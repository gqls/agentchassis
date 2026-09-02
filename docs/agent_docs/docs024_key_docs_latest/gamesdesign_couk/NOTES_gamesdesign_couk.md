# NOTES — gamesdesign.co.uk (append-only, newest at the bottom)

## 2026-09-02 — lane opens; brand rename executed

Session "gamesdesign.co.uk" [783baf] opened by the owner and announced to Portfolio
positioning per the owner's instruction. Positioning immediately relayed a waiting
owner ruling (they had been holding it because no thread owned this site): **stop
using the name "GameDesign.uk"** — it belongs to the gamedesign.uk rebuild.
gamedesign.uk session confirmed the ruling independently, verbatim: "Please tell
gamesdesign.co.uk lane to stop using our name."

### Re-verification of the handed-over measurements (all re-measured live, 2026-09-02)

- Peer claim "4 current specs carry 'GameDesign.uk'" — CONFIRMED: briefing
  (7814e6b0, 2 occ), design_intent (bf4ef822, 1), identity (b4b25fa1, 2), tools
  (528c662c, 1). The earlier count of six included two since superseded.
- Peer claim "23 of 49 active page titles" — CONFIRMED: 23 titles (21 as explicit
  `| `/`- ` suffix), plus 1 meta_description. `sites.company_name/logo_text/tagline`
  EMPTY — spec confirmed as only source.
- Beyond the handover: 22 `site_plan_pages.title` (the store the build reads —
  fixing only specs+pages would have been resurrected by the next plan-driven
  build), 30 `page_components.content_data` rows / 10 `rendered_html` rows across
  21 pages, `site_components`/plan sections/imagery/directives all ZERO. Served
  homepage carried 30 occurrences (title, og:title, JSON-LD name, tool-card titles
  + alt text).
- One lowercase `gamedesign.uk` occurrence in content: guide-p2p-architecture's
  "Launch P2P Simulator" link. Deliberately excluded from the replace (see below
  for its own fate).

### Decision + execution

Owner chose **"GamesDesign.co.uk"** (positioning's domain-as-brand recommendation)
via explicit question in-session. One transaction
(`SQL_2026-09-02_brand_rename_APPLIED.sql`, this dir): site-scoped backups
(`bak_gdcouk_rename_20260902_pages` 23 rows / `_plan_pages` 22 / `_page_components`
30) → spec supersede retire-then-insert by id, separate statements
(`source='operator'`, `created_by='claude-session-gamesdesign-couk-20260902'`) →
site_plan_pages.title 22 → pages title 23 + meta 1 → page_components content_data +
rendered_html 30 → in-transaction `DO` guard: zero residuals in every store, demand
control (exactly 4 current specs carrying the NEW string), `adopted_from` still
lowercase 'gamedesign.uk'. All counts landed exactly as measured; COMMIT.

Rerenders: 32 active pages (union of titled + component-affected), dispatched
per-page via kafka-publish-lib with receipts (not site-wide `rerender-pages`: queues
hours behind the estate, and `bugs_open/315` records a page skipped by site-level
rerenders). Chassis pods 2.5h old (>300s rule OK); open work items on the site are
the general improvement backlog, nothing touching branding.

### The P2P cross-link — target died the same afternoon

Flagged the inherited deep link
`https://gamedesign.uk/games/p2p-networking/index.html` to both peers as a pair
dependency. gamedesign.uk replied: **already dead since 17:05Z today** — they
retracted the whole old tree (sites-repo `40bd35f19`, 58 files, only 404.html kept)
on the owner's "clear the old files" ruling, before my note arrived. Verified
first-hand: old URL 404, my own copy `/games/p2p-networking/index.html` 200. No
redirect will come (their rebuild is the editorial seat, hosts no tools/games — the
tool kind lives HERE per the P5 split). Fix executed: repointed the
generic-text-block link to the local relative path
(`bak_gdcouk_p2plink_20260902` first), both content_data and rendered_html.

**Misstep, recorded:** first attempt's `DO` guard asserted "exactly 1 component
carries the local link" and ABORTED with 2 — the page's HERO already linked locally
to the same path, which I had not checked before writing the assertion. The abort
was the guard doing its job (nothing committed, backup rolled back with it); re-ran
with the pre-measured expectation (2) and it passed. Lesson: a demand-control
threshold is itself a claim — measure the pre-state it implies, don't infer it.

Positioning updated the register: GD1 name now DECIDED; the cross-link recorded on
both GD1/GD2. Class bug filed once by gamedesign.uk as `bugs_open/439` (adoption
carries source brand verbatim); the seam fix is theirs/whoever takes 439 — NOT this
lane. Per 439 §6, no fleet-zero claim from this site's census: our rows are manually
renamed, so seam verification needs a fresh cross-domain adoption.

### Rerender verification (same day, served artefacts, cache-busted)

Dispatch: 32/32 published with receipts (kafka-publish-lib, 0 failures), 32/32
orchestrations COMPLETED. Served verification per page (old string 0 case-sensitive
+ new string ≥1 demand control + body-size control):

- **24/32 clean PASS** (23 first pass; guide-p2p-architecture after one
  re-dispatch — its first rerender raced the link-fix commit by seconds and served
  the dead absolute link; re-dispatched, now serves local-link×2, old 0, new 3).
- **6 pages serve NEITHER string and that is CORRECT**: bayesian-ranking-guide +
  the five tool-*-guide flat pages carry their hits only in `content_data` fields
  the template does not render (`rendered_html` has zero new-string in the DB too);
  their titles never had the brand. My demand control was over-strict for them —
  the data is renamed, nothing served needed to change. All six 200 with old=0.
- **premium.html: pre-existing 404, NOT rename damage.** `deployed_at` and
  `last_built_at` NULL, file absent from the sites repo — the page row is active
  but was never built. Its rerender honestly reported `complete_skipped` (no
  deploy_result, no error): the never-deployed predicate refusing to "re"-render an
  unpublished page. Class = `bugs_open/349` (a never-built page row still wanted
  live); the page's fate is already queued several times over
  (needs_content_page/needs_content_planning/content_rewrite, most deferred or
  needs_human_review — including a verdict that a premium gate may be structurally
  irreconcilable with the site's no-data-storage promise). NOT released by this
  lane: publishing a Pro page is a positioning/owner decision (GD2 records the
  paid tier's future home as gamedesign.uk).

**Verify-script misstep, recorded:** the body-size control (≥2KB) was meant to
catch a broken page masquerading as "no-new-string" — but the 404 page is ~larger
than 2KB, so premium's 404 read as a content anomaly rather than a missing page.
Check the HTTP STATUS, not the body size; the status check is what separated the
six benign no-string pages (200) from the one real absence (404).

**Positioning residue to flag (flagged to the positioning session 2026-09-02):**
the mechanical replace also turned `pages.title` "GameDesign.uk Pro — For Studios
& Design Programmes" (premium) and the contact-index meta's "licensing
GameDesign.uk Pro" into "GamesDesign.co.uk Pro…" forms. Correct per the ruling
(the old brand had to go) but "GamesDesign.co.uk Pro" as a PRODUCT name is nobody's
decision yet — the peers verified "GameDesign.uk Pro" appears in no current spec,
and GD2 assigns the paid tier to the sibling. Premium 404s so nothing serves it;
the contact page does serve its meta. Owner/positioning call, not this lane's.

**End state 2026-09-02:** zero case-sensitive 'GameDesign.uk' anywhere in
site_specs (current), pages, site_plan_pages, page_components, or any of 31
serving pages; the only lowercase 'gamedesign.uk' remaining is
`identity.adopted_from` (a fact). Site serves as "GamesDesign.co.uk".
