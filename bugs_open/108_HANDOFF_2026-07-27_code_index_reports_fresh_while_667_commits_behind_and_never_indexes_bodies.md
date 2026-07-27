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

**ADDED 2026-07-27 (late, architecture_review thread) — a fourth consumer, and it
is worse than the three above because its prompt actively instructs reliance on
the broken half.** `review_architecture` (seated on `feature-designer` ~13:30
2026-07-27, i.e. after this case was filed) closes with:

> `kind "symbol" matches symbol names, "content" searches source bodies, "ls"
> lists indexed paths.`

Given this bug, the middle clause is false, so the seat will issue `content`
checks for exactly the things a forward-fitness judgement turns on — a route, a
registry key, a config key, whether anything still references a symbol — and get
a zero it has no way to read as manufactured. Note the same prompt *does* carry
the right instinct on the SQL tier (*"Treat an empty result as 'no precedent
found', NOT as 'this is novel'"*) and carries **no equivalent warning on the code
tier**, which is the tier that is actually broken. Verified live 2026-07-27:
`code_symbols` = 4,535 rows, **0 markdown**, 530 files, and `content` holds
declarations only (`max(length(content))` = **451**; the longest `func` row is its
signature line). Until candidate 2 or 3 lands, treat that clause as a known-false
promise rather than a seat defect — no prompt edit is worth making twice, and
fixing it in the prompt alone would leave the other three consumers lying.

Cross-ref: `docs/agent_docs/docs024_key_docs_latest/architecture_review/`
(`NOTES_architecture_seat.md`, and §6 item 5 of
`HANDOFF_2026-07-27_continue_here.md` — candidate 2 below is most of that item's
answer, so the two should be solved together rather than separately).

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

   > **WIDENED 2026-07-27 evening (architecture_review thread) — measured from the
   > live config, and it is worse than "the gate" in two ways.** Candidate 5 names
   > `council-gate`. In fact **`code_lookup` exists on `fix-proposer` ONLY**, and
   > even there it does not cover every seat:
   >
   > | lane | `code_lookup` step | whose `code_checks` are answered |
   > |---|---|---|
   > | `fix-proposer` | present | 6 seats — editquality, bug_historian, reuse_agent, guidelines, tooling_provenance, guardian |
   > | `feature-designer` | **absent** | **none** |
   > | `council-gate` | **absent** | **none** |
   >
   > **(a) `review_prior_art` is absent from `fix-proposer`'s `code_check_fields`.**
   > So the seat whose charter is *"does it propose BUILDING something that already
   > exists"* has its code questions dropped on **all three lanes**, not just the
   > gate — including the one lane that has the machinery. Its `checks` (SQL) are
   > collected; its `code_checks` are not.
   >
   > **(b) `feature-designer` has no `code_lookup` either**, which candidate 5 does
   > not mention because the architecture seat post-dates this case. So
   > `review_architecture` — now the platform's **only** forward-fitness voice after
   > the owner's 2026-07-27 D9 ruling — emits `code_checks` that are *never routed
   > anywhere*, on top of the `content` defect. Its prompt still says *"Answered
   > from the `code_symbols` index next round."* Doubly false on that lane.
   >
   > Reproduce (the whole finding in one query):
   > ```sql
   > SELECT type,
   >        default_config->'workflow'->'steps'->'code_lookup' IS NOT NULL AS has_code_lookup,
   >        default_config->'workflow'->'steps'->'code_lookup'->'config'->'code_check_fields' AS answered_for
   > FROM agent_definitions WHERE type IN ('fix-proposer','feature-designer','council-gate')
   >   AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
   > ```
   >
   > **Interim mitigation SHIPPED 2026-07-27 18:3x (config-only, live, no image).**
   > Not a fix — the seats still cannot look code up; they are merely no longer
   > lied to. A `CODE INDEX LIMITS` paragraph now closes **all 15** prompts that
   > mention `code_checks` (fix-proposer 7, council-gate 7 via the 099 mirror,
   > feature-designer 1), stating that the index holds declarations only, mirrors
   > the last *pushed* ref, and that on some lanes `code_checks` are not answered at
   > all — therefore **an empty or missing code result is NO INFORMATION, never
   > absence**, and an absence claim belongs in `missing` for a human. It is
   > deliberately **lane-agnostic**: the 099 mirror forces `fix-proposer` and
   > `council-gate` to carry identical prompt text, so a lane-specific sentence
   > could not survive it — which is itself worth noting as a constraint on any
   > future per-lane wording. Verify:
   > `… WHERE prompt_template LIKE '%CODE INDEX LIMITS%'` → 7 / 7 / 1.

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

## Re-verified 2026-07-27 15:5x UTC, post-roll triage sweep — the mechanism caught in the act

Chassis is on **v1.0.1174** (pod `agent-chassis-5994dc6d6c-pt8v9`, started
`2026-07-27T15:11:15Z`). This bug is **unaffected by the roll** — it is a data
defect, not an inert Go fix. Still **OPEN**, and the evidence is now stronger than
at filing.

**The refresh ran, completed, and re-indexed the same stale commit — measured, not inferred:**

```sql
SELECT name, enabled, last_triggered_at, last_completed_at, input_data::text
FROM scheduled_tasks WHERE name = 'code-index-refresh';
--  code-index-refresh | t | 2026-07-27 13:35:30+00 | 2026-07-27 13:35:30+00
--  input_data: {"repo": "agentchassis", "owner": "gqls", "language": "go"}

SELECT DISTINCT commit_sha, max(updated_at) OVER () FROM code_symbols;
--  e19aa5d | 2026-07-27 13:36:50+00
```

A **successful** run finished at 13:35 today. `updated_at` moved to 13:36 today
(age ≈ 2h, well under the 48h threshold → banner says **FRESH**). `commit_sha` did
**not** move: still `e19aa5d` (2026-07-24). That is Defect A observed end to end in
one cadence cycle, rather than reconstructed from two snapshots.

**The distance has grown while nothing changed.** 667 at filing (~06:00 today) →
**746** now (`git rev-list --count e19aa5d10..HEAD`). `git ls-remote origin
086_experience_loop` still returns `e19aa5d108…` — the pushed ref has not moved in
three days. Measured commit rate: **1,335 commits in 7 days** (≈191/day), so 746
commits ≈ **3.9 days** of drift, and the gap widens ~191/day for as long as nobody
pushes.

**The near-miss regression test now has its exact numbers:**

```
$ git ls-tree -r --name-only e19aa5d10 | grep -c '^internal/tools-api/'   # 0
$ git ls-tree -r --name-only HEAD      | grep -c '^internal/tools-api/'   # 12
```

The index returns zero rows for `internal/tools-api/` because **that code did not
exist at the commit the index describes** — while the banner reports the index
fresh. This is the cleanest possible statement of the bug: the absence is real *of
the indexed commit* and false *of the code under review*, and nothing in the answer
distinguishes the two.

**Defect B re-confirmed live**, and now with a real consumer artefact:
`max(length(content))` = **451**, `0` markdown rows, `content ILIKE '%stop_reason%'`
= **0**. A live `collected_data->'code_lookup'` from a run at 14:40 today shows
`code_context` composed purely of `path :: symbol` + signature + doc line for all
12 hits — **no body text anywhere in the answer**, exactly as the composer implies.

**Blast radius, measured:** **34 orchestration runs in the last 7 days** carry
`code_checks` in `collected_data` (latest 14:40 today). Every one was answered from
this index.

> **[CORRECTION to this file's own verification recipe]** *"Defect A, induced: point
> the indexer at a deliberately old ref"* is **not executable as written**. The
> `code-index-refresh` task's `input_data` is `{repo, owner, language}` — there is
> **no ref/branch/commit parameter to point**. `commit_sha` arrives from an upstream
> repo-analysis step (`code_symbols_actions.go:174`,
> `commit_field: "repo_analysis.commit_sha"`), so the ref is whatever that step
> fetched. Any fix under candidate 1 must therefore *add* the ability to name a ref
> before the induced test can be run at all — which is additional scope the
> candidate does not currently mention. Caught by reading the scheduled task's
> `input_data`, not the Go.

**Ownership note for the next reader:** the header says "unowned", but
`scripts/who-owns.py 108` names **`architecture_review`** (ACTIVE, 21 commits/14d,
16 mentions) — that thread grounded the same figures today and carries this as §6
item 5 of its `HANDOFF_2026-07-27_continue_here.md`. Route work there rather than
opening a fresh lane.

## Notes

- `platform/orchestration/actions/registry.go` counts **296** actions across **25**
  category strings against **10** declared in its own header; `site` alone holds
  107. Not this bug, but it is the other reason a "does this already exist?" search
  fails — recorded here so it is not rediscovered.
- Do not "fix" this by widening the schema hint alone. The hint governs SQL checks
  against ten platform tables; `code_symbols` is reached by `code_checks`, a
  different tier.
