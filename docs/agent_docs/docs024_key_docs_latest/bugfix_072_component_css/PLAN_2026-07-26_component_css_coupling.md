# PLAN — bugs_open/072, component markup ships without matching CSS

**Opened** 2026-07-26. Entry point for a cold start: `README_where_we_are.md`, then this.

**Note the name collision.** There are two `bugs_open/072` files. This workstream owns
`072_HANDOFF_2026-07-25_component_markup_ships_without_matching_css_on_some_sites.md`.
The other one — contact-info flat-vs-nested identity keys — belongs to
`brochure_component_library` and is nothing to do with this. Refer by slug, never by
number.

## The problem

Of the five sites emitting `class="news-card"`, two render the cards as bare unstyled
markup. Re-measured 2026-07-26, identical to the original filing 24 hours earlier.

## The cause

A site's `assets/css/styles.css` is a whole-file artefact written **only** by a
webdesign-agent design run. `RenderCSSFromSpecAction` appends the `css_snippets` matching
the site's component list *at that instant*; nothing re-renders it afterwards; page
rerender never touches it. The stylesheet is frozen at the last design run while the
component set keeps changing. Full evidence chain in the bug file's "Diagnosis" section —
not duplicated here, per "point at bugs, don't restate them".

## Decisions, and why

**D1 — fix it at page assembly, not by regenerating the stylesheet.**
Re-running webdesign-agent to refresh a site's CSS re-rolls the palette (the
`generic_theme` colour-churn problem). That is not a safe repair for two live customer
sites: it would change the design of the whole site to fix one section. Collecting the
snippets at assembly time also fixes the *class* rather than the instance — markup and
CSS are then produced by the same mechanism and cannot drift apart, which is exactly what
the bug file names as the defect class.

**D2 — do BOTH the assembly injection and the component-owned `<style>`, deduped.**
Owner's call. They are not redundant, because they cover different populations:

| | covers | delivered by |
|---|---|---|
| component-owned `<style>` in `html_template` | pages built or component-re-rendered *after* the change | the build path, which DOES re-render templates |
| assembly-time injection | every page built *before* the change | the cheap assemble-only rerender |

The dedup is what stops them being two mechanisms doing one job: the injection skips any
component whose stored `rendered_html` already carries a `<style>`, per component
function. As a page's sections get re-rendered over time, each component silently hands
over from the injected copy to its own — no flag day, no double-ship.

**D3 — copy the CSS from `css_snippets` at apply time; do not re-type it.**
Three of the five sites already serve those exact rules from `styles.css`. Re-typing them
into the migration would make any drift — a changed colour, a dropped rule — restyle three
live customer sites as a side effect of fixing two. Copying `css_content` in the `UPDATE`
makes the two byte-identical by construction.

**D4 — inject before `</head>`, i.e. after the stylesheet `<link>`.**
Where a frozen `styles.css` holds an older copy of the same snippet, the fresher copy must
win the cascade tie. Equal specificity means document order decides, so placement is the
whole mechanism, not a detail. The unit tests assert the ordering, not merely the presence.

**D5 — patch the bulk rerender path too.**
`rerender_pages_actions.go` assembles and deploys independently. Left alone, a bulk
rerender would strip the CSS back off a page the single-page path had just styled. The
same asymmetry is already recorded in that function for JS assets and has never been
fixed — this does not fix it for JS, but it does not repeat it for CSS.

**D6 — no live repair, no page re-render.**
Owner's call: let the fix reach the sites on the next image roll. So nothing about the two
failing sites has changed yet, and 072 stays OPEN. Recorded here because the temptation
to "just re-render the two pages and close it" is real and would have been a live change
to customer sites made as a side effect of a bugfix.

## Phasing

1. ~~Diagnose.~~ Done 2026-07-26.
2. ~~Go: `collectComponentCSS`/`injectComponentCSS` + both call sites; empty-list
   hardening.~~ Committed `7821ad7f5`. **Inert until an image roll.**
3. ~~Migration 222: the two components carry their own CSS.~~ Applied and recorded.
   **Invisible until sections are re-rendered.**
4. Council verdict `75d1a2af-afb8-492d-9587-4aa13bc440a2` — pending.
5. **After the roll:** pod-grep `data-component-css`, re-render `index` on
   ai-agent-orchestration.com and relojistas.com, measure all four sites, confirm the two
   controls are unchanged. Then and only then close 072.

## Deliberately not done

- **No detection check.** The bug file's fix candidate 3 asks for a check that a page's
  emitted classes have matching rules. It is not built, and the reason is worth recording:
  discovery checks are DB-only by house convention, and the bytes actually served at
  `/assets/css/styles.css` exist only in git — `css_themes.css_content` is empty for every
  composition-installed site. So a DB-only check can reconstruct the CSS a site *should*
  have and would report "fine" while the site serves a stale stylesheet. It would not have
  caught this bug. A check that would needs to read the deployed artefact, which is a
  different shape of tool (a report script, like `098_REPORT_*`). Left unbuilt rather than
  built wrong.
- **The `applies_to` granularity problem** (`FOCUS_visual_pipeline_css_and_component_lists.md`
  Cause 2): snippets keyed on generic terms (`card`, `feature`, `cta`) never match real
  component names (`features`, `differentiators`, `call-to-action`), so ~17 of the 21
  snippets have never shipped to any site. Real, known, documented, and **not this bug** —
  it makes sites plainer than intended; it does not make markup ship unstyled.
- **The JS asymmetry** in the bulk path. Named in a comment there since before this work.
