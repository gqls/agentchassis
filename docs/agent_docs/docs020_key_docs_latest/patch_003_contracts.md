# Patch for 003_contracts_and_standards_v7.md

Context: this doc captures cross-agent contracts (naming, JS
separation, CSS inheritance, linkage). It doesn't currently describe
`StoreGeneratedComponentAction`'s return shape, which downstream
workflows branch on. The April 2026 regeneration work added two new
status values and a side-effect worth documenting: the action now
raises `needs_rerender` work items when it regenerates existing
components.

## Change

Add a new top-level section after "JS Content Separation Contract"
and before "Site Component Linkage Contract". Place it around line
156 (after the `---` that closes the JS separation section).

```markdown
## Component Creation & Regeneration Contract

`StoreGeneratedComponentAction` persists LLM-generated component
templates to `content_components`. It has two primary paths that
downstream workflows may need to distinguish:

### Return status values

| `status` | Meaning | Primary key behaviour | Side effects |
|---|---|---|---|
| `created` | No existing active component with this `function`. Row INSERTed. | New UUID generated | Version 1 snapshot written to `component_versions` |
| `regenerated` | Existing component found. Old state snapshotted, new template UPDATE in place. | UUID preserved — foreign keys in `page_components`, `site_components`, etc. remain valid | Snapshot written to `component_versions` with `MAX(version_number)+1`; one `needs_rerender` work item created per affected site |

The `already_exists` status from earlier versions is REMOVED. Every
call now either creates or regenerates — the LLM's output is never
silently discarded just because a row with the same function already
exists.

### Rejection path

Layer 1 pre-store validation runs BEFORE the create/regenerate
branch. If the new template fails validation (zero template
placeholders despite a populated schema, malformed structure, etc.),
the action returns an error and does NOT touch the existing row.
An existing broken component stays in place — rejection never makes
things worse.

### Return payload (regenerated)

```json
{
  "component_id": "<existing-uuid>",
  "function": "provocation-feed",
  "status": "regenerated",
  "previous_version": 2,
  "new_version": 3,
  "pages_marked_rebuild": 4,
  "affected_sites": 2,
  "rerender_items_created": 2,
  "quality_score": 82,
  "quality_issues": []
}
```

### Return payload (created)

```json
{
  "component_id": "<new-uuid>",
  "function": "foo-section",
  "status": "created",
  "new_version": 1,
  "quality_score": 85,
  "quality_issues": []
}
```

### Downstream contract

Workflow steps consuming `StoreGeneratedComponentAction` output:

- MAY branch on `status` to distinguish the cases
- MUST NOT assume `component_id` is newly minted — it might point to
  a long-lived component whose template was just replaced
- SHOULD NOT create their own `needs_rerender` items when
  `status = "regenerated"` — the action has already done so (one per
  affected site), deduped via `item_key = component_regen_rerender:<uuid>`

### Version history preservation

Every create and regenerate writes a row to `component_versions`
with `change_source` recording the originating work item's `source`
field. See 014_site_snapshots_and_revert.md for population patterns
and 020_tool_lifecycle.md for the `update_component_html` path.
```

## Rationale

The regen semantics are visible to any workflow calling
StoreGeneratedComponentAction. Documenting the three outcomes
(created, regenerated, rejected), what state the component_id is in
each case, and the "don't duplicate rerender items" downstream rule
lets future workflow authors make correct assumptions without
reading the Go source.

The placement between JS separation and site linkage is thematic —
both bookend the component-creator flow (prompt → JS extraction →
store → linkage).
