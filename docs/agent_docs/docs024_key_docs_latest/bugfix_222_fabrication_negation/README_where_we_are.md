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
