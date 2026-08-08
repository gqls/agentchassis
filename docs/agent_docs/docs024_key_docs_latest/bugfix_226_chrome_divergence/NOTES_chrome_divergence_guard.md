# NOTES — bugfix 226 chrome divergence guard (append-only, newest at the bottom)

## 2026-08-08 — session 1 (bug selection + research)

- Picked 226 from `bugs_open/` after ownership sweep: `who-owns.py` clean on
  223–227; live-transcript grep found the FILER (session `0c5a11f2`, closed out
  with a final summary) and a second bug-picker (`98b5904b`) that declined 226
  ("ink is wet") and took 209. 223 self-routes to `architecture_review`;
  224/225 were touched today by the SQAM-003 arithmetic-oracle commit. 226:
  unowned, no open `site_work_items` on the mechanism (only unrelated
  favicon/og-card `image_url_404` rows).
- Premise re-verified at HEAD: the store UPDATE
  (`render_site_components_action.go:938-943`) replaces `rendered_html`
  unconditionally behind the 069 lock predicate; no comparison, no archive.
- **Misstep avoided (and worth recording): candidate 1 in the bug file is not
  implementable as written.** It says "re-render with the row's stamped inputs
  (117 stores them)". `render_inputs` is a jsonb of md5 DIGESTS
  (`datahelpers/chrome_render_inputs.go`) — nothing can re-render from it. The
  file's own "What already exists" section describes the fingerprint correctly
  two paragraphs earlier. Caught by reading the helper before planning, i.e.
  the LANDMINES "fix candidate refuted by its own file" check. Recorded as a
  CORRECTED note in `bugs_open/226` rather than WRONG_CALLS (it was never
  asserted by me as true; the check is the point).
- DB state, measured: `site_components` 57 rows, 57 with HTML, **11 stamped**
  with `render_inputs` (the 117 fix is built but the wave rides the next roll —
  which is the urgency: 46 unstamped rows get rebuilt when it lands).
- Prior art found: `page_component_history` (14,396 rows) is the house archive
  shape — but it archives `content_data` only; `rendered_html` is unarchived
  even there. Scoped OUT of 226 (page side has its own lanes); noted as a
  possible follow-on.
- Writer inventory for `site_components.rendered_html`: 6 classes (render
  overwrite, relink-erase, set/replace, append, core-manager admin dynamic SQL,
  raw psql). This is what decided trigger-over-call-sites: a Go guard can never
  see the psql writer, and the psql writer (mig 268 style) is the bug's origin.
- sqlmock (`DATA-DOG/go-sqlmock v1.5.2`) is in go.mod and used by neighbouring
  tests — behaviour tests are possible without a live DB.
- Migration home confirmed: `docs/agent_docs/sql_for_agents/`, next free number
  **344** (343 is the highest on disk); ROLLBACK sidecar naming per 339/340/341.
- Register: entry goes in `styling-render-pipeline.md` as **STY-054**.
  `rebuild-cascade.md` is DIRTY in the shared tree (another session's edit,
  3 add / 3 del) — deliberately NOT touching that file; same-file-passenger
  risk.

## 2026-08-08 — session 1 (implementation)

- Council submission: corr `cffbfec4-3bec-4577-8844-d17c546ded3e` (8 edits,
  trigger-over-call-sites rationale, risks named incl. fail-closed and
  grep-invisibility). Committing with `Council-Submitted:` per the 07-30 rule;
  budget ~30 min for the verdict, find the run by payload not printed id.
- Migration 344 applied to live DB at ~23:30Z and RECORDED
  (`--record-only`, note names the probe). Verified live:
  `trg_site_component_archive` enabled (`O`), `site_component_history` = 0
  rows (probe rows self-deleted), 0 digests stamped (no backfill — by
  design). **Did NOT use `--apply`: dry-run showed the pending backlog
  contains other threads' files** (335, 337, 338, 340×3, 341×2, 342, 343 —
  several "LIKELY ALREADY APPLIED" per their own guards). The 98-pending
  landmine is real; single-file psql + record-only is the honest route.
- In-file probe design note: the probe exercises the NEGATIVE first (byte-
  identical rewrite must archive nothing — the check-the-no-op-case rule),
  then archive + restore, then deletes its own ledger rows. `updated_at` is
  untouched (probe UPDATEs do not set it).
- Go: `site_component_divergence.go` (classify + emitter mirroring the 069
  emitter), store statement gains `rendered_html_digest = md5($1)` same-
  statement. 4 tests green; `gofmt` clean.
- **Pre-existing red at HEAD, NOT this lane's:**
  `TestValidDocSubjectTypes_LockstepWithMigrationCheck` fails because
  `e1628f7df` (RFC_015 decision records, 2026-08-08 20:21) shipped migration
  `340_doc_notes_decision_subject_type.sql` (adds `'decision'`) without
  adding it to `validDocSubjectTypes` — the FOURTH instance of the
  both-halves-in-one-commit landmine (LANDMINES.md:646). Owning sessions were
  live at 23:26 (mtimes), so left to them; flagged in this lane's close-out.
  `go build` is unaffected; only `go test` reddens.
