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
