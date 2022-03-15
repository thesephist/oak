---
title: Oak Notebook: an experiment in dynamic, interactive documents
date: 2022-03-14T07:34:17-05:00
---

On a whim this week, I decided to build a little experiment from some ideas I had simmering in my mind about a way to write documents with dynamic, interactive components interleaved into prose. I've published the experiment under the name **[Oak Notebook](https://thesephist.github.io/x-oak-notebook/)**, but it's not really a complete project or even a complete idea, so much as just a playground onto which I've thrown some things that seemed interesting.

!html <p>
    <img src="/img/oak-notebook-demo.gif" alt="A demo of me scrolling through the Oak Notebook demo website"
        style="border-radius: 1% / 1.5%;">
</p>

I'll quote at length from the demo site itself, since that's the best way I know of explaining what Oak Notebook is:

>Oak Notebook is an experimental tool for creating [dynamic documents](https://thesephist.com/posts/notation/#dynamic-notation) with Markdown and [Oak](https://oaklang.org/). It's both a way of writing documents with interactive, programmable "panels" for explaining and exploring complex ideas, and a "compiler" script that transforms such Markdown documents into HTML web pages. It's a bit like [MDX](https://mdxjs.com/), if MDX was focused specifically on input widgets and interactive exploration of information.
>
>I was inspired initially by [Streamlit](https://docs.streamlit.io/library/api-reference) and [Bret Victor's Scrubbing Calculator](http://worrydream.com/ScrubbingCalculator/) to explore ideas in this space, but what's here isn't a specific re-implementation of either of those concepts, and takes further inspirations from other products, experiments, and prior art.
>
>Oak Notebook provides a way to embed these interactive panels into documents written in Markdown without worrying about styling user interfaces or managing rich user input components.

I'm hoping that Oak Notebook will become a tool that I'll reach for to build little visualizations and interactive explorations to help me understand and think through situations; hopefully, in that process, Oak Notebook will improve slowly but surely to fit my use cases more and more. I also think this domain of interactive, dynamic documents is fascinating as a space for research, but it's always felt far too broad and deep for me to wade into comfortably. It's very difficult to even try to define an alphabet for the primitive components of dynamic data displays. Oak Notebook is a small and casual start to me cautiously exploring this domain of ideas with my favorite little toy language.

There are lots of question to answer, even for something as basic as what's on the demo site today. On the interface design side, some questions I face are:

- How do you clearly communicate what parts of the interactive embeds are inputs to be played with, and which are simply output display?
- How tightly should interactivity be woven into prose? Should authors be able to embed arbitrary inline components right into prose? Should dynamic pieces be offset from normal paragraphs in some way?
- Can we have an "alphabet" of basic input components from which more complex and situationally apt input mechanisms (rotatable knobs, color picker, 2D grid, text selection, time series input) can be built?
- How can we make ["running a program backwards" with automatic differentiation of programs](https://alpha.trycarbide.com/) easier to use for building explorable explanations?
- How can we make interactive embeds programmable without letting the complexity of these mini-programs get out of hand and turn into fully-fledged pieces of software to maintain?
- If there are errors in the program that runs each embed, how should they be reported to the user?

On the technical implementation side, some questions I face are:

- How can we make these dynamic documents easier to build and share? Can we minimize the toolchain requirements? Simplify the APIs?
- How might we take inherently computationally expensive simulations, like DL inference or long-running algorithms, and make them smooth to interact with? Can we use dynamic programming or caching to have the interface move faster than the underlying computation?

At a higher level, here are some guiding principles I used to help direct my brainstorming, quoting again from the demo site:

>**Words first.** Dynamic documents are still fundamentally documents, and in the spirit of [literate programming](https://en.wikipedia.org/wiki/Literate_programming), I think language is the most flexible and versatile tool we have to communicate ideas. Other dynamic, interactive components should augment prose rather than dominate them on the page.
>
>**Direct representation, direct manipulation.** When designing visualizations, we should favor representing quantities and concepts directly using shapes, graphs, animated motion, or whatever else is best suited for the idea that needs to be communicated. Likewise, when building input components, we should establish as direct a connection as possible between the user's manipulation of the controls and the data being fed into the simulation or visualization -- motion in time for time-series input, a rotating knob for angle input, a slider for size or magnitude, and so on.
>
>**Embrace programming and programmability.** What makes programming challenging is not the notation or syntax, but the toolchain and mental models that are often required to build software. I believe that with simple build steps and a well-designed API, the ability to express dynamic ideas with the full power of a general-purpose programming language can be a gift.
>
>**Composability.** Oak Notebook should come with a versatile set of inputs and widgets out of the box, but it should be easy and idiomatic to compose or combine these primitive building blocks together to make larger input components or reusable widgets that fit a given problem, like a color input with R/G/B controls or a reusable label for displaying dates.

In particular, I think "words first" is a good grounding principle whenever I think about making documents more interactive or smarter or amenable to machine understanding. Text documents are the skeleton to which authors can add color and texture with data displays and interactive explorations, but if we try to collapse narratives and ideas in prose down to simple knobs and graphs, the whole house collapses.

At a technical level, the Oak Notebook compiler is a [single Oak script that packs a serious punch](/highlight/https://raw.githubusercontent.com/thesephist/x-oak-notebook/da1513475af03504782040a448cc395c74b6263a/compile.oak?start=17&end=122). It parses Markdown using Oak's `md` standard library module. It calls out to `oak cat` to syntax highlight any Oak code blocks. It also shells out to child processes of `oak build --web` to compile a dynamically generated Oak program to JavaScript that runs on the Oak Notebook page. It's a great testament to the hackability of Oak and Oak's self-hosted tools, and I had a lot of fun writing the script.

I don't really know where this Oak Notebook project will go. It may go nowhere, and become a little tool I occasionally reach for to make fun demos. I think it would be pretty cool if it can become a part of the standard Oak toolchain -- that would be one step on the path to making Oak itself an ["erector set" for thinking](https://twitter.com/jessmartin/status/1451198781702111250). Perhaps it would even become a starting for more interesting and creative work on how we can write and share living, programmable documents to communicate bigger ideas better.

