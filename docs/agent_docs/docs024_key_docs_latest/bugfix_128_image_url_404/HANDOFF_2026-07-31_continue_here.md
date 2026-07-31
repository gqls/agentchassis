# HANDOFF 2026-07-31 — bugfix 128, cold start

**The bug is DONE: fixed, live on v1.0.1219, pod-verified with a negative control,
proven on two live sites, closed and moved to `bugs_closed/128`.** This doc exists for
the one thing still owed and the three things deliberately left.

Read in this order if you are new: `bugs_closed/128_HANDOFF_…md` (the case, with the
close section at the foot) → `NOTES_image_url_404.md` (what actually happened, including
three missteps) → `RUNBOOK_image_url_404.md` (every command, with its gotcha).

## What shipped

`image_url_404` compared asset **purpose names** against file **paths**, so a site owning
one `hero` asset could not be told its `/assets/images/hero.jpg` was a 404. Measured live
over all 127 rendered image paths on 13 sites, HTTP as ground truth: the shipped
predicate reported **21 working images and missed 6 live 404s**; the fix reports 1 and
misses 0. It now resolves through `storage.DeployedWebPath(asset_key, purpose)` — the
helper five writers and the deployer already share — which makes the check the exact
inverse of the render-time resolver. It also scans `site_components` (chrome, on every
page), flags `<img src="">`, keys dedup by filename-with-extension, and is **flag-only**:
its recognised-purpose routing branch was an exact duplicate of
`check_placeholder_image_in_use` and was deleted.

Commits: `beff42809` (fix + 10 tests), `6d3992213` (the drift audit), `b51e4879d`,
`e2b6a7dfd` (close), `fe35b322e`, `76643e8c0` (contribution to the filing lane).

## OWED — resubmit to the council after 2026-08-01 00:00 UTC

**The verdict of record is round 1's REVISE.** Round 2 answers both objections but never
sat: it terminated `complete_invalid`, which looks like a schema refusal and is not —
`__step_error` reads *"You have reached your specified API usage limits. You will regain
access on 2026-08-01 at 00:00 UTC"* (`bugs_open/130`'s cap).

Both commits carry `Council-Submitted:`, never `Council-Reviewed:` — that trailer asserts
nothing, so it cannot be a false claim, and `098` credits them automatically once a
verdict lands. **Do not write `Council-Reviewed:` on this correlation until you have read
an approved verdict.**

The round-2 submission file was written in this session's scratchpad, which will not
survive. **Rebuild it from `NOTES_image_url_404.md` § "Council round 1"** — that section
carries the full text of both objections and the five-query audit that answers the gating
one, which is all the submission's rationale contained.

```bash
RESUBMIT_CORR=99dca96a-413a-4bcb-b278-9577f920786d \
  ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <file>.json
# find the run by PAYLOAD, never by the printed id:
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = '99dca96a-413a-4bcb-b278-9577f920786d';
```

The two objections, in one line each, so you can judge whether they are answered:
**bug_historian (HIGH, gating)** — does adopting `DeployedWebPath` inherit a
shared-mechanism defect patched at one call site? Answered by an audit that is now a
comment block on `loadDeployedAssetPaths`. **editquality (medium ×3)** — bundling;
answered by grounding each edit in the diagnosis text, since all four are in the bug file
and three are causally forced by the core fix. Note `reuse_agent` **approved** the very
edit `editquality` objected to.

## LEFT DELIBERATELY — three items, none of them defects in the fix

1. **Nine stale `detected` rows** under the old extension-less item key. Five
   `finetuning.uk` `case-study-*` are still true 404s and will re-file under the new key.
   **Four are the OLD predicate's false positives and can be cancelled outright**:
   `fundamentallyai.com:brand-illustration` and three `robot-hands.com:content-hero-tool-*`
   (all serve 200). Not touched because `bugs_open/083`'s lane is actively working the
   `detected` population.
2. **The `storage.DeployedWebPath` durable fix.** A *new* purpose containing an underscore
   stored with `asset_key == purpose` would still be mis-rendered. The audit proves that
   set is empty today across all 267 active asset rows; it cannot prove it for assets that
   do not exist yet. The fix belongs in `platform/storage` with five other consumers —
   architecture-scope, the shape `bugs_closed/124` was vetoed for. Named in the round-2
   risks and in `LANDMINES.md`; **do not smuggle it into a bug patch.**
3. **The HTTP half.** Still deliberately unbuilt: `verifier_coverage_test.go:171` records
   the standing objection to an outbound probe on the completion path, and this fix keeps
   that promise. `features_open/026` Phase 3 remains its venue. ⚠ If it is ever built,
   **it must not probe an empty `src`** — that resolves against the current document and
   returns **200**, so the probe would confirm a broken image as healthy.

## Traps this lane paid for, so you do not

- **`who-owns.py` reads COMMITS**, so a filing workstream and an owning one look
  identical. Grep the live `.jsonl` transcripts for **code symbols**, never the bug number.
- **`complete_invalid` on a council run does not mean your JSON was wrong.** Read
  `collected_data->'__step_error'` — a FAILED step shows as COMPLETED with `error` NULL.
- **`curl` returning `000` is a connection error, not a status.** Re-probe every non-200
  before tallying; five of 127 needed it here and all five were 200.
- **Grep `LANDMINES.md` for your SYMBOLS yourself.** The `SessionStart` hook only matches
  entries against files already **dirty** in the tree. Two landmines written hours earlier
  by other lanes described the exact two traps this lane then hit.
- **The shared tree may not compile** — another session's in-flight edit broke this
  package. Test against `git archive HEAD` plus your own files, in `~/.cache`, **not
  `/tmp`** (16G tmpfs shared by ~30 sessions; it hit 100% during this work and the
  resulting error looks like a command failure when it is an output-capture failure).
