# Where we are — bugfix 222 (plain prose, append-only, newest at the bottom)

## 2026-08-08 evening — picked up

Picked up `bugs_open/222` off the open-bugs list. It's a false-positive bug in
the fabrication gate that protects tool recreations from shipping invented
data (the gate built for bug 020, after a real incident where a directory site
shipped a widget that made up fake vet practices). The gate looks for phrases
like "fake data" or "fabricated dataset" near each other in the generated
code. The problem: it can't tell the difference between a model *admitting* it
invented data and a model *denying* it did. A comment that says "no fabricated
data — starts empty" trips the same wire as one that says "here's our
fabricated dataset" — because both sentences contain the same two words close
together. The mortgagecalculator lane hit this for real on 2026-08-08: a
perfectly good, empty-by-design portfolio tool got discarded and dumped into
the human review queue, and it's happened at least once before to the same
page. They filed the bug, worked around it with a prompt-side instruction, and
left the actual code fix for someone else — which is what I'm doing now.

I checked nobody else was already on it (another thread has bug 220 mid-flight
right now, so gave that a wide berth), and confirmed the bug is still live in
the code as filed. Next: use Fable to draft the fix plan, then implement,
test, and put it through the council review gate before committing, per the
standing rules for platform-code changes.

## 2026-08-08 evening — fixed, tested, committed

The Fable-drafted plan turned up something better than the bug file's own
suggested fix. Fable found that this platform already has a "does this
sentence deny something, or admit it?" checker — built weeks ago for a
completely different problem (a marketing-copy checker that was wrongly
flagging honest disclaimers like "we do not claim this is verified" as if
they were the very overclaim it was meant to catch). Rather than write a
second, separate version of that same idea for the tool-recreation checker, I
pulled the general mechanism out into something both checkers can share, and
gave the tool-recreation checker its own list of "denial" words tuned for
code comments rather than marketing prose — the two need genuinely different
word lists (a marketing checker has to ignore "no" and "without" because they
show up as sales-speak intensifiers; a code comment saying "no fabricated
data" means exactly what it says).

I wrote the failing tests first, watched them fail against the real bug
(catching two of my own test-writing mistakes along the way — one test
didn't actually trigger anything, another one accidentally tripped a
different, unrelated glitch in the same checker), then wrote the fix and
watched them pass. Then I deliberately broke the fix on purpose twice, to
prove the tests would actually notice if it stopped working, and made an
unrelated tweak to prove the tests wouldn't falsely complain about that.

Submitted it to the review council before committing (correlation
`aa2d0d62-4aba-480e-aedc-8be264d53b01`) and committed it straight after,
using the "submitted, verdict not in yet" commit marker rather than waiting —
that's the normal practice here so code doesn't sit around unshipped for half
an hour. Also hit a small, harmless collision: while I still had an
unfinished note sitting in a shared log file, another person's session
appended their own note to the same file and committed it, which
automatically carried my note along with theirs. Nothing was lost, just
filed under someone else's commit instead of mine.

Still to do: check the council's verdict when it lands, and once this ships
in the next build, verify live and tell the mortgagecalculator team they can
drop their workaround.
