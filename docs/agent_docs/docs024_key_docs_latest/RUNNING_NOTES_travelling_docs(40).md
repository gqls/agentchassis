
---

## T33 — 2026-07-21 — bug 024 re-scoped: generic delivery path is forbidden for tools; section-editor delivers (PROVEN LIVE)

**Ground truth at start:** chassis+adapter **v1.0.1144** (not 1140). Call site 2 of
defect 5 (`3cb92dae4`) is an ancestor of the v1.0.1144 build commit `f9dfa0205`
and pod-verified (`isSelfContainedSection`=1, `toolTemplateValid`/`componentTemplateValid`=4,
`loadSingleComponentSchema`=5 in `agent-chassis-59c675c4f-pxr9f`). Migration 180's
config intact on tool-improver (`spec_literal.reason`, `spec_paths.component_id`,
`item_key_suffix_field`, `recurrence_expected`). Dedup index `idx_swi_dedup` puts
`unresolved`+`cancelled` in the TERMINAL set → the two stale pre-180 `needs_rerender`
rows are inert. Next free migration **183** (ledger+dir max 182).

**PROBE 1 — reason-bearing page_rerender, direct (kafka).** Hand-inserted then, when
the lane didn't move, drove directly: `system.agent.generic.requests`,
`config.agent_type=page-rerender`, `input_data.spec.reason=section_data_resolved`,
unique item_key. Corr `fdcff19b`. Reached `save_sections` — **first time in this
bug's history past defect 6** — and FAILED:
`save_page_sections: page … is rebuild_policy=owned (tool/widget-owned): a generic
section save would clobber it. Use apply_section_edit … Refusing to overwrite.`
- `rerender_page_sections` did NOT escalate first (reached save = escalated:false),
  so **defects 3+5 are sufficient** — the render logic works, only the WRITE is
  refused. Nothing was written/deployed (guard is before any write).
- Guard = `save_page_sections_action.go:138-160`, "guard rail 1, experience loop",
  `fb89f1071`, migration **164**. `164_pages_rebuild_policy.sql`:
  `UPDATE pages SET rebuild_policy='owned' WHERE page_type='tool'` (every tool page
  owned); note *"page_rerender/assembly is NOT gated — re-assembly is how owned
  pages deploy"* (they gated the section-WRITE, kept assemble open, deliberately).

**PROBE 2 — section-editor content_edit (the sanctioned path).** Corr `c3828d17`.
`input_data`: site_id, domain, `page_component_id=45229c85…`, page_name, slot_name,
`edit_type=content_edit`, `field_updates={}` (pure re-render; content_edit REQUIRES
field_updates or replacement_content_data — `applyContentEdit:594`). **COMPLETED.**
`page_components.rendered_html` 9,901→**10,705**; `.ltb-row-grid` now
`display:flex; flex-wrap:wrap; gap:0.75rem; align-items:end; min-width:0; max-width:100%`;
`pc_build=deployed`; **live page (curl) carries it**. section-editor workflow:
`load_edit_context → apply_section_edit → git_commit → update_page_status`.

**Durable traps banked this turn:**
- **The dispatch lane can be cron-starved.** page-rerender saw ~1 completion in 6h;
  33 triaged items queued. `build-dispatch-loop` is a STATELESS agent that idles out
  after 600s and dies — it is not a persistent poller; work-item dispatch is a
  scheduled (cron) burst. A hand-inserted work item can sit for hours. To prove a
  render NOW, drive the orchestration directly via kafka (the 085 envelope) instead
  of waiting on the lane.
- **UTC/BST clock trap (again).** A background monitor's `date` prints BST (UTC+1);
  the DB `now()` is UTC. A probe "waiting 65 min" was actually 6 min. Always read
  the DB clock, never the shell's, for age.
- **A guard can be invisible behind an earlier defect.** The ownership guard has
  been live since ~07-17 but NO prior 024 proof run reached it — each was blocked
  earlier (escalation pre-defect-3; suppression at defect 6). "We fixed 5 of 6 and
  the 6th is the last blocker" was wrong: fixing the 6th only exposes the 7th
  (the guard) that was there all along. Prove end-to-end BEFORE assuming the count.
- **content_edit needs a non-nil field_updates/replacement_content_data** or it
  errors "requires either 'field_updates' or 'replacement_content_data'"; `{}`
  satisfies it and is a pure re-render (merge-nothing).

Categories: (diagnosis, correction, proof)
