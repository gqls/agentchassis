# 337 — one section type reliably blows `generate_template`'s 16,000-token ceiling, so every site that plans it loses the page after burning three full generations

**Filed:** 2026-08-20 by the `bugfix_311_component_keys` lane, as the residual `311` named and
deliberately did not fix ("a cap decision, not a defect in the store"). **Status: OPEN, unowned.**
**Severity: low volume, 100% reproducible within it** — 3 items, 2 sites, 3 pages, and every
occurrence is the same section type.

**Not a duplicate of `bugs_open/205` or `bugs_open/257`, and the difference is the whole point.**
205 is a step whose budget nobody configured (runs at the hardcoded 2048); 257 is the class where a
configured budget never *arrives* at the provider. Here the budget is configured, it arrives, it is
respected — and it is simply **below what this task needs**. Different fix, different owner.

## The mechanism, in three lines

1. `component-creator`'s `generate_template` step is configured
   `{"model":"claude-sonnet-4-6","provider":"anthropic","max_tokens":16000}`
   (read from the live `agent_definitions` row 2026-08-20, not from a seed).
2. The LLM's template for one particular section type runs to **~47,000 characters**, hits the cap,
   and the transport correctly refuses a truncated body rather than storing a fragment
   (the `bugs_open/012` lesson, working as intended).
3. `needs_new_component` retries to `max_attempts=3`, each attempt paying a full 16,000-token
   generation, then parks `failed`. The page is left with no component and ships hollow.

## Evidence — every occurrence, and they are all ONE section type

`site_work_items.error` verbatim [MEASURED 2026-08-20 09:20Z]:

| site | page | error |
|---|---|---|
| loancalculator.co.uk | `tool-credit-roadmap` | `output_tokens=16000 reached the configured cap, 46637 chars recovered` |
| loanzy.uk | `tool-eligibility-checker` | `output_tokens=16000 reached the configured cap, 48553 chars recovered` |
| loanzy.uk | `tool-credit-health-check` | `output_tokens=16000 reached the configured cap, 47436 chars recovered` |

All three requested **`section_type = 'loans-credit-health-check'`**. All three reached
`attempt_count = 3`. Two different sites, three different page contexts, three consecutive
generations each — **nine cap-hits, zero successes, one section type.**

> **NARROWED the same day, 09:35Z, before anyone acts on it:** of the three pages, **two are the
> live loss** — `loancalculator.co.uk`/`tool-credit-roadmap` and `loanzy.uk`/`tool-credit-health-check`,
> both `status='active'`, both `build_status='needs_rebuild'`. The third,
> `loanzy.uk`/`tool-eligibility-checker`, is **`status='archived'`**, so its cap failure cost nine
> generations and no live page. Checked because a sibling repair in the same session spent a full
> build on a different archived page before the archived-page guard refused the stamp at the end.
> **The spend figure is unaffected; the "page ships hollow" claim applies to two pages, not three.** So this is not "sometimes
the LLM is verbose"; the ceiling is roughly a third of what this brief actually produces, every
time.

Adjacent cap failures on other agents, for scope only (do NOT fold them in — different steps,
different ceilings): `oufe.com` `improve_tool` (2026-08-11), and three `needs_diagnosis` items
(2026-08-14, 2026-08-19 ×2). Fleet total carrying a `stop_reason=max_tokens` error: **7 items.**

**Blast radius is small today and grows with the vertical:** 4 pages fleet-wide currently plan a
`credit-health`-shaped section. Both affected sites are loan/credit sites, which is the vertical the
portfolio buildout multiplies — so the next loans domain will plan it too and lose the same page.

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Make the writer's brief bound its own output, and prove the bound.** The ceiling is only
   wrong relative to a prompt that does not know it exists. A component brief that states a length
   budget (and a rich-tool decomposition rule: markup + a script the extractor separates, not one
   47k blob) closes the door for every future section type, not just this one. Requires reading
   `generate_template`'s prompt and measuring a real generation against it — the honest version of
   this candidate includes "and here is the length it actually came in at".
2. **Raise the step's cap to fit the measured need** (~24,000 would clear all three by ~50%).
   Cheapest, an owner threshold call, and it does not close the door — the next richer tool blows
   whatever number is chosen. Note the interaction: this multiplies the cost of a *failing* item's
   three attempts too.
3. **Handle truncation instead of refusing it** — a continuation call, or a retry at a higher
   ceiling on `stop_reason=max_tokens` specifically. Structural, cross-cutting (it would change the
   contract for every LLM step in the fleet), and therefore an `architecture_review` question
   rather than a bug patch. It is also the only candidate that helps the `improve_tool` and
   `needs_diagnosis` occurrences.
4. **Spend fewer attempts on a deterministic failure.** Three identical cap-hits is 48,000
   generated tokens thrown away per site. A cap refusal is not a transient — retrying it unchanged
   cannot succeed. Weakest as a fix (it saves money, repairs nothing) but it is the cheapest and it
   is honest about what a retry can and cannot do.

## How to verify a fix

Re-drive `loans-credit-health-check` on loanzy.uk (RUNBOOK recipe in
`docs/agent_docs/docs024_key_docs_latest/bugfix_311_component_keys/RUNBOOK_311_fix.md`) and assert:
the item completes with `attempt_count = 0`; a `content_components` row exists for the section with
`length(html_template) > 0` and containing `</section>`; and — the half that matters — the served
page gains real controls, `grep -c '<input'` on
`https://loanzy.uk/tools/credit-health-check/index.html` going above 0 from a pinned "before" of 0.
**Verify at the artefact, not at the item status:** a stored component is not a working calculator.

## Declared substitute for the `090` loop (owner ruling 2026-07-31)

No `090` run, because **no causal claim in this file is inferred.** The cap is read from the live
`agent_definitions` row; the failures are quoted verbatim from `site_work_items.error`, which names
the mechanism itself (`output_tokens=16000 reached the configured cap`); the census is a single
`ILIKE` over the same column and could have returned any number. What this file does NOT assert is
*why* this brief produces 47k chars, or that candidate 1 would fix it — those are open, and
candidate 3 in particular is architecture-scope and should not be taken as diagnosed.

## Related

- `bugs_open/311` — the parent. Its store-side collision is fixed and demand-proven; these two
  loanzy pages are the residual that fix cannot reach, because they die upstream of the store.
- `bugs_open/205` (a step with no configured cap, running at 2048; closed in substance) and
  `bugs_open/257` (the class: a configured budget that never arrives). Siblings, not duplicates.
- `bugs_open/012` — the truncation-and-config family. The refusal here is 012's lesson working:
  a cut completion is not stored as if it were whole.

---

## 2026-08-22 — taken; still valid; fix built (escalation seam + resize), council pending

Taken by a fresh session (lane docs:
`docs/agent_docs/docs024_key_docs_latest/bugfix_337_token_cap/` — PLAN, RUNBOOK, NOTES,
README). `who-owns` pointed at the 311 lane, but their handoff lists this as "filed from
this lane, unowned" and their live work is 311/345/351 — no competing fix in flight, and
no open work item targets this bug.

**Validity re-check [MEASURED 2026-08-22]:** cap still 16000 in the live row; the items
still parked `failed`/`attempt_count=3`; and a FOURTH occurrence this file predates —
loanzy.uk `needs_new_component:loans-credit-health-check_run1`, 08-18, same three-cut
signature. `llm_call_log` holds 9 generate_template truncations 08-15→08-19, recovered
46,441–48,817 chars each.

**Three facts this file did not have:**

1. **The cap was tight for the step's ORDINARY work, not just this section.**
   Successful calls, 14-day window: p95 13,633 of 16,000 (85%), max 15,374 (96%).
2. **The fleet headroom monitor flagged this step BEFORE and DURING this bug's life and
   nothing consumed the flag.** LCO-007 (`fleet-step-token-pressure`) doc_notes rows
   from 2026-08-18 onward: "T generate_template@16000 — n=229, p95 92.4%, peak 100.0%,
   truncated 9". Detection works; flag→action dispatch does not — that gap is a named
   residual here, not re-solved with a second monitor. (FIX-058's open question — does
   the near-miss threshold scale with the cap? "revisit when a 16000 seat first
   crosses" — has now had its trigger: this bug is that crossing.)
3. **The sibling whole-component writers were levelled and this step was missed:**
   `generate_tool_html`/`improve_tool` at 32000, `recreate_tool` at 64000
   (bugs_closed/012's table, migration 168 lineage) — generate_template sat at 16000.
   Per bugs_closed/067's sweep rule: generate_template is component-creator's only
   `execute_llm_prompt` step, so the levelling below is the whole sweep for this agent.

**Fix shipped (candidates 2+3 together, 3 in its narrow honest form):**

- **Framework seam (candidate 3, scoped):** `max_tokens_ceiling` — an opt-in
  `ai_service` key; on a typed `TruncatedError`, `execute_llm_prompt` retries ONCE at
  the ceiling, logging the cut call `success=false` with an `ESCALATED
  (bugs_open/337: …)` prefix; the escalated call's outcome flows through the existing
  machinery (tolerate/076 guard, transient ladder) verbatim. NOT a continuation call
  and NOT a fleet default — opt-in per owner ruling 2026-08-02 §2. Register
  **MDL-042**; code `platform/orchestration/actions/truncation_escalation.go` (+test,
  + wiring in `ai_actions.go`). Inert until the next chassis roll.
- **Resize from measurement (candidate 2):** migration
  `docs/agent_docs/sql_for_agents/549_generate_template_cap_resized_and_escalation_ceiling_armed.sql`
  — 16000→24000 routine (clears p95 by ~75% and this section's extrapolated ~19–22k
  need) + ceiling 32000 (clock-safe: step measures 92–127 tok/s, 600s non-streaming
  timeout; 40k+ risks trading truncation for a silent clock death). 415-pattern
  resolved-value asserts; 484-pattern gates; rollback sidecar. Live on apply.
- **Candidate 1 (bound the brief)** stays the named follow-up — FIX-059/migration-484
  precedent recorded in the lane PLAN; it changes what the step produces and needs a
  measured real generation to design. **Candidate 4** stays with the RSH-007/WII
  failure-ladder contract: a truncation AT the ceiling still burns the item's
  remaining attempts.

**Verification (per §How to verify, plus the RUNBOOK_311 corrections):** pre-flight
`pages.status` — `tool-eligibility-checker` is ARCHIVED, spend nothing on it; targets
are loanzy.uk `tool-credit-health-check` and loancalculator.co.uk `tool-credit-roadmap`.
Record WHICH half won: completion with zero `ESCALATED%` llm_call_log rows proves the
resize; with one, proves the escalation. Either passes; the attribution must be stated,
not assumed.
