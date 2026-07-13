# 021 — Site Spec & Classifier Architecture

Consolidates thinking from: 002c (system architecture), 002d (quality assurance), 003d (contracts), 004 (classifier notes), 002 (domain content strategy), 003 (deep research), 004 (dispatch loop), 009 (improvement loop).

---

## The Problem Right Now

Site planning data is scattered:

| Where | What's there | Who reads it |
|-------|-------------|--------------|
| `sites.content_data` | Planner output (pages, style, design notes), image URIs, LLM response blob | page-content-writer, webdesign-agent, build_render_context |
| `site_specs` | Empty for old sites; newer pipeline uses `aspect` rows (identity, classification, briefing, site_plan, strategy) | build-site-planner, read_site_spec |
| Agent definition prompts | Hardcoded assumptions about tone, audience | page-content-writer LLM prompt |

The audit agents query `site_specs` and find nothing for sites built by `pageflow-builder`. The content-quality-auditor reports "no target audience defined" because the data is in `content_data.response`, not in `site_specs`.

The planner and classifier have overlapping responsibilities — both decide what pages to create. The webdesign-agent guesses design direction from the industry name because no explicit design intent exists.

---

## Target State

One authoritative spec per site, stored in `site_specs`, versioned, readable by every downstream agent.

### The Unified Spec

Not a new concept — this is what `site_specs` was designed for. The issue is that the older pipeline (`pageflow-builder`) predates it and wrote to `content_data` instead. The newer pipeline (`build-site-planner`) already uses `site_specs` with aspects like `identity`, `classification`, `briefing`, `site_plan`, `strategy`.

The unified spec is the collection of all aspects for a site. Each aspect is a separate row in `site_specs` with `is_current = true`. Together they describe everything the site should be.

```
site_specs rows for gaswholesalers.com:
  aspect: classification  → build_approach, archetype, industry, builder
  aspect: identity        → company_name, tone, audience, key_messages, differentiators
  aspect: strategy        → content strategy framework answers (the 15 questions)
  aspect: design_intent   → style direction, colour mood, typography, imagery, layout
  aspect: content_direction → voice, avoid phrases, emphasis, social proof style
  aspect: site_plan       → pages array with status, priority, sections, handlers
  aspect: seo             → keywords, local_seo, schema_types
  aspect: maintenance     → audit config, refresh schedule, link checking
```

Each aspect is independently versioned. Changing the design intent doesn't create a new site plan version. The `is_current` / `superseded_at` columns already handle this — when a new version of an aspect is written, the old one gets `is_current = false, superseded_at = NOW()`. Full history is preserved.

### One Spec, Not Two

There is no separate "dream spec" and "build spec." Every item in the spec has a status:

- `deployed` — built and live
- `planned` — achievable with current agents, queued for build
- `blocked` — needs an agent/capability that doesn't exist yet

The "dream" is the full spec. The "build" is the subset that isn't blocked. Gap analysis is just `SELECT * FROM ... WHERE status = 'blocked'`.

When a new agent is deployed, the `feasibility-recheck` scheduled task promotes matching blocked items back to `planned`/`triaged`. The spec doesn't change — the system's capability catches up to the spec's ambition.

---

## Who Writes What

### Classifier

The classifier is the first agent. It runs research (domain analysis, competitor scraping, industry patterns) and produces the broadest possible spec for the site. It writes:

- `classification` — routing decisions (builder, archetype, hosting)
- `identity` — who the company is, tone, audience, messaging
- `design_intent` — what the site should look and feel like
- `content_direction` — voice, emphasis, things to avoid
- `seo` — keyword targets, schema types
- `maintenance` — which audit groups to enable

The classifier is aspirational — it describes the best version of this site. It includes pages and features the system can't build yet, marked as `blocked`. It draws from the domain content strategy framework (the 15 questions) and deep research principles to identify what would make this site genuinely useful in its niche.

The classifier does NOT decide which components to use or which style collection to apply — those are implementation details for the planner and design agent.

### Planner

The planner receives the classifier's spec and adds implementation detail. It writes:

- `site_plan` — enriched pages array with specific component names, section ordering, research requirements

The planner's job is validation and enrichment:
1. Check the classifier's page list against available components
2. Select specific components for each section (hero → `hero-split` or `hero-fullwidth`)
3. Set feasibility on pages/features (check agent registry)
4. Write work items for everything that's `planned`

The planner does NOT change the classifier's intent — it implements it. If the classifier says "professional-dark with blue/orange palette," the planner doesn't switch to "minimal-light." If it disagrees, it notes why in `strategy_notes` but defers to the classifier's judgement (or flags for HITL).

### Design Agent

The design agent reads `design_intent` from the spec and produces CSS. It writes:

- CSS theme to `css_themes`
- Style collection assignment to `style_collections` / `sites.style_collection_id`

It does not write to `site_specs` — it reads the intent and implements it.

### Content Writer

Reads `identity`, `content_direction`, and the page's section spec from `site_plan`. Writes content to `page_components`.

### Audit Agents

Read the full spec as ground truth. Compare current rendered state against the spec's intent. Deviations become work items. The audit enforces the spec — it doesn't make its own design or content decisions.

```
Classifier wrote: tone = "professional, direct, trustworthy"
Content writer produced: "We synergize cutting-edge solutions..."
Audit finds: tone mismatch → work item: content_rewrite
```

---

## Backward Compatibility

### pageflow-builder sites (finetuning.uk, gaswholesalers.com)

These have data in `sites.content_data` but no `site_specs` rows. Two options:

**Option A: Backfill.** Write a migration that extracts data from `content_data` and creates `site_specs` rows. This is a one-time operation. The old `content_data` stays (other code still reads it) but `site_specs` becomes the authoritative source.

```sql
-- Example backfill for identity aspect
INSERT INTO site_specs (site_id, aspect, data, source, created_by)
SELECT s.id, 'identity', jsonb_build_object(
    'company_name', COALESCE(s.company_name, s.domain),
    'tagline', COALESCE(s.tagline, ''),
    'tone', COALESCE(s.content_data->'response'->>'tone', ''),
    'target_audience', COALESCE(s.content_data->'response'->>'target_audience', '')
), 'backfill', 'migration'
FROM sites s
WHERE NOT EXISTS (
    SELECT 1 FROM site_specs ss WHERE ss.site_id = s.id AND ss.aspect = 'identity' AND ss.is_current = true
);
```

**Option B: Read from both.** The audit agents (and any agent needing spec data) try `site_specs` first, fall back to `content_data`. This is the `read_site_spec` action's job — it already exists and can be extended with fallback logic.

**Recommendation: Do both.** Backfill existing sites now (option A). Update `read_site_spec` to handle fallback (option B) for any edge cases where backfill missed something. Pageflow-builder continues working — it writes to `content_data` as before, and we add a post-build step that syncs to `site_specs`.

### pageflow-builder modification

Add a step after the planner in pageflow-builder that writes the plan to `site_specs`:

```json
"sync_to_site_specs": {
    "action": "write_site_spec",
    "config": {
        "aspect": "site_plan",
        "source": "pageflow-builder",
        "site_id": "site_record.site_id",
        "spec_data": "validated_plan"
    },
    "next_step": "... existing next step ...",
    "output_field": "spec_written"
}
```

This is additive — existing `content_data` writes stay. Both locations have the data. Over time, downstream agents migrate to reading from `site_specs`.

---

## The Content Strategy Framework Integration

The domain content strategy framework (the 15 questions from doc 002) and the deep research approach (doc 003) describe what makes a site genuinely valuable. These feed into the classifier's spec generation.

### How the 15 questions map to spec aspects

| Questions | Spec aspect | What's captured |
|-----------|------------|-----------------|
| 1-4 (who, intent, satisfaction, money flow) | `identity` + `classification` | Audience, purpose, monetisation model |
| 5-7 (competitors, table stakes, gaps) | `strategy` | Competitive landscape, opportunities |
| 8-9 (money flow, monetisation) | `strategy` | Revenue model, lead values |
| 10-11 (pages needed, format per page) | `site_plan` | Page list with purposes and formats |
| 12-13 (original elements, recurring value) | `strategy` + `content_direction` | Differentiation, engagement hooks |
| 14 (seasonal patterns) | `maintenance` | Content timing, traffic expectations |
| 15 (trust threshold) | `content_direction` | Research depth, authority signals needed |

The classifier doesn't need to answer all 15 questions in one LLM call. It can use the research agent (questions 5-7), the briefing agent (questions 1-4 from human input), and its own analysis (questions 8-15) across multiple steps.

### Deep research as a spec enrichment

The deep research framework (knowledge clusters, primary sources, gap analysis) is a future enrichment to the classifier's research step. Currently the classifier does basic domain analysis. The deep research approach would add:

1. Primary source identification per industry
2. Knowledge domain mapping
3. Content gap analysis against competitors
4. Authority-building page recommendations

This doesn't change the spec structure — it makes the classifier produce richer `strategy` and `content_direction` aspects. The spec format stays the same.

---

## Audit Agent Ground Truth

With the unified spec in `site_specs`, the audit agents read:

| Audit agent | Reads | Checks against |
|-------------|-------|---------------|
| visual-design-auditor | `design_intent` | Rendered CSS, component HTML |
| content-quality-auditor | `identity`, `content_direction` | Page content text, tone, CTAs |
| site-review-agent | All aspects | Overall site vs stated purpose |

The `load_brief` step in `content-quality-auditor` changes from:

```sql
-- Current: queries site_specs only
SELECT ... FROM sites s LEFT JOIN site_specs ss ON ...
```

To:

```sql
-- New: queries site_specs with content_data fallback
SELECT s.domain,
    COALESCE(s.company_name, s.domain) as company,
    COALESCE(
        (SELECT data->>'tone' FROM site_specs WHERE site_id = s.id AND aspect = 'identity' AND is_current = true),
        s.content_data->'response'->>'tone',
        ''
    ) as tone,
    COALESCE(
        (SELECT data->>'target_audience' FROM site_specs WHERE site_id = s.id AND aspect = 'identity' AND is_current = true),
        s.content_data->'response'->>'target_audience',
        ''
    ) as target_audience
    -- ... etc
FROM sites s WHERE s.id = $1
```

This works for both old and new sites. As backfill runs and new sites use `site_specs` natively, the fallback paths become dead code.

---

## Versioning and Rollback

`site_specs` already supports versioning:

```sql
-- Current version
SELECT * FROM site_specs WHERE site_id = $1 AND aspect = 'site_plan' AND is_current = true;

-- History
SELECT * FROM site_specs WHERE site_id = $1 AND aspect = 'site_plan' ORDER BY created_at DESC;

-- Rollback: restore a previous version
UPDATE site_specs SET is_current = false, superseded_at = NOW()
WHERE site_id = $1 AND aspect = 'site_plan' AND is_current = true;

UPDATE site_specs SET is_current = true, superseded_at = NULL
WHERE id = '<previous_version_id>';
```

Each spec write creates a new row. The old row stays with `is_current = false`. This gives full audit trail — you can see what the classifier originally proposed, what the planner changed, what a human adjusted.

The `source` and `source_agent` columns track who made each version:

```sql
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, source_item_id)
VALUES ($1, 'site_plan', $2, 'build-site-planner', 'build-site-planner', $3);
```

---

## Implementation Order

### Immediate (unblocks audit agents)

1. **Backfill `site_specs`** for finetuning.uk and gaswholesalers.com from `content_data`
2. **Update audit agent queries** to use fallback pattern (site_specs → content_data)
3. **Retrigger audits** — findings will now reference actual spec data

### Short-term (keeps both pipelines working)

4. **Add `write_site_spec` step** to pageflow-builder after planning
5. **Update `read_site_spec`** action with content_data fallback logic
6. **Add `design_intent` and `content_direction` aspects** to build-site-planner output

### Medium-term (unified spec)

7. **Extend classifier** to produce identity, design_intent, content_direction, strategy aspects
8. **Shift planner** to enrichment mode (reads classifier spec, adds component detail)
9. **Add feasibility checking** action (queries agent_definitions + component library)
10. **Migrate all downstream agents** to read from `site_specs` via `read_site_spec`

### Later (deep research)

11. **Integrate content strategy framework** questions into classifier workflow
12. **Add deep research** as classifier enrichment step
13. **Knowledge base integration** for authority-driven content sites

---

## Document Consolidation

This document (013) replaces the need for separate:
- `004_classifier_notes.md` → classifier responsibilities are section 3 here
- The "dream spec" concept from 002d → resolved as "one spec with status fields"

These documents remain as-is (referenced, not replaced):
- `002c_system_architecture_v3.md` → system-wide architecture (add reference to 013 for spec details)
- `002d_quality_assurance_architecture.md` → audit agent hierarchy (add reference to 013 for ground truth)
- `003d_contracts_and_standards_v4.md` → contracts (add query parameterisation rule)
- `002_domain_content_strategy_framework_v2.md` → the 15 questions (referenced by classifier)
- `003_deep_domain_research_authority.md` → knowledge base approach (future classifier enrichment)
- `004_site_work_orchestrator.md` → dispatch loop details (still accurate)
- `009_improvement_loop.md` → discovery/fix cycle (still accurate)
- `001e_development_guide_new_agents_v5.md` → agent development (still accurate)

---

## Resolved Decisions

24. **One spec, not two.** Items have status (deployed/planned/blocked), not separate documents. The "dream" is the full spec. The "build" is the non-blocked subset.

25. **`site_specs` is the authoritative source.** `content_data` is legacy storage. New code reads `site_specs` first with `content_data` fallback. Backfill migrates old sites.

26. **Classifier writes intent, planner writes implementation.** The classifier decides what the site should be. The planner decides how to build it with available components. The design agent implements the visual direction. Audit agents enforce all of it.

27. **Versioning is per-aspect, not per-site.** Changing the design intent doesn't version the site plan. Each aspect has independent history via `is_current` / `superseded_at`.

28. **Backward compatibility via fallback reads.** `read_site_spec` tries `site_specs` then `content_data`. Both pipelines continue working. Migration is gradual.

29. **The content strategy framework feeds the classifier.** The 15 questions become the classifier's research structure. Deep research is a future enrichment that produces richer spec aspects without changing the spec format.
