# 313 — `internal-linker`'s `check_candidates` can never be true, so `plan_links` has never run: the agent has completed 57 link jobs and produced no link plan

> ## ✅ CLOSED 2026-08-19 — FIXED, LIVE **and PROVEN AT THE ARTEFACT**
>
> **The proof arrived 2026-08-19 21:19Z** (canary correlation `50ea3037-4602-40f5-b7de-a0b3a537ce39`,
> webdesign.co.uk, target page `about`, chassis v1.0.1316). All three of this file's §How to verify
> arms, each against its established "before" arm:
>
> 1. **`plan_links` RAN — the first time in the agent's history.** `llm_call_log` now holds exactly
>    one row with `step_name='plan_links'` in all history: 21:18:53Z, `success=t`,
>    `claude-haiku-4-5`, 21,946 input / 446 output tokens.
> 2. **The prompt rendered PAGES, not map keys** — `## Candidate Pages (could link TO the target)`
>    followed by `### domains (/domains/index.html)` with real titles and content. No
>    `rows`/`count`/`columns`. Fix candidate 1's stated trap did not fire.
> 3. **The run COMPLETED at `complete`, not `complete_no_candidates`**, with
>    `collected_data->'candidate_pages'->>'count' = 68`. That pairing — non-empty candidates AND a
>    terminal step that is not `complete_no_candidates` — **is** the bug, inverted.
> 4. **Re-verified on the NEXT build, because a dated "live on v1.0.NNNN" claim expires.**
>    Chassis rolled to **v1.0.1319** on 2026-08-20; the Go halves are still aboard —
>    `5315c8a19` is an ancestor of HEAD, and a binary probe of pod
>    `agent-chassis-86b95b967b-2fqm5` found `fail_on_non_numeric` **PRESENT**, with both
>    controls behaving (`else_step` PRESENT, a fabricated literal absent). The config half is
>    DB-side and unaffected by rolls.
> 5. Downstream, unasked but recorded: the model planned **2 links** and **2 `content_rewrite`
>    items** were filed (`created_by='internal-linker'`, keys `internal_link_webdesign.co.uk_index`
>    and `…_domains`). See §Residual below — it did not bite.
>
> > **⚠ CORRECTION 2026-08-19 — THIS FILE'S NOMINATED CHECK IS A BROKEN INSTRUMENT, and it stays
> > at zero even now the fix works.** §How to verify says "a `plan_links` row in `llm_call_log` for
> > `agent_type='internal-linker'` — a table that has zero rows in all history today, so the
> > 'before' arm is already established and cannot be faked". The row was written under
> > **`agent_type='generic'`**, so that query still returns ZERO after a demonstrably successful
> > run. `agent_type` on `llm_call_log` carries the DISPATCH context, not the workflow's agent —
> > the same class as the `orchestration_states.owner_agent_type` landmine this lane already
> > recorded. **The instrument that works is `step_name='plan_links'`**, which is unambiguous and
> > was genuinely zero all-history. [MEASURED for a hand-fired inline-workflow dispatch;
> > whether a loop-dispatched `spawn_agent` run labels it differently is UNMEASURED and does not
> > affect this proof.]
>
> > **⚠ CORRECTION 2026-08-19 — "the proof should arrive within hours" was WRONG, and waiting would
> > have failed silently.** This banner previously read *"Natural traffic is ~2 runs/day with 20
> > open work items queued, so the proof should arrive within hours."* **Those 20 items are
> > `status='unresolved'`, which is TERMINAL** (`work_items_common.go:40-46`) and invisible to the
> > promoter (`detected` only), the selector and the atomic claim
> > (`workItemDispatchableStatuses = {triaged, approved}`, `work_items_common.go:172-175`). They
> > can never dispatch — and this very bug is what parked them, via the two-strike rule
> > (`load_work_item_actions.go:1336+`) counting its own no-op "complete" runs as attempts. Fresh
> > items arrive only on a `site-discovery-rotation-completeness` tick (hourly, ONE site, 7-day
> > per-site stamp) that finds new orphans — days, not hours. **The proof had to be hand-fired**
> > (script and envelope in the lane RUNBOOK). Recorded in `WRONG_CALLS.md`.
> > *Self-healing note:* `unresolved` does not hold the `idx_swi_dedup` slot, so once those
> > terminals age past 7 days the rotation re-raises the findings as `detected` and they dispatch
> > against the fixed config. The 20 dead rows need no manual revival.
>
> **Residual (named in this file, now measured): it did not bite.** The 2-link plan produced **2**
> distinct items, not one — `bugs_open/321`'s disjoint-key fix is CLOSED and live, and
> `create_rewrite_item` suffixes by source page. Precisely: the key is
> `internal_link_<domain>_<source_page>`, so two links from *the same* source page would still
> collapse to one item. Not observed here, and not this bug's subject.
>
> ---
>
> ## ✅ FIXED 2026-08-19 — config LIVE on apply (the record as it stood before the proof)
>
> **Fixed by the `bugfix_313_internal_linker` lane** (session `bugs_open/313`; lane docs:
> `docs/agent_docs/docs024_key_docs_latest/bugfix_313_internal_linker/`). Together with
> `bugs_open/298`, in one migration, in this file's own ordering.
>
> - **Migration `490_internal_linker_candidates_object_uncapped_fail_loud.sql`** — applied 2026-08-19
>   (direct psql, clean `UPDATE 1`, recorded in `schema_migrations`; snapshot taken first). Fix
>   candidate 1, both halves: `load_candidate_pages.output_format` → `object` AND `plan_links`'
>   template → `{{range .candidate_pages.rows}}`. Post-apply read-back: `output_format=object`,
>   no row LIMIT, truncation marked, template on `.rows`, condition/routing untouched.
> - **Fix candidate 2 BUILT: `config-key-audit --array-producer-conditions`** (register **WFA-018**,
>   `scripts/audit-array-producer-conditions.sh`). Pre-fix fleet run: 191 agents, 145 conditionals,
>   exactly this one finding. **Post-fix run: internal-linker clean — and it caught a brand-new
>   SECOND instance seeded the same day** (`meta-description-backfiller.check_has_pages`, the 320
>   lane's agent; finding contributed to their NOTES, commit `ab02e8145`). §Blast radius said the
>   affected set was one and was right at filing time; the class detector is why the next one cost
>   hours, not months.
> - **Fix candidate 3 SHIPPED in the sanctioned opt-in shape** (register **WFA-019**): `conditional`
>   gains `fail_on_non_numeric` (default OFF, lenient behaviour frozen by test, mutation-proven);
>   `check_candidates` opts in via the same migration as a re-drift tripwire. Go halves ride the
>   next chassis build after commit `5315c8a19` (v1.0.1315 predates it). Fix candidate 4
>   (resolveFieldValue synthesis) deliberately NOT done, per this file's own warning.
> - **Council: APPROVED, round 2, all reviewers** (corr `aef24a7f-2992-4d4f-a6e0-422cd77fcca3`,
>   verdict read 2026-08-19). Round 1 REVISE caught two real things — recorded in the lane NOTES
>   and WRONG_CALLS. **The second instance is already FIXED by its owner** (the 320 lane repaired
>   the backfiller and adopted the tripwire the same afternoon), so the fleet audit now reads
>   **CLEAN: 193 agents, 147 conditionals, 0 findings**.
> - **Residual, named:** `create_rewrite_item` files items under `item_key_prefix: internal_link`
>   with no suffix field, so until the `bugs_open/321` lane's disjoint config fix applies, an
>   N-link plan yields ONE `content_rewrite` item (up to ~2/3 of output dropped). Confirmed
>   disjoint both sides; do not misread 1-item-per-plan as this fix failing.
> - ~~**Still OPEN pending this file's own §How to verify**~~ — **SATISFIED 2026-08-19 21:19Z, see
>   the CLOSED banner at the top of this file.** (The two claims struck through here — that the
>   `agent_type='internal-linker'` query is the check, and that natural traffic would supply it
>   within hours — are both corrected up there; neither held.)

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

> **2026-08-21 (post-close re-verification, four-bug check session):** the Go halves
> (`5315c8a19`) remain ancestors of every chassis stamp rolled since the close —
> `a255551e0` (v1.0.1320), `0483e7f4e` (v1.0.1321), `bac189921…` (v1.0.1322, current) — each
> read from `/proc/1/exe` with a positive control. The config half is DB-side and roll-proof.
> Nothing owed; noted because a "live on v1.0.NNNN" claim expires with the tag it names.
