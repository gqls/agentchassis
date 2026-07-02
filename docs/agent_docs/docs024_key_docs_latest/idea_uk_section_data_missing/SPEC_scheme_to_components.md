# SPEC — scheme reaches components (fix specification)

Date: 2026-07-02. Evidence base: RUNBOOK Checks 1–4 + running notes Sk–So. Nothing below has been applied; the live idea.uk £29 VM is out of scope and untouched throughout.

## Decision record

The standard is **paired variables** ("on-colour" pairs): every paintable band colour has a matching text colour, curated per layout (and therefore per scheme), overridable per site through the palette's specialised slots (merge rule: theme-wins), with per-instance control later available via `site_plan_directives` scope=section if ever needed. This is not new architecture — Checks 4b/4c show it is the existing library standard incompletely adopted: 18/18 layouts define `--color-primary-text`, 17/18 define `--color-cta-text`, all define the header/footer pairs, several light layouts deliberately curate dark footer bands, and the legacy base `<head>` CSS already consumes `--color-cta-bg`/`--color-cta-text` and defines `--color-hero-title`/`--color-hero-subtitle`. The work is bringing the components (and one layout) into line with it.

Consequences: scheme is the base and light sites can render fully light; dark bands remain available as layout/palette choices; structural dark (hero image + overlay) is preserved; `buildSectionDefaults` stays unchanged as the whole-palette-darkness base; `is_dark_section` is demoted to selection/imagery metadata (6 of 37 declarers contradict their own flag — never key styling on it); `SectionStyles` stays retired; 025 Phase 4.5 (`data-section-bg` surface generalisation) is deferred as a separate dark-site concern.

**Variable names: reuse the existing ones.** Pairs already named in layouts/base CSS: `--color-header-bg/--color-header-text`, `--color-footer-bg/--color-footer-text`, `--color-cta-bg/--color-cta-text`, `--color-primary/--color-primary-text`, `--color-hero-title/--color-hero-subtitle`. Do not invent new names.

## The contract (replaces 003 "Dark Section Contract" checklist item 6; goes into the component-creator prompt)

1. Section and chrome components contain **no literal colours** (hex, `white`, `#fff`, `rgba(255,…)`) for text or backgrounds. Exception: an image-overlay treatment may use `rgba(0,0,0,x)` overlay gradients with white text — the overlay guarantees darkness (structural dark).
2. Backgrounds come only from: transparent (inherit the page) | `var(--color-background)` / `var(--color-surface)` | a semantic pair background (`--color-cta-bg`, `--color-header-bg`, `--color-footer-bg`) | a palette band (`--color-primary/secondary/accent`) used together with its on-colour.
3. A component that paints a band **re-exports the band's text context as variable references, never literals**, so the global element rules (`var(--section-*, var(--color-*))`) follow the pair. Canonical form:
   ```css
   .cta-section {
     background: var(--color-cta-bg, var(--color-primary));
     --section-text:       var(--color-cta-text, var(--color-primary-text, #fff));
     --section-heading:    var(--color-cta-text, var(--color-primary-text, #fff));
     --section-text-muted: color-mix(in srgb, var(--color-cta-text, #fff) 82%, transparent);
     --section-surface:    color-mix(in srgb, var(--color-cta-text, #fff) 8%,  transparent);
     --section-border:     color-mix(in srgb, var(--color-cta-text, #fff) 25%, transparent);
   }
   ```
   (`color-mix` is baseline-available in all evergreen browsers; if it is ruled out, add `--color-cta-text-muted` etc. to the pair set instead — decide once in W1.)
4. A component that does **not** paint its own background declares **no** `--section-*` at all.
5. Naming contract unchanged: root `data-component="{function}"`, kebab-case function.

## Workstreams

**W1 — layouts (small, first).** Add the CTA pair to `tool-portal-light` (`{{palette "cta_bg" "<value>"}}` + `{{palette "cta_text" "<value>"}}`), values curated to the layout in keeping with siblings (an accent band in the parchment family is consistent with soft-editorial's `#9b4020`; requirement: pair contrast ≥ 4.5). Sweep query across all 18 layouts: extract each `cta_bg`/`cta_text` pair and list any failing 4.5 contrast; check whether any layout defines `--color-hero-title`/`--color-hero-subtitle` (the base head CSS does; layouts may not — if absent, hero text falls back per rule 3). One `webdesign-agent`/`render_css_from_spec` re-render per affected site afterwards refreshes styles.css.

**W2 — hazard-class components (~18): straight bug fixes.** These declare dark `--section-*` while painting surface variables or nothing — white-on-light today. Set: `site-footer` (first — live on idea.uk), `site-head`, `header-docs`, the five `hero-*` page variants (verify each has the image-overlay-or-surface branch like the main hero; if the image can be absent, the no-image branch must be surface/inherit, not transparent-with-white), `contact-block`, `platform-comparison`, `lobby-grid`, `brief-explanation`, `game-master-explanation`, `provocation-feed`, `gripper-payload-calculator`, `archetype-grid`, `archetype-taster-quiz`, `tool-agent-complexity-estimator`. Fix = remove the dark declarations; text flows through the element rules; where an element paints from a pair (the footer's `--color-footer-bg`), re-export per rule 3 using the pair's text (`--color-footer-text`). Flags were per-template — read each template per-element before editing.

**W3 — band-class components (~19): pair conversion.** `call-to-action` first (idea.uk-visible): background `var(--color-cta-bg, var(--color-primary))`, context per rule 3, buttons `var(--color-cta-text)`-derived instead of `--color-white`. Hero: image branch untouched; gradient branch keeps the primary/secondary/accent gradient but text becomes `var(--color-hero-title, var(--color-primary-text, #fff))` / subtitle equivalent, context re-exported as references. Then the rest (social-proof, testimonials, system-stats, product-hero, case-studies-grid, gauntlet-*, tool-cta, tool-* calculators) — per component decide band-with-pair vs surface treatment; default to surface/inherit unless a band is the design intent. **Expected visual diff on dark sites:** the CTA moves from a primary-coloured band to the layout's curated cta band (e.g. tool-portal-dark `#1e1e1e`) — intended, but visible; note it in the change description.

**W4 — chrome path repair.**
(a) Structural, reuse-first: where `site_components.component_id` is NULL or points at an inactive component, `renderAndStoreSiteComponent` falls back to `GetComponentByFunction("site-header"/"site-footer")` (the same fallback `RenderHeader` already uses) with a `logger.Info` — so `site_components` can no longer pin dead components silently. `RenderFallbackHeader`/`RenderFallbackFooter` (component_library.go:1375/1405) swap `ctx.PrimaryColor` + literal white for `var(--color-header-bg)`/`var(--color-header-text, var(--color-text))` and footer equivalents (~10 lines each).
(b) idea.uk data: repoint `site_components.component_id` — header → the active generated `site-header`, footer → the active generated `site-footer` AFTER its W2 fix, head → decided by the InjectHead read (open item) — then `render_site_components` with `force_rerender`.
(c) Composition: keep function-lookup as the norm (it is what actually runs); either populate `style_collections.header/footer_component_id` at install later as a per-site-variety feature, or delete the misleading "webdesign-agent populates these later" comment. Smallest honest change: the comment.

**W5 — authoring + backstop alignment.** Rewrite 003's Dark Section item 6 to the contract above; update the New-Component Checklist; put the contract into the component-creator prompt (the generated footer proves the prompt currently emits half-migrated components). Re-aim `fix_forced_text_colors` to enforce the same contract (strip literal text colours; for band painters insert reference-declarations, not white literals; key on the template's own painting, not on `is_dark_section`); review `fix_hardcoded_colors`' dark-hex→`--color-primary` mapping (map to the appropriate pair or surface instead). The fixers remain improvement-loop backstops — never required for a correct first render.

**W6 — idea.uk sequence + verification.**
1. W1 layout patch → styles.css re-render.
2. W2 footer + site-head; W3 CTA (hero already correct via image branch).
3. W4b repoint + `render_site_components force_rerender`.
4. Full rebuild via `site_work_items` (pipeline=build, handler_agent=page-build-handler, status=triaged) — NOT `needs_rerender`: Check 4 proves rerender re-fossilises stale section renders (deployed hero still consumes legacy `--accent-color`).
5. Verify by re-running the Check 4a grep on the freshly deployed index.html — expect `site-header-section`, `site-footer-section`, hero consuming `--color-accent`, footer light, CTA per the pair.
Rollback: the site repo is git (revert); component edits are `UPDATE content_components SET html_template` — snapshot each row (copy html_template to a fork or a backup table) before editing, per the snapshot pattern.

## Open items (small, non-blocking)
- What ran on 2026-07-01 12:49–12:58 (DB shows deploy, git shows no new commits — likely a no-diff rerender or deploy-only touch). Identify before trusting `last_deployed_at` as a build signal.
- `InjectHead` (component_library.go:1661) selection logic — read before repointing the head slot; two head-ish components exist (`head` inactive, `site-head` active but section-level and in the hazard list).
- Planned-but-unbuilt pages (`news-index`, `guides-index`, `tool-audience-check`) are linked from the deployed nav/footer — build them in W6 or unlink.
