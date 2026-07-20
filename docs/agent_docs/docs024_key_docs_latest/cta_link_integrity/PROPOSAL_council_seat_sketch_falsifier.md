# Proposal — a council seat that falsifies sketches against the actual source

**From:** owner, via the cta_link_integrity trail (2525f980) · **Date:** 2026-07-20
**For:** the council-gate / concept-register workstream (seat creation or amendment)

## The paragraph to hand over

> **Problem.** The council gate reviews *plans* — edit sketches plus rationale — and every
> current seat judges them by a different lens: plausibility against the diagnosis
> (edit-quality), pattern history (bug-historian), process discipline (debug-historian),
> reuse, scope, provenance. **No seat's mandate is to open the file being patched and check
> that the sketched code would actually do anything.** On trail 2525f980 a sketched
> observe-log hinged on `prev, ok := resolvedData[fieldName]` finding a previously-written
> value — but `resolvedData` is a fresh local map in `planSection` and the per-field loop
> writes each key at most once, so the condition is statically unreachable: the log was
> **dead code**. It survived **three review rounds and ten-plus seats**, several of which
> praised the staging built on top of it. Had it shipped, the observe round would have
> reported a structural zero and the council itself would have approved the follow-up write
> on false evidence — the exact failure the staged design existed to prevent. It was caught
> only when a concurrent session read the target function (`b6e374fc2`, doc_notes
> corrections). One seat (edit-quality, round 4) did catch a *symbol-existence* gap in the
> same trail — a phantom helper the sketch called as if pre-existing — which shows the
> class is checkable; nothing checks the deeper level of whether the sketch's logic is
> *live* against the real data flow.
>
> **Proposed fix.** Create a seat — or extend edit-quality's mandate — as a
> **sketch-falsifier**: for every edit whose sketch modifies or reads an *existing*
> function, the seat must fetch that function's current body (at the submission's ref) and
> attempt to **refute** the sketch against it, defaulting to object when it cannot ground a
> claim. Its checklist, in order of the failures actually observed: (1) every symbol the
> sketch calls or reads exists, or the sketch explicitly adds it; (2) every conditional the
> sketch introduces is *reachable* given the surrounding code — a guard keyed on state the
> function provably never produces (fresh local maps, write-once loops, values consumed
> before assignment) is dead code and grounds an objection at high severity when the plan's
> evidence chain depends on it firing; (3) invariants the sketch borrows from prose
> ("merges last", "runs once per field", "stored value survives") are quoted from the
> actual source lines, not restated from the rationale. The seat's review must include the
> quoted lines of the target function it verified against — an assertion without the
> quote is treated as unverified. Footprint: fires whenever a submission's edits touch
> `platform/`, `internal/` or `pkg/` files with `operation: modify`; it can abstain on
> pure additions of new files. This is the reviewer analogue of the repo's standing rule
> that a verified fact needs its evidence inline: **a sketch is a claim about the codebase,
> and the gate currently has no seat whose job is to test that claim.**

## Supporting detail (for the implementing thread, not part of the handover paragraph)

- **Evidence trail:** submissions v3–v5 on 2525f980 carried the dead log; round-4
  edit-quality caught the phantom `sectionInputSchema` (symbol level) but no seat caught the
  unreachable conditional (flow level). Correction: doc_notes `pipeline/plan_sections`,
  2026-07-20 12:44, commit `b6e374fc2`; independently re-verified by the trail author.
- **Why existing seats don't cover it:** edit-quality reviews the plan *as written* and can
  withhold on verification gaps, but its round-4/5 behaviour shows its bar is
  evidence-in-the-submission, not evidence-from-the-source; debug-historian audits process
  discipline (staging, deploy verification), not code truth; bug-historian matches pattern
  history. The seat brief should quote this trail as its founding case, the same way
  bug-historian's pilot quoted its own.
- **Mechanism already exists:** reviewers have read-only checks (`code_check` /
  data_requests were referenced by prior_art_librarian in round 5) — the gap is mandate,
  not tooling. Seat as usual via fix-proposer, then run the 099 roster mirror; do not
  hand-patch the gate.
- **Cost note:** the seat reads only the functions named by `modify` edits — bounded by the
  ≤8-edit cap — so its token cost is proportional to the plan, not the repo.
