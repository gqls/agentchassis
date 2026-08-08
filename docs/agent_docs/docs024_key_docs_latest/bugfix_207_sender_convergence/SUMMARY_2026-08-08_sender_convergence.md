# SUMMARY 2026-08-08 — sender convergence: the series is complete

**What we're trying to do.** When a piece of work fails somewhere in the platform, the
message that reports the failure carries a verdict: "worth retrying" or "give up". For
a long time three different places in the code wrote that verdict, and two of them
guessed — one only trusted a flag nothing ever sets, and one simply always said "give
up". Because the first answer to arrive is the one that counts, a wrong "give up" from
a fast sender threw away work that a retry would have saved. The goal of this lane was
to make every sender decide the same way, through one shared classifier.

**Where we've come from.** Bug 207 (fixed, live 7 August) converged the first two
senders onto the shared classifier. That immediately exposed bug 216: the machinery
that acts on a "worth retrying" verdict was quietly broken — it did the paperwork of a
retry without ever sending one. 216 was fixed and proven live on 8 August. The same
investigation caught the third sender in the act: the coordinator's own
failure-notifier, which answers first on the most common real path, still said "give
up" unconditionally — that became bug 217.

**What we've done.** 217 is fixed. The shared classifier had to move to a package both
layers can use (the code's import structure forbade the obvious call), with the old
location keeping working aliases so nothing else changed. The notifier now classifies
the failure before stamping the verdict. The council approved the plan first round;
its two advisory objections were answered with extra evidence rather than waved off.
We measured before shipping: about half of these failure notices — six thousand in a
fortnight — were transient things like timeouts that deserved a retry and were being
thrown away.

**Where we are now.** Proven live in production tonight, on the new build. We staged a
real failure end to end: the notifier that yesterday said "give up" now says "worth
retrying", the paperwork agrees with the wire, and the retry actually went out — we
read it back off the queue. The safety cap (three retries, then stop) held exactly.

**Where we're going.** Nothing further is planned for this lane. Two watches remain:
the weekly failure-rate check (a retry that fails at one level can now trigger a retry
one level up, which is bounded but worth watching), and the classifier's vocabulary,
which should only ever change through its pinned tests. The lane's documents record
how to run both.
