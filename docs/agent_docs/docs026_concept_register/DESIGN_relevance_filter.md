# DESIGN — relevance-filtering for the council reviewer panel

**Status (2026-07-17): the Go ENGINE is BUILT, tested, committed. The SQL
WIRING is specified below (§6/§7), ready to apply. The chassis DEPLOY is the
one remaining, owner-gated step — it is fleet-wide and shared with the
council-gate thread, so it is not applied unilaterally.**

Purpose: let the ~7 specialist seats (and, optionally, the 3 current advisory
seats) run **only when relevant to the specific fix**, instead of every seat
firing on every council decision. The "activation" half of stage 3's original
design (`PLAN_concept_register.md` §Stage 3).

**What's built (committed `37468ba65`):**
- `platform/orchestration/actions/select_review_panel_action.go` — a
  deterministic, config-driven matcher. Verified mechanics below.
- `diagnose_council_decide_action.go` — now treats an absent reviewer field as
  an abstention (not a hard error), failing closed only if ALL abstain.
- `select_review_panel_action_test.go` — footprint matching, corpus fallback,
  fail-open, `[]string` coercion. `go build` + `go test` green.

**Verified mechanics that shaped the build** (read from the live action code,
not assumed):
- `plan_persisted.files` is a `[]string` of edited file paths the persist step
  already extracts — the clean primary signal, no JSON parsing needed.
- The `conditional` action supports `field == true` / AND / OR, but its array
  `contains` is exact-membership only — it **cannot** pattern-match file paths.
  So the filter genuinely needs a compute step (`select_review_panel`) to turn
  paths into per-seat booleans that the conditionals then check.
- `diagnose_council_decide` hard-failed on any absent reviewer field, so
  skipping a seat needed the abstention change. That change is why this is a
  chassis-image build, not pure SQL.

**A second reason to build this as shared code, learned 2026-07-17:** there
are now TWO council definitions — fix-proposer's (live) and the council-gate's
clone (`0NN_council_gate.sql`, file-only). Adding a seat means hand-patching
both, which already drifted within a day (fix-proposer's `check_fields` omitted
the advisory seats; the gate's roster lagged two seats behind). Two
hand-maintained rosters that must stay identical is *itself* the drift-shaped
failure the guidelines and bug-historian seats exist to catch. If the panel
composition (which seats, in what order, with what checks) lived in one place
— a `select_review_panel` Go action both workflows call — the drift class
disappears. That is an argument for the Go-action approach (§4 option b) beyond
just the relevance filtering.

---

## 1. The problem it solves

Each reviewer is a sequential `execute_llm_prompt` step. At 5 seats, a council
decision is 5 LLM calls; the "ten more" candidate list would take it to ~14,
every one on the critical path before a decision, re-run each revise round.
Most specialist seats are irrelevant to most fixes — a CSS-rendering fix has
no business consulting an LLM-reliability specialist or an adoption-pipeline
guardian. Relevance-filtering keeps the always-on cost near the original two
(edit-quality + guardian) and pays the LLM cost of a specialist seat only when
the fix actually touches its domain.

This matters more, not less, because of the concurrent **council-gate** thread
(`fixloop_eg_dartsonline/DESIGN_feature_builder_and_council_gate.md`): that
service will run this same panel over *every* platform commit. A filter is
what keeps that affordable.

## 2. The relevance signal (what to match on)

The fix-proposer already has the two inputs a filter needs, both available at
the point the council runs:

- **The fix plan** (`plan_persisted.plan_json`): a list of `edits[]`, each with
  `file` (repo-relative path), `symbol`, and `operation`
  (`modify|add|remove|config_change`). This is the primary signal — *which
  files/areas does this fix touch, and is it adding new code?*
- **The diagnosis** (`diagnosis_row.conclusion`): names the tables, actions,
  and mechanisms the bug involves.

A seat "fires" for a given fix if any edit's `file`/`operation`, or any
entity the diagnosis cites, matches that seat's **relevance footprint**.

## 3. Each seat's relevance footprint

The footprints are not invented — they come straight from the `verify-later`
fields of each seat's grounding concepts in the register (the same fields that
name the files/tables to check). This is exactly the "match the fix plan's
touched files/tables against each category's verify-later footprint" design
already sketched in `PLAN_concept_register.md` §Stage 3.

| Seat | Fires when the fix touches… | Footprint patterns (from grounding concepts' verify-later) |
|---|---|---|
| edit-quality | *always* (original seat) | — universal — |
| guardian | *always* (original seat, hard veto) | — universal — |
| bug-historian | a rebuild/rerender/render path | `*rerender*`, `*render*`, `save_page_sections`, `sectionHasVisibleContent`, `call_agent.go`, `missingkey`, tables `page_components`/`content_components` |
| reuse-agent | anything with `operation: add` (new file/func/table/migration) | any edit `operation == 'add'`, or a `file` matching `*_action.go` newly created, or a new `.sql` migration |
| guidelines | work-items, contracts, agent defs, component schemas | `input_contract`/`output_contract`, `idx_swi_dedup`, `save_page_sections`, `site_work_items`, `agent_definitions`, `input_schema` |
| LLM-reliability (#8) | LLM call config / the AI client | `platform/aiservice/*`, `ai_actions.go`, `ai_service`, `max_tokens`, table `llm_call_log` |
| render-pipeline guardian (#7) | styling/assembly/render | `*rerender*`, `styling`, `css`, `assemble`, `page_components` |
| adoption-pipeline guardian (#3) | site adoption/onboarding | `*adoption*`, `needs_domain_research`, `site_specs`, `domain-research-classifier` |
| diagnosis-loop guardian (#4) | the diagnosis machinery itself | `diagnose_*`, `diagnosis_artifacts`, `orchestration_states` |
| improvement-loop guardian (#5) | the self-healing/audit sweep | `improvement*`, `discovery_check*`, `write_audit_findings`, `scheduled_tasks` |
| compliance/legal eye (#6) | user-facing claims/pricing/content | site content paths, `site_specs`, pricing/claims fields — a broader, content-scoped footprint |
| debugging/incident-lore (#9) | *broad* — any platform code fix | could be near-universal; a candidate to keep always-on OR gate loosely |
| contextkit/docs specialist (#10) | doc-tooling / contextkit | `contextkit`, `docs026`, doc-generation paths |

## 4. The mechanism — three pieces, engine now built

1. **`select_review_panel` step** (BUILT). Deterministic, no LLM. Inserted
   after `persist_plan`, before the reviewer chain. Reads `plan_persisted.files`
   (edited paths) plus any `extra_text_fields` (e.g. `diagnosis_row.conclusion`),
   matches them against a **config-driven** `footprints` map (seat → substring
   patterns), and writes `panel.run_<seat>` booleans. Config-driven on purpose:
   the roster/footprints live in the workflow SQL, so adding/rescoping a seat
   is a config edit, no Go rebuild — and the same action serves both councils.
   Fail-open: a seat with an empty/absent footprint runs, never silently drops.
2. **Per-seat conditional gates** (SQL, §7). Each filterable reviewer step is
   preceded by a `conditional`: `panel.run_<seat> == true` → run the reviewer;
   else → route straight to the following step. Cheap (no LLM). Verified the
   `conditional` action supports exactly this (`field == true`, then/else via
   `next_step_override`, same as `check_approved`).
3. **council_decide abstention** (BUILT). It now treats an absent reviewer
   field as an abstention rather than a hard error, so a skipped seat's missing
   `review_<seat>.result` no longer fails the run — while still failing closed
   if ALL reviewers abstain. This is the one thing that makes the filter a
   chassis-image change rather than pure SQL (the old code hard-failed on any
   absent field). Backward-compatible: nothing is absent until skips are wired.

## 5. Scope decision baked in: which seats are "always-on" vs filtered

- **Always-on (no gate):** edit-quality, guardian — the original two, universal
  by construction.
- **Filter candidates:** everything else, including the 3 advisory seats
  already live (bug-historian, reuse-agent, guidelines). Those 3 can be
  retrofitted behind the filter once it exists — e.g. bug-historian only firing
  on rebuild/render fixes, reuse-agent only on `add` operations — recovering
  most of their per-run cost. That retrofit is optional and can follow once the
  filter is proven on the new specialist seats.
- **A judgment call for the owner:** the debugging/incident-lore seat (#9) has
  a near-universal footprint (any platform fix). Either keep it always-on
  (accepting its cost) or give it a loose gate (fires on any `platform/` or
  `internal/` edit, which is most fixes anyway). Not decided here.

## 7. The SQL wiring (ready to apply AFTER the image ships)

The migration (a `v10`) does four things to `fix-proposer`'s workflow. It is
**not** written/applied yet because it must not land before the chassis image
carrying `select_review_panel` — a workflow referencing an action the binary
doesn't have fails at that step. Deploy order: **image → this SQL → fire.**

1. **Insert the panel step.** `persist_plan.next_step` → `select_panel`;
   `select_panel` (`action: select_review_panel`) `next_step` → `review_editquality`.
   Its config carries the footprint map (v1 retrofits the 3 current advisory
   seats as the proof-of-concept):
   ```json
   {
     "plan_field": "plan_persisted",
     "extra_text_fields": ["diagnosis_row.conclusion"],
     "footprints": {
       "bug_historian": ["rerender","render","save_page_sections","sectionhasvisiblecontent","call_agent.go","missingkey","page_components","content_components"],
       "reuse_agent":   ["_action.go",".sql","migration","create table","new "],
       "guidelines":    ["input_contract","output_contract","idx_swi_dedup","site_work_items","agent_definitions","input_schema","save_page_sections"]
     }
   }
   ```
2. **Gate each advisory seat.** Before `review_bug_historian`, insert
   `gate_bug_historian` (`action: conditional`, `condition: "panel.run_bug_historian == true"`,
   `then_step: review_bug_historian`, `else_step: gate_reuse_agent`). Same for
   `gate_reuse_agent` (else → `gate_guidelines`) and `gate_guidelines` (else →
   `review_guardian`). Point `review_editquality.next_step` → `gate_bug_historian`,
   and each reviewer's `next_step` → the *next gate* so run/skip reconverge.
   edit-quality and guardian stay always-on (no gate).
3. **council_decide / escalate / run_checks unchanged** — their `review_fields`
   still list all five; the abstention change means a skipped seat's absent
   field is now tolerated.
4. **Footprints, not code** — to add seat #N later, extend the `footprints`
   map + add one gate; no Go rebuild.

## 8. Scope decisions and the one gated step

- **Always-on:** edit-quality, guardian.
- **v1 filtered:** the 3 current advisory seats (retrofit — proves the
  mechanism on live seats before the 7 specialists exist).
- **Then:** build specialists #3–10 behind the filter from the start.
- **Judgment call:** the debugging/incident-lore seat (#9) has a near-universal
  footprint — keep always-on or gate loosely on `platform/`/`internal/`. TBD.

**The one remaining step is the DEPLOY.** Per `CLAUDE.md` (build inversion
2026-07-17): `make build-<service>` builds from committed `HEAD` via git-archive
— it structurally cannot bundle WIP, so the WIP-collision worry is gone; my Go
is committed (`37468ba65`), so a HEAD build includes it. The care that remains
is inherent to a chassis image, not the build mechanism:
- It is **fleet-wide** — rebuilds the binary every agent runs, and a HEAD build
  ships **all** committed code, including other threads' committed-but-untested
  work. Rolling it is a whole-fleet act, not a fix-proposer change.
- Deploy discipline (`CLAUDE.md` §Building & deploying): bump `IMAGE_TAG`, build
  from HEAD, roll, **verify against the running pod** (`strings /app/agent-chassis
  | grep -c SelectReviewPanelAction`), **image before seeds** (the §7 SQL names
  `select_review_panel` — it must not land before the binary has it), and no
  orchestration dispatch within ~300s of a chassis restart.
- The Go is **shared with the actively-developed council-gate/feature-builder
  thread**; ideally the same `select_review_panel` binary serves both councils,
  so the deploy is best **sequenced with that thread**, not shipped standalone.

Given it ships the whole fleet's binary carrying every thread's committed work,
this is a coordinated release decision, not a single-workstream apply — hence
flagged for the owner rather than rolled unilaterally. Everything up to it is
done and committed (`37468ba65` + the §7 spec).
