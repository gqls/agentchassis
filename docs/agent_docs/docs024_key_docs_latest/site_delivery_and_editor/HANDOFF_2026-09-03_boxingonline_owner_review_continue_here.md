# HANDOFF 2026-09-03 — boxingonline.com owner review: continue here

**This is the cold-start for the OWNER-REVIEW thread on boxingonline.com** — the first paid
customer build (`d2aa5206-73bc-4707-a69c-2702c1eb9152`, order BR-9AUZ59). It is not the delivery
pipeline thread; that is `site_delivery_and_editor`'s own handoff series, and they own every
dispatch at this site. This thread's job is **reading the site against the owner's critique,
routing findings to the lanes that can fix them, and verifying claims at the artefact.**

Everything below was **measured 2026-09-03 between 10:30 and 10:45Z**, post-roll. Figures carry
their date because they go stale by addition.

---

## 0. READ THIS BEFORE MEASURING ANYTHING

Get these wrong and every check returns a confident, meaningless answer. Every one cost real time.

- **Serving host is `boxingonline.ugg2.com`** (`sites.publish_target='b2worker'`,
  `publish_project='boxingonline.ugg2.com'`). **`boxingonline.com` is the customer's own domain,
  NOT yet pointed at us** — it is a parked catch-all that answers 200 for any path with a
  114-byte stub. **Never probe it.** Control: an invented path on the slug returns 404 (verified
  10:42Z), so status codes ARE meaningful there.
  > A live `site_unreachable` item (`33a900b8`) is a FALSE POSITIVE for exactly this reason — the
  > checker probes `https://boxingonline.com/` and gets a DNS failure. The site is fine.
- **Every probe carries a must-be-present control.** A zero control means the fetch failed —
  report BLIND, never "clean".
- **Enumerate pages from the DB** (`pages WHERE deployed_at IS NOT NULL`), not from memory.
- **A completed work item is not a changed artefact.** Nor is a fresh `deployed_at`.
- **Date the artefact before accepting "latency":** published object OLDER than your change → wait.
  NEWER and still wrong → look upstream.
- **`</dev/null` on any inner `kubectl exec -i` inside a `while read` loop**, or it eats the loop's
  stdin and the sweep reports one row, cleanly.
- **Postgres regex: `\y` not `\b`** (`\b` is backspace; returns a zero that reads as an absence).

**The general lesson, and the reason this thread produced anything: every defect found here shipped
on a site where every page validated `valid=true, issues=0`, every work item completed, and every
build reported success. None announced itself.**

---

## 1. STATE OF THE SITE — measured 10:42Z, post-roll

Running: `agent-chassis-75b987cbd7`, both pods started 2026-09-03 08:57–08:58Z.

| page | imgs | email | contact links | GTM | control |
|---|---|---|---|---|---|
| `/index.html` | 7 | 0 | 0 | 0 | 19 |
| `/guides/index.html` | 1 | 0 | 0 | 0 | 7 |
| `/articles/index.html` | 21 | 0 | 0 | 0 | 7 |
| `/news/index.html` | 1 | 0 | 0 | 0 | 7 |
| `/tools/fight-calendar/index.html` | 1 | 0 | 0 | 0 | 7 |

`/contact.html` → **404** · invented path → **404** (so the 404 discriminates).
Cards on `/index.html`: 6 cards, **0 excerpt elements**, **6 suffixed titles**.

**VERIFIED DONE (owner's fourteen points):** email off every page · nothing links to contact ·
contact page 404 · logo-only header · guides index lists the 4 guides with decks · "Free Cost"
gone · one "News" in nav · fight calendar in the menu · logo carries no invented wordmark ·
six articles built and linked · footer regenerates properly (machine render, digest set).

**STILL WRONG ON THE SITE:** articles contain no news · comparator ships no data · fight calendar
has no calendar · news page serves raw feed residue · guides unrewritten (3,530–4,136 chars) ·
imagery thin (logo-only on every page except the two listings) · card decks not applied to
`/index.html`.

---

## 2. WHAT IS IN FLIGHT RIGHT NOW — the site's queue turn has ARRIVED

The build trigger serves the site holding the oldest eligible item. boxingonline reached
**position 2** and `ec92320f` was **CLAIMED at 10:42:15Z**. So the rest should follow in minutes,
not hours.

| item | type | state @10:42 | what it does |
|---|---|---|---|
| `ec92320f` | needs_rerender | **CLAIMED** | operator chrome refresh — should land GTM |
| `2d1f9c51` | needs_page | triaged | **the CTA rebuild** — the one this thread is waiting on |
| `d71b7877` | needs_imagery | triaged | the owner-authorised transparent-logo test |
| `aba585f5`, `06210ec6` | page_rerender | triaged | |
| `0cdddb6f` | stale_attestation | needs_human_review | **FALSE ALARM — see §5** |
| `33a900b8` | site_unreachable | detected | **FALSE POSITIVE — wrong host, see §0** |
| `c5614b00` | needs_page | failed | the 3× shrink-refused rebuild, superseded by `2d1f9c51` |

### 2.1 THE JOB THIS THREAD OWES WHEN `2d1f9c51` LANDS

**Read the regenerated copy against these seven checks** (agreed with the delivery lane; they will
send the row read and point at this verdict rather than duplicate it):

1. **A control-labelled field must read like a control.** Flag any `button`/`label`/`*_cta` field
   over ~120 chars. The current CTA is ~1,116. **Shorter is necessary and NOT sufficient — the
   acceptance is that it reads like a button, not a shorter essay.**
2. **Meta-copy:** `grep -Eci "we write|we'd rather|we cover|gets checked|we update|the list below"`.
3. **AI tells:** `plainly`, `honest`, `before your .* have to`, `starting point, not the final
   word`, and a CTA naming three or more tools in prose (the current one names four).
4. **Placeholders:** `placehold`, `EXAMPLE`, `TODO`, `_ADDRESS`, `Lorem`, ALL_CAPS_UNDERSCORE.
5. **Listing class:** home listing must stay 6 `/blog/` links, 0 `/guides/`.
6. **Empty slots:** post-682 a FILLED deck is the success state — do not file filled decks as a
   regression, and do not read collapsed slots as filled decks (see §4.2).
7. **Regression guards every time:** email 0 · contact links 0 · contact 404 · one fight-calendar
   ref per header · control non-zero per page.

**And the judgement no grep does:** the owner's item 1 is "telling the user what the site is doing
rather than talking to the user". A 167-character CTA that still explains what the section below
contains has been **abbreviated, not repaired**. Say so either way.

### 2.2 A PRE-REGISTERED PREDICTION — do not reinterpret it afterwards

Stated 2026-09-03 ~09:55Z, before the run, and banked by both the delivery and components lanes:

> **`2d1f9c51` is a BUILD, and all three build-path listing instances fleet-wide carry the new card
> shape. So it should produce decks with real copy, suffix-free headlines AND suffix-free alt text,
> with nothing further done. If it lands successfully with suffixes intact, the model is wrong.**

Alt text and headline interpolate the same value, so they must move together. **Suffix-free
headlines with suffixed alts would be a third outcome nobody predicted** — report it as a new
finding, not a pass.

---

## 3. OPEN WORK, BY OWNER — do not duplicate

| what | who | state |
|---|---|---|
| Every dispatch at this site, delivery chain, the 725 shrink-floor window | `site_delivery_and_editor` | active |
| Card decks / suffix — **why one instance re-resolves and another does not** | `components` lane | active, see §4.1 |
| Articles with no news; guide rewrites; the title-promise gate | `copy_quality_two_stage` | active |
| Fight calendar + comparator data; the research mechanism | `bugs_open/427` | **session ended — needs restarting** |
| Card visual design | offer analyser / visual designer lane | active |
| Article-header imagery (assets exist, no component can show them) | `editorial_design_uplift` | active |
| Transparent logo | `bugs_open/424` | fix aboard v1.0.1356; test queued here |
| Raw feed residue on the news page | `bugs_open/332` addendum | **unstaffed** |
| Nightly listing/tool/nav checks | `experience_loop` (rules A/B/C live) | running |

---

## 4. THE TWO LIVE MYSTERIES

### 4.1 Why the card fix reaches some listings and not others
**NOT the path.** The "build path works, rerender path doesn't" model is **REFUTED** — two
`page_rerender`s with `section_data_resolved` produced the NEW shape (designblog, websitepromotion)
while a third produced the OLD one (boxingonline/index). Same path, same reason, opposite outcomes.

**Eliminated by measurement** (mine and the components lane's, across all 17 content-listing
instances): path · reason · component · source · binary · time-of-run · locks ·
`component_version_id` · `content_item_id` · `schema_mode` · `build_status` ·
`built_from_plan_version` · `suppressed_sections` · per-section wiring (`pages.sections` is a flat
array of slot-name strings — nothing there to differ) · empty-resolve/eligible-population ·
overnight config change (`content-listing` last updated 09-02 10:43, before all three rerenders).

**Best remaining lead:** `garden-tools.uk` — `/index` NEW and `/care` old, same site, same
component, same source, same 4 stored items, same 4 eligible posts. Only the writer differs.

### 4.2 A measurement trap that nearly produced a false pass
Counting **empty** deck elements (`article-card__excerpt"></p>`) returns 0 **both** when the slot
collapsed (682) and when the deck is filled. Count the **element itself** and its inner length:
`0 elements` = collapsed · `6 with content` = filled · `6 empty` = the pre-682 defect.

---

## 5. TWO FALSE ALARMS HEADING FOR THE OWNER'S REVIEW

- **`0cdddb6f` stale_attestation** — claims a fact is overdue after 180 days on a site three days
  old. Cause: order-intake seeds `business_name` with `source.attested_by` and **no date**, and the
  checker treats undated as beyond cadence. **Fleet singleton: 146 attested facts, exactly 1
  undated, and it is this one.**
- **`33a900b8` site_unreachable** — probes the parked customer domain (§0).

---

## 6. THE GOOD NEWS NOBODY EXPECTED

**`evidence_base` went from 1 fact to 7**, written by `evidence-researcher` (09-02 12:41) and
re-run by `evidence-refresher` on schedule (09-03 09:11). Real, dated, cited facts with URLs and
verbatim quotes — including **"Canelo Alvarez is scheduled to fight Christian Mbilli on October
31"**, a forward-looking dated fixture, which is exactly the shape `bugs_open/427` says nothing
produces.

**But nothing consumes it.** Calendar still 0 inputs / 0 data / 0 fetch; Canelo/Mbilli appears only
on `/news/index.html` and those are the FEED's items. **427's title is now half false — restate it
before building, or the working half gets rebuilt.** Addendum committed to the bug.

---

## 7. ARTEFACTS THIS THREAD PRODUCED

| path | what |
|---|---|
| `site_delivery_and_editor/APPROVAL_READOUT_2026-09-02_boxingonline_what_is_actually_fixed.md` | **the owner's approval document** — three columns (verified / not fixed / built-but-inert) |
| `docs024_key_docs_latest/SITE_DEFECT_CATEGORIES.md` | **fleet-wide acceptance checklist**, written at the owner's direction; ~30 categories each with a runnable check. Other lanes are already appending |
| `site_delivery_and_editor/COMPARISON_2026-08-31_…why_theirs_looks_better.md` | ours vs the other builder's page |
| `site_delivery_and_editor/OWNER_REVIEW_2026-08-31_…what_each_finding_actually_is.md` | the original fourteen-point review |
| `static_site_form_endpoint/PLAN_2026-09-02_pre_plan_extensible_form_endpoint.md` | the owner's 1b pre-plan, D1 reviewed by the seam owner |
| `bugs_open/427` addendum, `bugs_open/332` addendum | the research gap and the feed-markdown gap |

---

## 8. OWNER RULINGS IN FORCE

1. **Fix everything before the delivery email.** No post-delivery bucket for this site.
2. **Best-in-class propagation approved** — `copy_quality_two_stage` building.
3. **No contact email or address on this site at all.** Contact page deleted; opt-in on explicit request only.
4. **Header is logo-only** — no wordmark beside the mark.
5. **A guides-index page on every site**, and guides rewritten to be more interesting, shorter where there is little to say.
6. **The palette stays** — "the cream off white decision is fine".
7. **A logo's background must not be baked into the logo asset.**
8. **Identity: whoever commissions a site is independent of the site's operation** — the identity model needs replumbing (`bugs_open/420`).
9. **Scoped override, one run** for the shrink floor (migration 725, claim-gated window).
10. **Test the transparent logo on boxingonline** — authorised 2026-09-03.

---

## 9. IMMEDIATE NEXT ACTIONS

1. **Watch `2d1f9c51`.** When it completes, run §2.1's seven checks and write the owner-facing
   verdict on his items 1 and 14. Report whether §2.2's prediction held — including if it did not.
2. **Verify the logo test (`d71b7877`) AT THE BYTES** — PNG colour type 6 or 4, **or** a `tRNS`
   chunk; test for both, either alone gives a false negative. **Not** at the adapter's
   `border_keyed`, which scored 1.000 on a 0.0%-transparent failure. Interim solid mark stays
   serving until the bytes say otherwise. `seotools` came back colour type 6 / 92% transparent —
   the first real success, so the matte works and the guard is the blind part.
3. **Confirm GTM lands** once `ec92320f` completes (`GTM-PQ3WCTBD` currently 0 on every page).
4. **Get a session onto `bugs_open/427`** — it is the root of the owner's items 7, 9 and 10 and its
   previous session has ended. Restate the title first (§6).
5. **Nobody re-seeds this site** until `420`'s class fix rolls — `build_queue.direction` still holds
   the ordering email and `sites.email` is empty, so a re-seed refills it. (`b60d66e3c` shipped the
   *mirror* fix, not this one.)
