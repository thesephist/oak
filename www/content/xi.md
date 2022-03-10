---
title: Xi: thinking different with concatenative programming languages
date: 2021-09-23T15:37:28-05:00
---

[**Xi**](https://github.com/thesephist/xi) (pronounced _Zai_) is a little stack-based concatenative language, written in Oak and using Oak types and semantics. I wrote Xi over the 2021 Labor Day weekend as a learning exercise to understand how stack languages like [Forth](https://en.wikipedia.org/wiki/Forth_(programming_language)) and [Factor](https://factorcode.org/) work and why they're interesting.

Before diving in, here's a glimpse of what Xi programs look like.

```oak
// print factorials of every number up to 10
factorial : nat prod
10 ( ++ factorial print ) each-integer

// Fibonacci sequence up to fib(25)
(fib) : dup 2 < ( drop swap drop ) ( ( swap over + ) dip -- (fib) ) if
fib : 1 1 rot (fib)
25 ( fib print ) each-integer
```

Xi is modeled mainly after the concatenative language Factor, but Xi's implementation is neither complete nor robust -- there's basically no error handling, for example, and Xi is not meant to be a faithful re-implementation of Factor. It should run correct programs correctly, but will often fail catastrophically on bad input. Nonetheless, building Xi and using it to write some basic programs has been an eye-opening experience that taught me a completely different way of structuring programs. This blog is my attempt to share some of those insights with you, but as with many interesting topics in programming languages, the best way to get a feel for why concatenative programming is interesting is to try it yourself. If you want to give it a shot yourself, you'll find some additional resources at the end of this post.

## Stack programming

All Xi programs operate on a single global data structure called the _stack_ (some sources call it the _data stack_, probably to avoid confusion with the function _call stack_). The stack contains all the values that a Xi program has access to at any given moment, in a [stack](https://en.wikipedia.org/wiki/Stack_(abstract_data_type)) data structure. Xi is dynamically typed, and works with the same basic values as Oak (like `float`, `string`, `bool`, `list`), the only difference being that Xi represents all numbers with `float`s for simplicity. So the Xi data stack is populated with these values.

A Xi program starts with an empty stack.

```oak
< >
```

Like other concatenative programming languages, each statement (line) in a Xi program is a sequence of _words_, where each word manipulates the data stack in some way, usually by moving and changing a few values at the top of the stack. Literal values like numbers and strings simply move those values onto the stack.

For example, the word `+` pops the top two values off the stack, adds them together, and pushes the sum back on the stack.

```oak
1
// stack: < 1 >
2 10
// stack: < 1 2 10 >
+
// stack: < 1 12 >
+
// stack: < 13 >
```

We can also write these words all next to each other, and have the same program.

```oak
1 2 10 + +
// stack: < 13 >
```

This is where the name _concatenative_ language comes from -- putting words next to each other composes those functions together in a predictable way.

Sometimes, we need to shuffle some items in the stack around to work on the right values without doing any other computation. These are called _stack shuffling_ words. Xi provides 4 basic ones -- `dup`, `dip`, `drop`, and `swap` -- from which more complex words can be defined:

```oak
2 dup
// stack: < 2 2 > — duplicates the top value

1 2 3 ( + ) dip
// stack: < 3 3 > — runs a quotation (words inside `( ... )`) underneath the
// topmost value on the stack

1 2 drop
// stack: < 1 > — simply drops the topmost value on the stack

10 20 swap
// stack: < 20 10 > — swaps the top 2 values' places on the stack
```

Keeping all program state in a single data structure like this seems a little bizzare at first, and it can sometimes be cumbersome. For example, there's no obvious way to tell which data belongs to whicih "call" of a function, because the function abstraction doesn't really exist in concatenative languages like Xi. But one clear benefit of the stack-oriented programming style is that it's very easy to introspect and debug programs, because all state is always visible. You can, at any point, print the entire data stack and see the entire "universe" of the program.

## Composable assembly

At first, writing concatenative code felt a bit like writing assembly. I was forced to think more about where my data was in the stack, and how my words and instructions moved and transformed them directly. But over time, I started to view programs less as "words manipulating data on the stack" and more as "words that string together to form longer instructions". In concatenative languages, words can compose very naturally together into longer "phrases" that do specific things.

As an example of basic composition, we can define `rot`, which rotates the top 3 items' places in the stack, like this.

```oak
// define the word "rot"
rot : ( swap ) dip swap

1 2 3 rot
// stack: < 2 3 1 >
```

Now that we have this word `rot`, rather than typing `( swap ) dip swap` everywhere and trying to imagine what's happening on the stack, I can just type `rot` and think one level higher, knowing that the top three elements on the stack are just being "rotated". Even higher level constructs like looping and iteration words can be composed together in exactly the same way.

What makes composition unique in concatenative languages is that functions (words) are composed together without naming or mentioning their parameters explicitly, but only by naming which words come after which other words. This programming style, called ["point-free" or "tacit" programming](https://en.wikipedia.org/wiki/Tacit_programming), is possible in other languages, but pervasive and natural in concatenative languages. (Hence the name "concatenative" -- programs are just "concatenations" of words.)

Even with higher level abstractions, though, reading concatenative code takes some getting used to. Here's a more complex Xi program, the [FizzBuzz](https://en.wikipedia.org/wiki/Fizz_buzz) program. Though each statement must be in a single line in Xi, I've broken them up here into multiple lines for readability.

```oak
// FizzBuzz in Xi

fizzbuzz : dup 15 divisible? (
    'FizzBuzz' print drop
) (
    dup 3 divisible?
    (
        'Fizz' print drop
    ) (
        dup 5 divisible?
        ( 'Buzz' print drop ) ( print ) if
    )
    if
) if

// main
100 ( ++ fizzbuzz ) each-integer
```

Here, the word `fizzbuzz` consumes a number at the top of the data stack and prints either 'Fizz', 'Buzz', 'FizzBuzz', or the number to output. The main program `100 ( ++ fizzbuzz ) each-integer` performs the quotation (`++ fizzbuzz`) for each integer counting up from 0 to 100, exclusive.

As a point of comparison, here's a solution to the same problem in Factor, from [Rosetta Code](https://rosettacode.org/wiki/FizzBuzz#Factor):

```
USING: math kernel io math.functions math.parser math.ranges ;
IN: fizzbuzz
: fizz ( n -- str ) 3 divisor? "Fizz" "" ? ;
: buzz ( n -- str ) 5 divisor? "Buzz" "" ? ;
: fizzbuzz ( n -- str ) dup [ fizz ] [ buzz ] bi append [ number>string ] [ nip ] if-empty ;
: main ( -- ) 100 [1,b] [ fizzbuzz print ] each ;
MAIN: main
```

You can see that the stack-manipulating words like `dup` and `nip` still appear, which makes it more difficult to get away completely from having to think about the low-level data stack.

## Pure abstraction

// Programming is mostly writing something many times, realizing it's a pattern, and abstracting it out. This is made very concrete in Xi. the concatenative/point-free abstraction lets me abstract fearlessly, which is a kind of a nice freedom.

The sample program computes factorials of every number from 1 to 10, inclusive, and prints it. This program is a great demonstration of how elegant and concise well-designed concatenative programs can be, if the right primitives are composed well. This program is just two short lines:

```oak
factorial : nat prod
10 ( ++ factorial print ) each-integer
```

First, we define the word `factorial` that takes a number, generates a list of numbers counting up from 1 to that number (`nat`), and takes their total product (`prod`). Then we loop through every number from 1 to 10, and compute the factorial and print it.

---

// Constraint as a learning aid, problem-solving tool and creative guide. After coding in Xi, writing Oak code is ... much easier! But constraints about stack still affect how I view things.

## Further reading

Learning about this completely new (to me) and esoteric topic, I found these resources to be particular helpful.

[Factor's website](https://factorcode.org/) is a good reference for broad information about Factor, which was the primary inspiration for Xi.

[A panoramic tour of Factor](https://andreaferretti.github.io/factor-tutorial/) is the most beginner-friendly treatment of Factor and concatenative programming I could find.

[A survey of stack shufflers](http://useless-factor.blogspot.com/2007/09/survey-of-stack-shufflers.html) helped me get a better sense of how to use stack shuffling words, and how to "think in Factor", i.e. think about programming by composing words together.

[Google TechTalk on Factor by its creator Slava Pestov](https://www.youtube.com/watch?v=f_0QlhYlS8g) gives a great high-level overview of what makes concatenative programming and Factor attractive.

[Bare metal x86 Forth](https://ph1lter.bitbucket.io/blog/2021-01-15-baremetal-x86-forth.html) is an advanced and insightful deep dive into bootstrapping a concatenative programming language from assembly.

