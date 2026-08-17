# PLAN — ai-agent-orchestration.com: images, carousels, contrast

**Opened 2026-08-17.** Continues `HANDOFF_2026-08-05_rebuild_scope.md` (whose figures are
re-measured in `NOTES_site_improvement.md`, not carried forward). Evidence for every claim below
is in NOTES; the commands are in RUNBOOK.

Site `2a8ebf9c-20a2-4c39-b191-840b012371da`, `deployed`, **UNLOCKED**, heavy scheduled automation.

---

## 0. The correction that shapes this plan

The originating brief assumed contrast and images were "already fixed on other sites, so run the
improvement loop". **Measured 2026-08-17, that is not the case, and the plan is built on the
measurement instead.** Recorded here rather than silently designed around:

- The contrast repair (bugfix 122's legible ink) **is already live and correct on this site** —
  `--color-primary-ink` computes to `#768eb2` — and the site still has **44 firm contrast
  failures on 4 pages**, 14 of them at 1.00:1.
- The improvement loop cannot reach two of the three defect families: one lives in
  component-embedded CSS that survives a fresh render (`index` rendered 08-17, still 17
  failures), the other on a page that **cannot be re-rendered at all** (`pricing`, 5/5 NULL
  `content_data`).
- The live image handlers do not generate images. Their only precedent site now has zero images.

## 1. Contrast — two defect families, and a scope decision

**Family A — bare token on its own ground.** Component CSS says
`color: var(--color-primary, #1a1a2e)`; this site's `primary` IS its surface (`#0D1117`).
20 of 44 failures.

Two routes, and they differ in blast radius, which is why this is an owner decision:

| | route A1 — fix the site's palette | route A2 — fix the components to use the ink token |
|---|---|---|
| change | one `site_specs` row: give `primary` a visible value | component CSS → `var(--color-primary-ink, var(--color-primary, …))` |
| reach | this site (and by argument oufe.com) | every site and every component of these classes |
| cost | ~1 spec row + a re-render | a migration over component CSS; precedent exists (mig 415 repointed `article-body` across 97 placements / 20 sites) |
| risk | contradicts a written pin in `design_intent.guidance` ("KEEP THESE VALUES") | it is bugfix 122's mechanism and that lane is ACTIVE — must be contributed, not competed |
| leaves behind | components still fragile on any site whose primary is dark | the degenerate palette still degenerate, just no longer fatal |

**These are not exclusive and the honest recommendation is both, in order: A1 to make the site
legible now, A2 as the durable fix contributed to the 122 lane.** A1 alone leaves a landmine;
A2 alone is slower and still leaves a palette whose `primary` means nothing.

**Family B — components hardcode a light ground on a dark site** (7 components across `index`
and `about`; `background:#fff` / `255,255,255`). 24 of 44 failures. Not fixable by re-render.
Needs the components rebuilt against the palette, or the offending declaration removed so the
themed surface shows through. This is the `hardcoded_section_colors` class the design-discovery
agent already names, and the site carries an unresolved `generic_theme` item.

**Family C — `pricing` is unrepairable in place.** 5/5 NULL `content_data`, last rendered
2026-04-13, 8 firm failures including 7 invisible headings. Only a rebuild through the framework
fixes it (082 submission per the owner ruling of 2026-08-04 — never hand-built).

**Verification, whichever route.** Re-run R1 over the same 4 pages and require firm failures to
fall. Then let the site's own render audit run: it retracts parked `contrast_failure` rows on a
fresh positive measurement (`write_render_audit_findings_action.go:479`), so the 17 parked items
drain as evidence rather than being promoted by hand. **Do not promote the parked rows manually**
— they were parked by migration 389 precisely so completions could not be minted ungraded.

## 2. Images — generate, then bind

Scope is exactly one component (`case-studies-grid`) on two pages, 10 images.

1. **Do NOT fill in `handler_agent` on the existing `image_url_404` /
   `image_source_unsatisfiable` rows.** Those handlers triage; the one site they ran against
   lost its images (RUNBOOK R5).
2. Generate real images through `image-generator` / `image-build-handler`. The prompts are
   already written and good — `cardN_image_alt` on the component describes each intended diagram
   ("supervisor-worker agent topology deployed across Kubernetes pods…").
3. Bind the results into `cardN_image_url` and deploy to stable `/assets/images/` paths, **not**
   pre-signed URLs (§4).
4. Verify at the artefact over HTTP (R4), not at the work-item status.

## 3. Carousels — a planner/spec hint, with the failure mode designed out

Nothing exists. The hint belongs on the component-level build path (the one this site uses), not
`html_actions.go`'s whole-page prompt where the only existing guidance sits.

Constraints the hint must carry, each from something that already went wrong:
- **CSS-first** (scroll-snap / CSS animation), vanilla JS only if unavoidable — the existing
  guidance already says this.
- **Every control must resolve to a real page.** `bind_site_experience_action.go:36` records
  "four dead carousel destinations found by hand on 2026-07-26" (`bugs_open/023`, `071`). The
  experience register checks destination roles against `pages` at bind time; a carousel spec that
  routes through it cannot promise a dead page.
- **Degrade to a legible list without JS**, so a carousel can never become an invisible-content
  defect of the kind §1 is about.
- Candidate components: `case-studies-grid` (5 cards, already image-bearing), `departments-grid`,
  `leadership-team`, `info-card-grid`.

## 4. Expiring image assets — separate, time-boxed

All 9 hero/`content_hero` rows in `assets` are pre-signed Backblaze URLs with
`X-Amz-Expires=604800` stamped 2026-08-11 → **lapse 2026-08-18**. Only `og-card.png` and
`favicon.png` are stable `/assets/images/` paths. No page component references a pre-signed URL
today, so the immediate blast radius looks nil. **[UNVERIFIED]** whether og tags, feeds or the
asset renderer hold one. Check before assuming; do not mint new pre-signed URLs into content.

## 5. Sequencing

1. Owner picks the §1 route (A1 / A2 / both).
2. A1 if chosen — single spec row, re-render, re-measure with R1. Fast, visible.
3. Contribute the family-A finding to `bugfix_122_contrast_ink_slots` regardless of route: their
   mechanism is live here and insufficient, which is information that lane does not have.
4. Family B — rebuild the 7 offending components against the palette.
5. Images — generate + bind (§2).
6. Carousels — write the hint (§3), then exercise it on `case-studies-grid`.
7. `pricing` rebuild through the framework (§1 family C).

**Council gate:** §1 route A2 and §3 are platform-scope (shared component CSS; a change to what
the planner is told). Both go through `097_TRIGGER_council_review_v1.sh` before or alongside the
commit. §1 A1, §2 and §4 are site data and do not.
