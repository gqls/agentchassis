# SUMMARY — 2026-08-02 — the diagnosis seed scope that never arrived (bug 174)

## What we're trying to do

Make the diagnosis pipeline honour the one instruction an operator gives it about
*where to look*. When you fire a diagnosis with `SEED_SCOPE`, you are naming the
functions you want read. The platform was accepting that instruction, storing it,
echoing it back to you, and then not using it — and there was no way to tell from
the outside.

## Where we've come from

The bug was filed on 2026-08-01 by the lane closing `bugs_open/164`, which
tripped over it while running its own verification: it fired a diagnosis with a
seed scope naming one very large file, expecting to induce a size cap, and the
cap never tripped because the scope it asked for was not the scope that ran.

That lane measured the damage on the intake table: of four intakes that ever
carried a seed scope, three were claimed by the dispatch loop and lost it, and
the only survivor was a manual dispatch fired minutes earlier to work around the
bug. Two of the three losses were other lanes' real investigations. It filed the
finding with the census done, unowned, and moved on.

## What we've done

Fixed it, and found the filing's proposed fix was insufficient in a way that
mattered.

The ticket named one gate. There were **three**, in series, and any two of them
fixed alone would have left the seed dropped — silently, in a new place:

1. The dispatch loop's `claim_item` SQL never *projected* the key out of the work
   item, so the mapping the ticket wanted to fix had nothing to read.
2. The `input_mapping` allow-list — the gate the ticket named.
3. A type gate. `QueryDatabaseAction` stringifies every column it scans, so a
   jsonb list arrives as text; the helper at the far end returned nil for text,
   which is indistinguishable from "nothing was supplied".

Gates 1 and 2 are config (migration 289, applied and verified live). Gate 3 is a
widening-only change to `ExtractStringListHelper`, pinned by a rejection test so
nothing that returned nil before can start returning something now.

We also made the failure visible rather than invisible. The consuming action's
scope fallback chain is correct by design and is precisely what converted a lost
parameter into a successful run with different inputs — so it now records which
arm supplied the scope, and warns in the report when nobody chose it. It does not
claim to know whether a seed was absent or confiscated, because it cannot.

And we built the class-closing check the ticket asked for —
`config-key-audit --relay-gaps` — after measuring and **rejecting** two more
obvious versions of it, one of which was blind to this very bug.

## Where we are now

- Config half: **live**, verified by reading it back off `agent_definitions`.
- Code half: **committed** (`f51acb2bb`, `10789dfe6`), inert until the next
  chassis roll. The bug therefore stays **OPEN** by the standing rule — fixed and
  live is the bar, and a committed-but-unrolled fix is still reproducible.
- Council: **APPROVED at round 1**, correlation `081d98b3`, six advisory
  objections. Four were answered with work: a helper reused instead of
  reinvented, a landmine written, the shared helper's other consumers measured
  rather than asserted safe, and an absence claim re-checked with the grep
  attached.
- One correction to our own submission: the blast-radius figure we gave the
  council ("14 live steps project jsonb") was measured with a filter that also
  caught text casts and predicates. The real figure is **1** — the projection
  this fix added. Logged in `WRONG_CALLS.md`.
- One thing deliberately left undone: `QueryDatabaseAction` stringifying jsonb is
  the deeper cause. It has zero currently-affected consumers, so it is a
  prospective trap rather than a live defect, and it is recorded in `LANDMINES.md`
  rather than fixed blind.

## Where we're going

1. The chassis roll, then verify at the **artefact**, not the tag: pod-grep both
   replicas for a string this change adds, fire a real seeded diagnosis through
   the default path (no `DISPATCH=1`), and assert both that the field arrived and
   that `scope_source` says `seed`. Field-present is a weaker claim than
   scope-used, and the fallback chain is what pulls them apart. Then close.
2. Wire `--relay-gaps` into `config-key-audit`'s argument dispatch. It is
   unwired today only because another session held that file mid-change; the
   audit script refuses to report clean rather than pretending, so the gap is
   loud, not silent.
3. Two dispatcher-shaped relays (`report-dispatch-loop`,
   `build-pipeline-trigger`) are reported as **uncovered**. Registering them
   means first reading their handlers' contracts — which nobody has done, and
   asserting it unread would put them in exactly the state 174 was already in.
