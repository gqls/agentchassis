# PLAN — getting dartsonline.com enough traffic for the affiliate route

Written 2026-08-18, after **Webgains rejected the application for insufficient traffic**.
Owning lane: `dartsonline_traffic`. Companion to `HANDOFF_2026-08-16_continue_here.md`,
which stays valid — this plan supersedes only its §6 (the affiliate recommendation).

Everything below marked `[MEASURED 2026-08-18]` was checked today against the live site,
the live database or the source. Everything else is marked `[UNMEASURED]` or `[INFERRED]`
**where the claim is made**, not in a footnote.

---

## 0. The one thing we do not have: a number

**We cannot currently state this site's traffic, and neither could I when writing this
plan.** `[MEASURED]` — there is no traffic figure anywhere in the platform's database, the
site is served from Backblaze B2 behind Cloudflare (`x-amz-*` + `server: cloudflare` on
every response), so **there are no origin nginx logs** of the kind that produced the
relojistas crawl-budget evidence on 2026-07-28. That analysis cannot be repeated here.

Three sources exist and **all three need the owner's login**:

| source | what it gives | cost to switch on |
|---|---|---|
| **Cloudflare zone analytics** | requests, unique visitors, country split, bot vs human — for the whole history already recorded | none. Log in, read it. **Fastest number available.** |
| **Google Search Console** | impressions, clicks, average position, **and the actual queries we surface for** — the only source of our own demand data | verify the domain (DNS TXT or a file we can deploy). No verification meta tag is served today `[MEASURED]` |
| **GA4 via GTM** | on-site behaviour | `GTM-PQ3WCTBD` is on every page `[MEASURED]`, but whether a GA4 tag actually fires inside that container is not checkable from here `[UNMEASURED]` — and the consent gap (handoff §3.3) sits on top of it |

**This is step 1 of everything below.** A traffic plan whose progress cannot be observed
is a wish. Webgains asked for a number; we should be able to say what it is before we
re-apply, and we should watch it move.

> I could not confirm from here whether Google has indexed the site at all. The search
> tool available to me does not honour `site:` or exact-phrase operators — a search for a
> verbatim sentence from the tungsten guide returned topically similar pages from other
> darts sites and nothing of ours, which is **suggestive and not proof** `[UNMEASURED]`.
> Search Console settles it in one screen and nothing else does.

---

## 1. Why the traffic is not there — three measured causes, in order of cheapness

### 1.1 Google is not being told what pages exist

`https://dartsonline.com/sitemap.xml` → **404** `[MEASURED]`. So do `/sitemap_index.xml`
and `/sitemap.txt`. `robots.txt` is Cloudflare's managed file and carries **no `Sitemap:`
line** `[MEASURED]`.

This is not normal for the estate. Sampled today `[MEASURED]`:

| site | /sitemap.xml | `Sitemap:` in robots.txt |
|---|---|---|
| relojistas.com | 200 | yes |
| webdesign.co.uk | 200 | yes |
| noted.co.uk | 200 | no |
| loancalculator.co.uk | 200 | no |
| **dartsonline.com** | **404** | **no** |
| robot-hands.com | 404 | no |

**The tool already exists**: `scripts/site-discovery-files.py <domain> --write` generates
robots.txt, sitemap.xml and llms.txt for any site, probes every URL first so it can never
advertise a 404, and builds llms.txt from the pages' own `<h1>` and first sentence. It was
written on 2026-07-28 for exactly this problem on relojistas. It does not deploy — that is
deliberate — so publishing is a separate, explicit step.

### 1.2 The site's freshness engine produces nothing a search engine can index

- **480 feed items ingested for this site, 0 ever published as a page** `[MEASURED]`
  (155/week, still running today).
- Fleet-wide the same: **10,694 feed items across 9 sites, `published_page_id` set on
  ZERO of them** `[MEASURED]`.
- And it cannot be otherwise: **`published_page_id` has no Go writer anywhere in the
  codebase** `[MEASURED]` — `grep -rn "published_page_id" --include=*.go platform/ internal/ pkg/ cmd/` returns nothing.
- What `/news/index.html` actually is: a listing whose **20 news links all point at other
  publishers** `[MEASURED]`, with 13 internal links, all of them navigation.

So the one part of the site that updates daily adds **no new indexable URL, ever**, and
passes authority outward to the darts publications we are competing with. The news page is
re-rendered several times a day (three `page_rerender` items today alone `[MEASURED]`) and
the site is no larger afterwards.

This is a **fleet-wide gap, not a dartsonline misconfiguration** — worth saying because it
means the fix is worth building once and is not this site's to fund alone.

### 1.3 Nothing is writing new content, and it is not the budget stopping it

New pages by week `[MEASURED]`:

```
w/c 07-06  14 pages   (initial build)
w/c 07-20   9
w/c 07-27   3
w/c 08-03   2         <- last new guide
w/c 08-10   0
w/c 08-17   1         (privacy, and it did not build — see §2.1)
```

The site's own `growth_config` permits **5 blog posts + 3 content pages + 3 structural
pages per rolling 7 days**, absolute max 60 pages `[MEASURED]`. It has 25 active pages and
is using **none** of that weekly allowance.

The cause is structural, not editorial `[MEASURED from source]`:

- `needs_blog_posts` — the item that drives `blog-content-planner` — is produced by
  **exactly one thing**: `discovery_checks/check_empty_blog.go`, which fires only when a
  blog index exists **with zero posts**. We have 11. It will never fire here again.
- `needs_content_planning` — the item that drives `content-gap-planner` — is produced only
  by `write_audit_findings_action.go`, i.e. when an audit finds a gap. It is a repair
  path, not a cadence.

**There is no recurring "publish more" driver in the platform.** After the initial fill, a
site's content stops growing unless a person files a work item. That is the whole
explanation for the flat weeks above, and it is why traffic has not compounded.

### 1.4 Two smaller things an affiliate reviewer will actually see

- **`/privacy.html` still 404s** `[MEASURED]` — see §2.1. Webgains, Awin and Amazon all
  look for it, and our own gap-planner filed it as *"required as a precondition for
  affiliate network applications"*.
- **`/shipping-returns.html` is live at 200** `[MEASURED]`, describing shipping and returns
  on a site whose own identity spec says it holds no stock, sells nothing and ships
  nothing. A network reviewer reads exactly these pages. The owner decision to retire it is
  still open (handoff §4.3) and this rejection is the argument for closing it.
- **Brand collision** `[MEASURED]`: `dartsonline.com.au` is an established Australian darts
  retailer that ranks for "Darts Online". Brand-name search will not find us for a long
  time, so every visit has to come from a specific question we answer better than anyone —
  which is what the guides are for, and why more of them is the plan.

---

## 2. Phase 0 — this week, before any content work

### 2.1 Unblock the privacy page. It is one sentence away from building.

The page row exists, the copy is registered in `evidence_base`, the planner did its job —
and the build **failed at content validation with 1 blocker** `[MEASURED]`, leaving the
work item parked at `needs_human_review` since 2026-08-17 12:24. The blocker, read from
`agent_error_log` (`CONTENT_VALIDATION_BLOCKER_DETAIL`):

```
type:        banned_claim
value:       "does not appear here"
category:    claims   severity: blocker
description: Banned claim "does not appear here" (completeness-of-exclusion, short form.)
location:    "...and a product does not appear here because it pays better. Your rights..."
```

The offending sentence is in the **owner-approved copy** (draft file, line 101):

> our guides are written on specifications and how equipment behaves, and **a product does
> not appear here because it pays better**.

The guard is fleet-wide (`platform/orchestration/datahelpers/claims_global.go:130`) and it
is a good guard: it exists to stop unverifiable completeness claims like *"a claim without
a source does not appear here"*. Our sentence means something different — it is an
affiliate-integrity statement — but it matches the pattern.

**Two remedies, and I recommend the first:**

- **(A) Reword, keep the meaning.** e.g. *"…and commission never decides what we
  recommend."* or *"…and we do not rank a product higher because it pays more."* Both say
  the same thing in a form the guard does not match. **This is a change to copy the owner
  approved, so it needs one word from him** — it is editorial, not legal.
- (B) Register a per-site exception to the fleet-wide pattern. Cheaper today, worse
  forever: it weakens a guard for a phrase shape that will recur on every affiliate site we
  build.

Then re-file and verify the served page carries the copy verbatim (handoff §3.1a — the
"verbatim" guarantee here is prompt-level, so it must be checked by eye, not assumed).

### 2.2 Publish a sitemap and point robots.txt at it

**Done as far as it can be without a deploy decision: the three files are generated and
waiting** `[MEASURED 2026-08-18]`. Dry run then `--write` into a scratch directory:

```
site dartsonline.com   live pages in DB: 23
probed: 23 fetchable, 0 dropped
wrote robots.txt (705 b), sitemap.xml (2407 b), llms.txt (4170 b)
NOT committed and NOT deployed -- deliberately.
```

All 23 live URLs probe 200, so the sitemap advertises nothing dead.

**The deploy route is concrete, and it is why the file is missing rather than broken**
`[MEASURED]`: static sites are served from B2 out of a per-domain folder in the `sites`
repo (local checkout `~/projects/sites/<domain>/`). `webdesign.co.uk/` in that repo
contains `robots.txt` and `sitemap.xml`; **`dartsonline.com/` contains neither** — which is
exactly the 404 we measure on the live site. So publishing is: pull, copy the three files
into `~/projects/sites/dartsonline.com/`, commit **by pathspec** and push.

**I have not pushed.** It deploys to a live site out of a repo other sessions also work in,
and one file in it is a live decision the owner has open: the sitemap would list
`/shipping-returns.html` (§1.4). Decide that page first, then ship the sitemap once.

⚠ Cloudflare's managed robots.txt is **prepended** to ours, not replaced — the tool detects
this and says so. Shipping our file will therefore **not** unblock any crawler: that is two
dashboard settings, per zone, and both are needed (*Block AI training bots* → allow, and
*Set your preference to block training in robots.txt* → no preference). Full recipe,
including the per-agent verification loop — a single `curl` cannot tell you, because the
file is served conditionally — is in `docs024_key_docs_latest/FLEET_GUIDANCE_discoverability.md` §2.

Currently disallowed on this zone `[MEASURED]`: **ClaudeBot, GPTBot, CCBot,
Google-Extended, Applebot-Extended, Amazonbot, Bytespider, meta-externalagent**. The owner
ruled *allow everything, including training* for the relojistas zone on 2026-07-28; **every
other zone, including this one, still carries the block.** The argument that decided it is
worth re-reading before answering for darts (fleet guidance §4): a robots block is
voluntary compliance, so it stops the crawlers that would cite and link us and does nothing
to the ones that scrape without credit.

> Note for anyone repeating the relojistas reasoning: **the "dead URLs are eating the crawl
> budget" causation was WITHDRAWN** in that guidance — crawl budget is a large-site concept
> and does not bind below a few thousand URLs. The sitemap is worth shipping because
> **discovery** is the constraint on a 23-page site nobody links to, not because of crawl
> budget. This plan does not rest on the withdrawn claim.

### 2.3 Get the number (owner)

Cloudflare analytics today; Search Console verified this week. Without these, §4's
"re-apply when X" has no X.

---

## 3. Phase 1 — content cadence, starting immediately (weeks 1–6)

**The route is proven on this very site and needs no code change.** File
`needs_content_page` work items directly, one per article, each carrying its own title,
purpose and sections, at `status='triaged'` (`detected` does not drain here). That is
exactly how the privacy page was requested on 2026-08-17 and how `/news/index.html` was
built on 2026-07-29 (`SQL_2026-07-29n_news_page.sql`). The framework writes the copy,
validates it, builds and deploys it. **We choose the topic; we do not write the words** —
which is also the standing owner ruling (CLAUDE.md: the framework writes the content).

The alternative — filing one `needs_blog_posts` item and letting `blog-content-planner`
choose — works too, but **it plans only 3–4 posts and its prompt has no topic input at
all** `[MEASURED from the live agent row]`: it reads `site_record`, `site_specs` and
`existing_posts`, and nothing else. So it cannot be aimed at a search query. Use it if we
want volume without direction; file items directly when we want to answer a specific
question people actually type.

**Cadence: 5 blog posts per rolling 7 days** (the budget's own limit), plus content pages
as needed. Six weeks of that takes the site from 11 guides to ~40 — and, more to the point,
gives Google a reason to crawl it weekly instead of never.

**Topic selection is the one thing I cannot ground yet** `[UNMEASURED]`. Nobody here has
keyword volume data, and inventing it would be the exact failure this lane keeps logging.
Two honest sources, in order:

1. **Search Console query data**, once verified (§2.3) — our own impressions tell us what
   we nearly rank for. That is the highest-value input and it arrives free.
2. Until then, the site's own editorial spec already sanctions the shape that is missing:
   `content_direction.editorial.analysis_scope` — *"Take something that happened … and
   answer the question a player actually has: what would I change about what I throw, and
   why?"* `[MEASURED — it is in the live spec]`. **We have 480 ingested news items and an
   editorial policy for turning them into original analysis, and have never once done it.**
   That is the cheapest topic pipeline we have, it is on-brand by construction, and it
   converts a link-out listing into indexable pages.

Candidate directions beyond that, all `[UNMEASURED]` on volume and offered as a starting
list for the owner to cut: checkout and scoring reference content (checkout charts, what to
leave, common finishes); board and oche setup dimensions; buying comparisons at price
points; brand-by-brand barrel profiles; and pages that give the two existing tools
(`dart-weight-comparator`, `setup-builder`) something to be linked from. Tools are the
site's most distinctive assets and currently nothing points at them but the nav.

---

## 4. Phase 2 — the affiliate route, which should not wait for Webgains

**Webgains is not the only door, and it is the one with the traffic gate on it.**

| network | darts merchant | gate | note |
|---|---|---|---|
| **Awin** | **Red Dragon** (10%, 60-day cookie) | **£5 refundable deposit**, review in ~2 working days | the deposit *is* the filter; credited to the account on approval, refundable if rejected |
| **Paid On Results** | Red Dragon | UK network, application-based | second route to the same merchant |
| **Adtraction** | **Darts Corner** (13,000+ products) | free account, apply to programme | best fit for comparison guides |
| **Webgains** | **Target Darts** (8%, 30-day) | **traffic — this is the one that rejected us** | reviewers check *Monthly Visits* and niche fit. Direct contact for the programme: `target-darts@webgains.com` |
| **Amazon Associates UK** | everything | no traffic minimum, but **3 qualifying sales within 180 days** or the account closes | a clock, not a gate. Apply when there is traffic to convert, not before |
| eBay Partner Network | darts inventory | generally lenient | worth a look, unglamorous but real |

**Question for the owner, because it changes the next move:** did Webgains reject the
**network account** or the **Target Darts programme**? A network-level rejection means
re-applying later with a traffic number; a programme-level one may be answerable by email
to the account manager above, with the guides as the argument.

**Recommended sequence:** apply to Awin and Adtraction now — their gate is not traffic —
so the site can carry real affiliate links while it grows. That also converts the privacy
page from a formality into an accurate document, and gives Webgains something to see on
re-application: a working publisher, not an application.

---

## 5. Phase 3 — the durable fix, and it is bigger than this site

The feed→page gap (§1.2) is worth building properly: **10,694 ingested items across 9 sites
and no mechanism to publish any of them.** A mechanism that turns a notable feed item into
an on-site analysis page — sourced, attributed, written to the site's own editorial rails —
would give every news-carrying site in the estate a daily indexable page instead of a daily
link-out.

That is a **shared mechanism on a shared seam**, so it is architecture-scope by CLAUDE.md's
own test and belongs in `architecture_review/` with a council round, not in this lane's
SQL. Filed here as the recommendation, not started.

---

## 6. What I need from the owner

1. **The reworded sentence for the privacy page** (§2.1) — pick one, or write your own, and
   the page builds.
2. **Which Webgains rejection was it** — network or Target Darts programme (§4)?
3. **Cloudflare + Search Console access or a screenshot** (§0/§2.3) so we can measure.
4. **`/shipping-returns.html`: retire it?** (§1.4) Still open from 08-16; the rejection
   makes it urgent rather than tidy.
5. **Apply to Awin/Adtraction now?** (§4) £5 refundable on Awin; nothing on Adtraction.
6. **AI crawlers**: Cloudflare currently blocks ClaudeBot, GPTBot and the rest (§2.2).
   Leave, or open?

---

## 7. What would tell us this is working

In order of how soon they move, and each one falsifiable:

- **week 1**: sitemap live and listed in robots.txt; privacy page serving 200 with the copy
  verbatim; Search Console verified and reporting *any* impressions.
- **weeks 2–6**: pages indexed climbing with pages published (if we publish 30 and Search
  Console indexes 6, the problem is quality or discovery, not volume — and we will know
  which because Search Console says why).
- **week 4 onward**: first non-zero clicks, and the query list — the first evidence of what
  this site is actually good at, which then picks the next 20 topics.
- **re-apply to Webgains** when there is a monthly-visits number worth stating. **I am not
  going to invent a threshold**: no published Webgains minimum exists `[UNMEASURED]`, and
  the honest version is "when the trend is up and to the right and we can show it".
