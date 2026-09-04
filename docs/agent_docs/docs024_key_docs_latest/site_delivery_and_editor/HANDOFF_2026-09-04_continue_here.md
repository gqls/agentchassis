# HANDOFF 2026-09-04 — the rehearsal is DONE and 466 is LIVE; the owner performed the hosting steps himself and found four surfaces that promise mechanisms we do not have; a chassis roll is imminent and it carries the delivery recipient record

**Supersedes `HANDOFF_2026-09-03b_continue_here.md`.** That file remains the record of the rehearsal
night; everything still live is carried here.

**Sibling handoffs, not competitors:** `HANDOFF_2026-09-03_boxingonline_owner_review_continue_here.md`
(the boxingonline review thread). `bugs_open/475` and `bugs_open/477` are **other sessions' lanes** as
of today — see §6 before touching their files.

---

## 0. State in one paragraph

The delivery pipeline **works end to end and has now been performed by a human being.** `idea.uk` was
delivered for real on 2026-09-03 19:30:31Z — first `handed_over_at` and first `customer_access_tokens`
in the estate's history — and on 2026-09-04 the owner took the ZIP, dropped it on `app.netlify.com/drop`
and put the site online himself, in about **forty minutes**. That run is the most valuable thing this
lane has produced, because it turned four confident pieces of copy into measured falsehoods. `bugs_open/466`
is **LIVE** (verified at the binary, §3). Fleet is on **`239ab3626`** across all 28 stamped services and
core-manager; **the owner is rolling a fresh chassis within the hour and it includes core-manager** — §1.1
is what that roll changes and how to prove it. Delivery state: **2** tokens, **1 of 60** sites handed
over, **1** delivery record, **0** transfer confirmations. boxingonline delivery is still **HELD** on the
owner's own cut-line.

> **The finding that ties the whole day together.** Four separate customer-facing surfaces promised
> mechanisms the code does not have — instructions inside the ZIP, a site you can open locally, reminder
> emails, and a subscription-management link. **Every one was found by a person reading the words**, none
> by any check we own. Nothing in this estate compares what we *say* to what we *do*. That is the gap
> worth a lane of its own, and nobody has one.

---

## 1. NEXT, in order

### 1.1 ⚠ FIRST, AFTER THE ROLL LANDS — the delivery recipient record is currently HALF-LIVE

`bugs_open/477`'s lane made a delivery record who it went to, in the same statement that claims the
handover (`698b144fa`, `platform/delivery/handover.go:173` — `StampHandover` now refuses an empty
recipient). **The table half is live and the Go half is not:**

| half | state `[MEASURED 2026-09-04 16:0xZ]` |
|---|---|
| `site_deliveries` table | **exists**, 1 row, backfilled (`recorded_by = backfill-orchestration-states-778`) |
| `698b144fa` in the running chassis | **NO** — live stamp is `239ab3626`, and `merge-base --is-ancestor` says absent (control: a commit newer than the stamp, correctly absent) |

**So a delivery that happened right now would stamp `handed_over_at` and record no recipient** — the
exact failure 477 exists to close. It matters today rather than in the abstract because the owner holds
voucher `WD-KN3WU-9PZN4` for another run through `webdesign.uk`, and that run can reach delivery.

**The roll closes it. Prove it at the artefact, do not infer it from the tag:**

```bash
# per SERVICE, never per fleet
psql -c "SELECT pod_name, git_commit FROM service_binary_capabilities
         WHERE kind='build' AND pod_name LIKE 'agent-chassis-%' ORDER BY started_at DESC LIMIT 1;"
git merge-base --is-ancestor 698b144fa <that stamp>     # must be YES
git merge-base --is-ancestor <a commit newer than the stamp> <that stamp>   # must be NO — the control
```
core-manager does **not** write `service_binary_capabilities`; read its own log line instead
(`kubectl -n ai-persona-system logs -l app=core-manager --tail=4000 | grep -oE '[0-9a-f]{40}'`) and run
the same ancestor test with the same control.

### 1.2 Tell the `477` lane that migration `778` is applied but UNRECORDED

`site_deliveries` exists and holds its backfilled row, but `778_site_deliveries_...sql` has **no row in
`schema_migrations`**. Harmless in itself — the file is `CREATE TABLE IF NOT EXISTS` plus
`INSERT … ON CONFLICT (site_id) DO NOTHING`, so a re-apply cannot double-write — but it means the runner
still lists it pending, and the next session doing a broad `--apply` will re-run it.

**It is theirs to record, not mine** (`--record-only`). I am flagging it rather than fixing it because a
migration recorded by a lane that did not apply it is worse bookkeeping than one recorded late. **This is
the same class of miss I made myself yesterday with `752`** (applied, then left uncommitted 13 hours) —
which is why it is written here rather than left for someone to trip over.

### 1.3 The FAQ now contradicts a decision the owner made the same afternoon

The FAQ reword shipped and is **verified at the served page** (§3). The surviving sentence is honest
today: *"it's one payment rather than a subscription."*

⚠ **The owner then ruled that we WILL build the £10/month domain-rental link.** The day that ships, that
FAQ line becomes false — on the page whose entire job is telling customers what they are buying, and by
the same mechanism as the four falsehoods above: copy that was true when written and was never re-read
when the product moved.

**Whoever ships the rental link owns this line.** It is one sentence on `webdesign.uk/faq.html`, and the
cheap check is `grep -i subscription` on the served page at ship time.

### 1.3b ⚠ A NEW BUILD WILL DETECT ITS OWN DEFECTS AND PARK THEM — read this before running the voucher

Raised by the `boxingonline.com` lane, **verified here first-hand at the code, the rows and the RFC**.
The owner complained about three things on boxingonline this hour. **All three were already filed, by
our own seats, in his own words, and parked undispatched:**

| row | his complaint |
|---|---|
| `48c0f927` | the "last night's result" blog post is a general essay on upsets, naming no fighter and no result |
| `2c38eec5` | the fighter comparator has **0 `<option>`, 0 `<select>`** — not populated at all |
| `3acce370` (+3 more) | the fight-night countdown's only date is **two days past**, matchups are generic archetypes |

All three: `status='deferred'`, empty `handler_agent`, `spec.filing_mode='record'`.

**This is deliberate, and the code says so in its own words** (`write_audit_findings_action.go`,
`recordOnlyFinding()`): *"a verdict row — the seat's finding is recorded, the repair it would have
dispatched is not"*, and *"nothing promotes this row"*. **Seven seats run in record mode, all
`is_active`**: `brief-fidelity-auditor`, `content-quality-auditor`, `improvement-loop`,
`offer-analyser`, `reader-experience-auditor`, `site-review-agent`, `visual-design-auditor`.

**Why it is deliberate rather than broken:** RFC_056 §3 records that every model finding used to
dispatch through two doors, and **`bugs_open/238` is what one such rewrite did to a page that was
fine.** Parking opinions is the remedy for auto-rewrites damaging good pages.

> ### The part that bears on the voucher run
>
> `[MEASURED 2026-09-04 ~16:1xZ]` **3,184 parked verdicts across 39 sites**; 58 on boxingonline;
> **newest filed minutes before this was written**, so it is live and accumulating.
>
> **And it is not confined to old sites.** Of **12** sites created since record mode went live
> (2026-08-25), **11 carry parked verdicts** — the newest site, `copyonline.co.uk` (created
> 2026-09-03), already has **29**. The single exception, `seotools.co.uk`, has 316 work items of
> other kinds and is unexplained.
>
> **So a site built on the voucher this week will be audited, its defects written down, and none of
> them repaired.** The build works and delivery works; the loop that would fix what the seats find is
> off by design. If the point of the run is a *better* site than boxingonline, the lever is the
> release decision, not the build.

**RFC_056 has no phase that releases them, and that is not an omission.** Phase 3 — already executed
as migration `624` on 2026-08-25 — was what *put* the seats into record mode. The RFC states the exit
explicitly and repeatedly: a record row is *"undispatchable by construction until **a person or a
deliberate migration releases it**"*, and is *"designed for **a person to release**"*. **Nobody has.**

**Do NOT hand-promote.** The `boxingonline` lane holds 46 content rewrites behind the owner's word and
is right to — that is 46 simultaneous rewrites on the estate's first paid site, by the same machinery
that produced `bugs_open/238`. Each row carries its own `spec.release_recipe` for when the decision is
made. **The decision is the owner's and it is not this lane's to take.**

Also flagged by that lane and still open: **`5edadfbe` sits `failed` representing work that is DONE**
(the owner's two approved copy edits, applied and verified live). It will misreport until reconciled.

### 1.3c ⚠ TWO CORRECTIONS FROM THE `parked findings` LANE — one of mine was backwards, and the recipe is largely inert

**1. My FIFO warning had the direction wrong, and wrong in the dangerous direction.** I told them
releasing 3,184 rows was "3,184 queue positions at the back". **It is the front.** The release recipe
leaves `created_at` untouched, and `find_dispatchable_site` is `ORDER BY MIN(created_at) ASC` — so
released rows keep their **August** dates and sort **AHEAD of today's live work**, fleet-wide. My
version made a bulk release sound self-throttling; it is the opposite. (Their other note: the
promoter is already a throttle at 900s × `LIMIT 20` ≈ **80 rows/hour**, so the population is ~40
hours of drip **if nobody defeats it by releasing straight to `triaged`**.)

**2. The documented `spec.release_recipe` is INERT for ~89% of the rows, and fails SILENTLY.** Every
parked row's `routed_status` is **`detected`**, not `triaged`. A `detected` row only reaches dispatch
via `detected-item-promoter`, whose live `pre_query` carries a fifth door from migration `629`:

```sql
(COALESCE(wi.spec->>'origin','') <> 'model_opinion') AS origin_ok
```

**`write_audit_findings` — the same producer that parks these rows — stamps
`spec.origin='model_opinion'` on them. 2,824 of 3,184 carry it.** Simulated across all five doors:
**352 flow, 2,832 stick.** A stuck row changes status, gets a named handler, and **errors on
nothing** — so it reads as released and is not.

> **The trap for the `boxingonline` lane specifically:** running the recipe on its 46 rewrites would
> move most of them from `deferred` to `detected`, where they would sit **looking released**. The
> `parked findings` lane is messaging them directly.

**Marked honestly, by them:** the door has never been *observed* to fire — zero `model_opinion` rows
have ever reached `detected`, because record mode parks them first. The claim is read off live config,
not observed, and they have filed a `090` diagnosis run rather than assert it alone.

**Also from them, and it clears a hold I wrote:** the chassis roll **has landed** — both pods stamp
`06c0b18f2`, started 16:01Z. I confirmed independently. `service_binary_capabilities` briefly listed
the old pods because it is a **two-hour survivor window, not a history**; `kubectl get pods` settles it.

### 1.4 Captions for the instructions page — MINE, blocked on placement

The owner's ten screenshots (`/home/ant/Downloads/idea_uk_netlify/`) are the right ones, and the
**signed-out "This site is private" wall is the most valuable image on the page** because it is the one
thing a customer will otherwise never see. The `475` lane owns the page mechanism and its council round
**c8ed56d2 came back APPROVED at 15:35:39Z**. I write captions once they give me placement.

### 1.5 `bugs_open/474`'s two frontend halves are still unclaimed

Both need a dashboard build, which is why they have sat: (a) Work Items filters pipelines server-side and
the dropdown names two of eight live pipelines, leaving **1,984** items reachable only via the unlabelled
catch-all; (b) the approve button's *visibility* condition (`isCheckpoint`) is not its *submittability*
condition (`editedReviewData != null`), so it renders for items it cannot submit. The strongest fix for
(a) derives the options from the data, so a new pipeline cannot go missing by omission.

### 1.5b ⚠ CORRECTED — payment→build IS automatic, and I told the owner otherwise

I read `billing.Service.HandleWebhook`, found it only marks the order paid and logs, and reported
that paying does not start a build. **Wrong.** The webhook does not dispatch; a poller does the join,
which is the better design because it survives a lost webhook.

| mechanism | `[MEASURED 2026-09-04 ~16:2xZ]` |
|---|---|
| `order-intake-collect` → `order-intake-collector` | every **900s**, `enabled`, last triggered 16:03:46Z, **80 COMPLETED** runs |
| `build-pipeline-trigger` | every **30s**, `enabled`, last completed 16:15:30Z, **503 COMPLETED** runs |

**The join is the order reference, not the brief** (owner ruling 2026-08-26, *"the brief will
change"*): the chat mints `BR-XXXXXX` when it stores the brief, the customer quotes it at payment,
and it lands in `billing_orders.external_reference`. The one real order carries `BR-9AUZ59`.

**Two documented paths that do NOT produce a build**, and any copy about payment must survive both:
a paid brief **naming no domain** → `needs_human_review` (*"inventing a domain here would put a name
nobody chose on a paid order"*); a domain **past `queued`** → `needs_human_review` (*"a paid customer
swallowed by a unique constraint is this product's worst failure"*).

**And delivery is still gated on a human approval** — `delivery-review-filer` files a checkpoint,
a person approves, then the zip and the email. So there is no SLA to quote to a customer, and any
timescale on a payment-success page would be invented.

> **The lesson, which is the day's fourth of this shape:** I read ONE function and generalised from
> it. `HandleWebhook` not dispatching is true; "payment does not start a build" is false. **A
> mechanism's absence at one site is not its absence** — the seam was one table and one
> `scheduled_tasks` row away, and I never looked.

### 1.6 Still owed to other lanes / the owner

- **`bugs_open/420` §C, carry it WITH the next delivery ask:** *"what CONSENT STATE may a classifier
  write on a contact row?"* Hand the answer back to the `bugfix_417_420` lane **verbatim**.
- **boxingonline delivery remains HELD**, and the **logo question is still with the owner.** Its four
  remaining review points are other lanes' code (`bugs_open/427`, `bugs_open/114`).
- **Four `icon` assets and one `og_card` reach no page**, and one `content_hero` is unserved where three
  siblings serve. Cited into `bugs_open/114` rather than filed fresh.

---

## 2. OWNER RULINGS 2026-09-04

| ruling | what it settles |
|---|---|
| **Instructions live in "all three" places** | a correctable page, a `README.txt` pointer in the ZIP, and per-site slots |
| **The instructions page is GENERIC, public, framework-built, on `webdesign.uk`** | so no exception to the 2026-08-04 "every site goes through the framework" ruling is needed. Per-customer content moves to the email and the README, which already know the domain |
| **"Leave it"** — the still-false *"The ZIP comes with instructions"* line | stays until the page exists, rather than shipping a stop-gap. ⚠ Risk stated and accepted: if a build reaches delivery first, that line goes out again. Mitigating fact: **the next build is his own trial run**, so the recipient would be him |
| **Follow-up interval: 3 days** | applied to `775` (the `477` lane's sender, seeded disabled) |
| **The delivery recipient goes in its own delivery record**, not a column on `sites` | because `sites` is read by many things and `420`'s contract split governs which address may live where |
| **Build the £10/month domain-rental link** | a reversal of the earlier position, and the trigger for §1.3 |
| **Accounts: not yet** — delegated to me, decided | build the link, defer accounts, and **capture the Stripe customer id the first time a subscription creates one**. Full reasoning and the named trigger: `PLAN_2026-09-04_preliminary_customer_accounts…` §4b |
| **Netlify's AI editor is an ASSET for the handoff**, not a threat | his reframe, and it is the right one — it turns *"no changes are included"* from an apology into an offer |
| **The HITL loop defaults to the SYSTEM approving changes** | *"for the moment"* — he does not want to check every fix. **Read the scope: it is conditional.** *"The damaged pages need fixing, the good ones don't. If a good one gets rewritten at this stage (fresh build) then as long as it is still good even if different then it is still ok."* He is accepting `bugs_open/238`'s risk **because the sites are fresh**. That acceptance does not obviously survive to a customer-approved or delivered site — re-confirm rather than inherit |
| **All parked findings get worked through**, by the new `parked findings` lane | *"I can approve each as they go if they wish."* Handed over in full this hour, with the 3,184 measurement, the RFC_056 citation correction and the `457` boxingonline ordering trap. **⚠ My sequencing warning was BACKWARDS — see §1.3c** |
| **Domain for the next run: `paper-cups.com`** | and **the build does NOT need it** — *"We can use a random word subdomain on ugg2.com like netlify does."* Nameservers pointed later, by the `dynadot` lane, *"when we need to"* |
| **`ugg2.com` is NOT the destination** | *"I'd want it hosted in our normal place (backblaze, cloudflare) eventually rather than ugg2.com."* So the end state is the customer's own domain, Cloudflare-fronted, served from Backblaze — passed to `dynadot` because it may change which records they set |
| **The `stripe` lane fixes `/pay/success`** | routed with the measurement and its controls. They have taken it, and found the framework CAN express it (`/pay/success/index.html`; extensionless directory URLs already resolve on that vhost), so no exception to the 2026-08-04 ruling is needed |

---

## 3. WHAT CLOSED TODAY, each verified at the artefact

| item | evidence |
|---|---|
| **`bugs_open/466` is LIVE** | core-manager's own stamp `239ab3626`; `merge-base --is-ancestor 33dfeed3a` → **YES**; control (a commit newer than the stamp) → absent. So the owner's copy approvals now fan out |
| **the delivery rehearsal, end to end** | all four agents ran, owner approved 19:20:58Z, zip cut, email arrived, ZIP and both links confirmed working by him |
| **`zip-link-refresher` actually refreshes** | the owner's own test design — set a short lifetime and watch, instead of waiting 7 days. `stored_url_expires_at` = **2026-09-10 21:08:16Z** against a token minted **19:30:31Z**: the presign advanced ~1h38m after minting. **The 30-day promise is real; it is not secretly 7** |
| **migration `776`** | applied **12:05:25Z**. The delivery email no longer promises reminders nothing sends. Config, so live on apply — no roll needed |
| **the FAQ reword** (`section_edit 75efbed1`) | `complete` is not evidence, so: served `webdesign.uk/faq.html` **lm 15:13:17Z**, promise phrasing **0**, liveness control (FAQ markers) **80**, invented-path control **404**. The one surviving `subscription` is the honest sentence — and §1.3 is its expiry date |
| **migrations `750` + `752`** | applied 2026-09-03 22:06:22Z / 22:03:42Z, both recorded |

---

## 4. FILED TODAY

| number | what |
|---|---|
| **`bugs_open/475`** | the delivery email tells the customer the ZIP contains instructions, and it contains none. Owner ruled "all three"; **another lane owns the mechanism** |
| **`bugs_open/476`** | diagnosed and deliberately **not** cheaply "fixed" — the honest fix is the instructions page, not a patch to the claim |
| **`bugs_open/477`** | the confirm button promises to stop reminders nothing sends, and `transfer_confirmed_at` has no reader. **Another lane owns it**, both council rounds approved |
| **`PLAN_2026-09-04_preliminary_customer_accounts…`** | for the client-accounts thread to own. Carries the finding that sets the problem: **60 sites, 42 with a `network_id`, 1 distinct network between them** — we cannot answer "which sites belong to this customer?" today |
| **`OPTIONS_2026-09-04_what_it_would_take_to_set_up_hosting_accounts…`** | four costed options; A is closed by email verification and credential custody, B is a mechanism not a destination, C is the owner's call, **D is done and ships regardless** |
| **`DRAFT_2026-09-04_customer_instructions_copy_v3…`** | the instructions rewritten from what actually happened. **v2 is kept, not corrected** — the gap between what I wrote from knowledge and what happened in a browser is the most useful thing in this directory |
| **voucher `WD-KN3WU-9PZN4`** | minted for another run through `webdesign.uk`. Recorded in `RUNBOOK…` — see §5.4 |

---

## 5. THE DAY'S OWN MISTAKES — five, all in `WRONG_CALLS.md`

1. **Migration `752` applied and left uncommitted for 13 hours.** Found by re-running `git status` on
   re-orientation, not by any check. Verified the disk file's md5 still matched the `schema_migrations`
   checksum before committing. Landmine filed.
2. **I attributed migration `767` to a lane from its FILENAME.** A peer corrected me with evidence. On
   this tree `git log` authorship is `cqls` for every commit, so **a filename is not a provenance
   signal** — the council correlation is.
3. **My uncommitted-migration check produced two false positives out of two.** No age gate (it fired
   inside a 60-second apply→commit window) and filename-equality (it flagged a renumber). The true
   estate rate is **zero orphans in 556**. A detector that has never been right about anything is worse
   than none, because its output gets quoted.
4. **`776`'s verify block was INERT and I proved it only by mutation.** `position(x in NULL)` is NULL,
   so both `<> 0` and `= 0` silently fail to fire. **The `761` lane hit the identical bug the day before
   and I had quoted their commit message that morning** — the existing landmine's title said "jsonb path
   with `<>`", which is why I did not recognise myself in it. Title widened.
5. **I handed the imagery task over without the trap I had written in my own sweep script.** The `475`
   lane's attribute-only grep found "zero content images"; the heroes are CSS `url()` backgrounds. And
   separately I joined `assets.storage_path`'s filename (a uuid) against `content_data` and got "14 of 14
   unreferenced" — the deployed name derives from `asset_key`. **The real answer is 14 of 20 serve.**

> The through-line in 3, 4 and 5 is one thing: **a check that cannot come out the other way is not
> evidence.** Two of these were dated and marked `[MEASURED]` and still could not have failed.

---

## 6. LANE BOUNDARIES — read before editing

- **`bugs_open/475` lane owns** the instructions-page mechanism (`{{instructions_link}}` placeholder,
  LinkConfig, the framework page on `webdesign.uk`, the zipper README) and one supplied-image canary.
  **`prepare.go` is theirs — do not touch it.**
- **`bugs_open/477` lane owns** the follow-up sender and `handover.go` / `delivery.go`. `774` and `775`
  are written and **not applied**.
- **The `stripe` lane owns** all payments: the £10/month rental Payment Link, the `/pay/success` and
  `/pay/cancel` 404s, and the "Fine Tune" branding on the live Stripe account.
- **This lane holds** the delivery half — handover stamp, tokens, the delivery email, the customer-facing
  `/c/` and `/d/` pages — and the captions in §1.4.

---

## 7. FALSIFIERS — check before believing this file

A newer handoff in this directory · **whether the chassis roll has landed** (re-read the stamp **per
service**; `239ab3626` is the pre-roll value here) · whether `698b144fa` is in the running chassis, which
flips §1.1 · whether `778` has been recorded · whether the rental link has shipped, which flips §1.3 ·
`customer_access_tokens` still **2** and `handed_over_at` still **1 of 60** · whether `474`'s two frontend
halves have been claimed · whether the owner has spent voucher `WD-KN3WU-9PZN4`, because that run reaches
delivery and carries the "leave it" line.

## 8. Read order, cold

This file → `SUMMARY_2026-09-03…` → `README_where_we_are.md` (the owner's plain-prose log, **newest at
the bottom** — today's entries are the Netlify run in his own words) →
`DRAFT_2026-09-04_customer_instructions_copy_v3…` **beside v2**, which is the point →
`PLAN_2026-09-04_preliminary_customer_accounts…` → `RUNBOOK…` (the rehearsal section, the corrected
dispatch envelope, and the voucher recipe) → `bugs_open/475`, `476`, `477` → the superseded
`HANDOFF_2026-09-03b_continue_here.md`.
