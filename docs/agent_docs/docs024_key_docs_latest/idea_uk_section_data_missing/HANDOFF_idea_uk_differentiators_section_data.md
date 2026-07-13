# Handoff — idea.uk differentiators / section-data investigation

Created 2026-06-19. Start a fresh chat from this. Deeper history is in
`running_notes.md` (checkpoints around (qq)/(rr)) and the design/build picture is in
`RUNBOOK_idea_uk_chassis_site_and_vm_deploy.md`. This file is self-contained enough to begin without them.

---

## The problem (the thing to solve)

On idea.uk's freshly built index page the **differentiators section renders its heading but seven empty cards**
— every item title and description blank — while the same page's writer-generated method narrative and 13-item
FAQ populated correctly. Since `reconcile_section_data` (wired 9 June) only re-triggers pages whose deferred
section data is *query-resolvable*, we need to establish where a differentiators component's items are meant to
come from — **query-resolved section data, a human-entered spec field, or page-content-writer prose** — and fix
whichever link is leaving them empty.

## What you need to know before digging

- **The page builds and deploys now.** A coordinator result-extraction fix (`result_spec.go`) was shipped and
  validated on 2026-06-19; idea.uk's index built in a real ~17-minute pass and deployed to B2. So this is **not**
  a build/coordinator problem — the writer produced a full page; the differentiators are the one empty part.
- **`reconcile_section_data` IS wired** — `registry.go` line 914, handler `ReconcileSectionDataAction`, category
  `site`, `IsLocal: true`, description *"Re-trigger pages whose deferred section data is now query-resolvable."*
  So the answer is **not** "wire the reconciler" (an earlier note said it wasn't wired — that note was stale).
- **The scope clue is in that description.** The action only refills section data that is *query-resolvable*
  (the tools/guides-list kind). So three possibilities for the differentiators items, and the fix differs per
  case:
  1. **Query-resolvable section data** — the reconciler should fill them, but the query yields nothing (or the
     section isn't in its scope). Fix is in the reconciler / the query / the section's data binding.
  2. **Human-entered spec field** — the reconciler correctly skips it (this is the case for the pricing section,
     whose `tier_1_*` come from a human `site_specs.pricing`). Fix is to capture the data into specs, or change
     the component's source, or drop the section.
  3. **Writer-generated prose** — the page-content-writer was supposed to fill the items and didn't. Fix is in
     the writer's section selection / generation for this component type.
- The other parked content gaps on the same page (for context, not part of this task): empty hero + CTA buttons
  (no destination pages), a dead contact form posting to `#contact`, thin nav/footer. Leave those to the main
  chat unless they turn out to share a root.

## Key facts / IDs

- idea.uk `site_id` = **97ed2f64-65ca-4b67-8a98-dfd8195a0d3a** (a FRESH build — has a `submission` spec; NOT an
  adoption).
- DB: `kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db`.
- gamesdesign.co.uk `site_id` = **e33263f4-74f8-494f-b191-546845dbbddf** — an ADOPTION build; useful as a
  contrast (did its differentiators populate? if so, from what source?).
- Deployed-page symptom: the differentiators component emitted seven `<div class="differentiator-item">` blocks,
  each with empty `<h3></h3>` and `<p></p>`, under a real generated heading ("Why ideas… survive longer than a
  chatbot"). The method section and the FAQ on the same page are fully populated.

## Bundle to gather for the chat

**Code (Go):**
- `reconcile_section_data_action.go` (`ReconcileSectionDataAction`) — what "query-resolvable" covers and how it
  decides which sections to refill.
- `registry.go` — confirms wiring (already in hand).
- The `page-content-writer` agent definition + the section-handling actions (`select_sections`,
  `process_sections`, `compile_page_sections` — likely in `v3_site_actions.go`) — to see whether the writer is
  expected to generate differentiator items and why it didn't.
- The action that gives a section its data before compile (the writer-prose vs section-data split), and
  whichever agent owns the `reconcile_section_data` step in the build flow.

**Data (run and paste):**
- The differentiators component's `content_components` row — its template and the data fields it expects
  (e.g. `items[].title/description`). This alone may reveal whether it's a structured-data component or a prose
  one. (Check `\d content_components` for the template/data-contract column name.)
- idea.uk's `needs_section_data` work items — for each flagged section, the source path (`query.*` vs a human
  field vs none) and target fields.
- The stored section/component data for the differentiators section on idea.uk's index — empty list, list with
  empty fields, or nothing.
- `site_specs` for idea.uk — is there a differentiators / USP / why-us source aspect, and what shape.

**Docs:**
- `026_component_regeneration_flow.md` and doc 030 (referenced in the registry descriptions for the reconcile
  actions); any FOCUS note on the section-data deferral.
- Whatever defines a component's data contract — how a component declares required section data, and the rule
  for what the writer fills vs what section-data supplies.

## First diagnostic queries (verify column names against `\d` first)

```sql
-- 1. The differentiators component's data contract (template + expected fields).
--    Confirm the template/data column via \d content_components first.
SELECT name, display_name, "function", category, component_level
     -- , <template/default_data column once confirmed>
FROM content_components
WHERE "function" = 'differentiators' OR name ILIKE '%differentiator%';

-- 2. Enumerate idea.uk's work-item types, then read the section-data specs.
SELECT DISTINCT item_type FROM site_work_items
WHERE site_id = '97ed2f64-65ca-4b67-8a98-dfd8195a0d3a';

SELECT id, item_type, status, jsonb_pretty(spec) AS spec
FROM site_work_items
WHERE site_id = '97ed2f64-65ca-4b67-8a98-dfd8195a0d3a'
  AND item_type ILIKE '%section%';     -- adjust to the real item_type from the DISTINCT result

-- 3. Any differentiators/USP source in specs (confirm the value column via \d site_specs).
SELECT aspect  -- , <value/content column>
FROM site_specs
WHERE site_id = '97ed2f64-65ca-4b67-8a98-dfd8195a0d3a'
  AND aspect ~* 'differ|usp|why|reason|advantage';
```

Then read the stored components for the index page (use the same components-per-page query that earlier showed
index at 0 components) and inspect the differentiators row's data.

## House rules (fresh chat won't have them)

Go not Python; plain language, no hype/flattery; confirm live schema/facts before asserting or writing SQL
(a `0 rows` result is not decisive — check the query first); reuse/alter existing functions before new; fix the
framework structurally over one-off patches; British English; low risk appetite; reasonable step sizes; ≤1
question per reply where possible; don't create summary docs unless asked. Every agent is an orchestrator owning
a workflow of steps calling actions; keep workflows simple with complexity in Go actions; check DB schema before
SQL; no `logger.Debug` (won't surface). k8s namespaces `ai-persona-system` and `kafka`; cluster
`personae-kafka-cluster`. Deploy = GitHub → GH Actions → Backblaze B2.
