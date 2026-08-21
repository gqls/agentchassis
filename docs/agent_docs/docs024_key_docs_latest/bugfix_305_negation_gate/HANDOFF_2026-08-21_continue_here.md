# HANDOFF 2026-08-21 — continue here (`bugfix_305_negation_gate`)

**Lane:** `bugfix_305_negation_gate`. **Bug:** `bugs_open/305` (owned by `copy_quality_two_stage`; this
lane built the platform half and contributed it back into that file, §8–§17).

> ## ▶ START HERE, IN THIS ORDER
> 1. `bugs_open/305` **§11 and §12 first** — what this fix deliberately does NOT do, and why the bug
>    file's own §7 verification instruction is wrong for it.
> 2. `SUMMARY_2026-08-20_the_rule_becomes_a_check.md` — five minutes, plain prose.
> 3. This file's "What is left" and "Decisions that are not a session's to make".
>
> **One-line state:** both halves are LIVE (migration `509` applied 2026-08-21 on chassis `v1.0.1321`,
> capability-probed on both replicas; `brief-negation-check` CronJob running daily since 08-20).
> Council **APPROVED** (`c48b7612`, round 4 of 4). The one open technical question is the per-page
> budget canary, which needs one page build to answer and cannot be answered before one.

## State, with how it was verified

| thing | state | how it was checked |
|---|---|---|
| the scanner + repair (`rewrite_negations`) | **LIVE** | `grep -ac 'rewrite_negations' /proc/1/exe` = 7 on BOTH replicas, control `rewrite_negationz` = 0 |
| the counting annotation | **LIVE, opt-in, on this agent only** | `copy_gate_annotate` = 1 in the binary; `true` at both step paths in `agent_definitions` |
| migration `509` | **APPLIED 2026-08-21 10:28Z** | chain reads `generate_content → rewrite_negations → render_section`, `page_budget` 2 |
| migration `517` | **APPLIED 2026-08-21 10:40Z** — the repair had NO `ai_service` and was live-but-blind on the first page | `copy_gate` marker went from `repair_unavailable` to needing its next run to confirm; both 509 and 517 recorded in the runner's ledger with `--record-only` |
| `brief-negation-check` | **LIVE daily 07:40 UTC**, `v1.0.1320` (rebuilt from HEAD), `imagePullPolicy: Always` | behavioural probe: reports "9 of 25" + a `regulatory (left alone)` column, which only the current code produces |
| council | **APPROVED**, `Council-Reviewed: c48b7612-3ecc-4345-912e-5966c079cb91` | 4 advisories at medium, none high; 11 of 14 seats approved outright |
| the three complained-of pages | **STILL SERVING IT** — 4 of 6 hero/CTA components | `page_components.content_data ~* '\w,\s+(not|never)\s+\w'` |
| the brief that supplies the tagline | **UNCHANGED** | `content_direction.formatted` still contains `in days, not months` |

## What is left

**0. CONFIRM THE REPAIR NOW WORKS — one query, and it is the first thing to do.** `517` fixed a
live-but-blind gate; nothing has exercised it since (no writer orchestration after 10:40Z at the time
of writing). The repair is proven the moment either of these is non-empty:
```sql
SELECT created_at, success, output_tokens, left(error_message,50) FROM llm_call_log
 WHERE error_message LIKE 'RETRY (bugs_open/305%' ORDER BY created_at DESC LIMIT 5;
SELECT collected_data->'copy_gate'->>'status', count(*) FROM orchestration_states
 WHERE collected_data ? 'copy_gate' GROUP BY 1;   -- 'repaired' is the goal; 'repair_unavailable' means still blind
```
⚠ If it still says `repair_unavailable` with a DIFFERENT error, read that error — it is a second cause,
not the same one.

**1. The per-page budget canary — still open, and the first run could not decide it.** ⚠ It cannot precede the apply
(the marker only exists once the step runs; my own precondition said otherwise and is corrected in
`509`'s header). Run it on the first page built after 2026-08-21:

```sql
SELECT jsonb_pretty(collected_data->'__copy_gate') FROM orchestration_states
 WHERE collected_data ? '__copy_gate' ORDER BY updated_at DESC LIMIT 5;
```
**MEASURED 2026-08-21 and it is worse than hoped, but the answer is still not decided.** `copy_gate`,
`copy_gate_0` and `copy_gate_1` all reach the durable row (the step's **output_field**, which
`saveStepResultWithRetry` copies); **`__copy_gate` does NOT** — that function
(`coordinator.go:1076`) loads a fresh state and copies only the step's own keys. The earlier
`agent_config` evidence was misleading: that key survives because it is written before an AWAIT, where
the full state is persisted.

⚠ **And the one available run could not discriminate:** iteration 0 had **0** hits, iteration 1 had 3,
so `page_hits: 3` is equally consistent with accumulating and with resetting. **It needs a page where
two sections both carry hits.** Until then: per-page budget UNPROVEN, live behaviour is the safe
fallback (per-section budget, every headline hit repaired regardless), page total counted at
`compile_page_sections` either way. **If you want it properly per-page, the mechanism is to carry the
count in the step's OUTPUT (which persists) and read the previous iteration's suffixed key
(`copy_gate_<N-1>`) — do not put it back in a bare `CollectedData` key.**

**2. `invented_superlative` is NOT in the running binary** (probed: 0). `v1.0.1321` predates commit
`1ac9b8890`. The **accuracy-claim family is covered** anyway — the fleet-wide banned-claim set already
catches *always accurate*, *definitive*, *guaranteed accurate*, *every claim is verified*, *never
wrong*. What is uncovered until the next roll is **hype**: industry-leading, best-in-class, 100%,
flawless. No action needed beyond the next roll; do not re-add it.

**3. The reuse follow-up this lane owes** (`reuse_agent`, approved round, medium): the truncation
three-state (`known-at-ceiling / known-below / UNKNOWN`) is inline in `rewrite_negations_action.go` and
wants to be a shared `aiservice.TruncationState(outTok, maxTok)`. **Whoever takes it should audit the
other call sites of `output_tokens >= max_tokens` at the same time rather than moving one.**

**4. The three pages, and the nine briefs — NOT this lane's, and not fixable by code.** See the
decisions below.

## Decisions that are not a session's to make

| # | decision | what turns on it |
|---|---|---|
| **D1** | **`RFC_044`: should the counting annotation run fleet-wide (default ON) again, or stay opt-in?** | Opt-in today, so outside `page-content-writer` "the copy improved" and "nothing was checking" are the same number. Fleet-wide costs ~10–60 µs per section and is a shared-contract change on two of the platform's most-invoked actions — which is why the council vetoed it arriving inside a bug fix. Nothing waits on this. |
| **D2** | **The nine briefs that hand the writer these phrases — corrected, or accepted?** | Each is a site-lane content decision. `ai-agent-orchestration.com`'s tagline is **MANDATED** onto four page types and is the exact sentence the owner objected to; the gate exempts it by design. ⚠ `remortgagecalculator.uk` is `portfolio_positioning`'s pilot and the fleet's worst brief — rerunning it before the brief is fixed would read as the gate failing. |
| **D3** | **Is `rather than` a tic or ordinary English?** | It is in **43%** of writer sections and currently counts toward the per-page budget of two. If it is ordinary English, drop it from the trip family and leave it as a density signal only. The rejection log after a week of traffic informs this; the call is taste, and taste is the owner's. |
| **D4** | **The post-deploy `negation_density` threshold, currently >12.** | Chosen to hold the `voice_tells` queue at its existing volume (14 of 189 pages). Lowering it surfaces more (>8 → 43 pages, >5 → 61, >3 → 87) into a queue with 45 parked items that has closed one, ever. Revisit when `bugs_open/033` gives that queue a working surface. |
| **D5** | **Do `brief_supplies_negation` findings need a route?** | Nine are open at `needs_human_review` with no handler — visible but unassigned, the same hole as `bugs_open/033`. |

## Can `bugs_open/305` close? — NO, and here is the bar it fails

The **defect** is fixed and live. The **damage** is not: four of the six hero/CTA components on the
three pages the owner named still serve the construction, and the brief still mandates the tagline
into four page types. That is the same shape as `bugs_open/327` ("the defect is fixed while the damage
is not"), and the estate's bar for `bugs_closed/` is *fixed AND live* for the thing the bug is about —
which here includes the owner's second instruction, "fix the affected pages".

**What would let it close:** `site_ai_agent_orchestration` edits that site's `content_direction`
(⚠ whole object, never a patch — `bugs_open/327`; and verify by label presence, never by diffing, because
`formatted` regenerates in random key order) and rerenders the three pages. Then the §15 canary should
come back with 0 repairable hits and the tagline gone rather than exempt.

**This LANE, as opposed to the bug, is done** once the budget canary reads. Its remaining items are
(2) and (3) above, both of which ride other people's work.

## Standing cautions for anyone touching this

- **A brief-supplied phrase passes the gate BY DESIGN.** If a page still reads "X, not Y", check the
  `__copy_gate` marker's `exempt` count before concluding the gate is broken (`LANDMINES.md`).
- **The durable `collected_data.generated_content` shows the copy BEFORE repair.** The splice reaches
  `render_section` in the same in-memory pass but the previous step's save is not rewritten.
  `page_components.content_data` is the truth.
- **Do not verify this fix on `llm_call_log` rates** — the gate's own repair calls are logged there, so
  the per-call rate can rise while the artefact improves (`bugs_open/305 §12`).
- **`\y`, never `\b`, in Postgres.** It cost this lane two wrong figures (`WRONG_CALLS.md`).
- **Run the §15 demand control after any change to the shapes** — recipe in `RUNBOOK` §7. It found a
  defect three test files had missed.

**Lane tooling:** none of its own — the scanner is `platform/orchestration/datahelpers/negationtells.go`
+ `negation_content.go`; the check is `cmd/brief-negation-check`. The scratch-only canary recipe is
`RUNBOOK` §7 (**`cmd/gatecanary` must never become a real command**).

**Migrations this lane owns:** `509` (applied 2026-08-21) with its `_ROLLBACK`.
