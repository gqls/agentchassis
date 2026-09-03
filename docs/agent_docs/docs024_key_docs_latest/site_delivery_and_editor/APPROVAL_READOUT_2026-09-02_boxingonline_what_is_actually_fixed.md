# APPROVAL READ-OUT 2026-09-02 — boxingonline.com: what is actually fixed

**For the owner's approval decision under his own ruling of 2026-09-02:** *"this is the first big
site through the system so let's build and fix everything before approval."* That ruling turns on
knowing which of his fourteen points are genuinely done — so this read-out has **three columns,
not two.** A list of green ticks would hide the third.

**Everything below was MEASURED at the served site or the live DB on 2026-09-02, between 18:50
and 19:0xZ**, not carried forward from earlier reports. Serving host is
`boxingonline.ugg2.com` (⚠ never probe `boxingonline.com` — it is a parked catch-all that
returns 200 with a 114-byte stub for any path). 21 pages deployed.

---

## A. VERIFIED WORKING — measured at the served artefact today

| # | item | evidence |
|---|---|---|
| 0 | **The owner's personal email is off the site** | 0 occurrences across all live pages; 0 in all four DB sources. Two independent sweeps agreed |
| 1 | **Nothing links to the contact page** | 45 inbound links → 0. Only self-links on the orphan itself remain |
| 2 | **Header is logo-only** | no wordmark beside the mark, as ruled |
| 3 | **Guides index lists the guides** | 4 × `/guides/tool-*-guide.html`, **0** `/blog/` links, 4 decks with real copy, 0 title-suffix remnants. Landed 18:52:31Z |
| 11 | **"Free Cost" removed** | 0 occurrences on the calendar page |
| 12 | **One "News" in the nav** | exactly 1 |
| — | **Fight calendar is in the menu** | 1 header entry per page, correct order, all 21 pages |
| — | **Logo carries no invented name** | single composition, zero lettering, verified by eye and pixel-decode |
| — | **Six articles built and linked** | 6 links on both listings, all targets HTTP 200 |
| 1 | **Contact page is GONE — both halves** `[VERIFIED 2026-09-03 08:28Z]` | `/contact.html` → **404**; three kept pages (`/index.html`, `/guides/index.html`, `/tools/fight-calendar/index.html`) → **200**; invented URL → **404** (so the 404 discriminates and is not a blanket); **0 inbound links** across all five pages probed. Verified independently by me and by two other sessions. Mirror sweep fired 22:53:51Z on the site's first post-roll tick, `deleted:1 ["contact.html"] accepted:true`. `bugs_closed/429`, fix `b60d66e3c` in v1.0.1355 |

## B. NOT FIXED — still visibly wrong on the site

| # | item | measured today |
|---|---|---|
| 7 | **The articles contain no news** | `last-nights-result…` and `saturday-fight-card…`: **0 mentions** of Hrgovic, Itauma, Cameron or Mayer — the real 31 Aug results the site holds on its own news page |
| 9 | **Comparator ships no data** | 18 manual inputs, **0** inline data arrays, **0** runtime fetches |
| 10 | **Fight calendar has no calendar** | **0** inputs, **0** data arrays, **0** fetches. The brief's core deliverable |
| 12b | **News page serves raw feed residue** | 5 literal markdown links, 12 truncated fragments, 11 UFC/MMA mentions on a boxing site |
| 3b | **Guides not rewritten** | 3,530 / 3,540 / 4,122 / 4,136 chars, unchanged since 09-01/02. His "more interesting, shorter if there's little to say" is unactioned |
| 8 | **Imagery still thin** | index 7 imgs, articles-index 21 — **everything else is the logo alone**: every article page, the news page, the about page, the calendar. Guides-index cards carry **0** images |
| 14 | **Card decks did not apply to the home page** | see §D — this one *reported success* |

## C. BUILT BUT CANNOT REACH THE SITE — done and inert, pending a release

| item | state |
|---|---|
| ~~**Contact page 404**~~ | **CLOSED 2026-09-03 — moved to §A.** See below |
| **Logo with no baked background** (his ruling 5) | Implemented, then its review caught a real defect — the prompt forbade the exact colour it also demanded. Fixed (`b2322a203`), but the running build predates it. **Do not regenerate the logo until the next roll** |
| **Card headline/deck producer** | Live in the binary since 12:28Z — but see §D |

## D. THE ONE THAT REPORTED SUCCESS AND FAILED — worth the owner's attention

The card fix ran on the home page **twice**, both items carrying the exact instruction our own
guidance says to verify (`reason='template_changed'`), both `complete`. The stored data is
unchanged: title still suffixed, excerpt still absent.

A controlled contrast on one site, one binary, ten minutes apart:

```
guides-index  BUILD path,    17:22:27  title suffix-free · excerpt PRESENT
index         RERENDER path, 17:32:26  title suffixed    · excerpt ABSENT   <- later, and broken
```

So it is a code-path difference, not deployment, data or ordering — and **a correct reason is
necessary but demonstrably not sufficient**, which contradicts the check three sessions spent the
day converging on. Under investigation by the components lane; the contrast above is with them.

## E. STAFFED AND IN PROGRESS ELSEWHERE

`bugs_open/427` — nothing populates dated, correctable facts. **This is the single root of items
7, 9 and 10.** Fleet measurement: 20 of 54 sites have a fact corpus at all; **42 of 54 hold five
facts or fewer**; boxingonline holds **one** — and of the two it was seeded with, one was the
owner's billing email mis-registered as a publishable claim. Three writers exist; none does
proactive research. Now staffed.

Also live: the guide rewrites and the title-promise gate (copy lane), card visual design
(designer + components), article-header imagery (editorial design — the six hero images already
exist and no article component can display one), `bugs_open/433` (71% of assets carry no
recorded MIME type).

---

## The honest summary for the approval decision

**The site is presentable and truthful.** Nothing on it misleads a reader, nothing of the owner's
personal data is exposed, every link resolves, and the pages that exist are well-formed.

**It is not yet the site the brief describes.** The calendar is the customer's named core
deliverable and contains no fights; the comparison tool asks the reader for the data it was meant
to supply; the editorial promises specific news and delivers general essays. All three share one
cause and it is now being worked.

Under his own cut-line, §B and §D are what stand between here and approval. §C will resolve
itself on the next roll without anyone doing anything.

---

# STATE AT END OF DAY 2026-09-03 — supersedes the tables above

**All figures measured at the SERVED site between 13:14 and 13:59Z, after that day's publish tick.**
The tables above are left intact rather than rewritten: the point of the three columns is to show
things MOVING between them, and a row that silently disappears teaches nothing.

## Moved INTO "verified working" today

| item | evidence, at the served artefact |
|---|---|
| **14 — card design** | `/index.html` lm 13:14:20 · **6 deck elements, 0 empty**, **0 suffixed headlines**, **0 suffixed image alts**. The summaries were always in the row under a different key; the rebuild joined them up. §D below is closed by this |
| **5 — logo, no baked background** | SERVED bytes, lm 13:14:09, 35,254 B: **PNG colour type 6, 80.10% fully transparent, border ring 99.84%**. One composition, no lettering. Verified at the artefact, not the row |
| **3 — guides index** | 4 guide links, 0 blog links, 4 decks with real copy. Plus a **"Guides" entry now on the home page**, so the guides are reachable from the front |
| **— analytics** | `GTM-PQ3WCTBD` ×2 and the `cc_v1` consent block on index, news and article pages |

## Item 1 — copy: REPAIRED, with one defect LIVE on the site

The CTA block is **210 visible characters, down from 1,347.** The four-tool walkthrough is gone.
Checks run against the published page: **0** register tells (`we write` / `we'd rather` /
`we cover` / `gets checked` / `the list below`), **0** AI tells (`plainly` / `honest` /
`starting point, not the final word`), **0** placeholders. Both CTA labels are short imperatives.

**The defect:** *"…and the calendar below tells you what's coming up next."* The call-to-action is
the **last section on the page** (`content-listing → info-card-grid → call-to-action`) and there is
no calendar on the home page at all. **A false spatial claim, now serving.**

**Fix filed as `needs_copy_edit a930e70c`** (13:53:07Z) — an INSTRUCTION to the framework about the
subheadline, not a typed value, and it **ends at `checkpoint_for_review`, so the rewrite lands in
the owner's admin queue for approve / request-changes.** It was deliberately held until the
components lane proved a rerender PRESERVES the card fix (batch 691, decks intact) — so correcting
item 1 could not revert item 14. **That ordering was the whole point and it held.**

## Still NOT fixed — re-measured 13:59Z, all unchanged

| item | measured |
|---|---|
| 7 — articles contain no news | `last-nights-result…`: **0** mentions of Hrgovic, Itauma, Cameron or Mayer — the real 30 Aug results the site now holds in `evidence_base` |
| 10 — fight calendar has no calendar | **0** inputs, **0** data arrays, **0** fetches |
| 9 — comparator ships no data | **18** manual inputs, **0** data |
| 12b — news page serves raw feed residue | **5** literal markdown links, **12** truncated fragments, **11** UFC/MMA mentions |
| 3b — guides not rewritten | 3,530 / 3,540 / 4,122 / 4,136 chars |
| 8 — imagery thin | index 7 · articles-index 21 · **guides-index 1 · news 1 · about 1 · every article page 1** (logo alone) |

**7, 9 and 10 remain ONE root cause** — `bugs_open/427`. Note its writer half now WORKS on this
site: `evidence_base` went 1 → 7 real cited facts including a dated forward fixture (Canelo v
Mbilli, 31 Oct). **Nothing consumes them.** That bug's session has ended and needs restarting.

## §D is CLOSED — and the prediction held

The item that "reported success and changed nothing" was the card fix via the rerender path. A
**BUILD** fixed it, exactly as pre-registered at 09:55Z before the run: decks with real copy,
suffix-free headlines **and** suffix-free alts, nothing further done.

⚠ **One claim to narrow, not to inherit:** "the path-split model stands" overstates it. Builds are
now **4 for 4**. But the model's other half — *rerenders never produce the new shape* — was
**refuted** yesterday (two rerenders elsewhere did) and batch 691 has since shown a rerender
PRESERVES the shape. So: **builds reliable; rerenders inconsistent; why, still unknown.** Best lead
remains `garden-tools.uk` — `/index` new, `/care` old, everything else held constant.

## Two false alarms and two traps, for the review

- **`0cdddb6f` stale_attestation** — a three-day-old site's fact "overdue after 180 days", because
  order-intake seeds `attested_by` with no date. **Fleet singleton: 146 attested facts, 1 undated.**
- **`33a900b8` site_unreachable** — probes the parked customer domain, not the serving slug.
- **`pages.rendered_head` reads GTM = 0 on all 21 pages while the served pages carry it** — the head
  is composed from `site_components` at assemble time. Checking that column would report the
  analytics rollout as failed when it succeeded.
- **`/about.html` has GTM = 0** — published ~20 s before its own assemble finished. Queued behind
  the next tick, not a failure.
- **Chrome refresh `ec92320f` is FAILED 3/3** (`bugs_open/457`), so chrome propagation on this site
  is hand-filed until that code fix.

## Logo — the fleet result, for context on his ruling 7

His ruling was general ("the background behind a logo shouldn't be part of the logo"), so the fleet
answer matters more than this site's. **All four sites regenerated through the fixed matte carry
PNG colour type 6 at the SERVED bytes:**

| site | fully transparent |
|---|---|
| seotools | 92.21% |
| designblog | 88.5% (dark-marked, so no light-background legibility risk) |
| websitepromotion | 87.4% |
| **boxingonline** | **80.10%** |

`bugs_closed/424_HANDOFF_2026-09-02_transparency_is_not_a_promptable_property_so_the_model_paints_a_checkerboard.md`
— and its closing statement is the sentence worth pinning: **the matte AND its guard are live on
v1.0.1356, not merely committed.** So the ruling is implemented estate-wide and demonstrated on
four sites, not just satisfied on this one.

⚠ **The guard was the blind part, and that is the transferable lesson:** `border_keyed` scored
**1.000 on a 0.0%-transparent failure** and on an 87.4% success alike, because it counted
border-flood REACHABILITY rather than transparency. **Verify a logo at the PNG bytes — colour type
6 or 4, OR a `tRNS` chunk, testing for both — never at the adapter's own confidence signal.**

---

# VERIFICATION 2026-09-03 18:00Z — the owner's two approved edits, checked at the served bytes

Measured against `https://boxingonline.ugg2.com/index.html`, **Last-Modified 17:32:30Z**,
cache-busted, `<body>` control non-zero, invented paths 404 while the page is 200.

## Both edits he approved at 16:21 are LIVE and correct

| his ruling | on the page now |
|---|---|
| *"The line can just be: News, previews and results from across the sport."* — he cut the model's trailing clause as slop | **verbatim, 1 occurrence**; the old *"odd strong opinion"* phrasing is **0** |
| **CUT IT** — the call-to-action's *"the calendar below"* line | `grep -ci 'calendar below'` → **0**; the subheadline is **emptied**, not reworded, and renders no element |

**The closing block is 96 visible characters**, from 1,347 at the start of the day (and 210 at
midday): a heading and two button labels — *Stay on top of every fight that matters · Catch the
latest boxing news · See the full fight calendar.* Every clause speaks to the reader; none describes
what the section below contains. **Against his item 1 this is a repair, not an abbreviation.**

His item 14 holds alongside it: 6 cards, **6 filled decks, 0 empty**, and **0 template-suffixed**
headlines or alt texts — the two agree character-for-character, which is the property that proves a
suffix has not crept back into either.

Regression guards, **20 of 20 served pages**: email **0** · contact links **0** · `/contact.html`
**404** · one fight-calendar reference in the header · GTM on every page.

## ⚠ How his approval reached the page is NOT how it was supposed to

The job his approval filed — `5edadfbe section_edit` — **failed three times and is still marked
failed**, last attempt 17:56Z:

```
step load_edit_context failed: need either page_component_id or both page_name + slot_name
```

The edits are on the page because **a person applied them by hand at 16:26 / 16:27**, five minutes
after he pressed approve. `page_components` 322ce532 and e5b848fa carry those timestamps.

This is the approve-button defect already diagnosed in this lane and fixed in `33dfeed3a`: the
`page_component_id` **is** in the spec, nested inside `approved_data.edits[]`, while the top-level
`copy_edit` and `page_target` that `load_edit_context` reads are `null` — and two edits in one job
also exceeds the one-target-per-job limit. **Both halves are fixed in code and inert until a roll.**

**Nothing is lost and he need do nothing.** But the record should say plainly: *his approval did not
apply itself.* And `5edadfbe` should be reconciled — it is a failed row standing for work that is
complete, and it will read as outstanding until someone closes it.

## Still wrong on the site, re-measured 17:32Z

- **`/articles/index.html` serves the six articles six times** — **36 cards**, 2 without decks, 14
  without a category label. `bugs_open/457`, **staffed** (7+ commits today, two lanes). Code fix plus
  a rebuild; no re-render clears it. The most visible remaining blemish.
- **Items 7, 9, 10** — articles carry no news, comparator ships no data, calendar has no calendar.
  One root cause, `bugs_open/427`; restate its title before building, because its writer half now
  works (`evidence_base` 1 → 7 cited facts) and nothing consumes them.
- **Item 12b** news feed residue · **3b** guides unrewritten · **8** imagery logo-only.
