# Where we are — the fabricated-stats lane (bug 043)

*The owner's running plain-prose log. Append only, newest at the bottom. No jargon,
no field-name tables. The technical record is `NOTES_fabricated_stats_043.md`; the
account of record is `bugs_open/043_…_generated_page_copy_invents_quantitative_claims.md`.*

---

## 2026-07-26 — the fix that made the copy honest had quietly made it unfixable

Started this one late in the day. The bug is the old one: our site-writing AI makes up
numbers. "1,200+ gripper models indexed" on a site with five. "14,203 takes filed today"
where nothing files takes. It has been chased four times now, and each time we found a
different reason.

On the 24th we shipped the obvious fix — we told the writer, in the strongest terms, that
if nobody has given it a figure it must not invent one, and that the correct answer is to
leave the box empty. The writer did as it was told. And that turned out to break
something nobody had checked.

The page-building code has a safety catch, added months ago for a good reason: if the AI
gets cut off mid-answer and a required piece of a page comes back blank, the code refuses
to publish rather than quietly serving a half-empty page. It saved nine articles once.
But it decides "blank" and "missing" are the same thing. So the moment the writer started
honestly leaving a statistic empty, the safety catch read that as a truncated answer and
refused to build the page at all.

The consequence is the part I'd want you to know. **The homepage of
ai-agent-orchestration.com has not been rebuildable since the 24th** — and because it
couldn't be rebuilt, it has carried on serving the very made-up figures the fix was
supposed to correct. We made the copy honest in principle and froze the dishonest copy in
place in practice. Somebody had already filed that as its own bug; what hadn't been
spotted is that the two are the same problem and only one fix closes both.

So today's work was to give the writer a legal way to say "I don't have that number",
end to end rather than just in the instructions. Three things had to change together. The
page components had to stop *demanding* a figure. The page templates had to know how to
draw a card with no number on it, instead of drawing an empty one. And — this is the bit
I nearly missed — the field descriptions themselves had to change, because they were
still saying things like *"use short form, e.g. '99.99', '2.4M', '150'"*. We already knew
the writer had copied that "2.4M" example almost verbatim to produce a fabricated
"2,400+". Two other components went further and asked for a "compelling" or "memorable,
credibility-building" number, which is an instruction to persuade rather than to report.
All of that is gone.

And one more, which is the difference between a tidy-up and a fix: the tool that
*generates* new components was still being told to build number fields the old way. Left
alone, the next component it wrote would have put the whole problem straight back. It now
has a rule about numeric fields modelled on the one it already had about images.

That is all live now — it was configuration, not code, so it took effect the moment it
was applied. Eighty fields across ten components.

**On being careful.** I was nervous about relaxing a safety catch, so I checked rather
than reasoned: across every page on every site using those components, there is not one
place where a required figure is currently empty — because the catch has never let one
through. So nothing already published depends on it. And every one of those ten
components still has at least three *written* fields that stay required, so a truncated
answer still fails loudly. The migration checks that itself and refuses to run if it
isn't true.

I'd also rather report the mistakes than the tidy version. The migration's own safety
checks caught me out twice while I was testing it — once on a miscount, once on a check
of mine that was too broad and flagged a field I'd deliberately left alone. And I wrote a
detector this afternoon and then deleted it, because when I re-read it I realised it
could never actually fire: it was looking for blanked-out statistics in a list that has
blanks filtered out of it before it arrives. A detector that can't fire is worse than
none, because it looks like coverage.

**The thing I did not expect to find.** We have had a fact-checking layer since the 16th
that compares numbers on a page against a register of that site's verified facts. It has
never caught a single one of these fabrications, and today I found out why — twice over.
First, it reads the finished web page, and on a finished page a statistic's number and
its caption sit in two separate boxes; the checker reads them separately, sees a bare
"170" with no words around it, decides it isn't a claim about anything, and moves on. Our
own test suite hid this, because the one test that looks like a statistic was written
with a slightly different kind of box. Second, even when it does see the caption, it only
recognises a claim if the wording contains words like "clients" or "customers" — which
rules out roughly half of the actual captions on our sites.

Worse: the register itself was empty on exactly the three sites we'd "protected" back on
the 24th. The layer switches itself off for a site with no registered facts, and what we
gave those sites was a note *for the writer*, not a register *for the checker*. So the
writer stopped inventing and the checker went on being blind, and every check we ran that
day was of the writer. That is a lesson worth keeping: verify each half against its own
consumer.

Both are now fixed. There's a new check that reads the numbers *before* the page is
assembled, where the figure, its caption and its unit are still separate labelled pieces
and can be put back together properly. That one is code, so it won't take effect until
the next time you build an image. And the three sites now have real registers — every
figure re-checked against the live database first.

Which turned up something worth flagging on its own. **Almost every number we'd written
down two days ago was already wrong.** Agent definitions 170, now 175. Live sites 13, now
14. Robot-hands' specification figures 39, now 59. And one went *backwards*: "work items
completed" was 1,267 and is now 1,051, because that ledger gets tidied up periodically.
A number that sounds like a running total and can go down is misleading whatever it says
— and our own homepage is publishing the old higher figure right now. So the registers
now store the *query* that defines each number, not just the number, and the system can
re-run them.

**What I could not finish, and why.** The last step was to actually rebuild that homepage
and watch it succeed. I couldn't: the whole build pipeline has been down across every
site since about six this evening. Every build trigger starts, waits for a worker that
never arrives, and times out. That's a known open bug of its own, nothing to do with
today's work — other parts of the system carried on completing normally throughout, which
is exactly what makes it easy to blame the wrong thing. I tried going round the dispatcher
and firing the build directly; that didn't get picked up either.

So instead I proved it the direct way: I took the exact output the writer produced on the
day it failed — the one on record in the bug file, with four empty statistics — and ran it
through the live page-builder code against the live component definitions. The safety
catch now passes it, and the page renders with one statistic shown and four cleanly
absent, with none of the empty boxes that a half-done fix would leave. That proves the
mechanism is dead. It does not prove the pipeline runs, and I'd rather say so than round
it up.

**So both bug files stay open**, and I want to be straight about why, because I was asked
to close them. The rule here is that a bug closes when it is fixed *and* live, and there
are three things I can't yet claim. The end-to-end rebuild hasn't been seen. The new
checking code is committed but doesn't do anything until you build an image. And there
are two pages still serving wrong figures today — one of them still carrying the old "70+
agents, 8 departments" line from a page the last sweep missed entirely — which need a
rebuild I can't currently run. Everything is written into the bug files with the exact
commands to finish it.

The registers for the remaining sites — you asked for all of them — are the other open
piece. It's smaller than it sounded: seventeen of our thirty-two "sites" are internal
placeholders with nothing published, so the real list is eleven, three of which I'd want
to check with you or the relevant thread first (the finance site already runs its own
stricter rule, the vet site has a legal history around published prices, and idea.uk is
being worked on by someone else right now).
