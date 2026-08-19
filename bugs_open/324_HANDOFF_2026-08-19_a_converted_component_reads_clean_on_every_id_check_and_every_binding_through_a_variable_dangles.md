# 324 — The instance-scope converter renamed ids and literal lookups; every binding that travels through a VARIABLE dangles. 32 of 69 converted templates are affected; 14 are serving live, structurally dead.

**Filed 2026-08-19 by the 283 lane. Status: OPEN.** Detector + mechanical repair + gate fix are
BUILT and tested (this session, commit pending); the repair itself rides the next chassis roll
through the framework. Parent: `bugs_open/283` (§14 records the correction to §13.7's
"COMPLETE and spot-checked"); RFC_034.

## The mechanism (read at the code, reproduced on live bytes)

`ConvertTemplateToInstanceScope` (`component_instance_conversion.go`) renames `id="x"`
declarations, literal `getElementById('x')` calls, id-reference attributes, `data-*` exact
values and `#x` selector references, then asserts completeness by grepping for those SAME
literal forms surviving. The premise — a binding that survives will *contain* the id literal at
the binding site — is false for three constructions the corpus uses constantly:

- **A. literal-through-variable**: `var ids = ['amount','interest','years'];
  ids.forEach(id => getElementById(id))` · `var fields = [{id:'gsfc-accel',…}];
  getElementById(field.id)` · `function el(id){…}; el('rw-ev')` — the declaration was renamed,
  the travelling literal was not; `getElementById` returns null; the first property access on
  it throws and the whole IIFE aborts.
- **B. declaration-by-concatenation**: `'<input id="name-' + index + '">'` — the declared-id
  regex captured the fragment as an "id", pass 1 prefixed the DECLARATION inside the JS string,
  and the paired lookup `getElementById('name-' + index)` and label `for="name-' + index`
  stayed bare.
- **C. lookup-by-concatenation**: `getElementById('step-' + n)` for declared `step-1…step-5`.

So a converted template reads clean on **every check the batch ran** — 0 unrendered tokens,
0 duplicate ids, prefixed markup, IIFE-scoped script — and the tool is dead the moment a user
interacts (or at IIFE start, where wiring runs at load).

## Measured, live, 2026-08-19 (Go detector `cmd/instanceaudit --bindings`, over all 69 converted rows)

| | count |
|---|---|
| converted templates | 69 |
| with ≥1 dangling binding | **32** |
| mechanically repairable (pass 5, see below) | **27** |
| need the judged pool after repair | **5** (3× `tool-automation-savings-estimator` — a `values['automation-level']` computed-key read that must match the now-prefixed write key; `tool-fuel-budget-forecaster` — composition hazard `'fg-' + id` where both `fleetSize` and `fg-fleetSize` are declared ids; `tool-loot-table-balancer` — dynamic id declarations with no static prefix) |
| **serving the broken bytes live** | **14 rows / 15 placements** — incl. webdesign.co.uk's own css-specificity demo, robot-hands.com gripper-safety (verified at the served page: bare `{ id: 'gsfc-accel' }` beside prefixed markup ids), oufe.com cramdown, gaswholesalers.com supplier-comparison, and `tool-affordability-complaint-checker` on THREE loan domains incl. loanandmortgagecalculator.co.uk |

The other 18 dirty rows sit on pages still serving pre-conversion HTML (loancalculator.co.uk's
'approved' pages, owned pages) — wrong in the template, not yet user-facing.

Severity per row varies with where the dangling lookup sits (load-time wiring = tool fully
dead; click-path = dead on interaction). `[INFERRED at runtime, structurally certain:
getElementById of an id absent from the rendered document returns null.]`

## Why every green check was green (the measurement lesson)

The batch's spot-checks measured what the converter changed (ids, tokens, duplicates) — the
question they encoded, not the question asked ("does the tool still work?"). The §13.7
"spot-checked at the served artefact" claims were TRUE as stated and blind to this class. The
one check that would have caught it on day one: grep the SCRIPT for any declared id surviving
bare — one line, now permanent (`UnprefixedBindings`, wired into `GateConvertedTemplate`, so
the gate itself refuses this shape from now on).

## 090 note (owner ruling 2026-07-31)

Not run through the diagnosis loop; equivalent first-hand verification substituted, stated
plainly: the mechanism was read at the converter's own completeness check; the census ran TWO
independent implementations (a python sweep and the Go detector, agreeing 32/32 with the Go
one also catching a composition hazard the python one missed); live damage verified at a served
page; and the repair is proven on pinned live bytes for all three classes with mutation
controls in both directions (`component_instance_bindings_test.go`).

## The fix (built this session; all through the framework)

1. **Pass 5 + detector** (`component_instance_bindings.go`): renames class A/B/C mechanically;
   refuse-contexts (comparisons, case labels, object keys, computed access) are skipped and
   REPORTED; `UnprefixedBindings` (incl. composition hazards) is now part of
   `GateConvertedTemplate` — the class cannot ship again silently.
2. **`repair_instance_scope_bindings` fix_type** (fifth writer in
   `fix_component_template_action.go`): load → repair → require detector-empty + two-instance
   gate clean + comparative write guard → snapshot (`change_source='repair_instance_scope_bindings'`)
   → write. Refusals write nothing and route to the judged pool.
3. **`cmd/instanceaudit --bindings`**: the census/verification mode (exit 3 while any converted
   row still dangles after repair — the done-check for the batch).
4. Repair seed (all 69 converted rows, derived at apply time; the arm no-ops on the 37 sound
   ones): `docs/agent_docs/sql_for_agents/487_seed_bindings_repair_HOLD.sql`, HOLD until the
   chassis roll carrying this code (verify RUNBOOK §1, then `git merge-base --is-ancestor
   <this commit> <stamp>`).

## How to verify the fix is done

- `go run ./cmd/instanceaudit <converted-export.json> --bindings` exits 0 with
  "still dangling after repair: 0" (the 5 judged rows leave via the judged pipeline).
- Per repaired serving page: the served script contains no bare declared-id literal; the tool
  computes (LMC oracle covers its 23; elsewhere one manual interaction per repaired page).

## Routes for the 5 judged rows

They join `bugs_open/283`'s judged pipeline (PLAN_2026-08-18_judged_pipeline.md; workflow
migration `486_judged_instance_scope_pipeline_HOLD.sql`). The 4 serving automation-savings
placements (ai-agent-orchestration.com, fundamentallyai.com, leopardessconsulting.co.uk,
finetuning.uk) are live-broken until then — if the judged pipeline slips, the contained
alternative is a snapshot rollback of those 3 rows + rerender, an owner call.
