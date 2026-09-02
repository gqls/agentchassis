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

### Rerender verification

(pending at time of writing — see next entry)
