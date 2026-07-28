# RUNBOOK — bugfix 102, the page-type gate on the claims layer

Commands that were hard to get right, with the gotcha attached. Fix them HERE.

---

## 1. The fleet measurement (the one that sizes this bug)

`cmd/claimscan` is the same scan engine as the gate and the audit. Run it against
**each opted-in site's own register** — a scan against an empty register answers a
different question (it surfaces every business-shaped number, not the ones the
platform would actually raise).

Which sites are armed:

```sql
SELECT s.domain,
       jsonb_array_length(COALESCE(ss.data->'facts','[]'::jsonb))         AS facts,
       jsonb_array_length(COALESCE(ss.data->'banned_claims','[]'::jsonb)) AS banned
FROM sites s
JOIN site_specs ss ON ss.site_id = s.id AND ss.aspect='evidence_base' AND ss.is_current
ORDER BY 1;
-- 2026-07-28: nine — ai-agent-orchestration.com, finetuning.uk, fundamentallyai.com,
-- gamesdesign.co.uk, leopardessconsulting.co.uk, oufe.com, relojistas.com,
-- robot-hands.com, vonc.com
```

Export per site. **The `page_type` prefix on field 1 is what makes the output
groupable; field 4 is what the scanner actually reads** (added by this fix):

```bash
for d in <domains>; do
  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
    psql -U clients_user -d clients_db -At -c \
    "SELECT ss.data::text FROM site_specs ss JOIN sites s ON s.id=ss.site_id
     WHERE s.domain='$d' AND ss.aspect='evidence_base' AND ss.is_current" > $d.evidence.json

  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
    psql -U clients_user -d clients_db -At -c \
    "SELECT COALESCE(p.page_type,'(none)') || '~' || p.name || E'\t'
         || COALESCE(pc.slot_name,'') || E'\t'
         || replace(encode(convert_to(pc.rendered_html,'UTF8'),'base64'), E'\n','')
         || E'\t' || COALESCE(p.page_type,'')
     FROM page_components pc
     JOIN pages p ON p.id = pc.page_id
     JOIN sites s ON s.id = p.site_id
     WHERE s.domain='$d' AND pc.rendered_html IS NOT NULL AND pc.rendered_html <> ''
       AND pc.locked_at IS NULL" > $d.tsv

  ./claimscan -evidence $d.evidence.json -components $d.tsv > $d.out 2>$d.err
done
```

### Three traps this run hit, all of which silently corrupt the numbers

1. **`kubectl exec` can truncate a large export and still exit 0.** One site came
   back 89 rows against 90. **Always diff the row counts between two exports of
   the same population before comparing their findings** — the error text
   (`read message: unexpected EOF`) went to stderr, which a redirect swallows.
2. **`grep -c` on the output files prints nothing for some sites**: the snippets
   carry en-dashes and grep decides the file is binary. Use
   `LC_ALL=C grep -ac -E '^(BANNED|NUMBER)'`.
3. **The glob `survey/*.out` also matches `survey/*.fixed.out`.** The "before"
   total came out as 187 = 124 + 63, i.e. before + after summed, and looked
   plausible. Count per site in an explicit list, never by glob.

Group findings by page type (field 1 is `page_type~page_name`):

```bash
cat *.out | grep -E '^(BANNED|NUMBER)' | awk '{split($2,a,"~"); print $1, a[1]}' \
  | sort | uniq -c | sort -rn
```

Exact suppressed set, before vs after (this is the positive control — it must be
a strict subset, and the "appeared" side must be empty):

```bash
comm -23 before.sorted after.sorted > suppressed.list   # 61 rows
comm -13 before.sorted after.sorted                     # must be EMPTY
```

## 2. Verifying the fix is live after a roll

The gate is Go, so it is inert until the image ships. **Pod-grep a symbol the
change introduced, with a positive control that proves the grep method works on
this binary at all.**

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis \
      -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec $POD -- sh -c \
  'strings /app/agent-chassis | grep -c "resolvePageType"'      # THE MARKER: 0 before, >=1 after
kubectl -n ai-persona-system exec $POD -- sh -c \
  'strings /app/agent-chassis | grep -c "scanComponentClaims"'  # POSITIVE CONTROL: 2 before and after
```

Go keeps function names in the binary, which is what makes this work — verified
on `agent-chassis-74dbd9c9f4-7p6d8`, 2026-07-28: `scanComponentClaims` → 2,
`businessClaimContextRe` → 1, `resolvePageType` → **0** (correctly, it had not
shipped). A zero on the marker WITH a non-zero on the control means "not rolled
yet"; zero on both means your grep is wrong, not that the fix is missing.

> **CORRECTED — this section first named `section-index` as the marker, and that
> was wrong twice.** Page-type string literals appear all over the Go source
> (`page_growth_budget.go`, `v3_site_actions.go`, `apply_gap_plan_action.go`,
> `populate_nav_tables_action.go` …), so a match proves nothing about this change;
> and `section-index` was then removed from the editorial set entirely, so the
> grep would have returned a confident, meaningless 1 on a binary without the fix.
> **A marker has to be a string that only YOUR change puts in the binary** — a
> new function name is usually the cleanest one available.

## 3. Which workflows can actually see a page type

```sql
SELECT ad.type, s.key, s.value->>'action', s.value->>'output_field'
FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
  AND s.value->>'action' IN ('validate_page_content','load_page_record')
ORDER BY ad.type, s.key;
```

2026-07-28: `page-build-handler` and `tool-recreation-handler` load
`page_record` before validating (so the type resolves); `content-reviewer` does
not (UNKNOWN — scans, as before); `report-builder` has `check_claims:false`
anyway.

## 4. Council submission

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh \
  <submission.json>
```

`SUBMISSION_CORR` for round 1 of this bug: **`de4a19f5-8f03-4e74-92cb-c23c10ab829d`**.
Find the run by payload, not by the printed id:

```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = 'de4a19f5-8f03-4e74-92cb-c23c10ab829d';

SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='de4a19f5-8f03-4e74-92cb-c23c10ab829d' AND kind='council_report'
ORDER BY created_at;
```
