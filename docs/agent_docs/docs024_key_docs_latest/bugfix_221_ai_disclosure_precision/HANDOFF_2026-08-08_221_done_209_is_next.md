# HANDOFF — 2026-08-08 (late). 221 is DONE. The next unowned bug is 209.

Read this first on a cold start. `SUMMARY_2026-08-08_*.md` is the read-aloud
version; `NOTES` has the evidence and every misstep; `RUNBOOK` has the commands.

---

## Part 1 — `bugs_open/221`: complete, nothing outstanding

**FIXED, council-APPROVED, LIVE and PROVEN on chassis v1.0.1268.**

| | |
|---|---|
| fix | `61c8cc6ff` (`validate_page_content.go` + `validate_page_content_meta_scope_test.go`) |
| council | `377a0488-214e-4e5c-bd3d-66343d34d9b2` — APPROVED round 1, 11 seats, 1 medium + 3 low, all dispositioned in the bug file |
| live proof | marker **0** on v1.0.1267, **1** on both v1.0.1268 replicas, shared-string control **1** everywhere; fleet uniform 45/45 |
| behaviour | the offending row (12,879 B, unchanged) returns **0** blockers under shipped code; injected apology still blocks |
| consumer told | `webdesign_couk/CONTRIB_2026-08-08_221_tools_index_unblocked.md` |

**The file stays in `bugs_open/`** per the owner's 2026-08-06 direction, not
because anything is owed.

### Two loose ends, neither blocking

1. **The landmine verifier never ran.** Fired by hand (`49f3a981`), FAILED at
   19:16 in the fleet-wide Anthropic credit exhaustion (33 failures 18:25–19:22;
   12 more in the following 20 minutes). **Re-fire when credit is back:**
   ```
   ./scripts/trigger-landmine-verifier.sh 'LANDMINES.md#a-new-pattern-in-validatepagecontent-is-a-blocker-by-default-and-a-blocker-there'
   ```
   ⚠ Per `bugs_open/223` its verdict is weak evidence **in both directions** for
   an entry with non-Go footprints. Do not delete or downgrade the entry on a
   STALE verdict. This is tidy-up, not risk.
2. **The queued rebuild was deliberately not fired.** webdesign.co.uk
   `tools-index` is `build_status='needs_rebuild'`. It is that lane's page, a
   rebuild regenerates content, and firing it to collect my own evidence would
   take their decision. They have been told; the queue will do it.

### What this fix explicitly did NOT do

The `bug_historian` seat's **medium** objection stands and is not closed: the
*generic* mechanism — any blocker-severity prose scan able to wedge a page's
rebuild forever — is untouched. The other **13** pattern entries are unchanged
and still substring-matched at blocker severity. `bugs_open/222` is the sibling
instance in `check_tool_fabrication_action.go` (a comment *denying* fabrication
is convicted), **owned by `mortgagecalculator_couk_adoption` — do not take it.**
The governance question is the RFC `bugs_open/221` names.

---

## Part 2 — the next bug: `bugs_open/209`, and why it is the one

**Every other file in `bugs_open/` (63 files) was swept through `who-owns.py`
this session. All are owned by an ACTIVE lane except 209.** The near-misses,
so you do not re-derive them:

| bug | why not |
|---|---|
| **116** | owner-gated; *"an owner decision, not a fix"*; two active lanes; D3 run 1 fired 08-08 |
| **226** | filed **three minutes** before I looked, by `0c5a11f2` (oufe rerender-safety), still in the file. Ink wet — competing |
| 222, 224, 225, 227 | owned (mortgagecalculator, loanandmortgagecalculator ×2, loancalculator) |
| 115, 146, 153, 155, 184, 186 | owned, all ACTIVE |
| 223 | `who-owns` now names **my** lane — that is a **false positive**: my docs merely *cite* 223 about the verifier's blind spot. I do not own it |

### 209 — `deploy_image_asset` still resolves a source by PURPOSE, and that lookup runs FIRST

`bugs_open/209_HANDOFF_2026-08-06_deploy_still_resolves_a_source_by_purpose_from_collected_data.md`

- Filed 2026-08-06 by the `bugfix_152_155_asset_source_identity` lane while
  running 155's closure test. **It is 155's second arm**, and the file says so:
  155 closed on its own recipe, and this exists *"because closing it without
  naming what survived would have retired the class on one arm's evidence."*
- The file declares itself **"OPEN, unowned"** — a deliberate handoff.
- Severity **medium**: same wrong-bytes outcome as 155, but confined to a single
  build workflow, and **no live instance is yet demonstrated.**
- Locus: `findStorageURI`, `platform/orchestration/actions/deploy_image_asset_action.go`.

**⚠ I committed this file in `a27462433` as an explicit `sweep:`.** It had been
**untracked for two days** — one `git clean` from gone, invisible to
`who-owns.py` and to 016b's index, so the standard "grep before you file" check
saw no 209 at all. The content is verbatim; I changed nothing.

### Before you start on 209 — the checks that actually mattered this session

1. **Re-check ownership yourself.** `who-owns.py` reads *commits*, so a session
   mid-fix is invisible. Also grep the live transcripts, which is what caught
   226 for me:
   ```bash
   ls -t ~/.claude/projects/-home-ant-projects-agentchassis/*.jsonl | head -16 | while read f; do
     c=$(grep -c "findStorageURI\|209_HANDOFF" "$f" 2>/dev/null); [ "$c" != "0" ] && echo "$c $(basename $f)"; done
   ```
   Now that 209 is committed, the 152/155 lane may also pick it back up — check
   their dir for movement before assuming it is still free.
2. **Re-verify the bug is live.** Mine was, but the file was 11 hours old on a
   tree taking ~1,500 commits/week. 209 is **two days** old and its own filing
   says no live instance was demonstrated — so *"is this still true, and was it
   ever reproducible?"* is a real question here, not a formality. **Run the
   code, do not grep the database** (see the RUNBOOK's §2 harness pattern — it
   generalises).

---

## Part 3 — process learnings worth carrying, all paid for this session

- **`head` is never an answer to "does this exist?"** `grep | head -30` over a
  7,000-line `LANDMINES.md` made me commit *"no entry exists for this"* when one
  did, at line 6,413, **already citing 221**. Count first (`grep -c`).
- **Assert the SIZE of any absence or set against something independent.** Three
  separate silent truncations this session — `kubectl exec -i` eating a
  `while read` loop's stdin; a `sed` range ending at `/^}/` which matched `}{`
  and compared 15 items against 0; the `head` above. Same shape every time: an
  operation yielding less than everything, **at exit 0, with no marker**.
  `pattern-check.py` has a detector for the first and it only scans `.sh` files —
  none of the three could have been caught, because all three were typed into a
  tool call.
- **"Not there yet" and "never going to be there" are the same observation.** I
  wrote that the verifier verdict was "queued"; it had FAILED. Ask
  `orchestration_states.status`, which separates them in one query.
- **A call-site count for an ACTION is a DB fact, not a repo fact.** I reported
  one call site (true of the Go function) when four live agent definitions carry
  the step. The council caught it.
- **Check whether you actually HAVE a negative marker before promising one.**
  `comm -23 removed added` was **0** for my commit — nothing removed survives
  un-re-added. When that happens, **build** the control: a throwaway pod on the
  previous image, plus a string present in both images so the zero is a real
  zero and not a typo. Full recipe in the RUNBOOK.
- **Write the test BEFORE the fix and run it against unfixed code.** 7/7
  must-not-block failed, 7/7 must-block passed — that ordering is what proves
  the test can fail *and* that the change only narrows. Then mutate each part.
- **Fleet credit was exhausted from 18:25.** Council got in at 18:22 by luck. If
  dispatches are failing, check `agent_error_log` for credit/quota/429 before
  diagnosing your own payload.
