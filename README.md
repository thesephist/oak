# Oak 🌳

[![Build Status](https://travis-ci.com/thesephist/oak.svg?branch=main)](https://travis-ci.com/thesephist/oak)

**Oak** is an expressive, dynamically typed programming language. It takes the best parts of my experience with [Ink](https://dotink.co/), and adds what I missed and removes what didn't work to get a language that feels just as small and simple, but much more ergonomic and capable.

Here's an example Oak program.

```js
std := import('std')

fn fizzbuzz(n) if [n % 3, n % 5] {
    [0, 0] -> 'FizzBuzz'
    [0, _] -> 'Fizz'
    [_, 0] -> 'Buzz'
    _ -> string(n)
}

std.range(1, 101) |> std.each(fn(n) {
    std.println(fizzbuzz(n))
})
```

Oak has good support for asynchronous I/O. Here's how you read a file and print it.

```js
std := import('std')
fs := import('fs')

with fs.readFile('./file.txt') fn(file) if file {
    ? -> std.println('Could not read file!')
    _ -> print(file)
}
```

Oak also has a pragmatic standard library that comes built into the `oak` executable. For example, there's a built-in HTTP server and router in the `http` library.

```js
std := import('std')
fmt := import('fmt')
http := import('http')

server := http.Server()
with server.route('/hello/:name') fn(params) {
    fn(req, end) if req.method {
        'GET' -> end({
            status: 200
            body: fmt.format('Hello, {{ 0 }}!'
                std.default(params.name, 'World'))
        })
        _ -> end(http.MethodNotAllowed)
    }
}
server.start(9999)
```

## Overview

Oak has 7 primitive and 3 complex types.

```js
?        // null, also "()"
_        // "empty" value, equal to anything
1, 2, 3  // integers
3.14     // floats
true     // booleans
'hello'  // strings
:error   // atoms

[1, :number]    // list
{ a: 'hello' }  // objects
fn(a, b) a + b  // functions
```

These types mostly behave as you'd expect. Some notable details:

- There is no implicit type casting between any types, except during arithmetic operations when ints may be cast up to floats.
- Both ints and floats are full 64-bit.
- Strings are mutable byte arrays, also used for arbitrary data storage in memory, like in Lua. For immutable strings, use atoms.
- Lists are backed by a vector data structure -- appending and indexing is cheap, but cloning is not
- For lists and objects, equality is defined as deep equality. There is no identity equality in Oak.

We define a function in Oak with the `fn` keyword. A name is optional, and if given, will define that function in that scope. If there are no arguments, the `()` may be omitted.

```js
fn double(n) 2 * n
fn speak {
    println('Hello!')
}
```

Besides the normal set of arithmetic operators, Oak has a few strange operators.

- The **assignment operator** `:=` binds values on the right side to names on the left, potentially by destructuring an object or list. For example:

    ```js
    a := 1              // a is 1
    [b, c] := [2, 3]    // b is 2, c is 3
    d := double(a)      // d is 2
    ```
- The **nonlocal assignment operator** `<-` binds values on the right side to names on the left, but only when those variables already exist. If the variable doesn't exist in the current scope, the operator ascends up parent scopes until it reaches the global scope to find the last scope where that name was bound.

    ```js
    n := 10
    m := 20
    {
        n <- 30
        m := 40
    }
    n // 30
    m // 20
    ```
- The **push operator** `<<` pushes values onto the end of a string or a list, mutating it, and returns the changed string or list.

    ```js
    str := 'Hello '
    str << 'World!' // 'Hello World!'

    list := [1, 2, 3]
    list << 4
    list << 5 << 6 // [1, 2, 3, 4, 5, 6]
    ```
- The **pipe operator** `|>`, which takes a value on the left and makes it the first argument to a function call on the right.

    ```js
    // print 2n for every prime n in range [0, 10)
    range(10) |> filter(prime?) |>
        each(double) |> each(println)

    // adding numbers
    fn add(a, b) a + b
    10 |> add(20) |> add(3) // 33
    ```

Oak uses one main construct for control flow -- the `if` match expression. Unlike a traditional `if` expression, which can only test for truthy and falsy values, Oak's `if` acts like a sophisticated switch-case, comparing values until the right match is reached.

```js
fn pluralize(word, count) if count {
    1 -> word
    2 -> 'a pair of ' + word
    _ -> word + 's'
}
```

This match expression, combined with safe tail recursion, makes Oak Turing-complete.

Lastly, because callback-based asynchronous concurrency is common in Oak, there's special syntax sugar, the `with` expression, to help. The `with` syntax sugar de-sugars like this.

```js
with readFile('./path') fn(file) {
    println(file)
}

// desugars to
readFile('./path', fn(file) {
    println(file)
})
```

For a more detailed description of the language, see the [work-in-progress language spec](docs/spec.md).

## Development

Oak (ab)uses GNU Make to run development workflows and tasks.

- `make run` compiles and runs the Oak binary, which opens an interactive REPL
- `make tests` or `make t` runs the Go tes suite for the Oak language and interpreter
- `make test-oak` or `make tk` runs the Oak test suite, which tests the standard libraries
- `make install` installs the Oak interpreter on your `$GOPATH` as `oak`, and re-installs Oak's vim syntax file

To try Oak by building from source, clone the repository and run `make install` (or simply `go build`).

