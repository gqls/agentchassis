# 410 — `next_fetch_at` is stamped at FETCH time (`NOW() + fetch_interval`), so a 6-hour interval on a 6-hour trigger falls due SECONDS after the next pass fires: every news site whose sources are all 6-hourly is served every OTHER run — a 12-hour cadence wearing a 6-hour label

**Filed 2026-08-26 by the `idea_uk_vm_site` lane** (`docs/agent_docs/docs024_key_docs_latest/idea_uk_vm_site/RUNNING_NOTES_idea_uk_vm_site.md` §X.65).
Found because one site's `content_sources.last_fetched_at` did not move across a trigger pass the
handoff had recorded as "COMPLETED".

**Diagnosed first-hand rather than via the `090` loop — declaring the substitute per the 2026-07-31
ruling.** Every link below is quoted from the live system (the live `agent_definitions` step query,
the `scheduled_tasks` row, `content_sources` timestamps, `orchestration_states` runs) or from the Go
at a cited line; the claim is closed-form arithmetic on those; a **positive control** (sites with a
sub-6h source) behaves as the arithmetic predicts; and a **prospective prediction** is recorded in §5
BEFORE the pass that tests it. A `090` run is welcome and would be cheap — the fixing lane should fire
it if any of §2 reads as inference.

**This is NOT `bugs_open/316`.** 316 is *who wins when more sites are due than the cap admits*
(alphabetical ordering; fixed by migration `556`, live). This is *nobody being due*: the cap was not
binding in the measured pass (10 due, 10 dispatched, `LIMIT 10`). 316's fix is in the query quoted
below and this defect survives it. The `bugfix_316_news_feed_ordering` lane's README (line 138,
2026-08-22) says the seven 6-hourly sites are *"now fully served"*: **refuted by §3** — they are
served every other run, exactly as before, for a different reason.

---

## 1. Mechanism in one paragraph

`scheduled_tasks.content-feed-refresh` fires `content-feed-trigger` every 21,600 s. Its
`find_news_sites` step selects a site only if some active source has `next_fetch_at IS NULL OR
next_fetch_at <= NOW()`. Both writers of `next_fetch_at` stamp it **relative to the moment of the
fetch**, not the moment of the trigger: `NOW() + fetch_interval`. A fetch happens 10 s – 9 min after
its trigger (dispatch is sequential, ~50 s per site). With `fetch_interval` = **6 h** (the column
default, and every source on 12 of 12 news sites), a source fetched at *T + δ* is due at *T + 6h + δ*,
and the next trigger fires at *T + 6h + ε* with ε ≈ 3–30 s of scheduler drift. δ > ε for every site
but (at best) the first one dispatched in the first seconds, so **the source is not due, the site is
skipped, and it is served at T + 12h.** Sites that ALSO hold a 3 h or 4 h source are due at every
pass and are served every pass — the control.

## 2. Evidence `[ALL MEASURED 2026-08-26 14:15–14:35Z unless dated]`

**The trigger's live site-selection query** (`agent_definitions` type `content-feed-trigger`,
`default_config->workflow->steps->find_news_sites->config->query`, post-316):

```sql
... EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true
            AND (cs.next_fetch_at IS NULL OR cs.next_fetch_at <= NOW()))
... ORDER BY due_at ASC NULLS LAST, domain ASC LIMIT 10
```

**The two stamp arms**, both `NOW() + fetch_interval`:

- `platform/orchestration/actions/dispatch_feed_sources_action.go:272-279` — *"Optimistically update
  next_fetch_at to prevent re-dispatch before completion"*: `SET next_fetch_at = NOW() + fetch_interval`.
- `platform/orchestration/actions/feed_actions.go` `UpdateSourceTimestampsAction` success arm (the
  `UPDATE content_sources SET last_fetched_at = NOW(), … next_fetch_at = NOW() + fetch_interval`
  block; the failure arm multiplies the interval by `LEAST(error_count+1, 4)` — same anchor).
- The per-site source selector the orchestrator itself uses is the same shape:
  `feed_actions.go:962` *"Returns all active sources for a site where next_fetch_at <= NOW()"*, :1007.

**The cadence and its drift** (`orchestration_states`, `owner_agent_type='content-feed-trigger'`):

| fired (UTC) | gap |
|---|---|
| 08-25 14:45:19 | |
| 08-25 20:45:32 | 6h 00m 12s |
| 08-26 02:45:35 | 6h 00m 04s |
| 08-26 08:46:06 | 6h 00m 30s |

**The interval**: `information_schema.columns` default for `content_sources.fetch_interval` =
`'06:00:00'`; every active source on all 12 news-eligible sites carries `06:00:00`; dartsonline adds a
`04:00:00`, relojistas a `03:00:00` and a `04:00:00` (`SELECT domain, string_agg(DISTINCT
fetch_interval) … GROUP BY domain`).

**The worked case — idea.uk at the 08:46:06 pass.** Its five sources were fetched 02:46:15–02:46:31
(dispatched 02:45:52, 17 s after the 02:45:35 trigger); `next_fetch_at` = **08:46:15 – 08:46:31**;
the trigger fired at **08:46:06** — **9 to 25 seconds too early**. `orchestration_states` holds ONE
`content-feed-orchestrator` run for the site today (02:45:52). Not the cap: exactly 10 sites were due
(the seven fetched at 20:45–21:02 on 08-25 and skipped at 02:45 for this same reason; the two
always-due sites; one newly armed site) and exactly 10 were dispatched, 08:46:20 → 08:54:23.

## 3. The 48-hour census — every 6h-only site runs 12 h apart; the sub-6h sites run every pass

`SELECT s.domain, string_agg(DISTINCT cs.fetch_interval::text,','), string_agg(to_char(o.created_at,'DD HH24:MI'),' | ' ORDER BY o.created_at) FROM orchestration_states o JOIN sites s ON s.id=o.site_id LEFT JOIN content_sources cs ON cs.site_id=s.id AND cs.is_active WHERE o.owner_agent_type='content-feed-orchestrator' AND o.created_at > now()-interval '48 hours' GROUP BY 1;` (distinct times, de-duplicated by hand):

| site | intervals | orchestrator runs (UTC) | effective cadence |
|---|---|---|---|
| ai-agent-orchestration.com | 6h | 25 20:47 · 26 08:47 | **12 h** |
| fundamentallyai.com | 6h | 25 20:49 · 26 08:48 | **12 h** |
| gaswholesalers.com | 6h | 25 20:58 · 26 08:51 | **12 h** |
| mortgagecalculator.co.uk | 6h | 25 21:00 · 26 08:52 | **12 h** |
| robot-hands.com | 6h | 25 20:51 · 26 08:48 | **12 h** |
| webdesign.co.uk | 6h | 25 20:45 · 26 08:46 | **12 h** |
| idea.uk | 6h | 25 16:25/16:29 (manual) · 26 02:45 — skipped 20:45 and 08:46 | **12 h** |
| remortgagecalculator.uk | 6h | 26 02:48 · 26 13:43 (off-cadence run) | — |
| loanandmortgagecalculator.co.uk | 6h | 26 08:54 (first run) | — |
| vetcomparison.uk | 6h | 25 21:02 · 25 23:00 · 26 08:53 | mixed |
| **dartsonline.com** | 4h + 6h | 25 14:47 · 20:56 · 26 02:47 · 08:50 | **6 h** ← control |
| **relojistas.com** | 3h + 4h + 6h | 25 14:45 · 20:54 · 26 02:46 · 08:49 | **6 h** ← control |

The disconfirming result would have been a 6h-only site appearing at ~02:4x AND ~08:4x; none does.
The control would have failed if the 3h/4h sites also alternated; they do not.

## 4. Blast radius and why nothing surfaced it

- **10 of 12 news sites** (as of 2026-08-26) hold only 6-hourly sources and refresh **twice a day, not
  four times** — half the designed and documented cadence (`scheduled_tasks.description`: *"every 6
  hours to refresh news feeds"*), since the trigger was armed.
- Nothing fails, so nothing files: the 6-hourly `stale_news_section` check keys on newest-item age vs
  `max_age_hours` (default **72**), which a 12 h cadence never approaches; the trigger's own run is
  COMPLETED; every orchestrator run that does happen is COMPLETED.
- The 316 lane measured "late" against each site's `fetch_interval` and read the cap boundary as the
  cause. Their fix made the queue fair (true) and did not change the cadence (this file).
- Aside, not the defect: dispatch is sequential at ~50 s per site (08:46:20 → 08:54:23 for 10), which
  is what makes δ minutes rather than seconds.

## 5. Prospective test — written at 14:38Z, BEFORE the ~14:46Z pass

idea.uk has been due since 08:46:15. Prediction: **(a)** the ~14:46 pass dispatches it (an
`orchestration_states` row, `owner_agent_type='content-feed-orchestrator'`, `site_id` idea.uk,
`created_at` ≈ 14:46–14:55); **(b)** its five `next_fetch_at` land at (fetch time + 6 h) ≈
20:46–20:55; **(c)** the ~20:46 pass does NOT dispatch it; **(d)** the ~02:46 pass on 08-27 does.
Refutation of (c) — an idea.uk orchestrator row at ~20:46–20:55 — kills this file's mechanism.
Outcome to be recorded below by whoever is watching (the filing lane will record (a) and (b)).

## 6. Fix candidates — ordered by what makes the bad state unrepresentable

1. **Give the due predicate a look-ahead of half the cadence, in BOTH layers, in one migration.**
   `next_fetch_at <= NOW() + interval '3 hours'` in `find_news_sites` (config, DB-live on apply) AND
   in the orchestrator's per-site source selector (`feed_actions.go:1007` — Go, so this half needs a
   roll; until it rolls the trigger would dispatch a site whose orchestrator then finds 0 due sources
   and completes as `no_sources`). A source due at any point before the midpoint to the next pass is
   served now; worst-case lateness falls from 6 h to ≤ 3 h and the phase lock cannot form. **The two
   layers are the trap** — 316 §"two layers" already documents that the site-level query and the
   source-level query drift apart.
2. **Anchor the stamp to the schedule, not the fetch**: `next_fetch_at = <this pass's trigger time> +
   fetch_interval`, carried into the ingester's payload. Structurally closes the door for any
   interval ≥ cadence, but the ingester has no notion of the trigger time today (Go, two arms, roll).
3. **Set `fetch_interval` below the cadence** — column default and seeder to `05:30:00`, and an
   `UPDATE` for the 12 existing sites. Cheapest, DB-live, and the least closing: it re-opens the moment
   anyone sets an interval equal to the cadence again, which is the obvious value to set.

**Per-site mitigation (candidate 3, one site) — left for the OWNER**: the filing session's attempt
was refused by its permission classifier (a production `UPDATE`), which is the right refusal — it
doubles that site's search-fetch spend and the owner has a cost-scare history on rotations.
```sql
CREATE TABLE bak_ideauk_fetch_interval_20260826 AS SELECT * FROM content_sources WHERE site_id='1244516d-014d-421c-88c6-090bb1e9552a';
UPDATE content_sources SET fetch_interval='05:30:00' WHERE site_id='1244516d-014d-421c-88c6-090bb1e9552a' AND is_active AND fetch_interval='06:00:00';
```

## 7. How to verify a fix

The §3 census, re-run 48 h after the roll: every 6h-only site shows **four** distinct run hours per
day (≈ 02:4x, 08:4x, 14:4x, 20:4x). Per site, `max(last_fetched_at)` should never be > 6 h 15 m old
while the trigger is running. Control stays: dartsonline/relojistas unchanged at every pass.

## 8. Consumers told / pointers

`bugs_open/316` (CONTRIB appended today, same commit) · `bugfix_316_news_feed_ordering/` (its
"fully served" claim, README:138) · `news_feed_pooling/` · `seed_content_sources_action.go` (inherits
the 6 h column default) · 016b §9 entry "a due-stamp of fetch-time + period is phase-locked to the
scheduler" (same commit) · idea.uk lane RUNBOOK 6g.

### §5 outcomes, recorded as they land

- **(a) CONFIRMED 2026-08-26 14:47Z** — trigger fired **14:46:32** (gap from 08:46:06 = 6h 00m 26s,
  the drift as measured); idea.uk `content-feed-orchestrator` run created **14:46:58**, the site's
  second of the day.
- **(b) CONFIRMED 2026-08-26 ~14:48Z** — the five `next_fetch_at` moved to **20:47:01–20:47:06**
  (the dispatch arm's optimistic stamp) and then **20:47:24–20:47:42** (the ingestion arm's), both
  AFTER the next trigger's expected ~20:46:58–20:47:02. Margin this cycle: **~25–45 s** (morning
  cycle was 9–25 s) — the phase lock re-arms itself every pass, δ > ε again.
- **(c)/(d) pending** — the ~20:46–20:47 pass should SKIP idea.uk (an idea.uk orchestrator row
  there refutes this file); the ~02:46 pass on 08-27 should serve it.
