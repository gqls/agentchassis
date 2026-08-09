# CONTRIB 2026-08-09 — your "ORIGINAL WRONG" stamp-duty finding: the owner decided, and the original is now fixed

From the bugfix-225 session (not this lane). Your `HANDOFF_2026-08-08c` §next-action 3
routed the stamp-duty ORIGINAL defect to the owner ("do NOT 'fix' the original
(owner controls it)"). **The owner has now decided: patch the original too**
(2026-08-09, during the bugs_open/225 fix — your original's `calcSDLT` was
byte-identical to the LMC page's, so it was the same one-block fix).

What changed on your side:

- `sites/mortgagecalculator.co.uk/stamp-duty.html` (sites `9d1a17202`) now
  computes post-2025-04-01 FTB relief (cap £500,000, withdrawn entirely above)
  and the £40,000 higher-rates floor. Your worked example — £595k FTB — now
  quotes £19,750, not £14,750. Live wire verified (sha `28e04d99…`,
  `grep -c 625000` → 0).
- Your `/tools/stamp-duty/index.html` rebuild was already correct and was used
  as a cross-reference; it is untouched.
- Your REPLAY-FAIL blocker (option VALUES missing from the id contract) is
  unaffected — the original still uses `ftb`/`next`/`additional` option values
  and the same element ids; only the inline script block changed.
- The defect pair is recorded durably in `bugs_open/225` §"Fix landed"
  (the twin was previously unrecorded there).

Nothing else of yours was touched. Your three §0.5 both-right model splits
remain with the owner as before.
