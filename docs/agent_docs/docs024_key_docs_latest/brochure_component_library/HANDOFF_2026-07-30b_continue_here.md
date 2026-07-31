# HANDOFF — brochure component library / fundamentallyai.com — 2026-07-30b

**Cold-start document. Supersedes `HANDOFF_2026-07-30_continue_here.md`** for state, but
that file is still worth reading once: its "READ THIS BEFORE DOING ANYTHING ON STEP 5"
section is unchanged and still binding, and its landmine list still holds.

## What changed since 07-30 (morning)

The morning handoff's **"NOT DONE — the actual next build"** is now **DONE and live**.

## The tool, in one paragraph

`tool-review-council-simulator`, live at `/tools/review-council-simulator.html`. Three
sliders (blocking threshold, reviewer relevance, revision rounds) plus a 26-seat roster
with four presets. It estimates how often a sound change passes an AI reviewer panel,
calibrated on **362 real `council_report` runs, 2026-07-10 to 2026-07-30**, with each
seat's measured objection rate at three severity thresholds. Page is `page_type='tool'`,
three sections: `hero-tool`, the widget, `tool-cta`.

**Its own travelling docs are the primary source, not this file:**
- PLAN: `doc_plans` where `subject_type='tool'` and
  `subject_key='tool-review-council-simulator'` — aim, behaviour contract, data
  provenance, and a numbered **"Deliberate decisions — do NOT fix these"** list.
- NOTES: `doc_notes`, same subject key, categories `["build","verification","tool"]`.
- Repo copy of the source: `components/tool-review-council-simulator/`
  (`template.html`, `install.py`, `travelling_docs.py`).
- Lane log: `NOTES_brochure_component_library.md`, entry `## 2026-07-30 (evening)`.
- Owner prose: `README_where_we_are.md`, entry dated 2026-07-30 evening.

```sql
SELECT body FROM doc_plans WHERE subject_type='tool'
 AND subject_key='tool-review-council-simulator' AND is_current;
```

## ANSWERED: the open question the morning handoff left as the next thread's first action

It asked whether `doc_plans`/`doc_notes` accept an arbitrary subject key cheaply enough
to use before step 5's wiring exists. **For a TOOL, yes, today, with no migration** —
`doc_plans_subject_type_check` already allows `'tool'` and 36 tool plans exist. This
tool's docs are in there now.

**For a section COMPONENT, no, not yet.** Commit `c659e312b` (another lane, same day)
added `subject_type='component'` in Go *and* in **migration 273, which is UNAPPLIED**.
Until it is applied and the image carries the Go half, a section component can carry
neither doc. That is the distinction, and it is why this build used the tool route.

## Verification — what is proven, and by what

- **S6 (does it operate when driven?):** `scripts/probe_council_simulator.py`. 44 checks,
  exit 0 clean / 1 on failure. Run it against the served page after ANY re-render:
  `python3 scripts/probe_council_simulator.py --url https://fundamentallyai.com/tools/review-council-simulator.html`
  Local template: no args. A specific file: `--template <path>`.
- **The probe is mutation-proven** (6 mutants, all correctly fail). Do not weaken it
  without re-running them; a check that has never been shown to fail is not evidence.
  The mutant generators are in the lane NOTES entry for this date.
- Live at the artefact: HTTP 200, three sections rendered, the widget's script inline
  **after** its markup (not extracted), real stats in the hero, screenshot inspected.

## Landmines this build paid for (both now in `LANDMINES.md` and `doc_notes`)

1. **`hero-tool` / `tool-cta` render NO buttons unless you set the `*_url` fields.** The
   `*_label` fields alone are dead data, and the two components spell the key in
   **opposite orders** (`cta_primary_url` vs `primary_cta_url`). The live
   llm-cost-calculator page has zero CTA anchors because of this. **Not fixed** — see
   "Loose ends".
2. **A browser probe injected before `</body>` runs BEFORE the component's
   `DOMContentLoaded` init** and reports exactly the bug it exists to catch. This
   probe's first run failed 7 checks against a correct component. Defer the driver to
   `load`, then mutation-test it.

Plus two smaller facts, in the NOTES entry: **a `grep -c` for a CSS class on a page that
inlines its own stylesheet always returns at least 1** (the class definition), which is
how I briefly talked myself out of landmine 1; and **`spec.filename` on a
`page_rerender` item does not set the served path** — `pages.url` does.

## Two corrections to figures other docs still carry

- **CLAUDE.md's council-gate "approval ran ~80%" is SOUND — I wrongly "corrected" it, so
  ignore any claim of mine that it is a two-day peak.** It is the **per-SUBMISSION**
  figure. Both denominators are real and they differ by 26 points:
  - **per ROUND: 50.7%** post-fix (211 rounds) — *"will this round approve?"*
  - **per SUBMISSION: 77.2%** (105 of 136 correlations) — *"will my plan get through?"*

  Already measured and recorded by another thread on 2026-07-28; see the memory topic
  `council-review-practice-index.md` line 24 and `council-gate-workstream.md`. My
  independent numbers reproduce theirs. **A REVISE or two is the median path, not a
  failure signal.** The commit message on `32653bd85` carries my wrong explanation and
  cannot be amended (forward-only); this is the correction of record, and the tool itself
  now prints both denominators.
- **There is no rounds-to-approval distribution in `doc_notes`.** All 266 council-gate
  verdict notes say `(round 1)`. I had planned to model it and did not, because it is
  not there. Do not build on that field.

## Loose ends, smallest first

1. **The sibling page's dead CTA labels** (landmine 1). `llm-cost-calculator.html` stores
   `"cta_primary_label": "Run the calculator"` with no URL, so its hero has no buttons.
   One `content_data` edit plus a re-render. Left undone because *what those buttons
   should say and point at* is a content decision, not a mechanical fix. The same is
   likely true of its `tool-cta` block and of `model-approach-selector`; **check all
   three before editing one.**
2. **No `tool-guide-intro` section on the new page.** The sibling has one (an
   8KB, `render_mode=agent` explainer). Deliberately omitted so the page needed no LLM
   pass. Adding it is a content task: it needs a real dispatch, not hand-written JSON.
3. **A guide page.** Both other tools have one (`/guides/...-guide.html`) and
   `tool-cta`'s copy on this site refers to companion guides that "set out the method".
   This tool has no guide, and it is the one that most needs one, because its model has
   stated assumptions worth a page of their own.
4. **`/tools.html` does not exist** on this site, so nothing links to a tools index. Both
   sibling pages carry an "Explore All Tools" label; this page does not, deliberately.
   Either build the index or drop those labels.
5. **`gated_by_truncation` is `false` on all 362 council reports.** Noted, not chased —
   `bugs_open/138`'s lane owns that field.

## Still NOT this lane's work (unchanged from the morning handoff)

The **staged step-by-step build system with stage gates** (`features_open/027`,
`staged_component_build/PROPOSAL_2026-07-30_...`). Owner's instruction is that it happens
in a separate thread. This build is a *user* of that ladder's ideas (it hand-rolled S2
and S6) and is worth citing as evidence, but do not build the system here.

## Next summary

**No SUMMARY was written for this build**, on purpose:
`SUMMARY_2026-07-30_the_panel_is_finished_and_two_new_fronts_open.md` was written the
same morning and already frames this as the next front, so a second file hours later is
the near-identical shelf the cadence rule warns against. **The next summary is owed and
should cover this tool plus whatever step 5 becomes** — it will be a genuine inflection
by then.

---

## INCOMING 2026-07-31 — a resolver your components depend on changed what it guarantees

Not this lane's work to do, but you must know it before authoring another
`input_schema`. From the "bugfix 9" thread fixing `bugs_open/072`:
**`CONTRIB_2026-07-31_identity_source_resolution_changed.md`** in this directory.

One line each:

- `plan_sections`' `sourceResolver` now falls back, **after** a literal
  `site_specs.identity.<leaf>` miss, to the writer's nested shape
  (`identity.contact.<leaf>`) and then to the canonical `sites` row columns
  (`email`, `phone`, `contact_address`, `company_name`, `tagline`, `logo_text`,
  `logo_url`). Literal always wins — **no path that resolves today changes value**.
  Registered PBP-026, committed `ef9e7e999`, inert until the next chassis roll.
- **Therefore:** a flat `site_specs.identity.*` path is no longer a reliable way to
  make `on_missing: needs_human_review` fire. `sites.email` is populated on 12 of
  15 real sites. Check any schema that relies on a *miss*.
- **And:** the flat/nested hand-patch workaround on six sites' `identity` aspect is
  now unnecessary. Please stop propagating it.
- **Three census findings that ARE yours** (measured, not acted on): 74 of 100
  declared `site_specs.*` paths name an aspect existing on no site (already
  diagnosed as "decorative" in `bugs_closed/018` — chrome runs a thinner path with
  no fallback machinery at all); the vocabulary carries near-duplicate aspect
  names (`nav`/`navigation`, `cta`/`ctas`); and `site_specs.pricing.tiers[0].name`
  style paths can never resolve because `navigateMap` has no array-index syntax.
- **One thing to veto if you disagree:** `identity.address` → `sites.contact_address`
  is the only mapping going beyond what `loadSiteDataFull` reads. Droppable.

---

## Contribution 2026-07-31 from the `bugfix_128_image_url_404` lane — 128 is CLOSED, and two things changed under you

You filed `128` and your `HANDOFF_2026-07-28b` correctly listed it as *"read, still
unowned"*. It is now **fixed, live on v1.0.1219 and closed** (`bugs_closed/128`,
commits `beff42809` / `6d3992213`). Two consequences you would otherwise find by
surprise:

1. **`check_placeholder_image_in_use` is now the SOLE owner of "the fallback path is
   rendered and no asset of that purpose exists, so build one".** `image_url_404` used to
   carry an exact duplicate of that branch — same two paths, purposes, `needs_hero_image`
   / `needs_logo` item types, `image-build-handler`, precondition, and both enabled on
   `design-discovery-agent`, differing only in `item_key` so they could not even dedup
   against each other. Neither had ever fired. The duplicate is deleted; **your check is
   unchanged and now unambiguous.** If you were relying on `image_url_404` to route image
   regeneration, it no longer does — every one of its emissions is flag-only.

2. **`image_url_404` now reports far more, and far more accurately.** Measured live over
   all 127 rendered image paths on 13 sites: the old predicate reported **21 working
   images and missed 6 live 404s**; the new one reports 1 and misses 0. It also scans
   `site_components` (your defect 3 — the chrome, on every page) and flags `<img src="">`.
   Expect roughly **18 page findings + 2 chrome findings** fleet-wide once every site has
   swept, against 13 items in the check's entire lifetime before. All flag-only, so they
   consume no dispatch budget.

Your imagery work is the main consumer here: `bugs_open/114` (imagery generated and never
referenced) is the *opposite* direction of the same pipeline, and the six
`/assets/images/hero.jpg` 404s this check now surfaces are the legacy
`sites.content_data.hero_url` default your own 07-29 contribution traced. **The check can
finally see them**, which makes 114's repair verifiable rather than assumed.

Nine stale `detected` rows carry the old extension-less item key; four of them
(`fundamentallyai:brand-illustration`, three `robot-hands:content-hero-tool-*`) were the
**old predicate's false positives** and can be cancelled outright. Left alone here because
`bugs_open/083`'s lane is actively working that population.

## INCOMING 2026-07-31 — 151 candidate 3 is built, council-APPROVED, and INERT

From the `gauntlet_dead_cta` lane. Not this lane's work to do, but the switch is yours
to throw, so you need to know it exists — a CONTRIB has been sitting unreferenced in
this directory since 09:13 (`CONTRIB_2026-07-31_151_candidate_3_is_built.md`, updated
since with the approval); this section exists because a cold start would never have
found it.

- **What is built:** `check_content_duplication` (detect) +
  `remove_duplicate_page_sections` (repair) + the `deduplicate-sections` handler agent —
  the deterministic half of your `bugs_open/151`. In-remit is same page + **same slot +
  byte-identical `content_data`** only; everything needing judgement (near-duplicates,
  cross-page, shared-fact overlap — your candidate 1's territory) files ONE
  `capability_gap` with `do_not_auto_rewrite: true` and no handler.
- **Council-APPROVED** round 3, corr `da3f2d9b-ae6f-492d-ad3b-748323b66367` (12 approve /
  2 advisory). Rounds 1–2 caught real defects, including a false positive that would have
  deleted a live section from vonc's home page — the identity rule was narrowed in
  response (`43492ec94`). Read the CONTRIB before relying on the earlier prose.
- **INERT, and enabling is YOUR decision:** zero discovery agents name the check
  (verified across all five workflow-bearing columns of `agent_definitions`).
  Enabling = adding `content_duplication` to a discovery agent's check list. **The first
  site it runs against starts deleting rows.** Today it would delete nothing anywhere
  (measured with the shipped functions over all 1,023 live rows), but that figure goes
  stale — re-run `gauntlet_dead_cta/scripts/dedup_census_shipped.go` before deciding.
- **A pre-delete guard consulting the plan stores is being BUILT now** (owner decision
  2026-07-31, answering the council's `bug_historian` seat): the repair will refuse to
  delete where the effective plan source (`site_plan_sections` → `site_specs.site_plan`
  aspect → `pages.sections`) itself specifies the repetition. Do not gate your enable
  decision on the older "guard is deferred" wording in the CONTRIB's first version.
- **Candidate 1 (assign facts to sections at plan time) is still yours**, and nothing
  here builds or forecloses it.
