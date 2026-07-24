# NOTES — about-page commercial (append-only, newest at the bottom)

## 2026-07-24 — design sessions (owner + Claude, "about page" thread)

- Owner brief: about pages across the portfolio should carry (a) a site-for-sale
  link (e.g. Afternic sales form), (b) a price *idea* without committing ("5
  figures") to deter lowballers, (c) an advertise-this-site link to advertise.co.uk
  (cheap, honest about low traffic), (d) as automatic as possible.
- Claude counter-proposals accepted by owner: sell the DOMAIN not the site;
  NO public price band (anchors low) — per-site tier + Afternic minimum-offer
  floor does the gatekeeping; advertising as flat-rate sponsored placement
  (nofollow, labelled) because CPM/CPC is dishonest at ~zero traffic; a third
  signal (built-by → storefront) completes the set and reframes "no traffic" as
  "freshly built demonstration".
- Owner added: a passive listing shouldn't wave "for sale" at an active
  advertiser → resolved via facts + render gates (passive Afternic listing always
  live; loud callout suppressed while advertising_active).
- Owner: flip = front-end API hook, admin in control now, advertise.co.uk later.
- Wording round: owner rejected "premium"/"serious offers" (domainer wallpaper).
  Register locked around "available to acquire" + representation ("domain team" —
  NOT "agent", overloaded in an AI company). Full ladder + advertise + built-by
  copy locked; see PLAN.
- Scout 1 (about-page/config machinery): NO existing for-sale/advertise machinery
  anywhere (greenfield); seams = site_specs aspect + content_components section +
  footer ContentData hook; the resolved_data-merges-last invariant makes
  site_specs-sourced fields LLM-proof. Full result in PLAN §Architecture.
- Scout 2 (admin API seam): admin dashboard → nginx → core-manager /api/v1/admin/*;
  PATCH /sites/:id/specs/:aspect already exists (HandleUpdateSiteSpec,
  site_admin_handlers.go:190); WriteSiteSpecAction is coordinator-ONLY, not HTTP;
  external-caller precedent = X-Bootstrap-Key (agentbootstrap.go:42-60).
- Owner picked **Incremental — DB first** (Phase 1 pilot without image roll).
- Scout 3 DISPATCHED: component INSERT recipe + whether a DB-only single-page
  attach+render is feasible (the render traps: page_rerender does NOT re-select
  components; bugs_open/024 tool fixes never rendered; content_data edits don't
  hold on derived fields). Phase 1 plan is conditional on its answer.
- Also this session: memory index lifecycle split (MEMORY.md active+durable /
  MEMORY_closed.md archive) — unrelated to this workstream, recorded in
  memory/memory-index-how-it-works.md.

Missteps so far: none material yet (design phase). Watch: do NOT assume the
single-page render path re-resolves site_specs sources — that is exactly the
class of assumption 024 punished.
