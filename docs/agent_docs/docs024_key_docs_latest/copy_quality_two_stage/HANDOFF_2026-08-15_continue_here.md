# HANDOFF 2026-08-15 — continue here

**Lane:** `copy_quality_two_stage`. **State: the approved 08-13 execution plan is fully
delivered or closed at its gates.** This file is the entry point for a fresh session;
NOTES carries the evidence for every claim below.

## What is DONE and verified (do not redo)

1. **The v2 house voice is live fleet-wide** (owner-approved 08-13 after reading both
   candidate texts and a capture-only sample run). All seven writer agents reference
   `{{.voice_style}}`; the block lives ONLY in `agent_default_configs.voice_style_block`
   (6,032 chars). Verified at the artefact: a real run's `prompt_rendered` carries the
   block exactly once, old rule zero. Fleet asserts: 0 agents match `'start with the
   fact'`, 7 match the reference. Register **CQ-022** (+ index row). Rollback: every
   pre-change text in `agent_definition_prompt_backups` (2026-08-13 rows) +
   `agent_def_backup_20260813_voicecarrier` + `agent_default_configs_bak_20260813_voicecarrier`.
   ⚠ v1 (H-vintage) was HELD by the owner and superseded — *"H was a starting point.
   See the mortgagecalculator thread."* Do not resurrect v1.
2. **bugs_closed/264** (audit_source) — closed by a concurrent session: migration 399
   (four `set_audit_source` steps) + Go `Required`/no-default, council-approved,
   rolled `v1.0.1298`, proven at the binary. `bugs_open/272` spun out (site-review-agent
   `findings_field`, not ours).
3. **D3 review** (loancalculator / loancash / lendzy) — findings in NOTES 08-14.
   Verdict: loancalculator carries the latent adversarial shape in TWO live fields;
   lendzy/loancash's frame is EARNED (rights-enforcement is the errand) — no change.
4. **LMC evidence_base** — NOT applied (their lane was mid-Track-B2; coordination gate).
   Candidate + full findings handed over:
   `loanandmortgagecalculator_couk/CONTRIB_2026-08-14_evidence_base_candidate_validated_do_not_apply_mid_B2.md`
   (+ `CANDIDATE_2026-08-14_evidence_base.json`). Key resize: the fleet banned set
   already covers unarmed sites; the numeric arm cannot be dry-run (claimscan runs
   `ScanBannedClaims` only — induced proof); the high-value move is reusing
   mortgagecalculator's GOV.UK SDLT facts against `bugs_open/225`.
5. **Roll 2026-08-15** (build `0115f2b45`, both replicas, proven at binary): the
   section-editor **mode split is LIVE** (`content_edit_mode`/`updated_field_count`/
   `total_field_count` on every `content_edit` result from now on; forward-only; filter
   `edit_type` first).

## OPEN OWNER DECISIONS (ask, don't assume)

- **D3's two one-field edits on loancalculator:** (a) `strategy.value_proposition`'s
  *"lenders have no incentive to volunteer"* clause → restate as the lever; (b) remove/
  replace the *"reveal the true cost of credit"* characteristic exemplar (the family the
  owner rejected 08-08; exemplars beat rules). Regenerate `formatted` via
  `datahelpers.FormatContentDirection` if (b) touches content_direction.
- **evidence_base sourcing:** the £5–7k facts cite the site's own pages (self-referential
  provenance). Options in the CONTRIB §"open question". Apply only after LMC's B2 lands,
  by supersede, watching the first sweep.

## NEXT WORK, in the standing order

1. **`bugs_open/033` — the human-review queue has no working surface.** This is stage 2's
   BLOCKING dependency (owner D2, 08-12: no unreviewed auto-rewrite, so stage 2's output
   parks in a queue nobody can read; `voice_tells`: 34 parked, 1 ever closed, by machine).
   Read 033 + `who-owns` before starting; it may be another lane's.
2. **Stage 2 build** (only after 1). The accumulated constraints, all measured:
   - **Proof case committed** (owner ruling 08-12): 6 links missing from LMC index
     `prose-0`; pass = `loanandmortgagecalculator_couk/gate_page_links.py` exits 0;
     must-not-change list in NOTES §"OWNER RULING". Fixtures:
     `loanandmortgagecalculator_couk/acceptance/BASELINE_2026-08-12_*`.
   - **Page-scoped READ, section-scoped WRITE** (the fleet lane's blindness finding).
   - **FORBIDDEN input: the stage-1 brief** (PLAN §1 corollary — it inherits the framing).
   - **Prefer `field_updates`** (blast-radius mitigation, NOT a guarantee — PLAN §10; now
     measurable via `content_edit_mode`); **type gate** via
     `datahelpers.SchemaContentFields`, `type IN ('array','list')`, REPORT `fromLegacy`
     and coverage (PLAN §9/§10 + addendum; `bugs_open/265`'s stale-extinction landmine).
   - **Set preservation is mechanical, never instructional** — proven 4× (rounds 4/6/7 +
     the 08-13 arm run, which ADDED 2 links against instruction).
   - **Locks, not instructions, protect approved copy** — proven 3×.
3. **Phase 4 acceptance checks** (fact-inventory diff; markup parity; the type gate) —
   induce each before trusting; a green gate that cannot fail is the lane's twice-hit
   armed-but-inert shape.

## Standing cautions for the next session

- LMC: never fire `run_improvement_sweep_once.sh` (promotes all `detected` rows); the
  oneshot discovery envelope is the read-only route. Check lane activity before ANY write
  to their site — they were mid-B2 batch 2 as of 08-14 evening.
- The capture-only arm harness (this lane's verification workhorse):
  `loancalculator_couk/voiceh_rewrite_v3.sh` + `SRC_ITEM` override + timed locks +
  `page_components_bak_*` + read `llm_call_log` via the CHILD orch id; cancel the
  dispatched item post-capture. Worked examples in NOTES 08-13/14.
- Timed locks on loancalculator index (prose-1/2/4) from 08-13 have EXPIRED by now —
  re-lock before any new capture run there.
- Concurrent sessions are fast here: 264 was fixed by another session within a day, with
  the identical design. **Re-verify "X does not exist" claims from NOTES against the live
  DB before building X** — this lane's whole history is that lesson.

## The five living docs

PLAN (design + §6 revised phasing + §8 decisions + §9/§10 stage-2 gates) · NOTES (the
evidence log — read the 08-13 → 08-15 tail first) · README_where_we_are (owner's log) ·
SUMMARY_2026-08-12 + SUMMARY_2026-08-14 (the series) · this HANDOFF.
