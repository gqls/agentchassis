# Design & Adoption: Problems To Fix

Reference case: gamedesign.uk adoption — original vs rendered output.

---

## A. Fingerprint Extraction (foundation — nothing downstream works without this)

1. **rawHTML is captured but never parsed for design data.** The crawl requests `["markdown", "rawHtml"]`, `crawlPageContent` stores both, but no action extracts colours/fonts/layout from the HTML. The design data that reaches the webdesign-agent is whatever the LLM infers from 500-char markdown summaries.

2. **The `analyze_site` LLM produces text descriptions, not concrete values.** The prompt says `"primary": "#hex or description"` — the LLM often gives "dark blue with warm accents" instead of actual hex values.

3. **The `design` spec written by `apply_adoption_plan` is the LLM's guess.** `specAspects["design"] = plan["design"]` — not parsed from actual CSS.

4. **The `needs_design` work item has an empty spec.** `spec: "{}"` — the webdesign-agent gets no design context from the adoption crawl.

5. **No `design_fingerprint` spec aspect exists.** No mechanism to pass concrete hex values, font stacks, or CSS variables from the crawled site to downstream agents.

6. **`<style>` blocks in rawHTML contain actual design tokens** (CSS variables, colours, font-family) but nothing reads them.

7. **Google Fonts `<link>` tags are in the rawHTML** but nobody extracts the font family names.

---

## B. Downstream Consumption (fingerprint exists but nobody uses it)

8. **The webdesign-agent doesn't read `adopt_from.design`.** That field doesn't exist yet. Even after producing the fingerprint, the webdesign-agent's prompt needs updating to use it.

9. **The site archetype `constraints` field isn't enforced** by the improvement loop. The archetype classification exists but isn't always persisted as a spec.

---

## C. CSS & Variable Mismatch

11. **The original's CSS was not used at all.** Original defines `--bg-color: #121212`, `--primary-color: #00bcd4`. Our version invented `--color-primary: #0f1923`, `--color-accent: #00d4ff` — different values, different naming.

12. **The font identity was lost.** Original uses `'Segoe UI', Roboto` (proportional system UI). Our version uses `'JetBrains Mono', 'Fira Code'` (monospace). The LLM guessed "game design = code font".

15. **The hero colour treatment differs.** Original: subtle gradient `#121212` → `#1e1e1e`. Ours: three-colour gradient `#1a1a2e` → `#16213e` → `#0f3460` — different colour family, more busy.

18. **CSS variable names don't match.** Original uses `--bg-color`, `--surface-color`, `--primary-color`, `--text-main`. Our system uses `--color-primary`, `--color-background`, `--color-text`. Components reference our naming convention, not the original's.

---

## D. Layout & Structure (the whole look and feel, not just colours)

10. **Page structure was replaced, not reproduced.** Original: hero → three pillar cards → resource grid. Ours: hero → generic features grid → CTA. The build pipeline mapped sections to generic component types instead of matching the original layout.

13. **Header was completely replaced.** Original: minimal dark header, brand + cyan `.uk` accent, three nav links, no CTA. Ours: gradient header, logo icon, eight nav items, "Get Started" button. Came from generic component library.

14. **Visual density and spacing are wrong.** Original is spacious with floating-panel cards (`box-shadow: 0 10px 30px rgba(0,0,0,0.5)`). Ours is compressed standard brochure.

16. **Content architecture was flattened.** Original has typed resource cards with colour-coded badges ("Tool", "Sim", "Guide", "Interactive"). Ours: uniform features grid with icon placeholders. Information hierarchy lost.

17. **Footer was genericised.** Original: one-line centred footer. Ours: four-column footer with service lists, contact, legal links. Pulled from generic footer component.

---

## E. Root Causes

19. **No mechanism to prefer the original's layout over generic components.** The adoption plan identifies sections, but the build pipeline always maps them to the component library. Nothing says "reproduce three pillar cards overlapping the hero".

20. **The design spec is disconnected from the build pipeline.** Even when the LLM correctly identifies "dark IDE aesthetic, cyan accent", the webdesign-agent and component selector don't constrain themselves to that.

---

## Suggested Work Order

**Phase 1 — Extract** (problems 1-7)
- Build the `extract_design_fingerprint` Go action
- Insert into adoption workflow
- Write fingerprint as spec in `apply_adoption_plan`

**Phase 2 — Consume** (problems 4, 8, 11, 12, 18)
- Pass fingerprint into `needs_design` work item spec
- Update webdesign-agent prompt to read and use `adopt_from.design`
- Map original CSS variable values to our variable naming convention

**Phase 3 — Structure** (problems 10, 13, 16, 17, 19)
- Capture layout patterns in fingerprint (card types, section arrangements, header/footer complexity)
- Feed layout data to component selector / content writer
- Allow adoption to generate custom components when library has no match

**Phase 4 — Fidelity** (problems 14, 15, 20)
- Spacing/density extraction and reproduction
- Gradient and visual treatment preservation
- Audit loop respects adopted design constraints
