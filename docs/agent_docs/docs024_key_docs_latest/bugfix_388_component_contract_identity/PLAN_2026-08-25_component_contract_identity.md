# PLAN — `bugs_open/388`: the store honours the identity the advisory resolved, carried BY ROW ID

**Date:** 2026-08-25. **Status:** drafted, council submission pending.
Drafted by a `fable`-model agent against this lane's first-hand evidence, then corrected in two places
by measurement (§7). Corrections are marked, never edited away.

---

## 1. The mechanism, restated correctly

The bug as filed says two resolvers name different rows. True, but incomplete — and the file marked
the incomplete part `[INFERRED]` itself, which is why it was the first thing checked.

**The store's write target is decided by a value the LLM emits.** `parseGeneratedTemplate`
(`store_generated_component_action.go:798-806`) takes `data["function"]` when non-empty and falls back
to `NormaliseToKebab(section_type)` only when the model supplies none. That value feeds
`resolveStorageIdentity` (`:115`), which looks the row up by `function`
(`component_storage_identity.go:167-176`). The advisory
(`load_existing_component_action.go:170-179`) picks its row by `section_type`, ordered by derived
usage sites. Since `e1951c24b` (08-22) a sentence in the prompt tells the writer to echo the advised
function — **so the two are bridged, and the bridge is prompt text.**

Five paths make the advised row and the stored row differ:

| | path | state |
|---|---|---|
| **P1** | the pin is not rendered (no field names), so the LLM names its own function | **live** — the pin block is gated on `{{if .existing_component.field_names}}`, so an identity that resolved with an empty schema is never pinned |
| **P2** | the LLM emits nothing, and the `section_type` fallback names a different row | **live, measured** — 27 of 120 section_types diverge; 15 would produce a loud wrong-cause refusal, 12 a silent duplicate |
| **P3** | **`function` does not name a row at all** | **live, measured — see §2. This is the finding that decides the fix's shape.** |
| **P4** | the advisory's PRIMARY path never runs the foreign-dependent census, so for a shared row it advises a contract the store will divert away from | **live** — the file's own header names these two cases and cures them only on the FALLBACK path |
| **P5** | a resolved row with no schema fields keeps its identity and loses its pin | **structural, presently benign** — 5 of 154 active section rows, 4 of which happen to have `function == section_type` |

## 2. `function` IS NOT AN IDENTITY, and that is what settles the design

`lookupBaseComponent` (`component_storage_identity.go:167-176`) reads, in full:

```sql
SELECT id::text, COALESCE(html_template,''), COALESCE(input_schema::text,'{}'), js_content
FROM content_components
WHERE function = $1 AND forked_from IS NULL
ORDER BY is_active DESC, updated_at DESC
LIMIT 1
```

**It filters neither `component_level` nor `is_active`** (the missing `is_active` filter is deliberate,
2026-05-06, and documented in the function's own comment). So:

**`[MEASURED 2026-08-25]`** over 330 non-forked rows: **25 `function` values carry more than one row**,
the largest carrying **5**. The two largest span component levels:

| function | non-forked rows | active | component levels |
|---|---|---|---|
| `site-footer` | 5 | 2 | `section`, `site` |
| `site-header` | 5 | 2 | `section`, `site` |
| `head` | 2 | 0 | `head`, `section` |
| `header-docs` | 2 | 1 | `header`, `section` |
| `tool-agent-complexity-estimator` | 2 | 1 | `section`, `tool` |

So on 25 function values the store picks **one of several rows by recency**, and which one it picks can
change with no code change at all — a `LIMIT 1` over an `updated_at` ordering is a resolver whose answer
moves when anything else touches a sibling row. **A pin carried by function name is structurally
incapable of naming a row.** RFC_034 §1 already ruled the general case for this estate: *"Convert by
`content_components.id`, never by `function`."*

The disconfirming result was available and would have been decisive: had every `function` been unique
among non-forked rows, a name-carried pin would have been sufficient and the simpler fix would have won.
It is not unique, so it is not sufficient.

## 3. The fix

**Carry the advisory's resolved row id through the workflow, and make the store's write target that id
— with `bugs_open/311`'s diversion re-applied at write time on top of it.**

Offer and enforcement become one computation's result — the shape `bugs_open/337`/`282` already proved
for the source vocabulary, and `016b` §9/092's council-approved precedent (one predicate behind the
writer's allow-list and the gate's accept-set). The LLM keeps authority over the function *name it
writes into its template*; it loses authority over *which row gets overwritten*, which is not a
question a language model should ever have been answering.

Why each awkward case survives:

- **`bugs_open/311` composes**, because the diversion decision is re-run on the pinned row at store
  time. `resolveStorageIdentity` splits into lookup + decide; the pinned entry point enters the *same*
  decide half verbatim — unknown requester → legacy regen, foreign dependents → mint
  `<function>-<domain-slug>`, scoped lookup, double-collision refusal.
- **A genuine creation is unchanged.** No row resolved → no pin → the store behaves exactly as today.
- **Both fail-opens survive and stay independent.** The advisory still never errors and still always
  returns a well-formed map; `component_id` simply joins the empty one. A resolver problem degrades to
  today's un-pinned advice — blind, never blocking. A store that receives no pin runs the legacy path,
  so an `error_step` routing around the load step costs only the pin.
- **The prompt pin is repaired, not retired.** It is still the only defence on un-pinned paths. It gets
  re-gated on `{{if .existing_component.function}}` so P5 closes at the prompt as well as in code.
- **CLC-006's constraint is respected.** The register's stated reason for never building this —
  *"multiple components per section_type can be legitimate"* — is untouched: this changes **which** row
  a regeneration lands on, never **how many** rows a section_type may have. Nothing is refused, nothing
  is renamed, nothing is reconciled.

### RFC_022 / the 2026-08-02 opt-in ruling, argued rather than asserted

It ships as an optional config key on `store_generated_component`; absent, behaviour is byte-identical.
All three of RFC_022's conditions are enumerated, not claimed: **opt-in ✓**; **the unsafe side is the
default ✓** (absent = today's LLM-chosen identity); **live consumers enumerated ✓** —
`[MEASURED 2026-08-25]` exactly one live agent uses either action, `component-creator`, two steps,
queried against `agent_definitions`. Note also that this change **removes** authority from a shared
seam rather than adding any, so the 2026-08-02 §2 remedy is being applied to a change that arguably
does not trigger it. Optional-key budget: **5 → 6 of N=10** for this action, so no accumulated-surface
review is owed under WFA-013.

## 4. Detection — the residual made durable

Three codes through the existing `LogActionFindings` → `agent_error_log` seam, declared in
`finding_code_registry.json` **in the same commit** (the scan test and the daily
`finding-code-registry-check` CronJob turn red within a day of first fire).

⚠ **`[MEASURED 2026-08-25]` the `unruled` bucket is at its cap — exactly 25 of 25.** Parking is not
available; these ship with disposition `instrumented`, owning doc `bugs_open/388`, `review_by`
2026-09-25.

1. `COMPONENT_FUNCTION_PIN_DIVERGENCE` — pin present, LLM's emitted function ≠ advised function. **This
   is the durable obedience meter that the 11-observation sample could not be.** It fires harmlessly
   from now on, because the pin already decided the row — which is the point: the residual becomes
   countable instead of inferred.
2. `COMPONENT_ADVISED_ROW_VANISHED` — pin present, row gone or forked by store time; legacy fallback
   ran. Counts the advise→store race, which is otherwise invisible.
3. `COMPONENT_PARALLEL_SECTION_BIRTH` — creation path, no pin, not diverted, and an active non-forked
   section-level row already exists for that `section_type`. This is the silent-duplicate shape made
   loud. On the pinned path it is unrepresentable, so it can only fire on the fail-open residual.

⚠ **The 283 lane's warning is binding here:** a "did the identity resolve correctly?" check must not be
answered by the same query that did the resolving — `bugs_open/324` shipped 32 dangling rows because
its completeness check re-grepped the renamer's own patterns. Code 3 is deliberately a **different**
query (a census over `section_type`) from the one that resolved the identity (a lookup by `id`).

## 5. Edits (8, council-submission form)

1. **`component_storage_identity.go`** — split `resolveStorageIdentity` into `lookupBaseComponent` +
   `decideStorageIdentity` (lines 80-150 moved verbatim); add `lookupComponentByID`
   (`WHERE id = $1::uuid AND forked_from IS NULL`, also scanning `function`) and
   `resolveStorageIdentityByID`. `resolveStorageIdentity` becomes lookup + decide, behaviour-identical.
2. **`load_existing_component_action.go`** — primary query gains `id::text`; the found row is routed
   through `resolveStorageIdentityByID` (fixing P4); every return path gains `component_id`, including
   the no-schema-fields branch (closing P5) and `empty`. Header comment's "stays primary so nothing that
   works today changes" is superseded and must say so.
3. **`store_generated_component_action.go`** — `"advised_identity"` joins the Optional spec; when the
   pin carries a `component_id` the identity resolves by id, else today's call verbatim. Vanished pin →
   finding + legacy fallback. Emitted function ≠ advised function → finding. Creation path, un-pinned,
   not diverted → one `COUNT(*)` census → finding. Nothing downstream of `functionName =
   ident.FunctionName` changes. **The birth INSERT's column list is not touched, and `usage_count` is
   not reintroduced** (`bugs_open/378`'s migration 610 depends on that).
4. **`component_storage_identity_test.go`** — the four store-side tests (§6 T-A..T-D). This is the only
   file driving the store's create path end-to-end, so the new census query ripples nowhere else.
5. **`load_existing_component_action_test.go`** — `TestLoadExistingComponent_SectionTypeHitDoesNotConsultTheResolver`
   pins the *defect*; it is **rewritten with its header saying why**, not deleted, so a peer holding it
   as evidence finds the correction rather than an absence.
6. **`docs/agent_docs/sql_for_agents/611_component_creator_store_honours_advised_identity.sql`** —
   snapshot-first, anchored, `DO`/`RAISE`-guarded (a verify block of bare `SELECT`s cannot stop a
   `COMMIT`). (a) adds `"advised_identity?": "existing_component"` to the `store_component` step config;
   (b) re-gates the two pin sentences into their own `{{if .existing_component.function}}` block.
   **No ordering constraint is claimed, because none exists** (2026-07-29 ruling): an optional key the
   old binary's spec does not declare is inert, and the new binary without the wire simply runs
   un-pinned. Either order is safe.
7. **`finding_code_registry.json` + `optional_explicit_wire_acks.json`** — three codes at disposition
   `instrumented`; one ack row `store_generated_component.advised_identity` whose `downstream` states
   what was actually read: absent → `GetRaw` returns nil → the legacy branch runs verbatim.
8. **`docs026_concept_register/register/component-lifecycle.md`** — new CLC entry in the **same commit**
   as the Go (2026-07-28 ruling, condition 2), superseding CLC-006's "flagged, not built" line and
   recording the landmine: **a pin by function cannot name a row — 25 function values carry more than
   one non-forked row, the largest 5, spanning component levels, as of 2026-08-25.** Index count bumped.

## 6. Tests, each with the mutation that must turn it red

- **T-A** pin decides the write target: pinned id X, LLM emits a different function → expect by-id
  lookup and `UPDATE … WHERE id = X`; **no expectation exists for any `WHERE function =` lookup**.
  *Mutation:* revert to `resolveStorageIdentity(functionName…)` → sqlmock fails on the unexpected query.
- **T-B** vanished pin falls back and says so. *Mutation:* remove the fallback → no write path has
  expectations; remove the finding → unmet INSERT.
- **T-C** **the 311-composition proof**: pinned row with a foreign dependent still diverts → scoped
  lookup then INSERT, never an UPDATE of the pinned row. *Mutation:* set `IsRegeneration=true` on the
  pinned path without calling `decideStorageIdentity` → the incumbent UPDATE hits no expectation. This
  is a real negative proven by mutation, not by mock bookkeeping.
- **T-D** silent parallel birth is recorded. *Mutation:* delete the census block → unmet expectations.
- **T-E** (rewritten) primary hit routes through the identity decision. *Mutation:* return advice
  directly → census expectation unmet.
- **T-F** a foreign-depended primary row advises what the store will actually do (empty contract).
  *Mutation:* advise the incumbent's fields regardless → assertion fails.
- **T-G** a no-schema-fields row still pins its identity. *Mutation:* drop `component_id` from that
  branch → red.
- Migration 611 self-verifies with `DO` guards on anchor drift. Afterwards run
  `cmd/config-key-audit --optional-key-budget` and `--optional-explicit-wires` plus the parity tests.

## 7. CORRECTIONS to the drafted plan, both found by measurement

> **CORRECTED 2026-08-25 (a) — the `?` marker is a SUFFIX, not a prefix.** The draft wrote
> `"?advised_identity"`. `MarkedConfigKey` (`datahelpers/action_inputs.go:631-635`) is
> `strings.HasSuffix(key, "?")`, and the live config surface agrees — `related_pages?`, `component_id?`,
> `page_type?`, `replace_existing?`. Written the draft's way the key would have been read as a literal
> field named `?advised_identity`, matched nothing, and the pin would have been silently absent on every
> run — **a failure indistinguishable from the fix working**, since an absent pin is a legal state that
> falls back to today's behaviour. Caught by reading the parser rather than the convention's name.
>
> **CORRECTED 2026-08-25 (b) — the function-duplication census was understated.** The draft said "at
> least 12 function values carry 2-5 rows" and named `site-header ×5`, `site-footer ×5`. Measured over
> the population `lookupBaseComponent` actually sees (non-forked, **no** `component_level` filter, **no**
> `is_active` filter): **25** function values carry more than one row, of 330. The draft's own scoping
> instinct was right and its filter was too narrow; the real number strengthens the argument rather than
> weakening it, which is why it was worth re-measuring a number that already supported the conclusion.

## 8. Deliberately NOT done

- **The third resolver, `component_selector.go:queryCandidates`.** It answers a different question —
  which component goes on a *page*, scored per request context — and the file's own comment
  (`:219-224`) argues the orderings rightly differ. `bugs_open/107` warns that forcing "consistency"
  there switches on a preferential-attachment loop. **Separate lane; not touched under 388.**
- **No data reconciliation** (the bug's candidate 3). Once the pin ships, the 27 stop being drift and
  become ordinary naming diversity. The two measured duplicate pairs need a human merge decision — a
  work item, not a side effect of this fix.
- **No change to 311's diversion semantics**, and no backfill of migration 581's standing NULL rows.
- **No new CronJob.** The finding-code registry check and the daily audit fleet already read what the
  three codes write.

## 9. Risks, each with the measurement that settles it

1. **Advice goes quieter on foreign-depended primary rows** (P4's cure means divert-to-create advises
   nothing — matching enforcement, but a behaviour change). *Measure:* the share of `generate_template`
   calls rendering the regeneration block, baseline **7 of 37 since 08-22**, against the scoped-twin
   birth rate, which must not rise.
2. **Advise→store race.** *Measure:* `COMPONENT_ADVISED_ROW_VANISHED` count; expected ≈0. Sustained
   non-zero means the pin should carry a schema snapshot too.
3. **Wire bypass by the aggressive whole-tree search** — closed by the `?` marker. *Measure:* zero
   bypass rows for this field, and `--optional-explicit-wires` exiting 0.
4. **Go inert until the roll while 611 is live immediately.** Safe in both orders (§5.6). Verify
   post-roll at the build-provenance line with `git merge-base --is-ancestor`, **per service**.
5. **First live exercise.** After the roll and 611, regenerate ONE component whose section_type is among
   the 27 and confirm the UPDATE landed on the advised id, no new row was born, and the divergence
   finding recorded the LLM's counterfactual choice. One demand-driven run, not a clean sweep — a
   post-fix zero with no demand measures nothing.
