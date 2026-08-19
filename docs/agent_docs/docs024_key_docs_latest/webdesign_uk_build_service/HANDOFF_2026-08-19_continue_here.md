# HANDOFF 2026-08-19 — MERGED cold-start: webdesign.uk build service + site delivery (Phase 4 next)

**SUPERSEDES** `HANDOFF_2026-08-18b_continue_here.md` (this lane) **and**
`../site_delivery_and_editor/HANDOFF_2026-08-18_continue_here.md` (that lane's joint
file, merged in here at the owner's direction, 2026-08-19). Both are bannered.

**Why merged:** one session drives both, and the work genuinely joins — the delivery
machinery (ZIP, live link, domain rent/buy, delivery email) *is* what webdesign.uk's
copy promises customers. **Each lane keeps its own NOTES / PLAN / RUNBOOK / README /
SUMMARY and its own register.** Write findings into the lane where the work happened.

**Read order, cold:** this file → both NOTES tails (2026-08-18 and 08-19 entries) →
`../site_delivery_and_editor/PLAN_2026-08-17_delivery_architecture_decisions.md`
(owner decisions + build order) → `README_where_we_are.md` (owner-facing) →
register `DGH-011` for the ZIP mechanism → `bugs_open/299`.

---

## 0. State in one paragraph

The commercial position is settled, applied at the register, and **verified at both
the served pages and the live bot**: £149 paid in full before the build, one-shot
with no approval stage, no changes included, no refunds, delivered as a ZIP to keep
plus the site live at a link we host for about a month, domain rentable at £10/mo or
buyable for £200 one-off, any sort of site (not businesses only), no pre-sales
service, **and delivery is now "two or three days"** (owner, 2026-08-19). The retired
£1,200 offer has been swept out of all nine site_specs. The chat box carries the
prompt-maker and is live on the VM. Phase 3 (the ZIP deliverable) is complete and
live-proven. **Phase 4 — handover + the delivery email — is the next build and has
not started.**

---

## 1. FIRST: four page rebuilds are in flight (queued 2026-08-19 ~10:00Z)

The register now says "two or three days"; four pages still published "next day",
so they contradict the bot until they rebuild.

```sql
SELECT spec->>'page_name', status FROM site_work_items
 WHERE spec->>'source'='owner-brief-2026-08-19'
    OR id='881c95ef-3fe9-4982-a676-4e61f5c5d368';
```
`index`, `faq`, `how-it-works` (severity high, priority 20) + the brief-starter
guide `881c95ef` (re-triaged).

**Verify at the SERVED page, never the item status** — a `complete` item is not a
repaired artefact, and this lane has already been bitten by a rewrite item that was
silently handled as a *rerender* and changed no copy at all:
```bash
curl -s "https://preview.webdesign.uk/index.html?cb=$(date +%s%N)" | grep -c "next day"   # want 0
```
Expect on every page: no "next day"; "two or three days" where turnaround is
mentioned; no approval/sign-off step; no post-payment "preview" naming (a *denial* —
"there is no preview to look at first" — is correct and wanted); £10/£200 on the
pages that discuss domains.

On failure the blocker is **persisted** — never grep pods:
```sql
SELECT occurred_at, jsonb_pretty(context) FROM agent_error_log
 WHERE error_code='CONTENT_VALIDATION_BLOCKER_DETAIL' ORDER BY occurred_at DESC LIMIT 3;
```

---

## 2. THEN: Phase 4 — handover + the delivery email (the next build)

**Not started: `sites.handed_over_at` does not exist** (checked 2026-08-19).

Source of truth: `PLAN_2026-08-14` Phase 4 mechanics + `PLAN_2026-08-17` decision 3.
The email carries: the ZIP link (dispatch `zip-deliverable-dispatch` with `{domain}`
— recipes in `sql_for_agents/459_zip_deliverer_agent_HOLD.sql`'s header, APPLIED),
the live-site link, a Netlify-connect invite (request-phase repeat), **both** domain
links (rent £10/mo subscription, buy £200 one-off with free transfer), and the
Stripe hosted portal. Needs the `sites.handed_over_at` migration plus a single
reader. `platform/mailer` is the sanctioned mailer. One council run; register entry
in the same commit.

**Plus, decided 2026-08-19 and not yet built:** a **6-week expiry** on the live link
(nothing expires today — serving is unbounded), and a **weekly chase email until the
customer confirms they have moved their files**, which needs a confirmed-transferred
state alongside `sites.handed_over_at`.

**Blocked on the owner, and this gates first revenue:** Stripe keys; the Stripe
webhook edge exception; second Nominet TAG (domain programme only).

> ### ⚠ OWNER DECISIONS 2026-08-19 (afternoon) — THREE THINGS TO BUILD IN PHASE 4
>
> 1. **The live link EXPIRES AT 6 WEEKS.** *"the preview link can expire in 6 weeks"*
>    (he said "preview"; he means the post-payment live link, which is never called a
>    preview in copy). **This is NEW WORK: nothing expires today.** Serving is
>    currently unbounded — no scheduled retraction, no retention job, no TTL — so a
>    6-week expiry is a mechanism that does not exist and must be built. The copy
>    still says "about a month", which under-promises against a 6-week reality. That
>    is the right direction and is deliberate.
> 2. **Chase by email, roughly WEEKLY, until the customer confirms they have moved
>    their files.** Needs: the delivery email (Phase 4), a repeating reminder, and a
>    *confirmed-transferred* state to stop it. No such state exists —
>    `sites.handed_over_at` does not exist either, so both land together.
> 3. **The move-it-yourself window is bounded: "within the next month."** Applied
>    2026-08-19 as `SQL_2026-08-19b` (writer_block only). **This RESOLVED the caps-ban
>    question the other way:** I was about to propose narrowing the 2026-08-09
>    "whenever you like" ban as an over-block, because it stopped a sentence stating
>    an attested freedom. The owner ruled the **ban is right and the copy was wrong** —
>    our hosting is not indefinite, so an unbounded time phrase is exactly the promise
>    that ban exists to stop. **Do not narrow that ban.** The carve-out written into
>    writer_block: a domain BOUGHT outright for £200 is the customer's property and is
>    genuinely theirs to move whenever, for ever. Bind the move, never the ownership.
>
> **CORRECTION to the paragraph below:** it says "do not build a month-long serving
> mechanism", written when the only question was whether customers might get *less*
> than a month. That still holds as written — but decision 1 above means an **expiry
> at 6 weeks** must be built. Do not read "nothing to build" as covering that.
>
> **THE LIVE LINK'S *MINIMUM* IS FINE — and my 2026-08-18 flag on it was WRONG IN DIRECTION.**
> Owner ruling 2026-08-19: *"The live link should be for a month or more; in reality
> the text should say a month."* The text already says "about a month"
> (`delivery_live_link_and_zip`), so **no wording change is needed**, and the
> mechanism already delivers it: sites serve from a git repo synced to B2 and
> **nothing takes them down** — no scheduled retraction, no retention job, no TTL
> (checked `scheduled_tasks`, the CronJobs, and `retract_asset_files`, which is
> manual, asset-scoped and called by no agent config). Serving is **unbounded**.
> I had reported "no month-long serving found" as a risk that customers might get
> LESS than a month; the same absence means they get MORE. Full correction in
> `WRONG_CALLS.md`, 2026-08-19. **Do not build a month-long serving mechanism.**
>
> ⚠ **The real exposure is the opposite one, and it is the owner's call:** because
> serving is unbounded, fact `keep_it_online` ("keeping the site online beyond the
> included month means the customer hosts it themselves") is **unenforced** —
> nothing stops a customer keeping our hosting indefinitely, free. Commercial
> decision, not a broken promise.
>
> ⚠ **The ZIP gap is real but narrower than I first wrote it.** The presign is
> `expiry_minutes: 10080` = **7 days** (`zip_deliverable_action.go`). "The ZIP is
> theirs permanently" is true once downloaded; the LINK to fetch it dies at 7 days
> and the claim does not say so. Under a 2-3 day build that is probably comfortable,
> but a customer away for a week loses the download. Owner to rule: lengthen the
> presign, or say so in the copy.

---

## 3. Owner rulings in force (do not re-litigate)

- **2026-08-19: delivery is "2 or 3 days."** `build_duration` re-attested, `value: 3`
  (the upper bound, because a stat field publishes a bare figure and a range cannot
  be one number — the stat must take the end that cannot over-promise). The hedge
  lives in `claim`/`writer_line`.
- **2026-08-19: leave the apex 302.** `webdesign.uk` → `webdesign.co.uk` stays.
  Not a defect. The chat API answers on `preview.webdesign.uk`.
- **2026-08-18: a better product beats a faster promise.** He judged pageflow
  probably cannot gain the other flow's checks and balances, and would rather move
  the delivery estimate. **Do not re-plumb the builder on this ruling** — it was
  about the trade-off, not the mechanism. `pageflow-builder` is still
  `recommended_builder` on **20 of 21** sites and owns the fleet's ONLY
  `briefing_questionnaire`.
- **2026-08-18: examples deferred.** No example-site links until this route has
  produced sites. (The `cross_site_domain` guard is NOT the reason — the allowlist
  `sites.content_data.allowed_reference_domains` already exists and
  `fundamentallyai.com` uses it. One row, if he ever un-defers.)
- **2026-08-04: every site goes through the framework.** Never hand-build a page.

---

## 4. What is LIVE, and how it was proven (not inferred)

| thing | state | proof |
|---|---|---|
| Delivery "two or three days" | **LIVE** 08-19 | bot, after a cache-beating restart: *"Usually two or three days from when we have everything we need from you."* |
| Chat prompt-maker | **LIVE** 08-18 | `434d2b64b` at the running service; smoke-tested — a darts-league enquiry got *"what would you want the site to actually do for people in the league?"* |
| Box deploy as makefile targets | **LIVE** | `make box-release` / `box-status` / `box-verify` (+5) |
| £1,200 swept from all 9 specs | **LIVE** 08-18 | seven phrases asserted nowhere; guard proven able to fail first |
| Questionnaire for any site type | **LIVE** 08-18 | 11 → 15 fields on `pageflow-builder`; backup taken |
| Refund ban narrowed | **LIVE** 08-18 | both real blocking sentences pass; retired promise still blocked |
| Commercial terms on served pages | **LIVE** | what-you-get / faq / how-it-works carry £10/£200, zero "preview" naming |
| Phase 3 ZIP deliverable | **LIVE** | register DGH-011, canary 8/8 byte-verified |
| Chassis | `v1.0.1314` | pod image + `build provenance` line |

---

## 5. STILL OPEN

1. **Phase 4** (§2) and its two over-promised delivery claims.
2. **`bugs_open/299`** — the home-page CTA names the Website Brief Starter and its
   href **dials the phone**. Filed, deliberately not patched. The producer question
   survives every rewrite: the section was written after the 268 fleet fix, so
   something still generates it.
3. **`what-you-get` shrink gate** — `SECTION SHRINK REFUSED, call-to-action 594→264
   visible chars (44% kept, floor 50%)`. Raising `section_shrink_floor` would
   silence a copy decision rather than make one, and it is the same CTA as 299.
4. **Contact email** `webdesign@contactforsales.com` (domain mismatch, item
   `a8d6f440`); Stripe webhook hostname; Stripe keys via terraform.
5. **HITL as a briefing step** — owner accepted the ordering: questions first
   (DONE), then HITL, routed through the **work-item** queue, which has a working
   screen. The orchestration HITL path has never fired: `collect_via_hitl` 0,
   `brief_answers` 0, `hitl_mode` 0 across 369 briefing orchestrations, against
   `briefing_answers` = 3 as the control. No consumer found for
   `system.notifications.ui`.
6. **A prompt-maker follow-up:** the conduct deliberately does NOT name the Website
   Brief Starter tool, because that tool's guide page was still selling the retired
   model. Once `881c95ef` lands, add the pointer.

---

## 6. Traps this joint work has paid for

- **The REGISTER is the wire.** Never steer via item-spec prose, never hand-edit
  HTML. `writer_block` edits are BY ANCHOR with exactly-once guards.
- **Never assert a fixed fact count on this register** — two lanes write it and the
  other edits IN PLACE. Compare against the row your transaction supersedes.
- **A bare-token `banned_claims` pattern bans the DENIAL too** (the `\brefunds?\b`
  precedent: 8 of 12 natural phrasings blocked). The negation guard scans
  **backwards only**, and bare "no"/"non-" are excluded cues by design. Some bans
  block denials **on purpose** — the `template` ban says so in its own reason — so
  read the reason before calling one a defect.
- **A guard that has only seen the state it was written for proves nothing.** Run it
  against the OLD data first and make it fail.
- **A `failed` work item is not work in flight**, but a `NOT EXISTS ... status NOT IN
  (terminal)` guard counts it as one. That silently skipped the index rebuild on
  2026-08-19; only the INSERT row count caught it.
- **`md5sum` cannot tell you which SOURCE a deployed binary came from** — these Go
  builds are not byte-reproducible. The chat binary now stamps its commit; ask the
  service. RUNBOOK corrected.
- **A `complete` work item is not a repaired artefact** — a rewrite item was handled
  as a *rerender* and changed no copy while reporting success.
- **Stat/figure fields publish only attested numbers; hedges stay prose.**
- **`submission` embeds its own differing copies** of mission_brief and
  roadmap_brief; **`content_direction` carries a rendered `formatted` duplicate**.
  Fix one copy, the other stays stale and authoritative.

---

## 7. Falsifiers

- A newer handoff in either lane dir.
- §1's four items: check the **served** pages, not the statuses.
- `sites.handed_over_at` existing (someone started Phase 4).
- Whether Stripe keys / webhook exposure / second Nominet TAG have landed.
- The webdesign.uk register's `updated_at` — two lanes write it.
- The chassis tag and its `build provenance` commit; it rolls often.
