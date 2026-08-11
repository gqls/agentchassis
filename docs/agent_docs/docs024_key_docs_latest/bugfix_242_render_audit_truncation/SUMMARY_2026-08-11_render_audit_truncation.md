# SUMMARY — bugfix 242, 2026-08-11 (milestone: fix live and behaviourally proven)

**What we're trying to do.** The platform photographs and measures every live page of
every site weekly with a robot browser. A limit caps each sweep at 25 pages, and when the
limit bit, the saved report looked exactly like a complete clean sweep — the two biggest
sites were silently under-measured on every run, the same skipped pages every week.
We're making a cut-short sweep impossible to mistake for a complete one.

**Where we've come from.** The bug was filed 2026-08-10 with the mechanism unverified.
The action computed the truncation correctly and returned it — but any step that sends a
request and awaits the answer has its own notes discarded by the platform's park
machinery. That mechanism turned out to be already established (RFC_012) and already
ruled on by the owner: durable facts either travel inside the adapter's reply or are
written to the permanent error log before dispatch; the coordinator itself stays
untouched pending a gated design question.

**What we've done.** Implemented exactly that: the request now tells the adapter "26
pages exist, you're getting 5", the adapter repeats it inside its answer's summary, the
ticket-filing step stamps it into its own durable result, and a permanent
`RENDER_AUDIT_TRUNCATED` log row lands before the request is even sent (order
mutation-tested). The weekly cap was raised 25 → 60 by a guarded, rolled-back-able
migration so no current site is actually cut short. The council reviewed it: round 1
REVISE found two real defects (we'd bypassed the estate's sanctioned log-writing door,
and the migration lacked guards against a known twin-row landmine) — both adopted; round
2 APPROVED.

**Where we are now.** Live on `v1.0.1288` and behaviourally proven: we deliberately
dropped the cap to 5, audited the 26-page loancalculator site, and all three records held
— summary `5 of 26, truncated`, ticket step stamped, log row with correct provenance —
then restored the cap. The bug file is marked done in substance and stays in `bugs_open/`
per the owner's filing rule.

**Where we're going.** Nothing further for this lane. The general class — 40 awaiting
actions can still lose their in-flight notes — remains RFC_012's open, owner-gated
question; this fix is one worked example of the sanctioned pattern for it, and the
register (VIZ-012/013/015) points future readers at both.
