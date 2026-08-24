# 384 — a card image lands, is linked correctly, and the listing that renders it is never re-rendered: the card appears only if something unrelated rebuilds the page

Filed 2026-08-24 by the `dartsonline_traffic` lane, from the owner's report that the
dartsonline.com homepage shows cards with no images (screenshot: a ragged grid, 4 of 12 cards
text-only).

**Not a duplicate of `bugs_open/114`, and the distinction is the whole point.** 114 is *"imagery
is deployed and never referenced"* — assets with `entity_type`/`entity_id` NULL, so the
listing-card join can never pick them up, plus a hero mapping that resolves to the site
fallback. **Here the asset is derived, linked and joinable, and the listing still does not show
it** — because nothing tells the listing page to re-render. Different failure, different fix,
same family. Cross-referenced both ways.

## Symptom, measured at the served artefact

`https://dartsonline.com/` `[MEASURED 2026-08-24 18:05Z]` — 12 `article-card` blocks, **4 with
no `<img>` at all**: `barrel-shapes`, `checkout-chart`, `dart-balance`, `dart-points`.

**And every one of those four cards EXISTS on disk:**

```
card-barrel-shapes.jpg    200      card-dart-balance.jpg    200
card-checkout-chart.jpg   200      card-dart-points.jpg     200
card-grip-styles.jpg      200      card-darts-calendar-density.jpg  200
```

So this is not a generation failure and not a deployment failure. The bytes are served; the page
that should reference them is stale.

## Why two of the six DO show, and why that is the tell rather than a contradiction

`grip-styles` and `darts-calendar-density` render their cards correctly today. Their cards landed
on **2026-08-22**, and the listings were re-rendered later that day by an unrelated 34-page
assemble wave. The other four landed **after** that wave and have had no reason to re-render
since.

**So a card reaches its listing only when something unrelated re-renders the listing.** That is a
race, not a mechanism — and it is why this looks intermittent and self-healing rather than broken.

## Mechanism, read from the code (not inferred)

The image-landed chain is:

1. **`flag_page_image_rebuild_action.go`** — a hero lands, so it emits a `needs_page` re-render
   **for the ARTICLE page** (`spec.reason: "image_landed"`, handler `page-build-handler`,
   `itemKey: page_rerender:<page>`) and, in the same transaction, calls
   `emitContentCardDerive` (added for 114, so the derive no longer waits for a sweep).
2. **`derive_card_asset_action.go`** — reads the hero, derives the card, writes the asset with
   `purpose='card'` and commits it through the git adapter.
3. **…and that is the end of the chain.** `derive_card_asset_action.go` contains **no**
   `rerender` / `rebuild` / `needs_rerender` / `flag_rebuild` reference of any kind
   `[MEASURED 2026-08-24: grep over the file returns zero hits]`. Nothing downstream of a landed
   card invalidates the page whose `query.blog_posts` consumer renders it.

**The listing rebuild machinery exists and simply has no caller on this path.**
`rebuild_blog_listing_action.go` is the action; `discovery_checks/check_orphan_pages.go` is its
only trigger, and it keys on **membership** (`orphan_blog_posts` — "a blog post appears in no
listing"), never on **card freshness**. Once a post is in the listing, no check asks whether the
listing's rendering of it is current.

Step 1 re-renders the ARTICLE for exactly this reason — the code knows an image landing must
invalidate a page. It invalidates the wrong one for the card case: the card is displayed by the
listing, not by the article.

## Why nothing caught it

- `check_content_image_missing` converges on the ASSET existing, not on the page referencing it.
  Its own header says *"The article page re-renders with its new hero via the normal image-landed
  flow"* — true for the hero, and silent about the listing.
- `check_orphan_pages` keys on membership, so a listed-but-stale card is invisible to it.
- `render_audit.py` reports `<img>` elements that **failed to load**. A card with no `<img>` at
  all produces nothing to fail — this site measured `broken-img=0` across 23 pages on 08-20
  while the gap was live.
- The `image_url_404` check keys on **unbacked paths**, i.e. a reference with no asset. This is
  the mirror image: an asset with no reference.

Four checks, and the defect falls between all of them because each asks about one side of a
reference that is missing on the other.

## Fix candidates, ordered by what closes the door

1. **Invalidate the listing where the card lands.** In `derive_card_asset_action.go`, after the
   card commits, emit a re-render for the pages whose components consume `query.blog_posts` for
   this site — the same transaction, the same shape as `emitContentCardDerive`'s own precedent in
   `flag_page_image_rebuild`. This makes the stale state unrepresentable: a card cannot land
   without its consumer being told. **Needs the consumer set to be derivable** — `queryresolve`'s
   `blog_posts` source already knows who those pages are, and `rebuild_blog_listing_action.go:82`
   says it deliberately shares that query.
2. **A discovery check for card-freshness**, the mirror of `check_orphan_pages`: a page whose
   listing markup omits a card image for an entity that HAS an active `purpose='card'` asset.
   Catches the class including any future producer, but detects rather than prevents, and this
   estate's `detected` items do not drain on every site.
3. **Widen `check_orphan_pages` from membership to membership-and-currency.** Cheapest to write,
   worst boundary: it makes an orphan check answer a freshness question, and the next reader will
   not expect that.

(1) is the fix; (2) is worth having anyway as the backstop that would have caught this one.

## How to verify a fix

Pick a listed article with no card, let its hero land, and assert **without touching the page**:
`curl` the listing and require the card `<img>` to appear. The disconfirming result is the one to
insist on — a listing re-rendered for an unrelated reason in the same window will show the card
whether or not the fix works, so pin the listing's `deployed_at` before and after and require
that the re-render was caused by the card landing rather than coinciding with it.

## 090 substitution, stated (owner ruling 2026-07-31)

Not filed through the diagnosis loop: `kubectl` has been `Unauthorized` fleet-wide since
~2026-08-24 18:00Z (the 3-day token expiry), so the loop cannot be dispatched. Substituted with
first-hand verification, all of it re-runnable without the cluster: the served homepage markup
(12 cards, 4 imageless, named), the six card URLs returning 200, and the three code files above
read in full — `flag_page_image_rebuild_action.go:194-206`, `derive_card_asset_action.go`
(grep for rebuild terms: zero hits), `check_orphan_pages.go:11,212`. **What I could not check
without the DB**: the `assets` entity link for the four new cards, and whether any listing
re-render is already queued. Both belong in the first cluster-enabled session.

## Relations

- `bugs_open/114` — the same family, the other failure: asset not linked / hero mapping wrong.
  **114's fix (`emitContentCardDerive` at the landing event) is what makes THIS gap reachable**:
  cards now land promptly and correctly, so the stale listing is the remaining hop.
- `bugs_open/083` — detected findings never reach a handler (why candidate 2 alone is not enough).
- `LANDMINES.md` — "a stale PAGE holds every improvement since it rendered".
