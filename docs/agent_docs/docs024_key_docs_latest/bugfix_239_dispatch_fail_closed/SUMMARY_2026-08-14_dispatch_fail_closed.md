# Summary — the dispatch and database-pool lane, 2026-08-14

*First summary for this lane. The series is the record: write a new one at the next real
milestone, never an edit of this file.*

---

## What we're trying to do

Make the chassis honest about failure. This lane started from a single complaint — a request sent
to the platform naming an agent that doesn't exist came back reporting success, having done
nothing at all — and grew into a broader question: how many other places in the message-handling
layer quietly do nothing while reporting that they worked?

The answer turned out to be "several", and they shared a cause. So the work became a sweep with a
principle behind it: when a mechanism cannot work, it should fail loudly, and when a mechanism
has never worked, it should be removed rather than left standing where the next person will read
it and believe it.

## Where we've come from

The original bug was a genuinely bad one. A request naming an unresolvable agent didn't refuse —
it fell through and ran whatever workflow the receiving pod happened to have, then filed the
result under the wrong owner. So the one field an operator would check to tell a real run from a
substituted one said "generic" on precisely the runs where it mattered. Six test dispatches
naming a non-existent agent all came back completed, having accomplished nothing.

Fixing that exposed something worse, and it is the thread that runs through everything since. The
fix was supposed to leave a permanent record of each refusal in the database. It never wrote one.
The reason was a second, unused database connection: the message processor held two, and the one
guarding this code was created from a setting that nothing in production sets. It was empty on
every pod, from the day it was written. So the code behind it had never run once.

That discovery reframed the lane. A neighbouring piece of work found the same two-handle
confusion causing a different problem — the processor was quietly resizing a connection pool it
didn't own, so an operator asking for twelve connections silently got four — and in the course of
proving that, we discovered we had no way to observe a connection pool at all. That lane closed
and handed us two follow-ups: build the missing instrument, and clean up the second handle.

The instrument came with its own lesson, and it is the one I'd tell someone else first. We built
it, put it through review, tested it, deployed it, and confirmed it was serving the right number
on both pods. It was being collected from neither. The monitoring configuration matched on a port
number that a pod *declares* rather than the port it actually serves, and our pods declared only
one of the two. What made it invisible was that around a hundred short-lived worker pods *did*
declare the right port and were all labelled as if they were the chassis — so the metric looked
perfectly healthy, from a group that contained neither of the two machines it existed for.

## What we've done

Four things, all now live and all verified against the running system rather than assumed.

The original refusal bug is fixed: a request naming an unknown agent now fails, loudly, and
leaves a record attributed to the agent that was *asked for*. The connection-pool ownership bug
is fixed, so an operator's configured value survives. The missing instrument exists, is genuinely
being collected, and reported real numbers for the first time — including the contention figures
the pool argument had previously to be conducted without.

And today we removed the second database connection entirely, along with everything that had been
built behind it. That came to six pieces of code, and the reason it's worth describing is that
each one needed a different argument. One was a check that read as though it sent a response to a
parent workflow and in fact sent nothing — both routes through it did exactly the same thing, so
deleting it provably changes nothing. One was a function with a real bug in it that had no callers
anywhere, so nothing had ever received the wrong data; the tempting fix, pointing it at the
working connection, would have brought a dead path back to life and changed what every parent
workflow receives. One was a duplicate-message check, which sounds alarming to delete until you
see that the layer underneath does exactly the same job on a working connection, continuously, and
that the specific rule about releasing a claim when a dispatch fails temporarily lives there too.
Two more were a field and a function nobody read, which weren't in the bug report at all — I
found them by re-checking the file rather than trusting the write-up.

The review council approved it first time. Its two objections were more useful than the approval:
both caught that I had *asserted* the key safety property instead of measuring it. Measuring it
took one file read and came out stronger than the assertion — the two things in question are set
on adjacent lines under a single condition, so one cannot exist without the other.

Verification after deployment produced the other story worth retelling. The usual way to confirm a
change shipped is to search the running program for something you added. This change only removed
things, and both obvious substitutes were the wrong tool — one piece of evidence had already
scrolled out of the logs, and searching for my own change's identifier would have failed *even on
a completely correct deployment*, because releases are built from whatever is current at the time
and that was later than my work. What worked was searching for what is now *absent*, paired with
something I deliberately kept, which must still be found. Absence on its own proves nothing, since
a broken search also finds nothing; the thing that must be present is what proves the search
works.

Two of the three behavioural checks passed outright. The third I could not complete, and I've
recorded it as incomplete rather than rounding it up: it depends on an event that happens about
twice a day and hasn't happened since the deployment, so finding nothing tells us nothing. I
closed the case anyway, on the grounds that this particular edit is provable without observing it
— the old code always ended up using exactly the connection the new code uses directly — and the
one assumption underneath that is what the passing checks measured.

## Where we are now

The lane has no outstanding bug work. Every case it owns is fixed, deployed, and checked against
the live system, and the case files, the register of platform mechanisms, and the debugging guide
all reflect that.

One loose end is deliberate and named: the direct sighting of that twice-a-day event, with the
query written down for whoever is passing when it next fires. It isn't load-bearing.

Two things are open but belong elsewhere. The monitoring configuration we fixed by hand is still
not part of the automated deployment, so nothing reconciles it and it could drift silently in
either direction — that sits with the lane that owns the monitoring plumbing, because wiring it in
changes what a full release applies. And a broken test elsewhere in the codebase, nothing to do
with this work, doesn't compile; I've left it alone but written down where it is, because the quick
build check people habitually use doesn't compile test files and so can't see it.

I should also record that this lane ran for four days on handoff notes alone, without the standing
documents it was supposed to have from the start. I created them today. That's late, and the cost
is real: the wrong turns from the first four days are only recoverable from commit history now,
and the wrong turns are the part nobody can rederive.

## Where we're going

One task remains, and it is a measurement task rather than a coding one: the index of notes that
loads into every session is over its size limit and will start silently dropping entries off the
end. I measured it rather than leaving it as a vague chore, and the measurement contradicts the
obvious plan. The sanctioned way to shrink it is to retire entries about closed bugs — but the bug
entries are comfortably *under* their budget, and the overspend is entirely in the practice
entries, which there's a standing decision to keep loading automatically. So retiring bugs cannot
fix it and moving the practices isn't allowed; what's left is making the existing entries shorter.

The trap there is documented and worth stating, because it has bitten before: before cutting any
detail, check that it genuinely exists in the file it points to, and check its other phrasings. On
the last attempt, most details that looked index-only turned out to be recorded elsewhere in
different words — and every one that really was index-only *contradicted* the file it pointed to.

Beyond that, this lane is done. The next thing anyone does here should be a new case, not a
continuation.
