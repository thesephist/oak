---
title: Oak and expressive power in programming languages
date: 2022-02-26T15:37:28-05:00
---

I've been interested in the idea of _expressiveness_ of programming languages since I read [_On the expressive power of programming languages_](https://doi.org/10.1016/0167-6423(91)90036-w) by Matthias Felleisen. Though I don't agree wholeheartedly with the way Felleisen co-opts the term _expressive power_ for his idea of capabilities in programming languages, I think it invites us to think more rigorously about expressiveness of languages.

_Expressiveness_ seems like a property universally sought after by programming language designers (Oak notwithstanding). Developers want to write programs in languages that are expressive. But what exactly does that quantify?

Intuitively, for me to say a language is _expressive_, I want to feel that its **representation of a program faithfully models my personal abstractions and theories** about the problems and ideas the program models. In this post, I want to make that intuition a bit more rigorous, and lightly discuss some ways in which Oak is expressive, and some ways in which it's not, both in general terms as well as specifically compared to its predecessor language [Ink](https://dotink.co/).

In the process, I want to show one way Felleisen's narrow definition of expressiveness falls short when we try to describe how programming languages feel to use day-to-day.

## Expressiveness as capability

Matthias Felleisen's definition of expressiveness is what I would call **expressiveness as capability**. For him in his paper, more expressive programming languages are _capable of expressing a wider set of programs_ than less expressive languages.

To explain this definition, Felleisen uses the example of syntax sugar -- syntax that can be trivially rewritten into a simpler or more primitive form, perhaps using something like a macro expansion. For example, in many C-like languages, the `+=` operator is syntactic sugar. `x += a` can be trivially rewritten to `x = x + a` without losing any meaning about the program's behavior. Further, this kind of rewriting is _local_ (doesn't interact with constructs outside of this particular expression), so the rewriting can be done with a macro (Felleisen calls this a _syntactic abstraction_).

Felleisen then defines two programming languages as _equally expressive_ if each language can be rewritten into the other with local macro expansions. Conversely, language A is _more expressive_ than language B if some programs in language A cannot be rewritten with local macro-expansions into language B, because it means language A contains syntactic constructs that aren't just "syntax sugar" for something in language B.

As an example, certain kinds of control flow mechanisms are impossible to express purely as syntactic rewrites. In a language with `goto`s, we cannot always rewrite control flows that jump across distant blocks of code using `goto` statements into more structured control flows using loops and `if` statements. Take this program in a hypothetical variant of Oak with `goto`s as an example:

```oak
fn doSomethingDangerous(poison) {
    if poison.brew() {
        :missingIngredient -> goto tryAgain
        :exploded -> goto cleanUp
        :error -> goto handleError
        _ -> poison.serve()
    }

    tryAgain:
        poison.addIngredients()
        // also runs cleanUp
    cleanUp:
        poison.reset()
        // also runs handleError
    handleError:
        print('Error.')
}
```

There is no way to rewrite parts of this program that use the `goto` keyword with only simple, local changes so that we don't need that keyword. This kind of goto-based control flow is common in low-level error handling code in C-like languages with `goto` support, but there isn't a way to express the `goto` syntax in terms of other syntactic abstractions we have, like loops or conditionals. This suggests that languages with `goto` statements are more expressive -- those programming languages have syntax that are more capable, in a sense. The control flow construct `goto` _adds expressive power_ to languages.

As a counterexample, many languages including Oak have _pipeline operators_. Usually, this operator provides a way to express function composition (passing the return values of one function into another to build a "pipeline") so that it's easier to read. For example, in Oak, we may write a program to find unique words in some text like this.

```oak
text |>
    lower() |>
    split(' ') |>
    sort!() |>
    uniq()
```

The `|>` operator is syntactic sugar -- the program `a |> b()` can be trivially rewritten as `b(a)`. So this above program is equivalent to writing

```oak
uniq(sort!(split(lower(text), ' ')))
```

As a result, Oak's pipeline operator doesn't add any expressive power (by Felleisen's definition) to Oak.

This definition of expressiveness in programming languages establishes a hierarchy between different languages, where some languages have larger syntactic vocabularies than others, capable of expressing a broader range of control flow, types, abstractions, and composition between ideas. It's a useful way to study languages, because the power we gain with completely new constructs in languages is real. Some examples of such power besides control flow, which I mentioned above, are type systems, ways of organizing code (functions, classes), generators, and support for asynchronous programming (`async` and `await`, coroutines).

When viewed through this lens, Oak gains little to no expressive power compared to Ink or other minimal scripting languages. Many operators added in Oak are syntactic sugar -- indeed, they're decomposed into their more "primitive" versions during the parse step in the interpreter, before any Oak code runs. Nonetheless, having written thousands of lines of code in both Ink and Oak, Oak programs _do_ feel more expressive, in the colloquial sense. To understand why, I think we need a different way to think about expressive power in programming languages, one based on how humans read and write programs day-to-day.

## Expressiveness as expression of theory

Peter Naur (of the [Backus-Naur form](https://en.wikipedia.org/wiki/Backus%E2%80%93Naur_form)) presents an idea in [Programming as theory building](https://doi.org/10.1016/0165-6074(85)90032-8) that the act of programming is less about laying down instructions for the computer in text, but more about the programmer _building a theory about the problem they are solving_, and the role the wider world plays in it. This theory the programmer builds contains their understanding of not only how the program works, but _why_ it's designed the way it is. Within the programmer's theory, they can explain why the program is concerned with certain parts of the problem but not others, and what has to change when basic assumptions about the problem change. More importantly, Naur asserts that this theory the programmer builds in the course of programming isn't something contained within the source code, nor the documentation, but something carried within the programmer's mind (or the minds of a team) and passed along like folklore.

>One way of stating the main point I want to make is that programming in this sense primarily must be the programmers' building up knowledge of a certain kind, knowledge taken to be basically the programmers' immediate possession, any documentation being an auxiliary, secondary product.
>
>The conclusion seems inescapable that at least with certain kinds of large programs, the continued adaptation, modification, and correction of errors in them, is essentially dependent on a certain kind of knowledge possessed by a group of programmers who are closely and continuously connected with them.

A critical part of this understanding that programmers have over their programs is a model of the world, and how the program represents it:

>The programmer having the theory of the program can explain how the solution relates to the affairs of the world that it helps to handle [...] the programmer must be able to explain, for each part of the program text and for each of its overall structural characteristics, what aspect or activity of the world is matched by it.

Naur finds that the act of programming is primarily about building a structured understanding of the world and the problem at hand. He also asserts many times in the essay that this understanding, or "theory", often stays confined within the minds of the original designers of a program, failing to be passed on effectively to future readers and maintainers of the program. If we follow these ideas, one desirable property of a good programming language must be its **ability to faithfully mirror the programmer's model of the problem at hand, and its place in the world**. I think this is another viable definition for _expressiveness_ in programming languages: how faithfully can the program represent the programmer's theory of the program within the language's syntax?

Expressiveness defined this way is actually a desirable quality of [notations](https://thesephist.com/posts/notation/) in general, programming languages being a specific kind of notation that also happens to be executable on a computer.

Using this definition of expressive power, syntactic sugar takes on a much different role than we found in the first section: syntax sugar can often make a programming language _more expressive_, because it may be able to encode some understanding of the program the programmer has without changing what the program does.

Let's return to the unique-words pipeline example from the first section. When I imagine the problem, my mental model of the solution is as a pipeline or an assembly line: first, we normalize the capitalization of all words in the string; then we split the text into words; then we `sort` and `uniq` to get a list of unique words in the list. I can encode my model of this problem much more faithfully in the program using Oak's pipeline operator (`|>`) than without it, even though the resulting program is ultimately identical in behavior.

The beauty of syntactic sugar is that it can express _the programmer's intent_ without modifying the program's behavior. It offers the programmer a way to communicate their theory of the world to future readers and programmers, right within the source code itself.

Viewed through this lens, Oak gained many different syntax constructs beyond what Ink had. As a result, Oak programs are consistently better at expressing the programmer's intent than Ink programs -- I find it easier to read Oak programs and understand why it works the way it does. Beyond the pipeline operator, Oak has:

The **push operator** (`<<`). In Ink, we added new items to lists or strings by assigning to a new key. But because pushing new items onto lists or strings is such a common operation, having a new operator specifically for this concept makes a lot of sense.

**Atoms (`:like_this`)**. Atoms in Oak are immutable strings that appear as if they were atomic tokens. Atoms are used pervasively in Oak to represent types or kinds of data, like results of asynchronous operations (`:data`, `:error`, `:resp`) or runtime types of values (`:int`, `:string`, `:function`). Atoms are used in Oak for many of the use cases for which Ink used strings, but by using atoms, Oak programs can signal to the reader that these values aren't user-generated or arbitrary, but are atomic labels that denote some part of the program's design.

The **`with` keyword**. The `with` keyword is pure syntactic sugar that allows Oak programmers to invoke functions that take callbacks in a nicer way, by placing the callback outside of the function argument list. Rather than writing

```oak
fs.readFile('data.txt', fn(file) {
    ...
})
```

we can write

```oak
with fs.readFile('data.txt') fn(file) {
    ...
}
```

Though these two notations are technically equivalent, and even visually quite similar, I like the way the `with` keyword visually separates information passed to `fs.readFile` (the file path) from _what happens after it_, in the callback function. It's an expression of _intent_, more than behavior. In the future, the `with` keyword may do something more (I'm a big fan of the way `with` statements work in Python), but even as pure syntactic sugar, I think this is a nice change.

As with all design decisions, expressiveness has a tradeoff -- more expressive programming languages tend to have a larger "vocabulary" size, which makes the language more complex. So Oak sometimes foregoes new specific syntax in favor of simplicity. For example, Oak doesn't have structured looping constructs like `for` or `while` loops in the language, instead relying on standard library functions like `std.each` and `std.loop` to fill in those gaps. So far, this has felt like a good choice. Rather than jumping straight to a basic loop, Oak programmers are guided to find an iteration abstraction that better describes their particular use case.

---

Programmers want to use expressive languages because of both of the reasons above -- the ability to express new programming constructs like control flows or asynchronicity, and the ability to write programs that better describe their mental models of the problem at hand. Sometimes, these two perspectives on expressiveness are at odds. A new bit of syntax may not add fundamentally new capabilities to the language, but it may help the programmer _communicate their mental model_ more faithfully. When we design programming languages (and any notation, for that matter), I think it's important to hold both of these standards in mind to build tools that are as ergonomic and they are capable.

