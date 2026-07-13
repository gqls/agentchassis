# TODO — chassis scheme work + idea.uk + carried backlog (annotated)

Consolidated from this chat, with each item explained for a reader coming in cold. Companion docs: `REPORT_scheme_does_not_reach_components.md` (the P0 investigation plan in full), `HANDOFF_scheme_to_components.md` (cold-start brief for the P0 thread), `RUNBOOK_idea_uk_chassis_site_and_vm_deploy.md` (idea.uk steps 1–8 + how we got here), `running_notes_2.md` (the running journal/evidence). idea.uk chassis site_id `1244516d-014d-421c-88c6-090bb1e9552a`; the live £29 product is a separate Go binary on a Hetzner VM and is untouched by any of this (the chassis build deploys to Backblaze B2, and DNS still points at the VM, so chassis changes are invisible to the live site).

---

## Done this chat (context — so you know the starting state)

- [x] **Scheme-aware weighted layout matcher — live.** The old layout matcher picked a layout purely on tag overlap and was blind to light/dark. We replaced it with a weighted, scheme-aware matcher (in `fork_theme_composition.go` + `resolve_composition_layout_action.go`) that treats a site's light/dark scheme as a near-hard constraint, and it is deployed in production.
- [x] **`layouts.scheme` column + new `tool-portal-light` layout** (`migration_layouts_scheme_and_light_tool_portal.sql`, applied). Added a curated light/dark/neutral property to layouts, and authored a light counterpart to the existing dark `tool-portal-dark` so light sites have a target.
- [x] **idea.uk re-resolved in place onto `tool-portal-light` + parchment** (runbook steps 1–4). We detached and cleared idea.uk's old (dark) composition and re-triggered the site-design-planner; it now resolves to the light layout with a site-specific parchment palette, with no library-gap flag.
- [x] **`styles.css` rendered, deployed and verified** (runbook steps 5–6; commit `05ef817`). The webdesign-agent rendered the stylesheet from the new composition and git-committed it (which deploys to B2). Verified it is exactly `tool-portal-light` with the parchment palette and no LLM drift.
- [x] **Established exactly how pages are built and root-caused the dark pages** (runbook step 7; notes lll–ooo). Traced the build end to end and found why the pages still render dark despite the correct stylesheet (see P0).
- [x] **Wrote the structural-gap report** for the dedicated fix thread.

---

## P0 — FUNDAMENTAL: a site's scheme does not reach its components  (dedicated thread)

**What it is.** A site's scheme (light/dark) is decided at composition and flows into the stylesheet's `:root` colour variables, but it never reaches the components that actually render each section, header and footer. Those components come from a dark-oriented library via a one-active-component-per-function lookup, and they self-style with inline CSS that hardcodes a dark treatment — so a light-resolved site still renders dark chrome (dark header, dark image hero, dark CTA, dark footer) over its light content. **Why it matters.** It blocks the platform from producing genuinely-light sites at all (idea.uk is just the case that exposed it), and it directly undercuts the design quality the mission promises. It is framework-level and must not be patched on a single page. The full plan is in `REPORT_scheme_does_not_reach_components.md`; the cold-start brief is `HANDOFF_scheme_to_components.md`.

- [ ] **Open the dedicated thread and orient.** Start from the handoff brief plus the report's §7 (a single up-front list of every agent definition, Go action, doc, schema and query to pull) and `running_notes_2.md`. The point is to begin from the established facts, not re-derive them.
- [ ] **Run the nine investigations (report §6).** Each is scoped to answer specific design questions and lists the exact files/queries it needs: (A) the precise render path for a section and a site component; (B) trace the scheme signal end-to-end and find where it stops; (C) inventory the whole component library against scheme (how many hardcode dark vs read the variables); (D) audit the layout-stylesheet ↔ component class-name contract; (E) the section-contrast model; (F) header/footer scheme + the `update_site_defaults` wiring; (G) the existing `--section-*` luminance mechanism; (H) migration + backfill safety; (I) a scheme-coherence audit guard.
- [ ] **Answer the eight design questions (report §5) before writing code.** These are the genuinely hard decisions (e.g. where scheme lives at render time; whether a section's darkness is a site property, a component property or a per-placement choice; the override mechanism). The gating one is Q4 — whether components should adopt the layout's class vocabulary so the stylesheet can style them, or stay self-styled but strictly variable-driven — because it determines the shape of everything else.
- [ ] **Validate or revise the provisional fix shape (report §8).** It is stated as a hypothesis to test or reject, not a plan to build blindly: introduce a scheme signal at render time → de-hardcode the dark components so they read the variables → make the renderer the single point that sets each section's light/dark treatment → make header/footer adaptive and wired → keep direct-function resolution → migrate and back-fill carefully → add the audit guard.
- [ ] **Hold to the override-not-variants steer.** Express the light/dark difference as a variable-value override (palette + `--section-*`) consumed by one component; create new component functions only where a component is genuinely too structurally different to share. Do not duplicate components into `*-light`/`*-dark`.

---

## P1 — idea.uk  (minimum now; finish after the P0 fix)

- [x] **Minimum now: no change, and that's deliberate.** idea.uk's composition and stylesheet are already correct; only the page chrome is dark, and that's the P0 gap. The active `hero` component already carries the corrected `--color-accent` variable (the inactive twin had the buggy `--accent-color` that fell back to navy), so whenever a rebuild eventually runs it will render the rust button for free. Nothing is gained by touching it before P0.
- [ ] **After P0: rebuild idea.uk's pages.** Once components are scheme-aware, re-run the page build so the pages pick up the now-light components and the fixed hero. Reuse the existing pattern — `flag_page_image_rebuild` emits a `needs_page` work item to `page-build-handler`, which re-runs `plan_sections` and re-renders — rather than inventing a new trigger.
- [ ] **Re-verify the deployed pages.** Confirm the pages on the B2 build read light/parchment throughout with no stray blue, i.e. the P0 fix actually reached idea.uk.
- [ ] **Review the built site before cutover.** Check the calls-to-action resolve to the live tool path `/request` (not a phantom `/contact`); that nothing collides with the reserved tool paths (`/request /confirm /approve /decline /stripe/webhook /internal/* /order/*`); and the outstanding content gaps (empty differentiator cards, pricing, a dead contact form).
- [ ] **VM cutover** (`RUNBOOK_idea_uk_vm_cutover.md`) — gated on P0 + the site review. This is the deliberate go-live: put nginx in front so general pages are served statically from the build while the reserved tool paths reverse-proxy to the live Go binary on `127.0.0.1:8080`. The biggest risks are reserved-path completeness and the Stripe webhook — prove the webhook works through nginx before cutting over; DNS is unchanged; rollback is restoring one nginx block. The live £29 binary keeps running throughout.

---

## P2 — carried chassis backlog  (independent of P0; can proceed in parallel any time)

- [ ] **Apply + test the build-standard classifier migration** (`migration_classifier_build_standard.sql` — proven correct by simulation, NOT yet applied). It prepends a best-in-class quality+fit instruction block (explicitly not scope) to the top of the classifier's `classify_and_extract` prompt, to raise design/content quality on new builds. Test it on a fresh build first, and confirm an adopted-site rebuild (e.g. gamesdesign) stays faithful to the original rather than drifting.
- [ ] **Deploy the dead-slot hardening** (three Go files). These add a `design_reference` fingerprint fallback to the palette/typography cascade so that when `design_intent` values are absent the resolver falls back to the adopted site's fingerprint instead of a generic default. Requires a chassis image rebuild and rolling `site-design-planner`.
- [ ] **improver-not-rewriter overlay for `webdesign-agent.analyze_design`.** Today the design step can re-invent the palette; the overlay makes it show the already-established palette plus a diff and an audit, so it improves rather than rewrites. Check first whether a suitable audit/log table already exists before adding one. (Targets `webdesign-agent`, not the separate `site-scraper`.)
- [ ] **Populate the remaining `layouts.scheme` rows** (~15 layouts still NULL) from each layout's real background colour. The query is already in the layouts migration; this just curates the rest of the library so the scheme-aware matcher has accurate data for every layout.

---

## P3 — known hazards / smaller items

- [ ] **A content rebuild silently de-tools a tool page (confirmed hazard, fix pending).** A `needs_page` / `link_resolution_rebuild` for a tool or game page is handled by `page-build-handler` → page-content-writer, which regenerates the page from `plan_sections`. The plan has no knowledge of the interactive tool (it lives as a section's `rendered_html`), so the tool is replaced with generated prose and the page falls back to a generic text block. Fix direction: route link maintenance through a preserve-sections re-render path, stamp `source_item_id`, and add an interactivity-aware save guard. This is **relevant to P1**: if idea.uk gains interactive tools before its post-P0 rebuild, that rebuild could hit this. See 005 / 016b / 020 / 026.
- [ ] **Smaller idea.uk runbook gaps:** empty differentiator cards, unresolved CTAs, a dead contact form, thin nav/footer and empty meta tags, and a full fresh rebuild as an end-to-end test. **Parked:** rewriting the £29 report's language/format.

---

## Quick status line
P0 (scheme reaches components) is the blocker for a genuinely-light idea.uk and for the VM cutover, and is going to its own thread. Everything in P2 is independent and can move in parallel. idea.uk is safe exactly as it is (staging only; the live tool on the VM is untouched).
