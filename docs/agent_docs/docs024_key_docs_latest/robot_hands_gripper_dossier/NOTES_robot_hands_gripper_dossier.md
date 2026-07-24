# NOTES — gripper dossier pilot

*Technical log. Append-only, newest at the bottom. Missteps included.
Design of record: `DESIGN_2026-07-24_gripper_dossier_pilot.md` (consolidated
from two grounded design passes; §5 seam reconciliation wins over either
half where they disagreed).*

---

## 2026-07-24 — workstream opened; scoring action built

Parent: per_site_ai (PLAN D15/D17), features_open/013. Owner unpaused after
robot-hands-site-fixes closed (R1–R9). Pilot scope locked: chat intake
(island) → pull → report-builder workflow → deployed report page + status
sidecar → island emails link.

### Owner decisions (2026-07-24, AskUserQuestion)
Shared sender (robot-hands@contactforsales.com) · prod E2E fixtures approved
(clean up after) · soft-launch unlinked · $50/mo cap on the new Anthropic
key. Defaults accepted: §5.3 mapping table, UUID+noindex, 24h/90d retention,
max_attempts=1. Tunnel VERIFIED LIVE (tools.apis.uk → 404 from island Caddy;
/api/v1/tools/health → 502 = no engine container yet, correct). Remaining
owner action: issue the $50-capped key.

### Built: score_grippers (first task)
- `platform/orchestration/actions/score_grippers_action.go` + registry entry
  + `score_grippers_action_test.go` (all green) +
  `sql_for_agents/204_robot_hands_matchmatrix_normalized_specs.sql`
  (NOT yet applied; pre-image safe — names no actions).
- Port source: dumped the live tool template (32,228 chars) to scratchpad;
  physics ported line-for-line (MATERIALS μ/ferrous map; dyn=m·a·S;
  fJaw=dyn/(μ·n); fDir=dyn; mEq=dyn/9.81; ipRow only under a requirement;
  unknown-never-pass; conflict note; verdict thresholds incl. the 1.25×
  marginal band; rank/headroom sort; the tool's exact no-match sentence).
- Input hardening beyond the tool (server context): material accepts μ key
  OR name alias; ip accepts "IP54" form; cycle_rate→safety tier (≤10→2,
  ≤30→3, >30→4, explicit safety_factor wins) per DESIGN §5.3; malformed
  spec = hard error (routes to error_step, never a guessed default).
- **Misstep, caught by the test run**: my conflict-note test asserted
  impliedMu = (mass·a·S)/(force·n) = 0.15; the tool actually derives it from
  the PUBLISHED PAYLOAD, (payload·a·S)/(force·n) = 0.26. The code was right
  (ported faithfully); the test expectation was wrong. Lesson: when porting,
  test expectations must be computed from the reference implementation, not
  re-derived from what the formula "should" be.
- fact_block contract: opens with the mandatory sentence when match_count=0;
  marks unpublished figures "NOT PUBLISHED by the manufacturer — say so if
  mentioned; never estimate it"; carries per-candidate source_url +
  verified_date; carries the substituted formula strings (these double as
  the discriminating E2E artefact — no other page carries them).

### Verified this session (live)
- 10 products rows, all with source_url + verified_date, none with
  matchmatrix block yet (seed 204 will add).
- Island tunnel public; API path answers 502 (nothing behind Caddy yet).

### Next
1. Commit (3 narrow pathspec commits: platform code+seed / this workstream's
   docs / the per_site_ai session docs) + council submission for the code.
2. create_report_page + report_charts.go (SVG) + verify_report_prose +
   pull_report_requests + emit_report_status_files.
3. Agent seeds 205–208; island service cmd/gripper-intake.
