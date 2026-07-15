# PLAN — Empty sections & fix-loop completion integrity

**Created 2026-07-14. Owner: uk@websy.uk. Testbed: robot-hands.com
(site_id `00ff3af5-dad8-4770-9f70-3edc267a3c92`).**
Origin: `../HANDOFF_2026-07-14_empty_product_sections.md` (read it for the full
evidence trail). Running record: `RUNNING_NOTES_empty_sections_loop_integrity.md`.
Operator tasks: `RUNBOOK_empty_sections_loop_integrity.md`.

## The problem, in two sentences

robot-hands.com ships live, planned product pages whose every value field is
empty — and the platform's own fix loop **detected this, "handled" it, and
marked it complete without fixing anything** (2026-07-10), then two-strike
parked the re-detections as non-dispatchable `unresolved` zombies. The
mechanism is generic: any handler saga that no-ops "successfully" gets its work
item stamped complete by the dispatch loop, so the fleet's success signal
cannot be trusted without completion-time verification.

## Root cause (established 2026-07-14, see RUNNING_NOTES for evidence)

Three stacked holes:

1. **Dispatch loops complete unconditionally.** `build-dispatch-loop` (and
   site-work-orchestrator's `fix_items_loop`) run `call_handler →
   mark_complete`; `complete_work_item` fires on every successful saga.
2. **page-build-handler's no-op exits are success-labelled.**
   `check_has_ready_sections` (ready_count 0) and `check_content_produced`
   (writer skipped) route to `complete_error` — a `complete_workflow` step —
   without flagging the item.
3. **gripper-detail guarantees the no-op.** It is an entity page with
   `pages.sections = []` / site-plan `sections: null`; its components were
   written by the 2026-05-02 pipeline and nothing in the current pipeline owns
   filling them. The handler is page-scoped, the item is section-scoped —
   every re-drive no-ops identically.

## Phases

### Phase 1 — Loop integrity: verification-gated completion  ✅ DONE, PROVEN LIVE
The fleet-wide win. `CompleteWorkItemAction` now consults a per-item-type
verifier registry before stamping `complete`; the `empty_section` verifier
re-runs the exact detection predicate against `spec.component_id`. Defect
persists → item routes into the existing fail/attempt machinery
(`attempt_count+1` → `triaged`/`failed`), claim released, handler result kept
under `result._verification` for forensics. Fail-open on verifier error
(re-detection + two-strike is the backstop).

- [x] Verifier registry (`discovery_checks/verifiers.go`)
- [x] `empty_section` verifier sharing the detection predicate
      (`check_empty_sections.go`), unit tests
- [x] Completion gate (`complete_work_item_verification.go`, wired in
      `load_work_item_actions.go`)
- [x] Deployed in chassis `v1.0.1117` (verified against pod binary)
- [x] Applied `sql_for_agents/149_page_build_handler_noop_flags.sql`
      (handler no-op exits park item at `needs_human_review` — defence in depth)
- [x] **Live re-drive PROVEN 2026-07-14**: re-drove the ORIGINAL falsely-
      completed item (`4e37b25b-bea1-4422-a16b-00018d61a8da`, product-details
      on gripper-detail). Result: `needs_human_review`, attempt_count=1 — SQL
      149's flag fired, the pre-existing completion guard held it there. Never
      reached `complete`. See RUNNING_NOTES Session 3.

### Phase 2 — Detection: required_fields_missing check  ✅ DONE, 🔶 first pass pending
New discovery check: deployed component instances whose schema-required,
LLM-sourced value fields are absent/empty in `content_data`. Flag-only at
`needs_human_review` (no automated handler can fix these honestly), capped 25
per pass. Scope deliberately excludes `query.*`/`site_assets.*`/`pages.*`
sources (render-time; dartsonline's working product grid must not flag) and
image fields (owned by `image_source_unsatisfiable`).

- [x] `check_required_fields_missing.go` + unit tests; in `v1.0.1117`
- [x] Applied `sql_for_agents/150_enable_required_fields_missing_check.sql`
      (confirmed in DB: `run_checks.config.checks` array includes it)
- [x] **First discovery pass PROVEN 2026-07-14**: fired
      `completeness-discovery-agent` directly via kcat for robot-hands — 8
      flags (broader than the ~6 estimate: also caught `tool-guide-intro` and
      3 `article-body` instances beyond the product family). Cross-check on
      dartsonline: **0 flags** — confirms the `query.*` exclusion holds and
      the working 14-card product grid is correctly left alone.

### Phase 3 — Fail-safe content: meta-commentary guard  ✅ DONE
`validate_page_content` check 7: blocks LLM prose-about-the-task shipping as
page copy (live case: product-card-with-cta stored a schema apology as
content). Missed the v1.0.1116 build by ~2 minutes; present in v1.0.1117.

- [x] `checkMetaCommentary` + unit tests (narrow, unambiguous patterns)
- [x] Deployed in v1.0.1117 (verified: `grep -ac "LLM meta-commentary in
      content" /app/agent-chassis` → 1)
- [ ] Not yet exercised by a real case — confirmed by unit test + binary
      presence only, no live meta-commentary has occurred since deploy

### Phase 4 — robot-hands product pages: category decision  🔶 IN PROGRESS
robot-hands is a spec/comparison site; Add-to-Cart furniture is category-wrong
regardless of data. Owner decided 2026-07-14: **Option C** (spec-sheet
component, no cart furniture) — with real data sourced by a new
discovery/scrape workflow, NOT fabricated. See RUNNING_NOTES Session 5 for
the feasibility research behind the split below.

**Correction to the original handoff's premise:** "wire a real gripper
catalog" assumed a comparison dataset already existed (MatchMatrix / gripper-
catalog). It doesn't — those pages are pure marketing copy, `products` and
`affiliate_products` have zero rows platform-wide for ANY site, and
dartsonline's "14 real product cards" (the handoff's proof the pipeline is
sound) are **frozen `rendered_html`, not a live mechanism** —
`query.products`/`query.affiliate_products` is not implemented in
`queryresolve.go` (only `pages_where_type`/`pages_under_section` are). If
dartsonline's product-grid were rebuilt today it would very likely go hollow
the same way gripper-detail did. This means Option C requires building real
infrastructure, not just swapping a component.

Split into two sub-phases:

**4a — Structural: ✅ DONE 2026-07-14, pending chassis rebuild to go live.**
- [x] New content_component `gripper-spec-sheet` (stroke, gripping_force,
      payload, weight, ip_rating, interface, voltage — each field
      conditionally rendered, no field forced). NO cart/Add-to-Cart/Buy-Now
      furniture. `on_missing: skip_section` on the `products` field.
      `sql_for_agents/151_gripper_spec_sheet_component.sql`.
- [x] Live `query.products` resolver added to `queryresolve.go`
      (`resolveProducts`, site-scoped, `category` arg — `query.products:gripper`).
      Same fix incidentally makes dartsonline's product-grid render-safe
      again (same query name, same mechanism, no longer stale-only).
- [x] Provenance (`source_url`, `verified_date`) stored in `products.content_data`
      — no migration; rendered on every spec-sheet card ("Source: … · Verified …").
- [x] Swapped `product-hero`/`product-specs`/`product-details`/
      `product-card-with-cta` on gripper-detail, and `product-hero`/
      `product-specs` on product-detail, for `gripper-spec-sheet`.
      `sql_for_agents/153_gripper_detail_page_swap.sql`. Verified live in DB:
      both pages now list `gripper-spec-sheet` as their first section;
      `site_specs.site_plan` (authoritative for gripper-detail) and
      `pages.sections` (product-detail's only source, since it isn't in the
      current site_plan at all — a pre-existing orphan, not something this
      migration introduced) both updated.
- [ ] **Blocked on chassis rebuild** — `resolveProducts` isn't in a deployed
      image yet. Until then, any rebuild of these pages will find
      `query.products` unresolvable and correctly `skip_section` (safe
      failure, not broken) rather than render real specs.

**4b — Data acquisition: ✅ DONE 2026-07-14 (manual, not a recurring workflow).**
- [x] Real specs for 5 manufacturers (Schunk, OnRobot, Robotiq, Zimmer Group,
      Festo) — fetched directly via WebFetch, not LLM recall. 4/5 from the
      manufacturer's own site; Festo's own site blocked automated fetch 4
      times running, so its row is sourced to a distributor listing
      (RS Online) instead — flagged in the seed file's header for
      transparency. Every field is either a number literally read off the
      fetched page, or absent (never inferred) — manufacturers disclose
      different subsets, so row completeness varies honestly.
      `sql_for_agents/152_gripper_products_seed.sql`. Verified live: 5 rows,
      category='gripper', all with source_url + verified_date.
- [x] Decided NOT to build the originally-scoped scrape/discovery workflow
      (web-search-adapter → firecrawl → LLM-extraction → products) this
      session — did the discovery + extraction directly instead, since I can
      read a manufacturer's own numbers with more confidence than an
      unsupervised extraction step, and it's faster for a 5-row first pass.
      **A reusable platform workflow for future re-verification/expansion is
      NOT built** — this is a real gap if the owner wants specs kept fresh
      or the catalog grown beyond 5 rows. Flagged as follow-up, not started.

**4 — LIVE AND PROVEN on gripper-detail (2026-07-15).** Chassis v1.0.1120
carries `resolveProducts`. gripper-detail rebuilt via the real dispatch path
(re-drove its `empty_section` item → `build-dispatch-loop` →
`page-build-handler` → content-writer). Verified:
- `gripper-spec-sheet` rendered 8,448 bytes; all 5 manufacturers present
  (Schunk/OnRobot/Robotiq/Zimmer Group/Festo), every distinctive real spec
  present (20–235 N, 85 mm, 11 kg, IP67, 1520 N, 218 N, …), all 5 source
  attributions with "Verified 2026-07-14".
- ZERO cart furniture, ZERO meta-commentary/apology text in the render.
- **Live on https://robot-hands.com/entities/gripper-detail.html** (HTTP 200):
  all 5 manufacturers present, ZERO empty `pd-title`/`pd-price` shells (the
  original bug's signature — gone), ZERO cart furniture.
- Bonus: this was a second live exercise of the Phase 1 gate — the work item
  completed with `result._verification = {"status":"verified","detail":
  "component no longer exists"}` (the old product-details component was
  deleted, so the verifier correctly cleared it).

**product-detail — DONE 2026-07-15** after one regression + fix. Its first
rebuild resurrected the deleted product-hero/product-specs because the
authoritative `site_plan_sections` table still held them (migration 153
updated only `pages.sections` + the `site_specs` aspect). Migration 154
corrected the table; rebuild #2 held. Now: `gripper-spec-sheet` deployed
(8,445 bytes, all 5 manufacturers, 5 sources, 0 cart, 0 meta); live at
https://robot-hands.com/product-detail.html HTTP 200 with all 5 manufacturers,
0 empty shells, 0 cart. Final consistency check: all three section sources
agree (`site_plan_sections`, `pages.sections`), gripper-detail unaffected.

**Two deployment gotchas discovered, both now in the RUNBOOK:**
- §5b: do NOT rebuild a page by orchestrating `page-build-handler` directly
  via kcat — the internal spawn→call_content_writer handshake never delivers
  (child idles out at 180s, parent hangs until the 90-min reaper). Use the
  dispatch path (re-drive an `empty_section` item). Cost 2 failed attempts.
- §5c: a page's section list has THREE sources in priority order
  (`site_plan_sections` table → `site_specs` aspect → `pages.sections`), and
  the winner is synced down over `pages.sections`. Editing only the lower
  sources silently regresses on rebuild. Cost the product-detail regression
  (migration 154).

### Phase 5 — Later / spun out
- `sectionHasVisibleContent` should measure **resolved data**, not text
  (`rerender_single_page_action.go:436`) — static labels currently keep hollow
  sections alive at assembly.
- `skip_section` ownership for components outside the spec-sections pipeline
  (entity pages) — who enforces `on_missing` when plan_sections never runs?
- Verifiers for more item types (`unresolved_cta`, `image_url_404`, …) — the
  registry is there; each is ~40 lines + test.
- Export the false-completion case to `fixloop_eg_dartsonline/` as a graded
  benchmark ("bug dissolves but isn't fixed").

## Coordination with the fixloop workstream (read 2026-07-14)

`docs024_key_docs_latest/fixloop_eg_dartsonline/DESIGN_triage_and_escalation.md`
designs the same problem class one layer up: a `diagnosis-triage` router that
turns a handler's loud failure, silent failure, or missing-capability into an
escalation to the diagnose→plan→council→PR loop (Tier 3). Its Phase 2 is
explicitly "a silent-failure verification checker... owned by THIS thread" —
overlapping territory with the gate here. Reconciled as follows (not a
duplication once you see the layers):

- **This gate operates INSIDE completion**, for the specific item being
  completed, re-using the detecting check's own predicate. **Their Phase 2
  operates AFTER completion**, scanning observable state independent of which
  item touched it (their example: section-index pages `active` with zero
  components).
- **Consequence: for any item_type with a registered verifier here, silent
  failure cannot occur** — a blocked completion routes into the existing
  attempt machinery (`attempt_count+1` → `triaged` → `failed` once exhausted).
  That is already triage's Phase-1 candidate (`status='failed' AND
  attempt_count >= max_attempts`). **No Phase-2 checker is needed for
  `empty_section`, or any future item_type registered here, once triage ships
  live.** Their Phase 2 remains necessary for defect classes with no work-item
  predicate to hook (e.g. structural checks with no registered verifier).
- **Recurrence detection (DESIGN's option b) already exists platform-wide** —
  `insertWorkItem`'s two-strike rule (`work_items_common.go`) marks an item
  `[unresolved after N attempts]` on repeat terminal completion within 7 days.
  Not something Phase 2 needs to build; every discovery check already gets it.
- **`result._verification`** (this gate's structured verdict: `status`,
  `detail`) is available on every blocked/verified completion — cheaper for
  their tooling to read than re-deriving a check, if a future Phase-2 checker
  ever wants to ask "did this specific completion work."
- **Ownership stays as DESIGN decided**: fixloop thread owns silent-failure
  verification checkers for classes without a registered verifier here; this
  thread owns the completion gate + registry + the underlying
  `site_work_items` machinery for classes it does register. Before either
  thread adds a new verifier/checker for the SAME item_type, check the other's
  running notes first.
- The robot-hands false-completion case (this workstream's origin) is
  structurally the SAME class as DESIGN's darts benchmark (§"the hard case").
  Worth folding into the fixloop benchmark suite once both threads agree on
  the write-up — flagged, not yet done.

## Success criteria

1. No `empty_section` item can reach `complete` while its component still
   renders empty (gate proves it on a live re-drive).
2. The ~36-item robot-hands empty_section backlog resolves to honest statuses
   (`needs_human_review` with reasons, not zombie `unresolved`).
3. `required_fields_missing` flags all 6 robot-hands product instances and
   zero dartsonline instances.
4. No LLM meta-commentary can pass `validate_page_content`.
5. robot-hands serves no category-wrong product furniture (Phase 4 decision
   executed).
