# START HERE — work_item_completion_integrity

**Cold-start entry point for this workstream.** Written 2026-07-20 by the thread that
closed `bugs_open/017` ("bugfix thread2"), so the next chat can resume without re-deriving
anything. Read this file, then the two inbound handoffs named in §3.

**Remit of this thread, in one sentence:** whether a `site_work_items` row can be trusted
to mean what it says.

---

## 1. State in 30 seconds

| | |
|---|---|
| **`bugs_open/017`** | ✅ **CLOSED 2026-07-20** — live in **v1.0.1139**, pod-verified. Moved to `/bugs_closed/017_…` |
| **Branch** | `085_debug_and_feature_loops` |
| **Commits** | `c82b2872c` (fix) · `c80fffc83` (council r2 follow-up) · `205b73a28` (§9 queue trap) · `41e3345b2` (standing five + corrections) · `93edb02f7` (closure) |
| **Council** | `SUBMISSION_CORR=319e23f6-b333-42ba-88ef-069b4426c057` — r1 REVISE, r2 REVISE at **8 approve / 2 object**. No `Council-Reviewed:` trailer claimed (verdict was never APPROVED) |
| **Live check** | defining sweep = **0**; no regressions |
| **In flight** | **nothing.** No dispatches pending, no uncommitted work of mine |
| **Next** | two inbound assignments — §3. One was assigned to this thread **by the owner** |

## 2. What 017 was, and what shipped

`site_work_items` could be stamped `status='complete'` while storing, in the same UPDATE,
the error proving the work never happened.

The root confusion is **two `status` fields one layer apart**:
`result.response_status` is **delivery** (the coordinator sets it to `'complete'` whenever
*any* reply arrives, `coordinator.go:2398-99`); `result.response.status` is the saga's own
**verdict**. The completion path read the first as if it were the second.

Shipped in v1.0.1139:

- **`handlerReportedFailure`** (`complete_work_item_verification.go`) — blocks completion on
  an explicit `failed`/`failure`/`error` verdict, routing into the *existing* attempt
  machinery via a generalised `failUnverifiedCompletion`. Runs **before** the per-item-type
  verifier.
- **`recordUnknownVerdict`** — an unfamiliar verdict still COMPLETES (a novel status is not
  evidence of failure) but is written to `agent_error_log` as
  `error_code='UNKNOWN_HANDLER_VERDICT'`, because a `zap.Warn` dies with the pod.
- **`fix_forced_text_colors` registered** in `GlobalActionRegistry` (leg 1; landed via
  `06376bcbf` when another session swept my tree).
- **`registry_parity_test.go`** — the build now fails if any action registers an
  `ActionInputSpec` with no registry entry, with a `dormantActions` allowlist.
- **Deleted** the dead `actioncheck.LocalActions` map + the comment and two guide docs that
  told authors to "register in TWO places".
- **Data:** 54 mis-stamped rows corrected to `failed`, reversible via `result._correction`.

**Design decisions and their reasons are in `PLAN_…md` — read that table before changing
any of it.** Two corrections to the original bug report are recorded there too (its root
cause was wrong; its scale was 27× understated).

## 3. What's next — two inbound handoffs, both already council-reviewed

Both sit in this directory. **Neither is started.** Between them they carry five council
rounds already paid for; the outstanding objections are enumerated so you do not re-spend
them.

### A. `bugs_open/032` + `bugs_open/021` §INSTANCE 2 — the verifier layer
`HANDOFF_2026-07-19_verifier_absent_row_defect_and_coverage.md`

> **⚠️ THE HANDOFF IS OUT OF DATE ON 032 — CHECKED 2026-07-20, DO NOT REDO THIS WORK.**
> It describes 032 as an open live defect with a fix "drafted". **The fix is written and
> committed** — `a467baa11`, *"a deleted component no longer verifies as a successful
> fix"* — and it is the conservative shape the handoff recommended: return an **error**,
> not a verdict, relying on our gate failing OPEN on verifier error.
> **But it is INERT.** Pod-grep of the running binary for its discriminating string
> `"genuinely fixed or silently deleted"` → **0**, with my own 017 guard string → **1** as
> a positive control. The commit (10:33 UTC) postdates the pod start (07:35 UTC). So 032
> is in exactly the state 017 was in yesterday — fixed, not live — which is why it
> correctly remains in `/bugs_open/`. **Next action on 032 is an image roll, not code.**
> Its file names `empty_sections_loop_integrity` as owner; closing it is theirs.
>
> Residual deliberately left open *in the code comment itself*: if the page still EXPECTS
> the component (a `plan_sections` entry, a slot reference), absence is not ambiguous — it
> is deletion, and `Resolved:false` is the honest verdict. Bigger change, assigned to
> `empty_sections_loop_integrity`, and the error-return floor does not preclude it.

- **Coverage gap (021 §2) — STILL OPEN and still this thread's policy call.** Re-verified
  2026-07-20: `RegisterVerifier` is called **exactly once** repo-wide
  (`check_empty_sections.go`), against ~50 item types with discovery checks. 4,570
  completions carry 5 `_verification` records. The mechanism is opt-in by construction
  (`verifiers.go:47-51`), so it stays at one unless an author remembers. **This, not 032,
  is what remains of handoff A.**
- **Plan:** `reasoning_dataset/submission_B_register_more_item_verifiers.json`
  (`SUBMISSION_CORR=66dbd0dd-de5f-4f50-acd3-f5f3d817dbd9`, 2 rounds, both REVISE).
  **Take as a starting point, not a finished plan** — the `phantom_internal_link` edit is
  "a stub dressed as an edit", `VerifierCoverage()` under-reports by iterating only the
  check registry, and the known-gap allowlist (~47 entries) is sketched, not enumerated.

### B. Submission A — work-item origin provenance (**owner-assigned 2026-07-20**)
`HANDOFF_2026-07-20_submission_A_work_item_origin_provenance.md`

One nullable `TEXT` column `site_work_items.origin_correlation_id`, populated at the single
INSERT in `write_audit_findings_action.go:657`, plus a partial index. **Three council
rounds, all REVISE, converging — "two small answers away", both drafted in the handoff.**

**Confirmed NOT started (2026-07-20):** the column does not exist on `site_work_items`, and
`origin_correlation_id` appears nowhere in `platform/`. So this one is genuinely
green-field, unlike A.

The standalone case: auditors make ~15,000 LLM judgements a month that become work items
with real terminal outcomes, and **nothing links the two** — so an auditor flagging twenty
non-issues is indistinguishable in the data from one flagging twenty real defects. The
filing thread declares a secondary interest (dataset ground truth) and states the platform
motive wins if they conflict.

**It also documents how to decline** if you judge it out of remit.

## 4. How to verify anything here (full commands in `RUNBOOK_…md`)

```sql
-- the defining query. MUST be 0. Any new row = the guard is not in the running pod.
SELECT count(*) FROM site_work_items
WHERE status='complete' AND result->'response'->>'status'='failed';

-- has the guard ever actually blocked in production? (still 0 as of 2026-07-20)
SELECT id, item_type, status, attempt_count FROM site_work_items
WHERE error LIKE 'completion blocked: handler saga reported failure%';

-- an unfamiliar handler verdict appeared → widen the allowlist
SELECT * FROM agent_error_log WHERE error_code='UNKNOWN_HANDLER_VERDICT';
```

## 5. Traps this thread paid for — do not re-learn these

1. **The pod-grep passes on a string your change merely USES.**
   `grep -c fix_forced_text_colors` returned 1 *before* the fix too (the action file always
   called `RegisterActionInputSpec`), so the old image passes identically. Grep a symbol
   that **cannot exist unless your change shipped** — for a config-shaped change, a literal
   from the changed line itself (`"Strip forced child-text colours"`). Always pair with a
   positive control. → 016b §9.
2. **A queued orchestration is indistinguishable from a dropped one.** No
   `orchestration_state_audit` rows meant *queued* (~16 min under backlog vs ~10 s quiet),
   not dropped. I resubmitted 3× on untested hypotheses = 3 wasted council runs. Ask when
   **other** orchestrations started. → 016b §9.
3. **"It rests on an author-run audit" from a reviewer is a defect report, not
   box-ticking.** I asserted a structural claim about 8 call sites having opened 4 and
   inferred 3 from filenames; the council objected twice and I waved it off. The claim held
   — by luck, not method. → NOTES misstep 2.
4. **CLAUDE.md changes mid-session.** Its diagnosis section *inverted* during this work
   (filing is now the DEFAULT for any durable claim). Re-read it from disk before asserting
   anything durable, not just at session start.

## 6. Owner calls outstanding

- **Nothing on 017.** One monitoring note only: the guard's *blocking* path has not yet
  fired in production, so it rests on tests rather than an observed live block. This is
  stated plainly in the case file rather than implied away.
- **Which inbound to take first.** Revised after checking the code rather than trusting the
  handoffs: **032 is already fixed and needs only an image roll**, so the real choice is
  between **submission A** (owner-assigned, 3 council rounds in, genuinely unstarted, small
  and additive) and **021's coverage policy** (largest, least urgent, needs a decision on
  shape before any code). Submission A is the obvious next move.
- **An image roll would land 032** (and anything else committed since 07:35 UTC on
  2026-07-20). Not this thread's call to make, but worth knowing it is the gate on someone
  else's closed bug.
- **Three stale `detected` `hardcoded_section_colors` items** remain (robot-hands.com,
  vonc.com, gamesdesign.co.uk). Deliberately **not** dispatched: it would edit live sites
  with an action `017` itself judged misconceived, against the "mark them failed and start
  fresh" ruling. If you ever want leg 1 proven end-to-end, vonc.com (1 component, the
  platform's own test site) is the right canary — but ask first.

## 7. Reading order for a cold start

1. this file
2. `PLAN_…md` — design decisions **and their reasons**, plus the two corrections to the
   original bug report
3. `NOTES_…md` — the technical log, **including three recorded missteps**; newest at bottom
4. `HANDOFF_2026-07-19_verifier_absent…` and `HANDOFF_2026-07-20_submission_A…` — the work
5. `RUNBOOK_…md` — commands, each with its gotcha
6. `/bugs_closed/017_…` — the closed case with its verification evidence inline
7. `README_where_we_are.md` — the owner's plain-prose log (append only; never rewrite)
