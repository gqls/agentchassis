# BUG 046 — the truncation fix stopped new damage; nothing ever repaired the old. 8 tools serve unterminated JavaScript on live customer sites

**Filed:** 2026-07-20 · travelling-docs thread
**Severity:** high — live customer-facing breakage on **6 domains**, present now,
and invisible to every check we have.
**Status:** OPEN. Detection (candidate 2) is **LIVE, ENABLED, and PROVEN
end-to-end** (v1.0.1149 + seed 186, 2026-07-22); grip-force restored at source
(census 9 → 8); 8 remain (no intact version → regeneration). The live *pages* are
still broken pending re-render delivery (bugs_open/024). See the UPDATEs below.

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
