# HANDOFF 2026-09-02 (evening) — the first paid site is through the full critique cycle; delivery HELD on the owner's fix-everything ruling; two roll-bound fixes gate the last quality items

**Supersedes** `HANDOFF_2026-08-31_continue_here.md` (every item closed or carried
below). **Joint picture** (site_delivery_and_editor + webdesign lanes, one session
driving both since 08-18). The running technical log is
`../webdesign_uk_build_service/NOTES_webdesign_uk_build_service.md` — the 09-01→09-02
entries are the story of this handoff and EVERY claim below carries its evidence
there. The owner's critique record is `OWNER_REVIEW_2026-08-31_boxingonline_what_he_found_and_what_each_finding_actually_is.md`
(this dir, boxingonline session's).

## 0. State in one paragraph

boxingonline.com (site `d2aa5206-73bc-4707-a69c-2702c1eb9152`, order BR-9AUZ59, the
first PAID build) serves at https://boxingonline.ugg2.com with **every brief promise
met and outside-verified**: six articles listed on News AND the home slot, fight
calendar in the nav (first production use of 407's header_slots declaration), owner's
email purged 19/19 twice-confirmed, contact page deleted (links half; see §2.5),
text-free single-composition logo on a solid header-matched ground. The owner
answered all six decision questions and raised eight further defects (2026-09-02,
his rulings verbatim in OWNER_REVIEW; my execution ledger in the NOTES); **his
cut-line ruling stands: EVERYTHING fixed before the delivery email — the 651 chain
stays HELD.** Tonight's chassis roll (v1.0.1354, pods 15:39/15:53Z) carried the 423
chrome-UTF8 fix — **the footer regenerated genuinely for the first time (all
acceptance criteria green at the row; serve-verify was in flight at write time)** —
and my request_changes endpoint (council **APPROVED r2**, consumption advisory
answered at the artefact). Two remaining quality items are **done-and-inert**,
blocked on the NEXT roll: the transparent logo (424 fix `b2322a203` postdates
v1.0.1354 — ⚠ do NOT regenerate until it rolls) and the card producer fix (under
diagnosis `fe4b8537`, components lane).

## 1. NEXT, in order

1. ~~**Footer serve-verify**~~ ✅ **DONE 16:5xZ, before this handoff was picked
   up**: wave drained, mirror published, served probes across three page types
   all read `footer_contact=0, email=0, control 7-19` —
   **"SERVED FOOTER VERIFIED: genuine regeneration, contactless,
   fleet-serving."** The last 423 consequence on this site is CLOSED: the
   footer is a real machine render at both the row and the served pages.
2. **Cards, producer half** — ~~WAIT for the components lane … `fe4b8537`'s
   answer~~ **ANSWERED 2026-09-02 ~17:3xZ, and it is a PATH split**: fe4b8537
   itself came back NOT CONFIRMED (iteration cap), but the A/B settled it on one
   binary (v1.0.1354), same site, canonical base, 3 min apart — **build path
   (7f1f4993, 17:23Z) emits the NEW item shape (excerpt present, suffix
   stripped); rerender path (item 684, 17:26:52Z) emits the OLD shape**, all
   controls clean (write happened, no floor refusal, post-682 template rendered).
   Stale-pod and timing theories are DEAD. The components lane owns the hunt;
   candidate mechanisms handed over (rerender's resolved-data STRIP layer
   ~:1541-:1617, or a silent resolve failure writing no key — `articles` was
   ABSENT from ResolvedData at merge time, since ResolvedData merges last and
   wins). Serve-check decks on /index.html still owed once their fix lands. The
   420-addenda discriminator (served object vs deployed_at) still applies.
3. **Logo, transparent regen — BLOCKED, do not fire** until a roll carries
   `b2322a203` (424 lane's self-contradicting-prompt fix; their handoff:
   `../bugfix_424_logo_transparency/HANDOFF_2026-09-02_continue_here.md`). The
   interim solid-#0a0a0a mark is correct and serving. Verify the fix at the
   BINARY (present + removed-string controls) before any dispatch. Owner status
   language: his no-baked-background ruling is implemented-and-inert, not pending.
4. ~~**Guides-index content** — blocked on a resolver vocabulary entry
   (`query.guide_pages` + pointing this instance at it; Go, roll-bound, NOT yet
   written by anyone — a small fix a session could own). CONTROLLED-PROVEN
   (NOTES ~13:5x): no pre-roll path holds guide items in a query-resolved
   listing~~ — **CORRECTED + EXECUTED 2026-09-02 ~17:2xZ: "roll-bound" was an
   overreach** (the experiment proved the build re-resolves, never that no
   existing vocabulary could resolve guides — `query.pages_where_type:guide`
   is deployed v1 vocabulary). Fix shipped pre-roll: fork
   `content-listing-guides-boxingonline-com` (`b475fe54`, only the articles
   source changed) + instance repoint (Path 0 stored-id binding — SURVIVES the
   build's delete/re-insert, proven) + rebuild `7f1f4993`. **Row-verified
   GREEN**: 4 guides, excerpt keys, suffix-free titles, guide-correct heading.
   **SERVE-VERIFIED 18:52Z (item CLOSED)**: exactly the 4 guide links, 0 blog
   item links, heading correct, 4 excerpt decks with real copy, no title
   suffixes — landed on the natural 18:48Z reconciler tick, no forcing.
   ⚠ Fingerprint reading for §1.2's acceptance: "article-card__excerpt MUST
   be 0" is the EMPTY-SLOT-era test (682 made the slots conditional) — on a
   page whose items CARRY excerpt data, non-zero rendered decks is the SUCCESS
   state, not a 682 regression. `query.guide_pages` (Go) is now OPTIONAL
   vocabulary sugar, no longer blocking. Full trail: webdesign NOTES
   17:1x-19:0x entries. The standing warnings hold: never retype pages to fix
   listings; never hand-write items into a resolved array (reverts on any
   build).
5. **Contact 404 half** — pinned to `bugs_open/429`, now OWNED (owner-routed,
   session `bugs_open/429`) and **FIX COMMITTED 2026-09-02 evening
   (`b60d66e3c`, Council-Submitted b576bcc6, verdict pending) — INERT UNTIL THE
   NEXT CHASSIS ROLL.** Ships: destination sweep with bulk floor (>20 AND >50%
   refuse; `allow_bulk_unpublish` opt-in the only override) + deleted-404/
   kept-200 acceptance pair + th1→th2. **Post-roll expectation, so the watch is
   not misread: ONE full republish per opted-in site on its normal rotation
   slot** — boxingonline's publish result should show published:true,
   deleted:1 (contact.html) — then no-drift again. ⚠ New flip-side landmine in
   their commit: anything hand-placed under a *.ugg2.com prefix is swept on
   next drift. The LINK half is closed 20/20 on two sessions' tables; the
   orphan serves 200 to direct navigation only, unlinked, unsitemapped, chrome
   frozen pre-rebuild (the discriminating observation in 429). Their pings:
   roll landed, then contact.html 404 verified → strike this item.
6. **Owner decisions still open**: guide reachability bridge (he chose
   guides-index; it provably can't hold guides pre-roll — the boxingonline
   session put the changed question back to him) · the 1b form-endpoint PRE-PLAN
   (boxingonline session drafting; reviews against the publish seam come to this
   lane) · RFC_058 identity model (420 lane owns; this lane is a named consumer).
7. **When the list is done**: the 651 rehearsal — delivery-review-filer →
   owner EDITS + **APPROVE** on admin.apis.uk (never resolve) →
   zip-deliverable-dispatch → delivery-email-sender with
   `customer_email = build_queue.direction->>'customer_email'` (NEVER
   sites.email — 420 split; recipe corrected in 651's header + this dir's
   RUNBOOK). Handover stamp is ONCE-ONLY. ~300s no-dispatch after chassis
   restarts.

## 2. What the roll(s) changed — all verified at binaries with controls

- **420 contract split LIVE** (yesterday's roll): re-seeding boxingonline is
  SAFE again (RUNBOOK block lifted with evidence); sites.email is the published
  contact only, empty = publish nothing.
- **423 fix LIVE** (tonight, v1.0.1354): `UpperFirst` present + control clean;
  the footer regenerated first try, every criterion green. (Half 1 —
  observability — check whether the store-failure branch now populates
  chrome_render_failed; the fix lane's file says which halves shipped.)
- **request_changes endpoint LIVE + APPROVED**: `POST /admin/work-items/
  :item_id/request_changes` files owner_critique at needs_human_review +
  approval_mode=manual (r1 shape was BOTH cluster-loadable AND
  schema-impossible — CHECK `swi_no_handlerless_promotable`; the whole story is
  in the NOTES ~13:4x-14:2x and the r2 rationale). Canary evidence: 3h+
  unclaimed with same-site demand control; found by the dispatcher poll
  (consumption advisory answered); canary retired complete. ⚠ Dashboard
  FRONTEND deploy state unverified — the API works; check
  frontends/admin-dashboard actually shipped before telling the owner the
  button is clickable.
- **682 template half LIVE + verified at served pages** (empty card slots gone);
  **425 producer half in-binary-but-not-executing** — five-times-reversed
  diagnosis converged: the ten completed pages DID re-render+archive
  (page_component_history keyed on page_id corroborates the served fingerprint
  exactly), but with pre-fix item shapes; fe4b8537 owns why.

## 3. The instruments this week forged (use these, in this order)

- **Read the artefact first.** Every instrument that survived reads the served
  page or the DB's own record of it (template fingerprint, excerpt-key
  presence, the floor guard's arithmetic, per-page probes with per-page
  controls); every one that fell read a column or a table about a column.
- **A positive control must exercise the same row population as the claim**
  (page_component_history: component_id is 98.4% NULL — join page_id; check
  your filter column is populated before trusting its zero).
- **Served object last-modified vs pages.deployed_at**: older = mirror lag
  (wait); newer = dirty source (look upstream). In 429/420 addenda.
- **Probe the capability, never the symbol alone**: symbols present verified
  the ROLL, not the executing path — three separate incidents this week.
- **Removed-string controls** prove which revision runs (absence of a deleted
  literal = the new code).
- The council gate is CHEAP and RIGHT: two REVISEs this week each found real
  defects in my shipped code. Submit before or with the commit.

## 4. Bugs filed/updated this cycle, with owners

`419` planner zero-section blog page (unowned; 090 UNVERIFIABLE, symptom+census
only) · `420` billing-email/published-contact + addenda (class fix ROLLED;
residual = RFC_058, 420-lane) · `423` chrome UTF-8 reasonless-false (half 2
FIXED+ROLLED; half 1 check) · `424` transparency capability gap (424-lane;
their fix awaiting NEXT roll — see §1.3) · `425` cards (components lane, live) ·
`427` calendar has no data (news_editorial owns acquisition; ONE seam with the
comparator per the owner's research instruction) · `429` mirror cannot
unpublish (~~unowned~~ **OWNED as of 2026-09-02 evening** — owner routed it to a
dedicated thread, session name `bugs_open/429`; cites 304. This lane is the
CONSUMER: the contact ruling's 404 half closes on their served-404 acceptance —
contribute into the bug file, do not compete). Landmines added: reconciler force sends you to
the back ("checked"≠"published"); plus the NOTES-recorded patterns (undoing an
override restores an unwatched default ×2; a true statement about one layer
restated as a system rule ×3).

## 5. Falsifiers (check before believing this file)

Newer handoff in either lane dir · the footer serve-verifier's actual output ·
whether a roll landed carrying `b2322a203` (unblocks §1.3) and any resolver
vocabulary fix (unblocks §1.4) · fe4b8537's verdict · the 098 report for
9f1cb042 crediting · `SELECT count(*) FROM customer_access_tokens` still 0 (no
delivery happened) · the owner's decisions may have landed in the boxingonline
session's thread — ask that session first (it holds the critique relationship).

## 6. Read order, cold

This file → NOTES 09-01→09-02 entries (the correction chains ARE the content)
→ OWNER_REVIEW (his words) → bugs 423/424/429 → RUNBOOK (this dir; the
corrected delivery recipes + lifted re-seed block) →
`../dispatcher_thread/DISPATCHER_README_start_here.md` (the owner's standing
critique surface — endpoint live, worked example = this whole cycle).
