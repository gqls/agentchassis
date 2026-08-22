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

## 2026-08-22, later — I had the danger backwards, and the truth is more interesting

I told you earlier that thirteen of these pages were primed to be destroyed: that a routine rebuild
would throw away the calculator and leave a title band. **That was wrong, and I want to correct it
plainly rather than quietly.**

What actually happens is the opposite, and it explains why this has been sitting there for months
without anyone noticing. When the platform rebuilds one of these pages it *does* generate a banner
for that slot — and then a safety net catches it. That net was built after an earlier accident where
rebuilds silently blanked working content, and it does its job: it sees that what is stored in the
slot is an interactive tool, sees that the freshly generated replacement is not, and puts the tool
back.

But it puts the tool back **under the banner's name**. So every rebuild faithfully preserves the
calculator and faithfully preserves the lie about what it is. **The mechanism that protects these
pages is the same mechanism that keeps them mislabelled.** I proved it on the vetcomparison
homepage: six rebuilds in the last four days, the tool intact every time, and the current database
rows were written *during* the most recent one.

The practical consequence is about ordering. These pages regenerate themselves, so tidying up the
existing 22 first would be wasted — the next rebuild would undo it. The producer has to be fixed
first, then the existing pages, and then the safety catch can be switched on.

## What the plan comes to

Four pieces, and only the first three are things I'd like to build now.

**Stop it happening.** When the platform meets a page it cannot identify, it should say so instead of
guessing from the plan. Today the "I don't know" placeholder gets silently upgraded to "banner"
because banner is first in the list. Instead the fragment gets its own honest type, with a template
that simply re-emits whatever is stored — which has the nice property that these pages become
genuinely rebuildable for the first time, rather than merely being protected from rebuilds.

**Check it.** Every page save funnels through one database write. Components stamp their own name
into their HTML, so at that one point we can compare what the row *claims* to be against what it
*is*, and record the disagreement. I want it to **record only** at first and not block anything —
there is a change from yesterday on the neighbouring code that made exactly this call, and it was
right to. The ability to actually refuse ships switched off, with a switch that has to be turned on
deliberately, one caller at a time.

**Look everywhere else.** Six different pieces of code write these rows and I have only fixed the
path used by one. A daily check covers the rest without touching them.

**Then repair the 22 — and that one is your call, not mine.** It changes what four live sites serve.
The repair itself is designed to change nothing a visitor sees: the same bytes, correctly labelled.
But 277 spent a fortnight learning how expensive it is to be casual about this class of change, so I
would rather show you the check-and-restore script and the before/after of one page than just run it.

## Where it stands right now

The plan has gone to the review council (reference 62aac6c2) and is being reviewed as I write. A
separate diagnostic run is still queued behind the fleet on the one question I could not settle from
the database alone. Nothing has been changed on any live site, and no code has been altered yet —
what exists so far is the diagnosis, the measurements, and the plan.

Two things I got wrong today are written down in the shared mistakes log with the checks that would
have caught them, and both corrections are recorded in the bug file itself rather than edited away.
