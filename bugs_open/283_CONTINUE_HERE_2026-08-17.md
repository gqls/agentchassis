# 283 — CONTINUE HERE (2026-08-17). Converter BUILT and approved; execution waits on a ROLL WITH A BUMPED TAG.

**Read the case file first** (`bugs_open/283_HANDOFF_…_element_ids_are_literal.md`) — §11–§12 are
the current state; §3a of `architecture_review/RFC_034` has the only trustworthy corpus numbers.
Supersedes `283_CONTINUE_HERE_2026-08-16.md`. Lane docs:
`docs/agent_docs/docs024_key_docs_latest/bugfix_283_component_instance_scope/` (the RUNBOOK's
deploy-digest chain in §1 is the first thing to run).

---

## 1. State in one paragraph

Owner ruled (RFC_034 DECIDED): **hybrid shape, LMC first, conversions THROUGH THE FRAMEWORK.**
The corpus is **91 component rows** (94 pages, 22 domains): **66** need only the deterministic
rename, **25** (the 23 LMC calculators + 2 tools) also need judged script work. The deterministic
converter is **built and council-approved** — fix_type `scope_component_instance` (CLC-017),
whose gate mechanically refuses the ids-only half-state — but it is **in no running image**:
the 2026-08-17 14:43 "fresh build" restart served yesterday's cached digest because the rebuild
reused tag `v1.0.1305`. **Nothing has been converted; 283 stays OPEN.** Council trail: rounds 2
and 3 both approved on correlation `07635a2f-3605-4e67-9a6d-7636b07f16ca` (round 3: "all
reviewers approve").

## 2. ⚠ Traps that will bite the next session first

- **The running pods are NOT the local image.** Pod digest `f90a7e88…` (revision `6a782274b`,
  08-16) vs local+pushed `v1.0.1305` digest `6039e19c…` (revision `89a0cbeb7`, 08-17 14:30, which
  DOES contain the detector fix `5b30a831b` and converter `b7b396cb3`). Same tag, two images —
  the CLAUDE.md same-tag cache trap, live. **Remedy: next roll under a BUMPED tag; never a
  one-service apply.** Verify by digest equality first, label second, ancestry third (RUNBOOK §1).
- **Only §3a's numbers are true.** 24, 88, "~30", "4 shared across 5 domains" all appear in
  history with visible corrections. 66/25 is the corrected split; re-run `cmd/instanceaudit`
  rather than trusting even that (the corpus drifts — 243→244 components in one day).
- **Convert by `content_components.id`, never `function`** — 4 functions carry forks; 9 rows
  silently skipped otherwise.
- **A `fixed:true` from the converter changes NO served page** until that page rerenders and
  redeploys. Sequence per component: convert → rerender → redeploy.
- **The first shipped conversion trips `instance-token-adoption-check`** (daily 07:40 UTC).
  That trip is the RFC_022 exception expiring — an owed architecture acknowledgement, NOT a
  defect. Do not "fix" it by reverting; retire the CronJob once acknowledged (CLC-016).

## 3. Do this next, in order

1. **After the next roll**: digest-verify, then run ONE canary conversion through the framework —
   file a work item for a single non-LMC row from the 66 (small blast radius, no oracle needed for
   a mechanical-only row), watch it convert, rerender, redeploy; diff the served page.
2. **Write the seed** for `instance_scope_conversion` work items (one per row; the 66 first or
   LMC-first per the ruling — LMC's 23 are all in the JUDGED pool, so the mechanical 66 can
   proceed while the judged pipeline is designed; that ordering respects "LMC first" for the
   judged work, which is what the oracle protects).
3. **Rebaseline `b2_verify`** before any LMC row converts; move `oracle.py` selectors in lockstep.
4. **Design the judged pipeline** for the 25: LLM rewrite per component (IIFE + `addEventListener`
   rewiring), the converter's gate as acceptance, byte-level truncation check on every result
   (`output_tokens == max_tokens` means CUT — `bugs_open/012`).
5. **Keep the trail on one correlation** (`RESUBMIT_CORR=07635a2f…`) for any further platform
   change in this lane.

## 4. Where everything lives

| thing | where |
|---|---|
| transform + gate | `platform/orchestration/actions/component_instance_conversion.go` |
| framework seam | `fix_component_template_action.go` → `fixScopeComponentInstance` |
| pinned live fixtures | `actions/testdata/instance_conversion_{mortgages_repayment,css_unit_converter}_283.html` |
| corpus classifier | `cmd/instanceaudit` (export query in its header; REFUSES an empty export) |
| detector + token seam | `component_instance_scope.go` (CLC-014) |
| adoption tripwire | CronJob `instance-token-adoption-check` (CLC-016) |
| decisions | `RFC_034` (DECIDED), `RFC_032` (open — ComponentID unification) |
| judged pool list | RFC_034 §3a / NOTES: 23 `loans-*`/`mortgages-*` + `tool-archetype-clash-calculator`, `tool-bayesian-ranking` |
