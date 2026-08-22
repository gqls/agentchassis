# NOTES — bugs_open/316, the news-feed cap serves the alphabet

Append-only, newest at the bottom. Technical log: evidence, commands, what the system said,
and every misstep.

---

## 2026-08-22 — session start: is the bug still valid?

**Ownership checked first.** `scripts/who-owns.py 316` returns OWNED-or-recently-active and names
`bugfix_275_silent_row_caps` as the likely owner. Read that lane's cold start — it is **CLOSED as of
2026-08-19 10:30Z** and its handoff lists 316 explicitly as one of *"three tickets that stand on their
own — they are not this lane and nobody owns them"*, suggested order **third**. So: not owned, free to
take. `git log` on the bug file shows one commit (`a996dbd73`, the filing) and nothing since.

**The defect is still live.** [MEASURED 2026-08-22 09:52Z, live `clients_db`]

```
SELECT jsonb_pretty(default_config->'workflow'->'steps'->'find_news_sites')
FROM agent_definitions WHERE type='content-feed-trigger'
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

still ends `... ORDER BY s.domain LIMIT 5`. Verbatim text saved to
`PREFIX_find_news_sites_query_2026-08-22.sql` in this directory — it is the fixture the detector's
positive control will be built from, and once the migration lands the live row can no longer supply it.

⚠ **`agent_definitions.updated_at` for this row reads 2026-08-22 08:36:05Z — today.** That is NOT
another session fixing this: the query text is byte-identical to the one the bug file quotes on 08-19,
and there are **zero snapshot rows** for `content-feed-trigger`. Recorded because a fresh `updated_at`
on a shared tree is exactly the shape that should make you look before assuming your target is untouched.

## The bug has WORSENED, and the new figure is much sharper than the filed one

Census of the retained `orchestration_states` window (~2 days), by the method `bugs_open/275` established
(`collected_data`, not the logs — the chassis log retains 15-90 s):

| run (UTC) | n | domains returned |
|---|---|---|
| 08-22 08:38 | 5/5 | dartsonline, gaswholesalers, mortgagecalculator, relojistas, vetcomparison |
| 08-22 02:38 | 5/5 | ai-agent-orchestration, dartsonline, fundamentallyai, relojistas, robot-hands |
| 08-21 20:37 | 5/5 | dartsonline, gaswholesalers, mortgagecalculator, relojistas, vetcomparison |
| 08-21 14:37 | 5/5 | ai-agent-orchestration, dartsonline, fundamentallyai, relojistas, robot-hands |
| 08-21 08:37 | 5/5 | ai-agent-orchestration, dartsonline, fundamentallyai, gaswholesalers, mortgagecalculator |

**Every run at the cap, 5 of 5** — the bug file's central claim, reproduced on a fresh window.

**`webdesign.co.uk` appears in ZERO of the five.** It is alphabetically LAST of the nine eligible sites.
It has been continuously due since 2026-08-21 08:45Z and was therefore eligible at four consecutive runs
(14:37, 20:37, 02:38, 08:38) and selected at none of them. Last fetched **2026-08-21 02:45Z** — over 31
hours on a **6-hour** cadence, i.e. **419% of its own cycle**. The filed file recorded this same site at
**7%**. The starvation is not stable; it compounds.

`relojistas.com` (the file's sentinel, 3 h cadence) is currently **served** — it appears in 4 of 5 runs.
Its alphabetical rank is 6, so it sits just past the old cut and wins whenever fewer than five
alphabetically-earlier sites are due. That is the same mechanism, not a contradiction: the alphabet does
not decide who is late in every window, it decides who is late **whenever contention happens**, and the
site it hurts most is whichever one is furthest down the alphabet at that moment. Today that is
`webdesign.co.uk`, decisively.

## Fleet census: how big is the class?

Over every live, active, non-snapshot agent definition, every `query_database` step whose query carries a
`LIMIT`, extracting the trailing `ORDER BY … LIMIT n`. [MEASURED 2026-08-22]

Findings, classified:

- ~19 steps are `LIMIT 1` fetch-one / claim-one, correctly ordered (`created_at DESC`, or
  `priority ASC, created_at ASC FOR UPDATE SKIP LOCKED`). Not this class — LCO-009 excludes them for the
  same reason.
- `build-pipeline-trigger.find_dispatchable_site` — `wi.created_at ASC, wi.priority ASC, wi.id ASC LIMIT 1`.
  Correct FIFO.
- `visual-design-auditor.load_design_context` (LIMIT 5) and `fix-proposer.load_last_bundle` (LIMIT 2) —
  the LIMIT is **inside a subquery**; the outer result is one row. Correctly ignored, as LCO-009 already
  records.
- `meta-description-backfiller.load_pages_missing_meta` — `ORDER BY p.name LIMIT 25`. Alphabetical **and**
  capped, so it looks like the same defect — but its candidate set is **CONSUMED**: a page that gets a
  meta description leaves `WHERE COALESCE(p.meta_description,'') = ''`. Alphabetical order there is a
  batching delay, not starvation. **This near-miss is the useful one** — it is what forced the
  distinction below.
- `model-directory-trigger.find_directory_sites` — `ORDER BY random() LIMIT 12`. Recurring set, but
  `random()` does not systematically starve, and it returns 3-4 against a cap of 12 so the cap never
  binds. This is the bug file's negative control and it holds up.
- **`content-feed-trigger.find_news_sites` — `ORDER BY s.domain LIMIT 5`. The only live member of the
  dangerous shape.**

**The distinction the census forced** (and it narrows register LCO-009's *"coverage is eventual, not a
defect"* gloss more precisely than the bug file did):

> A cap on a query whose candidate set is **replenished by the clock** — rows never leave the set, they
> merely acquire a later due-time — starves the tail **permanently** under a static `ORDER BY`.
> A cap on a query whose candidates are **consumed** — a row leaves the set once served — is only a
> batching delay, and there "coverage is eventual" is true.

`find_news_sites` is the first kind. `load_pages_missing_meta` is the second. The count alone cannot tell
them apart, which is why LCO-009's WARN cannot.

## The platform already has the right idiom — one layer down

Not a new convention. Two Go call sites already order feed work by its due-time:

- `platform/orchestration/actions/dispatch_feed_sources_action.go:101` — `ORDER BY next_fetch_at ASC NULLS FIRST LIMIT $n`
- `platform/orchestration/actions/feed_actions.go:1016` — `ORDER BY next_fetch_at ASC NULLS FIRST`

Both select **sources within a site**. The **site**-selection query, which lives in config rather than in
Go, is the single layer that skipped it. That materially changes the fix's argument: it applies an
existing platform convention to the one place that missed it, rather than proposing a new one.

## A trap in the bug file's own fix candidate 1

The file proposes `ORDER BY min_next_fetch_at NULLS FIRST`. Taken literally that is **wrong**, and it
would be worse than the alphabet.

The eligibility predicate has two arms:

```
(NOT EXISTS (active content_sources for this site)   -- arm A
 OR EXISTS (an active source with next_fetch_at IS NULL OR <= NOW()))   -- arm B
```

A site matching **arm A** has no active sources at all, so `min(next_fetch_at)` is NULL **and it is
permanently eligible** — nothing a fetch does can advance a timestamp it does not have. Under
`NULLS FIRST` such a site would win **every run for ever**: a head-of-queue squatter, which starves
everyone deterministically instead of merely starving the alphabetical tail.

[MEASURED 2026-08-22] Zero sites are in that state today — all nine eligible sites carry 1-9 active
sources — so this is a latent trap, not a live one. The arm is in the query deliberately, so the fix must
answer it rather than delete it.

Second, smaller trap in the same expression: a source with `next_fetch_at IS NULL` has **never been
fetched** and is genuinely maximally overdue, but SQL `min()` skips NULLs, so a bare
`min(cs.next_fetch_at)` hides it behind its siblings' timestamps.

## A thing I checked rather than asserted

The repo seed `docs/agent_docs/sql_for_agents/090_b_content_feed_trigger.sql` carries
`WHERE s.deleted_at IS NULL`; the live query does not. That looks like drift worth restoring — it is not.
`sites.deleted_at` **does not exist** (`ERROR: column s.deleted_at does not exist`). The live query is
correct and the seed is stale. Recorded because "live config dropped a guard the seed has" is a
convincing-looking finding that took one query to refute, and the memory note is right: the seed is
history, the live row is fact.
