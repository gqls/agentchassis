# NOTES — vetcomparison.uk

Running record, append-only, **newest at the bottom** (per CLAUDE.md "standing four").

> **CORRECTED 2026-07-18:** this file was previously named `RUNNING_NOTES_*` and ordered
> newest-first — I reordered it to newest-first earlier today, which was doubly wrong. Renamed
> and re-ordered to match the owner's standing-four directive. Caught by re-reading CLAUDE.md.

## 2026-07-14 (discovery)

- Fabrication identified: 3,124 practices with invented prices (22% at £48; zero source URLs;
  "_data_rights: do not scrape" claim), fabricated CMA quote, named-practice £33 claim
  (guides). DB: 997 price rows source='seed_import' on 235 real verified practices → approved
  quarantine (executed 07-15: is_current=false, retained).
- Real assets confirmed: 2,767 verified practices; ~330 with (aggregates-grade) prices; med
  pricing pipeline with evidence store. Handoff 2026-05-18 located: Go-A/Go-B never shipped;
  owed "query 4" answered.
- CMA research (two passes, grounded): final report 24 Mar 2026; dates/remedies as in SUMMARY.
## 2026-07-15 (strip live + guides + plan)

- Strip deployed: origin/master `92526ccd` via detached worktree cherry-pick (bot race handled;
  local clone unusable for pushes). Verified live: prices gone, calc/medicine/guides 404.
- Guides rewritten sourced + live (`f18eb395`), same URLs; banned-content audit clean (only
  CMA's own figures £21/£12.50/£500 on site). Homepage cards restored.
- Interim directory re-exported from 2,579 verified practices (later 2,389 after 07-16 triage).
- PLAN written (phases 0–5); owner decisions taken 07-16: attributed ON w/ opt-out; no RCVS
  badge; min_n=3; respond to consultations.
- LEGAL factual record written + updated with deploy evidence.

## 2026-07-16 (Phases 0–2 + consultation draft)

- **Phase 2 code done, deploy pending.** `directory_export_action.go` + registry entry
  (`directory_export_json`); shared `sendExportFilesToGit`; med exporter refactored onto it.
  **Found pre-existing bug: `med_export_json` was never in GlobalActionRegistry** — registered.
  008 (optout columns) + 009 (agent pair + disabled task) applied to prod.
- **Provenance finding:** all 803 historical business_prices rows have EMPTY source_url (NOT
  NULL but ''); nothing recoverable from data_observations either. Attributed output correctly
  = 0. Historical data → aggregates only (15 rows, 14 areas, min n=3). Fresh scrapes must
  persist per-price source_url.
- **Phase 1 done.** insertPrice → unified schema (+ insertMedicinePrice, loadCurrentPrices
  cutover); offeringSlug pinned byte-identical to 006's SQL (tests + prod cross-check). 006
  applied: 512 service products, 1,953 rows, 762 current, 0 seed_import current. 007 applied:
  36 CMA items (12/6/6/9/3), pet_band on product_prices.
- **Phase 0 done.** Export domain fail-closed (Go + tests); `.co.uk` blanked in
  agent_definitions ea5f6fac + scheduled_tasks 41735d49. Data triage: 20 fake "practices"
  dismissed (yelp/starofservice/bestlocalrated/allvets/calmshops/wheree/US/college), 17 names
  cleaned, 177 wheree-mirror rows → pending. Live directory 2,389 (commits e47a8c65, c80aa50c).
- **Consultation:** funding response drafted (CONSULTATION_RESPONSE_funding_DRAFT_2026-07-16.md)
  — levy basis confirmed flat per-FOP from the portal; owner to verify Notice ¶ + submit by
  30 Jul. Substantive draft Order not yet published as of 16 Jul.
- Harness lesson: `set -e` didn't abort after a failed verify → one push raced out before its
  gate (no harm — still an improvement). Verify and push are now separate steps (RUNBOOK).

## 2026-07-16 (later — Phase 3 claim flow)

- **010 applied:** `business_intel.claim_requests` (claim|optout|correction) with evidence_method,
  status, verifier and a **consent_text snapshot** (not a version pointer — rewording must never
  rewrite what a practice agreed to). `businesses.claimed_by` was an unused unconstrained uuid →
  now FKs to the granting request, so every claim traces to who asked / what we checked / what
  they consented to.
- **Full lifecycle dry-run against prod in a rolled-back transaction; zero residue confirmed.**
  claim → claimed+linked → 4 CMA prices w/ pet_bands → exporter claimed query returns them →
  attributed excludes claimed. Then: scraped price visible → opt-out → attributed 0, still in
  directory → later claim reverses opt-out. Audit trail correct (consent on claim, not opt-out).
- **Site front door live** (`58e2a837`): claim CTA now collects what verification needs (practice
  identity, role, callback number, link to their price list) instead of a bare mailto, states the
  terms, and cites the Dec 2026/Mar 2027 deadlines as the reason to act. **Opt-out route now
  actually on the page** (policy promised it since 07-15; it wasn't there). Verified live.
- Note: grepping live HTML for mailto body text gives false negatives — it's URL-encoded.

## 2026-07-16/17 (Phase 4 adoption — attempted, blocked on missing index)

- Triggered adoption via `082_submit_domain_unified.sh vetcomparison.uk --from
  https://vetcomparison.uk --fidelity locked` (correlation b0af4625-…). Crawl → fingerprint →
  css → analyze → classify_archetype → select_content → derive_content_direction all SUCCEEDED.
- **FAILED at apply_adoption_plan**: `insert needs_domain_research: no unique or exclusion
  constraint matching the ON CONFLICT specification (42P10)`. Cause: `insertWorkItem`
  (load_work_item_actions.go:1060) conflicts on `(site_id, item_key) WHERE item_key IS NOT NULL
  AND status NOT IN (<workItemTerminalStatuses>)`, which must match partial unique index
  `idx_swi_dedup` — **absent in prod**. work_items_common.go:29 comment says 'cancelled' joined
  the closed set in **migration 157 (2026-07-16, another workstream)** — schema + Go must land
  TOGETHER; predicate and clause must be byte-matched or every keyed insert 42P10s.
- **Do NOT hand-create the index** without confirming which statuses list the DEPLOYED binary
  emits (verify against the pod, not git — build/deploy practice memory).
- State left: `sites` row EXISTS (vetcomparison.uk, active/pending), zero site_specs, zero work
  items, live site untouched. Re-triggering adoption after the fix is safe (same command).
- **Unblock path:** next chassis deploy (ships fixloop migration-157 Go + our Phases 0–3) +
  apply migration 157/idx_swi_dedup schema in the same window → re-run the 082 trigger →
  then bump directory-exporter agent image_tags + smoke + enable per RUNBOOK.
- Adoption safety noted from FOCUS_adoption_faithfulness_via_locks(5): only the FIRST-plan
  faithful branch works today; convergence union can clobber adopted sections — after adoption
  succeeds, never re-plan this site to fill gaps (fleet landmine), hand-edits get permanent locks.

## 2026-07-17 (site working pass: dedupe + notice off; chassis v1.0.1130 live)

- Owner priority: site presentable before any documentation is submitted. Chassis v1.0.1130
  deployed (verified against the pod) — unblocks adoption retry + exporter enable, not done yet.
- **Directory deduplicated: 2,389 → 2,109 live.** 264 dismissed by (website host, postcode) —
  www/trailing-slash/http variants from different discovery routes; 9 more by (name, postcode);
  7 judged individually (corporate cvsvets.com pages shadowing practices' own sites, a
  rated.club junk row ×2, a digifarm staging domain; kept Harrogate/Swift Referrals as genuinely
  two businesses). Keep-rule per group: has_prices > latest verified > shortest URL. One
  scheme-less website_url fixed (`www.argyllclinic.co.uk` → https).
- **"Price comparison is being rebuilt" notice removed** from the homepage (`ae93824f`); the
  guides + claim CTA carry the price story. Live-verified: 2,109 entries, 0 dup groups, notice
  gone, claim CTA intact.
- Dedup keys worth keeping for the verifier/discovery deny-work: normalise website host
  (strip www., scheme, trailing slash) BEFORE upsert; rated.club + digifarm.uk join the junk
  host list.

## 2026-07-17 (later — ADOPTED; CLAUDE.md practices in force)

- Read repo CLAUDE.md (multi-session rules) and follow it: pathspec commits per task,
  forward-only, build-from-HEAD, pod-verified deploys, queue check before dispatch.
- Discovered our Phases 0–3 files had been SWEPT into concurrent sessions' commits
  (f51a7accc, d076c3c8e, 37468ba65) — per CLAUDE.md that's fine, forward-only; remainder
  committed narrowly (f604743e7).
- **Pod-verified v1.0.1130 (later v1.0.1134) contains our actions** (directory_export_json,
  insertMedicinePrice) — Phases 0–3 are DEPLOYED. `idx_swi_dedup` now exists in prod
  (migration 157 applied by fixloop session) with 'cancelled' in the predicate.
- **ADOPTION COMPLETED** (correlation 9cf8e0e8): crawl→classify→apply_plan seeded
  content_direction / site_archetype / structure / design_reference / design_intent;
  dispatch loop already ran domain-research-classifier (identity + classification current;
  site_type=hub). Site row active, build=pending; cascade rolling.
- Classifier emits NO content_features (lost 005 patch never landed anywhere) — news-feed
  decision needs the manual one-off spec patch if/when wanted.
- WATCH: build cascade will eventually rerender pages — verify the live site's hand-authored
  pages survive the first faithful pass; never re-plan to fill gaps.
- Still pending: exporter enable (bump directory-* agent image_tags v1.0.1126→current, smoke
  via kcat, enable task) per RUNBOOK.

## 2026-07-17 (latest — EXPORTER LIVE; first autonomous publish)

- directory-* agent image_tags bumped v1.0.1126 → v1.0.1134 (current pod-verified tag).
- Kcat smoke (correlation 3a0c7463): COMPLETED in seconds. Exporter queried prod, built 5
  files, committed via git-adapter (`ac3314fd` on sites/master), deploy served them. **First
  fully autonomous publish of this site's data.**
- All publication rules held in prod output: directory 2,109 (= dedupe), aggregates 13 rows
  min n=3 (was 15 pre-dedupe — two area groups correctly fell below the floor), claimed `[]`,
  attributed `[]` (no provenance). directory-metadata.json carries the policy string.
- `directory-export-json` task ENABLED (48h cycle); last_completed_at stamped by the smoke.
- Adoption cascade progressing meanwhile: classifier/tool-recreation/2 content pages complete,
  strategy triaged, **needs_rerender triaged — verify hand-authored pages survive the first
  rerender**; live site intact at time of writing.


## 2026-07-18 (fabrication returned via the platform; contained + filed)

- **The chassis regenerated fabricated data.** `tool-recreation-handler` (item status: `complete`)
  rebuilt our practice search as a client-side synthetic generator — PREFIXES×SUFFIXES names,
  Mulberry32 seeded RNG postcodes, its own comment "we generate a large, realistic, deterministic
  dataset" — and it went LIVE. Also added copy claiming pricing info + ownership data we don't
  publish, `Price: Low to High` sorting, and "a representative sample for demonstration purposes".
- **CONTAINED:** restored verified homepage from `b2896815`; live-verified 0 generator symbols,
  0 unsupported claims, real `/data/vet-full-index.json` wired, claim+opt-out back.
- **FILED** `/bugs_open/020` (commit `4e372119b`) + pattern in 016b §9 + index row. Root cause:
  (a) recreation path carries NO data-dependency contract (adoption's
  `extract_interactive_fingerprint` never passes the tool's fetch target to
  `tool-recreation-handler`); (b) that prompt's rule 9 "No fake data or dummy outputs —
  calculations must be mathematically correct" is scoped to ARITHMETIC, so it never forbade
  inventing records. Sibling: bug 001.
- > **CORRECTED 2026-07-18:** my earlier claim in this file that the rebuild "dropped the search
  > UI" was WRONG. The search component was present and well-built — it was populated with
  > invented practices. I had checked for the presence of markers, not for what the page said.
  > Caught by re-reading CLAUDE.md's "trust the rendered artefact, not the status".
- > **CORRECTED 2026-07-18:** the 2026-07-16 entry's claim that "all 803 rows have source_url"
  > was wrong — they are non-NULL but EMPTY. My check tested for NULL. Caught when the exporter
  > correctly produced an empty attributed-prices file.
- Working docs brought into line with the standing-four directive (this file renamed from
  RUNNING_NOTES_*, re-ordered newest-at-bottom).
- Live now: 2,109 practices, exporter enabled (48h), guides sourced, claim+opt-out routes live.
