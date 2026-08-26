# HANDOFF 2026-08-25b — the news feed is LIVE (the gap was a missing spec key, never a lost dispatch). Three failed rows read and routed. Class B dropped with a reason. Owner's queue: 23 of 49 are one cause.

> **SUPERSEDED 2026-08-26 by `HANDOFF_2026-08-26_continue_here.md`** — the news arc is proven unattended (3 trigger passes), the fresh build v1.0.1341 is verified at the binary, and the live headline is a FLEET LLM OUTAGE (Anthropic credits, `bugs_open/243` recurrence). Read that file first.

**Supersedes `HANDOFF_2026-08-25_continue_here.md` as the cold-start file** (its §4 item 1 is
done; items 2–4 are answered below). `HANDOFF_2026-08-16` §4 + §7 still hold the voice-arc
history and the head blind spot; `HANDOFF_2026-08-11` §3 for RFC_015.

Cold-start order: **this file → RUNBOOK Phase 6 → NOTES §X.61 → `README_where_we_are.md` tail**.

---

## 1. News — LIVE, verified at the artefact 2026-08-25 16:32Z

- `https://idea.uk/data/latest-news.json` **200**, `item_count 6`, `updated_at 2026-08-25T16:31:01Z`,
  `insights_url /news/index.html`; `news-archive.json` **200**. vm-sites commits `c1ca7e54`
  (16:26:11Z) and `b7c8efaf` (16:31:05Z), both `idea.uk/data/*.json`.
- `content_sources` idea.uk **5** (`news_search`, `error_count 0`); `content_feed_items` **12**
  (9 `relevant` avg 58.9, 3 `review`).
- **Mechanism, one line:** both news mechanisms select a site on
  `classification.content_features.news_feed.recommended = true`; idea.uk's spec had no
  `content_features` key, and the step that writes it cannot reach this site (never reads
  `industry_tags`; writes nothing on no-match; its carrier `improvement-sweep` is disabled).
  Fleet: 9/9 flagged sites have sources, 22/22 unflagged have none `[MEASURED 2026-08-25]`.
  Full account: NOTES §X.61 §2, the seed's header, LANDMINES (new entry, `site_specs` footprint).
- **What was done:** `sql/SQL_2026-08-25_arm_news_feed.sql` (authored spec row, supersede +
  insert; news page re-typed `news-index` on BOTH `pages` and the current plan); two receipted
  runs of `content-feed-orchestrator` via `scripts/dispatch_content_feed_orchestrator.sh`.
  Everything else — sources, fetch, triage, render, commit, sync — was the framework.
- **Decisions the owner may reverse (README has the plain version):** `news_search` only, no
  `api_news` (LLM-authored) — one word + a re-run to add; the five keywords are this session's.
- **From here the framework owns it:** `content-feed-refresh` runs 6-hourly (next ≈ 20:45Z
  2026-08-25). Watch, don't push. RUNBOOK 6d is the verification; 6e the retune recipe.

## 2. WHAT'S NEXT — in value order

1. **Retune the five keywords after a week of items** (RUNBOOK 6e; webdesign.co.uk needed to).
   The goto-redirect question is ANSWERED (16:45Z): 3 of the 6 served items ARE
   `google.com/goto` links, mortgagecalculator serves 1 too — **filed as `bugs_open/400`**
   (the ingester's ScrapingBee news provider; dedup keys on `source_url` so one story can
   double-list). The ingester lane's to fix; nothing owed here beyond watching.
2. **A/B test calculator page — owner's choice, not made:** rebuild via the tool writer
   (`create_tool_component`, as the 311 lane did for webdesign.co.uk on 08-19) or retire the
   page. DB holds 1 of 4 planned sections; served page works from an old deploy; every rerender
   fails. Three stale `tool_crosslink` `content_rewrite` items go with it.
3. **Empty tool headings** (funding-fit, patent-check): `{{.section_heading}}`/`{{.section_intro}}`/
   `{{.eyebrow_label}}` unfilled — real, reproduced every rebuild. Belongs to whoever owns tool
   templates (`bugs_open/357` nearest); the two `failed` `empty_section` rows are `bugs_open/300`'s
   class (CONTRIB appended) and will keep failing closed until 300's fix lands.
4. **Owner's review queue (49):** 12 `decision_blocked_change` + 11 `lock_blocked_change` = the
   guard refusing rebuilds of D-001/D-002/D-004/D-005-protected and locked sections — "keep
   stored" is already what happened. Recommend a batch close **with the owner**. Then 6
   `cta_names_unknown_destination` (08-24), 4 `content_rewrite`, 4 `dead_control` (07-18), singles.
5. Older residuals, unchanged: first organic signed Stripe webhook; tools-page card images and
   tool-page heroes; empty-kind → SDXL image-routing hole; ingress landmines (`ufw allow 80,443` FIRST).

## 3. Closed this session, so stop carrying them

- **Class B** (8 components / 3 sites, `content_data` NULL): DROPPED, not filed. Mechanism is
  `bugs_closed/194` (fixed, live); the copy is the honesty arc the owner closed 08-17; the
  field-edit gap is `bugs_open/357`'s. Live shape `[MEASURED 2026-08-25]`: 57 NULL-`content_data`
  deployed components / 21 sites; 11 / 5 sites carry visible "honest*". NOTES §X.61 §5.
- **The "08-04/05 dispatch rows are GONE" mystery** stays `[UNVERIFIED]` and is now irrelevant to
  news: even a landed dispatch would have exited at `complete_no_sources` on the same missing key.

## 4. Queue state `[MEASURED 2026-08-25 ~16:10Z]`

49 `needs_human_review` (§2.4) · 33 deferred (16 contrast, 12 `undeployed_asset`, 3
`capability_gap`) · 31 detected (`head_essentials_missing` family, `bugs_open/083`'s lane) ·
3 failed (all read, §2.2–2.3) · 0 triaged/approved/claimed. `complete` totals are a rolling
window — never compare across sessions.

## 5. Standing rulings — unchanged

Honesty arc CLOSED (owner 08-17, mig 454); D-005 guards the report hero; `whether you're` stays;
the voice gate cannot see `pages.title`/`meta_description` (08-16 §4).

## 6. Traps carried forward (+3 new)

- **`dispatch_sources` is ASYNC** — the same run's triage and render read what has landed. One
  orchestrator run is not a verification; run it twice or wait for the trigger's next pass.
- **Identical `length(rendered_html)` across a rebuild means the rebuild REPRODUCED, not
  repaired.** Re-run the finding's predicate on the replacement row before calling it stale
  (WRONG_CALLS 2026-08-25).
- **`news_render_result.item_count` is not the snippet's length** — read the served file.
- The 08-25 set still stands: rolling window; `attempt_count = 0` = never tried; a completion
  count is not an artefact check; the 08-16 §7 set.

## 7. Pointers out of this lane

- Queue timing → `dispatch_throughput/` (N=2 live, D0a–D16). CSS/theme → `bugs_closed/198`.
  Component-id churn on findings → `bugs_open/300`. Tool templates/rebuilds → `bugs_open/357`,
  `bugfix_311_component_keys/`. News machinery fleet-wide → `bugs_open/316` (cap/ordering),
  `news_editorial_features/` (the headline/subheadline the snippet carries is theirs).
