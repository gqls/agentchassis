# BUG 046 — the truncation fix stopped new damage; nothing ever repaired the old. 8 tools serve unterminated JavaScript on live customer sites

**Filed:** 2026-07-20 · travelling-docs thread
**Severity:** high — live customer-facing breakage on **6 domains**, present now,
and invisible to every check we have.
**Status:** OPEN (residual only). Detection **LIVE + ENABLED + PROVEN** (v1.0.1149
+ seed 186). **ALL 6 PLACED CASUALTIES REPAIRED & VERIFIED LIVE 2026-07-24** —
every live customer page that was serving unterminated JavaScript now serves
whole, balanced markup. Census 9 → 3; the 3 remaining are UNPLACED orphans
(0 page_components, serving nobody): `tool-archetype-clash-calculator-vonc-com`,
`tool-llm-cost-calculator-…-leopardessconsulting-co-uk`, `archetype-taster-quiz`
(section). Residual decision: deactivate vs regenerate the orphans. See the
UPDATEs below.

## UPDATE 2026-07-24 (later) — the batch: all remaining placed tools repaired & live

Owner-approved batch run (after the arena proof). The "6 remaining placed"
resolved to **4 components across 5 page placements** (the finetuning llm-cost
component is shared onto leopardessconsulting too; my earlier table header
mis-said 6 — the true reconciliation: 9 casualties = 6 placed components + 3
unplaced, and grip-force + arena were already done).

Recipe per component: tool-improver in-place rewrite → verify balanced + no
fabrication tells → section-editor `content_edit` delivery per placement →
verify live bytes.

| component | rewrite corr | template | live page | live script |
|---|---|---|---|---|
| drop-rate-tuner (gamesdesign) | `0ed21972` | 22,518 cut → 21,478 balanced | tools/tool-drop-rate-tuner.html | **3/3** |
| llm-cost (ai-agent-orchestration) | `e7ec3fe1` | 21,724 cut → 35,290 | tools/tool-llm-cost-calculator.html | **5/5** |
| llm-cost shared (finetuning) | `716203fb` | 17,828 cut → 31,324 | finetuning.uk + leopardess /tools/llm-cost-calculator.html | **5/5 both** |
| process-automation-scorer (leopardess) | `25019a4e` | 14,644 cut → 18,775 | tools/process-automation-scorer/index.html | **5/5** |

Delivery corrs: `3f5aa1d6` / `94434d81` / `564d1485` / `7e6c27e9` / `a228c040`,
all COMPLETED; all 5 `rendered_html` rows match the new templates (script 1/1,
`deployed`). All rewrites: zero fabrication tells (no PRNG/seed arrays, no
external fetch), all end clean.

Operational note: the 5 delivery publishes sat queued ~5 minutes behind another
session's council run (the bugs_open/030 serialisation) with **no orchestration
row** — the documented latency trap; waiting (not re-publishing) was correct.

---

## UPDATE 2026-07-24 — grip-force fully repaired LIVE; the delivery blocker (024) is gone

Two blockers I recorded on 07-21 are now **both resolved by other threads**:
- **bugs_open/024 (delivery) is CLOSED.** A tool page is `rebuild_policy='owned'`,
  so the generic rerender is (correctly) forbidden; the sanctioned delivery is the
  **section-editor** agent's `apply_section_edit` (`edit_type=content_edit,
  field_updates={}`) — a pure re-render from the *current* template, no LLM, git
  commit + deploy. Recipe: `features_open/009` + `bugs_closed/024`.
- **bugs_open/020 (tool-recreation fabrication) is CLOSED** — prompt fix (mig 183)
  + `check_tool_fabrication` gate live (v1.0.1150+). So regeneration is now safe.

**grip-force delivered LIVE via the section-editor** (corr
`06c6c158-4c0b-4eef-8479-9251c02480d1`, section-editor orchestration COMPLETED):
- `page_components.rendered_html`: 23,874 (script 1/0, damaged) → **23,526 (1/1,
  balanced)**; `build_status='deployed'`.
- **Live page** `https://robot-hands.com/tools/grip-force-friction-calculator/index.html`:
  `<script` 3 / `</script>` **3** (was 3/2). **The unterminated script is gone.**

First casualty fully repaired end-to-end (source + render + live page); proves the
delivery half of the recipe. Drive script:
`docs/agent_docs/docs024_key_docs_latest/truncation_casualties_046/scripts/deliver_via_section_editor.sh`.

### Repair recipe for the remaining 8 (all need regeneration — no intact version)
1. **Regenerate the template** — `needs_tool_recreation` → tool-recreation-handler
   (now fabrication-gated, bugs_closed/020). Produces a NEW tool (won't match the
   original design); for a broken tool, working ≠ broken is a win.
2. **Deliver** — section-editor `content_edit` (the proven recipe above).
   NB: a regenerated tool referencing an EXTERNAL `/tools/assets/{fn}.js` needs an
   assemble-only JS republish afterwards (gauntlet_dead_cta republish pattern) —
   `apply_section_edit` does not run `collectJSAssets`. grip-force's script is
   inline, so it needed none.

**Owner decision point:** whether to mass-regenerate the 8 live tools (LLM-heavy,
changes 8 live customer tools) or triage per-tool. Not auto-triggered.

---

## UPDATE 2026-07-22 — detection is LIVE, ENABLED, and PROVEN on real data

The chassis image carrying `check_truncated_component.go` reached production as
**v1.0.1149** (pod-verified: the literals `truncated_component query failed` and
`still truncated: unterminated` are in `/app/agent-chassis`; negative control 0).

- **Enabled:** seed `186_enable_truncated_component_check.sql` applied and
  ledgered (`schema_migrations`, applied_by=record-only). `completeness-discovery-agent`'s
  checks array now contains `truncated_component` (snapshot `b05773e0` taken).
- **Proven end-to-end** (not just deployed): a real `completeness-discovery-agent`
  pass on vonc.com (corr `c6721ab9`, orchestration COMPLETED clean) raised
  `truncated_component` item `ae5ab628` for `tool-arena-interface-vonc-com` —
  spec exactly as designed: `unterminated:["<script"]`,
  `intact_version_available:false`, `needs_human_review`, priority 35, the full
  restore/regenerate/remove `fix` note. The unplaced
  `tool-archetype-clash-calculator-vonc-com` was correctly NOT flagged (0
  page_components — the check sees only components on a site's pages, as
  documented). This is the "verify the failing branch, not the happy path"
  standard: the check's job is to detect, and it was induced against a real
  casualty and detected it.
- **Scope note for the census:** the sweep is per-site via the page join, so the
  3 truly *unplaced* casualties (archetype-clash-calculator, the leopardess
  llm-cost variant, archetype-taster-quiz — 0 page_components) will never be
  swept by a site pass. They are not serving visitors either; they need a
  fleet-level cleanup (deactivate or regenerate), tracked here rather than by the
  check.

Going forward the class is self-surfacing: each casualty on a live page becomes a
tracked `truncated_component` item as its site is next swept. Remaining work is
the *repair* (regeneration of the 8, owner-steered per 020) and *delivery* (024).

---

## UPDATE 2026-07-21 — bugfix-046 thread: sweep built, grip-force source-repaired, 8 remain

**Census re-grounded 2026-07-21: still exactly 9** (now 8 after the grip-force
restore below). Per-component facts added: 5 of 9 are on deployed pages; only
**grip-force had an intact prior version** (v2, 23,526 chars, balanced).

### Done

1. **Detection — candidate 2, the durable structural fix.** New discovery check
   `truncated_component` + verifier + drift-guard test, committed `1e5cb6fdc`
   (`platform/orchestration/actions/discovery_checks/check_truncated_component.go`).
   - Predicate: the 5-pair tag imbalance (`<script/<style/<section/<div/<fieldset`,
     open>close). Calibrated against the full live population 2026-07-21 — catches
     **exactly the 9 census rows, 0 over-fire.** Deliberately excludes
     `toolTemplateValid`'s ends-mid-token heuristic, which as a fleet *sweep*
     would flag **36** legitimate templates; a queue item must be high-precision.
   - **Not tool-scoped** (per the file's own warning): joins `content_components`
     at any level, so the `section` casualty is covered.
   - **Detect-and-surface** (`needs_human_review`, NO handler — the `dead_controls`
     pattern). It does NOT auto-route to a regenerator: the remedy varies, and
     tool recreation can **fabricate data** (`bugs_open/020`). The spec carries
     `intact_version_available` / `intact_version_number` so triage (restore vs
     regenerate) is one glance.
   - Verifier re-checks the current template with the same predicate; resolved
     when balanced or deactivated; a missing row is an error, never a false green.
   - **Inert** until an image roll **and** the enable seed
     `docs/agent_docs/sql_for_agents/186_enable_truncated_component_check.sql`
     (image-first) appends it to `completeness-discovery-agent`'s checks.

2. **grip-force restored at source (candidate 1's cheap first step).**
   `tool-grip-force-friction-calculator-robot-hands-com`'s `html_template`
   restored from its intact v2 (DB, live). Balanced (1/1), ends clean. **Census
   query now returns 8.** Damaged bytes backed up before the restore.

### NOT done — and deliberately not attempted by this thread

- **The live pages are still broken.** grip-force's page is `needs_rebuild` and
  the live URL still serves `<script`×3 / `</script>`×2 (2026-07-21). Restoring
  the template fixes the SOURCE; the live page only changes on the next
  **re-render** — and that delivery pipeline is `bugs_open/024`'s active, buggy,
  *owned* territory (defect 6 open). This thread did not touch it. The restore is
  a net improvement for delivery regardless: `toolTemplateValid` now ACCEPTS the
  template, so a re-render renders good bytes instead of carrying the cut.
- **The other 8 have no intact prior version** → they need **regeneration**, which
  is LLM-heavy and fabrication-guarded (`020`). Not auto-triggered. Once the check
  is live they surface as tracked `truncated_component` items; the owner/next
  thread decides regenerate-vs-remove per item.

### Verify what this thread changed
```sql
-- was 9, now 8 (grip-force dropped out):
SELECT cc.name FROM content_components cc
WHERE cc.is_active AND length(cc.html_template) >= 100
  AND (SELECT count(*) FROM regexp_matches(lower(cc.html_template),'<script','g'))
    > (SELECT count(*) FROM regexp_matches(lower(cc.html_template),'</script>','g'));
```
`go test ./platform/orchestration/actions/discovery_checks/ -run 'Truncat|Unterminated'`

Workstream docs: `docs/agent_docs/docs024_key_docs_latest/truncation_casualties_046/`.

---

## The gap in one sentence

`bugs_closed/012` fixed the *cause* of component truncation (the write guard +
`stop_reason` decoding + migration 168's token raise) and restored **one**
component — the loot-table balancer, by hand, because that was the one the
thread happened to be watching. **Nobody ever swept for the others.** They are
still damaged, still active, still deployed.

## Live evidence — not a DB inference

Fetched over HTTPS 2026-07-20, from the public sites:

| URL | HTTP | bytes | `<script` | `</script>` |
|---|---|---|---|---|
| `https://vonc.com/tools/arena/index.html` | 200 | 39,646 | **3** | **2** |
| `https://robot-hands.com/tools/grip-force-friction-calculator/index.html` | 200 | 38,656 | **3** | **2** |

One `<script>` is never closed on either page. The tail of both is
`… </style> </body> </html>` — the tool's own cut script block has swallowed
everything after it until the *next* `</script>`, so the tool's JavaScript never
runs and the markup after it is consumed as script text.

This is the failure the acceptance ladder was built to catch, and it is not
catching it: a Tier-4 run drives the page and reports on layout, not on whether
a `<script>` was terminated.

## Census — 9 damaged active components

Unbalanced `<script>` vs `</script>` in `content_components.html_template`,
`is_active = true`, template ≥ 100 chars:

```sql
SELECT cc.name, cc.component_level, length(cc.html_template)
FROM content_components cc
WHERE cc.is_active = true AND length(cc.html_template) >= 100
  AND (SELECT count(*) FROM regexp_matches(lower(cc.html_template), '<script',   'g'))
    > (SELECT count(*) FROM regexp_matches(lower(cc.html_template), '</script>', 'g'));
```

**8 `tool` + 1 `section`**, across **6 domains**:

| component | domain | template len | intact earlier version? |
|---|---|---|---|
| `tool-arena-interface-vonc-com` | vonc.com | 23,353 | **none** (0 versions) |
| `tool-archetype-clash-calculator-vonc-com` | vonc.com | 24,549 | **none** (0 versions) |
| `tool-grip-force-friction-calculator-robot-hands-com` | robot-hands.com | 24,278 | **yes — 23,526** |
| `tool-drop-rate-tuner-gamesdesign-co-uk` | gamesdesign.co.uk | 22,518 | **none** (0 versions) |
| `tool-process-automation-scorer-…` | leopardessconsulting.co.uk | 14,644 | **none** (0 versions) |
| `tool-llm-cost-calculator-…-com` | ai-agent-orchestration.com | 21,724 | **none** (0 versions) |
| `tool-llm-cost-calculator-…-finetuning-uk` | finetuning.uk / leopardess | 17,828 | 1 version, none intact |
| `tool-llm-cost-calculator-…-leopardessconsulting-co-uk` | (unplaced) | 21,724 | **none** (0 versions) |
| `archetype-taster-quiz` (`section`) | (unplaced) | 19,030 | — |

**The remedy `012` used is unavailable for 7 of 8.** It restored from
`component_versions`; these have no intact prior version to restore. They need
regeneration, which is a different and heavier operation.

Note the population is **not tool-only** — one `section` component is damaged
too, so any sweep must not be scoped to `component_level='tool'`.

## Why nothing surfaced it

- **`rendered_html` is damaged too**, so there is no template-vs-render
  disagreement to notice. Both sides carry the same cut.
- **Tier-4 acceptance passes them**: it checks layout and interactions against
  the PLAN's criteria, and a dead script fails no declared criterion.
- **`sectionTemplateValid` admitted them.** It keys on containing `</section>`,
  and 4 of these 8 contain one *upstream of the cut*. (This is the guard
  `bugs_open/024` replaced for tools with `toolTemplateValid`, live in
  v1.0.1140 — which now correctly REJECTS all 8. That is how the census was
  found.)

## Interaction with what is already live — read before repairing

`toolTemplateValid` (live, v1.0.1140) drops these from `loadComponentSchemas`.
Consequences, both intended:

- **Re-render path:** the component is "not found", so the section is CARRIED —
  the damaged stored HTML is preserved rather than re-rendered from a damaged
  template. No change to the live page, no further harm.
- **`plan_sections` path:** they flow to Path 3, i.e. a `needs_new_component`
  work item. **So up to 8 such items may appear as these pages are re-planned.**
  That is the correct remedy shape for a component with no intact version — but
  it will arrive unannounced unless someone is expecting it, and it is why this
  file exists as much as the breakage is.

## Fix candidates

1. **Sweep + regenerate.** For the 7 with no intact version, raise
   `needs_new_component` deliberately rather than waiting for an incidental
   re-plan, so the repair is tracked and ordered. Restore
   `tool-grip-force-friction-calculator` from its intact 23,526-char version
   first — that one is cheap and needs no LLM.
2. **Give the acceptance ladder a structural check.** A Tier-2/Tier-4 assertion
   that every `<script>`/`<style>`/`<section>` on a rendered page is terminated
   would have caught all 8 the day they landed, and costs nothing per run.
   `componentRegressionIssues`' `balancedPairs` already encodes the predicate —
   reuse it rather than writing a second one.
3. **Guard the render, not just the write.** `bugs_open/021` already asks for the
   write guard to cover `page_components.rendered_html` / `pages.rendered_*`.
   These 8 are the evidence for why: the damage reached the render surface and
   sat there.

## How to verify a fix

```sql
-- must return 0 rows when this is closed
SELECT cc.name, cc.component_level FROM content_components cc
WHERE cc.is_active = true AND length(cc.html_template) >= 100
  AND (SELECT count(*) FROM regexp_matches(lower(cc.html_template), '<script',   'g'))
    > (SELECT count(*) FROM regexp_matches(lower(cc.html_template), '</script>', 'g'));
```

Then re-fetch the two URLs above and confirm `<script` and `</script>` counts
match. **Check the live bytes, not the DB** — the whole point of this bug is
that both sides agreed and were both wrong.

## Related

- **`bugs_closed/012`** — the cause, fixed and closed. This is its unswept
  wreckage. 012's closure was correct on its own terms ("fixed AND live"); what
  it did not carry was a census, and the bar for closure does not currently ask
  for one.
- **`bugs_open/021`** — durable write guard covers one path only.
- **`bugs_open/024`** — how this was found: `toolTemplateValid` was written to
  stop healthy tools being dropped, and its calibration run against all 27 live
  tool templates produced this list as the other half of the output.
