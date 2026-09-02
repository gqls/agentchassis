# NOTES — bugfix 429 (append-only, newest at the bottom)

## 2026-09-02 — session 1 (took the bug on)

- Ownership: `who-owns.py 429` → "OWNED or recently active", but all activity is
  the filing commits themselves; the delivery lane's HANDOFF_2026-09-02 line 145
  says "`429` mirror cannot unpublish (unowned)". Tree grep: no uncommitted
  `platform/publish` work. Direct message exchange with site_delivery_and_editor
  confirmed handover to this session.
- Bug still valid `[MEASURED 2026-09-02]`: `contact.html` → 200,
  `index.html` sibling → 200, invented `zz-invented-control-9481.html` → 404
  (not a parked catch-all).
- Opted-in sites `[MEASURED 2026-09-02]`:
  `SELECT domain, publish_target, publish_project, published_hash, published_at
   FROM sites WHERE publish_target IS NOT NULL` → boxingonline.com
  (published_at 16:49Z TODAY — the post-retraction publish that copied survivors
  and left the orphan; this is why the code fix alone can never fire on the
  existing orphan: no drift) and noted.co.uk.
- Read: `platform/publish/{publisher,b2worker,cfpages}.go` + tests,
  `publish_site_action.go` + tests, `storage/s3.go` (ListObjects paginates to
  exhaustion, mid-page error aborts whole listing — load-bearing for the sweep's
  fail-danger analysis), `zip_deliverable_action.go` (TreeHash names zip cuts by
  last-12-hex TAIL only ⇒ th2 prefix bump does not change zip keys).
- 090 loop not run — substitution stated per the 2026-07-31 ruling: root cause
  established first-hand by two sessions with a discriminating observation
  (frozen chrome), independently re-verified here by code read (zero deletion
  handling) and live probe with controls. No new structural claim filed.
- Peer contributions (delivery lane): bulk-floor hazard (truncated listing is
  fail-DANGEROUS, unlike missed deletes), acceptance-pair (kept-200 control
  against over-deletion). Both adopted.
- Fork adversarial review: 12 findings. The ones that would have bitten:
  `treehash_test.go:43` pins `th1:` (would have broken shared HEAD — the
  a-shared-tree-commit-can-break-HEAD class); `check.py:165` literal
  `"publish_site": 3` + parity test; robots.txt 404-probe wedge.
- Landmine sweep before editing: the CDN-adds-bytes entry (15699) applies to
  byte-compares only — deletion acceptance is status-code-only, immune. The
  reconciler-force entry honoured: th2 needs no forcing.

- Working-tree test run showed TestNoNewSilentScanLoss failing — NOT this
  change: `resolve_internal_links_action.go` is another session's DIRTY file
  (their WIP lowered scan-loss 3→2; the ratchet-down belongs in THEIR commit —
  lowering the baseline here would break committed HEAD the other way). Verified
  by `verify-head-builds.sh --with <my files> --test`: green against clean HEAD.
  Also pre-existing at HEAD, not mine: `go vet` unreachable-code finding in
  `load_component_library_actions.go:207` (file not dirty, not touched here).
- Council: submitted 2026-09-02, SUBMISSION_CORR `b576bcc6-8434-4994-be26-f95485f4797c`
  (DRY_RUN admission passed first, check.py admitted). Committing with
  `Council-Submitted:` trailer per the 2026-07-30 rule; verdict to be read later —
  budget ~30 min for the queue.
- Budget audit after the check.py bump: `budget: 10 — 0 shared action(s) over it`;
  parity test green.

- Post-commit obligations DONE 2026-09-02 evening: `landmines-verify-dispatch.sh`
  dispatched 2 (this lane's entry armed); optional-key-budget-check overlay
  re-applied and VERIFIED AT THE ARTEFACT (cronjob volume → configmap
  `…t2tcc65744` → carries `"publish_site": 4`); MEMORY_workstreams entry added;
  WRONG_CALLS entry added (schema-guessed scheduled_tasks queries ×2 — the
  session's one misstep; no false claim reached a doc). Delivery lane given an
  early commit heads-up on top of the two agreed pings. Council verdict watch
  running in background (poll by payload corr, per the ~30 min queue-latency
  rule — a missing row is latency, not a dropped dispatch).

## 2026-09-02 — council verdict READ and advisories adjudicated

**APPROVED round 1** (corr `b576bcc6`, "4 advisory objection(s) — none high-severity",
9 abstained). Full report: `diagnosis_artifacts` kind=`council_report`, that corr.
Adjudication of each advisory, with the measurement where one was asked for:

1. **Interface widening "asserted, not evidenced"** (editquality, guardian,
   prior_art, architecture — the recurring one). ANSWERED BY MEASUREMENT
   `[MEASURED 2026-09-02]`: repo-wide grep for `ObjectStore` outside
   `platform/storage` returns exactly **7** files (publisher.go, b2worker.go +
   test, publish_site_action.go + test, zip_deliverable_action.go + test) — all
   in the commit or compile-covered (zip tests share `pubFakeStore`); `cfpages`
   takes NO store (`NewCFPages()` — `For` passes no Deps.Store), so it cannot
   break; `storage.S3Client.Delete` is real (s3.go:142, cited in grounded_in).
   Belt: full-tree `verify-head-builds.sh` at HEAD b60d66e3c+ — see below.
2. **"Why a third deletion-safety convention instead of the retraction
   actions' dry_run pattern"** (reuse_agent, prior_art). Adjudicated, recorded
   here as the reasoned answer the seats asked for: the retraction actions
   (`retract_page_deployment`, `retract_asset_files`) delete AT SOURCE with a
   human in the dispatch loop — dry_run-default-TRUE is the right guard there.
   The sweep runs UNATTENDED on the scheduled reconciler; a dry_run default
   there would mean the mirror never deletes anything on the automatic path,
   which IS bug 429 (an undriven mechanism). The floor+opt-in flag is the
   unattended-path analogue: routine convergence automatic, only the dangerous
   bulk shape gated on a human. Wiring retraction actions to also hit the
   mirror is the bug file's fix candidate 2, considered and rejected THERE
   (leaves orphans from any other cause unswept); `b2 sync --delete` is a CLI
   pattern in gqls/sites' GH workflow, not Go machinery in this repo — the
   sweep extends DGH-008's existing in-repo copy+verify mechanism instead of
   importing a CLI into the spawned pod.
3. **"Other code reading the th1: literal"** (guardian). Already measured
   pre-submission (fork review): fleet grep found one historical doc-comment
   quote (`check_page_content_divergence.go:196`) and no code, no
   source-scanning test outside the publish package; DB rows carrying th1: are
   the intended drift. In the lane PLAN; should have been in grounded_in.
4. **"No pod-verification step before trusting the sweep/flag in production"**
   (debug_historian). Accepted — added to RUNBOOK: before ANY hand dispatch
   with `allow_bulk_unpublish:true`, prove the spawned-pod image carries the
   sweep (stamp ancestor-check or capability probe with present+absent
   controls), never trust the roll.
- Belt landed `[MEASURED 2026-09-02]`: full-tree `verify-head-builds.sh` → "OK —
  HEAD 0ba331483 builds" (HEAD by then already carried other lanes' commits on
  top of b60d66e3c, so the widened ObjectStore compiles against the WHOLE
  shared tree, not just the packages the fix touched).

## 2026-09-02 ~21:15Z — the roll landed; convergence watch armed

- v1.0.1355 (stamp `0d2feee2f`) carries the fix: `git merge-base --is-ancestor
  b60d66e3c 0d2feee2f` exits 0 `[MEASURED 21:13Z]`; delivery lane verified the
  same ancestry on both chassis and image-generator-adapter pods (20:56/57Z).
- Pre-convergence state as expected: contact.html still 200 at ~21:0xZ
  (delivery lane's probe, index 200 control) — correct until boxingonline's
  first post-roll reconciler tick (~21:52Z by their three-tick observation
  [INFERRED, their marker]).
- ⚠ **Shared kubeconfig token EXPIRED ~21:1xZ** (both sessions Unauthorized —
  the 3-day-expiry landmine; owner refreshes). Consequences: the `th2:`
  published_hash flip is UNREADABLE until refresh (corroboration only — the
  close criterion is served bytes); rotation order unconfirmable, so if
  noted.co.uk is ahead in the queue, boxingonline converges on the SECOND tick
  (~22:52Z) — a 200 after one tick is not a failure signal; two
  boxingonline-serviced ticks with 200 would be.
- My watch: served pair every 4 min (contact + index, cache-busted), exits on
  404/200, alarms on any other non-200/200 shape (over-deletion can't hide).
  Delivery lane runs the same pair and strikes their §1.2 on the 404.
