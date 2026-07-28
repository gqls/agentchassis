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

## 2026-07-24 (later still) — pilot rebuild FAILED on a PRE-EXISTING platform bug; diagnosis filed

- Fired p1_trigger_rebuild.sh SEND=1 (corr 7a820803). Queue latency ~8 min, then
  the run FAILED at build_pages_loop_iter_0_write_page_content / resolve_links:
  `contract violation for agent 'internal-link-resolver': missing required
  fields: [sections]. Provided: [page_name page_type site_id]`.
- NOT our component's doing: identical failure fired TWICE on 2026-07-16 (also
  from a rebuild loop, agent_error_log), and 60 writer-bearing orchestrations
  COMPLETED in the last 8 days via the normal build path. Mechanism (pinned by
  read, filed for verification): page-content-writer's resolve_links maps
  `"sections?": "input_data.section_plan.sections_ready"` (OPTIONAL) while the
  internal-link-resolver input_contract REQUIRES [site_id, sections]; the
  pageflow-builder caller supplies section_plan, page-rebuild's writer dispatch
  does not ⇒ every page-REBUILD dies at resolve_links. The step's
  error_step:select_sections intent (non-fatal link-resolve) is BYPASSED because
  the violation happens at extraction, before the call.
- Greps clean: /bugs_open/ + /bugs_closed/ only hit 029 (different resolve_links
  defect — phantom tool links); needs_diagnosis queue was EMPTY. Filed 090 per
  CLAUDE.md (cause in shared infra, cross-cutting, fix would change fleet
  behaviour): CORR **38cffebf-d01a-4922-9f39-e2deb5930e0d**, item_key
  `needs_diagnosis:page-rebuild-s-per-page-writer-dispatch`. Advisory noted:
  local HEAD 51 ahead of origin — irrelevant here, defect is live-DB config.
- Live about page verified UNTOUCHED (0 grep hits) and build_status still
  needs_rebuild — the armed pilot survives; a fixed rebuild picks it up as-is.
- [INFERRED] the 07-16 pair was another thread's rebuild attempt that was
  abandoned or routed around; no bug was filed then — the sweep only sees
  recorded failures, and these two rows evidently didn't clear its bar.

---

## 2026-07-27 — the block is English-only, and the second site that wants it is Spanish (from traffic_probe)

**Not a bug in your build — a design gap this workstream is the right owner of.**
Raising it here rather than patching a shared fleet component from a site thread.

The owner has asked for the for-sale block on **relojistas.com** (tier 2, Afternic
listing confirmed live from the owner's dashboard). It is the **first non-English
site** to want it, and every string in `about-commercial-block.html_template` is an
English literal:

```
"Built by … We design and build sites like this one — see how it's done"
"The {{.domain}} name is available to acquire — register your interest."
"{{.domain}} is part of our portfolio and may be available to acquire — make an enquiry."
"Advertise on {{.domain}}. A small number of sponsored placements are available…"
```

relojistas' own `mission_brief` says *"todo el sitio debe estar íntegramente en
español"*. Rendering this block there would put English on a wholly Spanish page —
which is the exact defect class that site spent this morning documenting
(`bugs_open/071` sighting, commit `b96acad7d`): `component_library.go`'s English
defaults leaking onto it as a dead `/contact.html` CTA, `Browse all guides`,
`Explore All Archetypes`, and a footer `<h4>Contact</h4>`. Adding a fourth instance
deliberately would be worse than the three we found by accident.

**The awkward part for the design, and why it isn't just a translation:** the
tier-ladder wording was owner-approved as *register* — "acquire" not "for sale",
representation not adjectives, no price on the page. That register does not survive
machine translation intact; "available to acquire" has no single neutral Spanish
equivalent that avoids sounding either like a classified ad (*se vende*) or like
legalese (*susceptible de adquisición*). So this needs the owner's ear for Spanish
register, not a `es:` map filled in by whoever ships first.

**What I checked before raising it** (so you don't re-walk it):

| question | answer |
|---|---|
| Does the platform model per-site language? | **No.** No `language`/`locale` column on `sites`; `grep` finds it only inside `deploy_config.rss_feed.language` (`"es"` for relojistas) and in RAG config fields. There is no seam to hang this on yet. |
| Would writing relojistas' `commercial` aspect render anything today? | **No.** Zero Go references to `about-commercial-block` anywhere in `platform/`, `internal/`, `pkg/` — Phase 2's default-set hook is not built, so the block only appears where a section is explicitly inserted. |

I have therefore written relojistas' `commercial` aspect with the true facts
(`class=portfolio, tier=2, for_sale_requested=true, marketplace_url=…`) and
**deliberately not inserted the section**. It is inert and correct, and the day this
block speaks Spanish it is one section-insert away. `notes` on that row says so.

> **RETRACTED 2026-07-28 — please disregard; this was my error, not a finding.**
> I raised here that relojistas' Afternic listing had **Minimum Offer = 0** and that the
> locked anti-lowball protection was therefore absent. **Both false.** The floor is
> **$12,000** (owner). I misread a column-aligned dashboard paste. **No decision is needed
> and nothing about the design is in question** — sorry for the noise in your file.
>
> The one question that survives is a real one, unrelated to the number: **if
> `for_sale_requested` is the flag that means "represented", is a floor part of what that
> flag asserts?** Worth deciding on its own merits, not because of relojistas.

Contact for this: the relojistas thread, `traffic_probe/` (docs + NOTES 2026-07-27).
