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
