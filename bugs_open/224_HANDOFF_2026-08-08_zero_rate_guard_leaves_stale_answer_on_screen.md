
---

## 2026-08-14 (LMC lane) — the fix was TRANSIENTLY REVERTED on loans/standard-calc, by a restore, not by a rewrite

For the record on this bug's own page: between 2026-08-12 ~18:20Z and 2026-08-14, the
live `loans/standard-calc.html` served the PRE-FIX 0% APR guard again. Cause: a Track B
rollback used `load_lmc.py --restore`, whose backup table
(`page_components_bak_20260805_lmc`) predates this fix by three days. The restore was
deployed, a new decomposition pin was then taken from the poisoned tip, and the seeds
built from it propagated the old guard faithfully. `mortgages/overpayment.html` lost the
08-09 `btn-calculate` id the same way. Found by the full oracle sweep (6 FAILs, all this
bug's 0% signature), repaired from the last clean pin `7e6b993ef`, re-proven by oracle.
Full mechanism: `LANDMINES.md` §"A restore from a dated backup is a TIME MACHINE".
No change to this bug's status — the FIX is correct; what failed was rollback hygiene.
