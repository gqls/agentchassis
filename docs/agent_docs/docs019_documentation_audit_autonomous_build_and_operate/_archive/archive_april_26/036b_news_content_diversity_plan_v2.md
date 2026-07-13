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

---

## Original Research & Writing Pipeline

The sections above improve diversity of *sourced* news. This section describes producing *original* content — researched, written, and published articles that serve specific readership segments better than generic trade press.

### Readership Model

"Gas industry professionals" is too broad. The actual readership segments for a site like gaswholesalers.com might be:

- **Procurement managers** — price movements, contract terms, supply reliability
- **Operations/logistics** — infrastructure status, pipeline capacity, storage levels, delivery schedules
- **Finance/trading** — futures spreads, hedging strategies, regulatory cost impacts
- **C-suite/strategy** — market structure changes, M&A, policy shifts, energy transition

Each segment wants the same underlying events framed differently. Hormuz closure → procurement sees supply risk, trading sees price opportunity, operations sees delivery disruption, strategy sees long-term diversification argument.

The classifier already produces `target_audience` and `industry`. This needs extending to describe readership segments with enough detail to drive content targeting — what each segment cares about, what decisions they make, what data they need. If the current classification doesn't have this depth, the classifier prompt can be enriched to produce it, or a separate `readership_profile` spec aspect can be created.

For a given vertical, the number of meaningful angle combinations is roughly:

- ~10-20 recurring topic areas per vertical
- ~3-4 readership segments
- ~50-80 specific angle combinations per site

Across all sites and verticals, this reaches hundreds. The structure is the same everywhere — the content is parameterised by classification and readership.

### Research Agent Pipeline

For each news angle, a multi-step investigation rather than a single "find news about X" prompt:

```
Topic monitor detects signal (price moved, tweet from OPEC, regulatory filing)
  → Research agent: gather context from multiple sources
      → What happened? (fact gathering — multiple sources)
      → What's the history? (has this happened before, what was the impact)
      → Who said what? (official statements, analyst quotes, X discourse)
      → What are the numbers? (price data, volume data, contract data)
  → Writer agent: produce article draft targeted at readership segment
      → Frame for the audience (procurement angle vs trading angle)
      → Include actionable insight ("if you're hedging Q3, consider...")
      → Source attribution throughout
      → Reasoning explained — why we think X follows from Y
  → Eval agent: score against quality criteria
      → Factual accuracy (claims match cited sources?)
      → Audience fit (would a procurement manager find this useful?)
      → Originality (does this add value beyond summarising existing articles?)
      → Tone/compliance (no financial advice, no speculation presented as fact)
      → Structure quality (news hook → data → context → implications)
  → Publish or reject
```

The research depth is the key differentiator from current feed ingestion. Current api_news does one LLM call. The research agent does a multi-step workflow with verification.

### Continuous Timelines

Each topic thread maintains a running timeline — a persistent, evolving dataset that builds historical credibility over time. For a gas wholesale site, a thread might be:

- **NBP spot price** — daily data point, annotated with the news events that moved it
- **UK storage levels** — weekly data from Grid data, annotated with demand/supply context
- **LNG cargo arrivals** — shipping data, annotated with geopolitical context
- **Regulatory calendar** — Ofgem consultations, policy deadlines, annotated with market reactions

Each timeline produces:

1. **An interactive graph** — price/volume over time with labelled spikes and troughs. Each label links to the article that covered the event. The graph itself becomes valuable reference material that readers return to.

2. **Annotated history** — every data point has context: "Price rose 8% on 14 March — OPEC announced production cut of 500k bpd (see our coverage)." Over months, this builds a searchable, cited record that no single article provides.

3. **Pattern recognition** — as the dataset grows, the system can identify patterns: "The last 3 times UK storage dropped below 40%, NBP spot rose 12-18% within 2 weeks." These become the basis for scenario analysis.

### Scenario Analysis

Rather than predicting prices (which we shouldn't do), we can present structured "if/then" analysis:

- "If OPEC maintains current production targets, and UK storage stays below seasonal average, here are three scenarios for Q3 procurement costs" — with ranges, not point predictions
- "The EU carbon border adjustment takes effect in October. Here's what it means for gas-to-power economics under current vs proposed pricing"
- Decision framing: "If you're a procurement manager deciding between a fixed Q3 contract and spot exposure, here are the factors to weigh" — with current data, historical parallels, and cited analyst views

This is opinion-adjacent but grounded in data and cited reasoning. The eval agent needs to enforce the boundary: structured analysis with explained reasoning and cited sources is fine; confident directional predictions are not.

### Quality Framework

Every piece of original content must meet:

- **All claims cited** — no unsourced assertions. Every factual claim links to its source.
- **Reasoning explained** — "we think X because of Y and Z" not just "X is likely"
- **Multiple perspectives** — if analysts disagree, show both views
- **Audience-appropriate** — written for the readership segment, not for SEO robots
- **Structured consistently** — news hook → supporting data → expert context → implications for reader

Quality templates can be derived from studying how OilPrice, Argus Media, ICIS, and similar trade publications structure their articles. Not copying content — studying the *pattern*: lead structure, data placement, quote integration, conclusion framing. The eval agent scores against this pattern.

As reasoning models improve, the research depth and analytical quality improve automatically. The pipeline structure stays the same — better models produce better research and writing within the same workflow.

### Engagement Measurement

To know whether original content is working, we need to measure reader behaviour:

- **Read depth** — do readers scroll through the full article or bounce at the headline?
- **Return visits** — do timeline/graph pages get repeat visits? (indicates reference value)
- **Click-through from homepage** — which article framings get clicked? (indicates headline/segment fit)
- **Time on page** — proxy for engagement quality
- **Cross-article navigation** — do readers follow "related" links? (indicates topic interest mapping)

This data feeds back into the system:

- Topics with high engagement get more frequent coverage
- Readership segments with low engagement may need reframing or different angle selection
- Article structures that perform well become reinforced in the writer prompts
- Timelines that get return visits get more frequent data updates

The engagement pipeline is a separate instrumentation project (client-side analytics → data collection → feedback to content planning). It doesn't block the research/writing pipeline but is essential for calibrating quality over time.
