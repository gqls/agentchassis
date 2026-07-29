# NOTES — dartsonline.com traffic & affiliate-readiness

Append-only, newest at the bottom. Technical log: evidence, commands, what the system
actually said, and every misstep.

---

## 2026-07-29 — session 1 (workstream opened)

### Exploration corrections (things the brief and my own first pass got wrong)

Six claims I formed early and then had to correct. Recorded because each was locally
plausible and none survived a query.

1. **"The nav 404s mean three landing pages need building."** WRONG. `shop`, `brands`,
   `guides` are orphan `pages` rows from SUPERSEDED plans. The current plan
   (`site_plans.is_current`) contains `shop-index`, `brands-index`, `guides-index`
   (role `section-index`) and no landing rows at all. The hubs replaced them; the old
   rows kept `in_header=true` and `in_footer=true`. What caught it: querying
   `site_plan_pages` for the current plan instead of trusting the `pages` table alone.

2. **"`detected` work items can never dispatch, so `check_undeployed_assets` needs a
   2-line fix."** WRONG, and this one nearly became a code change. `triage_detected_items`
   (`platform/orchestration/actions/triage_detect_items_action.go`) promotes EVERY
   detected item to `triaged`/`build`; three live agents call it (site-review-agent,
   design-audit-agent, improvement-loop). The items are stranded because
   `scheduled_tasks.improvement-sweep` is `enabled=false` (last triggered **2026-05-02**),
   not because the check writes the wrong literal. And `bugs_open/083`'s owner has
   already gone further: *"routing is NOT the bottleneck … there is no reader for ANY of
   the queues"* — 325 items sit in `needs_human_review`, the queue humans CAN see, oldest
   2026-03-15. That file ends **"Decision pending — do not act on this section until it
   is recorded here."** So the planned generic fix is cancelled: it would have been a
   routing fix that its own bug file argues against, on someone else's open decision.
   Site-scoped SQL promotion for dartsonline is data, not mechanism, and stays in scope.

3. **"The nav-404 class is a generic framework defect."** WRONG — the framework already
   fixes it. `GetNavItems` (`nav_tables.go:215-240`) drops nav items whose target has
   never been deployed. The fleet measurement is decisive:
   ```sql
   SELECT s.domain, count(*) FROM pages p JOIN sites s ON s.id=p.site_id
   WHERE p.in_header AND p.deployed_at IS NULL AND p.status IN ('active','deployed','pending')
   GROUP BY 1;
   -- dartsonline.com | 4      <- one site, nothing else fleet-wide
   ```
   dartsonline serves STALE STORED CHROME (`bugs_open/117`). Data fix + chrome rebuild,
   no platform change. **The lesson is the order:** I measured the fleet before writing
   the fix, and the measurement removed the fix.

4. **"The fabricated identity is on the About page."** Partly wrong, and the important
   half was invisible from the page. The live page has NO Portland, NO brand names, NO
   darts.com contact (all greps 0). It DOES claim "We stock across the full range" ×2 and
   "We carry the manufacturers…" ×1. The Portland address, `sales@darts.com`,
   `(800) 526-1920`, the seven brands and the AUSTRALIAN company's Facebook URL
   (`facebook.com/dartsonlineau`) all live in the `identity` and `briefing` SPECS — i.e.
   in the source every future rebuild would draw from. Fixing only what was visible would
   have let the next rebuild reintroduce all of it.

5. **"S3 credentials block the imagery work."** WRONG for this site — that blocker is
   about writing hand-made exact bytes for three OTHER sites. All 33 dartsonline assets
   already serve HTTP 200.

6. **"The tool-imagery hold gates the setup-builder."** STALE — `bugs_closed/020` closed
   2026-07-23 and the owner lifted the hold 2026-07-24.

### The `gaps`-array trap (my own detector poisoned by my own text)

After the briefing reset my verification query still reported `portland: t` and
`stock_words: t` on the NEW row. Two different causes, and only one was a problem:
- `honesty_rails` (a key I had just added) contains *"Never claim to stock, carry…"* —
  the prohibition matches the pattern for the thing it prohibits. This is exactly
  [[prompt-text-poisons-its-own-detector]] and it bit inside five minutes of my writing it.
- `gaps` still read *"number sourced from associated Portland operation"*. That WAS a
  real leak risk: the phone had been replaced, so the sentence was untrue-of-now, and any
  writer reading the whole briefing blob into a prompt could re-contaminate copy from it.
  Rewrote `gaps` to current reality; provenance kept in the row's `notes` column and the
  backup table, neither of which is fed to a prompt.

**Transferable:** when a fix ADDS prohibition text, a substring check over the whole blob
can no longer distinguish "claim removed" from "claim named in order to forbid it". Check
per-key (`jsonb_each`), not over `data::text`.

### The gap that actually shipped the fabrication

`briefing.gaps` recorded, on 2026-07-06, that the phone was *"sourced from associated
Portland operation; not confirmed"* and that there were no *"confirmed brand partnership
or stock relationship details beyond signal-level inference"*. The research was honest
about its uncertainty **and the build rendered the claims as fact anyway.** The defect is
not bad research — it is that nothing between "recorded as unverified" and "rendered as
prose" reads the uncertainty. Worth a bug file of its own if a second site shows it.

### Applied this session (all live, all reversible via the bak_ tables)

| what | file | result |
|---|---|---|
| identity truth reset | `SQL_2026-07-29_identity_truth_reset.sql` | 1 current row; portland/stocking/AU-facebook all false; old row preserved |
| briefing truth reset | `SQL_2026-07-29b_briefing_truth_reset.sql` | about_us + services rewritten; `headquarters`/`location` keys dropped; `honesty_rails` added |
| stale `gaps` rewrite | inline (recorded above) | 0 keys mention Portland |
| nav reconciliation | `SQL_2026-07-29c_nav_reconciliation.sql` | 3 orphans archived; setup-builder out of nav; 3 hubs in with clean labels; 13 polluted nav_labels trimmed |

Backups: `bak_darts_identity_20260729`, `bak_darts_briefing_20260729`,
`bak_darts_pages_nav_20260729`.

### Verified facts worth not re-deriving

- Work items open: 17 `undeployed_asset` (detected/design), 8 `needs_page` (NHR),
  4 `needs_rerender`, 3 `page_rerender`, 3 `deactivated_component`, 3 `unresolved_cta`,
  1 each `capability_gap` / `empty_section` / `evaluate_tools` / `needs_section_data` /
  `owned_page_review`. (Exploration said 5 `unresolved_cta`; the live count is 3.)
- `bugs_closed/141` is CLOSED and proven live on v1.0.1198 (`9db57e426`) — news-index
  can enter nav. No pre-check needed before creating the news page.
- `check_missing_tools` (`discovery_checks/check_missing_tools.go`) already exists and is
  the natural home for the owner's 1-tool-per-6-articles ratio (D5). Today it is purely
  TIME-based: 7-day cooldown at 0 tools, 30-day at 1+. It counts deployed tools via
  `content_components.component_level='tool'`. It emits `evaluate_tools` →
  `tool-suggester` at `Status:"detected"` — same stranding as everything else.
- `loadSiteContactEmail` (`validate_page_content.go:1274-1298`) reads FLAT
  `identity.data->>'email'` / `->>'contact_email'`; `sync_site_identity_action.go:103-110`
  writes NESTED `identity.contact.email`. Both shapes now written (bug-072 class).
- `populate_nav_tables` DELETEs and rebuilds nav from `pages` — hand-editing
  `site_nav_items` is pointless, `pages.in_header` is the control surface.

---

## 2026-07-29 — session 1 continued: the third source of the same lie

### The nine guides were unblocked by a data fix, and it worked first time

`site_plan_sections` had zero rows for all nine blog-posts, and `pages.sections` was
`'[]'`. Inserted the canonical article layout — `["hero","article-body","call-to-action"]`,
taken from `create_blog_posts_action.go:183` rather than invented — into the PLAN table
(the authority; `pages.sections` is documented in `load_page_sections_from_spec_action.go`
as the materialised cache, and the loader syncs down to it itself).

Promoted two items only. **barrel-weight became the first guide page ever built on this
site** (13:32Z, 3 components, 9,039 bytes), then beginners. Both `complete`.

**Misstep worth recording:** while the first build was running I read the work item's
`error` column and saw *"page-build-handler no-op: no sections ready to build"* — the
exact failure the backfill was meant to cure — and briefly concluded the fix hadn't
worked. It had. `error` holds the text of the LAST failure and is not cleared when a row
is re-claimed; `beginners` still showed a `claimed_at` of 2026-07-20 next to it. The
authority on "is it working" was `orchestration_states`, which showed the run stepping
through `deploy_page` to `complete`. **A stale error column reads exactly like a live one.**

### No blog-index page was created — a deliberate departure from the plan

The plan said to hand-create one. `guides-index` (`/guides/index.html`) is already
deployed and already carries `content-listing`, whose `input_schema` sources
`query.blog_posts` — the same resolver a blog index would use. It IS the guides index;
it lists nothing only because no guide had ever been built. A second listing page would
have duplicated it and split internal links.

### THE FINDING: fixing the identity specs was necessary and NOT sufficient

barrel-weight came back well-written, on-voice, spec-accurate — and its call-to-action
read **"Filter our ranges by weight and tungsten percentage."** There are no ranges.

This was not the writer hallucinating. `content_direction` — untouched by the morning's
truth reset — instructs it to write shop copy:

| key | text |
|---|---|
| `writing_rules[0]` | "…on **product listings**" |
| `writing_rules[4]` | "Keep CTAs action-first: **'Add to Bag'**, 'Pick Your Weight'" |
| `writing_rules[7]` | "**Price copy** should be direct and confident: show savings…" |
| `writing_rules[8]` | "**Brand pages** should… not just list SKUs" |
| `persuasion_approach.method` | "Position the **store** as a trusted guide" |
| `content_depth.thoroughness` | "**Product pages** go deep on specs…" |

So the false premise had **three** homes — `identity` (who we are), `briefing` (what the
About page says), and `content_direction` (how every page is written) — and the third is
the one the writer reads most directly. Fixing the two obvious ones and building eight
more guides would have produced eight more invitations to browse a catalogue that does
not exist.

**How it was caught: by building one page and READING it.** No spec inspection would have
found it, because each aspect is internally coherent — it is only wrong relative to a
business that no longer exists. This is the case for building ONE and looking, rather than
promoting all nine and discovering it nine pages later. Two, then look, was the right size.

Fixed in `SQL_2026-07-29f_content_direction_editorial.sql`: replaced the four
shop-assuming keys, kept the voice keys untouched (the voice was never the problem), added
`editorial` (D2) and `honesty_rails` (D4). Both guides queued for rebuild.

**`formatted` is the load-bearing field and must be regenerated by hand here.**
`page-content-writer`'s prompt reads exactly one thing:
`{{.site_specs.specs.content_direction.formatted}}`. It is normally produced by the
`write_site_spec` action (`site_spec_actions.go:206-216` → `FormatContentDirection`). A
hand-authored spec that forgets it is **invisible to the writer** — the edit would look
applied and change nothing. The SQL reproduces the formatter exactly (string → `Key: v`;
array → `Key:\n- item`; object → `Key:\n` + per-subkey, joined `\n`; blocks joined
`\n\n`; `HumaniseKey` = underscores→spaces + capitalise first char). Go map order is
random, so block order carries no meaning.

`feed-triage` needed nothing: its prompt iterates every top-level `content_direction`
key, so `editorial` reaches news scoring the moment it exists.

### News feed armed

`classification` now carries `content_features.news_feed` (recommended, separate_page,
source_types rss+news_search+api_news, darts-specific keywords). Gate query verified
passing for this site. Also corrected `reasoning` and `detected_signals`, which were the
research provenance built on the Australian/Portland conflation — not rendered copy, but
what a future planner would read to re-derive the site's purpose, so leaving them would
re-seed the same error. `category`/`site_type`/`recommended_builder` untouched (they
drive builder selection).

Keywords are deliberately darts-specific: if `industry_tags` were ever fed to
`matchVerticalNews`, "sports-equipment" would token-match the generic `sports` entry
whose keywords are "sports news / match results / tournament / league standings" — true
of darts and useless for it.

### growth_config

Inserted (no prior row): `weekly_blog_posts_max` 5 (D3), plus `content_tools_ratio: 6`
(D5) carrying an explicit note that **nothing reads it yet** — the intended reader is a
change to `discovery_checks/check_missing_tools.go`, which today decides tool need on a
7-day/30-day timer with no reference to how much content a site has. Recorded as config
so the policy and its future reader sit together, marked UNREAD so it is not mistaken
for live behaviour.

### Truth reset verified on the artefacts, not the status

`about` and `shipping-returns` rebuilt and checked by counting phrases in the stored
`rendered_html`, not by trusting `status='complete'`:

| page | before | after |
|---|---|---|
| about | "we stock" x2, "we carry" x1, title "Specialist Darts **Retailer**" | all 0; now *"We don't sell darts, hold stock or ship products… an independent guide"* |
| shipping-returns | dispatch / tracking / courier / working days / checkout / cut-off / 30 days | all 0; now *"This site holds no stock and ships no orders… you buy your gear from your preferred retailer"*, and an FAQ that answers "Do you ship darts directly to me?" with "We hold no stock and ship nothing" |

The FAQ is the part I did not expect to come out well: asked "Who do I contact if I have
a problem with an order?" it answers *"the specific store where you completed your
purchase. We can't track shipments or process refunds"*. Setting
`pages.page_spec->>'purpose'` with an explicit FORBIDDEN list was what did it — the same
writer, one hour earlier, wrote the courier promises.

**Follow-up (small, not worth its own build cycle):** shipping-returns contains
"analyzing" — US spelling, against the platform's British-English convention. Sweep it
with the next rebuild of that page rather than churning a build for one word.

### Council + queue notes

- Tool-ratio change submitted to the council gate, correlation
  **`f5fc3014-973c-49a2-8d42-4bf9b401eaeb`**. Commit `f8190a7de` carries NO trailer yet;
  add `Council-Reviewed:` on a later commit only if the verdict is APPROVED.
- **The build dispatcher serves ONE SITE PER TICK, fleet-wide.**
  `build-pipeline-trigger.find_dispatchable_site` is
  `SELECT DISTINCT ON (wi.site_id) … ORDER BY wi.site_id, wi.priority ASC LIMIT 1`, so
  the winner is chosen by **site_id UUID order**, not by priority or age, and a site with
  any `claimed` item is excluded. I briefly diagnosed this as starvation when four
  dartsonline items sat `triaged` while `gaswholesalers` (UUID `5fe15466…`, one sort
  position ahead of `5fe8785b…`) held the slot with an `amend_asset:logo_failtest` row.
  **It was contention, not starvation** — that item completed, dartsonline became the
  selected site on the next tick, and the four items drained in sequence. Worth knowing
  for pacing: queue several items for one site and they run one at a time, a few minutes
  apart, not in parallel. Not filed as a defect; the throughput ceiling is real but no
  site was actually starved.
- Timing trap: I twice concluded "this has been stuck for minutes" from my own sense of
  elapsed time. `SELECT now()` said 56 seconds. Read the DB clock before calling
  something stalled.

### The council gate earned its keep — and I gave it work I should have done myself

Round 1 on corr `f5fc3014-973c-49a2-8d42-4bf9b401eaeb`: **REVISE**, 9 reviewers,
7 abstained (relevance filter), **1 unreadable**.

Two separate things came back, and they need separating because only one is about
the change:

1. **`decided_by: "unreadable reviewer(s): review_prior_art.result"`.** The verdict was
   DECIDED by a seat whose output failed to parse — `bugs_open/138`'s class exactly
   (a degraded review gates the round). Per [[council-revise-may-be-the-harness]], read
   `unreadable`, not `abstained`. With 7 of 9 abstaining, ONE readable seat reviewed
   this change.
2. **A real defect, found by the reviewers' own read-only check.** They asked for the
   fleet's `page_type` vocabulary, and it contains **`guide`** — separate from
   `blog-post` and `content`. My article-count query was
   `page_type IN ('blog-post','content')`, so it counted every `guide` page as nothing.
   Measured after being told:

   | page_type | deployed+active | sites |
   |---|---|---|
   | content | 117 | 13 |
   | blog-post | 52 | 9 |
   | **guide** | **52** | **5** |

   The third-largest article-shaped type, excluded entirely. A site whose written
   content is typed `guide` would count zero articles, never trip the ratio, and never
   be asked for a tool — **silently**, because an omitted page_type does not error, it
   just makes a site look emptier than it is.

**The lesson is about the ORDER, not the error.** My own submission listed this exact
hazard in its `risks` block and invited the reviewer to check it. CLAUDE.md is explicit
that this is not evidence — *"'no collision is possible' is a query, not an argument"* —
and it is right: the query took ten seconds. Writing "a reviewer should check whether
guides under another page_type would undercount" is a confession that I knew where the
hole was and asked someone else to look. Measure the blast-radius claim before you
submit; do not ask the reviewer to.

Fixed in `ced2bca08`: `articlePageTypes` is now a named var carrying the measurement and
a reason for every exclusion, queried with `= ANY($2)`.
`TestArticlePageTypes_CoversTheArticleShapedTypes` pins both the included AND excluded
sets, so the silent failure mode becomes a failing test — and so `tool` can never creep
in, which would let a site satisfy its tool ratio by publishing tools.

The `editquality` objection is separate and also fair: the edit bundles the
Sprintf→json.Marshal safety fix with the ratio change. Keeping it — reverting a genuine
quote-injection fix to tidy a submission boundary is the worse trade — and saying so
plainly rather than quietly, per [[answer-review-objections-with-evidence]].

### News lane fully armed (13:52Z)

The `content-feed-refresh` tick picked the site up on its first pass after the spec
change and seeded **6 sources** — five `news_search`, one per `vertical_keyword` I
authored, plus the `api_news` LLM source, all `error_count` 0. That confirmed the gate
query end to end. Because seeding had now HAPPENED, the all-or-nothing skip could no
longer bite, so the three verified RSS feeds went in immediately after: 9 sources total.
First fetch 19:54Z; items become visible after the triage pass that follows it.

### The page sweep is the reusable instrument

One query classifies every built page clean/SHOP-LANGUAGE over
`page_components.rendered_html`. Eight of nine came back clean, which is exactly what
made the ninth — the homepage, hero saying *"we've got the barrels, shafts, flights and
boards"* over a section headed **"What We Stock"** — worth trusting rather than
dismissing as a false positive. Queued for rebuild; it is the last page carrying the
claim.

### End of session 1 — the sweep is clean on all nine built pages

```
about :: clean          barrel-weight :: clean   beginners :: clean
contact :: clean        guides-index :: clean    index :: clean
new-arrivals :: clean   sale :: clean            shipping-returns :: clean
```

The homepage was the last to fall and it rebuilt well: *"Read the Specs Before You Hit
the Oche… We break down equipment specifications and track PDC tournament gear
releases"*, with card CTAs reading "Read the tungsten guide →" and "See flight
comparisons →". It even reuses `content_direction.example_phrases` verbatim — *"Your
flights aren't an afterthought. They steer your dart."* — which is the evidence that
replacing the four shop-assuming keys did NOT damage the voice keys.

Still emoji on the homepage card grid (⚖️ 📐 🖐️). That is the imagery phase and is
untouched; noted here so nobody reads "clean" as "finished".

**State at close of session 1**
- Honest: all 9 built pages, plus the identity/briefing/classification/content_direction
  specs behind them.
- Content: 9 guides have layouts; 2 built and deployed; 7 remain (the direction behind
  them is now correct, so they can be released in batches).
- News: 9 sources armed (5 news_search, 1 api_news, 3 verified RSS). First fetch 19:54Z,
  visible after the following triage pass.
- Generic: `content_tools_ratio` shipped (default-off, 1 of 14 sites opted in), council
  round 1 REVISE — real defect found and fixed (`guide` page_type), resubmission owed on
  `RESUBMIT_CORR=f5fc3014-973c-49a2-8d42-4bf9b401eaeb`.
- NOT started: imagery (33 assets still barely referenced), tools (setup-builder still
  unbuilt), affiliate resolver/ingester, SEO discovery files, WCAG palette (1.11:1),
  meta descriptions, chrome rebuild for the nav fix to reach the served pages.

**The nav fix is data-only so far.** `pages.in_header` is correct and the orphans are
archived, but the live header still serves the old chrome (bugs_open/117) — the three
404 links will not disappear from the served pages until a chrome rebuild runs. Do not
report the nav as fixed until `curl` says so.

---

## Session 2 — 2026-07-29 afternoon

### CORRECTION to the closing block above: the nav diagnosis was wrong

> **CORRECTED 2026-07-29 (session 2):** "the live header still serves the old chrome
> (bugs_open/117) — the three 404 links will not disappear from the served pages until a
> chrome rebuild runs" is **false**, and the fix it implied would have made things worse.

Measured with `curl`, which is what the same paragraph told the next thread to do and
which I had not done:

| served page | /shop.html /brands.html /guides.html | last built |
|---|---|---|
| `/index.html` | clean | 2026-07-29 |
| `/about.html` | clean | 2026-07-29 |
| `/blog/barrel-weight.html` | clean | 2026-07-29 |
| `/blog/beginners.html` | clean | 2026-07-29 |
| `/shipping-returns.html` | clean | 2026-07-29 |
| `/sale.html` | **3 dead links** | 2026-07-28 |
| `/new-arrivals.html` | **3 dead links** | 2026-07-26 |
| `/guides/index.html` | **3 dead links** | — |
| `/contact.html` | **3 dead links** | — |

The header is regenerated **per page at build time**, and `GetNavItems` already prunes
never-deployed targets. That is why every page rebuilt on 07-29 came out clean with
nobody touching chrome. `pages.rendered_header` is NULL on all nine, so the per-page
header is assembled at render, not stored. `bugs_open/117` (chrome is a stored artefact
no page re-render rebuilds) is real and is a different defect.

**What the wrong diagnosis would have cost.** The queued fix was a chrome rebuild. But
`site_nav_items` was itself stale — it still held the three archived orphan rows and had
never heard of `guides-index` — so rebuilding chrome first would have produced a header
with no dead links *and no Guides*, and every page would have needed building twice.
The right order, and the one used:

1. `nav_drift` → `nav-updater` alone, which re-runs `populate_nav_tables` from the
   corrected `pages` rows. Queued 16:17Z, complete 16:20Z.
2. Verify the table, not the item's status:
   ```
   primary:  Guides /guides/index.html · Start Here /new-arrivals.html · Deals /sale.html
   utility:  Home · About · Contact · Shipping & Returns
   ```
3. Only then the page rebuilds.

### CORRECTION: "all nine built pages clean" was a phrase-list result reported as a judgement

> **CORRECTED 2026-07-29 (session 2):** the closing sweep and the SUMMARY line built on
> it are both wrong. `/sale.html` was serving, and had been all along:
> *"We cut prices across our sale range."* · *"We move high-density tungsten barrels,
> shafts, and flights into clearance regularly."* · *"...when you shop the sale section."*

The sweep tested a fixed list of literals taken from the three fabrication sites I had
just fixed — `stock`, `Add to Bag`, `filter our ranges`, `Portland`, `darts.com`.
`clearance`, `cut prices` and `sale range` were not on it, so the row printed `clean`.
The check was correct; the sentence I wrote about it was not. Full entry, with the
general form, in `WRONG_CALLS.md`.

Confirmed at the source: `page_components.rendered_html ILIKE '%clearance%'` is **true**
for both of `sale`'s components (hero and call-to-action, updated 2026-07-28). The data
was there to catch this the whole time; only the query was narrow.

### THE FOURTH HOME of the false premise — `site_plan_pages`

Found by reading served `<title>` tags rather than page bodies. The current plan
(`0fb05b75`, 2026-07-22) still carried:

- `index` → `"Darts Online | Specialist Darts Equipment & Accessories"`
- `about` → `"About Darts Online | Specialist Darts Retailer"`
- `sale`  → `"Sale | Darts Deals & Clearance | Darts Online"`

`site_plan_pages` is what a reconcile rebuilds `pages` FROM. Session 1 corrected
`identity`, `classification`, `content_direction` and per-page `page_spec.purpose` — four
readers — and left the writer intact. **Fixing every reader of a false premise is not the
same as fixing its source.** Ask which table REGENERATES the ones you fixed.

Also found the same way: 18 of 21 pages had **no `meta_description` at all**, and
`assemblePage` emits `content=""` rather than omitting the tag — a blank authored answer
to "what is this page", which is worse than no tag. Now written for 11 more pages, into
both `pages` and `site_plan_pages`. My own two from session 1 (`index` 189 chars, `about`
180) **failed my own ≤160 assertion** and were trimmed to 145; I had written them without
measuring, in the same session I wrote the rule.

### Retail pages: repurposed, not archived, and why

`sale` and `new-arrivals` are retail landings on a site that sells nothing. Archiving is
the obvious move and is wrong here: `bugs_open/098` establishes **archiving does not
undeploy**, so an archived `/sale.html` goes on serving "we cut prices" to every visitor
and crawler indefinitely. A live page that tells the truth beats an archived one that
lies.

- `sale.html` → *How to Spot a Genuine Darts Deal* (nav: **Deals**) — evergreen
  buying-advice on judging a darts discount from the outside. High commercial intent,
  honest without a feed, and it becomes the natural affiliate page later.
- `new-arrivals.html` → *New to Darts? Start Here* (nav: **Start Here**) — signposting
  hub putting the beginner decisions in the order they arise.

Both keep their URLs and their existing `[hero, call-to-action]` section shape, so no new
section machinery was needed. `shop-index`/`brands-index` set `in_header=false`: retail
hubs with nothing to put in them until a feed exists.

Rebuilt `new-arrivals` (complete 16:24Z) reads: *"Build Your First Setup… We break down
how 22g versus 26g barrels, grip styles, and tungsten percentages actually change your
throw."* Honest and on-voice. `[OBSERVED, not fixed]` it does **not** link out to the
individual guides, because `hero + call-to-action` has nowhere to put a link list — the
signposting job is only half done and needs a listing section to finish.

### Council: APPROVED on round 2

`RESUBMIT_CORR=f5fc3014-973c-49a2-8d42-4bf9b401eaeb`, submitted 16:08Z, approved 16:17Z
— nine minutes, against the ~30 the runbook budgets. Five advisory objections, none
high-severity.

Round 1's three objections were answered with measurements, not argument:

- **prior_art (import cycle asserted, not shown):** package `actions` imports
  `discovery_checks` in **6** files; `discovery_checks` imports `actions` in **0**. And
  `loadGrowthConfig` is unexported (`page_growth_budget.go:209`) with no such field on
  `GrowthConfig`, so it was never callable anyway.
- **debug_historian (status filter used without enumerating it):** right, and my figure
  was wrong. `sites.status` = pool 17 / deployed 14 / system 1 = **32**. Re-measured with
  no filter: 32 sites, **1** carries the key. Also measured the check's real footprint —
  `evaluate_tools` items exist for 9 sites, all `deployed`.
- **editquality (bundled scope):** conceded and declared rather than folded in.

Three seats then independently made the same new point — the field lives only at its
call site, so a future reader of `GrowthConfig` would not know it exists and would add a
third ad-hoc reader. Acted on in `df682c339`: modelled in the canonical struct with the
cycle evidence in the comment. Trailer recorded there, since the approved code
(`f8190a7de`, `ced2bca08`) had already landed and forward-only means it cannot carry one
retrospectively.

### Queued and draining at time of writing

| item | status |
|---|---|
| `nav_drift` → nav-updater | complete 16:20Z, table verified |
| `new-arrivals` rebuild | complete 16:24Z, read and honest |
| `sale` rebuild | claimed 16:24Z |
| `contact`, `tungsten-guide`, `board-setup`, `brand-comparison`, `flight-shapes` | triaged |
| `missing_news_page` → content-gap-planner | triaged 16:24Z |

**Watch on the news page:** `page_type` must be the literal `news-index`. Several gates
key on it and `bugs_open/081` is the case where a mistyped page was DEPLOYED — no repair
path, looping ~3 months. Check the row BEFORE its build item runs, not after.
