# NOTES — brochure component library (append-only, newest at the bottom)

## 2026-07-20 — kickoff, initial checks

- Grepped the whole repo for `fundamentallyai` (code/docs/seeds/deployments,
  case-insensitive): **zero hits**. Confirms it has never been onboarded as a site
  here — nothing to reuse, nothing to conflict with.
- `WebFetch https://fundamentallyai.com` → 307 redirect to
  `https://www.afternic.com/forsale/fundamentallyai.com?...` — a domain-marketplace
  "for sale" parked page, not a live site. **[VERIFIED]** the domain is not currently
  the owner's. Recorded as a correction in the PLAN file, not silently absorbed into
  the brief.
- Checked `docs/leopardessconsulting/` for prior consultancy-site design lessons
  (README_where_we_are.md, PLAN_imagery_and_design_2026-07-18.md) since it's the
  closest existing precedent for "consultancy brochure site through this framework."
  Key transferable facts pulled into the PLAN file: house style is flat illustration
  via the Banana generator (`kind:"illustration"`), `kind:"hero"` historically fell
  through to SDXL which cannot render legible text (model-class limitation, not a
  prompt-quality problem), and the standing rule "anything carrying words or numbers
  is code-rendered [SVG] from evidence-base values, never generated" — directly
  relevant to any "stat band" / infographic-style component this workstream might
  propose.
- Read `docs024_key_docs_latest/036_REFERENCE_styling_render_pipeline.md` (existing,
  authoritative reference doc, not written by this workstream) for how components are
  actually assembled: `content_components.html_template` (Go template + inline
  `<style>`), resolved to a page section by function name via `plan_sections`
  (one active component per function), CSS assembled from the layout template +
  `css_snippets` (tagged `applies_to` — the existing extension point for new
  component CSS, e.g. carousel/hover-zoom rules) + a renderer-owned `--section-*`
  luminance block. Pages are static HTML artifacts (git → GH Actions → Backblaze B2),
  not server-rendered per request. **[FINDING, per that doc, not independently
  re-verified this session]** — treat as ground truth unless the Explore agent's
  independent pass contradicts it.
- Direct `WebFetch` of the three named reference sites: bain.com succeeded (full
  homepage component catalogue captured — hero triple-carousel, partnership
  spotlight card, an interactive "Mad Lib" industry/need selector, a 4-slide client
  case-study carousel with 2-3 stat callouts per card, a featured-insights strip,
  dual CTA promo blocks, email subscribe band, footer). bcg.com returned **HTTP 403**
  to WebFetch; mckinsey.com **timed out twice** (60s). Both are large JS-heavy sites
  that plausibly block or throttle a plain fetch — left to the `deep-research`
  workflow (run `wf_51d0513a-4d5`), which fans out via search rather than one direct
  fetch per site, and to secondary design-teardown sources it should surface. Do not
  re-try direct WebFetch on bcg.com/mckinsey.com repeatedly — two failures each is
  enough signal that the fetch path itself is blocked, not a fluke.
- Dispatched an Explore agent (background) to map: component-type registry, the full
  spec/mission → rendered-page chain, CSS/JS delivery mechanism (confirm whether any
  shipped component already carries real JS interactivity — none confirmed yet in
  this session's direct reading, the 036 reference doc's plumbing suggests
  CSS-template + inline `<style>` per component but says nothing about `<script>`
  usage, so **this is an open question, not yet answered either way**), imagery
  eligibility/kind logic, a worked recent component-addition example, and how a new
  site actually gets onboarded. Results not yet returned as of this note.
