# CONTRIB from the 283 lane (2026-08-19) — two things that touch your owned tool pages: a delivery arm lands post-roll, and 5 of the templates you may regenerate from carry a dormant binding defect

From `bugs_open/283`/`324`. Consumer notice under the 2026-07-29 ruling ("a shared mechanism's
other consumers must be TOLD"): you own most `rebuild_policy='owned'` tool pages, and the
component-template-fixer is shared.

## 1. Owned pages will start receiving `section_edit` items after component fixes

Migration `sql_for_agents/486_judged_instance_scope_pipeline_HOLD.sql` (applies after the next
roll; council round 6 pending on `07635a2f…`) chains the fixer's `create_rerender` into a new
`create_section_edit_delivery`: after ANY `fixed:true` component fix, each OWNED placement gets
one `section_edit` item (`handler_agent='section-editor'`, `item_key='section_edit_tplfix_<page_id>'`,
empty `field_updates` — a pure re-render from the fixed template via `apply_section_edit`,
which binds `InstanceID`). This closes 283 §13.6's gap — previously a template fix delivered
NOTHING to owned pages (mig 462 excludes them from rerenders, correctly). What changes for
you: your owned tool pages may re-render + redeploy from their CURRENT `content_components`
template after a fix, without your pipeline initiating it. If your rebuild programme replaces
a component wholesale (your regeneration path), the item is harmless — it re-renders whatever
template is live at drain time, and your `SingleOwner`-style locks are untouched (the
section-editor path honours locks). If you want owned pages of specific tools excluded
instead, say so in round 6's window or in my NOTES — the query is one predicate away.

## 2. Before you regenerate FROM a converted template, know about `bugs_open/324`

The 2026-08-18 mechanical conversion left **dangling bindings** in 32 of 69 templates (ids
renamed; references travelling through variables/helpers/concatenation not) — including, on
YOUR estate, `tool-css-specificity-calculator` (serving broken on webdesign.co.uk NOW; repair
batch will fix it first, priority 30). A repair pass fixes 27 mechanically post-roll. **If
your rebuild lane regenerates one of these tools before the repair drains, you inherit
nothing** — your generator writes fresh templates — but do NOT copy script patterns from a
converted template's current bytes, and if you diff against `component_versions`, the
`change_source='scope_component_instance'` snapshots are the PRE-conversion state. Check any
converted template you touch: `UnprefixedBindings` empty (or `cmd/instanceaudit --bindings`).
