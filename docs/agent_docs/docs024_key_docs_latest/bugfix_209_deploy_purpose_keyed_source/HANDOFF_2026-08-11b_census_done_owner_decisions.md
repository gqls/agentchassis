# HANDOFF — 2026-08-11b. 231 census CLOSED (all 3 arms). Detector live at HEAD. Three owner decisions pending. COLD-START HERE.

Supersedes `HANDOFF_2026-08-11_11_of_11_class_fix_approved.md` (the logo saga —
still the right doc for 235/380 history). Evidence + missteps: `NOTES_209…`
(08-11 afternoon section). Milestone read-out:
`SUMMARY_2026-08-11b_census_closed_and_the_signature_collapse.md`. Shared
accounts: `bugs_open/231` (census + all evidence), `bugs_open/213` (notified —
their producer key is lying), `bugs_open/240` (kafka) — contribute INTO them.

## What this session closed

- **231 census, all three arms** (`bugs_open/231` 2026-08-11 sections carry
  everything): 62/164 specs with Defaults, 232 fields. 195 findings = 24
  dead-mismatched / 75 dead-matching / 96 dotted-conditional. Read-path
  verification: 20/24 mismatched are honoured via DIRECT config reads (each
  read line cited) — **the only real damage is `audit_source` ×4**. Arm 2
  closed: every diverging dotted binding has a verdict (5 = the 348/380
  repairs; 28 = `*_field` direct-read family, extractor-irrelevant; 1 latent =
  `derive_card_asset entity_type`, benign until phases I5/I6 — grep 231 for
  it if you're building those).
- **Candidate 3 BUILT**: `cmd/config-key-audit --default-shadowed-keys` +
  `scripts/audit-default-shadowed-keys.sh`; `--specs` now emits `defaults`.
  Calibrated on both 231 faces as committed tests. Exit 1 only on
  dead+mismatched. Commits `279f80953` + `1221eb30f`, both ancestors of
  chassis v1.0.1288's stamp `bb5348642` (probed, two-sided control).
  **Council REFUSED on scope** (cmd/ + scripts/ outside platform/internal/pkg)
  — recorded in the commit body, submission JSON kept alongside; do NOT
  resubmit or FORCE.
- **THIRD LIVE FACE of 231**: four auditors (brief-fidelity-auditor,
  content-quality-auditor, site-review-agent, visual-design-auditor) set
  distinctive `audit_source` statics; all dead
  (`write_audit_findings_action.go:495` reads via inputs; Default
  `design-audit` wins). Artefact-proven: zero rows fleet-wide carry the
  intended labels; proof row `item_type='audit_finding_brief_fidelity'` +
  `audit_source='design-audit'` (07-24); 136-row merged stream through today.
  Filed in 231 · notified in 213's file · LANDMINES entry appended AND synced
  (5 doc_notes rows verified) · memory topic corrected.

## The three OWNER DECISIONS (all costed in bugs_open/231, 08-11 sections)

1. **audit_source repair — route and who.** No config-only fix exists (231
   explains why: a static can't be expressed against a Default, and removing
   the Default alone doesn't help — the `==""` fallback re-imposes
   design-audit). Option (a): four lines in `write_audit_findings` — read
   `audit_source` from config directly, the idiom ~20 sibling actions use;
   Go change, needs the next roll; **the file is the bugfix_213 lane's while
   their round is open — coordinate via their file, already notified**.
   Option (b): wait for candidate 2. Recommendation on record: (a) now —
   every day adds mislabelled rows, and (b) can still land later.
2. **Candidate 2 (config-static beats Default in ExtractActionInputs) — ship
   or drop.** Blast radius MEASURED: activation set = the 4 audit_source
   entries and NOTHING else (75 matching are no-ops by equality; the other 20
   mismatched are no-ops because their actions never consult the inputs value
   — re-verify the read-path table at implementation time). If decision 1(a)
   ships first, the activation set is EMPTY and candidate 2 is pure
   future-proofing — arguably the detector's exit-1 already guards new
   authors. Genuine judgement call now, not a necessity. If shipped: council
   round required (platform/), the three 231-pinned tests + 2 calibration
   tests flip DELIBERATELY, and the composite/empty-string design questions
   in 231 need answers.
3. **Stale logo.jpg deletion** (carried from the morning handoff, unblocked).
   Zero renderable references fleet-wide. fundamentallyai's index has a queued
   `needs_page:index:151census` rebuild (brochure lane's) — it regenerates
   from patched content_data; don't fight over that page.

## What remains, in order (after the decisions)

1. Whichever of decisions 1/2 the owner picks — implementation + council (if
   platform/) + roll + artefact proof (a new auditor run writing its OWN
   label; re-run `audit-default-shadowed-keys.sh`, the 4 findings drop out).
2. **240**: `tail ~/kafka-sweep-240.log` — the KUBECONFIG fix's first real
   APPLY run is the next 00:17/12:17 LOCAL slot the machine is awake for
   (crontab `17 */12 * * *`; the 11:17Z entry is the pre-fix refusal). Then
   C2 safe subset (scheduler-scoped transport) + the C1 question.
3. 209 Phase 3 (retire dead writers) and 236 — open, unowned by this thread.

## Cold-start checks

1. `go test ./cmd/config-key-audit/ ./platform/orchestration/actions/` — all
   green expected (includes the 7 new detector tests + the 8 lane tests).
2. `./scripts/audit-default-shadowed-keys.sh` — expect exit 1 with exactly 4
   dead-mismatched (the audit_source four) until decision 1 ships; anything
   NEW is a fresh instance of the class — investigate before anything else.
3. Deploy state: chassis v1.0.1288 = stamp `bb5348642` (this lane has nothing
   inert-awaiting-roll; per-service stamps differ — probe, don't assume).
4. The 380 mapping + 8 lane tests: HANDOFF_2026-08-11 §cold-start items 1–2,
   unchanged.
5. Before touching `write_audit_findings_action.go`: `scripts/who-owns.py 213`
   AND grep live transcripts — that lane was mid-round on 08-11.

## Traps for the next session (fuller list: LANDMINES.md, RUNBOOK §11)

- A `--default-shadowed-keys` finding asserts the INPUTS path only — grep the
  action's read line before calling it damage (20/24 were false this way).
- The read-path table in 231 is a point-in-time census, not a contract.
- `spec->>'audit_source'` currently means "any of five agents" — do not
  attribute audit findings by it until decision 1 ships (LANDMINES entry).
- Council gate refuses cmd//scripts/ on scope — that refusal is the recorded
  outcome, not a bug to work around.
