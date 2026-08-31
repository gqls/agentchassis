# 417 — the site planner's logo exemplar licenses a wordmark it never names, so the image model invents a brand ("Farm Shield Info" on farmerinsurance.uk)

**Filed 2026-08-31 by the loanzy_uk_example_site lane, from the owner's review of
farmerinsurance.uk.** Diagnosis loop NOT run — substituted equivalent first-hand
verification, stated per the 2026-07-31 owner ruling: every link in the causal chain below
is a read of the live row/artefact, each quoted, and the chain has no inferential hop.
(Owner-visible harm: a served UK insurance site whose logo carries **someone else's brand
name** — a trademark problem, not a cosmetic one.)

## The chain, each link verified 2026-08-31

1. **The exemplar** — `agent_definitions` row `build-site-planner` (live, is_active), worked
   example in `default_config`:
   `"prompt": "A precise, technical logomark — geometric, restrained, no human figures, no text outside the wordmark itself"`
   It PERMITS a wordmark and names no brand. A quoted exemplar ships verbatim-in-shape
   (memory class: *a QUOTED exemplar in a prompt is copied verbatim*).
2. **The paraphrase** — farmer's `needs_imagery` item (site `99cae989…`, item_key
   `needs_imagery:site:-:logo`, created_by `build-site-planner`): prompt ends *"no
   photorealism, **no text beyond the wordmark**, balanced proportions…"* — the exemplar's
   clause, re-worded, still naming no brand. `identity.company_name` ("Farmer Insurance
   UK") is in the identity spec and is NOT in the prompt.
3. **The generation** — `assets` row `a88c0e99…`, `origin_model
   banana/gemini-3-pro-image-preview`, `origin_prompt` = the paraphrase. Licensed to
   letter a wordmark with no text specified, the model invented one: **"Farm Shield
   Info"** (plausibly compressed from "farm … shield … information site" in the prompt).
4. **The serving** — `/assets/images/logo.png` live in the header (sole brand mark; no
   HTML text wordmark beside it), and the favicon + og_card were DERIVED from it, so the
   invented brand is stamped at three surfaces.

## Why this is a class, not a one-off

- The estate already RULED on this: `platform/orchestration/actions/discovery_checks/
  default_brand_prompt.go:231` — the fallback logo prompt says **"no lettering or
  words"**, with the rationale in-file: *"generated wordmarks reliably produce malformed
  text, and this asset is used at favicon size."* The planner's exemplar contradicts the
  estate's own craft rule; two prompt sources, one rule in one of them. Every planner-built
  site takes the exemplar path; the ruled path is the fallback nobody reaches.
- No seat reads pixels: nothing verifies rendered logo TEXT against
  `identity.company_name` (the acceptance council's gap #2 in the farmer review — see
  OWNER_REVIEW_2026-08-31 in the loanzy lane). The defect is silent until a human looks.

## Fix candidates (ordered by what closes the door)

1. **Fix the exemplar** (one migration on `build-site-planner`): either align with the
   ruled default — "no lettering or words" — or, if the owner wants wordmark logos, make
   the exemplar demand the exact brand string: *"the only text is the exact wordmark
   ‹company_name›"* with the planner instructed to substitute identity.company_name.
   Unnamed-wordmark becomes unrepresentable. **Council scope** (agent config migration).
2. **Belt**: a logo-text check (OCR-shaped, or a cheap model look) comparing rendered
   text to identity — the unowned designer-family gap; routed to the imagery/designer
   threads, not built here.
3. Farmer's instance: regeneration item ALREADY FILED through the framework
   (`3740f5f2-e6ff-4dd3-b60d-8ba502b1c636`, prompt names the exact wordmark, positive-
   framed because `bugs_closed/028` proved banana discards negative clauses). Favicon +
   og_card re-derivation owed AFTER it lands (`derive_brand_head_assets`) — presence-based
   discovery will NOT refile them on its own.

## Verify
- Exemplar: `SELECT substring(default_config::text from position('logomark' in default_config::text) for 120) FROM agent_definitions WHERE type='build-site-planner' AND is_active AND COALESCE(is_snapshot,false)=false;`
- The prompt that shipped: `SELECT origin_prompt FROM assets WHERE id='a88c0e99-6de9-4b7d-996d-3c16d530c8a8';`
- The ruled default: read `default_brand_prompt.go` header + :231.
- Related: 210 (needs_logo unhandleable), 235 (logo stored as hero), 322 (brand-head block page-blind), closed 028 (negative prompts discarded).
