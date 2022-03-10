---
title: Designing a try/catch interop API for Oak
date: 2022-02-17T15:37:28-05:00
---

One of Oak's most useful features is its ability to be compiled to JavaScript using `oak build --web` to run on all the places JavaScript can run, like Node.js, web browsers, and even microcontrollers and games using [embedded](https://duktape.org/) [JS engine](https://bellard.org/quickjs/). When writing Oak programs that run in a JavaScript context, one problem that was unsolved until recently is how Oak interacts with exceptions that are thrown in JavaScript contexts.

Oak can call naturally into JavaScript functions, but if the JavaScript function throws an exception somewhere, the calling Oak code can't catch or react to it appropriately. When this happens in code that updates a web UI or handles a web request, this can be a problem. **Oak needs to be able to recover from and react appropriately to exceptions thrown in JavaScript**. So I tried to design an Oak API around JavaScript's exceptions.

In the design process, I had two main goals for the Oak exceptions API:

1. It should feel at home in an Oak program, especially with other error-handling primitives in Oak. Errors are values in Oak, usually using either a special [sentinel value](https://en.wikipedia.org/wiki/Sentinel_value) like `?` or an error object like `{ type: :error, error: ... }`. Errors from JavaScript contexts should feel similar to manage.
2. It shouldn't add new constructs to the Oak language. Though errors are important, interoperation with JavaScript exceptions is only relevant in the compile-to-JavaScript context. Adding error handling facilities in this narrow context shouldn't disrupt the rest of the Oak language.

To get started, I began with how JavaScript itself deals with exceptions: _checked exceptions_.

## Checked exceptions and `try`/`catch`

In checked exceptions, JavaScript code can set up a `try { }` block, which can rescue the program from thrown exceptions and move the program execution into a `catch` block.

```
try {
    somethingAwful()
} catch (e) {
    handleError(e)
}
```

This is quite a nice solution, and it's possible because the `try { } catch (e) { }` statement is a part of JavaScript's syntax. To introduce something similar in Oak would require inventing some similar novel syntax or otherwise doing something ugly with closures:

```oak
with try(fn {
    somethingAwful()
}) fn catch(e) {
    handleError(e)
}
```

That seems... okay, but certainly not great. My biggest objection to this kind of interface that emulated the "block-based" feeling of Java-style exceptions is that this obscures true control flow of code. In Oak, callbacks usually mean the callback function is either executed asynchronously, or somehow executed out-of-flow from the rest of the program, like in a loop in `std.loop`. But here, there's a very straightforward control flow: we run the main function, check it for errors, and if any errors occur, we handle it in the catch function. Using a callback to handle the error feels like a lot of machinery for a very confusing flow of control.

This problem gets worse if we want to use the return value from the function with a fallback value:

```oak
number := with try(fn {
    somethingAwful()
}) fn catch(e) {
    number <- fallbackNumber
}

doSomethingWith(number)
```

This looks very awkward compared to what a more idiomatic Oak program may read like:

```oak
result := somethingAwful()
number := if result.type {
    :error -> fallbackNumber
    _ -> result.value
}

doSomethingWith(number)
```

After a few of these sketches, it seemed clear to me that Oak should handle JavaScript exceptions in a way that _preserved a clear sense of control flow_ in the resulting code, and that _let error handling code compose well with other Oak syntax_ like `if` expressions.

## Lua to the rescue

Lua's documentation covers a [rather unique way of dealing with runtime errors in Lua](https://www.lua.org/pil/8.4.html), using two builtin functions `error` and `pcall`.

>If you need to handle errors in Lua, you should use the `pcall` function (_protected call_) to encapsulate your code.
>
>Suppose you want to run a piece of Lua code and to catch any error raised while running that code. Your first step is to encapsulate that piece of code in a function [...] Then, you call [it] with `pcall`.
>
>```
>if pcall(foo) then
>  -- no errors while running `foo'
>  ...
>else
>  -- `foo' raised an error: take appropriate actions
>  ...
>end
>```
>
>The `pcall` function calls its first argument in _protected mode_, so that it catches any errors while the function is running. If there are no errors, `pcall` returns **true**, plus any values returned by the call. Otherwise, it returns **false**, plus the error message.
>```
>local status, err = pcall(function () error({code=121}) end)
>print(err.code)  -->  121
>```

There were two appealing properties of Lua's method, using a special `pcall` function to invoke potentially error-throwing functions.

1. Control flow is easily preserved. Rather than calling a function directly, we call it with `pcall`, and nothing else changes.
2. Errors are treated as normal values, rather than special things floating around in the aether of the program's runtime.

Both of these seemed quite well-suited to Oak in my eyes. Further, I also liked that the `pcall` function marked a _clear boundary_ in the source code between JavaScript-style thrown exceptions and Oak-style error values. So, this became the basis for Oak's try/catch interop API.

## Oak's solution: the `try()` function

_For the rest of this section, when I say "in Oak", I refer specifically to Oak programs compiled to JavaScript and running in a JavaScript engine like Node.js._

In Oak, exceptions raised by JavaScript code are handled by "trapping" any thrown exceptions and turning them into error values. We can trap thrown JavaScript errors using the `try()` function, which is built-in.

Let's say we have a function, `somethingVeryIllegal()`, that returns some number but may instead throw an exception. If we call it directly in Oak, this exception may bubble up through the stack and crash our Oak program.

```oak
// if this throws, the whole program crashes!
number := somethingVeryIllegal()
```

Instead, we can call the function using `try`. The `try` function catches any potential exceptions thrown within the called function, and returns either an `:ok` or `:error` object.

```oak
result := try(somethingVeryIllegal)

// result if no exception is thrown
{
    type: :ok
    ok: 10 // or whatever somethingVeryIllegal returned
}

// result if an exception was thrown
{
    type: :error
    error: {
        message: 'Error message'
        stack: 'Error message\n [...] /loader:822:12)'
    }
}
```

The `try` function "wraps" the JavaScript exception and returns it as an object, which we can handle like any other Oak error object. Though we used an existing function here, we can use anonymous functions instead to simulate a try "block":

```oak
result := with try() fn {
    1 + iAmNotDefined
}

// result
{
    type: :error
    error: {
        message: 'iAmNotDefined is not defined'
        stack: 'ReferenceError: iAmNotDefined is not defined [...] at Function.Module._load (node:internal/modules/cjs/loader:822:12)'
    }
}
```

The way we deal with errors from a `try()` call fits well among other error handling code in Oak, I think.

```oak
result := try(somethingMaybeIllegal)
if result.type {
    :error -> reportErr(result.error)
    :ok -> processInput(result.ok)
}
```

This solution, using a `try()` function that wraps a potentially-throwing call, is clear about the chain of control flow, and returns error objects are familiar to Oak programmers from other built-in functions like `open` and `input`. It does all this without introducing any fundamentally new constructs into the language, and as a bonus, any call to the `try()` visibly marks the boundary between JavaScript-style error handling and Oak-style error handling.

There are still some open questions. The main unresolved problem I face today is how Oak's error-handling patterns and the `try()` function interacts with any exceptions that are thrown asynchronously, in a callback. Node.js, for example, will let programs catch any uncaught exceptions at the top level in the program (for logging or reporting purposes). Oak currently has no facility to handle asynchronous runtime errors, because errors in correct Oak programs are reported as values. How might Oak be able to handle these asynchronous runtime errors that bubble up from JavaScript? These questions, I think, are good topics for future investigations.

