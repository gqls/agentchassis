# SUMMARY — 2026-08-12 — 040-kafka-dial

Third summary, seventeen days after the last one, written because the week of
data finally came in and it changed the picture: the case does not close, and a
second, much bigger fault was found that the first snapshot never saw. Current
state only; chronology is in `NOTES` and `README_where_we_are`.

---

## What we're trying to do

Kafka connections across the cluster fail intermittently. Originally filed as
plain network timeouts; the last summary left it as "instrumented, let it run a
week, see what the counter says."

## Where we've come from

Nobody had come back to it. The workstream directory sat untouched for
seventeen days — past its own one-week checkpoint — while the metric it built
quietly kept counting in the background. Picked back up as part of a routine
sweep through the open-bug backlog (the ownership checker's first answer
pointed at the *other* bug that happens to share the number 040, a known false
alarm noted at the top of the bug file).

## What we've done

Re-ran the week-old close test against live data. It fails outright: the rare
stall the last summary caught once is still happening (32 times in the latest
week, not zero). That alone means the case cannot close and has to fall to
"diagnose the residual" instead.

Diagnosing the residual is where the real finding is. Breaking the week down by
failure type turned up something the 22-hour snapshot never saw: **connection
refused, 71,832 times, in a single day and a half, across dozens of temporary
worker pods** — two orders of magnitude bigger than every other kind of failure
this metric has ever recorded, combined. (Caught myself nearly reporting a
number 340 times too small on the way there — used the wrong kind of average on
the first pass, re-checked before writing anything down.)

Ruled out the obvious explanation first: the Kafka servers themselves did not
restart or wobble during that day — they have been sitting untouched for
forty-five days. So it isn't the servers going away.

Reading the failure labels closely showed something more specific: almost all
of these failures are missing the "which server" label the metric is supposed
to always carry, and that turns out to mean the program tried to connect to an
address with no server name in it at all — just a bare port number, which on
this kind of connection quietly means "talk to yourself." Nothing listens
there, so it fails instantly, which is also why one pod can rack up over a
thousand of these in its lifetime — there's no ten-second wait dragging it out
like the rare stall has.

Found the exact piece of code that can produce that: when a worker asks the
Kafka cluster "who's in charge right now" (needed when creating a topic), it
builds the reply address from whatever the cluster says the answer is, with no
check that the answer isn't blank. Someone had already half-noticed this — a
different piece of code elsewhere in the same file already filters out that
exact blank address, without anyone connecting it to this bug.

Rather than guess and patch, filed it to the automated diagnosis loop first —
the loop exists for exactly this shape of question (fleet-wide, cause not
obviously where the symptom is). **It came back unable to settle it either
way** — not because the finding is wrong, but because the loop's tools can
search application logs and read our own code, and this fault lives entirely
in the metrics system and possibly inside a third-party library's internals,
neither of which the loop can reach. It did independently confirm, from real
log entries, that the related "who's in charge" lookup genuinely does produce
the *other*, already-known kind of failure (the slow ten-second one) fleet-wide
— which is at least evidence the loop was looking in the right neighbourhood.

Checked the application logs myself as a follow-up: not one line, anywhere,
about any of this. The failure is completely invisible except through the
counter. That's consistent with the rest of this bug's history — this system's
metrics have caught things nothing else ever could.

Shipped one small, safe fix regardless of which exact code path turns out to be
responsible: the "who's in charge" lookup now refuses to build a reply address
from a blank answer, and logs a warning when it happens — closing the silent
half of this finding either way, and giving whoever hits this next an actual
log line to look at instead of a mystery counter.

## Where we are now

**The case is open, bigger than it looked a week ago, and honestly reported as
such.** One narrow fix shipped; it is not claimed to be the whole answer, and I
have said so plainly in the bug file rather than let a partial fix read as a
closed case.

What's confirmed: the rare stall is real and ongoing (32/week). The refused-
connection burst is real, measured, time-bounded, and not caused by the
Kafka servers restarting. What's not confirmed: whether the code fixed today
is the actual site of the burst, or whether a second, harder-to-reach cause
(inside a third-party networking library, not this codebase) is doing the same
thing. The burst has not recurred in the twenty hours since it stopped, so
there is nothing live to test the fix against yet.

## Where we're going

1. **Watch for a recurrence.** If it comes back, the new warning log will say
   directly whether this fix's guard caught it — which settles, on the next
   occurrence, the question the diagnosis loop couldn't settle from history.
2. **The thirteen other services still have no visibility at all** — unchanged
   from the last summary, still the obvious next extension, still deliberately
   not bundled into this session's narrow fix.
3. Whoever next re-reads the week's data should re-run the close test again —
   it has now been wrong to close twice in a row, and there is no reason to
   expect the third read to be the last one either.
