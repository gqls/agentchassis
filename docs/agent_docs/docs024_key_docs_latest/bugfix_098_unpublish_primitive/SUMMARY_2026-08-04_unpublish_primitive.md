# SUMMARY — the unpublish primitive (`bugs_open/098`) — 2026-08-04

*(Milestone read-out. Previous: `SUMMARY_2026-08-03_unpublish_primitive.md`, written the
morning after the council veto. The series is the record; this file is current state only.)*

## What we're trying to do

Give the platform the ability to take a page OFF a live site as reliably as it puts one
on — and make sure that when the machinery declines to remove something, a human can see
that it declined, and why. Archiving a page used to remove it from every internal model
while the public site kept serving it forever.

## Where we've come from

The capability was built and proven, then vetoed on packaging grounds; the owner ruled
(RFC 011, option B) that the destructive verb stays but is reachable only through the
guarded retraction path. One page was retracted end-to-end as proof. Along the way we
found and fixed the opposite bug — a scheduler was silently re-publishing an archived
page twice a day — and paid five correctness debts the review raised, all approved by
the council across two rounds.

## What we've done (since the last summary)

Everything that was owed is done, and the population is gone. The evidence-preservation
fix (debt 5) shipped, went live, and was verified on the running binaries. The two
consolidation debts (3 and 4) shipped with a lockstep test guarding the shared link
census and a named lifecycle predicate now used at twelve call sites — and the council's
"did you find them all?" objection was right: its requested sweep found five sites the
first pass missed. With approval, all ten remaining frozen pages on leopardess were
retracted in one audited batch: nothing linked to them (checked, with a positive
control), the deletion landed as one commit, all ten now return 404 and the live pages
around them are untouched. The delegated decision is taken and written into the runbook:
archiving does NOT auto-retract — it is a deliberate two-step procedure, because an
automatic file-deleter keyed on a hand-edited flag is exactly the unguarded authority
this platform keeps having to claw back.

## Where we are now

The bug's population is zero for the first time since it was filed. Two caveats, both
recorded. First, the final acceptance check — the ten pages staying down past the next
scheduled refresh — runs tonight; everything indicates it will hold (the re-publishing
defect is fixed and live), but the rule here is to trust the artefact, not the
prediction. Second, the first live batch run refuted half of my own debt-5 fix: the
platform discards the in-record copy of the audit at an earlier point than the one I
guarded and tested. The half that matters — refusals written durably where monitoring
already looks — works and is proven. The small repair (write the full audit through the
same proven channel) is queued as debt 5b, and the underlying plumbing defect now has
three documented faces in RFC 012.

## Where we're going — the decisions that are YOURS

1. **RFC 012** (the await machinery destroys what actions compute — now three documented
   faces, including one that cost another lane a fleet-wide outage): choose between
   **A** — fix the coordinator to merge rather than replace (closes the class for
   everyone; largest blast radius; needs a reader census first); **B** — a named,
   DB-backed helper every findings-plus-await action calls (small, additive, testable;
   my recommendation, amended by tonight's finding: it must write to the database, not
   to a reserved in-memory namespace); **C** — nothing beyond the landmine entry (free
   until the next silent data loss). Addendum 1 (from the 192 lane) also asks who owns
   the shape-preservation question on the `storeActionResult` side — same ruling, same
   sitting.
2. **RFC 011's deferred general question**, only when the next destructive verb arrives:
   does a destructive verb differ in kind from an inert field on a shared adapter's
   vocabulary? (Option C there — a separate destructive vocabulary — was recorded as the
   honest general answer if it recurs.)
3. **Close 098?** Once tonight's re-check holds: the population is zero, the mechanism
   exists and is documented, the non-automation is a recorded decision. Closing then
   needs nothing from the code — say the word and it moves to `bugs_closed/` with
   debt 5b either done or explicitly carried.
