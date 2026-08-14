# Where we are — bug 223, the code index's blind spots being read as "it doesn't exist"

Plain prose, append-only, newest at the bottom. This is the running log.

---

## 2026-08-10, morning — picking the bug up, and what it is really about

We have a small robot whose job is to check the landmine notes we keep for each other.
When someone appends a note ("touch this file and here is the trap"), the robot looks up
the files and symbols the note names, reads them, and records a verdict: still true, gone
stale, or needs a human. The verdicts go into the database where the review council and
the next session can read them.

The robot looks things up in a code index — a table of every symbol in the codebase. The
trouble is that the index only contains **Go**. It has 5,837 entries this morning and
every single one is a `.go` file. But most landmine notes are not about Go: they are about
scripts, SQL migrations, database tables, config values, commands. So when the robot looks
those up, it gets nothing back — and it has been writing that nothing down as "this does
not exist", sometimes in so many words: *"the entire described workflow has no footprint."*

The best example of how wrong that is: one of those verdicts declared that three of our
own scripts, and a whole category of database rows, did not exist anywhere. It was
delivered **by those scripts**, and stored **in that category**. It disproved itself on
arrival.

The reason this matters more than one bad verdict is what a verdict is *for*. A "stale"
verdict is the signal a future session uses to delete a note. So a false stale does not
just fail to check something — it actively argues for throwing away a warning that was
correct.

## What we found today that changes the shape of the fix

Two things, both from looking rather than assuming.

**First**, the person who filed the bug thought the problem was "non-Go footprints". It is
wider than that. The index does not merely lack other languages — it also lacks two
categories of *Go* declaration: package-level `var`s and `const`s. There are around 930 of
those in the codebase and none is indexed. So when a note points at one, the robot says
things like *"no longer resolves as a standalone symbol (possibly inlined or renamed)"* —
which is worse than saying nothing, because it invents a plausible story a human will then
waste an afternoon chasing. Nothing was renamed. The index simply cannot hold that kind of
thing. The same gap stops our diagnosis loop cold: it can find every *use* of one of those
declarations and never the declaration itself, so it gives up with "unverifiable" while
naming exactly the thing it cannot see.

**Second**, and this is the useful discovery: the database table was *built* expecting
those two categories. Its own constraint permits them. The half of the code that reads the
index already treats them as ordinary code. Only the half that writes it never learned to
collect them. So this is not a design gap, it is an unfinished job.

## What is actually broken, in one sentence

The sentence the lookup prints when it finds nothing. It currently says, in effect, *"we
ran your query and it matched nothing; this is a real answer, not silence."* That wording
was itself a fix, from an earlier bug where empty answers were read as silence — and it
worked. But now it overshoots: for a thing the index cannot represent, "we ran it and
found nothing" is true and deeply misleading. A guard that is too confident is still a
defect.

And the deeper point, which the bug's own author corrected himself into: the blindness is
perfectly consistent, but the *conclusion drawn from it* varies run to run. Four verdicts
on identical empty input ranged from a careful "cannot be mechanically verified" to a flat
assertion of non-existence. You cannot tell which one you are holding. So asking the robot
more nicely to be careful is not a fix — three times in four it already is careful. The
one flat wrong answer is what does the damage, and only something structural removes it.

## Where we are right now

Bug confirmed still live, re-measured this morning rather than taken on trust. Ownership
checked two ways: no lane owns it, and no live session is working it. The mechanism is
read end to end and the four places that consume the same lookup are identified — this is
shared machinery, so a change here is seen by the review council's own seats, not just by
the landmine robot.

A design pass is running now to rank the candidate fixes. My own preference, going in, is
to fix it in the shared machinery rather than in the one agent: teach the lookup to know
what it cannot see, say so in words that cannot be read as absence, and hand the consumer
a machine-readable fact ("nothing you asked about was checkable") so the decision stops
being a matter of how the model felt that run. Then the separate, larger question of
actually indexing the missing Go declarations gets its own round, because it widens what
every diagnosis run searches and that deserves to be reviewed on its own merits.

Next entry will record what the design pass recommended and what the council said.

## 2026-08-10, afternoon — built, committed, and one uncomfortable moment

The fix is in and the config half is already live. Here is what happened, including the bit
I got wrong.

**The design.** I had a design pass run by a second model before writing anything, and it
agreed with the shape I had gone in with, which was reassuring rather than surprising. It
also caught a number of mine: I had sized the follow-on work using a quick text search that
counted 930 declarations, and the real figure is 1,173, because a grouped declaration block
counts as one line however many things it declares. Worth recording because it is the same
class of error as the bug itself — a search tells you what it can see, not what is there.

**What we actually changed.** Three things, in order of how much they matter.

The smallest and most important: every verdict the robot writes now carries a line, composed
by the machinery rather than by the robot, saying how many of its checks could not be
answered by the index at all. That matters because the verdict is read months later by
people and by other machines, long after the run that produced it is gone. A caveat the robot
is merely *asked* to include arrives most of the time — and "most of the time" is exactly the
problem we are fixing.

The second: the robot can no longer *reach* the "this note is stale" verdict when nothing in
its round was actually confirmed against real code. Not discouraged from it — unable to get
there. The route through the workflow now forks on a plain true/false fact, and the branch
for "we confirmed nothing" offers only two answers, neither of which is "stale". I put this
second because a determined model could still type the word into free text; the line above is
what protects the reader either way.

The third, and the one that helps most widely: the lookup itself now says what it cannot
see. It reads its own contents — which file types, which kinds of declaration — and when you
ask it about something it structurally cannot hold, it says so instead of saying "we ran your
query and found nothing". Crucially, all of that wording is *computed* from what the index
actually contains, so when we widen the index later the warnings retire themselves. Hardcoding
"we only hold Go" would have been quicker and would have quietly become a lie.

That third change benefits four different consumers, not just the landmine robot — including
the diagnosis loop, whose own source comments had already asked for exactly this and never
got it.

**Two things we found that the original bug report did not know.**

The blind spot does not only make things look absent — it can make them look *present*. If a
note points at a folder, and that folder happens to contain some Go files in its
subdirectories, the check comes back with a generous list and reads as confirmation, while
every file the note actually named is invisible. That is worse than a wrong accusation: a
wrong accusation makes you go and look. A flattering half-answer looks like diligence.

And the careful runs are not free either. I fired the robot at a fresh note on purpose this
morning, before changing anything, to bank the "before" picture. It drew the careful branch
— and still spent two model calls and eight queries to arrive at "either these files don't
exist or they aren't indexed, I can't tell", which one census query answers outright. It then
guessed the answer correctly and hung its whole verdict on the guess.

**The uncomfortable moment.** I wrote twelve tests, watched them all pass, and then did what
this project's rules require: broke each guard deliberately to check a test would notice.
Six of seven noticed. The seventh did not — I deleted the folder-listing fix entirely, at the
point where it is used, and every test stayed green, because they all tested the *function*
and none tested that anything *calls* it. And that was the guard for the false-confirmation
problem I had discovered myself an hour earlier. Finding a failure mode does not protect you
from it. Three more tests now go through the real path and the deliberate break is caught.
It is written up in the shared log of wrong calls, because the useful part is the question:
which test fails if I delete the *call*, not the function?

**Where we are.** Committed. The database half is applied and I proved it is safe to have
applied it early rather than assuming: I fired the robot again afterwards, on the unchanged
program, and watched the new fork evaluate to "no" and take the old route, complete
normally, and produce the same verdict as before. It could have failed three ways and did
not.

The code itself does nothing until the next build of the service goes out — that is normal
here, and I have left a marker so anyone can check in one command whether the running program
contains this change or not. It is currently absent, which is the useful thing about having
banked that check before making it.

**What is still owed.** The review council's verdict, which is running now and which I will
read and act on. Then, after the next build: re-run the robot on the two notes that started
this and compare against the banked "before" — including that the checks which *did* work
must still work, because a fix that buys quiet by checking less would look like success.

And then the larger piece, deliberately kept separate: actually teaching the index about the
~1,170 declarations it currently cannot see. That widens what every diagnosis in the system
searches, so it deserves its own review rather than riding in on this one's coat-tails.

## 2026-08-10, after the build went out — it works, and one honest deduction

The new build is running and the change is in it. I checked that at the program itself
rather than trusting the version number: the phrase the fix adds was absent before and is
present now, on both copies of the service, alongside a phrase I never added (still absent,
proving the check isn't just matching everything) and one from an older fix (present,
proving the check works at all).

Then I re-ran the robot on the note that started all this — the one it had declared
non-existent, in a message delivered by the very scripts it said did not exist. It now says:

> The entire footprint (Python and shell scripts) falls outside the Go-only index; existence
> and behaviour of these three scripts could not be checked.

That is the whole point. Same index, same absence of evidence, and it now reports the
absence of *evidence* instead of the absence of the *thing*. Four notes re-checked, and all
four stored verdicts carry the machine-written line saying how much of the round was
actually checkable. Importantly, the fix did not buy that quiet by checking less: on a note
with a mixed footprint, the parts that live in Go were still confirmed by name in the same
verdict that flagged the rest as unverifiable.

**The honest deduction, which I would rather write down than leave you to find.** One of the
two protections I built has never actually fired, and now I know why. The idea was that if a
round confirms nothing at all, the robot cannot even be offered the word "stale". But
"confirms nothing" turns out to be almost impossible to reach, because the text search
matches on fragments: a check looking for a Python constant called `VECTORS` came back with
eight results, all of them unrelated Go code that merely contains the letters "vector". One
accidental match like that is enough to count as "we confirmed something", so the guard
stands down.

Two things follow. The first is that the problem I found this morning — a search answering
a question about one file with a confident result from a completely different one — is not a
curiosity, it is the normal case. The new caveat caught it, and the robot's own verdict
repeated it back correctly: those matches "are unrelated to the VECTORS constant described
in the entry". The second is that the protection actually doing the work is the plainer one:
every answer now says what it could not see, and every stored verdict carries that summary.
I ranked those two that way when I built them, and the evidence agrees — but the guard that
has not fired should not get the credit, so I have said so in the bug file, in the register,
and here.

There is a clean follow-up, and it belongs to whoever owns the checking robot rather than to
this piece of work: it should not ask a text-search question about a file it can already see
is not Go. That would remove the wasted checks and let the standing-down guard actually stand
up when a note really is unverifiable.

**Where that leaves us.** The reporting half is done, live, and proven on the case that
motivated it. The remaining half — teaching the index about the ~1,170 declarations it still
cannot see — is written up as the next task, with its risks measured rather than guessed. A
fresh session can pick it up from the handoff without re-reading any of this.

---

## 2026-08-11 — the second half landed, and the robot changed its mind about the same file

The remaining half is done. The index can now see plain declarations — the named values a
programmer writes once at the top of a file and refers to everywhere else — and the checking
robot has been asked the same question it got wrong three days ago.

The clearest way to show it is the two answers side by side, same entry, same file, nothing
changed but what the index can see:

> **Friday:** the two pattern lists "no longer resolve as standalone symbols, **possibly
> inlined or renamed**" — so the note was flagged for a human to look at.
>
> **Today:** all six named things "**confirmed present at expected line ranges**" — note
> still valid, nothing for a human to do.

Nothing was ever renamed. On Friday the index simply had no way to hold that kind of thing,
so the honest answer "I cannot see this" came out as a guess about the code having changed.
That guess is what we set out to kill, and it is gone.

**Everything else was checked rather than assumed.** Every category of thing in the index was
counted before and after: the new ones appeared, and not one of the old ones moved. That
matters more than it sounds — the way the index stores things, a new entry can quietly
overwrite an old one of the same name, and the count is the only thing that would ever show
it. I also checked directly for that overlap and there was none, so nothing was lost.

**Two things worth telling you about, because both are the kind of mistake that looks like a
success.**

The first is mine, or rather this workstream's. The handoff said we should expect about 1,371
new entries. We got 1,204. That gap is not a fault — the 1,371 was wrong. It was measured by
running the real tool, which was itself the fix for two earlier bad guesses, but it was run
without the one setting the live system passes it (the live one skips the documentation
folder). So it was a proxy for the third time, and this time it had been written down as the
pass mark for the deploy. If I had trusted it, I would have read a perfectly healthy result
as 12% of the data going missing, and that particular symptom is the one we are told to stop
and investigate. I caught it by deciding "about 1,371" was too vague to be a real test, and
working out the exact number the live system should produce before looking at what it did
produce. Every category then matched to the last unit.

The second is a small piece of good news that arrived from another workstream. The handoff
warned that this change would be impossible to confirm in the running system, because it
adds no new text to search the program for. That was true when it was written, but somebody
else has since made every service announce which version of the source it was built from
when it starts up. So instead of hunting for a fingerprint, you just ask it, and the answer
is exact. I have written that up so nobody repeats the old workaround.

**One decision I deliberately did not take on your behalf.** Our working copy is 228 commits
ahead of the shared copy the indexer actually reads. I did not push, for two reasons: it is
228 commits of many people's work and not mine to send, and — more usefully — not pushing is
what made today's check trustworthy. Because the shared copy had not moved, re-running the
indexer was a fair comparison: anything that changed was down to the new code and nothing
else. Had I pushed first, the code would have changed at the same time and the before/after
count would have told us nothing. Whether to push now is a separate question, and yours.

**Where that leaves us.** The bug is finished: both halves built, reviewed, live and proven
on the cases that motivated them. What is still open is a governance question for you, not a
technical one — written up as RFC 022 — about whether the careful way we are told to extend
shared machinery should also be the thing our reviewer flags as needing extra scrutiny. At
the moment following the rule and breaking it draw the same warning, which makes the warning
worth less. There are three options costed in that document and a recommendation, but it is a
judgement about how you want the estate governed and I have not taken it.

---

**2026-08-14 — the staleness bug is closed, and the counter you asked for is built. Two
small decisions are yours when you have a minute.**

**The staleness problem is fixed and running in production.** The short version of what was
wrong: when someone asked our code index about a symbol it had never seen — usually because
the code was written after the last time anyone pushed — the answer "nothing found" came
with no explanation, and the reviewing model made one up ("that kind isn't indexed"),
confidently and wrongly. The subtle part, which the diagnosis loop caught me on: we already
HAD a warning about this, printed at the top of every report. The model read it and ignored
it, then quoted the vocabulary printed right next to the empty answer instead. So the fix
was placement, not another warning: every empty answer now says, in the same breath, "this
answer describes the code as of commit so-and-so, from such-and-such a date — anything newer
cannot appear here, so if what you asked about is newer, this is the index being behind, not
the code being absent." And every saved verdict now records which version of the code it was
judged against, so someone reading it months later can date it. The reviewers approved it
unanimously, it shipped in the current build, and we watched a real verification run write
the dated verdict correctly. The bug file has moved to the closed pile.

**The counter from RFC 022 now exists.** When you ruled on that RFC you accepted a known
blind spot: individually harmless opt-in fields no longer trigger a review, so nothing would
notice when a shared piece of machinery quietly accumulated its tenth one. The mechanical
counter that was promised is now built and run against the live fleet. What it found: 118
actions declare optional fields, 21 of those are used by two or more agents, and the widest
shared surfaces are the repo analyser (12 optional fields), the note-writing action that
started all this (11 fields, used by eight different agents), and the fix-commit preparer
(11). The reviewer prompts that used to say "that counter is not built yet" now point at it.

**The two decisions that are yours:** (1) **the budget** — at what count should a shared
action's optional fields trigger an architecture review? For scale: a budget of 10 would
flag exactly the three actions above today; 12 would flag none. (2) **whether the counter
runs on a clock** — the equivalent check from RFC 006 got a daily scheduled run because
nothing else re-checks live config; this one is currently run-on-demand only. Neither
decision blocks anything; the RFC stays open until the budget is ruled.

---

**2026-08-14 (later) — you ruled the budget: 10. Done, and one correction of yours is now
written into the reviewers' own instructions.**

You said the founding idea was that every agent should be somewhat independent so it can
be reused in other workflows, and that "a shared action nobody understands" was the wrong
way to talk about it. That correction is now part of the machinery, not just the chat: the
reviewer prompts say in so many words that sharing is estate design, that what gets
reviewed when the budget trips is the *accumulated pile of optional switches*, never the
reuse itself. The check now runs with 10 as its default, so anyone running it plainly gets
your ruling; RFC 022 is closed.

What the ruling means in practice: the three widest actions (the repo analyser at 12
switches, the note-writer at 11, the fix-commit preparer at 11) are over the line today,
so each owes one review of its accumulated switches — a look at the whole pile, then the
acknowledged level becomes that action's baseline. Nobody has scheduled those three
reviews yet; they are ordinary review rounds, not emergencies.

Still yours, no hurry: whether the counter runs on a daily clock. I've explained the
trade-off in chat — the short version is that the old scheduled check had to carry a
second copy of its logic in Python (with tests to stop the copies drifting) because the
scheduler's containers can't compile Go, but there is now a newer pattern in the estate
where the check is compiled into its own small Go image and scheduled like any other
service, no second copy at all. If you want it on a clock, that is the shape I'd build.
