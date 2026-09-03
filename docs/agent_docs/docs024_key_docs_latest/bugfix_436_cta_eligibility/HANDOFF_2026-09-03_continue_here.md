# HANDOFF — 2026-09-03. **START HERE.** `bugs_open/436` — lever + alarm BUILT, APPROVED, ROLLED, ENABLED. What is left is VERIFICATION, then close.

> ⚠ **SUPERSEDED 2026-09-03 mid-morning — read `HANDOFF_2026-09-03b_continue_here.md` instead.**
> Its §2 verification plan has been carried out: the check's first pass is observed and explained,
> and the induced canary passed two-way at the ranking. What remains is ONE blocked step (the header
> button at the served bytes) and the owner's opt-out decision — which now has real substance, not the
> "plausibly nothing to do" this file guesses. Kept for the trail; do not work from it.


**Lane:** `docs/agent_docs/docs024_key_docs_latest/bugfix_436_cta_eligibility/` · register **LNK-041**
(`docs026_concept_register/register/link-management.md`) · council **APPROVED round 2**, corr
`9faa2a23-f3bc-464e-8c3a-9d3d44759cc0` · commits `215c7eead` → `ffbbfc491` (all trailered; 098
credits them) · predecessor lane `bugfix_389_cta_relevance/` (391, CLOSED — the damage; 436 is the
cause).

## 0. Session-start checklist
`git log --oneline -10` · re-read this from disk · `scripts/who-owns.py 436` · re-run the §2
censuses (counts move on their own) · grep LANDMINES for any symbol you are about to trust.

## 1. WHAT IS DONE — everything except the live verification

| | state | proof |
|---|---|---|
| `pages.eligible_as_cta_target` (mig **714**) | **APPLIED 2026-09-02**, ledger-recorded | column present; 0 rows opted out (default true everywhere — zero behaviour change until someone flips a page) |
| Ranking binding (all 3 `chooseCTATargets` callers incl. header fallback) | **LIVE** since the 2026-09-03 roll | `service_binary_capabilities`: both running commits (`0d2feee2`, `7bf1ff67`) are descendants of `24b871535` |
| Label-match binding (refuse-not-drop) + judge reason `names_ineligible_page` | **LIVE**, same roll | same ancestry proof |
| `cta_rank_anomaly` check registered in the binary | **YES — 412 pods**, = positive control `misdirected_cta`'s 412; negative control absent | `[MEASURED 2026-09-03 ~09:20Z]` |
| Check ENABLED (mig **715_HOLD**) | **HAND-APPLIED 2026-09-03 ~09:22Z**, after the registration proof | checks array carries the name; `agent_definitions_backup` snapshot holds the PRE-change config (`has_old=t`). ⚠ NO ledger row — `--record-only` refuses `_HOLD` sidecars by design; NOTES is the record |
| Council | **APPROVED r2** (r1 REVISE actioned: near-empty-after-filter tested; snapshot + `715_ROLLBACK` added) | report: `diagnosis_artifacts` kind=`council_report`, corr `9faa2a23`, latest row |

**Nothing is opted out anywhere.** The lever exists; nobody has pulled it. That is deliberate —
owner decisions only (§4).

## 2. ⏳ WHAT IS LEFT BEFORE CLOSE — two verifications and one owner question

The close bar is **fixed AND live** (CLAUDE.md): the mechanism is live; what is missing is the
*verified* half. In order:

### 2a. The induced two-way canary (the bug file's own verification bar). ~1–2 h
1. Pick a low-stakes site with ≥2 tool pages. Read its ranking exactly as the code does:
   ```sql
   SELECT name, COALESCE(nav_order,100) FROM pages
   WHERE site_id='<site>' AND page_type IN ('tool','game') AND status IN ('active','deployed')
   ORDER BY COALESCE(nav_order,100), name LIMIT 3;
   ```
2. `UPDATE pages SET eligible_as_cta_target=false WHERE site_id='<site>' AND name='<rank1>';`
3. Dispatch a resolve/rebuild for ONE page (the RUNBOOK's canary section; `spec.page_name` is
   load-bearing on a rerender — without it the result is discarded).
4. Assert at the STORED field (`page_components.content_data` cta url ≠ the opted-out page) **and
   at the SERVED bytes for the header button** (`scripts/probe-page-url.sh`; the header's pick is
   NEVER persisted — no DB check can see it; this is council round-2 bug_historian's advisory and
   the one thing only a live look verifies).
5. Flip back to `true`, re-run, assert the page wins again. Both directions = done.
6. ⚠ Verify by artefact, never by work-item status (`complete` is not evidence — 391's rule).
   ⚠ Item keys dedup in ANY status (bugs_open/326) — a retry needs a fresh key.

### 2b. Observe the check's first fleet pass. Passive, or induce one
- `SELECT count(*), min(created_at) FROM site_work_items WHERE item_type='cta_rank_anomaly';`
  (was **0** at 09:25Z 2026-09-03 — the rotation had not run yet).
- ⚠ **"0 items" alone is UNREADABLE**: on a healthy site the check files nothing and positively
  RETRACTS. Pair the count with proof the check RAN — the discovery runner logs its enabled +
  registered arrays per run, or induce a run on one site. Expected: silent on all-default sites
  (webdesign's shape), silent on the demoted password-entropy trio (nav_order 900), fires only on
  a true fossil (unique min < 100, lead ≥ 50, ≥ 3 tools).
- If it FIRES anywhere: that is a real finding for a human — read it before assuming misfire. If
  it misfires: `715_enable_cta_rank_anomaly_check_ROLLBACK.sql` (hand-apply; snapshots first).

### 2c. Owner question (raise, don't assume — Phase 3 of the PLAN)
Does he want any live page opted out today? The demoted password-entropy trio are already
harmless at nav_order 900, so plausibly nothing to do. This is a USAGE decision, not a close
blocker — record his answer in the lane and act on it if any.

### Then: CLOSE
Move `bugs_open/436_…` → `bugs_closed/` (bar: fixed AND live AND verified per 2a+2b), update
`MEMORY_workstreams.md`'s 436 entry and MEMORY_closed, final SUMMARY file for the lane (first one
— none written yet; the close IS the milestone), and add the transferable pattern to 016b §9 if 2a
surfaces anything new (the lock-in pattern itself is already recorded in 391/LNK-040/LNK-041).

## 3. Traps this lane hit or dodged — do not re-derive
- **Migration numbers go stale mid-session** — 710 was taken under us; ours are **714/715**. Two
  WRONG_CALLS rows from this lane (stale directory listing; jsonb path composed from purpose
  instead of read from the row — the live step is `run_checks`, NOT `run_discovery_checks`).
- **The discovery runner FAILS the whole step on an unregistered check name**
  (`discovery_checks.go:195-218`) — that is why 715 was `_HOLD` and why its recipe probes
  `service_binary_capabilities` with BOTH controls before applying. Never flip
  `allow_unregistered_checks` fleet-wide to smuggle one name through.
- **`snapshot_agent` two-arg overload** writes `agent_definitions_backup` (one-arg writes an
  `is_snapshot` row into `agent_definitions`) — verify a snapshot holds the PRE-change config,
  not that one exists (LANDMINES).
- **Refuse-not-drop is load-bearing** at the label match: removing an opted-out page from the
  universe lets a one-token runner-up win (the measured self-link failure).
  `TestBestLabelMatchForPageRefusesAnOptedOutPage`'s runner-up subtest fails on a pre-filter
  rewrite — do not "simplify" the rule into `CTALabelUniverseSQL`'s WHERE clause.
- **The Go field is INVERTED** (`IneligibleAsCTATarget`; zero value = eligible) so literal
  candidates stay eligible. A polarity flip silently refuses every label match fleet-wide;
  `TestCTALabelCandidateRowCarriesEligibility` pins it.
- **The lever does NOT unlock label-locked fields on the recompute path** (KEEP #2 holds any
  valid stored page). Full rebuilds replace; recomputes keep. Stated in PLAN, bug file, LNK-041 —
  do not re-file it as a defect.
- The in-tree actions package was untestable on 2026-09-02 (a third session's untracked,
  non-compiling `invalid_banned_claim_pattern_test.go`); verify at committed HEAD in a worktree
  if it is still there.

## 4. Who was told what
- `bugs_open/399` lane: CONTRIB (judge gained `names_ineligible_page`).
- `cta_target_content_pass` lane: CONTRIB (decision-5 re-scope unblocked; re-measure the locked
  population AFTER the roll before scoping — it drains on its own).
- `bugs_open/114` session: two-way exchange on the tombstone test (they fixed it, `d1cf3aac3`)
  and file attributions.
- `bugs_open/384`: deliberately NOT contacted — listing arrays do not pass through CTA
  resolution; reason recorded in the PLAN.

## 5. The one-query status board
```sql
-- lever exists, nobody has pulled it:
SELECT count(*) FILTER (WHERE NOT eligible_as_cta_target) AS opted_out, count(*) AS pages FROM pages;
-- check enabled:
SELECT (default_config #> '{workflow,steps,run_checks,config,checks}') ? 'cta_rank_anomaly'
FROM agent_definitions WHERE type='completeness-discovery-agent' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- check registered in the running fleet (both controls):
SELECT name, count(*) FROM service_binary_capabilities WHERE kind='discovery_check'
  AND name IN ('misdirected_cta','cta_rank_anomaly','no_such_check_zz') GROUP BY name;
-- first-pass evidence:
SELECT count(*) FROM site_work_items WHERE item_type='cta_rank_anomaly';
```
