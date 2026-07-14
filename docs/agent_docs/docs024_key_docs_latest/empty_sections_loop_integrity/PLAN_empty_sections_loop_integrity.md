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

### Phase 1 — Loop integrity: verification-gated completion  ✅ code, 🔶 rollout
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
- [x] Deployed in chassis `v1.0.1116` (verified against pod binary)
- [ ] Apply `sql_for_agents/149_page_build_handler_noop_flags.sql`
      (handler no-op exits park item at `needs_human_review` — defence in depth)
- [ ] **Live re-drive** of one gripper-detail `empty_section` item; expected:
      item ends `needs_human_review` (149 applied) or `triaged/failed` with
      "completion blocked: post-fix verification…" (gate) — NOT `complete`

### Phase 2 — Detection: required_fields_missing check  ✅ code, 🔶 rollout
New discovery check: deployed component instances whose schema-required,
LLM-sourced value fields are absent/empty in `content_data`. Flag-only at
`needs_human_review` (no automated handler can fix these honestly), capped 25
per pass. Scope deliberately excludes `query.*`/`site_assets.*`/`pages.*`
sources (render-time; dartsonline's working product grid must not flag) and
image fields (owned by `image_source_unsatisfiable`).

- [x] `check_required_fields_missing.go` + unit tests; in `v1.0.1116`
- [ ] Apply `sql_for_agents/150_enable_required_fields_missing_check.sql`
      (adds check to completeness-discovery-agent)
- [ ] First discovery pass on robot-hands: expect ~6 product-component flags

### Phase 3 — Fail-safe content: meta-commentary guard  ✅ code, ❌ not deployed
`validate_page_content` check 7: blocks LLM prose-about-the-task shipping as
page copy (live case: product-card-with-cta stored a schema apology as
content). **Missed the v1.0.1116 build by ~2 minutes** — needs a rebuild.

- [x] `checkMetaCommentary` + unit tests (narrow, unambiguous patterns)
- [ ] Rebuild + redeploy chassis (v1.0.1117 or rebuilt 1116)
- [ ] Verify against pod: `grep -ac "LLM meta-commentary in content" /app/agent-chassis` → 1

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
