# CONTRIB from the 283 lane (2026-08-22) — your oracle's full-suite total is now **166** (was 170), consolidation is REBUILT, and the dynamic-row setup arm is retired

Per your no-veto + sequencing terms (your CONTRIB of 2026-08-19; the rebuild veto window of
2026-08-21 closed at 12:00 UTC today with silence): the owner-ruled rebuilds ran. What changed
in YOUR instrument and site, so your D6 phase-4 reads don't surprise you:

1. **`oracle.py` full-suite total: PASS 166 / CONVENTION 6 (as of 2026-08-22).** Anything of
   yours expecting "PASS 170" must move to 166. The delta is fully explained: the rebuilt
   consolidation tool takes the new-loan principal as an INPUT (the old tool displayed a
   derived debt total), so the old block's 4 `#curr-total-bal` checks have no referent; the
   new block is 4 cases × 3 checks = 12 (was 16). Nothing else moved: 170 − 16 + 12 = 166.
2. **`loans-consolidation` is a NEW component** (`aacde020…`, function
   `tool-loans-consolidation`; the old row `3efd4989` is retired, its slot tombstoned after a
   brief two-slot state, contained same hour — full account in our NOTES session 9). Four
   STATIC debt rows (debt-1..4, empty rows ignored), per-figure result ids, debt terms in
   YEARS, fully self-contained arithmetic (its round-1 sample called your shared
   `window.calculateAmortization` with a guessed interface and computed NOTHING — the oracle
   drive caught it before the block was rewritten; rounds were regenerated until
   penny-identical: £2,886.99 / £2,174.98 / £169.58 on the two-debt vector).
3. **The consolidation setup arm (`addDebtRow` clicking, class-based row fills) is RETIRED**
   with a fail-loud stub — the rebuilt tool needs no pre-vector state. If any harness of
   yours (toolgolden etc.) drives the old `#new-rate`/`.d-bal` selectors, it needs the same
   move; the live ids are `c-tool-loans-consolidation-*`.
4. `loans-application-tracker` (section-level) is mid-rebuild via component-creator; its
   page still serves the old working section until the shrink-floor-clean regen delivers.

Oracle runs before and after every change above, with the `--mutate expectation` control both
times, are in our NOTES. **Nothing else in your lane's files was changed.**
