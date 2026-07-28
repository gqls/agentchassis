# 135 — the code_symbols prune has no floor: one bad run can delete a working index

**Filed 2026-07-28.** Status: OPEN, unowned.
**Found by `bug_historian` during council review of an unrelated plan** (corr
`7ba5b8c4`, D11 layer 1b), reading the call site rather than the symptom. There is
no failure to point at yet — this is a latent defect with a named recurrence
history, filed before it bites rather than after.

**Split out of the markdown-indexing plan deliberately.** It arrived as an
objection to that plan, I fixed it *inside* that plan, and the fix then generated
three rounds of objections about its own machinery while the actual subject of the
plan drew none. The defect is **pre-existing and independent** — it affects the
Go-symbol rows today, with or without markdown — so it belongs here, not as a
rider on someone else's change.

---

## The defect

`IndexCodeSymbolsAction` (`platform/orchestration/actions/code_symbols_actions.go`)
ends every run with an unconditional reconciliation delete:

```sql
DELETE FROM code_symbols WHERE repo = $1 AND commit_sha IS DISTINCT FROM $2
```

Everything not re-upserted by **this run** is destroyed. That is correct when the
run saw the whole repo, and catastrophic when it did not. **Nothing checks that it
did.**

The run's row count comes from walking a fetched tarball. Any of these produces a
small-but-nonzero result with no error:

- a partial or truncated tarball fetch;
- a moved/renamed directory, so a walk finds far fewer files;
- a permissions change on the extracted tree;
- an analyser failure that yields a short `Output` (the action logs and continues).

In each case the run upserts what little it saw, then **deletes everything else**.
The index is not corrupted — it is *emptied of everything the bad run missed*, and
the next reader gets confident zero-row answers from a corpus that no longer
describes the codebase.

## Why this is worth filing rather than shrugging at

This is the platform's most literally recurrent destructive pattern. From
`016b` §9, quoted by `bug_historian` in the review:

> *"the identical DELETE+INSERT rebuild pattern destroyed a working A\*
> pathfinding game; recurred independently on a second site months later."*

And the read-side consequence is the one `bugs_closed/108` was about: an empty
answer from this index is indistinguishable from genuine absence. **108's fix made
the empty answer more confidently worded** (*"the query was RUN and matched
none"*), so a silently-emptied index now lies harder than it did before that fix.
135 is the write-side of the same exposure.

## What makes it currently invisible

- The prune is not conditional on anything, so there is no branch to log.
- `index_code_symbols` reports `pruned: <n>` as a success counter. A run that
  deleted 4,000 rows because it only saw 900 files reports `pruned: 4000` and
  `indexed: true`. **The number that should be the alarm is presented as output.**
- `orchestration_states` retains ~2 days, so the evidence of a bad run ages out
  before anyone correlates it with a suddenly-thin index.

## Fix candidates — ordered by what makes the bad state unrepresentable

1. **A floor on the prune, per kind.** Refuse to delete a kind's rows when this
   run produced fewer than a configured fraction (suggest 0.5) of what is stored.
   Makes "one bad run empties the index" unrepresentable rather than unlikely.
   **The floor must be resolvable**: a legitimate large deletion (a big refactor
   merged upstream) is a real event, so the ratio wants to be step config
   (`prune_floor_ratio`, 0 disables) and the refusal must name that remedy — a
   guard with no exit is a defect in a guard's costume (`guardian`, round 5).
2. **A durable surface for the refusal.** A refusal that lives only in
   `logger.Error` is the 034/076 shape: *"'No error anywhere' usually means no
   error surface, not no error."* **Learn from the failed attempt in the
   originating plan: `site_work_items` is NOT usable here** — `site_id`,
   `handler_agent` and `pipeline` are all `NOT NULL` and this indexer is
   repo-wide with no site. `doc_notes` fits (nullable `site_id`, append-only, and
   this subsystem already keeps its trail there under
   `subject_key='index_code_symbols'`). There is no automated consumer for either;
   say so rather than implying a triage lifecycle.
3. **Stop reporting `pruned` as a bare success counter.** Report it beside the
   run's input size so a disproportionate delete is legible in the result.
4. **Consider making the prune conditional on a whole-repo signal** — e.g. the
   analyser's `file_count` against the previous run's — rather than on row ratios
   alone. Not costed.

## Verification, when someone fixes it

The failing branch is inducible and must be **seen** to fire, not assumed
(`verify-the-failing-branch`): point the indexer at a ref with a deliberately
truncated tree, confirm the prune is REFUSED, the row count is unchanged, and the
refusal is durable. A green run over a healthy repo proves only that the guard is
inert.

## Provenance

- Objection raised by `bug_historian` (medium) in council corr `7ba5b8c4`,
  rounds 3 and 4 — round 4's text: *"This plan had the context and the call site
  open in front of it and chose to gate only the new addition, leaving the older,
  larger, already-populated half of the same table exposed to the identical
  mechanism."* It was right, and it applies with the markdown plan removed.
- The design refinements in candidates 1–2 are what six council rounds produced
  before the scope was split; they are recorded here so the next thread starts
  from the end of that argument rather than the beginning.
