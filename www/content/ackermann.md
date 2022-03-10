---
title: Playing with the Ackermann function: a tour of computational complexity
date: 2022-03-10T15:34:17-05:00
---

Once in a while, I like to play around with Oak to test new mathematical ideas I come across. Today, I was reminded of [Ackermann functions](https://en.wikipedia.org/wiki/Ackermann_function), and thought I'd write a small Oak program to get a feel for how the function behaved.

The Ackermann function is quite simple to write. It takes two integers and returns an integer.

```oak
fn ackermann(m, n) if {
    m = 0 -> n + 1
    n = 0 -> ackermann(m - 1, 1)
    _ -> ackermann(m - 1, ackermann(m, n - 1))
}
```
It's best known in mathematics and computer science for two properties:

1. It increases super-exponentially, even for very small input values.
2. It's not a [primitive recursive](https://en.wikipedia.org/wiki/Primitive_recursive_function) function, which means that it isn't expressible in terms of simple, finite `for` loops.

We'll take a look at both of those properties here.

## The script

To quickly play around with values of the Ackermann function, I wrote a simple script that let me type in input values into an interactive prompt and measure the run-time of the function:

```oak
Cli := cli.parse()
Time? := Cli.opts.t != ? | Cli.opts.time != ?

std.println('Ackermann function calculator.')
with std.loop() fn {
    print('.> ')
    // input is of the form "M N, M N, M N"
    // assume input never fails b/c it's just a toy calculator
    args := input().data |> str.split(',') |> std.map(fn(pair) {
        pair |>
            str.split(' ') |>
            std.map(int) |>
            std.compact() |>
            std.take(2)
    })
    args |> with std.each() fn(pair) {
        [m, n] := pair
        prefix := if len(args) {
            1 -> ''
            _ -> 'ack({{0}}, {{1}}) = ' |> fmt.format(m, n)
        }
        if m != ? & n != ? {
            true -> {
                start := nanotime()
                a := ackermann(m, n)
                elapsed := nanotime() - start
                (prefix + if {
                    Time? -> '{{0}} ({{1}}ms)'
                    _ -> '{{0}}'
                }) |> fmt.printf(a, math.round(elapsed / 1000000, 3))
            }
            _ -> std.println('Invalid input. Try again.')
        }
    }
}
```

It works like this:

```
$ oak ackermann.oak --time
Ackermann function calculator.
.> 1 1
3 (0.008ms)
.> 2 2
7 (0.068ms)
.> 3 4
125 (17.101ms)
.> 3 5, 3 6, 3 7
ack(3, 5) = 253 (54.885ms)
ack(3, 6) = 509 (220.24ms)
ack(3, 7) = 1021 (973.239ms)
.> 3 10
8189 (157760.616ms)
.>
```

## Big numbers are big

The Ackermann function stays quite tame for `m = 1` and `m = 2`.

```
.> 1 1, 1 2, 1 3, 1 4, 1 5, 1 6, 1 7, 1 8, 1 9, 1 10
ack(1, 1) = 3 (0.009ms)
ack(1, 2) = 4 (0.009ms)
ack(1, 3) = 5 (0.012ms)
ack(1, 4) = 6 (0.015ms)
ack(1, 5) = 7 (0.019ms)
ack(1, 6) = 8 (0.022ms)
ack(1, 7) = 9 (0.024ms)
ack(1, 8) = 10 (0.023ms)
ack(1, 9) = 11 (0.024ms)
ack(1, 10) = 12 (0.034ms)
.> 2 1, 2 2, 2 3, 2 4, 2 5, 2 6, 2 7, 2 8, 2 9, 2 10
ack(2, 1) = 5 (0.025ms)
ack(2, 2) = 7 (0.04ms)
ack(2, 3) = 9 (0.067ms)
ack(2, 4) = 11 (0.095ms)
ack(2, 5) = 13 (0.131ms)
ack(2, 6) = 15 (0.187ms)
ack(2, 7) = 17 (0.208ms)
ack(2, 8) = 19 (0.24ms)
ack(2, 9) = 21 (0.29ms)
ack(2, 10) = 23 (0.351ms)
```

Both sequences increase linearly by 1 and 2, and are quick to compute. For `m = 3`, though, things start looking different:

```
.> 3 1, 3 2, 3 3, 3 4, 3 5, 3 6, 3 7, 3 8, 3 9, 3 10
ack(3, 1) = 13 (0.159ms)
ack(3, 2) = 29 (0.759ms)
ack(3, 3) = 61 (3.57ms)
ack(3, 4) = 125 (12.222ms)
ack(3, 5) = 253 (50.915ms)
ack(3, 6) = 509 (212.928ms)
ack(3, 7) = 1021 (980.088ms)
ack(3, 8) = 2045 (4745.548ms)
ack(3, 9) = 4093 (26737.925ms)
ack(3, 10) = 8189 (157760.616ms)
```

Not only do the numbers increase rapidly, the runtime increases even faster -- `ack(3, 10)` took over 2 minutes to compute. The runtime of the Ackermann function increases so quickly because the function changes its return value by 1 each iteration, and only sometimes. That means every increment of 1 in the function's return value corresponds to several invocations of the Ackermann function during its computation.

Some facts about the next sequence, `ack(4, _)`

- I tried to compute a sequence for `ack(4, _)` but only got as far as `ack(4, 0) = 13 (0.574ms)`.
- `ack(4, 1)` too long to compute, and I had to stop the program, but according to Wikipedia, its value should be `65533`.
- `ack(4, 2)` is too large to even write down here, and is best described as `2^65536 − 3`.

I expected the function itself to increase quickly, from its reputation, but I was surprised the run-time increased much faster than even the function itself.

## To recurse or not to recurse

After getting these results, I tried to compile the Ackermann function to JavaScript and run the script on Node.js, where Oak often runs faster especially for numerical computations. A naive attempt, compiling the same file using `oak build --web` and running on Node.js, resulted in a stack overflow at `ack(3, 10)`.

_That's fine_, I thought. Oak's recursion limit must be higher than JavaScript's, so the stack overflowed deep in the recursive call. Usually, in this situation, I'd rewrite the function so that the function is [tail-recursive](https://en.wikipedia.org/wiki/Tail_call), so that Oak's compiler could optimize the recursion down to a loop. But for the Ackermann function, there isn't a straightforward way to unroll the recursion into a simple tail recursion, because as noted above, the Ackermann function isn't primitive recursive!

The most common way to rewrite the Ackerman function using loops (or basic tail recursion) is actually a kind of a cheat: rather than using the programming language's stack, which can overflow, we can simulate a stack manually, using a growable list of numbers. The Ackermann function then operates not on two parameters to a function, but the top two numbers of this manually-managed stack.

I wrote this new loop-based variant of the Ackermann function (which I affectionately called `stackermann`) in Oak using JavaScript's arrays:

```oak
fn stackermann(m, n) {
    stack := [m, n] // begin with the stack [m, n]
    with std.loop() fn(_, break) if len(stack) {
        // if stack has only 1 value, return it
        1 -> break(stack.0)
        _ -> {
            // Ackermann function operating on
            // the top 2 values of the stack
            n := stack.pop()
            m := stack.pop()
            if {
                m = 0 -> stack.push(n + 1)
                n = 0 -> {
                    stack.push(m - 1)
                    stack.push(1)
                }
                _ -> {
                    stack.push(m - 1)
                    stack.push(m)
                    stack.push(n - 1)
                }
            }
        }
    }
}
```

This function, based on a loop and a manually managed stack, runs slightly faster on Node.js than the original Oak version runs natively:

```
.> 3 1, 3 2, 3 3, 3 4, 3 5, 3 6, 3 7, 3 8, 3 9, 3 10
ack(3, 1) = 13 (1ms)
ack(3, 2) = 29 (1ms)
ack(3, 3) = 61 (3ms)
ack(3, 4) = 125 (6ms)
ack(3, 5) = 253 (5ms)
ack(3, 6) = 509 (19ms)
ack(3, 7) = 1021 (73ms)
ack(3, 8) = 2045 (301ms)
ack(3, 9) = 4093 (1095ms)
ack(3, 10) = 8189 (4375ms)
```

Armed with this slightly faster implementation of the algorithm, we can now attack `ackermann(4, 0)` and `ackermann(4, 1)`.

```
ack(4, 0) = 13 (0ms)
ack(4, 1) = 65533 (275487ms)
```

Even with the speed improvements, the run-time of `ackermann(4, 1)` still dwarfs that of any in the `ackermann(3, _)` sequence. `ackermann(4, 2)` and beyond are, unfortunately, still out of our reach. It would take many days of compute on my lowly laptop to find those answers, and I wasn't about to put mine through that today.

As a last step in my journey to make the function run even faster, I explored the possibility of [memoizing](https://en.wikipedia.org/wiki/Memoization) the Ackermann function. But [this Stack Overflow answer](https://stackoverflow.com/a/13088510) confirmed my suspicions that memoizing the Ackermann function doesn't yield much speed-up, because the domain of the function is two-dimensional, and memoization yields minor speed improvements at the cost of much more heap memory usage. Even computing `ackermann(4, 2)`, even with more efficient programming languages and memoization, can result in out-of-memory errors.

At this point, I decided to put my own exploration to a halt. Looking at the two different implementations of the Ackermann function -- the recursion-based and loop/stack-based -- gave me a good sense of how the function achieves its massive computational complexity, and why simplifying it isn't trivial.

If you're interested in digging deeper, both the [Wikipedia page](https://en.wikipedia.org/wiki/Ackermann_function) and [Computerphile's video on the subject](https://youtu.be/i7sm9dzFtEI) seem like great starting points.

