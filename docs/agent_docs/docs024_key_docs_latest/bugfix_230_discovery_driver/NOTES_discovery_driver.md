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

## 2026-08-09 — council APPROVED r1; applied; LIVE and behaving

**Verdict read in full** (report saved to scratchpad; doc_notes verdict row + council_report
artifact): **APPROVED round 1, "4 advisory objection(s) — none high-severity", 6 abstained,
no truncation gating.** Every objection dispositioned:

- **editquality M — agent_type strings unguarded against agent_definitions.type:** REAL;
  closed before apply — migration guard now asserts each of the 3 tasks' target_agent_type
  matches a live active non-snapshot agent_definitions row (would have raised on a typo).
- **tooling_provenance M — no doc_notes subject row for the new pipeline:** closed — the
  migration now writes subject_key `site-discovery-rotation` (source `migration-346`);
  rollback removes it.
- **guardian M — enabled=true fleet-wide:** answered by the owner ruling recorded in
  bugs_open/230 §4 between submission and verdict ("probably a defect, I haven't made any
  costs decisions lately") plus IMP-016's real gate being handler-readiness, which
  observe-only respects. Kept ON; one-UPDATE pause stands in the RUNBOOK.
- **guardian M / tooling_provenance L — doc_notes.source landmine:** verified consistent —
  the sibling checks write the CHECK name (not the script filename) as source; ours matches
  its CronJob/service/dir name exactly, comment added in check.py, runbook queries by
  categories.
- **debug_historian M — sites.status filter unenumerated:** it WAS enumerated this session
  (GROUP BY: deployed 20 / pool 17 / system 1; no other values). 'active' matches nothing
  today and is kept only to mirror the designed driver's predicate. pool/system excluded
  deliberately.
- **editquality M / guardian L — ImagePullBackOff on the CronJob:** verified at the POD, not
  the Job: first manual run reached `Completed`, logs show the full report.
- **reuse L — 4th bespoke watchdog fork:** no generic scheduled-task-coverage watchdog
  exists (services dir enumerated); the architecture seat's forward note adopted: the
  stamp-table rotation is instance ONE of a potential shared primitive — if a second
  rotation of this shape is needed, THAT is the moment for an RFC, not a third copy.
- **improvement_guardian L — bespoke concurrency group:** deliberate; documented in PLAN §4b
  (avoids contending with the shared 'dispatch' group).
- **prior_art M — corroborate 0-of-5 from a DB-access session:** this session IS that
  corroboration (measured twice: filing lane 08-08/09, me 08-09 morning).
- **debug_historian L — no standalone verify script:** RUNBOOK §1/§3 hold the queries; the
  watchdog is the standing verifier.

**Applied ~10:47Z** by psql -f (whole file, guard passed) + `--record-only` with a note
(`--apply` NOT used: pending 342/345 belong to other threads — 345 is the 227 lane's
deliberately-parked fix). Watchdog applied with `kubectl apply -k`.

**Live behaviour, first 10 minutes — all three links working:**
- 09:49:51/09:50:21/09:50:51Z: the three tasks fired 30s apart (concurrency spacing as
  designed), all picking robot-hands.com (lowest site_id among unstamped, each agent's own
  rotation), stamps written.
- All three orchestrations **COMPLETED** (`*-orchestrate-0809-094x`, domain robot-hands.com)
  within ~1.5s of their stamp — selections produce runs.
- The runs FILED: 17 undeployed_asset + 4 literal_markdown + others, `status='detected'`,
  observe-only as designed.
- Watchdog manual run: pod `Completed`, exit 0, clean report (grace suppressing install-day
  NULLs; stamps 1/1/1 vs orchestrations 5/5/7 — the extra runs are the hand-fired
  dartsonline cycle in the same window), doc_notes row written.

**Still owed:** bugs_open/230 §6's canary — an `empty_section` item for finetuning.uk
`featured-content` appearing WITHOUT dispatch when the completeness rotation reaches it
(site_id order: robot-hands 00ff…, loancalculator 0162…, finetuning 1368… — ETA a few
hourly ticks). Until it lands, 230 stays OPEN; the fix is live but the bug's own
verification criterion is the arrival of that item.

## 2026-08-09 (afternoon) — the canary landed unprompted; bug 230's own criterion is satisfied

**§6 SATISFIED at 13:52:04Z.** The completeness rotation reached `finetuning.uk` on its
fourth tick (site_id order: robot-hands 00ff… 09:50 → loancalculator 0162… 10:51 → cookly
11:51 → idea 12:51 → **finetuning 1368… 13:52**) and filed both `featured-content`
`empty_section` items — the outstanding true positives this bug was filed over — with **no
session dispatching anything**. Orchestration `b5cf2140` /
`completeness-discovery-agent-orchestrate-0809-1352`, COMPLETED. 58 items on that site in
one unattended pass. All three links (schedule → run → finding) now demonstrated, which is
the thing [[detection-works-schedule-and-dispatch-do-not]] measured broken at link 2.

**Survived a fleet roll mid-verification**: chassis + kafka-scheduler went v1.0.1273 →
**v1.0.1274** around 12:25Z, between the 3rd and 4th rotation ticks. The rotation missed
nothing — 12:51 and 13:52 both fired and both produced COMPLETED runs. Worth recording
because a roll is the obvious thing to blame for a gap, and here there was no gap.

### Misstep this session, logged in WRONG_CALLS.md

I told the owner a tick "appears to have been missed" and offered the roll as the cause.
**False, from a BST-vs-UTC clock mix**: my watcher printed local time, the DB prints UTC,
so a 15-minute-old fire on an hourly task looked an hour overdue. `SELECT now()` in the
same result set killed it in one query — and the same query already contained
`now() - last_triggered_at AS since_fire`, which cannot be misread across zones. The
durable defence is reading the DB's own elapsed-time column, never pairing two wall clocks.
`bugs_open/085` records this exact trap and I had read that file **this session**; reading a
warning is not applying it.

### A second landmine found by verifying, not by suspecting

`site_work_items.created_by` carries the **SENDER**
(`platform/orchestration/actions/discovery_checks.go:132` —
`params.ExecutionContext.Sender.AgentType`), not the agent that ran the check. A
spawn-fired discovery run stamps the real agent type; a **scheduler**-fired one stamps
`generic`. Measured today: 177 items across the 5 rotated sites under `generic`, versus 33
+ 25 from one hand-fired cycle under the real names. **Any coverage census grouped by
`created_by` will therefore omit every scheduled run** — precisely the population this fix
created. Filed to LANDMINES with the join to use instead
(`orchestration_states.owner_agent_type`, or `site_discovery_rotation`). The `generic`
spelling itself is `[UNEXPLAINED]` — the scheduler sends `from_agent_type: kafka-scheduler`
— flagged rather than smoothed over.

### Bookkeeping honesty

My `WRONG_CALLS.md` commit (`5873985b0`) took a **same-file passenger**: the `bugfix_136`
lane appended its own entry between my read and my commit, so their text is in my commit
and their own commit (`7978889e2`, message mentions wrong_calls) carries only LANDMINES.
Nothing lost, attribution in the log is crossed. This is the documented pathspec limitation,
not a mistake to fix by rewriting history (forward-only).

### State at close of session

- Rotation: **live and self-driving**, 5 sites × 3 agents examined so far, ~1 site/agent/hour.
- Watchdog: deployed, daily 06:35Z, first manual run clean, doc_notes row written.
- 230: **fixed, live, verified**; stays in `bugs_open/` (owner ruling 2026-08-06).
- Untouched on purpose: `bugs_open/083`'s drain decision, and the `improvement-sweep` row.

## 2026-08-10 — CONTRIBUTION from the bugfix_236_site_availability lane (not a change to yours)

Telling you rather than only measuring you, per the owner ruling of 2026-07-29 §3.

**What arrives in your table.** `site_discovery_rotation` gains rows with a **fourth
`agent_type`, `availability-discovery-agent`**, written by a new scheduled task
`site-discovery-rotation-availability`. It is your §4b pre_query **verbatim** with two
values changed: the agent type, and the cooldown `'7 days'` → **`'4 hours'`** (an outage
is not a content defect — the bug is `bugs_open/236`'s 522 half, where a finished site
served an error page to every visitor indefinitely and nothing noticed). Tick 300s, own
`concurrency_group='site-availability'` so the `bugs_open/048` in-memory head-of-queue
class cannot couple it to your three content rotations. **No schema change**, and your
three rows are untouched.

**What this does to your watchdog, precisely.** `site-discovery-staleness-check` keys on
the three content agents' stamps, so it neither breaks nor covers the new one: an
availability rotation that silently stopped would be invisible to it. **Whether to extend
its coverage query to a fourth agent_type is your call, not mine** — I have not edited
your CronJob or your migration.

**The thing worth knowing if you re-measure cadence.** Your steady state was ~9
runs/day fleet-wide; this adds ~126 lightweight orchestrations/day (21 sites × 6 probes,
no LLM steps, no spawns). If a future census of `site_discovery_rotation` reports
"discovery is running hot", **split on `agent_type` before concluding anything** — the
availability rows will dominate the count and mean something different.

Lane docs: `docs024_key_docs_latest/bugfix_236_site_availability/`. Code committed
`4a5d77004` (check + tests + register IMP-053); config held as
`sql_for_agents/372_site_availability_driver_HOLD.sql` until the chassis rolls.
