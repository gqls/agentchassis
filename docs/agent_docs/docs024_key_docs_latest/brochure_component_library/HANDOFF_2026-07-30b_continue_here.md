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
