# Summary — bug 315: the platform cannot tell "didn't need publishing" from "failed to publish"

2026-08-19. Written to be read aloud.

## What we're trying to do

Make it possible to answer one question honestly: **has this page actually reached the website?**
Today the database has a field called `deployed_at` that everything treats as the answer, and it is
not the answer to that question — it is the answer to a different one.

## Where we've come from

The lane rebuilding webdesign.co.uk's tools hit the sharp end of it. A tool page was rebuilt four
times, every rebuild reported success, the timestamp was refreshed each time, and the public site
carried on serving the old tool for about six hours before publishing itself with nobody doing
anything. They filed it, said clearly it was not theirs to fix, and moved on. A second lane added a
similar-looking case from another site and sized what it thought was the affected population.

## What we've done

We took it as a fix lane, checked nobody else was on it, and confirmed it is still real.

We traced the whole path from "page rebuilt" to "page on the website", and measured rather than
assumed. Five agents write that timestamp. **Two of them write it before the deployment has even been
requested**; the other three write it after handing the page over and then throwing away what came
back. There is no arrangement of those five in which the field could mean what people read it to mean.

We found the machinery for the honest answer already exists and was abandoned: two database fields
built for exactly this, both completely empty, with **no code anywhere in the repository — including
tests — that has ever written to either**. We also found the platform's own reference document
claiming this traceability already exists, which is false in all three of its parts; we corrected it,
because the automated reviewers read that document as fact.

We put a fix plan through the review council. It came back **revise**, and the round paid for itself
twice over: it caught that our headline claim was broader than what we were actually shipping, and it
forced a check that revealed our plan was quietly re-adding a column somebody had deliberately
removed — on a decision the estate's own migration notes say belongs to the owner, not to a
bug-fixer. Both are now corrected. The architecture seat ruled that changing what the deployment
service reports back is not a bug fix, because nineteen live steps across sixteen agents consume that
response; we have written that up as RFC 038 and taken it out of the fix.

And we caught ourselves being wrong about something larger. We had reported forty pages as apparently
stale. **All forty were fine** — proven by pulling content out of the database and finding it in the
served page. They had been rebuilt into identical content, which produces an empty change that
correctly copies nothing.

## Where we are now

That last mistake is the finding. It took four steps and a judgement call to establish that *one*
page was correctly published, and until we did it, "this page never needed republishing" and "this
page failed to republish" were **identical in every signal the platform produces** — same work-item
status, same orchestration outcome, same timestamp, same success from the deployment service, same
untouched file date on the website.

So the bug is not that pages fail to publish. **It is that nothing we have can tell those two apart.**
That also settles the design: the cheap fix the bug report proposed — alerting when the timestamp is
newer than the website's file — we ran, and it produced forty confident false alarms on our busiest
site that stayed false for eighty-five minutes. It cannot be rescued with a grace period. Recording a
fingerprint of what we sent is not a refinement of that idea; it is the only version that works.

Nothing has shipped. The investigation, the plan, the review round and the architecture referral are
all written down and committed.

## Where we're going

Three things, in order.

**One decision belongs to the owner**, and everything else waits on it: wire up the two abandoned
fields, or drop them. This is the third independent discovery that they are empty, and the note left
by the last person to find them says explicitly that the call is the owner's.

**One change is ready and needs no code**: delete the premature "deployed" stamp from the two agents
that write it before requesting a deployment. Both already call a routine that stamps it properly
afterwards, on a better identifier. The architecture reviewer looked at this piece specifically and
said proceed. It is held pending the owner's word only because it alters the live build pipeline the
instant it runs.

**One is in architecture review**: making the deployment service report the commit reference it
already computes and throws away, plus a fingerprint of the bytes it sent. That is RFC 038, and it
needs the survey of nineteen consumers that the reviewer correctly said we had not done.
