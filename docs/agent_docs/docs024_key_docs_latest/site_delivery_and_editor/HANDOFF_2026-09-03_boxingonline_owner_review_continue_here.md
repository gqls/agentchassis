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
| Transparent logo | ~~`bugs_open/424`~~ → **`bugs_closed/424_HANDOFF_2026-09-02_transparency_is_not_a_promptable_property_so_the_model_paints_a_checkerboard.md`** | **CLOSED.** Matte AND its guard live on v1.0.1356 (not merely committed). boxingonline verified at the SERVED bytes. ⚠ Re-pointed 2026-09-03: `bugs_open/424` resolves to nothing — verified at HEAD with `git ls-tree`, exactly one path, so the move is complete and not a pathspec-commit copy |
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

---

# UPDATE — end of 2026-09-03. §§1, 2 and 4 above are SUPERSEDED by this.

The site's queue turn came at 10:42Z and again at ~12:07Z; the publish tick landed 13:14Z.
**Full end-of-day state, measured at the served site, is in
`APPROVAL_READOUT_2026-09-02_boxingonline_what_is_actually_fixed.md` under
"STATE AT END OF DAY 2026-09-03"** — read that for the owner-facing verdict. This section records
only what a NEW SESSION needs that the read-out does not carry.

## The seven checks in §2.1 HAVE BEEN RUN. Verdict: pass, one live defect.

Against `/index.html` lm 13:14:20, cache-busted, control 7, invented path 404:
CTA block **210 visible chars** (from 1,347) · **0** register tells · **0** AI tells · **0**
placeholders · 6 blog links · **6 deck elements, 0 empty** · **0 suffixed headlines, 0 suffixed
alts** · email 0 · contact links 0 · one fight-calendar header entry · GTM ×2.

**THE ONE LIVE DEFECT: "…and the calendar below tells you what's coming up next."** The CTA is the
last section (`content-listing → info-card-grid → call-to-action`) and there is no calendar on the
page. **Fix filed `needs_copy_edit a930e70c`, ends at `checkpoint_for_review` → it lands in the
OWNER'S ADMIN QUEUE for approve / request-changes.** Do not hand-edit `content_data`; the
framework writes the content.

## §2.2's PRE-REGISTERED PREDICTION HELD — and one inherited claim must be narrowed

A BUILD produced decks with real copy, suffix-free headlines **and** suffix-free alts, nothing
further done. Builds are now **4 for 4**.

⚠ **Do NOT inherit "the path-split model stands".** Its other half — *rerenders never produce the
new shape* — was refuted (two rerenders elsewhere did), and the components lane's batch 691 has
since shown a rerender **PRESERVES** the shape on a new-shape baseline. So: **builds reliable;
rerenders inconsistent; cause still unknown.** §4.1's elimination list and the `garden-tools.uk`
lead both stand. That 691 result is also what unblocked the CTA fix — it proved correcting item 1
cannot revert item 14.

## New traps for §0

- **`pages.rendered_head` reads GTM = 0 on all 21 pages while the served pages carry it.** The head
  is composed from `site_components` at assemble time. **Checking that column would report the
  analytics rollout as failed when it succeeded.**
- **A `rerenders_since` count on a component whose `updated_at` has not moved is EMPTY of meaning**
  — an unmoved timestamp proves those rerenders never wrote that row, so the number describes
  rerenders elsewhere on the site. I nearly reported "survived 19 rerenders" as good news. Kept out
  of every file deliberately; do not reintroduce it.
- **Chrome refresh `ec92320f` FAILED 3/3** (`bugs_open/457`) — chrome propagation on this site is
  hand-filed until that code fix.
- **`/about.html` GTM = 0**: published ~20 s before its own assemble finished. Queued, not failed.

## What is still open, unchanged from §3

Items **7, 9, 10** (articles / comparator / calendar) — re-measured 13:59Z, all unchanged, all one
root cause in **`bugs_open/427`, whose session has ENDED and needs restarting.** Its writer half
now works here (`evidence_base` 1 → 7 real cited facts, including a dated forward fixture) and
nothing consumes them — **restate the bug's title before building or the working half gets rebuilt.**
Also open: **12b** news feed residue (5 md links / 12 truncations / 11 UFC), **3b** guides
unrewritten, **8** imagery logo-only on every page but the two listings.

## Immediate next actions, replacing §9

1. **Watch the owner's admin queue** for `a930e70c` — the CTA rewrite needs his approve/request-changes.
2. **Re-run the seven checks after that rewrite lands**, and re-read the deck keys afterwards
   (the 691 result says they survive, but verify rather than inherit).
3. **Get a session onto `bugs_open/427`** — root of items 7, 9, 10; restate the title first.
4. **`/about.html` GTM** — confirm on the next publish tick.
5. **Nobody re-seeds this site** until `420`'s class fix rolls.

---

# UPDATE 2 — 2026-09-03 17:08Z. Supersedes §4.1 entirely, and adds a NEW site defect.

## Served state now — measured 17:08Z, cache-busted, control non-zero on every row

| page | cards | decks | empty deck | empty category | suffix | GTM |
|---|---|---|---|---|---|---|
| `/index.html` | 6 | 6 | 0 | 0 | 0 | 2 |
| `/articles/index.html` | **36** | 36 | **2** | **14** | 0 | 2 |
| `/guides/index.html` | 4 | 4 | 0 | 0 | 0 | 2 |
| `/news/index.html` | 0 | 0 | 0 | 0 | 0 | 2 |
| `/about.html` | 0 | 0 | 0 | 0 | 0 | 2 |

`/about.html` GTM has caught up — the whole site now carries the tag.

## §4.1 IS CLOSED. The cause was ONE MISSING ASSIGNMENT.

`bugs_open/454`: when a long function was split on 2026-09-02, `classifyStoredSection` computed a
section plan and **never assigned it to the struct it returns**, so `renderPlannedSection` read a
zero value and the render composed `base ⊕ stored content_data ⊕ nil`. For a fortnight **every
light re-render on the estate rendered each page's own stored data back at itself and reported
success** — no errors, healthy counts, nothing blanked. One assignment (`9831e9ab4`), live in
v1.0.1358 since 12:18Z. Filed by the `bugs_open/427` lane from a completely unrelated symptom.

**Do not re-open the per-instance hunt in §4.1.** Its sixteen eliminations are individually true
and its `garden-tools.uk` pair is what answered the bug — but the difference it was hunting **does
not exist.**

⚠ **Two corrections that must not be inherited backwards:**
- **This thread's original framing was RIGHT** — "the rerender path does not resolve", and "a
  correct reason is necessary and NOT sufficient" (rows completed twice with `template_changed` and
  produced the old shape).
- **The components lane's refutation of it was WRONG, and they have logged it against themselves.**
  The two rerenders said to have "produced" the new shape on designblog/websitepromotion had not: a
  **BUILD** wrote `excerpt` into stored on 09-02 20:51 and 23:02, and every later re-render merged
  `stored ⊕ nil` and carried it forward. The disconfirming column was in the same
  `page_component_history` rows already being read — archived `content_data` states what each write
  REPLACED.

Deck class is draining on its own — **10 new / 7 old of 17** at 15:15Z, from 5/12 the day before,
because any re-render for any reason now fixes it. **No repair wave is needed or filed.**

## NEW DEFECT ON THIS SITE — `/articles/index.html` serves the six articles SIX TIMES

**36 cards where there should be 6.** Six distinct `/blog/` URLs, 36 link occurrences. And because
each duplicate is frozen at the template of the day it was created, they are not even alike: **34
carry decks, 2 are empty, 14 are missing their category label.**

At the source — `page_components` for `articles-index`:

```
position  component_id  slot                 rendered_html  created_at
   1      set           hero                     3,658      08-31 14:30:43
   2      set           generic-text-block         NULL     08-31 14:30:43
   3      set           call-to-action           3,049      08-31 14:30:43
   3      NULL          generic-text-block       4,429      08-31 16:29:47   <- orphan
   3      NULL          generic-text-block       4,638      08-31 18:14:32   <- orphan
   3      NULL          generic-text-block       5,195      09-01 01:31:57   <- orphan
   3      NULL          generic-text-block       6,335      09-01 01:58:36   <- orphan
   3      NULL          generic-text-block       6,334      09-01 02:34:51   <- orphan
   3      NULL          generic-text-block       6,070      09-02 16:28:02   <- orphan
```

**Six orphan rows stacked at position 3, `component_id` NULL, each holding its own frozen
`rendered_html`, all six assembled into one page.**

**`bugs_open/457`. NO RE-RENDER WILL FIX IT** — `component_id` NULL means `resolveComponent` misses
and the carry branch runs, so 454's fix cannot reach these rows. **It needs 457's code fix and a
rebuild.** Nothing has been touched: paid site, wrong instrument.

## `bugs_open/424` is CLOSED and MOVED

→ `bugs_closed/424_HANDOFF_2026-09-02_transparency_is_not_a_promptable_property_so_the_model_paints_a_checkerboard.md`
(verified at HEAD with `git ls-tree`: exactly one path, so a complete move, not a pathspec copy).
**Matte AND guard live on v1.0.1356, not merely committed.** All four regenerated sites carry PNG
colour type 6 at the served bytes: seotools 92.21%, designblog 88.5%, websitepromotion 87.4%,
boxingonline 80.10%. §3's row is re-pointed.

## Revised immediate next actions

1. **The owner's admin queue** — `needs_copy_edit a930e70c`, the "calendar below" rewrite, awaits
   his approve / request-changes. Re-run the seven checks after it lands.
2. **`bugs_open/457`** — the 36-card duplication on `/articles/index.html`. Unstaffed, code fix +
   rebuild, and it is the most visible remaining defect on the site.
3. **`bugs_open/427`** — root of items 7, 9, 10. Session ENDED. **Restate the title first:** its
   writer half now works here (`evidence_base` 1 → 7 cited facts incl. a dated forward fixture) and
   nothing consumes them.
4. **`bugs_open/332`** — news feed residue (5 md links / 12 truncations / 11 UFC). Unstaffed.
5. **Nobody re-seeds this site** until `420`'s class fix rolls.

---

# UPDATE 3 — 2026-09-03 18:00Z. The seven checks are RE-RUN and PASS. Action 1 is CLOSED.

## Do not go looking for `a930e70c` in the owner's admin queue — he already acted, at 16:21.

UPDATE 2's action 1 (*"the CTA rewrite awaits his approve / request-changes"*) was **stale when it
was written**. Reconstructed from the DB:

```
a930e70c needs_copy_edit     complete 14:24  approval_mode=auto   ← detection, correctly auto
  └─ 1cb907ee copy_edit_proposed complete 14:24→16:21             ← the review he actually saw
       └─ OWNER APPROVED 16:21   spec.approved_by='admin', 2 edits, each with owner_ruling text
            └─ 5edadfbe section_edit  FAILED 3/3, last attempt 17:56Z
               "step load_edit_context failed: need either page_component_id
                or both page_name + slot_name (CHILD_ORCHESTRATION_FAILED)"
```

**His edits are live on the page regardless** — `page_components` 322ce532 (`content-listing`) and
e5b848fa (`call-to-action`) were updated **16:26 / 16:27**, by hand, five minutes after he pressed
approve and while the job was failing. That is this lane's own hand-application, described in
`README_where_we_are.md`.

⚠ **`5edadfbe` is still sitting `failed` at attempt 3/3 while representing work that is DONE.**
Someone should reconcile it or it reads as outstanding for ever. Its `spec` is the textbook case of
the approve-button defect this lane fixed in `33dfeed3a`: `page_component_id` **is** present but
nested inside `approved_data.edits[]`, while top-level `copy_edit` and `page_target` are `null` —
which is where `load_edit_context` looks. Two edits in one job also hits the one-target-per-job
limit. **Both halves are fixed and inert until a roll.**

## §2.1's seven checks — PASS. Served `/index.html`, Last-Modified 17:32:30Z, cache-busted.

| # | check | result |
|---|---|---|
| 1 | CTA reads like a control | **96 visible chars** (1,347 → 210 → **96**); heading + 2 imperative button labels |
| 2 | register tells | **0** |
| 3 | AI tells | **0** |
| 4 | placeholders | **0** / **0** / **0** |
| 5 | listing class | 6 distinct `/blog/`, 6 occurrences; both `/guides/` hrefs are **FOOTER**, listing is pure `/blog/` |
| 6 | deck slots | **6 elements, 0 empty**, inner 103–153 chars — filled |
| 7 | regression | email **0** · contact links **0** · `/contact.html` **404** · **1** header fight-calendar ref · GTM ×2 — on **20/20** served pages |

Controls: three invented paths **404** while `/index.html` is **200**; `<body>` non-zero on every
200. Pages enumerated from `pages WHERE deployed_at IS NOT NULL` (22 rows).

**The judgement no grep does — owner item 1: REPAIRED, not abbreviated.** `grep -ci 'calendar
below'` → **0**; the subheadline was *emptied*, not reworded, which is what he ruled. The block now
reads *"Stay on top of every fight that matters / Catch the latest boxing news / See the full fight
calendar"* — every clause addressed to the reader, none describing what the section below contains.
The listing subtitle is his verbatim line, *"News, previews and results from across the sport."*

**Owner item 14 — deck keys VERIFIED, not inherited** (UPDATE 2 asked for this rather than trusting
batch 691): 6 card headlines and 6 card alts agree character-for-character, **0 template-suffixed**
either side. ⚠ **Use an anchored suffix pattern** — see the new LANDMINES entry; a `\s[|\-–—]\s*\S`
test reports **3 of 6 alts suffixed** on this correct page, because the headlines contain editorial
em-dashes, and that false positive wears the exact shape §2.2 tells you to escalate as a third
outcome.

## Two corrections to UPDATE 2's own action list

- **`bugs_open/457` is NOT unstaffed** — 7+ commits today across `site_delivery_and_editor` and the
  components lane; `scripts/who-owns.py 457` names the owning lane. Contribute into the bug file,
  do not compete. Re-measured 17:32Z: `/articles/index.html` still **36 cards / 36 decks / 2 empty**,
  unchanged, as expected for a defect needing a code fix + rebuild.
- **`bugs_open/332` and `bugs_open/427` are both being worked** — 332 has a commit at HEAD and 427's
  file was written at 17:39Z. Check `git log` on the file paths before adopting either.

## New for §0's trap list

**`/guides/tool-fight-calendar-guide.html` 404s while `pages` says `deployed`** — this is **latency,
not a defect.** The row was **created 17:35**; the site's served publish is **17:32**. Published
object older than the change ⇒ wait. Worth a second look only if it survives the next publish, since
`build_status='deployed'` sits on a row whose `last_built_at` is **NULL**.

## Immediate next actions, replacing UPDATE 2's

1. ~~Watch the admin queue for `a930e70c`~~ — **DONE, approved 16:21, applied by hand, verified at
   the served bytes.** Nothing to watch.
2. **Reconcile `5edadfbe`** (failed 3/3, work actually complete) — or it misreports for ever.
3. **`bugs_open/457`** — the 36-card duplication, still the most visible defect on the site. Staffed;
   join, don't duplicate.
4. **`bugs_open/427`** — root of owner items 7, 9, 10. Restate the title first (§6).
5. **Nobody re-seeds this site** until `420`'s class fix rolls.

---

# UPDATE 4 — 2026-09-04. A CORRECTION TO THIS FILE THAT COST ANOTHER LANE A WRONG TURN.

## ⚠ "`bugs_open/427`'s session has ENDED and needs restarting" is FALSE, and has been since 20:00Z.

That claim appears in §3, in UPDATE 1, in UPDATE 2's action 3 and in this file's action lists — four
places. **The 427 lane restarted the same evening and is active:** `0e086d854` 20:00, `527961354`
20:13, `f35a6dfd5` 20:14, `a12d787d0` 21:19 (s22.4 → s23), working verified-facts plans with
negative and positive controls.

**It propagated.** The `bugs_open/463` lane went looking for a live 427 session, was routed here by
site-design-planner, and arrived asking this thread questions on the strength of it — plus a second
misattribution (migration **687** is `05905ebcb`, the **428** lane's build-site-planner prompt work;
nothing in this thread shipped a migration). *Do not route 427 work at this thread.* Reach the 427
lane directly, and **check `git log --since` on `bugs_open/427*` before repeating any status from
this file** — "session ended" is the single most perishable claim a handoff can carry, because
nothing updates it when the session comes back.

## Census run answering that lane's question, recorded here so the number carries its date

`[MEASURED 2026-09-04]` `parent_section` is persisted on **`site_plan_pages`**, not on `pages`
(`pages` has no such column; `datahelpers.ParentSectionFromURL` exists to recover the directory from
the URL instead). `site_plans` is full history — 56 plans, 2026-05-12 → 2026-09-04, 21 superseded
and retained — so this is a complete census, not a rolling window.

- **881 planned pages; 860 (97.6%) declare no `parent_section`.**
- The 21 that do are **four values on two days**: `archetypes` (8) and `tools` (1) on 2026-06-22,
  `glosario` (8) and `guias` (4) on 2026-07-16.
- **Since 2026-07-16: 44 plans, 702 pages, ZERO declarations.**

The 463 lane measured independently that the live 32,191-char plan_site prompt never contains the
string `parent_section`. The two together say the field is **undriven, not disputed** — the model is
never asked, so production holds no sample of what it would honour. And the only three occasions it
ever named a section unprompted it **invented vocabulary** (`guias`, `glosario`) beside an estate
whose head is `guides`.

⚠ **Do not compare 881 against the estate's 1,362 pages.** `site_plan_pages` is planning intent;
pages created by adoption or by `deploy_tool_action` never appear in it.

## This site is a worked example of the directory problem, if anyone writes that RFC

```
blog-index     /articles/index.html                      <- children are NOT in its directory
blog-post      /blog/<six>.html
blog-post      /guides/tool-fight-calendar-guide.html    <- role differs from its four siblings
guide          /guides/tool-<four>-guide.html
section-index  /guides/index.html
```

A `blog-index` at `/articles/` whose every child is `/blog/` — stem-matching index to children
returns **zero**. And the four `role=guide` companions sit in the same directory as one
`role=blog-post` companion produced by `deploy_tool_action.go:739`, which hardcodes
`Role: "blog-post", ParentSection: "guides"`.

**A second candidate cause for the `/guides/tool-fight-calendar-guide.html` 404 in UPDATE 3.** I put
it down to latency (row created 17:35, site published 17:32) and that remains the simpler
explanation. But it is also **the only companion guide on this site with a different role**, and the
only one not serving. `[UNVERIFIED]` — two candidate causes, not one. Watch it on the next publish
tick; if it is still 404 after a publish that postdates 17:35, latency is refuted and the role
difference is where to look.

---

# UPDATE 5 — 2026-09-04 14:36Z. §0's description of the customer domain is WRONG as of today.

## `boxingonline.com` does not answer anything. It does not resolve at all.

§0 says it "is a parked catch-all that answers 200 for any path with a 114-byte stub" and builds the
`site_unreachable` false-positive ruling on that. **Measured 2026-09-04 14:36Z, that is not the
state.** The domain is registered (Dynadot, created 2023-10-05) and delegated to Cloudflare
nameservers `ALEXIS`/`LEAH.NS.CLOUDFLARE.COM` — and **those exact nameservers return `REFUSED` for
it**, so public resolution is `SERVFAIL`. That is a **dangling delegation**: the registrar points at
Cloudflare, but no active zone for this domain exists in the account those nameservers serve.

**Control, in the same breath** — the same two nameservers return `NOERROR` for
`boxingonline.ugg2.com`, and `dig` returns records for `google.com`. So the empty answer is the
domain's, not the instrument's.

**The CONCLUSION in §0 still stands and the ruling does not change:** `site_unreachable` is still a
false positive *about the site*, which serves fine at the slug. But the stated REASON is wrong, and
the corrected reason matters more, because "a parked page answers 200" and "the domain does not
resolve" imply completely different next actions. ⚠ **A fresh item has since fired —
`55df8d35 site_unreachable unresolved` at 09-04 13:08**, a different id from §2's `33a900b8`. Expect
it to keep re-firing until the zone exists; do not re-rule it each time without reading this.

**Going live at the customer's domain is therefore blocked on a Cloudflare zone**, not on the build.
This lane already owns the tool: `cf_customer_domain_zone.sh` (dry-run default, `--apply` executes),
which is severable by design — the site keeps serving at the slug regardless. **Not run here; that
is the delivery lane's call and the owner's, and ruling 1 (fix everything before the delivery email)
is still in force.** `sites.email` remains empty and must stay so until `420`'s class fix rolls.

## Serving state, re-measured 14:36Z — the site IS the latest build

Whole site republished **12:12:50–12:13:20Z today**; **zero** page writes and **zero**
`page_components` writes since. Nothing is waiting to publish.

- Yesterday's verified fixes all survive the republish: owner's subtitle verbatim ×1, `calendar
  below` **0**, **6 deck elements / 0 empty**, email **0**, contact links **0**, GTM ×2.
- **`/guides/tool-fight-calendar-guide.html` now serves 200.** UPDATE 4 left two candidate causes for
  its 404; **latency is CONFIRMED and the role-difference hypothesis is REFUTED.** No further action.
- `/contact.html` still 404 — correct, ruling 3.
- **`bugs_open/457` unchanged: 36 card titles, 6 distinct `/blog/` URLs** on `/articles/index.html`.
  Still the most visible defect, still needs the code fix plus a rebuild.

⚠ **The serving host is `boxingonline.ugg2.com`.** `boxingonline.com.ugg2.com` — the domain with the
slug appended — **does not resolve**; it is an easy mis-type that returns a DNS failure looking like
an outage.

---

# UPDATE 6 — 2026-09-04. Why the owner's complaints were never actioned, and a CORRECTION to my own claim.

## The answer: the findings exist, and the loop is switched off BY DESIGN.

Every defect the owner raised on 2026-09-04 was already written up as a verdict row on this site,
days earlier, in near-identical words — and deliberately not dispatched.

| owner's complaint (2026-09-04) | verdict row, filed | measured at the served bytes 2026-09-04 ~15:00Z |
|---|---|---|
| the "last night's result" article is process, not a story | `48c0f927` 09-01 | general essay on upsets; worked example is **Douglas/Tyson 1990**; no recent fighter, date or score |
| the comparator "isn't populated" | `2c38eec5` 09-01 | **0 `<option>`, 0 `<select>`** |
| the countdown "has previous fights in it" | (calendar ×4: `3acce370` `dd07d0c7` `9877c9a0` `70e4ed6c`) | only ISO date in page is **2026-09-02**, two days PAST; matchups are archetypes (Joshua/Wilder, Usyk/Fury) |
| "latest articles section is still duplicated" | `462ac4da` 09-01 (says **four** times) | **six** times now — `bugs_open/457`, growing |

**Mechanism, from the code's own spec fields** (`write_audit_findings_action.go`, `recordOnlyFinding()`):
`filing_mode=record` (RFC_056) → `status='deferred'`, `handler_agent=''`, summary prefixed
`[verdict, not dispatched]`, and `not_dispatchable = "…deliberate; … nothing promotes this row"`.
Seven seats live in record mode (`brief-fidelity-auditor`, `content-quality-auditor`,
`improvement-loop`, `offer-analyser`, `reader-experience-auditor`, `site-review-agent`,
`visual-design-auditor`). Each row carries `spec.release_recipe`, a one-line UPDATE.

`[MEASURED 2026-09-04]` **46** parked content verdicts on this site, **58** of all types;
**3,181 across 39 sites** (the `site_delivery_and_editor` lane read **3,184** within the hour —
the gap is GROWING, not a disagreement).

## ⚠ CORRECTION to my own claim, caught by `site_delivery_and_editor`

I told that lane, and implied here, that **"RFC_056 §7 phase 3 = switch the loop back on"** and that
releasing was therefore a queued phase to defer to. **That is FALSE.** Verified in the RFC's own
phasing table: **phase 3 IS what put the seats into record mode** — *"`filing_mode: record` on all
five model seats' write steps"*, shipped as
`624_acceptance_council_seats_record_only_and_seat_failures_are_rows_HOLD.sql`. The filename says
`record_only`. **There is no later phase that releases anything.** The RFC states the exit only as
prose: a record row is *"undispatchable by construction until a person or a deliberate migration
releases it"* and is *"designed for a person to release"*.

**Where my error came from, because it is the reusable part:** I read the *plan document's filename*
— `PLAN_2026-08-25_switch_off_the_evolutionary_rewrites_and_switch_the_loop_back_on.md` — as if it
named a phase. It names the whole plan's ARC, of which only the switching-off half has been built.
**A filename is a title, not a status.** I had the phasing table in front of me and quoted it in the
same breath.

This **strengthens** the case for holding rather than weakening it: there is no plan queued behind
the release to defer to, so releasing is a **new owner decision**.

⚠ `schema_migrations` has **no row for 624** — but the effect IS live (7 seats measured directly in
`agent_definitions`). A `_HOLD` file is applied by hand, and hand-applying and recording are two acts.
**Verify record mode at `agent_definitions`, never at `schema_migrations`.**

## The reframing that makes this bigger than this site (`site_delivery_and_editor`, verified by me)

This is **not** confined to sites built before the switch. `[MEASURED 2026-09-04]` of **12** sites
created on/after 2026-08-25, **11** carry parked verdicts. The newest, **`copyonline.co.uk`
(created 09-03), already has 29**, the newest filed 09-04 11:09Z — hours old. The owner holds
voucher `WD-KN3WU-9PZN4` for another full webdesign.uk run, so the next site will be audited, its
defects recorded, and none repaired: boxingonline's end state, reached freshly.

**So the question to put to the owner is not "clean up boxingonline" but "what happens to every site
we build from here".**

## If he rules to release, sequencing is the hard part, not the SQL

`find_dispatchable_site` is `ORDER BY MIN(created_at) ASC LIMIT 1` on a 30-second tick — fleet-wide
FIFO. 46 rows fired behind ~3,138 others is a **queue position, not a batch**. Cost it before firing.
And a rewrite dispatched today would mostly reproduce the same output: `evidence_base` holds 7 real
dated cited facts (incl. a forward fixture, Canelo v Mbilli 31 Oct) and **nothing consumes them**
(`bugs_open/427`, title now half false). **Close the consumption gap first or pay twice.**

Cross-recorded by that lane as §1.3b of
`docs/agent_docs/docs024_key_docs_latest/site_delivery_and_editor/HANDOFF_2026-09-04_continue_here.md`.

---

# UPDATE 7 — 2026-09-04 ~16:10Z. TWO CORRECTIONS TO UPDATE 6, both caught by other lanes.

## ⚠ CORRECTION 1 — the countdown. My date was wrong; the truth is worse.

UPDATE 6 said *"only ISO date in page is 2026-09-02, two days PAST"*. **False.** The `calendar` lane
read `page_components.rendered_html` directly and found no such string; I re-fetched the served page
to discriminate (they have no outbound internet) and they are right.

`[MEASURED 2026-09-04 ~16:0xZ]` served `/tools/fight-countdown/index.html`, cache-busted, control
`<body>`=1: **`var fights` ×1 · `year: 2025` ×6 · `year: 2026` ×0** (months 5,6,5,8,9,6). My
`2026-09-02` returns **no context match at all** on a fresh fetch.

**So: served == DB, NO publish/cache gap** (a clean negative — that branch is closed). And the defect
is not one stale entry among working ones: **all six fights are dated 2025**, about a year past, so
selecting ANY option immediately trips the tool's own *"This fight has started or concluded"*. **Every
option is dead on arrival.** The six are also *fabricated* — Usyk/Fury II, Joshua/Wilder II,
Canelo/Benavidez, Inoue/Nery, Davis/Garcia II, Taylor/Serrano III — none is the genuine Canelo/Mbilli
31 Oct 2026 fixture sitting in `evidence_base`.

**How I got it wrong, and it is the reusable part:** I ran a generic `\b20\d{2}-\d{2}-\d{2}\b` regex,
got one hit, and reported it as *the fight date* without establishing what the string was. The real
dates are `year: 2025, month: N` — a shape that regex cannot see. **I found a date-like string and
published it as the defect.** A structural read of the data beat a pattern match over the text.

**Root cause is NOT 427** (`calendar` lane, verified): the component is `component_level='tool'`, has
**no `input_schema`**, and templates only `{{.InstanceID}}` — every other value is a static literal
baked in at generation. **There is no `content_data` path for this tool to consume `evidence_base`
through at all.** So it is a generator inventing plausible content where it had none — the same shape
as the logo inventing a wordmark. Filed as its own bug, cross-referenced to 427.

⚠ **And it was undetected until now** — none of the five deferred verdicts names it, and the four
calendar ones are SUPERSEDED (they describe a pre-build "buried in a paragraph" state). **So the
record-mode backlog is NOT a complete inventory of this site's defects. "It is already filed" is not
a safe assumption.**

## ⚠ CORRECTION 2 — the articles duplication is NOT growing. It stopped 48 hours ago.

UPDATE 6 implied "still growing" and I told three parties so. **False**, caught by session 457.
Verified live: the six NULL-`component_id` rows are dated 08-31 16:29:47, 08-31 18:14:32,
09-01 01:31:57, 09-01 01:58:36, 09-01 02:34:51, **09-02 16:28:02** — nothing since.

I compared a verdict written 09-01 ("four times") against a measurement taken 09-04 (six) and
attributed the delta to *ongoing* growth. The interval contained the growth; the growth had already
ended inside it. **UPDATE 2 of this very file lists those six timestamps — I wrote them myself two
days ago and did not read them before asserting.** The check was free and already done.
Appending stopped because migration 316's duplicate guard began hard-failing the action; the
producer-side fix shipped 09-03 (`f895616d7`).

**This inverts the remedy: deletion is now a ONE-WAY REPAIR**, not a losing race. Two conditions:
1. **ROLL FIRST, DELETE SECOND.** `828b22c7c` is inert until the chassis rolls. The shipped fix
   refuses to CREATE a listing in a guessed slot but still UPDATEs one if a single row occupies it.
   Seven rows sit there today so it refuses; delete six and occupancy falls to 1 — the page's own
   prose block — and the next rerender overwrites that prose with a listing. *(The roll landed
   16:01Z; dispatch safe from ~16:07Z.)*
2. **THE LIST DOES NOT COME BACK ON ITS OWN.** Verified: `pages.sections` for `articles-index` is
   `["hero", "generic-text-block", "call-to-action"]` — **no listing-class section named**. After the
   deletion the page correctly serves **no** article list until one is added to the plan. Say this to
   the owner in advance or it reads as a fresh fault.

## The carousel request — the section he means cannot be carouselled, and a held migration would put one in the WRONG section

`offer analyser / visual designer` lane, verified by me. `/index` (`build_status='needs_rebuild'`):
pos 1 `content-listing` "Latest from the site" ← **what he means; schema has NO carousel field**;
pos 2 `info-card-grid` "A few places to start" ← **schema HAS one**; pos 3 `call-to-action`. None
carries a `carousel` key today.

⚠ **`740_info_card_grid_carousel_defaults_on_HOLD.sql` is approved, UNAPPLIED (0 rows in
`schema_migrations`), and carries OWNER RULING 2026-09-03 "switch the switches".** If it lands and
this page rebuilds, **a carousel appears in "A few places to start" while "Latest from the site"
stays a grid** — the owner would see a carousel materialise on the page he complained about, in the
wrong section, and could reasonably read his request as actioned.

⚠ **740's own blast-radius census has gone STALE BY ADDITION.** Its header states *"39 do not carry
the key at all"*. `[MEASURED 2026-09-04]` live: **55 instances across 55 pages, 1 with the key, 54
without.** The ruling was taken against 39 and would now flip **54** — a 38% under-count. **Re-take
that census before applying.**

## Counts, reconciled — quote the FILTER with the number

`[MEASURED 2026-09-04]` boxingonline: **63** rows `status='deferred'` · **58** of those carrying
`[verdict, not dispatched]` · **46** of those content-type. Three lanes quoted 46, 58 and 63 at each
other as if they disagreed; all three are right and measure different sets. Fleet: 3,181 → 3,184
within the hour — drift, not disagreement.

---

# UPDATE 8 — 2026-09-04 ~16:15Z. Two more corrections, both from lanes, both sharper than mine.

## ⚠ CORRECTION 3 — 740's blast radius is 51, not 39 and NOT my 55. And "it grew 38%" is the wrong story.

I told the visual-designer lane their migration's census (39) was stale against my live 55. They
re-took it rather than accepting my figure, and **neither of us had the right number.**
`[MEASURED 2026-09-04 16:11:36Z]`:

| set | instances | with key |
|---|---|---|
| served (`build_status='deployed' AND status='active'`) | 44 | 1 |
| + active but NOT deployed (mid-rebuild) | +8 | 0 |
| **= ACTIVE, any build_status** | **52** | **1 → 51 would flip** |
| + not active (archived) | +3 | |
| = all rows (my number) | 55 | |

**51 is the answer**, because the default lands at **plan/render time**: an active page that is
`needs_rebuild` *will* be rebuilt — that is what the status means — so those 8 count; archived pages
do not. Their 39 was the SERVED pairing; my 55 was unfiltered.

**And the decomposition changes the story I was about to tell the owner.** Of the movement from 39,
only **+4 is genuine growth** on the identical filter in ~28 hours; **+8 is the filter, not the
estate changing.** So "the number he ruled on has grown by 38%" would be **wrong**: he ruled on 39,
the honest figure is 51, and the gap is mostly that the original census measured the wrong set.

That is the `build_status='deployed' AND status='active'` trap the same lane filed as a landmine an
hour earlier — and then walked into on their own migration's headline figure. **"What is this page
made of" is `status='active'` alone; the pairing answers "what is being served."**

## ⚠ CORRECTION 4 — "a rewrite today would repeat the failure" is a PREDICTION, not a measurement.

UPDATE 6 and my report to the owner said promoting the parked rewrite would mostly reproduce the same
essay because nothing consumes `evidence_base`. The `copy_quality_two_stage` lane checked the
writer's inputs — their mechanism, not mine — and found a **simpler and better-evidenced** reason:

**When that article was written, the site had ONE fact, and it was "The business is called Boxing
Online."** `evidence_base` history: 2 facts 08-31 12:21Z → 1 at 15:54Z → **7 at 09-02 12:41Z** → 8
today. The verdict was filed 09-01. **The writer produced a general essay because there was nothing
else to produce**, which holds whether or not a consumption gap also exists.

**So the inputs genuinely changed on 09-02, and a rewrite dispatched today is NOT a repeat of the run
that failed.** My "it would produce another general essay" is a prediction. The reason to hold is
that **nobody has checked whether the facts reach the writer** — not that we know they do not.

⚠ **And the obvious check does not work.** That lane sampled five writer prompts, found "Hrgovic" in
one, and was a sentence from reporting that consumption works. It does not: the match sits inside the
page's **own existing rendered HTML** — a BBC feed headline in the Existing Content block that every
rewrite prompt carries. Testing for "Verified Facts" returned true on all five, matching **strict rule
14's own prose listing the input block names**. **Neither instrument can distinguish a prompt CARRYING
facts from a prompt TALKING ABOUT facts.**

**The decisive test is one dispatch and a read of the resulting prompt for a facts block** — not a
grep for a fact string. One run settles it. Bounded claim available now: in five sampled 09-03
prompts, **no `evidence_base` fact appears as supplied evidence**; whether a facts block is assembled
for this site at all is **not established** and needs the assembly path read. Same seam as that lane's
empty-writer-context finding (`Company:` empty on 41% of 6,931 writer calls).

## Standing corrections to quote, since four of my claims have now been improved by other lanes

| my claim | corrected to | by |
|---|---|---|
| countdown counts to 2026-09-02 | six FABRICATED fights, all `year: 2025`, every option dead on arrival; served == DB | `calendar` |
| the duplication is still growing | flat since 09-02 16:28; deletion is a ONE-WAY repair | `457` |
| RFC_056 phase 3 = "switch the loop back on" | phase 3 CREATED record mode (`624_…_record_only_…_HOLD`); no phase releases | `site_delivery_and_editor` |
| 740 would flip 54, census stale by 38% | 51; +8 of the movement is my filter, only +4 is growth | `offer analyser / visual designer` |
| a rewrite today repeats the failure | a prediction; inputs changed 09-02; one dispatch settles it | `copy_quality_two_stage` |

**The pattern in four of the five is the same: I measured a proxy and reported it as the thing.** A
date-shaped string as a fight date; an interval containing growth as growth continuing; a filename as
a phase; an unfiltered row count as a blast radius. Every correction came from a lane that read the
STRUCTURE where I had matched a PATTERN.
