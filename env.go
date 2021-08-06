package main

import (
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type typeError struct {
	reason     string
	stackTrace stackEntry
}

func (e typeError) Error() string {
	// TODO: display stacktrace
	return fmt.Sprintf("Type error: %s", e.reason)
}

type mathError struct {
	reason     string
	stackTrace stackEntry
}

func (e mathError) Error() string {
	// TODO: display stacktrace
	return fmt.Sprintf("Math error: %s", e.reason)
}

func (c *Context) requireArgLen(fnName string, args []Value, count int) error {
	if len(args) < count {
		return runtimeError{
			reason:     fmt.Sprintf("%s requires %d arguments, got %d", fnName, count, len(args)),
			stackTrace: c.generateStackTrace(),
		}
	}

	return nil
}

type builtinFn func([]Value) (Value, error)

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
	// core language and reflection
	c.LoadFunc("import", c.mgnImport)
	c.LoadFunc("string", c.mgnString)
	c.LoadFunc("int", c.mgnInt)
	c.LoadFunc("float", c.mgnFloat)
	c.LoadFunc("atom", c.mgnAtom)
	c.LoadFunc("codepoint", c.mgnCodepoint)
	c.LoadFunc("char", c.mgnChar)
	c.LoadFunc("type", c.mgnType)
	c.LoadFunc("len", c.mgnLen)
	c.LoadFunc("keys", c.mgnKeys)

	// os interfaces
	c.LoadFunc("args", c.mgnArgs)
	c.LoadFunc("env", c.mgnEnv)
	c.LoadFunc("time", c.mgnTime)
	c.LoadFunc("nanotime", c.mgnNanotime)
	c.LoadFunc("exit", c.mgnExit)
	// TODO: rand(type, max)
	c.LoadFunc("wait", c.callbackify(c.mgnWait))
	// TODO: exec(path, args, stdin)

	// i/o interfaces
	// TODO: input()
	c.LoadFunc("print", c.mgnPrint)
	// TODO: ls
	// TODO: mkdir
	// TODO: rm
	// TODO: stat
	c.LoadFunc("open", c.callbackify(c.mgnOpen))
	c.LoadFunc("close", c.callbackify(c.mgnClose))
	c.LoadFunc("read", c.callbackify(c.mgnRead))
	c.LoadFunc("write", c.callbackify(c.mgnWrite))
	// TODO: close := listen(host, handler)
	// TODO: req(data)

	// math
	c.LoadFunc("sin", c.mgnSin)
	c.LoadFunc("cos", c.mgnCos)
	c.LoadFunc("tan", c.mgnTan)
	c.LoadFunc("asin", c.mgnAsin)
	c.LoadFunc("acos", c.mgnAcos)
	c.LoadFunc("atan", c.mgnAtan)
	c.LoadFunc("pow", c.mgnPow)
	c.LoadFunc("log", c.mgnLog)
}

func errObj(message string) ObjectValue {
	return ObjectValue{
		"type":  AtomValue("error"),
		"error": MakeString(message),
	}
}

func (c *Context) callbackify(syncFn builtinFn) builtinFn {
	return func(args []Value) (Value, error) {
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

func (c *Context) mgnString(args []Value) (Value, error) {
	if err := c.requireArgLen("string", args, 1); err != nil {
		return nil, err
	}

	switch arg := args[0].(type) {
	case *StringValue:
		return arg, nil
	default:
		return MakeString(arg.String()), nil
	}
}

func (c *Context) mgnInt(args []Value) (Value, error) {
	if err := c.requireArgLen("int", args, 1); err != nil {
		return nil, err
	}

	switch arg := args[0].(type) {
	case IntValue:
		return arg, nil
	case FloatValue:
		return IntValue(math.Trunc(float64(arg))), nil
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

func (c *Context) mgnFloat(args []Value) (Value, error) {
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

func (c *Context) mgnAtom(args []Value) (Value, error) {
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

func (c *Context) mgnCodepoint(args []Value) (Value, error) {
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

func (c *Context) mgnChar(args []Value) (Value, error) {
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

func (c *Context) mgnType(args []Value) (Value, error) {
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

	panic("Unreachable!")
}

func (c *Context) mgnImport(args []Value) (Value, error) {
	if err := c.requireArgLen("import", args, 1); err != nil {
		return nil, err
	}

	pathBytes, ok := args[0].(*StringValue)
	if !ok {
		return nil, runtimeError{
			reason: fmt.Sprintf("path to import() must be a string, got %s", args[0]),
		}
	}
	pathStr := pathBytes.stringContent()

	// if a stdlib, just import the library from binary
	if isStdLib(pathStr) {
		return c.LoadLib(pathStr)
	}

	filePath := pathStr + ".mgn"
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(c.rootPath, filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, runtimeError{
			reason: fmt.Sprintf("could not open %s, %s", filePath, err.Error()),
		}
	}
	defer file.Close()

	if imported, ok := c.eng.importMap[filePath]; ok {
		return ObjectValue(imported.vars), nil
	}

	ctx := c.ChildContext(path.Dir(filePath))
	ctx.LoadBuiltins()

	ctx.Unlock()
	_, err = ctx.Eval(file)
	ctx.Lock()
	if err != nil {
		return nil, err
	}

	c.eng.importMap[filePath] = ctx.scope
	return ObjectValue(ctx.scope.vars), nil
}

func (c *Context) mgnLen(args []Value) (Value, error) {
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
		return nil, runtimeError{
			reason: fmt.Sprintf("%s does not support a len() call", arg),
		}
	}
}

func (c *Context) mgnPrint(args []Value) (Value, error) {
	if err := c.requireArgLen("print", args, 1); err != nil {
		return nil, err
	}

	outputString, ok := args[0].(*StringValue)
	if !ok {
		return nil, runtimeError{
			reason: fmt.Sprintf("unexpected argument to print: %s", args[0]),
		}
	}

	n, _ := os.Stdout.Write(*outputString)
	return IntValue(n), nil
}

func makeIntListUpTo(max int) Value {
	list := make(ListValue, max)
	for i := 0; i < max; i++ {
		list[i] = IntValue(i)
	}
	return &list
}

func (c *Context) mgnKeys(args []Value) (Value, error) {
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

func (c *Context) mgnArgs(_ []Value) (Value, error) {
	goArgs := os.Args
	args := make(ListValue, len(goArgs))
	for i, arg := range goArgs {
		args[i] = MakeString(arg)
	}
	return &args, nil
}

func (c *Context) mgnEnv(_ []Value) (Value, error) {
	envVars := ObjectValue{}
	for _, e := range os.Environ() {
		kv := strings.SplitN(e, "=", 2)
		envVars[kv[0]] = MakeString(kv[1])
	}
	return envVars, nil
}

func (c *Context) mgnTime(_ []Value) (Value, error) {
	unixSeconds := float64(time.Now().UnixNano()) / 1e9
	return FloatValue(unixSeconds), nil
}

func (c *Context) mgnNanotime(_ []Value) (Value, error) {
	return IntValue(time.Now().UnixNano()), nil
}

func (c *Context) mgnExit(args []Value) (Value, error) {
	if err := c.requireArgLen("exit", args, 1); err != nil {
		return nil, err
	}

	switch arg := args[0].(type) {
	case IntValue:
		os.Exit(int(arg))
		// unreachable
		return null, nil
	default:
		return nil, typeError{
			reason: fmt.Sprintf("Mismatched types in call exit(%s)", args[0]),
		}
	}
}

func (c *Context) mgnWait(args []Value) (Value, error) {
	if err := c.requireArgLen("wait", args, 1); err != nil {
		return nil, err
	}

	switch arg := args[0].(type) {
	case IntValue:
		time.Sleep(time.Duration(float64(arg) * float64(time.Second)))
	case FloatValue:
		time.Sleep(time.Duration(float64(arg) * float64(time.Second)))
	default:
		return nil, typeError{
			reason: fmt.Sprintf("Mismatched types in call wait(%s)", args[0]),
		}
	}

	return null, nil
}

func (c *Context) mgnOpen(args []Value) (Value, error) {
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
		return nil, typeError{
			reason: fmt.Sprintf("Mismatched types in call open(%s, %s)", args[0], args[1]),
		}
	}

	var flags int
	switch string(flagsAtom) {
	case "readonly":
		flags = os.O_RDONLY
	case "readwrite":
		flags = os.O_RDWR
	case "append":
		flags = os.O_RDWR | os.O_APPEND
	case "create":
		flags = os.O_RDWR | os.O_CREATE
	case "truncate":
		flags = os.O_RDWR | os.O_CREATE | os.O_TRUNC
	default:
		return nil, typeError{
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

func (c *Context) mgnClose(args []Value) (Value, error) {
	if err := c.requireArgLen("close", args, 1); err != nil {
		return nil, err
	}

	fdInt, ok1 := args[0].(IntValue)
	if !ok1 {
		return nil, typeError{
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

func (c *Context) mgnRead(args []Value) (Value, error) {
	if err := c.requireArgLen("read", args, 3); err != nil {
		return nil, err
	}

	fdInt, ok1 := args[0].(IntValue)
	offsetInt, ok2 := args[1].(IntValue)
	lengthInt, ok3 := args[2].(IntValue)
	if !ok1 || !ok2 || !ok3 {
		return nil, typeError{
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

func (c *Context) mgnWrite(args []Value) (Value, error) {
	if err := c.requireArgLen("write", args, 3); err != nil {
		return nil, err
	}

	fdInt, ok1 := args[0].(IntValue)
	offsetInt, ok2 := args[1].(IntValue)
	dataString, ok3 := args[2].(*StringValue)
	if !ok1 || !ok2 || !ok3 {
		return nil, typeError{
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

func (c *Context) mgnSin(args []Value) (Value, error) {
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
		return nil, typeError{
			reason: fmt.Sprintf("Mismatched types in call sin(%s)", args[0]),
		}
	}

	return FloatValue(math.Sin(val)), nil
}

func (c *Context) mgnCos(args []Value) (Value, error) {
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
		return nil, typeError{
			reason: fmt.Sprintf("Mismatched types in call cos(%s)", args[0]),
		}
	}

	return FloatValue(math.Cos(val)), nil
}

func (c *Context) mgnTan(args []Value) (Value, error) {
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
		return nil, typeError{
			reason: fmt.Sprintf("Mismatched types in call tan(%s)", args[0]),
		}
	}

	return FloatValue(math.Tan(val)), nil
}

func (c *Context) mgnAsin(args []Value) (Value, error) {
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
		return nil, typeError{
			reason: fmt.Sprintf("Mismatched types in call asin(%s)", args[0]),
		}
	}

	if val > 1 || val < -1 {
		return nil, runtimeError{
			reason: fmt.Sprintf("asin() takes a number in range [-1, 1], got %f", val),
		}
	}

	return FloatValue(math.Asin(val)), nil
}

func (c *Context) mgnAcos(args []Value) (Value, error) {
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
		return nil, typeError{
			reason: fmt.Sprintf("Mismatched types in call acos(%s)", args[0]),
		}
	}

	if val > 1 || val < -1 {
		return nil, runtimeError{
			reason: fmt.Sprintf("acos() takes a number in range [-1, 1], got %f", val),
		}
	}

	return FloatValue(math.Acos(val)), nil
}

func (c *Context) mgnAtan(args []Value) (Value, error) {
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
		return nil, typeError{
			reason: fmt.Sprintf("Mismatched types in call atan(%s)", args[0]),
		}
	}

	return FloatValue(math.Atan(val)), nil
}

func (c *Context) mgnPow(args []Value) (Value, error) {
	if err := c.requireArgLen("pow", args, 2); err != nil {
		return nil, err
	}

	var base float64
	var exp float64
	err := typeError{
		reason: fmt.Sprintf("Mismatched types in call pow(%s, %s)", args[0]),
	}

	switch arg := args[0].(type) {
	case IntValue:
		base = float64(arg)
	case FloatValue:
		base = float64(arg)
	default:
		return nil, err
	}

	switch arg := args[1].(type) {
	case IntValue:
		exp = float64(arg)
	case FloatValue:
		exp = float64(arg)
	default:
		return nil, err
	}

	if base == 0 && exp == 0 {
		return nil, mathError{
			reason: fmt.Sprintf("pow(0, 0) is not defined"),
		}
	} else if base < 0 && float64(int64(exp)) != exp {
		return nil, mathError{
			reason: fmt.Sprintf("pow() of negative number to fractional exponent is not defined"),
		}
	}

	return FloatValue(math.Pow(base, exp)), nil
}

func (c *Context) mgnLog(args []Value) (Value, error) {
	if err := c.requireArgLen("log", args, 2); err != nil {
		return nil, err
	}

	var base float64
	var exp float64
	err := typeError{
		reason: fmt.Sprintf("Mismatched types in call log(%s, %s)", args[0]),
	}

	switch arg := args[0].(type) {
	case IntValue:
		base = float64(arg)
	case FloatValue:
		base = float64(arg)
	default:
		return nil, err
	}

	switch arg := args[1].(type) {
	case IntValue:
		exp = float64(arg)
	case FloatValue:
		exp = float64(arg)
	default:
		return nil, err
	}

	if base == 0 {
		return nil, mathError{
			reason: fmt.Sprintf("log(0, _) is not defined"),
		}
	} else if exp == 0 {
		return nil, mathError{
			reason: fmt.Sprintf("log(_, 0) is not defined"),
		}
	}

	// we use math.Log2 here because we want logs of base 2 to give exact
	// answers, where we care less about other bases
	return FloatValue(math.Log2(exp) / math.Log2(base)), nil
}
