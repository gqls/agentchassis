# HANDOFF — bugfix 203 phantom-CTA cleanup — continue here (supersedes 08-07)

**Written 2026-08-09.** Evidence for everything below: `NOTES_phantom_cta_cleanup.md`'s
2026-08-08/09 entries. Milestone read-out (plain prose): `SUMMARY_2026-08-09_phantom_cta_cleanup.md`.
The 08-06 and 08-07 handoffs are superseded in full — their "cannot be finished from the
outside, fix the resolver" conclusion is exactly what happened since.

## State in one line

**The resolver capability is built, tested, calibrated against the live fleet, and
resubmitted for review after an unrelated fleet-wide outage ate the first attempt.**
Verdict not yet read. **Start here — read it before doing anything else:**

```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='258e4ed7-55a2-4280-a919-2713363c8b89' AND kind='council_report'
ORDER BY created_at;
```

If empty, it's still running — check `orchestration_states` for orchestration
`e1c497e0-2be0-4a1a-821c-446166404451` (`current_step`/`status`). Give it time before
re-firing; a missing row is not a lost dispatch (see the fixloop runbook on this).

## What shipped (commits, in order)

- `bd6e3320c` — the fix itself: `datahelpers.LabelTokens`/`BestLabelMatch`/
  `NewLabelMatchCandidate` (extracted from `check_misdirected_cta.go`, behaviour-preserving),
  wired into `resolve_internal_links_action.go`'s `setCTAField` (write-time: match a page's
  *currently published* label to a real candidate before falling back to the old positional
  pick) and `rerender_page_sections_action.go`'s `applyCTARecompute` (the repair path
  `check_misdirected_cta`'s own `cta_links_stale` item triggers — which previously could NOT
  fix the exact defect it exists to fix; see NOTES for why). Six new tests, all
  mutation-proven.
- `465e45531` — a gofmt fix to the previous commit (caught by the commit-scope hook, not by
  `go build`/`vet`/`test` — all three are gofmt-blind). No behaviour change.
- `1ab9792e5` — docs: the NOTES entry with the full calibration methodology, the false
  positive found and fixed (interrogatives weren't in the stopword list), and
  `CALIBRATION_2026-08-08_label_match_report.txt` (898 lines, the real shipping matcher run
  against the live fleet, kept as evidence — not a rebuildable artefact, the `cmd/ctacalibrate`
  harness that produced it was deliberately deleted before commit per its own header comment).

Both commits carry `Council-Submitted: 258e4ed7-55a2-4280-a919-2713363c8b89`. **Per the
standing rule: do not write `Council-Reviewed:` retroactively even once approved** —
forward-only forbids the amend, and `098`'s coverage report resolves the correlation and
credits these commits automatically once the verdict lands approved. Nothing to do there
but read the verdict.

## Do not redo these

- **The extraction, the wiring, the tests, the mutation-proofs.** All done, all passing.
  `go test ./platform/orchestration/...` is clean except two **independently confirmed
  pre-existing failures at committed HEAD**, unrelated to any file this lane touches
  (`TestValidDocSubjectTypes_LockstepWithMigrationCheck`, `TestEveryCheckProducedItemTypeIsClassified`
  — verified via `git stash`, they fail identically with none of this lane's changes present).
  Don't spend a session re-confirming this; re-run the specific test names above if you want
  the receipt.
- **The calibration.** Already run against the real shipping function over live data, not a
  SQL approximation, using a `kubectl port-forward` to `postgres-clients` + `go run` importing
  the actual `datahelpers` package — that's the right method if you ever need to re-run it
  (e.g. after a stopword-list change), but the numbers as of 2026-08-08 stand: 1,251 CTAs
  examined, 634 matched, 315 newly resolved (no prior link to override), 162 would override
  an existing one (spot-checked, dominated by clear improvements).
- **The safe-subset re-check from 2026-08-08 (earlier the same day).** Already established:
  of the original 13-instance census, 3 self-healed naturally; what's left (2 real-tool-CTA
  pages, 4 parked "Get Started" heroes) has NO safe bare-rerender option — the tool CTAs would
  lose an achievable correct link, the heroes carry ~237 commits of unrelated rerender blast
  radius. Don't re-derive this; it's in NOTES with the exact query and reasoning.

## The credit-outage detour, so you don't re-diagnose it

The first submission (2026-08-08 evening) reached `orchestration_states.status='COMPLETED'`
with `current_step='complete_invalid'` — **read `collected_data->'__step_error'`, not
`error`** (a `COMPLETED`/`error IS NULL` row can still have failed a step; this is an
already-known landmine shape). It was a genuine Anthropic API 400 ("credit balance too low"),
fleet-wide, independently hit and documented the same evening by the `finetuning_uk_service`
lane (`bc6c99cff`) — not a defect in this submission. Confirmed restored before resubmitting
(`max(created_at) FROM llm_call_log WHERE provider='anthropic' AND success=true` moved to
minutes-before-resubmit, with a run of consecutive successes) and resubmitted on the
**same trail** via `RESUBMIT_CORR=258e4ed7-55a2-4280-a919-2713363c8b89` against the original,
unedited submission JSON — do not rebuild it from scratch if a third attempt is ever needed;
the file's content is reproduced in `git show bd6e3320c` and `1ab9792e5`'s trail if the
scratch copy is gone.

## Next, in order

### 1. Read the verdict (see "State in one line" above)

- **APPROVED**: nothing to commit (see above). Move to step 2.
- **REVISE**: disposition every objection explicitly, resubmit with a **fresh**
  `RESUBMIT_CORR=258e4ed7-...` against the same trail so it accumulates. The submission
  itself already flagged the one question most likely to draw an objection — whether this
  needs full architecture review rather than a normal round, given it's a guarantee change on
  a shared mechanism (CTA resolution: "some real page" → "a page matched to what the button
  claims"). If the architecture seat says `needs_rfc`, write the RFC — do not resubmit the
  same plan with better measurements per the owner's 2026-07-28 ruling on exactly this
  situation (see CLAUDE.md).
- **REJECTED**: read the guardian's notes (`SELECT body FROM doc_notes WHERE categories ?
  'council-gate' ORDER BY created_at DESC LIMIT 1`) — it names the safest contained
  alternative. Don't resubmit the same plan; the veto is about scope/soundness, not
  measurements.

### 2. Once approved: build, roll, verify at the pod — this fix is INERT until then

Go changes don't do anything live until an image is built from committed HEAD and rolled.
`make build-agent-chassis`, bump `IMAGE_TAG`, roll, then verify against the running pod, not
git and not the tag:
```bash
for P in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[*].metadata.name}'); do
  kubectl -n ai-persona-system exec "$P" -- sh -c \
    'strings /app/agent-chassis | grep -c BestLabelMatch; strings /app/agent-chassis | grep -c loadExistingSectionContentData'
done
```
Both should be non-zero, both replicas, same exec (a positive control proves the pipeline,
never your spelling — see the standing landmine on this).

**Then re-check whether the two-real-tool-CTA-pages / four-parked-heroes set from the earlier
cleanup pass can NOW be safely fixed differently.** They were unsafe specifically because the
only available repair routes (bare rerender, `edit_live`) couldn't preserve/re-aim the correct
link. Once the resolver itself does the matching, a **full rebuild** of those pages (which
re-runs the resolver unconditionally, per `bugs_closed/023`'s addendum) may now produce the
right link directly — worth testing on ONE page as a fresh canary before assuming it for all
six, the same discipline this lane used throughout. Do not assume; measure.

### 3. The two named follow-ons, not started

- **Writer-prompt coordination**: `cta_target_title` still never reaches the content-writer's
  LLM prompt (traced in full, see NOTES 2026-08-08 entry) — a real gap in the same file's own
  stated design intent, separate council footprint, not bundled into this round on purpose.
- **Persist-time self-healing arm**: a `repairSectionsBeforePersist` addition (alongside the
  existing `RepairContentDataLinks`, LNK-028) so a mismatched page corrects itself on its next
  ordinary save without a per-page dispatch. Step 2 above may make this partially moot for the
  six specific pages this lane already knows about — but not for the fleet in general.

## Tree hazards live right now (checked 2026-08-09)

This branch is heavily concurrent. Confirmed unrelated, do not investigate if seen again:
`check_decision_guards.go` / `decision_guard.go` (untracked, another lane, causes the
`TestEveryCheckProducedItemTypeIsClassified` failure above) and the doc-subject-type lockstep
test failure (already-committed, pre-existing, unrelated migration numbering issue per
`bugs_open/064`). Verify HEAD builds clean in isolation if in doubt:
`git archive HEAD | tar -x -C $(mktemp -d)` and build there, not in the shared working tree.
