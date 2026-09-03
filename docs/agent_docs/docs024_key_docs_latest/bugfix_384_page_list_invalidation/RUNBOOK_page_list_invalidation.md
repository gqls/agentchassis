# RUNBOOK — bugs_open/384 page-list consumer invalidation

Every command here was got wrong at least once before it was got right; the gotcha is attached.

## DB access
```bash
PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -X -q"
$PSQL <<'SQL'
\pset format aligned
...
SQL
```
- `-X -q` keeps psqlrc noise out. Heredoc + `<<<` in the same command clashes — write HTML to a scratch file first.
- The shell cwd DRIFTS after a `cd` in an earlier call — use absolute paths or `cd /home/ant/projects/agentchassis;` first.

## Schema facts that cost a query each
- `pages` has **no `slug`** — key on `pages.url` (per site: `/index.html` exists once per site, so ALWAYS join `sites`) or `pages.name`.
- `content_components.input_schema->'fields'` is an **OBJECT** (name → spec), not an array: `jsonb_each(...)`, never `jsonb_array_elements`.
- `agent_definitions` has no `name` column; `orchestration_states` keys the agent by `owner_agent_type`.

## Who consumes a query source (the consumer set — the seam's own query)
```sql
SELECT f.value->>'source' AS source, cc.name, f.key AS field,
       (SELECT count(*) FROM page_components pc WHERE pc.component_id=cc.id AND pc.build_status<>'removed') AS live_instances
FROM content_components cc, jsonb_each(coalesce(cc.input_schema->'fields','{}'::jsonb)) f
WHERE f.value->>'source' LIKE 'query.%' ORDER BY 1,2;
```
`[MEASURED 2026-08-24]` 43 fields / 25 components. Page-IMAGE sources (splice `pageImageJoins`): `pages_where_type:*`, `blog_posts`, `pages_under_section:*`. `section_index_for` does NOT.

## Pair census: (card asset, stored entry) — the disconfirmable measurement
```sql
WITH qf AS (SELECT cc.id component_id, cc.name component, f.key field, f.value->>'source' source
            FROM content_components cc, jsonb_each(coalesce(cc.input_schema->'fields','{}'::jsonb)) f
            WHERE f.value->>'source' LIKE 'query.%' AND f.value->>'type'='array'),
ent AS (SELECT p.site_id, p.url listing_url, qf.component, qf.source, pc.updated_at array_written_at,
               e.value->>'url' entry_url, coalesce(e.value->>'image','') entry_image
        FROM page_components pc JOIN qf ON qf.component_id=pc.component_id JOIN pages p ON p.id=pc.page_id
        CROSS JOIN LATERAL jsonb_array_elements(CASE WHEN jsonb_typeof(pc.content_data->qf.field)='array' THEN pc.content_data->qf.field ELSE '[]'::jsonb END) e
        WHERE pc.build_status<>'removed' AND p.status='active')
SELECT ent.source, ent.component, count(*) pairs_with_card,
       count(*) FILTER (WHERE position(ca.url in ent.entry_image)>0) current,
       count(*) FILTER (WHERE position(ca.url in ent.entry_image)=0) stale,
       count(*) FILTER (WHERE position(ca.url in ent.entry_image)=0 AND ent.array_written_at > ca.created_at) stale_though_written_after_card
FROM ent JOIN pages tp ON tp.url=ent.entry_url AND tp.site_id=ent.site_id
JOIN assets ca ON ca.site_id=tp.site_id AND ca.entity_type='page' AND ca.entity_id=tp.id AND ca.purpose='card' AND ca.status='active'
GROUP BY 1,2 ORDER BY 5 DESC;
```
- Gotcha: a first cut keyed `empty image` over ALL sources and showed news/directory arrays as "20/20 empty" — those entries have no `image` key at all. Join the card, don't count empties.
- Gotcha: verify the URL join on one site before believing a zero (leopardess `/blog.html` = genuinely no cards, checked entry-by-entry).

## Which re-render mode a producer gets
`page-rerender`'s `check_rerender_mode` condition (live, read from `agent_definitions`): reason ∈ {image_landed, section_data_resolved, cta_links_stale, template_changed, literal_markdown} → `rerender_page_sections` (re-resolves `query.*`); anything else → assemble (re-ships stored arrays). Dump it:
```bash
$PSQL -At -c "SELECT default_config::text FROM agent_definitions WHERE type='page-rerender' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL" > /path/agentdef.json
```

## Existing producers of `section_data_resolved` (precedent shapes)
```sql
SELECT DISTINCT ON (created_by) created_by, item_type, handler_agent, item_key, spec::text
FROM site_work_items WHERE spec->>'reason'='section_data_resolved' ORDER BY created_by, created_at DESC;
```
`render_news_section` → `page_rerender`/`page-rerender`, key `page_rerender_<name>_<site>_section_data_resolved` (THE shape to copy). `render_directory`/`reconcile_section_data` → `needs_page`/`page-build-handler` (LLM chain; not ours).

## Risk meters for the handler that receives the new items
```sql
-- escalation rate of reasoned re-renders (baseline 1/25 on 2026-08-24)
SELECT collected_data->'input_data'->'spec'->>'reason' reason, status, coalesce(collected_data->'rerender_sections'->>'escalated','(n/a)') escalated, count(*)
FROM orchestration_states WHERE owner_agent_type='page-rerender' AND created_at > now()-interval '14 days' GROUP BY 1,2,3;
-- owned pages FAIL save_sections on this path → excluded at the lookup
SELECT left(coalesce(error,'(null)'),90), count(*) FROM orchestration_states WHERE owner_agent_type='page-rerender' AND status='FAILED' AND created_at > now()-interval '14 days' GROUP BY 1;
```

## Ownership checks (re-run at EVERY phase boundary — before first code, before each commit, before a council round)
```bash
git log --oneline --since='90 minutes ago' -- bugs_open/384* platform/orchestration/actions/derive_card_asset_action.go platform/orchestration/actions/flag_page_image_rebuild_action.go platform/orchestration/actions/queryresolve
CUT=$(date -u -d '30 minutes ago' +%Y-%m-%dT%H:%M:%SZ); find ~/.claude/projects/-home-ant-projects-agentchassis/ -maxdepth 1 -name '*.jsonl' -newermt "$CUT" | grep -v <own-session-id> | xargs grep -c "PageListConsumerPages\|requestPageListReresolve\|derive_card_asset_action" | grep -v ':0$'
```
Peer sessions are addressable by name via ListAgents/SendMessage (`bugs_open/326`, `bugs_open/357`, `bugs_open/352`, `bugs_open/333 [cb419e]`). The filing lane (`dartsonline_traffic`) has no named session — coordinate through `bugs_open/384` itself.

## Phase 2 — enabling the sweep (ONLY after the roll)
1. Prove the binary registers the check (capability list, not `strings`):
   ```sql
   SELECT service, built_from, capabilities ? 'page_list_stale' FROM service_binary_capabilities ORDER BY recorded_at DESC LIMIT 3;  -- schema: \d first
   ```
   and `git merge-base --is-ancestor <phase-2 commit> <build provenance sha>` per CLAUDE.md.
2. Apply by hand: `docs/agent_docs/sql_for_agents/603_enable_page_list_stale_HOLD.sql` (snapshot_agent first, DO/RAISE verify inside). Rollback file beside it.
3. First sweep proof (demand control = the 4 sites with stale tool-cta entries, 14 pairs on 2026-08-24 — re-run the pair census first, it may have moved):
   ```sql
   SELECT s.domain, w.status, w.spec->'stale' FROM site_work_items w JOIN sites s ON s.id=w.site_id
    WHERE w.item_type='page_rerender' AND w.spec->>'check'='page_list_stale' ORDER BY w.created_at DESC;
   ```
   Disconfirming result: the completeness agent visits a stale site (`site_discovery_rotation`) and files nothing; or files against a page whose stored array matches a fresh resolve (re-run the comparison by hand before calling it wrong).
4. The per-run summary is in the discovery run's findings (`"summary":true` with stale/current/unknown) — `unknown > 0` on a site means a source did not resolve; that is not "current".

## Post-roll acceptance (the protocol that PASSED 2026-08-25)
```bash
./docs/agent_docs/docs024_key_docs_latest/bugfix_384_page_list_invalidation/scripts/induce_card_landing.sh dartsonline.com barrel-shapes
```
⚠ **That script's kcat route FAILS** — `asset-deployer` dispatched to `system.agent.generic.requests` lands on a pod with no S3 client (`derive_card_asset: storage client not available`). Use the **work-item route** instead, which is production's own path:
```sql
INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary, spec, priority,
                             handler_agent, status, created_by, item_key, batch_id)
VALUES ('<site>','image-build-handler','build','needs_content_image','low','<why>',
        '{"mode":"content_card","check":"<your_tag>","entity_type":"page","entity_id":"<page uuid>","page_name":"<page>","purpose":"card"}'::jsonb,
        65,'asset-deployer','triaged','<your_tag>','content_image:<page>', gen_random_uuid())
ON CONFLICT DO NOTHING;
```
Then wait for `build-dispatch-loop` — it runs **per site** (`load_items` takes `input_data.site_id`), roughly every 90 minutes per site, so check the rotation rather than assuming it stalled:
```sql
SELECT s.domain, count(*), max(o.created_at)::timestamp(0) FROM orchestration_states o
  LEFT JOIN sites s ON s.id::text = o.collected_data->'input_data'->>'site_id'
 WHERE o.owner_agent_type='build-dispatch-loop' AND o.created_at > now()-interval '3 hours' GROUP BY 1 ORDER BY 3 DESC;
```
The assertion (expect exactly N, none on an `owned` page):
```sql
SELECT p.name, COALESCE(p.rebuild_policy,'generic') AS policy, w.status, w.spec->>'reason', w.created_by, w.item_key
  FROM site_work_items w JOIN pages p ON p.id=w.page_id
 WHERE w.site_id='<site>' AND w.item_type='page_rerender' AND w.spec->>'cause'='card_landed:<page>' ORDER BY p.name;
```
And the causation leg. ⚠ **Do NOT require `pages.deployed_at` to advance** — corrected 2026-08-25 by running it: on a listing whose array is already current the re-resolve produces byte-identical HTML, the deploy is a no-op, and `deployed_at` legitimately does not move (measured: `index` re-rendered 4 of 4 sections, 0 carried, deploy step visited, `deployed_at` unchanged). The causation signals that DO discriminate are (a) the item row itself carries `spec.cause`, (b) the run carries it too, and (c) `page_components.updated_at` advances when the array is rewritten:
```sql
SELECT coalesce(collected_data->'input_data'->'spec'->>'page_name','?'), status,
       coalesce(collected_data->'rerender_sections'->>'escalated','n/a'), current_step
  FROM orchestration_states WHERE owner_agent_type='page-rerender'
   AND collected_data->'input_data'->'spec'->>'cause'='card_landed:<page>';
```
⚠ **The served page is not the measurement on dartsonline** — it serves 12/12 because the filer hand-repaired it on 2026-08-24. The ROWS are.

## Applying a HELD migration by hand (603, and the verify that fooled me)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -X -q -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/603_enable_page_list_stale_HOLD.sql
```
By hand, never `run-migrations.sh` — the runner takes EVERY pending file (LANDMINE), and a
`_HOLD` file is held precisely so a human applies it.

⚠ **`snapshot_agent()` returns the SOURCE row's id, not the snapshot's.** Verifying against
that id compares the live row WITH ITSELF and prints a clean result either way (45→45, nothing
added, nothing lost). The real pre-image is in a different table:
```sql
WITH live AS (SELECT default_config->'workflow'->'steps'->'run_checks'->'config'->'checks' c
                FROM agent_definitions WHERE type='completeness-discovery-agent'
                 AND is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL),
     snap AS (SELECT default_config->'workflow'->'steps'->'run_checks'->'config'->'checks' c,
                     snapshot_taken_at
                FROM agent_definitions_backup
               WHERE type='completeness-discovery-agent' AND snapshot_reason LIKE '603_%'
               ORDER BY snapshot_taken_at DESC LIMIT 1)
SELECT snap.snapshot_taken_at, jsonb_array_length(snap.c) before_n, jsonb_array_length(live.c) after_n,
       live.c @> snap.c AS every_pre_apply_name_survives,
       (SELECT jsonb_agg(x) FROM jsonb_array_elements(live.c) x WHERE NOT snap.c @> jsonb_build_array(x)) AS names_added,
       (SELECT jsonb_agg(x) FROM jsonb_array_elements(snap.c) x WHERE NOT live.c @> jsonb_build_array(x)) AS names_LOST
FROM live, snap;
```
Print `snapshot_taken_at` in the result — if the snapshot side is NULL or the two sides are the
same row, the diff is theatre.

## Reading the sweep's summary finding (what "it worked" actually looks like)

```sql
SELECT s.domain, o.created_at::timestamp(0), jsonb_pretty(f.value)
FROM orchestration_states o
LEFT JOIN sites s ON s.id::text = o.collected_data->'input_data'->>'site_id'
CROSS JOIN LATERAL jsonb_array_elements(
    CASE WHEN jsonb_typeof(o.collected_data->'run_checks'->'findings')='array'
         THEN o.collected_data->'run_checks'->'findings' ELSE '[]'::jsonb END) f
WHERE o.owner_agent_type='completeness-discovery-agent'
  AND f.value->>'check' = 'page_list_stale'
ORDER BY o.created_at DESC LIMIT 5;
```
- `stale=0` **with `current>0`** = it looked and everything was current. THE PASS.
- `stale=0, current=0, unknown=N` = it did not compare anything. On a site with a legitimately
  EMPTY listing this is correct and expected (lampenkap.com: one page, zero `tool` pages, so its
  `tool-list` array is empty and the resolve is classified UNKNOWN). It is NOT a pass, and at a
  glance it is indistinguishable from the blind case — read `consumer_pages` to tell whether the
  lookup itself ran.

## Fanning out a template change made by SQL (615's shape — reuse this)

Nothing re-renders on a `content_components` write; see the LANDMINE. Read the LIVE fixer query
first (`agent_definitions`, not migration 460 — the live row has since gained the owned-page
exclusion), then ADD the page-status filter it still lacks:

```sql
CREATE TEMP TABLE _targets AS
SELECT DISTINCT p.id AS page_id, p.name AS page_name, p.site_id, s.domain
  FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
  JOIN content_components cc ON cc.id=pc.component_id
 WHERE cc.name = '<component>' AND pc.build_status <> 'removed'
   AND p.status = 'active'                                    -- bugs_open/098; the fixer LACKS this
   AND COALESCE(p.rebuild_policy,'generic') <> 'owned';       -- save_sections refuses owned pages
```
item_key `page_rerender_<page>_<site>_template_changed` (the shared spelling — DB-level dedup),
KEEP a `NOT EXISTS` too (334 of 338 live `template_changed` items are keyless and the unique
index cannot see them), and SET `page_id` (only 4 of 272 carry one, and without it you cannot
check where the items landed).

**Verify at the artefact, never the migration:**
```sql
SELECT s.domain, p.name, (pc.rendered_html LIKE '%<your new class>%') AS has_it
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
JOIN content_components cc ON cc.id=pc.component_id AND cc.name='<component>'
ORDER BY 1,2;
```
plus an empty-attribute control (`rendered_html LIKE '%src=""%'` must be 0 if the render is gated).

## Filing the production card-derive (D1), and what it can and cannot do

Work-item route only (kcat lands on a pod with no S3 client — §7a). ⚠ **`derive_card_asset`
CROPS AN EXISTING HERO; it does not generate imagery.** A page with no hero of any kind
completes with `derived:false` and the reason `"no hero asset to derive from: no active page,
content, or site hero"` — a truthful completion that produces nothing. So check the ASSETS, not
the item status:
```sql
SELECT s.domain, count(*) FILTER (WHERE ca.id IS NOT NULL) AS with_card, count(*) AS pages
FROM pages p JOIN sites s ON s.id=p.site_id
LEFT JOIN assets ca ON ca.site_id=p.site_id AND ca.entity_type='page' AND ca.entity_id=p.id
     AND ca.purpose='card' AND ca.status='active'
WHERE p.page_type='tool' AND p.status='active' GROUP BY 1;
```

## The 090 diagnosis run on defect #1 (fired 2026-09-02, owner authorised)

**Question put to the loop:** a `page_rerender` item carrying `spec.reason=section_data_resolved`,
correctly specced and consumed on a healthy queue, reaches `complete` and deploys — and
`page_components.content_data` for the listing component is never rewritten.

```
INTAKE_CORRELATION=d4f745e6-3f79-42a8-8f71-bb611736912c   # the intake record
RUN_CORRELATION_ID=149ec925-ffb7-41eb-806a-1595b8ff2226   # <- artifacts are written under THIS
```
Env used: `RUNTIME_SITE=leopardessconsulting.co.uk`, `SITE_ID=4851f6fc-…`, `PAGES=blog`,
`SLUG=section_data_resolved_completes_without_rewriting_array`, `SEED_SCOPE=` the three files
(`rerender_page_sections_action.go:RerenderPageSectionsAction`,
`queryresolve.go:resolvePagesWhereType`, `page_list_reresolve.go`).

⚠ **It refused first, and `FORCE=1` was correct here — read this before copying the flag.** The
page-keyed coverage probe (`PAGES=blog`) matched ~30 open items and blocked. Two reasons none of
them was in-flight work on this mechanism:
1. The biggest block of hits **was this lane's own finding** — the nine
   `page_rerender_blog_…_section_data_resolved` rows at `unresolved`. They are born TERMINAL
   (see `bugs_open/384` UPDATE 09-02): nobody is working them and nobody can.
2. The rest are other sites' pages that happen to be named `blog` (`misdirected_cta` on
   ai-agent-orchestration, `required_fields_missing` from July, `voice:` from 2026-07-17).

**A property of the probe worth knowing:** its coverage clause is
`status NOT IN ('complete','cancelled','rejected')`, so a row parked in the terminal-but-unlisted
status `unresolved` reads to it as OPEN, IN-FLIGHT work. A born-dead backlog therefore *blocks
diagnosis of the very defect that created it*. `PAGES=` matches page NAME across ALL sites — narrow
it, or expect this.

Read the verdict (use the RUN correlation, not the intake one):
```sql
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'correlation_id' = '149ec925-ffb7-41eb-806a-1595b8ff2226';
SELECT body FROM doc_notes WHERE categories ? 'diagnosis' ORDER BY created_at DESC LIMIT 1;
```

### Outcome of that run, and the two seeding mistakes to avoid next time

**Verdict `UNVERIFIABLE` after 5 iterations (iteration-cap).** Useful anyway — its "still needed"
list is what cracked the case. Both gaps were MINE, in how I seeded it:

1. **Symbol-scoping the seed to the ENTRY function omits the callee that actually decides.** I
   passed `rerender_page_sections_action.go:RerenderPageSectionsAction`. The branch that chooses
   carry-vs-re-resolve lives in `rerenderFlatSections`, which the bundle then showed as an omitted
   body — so the loop could see `carryStoredSection` existed and could not see who calls it.
   **Seed the WHOLE FILE unless you have already read the call graph and know the deciding branch
   is inside the symbol you name.**
2. **The bundle's agent auto-gather does NOT include workflow steps.** It reported
   `agent_definitions[page-rerender]: root ai_service present=false` and nothing else, so the loop
   listed `check_rerender_mode`'s config as missing evidence — **a fact I already had in hand and
   had not put in the symptom.** If a live `agent_definitions` value is load-bearing, QUOTE IT IN
   THE SYMPTOM; do not assume the bundle fetches it.

**The check that actually settled it, and the one to reach for first next time:**
```sql
-- page_components.updated_at is NOT trigger-maintained; page_component_history IS.
-- The archive triggers fire on rendered_html change, or content_data change with html static,
-- so a MISSING row means the writer changed neither. history.component_id is the
-- page_components ROW id (not content_components.id) — the obvious join returns 0 rows.
SELECT created_at, source, op, application_name,
       length(rendered_html) AS html_len,
       jsonb_array_length(content_data->'articles') AS n_art
  FROM page_component_history
 WHERE page_id='<page uuid>' AND created_at > '<date>'
 ORDER BY created_at DESC;
```
`application_name` on each row NAMES THE WRITER (`action:rebuild_blog_listing`, `psql`, …). That is
how "which action actually maintains this array" became a query instead of an argument — and it is
the check that showed `rerender_page_sections` has written a listing array **zero times in 14 days**.

## Did a write REPAIR the listing? (the corrected census — added 2026-09-03 after the first one was published wrong)

Answers "of the writes that landed on a listing with a real image deficit, how many fixed it".
Full SQL: `scripts/census_repair_rate.sql` in this directory. Four gotchas, all paid for:

1. **JOIN THE CARD. Do not count `image=''`.** An entry whose target page has no active card is
   *correctly* blank — the resolver has nothing to project. Counting bare empties scored the
   resolver's correct behaviour as failure and produced a published "~37% success" where the true
   figure was 132/132. Same rows, same window: bare = 19 writes/5 repaired, card-joined = 8/6.
   The card must also have existed **before the write** (`ca.created_at < write.created_at`), or
   you are asking a write to have used data that did not exist yet.
2. **`PARTITION BY (page_id, slot_name)`, never `component_id`** — it is NULL on trigger rows, so
   partitioning by it collapses every slot on the page into one series and `LEAD` returns a
   different component's `content_data`. (Carried over from the 11:0x attempt; still true.)
3. **Trigger rows only** (`source='artefact_archive_trigger'`). Each write is recorded twice — an
   explicit `save_page_sections_overwrite` audit row and a trigger row ~42 ms later; only the
   trigger row carries `slot_name`, and counting both double-counts.
4. **`WITH … AS MATERIALIZED` at every stage.** PostgreSQL does not guarantee `AND` evaluation
   order, so an unmaterialised `jsonb_typeof(x)='array'` guard gets reordered behind
   `jsonb_array_length(x)` and the whole census dies with `cannot get array length of a scalar` —
   which reads like a data problem and is a planner one.

⚠ **The census sees only writes that MOVED BYTES.** The archive triggers fire on
`UPDATE OF rendered_html` when it changes, or `UPDATE OF content_data` when it changes and
`rendered_html` does not. A write that changed **neither** leaves no row — and a byte-identical
no-op is exactly what `bugs_open/454` produced. Every failure count from this census is a
**lower bound**, and must be quoted as one.

⚠ **You cannot attribute a write older than ~25 hours to a code path.** `orchestration_states`
held **25.0 hours** of history `[MEASURED 2026-09-03 12:4xZ]` (`SELECT min(created_at) FROM
orchestration_states` — re-measure, do not trust this figure). Beyond that the runs are gone and
the census gives you an OUTCOME only. To attribute, join each write to the last orchestration on
its page within 20 minutes:
```sql
LEFT JOIN LATERAL (
  SELECT o.owner_agent_type, o.collected_data->'input_data'->'spec'->>'reason' AS reason
    FROM orchestration_states o
   WHERE coalesce(o.collected_data->'input_data'->'spec'->>'page_id',
                  o.collected_data->'input_data'->>'page_id') = w.page_id::text
     AND o.created_at <= w.created_at AND o.created_at > w.created_at - interval '20 minutes'
   ORDER BY o.created_at DESC LIMIT 1) o ON true
```

## Reading which commit is actually running (the startup line had already scrolled)

```sql
SELECT service, git_commit, min(started_at), max(last_seen_at), count(*) AS pods
  FROM service_binary_capabilities WHERE kind='build' AND name='provenance'
 GROUP BY 1,2 ORDER BY 3;
```
then `git merge-base --is-ancestor <your commit> <git_commit>`. This has no shelf life, unlike
`kubectl logs | grep 'build provenance'` — on this fleet that line was out of range within half an
hour, and grepping the busy log for `provenance` returns **matches inside JSON payloads** that look
like hits and are not.
⚠ **GROUP the commits; do not take the newest row.** Ephemeral job pods spawn constantly, and the
427 lane hit a newest-first read returning a job pod still on the OLD commit. Group by commit and
read the standing Deployment's pods, or read the pod type that will actually run your work.

## Filing a page_rerender by hand (production's own row, not a kcat envelope)

Copy `insertPageRerenderItem` (`platform/orchestration/actions/create_rerender_items_action.go:240`)
— it is THE one INSERT for these: `pipeline='build'`, `priority=80`, `handler_agent='page-rerender'`,
`status='triaged'`, `created_by = source`. Give it a distinct `item_key` suffix so it cannot collide
with a real in-flight seam item (`idx_swi_dedup` is unique on `(site_id, item_key)` for
non-terminal rows only, so a `complete` predecessor does not block you).
⚠ **Check `bugs_open/450`'s guard before reading a null result as a 384 failure**: `save_page_sections`
refuses a page whose `rebuild_policy='owned'` OR that is a pending tool shell
(`genericBuildRefusal`, `owned_page_guard.go:175`). Confirm your target's `page_type` and
`rebuild_policy` first — a refusal and a failed re-resolve look identical at the array.
