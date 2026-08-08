# RUNBOOK — bugs_open/221, meta-commentary pattern precision

Every command here had a gotcha attached. The gotcha is the reason the line is
in this file; do not strip it.

## 1. The census — locate candidate rows (a locator, NOT the check)

```sql
SELECT count(*) AS rows_scanned,
       count(*) FILTER (WHERE lower(pc.rendered_html) LIKE '%as an ai%')            AS as_an_ai,
       count(*) FILTER (WHERE lower(pc.rendered_html) LIKE '%as a language model%') AS lang_model,
       count(*) FILTER (WHERE lower(pc.rendered_html) LIKE '%i cannot generate%')   AS refusal,
       count(*) FILTER (WHERE lower(pc.rendered_html) LIKE '%calculator%')          AS positive_control
FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE p.status = 'active';
```

⚠ **`pages` has no `deleted_at`** — psql will hint at `deployed_at`, which is a
different question. Filter on `p.status='active'`. Filtering on `deployed_at`
instead is `bugs_open/185`'s trap (detectors that select deployed miss live pages).

⚠ **Keep the positive control in the same statement.** `calculator` = 281 is what
makes a 1 or a 0 in the other columns mean anything. A count with no control
cannot tell "clean fleet" from "broken predicate".

⚠ **This SQL is not the check.** It matches substrings; `checkMetaCommentary`
matches substrings *inside extracted prose blocks only*. The two disagree by
design after `bugs_open/219` — the SQL over-reports (it sees `<script>` bodies).
Use it to find rows to feed the Go, never as the verdict.

## 2. Run the REAL check over live bytes

Dump the candidate rows:

```bash
SP=<scratchpad>; mkdir -p $SP/rows
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -F'|' -c "
SELECT pc.id, s.domain, p.name, pc.slot_name
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE p.status='active' AND (lower(pc.rendered_html) LIKE '%as an ai%'
   OR lower(pc.rendered_html) LIKE '%input_schema%'
   OR lower(pc.rendered_html) LIKE '%on_missing%');" > $SP/rows/manifest.txt

while IFS='|' read -r id dom pname slot; do
  kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
    -At -c "SELECT rendered_html FROM page_components WHERE id='$id';" \
    < /dev/null > $SP/rows/$id.html            # <-- the </dev/null is load-bearing
  echo "$(wc -c < $SP/rows/$id.html) bytes  $dom/$pname/$slot  $id"
done < $SP/rows/manifest.txt
```

⚠⚠ **`kubectl exec -i` inside a `while read` loop DRAINS THE LOOP'S STDIN.**
Without `< /dev/null` on the `kubectl` line the loop processes exactly one row
and exits **0** with no error — indistinguishable from a one-row result set.
This cost a wrong "only one row exists" reading in this lane on 2026-08-08
(`WRONG_CALLS.md`). Always print a per-row line and **count it against the
manifest**; do not trust the loop to have run.

Then execute the real function in-package (scratch file, **not for commit**):

```go
// platform/orchestration/actions/zz_scratch_221_live_test.go
func TestScratch221LiveRows(t *testing.T) {
    files, _ := filepath.Glob(filepath.Join(os.Getenv("ROWS_DIR"), "*.html"))
    for _, f := range files {
        b, _ := os.ReadFile(f)
        issues := checkMetaCommentary(string(b))
        t.Logf("%s -> %d issue(s)", filepath.Base(f), len(issues))
        for _, i := range issues { t.Logf("    BLOCKER value=%q location=%q", i.Value, i.Location) }
    }
}
```

```bash
ROWS_DIR=$SP/rows go test ./platform/orchestration/actions/ -run TestScratch221LiveRows -v
```

⚠ **Delete the scratch file before committing.** It is `zz_`-prefixed so it
sorts last and is easy to spot in `git status`.

⚠ **The four `input_schema`/`on_missing` rows are the negative control** — they
must return **0** issues. If they ever return non-zero, `bugs_open/219`'s scope
fix has regressed and that is a bigger finding than 221.

## 3. Prove the check can still FAIL (the mutation)

A pass proves nothing unless the same harness can produce a failure. Induce the
disclosure the check exists for and require a blocker:

```bash
go test ./platform/orchestration/actions/ -run TestMetaCommentaryStillBlocksVisibleCopy -v
```

Then mutate: delete the first-person pattern entry from `metaCommentaryPatterns`
and re-run — that test **must** fail. A test suite that passes with the rule
removed is not testing the rule (`a-quiet-test-passes-when-the-rule-is-gone`).

## 4. Prove it live after the roll

Go changes are inert until an image is rebuilt and rolled. Verify at the **pod
that will run the action**, never at git and never at the tag:

```bash
kubectl exec -n ai-persona-system <pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "<a string the change ADDED>"'      # expect >=1
kubectl exec -n ai-persona-system <pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "<a string the change REMOVED>"'    # expect 0
```

⚠ **Both greps, same exec, every replica.** A positive control alone proves the
pipeline shipped *something*; the negative one proves it shipped *yours*
(`bugs_open/153`). ⚠ The fleet can be MIXED — the `agent-chassis` deployment may
have rolled while spawned agent pods still run the old image (loancalculator
lane, 2026-08-08). Grep the pod that will actually execute the step.

## 5. Has this check ever actually blocked a build?

```sql
SELECT iss->>'category' AS category, iss->>'value' AS value, count(*) AS hits,
       count(DISTINCT domain) AS domains, max(occurred_at)::date AS newest
FROM agent_error_log ael, jsonb_array_elements(ael.context->'issues') iss
WHERE ael.occurred_at > now() - interval '14 days'
GROUP BY 1,2 ORDER BY 3 DESC;
```

⚠ **Do NOT search `error_message`.** It carries no detail — the whole of it is
*"content validation failed: 1 blockers, 0 errors"* and *"see context.issues for
detail"*. Every spelling of `%meta_commentary%` against that column returns **0**,
fleet-wide, all history, and reads as "this check has never fired" when it had
blocked six builds that week. Cost this lane a near-miss on 2026-08-08.

⚠ **Do NOT use `context::text LIKE '%"category":"meta_commentary"%'`.** jsonb
renders a **space** after the colon, so that form matches nothing and hands you a
confident zero. Enumerate with `jsonb_array_elements` and read the keys.

⚠ **`agent_error_log` has no `created_at`** — the column is `occurred_at`. And
`pages` has no `deleted_at`. Both mistakes error loudly, which is the good case;
the two above do not.
