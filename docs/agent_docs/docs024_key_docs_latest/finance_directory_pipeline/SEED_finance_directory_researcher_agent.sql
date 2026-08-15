-- SEED — finance-directory-researcher agent (Phase B: the finance/insurance
-- provider directories — kinds mortgage-lender, savings-provider,
-- health-insurer; portfolio_positioning lane, DIR-001)
--
-- SIBLING of directory-researcher and adoption-researcher, same 6-step
-- workflow, one agent emitting all three kinds from one prompt (the
-- adoption agent emitting company+protocol is the precedent). The prompt is
-- the whole value: this one's distinctive burden is the NON-PRICE ruling.
--
-- REVISED 2026-08-15 (B4 supervised run 1): extract_claims gained THE SECOND HARD RULE
-- (an entity is ONE NAMED FIRM, never a sector/aggregate) after run 1 registered two
-- category-shaped entities off market-level pages. Applied to the LIVE row by
-- sql_for_agents/423 (snapshot-first); this file updated to match so a re-apply cannot
-- resurrect the gap. Evidence: portfolio_positioning/NOTES 2026-08-15.
--
-- ⚠ ORDERING — DO NOT LET THIS RUN BEFORE THE PHASE B IMAGE IS LIVE.
-- The agent's ACTIONS all ship in the old binary (web_search, prepare_urls,
-- batch_webscrape, execute_llm_prompt, verify_and_register_directory_claims
-- — kind-agnostic since the model/adoption phases), so this seed APPLIES
-- cleanly pre-roll. But the registration-time field allowlist
-- (financeKindFieldAllowlist, directory_claims.go — the MECHANICAL half of
-- the owner's non-price ruling, added at council direction, corr 69a619e6)
-- only exists in the Phase B binary. A research run on the OLD binary would
-- enforce the ruling by prompt alone, which is exactly the
-- instruction-not-a-control gap the council objected to. Hence the companion
-- scheduled-tasks seed ships its three tasks ENABLED=FALSE, and the runbook
-- step that enables them requires the pod-grep for the Phase B log literal
-- first ("per kind (Phase B kind-scoped keys)").
--
-- THE NON-PRICE RULING (owner, 2026-08-12/13; DIR-001): these directories
-- carry durable facts only — FCA firm reference, regulator status, product
-- types, structure. NEVER interest rates, APRs, AERs, premiums, fees or any
-- price-shaped figure: a stale price published under a named FCA-regulated
-- firm is a financial-promotion exposure, where a stale "established 1989"
-- is not. The field vocabulary is CLOSED per kind, enforced twice: in this
-- prompt (instruction) and at registration (control — candidates naming any
-- other field are rejected to human review, never registered).
--
-- SOURCE POSTURE: exclude_domains only, NO prefer_domains — the adoption
-- agent's reasoning transfers directly (a provider's own page is the most
-- INTERESTED source for claims about that provider). The prompt points at
-- the FCA register, trade bodies (ABI, UK Finance, BIBA), gov.uk,
-- MoneyHelper and trade press. Comparison sites this portfolio competes
-- with are excluded outright.
--
-- TIMEOUT GOTCHA (bugs_open/062): timeout_seconds INSIDE the step's config.
-- AI_SERVICE GOTCHA (bugs_open/009): ai_service at STEP level only.
--
-- Invocation:
--   input_data: { research_query }
--   examples:
--     "UK mortgage lenders FCA authorised residential buy-to-let product ranges"
--     "UK savings account providers FSCS protection banks building societies"
--     "UK private medical insurance providers underwriters cover types"

INSERT INTO agent_definitions (
    type, display_name, description, category, agent_category, status, is_active,
    image_repository, image_tag, input_contract, output_contract, default_config
)
SELECT
    'finance-directory-researcher',
    'Finance Directory Researcher',
    'Finance/insurance directory acquisition lane (DIR-001, kinds mortgage-lender / savings-provider / health-insurer): researches UK providers, extracts atomic NON-PRICE candidate claims (FCA firm reference, regulator status, product/cover types, underwriter, established year) with verbatim quotes, and registers only those whose cited source demonstrably contains the quote (verify_and_register_directory_claims). Price-shaped fields are refused at registration by the closed per-kind vocabulary. Unverifiable candidates terminate at human review.',
    'analyst', 'analyst', 'active', true,
    'docker.io/aqls/agent-chassis',
    (SELECT image_tag FROM agent_definitions WHERE type = 'claims-auditor'),
    '{"required": ["research_query"]}'::jsonb,
    '{"produces": {"registration": "verified UK finance/insurance provider claims added to the directory register; rejects raised for human review"}}'::jsonb,
    $cfg${
  "workflow": {
    "start_step": "search_web",
    "processing_mode": "orchestrator",
    "timeout_seconds": 600,
    "steps": {
      "search_web": {
        "action": "web_search",
        "config": {"query_from": "input_data.research_query", "num_results": 10},
        "next_step": "prepare_urls",
        "output_field": "search_results",
        "description": "Search the open web for candidate sources"
      },
      "prepare_urls": {
        "action": "prepare_urls",
        "config": {
          "max_scrapes": 4,
          "max_snippets": 5,
          "exclude_domains": ["pinterest.com", "facebook.com", "twitter.com", "instagram.com", "reddit.com", "moneysupermarket.com", "comparethemarket.com", "gocompare.com", "confused.com", "uswitch.com", "money.co.uk"]
        },
        "next_step": "scrape_pages",
        "output_field": "prepared_urls",
        "description": "Pick the strongest sources. NO prefer_domains, deliberately (the adoption agent's reasoning transfers): a provider's own page is the most INTERESTED source for claims about that provider. The register (register.fca.org.uk), trade bodies (abi.org.uk, ukfinance.org.uk, biba.org.uk), gov.uk, moneyhelper.org.uk and trade press are all legitimate. Comparison sites this portfolio competes with are excluded outright — their listings are the thing being replaced, not a source."
      },
      "scrape_pages": {
        "action": "batch_webscrape",
        "config": {
          "urls_field": "prepared_urls.urls_to_scrape",
          "scrape_config": {"only_main_content": true, "capture_screenshot": false, "formats": ["markdown"]},
          "timeout_seconds": 240
        },
        "next_step": "extract_claims",
        "output_field": "scrape_results",
        "description": "Scrape the selected sources"
      },
      "extract_claims": {
        "action": "execute_llm_prompt",
        "config": {
          "ai_service": {"model": "claude-sonnet-4-6", "provider": "anthropic", "max_tokens": 8000, "api_key_env_var": "ANTHROPIC_API_KEY"},
          "input_fields": ["scrape_results", "prepared_urls", "input_data"],
          "output_format": "json",
          "prompt_template": "You are extracting ATOMIC, CITABLE, NON-PRICE facts about UK FINANCIAL SERVICES PROVIDERS for a public cited directory. Research question: {{.input_data.research_query}}\n\nSCRAPED SOURCES (each has a url):\n{{.scrape_results}}\n\nExtract up to 15 candidate claims, one per (provider, fact) pair. Each MUST be atomic and verifiable:\n- name the specific provider (entity_slug: lowercase-hyphenated, e.g. 'nationwide-building-society'; entity_name: display name; entity_owner: the parent group if the source names one, else same as entity_name);\n- entity_kind: EXACTLY ONE OF \"mortgage-lender\", \"savings-provider\", \"health-insurer\" — whichever the research question concerns. Never any other value.\n- field: ONLY from this CLOSED list, per kind — mortgage-lender: fca_firm_reference, regulator_status, product_types, established_year, lender_type · savings-provider: fca_firm_reference, regulator_status, product_types, established_year, protection_scheme · health-insurer: fca_firm_reference, regulator_status, cover_types, established_year, underwriter. Any other field will be MACHINE-REFUSED at registration — do not emit one.\n- a VERBATIM quote copied EXACTLY from ONE source's text that states it — do not paraphrase, do not stitch sentences, do not normalise numbers or names;\n- the url of that exact source page (from the material above — never invent or shorten one).\n\nEvery claim will be machine-checked twice: the field against the closed vocabulary, and the url re-fetched with your quote required verbatim. A paraphrase fails. A summary fails.\n\nTHE RULE THAT MATTERS MOST — NO PRICES, EVER. Do not extract interest rates, APRs, AERs, premiums, fees, LTV percentages, or any figure a customer would compare on. This directory records WHO a provider is (regulated by whom, offering what product classes, structured how), never what anything costs today. A price in this register would go stale and mislead under a named regulated firm's name — that is a compliance exposure, and it is why the field list above contains no price-shaped field. If a source's best content is pricing, skip it.\n\nTHE SECOND HARD RULE — AN ENTITY IS ONE NAMED FIRM. Every entity_slug / entity_name must be a single company or brand that holds (or could hold) its own FCA authorisation. Never register a sector, market segment, product category, regulator, statistics series or any other aggregate as an entity — shapes like 'UK specialist lenders', 'FCA-regulated mortgage lenders (general)' or 'the later life mortgage market' are wrong even when the facts about them are true and citable, and a human reviewer will reject the set for them. A market study, trade-body overview or statistics page may only yield claims about the individual firms it NAMES; if a source discusses the market solely in aggregate, extract nothing from it.\n\nALSO:\n1. WHO IS SPEAKING. The FCA register, a trade body, or the provider's own about/regulatory page are all fine sources for these durable facts (unlike results claims, a firm's own statement of its FCA reference is authoritative). Third-party listicles are weak — prefer the register or the firm.\n2. regulator_status is the source's own words (e.g. 'authorised and regulated by the Financial Conduct Authority', with the FRN if stated in the same sentence). fca_firm_reference is the bare reference number, only when the source states it explicitly.\n3. product_types / cover_types are the source's own enumeration (e.g. 'residential, buy-to-let and later-life mortgages'), never your synthesis across pages.\n4. NEVER COPY A VALUE FROM THESE INSTRUCTIONS — they illustrate shape, not data.\n\nOptionally, on the FIRST claim you emit for a given entity, also include entity_summary (one plain sentence: what the provider is), entity_links (object: docs [the provider's own regulatory/about page]), and entity_attributes (object: sector e.g. \"building-society\"|\"bank\"|\"insurer\", region — coarse filing tags, not claims). Omit rather than guess.\n\nReturn ONLY a JSON array (no commentary). Each element:\n{\"entity_kind\": \"mortgage-lender\"|\"savings-provider\"|\"health-insurer\", \"entity_slug\": \"...\", \"entity_name\": \"...\", \"entity_owner\": \"...\", \"entity_summary\": \"...\" (optional), \"entity_links\": {...} (optional), \"entity_attributes\": {...} (optional), \"field\": \"...\", \"value\": \"...\", \"unit\": \"...\", \"quote\": \"verbatim sentence from the source\", \"url\": \"...\", \"publisher\": \"...\", \"title\": \"...\", \"published\": \"YYYY-MM or YYYY if shown\", \"staleness_days\": <400 for fca_firm_reference, regulator_status, established_year, lender_type, underwriter, protection_scheme — structural facts; 200 for product_types and cover_types, which drift>}\n\nIf nothing extractable meets the bar, return []."
        },
        "next_step": "verify_and_register",
        "output_field": "candidate_claims",
        "description": "Extract atomic non-price provider claims with verbatim quotes — the checkable form"
      },
      "verify_and_register": {
        "action": "verify_and_register_directory_claims",
        "config": {
          "candidates": "candidate_claims.result"
        },
        "next_step": "complete",
        "output_field": "registration",
        "description": "The disposal step: closed-vocabulary check, then re-fetch every cited url; register only quotes that are really there; rejects to human review under directory_citation_unverified:<kind>"
      },
      "complete": {
        "action": "complete_workflow",
        "config": {"output_fields": ["registration"]}
      }
    }
  }
}$cfg$::jsonb
WHERE NOT EXISTS (SELECT 1 FROM agent_definitions WHERE type = 'finance-directory-researcher' AND deleted_at IS NULL)
RETURNING type, status;

-- ── Post-apply verification ────────────────────────────────────────────────
-- 1. Row exists:
--      SELECT type, status FROM agent_definitions WHERE type='finance-directory-researcher';
-- 2. DO NOT smoke-run until the Phase B binary is live (see header). Gate:
--      kubectl -n ai-persona-system exec <chassis-pod> -- grep -aq "per kind (Phase B kind-scoped keys)" /proc/1/exe
--    with a control grep for an absent string in the same breath.
-- 3. First run, what to CHECK rather than assume:
--    - rejects land as directory_citation_unverified:<kind> (kind-scoped), not the bare legacy key;
--    - NO price-shaped field appears even as a REJECT with class citation_invalid mentioning the
--      closed vocabulary — some will; that is the control working, not a failure;
--    - SELECT de.kind, de.name, dc.field, dc.value, dc.status FROM directory_claims dc
--      JOIN directory_entities de ON de.id=dc.entity_id
--      WHERE de.kind IN ('mortgage-lender','savings-provider','health-insurer') AND dc.is_current
--      ORDER BY de.kind, de.name, dc.field;
