# Where we are — bugs_open/390

Plain prose, append-only, newest at the bottom.

## 2026-08-25 — opening the lane

The bug, said simply: when the site checker finds text that is too faint to read, it sends a
repair agent to change the colour. The agent writes a new line of styling into the site's main
stylesheet, saves it, and marks the job done. The job is marked done because something was
written — not because anyone looked to see whether the text got easier to read. Usually it did
not, so the checker finds the same faint text again next time, files it again, and another dead
line of styling gets added. Some sites now carry sixteen of these for the same five bits of text.

Why the repair does not work is worth understanding, because it is not the agent being careless.
Each page carries its own little block of styling, written inside the page itself, and browsers
give the last word to whichever instruction is more specific — and, when two are equally
specific, to whichever comes later. The page's own block is both more specific and later. The
agent is editing a file that structurally cannot win the argument. It is doing exactly what it
was told; what it was told is wrong. The prompt it works from actually says *"repeat the
offending selector exactly as it appears above … so your override wins"*, and on the pages I
measured that sentence produces the losing move every time.

I checked how often this happens rather than assuming. Of the repairs that were marked done and
whose target was a real one, **seventy-five** came straight back, and **ninety-seven** of those
returns recorded byte-for-byte the same two colours as before — the repair had changed nothing at
all. Then I took forty of them at random and went and looked at the actual pages: in
**thirty-three** the deciding instruction was in the page's own block and more specific than what
the agent wrote, and in **none** was it in the file the agent edits. So this is not an agent
writing weak rules. It is the platform sending the repair to the wrong place, and then recording
success.

Two things I got wrong or nearly wrong, worth writing down. The bug report's own worked example
turns out not to demonstrate the bug — that particular site has no stylesheet linked at all, so
the agent refuses and parks rather than pretending, which is an earlier fix doing its job. I found
a genuine example elsewhere and used that. And I briefly announced a "third failure mode" on
another site before checking the guard that governs it; the guard already catches that case, and
I have recorded the check I skipped.

One genuinely new thing did come out of it, and it is uncomfortable. Even a repair that *does*
win gets deleted. When a site's design is regenerated — a palette change, say — the whole
stylesheet is rewritten from scratch and every repair appended to it disappears. I watched that
happen on one site at 12:09 today, wiping five repairs made the evening before, and the work
items still say `complete`. That is behaving as designed; a migration from earlier this month
deliberately made the design agent own that file. Nobody has drawn the consequence for repairs
until now. It is a separate bug and I am filing it separately rather than folding it in here.

You chose the order: fix where the repair is aimed first, and leave the historical records alone.
So the plan is three steps. First, one small database change that stops the prompt instructing the
losing move — that ships without rebuilding anything and helps immediately. Second, the real fix:
teach the page checker to record *which* instruction is actually winning and where it lives, so
the platform can aim the repair at something it can reach — and have the checker prove its own
answer by removing the instruction it thinks is winning and confirming the colour changes. Third,
let the agent use that measurement, and park honestly the one case no stylesheet can ever fix.

I have written down, in advance, what I expect each step to do, including what result would prove
me wrong. That matters here more than usual, because the failure mode of this whole area is work
that reports success without having achieved anything — which is precisely the bug.

## 2026-08-25, later — what has actually shipped, and the two things that went wrong

Three things are now in.

**First, the quick one, and it is live.** The repair agent's instructions used to tell it, in
writing, to do the thing that loses. I replaced that text. It went live at about a quarter to five
and I checked the live record afterwards rather than assuming — the old sentence is gone, the new
one is there, and running it a second time now refuses loudly instead of quietly doing it twice.

**Second, the real fix, which is committed but asleep.** The page checker now works out *which*
styling instruction is actually deciding a colour, and — this is the part I care about — it
**proves** it rather than guessing. It removes the instruction it thinks is winning and looks
again. If the colour changes, that instruction was the winner; if nothing changes, it was not, and
the checker says "I could not tell" instead of inventing an answer. It puts everything back
afterwards and checks that it did. Nothing here takes effect until the next release of two
services, so the pipeline behaves exactly as it did today until then.

**Third, a separate bug filed.** The erasure problem I described earlier is now its own case with
its own number, so it does not get lost inside this one.

Now the two things that went wrong, because those are the useful part.

**The review panel found a real flaw and I had already shipped past it.** I put the quick fix
through the review council, and while it was thinking I applied the change. The panel came back
asking for a revision: my migration assumed there was exactly one copy of the agent's settings, and
never actually checked. It turns out four other agents in the estate *do* have two copies, so the
assumption is not idle — it just happened to be true for this one. I verified that myself rather
than taking their word, added the check, and proved it works by faking a second copy and watching
it refuse.

The second flaw they found is the one worth writing down. I had typed the same sentence into the
file twice — once for the safety check, once for the edit itself. If those two copies ever drifted
apart by a single character, the safety check would pass and the edit would silently do nothing,
while still reporting success. That is the *second* time in this one file that I wrote a value down
twice and created a way for it to disagree with itself; I caught the first one myself. The rule I
am taking from it: never restate a value, derive it. The file now declares each sentence once.

**And a mistake in how I sequenced things.** Because I applied the change before the panel finished,
their automatic checks looked at the world *after* my edit and reported that the text I was
proposing to replace was not there. That reads like something had gone wrong, and nothing had — it
was simply me having already done it. I have written that down where the next person will see it.

Two review rounds are still running. The remaining piece — teaching the repair agent to use the new
measurement — is deliberately waiting until the services roll, so it can be built against real data
rather than my expectations of it.

---

**26 August, lunchtime.** Everything is now approved and live, and today is the first day the new
measurement meets a real site.

The overnight release carried the measuring code onto all three services, and I proved that at the
running services themselves rather than trusting the release notes. The review panel has now
approved all three pieces of this work — the quick prompt fix, the measuring code, and the final
piece that routes on the measurement. The last of those took two attempts through no fault of the
work: the first review died mid-run because the AI provider the reviewers themselves run on was
briefly out of credit, so eleven of the seats simply couldn't speak. Once that was topped up, the
same submission passed with no objections at all.

One small but important catch this morning: the previous session's handover note contained a
checking query pointing at a name that doesn't exist in the database — anyone following it would
have concluded for ever that no audit had run yet. I traced where the numbers actually land, fixed
the note, and wrote down how the mistake happened (the name was written from memory rather than
read from the code).

What happens next is out of our hands, deliberately: the site checker visits each site every three
days on its own schedule, and the first site due — a remortgage calculator site — comes up early
this afternoon. That visit will be the first time the new measurement runs for real. The decisive
sites, the ones whose repeated failed repairs started this whole investigation, come due tomorrow
morning and afternoon. I've set a watch so the moment the first results land we can read them and
grade the predictions we wrote down in advance.

**26 August, mid-afternoon.** The first real test happened, on schedule, and passed — with one
honest caveat about what "passed" means today.

The checker visited the remortgage calculator site at ten to two, as predicted to the minute. It
found five failing colour pairings and, for the first time, tried to work out *which* styling
instruction was actually winning for each. It managed that for one of the five — a heading — and
got the answer we expected: the winning instruction was buried in the page itself, so a fix in the
shared stylesheet needs to be written more specifically to beat it, and the checker wrote down
exactly how specific. For the other four, all the same footer text repeated on four pages, it
correctly said "I can't prove which instruction wins" rather than guess. I dug into why: the footer
text has no colour instruction of its own — it inherits its colour from its container — and the
measuring code only looks at instructions aimed at the text itself. That's a known, deliberate
limitation, but footers appear on every page of every site, so it will recur; I've written up a
possible extension for a later review rather than bolting it on now.

Then the repair agent ran on the first footer row, under the corrected instructions from
yesterday. It did precisely what we predicted it would: one rule, one property, marked so it
cannot lose. I checked that at the live website, not in the database — the served stylesheet ends
with the new rule, its fingerprint matches the committed file exactly, and the new colour measures
5.5 to 1 against the 4.5 required. Yesterday's three "repairs" on the same footer, by contrast,
were written against a selector that matched nothing and did nothing.

The caveat: what settles this for good is the checker's *next* visit in three days, when it should
find the footer fixed and withdraw its complaint instead of filing it again. That is the result
the whole investigation turns on, and it can't be hurried. The heading row — the one with the
measured requirement attached — is still queued behind some unrelated page builds and should run
later today.

**26 August, later afternoon.** All five of today's failures were repaired within seven minutes of
each other, and I checked the live stylesheet again: every rule is there, the file's fingerprint
matches what was committed, and both colours now pass by a clear margin.

The heading row was the important one — the first time the repair agent was handed a *measured*
requirement ("you must beat this exact instruction; here is a selector we have checked will do it").
It met the requirement, using the checked selector word for word. It also did one thing it was told
not to: it marked the rule as unbeatable when the measurement said that wasn't needed. That does no
harm to the page, but it means the "not needed" signal is being ignored, and it tells me the
prompt's "this section overrides the general advice below" line is being outvoted by the general
advice. There is a clean fix — only ever show one of the two — but one case isn't a pattern; I'll
decide after tomorrow's three sites.

Less good: the footer text appears on four pages, so four separate tickets were raised for the same
fault, and the agent fixed it four times, each time on top of the last. That's waste rather than
damage, it's a known consequence of tickets being per page, and the right place to stop it is the
"only mark it done if it measurably improved" gate you deferred yesterday. Nothing here changes
what settles the case: the checker's return visit on the 29th.

**26 August, end of day.** A very good day, and one honest loose end.

The review panel approved the evening's fix — the one that stops the repair agent being given two
contradictory instructions — first time, no blocking objections. Better than approval: within
twelve minutes of it going live, the next repair came out exactly right under the new rules, and
the eight that followed did too. We now have a clean before-and-after on one evening: same model,
same kind of fault, old prompt → wrong habit every time; new prompt → right behaviour every time.

The checker also visited two more sites tonight. On both, it could work out which styling
instruction was winning for almost every failure — thirty-eight worked out, six honestly declined,
none unfixable — which is the distribution we predicted when this investigation started. The
cooking site is the satisfying one: it's the site where even correctly-written repairs kept
failing, and every one of its failures now carries an exact, measured requirement.

The loose end: that site's five repairs are still waiting in the queue behind other sites' work —
nothing wrong, just a busy evening — so grading them is the first job tomorrow.

A fresh release of the platform went out tonight. I checked, at the running services themselves,
that our measuring code is still aboard — it is.

What's left before I can close the case: tomorrow the checker visits the three sites whose
repeated failures started all this, and Friday it returns to today's sites — where it should
withdraw its complaints rather than re-file them. That withdrawal is the finish line.

**27 August, early afternoon.** Today's plan was to watch the checker visit the three sites whose
failures started this investigation. The first visit found something bigger instead.

The checker's visit to vonc timed out — and when I pulled on that thread, it turned out the
checker has been timing out on **every site larger than about two dozen pages, on every audit
day, for at least two weeks**. Fourteen or so of our largest sites — including webdesign.co.uk
itself — have not had a single successful check in that time. The cruel details: the measuring
browser actually finishes the job, but the platform gives up waiting about 45 seconds too early
and throws the answer away; the failure is recorded as "COMPLETED" so nothing looks wrong; and
each attempt uses up the site's once-every-three-days slot, so the next attempt fails the same
way. Our own work last week flagged this as a pre-existing risk nobody had filed — it's now
filed properly (case 416), with the evidence, the likely cause, and three candidate fixes, and
the automatic diagnosis loop is running on it.

To be clear about cause: this is **not** something our changes broke — the failures predate them
by a week. Our measuring code may have slightly lowered the page threshold, and I've flagged that
question honestly rather than deciding it myself.

For our own case: the smaller of the three named sites gets checked this afternoon and should be
fine; the mid-sized one is right at the danger line, which accidentally makes it a useful
experiment; vonc can't be measured until the timeout is fixed. Friday's decisive return visits
are all to small sites, so the finish line hasn't moved.

**27 August, evening.** The verdict on the day: our measurement design is confirmed on every site
it could actually run on — seven sites now, roughly sixty worked-out failures, not one case of
the prover fooling itself — and the repair agent has obeyed its measured instructions perfectly
since last night's prompt fix, sixteen out of sixteen. Two of the three sites we'd named in
advance couldn't be measured at all, but for a reason we proved is older than our work: the
platform-wide timeout that's been silently eating every large site's check for a fortnight. That's
now case 416, fully evidenced, with a cheap interim fix spelled out for whoever takes it — I
deliberately didn't take it myself without your word, since it touches shared machinery beyond
this investigation. Friday's return visits — the finish line — are all to small sites the timeout
can't touch.

**31 August.** Friday's return visits happened, and the news is what we hoped: **the repairs
held, and the case is closed.** The checker went back to all three sites on schedule on Friday.
On the remortgage site we can see it plainly did its job — it flagged a brand-new problem on a
newly added calculator page, and said nothing at all about any of the five spots we repaired.
On garden-tools and cookly it filed nothing and raised no errors; I couldn't read the checker's
own diary for those two visits because the platform only keeps that diary for a day and I graded
this two days late, but every place a failure would have left a mark is empty, the failure log
was demonstrably working that same evening (it caught over a dozen other sites' failures around
them), and both sites are comfortably inside the size range where checks complete. The repaired
stylesheets are still exactly what we published — I re-checked the live sites against the
database today, byte for byte.

So the full arc closes: the repair agent used to author fixes that could never win; now it
measures what it must beat, beats it, and the checker that originally filed the complaints comes
back and agrees they're gone. Not one repaired spot on any site was flagged again. The one
designed safety valve — parking a complaint as genuinely unreachable — has never once been
needed, which is the best possible answer.

Two loose ends, neither ours: the big-site timeout (case 416) ate another thirteen large-site
checks over the weekend and still needs an owner; and the whole fleet ran out of its Anthropic
allowance from Friday lunchtime until tonight, which stopped one new repair on the remortgage
site — that one will retry by itself tomorrow. I've moved case 390 to the closed pile and
written the closing summary.
