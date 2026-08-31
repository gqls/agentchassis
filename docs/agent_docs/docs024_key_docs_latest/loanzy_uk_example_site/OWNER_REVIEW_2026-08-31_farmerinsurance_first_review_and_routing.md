# OWNER REVIEW — farmerinsurance.uk, 2026-08-31: six findings, verified, joined to the council's held verdicts, and routed

**The owner reviewed the served site** (39 active pages; invented-path control 404s, so the HTTP
readings are informative). Each finding below carries: his words in substance, the verification at
the artefact/spec, the council-verdict sibling where one exists (evidence the acceptance council
sees what he sees), and the routing. **Four of six findings have HELD verdict siblings already in
the queue** — filed by the seats within minutes-to-hours of the build, dispatched to nobody, exactly
as designed. Where a verdict sibling exists, the natural fix path is the RUNBOOK's release recipe on
THAT row, not a fresh filing.

## 1. Copy: outdated agent/flow — negative framing + explaining the AI request instead of talking to the user
- **His example, verified live on `/index.html`:** *"so every guide starts from the risk on your
  land and works outward to the policy wording"* — methodology meta-commentary, not user-facing copy.
- **Council sibling (root cause, held):** content-quality-audit — *"The site has no tagline, stated
  industry, target audience, or tone definition"* (the tone vacuum the writer filled with meta-voice).
- **ROUTED: `copy_quality_two_stage`** (owner named them) — message sent 2026-08-31 with examples;
  their deep-refresh + every-prompt audit is the standing owner instruction this lands inside.

## 2. The logo reads "Farm Shield Info" — someone else's brand; ours is farmerinsurance
- **Verified:** identity spec `company_name = "Farmer Insurance UK"` (CORRECT); "Farm Shield"
  appears in **zero** specs; the served logo's alt is `farmerinsurance.uk`. **The logo image's text
  was invented by the asset generator, unmoored from identity** — and nothing verifies rendered
  logo text against the identity (no seat reads pixels). Third-party brand risk, owner-flagged.
- **ROUTED:** imagery/brand-asset machinery (`derive_brand_head_assets` family) — regenerate the
  logo from `identity.company_name`; the systemic half (logo-text-vs-identity verification) noted
  for the vigilant/designer thread. This lane files the regeneration through the framework.

## 3. News is American — owner wants UK news default for all .uk/.co.uk sites (a region flag)
- **Verified:** farmer's `content_sources` are regionless search queries ("News Search: insurance
  regulation / claims / premiums / insurance market") → US results dominate.
- **Council siblings (held):** site-review AND content-quality both filed the news section serving
  **malformed Google redirect URLs** instead of real content — same subsystem, second defect.
- **ROUTED: `news editorial` lane** — feature ask: region default (UK for .uk/.co.uk TLDs, a flag),
  plus their existing redirect-URL defect siblings to fold in.

## 4. Directory-style entity mentions should LINK OUT (e.g. Drewberry → drewberry's site)
- **Verified:** "Drewberry … FCA-authorised independent insurance broker …" appears as bare text
  (blog page `finding-a-farm-insurance-broker`), no outbound link.
- **Council sibling (held):** offer-analysis — *"The growth_path specifies a verified directory of
  FCA-regulated agricultural insurance [brokers…]"* — the directory the entity mentions belong in.
- **ROUTED: `bugfix_206_directory_build_handler` lane** (entity directories) via CONTRIB — the ask
  joins their existing "entity-directory pages have no real builder" case; plus the writer-side
  rule (an entity described gets its outbound link) for the copy lane's prompt audit.

## 5. Nicer components: CAROUSELS as a default pattern, not card-stack-after-card-stack
- **His words in substance:** different carousel types (same cards fine) as the default, because
  card-after-card scrolling on mobile is a bad experience.
- **Verified:** homepage = **0** carousels, **72** card-class elements.
- **Council sibling (held):** brief-fidelity ×2 — layout does not match the design intent's
  "Editorial hub layout — featured guide(s) prominent / clear category navigation".
- **ROUTED:** `offer analyser benefit analyser visual designer` (vigilant/designer thread — owner
  named it; their in-body-image + visual-designer placement design is the adjacent work) + CONTRIB
  to `staged_component_build` (the component-library maker) for carousel component types; the
  experience-loop machinery reads this doc via the council joins above.

## 6. (Standing from 08-27, still open) Seven unresearched `add_tool` rows + the growth-refusal decision
- Unchanged asks on the owner's desk: park-or-build the 7 farmer tools; the per-site growth
  refusal (FOLLOW-UP 1, now three worked cases: apis.uk, farmer's tools, and — by shape — the
  Farm Shield logo: generators inventing what no spec asked for).

**The meta-reading, for RFC_056's file:** the owner's six findings vs the council's held verdicts =
4 overlaps, 2 gaps (the meta-voice copy itself; the logo text). The gaps are both *seats that would
need to read what the owner reads* (prose register; image pixels) — the reader seat's prompt can
learn the first; the second needs an OCR-shaped check nobody has built. Both noted in FOLLOW-UPS.

---

## ROUTING OUTCOMES — appended 2026-08-31, same day, after the lanes replied

- **§1 copy — TAKEN** by copy_quality_two_stage. The tone-vacuum reading enters their audit
  question set as a new mechanism variant (an absent identity/tone doesn't yield neutral copy;
  the writer fills the vacuum with its own method). Farmer becomes worked case #2 for their
  owner-ordered benefit-led co-design, and a candidate for their model screen (their
  pre-registered benchmark: fable-5 scored NEG 0 on the worst canary section's exact
  production prompt vs shipped sonnet's 5).
- **§2 logo — DONE here + FILED.** Root cause traced end-to-end and filed as
  **`bugs_open/417`**: build-site-planner's worked-example logo prompt licenses "a wordmark"
  it never names → banana invents brand text; it contradicts the estate's own ruled default
  (`default_brand_prompt.go`: "no lettering — generated wordmarks reliably produce malformed
  text"). Farmer's regeneration filed THROUGH THE FRAMEWORK: `needs_imagery` item
  `3740f5f2-e6ff-4dd3-b60d-8ba502b1c636` (prompt names the exact wordmark
  "farmerinsurance", positive-framed per closed 028). **Owed after it lands:** favicon +
  og_card re-derivation (they derive from the logo and presence-based discovery will not
  refile them). The systemic fix (exemplar migration) is council scope — flagged, not shipped.
- **§3 news — RE-ROUTED, my miss.** news_editorial_features is editorial FEATURE pages, not
  feed ingestion — I matched a lane NAME, not a lane (they said so, correctly, and verified
  the mechanism anyway). Their measurement: **zero region keys across all 48 fleet
  news_search configs** — the capability is ABSENT, not undriven; seam `web_search_action.go`;
  TLD default belongs at seed time; affected-population count owed before it ships. Now a
  CONTRIB in `bugfix_316_news_feed_ordering/`.
- **§4 directory — CONTRIB filed** in `bugfix_206_directory_build_handler/`; writer-side
  entity-link rule also noted with the copy lane.
- **§5 carousels — INVERTED TWICE by measurement (offer-analysis lane), and the ask has
  nearly vanished as a build:** (i) two carousel components already exist (1 live instance
  between them) and `component_expresses` has no traversal token — planner-blindness, the
  381/IMG-074 shape; then (ii) **`info-card-grid` — the 42-instance grid itself — already has
  a declared boolean `carousel` field, ON for exactly 1 of 42.** The owner's ask is closest to
  a switch that exists and is off. CONTRIBs in `staged_component_build/` (theirs + mine).
  **Two decisions are the owner's:** default the flag on fleet-wide or not (nothing resolves
  it automatically), and whether carousel-as-default is right (items behind an interaction
  are items many readers never see — his mobile-UX reason vs that trade).
- **Verdict release:** both peer lanes declined to release the held brief-fidelity/offer
  verdicts off a relay — right call; release stays the owner's verb.
