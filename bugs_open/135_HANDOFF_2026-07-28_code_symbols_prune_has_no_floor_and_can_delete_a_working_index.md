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

---

## FIXED 2026-07-31 — the prune now has to earn the right to run

**Taken by:** the "bugfix 11" session. Workstream:
`docs/agent_docs/docs024_key_docs_latest/bugfix_135_prune_floor/` (standing five).
**Commit:** `10524a03c`. **Register:** CTXA-025. **Council:**
`14239fa4-552f-4821-abaf-ea15ccee4ea5`.

### Still valid when taken, re-grounded 2026-07-31

The unconditional DELETE was unchanged, guarded by nothing but `if commitSHA != ""`.
Live index: one repo, **4,992 rows, five kinds, all at one commit** (func 3048,
method 1025, struct 857, interface 33, alias 29; **592 distinct paths**) — healthy,
so a floor could be armed at no cost. Note `kind` is under a CHECK of eight Go
kinds, so markdown rows cannot exist yet.

### What was built (candidates 1–3, plus 4's intent by another route)

**Candidate 1 — the floor, per cohort.** `platform/orchestration/actions/prune_floor.go`
holds the rule as pure functions of counts: no SQL, no knowledge of any table.
Cohorts are **per symbol kind** *and* **distinct paths** — the second in a
different unit on purpose, because rows measure what a run WROTE and paths measure
what it SAW, which is the property that fails on a truncated tarball. That is
candidate 4's "whole-repo signal" using data already in the table, rather than the
analyser's `file_count`, which has nowhere to be stored for comparison.
Resolvable: `prune_floor_ratio` step config, default 0.5, **0 disables**, `>1`
clamps (an unsatisfiable floor refuses for ever while looking like a working
guard), a cohort exactly AT the floor passes, `stored=0` never refuses, and an
unparseable config value falls back to the DEFAULT — never to 0.

*Detection is per cohort; the refusal is all-or-nothing.* A refused prune is
**self-healing** — the delete is defined against the current commit, not against
this run, so the next healthy run removes what a refusal retained. The cost of
refusing too much is one cycle of staleness; of deleting too much, the index.

**Fails CLOSED** when the cohorts cannot be measured (`refused_unmeasurable`).

**Candidate 2 — the durable surface.** `recordPruneRefusal` writes `doc_notes`
(`subject_type='action'`, `subject_key='index_code_symbols'`, category
`prune-refused`), reusing the existing `insertDocNote` rather than adding an INSERT
path. Suppressed to one row per repo per 24h; never fatal. **There is still no
automated consumer** — it is a record for whoever asks, and the code comment says
so rather than implying a triage lifecycle.

**Candidate 3 — `pruned` is no longer a bare success counter.** The result now
carries `prune_status` (`pruned` · `pruned_floor_disabled` · `refused_floor` ·
`refused_unmeasurable` · `failed` · `skipped_no_commit`), `files_analysed` (the
run's INPUT size, beside the delete count), `prune_floor_ratio`,
`prune_floor_from_config`, `prune_reason`, `prune_cohorts`.

### One thing this case file did not ask for, and needed

A refused prune **retains stale rows by design**, and the read side's freshness
banner reads ONE row (`ORDER BY updated_at DESC LIMIT 1`) and names *its* commit —
so a part-stale index would have been announced as being at the new one. That is
108's lie one layer along, and 108's fix made the wording *more* confident. So
`codeIndexScope` gained a distinct-commit count (riding the COUNT query it already
runs) and `mixedCommitNote()`, wired into **both** readers. Checked against council
18fe4035 / migration 243 before adding it: that reader-side fix was about symbol
BODIES, the freshness half came from 059/108-A (`87d0bcf97`), and **nothing in the
tree measured commit spread** — the only `DISTINCT commit_sha` in the repo is the
new one. The state was already reachable via the pre-existing "no commit_sha →
prune skipped" branch, so this closes a pre-existing gap the fix would have made
more likely.

### The three siblings — NOT converted, deliberately

`populate_nav_tables_action.go:147` (whole-site `site_nav_items`/`site_nav_groups`),
`site_db_actions.go:1474` (`link_registry` per page),
`save_page_sections_action.go:532` (agent-writable `page_components` per page). All
three delete-what-they-did-not-see with no completeness check; one of them is the
016b `game-pathfinding` precedent. **The rule is reusable, the cohorts are not** —
each site must choose classes it can lose independently plus one signal in a
different unit. Two are another lane's live territory this week. Named in
`prune_floor.go`'s header and in CTXA-025's open-review-question.

### Verification — the failing branch was SEEN, not assumed

Per this file's own instruction ("a green run over a healthy repo proves only that
the guard is inert"):

1. **Rule:** 11 unit tests, every branch — refusal text and its remedy, the wiped-kind
   case that a whole-corpus ratio would hide (asserts its own 99% premise first),
   the inclusive boundary, disabled/negative/clamped, `stored=0`, config parsing
   including a garbage value falling back to the default.
2. **SQL:** all three statements PREPAREd + EXECUTEd against the live schema. At the
   current commit every cohort reads `confirmed == stored`; against a bogus
   `commit_sha` every cohort reads 0 — which is what a total-loss run measures.
3. **Image:** `v1.0.1217` pod-grepped with a positive control in the same exec.
4. **Induced refusal live:** see the workstream RUNBOOK for the exact SQL.
