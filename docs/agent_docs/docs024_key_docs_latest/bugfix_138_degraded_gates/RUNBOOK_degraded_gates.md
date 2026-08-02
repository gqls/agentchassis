# RUNBOOK — bugs_open/138, a truncated advisory review becomes a blocking one

Every command here was hard to get right at least once. The gotcha is attached to
the command, not filed separately.

---

## 1. Per-SEAT: which seats truncate, and how often (the file's headline table)

```sql
WITH r AS (
  SELECT rev->>'reviewer' AS reviewer, rev->>'verdict' AS verdict,
         COALESCE((rev->>'degraded')::boolean,false) AS degraded
  FROM diagnosis_artifacts d, LATERAL jsonb_array_elements(d.body::jsonb->'reviews') rev
  WHERE d.kind='council_report' AND d.created_at > now() - interval '14 days'
)
SELECT reviewer, count(*) AS reviews,
       count(*) FILTER (WHERE degraded) AS degraded,
       count(*) FILTER (WHERE degraded AND verdict='object') AS degraded_objections_that_gate
FROM r GROUP BY 1 HAVING count(*) FILTER (WHERE degraded) > 0 ORDER BY 3 DESC;
```

> **GOTCHA — this counts SEATS, not ROUNDS.** A round can carry more than one
> degraded gate, so this number is always ≥ the number of affected rounds and the
> two will not match. Both are correct; they answer different questions. Section 2
> is the round-level one. Reconciling them is how the "17 vs 10" apparent
> contradiction in the bug file resolves.
>
> **GOTCHA — the window ROLLS.** `now() - interval '14 days'` re-run six hours
> later is a different population. 17 became 18 the same day. Never quote this
> figure without its timestamp, and never treat a changed figure as a regression.

## 2. Per-ROUND: replay the labelling rule over history

This is the query that measured candidate 1's blast radius **before** submitting
it — it re-derives the new decision from the stored `reviews[]`, so it answers
"what would this change actually have done" rather than "what do I expect".

```sql
WITH rep AS (
  SELECT d.id, d.created_at, d.body::jsonb AS b
  FROM diagnosis_artifacts d
  WHERE d.kind='council_report' AND d.created_at > now() - interval '14 days'
), rv AS (
  SELECT rep.id, rep.created_at,
         rep.b->>'decision' AS decision, rep.b->>'decided_by' AS decided_by,
         r->>'reviewer' AS reviewer, r->>'verdict' AS verdict,
         COALESCE((r->>'degraded')::boolean,false) AS degraded,
         COALESCE(jsonb_array_length(CASE WHEN jsonb_typeof(r->'objections')='array'
                  THEN r->'objections' ELSE '[]'::jsonb END),0) AS n_obj,
         EXISTS (SELECT 1 FROM jsonb_array_elements(
                   CASE WHEN jsonb_typeof(r->'objections')='array'
                   THEN r->'objections' ELSE '[]'::jsonb END) o
                 WHERE lower(btrim(COALESCE(o->>'severity',''))) NOT IN ('low','medium')
                ) AS has_gating_obj
  FROM rep, LATERAL jsonb_array_elements(rep.b->'reviews') r
), g AS (
  SELECT *, (verdict='object' AND (degraded OR n_obj=0 OR has_gating_obj)) AS gates,
            (verdict='object' AND degraded AND NOT has_gating_obj)          AS trunc_only
  FROM rv
), per AS (
  SELECT id, min(created_at) AS at, min(decision) AS decision, min(decided_by) AS decided_by,
         count(*) FILTER (WHERE gates AND NOT trunc_only) AS merits,
         count(*) FILTER (WHERE gates AND trunc_only)     AS trunc,
         string_agg(DISTINCT reviewer, ',') FILTER (WHERE gates AND trunc_only)     AS trunc_seats,
         string_agg(DISTINCT reviewer, ',') FILTER (WHERE gates AND NOT trunc_only) AS merits_seats
  FROM g GROUP BY id
)
SELECT to_char(at,'MM-DD HH24:MI') AS at, decision, left(decided_by,34) AS decided_by_today,
       merits, trunc, COALESCE(trunc_seats,'-') AS truncated, COALESCE(merits_seats,'-') AS on_merits
FROM per WHERE trunc > 0 ORDER BY at;
```

> **GOTCHA — `jsonb_array_length` blows up on a non-array.** A salvaged review can
> carry `objections` as an OBJECT (that is `bugs_open/036`'s schema slip, and
> `salvageMistypedReview` deliberately keeps such a review). The
> `jsonb_typeof(...)='array'` guard is not defensive padding — without it the query
> errors on real rows.
>
> **GOTCHA — filter on `decided_by LIKE 'gating objection from %'` if you want
> rounds the gate actually DECIDED.** Two populations are excluded by that filter
> and both are legitimate: a round `rejected` by hard veto (the veto short-circuits
> before gate counting, so truncation never mattered) and pre-2026-07-22 rounds
> whose label reads `objection from X` without the leading `gating`. Dropping the
> filter changes 10 to 12 and neither number is wrong — say which you mean.

## 3. After the fix ships: the rate becomes one line

`gated_by_truncation` is persisted in `metadata` (jsonb) as well as `body` (text),
specifically so this needs no cast and no `jsonb_array_elements`:

```sql
SELECT date_trunc('day', created_at) AS day,
       count(*) AS reports,
       count(*) FILTER (WHERE metadata->>'gated_by_truncation' = 'true') AS truncation_gated
FROM diagnosis_artifacts
WHERE kind='council_report' AND created_at > now() - interval '14 days'
GROUP BY 1 ORDER BY 1;
```

> **A NULL here means "written before the fix", not "clean".** The field is emitted
> unconditionally, true or false, exactly so those two are distinguishable. Any
> alert built on this must treat NULL as unknown, never as zero.

## 4. Read a seat's ACTUAL token cap

```sql
SELECT s.key AS seat,
       max(CASE WHEN a.type='fix-proposer' THEN (s.value->'config'->'ai_service'->>'max_tokens')::int END) AS fix_proposer,
       max(CASE WHEN a.type='council-gate'  THEN (s.value->'config'->'ai_service'->>'max_tokens')::int END) AS council_gate
FROM agent_definitions a, LATERAL jsonb_each(a.default_config->'workflow'->'steps') s
WHERE a.type IN ('fix-proposer','council-gate')
  AND a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
  AND s.key LIKE 'review_%'
GROUP BY s.key ORDER BY 2 DESC NULLS LAST, 1;
```

> **GOTCHA — the two keys you need are at DIFFERENT depths, and neither wrong path
> errors.** Learn them as a pair or you will get one of them wrong:
>
> | what | path |
> |---|---|
> | token cap | `config` → **`ai_service`** → `max_tokens` |
> | prompt | `config` → `prompt_template` (a SIBLING of `ai_service`, not inside it) |
>
> Querying either at the wrong depth returns a confident uniform answer for **every**
> seat instead of failing: `(unset→default)` for the cap, which reads as "nobody has
> ever right-sized these", and `NULL` for the prompt, which reads as "these 51 seats
> have no prompts". I nearly recorded editquality as unfixed on the first (2026-07-29)
> and did briefly believe the second (2026-07-30) — **the second happened while
> knowing about the first**, because "watch the depth" does not tell you which keys
> are nested. That is why the table above gives both. Silent-zero family, same as
> `[[a-count-you-kept-is-not-a-census]]`.

## 5. Build & test when the working tree is broken by another session

`platform/orchestration/datahelpers/claims.go` was mid-edit by another thread
(undefined `negatedClaimMatch`), so `go test ./platform/orchestration/actions/`
could not compile — through no fault of the change under test.

```bash
SCRATCH=<scratchpad>/head138
rm -rf "$SCRATCH"; mkdir -p "$SCRATCH"
git archive HEAD | tar -x -C "$SCRATCH"
cp platform/orchestration/actions/diagnose_council_decide_action.go \
   platform/orchestration/actions/diagnose_council_test.go "$SCRATCH/platform/orchestration/actions/"
(cd "$SCRATCH" && go build ./platform/... && go test ./platform/orchestration/actions/)
```

> This is the same recipe as `[[shared-tree-wont-compile]]` and it doubles as the
> diagnosis: HEAD built clean here, which is what proved the breakage was another
> session's uncommitted WIP rather than something I had done or something already
> committed. **A red build in a shared tree is not evidence about your own change
> until you have separated the two.**

## 6. Verify the deploy at the POD — with controls, or it proves nothing

Raised as a council objection (`debug_historian`, medium, corr `919a05bf`) against a
submission that said only "inert until the next chassis image is rolled". It was
right: naming the check is part of the change.

**Capture the baseline BEFORE the roll.** After it, a 0 is unfalsifiable.

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec "$POD" -- sh -c '
echo "gated_by_truncation   : $(strings /app/agent-chassis | grep -c "gated_by_truncation")"
echo "gating TRUNCATED      : $(strings /app/agent-chassis | grep -c "gating TRUNCATED objection from")"
echo "all reviewers approve : $(strings /app/agent-chassis | grep -c "all reviewers approve")"
echo "gating objection from : $(strings /app/agent-chassis | grep -c "gating objection from")"
'
```

| marker | pre-roll (measured 2026-07-29, `agent-chassis-6fd7d88c4d-f6pgj`) | after |
|---|---|---|
| `gated_by_truncation` | **0** | **> 0** |
| `gating TRUNCATED objection from` | **0** | **> 0** |
| `all reviewers approve` | 1 | 1 — **positive control** |
| `gating objection from` | 1 | 1 — unchanged string |

> **The controls are not padding.** This change DELETES no string, so there is no
> delete-marker (the strongest kind, where a count must go 1→0 and cannot be faked
> by a bad grep). With additive markers only, a post-roll 0 is indistinguishable
> from: wrong pod, wrong binary path, `strings` absent, a typo in the pattern. The
> two controls must read non-zero **in the same command** or the whole result is
> void. See `[[date-the-build-when-a-change-adds-no-new-string]]` for the case
> where not even that is available.
>
> **A pod-grep proves the binary CONTAINS the change; it does not prove it RAN.**
> For that, wait for a real council round and check the artifact:
> `SELECT metadata->>'gated_by_truncation' FROM diagnosis_artifacts WHERE kind='council_report' ORDER BY created_at DESC LIMIT 1;`
> — NULL means a pre-fix binary wrote it, `false` means the new code ran and found
> no truncation gate. Those are different facts and the field is emitted
> unconditionally so they stay distinguishable.
>
> **And a green round proves nothing about the failing branch.** To prove the
> TRUNCATED wording, induce it: drop a scratch seat's `config.ai_service.max_tokens`
> to ~500 and run one round.

## 7. HEADROOM — the leading indicator, and the two ways to compute it wrongly

Truncation counts are the lagging indicator and they read ~0 once the caps have been
raised. What predicts the next truncation is how close a seat's output already gets
to its cap. Both traps below produce plausible numbers, not errors.

```sql
WITH live AS (
  SELECT a.type AS council, s.key AS seat,
         (s.value->'config'->'ai_service'->>'max_tokens')::int AS cap
  FROM agent_definitions a, LATERAL jsonb_each(a.default_config->'workflow'->'steps') s
  WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
    AND s.key LIKE 'review_%'
    AND (s.value->'config'->'ai_service'->>'max_tokens') IS NOT NULL
), pairs AS (
  SELECT seat, cap, string_agg(DISTINCT council, ',') AS councils FROM live GROUP BY 1,2
), calls AS (
  SELECT step_name AS seat, max_tokens AS cap, agent_type,
         output_tokens::numeric / max_tokens AS frac
  FROM llm_call_log
  WHERE created_at > now() - interval '14 days' AND step_name LIKE 'review_%'
    AND max_tokens > 0 AND output_tokens IS NOT NULL
)
SELECT p.seat, p.cap, count(c.frac) AS n,
       count(c.frac) FILTER (WHERE c.agent_type = ANY(string_to_array(p.councils,','))) AS n_holder,
       round(100*(percentile_cont(0.95) WITHIN GROUP (ORDER BY c.frac))::numeric,1) AS p95_pct,
       round(100*max(c.frac),1) AS peak_pct,
       count(*) FILTER (WHERE c.frac >= 1) AS truncated
FROM pairs p LEFT JOIN calls c ON c.seat = p.seat AND c.cap = p.cap
GROUP BY 1,2 ORDER BY peak_pct DESC NULLS LAST;
```

> **GOTCHA — THE DENOMINATOR CHANGES INSIDE THE WINDOW, and mixing caps invents a
> finding.** `llm_call_log.max_tokens` is per call. editquality went 8000→16000 on
> 07-28, guidelines and prior_art on 07-29. Compute the ratio PER ROW and join on
> `c.cap = p.cap` so only calls at the seat's *current* cap count. Take the p95 of a
> mixed population and you get "editquality is at 95% of a 16000 cap" — an artefact
> of 8000-cap rows, when the 16000 rows peak at 62.9%. I drafted that as a headline
> finding ("the raise created no headroom!") minutes after writing this warning in my
> own query. See `WRONG_CALLS.md`, 2026-07-30.
>
> **GOTCHA — `agent_type` CANNOT tell you which council made the call.** Every review
> call before 2026-07-26 14:54 logged `agent_type='generic'`; from 15:03 the same
> calls log `council-gate`; `fix-proposer` has **never** appeared. So a per-council
> split silently cuts its own history at that relabelling, and `WHERE
> agent_type='council-gate'` discards 1,798 rows. Key on **(seat, cap)** — the unit
> the risk belongs to — and use `n_holder` to say how much of the population is
> actually attributable to a council still holding that cap. It is exact for
> feature-designer and the experience councils, a LOWER BOUND for the fix lane.
>
> **Why peak AND p95.** Truncation is a tail event, so the maximum is the primary
> signal: `review_guardian` sits at p95 81.8% (unremarkable) with a peak of 99.2%
> over 278 calls — one review that came within 64 tokens of being cut. A p95-only
> rule rates it "ok" and flags two pairs whose evidence is 4 calls each.

## 8. The two live instruments (candidate 2, 2026-07-30)

```bash
# pull: the full table, any time, with its own honest limits in the header
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/104_REPORT_seat_token_pressure_v1.sh [days]
```

```sql
-- push: the CTE-only scheduled task (no LLM, no message, no credits) and its notes
SELECT name, interval_seconds, enabled, fire_message, last_triggered_at
FROM scheduled_tasks WHERE name='council-seat-token-pressure';

SELECT created_at, subject_key, body FROM doc_notes
WHERE categories ? 'seat-token-pressure' ORDER BY created_at DESC LIMIT 5;
```

> **The task is an EVENT, not a heartbeat.** `subject_key` holds an md5 of the flagged
> set, and the insert is skipped if a note with that key already exists (30-day
> look-back). A persisting condition is announced once; a new seat crossing, or one
> escalating near-miss→truncated, changes the digest and speaks again. Before trusting
> a quiet week, check `last_triggered_at` is recent — silence from a task that stopped
> running looks exactly like silence from a healthy fleet.
>
> **The thresholds live in the task's `pre_query` and NOWHERE else.** The report
> prints numbers and points at it. Two copies of a threshold is the drift class 099
> and 102 exist to fight.
>
> **Prove the no-op branch, not just the firing branch.** Substituting impossible
> thresholds into the pre_query must return **zero rows** — that is the path the
> scheduler takes 23 hours out of 24, and an aggregate over an empty set still returns
> one row, so the `n_flagged > 0` guard is doing real work.

## 9. Reading a seat's output-schema field order (candidate 4's survey)

```sql
WITH t AS (
  SELECT a.type AS council, s.key AS seat, s.value->'config'->>'prompt_template' AS p
  FROM agent_definitions a, LATERAL jsonb_each(a.default_config->'workflow'->'steps') s
  WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
    AND s.key LIKE 'review_%'
)
SELECT council, seat, array_to_string(ARRAY(
         SELECT m[1] FROM regexp_matches(substring(p from position('## Output' in p)),
                                         '"([a-z_]+)"\s*:', 'g') AS m), ',') AS key_order
FROM t ORDER BY 1,2;
```

> `## Output` is the only anchor present in all 51 templates, which is why the
> length-budget script inserts against it. Note the nested keys come out inline
> (`objections,edit,problem,severity`) — that is a feature: it shows the order INSIDE
> an objection, which is where the severity-loss theory lived until it was refuted.

## 10. Apply the length budget, and verify it the way that can actually fail

```bash
./scripts/apply-seat-length-budget.py            # dry run — lists targets + the block
./scripts/apply-seat-length-budget.py --apply    # snapshot each agent type, then write
./scripts/apply-seat-length-budget.py --verify   # live state only
```

> **The row count IS the check, not decoration.** The update uses
> `jsonb_set(..., create_if_missing := false)`, so a wrong path is a **silent no-op**
> rather than an error. The script asserts exactly one row per target and stops
> otherwise. Applied 2026-07-31: 7 of 7 at one row each.
>
> **Prove idempotency, because a re-run is how this gets used.** A second `--apply`
> must report "nothing to do" — the block is delimited by a start phrase plus
> `— end length budget —`, and a hand-authored block (which `review_architecture` has)
> lacks the sentinel and is deliberately **refused**, not overwritten. Both branches
> were exercised against real live seats before the first write.
>
> **Where snapshots actually go — the obvious check says they did not happen.**
> `snapshot_agent()` copies the live row into **`agent_definitions_backup`**, stamping
> `snapshot_taken_at`/`snapshot_reason`. It does NOT create an `is_snapshot` row in
> `agent_definitions`, even though every live-config query in this repo filters
> `COALESCE(is_snapshot,false)=false` and so implies it would. `created_at` is copied
> verbatim from the source, so ordering backups by `created_at` finds nothing recent
> either. And a snapshot is only a rollback if it predates the write — assert that:
>
> ```sql
> SELECT type, snapshot_taken_at, snapshot_reason,
>        (default_config #>> '{workflow,steps,review_guardian,config,prompt_template}'
>         LIKE '%end length budget%') AS has_change   -- MUST be false on the backup
> FROM agent_definitions_backup
> WHERE snapshot_taken_at > now() - interval '30 minutes' ORDER BY snapshot_taken_at;
> ```
>
> **Cutover time — take it from the row, never from your shell history.** Measured
> 2026-07-31: feature-designer `15:12:49`, fix-proposer `15:12:55`, council-gate
> `15:12:58` (`agent_definitions.updated_at`). An orchestration keeps the workflow
> definition it loaded **at spawn**, so a round already in flight carries the OLD
> prompt and is baseline, not signal. `scripts/council-adoption-report.sh` was wrong
> by 45 minutes once for exactly this reason and reclassified five pre-change rounds
> as evidence.
>
> **Then measure whether it did anything**, on rounds spawned after the cutover only:
>
> ```sql
> SELECT step_name, count(*), max(output_tokens),
>        round((100.0*max(output_tokens)/max(max_tokens))::numeric,1) AS peak_pct_cap
> FROM llm_call_log
> WHERE step_name IN ('review_guardian','review_improvement_guardian','review_debug_historian')
>   AND created_at > '2026-07-31 15:13:00+00'
> GROUP BY 1 ORDER BY 1;
> ```
>
> Pre-cutover peaks to beat: guardian **99.2%**, debug_historian **99.8%**,
> improvement_guardian **96.6%** (all of an 8000 cap). The architecture seat's
> precedent says to expect outputs to get *shorter*, not merely to stay under — if the
> peaks do not move, the block is being ignored and that is the finding.

## 11. Inducing the TRUNCATED branch — designed, pre-proven, BLOCKED on permission

The `true` branch of FIX-055 has never fired: **0 of 68 rounds** decided by the live code.
Not because it is unreachable — 3 reviews came back degraded and each was correctly
excluded (two had a high-severity objection survive the cut, so the round gated on merits;
one was degraded but approved). The reason it is now *rare* is that the length budget
removed the pressure: **0 degraded in the last 399 reviews.** So waiting is a weak plan —
we would be waiting to observe the thing we have been suppressing.

The induction below is fully prepared. `--apply`-equivalent steps were **refused twice by
the session permission classifier**, so it has NOT been run. Nothing is half-applied: the
seat still reads `cap=8000` and prompt md5 `94deeb7c39db83a6cd6263102ab2eb3e`, verified
after the refusal.

### Why this seat, and why the blast radius is near zero

`review_adoption_guardian` on **council-gate**: 38 calls in 14 days, the rarest seat, and
footprint-gated on `adoption` words — so another session's round only invokes it if their
edited paths match that footprint. Compare an always-on seat (`editquality`, `guardian`):
council-gate ran **15 rounds in 12 hours**, so crippling one of those would very likely
degrade someone else's round.

**And the window is shorter than it looks.** An orchestration keeps the workflow definition
it loaded **at spawn**, so the config only has to be crippled until *my round spawns* — not
until it completes. Restore the moment the row appears in `orchestration_states`; every
later round then loads the healthy config.

### The four conditions the branch needs (all four, or no label)

`degraded` **and** verdict `object` **and** no surviving high-severity objection **and** no
other seat gating on merits.

> **CORRECTED 2026-08-02 — the cap rationale below was reasoned on the wrong axis, and the
> mechanism it named is UNREACHABLE.** It read: *"120 guarantees the cut lands before any
> objection is written (`len(Objections)==0` → gates as truncated)"*, justified by the seat's
> p95 (~286). Two errors. (1) `repairTruncatedJSON` scans **backward for the last `}` or `]`
> and returns `""` when there is none** (`apply_adoption_plan_action.go:991`) — so a cut
> before the first objection closes discards the WHOLE review, yielding no verdict and no
> label. The `len(Objections)==0` route cannot be reached by truncating this seat. (2) The
> binding constraint is not p95 at all; it is **where the cut lands relative to the first
> objection's `severity` field**, because `severityGates` returns true for `"high"`, *unset*,
> **or anything unrecognised** (`diagnose_council_decide_action.go:691`) — so an objection
> whose severity was cut off reads as ungraded and **gates on merits**, taking the `false`
> branch. Caught by reading both functions before spending the round, not by running it.

**The window is bounded on BOTH sides.** Three landing zones:

| cut lands | repair yields | branch |
|---|---|---|
| before any closer (~<60 tok) | `""` — whole review discarded | nothing fires |
| after a complete **high** objection | one gating objection | `false` — merits |
| after a complete **low/medium** objection | non-gating objection + `Degraded` | **`true`** ✓ |

So the target is row 3, and **the injection must pin exactly one `severity: "low"` objection**
— pinning only `verdict: object` (as the first draft did) invites a guardian seat to grade it
`high`, which lands in row 2 and looks like a successful run while proving nothing.

Cap **120** survives the re-derivation, for the corrected reason: one complete objection is
~60–70 tokens `[ESTIMATED — not measured]`, and 120 is comfortably under this seat's observed
floor of **161** output tokens (55 calls/21d; daily floor 161–208, no downward drift since the
brevity rollout, so the floor is not an artefact of the pre-budget era). 500 was rejected
because it sits above the p50 and would truncate only sometimes.

**The row-3 fixture is also the FAITHFUL one**, which is the reassuring part: a low-severity
objection that gates only because the reply was cut is exactly what the code comment describes
— "would have been approved (or carried advisory objections only) had the seat been read in
full" (`diagnose_council_decide_action.go:728-731`).

### The recipe

```bash
SCR=<scratchpad>
# 0. capture EXACTLY, and build+prove the restore BEFORE breaking anything
#    (run the restore while it is still a no-op; the prompt md5 must be unchanged after)
# 1. snapshot_agent('council-gate', 'pre-update: INDUCED FAULT …')
# 2. jsonb_set cap -> 120 AND append the induced-fault paragraph to the prompt
# 3. fire 097 with a submission whose path matches the adoption footprint
# 4. poll: SELECT 1 FROM orchestration_states WHERE collected_data->'input_data'->>'fix_correlation_id' = '<corr>';
#    the moment it exists -> RUN THE RESTORE
# 5. verify restore: cap=8000 AND prompt md5 back to 94deeb7c39db83a6cd6263102ab2eb3e
# 6. read the verdict:
#    SELECT metadata->>'gated_by_truncation', body::jsonb->>'decided_by'
#      FROM diagnosis_artifacts WHERE kind='council_report' AND correlation_id='<corr>';
```

> **Build and PROVE the restore first.** Writing the undo before the do, and running it
> while it is still a no-op, is what turns "I can revert this" from a claim into a checked
> fact. Done here: the no-op restore returned `cap=8000` and left the prompt md5 identical.
>
> **The prompt injection is deliberate and must be disclosed.** The seat is told to emit
> verdict `object` **with exactly one objection graded `severity: "low"`** for that round.
> Without the verdict pin it is a coin flip (~32% object fleet-wide); without the SEVERITY
> pin the round lands in row 2 of the table above and silently proves nothing. It makes the
> induced review NOT a genuine adoption objection — so the correlation must be recorded
> wherever the artifact is read, or a future reader will mistake a test fixture for a real
> verdict.
>
> **The submission must be a REAL change.** A fabricated plan spends ~10 seats' credits
> reviewing fiction and pollutes `fix_plan` artifacts. The prepared one expands
> `repairTruncatedJSON`'s one-line comment in `apply_adoption_plan_action.go` to name its
> cross-pipeline consumer (`salvageTruncatedReview`) and the `output_tokens IS NULL`
> logging signal — comment-only, genuinely useful, and its path matches `adoption`.
>
> **What it would prove, and what it would not.** It proves the string selection and that
> the TRUNCATED wording survives persistence. It does NOT prove more than the unit tests
> about the decision rule itself — the persistence path is already demonstrated 68 times by
> the `false` branch, which writes the same two fields.

### §11 OUTCOME (2026-08-02) — fired once, true branch NOT reached; read this before designing round 3

Corr `4d7363d7-63dc-4b01-a912-ac0bb73c3031` → `rejected`, `decided_by="veto from
editquality"`, `gated_by_truncation=false`; adoption review counted `unreadable`.
Full evidence in NOTES 2026-08-02 and WRONG_CALLS (two entries, same date). The two
blockers any round 3 must clear:

1. **Size the cap in VISIBLE-TEXT terms, not output_tokens.** `claude-sonnet-5` thinks
   by default; the chassis neither requests nor records it (`thinking_*` columns are
   ALWAYS null — not evidence), and `output_tokens` counts thinking+text. At cap 120,
   ~108 tokens went to thinking and 42 CHARS of JSON came out — under the discard
   floor. Rehearse the exact crippled call offline first (`prompt_rendered` is in
   `llm_call_log`), or size generously: cap ~400 with the fixture demanding ≥800 words
   of notes keeps truncation certain while giving thinking room to finish and the
   objection room to close.
2. **The submission must be one that EVERY non-crippled seat approves.** editquality
   vetoes comment-only plans by explicit rule ("audits/comments are not edits"), and a
   veto — or any merits gate — short-circuits `decideCouncil` before the truncation
   rule runs. The true branch is a conjunction: near-perfect submission AND crippled
   seat. That conjunction is why 68 organic rounds never showed it, and it is the real
   cost of witnessing the label: a genuine, approvable CODE change must ride along.

What the failed round DID witness live (previously unit-tested only): the
tolerate+retry machinery on a truncated seat (bugs_open/119), the salvage-failure arm
(`repairTruncatedJSON` → `""` → honest `unreadable`, no fabricated verdict, round
decided on the remaining 11 seats), the frozen-plan mechanics, and the veto
short-circuit. Collateral: none — both 120-cap calls carry the induction round's
orchestration id; the window round `0802-1034` shows no adoption reviewer and no
unreadables.
