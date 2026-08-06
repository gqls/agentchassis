# SUMMARY — bugfix 196, 2026-08-06 (milestone: closed)

**What we're trying to do.** When one of our automated agents fails at a task, the
system that asked it to do the task should be told "that failed, and here's why" —
and should then take its failure route: retry, fall back, or stop cleanly. Bug 196
was the discovery that this wasn't happening: a failed agent answered its parent
with a message stamped "complete, success", with the real failure buried inside the
message body where nothing looked. The parent recorded the failure as if it were
the step's result and carried on building with garbage. Nothing errored, so nothing
was ever investigated.

**Where we've come from.** The bug was filed on 4 August by the team closing bug
195 (a related mis-classification bug), at the insistence of a council reviewer who
refused to let the finding live only in a side note. It was filed honestly as
"read from the code, not yet proven live". We picked it up on 5 August after
checking no other session was working it. Investigation confirmed the code read and
sharpened it twice: there is a second sender with the same defect, and the system
actually sends a *correct* failure message moments after the wrong one — but the
wrong one always arrives first, claims the parent's attention, and the correct one
is thrown away as a duplicate. We also established this was a regression: an older
version of the code did this properly, and a refactor lost it. Notably, the fix
routes failures onto error-handling machinery that already exists and is already
exercised daily by other parts of the fleet — we weren't building anything new,
just stopping the lie at the sender.

**What we've done.** Measured the blast radius before touching anything: no live
workflow depends on the broken behaviour, and every piece of code that reads the
old message format was found and deliberately left working. Wrote the fix plan
preferring the structural remedy — one message-sender through which every response
travels, where a failure now says it is a failure — and put it through the review
council, which approved it first round with four advisory points, each of which we
answered with evidence. The implementation was done by a delegated agent, verified
line by line, and proven the hard way: the new tests were run against the un-fixed
code and fail at exactly the defect, so they genuinely guard it. We then built a
live experiment: park a parent workflow waiting on a child, make the child fail,
and capture the actual message sent back. On the old software the capture shows
"complete, success: true" wrapping a failure. On the new build — rolled overnight —
the identical experiment shows "error, unrecoverable, success: false" with a proper
error record, and the old body format untouched for its existing readers. Getting
that experiment right took three attempts, and every wrong turn is written down
where the next person will trip on it: two entries in the wrong-calls log and one
new landmine (a subtle platform trap the probe uncovered — messages sent between
agents are wrapped in an envelope that hides the fields a child would need to pick
its own workflow).

**Where we are now.** Closed, in full. The fix is committed, council-approved,
live on both running replicas (verified against the running binaries, not the
tag), and proven by the same experiment run before and after. The bug file has
moved to the closed pile with the before/after table in its header. The two teams
whose ground this touches — the hung-spawns investigation and the work-queue lane
— have dated notices in their files saying exactly what changed about their
guarantees. The new sender seam is registered (CTS-058) so future sessions can
find it, and the transferable lesson — the first answer to arrive claims the
question, so a wrong answer produced earlier permanently beats a right answer
produced later — is in the debugging guide. All probe scaffolding is cleaned up.

**Where we're going.** Nothing further is owed on 196 itself. Two follow-ons
belong to other work: bug 197 (the sibling classifier that still judges
retryability by matching words in error text — deliberately untouched here, and
now cheaper to fix because 195's recorder is accumulating the real error
population) and a watch item — now that failures are honest, workflows that only
ever "succeeded" because they ignored child failures will start failing visibly.
Those failures are real and were always real; the record just stopped lying about
them. If a wave of them appears, that is the fix working, and the error messages
now carried in the envelope say exactly what broke.
