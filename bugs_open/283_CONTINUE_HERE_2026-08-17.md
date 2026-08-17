# 283 — CONTINUE HERE (2026-08-17). Converter LIVE on v1.0.1307; CANARY IN FLIGHT (§5). §1–§3 predate the roll — §5 supersedes their "waits on a roll".

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

---

## 5. UPDATE 2026-08-17 evening — the roll is REAL this time, and the canary is IN FLIGHT

**`v1.0.1307` verified at the artefact**: bumped tag, pod digest `8339bdbd…` == local digest,
revision `a6d1c53c0`, and **both `5b30a831b` (detector fix) and `b7b396cb3` (converter) are
ancestors**. The converter is live on the fleet for the first time.

**Canary dispatched through the framework** (step 1 of §3, done):

- work item **`38efde3b-10df-40c5-b3dc-691dddcd57b9`**, `item_type=instance_scope_conversion`,
  `status=triaged`, `handler_agent=component-template-fixer`, `item_key=instance-scope:24faa765`
- target: `tool-css-unit-converter`, component row `24faa765…`, ONE page on webdesign.co.uk —
  **the pinned-fixture component**, so the expected result is predicted to the digit:
  `fixed:true, ids_declared 12, id_attrs 12, get_element_by_id 11, id_ref_attrs 6, hash_refs 0,
  data-target ×5`, then the workflow's `check_needs_rerender` raises the rerender automatically.
- pre-dispatch checks done: no in-flight work on the component (one PARKED `vision_finding` —
  a mobile-layout CSS gap, pre-existing, orthogonal to id namespacing, noted in the item's spec
  so the served-page diff is not misattributed); chassis restart was >1h before dispatch.
- ⚠ webdesign.co.uk is the TOOL REBUILDS lane's territory (blocked on a roll at the time). A
  native rebuild of this tool would REPLACE the template and take the conversion with it
  (regeneration replaces; rerender merges) — acceptable for a canary, recorded here so nobody
  reads a later disappearance as a revert.

**How to verify the canary (next session, in order):**
```sql
SELECT status, result FROM site_work_items WHERE id='38efde3b-10df-40c5-b3dc-691dddcd57b9';
-- ⚠ a COMPLETED item's result may be the SPAWN record — read the orchestration:
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'id' = '38efde3b-10df-40c5-b3dc-691dddcd57b9';
SELECT count(*) FROM content_components WHERE is_active AND html_template LIKE '%{{.InstanceID}}%';
-- expect 1 after conversion; the template itself:
SELECT html_template FROM content_components WHERE id='24faa765-845e-4d1d-b7df-d71a06a0a617';
-- and the doc_notes entry the workflow appends (subject_type='pipeline', subject_key='build',
-- created_by='component-template-fixer', newest)
```
Then diff the SERVED page once the rerender lands. **Tomorrow 07:40 UTC the
`instance-token-adoption-check` TRIPS (adopters 0→1): expected, acknowledged by RFC_034 DECIDED —
retire the CronJob after observing the trip, do not treat it as a defect.**

---

## 6. CANARY RESULT — converted EXACTLY as predicted; rerender fan-out in flight

**18:51:54 UTC: item `38efde3b` claimed and completed in the same second.** The fix result, against
the prediction made from the pinned fixture before dispatch:

| field | predicted | actual |
|---|---|---|
| `fixed` | true | **true** |
| `ids_declared` | 12 | **12** |
| `id_attrs_renamed` | 12 | **12** |
| `get_element_by_id` | 11 | **11** |
| `id_ref_attrs` | 6 | **6** |
| `hash_refs` | 0 | **0** |

Verified at the artefact, not the status: the live `html_template` carries
`id="{{.InstanceID}}-input-value"` and `data-target="{{.InstanceID}}-result-px"`; snapshot v2 in
`component_versions` (`change_source='scope_component_instance'`); **adoption count is now 1 of
244** — so the 07:40 UTC tripwire WILL trip tomorrow, which is the RFC_022 expiry doing its job.
The workflow's own `doc_notes` entry is written and accurate (created_by='component-template-fixer').

**The rerender is a FAN-OUT, and that is worth knowing before reading any status:** the workflow's
`needs_rerender` item completed by spawning a **111-page site-wide rerender batch**
(`batch_id 486e96c9…`), not by rerendering the one page. At 19:05 the batch stood 58 complete / 52
queued, with the canary page's own `page_rerender` item still `triaged`. The stored
`rendered_html` therefore still carried the OLD ids at that moment — **"complete" on the
needs_rerender item is not a repaired artefact**, it is a dispatched batch. A background watcher is
armed on the moment the stored render flips; the served page
(`/tools/css-unit-converter/index.html`) follows on deploy.

**Still owed on the canary before calling it done end-to-end:** stored render flipped → served page
carries `c-tool-css-unit-converter-…` ids → the page still works by hand (one click of a copy
button exercises the renamed `data-target` chain). Then release the HOLD seed for the other 68.
