# HANDOFF — tool-improver truncates a whole component and saves the wreckage over the durable source

**Created 2026-07-18 from the travelling-docs / self-verifying-tools workstream**
(HANDOFF_2026-07-10 T26 has the surrounding run). **Severity: destructive, and
it fires silently.** A fix agent destroyed a working tool's durable source and
reported success; nothing in the pipeline noticed. The live page survived by
luck of timing, not by design.

**Evidence site:** `tool-loot-table-balancer`, gamesdesign.co.uk
(`site_id e33263f4-74f8-494f-b191-546845dbbddf`, component
`3862f72f-8a67-4dda-b0ef-6be83bc22fe6`). DB:
`kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db`.

---

## What happened

`tool-improver` was dispatched to fix a mobile overflow. Its `improve_tool` step
(claude-sonnet-5) rewrote the whole component and returned
**`output_tokens=8000` against `max_tokens=8000`** — it hit the ceiling exactly,
i.e. the completion was cut off mid-stream. The truncated text was then written
straight to `content_components.html_template`:

| State | Length | Contents |
|---|---|---|
| Working (attempt 2, 2026-07-17 20:00) | 10,272 | complete: `<style>`, markup, `</script>` |
| After this run (2026-07-18 10:01) | **1,253** | **CSS only** — no `<script>`, no `<div>`, no `<fieldset>`; ends mid-declaration (`font-weight: bold,`) |

An intermediate version row from the same write (6,765 chars) is *also*
truncated — it ends mid-JavaScript (`var tiers = ['Common', 'Uncommon', 'Rare',
'Epic`). So the run produced two successive mangled writes.

**The fix agent reported success.** It wrote a machine `fix` note claiming it had
set `min-width:0` on grid items and made the columns responsive. That claim is
not false in intent — the reasoning was right (see bugs_open/010) — but the
artifact it persisted was a fragment.

**Why the site did not go down:** the page had been rendered from the previous
complete component and the render did not re-propagate before this was caught.
One `refresh`/rerender from the durable source would have shipped a page with no
markup and no JavaScript. This is the `rendered artifact vs durable source`
distinction the vonc footer arc established — here it accidentally saved us.

## Root cause

Two independent failures, both required:

**1. Whole-component writers were sized inconsistently.** The agents that emit an
entire component had different ceilings for the same job:

| Step | max_tokens (before) |
|---|---|
| `tool-recreation-handler.recreate_tool` | **64000** ✅ correctly sized |
| `tool-improver.improve_tool` | **8000** ← truncated here |
| `tool-generator.generate_tool_html` | **8000** ← same exposure at BIRTH |

The birth of this very tool used **6094/8000** — the generator was one slightly
larger tool away from shipping a truncated component to a live site.

**2. Nothing checks that what is being saved is still a component.** The save
path accepted a CSS-only fragment as a valid `html_template`. No length-collapse
check, no structural check, no "did the `<script>` survive" check. Compare the
loop-integrity workstream's verifier-registry gate for work items — the same idea
has never been applied to component writes.

## What has been done (2026-07-18)

- **Component restored** from `component_versions` — the last complete version
  (10,272 chars, `</script>` intact, matching what the live page serves). The
  truncated state is preserved in `tmp_loot_truncated_20260718` for inspection.
- **Migration 168 applied**: `improve_tool` and `generate_tool_html` raised
  8000 → **32000** (snapshots taken). This removes the immediate exposure only.

## STATUS 2026-07-18 (later) — both real fixes are BUILT; both await an image roll

Candidates (a) and (b) below were **built by other threads** while this case was
open. Verified by this thread rather than rebuilt:

- **(a) completeness guard — BUILT.** `platform/orchestration/actions/component_write_guard.go`
  (`componentRegressionIssues`), wired into `update_component_html_action.go`
  — the exact path that destroyed this component. On a blocked write it
  hard-errors: component left untouched, step fails, and a
  structured row lands in `agent_error_log`
  (`error_code='component_write_regression_blocked'`). The
  `allow_structural_regression` escape hatch is step-config-only, so an agent
  cannot talk its way past it. Three comparative checks: size collapse (<50%
  retained), unterminated `<script>/<style>/<section>` where the current row was
  balanced, and a mid-token tail where the current row ended on a closed tag —
  all gated on "truncation cannot grow an artifact", and calibrated against all
  29 live `component_versions` transitions (1 block = this incident, 0 false
  positives).
- **(b) truncation detection — BUILT.** `f32b208e5` decodes `stop_reason` /
  `done_reason` in `GenerateText` and hard-errors on a capped completion
  (`bugs_open/008`). Council-reviewed.

> **CORRECTED 2026-07-18** by the thread that wrote the guard. The section above
> was written by a second thread which found the guard's code **uncommitted in
> the shared working tree**, assumed another thread had finished it, and
> reasonably inferred the rest. (Its independent verification against the real
> stored artifacts stands and is valuable — that part is confirmed.) Two claims
> were wrong:
>
> **(i) The guard is wired into `update_component_html` ONLY** — not into
> `store_generated_component`. The birth path keeps its own separate,
> schema-shaped gate. A proposed consolidation of the two recorders was
> **withdrawn** after the council gate's edit-quality and guardian seats objected
> that it was scope creep on working code; see the NOTE in
> `store_generated_component_action.go`.
>
> **(ii) There was NO `error_step` routing to `needs_human_review`.**
> `tool-improver`'s `update_component` step had `error_step = null`, so a refusal
> would reach `failWorkflow`: orchestration FAILED, work item left to the reaper,
> **no note** — a thread reading that would see a generic failure, not "a fix was
> rejected as mangled". Migration **`169_tool_improver_refusal_path.sql`** adds
> that route (`refuse_mangled_write` → `note_refusal` → `complete`, both reading
> `__step_error.message` so the recorded reason is the guard's own).
> **169 APPLIED 2026-07-19 09:31:37** (ledger row present; `snapshot_agent`
> taken 09:31:35; live config verified field-by-field against the migration).
> `update_component.error_step = refuse_mangled_write`, and both new steps exist
> with the intended config. So the refusal now routes to `needs_human_review`
> plus a NOTE — **as soon as the guard can fire**, which still needs the image.
>
> *What caught it:* writing 169 required dumping the live `tool-improver` step
> graph, which showed the null `error_step`.

**Guard committed** `cc7bcc881`; council revisions `f485eb8cb` (gate correlation
`e8827490-764a-4c90-b4db-72e358f9be87`). **Follow-up filed as `bugs_open/021`**:
the guard covers this one write path, while `page_components.rendered_html` and
`pages.rendered_header/footer/head` share the same unguarded overwrite shape —
raised by the council's bug-historian seat, which asked that a human confirm the
follow-up exists before the next incident lands there.

**Verified against the REAL artifacts of this incident** (not fixtures) —
exported from `component_versions` + `tmp_loot_truncated_20260718` and run
through the live guard:

| Write | Result |
|---|---|
| 10,280 → 1,253 (final wreck) | **BLOCKED** — 12% retained; `<style>` unterminated; ends `font-weight: bold;` |
| 10,280 → 6,771 (intermediate) | **BLOCKED** — passes the 50% size floor at 66%, caught by unterminated `<script>` + tail `'Epic` |
| 1,253 → 10,280 (the restore) | **ALLOWED** |

The intermediate case is the important one: the size check alone would have let
it through. Both fixes are committed but **NOT in the deployed chassis
(v1.0.1135)** — confirmed by pod-binary grep. Until an image ships, the exposure
is mitigated only by migration 168's raised ceilings.

## Fix candidates

**(a) A completeness guard on component writes — the real fix.** Refuse to
persist an `html_template` that (i) collapses in size versus the current row
beyond a sane threshold, or (ii) loses structure the current row had (had
`<script>`/`<div>` and the replacement does not), or (iii) fails a balanced-tag /
unterminated-`<script>` check. On refusal: leave the component untouched, fail
the work item honestly (`needs_human_review`), and write a NOTE saying the fix
was rejected as mangled — never a silent success. This is the "docs never fail
the work, and a fix never destroys the work" rule made mechanical.

**(b) Detect the truncation at its source.** `output_tokens == max_tokens` is a
reliable truncation signature. The LLM action already records both in
`llm_call_log`; it should flag the call as truncated and let the caller refuse
the result rather than parse it as if complete. Cheap, and it generalises to
every whole-artifact writer (the article-body arc, 2026-07-15, was the same
signature in a different agent).

**(c) Raising ceilings — done (168), necessary but never sufficient.** A bigger
budget makes truncation rarer; only (a)/(b) make persisting a wreck impossible.

## Verify after fixing

1. Force a truncation (temporarily set `improve_tool` `max_tokens` very low on a
   scratch tool) → the component must be UNCHANGED, the item must not read
   `complete`, and a note must record the refusal.
2. `SELECT length(html_template)` before/after any improve cycle on a real tool —
   never a collapse without an explicit, logged reason.

## References

- Travelling-docs `HANDOFF_2026-07-10_stage5_live_and_next_fronts.md` T26.
- `bugs_open/010` — the non-convergence finding from the same tool; this bug was
  found while verifying 010's fix.
- Article-body truncation arc (2026-07-15) — same `max_tokens` signature, different agent.
- Save path: `update_component_html`; config: `agent_definitions.default_config
  → workflow.steps.<step>.config.ai_service.max_tokens`.
