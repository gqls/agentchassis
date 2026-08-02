# Where we are — the improvement loop saying "site is clean" when it isn't

*(Plain-prose log, append-only, newest at the bottom. Owner's document.)*

---

**2026-07-31, evening — session "bugfix 27"**

We have a loop that is supposed to sweep a site, find what is wrong with it, queue the
fixes, and finish by re-rendering the pages so the fixes actually appear. It has been
ending with the message **"No issues found — site is clean"** on sites where it had just
queued dozens of problems.

The reason turns out to be a small thing about who does the work. The step that moves
findings from "spotted" to "queued" exists in three different agents, and it is greedy: it
moves *everything* for the site, not just its own findings. The main loop calls two helper
agents first, and only then runs its own copy of that step. By that point there is nothing
left to move, so its copy quite honestly says "I moved nothing" — and the loop reads that
one sentence as "the site is fine". It then skips the closing re-render and reports success.

Nobody noticed for months because there is a separate scheduled job that picks the queued
fixes up anyway, every two minutes. So the fixes did happen. What was lost each time was
the final re-render pass — the thing that guarantees the fixed content actually reaches the
live pages — and, less visibly, the truth: anything reading the loop's result was told a
site with open problems was clean.

I ran the loop once, deliberately, at vetcomparison.uk, on the current code, to watch it
happen rather than take the previous session's word for it. It did exactly what was
predicted: the first helper queued 24 findings, the second queued 3 more, the main loop's
own copy found nothing left to queue, and the run ended on "site is clean". Twenty-seven
findings queued in that run; zero closing re-renders created. That matters beyond this bug
— the original report had honestly marked "this happens every time" as an *assumption*,
because the database had already deleted the history that would prove it. Now there are two
independent sightings, on two sites, two days apart.

**The fix.** Rather than argue about which agent should be allowed to do the queueing, I
made the step able to answer a different and better question: not *"did I personally queue
anything just now?"* but *"does this site have work waiting?"* — which is true regardless of
who queued it, in what order, or which agent gets that step added next. The loop will read
that instead. The old signal is left exactly as it was, because three other places in the
system use it correctly to mean "my own results were empty", and changing its meaning to fix
one of them would quietly break the other three.

**What is not done yet.** The code half is committed but does not do anything until a new
chassis image is built and rolled out, and you asked me not to roll one (a roll interrupts
other sessions' work). The one-line configuration change that switches the loop over is
written and **deliberately not applied** — applied too early, against the old code, it would
make *every* run report "clean", which is worse than the bug. So the ticket stays open. What
is owed is two steps, both written down: roll an image containing the commit, then apply the
config and re-run the same sweep to watch the loop take the other branch.

I also found a second way the same false "clean" message can appear: if a site has already
been audited three times, the loop skips straight to the end and says "clean" without
looking at all. No site is currently in that state, so it is a trap rather than a live
problem — recorded, not fixed, because it needs a different answer (an honest "we skipped
this" message) rather than the change I made here.

The change has gone to the review council; the verdict lands in about half an hour and I
will act on it whatever it says.

**Later the same evening**

A new chassis image went out while I was writing this up — v1.0.1218 became v1.0.1223 — and
it does **not** contain the fix. I checked the running containers rather than the version
number: the string my change adds is absent from both, while a control string that has been
there for months is present, so the check itself is sound. The image was simply built from a
snapshot of the code taken before my commit. Nothing is wrong; it just means the two steps
owed are still owed, and the next person should not read "there was a roll" as "the fix is
live". The command that answers it properly is written into the ticket.

**End of the evening — the review came back approved**

The council approved it, with eight advisory comments and nothing blocking. Four of those
comments were questions I could answer with a database query rather than an argument, so I
ran them, and all four came back in the change's favour. One of them was genuinely
important: a reviewer asked whether the three agents that share the promoting step all file
their work under the same label — because if they didn't, my new count would quietly miss
exactly the items it exists to find. They do. But I had taken that on trust from a comment in
the code instead of checking it, and the reviewer was right to make me.

Two things the reviewers asked for that I had only written down in prose, they wanted as
proper tickets, and they were right about that too. So there is now a separate ticket for the
second way this same false "clean" message can appear (the audited-three-times shortcut), and
a written proposal for the underlying question — should a step that sweeps up everything on a
site have exactly one owner, rather than three agents racing for it? That one is a decision
for you rather than for me; the proposal lays out three options with what each costs.

Where it stands: the code is committed and approved, the switch-over is written and held
back, and the ticket stays open until someone rolls an image and applies it. Both of those
steps are written into the ticket in order, with the exact commands. I checked the committed
code builds and its tests pass in isolation, separately from whatever else is half-finished
in the shared working directory tonight.

**2026-08-02 — you asked to be walked through the proposal, so I ran the one measurement it
was missing first**

The proposal I left you had a hole in it that I flagged at the time: it said option (a) — give
the sweeping step exactly one owner — depended on a fact nobody had checked, namely whether
anything *other* than the improvement loop ever calls the two child agents. I said whoever
took the proposal should check that before deciding. Since the person taking it is you, I
checked it.

The answer is that nothing else calls them. Not "probably nothing" — a scan of every live
agent's definition for anything naming those two children returns exactly two results, and
both of them are the improvement loop. There is no second caller to break. On top of that,
the run counters show the three agents have run three times each, the same three times, which
is what you would see if the children have only ever run as part of a loop run and never on
their own. And neither child does anything with the result of the shared step — they run it
and then stop — so taking it away from them changes nothing about how they behave.

That changes the shape of the decision. The reason we did not do the structural fix in the
first place was that auditing every other caller looked like an open-ended job. The audit is
now done, and it found nothing to audit. What is left is deleting one step from each of two
agents, in the database, reversible, with no code change at all.

One honest caveat, because it is the only soft edge: the run counters only go back to
2026-07-26, and there have only been three loop runs in the fleet's whole recorded history —
because the sweep is switched off and every run so far was one of us firing it by hand. So
"nothing else calls them" is solid (that comes from reading the definitions, not the history),
but "nothing has ever called them" leans on a short window. It does not change my
recommendation; it is the sort of thing that should be said out loud rather than buried.

**2026-08-02, later — you chose "one owner", so that is what we did**

The change itself was small and the interesting parts were around it.

What went out: the two child agents no longer do the sweeping-up. Only the
improvement loop does. That is a database change, applied by hand, with a snapshot
of both agents taken first so it can be undone.

But deleting the duplicate steps once does not stop the problem coming back — the
next agent someone adds could easily be given the same step again, and nobody would
notice until a site was wrongly called clean. So the more important half is a piece
of code: an action can now *declare* that it is meant to have exactly one owner, and
there is a check that reads the whole fleet and reports any action that has picked
up a second one. I ran it before the change and it reported the problem, naming all
three agents. I ran it after and it reported nothing. Same command, same fleet — that
pair is the proof, rather than my say-so.

Then I fired a real sweep, because config that looks right and a run that works are
different claims. Both child agents ran to completion with their step removed, and
the loop's own count went from **zero — which it had been in every run we have ever
observed — to twelve**. That number moving off zero is the whole point: the loop is
now doing the promoting itself instead of finding the cupboard bare and concluding
there was nothing to do.

**And I made a mistake worth telling you about, because my own safety check caught
it and I let it through anyway.**

When I deleted the step, I remembered to redirect the normal path that led into it
and forgot the *error* path — what the agent does if the preceding step fails. That
left a pointer to a step that no longer existed, which would have stranded any run
unlucky enough to hit an error there. My migration had a check for exactly this, it
ran, and it printed exactly the right warning. The transaction then committed
regardless, because the check was written as a question rather than as a stop:
the database prints the answer and carries on. I had written "if this returns
anything, stop" in a comment — which is an instruction to a human, not a mechanism.

Fixed within minutes, and the fix's version of the check genuinely halts. I then
proved it halts by deliberately re-breaking the thing inside a transaction I threw
away, rather than assuming. And I widened the same question to the whole fleet: no
other agent anywhere has a pointer to a step that does not exist, so this was the
only instance — but nothing had ever asked that question before, which is why it is
now written down as a standing check.

The error path is the nasty half of this class: it might not fire for months, so a
successful test run tells you nothing about it.

**2026-08-02, later — you asked for the check to run automatically, so it does**

It now runs itself once a day, against the live system, and records what it found each
time.

The reviewer's point that prompted this was the best one of the round: the check
existed, worked, and reported problems — but nothing ever ran it. I confirmed that was
true of every other audit script here too, not just mine.

One design choice worth explaining, because the obvious answer is wrong. The natural
place to put a check is "when someone commits code". That would not work here: when
somebody writes a database change, the change has not been applied yet at the moment
they commit it, so the check would look at the live system, see everything in order,
and wave it through — passing the exact change that causes the problem. And a good deal
of configuration here gets changed directly in the database without any commit at all,
which a commit-time check cannot see even in principle. So it runs on a clock against
the live system instead.

It writes a note every time it runs, including when it finds nothing. That is
deliberate: if it only wrote when something was wrong, silence would mean either "all
clear" or "the job is broken and has not run for a month", and those must not look
alike.

I proved it works in both directions rather than just watching it pass: a real run
checked 179 live agents and found nothing, and then I deliberately fed it a question it
should fail — without touching any real configuration — and confirmed it fails loudly
and records the failure. I removed that deliberate false alarm afterwards so nobody
reads it later as a real one.

The honest weak spot: the scheduled job runs a small Python copy of the check rather
than the original, because running the real one inside the job would mean downloading
and compiling the whole codebase every night — and a check that breaks for plumbing
reasons is one people learn to ignore, which puts us back where we started. A copy can
drift from the original, so there are now two tests that compare them and fail if they
ever disagree, and I verified those tests genuinely catch a difference rather than
assuming they would.
