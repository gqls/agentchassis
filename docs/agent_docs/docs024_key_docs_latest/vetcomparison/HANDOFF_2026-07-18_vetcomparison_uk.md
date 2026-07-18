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

## ⚠️ TOP PRIORITY — homepage regression from the chassis rebuild

The first-build rerender replaced the hand-authored homepage. Verified 2026-07-18:
- **Directory search UI GONE from the homepage** (`#vet-list`, search inputs absent) — the
  site's core function. Data files all still live under `/data/`; only the page markup lost it.
- **Claim CTA + opt-out routes GONE** from the homepage (they are the licensing funnel and a
  policy promise — LEGAL §7.5 says practices can request removal via a published route, so
  their absence is a live compliance gap with our own policy).
- Guides SURVIVED at original flat URLs with sourced content intact (last-reviewed stamps,
  sources lists). New about/contact pages exist; £21/£12.50 on about.html are CMA's own
  figures. **No unsourced prices anywhere** (audited all pages).
Fix path: hand-edit the rebuilt homepage to restore the directory component + claim/opt-out
section (hand edits during the adoption window get PERMANENT locks — the right protection), via
the sites-repo worktree pattern (RUNBOOK). Note the chassis renders pages from specs — check
`pages` table/spec for where the homepage sections live so the fix is spec-level if possible,
not just HTML (else the next rerender may revert it: `source: "pages.*"` fields revert on
render — fleet landmine).
Also pending: **7 HITL items** need the owner in the admin UI: 4 needs_page + 1
owned_page_review + 1 needs_section_data in needs_human_review, 1 needs_page failed.

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
