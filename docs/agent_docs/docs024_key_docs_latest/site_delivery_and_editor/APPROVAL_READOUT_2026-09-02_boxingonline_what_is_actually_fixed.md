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
| **Contact page 404** | Deletion worked; the publish mirror has no delete capability at all. Fix committed (`b60d66e3c`), roll-bound. Converges automatically on the site's normal hourly slot after the roll — **no forcing** |
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
