# RUNBOOK — `bugs_open/381` inexpressive composition

Every command here was needed and had to be got right. Gotcha attached to each.
DB: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`

## 0. ⚠ Two traps that cost this lane real time

**(a) A PostgreSQL regex `\b` is BACKSPACE, not a word boundary.** `~* '<(ul|ol)\b'` matches
nothing and raises no error, so the first structure baseline read **0 across every slot** and
looked like a clean, dramatic finding. The word-boundary operators here are `\y` (either end),
`\m` (start), `\M` (end). Use a character class instead — `'<(ul|ol)[\s>]'` — which is what every
query below does. **The check: a positive-control row that MUST match.** `article-body` at 76% is
that control; a baseline reading 0 everywhere including the control is the instrument, not the
estate.

**(b) Long `jsonb_each` sweeps over `content_components` hang.** Two queries with an unguarded
`jsonb_each` over `input_schema` timed out at 2 minutes and killed the exec. Always
`SET statement_timeout='30s';` at the top of a psql heredoc and wrap the tool call in
`timeout 60`, and use `CROSS JOIN LATERAL jsonb_each(CASE WHEN jsonb_typeof(x)='object' THEN … ELSE '{}'::jsonb END)`
— 3 of 151 rows have a NULL/non-object `input_schema`.

## 1. The acceptance measure (Phase D) — fleet structure share

Pre-fix baseline `[MEASURED 2026-08-24]`: **741 pages / 29 sites; 260 with a list, 64 with a
table, 289 with `<strong>`; 327 (44%) with none of the three.**

```sql
WITH pg AS (
  SELECT p.id, p.site_id,
         bool_or(pc.rendered_html ~* '<(ul|ol)[\s>]')  AS has_list,
         bool_or(pc.rendered_html ~* '<table[\s>]')    AS has_table,
         bool_or(pc.rendered_html ~* '<strong[\s>]')   AS has_strong
  FROM pages p JOIN page_components pc ON pc.page_id = p.id
  WHERE pc.updated_at > now() - interval '30 days' AND pc.rendered_html IS NOT NULL
  GROUP BY 1,2)
SELECT count(*) AS pages, count(DISTINCT site_id) AS sites,
       count(*) FILTER (WHERE has_list)   AS with_list,
       count(*) FILTER (WHERE has_table)  AS with_table,
       count(*) FILTER (WHERE has_strong) AS with_strong,
       count(*) FILTER (WHERE NOT has_list AND NOT has_table AND NOT has_strong) AS none_of_three
FROM pg;
```
**Gotcha: measure only pages whose `page_components.updated_at` POSTDATES the apply.** A 30-day
window after the apply is mostly pre-fix pages and will read as "no change". Swap the interval
for `> TIMESTAMP '<apply time>'` when reading the result.

## 2. The controlled comparison (the finding that set the design)

Same declared type, same renderer, different `llm_guidance`:
```sql
SELECT cc.function, count(*) AS n30d,
       count(*) FILTER (WHERE pc.rendered_html ~* '<(ul|ol)[\s>]') AS w_list,
       count(*) FILTER (WHERE pc.rendered_html ~* '<h3[\s>]')      AS w_h3,
       count(*) FILTER (WHERE pc.rendered_html ~* '<strong[\s>]')  AS w_strong,
       count(*) FILTER (WHERE pc.rendered_html ~* '<table[\s>]')   AS w_table
FROM page_components pc JOIN content_components cc ON cc.id = pc.component_id
WHERE cc.function IN ('generic-text-block','article-body','about-content',
                      'illustrated-text-block','report-dossier','ported-prose','ported-page')
  AND pc.updated_at > now() - interval '30 days'
GROUP BY 1 ORDER BY 2 DESC;
-- 2026-08-24: article-body 153/116 lists (76%) | generic-text-block 173/12 (7%)
```

## 3. Reading what the writer actually sees for one field

```sql
SELECT function, jsonb_pretty(input_schema) FROM content_components
WHERE function = 'article-body' AND is_active;
```
**Gotcha: `llm_guidance` is the key the writer reads, NOT `description`.** The prompt renders
`llm_field_specs[].description`, which `plan_sections_action.go:2328` populates from
`fieldDef["llm_guidance"]`. A component's top-level `description` column goes to the PLANNER's
menu; the field's `llm_guidance` goes to the WRITER. Two different audiences, easily confused.
Fleet view of what the writer sees: `audit_writer_brief.py` (register CQ-025).

## 4. Post-apply verification

```sql
-- (a) the derived vocabulary, on known components
SELECT function, component_expresses(html_template, input_schema)
FROM content_components WHERE is_active
  AND function IN ('generic-text-block','info-card-grid','pricing','faq','ported-prose');
-- expect: {html-block,list,table} | {items} | {table} | {items} | {html-block,list,table}

-- (b) THE TYPE MUST READ BACK LITERALLY 'html' — the type checker cannot tell you this
SELECT function, input_schema->'fields'->'content'->>'type'
FROM content_components WHERE is_active
  AND function IN ('generic-text-block','about-content','illustrated-text-block','article-body');
```
**Gotcha on (b): `DeclaredTypeSatisfied` is default-TRUE** — only `array`/`list` are checked
(`datahelpers/content_type_violations.go:262`), so `hmtl`, `HTML` or `""` behave identically to
`html` and no downstream check could ever surface the typo. This read-back is the only thing that
would catch it.

## 5. Applying, and proving the live rows changed

Dry run first, scoped to this lane's files (`--apply` takes EVERY pending file otherwise):
```bash
# per session, and again after any roll
python3 scripts/apply-migrations.py --dir docs/agent_docs/sql_for_agents   # verify flag names first
```
Then read the agents themselves, never the migration's exit code:
```sql
SELECT type, updated_at FROM agent_definitions
WHERE type IN ('build-site-planner','site-planner','content-gap-planner','page-content-writer')
  AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;
```
**Gotcha: the id-scoped `UPDATE` is a silent no-op if a second active row exists** — which is why
every migration here opens with a `count(*) = 1` guard that RAISEs. And a verify block made of
bare `SELECT`s **cannot stop the COMMIT** (`ON_ERROR_STOP` ignores a non-empty result): use
`DO $$ … RAISE EXCEPTION $$`.

## 6. Reading a canary run's rendered prompts

```sql
SELECT collected_data#>>'{...}' FROM orchestration_states WHERE ...;
```
**Gotcha: `orchestration_states` is a ~24h rolling window on a sliding clock.** The garden-tools
build's rows are already gone. Read a canary within the day or not at all; the pre-fix `plan_site`
prompt text is preserved in the `loanzy_uk_example_site` lane's NOTES instead.

## 7. Council

```bash
DRY_RUN=1 ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>
```
Migrations have been in scope since 2026-08-19 (`bugs_open/314`). Budget ~30 minutes: the council
takes 2–5, the dispatch queues behind the fleet. Find the run by payload, never by the printed id:
```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
```
**Gotcha (from the `305` lane, 2026-08-24): a fleet roll KILLS an in-flight council run, and a
killed run is indistinguishable from a queued one.** The tell is the run's `updated_at` against
the pod's `.status.startTime`.

## 8. APPLIED — arm A live 2026-08-24, arm B HELD

**Applied by hand, NOT by the runner, and that was deliberate.** `run-migrations.sh` has **no
directory or file scope** — `--apply` takes EVERY pending file — and at the time `600_claims_audit_rotation.sql`
was pending from another lane. So:
```bash
for f in 591_component_expresses_and_build_site_planner_menu \
         592_site_planner_menu_capability \
         593_content_gap_planner_menu_capability; do
  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
    psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < docs/agent_docs/sql_for_agents/$f.sql
done
# then, because recording is a separate human act:
./scripts/migration/run-migrations.sh --record-only docs/agent_docs/sql_for_agents/<f>.sql --note "..."
```
**Gotcha: the runner's default dry run re-executes every pending file inside a doomed transaction**,
which took longer than a 300s timeout here. `--no-probe` skips that; or read the ledger directly
(`SELECT filename FROM schema_migrations ORDER BY filename DESC LIMIT 10;`) and diff it against the
non-sidecar files on disk.

### The post-apply checks that actually mean something

```sql
-- 1. the function discriminates — POSITIVE and NEGATIVE control in one breath
SELECT function, component_expresses(html_template, input_schema)
FROM content_components WHERE is_active AND function IN ('ported-prose','call-to-action');
-- ported-prose {html-block,list,table} | call-to-action {}   [VERIFIED 2026-08-24]

-- 2. each menu query RUNS, bound as the chassis binds it (a query that fails at runtime
--    fails the whole planner step, and nothing in the migration would have caught it)
DO $t$ DECLARE q text; n int; BEGIN
  SELECT default_config#>>'{workflow,steps,load_components,config,query}' INTO q
    FROM agent_definitions WHERE type='build-site-planner' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  EXECUTE 'SELECT count(*) FROM ('||q||') z' INTO n USING '<a site id>'::uuid;
  RAISE NOTICE 'rows: %', n;
END $t$;
-- build-site-planner 149 | content-gap-planner 149 | site-planner 151   [VERIFIED 2026-08-24]
```

**3. THE GATE MUST DISCRIMINATE IN BOTH DIRECTIONS — "the clause is present" is not a check.**
Strip the clause from the live query text and compare counts on two sites:
```
evidence-LESS site  (garden-tools): with gate 149, without 151  -> excludes exactly 2   ✓
evidence-BEARING site            : with gate 151, without 151  -> excludes nothing      ✓
```
`[VERIFIED 2026-08-24]` Both arms were run. A gate tested only on the site it should filter
cannot distinguish "correctly filtering" from "filtering everything".

### Arm B is HELD, and held mechanically

`594` and `595` are renamed `*_HOLD.sql` so `SIDECAR_RE` excludes them from the runner — because
the runner has no scope, a documented "do not apply yet" would not have survived another session's
`--apply`. **Release condition: the `bugs_open/305` lane's `714789d7b` (`</th`/`</tr` sentence
boundaries) must be live in the chassis.** How to check — and how not to:
```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor 714789d7b <the stamped commit> && echo LIVE
```
⚠ **Do NOT use `grep -a <sha> /proc/1/exe` for this.** Tried here: the 40-zeros control came back
**PRESENT** (it matches Go's internal digit table), so the probe cannot discriminate. And a pod's
`.status.startTime` dates the ROLL, not the IMAGE — ours started 15:39Z, minutes after the fix was
committed at 14:39:30Z, which proves nothing either way.
