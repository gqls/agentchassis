# HANDOFF 2026-08-21 — SUPERSEDED (`bugfix_305_negation_gate`)

> ## ⛔ DO NOT START HERE — read `HANDOFF_2026-08-22_continue_here.md` (same directory).
> This file is kept as HISTORY and its banners below are not all true any more. In particular
> **"THE GATE DOES NOT YET CHANGE PAGES" is FALSE as of 2026-08-22** — it does, proven at the
> artefact. The value of this file is the sequence: the fix was wrong twice and then right, and
> the wrong turns are written down where they happened.

**Lane:** `bugfix_305_negation_gate`. **Bug:** `bugs_open/305` (owned by `copy_quality_two_stage`; this
lane built the platform half and contributed it back into that file, §8–§17).

> ## ▶ START HERE, IN THIS ORDER
> 1. `bugs_open/305` **§11 and §12 first** — what this fix deliberately does NOT do, and why the bug
>    file's own §7 verification instruction is wrong for it.
> 2. `SUMMARY_2026-08-20_the_rule_becomes_a_check.md` — five minutes, plain prose.
> 3. This file's "What is left" and "Decisions that are not a session's to make".
>
> **One-line state:** both halves are LIVE and the repair is **PROVEN working on real copy**
> (migrations `509` + `517` applied 2026-08-21; chassis `v1.0.1321`, capability-probed on both
> replicas; `brief-negation-check` daily since 08-20). Council **APPROVED** (`c48b7612`, round 4 of 4).
> **UPDATED 2026-08-21 evening — the splice fix HAS ROLLED.** Chassis `v1.0.1322`, build
> `bac189921`, pods up 16:54Z. `git merge-base --is-ancestor 0eea9e597 bac189921` → **yes**; the
> superlative guard (`1ac9b8890`) is in too and was independently binary-probed on **both** replicas
> (`invented_superlative` = 1). ⚠ Control: HEAD (19:07 BST) is NOT an ancestor of the build, so the
> ancestry test discriminates — my first attempt used a control that predates the build and therefore
> could not fail.
>
> ## ⚠ UPDATED AGAIN, 2026-08-21 EVENING — THE GATE DOES NOT YET CHANGE PAGES
> The end-to-end confirmation happened and it FAILED. Two pages built after the roll reported
> `status: repaired, hits_after: 0`, both COMPLETED — and the stored `content_data` was
> **byte-identical to the pre-repair value**. The in-place mutation of the writer's content map does
> not survive the step boundary (`saveStepResultWithRetry` copies only the CURRENT step's own keys), so
> the renderer read the unpatched map. An honest marker over a page that never changed.
>
> **Fixed in code** (the action now returns the patched content as its own `result`) **and migration
> `548` points `render_section.content_from` at it — HELD, because the live build predates the change.**
> Full account: `bugs_open/305 §22`.
>
> **So today's state, precisely: the gate detects correctly, selects correctly, rewrites well, and does
> not yet change pages.** It needs one roll, then `548`, then one page to confirm at the artefact.
>
> ## ✅✅ 2026-08-22 10:00Z — ALL THREE STEPS ARE DONE. THE GATE CHANGES PAGES, PROVEN AT THE ARTEFACT.
> `loanzy.uk/tool-interest-rate-stress-test`, rebuilt 09:57Z post-548, saved 09:59Z: `copy_gate_2`
> **8 → 2 hits, 6 rewritten, 0 rejected**, and in `page_components.content_data` **0 of 6 removed
> constructions are present and 6 of 6 replacements are** (the second is the demand control). The
> pre-548 control on an accepted save inverts exactly (`RUNBOOK` §8). It also closes §20's same-field
> splice race at the artefact — six rewrites in ONE field, six landed. **Nothing below is outstanding
> for the defect half; what remains for the BUG is the damage half (the briefs and the three pages),
> which is not this lane's.** Details: bug file §23, NOTES 2026-08-22.
>
> ## ✅ UPDATED 2026-08-22 MORNING — STEPS 1 AND 2 ARE DONE; ONLY THE ARTEFACT PROOF REMAINS (superseded by the block above)
> The roll landed (`v1.0.1323`, stamp `70e7b4f9c`, pods up ~08:36Z; `dd9fc6197` is an ancestor,
> binary-probed on BOTH replicas with a discriminating absent control). **Migration `548` applied
> 2026-08-22 ~09:19Z** (UPDATE 1, verify block passed, recorded 09:20:25Z via `--record-only`;
> `_HOLD` dropped from both files) and the live row now reads `content_from = copy_gate.result`.
> Every `RewriteNegationsAction` exit path that precedes `render_section` was re-checked to carry
> `result` (clean pages included). ⚠ Markers COMPLETED before 09:20Z today (two loanzy.uk repairs,
> the overnight remortgagecalculator pair) still ran on the old wiring — honest about the map, wrong
> about the page. **What is left: the first post-09:20Z `repaired` page, verified at
> `page_components.content_data` on the part that DIFFERS.** Details: NOTES session 2026-08-22.

## State, with how it was verified

| thing | state | how it was checked |
|---|---|---|
| the scanner + repair (`rewrite_negations`) | **LIVE** | `grep -ac 'rewrite_negations' /proc/1/exe` = 7 on BOTH replicas, control `rewrite_negationz` = 0 |
| the counting annotation | **LIVE, opt-in, on this agent only** | `copy_gate_annotate` = 1 in the binary; `true` at both step paths in `agent_definitions` |
| migration `509` | **APPLIED 2026-08-21 10:28Z** | chain reads `generate_content → rewrite_negations → render_section`, `page_budget` 2 |
| migration `548` | **APPLIED 2026-08-22 ~09:19Z**, recorded 09:20:25Z | `UPDATE 1`, verify `DO` passed; live row re-read as `content_from = copy_gate.result` |
| the repaired copy reaching `render_section` + `compile_page` | **PROVEN 2026-08-22** | `loanzy.uk/tool-interest-rate-stress-test`: `section_output_2.rendered_html` and the final `page_html` both carry the rewrite and lack the removed clause |
| the repaired copy reaching `page_components` | **STILL UNPROVEN** — the one open item | that page's save was refused wholesale by `bugs_open/253`'s component floor; pre-548 control shows the old defect on an accepted save (`RUNBOOK` §8) |
| migration `517` | **APPLIED 2026-08-21 10:40Z** — the repair had NO `ai_service` and was live-but-blind on the first page | `copy_gate` marker went from `repair_unavailable` to needing its next run to confirm; both 509 and 517 recorded in the runner's ledger with `--record-only` |
| `brief-negation-check` | **LIVE daily 07:40 UTC**, `v1.0.1320` (rebuilt from HEAD), `imagePullPolicy: Always` | behavioural probe: reports "9 of 25" + a `regulatory (left alone)` column, which only the current code produces |
| council | **APPROVED**, `Council-Reviewed: c48b7612-3ecc-4345-912e-5966c079cb91` | 4 advisories at medium, none high; 11 of 14 seats approved outright |
| the three complained-of pages | **STILL SERVING IT** — 4 of 6 hero/CTA components | `page_components.content_data ~* '\w,\s+(not|never)\s+\w'` |
| the brief that supplies the tagline | **UNCHANGED** | `content_direction.formatted` still contains `in days, not months` |

## What is left

**THE ORDER, and it is now three steps rather than one:**
> **2026-08-22: steps 1 and 2 are DONE (details in the banner above and NOTES). Step 3 is the
> only open item, and it now has a one-query recipe with a measured control — `RUNBOOK` §8.
> The three-step list is kept below as the record of what was required.**
>
> Step 3, as it now stands: the gate is proven through `render_section`, `compile_page` and the
> writer's final `page_html` on a real post-548 page (`loanzy.uk/tool-interest-rate-stress-test`,
> `copy_gate_2` repaired 6→5, one surgical rewrite). What is NOT yet proven is **persistence to
> `page_components`**, because that page's save was refused wholesale by `bugs_open/253`'s
> component floor (`hero-tool` 12→5 classes) — not by the gate, and with two controls. Needed: one
> post-548 repaired page whose save is ACCEPTED. Run `RUNBOOK` §8 on it; the pre-548 control
> (`tool-loan-repayment-calculator`, save accepted, `stored_matches_PRE=true/POST=false`) is what
> it has to invert. ⚠ Do NOT fire a loanzy rerender to force it — those builds are `bugfix_311`'s
> live work.

1. ~~**A chassis roll**~~ **DONE 2026-08-22** — `v1.0.1323`, stamp `70e7b4f9c`, both replicas probed
   with an absent control; `dd9fc6197` is an ancestor and the control commit is not.
2. ~~**Apply migration `548`**~~ **DONE 2026-08-22 ~09:19Z**, recorded 09:20:25Z; live row reads
   `copy_gate.result`.
3. **Then confirm at the artefact**, on the part that DIFFERS:
   ```sql
   SELECT pc.slot_name, pc.content_data->>'subheadline'
     FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
    WHERE s.domain='<the site>' AND p.name='<the page>';
   ```
   The removed construction must be ABSENT and the rewrite PRESENT. Never test the rewrite's opening —
   `from` and `to` share it, and the marker truncates both at 160 chars.

**0. ~~CONFIRM THE REPAIR NOW WORKS~~ — ANSWERED 2026-08-21 11:06Z. IT DOES.** First real repair,
`mortgagecalculator.co.uk`/`scorecard-simulator`: 5 hits found, **1 left alone as regulatory**, 2
allowed by the budget, **2 rewritten, 0 rejected**, `hits_before 5 → hits_after 3`, one call at 447
output tokens. Both rewrites surgical — a "rather than" clause removed, claim and facts intact. Full
marker in `bugs_open/305 §19`.

**0b. STILL OWED: ARTEFACT-LEVEL PROOF — one page that trips the gate AND renders.** The marker is a
*status*; this estate trusts the artefact. That page did not render: it failed at the next step on the
pre-existing `bugs_open/260` type gate (`mechanism-flow`'s `steps[N].branches` arrives as prose where
the schema declares an array of objects). **NOT caused by the gate** — control: on the earlier 10:30
run the repair spliced nothing (`repair_unavailable`, no `ai_service` yet) and the identical failure
occurred, and `steps[1].branches` fails too, which the gate never touched. Reported to the 260 lane.
The query that closes this item:
```sql
SELECT collected_data->'input_data'->'site_record'->>'domain' AS domain,
       collected_data->'input_data'->'current_page'->>'name'  AS page, status,
       collected_data->'copy_gate'->>'hits_before', collected_data->'copy_gate'->>'hits_after'
  FROM orchestration_states
 WHERE collected_data ? 'copy_gate' AND (collected_data->'copy_gate'->>'status')='repaired'
   AND status <> 'FAILED' ORDER BY updated_at DESC;
```
Then read that page's `page_components.content_data` and confirm the rewritten sentence is what is
stored — **never `updated_at`**, which a rerender bumps without regenerating.

**1. ~~The per-page budget canary~~ — ANSWERED 2026-08-21: it is PER SECTION.** The first page whose
first two sections both carried hits settled it — iteration 0 `page_hits 1`, iteration 1 `page_hits 8`
(not 9) with a fresh budget of 2. `__copy_gate` does not survive `saveStepResultWithRetry`, which
copies only the step's own keys. Live behaviour is the documented safe fallback: per-section budget,
**every headline hit repaired regardless**, page total counted at `compile_page_sections`. To make it
truly per-page, carry the count in the step's OUTPUT (`copy_gate_<N-1>`) — never a bare state key.

**1b. ~~THE REPAIR UNDER-APPLIES~~ ~~FIXED AND ROLLED~~ — SUPERSEDED BY §22: THE REPAIR DID NOT REACH
THE PAGE AT ALL. See the banner above. What follows is the older, narrower item, kept because its
verification recipe is still the right one once `548` is applied.** ~~FIXED AND ROLLED (`0eea9e597`, live in `bac189921` since
2026-08-21 16:54Z). ONE END-TO-END CONFIRMATION IS STILL OWED, and it is traffic-bound.** Several hits
in ONE field used to race: six accepted rewrites, one landed, confirmed at the artefact. The unit test
composes two rewrites in one field and a mutation probe reproduces the race, so the fix is proven in
code; what is not yet proven is the same thing through the real renderer and save. **No page has been
built since the roll** — writer traffic ran 3–12 calls/hour from 10:00 to 15:00 and then stopped, with
zero queued `page_rerender`/`page_build` items. **Nothing is wrong; there is simply no work.** The
check, on the first multi-hit page after 16:54Z: a page with several hits in one field should show
`hits_after ≈ hits_before − len(rewritten)`, and the removed constructions should be absent from
`page_components.content_data`. ⚠ Verify on the part that DIFFERS (the removed construction), never on
the rewrite's opening — `from` and `to` share it, and the marker truncates both at 160 chars. ⚠ It cannot precede the apply
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

**2. ~~`invented_superlative` is NOT in the running binary~~ — CLOSED 2026-08-22: it IS, on
`v1.0.1326`** (probed 1 on both replicas, control `invented_superlativz` 0). The hype family is
covered. Original note kept below for the reasoning.
~~(probed: 0). `v1.0.1321` predates commit `1ac9b8890`.~~ The **accuracy-claim family is covered** anyway — the fleet-wide banned-claim set already
catches *always accurate*, *definitive*, *guaranteed accurate*, *every claim is verified*, *never
wrong*. What is uncovered until the next roll is **hype**: industry-leading, best-in-class, 100%,
flawless. No action needed beyond the next roll; do not re-add it.

**3. ~~The reuse follow-up this lane owes~~ — DONE 2026-08-22, council `a696e2a3` APPROVED round 1.**
Shipped as `aiservice.ClassifyTruncation` (register **MDL-043**), three mutation-proven tests, adopted
at one call site. ⚠ **The audit changed the shape:** the platform already detects truncation
STRUCTURALLY (`IsTruncated`, MDL-038) and `GenerateText` returns it, so the numeric helper ships
documented as a BACKSTOP for the two cases that signal cannot reach — do not let a future caller
prefer it. The audit's third site (`cmd/reasoningset`, two places) is **`bugs_open/366`**, filed
rather than patched because it changes which rows reach an eval corpus. Two corrections came out of
it: pass `__sent_max_tokens` (the APPLIED ceiling), not `options["max_tokens"]`; and this action does
NOT inherit MDL-042's escalate-on-truncation retry. Detail: NOTES 2026-08-22 evening.

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
