# HANDOFF — portfolio_positioning — 2026-09-02 (night). **START HERE.**

Supersedes `HANDOFF_2026-08-31_continue_here.md` + its three addenda. Owner read-out:
`SUMMARY_2026-09-02_first_four_remakes_live.md` (new tonight). Counts carry their dates.

# 0. STATE IN ONE PARAGRAPH

**The first FOUR remakes are LIVE**: advertise.co.uk, websitepromotion.co.uk, seotools.co.uk,
designblog.co.uk — briefed, released, built, deployed and serving through the Cloudflare worker,
all on 2026-09-02, serve-verified at the bodies. The release recipe with every gotcha is
`RUNBOOK_remake_release.md` (new). The owner critiqued the results same night → **the next 18
briefs are HELD** on two inputs: `bugs_open/444` (listing pages ship EMPTY with brief-echo prose
— three producer-absence mechanisms; enablement per site or hold those page types) and the
theme-kits differentiation levers (layout in the brief NOW, colour by referent NOW, chrome
behind №5's pin experiment — runbook §5/§6). Sitemap machinery (642) is untouched and
self-maintaining; the four new sites join its rotation. **A fresh chassis build was deployed
late 2026-09-02 (owner message)** — nothing of THIS lane's was inert-until-roll (all our work
is DB config + docs + work items); the ~300s no-dispatch window after a chassis pod (re)start
applies to anything you fire immediately; lanes whose fixes ride the roll (424 imagery class,
444's committed-but-gated code) verify per-SERVICE at the provenance stamp, never per-fleet.

# 1. FIRST TASKS

1a. **Convergence checks on the four live sites** (~10 min, judge at served bodies):
    advertise `/sitemap.xml` should exist after 642's next selection of it — 404 tonight was
    expected; seotools' 7 tools arrive when design-discovery's rotation reaches it (its 7
    `owned_page_review` holds are TRUE until then — do NOT close them on the advertise
    precedent, their condition differs, NOTES (o)); then batch-verify the trio's ~50
    `unresolved_cta` items at served pages and close-with-evidence only what resolved.
1b. **`642`/`622` routine checks** — unchanged, queries in HANDOFF_2026-08-26 §1a/§1b.
1c. **The owner's open items**: advertise negation flag (recommendation standing: keep) ·
    websitepromotion claims_unverified (4 claims, 1 high) · fleet-wide claims backlog (8 sites,
    never drained) · two test briefs (indoorplanters.co.uk, buytoletcalculator.uk).
1d. **Next briefs UNBLOCKED (updated same night)**: the 444 thread answered the enablement
    shape — pre-enablement recipes are in runbook §6 (feeds via content_features.news_feed or
    source seeds; directories via kind checklist or exporter config row; glossary/showcase
    held back), fire direction unchanged, their validator lands via council as belt-and-braces.
    A brief may fire once its verticals' pre-enablement is DONE and the direction carries
    template v2. Single-pagers first; 3 protected; twins ⚑OWNER; insurance last.

# 2. LIVE vs NOT-YET-PROVEN

| | |
|---|---|
| Release recipe | proven ×4, runbook'd |
| Serve path (zone→NS→A→route) | proven ×4 — and its failure shape (NS-before-zone = dark domain) happened and is runbook'd |
| Brief-writer → owner gate → build | proven end-to-end ×4 in one day |
| Listing pages (directory/glossary/feed/showcase) | **BROKEN as shipped — bugs_open/444**, every naive check passes |
| Chrome pinning | **UNPROVEN** — experiment staged for №5 (runbook §5, three-way read) |
| Layout/colour differentiation | proven levers, in template v2, unexercised by any remake yet |

# 3. OPEN (beyond §1)

- **18 remakes remain** (`DECISION_2026-08-20...`§2). **№5 carries TWO canary duties**:
  the chrome-pin experiment (runbook §5, three-way read) AND the planner imagery-prompt canary
  (runbook §5b — designblog lane's owner-directed prompt change, live 19:59:56Z; check the
  plan's imagery.sections; article pages exempt). ⚠ TWO lanes cite "migration 718" — resolve
  by FILENAME.
- **444 routing**: OWNED — session "bugs_open/444". Their diagnosis CORRECTED the bug file
  (two-guards-in-series mechanism for directories; news legal-by-design; writer-prose is
  glossary-only) and answered the enablement shape → runbook §6 verbatim. Council corr
  `c0990eb3`: 2 REVISE rounds (real improvements, landed by sha), ~~**round 3 pending;
  migration 720 committed but HELD unapplied until the verdict**~~ **→ APPROVED 20:53:22Z,
  720 APPLIED per the fixing session (addendum below, 21:20Z)** — post-roll runbook §2 traps:
  deferred listing section = source config; missing gate receipt ≤~3h of a terminal row =
  anti-churn (326), not failure; a replan never drops a BUILT page.
- **gamesdesign/gamedesign pair**: rename DONE + serve-verified (their sessions); Pro-tier
  name+home UNDECIDED (GD1/GD2); p2p cross-link pair dependency recorded; GD2 brief is the
  reckoning point. `bugs_open/439` = adoption-carries-name class. **NEW same night:
  `bugs_open/447` — tool-suggester is SEAT-BLIND (reads identity+classification only) and
  proposed SIX of the sibling's tools by name onto gamedesign.uk; held reversibly by their
  lane. GD2 states hosts_tools=FALSE machine-readably (the consumer is 447's opt-in build);
  the class is stamped on GDN1b and runbook §2 — eye every add_tool wave on any P5 pair until
  447 lands. Their `bugs_open/446`: owner re-ruled gamedesign's execution (louder, gamier,
  re-dispatched 20:11Z) — GD2 seat UNCHANGED.**
- **indoorplanters.uk stub GDN1b** — P5 pair decision owed before either twin builds.
- Carried: 21+ domains no register row (NOW A SERVING BLIND SPOT — orphan-served domains are
  invisible to every pages-row detector, NOTES (d)) · two-copies register question ·
  Christmas sender · adversecreditmortgage halted · Cloudflare robots (owner).

# 4. TRAPS (today's, all with worked cases in NOTES (a)–(q))

`ensure_site_record` scans name+network_id bare (row must be complete at CREATION) ·
handlerless notifications can't be re-queued (`swi_no_handlerless_promotable`) · site_specs
supersede = separate statements, never chained CTEs · validation blocker detail lives in
agent_error_log's SECOND row · a claimed "race" needs timestamps (my owned_page_review wording
was corrected by them) · a spec census is not a pages census (Pro-name lesson) · the 404-token's
empty zone list ≠ zone absence · a branded 404 title doesn't tell you which stack answered
(read the Generator meta) · count ITEMS on listing pages, never bytes (444/016b §9) ·
tool-suggester is seat-blind on P5 pairs (447) · a transient `.git/index.lock` failure on this
shared tree = another session mid-commit — retry, never delete the lock.

# 5. FILES OF RECORD

Cold start: this file → `SUMMARY_2026-09-02_first_four_remakes_live.md` →
`RUNBOOK_remake_release.md` → NOTES 2026-09-02 (a)–(t) → README_where_we_are (owner log).
Differentiation: `CONTRIB_2026-09-02_themes_lane_differentiation_levers_measured.md`.
Class bug: `bugs_open/444...md` + 016b §9 tail + §10 row. Register: `positioning_register`
DB (GD1/GD2/GDN1b written today; MD copy deliberately untouched).
Peer sessions live tonight: designblog.co.uk · gamesdesign.co.uk · gamedesign.uk · theme kits ·
feed lane (WebProNews CONTRIB with them) · bugs_open/444 (the gate; ~~will ping its round-3
verdict~~ **landed APPROVED 20:53Z — runbook §2 carries it, addendum below**).

# ADDENDUM 2026-09-02 ~21:20Z — read this before §1 (cluster token EXPIRED at 21:08:03Z)

- **The kubeconfig token expired 21:08:03Z** — 3-day JWT, issued 08-30 21:08:03Z (memory
  `kubeconfig-token-expires-every-3-days`). Every kubectl/psql answers `Unauthorized` until the
  owner refreshes `~/.kube/config_production_uk001`; only the owner can. Everything marked [DB]
  below was read in the minutes BEFORE expiry; every cluster-side step in §1 is BLOCKED on the
  refresh. Served bodies stay readable (curl).
- **444 round 3 APPROVED** 20:53:22Z (corr `c0990eb3`; [DB] `complete_approved` COMPLETED, read
  by this lane). The fixing session's close-out, committed `2d7a98446` (and their direct message
  to this session, 21:1xZ):
  `Council-Reviewed: c0990eb3…`, **migration 720 APPLIED + verified live**, ~~Go gate `6525b45ae`
  INERT until a chassis roll carries it~~ **→ GATE PROVEN LIVE after the token refresh (their
  `560a24c07`, build ∈ [`6525b45ae`, `c610898d1`)); the r2 refinements incl. receipt anti-churn
  ride the NEXT roll. Class half of 444 meets fixed-AND-live; bug stays open for instance work
  + candidate (3) + first-fire confirmation.** Runbook §2 updated (and its deploy-proof line
  now points at their NUL-split probe — CLAUDE.md's two recipes are landmined). **§1d consequence**: the PROMPT half
  is live now — the planner is told not to plan a listing page without a live item source —
  so runbook §6's pre-enablement is the belt and the Go gate the braces; nothing further to
  wait on from 444 before firing. ⚠ [DB] designblog filed `needs_section_data` at 21:04:44Z
  reading *"required query source errored: queryresolve.Resolve: unknown query name
  featured_post"* — that is the SHAPE of 444's "errored REQUIRED field defers loudly", ~~which
  would mean the late roll DID carry `6525b45ae` [INFERRED — stamp unread]~~ **CORRECTED by the
  444 session (21:3xZ): CONFIRMED as their repair firing — "required query source errored:" is
  a novel Reason string born in the defer branch that rode 443's `dbb218a41` (committed
  20:08:04Z), and `featured_post` is one of the FIVE unregistered query bases their round-2
  census predicted. So the row proves the roll carried `dbb218a41` (the DEFER half) and says
  NOTHING about `6525b45ae` (the GATE, a separate commit 16 min later) — the per-service stamp
  is the only discriminator, and their three-part liveness check is their first task after the
  refresh (their NOTES `d6429a362`).** Recorded here so the row is not read as a new bug.
- **§1a, measured at the bodies ~21:15Z:**
  - **sitemaps**: websitepromotion 200/1,401 B · designblog 200/1,854 B · **advertise 404 ·
    seotools 404 — explained, not a fault**: [DB] both were selected by 642 while lame-delegated
    (advertise 16:37:46Z `url_count` 0 / `probe_dropped` 22; seotools 17:38:46Z 0/14). Runbook
    §3 carries the trap. Rotation ticking ([DB] last 20:43Z; due set age/change-quiet/change-busy
    = 0/14/5 at 21:0xZ). If either still 404s on 09-03 evening, ask whether any page changed
    after the stamp BEFORE suspecting 642.
  - **seotools' 7 `owned_page_review` holds are TRUE at the served body — do NOT close**: all
    seven `/tools/<slug>/` serve 200 at ~56 KB with the tool's H1 and 0 `<form>`, 0 `<input>`,
    0 `<select>`, 1 `<button>` (mobile-menu toggle). Control, same probe: advertise's 3 real
    tools = 1 form, up to 11 inputs, 2 selects. Prose shells at the tool URLs; WHO built them
    (generic `needs_page`, or tool-deployer minus its component) needs [DB] `page_components`
    for those pages. Two [DB] facts to read first after the refresh: `save_refused_incomplete`
    on tool-serp-snippet-previewer at 20:40Z ("a section would lose its layout components" —
    something DID try to write that page post-deploy), and `site-discovery-rotation-design`
    last ticked **18:47:45Z** while the other three rotations ticked 20:15–21:00Z [UNVERIFIED
    whether that is its cadence or a stall — seotools' tools arrive on THAT rotation].
    seotools has a `/tools.html` hub in nav (so the 444 "no tools hub" class is not universal).
  - **advertise** [DB]: `site_unreachable` DETECTED ×1 — dark-window artefact (serves 200 /
    75,562 B now); close with the curl once the DB is back. `needs_page` FAILED ×2 — the
    credit-outage pair from NOTES (g), news-index + uk-ad-spend-reference [INFERRED from (g);
    result unread]; both URLs are in nav and `/news/index.html` serves 200 / 61 KB — read the
    items' `result` before touching them.
  - **CTA piles** [DB 21:0xZ]: seotools 18 · websitepromotion 17 · designblog 1 = 36 open (from
    ~54 at (k); 18 already overtaken). Batch-verify needs each spec's section + field → after
    the refresh. websitepromotion `needs_page` d0a5c53f sits at `needs_human_review`
    ("Re-render promote-website-free-uk after its image asset landed") — an odd shape; read it
    before acting.
- §1b [DB 21:0xZ]: 642 rotation COMPLETED on every selection in the last 12 h, `dropped` 0
  except the two dark-window runs above and leopardessconsulting (1); 622 guard holds (min
  deployed_pages 1, apis.uk). §1c unchanged.

# ADDENDUM 2 — 2026-09-02 ~22:2xZ: token back, §1a worked, one class bug filed (`bugs_open/450`)

The owner refreshed the kubeconfig ~21:2xZ (the 444 session noticed first). Everything below is
[DB] or body-verified after that.

- **§1a CTA batch — DONE for what resolved:** 36 open → **11 closed with evidence** (content_data
  field filled AND `href` present on the cache-busted served body; result carries page/fields/
  href counts; two note that both CTAs resolve to the SAME page — websitepromotion `advertise`
  and `channels-index`, cosmetic). **25 remain** (seotools 15 · websitepromotion 10): every one
  still has an EMPTY `secondary_cta_url` (or both) because `resolve_internal_links` found "no
  eligible content hub" — the gated template renders no button, so nothing is broken on the
  page; they re-resolve only if a hub becomes eligible. Leave them; re-run the join query in
  NOTES (u) after any hub/rerender wave, close what fills.
- **seotools' 7 tool holds are TRUE — and now explained: `bugs_open/450`.** The plan named
  every tool page as `hero-tool,generic-text-block` (no tool existed at 16:13Z); the
  phantom-link repair built them as prose 19:57–20:41Z (26 writes / 6 pages, `unbuilt_internal_link`
  → `page-build-handler`); the owned-page guard keys on `rebuild_policy='owned'`, which they are
  not. Control: advertise's plan (13:09Z) named its real tool components because the design
  rotation had reached it at 12:43Z. Design-discovery has NEVER selected seotools or
  websitepromotion (rotation ≈ 1 site / 3 h; last tick 18:47Z — cadence, not a stall). 090 run
  `96e97dc4` fired at filing, `diagnosing` at 21:42Z — **read its verdict first**
  (`SELECT body FROM doc_notes WHERE body LIKE '%96e97dc4%' ORDER BY created_at DESC LIMIT 1;`)
  and fold it into 450. ⚠ 444's CONTRIB in 450 (`ad1b3b1fa`) re-ordered the candidates: the
  plan-side hold (1) cannot close the phantom-link door and may starve the tool rotation — (2)
  make the hold a control / (3) never route a tool target to the generic builder are the
  door-closers; (1) only after §7 and behind a sibling key. Mitigation recipe for №5+: runbook §2b (one-shot design discovery at
  plan time) — do NOT point it at seotools until 450 §7 (tool-deployer on a pre-existing page
  row) is answered. `save_refused_incomplete` 540f5359 closed as a by-product (it was one of
  those shell writes tripping the component floor). `dead_internal_link_live` ×7 will self-clear.
- **advertise:** `site_unreachable` 59309883 (17:46Z, DNS "server misbehaving" in the dark
  window) — LEFT for the probe: the same producer auto-resolves on the next availability pass
  ("probe recovered", gamedesign.uk 21:22Z precedent); advertise's next pass is hours away.
  `needs_page` FAILED ×2 = **`uk-advertising-regulation-map` never built**: component
  `mechanism-flow`, `steps[].branches` declared array got string, 4 failures 17:21–19:07Z —
  **an instance of `bugs_open/437`** (loanzy lane, 119 failures / 6 sites in 14 d, filed today,
  no owner yet). Page is `active`, `deployed_at NULL`; LNK-038 suppresses links to it at render
  so nav stays clean. Nothing to do here until 437's candidate 1 lands; the page then builds
  on retry (the items are terminal `failed` — a fresh `needs_page` may need minting, read 437
  first).
- **websitepromotion:** `needs_page` d0a5c53f re-queued `triaged` — its 18:25Z block was
  `placeholder_text` "to be added" matched inside natural prose ("asking to be added to the
  relevant page"), a false positive; fresh writer run per NOTES (f) — **DONE**: claimed within
  2 min, deployed 21:41:02Z (`869b1eca`), page `deployed_at` 21:45:39Z, 4 components refreshed,
  no placeholder block; served body verified 200 / 70,325 B, `hero-free-promotion.jpg` present
  as the hero background, 0 × "to be added". One `CTA_LABEL_MISMATCH` warning (cosmetic).
  Its `tool-channel-prioritiser` is the one undeployed page: 450's class, no shell yet (no link
  repair reached it) — the hold cd2fda11 stays.
- **designblog** rows (`needs_section_data` featured_post, `capability_gap` deferred,
  `claims_unverified`) are the designblog session's.
- **Queue left for the owner (§1c) unchanged** + the two test briefs.
- **Next session's §1**, in order: 090 verdict on 450 → fold in · advertise sitemap re-check
  (still 404 → "did any page change after 16:37Z?") · seotools/websitepromotion: has the design
  rotation reached them yet (`site_discovery_rotation` rows) — when it does, watch what
  tool-deployer does to the shell rows and record it in 450 §7 · then §1d briefs.

