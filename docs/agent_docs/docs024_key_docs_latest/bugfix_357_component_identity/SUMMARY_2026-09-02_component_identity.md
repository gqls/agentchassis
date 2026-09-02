# SUMMARY — component identity (bugs_open/357) — 2026-09-02

**What we're trying to do.** Twenty-two live pages across three sites each stored a whole
working interactive tool — a calculator, a simulator, a comparison widget — in a database row
that claimed to be the shared "hero" component, the simple title band most pages open with.
The danger was never the label itself: it was that any routine regeneration of a "hero" would
replace a working tool with a title band. The job: make that impossible, without moving a
byte of what the sites serve.

**Where we've come from.** The lane built its machinery in phases: a provenance stamp, a
save-time guard, and a birth-fix (live since 25 August) that stops new rows being mislabelled.
The repair of the existing 22 was designed twice. The first design (migration 578) retyped
them onto one shared "adopted fragment" component and depended on an untested save-time
conservation mechanism — and this lane proved that dependency could only ever be tested
vacuously. A sister lane (lendzy) then landed the alternative shape on their own smaller case:
give each tool its OWN component whose template IS the stored bytes, so regeneration
reproduces the tool by construction. The owner chose that shape ("Option B"). Along the way
the lane also found, fixed, shipped and closed a pod-killing crash bug (408) that any of these
pages could trigger.

**What we've done.** Measured everything the design rested on (drift mechanics, a third
repoint leg nobody had named, one function-name collision requiring a fork, losslessness and
template-validity across all 22 bodies); drafted migration 701 against the sister lane's crib;
took it through three council rounds, each of which genuinely improved it; the owner applied
it by hand — pilot first, then the remainder — with every guard passing and every byte
unchanged. Verified at all 22 served pages with discriminating controls.

**Where we are now.** The bug is CLOSED: its own predicate returns zero, verified at the
served artefacts. The best evidence arrived unplanned: hours after the repair, the ordinary
news-refresh machinery rebuilt one repaired page end to end — planned the new identity,
rendered the adopted template, reproduced the tool to within a trailing newline, and kept the
adopted identity through a full row rewrite. The safety property we chose this design for
held in production without being asked. One honest correction stands in the record: the
migration's own verification rerenders took a byte-reship path (a prose value in a parsed
field — a known trap another lane spotted), so they proved delivery, not regeneration; the
organic rebuild is what proved regeneration.

**Where we're going.** The lane closes. What it hands on: bugs_open/406 (the sibling refusal
defect — diagnosed, fix owed, council-gated); three parked contrast items on the vetcomparison
tool whose natural home is now its own template (their design pass is queued); the
walker-family census and the 578 file kept as the record of the road not taken. The class
itself is guarded at birth, and a regression re-arms nothing quietly — the growth-guard in 701
aborts loudly if a twenty-third row ever appears.
