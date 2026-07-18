# HANDOFF — vetcomparison.uk (fresh thread starts here)

**Date:** 2026-07-18. **Predecessor:** HANDOFF_2026-05-18 (historical). Read this file, then
SUMMARY_2026-07-18_readout.md (narrative), PLAN_2026-07-15_rebuild.md (phases, all statuses
current), RUNBOOK_vetcomparison.md (operator steps), RUNNING_NOTES (log, newest first),
LEGAL_2026-07-15 (factual record — update it whenever publication-relevant facts change).
**Obey repo CLAUDE.md**: concurrent sessions; pathspec commits per task; forward-only; builds
from committed HEAD; verify deploys against the pod; queue-check before dispatch.

## Current state (all pod/live-verified 2026-07-18)

- **Site LIVE and chassis-managed**: adopted 2026-07-17 (fidelity locked recorded), full build
  cascade ran autonomously (strategy→plan→composition→design→imagery→pages→rerenders).
  Site row active; classification site_type=hub.
- **Exporter LIVE**: `directory_export_json` action deployed (pod-verified in v1.0.1134);
  agents `directory-json-exporter`/`directory-export-orchestrator` at image_tag v1.0.1134;
  scheduled task `directory-export-json` ENABLED, 48h. First autonomous publish `ac3314fd`:
  directory 2,109 / aggregates 13 rows (min_n 3, n shown) / claimed `[]` / attributed `[]` /
  metadata. All publication rules held.
- **Data**: 2,109 verified deduplicated practices (dedupe keys: website host+postcode, then
  name+postcode; keep-rule has_prices > latest verified > shortest URL). 176 real practices
  parked `pending` (wheree.com mirror websites). 997 fabricated price rows quarantined
  (`source='seed_import'`, never publish). Historical prices have EMPTY source URLs →
  aggregates only, forever. CMA 36-item taxonomy seeded (`cma_item` products, 12/6/6/9/3).
- **Claim flow proven** (rolled-back prod dry-run): claim_requests table (consent snapshot),
  claimed supersedes scraped, opt-out hides prices but keeps directory+aggregates, claim
  reverses opt-out. Operator SQL in RUNBOOK.

## ⚠️ TOP PRIORITY — the chassis REGENERATED FABRICATED PRACTICE DATA (contained, not fixed)

**An earlier version of this handoff said the rebuild "dropped the search UI". That was wrong —
see SUMMARY_2026-07-18_bugs_journey.md §9-10.** The truth, found by checking the rendered page
rather than work-item status (CLAUDE.md: *trust the rendered artefact, not the status*):

The adoption rebuild's `needs_tool_recreation` (status: **complete**) recreated our practice
search as a component that **generates synthetic practices client-side** — its own comment:
*"The original directory holds 2,100+ UK practices. For this recreation we generate a large,
realistic, deterministic dataset"* — assembling fake names from PREFIXES×SUFFIXES arrays and
inventing postcodes via a Mulberry32 seeded RNG, then calling `render()`. **Live visitors were
shown fabricated veterinary practices**: the exact defect this site was remediated for on 14-15
July, reintroduced autonomously by our own tooling.
The same rebuild added unsupported claims: "pricing information, ownership data" (we publish
neither), "Price: Low to High" sort controls (no published prices), and called our real
2,109-practice directory "a representative sample for demonstration purposes".

**CONTAINED 2026-07-18**: verified homepage restored (commit on sites/master, restored from
`b2896815`), live-verified clean — 0 generator, 0 unsupported claims, real
`/data/vet-full-index.json` wired, claim + opt-out routes back. Guides were unaffected
throughout (sourced content, review stamps intact); no unsourced prices anywhere on the site.

**SOURCE FIXED 2026-07-18 — one thing left to verify:**
1. ✅ **Fixed at source, not just in the file.** The fabrication was still in
   `page_components.rendered_html` (deployed, unlocked) — the generator lived in the **`hero`**
   slot, not `filtered-result-grid`. Hero's data layer rewritten to `fetch('/data/vet-full-
   index.json')` keeping the chassis's better UI (region filter, pagination); demo-sample
   disclaimer, price-sort controls, "pricing information / ownership data" claims and a false
   about-page "we distinguish independent practices" differentiator all removed. Four components
   now `lock_type='permanent'`. Two other hits (about/faq, guide-cma-market-investigation) were
   checked and are ACCURATE statements about CMA findings — leave them; a regex sweep would
   wrongly "fix" correct content.
   ⚠️ **NOT PROVEN: no render has been run against the fixed source.** Manual dispatch failed
   (`rerender-pages` is `experimental`; neither site-builder.requests nor
   page-rerender.process produced an orchestration state from kcat). **Watch the first natural
   render and diff the homepage.** Verification one-liner:
   `curl -s https://vetcomparison.uk/ | grep -ciE 'Mulberry32|makePostcode|representative sample'`
   must be 0, and `grep -c vet-full-index` must be ≥1.
   Note `page_components.data_path` is empty fleet-wide — vestigial, do not build on it.
2. **Platform bug FILED as `/bugs_open/020`** (2026-07-18, commit `4e372119b`) — read it before
   touching the recreation path; pattern + grep tells also in 016b §9, bug index row 020.
   Root cause is structural, in two parts: (a) the recreation path has no **data-dependency
   contract** — adoption's `extract_interactive_fingerprint` never passes the original tool's
   `fetch()` target to `tool-recreation-handler`, so a data-backed tool cannot be rebuilt
   faithfully; (b) that agent's prompt rule 9 ("No fake data or dummy outputs — calculations
   must be mathematically correct") is **scoped to arithmetic**, so it does not forbid inventing
   records. Fix candidates ranked in 020; (1)+(2) are structural, (3) a cheap grep gate.
   Sibling case `001`. Fixing this is a platform job, not a vetcomparison job — but this site is
   the reason it is urgent, since fabrication here is the exact defect we remediated.
3. **7 HITL items** need the owner in the admin UI (see "Owner review queue" below).

## Owner review queue (7 items, admin UI)

Needs the owner's judgement:
- `tool-compliance-deadline-calculator` (owned_page_review) — chassis proposes an interactive
  CMA-deadline calculator; unbuilt, 0 sections. Good idea on merits (drives claims) but owner
  decides. **If built, it must not invent data** (see top priority above).
- Three planned pages with **0 sections**, so the builder correctly refused: `directory-index`,
  `guides-index`, `practice`. Decide purpose or cancel. `directory-index` is the interesting
  one — its existence suggests the planner wanted the directory on its own page, which bears on
  where the search lives.
Needs a fact only the owner has:
- `needs_section_data` on contact — identity spec has no email/phone. Site already publishes
  `vetcomparison@contactforsales.com` in claim/opt-out links; confirm that (and whether to
  publish a phone).
Technical, next session can clear:
- `needs_page` about — **failed**, "claim timed out, handler pod likely died" (transient;
  about.html is live). Retry or cancel.
- `needs_page` about re-render — validate_content "1 blockers"; read the blocker.

## Open items, priority order

1. Homepage restoration (above) + owner works the HITL review queue.
2. **Funding consultation closes 30 Jul 23:59** — draft ready:
   CONSULTATION_RESPONSE_funding_DRAFT_2026-07-16.md (owner verifies levy ¶ in the Notice PDF,
   submits via connect.cma.gov.uk portal). Case team VetsMI@cma.gov.uk.
3. **Watch the CMA case page** for the substantive draft Order (expected any day; Order due
   23 Sep). When it lands: clause-referenced response from CONSULTATION_2026-07-16_briefing.md
   §3 positions (pro-independent + express reuse right + machine-readable + no
   selective-blocking). Nothing submitted without owner sign-off.
4. **Phase 5 — provenance-first price scraping** (the big build item): re-verify the 176
   wheree practices; scrape practice price pages persisting per-price `source_url` +
   `observed_at` (without which output is unpublishable — exporter enforces); extend
   sweep/verifier deny-list (wheree.com, bestlocalrated, yelp.*, starofservice, threebestrated,
   allvets.co.uk, calmshops, rated.club, digifarm.uk); normalise website host before upsert
   (dedupe root cause). Post-Order (~Dec 2026) the mandated lists make this trivial +
   compliance-watch becomes the claim funnel.
5. Solicitor review (LEGAL §8): the factual record + database-right position. Owner decision
   2026-07-16: attributed publication proceeds meanwhile under narrowing conditions.
6. Optional: classifier `content_features` one-off spec patch if a news feed is wanted (lost
   005 patch never landed).

## Hard rules (do not relax)

- **No price without provenance or consent.** seed_import rows never publish. We do not own
  vetcomparison.co.uk — no default domains anywhere (fail-closed enforced in Go).
- **Never re-plan this site to fill gaps** (first-plan branch is the only faithful path;
  re-plans clobber). Hand edits → permanent locks.
- Sites repo deploys: detached worktree off origin/master, verify and push as SEPARATE steps,
  rebase-retry on bot races, never force-push, never touch the stale dirty local clone.
- Aggregates: min_n 3, always publish n. Claimed supersedes scraped. Opt-out promptly.

## Pointers

- DB: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`
  (password: secret postgres-clients-secret). Tables: business_intel.{businesses, products,
  product_prices, claim_requests, business_prices(deprecated)}; sites/site_specs/site_work_items.
- Migrations 006–010 in this dir, all applied, all idempotent.
- Exporter config lives in scheduled_tasks 'directory-export-json' input_data (attributed ON,
  min_n 3, filename vet-full-index.json).
- Adoption re-trigger (if ever needed): `bash 082_submit_domain_unified.sh vetcomparison.uk
  --from https://vetcomparison.uk --fidelity locked`.
- Live checks with curl (python urllib → Cloudflare 403); mailto bodies are URL-encoded
  (grep accordingly).
