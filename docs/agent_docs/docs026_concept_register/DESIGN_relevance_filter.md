# DESIGN — relevance-filtering for the council reviewer panel

**Status: DESIGN ONLY, not built.** Written 2026-07-17 after seat #3
(guidelines) brought the council to 5 always-on sequential reviewers. The
purpose: let the remaining ~7 specialist seats (and, eventually, the 3
current advisory seats) run **only when relevant to the specific fix**,
instead of every seat firing on every council decision. This is the "activation"
half of stage 3's original design (`PLAN_concept_register.md` §Stage 3), now
that there are enough seats to make it worth building.

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

## 4. The mechanism — and the one thing that makes it NOT pure-SQL

The seat additions so far have all been pure DB migrations (no chassis image).
Relevance-filtering **cannot** be, and here is exactly why — a concrete
finding, not a guess:

`diagnose_council_decide_action.go` (the deterministic aggregator) loops over
its configured `review_fields` and, for any field whose reviewer output is
absent, **hard-fails**:

```go
raw := datahelpers.ExtractNestedField(params.CollectedData, field)
if raw == nil {
    return nil, fmt.Errorf("reviewer output missing at %q", field)   // line 100-102
}
```

So if a seat is *skipped* (didn't run because it wasn't relevant), its
`review_<seat>.result` field never gets written, and `council_decide` fails
the whole run. A relevance-filter that skips steps therefore requires a small
Go change: **treat an absent review field as an abstention (skip it), not an
error.** That is a ~2-line change (`if raw == nil { continue }` plus a logged
count of how many seats abstained), but it means a chassis image build +
deploy sequencing — a different class of change from the SQL-only seat adds,
and the reason this is a design doc awaiting a build decision rather than
another same-day apply.

### The three pieces, in order

1. **A `select_panel` step** (deterministic, no LLM), inserted right after
   `persist_plan`, before the reviewer chain. It reads `plan_persisted.plan_json`
   and `diagnosis_row.conclusion`, matches edits/cited-entities against each
   filterable seat's footprint, and writes a boolean per seat into
   collected_data (e.g. `panel.run_llm_reliability = true`). Two ways to build
   this: (a) a `query_database` step doing jsonb extraction + SQL regex
   matching against a footprint table — pure DB, but clunky and the patterns
   live in SQL; (b) a small Go action `select_review_panel` with the footprint
   map in code — cleaner, testable, but part of the same image build the
   council_decide change already forces. **Recommend (b)** since the image
   build is already required.
2. **Per-seat conditional gates.** Each *filterable* reviewer step is preceded
   by a `conditional` action (the type already used by `check_approved` etc.):
   `panel.run_<seat> == true` → run the reviewer; else → `next_step` straight
   to the following step, skipping it. Cheap (no LLM). The always-on seats
   (edit-quality, guardian) get no gate.
3. **The council_decide abstention change** (§4 above) so a skipped seat's
   absent field is tolerated. Without this, steps 1–2 break the aggregator.

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

## 6. Recommendation and the decision this needs

Build order, when green-lit:
1. The `select_review_panel` Go action + the `council_decide` abstention change,
   together in one chassis image (they're interdependent).
2. Wire `select_panel` + per-seat conditionals into the workflow (SQL migration,
   applied after the image ships — same deploy-sequencing discipline the
   fix-proposer file already documents).
3. Build the remaining specialist seats (#3–10) behind the filter from the
   start, rather than as always-on steps.
4. Optionally retrofit the 3 current advisory seats behind the filter.

**The decision for the owner:** this is a chassis-image change to a
production, actively-developed workflow shared with the fix-loop and the
council-gate threads — bigger than the SQL-only seat adds. It should be
sequenced with those threads (the council-gate especially, since it will
consume the filtered panel). Recommend: confirm the approach here, then build
step 1 (the Go action + council_decide change) as its own reviewable change,
rather than folding it into a seat migration. Not started.
