# PLAN 2026-08-25 — the fleet prompt audit ("is this prompt encouraging AI styles of writing?")

**The instruction** (owner, 2026-08-25, via `loanzy_uk_example_site` — canonical record
`loanzy_uk_example_site/OWNER_REVIEW_2026-08-25_homegarden_and_what_it_says_about_every_site.md` §7):
*"They should also audit every prompt in the database and code and ask of it whether it is
contributing to good readable copy or whether it is encouraging AI styles of writing (bad)."*

**Prerequisite met:** the context refresh he ordered first is done —
`REFRESH_2026-08-25_deep_context_the_accumulated_copy_discussion.md`. The audit's criteria (§3
below) are derived from it, not from this session's taste.

**Scope note:** this is a workstream, not a task (the loanzy lane flagged the same reading). This
PLAN is the design; execution spans sessions. It lives in this lane's directory; split it out with
its own standing five only if it outgrows the lane.

---

## 1. The census — what "every prompt" is `[MEASURED 2026-08-25]`

All queries re-runnable; reproduced in full in the census below so no scratchpad survives to matter.
AF = `is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL`.

| population | where | count as of 2026-08-25 |
|---|---|---|
| active agent definitions | `agent_definitions` | 200 |
| prompt/instruction/guidance strings in `default_config` (recursive walk) | 73 agent types | **173 strings, 690,763 chars** |
| — of which `prompt_template` at any depth | `jsonb_path_query(default_config,'strict $.**.prompt_template')` | 137 (min 115 / median 3,219 / max 31,879 chars) |
| per-field writer instructions | `content_components.input_schema` `llm_guidance` | **2,442 strings, 265,269 chars** across 140 of 309 active components |
| live per-site briefs | `site_specs` `content_direction`, `is_current` | **31 rows** (largest aspect, 774,552 JSON chars incl. data) |
| voice-family aspects | `site_specs` | 16 rows — ⚠ most are DEAD surfaces (REFRESH §5); audit only what reaches a prompt |
| Go prompt-construction sites | `platform/ internal/ pkg/ cmd/`, non-test | **57 occurrences in 26 files**; writer-adjacent: `plan_sections_action.go`, `internal/agents/contentcreator/agent.go`, `ai_actions.go`, `html_actions.go`, `write_site_plan_action.go`, `validate_page_content.go` and kin |
| migrations carrying prompt text (audit trail, not a separate live population) | `sql_for_agents/` | 173 of 995 files |

Known census traps, inherited: `agent_definitions` has NO `prompt_template` column — everything is
inside `default_config` JSON; a steps-only walk misses `sub_workflow` (the page-content-writer's
14,897-char prompt — LANDMINES :7807); `task_workflow`/`orchestrator_workflow` columns carry 15
ILIKE-detected rows not yet JSON-walked (**unsized — phase 1 sizes them**); Go literal char volume
unmeasured (grep counts sites, not lengths).

## 2. The audit question, operationalised — six questions per prompt

From REFRESH §2–§6. The owner's single question splits into what the evidence says actually moves
the writing:

1. **What does this prompt DEMONSTRATE?** Count, in the prompt's OWN prose and its worked examples:
   the five gate shapes (`x_not_y`, `not_x_but_y`, `staccato`, `rather_than`, `negative_reveal` —
   `ScanDefineByNegation` runs on any text), plus the wider register the owner catches and lists
   don't: methodical scaffold, performed candour, presumption, both-direction word-weight. **This is
   the primary axis** — an instruction is also an example, and the classes track their
   demonstration counts (REFRESH §2).
2. **Does it talk ABOUT methodology/integrity in ways that can leak as content?** (REFRESH §4
   hypothesis — match instruction text against shipped self-description copy; homegarden's
   about.html and finetuning's honesty beat are the reference specimens.)
3. **Does it prescribe an opening/template that becomes a tic**, or **prohibit where it could
   exemplify the good move**? (Displacement evidence, REFRESH §2.)
4. **Do its exemplars obey v3's rules** (negative-frame, word-weight both directions, self-flagging
   commentary, manufactured contrast/cadence)? A concrete on-topic exemplar is also a LIFT risk —
   does it carry `how_to_use_these`-style guarding where lifting would hurt?
5. **Does it reach a writer at all?** Reachability per REFRESH §5 — findings against dead surfaces
   are filed as dead-surface findings (their own defect class), not as copy findings.
6. **Is any of its text load-bearing for a detector?** (prompt-text-poisons-its-own-detector;
   016b :12508) — a fix that rewrites prompt prose can silence or trigger detectors; name them.

## 3. Phases

- **Phase 0 — census.** DONE above, dated.
- **Phase 1 — mechanical pre-scan (cheap, complete).** Extract every string in the census to a
  working table; run `ScanDefineByNegation` + a scaffold/candour/presumption pattern pass over each;
  size the two unsized populations. Output: a league table — demonstrations per prompt ×
  reachability × call volume. No judgments yet; this ORDERS the work, it does not decide it
  (a pattern list is not the owner's ear — the scan finds teachers of the enumerable tells, the
  judgment pass finds the rest).
- **Phase 2 — judgment pass, highest-leverage first.** Read each prompt against the six questions.
  Priority from phase 1, seeded by what is already known: `page-content-writer`'s sub_workflow
  prompt (16 demonstrations/call, every section of every page) · `plan_sections_action.go`'s
  brief assembly · `content-gap-planner` (busiest planner by 27×) · the house voice row itself ·
  the 31 `content_direction.formatted` briefs · top `llm_guidance` components by render volume ·
  the about-page planning path (the premise layer) · audit/evaluator prompts whose language defines
  what gets flagged. Each prompt gets ONE verdict row: teaches-good / neutral / teaches-AI, with
  the specific demonstrating sentences quoted.
- **Phase 3 — fixes, separately and through the gates.** Not this PLAN's scope by owner ordering.
  Shape constraints already established: de-demonstrate rather than add prohibitions; replace bad
  exemplars with good ones in the same commit as any rule change (D1: "the exemplars change with
  the rule, or the change is theatre"); prompt changes to `platform/` go through the council;
  DB prompt changes ship as migrations (council scope since bugs_open/314); re-run the 305 gate's
  detectors after any prompt rewrite they read (question 6).
- **Acceptance throughout:** a sample the owner would pass — v3-as-proxy with judgment, never a
  regex count alone (REFRESH §9.5).

## 4. Records

Findings ledger: `AUDIT_prompts/` subdirectory of this lane as phase 2 begins — one file per
audited population, league table + verdict rows, every count dated. NOTES carries the running log;
README_where_we_are carries the plain-prose account; this PLAN is corrected in place, visibly, as
decisions land.
