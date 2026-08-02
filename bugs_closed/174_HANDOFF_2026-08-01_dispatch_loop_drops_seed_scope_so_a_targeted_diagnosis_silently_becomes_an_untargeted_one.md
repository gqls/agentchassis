# 174 — the diagnosis dispatch loop drops `seed_scope` (and `runtime_page`), so a diagnosis you aimed at specific symbols silently runs against whatever the code search happened to return

**Filed 2026-08-01 by the `bugfix_164` lane.** OPEN, UNOWNED. **LIVE — it has already
degraded two other lanes' diagnoses**, see § Measured.

## How it was found

Running `bugs_open/164`'s own post-roll VERIFY, which says to induce the body cap by firing a
diagnosis with a `seed_scope` naming one very large file. I fired it, and the bundle came back
with **7 symbols, none oversized, `truncated=false`** — the cap never tripped. The scope I
asked for was not the scope that ran.

**The knob is not merely ignored — every observable says it worked.** `090_TRIGGER` accepts
`SEED_SCOPE`, documents it (`# SEED_SCOPE comma-separated path[:Symbol] entries for iteration
1`), writes it correctly into the work item spec, keys its own pre-dispatch coverage probe on
it (`:282-288`), and prints it back. The work item still holds it now. It simply never reaches
the agent.

## The defect

Two `call_agent` `input_mapping` blocks on the same path disagree, and the one in front wins.

**`diagnose-orchestrator` → `call_diagnoser` FORWARDS it** (live config):

```json
{"ref?":"input_data.ref", "repo?":"input_data.repo", "owner?":"input_data.owner",
 "symptom":"input_data.symptom", "site_id?":"input_data.site_id",
 "seed_scope?":"input_data.seed_scope",          ← present
 "runtime_page?":"input_data.runtime_page",      ← present
 "subject_key?":…, "runtime_site?":…, "subject_type?":…, "correlation_id?":…}
```

**`diagnose-dispatch-loop` → `call_handler` DOES NOT** (live config): a 10-key allow-list —
`ref? repo? owner? symptom site_id? subject_key? work_item_id runtime_site? subject_type?
correlation_id?`. **No `seed_scope`. No `runtime_page`.**

`input_mapping` is an **allow-list**, so an unlisted key is dropped in silence — the same
mechanism that dropped `fidelity` until migration 274 (see LANDMINES, *"`--fidelity high` is
not a milder `locked`"*). Diff of the two mappings, run live:

```
ORCHESTRATOR FORWARDS BUT LOOP DROPS: seed_scope
ORCHESTRATOR FORWARDS BUT LOOP DROPS: runtime_page
LOOP SENDS BUT ORCHESTRATOR IGNORES:  work_item_id   (expected — the loop's own bookkeeping)
```

**And the dispatch loop is the DEFAULT and live path.** `090` prints it itself:
*"`diagnose-pipeline-trigger` is enabled, so `diagnose-dispatch-loop` claims this item on its
next 60s tick"*. The only way to get a seed scope through today is `DISPATCH=1`, which forces a
direct publish to the orchestrator and bypasses the loop entirely.

## Why the failure is invisible rather than loud

`diagnose_assemble_bundle`'s scope resolution is a **fallback chain**, by design
(`diagnose_assemble_bundle_action.go:135-151`): loop scope → `input_data.seed_scope` →
`lookup_code_symbols`' `code_results`. With the seed dropped, step 2 finds nothing and step 3
quietly supplies a *different, plausible* scope. The action cannot tell "the caller gave no
seed" from "the seed was confiscated in transit", so it correctly does not complain. **A
fallback chain converts a lost parameter into a successful run with different inputs** — there
is no error, no warning, and the bundle looks entirely normal.

## Measured — live, and it has already cost two other lanes

Every intake that ever carried a non-empty `seed_scope` (`site_work_items`, `item_type =
'needs_diagnosis'`), with who claimed it:

| item_key | claimed_by | date | seed survived? |
|---|---|---|---|
| `scheduler-stamps-completed-at-publish` | `diagnose-dispatch-loop` | 2026-07-28 | **NO** |
| `deploy-image-asset-purpose-not-assetid` | `diagnose-dispatch-loop` | 2026-07-31 | **NO** |
| `verify164_bodycap` | `diagnose-dispatch-loop` | 2026-08-01 | **NO** |
| `verify164_bodycap_direct` | `090_TRIGGER_needs_diagnosis` (direct) | 2026-08-01 | **yes** |

**3 of 4, and the only survivor is a manual direct dispatch fired minutes ago to work around
this bug.** Corroborated from the other end:

```sql
SELECT count(*) FILTER (WHERE collected_data->'input_data' ? 'seed_scope') AS with_seed,
       count(*) AS retained
  FROM orchestration_states WHERE owner_agent_type='diagnose-agent';
```
→ **1 of 10.** That 1 is `d0fb8c27`, my direct dispatch.

**Two of the three losses are other lanes' real work**, not test traffic:
`deploy-image-asset-purpose-not-assetid` is the `bugs_open/155` lane (2026-07-31) and
`scheduler-stamps-completed-at-publish` is from 2026-07-28. Both completed. **Both were aimed
at chosen symbols, both silently ran against the code-search fallback instead, and neither
author has any way to know** — which also means any conclusion those runs reached about "the
code shows X" rests on a scope nobody chose.

⚠ **Window caveat:** `orchestration_states` retains barely a **day** here, so the `1 of 10` is
a snapshot, not a census. `site_work_items` retains longer, hence the 4-row table above
(2026-07-28 → 2026-08-01). Historic losses before 07-28 are `[UNMEASURED]` and unknowable.

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Add `seed_scope?` and `runtime_page?` to `diagnose-dispatch-loop.call_handler`'s
   `input_mapping`**, sourced from the claimed spec (`claimed.seed_scope`,
   `claimed.runtime_page`) — the other ten keys already use exactly that `claimed.*` form, and
   the spec demonstrably carries them. Config-only, no image. **Smallest correct fix, and it
   makes the two mappings agree**, which is the actual invariant.
2. **Make the two mappings lockstep-testable.** They are a `dedup-index/Go-list` shaped pair:
   two hand-maintained lists on one path that must not drift, with nothing checking. A
   source-or-DB test asserting *every key the orchestrator accepts is forwardable by the loop*
   (modulo a named allow-list of loop-internal keys like `work_item_id`) closes the class, not
   just today's two instances. Compare the `LOCK-007` construction (`bugs_closed/143`).
3. **Make the fallback say which branch it took.** `diagnose_assemble_bundle` could log/report
   *which* of the three scope sources supplied the scope, so "seed ignored" is visible in the
   artefact rather than indistinguishable from "no seed given". Weaker than 1 — it reports the
   symptom rather than removing it — but it is the part that would have caught this in a day
   instead of never, and it generalises to every fallback chain in the platform.
4. Document `DISPATCH=1` in `090` as the seed-scope workaround. **Do NOT do this instead of 1**
   — it writes the bug into the runbook as a feature.

## How to verify a fix

- Fire `090` **without** `DISPATCH=1` (i.e. let the loop claim it) with a `SEED_SCOPE`, then
  assert the agent actually received it:
  ```sql
  SELECT collected_data->'input_data'->'seed_scope' FROM orchestration_states
   WHERE owner_agent_type='diagnose-agent' AND correlation_id='<run corr>';
  ```
  Must be the array you passed, **not** absent. Use the **run** correlation stamped back onto
  the item as `spec.dispatch_correlation_id`, not the intake correlation — they differ.
- **Assert the effect, not just the field:** the bundle's `## In-scope code` must contain the
  symbols you named. Field-present is a weaker claim than scope-used, and the fallback chain is
  exactly what makes them come apart.
- **Negative control:** an intake with no `SEED_SCOPE` must still work and fall through to
  `code_results` as it does today.
- Re-run the mapping diff (§ The defect) and require it to report **no** orchestrator-forwards/
  loop-drops rows.

## Related

- **`bugs_open/164`** — this was found running its VERIFY, and it **blocks that verification via
  the default path**. 164's VERIFY has been amended to require `DISPATCH=1` until this is fixed.
- **LANDMINES**, *"`--fidelity high` is not a milder `locked`"* — same mechanism (`call_agent`
  `input_mapping` is an allow-list; `fidelity` was dropped until migration 274), different
  path. That entry's check — *"confirm the branch actually taken by reading `input_data` back
  off the orchestration row"* — is precisely what catches this one.
- `bugs_open/124` (double dispatch) is why the loop exists in front of the orchestrator at all,
  so the fix must not reintroduce a second publish.
- Family: a parameter accepted, stored, echoed and never applied — see also `bugs_open/136`
  (config says `domain` where the code reads `pipeline`) and MEMORY's *"a dead config key looks
  exactly like a live one"*.

---

# TAKEN AND FIXED — 2026-08-02, `bugfix_174_seed_scope_relay` lane

**Workstream docs:** `docs/agent_docs/docs024_key_docs_latest/bugfix_174_seed_scope_relay/`
**Council:** corr `081d98b3-75e1-4926-a17a-b0c72e5ccece` — **APPROVED, round 1**, 6 advisory objections, none high-severity.
**Commits:** `f51acb2bb` (Go + migration), `10789dfe6` (detector + 285→289 renumber).

## CORRECTION to this filing: fix candidate 1 was insufficient, and would have failed SILENTLY

> The ticket names one gate and proposes sourcing the new mapping keys from
> `claimed.seed_scope` / `claimed.runtime_page`. **Those paths do not exist.**
> `claim_item`'s SQL `RETURNING` clause is itself an allow-list — it projected
> nine spec keys, and neither of these was among them. Candidate 1 alone would
> have added an optional mapping entry resolving to nothing, which
> `ResolveInputMapping` skips at Info level: **a fix for a silent-drop bug that
> itself drops silently.**

**There are THREE gates in series, and all three had to move together:**

| # | gate | fixed by |
|---|---|---|
| 1 | `claim_item`'s `RETURNING` projection — never produced the keys | migration **289** (config, live) |
| 2 | `call_handler`'s `input_mapping` allow-list — the gate this ticket named | migration **289** (config, live) |
| 3 | **type.** `QueryDatabaseAction` stringifies every `[]byte` a column scan returns, so a jsonb value arrives as the *string* `["a","b"]`; `ExtractStringListHelper` returned nil for a string — indistinguishable from "nothing supplied" | `data_helpers.go` (Go, **awaiting roll**) |

Gate 3 means the config fix is **inert on its own**: without the Go half the
string still reads as nil and the code_results fallback runs exactly as before.
Applying 289 early therefore cannot half-enable anything.

## What shipped, against this ticket's own fix candidates

- **Candidate 1 — done, and corrected.** Migration `289_diagnose_dispatch_loop_forwards_seed_scope.sql`,
  applied 2026-08-02 ~11:15 (snapshot `f4055640`). Both allow-lists. Its final
  assertion checks the **invariant**, not the two keys: every key
  `diagnose-orchestrator`'s `input_contract` declares must be forwardable.
  (Numbered 285 at first; renumbered to **289** when another session claimed 285
  eight minutes later.)
- **Candidate 2 — done, but NOT as specified.** See below; the general form was
  measured and rejected.
- **Candidate 3 — done.** `diagnose_assemble_bundle` now records `scope_source`
  (`route` | `seed` | `code_results`) on its result every run, and renders a
  warning into the bundle **only** on the ambiguous `code_results` arm — so
  seed/route bundles stay byte-identical and no archived baseline moves.
- **Candidate 4 — correctly NOT done.** `DISPATCH=1` is not documented as a
  workaround.

## Candidate 2: the general check was measured and REJECTED — read this before rebuilding it

The ticket proposes asserting "every key the orchestrator accepts is forwardable
by the loop". Generalised to the fleet, that rule is **not sound**, measured live
2026-08-02:

| version | findings | why rejected |
|---|---|---|
| every `call_agent` forwards every key its callee declares | **31** of 75 resolvable call sites | Legitimate. `pageflow-builder.apply_site_design` omits `site_context`; `webdesign-agent` has `else_step: load_site_context` and loads it itself. |
| …and the callee actually READS `input_data.<key>` | **3** | Still cannot separate "the caller dropped it" from "the caller never had it". |
| **both** | — | **Blind to THIS bug.** `call_handler` resolves its callee at runtime (`agent_type_field: claimed.handler_agent`), so a static resolver skips the one site the check exists for. |

Shipped instead: **`config-key-audit --relay-gaps`** + `scripts/audit-relay-gaps.sh`
(concept register **WFA-007**) — a declared registry over the **3** dispatcher-shaped
relays that exist fleet-wide, where the question is answerable because the caller's
envelope *is* the work item spec. It reports findings, **unmatched registry entries**
(an assertion that stopped running — louder than a finding), and **uncovered**
dispatcher-shaped relays so the registry cannot silently fall behind.

Proven **by firing**: exit 1 against the pre-fix config rebuilt from 289's own
snapshot, naming both keys in both categories; exit 0 against live; exit 2 on an
empty export.

## Deliberately not fixed — and the blast-radius figure corrected

`QueryDatabaseAction`'s blanket jsonb stringification is the deeper cause and is
**not** fixed here.

> **CORRECTED: this lane's own council submission said "14 live query_database
> steps project json/jsonb". That was measured with a loose regex that also
> matched `->>` text casts and arrows inside WHERE predicates. The real figure is
> ONE — the projection this fix added.** Every other one of the 14 is a text cast
> or a predicate. Logged in `WRONG_CALLS.md`.

So the deferred fix has **zero currently-affected consumers**: a prospective trap,
not a live defect. Recorded in `LANDMINES.md` (two entries) at the council
`bug_historian` seat's insistence, which named it the minimum acceptable
mitigation.

## STATUS: OPEN until the chassis roll

Gates 1 and 2 are **live**. Gate 3 is **committed and inert** until an image
carrying `f51acb2bb` is rolled — so the defect is still reproducible through the
default path, and by the standing rule (fixed AND live) this stays in
`bugs_open/`.

**To close, verify at the artefact — not the tag, not git:**

1. Pod-grep **both** replicas for a string this change adds, plus a negative control:
   `kubectl exec -n ai-persona-system <pod> -- sh -c 'strings /app/agent-chassis | grep -c "This scope was NOT chosen"'` → 1
2. Fire `090` **without** `DISPATCH=1` (let the loop claim it) with a `SEED_SCOPE`.
3. Assert the field arrived **and** that it was used — they come apart, which is
   the whole bug:
   `SELECT collected_data->'assembled'->>'scope_source' FROM orchestration_states WHERE correlation_id='<run corr>';` → **`seed`**
   and the bundle's `## In-scope code` must contain the symbols you named.
4. Negative control: an intake with **no** seed must still work and still report
   `scope_source = code_results`.
5. `./scripts/audit-relay-gaps.sh` → 0 findings, 0 unmatched.

**One follow-up, not a blocker:** `--relay-gaps` is not yet wired into
`config-key-audit`'s `os.Args` dispatch — `main.go` was held by a concurrent
session (WFA-006) and committing it would have broken the build at HEAD. The
audit script refuses to report clean rather than pretending, so the gap is loud.

---

# CLOSED — 2026-08-02 evening — LIVE on chassis v1.0.1229 and PROVEN AT THE ARTEFACT

## 1. The binary, on both replicas (not the tag, not git)

| grep over `/app/agent-chassis` | `-g7fbt` | `-n8nbj` |
|---|---|---|
| `This scope was NOT chosen` (**added**) | 1 | 1 |
| `scope_source` (**added**) | 1 | 1 |
| `diagnose_assemble_bundle: scope resolved` (**added**) | 1 | 1 |
| `diagnose_assemble_bundle: no scope` (**positive control**, pre-existing) | 1 | 1 |
| `This scope was DEFINITELY not chosen` (**negative control**) | **0** | **0** |

The negative control is what makes the three 1s mean something: it proves the
grep discriminates rather than matching anything.

## 2. The behaviour, through the DEFAULT path — the path that was broken

Fired `090` with a `SEED_SCOPE` and **no `DISPATCH=1`**, so `diagnose-dispatch-loop`
claimed it. Intake `1a35f000-a95d-46cd-b4ea-8e61bff7bcea`, run
`12fdf121-04e8-431d-9245-38767971e9ea`.

**Iteration 1 — the seed arm:**
- `collected_data->'bundle'->>'scope_source'` = **`seed`**
- `symbol_count` = **2**
- The bundle artefact's `## In-scope code` contains **exactly** the two symbols
  named and nothing else:
  `data_helpers.go:ExtractStringListHelper` and
  `diagnose_assemble_bundle_action.go:DiagnoseAssembleBundleAction`
- The fallback warning is **absent**, correctly — the scope was chosen.

Both halves asserted, because they come apart: *field-present* (the seed
arrived) and *scope-used* (it governed the bundle) are different claims, and the
fallback chain is exactly what separates them.

**Iteration 2 — the loop-back arm, unasked-for confirmation:** `scope_source` =
**`route`**, 4 symbols, the verdicter's own revised scope. The seed did not pin
the loop to iteration 1. That is `TestScopeSource_LoopScopeOutranksSeed`
confirmed in production.

## 3. GATE 3 CONFIRMED IN PRODUCTION — the config half alone would NOT have fixed this

The live orchestration row for `diagnose-orchestrator`:

```
jsonb_typeof(collected_data->'input_data'->'seed_scope')  ->  string
```

**Not `array`.** The jsonb column travelled through `QueryDatabaseAction`'s
`[]byte`→`string` conversion and reached the agent as JSON *text*, exactly as the
three-gate analysis predicted. Before this fix `ExtractStringListHelper` returned
nil for that, so **migration 289 on its own would have dropped the seed a third
time, in a new place, just as silently.**

This was the one claim in this lane that could not be proven offline — pgx's
jsonb handling — and it is now measured rather than argued. It also vindicates
writing the helper to accept **both** `[]byte` and `string`: I did not have to be
right about the driver, and I was not sure.

An independent confirmation from the same query: the last pre-roll diagnosis
(`ae9404bd`, 09:25) has `scope_source` NULL, the post-roll one has it populated.

## 4. Not yet observed live, stated rather than implied

**The `code_results` negative control** — an intake with *no* seed must still work
and fall through to the code search — has **not** been observed post-roll, because
no unseeded diagnosis has run since v1.0.1229. It is covered offline by
`TestScopeSource_NoSeedFallsThroughToCodeResults` across 5 input shapes (absent,
`[]`, non-JSON text, a JSON object, an empty decoded list), and the next ordinary
unseeded diagnosis will confirm it:

```sql
SELECT collected_data->'bundle'->>'scope_source'
  FROM orchestration_states
 WHERE owner_agent_type='diagnose-agent'
   AND NOT (collected_data->'input_data' ? 'seed_scope')
 ORDER BY created_at DESC LIMIT 1;    -- expect 'code_results'
```

## 5. The class check

`./scripts/audit-relay-gaps.sh` → **0 findings, 0 unmatched, 2 uncovered**, exit 0.
The two uncovered relays are deliberate and must not be registered unread — see
the ticket's fix-candidate section and WFA-007's landmine.

**MOVED TO `bugs_closed/`.**
