# CONTRIB 2026-08-25 — from `apis_uk_bees_homepage` (session "apis.uk"): owner ruling on the split — you take everything Google, we keep the rest

Owner, 2026-08-25, in the apis.uk session, on reading our handoff's §4 (traffic and tracking):

> "section 4 has google in it which is taken by another lane, please communicate to that lane that
> that is what they take and we will take the rest here."

So this is the division of labour, stated once so neither lane re-derives it. Your
`CONTRIB_2026-08-25_from_analytics_gtm_…` was read in full; §2 is accepted — our §3 "per-site
analytics id in `RenderFallbackHead`" build is **dropped**, and our handoff now says so visibly.

## Yours — everything Google, including the parts our handoff still describes

| item | where it was written in our lane | state as we hand it over |
|---|---|---|
| **GA4 publication** + the consent decision that precedes it | our `HANDOFF_2026-08-25` §2.1 and the §4a walkthrough (owner's verbatim request) | still `"tags":[]` on `GTM-PQ3WCTBD` version 2, re-read 2026-08-25 ~17:00 BST. Copy the walkthrough into your lane if you want to maintain it; we will not edit it further |
| **The durable-tag rebuild** (`bugs_open/397`, `sql/c2_…`) | our 08-24 backfill caused it | owner-timed, yours |
| **The structural half** — new sites born untagged; how a handed-over site is guaranteed never to get our id (397 §6.2, council scope) | our §3 bullet 1 was a wrong-place design for this | yours; the `sites.settings->>'analytics_container_id'` key we proposed has **0 rows** and should stay that way — one seam, STY-050 |
| **Search Console** (service account, 039 §5) | our §2.2 | yours |
| **The fleet traffic dashboard script** — 039 §2's Cloudflare query batched 8 zones at a time, our own `curl`/headless traffic as its own visible line | our §2.3, "offered, not built" | yours to build or decline; it was never started here |
| `039_REFERENCE_traffic_and_tracking.md` (at `docs024_key_docs_latest/039_…`, not in either lane dir) | written from our lane on 08-25 | yours to keep current; we will only append dated corrections |

## Ours — what stays in `apis_uk_bees_homepage`

- **The apis.uk page itself**: 6 illustrated sections, 7 `page_components` `lock_type='permanent'`,
  `pages.build_status='deployed'` `[MEASURED 2026-08-25 ~17:00 BST]`, served 200 / 67,877 B, GTM
  snippet present once, `tools.apis.uk` real endpoint 200.
- **Per-section subjects** — `pages.sections` is `[]string` (`PlannedSections`), so every slot gets
  one brief; council scope, ours to build.
- **Image accuracy A + C** — front-loaded distinguishing features in the imagery prompt composition,
  and a vision-critique step (`execute_vision_prompt`) on `visual-design-auditor`; ours.
- The two `deferred` `content_rewrite` items on apis.uk (swarm, pollination), unblocked by the
  subjects build.

## Two things you should know about apis.uk before `c2` fires

1. **apis.uk is on your bucket-B list, and its index page refuses re-renders.** The permanent locks
   are on `page_components`, not the head slot (your §1 is right about that) — but a page-level
   re-render is refused by them: today's `page_rerender` from the 383 lane on apis.uk is `failed`,
   and their own commit `9a843c06a` reads "apis.uk cannot re-render". So when `c2`'s `stale_chrome`
   wave reaches apis.uk, expect the **page** item to fail with an overwrite refusal `[INFERRED from
   the 08-25 handoff's measured "overwrite: REFUSED for page index" — not re-tested this session]`.
   That is not damage: the served page keeps the tag it already carries, and the head **artefact**
   will be right once the key exists. Please list it as expected in 397 §9, and tell us when `c2`
   has run — we will verify apis.uk at the served bytes ourselves rather than have your wave
   report it as a failure nobody owns.
2. 397 §9 names the lanes to tell (loanzy, webdesign, 357/384, agritec). **Add
   `apis_uk_bees_homepage`** — apis.uk is in the wave, and the trap this lane paid for is that a
   render re-queues the page (`build_status='needs_rebuild'` is queue membership), so we will need
   to settle it again afterwards.

Nothing here needs a reply. If any row in the "yours" table is one you would rather not carry, say
so in our directory and we will take it back — the owner's ruling is about Google, not about
dumping work.
