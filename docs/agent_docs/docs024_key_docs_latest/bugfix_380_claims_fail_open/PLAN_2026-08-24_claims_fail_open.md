# PLAN 2026-08-24 — bugs_open/380: the claims layer fails open on the sites that need it most

Owning session: `bugs_open/380` (adopted 2026-08-24 from the `loanzy_uk_example_site` handoff,
which filed it UNOWNED). The full research record is in NOTES; this file holds the decisions.

## What we are fixing

Three mechanisms degrade to "no constraint" when a site has no `evidence_base` facts:
(a) nothing on the greenfield chain creates a register; (b) `build-site-planner` tells the writer
to use plain string sections ("no facts keys") and the writer gets NO Verified Facts block at all;
(c) `claims-auditor.check_opted_in` branches to `complete` — a skipped audit indistinguishable from
a clean one. The writer is freest exactly where it knows least (garden-tools.uk's invented review
methodology; loanzy.uk's invented credit broker).

## Corrections to the bug file, established before designing (evidence in NOTES)

1. The gate is on FACTS, not the register: `facts_text` is `string_agg` over `facts[]` → NULL on an
   empty array, so the four zero-fact registers are skipped too (33 of 48 sites unaudited, not 29,
   `[MEASURED 2026-08-24]`).
2. Candidate 3 ("mint an empty-but-present register") cannot work as filed: `ParseEvidenceBase`
   returns nil for it (CLM-005) and both prompt arms test facts, not the row.
3. The LLM auditor is an ORPHAN: no seed file (mig 350 records it), no schedule, no spawner, ONE
   `llm_call_log` row ever (2026-07-18, returned `[]`), `claims_llm%` items = 0 rows. Every
   `claims_unverified` item came from the Go discovery check. "Fleet claims coverage ~40%" was
   [INFERRED]; measured, the LLM layer's coverage is ~0%.
4. This bug was first found 2026-07-20 (`claims_verification/PLAN_2026-07-20_gaswholesalers_second_site.md`
   §3, the unbuilt `cold_audit` design, benchmark 174 assertions on gaswholesalers.com).

## Decisions (owner, 2026-08-24, via in-session questions)

| # | decision | why |
|---|---|---|
| D1 | **No minting, no backfill** — absence IS the cold posture | a shell register parses to nil (CLM-005), flips `rowExists` (29 sites of low-sev stat items), and reads as protection it isn't; `evidence_citations.go` mints the row at the first real fact |
| D2 | **Rotation: 3600s tick, 7-day per-site window**, never-audited first | ends the "cadence is an owner call" hold open since 2026-07-16; ~4 sites/day, ~120 Sonnet calls/month |
| D3 | **Go practice-claims family ships at default `warning`** (record, never refuse); escalation is per-step `practice_claims_severity`; a fleet-wide flip is RFC_003 Q1 → architecture review | RFC_023: a rule that can newly refuse is a contract change; bugs_open/364 is the false-positive cost |
| D4 | **Writer-prompt arm gated on an owner plaintext read** (`_HOLD` migration + committed v5 plaintext) | RFC_016 §5.2: any edit voids the 2026-08-09 v4 approval |

RFC_003 Q2 is thereby answered NO (append to the RFC); Q1 stays open, with `warning` as the interim.

## The four slices

- **597_claims_auditor_runs_cold_and_fails_closed.sql** (config, live on apply): delete
  `check_opted_in`; `load_evidence_facts.next_step → load_page_text`; remove
  `load_evidence_facts.config.error_step` (a DB error now FAILS the run — RFC_017); cold-register
  prompt arm (`{{if .evidence_facts.facts_text}}` roster `{{else}}` report practice/possession/
  track-record/named-relationship assertions, do-not-report list for could-framed/negations/quotes/
  industry statements); `ALLOWED ENTITIES` nil-guarded; page-text cap 3500→12000; doc_notes receipt
  per run (`pipeline`/`claims-audit`, categories `audit-ran|audit-findings`); explicit
  `recurrence_expected: false` on `request_claims_review` (mig 572's reasoning). Seed committed:
  `claims_verification/SEED_claims_auditor.sql` (the agent has none).
- **598_build_site_planner_distinct_no_facts_arms.sql** (config): the two identical `{{else}}` arms
  become distinct; both instruct the OBJECT form with `"facts": []` (verified end-to-end: that
  reaches the writer as `facts_scoped=true` with no Go change); the spec-absent arm adds the
  no-operating-history planning rule; rule 17's contradictory last sentence edited.
- **599_page_content_writer_no_register_arm_HOLD.sql** (config, HELD for D4): the scoped-empty arm's
  false clause fixed; an `## Operating history: NONE RECORDED` block (double-guarded — text/template
  errors through absent intermediates); the STRICT-RULE "say what we DO" qualified.
- **600_claims_audit_rotation.sql** (config, applied after the hand-dispatch proof): 590-template
  rotation, `claims-auditor`, 3600s / 7-day, shipped-page predicate, `locked_at IS NULL`, stamp in
  `site_discovery_rotation`.
- **Go commit (inert until image roll)**: `datahelpers/claims_practice.go` — first-person
  physical-practice family, exempted by an `operating_history` attestation in `evidence_base`;
  `ParseEvidenceBase` nil rule widened AND the numeric scan's arming guarded on register content
  (closes CGV-033's latent hazard); NOT unioned into the refusing set (mutation test pins it);
  gate wiring at `warning`; claimscan `PRACTICE` lines. Fleet dry-run before the image ships.

## Council

Two ordinary-gate submissions (RFC_023): S-config (597+598+600, 599 named as HELD) and S-go.
`Council-Submitted:` trailers on commits; 098 resolves on verdict.

## Not in this fix (named follow-ons)

Wiring `evidence-researcher` into the greenfield chain (unattended research breaks agritec RUNBOOK
§9's mandatory review — owner decision, the cold-audit lists are its demand signal); discovery-check
wiring of the practice family (slice 2b, needs the measured fleet count); RFC_003 Q1 (escalation to
refusal); the writer's standalone `plan_sections` step (unscoped by RFC_016 slicing).
