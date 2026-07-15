# relojistas.com — Rebuild Running Notes (dated log)

Append-only log for the relojistas rebuild (HANDOFF Thread A). Companions:
`relojistas_rebuild_manifest.md`, `_plan.md`, `_runbook.md`. Umbrella workstream log
remains `traffic_probe_running_notes(28).md`; this file is the domain-scoped Thread-A log.

---

## 2026-07-15 — last traffic read + rebuild kicked off (manifest/plan/runbook drafted)

**Context.** Operator asked for a final read of relojistas traffic, then to submit the
domain for a complete rebuild onto static hosting including an outbound newsfeed sent via
RSS to the *same* legacy RSS link that is still being hammered, linking out to sources for
now, content + news in Spanish.

**Traffic read (live off the box, 9–14 Jul window, 5.5 days).**
- 405,701 requests; **83.5% 404** (dead vBulletin paths), **15.8% 301** (www/https),
  **only 0.34% 200** (1,367) — of which **89 were the real homepage**, 8 beacon. Still
  ~99.7% bot/crawler noise at the request level.
- **RSS is the asset:** `/external.php?type=RSS2` pulled **749×** (~136/day, steady:
  87/145/172/144/112/89), 31 distinct CF-masked sources, UAs incl. **FeedFetcher-Google**
  (real Google subscriptions), **meta-webindexer**, **Applebot**, DotBot, feed readers —
  **all 404**. Subscribed variants: `forumids=2,4,13,44,58,78,145,288`, per-user `cat/ppuser`.
- **Organic intent now exists:** 9 captured events, **~8 organic**, over 3.5 weeks from
  **ES/CL/MX** (`Omega seamaster` CL, `Certina 919` ES, `Casio shock ga2100` MX, `Aguja 0.18`,
  `president 22`, `Rolen date`, `Omega 285`, `Casio`). This **updates** the day-one
  "human ≈ 0" verdict — small but real demand.
- **Interpretation:** reactivating the feed with live Spanish watch news is the
  highest-leverage move (136 daily 404s → 136 daily subscriber touches). Rebuild justified.

**Capability check (what already exists vs net-new).**
- News **ingest→triage→curate is live** (`content-feed-trigger`→`orchestrator`→`ingester`→
  `triage`→`render_news_section`→`/data/latest-news.json`); sources are pure data rows.
- **VM deploy plumbing has landed** since the June handoff: `resolveGitRepoName`,
  `deploy_config.target='vm'`, `check_backend_unreachable` all present in the Go source.
- **Net-new = outbound RSS.** The pipeline emits JSON, not RSS XML; `feed-publisher` is
  unbuilt. So we add one thin `render_rss_feed` step + a legacy-URL handler.

**Decisions taken (via operator Q&A).**
- **D1 Host = existing CF-proxied box.** Pure object storage can't answer a query-string
  URL (`/external.php?type=RSS2`); the box + nginx can rewrite/serve it. (chosen over
  object-storage + CF-edge migration.)
- **D2 Keep the engine + a clever backend feature.** Operator: "we can still have the engine
  and think of something clever to deliver with it along with the static pages." Proposed
  A+B: (A) dynamic per-category legacy feeds mapping old `forumids`/`cat` → our topic feeds
  (the thing static can't do; maximises subscriber reactivation); (B) search that *answers*
  from curated news while still capturing intent. C (live "búsquedas recientes" wall) held
  to v2 (needs moderation).
- **D3 Manifest/plan first, then build** (operator chose review-before-build).
- **D4** News links out to sources, title+summary only (brief + rights-safe).
- **D5** Spanish throughout; vertical = relojería; framing = watch news portal on forum
  heritage (to confirm).

**Deliverables this session.** `relojistas_rebuild_manifest.md` (full spec),
`relojistas_rebuild_plan.md` (phases P0–P6), `relojistas_rebuild_runbook.md` (exact
commands; §0 traffic read proven), and this log.

**Next.** On approval: P0 CF real-ip re-run → P1 sites row + `news_feed.recommended` →
P2 seed sources + first ingest → P3 chassis build → P4 `render_rss_feed`/`/feed.xml` →
P5 engine legacy handler + search-answers → P6 verify subscriber 404→200 flip.

**Open.** Confirm framing (§ manifest 2); pick clever-engine option(s) (A+B recommended);
verify real RSS source URLs before seeding (prefer api_news); CF real-ip is a P0 prereq.
