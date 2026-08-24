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
Then read the agents themselves, never the migration's exit code — **but read the CONTENT, not
`updated_at`:**
```sql
-- ⚠ THIS IS THE WRONG CHECK, and it was in this runbook until 2026-08-24:
--     SELECT type, updated_at FROM agent_definitions WHERE type IN (...);
-- `[MEASURED 2026-08-24]` 199 of the 200 live agent rows share ONE updated_at value, to the
-- microsecond (2026-08-24 18:31:14.450827+00) — some bulk write touches the whole table, and it
-- took no snapshot and left no agent_definitions_backup row. So updated_at is DEGENERATE as a
-- per-lane change signal: it moves for rows nobody edited, and a lane checking "did my migration
-- land?" gets an answer that is true for everyone and evidence for no one.
-- Ask instead whether YOUR OWN needle is in the live text:
SELECT type,
       default_config#>>'{workflow,steps,plan_site,config,prompt_template}' LIKE '%19. MATCH STRUCTURE TO PROMISE%' AS my_rule_present,
       default_config#>>'{workflow,steps,load_components,config,query}'     LIKE '%component_expresses%'            AS my_menu_present
FROM agent_definitions
WHERE type = 'build-site-planner'
  AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;
```
**And on a shared agent, check the OTHER lane's needle in the same query** — two lanes edited
`build-site-planner` and `page-content-writer` today (591/595 here, 598/599 from the `bugs_open/380`
lane). Both sets survived `[VERIFIED 2026-08-24]`, because both used anchored `replace()` with
exact-count guards rather than a wholesale rewrite. A whole-prompt `jsonb_set` from either side
would have silently reverted the other, and **`updated_at` could not have told you**.
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

### Arm B was held, then RELEASED the same day — and the method that dated the blocker

`594`/`595` were briefly renamed `*_HOLD.sql` (so `SIDECAR_RE` kept the runner off them — a
documented "do not apply yet" does not survive another session's `--apply`). The release condition
was the `bugs_open/305` lane's `714789d7b` being live. **It was, and both are now applied and
recorded.**

**THE ONLY RELIABLE WAY TO DATE A DEPLOYED COMMIT — the binary's own record, which has no shelf
life** (`platform/buildcapability`, RFC_040):
```sql
SELECT git_commit FROM service_binary_capabilities
 WHERE service = 'agent-chassis' AND kind = 'build'
 ORDER BY last_seen_at DESC LIMIT 1;
-- 70fd163c24eae0c444bae7a425bb3d3c3096f7e4   [VERIFIED 2026-08-24]
```
Then ask git the ancestry question, **with a control in each direction**:
```bash
git merge-base --is-ancestor <your commit>  <that sha> && echo LIVE
git merge-base --is-ancestor <a later commit> <that sha> || echo "control fails, good"
git merge-base --is-ancestor <an old commit>  <that sha> && echo "control passes, good"
```

⚠ **Three methods that DO NOT work for this, all tried here first:**
1. **`grep -a <sha> /proc/1/exe` cannot answer it at all.** `buildinfo.GitCommit` is **one string,
   not an ancestry**, so a binary that certainly contains your commit reports it **absent**. Two
   lanes were burned by this (`bugs_open/215` on v1.0.1288, `bugs_open/299` on v1.0.1316).
2. **The 40-zeros "must be absent" control comes back PRESENT** — Go's internal digit table matches
   a run of zeros. A control that cannot fail is worse than none: it converts *"I did not check"*
   into *"I checked and it passed"*.
3. **A pod's `.status.startTime` dates the ROLL, not the IMAGE.** Ours started 15:39Z, an hour
   *after* the fix was committed at 14:39:30Z — which feels like evidence and is not; a tag can be
   built before a commit and rolled after it.
The startup `build provenance` line is honest but **scrolls** — gone from a full `kubectl logs`
within hours, and already past `--since-time` here.

### Post-apply checks for arm B

```sql
-- the LITERAL type, because DeclaredTypeSatisfied is default-TRUE and could never tell you
SELECT function,
       input_schema->'fields'->'content'->>'type'         AS declared,
       length(input_schema->'fields'->'content'->>'llm_guidance') AS guidance_len
FROM content_components WHERE is_active
  AND function IN ('generic-text-block','about-content','illustrated-text-block','article-body');
-- all four: html, 401-826 chars.  report-dossier stays type=text — deliberately excluded.

-- the second-order effect the two arms exist for
SELECT function, component_expresses(html_template, input_schema) FROM content_components
WHERE is_active AND function = 'generic-text-block';
-- {html-block,list,table}  (was {} before 594)   [VERIFIED 2026-08-24]
```
And on the prompt, all five in one row — `rule10_html`, `rich_text_gone`, `rule9_narrowed`,
**`md_ban_intact`** (304's markdown ban must have survived the replace) and `img_forbidden`:
all true `[VERIFIED 2026-08-24]`.

⚠ **Forward note from the 305 lane:** `</blockquote`, `</dd`, `</dt`, `</caption`, `</section` are
still absent from their sentence-boundary set, deliberately, because RULE 10 does not emit them.
**Adding any of those to the guidance or to RULE 10 obliges you to tell that lane** so they can
probe and fixture it — guessing at prefixes is how the `</th`/`</td` asymmetry arose.
