# HANDOFF 2026-09-03b (evening) — the owner's copy edits are LIVE, the approve-to-apply seam is fixed and council-approved, and the delivery rehearsal is running on a non-customer site, blocked on one click

**Supersedes `HANDOFF_2026-09-03_continue_here.md`.** That file is still the record of the day's
first half and carries two in-place corrections worth reading (the links host; the `cta-subtitle`
counting trap). Everything live is carried here.

**Sibling handoff, not a competitor:** `HANDOFF_2026-09-03_boxingonline_owner_review_continue_here.md`
is the boxingonline owner-review thread. This one is the delivery pipeline.

---

## 0. State in one paragraph

boxingonline.com (site `d2aa5206-73bc-4707-a69c-2702c1eb9152`, order BR-9AUZ59, first PAID build)
serves at **https://boxingonline.ugg2.com**, and the owner's two approved copy edits went live at
**17:32:30Z** — verified at the served page, four checks, all pass. Delivery for that site remains
**HELD** on his own cut-line; `customer_access_tokens` is still **0** and **0 of 60** sites have ever
been handed over. The approve-to-apply seam that failed on his first use (`bugs_open/466`) is fixed,
council-**APPROVED** at round 3, and **inert until a release**. He then authorised a **full** delivery
rehearsal on a non-customer site: it is running on **idea.uk**, step 1 of 4 is done, and it is now
run END TO END: he approved at 19:20:58Z, the zip cut, the email arrived, and he confirmed the
zip and both links work. **Three bugs were filed from the rehearsal itself** (`474` twice over, the
`651` recipe defect, and `475` — the email making a false claim to the customer); the first two are
fixed and `475` has the owner's ruling and draft copy. Fleet on **v1.0.1359**.

---

## 1. NEXT, in order

> **THE REHEARSAL IS DONE. All four agents ran, the owner approved, the email arrived, the zip
> opened and both customer links worked.** First `handed_over_at` stamp and first
> `customer_access_tokens` rows in the estate's history. Full account in `RUNBOOK…`.

1. **`bugs_open/475` — the delivery email tells the customer the ZIP contains instructions, and it
   contains none.** The owner found it by reading the email as a customer would. It is a FALSE claim
   to a paying customer, on the line their whole "you host it yourself" story depends on, and the
   failure mode is them blaming the download. **Owner ruled "all three":** a correctable page, a
   `README.txt` pointer in the ZIP, and per-site slots. Draft copy written to house style:
   `DRAFT_2026-09-03_customer_instructions_copy.md`.
   ⚠ **Do not ship §4 of that draft.** Nobody has ever performed the hosting steps. Somebody must
   drag this actual ZIP onto `app.netlify.com/drop` once, end to end, and write down what they saw —
   shipping unperformed instructions repeats 475's own root cause one level along.
2. **THE THREE-LIFETIME FINDING, and it has a deadline.** The presign lasts **7 days**, the email
   says **30**, the tokens run **42**. Not a broken link — `HandleZipDownload` renders a refresh page
   rather than redirecting to a stale presign — but the 30-day promise rests on `zip-link-refresher`,
   which is scheduled, enabled, and **has never refreshed anything** because there were no zip tokens
   until tonight. **Before 2026-09-10:**
   `SELECT purpose, stored_url_expires_at, expires_at FROM customer_access_tokens WHERE purpose='zip_download';`
   `stored_url_expires_at` advancing is the proof. If it does not move, our 30 days is really 7.
3. **`bugs_open/466` needs a release to do anything.** Committed `33dfeed3a` → `bfc2925ce`, council
   APPROVED (`d04c1bc1`). Until an admin/core-manager image carrying it rolls, the owner's copy
   approvals still fail the way they did this morning.
   `git merge-base --is-ancestor 33dfeed3a <the admin-dashboard build stamp>`. **`make release` is
   the owner's.**
4. **Migration `752` is written and NOT applied, and its resubmission is NOT sent.** The council's
   REVISE of `751` was right twice: pointing `review_data` at the whole `input_data` map is
   unconstrained (a future change to input_data silently changes what the owner is shown), and it
   left the backfilled item and future items with two shapes for one field. `752` narrows it to an
   object the dispatcher composes, with the RUNBOOK recipe updated in step. **Submit BEFORE applying
   this time** — the round's other fair hit was that `751` was applied before review, so the seats
   audited a change already live. `RESUBMIT_CORR=95eaf57c-74d4-46f3-bad3-f8ef46bc1df8`.
5. **`bugs_open/474` §2 is unclaimed** — Work Items filters by pipeline server-side, defaults to
   `build`, and the dropdown names two of the eight live pipelines. Five pipelines holding **1,984**
   items are reachable only via the unlabelled catch-all. Strongest fix derives the options from the
   data so a new pipeline cannot go missing by omission. Frontend build.
6. **`bugs_open/474` candidate 2 is also unclaimed** — the approve button's VISIBILITY condition
   (`isCheckpoint`) is not its SUBMITTABILITY condition (`editedReviewData != null`), so it renders
   for items it cannot submit. Frontend build.
6b. **`bugs_open/477` is CLAIMED by another session (2026-09-04) — do not compete.** The confirm
   button promises to stop reminders nothing sends, and `transfer_confirmed_at` has no reader. A
   fresh session picked it up, asked before touching anything, and is taking candidate 3 (honest copy
   now → scheduled follow-up sender suppressed by `transfer_confirmed_at` → restore the stronger
   wording). They own `delivery.go` and `handover.go`; this lane stays out.
   > **⚠ THIS ENTRY EXISTS BECAUSE THE NEXT LIST OMITTED IT.** I filed 477 the same evening and never
   > added it here, so `who-owns.py` read OWNED off the filing commit alone and they had to ask
   > whether I was already building it. **Filing a bug is not the same as queueing it**, and the
   > handoff is what the next session actually reads. Two things I owed them that were in neither the
   > bug file nor this one: the owner has **NOT** ruled on the follow-up (his "maybe repeat them in
   > the first follow up email" is a suggestion, not a decision — do not cite a ruling), and
   > **`idea.uk` is the only site in the estate a follow-up sender could select** (`handed_over_at`
   > set, `transfer_confirmed_at` NULL, 1 of 60) — whose delivery address is the owner's own, so a
   > working sender WILL email him unannounced unless someone warns him first.

7. **Carry `bugs_open/420` §C to the owner WITH the next delivery ask** — unchanged, still owed:
   **"what CONSENT STATE may a classifier write on a contact row?"** Hand the answer back to the
   `bugfix_417_420` lane verbatim.
8. **boxingonline's delivery is still HELD** on the owner's own cut-line, and its four remaining
   review points are other lanes' code (`bugs_open/427`, `bugs_open/114`). The logo question is
   still with him.

---

## 2. OWNER RULINGS TODAY

- **"Carry on fixing the tools that make these work properly."** Delivery for boxingonline NOT
  unblocked.
- **Full rehearsal, on our own site** — chosen over a dry prefix and over waiting, in the knowledge
  that it burns that site's once-only handover stamp and sends a real email through the live SMTP
  account.
- **Delivery email recipient: `aaa@designconsultancy.co.uk`.**
- Earlier and unchanged: *"palette: the cream/off-white STANDS"* · *"header stays LOGO-ONLY. Closed."*
  · the CTA line *"cut it"* (applied, live) · *"Guides should be a type of their own"* (routed to
  session `428`).

---

## 3. WHAT CLOSED TODAY, each verified at the artefact

| item | evidence |
|---|---|
| **the owner's two approved copy edits** | ✅ **re-verified 2026-09-04 07:41Z after an overnight republish** (lm `04 Sep 03:46:12Z`) — all seven checks still pass, so they survived a regeneration. Originally: served `/index.html` lm **17:32:30Z**: his verbatim line present ×1, old subtitle gone, `calendar below` = 0, **rendered `cta-subtitle` elements = 0** with no empty `<p>`, excerpts = 6, `\| Boxing Online` = 0 |
| **`bugs_open/466`** | fixed, 9 tests, **10 mutations across 4 council rounds all killed**, APPROVED `d04c1bc1`. ⚠ **committed, NOT live** |
| **`bugs_open/474`** | filed AND fixed same hour; migration `751` applied by hand, verified by independent re-read |
| **the `651` dispatch recipe** | corrected; the envelope that actually lands is proven (§5) |
| **the links host** | `links.webdesign.uk` VERIFIED — `/c/` and `/d/` both live and route-specific against an invented-token probe with a 404 control. Recipe in `RUNBOOK…` |
| **delivery prerequisites** | migration 650 applied · `DELIVERY_SMTP_*` env + secret present · all four agents active. Remaining first-time risk is now **three** things, not four: the agents have never run, no token has ever been minted, DKIM/DMARC at the mail host is still unchecked |

**⚠ `cta-subtitle` CANNOT be checked with a bare grep.** The page carries the string twice — the
rendered `<p class="cta-subtitle">` and the CSS rule `.cta-subtitle {…}` in its own `<style>`. The
rule survives the edit, so a bare count reads **1** for ever and looks like a failed edit; an older
monitor reported exactly that on this publish. The check is `class="cta-subtitle"` = 0, with the bare
count ≥1 kept as the **liveness control**.

---

## 4. FILED TODAY

| number | what |
|---|---|
| **`bugs_open/474` §2** | **the same item is in a pipeline the dashboard does not offer** — Work Items filters server-side, defaults to `build`, and the dropdown lists only build/content/all. `[MEASURED]` **five of the eight** live pipelines have no option, holding **1,984** items (`design` alone 1,933). Owner unblocked by "all pipelines"; no candidate taken, all are frontend |
| **`bugs_open/474`** | **the delivery review item was filed in a shape the approve screen refuses.** The button renders on `isCheckpoint` alone but submits only with `spec.review_data`, which `create_work_item` never writes. `[MEASURED]` 1 of 27 checkpoint rows lacks it, and it is the only delivery review ever filed. **Our own `prepare.go` predicted it and assigned it to someone else in a parenthesis.** Fixed by `751`; candidate 2 unclaimed |
| **migration `750`** | copy-editor's approval fans out — `fan_out_from: "edits"`, `defaults: {edit_type}`, `include_fields: ["domain"]`. **Not yet applied**; inert until 466's binary rolls, harmless before |
| **migration `751`** | **APPLIED + recorded.** `delivery-review-filer` files `review_data` from `input_data`; the already-filed item patched from its own spec |
| **LANDMINE** | a dispatch REFUSED for missing headers is indistinguishable from queue latency — and the do-not-retry guidance then tells you to wait for ever |
| **LANDMINE** | `checkpoint_for_review` writes a FIXED key set, so `on_approve.include_fields` could only ever copy nulls |
| **RFC_022 addendum** | the escalation trigger for `on_approve.fan_out_from`, with the RECURSIVE census query |
| **register `ADM-012`** | approval fan-out, with the delivery-gate hazard recorded |

---

## 5. THE DAY'S OWN MISTAKES — three, all in `WRONG_CALLS.md`

1. **I told the owner twice a dispatch was "inside the measured latency window". It had been REJECTED
   within seconds** — `INCOMING_MESSAGE_REJECTED :: missing required header(s): client_id,
   orchestration_id`. `651`'s recipe named three headers; five are required. A refused message and a
   slow one leave the identical evidence (no orchestration row), and CLAUDE.md correctly says not to
   retry on that — so the incomplete recipe and the sound guidance **combine into a trap**.
   **I used a remembered number as a reason not to run a one-line check.**
   > **THE ENVELOPE THAT LANDS** (proven: COMPLETED in under **25 seconds**, so the "29 minute"
   > figure did not apply at all): `orchestration_id`, `orchestration_name`, `client_id`,
   > `request_id`, `message_id`, `step_name`, `message_type`, `action`, `from_agent_type`,
   > `from_agent_id`, `responses_topic`. Copy from `097_TRIGGER_council_review_v1.sh:300-310`, a
   > script known to run today — never from a prose recipe.
   > **And run `kafka_verify_landing "$corr" 30` FIRST, not as an escalation:** 13 = really latency,
   > **12 = consumed and refused**, and it names what is missing. A green `kafka_publish_checked`
   > receipt asserts PUBLICATION, not acceptance; only the second predicts a run.
2. **I built two guards on things `LANDMINES` already warns about**, both caught by the council, not
   by me: a top-level `jsonb_each` walk blind to `sub_workflow` steps, and a refusal resting on
   `page_components.updated_at`, which cannot distinguish a content edit from a status-only write.
   **I grep the register for what I am FIXING; both traps were filed under what I was fixing it
   WITH** — the query idiom, the column. Run the grep twice.
3. **My `sweep_site_defects.sh` §1.4 is blind by construction** (found by session `332`): it greps
   served HTML on pages whose own JS overwrites it from `/data/news-archive.json`. Fleet-wide, five
   hosts. They own the fix.

---

## 6. FALSIFIERS — check before believing this file

A newer handoff in this dir · whether the owner approved `e370e0bb` (and whether he pressed **approve**
or **resolve** — `result ? 'approved_by'` is the only thing that counts) · whether a release has
carried `33dfeed3a`, making 466 live · whether migration `750` has been applied · `customer_access_tokens`
still 0 and `sites.handed_over_at` still 0 of 60 · a NEWER chassis roll (v1.0.1359 here; re-read the
stamp **per SERVICE**) · the council verdict on `95eaf57c` · whether `474` candidate 2 has been taken.

## 7. Read order, cold

This file → `SUMMARY_2026-09-03…` → `README_where_we_are.md` (the owner's plain-prose log, newest at
the bottom — the last three entries are today) → `RUNBOOK…` (**the rehearsal section and the corrected
dispatch envelope**) → `bugs_open/466` and `bugs_open/474` → `HANDOFF_2026-09-03_continue_here.md`
(superseded, but its two correction blocks are the working record) → the boxingonline 09-03 handoff.
