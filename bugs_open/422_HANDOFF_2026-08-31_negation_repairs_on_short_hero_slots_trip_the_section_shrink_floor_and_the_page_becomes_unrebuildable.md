# 422 — negation repairs on short hero/CTA slots trip the section-shrink floor, and the page becomes unrebuildable

**Filed:** 2026-08-31, `copy_quality_two_stage` lane, from the v1.0.1349 canary widening
(finetuning.uk, nine-page rebuild through the corrected copy stack).

**Diagnosis-loop note (owner ruling 2026-07-31):** filed on first-hand verification — the
substitution is four independent reproductions of the same arithmetic on two pages, each with
both mechanisms' own error text naming their halves, and the failure ladder's terminal state
reached on one of them. No inference bridges the chain; a 090 run would re-read the same rows.

## Symptom

A `needs_page` rebuild of a page whose hero/CTA slots are short and negation-dense fails at
`save_sections` with `SECTION SHRINK REFUSED`, retries through the failure ladder, and lands
`failed` (unresolvable). The page keeps serving its OLD copy; the corrected stack can never
reach it.

`[MEASURED 2026-08-31]`, finetuning.uk canary widening:
- `services`: **3 of 3 attempts** refused — hero-services 378→169 visible chars (45% kept),
  call-to-action 426→163 (38%); floor 50%. Item `failed`, terminal.
- `about`: **3 of 3 attempts** refused the same way (hero-about 424→183, 43%). Item `failed`,
  terminal — measured complete ~17:45Z, same night: **2 of the 9 canary pages end the ladder
  unrebuildable**, and both keep serving pre-A+B copy.
- The other 7 pages passed the same rebuild — their generations happened to be longer or less
  negation-dense, so this selects for exactly the pages whose heroes the register work most
  wants to reach.

## Mechanism — two correct guards colliding

1. `rewrite_negations` (Decisions A+B + ruling 7, live in v1.0.1349) repairs
   define-by-negation sentences by ending them before the comparison. On body prose this is
   measured-safe (canary: 10/12 splices meaning-intact, 2 refused by the acceptance guards).
   On a 3–4-sentence hero whose EVERY sentence carries a shape, accepted repairs remove 50%+
   of the slot's visible text.
2. `save_page_sections`' shrink floor (bugs_open/178, axis corrected by 293) refuses a
   same-named prose slot losing >50% visible text in one save — the guard that stops content
   loss. It cannot tell sanctioned repair-shrink from loss; its own error names the knob
   (`section_shrink_floor` in step config).

Neither is wrong. The repair does not know the save's budget; the save does not know the
shrink is sanctioned. The page loses.

## Fix candidates, ordered by what closes the door

1. **Make the repair self-limiting: a per-field shrink budget inside `rewrite_negations`.**
   The action already composes per-field splices (`spliced` map) and records every
   acceptance/rejection; add a running visible-length check that stops ACCEPTING further
   repairs for a field once cumulative shrink approaches the save floor (reject with reason
   `shrink_budget`, keeping the marker honest). The save never sees an over-shrunk slot; no
   operator knowledge needed; the bad state is unrepresentable. Costs: the floor value must be
   readable from one shared place (or conservatively hardcoded below 50%).
2. **Marker-aware floor**: `save_page_sections` accepts shrink attributable to the copy-gate
   marker's accepted repairs (byte accounting travels with the section). Closes the door but
   couples the two mechanisms across an orchestration boundary — the marker lives in
   CollectedData, and a floor that trusts a marker is a floor a bug in the marker disarms.
3. **Dispatch-time knob**: set `section_shrink_floor` in the rebuild step config for
   copy-stack rebuilds. Closes nothing — every future dispatcher must remember ("operators
   must remember X" is a defect); and a floor lowered for repair-shrink is also lowered for
   real loss in the same save.
4. Do nothing + hand-rescue: the failure ladder already parks these loudly, and the pages
   keep serving old copy. The cost is silent-looking: the register work reaches every page
   EXCEPT the worst ones.

Related, not duplicate: the offer lane's finding tonight (their gate, `isTruncationOf`) that
truncation is structurally wrong for `differentiated: true` points — different table, same
repair family. A candidate-1 budget is about VOLUME on undifferentiated prose; their prefix
test is about MEANING on differentiated points; both bound the same repair from different
sides.

## How to verify a fix

Re-fire `needs_page` for finetuning `services` (item_key `page_rerender:services` is terminal,
key free). Success = item completes, served hero changes, battery drops, and the copy-gate
marker shows repairs accepted up to the budget with `shrink_budget` rejections recorded for
the remainder. Control: a page with a long hero (approach) must still repair to the same
counts as today.
