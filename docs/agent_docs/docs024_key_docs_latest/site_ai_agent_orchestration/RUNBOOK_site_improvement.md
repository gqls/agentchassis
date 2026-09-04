# RUNBOOK — ai-agent-orchestration.com improvement

Commands that were hard to get right, with the gotcha attached. Site id
`2a8ebf9c-20a2-4c39-b191-840b012371da`.

---

## R1 — Measure contrast at the artefact (the only instrument that sees it)

```bash
timeout 300 python3 scripts/render_audit.py --json out.json \
  https://ai-agent-orchestration.com/index.html \
  https://ai-agent-orchestration.com/about.html \
  https://ai-agent-orchestration.com/pricing.html \
  https://ai-agent-orchestration.com/services.html
```

**Gotchas.**
- It needs a Chromium; it finds one itself (`$CHROME`, the Playwright cache, `/usr/bin/chromium`).
  **`import playwright` is NOT available in this environment** — the script shells out to
  `chrome --headless=new --dump-dom` instead. Do not write a probe that imports playwright; it
  fails with `ModuleNotFoundError`. Reuse this script's own technique (below, R2).
- **Filter `overImage` before quoting a total.** 47 raw findings were 44 firm + 3 approximations;
  the adapter itself calls an over-image backdrop unknown, and a firm/approximate mix quoted as
  one number overstates the defect.
- `images` in the JSON is *confirmed-broken after a serial re-check*, and `verify_broken` **skips
  any image with an empty `src`**. So a page of empty `<img>` tags reports `broken=0`. An
  `images_reported=5, broken=0` therefore means "5 images failed in-browser and none survived
  re-checking" — which on `index` is 5 empty srcs, not 5 healthy images. Read both numbers.

## R2 — Ask the browser for a computed token (never grep the stylesheet)

Reuses `render_audit.py`'s injection technique: fetch the page, add `<base href>` so relative
assets still resolve against the live origin, inject a probe that appends `<pre id="AUDIT_RESULT">`,
then read it back out of `--dump-dom`. Full script:
`scratchpad/probe2.py` pattern in `NOTES`; the load-bearing part is

```python
sys.path.insert(0, os.path.abspath("scripts")); import render_audit as ra
chrome = ra.find_chrome(); page = ra.fetch(url)
inj = re.sub(r"<head([^>]*)>", r"<head\1><base href='%s'>" % base, page, count=1)
inj = inj.replace("</body>", "<script>%s</script></body>" % PROBE)
subprocess.run([chrome,"--headless=new","--disable-gpu","--no-sandbox",
                "--virtual-time-budget=10000","--dump-dom","file://"+path], ...)
```

**Why it matters here:** the served stylesheet contains `h3, .site-footer h4 { color: #ffffff; }`
— white, legible, and **not the winning declaration**. The component's own embedded `<style>`
block in `rendered_html` overrides it. Reading the site stylesheet would have produced a
confidently wrong answer.

## R3 — Extract a component's embedded CSS rule

The rule lives inside `page_components.rendered_html`, not in any stylesheet file.

```sql
SELECT p.name, pc.slot_name,
       regexp_replace(substring(pc.rendered_html from '[^{};]*background[^;]*(#fff|255, *255, *255)[^;]*'),
                      '\s+',' ','g') AS decl
FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE p.site_id='<site>' AND pc.rendered_html ~* 'background[^;]*(#fff|255, *255, *255)';
```

**Gotcha.** A naive `substring(... from 'h3[^{]*\{[^}]*color[^}]*\}')` matches the page's **prose**
(`<h3>Agile Orchestration Architecture</h3>` … ) long before it matches CSS, and returns a
paragraph of marketing copy that looks like a failed query. Anchor on `<style>` or on a
declaration-shaped pattern, and be aware `rendered_html` is multi-line — a bash line-oriented
split over psql output silently drops every component.

## R4 — Census the `<img>` srcs for a site

```sql
WITH srcs AS (
  SELECT p.name, pc.slot_name,
         (regexp_matches(pc.rendered_html,'<img[^>]*src="([^"]*)"','g'))[1] AS src
  FROM page_components pc JOIN pages p ON p.id=pc.page_id
  WHERE p.site_id='<site>')
SELECT name, slot_name, CASE WHEN src='' THEN '(EMPTY)' ELSE src END FROM srcs ORDER BY 1,2;
```

Confirm a 404 over HTTP rather than from the work item, which only compares against the `assets`
table: `curl -s -o /dev/null -w '%{http_code}' <url>`.

## R5 — Check whether a handler actually repairs, before routing work at it

Two steps, both cheap, and skipping either is how work gets routed at a handler that removes the
thing you wanted built.

```sql
-- 1. what does it DO?  (an action list is enough to spot a triage-only handler)
SELECT jsonb_pretty(jsonb_path_query_array(default_config,'$.workflow.steps.*.action'))
FROM agent_definitions WHERE type='<handler>'
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- 2. where has it run, and what did the ARTEFACT look like afterwards?
SELECT s.domain, w.status, count(*) FROM site_work_items w JOIN sites s ON s.id=w.site_id
WHERE w.item_type='<type>' AND w.handler_agent<>'' GROUP BY 1,2;
```

Then go and look at that site's artefact (R4). `image-source-unsatisfiable-handler` reads as the
obvious fix for a missing image and its only precedent site now has **zero** `<img>` tags.

## R6 — Fleet palette outlier check

```sql
SELECT s.domain,
       ss.data->'palette'->'reference_values'->>'primary' AS primary,
       ss.data->'palette'->'reference_values'->>'surface' AS surface,
       (ss.data->'palette'->'reference_values'->>'primary')
     = (ss.data->'palette'->'reference_values'->>'surface') AS degenerate
FROM site_specs ss JOIN sites s ON s.id=ss.site_id
WHERE ss.is_current AND ss.aspect='design_intent'
  AND ss.data->'palette'->'reference_values' ? 'primary'
ORDER BY degenerate DESC NULLS LAST;
```

**Gotcha.** The table is `site_specs` with columns `aspect` / `data` / `is_current` — **not**
`spec_type` / `content`. `\d site_specs` first.

## R7 — Is a page repairable at all?

```sql
SELECT p.name, count(*) comps, count(*) FILTER (WHERE pc.content_data IS NULL) nulls,
       min(pc.updated_at)::date oldest, max(pc.updated_at)::date newest
FROM pages p JOIN page_components pc ON pc.page_id=p.id
WHERE p.site_id='<site>' GROUP BY 1 ORDER BY oldest;
```

A page with `nulls = comps` **cannot be re-rendered** — `rerender_page_sections` rebuilds a
section from `content_data` and there is none. It must be rebuilt through the framework.
NEVER restore from `page_component_history` (its `component_id` is NULLed by
`ON DELETE SET NULL`; see `bugs_closed/194` §4).

## R8 — Propagate a component-TEMPLATE change to the live pages

A `content_components.html_template` edit is inert until each placement re-renders. Getting this
wrong is silent: the wrong route reports **success** and ships the old bytes.

```sql
-- Page-scoped rerenders for every page carrying a changed component, ON ONE SITE.
INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary,
                             priority, handler_agent, status, created_by, spec)
SELECT DISTINCT p.site_id, 'side_effect', 'build', 'page_rerender', 'low',
       'Rerender page after template fix: ' || p.name,
       80, 'page-rerender', 'triaged', '<who-you-are>',
       jsonb_build_object('reason','template_changed','page_id',p.id::text,
                          'page_name',p.name,'domain',s.domain)
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE pc.component_id IN (<the component ids>)
  AND p.site_id = '<site>'
  AND p.rebuild_policy IS DISTINCT FROM 'owned'
  AND NOT EXISTS (SELECT 1 FROM site_work_items w
                  WHERE w.site_id=p.site_id AND w.item_type='page_rerender'
                    AND w.spec->>'page_id'=p.id::text
                    AND w.spec->>'reason'='template_changed'
                    AND w.status IN ('detected','triaged','claimed'));
```

**Gotchas, all four of which cost something if missed.**

- **`spec.reason` is what selects the code path.** `page-rerender.check_rerender_mode` routes to
  `rerender_sections` (regenerates from `content_data` + the template) ONLY for
  `image_landed | section_data_resolved | cta_links_stale | template_changed`. Every other value,
  **including none**, falls to `render_page`, which assembles the STORED `rendered_html` — old CSS,
  green status. Read the live condition before trusting this list; it has grown:
  ```sql
  SELECT default_config #>> '{workflow,steps,check_rerender_mode,config,condition}'
  FROM agent_definitions WHERE type='page-rerender'
    AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  ```
- **Scope to the pages that carry the component; never fire a site-wide `needs_rerender`.** On this
  site `privacy` and `terms` hold **permanently locked** `generic-text-block` components, and a
  rerender aimed at a locked positionally-named section **duplicates** it (`bugs_open/189`). Count
  `page_components` for the page before and after if you ever do hit one.
- **Copy the shape from the LIVE agent row, not from the migration that introduced it.**
  `460_template_changed_rerender_reason.sql` puts `p.filename` in the spec and **`pages` has no such
  column** — that is what `461_fix_460_…` exists to fix. The live query is the corrected one:
  ```sql
  SELECT default_config #>> '{workflow,steps,create_rerender,config,query}'
  FROM agent_definitions WHERE type='component-template-fixer'
    AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  ```
- **`p.rebuild_policy IS DISTINCT FROM 'owned'`** — an `owned` page refuses with
  `save_page_sections: overwrite: REFUSED` rather than silently doing nothing, but filtering keeps
  the queue honest.

Then verify at the artefact, **by colour pair** (R1). A pair present in the after-set and absent
from the before-set is a regression you introduced — the check migration 456 lacked, which is how
its `.stats-cta` regression survived a net 44→33 improvement.

## R9 — Apply ONE migration without sweeping other lanes'

`run-migrations.sh --apply` takes **every** pending file; on 2026-08-18 that was 17 files from at
least six lanes. Apply yours alone, then record it:

```bash
# 0. rehearse: the real file, guards and all, with the final COMMIT swapped for ROLLBACK
sed 's/^COMMIT;$/ROLLBACK;/' <file>.sql > /tmp/rehearsal.sql
kubectl exec -i -n ai-persona-system postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < /tmp/rehearsal.sql
# 1. apply (same invocation the runner itself uses)
kubectl exec -i -n ai-persona-system postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < <file>.sql
# 2. record
./scripts/migration/run-migrations.sh --record-only <file>.sql --note "<what you checked>"
```

**Gotchas.**
- **A dry-run `SELECT` proves your ANCHORS match; only the rehearsal proves your GUARDS pass.** Mine
  passed the anchor dry run and would have failed at COMMIT on a guard I had scoped too widely.
- **"Pending" ≠ "unapplied".** The listing shows files applied by hand and never recorded. 460 was
  listed pending while `template_changed` was already live in `page-rerender`. **Check the live row,
  not the ledger**, before concluding a mechanism does not exist.
- **`--no-probe` for a fast listing.** The default probe executes every pending file inside a doomed
  transaction, which took longer than a 240 s timeout here.
- **Next number = highest in the directory + 1, and numbers still collide** (457 and 458 each name
  two unrelated migrations). Re-check immediately before writing: 470 appeared during this session.

## R10 — Apply the managed flip (617 HOLD) ONLY after the carry has rolled, then prove it survived

`617_aiao_writer_block_managed_with_guidance_carry_HOLD.sql` is inert-by-name (the runner's `SIDECAR_RE`
excludes `_HOLD`). It needs the CLM-029 carry (`c17a18620`) in the RUNNING chassis, and the file cannot
check git ancestry itself — so this one-liner does the check first and only then pipes the file:

```bash
PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"
LIVE=$($PSQL -Atc "SELECT git_commit FROM service_binary_capabilities WHERE service='agent-chassis'
        AND last_seen_at > now()-interval '30 minutes' GROUP BY 1 ORDER BY max(last_seen_at) DESC LIMIT 1")
echo "running chassis: $LIVE"
git merge-base --is-ancestor c17a18620 "$LIVE" && echo CARRY-IS-LIVE || { echo "NOT YET — do not apply"; false; }
# only if CARRY-IS-LIVE:
$PSQL -v ON_ERROR_STOP=1 -v live_chassis="$LIVE" \
  < docs/agent_docs/sql_for_agents/617_aiao_writer_block_managed_with_guidance_carry_HOLD.sql
./scripts/migration/run-migrations.sh --record-only 617_aiao_writer_block_managed_with_guidance_carry_HOLD.sql \
  --note "applied by hand after merge-base proved c17a18620 in $LIVE"
```

**Gotchas.**
- **The guard refuses `4c996e1b5…` by name** (the chassis live on 2026-08-25) and any sha that is not the
  one heartbeating. It does NOT know whether a NEWER sha contains the carry — that is what the
  `merge-base` line is for. Skipping it and typing the running sha after a roll that lacks `c17a18620`
  would pass the guard and delete the prohibitions at the next refresh.
- **`-v live_chassis` unset → `617 REFUSED … Got: (unset)`, exit 3.** Rehearsed. The `\if :{?live_chassis}`
  at the top turns a missing variable into an empty string so the refusal is the guard's message, not a
  psql syntax error with exit 0.
- **A `--record-only` for a `_HOLD` file works** (the record path takes any filename); the apply path
  never sees it.

**The survival check — the first ~09:06Z refresh after applying.** Managed regeneration replaces
`writer_block` only when the composed text differs (`existing != block`,
`refresh_evidence_base_action.go` "Whitelist regeneration"). 617 wrote the composer's exact output, so:

```sql
SELECT created_by, created_at,
       md5(data->>'writer_block') = md5((SELECT old_value->'data'->>'writer_block' FROM migration_backups
                                          WHERE migration_name='617_aiao_writer_block_managed_with_guidance_carry')) AS still_611,
       md5(data->>'writer_block') AS wb_md5,
       (data->>'writer_block') ~ 'NOT TRACKED / DOES NOT EXIST, NEVER STATE' AS has_never_state,
       (data->>'writer_block') ~ '\mNNN\M' AS has_nnn,
       jsonb_array_length(data->'facts') AS facts
FROM site_specs
WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND aspect='evidence_base' AND is_current;
```
Expect: `created_by='evidence-refresher'`, `still_611=f`, `has_never_state=t`, `has_nnn=f`, `facts=8`, and
`wb_md5` equal to the md5 of the 617 constant (compute it from the file:
`python3 -c "import re,hashlib;t=open('docs/agent_docs/sql_for_agents/617_aiao_writer_block_managed_with_guidance_carry_HOLD.sql').read();print(hashlib.md5(re.findall(r'\\$WB\\$(.*?)\\$WB\\$',t,re.S)[0].encode()).hexdigest())"`).
A different md5 with `has_never_state=t` means the composer's output differs from the prediction in some
harmless way — diff it and record it; `has_never_state=f` means the carry is NOT in the running binary
after all — run the ROLLBACK sidecar and re-check the merge-base.

**Compose offline with the REAL function, without touching the shared tree.** Map a scratch test file
into the package with `go test -overlay` (the harness used to build 617; the live row's output is in
NOTES 2026-08-25 §3):

```bash
S=<scratchpad>; kubectl … psql -At -c "SELECT data FROM site_specs WHERE site_id='…' AND aspect='evidence_base' AND is_current" > $S/eb.json
cat > $S/zz_preview_test.go <<'GO'
package actions
import ("encoding/json";"os";"testing")
func TestPreview(t *testing.T){ raw,_:=os.ReadFile(os.Getenv("EB")); var eb map[string]interface{}
  if err:=json.Unmarshal(raw,&eb); err!=nil { t.Fatal(err) }
  os.WriteFile(os.Getenv("OUT"), []byte(composeWriterBlock(eb)), 0o644) }
GO
printf '{"Replace":{"%s":"%s"}}' "$PWD/platform/orchestration/actions/zz_preview_test.go" "$S/zz_preview_test.go" > $S/overlay.json
EB=$S/eb.json OUT=$S/composed.txt go test -overlay $S/overlay.json -run TestPreview -count=1 ./platform/orchestration/actions/
```
`git status platform/orchestration/actions/` shows nothing of yours afterwards — the file never existed
on disk in the tree, so no other session's `git add -A` can sweep it.

**R10 addendum (council `35ab8b23` r1, applied same day).** Two things changed after the approval:
- The chassis guard also refuses any running sha whose pods **started before 2026-08-25 12:49:19Z** (the
  commit time of `c17a18620`) — a binary built before the carry cannot contain it. **Necessary, not
  sufficient**: a later restart on an old image passes it, and so would a roll that reverted the carry. The
  `merge-base` line above stays load-bearing. (Mutation-proven: with the by-name refusal deleted and today's
  sha passed, the file refuses on `started_at`.)
- **Run the survival query again after the SECOND ~09:06Z refresh.** The first refresh reads the 617 row
  (migration-written); the second is the first to re-read a refresher-WRITTEN row, which is where a typed
  struct round-trip would bite if one existed. CLM-029's round surveyed all 9 `ParseEvidenceBase` callers
  (readers/guards only) and pinned both real write paths with round-trip tests, so expect the same md5 on
  day 2 — but expect it by looking.

> **CORRECTED 2026-08-26 (found by running it):** R10's line "*a `--record-only` for a `_HOLD` file
> works*" is FALSE — the runner refuses UPPERCASE-suffixed sidecars for recording too ("recording one is
> meaningless"). Nothing dangles: `_HOLD` files are also excluded from the pending listing, and the
> application record IS the database — the `created_by='617_migration'` spec row plus the
> `migration_backups` row. Skip step 4; verify those two rows instead. (617 was applied this way
> 2026-08-26 09:41:16Z against chassis `2fb40a960`, all guards passing; wb md5 `fa0a4710…` verified.)

## R-council — before submitting: confirm every symbol your plan rests on is INDEXED

One query, before the gate reads your submission. The council's `code_checks` consults
`code_symbols`, and **a Go closure (`name := func(…)`) is not in it** — so a seat asking "does this
symbol exist?" gets zero rows and can object on a premise that is not actually false.

```sql
SELECT symbol, path, line_start FROM code_symbols WHERE symbol = '<the symbol your plan names>';
```

**A zero means "check whether it is a closure" before it means anything else.** Verified with
controls 2026-09-04: `resolveComponent`'s closure (`rerender_page_sections_action.go:361`) is
absent while `escalateRerenderToWriter` (:1178) and `isSelfContainedSection` (:1149) — top-level
funcs in the **same file** — are both present. The file is indexed; the closure form is not. Not
staleness: the indexed commit is an ancestor of HEAD and the closure is at 361 in that commit.

Round 3 of `458` rested on `buildLegibleInkDefaults` (`palette_specialised_slots.go:703`), which
**is** indexed — so it drew no such objection. That was luck, not design, and this check is what
converts it to design. Full trap: `LANDMINES.md`, *"`grep 'func <name>'` CANNOT MATCH A CLOSURE"*.

**And when the landmine-verifier returns `NEEDS_HUMAN_REVIEW`, read the BODY, never the status.**
It says one of three things: *"I checked and found nothing"* (evidence), *"I cannot see this class"*
(silence — `LANDMINES.md:7731`, the index is 100% Go and **81% of footprints are not**), or *"I
found nothing AND here is a property of my instrument that would produce exactly this"* (a lead
worth more than the verdict). ⚠ **Only the third names a testable hypothesis, and triaging on the
status column discards it along with the second.**
