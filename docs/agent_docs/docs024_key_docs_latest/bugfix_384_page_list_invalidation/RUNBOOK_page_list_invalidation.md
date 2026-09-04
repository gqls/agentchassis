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

---

## Running the census WITHOUT getting a truncated answer (added 2026-09-04)

`scripts/census_repair_rate.sql` emits one row per qualifying write — 163 of them on 09-03. Piped
through `kubectl exec`, that is enough to hit the known truncation trap (`LANDMINES.md`,
*"`kubectl exec` truncates a large export mid-stream, and the short file looks complete"*): the run
exits **1**, the file ends **mid-row** with the kubectl error text spliced onto the last data line,
and the 81 surviving rows are individually perfect. Bucketing that with `awk` yields a plausible,
wrong census. It happened here on the first attempt.

**Do not retry-until-it-matches for this query — aggregate SERVER-SIDE so the result is 4 rows.**
Keep the script's CTE chain verbatim (it is the definition of "a write over a real deficit") and
replace only the final `SELECT`:

```sh
python3 - <<'PY' > /tmp/census4.sql
src = open('docs/agent_docs/docs024_key_docs_latest/bugfix_384_page_list_invalidation/scripts/census_repair_rate.sql').read()
head, sep, _ = src.partition('SELECT s.domain, p.name AS page')
assert sep, 'anchor not found — the script final SELECT was reworded'
head += """SELECT CASE
         WHEN sc.created_at < '2026-09-02 11:27:53+00' THEN '1 pre-regression'
         WHEN sc.created_at < '2026-09-03 12:05:34+00' THEN '2 DURING 454'
         WHEN sc.created_at < '2026-09-03 22:06:39+00' THEN '3 post-fix d0252fd4'
         ELSE '4 post-fix 239ab3626' END AS era,
       count(*) AS writes_over_deficit,
       count(*) FILTER (WHERE sc.produced IS NOT NULL AND sc.post_deficit = 0) AS repaired,
       count(*) FILTER (WHERE sc.produced IS NOT NULL AND sc.post_deficit > 0) AS left_blank,
       count(*) FILTER (WHERE sc.produced IS NULL) AS unknown,
       min(sc.created_at)::timestamp(0) AS first_write,
       max(sc.created_at)::timestamp(0) AS last_write
  FROM scored sc WHERE sc.pre_deficit > 0 GROUP BY 1 ORDER BY 1;
"""
open('/dev/stdout','w').write(head)
PY
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db < /tmp/census4.sql
```

⚠ **`assert sep` earns its line.** The anchor is a literal from the script; if someone rewords that
`SELECT`, a silent `partition` miss would run the CTEs and emit nothing, which reads as "no writes".

⚠ **ALWAYS read `last_write`, not just the ratio.** A clean era over a period with no writes is not
evidence — on 2026-09-04 era 4 read 7/7/0 with `last_write 08:12:16`, i.e. nothing in the preceding
3.6 hours. The zero needs a demand control.

⚠ **The era-1 boundary is a rolling 10-day window** (`now() - interval '10 days'`), so its COUNT
shrinks monotonically for ever (132 → 131 → 130 over three days). The ratio is the claim.

## Is a page in the "cannot render, cannot escalate" hole? (added 2026-09-04)

The conjunction behind `bugs_open/384`'s last generic residual: the render gate refuses the page
(a component has a required `source:"llm"` field absent from stored `content_data`) **and** the
escalation that would fix it is suppressed (`pageSectionsSatisfiable` false). Reproduce all three
sources `declaredPageSections` reads — `site_specs` current `site_plan`, `pages.sections`, and
`site_plan_pages` membership — or the answer is wrong in the safe-looking direction.

⚠ **Two traps, both hit while writing this.**
- `site_specs` has a **`data`** column, not `spec_data`.
- `jsonb_array_length` over `data->'pages'` and `pg->'sections'` **dies with
  `cannot get array length of a scalar`** unless each is behind its own `jsonb_typeof(...)='array'`
  CASE. This is the AND-ordering trap already in this runbook, one level deeper: materialising the
  outer CTE is not enough, because the guard and the length sit in the same expression.

```sql
WITH req AS MATERIALIZED (             -- required llm fields per component
  SELECT cc.id AS component_id, f.key AS field
    FROM content_components cc
    CROSS JOIN LATERAL jsonb_each(COALESCE(cc.input_schema->'fields','{}'::jsonb)) f
   WHERE f.value->>'source' = 'llm' AND (f.value->>'required')::boolean IS TRUE),
refused AS MATERIALIZED (              -- page_components missing at least one of them
  SELECT DISTINCT pc.page_id, pc.slot_name FROM page_components pc
    JOIN req ON req.component_id = pc.component_id
   WHERE pc.build_status <> 'removed' AND COALESCE(pc.content_data->>req.field,'') = ''),
spec_sections AS MATERIALIZED (
  SELECT ss.site_id, pg->>'name' AS page_name,
         jsonb_array_length(CASE WHEN jsonb_typeof(pg->'sections')='array'
                                 THEN pg->'sections' ELSE '[]'::jsonb END) AS n
    FROM (SELECT site_id, CASE WHEN jsonb_typeof(data->'pages')='array'
                               THEN data->'pages' ELSE '[]'::jsonb END AS pages
            FROM site_specs WHERE aspect='site_plan' AND is_current) ss
    CROSS JOIN LATERAL jsonb_array_elements(ss.pages) pg),
plan_member AS MATERIALIZED (
  SELECT sp.site_id, spp.name AS page_name FROM site_plan_pages spp
    JOIN site_plans sp ON sp.id = spp.plan_id WHERE sp.is_current)
SELECT s.domain, p.url, r.slot_name
  FROM refused r JOIN pages p ON p.id=r.page_id JOIN sites s ON s.id=p.site_id
 WHERE p.status='active'
   AND COALESCE((SELECT max(n) FROM spec_sections x
                  WHERE x.site_id=p.site_id AND x.page_name=p.name),0) = 0
   AND COALESCE(jsonb_array_length(CASE WHEN jsonb_typeof(p.sections)='array'
                                        THEN p.sections END),0) = 0
   AND NOT EXISTS (SELECT 1 FROM plan_member m
                    WHERE m.site_id=p.site_id AND m.page_name=p.name)
 ORDER BY 1,2;
```

`[MEASURED 2026-09-04 12:0xZ]` → **4 slots / 3 pages / 3 sites**, one of them a 384 consumer.
Drop the last three predicates to get the *refused but escalatable* population (**64 slots /
60 pages / 8 sites** the same day) — a different question, and not this lane's.

## Verifying a card-image repair AT THE SERVED PAGE (corrected 2026-09-04)

⚠ **`grep -c 'src=""'` IS NOT A SUFFICIENT CHECK and cannot fail on half our templates.** A blank
image behind `{{if .image}}` renders **no element at all**. `[MEASURED 2026-09-04]` 8 of the 15
components that render `.image` guard it that way. Count containers against images:

```sh
curl -s "https://<domain>/<page>.html" > /tmp/p.html
python3 - <<'PY'
import re
h=open('/tmp/p.html').read()
cards=re.findall(r'<article[\s\S]*?</article>', h)
bad=[c for c in cards if not re.search(r'<img', c)]
print('articles:', len(cards), 'without img:', len(bad))
for c in bad: print('  MISSING:', re.findall(r'href="([^"]*)"', c)[:1])
PY
```

Then require the count to agree with the stored row —
`SELECT jsonb_array_length(content_data->'articles') FROM page_components WHERE …` — so the served
page and the database have to corroborate each other. Keep `src=""` as a SECOND assertion only.

### ⚠ CORRECTED 2026-09-04 12:3xZ — the hole query above UNDERCOUNTS; use this one

The `refused` CTE above expresses only the gate's **second** branch. The gate refuses **two** ways
(`rerender_page_sections_action.go:427-431`) and applies an exemption before either:

- **exemption** — `isSelfContainedSection`: `component_level='tool'` AND **empty `input_schema`** →
  `continue`, never tested. **A section that is never tested can never be refused.**
- **(a)** `len(s.contentData) == 0` → *"no stored content_data"*. **Schema-independent** — the
  version above cannot see it.
- **(b)** a schema-`required` `source:"llm"` field empty or absent.

Replace the `req`/`refused` CTEs with:

```sql
selfc AS MATERIALIZED (        -- the exemption; join it or you will count tool shells
  SELECT cc.id FROM content_components cc
   WHERE COALESCE(cc.component_level,'section')='tool'
     AND COALESCE(cc.input_schema,'{}'::jsonb) = '{}'::jsonb),
req AS MATERIALIZED (
  SELECT cc.id AS component_id, f.key AS field FROM content_components cc
    CROSS JOIN LATERAL jsonb_each(COALESCE(cc.input_schema->'fields','{}'::jsonb)) f
   WHERE f.value->>'source'='llm' AND (f.value->>'required')::boolean IS TRUE),
refused AS MATERIALIZED (
  SELECT pc.page_id, pc.slot_name,
         CASE WHEN COALESCE(pc.content_data,'{}'::jsonb) = '{}'::jsonb
              THEN 'a: no stored content_data' ELSE 'b: missing required llm field' END AS branch
    FROM page_components pc
   WHERE pc.build_status <> 'removed'
     AND NOT EXISTS (SELECT 1 FROM selfc sc WHERE sc.id = pc.component_id)
     AND ( COALESCE(pc.content_data,'{}'::jsonb) = '{}'::jsonb
        OR EXISTS (SELECT 1 FROM req r WHERE r.component_id = pc.component_id
                     AND COALESCE(pc.content_data->>r.field,'') = '') ))
```

`[MEASURED 2026-09-04 12:2xZ]` hole **4 slots** (3 branch a / 1 branch b) / 3 pages / 3 sites —
membership unchanged, because empty `content_data` also leaves every required field absent, so the
old predicate caught them by accident. **Refused-but-escalatable is 73 slots / 66 pages, not the
64/60 first published.** Keep the `branch` column: (a) means the writer authors the WHOLE slot,
(b) means it tops up one field — different repairs.

⚠ **`input_schema.fields` is an OBJECT keyed by field name, not an array.**
`jsonb_path_query_array($.fields[*] ? (@.required == true))` returns a **clean empty result**,
which reads as *"declares no required fields"* rather than *"wrong shape"*. Use `jsonb_each`.
(Found by the `ai-agent-orchestration` lane, 2026-09-04.)

⚠ **Before publishing any "latent / at-risk" population, ask whether it is ELIGIBLE for the
mechanism.** Keying on *unsatisfiable alone* — the pages whose fallback is already gone, held out
of the hole only by intact content — returns **121 pages / 29 sites**, and the true latent exposure
is **ZERO**: 120 of them carry a single self-contained tool component (exempt, never tested) and
the 121st carries no components. Add `LEFT JOIN selfc` and split on it before quoting the number.

### The standing watch (added 2026-09-04 12:4xZ) — the only one of the three figures worth tracking

Append to the corrected CTEs above. It counts **unsatisfiable pages carrying a NON-EXEMPT
component** — the set that can ever enter the hole:

```sql
SELECT ps.domain, ps.url, pcx.slot_name,
       CASE WHEN r.page_id IS NULL THEN 'LATENT (content intact — joins the hole if lost)'
            ELSE 'already in the hole: ' || r.branch END AS state
  FROM pages_scored ps
  JOIN page_components pcx ON pcx.page_id = ps.page_id AND pcx.build_status <> 'removed'
  LEFT JOIN selfc sc ON sc.id = pcx.component_id
  LEFT JOIN refused r ON r.page_id = pcx.page_id AND r.slot_name = pcx.slot_name
 WHERE ps.unsatisfiable AND sc.id IS NULL ORDER BY 4,1,3;
```

`[MEASURED 2026-09-04 12:4xZ]` **5 slots / 3 pages / 3 sites — 4 in the hole, 1 latent**
(ai-agent-orchestration.com `/blog.html` `blog-listing`).

⚠ **Do NOT track "tool components with a non-empty schema" as the early warning.** It sounds like
the right proxy and it is already **56 of 366** — mounted on live pages, harmless, and it will
never trend to zero. Eligibility is the INTERSECTION with an unsatisfiable page; either side alone
is uninformative. ⚠ And the exemption is **per-section**, so a page mixing an exempt tool slot with
a non-exempt one still qualifies through the second — join `page_components`, never `pages`.

### ⚠ CORRECTED 2026-09-04 13:2xZ — the guarded-template figure, and the right patterns

The "Verifying a card-image repair" section above first said *"8 of the 15 components that render
`.image` guard it"*. **Wrong. `[RE-MEASURED 13:2xZ]` it is 14 of 14 — ZERO unguarded**, so
`grep -c 'src=""'` can **never** fire for a missing listing image on this estate.

```sql
-- population: templates that actually BIND an image into an <img src> (a bare
-- '.image' match also catches a JS blob referencing data.imagesBy — not a binding)
WITH t AS (SELECT name, html_template AS h FROM content_components
            WHERE html_template ~ '<img[^>]*src="\{\{[^"]*\.image')
SELECT count(*) AS binds_image,
       count(*) FILTER (WHERE h ~ '\{\{-? *if +[\$a-zA-Z_.]*\.image') AS guarded,
       count(*) FILTER (WHERE h !~ '\{\{-? *if +[\$a-zA-Z_.]*\.image') AS unguarded
  FROM t;   -- 14 / 14 / 0
```

⚠ **The guard pattern must allow a VARIABLE, not just dot.** My first cut,
`\{\{ ?if \.image ?\}\}`, required `}}` immediately after `.image` and at most one space — it
missed `{{if .image_url}}`, `{{if $card.image}}` and `{{if $item.image_url}}`. **A guard is a guard
whichever variable it tests**, and inside a `{{range}}` it will be `$item`/`$card`, which is the
common case for exactly the listing components this lane cares about.

⚠ **`component-render-check` does not cover this and is not meant to.**
`cmd/component-render-check/rendercheck.go` is *"the OUTPUT-level empty-element check"* (its own
first line); `broken_img` is `<img[^>]*\ssrc=""` and a finding is an empty-element shape whose
count **INCREASES** when a field is removed. A guarded field's removal makes the element vanish, so
nothing increases and nothing is reported. Correct scoping — **do not read a clean run as cover**,
and do not file it as a defect. A detector for the vanishing case wants a differential on element
COUNT.

### ⚠ CORRECTED AGAIN 2026-09-04 13:4xZ — resolve the component the way the ACTION does, not by `component_id`

Both corrected versions above still join `content_components` on `pc.component_id`. **That is wrong
and it undercounts.** `rerender_page_sections_action.go:390` falls through to `schemas[s.slotName]`,
and `loadComponentSchemas` (`plan_sections_action.go:1981`) **indexes by BOTH `name` AND
`function`** — so a NULL-`component_id` row still resolves and the gate still judges it.
`[MEASURED 13:4xZ]` **16** active rows carry a NULL `component_id`, **14 resolve** this way; keying
on the column alone wrongly called **3** of them non-exempt and hid **7** branch-(b) refusals.

Replace the `eff`/`refused` head with this and keep everything downstream:

```sql
WITH eff AS MATERIALIZED (              -- component_id FIRST, then the name/function fallback
  SELECT pc.page_id, pc.slot_name, pc.content_data,
         COALESCE(cc_id.id, cc_nm.id)                           AS eff_cid,
         COALESCE(cc_id.component_level, cc_nm.component_level)  AS lvl,
         COALESCE(cc_id.input_schema,   cc_nm.input_schema)      AS schema
    FROM page_components pc
    LEFT JOIN content_components cc_id ON cc_id.id = pc.component_id
    LEFT JOIN LATERAL (
      SELECT c.id, c.component_level, c.input_schema FROM content_components c
       WHERE pc.component_id IS NULL AND (c.name = pc.slot_name OR c.function = pc.slot_name)
       ORDER BY (c.name = pc.slot_name) DESC LIMIT 1) cc_nm ON true
   WHERE pc.build_status <> 'removed'),
refused AS MATERIALIZED (
  SELECT e.page_id, e.slot_name,
         CASE WHEN COALESCE(e.content_data,'{}'::jsonb) = '{}'::jsonb
              THEN 'a: no stored content_data' ELSE 'b: missing required llm field' END AS branch
    FROM eff e
   WHERE NOT (COALESCE(e.lvl,'section')='tool' AND COALESCE(e.schema,'{}'::jsonb)='{}'::jsonb)
     AND ( COALESCE(e.content_data,'{}'::jsonb) = '{}'::jsonb
        OR EXISTS (SELECT 1 FROM jsonb_each(COALESCE(e.schema->'fields','{}'::jsonb)) f
                    WHERE f.value->>'source'='llm' AND (f.value->>'required')::boolean IS TRUE
                      AND COALESCE(e.content_data->>f.key,'') = '') ))
```

`[MEASURED 13:4xZ]` **hole 5 slots** (3 branch a / 2 branch b) / 3 pages / 3 sites ·
**refused-but-escalatable 76 slots** · **LATENT ZERO** — every non-exempt slot on an unsatisfiable
page is already refused. `ai-agent-orchestration.com /blog.html` holds three of the five.

⚠ **`content_data` PRESENT is not `content_data` SUFFICIENT.** Branch (b) fires on a **populated**
map — it tests each required `source:"llm"` field individually. Classifying a slot as healthy
because `content_data` is non-empty is the mirror of classifying it by `component_id`; both were
made today, on the same row, by two lanes.

⚠ **The watch must re-evaluate the JOIN, never cache either side** (the `ai-agent-orchestration`
lane's point, and their own slot is the worked example): a page can become unsatisfiable — a plan
superseded, a `sections` array emptied — with **nothing about the component changing**, and a
component can leave the tool exemption by gaining a schema with nothing about the page changing.
Both sides move independently, so neither count is a proxy for the intersection.

### ⚠ Spot-checking branch (b) BY EYE — near-miss field names read as satisfied

Raised by the `ai-agent-orchestration` lane, 2026-09-04, from the live case. `blog-listing_pre_037`
requires **`section_heading`** and **`section_intro`**; the stored map on both affected pages holds
**`section_title`** and **`section_subtitle`**. Reading that map, it *has* a heading and an intro —
it looks complete, and a by-eye check passes on a row the gate refuses.

**Compare the schema's required keys to the map's keys as SETS, never by reading the map:**

```sql
SELECT f.key AS required_llm_field, COALESCE(pc.content_data->>f.key,'(ABSENT)') AS stored
  FROM page_components pc
  JOIN content_components cc ON cc.id = pc.component_id   -- or the name/function fallback above
  CROSS JOIN LATERAL jsonb_each(COALESCE(cc.input_schema->'fields','{}'::jsonb)) f
 WHERE pc.page_id = $1 AND pc.slot_name = $2 AND pc.build_status <> 'removed'
   AND f.value->>'source'='llm' AND (f.value->>'required')::boolean IS TRUE;
```

⚠ **And `resolveComponent` NAMES TWO DIFFERENT THINGS in this package** — a package-level func in
`rerender_single_page_action.go:1240` (chrome slots) and the **closure** at
`rerender_page_sections_action.go:361` that does the name-or-function fallback. `grep 'func
resolveComponent'` finds only the first, because a closure is `name := func(…)`. **Grep the bare
symbol and cite file:line.** I used the `func` form to tell a peer their citation was wrong; it was
not. Full trap in `LANDMINES.md` under the `content_components.name`/`.function` entry.

### The reflex neither lane had: grep LANDMINES for the TABLE before the first query against it

```sh
grep -n '<table>' docs/agent_docs/docs024_key_docs_latest/LANDMINES.md
```

Same reflex as for a file path — and it is the **only** thing that finds a table-footprinted entry,
because the `SessionStart` hook matches entries against **dirty paths** and a table has none. Both
this lane and `ai-agent-orchestration` skipped it on 2026-09-04 and each re-derived, the hard way,
figures already written down. ⚠ **Do it for every table in the statement, including JOINED ones** —
the query that went wrong read `page_components` and joined `content_components`, and the entry was
footprinted on the joined one.

⚠ **Expect volume, and scan the headings not the lines:** `[MEASURED 2026-09-04]`
`content_components` appears on **173** lines of that file and `page_components` on **333**. That is
why the answer is to read the entry HEADINGS your table appears under, not to grep-and-read.
