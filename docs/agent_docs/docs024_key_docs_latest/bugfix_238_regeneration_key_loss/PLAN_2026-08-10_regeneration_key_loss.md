# PLAN 2026-08-10 — bugs_open/238: a regeneration loses the keys nobody asked the LLM for

**Lane opened 2026-08-10 (evening), continuing 08-11.** Session name `bugfix 238`.
Owner brief: take the next unowned bug in `bugs_open/`, research prior art and the
live data, plan a robust framework-level fix with fable, check the council, keep
the docs updated, put missteps in `WRONG_CALLS.md`.

## The bug as filed, and how the filing was wrong about its own cause

`bugs_open/238` (filed 2026-08-09 by the finetuning imagery lane) records five
`<img class="csg-card-image" src="">` live on finetuning.uk's homepage after a
`tone_shift` work item regenerated the `case-studies-grid` section. Its stated
mechanism:

> "The generator is not told that certain keys are structural rather than
> editorial, so it reproduces the ones that look like copy (`*_image_alt`) and
> drops the ones that look like plumbing (`*_image_url`)."

**That is not what happened, and the difference decides the fix.** The LLM never
saw those keys and could not have emitted them. `plan_sections` splits a
component's schema fields by `source`: only `source:"llm"` fields are put in
`llm_field_specs`, and the writer prompt closes with *"Return a JSON object with
exactly the keys listed in 'What To Write'. Do not add any keys not in that
list."* Every lost key is **resolver-sourced**, not LLM-sourced.

The lost set is exactly the component's 11 non-`llm` fields — measured, not
inferred (`page_component_history` 58 keys → 47 at 2026-08-09 15:17:19Z):

| field | declared source | required |
|---|---|---|
| `card1..5_image_url` | `site_assets.image` | true |
| `card1..5_link_url` | `site_specs.case_studies.cardN_url` | true |
| `cta_link_url` | `site_specs.pages.contact_url` | true |

Zero LLM fields were lost. Every `*_image_alt` survived.

## The four-fault chain

1. **The sources resolve nothing on this site.** `site_assets.image` aliases to
   role `hero` (`imageryplan.ImageRoleForPath`), and `r.assets` is populated only
   from `site_plan_imagery` joins — finetuning has **0** imagery rows and no
   current plan. `site_specs` has no `case_studies` and no `pages` aspect — and
   **no site fleet-wide has ever had either**, so the six link fields have never
   been resolvable anywhere. (`plan_sections_action.go:538-559`, `:595-610`.)
2. **A required field's absence is silent by default.** `on_missing` defaults to
   `skip_field` when the schema omits it (`:1844-1846`), and the required branch
   honours `skip_field` by omitting the field (`:1893-1899`) on the stated premise
   *"templates gate on the field"* — **false for this template**: the `<img>` is at
   root scope with a bare `src="{{.card1_image_url}}"`. Info log, no work item.
3. **The render gate is exempt from exactly this class.** `missingRequiredLLMFields`
   refuses to render a section missing a required field — but only
   `source == "llm"` (`json_envelope.go:463-467`). A required `site_assets.*`
   field is skipped by construction.
4. **The persist is a wholesale replace.** `save_page_sections` DELETEs the page's
   agent-writable rows and INSERTs the fresh set (`:756`, `:904`) — it snapshots
   the old row into `page_component_history` first (`:713-727`), so it *holds the
   values in hand at the moment it destroys them*.

**Why it had never happened before.** The values survived from 2026-05-01 to
2026-08-03 byte-stable because every intervening run was a *rerender*, and
`rerender_page_sections` **merges** `stored ⊕ fresh resolved_data`
(`rerender_page_sections_action.go:498-506`, comment: *"so the row remains a
complete render source"*). The first true regeneration destroyed them.

> **The asymmetry between those two functions is the bug.** One path treats
> `content_data` as a complete render source to be preserved; the other treats it
> as output to be replaced. Both are shipped, both are deliberate, and nothing
> reconciles them.

## What the filing missed, found this session

- **The images are the visible third of the damage.** All five
  `card*_link_url` and `cta_link_url` went too. Those are `{{if}}`-gated, so the
  five "Read case study" links and the section CTA button **vanished silently** —
  no empty attribute, no gap, nothing for a markup-shape check to see. Verified
  at the served page: 0 `<a class="csg-card-link">` and 0 `<a class="csg-cta-btn">`
  anchors (every `csg-card-link`/`csg-cta-btn` hit in the HTML is a CSS rule).
  **A guarded field fails more quietly than an unguarded one** — the opposite of
  the intuition that gating is the safe pattern.
- **Fleet blast radius** (the bug's own open question, now answered): 10 active
  components carry a `{{.*_image_url}}`; **5 deployed rows across 4 sites** are
  missing at least one — finetuning.uk `/index.html`, ai-agent-orchestration.com
  `/index.html`, leopardessconsulting.co.uk `/blog.html` (post1-6) and its
  automation-savings-estimator tool page, oufe.com's recovery-waterfall tool page.
- **26 fields across 8 components are `type:"url"` with `source:"site_assets.*"`** —
  invisible to all three existing image checks, which key on
  `type in ('image','image_url')`.

## Design decisions

### A — the door-closer: carry structural keys at PLAN time, not save time

When a non-llm field's declared source resolves nothing (after every alias), fall
back to the value the page's own deployed `content_data` already carries. It flows
through the existing `merge_with` overlay (PBP-014) into the rendered HTML **and**
the persisted row — no new seam downstream.

**Why plan-time and not a merge in `save_page_sections`:**

- A save-time merge repairs the row *after* the HTML rendered without the value:
  the build still ships `src=""`, and the row then claims a
  `rendered_html_digest` that no longer reproduces from its own `content_data` —
  breaking the semantics `bugs_open/229` just established.
- `planSection` is shared by the build path and the rerender path, so one edit
  covers both. This is the direct answer to the `bugs_open/178` / `021` / `093`
  "one call site guarded, sibling unchecked" family — not a third path that
  merges, but a plan that is complete before anyone renders.
- `save_page_sections` already carries five refusing guards and `bugs_open/178`
  left a standing instruction that a sixth is the trigger for a unified
  content-loss detector as its own submission. This lane does not add one.

**Scope rule.** Schema-declared, non-`llm` sources only (`renderer`/`static`
resolve at render time by design), required *and* optional, non-empty stored value
only. **LLM fields are never carried** — a `tone_shift` rewriting the copy is the
regeneration working correctly, and carrying old copy would defeat it. Live
resolution always wins, so a repaired source takes precedence on the next build
and a carried value is bounded to "while the source resolves nothing at all".

**A required field that resolves nowhere** — not from its source, not from a
stored row — is recorded durably as `STRUCTURAL_KEY_CARRY_MISS` in
`agent_error_log` via the existing `LogActionFindings`. It does **not** defer the
section: that would be RFC_009's option A (owner: NOT taken) arriving through the
plan-time door.

### B — the guard: stop discarding a report we already compute

`missingBareFields` (`component_library.go:860`) already parses the template, walks
root-scope actions only, and returns the fields sitting inside `href=`/`src=` that
rendered empty. At the exact render that shipped this bug it returned
`[card1_image_url … card5_image_url]` — and `RenderTemplate` (`:953`) throws the
result away (`out, _, _ :=`). The site-chrome renderer is the only consumer.

`RenderComponentAction` will consume it and **refuse** rather than ship a dead
control, shipped as an **opt-in config field defaulting OFF** (the owner's own
RFC_022 pattern), flipped ON for `page-content-writer` — measured as the only live
caller of `render_component`. Refusal leaves the stored row and the live page
intact, and mints a work item keyed by **page + slot** (the existing
`image_url_404:empty-src` key is site-wide, and finetuning's `blocked` row has held
that fleet slot since 08-03, so new damage cannot even mint an item).

The rerender path gets the same emit **record-only**: it merges, so it cannot lose
a key — and it is the repair vehicle, so it must not refuse.

**This is not RFC_009 option A.** A was a render gate driven by `on_missing`
*declarations* — inert for ~90% of fields, and able to break live pages. This is
template-authority (the ungated `src=` placeholder is the contract) and scans the
final data map, which is where every supply path has already converged — RFC_009's
own hard question 2.

### C — the site repair, through the framework

Restore the 11 keys as data, then re-render with **no LLM** (`reason:
section_data_resolved`). The cards were rewritten, so image assignment is by
**subject**, corroborated independently by the regenerated alt texts; each of the
five assets is used exactly once. The old card links 404 today, so they are
re-pointed at `/case-studies.html` rather than restored. Seeding the two missing
`site_specs` aspects makes six of the eleven fields resolve properly from now on.

Owner decisions taken 2026-08-10: guard = refuse-and-raise opt-in; mapping =
subject-based; links → `/case-studies.html`.

## Out of scope, deliberately, and why

- **The other four damaged rows.** ai-agent-orchestration's historical URLs 404 and
  the site has no case-study assets — restoring would trade an empty `src` for a
  404 one. leopardess/oufe never had the keys (fresh-miss class, not regressions)
  and have no candidate imagery. All are named in the bug file; B's widened checks
  flag them when discovery resumes.
- **`bugs_open/236`** (hero/logo lose `image_url` one layer down) — same family,
  different layer, root cause not established. Named as a relation; nothing claimed.
- **`sectionHasImageField` widening** — measured benefit zero (none of the 8
  url-source components' templates reference `hero_url`/`background_image`) and it
  writes live-path data. Deferred with the measurement recorded.
- **Re-sourcing the shared component schema** so five cards can name five distinct
  assets. `content_components` has no `site_id`; `3f946437` is used by four pages
  across three sites. Per-card image sources are **not expressible today** — the
  resolver never looks assets up by literal key. Named as a future resolver-arm
  candidate for the owner, not built here.

## Sources

`bugs_open/238` · design reports this session (mechanism; guard+repair) ·
RFC_009 (DECIDED 2026-08-03, "C now, B next", A not taken) · RFC_016/RFC_022 ·
`bugs_open/178` (the stop sign, and the `edit_live` channel) · `bugs_open/151`
PBP-037 fact-carry (the pattern this extends) · `bugs_open/229` (digest
semantics) · `bugs_open/230` (discovery driver, currently disabled) ·
`WRONG_CALLS.md` 2026-08-10 (the refuted `case-studies.html` "cheapest win").
