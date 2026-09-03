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
> **2026-09-03 08:37Z — MORNING STATE (session slept 21:5x→08:2xZ).** §1.2 DONE (429 CLOSED, moved
> `640e4041c`: contact 404 with index 200 + invented 404; both sites `th2:`; converged at
> boxingonline's FIRST serviced post-roll tick 22:53:51Z per the 429 lane's series read).
> §1.3 the owner-approved index rebuild `c5614b00` **FAILED ×3** (21:51/22:xx/23:25Z) on the
> content-loss guard: `SECTION SHRINK REFUSED … call-to-action 1116→167 chars of VISIBLE text
> (15% kept, floor 50%)` — the build path regenerates a ~1.1k-char news-teaser CTA as a 167-char
> one; do NOT set `section_shrink_floor` (step-config escape hatch, fleet-wide). Also filed by
> that attempt: `needs_section_data` `97db7b0f` (featured-content's `query.featured_post` is an
> UNKNOWN query name → the section would be skipped on any rebuild; the two open
> `empty_section:…:featured-content` items are the same gap). §1.4 the GTM/consent wave was
> never coming: the detector's `stale_chrome` item was born `unresolved` by the two-strike
> ladder → **`bugs_open/451`** (fleet: 75/76 parked, 12 sites; diagnosis row `0639080d`),
> operator re-file **`ec92320f`** filed 08:35:44Z (watcher armed; queue ~3 h deep per the
> components lane). Chassis still `v1.0.1355` (no roll overnight) → §1.1 unchanged.
> `customer_access_tokens` 0. Token valid.
> **⚠ ROLL INBOUND (owner, 08:39Z): a new chassis is being prepared for deployment within the
> hour.** After it lands: (a) verify at the artefact, per service — image-generator-adapter
> provenance line + `git merge-base --is-ancestor fcbe6071c <stamp>` (the 424 guard fix) and
> the same for the chassis (⚠ `fcbe6071c` adds NO function name or string literal — only a
> local `finalAlpha` — so there is no binary-probe literal for it; ancestry at the stamp IS the
> check, with `KeyOutBackground` PRESENT as the control that the binary is the adapter);
> if ABOARD, §1.1's unblock (a) is met and only owner decision #2
> (first test site) remains before a calibration run; (b) a roll kills whatever is CLAIMED
> mid-run — re-read `ec92320f` (chrome refresh), `06210ec6` (components' discriminator) and
> diagnosis row `0639080d` afterwards; a `failed`/reaped state there is the roll, not the
> mechanism — re-file, do not diagnose; (c) no dispatch within ~300 s of the new pods; (d) the
> pre-delivery sweep (§1.5) waits for the roll AND the chrome wave AND the mirror tick.
> **ROLL LANDED 08:58Z: `v1.0.1356`**, chassis pods started 08:56:01/08:56:54/08:57:15Z (two
> ReplicaSets, same tag), image-generator-adapter 08:56:08/08:56:20Z; provenance line read
> from EVERY new pod = **`7bf1ff674`**. Ancestry: **`fcbe6071c` (424 guard fix) ABOARD**;
> controls `b2322a203` + `6440ec968` ABOARD, `433179904` (post-stamp) NOT — the check
> discriminates. So §1.1's unblock (a) is MET on both services. **Remaining before any
> boxingonline logo generation: (b) the owner's decision on the first deliberate test site.**
> This lane's recommendation: the 424 lane's post-roll reset of the three portfolio sites
> carrying broken logos (designblog / seotools / gamedesign — theirs to do) IS the
> lower-stakes calibration; read those results at the PNG bytes (424 RUNBOOK chunk scan +
> `border_keyed` in the adapter log), and only then regenerate boxingonline's. Not fired.
> **OWNER GO 09:25Z ("yes try the logo on boxingonline") — FIRED:** `needs_imagery`
> **`d71b7877-b42a-4019-9ede-74be363209ff`**, `item_key needs_imagery:site:-:logo`, filed
> 09:24:42Z `triaged`, priority 10, `image-build-handler`; spec = the fleet shape of the run
> that succeeded on websitepromotion (`95333588`), boxingonline's base prompt with the voided
> wordmark phrase removed, **no GROUND clause** (the key-colour clause is code, 424; text-free is
> code, 417). Baseline before firing: asset `20ce80fb` `updated_at` 09-02 10:40:12Z; served
> logo 139,777 B, sha256 `4aff0f99383c3bf5…`, 400×218 depth 16 colour type 2 no tRNS (the
> interim solid mark). **Success = BOTH** (424 RUNBOOK): stored/served PNG colour type 4/6 or
> `tRNS`, with a real fully-transparent fraction, AND the adapter log's `border_keyed` for
> this run; plus eyeball for a single text-free composition and the CONTRIB's magenta fringe.
> A refusal (guard) leaves the interim mark in place — not a failure of this test, a data
> point. Watcher armed; 424 lane + boxingonline session told. The three portfolio resets
> (424 lane, 09:23:49Z) run in the same window — read all four.
> Binary control probe on the new adapter pod: first attempt KILLED by the rollout (exit 137
> / "container not found") — inconclusive, not "absent"; retried on a settled pod (NOTES).
> **~21:18Z (yesterday):** two of §1's four premises are corrected below in place — §1.1 (logo
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
   > **CORRECTED 2026-09-02 ~21:14Z (this lane, next session, cold-start re-read of
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
   > **UPDATE 21:26Z (clock-read):** unblock (a) is now COMMITTED and council-APPROVED —
   > `fcbe6071c` (21:17:18Z, 424 lane: `BorderKeyed` computed from each pixel's FINAL
   > alpha==0, not BFS reachability; mutation-proven regression test), verdict `52bd50a1`
   > approved with no objections — but **NOT aboard `v1.0.1355`** (`git merge-base
   > --is-ancestor fcbe6071c 0d2feee2f` → no). So (a) = the NEXT roll of
   > image-generator-adapter (+ chassis; the 424 RUNBOOK says both). After that roll,
   > (b) the owner's decision #2 (boxingonline vs a lower-stakes site first) STILL
   > stands, and the first real generation is a CALIBRATION run (424 handoff item 3):
   > read `border_keyed` AND chunk-scan the PNG bytes, never one alone. The 424 lane
   > also reports designblog / seotools / gamedesign still serve the broken near-opaque
   > logos and has flagged that to the owner as urgent; their work items are NOT to be
   > reset before that roll.
2. ~~**Contact-404 — WATCH ONLY, never force the reconciler.**~~ **DONE 2026-09-03 (429 CLOSED `640e4041c`; served 404/200/404 at 08:23:09Z; both sites `th2:`).** `b60d66e3c` (429,
   council b576bcc6 APPROVED 18:52Z) converges each opted-in site ONCE on its
   normal rotation slot: boxingonline's publish result should read
   `published:true, deleted:1` (contact.html), then no-drift again. At ~21:0xZ
   contact.html still served 200 (index control 200) — pre-convergence, expected.
   `[ADDED ~21:14Z]` This site's reconciler slot is ~:52 past the hour `[INFERRED
   from three ticks — COMPLETED orchestrations with domain=boxingonline at
   18:51:51/18:52:10, 19:52:29, 20:52:47/20:53:03Z; index last-modified 20:53:26Z]`,
   so the first POST-roll tick (pods 20:56/57Z) is ~21:52Z — probe the pair after
   ~21:55Z, and treat a still-200 before then as nothing at all.
   `[429 lane, ~21:14Z]` Rotation ORDER is unconfirmed without the token: if
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
   `[21:21Z, DB, token back]` Both opted-in sites still `th1:` — boxingonline
   `published_at` 20:53:33Z (the pre-roll tick); **noted.co.uk last published
   2026-08-16** (17 days of no drift — the th1→th2 prefix change is what will make it
   republish). Which of the two the ~21:52Z tick services first is unknown.
   > **CORRECTED 2026-09-02 ~21:18Z (next session, from the components lane's own
   > post-roll handoff `../components_lane_425/HANDOFF_2026-09-02_continue_here.md`,
   > commit `753c3e6bf`, 21:10Z): there is NO HOLD apply to wait for, and decks will
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
   > ~21:18Z, same session, by running the check I had cited before it ran (classifier
   > outage delayed it): FOUR other roll commits touch those files** — `6525b45ae`
   > (444's listing-page item-source gate: `plan_sections_action.go` +10,
   > `queryresolve/business_directory.go`), `dbb218a41` (443 fallback-tier subjects:
   > `plan_sections_action.go` +102), `3b1389ca0` (137 runtime-fill exemption:
   > `rerender_page_sections_action.go`, 1 line), `987ed3b3b` (427: `queryresolve.go`,
   > `upcoming_events.go`). None names 425 or `list_item_text.go`; **diffs NOT read.**
   > So "`v1.0.1355` should not change the rerender-path result" is `[INFERRED from
   > commit messages]`, not measured — the components lane's §0 item 2 discriminator
   > re-run IS the measurement. Item 4's chrome wave is still the rerender path either way.
   > **MEASURED 21:26Z: the discriminator ran, and the roll did NOT change the answer.**
   > The components lane filed `page_rerender` `b238bed9` on index at 21:21:04Z
   > (summary "bugs_open/425 §0 item 2: post-roll discriminator on v1.0.1355",
   > reason `template_changed`), claimed by `build-dispatch-loop` 21:21:54Z, complete
   > 21:22:26Z; `pages.deployed_at` 21:22:23Z; all four `page_components` on the page
   > `updated_at` 21:22:14Z; **8 `page_component_history` rows keyed on `page_id`** at
   > 21:22:14Z (the write happened). At the row: content-listing `7dead3e5`
   > `articles[0]` has **NO `excerpt` key**; `rendered_html` carries **0**
   > `article-card__excerpt` vs **24** `article-card`. Baseline control: guides-index
   > (build path, `2e738efd`) `articles[0]` HAS `excerpt`. So on `v1.0.1355` the
   > rerender path still does not execute the producer, the four other roll commits did
   > not change it, and the served /index.html will STILL show 0 decks after the
   > ~21:52Z mirror tick — **not a new failure**. The route to decks on this site is the
   > components lane's ⭐ experiment (their handoff: a `needs_page` rebuild of index
   > "would not destroy the repro — it makes it strictly better", then a rerender to see
   > whether the key survives). They hold the sequencing; this lane did not fire it.
   > **21:28Z: the components lane has handed the rebuild to this lane** ("fire it
   > whenever suits; it cannot damage the repro") and confirmed the defect SURVIVES
   > v1.0.1355 with both chassis pods probed carrying the fix. **Prepared, NOT fired:**
   > exact INSERT + pre-flight + verification in this dir's RUNBOOK ("BUILD-path
   > rebuild for the card decks"). Held for the OWNER's go because the build path
   > REGENERATES the page's LLM-written copy (webdesign NOTES 17:3xZ: build
   > regenerates, rerender carries) — on the home page he reviewed point by point,
   > that is a copy change he did not ask for, on the page that passed. If he says go:
   > run the RUNBOOK block, verify at the row, then hand the row id to `components`
   > for their step-2 rerender.
   > **FIRED 21:35Z on the owner's go** ("Fire it now", his answer at ~21:31Z):
   > `needs_page` **`c5614b00-9415-4703-a383-2da5633ddced`**, filed 21:34:14Z `triaged`,
   > pre-flight clean (no in-flight build/rerender on the page; chassis pods 37 min old).
   > Baseline recorded at filing: content-listing instance `9e643633` (NB: re-inserted
   > at 21:30:07Z by something other than this lane — chased in NOTES), `has_excerpt`
   > false, title suffixed. Watcher armed on the item; verify per the RUNBOOK block,
   > then hand the id to `components` (messaged at filing). Expect claim ~21:36–21:41Z,
   > complete ~2–8 min later; a missing claim at 10 min = read the row, do not re-file.
   > **RESULT (read 08:2xZ, 09-03): FAILED, 3 attempts** (claimed 21:42:49Z; refusals 21:51Z,
   > ~22:3xZ, 23:25Z; `mark_item_failed` 23:25:39Z). Error each time: `save_page_sections:
   > SECTION SHRINK REFUSED for page "index" — call-to-action 1116→167 chars of VISIBLE text
   > … (15% kept, floor 50%)`. The stored CTA is a ~1.1k-char running news-teaser block; the
   > current writer produces a normal short CTA; the content-loss guard cannot know which is
   > right and refuses the whole save. **The build route on index is therefore CLOSED** until
   > either the 425 rerender-path defect is fixed (the real fix) or someone decides the CTA's
   > stored text is not the baseline — and that is a copy decision, not a floor to lower.
   > Side-finding from attempt 1: `featured-content` on index sources `query.featured_post`,
   > which the resolver does not know (`needs_section_data` `97db7b0f`, needs_human_review;
   > `on_missing: skip_section`) — a vocabulary gap of the guides-index class; owner of the
   > `featured_article` component to route. Stored/served index = the 21:30Z rerender.
   > The components lane's guides-index discriminator (`06210ec6`, template_changed, filed
   > 08:26Z) runs instead — guides-index already IS a new-shape baseline; ordering is by
   > `page_component_history.source_item_id`, not by who fires first.
   > **OWNER RULING 09:54Z ("Scoped override, one run") — the refusal was protecting the
   > copy he rejected.** The boxingonline session read the defended CTA: ~1.1k visible chars
   > narrating the news feed and walking through four tools in full sentences — his item 1
   > verbatim; the 167-char rebuild IS the repair, and a percentage floor cannot tell padding
   > from substance. No per-item override exists (`save_sections_shrink_guard.go:246`,
   > `single_slot_floors.go:112` read only the saving STEP's config); page-build-handler's and
   > page-rerender's `save_sections` are separate configs (measured at the live rows).
   > **Migration `725_boxingonline_index_rebuild_shrink_floor_window_HOLD.sql`** (applied by
   > hand 09:53:26Z; `snapshot_agent` taken; `section_shrink_floor = 0.1` on page-build-handler
   > `save_sections` ONLY; open `needs_page` fleet-wide at apply = 0) + its `_ROLLBACK`. Rebuild
   > refiled as **`2d1f9c51-d5ce-433e-8d67-f3bfd79916b4`** 09:53:32Z. A monitor runs the
   > ROLLBACK at that item's terminal state and prints the floor, the row read and the count of
   > other builds that saved inside the window. **If you find the floor still 0.1 with `2d1f9c51`
   > terminal, run the ROLLBACK by hand — the window must not outlive the item.** After it lands:
   > new-shape baseline for the components lane's step 2; read the regenerated copy against the
   > OWNER_REVIEW classes; featured-content drops (unknown `featured_post` source).
   > **10:27Z: window CLOSED at 10:26:44Z (ROLLBACK; 0 other builds completed while open) and
   > made CLAIM-GATED** — monitor `bpphsj4ji` re-applies the HOLD the moment `2d1f9c51` reads
   > `claimed` (page-build-handler takes ~8 min from claim to save) and runs the ROLLBACK at
   > terminal. Why: the fleet trigger's `find_dispatchable_site` (read at the live row) serves
   > the site whose 8-item window holds the OLDEST `created_at`; boxingonline's minimum is
   > 08:26Z today against 21 sites / 270 triaged items with week-old minima, so its turn is
   > hours away and the flat window would have exposed every build until then. No honest
   > override exists (site selection has no input; the 08-08 "dispatch the loop directly"
   > bypass no-ops under bare orchestrate; direct-handler has the 029 hang). **If you find the
   > floor at 0.1 with `2d1f9c51` terminal, run the ROLLBACK; if you find it 0.1 with the item
   > `claimed`, that is the monitor working — leave it.** The logo `d71b7877`, the chrome
   > refresh `ec92320f` and the components' `06210ec6` wait on the same turn and will load in
   > the same run (max 8), expected order 8 → 10 → 10 → 80.
   > **CORRECTED 10:31Z: "hours away" was WRONG — boxingonline is POSITION 2.** The
   > components lane produced a site (adversecreditmortgage.co.uk) holding the fleet's oldest
   > window minimum yet untouched for two days; the clause I had read and not applied is the
   > SITE LOCK (`sites.locked_at` 08-18, owner halt → invisible to the trigger). My "21 sites /
   > 270 items" was a proxy census (`status='triaged' AND pipeline='build'`) that ignores the
   > lock, attempts, retry, deps, governor and claimed-skip. Replicating the trigger's FULL
   > SQL minus LIMIT (`[MEASURED 10:31Z]`): 1 finetuning.uk 08:25:16 · **2 boxingonline.com
   > 08:26:24 (5 in window)** · 3 garden-tools.uk … gaswholesalers/gamesdesign not in the top
   > 15. So the four items load in the NEXT boxingonline run, minutes away. The claim-gated
   > window stands (it is now the better design regardless). WRONG_CALLS row filed.
   > **What actually produces decks here:** the BUILD path — proven on this very site
   > (guides-index `needs_page` rebuild 17:23:02Z → excerpt PRESENT, suffix stripped;
   > §1.4 closed that way). A `needs_page` rebuild of `index` is the known-working route
   > if the components lane's post-roll re-run of their §2 discriminator (their §0 item
   > 2, needs the token) still shows the rerender path broken. Not fired tonight: no DB,
   > and it is a joint call with the components lane (messaged ~21:18Z). Success criterion
   > unchanged: NON-ZERO `article-card__excerpt` with real copy + suffix-free titles on
   > /index.html. Also owed after ANY rerender here: `721`'s effect on the six hero
   > components (applied pre-roll, "needs a re-render to show") — approval-readout B.8.
4. ~~**GTM wave — watch, third-party-owned.**~~ **CORRECTED 2026-09-03: it was never inbound.**
   The detector's `needs_rerender`/`stale_chrome` for this site (`5b4eb7a0`, 09-02 06:19Z) was
   written `unresolved` AT BIRTH by the loader's two-strike ladder (two `complete`/`failed`
   siblings in 7 d — one of them a SUCCESS, 09-01) → `bugs_open/451`, landmine + 016b §9
   entries, CONTRIB to `analytics_gtm`, diagnosis row `0639080d`. Operator re-file
   **`ec92320f-3037-448a-bd55-de8385404d92`** (08:35:44Z, triaged; spec = the last completed
   row's). Success = `site_components` head/header/footer `updated_at` > 08:35Z with
   `GTM-PQ3WCTBD` AND `cc_v1` present, then the served pages after the mirror tick (monitor
   armed: GTM count on index). Assemble-only prediction under test on guides-index: no
   history row attributable to the wave / archived content still carries `excerpt`.
   Original text follows. Analytics lane wrote
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

> ⚠ **21:08:03Z (JWT `exp`): the shared kubeconfig token EXPIRED mid-session** (every kubectl →
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

> **Timestamp correction, 21:20Z:** the in-place labels above originally read "~21:2xZ / ~21:4xZ / ~21:5xZ"; they were INFERRED, not read from the clock, and overstated elapsed time by ~30 min (the whole session's work sits between 21:03Z and 21:19Z). Re-anchored to the commit clock (21:14Z, 21:18Z) and `date`. The reconciler tick at ~21:52Z had NOT happened when any of it was written.
