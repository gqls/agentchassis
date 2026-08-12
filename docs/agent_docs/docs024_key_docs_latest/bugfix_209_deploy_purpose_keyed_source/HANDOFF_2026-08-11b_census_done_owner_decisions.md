# HANDOFF — 2026-08-11b. 231 census CLOSED (all 3 arms). Detector live at HEAD. **Owner decisions RULED (evening) — implementation is the next session's work.** COLD-START HERE.

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

## OWNER RULINGS 2026-08-11 (evening) — all three decided; do not re-litigate

1. **audit_source repair: option (a) NOW, and (b) later** — i.e. the four-line
   direct-config read in `write_audit_findings` ships first as its own task,
   AND candidate 2 still ships afterwards on its own timetable.
2. **Candidate 2: SHIP.** "An explicit config value beats a default" becomes
   the resolver's rule. Full costing and design questions below.
3. **Stale logo.jpg files: LEAVE for now.** Off the list; do not delete.

## What remains, in order (the next session's work-list)

1. **Task A — the audit_source fix (ruling 1a).** Four lines in
   `write_audit_findings_action.go` (~:495): read `audit_source` from
   `params.StepConfig.Config` first (the idiom ~20 sibling actions use, e.g.
   `datahelpers.GetStringField(config, "audit_source", inputs.Get("audit_source"))`
   — decide the exact fallback chain deliberately: config-static → resolved
   inputs → the "design-audit" default). BEFORE touching the file:
   `scripts/who-owns.py 213` AND grep live transcripts — that lane was
   mid-round on this file 08-11; coordinate via `bugs_open/213` (already
   notified, offer already on record there). Council round: platform/ — IN
   scope this time. Then: commit, next roll, artefact proof = a fresh auditor
   run writing its OWN label + `audit-default-shadowed-keys.sh` finding count
   drops 4→0 dead-mismatched. Update `bugs_open/231` and the LANDMINES
   `audit_source` entry (its "until the shadow is fixed" clause) when proven.
2. **Task B — candidate 2 (ruling 2), its own council round.** Change
   `ExtractActionInputs` so an explicit config value for a spec field beats
   `spec.Defaults` ("config-static beats Default, Strategy 0 beats both").
   Everything known is in `bugs_open/231`: blast radius measured 2026-08-11
   (activation set = the 4 audit_source entries; EMPTY once Task A ships —
   re-run the census + read-path check at implementation time, the table is a
   point-in-time snapshot). Design questions that round must answer, on
   record in 231: composites (Strategy 5 deliberately excludes them today) ·
   explicit empty string (`changed_files_field: ''` live on
   feature-implementer, authored as "disable") · whether the Deprecated
   bridge (Strategy 3) also beats Defaults or keeps losing. **Tests flip
   DELIBERATELY, citing the ruling:** the three 231-pinned tests in
   `platform/orchestration/actions/deploy_image_asset_purpose_source_test.go`,
   AND the detector must be re-specified in the SAME round —
   `--default-shadowed-keys`' `static_string`/`non_string_literal` classes
   describe the OLD resolver; after candidate 2 they become live config, so
   the mode's dead classes shrink to `unextractable_field` (+ whatever the
   bridge/composite decisions leave dead) and `defaultshadow_test.go`'s
   calibration flips. Shipping the resolver change without the detector
   update makes the detector lie — one round, one commit, both files.
   Go change → inert until an image roll; DB config untouched.
3. **240**: `tail ~/kafka-sweep-240.log` — the KUBECONFIG fix's first real
   APPLY run is the next 00:17/12:17 LOCAL slot the machine is awake for
   (crontab `17 */12 * * *`; the 11:17Z entry is the pre-fix refusal). Then
   C2 safe subset (scheduler-scoped transport) + the C1 question.
4. 209 Phase 3 (retire dead writers) and 236 — open, unowned by this thread.

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

## ⚠ `cmd/config-key-audit/` is CONTENDED — check before you edit it (added 08-12)

Another lane is building **RFC_022's optional-key-budget counter** in the same
package: `optionalbudget.go` + `optionalbudget_test.go` (untracked as of
2026-08-12 morning) and a 13-line addition to `main.go`'s header + dispatch.
Task B has to touch `main.go` and `defaultshadow*.go` in the same package.
**Their 13 lines are pure additions and displace nothing of ours** (checked
`git diff` before this handoff was committed) — but a pathspec commit CANNOT
protect you from a same-file passenger, so:

- Before editing `main.go`: `git diff cmd/config-key-audit/main.go` and read
  whose lines are there. Commit only when the diff is yours, or say so in the
  message if you deliberately carry theirs.
- Their work was briefly absent from the tree and then back within minutes on
  08-12 — treat any "their mode has vanished" reading as a stale snapshot,
  not as a licence to re-add or delete it.
- The two modes are independent: theirs counts optional keys per shared
  action, ours classifies default-shadowed entries. No shared symbols beyond
  `liveAgent`/`decodeLiveAgents`. Do not merge them.

## Traps for the next session (fuller list: LANDMINES.md, RUNBOOK §11)

- A `--default-shadowed-keys` finding asserts the INPUTS path only — grep the
  action's read line before calling it damage (20/24 were false this way).
- The read-path table in 231 is a point-in-time census, not a contract.
- `spec->>'audit_source'` currently means "any of five agents" — do not
  attribute audit findings by it until decision 1 ships (LANDMINES entry).
- Council gate refuses cmd//scripts/ on scope — that refusal is the recorded
  outcome, not a bug to work around.
