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

---

## ⚠ CORRECTED 2026-08-24, same session, ~1 hour later — the mechanism above is WRONG in its central claim, and the fix proved it

**What I filed:** *"nothing re-renders the listing when a card lands"*, and *"the two that show do
so because an unrelated 34-page assemble wave re-rendered the listings after their cards landed"*.

**Both are false, and my own timeline refutes them.** Measured after filing:

```
listing re-renders since the four cards landed (all page_rerender on `index`):
  2026-08-23 11:41:45   (assemble: no reason)   complete
  2026-08-23 14:27:24   (assemble: no reason)   complete
  2026-08-24 14:59:43   (assemble: no reason)   complete
```

The listing was re-rendered **three times** after the cards landed and still showed nothing. So
"nothing re-renders the listing" is wrong, and the coincidence I built the "why two show" story
on was just that — a coincidence.

**The actual mechanism, and it is one layer down.** The listing's items live in the
`content-listing` component's `content_data->'articles'`, and that array is written by exactly
one thing — `save_page_sections_overwrite`, i.e. a **section re-resolve**
(`page_component_history` shows no other writer for this page). Its `articles` field declares
`"source": "query.blog_posts"` (`content_components.input_schema.fields`), so the images come
from `queryresolve`'s `pageImageProjection` — which joins the card correctly:

```sql
LEFT JOIN assets ca ON ca.site_id = p.site_id AND ca.entity_type = 'page'
                   AND ca.entity_id = p.id AND ca.purpose = 'card' AND ca.status = 'active'
```

**Assemble-mode re-render (`page_rerender` with NO `spec.reason`) re-assembles the STORED array
verbatim, empty `image` strings included.** Only a re-render carrying
`spec.reason='section_data_resolved'` re-runs the query and refreshes the array. So the defect is
sharper than filed:

> **The listing IS re-rendered, in the mode that structurally cannot pick the change up, and
> nothing in the card-landing chain ever requests the mode that can.**

That is worse than "no trigger", because every routine chrome propagation — the assemble mode
this estate rightly prefers, since it cannot escalate a page to the content writer — re-renders
the listing and silently re-affirms the stale data. The page looks freshly built and is not.

**Proven by fixing it.** Dispatched `page-rerender` with `spec.reason='section_data_resolved'` for
`index` and `guides-index` at 2026-08-24 ~19:25Z, after checking the escalation precondition
(no schema-required `source:"llm"` field missing on either page — the check the leopardess
runbook's header warns about). Result within 90 seconds:

```
stored array:  0 empty image fields of 12   (was 4 of 12)
served page:   12 cards, 12 with an image, 0 without   (was 8 of 12)
```

**What survives from the original filing:** the symptom, the evidence that the assets were all
correctly derived/linked/keyed (`asset_key` matches `contentCardKey`, `entity_id` matches the
page, all `active`), the four-checks-miss-it analysis, and fix candidate 1 — which is now better
targeted: the card-landing chain should emit a **section re-resolve for the consuming listings**,
not merely "a re-render", because a re-render of the wrong mode is what has been happening all
along.

**What this cost, recorded because it is the lesson:** I built a mechanism out of a plausible
timeline (cards landed, then a wave, then they showed) without checking whether that wave's MODE
could have caused the effect. The check that would have caught it is one query —
`spec->>'reason'` on the re-renders I was crediting — and I ran it only after the file was
committed. **A correlation that explains two cases is not a mechanism; the mode was the variable
I never looked at.**

---

## 2026-08-24 ~20:50Z — taken up by session `bugs_open/384` (platform fix; the filer hand-repaired the instance and moved on)

**Ownership.** The filing session was still alive at pick-up (19:14Z, on other work) but had committed nothing further and had no WIP on the fix site; `who-owns` → "(none identified)". Resumed here rather than in parallel. Lane docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_384_page_list_invalidation/` (PLAN / RUNBOOK / NOTES / README_where_we_are).

**Corrections to this file's premises, found while researching:**
- `bugs_open/083` → it is `bugs_closed/083` (closed 2026-08-22): the `detected-item-promoter` drains handler-bearing `detected` items since 08-15, so "this estate's `detected` items do not drain" holds only for handler-LESS findings. Candidate 2 (a check emitting `page_rerender`/`page-rerender`, a pair with 1,323 completes in 14d) would drain.
- `bugs_open/052` → `bugs_closed/052`.

**The class is wider than cards `[MEASURED 2026-08-24 19:2xZ]`.** `queryresolve.pageImageProjection` (card, else current plan hero, else "") is shared by `pages_where_type:*`, `blog_posts` and `pages_under_section`; `flag_page_image_rebuild` re-renders only the ARTICLE when a hero lands, so a hero landing leaves listings stale the same way. Fleet pair census (card asset ↔ stored entry): `content-listing`/`blog_posts` 32 pairs / 0 stale (after the hand repair); `tool-list` 37 / 0; **`tool-cta` 62 / 14 stale** (5 written after the card landed) — no served defect only because `tool-cta`'s template renders no `image`. Demand: 41 card landings / 14 days across 8 sites.

**What was built (committed 2026-08-24, inert until the next chassis roll — register PBP-048):** ONE seam. `queryresolve.PageListConsumerPages` derives "which pages consume a page-image query source" from `content_components.input_schema` (owned pages excluded — page-rerender's reasoned branch fails `save_sections`' ownership refusal); `actions.requestPageListReresolve` files one `page_rerender` / `spec.reason='section_data_resolved'` per consumer via the canonical `insertPageRerenderItem`, never failing the caller. Called from `derive_card_asset` (cause `card_landed:<page>`) and from `flag_page_image_rebuild` (cause `page_image_landed:<page>`, deferred when a card derive was raised). The page-image source set is pinned to the resolvers by a test that drives every handler and records which SQL reads the card join. Fix candidate 1, as this file's correction re-targeted it: a **section re-resolve**, not merely a re-render.

**Candidate 2, built the same evening (Phase 2):** `discovery_checks/check_page_list_stale.go` (`page_list_stale`) compares each consumer page's stored array against a fresh resolve per url on `image` and files one `page_rerender`/`section_data_resolved` at `detected` under the key the event emitter uses (so the two collapse). Unknown (erroring/empty resolve, or unreadable `content_data`) is counted, never treated as current; **no retraction arm** — measured live+archive 2026-08-24: `page_rerender` has 18,360 rows from 122 producers and 0 ever retracted, while `needs_rerender` (635 rows, 21 filers, 17 retracted) works because it has exactly ONE retraction authority; the condition is single authority, not few producers, and this sweep cannot be that authority for a shared key. **Enablement is HELD** — `sql_for_agents/603_enable_page_list_stale_HOLD.sql`, to be applied by hand after the registering binary has rolled and its capability list names the check. ~~Its first sweep will re-render the 4 sites holding the 14 stale `tool-cta` entries.~~ **CORRECTED 2026-08-25: FALSE, and this lane's own council revision made it false.** The round-2 bound (count only consumers whose template actually renders `.image`) narrows the shared lookup, so it narrows the SWEEP too — and `tool-cta` (**59** live instances as of 2026-08-25) renders no image. Simulated against the live fleet under the shipped predicate `[MEASURED 2026-08-25 09:42Z]`: **the sweep would file ZERO items today, on every site.** That is correct behaviour (a stale-but-invisible array is a re-render for no visible change; and `template_changed` re-resolves the array if such a template is ever changed to render the image) — but it means enabling the sweep buys INSURANCE, not a repair. `WRONG_CALLS.md` 2026-08-25. Still latent, not fixed: `rebuild_blog_listing_action.go:212-220` writes `"image": ""` for every listed post (0 of 3 `blog-index` pages list a carded post, 2026-08-24) — the sweep now catches it when it fires. Why owned pages are excluded, with the unit stated: `[MEASURED 2026-08-24, live table, by CAUSE — `error LIKE '%rebuild_policy=owned%' OR '%OWNED_PAGE_GUARD%'`]` **13 of 18** `page-rerender` failures on `rebuild_policy='owned'` pages in the last 14 days are ownership refusals, all of them `cta_links_stale` items from the discovery checks — a population this change does NOT touch (the exclusion keeps only *this emitter's* items off owned pages). Two earlier figures written by this lane and its peers were wrong in different ways and are retracted: "12 failures" (those were `orchestration_states` RUNS, retries included) and "4 items" (classified by the `OWNED_PAGE_GUARD` marker, which was only added on 2026-08-19 — a marker-based classifier has a birth date). Recorded in `WRONG_CALLS.md`.

**Verification (post-roll, at the artefact):** in the lane RUNBOOK — an induced card landing on a site with a known consumer count N must produce exactly N items with `spec.cause='card_landed:<page>'`, N `page-rerender` COMPLETED with `rerender_sections.escalated=false`, and `pages.deployed_at` advanced on the listings BECAUSE of them (`page_component_history.source_item_id`). The served 12/12 on dartsonline is not discriminating; the rows are.

---

# STATUS 2026-08-25 — the seam is LIVE and swept; all four open decisions are RULED and SHIPPED

Not closed: this stays in `bugs_open/` until the Go half has ROLLED (the migrations are live,
Go is not — see "what is not live" below). The defect itself is fixed and proven end-to-end.

## What is live and proven

| piece | state | proof |
|---|---|---|
| the event seam (card lands → consumers re-resolve) | LIVE | induced landing filed exactly N=2 consumer items; `index` chain closed, `escalated=false`, array rewritten `[MEASURED 2026-08-25 09:51Z]` |
| `page_list_stale` sweep | **ENABLED and PROVEN LOOKING** (migration `603`, applied by hand 11:37Z) | checks array 44 → 45, verified against the pre-image in `agent_definitions_backup`; registered on **301 pods** at `4c996e1b5cb9`; and `[MEASURED 2026-08-26 08:25Z]` loancalculator `stale:0 / current:**25** / unknown:0` — the `current>0` pass, not the blind zero |
| `tool-cta` renders the image (decision 4) | LIVE | migrations `614`/`615`; 40 items filed, 0 on archived/owned; 6 pages re-rendered so far, all carrying real card URLs, **0 empty `src`** |
| `rebuild_blog_listing` blank-image fix (decision 3) | committed, **NOT rolled** | `7720dc76c` + `bafd4411c` |
| dependency-scoped consumer lookup (decision 2, RFC_052) | committed, **NOT rolled** | `72469c556`; council `e1d32ca2` APPROVED |

## The owner's four rulings, and what each one turned out to mean

1. **Enable the sweep** — done. Predicted and confirmed: it files ZERO items today, because
   every image-rendering listing is current. It is insurance, not a repair.
2. **Generalise the lookup now** (RFC_052) — done and CLOSED. The declaration is now a per-source
   dependency set (class → the item keys it feeds), and both producers migrated onto it. All
   three behavioural changes measured as no-ops on today's fleet before shipping.
3. **Fix the action** — done. It was NOT "latent": `rebuild_blog_listing` is an unconditional
   step of `rerender-pages` (42 runs/14d) and leopardess's `blog-listing_pre_037` IS a consumer,
   so the two were competing writers of one field. Still changes no stored byte today, because
   0 of the 47 listed articles has a card or hero.
4. **Fix the `tool-cta` entries by changing the template** — done, after a SECOND ruling. The
   measurement that prompted it: the change would have put 144 of 228 entries onto full-bleed
   page HEROES, all on loancalculator. Owner: derive the cards first. 10 of 10 landed there;
   the fleet now resolves **206 card crops / 0 heroes / 42 gated-blank**.

## Three findings this work produced that outlive it

- **`derive_card_asset` CROPS AN EXISTING HERO — it does not generate imagery.** All 29 D1
  items completed; only 10 cards landed. loanandmortgagecalculator's 19 tool pages have no hero
  of any kind, so the action completed truthfully with `derived:false` and produced nothing.
  Their 12 tool-cta entries stay blank permanently, via the gate. Check the ASSETS, never the
  item status.
- **A stored-but-unrendered key RE-STALES after any repair**, because the resolver always
  returns it and the seam always skips non-rendering consumers. Only two states are stable:
  RENDERED, or NOT STORED. **"Not stored" is UNSAFE today** — of 28 live (component,
  query-array-field) declarations, **17 render an item key their own schema omits** (every
  directory listing renders `.url` without declaring it). Do not reach for it as the tidy fix.
- **A template edited by SQL ships NOTHING** without a hand-written fan-out, and the estate's
  own fan-out query (`component-template-fixer.create_rerender`) has **no page-status filter** —
  it would have re-rendered and re-published the 16 `tool-cta` instances sitting on ARCHIVED
  pages (`bugs_open/098` exactly). Both written up in `LANDMINES.md`.

## What is NOT live, and how to check rather than assume

The Go changes (decisions 2 and 3) are inert until a chassis image is built and rolled. Verify
at the binary — `service_binary_capabilities`, or the pod's `build provenance` line plus
`git merge-base --is-ancestor <commit> <that sha>`. Never `strings`, never a discovery grep for
"some 40-hex string", and always run a known-present and a known-absent control in the same
breath.

## Owed, and honestly still open

- ~~**The sweep's `current > 0` proof.**~~ **OBTAINED 2026-08-26.** It arrived overnight, once the
  rotation reached sites with non-empty listings: **loancalculator.co.uk `consumer_pages: 25,
  stale: 0, current: 25, unknown: 0`** (08:25Z), robot-hands 3/3, finetuning 3/3,
  loanandmortgagecalculator 2/2, vonc 1/1, webdesign 1/1, garden-tools 1/1. That is the PASS —
  it looked at 25 listings on one site and found every one current. **Items filed all-time: 0**,
  which is the predicted and correct result.
  The earlier `unknown` readings were not faults: lampenkap has ONE page and ZERO `tool` pages,
  so its `tool-list` array is legitimately empty and an empty resolve is classified UNKNOWN by
  design. agritec shows the mixed case (`current: 1, unknown: 1`). ⚠ The reporting hazard stands
  and is worth keeping in mind: `stale=0, current=0, unknown=N` looks identical at a glance to
  the blind case, and `consumer_pages` is the field that tells you the lookup ran at all.
- **The escalation rate**, one week on, against the refreshed baseline **1 of 36**
  `section_data_resolved` runs in 14 days. 0 of 5 so far on the tool-cta batch.
- **Council verdicts**: `7553c120` (decision 4) and `e1d32ca2` (decision 2) APPROVED;
  `170147b4` (decision 3) REVISE at round 1 → revised and resubmitted on the same correlation,
  verdict pending. The round-1 objection was correct and is described in NOTES.
- **Not this lane's, but blocking a fleet-wide rule:** `go test ./cmd/config-key-audit/` does not
  COMPILE — `livedeclarations_test.go:151,153` reference `livespec.DeferredDeclarations`, renamed
  2026-08-23 (`livespec.go:77`). Committed at HEAD by the 363 lane (`18661b3c7`). So the
  optional-key parity test CLAUDE.md instructs every author to run cannot be run by anyone, and
  the pre-commit hook reports it as "the tree does not build", which reads as a local problem and
  is not one. `go build ./...` is clean.

---

## ⚠ CORRECTION to MY "How to verify a fix" section, 2026-08-26 — do not require `deployed_at` to advance

My verification protocol above says to *"pin the listing's `deployed_at` before and after and require
that the re-render was caused by the card landing rather than coinciding with it"*. **The
`deployed_at` half is wrong and produces a false NEGATIVE.** Corrected by the lane that implemented
the fix, who wrote the same requirement into their first draft and then disproved it by running it:

**On a listing whose array is already current, the re-rendered HTML is byte-identical and the deploy
is a legitimate no-op.** So a correct fix, doing exactly the right thing, fails my test. The
protocol demanded evidence the mechanism has no reason to produce.

**What to require instead** — the causal claim was the right instinct, the artefact was the wrong
place to look for it:

- the consumer set is **derived**, not named: N items filed, one per page whose
  `content_components.input_schema` declares the consuming query;
- each carries `spec.reason='section_data_resolved'` and a cause naming the landing
  (`card_landed:<page>`);
- the stored array's empty `image` count for that entity goes to zero;
- `escalated=false`.

Byte-identical output with those five facts true is a **pass**, not a failure.

**The general shape, which is why this is worth a correction block rather than an edit:** I wrote a
verification that could fail for a reason unrelated to the defect — the mirror of the failure this
file is otherwise about, and the same family as my own "a correlation that explains two cases is not
a mechanism" note above. **A test that can go red while the fix is working is as useless as one that
cannot go red at all**, and it is more expensive, because it sends the next person to un-fix a
working thing.


---

# CLOSE-OUT VERIFICATION 2026-08-26 — proven TWICE on natural triggers, and the one residual this fix structurally cannot reach

## The Go halves have ROLLED — the "committed but inert" caveat is retired

`[MEASURED 2026-08-26]` both running chassis builds (`e7f1045fddec`, 1,261 pods; `2fb40a960f88`,
10 pods) carry every commit of this lane — `7720dc76c`, `72469c556`, `bafd4411c`, `efc0db7bc`,
by `git merge-base --is-ancestor` against the stamp the pods report. **And they have RUN, not
merely shipped:** `render_news_section` filed 6 items and `render_directory` 1 through the
migrated `ConsumerPages` lookup in 12 hours; `rerender-pages` completed 62 runs; all three
blog-index listings were rewritten by the new projection path.

⚠ The ancestry check must be run against the CURRENTLY reported stamp. My first attempt compared
against yesterday's `4c996e1b5cb9` and returned "NOT in the running build" for all four commits —
a false negative produced entirely by a hardcoded sha that a roll had superseded.

## Proven twice on NATURAL triggers — not induced

Every earlier proof of this seam was an induced landing. Today it demonstrated itself unprompted,
on two sites:

| site | card landed | seam response | outcome |
|---|---|---|---|
| leopardessconsulting.co.uk | 14:42:45 | 3 consumer items filed at **14:42:46** — one second later | blog listing array rewritten 15:30:34; **11 of 11 entries carry an image**, 0 blank-with-imagery (was 4) |
| finetuning.uk | 17:25:45 | 3 consumer items filed at **17:25:45–46** | in flight at time of writing |

Three further leopardess landings at 14:55/14:56 each reported
`page_list_reresolve: "deduped", deduped: 3, queued: 0` — collapsing onto the open items via the
shared `PageRerenderItemKey`. **Four landings in fourteen minutes produced three re-render items,
not twelve**, which is the dedup contract behaving exactly as PBP-048 specifies. All nine items
completed with `attempt_count = 0`.

## THE RESIDUAL: `owned` pages accumulate stale listing arrays permanently, and this fix cannot reach them

`[MEASURED 2026-08-26]` fleet-wide, 1,008 stored listing entries; **17 are blank where a card
exists**. Every one resolves:

- **14 sit on `rebuild_policy='owned'` pages** — finetuning.uk/`llm-cost-calculator`,
  leopardessconsulting.co.uk/`llm-cost-calculator` and `/tool-ai-vendor-trust-checklist`. None was
  in the `615` fan-out and none can be, because `PageListConsumerPages` excludes owned pages BY
  DESIGN: page-rerender's reasoned branch runs `save_sections`, whose ownership refusal
  (`bugs_open/208`, OWNED_PAGE_GUARD) would FAIL the run.
- **3 sit on generic pages and are the seam in flight** (finetuning, card landed 25 minutes
  before the census).

So there are **zero genuine misses** — but the owned-page exclusion has a consequence this file
has not stated until now, and it should not be discovered by someone else later:

> **An `owned` page's `query.*` listing array is never re-resolved by this seam, by the sweep, or
> by the `template_changed` fan-out. It goes stale on the first card landing after its last
> resolve and STAYS stale indefinitely.**

That was tolerable while `tool-cta` rendered no image. **It is now visible**: migration `614` made
`tool-cta` render `.image`, and two of the three affected pages carry `tool-cta`. Their tiles show
images for entries that have them and nothing for these 14 — on pages a human owns and did not
change.

The exclusion itself is correct and should not be removed; the gap is that nothing else covers
those pages. `bugs_open/333`'s lane established that a per-BRANCH refusal can only be expressed at
selection time, and migration `486` shows the existing remedy shape for owned pages: route them to
`section_edit` → `section-editor` instead of `page_rerender`. **Nobody has applied that shape to
listing staleness**, and doing so is the natural follow-up — filed here rather than fixed, because
it is a new seam for owned pages and belongs in its own round, not in this lane's close-out.

## What remains, and none of it blocks closing

- **The escalation watch.** 603's header asks for the rate re-read a week on, against the
  refreshed baseline of **1 in 36** `section_data_resolved` runs in 14 days. Zero escalations
  across every seam-driven run this lane produced (42 tool-cta runs + the natural landings).
- **`bugs_open/404`** — a separate defect this lane FOUND, not part of this one. Candidate 0
  (the vocabulary/reader parity test) is unclaimed.
- **One permanently-failing page**, `ai-agent-orchestration.com/tool-automation-savings-estimator`,
  refused by the section component floor (`77→37` class attributes). Pre-existing: it failed three
  times on 2026-08-24, before this lane touched anything, and those were the fleet's only other
  floor refusals in 14 days.
- **The owned-page residual above**, which is the only thing here that is genuinely unfinished
  work rather than a watch item.


## UPDATE 2026-08-26 20:45Z — fresh build, fourth natural demonstration, residual now measured exactly

- **Fresh chassis build `b34c24f4c65b`** (95 pods) rolling alongside `e7f1045fddec` (700). All four
  of this lane's Go commits are ancestors of **both**, and the new build is a strict descendant of
  the old. The lane's behaviour holds whichever pod serves.
- **Fourth natural trigger:** finetuning.uk's three items (filed 17:25:45–46) all completed by
  18:44 with `attempt_count = 0`; its arrays were rewritten 19:13–19:15 and now carry **0 blank
  entries on every generic listing**. vonc.com's card landed 19:59:30, the seam fired, and its
  re-render is pending.
- **The residual, measured exactly** `[MEASURED 2026-08-26 20:40Z]`: blank-where-a-card-exists is
  **14 on `owned` pages (3 pages) and 1 on a generic page** (vonc, seam in flight). Every generic
  page repairs itself; owned pages never do. That is the §CLOSE-OUT residual, now with a number.
- **Cold-start doc for this lane is now**
  `docs/agent_docs/docs024_key_docs_latest/bugfix_384_page_list_invalidation/HANDOFF_2026-08-26_continue_here.md`.


## UPDATE 2026-08-31 16:10Z — **STAYS OPEN.** A live recurrence on a generic page, and the clean census was flattered

Re-measured after five idle days (last lane commit 2026-08-26). **This reverses the 08-26 update's
"ready to close" and the handoff's §1/§6.**

**Still true:** the lane's code is live. Fleet is on a SINGLE build `ef06af0e0afc` (342 pods,
current `last_seen_at`); all four lane commits are ancestors of it. The owned-page residual is
unchanged at **14 blanks / 3 pages** and is still correctly described.

**What is new, and why this cannot close:**

1. **The defect is reproducing on a GENERIC page.** `leopardessconsulting.co.uk/blog` serves
   **2 text-only cards where an active card asset exists**, and they are the **first two tiles in
   the grid**. Verified at the served artefact — `curl .../blog.html` returns 11
   `src="/assets/images/card-*.jpg"` for 13 array entries, and both guides render
   `<article class="article-card hover-lift">` directly into `<div class="article-card__content">`
   with no image node. Cards landed **2026-08-27 22:37:25** and **22:37:49**; still blank on 08-31.

2. **The seam did its job — this is a consumption failure, not a detection failure.** Nine
   `page_rerender` items were filed for the page, the first within **40 ms** of the card landing.
   Every spec is correct: `reason=section_data_resolved`, right `page_id`,
   `consumes=["query.blog_posts"]`, and the component's own field source IS `query.blog_posts`
   (`content_components.input_schema`, component `blog-listing_pre_037`). Dependency scoping matches.

3. **Two of those items COMPLETED GREEN and repaired nothing.** `e1f2dd23` (22:37:25.98 →
   complete 22:58:18) and `5f78c1e4` (08-28 09:39:12 → 09:40:36) each deployed `blog.html` +
   `tools/assets/blog-listing.js` with real commit shas. Yet `page_components.updated_at` for the
   listing row is still **2026-08-27 21:34:20** — an hour BEFORE the cards landed. The RUNBOOK's
   causation leg (measured 08-25) establishes that column advances whenever the array is rewritten.
   **So the array was never rewritten.** The remaining seven items sit `unresolved`,
   `attempt_count=0`, oldest 08-28 09:30, newest 08-31 10:37 — never picked up.

4. **No mechanism is asserted for (3), deliberately.** The runs are unrecoverable —
   `orchestration_states` retains ~1 day (oldest row 2026-08-30 15:07). The live gate is NOT the
   suspect: the live `page-rerender` row's `check_rerender_mode` condition does include
   `section_data_resolved` (read from `agent_definitions`, not a seed). Unfalsified candidates:
   the `plan.Status != "ready"` carry branch (`rerender_page_sections_action.go:509`) and the
   `listedOnly` floor in the `blog_posts` resolver. **A `090` diagnosis run is owed before anyone
   writes a root cause here** — durable, cross-cutting, cause not where the symptom is.

5. **⚠ The census that supported closing was FLATTERED, and would read clean today for reasons
   unrelated to the fix.** Two demand controls, neither present in the 08-26 read-out:
   - **Card production has been ZERO for two days.** Cards landed/day: 08-26 **89**, 08-27 **109**,
     08-28 **46**, 08-29 **18**, 08-30 **0**, 08-31 **0**. No demand on the seam since 08-29.
   - **Fleet work-item completion collapsed.** `page_rerender` created vs now-complete: 08-24
     1400/1390 (**99%**), 08-27 2138/1947 (91%), 08-28 338/146 (43%), 08-29 179/7 (**4%**),
     08-30 210/3 (**1.4%**), 08-31 300/144 (48%). **1,076** `page_rerender` rows `unresolved` in 7
     days, beside 1,395 `undeployed_asset` and 466 `required_fields_missing`. Fleet-wide, matching
     the `bugs_open/413` starvation shape (owned by the `dispatch_throughput` lane) — **not this
     lane's to fix, but this lane must not close on a queue that is not running.**

**Bar for closing, restated:** a generic-page landing that lands, fires, is CONSUMED, and rewrites
the array — observed on a queue that is actually draining. Item (1) is a standing counter-example
today.

Measurements and traps in
`docs/agent_docs/docs024_key_docs_latest/bugfix_384_page_list_invalidation/NOTES_page_list_invalidation.md`
(entry 2026-08-31); the handoff carries a CORRECTED block at its head.

### CORRECTION to the update above, same day 16:25Z — the queue collapse is a KNOWN, ENDED credit outage; the open question narrows to ONE run

Item 5's second bullet called the fleet collapse "the `bugs_open/413` starvation shape". **That was
a guess from shape, made without reading the owning lane's record, and it is wrong.** The
`dispatch_throughput` lane had already measured it `[MEASURED 2026-08-31 09:45Z]`: an **Anthropic
credit-balance outage**, window **~2026-08-28 → 2026-08-31 ~08:40–09:00Z (~2.5 days)**, their
"D4 case 4". Recovery ~09:00Z today; their post-recovery read is clean (zero stuck claims, zero
pins, nothing owed dispatch-side). 413 is a different, real defect and is not this.

Consequences, which **narrow rather than weaken** the case for keeping 384 open:

- **Items 5a and 5b keep their caveat but lose their mystery.** Zero card production on 08-30/31
  and the 7 `unresolved` re-render items are the outage. They should drain now that it has ended —
  **to be re-read, not assumed.**
- **Item 3's first run survives every confound.** The month's credit outages ran 08-25 23:47→08-26
  08:55, 08-27 11:30→13:35, and 08-28→08-31 09:00. The completion at **2026-08-27 22:58:18** sits
  in a HEALTHY window between the second and third. It deployed `blog.html` with a real sha and
  left the array's `updated_at` at 21:34 — an hour before the cards that triggered it landed.
- **That one run is now the whole of the open question**, and it is exactly what a `090` should be
  pointed at: a correctly-specced `section_data_resolved` item, consumed in a healthy window,
  completing green without rewriting the array it was filed to refresh.

**The close bar is unchanged:** one generic-page landing observed end-to-end — lands, fires, is
consumed, array rewritten — on a queue that is now, post-recovery, actually draining.

## UPDATE 2026-09-02 16:0xZ — re-read on a HEALTHY queue: the page did NOT repair, and the sweep has NEVER RUN ONCE

Owner asked for the page to be re-read before firing a `090`. Done, on a fleet that has recovered.
**The re-read removes the last confound and makes the case stronger, not weaker.**

**The demand controls now PASS — this is no longer an idle-system reading.** `page_rerender`
created vs complete: 09-01 **1164/953 (82%)**, 09-02 **1179/1084 (92%)** (against 4% and 1.4% at
the outage trough). Cards are landing again (10 on 09-01, 6 on 09-02). The queue is draining.

**The page did not repair.** `curl https://leopardessconsulting.co.uk/blog.html` is **byte-identical
to 08-31 — 38,319 bytes, the same 11 card images for 13 entries.** Both guides still render as
text-only tiles, still the first two in the grid, six days after their cards landed.
`page_components.updated_at` is still **2026-08-27 21:34:20**.

**Attribution corrected — I got this wrong on 08-31.** I wrote "the seam fired nine times". It did
not. Splitting by `created_by`:
- **The card-landing seam (`derive_card_asset`) filed 6 items and ALL 6 COMPLETED**, including both
  of ours (`card_landed:tool-gdpr-ai-risk-assessment-guide` 08-27 22:37 → complete;
  `card_landed:tool-monitoring-scope-estimator` 08-28 09:39 → complete). **Defect #1 stands and is
  still unexplained: a correctly-specced item, consumed in a healthy window, completing green
  without rewriting the array.**
- **The 9 stuck items are the SWEEP's** (`completeness-discovery-agent`, `spec.check=page_list_stale`),
  filed daily 08-28 → 09-01. Not the seam.

**DEFECT #2, new, and it retires a claim in §3 of the handoff.** The sweep (migration `603`) is
listed there as LIVE. It detects correctly and **it has never once run**: `page_list_stale` has
filed **12 items in its entire lifetime — live AND archive, all time — every one `unresolved`,
`attempt_count = 0`, zero ever claimed or completed.** Three pages, three sites (leopardess/blog 9,
agritec.uk/tools 2, garden-tools.uk/tool-watch-service-interval-calculator 1).

**Mechanism, verified first-hand at source and artefact:**
1. `insertWorkItem`'s two-strike anti-churn arm counts prior siblings on the same `item_key` with
   `status IN ('complete','failed')` over 7 days (`load_work_item_actions.go:1985-1993`).
2. At `>= 2` it brands the new row `status='unresolved'` (`:2033`).
3. `unresolved` is in `workItemTerminalStatuses` (`work_items_common.go:48`) and
   `claim_work_item_action.go:135` claims only `('triaged','approved')` — **so the row is born
   terminal and unclaimable**, and holds no dedup slot, so the next sweep files another born-dead row.
4. **Every stuck row carries the brand**: `[unresolved after 6 attempts] Page-list on blog shows 2
   stale image(s) …`. Strike composition on that key: **6 `complete`, 0 `failed`** — every counted
   "attempt" was one of defect #1's false successes.

**So the two defects are coupled, and that is the finding.** Defect #1 makes the seam complete
without repairing; each false success is a strike; after two, the anti-churn arm kills the sweep's
item at birth. **384's own second line of defence is disabled by 384's own first line, on exactly
the pages that need it.**

**This is NOT a new bug — it is `bugs_open/389`'s class**, whose class 1 already names the chain
("a no-op rerender that completes, strikes, and manufactures `unresolved` stock"). 389 is OWNED by
the `bugfix_308` lane (`who-owns.py`), so the evidence went **into 389 as a CONTRIB**, with the one
property I could not find stated there: the strikes and the parked item came from **two different
producers sharing an `item_key` on purpose**, and all six strikes were successes. **No fix attempted
here.**

**§8.1's escalation watch is VACUOUS and should not be re-read as written.** "Zero escalations
against a 1-in-36 baseline" is zero over an empty denominator — the sweep never ran. It read as a
clean bill of health for eight days.

**Still open, unchanged:** defect #1's mechanism. The `090` the owner deferred is still the right
next step and is now sharply targeted: *a `section_data_resolved` item, correctly specced and
consumed in a healthy window, completes green and does not rewrite the array it was filed to
refresh.* The owned-page residual (14 blanks / 3 pages) is also unchanged.

## UPDATE 2026-09-02 17:1xZ — the `090` came back UNVERIFIABLE, and its "still needed" list cracked it: **the seam's vehicle has never written a listing array**

Owner authorised the run. `090` fired: intake `d4f745e6-3f79-42a8-8f71-bb611736912c`,
**run correlation `149ec925-ffb7-41eb-806a-1595b8ff2226`**, 5 iterations, all three orchestrations
COMPLETED.

**Verdict: `UNVERIFIABLE` — "NOT CONFIRMED (stopped: iteration-cap)". Not a refutation, and not a
waste.** It independently reproduced the state evidence (the 21:34:20 freeze; the two
`section_data_resolved` items completing at 22:58:18 and 09:40:36 without moving it) and then said
precisely what it could not see: the body of `rerenderFlatSections` that chooses carry-vs-resolve,
the `queryresolve.Resolve` body, and the live `check_rerender_mode` config. **It also named an
alternative I had not considered and could not rule out from `updated_at` alone: "the row was never
touched" vs "touched but the value was unchanged".** Chasing exactly that is what produced
everything below.

**First: `updated_at` on `page_components` is NOT trigger-maintained** (no such trigger exists;
only the two `page_component_artefact_archive` triggers). So a frozen `updated_at` really was
correlation, not proof — the loop was right to withhold.

**`page_component_history` settles it** `[MEASURED 2026-09-02]`. The archive triggers fire whenever
`rendered_html` changes, or `content_data` changes with `rendered_html` static. For this listing
row the only rows since 2026-08-25 are:

| archived_at | application_name | n_articles | blank |
|---|---|---|---|
| 2026-08-27 21:34:20 | `action:rebuild_blog_listing` | 11 | 0 |
| 2026-08-27 06:55:03 | `action:rebuild_blog_listing` | 11 | 0 |
| 2026-08-26 22:02:54 | `action:rebuild_blog_listing` | 11 | 0 |
| 2026-08-26 15:30:33 | `action:rebuild_blog_listing` | 11 | **11** |

**Nothing at 22:58:18 or 09:40:36.** Neither completed run changed `content_data` OR
`rendered_html`. The "touched-but-unchanged" alternative is dead: they wrote nothing.

**And the writer is `rebuild_blog_listing` — not the path the seam files for.** Fleet-wide census
of every write to a component carrying a `query.*` array field, last 14 days `[MEASURED 2026-09-02]`:

| application_name | writes | pages | newest |
|---|---|---|---|
| `action:rebuild_blog_listing` | 5 | 1 | 2026-08-27 21:34:20 |
| `psql` (a hand write, another lane) | 1 | 1 | 2026-08-31 13:37:04 |

**`rerender_page_sections` has written a listing array ZERO times in 14 days.**

**The live `page-rerender` workflow has no rebuild step.** Its steps are exactly:
`check_rerender_mode → rerender_sections (rerender_page_sections) → check_escalated →
save_sections (save_page_sections) → render_page (rerender_single_page) → check_skipped →
deploy_page → update_status → complete`. **No `rebuild_blog_listing`.** So the item the seam files
runs a workflow that renders and deploys — which is why it completes green with a real commit sha —
and, on the evidence above, does not rewrite the array it was filed to refresh.

**This puts §4's "four natural demonstrations" in doubt and they must be re-read before being
quoted again.** The 2026-08-26 15:30:33 archive row is the celebrated leopardess repair
("array rewritten 15:30:34 — 11 of 11 entries carry an image"); its pre-image is 11-blank-of-11 and
its writer is **`action:rebuild_blog_listing`**. The seam files `page_rerender`; the repair was
performed by a different action on a different trigger. **[INFERRED, not yet measured]** that the
other three demonstrations are the same shape — nobody has checked their `application_name`, and
that is the first thing the next session should do.

**What is now MEASURED vs still open.** Measured: no write by the two completed runs; the writer
census; the workflow's step list. **Still open, and it is a sharper question than the one I sent to
the loop:** is `rerender_page_sections` *supposed* to rewrite a `query.*` array and failing to
(the carry branch / resolver), or was the seam pointed at the wrong vehicle from the start? A
re-run with whole-file `SEED_SCOPE` (see the RUNBOOK note) is the cheap next step.

**Unchanged:** defect #2 (the sweep born terminal, contributed to `bugs_open/389`), and the
owned-page residual (14 blanks / 3 pages). 384 stays open.

### RETRACTION, same day 17:3xZ — my "zero writes in 14 days" fleet claim is WRONG. The page-specific finding stands.

**RETRACTED:** *"`rerender_page_sections` has written a listing array ZERO times in 14 days"*, and
the writer table under it (`rebuild_blog_listing` 5 / `psql` 1). **Do not quote either.**

**Why it was wrong.** I joined `page_component_history h JOIN page_components pc ON pc.id =
h.component_id`. **`save_page_sections` DELETES and RE-INSERTS the `page_components` row** rather
than updating in place (visible as `op=delete` / `source=save_page_sections_overwrite`), so its
history rows point at a row id that no longer exists. Measured `[2026-09-02]`: **44,781 of 45,540
history rows — 98.3% — match no live `page_components` row.** My census therefore ran over ~1.7% of
the table, and the only writer it could still see was the one that updates IN PLACE
(`rebuild_blog_listing`). **The instrument selected for the answer it gave me.**

**The corrected census, keyed on `page_id` (stable across the delete/reinsert), listing-hosting
pages, last 14 days `[MEASURED 2026-09-02]`:**

| writer | source | rows | pages | newest |
|---|---|---|---|---|
| (empty) | `save_page_sections_overwrite` | **4,969** | **161** | 2026-09-02 17:26:19 |
| `action:section_editor.update` | artefact_archive_trigger | 93 | 49 | 2026-09-02 08:20 |
| `psql` | artefact_archive_trigger | 12 | 9 | 2026-09-02 15:46 |

So `save_page_sections` — the rerender workflow's own write step — is a **prolific** writer. The
claim that the seam's vehicle never writes is false.

**What SURVIVES, because it was keyed on `page_id` from the start:** for
`leopardessconsulting.co.uk/blog` (page `05269427-…`), the only history rows since 2026-08-25 are
**four `action:rebuild_blog_listing` writes** (08-26 15:30:33, 08-26 22:02:54, 08-27 06:55:03,
08-27 21:34:20) and **nothing at 22:58:18 or 09:40:36**. Not even a `save_page_sections_overwrite`
row. That finding is sound and is the one that matters.

**And it sharpens the question rather than dissolving it.** `save_page_sections` writes 4,969 times
across 161 listing pages in a fortnight — but produced **no row at all** for this page on two runs
that reached `deploy_page` (a real commit sha). So the live question is now: **on those two runs,
why did the workflow reach deploy without `save_sections` writing anything?** Candidates, none
tested: `rerender_sections` carried every section so the save was a no-op; `check_escalated`
routed around the save; or `render_page` (`rerender_single_page`) deploys from stored
`content_data` independently of the save step.

**Also still true, and unaffected:** the §4 doubt. The 08-26 15:30:33 leopardess repair really was
written by `action:rebuild_blog_listing`. The other three demonstrations remain **unchecked** —
and note that a `save_page_sections` repair would NOT show under a component-id join, so anyone
re-checking them must key on `page_id`.

### RESOLVED, same day 17:5xZ — §4's demonstrations are GENUINE (3 of 4), and the defect is specific to BLOG-LISTING pages

I checked the other three demonstrations, keyed on `page_id` as the retraction above requires.
**My doubt was wrong for three of them, and I withdraw it.** `[MEASURED 2026-09-02]`

| demonstration | what actually wrote it | genuine? |
|---|---|---|
| finetuning.uk `tool-ai-readiness-checker` 19:13:31 | `save_page_sections_overwrite` (+ 4 slot deletes) | **YES** |
| finetuning.uk `tools` 19:14:21 | `save_page_sections_overwrite` (+ 3 slot deletes incl. `tool-list`) | **YES** |
| finetuning.uk `model-approach-selector` 19:15:17 | `save_page_sections_overwrite` (+ 4 slot deletes) | **YES** |
| vonc.com `archetypes` 2026-08-27 08:12:56 | `save_page_sections_overwrite` (slots `archetype-grid`, `archetype-combinations`) | **YES** (it did repair, later than the handoff's "pending") |
| leopardess `blog` 2026-08-26 15:30:33 | `action:rebuild_blog_listing` | **NO — this one was another action's repair** |

So the seam does work, on pages whose sections are written by `save_page_sections`. §4 stands for
four of its five claimed repairs; only the leopardess attribution was wrong.

**And that isolates the real defect.** `leopardessconsulting.co.uk/blog`, **complete history, all
time**:

| source | writer | rows | window |
|---|---|---|---|
| `artefact_archive_trigger` | `action:rebuild_blog_listing` | 5 | 2026-08-24 → 2026-08-27 21:34:20 |
| `artefact_archive_trigger` | app connections | 2 | 2026-08-11 |
| `save_page_sections_overwrite` | — | **1** | **2026-07-12 17:47** — and never since |

**A blog-listing page's array is maintained exclusively by `rebuild_blog_listing`. `save_page_sections`
has written this page ONCE in seven weeks.** The `page-rerender` workflow contains no
`rebuild_blog_listing` step (step list in the 17:1x update). So on a blog-listing page the seam's
item renders from stored `content_data`, deploys a real commit, and writes nothing — while the same
seam on a tool-listing or archetype page repairs correctly, because there the write goes through
`save_page_sections`.

**That is why exactly one generic page in the fleet is stuck**, and why every fleet-level aggregate
looked healthy.

**The remaining question, and it is the last one:** `rebuild_blog_listing` wrote this page five
times between 08-24 and 08-27 21:34 and **has not run since** — the cards landed at 22:37, 63
minutes after its last run. So: what triggers `rebuild_blog_listing`, and why has it not fired in
six days? Answer that and defect #1 is closed. **Not yet investigated.**

**Status of the three findings:** defect #1 now localised to the blog-listing path (above);
defect #2 (sweep born terminal) contributed to `bugs_open/389`; owned-page residual unchanged.
384 stays open.

## UPDATE 2026-09-02 18:3xZ — why `rebuild_blog_listing` stopped: the site is out of rerender-pages service, and the two-strike arm is why

`[ALL MEASURED 2026-09-02]`

**The chain, end to end:**
1. `rebuild_blog_listing` exists in exactly **one** live agent — `rerender-pages` (queried live
   `agent_definitions`; `page-rerender` does not have it).
2. `rerender-pages` is spawned by `build-dispatch-loop` — **36 of its 37 runs** in 24h.
3. In those 24h it ran **37 times across 17 distinct sites, ZERO of them leopardess.**
4. leopardess DOES get items handled by `rerender-pages` — **18 in 7 days** (live+archive). But:
   **`deactivated_component` 10 unresolved / 10 branded `[unresolved after N attempts]`;
   `needs_rerender` 5 unresolved / 5 branded; only 3 complete, all on 08-26/27.** The last of those
   completed **21:22–21:33 on 08-27** — and `rebuild_blog_listing`'s last write was **21:34:20**.
   The site's rerender pipeline stops at that minute and has not restarted.
5. So: born-terminal items → nothing eligible for `rerender-pages` → no `rebuild_blog_listing` →
   the blog array is never rebuilt → the two cards stay blank.

**This is defect #2 (389's mechanism) CAUSING defect #1's persistence.** The strikes that closed the
door were this lane's own seam succeeding repeatedly on 08-26/27 after card landings.

**⚠ But the brand is NOT unique to leopardess — do not state that it is.** Control across served
sites, `rerender-pages` items, 7 days:

| site | items | unresolved | branded | complete | rerender-pages runs 24h |
|---|---|---|---|---|---|
| relojistas.com | 43 | 36 | 36 | 7 | 3 |
| dartsonline.com | 24 | 17 | 17 | 7 | 3 |
| **leopardessconsulting.co.uk** | 18 | 15 | 15 | **3 (none since 08-27)** | **0** |
| gaswholesalers.com | 22 | 13 | 13 | 9 | 4 |
| boxingonline.com | 10 | 3 | 3 | 4 | 10 |

The arm suppresses a large share of rerender work fleet-wide. Served sites survive because SOME
keys still get through; leopardess has none left that do. **[INFERRED, not proven]** that is the
discriminator.

**TESTABLE PREDICTION — check this first next session.** The brake counts terminal siblings over a
**rolling 7 days**. leopardess's strikes fall on 08-26/27, so they age out **2026-09-02 → 09-04**.
If the reading is right the keys become filable again and the site should resume unaided, and the
blog page may repair itself. **If it repairs without intervention, the chain above is confirmed; if
it does not, the chain is wrong and something else holds the site.** Either outcome is decisive —
this is the cheapest experiment available and it costs nothing but a re-read.
It also implies a **weekly sawtooth**: a site repaired successfully enough freezes its own rerender
service for a week. Put to the `dispatch_throughput` lane; unanswered as of writing.

## Correspondence with the `dispatch_throughput` lane (2026-09-02) — two answers and one finding that is THEIRS, not this bug's

Asked them directly (they own dispatch). Their replies, and what came back:

- **Promotion is `detected-item-promoter`**, a 900s scheduled task, enabled and firing (last tick
  18:18Z), **independent of site selection** — so my "deadlock" reading was wrong in that form too.
  It promotes through doors: `pipeline IN ('build','content','design')`, handler registered+active,
  plus known-good doors from bugs 444/430/454.
- **`sites.build_status` is inert for dispatch** — not in the selector's clause list. `pending`
  frozen on leopardess is a symptom of non-visit, not a cause. (I had flagged it as a suspect.)
- They have added a **zero-eligible starvation census** to their runbook beside the per-site floor
  (their commit `155c36812`), because a zero-eligible site contributes nothing to the floor meter —
  the same absence-shaped damage as 413, one layer earlier.

**And the dig they asked for, which generalised:** why do leopardess's 51 `detected` rows fail a
door? **All 51 pass the pipeline door and fail the handler door — `handler_agent` is EMPTY.**
Fleet-wide: **1,386 `detected` rows across 35 sites, 1,386 with an empty `handler_agent` (100%);
zero with a merely-inactive handler; zero failing the pipeline door.** 12 item_types, oldest
2026-07-26 (`head_essentials_missing` 978 rows / 31 sites leads it).

So the handler door parks the entire detected population **by construction**, because no detector
populates the field it reads — it is not selecting low-value work over high-value work. **That is
the `dispatch_throughput` lane's finding to take forward, not this bug's**, and it does NOT explain
the blog listing (different population, different handlers). Handed to them with that separation
stated; I have not read the promoter's pre_query or bugs 444/430/454, and have filed nothing.

### RETRACTION 2026-09-02 19:1xZ — my "handler door parks 100% of the detected population" finding is WRONG (and the `dispatch_throughput` lane refuted it)

**RETRACTED:** the claim in the update above that the promoter's handler door "parks the entire
detected population by construction, because no detector populates the field it reads". **Do not
quote it. There is no bug there.**

**Verified myself at the source** (not taken on the peer's report — `scheduled_tasks.pre_query` for
`detected-item-promoter`, 94 lines):
- line 51, inside the `scored` CTE: `AND COALESCE(wi.handler_agent, '') <> ''`
- the `held` CTE comment, verbatim: *"What the doors refused. Flag-only rows (no handler_agent) are
  NOT here: they are excluded by `scored` itself, because `detected` is where they belong
  permanently and holding is not what is happening to them."*

So handler-less rows are **excluded from scoring upstream** — the doors never see them.
`head_essentials_missing` and the other 11 types are **FLAGS**: records with no automated handler,
whose permanent resting place IS `detected`. The doors govern only handler-bearing rows, and the
promoter reports those holds with reasons every tick (`held_detail`). The fork I left the peer
("is handler_agent set at detection, or filled by the promoter?") has answer **neither** — handler
assignment is a deliberate third act that turns a flag into work.

**My error, precisely:** I reconstructed the door from the peer's prose description and tested rows
against **my reconstruction** (`EXISTS(SELECT 1 FROM agent_definitions WHERE type=handler_agent…)`),
which is trivially false for an empty string. I never read the query. **The rows do not fail that
door; they never reach it.** The tell was in my own output and I walked past it: **1,386 of 1,386
failing exactly one door and 0 failing any other.** A real filter discriminates — a 100%/0% split
is the signature of having measured a **definition**, not a filter.

## CONFIRMATION FROM THE `dispatch_throughput` LANE — a better control than mine for §4's chain

They ran 21-day daily rerender-family completions (live+archive) across my five comparison sites.
**The discriminating signal is post-outage recovery**, which my snapshot could not see:

| site | 09-01 | 09-02 | recovered? |
|---|---|---|---|
| dartsonline.com | 268 | 106 | yes |
| gaswholesalers.com | 52 | 139 | yes |
| relojistas.com | 40 | 72 | yes |
| **leopardessconsulting.co.uk** | **2** | **1** | **no** |

leopardess froze at 08-28 (…89, 179, then 5, 4, 2, 1) exactly as the two-strike reading predicts,
and is **uniquely unable to recover** while its comparators did. That is much stronger than my
branded-count control, which could not separate leopardess from relojistas.

⚠ **The weekly PERIODICITY remains untestable** — the 08-28→31 trough is the fleet-wide LLM outage
(99%+ call failure; every site dips), so there is no clean sawtooth baseline in this window. **Do
not claim a sawtooth.** The 09-03 self-resume prediction is the clean test.

**They also confirmed:** branded-only sites are invisible to their per-site floor (no eligible rows,
no attempts, no losses — the 413 absence shape), so **this two-strike mechanism is now the one
absence-maker their meters cannot see**, which argues for `bugs_open/389` fixing it at source rather
than metering around it. On ">80% of a healthy site's rerender work suppressed by a brake that counts
SUCCESSES as strikes" they have no baseline but would not defend it as design — that judgement is
389's owner's.

**If leopardess resumes on/after ~2026-09-03 21:30Z, stamp the resume time into 389's evidence** —
their request, and it is the confirmation that mechanism needs.

## UPDATE 2026-09-03 09:2xZ — **THE PAGE IS REPAIRED, AND THE §4 CHAIN IS CONFIRMED** (my predicted date was wrong; the mechanism was not)

`[ALL MEASURED 2026-09-03 09:1x–09:2xZ]`

**The artefact.** `curl https://leopardessconsulting.co.uk/blog.html` → **13 of 13 card images**,
42,483 bytes (was 11 of 13 / 38,319 bytes, byte-identical across 08-31, 09-02 morning and 09-02
evening). Both guides carry their image. **The defect that kept 384 open is gone from the page.**

**When and by whom.** `page_component_history`, keyed on `page_id`:
- **2026-09-02 23:20:15** — `action:rebuild_blog_listing`, pre-image **13 articles / 2 blank**.
  **This is the repairing write.**
- 2026-09-03 00:28:14 — `action:rebuild_blog_listing` again, pre-image 13 / 0 blank (already fixed).

**THE CHAIN IS CONFIRMED — by the BRAND, which is the discriminator the handoff specified, not by
the timing.** The items that ran:

| item_key | created | branded? | outcome |
|---|---|---|---|
| `deactivated_head` | 2026-09-02 23:12:18 | **NO** | complete 23:20:18 |
| `stale_chrome` | 2026-09-02 23:12:18 | **NO** | complete 23:20:45 |
| `improvement_rerender_leopardessconsulting.co.uk` | 2026-09-02 23:22:49 | **NO** | complete 09-03 00:28:20 |

For six days every such item was born `unresolved` with the `[unresolved after N attempts]` brand.
The moment the strike count on those keys fell below 2, items were filed **unbranded**, were
dispatched, `rerender-pages` ran, and `rebuild_blog_listing` rebuilt the array. That is precisely
the §4 chain, and the growth-posture door did **not** park anything (`growth_release_recipe` absent
on every row).

**⚠ MY PREDICTED DATE WAS WRONG BY ~21 HOURS, AND THE ERROR IS INSTRUCTIVE.** I computed the
age-out from the **blog-listing key** (`page_rerender_blog_…_section_data_resolved`, second-newest
strike 08-27 22:37 ⇒ ~09-03 22:37). But that key is not what gates `rerender-pages` service. The
gating keys are `deactivated_head` (strikes 08-26 01:57:58, 08-26 21:21:48 ⇒ lifts **09-02
01:57:58**) and `improvement_rerender_…` (08-26 02:02:26, 08-26 22:01:22 ⇒ lifts **09-02
02:02:26**). Both lifted on **09-02 ~02:00**, ~21 hours before the repair.
**Right mechanism, wrong key, therefore wrong date.** I predicted the age-out of the key whose
SYMPTOM I was watching rather than the key whose SERVICE I needed. Logged in `WRONG_CALLS.md`.

**On the roll as a confound — it is not one, but I cannot fully exclude it either.** The brand
lifted ~02:00 on 09-02, **19 hours before** the 21:00Z roll, so the unblocking condition predates
the roll and the roll cannot be the cause of the eligibility change. But the new items were filed
at 23:12, **after** the roll, and I cannot prove the filing latency was rotation rather than a pod
restart. **[UNRESOLVED, and it does not matter for the verdict]**: the brand evidence is decisive on
its own, because an unbranded-then-dispatched item is exactly what the chain predicts and a pod
restart does not un-brand anything.

## WHAT REMAINS BEFORE 384 CAN CLOSE — three items, one of which is an owner decision

`[MEASURED 2026-09-03 09:1xZ]` Fleet census, blank-where-a-card-exists:

| policy | blanks | pages | assessment |
|---|---|---|---|
| generic | 5 | 2 | **all IN-FLIGHT** — designblog.co.uk/index (4 entries, cards landed 4.3h ago), oxenunity.com/tool-take-strength-scorer (1, 7.1h). Nothing stuck. |
| owned | 14 | 3 | unchanged; structural, out of this seam's reach by design |

1. **THE OWNER DECISION — does 384 close on "blog listings recover by rotation"?** The seam files
   `page_rerender` → `page-rerender`, whose workflow has **no `rebuild_blog_listing` step**, so on a
   blog listing the seam's own item still completes without rebuilding. leopardess was repaired by
   `rerender-pages` on its ordinary rotation, not by the seam. Two readings, both defensible, laid
   out in the handoff §2.
2. **The owned-page residual (14/3)** — needs its own seam (`486`'s `section_edit` route). Must not
   close inside this bug.
3. **The sweep has still never run** (`page_list_stale`, 12 items, all born terminal). It is this
   lane's artefact but the cause is `bugs_open/389`'s. Once 389 lands, the sweep needs re-validating
   and §8.1's escalation watch re-doing from zero — the old "zero escalations" figure is vacuous.

**NOT CHECKED, and it bears on item 1:** whether designblog.co.uk/index and
oxenunity.com/tool-take-strength-scorer are `rebuild_blog_listing`-maintained (slow rotation path)
or `save_page_sections`-maintained (the seam repairs them). If the former, they are a live second
instance of the gap in item 1 and should be re-read in a few hours. The query is in the handoff.

## UPDATE 2026-09-03 10:0xZ — **OWNER RULING: STAYS OPEN.** And a LIVE REPRODUCTION on the seam's own path — the defect is NOT blog-listing-specific

**OWNER RULING (2026-09-03):** *"keep it open until those are checked and fixed."* So the §2
decision in the 09-03 handoff is settled as **Option B** — 384 does not close on "recovers by
rotation". The remaining items are to be checked AND fixed, not accepted. My Option-A
recommendation is superseded; recorded here so no later session re-opens the question.

The first check found the defect live, and it **falsifies my own 09-03 §2 framing.**

### designblog.co.uk/index — reproducing NOW, on the current build, on a `save_page_sections` page

`[ALL MEASURED 2026-09-03 09:3x–10:0xZ]`

- **Not a blog listing.** Component is `content-listing` (`query.blog_posts` array field), page
  `index`, and the page is **`save_page_sections`-maintained** — the path where I said the seam
  "repairs correctly".
- **The seam fired correctly:** 5 `page_rerender` items from `derive_card_asset`,
  `reason=section_data_resolved`, `cause=card_landed:tool-aspect-ratio-guide`, created 04:56:39,
  **all complete** by 05:25:51. Two carry `consumes: ["query.blog_posts"]` — scoping matches the
  component's own field source exactly.
- **The projection inputs are all present and correct:** four target pages `status=active`,
  `page_type='blog-post'`; four card assets `status=active`, `asset_key` set
  (`card_tool_aspect_ratio_guide` etc.), `site_id` matching. Cards landed 04:56:39–05:05:10.
- **The array was rewritten twice AFTER all four cards existed** (05:06:21 and 05:25:28) and holds
  **4 entries, 4 blank**. Pre-image at each write: also 4 blank. So the re-resolve ran and produced
  the same empty images each time.

### THE CARRY HYPOTHESIS IS DEAD — the run says the section WAS re-rendered

`orchestration_states` still holds these runs (they are hours old, not days — the retention problem
that defeated the leopardess diagnosis does not apply):

| run | section_count | rerendered | carried | escalated | skipped |
|---|---|---|---|---|---|
| 2026-09-03 05:06:19 | 4 | **4** | **0** | false | false |
| 2026-09-03 05:08:24 | 4 | 4 | 0 | false | false |
| 2026-09-03 05:09:01 | 4 | 4 | 0 | false | false |

**Nothing was carried.** So `plan.Status != "ready"` (`rerender_page_sections_action.go:509`) — the
candidate I have carried since 09-02 and put in the last two handoffs — **is refuted for this case.**
The section re-rendered and the image field still resolved empty.

**Remaining hypothesis, UNTESTED and stated as such:** the `query.blog_posts` resolve returns without
populating `articles` (error, or an empty/short-circuited result), so `plan.ResolvedData` lacks the
key, `mergedContent` keeps the stored blank array, and the section still counts as `rerendered`
because it DID render HTML. That would also explain `content_data` unchanged while `rendered_html`
changed (only the html archive trigger fired). **Do not write this into a fix until it is tested.**

### What this does to the lane's state — three corrections to my own recent claims

1. **RETRACTED: "the defect is BLOG-LISTING-specific."** It is not. designblog/index is a
   `content-listing` on the `save_page_sections` path and it is broken right now.
2. **WEAKENED: "four of five demonstrations are genuine."** I verified that a `save_page_sections`
   **write happened** at 19:13/19:14/19:15 on finetuning — **I did not verify those writes produced
   non-blank images.** The 0-blank figure for them comes from the 08-26 census, not from checking
   those writes. **Re-verify before quoting §4 again.**
3. **The 09-03 handoff's §2 decision framing is superseded** by the owner ruling above AND by this
   reproduction — the choice was never "close on rotation vs fix blog listings"; the seam does not
   reliably repair its own path either.

**This is the best diagnostic opportunity this lane has had:** a live, reproducing case with the
orchestration runs still inside retention, correct inputs, and a falsified leading hypothesis.

### TEMPERING my own alarm, 2026-09-03 10:3xZ — the population is HEALTHY; what is broken is the ATTRIBUTION

I raised the designblog case an hour ago as a live reproduction. It is one, but I set the alarm too
high and the population measurement corrects it. **Both facts below are true and they matter
together.**

**The fleet is not accumulating damage** `[MEASURED 2026-09-03 10:2xZ]` — generic listing entries
whose target card exists, bucketed by card age:

| card age | entries | still blank | with image |
|---|---|---|---|
| < 12h | 26 | 5 | 80.8% |
| 24–72h | 18 | 0 | **100%** |
| older | 600 | 0 | **100%** |

**Every entry whose card is more than 24 hours old carries its image — 618 of 618.** So there is no
standing fleet-wide breakage, the blanks are a transient window, and my "reproducing NOW" framing,
while literally true, implied a persistence the data does not support. The leopardess six-day case
was the STARVED exception, not the norm.

**But the sharp question survives, and it is now sharper.** designblog/index is **still 4-of-4 blank
at 10:3xZ, with no write of any kind since 05:25:28** — five and a half hours after the cards
landed, and after **two** `section_data_resolved` re-renders that the runs themselves report as
`rerendered=4, carried=0`. So:

- something DOES repair these entries inside 24h — the 618 prove it;
- **and it is demonstrably not the seam's immediate re-resolve**, which ran twice here and produced
  blanks both times, exactly as it did on leopardess.

**That is the real defect and it is an ATTRIBUTION defect.** This lane's evidence — §4's "proven
four times on natural triggers" — credits the seam with repairs that, on both cases examined
closely, the seam did not perform. The user-visible impact is a bounded window (<24h), not permanent
damage; the correctness problem is that **we do not know what closes that window**, and this lane
has been asserting that we do.

**So the question for the `090` is not "why is this page broken for ever" — it is "what actually
writes the image, and why is it not the re-resolve that was filed to do it".** The intake fired at
09:41:56 is still `awaiting_diagnosis` (queue latency; **do not re-fire** — the trigger's own
guidance and `bugs_open/124`).

## UPDATE 2026-09-03 11:0xZ — **THE DEFECT, MEASURED: the re-resolve repairs a blank listing only ~37% of the time**

This supersedes both of today's earlier framings — it is neither "the seam never repairs" (my 10:0x
alarm) nor "the seam works" (this lane's standing claim). `[MEASURED 2026-09-03 11:0xZ]`

**Method** (stated first, because two earlier attempts at this were invalid — see the trap below):
every `page_component_history` row from the archive TRIGGER carrying an `articles` array in the last
7 days, partitioned by `(page_id, slot_name)`, with `LEAD` giving the value each write PRODUCED
(falling back to the live row for the newest write). A write "repaired" if its pre-image had blanks
and the value it produced had none. Attribution by pairing each trigger row with the
`save_page_sections_overwrite` audit row the action writes ~42 ms earlier.

**24 writes landed on a blank listing array in 7 days. 9 repaired it. 15 left it blank.**

| writer | writes over a blank array | repaired | left blank |
|---|---|---|---|
| `save_page_sections` (the seam's own path) | 21 | **8** | **13** |
| `action:rebuild_blog_listing` | 3 | 1 | 2 |

**So the re-resolve is UNRELIABLE, not absent — ~37% success over a blank array.** That is the
defect, and it explains every observation this lane has collected: pages do eventually repair (a
later write succeeds), the fleet shows no standing damage (618/618 past 24h), and yet two cases
examined closely showed completed re-renders that changed nothing.

**The failures cluster, and repeated attempts on one page keep failing:**
- `boxingonline.com/index` — 4 consecutive writes left blank (08-31 14:12, 14:21, 14:27, 16:57),
  then **repaired** 09-01 01:34.
- `advertise.co.uk/index` — 3 writes left blank (09-02 16:17, 16:31, 16:53).
- `designblog.co.uk/index` — **5 writes left blank** (09-02 20:51, 21:07, 21:11; 09-03 05:06,
  05:25) and still blank at 11:0x.
- The 8 `save_page_sections` repairs are homegarden.uk ×6 and idea.uk ×1 on **2026-08-27**, plus
  boxingonline ×1 on 09-01.

**⚠ A TEMPORAL PATTERN worth a look, stated as a LEAD and NOT a finding:** the successes are mostly
2026-08-27 and the failures mostly 08-31 → 09-03. That is consistent with a regression in that
window, and equally consistent with which sites happened to receive card landings. **Do not treat
it as a regression without controlling for site and for card-landing volume** — this lane has
already published one conclusion built on exactly that kind of uncontrolled split.

**TRAP, paid for twice in twenty minutes — read before re-measuring this.** Two prior attempts at
this table were invalid and I nearly published the first:
1. *"Pre-image had blanks"* is NOT a repairing write — it includes blank→blank rewrites, which are
   the majority here (15 of 24). You must compute the value the write PRODUCED.
2. **`page_component_history.component_id` IS NULL on these rows**, so `PARTITION BY (page_id,
   component_id)` silently collapses every slot on the page into ONE series and `LEAD` returns a
   DIFFERENT component's `content_data`. That produced a clean-looking table showing
   `save_page_sections` with 0 repairs, which is false. **Partition by `(page_id, slot_name)`, and
   only trigger rows carry `slot_name`.**
3. One write is recorded TWICE — an explicit `save_page_sections_overwrite` audit row (no
   slot_name, no html) and an `artefact_archive_trigger` row (slot_name + html) ~42 ms later.
   Counting both double-counts; the pairing is also what makes attribution possible.

**This is the sharpest statement of the defect the lane has produced, and it is the right input to
the `090` now running** (`198a7b12-f465-4cc0-a414-cec69e5f3392`): not "why did this page fail" but
**"why does the same re-resolve succeed on some runs and fail on others, on the same page, hours
apart"**.

### 2026-09-03 11:2xZ — THE QUERY IS NOT THE BUG. Ran it verbatim against live data; it returns the cards.

I reconstructed the resolver's SQL exactly as `resolvePagesWhereType` builds it — the projection
(`COALESCE(ca.asset_key,'')`), `PageImageJoinsSQL`, and `ListedPageEligibilitySQL` (the `listedOnly`
branch `blog_posts` uses: `deployed_at IS NOT NULL AND jsonb_typeof(sections)='array' AND
jsonb_array_length(sections) > 0`) — and ran it against the live database for
`designblog.co.uk` / `page_type='blog-post'` `[MEASURED 2026-09-03 11:2xZ]`:

| name | card_key |
|---|---|
| tool-aspect-ratio-guide | `card_tool_aspect_ratio_guide` |
| tool-css-unit-converter-guide | `card_tool_css_unit_converter_guide` |
| tool-css-variables-guide | `card_tool_css_variables_guide` |
| tool-smart-contrast-guide | `card_tool_smart_contrast_guide` |

**All four rows pass the eligibility floor AND carry a non-empty `card_key`.** So:

- **The SQL is correct.** The join, the projection and the eligibility floor all work.
- **The data is correct.** Verified independently earlier (pages active, cards active, `asset_key`
  set, `site_id` matching).
- **The timing hypothesis is dead.** The cards were all present by 05:05:10 and the failing run was
  05:25:27 — and the same query returns them now.

**So the defect is DOWNSTREAM of the resolver query.** The array the run persisted holds the four
correct URLs with `image=''` — the run's own `sections_metadata` shows it — while the query that
supposedly produced it returns card keys.

**The surviving hypothesis, now the only one standing:** `plan.ResolvedData` does not contain
`articles` on the failing runs, so `mergedContent` (stored ⊕ resolved, resolved winning) keeps the
STORED blank array, and the section still counts as `rerendered` because it did render HTML from
that merged content. The 8 successful repairs would be runs where `ResolvedData` did carry it.
**That makes the question "why does planSection populate `articles` on some runs and not others,
for the same component on the same page hours apart" — a conditional inside the resolve path, not
a broken query.** Still UNTESTED; `WebPath()` returning "" for a non-empty `CardKey` is the only
other survivor and is a two-line read.

This is the input the `090` (`198a7b12-f465-4cc0-a414-cec69e5f3392`, at its verdict step) most
needed and did not have when it was seeded.

### 2026-09-03 11:4xZ — the `090` FAILED (truncation), and that is a known landmine, not a new finding

`198a7b12-f465-4cc0-a414-cec69e5f3392` did **not** return a verdict. It **FAILED** at the `verdict`
step after all 5 iterations of evidence-gathering:

```
step verdict failed: execute_llm_prompt: AI call failed with unhandled error:
response truncated: stop_reason=max_tokens (output_tokens=32000 reached the
configured cap, 3440 chars recovered)
```

**So the cost of a truncated verdict is the WHOLE RUN** — five iterations of bundles, ~30 minutes
and the credits, discarded for a cut answer. Worth knowing before anyone re-fires: this failure mode
is not "a shorter answer", it is "no answer".

**This is `LANDMINES.md`'s documented trap, already there since 2026-07-31** — *"A TRUNCATED LLM
call has `output_tokens = NULL`…"*. I re-derived it from scratch before finding it (logged in
`WRONG_CALLS.md`: the `SessionStart` hook cannot match a **table** footprint, so it never surfaced,
and I did not grep). **Reconfirmed and appended there rather than filed as new**
`[MEASURED 2026-09-03]`: the documented `output_tokens >= max_tokens` census returns **4** while
`error_message ILIKE '%stop_reason=max_tokens%'` returns **74** over 14 days — still ~5% visibility.
Today's offenders include **`council-gate` `review_debug_historian`, truncated 17 times at cap 8000**,
which is a gate seat silently losing its review on the mechanism CLAUDE.md tells every session to use.
**Not this lane's to fix; flagged where it belongs.**

**For this bug it changes nothing about the diagnosis and one thing about the plan:** the `090`
route has now produced UNVERIFIABLE once and FAILED once on this question. **A third attempt should
not be a re-fire of the same shape.** The first run's own missing-evidence list, plus this session's
two narrowings (the query is exonerated; `ResolvedData` is the suspect), are a better starting point
than another cold run — and the remaining question is now small enough to answer by reading
`planSection` against a failing and a succeeding run, which is a code read, not a loop.

## UPDATE 2026-09-03 12:4xZ — **THE ~37% FIGURE IS RETRACTED.** The suspect was right, the cause is `bugs_open/454`, and the seam's own path repaired **132 of 132** before that regression landed

This supersedes every framing published earlier today, including my own 11:0x "the re-resolve is
UNRELIABLE (~37%)" and the 10:3x "ATTRIBUTION defect". Both were wrong, in different ways, and the
corrections are below with the measurements that caught them.

### 1. The surviving hypothesis was RIGHT, and someone else had already found the mechanism

11:2x left exactly one hypothesis standing: *"`plan.ResolvedData` does not contain `articles` on the
failing runs, so `mergedContent` keeps the STORED blank array while the section still counts as
`rerendered`."* That is **correct, and it is `bugs_open/454`** — filed 2026-09-03 10:00Z by the
`bugs_open/427` lane, ~90 minutes before I wrote that paragraph, while I was measuring.

`classifyStoredSection` computes `plan := planSection(...)`, branches on `plan.Status`, and returns
**without assigning `c.plan`**. `renderPlannedSection` reads `cls.plan` and gets the zero value, so
`plan.ResolvedData` is nil for *every* section of *every* light re-render. Not conditional, not
`articles`-specific, and not a query-path bug — which is why my exoneration of the resolver SQL at
11:2x was sound and led nowhere on its own.

- **Introduced by `94f81cc60`, 2026-09-02 11:27:53Z** ("035 P1: extract classifyStoredSection").
- **Fixed by `9831e9ab4`, 2026-09-03 10:00:40Z.** Council APPROVED (`075cfedd`), round 1.
- **LIVE since ~12:05Z today** in chassis `d0252fd4dab2a3a583d1cc8eb8e1b26e9c422d85` (v1.0.1358)
  `[MEASURED 2026-09-03 12:33Z]` — read from `service_binary_capabilities` (kind='build',
  name='provenance'), not from the startup log line, which had already scrolled. The 427 lane
  proved it live independently at the artefact (`6f3116af0`).
- **`WebPath()` is exonerated too** — the other survivor, and a two-line read as predicted:
  `DeployedWebPath` → `DeployedAssetPath` returns a `RelativeURL` built from the asset key, and
  **cannot return "" for a non-empty `CardKey`** (`platform/storage/url_helpers.go:317-347`).

**Do not fix any of this here.** 454 is owned by the `bugfix_427_event_render` lane, fixed,
approved and live. 384's job was to say whether the seam is sound underneath it.

### 2. The `~37%` was a MEASUREMENT ARTEFACT — I did not join the card

`[MEASURED 2026-09-03 12:1x–12:3xZ]` The 11:0x census counted an entry as a failure whenever
`image` was `''`. **An entry whose target page has no card asset is *correctly* blank** — the
resolver has nothing to project. So the census scored the resolver's correct behaviour as failure,
and did it more often in the later window because that is where the un-carded pages were.

This is my own RUNBOOK's documented gotcha, one section above the query I wrote:
*"a first cut keyed `empty image` over ALL sources and showed news/directory arrays as '20/20
empty' … **Join the card, don't count empties**."* I applied it to the pair census in August and
not to this one.

**Both shapes, over the identical rows, same 7-day window, `articles` only:**

| measure | writes over a blank array | repaired | left blank |
|---|---|---|---|
| bare `image=''` (the 11:0x shape) | 19 | 5 | 14 |
| card-joined (entry's page had an active card **before** the write) | **8** | **6** | **2** |

### 3. What the corrected census actually says — and it splits cleanly at the regression

Same method as 11:0x (archive-TRIGGER rows only, `PARTITION BY (page_id, slot_name)`, `LEAD` for
the value each write PRODUCED), plus two fixes: restricted to `query.*`-sourced array fields, and
an entry counts as a deficit only if its target page had an **active card created before that
write**. 10-day window. `[MEASURED 2026-09-03 12:2xZ]`

| era | writer | writes over a real deficit | repaired | left blank |
|---|---|---|---|---|
| **before** `94f81cc60` | `save_page_sections` | **131** | **131** | **0** |
| **before** `94f81cc60` | `action:rebuild_blog_listing` | 1 | 1 | 0 |
| **after** `94f81cc60` | `save_page_sections` | 11 | 4 | **7** |
| **after** `94f81cc60` | `action:rebuild_blog_listing` | 1 | 1 | 0 |

**132 of 132 before the regression. Zero failures.** And in the post-regression window the
attribution is total `[MEASURED 2026-09-03 12:3xZ]` — each write joined to the last orchestration
on its page within 20 minutes:

| attributed to | writes | repaired | left blank |
|---|---|---|---|
| `page-rerender`, `reason=section_data_resolved` (**the light re-render**) | 7 | **0** | **7** |
| `page-build-handler` (**a full build**) | 1 | 1 | 0 |
| no run found in window (full-build chains, keyed differently) | 4 | 4 | 0 |

**Every light re-render failed. Everything else repaired.** That is 454 exactly: the full build
goes through `plan_sections` normally and is untouched, the light re-render drops its plan.
designblog/index sat blank for 7 hours because it is a listing page that only ever receives the
LIGHT re-render — no image lands on `index` itself, so no full build ever came to rescue it, unlike
its three sibling `tool-cta` slots which failed at 05:08–05:09 and were repaired at 05:42–05:50 by
`page-build-handler` runs (`reason=image_landed`) 5 minutes earlier.

### 4. Three corrections to today's earlier claims, and what each of them cost

1. **RETRACTED: "the re-resolve repairs a blank listing only ~37% of the time" (11:0x).** The
   figure came from an uncarded denominator. Corrected figure: **100% (132/132) before 454's
   regression; 0% (0/7) after it, for a reason that is not this bug.**
2. **RETRACTED: "the defect is an ATTRIBUTION defect — this lane credits the seam with repairs it
   did not perform" (10:3x).** That framing rested on exactly **two** closely-examined cases,
   leopardess (09-02 23:20) and designblog (09-03 05:06–05:25). **Both lie inside 454's regression
   window**, which opened 09-02 11:27Z. A two-case sample drawn entirely from another lane's
   19-hour-old regression cannot say anything about the seam's own record, and I published a
   conclusion about the lane's whole evidence base from it. §4's "proven four times" is **not**
   falsified; it is un-recheckable at this distance (see the limit below), and nothing now
   contradicts it.
3. **STANDS: the leopardess starvation (08-26 → 09-02) and its resume.** That is `bugs_open/389`'s
   two-strike arm, is pre-regression, and is unaffected by any of this. The seam was never given
   an item to run; it was not failing.

### 5. Two limits on the above — state them wherever these numbers are quoted

- **The pre-regression 132 cannot be attributed to a code path.** `orchestration_states` holds
  **25.0 hours** of history (oldest run 2026-09-02 11:44Z) `[MEASURED 2026-09-03 12:4xZ]`, so every
  pre-regression write's run is aged out. "132/132 repaired" is an OUTCOME over whatever mix of
  light re-renders and full builds was operating. It is strong evidence of **no standing failure**;
  it is **not** proof the light re-render did the repairing. The post-regression attribution table
  is the one that discriminates by path, and it only covers 12 writes.
- **The post-regression failure count is a LOWER BOUND.** The archive triggers fire only when
  `rendered_html` changes, or when `content_data` changes and `rendered_html` does not
  (`trg_page_component_artefact_archive_upd` / `trg_page_component_content_archive_upd`). A write
  that changed **neither** leaves no history row at all — and a byte-identical no-op is precisely
  what 454 produces. So this census can only see the 454 failures that happened to move some
  bytes; it cannot see the ones that moved none.

### 6. What this does to 384's remaining work

The owner ruling of 2026-09-03 ("keep it open until those are checked and fixed") stands, and the
list in §3 of `HANDOFF_2026-09-03_continue_here.md` is unchanged by any of this — **except** that
the item that looked biggest this morning, "the seam repairs only ~37% of the time", was not a 384
item at all. What remains genuinely owed:

1. **Prove the seam repairs on its own path, post-fix.** A `section_data_resolved` re-resolve at
   designblog.co.uk/index, filed 12:35:51Z as `created_by='bugs_open/384_postfix_verify'`
   (item `80a1c536-b75f-416d-ac72-952177229b5c`, item_key `…_section_data_resolved_384verify`).
   The page is `page_type='landing'`, `rebuild_policy='generic'`, so **450's `pageRefusesGenericBuild`
   does not fire on it** — checked before filing, because a refusal would have read as a 384 failure.
   Success = the `content-listing` `articles` array goes 4-blank → 0-blank.
2. **The owned-page residual (14 blanks / 3 pages)** — unchanged, still structurally out of this
   seam's reach, still must not close inside 384.
3. **The sweep (`check_page_list_stale`, migration 603) has still never run** — unchanged, still
   blocked behind 389's arm.

## UPDATE 2026-09-03 13:0xZ — **THE SEAM'S OWN PATH IS PROVEN ON THE FIXED BINARY.** designblog.co.uk/index: 4 blank → 0 blank, verified at the served page

`[ALL MEASURED 2026-09-03 12:54–13:0xZ]` The verification filed at 12:35:51Z has run and repaired.

| | before | after |
|---|---|---|
| `content-listing.articles` blank images | **4 of 4** | **0 of 4** |
| `page_components.updated_at` | 2026-09-03 05:25:29Z | **2026-09-03 12:54:43Z** |
| `rendered_html` | 2,494 B | **3,327 B** |
| served `designblog.co.uk/` | — | HTTP 200, **4 card `<img src>`, zero `src=""`** |

- **The run:** `page-rerender`, `reason=section_data_resolved`, COMPLETED 12:54:41Z —
  `section_count 4, rerendered 4, carried 0, escalated false`. Item
  `80a1c536-b75f-416d-ac72-952177229b5c` went `complete` at 12:54:53Z, `attempt_count 0`.
- **The four card files resolve:** `card-tool-aspect-ratio-guide.jpg` 200/35,559 B,
  `…css-unit-converter-guide` 200/53,603 B, `…css-variables-guide` 200/62,028 B,
  `…smart-contrast-guide` 200/35,984 B. So this is a repair a visitor can see, not a row.
- **No `bugs_open/450` confound:** `page_type='landing'`, `rebuild_policy='generic'`, checked
  before filing — neither arm of `genericBuildRefusal` fires, so the fix ran through
  `save_page_sections` cleanly. (The 427 lane's own post-fix case was refused at the save by that
  guard, so it had a proven re-render and no proven artefact. This one has both.)

**⚠ THE RUN'S COUNTS ARE NOT THE PROOF, AND MUST NOT BE QUOTED AS ONE.**
`rerendered=4, carried=0, escalated=false` is **byte-for-byte what the BROKEN runs reported** —
compare the 10:0xZ table above, where 05:06:19 reads `4 / 4 / 0 / false` and produced four blanks.
That is 454's whole signature. The proof is the **artefact** plus a baseline someone else read:
the `bugs_open/427` lane independently recorded this row at 12:54Z as *"4 articles, 0 with images,
updated_at still 05:25:28Z, html 2,494 bytes"* — seconds before the write landed. Their read is the
control, and it was taken by a different session with no stake in this passing.

### What this settles, and what it does not

**Settles:** the 384 seam's own path — `derive_card_asset` → `page_rerender`/`section_data_resolved`
→ `rerender_page_sections` → `save_page_sections` — **re-resolves a stale page-list array and lands
the card images, on a page that had been blank for 7.5 hours across two earlier attempts.** Item 1
of the remaining-work list is closed. Combined with §1's 132/132 pre-regression record, the
mechanism this workstream built is sound and the 09-02→09-03 breakage was `bugs_open/454`.

**Does NOT settle** — and the `components` lane (`bugs_open/425`) is right to flag it: this is a
positive for the **`image` field only**. That deck already carried `articles[0].excerpt` in every
archived state back to 09-02 20:51, put there by a BUILD, so on the *item-shape* axis the baseline
could not have failed and this run proves nothing about it. My control was the blank images and the
claim is scoped to them. **The general trap they paid for and I did not: "a re-render wrote a row
carrying key K" is not "the re-render produced K" — a stored ⊕ nil merge carries K forward
unchanged.** The discriminator is to project the state each write REPLACED:
`SELECT h.created_at, (h.content_data->'articles'->0 ? '<key>') AS present_BEFORE_this_write FROM
page_component_history h WHERE h.page_id = $1 ORDER BY h.created_at;` — joined on `page_id`, never
`component_id`.

## UPDATE 2026-09-03 15:2xZ — the standing residual, RE-MEASURED with the card joined: **generic is clean, owned is 14/14 and stale by 22–48 days**

The "14 blanks / 3 pages" owned figure was measured with the same uncarded shape as the retracted
~37%, so it had to be re-checked before it could be carried anywhere. **It survives, and it is
worse than a rate — it is total.** `[MEASURED 2026-09-03 15:2xZ]` every entry in a live
`query.*`-sourced listing array whose target page has an active card:

| `rebuild_policy` | carded entries | still blank | pages | with image |
|---|---|---|---|---|
| `generic` | 640 | **1** (in-flight) | 1 | **99.8%** |
| `owned` | 14 | **14** | 3 | **0.0%** |

- **The one generic blank is NOT damage.** advertise.co.uk/index, card landed **0.1 h** ago, array
  last written 22 minutes ago — i.e. **written before the card existed**. Its re-resolve has not
  come round yet. That is the transient window this lane has measured all along, caught mid-flight.
- **The owned 14 are all `tool-cta` slots on three pages**, and their arrays have not been rewritten
  in **three to seven weeks**:

| site | page | array last written | card age |
|---|---|---|---|
| leopardessconsulting.co.uk | `llm-cost-calculator` | **2026-07-17** | 548 h |
| leopardessconsulting.co.uk | `tool-ai-vendor-trust-checklist` | **2026-07-30** | 548 h |
| finetuning.uk | `llm-cost-calculator` | **2026-08-12** | 598 h |

**This is the shape that distinguishes it from everything else in this file.** A generic page's
blank is a window measured in hours and it closes by itself; an owned page's blank is permanent,
because `save_sections` refuses the page and no other writer touches that array. 22–48 days is not
a slow repair, it is no repair. **It is still NOT 384's to close** (the seam is working as
designed — it declines to overwrite a customer-owned page), but the figure is now card-joined and
therefore quotable, and it names the exact three pages and the `tool-cta` slot for whoever takes
the `section_edit` → `section-editor` route (migration `486`).

Detail query: `scripts/residual_by_policy.sql` in the workstream directory.

## UPDATE 2026-09-03 15:4xZ — the enumerated exception to "every light re-render now delivers", **corrected**: it is 2 rows, not 14, and `component_id IS NULL` is NOT the test

The `components` lane (`bugs_open/425`) flagged a shape my census is **structurally blind to** and
which would otherwise read as a 384 residual: a `page_components` row whose `component_id` is NULL.
My census joins `page_components pc JOIN qf ON qf.component_id = pc.component_id`, so such a row is
**excluded from every figure in this file**. That part of their warning is right and important.

**But the screening rule they offered — "`component_id IS NULL` ⇒ can never be repaired" — over-flags
by 6×, and I checked the code rather than adopting it.** `resolveComponent`
(`rerender_page_sections_action.go:361-393`) does NOT give up on an empty `componentID`: it falls
through to `schemas[s.slotName]`, and `loadComponentSchemas` (`plan_sections_action.go:1981-2002`)
indexes **by BOTH `name` AND `function`** — *"Index by both name and function for fast lookup in
the section loop."* So a NULL-id row resolves fine whenever its `slot_name` matches either.

`[MEASURED 2026-09-03 15:4xZ]` all 14 live NULL-`component_id` rows, tested against the map the
code actually builds (`cc.name = slot_name OR cc.function = slot_name`, `is_active`):

| resolves? | rows | pages | slot names |
|---|---|---|---|
| **yes** | **12** | 6 | `blog-listing`, `generic-text-block` ×8, `faq`, `tool-funding-fit`, `tool-loan-vs-savings` |
| **no — genuinely stranded** | **2** | 2 | `article-grid` (finetuning.uk `/blog`), `section` (gamesdesign.co.uk `/game-jelly-invaders`) |

**The trap in the middle is that `name` and `function` are different columns and the slot names use
the FUNCTION one.** `blog-listing`, `generic-text-block`, `faq`, `tool-funding-fit` and
`tool-loan-vs-savings` all have **0 rows** in `content_components` by `name` and ≥1 by `function`.
A screening query written the obvious way — `WHERE cc.name = pc.slot_name` — therefore returns
"unresolvable" for every one of them. **I wrote exactly that query first and it said 14 of 14 were
stranded**; `content-listing` (which I knew resolves, because I had just watched it repair) came
back false too, which is the only reason I looked again. **Keep a known-good control in any
resolution census, or this returns a clean, plausible, entirely wrong answer.**

**So the correct screening rule for a stuck page after the 454 fix:**
```sql
pc.component_id IS NULL
AND NOT EXISTS (SELECT 1 FROM content_components cc
                 WHERE (cc.name = pc.slot_name OR cc.function = pc.slot_name) AND cc.is_active)
```
Two pages match today, neither of them a listing this seam feeds. Separately — and this half of the
lane's report stands unchallenged — boxingonline.com `/articles-index` carries **six** stacked
`generic-text-block` rows all at `position 3`, which is `bugs_open/457`'s orphan-append and is why
that page serves 36 cards where there should be 6. Those rows resolve; the defect is that they
exist, not that they cannot render.

## UPDATE 2026-09-03 15:3xZ — **17/17 after the fix**, completing the partition; plus a correction to my own page count and a population that grows by the hour

### The third era, from the `components` lane (`bugs_open/425`)

They drained their whole class post-fix and it went **17 new shape / 0 old at 15:27Z** (5/12 the
previous day, 9/8 at 15:05, 11/6 at 15:18). Six of those were a batch they filed with baselines
recorded **before** dispatching; each confirmed `excerpt` absent → present from its own archiving
write. So this seam's record now reads across all three eras:

| era | writes over a real deficit | outcome |
|---|---|---|
| before `94f81cc60` | 132 | **132 repaired, 0 blank** |
| during 454's regression | 7 light re-renders | **0 repaired, 7 blank** |
| after `9831e9ab4` | 17 (their class) + 1 (designblog) | **18 repaired, 0 blank** |

Two of their details are worth keeping. `idea.uk /guides-index` came back with **9** items where the
frozen array held 7 — the resolver returning more eligible posts than the stale snapshot, i.e. the
mechanism working, not a discrepancy. And **four of the six repairs moved `rendered_html` in both
directions** — one of their earlier canaries *shrank* 928 bytes on a clean repair (stripped
site-name suffixes and collapsed empty elements outweighing added text). **So "html grew" is not a
pass condition either**, and my 12:5xZ table quotes 2,494 → 3,327 B: that byte figure is
descriptive, NOT part of the verdict. The verdict is the blank count and the served page.

### ⚠ CORRECTION to my own 15:4xZ table — the page count was wrong, and the population is MOVING

`[MEASURED 2026-09-03 15:30:39Z]`, against `[MEASURED 15:4xZ]` forty minutes earlier:

| | at 15:4xZ (as I published it) | at 15:30:39Z | correction |
|---|---|---|---|
| resolving rows | 12 | **13** | population grew |
| resolving **pages** | ~~6~~ | **6** | **wrong when written — it was 5** |
| stranded | 2 rows / 2 pages | 2 rows / 2 pages | unchanged |
| **total NULL-id rows / pages** | 14 / 7 | **15 / 8** | grew |

**The page count was a miscount and I own it** — I counted the resolving pages by eye off a 14-row
listing instead of asking the database, and got 6 where the distinct count was 5. The peer lane's
independent census is what surfaced it (they read 7 total pages against my 8). One `count(DISTINCT
page_id)` would have settled it; eyeballing a short list feels too cheap to be worth a query and is
exactly where this kind of error lives.

**And the population is not static — it grows by the hour.** A new NULL-`component_id` row was
created at **15:27:34Z**, three minutes before I read it: advertise.co.uk
`/tool-cpm-cpc-benchmark-comparator`, slot `tool-cpm-cpc-benchmark-comparator` (it resolves).
`bugs_open/457`'s orphan append is producing these continuously, so **both my count and the peer's
were correct when taken and stale within the hour.** CLAUDE.md's rule is that a count of things
carries the date it was counted; on this population **a date is not enough — it needs the time**.

**The one figure that has not moved, and is the sharpest form of the `name`/`function` trap:**
`[MEASURED 15:30Z]` of all 15 NULL-id rows, **zero match an active component by `name`** — every
one of the 13 that resolves does so by `function`. So a screening query joined on `cc.name` returns
**100% stranded on this population, every time, whatever its size.** It cannot come out right.

### One inference from that lane, carried as theirs and UNTESTED

`[INFERRED, not measured — theirs]` a re-render of boxingonline.com `/articles-index` may still fail
at the save: migration `316`'s `uq_page_components_no_byte_identical_duplicate` refuses a row
byte-identical to one already present, and six rows rendering the same deck from the same data is
that shape — which is `457`'s own reported failure mode. **Do not build on it without testing it.**
It does not change this seam's record; it bounds what a re-render can do for that one page.

## CORRECTION 2026-09-03 15:5xZ — **I attributed the growing NULL-id population to `bugs_open/457` and that is WRONG.** 457 has appended nothing in 23 hours; the growth is the ordinary save path

The 15:3xZ update above says *"`bugs_open/457`'s orphan append is producing these continuously"*.
**It is not.** The `components` lane chased the 15:27:34Z row independently, and I have now verified
their account at both the code and the data rather than adopting it.

**At the code — two different INSERTs, and they are not confusable once read:**

| | `rebuild_blog_listing_action.go:403-407` (457) | `save_page_sections_action.go:1124-1127` (the save path) |
|---|---|---|
| position | **hard-coded `3`** | **`i+1`** — the section index |
| `component_id` | **column absent from the INSERT entirely** | `$5` = `componentIDPtr`, a POINTER — NULL when the section metadata carries no id |
| fires when | a blog-index page and `findBlogListingSlot` misses | every ordinary section save |

**At the data** `[MEASURED 2026-09-03 15:5xZ]` — the population splits cleanly on that signature:

| producer | rows | pages | oldest | newest |
|---|---|---|---|---|
| 457's append (`position 3` + `generic-text-block`) | 6 | **1** (boxingonline `/articles-index`) | 2026-08-31 16:29:48 | **2026-09-02 16:28:02 — nothing in 23 h** |
| the save path | 9 | 7 | **2026-03-16** 14:46:47 | 2026-09-03 15:27:34 |

And the decisive one: **all five rows on advertise.co.uk `/tool-cpm-cpc-benchmark-comparator` were
created inside 0.2 s** (15:27:34.043 → .248), positions 1–5, and **only position 5 lacks its id**.
Four healthy siblings. That is one full page build writing a NULL `componentIDPtr` for one section,
not an append onto an existing page. 457's rows arrive alone, at position 3, on one page.

**What survives, unchanged:** the population IS growing, both censuses were stale within the hour,
and on this population a count needs the **time** and not just the date. Only the attribution was
wrong — and 457's own file predicts its silence, because the action now hard-fails on migration
`316`'s duplicate guard before reaching that insert.

`[UNVERIFIED — theirs, and not this lane's]` why those tool-page sections' metadata lacked a
component id at all. Two older rows of the same shape are already in the population (idea.uk
`/tool-funding-fit` 09-02 12:27, loanzy.uk `/tool-loan-vs-savings` 08-28 07:33), both tool pages
whose slot is named after the tool. That belongs to whoever owns the tool-page build.

**And the three-era table's "after" column is still moving** — quote it as **18/18 as of 15:27Z**,
not as a settled figure. Three of those repairs were dispatched by nobody at either bug.

## UPDATE 2026-09-03 16:0xZ — the census's DENOMINATOR is now verified, not assumed: the assemble path writes nothing at all

Relayed by the `components` lane and **re-measured here before recording**, per the lesson two
sections up — this is a claim I agreed with, which is exactly why it needed grading.

`[MEASURED 2026-09-03 16:0xZ]` robot-hands.com, over the window in which **39 assemble-mode
`page_rerender` items completed** (14:21:30 → 15:13:36): **zero `page_component_history` rows, any
source.** The first row is **15:14:34**, `save_page_sections_overwrite` on `learning-center-hub` —
their `template_changed` canary. So one page, one hour, two paths, opposite results: an assemble
re-render on `learning-center-hub` at 14:27:36 left no trace; the re-resolving re-render on the
same page 47 minutes later wrote six rows.

**The demand control, and mine is stronger than the one they used.** They controlled with three
later same-site re-renders that did write. Better: the table processed **174 rows fleet-wide inside
that same window** while robot-hands wrote zero. So the absence is not an idle table, an outage, or
a quiet hour — it is that path writing nothing while the table was busy.

**What this does for every figure in this file.** My census counts rows in
`page_component_history`, so its denominator is **exclusively the re-resolving modes**
(`image_landed`, `section_data_resolved`, `cta_links_stale`, `template_changed`,
`literal_markdown` → `rerender_page_sections`). Assemble-mode re-renders re-ship stored HTML and
are **structurally absent**. That is the CORRECT population for the question this lane asks — "when
a re-resolve ran over a blank listing, did it repair it?" — and it is now verified rather than
assumed, on a site neither lane chose for the purpose.

⚠ **But do not read "132 writes in 10 days" as "the seam ran 132 times."** It ran far more often;
132 is the count of *re-resolving* runs that landed on a real deficit **and moved some bytes**. Two
things are invisible: the assemble path (by design, it cannot repair) and any byte-identical no-op
(the archive triggers fire only on a change). The first is correctly excluded; the second is the
lower bound already stated at 12:4xZ.
