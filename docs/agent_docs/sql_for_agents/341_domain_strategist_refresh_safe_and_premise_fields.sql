-- 341_domain_strategist_refresh_safe_and_premise_fields.sql
--
-- vigilant_designer_offer_analysis programme, Phase B2 (PLAN 2026-08-02; owner
-- decision 2026-08-08: B1+B2 jump the queue — see the lane PLAN's decision log
-- and features_open/030).
--
-- THE DEFECT: domain-strategist's workflow (read_specs -> analyze_strategy ->
-- write_strategy_spec -> create_next_item -> complete) has NO conditional
-- anywhere. create_next_item UNCONDITIONALLY files needs_briefing ->
-- build-briefing-agent, which files needs_site_plan -> build-site-planner
-- (all three links read from live config 2026-08-08). On greenfield that chain
-- is the point, and it is the only way it has ever run (3/3 historical
-- needs_briefing rows). On a DEPLOYED site, a premise refresh would re-plan the
-- live site as a side effect — which is why nobody dares refresh one, and why
-- B3's check_premise_incomplete cannot ever enable before this gate exists.
--
-- THE CHANGE, config only, live on apply:
--   1. NEW step check_site_deployed (query_database): is_deployed :=
--      count(pages WHERE build_status='deployed') > 0. DB state ONLY — never an
--      input_data.spec.* path (a missing spec key fails input_mapping: the dead
--      needs_logo path is the worked example), and never sites.status
--      (loanandmortgagecalculator is 'active' with 41 deployed pages — measured
--      2026-08-08; status lies).
--   2. NEW step gate_next_item (conditional_branch — the registered name;
--      "conditional" is a deprecated alias of the same handler, registry.go:71):
--      deployed -> complete (chain suppressed); else -> create_next_item
--      (greenfield behaviour byte-identical, same step, same config).
--      write_strategy_spec.next_step repoints to check_site_deployed.
--      The complete step's output_fields list next_item_created, which the gated
--      path never writes: extractWorkflowResult's ResultModeFields SKIPS missing
--      fields (coordinator.go:3757) — verified, not assumed.
--   3. analyze_strategy's output schema gains the four premise fields. This is a
--      RESTORATION, not an invention: gaswholesalers.com's strategy row has
--      carried satisfaction_condition / trust_threshold / recurring_value (plus
--      visitor_type/primary_intent) since 2026-04-17 — the sole survivor of the
--      older, better shape; the other 16 sites carry the 12-key shape with no
--      premise fields. Live spellings reused; money_flow is PLAN B2's name for
--      the fourth. Plus a refresh-preserves instruction: read_specs loads ALL
--      current aspects, so an existing strategy row IS in the model's context on
--      a refresh; and write_site_spec DEEP-MERGES (site_spec_actions.go), so an
--      omitted key structurally survives — the instruction is belt, the merge is
--      braces.
--
-- CONSUMERS TOLD (owner ruling 2026-07-29 3): the existing producer of
-- needs_strategy is vertical-exemplar-researcher (3 rows, all greenfield).
-- Greenfield behaviour is unchanged by construction; the guarantee that CHANGED
-- is "running domain-strategist always chains a build" -> "chains a build only
-- on a site with no deployed pages". CONTRIB note filed with the
-- portfolio_positioning lane (premise->writer wiring owners).
--
-- ROLLBACK: restore from agent_definitions_backup (two-arg snapshot below;
-- NOT an is_snapshot row — LANDMINES 2026-07-30), picking the row by
-- snapshot_taken_at DESC with snapshot_reason LIKE 'pre-update: B2%'.
--
-- VERIFY (after apply): the B2 witness is a hand-filed needs_strategy for
-- loancalculator.co.uk (0162cde4-…, 27 deployed pages, NO strategy aspect):
-- orchestration completes, a strategy row with revenue_models.primary_model AND
-- a premise field exists, and ZERO needs_briefing rows appear (row identity, not
-- count). See the lane PLAN_2026-08-08_B1_B2_premise_first.md.

-- Probe guard: refuse a second application.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM agent_definitions
        WHERE type = 'domain-strategist'
          AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
          AND default_config #> '{workflow,steps,gate_next_item}' IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'B2/341: already applied — gate_next_item exists';
    END IF;
END $$;

-- Drift guard: composed against the exact live texts fetched 2026-08-08 ~17:00Z.
-- If another session has since changed the step or prompt, REFUSE and recompose.
DO $$
DECLARE
    p_md5 text;
    nx text;
BEGIN
    SELECT md5(default_config #>> '{workflow,steps,analyze_strategy,config,prompt_template}'),
           default_config #>> '{workflow,steps,write_strategy_spec,next_step}'
      INTO p_md5, nx
      FROM agent_definitions
     WHERE type = 'domain-strategist'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF p_md5 IS DISTINCT FROM '489d24564437bb85e574f12726c2b370'
       OR nx IS DISTINCT FROM 'create_next_item' THEN
        RAISE EXCEPTION 'B2/341: DRIFT — live config differs from what this migration was composed against (prompt %, next_step %). Recompose.', p_md5, nx;
    END IF;
END $$;

BEGIN;

SELECT snapshot_agent('domain-strategist',
    'pre-update: B2 (vigilant_designer_offer_analysis) — refresh gate + premise fields');

UPDATE agent_definitions
SET default_config =
    jsonb_set(
    jsonb_set(
    jsonb_set(
    jsonb_set(
        default_config,
        '{workflow,steps,check_site_deployed}',
        $step$
        {
            "action": "query_database",
            "config": {
                "query": "SELECT (COUNT(*) > 0) AS is_deployed FROM pages WHERE site_id = $1 AND build_status = 'deployed'",
                "params": ["input_data.site_id"],
                "output_format": "object"
            },
            "next_step": "gate_next_item",
            "description": "B2 gate input: deployed = any deployed pages (never sites.status, which lies: 'active' with 41 deployed pages measured 2026-08-08)",
            "output_field": "site_state"
        }
        $step$::jsonb
    ),
        '{workflow,steps,gate_next_item}',
        $step$
        {
            "action": "conditional_branch",
            "config": {
                "condition": "site_state.is_deployed == true",
                "then_step": "complete",
                "else_step": "create_next_item"
            },
            "description": "B2 refresh-safety gate: a deployed site's strategy refresh must NOT enqueue the briefing->site-plan rebuild chain; greenfield path unchanged"
        }
        $step$::jsonb
    ),
        '{workflow,steps,write_strategy_spec,next_step}',
        to_jsonb('check_site_deployed'::text)
    ),
        '{workflow,steps,analyze_strategy,config,prompt_template}',
        to_jsonb($prompt$You are a domain strategist. Your job is to determine the best website strategy for a domain name. You determine WHAT kind of site to build and WHY. You do NOT design the page architecture — a separate planner agent handles that.

Domain: {{.input_data.domain}}

## Research Data
{{.site_specs}}

## Your Task

Think carefully about this domain. Work through each section below IN ORDER.

### 1. Domain Name Analysis

Classify the domain name:
- **company_brand**: A specific business name (acmeplumbing.com, smithandjoneslaw.co.uk). The site should represent THAT business.
- **generic_industry**: A generic industry or service term (gaswholesalers.com, londonplumbers.co.uk). Much higher value as an authority/directory — pretending to BE a single gas wholesaler wastes the domain.
- **geographic_service**: A location + service combination (manchesterelectricians.com). Ideal for a local directory or lead generation.
- **product_category**: A product type (standingdesks.com, bestcoffeegrinders.com). Good for reviews, comparisons, or affiliate content.
- **ambiguous**: Could be either a brand or a generic term. Explain why and make a judgment call.

### 2. Search Intent

Who would search for these keywords?
- What are they looking for?
- Commercial intent (ready to buy/hire) or informational (researching)?
- What are the likely high-value search terms?

### 3. Revenue Model Assessment

Rate each model for THIS domain: strong_fit / possible / poor_fit.

- **lead_generation**: Capture enquiries, sell leads to businesses in the industry
- **affiliate**: Link to businesses with affiliate programs, earn commission
- **display_advertising**: Build traffic via content/SEO, monetise with ads
- **sponsored_listings**: Businesses pay for premium placement in a directory
- **direct_business**: Domain represents an actual business, revenue from the business itself
- **saas_tools**: Provide a useful tool, monetise via premium features or leads

Pick ONE primary and up to TWO secondary models.

Then answer the four PREMISE QUESTIONS for the chosen primary model — one plain, concrete sentence each, specific to THIS domain, no hedging:
- **satisfaction_condition**: what a visitor must have achieved for the visit to have been worth their while
- **money_flow**: how money actually reaches us — who pays, for what, roughly when
- **recurring_value**: what brings a visitor back — the thing that changes, accrues or updates
- **trust_threshold**: how much trust the visitor needs before acting, and why

### 4. Competitive Positioning

Based on the research:
- Who currently ranks for these keywords?
- What kind of sites are they?
- What gap exists that this domain could fill?

### 5. Site Type Recommendation

Choose EXACTLY ONE from this canonical list:

| Site Type | When to use |
|-----------|-------------|
| `brochure` | Domain represents a specific business — showcase their services/products |
| `authority-portal` | Generic industry domain — be THE resource with directory + editorial content |
| `local-directory` | Geographic + service domain — local service directory with listings |
| `review-site` | Product category domain — reviews, comparisons, affiliate content |
| `content-hub` | Informational domain — blog/magazine style, articles as primary content |
| `landing-page` | Specific product/offer — single high-conversion page |
| `portfolio` | Creative/agency domain — showcase of work/projects |

### 6. Page Type Recommendations

Based on the strategy, recommend which PAGE TYPES the planner should consider. These are recommendations, not a page plan — the planner decides the actual pages.

Choose from this canonical list:

| Page Type | Description |
|-----------|-------------|
| `content` | Standard content page (about, services, contact, etc) |
| `index` | Home page |
| `landing` | Conversion-focused page |
| `entity-directory` | Searchable/filterable directory of entities |
| `entity-page` | Individual entity profile page |
| `tool` | Interactive tool or calculator |
| `blog-index` | Blog/news listing page |
| `blog-post` | Individual blog article |

For each recommended page type, explain WHY the strategy calls for it.

### 7. Content & Tone

What tone suits this site? What kind of content draws the target audience?

Return your analysis as JSON:
```json
{
  "domain_type": "company_brand|generic_industry|geographic_service|product_category|ambiguous",
  "domain_type_reasoning": "why this classification",
  "search_intent": {
    "primary_intent": "commercial|informational|navigational",
    "likely_searches": ["search term 1", "search term 2"],
    "high_value_terms": ["term with commercial value"]
  },
  "revenue_models": {
    "lead_generation": {"fit": "strong_fit|possible|poor_fit", "reasoning": "..."},
    "affiliate": {"fit": "...", "reasoning": "..."},
    "display_advertising": {"fit": "...", "reasoning": "..."},
    "sponsored_listings": {"fit": "...", "reasoning": "..."},
    "direct_business": {"fit": "...", "reasoning": "..."},
    "saas_tools": {"fit": "...", "reasoning": "..."},
    "primary_model": "one of the above keys",
    "secondary_models": ["key1"]
  },
  "competitive_position": {
    "current_landscape": "brief description",
    "gap_opportunity": "what gap this site fills",
    "defensible_moat": "why competitors cant easily replicate"
  },
  "site_type": "one of the canonical site types",
  "site_type_reasoning": "why this site type was chosen",
  "recommended_page_types": [
    {"page_type": "entity-directory", "reasoning": "core to the authority portal strategy"},
    {"page_type": "content", "reasoning": "about page, contact, services overview"},
    {"page_type": "blog-index", "reasoning": "editorial content for SEO traffic"}
  ],
  "tone": "professional|friendly|authoritative|editorial|technical|bold",
  "content_strategy": "what content draws visitors and keeps them coming back",
  "growth_path": "how the site scales over time",
  "value_proposition": "one sentence describing what this site offers visitors",
  "satisfaction_condition": "what a visitor must have achieved for the visit to have been worth their while",
  "money_flow": "how money actually reaches us: who pays, for what, roughly when",
  "recurring_value": "what brings a visitor back",
  "trust_threshold": "how much trust the visitor needs before acting, and why"
}
```

Return ONLY valid JSON.

## Vertical Landscape
If the Research Data includes a `vertical_landscape` aspect (best-of-niche exemplar research), weigh it heavily in section 4 (Competitive Positioning) and section 6 (Page Type Recommendations): build on the success factors and lessons it records — reasons, not copies — and exploit the differentiation_opportunity it names.

## Existing Strategy (refresh runs)
If the Research Data already contains a `strategy` aspect, this is a REFRESH: treat the existing strategy as prior work by a colleague. Revise what the evidence contradicts, sharpen what is vague, and fill what is missing — especially the four premise questions, which older strategies lack. Do NOT reinvent a coherent strategy from scratch, and restate conclusions you mean to keep (your output is merged over the existing record; keeping a conclusion means saying it again).$prompt$::text)
    )
WHERE type = 'domain-strategist'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Verify inside the transaction.
DO $$
DECLARE
    n int;
BEGIN
    SELECT count(*) INTO n
      FROM agent_definitions
     WHERE type = 'domain-strategist'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
       AND default_config #>> '{workflow,steps,write_strategy_spec,next_step}' = 'check_site_deployed'
       AND default_config #>> '{workflow,steps,check_site_deployed,next_step}' = 'gate_next_item'
       AND default_config #>> '{workflow,steps,gate_next_item,config,condition}' = 'site_state.is_deployed == true'
       AND default_config #>> '{workflow,steps,gate_next_item,config,else_step}' = 'create_next_item'
       AND default_config #>> '{workflow,steps,analyze_strategy,config,prompt_template}'
           LIKE '%satisfaction_condition%'
       AND default_config #>> '{workflow,steps,analyze_strategy,config,prompt_template}'
           LIKE '%Existing Strategy (refresh runs)%';
    IF n <> 1 THEN
        RAISE EXCEPTION 'B2/341 VERIFY FAILED: expected exactly 1 updated active row, found %', n;
    END IF;
    -- The snapshot must hold the PRE-change chain or it restores nothing.
    IF NOT EXISTS (
        SELECT 1 FROM agent_definitions_backup
        WHERE type = 'domain-strategist'
          AND snapshot_reason LIKE 'pre-update: B2%'
          AND default_config #>> '{workflow,steps,write_strategy_spec,next_step}' = 'create_next_item'
          AND default_config #> '{workflow,steps,gate_next_item}' IS NULL
    ) THEN
        RAISE EXCEPTION 'B2/341 VERIFY FAILED: no backup row carrying the PRE-change chain';
    END IF;
END $$;

COMMIT;
