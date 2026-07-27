# 108 — the code index reports FRESH while 667 commits behind, and never indexes function bodies

**Filed 2026-07-27.** Status: OPEN, unowned.
Two compounding defects in `code_symbols`, the index behind every code-shaped
council check (`code_checks`) and the diagnosis loop's code tier. Together they
make the index answer **"absent"** for code that exists, to the two council seats
whose entire charter is policing claims of absence.

Distinct from `bugs_closed/059` (*index stale 3 weeks — no reindex cadence*).
That was "nothing refreshes it"; the 24h `code-index-refresh` cadence fixed it and
is running. **This is "the refresh runs, resets the freshness clock, and re-indexes
the same stale commit."** 059's own read-time guard is what now reports FRESH.

---

## Symptom

A prior-art search for a service that exists in the working tree returns zero
rows, and the answer is prefixed with a banner saying the index is fresh.

Concretely — the near-miss that surfaced this
(`robot_hands_gripper_dossier/`, `WRONG_CALLS.md` 2026-07-27): a design proposed
building `cmd/gripper-intake/` on the island VM. `cmd/tools-api` +
`internal/tools-api/` — which already did all of it, multi-tool and multi-site —
had shipped to that same VM the next day. **`internal/tools-api/` has zero rows in
the index**, so no `code_checks` lookup could ever have found it.

## Evidence (live, 2026-07-27)

```sql
SELECT count(*) FROM code_symbols;                                    --  4535
SELECT count(*) FROM code_symbols WHERE path LIKE 'internal/tools-api/%';  --  0
SELECT count(*) FROM code_symbols WHERE path LIKE '%score_grippers%';      -- 22
SELECT DISTINCT commit_sha, max(updated_at) OVER () FROM code_symbols;
--  e19aa5d | 2026-07-26 13:36:45+00
```

```
$ git log -1 --format="%h %ad %s" --date=short e19aa5d10
e19aa5d10 2026-07-24  feat(gripper-dossier): score_grippers — server-side port + seed 204
$ git ls-remote origin 086_experience_loop
e19aa5d108eddefb81196363908fe379ac86e447   refs/heads/086_experience_loop
$ git log -1 --format="%h %ad" --date=short HEAD
5c3081e3f 2026-07-27
$ git rev-list --count e19aa5d10..HEAD
667
```

So: rows written **2026-07-26 13:36** (≈24h old → under the 48h threshold →
**reports FRESH**), describing code from **2026-07-24**, **667 commits** behind
local HEAD.

### Defect A — the freshness verdict measures the wrong thing

`platform/orchestration/actions/diagnose_code_lookup_action.go:64-74`:

```go
err := db.QueryRowContext(ctx,
    `SELECT COALESCE(commit_sha,''), updated_at FROM code_symbols
     ORDER BY updated_at DESC LIMIT 1`).Scan(&sha, &updatedAt)
...
age := now.Sub(updatedAt)
if age > staleAfter { /* STALE banner */ }
```

`age` is **how long ago we last wrote a row**, not **how far the indexed commit is
behind the code**. A refresh that re-indexes the same ref resets the clock and
reports FRESH forever. The banner does print `sha`, so the fact is on screen — but
the *decision* ignores it, and the prompts tell reviewers to trust the verdict.

The index tracks a **pushed ref**. In this repo sessions commit locally and push
rarely: `origin/086_experience_loop` has not moved since 07-24 while the branch has
gained 667 commits. **The index is structurally a mirror of what was last pushed,
and nothing anywhere says so.**

### Defect B — `content` never contains function bodies

`platform/orchestration/actions/code_symbols_actions.go:336-352`:

```go
// composeSymbolContent builds the searchable text (embedded AND trigram-matched):
// kind + symbol + signature + doc + path.
```

Live confirmation — the whole `content` column for one row:

```
func init
func init()
platform/orchestration/actions/feed_normalize_action.go
```

Bodies are read on demand by `ReadSymbolBody` (`internal/analysis/symbolbody.go:31`),
which the indexer never calls. So **every `content` check for a route, registry
key, table name, config key or any string literal returns zero rows** —
verified: `'%/api/v1/tools/gauntlet%'`, `'%stop_reason%'`, `'%med_export_json%'`
all return 0.

The contract documented at `diagnose_code_lookup_action.go:29-31` says `content`
matches *"symbol source bodies … (e.g. `"%stop_reason%"`)"*. It does not, and
**that documented example cannot work** — `stop_reason` is a literal that only ever
appears inside a body.

## Root cause

Both defects share one shape: **an empty result is indistinguishable from a
genuine absence, and every qualifier that would have said so is computed from
something other than the fact in question.** 059 identified exactly this
("*a stale index answers 'absent' identically to a genuine absence*") and fixed the
read-time banner — but keyed the banner on write-time, which the refresh itself
resets.

## Who this hurts

`review_prior_art` (`0NN_fix_proposer_v20_prior_art_librarian.sql`) — charter:
*"does it propose BUILDING something that already exists"* and *"does it rest on a
load-bearing absence claim with no evidence"*. Also `review_reuse_agent`, and the
diagnosis loop's `lookup_code_symbols` step (`122_diagnose_agents.sql`).

**The seats policing absence claims are fed manufactured absence, in the one
direction that approves the plan.**

Compounding, and worth its own line: on the **council-gate** the `code_lookup` step
is deliberately not mirrored (`0NN_council_gate.sql:40-45`,
`099_SYNC_gate_roster.py:28`), yet `review_prior_art` — which 099 *does* mirror —
still emits `code_checks` and its prompt promises they are *"answered from the
code_symbols index next round"*. On the gate that promise cannot be kept at all.

## Fix candidates — ordered by what makes the bad state unrepresentable

1. **Make the freshness verdict a function of the indexed commit, not the row
   timestamp.** Compare the indexed `commit_sha` against the current ref and report
   the **commit distance**; STALE when the index cannot be shown to describe the
   code being reviewed. This makes "fresh clock, stale code" unrepresentable
   instead of merely unlikely. Requires the indexer to record which ref it indexed.
2. **Index bodies.** Populate a `body` column (or extend `content`) from the
   `[line_start, line_end]` span already stored on every row. This also answers
   the schema half of the architecture thread's D8b (indexing markdown so
   `WRONG_CALLS.md` and `bugs_open/` become reachable) — same question, already
   settled.
3. **Correct the `content` contract comment** at `diagnose_code_lookup_action.go:29-31`
   so it stops promising body matching, and replace the `"%stop_reason%"` example
   with one that can actually work. Cheapest; do it even if 2 is deferred.
4. **State the ref-tracking fact in the banner** — "indexed from `origin/<branch>`
   @ `<sha>`, N commits behind local HEAD". Even without 1, this makes the failure
   visible to a reader.
5. **Either mirror `code_lookup` onto the gate, or stop `prior_art` promising an
   answer there.** A seat that asks questions into a void is worse than one that
   does not ask, because the asking reads as diligence.

## How to verify a fix

- **Defect B, induced:** `SELECT count(*) FROM code_symbols WHERE content ILIKE
  '%stop_reason%'` — currently **0** while the literal demonstrably exists in Go
  source. Non-zero after the fix. Negative control: a string that appears nowhere
  must still return 0.
- **Defect A, induced:** point the indexer at a deliberately old ref, run the
  refresh, and assert the banner says STALE despite `updated_at = now()`. This is
  the failing branch — a guard whose job is catching a fault must be *seen*
  catching it (the file's own doctrine at `:78-82`).
- **The near-miss regression test:** after reindexing at current HEAD,
  `SELECT count(*) FROM code_symbols WHERE path LIKE 'internal/tools-api/%'` must
  be non-zero. That single query is the whole bug in one line.

## Notes

- `platform/orchestration/actions/registry.go` counts **296** actions across **25**
  category strings against **10** declared in its own header; `site` alone holds
  107. Not this bug, but it is the other reason a "does this already exist?" search
  fails — recorded here so it is not rediscovered.
- Do not "fix" this by widening the schema hint alone. The hint governs SQL checks
  against ten platform tables; `code_symbols` is reached by `code_checks`, a
  different tier.
