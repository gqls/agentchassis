# RUNBOOK — bugfix 220

## Read the live dispatcher mapping (the defect's home)
```sql
SELECT jsonb_pretty(default_config->'workflow'->'steps'->'process_item'->'config'->'sub_workflow'->'steps'->'call_handler')
FROM agent_definitions
WHERE type='build-dispatch-loop' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```
Gotcha: the call_handler step is nested under `process_item.config.sub_workflow.steps`,
not under `workflow.steps` directly. For agents where you don't know the nesting, use
the LATERAL jsonb_each form (see NOTES) — `jsonb_path_query` with `?(...)` filters
fails on this Postgres with "syntax error at or near (" for `@.keyvalue()`.

## Find every dispatcher that maps a given path (live, not seeds)
```sql
SELECT type FROM agent_definitions
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND default_config::text LIKE '%current_item.spec.page_name%';
```
Gotcha: this is a VALUE substring search, which works; the `jsonb::text LIKE
'%"k":"v"%'` KEY:VALUE form does NOT (jsonb renders a space after the colon). Induce a
non-zero control before trusting a zero (`input_data.spec.page_name` → 5 rows).

## The item-type census (who disagrees column vs spec)
Bug file § CONTRIB 2026-08-08 has the query shape and the measured answer
(only unbuilt_internal_link; re-run before shipping — the census moves daily).

## Apply + record the migration (after the Go is committed)
```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  < docs/agent_docs/sql_for_agents/340_unbuilt_link_dispatch_authoritative_page_id.sql
./scripts/migration/run-migrations.sh --record-only 340_unbuilt_link_dispatch_authoritative_page_id.sql \
  --note "<what you verified>"
```
Gotcha: record in the same motion — an idempotent file probes `ok` and reads as
pending for ever (mig 335 lesson). The `_ROLLBACK` sidecar is excluded by SIDECAR_RE.

## Prove the deploy (after the next fleet roll — owner runs make release)
```bash
kubectl -n ai-persona-system exec <chassis-pod> -- sh -c 'strings /app/agent-chassis | grep -c "authoritative_page_id"'
# expect >0 on EVERY replica; this change removes no string, so there is no negative
# control — corroborate with behaviour (below), not the tag.
```

## Behavioural acceptance
Find a site with a live link to a never-deployed page (`pages.deployed_at IS NULL` +
an href to its url in any rendered component), fire the one-shot improvement loop,
then assert: the minted `unbuilt_internal_link` item's `result` names the TARGET's
file, `pages.deployed_at` is now set on the target, and `result._verification.status`
= `verified`. The wrong outcome (pre-fix) deploys the CONTAINER's file and leaves the
target NULL.

## Re-check the three config legs are still live (do this before every acceptance run)
```sql
-- leg 1: the dispatcher mapping. NOTE THE DEPTH: input_mapping lives under
-- call_handler->'config', not on the step. Reading the step and asking for
-- ->'input_mapping' returns EMPTY, which looks exactly like a missing leg.
SELECT jsonb_pretty(default_config->'workflow'->'steps'->'process_item'->'config'
                    ->'sub_workflow'->'steps'->'call_handler'->'config'->'input_mapping')
FROM agent_definitions
WHERE type='build-dispatch-loop' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- expect: "page_id?": "current_item.page_id"

-- legs 2 and 3 in one read, without needing to know either step's nesting:
SELECT s.key AS step, jsonb_pretty(s.value->'config')
FROM agent_definitions a, LATERAL jsonb_each(a.default_config->'workflow'->'steps') s
WHERE a.type='page-build-handler' AND a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
  AND s.key IN ('load_page_record','save_sections');
-- expect: load_page_record.authoritative_page_id = input_data.page_id   (mig 340)
--         save_sections.page_name_field         = page_record.name      (mig 342)
```
Gotcha: always confirm the agent row count is 1 first — a zero-row result and a
wrong-path result are the same empty output.

## Re-grep the binary on the CURRENT pods (an earlier pod-grep expires at the next roll)
```bash
for p in $(kubectl -n ai-persona-system get pods -l app=agent-chassis \
             -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}'); do
  echo "--- $p"
  kubectl -n ai-persona-system exec "$p" -- sh -c \
    'strings /app/agent-chassis | grep -c "authoritative_page_id"; \
     strings /app/agent-chassis | grep -c "unbuilt_internal_link"; \
     strings /app/agent-chassis | grep -c "authoritative_page_identity_XYZ"' </dev/null
done
# expect per replica: 3, 7, 0 — the third is an INVENTED string, the negative
# control that proves the grep could have come out zero.
```
Gotcha: `kubectl exec` inside a loop eats the loop's stdin — `</dev/null` it. The
final `grep -c` returning 0 makes the whole `sh -c` exit 1; that is the control
working, not a failure.

## Fire the improvement loop at a chosen site (the shipped trigger CANNOT do this)
`scripts/initial_messages/060improvement_loop/076_improvement_loop_trigger.sh`
re-assigns SITE_ID/DOMAIN to robot-hands.com *after* parsing its own arguments, so
passing a site is silently ignored. Build the payload by hand instead — working
copy kept at `scratchpad/fire_improvement_loop_dartsonline.sh`. Two rules:
payload goes in the container COMMAND (never stdin — `kubectl run -i | kcat -P`
drops ~4 of 5 silently at exit 0), and the command ends `&& echo PUBLISH_OK`.
**No `PUBLISH_OK` in the output means nothing was published.**

## Verify a page at the served artefact — READ THE URL, NEVER BUILD IT
```sql
SELECT name, url, build_status, COALESCE(deployed_at::text,'NEVER') AS deployed
FROM pages WHERE site_id='<site>' AND name IN ('<container>','<target>');
```
then `curl` those `url` values verbatim. Gotcha, and it has now bitten twice in
this lane: the URL is **not** derivable from `pages.name` — `beginners` serves at
`/blog/beginners.html`, `guide-first-time-buyer` at
`/guides/first-time-buyer/index.html`. A guessed URL 404s, and a 404 is exactly
the signal that means "never deployed", so the wrong guess reads as a finding.
For a dead-link claim the honest triple is: container **200**, target **404**, and
`grep -c 'href="<href>"'` ≥ 1 in the container's served bytes.

## The residue census — what did the pre-verifier completions leave behind?
```sql
SELECT left(w.id::text,8) AS item, s.domain, w.updated_at AS completed,
       COALESCE(t.name,'(no row)') AS target, COALESCE(t.deployed_at::text,'NEVER') AS tgt_deployed,
       (w.result->'_verification'->>'status') AS verification
FROM site_work_items w
LEFT JOIN pages t ON t.id=w.page_id
LEFT JOIN sites s ON s.id=w.site_id
WHERE w.item_type='unbuilt_internal_link' AND w.status='complete'
ORDER BY w.updated_at;
```
Gotcha: ask by IDENTITY, not `count(*)`. The count says "6 complete" and hides
that three targets shipped days later by unrelated work, one had its link removed,
and only one is genuine residue. Then settle each candidate with the verifier's own
first disjunct rather than by eye —
`position('href="'||href||'"' in COALESCE(pc.rendered_html,'')) > 0` — because
"link still rendered?" is what separates residue from a legitimately resolved item.

## THE acceptance assertion — all four legs of one dispatch, in one row

> **CORRECTED 2026-08-09 (afternoon): the deploy leg's jsonb path in the previous
> version of this section was ONE LEVEL TOO SHALLOW, and the control below could
> not catch it.** The old path was
> `result->'response'->'deploy_result'->'rendered_page'->>'page_id'`. The real shape
> nests another `response` in between, so that path returns **empty on every row,
> converged or not**. Caught by dumping `jsonb_pretty(result)` on the first genuinely
> converged item (`69818add`), whose deploy was plainly successful in the JSON while
> the documented query printed a blank deploy leg.
> **Why the control did not catch it:** the control's expected value for that column
> was *empty*, and a wrong path and absent data both render as empty. The control
> agreed with the broken instrument. **A control only tests a column if the column is
> expected to be NON-EMPTY in the control case** — see the corrected control below.

```sql
SELECT left(w.id::text,8) AS item,
       w.result->'response'->'sections_saved'->>'page_name'              AS saved_page_name,
       left(w.result->'response'->'sections_saved'->>'page_id',8)        AS saved_page_id,
       -- NOTE the second ->'response': deploy_result wraps its payload one level deeper
       left(w.result->'response'->'deploy_result'->'response'->'rendered_page'->>'page_id',8) AS rendered_pid,
       length(w.result->'response'->'deploy_result'->'response'->'rendered_page'->>'html')    AS rendered_len,
       w.result->'response'->'deploy_result'->'response'->'deploy_result'->'response'->'data'->>'success'   AS deploy_ok,
       w.result->'response'->'deploy_result'->'response'->'deploy_result'->'response'->'data'->>'file_path' AS deployed_file,
       w.result->'_verification'->>'status'                              AS verif,
       left(w.result->'_verification'->>'detail',90)                     AS detail
FROM site_work_items w WHERE w.id::text LIKE '<item-prefix>%';
```
`response.sections_saved` and `response.deploy_result` are the two step outputs the
handler saga returns; `_verification` is stamped at completion. A CONVERGED item
reads: `saved_page_name` = the **target**, `saved_page_id` = `rendered_pid` = the
target's id, `rendered_len` non-zero, `deploy_ok` = `true`, `deployed_file` = the
target's `pages.url`, `verif` = `verified`, and the detail says *"target page … has
shipped"* (disjunct **a**).

**`rendered_len` and `deploy_ok` are the legs that separate a real deploy from an
honest skip, and the old query had neither.** A skipped deploy still emits a
`rendered_page` object carrying the correct target `page_id` — so `rendered_pid`
alone does NOT prove anything shipped. On `338deb27` the render is present, points at
grip-styles, and is **empty** (`rendered_len` 0, `deploy_ok` NULL).

**Two controls, and the second is the one that tests the deploy path.**
1. *Negative control* — run against `338deb27` (08-09 morning, pre-mig-342). It must
   print the known FAILURE: `saved_page_name` = `beginners` (the container),
   `saved_page_id` = `5009f5c8` (the container's id), `rendered_len` = 0, `deploy_ok`
   empty, `verif` = `verified` via disjunct (b) *"href … is no longer rendered"*.
2. *Positive control for the deploy leg* — run against `69818add` (08-09 14:24, the
   first genuine convergence). `deploy_ok` must read **`true`** and `deployed_file`
   **`/brands/index.html`**. **If this column comes out empty your path is a typo**,
   which is exactly the failure the single negative control could not see.
