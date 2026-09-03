# HANDOFF 2026-09-03 (rev 2, evening) — bugfix 440: phases 1a+1b APPROVED and shipped; the producer census RESIZED what remains; owner decisions D1–D5 are RULED

Written for a session with none of this context. Every claim carries its check; cite symbols,
never line numbers. Rev 2 supersedes this morning's rev 1 in place (same filename, deliberately:
one canonical "continue here" per lane) — rev 1's five open decisions are now ruled, and the
producer census below is new and material.

## The bug in one paragraph

`spec.reason` on `page_rerender` items is TWO fields wearing one name: the gate's routing key AND
free human prose. The live gate (`check_rerender_mode`: five-value `==` disjunction on
`spec.reason`, `else_step: render_page`) silently ASSEMBLES anything it doesn't recognise, so a
routing key nobody understands completes green having changed nothing. Refusal is impossible
until the fields split (`spec.routing_reason` = vocabulary-only; `spec.reason` = annotation, free
forever). Split + refusal = RFC_062; evidence = `bugs_open/440_HANDOFF_2026-09-02_…`.

## OWNER RULINGS, 2026-09-03 — settled, do not re-open (full text: RFC_062 §Rulings)

| # | ruling |
|---|---|
| D1 | a refused item routes to **`needs_human_review`** (message names the bad key AND the vocabulary) — not a silent assemble, not a blunt orchestration failure |
| D2 | the **404 lane CO-SIGNS** the phase-3 gate migration (the declarations are theirs) |
| D3 | **YES to the write-door CHECK constraint** (NOT VALID first, validated after census) |
| D4 | **NO policing of `spec.reason`** — the annotation stays free prose forever |
| D5 | phase 1b's courtesy gate on the 404 lane **LIFTED** (acted on: 1b shipped same day) |

## State — approved, shipped, verified

| what | state | re-check |
|---|---|---|
| phase 1a (livespec foundation, REB-008) | **APPROVED r1 `55def842`; SHIPPED** | `platform/livespec/rerender_routing_key{,_test}.go`, commit `a3758c399`; in build `7bf1ff674021` by ANCESTRY. ⚠ never literal-probe zero-caller code — DCE strips it (LANDMINES 2026-09-02, verifier `dd777a91`) |
| phase 1b (creator stamps the key) | **APPROVED r1 `934327db`; committed `ec66ed12b`** (+ advisory hardening) | `create_rerender_items_action.go` + `create_rerender_items_routing_key_test.go`. Stamps **in lockstep** with `spec.reason` — never on merely-known — so RFC_062's flip is byte-neutral for REB-001's designed degrade (`image_landed` without a component stamps neither) |
| owner rulings | recorded | RFC_062 §Rulings, lane PLAN phase table, REB-008 |
| coordination | done | CONTRIBs in 404's and 384's NOTES (`5b5c669dd`); 404's r4 = approved |

## ⚠ THE FINDING THAT RESIZED THE LANE (phase 1b's council round, `bug_historian` [medium])

`[MEASURED 2026-09-03]` by `created_by` over live `page_rerender` items: **the creator phase 1b
fixed mints 1 of 3,172 reason-bearing items.** `completeness-discovery-agent` mints 1,882;
`generic` 388; `derive_card_asset` 313; `render_news_section` 275; `component-template-fixer` 94;
plus lane/migration producers. And **13 Go files write an in-vocabulary reason directly**, mostly
as raw `{"reason":"x"}` literals that never touch the vocabulary constants
(`render_news_section_html.go`, `refresh_evidence_base_action.go`, `render_directory_action.go`,
`reconcile_section_data_action.go`, `flag_page_image_rebuild_action.go`,
`store_generated_component_action.go`, `discovery_checks/check_misdirected_cta.go`,
`…/check_literal_markdown.go`, `…/check_contact_form_undeliverable.go`, …).

**Consequences (already written into RFC_062 and the bug file):** phase 2 is a **producer
conversion programme**, not just a migration-authoring rule; the transition clause is
**load-bearing**, not a drain-window nicety; and phase 3 gains a gate condition — **a census
showing zero reason-bearing items from unconverted producers** before any narrowing. Narrowing
early would route ~3,100 items to assemble: this bug's own shape, fleet-wide, inside its own fix.

## What REMAINS (this is the close-out list)

1. **Phase 2 — producer conversion.** Convert the 13 Go writers to stamp `routing_reason`
   alongside `reason` (prefer the `livespec.Reason*` constants over new raw literals), plus the
   agent/config producers; then the migration-authoring rule + `scripts/pattern-check.py`
   advisory for the raw-SQL door. ⚠ pattern-check.py is IN council scope (2026-08-24). Expect
   several small council rounds rather than one large one.
2. **Phase 3 — the flip** (RFC_062, co-signed per D2): gate reads the transition clause
   (`TransitionRerenderModeConditionClause()`), refusal branch routes to `needs_human_review`
   (D1) via `CheckRoutingKnownConditionClause()`, plus the CHECK constraint (D3). **Blockers:**
   (a) **confirm the conditional evaluator's behaviour on a MISSING key vs `''`** — inverting it
   refuses every legacy item (flagged in module header, RFC, both approved submissions);
   (b) the unconverted-producer census reads zero; (c) 404 co-sign obtained.
3. **Close 440** when refusal is **fixed AND live**: induce an unknown routing key → it lands in
   `needs_human_review`; annotation-only prose still assembles unwarned; census clean. Then
   `git mv` to `bugs_closed` naming BOTH paths on the commit, verify at HEAD.
4. Riders: read the landmine verifier's verdict (`dd777a91`); flip REB-008's status wording at
   phase 3 and lift its no-second-producer constraint only when RFC_062 lands.

## Craft notes for the next session (each cost a round or a correction here)

- **Submission accuracy is where this lane keeps bleeding — not design.** Three consecutive
  rounds (404's r3/r4, our 1a, our 1b) were gated or objected-to on what the submission SHOWED,
  never on what the code did. Attach the query/grep OUTPUT; list EVERY file the commit touches
  (our 1b omitted the register file and drew a medium for it); name the prior round's commit
  when an edit builds on shipped-but-unlisted symbols (two seats asked whether livespec existed).
- **On a shared tree, a test FAIL naming symbols you have never heard of is a neighbour's WIP —
  NO DATA, not a result.** `platform/orchestration/actions` broke under this lane three times
  today from other lanes' in-flight files.
- **Gate a wait on the artefact you need**: `go test ./pkg -run ZzzNoSuchTestZzz` compiles test
  files and runs nothing. `go build` skips `_test.go` (opens too early); `go vet` fails on any
  lint anywhere in the package (never opens). Both mistakes made here — WRONG_CALLS 2026-09-03.
- **Counting emissions**: exclude `fix_correlation_id IS NOT NULL` or you count council payloads
  quoting your own string (WRONG_CALLS 2026-09-02).
- `platform/livespec` package run FAILS at HEAD regardless of this lane
  (`TestNoNewMigrationFileReadersOutsideTheAllowList`, 405 lane's `ffa1707b3`). Theirs.

## Key artefacts

| what | where |
|---|---|
| bug file (producer census at the tail) | `bugs_open/440_HANDOFF_2026-09-02_a_routing_key_nobody_understands_completes_green.md` |
| design + rulings | `docs/agent_docs/docs024_key_docs_latest/architecture_review/RFC_062_routing_key_annotation_split.md` |
| lane docs | this directory — PLAN (phases, consumers) · NOTES (evidence, dispositions, missteps; newest at bottom) · RUNBOOK (census, emission-counting, inert-verification) · README (owner-facing prose) |
| code | `platform/livespec/rerender_routing_key{,_test}.go` · `platform/orchestration/actions/create_rerender_items_action.go` (`rerenderMode.RoutingKey`) + `…_routing_key_test.go` |
| commits | `ec2efc06e` `a3758c399` `0600eb6b3` `5b5c669dd` `544de50e0` `35de364dd` `624d3d2e8` `ec66ed12b` `8657c3cb4` |
| council | `55def842` (1a) · `934327db` (1b) — both APPROVED r1 |
