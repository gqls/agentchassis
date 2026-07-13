# PLAN — brief-explanation

**Tool/component function:** `brief-explanation`
**Component id:** 58363894-9db9-4d2f-81ac-c47b54d97fc3
**Site(s) using it:** vonc.com (Spark) — index page
**Created:** 2026-07-01 (retroactively, from session history)
**Status:** Broken (Mode-B empty shell); regeneration in progress.

---

## Aim
A concise "how it works / what is Spark" explainer on the index page — tells a
first-time visitor what the daily game is and how to take part, in three quick
steps, with a couple of stats and the two primary actions. STATIC brand content
(does not change day to day).

## Source spec (Spark v1 roadmap)
index section_types include `brief-explanation`. Purpose: a quick, energetic
explanation of the game (arena metaphor, AI-as-producer). Companion to the hero and
the provocation card on the landing page.

## Behaviour contract
STATIC content, filled at BUILD time by the content writer (NOT a runtime loader —
that distinction matters: this is not daily-changing data, so it belongs in the HTML
for SEO and no-JS robustness). Intended structure:
- eyebrow label; heading (with one emphasised word/phrase); short description
- an ordered list of EXACTLY THREE numbered steps (each a short bold title + a
  one-sentence explanation of the daily flow)
- a row of EXACTLY THREE stats (value + label)
- two CTAs (primary + secondary)
- an illustrative image (from the site illustration asset) with a small badge

## Delivery mechanism
**Build-time content.** component-creator generates the template + input_schema;
`page-content-writer` fills the Tier-A voice fields at page build. No JS loader.

## Field tiers (target schema)
- Tier A (voice/llm): eyebrow, heading, description, step1/2/3 title + text, badge,
  CTA labels.
- Tier B (tunable labels + fallback): the three stat values and labels.
- Tier C (site data): the image (site illustration asset), image alt.

## Dependencies
- component-creator (generation) → in-place regeneration of id 58363894.
- page-content-writer (content fill) via a needs_page rebuild of the index.
- site illustration asset (`/assets/images/illustration.jpg`).

## Deliberate decisions
- **Regenerate, do NOT add a JS loader.** brief-explanation is static; Option-2
  (runtime JS) is only for daily-dynamic shells (provocation-card, lobby-grid).
- Regeneration produces a FRESH template (component-creator doesn't see the old
  one); the new markup/CSS will differ from the current broken shell — acceptable,
  since the shell is unusable. The description pins the intended structure.
