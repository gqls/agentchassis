# HANDOFF 2026-08-15 — continue here

**Lane:** `copy_quality_two_stage`. **State (updated ~22:30Z): the FOUR OWNER DECISIONS
LANDED this evening and 1+2+3 are ALL APPLIED AND PROVEN; only the stage-2 build (4,
ruled BUILD IN PARALLEL with the 033 thread) remains.** NOTES 2026-08-15 (evening + late
evening) carries the evidence:
- **1+2:** loancalculator's specs were REGENERATED the same morning by the rebuild lane's
  fire and the offending clauses survived verbatim, so the edits were applied to the
  fresh rows (old: strategy `b82e2c7e…`, content_direction `a1feaaa7…`; new: `bca3b9ee…`
  / `7f39172e…`; rollback = flip `is_current`). ⚠ A future re-fire of the research
  agents clobbers both corrections — CONTRIB filed in `loancalculator_couk/`.
- **3:** LMC's register is LIVE (row `7268d235…`, its first): mc's 13 GOV.UK SDLT facts
  verbatim, £5–7k facts deferred to sourcing. Flood risk MEASURED AT ZERO (claimscan
  numeric arm 0/82 strictest mode, induced controls firing) — and note the correction:
  the numeric arm always WAS dry-runnable; 08-14's contrary finding is refuted in place
  (WRONG_CALLS 08-15). 225's calculator fix has been live since 08-09; the register is
  regression insurance, not a live catch. First scheduled sweep reaches LMC ≥08-16
  ~09:49Z stamps-permitting; expectation is 0 claims items — any finding = a page
  changed after 08-15 late evening.

## The four decisions as originally posed (RULED — kept for context)

1. **loancalculator `strategy.value_proposition`** — its final clause promises the site
   reveals what *"lenders have no incentive to volunteer"* (the insider-secrets family,
   the mc corrections' named anti-pattern). **Restate as the lever, or leave?**
   Recommended: restate. One field.
2. **loancalculator's characteristic exemplar** — `content_direction.example_phrases`
   still lists *"reveal the true cost of credit"* (the family the owner rejected 08-08)
   as a model. Exemplars beat rules, so this teaches the rejected register. **Remove, or
   replace with a lever-shaped equivalent?** Recommended: replace, not remove — the slot
   does teaching work. Regenerate `formatted` via `datahelpers.FormatContentDirection`
   or the edit is inert (the writer reads `formatted`).
3. **evidence_base sourcing for LMC** — the £5–7k facts in the candidate cite the site's
   own pages (self-referential provenance). **External source / register as the site's
   own estimate / open with mortgagecalculator's GOV.UK SDLT facts instead?**
   Recommended: SDLT first (verified, external, catches `bugs_open/225`'s live wrong
   figure), sourcing work later. Apply only after LMC's B2 lands, by supersede, watching
   the first sweep.
4. **Stage 2 timing** — `bugs_open/033` (the review surface, stage 2's D2 dependency) is
   being worked in a SEPARATE thread (owner, 08-15). **Build stage 2 in parallel so the
   two land together, or wait for the surface?** Recommended: parallel — stage 2's first
   output is the committed proof case, which the owner reviews directly either way; if
   033's thread stalls, stage 2 pauses exactly where it would have been.

Plain-prose versions of all four: `README_where_we_are`, 2026-08-15 entry.

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

## NEXT WORK once the decisions land

1. **Decisions 1+2** are minutes of config work each (supersede, never in-place; snapshot
   first; the `formatted` regeneration on decision 2 is load-bearing).
2. **Decision 3** waits on LMC's B2 finishing — check their lane activity first, as ever.
3. **Stage 2 build** (timing per decision 4; `bugs_open/033` is another thread's — do not
   start it from here). The accumulated constraints, all measured:
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
   - **Never compare prose to prose in an acceptance gate** — the 278 natural experiment
     (same section, fixed inputs, filled twice): titles identical, **2 of 4 card bodies
     diverged** with nothing wrong. Set/type/structure comparison passes that pair
     correctly AND still catches the duplication itself. Evidence banked in `278` §8,
     which also carries the instruction to capture a free third generation if their
     repro re-fills the section.
4. **Phase 4 acceptance checks** (fact-inventory diff; markup parity; the type gate) —
   induce each before trusting; a green gate that cannot fail is the lane's twice-hit
   armed-but-inert shape.

## NEW 2026-08-15 (late): webdesign.co.uk — the first live v2-era case, WAITING on a composition bug

Owner feedback (via the bugfix-122 front): the homepage copy "sounds like AI with 'this
not that'" — h2 *"A workbench, not a sales pitch"* + the same shape in the body. Verified;
NOTES 08-15 has the full record. The site has NO voice spec/gate (pure house-voice
governance) and the copy predates v2. **Action when unblocked:** voice-only
`content_rewrite` via `page-build-handler` (renders under v2 now). **Blocked on: `bugs_open/278`** — the duplicate is LOCATED in `site_plan_sections`
(`info-card-grid` planned twice in one transaction, N=1 fleet-wide); the "why" is
090-flagged and undiagnosed; likely fix is a plan edit + re-render. The 122 front pings
this lane on close. Fix composition FIRST or the two fixes muddy each other. Do not hand-write
copy (owner rulings 08-04/08-06), and do not touch the tan link colour (intended, 122's
ink canary).

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
evidence log — read the 08-13 → 08-15 tail first; corrections are marked in place) ·
README_where_we_are (owner's log; the four decisions in plain prose at the 08-15 entry) ·
SUMMARY_2026-08-12 / 08-14 / 08-15 (the series: why it's wrong → the fix ships → what's
left is choices) · this HANDOFF.
