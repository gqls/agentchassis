# SUMMARY — 2026-08-02b — from "a site was wrongly called clean" to a check that runs itself

*(Second summary of the day. The first — `SUMMARY_2026-08-02_one_promoter_one_owner.md` —
ended with one question outstanding: the new check existed but nothing ran it. That is
now answered, which is why this is a new file rather than an edit of that one.)*

## What we're trying to do

A site accumulates problems — a broken link, a weak call to action, a page that needs
re-rendering. Something has to notice them, queue them, and finish by rebuilding the
pages so the fixes actually reach the public site.

The step that does the queueing is greedy on purpose: it takes **everything** waiting on
a site, not just the things the current run found. That is sensible when one thing calls
it. Three things called it.

## Where we've come from

The improvement loop calls two helper agents and then does its own sweep. All three ran
the same greedy step, so the helpers cleared the pile first and the loop's own sweep
found nothing left. It reported "I queued nothing" — perfectly true — and concluded
there was nothing wrong with the site. It then announced *"No issues found — site is
clean"* and skipped its closing rebuild, on a site where sixty-seven problems had just
been queued.

Nobody had noticed for months, because a separate scheduled job picks the queue up
anyway every two minutes. The fixes happened. Only the honest ending went missing.

**Yesterday we fixed the symptom.** The shared step now also reports the state of the
*site* — is there work waiting here, whoever queued it — and the loop reads that
instead. Proven live: same site, same command, one day apart, the decision going the
other way.

**The review council said that did not close the underlying problem**, from two
independent directions, and they were right. It left the next agent someone adds
inheriting two similar-looking signals and a rule for choosing between them buried in a
code comment. That became a written proposal with three options.

The proposal admitted it was missing one fact, and that the missing fact decided the
cheapest option. When you asked to be walked through it, we checked that fact first:
does anything other than the improvement loop call those two helper agents? Nothing
does — not "probably nothing", but a scan of every live agent returning exactly two
results, both the improvement loop. The audit that made the structural option look
expensive was done, and it found nothing to audit.

## What we've done

**You chose the structural option and it shipped the same day.** The helper agents no
longer do the queueing. Only the improvement loop does.

**But deleting the duplicates is a one-off that the next person undoes without knowing.**
So the durable half is different in kind: an action can now *declare* that it is meant to
have exactly one owner, and a check reads the whole fleet and reports any action that
has quietly picked up a second one.

**And then you asked for that check to run automatically**, which it now does — daily,
against the live system, recording what it found each time.

The obvious place for such a check is "when somebody commits code", and that turns out
to be the wrong answer. When somebody writes a database change, it has not been applied
at the moment they commit it — so a commit-time check would look at the live system, see
everything in order, and wave through the exact change that causes the problem. A good
deal of configuration here is also changed directly in the database with no commit at
all, which a commit-time check cannot see even in principle.

The council approved the work, with two seats raising advisory points and none blocking.
Four seats independently asked for the same check to be done before deleting the step; it
was done and came back in our favour. The sharpest point in the round was the one that
led to the scheduled job.

## Where we are now

Finished, live, reviewed, and proven — with the proving done in both directions rather
than by watching things pass. A real run checked 179 live agents and found nothing. Then
we deliberately fed the check a question it ought to fail, without touching any real
configuration, and confirmed it fails loudly and records the failure. That deliberate
false alarm was removed afterwards so nobody later reads it as real.

The check writes a note every time it runs, including when it finds nothing. If it only
wrote on problems, silence would mean either "all clear" or "this has been broken for a
month", and those must not look alike.

**Two things are worth saying plainly rather than burying.**

The first is a mistake. When the duplicate step was deleted, the normal path into it was
redirected and the *error* path was not — leaving a pointer to something that no longer
existed, which would have stranded any run unlucky enough to fail at that point. The
migration had a check for exactly this; it ran, and printed exactly the right warning;
and the change committed anyway, because the check was phrased as a question rather than
as a stop. Fixed within minutes, with a version that genuinely halts, and proved to halt
by deliberately re-breaking things inside a transaction that was thrown away. The same
question asked across the whole fleet found no other instance.

The second is a known weak spot, not a defect. The scheduled job runs a small copy of the
check rather than the original, because running the original inside the job would mean
downloading and compiling the entire codebase every night — and a check that breaks for
plumbing reasons is one people learn to ignore, which would put us back where we started.
A copy can drift from the original, so two tests compare them and fail if they ever
disagree, and those tests were themselves verified to catch a difference rather than
assumed to.

## Where we're going

Nothing here is outstanding. Two things sit nearby and are deliberately not part of this:

There is a **second, unrelated route** to the same false "clean" message — a site that
has been audited three times is skipped entirely and told it is fine. It is filed as its
own case. No site is currently in that state, so it is a trap rather than a live fault,
and it becomes live on the day the improvement loop is switched back on.

And **the improvement loop is still switched off**, so it only runs when someone fires it
by hand. Everything above is worth more the day that changes — which is also the day the
second route starts biting, which is why it was filed rather than mentioned in passing.
