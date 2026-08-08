# Where we are — bugfix 216 (the retry that books itself and then refuses to run)

Append-only, newest at the bottom. Plain prose for reading aloud.

## 2026-08-08 — picked up, fixed in code, not yet shipped

Bug 207 taught the system to say "this failure is worth retrying" correctly. Bug 216
is the discovery that the part which actually performs the retry then drops the ball:
it writes down "retry number 1" and, milliseconds later, declares it has nothing to
retry with and kills the whole job. The reason is almost comic: the code claims the
work (which stamps it "processing"), then asks the database "give me that work — but
only if it's NOT processing". It can never get it back, so it falls to a spare copy in
memory that deliberately never carries the message payload, and without a payload the
replay refuses. The refusal is correct behaviour; feeding it a copy that can never
have the payload is the bug.

The fix is the obvious one and the bug report already argued for it: the claim itself
hands back the complete row, payload included — so pass that row along instead of
asking the database a question that can't succeed. The double-read and the doomed
fallback are gone entirely. One nice sharpening while reading the code: the bug report
guessed that responses landing on the same pod might survive; they don't — every
response reloads state from the database, which never stores the payload. The retry
arm was completely dead for response-driven failures, not just mostly dead.

It's coded, and it's tested in a way I trust: the new test recreates the hostile world
(database unreachable, in-memory copy payload-less) and passes only if the replay
actually goes out on the wire; I then deliberately re-broke the code the old way and
watched the test fail, so it genuinely discriminates. Next: council review, commit,
build, roll, and then the real proof — re-run the live induction from the 207 lane and
watch the retried request actually arrive where it's addressed, because a bumped
counter is exactly what the broken version produced too.

## 2026-08-08 evening — approved, shipped, and proven on the wire

The review approved it first time round (nine reviewers, one minor note). The fleet
release later that afternoon carried it out, and I checked the actual running binaries
rather than trusting the version number: the new code is in, the old fallback is gone.

Then the real test. I staged the same little drama as before: a parent workflow waiting
on a request that nothing will ever answer, a child that fails in the "worth retrying"
way, and the child's answer delivered to the parent. Last time, the parent wrote down
"retry number 1" and killed the whole job four milliseconds later. This time, the retry
actually went out: the original request reappeared on its queue, marked as attempt
number one, three minutes into a ten-minute wait — too early for anything but the fixed
path to have sent it. The parent stayed alive, kept waiting, retried twice more on
schedule, and only gave up when the retry budget ran out — which is exactly right,
because in this staged scenario nobody was ever going to answer.

So the retry machinery the last three fixes were feeding now actually fires. The
remaining piece is the sibling bug (217): a different corner of the code that answers
for failed child workflows still stamps every failure "unrecoverable" — and it answers
first, so its verdict wins. We saw it do this again today during the test. That's the
next job, and it's now worth doing, because a retry it permits will actually happen.
