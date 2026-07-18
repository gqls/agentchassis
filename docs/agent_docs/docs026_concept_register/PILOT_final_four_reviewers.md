# PILOT — the final four council reviewers (seats #8-11; candidates #6-9)

**Status: ALL LIVE as of 2026-07-18**, applied sequentially via
`fixloop_eg_dartsonline/0NN_fix_proposer_v15..v18_*.sql` — each gated behind
the relevance filter, each surgical (chained `jsonb_set`), each pre-flight
checked, snapshotted, and verified. **This completes the council: 13 reviewers**
(edit-quality + guardian always-on; 11 gated specialists). One combined pilot
doc for the four, since they share the settled pattern — each seat's full
charter/prompt is in its migration file.

---

## Seat #8 — Compliance / legal eye (`review_compliance`, v15, candidate #6)

The severity-justified seat: two live incidents (fabricated veterinary prices,
legally recorded; fabricated marketing claims — including a poisoned writing
rule in a site spec that *instructed* the fabrication). Judges: unevidenced
user-facing claims/prices/testimonials (the evidence_base / audit-row
discipline, `CQ-017` smallest-true-claim); weakened claims gates/scanners; the
poisoned-spec class in prompt/spec edits; removed or obscured disclaimers and
legal pages (`LGL-001` — conspicuous-and-proximate, information-not-advice).
Footprint: pricing, evidence_base, legal/compliance/disclaimer terms, claims
machinery.

## Seat #9 — Render-pipeline guardian (`review_render_guardian`, v16, candidate #7)

Where most silent, visually-invisible bugs live. Judges: render paths that can
silently drop/blank content (fail-loud-not-silent, the escalate-not-blank
posture); the two rerender modes' skip semantics (`STY-048` — scoped mode
hash-skips, silently wrong for chrome changes; assemble mode re-embeds
unconditionally); the `data-runtime-fill` exemption and
rendered-artifacts-are-not-sources landmine (`STY-019`, vonc); the `var()`
colour inheritance chain (`CTS-011`); the three validation layers (`STY-004`).
Footprint: rerender/assemble/styling/page_components terms. Overlaps the
bug-historian on render fixes by design — recurrence vs. contracts are
different questions.

## Seat #10 — LLM-reliability specialist (`review_llm_reliability`, v17, candidate #8)

Born from two bugs found in one week. Judges: `ai_service` config placement
(`MDL-039`/BUG B — root SHADOWS step; step-level config is dead under a root
block, the old runbook rule was backwards); truncation-looks-like-success
(`MDL-038`/BUG A — `stop_reason` undecoded; the `output_tokens == max_tokens`
signature); thinking-spend and tokenizer-growth in budgets (the Sonnet-5
gotchas); model-swap discipline (snapshot/rollback + `llm_call_log`
verification, `MDL-005/006`). Footprint: aiservice, max_tokens, llm_call_log,
stop_reason, execute_llm_prompt terms.

## Seat #11 — Debugging / incident-lore historian (`review_debug_historian`, v18, candidate #9)

The deliberately-broad seat, carrying the register's largest category (74
lessons) — gated **loosely** (fires on most code fixes: `.go`, `platform/`,
`internal/`, `cmd/`, `.sql`), per the design doc's judgment call. Its question
is cheap and always worth asking: "we've been burned by this before — does the
plan account for it?" Judges the fix's *verification and surgery approach*:
needle-gate SQL discipline + the Postgres pitfall catalogue (`DBG-016` —
`replace()` silently no-ops, LIKE `%` literals, sticky aborted transactions,
counted needles never from memory); informational-column blast radius
(`DBG-017` — never scope by `sites.status='active'`); verify against the
running pod, never git/tags; the repair-vs-regenerate taxonomy for stored
templates (`DBG-065` — Mode B is unrepairable, route to regeneration).

---

## The completed council (execution order)

```
select_panel (deterministic relevance filter)
→ edit-quality                    [always-on]
→ bug-historian                   [gated: rebuild/render]
→ reuse-agent                     [gated: additions]
→ guidelines                      [gated: contracts/work-items]
→ tooling-provenance              [gated: contextkit/travelling docs]
→ adoption-guardian               [gated: adoption pipeline]
→ diagnosis-guardian              [gated: diagnosis machinery]
→ improvement-guardian            [gated: improvement loop]
→ compliance                      [gated: claims/pricing/legal]
→ render-guardian                 [gated: rendering/styling]
→ LLM-reliability                 [gated: LLM calls/config]
→ debug-historian                 [gated loosely: most code fixes]
→ guardian (hard veto, stability proviso)  [always-on]
→ council_decide (deterministic aggregation; skipped seats = abstentions)
```

Every specialist is advisory (`approve|object`, no veto); only the guardian
blocks. A typical fix wakes 2-5 seats, not all 13. Not yet exercised on a real
fix-proposer run — BUG A's fix dispatch remains the owner's call, and will be
the completed council's first real outing.
