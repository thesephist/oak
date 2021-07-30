package main

import (
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

type typeError struct {
	reason     string
	stackTrace stackEntry
}

func (e typeError) Error() string {
	// TODO: display stacktrace
	return fmt.Sprintf("Type error: %s", e.reason)
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

type BuiltinFnValue struct {
	name string
	fn   func([]Value) (Value, error)
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

func (c *Context) LoadFunc(name string, fn func([]Value) (Value, error)) {
	c.scope.put(name, BuiltinFnValue{
		name: name,
		fn:   fn,
	})
}

func (c *Context) LoadBuiltins() {
	c.LoadFunc("import", c.mgnImport)
	c.LoadFunc("type", c.mgnType)
	c.LoadFunc("string", c.mgnString)
	c.LoadFunc("int", c.mgnInt)
	c.LoadFunc("float", c.mgnFloat)
	c.LoadFunc("codepoint", c.mgnCodepoint)
	c.LoadFunc("char", c.mgnChar)
	c.LoadFunc("len", c.mgnLen)
	c.LoadFunc("print", c.mgnPrint)
	c.LoadFunc("keys", c.mgnKeys)
	c.LoadFunc("open", c.mgnOpen)
	c.LoadFunc("close", c.mgnClose)
	c.LoadFunc("read", c.mgnRead)
	c.LoadFunc("write", c.mgnWrite)
}

func errObj(message string) ObjectValue {
	return ObjectValue{
		"type":  AtomValue("error"),
		"error": MakeString(message),
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
		n, err := strconv.ParseInt(string(*arg), 10, 64)
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
		f, err := strconv.ParseFloat(string(*arg), 64)
		if err != nil {
			return null, nil
		}
		return FloatValue(f), nil
	default:
		return null, nil
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
	pathStr := string(*pathBytes)

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

	if imported, ok := c.importMap[filePath]; ok {
		return ObjectValue(imported.vars), nil
	}

	ctx := NewContext(path.Dir(filePath))
	ctx.importMap = c.importMap
	ctx.LoadBuiltins()
	_, err = ctx.Eval(file)
	if err != nil {
		return nil, err
	}

	c.importMap[filePath] = ctx.scope
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
	case "readwrite":
		flags = os.O_RDWR
	case "append":
		flags = os.O_RDWR | os.O_APPEND
	case "create":
		flags = os.O_RDWR | os.O_CREATE
	case "truncate":
		flags = os.O_RDWR | os.O_TRUNC
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
	c.fileMap[fd] = file
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

	file, ok := c.fileMap[uintptr(fdInt)]
	if !ok {
		return errObj(fmt.Sprintf("Unknown fd %d", fdInt)), nil
	}

	err := file.Close()
	if err != nil {
		return errObj(fmt.Sprintf("Could not close file: %s", err.Error())), nil
	}

	delete(c.fileMap, uintptr(fdInt))

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

	file, ok := c.fileMap[uintptr(fdInt)]
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

	file, ok := c.fileMap[uintptr(fdInt)]
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
