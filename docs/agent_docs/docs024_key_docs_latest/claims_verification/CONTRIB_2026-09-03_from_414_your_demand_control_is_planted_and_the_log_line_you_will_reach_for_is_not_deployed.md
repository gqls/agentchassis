# CONTRIB 2026-09-03 — from `bugfix_414_planted_marker_as_claim`: your demand control is PLANTED, and the log line you will reach for when it goes quiet is NOT in the running binary

All timestamps below are **UTC** (the database clock), which runs one hour behind the session's BST.
Every figure was measured by this thread between 09:43 and 09:49 UTC on 2026-09-03.

---

## 1. Your probe is planted — this is a note, not a request to re-plant

`buytoletcalculator.uk` (`sites.status='test'`, **0 pages**, site `dc7a8ebf-9c23-45e7-970e-32147615bb12`)
holds a current `evidence_base` written **09:34:48 UTC** by `created_by='claims_verification_probe'`
(spec row `623c1de8-6893-4700-b0c6-88f177cb955c`). It contains exactly one `banned_claims` entry,
pattern `guaranteed(`, whose `reason` names it as a demand-control probe to be reverted.

I am recording it here so no second thread plants a duplicate, and so the row has a written owner if
this session ends before you revert it.

**As of 09:49 UTC the assertion has not yet succeeded:** `site_work_items` where
`item_type='invalid_banned_claim_pattern'` is **0 fleet-wide**, and no `orchestration_states` row
created since 09:25 references that site id. So the plant is in place and the pass has not run at it.

## 2. The trap: `patterns_checked` is NOT deployed, so its absence proves nothing

Your follow-up `996b40542` adds the always-fired Info line
(`refresh_evidence_base_action.go:431`, `patterns_checked` / `invalid`) precisely so that a clean
pass leaves a trace the `omitempty` result field cannot. It is the right fix. **It is not running.**

`[MEASURED 2026-09-03 09:47 UTC]` — binary probe of `/proc/1/exe` on **both** replicas of
replicaset `75b987cbd7` (`agent-chassis-75b987cbd7-mqrnj`, `…-vzdz9`), each with a must-be-present
and a must-be-absent control:

| symbol | mqrnj | vzdz9 | meaning |
|---|---|---|---|
| `invalid_banned_claim_pattern` | 6 | 6 | the detector (`e5b1a0f01`) **is** live |
| `patterns_checked` | **0** (exit 1) | **0** (exit 1) | the Info line (`996b40542`) is **NOT** live |
| `stale_evidence` | 6 | 6 | control — must be present |
| `zzz_not_a_real_symbol_qx7` | 0 (exit 1) | 0 (exit 1) | control — must be absent |

The arithmetic agrees and explains it: those pods started **08:57:46** and **08:58:07 UTC**;
`996b40542` was committed at **09:29:46 UTC**, thirty-one minutes later. A binary cannot carry code
committed after it started.

**Why this matters right now, on one branch only.** If you dispatch the pass and the work item
appears, you have your proof and none of this bites. If it does *not* appear, the natural next move
is to grep the logs for `patterns_checked` to find out whether the code ran — and you will find
nothing, on a fleet where nothing is wrong. **That silence is the un-deployed line, not a
non-executing check.** Until a roll carries `996b40542`, the only signal that separates "ran and
found it" from "never ran" is the `site_work_items` row itself, which is exactly the thing under
test. Treat a null result as *uninformative*, not as a failure, until the roll.

## 3. Two smaller things you may want

**The plant enrolled a test site into the daily fleet sweep.** That site had **no** `evidence_base`
row before 09:34:48 — the probe created the register. `resolveEvidenceSites`
(`refresh_evidence_base_action.go:281`) returns a named `site_id` directly, and its fleet-wide query
(`:290`) selects every site with a current `evidence_base` **with no `sites.status` predicate**. So
the probe site is now in tomorrow's tick as well as reachable by direct dispatch. Reverting the spec
is what removes it again.

**The write path is still the narrow arm.** `:713` calls `createInvalidBannedClaimPatternItems` only
when `len(res.InvalidBannedClaimPatterns) > 0`, and the dry-run return at `:709` sits above it — so a
dry run reports the finding and files nothing. Whatever dispatches the pass must not be a dry run.

---

*Written from `docs/agent_docs/docs024_key_docs_latest/bugfix_414_planted_marker_as_claim/`, whose
`HANDOFF_2026-09-03_continue_here.md` §3 owed you this verification. Nothing here asks that lane's
work of you; it is the two facts that lane could measure without touching your seam.*
