#!/usr/bin/env python3
"""Write this tool's travelling docs into doc_plans / doc_notes.

Per convention 037 each tool carries two docs keyed by its component `function`:
a PLAN (intent, changes rarely) and NOTES (append-only history). The owner asked
for this build to be documented "as per the doc traveller", so it goes in the DB
tables the platform actually reads, not only in files.

VERIFIED BEFORE USING THIS PATH (2026-07-30): doc_plans' CHECK already allows
subject_type='tool' (36 tool plans exist), so no migration is needed here. This
is NOT true for subject_type='component' -- commit c659e312b added that in Go and
in migration 273, and 273 is UNAPPLIED. A section component still cannot carry
these docs today; a tool can. That distinction is the whole answer to the open
question the 2026-07-30 handoff left for this build.

Usage:
  python3 travelling_docs.py --emit-plan   | kubectl ... psql ...
  python3 travelling_docs.py --emit-note   | kubectl ... psql ...
"""
import sys

SUBJECT_TYPE = "tool"
SUBJECT_KEY = "tool-review-council-simulator"
SOURCE = "brochure_component_library"
CREATED_BY = "operator:brochure_component_library"

PLAN = """# PLAN -- tool-review-council-simulator

## Aim

Let a visitor answer a question our own team had to answer with money: if you put a
panel of AI reviewers in front of every change, what proportion of sound work comes
back, and what does tightening or loosening the panel actually do to that number?

It exists to make one of this company's two flagship capabilities (the multi-agent
review council) legible by letting someone operate it, rather than read a page about
it. It is the third tool on fundamentallyai.com, after the LLM cost calculator and
the model-approach selector.

## Source spec

Owner instruction, 2026-07-30, verbatim in the lane handoff: "An interactive tools
page, with sliders/inputs, built on real live platform numbers (pages/sites hosted;
council decisions by verdict and by reviewer seat)", with the requirement that the
numbers be real and that the style match the two existing tool pages.

The "pages/sites hosted" half was measured and deliberately NOT used as the tool's
engine (see Deliberate decisions). The council half is the engine.

## Behaviour contract

Inputs, all live-updating on the `input` event (no calculate button; a slider that
needs a button to take effect is a broken slider):

- `blocking threshold` (range 0-2): any objection / medium and up / high only.
- `reviewer relevance` (range 10-100, step 5): the share of seats that have an
  opinion on a given change. Default 70, which is what our own runs show.
- `revision rounds` (range 1-6): how many attempts before giving up.
- a 26-row seat roster of checkboxes, each showing that seat's measured objection
  rate at the current threshold; four presets (typical 8 / minimal pair / all 26 /
  clear).

Outputs: first-round pass probability; pass-within-N-rounds; mean rounds to pass;
reviews spent; expected seats firing; a ranked blocker bar chart (top 8); and a
"reality band" placing the user's configuration against three measured figures.

Model: for each enabled seat, P(blocks) = relevance x that seat's measured rate at
the chosen threshold. P(pass) = product of (1 - P(blocks)). P(within N) =
1 - (1 - P(pass))^N.

Empty roster renders "n/a" and an explanation, never a misleading 100%.

## Data provenance

Every rate is measured, none invented. Source: `diagnosis_artifacts` where
`kind='council_report'`, 362 rows spanning 2026-07-10 to 2026-07-30, aggregated by
`body::jsonb -> 'reviews'` (each element carries `reviewer`, `verdict`,
`objections[].severity`). Queries are in the lane RUNBOOK.

The three comparison figures:
- 2.6% (4 of 154) first-round approval before 2026-07-22
- 51.0% (106 of 208) after 2026-07-22
- 30.4% (110 of 362) across all runs

The 07-22 step is the decision-rule fix in bugs_closed/057, and the data shows it
plainly, which is why the tool leads on that lever.

## Delivery mechanism

Path 1 equivalent, self-contained: the component's `html_template` holds its own
`<style>`, its markup, and its `<script>` as an IIFE **after** the markup. It uses
NO `js_snippets` entry and NO `/assets/js/snippets.js`. This matches the sibling
tool (`tool-llm-cost-calculator`, `input_schema` NULL, `template_variable_count` 0)
and it is the reason this component cannot suffer the head-loaded init failure that
cost the teaser-reveal-panel five rounds: a script placed after its own markup
cannot run before that markup exists.

A `document.readyState` guard is present anyway, so the component still works if
the template is ever bundled into a head-loaded file.

## Dependencies

- `content_components` row `tool-review-council-simulator-fundamentallyai-com`
  (`function=tool-review-council-simulator`, render_mode=template,
  category=interactive, component_level=tool).
- Page `/tools/review-council-simulator.html`, page_type=tool, three sections:
  `hero-tool`, this widget, `tool-cta`. hero-tool and tool-cta reuse the exact
  component rows the llm-cost-calculator page uses.
- The site palette CSS variables (`--color-primary`, `--color-surface`,
  `--color-border`, `--color-text`, `--color-text-muted`, `--color-accent`).
- NOT in `site_plan_sections`. Deliberate: no tool page on this site is, and adding
  one invites the plan-driven rebuild to regenerate its copy with an LLM.

## Deliberate decisions -- do NOT "fix" these

1. **The numbers are a dated snapshot, and the page says so.** This is static HTML
   with no API to call at load. Baking the figures in at build time is the only
   honest option, so the tool states the measurement date and the run count in the
   body rather than implying a live feed. Re-measure before quoting them in a month.

2. **Site/page counts were measured and left out of the engine.** 442 pages, 419
   active, 383 deployed, 14 sites, 110 tool pages. They make a fine stat band but
   they are not a *tool*: nothing a visitor could slide would change them. Using
   them as decoration on an interactive page would have been the passive dashboard
   the owner explicitly did not ask for. They are recorded here so the next thread
   knows they were considered rather than missed.

3. **Seats with under 20 recorded runs are flagged "thin" in the UI.** Four seats
   sit at 100% objection over 4 to 21 runs. Presenting that as a rate would be the
   most misleading thing on the page, so the roster marks them and a note names
   which of the user's selected seats are thin.

4. **The model's two false assumptions are printed under the results, not hidden.**
   Seats are treated as independent (they are not; a badly shaped change draws
   objections from several at once) and each round as a fresh roll (real
   resubmissions answer the objection they were given, so this understates a real
   council). The page says to read the pass rates as a floor.

5. **No `tool-guide-intro` section.** That component is render_mode=agent and would
   need an LLM pass to fill. The page is deliberately self-contained; adding the
   intro later is a content task, not a fix.

6. **No `/tools.html` link anywhere.** That index page does not exist on this site.
   The sibling page's "Explore All Tools" label has no valid target; it was not
   copied. Every URL in this page's content_data was checked against `pages` first.

## Verification contract

`scripts/probe_council_simulator.py` is the gate, and it is mutation-proven: it
fails on a no-init mutant, on a script-before-markup mutant (the exact
teaser-reveal-panel bug class), and on a dead-slider mutant. Exit 0 clean, 1 on any
failure. Run it with `--url` after any re-render. A check that has never been shown
to fail is not evidence.
"""

NOTE = """2026-07-30 -- built, S6-gated, and installed. Category: none of the usual (no bug
to record); this is the build log.

Investigated the two existing tool pages before designing anything, per the owner's
instruction. Found the pattern: page_type='tool', a 4-part stack (hero-tool,
tool-guide-intro, tool-<slug> widget, tool-cta), and the widget is ONE
self-contained html_template of ~35KB with input_schema NULL, js_content NULL and
template_variable_count 0. Its inline `<script>` sits AFTER its markup and calls
init() synchronously. Forked per site: four rows share function
'tool-llm-cost-calculator', one per site. Copied that shape exactly.

Measured the numbers rather than reusing any figure already written down, since the
lane's own doctrine is that counts go stale within days. Two things fell out that
changed the design:

- **All 266 council-gate verdict notes say "(round 1)".** I had intended to model
  rounds-to-approval on the real distribution. There is no such distribution in that
  source. Nothing was built on it. Had I not looked, the tool would have shipped a
  fabricated curve, which is precisely the failure this file exists to prevent.

- **CLAUDE.md's "approval ran ~80%" is a two-day peak, not the sustained rate.**
  Measured by day, the post-fix rate is 51.0% (106 of 208 runs, 07-22 onward)
  against 2.6% (4 of 154) before. The daily figures do reach 62-73% on four days,
  which is where ~80% came from. The tool quotes 51%, the sustained figure, and the
  pre/post pair is now the most interesting thing on the page.

Also checked the denominator rather than assuming it: 284 doc_notes carry the
'council-gate' category but 18 of them are threads' own notes, not verdicts. The
honest verdict denominator is 266. Separately, `diagnosis_artifacts` holds 362
council_report rows going back to 07-10 (the gate's notes only start 07-17) because
the reports include the fix-loop's own council runs. Two real denominators for two
different questions; the tool uses the 362 because per-seat data only exists there.

**S6 (real clicks in real Chromium) earned its place on the first run, by failing.**
The probe reported init had never run: headline still '--', roster empty, 26 seats
missing. The component was fine. The PROBE was wrong: it is injected inline before
`</body>`, so it executed during parsing, BEFORE the component's own
DOMContentLoaded init. It was measuring the pre-init page. Fixed by deferring the
driver to `load`. Recording it because the failure is indistinguishable at a glance
from the real bug it is designed to catch, and a thread that "fixed" the component
in response would have broken a working one.

Which raised the obvious question: can this probe fail for the RIGHT reason? Built
three mutants and confirmed it can. (1) init never called: 7 checks fail. (2) script
moved BEFORE its markup with the readiness guard removed -- the exact
teaser-reveal-panel failure -- 7 checks fail. (3) slider listeners removed: 7 checks
fail, and specifically the "changes the result" ones, while the seat-checkbox checks
still pass because that listener is separate. Exit codes: 0 clean, 1 on mutants.
35 checks pass on the real template.

Installed as one transaction across content_components, pages and page_components
with a DO block raising on any wrong count before COMMIT. Verified the stored
template is byte-identical to the file (md5 f04740f1, 28,725 bytes) rather than
merely "long enough" -- a length check would have passed on a truncated store.

One thing worth knowing about this site's tool pages: they are NOT in
`site_plan_sections`. Only the 7 plan-managed pages are. So the "three places a
placement lives" landmine reduces to two here (pages.sections and
page_components.slot_name), and adding plan rows would be actively wrong -- it
invites the plan-driven rebuild to rewrite the copy.

`spec.filename` on a page_rerender work item does NOT determine the served path.
History for the sibling pages carries three mutually inconsistent filenames for the
same page ('llm-cost-calculator.html', 'index.html',
'tools/model-approach-selector/index.html'), all runs complete, and the root-level
paths those filenames imply return 404 while the real /tools/ paths return 200. The
path comes from `pages.url`. [INFERRED from that evidence; the deploy action's code
was not read.] Used the explicit full path anyway.

Two latent defects on the SIBLING page, noticed while copying its content_data,
neither fixed here because neither is this build's scope: `hero-tool` renders its
CTA row only `{{if or .cta_primary_url .cta_secondary_url}}` and `tool-cta` only
`{{if .primary_cta_url}}`, but the llm-cost-calculator's stored content_data sets
only the *_label fields and no URLs. So both CTA rows on that page render as
nothing. This page supplies both labels and URLs.
"""


def q(s):
    return "'" + s.replace("'", "''") + "'"


def dollar(s, tag="BODY"):
    d = "$%s$" % tag
    if d in s:
        raise SystemExit("tag %s appears in body" % d)
    return d + s + d


def emit_plan():
    print("\\set ON_ERROR_STOP on")
    print("BEGIN;")
    # Supersede any prior current plan for this key, so is_current stays single.
    print("UPDATE doc_plans SET is_current=false, superseded_at=NOW()"
          " WHERE subject_type=%s AND subject_key=%s AND is_current;"
          % (q(SUBJECT_TYPE), q(SUBJECT_KEY)))
    print("INSERT INTO doc_plans (subject_type, subject_key, body, source, created_by,"
          " is_current, created_at, updated_at)")
    print("VALUES (%s, %s, %s, %s, %s, true, NOW(), NOW());"
          % (q(SUBJECT_TYPE), q(SUBJECT_KEY), dollar(PLAN), q(SOURCE), q(CREATED_BY)))
    print("""DO $CHK$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM doc_plans
   WHERE subject_type=%s AND subject_key=%s AND is_current;
  IF n <> 1 THEN RAISE EXCEPTION 'expected exactly 1 current plan, got %%', n; END IF;
END $CHK$;""" % (q(SUBJECT_TYPE), q(SUBJECT_KEY)))
    print("COMMIT;")
    print("SELECT subject_type, subject_key, is_current, length(body) AS body_len"
          " FROM doc_plans WHERE subject_key=%s;" % q(SUBJECT_KEY))


def emit_note():
    print("\\set ON_ERROR_STOP on")
    print("INSERT INTO doc_notes (subject_type, subject_key, site_id, body, categories,"
          " source, created_by, created_at)")
    print("VALUES (%s, %s, '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd', %s,"
          " '[\"build\", \"verification\", \"tool\"]'::jsonb, %s, %s, NOW());"
          % (q(SUBJECT_TYPE), q(SUBJECT_KEY), dollar(NOTE), q(SOURCE), q(CREATED_BY)))
    print("SELECT subject_key, categories, length(body) AS body_len, created_at"
          " FROM doc_notes WHERE subject_key=%s ORDER BY created_at DESC;" % q(SUBJECT_KEY))


if __name__ == "__main__":
    if "--emit-plan" in sys.argv:
        emit_plan()
    elif "--emit-note" in sys.argv:
        emit_note()
    else:
        raise SystemExit(__doc__)
