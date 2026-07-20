# 005 FEATURE — pilot onboarding + first pool activation

**Raised:** 2026-07-20, parked by the owner: *"we're not quite ready to onboard
more domains."*
**Status: ON HOLD — do not onboard. Nothing here starts without an explicit
owner go.** Everything upstream is done and waiting; this file is the handoff
for whichever thread runs it when the hold lifts.
**Related:** `docs024_key_docs_latest/news_feed_pooling/PLAN_2026-07-19_*.md`
(Decisions 1–21), `RESEARCH_2026-07-20_dormant_domain_history.md`,
`features_open/002` (the gate), `features_open/004` (the seat).

## What is already in place (verified live, 2026-07-20)

- 17 pool synthetic sites (`pool-<slug>.internal`, `status='pool'`), each with a
  pool-default `audience.v1` profile. **Structurally inert**: no classification
  specs, no sources, invisible to every fleet loop. Safety invariants in the
  workstream RUNBOOK (both must read 0).
- All 11 live sites carry current `audience.v1` profiles.
- Pilot cohort chosen: **~37 traffic-bearing domains** across 14 pools
  (list: PLAN Decisions 11 + 19 + 20).

## The three steps, in order, when the hold lifts

### 1. Pilot onboarding (~37 domains → classified sites with profiles)

- Onboard through the normal chain (`domain-research-classifier` →
  identity/classification/strategy). Audience profiles are written **at**
  onboarding from the classifier's research — **never from the domain name
  alone** (PLAN Decision 4; the binding authoring rules are PLAN Decision 15).
- Each member site's profile forks its pool's default; `position` gets written
  at fork for any site with a live sibling (the makeitaquote /
  memecreator.co.uk / memegenerator.uk trio is the known second cluster).
- Re-add the dropped "who is your target audience?" question to the live
  briefing questionnaire (`026_pageflow_builder.sql:868` dropped it; the backups
  at `000_agent_definitions_backup_070_refactor.sql:855` have the section shape).
- Special cases inside the cohort, each with its own directive (PLAN D16–20):
  nanangmrk = **adoption not build** (owner-run live site — separate workstream);
  buysportskit / smartbusinesssupplies / outfax = retailer-directory utilities
  (D17); makeitaquote = differentiated tool build (D18); komunikatif =
  Indonesian-language rebuild with dedicated sources (D20). zdec is **not** in
  the pilot — measure first via the relojistas traffic-probe mechanism
  (CF real-ip prerequisite).

### 2. First pool armed — ONE pool only

- Curate real sources for a single pool (candidate: whichever pool the largest
  onboarded sub-cohort lands in; `savings-investing` or `travel-leisure` on
  current traffic). **Never fabricate a feed URL** — verify every RSS source
  live before inserting (the relojistas source-vetting pattern: 5 verified,
  3 rejected).
- The arming act is writing the pool site's classification spec with
  `news_feed.recommended=true` (+ the trigger's deployed-page condition — check
  whether a pool needs a stub page or a trigger tweak; **resolve this without
  weakening the trigger for real sites**).
- This step starts spending credits (api_news calls per fetch interval). Cost it
  and say so before switching on.

### 3. The 002 similarity baseline — gates everything downstream

- Build the pairwise-similarity check as a sibling of
  `check_duplicate_palette.go:69-83` (the platform's only cross-site discovery
  check — same join shape, different comparand).
- Run it on the armed pool's real articles across its onboarded members'
  *selections* before any member renders a feed. Distribution not mean;
  worst-decile pairs; threshold set before seeing the data.
- Only acceptable numbers open the render path. `features_open/004` (council
  seat) becomes seatable at this point — it needs this baseline to cite.

## Blockers that must clear before step 1 completes

- `bugs_open/027` — news pages render nothing without JavaScript (another
  thread owns it). Defeats the SEO purpose fleet-wide if unfixed.
- `bugs_open/026` — news-listing hardcodes English; blocks the Romanian, Dutch,
  German, Spanish-market and Indonesian domains in this cohort.

## Why parked rather than started

Onboarding 37 domains is the platform's biggest single batch to date (current
fleet: 11), lands on machinery with two known rendering bugs, and every site
built before the similarity gate exists adds to the surface the gate has to
retro-check. The owner's hold is the right order: fix the gates, then grow.
