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

## 2026-07-24 (later) — Phase 1 pilot armed on finetuning.uk

- Scout 3 verdict: DB-only pilot IS feasible, but ONLY via the REBUILD path
  (page-rebuild agent re-runs plan_sections → SelectComponentByType → resolves
  site_specs.*); page_rerender can never ADD a section (016b §RerenderSinglePage,
  rerender_page_sections_action.go:153). Full recipe + traps → RUNBOOK.
- DESIGN CORRECTION vs scout sketch: scout proposed precomputed `show_*` booleans
  in the aspect; rejected — that is write-time derivation (advertise.co.uk would
  have to know to flip show_for_sale when setting advertising_active). We store
  raw FACTS and gate IN-TEMPLATE with and/or/not/eq. missingkey=zero ⇒ absent
  facts falsy ⇒ fail-closed.
- Misstep (mine): queried sites.status='active' / page_type='about' by guess —
  both wrong (status='deployed'; about pages are page_type='content' matched by
  name/url). Then guessed wi.payload & sps.page_id columns — also wrong (spec;
  plan_id/page_name). Schema-first exists for a reason; cost ~3 round trips.
- Pilot candidate sweep (live DB): every deployed site carries open work items
  (fleet norm — "zero open items" is unachievable); differentiator = items fresh
  TODAY touching the target. ai-agent-orchestration.com churning today (19 fresh
  page_rerender + a claimed content_rewrite) — rejected. gaswholesalers quiet but
  is the claims-verification thread's cold-audit pilot — rejected on ownership.
  **finetuning.uk chosen**: quiet, one fresh section_edit but on a DIFFERENT page
  (c67ed17b ≠ about c0c68034), no workstream claims it.
- finetuning.uk + gaswholesalers both have ZERO site_plans rows → authoritative
  section store = site_specs.site_plan ASPECT (the ~5-older-sites path); edit =
  supersede-then-insert of the aspect + pages.sections mirror.
- Pilot ships BUILT-BY ONLY (for_sale_requested=false, inventory_open=false, no
  marketplace_url stored): no Afternic listing confirmed for finetuning.uk,
  advertise.co.uk not built. Honesty rails hold AND the two suppressed lines
  prove the gates. tier="2" pre-set (inert until for-sale flips).
- APPLIED (all verified in one tx): seed 202 (component; sole selector candidate
  confirmed by needle-gate), commercial aspect, site_plan aspect edit + cache
  mirror (both stores agree), about page → needs_rebuild.
- Dispatch: build-pipeline-trigger keys on TRIAGED WORK ITEMS not needs_rebuild
  pages — flagging alone fires nothing. Trigger = orchestrate envelope
  (action=orchestrate, config.agent_type=page-rebuild, input_data {domain,
  site_id}) per travelling_docs/086 pattern → p1_trigger_rebuild.sh (dry-run
  default). Chassis pod started 12:44Z today — well past the 300s rule.
- rel="nofollow" on ALL THREE lines' links including built-by: fleet-wide
  same-owner footer/about links to fundamentallyai.com are exactly the
  widget/footer-credit pattern Google discounts; traffic value is clicks, not
  PageRank. [DECIDED without owner — flag if he wants built-by followed]
