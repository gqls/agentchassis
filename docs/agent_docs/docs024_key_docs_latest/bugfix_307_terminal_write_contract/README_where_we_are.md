# README — where we are (bugs_open/307), plain prose, newest at the bottom

## 2026-08-20, evening — what this is and why it needed a plan first

You asked me to look at bug 307: a three-hour GitHub outage on 17 August killed a hundred
pieces of queued work outright, and your ruling at the time was that a passing blip should
put the work back in the queue rather than kill it.

The retry machinery already existed and it already ran. The problem was that it retried
*immediately* — three tries inside a few minutes, all three into the same dead dependency,
and then the item was marked failed for good. Retrying with no wait is the same as not
retrying at all, if the thing you depend on is down for two hours.

While reading it I found something bigger than the outage, and it is the reason this is
worth doing properly rather than adding a wait. There are two different bits of code that
write "this work item failed", and only one of them counts attempts. The other one just
writes `failed` and stops. Five live agents use that second one — so for those, **one
failure is the end, on a normal day with nothing wrong**. Measured over the last fortnight:
141 items died before using up their allowance, 139 of them on their very first try. The
outage cost a hundred items once; this costs a handful every day and nobody had noticed,
because a failed item looks the same either way.

There is a third thing, contributed to the bug by another lane last week. When a handler
deliberately decides something — "a human needs to look at this", "we are not doing this
one" — the failure path can silently write over that decision. On the *success* path, two
sibling pieces of code both refuse to do that, with a comment explaining why. The failure
path has no such refusal. Until this week you could not even see it happening, because
everything wrote the same word. As of yesterday that changed: nine items now carry a
deliberate "won't fix", so from now on an overwrite would be a visible wrong answer.

So the fix is one thing, not three: **a single shared piece of code that every failure path
goes through**, which counts the attempt, waits before the next one, refuses to overwrite a
deliberate decision, and — when it can see that the whole fleet is failing on the same error
at once — puts the item back without spending an attempt at all.

That last part deserves explaining, because the naive version is dangerous. A deleted
repository and a GitHub outage produce the *identical* error message: "404 not found". If we
just treated 404 as temporary and retried forever, a genuinely deleted repo would grind away
until someone noticed. So we do not judge the error text at all. We ask a different
question: *is anything else failing the same way right now?* An outage shows up as many
items, across different customer sites and different kinds of agent, failing with the same
message inside a few minutes. A deleted repo fails alone. I measured a week of history for
false alarms: with that test, exactly three things trigger it, and all three were real
infrastructure outages. Nothing that was one item's own fault ever triggered it.

Two things I want to flag rather than bury.

First, I ran this through the platform's own diagnosis loop, as you asked. It read the code
and the database and built four rounds of evidence that agree with the above — and then its
final write-up step ran out of room mid-sentence and the run died. So there is no formal
verdict, only its working. The bug itself was filed on first-hand verification under the
rule that allows that, and I re-checked everything in the code and the live database this
session, so I am confident in the finding. But I am not going to claim a verdict I do not
have. I also think the run died partly because I asked it about three mechanisms at once
when it is designed for one.

Second, the waiting period needs somewhere to live, and the obvious choices were all traps.
Parking an item as "blocked" looks right and is quietly useless: a cleanup job runs every
ten minutes, releases every blocked item, and wipes the note saying why it was blocked.
That is why there are zero blocked items in the database and always have been. So instead
the item stays in the queue and carries a "not before this time" stamp, which the three
places that pick work up learn to respect. And the waiting *numbers* come from a policy
table that already exists for exactly this purpose and has been waiting for a second user
since it was built.

Next: build it, put it through the review council, and then it needs a deploy before any of
the code half takes effect. The database half goes live the moment it is applied and is
harmless before the deploy.
