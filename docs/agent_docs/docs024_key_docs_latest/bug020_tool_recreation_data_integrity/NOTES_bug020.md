# NOTES — bug 020 (append-only, newest at the bottom)

Technical log: evidence, commands, what the system said, and the missteps.

---

### 2026-07-21 — session start, orientation

- Read `/bugs_open/020`, the vetcomparison HANDOFF, and the live
  `tool-recreation-handler` config. `who-owns.py 020` → owned/active by the
  vetcomparison + imagery workstreams (imagery placed a HOLD on tool imagery
  until 020 is fixed, `3f6f1febf`). This session is the platform fix.

- **Confirmed root cause (b) LIVE**, not just in the seed file: the live
  `recreate_tool` prompt still carries rule 9 verbatim: `9. No fake data or dummy
  outputs — calculations must be mathematically correct`. The seed file
  `099_tool_recreation_handler.sql` is STALE — the live row also carries migration
  138's "Mandatory Behaviour Requirements" section, which the seed does not. So
  always read the live row, never the seed.

- Live models: `analyze_tool = claude-sonnet-5 @8000`, `recreate_tool =
  claude-opus-4-8 @64000`. The agent_definitions row was `updated_at 08:46 today`
  — a fleet re-seed by another thread. Confirms: **patch in place, never
  re-INSERT.**

### 2026-07-21 — a key realisation that shrank candidate (1)

The case file's candidate (1) wants `extract_interactive_fingerprint` to capture
the `fetch()` target and carry it through adoption. Read the fingerprint action:
it detects that `fetch` is *present* (`ifpFetchRe`, `pageSignals["fetch"]`) but
**never records the target URL** — so (a) is real. BUT `recreate_tool` already
receives the full original source (`existing_content.existing_content.raw_html`),
which contains the original `fetch('/data/vet-full-index.json')`, and it renders
the whole analysis JSON via `{{.tool_analysis.result | toJSON}}`. So the
data-source contract can flow analyze→recreate **with a prompt-only change** — no
adoption-crawl plumbing. That is what migration 183 does (adds `data_source` to
the analysis schema). Cheaper (1), same end.

### 2026-07-21 — Half A shipped (migration 183)

- Confirmed all four anchor strings UNIQUE against the live row before writing the
  replaces (`## Requirements` ×1, rule-9 line ×1, `"site_context":` ×1, `assets it
  relies on` ×1). The migration's DO block RAISEs on a no-op replace, so a bad
  anchor rolls back clean.
- Next free migration number: `ls` said 182 was the max file AND `schema_migrations`
  showed `182_legal_pages_aao_finetuning.sql` applied 09:21 today (another thread's
  untracked file). Took **183**.
  - **Trap avoided:** `schema_migrations` has no `version` column — it is keyed on
    `filename`. My first `ORDER BY version` query errored; corrected to `filename`.
- Applied out of band (`psql < 183...sql`), all four UPDATEs = `UPDATE 1`, DO block
  NOTICE fired, COMMIT. Recorded with `run-migrations.sh --record-only`.
- **Independent** verification (not the migration's own DO block) against the live
  row: `integrity_section=true, old_rule9_gone=true, new_rule9=true,
  data_source=true, item8_patched=true`. Committed `266f900e5`.

### 2026-07-21 — Half B built (the mechanical gate)

- Read `checkpoint_for_review_action.go`: it creates a `needs_human_review`
  `site_work_items` row and lets the workflow complete normally. This is the
  machinery to reuse — so the gate needs only ONE new action (the detector), not a
  new terminal.
- **Second defect noticed (noted, not fixed):** in the live workflow
  `validate_tool` (`validate_page_content`) has `error_step = save_sections` — so
  ALL validation blockers (incl. cross-site contamination) are *swallowed* and the
  tool deploys anyway. Fixing that broadens the change and risks blocking legit
  tools on placeholder/template false positives, so the fabrication gate is a
  separate, independent step instead. Left as a noted concern.
- Wrote `check_tool_fabrication_action.go` with a pure `DetectToolFabrication`
  core. **Precision was the hard part** — a seeded PRNG alone is NOT fabrication
  (games use it for gameplay). Tiered detection with a corroboration gate on the
  ambiguous tier (original was data-backed AND recreation dropped the fetch). The
  corroboration cleanly separates the vetcomparison directory (original fetched
  data) from a legitimate name-generator tool (original had the fragment arrays
  too, no fetch).
- `go build` green; `go test -run TestDetect_` → 11/11 pass, incl. the real
  vetcomparison fabrication gated both WITH the confessing comment (Tier A) and
  with it removed (Tier B), and NEGATIVE cases (dice game, calculator, faithful
  fetch-preserving recreation, honest empty state, name-generator) all NOT gated.
- Registered in `registry.go` (verified the 6-line diff is the only change — no
  foreign edits riding). Committed `61f5fe567`.
- Wiring staged **image-first** (out of `sql_for_agents/` so a `--apply` sweep
  can't apply it before the action exists in the pod). Council submitted,
  `SUBMISSION_CORR 8eef369f`.

### 2026-07-21 — concurrent finding folded in (not mine)

Another session appended an addendum to `/bugs_open/020`: the `permanent`
`lock_type` on vetcomparison's corrected components did **not** survive a full page
rebuild (08:08 today) — the rebuild delete-and-recreates the component rows (the
`hero` primary key changed), so a per-row lock cannot survive by construction.
Benign this time (the rebuild regenerated `hero` from a clean source; zero
fabrication live). Reinforces why 020's fix must live in the generator's contract
(Half A) + a deploy gate (Half B), not in a per-row flag.
