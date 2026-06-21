# Running notes — checkpoint (uu): hardened reconciler shipped (decision 1B)

Date 2026-06-21. Continues (tt). Append to `running_notes.md`.

## Decisions taken

- **1B — schema-sourced reconciler.** Adopted. The render-time reconciler now sources its
  expected per-item keys from the component's own `input_schema` instead of the section plan's
  `llm_field_specs`. Removes the plan-freshness dependency entirely: the fix lands on a
  `page-rebuild` whether or not the rebuild reconstructs the section plan, and whether or not
  the prompt change is present. The prompt migration + `plan_sections` `ItemFields` population
  are now **optimisation only** (they steer the LLM so fewer remaps are needed), not a
  correctness requirement.
- **2 — non-fatal.** Kept. Unrecoverable item (expected key missing, no case/separator variant
  and no known synonym) is logged at ERROR and the build continues, shipping that one sub-field
  empty. Rationale: a missing sub-field is cosmetic; failing a whole page build over it is
  higher blast-radius; the ERROR makes it visible for follow-up. Revisit only if these ERRORs
  actually appear.
- **3 — skipped.** No longer need to determine whether `page-rebuild` rebuilds or replays the
  section plan, because 1B makes that immaterial to correctness.
- **No component-level regeneration trigger exists** (user confirmed). So the remedy for the
  already-deployed broken cards is a whole-index `page-rebuild`, which regenerates *all* index
  sections — hero, the 13-item FAQ, and the method narrative get rewritten too. Accepted as the
  cost; `page-rerender` cannot be used (it concatenates stored `rendered_html`; see (tt)).

## Change made (this session)

File `v3_site_actions.go` (in outputs), `RenderComponentAction` + helper, package `actions`:

- **Wire-in** (content-extraction point in `RenderComponentAction`): replaced
  `if specs := datahelpers.ExtractNestedField(params.CollectedData, "current_section.llm_field_specs"); specs != nil { reconcileGeneratedItemKeys(contentData, expectedItemFieldsFromSpecs(specs), …) }`
  with
  `if comp != nil && len(comp.InputSchema) > 0 { reconcileGeneratedItemKeys(contentData, expectedItemFieldsFromComponentSchema(comp.InputSchema), …) }`.
  `comp` is the component already resolved/loaded fresh earlier in the function
  (`GetComponentByID` / `GetComponentWithFallback`), so its `input_schema` is always current.
- **Helper**: removed the now-dead `expectedItemFieldsFromSpecs`; added
  `expectedItemFieldsFromComponentSchema(inputSchema map[string]interface{}) map[string][]string`.
  It walks `inputSchema["fields"]`, keeps only fields with `source:"llm"` (so reconciler reach
  stays identical to the old writer-loop scope — query-resolved/static arrays are left alone),
  and reuses `extractArrayItemFields` (from `plan_sections_action.go`) to read each field's
  `items`/`item_schema`. Returns empty (reconcile no-ops) when there's no `fields` map or no llm
  array fields.
- `reconcileGeneratedItemKeys`, `itemKeySynonyms`, `synonymsFor`, `normaliseKeyForMatch`
  unchanged. No import changes (`datahelpers` still used ~93× elsewhere). Braces/parens/brackets
  balanced; no Go toolchain in the authoring env, so `go build` + `make test-unit`
  (`go test ./...`) before deploy.

**Cross-file dependency**: this reuse means `v3_site_actions.go` and the patched
`plan_sections_action.go` (which introduced `extractArrayItemFields`) must deploy in the same
chassis image. They were always going together (the `ItemFields` population feeds the prompt
half), so no new constraint — just don't ship one without the other.

## Confirmed during the hardening review

- `Component.InputSchema` is `map[string]interface{}` of shape `{"fields": {...}}`
  (`component_library.go` line 37, unmarshalled line 250), the same shape `plan_sections`
  already walks — so the schema walk is type-safe.
- Failure modes degrade safely: a component whose schema lacks `fields`, or whose field value
  isn't a map, yields an empty expected-map → no-op (never a panic; `extractArrayItemFields`
  uses comma-ok asserts).
- Broader activation (now fires for every llm array component, not only plan-carried ones) is
  safe: the reconciler only acts when an expected key is *missing* and never moves a synonym
  onto a key that is itself expected (`wantSet` guard).
- One theoretical regression vs the old plan-sourced version: if a component's
  `input_schema.items` is edited to diverge from its `html_template` *after* a good plan was
  built, the fresh-schema reconciler targets the new (wrong) keys. But that component is already
  misconfigured (fresh generation would break too, since the prompt also derives from
  input_schema). The governing invariant — a component's schema `items` must match its template
  tokens — is the right thing to hold; the reconciler enforces consistency toward the current
  schema. (Note: info-card-grid violates a related invariant via literal `<no value>` in its
  stored template — separate issue, see earlier notes.)

## Deploy / rollback (from the makefile)

- Single global `IMAGE_TAG`; one `agent-chassis` binary runs all dynamic agents. Targeted
  deploy: `make quick-agent-update IMAGE_TAG=v1.0.1067` (build → push → kustomize → DB
  `agent_definitions.image_tag` → restart `agent-chassis` deployment) + `make
  update-and-restart-orchestrator IMAGE_TAG=v1.0.1067` (generic-orchestrator statefulset). A
  full `make release` would bump *every* service to the new tag — use the targeted path for an
  isolated change.
- Rollback: re-point to `v1.0.1066` (image already exists, no rebuild) via
  `make update-agent-images-v2 IMAGE_TAG=v1.0.1066` + rollout restart; prompt via the 019
  down-migration. Reconciler stays low-risk to leave deployed during a prompt-only rollback now
  that it's schema-sourced (no plan/prompt coupling).

## State after this checkpoint

- Three artefacts now final in outputs: `plan_sections_action.go` (ItemFields population),
  `v3_site_actions.go` (hardened schema-sourced reconciler), `019_pcw_prompt_item_fields.sql`
  (+ down). Prompt migration already applied; code awaits a chassis image bump.
- Runbook `RUNBOOK_pcw_item_fields_fix.md`: page-rebuild (index only) is the route; the
  plan-freshness caveat in §3 is now moot for correctness (1B), though the §3 page-rebuild
  workflow dump is still harmless to run; the schema-sourced hardening it recommends is now the
  shipped behaviour.
- Open/parked: whole-index copy churn from the rebuild (no component-level trigger to avoid it);
  info-card-grid `<no value>` template; idea.uk parked gaps (hero/CTA/contact form/nav-footer).
