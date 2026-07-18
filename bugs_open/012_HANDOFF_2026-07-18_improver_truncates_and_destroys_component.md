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
