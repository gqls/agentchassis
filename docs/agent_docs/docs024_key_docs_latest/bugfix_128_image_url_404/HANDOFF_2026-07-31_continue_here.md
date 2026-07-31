# HANDOFF 2026-07-31 — bugfix 128, cold start

**The bug is DONE: fixed, live on v1.0.1219, pod-verified with a negative control,
proven on two live sites, council-APPROVED at round 2, and closed to `bugs_closed/128`.**
Nothing is owed. This doc exists for the three things deliberately left, and the traps
this lane paid for.

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

## DONE — council APPROVED at round 2, and its advisories discharged

**Nothing is owed to the council.** `SUBMISSION_CORR=99dca96a-413a-4bcb-b278-9577f920786d`,
run `06573962`: **12 approve / 3 object / 3 abstain, "approved with 3 advisory
objection(s) — none high-severity"**. `bug_historian` withdrew its own round-1 gating
objection in terms: *"a direct, evidenced answer to my own prior gating objection …
correctly declining architecture-scope creep."* The approval commit carries
`Council-Reviewed:`; the earlier ones carry `Council-Submitted:` and are credited by `098`
at report time.

The first resubmission attempt (round 2a) died on the API usage cap; the second, once the
cap was raised, was approved. **Trap worth keeping:** an attempt before that produced no
run at all and printed nothing I was grepping for, because a `cd` earlier in the same
compound command had made the trigger's relative path invalid. **Confirm a submission
landed by querying `orchestration_states` for the correlation — never by the trigger's
stdout**, which you may not have actually read.

The three advisories were checked rather than banked; the evidence is in
`bugs_closed/128` § "Council: APPROVED at round 2". The one that changed anything:
`bug_historian` objected that the durable `DeployedWebPath` fix was *"deferred as its own
item rather than filed"* — so it is now **`bugs_open/168`**, which is item 2 below.

## LEFT DELIBERATELY — three items, none of them defects in the fix

1. **Nine stale `detected` rows** under the old extension-less item key. Five
   `finetuning.uk` `case-study-*` are still true 404s and will re-file under the new key.
   **Four are the OLD predicate's false positives and can be cancelled outright**:
   `fundamentallyai.com:brand-illustration` and three `robot-hands.com:content-hero-tool-*`
   (all serve 200). Not touched because `bugs_open/083`'s lane is actively working the
   `detected` population.
2. **The `storage.DeployedWebPath` durable fix — now FILED as `bugs_open/168`.** A *new*
   purpose containing an underscore stored with `asset_key == purpose` would still be
   mis-rendered. The audit proves that set is empty today across all 267 active asset
   rows; it cannot prove it for assets that do not exist yet. The fix belongs in
   `platform/storage` with **six** consumers — architecture-scope, the shape
   `bugs_closed/124` was vetoed for. **Do not smuggle it into a bug patch**, and note
   that its best candidate inverts `TestDeployedWebPathCannotExpressBrandHeadPaths`,
   which is the `bugs_open/142` lane's deliberate pin — a conversation with them, not a
   silent edit.
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
  Here it was the fleet-wide API usage cap.
- **A trigger that printed nothing is not a trigger that ran.** One resubmission created
  no orchestration at all, because a `cd` earlier in the same compound command made the
  script's relative path invalid — and the grep I filtered its output through matched
  nothing, which looks identical to a quiet success. **Confirm by querying
  `orchestration_states` for the correlation**, never by stdout you did not read in full.
- **`curl` returning `000` is a connection error, not a status.** Re-probe every non-200
  before tallying; five of 127 needed it here and all five were 200.
- **Grep `LANDMINES.md` for your SYMBOLS yourself.** The `SessionStart` hook only matches
  entries against files already **dirty** in the tree. Two landmines written hours earlier
  by other lanes described the exact two traps this lane then hit.
- **The shared tree may not compile** — another session's in-flight edit broke this
  package. Test against `git archive HEAD` plus your own files, in `~/.cache`, **not
  `/tmp`** (16G tmpfs shared by ~30 sessions; it hit 100% during this work and the
  resulting error looks like a command failure when it is an output-capture failure).
