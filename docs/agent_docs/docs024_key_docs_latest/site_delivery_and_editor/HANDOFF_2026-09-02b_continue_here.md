# HANDOFF 2026-09-02b (late evening) — the roll LANDED (v1.0.1355, all four fixes aboard); one dispatch is ours to fire, three convergence waves to watch, then the pre-delivery sweep and the 651 rehearsal

**Supersedes** `HANDOFF_2026-09-02_continue_here.md` (every item resolved or carried
below — its §1.2/1.5/1.6/2 in-place updates are folded in). **Joint picture**
(site_delivery_and_editor + webdesign lanes, one session driving both since 08-18).
Running technical log: `../webdesign_uk_build_service/NOTES_webdesign_uk_build_service.md`
(the 09-02 19:4x→21:0x entries are this handoff's evidence). Owner critique record:
`OWNER_REVIEW_2026-08-31_boxingonline_what_he_found_and_what_each_finding_actually_is.md`
(this dir). Approval decision surface:
`APPROVAL_READOUT_2026-09-02_boxingonline_what_is_actually_fixed.md` (this dir,
boxingonline session's — three columns: verified / not fixed / built-but-inert).

## 0. State in one paragraph

boxingonline.com (site `d2aa5206-73bc-4707-a69c-2702c1eb9152`, order BR-9AUZ59, first
PAID build) serves at https://boxingonline.ugg2.com, green on everything previously
verified (footer machine-render proven at row AND serve — the digest discriminator;
guides index 4/4; email 0/contact-links 0 on served pages). **The chassis roll this
lane was waiting on LANDED**: `v1.0.1355`, pods 20:56:43/20:57:10Z, provenance stamp
`0d2feee2f`, and ancestry-verified ABOARD: `b2322a203` (424 logo), `b60d66e3c` (429
contact sweep + th1→th2), `9f6f91325`+`c1178442d` (425 cards, Go half). **Delivery
stays HELD on the owner's fix-everything cut-line**; `customer_access_tokens` = 0
(checked 19:3xZ). As of ~21:0xZ NONE of the three convergence waves had fired yet
(contact.html 200, GTM count 0, excerpt decks 0 — all with controls held; that is the
EXPECTED pre-convergence shape, not a failure).
> **21:4xZ:** two of §1's four premises are corrected below in place — §1.1 (logo
> regen: DO NOT FIRE, owner decision) and §1.3 (cards: no HOLD pending; the rerender
> path is the open defect and the roll does not touch it). §1.2 and §1.4 stand.

## 1. NEXT, in order

1. **FIRE the transparent logo regen — unblocked, ours to do.** `b2322a203` is
   aboard by stamp ancestry (the sanctioned check; CLAUDE.md "did my fix ship?").
   Recipe + any belt-and-braces literal probe: the 424 lane's
   `../bugfix_424_logo_transparency/HANDOFF_2026-09-02_continue_here.md`. The
   interim solid-#0a0a0a mark is correct and serving meanwhile. Disciplines:
   ~300s no-dispatch after a chassis restart (pods 20:56/57Z — already passed by
   any new session's clock); expect 25–36 min queue latency; a missing
   orchestration row is latency, NOT a dropped dispatch — find the run by
   payload, never re-fire on that evidence. Owner's ruling being satisfied here:
   no baked background, text-free single composition.
   > **CORRECTED 2026-09-02 ~21:2xZ (this lane, next session, cold-start re-read of
   > the dir this item points at): item 1 is NOT unblocked — DO NOT FIRE.** The roll
   > half is true and is now verified on BOTH services: image-generator-adapter's
   > provenance line read from pod `588ffc76b9-fddqd` (20:56:58Z) = stamp `0d2feee2f`,
   > same as the chassis; ancestry of `6440ec968`/`b2322a203`/`b60d66e3c`/`9f6f91325`/
   > `c1178442d` → all ABOARD. What this item did not read: the 417 lane's CONTRIB in
   > the 424 lane dir (rounds 1–3, last 19:45Z, committed `7fc657116` at 19:45Z — ~75
   > min BEFORE this item was written — and folded into `bugs_open/424` §"the matte
   > RAN"). The matte's fail-closed guard gates on `stats.BorderKeyed`
   > (`dynamic_adapter.go:683`), which counts border-flood MEMBERSHIP at
   > `dist <= outer` (`keyground.go:104/131/149`); a pixel only reaches alpha 0 at
   > `dist <= inner` (`:176`). Five fleet runs on `v1.0.1354`, queue-triggered:
   > **1 usable of 4 stored + 1 correct refusal**, and `border_keyed=1.000` on BOTH a
   > 0.0%-transparent failure (designblog) and an 87.4% success (websitepromotion).
   > `git log b2322a203..HEAD -- internal/adapters/imagegenerator/{keyground,dynamic_adapter}.go`
   > is EMPTY (checked 21:1xZ): `b2322a203` fixed the PROMPT contradiction; the guard
   > on `v1.0.1355` is byte-identical to the one that passed those failures. There is
   > NO asset revert seam (424 §"why there is no clean revert"), and the 424 lane's own
   > handoff lists "is boxingonline the right first real test" as OWNER decision #2,
   > "not this session's call". Firing = ~3-in-4 chance of replacing the correct
   > interim mark on the first paid customer's live site with an unusable one that
   > reports success. **Unblocks when EITHER** (a) the guard gains a border-
   > TRANSPARENCY statistic (final alpha 0) and gates on it — 424 lane's code, council
   > scope, another roll — **or** (b) the owner rules to accept the risk / test on a
   > lower-stakes site first. Until then the interim solid mark stays. The fleet cannot
   > pre-empt us: no open `needs_imagery:site:-:logo` for `d2aa5206` at 21:0xZ (3 open
   > fleet-wide, two other sites); last logo run here = interim `00aa1796` 10:40Z;
   > asset `20ce80fb` updated 10:40:12Z. CONTRIB r3 also notes the one good run carries
   > a magenta fringe at the mark's edge (despill incomplete) — even a pass wants an
   > eyeball before it serves. Messaged: boxingonline.com, bugs_open/424, bugs_open/429.
2. **Contact-404 — WATCH ONLY, never force the reconciler.** `b60d66e3c` (429,
   council b576bcc6 APPROVED 18:52Z) converges each opted-in site ONCE on its
   normal rotation slot: boxingonline's publish result should read
   `published:true, deleted:1` (contact.html), then no-drift again. At ~21:0xZ
   contact.html still served 200 (index control 200) — pre-convergence, expected.
   `[ADDED 21:2xZ]` This site's reconciler slot is ~:52 past the hour `[INFERRED
   from three ticks — COMPLETED orchestrations with domain=boxingonline at
   18:51:51/18:52:10, 19:52:29, 20:52:47/20:53:03Z; index last-modified 20:53:26Z]`,
   so the first POST-roll tick (pods 20:56/57Z) is ~21:52Z — probe the pair after
   ~21:55Z, and treat a still-200 before then as nothing at all.
   `[429 lane, 21:3xZ]` Rotation ORDER is unconfirmed without the token: if
   `noted.co.uk` (the other opted-in site) is ahead in the queue, the ~21:52Z tick
   services IT (expect a full th2 republish there) and boxingonline waits for ~22:52Z.
   **A still-200 after ONE tick is not a signal; two boxingonline-serviced ticks with a
   200 would be.** Both lanes have a monitor on the pair; whoever sees the 404 first
   pings the other; the 429 lane does the `published_hash` th2: read + `bugs_closed`
   move once the token returns.
   When it 404s (and a kept page still 200s — probe the PAIR), strike this item;
   the 429 lane also verifies server-side. ⚠ Flip-side landmine now live:
   anything hand-placed under a `*.ugg2.com` prefix is swept on next drift.
   The LINK half closed long ago (20/20, two sessions).
3. **Cards — wait for the components lane's HOLD apply, then serve-check.**
   Their Go fix is aboard; the rerender is driven by their
   `…83_content_listing_rerender_after_roll_HOLD.sql` (a HOLD is applied BY HAND,
   by THEM — do not apply another lane's migration; coordinate via session
   `components` / bug file `bugs_open/425_HANDOFF_2026-09-02_…`). Then serve-check
   /index.html: **success = NON-ZERO `article-card__excerpt` decks carrying real
   copy + suffix-free titles** (the 682 fingerprint INVERSION — on data-carrying
   items, rendered decks are the success state; "must be 0" was the empty-slot
   era). At ~21:0xZ: 0 decks / 24 `article-card` control = pre-convergence.
   **Do NOT report cards fixed until this check passes.** The 420-addenda
   discriminator applies if confused: served object last-modified older than
   pages.deployed_at = mirror lag (wait); newer = dirty source (look upstream).
   > **CORRECTED 2026-09-02 ~21:4xZ (next session, from the components lane's own
   > post-roll handoff `../components_lane_425/HANDOFF_2026-09-02_continue_here.md`,
   > commit `753c3e6bf`, 21:2xZ): there is NO HOLD apply to wait for, and decks will
   > NOT arrive by themselves.** (a) `683_…_HOLD` was applied BEFORE the roll — batch
   > `…000683`, 10 complete / 4 cancelled (section-component floor, by design); on this
   > site those are the `page_rerender` rows `22421f7b` (17:25:06Z) and `68b4fb82`
   > (17:31:14Z), both `complete`, both the APPROVAL_READOUT §D "ran twice, reported
   > success, stored data unchanged". (b) The open defect is upstream of any roll: on ONE
   > binary (`v1.0.1354`, Go fix `f57f5ad1f` aboard since 12:28Z) the BUILD path writes
   > `excerpt` + strips the suffix and the RERENDER path does neither — reproduced three
   > times, seven branches eliminated by reading, three diagnosis runs failed
   > (their §2). `9f6f91325`/`c1178442d` are council-r3 refactors (reuse
   > `datahelpers.SafeCut`/`TruncateString`; 683 header wording) — not a fix for §2.
   > ~~`git log …` shows nothing else in the roll touching that path~~ **CORRECTED
   > 21:5xZ, same session, by running the check I had cited before it ran (classifier
   > outage delayed it): FOUR other roll commits touch those files** — `6525b45ae`
   > (444's listing-page item-source gate: `plan_sections_action.go` +10,
   > `queryresolve/business_directory.go`), `dbb218a41` (443 fallback-tier subjects:
   > `plan_sections_action.go` +102), `3b1389ca0` (137 runtime-fill exemption:
   > `rerender_page_sections_action.go`, 1 line), `987ed3b3b` (427: `queryresolve.go`,
   > `upcoming_events.go`). None names 425 or `list_item_text.go`; **diffs NOT read.**
   > So "`v1.0.1355` should not change the rerender-path result" is `[INFERRED from
   > commit messages]`, not measured — the components lane's §0 item 2 discriminator
   > re-run IS the measurement. Item 4's chrome wave is still the rerender path either way.
   > **What actually produces decks here:** the BUILD path — proven on this very site
   > (guides-index `needs_page` rebuild 17:23:02Z → excerpt PRESENT, suffix stripped;
   > §1.4 closed that way). A `needs_page` rebuild of `index` is the known-working route
   > if the components lane's post-roll re-run of their §2 discriminator (their §0 item
   > 2, needs the token) still shows the rerender path broken. Not fired tonight: no DB,
   > and it is a joint call with the components lane (messaged 21:4xZ). Success criterion
   > unchanged: NON-ZERO `article-card__excerpt` with real copy + suffix-free titles on
   > /index.html. Also owed after ANY rerender here: `721`'s effect on the six hero
   > components (applied pre-roll, "needs a re-render to show") — approval-readout B.8.
4. **GTM wave — watch, third-party-owned.** Analytics lane wrote
   `analytics.gtm_container_id=GTM-PQ3WCTBD` into `site_specs.site_config`; the
   `chrome` key SURVIVED the merge (header_slots + fight-calendar CTA verified
   intact — the nav fix did not revert). A 22-page chrome rerender propagates it
   via the stale_chrome pass; count still 0 at ~21:0xZ. **It is the RERENDER
   path: a still-suffixed card after THIS wave is NOT a new failure** — only
   item 3's HOLD rerender carries the card fix. If it does not fire in
   reasonable time, check whether stale_chrome has a tick scheduled rather than
   assuming latency. Chrome rerender is SAFE now (423 live; footer gate proven
   on a genuine render — digest set, len 2289, 16:27:55Z, all three contact
   markers FALSE).
5. **Pre-delivery sweep — ONLY after items 1–4 have all settled.** Run
   `docs024_key_docs_latest/SITE_DEFECT_CATEGORIES.md` (fleet acceptance
   checklist, ~30 runnable checks — **read its §0 first**: probe the serving
   host not the parked customer domain; must-be-present control on every probe;
   enumerate pages from the DB; date the artefact before accepting latency)
   against boxingonline, and read `APPROVAL_READOUT_2026-09-02_…` (this dir) —
   the owner's approval rests on its three-column honesty; do not collapse
   built-but-inert into fixed. Measuring mid-transition buys a state nobody
   will serve.
6. **Then the 651 rehearsal** — delivery-review-filer → owner EDITS + **APPROVE**
   on admin.apis.uk (never resolve; the request_changes button is live and
   verified end-to-end, API + frontend at the artefact) →
   zip-deliverable-dispatch → delivery-email-sender with
   `customer_email = build_queue.direction->>'customer_email'` (**NEVER
   sites.email** — 420 split; corrected recipe in 651's header + this dir's
   RUNBOOK). Handover stamp is ONCE-ONLY. Zip keys changed once at th2 — cut
   the zip fresh post-roll and record no zip link anywhere before the cut.
7. **Owner decisions outstanding** (none block 1–5): the 1b form-endpoint
   PRE-PLAN awaits him
   (`../static_site_form_endpoint/PLAN_2026-09-02_pre_plan_extensible_form_endpoint.md`;
   this lane's D1 review is DONE —
   `REVIEW_2026-09-02_form_endpoint_preplan_D1_vs_publish_seam.md`, D1(c)
   opposed on four grounds, (b) supported) · RFC_058 identity model (420 lane
   owns; this lane a named consumer; the pre-plan's D2 defers to it). The
   guide-reachability question is DEAD — withdrawn by the boxingonline session
   before the owner spent a decision; do not resurrect it.

## 2. What tonight established (all verified with controls; NOTES has the evidence)

- **Roll verification method worked as designed**: provenance line read from the
  fresh pod (`main.go:58`), stamp `0d2feee2f`, four ancestry checks — per
  SERVICE, at the artefact, no tag-trusting.
- **Footer closure upgraded to conclusive**: serve-verify (16:5xZ, three page
  types, contactless, control 7-19) + the boxingonline session's row
  discriminator — `rendered_html_digest` SET beside the bytes, and only the
  render path writes it. "The site ships a chrome artefact the pipeline cannot
  regenerate" is RETIRED. Whether 423's slicer is fixed vs un-triggered on this
  input stays 423's question — neither session claims it.
- **request_changes verified END-TO-END**: council `9f1cb042` revise→approved
  (09-02 14:23Z), both commits credit via `Council-Submitted` at 098; dashboard
  FRONTEND verified at the pod bundle (`index-D46-1-nI.js` carries
  `request_changes` + `Review Queue`, nonsense control absent). The owner's
  button is genuinely clickable.
- **423 half-1 (observability): a NO-DEMAND ZERO** — 0 `chrome_render_failed`
  rows while both post-roll rerenders succeeded; branch live-in-binary,
  unexercised. Neither pass nor fail; nothing to check until a failure occurs.
- **D1 publish-seam review delivered** (see §1.7) — committed `df56c1cf0`, the
  boxingonline session updating the pre-plan's §5 pointer.

## 3. Standing instruments (unchanged; use in this order)

Read the artefact first · a positive control must exercise the same row
population as the claim · served last-modified vs pages.deployed_at
discriminator · probe the capability, never the symbol alone · removed-string
controls prove which revision runs · the council gate is cheap and right —
submit before or with the commit.

## 4. Falsifiers (check before believing this file)

> ⚠ **21:1xZ: the shared kubeconfig token EXPIRED mid-session** (every kubectl →
> `Unauthorized`; memory `kubeconfig-token-expires-every-3-days`; the sanctioned
> expiry check confirms). Owner-only refresh. Served-site probes (curl on
> `boxingonline.ugg2.com`) still work and are enough to watch all three waves at
> the artefact; DB verification (`published_hash` `th2:`, the HOLD's rerender rows,
> `customer_access_tokens`) resumes when the token does. Last DB reads: 21:0xZ.

A newer handoff in either lane dir · the actual served state of
contact.html / GTM count / excerpt decks (my ~21:0xZ zeros go stale the moment
any wave fires — re-probe, don't quote) · whether the components lane applied
the HOLD (ask them; their bug file/handoff) · whether the 424 regen was already
fired by someone (check `orchestration_states` by payload domain before
dispatching — double-fire costs a round) · `SELECT count(*) FROM
customer_access_tokens` still 0 · a NEWER chassis roll (re-read the stamp; per
service) · owner decisions may have landed in the boxingonline session's thread
(session name `boxingonline.com`; it holds the critique relationship).

## 5. Read order, cold

This file → webdesign NOTES 09-02 19:4x→21:0x entries → APPROVAL_READOUT (this
dir) → SITE_DEFECT_CATEGORIES §0 → bugs 424/425/429 handoffs (all
2026-09-02) → RUNBOOK (this dir; corrected delivery recipes) →
`../dispatcher_thread/DISPATCHER_README_start_here.md`.
