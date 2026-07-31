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
