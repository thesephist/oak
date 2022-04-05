---
title: The shape of code: plotting line frequencies in large codebases
date: 2022-04-04T15:50:30-05:00
---

Last week, [@rsms](https://twitter.com/rsms) shared a [little script to draw a histogram of the distribution of line lengths in a codebase](https://gist.github.com/rsms/36bda3b5c8ab83d951e45ed788a184f4).

!html <blockquote class="twitter-tweet"><p lang="en" dir="ltr">Was curious about source code-line length so wrote a horribly hacky bash script that draws a histogram.<a href="https://t.co/GwQeluW5gj">https://t.co/GwQeluW5gj</a> <a href="https://t.co/lacGxMMrzw">pic.twitter.com/lacGxMMrzw</a></p>&mdash; Rasmus Andersson (@rsms) <a href="https://twitter.com/rsms/status/1508900257324666882?ref_src=twsrc%5Etfw">March 29, 2022</a></blockquote> <script async src="https://platform.twitter.com/widgets.js" charset="utf-8"></script>

I've [played around with ways to visualize shapes of code before](https://github.com/thesephist/codeliner), because I think we read code so differently than we do prose — we pay less attention to individual tokens and glean more information from the silhouette and visual rhythm of lines and spaces.

There were a few things I liked about this particular take on visualizing the shape of code:

1. It proved to me that [Unicode box drawing characters](https://en.wikipedia.org/wiki/Box-drawing_character) actually let you draw some pretty nice looking graphics in the terminal.
2. A fairly simple program can generate a chart like this, if it doesn't have to run at light speed. It's probably a fun toy problem for learning a new programming language (or polishing an existing one, like Oak).
3. This got me curious about what line length distributions of different languages and codebases would look like, including those belonging to some of my own projects.

I really enjoy building small toy utilities with Oak to distract myself from real work, so I wrote a little CLI called [`codecols`](/highlight/https://oaklang.org/oak/codecols.oak), which does almost exactly what Rasmus's script does, with a few extra tricks:

- An option to render the plot up to an arbitrarily large number of columns, instead of maxing out at 100 columns of text
- Command-line flags `--max-cols` and `--histo-width` for adjusting plot scale and size
- Better text alignment and formatting for large numbers

Armed with this new toy, the first codebase I tried to measure was, of course, Oak's very own. I fed all `*.oak` files in the [Oak codebase as of today](https://github.com/thesephist/oak/tree/8bf3f657f229bb6f7f748da62ccbfacd8c4ffa98) to codecols.

```
$ cat oak/**/*.oak | codecols --max-cols 100
cols count
2    835   ███████████████████████████████████████▍
4    1272  ████████████████████████████████████████████████████████████
6    850   ████████████████████████████████████████
8    432   ████████████████████▍
10   390   ██████████████████▍
12   643   ██████████████████████████████▎
14   672   ███████████████████████████████▋
16   661   ███████████████████████████████▏
18   631   █████████████████████████████▊
20   581   ███████████████████████████▍
22   413   ███████████████████▍
24   417   ███████████████████▋
26   371   █████████████████▌
28   358   ████████████████▉
30   323   ███████████████▏
32   250   ███████████▊
34   286   █████████████▍
36   227   ██████████▋
38   231   ██████████▉
40   176   ████████▎
42   221   ██████████▍
44   213   ██████████
46   190   ████████▉
48   134   ██████▎
50   153   ███████▏
52   152   ███████▏
54   145   ██████▊
56   126   █████▉
58   100   ████▋
60   113   █████▎
62   96    ████▌
64   98    ████▌
66   95    ████▍
68   107   █████
70   101   ████▊
72   95    ████▍
74   94    ████▍
76   124   █████▊
78   126   █████▉
80   71    ███▎
82   19    ▉
84   23    █
86   15    ▋
88   19    ▉
90   17    ▊
92   9     ▍
94   7     ▎
96   11    ▌
98   9     ▍
100  10    ▍
average columns per line: 25.56
```

Despite representing a codebase written a completely different language, this plot looks a lot like Rasmus's original — it's a [bimodal distribution](https://en.wikipedia.org/wiki/Multimodal_distribution) with a long tail. There's one big peak at around 2-4 columns, a gap at around 6-12 columns, and a wider, slightly smaller peak thereafter. In this plot of the Oak codebase, the count drops off quickly after 80-column-wide lines, because I use a text editor that likes to wrap lines at 80 characters.

This bi-modality is interesting, because it doesn't occur in natural language prose. In English prose, sentence length and paragraph length have unimodal distributions, [at least in my personal writing](https://thesephist.com/posts/blog-analysis/), more closely resembling the normal bell-curve shape.

After noticing this bimodal trend, I got curious about just how universal this pattern is among other programming languages, especially languages that don't have the traditional C-like syntax. I spent the next hour excitedly running `codecols` on a bunch of different large codebases with idiomatic code style representing different languages, and I thought I'd show them to you here, starting with more classic, bimodally-distributed languages. I also speculate towards the end about why some languages might display this behavior, while some do not.

Here's Go's [`net/http` package](https://github.com/golang/go/tree/master/src/net/http) from its standard library. Go is a very C-like language in many ways, syntax being one of them. Predictably, the line length distribution looks very much like C's.

```
cols count
2    6998  ████████████████████████████████████████████████████████████
4    4035  ██████████████████████████████████▌
6    894   ███████▋
8    707   ██████
10   1526  █████████████
12   1703  ██████████████▌
14   1880  ████████████████
16   2404  ████████████████████▌
18   2823  ████████████████████████▏
20   2222  ███████████████████
22   1815  ███████████████▌
24   1764  ███████████████
26   1608  █████████████▊
28   1603  █████████████▋
30   1415  ████████████▏
32   1429  ████████████▎
34   1231  ██████████▌
36   1241  ██████████▋
38   1129  █████████▋
40   1112  █████████▌
42   1115  █████████▌
44   1001  ████████▌
46   959   ████████▏
48   986   ████████▍
50   1015  ████████▋
52   910   ███████▊
54   1038  ████████▉
56   871   ███████▍
58   895   ███████▋
60   974   ████████▎
62   1013  ████████▋
64   860   ███████▎
66   781   ██████▋
68   707   ██████
70   880   ███████▌
72   553   ████▋
74   688   █████▉
76   395   ███▍
78   321   ██▊
80   289   ██▍
82   228   █▉
84   209   █▊
86   154   █▎
88   157   █▎
90   118   █
92   93    ▊
94   95    ▊
96   68    ▌
98   55    ▍
100  39    ▎
average columns per line: 30.74
```

Another C-like language is JavaScript. Here's the distribution for [the `react` package, which is the core of the React library](https://github.com/facebook/react). We continue to see the general pattern of a peak near zero, a trough around 6-12 columns, and then a second peak. I also think it's quite cool to see the drop-off after 80 characters that signals code style with a maximum line length.

```
cols count
2    269   ██████████████▍
4    386   ████████████████████▊
6    1114  ████████████████████████████████████████████████████████████
8    884   ███████████████████████████████████████████████▌
10   373   ████████████████████
12   325   █████████████████▌
14   326   █████████████████▌
16   658   ███████████████████████████████████▍
18   499   ██████████████████████████▉
20   417   ██████████████████████▍
22   462   ████████████████████████▉
24   428   ███████████████████████
26   343   ██████████████████▍
28   402   █████████████████████▋
30   322   █████████████████▎
32   299   ████████████████
34   379   ████████████████████▍
36   397   █████████████████████▍
38   311   ████████████████▊
40   352   ██████████████████▉
42   295   ███████████████▉
44   275   ██████████████▊
46   404   █████████████████████▊
48   320   █████████████████▏
50   263   ██████████████▏
52   373   ████████████████████
54   193   ██████████▍
56   225   ████████████
58   249   █████████████▍
60   197   ██████████▌
62   185   █████████▉
64   163   ████████▊
66   205   ███████████
68   126   ██████▊
70   152   ████████▏
72   172   █████████▎
74   164   ████████▊
76   170   █████████▏
78   160   ████████▌
80   145   ███████▊
82   50    ██▋
84   38    ██
86   21    █▏
88   17    ▉
90   19    █
92   9     ▍
94   22    █▏
96   10    ▌
98   7     ▍
100  14    ▊
average columns per line: 33.02
```

Here's the plot for `*.rb` files (excluding tests) in the [Rails codebase](https://github.com/rails/rails). Ruby's syntax is similar to C-style languages in a few ways. It's an object-oriented imperative language with pretty conventional syntax for that style. But it uses `do ... end` and `def ... end` rather than `{ ... }` for block delimiters, and in general prefers keywords over characters. I suspect that's why, even though we still see a C-like distribution here, the distribution is shifted up a few columns.

```
cols count
2    110   ▎
4    6726  ████████████████▋
6    24119 ████████████████████████████████████████████████████████████
8    19542 ████████████████████████████████████████████████▌
10   14453 ███████████████████████████████████▉
12   11422 ████████████████████████████▍
14   8542  █████████████████████▏
16   6409  ███████████████▉
18   6007  ██████████████▉
20   6641  ████████████████▌
22   7448  ██████████████████▌
24   7832  ███████████████████▍
26   7557  ██████████████████▊
28   7994  ███████████████████▉
30   11033 ███████████████████████████▍
32   8261  ████████████████████▌
34   8634  █████████████████████▍
36   8102  ████████████████████▏
38   8153  ████████████████████▎
40   7941  ███████████████████▊
42   8128  ████████████████████▏
44   8282  ████████████████████▌
46   7549  ██████████████████▊
48   7571  ██████████████████▊
50   6974  █████████████████▎
52   6540  ████████████████▎
54   6230  ███████████████▍
56   5836  ██████████████▌
58   5434  █████████████▌
60   5122  ████████████▋
62   4913  ████████████▏
64   4668  ███████████▌
66   4320  ██████████▋
68   4123  ██████████▎
70   4082  ██████████▏
72   3876  █████████▋
74   4018  █████████▉
76   3849  █████████▌
78   4032  ██████████
80   3472  ████████▋
82   2756  ██████▊
84   2297  █████▋
86   2046  █████
88   1778  ████▍
90   1800  ████▍
92   1535  ███▊
94   1340  ███▎
96   1367  ███▍
98   1244  ███
100  1114  ██▊
102  1067  ██▋
104  923   ██▎
106  861   ██▏
108  739   █▊
110  696   █▋
112  666   █▋
114  564   █▍
116  530   █▎
118  561   █▍
120  434   █
average columns per line: 40.56
```

Lastly, here's Typescript's infamous [45,000-line `checker.ts` type checker file](https://github.com/microsoft/TypeScript/blob/main/src/compiler/checker.ts). It's notable for its almost complete lack of short lines, and a long "tail" of super long lines. If you read the `checker.ts` file, you'll see why this is the case — this file is very notationally heavy with long variable names and complex function and type signatures that are compressed into single lines rather than spread out for readability.

```
cols count
2    3426  ████████████████████████████████████████████████████████████
4    0
6    31    ▌
8    0
10   1783  ███████████████████████████████▏
12   388   ██████▊
14   2265  ███████████████████████████████████████▋
16   182   ███▏
18   1859  ████████████████████████████████▌
20   190   ███▎
22   1265  ██████████████████████▏
24   329   █████▊
26   1071  ██████████████████▊
28   508   ████████▉
30   852   ██████████████▉
32   595   ██████████▍
34   775   █████████████▌
36   543   █████████▌
38   672   ███████████▊
40   612   ██████████▋
42   683   ███████████▉
44   754   █████████████▏
46   698   ████████████▏
48   854   ██████████████▉
50   957   ████████████████▊
52   840   ██████████████▋
54   785   █████████████▋
56   806   ██████████████
58   790   █████████████▊
60   730   ████████████▊
62   750   █████████████▏
64   724   ████████████▋
66   681   ███████████▉
68   647   ███████████▎
70   645   ███████████▎
72   614   ██████████▊
74   563   █████████▊
76   610   ██████████▋
78   558   █████████▊
80   563   █████████▊
82   491   ████████▌
84   533   █████████▎
86   475   ████████▎
88   455   ███████▉
90   444   ███████▊
92   471   ████████▏
94   447   ███████▊
96   466   ████████▏
98   424   ███████▍
100  428   ███████▍
102  405   ███████
104  403   ███████
106  418   ███████▎
108  373   ██████▌
110  407   ███████▏
112  356   ██████▏
114  369   ██████▍
116  352   ██████▏
118  362   ██████▎
120  306   █████▎
122  303   █████▎
124  308   █████▍
126  249   ████▎
128  259   ████▌
130  248   ████▎
132  209   ███▋
134  192   ███▎
136  188   ███▎
138  156   ██▋
140  175   ███
142  188   ███▎
144  177   ███
146  150   ██▋
148  137   ██▍
150  130   ██▎
152  115   ██
154  124   ██▏
156  88    █▌
158  90    █▌
160  97    █▋
162  121   ██
164  87    █▌
166  86    █▌
168  74    █▎
170  56    ▉
172  57    ▉
174  66    █▏
176  47    ▊
178  54    ▉
180  44    ▊
182  35    ▌
184  44    ▊
186  30    ▌
188  39    ▋
190  40    ▋
192  26    ▍
194  30    ▌
196  26    ▍
198  24    ▍
200  19    ▎
average columns per line: 61.35
```

I also decided to look at a few languages that don't have some of the characteristic C-style syntax that give JS, Go, and C their shape. For example, here's the Python code behind the [Black code formatter for Python](https://github.com/psf/black/tree/main/src/black), formatted with Black itself.

We no longer see the double-peak shape here, and instead see something much closer to a unimodal distribution with a peak around 22-24 columns. We also see the distribution max out at exactly 88 characters, which is the strict line-length limit set by the Black formatter.

This plot began to convince me that the bimodal distribution was an artifact of C-like syntax, specifically the pattern of ending parentheses or double braces that end up in their own line, as `}` or `});`. Python has fewer such constructions (for example, in argument lists), and uses indentation instead for block nesting, so we don't see such short lines emphasized in the plot.

```
cols count
2    49    █████████▊
4    18    ███▌
6    96    ███████████████████▏
8    178   ███████████████████████████████████▍
10   234   ██████████████████████████████████████████████▋
12   194   ██████████████████████████████████████▋
14   281   ████████████████████████████████████████████████████████
16   217   ███████████████████████████████████████████▎
18   274   ██████████████████████████████████████████████████████▌
20   249   █████████████████████████████████████████████████▋
22   212   ██████████████████████████████████████████▎
24   262   ████████████████████████████████████████████████████▏
26   255   ██████████████████████████████████████████████████▊
28   301   ████████████████████████████████████████████████████████████
30   232   ██████████████████████████████████████████████▏
32   291   ██████████████████████████████████████████████████████████
34   208   █████████████████████████████████████████▍
36   255   ██████████████████████████████████████████████████▊
38   223   ████████████████████████████████████████████▍
40   227   █████████████████████████████████████████████▏
42   184   ████████████████████████████████████▋
44   199   ███████████████████████████████████████▋
46   166   █████████████████████████████████
48   129   █████████████████████████▋
50   152   ██████████████████████████████▎
52   156   ███████████████████████████████
54   144   ████████████████████████████▋
56   131   ██████████████████████████
58   126   █████████████████████████
60   129   █████████████████████████▋
62   101   ████████████████████▏
64   115   ██████████████████████▉
66   113   ██████████████████████▌
68   123   ████████████████████████▌
70   119   ███████████████████████▋
72   104   ████████████████████▋
74   130   █████████████████████████▉
76   162   ████████████████████████████████▎
78   186   █████████████████████████████████████
80   144   ████████████████████████████▋
82   107   █████████████████████▎
84   107   █████████████████████▎
86   102   ████████████████████▎
88   64    ████████████▊
90   0
92   0
94   0
96   0
98   0
100  0
average columns per line: 40.53
```

Another more exotic language is assembly. This plot shows assembly files (`*.s`) from Go's [`crypto`](https://github.com/golang/go/tree/master/src/crypto) standard library packages. Once again, we don't see any bimodal peaks here. In fact, there's a very strong preference in these files for lines of around 12-22 columns long, which makes sense considering how "rectangular" most assembly code looks. I have a feeling that lines longer than 30 columns are comments, rather than code.

```
cols count
2    111   ███▏
4    170   ████▉
6    83    ██▍
8    232   ██████▊
10   690   ████████████████████▏
12   1111  ████████████████████████████████▍
14   2057  ████████████████████████████████████████████████████████████
16   1313  ██████████████████████████████████████▎
18   1624  ███████████████████████████████████████████████▎
20   1288  █████████████████████████████████████▌
22   1291  █████████████████████████████████████▋
24   934   ███████████████████████████▏
26   829   ████████████████████████▏
28   764   ██████████████████████▎
30   593   █████████████████▎
32   497   ██████████████▍
34   166   ████▊
36   388   ███████████▎
38   638   ██████████████████▌
40   418   ████████████▏
42   176   █████▏
44   245   ███████▏
46   558   ████████████████▎
48   153   ████▍
50   249   ███████▎
52   122   ███▌
54   241   ███████
56   153   ████▍
58   92    ██▋
60   140   ████
62   106   ███
64   96    ██▊
66   122   ███▌
68   136   ███▉
70   101   ██▉
72   92    ██▋
74   43    █▎
76   186   █████▍
78   23    ▋
80   27    ▊
82   29    ▊
84   27    ▊
86   33    ▉
88   14    ▍
90   9     ▎
92   11    ▎
94   22    ▋
96   10    ▎
98   0
100  0
average columns per line: 27.35
```

I'll conclude this exploration by lingering a bit on Lisp programs and their shape. This plot is for all Clojure source files (`*.clj`) in the [Clojure source tree](https://github.com/clojure/clojure). Once again, because Lisps don't tend to leave parentheses or braces hanging on their own lines at the end of blocks, we see a more unimodal distribution.

```
cols count
2    68    ██▎
4    109   ███▊
6    198   ██████▊
8    286   █████████▊
10   447   ███████████████▍
12   542   ██████████████████▋
14   591   ████████████████████▎
16   1741  ████████████████████████████████████████████████████████████
18   736   █████████████████████████▎
20   644   ██████████████████████▏
22   646   ██████████████████████▎
24   728   █████████████████████████
26   585   ████████████████████▏
28   579   ███████████████████▉
30   616   █████████████████████▏
32   571   ███████████████████▋
34   539   ██████████████████▌
36   474   ████████████████▎
38   457   ███████████████▋
40   489   ████████████████▊
42   430   ██████████████▊
44   456   ███████████████▋
46   432   ██████████████▉
48   411   ██████████████▏
50   351   ████████████
52   399   █████████████▊
54   357   ████████████▎
56   337   ███████████▌
58   382   █████████████▏
60   303   ██████████▍
62   361   ████████████▍
64   382   █████████████▏
66   584   ████████████████████▏
68   568   ███████████████████▌
70   614   █████████████████████▏
72   302   ██████████▍
74   286   █████████▊
76   206   ███████
78   147   █████
80   175   ██████
82   180   ██████▏
84   111   ███▊
86   93    ███▏
88   68    ██▎
90   79    ██▋
92   56    █▉
94   66    ██▎
96   43    █▍
98   37    █▎
100  26    ▉
average columns per line: 40
```

I did something similar with Lisp code from [Klisp, my own flavor of Lisp](https://github.com/thesephist/x-oak-klisp) I wrote for hobby interpreter projects. Unsurprisingly, I saw a very similar pattern of a single large peak.

```
cols count
2    0
4    0
6    50    █████████████████████████
8    72    ████████████████████████████████████
10   81    ████████████████████████████████████████▌
12   69    ██████████████████████████████████▌
14   89    ████████████████████████████████████████████▌
16   94    ███████████████████████████████████████████████
18   120   ████████████████████████████████████████████████████████████
20   108   ██████████████████████████████████████████████████████
22   82    █████████████████████████████████████████
24   82    █████████████████████████████████████████
26   76    ██████████████████████████████████████
28   93    ██████████████████████████████████████████████▌
30   89    ████████████████████████████████████████████▌
32   77    ██████████████████████████████████████▌
34   52    ██████████████████████████
36   34    █████████████████
38   53    ██████████████████████████▌
40   32    ████████████████
42   25    ████████████▌
44   22    ███████████
46   31    ███████████████▌
48   29    ██████████████▌
50   9     ████▌
52   11    █████▌
54   23    ███████████▌
56   8     ████
58   7     ███▌
60   12    ██████
62   9     ████▌
64   1     ▌
66   8     ████
68   4     ██
70   1     ▌
72   7     ███▌
74   6     ███
76   9     ████▌
78   7     ███▌
80   8     ████
82   0
84   0
86   0
88   1     ▌
90   0
92   1     ▌
average columns per line: 26.71
```

I'm not really sure if there's a lesson in any of this, except that (1) charts are cool! and (2) some syntactic properties of programming languages show up in meaningful and interesting ways in these visualizations. If you want to try running this tool on some of your code, you can find my Oak script on [Github Gist](https://gist.github.com/thesephist/147ed8f22eef181b33333f6c4b742e75) or try a [Rust implementation](https://twitter.com/svoisen/status/1511733720188809216) if that's more your thing.

I'll leave you with one last set of charts I found on Twitter, comparing the impact of different code formatters on the distribution of lines of Python code:

!html <blockquote class="twitter-tweet"><p lang="en" dir="ltr">V nice. Made me curious how various Python formatters affect linecol distribution in a fairly large Python project. Quick n dirty script using termplotlib <a href="https://t.co/2gC0QY0X8v">https://t.co/2gC0QY0X8v</a><br>gist: <a href="https://t.co/SnaMTgIymo">https://t.co/SnaMTgIymo</a> <a href="https://t.co/GcQ6m6oVzf">pic.twitter.com/GcQ6m6oVzf</a></p>&mdash; nyc.russell.mtb (@RussellRomney) <a href="https://twitter.com/RussellRomney/status/1509038933425930245?ref_src=twsrc%5Etfw">March 30, 2022</a></blockquote> <script async src="https://platform.twitter.com/widgets.js" charset="utf-8"></script>

