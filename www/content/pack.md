---
title: Oak pack: statically linked, executable Oak binaries
date: 2022-03-18T09:00:30-05:00
---

This week, I added **oak pack**, the newest in Oak's collection of self-hosted (written in Oak itself) language tools. `oak pack` is a tool to package or "pack" Oak programs into stand-alone, self-contained binary executables that can be distributed independently without the Oak interpreter. Before diving into how it works, here's a quick look at what `oak pack` can do.

Let's say we have this Oak program. It says `Hello from \`oak pack\`!` and tells us the name of the running executable itself.

```oak
std := import('std')
fmt := import('fmt')

std.println('Hello from `oak pack`!')
fmt.printf('Current executable: "{{0}}"', args().0)
```

When we run it with the `oak` executable, we get

```
$ oak hello-oak-pack.oak
Hello from `oak pack`!
Current executable: "oak"
```

We can now "pack" this Oak program into an executable called "hello".

```
oak pack --entry hello-oak-pack.oak --output hello
```

We can now run the same program by simply running the `./hello` binary executable, instead of using the Oak CLI. If we do so, we see a nearly identical output. But this time, the running executable is called `./hello`!

```
$ ./hello
Hello from `oak pack`!
Current executable: "./hello"
```

`oak pack` can also pack multi-file Oak programs with imports between them, pre-include other Oak modules, and cross-pack binaries between macOS, Linux, and BSDs if given Oak interpreters for those other platforms. This post is about why this may be useful, how other languages' tools do it, and how it works within Oak.

## Why static binaries?

Static binaries are _simple_, and most of the benefits of static binaries come from that simplicity.

They're simple to **distribute**. If I write an Oak program and try to share it or back it up as a collection of files, I have to somehow move many files in a specific hierarchy of folders all together, so that the overall structure is preserved. I also need to ensure that the right version of the Oak interpreter itself is installed on the system on which I want to run the program. If instead I can "package" an Oak program somehow into a single executable binary, all I need to do to share or deploy that program elsewhere is to transfer that single file securely.

Because static binaries are self-contained, they also have **longer lifetimes**. With a typical web app or CLI written in Oak, for that program to keep working, it needs to be maintained alongside its dependencies and its supported version of the Oak interpreter. If I can bundle all of that into a single-file executable, its dependencies or interpreter won't go out of date or out of sync with the program itself, and it has a much better chance of surviving for years, maybe even decades.

Because they can be invoked simply, they **fit better into workflows**, especially UNIX-style workflows in the command line. Instead of typing out `oak my-cli.oak` every time, I can simply type `my-cli` to run the program. A [shebang](https://en.wikipedia.org/wiki/Shebang_(Unix)) line at the top of an executable script can get some of the same benefits, but a self-contained executable file is also easier to invoke as a subprocess from other programs, and avoids [portability issues with shebang support between different platforms](https://en.wikipedia.org/wiki/Shebang_(Unix)#Portability).

Because of these benefits, especially easy distribution and longevity, more programming language communities are prioritizing a language's ability to be packaged into static binaries. If my memory serves, Go really led this culture change, but other languages like Rust and Deno have embraced this idea as well. Now, it's time for Oak to do the same.

## Prior art

Historically, static linking was not the norm. Programmers avoided it in favor of _dynamic linking_, which let many different programs share components to save on storage and memory. These shared components are called [dynamically linkable libraries or "shared object" files (.dll or .so)](https://en.wikipedia.org/wiki/Library_(computing)#Shared_libraries), if you've come across them in your operating system. Statically linked binaries package up all those shared components into self-contained units that each contain their own versions of the shared files, and lose those efficiency benefits.

It seems static linking became more fashionable as storage and memory became less constraining, and those were trumped in priority by how easy it was to distribute and deploy programs reliably. Ahead-of-time compiled languages achieve this self-contained-ness quite differently than interpreted languages with large runtimes. So while Go was the first major language (if I'm mistaken, please do correct me) in this recent drive towards static linking, it's not very helpful to study how Go builds its binaries. Instead, I took inspiration for `oak pack` from similar initiatives in three other language ecosystems: Node.js, Deno, and Python.

### Vercel `pkg`

[Vercel's `pkg` tool](https://github.com/vercel/pkg) is the oldest of the three here. I remember being quite surprised that this kind of "compile an interpreted language into an executable" was even possible when I first came across it. With it, a Node.js programmer can just run `pkg my-program.js` to get a self-contained executable called `./my-program`. `pkg` achieves this by first pre-compiling parts of the JavaScript program into bytecode using the V8 JavaScript engine, and combining that bytecode program with a snapshot of the project's filesystem with other Node.js modules and any other included assets. All of these ingredients are packaged into a single binary built on top of a patched version of the Node.js interpreter, so that when the binary is run, the included Node.js interpreter will spin up and begin running the included program embedded elsewhere in the same executable file.

`pkg` is very versatile, because in addition to the program being compiled, it can also include a [snapshot of the project's filesystem layout](https://github.com/vercel/pkg#snapshot-filesystem) that can include other assets, and can be read by the program as if it were a real filesystem, despite the fact that the entire snapshot is packed into an executable in reality.

### Deno compile

Deno, a TypeScript runtime, shipped a feature called `[deno compile](https://github.com/denoland/deno/issues/986)` that achieves something very similar to Vercel's `pkg` tool, but built into the Deno CLI itself.

Deno compile works in a very similar way to `pkg`. The implementation of Deno compile is written in Rust and quite readable, even for a Rust newbie like me. You can find it in [this pull request](https://github.com/denoland/deno/pull/8539/files#diff-8301906666625ca9014051b459e158d180ab3c3fbee191b0a6c715f8c8036ad7R127) on Deno's GitHub repository. Here's how it works:

1. First, Deno uses their built-in bundler, [Deno bundle](https://deno.land/manual/tools/bundler), to produce a single-file JavaScript bundle from a TypeScript program.
2. Then they produce an executable binary by making a copy of the Deno CLI and appending the JavaScript bundle to the end of it.
3. Deno then marks the end of that modified Deno executable with an 8-byte magic sequence (`d3nol4nd`, which is kind of cute) and the size of the included JavaScript bundle, so that when the Deno executable in that compiled file runs, it can check the last 8 bytes of itself and extract the bundled JavaScript program to run it.

If you think about it, Deno compile works by cloning itself, and then making that clone...pull its own program out of its butt.

Corey Butler has a [good write up about building JavaScript executables](https://medium.com/swlh/javascript-executables-5644ead7016d) that discusses Deno's "compile" feature in more depth, if you're interested in their rationale. Because Deno's implementation was the most readable implementation of this general method that I could find, Oak's implementation of `oak pack` is modeled very closely after Deno compile.

### PyOxidizer

Python is [rather well known for being difficult to package and deploy](https://xkcd.com/1987/), and [PyOxidizer](https://github.com/indygreg/PyOxidizer) is a project to put an end to those woes with a tool that can package a Python project along with all of its dependencies into a static executable that "just works".

PyOxidizer makes use of two tricks:

1. It "bundles" a Python program at compile-time by storing bytecode representations of all Python modules from a given program in a hashmap stored in the final executable.
2. It uses a modified version of the Python interpreter that imports modules from this hashmap, rather than the Python distribution on the filesystem.

This method follows the pattern we've seen across all three implementations of compiling interpreted languages so far: we need some way to bundle all dependencies of a program into a single "thing" of some kind, and we need to include the language's interpreter into the final executable binary, such that it will execute the bundled program on startup.

Gregory Szorc, the creator of PyOxidizer, gave a great talk about how it works at [this Facebook Rust meetup](https://youtu.be/uHm939mXefs) (his talk is the first half of the recording). If you're interested in PyOxidizer or implementing something similar for another language, I think this talk is a good resource.

## How `oak pack` works

As I mentioned briefly above, `oak pack` is modeled closely after `deno compile`, and shares the same basic structure. You can find my changes to the Oak CLI adding this feature [in my `oak pack` commit to the Oak CLI](https://github.com/thesephist/oak/commit/9b917f4de2c8269f58af81fea10da262912f93f0). There are two components to this change: what happens when I run `oak pack`, and what happens when Oak starts up.

The new [`oak pack` command](https://github.com/thesephist/oak/blob/main/cmd/pack.oak) builds a static executable binary by the following steps:

1. It makes a copy of itself (the currently running Oak CLI) in memory.
2. It makes a single-file Oak bundle of the Oak program being compiled using `oak build`, and appends that to the file in memory.
3. It encodes the length of that Oak bundle from step 2 into a 24-byte buffer, and appends that to the file.
4. Finally, it marks the file as an `oak pack` file using 8 magic bytes added to the end of the file, `oak \\x19\\x98\\x10\\x15`.

When I run this new binary, the Oak CLI starts up normally, but checks the last 8 bytes of its own file before doing anything else. If it finds those 8 magic bytes, it extracts the bundled Oak program using the length of the program embedded into the binary, and executes it immediately.

We can see the result of this process in a binary packed using `oak pack`. For example, here's a byte-level read-out of the last 32 bytes of the `./hello` binary I used to open this blog post:

```
$ oak eval "fs.readFile('./hello') |> std.takeLast(32)" | xxd -
00000000: 2020 2020 2020 2020 2020 2020 2020 2020
00000010: 2020 2020 2037 3539 6f61 6b20 1998 1015       759oak ....
```

The first 24 bytes here encode the number "759" padded with a bunch of spaces, which means that the bundled program is 759 bytes long. The last 8 bytes are the 8 magic bytes marking an `oak pack`-produced binary.

Although it seems like adding arbitrary data to the end of executables would somehow corrupt the executable, at least on the common operating systems (Linux, macOS, Windows), executable files seem to be structured in a way that makes them invariant to random bytes being added to the end. Linux's ELF executable format, for example, indexes all sections in the file from the beginning of the file, [as I wrote about before](https://dotink.co/posts/elf/). It doesn't seem like this is a universal property of executable file formats we can always assume, since other file formats (like ZIP archives) index into the file from its end, but the executable formats used by common operating systems are stable because they are a part of the stable interfaces to these operating systems. I don't expect them to change dramatically, and my confidence in that is bolstered by other tools like Deno using this same strategy to pack up their programs into static binaries.

## Unresolved threads

Now that `oak pack` itself works, some new questions have emerged. The most obvious question to ask after this change is whether "packing an Oak program" can break an Oak program, and the answer is currently _yes_...sometimes, if you don't prepare for it.

Most things that can be observed from within an Oak program don't change in the `oak pack` process, except _what exactly is running_. Oak programs can answer that question using the `args()` built-in function, which returns a list of command-line arguments, starting with the currently running executable. When a packed Oak program runs, the command-line arguments will look different, because there is no longer a CLI called `oak` running a program file ending in `.oak` — there is only whatever static binary executable is running. Oak programs that depend on `args()` can sometimes break as a result of this, depending on how exactly that list is parsed.

Different runtimes deal with this differently. Some allow programs to detect whether it's running in a compiled mode; others transparently modify the result of `args()` so that it still looks from inside the program as if an Oak CLI is running the program normally. Neither seem very elegant, and I'm waiting to see what I need and prefer in usage before committing to an API change.

It would also be useful to be able to include assets other than `.oak` source files in the packed binary, so that I can include assets when building and distributing a web server written in Oak, for example. Go does this using compiler pragmas (`//go:embed`), which Oak itself uses to include the Oak standard library in the Oak CLI. Rust uses a macro called `include_bytes!` to do something similar. Once again, I'm not sure if there's a good API to do this kind of thing, especially in an interpreted language like Oak. But I'll continue exploring this space.

Since shipping `oak pack`, I've really enjoyed the ability to build Oak programs that are stand-alone executables. It feels really _right_ to write personal CLI tools like [`times`](/posts/times/), and install it on my system as an executable just like any other UNIX utility. All this, mind you, builds on a toolchain that's entirely written in Oak — from parsing and bundling Oak code, to actually constructing the final binary. I think that's pretty great.

