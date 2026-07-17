# PILOT — "Adoption-Pipeline Guardian" council reviewer (stage 3, seat #5; candidate #3)

**Status: LIVE as of 2026-07-17.** Applied via
`fixloop_eg_dartsonline/0NN_fix_proposer_v12_adoption_guardian.sql`. The
**first seat built behind the relevance filter** (gated from the start, not
always-on) and the first built with a **surgical** migration (chained
`jsonb_set`, not a full-config reapply) to respect the now co-edited workflow.
Council is now **7 reviewers** (2 always-on + 5 gated specialists).

---

## 1. Why this seat

Candidate #3 from the roster. Grounded in `ADO-006` — "adoption writes specs
first, classifier consumes under fidelity rules" — which is one of the *two*
original stage-1-flagged rediscovered concepts (independently re-derived and
confirmed by five different units across two months). Plus `ADO-003` (the
adoption pipeline shape). The adoption pipeline has strict architectural
contracts that a fix could easily break, and the register shows people keep
having to relearn them.

## 2. Charter — the adoption pipeline's load-bearing contracts

**The adoption guardian judges one question: does this fix respect the adoption
pipeline's event-driven, write-then-relay contract, or break it?**

- **WRITE-THEN-RELAY** (`ADO-006`): `apply_adoption_plan` writes the specs
  (`site_archetype`, `design_reference`, `design_intent`, `content_direction`,
  identity), pages, and work items itself, then emits **exactly one** strategic
  item — `needs_domain_research`. It never calls the classifier/planner
  directly, and never emits build-stage items (`needs_composition`/`needs_design`).
- **ADOPTED SPECS ARE GROUND TRUTH**: the `domain-research-classifier` treats
  the adopted identity/archetype/content_direction/design_intent as ground
  truth that outranks its own search+scrape — reads-and-extends, never
  overwrites.
- **NO BYPASS**: adopted sites run the full strategist → briefing → planner
  chain exactly as fresh builds — adoption routes *through* the planner.
- **LLM FOR REASONING, GO FOR EXTRACTION** (`ADO-003`): never pay an LLM to
  transcribe what a regex/goquery can read (hex colours, fonts, CSS vars).

**Verdicts: `approve | object`, no `veto`** — same advisory design as the other
specialists.

## 3. Relevance footprint (gated)

Fires only when the fix touches the adoption pipeline:
```
"adoption": ["apply_adoption_plan","site-adoption","adoption","needs_domain_research","domain-research-classifier","site_archetype","design_intent","design_reference","content_direction"]
```
So on the large majority of fixes (which don't touch adoption) this seat
abstains — the filter doing its job. The prompt, full charter text, and exact
wiring are in the migration file.

## 4. Wiring (gated, surgical) — the pattern for every future seat

Because the workflow is co-edited (the guardian carries a `code_checks`
mechanism and the stability proviso, neither in the seat migrations), this seat
was added by **surgical `jsonb_set`** on the live config, NOT the v6-v11
full-config reapply. Eight operations in one atomic UPDATE: add the footprint,
add the reviewer step, add the gate, rewire `review_tooling_provenance.next_step`
and `gate_tooling_provenance.else_step` → `gate_adoption`, and append the seat
to `council_decide`/`escalate`/`run_checks`. Verified afterward that the
guardian proviso, `code_checks`, the other seats, and the filter wiring were
byte-intact. **This surgical pattern is now the standard for all further seat
additions** (see `DESIGN_relevance_filter.md` and the memory note).
