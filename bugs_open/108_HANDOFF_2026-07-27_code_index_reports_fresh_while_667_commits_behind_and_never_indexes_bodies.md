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
   > **ROUTING HALF SHIPPED 2026-07-27 19:4x (config-only, live, no image)** —
   > candidate 5 is now **half done, deliberately**:
   > - `fix-proposer.code_lookup.code_check_fields` 6 → **7**: `review_prior_art`
   >   added. It was the only seat emitting `code_checks` and missing from the
   >   answer list, on the one lane that could answer it.
   > - **`feature-designer` gains a `code_lookup`** (`run_checks → code_lookup →
   >   repropose`), answering `review_architecture`. Reachability re-checked: 25/25
   >   steps, no orphans; `review_fields` 6 and `hard_veto_from` unchanged.
   > - **`council-gate` deliberately NOT changed.** Candidate 5 offers "mirror
   >   `code_lookup` onto the gate" as an option; `099_SYNC_gate_roster.py:24-29`
   >   already states why it is excluded — `code_lookup`/`repropose` *"serve the
   >   blind reproposer, which the gate has no equivalent of (its authors read the
   >   objections themselves)"*. That reason is sound and still true. The same test
   >   is what **includes** `feature-designer`, which does have a blind reproposer.
   >   **The residual asymmetry, which candidate 5 should be read as still covering:**
   >   the gate's authors receive SQL `check_results` in the verdict note but never
   >   code ones, so a gate seat's `code_checks` remain unanswerable. Fixing that
   >   means surfacing code results into the verdict note — NOT bolting a
   >   reproposer-shaped step onto a lane that has no reproposer.
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

---

## 2026-07-28 07:10 — DEFECT B IS FIXED AND LIVE. Defect A is now demonstrated on a running diagnosis, and B's fix has made A *more* dangerous.

Contributed by the owning thread (`architecture_review`), which built candidate 2
as D11 layer 1 under council approval `18fe4035-4fa6-4079-ab44-8541d6e58944`.

### Defect B (bodies never indexed) — CLOSED, verified live

| check | before | after |
|---|---|---|
| `count(body)` | 0 of 4,535 | **4,535 of 4,535** |
| `body ILIKE '%stop_reason%'` — the contract's own example | 0 | **6** |
| `content_hash` drift (nothing silently re-embedded) | 0 | **0** |
| indexer log | — | `bodies_sliced=4536 slice_errors=0 file_read_errors=0` |

Shipped: migration 243 (`body` column + `idx_code_symbols_body_trgm`), commit
`37f7deff9`, live on chassis `v1.0.1180`. **Candidate 3 (the false contract
comment at `:29-31`) is done in the same commit**, and candidate 4's ref-tracking
fact is now stated in that comment. Full trail in `doc_notes` under
`index_code_symbols` / `diagnose_code_lookup`.

### Defect A — a LIVE diagnosis was misled today, and here is the instance

Correlation `914dc844-7dad-4d5a-8d1b-a9c4296880c4`, bundle rendered
**2026-07-28 07:07:27**, diagnosing robot-hands.com 404 links. It asked, correctly,
whether `RepairPageLinks` is on the build path:

```
[code_request 1] kind=symbol query="RepairPageLinks"
  answered: 0 rows — searched the names of 4535 indexed symbols. The query was RUN
  and matched none; this is not an unanswered question.
[code_request 2] kind=ls query="datahelpers/"
  answered: 0 rows — no indexed path has that prefix, out of 4535 indexed symbols.
```

**Both exist.** `platform/orchestration/datahelpers/link_repair.go:139` defines
`func RepairPageLinks(html string, index PageURLIndex)`. They are absent only from
the indexed snapshot:

```
$ git cat-file -e e19aa5d:platform/orchestration/datahelpers/link_repair.go
  fatal: path ... exists on disk, but not in 'e19aa5d'
$ git rev-list --count e19aa5d..HEAD
  955
```

The banner above it read `(index freshness: refreshed 17h ago at commit e19aa5d)`.
**Seventeen hours sounds fresh. Nine hundred and fifty-five commits is not.**

**Distance over time — it only ever grows:** 667 at filing (06:00 07-27) → 946
(21:00 07-27) → **955** (07:10 07-28). A reindex ran successfully at 07:04 today
and the distance did not move, because the indexer fetches the last **pushed** tip
of the branch and `origin/086_experience_loop` has been at `e19aa5d10` throughout.
**This confirms the filed mechanism precisely: the refresh resets the freshness
clock without advancing the code.** The immediate cause is not the indexer — it is
that ~955 commits of work sit unpushed, and the index can only ever mirror what
was pushed.

### THE NEW THING, and it inverts the priority of the candidates

**Fixing defect B raised the cost of defect A.** The empty-answer rendering shipped
alongside the body fix deliberately replaced `"(no symbol matches in the index)"`
with `"answered: 0 rows — the query was RUN and matched none; this is not an
unanswered question."` That is the correct fix for the empty-vs-unanswered
confusion this bug is largely about — and it makes a **stale-index false negative
read as a stronger, more confident denial than it used to.** The old wording at
least kept the word "index" in the sentence.

So the two halves are not independent improvements that can land in either order:

> **Candidate 1/4 (freshness by COMMIT DISTANCE, not clock) is now a prerequisite
> for candidate 2/3 being safe, not a parallel nicety.** An honest "I searched and
> found nothing" is only honest if the thing searched is the thing being asked
> about. We have made the sentence more trustworthy without making the corpus more
> current, and trustworthiness applied to a stale corpus is exactly how a reviewer
> is led to a confident wrong answer.

This is a general shape worth naming: **two fixes that each increase honesty can
combine to increase harm, when one raises confidence in a signal the other has not
yet made correct.** It belongs in 016b §9.

### What candidate 1 actually needs, now that the ref question is concrete

The earlier note in this file — that the `code-index-refresh` task has no ref
parameter — still stands, and today's run shows the other half: the indexer's
`analyse_repo_local` step DOES receive an explicit `ref` when one is dispatched
manually (this run used `ref=086_experience_loop` and fetched `e19aa5d`, the pushed
tip). So the indexer already knows the ref it fetched; what it does not do is
**store** it. Recording `ref` alongside `commit_sha` on the row is the small change
that makes "N commits behind `<ref>`" computable at read time without giving the
shared chassis pod a GitHub token it deliberately does not hold.

---

## Measured cost, 2026-07-28 (bugs thread) — this defect defeated a diagnosis run on a DIFFERENT bug

Until now this file argued the false-zero risk. Here is what it actually cost, end to end.

**The setup.** `bugs_open/097` needed one question answered: of three mechanisms that claim
to cover in-body link integrity, which was on the build path when
`robot-hands.com/learning-center.html` deployed on 07-25 with six live 404s. I filed a
diagnosis run for exactly that (`914dc844-7dad-4d5a-8d1b-a9c4296880c4`).

**The outcome.** `UNVERIFIABLE` after all five iterations. And its own citation names why:

```
tier:  static
quote: "answered: 0 rows — searched the bodies and declarations of 4535 indexed symbols
        (4535 with bodies). The query was RUN and found nothing;
        this is not an unanswered question."
where: code_symbols index (code_request kind=content/symbol query="repairpagelinks")
```

**That is a false zero, and the wording is the problem.** `RepairPageLinks` exists and is
on the very path under investigation:

```
platform/orchestration/datahelpers/link_repair.go:139   func RepairPageLinks(...)
platform/orchestration/actions/validate_page_content.go:357
        cleanHTML, repairs = datahelpers.RepairPageLinks(cleanHTML, pageIndex)
```

The index cannot know that, because:

| fact | value |
|---|---|
| distinct `commit_sha` in `code_symbols` | **1** — `e19aa5d` |
| that commit's date | 2026-07-24 17:43 |
| commits behind HEAD | **970** (was 667 when this file was filed — drift is growing) |
| when `link_repair.go` was added | `43f254be5`, **2026-07-26** — two days AFTER the indexed commit |
| rows matching `RepairPageLinks` | **0** |
| rows from `link_repair.go` at all | **0** |

So the file did not exist at the commit the index holds. The query was answered honestly
against a snapshot two days stale, and the *phrasing* — "The query was RUN and found
nothing; this is not an unanswered question" — converts that into a positive claim of
absence. A diagnosing agent reading that reasonably concludes the symbol does not exist.

**Why this is worse than a stale cache.** The loop's own `next_scope` shows it then spent
an iteration on `"Retry with exact casing (prior searches used lowercase 'repairpagelinks')
to settle whether this symbol is indexed at all"` — i.e. it correctly suspected its own
evidence and burned budget chasing a casing theory for a file that was simply absent. The
index did not merely fail to help; it sent the investigation down a wrong path and then
ran out of iterations.

**What this adds to the case for fixing it:** the cost is no longer hypothetical. One
diagnosis run, five iterations, ~254KB of bundles, and `bugs_open/097` still has its
central question unanswered — because the evidence tier lied with confidence. Any
`content`/`symbol` answer from this index is currently worthless for anything added since
2026-07-24, which after 970 commits is most of the interesting surface.

---

## 2026-07-28 09:2x — defect A measured again, and it is now impairing a council run in flight

*Added by the gripper-dossier/consolidation thread. Evidence only; this case is owned.*

The drift is not slowing. Measured just now, with the commands, so the figure carries
its own provenance:

```
$ git ls-remote origin 086_experience_loop
e19aa5d108eddefb81196363908fe379ac86e447   refs/heads/086_experience_loop   # 2026-07-24

$ git rev-list --count <that sha>..HEAD
1003
```

```sql
SELECT DISTINCT commit_sha, max(updated_at) OVER () FROM code_symbols;
--  e19aa5d | 2026-07-28 07:05:29+00      <-- REINDEXED TODAY, still 1,003 commits behind
SELECT count(*) FROM code_symbols WHERE path LIKE 'platform/mailer/%';     -- 0
SELECT count(*) FROM code_symbols WHERE path LIKE 'platform/httpguard/%';  -- 0
SELECT count(*) FROM code_symbols WHERE path LIKE 'internal/tools-api/%';  -- 0
```

**The reindex at 07:05 today did not help, and cannot**, which is the part worth
restating: the indexer mirrors the last **pushed** tip, so re-running it re-indexes the
same 24-July tree. Only a push moves it. That makes this the rare defect whose fix is
not a code change at all.

**A live consequence, happening as this is written.** Council submission
`6db59c8b-829f-4e4f-8273-511e1714d6ce` is at `review_prior_art` right now, reviewing two
new packages (`platform/mailer`, `platform/httpguard`) and citing `internal/tools-api` as
its evidence. **All three are 0 rows in the index.** So the seat whose entire charter is
*"does this platform already have something that does this?"* is answering from a tree in
which none of the code under review, and none of the code it is being compared against,
exists. Whatever it returns, it cannot be evidence either way.

That is a sharper statement of the harm than "answers are stale": for a NEW package the
index cannot even be wrong usefully — it returns the same empty answer for
*"this is genuinely novel"* and for *"this shipped three days ago"*.

**Not actioned here, deliberately.** The fix is `git push`, which publishes 1,003 commits
from a shared branch that several sessions are working. That is an owner call, not a
side-effect of a bug-file update.
