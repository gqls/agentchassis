# Where we are — component identity (`bugs_open/357`)

The owner's running plain-prose log. Append only, newest at the bottom. No jargon where plain words
will do.

---

## 2026-08-22, afternoon — asked for 277, found it already closed, moved to its successor

You asked me to look at bug 277. It had already been closed a few hours earlier, on your own
ruling — so rather than take the file's word for it I checked the live system: the routing it
promised is there, everything of that type is classified, and the two pages it used as proof still
serve what they are supposed to. The close is real. Its own closing note says where to go next, and
that is bug 357, which nobody is working on because the lane that filed it finished at the same
time.

## What 357 is about, in plain terms

A web page here is stored as a list of blocks. Each block says which *kind* of block it is — a
banner, a text panel, a calculator — and stores the HTML it should show. The kind matters because
everything else in the platform reasons about it: what fields the block needs, whether it can be
rebuilt, what to check it for.

On 22 live pages, the block is lying. It says it is a **banner** (the "hero" — a title, a sentence
and a button) and what it actually holds is an **entire working calculator or interactive tool**,
ten to twenty thousand characters of it. The block was born that way. Nothing overwrote it later.

## Why it happens, and it is a simple mistake with a big shadow

When the platform saves a page, it splits the HTML into blocks and works out what each one is. Tool
pages don't arrive in the shape it expects — they come as one single lump with no internal markers —
so the splitter gives up on identifying it and labels it with a placeholder meaning *"I don't know
what this is."*

Then the next step does something reasonable-sounding and wrong. It sees the placeholder, looks up
the page's plan, and fills in the name **by position**: this is the first block, the plan says the
first block is a banner, therefore this is a banner. It never looks at the thing itself. And because
every one of these pages plans a banner first, every single tool gets labelled a banner.

So the honest answer "I don't know" is quietly converted into a confident wrong one. That is the
whole bug, and it is the part worth fixing properly rather than case by case.

## Two things I found that the bug file does not say, and one of them is worse than it looks

**It is still happening.** The bug file describes nine pages, the newest from a fortnight ago. Its
own re-runnable query actually returns **22**, and the newest was created **today** — the homepage
of vetcomparison.uk. That page currently opens with a vet-practice search tool sitting where its
banner should be, and the banner text that was written for it is stored but never shown to anyone.

**Thirteen of them are already primed to be destroyed.** The bug file's reasoning is that these
pages are safe as long as nobody writes data to them, because writing data would let the platform
decide it can rebuild the block from scratch — and rebuilding it would render the *banner* it claims
to be, throwing away the calculator. That is right for the nine it looked at. But thirteen of the
other rows **already have that data**, complete and valid, sitting next to the tool. Nothing more
needs to be written. A routine rebuild of any of those pages would replace a working calculator with
a title band, and sixteen of the 22 are on the rebuild-eligible setting.

I want to be straight about one thing I could not establish: whether this has already happened to
some page. If it had, that page would look normal now and would have quietly dropped out of every
query I can run, because there is no systematic history of these blocks to check. I have candidates
but no proof, and I am not going to claim it in either direction.

## What I am proposing to do about it

The fix I want is not "relabel these 22 rows". It is to stop the platform ever again asserting what
a block is without checking. The encouraging part is that the check is cheap and I have measured it
rather than guessed: components stamp their own name into their HTML, and comparing the two across
every block on the fleet gives **1,550 that agree, none that disagree, and 27 where the stamp is
simply missing** — which is this problem plus five other rows worth looking at. The obvious cruder
test flags 158 and is useless, because it also convicts pages whose styling was legitimately edited
later. So there is a precise test available, it does not cry wolf, and the 1,550 agreements are what
tell me the test is actually doing something rather than passing everything.

Detailed plan being drafted now. I will come back to you before anything is applied, and separately
on the question of what to do with the 22 existing pages — that is a decision about changing what
live sites serve, which is yours and not mine, and 277 has just finished teaching everyone how
expensive it is to get that decision wrong.
