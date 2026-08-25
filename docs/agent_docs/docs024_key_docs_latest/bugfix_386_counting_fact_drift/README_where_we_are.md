# Where we are — counting-fact drift (bugs_open/386)

Plain prose, append-only, newest at the bottom.

---

## 2026-08-25 — what this bug actually is, and why your ruling changed the plan

The short version of the bug: some of the numbers on our sites are live counts — "feed items
collected: 11,513". We keep those counts in a register, and a job re-reads them every night. When
the count ticks up, the register moves and the page does not. Our own honesty checker then compares
the page against the register, finds a number the register no longer holds, and reports the page
for making up a figure. Nothing false is published. The page is simply a day old. But the report is
filed at a severity that refuses a rebuild, so the page cannot be rebuilt to fix itself, and in the
review queue it looks exactly like a page that invented a number.

It re-arms every night on its own, which I could show rather than assert: when the bug was filed on
Sunday the register held 11,646; today it holds 11,828. The page still says 11,513.

The good news is that this is small and precisely bounded. Of 295 facts across the whole estate,
only 29 are the kind a nightly job moves, and only 13 of those are matched strictly enough to
convict a page. They sit on six sites — fundamentallyai, leopardess, robot-hands, vonc, and two
others. This is not a fleet-wide rework.

Then your ruling arrived, and it reframed the thing properly. I had been asking "how do we make the
checker tolerate a number that moved". You asked the better question: should the page be printing a
live counter at all? A counter's honest form is a lower bound. "We have collected 11,646 items" is
false a minute later; "at least 11,000" stays true for months and needs nobody to re-render
anything. So the default is now "at least N", or cut the number entirely where the page does not
need it.

One caution I want to flag before I act on it, because it is the kind of fix that can quietly make
things worse. Switching a fact to "at least N" tells the checker to accept any number below N that
appears near certain words. Those words are matched loosely — as fragments, not whole words — so if
the words are broad, the fact ends up vouching for almost any number on that part of the site. We
have a live example: one fact on ai-agent-orchestration.com is set to "at least" with the single
word "orchestration" attached, and it currently blesses any figure up to 7,281 sitting anywhere
near that word. Your ruling already warned about this and quoted the figure as 4,068 — it was 4,068
when someone wrote it down, and it is 7,281 now, which is the same bug biting the document that
describes the bug.

So the order of work is: measure what each proposed "at least" would start accepting, *then* change
it — never the other way round. Where a page does not need the number, I would rather delete it than
register anything, which is your stronger option and the only one with no ongoing cost.

After that there is a second, deeper fix, which your ruling explicitly left standing: teach the
register to remember the values it used to hold, so a page that printed last Tuesday's number is
recognised as having been right on Tuesday rather than accused of invention. I checked whether that
is even buildable and it is — we never actually delete the old register entries, we just mark them
superseded, so the whole history is still there: 315 old versions going back to mid-July. The
number the convicted page prints, 11,513, is sitting in that archive dated Sunday. So the history
does not need to be reconstructed by guesswork; it is already on disk.

Two smaller things I corrected on the way through, both cases of a document being more confident
than the code. The bug file proposed marking counting facts as "always goes up" — but ours don't
always: one of them falls whenever the table behind it is cleared out, which I found in our own
migration history. And two documents say we can tell a stale page from an invented number by
comparing timestamps on something called `nearest_fact_id`. That field does not exist in the code
at all; it is a label our audit model writes in its own output, so the comparison the documents
describe cannot happen where it would matter. I will correct both in place rather than only noting
it here.

Nothing has been changed on any live site yet. The only thing written so far is this lane's own
notes.

## 2026-08-25 — the number on the broken page was never written by the writer

An hour after I wrote the plan above I checked something I should have checked first: how the
number on the convicted page actually got there. It changes the order of the work, so I want to be
straight about it rather than quietly re-planning.

Two things came out of it.

The first is good news. Your ruling — say "at least N", or don't print the number — is already
running in production on five facts, and somebody arrived at it by hand before you ruled. One of
them reads "more than 10 live production sites", with the live count of 26 kept available behind the
scenes and an explicit note telling the writer to state a floor and never the exact figure. That is
your ruling, working, on a live site. So Phase A copies a template we already have rather than
inventing one, which makes it a much smaller and safer job than I had assumed.

The second is more awkward. The five facts that actually convict the fundamentallyai page carry no
writer instruction at all — they never told the writer anything. The numbers on that page came from
a chart component whose whole job is to render the register, and the figure is frozen into the
stored component as it was on Saturday. So "phrase it as at least N" has nothing to attach to
there: there is no sentence, only a chart.

Which means, for the case that started this bug, your ruling's *other* half is the one that applies
— the part you explicitly left standing, where the register learns to remember the values it used
to hold. And it is the right answer for a chart rather than a workaround, because the chart already
prints its own date: "11,513, verified 23 August" is true for ever. Nobody needs to re-render
anything. The register just has no way to agree with a page about last Saturday.

There is also a smaller correction worth having. I said the problem covered thirteen facts. It
covers far fewer in practice. Most of those thirteen are counts of things you could list on your
fingers — eight archetypes, six manufacturers, four guides — and the writer instruction for several
of them names every item. An exact count is the honest form there, and converting them to "at
least" would weaken the checker for nothing. The genuinely fast-moving ones are four figures on
fundamentallyai and two on leopardess. That is the real size of it.

Still nothing changed on any live site.

## 2026-08-25, later — the fix is built, and one thing you should know about how it got committed

The durable fix is written, tested and submitted for review. What it does, in one sentence: the
register now keeps the readings it used to hold, so when a page prints "11,513 verified 23 August"
and the counter has since moved on, the platform recognises that as something that was true when it
was written rather than a number somebody made up.

It is off by default and stays off until someone deliberately turns it on for a specific fact. That
is on purpose — the thing being turned on is "accept more numbers", and that should never be the
default. It also means that when this ships, nothing will change anywhere until we arm a fact, and
a quiet result after the roll is the expected outcome rather than proof it works. I have written
that at the top of the register entry so nobody later mistakes silence for success.

Fourteen tests. More usefully, I broke the code seven different ways on purpose — in a throwaway
copy of the repo, so no broken version ever sat in the shared tree — and checked that each break
made the right test fail. A test you have never seen fail is not evidence of anything, and there is
a case in our own history of a test in this exact area that passed with its fix removed.

I also had to correct a colleague's session twice today, and it corrected me once, which is the
system working. It had concluded my fix would rescue a different case — a claim like "over 1,600 a
day" when the real figure dips below that. It does not, and it should not: making it do so would
mean the platform vouching for the highest number we ever recorded, for ever, on every quiet day
afterwards. The right answer for that case was already sitting in the register — the instruction
there says "over a thousand", which is below the lowest figure we have ever seen, so it is safe
permanently. That is the owner's own ruling working exactly as intended; the copy on the page had
just drifted above it.

The one thing worth flagging. My change to the main file was committed by another session rather
than by me. Several of us work in one shared copy of the code, and while I was testing, a colleague's
commit swept my half-finished work in alongside their own. Nothing is lost — I checked the code
arrived intact and the tests still pass — but it means my commit message describes a mechanism that
is not actually in my commit, and the review reference is attached to the wrong one. There is no way
to correct that after the fact without rewriting history, which we do not do here, so I have written
down where everything actually landed and told the other session. It is a known hazard of the shared
tree rather than anything that went wrong today, and the practical cost is that an audit report will
show a gap that has a written answer.

Still nothing changed on any live site.
