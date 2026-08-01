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
