# Two dispatchers, one queue — `bugs_open/124`, closed

*2026-07-28. Written to be read aloud.*

## What we're trying to do

Stop paying twice for every bug we send into the diagnosis loop, and make the
record of a diagnosis findable from the job that asked for it.

## Where we've come from

The diagnosis loop has one documented front door: a script called `090_TRIGGER`.
You give it a symptom; it writes a row into the work queue saying "this needs
diagnosing", and then it sends the job off to be diagnosed. That was correct when
it was written, because nothing else was watching that queue.

Later we built the thing that watches the queue — a dispatcher that checks every
minute for anything waiting and picks it up. It shipped switched off, on purpose,
and the script's own notes said so: *"the task ships disabled — until then, this
script is the dispatcher."*

Somebody switched it on. Switching it on is a single line of SQL, and it does not
touch the script. From that moment the script sent the job, and about a minute
later the watcher found the same row still sitting there and sent it again.

## What we've done

Found that every manually-filed diagnosis has been running twice — two
independent twelve-to-fourteen minute runs of a large model on the same bug, at
the same time, neither aware of the other, on two reference numbers that could
not be joined to each other. We caught one happening live: another session filed
a bug at 16:37:51 and both dispatchers took it forty-seven seconds apart.

Two things in the original report turned out to be wrong, and both are recorded
as corrections rather than quietly fixed. It said nothing ever closes these queue
rows on success — it does; that claim came from a line of text the script
*prints*, not from the system's actual configuration. And it described the
watcher as re-dispatching a job forty-three minutes after it had been diagnosed —
it did not; it started ninety seconds after the job was filed, and what happened
forty-three minutes later was that run retrying itself. One database query caught
each. The conclusions of both reports survive; only the mechanisms were wrong.

We also found a half of the problem nobody had noticed. The reference number the
script prints and tells you to save only ever worked because the *duplicate* run
used it. Jobs the watcher filed by itself, back in July, carry reference numbers
that point at nothing at all. So simply deleting the duplicate would have made
things worse for everyone.

The fix has three parts. Whoever takes the job out of the queue is the one who
gets to run it, decided in a single atomic step so two dispatchers cannot both
win. The script asks the database whether the watcher is switched on rather than
assuming — so if we switch it off again, manual dispatch resumes with nobody
having to remember. And the watcher now writes its own reference number onto the
queue row as it picks the job up, so one key joins the job to the run that did
it, whichever route it came in by.

That last part is a small general addition to the platform, not something
specific to diagnosis: any queue-driven job can now record which run picked it up
without new Go code being written for it.

## Where we are now

Fixed, live and verified against a real diagnosis rather than a test one. The run
we fired at 17:04 produced exactly one diagnosis instead of two, nothing at all
under the script's own reference number, and a queue row that closed itself and
now points at the run that did the work. We chose a symptom that was a genuine
open question, so the credits bought a finding as well as a proof.

The ticket is closed and moved to `bugs_closed`. The transferable lesson is
written into the debugging guide, because this shape is not specific to
diagnosis: most of our lanes grew a manual trigger script first and an automatic
loop afterwards, and turning the loop on never turns the script off.

Two things went wrong along the way and are on the record. Deploying the new
image reset the chassis from two replicas to one, because the extra replica had
been added by hand and the deployment file still said one — so the next deploy by
anybody was always going to undo it silently. Restored within a minute, and the
file now says two. And the deploy killed the review that was assessing this very
change, because reviews run on the machines the deploy replaces. Resubmitted.

## Where we're going

Nothing is left open on this bug. Two things it touched are worth someone's time:
the same "manual script plus automatic loop" pattern exists in other lanes and
nobody has audited them; and the rate of the neighbouring bug about hung
dispatches was partly inflated by these duplicates, so it needs re-measuring from
today forward rather than from its filed history.
