# RUNBOOK — bug 161 residual / `bugs_open/456`

Every command here was got wrong at least once first. The gotcha is attached to each.

## Is a site's evidence register actually readable?

The question no dashboard answers. `regcheck` runs the **real** `ParseEvidenceBase`.

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At \
  -c "SELECT sp.data::text FROM site_specs sp JOIN sites s ON s.id=sp.site_id
      WHERE sp.aspect='evidence_base' AND sp.is_current AND s.domain='<domain>';" < /dev/null > eb.json
go run ./cmd/regcheck -evidence eb.json
go run ./cmd/regcheck -evidence eb.json -claim "<a sentence its bans should refuse>"
```

⚠ **`< /dev/null` is load-bearing.** `kubectl exec -i` forwards stdin; inside a `while read`
loop it eats the domain list and the census silently reports on ONE site with exit 0. Mine
printed `parsed OK: 1  FAILED: 0` over 27 domains and read as a clean fleet.

⚠ **Read the whole register into an array first and assert its size:**
`mapfile -t DOMS < domains.txt; echo "loaded: ${#DOMS[@]}"`. A summary line cannot show you the
rows it never saw.

## The fleet census (the measurement 456 rests on)

```bash
mapfile -t DOMS < domains.txt          # from: SELECT s.domain ... aspect='evidence_base' AND is_current
go build -o /tmp/regcheck ./cmd/regcheck
for d in "${DOMS[@]}"; do
  ... psql ... < /dev/null > eb.json
  /tmp/regcheck -evidence eb.json 2>&1 | head -1
done
```
Baseline **2026-09-03: 25 parse, 2 do not** (finetuning.uk, noted.co.uk). **That baseline is the
demand control** — a post-roll run showing 27 OK proves nothing unless you know it was 25.

## Find the malformed facts without running Go

```sql
SELECT s.domain, count(*) FILTER (WHERE jsonb_typeof(f->'value')='string')  AS string_valued,
                 count(*) FILTER (WHERE jsonb_typeof(f->'source')<>'object') AS bad_source
FROM site_specs sp JOIN sites s ON s.id=sp.site_id, jsonb_array_elements(sp.data->'facts') f
WHERE sp.aspect='evidence_base' AND sp.is_current GROUP BY 1 HAVING ... > 0;
```
Catches the commonest cause. **Not a substitute for the parser** — it knows one failure shape.

## Facts nothing re-proves (the 27)

```sql
SELECT s.domain, count(*)
FROM site_specs sp JOIN sites s ON s.id=sp.site_id, jsonb_array_elements(sp.data->'facts') f
WHERE sp.aspect='evidence_base' AND sp.is_current
  AND jsonb_typeof(f->'source')='object'
  AND NOT (f->'source' ? 'sql' OR f->'source' ? 'query' OR f->'source' ? 'citation'
        OR f->'source' ? 'attested_by' OR f->'source' ? 'artifact_check')
GROUP BY 1 ORDER BY 2 DESC;
```
After the roll this is a field: `facts_unverifiable` in the sweep result.

## Does a ban actually fire on live copy?

```sql
SELECT string_agg(regexp_replace(pc.rendered_html,'<[^>]+>',' ','g'),' ')
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE s.domain='<domain>' AND pc.build_status='deployed';
```
then match each `banned_claims` pattern with `(?i)` in Python. ⚠ **This is stored HTML with
tags stripped** — `<title>` and JSON-LD are OUTSIDE it. Say so wherever you quote the result.

## Building against HEAD while the tree is broken by a peer

The shared tree frequently will not compile because another session is mid-edit. **Do not
conclude your change is broken.**

```bash
scripts/verify-head-builds.sh --with <your file> [--with ...] --test \
   ./platform/orchestration/actions/... ./platform/orchestration/datahelpers/...
```
Builds committed HEAD **plus only your named files**, so a peer's dirty file cannot fail your
run. ⚠ **Pass targets** — the bare form runs the whole `test/` tree, which needs Kafka on
localhost and fails for reasons that are not yours. ⚠ If the failure names **your** file but a
symbol you never wrote, the file has a peer's uncommitted passenger: check
`git status`/`git diff` before committing it.

## Proving a test is worth having

```bash
cp <file> /tmp/bak; python3 - <<'PY'   # apply the mutation the test claims to catch
...
PY
go test ./<pkg>/ -run '<Tests>' -count=1     # MUST be red
cp /tmp/bak <file>
go test ./<pkg>/ -run '<Tests>' -count=1     # MUST be green
```
⚠ `gofmt -l <file>` **prints the path when the file is NOT formatted** and exits 0 either way.
Never follow it with an unconditional `echo "ok"`; label it `(empty above = clean)` or use
`test -z "$(gofmt -l <path>)"`.

## Council + diagnosis

- `DRY_RUN=1 097_TRIGGER_council_review_v1.sh <submission.json>` — free, and it catches real
  errors (mine: one edit naming two files; the server refuses whitespace in `.file`). **≤8 edits,
  one file each.**
- Verdict: `SELECT current_step, status FROM orchestration_states WHERE
  collected_data->'input_data'->>'fix_correlation_id' = '<corr>';` → `complete_approved`.
  Full objections: `SELECT body FROM diagnosis_artifacts WHERE kind='council_report' AND
  correlation_id='<corr>'` (column is **`body`**, not `content`).
  **Read an APPROVED verdict in full** — three of the four things I acted on came from a round
  that had already approved the change.
- `090` diagnosis: **one mechanism per symptom.** Mine bundled two and returned
  `UNVERIFIABLE — stopped: scope-not-narrowing`: a full run, no verdict on either half.
