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

### 2026-07-21 — tolerated migration-number collision on 183 (not a problem)

A concurrent session (bugfix 045) also took **183** —
`183_generic_hero_tool_component.sql` — for its own DB-config patch. So there are
now TWO `183_*.sql` files. Checked the ledger: both are recorded, mine first
(`183_tool_recreation_no_invented_data.sql` at 10:52:58, theirs at 10:56:37). No
functional conflict — `schema_migrations` is keyed on **filename**, not number, so
they are distinct rows, both applied. The 045 session already noticed and
explicitly tolerated it. **No renumber** (forward-only; renaming an applied+
recorded migration would orphan its ledger row). Next free number is 184. This is
the concurrent-sessions numbering race the repo warns about — both of us read
max=182 and took 183 within the same minute; the collision is cosmetic because the
ledger keys on filename.

### 2026-07-21 — council run WEDGED (003-class), and a verdict-query trap I nearly hit

- The first council submission (`SUBMISSION_CORR 8eef369f`, orch `0b0552b8`) **wedged**:
  it ran the early seats (`gate_guidelines` → `gate_tooling_provenance` →
  `review_tooling_provenance`) then stopped at `review_tooling_provenance |
  EXECUTING_STEP` and did not move for **3h42m**. Only a `fix_plan` artifact exists;
  **no `council_report`**. This is the 003-class EXECUTING_STEP hang (a dropped
  awaited response), NOT a REVISE/REJECT — it is an infra failure, not a judgement
  on the change. Other submissions completed in the same window (`ed4851c9` got a
  REVISE), so the council infra is not globally down; this was a transient per-run
  drop. Resubmitted once with `RESUBMIT_CORR=8eef369f` (same fix-correlation, so the
  trail accumulates).
- **TRAP nearly hit:** the CLAUDE.md verdict query
  `SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at
  DESC LIMIT 1` returns the most recent council note **FLEET-WIDE**, not yours. My
  background poller returned `ed4851c9`'s REVISE note — another thread's — which
  reads as *my* verdict if you don't check the "submission correlation:" line in the
  body. **Always confirm the verdict against the correlation-keyed source:**
  `diagnosis_artifacts WHERE correlation_id=<yours> AND kind='council_report'`
  (that returned zero rows for me, correctly — the run never produced a report).
  Same family as the documented "council_report source_agent='generic' fleet-wide"
  trap. RUNBOOK query corrected.
