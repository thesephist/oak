# Oak Programming Language Documentation

## Syntax

Oak, like [Ink](https://dotink.co), has automatic comma insertion at end of lines. This means if a comma can be inserted at the end of a line, it will automatically be inserted.

```go
program := expr*

expr := literal | identifier |
    assignment |
    propertyAccess |
    unaryExpr | binaryExpr |
    prefixCall | infixCall |
    ifExpr | withExpr |
    block

literal := nullLiteral |
    numberLiteral | stringLiteral | atomLiteral | boolLiteral |
    listLiteral | objectLiteral |
    fnLiterael

nullLiteral := '?'
numberLiteral := \d+ | \d* '.' \d+
stringLiteral := // single quoted string with standard escape sequences + \x00 syntax
atomLiteral := ':' + identifier
boolLiteral := 'true' | 'false'
listLiteral := '[' ( expr ',' )* ']' // last comma optional
objectLiteral := '{' ( expr ':' expr ',' )* '}' // last comma optional
fnLiteral := 'fn' '(' ( identifier ',' )* (identifier '...')? ')' expr

identifier := \w_ (\w\d_?!)* | _

assignment := (
    identifier [':=' '<-'] expr |
    listLiteral [':=' '<-'] expr |
    objectLiteral [':=' '<-'] expr
)

propertyAccess := identifier ('.' identifier)+

unaryExpr := ('!' | '-') expr
binaryExpr := expr (+ - * / % ^ & | > < = >= <= <<) binaryExpr

prefixCall := expr '(' (expr ',')* ')'
infixCall := expr '|>' prefixCall

ifExpr := 'if' expr? '{' ifClause* '}'
ifClause := expr '->' expr ','

withExpr := 'with' prefixCall fnLiteral

block := '{' expr+ '}' | '(' expr* ')'
```

### AST node types

```c
nullLiteral
stringLiteral
numberLiteral
boolLiteral
atomLiteral
listLiteral
objectLiteral
fnLiteral
identifier
assignment
propertyAccess
unaryExpr
binaryExpr
fnCall
ifExpr
block
```

## Language Functions

- `import(path)`: Imports a module located at the specified `path`.
- `string(x)`: Converts the argument `x` to a string.
- `int(x)`: Converts the argument `x` to an integer.
- `float(x)`: Converts the argument `x` to a floating-point number.
- `atom(c)`: Creates an atom with the specified character `c`.
- `codepoint(c)`: Returns the Unicode code point of the character `c`.
- `char(n)`: Converts the Unicode code point `n` to a character.
- `type(x)`: Returns the type of the argument `x`.
- `len(x)`: Returns the length of the argument `x`.
- `keys(x)`: Returns an array of keys of the argument `x`.

## OS Functions

- `args()`: Returns command-line arguments as an array of strings.
- `env()`: Returns the environment variables as an object.
- `time()`: Returns the current time as a float.
- `nanotime()`: Returns the current time in nanoseconds as an integer.
- `exit(code)`: Exits the program with the specified exit code.
- `rand()`: Generates a random floating-point number between 0 and 1.
- `srand(length)`: Seeds the random number generator with the specified length.
- `wait(duration)`: Pauses the program execution for the specified duration.
- `exec(path, args, stdin)`: Executes a command specified by `path` with the given `args` and optional standard input `stdin`. Returns stdout, stderr, and end events.

## I/O Interfaces

- `input()`: Reads input from the standard input.
- `print()`: Writes output to the standard output.
- `ls(path)`: Lists files and directories in the specified path.
- `mkdir(path)`: Creates a directory at the specified path.
- `rm(path)`: Removes the file or directory at the specified path.
- `stat(path)`: Retrieves file or directory information at the specified path.
- `open(path, flags, perm)`: Opens a file at the specified path with the given flags and permissions.
- `close(fd)`: Closes the file descriptor `fd`.
- `read(fd, offset, length)`: Reads data from the file descriptor `fd` starting at the specified `offset` and reading `length` bytes.
- `write(fd, offset, data)`: Writes data to the file descriptor `fd` starting at the specified `offset`.
- `close := listen(host, handler)`: Listens for incoming connections on the specified `host` and handles them with the provided `handler` function.
- `req(data)`: Sends an HTTP request with the provided data.
  
  ```go
  // Req syntax:
  // ---
  
  req({
    url: ''
    method: 'GET'
    headers: {}
    body: _
  })
  ```
