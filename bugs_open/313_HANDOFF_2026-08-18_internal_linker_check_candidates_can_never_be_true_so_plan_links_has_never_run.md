# 313 — `internal-linker`'s `check_candidates` can never be true, so `plan_links` has never run: the agent has completed 57 link jobs and produced no link plan

**Filed 2026-08-18** by the `bugfix_275_silent_row_caps` lane, while answering `bugs_open/298`'s
reachability question. **Diagnosis loop: CONFIRMED, first iteration** (`RUN_CORRELATION_ID=
c4aa3559-86b1-4356-a28b-c71dfa661465`) — it re-read the same config and code independently and cited
the same lines, plus one link in the chain this file had not spelled out (see §The chain).

## The defect, in one line

`load_candidate_pages` declares `"output_format": "array"`, so the result has **no `count` key**;
`check_candidates` tests **`candidate_pages.count > 0`**; the path cannot resolve, the numeric
comparison returns false, and every run takes `else_step: complete_no_candidates`.

## The chain, end to end

| step | action | shape | what happens |
|---|---|---|---|
| `load_target_page` | `query_database` | **`object`** | has a `count` key |
| `check_target_found` | `conditional` | `target_page.page_id != null` | **works** — this is why runs get *past* the first gate |
| `load_candidate_pages` | `query_database` | **`array`** | bare slice, **no `count` key** |
| `check_candidates` | `conditional` | `candidate_pages.count > 0` | **cannot resolve → always false** |
| `complete_no_candidates` | `complete_workflow` | `"status": "skipped"` | where every run ends |
| `load_specs` → `plan_links` | the `then` chain | | **never entered** |

`check_candidates`' `then_step` is `load_specs`, and `load_specs` is the **only** step that chains to
`plan_links`. So taking the else branch is sufficient to make `plan_links` unreachable — that link was
the diagnosis loop's addition.

### Why the path cannot resolve (read, not inferred)

- `QueryDatabaseAction` returns the bare slice for `output_format: array`
  (`platform/orchestration/actions/database_actions.go:129`); `count` is set **only** in the `object`
  branch, alongside `rows` and `columns`.
- `resolveFieldValue` (`platform/orchestration/actions/conditional_branch_action.go:320`) tries five
  strategies. The only one that reaches a nested key on a non-map base is **Strategy 5**, and it is
  guarded by `if baseMap, ok := baseObj.(map[string]interface{}); ok` — an array fails that assertion.
  All five fail, and it returns `nil`.
- `evaluateSingleComparison`'s numeric arm then fails `datahelpers.ToFloat64(nil)`, logs
  `Numeric comparison: left side is not numeric`, and **`return false, nil`** — no error, no
  distinguishable failure, just the `else` branch.

**There is no input that makes this true.** It is not "fails when the site has few pages"; it is
unsatisfiable as configured.

## Evidence

**Two runs whose state survives** (`orchestration_states` is pruned to ~2 days):

| run | `candidate_pages` loaded | ended at |
|---|---|---|
| 2026-08-17 22:22:33Z | 7 | `complete_no_candidates` |
| 2026-08-18 01:01:46Z | **15** | `complete_no_candidates` |

The second is the important one: fifteen real candidate pages sat in `collected_data` while the step
after them reported there were none.

**[MEASURED] Volume:** 82 `site_work_items` name `internal-linker` as `handler_agent` —
**57 `needs_internal_links` marked `complete`**, spanning **2026-07-24 → 2026-08-18**, plus 20
`unresolved` and 5 `cancelled`.

**[MEASURED] Age: the shape dates to the agent's creation.** The live row was created **2026-04-12**,
and the seed that created it (`docs/agent_docs/sql_for_agents/101_internal_linker.sql`, added
2026-04-13) carries the **identical** pairing — `output_format: array` on `load_candidate_pages`,
`candidate_pages.count > 0` on `check_candidates`. Live config and seed agree, so there is no drift and
no later regression to find.

**[INFERRED, deterministically]** every run since creation took the else branch. The mechanism has no
data dependence and no randomness, so the two observed runs plus the unchanged config are sufficient —
but only two runs were *observed*, and that is the honest bound.

⚠ **`agent_definitions.updated_at` cannot date this.** The internal-linker row reads
`updated_at = 2026-08-18 17:59:19Z`, which looks like a targeted edit an hour before this was filed —
**200 rows share that minute**, so it is a bulk write. One `GROUP BY date_trunc('minute', updated_at)`
separates the two readings; without it, "someone is working on this agent right now" is the natural and
wrong conclusion.

## Why nobody noticed for four months

1. **`complete_no_candidates` declares `"status": "skipped"`, and the work items read `complete`.** A
   run that linked nothing is indistinguishable, from `status`, from one that linked successfully —
   the `a complete work item is not a repaired artefact` class again.
2. **The agent has zero `llm_call_log` rows**, which reads as "dormant, never dispatched" — the
   comfortable answer. It is actually a *symptom of this bug*: the only LLM step is `plan_links`, and
   `plan_links` is downstream of the dead branch. `bugs_open/298` recorded that zero honestly and drew
   the cautious conclusion; the zero meant the opposite of dormancy.
3. The failure is silent by construction: no error, no exception, no retry — one `return false`.

## What is NOT claimed

**This does not say sites have no internal links.** `resolve_internal_links` is a separate registered
action used elsewhere in the build path; whether pages get internal links by that route is a different
question and is not measured here. What is established is that **the dedicated internal-linking agent
has never produced a link plan.**

## Blast radius of the class — measured, and it is ONE

Censused across all live agent definitions: **7** conditionals test some `X.count`. Six read a producer
that really emits `count` (four `query_database` steps with `output_format: object`, plus
`load_unswept_areas` and `load_ch_enrichment_batch`, whose Go actions both set `"count": len(...)` —
`load_unswept_areas_action.go:145`, `companies_house_actions.go:122`). **`internal-linker` is the only
mismatch.** So this is an instance defect, not a fleet-wide class — which is worth stating because the
opposite assumption would point at a risky shared-mechanism change (see fix candidate 4).

## Fix candidates, ordered by what closes the door

1. **Make the producer and the consumer agree, in one migration** — `output_format: "object"` on
   `load_candidate_pages` **and** `{{range .candidate_pages.rows}}` in `plan_links`' prompt template.
   ⚠ **Both halves, or you trade a dead branch for a broken prompt:** the template currently does
   `{{range .candidate_pages}}` over the bare array, and ranging a map yields values in key order —
   `rows`, `count`, `columns` — which is not a page list. This is DB config, so it is live on apply with
   no roll, and it makes the pair match the working `load_target_page` / `check_target_found` pair in
   the same agent. **Fix `bugs_open/298`'s cap in the same migration** (see below).
2. **Add the config-time audit that would have caught it** — for every `X.count` condition, assert the
   step producing `X` emits a `count`. The census query is in this file's §Blast radius and in the
   `bugfix_275_silent_row_caps` RUNBOOK; it runs in one statement against live config and is a natural
   fit for `scripts/pattern-check.py` or `cmd/config-key-audit`. **This is the half that stops the next
   one**, and it is cheap because the shape is machine-checkable.
3. **Make the failure loud rather than silent** — `evaluateSingleComparison` already logs
   `Numeric comparison: left side is not numeric` at WARN and then returns false. On this fleet a WARN
   is unreadable within ~90 seconds (see `bugs_open/275` §2026-08-18 evening), so consider whether an
   unresolvable path in a *routing* condition should fail the step rather than silently pick `else`.
   That is a shared-seam change and belongs in architecture review, not in this bug's patch.
4. **Do NOT teach `resolveFieldValue` to synthesise `.count`/`.length` for arrays.** It looks like the
   root fix and it is the dangerous one: it changes what a shared conditional *guarantees*, and every
   currently-always-false condition of this shape would start evaluating true — behaviour change in
   agents nobody is looking at, from a change whose diff touches no agent. The census above says the
   affected set is exactly one, so the shared change buys nothing this bug needs.

## How to verify a fix

- **Disconfirming pair, at the artefact:** on a site with candidates, a run must produce a `plan_links`
  row in `llm_call_log` for `agent_type='internal-linker'` — a table that has **zero rows in all
  history** today, so the "before" arm is already established and cannot be faked.
- **And check the prompt actually rendered the pages** — `prompt_rendered` must list candidate page
  names under `## Candidate Pages`, not `rows`/`count`/`columns`. That is the fix-candidate-1 trap
  showing up as evidence rather than as an outage.
- `orchestration_states.current_step` for a completed run must no longer be `complete_no_candidates`
  when `collected_data->'candidate_pages'` is non-empty. That pairing **is** the bug, and it is
  readable from the durable channel without touching logs.

## Filing basis (owner ruling 2026-07-31)

Put through the diagnosis loop **before** being asserted here, because the cause is not where the
symptom is (the symptom was an empty `llm_call_log`, filed under a *row cap* ticket; the cause is a
routing condition two steps away). **Verdict: CONFIRMED**, first iteration, independently citing
`check_candidates`, `load_candidate_pages`, the surviving `orchestration_states` row, the
`resolveFieldValue` type assertion, and the `load_specs → plan_links` chain. Grepped `bugs_open/` and
`bugs_closed/` before filing: nothing covers this step or this mismatch.

## Related

`bugs_open/298` — the row cap on the very step whose result this branch discards. **Fix them together
and in this order:** repairing the cap alone changes nothing observable, and repairing the branch alone
immediately makes the cap live on the 8 of 24 sites that exceed 15 candidates (worst site 69).
`bugs_open/275` (the lane that found this; its §2026-08-18 evening entry has the durable-census method)
· register **LCO-009** · `bugs_open/287` (why some of this agent's work-item results are unreadable) ·
MEMORY `a-complete-work-item-is-not-a-repaired-artefact`.
