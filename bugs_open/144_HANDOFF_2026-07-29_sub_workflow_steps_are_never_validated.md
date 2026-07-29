# 144 — Steps inside a `sub_workflow` are never validated, by anything

**Filed** 2026-07-29 · **Status** OPEN, unowned · **Class** silent coverage gap (sibling of
`bugs_closed/101`) · **Found by** a council guardian objection on corr
`30a8785b-8cad-4d10-8633-486d81e837e9`, which correctly flagged that a "the only live carrier
is X" claim had been measured incompletely. It had been.

---

## 1. The one-line version

`ValidateWorkflow` runs **once**, on the top-level workflow. Steps nested inside a
`sub_workflow` are extracted and executed directly by the loop action, **bypassing the
validator entirely** — so every invariant the validator enforces is unenforced for them.
**64 live steps across 17 active agents** are in that blind spot. The offline config-key audit
has the same blindness, so neither half can see it.

## 2. Evidence

**The validator is called exactly once, on the top-level workflow** —
`platform/messaging/processor.go:276`:

```go
if err := p.validator.ValidateWorkflow(agentConfig.Workflow); err != nil {
```

**`ValidateWorkflow` does not descend into sub-workflows.** It iterates `workflow.Steps` only
(`platform/validation/workflow.go:50-54`), and `grep -rn "sub_workflow\|SubWorkflow"
platform/validation/` returns **nothing**.

**The nested steps are executed anyway**, pulled straight out of config at runtime —
`platform/orchestration/actions/loop_actions.go:70-77`:

```go
// Get substeps (supports both 'substeps' and 'sub_workflow.steps')
…
// Fallback to sub_workflow structure (used by workflow definitions)
if subWorkflow, swOk := config["sub_workflow"].(map[string]interface{}); swOk {
```

So the nested step runs, and nothing ever validated it.

**MEASURED 2026-07-29** over live `agent_definitions` (`is_active`, non-snapshot, not deleted),
walking every `sub_workflow.steps` in the tree:

| | count |
|---|---|
| top-level `(action, key)` pairs — what the audit sees | 816 |
| nested `sub_workflow` steps | **64**, across **17** agents |
| nested `(action, key)` pairs | 65 |
| pairs that exist **only** nested — seen by nothing | **24** |

Examples of the invisible 24: `assemble_page.inject_head`, `companies_house_fetch.fetch_psc`,
`companies_house_search.sic_filter`, `complete_work_item.commit_sha`, `git_commit.page_field`.

## 3. Why this is worse than a config-key gap

The unknown-config-key detector is only *one* of the things `ValidateWorkflow` does. A nested
step also escapes:

- **`start_step` / empty-workflow checks** (`:36-47`)
- **per-step validation** (`:50-54`) — including *"step must have an action"* (`:71-73`)
- **cycle detection** (`:57-59`)
- **dependency existence** — `depends_on` naming a step that does not exist (`:61-64`, `:96-102`)
- **`fan_out` sub-task topic validation** (`:127-136`)
- the unknown-config-key warning (`:115-124`)

A nested step with a typo'd `depends_on`, or a `fan_out` with no topic, fails at **runtime**
instead of at validation — if it fails loudly at all.

## 4. Why nobody has noticed

**Both halves are blind in the same direction, so they agree with each other.**
`scripts/audit-config-keys.sh` reads

```sql
jsonb_each(ad.default_config->'workflow'->'steps')
```

— top level only. So the offline report and the runtime validator have *identical* coverage,
and cross-checking one against the other can never reveal the gap. Consistent blindness reads
exactly like correctness.

It also means **the config-key ratchet's denominator is wrong**: "151 undeclared actions, 566
pairs" is computed over a population that excludes all 24 nested-only pairs. The gap is
understated, not overstated — nothing already declared becomes wrong, but "we have covered X%"
is measured against the wrong total.

## 5. Fix candidates, ordered by what closes the door

1. **Make the bad state unrepresentable — validate sub-workflows where they are defined.**
   Have `ValidateWorkflow` recurse into `config.sub_workflow.steps` (and `substeps`) and
   validate each as a workflow in its own right. One traversal change, and every invariant
   above starts applying to all 64 steps at once. **Check first whether any of the 64 would
   newly FAIL validation** — this is warn-only for config keys but a *hard error* for a bad
   `depends_on` or a missing action, so a naive recursion could start rejecting workflows that
   run today. Measure before shipping; that is the whole risk.
2. **Fix the audit's SQL in the same change**, or the offline report will keep disagreeing
   with the runtime once (1) lands — a `jsonb_path_query` over `$.**.steps` rather than a
   single `->'workflow'->'steps'`.
3. **Do not do (2) alone.** Making the report see 24 more pairs while the validator still
   cannot check them just moves the blindness rather than removing it.

## 6. Landmines

- **A "which definitions carry action X?" query over `->'workflow'->'steps'` is incomplete**
  and will silently under-report. That is how this was found: a claim that `page-build-handler`
  was the only live `plan_sections` carrier missed `page-rebuild`, whose step is at
  `/workflow/steps/build_pages_loop/config/sub_workflow/steps/plan_sections`. Use a full-text
  `default_config::text LIKE '%<action>%'` as the cross-check — it is crude, and it is the one
  that caught this.
- **The nested `plan_sections` carrier turned out clean** (keys `page_name`, `sections`,
  `site_id`, all declared), so `bugs_open/136`'s opt-in is unaffected. The incompleteness of the
  *method* was real even though the *conclusion* survived — which is the point: it survived by
  luck, not by the check.
- Any fix here is a **platform seam** (it changes what "validated" means fleet-wide) — see
  CLAUDE.md's ordering-exemption rule before shipping it inside another change.

## 7. Verify

```bash
# the 64 nested steps, and whether any would newly fail
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc "
SELECT type, jsonb_pretty(default_config) FROM agent_definitions
WHERE deleted_at IS NULL AND COALESCE(is_snapshot,false)=false AND is_active
  AND default_config::text LIKE '%sub_workflow%';"
```
