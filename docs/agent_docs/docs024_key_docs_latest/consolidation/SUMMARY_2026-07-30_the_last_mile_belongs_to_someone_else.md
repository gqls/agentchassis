# Consolidation programme — where we are, 2026-07-30

*Fourth in the series. The three before it are 07-27 (what we found), 07-28 (what the
audit got wrong) and 07-29 (getting it used). Written to be read aloud.*

---

## What we're trying to do

The platform grew by copying. When a new site or service needed something that
already existed somewhere else, the quickest route was to write it again nearby, and
over a couple of hundred builds that became the normal way to add anything. The
result is not messy code so much as *divergent* code: several versions of the same
idea, each slightly different, none of them wrong enough to notice, and no way to
know which one you are looking at.

This programme picks the cases where that divergence is actually dangerous, builds
one shared version, and — since the owner's ruling last week — gets it **used**. That
last clause is the whole of the current phase. A shared version that nothing calls
has not removed a single copy; it has only added one more.

## Where we've come from

Two shared packages came out of the earlier phases. One sends email, which sounds
trivial until you notice that until last week there was no email sender anywhere in
the code we actually build and deploy — the only working one lived in a virtual
machine's own app, outside the build, untested, and unshippable by our image
pipeline. The other holds the defences a public web endpoint needs: working out who
a visitor is, limiting how often they can call you, and spotting a bot filling in a
form. Three versions of that limiter existed, and the weakest of them was guarding
the only genuinely public endpoint we had.

Both were built, tested and approved by the review council on the 28th. Both had
nothing calling them. On the 29th the owner settled the question I had raised about
that — does this programme's job end at building the shared thing, or does it run to
getting it adopted — with two words: *get it adopted*.

## What we've done

The immediate blocker turned out to be in our own shared code, and finding it was
the most useful thing that happened.

The security package decided who a visitor was by reading two particular labels that
web servers attach to a request. That was correct, and it was justified in the code's
own explanatory comment by how one specific web server behaves. We have two. The
other one — the one in front of the service we were about to adopt this into —
behaves differently: it does not write the first label at all and passes a
visitor-supplied one straight through, and it *replaces* the second rather than adding
to it. So adopting our shared package, unchanged, into that service would have made
every visitor look like the same internal address. It would have produced a useless
answer while sitting in the code exactly where a reader would expect the fix to be.

The fix is small and it is about honesty rather than cleverness. The function now
*requires* the caller to say which web server is in front of it. There are three
ready-made answers to that question and no default, because a default that happens
to be right in one place and silently wrong in another is precisely the defect. The
old behaviour is one of the three, unchanged, so the fix we proved against a real
production incident earlier this month is still proved.

That went through the council and was **approved first time**. It also settled a
question I had raised against my own submission — whether tightening a shared
mechanism needs the heavier architecture review — in favour of the normal route, on
the grounds that hardening a shared thing's contract *before* its first real consumer
is the cheapest moment in its life to do it. That is a distinction worth keeping: it
is the opposite of the case a fortnight ago where a shared capability arrived hidden
inside a bug fix and was rightly vetoed.

Two more things got done rather than deferred. Both packages are now written into the
register of things that exist and can be called — neither was, which is exactly the
blindness that register's own warning banner exists to catch. And the one reviewer who
asked for something rather than approving outright wanted a loose end tied before it
became folklore: the safety of the new arrangement rests on a check about where a
request came from, and that check's exact boundary was implied rather than stated. It
is now pinned by tests, including one driven through a real network socket rather than
a hand-typed string, and both tests were verified to *fail* when I deliberately broke
the thing they guard. That turned up a fact worth knowing before anyone deploys this:
a proxy sitting behind certain kinds of shared internet address is not trusted, and
its answer is ignored. It fails in the safe direction, but it decides where this can
be used.

## Where we are now

Everything on this programme that is ours to do is done, committed and approved.

What is left is not code, and this is the part worth your attention. Both remaining
adoptions land inside the same service — it is the security package's only target, and
the email package's only waiting consumer would live there too — and that service
belongs to another thread. So I wrote up the evidence and a ready-to-apply patch,
following that service's own existing conventions rather than imposing ours, and filed
it in their directory. That was the 29th, early afternoon.

This morning I checked what happened to it, and the answer is nothing. They have been
busy — six commits since — but none of them touched that service or the patch. And
when I looked at *why*, I found something more useful than a nudge: their own
start-here document contains a line saying, in effect, "the consolidation people may
get in touch, nothing owed yet." I dated that line. It was written five hours **before**
my patch arrived, and it has survived four later edits of the same file untouched. So
the next person to pick up that thread would read "nothing owed" with a finished patch
sitting two files away.

That is not their failure and it is not mine. It is a gap we have hit before in a
different form: **putting a file in the right place is authoring, not delivering.**
Nothing in our setup tells a thread that a new document in its own directory applies
to it. I have appended a dated note to their start-here document — appended, with
none of their words changed — saying the contact has arrived, what the finding is, and
that it concerns their service whether or not they care about our shared code. Because
it does: the address they store for every visitor to that tool has been the same
single value in all eighty-three records since it went live, which means the limiter
they think is per-visitor is one shared bucket, and the identity they store has never
told anyone apart. No attacker required, and it is measured rather than reasoned.

## Where we're going

Three things, and the first is a genuine choice for you rather than a task.

The adoption now needs that other thread's agreement, not more work from me. The note
is in the place their next session will read, and the patch is three small edits that
will not go stale. So the honest options are to leave it and let them pick it up, or —
if it stays untouched — to raise the routing with you, because a public endpoint whose
rate limit does not actually distinguish visitors is a live defect rather than a tidy-up.
What I am deliberately not going to do is apply it myself. That service has an open bug
against it owned by that thread, and reaching into someone else's code is the exact
thing our contribute-don't-fix convention exists to prevent.

There is one debt I owe the council and intend to honour: a reviewer approved an earlier
round but asked that a claim I made about how our workflows declare their data contracts
be independently checked before anyone treats it as settled precedent. I am still the
only person who has read that. It is contained and worth doing properly.

And there is one question I have deliberately left as a question. A reviewer suspected a
particular unsafe coding pattern recurs across the platform. I showed the specific
evidence offered for it was wrong — the file it cited retracts its own headline number —
but the underlying suspicion is reasonable and unmeasured. It needs counting before it is
a claim, and I would rather leave it labelled as unmeasured than file a bug on a hunch
and have someone inherit it as a fact.
