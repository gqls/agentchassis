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
