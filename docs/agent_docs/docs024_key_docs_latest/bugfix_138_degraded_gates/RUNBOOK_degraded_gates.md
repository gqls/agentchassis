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

> **GOTCHA — it is `config.ai_service.max_tokens`, NOT `config.max_tokens`.**
> Querying the wrong depth does not error: it returns a confident `(unset→default)`
> for **every** seat, which reads exactly like "nobody has ever right-sized these"
> and is completely false. I nearly recorded editquality as unfixed on that output.
> A wrong-depth JSON path is a silent-zero trap, same family as
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
