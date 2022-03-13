---
title: Oak performance I: Fibonacci faster
date: 2022-03-11T19:37:28-05:00
---

I'm in a benchmarking mood today. So I implemented a (naive, not-memoized) Fibonacci function in both Oak and Python (the language it's probably most comparable to in performance) and ran them through some measurements. This post is a pretty casual look at how they both did.

Here's the Oak version of the program. I call `print()` directly here to avoid any overhead of importing the standard library on every run of the program (though it's probably within the margin of error). I picked the number `34` sort of by trial-and-error, to make sure it was small enough for me to run the benchmarks many times but big enough that the program ran long enough to allow for consistent measurements.

```oak
fn fib(n) if n {
    0, 1 -> 1
    _ -> fib(n - 1) + fib(n - 2)
}

print(string(fib(34)) + '\n')
```

Here's the same program in Python, which I ran with Python 3.9.10 on my MacBook Pro. It's almost a direct port of the Oak program.

```py
def fib(n):
    if n == 0 or n == 1:
        return 1
    else:
        return fib(n - 1) + fib(n - 2)

print(fib(34))
```

I also made a Node.js version compiled from the Oak implementation, with `oak build --entry fib.oak -o out.js --web`.

I measured 10 runs of each program with the magic of [Hyperfine](https://github.com/sharkdp/hyperfine). Here's the (abbreviated) output:

```
Benchmark 1: node out.js
  Time (mean ± σ):     259.4 ms ±   1.9 ms    [User: 252.6 ms, System: 11.7 ms]

Benchmark 2: python3 fib.py
  Time (mean ± σ):      2.441 s ±  0.042 s    [User: 2.421 s, System: 0.011 s]

Benchmark 3: oak fib.oak
  Time (mean ± σ):     13.536 s ±  0.047 s    [User: 14.767 s, System: 0.729 s]

Summary
  'node out.js' ran
    9.41 ± 0.18 times faster than 'python3 fib.py'
   52.19 ± 0.43 times faster than 'oak fib.oak'
```

The gap between Python and native Oak isn't too surprising -- Oak is generally 4-5 times slower than Python on numerical code, so this looks right. But I was quite surprised to see just how much faster the Node.js version ran. V8 is _very, very good_ at optimizing simple number-crunching code, and it shows clearly here.

