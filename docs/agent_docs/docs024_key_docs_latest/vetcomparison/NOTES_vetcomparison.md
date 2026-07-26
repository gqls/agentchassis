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

## 2026-07-18 (later — spec-level fix: the fabrication was in the DB, not just the file)

- **The file restore was necessary but insufficient.** `page_components.rendered_html` still held
  the fabrication, `deployed` and unlocked, so the next render would have republished it. Found by
  querying the components rather than trusting the restored file.
- **It was not where I assumed.** The generator lived in the **`hero`** slot (18,101 chars — the
  whole recreated tool), not `filtered-result-grid`. Spread across the index page:
  hero = generator + "representative sample" demo claim; filtered-result-grid = `Price: Low to
  High` sort controls; info-card-grid = "pricing information, ownership data" claim.
- **Fixed at source, keeping the chassis's better UI.** Rewrote the hero's data layer only:
  same markup/ids/CSS (search, region filter, pagination, help text) but `fetch('/data/vet-full-
  index.json')` instead of the generator, honest disclaimer, and a comment forbidding record
  generation with a pointer to bug 020. 18,101 → 11,326 chars.
- Copy fixes: price-sort options stripped; info-card intro rewritten to what we actually publish;
  the "Ownership and Group Information" card (a feature we do not have) replaced with the
  claim-your-listing route (a product we do).
- **Swept the other pages** — 3 more hits. Only ONE was a false claim: about/differentiators said
  "The directory identifies independently owned practices separately from those owned by corporate
  groups" (we stripped ownership data on 15 July) → replaced with a true differentiator. The other
  two (about/faq, guide-cma-market-investigation) describe the CMA's *findings and obligations*
  accurately and were LEFT ALONE — a regex sweep would have wrongly "fixed" correct content.
- All four corrected components set `lock_type='permanent'` (only 1 component in the whole fleet
  had a permanent lock before this).
- `page_components.data_path` exists but is **empty fleet-wide (0 rows)** — vestigial, not the
  live data mechanism. Do not build a fix on it.
- > **NOT PROVEN:** I could not dispatch a rerender manually to confirm the fixed source renders
  > correctly — `rerender-pages` is `experimental` and neither
  > `system.agent.site-builder.requests` nor `system.agent.page-rerender.process` produced an
  > orchestration state from a kcat trigger. DB state and live site are both verified correct;
  > the render path is unverified. **Next thread: watch the first natural render.**
- Self-inflicted, recorded so nobody repeats it: `\set html \`cat file\`` in a piped psql runs
  `cat` **inside the pod**, which has no such file — it silently blanked the hero to 0 chars
  (reported `UPDATE 1`). Correct method: generate dollar-quoted SQL locally, pipe via stdin.

## 2026-07-19 (state grounded for handoff; nothing changed)

- Re-checked every figure against the live system rather than carrying them forward. Live page:
  0 fabrication, 0 false claims, real directory wired, claim CTA present, 6,133 bytes.
  DB: 2,109 published / 238 pending / 0 fabricated current / 0 claims / 0 claimed.
  4 permanent component locks intact, no generator in any of them.
- **No render has run since the 07-18 restore** (no commit has touched vetcomparison.uk/), so the
  bug-020 fix remains verified at DB + live-file level but the render path is STILL unexercised.
  When it does run the homepage will change to the richer chassis component (11.3 KB, region
  filter + pagination) — expected and an improvement; verify with the greps in the handoff.
- Exporter last completed 2026-07-17 20:25Z on a 48h cycle → next due ~2026-07-19 20:25Z.
- CMA case page re-fetched: substantive draft Order **still not published** (latest entry remains
  30 June 2026); funding consultation still closes 30 July 23:59. Order not yet made.
- Found files in this dir not written by this workstream: `README_where_we_are.md` (a paste of my
  15 July chat messages — now STALE and misleading: says the strip needs pushing, quotes 2,579
  practices), `README_vet_legal.md` (paste of the copyright discussion), and ~150 KB of
  owner-supplied RCVS research (`RCVS_mismanagement.md`, `RCVS2_*`, `RCVS3_report_and_discussion`)
  on RCVS institutional efficacy and its software-project record. The RCVS material is
  strategically live — they build the competing official tool and set third-party approval
  criteria — and is not yet folded into any plan. All flagged in the handoff.
- Docs brought to the standing four + handoff refreshed to 2026-07-19 with today's grounded
  figures. Runbook gained: DB component editing (dollar-quoted SQL via stdin), the whole-site
  component audit query, and the unsolved render-trigger.

## 2026-07-19 (later — CLAUDE.md re-read: standing FIVE, and a doc I misread)

- CLAUDE.md gained a fifth standing doc since this morning: **`README_where_we_are.md` is the
  OWNER'S plain-prose running log**, append-only, to be written at every natural break ("if you
  wrote a substantial reply in chat, it belongs here too").
- > **CORRECTED 2026-07-19:** earlier today I told the owner this file was "STALE, do not act on
  > it" and wrote that into the handoff, because its opening still describes the 15 July strip.
  > That was a misreading — it is a chronological history, so early entries are supposed to be
  > old. Handoff corrected to describe it properly. Caught by re-reading CLAUDE.md.
- Checked whether I was the session that overwrote it on 07-19 (CLAUDE.md now warns about this):
  **not me** — `git log` shows only the owner has committed that file, and the working-tree change
  is a +35-line append, not a rewrite. I only ever read it.
- Appended this session's entry to it in the owner's register (prose, `--` separator, `※ recap:`).
  Never rewrote or reordered.
- `/bugs_closed/` now exists (split 07-19). Grep BOTH dirs before filing. Numbering is one shared
  sequence, never reassigned; `016` and `017` each name two different cases, so resolve by slug.
  Checked ours: **020 is unique across both dirs** (019/020/021 all distinct).
- Added `SUMMARY_2026-07-19_readout.md` in the five-part structure CLAUDE.md now specifies
  (trying to do · come from · done · where we are · where going). Earlier summaries kept as
  milestone records at their dates.

## 2026-07-19 (later still — the render was going to ship two dead sections; news feed built)

Picked up from `HANDOFF_2026-07-19`. Its priority 1 was "watch the first render". Rather than
wait for it, I assembled what the render *would* produce and inspected that. Two of the five
index sections were defective, so the handoff's framing — "that change is expected and is an
improvement" — was only half right.

**State re-verified first (all still true):** live page 6,154 bytes = still the hand-authored
version; last `page_rerender` item 2026-07-17 22:45Z, restore was 07-18, so **no render has
exercised the fix**. DB: 4 components locked `permanent`, hero carries the real fetch, zero
fabrication markers anywhere. `/data/vet-full-index.json` = 2,109 real practices.

### The handoff shipped a verification command that always fails

```
curl -s "https://vetcomparison.uk/?cb=$RANDOM" | grep -c 'vet-full-index'   # "must be >= 1"
```
Returns **0**, and always will. The live page loads `assets/js/vet-search.js?v5`; the fetch of
`/data/vet-full-index.json` lives *inside that script*, not in the HTML. The check greps the
wrong artefact. A session running it would read a clean site as a regression. Correct form:

```
curl -s "https://vetcomparison.uk/assets/js/vet-search.js?cb=$RANDOM" | grep -c 'vet-full-index'   # >= 1
```
Handoff corrected. **The underlying site was fine throughout** — this was a bad check, not a bug.

### What the render would have published

| slot | verdict |
|---|---|
| `hero` | fine — real `fetch('/data/vet-full-index.json')`, region filter, pagination, honest disclaimer, graceful error state |
| `filtered-result-grid` | **dead** — a *second* search box over an empty grid with a hardcoded "No results found."; its script binds `.filter-btn`/`.result-card`, neither of which exists |
| `info-card-grid` | fine — careful copy ("We do not publish figures we cannot attribute") |
| `latest-news` | **empty** — headline with nothing under it; its JS fetches `/data/latest-news.json`, which **404s** |
| `call-to-action` | fine |

Neither defect is fabrication — both correctly decline to invent data, which is the site's whole
point. They are dead UI, and `hero` already does the directory job completely, so
`filtered-result-grid` was redundant as well as broken.

### Two mechanisms worth keeping (both cost me a wrong assumption first)

- > **MISSTEP, caught before acting:** I was about to remove the dead grid by editing
  > `pages.sections`, which lists exactly the five slots in order. **That would have been a
  > silent no-op.** Both assembly paths — `rerenderLoadSections`
  > (`rerender_pages_actions.go:592`) and `getPageSections` (`rerender_single_page_action.go:381`)
  > — `SELECT ... FROM page_components WHERE page_id = $1 ORDER BY position` and **never read
  > `pages.sections` at all**. To drop a component you must delete or blank the
  > `page_components` row. Caught by reading the functions, which is precisely what CLAUDE.md's
  > corrected "diagnosis before debugging" section tells you to do — it names
  > `rerenderLoadSections` as the function a confident thread skipped.
- **The empty-section guard does not catch shells.** `sectionHasVisibleContent`
  (`rerender_single_page_action.go:446`) strips style/script/tags/entities/whitespace and
  requires `len > 10`. `filtered-result-grid` reduces to ~34 chars — *its own empty-state copy*
  ("Sort: Recommended", "No results found.") is what admits it. `latest-news` passes on its
  headline while its body is a 404. And only one of the two paths has the guard at all:
  `rerenderLoadSections` applies none. Filed to the diagnosis loop, corr
  `be60b0d7-21c4-4e02-be95-2ec37387004f` (queue was clear — no covering item).

- > **MISSTEP:** I first concluded "this site has no news" from `SELECT count(*) FROM
  > content_items` = 0. **Wrong table.** News lives in `content_feed_items`. The conclusion
  > survived re-checking against the right table (also 0), but the evidence I'd have written
  > into a doc would have been wrong, and the next thread would have grepped a table that has
  > nothing to do with news.

### Actions taken

1. **Removed `filtered-result-grid`** from the index page. Snapshot first:
   `_vetcomparison_bak_20260719_index_components` (all 5 rows), then `DELETE`. Page is now
   hero → info-card-grid → latest-news → call-to-action. Positions are now 1,3,4,5 — gaps are
   harmless, both paths `ORDER BY position`.
2. **Built the news feed properly** (owner's call — the empty slot is real work, not something
   to delete). Chain, verified end to end:
   - **The gate was `content_features`, exactly as handoff item 7 suspected.**
     `content-feed-trigger`'s selection query requires
     `(data->'content_features'->'news_feed'->>'recommended')::boolean = true` on the *current
     classification spec*. vetcomparison had **no `content_features` key at all**, so it was
     never selected — the source alone would have been inert. Patched by superseding the spec
     row (`is_current=false, superseded_at=now()`) and inserting a new current one, per the
     unique index `idx_site_specs_current`.
   - **`source_types: ["rss"]` only — this is a load-bearing integrity decision, not a default.**
     `seed_content_sources_action.go:195-219` auto-creates sources per declared type:
     `api_news` spawns an **xAI/Grok LLM source** that *authors* news text, `news_search` one per
     keyword. `rss` is explicitly skipped ("requires manual URL config"). On a site remediated
     for fabricated content, LLM-authored news is exactly the wrong thing, so listing only `rss`
     structurally prevents the seeder from ever adding one. The reason is written into the spec's
     `reason` field so a later thread cannot "helpfully" add `api_news` back without reading why.
   - **Source:** the *keyword-filtered* GOV.UK feed, not the CMA org feed —
     `https://www.gov.uk/search/all.atom?keywords=veterinary&organisations%5B%5D=competition-and-markets-authority`.
     The unfiltered org feed has **zero** vet mentions (it is mergers, parking, dental); the
     filtered one returns 10 entries, all veterinary, top item *"Vets market investigation: draft
     funding Order and Undertakings"*. Checked the parser handles it: `feed_actions.go:203`
     tries RSS then Atom, and the real XML has exactly one `<link rel="alternate">` per entry (so
     `entry.Link.Href` is unambiguous), `<summary>` present, `<published>` absent but `<updated>`
     is RFC3339 and `parseAtomBody` falls back to it.
   - **No LLM in the display path either:** `loadNewsItems` selects `source_title`,
     `source_summary`, `source_url`, `source_published_at` straight from `content_feed_items`.
     Triage only scores `relevance_score`/`status`; it does not author.
   - Verified by running `content-feed-trigger`'s selection query **verbatim**: vetcomparison
     now returns. Feed due at the next 6-hourly sweep (~13:40Z; last 07:40Z).

### Landmine found while doing it

That selection query ends `ORDER BY s.domain LIMIT 5`, and there are now **exactly 5** eligible
sites. `vetcomparison.uk` sorts **last**. A sixth news site would starve it — deterministically,
by alphabet, not at random, and silently. Not filed yet; recorded here and in the handoff.

### Server-rendering the news (owner's second question) — see `/bugs_open/027`

Measured rather than estimated; full findings appended to the bug file, not restated here. The
short version: cheaper than 027 assumed, because **`latest-news.js` is already
progressive-enhancement-safe** (it only overwrites the container when the fetch returned items,
and swallows errors), so server-rendered markup survives an empty or failed fetch with no JS
change. Against that: `data_sources`/`go_template` are **dead metadata** — zero Go readers
repo-wide — so there is no binding engine to reuse. `news-listing.js` is **unverified**.

### Owner queue: 7 → 5 (two cancelled, owner-approved)

Both about-page items had stale premises — `about.html` is built and live (36 KB, sound heading
structure, zero fabrication markers, `curl`-verified 2026-07-19).

- `ff65cc65` "Build about page (not_built)" — was `failed` at **attempt 2 of 3**. Cancelling was
  *protective, not tidying*: a third attempt would rebuild a page that already exists, which is
  the clobber risk in `/bugs_open/001` and the "never re-plan this site to fill gaps" hard rule.
- `f9bf92e7` "Re-render about after its image asset landed" — moot; the live page carries only
  the logo, and a re-render is the same unexercised render path we had just finished cleaning.

Reason recorded in each item's `spec.cancelled_reason` so the decision is auditable from the DB.
Remaining 5 all genuinely need the owner (three 0-section pages, the contact email fact, the
deadline-calculator review).

### The publish chain, verified before trusting it

Ingested feed items are not the same as news on the page. Checked `content-feed-orchestrator`'s
step list rather than assuming:

```
dispatch_feed_sources → render_news_section → render_rss_feed → git_commit
```

So rendered files reach the site repo and deploy — the chain is complete. Two things fall out:

- **`seed_content_sources` runs on every cycle**, not once at setup. So `source_types: ["rss"]`
  is a standing guard, not a one-time choice: on any cycle where that list said `api_news`, the
  seeder would create the Grok source. This is the single most important line of config on this
  site.
- **vetcomparison also gets `/feed.xml`** from `render_rss_feed` — server-rendered and complete,
  which is precisely the artefact `/bugs_open/027` notes is *unaffected* by the client-side news
  defect. So the site gets a crawler-visible news surface even before 027 is fixed.

## 2026-07-20 — the feed ingests but does not publish, and two things I got wrong

Feed is **working at the ingestion end**: 2 real CMA items, genuine gov.uk URLs, real summaries,
no LLM anywhere in the path.

| item | published | age |
|---|---|---|
| Vets market investigation: draft funding Order and Undertakings | 2026-06-30 | 463 h |
| Veterinary services for household pets | 2026-06-30 | 463 h |

**But nothing reaches the site.** `/data/latest-news.json` and `/feed.xml` both 404. Cause found
by reading the orchestration's own `collected_data`, not by guessing:

```
news_render: { rendered: true, item_count: 0, items: [] }
news_commit: null
```

`render_news_json` runs with **`max_age_hours: 72`**; our items are 463 h old, so `loadNewsItems`
filters both out → `item_count: 0` → `check_has_news` routes `0 → complete`, **skipping
`commit_news`** → the JSON is never written. The component then renders its headline over
nothing, which is the exact defect I removed the dead grid to avoid.

72 h is right for daily-churn verticals (gas prices, watch news). It is wrong for **regulatory**
news, which arrives a few times a month. Owner approved widening to **720 h (30 days)**, chosen
to match the 30-day expiry `RenderNewsSectionAction` already applies — so it cannot surface
anything the platform itself considers stale, and fresh items still win the ORDER BY under the
`max_items: 6` cap.

> **MISSTEP 1 — I declared a fleet-wide outage that was not one.** From three readings (13:37
> sweep completed instantly; nothing fetched anywhere; no feed orchestrator pod) I concluded the
> 6-hourly sweep was completing without doing any work fleet-wide, and **filed that to the
> diagnosis loop** (`12ff5852`). It was **late, not broken** — it fetched at 14:30 and our source
> went through cleanly. The likeliest explanation is the one CLAUDE.md already documents: a chassis
> pod roll silently drops spawns within ~300s, and a fresh build had just been deployed. I filed a
> platform-wide claim on ~10 minutes of absence-of-evidence. **Absence of a row is not evidence of
> a defect on a cluster that queues.**
>
> **MISSTEP 2 — a DB-only config edit does not hold.** I set `max_age_hours: 720` on the live
> `agent_definitions` row at ~18:23. The 19:41 run still rendered `item_count: 0`, i.e. it used 72.
> The seeds carry the value (`sql_for_agents/090_...sql:493`, `087_...sql:1913,2297`), so a re-seed
> re-applies it — the clobber landmine, met head-on. Seed 090 now carries 720 with the reason
> inline. **Change config in the seed AND the row, or it silently reverts.**

### Landmine: the reaper eats a queued diagnosis

Run `be60b0d7` (empty-section guard) **FAILED**: `reaper: stale AWAITING_RESPONSES for >90 min`.
It queued ~32 min behind a busy cluster, then was reaped before diagnosing. Its intake item sat at
`awaiting_diagnosis`, which makes the 090 trigger refuse a refile — close the stale item first
(done, with the reason in `spec.failure_reason`), then refile. Refiled as `459fbdf3`.
**A diagnosis filed into a busy cluster can die without ever being answered, and the intake record
will still look open.**

### Separate, and more serious: the directory exporter is now failing

`directory-export-json` ran 2026-07-19 20:25 and **failed**:

```
step export_json failed: failed to execute action directory_export_json:
directory export requires an explicit domain; refusing to export without one
```

The config is **not** missing — `scheduled_tasks.input_data.input_data.domain` is
`vetcomparison.uk`. `DirectoryExportAction` (`directory_export_action.go:123-136`) merges
`params.CollectedData["input_data"]` over `params.StepConfig.Config` and then aborts on empty
Domain, so the key is not arriving in that merged map. Last **successful** run was 2026-07-17
20:25 (48 h cycle, so this was the first run since), with several chassis builds in between —
consistent with a regression rather than a config change. `vet_med_export_action.go:152` has the
identical merge-then-guard shape, so this is unlikely to be one exporter's problem.
**Consequence: `/data/vet-full-index.json` stops refreshing.** The live file still serves 2,109
practices, so nothing is broken *on the page* — it just goes stale. Filed as `2c5bb9e2`.

### Not changed, deliberately

`render_rss_xml` passes no `max_age_hours`, so `render_rss_feed_action.go:131` applies its own
default of **336 h** with a stated reason ("feeds carry more history than the homepage card").
Our items are 463 h old, so `/feed.xml` stays empty too. I did **not** widen it: unlike the
homepage card's 72, that value was chosen deliberately and documented, so changing it is an
owner call rather than a defect fix. Worth raising, because `/feed.xml` is server-rendered and
is the one news surface `/bugs_open/027` does not affect.

### 2026-07-20 (later) — the window was never the bug: numeric step config is inert fleet-wide

The 720h fix did not work, and chasing why found something much bigger.

After the change, the 14:02 run still rendered `item_count: 0`. Checks, in order:

1. The run **did** carry 720 — read from its own `initial_request_data.agent_config`:
   `{"site_id":"input_data.site_id","max_items":6,"page_name":"index","max_age_hours":720}`.
2. The renderer's query, run verbatim against the live DB with 720, **returns both items**
   (status `relevant`, score 55, ~483 h old).
3. So the pod was not behaving like 720. Verified what is actually deployed, per CLAUDE.md —
   against the pod, not git:
   ```
   strings /app/agent-chassis | grep -c loadNewsItems          -> 5   (current query IS deployed)
   strings /app/agent-chassis | grep -c persistNewsSectionHTML -> 0   (027 server-render is NOT)
   ```

**Root cause — `ExtractActionInputs`, `platform/orchestration/datahelpers/action_inputs.go`.**
Every branch that consults step config reads `config[field].(string)` — Strategy 0 (:126),
the `input_fields` branch (:144), the deprecated `*_field` branch (:180), Strategy 4 (:233).
There is **no branch that takes a literal config value.** Consequences:

- a **numeric** config value fails the type assertion and is dropped silently;
- a **plain string** without a dot is treated as a single-segment *reference* and looked up as a
  key in `collectedData` — so a literal like `"vetcomparison.uk"` resolves to nothing.

So `max_age_hours: 720` never arrived and `GetInt("max_age_hours", 72)` returned its fallback.
**`max_items: 6` has never been read either** — it just happens to equal its fallback.

> **Why this hid so well, and the lesson.** The seeded value was **72** and the Go fallback is
> **72**. Config and behaviour agreed, so the config looked live and load-bearing when it was
> decorative. It took *changing* the value to reveal that changing the value does nothing.
> **A config setting that matches its code default proves nothing about whether it is wired up.**
> Predicted-then-observed: configure 720, observe behaviour identical to 72. That is the
> confirmation, not the code read alone.

Filed as `f155b0c4-881b-4369-abe4-569d7b2ad4c8`. Fleet-wide: any action tuned by a numeric step
config value is silently running on its Go default.

> **CORRECTION to my entry above.** I wrote that the fix "is in both places now" and that the
> seed change made it durable. **Both places are inert.** Seed 090 keeps 720 with a loud comment
> recording that it does nothing until the input-plumbing defect is fixed — the intent survives
> and takes effect on the fix, but nobody should read it as working.

**Effective window today is still 72 h**, so the homepage news card stays empty until either the
plumbing is fixed or `render_news_section_action.go`'s fallback is changed (inert until an image
roll). Not attempted here: another session is actively editing that file (`1005e1af2`
server-renders news into the page, per `/bugs_open/027` — which is my own addendum being acted on).

### Related, unresolved

`/bugs_open/027`'s fix is **committed but not deployed** — `persistNewsSectionHTML` is absent from
the running binary. It is inert until the next image roll, so 027 stays OPEN by the standing bar
(fixed AND live).

### 2026-07-20 evening — 042 fixed in code; new build carries 027 but not 042

Fresh chassis build deployed ~17:58 UTC. Verified against the **pod**, not the tag:

```
strings /app/agent-chassis | grep -c persistNewsSectionHTML   -> 8   (027 server-render IS live)
strings /app/agent-chassis | grep -c components_rendered      -> 1
```

So `/bugs_open/027`'s fix is now deployed. **042 is not** — `config[field].(string)` unchanged,
no commit touched `action_inputs.go`. So the effective window is still 72 h.

**Checked before assuming the new build was safe for our page.** With 042 unfixed the render
still sees zero items, and 027's new code writes component HTML — so it could have overwritten
the homepage's `latest-news` component with an empty section. It does not:
`renderLatestNewsCardsHTML` returns `""` for zero items, and `persistNewsSectionHTML` starts
`if db == nil || inner == "" { return 0 }`. It also **skips locked components**
(`lock_type IS NULL OR lock_expires_at < NOW()`) and `html.EscapeString`s third-party titles and
URLs. Nothing regresses; the card just stays empty.

**042 fixed** (owner-approved, non-string scalars only): Strategy 5 in `action_inputs.go`, plus
`action_inputs_literal_test.go`. 5 tests pass, `datahelpers` + `actions` build and pass.
Narrowed from what I first wrote in the bug file — see the correction there: taking *any* type
literally would turn an unresolved reference into a silent literal string, which is a worse and
less visible failure than the one being fixed. Inert until the next image roll.

> **Landmine — a submission fired into the post-deploy quiet window is silently dropped.**
> The orchestration layer created **nothing at all** between 18:53 and 19:06 UTC while the
> scheduler kept beating (30 s tasks firing normally). My council submission at ~19:03 landed in
> that window: zero `orchestration_state_audit` rows, zero artifacts, never started. Resubmitted
> at 19:12 once orchestrations were flowing again.
> **I did NOT declare an outage this time** — that was this morning's mistake, and 12 minutes of
> silence looked identical. Waited, re-checked, saw 8 orchestrations in 5 minutes, concluded
> settling rather than failure. CLAUDE.md documents a ~300 s no-dispatch window after a chassis
> restart; the observed quiet here was longer and started ~55 min *after* the roll, so the
> documented rule does not fully cover it. Check that orchestrations are actually being created
> before trusting any dispatch.

**Where this leaves the news feed:** every link is now in place except the image roll. On the
next build, 042's fix makes `max_age_hours: 720` real, the two CMA items load, the JSON
publishes, and 027 server-renders them into the page for crawlers. Verify in that order —
pod, then artefact, then page — and trust none of the statuses.

## 2026-07-21 — v1.0.1144: the feed works, the render landed, and what it dragged back in

Re-read CLAUDE.md from disk first (it had moved: council ~30 min latency clause, WRONG_CALLS,
`[INFERRED]` markers, who-owns.py). Verified v1.0.1144 against the **pod**.

**042 is live and the news feed works.** `strings /app/agent-chassis | grep -c 'took literal
scalar config value'` → 1 (Strategy 5 present). Behavioural proof: `/data/latest-news.json` → 200,
`item_count: 3`, real CMA items 460h+ old — impossible under the silent 72h fallback. 042 moved to
`bugs_closed/`. The homepage is now the 46.7 KB chassis render with 3 server-rendered news cards in
the raw HTML.

**027's `persistNewsSectionHTML` is gone** (pod grep → 0). Another session reworked 027 through the
council (REVISE → queryresolve resolvers) and the v1.0.1142 sweep `2d529d6dc` removed the function.
News is still server-rendered, so 027's goal is met by their approach — 027 is theirs now.

**The 08:08 render dragged two things back:**
- The dead `filtered-result-grid` I deleted on 07-19 returned. `pages.sections` still listed it, and
  the regenerate path reads the manifest.
  > **CORRECTION to my 07-19 claim.** I wrote then that editing `pages.sections` is "a silent
  > no-op" because both *assembly* paths read `page_components` by position. True for assembly —
  > but the *regeneration* path (full rebuild) reads `pages.sections` to decide which components to
  > materialise. So sections is load-bearing for rebuilds, not for assembly. Removing a component
  > durably needs BOTH: delete the `page_components` row AND remove it from `pages.sections`.
  > Done today; snapshot `_vetcomparison_bak_20260721_index_components` first. DB clean; live page
  > updates on the next render.
- Every `bug-020` `permanent` lock was stripped. Verified delete-and-recreate, not in-place clear:
  `hero`'s row id changed `c8df695e…` → `24060593…`. Filed to `bugs_open/020`: a lock on the old
  row cannot survive a rebuild that replaces the row. Content regenerated clean (real fetch intact,
  zero fabrication), so the exposure is latent — but no lock-based mitigation here survives a
  rebuild.

**Reconciled the pending runs by PAYLOAD (per new CLAUDE.md), not the printed id:**
- Both council submissions (712be028, 563462b8) COMPLETED ~1 h after submission at
  `complete_invalid` — `__step_error`: *"edit 2: operation 'create' not in the allowlist"*. They
  were structurally invalid (I included a file-create edit for the test), never dropped. My
  "dropped → resubmit" call on 07-20 was wrong on both counts (latency, and the plan was invalid
  regardless). Logged in WRONG_CALLS; corrected the false council-linkage in `bugs_open/043`.
- All three diagnosis runs (f155b0c4, 55dc0fa4, 459fbdf3) produced bundles but **no verdict** —
  each has a FAILED/`route` row. `bugs_open/043` route-hang stands on these alone.

**Exporter (55dc0fa4) still unresolved.** Last run 07-19 20:25 FAILED; next ~tonight 20:25. My 042
fix is numeric-only and does NOT cover the literal-string domain case, so it will likely fail
again — needs a human read of `directory_export_action.go`, not the loop (which isn't returning
verdicts anyway).

## 2026-07-23 — med retailer pipeline revived, provenance-first (session "bugfix 054")

Owner directed (yesterday + today): med-discover/med-export will be active; med scrape
prices feed vetcomparison.uk; read the latest docs first. Decisions taken with the owner:
strip `typical_vet_price` from exports entirely; export-side fail-closed provenance guard;
data files only (no medicine page rebuild — separate task, bug-020 class).

**Shipped:**
- `f82f8b425` + gofmt `ff5e2f7df` (v1.0.1151): `vet_med_export_action.go` — TVP stripped
  end-to-end (struct/SELECT/Scan/option/assignment; was `omitempty`, so it had to be gone at
  compile time); `filterMedExportProvenance` between load and grouping (drops url-less/
  zero-date rows); the previously SILENT `rows.Scan` error `continue` now feeds the same
  counter; always-present `skipped_missing_provenance` in price-metadata.json.
  `scan_discovery_candidates.go`: +7 aggregator domains (5 RUNBOOK families). Discriminating
  tests for the failing branch (drop 3 of 4, count them; metadata field present at 0).
- `b28137859`: seed 037 worker-config domain BLANKED (the def INSERT is ON CONFLICT DO
  UPDATE SET default_config — an unblanked seed re-run would reinject vetcomparison.co.uk
  into the LIVE worker config); NEW vetcomparison/011_med_scrape_prices_task.sql (the row
  NEVER EXISTED — 096's UPDATE was a silent no-op, which is why nothing ever populated
  med_price_snapshots on a schedule); 096 annotated.
- Deploy: business-intel rolled to v1.0.1151 (the export runs IN-PROCESS there — the task
  targets med-json-exporter directly, no spawn); all 8 med-* agent_definitions bumped to
  v1.0.1151 (spawned temp pods run the DEF tag, not the deployed image —
  spawn_actions.go:2127-2139,:2755). Pod-grep: needle skipped_missing_provenance=3,
  positive control=1, stripped literal=0, deny-list needle=1. Council submission corr
  abf75d33-c9ac-42fc-99b3-47ddf2694422 (verdict pending at commit time; committed without
  trailer per discipline).

**Missteps this session (also in WRONG_CALLS 2026-07-23):**
- Yesterday's close-out of bugs_closed/054 claimed "no %med% table exists live" — FALSE,
  the query filtered table_schema='public'; everything is in business_intel.*. Caught by
  the owner's "read the latest docs" + `\dn`. Corrected in bugs_closed/054.
- I initially framed the empty med-export domain as a "gap to fill" and left seed 037
  carrying vetcomparison.co.uk — the RUNBOOK rail says never reintroduce it. Corrected
  yesterday (`2377ba5c4`), and today the WORKER-config copy of the same domain was also
  found and blanked (the 2377 fix only covered the scheduled-task payload).
- Enabled med-discover-urls at 11:49:19, ~187s after the business-intel restart — inside
  the ~300s no-dispatch window. Got away with it (orch 5bb6cc19 EXECUTING at 11:49:47);
  do not copy the timing.

**State at writing:** discover run 5bb6cc19 EXECUTING; scrape + export still disabled,
next in sequence. Enable order + verification queries now in the RUNBOOK §Med retailer
pipeline.

## 2026-07-23 (later) — pipeline verified END TO END; two data-quality findings filed

**D1 discover:** orch 5bb6cc19 COMPLETED; listings 304→306 (+1 animed, +1 pdo).
**D2 scrape:** orch 8e2eaa07/parent 5717ab5d COMPLETED (~45 min for batch of 20 — the LLM
fallback runs at ~3 tok/s on CPU ollama; a regex-miss page costs 10+ min. If retailer page
formats drifted since April, fallback will dominate runtime — refreshing the regex
strategies is the real speed fix). 10 fresh snapshots + 19 evidence rows.
**Spot-check method that WORKS:** the retailer pages are JS shells — curl sees zero prices;
Firecrawl renders them. So verify against `med_scrape_evidence.markdown_content` (the
retained rendered artefact), NOT a raw curl of the retailer. 8/10 prices verified verbatim
in their own evidence. **2/10 did NOT** (Advocate ±£0.20) → filed `bugs_open/061` (scrape
can store a price absent from its own evidence; gate checks provenance PRESENCE not parse
FIDELITY). Also flagged there: legacy category-page listings ("Cat Tick"/"Horse", from
2026-04-02 discovery) publish cheapest-in-category prices under non-product names —
hygiene = owner call; note export ignores `med_retailer_listings.is_active`.
**D3 export:** orch da5345e3 COMPLETED (in-process on business-intel, ~1.5s); git commit
`a52fbf0` (9 files); LIVE minutes later — price-metadata.json exported_at=2026-07-23T12:31:28Z
with `skipped_missing_provenance: 0` PRESENT (the new-code proof), medicine-prices.json
10/10 options carry url+collected_at, ZERO typical_vet_price. Chain verified: DB → export →
git → live artefact.
**Council:** APPROVED round 1 (corr abf75d33, 3 advisory objections none high-severity).
All three med tasks now ENABLED (weekly discover / 6h scrape / 48h export).

## 2026-07-24 — the dead search grid recurs a THIRD time: root cause is the PLAN, fixed at source

**Symptom (continuation of HANDOFF_2026-07-21 §2 watch-item 2).** The dead
`filtered-result-grid` component was back on the live homepage again (`curl` → 28
`filtered-result-grid` occurrences, the empty "No results found." grid, and the
sort dropdown whose `Price: Low to High` option is the last surviving entry in the
07-19 fabrication-marker grep). It had been hand-removed twice — 07-19 (page_components)
and 07-21 (page_components **and** `pages.sections`) — and returned both times.

**Diagnosis (self-evidencing, local — no diagnosis-loop needed).** The index page
(`9fad89c1-…`) was re-rendered and deployed **2026-07-23 20:36:41**; that render
re-materialised all components (consistent with the 07-21 note that a render
delete-and-recreates every component and strips the bug-020 locks). Traced the source
of the section list upward:
- `pages.sections` again listed `filtered-result-grid` (the 07-21 removal had not held).
- `site_plan_sections` for the **current** plan (`9d9c601d`, `is_current=t`,
  `source_agent=build-site-planner`, never superseded) listed it at `ordering=1`,
  between `hero` and `info-card-grid`.
- Chain confirmed in code: `reconcile_site_plan_action.go:393` regenerates
  `pages.sections` from the plan; `plan_sections`/render then materialise
  `page_components`; `getPageSections` (rerender_single_page_action.go:393) assembles
  the deployed HTML **from `page_components`**. So the plan is upstream of everything a
  hand-delete touches — deleting downstream can never hold.

> **CORRECTION to HANDOFF_2026-07-21 §2:** it framed the recurrence as a
> `bugs_open/001` **re-plan-clobber**. It is not. Nothing re-planned the site — the
> plan (unchanged since 07-17) has **always** contained the grid, and every faithful
> render reproduces it. This is a plan-content correction, not a clobber, and not a
> lock problem (`bugs_open/020`). The 07-21 instinct — "the source is the site plan" —
> was right; the bug-number attribution was wrong.

**`suppressed_sections` does NOT help here** `[verified by code read]`: the column is
consulted only by `pageSectionShortfall` (v3_site_actions.go:844) and the
`check_empty_sections` / `check_required_fields_missing` discovery checks — i.e. it
suppresses *nagging*, not *rendering*. Neither the render (`plan_sections`) nor the
assembly (`getPageSections`) excludes a suppressed section, so the admin "hide section"
endpoint (page_admin_handlers.go:562, comment "Phase 5 prep — column may not exist yet")
would not actually remove a section from the deployed page. Noted here as a latent
platform gap; NOT asserted as a filed bug (fleet-wide claim, unverified beyond the read).

**Fix (DB, live immediately — no image roll).** Snapshotted first
(`_vetcomparison_bak_20260724_plan_sections` / `_index_pagecomponents` / `_index_page`),
then in one transaction: deleted the grid row from `site_plan_sections` and closed the
ordering gap (→ 0 hero,1 info-card-grid,2 latest-news,3 call-to-action); removed
`filtered-result-grid` from `pages.sections`; deleted the grid `page_component` and
renumbered positions (→ 1..4). All three layers now consistent at 4 sections, so
`pageSectionShortfall` sees planned=4==rendered=4 (no false shortfall).

**Why it holds now:** the plan no longer contains the grid, so a reconcile can only ever
regenerate `pages.sections` as the 4 real sections. `[UNVERIFIED — awaiting render]` the
LIVE page still shows the grid until the next full re-render. `content-feed-refresh`
(6 h cycle, last 2026-07-24 13:44, **next ~19:44 UTC**) re-renders the whole homepage, so
that run should flush it. Verify then:
`curl -s "https://vetcomparison.uk/?cb=$RANDOM" | grep -c 'filtered-result-grid'` must be 0.
If it returns AGAIN after a render, the plan edit did not survive → the planner regenerated
the plan (a genuine re-plan), which is the real `bugs_open/001` case — do NOT re-delete.

### 2026-07-24 (later) — correction: the 6h content-feed cycle does NOT reliably re-render the homepage
My prediction above ("next content-feed-refresh ~19:44 … should flush it") was **wrong** and I
watched it fail. The 19:45:02 `content-feed-refresh` run completed in <1s (`last_triggered_at`
== `last_completed_at`) — a **no-op**: no new/changed news, so no homepage render dispatched, so
no flush. `[OBSERVED]` The index page re-renders *periodically*, not every 6h cycle: `deployed_at`
went 2026-07-23 20:36:41 → 2026-07-24 14:05:29 (~17.5h apart), each a real content-feed render
that had work to do. The news window is 720h and the CMA items are ~530h old with no new CMA
veterinary news imminent, so the next real render is driven by whatever next changes the news
HTML (likely the daily relative-date rollover) — **expect the live flush within ~a day, not on a
6h clock.** The DB fix is durable regardless; this only affects *when* the live page catches up.
The verification one-liner is unchanged: `curl -s "https://vetcomparison.uk/?cb=$RANDOM" | grep -c
'filtered-result-grid'` → 0 once it flushes. Did NOT force a render: a `page_rerender` assemble-only
item is the safe mechanism, but create_rerender_items_action.go warns against hand-rolling the
INSERT (item_type/pipeline/dedup-key), and no enabled sweep picks a reason-less item up cleanly —
not worth the risk on this multi-session live site for a cosmetic empty box. Left to natural cadence.

### 2026-07-24 — CMA substantive draft Order PUBLISHED 21 Jul 2026; two live consultation deadlines
`[VERIFIED against the CMA case page today]` https://www.gov.uk/cma-cases/veterinary-services-market-for-pets-review
- **Substantive remedies Order**: "draft Veterinary Services Market Investigation Order 2026" +
  draft Explanatory Note + draft RCVS Undertakings 2026 — **published 21 Jul 2026**, consultation
  **closes 23:59 on 20 Aug 2026**. This resolves HANDOFF_2026-07-21 open item #3 ("still not
  published; overdue and imminent" — it landed the same day that handoff was written). Response
  basis = `CONSULTATION_2026-07-16_briefing.md` §3 (express reuse right, machine-readable lists,
  no selective blocking). OWNER SIGN-OFF REQUIRED — not drafted/submitted by me.
- **Funding Order consultation**: **closes 23:59 on 30 Jul 2026** (~6 days) — confirmed on the case
  page (item #2 asked for this to be re-verified against the notice). Draft ready at
  `CONSULTATION_RESPONSE_funding_DRAFT_2026-07-16.md`; owner verifies levy figures + submits via
  connect.cma.gov.uk (VetsMI@cma.gov.uk).
- Most recent case-page entry: 21 Jul 2026 "Consultation on draft substantive Order and Undertakings
  published."

### 2026-07-24 — bugs_open/061 DIAGNOSED, fixed, remediated (session "bugfix 061 med scrape")
`[VERIFIED against llm_call_log + med_scrape_evidence, queries in RUNBOOK §fidelity sweep]`
- Mechanism CONFIRMED, no diagnosis loop needed — the fabricated values sit verbatim in
  `llm_call_log` ollama responses, timestamps matching the snapshots to the second. The LLM
  fallback (regex found 0 variants → Mistral on a 1,500-char window) invents price tables
  when its window holds no product prices; the £-gate had checked the FULL section, so a
  delivery banner's £49 fired it. One call echoed the prompt's worked example verbatim
  (£17.48) — 79 April-era snapshots carry that exact price fabricated (19 others are
  genuine 17.48s; only the evidence check separates them).
- Both filed hypotheses REFUTED: no was-price (LLM never saw the real prices); no markdown
  divergence (9293 vs 9256 = Go bytes vs PG chars; octet_length = 9293 exactly).
- Blast radius corrected 2 → **212** (8 in the export window incl. the 2 live Advocate
  rows; 204 April-era). All quarantined to `med_price_snapshots_quarantine_061` +
  deleted; MV refreshed; 33 poisoned `med_retailer_listings.last_price` reset (column is
  write-only in Go — checked). Full-table sweep now 0 PRICE_ABSENT. Live JSON self-cleans
  at the next med-export-json run (48h cadence).
- Fix BUILT + 17 unit tests green (archive-overlay build): write-time parse-fidelity
  guard, `scrape_llm` provenance label, gate on the actual LLM window, prompt hardening.
  INERT until image roll — until then re-run the RUNBOOK fidelity sweep after scrapes.
- Wrong turn worth keeping: my first sweep pattern `FM999999D00` renders 0.42 as `.42`,
  which LIKE-matches nearly anything — sub-£1 fabrications would false-OK. Re-ran with
  `FM999999990D00`; counts happened to be identical, but the corrected form is the one in
  the RUNBOOK.

### 2026-07-25 — dead grid CONFIRMED gone (two renders); and the homepage's main grid is 100% dead links
`[VERIFIED live — curl + DB, queries in RUNBOOK]`

**1. The 07-24 plan-level grid fix held.** Two full renders have since run
(`page_components.updated_at` 07-25 01:49:55 and 13:51:20) and both produced **four** sections,
not five. Live homepage 42,051 bytes (was 46,656 with the grid); `filtered-result-grid` = 0,
`No results found` = 0, fabrication markers = 0. Plan / `pages.sections` / `page_components` all
agree at 4. `SELECT ... WHERE slot_name ILIKE '%filtered%'` returns 0 rows site-wide.
**This closes the recurrence** — the two prior hand-deletes (07-19, 07-21) were downstream of the
plan and came back; the plan edit did not. My 07-24 prediction that it would flush on a ~6h
content-feed cycle was wrong (recorded then); it flushed on the next render, ~1 day.

**2. Renders delete-and-recreate rows.** Comparing `_vetcomparison_bak_20260724_index_pagecomponents`
to live: `content_data` and `rendered_html` byte-identical, **but `id` differs**. So content is
carried through an assemble-only render while row identity is destroyed. That is the same
mechanism the 07-21 handoff recorded for the stripped `bug-020` locks, now confirmed on a second
component, and it is why `save_page_sections_action.go:498` needs its locked-row carve-out
(`bugs_open/058`). Consequence for this site: **any DB edit to a component is safe from
assemble-only renders but not from a full content re-render.**

**3. NEW, and worse than the grid: every link on the homepage's main content grid is a 404.**
`info-card-grid` has six anchors and all six are dead — `/search`, `/about-pricing`,
`/about-ownership-disclosure`, `/guides/pet-owner-rights`, `/claim-listing`, `/guides/cma-compliance`.
Five have no `pages` row at all. The sixth points at a page that **exists and is live** at
`/guides/cma-compliance/index.html` — a URL-form miss (no directory-index rewrite on this host;
both `/guides/cma-compliance` and `/guides/cma-compliance/` 404). Three further chrome links point
at `planned`, never-deployed pages (`/directory/index.html`, `/guides/index.html`,
`/tools/compliance-deadline-calculator/index.html`) = 9 live 404s on the homepage.
- The URLs are **authored**, not resolver-derived: no `site_specs` row for this site contains them
  (all 12 current aspects checked) and `site_plan_sections` carries no content at all. They live
  only in `page_components.content_data`.
- `info-card-grid` is **absent from `ctaFieldNames`** and — the real point — its links are
  `cards[].link_url`, inside a repeating array, which that map's `[2]string` value type cannot
  address. Not merely unenrolled: **unrepresentable**. Filed as a contributed finding on
  `bugs_open/023` (OWNED, active — `who-owns.py` checked first; I did not fork a fix).
- The audit backstop did not catch it either: vetcomparison has **zero** rows of every link/CTA
  item type across all time, against 188 fleet-wide. `[INFERRED]` the completeness discovery agent
  has never run against this site's deployed HTML — *not provable*, because `orchestration_states`
  is pruned at ~24h (the `bugs_open/044` over-flagging trap), so its silence is not evidence.
- **Fixed one of six** (`/guides/cma-compliance` → `/guides/cma-compliance/index.html`, in
  `content_data` AND `rendered_html`; snapshot `_vetcomparison_bak_20260725_index_components`).
  The other five have no destination to point at — an owner decision, and three of those
  destinations are already sitting in his review queue as the 0-section pages.

**4. Wrong turn, self-caught.** I first read the CMA landing page and concluded two of the three
remedies named in the site's news subheadline (prescription fee cap, ownership disclosure) were
unverifiable, because the page lists only document titles. That was reading the wrong artefact.
Extracting the actual PDFs (`pdftotext -layout`) confirms **all three**: Article 7 Price List,
Article 18 "imposes maximum fees (also referred to as price caps)" with a Primary and an
Additional Prescription Fee Cap, Article 5 Ownership Information. The generated copy is accurate.
**The check that mattered was the primary source, not the landing page.**

**5. The trap in that Order, and it is this site's exact failure mode.** Every figure and date in
the draft is a **bracketed placeholder**: `'Initial Primary Prescription Fee Cap' means [£21
inclusive of VAT. This will be adjusted for inflation ... before the Order is made]`, likewise
`[£12.50]`, and every compliance date reads `[X March 2027]` / `[X December 2026]`. The *relative*
periods are firm (3/6/9/12 months, Large vs Small businesses); the absolute dates and the amounts
are not settled until the Order is made. `[VERIFIED]` the live site publishes **none** of them
(grepped £21 / £12.50 / the four date strings across every live page — 0 hits).
**Directly decides the `tool-compliance-deadline-calculator` build-or-cancel** in the owner queue:
it cannot emit absolute deadlines today without asserting dates the CMA has not fixed — which is
precisely what this site was remediated for. Buildable honestly only as relative periods from a
"date the Order is made" input.

### 2026-07-26 — bugs_open/061 CLOSED: the guard caught a live fabrication (session "bugfix 61")

Picked up 061 to fix it and found there was no code left to write — `ca2cd7535` had shipped
in `v1.0.1165` at some point after it was committed, so the case was only open because
nobody had checked. Deployment proof, taken from the **spawned worker's own binary** rather
than the deployment (med workers run `agent_definitions.image_tag`, all 8 = `v1.0.1165`):
`fidelity guard dropped variant` ×1, `never copy its values` ×1,
`never estimate, recall, or compute` ×1, positive control `MedScrapePrices` ×23. The 12:01
production run returned `{"scraped":15,"variants_stored":26,"failed":5,"fidelity_skipped":0}`
— `fidelity_skipped` is a field the fix introduced, so the guarded path runs every scrape.

**MISSTEP, and it is the whole point of this entry.** I had all the green evidence — full-table
sweep 0 PRICE_ABSENT, live JSON re-exported clean, and **16 consecutive post-fix LLM fallback
calls returning `[]`** — and I wrote in the plan that the prompt hardening had removed the
fabrication at source, so the guard's drop branch was belt-and-braces and *could not be
exercised live*. I ran the induced test anyway, only because the standing rule says a green
happy path proves deployment and not correctness.

**It fabricated on the first run.** Induced re-scrape of listing
`0b50fd2d` (`petdrugsonline.co.uk/advocate`, the original page), worker
`agent-med-price-collector-eb928d3a-fmrph`, 15:01:07Z:

- regex → 0 variants (correct, category page); £-window gate → **passed**, fallback fired;
- `llm_call_log` `baa8b777`, latency **495,177 ms**, returned three invented variants:
  `{"Large Cats 80mg/8mg (4kg-8kg)", 19.25, tvp 36.75}`, `{"Small Dogs 10kg Pack of 3", 34.99}`,
  `{"Medium Dogs 25kg Pack of 3", 68.75}` — a size label that is not on the page at all;
- 3 × `MedScrapePrices: fidelity guard dropped variant — price absent from scraped markdown`,
  each `collection_method=scrape_llm`; **0 snapshots stored**; evidence row
  `variants_found=3, prices_stored=0`.

The guard was right rather than merely strict: `19.25` / `34.99` / `68.75` / `36.75` appear
nowhere in the 9,256-char evidence, while the page's real `17.75` and `29.75` are both present.
Note the invented values **differ from July's** (17.95/29.95) — the model hallucinates afresh
each time rather than repeating, which is why "we've seen the fabricated values" is never a
purge criterion. Full-table sweep after: **2,577 OK / 0 PRICE_ABSENT**.

So the correction: **the guard is load-bearing, not defence-in-depth**, and the 16 `[]`
responses were other pages, not evidence of a cured model. 016b §9 had already said "prompt
rules alone are hope, not enforcement" — I nearly contradicted a pattern this workstream
itself filed. Logged in `WRONG_CALLS.md`.

**[OBSERVED, one sample] Throughput ceiling, no data harm.** That fallback call blocked the
per-listing loop for 8m15s; the batch is 20 listings 6-hourly and this run had managed ~4 by
the 10-minute mark. Worth measuring before anyone raises `batch_size`. Not filed — one sample,
and the guard holds regardless.

**Council trailer.** `ca2cd7535`'s APPROVED verdict (`7cf73cc1`) landed 2026-07-24 20:39:50Z,
**after** the commit, so it can never carry a `Council-Reviewed:` trailer and 098 will always
list it as unreviewed. Recorded in the case file, which is the only place the join survives.

### 2026-07-26 — why the site is thin: the data pipeline has been OFF since March
`[VERIFIED — scheduled_tasks + businesses.last_verified_at]`
Owner dropped both CMA consultations ("not strong enough yet") and chose all four candidate
directions. New plan: `PLAN_2026-07-26_site_strength.md`.

**The diagnosis is not "pages are missing".** Filling the nine homepage 404s would make the site
look complete without making it stronger. The site is thin because nearly everything we hold is
barred by our own provenance rule:
- prices: **0 of 762** current rows carry a source URL (unrecoverable — they predate the rule)
- `group_name`: no evidence trail, and **contradicts `is_independent` on 870 practices** (all 55
  Medivet, 27 CVS Vets, 34 Vets4Pets flagged "independent")
- `vet_practice_details`: 2,781 rows (species 2,486, emergency 1,648, accreditations 957) — no
  source, no as-at date
- Companies House: 5,798 companies collected, but only **158** matched deterministically
  (tier1/tier2 ≥0.91); 542 more at 0.50–0.81 by postcode-proximity or LLM — a guess, not a fact
- none of it is live: the export publishes exactly 6 fields and no ownership/price field, so this
  is **under-publishing, not mis-publishing**. No correctness emergency.

**Root cause, and it is simple: every vet collection task is disabled and has been since March.**
`vet-sweep-continue` (last 03-17), `vet-batch-verify` (03-19), `ch-vet-collect` (03-29) — all
`enabled=false`. `max(last_verified_at)` = 2026-03-18; verified in the last 30 days = **0**. They
were switched off during the July fabrication remediation, correctly, and everything since has
been publishing-side work (exporter, adoption, guides, news). **Nobody restarted collection.**

**Two wrong turns, both caught before they cost anything:**
1. I wrote a plan whose first step was "apply seed 082, it was never applied". Reading the seed to
   the end shows it **creates the agent then DELETEs it** — *"forget it, we'll use the regex in the
   verifier"*. Applying it is a no-op; the standalone scraper was deliberately retired. Caught by
   the seed-037 landmine habit (read the whole seed before applying).
2. I then assumed the replacement was never built. It was:
   `StoreBusinessVerification` takes the LLM `registration_number`, else falls back to a
   **deterministic regex** over the scraped page (`business_intel_actions.go:350-372`;
   `updateBusinessFields` maps `registration_number` → `company_number_scraped`, line 909).
   Both readings were "the mechanism is missing"; the truth was "the mechanism is idle".

**The one unknown that sizes everything downstream.** The verification path writes
`data_observations` *including* a `source_url` column (`business_intel_actions.go:328-334`), yet
all 2,970 existing rows have it **empty**. So we do not know whether a live run records provenance
or not. If it does, restarting collection makes the data publishable. If it does not, restarting
just refreshes unpublishable data and the real fix is a Go change first (council → build → roll).
**Next action is a ~10-practice pilot to answer exactly that** — not re-enabling the tasks, which
is one UPDATE and the wrong first move.

---

## 2026-07-26 ~22:45 BST — P1 answered WITHOUT a live crawl, and the hit rate measured read-only

Session "bugfix 061", continuing from `HANDOFF_2026-07-26_continue_here.md`.

### 1. The provenance unknown is ANSWERED — and it did not need the pilot

The previous entry (above) says *"we do not know whether a live run records provenance"* and
proposes a ~10-practice live pilot to find out. **It is answerable statically, and the answer is
that `source_url` is structurally guaranteed empty.** Three independent artefacts agree:

- **The writer reads provenance from the LLM's own output.**
  `StoreBusinessVerificationAction` sets `sourceType/sourceName/sourceURL` from
  `verResult["source_type"|"source_name"|"source_url"]` (`business_intel_actions.go:322-324`),
  where `verResult := extracted["verification_result"]` (line 180).
- **The prompt never asks for them.** `vet-practice-verifier`'s `extract_and_reconcile` step
  requests exactly six sections — `business`, `vet_details`, `vet_staff`, `prices`,
  `confidence_score`, `extraction_notes`. There is no source/provenance field in the prompt.
- **The real fetched URL never reaches the writer.** The step wiring, read from
  `default_config->'workflow'->'steps'`:
  `search_practice→search_results` → `scrape_website→scraped_data` → `prepare_context→
  extraction_context` → `extract_and_reconcile→verification_result` → `store_results`.
  `store_results.config.input_fields` is `["business_id","verification_result","task_id"]` —
  **`scraped_data` is not in it**, so the URL `scrape_web` actually fetched is unreachable from
  the writer even though `scrape_website.config.url_field` names it deterministically.

**Empirical confirmation, which is the strongest of the three** — `raw_data` *is*
`json.Marshal(verResult)`, so its keys are the object's keys:

```sql
SELECT count(*) AS total,
       count(*) FILTER (WHERE raw_data ? 'source_url')  AS has_source_url_key,
       count(*) FILTER (WHERE raw_data ? 'source_type') AS has_source_type_key
FROM business_intel.data_observations;
--  total | has_source_url_key | has_source_type_key
--   2970 |                  0 |                   0
```
The key set present across all 2,970 rows is exactly the prompt's six sections plus LLM
improvisation (`opening_hours` 43, `services` 14, `branches` 5, …). So this is **not** "the column
was added later" and **not** "`sourceURL` resolves empty at runtime" — the two hypotheses the plan
offered. It is a contract mismatch: the writer reads three fields the producer is never asked to
emit, and the component that *does* know the URL is not wired to the writer.

**The fix is NOT "ask the LLM for the source URL."** On this site of all sites, an LLM-asserted
provenance string is a fabrication surface — it would be a model claim about evidence, which is
precisely the class this site was remediated for. The fix is to pass `scraped_data` into
`store_results` and record the URL actually fetched. `input_fields` is config (live immediately),
but the writer must also be taught to read it — **so a Go change is required**: council → build →
roll, exactly the slower branch the plan reserved for this answer.

**Filed to the diagnosis loop before asserting it** (CLAUDE.md: spend it *before* you assert):
`SUBMISSION_CORR = e6580fe5-7537-4eba-a3aa-7863ce4dbfc7`. Verdict pending at time of writing —
**if it refutes any of the above, correct this entry in place and say so.**

### 2. Company-number hit rate — measured, read-only, nothing written

The owner asked for a pilot on ~25 with the hit rate reported. **I did not run the live verifier
to get it.** Running it would have written LLM-extracted, unsourced facts over 25 practices'
current rows to obtain a number that a read-only probe gives for free. Instead: a Go probe
(`scratchpad/pilot/probe.go`) that copies `companyRegNumberPatterns` and the footer-first strategy
**verbatim** from `business_intel_actions.go:1543-1603`, over a deterministic sample
(`ORDER BY md5(id::text) LIMIT 25` of active, non-opted-out, verified practices with a website).
No DB writes, no LLM, no fabrication surface. Polite: 1 req/s, 15s timeout, identifying UA.

| arm | what it models | result |
|---|---|---|
| **PROD** — homepage only | what the live verifier sees today | **4 / 25 (16%)** |
| **HEADROOM** — + `/terms`, `/privacy`, `/legal`, … | if `follow_links` were widened | **7 / 25 (28%)** |

22 of 25 reachable (1×403, 2× connection error).

**`[SMALL SAMPLE]` 25 is a small n and these rates carry wide intervals** (16% is roughly 5–36%;
28% roughly 12–49%). Treat them as "roughly a sixth today, roughly a quarter achievable", not as
point estimates. A larger sample is cheap now the probe exists.

**The headroom finding is a config-only win.** `scrape_website.config.follow_links` is
`[fees, prices, about, team, contact, services]` — **not one of those is a legal/terms page**,
which is where UK companies most often print a registration number. Three of the seven hits came
from `/privacy`, `/terms`, `/terms-and-conditions`. Widening that list is a DB config change:
live immediately, no build, no roll.

**Found numbers are high quality — 6 of 7 resolve to a real CH vet company:**

| number | `ch_vet_companies` | found on |
|---|---|---|
| 10084952 | VETPARTNERS PRACTICES LIMITED | homepage |
| 03777473 | CVS (UK) LIMITED | homepage |
| 07674796 | HENLEY VETS LIMITED | homepage |
| 10687455 | *(no CH match)* | homepage |
| 05185406 | DNA VETCARE LTD | `/privacy` |
| 05886364 | ARK VETS LIMITED | `/terms` |
| 06798554 | LANGFORD VETERINARY SERVICES LIMITED | `/terms-and-conditions` |

**This is the P2 unlock showing itself.** Two of the seven immediately expose true group ownership
— Heywood Veterinary Centre → VetPartners, Animed Whitstable → CVS (UK) — sourced to a company
number anyone can check, which is exactly what the 870-practice `is_independent`/`group_name`
contradiction needs. When the regex fires it is worth a lot; the constraint is how often it fires.

### 3. Incidental data-quality finding

1 of the 25 sampled "practices" is **not a practice**: *"Vets in Blackburn - Lancashire Telegraph
Business Directory"* → `directory.lancashiretelegraph.co.uk`. Same family as the 176 `wheree.com`
rows the RUNBOOK flags. `[UNMEASURED]` — I have not counted how many of the 3,419 are directory
listings rather than practices; the sample suggests it is worth counting before P3 builds a page
per practice.

### 4. `> **CORRECTED same session, ~23:15 BST** — my own §2 "config-only win" was wrong

Above (§2) I wrote that widening `scrape_website.config.follow_links` to include legal pages was
"a DB config change: live immediately, no build, no roll". **That is false.** I recommended a
config change without checking that the config key is read.

```
grep -rn "follow_links" --include=*.go .   # -> no hits, anywhere in the repo
grep -rn "extract_mode\|fallback_url_field" --include=*.go .   # -> no hits
```

`WebscrapeAction` (`webscrape_actions.go:27-147`) reads only `url_field`, `url`, `action`,
`upload_results` and `scrape_config`, resolves **one** URL, and dispatches it to the webscrape
adapter. So **four of the six keys on that step do nothing**:

| key on `scrape_website` | reads as | actually |
|---|---|---|
| `max_pages: 3` | fetch up to 3 pages | inert — one page |
| `follow_links: [fees, prices, about, team, contact, services]` | follow six link types | inert |
| `extract_mode: "text"` | text extraction mode | inert |
| `fallback_url_field: "search_results.results.0.url"` | no website → use top search hit | inert |

**What this changes:**
1. **The 16% is exact, not approximate.** My PROD arm fetched the homepage only — which is
   precisely what production does. I had described it as a "conservative lower bound"; it is the
   actual figure.
2. **28% is not free.** It needs a Go change (teach `scrape_web` to honour `max_pages`/
   `follow_links`) or additional explicit scrape steps in the workflow. It should therefore be
   **bundled with the provenance fix** — one council round, one build, one roll — rather than
   sequenced as a quick win beforehand.
3. **`fallback_url_field` is a silent dead path.** The intended "practice has no website → scrape
   the top search result" never fires. Moot for now (all 3,419 rows carry a `website_url`) but it
   is a trap for anyone who assumes the fallback protects them.

**This is a second structural finding, not just my error:** a step whose config reads like a
six-page crawl with a search fallback, and is a single GET. Nothing warns — unknown config keys
are silently ignored, so the config is documentation that cannot go stale-checked. `[INFERRED]`
that this class is fleet-wide; I have only verified these four keys on this one step.

**Cheap check that would have caught it:** grep the key in the Go source before calling a config
change a win. One command. Logged in `WRONG_CALLS.md`.
