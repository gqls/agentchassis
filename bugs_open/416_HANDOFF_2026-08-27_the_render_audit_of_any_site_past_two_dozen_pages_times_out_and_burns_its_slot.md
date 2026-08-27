# 416 — the render audit of any site past ~two dozen pages TIMES OUT at its step budget, discards the adapter's finished work, and burns the site's rotation slot for 3 days

**Filed 2026-08-27 by the `bugfix_390_cascade_attribution` lane** (found while grading P4 — vonc.com's
audit was the lane's formal test and it timed out). **Diagnosis loop running: 090
`RUN_CORRELATION_ID=0df41fb3-e386-416c-bffc-1d338fb82e6d`** — root-cause section below is a
HYPOTHESIS until that verdict lands; the damage and census are first-hand measurements.

## 1. The mechanism (hypothesis pending 0df41fb3, evidence below)

The `render-audit-agent` workflow's `audit` step dispatches the whole site to the render-audit
adapter in ONE request and awaits ONE reply. The step declares **no `timeout_seconds`** (checked
in the live `agent_definitions` row: every step's config lacks it), so a chassis-side default —
**~185 s by observation** (17 of 17 timed-out orchestrations ran 3m03–3m13 end-to-end) — governs
the await. The adapter audits ~**8 s/page** (vetcomparison 15 pages: complete at 1m59; vonc 28
pages: adapter produced its reply at **+3m47**, 45 s AFTER the step gave up). So any site past
roughly **23–25 deployed pages** cannot finish inside the budget, the finished reply is
discarded, and the orchestration exits via `complete_error`.

**Two structural aggravators:**
- **The rotation stamps at SELECTION** (`site_discovery_rotation.last_selected_at`, written by the
  pre_query in the same statement that selects) — so a timed-out audit consumes the site's turn
  and it waits 3 more days to fail the same way.
- **The failure is recorded as `COMPLETED` with `error` NULL** (`bugs_open/354`'s defect, this
  audit workflow is its worked example) — so nothing looks wrong anywhere except
  `agent_error_log` and the absence of findings, and an absence has no row.

## 2. Damage [MEASURED 2026-08-27 13:15 UTC]

- `agent_error_log` (`agent_type='render-audit-agent' AND error_code='TIMEOUT'`): **53 timeouts
  across 8 audit days, 2026-08-17 → 2026-08-27** — 4/7/5/9/3/10/5/10 by day, every audit day in
  the window. Domains: webdesign.co.uk (151 pages), finetuning.uk (55), ai-agent-orchestration.com
  (46), mortgagecalculator.co.uk (42), gaswholesalers.com (41), robot-hands.com (38),
  dartsonline.com (38), leopardessconsulting.co.uk (36), idea.uk (34→28 audited), fundamentallyai
  (26), lendzy (27), vonc (28), loancalculator (28), loanandmortgagecalculator, gamesdesign,
  adversecreditmortgage, loancash. **These sites have had no successful render audit for at least
  two weeks** (bounded by the error-log window read; older retention unchecked).
- Sites ≤15 pages complete comfortably (remortgagecalculator, garden-tools, cookly, webdesign.uk
  13, vetcomparison 15 — all `complete` on 08-26/27, all under 2m04).
- **Consequences beyond missed findings:** the WII-016 retraction runs in `write_findings` — a
  timed-out audit retracts NOTHING, so completed repairs on affected sites are never confirmed
  and stale findings never close. The 390 lane's cascade-attribution telemetry never populates
  for these sites. Every fix that verifies "at the next audit" silently exempts the big sites.
- **NOT a regression from the 390 attribution roll (08-25 23:11):** the census predates it by a
  week+. Attribution adds per-element CSSOM work and MAY have nudged the ceiling down slightly —
  vonc (22 pages on 08-24, no timeout; 28 pages on 08-27, timeout) and lendzy/idea are new
  entrants whose page counts ALSO grew — the marginal-cost question is for the diagnosis loop,
  and commit 2's council submission named this exact risk and this exact confound in advance
  ("pre-existing budget pressure, unfiled").

## 3. How to verify

```sql
SELECT occurred_at::date, count(*) FROM agent_error_log
WHERE agent_type='render-audit-agent' AND error_code='TIMEOUT' GROUP BY 1 ORDER BY 1;
-- a timed-out run: current_step='complete_error', __step_errors->'audit', ran ~3m05:
SELECT current_step, updated_at-created_at, collected_data->'__step_errors'->'audit'
FROM orchestration_states WHERE owner_agent_type='render-audit-agent' ORDER BY created_at DESC;
-- the adapter finishing late (per-pod logs, structured field):
kubectl -n ai-persona-system logs <render-audit-adapter-pod> | grep '"Successfully produced'
```

## 4. Fix candidates, ordered by what closes the door

1. **Make the await budget a function of the work**: the dispatch metadata already carries
   `pages_total` BEFORE the wait begins — size the step timeout from it (pages × measured rate ×
   headroom), or split the audit into bounded per-page/per-batch requests so no single await can
   exceed any constant. Splitting is the version where the bad state is unrepresentable.
2. **Stop burning the slot on failure**: move the rotation stamp from selection to
   `write_findings` success, or reset `last_selected_at` on the `complete_error` path — otherwise
   any future timeout still costs 3 days of coverage per occurrence.
3. **Interim, config-only, live immediately** (if the chassis honours step-level
   `timeout_seconds`): set it on the `audit` step to cover the current worst site
   (151 pages × ~8 s ≈ 21 min — verify the chassis cap first). Cheap, reversible, but a constant
   that page growth will eventually beat — pair with (1) or (2).
4. The recording half (COMPLETED/error-NULL) is `bugs_open/354` — fixing it does not fix this,
   but until one of them lands this failure is invisible outside `agent_error_log`.

## 5. Who this touches

Render-audit rotation coverage (all consumers of its findings and retractions), `bugs_open/296`
(parked findings on affected sites can never be retraction-tested), `bugs_open/354` (this is its
worked example), the 390 lane (P4-formal on vonc blocked until fixed; noted/loanzy graded on
their own audits). Filed unowned beyond the 090 run; the 390 lane will grade any fix's effect on
its own verification but does not claim the fix.
