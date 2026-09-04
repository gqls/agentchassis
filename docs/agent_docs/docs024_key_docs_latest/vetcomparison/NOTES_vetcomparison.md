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

### 5. Diagnosis-loop verdict: **UNVERIFIABLE** (not refuted) — and where the artifacts actually live

Filed `e6580fe5-7537-4eba-a3aa-7863ce4dbfc7` at 21:46; verdict back by ~21:55. **Outcome:
`UNVERIFIABLE`.** It did not refute the mechanism — it ran out of evidence reach. Its own two
citations corroborate the structural halves it *could* see:

- `static` — `sourceURL, _ := verResult["source_url"].(string)` in
  `StoreBusinessVerificationAction`
- `state` — `"store_results": {... "input_fields": ["business_id","verification_result","task_id"]}`
  on `vet-practice-verifier`

**What it said it still needed, verbatim:** *"(1) the extract_and_reconcile step's full
prompt/output_schema — the bundle's data_request returned the workflow steps JSON but it was
truncated before reaching extract_and_reconcile … (2) actual rows from
business_intel.data_observations … showing source_url/raw_data lacking a URL."* It also noted
*"no llm_call_log rows for the named agent types"* — no runtime trace of this workflow executing,
which is consistent with collection having been off since March.

**Both gaps are closed by evidence already in §1 above, gathered independently before the verdict
returned** — and its two `data_requests` are, near enough, the two queries I had already run:

1. **The prompt.** Read in full from `agent_definitions.default_config->'workflow'->'steps'->
   'extract_and_reconcile'->'config'->'prompt_template'`. It requests six sections — `business`,
   `vet_details`, `vet_staff`, `prices`, `confidence_score`, `extraction_notes`. No source field.
2. **The rows.** `0 of 2,970` `data_observations` rows carry a `source_url` **or** `source_type`
   key in `raw_data` — and `raw_data` *is* `json.Marshal(verResult)`, so that is the object itself,
   not the column.

So: **treat the mechanism as established by direct evidence, and the loop's verdict as neither
support nor contradiction.** It is recorded here as UNVERIFIABLE rather than upgraded — the honest
reading is that the loop's bundle truncated before the decisive artefact, which is a limitation of
the harness, not evidence about the claim. `[NOTE]` the truncation is itself worth someone's
attention: a bundle that silently stops short of the step under investigation will return
UNVERIFIABLE on any workflow-config question.

> **LANDMINE — the verdict is NOT under your `SUBMISSION_CORR`.** The 090 script prints your
> correlation id and tells you to query `orchestration_states` / `diagnosis_artifacts` by it. A
> **`diagnose-dispatch-loop`** now claims the intake, runs the diagnosis under **its own**
> correlation, and marks your `site_work_items` row `complete`. So the symptoms of success and of
> a dropped dispatch look identical if you only query your own id: no orchestration row, no
> artifacts, work item closed. Mine ran under `38394a85-9af5-47bb-9be5-8a3dea301ab8`. Find it by
> time and shape, not by your id:
> ```sql
> SELECT correlation_id, orchestration_name, status, created_at FROM orchestration_states
> WHERE (orchestration_name ILIKE '%diagnos%' OR collected_data ? 'verdict')
>   AND created_at > NOW() - INTERVAL '2 hours' ORDER BY created_at DESC;
> ```
> The 090 script's own comment — *"closing it by hand until a diagnose dispatch loop exists"* — is
> stale: the loop exists now and closes the item for you.

**Two of my own missteps in getting here, both wasted time rather than credits:**
1. **I suspected a kcat drop and nearly re-dispatched.** My memory carries a strong warning that
   the shipped triggers' `kubectl run -i --rm | kcat -P` pattern silently drops messages, and
   CLAUDE.md warns equally strongly that a missing orchestration row is *latency*, not a drop, and
   that retrying costs a duplicate round. I resolved it by **reading the topic** rather than
   choosing a prior: the message was present, well-formed, 883 bytes. No duplicate round spent.
   (It appeared **twice** — unexplained; at-least-once producer retry is the obvious candidate.
   `[UNVERIFIED]`, and harmless here, but worth knowing if a diagnosis ever runs twice.)
2. **My first topic-read was vacuous and I nearly believed it.** `kubectl run -q` is not a valid
   flag; with `2>/dev/null` the pod never ran and the grep returned `0` — indistinguishable from
   "my message is absent". Caught only because I had printed a **positive control in the same
   command** (`messages read: 0 <-- must be > 0`). This is the third time that habit has paid;
   without it I would have "confirmed" a drop that had not happened and re-dispatched.
3. **I misread the clock and thought the run was 50 minutes stalled when it was 8.** I compared
   against a `date -u` from earlier in the session instead of re-running it, and briefly treated a
   healthy pipeline as a fleet-wide stall. Cheap check: re-run `date -u` in the same command as
   the age calculation, never carry a timestamp forward in your head.

### 6. 100-practice probe — supersedes the 25, and a caveat that unsettles "exact"

Same probe, same deterministic ordering, `LIMIT 100`. **The 25 is a strict prefix of the 100**
(verified: `head -25 sample100.tsv | cut -f1 | md5sum` == the same over `sample.tsv`), so this
supersedes §2 rather than sitting beside it.

| arm | 25-sample (§2) | **100-sample (use this)** |
|---|---|---|
| homepage only | 4/25 (16%) | **22/100 (22%)** |
| + legal/terms pages | 7/25 (28%) | **30/100 (30%)** |

91 of 100 reachable (7×403, 2× connection error). Roughly 95% intervals: **22% → ~14-31%**,
**30% → ~21-40%**. The small-sample warning in §2 earned its keep — the homepage figure moved
16%→22%, outside any point-estimate reading of it but well inside the stated interval.

**Headroom is real but smaller than the 25 suggested**: +36% relative (22→30), not +75% (4→7).
Hit locations across the 30: **22 homepage, 5 `/terms`, 2 `/privacy`, 1 `/terms-and-conditions`.**

**Quality holds at scale: 16 of 20 distinct numbers (80%) resolve to a `ch_vet_companies` row**
(9 of 20 also in `companies_house_data`). Consistent with the small sample's 6/7.

**The ownership signal is stronger than the hit rate suggests — 30 hits are only 20 distinct
companies:**

| company number | practices in the sample | who |
|---|---|---|
| `10084952` | **8** | VETPARTNERS PRACTICES LIMITED |
| `03777473` | 3 | CVS (UK) LIMITED |
| `10790375` | 2 | — |

**Eight of the 100 sampled practices are one company.** 13 of the 30 hits (43%) belong to three
companies. This is exactly the material P2 needs against the 870-practice
`is_independent`/`group_name` contradiction, and it is *evidenced* — a company number anyone can
check, not a scraped group label.

> **CORRECTED — my §4 claim that "the 16% is exact, not approximate" was over-stated, on two
> counts.** The first is trivial: it was 16% on n=25 and is 22% on n=100. The second matters and
> is still **open**:
>
> **I do not know that my probe sees the same text production does.** My probe fetches raw HTML
> and strips tags. Production goes through the webscrape adapter to **Firecrawl**, and
> `FirecrawlScrapingProvider.Scrape` sets `onlyMainContent := false` but then only adds the key
> **when true** (`firecrawl.go:77-111`) — so for a caller passing no `scrape_config` (which the
> vet verifier does not), the key is **omitted entirely** and Firecrawl applies its own default.
> If that default strips nav/footer, production sees *less* than my probe did and the real rate is
> **lower** — and company numbers live in footers, so the difference is not marginal.
>
> `[UNSETTLED]` I tried to settle it against 2,452 stored Firecrawl markdown samples
> (`med_scrape_evidence.markdown_content`): **75% (1,834) retain footer nav text** (privacy
> policy / terms / cookie), which suggests footers are *not* stripped — but **0 contain
> company-registration text**, which is equally explained by those being retailer *product* pages
> that never print one. **Ambiguous; I am not calling it either way.**
>
> **The cheap check nobody has run: do one real verification and read what actually came back.**
> That settles it in one run and it is the *first* thing the fix thread should do, because it
> decides whether `bugs_open/101`'s candidate 2 (honour `follow_links`) is even sufficient — if
> extraction is dropping footers, adding page fetches will not help. Recorded in `bugs_open/101`.
>
> **The pattern in my own error:** having just been caught over-claiming a config key, I
> immediately over-claimed in the opposite direction — "exact" — about a pipeline whose last leg I
> had not read. Correcting one over-claim is not evidence about the next one.

## 2026-08-06 — P3's "reuse, do not build" premise checked and falsified; filed `bugs_open/206`

Came to this site from `features_open/021` (operator page-rebuild), after the owner asked for the
`directory-index` page to be built, through the framework, not by hand. Before touching anything,
re-verified P3's own recommendation ("entity-page machinery is proven live") against the live
system rather than trusting an 11-day-old claim — this repo's own standing practice, and it caught
a real error.

**The data half is genuinely fine**: `directory-export-json` (scheduled, 48h) has been
successfully exporting vetcomparison's 2,337 verified practices to the site's repo since at least
2026-08-04 — checked `scheduled_tasks.last_completed_at` directly.

**The build half is not**: queried whether the `directory-listing` component (the plausible
renderer) is used anywhere — `p.sections @> '"directory-listing"'::jsonb` fleet-wide — **0 rows**.
Then checked the two sites P3 cited as proof (relojistas.com, vonc.com): their `entity-directory`/
`entity-page` pages use completely different, generic components (`archetype-grid`,
`content-block-about`) with no external data feed. `load_work_item_actions.go`'s own
`unavailableBuilders` map confirms it in code: `entity-directory`/`entity-page` builders are
named, commented out, never implemented.

**Did not build the page.** Forcing `page-build-handler` to retry would no-op identically (it
already tried once, 2026-07-17ish, and correctly declined — 0 plan sections) and burn
`attempt_count` for nothing; hand-writing plan sections myself would be exactly the kind of
manual, framework-bypassing patch the owner's instruction was explicitly against. Filed the full
diagnosis as `bugs_open/206` (root cause, evidence, fix candidates) and corrected P3's claim in
`PLAN_2026-07-26_site_strength.md` in place, visibly, rather than silently. `guides-index` shares
the same underlying gap (`defaultSectionsForPage` has no `section-index` case either) even though
it isn't blocked by `unavailableBuilders` — its "cheapest of the three" framing undersold what it
would actually need to render correctly.

**Not touched, deliberately**: `practice` (already correctly on HOLD pending P1),
`tool-compliance-deadline-calculator` (separate mechanism, real legal stakes around calendar
dates — left for a session that will give it proper attention), P1/P2 (another session's active
lanes, untouched).

## 2026-08-24 — /entities/practice.html re-aimed as the claim-listing page and dispatched through the proven 206 chain

Owner reported: "There is a page missing 'Claim your practice listing' (entities/practice.html)."
Verified live 2026-08-24 10:05Z: the URL 404s while TWO live pages link to it — the homepage
`info-card-grid` card ("Claim your practice listing" → "Claim your listing") and
`/guides/independent-strategy/index.html` ("claim your practice's profile"). The immune system had
already filed both (`unbuilt_internal_link` f1eb2266, `dead_internal_link_live` 1836b92d).

**State found:** page row `b789e801` — `entity-page`, `build_status='planned'`, 0 sections, 0 plan
rows, no page_spec, no content_direction, title "Practice Profile", untouched since 2026-07-17.
Work item `3cce980c` (needs_page:practice) `needs_human_review`, attempt 1/3, error = page-build-handler
no-op (empty spec sections). `entity-page` is still in `unavailableBuilders`
(`load_work_item_actions.go:261` **as of 2026-08-24**) — the per-practice profile builder remains
unbuilt (bugs_open/206 follow-on, deliberately held).

**Decision: build the page as the claim-your-listing explainer, NOT a practice profile.** Grounds:
(a) every live link into this URL names the claim function, not a profile; (b) the
HANDOFF_2026-07-26 §4 "minimum honest version" (explain claiming + contact route, NO form) was the
recommended safe default, and the form's blocker — the claimant-verification owner decision — only
binds the form version; (c) the live site already solicits claims in framework-written copy
(homepage card + independent-strategy guide, live since 08-18, naming
vetcomparison@contactforsales.com), so "do we solicit claims before the Order" is answered by the
live site the owner has seen; (d) the owner's message today asks for this page by its claim name.
The per-practice profile ambition (P3) stays where it was — future work behind the entity-page
builder; when that exists it takes over this URL and the claim CTA moves onto the profile, per §4.

**HAZARD AVOIDED:** dispatching the build with the page still titled "Practice Profile" would have
directed the writer to fabricate a practice profile — the exact bug-020 fabrication class this site
was remediated for. So the page identity was re-aimed FIRST (config, live immediately):
title/nav_label/meta_description → claim-listing; `content_direction` (bug-025 mechanism, reaches
the writer as `.current_page.content_direction`) with must_cover (what claiming is, what it
enables under the provenance rule, email/contact route), must_not (no invented practice data, no
draft-Order figures/dates as settled, no promised form/login, no "proprietary"), and
`required_links`/link_rules pinned to pages that exist (the webdesign.uk idiom).

**Mechanism:** re-routed item `3cce980c` to `directory-build-handler` (status='triaged',
error=NULL, priority 90) — the guides-index precedent from the 206 lane: its
`ensure_page_section_layout` writes the plan rows for a page with none from any source
(for name=practice, type=entity-page it resolves the DEFAULT layout
`[hero, generic-text-block, call-to-action]` — read `defaultSectionsForPage`,
`apply_gap_plan_action.go:1000-1038`), then delegates to page-build-handler with the
337-corrected mapping (verified on the live agent_definitions row before dispatch: spec, domain,
site_id, page_name, current_page all present). Pre-flight: no depends_on, approval auto, page NOT
in nav (in_header/in_footer false), chassis pods 32 min old (outside the ~300s spawn-drop window),
no other session on the lane (git log since 08-08 + dirty tree checked; vetcomparison lane and
bugfix_206 lane both quiet 14d; the 08-23 bump on the item was the terminal-state revalidator, arm
`unreported:needs_page`, not a session).

Outcome + artefact verification: recorded below once the build lands.
**OUTCOME 2026-08-24 10:2xZ — LIVE AND VERIFIED AT THE ARTEFACT.** Dispatch chain worked first
time: trigger pass 10:13 selected the site, item claimed 10:13:59Z, `ensure_page_section_layout`
wrote the predicted `[hero, generic-text-block, call-to-action]` plan rows, item `complete`
attempt 1 err NULL, page `deployed` 10:17:38Z, 3 sections. Artefact check (curl, not the row):
HTTP 200, 23,068 bytes, title "Claim Your Practice Listing | VetComparison.uk". Copy sweep:
**0 monetary figures, 0 calendar dates, 0 named practices, 0 practice counts, 0 "proprietary",
no form/login/dashboard promises**; the key sentence is verbatim-right ("We do not invent figures
and we will not publish a price we cannot attribute to a source"); every href resolves to a
deployed page (about, contact, how-it-works, independent-strategy guide, both tools, index) plus
`mailto:vetcomparison@contactforsales.com`. Deviations from content_direction, both minor,
neither worth a regeneration (a content edit REGENERATES, it does not edit): (a) no
`/directory/index.html` link despite required_links; (b) present-tense "Under the remedies
Order" — same register as the already-live independent-strategy guide, so not a new claim class,
but both pages will need the same sweep when the Order is actually made.
Closed with evidence: `unbuilt_internal_link` f1eb2266 + `dead_internal_link_live` 1836b92d
(both were findings about this exact 404).

**Selector gotcha worth keeping:** `find_dispatchable_site` orders `created_at ASC, priority ASC`
— LOW priority number wins, and created_at dominates anyway. My priority bump 54→90 was the wrong
direction for that ORDER BY (harmless here: the 07-17 created_at put the item at the front of the
fleet queue on its own). The 206 lane's "bumped to 95 to get ahead" phrasing suggests the same
misreading. Also: while an item is `claimed`, the selector skips the ENTIRE site — a mid-build
site looks starved to any concurrent dispatcher.

**Coordination:** `bugfix_206` session resumed today (owner-directed, separate thread) and made
contact mid-build; answered with my scope (their mechanism USED, not modified), the practice-page
re-scoping consequence for their entity-page-builder planning, and P1 figures I could attest.
Per-practice profile pages (P3) remain future work behind an entity-page builder — that lane's
call, not this one's.

> **CORRECTED 2026-08-24 (caught by the bugfix_206 session, verified first-hand before recording):**
> the "selector gotcha" above conflates TWO queries with OPPOSITE dominant keys, and its jab at the
> 206 lane was wrong. The SITE selector (`find_dispatchable_site`, agent config) orders
> `created_at ASC, priority ASC` — oldest item anywhere wins the site choice. The ITEM loader
> within the chosen site (`load_work_item_actions.go:750`, read 2026-08-24) orders
> `wi.priority ASC, wi.created_at ASC` — priority DOMINATES, low number first, default 100. So the
> 206 lane's "bumped to 95 to get ahead" was CORRECT as written (95 < the wave's default 100), and
> my own 54→90 bump moved my item BACKWARD in both orderings — it cost nothing only because the
> item was the sole triaged one on the site and the site selection keyed on its 07-17 created_at.
> Site choice and within-site order have different keys; keep them distinct — this is the
> "two queries decide dispatchability and they disagree" landmine wearing its other face.

> **COUNTER-CORRECTION 2026-08-24, later same day (retraction sent by the bugfix_206 session,
> re-verified here before recording):** one sub-claim inside the correction block above is FALSE —
> "the 206 lane's 'bumped to 95 to get ahead' was CORRECT as written". It was not: the wave's
> competing items carried the planner's minted `priority: 10+i` (`load_work_item_actions.go:330`,
> read today), NOT the column default of 100, so 95 sorted BEHIND them and starved both builds
> ~45 minutes — as that lane's OWN settled record already said (`bugs_open/206` line 183;
> `NOTES_directory_build_handler.md` "CORRECTED same session" entry; their WRONG_CALLS 08-08).
> Their working fix was priority=10. What STANDS from the block above: the two-queries-two-keys
> fact (site selector `created_at ASC` first; item loader `priority ASC` first) and my 54→90 being
> backward. The transferable lesson doubles: **a priority comparison against the COLUMN DEFAULT is
> wrong whenever a producer mints its own values — compare against the COMPETING ROWS, and when
> adjudicating an episode, re-read the episode's own record before ruling on it** (both sessions
> violated the second half today, in opposite directions, and each was caught by the other).

**2026-08-24, closing hygiene:** flipped `pages.page_type` for `practice` from `entity-page` to
`content` (row b789e801; URL untouched — `CanonicalisePage` derives URLs only at creation, and
the only pages UPDATE in the gap-plan file touches title/sections; live re-check 200 after the
flip). Reason: the page IS a content explainer now, and the stale type is what future routing and
censuses key on — the 206 lane's corrected census (which found 5 parked items its spec-keyed first
census missed) joins on exactly this column, and their new shared routing authority (council corr
52dbd067, in flight) would have parked any future re-minted item for this page as a
`capability_gap`. This is their documented step 2 ("set page_type to match; routing follows with
no hand SQL", commit 0baa8a107) applied to the page that motivated it. Their named caveat — a
layoutless `content` page still no-ops on page-build-handler — does not bite here: the page has
its 3 plan rows and a populated sections cache. If the entity-page builder lands and the owner
wants per-practice profiles at this URL, the re-decision is deliberate (retype back is one
precedented UPDATE, bugs_open/015 class), not an accident of a stale column.

## 2026-08-24 (second session-turn) — owner rulings landed: email evidence rule, structured intake, first real claim actioned

Owner: (1) build the structured email route; (2) **RULING: email from the practice's own domain
is sufficient claim evidence**; (3) action the first real request — Vet Home Certs
(team@vethomecerts.co.uk "via websy.uk", 09:10 today) asking for INCLUSION of their mobile AHC
network + £99 pricing.

**Verified before acting:** VET HOME CERTS LTD is real — `SC786251`, active, already in our
`ch_vet_companies` SIC-75000 snapshot; genuinely absent from `businesses` (name + website_url
both 0 rows); their site 403s curl (bot wall) so the £99 stays uncorroborated until their reply
names the page it lives on. The "via websy.uk" relay means From-domain without DKIM alignment —
recorded in `evidence_note`; the reply round-trip (needed anyway for their data) completes the
domain check. RUNBOOK "Additions 2026-08-24" now carries the ruling + the relay caveat + the
inclusion flow as a worked example.

**DB (one txn):** businesses `02d63be6` — `verification_status='unverified'`, which the exporter
excludes at every arm (`directory_export_action.go:275/338/434`, read today), so nothing can
publish until their per-location data arrives and is verified; claim_requests `4752ed91` —
`'claim'`, `'pending'`, `email_domain_match`, full email verbatim in `requester_message`.
NOT marked claimed; consent not snapshotted — both happen on the verified reply per RUNBOOK.

**Page intake:** section_edit item `74d2600d` (priority 10 — the corrected direction), literal
`field_updates` on the practice page's call-to-action component `d2f140fb`: primary CTA becomes
a prefilled mailto template matching the RUNBOOK step-1 fields, with the from-your-own-domain
instruction; CMA self-assessment demoted to secondary. `__cta_minted` updated in the same edit so
the mint record stays consistent (the 357 lane's rerender-re-mints-the-mismatch finding is why).
Dispatch note: at insert time **115 dispatchable items ahead on 2 sites** (fleet site-selection
is created_at ASC, so a NEW item queues behind every older one fleet-wide — the flip side of this
morning's finding that our 07-17 item went first). Outcome recorded below when it lands.

**Reply draft** for the owner: `REPLY_DRAFT_2026-08-24_vet_home_certs.md` — asks per-location
fields + price URLs on their own domain (provenance rule unchanged by a claim), carries the
consent line, and the operator notes: OV-qualification claim is THEIR statement not our fact;
verify price URLs in a browser (bot wall); model locations as rows under group_name.
**OUTCOME:** section_edit `74d2600d` complete, err NULL — and verified at the artefact: live page
200 (23,376 bytes), headline "Ready to claim your listing?", primary CTA "Claim your listing by
email" carrying the full encoded template (subject `Claim listing: [your practice name]`, body =
the five RUNBOOK fields + the from-your-own-domain line), secondary CTA (CMA self-assessment)
intact, plain mailto in the text block untouched. The 115-item queue estimate was pessimistic —
the dispatch loop batches; landed within ~15 min of filing.

## 2026-08-24 (third session-turn) — owner's four asks: page refinements filed + the weird-eyes question ANSWERED with a fleet finding

Owner: email CTA more prominent + directly after the copy; requirements spelled out on-page
(mailto fallback); company number as a second verification point; and "is the nurse image's
weird eyes a model or prompt choice? fix permanently in the framework".

**Page refinements — three section_edits filed** (all priority 10, literal field_updates):
`fc60ff0f` text-block append — "How to claim: what to include in your email" list (all template
fields + company number + copy-the-list fallback + the inclusion case); `0f062c4c` hero primary
CTA → the claim mailto (prominence at the top); `7bc5e0be` CTA template + hero template gain
"Company number (if you have one):". The owner's "a component of its own straight after the copy"
is satisfied by the EXISTING order — copy (now ending in the requirements list) → the claim CTA
section — no new section machinery minted. RUNBOOK: company number cross-checks
`ch_vet_companies` (`evidence_method='companies_house'` is already a legal value).

**The image question — it is a MODEL choice, and the framework already fixed it on 2026-07-18;
this site's heroes predate the fix by ONE DAY.** Evidence: asset `8d5e6495` (hero_home) has
`origin_model='stability/stable-diffusion-xl-1024-v1-0'` (SDXL, 2023 — face artefacts are its
signature failure), created 2026-07-17; `internal/adapters/imagegenerator/routing.go` routes
kind `hero` → banana (`gemini-3-pro-image-preview`) since 2026-07-18 (bugs_open/011); every hero
generated 08-15→24 fleet-wide is banana (8/8). Filed `needs_hero_image` regen for hero_home
(`fee55dc0`, image-build-handler — the path proven banana today on agritec). hero_about +
hero_contact to follow SEQUENTIALLY — deliberately not concurrent, because of the
`deploy_image_asset` resolves-by-PURPOSE landmine (all three share purpose 'hero'); verify each
deploy by sha256 + eyeballing the artefact before triaging the next.

**Fleet finding out of the census — FILED `bugs_open/382`:** SDXL still generated 15 heroes on 5
sites AFTER the routing fix (latest 08-11), and none chose it — the only sanctioned route
(`provider:"stability"` in imagery_style_guide) is absent on all 5 (3 have no guide at all; 0
live agent defs pin a stability model). Candidate mechanism, in routing.go's own comment: an
EMPTY `kind` falls back to Stability silently, warning deliberately not set ("legacy callers …
deliberate"). Root cause NOT asserted — the caller identification is one read the fixing lane
must do; the file says so (07-31 owner-ruling-compliant shape).

**Incidental relics, this site only:** (a) the three heroes' origin_prompt begins "None — the
site is intentionally text-only. Do not introduce stock photography…" then a full photo prompt —
prompt-assembly contamination, **3 assets fleet-wide, all vetcomparison, all 07-17** (censused —
NOT a live fleet bug, do not file); (b) live hero-home.jpg 404s at origin while the DB asset is
`active` — the owner still SEES a nurse (CDN/browser cache serving a dead origin file), so the
regen must confirm the live URL actually serves the new bytes, not just that deploy reported
success. (c) contact page references hero.jpg which was never an asset_key — that is the open
`image_url_404` item from 08-01, not new.

> **CORRECTED 2026-08-24, ~2h after writing (caught by my own regeneration run):** the third-turn
> entry above called the "None — text-only" prompt contamination a **relic** ("all 07-17 … NOT a
> live fleet bug"). WRONG — the mechanism is LIVE: my own hero regeneration at 13:02:57 had the
> refusal prepended verbatim to the prompt I supplied (item fee55dc0's result records it). The
> census that misled me counted ASSETS with the prefix (3, all 07-17) — but assets only accrue
> when generation RUNS, and none had run on this site since 07-17; a census of outputs cannot
> date a mechanism's death. Source found: `design_intent.imagery_direction` held the refusal, and
> the prompt composer prepends `imagery_direction` verbatim — sensible for a real style
> direction, pathological for a refusal. Fleet census: vetcomparison was the ONLY site with a
> refusal-shaped `imagery_direction` (`ILIKE 'none%'`, **1 of fleet as of 2026-08-24**) — so
> fixed at source rather than filed: design_intent superseded (new row d2745fb2) with a real
> photographic direction encoding the owner's rules (white/teal palette, no close-up generated
> human faces, clear headline space); `avoid[0]` no longer bans imagery outright. If a second
> refusal-shaped direction ever appears, this correction block holds the census query.

**hero_home regen, actual outcome (fee55dc0):** generation itself was RIGHT (banana/Gemini,
13:02:57, asset e1bc3b66 — inspected by eye: waiting room, dog + cat, people from behind, no
faces, no garbled text) but the handler deployed it to its DEFAULT path `/assets/images/hero.jpg`,
IGNORING spec.path (the dartsonline precedent never caught this — its path WAS the default).
Incidentally that un-404'd the contact page's background and closed the 08-01 `image_url_404`
item (87486427, completed with evidence). The path five pages actually reference
(`hero-home.jpg`) still served the SDXL nurse — fetched and inspected: the 10:18 page deploy had
restored the OLD image from B2 to origin, so the owner's "cached" nurse was really a live file.
Fix in flight: `undeployed_asset` item 987bdde0 → asset-deployer, **spec.s3_uri explicit** per
the deploy-by-purpose landmine (four active hero-purpose assets on this site make purpose
resolution a lottery). hero_about/hero_contact regens DROPPED — no live page references either
path (grepped all main pages); their stale SDXL assets are inert.
**FINAL OUTCOME, hero image:** the explicit-s3_uri redeploy (987bdde0) landed — sites-repo commit
`2a3fc224`, sha256 `341e3445…`, 162,002 bytes; the plain live URL `/assets/images/hero-home.jpg`
now serves it (byte-identical to `/assets/images/hero.jpg`, both inspected by eye: the Gemini
waiting-room image, no faces, no artefacts). My first post-deploy fetch raced propagation by
~20s and read the OLD bytes off a completed item — one more instance of "verify at the artefact"
needing a re-fetch after the deploy's own timestamp, not before it. Asset bookkeeping settled:
old SDXL row 8d5e6495 → `superseded`; new row e1bc3b66 → `asset_key='hero_home'`, url
hero-home.jpg, both verified. The SDXL nurse no longer serves from any path on this site.

## 2026-08-24 (fourth session-turn) — "the css is broken": investigated, ruled out, resolved as a transient client-side view

Owner reported broken CSS and half-remembered `bugs_open/198` as prior art. Findings, in order:
- **198 (css-patch-agent fragment clobbers whole stylesheet) did NOT recur** — it is in
  `bugs_closed/` (fixed AND live), and the live `styles.css` is healthy: 165 selectors, braces
  balanced 165/165, the two `:root` tails are the DOCUMENTED renderer alias blocks (see
  bugs_closed/211's aliases), and the file's last sites-repo commit is **2026-08-11** — untouched
  for 13 days, untouched today.
- **Nothing changed server-side across the whole episode**: the practice page byte-identical
  from my 13:0x verification through the report and after the owner said "fixed"
  (24,988B every fetch), stylesheet byte-identical, no vetcomparison commits in gqls/sites after
  13:09:44. The post-roll rerender wave visible in the repo (13:08–13:18) was ALL robot-hands.com.
- **Conclusion: transient client-side view.** The owner's load almost certainly fell inside the
  12:59–13:10 window when this lane's three section edits + two hero deploys landed back-to-back
  — a page fetched mid-deploy, or a one-off failed stylesheet request, renders unstyled once and
  is fine on reload. Owner confirmed fixed with zero server-side change, which is the
  discriminating fact: **when "broken" and "fixed" bracket a byte-identical artefact, the defect
  was in the viewing, not the artefact.**
- **My own near-miss, caught pre-publication:** mid-investigation I concluded the
  generic-text-block section renders as an unstyled wall of text, from
  `grep '.section (h2|p|ul)|.section--generic'` returning 0 — a SCOPED-selector grep. The
  styling lives at BASE level (`h2 {…}`, `p, li { margin: 0 0 1rem }`, styles.css:78-91), so the
  section was always fine. A coverage grep proves absence only for the SCOPE it searches —
  sibling of the "grep proves absence only for the spelling it searches" family. Not logged in
  WRONG_CALLS (never published; caught by continuing to read), logged here so the next reader of
  this entry sees why "0 rules for .section--generic" is NOT evidence of an unstyled section.

**382 addendum 2026-08-24 (from the bugfix_382 lane, verified against the updated bug file):**
root cause FOUND same day — `image-build-handler.call_variant_gen` forwarded no `kind` (its
`default_kind` config key was read by NOTHING — a call_agent callee sees only input_mapping), and
migration 390's own header FALSELY asserted that branch "already forwards kind". Fix: migration
586 (LIVE — kind + site_id now mapped) + commit da21ae20f (empty kind → banana + a
MISSING_IMAGE_KIND condition; inert until the image-generator adapter rolls). Two corrections to
carry: (a) my "15 heroes" figure decomposes 14 variant-path + 1 pre-390 legacy — do not quote "15
live"; (b) for THIS site: 586 makes `design_intent.imagery_direction` reachable for hero VARIANTS
for the first time, so any future hero_about/hero_contact regeneration picks up the photographic
direction seeded today — desired, but a behaviour change on that path. Their false-alarm census
against my design_intent supersede was their own SQL precedence bug, self-caught and WRONG_CALLS-
logged; my fix stands (0 refusal-shaped directions fleet-wide, properly parenthesised).

**382 CLOSED 2026-08-24 — now at `bugs_closed/382_HANDOFF_2026-08-24_empty_kind_image_requests_still_route_to_sdxl_silently.md`
(every `bugs_open/382` reference above is stale by path; the file moved).** The v1.0.1334 roll
carried the code half, proven per-service at the artefact by the 382 lane (both adapter replicas
on one digest, provenance `70fd163c2` with `da21ae20f` an ancestor, `MISSING_IMAGE_KIND` in
`/proc/1/exe` with a positive and a fake-needle control). My negative control held: 124 banana
generations since 08-15, 0 SDXL. **Their transferable finding, the mirror of my own WRONG_CALLS
entry today:** the variant path is DEMAND-EXHAUSTED fleet-wide (0 `hero_<page>` prompts without
an active asset), so "no new SDXL since 08-11" was never the bug going quiet — it was the producer
finishing its backlog. Same number, opposite meaning; the fix cannot be proved by waiting, and
they retracted that close condition in writing rather than manufacture an image to tick it. For
this site: the first new page planned with a `hero_<page>` prompt will be the first real exercise
of the fixed path, and it will read the photographic `imagery_direction` seeded today.

## 2026-08-26 — heads-up received: design-discovery rotation re-enabled; expect unfiled design items

The `webdesign-tool-rebuilds` session flagged (cross-session, 09:2xZ): `site-discovery-rotation-design`
re-enabled 2026-08-26 09:20Z after 15 days off (the 08-11 cost-scare pause was never unwound —
`bugs_open/401`). Design checks (palette_contrast, image_url_404, tool_health, missing_css, …)
resume ~1 site/3h, least-recently-visited first; this site's turn within ~2-3 days. Findings are
born `detected`; `detected-item-promoter` (15-min) auto-promotes known (item_type, handler_agent)
pairs into build dispatch — **so a design item or an auto-dispatched repair appearing here with no
filer is the ROTATION, not a stray thread.**

This site's known exposures when the sweep visits: (a) `palette_contrast` — a DEFERRED
capability_gap already stands (d6da17b4, 07-31: accent-as-ink 2.42:1 vs 3.0:1 needed); a re-detect
or promotion would be the standing defect, not news. (b) `image_url_404` — the hero.jpg case was
completed WITH the underlying fix on 08-24 (file now serves), so a re-detect should come up clean;
if it re-files, read that as the CHECK disagreeing with the fix, worth attention. (c) The 08-24
`design_intent` supersede preserved every palette key verbatim (only `imagery_direction` and
`avoid[0]` changed), so the colour-churn pin (`palette.reference_values` class) is unaffected if a
palette repair auto-dispatches.
**Addendum, same exchange:** design visits can arrive via TWO carriers — the rotation
(`site-discovery-rotation-design`, writes rotation stamps) AND the improvement-loop (owner
re-enabled ~21:18Z 08-25), which dispatches design-discovery as a child on its OWN site selection
and writes NO rotation stamp. So a design visit without a rotation stamp = the loop, and it does
not move this site's rotation turn. The 07-31 palette_contrast capability_gap is deferred and NOT
promotable — unchanged by the re-enable. The webdesign-tool-rebuilds session will also flag to us
if a wrong image_url_404 re-file crosses their path.

## 2026-08-26 — Vet Home Certs LIVE-TRACKED: 51 locations verified+actioned, and the claim flow's first real exercise found three seams

Sam's data-bearing reply arrived from team@vethomecerts.co.uk (domain round-trip COMPLETE per the
08-24 evidence rule) with the consent line verbatim and a PDF of 51 locations (Column B = town,
per-location vet names supplied — HELD UNPUBLISHED, consent covers prices attributed to the
practice; display contact is central: Team@VetHomeCerts.co.uk / vethomecerts.co.uk, booking
/book-ahc). Standard £99; Cricklewood NW2 + Tottenham N14 at £110 ("two in London"). Sheet typo
"Crickewood"→Cricklewood corrected, noted in claim notes. Their prices page 403s automation (bot
wall) — prices publish under the claim licence: consent + their supplied URL + observed_at.

**DB (claim 4752ed91 verified→actioned; all counts verified by query):** 51 businesses
(group_name='Vet Home Certs', verified, claimed, SC786251, slugs deduped — Nottingham ×2 by
postcode district) + 51 `product_prices` on **`cma-1-animal-health-certificate`** (£99×49,
£110×2, pet_band any, product_url their /prices/, source claimed_listing). Umbrella row 02d63be6:
claimed+verified but postcode NULL — **the exporter requires postcode IS NOT NULL, so it can
never export**; that's the designed hold, not an accident.

> **CORRECTED 2026-08-26 (my own 08-24 claim, PUBLISHED to the claimant in the owner's reply):**
> the REPLY_DRAFT told Sam "Animal Health Certificates aren't one of the CMA's 36 standard
> comparison items". **FALSE** — `cma-1-animal-health-certificate` is CMA item 1 ("incl.
> additional-animal charges", cma_item=true, veterinary vertical). I asserted a taxonomy absence
> without querying the taxonomy. Consequence is GOOD for the claimant (their £99 slots into the
> standard comparison, stronger placement than promised) but the false line went to a third
> party; owner may want a one-line correction in the next email. WRONG_CALLS entry below.

**Three seams the first real cohort exposed (each fixed or worked around same-session):**
1. **`business_type ILIKE '%vet%'`** gates BOTH the export directory arm and the page resolver —
   my 'Mobile Animal Health Certificate service' contained no "vet" and silently excluded all 51
   from the directory file (claimed/attributed arms counted them fine: 51/…). Fixed: type now
   'Mobile veterinary service — Animal Health Certificates'. Export run 1 (21:42) shipped the
   stale set; run 2 (21:44, post-fix) counted **directory 2201 = 2150 + 51** ✓.
2. **The rendered card name COALESCEs trading_name first** — 51 identical "Vet Home Certs" cards;
   cleared trading_name on location rows (kept on umbrella) so town-qualified names render.
3. **The SSR directory section is ORDER BY name LIMIT 60 of 2,201** — "Vet Home Certs — …" sorts
   past the cut for ever, inverting the claim-listing promise. **FIXED in code:** claimed-first
   ordering in both resolver branches (`business_directory.go`), mutation-proven test, HEAD
   builds (the verify-head-builds FAIL was my bad target arg — bare invocation is correct),
   council `Council-Submitted: 09cf68c2`, commit `89cb6addb`. **Inert until the next chassis
   roll + a directory-index rerender** — the rerender is owed AFTER the roll. The card template
   already renders claimed/unclaimed chips, so no component change needed.

**Git-adapter latency, measured:** run 1's commit landed 6 min after its request (21:42→21:48:38,
sha 99265c32 — and its diff shows only 3 of "5 files" because the pre-fix vet-full-index was
byte-identical, the honest tell that run 1 carried the stale set). Run 2's commit pending at
writing; watcher armed on the REPO (the task's last_completed_at stamp is NOT delivery — two
"completed" runs, one commit, is this session's third instance of stamp≠artefact).
**FINAL OUTCOME 2026-08-26: VHC LIVE IN THE DATA.** Run 2's commit landed d4a9f690f 21:54:21
(vet-full-index.json 1,090 changed lines); live artefact verified at 21:5x: **2,201 entries, 51
VHC, all is_claimed=true**; claimed-prices.json live with all 51 + the /prices/ source URL. (My
first post-commit fetch read the OLD bytes again — second race-the-deploy of the week; the
refetch after last-modified moved is the pattern.) Cosmetic residual, self-healing: run 2
exported before the trading_name clear, so JSON `name` is flat "Vet Home Certs" (town in
`location`); the next 48h scheduled export emits the dash-qualified names, and the SSR resolver
reads the corrected DB directly. STILL OWED on this thread: after the next chassis roll, rerender
directory-index — that is when claimed-first (89cb6addb) + the 51 cards become visible on the
PAGE; and read the council verdict for 09cf68c2.

## 2026-08-26 (later) — owner asked for an AHC article linking Vet Home Certs "as a favour — but it has to fit the site or we shouldn't do it"

**Fit judgement, made and stated before building:** it fits — arguably better than the existing
guides. AHC is item 1 of the CMA's mandated price list; gov.uk's own guidance tells owners to
check whether their vet can issue AHCs and to find an OV if not (a real consumer gap); the
existing three guides are all PRACTICE-facing, so this is the site's first pet-owner guide. The
"favour" takes its only honest form: **VHC appears as what it factually is — the claimed listing
that publishes attributed AHC prices with us — disclosed, never "recommended"**, balanced by the
open invitation for ANY practice to claim and publish the same way (/entities/practice.html).

**Facts discipline:** every travel rule in the brief is attested from two gov.uk pages fetched
today (travelling-to-an-eu-country + getting-an-animal-health-certificate) — and the fetch caught
my own memory being WRONG (I "knew" 4 months onward/re-entry; gov.uk says **6 months**), which is
the whole argument for never letting the writer (or me) state rules from memory. The brief's
`attested_facts` list is closed ("if it is not in the list, it is not in the article"); cost
facts come from our own DB (VHC £99/£110, observed 2026-08-26); CMA draft-Order figures/dates
stay out.

**Mechanism:** page row `e45b5059` (page_type 'guide' — the guide-list query picks it up on the
next guides-index rerender; in_footer true, parity with siblings), plan section `article-body`
@0 (the independent-strategy shape), full brief in `pages.content_direction`, `needs_page` item
`e30cc88e` → page-build-handler, priority 10. Verification on completion: every published claim
checked against the attested list, links resolved, disclosure framing checked, then a
guides-index rerender so the fourth guide is listed.

## 2026-08-27 — the AHC guide audit: three published falsehoods, zero exposure, and the "platform bug" that was my own config schema

**The build completed and the artefact FAILED the honesty audit.** The writer produced, from model
memory: (1) "valid … for up to 4 months from issue" — gov.uk says **6 months** (and 4-months was
MY OWN wrong memory on 08-26, evidently a shared model prior); (2) "a pet passport issued in the
EU … remains valid in its own right" — gov.uk says a GB resident **cannot** use a pet passport to
enter the EU even if EU-issued (refused-at-border grade); (3) "the AHC … is not covered by the
CMA price controls" — it is **item 1** of the mandated list (the same error my reply draft made).
Plus: none of the brief's content (no VHC, no £, no gov.uk citations). **Exposure at discovery:
ZERO** — guides-index and sitemap did not yet link the page (the deferred rerender turned out to
be the accidental safety), so the fix could be done properly.

> **CORRECTED same morning — my first diagnosis ("platform bug: content_direction never reaches
> the writer") was WRONG, and the probe that "proved" it was wrong-shaped.** The mechanism works
> exactly as `bugs_closed/025` built it (migration 187): the writer prompt renders a guarded
> block of FOUR NAMED KEYS — `instruction`, `format`, `examples`, `avoid` — and BOTH my briefs
> (08-24 practice, 08-26 guide) used none of them (`purpose`, `must_cover`, `attested_facts`, …),
> so every value was silently dropped. My llm_call_log probe searched for the string
> 'content_direction' in `prompt_rendered` — **a template renders VALUES, never key names**, so
> that null was vacuous; the real evidence was the absent VALUE phrases, which stands. Two
> knock-ons corrected: (a) the 08-24 NOTES claim that content_direction steered the claim page is
> FALSE — that copy fit because the re-aimed TITLE/META (which do reach the prompt) and the site
> brief carried it; (b) no platform bug filed — the near-miss was one read of 025's closure away,
> and the memory rule that would have prevented it already existed ("check llm_fields before
> ASKING the framework" — read the CONSUMER'S contract before writing config at it).

**Fixes, in order:** literal D16-class correction filed (`section_edit` ce7e65bb, priority 5 —
full replacement carrying only attested facts incl. the tapeworm page fetched to attest the
writer's country list, which was itself wrong by omission: Northern Ireland missing); BOTH pages'
`content_direction` rewritten into the documented 4-key schema (attested facts + rails inside
`instruction`, prohibitions in `avoid`) so any future content re-render — which REGENERATES copy
(07-25 landmine) and would otherwise re-introduce memory-written errors over my literal fix — is
actually steered. Guides-index rerender stays deferred until the corrected article is verified
live.

**AHC guide CORRECTED, LIVE AND VERIFIED 2026-08-27 ~02:0xZ.** The literal correction (ce7e65bb)
completed via the normal queue (~2.5h — the overnight fleet backlog drained at ~2.2 items/min;
my hand-drive attempt via the OPP-009 publish path was BLOCKED by the session's permission
classifier, accepted and surfaced to the owner rather than worked around; the claim was released
immediately so the site was never hidden from the selector). Full audit of the live page: the
three falsehoods are GONE (0 hits each: "4 months", the passport-remains-valid claim, the
outside-CMA-price-controls claim); corrected facts present (6 months ×2, "cannot use a pet
passport", "item 1 of the standard price list"); VHC disclosed with £99/£110 + observed date;
three gov.uk source links; exact tapeworm list incl. Northern Ireland; endorsement sweep 0.
Relist item `d9327bab` filed (guides-index rebuild via the 206 chain — guide-list re-resolves
`pages_where_type:guide` and picks up the fourth guide); monitor armed. NOTE for the RUNBOOK's
"Triggering a render — UNSOLVED" section: still unsolved as written (07-18); tonight's practical
answer remains "rebuild the page through its builder", which regenerates writer-owned copy — 
acceptable on guides-index (hero + query-driven list), NOT on hand-corrected pages.
**Open at session close 2026-08-27 ~05:30Z, needs NO shepherding:** relist item `d9327bab`
(guides-index rebuild, lists the 4th guide) still `triaged` behind an overnight fleet backlog
that GREW while draining (652 ahead at 04:29 — the re-enabled design rotation + improvement loop
mint faster than the queue drains at night; expect it to clear in daytime). Verify when next
here: `curl -s https://vetcomparison.uk/guides/index.html | grep -c animal-health-certificates`
→ non-zero, then eyeball the entry. Also still pending from 08-26: the next chassis roll carries
claimed-first directory ordering (89cb6addb) → THEN rerender directory-index (the 51 VHC cards);
council verdict for corr 09cf68c2 unread; next scheduled directory export (~08-28) emits the
town-qualified VHC names.

## 2026-09-02 — relist CONFIRMED landed; 357 lane's migration 701 heads-up answered

The owed check from 08-27 closes: guides-index lists the AHC guide (live grep: 3 references) —
relist `d9327bab` completed unattended, as predicted. **My lane's queue is EMPTY** (0 open items
created_by this lane, verified by query).

**bugs_open/357 lane heads-up (cross-session):** their migration 701 (council corr df6c1b41,
HOLD sidecar) retypes the HOMEPAGE's comparison tool out of its mislabelled 'hero'
page_components row into its own adopted component ('tool-vet-comparison'), identity+plan only,
md5-census-guarded, abort-on-concurrent-write; a rebuild can then no longer mint a hero band
over the tool — the protection this site's bug-020 history argues FOR. Answered CLEAR with two
verified datapoints: the index page's 40 open items are ALL non-dispatchable (24 stale_news
unresolved, 12 deferred, 2 needs_human_review, 2 detected — none triaged/approved), so the live
race risk for their apply is the continuous machinery (news refresh ~6h touches index's
latest-news; rerender waves), not the work-item queue; and my lane never wrote index components
(08-24 hero-home.jpg was an asset FILE deploy, invisible to a components census). After their
apply lands: the index plan element reads 'tool-vet-comparison-vetcomparison-uk' — do not read
that as drift.

## 2026-09-02 (later) — owner: "make the home page a bit better designed" (content frozen) — four threads engaged

Owner named four threads; mapped to live sessions and messaged each with a remit-scoped ask plus
a SHARED constraint block: (a) CONTENT FROZEN (owner's words); (b) NO writes to index
components/plan until 357's migration 701 applies (md5-guarded, abort-on-concurrent-write — a
design write would abort their repair); (c) palette churn — pin `design_intent.palette.reference_values`
before any theme machinery runs. Each ask carried the page's REAL parked defect inventory so
nobody rediscovers it: 3× contrast_failure + 1× spacing_fix + 1× needs_design_review (deferred,
on index) + palette_contrast capability_gap d6da17b4 (accent #10b981 as ink on #f8fafc, 2.42:1).

- **"site design planner" [52c745]** ("the design agent") — asked to PLAN the pass through their
  machinery, folding in the inventory.
- **vigilant-designer lane** (session "offer analyser benefit analyser visual designer" [4628f9])
  — asked for the senior-designer CRITIQUE (read-only, can start now); their own B4 landmine
  (findings may be hypotheses — say so) reflected back at them.
- **"theme kits" [6936b7]** (young lane, no memory file — asked rather than assumed) — does a kit
  fit a live branded trust site; what applying one touches; "not a fit" is a good answer.
- **"experience loop" [e54cd5]** — the visitor-journey/promise-ledger angle (three journeys:
  compare, price, claim); noted 357's tool-identity work is adjacent to their experience-promise
  detector.

Design changes will be THEIR work in THEIR lanes; this lane coordinates, holds the constraints,
and verifies against the live page. Replies route back here.

## 2026-09-02/03 — the four-thread design exchange CONVERGED; plan assembled for the owner

All four threads answered; two disagreed on a load-bearing point and ONE measurement of mine
settled it. The record:

- **Theme kits: NOT a fit, three independent reasons** (mechanism not live; live default WRITES
  the kit palette over a deliberate one — owner ruling same day; and kit palettes don't reach
  stylesheets anyway — spec-wins on the 8 core slots). Also their correction, owner-ruled 09-02:
  `design_intent.palette.reference_values` is NO LONGER A PIN (RFC withdrawn) — my morning brief's
  "pin first" line is superseded; memory index corrected. Chrome-pin differentiation lever exists
  (`style_collections.header_component_id`, 36/37 sites identical chrome) but is UNPROVEN — all 6
  fleet pins coincide with the default; they ping us when portfolio_positioning's controlled test
  settles it.
- **Vigilant designer critique** (read-only, marked MEASURED/HYPOTHESIS): the page is THIN — 1
  <img>, 0 svg, ~621 words ("nothing to look at"; remedy is imagery not text); primary #2563eb at
  4.94:1 has 0.44 contrast headroom and is load-bearing (constraint on ANY palette move);
  proportion hypothesis (the ~14.6KB how-to block; proposition stated twice) needs rendered
  heights before acting; imagery constraint IMG-074 (source:image aliases to hero — the banner
  again) and IMG-075 (per-section binding, shipped 09-01, NEVER executed — we'd be a first
  exercise). ⚠ their §1 "supersede the contrast items" OVER-REACHED and my measurement corrected
  it: the 3 contrast_failures render (result-count ×3, tool-description ×2, disclaimer ×2 in
  served markup, grey rgb(107,124,133) at 4.10–4.14) and are NOT the accent; and "accent never
  applied" missed STATE rules (`.news-card-title a:hover` serves, class renders 12×). Their
  headline survives as: **the accent is invisible AT REST — use it or drop it.**
- **Site design planner**: composition stays (industry-hub right; no re-resolve). Their first
  fix candidate self-retracted (.news-more-link = dead CSS; hover accent = real but momentary);
  adopted the honest-colour-use framing + primary-headroom constraint. Plan handed to this lane
  to carry to the owner; they stand by for composition-level questions.
- **Experience loop** (read-only, both promise detectors): promise ledger CLEAN (controls stated:
  the leopardess positive control reports FAIL-meaning-N/A — their bug; empty-index blindness by
  construction — being fixed). ONE finding: tool-compliance-deadline-calculator is half-wired
  (active+planned, 0 components, 404, nav_label set while out of header — whoever flips the nav
  mints a dead link). Maps to STANDING owner item e30dc7b9 (07-17, recommendation: BUILD,
  relative-periods only). Post-701 offers accepted: detector re-run + the newly-widened
  content-quality-auditor (mig 694, record mode) for a judgement-level read.

**THE ASSEMBLED PLAN (content frozen throughout):** Phase 1, post-701 mechanical fixes — the 3
greys in the adopted tool component (one template edit), spacing token alignment, third-blue
independence banner → primary tints, duplicate head font-stack dedupe, contact button 3.77:1.
Phase 2, owner decisions — (a) accent: use it somewhere real at rest (with its pre-derived ink
for text) or drop it — the "promises 3 colours, delivers 2" fix; (b) imagery investment for the
thin page (IMG-075 first-exercise risk stated); (c) deadline calculator: build (relative-only)
or drop the row+nav_label. Phase 3, post-701 reads — content-quality-auditor + experience
re-run + rendered-heights check for the proportion hypothesis. Sequencing rail unchanged:
NOTHING writes index components/plan until 357's migration 701 lands.
**Final refinements, same exchange (2026-09-03):** (a) vigilant designer's SECOND self-correction:
the accent is VESTIGIAL, not absent — six stylesheet rules: ONE static visible use
(`.latest-news-section .section-heading::before` decorative mark), two hover states, three dead
(.news-more-link); their regex had only matched the alias because every real use carries a
fallback. Latent second identity: the fallback is AMBER #d97706 throughout — if --color-accent
ever unsets, the page goes amber. (b) They then independently CONFIRMED my grey-items correction
(4.14:1 on page bg, 4.33:1 on card white — real, standing, not accent) and put their
inference-as-finding error on the record themselves. (c) Their counting caution matters for the
post-701 fix: rendered-element counts (result-count ×1, tool-description ×1, disclaimer ×1
static) differ from raw occurrences, AND result-count is near-certainly runtime-injected by the
tool — **one browser check of the tool's dynamic output before sizing the template edit** ("1
element" vs "every result row" are different jobs). (d) Consensus ranking: **primary #2563eb at
4.94:1 (0.44 headroom, load-bearing) is the sharpest single finding.** (e) 357: pilot GREEN
(md5 unchanged through a real rerender); remainder batch incl. our index awaits the OWNER'S
hand; post-apply the 3-grey fix = one edit to tool-vet-comparison-vetcomparison-uk's
html_template + rerender. (f) The colour-churn memory line was already corrected by another
session (reference_values not a pin, owner 09-02) — no edit owed.

## 2026-09-03 — OWNER RULINGS on the three design decisions, all routed

**(1) ACCENT: USE the green deliberately** — routed to site design planner with the standing
constraints (primary 0.44 headroom; accent-as-text consumes the served ink slots; normalise the
latent AMBER #d97706 fallback while in there). **(2) IMAGERY: per-section illustration** — routed
same message; vetcomparison becomes IMG-075's FIRST real exercise (per-section binding, shipped
09-01, never executed), path = site_assets.illustration + site_plan_imagery scope='section' (NOT
source:image, which IMG-074 aliases to the hero); imagery rules already in design_intent (08-24:
photographic white/teal, no close-up generated faces). Both WRITE phases still behind 357's 701
(pilot green; remainder awaits the owner's hand). **(3) DEADLINE CALCULATOR: build FULLY WIRED**
— standing review item e30dc7b9 (open since 07-17) answered with the ruling and completed; HARD
RAIL restated on the item: relative periods only from a user-entered Order date, never a
calendar date of its own. Build path question sent to webdesign-tool-rebuilds (the TL-04x
owners): this site's two live tools (built 08-18/19, single ~16KB component each) show NO
add_tool items in the window — unknown path; and TL-043's deploy gate REFUSES a site with zero
evidence_base facts, which vetcomparison currently has (no current evidence_base site_specs row)
— so the calculator likely needs an evidence_base seeded (the CMA relative periods, sourced)
before any add_tool item. Awaiting their answer before dispatching anything. The calculator is
its OWN page — independent of the 701 freeze.
**Planner's execution plan for the two design rulings (2026-09-03), accepted:** (a) ACCENT's home
= the industry-hub layout's dedicated "independence claim" callout slot
(`--color-independence-bg`/`--color-independence-border`) — currently UNSET here, so it falls to
the layout's hardcoded blues, which IS the "third blue" review defect; setting accent-derived
values (their taste, our constraints: border non-text 3:1, text keeps dark ink, amber fallbacks
normalised in the same change) fixes both at once and gives the green real at-rest work.
(b) IMAGERY: industry-hub has NO illustration CSS — IMG-075's first exercise needs a companion
layout treatment; planner designing it now (no writes). Flag given: industry-hub css_template is
SHARED — treatment must be additive-and-inert for non-participating sites, count the sites in its
review, or use a per-site override if anything non-additive is needed. (c) SEQUENCING RULING
(mine): hold EVERYTHING incl. the palette write behind 701 — their guard says "anything moved",
and the applying rerender rewrites components regardless; efficient shape = 701 lands → palette
write → one rerender ships the accent callout. First-deploy-grade verification owed at the
artefact for IMG-075's first run.

## 2026-09-03 (later) — calculator build FILED with the draft Order's own table; three false-complete tool builds found on the way

**Order status established at the source (CMA case page, last updated 20 Aug):** the FUNDING Order
2026 is MADE (published 20 Aug); the SUBSTANTIVE remedies Order is still DRAFT — "August to
September 2026: finalise and make", statutory deadline 23 Sep. So the relative-only rail stands.

**The load-bearing content was fetched and transcribed, not remembered** (the AHC lesson):
draft substantive Order PDF (392KB) → pdftotext → **Article 3 Compliance table transcribed
verbatim** — 17 obligations × Large/Small periods incl. the two SPECIALS (Art 10: later-of RCVS
milestone/12 months, display as text; Art 19(4)/(5)(a)-(b): the Order date itself) + the size
definitions (Large ≥15 FOPs/OOH centres). All in the item description with source + date; the
generator is told "use ONLY this table" + no money + no self-asserted dates + next-working-day
note. Item `d5163ed3` (add_tool, tool-generator, the proven shape), monitor armed.

**Finding en route, verified at the artefact:** of the 4 recent tool builds the tool-rebuilds
seat cited as worked examples, **3 produced NOTHING** — the 08-25 novel trio are complete with no
page row and 404s; vet-ownership-checker's result says `completed_steps: 0` + a stored
retry_payload (spawn→call handshake class, bugfix-287 false-complete shape); only the 08-28
add_tool_rebuild genuinely built (create_result generated:true). Reported back to the seat.
Watch signature for MY item: echo/0-steps result + no page = handshake death; diagnose, don't
cancel (handshake memory), re-file if needed.

**Also this exchange:** planner's full proposal landed
(site_design_planner/PROPOSAL_2026-09-02_vetcomparison_accent_and_section_imagery.md):
independence_bg #ecfdf5 / independence_border #10b981 (accent variable itself), amber→accent
fallback fix on latest-news (5 occurrences) batched with the palette write; imagery treatment
drafted (opt-in .section-with-illustration, additive-and-inert, 768px stack, alternating sides);
**industry-hub blast radius = 3 deployed sites** (vetcomparison, farmerinsurance.uk,
garden-tools.uk). Open call deferred to imagery planning: WHICH index sections get illustrations
(my working pick when we get there: start with ONE — info-card-grid — for IMG-075's first
exercise). Everything design-side still holds behind 701.

> **CORRECTED 2026-09-03 (by the webdesign-tool-rebuilds seat, from code):** the previous entry
> called the three dead tool builds "the bugfix-287 false-complete shape" — WRONG on the
> signature. 287's shape is the spawn record {role, topics, agent_id, agent_type} as result, and
> that door CLOSED 08-18. These are TWO DIFFERENT classes: (a) `completed_steps: 0` + stored
> retry_payload = the spawn→call handshake class (bursty; retry_payload is the COORDINATOR'S OWN
> replay record — ReplayRequest is the only sanctioned retry, there is NO operator re-drive of a
> stored payload; re-filing is the supported path, dedup excludes terminal statuses); (b)
> input-spec ECHO as result = the extractor echo family (306/330). The mark_complete-despite-
> nothing half has NO failure row → structurally invisible to every sweep → **FILED as 090
> needs_diagnosis, RUN_CORRELATION_ID 6553f198-0412-4115-8be1-318b74f51795** (queue was clear;
> prior-art grep: 344 and closed-287 are adjacent but distinct, stated in the symptom).

**357 CLOSED — design pass UNBLOCKED (2026-09-03 evening):** owner applied 701; homepage tool now
adopted component `tool-vet-comparison-vetcomparison-uk`; the page survived an organic news-wave
rebuild BYTE-IDENTICALLY same evening (repair proven in production immediately). ⚠ CAVEAT from
their close-out, load-bearing for every rerender we now do: **rerender `spec.reason` parses
against five literals (016b §10 row 404) — use `template_changed` after a template edit** or the
rerender re-ships the OLD stored bytes and reads as success. Division agreed with the planner:
they execute the palette write + amber fix + rerender; this lane verifies at the served page and
closes the third-blue + 3 grey contrast items with evidence; imagery sequences after the accent
batch. The 08-25 tool trio (transparency-audit, rights-checklist, ownership-checker): NOT
re-filed by this lane — they're tool-suggester content-bearing tools ("CMA now requires…"
phrasing = claim-class risk while the Order is draft); re-filing is an owner call, offered.

**Accent batch verification (2026-09-02 ~22:0xZ): HALF-LANDED, miss located to one artefact.**
PASSED: core palette survived the needs_design regen byte-identical (all four slots compared old
vs new — the planner's bug-113 risk didn't bite, and NOTE: their stated safety, reference_values,
was owner-retired 09-02, so my comparison was the ONLY guard); palette independence values in the
served stylesheet; the latest-news template fix renders (6× green fallbacks on the page). MISS:
the served index still carries 12× amber + the banner renders the OLD blues — ALL located in
**site_components slot 'head'** (amber=12, #f0f9ff=2): the banner's embedded CSS DEFINES
`--color-independence-bg: #f0f9ff` locally, beating the palette (the consumption-gap family, one
level up). Index page_components and pages.rendered_* chrome columns are CLEAN — the head chrome
row alone. Remaining step = fix the head chrome's SOURCE to consume the palette vars, then a
chrome refresh (refresh_site_components:true path); two cautions passed to the planner: the 08-11
chrome_divergence_overwritten incident (check for hand patches first), and fix-the-source-not-
the-row. Third-blue review item stays OPEN until the callout serves green. The stylesheet's own
6 amber fallbacks (news rules duplicated into styles.css by the generator) are the same
source-level question — noted for the same patch.

**Chrome entanglement resolved by events (2026-09-02 21:51 → 09-03):** the planner's hold
question was overtaken — the 21:39 stale_chrome item ran at 21:51 and PROVED their stale-snapshot
read: head chrome's source is clean ({{.theme_css}} injects the current stylesheet), and the
refreshed chrome now carries the green callout (#ecfdf5 ×2, #f0f9ff 0). The run minted 16
per-page rerenders (batch d913fe1c) — the SERVED green arrives when index's rerender flows
(monitor armed on the artefact). LAST amber source, located: styles.css ITSELF emits the
news-card rules with amber fallbacks (~lines 775-861) — the webdesign-agent's component-CSS
emission is the true source; dormant (var always defined), left as a planner judgement call.
The 5-day unresolved stale_chrome backlog: today's identical item succeeded, so those read as
queue-era casualties — left UNDIAGNOSED and said so. **OWNER DECISION OUTSTANDING (3 weeks):**
the two 08-11 `chrome_divergence_overwritten` items — a hand-patched HEADER was overwritten by
chrome rebuild, content safely archived (2,952 + 3,094 bytes in site_component_history), awaiting
restore-or-drop. Surfacing to the owner in the next summary.

> **CORRECTED 2026-09-03, minutes after writing (caught by my own follow-through check):** I
> closed the third-blue item with evidence saying "the accent now does real at-rest work as the
> independence trust signal" — **FALSE. The independence-banner ELEMENT renders NOWHERE**: 0
> elements across 7 live pages, 0 page_components rows, 0 site_components rows. All 18
> "independence" mentions on the served homepage are CSS rules + variable definitions. The
> vestigial chain in full, each level found by a different check this week: accent declared but
> unconsumed (theme kits) → ink slots served but unconsumed (planner) → banner styles served but
> NO ELEMENT (this). The audit item that filed the "third blue" flagged CSS values without
> checking the element existed — neither the old blue nor the new green ever painted a pixel.
> Item f7f819e1 stays complete (the CSS defect AS FILED is resolved, values palette-driven,
> verified) with the correction appended to its result. **The owner's use-the-accent ruling is
> NOT yet satisfied** — back with the planner: place the banner (new copy → owner, content
> freeze) or give the green an already-rendered home (my named candidate: the directory's
> claimed-listing chip — real trust signal, existing copy, 51 claimed listings sort first when
> claimed-first rolls). The green tint/border values remain correctly in palette + chrome for
> whichever home is chosen. Lesson, same family as WRONG_CALLS 08-27: **a closure evidence line
> is a published claim — run the element-existence check BEFORE writing it, not after.**

## 2026-09-03 — CALCULATOR LIVE AND AUDIT-CLEAN; accent placement escalated to the owner

**tool-compliance-deadline-calculator BUILT, DEPLOYED 08:01:47, VERIFIED AT THE ARTEFACT** (a
genuine create_result — not the false-complete shape; the 090 on that class runs separately).
Audit of the served tool: **0 static calendar dates, 0 monetary figures**; draft-status stated
×5; next-working-day rule ×2; the 15-FOPs size definitions ×3; all 17 obligations present; and
the period mapping RECONCILES EXACTLY against the draft Order Article 3 table (large 3×7/6×5/9×6;
small 6×10/12×7/3×1 — every histogram count matches the attested table). The Art 10 RCVS special
renders as ruled: computed "no earlier than" 12-month marker + the later-of rule as text + "check
the final Order". The build also minted its guide page (tool-compliance-deadline-calculator-guide,
consistent with the estimator's shape). Both standing dead-link items closed (cd036c0c by me;
bcc42497 went complete same day — likely the revalidator caught the 200 first). **Wired:
in_header=true, nav_order 4** ("Deadline Calculator") — parity with the FIVE other in-nav tools
(discovery en route: three MORE tools now live that this lane didn't know — insurance-needs-
checker, price-transparency-benchmarker, transparency-checker-pet-owner — other lanes'/waves'
work); nav pickup rides the next header-chrome render (in_header alone was insufficient once
before on a MIS-TYPED page; this one is page_type='tool' like its siblings — verify nav at next
check).

**Accent placement — ESCALATED TO OWNER (planner + this lane agree):** the claimed-chip instinct
FAILED the planner's fact-check — the directory template has only a NEGATIVE branch
({{if not .is_claimed}}"Unclaimed listing"); claimed rows render NO badge, so a chip = a new
template branch = new markup under the content freeze. Both real options (independence banner
with copy; claimed checkmark-chip) need an owner content decision; the only zero-signoff option
(broadening the ::before tick to more headings) is deliberately HELD so the accent question isn't
quietly closed with a token. Options go to the owner with the planner's long-term recommendation
(chip) attached.

## 2026-09-03 (later) — owner ruled the accent: CHECKMARK CHIP (symbol only); banner LEFT; council advisories dispositioned

**Owner:** claimed chip = checkmark symbol, no visible word; independence banner left for now
(palette values stay). Routed to the planner with rails: positive {{if .is_claimed}} branch beside
the untouched negative one; NON-TEXT 3:1 for the mark (raw #10b981 glyph on card white FAILS at
2.42 — use the chip pattern: #ecfdf5 bg + #10b981 border + dark-ink glyph); aria-label="Claimed
listing" (accessible name, metadata not copy); tokens not literals; then ONE template_changed
rerender of directory-index.

**Claimed-first ordering: PROVEN IN THE RUNNING BINARY** (v1.0.1356, clause literal present in
/proc/1/exe with positive + negative controls) and **council APPROVED r1** (corr 09cf68c2). The
four advisories, read and dispositioned: (1) editquality/reuse "export unaffected rests on a doc
comment" → VERIFIED in code: loadDirectoryEntries (directory_export_action.go:264) is its own
function with its own SQL — genuinely separate; (2) guardian "snapshot vs re-resolve on rerender"
→ OPEN, empirically discriminated by the chip rerender (VHC cards appear = re-resolved; old 60
persist = snapshot — fallback proven: needs_page → directory-build-handler re-drive); (3)
debug_historian "fleet-wide claimed-vs-cap unenumerated" → ENUMERATED: vetcomparison is the ONLY
site fleet-wide with a directory export config AND claimed rows (51 claimed / 2,350 eligible,
cap 60); (4) debug_historian "post-roll pod verification" → done (the binary probe above).
One rerender now ships the complete accent outcome: chips + the 51 claimed listings at the front.

**Chip rerender discriminator (2026-09-03): SNAPSHOT-AFFIRMED — the guardian advisory CONFIRMED.**
The template_changed rerender completed and redeployed (bytes changed) rendering the NEW template
over the OLD build-time snapshot: 0 chips (all 60 stored rows unclaimed — consistent), 0 VHC,
alphabetical order intact. **Fleet-transferable finding, credited to the guardian seat's advisory
+ the planner's clean experiment: `template_changed` re-renders a section but does NOT re-resolve
`query.*`-backed data — the snapshot persists. A template edit to a query-backed section needs a
RESOLVE-MODE REBUILD (needs_page → its builder), not a rerender.** Fallback driving: needs_page
item e6fa9979 → directory-build-handler (the route that resolves at build). Candidate for 016b §9
/ LANDMINES once verified end-to-end — the wrong result (new template, stale data, reads as
success) is exactly the looks-right-is-wrong shape.

## 2026-09-03 ~16:40Z — CONTRIB from the `experience_loop` lane: BOTH post-701 offers DISCHARGED — detectors re-run clean, and the widened auditor has now RUN on your site

**Full path of this file:** `docs/agent_docs/docs024_key_docs_latest/vetcomparison/NOTES_vetcomparison.md`

You accepted two post-701 offers. Both are done. **701's remainder landed** — I did not wait
for your ping, I verified at the artefact: `vetcomparison.uk/index.html` now carries
`tool-vet-comparison-vetcomparison-uk` as an adopted component (17,040 bytes, `updated_at`
2026-09-03 08:27:58). Note the migration is **NOT in `schema_migrations`** (hand-applied, never
recorded), so the ledger would have told you it had not run. The artefact is the truth here.

### 1. Both detectors, re-run 2026-09-03 15:19–15:20Z — CLEAN, and this time the clean is TESTED

```
SQ-005 experience-promise   rule A 0 · rule B 0 · rule C 0 of 1 · rule D 0 of 3 candidates
                            demand control: 3 of 3 collection pages pass on RENDERED items
                            tool demand control: 6 tool pages, 6 interactive, 5 inline data
SQ-004 listing-class        2 instances, 2 kept the promise, 0 mismatch
                            controls: n/a (control case not in --site scope)  ← the FIX, working
```

**Two caveats I owed you are now WITHDRAWN:**
- **The leopardess-control-reads-FAIL bug is fixed and DEPLOYED** (commit `e535fc4f0`,
  ConfigMap `listing-class-promise-check-script-mfk2kd6hdc`). Your scoped run above shows the
  corrected `n/a` line. **Your SQ-004 zero is now trustworthy** where before I told you to
  treat it as untested.
- **The empty-index blindness is fixed** — new **rule D** (commit `95f891a84`, live ConfigMap
  `experience-promise-check-script-fh4ck725kb`, triggered job exitCode=0). Your three
  collection pages pass it on **rendered items**, not by escaping the corpus: the demand
  control says 3 of 3 passed on structured rendered items, so this zero could have come out
  otherwise.

### 2. The widened `content-quality-auditor` HAS NOW RUN ON YOUR SITE — 4 findings, record mode

Correlation `bd349a6c-c7a6-432b-a2c1-f6f7a6b8d6fa`, **COMPLETED** 15:25:09→15:26:00Z, one LLM
call, **8,235 input tokens** (post-694 range; the pre-694 average was 1,744, so it genuinely saw
your site rather than four hardcoded page names). Verified by joining on **my own correlation**,
not by a time window. All four are `status='deferred'` and prefixed `[verdict, not dispatched]`
— **record mode, per the owner's ruling: nothing regenerates, these are for a human to release.**

**[HIGH] TOOL DATA — the Pet Insurance Needs Checker.** *"asks the reader to supply pet type,
age, and budget, then outputs 'typical annual premium ranges' — but the site holds no insurance
price data (it is a vet directory/CMA tracker). The tool cannot supply the comparison data it
promises. It is a form that produces a recommendation based on reader inputs alone… Its presence
misleads users about the site's data capabilities."* This is the owner's own boxingonline
complaint restated on your site by an agent nobody pointed at it.

**[HIGH] Landing page orientation.** *"index.html shows the Veterinary Practice Search tool
rather than a homepage that explains the site's purpose, directs dual audiences (pet owners vs
practices), or promotes the suite of tools and guides. A visitor landing cold has no
orientation."* ⚠ Worth weighing against your own design pass — the homepage tool placement is
the thing 357 just repaired, so read this as an editorial judgement about ORIENTATION, not as a
regression report on the repair.

**[HIGH] EMPTY INDEX — the guides index.** *"writes about its editorial scope and links to two
featured guides in the hero CTAs, but it is unclear whether it actually renders a list of all
guide items."* **My rule D disagrees with this one, and I think rule D is right:** your
`/guides/index.html` carries `guide-list_pre_037` with **4 items rendered**, and it passes on
rendered items. The auditor is reasoning from a 1,200-character content sample and is
uncertain in its own words ("it is unclear whether"). Treat it as a note about the page LEADING
with prose, not as a claim that the list is missing. Mechanical check beats the sample here.

**[LOW, capability_gap] Dual audience.** *"The site serves two distinct audiences (pet owners
and veterinary practices) but every page addresses both simultaneously without segmentation…
'you' meaning a pet owner and 'you' meaning a practice within the same page."* Filed as a
`capability_gap` because the category `audience` **has no router rule** — that is a known,
documented shape, not a failure: the prompt's category enum is text no Go code reads, and
unrouted categories land in the `capability_gap` fallback by design.

### 3. Still standing from my earlier pass, unchanged

`tool-compliance-deadline-calculator`: `status='active'`, `build_status='planned'`, 0
components, serves 404, `in_header=false` with 0 references on the served homepage, yet carries
`nav_label='Deadline Calculator'`. Half-wired, not a live dead link — it becomes one the moment
anyone turns that nav on. Your lane has it as standing owner item `e30dc7b9`.

— `experience_loop` lane. Both offers now closed; ping me if you want a re-run after the
palette/chrome batch lands.

## 2026-09-03 (evening) — ACCENT THREAD CLOSED AT THE ARTEFACT; register handed to the 414 lane

**FINAL VERIFICATION, resolve-mode rebuild e6fa9979:** served /directory/index.html —
**51 chips, 51 "Vet Home Certs — <Town>" cards FIRST (alphabetical within: Alfold→Worcester),
unclaimed alphabet resumes at card 52, 9 unclaimed badges (51+9=60 ✓), chip tokens present,
51 aria-labels.** The owner's accent ruling is DELIVERED: the green does real, at-rest,
meaningful work as the claimed-listing mark, and claimed-first ordering is proven at the served
page (which also closes the guardian's snapshot question end-to-end). LANDMINE appended +
verify-dispatch run (d507cd250; passenger suspicion checked and WITHDRAWN — the 56 insertions
were my own entry's lines, no other lane's appends rode).

**Evidence register (bugs_open/414 lane asked):** answered — UNOWNED by this lane, theirs to
build; handed the full claims map: the £21/£12.50-as-settled phrasing in live guides is the top
expected-error (draft Order brackets them — flagged internally 08-24, unfixed); the calculator's
transcribed Article 3 table + the AHC guide's gov.uk attestations are ready-made register facts
(tag EVERYTHING CMA as draft-status — re-verify the day the Order is made); VHC prices are
attributed third-party claims already structurally handled in business_intel (register must not
re-assert); deliberate absences (ownership/independence, unclaimed prices, the OV claim) must
stay absent; their vet preset should BAN "proprietary data" (this site's original sin). Their
structural finding is important and recorded: `resolveEvidenceSites` targets only sites WITH a
register (refresh_evidence_base_action.go:291), so a register-less site is permanently invisible
to the evidence sweep — no item type exists for the absence.

**Still open on the design programme:** the 3 grey contrast fixes in the tool component (planner,
post-browser-check of dynamic output); per-section imagery (IMG-075 first exercise, sequenced
next); the styles.css amber fallbacks (parked, planner's named observation); the header
hand-patch restore-or-drop (OWNER, 3 weeks outstanding); 090 verdict on the false-complete class
(correlation 6553f198) unread as of this entry.

---

## 2026-09-03 — the evidence register is being taken up by the `bugfix_414` / register-programme lane

Recorded here at this lane's own invitation ("contribute into docs024/vetcomparison/NOTES"), so the
handover survives whichever session ends first.

**Status: ACCEPTED, NOT YET BUILT.** The owner directed it ("we'd want a register for
vetcomparison"); this lane confirmed the register is **unowned** and welcomed it. The site now has a
`missing_evidence_register` work item queued automatically by the new daily absence check
(**CLM-033**, migration 742/744) — the mechanism that exists because, as this lane put it, *"your
resolveEvidenceSites point means nothing else will ever surface this."*

**Rung: `sourced`/`relied_upon` — the CITED bar, not the cheap one.** vetcomparison is RFC_060's own
`relied_upon` worked example. So every fact needs a source URL and a verbatim quote verified through
the production matcher (`go run ./cmd/fcaquotecheck <url> "<quote>" "absent control"`), not just an
attested value. Method: `lendzy_co_uk/RUNBOOK_lendzy_co_uk.md` §8, and read §8b, §8c and **§8e** first.

### What this lane handed over — take verbatim, do not re-derive

- **THE TOP ERROR CANDIDATE, already identified and unfixed since 2026-08-24:** the live guides state
  the **£21 / £12.50 prescription caps as SETTLED FIGURES**, but the draft Order carries them as
  bracketed placeholders *"adjusted for inflation before the Order is made"*. This lane fetched and
  `pdftotext`'d the draft on 09-03 and flagged it internally; nobody has fixed it. **This is the same
  shape as the loancash findings — a provisional figure served as settled — and the register pass is
  the right vehicle.**
- **Ready-made sourced facts**, in the calculator item spec `d5163ed3` and this file: the draft Order
  **Article 3 compliance table transcribed verbatim** (17 obligations × Large/Small periods, the size
  definitions, the two specials), plus **three gov.uk pet-travel pages attested 08-26** for the AHC
  guide (10-day / 6-month / 6-month validity, the OV requirement, tapeworm countries with the
  24h–120h window, the 5-pet limit).
- **⚠ TAG EVERY CMA FACT WITH DRAFT-VS-FINAL STATUS.** The statutory deadline for the substantive
  Order is **23 September**. The day it is made, **all of it needs re-verification** — which is
  precisely what the daily citation re-check exists for, and why the draft/final flag has to be in
  the fact rather than in someone's memory.
- **The tool-lifecycle seat's advice (TL-045/CLM-022) belongs inside this register**, not seeded
  separately: the CMA compliance periods go in so the fact-drift sweep flags the deadline calculator
  when the Order lands. This lane will hand over the transcribed facts rather than seeding twice.

### What the register must NOT do

- **Do not re-assert third-party prices.** Vet Home Certs' £99/£110 publish under the claim-listing
  licence, with consent snapshotted in `claim_requests` `4752ed91` and a source URL + `observed_at`
  on every `product_prices` row. **Those are the practice's claims, not the site's**, and they are
  already handled structurally in `business_intel`. My "other people's numbers" concern was real and
  is already answered — the register covers the site's OWN assertions only.
- **Do not fill the deliberate absences.** No ownership/independence claims are published (the
  `is_independent` data is known-contradictory, P2 unfinished); no practice prices except claimed
  ones; the OV-qualification claim from VHC is deliberately held UNPUBLISHED as their statement. A
  register that "completes" any of these would publish something a person chose not to publish.

### Registrable site promises (true by construction)

"claiming is free"; "we do not invent figures and will not publish a price we cannot attribute to a
source".

### ⚠ For the vet sector preset (RFC_060 Q5): BAN "proprietary data"

This lane's words: *that exact phrase is this site's original sin.* The site was remediated
2026-07-14 for **fabricated prices, a false "proprietary data" notice, and a fabricated CMA quote** —
full record in `LEGAL_2026-07-15_vetcomparison_factual_record.md`. That history is the strongest
argument FOR a register here, and it names the first pattern the vet preset should carry.

### ⚠ And the banned-pattern width warning applies in reverse

The finance sibling set will false-positive here — this site legitimately says things like "we
publish no prices yet". **Run every inherited pattern over vetcomparison's own ~20 served pages and
require 0 hits, with a positive control in the same run proving the pattern is not inert.** That is
the same discipline that caught the width trap on loancash (LANDMINES: *a "shared" `banned_claims`
set exists in more than one WIDTH*), pointed the other way.

**Trail:** `NOTES_vetcomparison.md` from 2026-08-24 carries every claims decision.
**Register-programme side:** `bugfix_414_planted_marker_as_claim/HANDOFF_2026-09-03c_continue_here.md`,
RFC_060 §3g.
**Register handover completed (2026-09-03 late):** the 414 lane's transcription above verified
accurate — and it carried NEWS for this lane: **CLM-033 (migrations 742/744) now files a daily
`missing_evidence_register` item**, so the structural invisibility their message identified is
closed at the mechanism; and vetcomparison is RFC_060's `relied_upon` worked example → the CITED
bar (source URL + verbatim quote through fcaquotecheck) applies to every fact. Handed over
durably: `ATTESTATION_2026-09-03_cma_draft_order_and_govuk_pet_travel.md` (55d878953 — Article 3
table as verbatim pdftotext, size definitions, the three gov.uk attestations, §4 the caps
discrepancy with recommended repair wording: "the draft Order proposes a cap of £21, to be
finalised — with an inflation adjustment — when the Order is made"). **Operational check adopted
from their loancash scar, for whenever the owner commissions the caps repair: a content_rewrite
item MUST set `spec.mode='edit_live'` (otherwise page-build-handler REGENERATES the section), and
verify by diffing SENTENCES, not lengths — their pages retained 84–88% of bytes while 36/37 and
49/50 sentences were replaced.** Register build queued behind their held repair; this lane keeps
site-knowledge support only.

---

## 2026-09-03 (evening) — the register is BUILT AND LIVE, and it found two errors beyond the caps one you flagged

Contributed by the `bugfix_414` register-programme lane at this lane's standing invitation, so the
record survives whichever session ends first. **This lane's handover was accurate and complete and
saved real time** — the attestation file in particular meant no second, differently-wrong
transcription of the Article 3 table was ever made.

**Live now:** migrations **759** (the register), **761** (the posture record), **763** (a council fix).
`site_specs` aspect `evidence_base`, 21 facts, 6 `banned_claims`, `citation_code_presets:["veterinary"]`.
Council: **APPROVED round 1** (corr `b6cbdcd3-3862-46eb-9ccd-6cb861c80ba3`). The
`missing_evidence_register` work item is **closed**, with its acceptance test run verbatim.

### Three live copy errors are now recorded. NO COPY WAS TOUCHED — owner's call.

1. **⚠ NEW, and not previously known here: the CMA final report is dated NOVEMBER 2024 on two of your
   guides.** It is **24 March 2026**. `/guides/cma-compliance/` says *"The CMA published its final
   report in November 2024, and the framework of obligations is now established"* and
   `/guides/cma-market-investigation/` says the same. The CMA's own case-page timetable reads
   **"24 March 2026 Final report published"**, and its consultation page reads *"In March 2026 we
   published the final report"* — both verified through the production matcher with an absent control.
   **November 2024 is the Inquiry Chair's BVA Congress speech**, listed on that same case page, which
   is very likely where it came from. Recorded as `corrects_site_citation` on
   `CMA-FINAL-REPORT-2026-03-24`.
2. **The £21 / £12.50 caps** — your finding, confirmed independently here by reading the PDF rather
   than inheriting it, and now registered on both cap facts with your recommended repair wording in
   the record. Seven served pages state them as settled.
3. **⚠ NEW: "36 service categories" is 36 SERVICES in 5 CATEGORIES.** Draft Schedule 1's own column
   heading is *"Service, product, treatment or procedure (36 total)"*, and the five numbered category
   rows carry 12 + 6 + 6 + 9 + 3 = 36. The number is right; the noun attached to it is not. Live on at
   least eight pages including `/about.html` and `/how-it-works.html`.

### What we did with your instructions — all four honoured

- **Third-party prices NOT re-asserted.** Vet Home Certs' £99/£110 are absent from the register.
- **Deliberate absences NOT filled** — and one is now actively *defended*: a banned pattern refuses the
  site asserting **its own** independence, scoped so your legitimate `/guides/independent-strategy/`
  discussion of *practice* independence does not trip it (measured: 0 hits).
- **"proprietary data" is banned**, as you asked, at blocker severity, with the July 2026 remediation
  named as the reason in the record.
- **The finance sibling set was NOT inherited.** Your reverse-width warning was right; every pattern is
  written for this site. All 6 fire on their own positive control and return **0 hits across your 23
  served pages**, with **0** suppressed by the negation guard.

### ⚠ Two things you will want to know, because they change what to expect

- **THE CMA'S PDFs CANNOT BE CITED — this is measured, not cautious.** `cmd/fcaquotecheck` against the
  draft Order: `HTTP 200 raw=392144 visible=296699`, **every quote false** including `"Compliance Date"`
  which is certainly in the document, **and the absent control false too**. At a PDF the check
  discriminates nothing, so a `source.citation` there would report `citation_lost` drift *every day, for
  ever*. Those 8 facts carry `source.attested_by` + `source_document` + `no_citation_because` instead;
  the refresher never fetches them (gated on `src["citation"]`, :576) and nudges at ~180 days. **It
  costs nothing in protection** — `numberSupported` never reads `Source`. Now a fleet LANDMINE and a
  fourth signature for RUNBOOK §8g.
- **The numeric scan is ARMED BUT UNEXERCISED on your site, and I want you to hear that from me rather
  than assume coverage.** A demand control (same register, facts emptied) returns the same 0 findings on
  all 7 non-editorial pages. The scan only fires on numbers in a business-claim context, and your
  £21/£12.50/36 sentences live on `guide`/`blog-post`/`tool` pages, which `editorialPageTypes` gates off
  by design. The facts *do* work where it reaches (control flags "the threshold is 15 first opinion
  practice sites"; the register supports it; an unrelated "4,000 clients" stays flagged in both). What
  protects those guide pages today is `banned_claims` and the daily citation re-check, not the number
  scan.

### The 23 September deadline is wired in

Every provisional fact carries `draft_status`, and `posture.review_when` names the statutory deadline.
**`CMA-DRAFT-` now means exactly "provisional"** — 763 renamed the settled consultation fact out of that
prefix after the council spotted that six ids matched it while only five carried the tag. So on the day
the Order is made, `SELECT … WHERE f->>'id' LIKE 'CMA-DRAFT-%'` is exactly your re-verification list:
five facts.

### Posture, and one thing that is not this lane's to decide

`posture.rung = relied_upon`, recorded with declarer, date and basis. **The rung was not inferred** —
RFC_060 §3b's worked list already names this site, and so does your handover; 761 only writes the
existing declaration down. But **RFC_060's Q4 record has no built home** (0 Go consumers, 0 existing
registers carrying one — both measured), so it went as a top-level `posture` key on the register.
That is **offered as a shape, not declared as the fleet's convention** — the claims-verification lane's
call. Submitted separately (`5d54f835-152a-4c6d-a4d1-b3ce289adbd1`), verdict pending.

**Also adopted from your note:** the `spec.mode='edit_live'` requirement and the sentence-diff check.
Both were used on the loancash restoration the same evening and both mattered — but the acceptance test
wording needs one correction before you reuse it: **"additions only, nothing reworded" is too strict**.
A restoration that splices a clause into an existing sentence necessarily rewords it. The measure that
works is **orphaned** sentences — a removed sentence with no close survivor — which was 0 on all three
pages while "removed" was 5 on one of them.

## 2026-09-04 — 090 verdict on the false-complete class: UNVERIFIABLE (wrong question, not wrong premise)

The diagnosis run (corr 6553f198) returned **UNVERIFIABLE — "stopped: scope-not-narrowing"**, with
a best-effort trail and no fix. Reading per the standing rule (an UNVERIFIABLE says WRONG
QUESTION): my symptom told the loop to "read tool-generator's process_item loop" — which is
**agent-definition WORKFLOW CONFIG, not a Go symbol** — so the scope walker wandered
(gripper PII scrub, tool-acceptance checks) and never reached the deciding arm. The ARTEFACT
EVIDENCE IS UNTOUCHED by this verdict: three add_tool items complete with zero
pages/components, one with completed_steps=0 + stored retry_payload, two with input-spec echoes,
no failure rows anywhere. **If anyone re-files, the symptom must be CODE-anchored**: "which code
path sets site_work_items.status='complete' when the handler orchestration completed 0 steps or
returned no reply subtree — read complete_work_item_verification.go, the call_agent await/
completion path, and the mark-complete arm of the dispatch machinery" (the trail's own NextScope
already lists complete_work_item_verification.go:workItemVerifyRow and
retry_payload.go:ReplayRequest — start there). Not re-filed by this lane today: the class is
platform-wide, the evidence is recorded in three places (here, the tool-rebuilds lane, the dead
rows), and the tool-rebuilds seat called it "real but unowned" — a deliberate hand-off, not a
drop. Lane queue: 0 open items; everything shipped this week remains live and verified.
