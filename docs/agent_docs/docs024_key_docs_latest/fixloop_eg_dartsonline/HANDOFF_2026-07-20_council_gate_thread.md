# HANDOFF — continue the COUNCIL GATE thread

**Filed 2026-07-20 evening** from "fixloop council on every bugfix". Cold-start
for a NEW chat. Supersedes `HANDOFF_2026-07-17_council_gate_thread.md`, which
described a gate that was not yet built. **Read `RUNBOOK_council_gate.md` first
for the commands**, then this for state and what is open.

## Standing docs (CLAUDE.md's five)

| doc | file |
|---|---|
| PLAN | `PLAN_2026-07-17_council_gate.md` |
| RUNBOOK | `RUNBOOK_council_gate.md` |
| NOTES | `NOTES_running_council_gate.md` |
| owner prose log | `README_where_we_are.md` (shared with the fix-loop + feature-builder threads in this directory; **append only**) |
| SUMMARY | `SUMMARY_2026-07-19_council_gate.md` (newest; write a NEW file only at a real milestone) |

## Where it stands (all verified live, 2026-07-20)

The gate is **LIVE, advisory, and in real use by other threads**. It is the fix
loop's reviewer council opened as a service: any session submits a change, gets
a verdict, and commits with a `Council-Reviewed: <SUBMISSION_CORR>` trailer.

- **Roster: 16 seats**, both councils in sync (`099_SYNC_gate_roster.py` dry run:
  zero drift). 2 always-on (edit-quality, guardian); 14 relevance-gated. A real
  submission on 2026-07-20 drew **10 of 16**.
- **Adoption is real and not prompted**: 32 gate verdict notes on 2026-07-20
  alone, from several threads (imagery, feature-builder, bugfix threads).
- **Chassis v1.0.1140** — pod-verified. Carries the `bugs_open/036` fix
  (`objectionEdit`, 7 symbol hits: one seat's type slip no longer voids a whole
  paid round) and the `019` truncation fix. Abstention for skipped seats intact.
- **Owner rulings** (unchanged): scope `platform/`+`internal/`+`pkg/`; advisory
  first; one run per coherent task; PR-mode is **step 4 and needs an explicit go**.

## What is open

1. **`bugs_open/016` second finding — applied, still UNEXERCISED.** The reviser
   now reads the whole `council_report` artifact instead of per-seat prompt
   sections (`load_council_reviews`, patch script
   `PATCH_fix_proposer_016b_reviser_reads_council_report.py`, applied
   2026-07-18 16:21:44Z). **No `fix-proposer` run has started since**, so it has
   never executed. Reliable probe — `load_council_reviews` exists on
   `fix-proposer` **only**:
   `SELECT count(*) FROM orchestration_state_audit WHERE new_current_step='load_council_reviews';`
   → 0 means still unproven. Two probes that LIE, both burned already:
   `llm_call_log.agent_type='fix-proposer'` never matches (rows are stamped
   `generic`), and `persist_plan` is shared with `feature-designer` and
   `experience-planner`.
2. **The digest gate-verdicts change, mid-flight on correlation
   `bd12762a-5b10-416b-a70d-90ee3067ce7d`.** Round 2 verdict: **REVISE** (10
   seats, 2026-07-20 19:28). Round 1's partition assumption was refuted by the
   council's own live queries (every `council_report` carries
   `source_agent='generic'`); round 2 replaced it with a verified discriminator
   (does the correlation have a `kind='bundle'` artifact → diagnosis-backed).
   Round 2's objections, if someone takes this forward: the `body::jsonb` cast is
   unguarded (**three seats caught this independently** — a non-JSON body would
   error the whole gatherer); the edit declares one file but changes
   `fixloop_digest_test.go` too; error handling from the gatherer is unstated;
   and the travelling-docs convention (a NOTES entry for the subject) is
   unaddressed. Submission JSON is in the session scratchpad — rewrite it rather
   than hunt for it.
3. **PR-mode** — not built, owner's call, and the coverage number is the input.
4. **Approved verdicts are deletable** — an open question in the PLAN. A
   `council_report` is what a commit trailer points at, and one was destroyed
   mid-day on 2026-07-18 by a documented practice (since retired). Nothing makes
   an approved verdict immutable.

## Traps this thread paid for (do not re-pay)

- **Dispatch latency is ~30 minutes, not ~2.** Measured 2026-07-20: publish →
  run start = 29 minutes under normal load. A missing orchestration row is
  latency, not a drop. I called it a drop at 13 minutes, retried, and spent ~25
  minutes on publish-path forensics that found nothing — logged in
  `WRONG_CALLS.md`. The messages were on the topic the whole time (`kcat -C
  -o -400` proves publish independently of the consumer).
- **Status is `COMPLETED`, uppercase.** A `case "$R" in *completed*)` poll loop
  never breaks; that cost 10 minutes.
- **`kubectl exec` needs `-i` only when SQL arrives on stdin**, never with `-c`.
  Both directions have bitten: with `-i` inside a read-loop it eats the loop's
  stdin (the 098 report silently counted 4 of 41 commits); without `-i` a heredoc
  reaches nothing and psql exits 0 having done nothing. **Verify the write.**
- **Never hand-patch the gate's roster** — seat `fix-proposer`, then run
  `099_SYNC_gate_roster.py` (`--apply` snapshots first).
- **Reviewers judge the `sketch`**, not your prose. On a resubmit, update the
  sketches or you will draw confident objections about code you already changed.

## The honest limit, worth carrying into any PR-mode discussion

Another thread recorded in `WRONG_CALLS.md` that a submission whose headline
proposal was *unimplementable* passed **two full council rounds** and reached an
implementing thread, which hit the wall immediately. Plan-stage review cannot
refute what only implementation can. That is an argument for handing a plan over
sooner rather than polishing it through another round — and a caution against
treating a green verdict as proof of feasibility.

## Cross-links

- `DESIGN_feature_builder_and_council_gate.md` §2 — the design this executes.
- `bugs_open/016_HANDOFF_2026-07-18_council_revise_prompts_drop_reviewer_output.md`
  — both findings, what is fixed, what is unproven.
- `docs026_concept_register/` — where seats come from; `PILOT_*_reviewer.md` is
  the pattern for adding one.
- `WRONG_CALLS.md` — this thread's entry (queue latency) and the plan-stage-review
  limit above.
