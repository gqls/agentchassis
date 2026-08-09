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
