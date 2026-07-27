# HANDOFF — webdesign.co.uk, continue here

**Written 2026-07-27 ~18:15 UTC.** Supersedes
`HANDOFF_2026-07-27_phase2_uk_authority.md` as the cold-start document. That file
is still worth reading for the Phase 2 brief and the carried-forward landmines, but
its news sequence is corrected in place and its open questions are all answered —
**start here, then read it for background.**

---

## TL;DR — what changed today, in one screen

| thing | state |
|---|---|
| **Dead home page links** | **FIXED & LIVE.** 10 of 13 hrefs were 404. Verified live, all 99 pages swept. |
| **Why nothing flagged it** | The 3 link checks **have never run, on any site.** `bugs_open/116` (new, unowned). |
| **News feed** | **ARMED but has ingested nothing.** First tick lost to `bugs_open/029`. Re-armed for **19:49 UTC** — check this first. |
| **Cloudflare analytics** | **Still blocked on the owner.** One dashboard step. No beacon on the live page as of 13:05. |
| **Phase 2 direction** | **Redirected by the owner** to an enterprise buyer section, "Buying design". New plan written. |
| **Platform code** | **Nothing changed by me.** Deliberately — see "what I did not do". |

---

## 1. DO THIS FIRST — the news feed

Everything about news is armed and waiting on one tick.

```sql
SELECT count(*) FROM content_feed_items cfi JOIN sites s ON s.id=cfi.site_id
 WHERE s.domain='webdesign.co.uk';
```

- **If > 0:** the feed works. Next: **build the news page** (`/news/index.html`,
  `build_status='planned'`), and **only after it builds**, re-render chrome to
  publish the News nav link. That order is not optional — the nav row already
  exists in the DB, so re-rendering chrome early puts a 404 in the header of all
  98 pages (`bugs_open/049`'s exact shape).
- **If still 0 after ~20:15 UTC:** re-read `bugs_open/029` before anything else and
  check the chassis pod's `startTime` against the tick time. Do **not** conclude
  the arming is broken — it is verified working (see below).

**State as of 18:13 UTC:** 5 of 5 sources armed (`next_fetch_at IS NULL`),
0 fetched, 0 items, 0 recent failures. Task last fired 13:49:16, so **next tick
19:49:16 UTC**.

### What was wrong and what fixed it

The previous handoff said "wait for the feed". **The feed could never have fired.**
`content-feed-trigger`'s `find_news_sites` step enumerates sites on
`site_specs.classification.data->'content_features'->'news_feed'->>'recommended'`,
and this site's classification spec had **no `content_features` key at all** —
`NULL::boolean = true` is NULL, so the row was dropped from every tick.
**Creating `content_sources` rows does not arm a feed.** `SQL_p9` fixed it.

Verified by running `find_news_sites` **unfiltered and unmodified** — the site now
returns in slot 3 of 5. Run it unfiltered; filtering it to your own domain answers
a different question and hides the `LIMIT 5`.

### Then the first tick ingested nothing, and that was not us

All 5 sources dispatched correctly with the editorial queries intact, then died at
`spawn_ingester` — *"timed out after 3 retries"*, ~8 min each. **`bugs_open/029`**
(hung spawns — resolve **by slug**, `bugs_closed/029` is the unrelated
phantom-links case), roll-adjacent: chassis `startTime` 13:45:31Z, tick 13:49:09 =
**218 s**, inside the ~300 s window.

**The check that made attribution cheap, and the transferable lesson:** my feed had
been armed that same hour, so the obvious reading was "my change is wrong". One
query on **vetcomparison.uk — same tick, a site I never touched — showed the
identical failure.** Not ours. 029 is heavily owned (6 council rounds); occurrence
contributed to the bug file, no competing fix.

### Two measured facts about the feed worth keeping

- **The 6-hourly task effectively runs 12-hourly.** Across 3 days, items land only
  in the **07:xx and 19:xx–20:xx** hours, never 01:xx or 13:xx. Cause:
  `UpdateSourceTimestamps` sets `next_fetch_at = NOW() + interval` at *ingestion*,
  minutes after the trigger, while the next tick is `last_triggered_at + interval`
  — so a fetched source comes due *just after* the following tick and misses it (by
  **37 seconds** on ai-agent-orchestration.com). **So 13:49 is structurally the
  quiet tick**, which is why one roll cost 100% of it, and why 19:49 is a real
  chance.
- **`find_news_sites` ends `ORDER BY s.domain LIMIT 5` with no rotation**, and this
  site is now the **6th and alphabetically last** recommended site. If 5+ are ever
  due together, ours is dropped, silently, every time. Not biting yet only because
  of the staggering above. Recorded, not filed.

### The landmine in `SQL_p9` — do not edit the keywords casually

`content-feed-orchestrator` runs `seed_sources` **before** `dispatch_sources`, and
`seed_content_sources` creates one source **per `vertical_keyword`**, named
`"News Search: <keyword>"`, `ON CONFLICT (site_id, name) DO NOTHING`. The keywords
in `SQL_p9` are set to **exactly the five source-name suffixes `SQL_p8` used**, so
the auto-seeder collides and no-ops. **Change one character and a sixth source
appears with the bare keyword as its query**, quietly diluting the editorial ones.
`source_types` is `news_search` only — `api_news` would add an unchosen xAI source.

---

## 2. The dead home page links — fixed, and what it exposed

**Fixed and live.** Every link on the home page resolves. Verified against the live
URL (not the DB), then all **99** deployed HTML files swept against the artefacts
on disk.

**What was wrong:** 10 of 13 hrefs 404'd; only the three nav links worked, which is
why it survived inspection — the menu works, so the site feels fine until you click
a card. All 12 cards across two `info-card-grid` components were dead. Two
independent faults:

1. **Invented targets** — `colour-contrast-checker` and `css-layout-generator` name
   pages that exist under other names (`smart-contrast`, `layout-generator`);
   `spacing-scale-calculator` and `typography-scale` name tools that exist in **no**
   form among the 63 built; four category links point at category pages never built.
   Slugs absent from `cmd/webdesignport`, so this is generation — `bugs_open/092`.
2. **Wrong path shape** — `/tools`, `/guides` and the category links have no
   `/index.html`.

**Fix:** `SQL_p10` corrects `content_data` **and** `rendered_html` with one shared
replacement list — both, because assemble republishes *stored* HTML, so fixing the
data alone changes nothing served. Card copy was corrected where it described a
tool we do not have. Also fixed in `gqls/sites` (commit `b0dbe8358`).

### THE DEPLOY FACT I DID NOT KNOW — read this before touching any page

**The site is NOT served from the database.** Pages are published as files into
**`~/projects/sites`, branch `master` (not `main`)**, and a GitHub Action
(`deploy-to-b2.yml`) ships changed `<domain>/` dirs to B2 behind Cloudflare.

My local clone was **394 commits behind with no `index.html` at all** — which looks
exactly like "the site is not in the repo" and is not. **`git fetch` first.**
`gh run list` / `gh run watch` to follow the deploy (~25 s).

### Object stores cannot resolve directory indexes — fleet-wide

```
webdesign.co.uk  /tools/ 404  /tools 404  /tools/smart-contrast/index.html 200
relojistas.com   /about/ 404  /about 404
robot-hands.com  /about/ 404  /about 404
gaswholesalers   /about/ 404  /about 404
```

`x-amz-*` headers on every response: S3-compatible bucket behind Cloudflare. **This
is inherent, not a misconfiguration, and it will never start working.** Roots serve
because the bucket has a default root object; subdirectories never will.

### Why nothing flagged it — `bugs_open/116` (NEW, unowned)

`phantom_internal_links`, `dead_controls` and `misdirected_cta` are enabled on
exactly one agent, `completeness-discovery-agent`, and **that agent has zero runs
across the full 13-day `orchestration_states` retention, on every site.** The only
discovery agent ever to run is `design-discovery-agent` (8 runs), which carries none
of them — and it *did* run here, at 2026-07-26 21:31, producing no link finding
because it never looked. Its only recurring caller is `improvement-loop` (**off**);
the only scheduled task pointing at it is disabled, a one-shot, for another site.

**Owner steer, 2026-07-27:** *"whilst the improvement loop will return the checkers
should run after every build or change I think."* That is candidate 1 in 116 and
the right one — a sweep only reports after damage ships.

> **"All 98 pages return 200" and "the site works" are different claims.** The first
> was true and I had only ever checked it. Any future "the site is live and clean"
> must name which of the two it means.

---

## 3. What I deliberately did NOT do — read before you pick this up

**I changed no platform Go code**, and the reason should survive into the next
thread.

My approved plan said I would change `NormalizePagePath`
(`platform/orchestration/datahelpers/links.go:169`) so the platform stops treating
`/tools` as equivalent to `/tools/index.html`. Before editing I checked ownership:

- **`bugs_open/071` already contains a section reasoning about that exact
  function**, concluding the repair belongs at the writer, not the normaliser.
  Owned by `brochure_component_library` (68 commits/14d) and `gauntlet_dead_cta`
  (60 commits/14d).
- **`bugs_open/092`** (the generation half) is owned by
  `bugfix_079_phantom_link_gate`, which committed to it the same day.

I would have shipped a competing fix into three active workstreams' code.

**My finding is still new and still real**, and is contributed *into* 071: 071's
analysis covers the flat-file shape (`pages.url = /about.html`), where the mismatch
is flagged correctly. **The `dir/index.html` shape inverts it into a FALSE MATCH** —
`/tools/index.html` normalises to `/tools`, the href `/tools` normalises to
`/tools`, they match, and it 404s live. That is the one link of the ten the audit
would have passed even had it run. Also note
`rerender_page_sections_action.go:429` compares normalised paths, so the
`index.html` strip may be load-bearing there — that interaction wants the owner's
judgement.

**Logged in `WRONG_CALLS.md`:** I ran `scripts/who-owns.py` *after* writing the plan
rather than before. **Writing a plan that touches a bug's code IS routing work at
it**; the check belongs before the plan, and costs 0.3 s.

---

## 4. Phase 2 was redirected — "Buying design", not a generic buyer track

**Live plan: `PLAN_2026-07-27b_buying_design.md`.**
`PLAN_2026-07-27_phase2_buyer_track.md` is **superseded**, kept as evidence of how
the audience was first mis-sized. Owner rulings D10–D12 are in
`PLAN_2026-07-25_webdesign_couk.md`.

**Owner rejected:** the nav label `Hire` (*"a bit cheap-sounding… sounds like we're
hiring an Upwork or Fiverr worker, completely the wrong end of the quality and
price scale"*) and the page "how to judge a quote" (*"focuses the content on money
and not design"* — a better objection than my no-figures one).

**Audience:** the **£100k–£multi-million commissioner**. Scalability and reliability
as much as brand proposition; expressing an offline brand online; extending an
online brand without diluting it. **Not** the small-business buyer.

**Measured:** the buyer track is **100% new writing** — all 63 tools and all 31
guides are practitioner-facing; there is no buyer content to improve.

### Positioning (mine, owner endorsed, one decision outstanding)

We run an AI build system across ~1,000 sites and are **not bidding for the buyer's
project** — the only party in that market with nothing riding on the answer about
AI. **The differentiator is publishing where it fails, not where it works.**

Owner: *"I am happy to admit our own failures completely, but the overall message
must be one of success — our successes must triumph over the failures."* He also
wants the positioning extended to **human design helped by AI, and AI automation
nurtured and guided by humans — used fully where it works, not used where it
doesn't**, and to explore competing against AI copy that is *"looking less and less
like slop"*, being a destination in the noise, and **owning the customer** because
*"relationships are and will always remain human to human"*.

> ⚠️ **DECISION STILL OPEN — do not write at this level until it is made.**
> How exposing: (a) generic failure classes, (b) anonymised cases, (c) **our own
> named failures with evidence**. (c) is most credible and most exposing. Nothing
> has been written at (c).
>
> **My unraised caution, worth putting to him:** if every published failure is
> followed by a redemption, readers detect the pattern within three examples and it
> reads as humblebrag. A workable rule: **publish failure modes inherent to the
> medium** (every AI system invents figures — a class statement) and **be discreet
> about failures merely idiosyncratic to us**. That keeps the honesty real and the
> commercial position intact.
>
> **Also unraised:** "human + AI, best of both" is the single most common
> positioning line in this market. What makes it credible is not claiming the blend
> but **publishing exactly where the line is drawn and why** — which nobody else
> does, because most don't know their own line.

### Tools for this market

Test that survived: not "is it useful" but **"does the same tool get used at three
stages of a supplier selection?"** (build the internal case → brief → shortlist →
pitches → contract → oversee delivery).

- **Anchor: side-by-side site benchmark.** Your site vs five others. Used to build
  the board case, to check whether an agency's *delivered* work stands up rather
  than the case study about it, and to verify what you paid for. Three returns.
- **BUILD FIRST: accessibility duty check.** The contrast, 44px and focus-visibility
  tools **already exist** — the buyer version is a reframe, not a build. One
  engine, two audiences, and it carries boardroom weight.
- Then brand-consistency, ownership/handover checklist, brief scaffold.

**Grounded, not assumed:** `internal/adapters/browserrunner/` really does run
headless Chromium with desktop+mobile profiles and `no_horizontal_overflow`,
`no_console_errors`, `page_status_ok` (`run_checks_action.go`). **It is NOT a
performance or accessibility suite** — do not imply it is.

**RAIL:** tools run on URLs the **buyer** supplies and results go to the buyer.
**Never publish comparative rankings of named agencies** — different risk class, and
it destroys the neutrality the section depends on. Therefore **D11's directory means
tools, not suppliers.**

### Figures

**No figures, no page** — strengthened at this tier, where an unsourced number does
more damage than a missing one. Owner wants **verified figures from top agencies**.
Best honest source: **Companies House** filed accounts (and note
`platform/orchestration/actions/companies_house_fetch_accounts_action.go` **exists**)
— but the related matcher cannot record provenance (`bugs_open/100`/`101`), so treat
it as a lead, not a dependency. Second: HTTP Archive / CrUX. Legal claims: primary
source only.

### Design/experience — owner's latest steer

He wants a *"churning exciting builder feel… a real design agency and not just suits
pitching for money"*, showing what works now, what others are doing, what we're
doing including mistakes, as *"active participants at the head of the game"* — while
**not scaring them away**, and being **excited about other people's solutions too**
so buyers see the whole marketplace.

**He explicitly said this is not direction and wants views either way.** My view,
for the next thread to put to him: **the excitement should come from visible
liveness, not visual pyrotechnics.** A page showing what shipped this week, what
broke, what was measured is more impressive to an enterprise buyer than any hero
animation, and far harder to fake. Counter-argument to hold honestly: a design buyer
judges a design site on its design, immediately — so the visual bar is genuinely
high; it just shouldn't be met with generic agency motion.

**IMAGERY — owner corrected me and this is settled:** *"we should keep to the design
feel as it is a good one."* **Do not force imagery onto every page.** Current
baseline measured: 19 of 63 tool pages and 7 of 31 guides carry any visual; only 13
carry a real `<img>`. Note also that a site positioned on honesty about AI's
weaknesses **cannot** use obviously-AI stock imagery without self-refuting — real
screenshots, measurements and diagrams are both safer and better for this audience.
And `prefers-reduced-motion` appears **0 times** in the live `styles.css`, which
matters for a site about to publish on accessibility duty.

### Future, recorded not started

Owner's next build after this section: **a website creation form using our system** —
migrate the site onto a VM, use **`tools-api`** (`cmd/tools-api` and
`internal/tools-api` exist; one live endpoint, `/api/v1/tools/gauntlet`), with a
**separate cluster and separate database**. That isolation constraint is the bit
that gets quietly eroded later — hold it. Precedent: `idea_uk_vm_site` has run the
full chain to a completed paid transaction. The brief-scaffold tool is the same
interaction shape and may become its first step.

### The tension, named not averaged

Owner wants *"a whole load of webdesign related traffic any way we can"* **and** to
pitch at the multi-million buyer. Different games. My reading: the 94 practitioner
pages are the **traffic engine** (already built, they rank, keep them); buying
design is the **positioning layer** and needs no volume. **Real risk is register
collision** — a brand director landing on *Fractional Layouts: The Math of CSS Grid*
concludes this is a developer site and leaves.

**Reframe that raises the rewrite's priority:** because the two source domains stay
live, the 94 imported pages **duplicate** them — so the copy rewrite is doing double
duty, quality **and** de-duplication. Buying design is the only part of the site with
zero duplication risk.

> **STILL OPEN AND IT DECIDES REAL WORK:** the owner said *"ultimately we will not
> focus on the designers"*. I argued they stay as the traffic engine with no
> investment beyond the rewrite. **Confirm or correct** — it is the difference
> between W2's practitioner half being a real rewrite or a holding action.

---

## 5. Still blocked on the owner

1. **Cloudflare Web Analytics — one dashboard step.** Re-checked 13:05 UTC: still no
   beacon on the live home page. Dashboard → Web Analytics → add `webdesign.co.uk`
   → **Automatic Setup** (the zone is already proxied; nothing in this repo
   changes). Until then **nothing is being collected**, and every "order by
   popularity" decision waits on it.
2. **The (a)/(b)/(c) exposure decision** above.
3. **`Buying design` confirmed** as the nav label, and the eight-page inventory.
4. **Do the designers stay?**
5. **Favicon** — `/favicon.ico` 404s on all 98 pages. No fleet site has one; picking
   a brand mark is his call, not something to invent.

---

## 6. Smaller leftovers

- **Two near-identical "What's here" sections** on the home page, with overlapping
  cards ("Front-end guides"/"Practical front-end guides", "Full tool
  library"/"Find the tool you need"). A design question, not a link bug — but now
  visible.
- **No category pages exist.** After `SQL_p10`, five of six cards in the first
  component point at `/tools/index.html` because there is nothing more specific.
  Building real category pages (colour / CSS / typography / accessibility) is the
  better answer.
- **W2's measured starting item:** American spellings in body copy on **23 of 98**
  pages. The pattern must exclude `color`/`center`/`gray`/`behavior` (CSS tokens;
  `behavior` contributes zero, all `scroll-behavior`). **Two of the three affected
  titles are also live slugs — Britishise the title, never the URL.**
- **Ordering by popularity stays last**, and only after stats accumulate against
  rewritten content.

---

## 7. Commits from this session

Chassis repo (branch `086_experience_loop`):

```
7e902ccea docs(README): dead home page links in plain prose — and that the checks had never run
b274c4ed5 fix(webdesign.co.uk): every content link on the home page was a 404; file what that exposed
b99eec2a1 plan: "Buying design" — the enterprise buyer section replaces the buyer track
cb63098a6 docs(HANDOFF): correct "wait for the feed" in place, close all four owner questions
1cc3272ce docs: first tick dispatched correctly and ingested nothing — 029, not us
e4aa13c03 docs(029-hung-spawns): fresh roll-adjacent instance, in feed ingestion not builds
b2daf54b0 docs: plan the buyer track D10 created, size W2 with measurements not estimates
de0aca43e docs(PLAN): D10-D12 — the owner's three Phase 2 rulings
169639800 fix(news): the feed was unarmed — sources are not the switch, the classification flag is
```

`gqls/sites` (branch `master`): `b0dbe8358` — the home page link fix, deployed.

**New/changed files:** `SQL_p9`, `SQL_p10`, `PLAN_2026-07-27b_buying_design.md`,
`bugs_open/116`, plus additions to `bugs_open/029`, `bugs_open/071`,
`WRONG_CALLS.md`, `NOTES_webdesign_couk.md`, `README_where_we_are.md`.

**No summary written this session, deliberately** — per the rarity rule, the five
headings would largely repeat `SUMMARY_2026-07-27_corrections_and_handover.md`.
Write one when the news page is live and the buying-design shape is agreed; that
will be a genuine inflection.
