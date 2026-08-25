# HANDOFF 2026-08-25 — bugfix_380_claims_fail_open: the bug is CLOSED; this is what a new chat needs

**Cold start = this file.** Then `NOTES_claims_fail_open.md` (technical record, newest at the bottom),
`PLAN_2026-08-24_claims_fail_open.md` (decisions D1-D4 and their reasons), `RUNBOOK_claims_fail_open.md`
(every query that was hard to get right), `README_where_we_are.md` (the owner's prose log),
`SUMMARY_2026-08-25_claims_fail_open.md` (the milestone read-out). Bug file, now closed:
`bugs_closed/380_HANDOFF_2026-08-24_a_site_with_no_evidence_base_gets_no_fact_assignment_and_no_claims_audit_so_the_writer_is_freest_where_it_knows_least.md`.

## 1. State `[MEASURED 2026-08-25 ~10:00Z, at the running system]`

| piece | where | live? | proof |
|---|---|---|---|
| Auditor runs cold, fails closed, receipts every run | mig 597 (+601 extraction fix) | YES since 08-24 16:20Z | 18 COMPLETED runs in 24h, 0 FAILED; garden-tools cold run's first two findings = the owner's two quoted sentences (corr bcf23316) |
| Planner: object form with `facts: []` on factless sites | mig 598 | YES since 08-24 16:21Z | anchors read back; next greenfield build should show `site_plan_sections.assigned_fact_ids = '[]'` everywhere (not yet observed — no greenfield build since) |
| Writer: no-register / no-operating-history arm | mig 599 (owner read the plaintext, "approved") | YES since 08-24 20:10Z | 150 of 150 `generate_content` calls since carry `## Operating history: NONE RECORDED` (the 71 `rewrite_negations` calls are a different prompt) |
| Auditor rotation, 3600s / 7-day | mig 600 | YES since 08-24 19:04Z | 15 sites selected overnight, 15 receipts, 8 new `claims_llm_*` items + 4 clean |
| Go practice-claims family + `operating_history` attestation + number-scan arming guard | commit c9cd817d9 | YES — chassis **v1.0.1337**, pods 09:27Z, provenance `4c996e1b5`, `c9cd817d9` is an ancestor; binary probe mine=1 / present-control=1 / absent-control=0 | **not yet positively exercised**: the 3 builds since the roll (loancash tool pages) carry zero practice sentences, so their silence is expected |
| Seed for the auditor | `claims_verification/SEED_claims_auditor.sql` | n/a (bring-up artefact) | regenerated from the post-601 live row |

Council: config slice `e684fc8d` APPROVED r1 (5 advisories, all answered in NOTES); Go slice `1d87615f`
APPROVED r1 (1 medium: a human re-check of the `ParseEvidenceBase` widening's consumer list — the list is
in the Go submission's risks §4 and NOTES). Commits: `856d0e1fd` config, `c9cd817d9` Go, `171ffed55` /
`ac6334456` / `ff9c55cb6` docs, `eb347dc12` 599 release, plus this close-out.

## 2. The ONE verify-later (not a gap — a demand control that has not had its input yet)

When any register-less page carrying a practice sentence is rebuilt (candidates: cookly.uk `about` —
"How we test what goes up"; dartsonline.com `shipping-returns` — "We test barrel profiles"; idea.uk
`about`; anything on garden-tools.uk — but do NOT rebuild garden-tools to make this pass, the loanzy lane
keeps it as an untouched measurement), its `validate_page_content` result should carry a
`practice_claim` issue at severity `warning` and the page should still deploy (warnings never affect
validity). Check:
```sql
SELECT o.created_at, s.domain, o.collected_data->'input_data'->>'page_name',
       (o.collected_data::text LIKE '%practice_claim%') AS practice_warning_present
  FROM orchestration_states o JOIN sites s ON s.id=o.site_id
 WHERE o.owner_agent_type='page-build-handler' AND o.created_at > '2026-08-25 09:27+00'
 ORDER BY 1 DESC;
```
Before reading a `false` as a dead gate, run claimscan on that page's components (RUNBOOK §5's export,
then `go run ./cmd/claimscan -components <tsv>`): if claimscan prints no PRACTICE line either, the page
had nothing to flag. `orchestration_states` reaps on a 24h clock — check within a day of the build.

## 3. What is NOT on this lane any more (residuals, each with its home)

- **`bugs_open/386`** (unowned, filed by the 364 lane): refreshing a counting fact convicts every page
  still rendering the old value. The 7-day rotation now AMPLIFIES this — a rotation finding whose
  `nearest_fact_id` has a `verified_at` newer than the page's render is a stale render, not an
  invention (CLM-027 relations). The durable fix is 386's (re-render on fact refresh).
- **`bugs_open/033`**: the `needs_human_review` queue the auditor's findings feed (now 8 more
  `claims_llm_*` items, one per audited site with findings) has no working surface. The findings are
  the demand signal the owner asked for; reading them is the missing half.
- **Skip-as-success census** → handed to the `bugs_open/354` lane (owner of the "a skip is a third
  state" seam): 38 conditionals fleet-wide route a branch to a `complete_workflow` step; ~6 are the 380
  shape (`site-adoption-agent.check_crawl_content`, `tool-auditor`/`tool-improver.check_tool_found`,
  `internal-linker.check_target_found`, `page-build-handler`/`tool-recreation-handler.check_page_found`).
  CONTRIB file in this dir. Not patched here.
- **Greedy-first Postgres regex on four more agents** (`meta-description-backfiller.load_pages_missing_meta`,
  `webdesign-agent.load_decisions`, `internal-linker.load_target_page`, `visual-design-auditor.
  load_design_context`) — LANDMINES entry + addendum; their lanes own them. The mechanism was first
  fixed by migrations 517/518 (`bugs_open/320` lane); a shared `page_visible_text(uuid)` TEXT function
  (517 has only the length) would let 601's query stop being a third formulation.
- **Slice 2b**: wiring the practice family into the discovery check (`check_unverified_claims.go`) at
  severity medium. Needs its own council round; the input it was waiting for is measured — the full-corpus
  dry run: 12 findings / 1,867 components on 6 sites (7 garden-tools, 3 true practice on operating sites,
  1 false positive `idea.uk` "how we test your idea"). Cost is bugs_open/033's queue.
- **"Source more"** — wiring `evidence-researcher` into the greenfield chain. OWNER DECISION: unattended
  research breaks the agritec lane's RUNBOOK §9 (every research run needs a human read — 4 of 5 facts
  unusable on a run that reported COMPLETED). The cold audit's findings lists are the demand signal.
- **True-practice sites** (leopardessconsulting.co.uk "We test every workflow on our own sites",
  loanandmortgagecalculator.co.uk "we test them"): they will carry a `practice_claim` warning on rebuild
  until someone records an `operating_history` attestation (`attested_by`, `attested_at`, `evidence`) in
  their `evidence_base`. Told via CONTRIB notes; no lane has done it yet.
- **The LLM item has no `spec.page_id`** and parks under the revalidator's `spec_no_page_id` arm
  (pre-existing shape). A per-page follow-up would let the revalidator close it automatically.
- **Cosmetic**: `validate_page_content.go` `practiceClaimsSeverity` logs a variable named `raw` (step
  config, not model output) — the pattern check flags it as `logged-model-output`; rename to
  `configured` when the file is next touched.
- **RFC_003 §8 Q1** (may a never-opted-in site have a build refused?) stays open with `warning` as the
  interim; the flip lives in `practice_claims_severity` per step and goes to architecture review fleet-wide.
- **`cmd/config-key-audit`'s `TestBudgetCronCountsLiteralMatchesTheRegistry` is red at HEAD** for
  `create_tool_component` (4 vs 5) and `deploy_tool_to_site` (3 vs 4) — another lane's Optional keys
  without a `check.py` regeneration. Not this lane's; reported.

## 4. Traps a new session will hit, in the order it will hit them

1. **Do not read `agent_definitions.updated_at`** — 199 of 200 rows share one microsecond, rewritten by
   a bulk touch at least twice a day (381 lane's LANDMINE). Verify a migration by a needle in the live
   text (RUNBOOK §5).
2. **The committed prompt text is not the live prompt** — generate any human-read plaintext from a live
   dump; the v5 file the owner approved is `brochure_component_library/sql/page_content_writer_prompt_v5_
   2026-08-24.txt`, the live-after-599 dump is the `v5b` file beside it (delta = the 381 lane's rules 9/10).
3. **`string_agg` + `regexp_replace` in SQL** — the greediness rule (LANDMINES). Strip per row, order the
   aggregate, assert a known sentence survives.
4. **A `kubectl exec … psql` export can truncate with only a stderr line** — count rows against the DB
   before trusting a dry run (it cost this lane a wrong calibration number once).
5. **`ensure_site_record` creates sites by domain** — any hand dispatch must refuse unknown domains
   (`TRIGGER_claims_audit.sh` does).
6. **`orchestration_states` reaps on a 24h clock** — measure runs via `llm_call_log` (the training
   corpus; never delete from it) and the doc_notes receipts, not via the orchestration table, after a day.
7. **Shell `timeout` on a kubectl-exec psql orphans the server backend**; a dead-client `COPY` sits in
   `ClientWrite` holding locks (85 waiters on 08-24) — `pg_terminate_backend`, not cancel.

## 5. Falsifiers for this handoff

- A `page-build-handler` run on a page claimscan flags with PRACTICE that shows no `practice_claim`
  warning → the gate is not wired as believed; re-read `validate_page_content.go` `checkClaims` branch.
- A rotation tick with no receipt for the selected site (RUNBOOK §3's `selected_but_no_receipt`) → a
  FAILED run; look in `agent_error_log`.
- A `generate_content` call after 599 without the arm on a site with no `operating_history` → the
  template was rewritten wholesale by someone (check the 599 needles).
- Anyone applying `599_…_HOLD` again — it is not held; the file is `599_page_content_writer_no_register_arm.sql`, applied and recorded.
