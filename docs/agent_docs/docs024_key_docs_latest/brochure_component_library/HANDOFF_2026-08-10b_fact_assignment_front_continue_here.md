# HANDOFF 2026-08-10b — fact-assignment front (bug 151 / RFC_016): cold-start for a fresh chat

**Supersedes `HANDOFF_2026-08-10_…`.** Written ~23:00 BST, after the Slice B round
came back **APPROVED**, all three seeds were **applied**, and chassis **v1.0.1283**
(carrying this front's three new durable detections) was rolled and
**artefact-verified**. The building is finished. **What remains is measurement
only** — §4 is the whole of the next session's job.

**This is ONE OF TWO fronts in this directory. Do not confuse them.**
- **This file = the fact-assignment front** (bug 151 candidates 1 + 1b, RFC_016).
- **`HANDOFF_2026-08-09_sweep_front_continue_here.md` = the fundamentallyai sweep
  front** — different live thread, same site. **Read it BEFORE dispatching
  anything at the site** (§4 step 1 is exactly this).

Site id, needed everywhere: **`199733a8-ac9c-4c30-b2ce-65ecdac6f3bd`**.

## 1. Verified live state (all artefact/row-verified 2026-08-10 evening)

| thing | state |
|---|---|
| Chassis | **v1.0.1283**, both replicas (`agent-chassis-696d88b4c7-95mgb`, `-wnbs8`), started 21:43Z |
| The three detections | **LIVE + ARTEFACT-VERIFIED**: `FACT_ASSIGNMENT_ABSENT`, `RECOMPOSE_INTENT_NOT_REALISED`, `PLAN_PAGE_MERGE_LOSSY` each →1 on both replicas; NEG spelling →0; CTRL pre-existing literal →1 |
| Slice B council round | **APPROVED**, corr `a06ff850-aff6-4ed0-8e0a-93d57b0cbc45` (resubmission, same corr as the REVISE — the trail accumulates). Verdict pinned: `COUNCIL_VERDICT_slice_b_2026-08-10b_approved.json`. 3 advisory objections, all dispositioned same evening (NOTES, "night" entry) |
| Seed 362 (planner sees realised sections) | **APPLIED + row-verified** (prompt 18,738→19,685 B; `\| sections: ` listing + re-emit instruction both present). `agent_definitions_bak_362` exists |
| Seed 328 (`section_facts` wiring) | **APPLIED + row-verified** (`section_facts=spec_sections.section_facts` on page-build-handler's `plan_sections`). Own pod-check passed pre-apply. `bak_328` exists |
| Seed 330 (writer prompt v4) | **APPLIED + row-verified** (three-way branch at the NESTED path `{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}`). Compliance read done + owner-countersigned (`COMPLIANCE_READ_2026-08-10_writer_prompt_v4.md`). `bak_330` exists |
| Migration ledger | all three `--record-only`'d under their new (un-`_HOLD`ed) names |
| Owner rulings | all four §6 decisions ruled + executed: `DECISIONS_2026-08-10_owner_rulings_after_relook.md` |
| Consumption of assignments | **STILL ZERO — measuring this is the next session's job.** Nothing has replanned/rebuilt since the seeds applied |

## 2. The mechanism as it now stands (read this before measuring)

Since 362+328+330 and v1.0.1283, a replan of a site with an `evidence_base` works
like this — every link verified this session:

1. **Planner** (`build-site-planner`) sees each built page's realised section
   list and re-emits it verbatim, attaching per-section `facts` (mandatory since
   seed 333; `[]` = deliberately factless). Composition compares EQUAL by NAME
   (`sameSectionList` fixed), so **nothing is restored and no carry is needed** —
   the 1b (ii) carry is the safety net, not the mechanism.
2. **Validate** records disobedience durably instead of hiding it:
   `FACT_CARRY_UNMATCHED_SECTION` (facts scoped to a section the page lacks),
   `FACT_ASSIGNMENT_ABSENT` (object entry with no usable `facts` key — only on
   restored pages, i.e. only when the carry runs), `RECOMPOSE_INTENT_NOT_REALISED`
   (a `recompose_pages` page proposed verbatim or absent).
3. **page-build-handler**'s `plan_sections` (the serving agent — proven by config
   handshake + 30/30 live census + in-place resolver transit) stamps
   `facts_scoped`/`assigned_writer_block` per ready section via the 328 key.
4. **Writer** iterates the handler's plan (its own `plan_sections` is a no-caller
   fallback) and v4 branches per section: assigned facts only / explicit
   factless / site-wide fallback.
5. **Plan write**: colliding pages dedupe; a composed-vs-composed merge records
   `PLAN_PAGE_MERGE_LOSSY` with both full section lists.

## 3. What happened today (compressed; NOTES has the evidence)

Re-look corrected three of four decision recommendations → owner ruled all four
→ compliance read performed at owner direction + countersigned → REVISE answered
(census settled three ways; §3.5 absent-facts hole fixed; two ruled detections
built; seeds guarded with induced DO/RAISE; mutations caught by name, incl. one
build-failure mutant redone) → resubmitted BEFORE committing → **APPROVED** →
advisory objections closed (LogActionFindings refactor — the reuse seat was
right; counter-consumer grep zero) → seeds applied 362→328→330 → landmine armed
→ owner rolled v1.0.1283 → artefact-verified. One outbound same-file passenger
(my v3_site_actions.go refactor rode in `fba05b83a`, another lane's commit) —
recorded, nothing lost.

## 4. THE NEXT SESSION'S JOB — replan, rebuild, census. In this order

1. **Read the sweep front's handoff first**
   (`HANDOFF_2026-08-09_sweep_front_continue_here.md`) and check its lane's
   recent commits (`git log --oneline -20 -- <this directory>`). A replan on
   this site IS a build dispatch and, until the quiet mode is fixed, still a
   phantom-page generator. If the sweep front has work in flight at the site,
   coordinate in its handoff file before dispatching.
2. **Check the queue before dispatching** (the standing rule):
   `SELECT summary, status FROM site_work_items WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND status NOT IN ('complete','cancelled','rejected');`
3. **Dispatch the replan** (mechanism in `RUNBOOK_brochure_component_library.md`;
   it is a `needs_site_plan` work item). Do NOT set `recompose_pages` — this
   measurement wants preservation, not redesign. Respect the ~300s
   no-dispatch-after-pod-restart rule (pods started 21:43Z on 08-10, so only
   relevant if another roll happens).
4. **Read the plan the replan writes** — this alone proves/refutes 362:
   - Composition preserved? Every built page's `site_plan_sections` list should
     be UNCHANGED in membership+order vs the realised pages.
   - Assignments landed? `SELECT page_name, ordering, assigned_fact_ids FROM
     site_plan_sections WHERE plan_id=<new> AND assigned_fact_ids IS NOT NULL;`
     — non-empty on engaged pages is candidate 1 consuming for the first time.
   - Disobedience? `SELECT error_code, count(*) FROM agent_error_log WHERE
     created_at > <dispatch time> AND error_code IN
     ('FACT_CARRY_UNMATCHED_SECTION','FACT_ASSIGNMENT_ABSENT') GROUP BY 1;`
     Expect ~0 carries (nothing should be restored now) — a burst of carries
     means the planner is NOT re-emitting realised lists, i.e. 362 failed.
5. **Rebuild the flagged pages** (the sweep front's flagged set; coordinate per
   step 1) and read the WRITER's outputs: scoped sections must state ONLY their
   assigned facts; factless sections no business figures; **the overlap pairs
   must fall on engaged pages**.
6. **The disconfirming half — do not skip it:** the five fact-blind sites (no
   `evidence_base`) must NOT move. Their writer prompts fall through to the
   unscoped arm; any change on them is a regression this round caused.
   [Site list: RFC_016 / the 08-08 slice A notes — re-derive with
   `SELECT s.domain FROM sites s LEFT JOIN site_specs ss ON ss.site_id=s.id AND
   ss.aspect='evidence_base' WHERE ss.id IS NULL AND <live-sites filter>;`
   schema-first — check `\d site_specs` before trusting this sketch.]
7. **Close the loop in the docs**: census results into NOTES + README; if the
   round proves out, bugs_open/151's candidate-1 arc is measurable-done — but
   remember the owner keeps finished bugs in `bugs_open/` (ruling 08-06).

## 5. Traps (updated; the standing ones from the 08-10 handoff still hold)

- **`orchestration_states` prunes at ~24h** — the 30/30 serving-path census and
  any error census must be re-run fresh, never quoted from this file as live.
- **Verdicts by CORRELATION, never `doc_notes … LIMIT 1`.**
- **A roll is not evidence** — v1.0.1283 is verified for THIS front's literals;
  re-verify after any further roll (pod names change: they did twice today).
- **`recompose_pages` + seed 362**: redesign intent must be in BOTH the field
  and the briefing, or it silently no-ops — LANDMINES entry (2026-08-10). The
  durable tell (`RECOMPOSE_INTENT_NOT_REALISED`) is live in v1.0.1283.
- **The working tree is shared and was NOT clean today** — other sessions had
  WIP across `platform/orchestration/actions` all evening (one swept my refactor
  into their commit). Build/test proofs via `git archive HEAD` overlay; check
  `git status` fresh; expect passengers in both directions.
- **`who-owns` reads commits** — a session mid-work is invisible; grep live
  transcripts if routing matters.
- Rollbacks if needed: `agent_definitions_bak_362` / `_bak_328` / `_bak_330`
  (restore recipe in each seed's tail). Data-half rollback un-does config only —
  the Go in v1.0.1283 is inert without the config keys, by design.

## 6. Owner items outstanding (not blockers)

1. **`features_open/012` needs a target date** — the council's bug_historian
   asked that the recompose field-based fix not drift into indefinite deferral.
2. **Standing 215 revisit trigger**: `SELECT count(*) FROM agent_error_log WHERE
   error_code='PLAN_PAGE_MERGE_LOSSY';` — richer-wins is ratified while this
   stays 0; a non-zero is the agreed cue to look again.
3. The compliance read logged three pre-existing follow-ups (invented
   commitments clause; edit-mode legacy claims; testimonial trade-dress) —
   recorded in `COMPLIANCE_READ_2026-08-10_writer_prompt_v4.md` §3.4–3.6,
   unscheduled.

## 7. Commit trail (this front, today)

`2c4278d6b` owner rulings + compliance read + 012 registration · `4e3c96e64`
the REVISE answered (code + guards + resubmission) · `1f0cd3bcd` gofmt ·
`fba05b83a` (ANOTHER LANE's commit — carries my LogActionFindings refactor of
`v3_site_actions.go` as an undeclared passenger, declared in NOTES instead) ·
`177454a87` APPROVED: seeds applied, verdict pinned, landmine armed ·
`d77c58683` the passenger recorded · this file's commit.
