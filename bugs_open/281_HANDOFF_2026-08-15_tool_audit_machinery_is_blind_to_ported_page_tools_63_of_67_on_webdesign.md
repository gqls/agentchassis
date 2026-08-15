# 281 — the tool-audit machinery is blind to ported-page tools: 63 of 67 tools on webdesign.co.uk are invisible to `tool_health`, and `load_tool` would audit an arbitrary instance if pointed at one

Filed 2026-08-15, from the owner's visual gate on webdesign.co.uk (bug 122's second canary).
The owner found the Mind Map Studio tool with illegible pale-on-pale UI text, keyboard-mash
seeded content ("Centrdgsdgsdgsdal Idea") and no usage guidance — and **no work item had ever
been filed against it**. This file explains why not, with the two mechanisms quoted.

**First-hand verification substituted for a 090 run, stated per the 2026-07-31 ruling:** both
mechanisms below are single visible clauses in code/config that were read directly, and the
census is one query, reproduced here. There is no inference step a diagnosis loop would
independently re-derive.

## Mechanism 1 — the producer never sees ported tools

`platform/orchestration/actions/discovery_checks/check_tool_health.go:68` scopes the entire
check (both the Tier-1 structural audit that files `improve_tool` and the Tier-2 queue that
files `audit_tool` for LLM review):

```sql
AND cc.component_level = 'tool'
```

But on webdesign.co.uk `[MEASURED 2026-08-15]`:

| component_level | function | tool pages |
|---|---|---|
| section | `ported-page` | **63** |
| tool | tool-ab-test-calculator, tool-css-specificity-calculator, tool-css-unit-converter, tool-llm-cost-calculator | 4 |

The `ported-page` component is one shared `content_components` row
(`a7daa5c5-8cfd-4f2c-8e09-de6abcb637ef`, **115 page instances fleet-wide**); each instance's
actual tool code lives in its `page_components.rendered_html`. `component_level='section'`, so
the health check's query returns 4 of 67 tools and the other 63 have never been examined. The
Mind Map Studio (`/tools/mind-map/index.html`) is one of the 63 — which is why its defects
waited for a human to stumble on them.

## Mechanism 2 — the consumer cannot address a ported instance either

The tool-auditor's `load_tool` step (live `agent_definitions` row, `type='tool-auditor'`):

```sql
... FROM content_components cc
JOIN page_components pc ON pc.component_id = cc.id
JOIN pages p ON pc.page_id = p.id
WHERE cc.id = $1::uuid AND cc.is_active = true LIMIT 1
```

`$1` is `spec.component_id` and **`spec.page_id` is never used in the WHERE**. Pointed at the
shared ported-page component, the join yields ~115 rows and `LIMIT 1` picks an arbitrary one —
the audit would run against a random site's random tool and *complete successfully*. So
hand-filing `audit_tool` items for ported tools is not a workaround; it produces confident
reviews of the wrong artefact. (This was caught before any such item was filed; none exist.)

## What this is NOT

- Not the decomposition bug (`bugs_open/263`) — that is about ported wrappers being dissolved
  during decomposition; this is about audit coverage.
- Not a claim that ported tools get no checks at all: **Tier-4 acceptance CAN and does target
  ported instances** (`acceptance_run:animated-favicon:6b49db8e-…` completed with
  `component_id=a7daa5c5…` + per-item `page_id`/`function` — the acceptance spec carries the
  instance identity the auditor's spec lacks; 15 acceptance runs on webdesign in the last 30
  days). But Tier-4 grades documented acceptance criteria and declines undocumented tools, so
  it does not substitute for the structural/legibility audit.

## Fix candidates, ordered by what closes the door

1. **Teach `check_tool_health` to enumerate tool PAGES, not tool components** — e.g. scope on
   `p.page_type='tool'` (or slot `ported-page` on tool pages) and audit
   `pc.rendered_html` per instance. The acceptance producer already demonstrates the correct
   identity shape (component_id + page_id + function per item). `load_tool` must then resolve
   by `page_id` (or `page_component_id`), falling back to component only for real tool forks.
   This makes the blind spot unrepresentable: a new ported tool is auditable the day it ships.
2. Narrower: keep the level filter but add a second branch for `ported-page` instances on
   `page_type='tool'` pages. Same coverage, two code paths to keep in step.
3. "Operators should hand-file audits for ported tools" — **not a fix** (and currently
   impossible: mechanism 2 loads an arbitrary instance).

## How to verify a fix

- Census before/after: items filed by `tool_health` on webdesign should cover ~67 tools, not 4.
- The motivating case: the Mind Map Studio's illegible controls must be *detectable* by the
  fixed check (its pale-on-pale styles are in its own `<style>` block — note the existing
  `hardcoded_colors` check already looks for exactly this class of thing, it just never ran).
- Negative control: a non-tool ported page (e.g. `/learn/...` prose ports) must NOT start
  drawing `improve_tool` items from the tool branch.

## Interim state (so nobody re-derives it)

- The owner's three Mind Map defects are filed as
  `section_edit:owner-gate:tool-mind-map` (section-editor lane; 171 complete / 2 failed —
  the healthy lane for ported instances), 2026-08-15, status `triaged`.
- The 4 real tool components got owner-requested `audit_tool` items the same day
  (`created_by='owner-visual-gate-20260815'`).
- The other 62 ported tools remain unexamined for legibility/junk-content until this bug is
  fixed; acceptance rotation continues but only grades documented criteria.
