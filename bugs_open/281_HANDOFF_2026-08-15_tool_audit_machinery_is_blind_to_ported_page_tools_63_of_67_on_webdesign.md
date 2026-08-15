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
  ~~the healthy lane for ported instances~~), 2026-08-15, status `triaged`.
  > **CORRECTED 2026-08-15 ~18:3xZ:** the section-editor is healthy for real FORK slots — its
  > 171 completions are those — but it REGENERATES a section through an LLM, which is
  > structurally wrong for a ported blob: both owner-gate items failed with **SLOT FLOOR
  > REFUSED** (mindmap 28→4 class attributes, occlusion 11→4; bugs 178/253's floors, working
  > as designed — nothing was written). So ported instances have NO safe framework editor,
  > which is this bug's thesis restated at the third mechanism. The two pages were repaired
  > by guarded surgical `replace()` on `rendered_html` (archive trigger banked pre-states)
  > + `page_rerender`; items cancelled with the supersession recorded in `result`.
- The 4 real tool components got owner-requested `audit_tool` items the same day
  (`created_by='owner-visual-gate-20260815'`).
- The other 62 ported tools remain unexamined for legibility/junk-content until this bug is
  fixed; acceptance rotation continues but only grades documented criteria.

## Addendum 2026-08-15 (fixing session, `bugfix_281_tool_audit_ported/`) — three findings and the fix

**Finding A — mechanism 2 HAS fired, via a third producer this file did not name.** "None
exist" above is true of `audit_tool` items only. `check_tool_acceptance.go` (Tier 2) already
admits ported instances (TL-033) and filed `improve_tool` → tool-improver with the SHARED
`component_id`; tool-improver's `update_component_html` then rewrote the ported-page
wrapper's `html_template` and flipped every placement to `pending` — `component_versions`
`[MEASURED]`: v1 (2026-08-05, pre-edit = the 77-char `{{.body}}` passthrough) and v3
(2026-08-14 18:48Z, trigger `tool_acceptance:asset-formatter:<webdesign>`, complete). **Live
latent hazard as of 2026-08-15:** the shared template is 8,864 chars of asset-formatter tool
markup (`{{.body}}` still present); all 115 instances sit at `build_status='pending'`;
verified NOT yet propagated (every `pc.updated_at` == the write instant, `rendered_html`
content unchanged, no `component_template_corrupted` item). Restoring the passthrough (seed
208 / v1) is the owning lane's call — flagged to the owner.

**Finding B — Tier-4's judge cannot file a fix for a ported tool either.**
`tool_acceptance_actions.go` `JudgeAcceptanceResultsAction` (~:867-873) re-derives the
component by `cc.function = <subject key>` — no `content_components` row carries a ported
tool's function (the shared row is `ported-page`), so `componentID==""` and the run lands in
the "no content_components row … route manually" arm: a verdict is produced, no item is
filed. Its `LEFT JOIN … LIMIT 1` is the same arbitrary-instance shape as `load_tool`. Not
fixed here (Tier-4 path); recorded for whoever picks it up.

**Finding C — the eligibility count is 66, not 67.** `tool-ab-test-calculator`'s page carries
BOTH a fork and a ported-page instance; clause (a) audits the fork, clause (b) rightly skips
the page. Post-fix census should read 66 subjects.

**Fix (Track 1, TL-042; council-submitted; Go rides the next roll, seeds 425/426 applied):**
`check_tool_health` enumerates the ladder's population per page instance; ported findings from
it AND from `check_tool_acceptance` file a new handler-less `ported_tool_fix`
(`needs_human_review`, key `ported_tool_fix:<check>:<subjectKey>:<site>`); item keys and
cooldowns are per instance for ported tools (forks byte-identical); template-contract checks
run only for forks; Tier-2 audit queueing capped 12/run. tool-auditor's and tool-improver's
`load_tool` pin `component_id AND spec.page_id`; the auditor reviews `source_html` and routes
ported findings to human review only. `update_component_html` refuses a `component_level<>'tool'`
component placed on >1 page unless `allow_shared_component_write` (opt-in, default OFF).
Decomposition of the 63 (owner's ask) is a proposal with preconditions, not executed —
`bugfix_281_tool_audit_ported/PROPOSAL_2026-08-15_decompose_webdesign_tools.md`.
Status: fixed at source, OPEN until the Go is live and the first-sweep census in the RUNBOOK
is recorded.

## Contribution 2026-08-15 ~17:00Z (filing session, with the owner) — Finding A hazard CLOSED; replacement lane opened

> **Finding A now has its own case: `bugs_open/285`** (owner-directed) — the full two-firing
> timeline, root-cause decomposition, restoration record, and the residuals
> (`fix_component_template` unfenced; Tier-4 judge). Close-out of the write-path defect tracks
> THERE; this file stays about audit coverage.

- **The owner ruled: do both** (Track 1 audit fix — yours — and accelerate native rebuilds).
- **Finding A's latent hazard is closed.** With the owner's sanction this session restored the
  shared wrapper's `html_template` from the poisoned 8,864-char state to **v3's pre-edit
  snapshot** (4,664 chars, `{{.body}}` intact) — *not* v1: v1 (77-char seed) predates
  legitimate 08-05/08-08 edits; v3 is the state the poisoning write itself snapshotted.
  The poisoned state is banked as **v4** (`change_source='manual-restore'`) so nothing is
  lost. The **114 `pending` placements were un-flipped to `deployed`** (guarded on
  component+status; `pending` is a live re-render signal — `chrome_link_policy.go:133` — and a
  re-render against the stub `content_data` would have clobbered rendered pages). Verified
  after: template 4,664/passthrough, 0 pending. Your `allow_shared_component_write` guard
  remains the fix that makes this unrepresentable; this was the cleanup.
- One correction to the addendum's Finding A text for the record: the single `deployed`
  placement (`learn-ai-builders-content-first`, updated 18:51Z) was checked — its
  `rendered_html` is clean; the 18:51 touch was not propagation.
- **Native rebuilds of the 63 are now an owned lane**:
  `docs024_key_docs_latest/webdesign_tool_rebuilds/` (PLAN/RUNBOOK/NOTES). Found on the way:
  the suggester→generator path only ADDS tools (gap analysis); the one native-beside-ported
  page (`tool-ab-test-calculator`) served a raw `{{.section_heading}}` and kept its ported
  slot — retired to `removed` + repair item filed. Pilot replacement
  (`tool-aspect-ratio`, `add_tool_novel_webdesign.co.uk`) is in the queue; rich apps are
  excluded from rebuild by decision (PLAN §3) and wait on your decomposition proposal.
