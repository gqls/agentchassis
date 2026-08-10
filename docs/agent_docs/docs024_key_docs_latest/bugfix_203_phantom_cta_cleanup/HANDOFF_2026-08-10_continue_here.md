# HANDOFF — bugfix 203 phantom-CTA cleanup — continue here (supersedes 08-09)

**Written 2026-08-10.** Evidence for everything below: `NOTES_phantom_cta_cleanup.md`'s
2026-08-09 (cold-start continuation) and 2026-08-10 entries. The 08-09 handoff's own
open question ("read the verdict, then test row 3 as a canary") is answered in full —
both happened, and the canary surfaced a second, un-shipped bug that is now itself
mid-review. Read this file, not 08-09 — its "Next, in order" is stale.

## State in one line

**The original 203-follow-on resolver fix (`bd6e3320c`) is APPROVED, live, and
verified working on a real page.** That same live exercise found a second bug — a
scoring-priority defect in the matcher it ships — which is fixed (`3bc0486d7`) and
**awaiting its own council verdict**. **Start here:**

```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='6cb8c72b-0abc-4eb6-b4d2-4cbf01eed515' AND kind='council_report'
ORDER BY created_at;
```

If empty, it's still running — check `orchestration_states` for orchestration
`76b19b7e-3127-41a6-a1ad-b32efcad5f9c`. Give it ~30 minutes before assuming anything is
wrong (see CLAUDE.md's council-gate timing note); a missing row is latency, not a
dropped dispatch.

## What shipped since the 08-09 handoff (commits, in order)

- `3a41c5ea8`, `3bd57fc56` — NOTES entries only (verdict-read + row-3 canary write-up),
  no code.
- **`3bc0486d7`** — the second fix: `datahelpers.BestLabelMatch`
  (`platform/orchestration/datahelpers/label_match.go`) reordered so token-overlap
  count is compared BEFORE interactive-vs-hub category, with category only breaking a
  genuine tie. Two new tests, one (`TestBestLabelMatchOverlapBeatsCategory`)
  mutation-proven against the pre-fix comparator via `git stash`. **This commit carries
  NO council trailer at all** (committed before the submission existed — a process
  slip, logged in full at `WRONG_CALLS.md` 2026-08-10 and in NOTES). It will read as
  UNREVIEWED in `098`'s report forever, even once approved — this HANDOFF and NOTES are
  the only durable link between the commit and `SUBMISSION_CORR=6cb8c72b-0abc-4eb6-b4d2-4cbf01eed515`.
  **Do not try to retrofit a `Council-Reviewed:` trailer onto `3bc0486d7`** — there is
  nothing to amend it with (forward-only). If approved, just record it; don't chase the
  098 bookkeeping.
- `018923568` — NOTES + WRONG_CALLS entry for the above.

## What the row-3 canary actually proved (don't redo it)

Dispatched `robot-hands.com/how-to-specify-a-gripper`'s existing detector-created
`misdirected_cta` item (`fab86424-0078-469b-b355-76c6a625b67e`) by promoting it to
`triaged` and firing `build-dispatch-loop` — no new work item authored, the framework's
own suggestion already named the right target. Result, checked at three layers
(content_data, not just work-item status):

- **hero "Run MatchMatrix" → `/tools/matchmatrix/index.html`: CORRECT.** First live
  proof the original 203 follow-on's resolver fix does what it was built to do.
- **call-to-action's secondary_cta "Browse the Gripper Catalog" →
  `/tools/gripper-cycle-time-estimator/index.html`: WRONG.** Should have been
  `gripper-catalog-index` (`/gripper-catalog/index.html`, 2-token overlap) — lost to a
  1-token tool match because of the priority bug `3bc0486d7` now fixes.

**Do not re-run this canary.** Both outcomes are recorded with full evidence in NOTES
2026-08-09/10. The robot-hands.com page's `secondary_cta_url` is STILL wrong as of this
writing — the priority fix landing does not retroactively repair pages that already
rendered under the old code (see "Next" below).

## Do not redo these (carried forward from 08-09, still true)

- The original resolver extraction/wiring/tests/mutation-proofs and its calibration —
  see 08-09 handoff for detail, unchanged.
- Row 2's structural finding: `finetuning.uk/about`'s verified target
  (`/how-we-work.html`) is `page_type='content'`, which is outside BOTH the interactive
  and hub candidate pools by design — **no version of either fix can reach it**. Not a
  bug, a scope boundary. Don't re-diagnose it.

## Next, in order

### 1. Read the verdict (see "State in one line" above)

- **APPROVED**: nothing to commit — see the trailer note above, this commit stays
  bookkeeping-unreviewed by design-of-the-mistake. Move to step 2.
- **REVISE**: disposition every objection, resubmit with a fresh
  `RESUBMIT_CORR=6cb8c72b-...` against the same trail (this time, if any further edit is
  needed, put the trailer in the SAME commit that makes it — don't repeat the 08-10
  slip).
- **REJECTED**: read the guardian's notes
  (`SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1`).

### 2. Once approved: build, roll, verify at the pod

Same drill as the original fix — `make build-agent-chassis`, bump `IMAGE_TAG`, roll,
then positive-control grep both replicas for a symbol the fix's diff actually touches
(e.g. a literal from the new doc comment, or just re-run the two existing symbol checks
plus confirm the binary is NEWER than this fix's commit).

### 3. Re-run the row-3 canary's OWN page as the verification

Once the priority fix is live, re-dispatch (or wait for the automatic loop to reach)
`robot-hands.com/how-to-specify-a-gripper` again and confirm `secondary_cta_url`
resolves to `/gripper-catalog/index.html`. This is the cheapest possible verification —
the exact page and exact wrong value are already known, no new census needed.

### 4. Only then: decide on the remaining five held-back pages

2 leftover real-tool-CTA pages minus row 3 (now done) = **just robot-hands's own
secondary CTA, already covered by step 3** — re-check whether there's a genuinely
separate "row 2" successor once the fix is live (row 2 itself is out of reach per the
structural finding above, so this is really down to the 4 leopardessconsulting.co.uk
"Get Started" heroes). **Check each of those four's verified target's `page_type`
before dispatching any of them** — not done this session — since the row-2 lesson
means a `content`-type target will never resolve regardless of this fix.

### 5. The two named follow-ons, still not started (unchanged from 08-09)

- Writer-prompt coordination (`cta_target_title` never reaches the content-writer's LLM
  prompt).
- Persist-time self-healing arm (`repairSectionsBeforePersist`).

## Tree hazards live right now (checked 2026-08-10)

Same as 08-09: `check_decision_guards.go`/`decision_guard.go` untracked-and-unrelated,
doc-subject-type lockstep failure pre-existing. This session additionally observed a
large concurrent fleet release in flight (17 services' kustomization `newTag` bumps,
uncommitted, `agent-chassis` at `v1.0.1274`) — not this workstream's concern, but if
`IMAGE_TAG`/pod state looks inconsistent with what you expect, check whether another
session's release finished rather than assuming this workstream's build is broken.
