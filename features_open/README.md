# features_open/ — designed but not built

Created 2026-07-19, mirroring `/bugs_open/` and `/bugs_closed/`.

## What belongs here

Forward-looking work: **features we intend to build, and risks in designs we
intend to build**, captured at the moment the thinking happened so it is not
lost between sessions.

## What does NOT belong here

**Anything biting production now — that is `/bugs_open/`.** The two directories
answer different questions and must not blur:

| directory | answers |
|---|---|
| `/bugs_open/` | what is broken in production right now |
| `/bugs_closed/` | what was broken and is now fixed **and live** |
| `/features_open/` | what we have decided to build, and what will bite us when we do |

A latent defect in an unbuilt design is a **`RISK_` file here**, not a bug. It
becomes a bug the day the design ships and the defect is reproducible in prod.

## Naming

`NNN_FEATURE_<slug>.md` — something to build.
`NNN_RISK_<slug>.md` — a known hazard in something we are about to build.

Numbering is its **own** sequence, independent of the bugs directories (those
two share one sequence between them). Never reassign a number.

## What a good entry holds

The originating intent in the owner's own framing, what makes it non-obvious,
the open questions, and — for a `RISK_` — how you would *measure* whether the
hazard has materialised. A risk with no test is an opinion.

## Index

| # | type | title |
|---|---|---|
| 001 | FEATURE | [Packaged topic features — living dossiers](001_FEATURE_packaged_topic_features.md) |
| 002 | RISK | [Portfolio-wide duplicate content from pooled feeds](002_RISK_portfolio_duplicate_content.md) |
| 003 | FEATURE | [Paid tier — per-site sources and beyond](003_FEATURE_paid_tier_beyond_news.md) |
| 004 | FEATURE | [Duplicate-content council seat](004_FEATURE_duplicate_content_council_seat.md) |
| 005 | FEATURE | [Pilot onboarding + first pool activation](005_FEATURE_pilot_onboarding_and_first_pool.md) — **ON HOLD, owner gate** |
| 006 | FEATURE | [gaswholesalers.com repositioning + "AI influence" page](006_FEATURE_gaswholesalers_repositioning_and_ai_influence_page.md) |
| 007 | FEATURE | [AI-advisory chatbot, freemium/paid](007_FEATURE_ai_advisory_chatbot_freemium.md) |
| 008 | FEATURE | [Halve image-generation cost via the Gemini Batch API](008_FEATURE_image_generation_batch_api.md) — **deferred, owner: revisit on volume** |
