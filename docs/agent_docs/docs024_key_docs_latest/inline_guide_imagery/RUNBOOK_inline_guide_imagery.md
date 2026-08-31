# RUNBOOK — inline_guide_imagery

The commands this lane had to get right, with the gotcha attached. Change them HERE when
one changes.

---

## Is the illustrated section component still there, and what does it source?

The component the PLAN proposed to fork already exists — do not build a second one.

```sql
SELECT id, name, function, is_active, jsonb_pretty(input_schema)
  FROM content_components WHERE function='illustrated-text-block';
```

⚠ **Read `input_schema`, not the name.** Migration 644 repointed `image_url` from
`site_assets.image` to `site_assets.illustration` on 2026-08-26. Under the old source the
field alias-resolved to **the page's own hero** (`imageryplan.imageRoleAliases` maps
`image → hero`), so a section "illustration" was silently the banner image again. The name
is identical either way; only the source tells you which world you are in.

## Where does a section's picture actually come from?

```sql
SELECT spi.scope, spi.scope_ref, spi.kind, spi.key, a.asset_key, a.status
  FROM site_plan_imagery spi
  JOIN site_plans sp ON sp.id=spi.plan_id AND sp.is_current
  JOIN sites s ON s.id=sp.site_id
  LEFT JOIN assets a ON a.site_id=sp.site_id AND a.asset_key=spi.key AND a.status='active'
 WHERE s.domain=:domain AND spi.scope='section' AND spi.scope_ref LIKE :page || ':%';
```

⚠ **`scope='page'` rows do NOT feed a section field.** The resolver's section arm filters
`scope='section'` (`plan_sections_action.go:487-493`). apis.uk has five `scope='page'`
illustration rows and six illustrated sections, and those two facts are unrelated to each
other — the sections' images live only in stored `content_data`.

⚠ **The `scope_ref` ordinal is 0-BASED and indexes the page's section list**
(`write_site_plan_imagery_scope.go:276`, `case ordinal >= sectionCount[canonical]`).
`page_components.position` is **1-based** on 847 of 1,065 pages `[MEASURED 2026-08-31]` and
neither 0- nor 1-based-contiguous on 128 of them. The two numbering schemes are not the
same number; never carry one into the other without saying which you have.

## Does a page serve its section images, and how many?

```bash
curl -s https://<domain>/<path> | grep -o '<img[^>]*>' | wc -l
curl -s https://<domain>/<path> | grep -o 'src="[^"]*"' | sort | uniq -c
```

⚠ Every site serves its **logo** as an `<img>`, so the floor is 1, not 0. A guide with "one
image" almost always has zero content images.

## What the page thinks it is composed of — three places that can disagree

```sql
-- 1. the plan (authority)
SELECT sps.ordering, sps.component_name, sps.subject
  FROM site_plan_sections sps JOIN site_plans sp ON sp.id=sps.plan_id AND sp.is_current
  JOIN sites s ON s.id=sp.site_id WHERE s.domain=:domain AND sps.page_name=:page
 ORDER BY sps.ordering;

-- 2. the materialised cache on the page row
SELECT sections FROM pages p JOIN sites s ON s.id=p.site_id WHERE s.domain=:domain AND p.url=:url;

-- 3. what was actually built
SELECT pc.position, pc.slot_name, cc.function, pc.build_status
  FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
  LEFT JOIN content_components cc ON cc.id=pc.component_id
 WHERE s.domain=:domain AND p.url=:url ORDER BY pc.position;
```

⚠ **They diverge, and id wins over name** (`plan_sections_action.go:1236-1245`, observe-only
log). apis.uk `/index.html` is the live example: the plan says `generic-text-block` six
times, the built rows point at `illustrated-text-block`, and `slot_name` still says
`generic-text-block`. A query against any one of the three will read as consistent.

## Is a page's per-section content protected on the next write?

```sql
SELECT pc.slot_name, count(*), count(DISTINCT pc.content_data::text)
  FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
 WHERE s.domain=:domain AND p.url=:url AND pc.build_status='deployed'
 GROUP BY 1 HAVING count(*)>1;
```

⚠ **Any row this returns is a slot with NO carry-forward protection.**
`ensureStoredContent` (`plan_sections_action.go:233-245`) deletes any slot whose rows repeat
with differing `content_data` — repetition is judged by `slot_name`, and
`save_page_sections_action.go:1001` writes `slot_name = component name`. So N sections of
one component on a page are always conflicted, and a save-path regeneration drops every
resolver-sourced key they carry.

## Fleet supply — how many section figures exist at all

```sql
SELECT spi.kind, count(*) AS rows, count(DISTINCT sp.site_id) AS sites
  FROM site_plan_imagery spi JOIN site_plans sp ON sp.id=spi.plan_id AND sp.is_current
 WHERE spi.scope='section' GROUP BY 1 ORDER BY 2 DESC;
```

⚠ **`icon` dominates this table and behaves differently from `illustration`.** Icons run to
ten section-scope rows on one page (webdesign.co.uk index) and are consumed **by key**;
illustration/infographic had at most **one** row per page fleet-wide as of 2026-08-31. Any
change to the kind-alias resolution must be measured against icons separately or it changes
live pages.

## Filing a diagnosis for this lane

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/090_TRIGGER_needs_diagnosis_v1.sh "<symptom>"
```

⚠ The trigger prints TWO ids. The intake `CORRELATION_ID` is not the key the artifacts are
written under — save the **`RUN_CORRELATION_ID`** printed after "waiting for the loop to
claim it". Progress:

```sql
SELECT current_step, status, updated_at FROM orchestration_states
 WHERE correlation_id::text='<RUN_CORRELATION_ID>' ORDER BY updated_at DESC LIMIT 3;
```
