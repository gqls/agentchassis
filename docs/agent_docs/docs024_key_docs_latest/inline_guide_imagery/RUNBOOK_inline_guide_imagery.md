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

## Is the per-section binding live in the running chassis?

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis --no-headers -o custom-columns=NAME:.metadata.name | head -1)
for sym in PlanSectionsAction sectionRefForOrdinal planSectionOrder sectionOrderAgrees newSectionRef NextOccurrence sectionOrderAgreesNOTREAL; do
  kubectl -n ai-persona-system exec $POD -- grep -aq "$sym" /proc/1/exe && echo "PRESENT $sym" || echo "absent  $sym"
done
```

⚠ **CORRECTED 2026-09-03 — the "only `image_landed`/`section_data_resolved` re-resolve" line I put
in this RUNBOOK is WRONG.** I quoted `rerender_page_sections_action.go:47`'s comment; the LIVE
`page-rerender` config gates on **FIVE** reasons — `image_landed`, `section_data_resolved`, `cta_links_stale`, `template_changed`, `literal_markdown` — and the comment has drifted.
**Read the config, not the header:**

```sql
SELECT s.key, s.value->'config'->>'condition'
  FROM agent_definitions a, jsonb_each(a.default_config->'workflow'->'steps') s
 WHERE a.type='page-rerender' AND a.is_active AND COALESCE(a.is_snapshot,false)=false
   AND a.deleted_at IS NULL AND s.value->>'action'='conditional';
```

⚠ **AND the second half is UNSETTLED:** whether the sections path actually re-resolves
`site_assets.*` when it runs is under test (`bugs_open/425` §2 reports it does not for `query.*`,
reproduced four times; the one page traced as recovering did so via the BUILD path). Do not quote
"a re-render will pick it up" as fact until that experiment reports.

⚠ **`kubectl logs … | grep 'build provenance'` does NOT work on this service** — verified
2026-09-02, and its failure is the dangerous kind: the phrase DOES appear in the logs, inside LLM
prompt text describing the check, so a careless grep returns a hit that looks like a stamp. Probe
the binary, and always run BOTH controls (`PlanSectionsAction` must be present,
`sectionRefForOrdinalNOTREAL` must be absent) — a grep that matches everything and a grep that
matches nothing produce the same confident answer otherwise.

⚠ **Probe the symbols of the round you are checking, not the round you remember.** `debug_historian` flagged this on IMG-075's round 3: the probe recorded here originally listed only round 1's symbols (`sectionRefForOrdinal`, `planSectionOrder`), which are PRESENT on `v1.0.1351` and `v1.0.1352` while round 2/3's guards (`sectionOrderAgrees`, `newSectionRef`, `NextOccurrence`) are absent from both — so the old list returns all-present on a binary carrying half the change, and reads as a clean deploy.

⚠ Probe the **capability** (a symbol the change introduced), not the commit sha: the stamp names
one commit, so a sha probe answers "was the build cut exactly here", which is not the question.

## Would a given page bind, or stand down?

The binding engages only when the plan's section order and the page's live section order describe
the same sequence of slots. Check both sides before predicting anything:

```sql
-- plan side (site-level slots are filtered out of the comparison)
SELECT sps.ordering, sps.component_name FROM site_plan_sections sps
  JOIN site_plans sp ON sp.id=sps.plan_id AND sp.is_current
  JOIN sites s ON s.id=sp.site_id
 WHERE s.domain=:domain AND sps.page_name=:page ORDER BY sps.ordering;

-- live side (what the build loop / rerender actually iterates)
SELECT pc.position, pc.slot_name FROM page_components pc
  JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
 WHERE s.domain=:domain AND p.url=:url ORDER BY pc.position;
```

⚠ **Compare `slot_name`, not the component the row points at.** apis.uk/index is the worked
example: its slots are `generic-text-block` while its `component_id` resolves to
`illustrated-text-block`. The plan and the slots agree, so it binds; the component names do not,
and comparing those would predict the opposite.

## Seeding a per-section figure (the supply half)

```sql
INSERT INTO site_plan_imagery (plan_id, scope, scope_ref, key, kind, prompt, ordering, source, locked_at, locked_by)
SELECT sp.id, 'section', :page || ':' || :ordinal, :asset_key, 'illustration', :prompt, :ordinal, 'manual', now(), :lane
  FROM site_plans sp JOIN sites s ON s.id=sp.site_id AND s.domain=:domain
 WHERE sp.is_current;
```

⚠ `prompt` is **NOT NULL**; for an asset that already exists it is documentation for the next
planner, not a generation request. ⚠ The ordinal is the **plan's** `ordering` (0-based, counts
site-level slots) — on a page whose plan starts with a hero, live `position - 1`. ⚠ `locked_at`
set, or IMG-013 lock transfer will not carry the row across the next replan.

## Is a planned section figure actually reaching its page? (the Phase-4 tripwire, by hand)

The check `check_unrendered_section_imagery` would automate this. Until it exists, run it:

```sql
WITH planned AS (
  SELECT sp.site_id, s.domain, split_part(spi.scope_ref,':',1) AS page_name, spi.scope_ref, spi.kind, a.asset_key
  FROM site_plan_imagery spi
  JOIN site_plans sp ON sp.id=spi.plan_id AND sp.is_current
  JOIN sites s ON s.id=sp.site_id
  JOIN assets a ON a.site_id=sp.site_id AND a.asset_key=spi.key AND a.status='active'
  WHERE spi.scope='section' AND spi.kind IN ('illustration','infographic'))
SELECT pl.domain, pl.scope_ref, pl.asset_key,
       EXISTS (SELECT 1 FROM pages p JOIN page_components pc ON pc.page_id=p.id
                WHERE p.site_id=pl.site_id AND p.name=pl.page_name
                  AND pc.rendered_html LIKE '%'||replace(pl.asset_key,'_','-')||'%') AS rendered
FROM planned pl ORDER BY 1,2;
```

`[MEASURED 2026-09-02]` 4 rows, **2 rendered nowhere** — fundamentallyai `about:2`, vonc
`about:2`. Contributed to `bugs_open/114`.

⚠ **This is an ABSENCE query, so it needs its control** — the `rendered` column being false and
the `LIKE` pattern being wrong are the same output. The two `true` rows ARE the positive control
here (same predicate, same derivation, they match); if a run ever returns all-false, suspect the
pattern before believing the finding. The pattern approximates
`storage.DeployedWebPath(asset_key, purpose)`; read that function if the spelling ever changes.

⚠ **A false row is not yet a defect — read the section's field source before concluding.** Both
2026-09-02 misses were the component's own declaration, not a queue: one sourced
`site_assets.image` (which the alias map resolves to the page HERO, unconditionally), the other
declared **no source at all**, which is written by nobody and resolved by nobody:

```sql
SELECT cc.function, f.k, f.v->>'source', f.v->>'on_missing'
FROM content_components cc, jsonb_each(cc.input_schema->'fields') AS f(k,v)
WHERE cc.is_active AND cc.function = :section_component AND (f.k ILIKE '%image%' OR f.v->>'source' LIKE 'site_assets%');
```
