# Summary — bugfix 291: the reviewer that never existed

*Written 2026-08-17, at closure. Read aloud-able; no jargon that is not explained.*

## What we're trying to do

Our tool auditor inspects the small interactive tools on our sites — mortgage
calculators, image comparers, that sort of thing — and writes up genuine problems it
finds. Some of those it can hand to another agent to fix automatically. The rest need a
person to look at them. The aim of this piece of work was to make sure that second kind
actually reaches a person, instead of vanishing.

## Where we've come from

They were vanishing. For months, every finding the auditor decided needed a human was
addressed to a reviewer called "hitl-review" — an idea written down in April 2026, given
a name, and never built. The system dutifully tried to deliver each finding, found
nobody home, and marked it "blocked" for ever. Fourteen had piled up, and the rate was
increasing: five on the 16th, fourteen by the 17th.

The quiet part was worse than the visible part. While a stuck finding sat against a
page, any *new* finding about that same page was silently discarded as a duplicate. So
the fourteen understated the loss, and nobody could say by how much.

Underneath it was one missing line of configuration. The auditor's instructions never
said what state a review item should start in, and the system's default is "ready to
send out". Give an item a name that sounds like a resting state, and it still gets
dispatched — the name does nothing; only the state field does.

## What we've done

Three things, in a deliberate order, because the obvious order would have caused a worse
outage than the bug.

**First, stop the bleeding.** A configuration change so review items start in the
"waiting for a person" state. Live within minutes, because configuration takes effect
immediately.

**Second, recover what was lost.** All fourteen findings were repaired back into the
human-review queue, each stamped with a record of what had happened to it. They are real,
useful findings — missing labels on form fields, a tool depending on a script that can
never load — and there is already a working button in the admin that turns a confirmed
review item into a fix task. That path had never been reachable for them before.

**Third, fix the class, not the case.** The platform now refuses to create dispatchable
work addressed to an agent that does not exist: it is stopped at the moment of writing,
marked blocked with a clear explanation, and released automatically if that agent is ever
built. It also stopped *forcing* the mistake — the old code demanded every configuration
name some handler, which is precisely why the auditor named an imaginary one. That went
through the review council, which sent it back once for a good reason (it wanted a way to
switch the new check off without a release, and a measurement of its cost rather than an
assurance); both were supplied, and it was approved.

The reason for the strict ordering: the tidy-up step — removing the imaginary name
altogether — is only safe once the new code is running. Done a day earlier it would have
made the auditor unable to file *any* review item, and the failure would have been
swallowed silently. The staged script refused to run until we could prove the new code
was live.

## Where we are now

Closed, and proven rather than assumed.

The release needed two attempts. The first one looked deployed — new pods, real rollout
— and contained none of the new code, because the version label had not been changed and
the machine kept serving the copy it already had. That was not just us: another lane
measured the same thing and found 203 commits of the day's work sitting inert. We wrote
the ten-second check into the shared traps file: compare the fingerprint of the image you
built with the fingerprint of the one running.

The second release was real, and we checked it three ways before touching anything. Then
we proved the new safety net actually catches things, in production, by sending the live
system three work items at once: one addressed to a made-up handler, one to a real
handler, and one addressed to nobody at all. The made-up one was stopped at writing time.
The real one went straight through. The one addressed to nobody was accepted — which is
the shape the auditor now uses, so that arm confirmed the tidy-up was safe. All three test
items were cleaned up afterwards.

That middle arm is the point worth remembering: without it, a check that blocked
*everything* would have looked exactly like success.

There is no longer a single row anywhere in the work queue addressed to the reviewer that
never existed.

## Where we're going

Nothing further is owed on this bug. Three things are written down and belong elsewhere:

- **The duplicate-suppression key is still one-per-page.** A second, different finding
  about the same page can still be squeezed out. No amount of fixing the delivery address
  helps this; it needs one key per finding, which the tools lane already owns as a
  follow-on. The review council raised this and was right to.
- **About forty places in the code write to the work queue directly**, bypassing the new
  door. The older check still catches those, just later and more expensively.
- **Five other places name a different non-existent handler**, harmlessly, because they
  put their items in a resting state. Inert, recorded, not worth a change of its own.
