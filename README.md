# Magnolia

A friendly, expressive programming language.

## Syntax

Mgn, like [Ink](https://dotink.co), has automatic comma insertion at end of lines. This means if a comma can be inserted at the end of a line, it will automatically be inserted.

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
    identifier '<-' expr | // nonlocal update
    identifier ':=' expr |
    listLiteral ':=' expr |
    objectLiteral ':=' expr
)

propertyAccess := identifier ('.' identifier)+

unaryExpr := ('!' | '-') expr
binaryExpr := expr (+ - * / % ^ & | > < = >= <=) binaryExpr

prefixCall := expr '(' (expr ',')* ')'
infixCall := expr expr (
    expr |
    '(' (expr ',')* ')'
)?

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
    _ -> n * (n - 1 factorial)
}
```

```
// methods are emulated by infix notation
// for example, for lists
n times fn(i) std.println(i)
names join ', ' // join(names, ', ')
scores sum // sum(scores)
mgnFiles := fileNames std.filter fn(name) name endsWith? '.mgn'
	// mgnFiles := std.filter(fileNames, fn(name) { endsWith?(name, '.mgn') })
fn sum(xs...) xs std.reduce(fn(a, b) a + b, 0)
```

```
// "with" keyword just makes the last fn a callback as last arg
with fetch('some.url.com')
	fn(resp) with resp.json()
		fn(json) console.log(json)
```

```
// file read
with file := open('name.txt') fn(evt) if evt.type {
	:error -> std.println(evt.message)
	_ -> with read(file, 0, -1) fn(evt) {
		if evt.type {
			:error -> std.println(evt.message)
			_ -> std.printf('file data: {0}', evt.data)
		}
		close(file)
	}
}

// with stdlib
with std.readFile('name.txt') fn(file) if file {
    ? -> std.println('could not read file')
    _ -> std.println(file slice(0, file len))
}
```

```
// while loop
fn {
    std.println('reading file...')
} while fn { signal nil? }
```

