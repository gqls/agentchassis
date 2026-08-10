# SUMMARY — 2026-08-10 (b): the acceptance tests lost their eyes, and nobody could tell

*The read-out on the missing storage client (`bugs_open/243`), written to be said aloud.
The same day's earlier summary covers the interactive pile and the forms; this one is only
about the storage issue.*

## What we're trying to do

Every tool on our sites — the calculators, the quizzes, the builders — is supposed to be
tested automatically, without anyone asking. Each test has two halves. The mechanical
half drives the tool in a real browser: click this, type that, check the answer is
exactly right. The seeing half takes screenshots of the page on desktop and mobile and
has a vision model actually look at them — because a tool can pass every mechanical check
and still be broken in a way only eyes catch: invisible text, a collapsed layout, a
button rendered off screen. We want both halves working on every run.

## Where we've come from

While proving the darts setup-builder tool yesterday, we noticed its test ended on an odd
final step and went looking. The mechanical half had passed cleanly — but the seeing half
had failed with "no storage client — cannot download screenshots". Checking the records:
**every single acceptance run we have a record of — twenty-six out of twenty-six — failed
the seeing half the same way.** Nobody had noticed because the runs still report success:
the failing step's name looks like a deliberate "nothing to look at" branch, and the part
of the result everyone reads — the mechanical checks — was genuinely fine. So all our
green results are real, but they have only ever been half the test.

## What we've done

Traced it end to end, at your direction, and the cause is now certain rather than
suspected. The pointer you gave — storage is handed to spawned containers only if their
type is on the list in the spawn code — is exactly right, and it turned out to be half of
a two-part answer, because these tests run in two different places:

- **The overnight sweep** (twenty of the twenty-six runs) spins up a fresh container per
  test. The spawn code grants storage access only to container types on its list, and the
  tool-acceptance type is not on it. So those containers start without the storage
  settings, and the agent inside never builds a storage client at all.
- **Manual runs** (the other six, including both of ours) don't spawn a container — they
  run inside the standing shared service. That service has no storage configured **on
  purpose**: you ruled on 8 August that bucket access should not be spread across that
  shared deployment, and the note recording the ruling is written into its config file.

Two different reasons, one identical error. Meanwhile the screenshots themselves are
fine — they are taken and uploaded correctly on every run by a different component; it is
only the reading-them-back step that has nowhere to read from. And the scope is exactly
this one agent: nothing else in the live fleet uses this screenshot-reading step, so
nothing else is affected.

It is all written up as **bug 243**, with the evidence for every link in the chain, and
the pattern has gone into the shared debugging guide — including the honest record that
mid-investigation we briefly concluded all runs were inline, which the per-run
"which pod ran this" column disproved for twenty of the twenty-six.

## Where we are now

Diagnosed, filed, nothing yet fixed. The tests keep running and their mechanical halves
keep being trustworthy — every green we have quoted stands. But the seeing half has never
run, on any tool, ever, and until it does we are certifying tools nobody has looked at.

## Where we're going

Three moves, in order of confidence:

1. **The one-line fix for the part that matters most.** Put the tool-acceptance type on
   the spawn list. That mends the entire unattended overnight sweep — the path that runs
   for ever without anyone watching — and it uses precisely the mechanism your 8 August
   ruling intended for granting storage to a specific container type. It is ready to go
   through the review gate on your word.
2. **A decision from you on the manual runs**, which the one-liner does not touch: accept
   that hand-fired tests keep losing the seeing half, or make them spawn a container like
   the sweep does, or revisit the 8 August ruling. Any of the three is workable; choosing
   silently is the only wrong option.
3. **Make this kind of loss visible.** Twenty-six consecutive failures read as successes
   because a quiet final step and a healthy-looking status let them. Whatever else we do,
   a run whose seeing half failed should say so where the result is read, so a green with
   eyes and a green without are never again the same colour.
