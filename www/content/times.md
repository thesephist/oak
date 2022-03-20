---
title: Building a personal time zone CLI with Oak
date: 2022-03-20T00:50:30-05:00
---

`times` is a little command-line utility I wrote to show me dates and times in all the time zones I care about quickly. It's the next in a series of [toy Oak programs](http://localhost:9898/posts/ackermann/) I've been writing when I get bored, but what sets it apart is that I've used Oak's new [`oak pack` feature](/posts/pack/) to install it as a stand-alone executable on my Mac, so I can call it with `times` anywhere, and share it with my friends, too.

`times` takes no input or command-line arguments, and does exactly one thing: print this list:

```
$ times
US Eastern   0:52 3/20  -4h
US Pacific  21:52 3/19  -7h
     Korea  13:52 3/20  +9h
   Ukraine   6:52 3/20  +2h
    France   5:52 3/20  +1h
   Germany   5:52 3/20  +1h
        UK   4:52 3/20  +0h
```

Though you can't see it here in the copy-pasted output, the terminal output also color-codes different parts of this table. The utility doesn't really interoperate with any other tools, and it's not particularly configurable, so it doesn't quite fit neatly into the UNIX philosophy. But hey — if I want to change the time format or add a new time zone, I can just dig into the Oak source code, re-compile, and re-install, all in under a minute.

Especially after shipping `oak pack`, I think Oak is turning out to be a really pleasant way for me to customize my desktop working environment with little utilities like this. Not only does it produce simple self-contained programs, but each program doesn't need too many dependencies, because the standard library contains most of the functionalities I routinely need in my tools, from date and time formatting to Markdown compilation. I've always maintained that even though Oak is a general purpose programming language, it isn't trying to fill every niche. It's trying to be a [tool for building personal tools and projects](/posts/why/), and I'm excited about the way Oak has been moving towards that vision steadily over time.

---

Here's the full `times.oak` program, for sake of completeness:

```oak
std := import('std')
str := import('str')
math := import('math')
fmt := import('fmt')
datetime := import('datetime')

Zones := [
	{ name: 'US Eastern', offset: -4 } // Daylight Saving Time
	{ name: 'US Pacific', offset: -7 } // Daylight Saving Time
	{ name: 'Korea', offset: 9 }
	{ name: 'Ukraine', offset: 2 }
	{ name: 'France', offset: 1 }
	{ name: 'Germany', offset: 1 }
	{ name: 'UK', offset: 0 }
]

Now := int(time())
MaxNameLen := math.max(Zones |> std.map(:name) |> std.map(len)...)

fn yellow(s) '\x1b[0;33m' << s << '\x1b[0;0m'
fn gray(s) '\x1b[0;90m' << s << '\x1b[0;0m'

Zones |> with std.each() fn(zone) {
	{
		month: month
		day: day
		hour: hour
		minute: minute
		second: second
	} := datetime.describe(Now + zone.offset * 3600)
	fmt.printf(
		'{{0}}  {{1}}  {{2}}'
		zone.name |> str.padStart(MaxNameLen, ' ')
		fmt.format(
			'{{0}}:{{1}} {{2}}/{{3}}'
			hour |> string() |> str.padStart(2, ' ')
			minute |> string() |> str.padStart(2, '0')
			month
			day
		) |> yellow()
		if zone.offset < 0 {
			true -> '-' + string(math.abs(zone.offset)) + 'h'
			_ -> '+' + string(zone.offset) + 'h'
		} |> gray()
	)
}
```

