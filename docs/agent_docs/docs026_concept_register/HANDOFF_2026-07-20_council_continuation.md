# HANDOFF — council work, cold-start for the next session (2026-07-20 evening)

**Read this first, then the standing five** (`PLAN_concept_register.md`,
`PLAN_2026-07-20_direction_reach_and_drift_guard.md`, `RUNBOOK_concept_register.md`,
`RUNNING_NOTES_concept_register.md` — turns 25-32 are the council era,
`README_where_we_are.md` for the owner's plain-prose history). This file is the
map; the notes are the evidence. **Re-read CLAUDE.md from disk before acting on
anything here** — it changes daily, and three of its recent changes were caught
only by re-reading (see the memory `reread-claudemd-and-standing-docs`).

## What exists, live in production, as of 2026-07-20 ~19:30 BST

**The council: 16 reviewer seats, on BOTH councils** (`fix-proposer` — the fix
loop's own; `council-gate` — the advisory gate any thread submits to via 097),
kept identical by the mechanical mirror (`099_SYNC_gate_roster.py`; NEVER
hand-patch the gate). Roster:

- **5 always-on:** `review_editquality` · `review_constitution` (root-cause-not-
  workaround FIRST, reuse-before-recreate, schema-first, parameterised-only —
  v19) · `review_mission` (best-site-per-domain; revenue model shapes the site;
  silent override is THE failure mode — v19) · `review_prior_art` (the
  librarian: verifies "does not exist / needs to be built" claims with
  code_checks + SQL checks; asserted-absence / dormant-machinery — v20) ·
  `review_guardian` (the ONLY veto; stability proviso; code_checks).
- **11 gated** behind the deterministic relevance filter (`select_panel`
  footprints; skipped seat = abstention in `council_decide`): bug_historian,
  reuse_agent, guidelines, tooling_provenance, adoption, diagnosis, improvement,
  compliance, render_guardian (carries the owner's plain-rerender false-green
  trap, patched 2026-07-20), llm_reliability, debug_historian.
- Every specialist is advisory (`approve|object`); only the guardian vetoes
  (any seat's `veto` string would reject — that's why none offer it).
- Council chain: `select_panel → editquality → constitution → mission →
  prior_art → [11 gates] → guardian → council_decide`. Typical fix wakes 5-8.

**The direction guard (all live, all tested):**
- `DIRECTION_LEDGER.md` (this dir) — blessed canonical constitution + mission,
  sanctioned copies, sha256s. The ledger guards itself.
- `.githooks/commit-msg` — REAL GATE: commits touching blessed paths need a
  `Direction-Approved: <name>` trailer (earned by the owner's word only). Its
  first live firing was its own creation commit (`fc0652ce8`).
- `fixloop_eg_dartsonline/100_CHECK_direction_integrity.py` — read-only, three
  surfaces (files vs ledger, copies vs canonical, seat-prompt anchors in BOTH
  councils). **Run it after ANY seat migration** and after any deploy (re-seed
  clobber risk). Last run: ALL GREEN, post-v1.0.1140.

**The mission reach (R0+R1, observe-only, live on `domain-research-classifier`):**
R0 = platform-mission digest in `classify_and_extract`'s prompt. R1 =
`review_mission_alignment → gate_mission_note → append_mission_note` after the
spec writes; objections land in `doc_notes` (categories `['mission-review']`);
every error path routes to `create_next_item` — **structurally cannot block a
build**. Consumer: `fixloop_eg_dartsonline/101_REPORT_mission_review_findings.sh`.

**The binary (v1.0.1140, rolled 2026-07-20 18:58 BST, pod-verified):** the
truncation family is live — `TruncatedError` carries the partial, ollama
`done_reason=length`, `stop_reason=refusal` decode. Bug 008 CLOSED by its
owning thread. Council rounds can no longer be voided by a silently-truncated
reviewer (019's upstream cause).

## Proven end-to-end (don't re-prove; cite)

- First real council outing: BUG A run `53da3a30` — 3 revise rounds, APPROVED,
  filter woke 5/13, 6 abstentions counted, revise loop EXPANDED the fix to both
  providers (notes turn 27).
- Implementer + build gate: run `70680566` — correct code, gate rightly blocked
  a cosmetic gofmt miss, no PR (bugs_open/013 filed for the gofmt-at-commit-prep
  fix; still open). Hand-finished as `f32b208e5`; SUPERSEDED by the better
  TruncatedError form, shipped in v1.0.1140.
- The gate's own first firings: render seat objections, verdict flow, 098
  coverage — see `council-gate-workstream` memory (that thread owns the gate).

## Open threads, in order of likely next ask

1. **R2 — promote mission review to enforcement.** Wait ~a week from 2026-07-20,
   run `101_REPORT_mission_review_findings.sh 7`, hand-grade with the owner
   (consultancy shape is LEGITIMATE when evidenced; the DEFAULT is the failure).
   Owner decision. Design note: classifier has no revise loop yet — R2 needs one
   (small; re-run classify_and_extract once with the objection injected).
2. **R4 — seat constitution/mission/librarian on the feature-designer and
   experience-planner councils.** Same v19/v20 surgical pattern BUT those are
   OTHER THREADS' co-edited machinery (feature-builder + experience-loop
   workstreams) — read their live rows and memories first, coordinate, drift-
   check, then patch + their own smoke run.
3. **R3 — `mission_alignment_check` discovery check** (fleet audit of built
   sites; shape-mixing detectors). Go + image window; observe-only; findings at
   `detected` are fine HERE (it's the improvement loop's own contract) — unlike
   R1 (see landmine 4).
4. **D4 — the `standards` table** (CTS-029's unbuilt destination; verified
   absent 2026-07-20). File canonical → seeded rows → assembler + seats read
   rows. Council-review the change (it's platform machinery).
5. **Multi-model diagnosis gauntlet** — proposed subproject, recorded in
   `PLAN_concept_register.md`; its prerequisite (MDL-038 truncation detection)
   SHIPPED in v1.0.1140, so it is now unblocked if the owner green-lights.
6. **Register corrections:** CTXK-004 claims `cmd/assembler` exists — it does
   not (checked 2026-07-20). A live asserted-absence specimen in our own
   register; correct via the register's own correction discipline.

## Landmines (each cost something real; do not relearn)

1. **Surgical only.** Both council rows and the classifier are co-edited live.
   Chained `jsonb_set` + needle-gates; NEVER full-config reapply. Seat
   `fix-proposer`, then `099 --apply` — never hand-patch the gate.
2. **Re-assert needles INSIDE the UPDATE's WHERE.** The render seat's prompt
   changed under this session BETWEEN a needle-count and a dump (another thread
   patching the same seat). A gate from an earlier read is worthless.
3. **Pre-flight runs:** structural changes (new steps/rewires) need 0 active
   council runs; text-only additive prompt patches are safe mid-traffic (the
   step reads its prompt at execution). Queue races happen — check WHERE the
   active run is.
4. **The triager is site-scoped and TYPE-BLIND** (`triage_detect_items_action.go`
   :91-103): it promotes ALL `detected` items into dispatch. Observe-only
   findings must NOT be `detected` work items (they'd dispatch to a nonexistent
   handler — the 023 class). Use `doc_notes` + a named report consumer.
5. **Check BOTH bug dirs before UPDATING a case, not just filing** —
   `ls bugs_open/NNN* bugs_closed/NNN*` first. A case closed mid-evening under
   this session and an append recreated it as an orphan fork (WRONG_CALLS.md,
   2026-07-20 entry).
6. **Deploys can re-seed configs.** After any chassis roll: `100_CHECK` + verify
   the classifier R0/R1 wiring. Both survived v1.0.1140, but the classifier row
   was touched 35s pre-deploy by something — anchors held, stay watchful.
7. **No dispatch within ~300s of a chassis pod (re)start**; verify deploys
   against the RUNNING pod (`strings /app/agent-chassis | grep -c "<symbol>"`),
   never git, never the tag.
8. **Seat model/config standard:** claude-sonnet-5 @ 8000, temp 0.0,
   `input_fields [diagnosis_row, plan_persisted, schema_hint]`, error_step
   `complete_refused`; new seats advisory-only. The 099 mirror transforms
   diagnosis→rationale for the gate — keep the exact `## The diagnosis` /
   `{{.diagnosis_row.conclusion}}` tokens so the transform matches.
9. **`Direction-Approved:` trailer = the owner's word for THAT change.** Never
   add it to pass the gate. `Council-Reviewed:` = APPROVED verdicts only
   (a REVISE trailer is a permanent false claim).

## Verify-current-state one-liner (run before believing this file)

```bash
python3 docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/100_CHECK_direction_integrity.py
# then: seats + sync
python3 docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/099_SYNC_gate_roster.py   # dry run; expect 16/16, drift (none)
```

Figures in this file were live on 2026-07-20 evening; per CLAUDE.md, re-ground
any of them before repeating them.
