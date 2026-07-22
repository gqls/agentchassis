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

### 2026-07-21 — council round 2 = REVISE (a good one); addressed, resubmitted round 3

Round 2 (resubmit orch `d9aaa0b5`) returned **REVISE**, 5 approve / 4 object,
decided by editquality. The objections were sharp and mostly correct:

- **editquality (HIGH) — the central one, and it was right:** the detector is
  INERT without the wiring, and I had NOT put the wiring in the reviewed plan — yet
  my rationale asked them to "confirm the gate routing cannot deploy a fabricated
  tool". You can't confirm deploy-blocking from a plan that omits the deploy-block.
  → **Fixed round 3: the wiring is now edit 5 (config_change), naming
  tool-recreation-handler as owner and showing the fabricated=true branch never
  reaches save_sections/deploy_page.** guardian, compliance, prior_art_librarian all
  independently flagged the same dormant-machinery risk.
- **guardian (MED) — field paths unverified.** Legit concern (a wrong path = silent
  fail-open while unit tests pass on hand-built payloads). **Verified, and they are
  correct:** `completeness_check.clean_html` is the exact path `validate_tool` reads;
  `existing_content.existing_content.raw_html` + `tool_analysis.result` are
  referenced by the recreate_tool prompt itself. No code change — documented.
- **reuse_agent (MED) — did I search for an existing quality pipeline?** Did the
  negative-result search: `compute_component_quality` scores `content_components`
  template CONTRACT compliance (tag balance, schema-field match), has ZERO
  fabrication signals, runs **0** times in this workflow, and recreated tools are
  `page_components` (invisible to it). No hook to extend — the detector is novel.
- **editquality (LOW) — test filename.** Renamed to
  `check_tool_fabrication_action_test.go` (commit `ec6f5788b`), tests still green.
- **tooling_provenance (MED) — leave a travelling doc_note.** Inserted a
  design-rationale note keyed `subject_type=pipeline / subject_key=tool-recreation-handler`
  (id 0547229b) recording the Tier A/B corroboration tradeoff for the next fixer.
- **debug_historian (MED) — pod-grep before trusting the gate.** Already in the
  RUNBOOK and the wiring header; cited in the round-3 rationale.

Approvers (compliance, constitution, mission, prior_art_librarian) all endorsed the
DESIGN (tiered corroboration targets the real bug-020 signature, not a generic PRNG
grep; reuses checkpoint_for_review; fail-open is a reasonable narrow default, not a
gate weakening) — the only blocking theme was "review the wiring too", now done.
Round 3 resubmitted (orch `8fba93dd`, same corr `8eef369f`).

### 2026-07-21 — round 3 VOIDED at persist_submission (MY error, not infra)

Round 3 ended `complete_invalid` with **no fix_plan and no council_report** — it
voided at `persist_submission`, before any review seat ran. Cause was NOT the
019/036 truncation-void class and NOT infra: **edit 5's `file` field was a
descriptive string** ("platform/... (agent_definitions ...) — WIRING_...sql")
rather than a repo-relative path, and `diagnose_persist_fix_plan` validates
repo-relative paths server-side → rejected the plan → `complete_invalid`. The 097
trigger's client-side check does NOT validate the per-edit `file` path shape, so it
passed locally and failed server-side. **Lesson: every edit `file` must be a clean
repo-relative path (jq `.plan.edits[].file`), even for a config_change/docs edit.**
Corrected edit 5 to the real wiring SQL path and resubmitted **round 4** (orch
`e07a6629`). Council-run tally so far on this change: R1 wedged (003-class), R2
REVISE (real, addressed), R3 voided (my bad path), R4 pending.

### 2026-07-21 — council round 4 = REVISE (found a real flaw); addressed; STOPPING the loop

Round 4 (orch `e07a6629`) ran clean this time: `complete_revise`, 11 reviewers, 5
abstained, decided by editquality. It reviewed the fuller plan INCLUDING the wiring
(edit 5). Two classes of objection:

- **REAL FLAW — bug_historian (HIGH), echoed by editquality/compliance: the detector
  failed OPEN.** Missing recreation HTML returned `fabricated=false` → deploy. That is
  the exact `missingkey=zero` silent-default class bug 020 itself belongs to. **Fixed
  (commit `37d3bb119`): fail-SAFE now** — un-inspectable output returns
  `fabricated=true, tier=uninspectable` → held for review. Bonus: this also kills
  guardian's separate "silent no-op if the field path drifts" fear — a drifted path
  now fails LOUD (everything held) instead of silently passing. New test. This was the
  single most valuable thing the whole council thread surfaced.
- **SQL/plan discipline on edit 5 (the wiring) — debug_historian (HIGH), prior_art,
  guardian:** jsonb surgery on a live workflow row is needle-gate class; show a
  pre-count + RETURNING; put the "verified" field-path SQL check IN the plan, not
  asserted. **Partly addressed:** the WIRING SQL now opens with a needle-gate DO block
  (assert exactly 1 live row AND not-already-wired = idempotency guard) and the repoint
  UPDATE carries RETURNING. The "evidence in the plan" ask is inherent to a
  config_change edit that can't be applied/verified until the action ships (image-first)
  — the verification queries live in the RUNBOOK and were run live this session.
- tooling_provenance: the doc_note isn't a plan edit — true; it was written out of band
  (id 0547229b). guardian/prior_art (LOW): want the reviewers' own code_checks on
  "registry additive"/"new work". bug_historian (MISSING): audit OTHER regeneration
  pipelines for the same fabrication class — a legitimate scope-broadening follow-up
  (candidate 2's "apply to every generative prompt"), noted, out of scope for this fix.

**DECISION: stop the council loop here (after R4).** Tally: R1 wedged, R2 REVISE
(addressed), R3 self-voided (bad path, fixed), R4 REVISE (addressed). The two REVISE
rounds each found something real (wiring-must-be-in-plan; fail-open) and both are now
fixed in committed code/docs. The residual R4 objections are plan-*presentation*
(show the SQL check in-plan, run code_checks) and scope-broadening (audit other
pipelines) — not defects in the change. Marginal value of another round is low, the
infra is flaky (a wedge + a void already), and a config_change edit is inherently hard
to fully evidence before the image roll. **No APPROVED was obtained, so NO
`Council-Reviewed:` trailer** (CLAUDE.md: earned by APPROVED only). The change stands
on: verified field paths, a 12-test precision+fail-safe contract, needle-gated wiring,
and a design the approving seats (compliance, constitution, mission, prior_art,
render_guardian) endorsed. The owner reviews before rolling the image.

### 2026-07-22 — gate WIRED & LIVE on v1.0.1146; fail-safe gap found; owner: finish next roll

The image rolled. Verified BEFORE touching anything (never trust the tag):
- Pod `agent-chassis-687cdf6db5-fq2fd`, **v1.0.1146**, started 2026-07-21T18:50:01Z.
- `strings /app/agent-chassis | grep -c`: `check_tool_fabrication`=**4**,
  `corroborated_corpus`=**1** (Tier B live), positive control
  `check_tool_completeness`=2, negative control=0. Detector IS live.
- **BUT `uninspectable`=0** — the fail-SAFE commit `37d3bb119` (18:27:40Z) is NOT in
  this image; it was built ~20 min before that commit. **The pod runs the fail-OPEN
  detector.** Edge-case only (empty output isn't fabrication — real fabricated content
  is non-empty and still inspected), so the core bug-020 hole is closed; the hardened
  version needs the next roll.
- The gate was NOT wired (`check_completeness.next_step` was still `save_training_data`;
  no `check_fabrication` step) — dormant machinery, exactly what prior_art warned of.

WIRED via **migration 189** (`sql_for_agents/189_wire_tool_fabrication_gate.sql`,
needle-gated: pre-count + idempotency guard + RETURNING; applied out of band,
ledger-recorded). Routing independently re-queried (not the migration's own DO block):
`check_completeness → check_fabrication → route_fabrication`; conditional
`fabrication_check.fabricated == true → request_fabrication_review → complete` (NEVER
save_sections/deploy_page), else → `save_training_data → validate_tool → … →
deploy_page`. **The WIRING_..._APPLY_AFTER_IMAGE.sql staged file is now SUPERSEDED by
189** (re-applying it would safely abort on the idempotency guard).

Did NOT run a bespoke induced-fault today: the routing PRIMITIVES are already
production-proven in this same workflow (`check_page_found` uses the same `conditional`
+ dotted-path `page_record.found == true` shape every run) and the detection is
unit-proven, so the residual runtime risk is low — and doing a scratch reproduction on
the fail-OPEN version now and AGAIN on the fail-safe version after the roll is wasteful.

**Owner decision (2026-07-22): keep 020 OPEN, finish on the next image roll.** Core
protection is live now, so no urgency. TO CLOSE (next roll): (1) pod-grep confirms
`uninspectable` >= 1 (fail-safe landed); (2) ONE induced-fault test — recreate a
data-backed tool, confirm a `needs_human_review` item is raised and the page is NOT
deployed with generator symbols (grep the rendered page, not `complete`); (3) move 020
to bugs_closed.

### 2026-07-22 (2nd roll) — the induced-fault EARNED ITS KEEP: caught a silent no-op bug

v1.0.1149 rolled WITH the fail-safe (`uninspectable` grep = 1, `check_tool_fabrication`=4,
`corroborated_corpus`=1). Wiring had persisted from 189. Ran the **induced-fault probe** —
the fixloop-036 pattern: a scratch agent `tool-recreation-020probe` (all-local workflow,
run in-process via generic orchestrate like council-gate) whose `check_completeness` reads
a stubbed fabrication from `input_data.stub_html`, then the LIVE `check_fabrication →
route_fabrication → checkpoint`. Terminals: `complete_held` (PASS) vs `complete_clean` (FAIL).

Result: reached `complete_held` and raised a `needs_human_review` item — routing WORKS. **But
the verdict was `tier:"uninspectable"`, NOT `declaration`.** The fabrication (with the
confessing comment + makePostcode, `completeness_check.clean_html` = 1412 chars, present)
never reached the detector. Chased it: `input_data.stub_html`=1447 present, `completeness_check.clean_html`
=1412 present — yet `check_fabrication` read empty.

**Root cause — a REAL production bug the probe exposed:** the wrapper read `html_field` via
`datahelpers.ExtractActionInputs`, whose **Strategy 0** (`action_inputs.go`) resolves any
dotted config VALUE against `collected_data` BEFORE the handler runs. So
`html_field="completeness_check.clean_html"` was resolved to the 1412-char HTML, `inputs.Get`
returned the content, and the handler extracted AGAIN with the content as a path → `""` →
`recreation=""` → `uninspectable`. `check_tool_completeness` avoids this by reading
`config["html_field"]` directly; I used `ExtractActionInputs` and double-resolved.

**Severity:** on the fail-OPEN detector (v1.0.1146, which I had WIRED), this reads empty for
EVERY recreation → `fabricated=false` → deploys everything, incl. real fabrications — a silent
no-op that looks wired. The fail-SAFE change (council R4) turned it into a loud over-HOLD,
which is the ONLY reason it was catchable. Without inducing the fault + the fail-safe, I'd
have "verified" the gate green and closed 020 while it silently passed fabrications.

**Actions:** (1) UNWIRED the gate live (snapshot; `next_step`→`save_training_data`, 3 gate
steps removed) — 0 real recreations had been over-held; (2) FIXED the action (`1a2718213`):
read config paths directly + new wrapper regression test `TestCheckToolFabricationAction_ReadsDottedConfigPath`
(fails against the bug); (3) deleted the scratch agent + cancelled the probe review item;
(4) logged WRONG_CALLS.

**To CLOSE (next roll):** pod-grep the fix is in the binary → re-apply 189 → re-run the probe,
require `tier:"declaration"` (real detection) + HELD not deployed → bugs_closed + lift the
imagery HOLD. The probe agent SQL + trigger are in the session scratchpad if needed again.
