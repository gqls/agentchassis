# RUNBOOK — bugs_open/386 counting-fact drift

Every query here was run 2026-08-25 and returned what the PLAN quotes. Gotchas attached.

DB prefix for all of them:
`kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At`

**Three traps that apply to every query below.** (1) Never wrap a `kubectl exec … psql` in a shell
`timeout` — it orphans the server backend, and a dead-client `COPY` sits in `ClientWrite` holding
locks; use `PGOPTIONS='-c statement_timeout=60000'` and `pg_terminate_backend` to clear one.
(2) An export can truncate with only a stderr line at exit 0 — count rows in the DB and assert the
export matches before trusting any scan built from it. (3) Record every count as `N as of <date>`;
a census does not go wrong, it goes stale by addition and reads as current for ever.

## §1 — the census: which facts are exposed at all

```sql
SELECT COALESCE(NULLIF(f->>'kind',''),'(none)') AS kind,
       COALESCE(NULLIF(f->>'tolerance',''),'(none/exact)') AS tolerance,
       CASE WHEN f->'source' ? 'sql' OR f->'source' ? 'query' THEN 'sql'
            WHEN f->'source' ? 'artifact' THEN 'artifact'
            WHEN f->'source' ? 'attested_by' THEN 'attested'
            ELSE 'other' END AS src,
       (f ? 'observations') AS has_obs,
       count(*) AS facts, count(DISTINCT ss.site_id) AS sites
  FROM site_specs ss, LATERAL jsonb_array_elements(ss.data->'facts') f
 WHERE ss.aspect='evidence_base' AND ss.is_current
 GROUP BY 1,2,3,4 ORDER BY 5 DESC;
```
**Gotcha:** `is_current` is load-bearing — without it you scan the whole archive and every count
roughly triples. `(none/exact)` and `exact` must be counted together: an absent `tolerance` IS
exact (`claims.go:1114` default arm).

## §2 — the blast radius: sql-sourced number facts with their tolerance and scoping

```sql
SELECT s.domain, f->>'id', COALESCE(NULLIF(f->>'tolerance',''),'exact') AS tol,
       CASE WHEN jsonb_array_length(COALESCE(f->'context_terms','[]'::jsonb))=0
            THEN 'NO context_terms -> degrades to EXACT' ELSE 'scoped' END AS ctx,
       f->>'value', f->>'verified_at'
  FROM site_specs ss JOIN sites s ON s.id=ss.site_id,
       LATERAL jsonb_array_elements(ss.data->'facts') f
 WHERE ss.aspect='evidence_base' AND ss.is_current
   AND (f->'source' ? 'sql' OR f->'source' ? 'query') AND f ? 'value'
 ORDER BY 1,2;
```
**Gotcha, and it is the one that matters:** a non-exact tolerance with **no** `context_terms`
silently degrades to exact (`claims.go:1098-1101`). So a `gte` fact with no terms is a **no-op at
the scan** while still changing refresh drift semantics — a config edit that looks like a fix and
is not. Always select the term count alongside the tolerance; the tolerance alone tells you
nothing.

## §3 — is the fact history reconstructible? (the backfill source)

```sql
SELECT count(*) AS superseded_rows, count(DISTINCT site_id) AS sites,
       min(created_at)::date AS oldest, max(created_at)::date AS newest
  FROM site_specs WHERE aspect='evidence_base' AND NOT is_current;

SELECT ss.created_at::date, f->>'id', f->>'value'
  FROM site_specs ss, LATERAL jsonb_array_elements(ss.data->'facts') f
 WHERE ss.aspect='evidence_base' AND ss.site_id='<site>'
   AND f->>'id' = '<fact id>'
 ORDER BY 1;
```
Answered `315 rows / 15 sites / back to 2026-07-16` as of 2026-08-25, and fundamentallyai's
`F9-feed-items-collected` reconstructs day by day.
**Gotchas:** the refresh supersedes rather than overwrites
(`refresh_evidence_base_action.go:1289-1318`), which is why this works at all — do not assume it.
Two rows can share a date (2026-08-16 has two refreshes), so de-duplicate before treating the
series as one-per-day. And `site_specs` has **no** retention job, unlike `orchestration_states`
which reaps at ~24h — so this is the durable source for anything older than a day.

## §4 — the accidental-support check before arming any `gte`

```sql
SELECT s.domain, f->>'id', f->>'value', COALESCE(NULLIF(f->>'tolerance',''),'exact'),
       COALESCE(f->'context_terms','[]'::jsonb)::text
  FROM site_specs ss JOIN sites s ON s.id=ss.site_id,
       LATERAL jsonb_array_elements(ss.data->'facts') f
 WHERE ss.aspect='evidence_base' AND ss.is_current AND s.domain='<domain>'
 ORDER BY 2;
```
Read the terms, not the tolerance. `context_terms` are matched with `strings.Contains` against a
±70-byte window (`claims.go:1086-1096`) — a **substring** test, so `"orchestration"` matches
"orchestrations", "orchestration layer" and any prose containing the word, and the fact then
vouches for every number below its value in that window. This is how a `gte` conversion switches
the check off for a whole vocabulary while reading as "no findings".

## §5 — the verification instrument

`go run ./cmd/claimscan -evidence <site.json> -components <tsv>`. Before believing any before/after
diff, assert `git diff <live build stamp> HEAD -- platform/orchestration/datahelpers/` is empty —
otherwise you are scanning with an engine the fleet is not running. Judge findings in ≥300 chars of
context, not the printed snippet: the snippet is ±60 bytes (`claimSnippet`, `claims.go:1131`) and a
citation for a figure routinely lives just outside that window.
