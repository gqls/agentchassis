# HANDOFF 2026-08-04 (rev 5) — both rev-4 unknowns read; REVISE answered with code evidence and resubmitted (round 2 in flight); root-cause diagnosis genuinely UNVERIFIABLE, not silent

**This revision replaces rev 4** (same file, same date). Rev 4 left two things
unread: a council verdict and a root-cause diagnosis, both in flight. Both are
now read; this revision reports what they said and what was done about them.
Read `bugs_open/178`'s own file in full first — it is still the authoritative
account, seven updates deep now.

## What happened since rev 4, in order

1. **Read the council round-1 verdict**: `56f9a5a2-4d37-4114-9442-239861acd36e`
   came back **REVISE**, decided by a HIGH-severity gating objection from
   `bug_historian`, plus two MEDIUM objections (`architecture`,
   `prior_art_librarian`). 8 of 11 reviewers approved outright, several
   explicitly endorsing the narrow opt-in scoping.
2. **Checked all three objections against actual code/docs rather than taking
   them on trust — two turned out to be wrong as stated, one stands.**
   - `bug_historian` claimed `save_page_sections_action.go` and
     `rerender_page_sections_action.go` share the exact-slot_name-join failure
     this fix patches. **Checked, does not hold for either file**:
     `rerender_page_sections_action.go` resolves `component_id` before
     slot_name since `bugs_open/182` (immune to naming drift, observe-only log
     on disagreement); `save_page_sections_action.go` is delete+insert per
     build, not a stale-row join — its one slot_name lookup is the disclosed
     `bugs_open/058` locked-row exception. Full citations in `bugs_open/178`'s
     new update.
   - `prior_art_librarian` pointed at `datahelpers.SectionIdentityKey` as
     possible prior art. **Checked, wrong tool**: its own doc comment requires
     slot equality as *necessary* plus byte-identical content — built for
     exact-duplicate deletion (`bugs_open/156`), not cross-slot content
     matching. No existing mechanism solves this fix's actual question.
   - `architecture`'s objection (root cause deferred, no owner/timeline)
     **stands** — see point 3 below.
3. **Read the root-cause 090 diagnosis** (`167d2cc2-0b98-405c-a1d7-d54d80ed37c9`):
   it completed and returned **UNVERIFIABLE** (`stopped_by:
   "scope-not-narrowing"`, 5 iterations), not silence. It never obtained
   `plan_sections_action.go`'s actual Path 1/Path 2 body, and — the genuine
   dead end — no `page_component_history` row predates the overwrite event
   under investigation, so there is no forensic trail for what wrote the
   mismatched slot or when. Pulled the one pre-overwrite snapshot directly:
   real prose in its `content_data`, but no `slot_name` column to answer the
   actual question. **The mechanism is genuinely unknown, not merely
   unexamined** — this is new information, not the same open item restated.
4. **One claim inside the diagnosis artifact was wrong, and needed checking
   before it could be repeated**: it flagged `page_component_history`'s
   snapshot INSERT as "a real bug" for writing `pc.id` instead of
   `pc.component_id`. Checked the schema — `component_id` FKs to
   `page_components(id)`, not `content_components`, so `pc.id` is correct.
   The diagnosis loop misread its own evidence. Recorded so nobody downstream
   inherits it from the artifact alone.
5. **Resubmitted to the council, no code change** — the shipped fix
   (`4b3f9f89b`, live, pod-verified) is unchanged; only the rationale changed,
   answering all three round-1 objections with the code/doc evidence from
   points 2–3. `SUBMISSION_2026-08-04b_component_identity_drift_fallback_revise_response.json`,
   `RESUBMIT_CORR=56f9a5a2-4d37-4114-9442-239861acd36e` (same trail id, so
   round 2 accumulates against round 1 per this repo's practice). **Confirmed
   genuinely dispatched** — `council-gate-orchestrate-0804-2021` seen
   executing `review_editquality` seconds after submitting, and progressed to
   `review_prior_art` by 20:23 — not a repeat of the round-1 `kcat`
   silent-drop.

## State

The fix is still **live and pod-verified** (unchanged since rev 3/4). Council
round 2 is **in flight, unread** as of this revision. Root cause (candidate 3)
is now **investigated and still unknown** — a materially stronger claim than
"undiagnosed," worth distinguishing if this file is skimmed rather than read.
`bugs_open/178` stays OPEN.

## OPEN — in priority order

1. **Read the council round-2 verdict** before doing anything else:
   ```sql
   SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
   WHERE correlation_id='56f9a5a2-4d37-4114-9442-239861acd36e' AND kind='council_report'
   ORDER BY created_at;
   ```
   Two rows expected now (round 1 `revise` at 19:32:30, round 2 pending). If
   the second is APPROVED, commit a follow-up with trailer
   `Council-Reviewed: 56f9a5a2-4d37-4114-9442-239861acd36e` (forward-only — a
   new commit, not an amend of `4b3f9f89b`). If REVISE again, read the fresh
   objections (`SELECT body FROM doc_notes WHERE categories ? 'council-gate'
   ORDER BY created_at DESC LIMIT 1`) and judge whether they're answerable or
   this stays advisory-only — this repo's practice does not require chasing
   APPROVED indefinitely.
2. **Root cause (candidate 3) has no live lead.** The one witnessed occurrence
   no longer reproduces and its own forensic trail (`page_component_history`)
   doesn't predate the event in question. Nothing to re-investigate until a
   fresh occurrence is caught with an intact history trail — see point 4.
3. **The ambiguous case is still open by design** (two-or-more unmatched
   sections, or two-or-more candidate prose slots) — unknown severity, not yet
   observed.
4. **Watch for the fallback's first natural firing**, unchanged query from rev
   4, subject to ~24h `orchestration_states` retention:
   ```sql
   SELECT id, created_at, collected_data->'section_plan'->'edit_live_meta'
   FROM orchestration_states
   WHERE collected_data->'section_plan'->'edit_live_meta'->>'fallback_matched' = '1'
   ORDER BY created_at DESC LIMIT 10;
   ```
5. **Watch list, unchanged for four revisions**: the shrink guard doesn't fire
   on a whole-slot rename.

## Landmines specific to this lane (carry-forward + two additions)

- All landmines from the 2026-08-03/rev-2/rev-3/rev-4 handoffs still apply
  unchanged (shrink guard, dependency release rules, dispatch quiet-spell
  reading, `orchestration_states` retention, `output_field`
  shape-preservation, `build-dispatch-loop` IS scheduled, `bugs_open/087` vs
  `192` error-string collision, `content_components` has no site_id/page_id,
  "competing candidates" theories need a COUNT not a plausibility check, a
  council-trigger's printed correlation is not proof of dispatch).
- **A council objection's PLAUSIBILITY is not its CORRECTNESS — check it
  against the code before either conceding or resubmitting.** Round 1's
  gating objection named two specific files as sharing this fix's failure
  mode; reading those files showed neither actually does. A REVISE verdict is
  a claim like any other artifact — verify before acting on it, same as
  everything else this repo's memory insists on.
- **A diagnosis loop's own `NeededEvidence`/analysis text can assert a wrong
  claim with the same confidence as a verified one.** This run flagged a
  schema "bug" (`page_component_history.component_id` from `pc.id`) that
  wasn't one — the column correctly FKs to `page_components`, not
  `content_components`. `UNVERIFIABLE` on the main question did not stop it
  from stating a false side-claim in passing. Check anything a diagnosis
  artifact asserts before repeating it, exactly as with any other claim.

## Cold-start pointers

- `bugs_open/178`'s own file — still the authoritative account, seven updates
  deep now, including this session's council-answer and diagnosis-read work.
- `NOTES_work_item_routing_columns.md`'s 2026-08-04 tail (four entries now:
  192-lane verification, the measurement+fix, the deploy-verification
  session, and this session's REVISE-answer + diagnosis-read).
- `SUBMISSION_2026-08-04_component_identity_drift_fallback.json` (round 1) and
  `SUBMISSION_2026-08-04b_component_identity_drift_fallback_revise_response.json`
  (round 2, same correlation, rationale-only diff).
- Register entry PBP-028 (`docs026_concept_register/register/page-build-pipeline.md`)
  still has an open review question about matching by `slot_name` — still not
  folded in; worth adding the round-2 evidence (which files do and don't
  share the exposure) when someone next edits that entry.
- Commits: `08d0515f3` (178's original fix), `2b9d84072`/`71ecbb013` (192's fix
  + 178's correction), `4b3f9f89b` (this bug's fallback fix, rev 3, still the
  current code — rev 5 made no code changes, docs + a council resubmission
  only; check `git log` for this revision's own doc commit).
