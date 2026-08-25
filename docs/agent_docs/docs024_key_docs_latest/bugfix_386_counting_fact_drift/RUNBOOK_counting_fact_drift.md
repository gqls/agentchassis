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

## §6 — running the scan properly (do NOT hand-roll text extraction)

The scanners read **extracted text blocks**, not markup and not `content_data`. Three surfaces that
look interchangeable and are not:

- `content_data` — staging. Holds the whole register snapshot including fact claim text and its
  writer guidance ("… always say so"). Scanning it over-reports, and presence here does **not** mean
  the value is rendered.
- `rendered_html` raw — chart markup. `evidence-chart` draws SVG, so a bare number regex returns
  viewBox bounds and coordinates (127, 320, 555, 600, 625, 700, 1200 …). Two of them coincided with
  genuine former fact values, so a hand-rolled scan can "confirm" a premise on chart geometry.
- what `cmd/claimscan` extracts — the same engine as the deploy gate and the post-deploy audit.
  **Use this one.**

```bash
S=<scratch>; SITE=<uuid>
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -c "
SELECT p.name || E'\t' || COALESCE(pc.slot_name,'') || E'\t' ||
       replace(encode(convert_to(pc.rendered_html,'UTF8'),'base64'), E'\n', '') ||
       E'\t' || COALESCE(p.page_type,'')
  FROM page_components pc JOIN pages p ON p.id = pc.page_id
 WHERE p.site_id = '$SITE' AND pc.rendered_html IS NOT NULL
   AND pc.rendered_html <> '' AND pc.locked_at IS NULL" > "$S/components.tsv"
# ASSERT the export before trusting it — it truncates at exit 0
wc -l < "$S/components.tsv"   # must equal the same predicate's count(*) in the DB
go run ./cmd/claimscan -evidence "$S/evidence.json" -components "$S/components.tsv"
```
**The 4th column is not optional in practice.** Without `page_type` every page reads UNKNOWN, prose
numbers are scanned on every page, and the tool reports editorial false positives the platform no
longer raises — so it disagrees with the gate it exists to predict.

**And check engine parity before quoting any result as "what the fleet does":**
`git diff <rolled provenance>..HEAD -- platform/orchestration/datahelpers/` must be empty. On
2026-08-25 it was 122 lines (`52958897f`, the 364 lane's component-grain change, committed and
unrolled), so a HEAD run predicts the post-roll gate, not the live one.

## §7 — does this counter REAP? (run before registering any floor)

A floor is only honest if the counter cannot fall below it. Establish which kind you have from the
archive, never from the fact's name:

```sql
WITH series AS (
  SELECT s.domain, f->>'id' AS fid, ss.created_at, (f->>'value')::numeric AS v,
         lag((f->>'value')::numeric) OVER (PARTITION BY s.domain, f->>'id' ORDER BY ss.created_at) AS prev
    FROM site_specs ss JOIN sites s ON s.id=ss.site_id,
         LATERAL jsonb_array_elements(ss.data->'facts') f
   WHERE ss.aspect='evidence_base' AND f ? 'value'
     AND (f->'source' ? 'sql' OR f->'source' ? 'query')
)
SELECT domain, fid,
       count(*) FILTER (WHERE prev IS NOT NULL) AS transitions,
       count(*) FILTER (WHERE prev IS NOT NULL AND v < prev) AS falls,
       min(v) AS lowest, max(v) AS highest
  FROM series GROUP BY 1,2
 HAVING count(*) FILTER (WHERE prev IS NOT NULL) > 0
 ORDER BY 4 DESC;
```
Read **three** columns, not one. `falls > 0` disqualifies a naive floor. `lowest = highest` means the
fact is **static** — it cannot drift, so it is not a 386 case at all and converting it to `gte` only
widens what the checker accepts. Only `falls = 0 AND lowest < highest` is a genuine rising counter,
and even then the floor must sit below `lowest`, not below today's value.

Measured 2026-08-25: 8 of 29 reap, 6 are static, and the largest faller ranges 1,625–90,790.
