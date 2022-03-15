---
title: Why Oak?
date: 2021-09-08T15:37:28-05:00
---

In mid-2019, I wrote my first programming language as a hobby learning project. It was called [Ink](https://dotink.co/), and was a very simple dynamically typed scripting language in the spirit of JavaScript and Lua. For the next two years, I built [lots and lots of projects and tools](https://dotink.co/docs/projects/) with Ink, and learned a lot about both building and using my own programming language.

My biggest takeaway was this:

>When we design the vocabulary with which we build our tools and projects, we can move much faster, build more fearlessly, and gain a deeper understanding of the systems we build.

This post, in addition to being an introduction to Oak, is my attempt to unravel this lesson and invite you to adopt it in your work. But before we get into that, I want you to meet Oak — a simple, expressive, friendly programming language I designed for building my personal projects and tools.

```oak
std := import('std')

std.println('Hello, World!')
```

## Inventing Oak

Oak came to be as an evolution of Ink, which was my first toy programming language. It was a dynamically typed scripting language that was bare (you might say _minimal_) in features, but had the essentials. I really enjoyed writing programs in Ink, but it had a handful of shortcomings and misfeatures that became more and more obvious over time as I used it.

Oak is a sequel to Ink that tries to correct many of these early mistakes, while being [more expressive](/posts/expressive/) with much more robust first-party tooling support. Oak is intentionally not "Ink 2.0" — it's not a simple upgrade or bug fix, but a different language with many well-considered updates in the details.

I was never really proud to have other developers try to program in Ink, because it was a rough learning project; Oak, on the other hand, is something I can (with some chagrin) hand off to a fellow dev without being crushed under the weight of my own self-doubt.

### What Ink got right, and what it got wrong

Ink was designed to be very minimal, partly because I prefer simplicity and partly because I didn't want to commit myself to building lots of language features for my first toy language. It had few basic data types, no classes or inheritance mechanism, and a small set of built-in operators. Every blocking operation (like reading files, working with timers) was fully asynchronous. In these ways, Ink was very "pure", value-oriented, fully async, without any hidden details to learn. But that simplicity and purity came at the cost of ergonomics and expressiveness. As I used Ink, I began to convince myself that some features and complexity that was left out of Ink _did_ earn their places in even a small, minimal language. Let me walk you through a handful of them.

#### Richer set of operators

The most obvious kind of missing feature was operators. Oak has a richer set of operators than Ink, for things that I wanted to express in Ink programs but couldn't do so elegantly:

The **nonlocal assignment operator** `<-` binds a value to a variable _without declaring the variable in the local scope_, meaning it can update values that are in a parent scope of some block or function. This operation wasn't possible in Ink. This operator is similar to `:=`, with one difference:

```oak
n := 10
m := 20
{
    n <- 30
    m := 40 // m re-declared in this scope
}
n // reassigned to 30
m // still 20
```

The **push operator** `<<` pushes values onto the end of an iterable value (a string or a list), and returns the changed string or list. This was possible in Ink, but had a much clunkier notation:

```oak
// Ink
list.(len(list)) := item
// Oak
list << item
```

The **pipe operator** `|>` takes a value on the left and makes it the first argument to a function call on the right.

```oak
2 |> double()
// ... really means ...
double(2)
```

This operator helps express chained function calls or "streams" of operations, which are common in functional programming patterns.

```oak
// print 2n for every prime n in range [0, 10)
range(10) |> filter(prime?) |>
    each(double) |> each(println)

// adding numbers
fn add(a, b) a + b
10 |> add(20) |> add(3) // 33
```

#### Better operator precedence rules

Ink's syntax also suffered from bad operator precedence, which I didn't have a chance to go back and fix, because I was then just learning how to write parsers. In Ink, the dot `.` used for property access had lower precedence than function calls, which meant an expression like `stack.pop()` parsed as `(stack) . (pop())`, when it should really have been parsed as `(stack.pop) ()`. Oak corrects this mistake, so that the object-oriented programming style of calling methods defined on objects is much easier on the eyes.

#### Optional and variadic parameters

Lastly, Ink didn't support optional function parameters, nor variadic functions, which meant that every function call had to be called with precisely the correct number of arguments. I didn't expect this to be so limiting, but over time, I discovered there are many, many places where a parameter list of varying length is useful:

- Without optional parameters, I was forced to sometimes write multiple versions of the same function, like `sort()` and `sortBy(predicate)`, one of which took some property to sort a list by.
- It's often more natural for a function to take a variadic number of arguments than to take a list, like `math.sum(10, 20, 30, 40)`, but in Ink, I couldn't express it. I had to write `(math.sum)([10, 20, 30, 40])`.
- Without variadic functions, anytime I wanted to give a function a varying number of arguments, I had to use a list. For example, `fmt.format('{{0}} {{1}}', 10, 12)` has a much more clunky signature in Ink: `(fmt.format)('{{0}} {{1}}', [10, 12])`.

#### The good ideas that survived

Despite these flaws, there were many ideas in Ink that I felt I got right, and I wanted to adopt into Oak.

I thought an emphasis on programming in a functional style was good, and I wanted to keep language features that encouraged functional programming patterns. So I kept:

- Deep equality, over identity. In other words, objects like `{a: 'hi'}` are compared not by reference, but by whether their internal values were fully equal.
- Iteration using recursion, rather than using structured looping keywords like `for` or `while`. In practice, most loops are written using standard library functions like `std.each` or `std.loop`.
- Values that are just values, rather than instances of classes or other complex constructs that can easily bloat into a ball of methods and internal state. I'm a big proponent of the idea that [simple values are easier to program with than objects that combine state and behavior](https://youtu.be/-6BsiVyC1kM). State lives in values; behavior lives in functions.

There were also a few more novel ideas in Ink that proved their worth. I expanded on these further when designing Oak.

- The [_match operator_ `::`](https://dotink.co/docs/overview/#match-expressions), a kind of a versatile switch-case, was the only branching/conditional construct in Ink. It worked so well that Oak adopted the idea, but using a keyword `if` instead. Oak's `if` expression also has a few shorthand forms to make it easier to use, like `if condition? -> doSomething()`.
- The _empty name_ `_` was equal to any other value (matched every other value; `x = _` for any and all `x`). This simple idea made writing complex `if` expressions and other structural comparisions easy: for example, `result = { type: :error, error: _ }` checks if a result value is structured like an error value.
- Asynchronous programming with callbacks, like in JavaScript, was a good idea. But sometimes, I wanted an escape hatch to write synchronous functions for reading and writing files or waiting on a timer. Oak improves on this by offering both synchronous and asynchronous versions of all the blocking built-in functions, and light syntax support with the `with` syntax sugar for callbacks.

## Inventing _with_ Oak

Oak is a general-purpose programming language, but it's designed (and continuing to be developed) for the main use case of building personal projects and tools. Beyond the obvious pleasure of building projects with my own programming language, I think using a small language that I understand and control fully to build projects has had some unexpected upsides.

### Oak all the way down

Andreas Kling, who created [SerenityOS](https://github.com/SerenityOS/serenity), describes the virtue of owning the "full stack" of software in his [podcast interview on Corecursive](https://corecursive.com/serenity-os-with-andreas-kling/):

>**Adam [interviewer]**: One thing he picked up at Apple was a style of development that’s a bit different than what I’m used to. A lot of development today for me seems to be gluing various components together into a working system. But at Apple, everything is in-house. The web browser you use, the system calls you make, maybe even the programming language you use if you’re using Swift. They’re not black boxes, they’re just something made for you by one of your colleagues. You can ask questions, you can make improvements. It’s all just code there in source control.
>
>**Andreas**: I still feel that nobody really does that better than Apple. They control the whole stack, and they really take advantage of that. Especially lately with putting out their own CPUs and everything now as well. That’s been really awesome. And I enjoyed learning from that environment what is really possible if you control more of the stack.

I think Andreas makes a valuable observation here, that **owning and understanding further up and down your software stack can let you move faster and build things others couldn't have built**. Of course, there is a reason this isn't common in the industry -- for most software products and companies, it's simply not practical to build and own the full stack behind complex and ever-evolving products. But I think that calculus changes for personal projects or projects that we build to learn new ideas, and I think Andreas recognizes that as well, in his experience building SerenityOS from scratch as this kind of a "full stack" system:

>Everything is just a piece of code that somebody writes. And if we just make all those pieces of code and stack them up, it’s going to work. I had no illusions about how an operating system looks once it is put together and works. Now I didn’t know how to get there, but I reasoned that if you just start building these components one by one, eventually you’ll have the full stack and it will just gel together. So that’s what I started doing.

In my case, there are two very clear benefits of build personal tools and projects in my own language:

First, **my programs are only nearly as complex as my problems, and don't inherit the incidental complexity that comes from "generic" software design**. For example, Oak's `http` standard library has a server that can parse different routes and parameters in URLs and query strings (you can see it in action [here](https://github.com/thesephist/stream/blob/6f045ec309366f9e32dd5b764cbe7477edeff73f/src/main.oak#L31-L209)). A general-purpose library that does something similar is [Express](https://expressjs.com/) in the Node.js ecosystem or [Gorilla](https://github.com/gorilla/mux) in the Go ecosystem. Both of these libraries are powerful and extensible, but they're also too complex for me to understand fully. By comparison, I wrote the entire `http` library in Oak, all 300 lines, by myself, and I understand what each line is doing perfectly. If I need to dig deeper, I also know how the interpreter that runs that library works, because I wrote it! This deep level of understanding means if there's ever a bug or unexpected behavior, I can immediately dig into the origin of the issue and debug more effectively, than if parts of my stack belonged to someone else.

**I learn much more much faster about computers**, because whenever I write a new project that touches a new domain like operating system signals or WebSockets or date/time parsing or Markdown rendering, I come to understand how everything works much deeper than if I had simply pulled in a library from NPM.

### The power of a self-hosted toolchain

While I was building projects with Ink in the early days, I sort of stumbled into a path of building a lot of language tooling in Ink itself -- I built a syntax highlighter, a code formatter, a compiler, an assembler, an online programming environment, and so on, all for the Ink language, all in Ink itself. While I ended up on this path without an agenda in the beginning, my experience here convinced me that the right way to build language tools (perhaps with the exception of the main compiler/interpreter itself) was to _self-host_. To do it in the language itself. Without this hindsight, Ink's self-hosted tooling was scattered and disorganized. The compiler used a different parser than the code formatter, which lived in a different codebase than the syntax highlighter. But armed with this perspective, I sought to build a rich set of language tooling into Oak itself from the start.

To support this effort, I built into Oak a standard library module for working with Oak syntax and Oak programs. The `syntax` standard library module contains an "official" tokenizer, parser, and code formatter for Oak syntax, which is used by the syntax highlighter, code formatter, bundler, and compiler that ship with the Oak executable. I borrowed this idea from Go and Rust, which both have official libraries for working with the language, written in the language itself. For Go, this lives in [golang.org/x/tools](https://pkg.go.dev/golang.org/x/tools). Similar libraries for Rust can be found in various components of the `rustc` compiler, like [the `rustc_ast` crate](https://github.com/rust-lang/rust/tree/8073a88f35728289ef535cca5cf13302faba5972/compiler/rustc_ast) for example.

When all of the supporting tools for a programming language are built together on the same foundation, it leads to a programming _environment_ that feels much more coherent and integrated together. There's one right tool for any given task, which is designed exactly for that task, and every tool ships with the programming language itself (in this case, the `oak` executable). I think this coherence makes me more productive and less scattered when I program with Oak. Go, Deno, and Zig also ship with tightly integrated and coherent tools, and I think deliver the same benefit.

### A personal vocabulary for computing

The programming language we use to build projects and tools define the _vocabulary_ with which we speak these ideas into existence. Because most languages are extremely general-purpose, if I know my use case exactly, it turns out I can get a huge amount of leverage and learning by building a custom language and surrounding ecosystem for my exact purpose.

With the language, libraries, and tools tailored to the way I want to work, I rarely have to fight the language to understand what went wrong or why something is behaving unexpectedly. The vocabulary of Oak lines up well with how I want to describe my problems and designs, working at the level of abstraction that feels right for me, with the concepts I'm familiar with. And unlike programming in Ink, where I had to often work around design flaws or deficiencies, Oak feels just right under my fingers. I'm looking forward to building an ever wider universe of interesting projects and experiments on this new foundation.

