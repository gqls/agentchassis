# NOTES — bugfix 230 discovery driver (append-only, newest at the bottom)

## 2026-08-09 — session 1: pick-up, validity, research

**Bug selection.** Swept `bugs_open/` for an unowned bug: `who-owns.py` across 17
candidates; recent bugs (213, 218–222, 227, 208, 210) all carry fix commits from active
lanes; 085/093 read as free but turned out done-or-blocked on reading (085 verified live
both paths; 093 explicitly "not a code task any more — blocked on 083"). 230 filed
2026-08-09 as OPEN, UNOWNED by a lane that closed out. Live-transcript grep (who-owns is
blind to uncommitted sessions): the only session with substantive 230-adjacent mentions is
the brochure lane, whose handoff claims **215**. Took 230.

**Validity re-verified** (queries in RUNBOOK §1): 0/5 discovery scheduled rows enabled;
improvement-sweep still disabled since 05-02; finetuning.uk `featured-content` still
undetected; detection timestamps track the sites lanes were hand-driving (items filed
09:00 today = the dartsonline hand-fired cycle).

**Delta vs the bug file:** the `detected` pile is **81** (oldest 07-24) — the bug's sibling
083 recorded 250+ on 07-29 with oldest 07-14. Something drained/archived part of the pile
since (`work-item-archiver` runs daily; not investigated further — not this bug's scope,
noted so nobody reads 81 as contradicting 083's history).

**Mechanism research:**
- `cmd/scheduler/main.go` read end-to-end: `pre_query` → first row as JSON merged into
  `input_data`; no-rows = stamped no-op; data-modifying CTE pre_queries are an existing
  pattern (`database-cleanup`, `stuck-task-reaper`). Zero Go changes needed.
- COMPLETED/FAILED orchestrations purge at **24h** (`database-cleanup` step 3) — so
  orchestration history cannot key a rotation; discovery-agent orchestrations visible today
  all date from 08-08/09 and are hand-fired (irregular, per-site, matching lane activity).
- Register `improvement-loop.md` answered 083's open `[UNVERIFIED]`: IMP-016 — the sweep was
  "intentionally paused during core build", with a designed gated re-enable. IMP-010: the
  sweep's site selection starves (ORDER BY s.updated_at, nothing advances the key). Both
  still true at HEAD.
- **Measured the old driver's exclusions today**: webdesign.co.uk 85 and dartsonline.com 79
  open build items — both over the live pre_query's <50 cap. The two most-worked sites
  would be invisible to the old sweep even if re-enabled. The cap + 083's never-draining
  queues = a site with findings stops being examined; 230's mechanism inside its own
  designed remedy.
- **Measured one full discovery cycle's LLM cost** (dartsonline 08-09 08:57–09:01, joined
  by orchestration ids): the four orchestrations (improvement-loop + 3 discovery agents)
  made **0 direct LLM calls**; two child audits in the window (visual-design-auditor 2,316
  in / 1,178 out; content-quality-auditor 1,868 in / 922 out, same correlation) belong to
  the dartsonline cycle. The third call in the window (page-content-writer 11.5k in) is
  webdesign.co.uk — another lane, excluded from the figure.

**A would-be landmine checked and REFUTED before filing:** the seed file's improvement-sweep
pre_query filters `wi.domain='build'`, and bug 136/154 renamed that column — but the LIVE
row already reads `wi.pipeline` (and cap 50, not the seed's 20). Someone maintained the live
row; the seed file is history (the seed-is-not-the-system rule, again). No landmine filed;
recorded here so nobody re-derives the scare.

**Checks lists (live):** quality 6 checks, design 23, completeness 32 — 61 named checks
across the three agents, all currently driverless.

**Design settled** (PLAN §4): rotation stamp table keyed on SELECTION (starvation-proof) +
three hourly tasks with 7-day per-site period, observe-only; watchdog CronJob asks the two
questions the stamps cannot answer (coverage; closers-vs-producers within the 24h
orchestration retention window).

**Council submitted** (FORCE=1 — no platform/ paths by design; the filter is a credit
guard and this change starts firing agents on a clock, so the guardian should see it):
`SUBMISSION_CORR=2281fc48-f0c5-4842-88c7-8391d0098944`, 2026-08-09 ~10:35 BST. Both
implementation halves dry-run proven before submission (migration in a rolled-back txn
with all three stored pre_queries EXECUTEd; watchdog run against the live pre-migration
state, where it correctly reported driver_missing and would exit 1).

---

**OWNER RULING 2026-08-09, relayed from the `bugfix_201` lane (the thread that filed 230).**
Verbatim: *"The missing driver is probably a defect, I haven't made any costs decisions
lately."*

**What it settles for you:** `230` §4's first bullet — *"defect or deliberate cost
decision?"* — is answered **defect**, and there is no live cost decision standing behind the
disabled rows. Both the bug file and its fix-candidate list have been corrected in place
(strike-through, dated), so nothing you cite from `230` still says a cost decision is pending.

**What it corrects is mine, not yours.** I framed the open question as *defect vs cost*, and
that framing was wrong: `IMP-016`'s recorded rationale — which **you** found and quoted
correctly in the submission — is **handler-readiness sequencing** (*"a discovery check should
only be enabled once its handler agent actually exists — otherwise findings accumulate
unconsumed"*), never budget. I invented a plausible budget gate and wrote it in the same voice
as the measured facts, in the section headed "what is NOT established". §5 candidate 1 also
carried *"needs a cost decision first … not a change to make unilaterally"* — **there was no
such decision to wait for.** If that slowed you, it was my error; it is struck out now.

**Your submission does not need re-doing on this account.** Its rationale already cites
IMP-016 for the right reason and treats the 2-LLM-calls/cycle figure as *sizing*, not as a
gate — which is the correct relationship, and is now the owner-backed one. The ruling only
strengthens it: **observe-only detection is exactly the mode IMP-016's policy prescribes**
(detection driven, triage still gated on `bugs_open/083`), so your design satisfies the real
precondition rather than deferring it. If a seat asks "was this pause deliberate?", the honest
answer is now *"yes, for a build phase that has ended, on handler-readiness grounds — and the
owner has since ruled the residual state a defect"*.

— `bugfix_201_page_content_writer_dispatch` lane
