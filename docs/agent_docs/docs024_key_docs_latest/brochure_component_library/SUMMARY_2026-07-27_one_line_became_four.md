# Summary, 27 July 2026 — one line became three, and then four

*Written to be read aloud. Previous in the series:
`SUMMARY_2026-07-26b_the_chart_is_built_and_what_it_cost.md`.*

---

## What we're trying to do

Build a consultancy brochure site, fundamentallyai.com, that markets this platform
using only things the platform has actually done — and build it *through* the
platform, so that everything the site needs is a capability the fleet gains rather
than a one-off. The rule that makes it hard is also the point: no figure appears on
the site unless the system can re-derive it from a real query. That is what the
site is selling.

Along the way we build the reusable components a consultancy site needs, and every
time the platform can't do something we need, we fix the platform rather than
working around it.

## Where we've come from

The site went live on 22 July. Over the following week we built five interactive
components, proved they survive a full rebuild, and repaired the site's links twice.
On 26 July we shipped the piece the brief had always asked for and nobody had ever
built: a chart component whose numbers come from the evidence register, so a chart
*structurally cannot* display a figure the system can't defend. That needed no new
Go code, which was the pleasant surprise of that day.

It shipped on the home page only, because of a bug we found while building it: no
component could tell which page it was on. The chart carried definitions saying
"this one belongs on capabilities", and the platform had no way to honour them.

We filed that as bug 085 and wrote in the file, twice, that the fix was one line.

## What we've done

**We fixed 085, and it was not one line.**

Following the value from where it enters the system to where a component could read
it — rather than reading the one function that looked wrong — showed it is dropped at
**three** separate points on a single journey. The page's name is passed in by
configuration and thrown away by a filter; the thing that saves the context for the
next step doesn't save that field at all; and the thing that restores it on the other
side doesn't restore it either. Every one of those looks entirely reasonable when
you're looking at it. Had we shipped the one line we filed, nothing visible would have
changed, and the next person would have reasonably concluded the diagnosis was wrong.

What caught the other two was asking the database what the saved context actually
contained, rather than trusting the code. The field wasn't empty — it was *missing*,
which is a different fact and points at a different function.

**The review council sent it back, correctly.** Its objection was that we had fixed
the field and not the mechanism that dropped it: the next field anyone adds will
vanish the same way, silently. So the fix now also carries a check that fails the
moment someone adds a field the templates can see but the pipeline can't deliver —
and running that check found **three fields already in that state**. Those are filed
as their own case rather than quietly bundled in. Round two was approved.

**Then it shipped, and the live test failed.** The new build went out this afternoon.
We confirmed our code was genuinely in the running binary, fired the cheap re-render
that should have proved it, and the page *still* showed charts belonging to another
page. Not a bad fix and not a bad deploy — a **fourth** drop point, on a different
route through the system. The route we had fixed is the one used when a page is
written from scratch; there is a second, cheaper route used when a page is
re-rendered without rewriting its words, and it builds its own context and never
included the page's name. The name was sitting in a variable one line above the call.

That fix is written, tested and with the council now.

**Separately, the em-dash question turned out to be measuring the wrong thing.** You
were asked to choose between a mechanical pass over the finished text and fixing the
two components producing most of them. Re-measuring first showed the count was adding
together two different things: em dashes the writing model produced, and em dashes
typed into component templates and reprinted on every render. No instruction to the
writer moves the second kind, and neither would the mechanical pass, because the pass
would work on text and these aren't in the text.

Split properly: 66 on the site, **43 written and 23 baked into templates**. That
overturns both halves of what we told you. The capabilities page, reported as showing
no improvement at all, is four template ones and two written ones. And one of the two
components we named as the main offenders contributes nothing from the writer at all.

## Where we are now

The chart is live on the home page and correct. Its figures refreshed themselves
overnight without anyone retyping a number, which is the whole design working.

Bug 085's first fix is **live in production and verified in the binary**. It is not
yet *proven* to work end to end, because the cheap way to prove it is the route we
have just discovered was also broken. The second fix is written and awaiting review;
it needs the next build to take effect. Once it does, proving the whole thing takes
about two minutes and costs nothing.

The wider mechanism — four hand-maintained lists that nothing checks for agreement —
is filed as its own case, unowned, with the three fields it is currently dropping
named.

On the voice question, our recommendation is now **neither of the two options we
offered**: leave the writer alone, because where it was re-run it roughly halved, and
fix the templates instead. Three components hold 21 of the 23, and two of those are
*generated* tool pages — so the same em dashes will be printed into the next tool we
generate unless the generator's own instructions are fixed. That is the durable win
and it is a small job.

On the decision-record page, nothing technical is blocking it and we are not going to
decide it for you. What we have done is put the real numbers in front of you,
including the one that cuts both ways: our review council approved nothing at all for
ninety-one consecutive rounds, because the council's own decision rule had a bug we
later found and fixed. That is simultaneously the best argument for the practice and
the most quotable line against us.

## Where we're going

Next, in order:

1. **Prove 085 end to end** once the second fix ships — one scoped re-render of the
   home page should leave one chart where three stand now, and the capabilities page
   should then carry the two that belong to it. We will also deliberately test the
   case that *should* show nothing, because a green happy path proves a deploy and
   not a fix.
2. **Restore the capabilities chart**, which has been held back since the day it was
   built and needs no data change.
3. **The template em dashes**, if you want them gone — including fixing the tool
   generator so new tools are born without them.
4. **The decision-record page or the softened sentence** — your call. The live page
   currently says self-correction is "something you can read", and nobody can.
5. The dark-theme render of the chart is still unverified, waiting on leopardess
   adding a charts key to its own register. No code needed; nothing blocked.
