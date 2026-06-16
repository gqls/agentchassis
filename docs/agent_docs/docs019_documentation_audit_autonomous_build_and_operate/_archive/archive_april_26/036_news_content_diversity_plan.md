# News Feed Content Diversity & Research Pipeline

## Problem

The homepage news feed draws from 4 sources (RSS, news_search, api_news/Grok, scrape) but in practice one source dominates because RSS is the most frequent and reliable. All sources search the same broad topic space independently, so they find the same stories. The result is 6 items from OilPrice rather than a mix of perspectives.

## Current Source Setup (gaswholesalers.com)

| Source | Type | Frequency | Typical yield |
|--------|------|-----------|---------------|
| OilPrice Energy News | rss | 6h | 5 items/run |
| Gas wholesale energy news search | news_search | 6h | 2-5 items/run |
| Grok energy news | api_news (xAI Responses API, web_search + x_search) | 6h | 4-5 items/run |
| OilPrice Latest News Scrape | scrape | 6h | ~20 raw, filtered by triage |

## Root Cause

Each source independently fetches "gas/energy news" without knowing what the others found. The render query picks the highest-scored most-recent items, which cluster around the same breaking stories from the most prolific source.

## Approach

### 1. Topic-Focused Source Splitting

Replace the single broad Grok source with multiple specialised sources, each with a different prompt template targeting a different angle of the vertical:

- **Geopolitical/supply chain** — OPEC decisions, pipeline disruptions, sanctions, shipping routes
- **UK/EU regulatory** — Ofgem, energy policy, carbon pricing, storage obligations
- **Market data/pricing** — spot prices, futures, spreads, seasonal patterns
- **Emerging/transition** — hydrogen, biogas, renewable gas, LNG infrastructure

Each is a `content_sources` row with a different `prompt_template` in its config. No code changes needed — just SQL inserts.

### 2. Coverage Gap Analysis

A pre-fetch step that reads existing feed items and identifies what's already covered before dispatching ingesters. The output is a list of saturated topics and under-reported angles.

This becomes a new step in the content-feed-orchestrator workflow, between `seed_sources` and `dispatch_sources`. The dispatch step passes coverage gaps as context to the api_news prompt templates.

### 3. Multi-Language/Region Discovery

Grok's X search can find content in any language. Regional sources surface stories that English-language feeds miss:

- German energy policy (drives EU gas pricing)
- Japanese LNG buyer behaviour (spot market signals)
- Arabic Gulf production decisions
- Norwegian pipeline/storage data

Each regional search is a separate `content_sources` entry. Triage needs a translation/summary sub-task — the LLM scores relevance from a translated summary, and the rendered item links to the original source with a note like "Originally reported by Handelsblatt (Germany)."

### 4. Triage Diversity Scoring

The triage prompt gets additional context: titles of items already in the display set. Items covering a different angle score higher. This is a prompt change to the feed-triage agent's `score_relevance` step.

### 5. Source Chain Provenance

When Grok finds a tweet that links to a Reuters article, the `source_attribution` chain records both the discovery path (X/@energy_analyst) and the original source (Reuters). This supports:

- Credibility assessment (tweet found it, Reuters wrote it)
- Deduplication (same story found via different paths)
- Display diversity (show the Reuters attribution, not "found on Twitter")

Already partially implemented in the triage scoring — `source_attribution.original_source`, `found_via`, `source_tier` fields exist in `content_feed_items`.

### 6. Render Interleaving

The render query uses `ROW_NUMBER() OVER (PARTITION BY source_id)` to interleave items from different sources. With 4+ sources and 6 display slots, each source contributes at most 2 items. Already implemented in the current `render_news_section_action.go`.

With topic-focused sources, this naturally produces topical diversity too — one geopolitical item, one regulatory, one pricing, etc.
