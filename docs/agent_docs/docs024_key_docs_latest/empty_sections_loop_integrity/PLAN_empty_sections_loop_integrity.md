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
- [ ] First discovery pass on robot-hands not yet observed: expect ~6
      product-component flags, 0 on dartsonline (see RUNBOOK §5)

### Phase 3 — Fail-safe content: meta-commentary guard  ✅ DONE
`validate_page_content` check 7: blocks LLM prose-about-the-task shipping as
page copy (live case: product-card-with-cta stored a schema apology as
content). Missed the v1.0.1116 build by ~2 minutes; present in v1.0.1117.

- [x] `checkMetaCommentary` + unit tests (narrow, unambiguous patterns)
- [x] Deployed in v1.0.1117 (verified: `grep -ac "LLM meta-commentary in
      content" /app/agent-chassis` → 1)
- [ ] Not yet exercised by a real case — confirmed by unit test + binary
      presence only, no live meta-commentary has occurred since deploy

### Phase 4 — robot-hands product pages: category decision  ⬜ owner decision
robot-hands is a spec/comparison site; Add-to-Cart furniture is category-wrong
regardless of data. Options:

| | Option | Effort | Note |
|---|---|---|---|
| A | Wire a real gripper catalog (`query.*` source) and keep product pages | High | Only if a catalog data source is actually wanted |
| B | Remove product components + delete/unpublish the two pages (`gripper-detail`, `product-detail`) | Low | **Recommended.** Pages are unlinked (`in_header/footer=false`) but live + indexable |
| C | Replace with a spec-sheet component (no cart furniture), data from the comparison dataset | Medium | Fits the site's actual category |

Destructive → owner call. Prepared SQL can be drafted once decided.

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
