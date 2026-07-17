# RUNBOOK — Council Gate (advisory review service)

**Thread:** "fixloop council on every bugfix" (started 2026-07-17 from
`HANDOFF_2026-07-17_council_gate_thread.md`). Design:
`DESIGN_feature_builder_and_council_gate.md` §2. Turn record:
`NOTES_running_council_gate.md`.

## State: BUILT, NOT LIVE

All three advisory-mode components are built and verified as far as they can
be without applying anything. **Nothing has been applied to the database** —
owner ruling 2026-07-17: the gate launches only once more concept-register
stage-3 seats are live.

| # | Component | File | State |
|---|---|---|---|
| 1 | Submission wrapper + trigger | `097_TRIGGER_council_review_v1.sh` | built; validations dry-run tested (single-line payload proven — kcat trap) |
| 2 | Orchestrator seed | `0NN_council_gate.sql` | built, apply-ready, **synced to the 5-seat v8 roster** (19 steps); literal-balance verified; **NOT applied** |
| 3 | Visibility report | `098_REPORT_unreviewed_commits_v1.sh` | built; live-run 2026-07-17: 28 in-scope commits / 3 days, 0 reviewed |
| 4 | PR-mode (enforcement) | — | **not built** — owner's explicit go required (build order rule) |

## Owner decisions on record (collected 2026-07-17, this thread)

1. **Scope:** `platform/`, `internal/`, `pkg/`. Docs/site content never spend
   council credits (the 097 script refuses them client-side; `FORCE=1` overrides).
2. **Mode at launch:** advisory first (steps 1–3). PR-mode stays a later,
   separate owner call.
3. **Credit policy:** one council run per submission = per task/commit,
   matching the commit-per-task rule.
4. **Roster:** WAIT for more concept-register stage-3 seats before launch.
   **The seats arrived the same day**: the concept-register thread built and
   applied reuse-agent (v7) and guidelines-agent (v8) to the live
   fix-proposer — the council is now **5 reviewers**, and this gate's seed is
   synced to that roster. Whether 5 satisfies the ruling, or the launch also
   waits for the relevance filter (`DESIGN_relevance_filter.md` — needs a
   chassis Go change), is the owner's call — see launch checklist step 0.

**Flag → RESOLVED same day:** the v6-inherited `run_checks.check_fields`
omission (three advisory seats' checks solicited but never run) was flagged
by this thread and fixed by the concept-register thread as **v9, applied
live**. The gate seed matches. Their relevance-filter Go engine
(`select_review_panel` + council_decide abstention) is also now built and
committed (`37468ba65`) but **inert until a chassis image ships** — they are
holding that deploy to sequence with this thread; the gate seed notes the
coming lockstep change in its header.

## Launch checklist (when the owner says go)

0. Owner confirms the roster ruling is satisfied: the live council is now 5
   seats (edit-quality, bug-historian, reuse-agent, guidelines, guardian) —
   launch on these, or wait for the relevance filter / further seats?
1. Roster lockstep: any new seat lands in **both** `0NN_fix_proposer` (v9+)
   **and** `0NN_council_gate.sql` in the same migration — the two files'
   reviewer steps are deliberately name-matched; letting them drift is the
   dedup-index/Go-list class of failure. (Synced to v8 as of 2026-07-17.)
2. Apply `0NN_council_gate.sql` to clients_db
   (`kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`)
   — needs the owner to name the target, per the standing permission gate.
   No image build: every action is registered and live (≥ v1.0.1127).
3. Run the post-apply verification queries at the bottom of the seed file
   (17 steps; `review_fields` lists all three reviewers).
4. Smoke-run: submit a small real change via 097, watch
   `orchestration_state_audit`, read the verdict note
   (`SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1`).
5. Announce in the digest/docs: threads submit before committing; approved
   commits carry the trailer.

## How a thread uses the gate (once live)

1. Write a submission JSON (schema in the 097 header: `rationale`,
   `submitter`, and a fix_plan-shaped `plan` — ≤8 edits, ≤64KB, real diff
   hunks in `sketch`, evidence quotes in `grounded_in`).
2. `./097_TRIGGER_council_review_v1.sh submission.json` → save the printed
   `SUBMISSION_CORR`.
3. Verdicts: **APPROVED** → commit with trailer `Council-Reviewed: <corr>`.
   **REVISE** → objections + the reviewers' checks come back answered
   (doc_note + council_report); revise, resubmit with
   `RESUBMIT_CORR=<corr>` so the trail accumulates. **REJECTED** (guardian
   veto) → do not ship as-is; the guardian's notes name the safest contained
   alternative.
4. `./098_REPORT_unreviewed_commits_v1.sh [days]` shows fleet coverage;
   `PERSIST=1` files it to doc_notes (categories digest+council-gate).

## Honest limits (advisory mode)

- The gate cannot intercept a hand-commit to the shared branch; the 098
  report makes the gap visible, nothing more. The first live run of 098
  (3-day window) found **28 in-scope commits, 0 reviewed** — that number is
  the baseline the gate is judged against.
- A council run costs credits and minutes. Submit per coherent task, not per
  iteration; PR-cadence batching arrives only with PR-mode.
- The trailer is self-declared. MISMATCH (trailer without a green report) is
  bucketed separately by 098 precisely so a false claim of review is visible.

## Cross-links

- `HANDOFF_2026-07-17_council_gate_thread.md` — this thread's cold-start.
- `DESIGN_feature_builder_and_council_gate.md` §2 — the design executed here.
- `../../docs026_concept_register/` — stage 3 = the seat roster (launch gate);
  `PILOT_bug_historian_reviewer.md` is the proven seat-adding pattern;
  `PILOT_reuse_agent_reviewer.md` is seat #4 awaiting sign-off.
- Multi-session coordination workstream — commit-per-task + build-from-ref;
  the `Council-Reviewed:` trailer composes with those rules.
