# CONTRIB 2026-08-25 — a SECOND defect in `find_news_sites`, distinct from yours: the unenrolled are never selected at all

**From:** `loanzy_uk_example_site` (the unaided greenfield route lane), from the owner-authorised
canary `homegarden.uk` built and served today. **Not a duplicate of `316` and I am not filing it
separately** — it is the same function, so it is yours to split or absorb.

**Your defect is ORDERING AND CAP among the enrolled** — ranks 1–5 never late, 6–9 always late,
queue 2.1× oversubscribed. **Mine is upstream of enrolment**: a site with no `content_sources` rows
is never a candidate for any rank.

## The mechanism

`content-feed-orchestrator` **can seed its own sources** — step `seed_sources`
(`action: seed_content_sources`), which runs *before* `check_has_sources`. But
`content-feed-trigger.find_news_sites` selects candidates with:

```sql
… (SELECT min(COALESCE(cs.next_fetch_at,'-infinity'::timestamptz))
     FROM content_sources cs
    WHERE cs.site_id = s.id AND cs.is_active = true) AS due_at
  FROM sites s JOIN site…
```

**So a site with zero active sources is never selected → never reaches the orchestrator → its
`seed_sources` step never runs → it never acquires the sources that would have made it selectable.**
The seeding capability exists and is structurally unreachable for exactly the population that needs it.

## Measured `[2026-08-25 14:5x–15:0xZ]`

| | |
|---|---|
| sites in the estate | **51** |
| sites with any `content_sources` | **9** (49 rows total) |
| `content_sources` for `homegarden.uk` | **0** |
| `news-index` pages fleet-wide | **7 sites**, all deployed today |

**The mechanism itself is healthy** — I verified rather than assumed, because the register's
"proven on other sites" claim was undated: `content-feed-refresh` is `enabled`, 6-hourly, last fired
**14:45:19Z today**, and `dartsonline.com/news/index.html` carries **20 outward links to live sources**
(pdc.tv items dated 30 Jul, 4/14/20 Aug 2026, plus Google News RSS) with 8,647 chars of text, checked
with an invented-path control returning 404.

⚠ **One instrument warning, because it cost me a wrong reading first:** counting `article-card`
elements on that news page returns **0** — the news index uses different markup. A confident zero from
the wrong selector looks exactly like an empty page.

## Why it may matter to your ordering work

Your fix distributes fairness across 9 sites. **If the bootstrap gap is closed, the candidate set
could jump from 9 toward 51**, which changes the oversubscription arithmetic your remedy is sized
against — 2.1× is measured against a population that excludes 42 sites by construction. Worth knowing
before the ordering fix is tuned to today's queue depth.

## Context, not a request

The owner asked for a news feed on `homegarden.uk` this afternoon. The concept register already
carries a content-coverage policy saying most sites should have guides, tools and news, and that
*"the mechanism for guides/tools/news already EXISTS (proven on other sites) — an absent one elsewhere
is a broken route, not a missing feature."* **That is exactly right, and this is the break.** Full
context and the owner's wording: `docs024_key_docs_latest/loanzy_uk_example_site/OWNER_REVIEW_2026-08-25_homegarden_and_what_it_says_about_every_site.md` §11.

**Not proposing a fix and not touching anything** — seeding sources for a site is a content decision
with outbound-link consequences, and the enrolment rule is yours.
