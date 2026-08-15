# CONTRIB 2026-08-15 (~18:10Z) — two register fields superseded mid-rebuild, by the copy_quality_two_stage lane

**From:** `copy_quality_two_stage` (session `claude-session-copyquality-20260815`).
**To:** the loancalculator rebuild thread (fire-in-flight handoff, phase 1 corr `2d950ecc`).
**What changed about your guarantee:** any writer run that starts after ~18:0xZ on
2026-08-15 reads a corrected `strategy` and `content_direction` for this site. Pages
already built this morning were written under the old register. Your locks, compositions
and the 282 sequence are untouched — both edits are prose-register only.

## What was changed, and why

The owner ruled today (copy_quality_two_stage D3, decisions 1+2) that two fields carried
the insider-secrets register rejected on 08-08 on the mortgage site:

1. `strategy.value_proposition` final clause — *"…that lenders have no incentive to
   volunteer"* → *"…that decide what you pay"*.
   Old row `b82e2c7e-8e7e-432e-8130-b0d981805a11` (superseded, preserved) →
   new current `bca3b9ee-2278-4706-a158-a7d30b6ea6c8`.
2. `content_direction.example_phrases.characteristic[6]` — *"Our calculators and guides
   are built to reveal the true cost of credit."* → *"Our calculators and guides show
   what a loan costs in total, and which parts of that cost you can change."* — changed
   in BOTH the array and `formatted` (the field writers actually read; proven the two
   `formatted` values differ by exactly that one line).
   Old row `a1feaaa7-0543-4bcf-b401-abbe03c26b3c` (superseded, preserved) →
   new current `7f39172e-9933-4688-ae99-abab17034ac9`.

Both by supersede mirroring `write_site_spec`'s transaction shape; the old rows are the
rollback (flip `is_current` back).

## The one thing to watch

Your fire regenerated these specs this morning (07:58–08:31Z, `domain-research-classifier`
/ `domain-strategist`) and the offending clauses survived that regeneration verbatim. If
your lane re-fires the research/strategy agents for this site, **the corrections will be
clobbered by the fresh classifier output** — `write_site_spec` merges the incoming spec
over current, so a regenerated `value_proposition` / `example_phrases` replaces ours. If
that happens, re-apply from the row ids above or ping `copy_quality_two_stage` (NOTES
2026-08-15 has the full record). We deliberately did NOT pin the rows — `pinned` has
meaning only on the evidence-base path (`refresh_evidence_base_action.go`), not in
`write_site_spec`.

Pages built earlier today carry the old register; nothing here asks you to rebuild them —
the copy lane will pick divergence up in its own sweeps.
