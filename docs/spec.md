# Oak programming language

This is a work-in-progress rough draft of things that will end up in a rough informal language specification.

## Syntax

Oak, like [Ink](https://dotink.co), has automatic comma insertion at end of lines. This means if a comma can be inserted at the end of a line, it will automatically be inserted.

```
program := expr*

expr := literal | identifier |
    assignment |
    propertyAccess |
    unaryExpr | binaryExpr |
    prefixCall | infixCall |
    ifExpr | withExpr |
    block

literal := nullLiteral |
    numberLiteral | stringLiteral | atomLiteral | booleanLiteral |
    listLiteral | objectLiteral |
    fnLiterael

nullLiteral := '?'
numberLiteral := \d+ | \d* '.' \d+
stringLiteral := // single quoted string with standard escape sequences + \x00 syntax
atomLiteral := ':' + identifier
booleanLiteral := 'true' | 'false'
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

ifExpr := 'if' expr '{' ifClause* '}'
ifClause := expr '->' expr ','

withExpr := 'with' prefixCall fnLiteral

block := '{' expr* '}' | '(' expr* ')'
```

### AST node types

```
nullLiteral
stringLiteral
numberLiteral
booleanLiteral
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

## Builtin functions

```
-- language
import(path)
string(x)
int(x)
float(x)
atom(c)
codepoint(c)
char(n)
type(x)
len(x)
keys(x)

-- os
args()
env()
time() // returns float
nanotime() // returns int
exit(code)
rand()
wait(duration)
exec(path, args, stdin) // returns stdout, stderr, end events

---- I/O interfaces
input()
print()
ls(path)
mkdir(path)
rm(path)
stat(path)
open(path, flags, perm)
close(fd)
read(fd, offset, length)
write(fd, offset, data)
close := listen(host, handler)
req(data)

-- math
sin(n)
cos(n)
tan(n)
asin(n)
acos(n)
atan(n)
pow(b, n)
log(b, n)
```

## Code samples

```
// hello world
std.println('Hello, World!')

// some math
sq := fn(n) n * n
fn sq(n) n * n
fn sq(n) { n * n } // equivalent

// side-effecting functions
fn say() { std.println('Hi!') }
// if no arguments, () is optiona
fn { std.println('Hi!') }

// factorial
fn factorial(n) if n <= 1 {
	true -> 1
    _ -> n * factorial(n - 1)
}
```

```
// methods are emulated by pipe notation
n |> times(fn(i) std.println(i))
names |> join(', ')
scores |> sum()
oakFiles := fileNames |> filter(fn(name) name |> endsWith?('.oak')) 
fn sum(xs...) xs |> reduce(0, fn(a, b) a + b)
```

```
// "with" keyword just makes the last fn a callback as last arg
with fetch('some.url.com')
	fn(resp) with resp.json()
    fn(json) console.log(json)
```

```
// file read
with open('name.txt') fn(evt) if evt.type {
	:error -> std.println(evt.message)
	_ -> with read(evt.fd, 0, -1) fn(evt) {
		if evt.type {
			:error -> std.println(evt.message)
			_ -> std.printf('file data: {0}\n', evt.data)
		}
		close(evt.fd)
	}
}

// with stdlib
with std.readFile('name.txt') fn(file) if file {
    ? -> std.println('could not read file')
    _ -> std.println(file |> slice(0))
}
```

