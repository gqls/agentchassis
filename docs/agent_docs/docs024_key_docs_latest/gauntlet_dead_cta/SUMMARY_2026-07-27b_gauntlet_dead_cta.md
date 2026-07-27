# SUMMARY — 2026-07-27b · the site stopped lying, and now the engine has to explain itself

*Second read-out of the day. The morning one (`SUMMARY_2026-07-27`) covers the Arena
rebuild; this one covers what came after, and the decisions are the interesting part.*

## What we're trying to do

vonc.com is built around a daily argument: you read a provocation, take a position,
and an AI opponent argues back and judges you. Two things have to be true for that
to be worth anything. The page must not pretend — no invented people, no invented
numbers, no button that looks like it does something and doesn't. And the machine
behind it must actually work, and say so plainly when it doesn't.

The first of those is now finished. The second is what we are on.

## Where we've come from

The engine was built, deployed to its own small server, and proven with a real
round-trip. The Gauntlet page was rebuilt against it so the clock, the objectives
and the verdict are all consequences of real replies. The archive was rebuilt so
each past provocation opens its real written case.

That left the Arena, which we had on record as a page that never finished loading.
That record was wrong, and the truth was worse: it loaded perfectly and was almost
entirely invented — twenty-six fictional users with handles, posting opinions they
never wrote, each with a made-up count of how many people had voted it "Genius" or
"Delusional", plus a chain of invented arguments crediting invented contributors.
The box inviting you to file a take saved it to your own browser and nowhere else.
On your instruction we scoped it down honestly rather than building a backend for
it, and that shipped this morning.

## What we've done

**Nothing on vonc.com is fabricated any more.** The platform's own claims scanner
returns zero findings across all forty-nine components. That is the headline, and
it is the thing the whole workstream existed for.

Then three findings that were not on anyone's list.

**The site was publishing its own build instructions to Google.** The description
search engines show under a result was, on sixteen tool pages across six different
sites, the internal specification written for the machine that built the page. The
Arena's was the worst: 1,206 characters that began by announcing the page was "a
fully self-contained client-side experience (no fetch calls, no backend)" and went
on to instruct itself to "embed 5 sample provocations in JS". The cause is one line
— the code that deploys a tool copies the build brief straight into the public
description — and the same function does it *correctly* for the companion guide page
a hundred lines below. I fixed vonc's page and filed the rest as `bugs_open/103`,
because a code fix alone repairs nothing already published.

**The engine throws away the reason it failed.** This was known and filed. What was
not known is how far it went. The bug named two places where the error was
discarded. Reading the code found nine. Auditing *every* error return found seven
more on a second class of failure. Sixteen in total — every single way this service
can fail was returning the same shrug to the visitor and recording nothing.

**And one of those failures was actively destroying its own evidence.** The `/round`
endpoint returned a status code — 502 — whose response body Cloudflare replaces with
its own error page. So the one message explaining what went wrong never reached the
browser at all. We had fixed exactly this on the two sibling endpoints back in July;
`/round` was missed, and it is the endpoint the other two depend on.

The fix went to the review council and **came back asking for revision**, which was
the right call and worth more than it first appeared. The objection was narrow — the
reviewer could not tell from my summary whether I had edited four places or two. My
code was in fact complete. But checking it properly showed that *my own stated count
was wrong*, and that audit is what turned up the seven further failures nobody had
noticed. A reviewer who could not see the whole diff still caused the work to
roughly double in coverage.

## Where we are now

The front end is finished and honest. The engine fix is written, tested and
committed, and back with the council for a second look.

Four decisions are worth recording, because each one could easily have gone the
other way:

**We did not fix a flag that looked wrong.** The Gauntlet's page is marked in the
database as needing a rebuild, which is untrue. I was one command from correcting it
when I found another workstream had already measured thirty-four pages in exactly
that state — all serving perfectly — and had deliberately decided to leave them.
Wrong-looking data is not reason enough to change it; you have to ask what still
reads it, and in this case three separate safeguards had already made it harmless.

**We are instrumenting a theory rather than acting on it.** The tempting explanation
for the failures is that the AI's answer is being cut off mid-sentence. The evidence
does not support it — the successful answers are nowhere near the limit — so instead
of raising the limit and hoping, the new logging labels that specific case distinctly.
If it never appears, that is the answer, and we will have proved it rather than
assumed it.

**We log faults, not mistakes.** When someone sends a malformed request, that is
their error, not ours, and recording it would bury the real failures in noise. So
sixteen genuine fault paths now explain themselves and six caller-error paths
deliberately stay quiet.

**One judgment we did not make ourselves.** To tell a broken AI response apart from
a merely awkward one, the log records a short extract of what came back — and on
some failures that quotes the visitor's own argument. It is public content on a
public game, not personal data, and it is capped. The reviewer flagged it as a
question for a person rather than for code review, and I agree, so it is written up
and waiting rather than quietly decided.

There is also a trap worth knowing about: the bug number **083 belongs to two
different open cases**. Almost every mention of "083" in the project history refers
to the other one, which someone else is actively working. Anyone glancing at the
history would conclude this bug was crowded and leave it alone. It had exactly one
prior mention — mine, filing it.

## Where we're going

The engine fix is inert until the small server it runs on is rebuilt — and it is
worth being clear that the main system was rebuilt three times today and none of
those touch it. That rebuild, and then verifying against the running container
rather than trusting the build, is the next step.

After that, the honest failure rate becomes measurable for the first time, because
until today the service had no request log at all — there was no denominator to
divide by. Once we can see what is actually failing, the remaining candidates on the
bug become answerable instead of speculative: give the call a timeout, and retry once
when the upstream is briefly overloaded.

Then the acceptance harness, which still waits three-tenths of a second for answers
that take ten to twenty seconds and would therefore fail a page that is working
correctly. It must be fixed properly rather than by teaching the page to print
reassuring text that would also appear with the engine switched off.

And still open, waiting on a person rather than on us: the sixteen pages across six
sites publishing their build instructions to search engines, and the question of
whether we should be logging extracts of what visitors write.
