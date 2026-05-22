There's the news pipeline, named explicitly in the agent list:

content-feed-orchestrator — the top-level news/feed orchestrator
content-feed-trigger — schedules/kicks off feed runs
feed-ingester — fetches items (the rss/news_search/api_news populate step) → writes content_feed_items
feed-triage — relevance/dedup/credibility scoring (the relevance_score, credibility, status columns)

So the populate path is content-feed-trigger → content-feed-orchestrator → feed-ingester (fetch+store) → feed-triage (score/filter) → render_news_section (→ /data/*.json). That's the well-trodden path you mentioned. And the content_feed_items schema is rich — it already has the bones for your enhancements:

Your price-filter + expiry enhancement is partly pre-built. There's an expires_at column already, and topics jsonb, relevance_score, credibility. So expiry is a supported concept; the work is (a) populating expires_at short for price-movement items, and (b) a triage rule that tags price items via topics. The 1-2-day TTL you want maps directly onto expires_at.
source_published_at, published_at, status (ingested→relevant→published/duplicate/expired/rejected) — a full lifecycle. The idx_cfi_dedup excludes duplicate/expired/rejected, so expired items auto-drop from feeds.
