# Where we are — bug 124, diagnoses running twice

Append-only. Newest at the bottom. Plain prose.

---

**2026-07-28, first look.**

Every bug we send into the diagnosis loop has been diagnosed twice, and we have
been paying for both.

The way you file a bug for the loop is to run a script called `090_TRIGGER`. That
script does two things: it writes a row into the work queue saying "this needs
diagnosing", and then it sends the job off to be diagnosed. That was the right
design when it was written, because at the time nothing else was watching the
queue.

Since then we switched on the thing that watches the queue. It checks every
minute for anything waiting to be diagnosed, and it picks it up. So now the
script sends the job, and about a minute later the watcher finds the same row
still sitting there and sends the job *again*. Two diagnoses run, on the same
bug, at the same time, neither aware of the other. Each one is roughly a
twelve-to-fourteen minute run of a large model — the most expensive single thing
in that pipeline — and the longest we measured was thirty-one minutes.

Nobody did anything wrong exactly. The script even says, in its own notes at the
top, "the watcher ships switched off — until then, this script is the
dispatcher." Somebody switched the watcher on. The script was never told. That is
the whole bug.

**What it has been costing beyond the money.** When the two runs disagree about
how they finish — one succeeds, the other trips over a separate known fault —
the failing one writes "failed" over the record of the one that worked. We then
count those failures as evidence for a *different* open bug, and that bug's
apparent rate goes up. So this has been quietly corrupting the evidence we use to
prioritise other work.

**Two things in the original bug report turned out to be wrong**, and I want them
on the record rather than quietly fixed. The report said nothing ever closes
these queue rows when a diagnosis succeeds — it does; that claim was drawn from a
line of text printed by the script rather than from the system's actual
configuration. And both this report and its sibling describe the watcher as
"re-dispatching a job forty-three minutes after it had already been diagnosed";
it did not — it started ninety seconds after the job was filed, alongside the
first run, and what happened forty-three minutes later was that run retrying
itself. The *conclusion* of both reports survives and is if anything stronger. The
stories about how were wrong. One database query caught each.

**A thing I checked before believing it.** The obvious fix is to stop the script
dispatching and let the watcher do it. But the watcher is configured to run "one
at a time", so I expected that to mean every diagnosis in the fleet would now
queue up behind whichever one was running — thirteen minutes each, single file.
That would have been a bad trade. It turns out not to be true: the scheduler
marks a job "done" the moment it has *sent* it, not when it finishes, so the next
one goes out on the following tick. I found two of them running side by side in
the live data. And the watcher is actually *faster* off the mark than the script
was — under a minute, against four and a bit — because the script's own send has
to queue behind everything else on a shared lane.

**One complication, which is why this is a slightly bigger change than "delete
one line".** The script prints a reference number and tells you to save it; it is
how you find the diagnosis afterwards. That reference only works today because
the *duplicate* run happens to use it. The watcher's run files everything under a
different number of its own. So if I just delete the script's send, the reference
it prints stops meaning anything — which is worse than what we have now. The two
bug reports the watcher filed by itself, back in July, already have this problem:
their reference numbers point at nothing at all, and nobody had noticed because
nobody had gone looking.

So the fix has three parts. Whoever picks the job up out of the queue is the one
who gets to run it, decided atomically so two dispatchers cannot both win. The
script asks the database whether the watcher is switched on, rather than assuming
— so if we ever switch it off again, the script goes back to dispatching by
itself with nobody having to remember. And the watcher now writes its own
reference number onto the queue row when it picks it up, so there is one key that
joins the job to the run that did it, whichever route it came in by.

That last part is a small general addition to the platform rather than something
specific to diagnosis: any queue-driven job can now record which run picked it up.

---

**2026-07-28, evening. Done and verified.**

It works. We fired one real diagnosis through the fixed path at 17:04 and it
produced exactly one run instead of two: nothing at all under the script's own
reference number, one set of orchestrations under the watcher's, and a queue row
that closed itself at 17:11 and now carries the reference number of the run that
actually did the work. That last bit is new — until today there was no way at all
to get from one of these jobs to its own diagnosis if the watcher had dispatched
it.

We chose the symptom for that test deliberately rather than using a throwaway.
It asks whether our "only run one of these at a time" setting means anything for
jobs whose work outlives the moment they are sent — a real open question I ran
into while checking this fix wouldn't slow the fleet down. So the credits bought
a finding as well as a proof.

**Two mistakes of mine, both on the record rather than tidied away.**

Deploying the new image quietly halved the chassis: it went from two copies to
one. The second copy had been added by hand earlier today and the deployment file
still said one, so the next deploy by anybody was always going to undo it, in the
direction of less capacity, with nobody told. Restored inside a minute, and the
file now says two so it cannot happen again.

The deploy also killed the review that was assessing this very change. Reviews
run on the same machines the deploy replaces, and it died mid-sentence at exactly
the second the new machine came up. Nothing recovers that — it just sits there
looking busy. Resubmitted on the same reference so the trail stays in one place.
The lesson is simply: get the verdict, then deploy. Where that is impossible —
and it was here, because the database change could not be applied against the old
code without stopping the diagnosis lane altogether — submit after the deploy,
not before.

**What I'd flag for someone else.** Almost every lane we have grew the same way:
a hand-run script first, an automatic loop bolted on later, with the script left
in place "until the loop is switched on". Switching a loop on is one line of SQL
and it never touches the script. I have written the shape into the debugging
guide, but nobody has actually gone and audited the other lanes for it, and I
would not assume this was the only one.
