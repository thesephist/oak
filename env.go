package main

import (
	"bufio"
	"bytes"
	"context"
	crand "crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

func (c *Context) requireArgLen(fnName string, args []Value, count int) *runtimeError {
	if len(args) < count {
		return &runtimeError{
			reason: fmt.Sprintf("%s requires %d arguments, got %d", fnName, count, len(args)),
		}
	}

	return nil
}

type builtinFn func([]Value) (Value, *runtimeError)

type BuiltinFnValue struct {
	name string
	fn   builtinFn
}

func (v BuiltinFnValue) String() string {
	return fmt.Sprintf("fn %s { <native fn> }", v.name)
}
func (v BuiltinFnValue) Eq(u Value) bool {
	if _, ok := u.(EmptyValue); ok {
		return true
	}

	if w, ok := u.(BuiltinFnValue); ok {
		return v.name == w.name
	}
	return false
}

func (c *Context) LoadFunc(name string, fn builtinFn) {
	c.scope.put(name, BuiltinFnValue{
		name: name,
		fn:   fn,
	})
}

func (c *Context) LoadBuiltins() {
	// global initializations
	rand.Seed(time.Now().UnixNano())

	// core language and reflection
	c.LoadFunc("import", c.oakImport)
	c.LoadFunc("int", c.oakInt)
	c.LoadFunc("float", c.oakFloat)
	c.LoadFunc("atom", c.oakAtom)
	c.LoadFunc("string", c.oakString)
	c.LoadFunc("codepoint", c.oakCodepoint)
	c.LoadFunc("char", c.oakChar)
	c.LoadFunc("type", c.oakType)
	c.LoadFunc("len", c.oakLen)
	c.LoadFunc("keys", c.oakKeys)

	// os interfaces
	c.LoadFunc("args", c.oakArgs)
	c.LoadFunc("env", c.oakEnv)
	c.LoadFunc("time", c.oakTime)
	c.LoadFunc("nanotime", c.oakNanotime)
	c.LoadFunc("rand", c.oakRand)
	c.LoadFunc("srand", c.oakSrand)
	c.LoadFunc("wait", c.callbackify(c.oakWait))
	c.LoadFunc("exit", c.oakExit)
	c.LoadFunc("exec", c.callbackify(c.oakExec))

	// i/o interfaces
	c.LoadFunc("input", c.callbackify(c.oakInput))
	c.LoadFunc("print", c.oakPrint)
	c.LoadFunc("ls", c.callbackify(c.oakLs))
	c.LoadFunc("rm", c.callbackify(c.oakRm))
	c.LoadFunc("mkdir", c.callbackify(c.oakMkdir))
	c.LoadFunc("stat", c.callbackify(c.oakStat))
	c.LoadFunc("open", c.callbackify(c.oakOpen))
	c.LoadFunc("close", c.callbackify(c.oakClose))
	c.LoadFunc("read", c.callbackify(c.oakRead))
	c.LoadFunc("write", c.callbackify(c.oakWrite))
	c.LoadFunc("listen", c.oakListen)
	c.LoadFunc("req", c.callbackify(c.oakReq))

	// math
	c.LoadFunc("sin", c.oakSin)
	c.LoadFunc("cos", c.oakCos)
	c.LoadFunc("tan", c.oakTan)
	c.LoadFunc("asin", c.oakAsin)
	c.LoadFunc("acos", c.oakAcos)
	c.LoadFunc("atan", c.oakAtan)
	c.LoadFunc("pow", c.oakPow)
	c.LoadFunc("log", c.oakLog)

	// language and runtime APIs
	c.LoadFunc("___runtime_lib", c.rtLib)
	c.LoadFunc("___runtime_lib?", c.rtIsLib)
	c.LoadFunc("___runtime_gc", c.rtGC)
	c.LoadFunc("___runtime_mem", c.rtMem)
}

func errObj(message string) ObjectValue {
	return ObjectValue{
		"type":  AtomValue("error"),
		"error": MakeString(message),
	}
}

func (c *Context) callbackify(syncFn builtinFn) builtinFn {
	return func(args []Value) (Value, *runtimeError) {
		if len(args) == 0 {
			return syncFn(args)
		}

		lastArg := args[len(args)-1]
		callback, isCallbackFn := lastArg.(FnValue)
		if !isCallbackFn {
			return syncFn(args)
		}

		syncArgs := args[:len(args)-1]
		c.eng.Add(1)
		go func() {
			defer c.eng.Done()

			evt, err := syncFn(syncArgs)
			if err != nil {
				c.eng.reportErr(err)
				return
			}

			c.Lock()
			defer c.Unlock()
			_, err = c.EvalFnValue(callback, false, evt)
			if err != nil {
				c.eng.reportErr(err)
				return
			}
		}()

		return null, nil
	}
}

func (c *Context) oakImport(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("import", args, 1); err != nil {
		return nil, err
	}

	pathBytes, ok := args[0].(*StringValue)
	if !ok {
		return nil, &runtimeError{
			reason: fmt.Sprintf("path to import() must be a string, got %s", args[0]),
		}
	}
	pathStr := pathBytes.stringContent()

	// if a stdlib, just import the library from binary
	if isStdLib(pathStr) {
		return c.LoadLib(pathStr)
	}

	filePath := pathStr + ".oak"
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(c.rootPath, filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, &runtimeError{
			reason: fmt.Sprintf("Could not open %s, %s", filePath, err.Error()),
		}
	}
	defer file.Close()

	if imported, ok := c.eng.importMap[filePath]; ok {
		return ObjectValue(imported.vars), nil
	}

	ctx := c.ChildContext(path.Dir(filePath))
	c.eng.importMap[filePath] = ctx.scope
	ctx.LoadBuiltins()

	ctx.Unlock()
	_, err = ctx.Eval(file)
	ctx.Lock()
	if err != nil {
		if runtimeErr, ok := err.(*runtimeError); ok {
			return nil, runtimeErr
		} else {
			return nil, &runtimeError{
				reason: fmt.Sprintf("Error importing %s: %s", pathStr, err.Error()),
			}
		}
	}

	return ObjectValue(ctx.scope.vars), nil
}

func (c *Context) oakInt(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("int", args, 1); err != nil {
		return nil, err
	}

	switch arg := args[0].(type) {
	case IntValue:
		return arg, nil
	case FloatValue:
		return IntValue(math.Floor(float64(arg))), nil
	case *StringValue:
		n, err := strconv.ParseInt(arg.stringContent(), 10, 64)
		if err != nil {
			return null, nil
		}
		return IntValue(n), nil
	default:
		return null, nil
	}
}

func (c *Context) oakFloat(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("float", args, 1); err != nil {
		return nil, err
	}

	switch arg := args[0].(type) {
	case IntValue:
		return FloatValue(arg), nil
	case FloatValue:
		return arg, nil
	case *StringValue:
		f, err := strconv.ParseFloat(arg.stringContent(), 64)
		if err != nil {
			return null, nil
		}
		return FloatValue(f), nil
	default:
		return null, nil
	}
}

func (c *Context) oakAtom(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("atom", args, 1); err != nil {
		return nil, err
	}

	switch arg := args[0].(type) {
	case *StringValue:
		return AtomValue(arg.stringContent()), nil
	case AtomValue:
		return arg, nil
	default:
		return AtomValue(arg.String()), nil
	}
}

func (c *Context) oakString(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("string", args, 1); err != nil {
		return nil, err
	}

	switch arg := args[0].(type) {
	case *StringValue:
		return arg, nil
	case AtomValue:
		return MakeString(string(arg)), nil
	default:
		return MakeString(arg.String()), nil
	}
}

func (c *Context) oakCodepoint(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("codepoint", args, 1); err != nil {
		return nil, err
	}

	switch arg := args[0].(type) {
	case *StringValue:
		if len(*arg) != 1 {
			return null, nil
		}
		return IntValue(int64((*arg)[0])), nil
	default:
		return null, nil
	}
}

func (c *Context) oakChar(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("char", args, 1); err != nil {
		return nil, err
	}

	switch arg := args[0].(type) {
	case IntValue:
		codepoint := int64(arg)
		if codepoint < 0 {
			codepoint = 0
		}
		if codepoint > 255 {
			codepoint = 255
		}
		bytes := StringValue([]byte{byte(codepoint)})
		return &bytes, nil
	default:
		return null, nil
	}
}

func (c *Context) oakType(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("type", args, 1); err != nil {
		return nil, err
	}

	switch args[0].(type) {
	case NullValue:
		return AtomValue("null"), nil
	case EmptyValue:
		return AtomValue("empty"), nil
	case IntValue:
		return AtomValue("int"), nil
	case FloatValue:
		return AtomValue("float"), nil
	case BoolValue:
		return AtomValue("bool"), nil
	case AtomValue:
		return AtomValue("atom"), nil
	case *StringValue:
		return AtomValue("string"), nil
	case *ListValue:
		return AtomValue("list"), nil
	case ObjectValue:
		return AtomValue("object"), nil
	case FnValue, BuiltinFnValue:
		return AtomValue("function"), nil
	}

	panic("Unreachable: unknown runtime value")
}

func (c *Context) oakLen(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("string", args, 1); err != nil {
		return nil, err
	}

	switch arg := args[0].(type) {
	case *StringValue:
		return IntValue(len(*arg)), nil
	case *ListValue:
		return IntValue(len(*arg)), nil
	case ObjectValue:
		return IntValue(len(arg)), nil
	default:
		return nil, &runtimeError{
			reason: fmt.Sprintf("%s does not support a len() call", arg),
		}
	}
}

func makeIntListUpTo(max int) Value {
	list := make(ListValue, max)
	for i := 0; i < max; i++ {
		list[i] = IntValue(i)
	}
	return &list
}

func (c *Context) oakKeys(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("print", args, 1); err != nil {
		return nil, err
	}

	switch arg := args[0].(type) {
	case *StringValue:
		return makeIntListUpTo(len(*arg)), nil
	case *ListValue:
		return makeIntListUpTo(len(*arg)), nil
	case ObjectValue:
		keys := make(ListValue, len(arg))
		i := 0
		for key := range arg {
			keys[i] = MakeString(key)
			i++
		}
		return &keys, nil
	default:
		return MakeList(), nil
	}
}

func (c *Context) oakArgs(_ []Value) (Value, *runtimeError) {
	goArgs := os.Args
	args := make(ListValue, len(goArgs))
	for i, arg := range goArgs {
		args[i] = MakeString(arg)
	}
	return &args, nil
}

func (c *Context) oakEnv(_ []Value) (Value, *runtimeError) {
	envVars := ObjectValue{}
	for _, e := range os.Environ() {
		kv := strings.SplitN(e, "=", 2)
		envVars[kv[0]] = MakeString(kv[1])
	}
	return envVars, nil
}

func (c *Context) oakTime(_ []Value) (Value, *runtimeError) {
	unixSeconds := float64(time.Now().UnixNano()) / 1e9
	return FloatValue(unixSeconds), nil
}

func (c *Context) oakNanotime(_ []Value) (Value, *runtimeError) {
	return IntValue(time.Now().UnixNano()), nil
}

func (c *Context) oakRand(_ []Value) (Value, *runtimeError) {
	return FloatValue(rand.Float64()), nil
}

func (c *Context) oakSrand(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("srand", args, 1); err != nil {
		return nil, err
	}

	bufLen, ok1 := args[0].(IntValue)
	if !ok1 {
		return nil, &runtimeError{
			reason: fmt.Sprintf("Mismatched types in call srand(%s)", args[0]),
		}
	}

	buf := make([]byte, bufLen)
	_, err := crand.Read(buf)
	if err != nil {
		return null, nil
	}

	bytes := StringValue(buf)
	return &bytes, nil
}

func (c *Context) oakWait(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("wait", args, 1); err != nil {
		return nil, err
	}

	// in both Oak & Go, duration <= 0 results in immediate completion
	switch arg := args[0].(type) {
	case IntValue:
		time.Sleep(time.Duration(float64(arg) * float64(time.Second)))
	case FloatValue:
		time.Sleep(time.Duration(float64(arg) * float64(time.Second)))
	default:
		return nil, &runtimeError{
			reason: fmt.Sprintf("Mismatched types in call wait(%s)", args[0]),
		}
	}

	return null, nil
}

func (c *Context) oakExit(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("exit", args, 1); err != nil {
		return nil, err
	}

	switch arg := args[0].(type) {
	case IntValue:
		os.Exit(int(arg))
		// unreachable
		return null, nil
	default:
		return nil, &runtimeError{
			reason: fmt.Sprintf("Mismatched types in call exit(%s)", args[0]),
		}
	}
}

func (c *Context) oakExec(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("exec", args, 3); err != nil {
		return nil, err
	}

	path, ok1 := args[0].(*StringValue)
	cliArgs, ok2 := args[1].(*ListValue)
	stdin, ok3 := args[2].(*StringValue)
	if !ok1 || !ok2 || !ok3 {
		return nil, &runtimeError{
			reason: fmt.Sprintf("Mismatched types in call exec(%s, %s, %s)", args[0], args[1], args[2]),
		}
	}

	argsList := make([]string, len(*cliArgs))
	for i, arg := range *cliArgs {
		if argStr, ok := arg.(*StringValue); ok {
			argsList[i] = argStr.stringContent()
		} else {
			return nil, &runtimeError{
				reason: fmt.Sprintf("Mismatched types in call exec, arguments must be strings in %s", cliArgs),
			}
		}
	}

	cmd := exec.Command(path.stringContent(), argsList...)
	stdoutBuf := bytes.Buffer{}
	stderrBuf := bytes.Buffer{}
	cmd.Stdin = strings.NewReader(stdin.stringContent())
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Start()
	if err != nil {
		return errObj(fmt.Sprintf("Could not start command in exec(): %s", err.Error())), nil
	}

	err = cmd.Wait()
	exitCode := 0
	if err != nil {
		// if there is an err but err is just ExitErr, this means the process
		// ran successfully but exited with an error code. We consider this ok
		// and keep going.
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
		}
	}

	stdout, err := io.ReadAll(&stdoutBuf)
	if err != nil {
		return errObj(fmt.Sprintf("Could not read stdout from exec(): %s", err.Error())), nil
	}
	stdoutVal := StringValue(stdout)
	stderr, err := io.ReadAll(&stderrBuf)
	if err != nil {
		return errObj(fmt.Sprintf("Could not read stderr from exec(): %s", err.Error())), nil
	}
	stderrVal := StringValue(stderr)

	return ObjectValue{
		"type":   AtomValue("end"),
		"status": IntValue(exitCode),
		"stdout": &stdoutVal,
		"stderr": &stderrVal,
	}, nil
}

var inputReaderInit sync.Once
var inputReader *bufio.Reader

func initInputReader() {
	inputReader = bufio.NewReader(os.Stdin)
}

func (c *Context) oakInput(_ []Value) (Value, *runtimeError) {
	inputReaderInit.Do(initInputReader)
	str, err := inputReader.ReadString('\n')
	if err == io.EOF {
		return ObjectValue{
			"type":  AtomValue("error"),
			"error": MakeString("EOF"),
			// if any data was read before encountering EOF, ensure the caller
			// still gets that data.
			"data": MakeString(str),
		}, nil
	} else if err != nil {
		return errObj(fmt.Sprintf("Could not read input: %s", err.Error())), nil
	}

	inputStr := strings.TrimSuffix(str, "\n")

	return ObjectValue{
		"type": AtomValue("data"),
		"data": MakeString(inputStr),
	}, nil
}

func (c *Context) oakPrint(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("print", args, 1); err != nil {
		return nil, err
	}

	outputString, ok := args[0].(*StringValue)
	if !ok {
		return nil, &runtimeError{
			reason: fmt.Sprintf("Unexpected argument to print: %s", args[0]),
		}
	}

	n, _ := os.Stdout.Write(*outputString)
	return IntValue(n), nil
}

func (c *Context) oakLs(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("ls", args, 1); err != nil {
		return nil, err
	}

	dirPath, ok1 := args[0].(*StringValue)
	if !ok1 {
		return nil, &runtimeError{
			reason: fmt.Sprintf("Mismatched types in call ls(%s)", args[0]),
		}
	}

	fileInfos, err := ioutil.ReadDir(dirPath.stringContent())
	if err != nil {
		return errObj(fmt.Sprintf("Could not list directory %s: %s", dirPath.stringContent(), err.Error())), nil
	}

	fileList := make(ListValue, len(fileInfos))
	for i, fi := range fileInfos {
		fileList[i] = ObjectValue{
			"name": MakeString(fi.Name()),
			"len":  IntValue(fi.Size()),
			"dir":  BoolValue(fi.IsDir()),
			"mod":  IntValue(fi.ModTime().Unix()),
		}
	}

	return ObjectValue{
		"type": AtomValue("data"),
		"data": &fileList,
	}, nil
}

func (c *Context) oakRm(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("rm", args, 1); err != nil {
		return nil, err
	}

	rmPath, ok1 := args[0].(*StringValue)
	if !ok1 {
		return nil, &runtimeError{
			reason: fmt.Sprintf("Mismatched types in call rm(%s)", args[0]),
		}
	}

	err := os.RemoveAll(rmPath.stringContent())
	if err != nil {
		return errObj(fmt.Sprintf("Could not remove %s: %s", rmPath.stringContent(), err.Error())), nil
	}

	return ObjectValue{
		"type": AtomValue("end"),
	}, nil
}

func (c *Context) oakMkdir(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("mkdir", args, 1); err != nil {
		return nil, err
	}

	dirPath, ok1 := args[0].(*StringValue)
	if !ok1 {
		return nil, &runtimeError{
			reason: fmt.Sprintf("Mismatched types in call mkdir(%s)", args[0]),
		}
	}

	err := os.MkdirAll(dirPath.stringContent(), 0755)
	if err != nil {
		return errObj(fmt.Sprintf("Could not make a new directory %s: %s", dirPath.stringContent(), err.Error())), nil
	}

	return ObjectValue{
		"type": AtomValue("end"),
	}, nil
}

func (c *Context) oakStat(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("stat", args, 1); err != nil {
		return nil, err
	}

	statPath, ok1 := args[0].(*StringValue)
	if !ok1 {
		return nil, &runtimeError{
			reason: fmt.Sprintf("Mismatched types in call stat(%s)", args[0]),
		}
	}

	fileInfo, err := os.Stat(statPath.stringContent())
	if err != nil {
		if os.IsNotExist(err) {
			return ObjectValue{
				"type": AtomValue("data"),
				"data": null,
			}, nil
		}
		return errObj(fmt.Sprintf("Could not stat file %s: %s", statPath.stringContent(), err.Error())), nil
	}

	return ObjectValue{
		"type": AtomValue("data"),
		"data": ObjectValue{
			"name": MakeString(fileInfo.Name()),
			"len":  IntValue(fileInfo.Size()),
			"dir":  BoolValue(fileInfo.IsDir()),
			"mod":  IntValue(fileInfo.ModTime().Unix()),
		},
	}, nil
}

func (c *Context) oakOpen(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("open", args, 1); err != nil {
		return nil, err
	}

	// flags arg is optional
	if len(args) < 2 {
		args = append(args, AtomValue("readwrite"))
	}

	// perm arg is optional
	if len(args) < 3 {
		args = append(args, IntValue(0644))
	}

	pathString, ok1 := args[0].(*StringValue)
	flagsAtom, ok2 := args[1].(AtomValue)
	permInt, ok3 := args[2].(IntValue)
	if !ok1 || !ok2 || !ok3 {
		return nil, &runtimeError{
			reason: fmt.Sprintf("Mismatched types in call open(%s, %s, %s)", args[0], args[1], args[2]),
		}
	}

	var flags int
	switch string(flagsAtom) {
	case "readonly":
		flags = os.O_RDONLY
	case "readwrite":
		flags = os.O_RDWR | os.O_CREATE
	case "append":
		flags = os.O_RDWR | os.O_CREATE | os.O_APPEND
	case "truncate":
		flags = os.O_RDWR | os.O_CREATE | os.O_TRUNC
	default:
		return nil, &runtimeError{
			reason: fmt.Sprintf("Invalid flag for open(): %s", flagsAtom),
		}
	}

	file, err := os.OpenFile(pathString.stringContent(), flags, os.FileMode(permInt))
	if err != nil {
		return errObj(fmt.Sprintf("Could not open file: %s", err.Error())), nil
	}

	fd := file.Fd()

	c.eng.fdLock.Lock()
	defer c.eng.fdLock.Unlock()
	c.eng.fileMap[fd] = file

	return ObjectValue{
		"type": AtomValue("file"),
		"fd":   IntValue(fd),
	}, nil
}

func (c *Context) oakClose(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("close", args, 1); err != nil {
		return nil, err
	}

	fdInt, ok1 := args[0].(IntValue)
	if !ok1 {
		return nil, &runtimeError{
			reason: fmt.Sprintf("Mismatched types in call close(%s)", args[0]),
		}
	}

	c.eng.fdLock.Lock()
	defer c.eng.fdLock.Unlock()
	file, ok := c.eng.fileMap[uintptr(fdInt)]

	if !ok {
		return errObj(fmt.Sprintf("Unknown fd %d", fdInt)), nil
	}

	err := file.Close()
	if err != nil {
		return errObj(fmt.Sprintf("Could not close file: %s", err.Error())), nil
	}

	delete(c.eng.fileMap, uintptr(fdInt))

	return ObjectValue{
		"type": AtomValue("end"),
	}, nil
}

func (c *Context) oakRead(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("read", args, 3); err != nil {
		return nil, err
	}

	fdInt, ok1 := args[0].(IntValue)
	offsetInt, ok2 := args[1].(IntValue)
	lengthInt, ok3 := args[2].(IntValue)
	if !ok1 || !ok2 || !ok3 {
		return nil, &runtimeError{
			reason: fmt.Sprintf("Mismatched types in call read(%s, %s, %s)", args[0], args[1], args[2]),
		}
	}

	c.eng.fdLock.Lock()
	file, ok := c.eng.fileMap[uintptr(fdInt)]
	c.eng.fdLock.Unlock()

	if !ok {
		return errObj(fmt.Sprintf("Unknown fd %d", fdInt)), nil
	}

	offset := int64(offsetInt)
	length := int64(lengthInt)

	_, err := file.Seek(offset, 0)
	if err != nil {
		return errObj(fmt.Sprintf("Error reading file during seek: %s", err.Error())), nil
	}

	readBuf := make([]byte, length)
	count, err := file.Read(readBuf)
	if err != nil && err != io.EOF {
		return errObj(fmt.Sprintf("Error reading file: %s", err.Error())), nil
	}

	fileData := StringValue(readBuf[:count])
	return ObjectValue{
		"type": AtomValue("data"),
		"data": &fileData,
	}, nil
}

func (c *Context) oakWrite(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("write", args, 3); err != nil {
		return nil, err
	}

	fdInt, ok1 := args[0].(IntValue)
	offsetInt, ok2 := args[1].(IntValue)
	dataString, ok3 := args[2].(*StringValue)
	if !ok1 || !ok2 || !ok3 {
		return nil, &runtimeError{
			reason: fmt.Sprintf("Mismatched types in call write(%s, %s, %s)", args[0], args[1], args[2]),
		}
	}

	c.eng.fdLock.Lock()
	file, ok := c.eng.fileMap[uintptr(fdInt)]
	c.eng.fdLock.Unlock()
	if !ok {
		return errObj(fmt.Sprintf("Unknown fd %d", fdInt)), nil
	}

	offset := int64(offsetInt)
	writeBuf := []byte(*dataString)

	var err error
	if offset == -1 {
		_, err = file.Seek(0, 2) // "2" is relative to end of file
	} else {
		_, err = file.Seek(offset, 0)
	}
	if err != nil {
		return errObj(fmt.Sprintf("Error writing file during seek: %s", err.Error())), nil
	}

	_, err = file.Write(writeBuf)
	if err != nil && err != io.EOF {
		return errObj(fmt.Sprintf("Error writing file: %s", err.Error())), nil
	}

	return ObjectValue{
		"type": AtomValue("end"),
	}, nil
}

type oakHTTPHandler struct {
	ctx         *Context
	oakCallback FnValue
}

func (h oakHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := h.ctx
	cb := h.oakCallback

	// unmarshal request
	method := r.Method
	url := r.URL.String()
	headers := ObjectValue{}
	for key, values := range r.Header {
		headers[key] = MakeString(strings.Join(values, ","))
	}
	var body *StringValue
	if r.ContentLength == 0 {
		body = MakeString("")
	} else {
		bodyBuf, err := io.ReadAll(r.Body)
		if err != nil {
			ctx.Lock()
			_, err = ctx.EvalFnValue(cb, false, errObj(
				fmt.Sprintf("Could not read request in listen(), %s", err.Error()),
			))
			ctx.Unlock()

			if err != nil {
				ctx.eng.reportErr(err)
			}
		}
		bodyStr := StringValue(bodyBuf)
		body = &bodyStr
	}

	// construct request object to pass to Oak, call handler
	responseEnded := false
	responses := make(chan Value, 1)
	endHandler := func(args []Value) (Value, *runtimeError) {
		if err := ctx.requireArgLen("listen/end", args, 1); err != nil {
			return nil, err
		}

		if responseEnded {
			ctx.eng.reportErr(&runtimeError{
				reason: fmt.Sprintf("listen/end called more than once"),
			})
		}

		responseEnded = true
		responses <- args[0]

		return null, nil
	}

	go func() {
		ctx.Lock()
		defer ctx.Unlock()

		_, err := ctx.EvalFnValue(cb, false, ObjectValue{
			"type": AtomValue("req"),
			"req": ObjectValue{
				"method":  MakeString(method),
				"url":     MakeString(url),
				"headers": headers,
				"body":    body,
			},
			"end": BuiltinFnValue{
				name: "end",
				fn:   endHandler,
			},
		})
		if err != nil {
			ctx.eng.reportErr(err)
		}
	}()

	// validate responses
	resp := <-responses
	rsp, isObject := resp.(ObjectValue)
	if !isObject {
		ctx.eng.reportErr(&runtimeError{
			reason: fmt.Sprintf("listen/end should return a response, got %s", resp),
		})
		return
	}

	// unmarshal response from the return value
	// response = { status, headers, body }
	statusVal, okStatus := rsp["status"]
	headersVal, okHeaders := rsp["headers"]
	bodyVal, okBody := rsp["body"]

	resStatus, okStatus := statusVal.(IntValue)
	resHeaders, okHeaders := headersVal.(ObjectValue)
	resBody, okBody := bodyVal.(*StringValue)

	if !okStatus || !okHeaders || !okBody {
		ctx.eng.reportErr(&runtimeError{
			reason: fmt.Sprintf("listen/end returned malformed response, %s", rsp),
		})
		return
	}

	// write values to response
	// Content-Length is automatically set for us by Go
	for k, v := range resHeaders {
		if str, isStr := v.(*StringValue); isStr {
			w.Header().Set(k, str.stringContent())
		} else {
			ctx.eng.reportErr(&runtimeError{
				reason: fmt.Sprintf("Could not set response header, value %s was not a string", v),
			})
			return
		}
	}

	code := int(resStatus)
	// guard against invalid HTTP codes, which cause Go panics
	// https://golang.org/src/net/http/server.go
	if code < 100 || code > 599 {
		ctx.eng.reportErr(&runtimeError{
			reason: fmt.Sprintf("Could not set response status code, code %d is not valid", code),
		})
		return
	}

	// status code write must follow all other header writes, since it sends
	// the status
	w.WriteHeader(int(resStatus))
	_, err := w.Write(*resBody)
	if err != nil {
		ctx.Lock()
		defer ctx.Unlock()

		_, err = ctx.EvalFnValue(cb, false, errObj(
			fmt.Sprintf("Error writing request body in listen/end: %s", err.Error()),
		))
		if err != nil {
			ctx.eng.reportErr(err)
		}
	}
}

func (ctx *Context) oakListen(args []Value) (Value, *runtimeError) {
	if err := ctx.requireArgLen("listen", args, 2); err != nil {
		return nil, err
	}

	host, ok1 := args[0].(*StringValue)
	cb, ok2 := args[1].(FnValue)
	if !ok1 || !ok2 {
		return nil, &runtimeError{
			reason: fmt.Sprintf("Mismatched types in call listen(%s)", args[0]),
		}
	}

	sendErr := func(msg string) {
		ctx.Lock()
		defer ctx.Unlock()

		_, err2 := ctx.EvalFnValue(cb, false, errObj(msg))
		if err2 != nil {
			ctx.eng.reportErr(err2)
		}
	}

	server := &http.Server{
		Addr: host.stringContent(),
		Handler: oakHTTPHandler{
			ctx:         ctx,
			oakCallback: cb,
		},
	}

	ctx.eng.Add(1)
	go func() {
		defer ctx.eng.Done()
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			sendErr(fmt.Sprintf("Error starting http server in listen(): %s", err.Error()))
		}
	}()

	closer := func(_ []Value) (Value, *runtimeError) {
		// attempt graceful shutdown, concurrently, without blocking Oak
		// evaluation thread
		ctx.eng.Add(1)
		go func() {
			defer ctx.eng.Done()

			err := server.Shutdown(context.Background())
			if err != nil {
				sendErr(fmt.Sprintf("Could not close server in listen/close: %s", err.Error()))
			}
		}()

		return null, nil
	}

	return BuiltinFnValue{
		name: "close",
		fn:   closer,
	}, nil
}

func (c *Context) oakReq(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("req", args, 1); err != nil {
		return nil, err
	}

	argErr := runtimeError{
		reason: fmt.Sprintf("Mismatched types in call req(%s)", args[0]),
	}

	data, ok1 := args[0].(ObjectValue)
	if !ok1 {
		return nil, &argErr
	}

	// unmarshal request data
	methodVal, ok1 := data["method"]
	urlVal, ok2 := data["url"]
	headersVal, ok3 := data["headers"]
	bodyVal, ok4 := data["body"]

	// default args
	if !ok1 {
		methodVal = MakeString("GET")
		ok1 = true
	}
	if !ok3 {
		headersVal = ObjectValue{}
		ok3 = true
	}
	if !ok4 {
		bodyVal = MakeString("")
		ok4 = true
	}

	if !ok1 || !ok2 || !ok3 || !ok4 {
		return nil, &argErr
	}

	method, ok1 := methodVal.(*StringValue)
	url, ok2 := urlVal.(*StringValue)
	headers, ok3 := headersVal.(ObjectValue)
	body, ok4 := bodyVal.(*StringValue)
	if !ok1 || !ok2 || !ok3 || !ok4 {
		return nil, &argErr
	}

	client := &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			// do not follow redirects
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest(
		method.stringContent(),
		url.stringContent(),
		strings.NewReader(body.stringContent()),
	)
	if err != nil {
		return errObj(fmt.Sprintf("Could not initialize request in req(): %s", err.Error())), nil
	}

	// construct headers
	// Content-Length is automatically set for us by Go
	req.Header.Set("User-Agent", "") // remove Go's default user agent header
	for k, v := range headers {
		if valStr, ok := v.(*StringValue); ok {
			req.Header.Set(k, valStr.stringContent())
		} else {
			return nil, &runtimeError{
				reason: fmt.Sprintf("Could not set request header, value %s is not a string", v),
			}
		}
	}

	// send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, &runtimeError{
			reason: fmt.Sprintf("Could not send request: %s", err.Error()),
		}
	}
	defer resp.Body.Close()

	respStatus := IntValue(resp.StatusCode)
	respHeaders := ObjectValue{}
	for key, values := range resp.Header {
		respHeaders[key] = MakeString(strings.Join(values, ","))
	}

	var respBody *StringValue
	if resp.ContentLength == 0 {
		respBody = MakeString("")
	} else {
		buf, err := io.ReadAll(resp.Body)
		if err != nil {
			return errObj(fmt.Sprintf("Could not read response: %s", err.Error())), nil
		}
		strBuf := StringValue(buf)
		respBody = &strBuf
	}

	return ObjectValue{
		"type": AtomValue("resp"),
		"resp": ObjectValue{
			"status":  respStatus,
			"headers": respHeaders,
			"body":    respBody,
		},
	}, nil
}

func (c *Context) oakSin(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("sin", args, 1); err != nil {
		return nil, err
	}

	var val float64
	switch arg := args[0].(type) {
	case IntValue:
		val = float64(arg)
	case FloatValue:
		val = float64(arg)
	default:
		return nil, &runtimeError{
			reason: fmt.Sprintf("Mismatched types in call sin(%s)", args[0]),
		}
	}

	return FloatValue(math.Sin(val)), nil
}

func (c *Context) oakCos(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("cos", args, 1); err != nil {
		return nil, err
	}

	var val float64
	switch arg := args[0].(type) {
	case IntValue:
		val = float64(arg)
	case FloatValue:
		val = float64(arg)
	default:
		return nil, &runtimeError{
			reason: fmt.Sprintf("Mismatched types in call cos(%s)", args[0]),
		}
	}

	return FloatValue(math.Cos(val)), nil
}

func (c *Context) oakTan(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("tan", args, 1); err != nil {
		return nil, err
	}

	var val float64
	switch arg := args[0].(type) {
	case IntValue:
		val = float64(arg)
	case FloatValue:
		val = float64(arg)
	default:
		return nil, &runtimeError{
			reason: fmt.Sprintf("Mismatched types in call tan(%s)", args[0]),
		}
	}

	return FloatValue(math.Tan(val)), nil
}

func (c *Context) oakAsin(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("asin", args, 1); err != nil {
		return nil, err
	}

	var val float64
	switch arg := args[0].(type) {
	case IntValue:
		val = float64(arg)
	case FloatValue:
		val = float64(arg)
	default:
		return nil, &runtimeError{
			reason: fmt.Sprintf("Mismatched types in call asin(%s)", args[0]),
		}
	}

	if val > 1 || val < -1 {
		return nil, &runtimeError{
			reason: fmt.Sprintf("asin() takes a number in range [-1, 1], got %f", val),
		}
	}

	return FloatValue(math.Asin(val)), nil
}

func (c *Context) oakAcos(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("acos", args, 1); err != nil {
		return nil, err
	}

	var val float64
	switch arg := args[0].(type) {
	case IntValue:
		val = float64(arg)
	case FloatValue:
		val = float64(arg)
	default:
		return nil, &runtimeError{
			reason: fmt.Sprintf("Mismatched types in call acos(%s)", args[0]),
		}
	}

	if val > 1 || val < -1 {
		return nil, &runtimeError{
			reason: fmt.Sprintf("acos() takes a number in range [-1, 1], got %f", val),
		}
	}

	return FloatValue(math.Acos(val)), nil
}

func (c *Context) oakAtan(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("atan", args, 1); err != nil {
		return nil, err
	}

	var val float64
	switch arg := args[0].(type) {
	case IntValue:
		val = float64(arg)
	case FloatValue:
		val = float64(arg)
	default:
		return nil, &runtimeError{
			reason: fmt.Sprintf("Mismatched types in call atan(%s)", args[0]),
		}
	}

	return FloatValue(math.Atan(val)), nil
}

func (c *Context) oakPow(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("pow", args, 2); err != nil {
		return nil, err
	}

	var base float64
	var exp float64
	err := runtimeError{
		reason: fmt.Sprintf("Mismatched types in call pow(%s, %s)", args[0], args[1]),
	}

	switch arg := args[0].(type) {
	case IntValue:
		base = float64(arg)
	case FloatValue:
		base = float64(arg)
	default:
		return nil, &err
	}

	switch arg := args[1].(type) {
	case IntValue:
		exp = float64(arg)
	case FloatValue:
		exp = float64(arg)
	default:
		return nil, &err
	}

	if base == 0 && exp == 0 {
		return nil, &runtimeError{
			reason: fmt.Sprintf("pow(0, 0) is not defined"),
		}
	} else if base < 0 && float64(int64(exp)) != exp {
		return nil, &runtimeError{
			reason: fmt.Sprintf("pow() of negative number to fractional exponent is not defined"),
		}
	}

	return FloatValue(math.Pow(base, exp)), nil
}

func (c *Context) oakLog(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("log", args, 2); err != nil {
		return nil, err
	}

	var base float64
	var exp float64
	err := runtimeError{
		reason: fmt.Sprintf("Mismatched types in call log(%s, %s)", args[0], args[1]),
	}

	switch arg := args[0].(type) {
	case IntValue:
		base = float64(arg)
	case FloatValue:
		base = float64(arg)
	default:
		return nil, &err
	}

	switch arg := args[1].(type) {
	case IntValue:
		exp = float64(arg)
	case FloatValue:
		exp = float64(arg)
	default:
		return nil, &err
	}

	if base == 0 {
		return nil, &runtimeError{
			reason: fmt.Sprintf("log(0, _) is not defined"),
		}
	} else if exp == 0 {
		return nil, &runtimeError{
			reason: fmt.Sprintf("log(_, 0) is not defined"),
		}
	}

	// we use math.Log2 here because we want logs of base 2 to give exact
	// answers, where we care less about other bases
	return FloatValue(math.Log2(exp) / math.Log2(base)), nil
}

// ___runtime_lib returns the string content of the bundled standard library by
// the given name, or ? otherwise.
func (c *Context) rtLib(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("__runtime_lib", args, 1); err != nil {
		return nil, err
	}

	switch arg := args[0].(type) {
	case *StringValue:
		libName := arg.stringContent()
		if libSource, ok := stdlibs[libName]; ok {
			return MakeString(libSource), nil
		}
		return null, nil
	default:
		return nil, &runtimeError{
			reason: fmt.Sprintf("Mismatched types in call ___runtime_lib(%s)", args[0]),
		}
	}
}

// ___runtime_lib? reports whether a bundled standard library by the given name exists
func (c *Context) rtIsLib(args []Value) (Value, *runtimeError) {
	if err := c.requireArgLen("__runtime_lib?", args, 1); err != nil {
		return nil, err
	}

	switch arg := args[0].(type) {
	case *StringValue:
		libName := arg.stringContent()
		_, ok := stdlibs[libName]
		return BoolValue(ok), nil
	default:
		return nil, &runtimeError{
			reason: fmt.Sprintf("Mismatched types in call ___runtime_lib?(%s)", args[0]),
		}
	}
}

// ___runtime_gc runs a garbage collection cycle for both Oak and the
// underlying Go runtime. It blocks until the GC cycle is complete.
func (c *Context) rtGC(_ []Value) (Value, *runtimeError) {
	runtime.GC()
	return null, nil
}

// ___runtime_mem reports a dictionary of memory usage statistics for diagnostics
func (c *Context) rtMem(_ []Value) (Value, *runtimeError) {
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)
	return ObjectValue{
		// number of allocations
		"allocs": IntValue(memStats.Mallocs),
		"frees":  IntValue(memStats.Frees),
		"live":   IntValue(memStats.Mallocs - memStats.Frees),
		// number of bytes
		"heap": IntValue(memStats.HeapAlloc),
		"virt": IntValue(memStats.HeapSys),
		// total gc cycles count
		"gcs": IntValue(memStats.NumGC),
	}, nil
}
