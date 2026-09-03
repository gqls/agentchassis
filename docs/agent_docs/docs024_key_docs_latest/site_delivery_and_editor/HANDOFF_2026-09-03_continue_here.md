# ⚠ SUPERSEDED 2026-09-03 evening by `HANDOFF_2026-09-03b_continue_here.md` — read that first

> This file remains the working record of the day's first half, and its two in-place correction
> blocks are worth reading: the links host (§1 item 4) and the `cta-subtitle` counting trap
> (§1 item 1). Everything still live is carried into 09-03b.

# HANDOFF 2026-09-03 (evening) — the first paid site is verified at the artefact on four of the owner's points; delivery is still HELD, and the delivery chain has never run

**Supersedes `HANDOFF_2026-09-02b_continue_here.md`**, which grew fifteen in-place correction
blocks over one day and is no longer a cold read. Everything live is carried here; the old file
stays as the working record of how today actually went, including the wrong turns.

**Sibling handoff, not a competitor:** `HANDOFF_2026-09-03_boxingonline_owner_review_continue_here.md`
(same dir) is the boxingonline session's owner-review thread. This one is the delivery pipeline.
Anything dispatch-shaped is here; anything about what the owner thinks of the site is there.

---

## 0. State in one paragraph

boxingonline.com (site `d2aa5206-73bc-4707-a69c-2702c1eb9152`, order BR-9AUZ59, first PAID build)
serves at **https://boxingonline.ugg2.com**. Today four of the owner's fourteen review points closed
**at the served artefact, not at a status**: the transparent logo, the card decks, the contact-page
404, and the analytics tag with its consent banner. His approved copy edit is applied at the source
and awaiting the next publish tick. **Delivery remains HELD** on his own cut-line
(*"build and fix everything before approval"*), and `customer_access_tokens` is **0**.
**UPDATE 17:32Z: his approved copy edits are now LIVE at the served page** (§1 item 1), so five of
the fourteen are closed at the artefact rather than four. The four
remaining review points are other lanes' code, and he ruled today: *"carry on fixing the tools that
make these work properly."* Chassis and adapter both on **v1.0.1359**.

---

## 1. NEXT, in order

1. ~~**Verify the owner's approved copy edit at the SERVED page.**~~ **DONE 2026-09-03 17:32:30Z —
   ALL FOUR CHECKS PASS at the served artefact.** Published lm `Thu, 03 Sep 2026 17:32:30 GMT`
   (the tick ran ~17 min later than the ~17:15Z estimate; the publisher job was mid-run at 17:24Z).
   Owner's line present ×1, old subtitle gone, `calendar below` = 0, rendered `cta-subtitle`
   elements = 0 with no empty `<p>`, excerpts = 6, `| Boxing Online` = 0.

   > ⚠ **CHECK 3 AS WRITTEN ABOVE CANNOT BE EXECUTED AS A BARE GREP, and doing so reads as a
   > FAILURE.** The served page carries **two** things called `cta-subtitle` and they are different:
   > the rendered `<p class="cta-subtitle">` **and** the CSS rule `.cta-subtitle { margin-bottom:
   > 2rem; }` inside the page's own `<style>` block. The CSS rule survives the edit, so a bare
   > `grep -c cta-subtitle` reads **1** for ever. The check is `class="cta-subtitle"` = 0, with the
   > bare count ≥1 kept as the **liveness control** that the section rendered at all. An older
   > monitor reported `cta-subtitle els=1` on this very publish; that is the bare-string count, and
   > it is a false alarm — re-measured both ways at 17:33Z, the single remaining occurrence is the
   > CSS rule. No checklist-5.1 empty `<p>` is possible here either: the template gates the element
   > on `{{if .subheadline}}` and Go treats `""` as false, so an empty value omits it entirely.
2. **The logo question is with the OWNER and is the only open ask on this site.** New today, and
   NOT covered by any prior ruling: the mark carries **zero pixels within ±60 of brand red
   `#C0392B` or gold `#D4A017`** (53% blue / 45% neutral), and the prompt asked for *"a stylised
   boxing glove or ring ropes"* while a fist-in-a-diamond came back — a **subject-fidelity** miss
   (417 family), not a transparency one (424, closed). Reproduces across TWO artefacts: this lane
   on the source (206,018 opaque px), the designer seat on the served copy (16,372). Their
   consequences, credited not re-derived: 48.4% of the ink at or below weak contrast on the
   near-black header vs 25.4% on cream, and detail collapse at `max-height:40px`. **Sequencing
   agreed with that seat: fix the mark first; whether the header still fails to name the site is a
   question for him AFTER he sees a new one — ruling (2) "header stays LOGO-ONLY" is CLOSED and
   must not be actioned against.** ⚠ Do NOT re-seed to obtain a regen (§9.5, `bugs_open/420`).
3. **The four remaining review points are not this lane's to fix** — his ruling was to fix the
   mechanisms. Verify at the artefact when their lanes land: articles with no news, the dataless
   comparator, the calendar with no calendar (all one root, **`bugs_open/427`**), and imagery —
   where the sharper framing found today is that **12 generated images on this site are referenced
   on ZERO pages** (6 content-heroes, hero-about, hero-contact, hero-calendar, 3 icons), so the
   problem is that components cannot hold them (**`bugs_open/114`**), not that too few exist.
4. **The 651 delivery rehearsal — and the recommendation is to rehearse on a NON-customer site.**
   Read at the live config today: the chain is `delivery-review-filer` → owner APPROVES on
   admin.apis.uk → `zip-deliverable-dispatch` → `zip-deliverer` → `delivery-email-sender`, and
   **not one of the four has ever run** (zero orchestrations all time, zero `customer_access_tokens`
   fleet-wide, zero zips on this site). Four unexercised agents, a token type never minted and an
   ~~**unverified links host**~~ is a lot to meet for the first time on a paying customer's delivery.

   > **CORRECTED 2026-09-03 evening — the links host IS verified, and the earlier reading was mine.**
   > I probed `/` and an invented path, both 404, and concluded "unverified". That was the wrong two
   > URLs: **the routes are `/c/:token` and `/d/:token`, and neither was among them.** Probed with a
   > deliberately invented token, both answer **200 with route-specific HTML**, against an invented
   > top-level path that returns nginx's own 404 — so the control discriminates and this is not a
   > catch-all. `/d/` says *"That download link is no longer active"*, which is the correct refusal
   > for an unknown token. `/c/` serves the button page, and that is **deliberate and owner-ruled
   > (2026-08-25)**: `HandleConfirmPage` performs **no database access at all**
   > (`internal/core-manager/handlers/delivery.go:145-188`, which states it as a property), because a
   > read-only lookup would hand anyone holding a guessed token a free validity oracle. The customer
   > learns on pressing. Do not file this as a bug — I nearly did, off the probe alone, before
   > reading the handler.
   > **So the host is one fewer first-time risk, and "it cannot be tested until a token exists" was
   > also wrong: the refusal path is exactly what an invented token tests.** Recipe in the RUNBOOK.
   > Also re-checked the same evening: migration **650 applied** (`stored_url`,
   > `stored_url_expires_at` both present), **`DELIVERY_SMTP_*` env present on `agent-chassis`** with
   > `DELIVERY_SMTP_PASS` resolving to secret `delivery-smtp-secrets` (which **exists**, 16 bytes —
   > note the ref is `optional: true`, so a missing secret would be silently empty, not a start
   > failure), and all four agents **exist and are active** (`zip-deliverable-dispatch` and
   > `zip-deliverer` carry status `experimental`). What genuinely remains first-time: the four agents
   > have never run, no token has ever been minted, and DKIM/DMARC at the mail host (absent as of
   > 2026-08-26, and not checkable from here).
5. **Carry `bugs_open/420` §C to the owner WITH the next delivery ask** — a commitment with a
   trigger, agreed with the `bugfix_417_420` lane, in the exact form: **"what CONSENT STATE may a
   classifier write on a contact row?"** Not "may it write one" (it plainly may; 24 of 57 sites
   got legitimate contacts that way). A two-state answer designs the fill-only-if-empty inversion
   back in at row level; the honest third state is probably *recorded, not published, not asked*.
   Hand the answer back to that lane VERBATIM; it is their record.

---

## 2. OWNER RULINGS TODAY — all in his words

- **Items 1–4: "carry on fixing the tools that make these work properly."** Delivery NOT unblocked.
- **The CTA line: "cut it."** Applied. Line 1 is his verbatim text (above).
- **The domain price: he caught a stale figure.** *"I think we reduced the domain pricing"* — right,
  by eight days. See §4.
- **"Guides should be a type of their own."** Routed (after two misroutes, §5) to the session
  named **`428`**, who owns build-site-planner's page-type vocabulary and has accepted it as an
  open item pending their user's sequencing call. **Keep the two halves apart:** adding a `guide`
  type is additive and inert; RE-TYPING the 167 existing pages changes what every blog and guide
  listing resolves on 20 live sites, which is architecture-scope under the 2026-07-29 ruling §1.
- **RFC_058: "Option C would be my preference of the original choices"** — four identities
  (ordering party, operating party, published contact, SUBJECT), plus two additions he made
  himself: a fifth **selling-party** identity, which he explicitly DEFERRED (*"We don't have to do
  the first one today"*), and **multi-valued contacts** (*"there may be more than one contact
  detail for any of these"*), which he did not. ⚠ **The two are NOT independent** (owning lane's
  point, and it is the sharpest thing said about this): a relation is what multi-valued contacts
  force, and a relation is ALSO what makes the deferred fifth identity cheap, because under it a
  new identity is a ROW and under columns it is a migration. **A session that shortcuts the
  contacts change back into columns "for now" silently converts his deferral into future migration
  work.** Consent then attaches per contact, so three states must be distinguishable: has contacts
  none published · has none recorded · asked and the answer is none. System of record: `RFC_058`
  (`b50c6af84`, `70932b173`), owned by the `bugfix_417_420` lane.
- **ALREADY RULED 2026-09-02, and this lane wrongly re-opened both today** (§5): *"palette: the
  cream/off-white STANDS — no flip"* and *"header stays LOGO-ONLY. Closed."*
  (`webdesign_uk_build_service/NOTES…:6904-6910`, "Rulings (owner, via boxingonline thread)").

---

## 3. WHAT CLOSED TODAY, each verified at the artefact

| item | evidence |
|---|---|
| **5 — transparent logo** | served `/assets/images/logo.png` lm 13:14:09Z, PNG **colour type 6**, **80.10% fully transparent**, border ring 99.84%. Source read the same way: 80.82% / 99.91%, fringe 0.038% |
| **14 — card decks** | served `/index.html`: **6 × `article-card__excerpt`**, **0 × "\| Boxing Online"**, 30 `article-card` control |
| **contact 404** | contact.html **404** with index **200** and an invented path **404**; both b2worker sites on `th2:`. `bugs_open/429` CLOSED and moved |
| **GTM + consent** | 6/6 sampled pages carry `GTM-PQ3WCTBD`, 404 control holding |
| **the logo incident, all four sites** | seotools 92.21% · websitepromotion 87.4% · designblog 88.5% · boxingonline 80.10%, every one colour type 6 at the SERVED bytes. `bugs_open/424` CLOSED |

**Also built today: `sweep_site_defects.sh` (this dir)** — the mechanisable checks of
`SITE_DEFECT_CATEGORIES.md` with its §0 disciplines executed rather than remembered. Run it as
`./sweep_site_defects.sh <domain>`; exit 0 clean, 1 findings, and **a blind check always prints and
always makes the exit non-zero**. Current boxingonline result: 8 findings, 0 blind, all owned
elsewhere. ⚠ **Its own first run carried four false positives and one false clean** — every arm now
carries the measured reason beside it; do not "simplify" them back (details in the script header).

---

## 4. FILED TODAY

| number | what |
|---|---|
| **`bugs_open/451`** | the two-strike ladder parks a RECURRING detection as `unresolved` because a COMPLETED refresh counts as a strike — `stale_chrome` 75 of 76 parked across 12 sites, so the GTM/consent rollout could never arrive |
| **`bugs_open/457`** (components lane's; this lane contributed the finder fall-through + census) | `rebuild_blog_listing` appends an orphan row every run and now hard-fails on the duplicate guard. **On this site it serves 36 card links for 6 distinct articles under 6 "Latest Articles" headings.** A re-render clears the EMPTINESS; only deletion removes the DUPLICATION |
| **`bugs_open/466`** | **an approved review hands the editor a payload it cannot read.** First real use of approve-to-apply, on this site: the approval wrote `copy_edit: null` / `page_target: null` with the payload under `approved_data`, and `copy-editor` emits N edits where `section-editor` applies one. The review row reads `complete` while nothing applied — **the owner has no way to learn his decision did not land** |
| **migration `726`** | the delivery email quoted the **superseded** £200 domain buy-out. His 2026-08-26 ruling moved it to **£59.99**; `SQL_2026-08-26e` censused *"every £200 in the live SPECS"* and the email is a step config on an AGENT DEFINITION, outside that population by construction. Applied by hand with a snapshot, an anchor guard, and a verify that also asserts the RENTAL price was undisturbed. **Nothing shipped it — the chain has never run** |
| **migration `725`** | the temporary, owner-authorised `section_shrink_floor` window that let the index rebuild through. **CLOSED**; total exposure across all its openings was ONE other fleet build, whose outcome the lowered floor did not change (kept 0.88 / 0.89) |

---

## 5. THE DAY'S OWN MISTAKES, and the checks that catch them

Six went into `WRONG_CALLS.md` today. The three worth reading before working this lane:

1. **Told the owner a decision was open when he had already ruled it** — and the ruling was in this
   lane's own NOTES under a differently-worded heading. Check: `grep` the rulings ledgers before
   listing ANY decision as outstanding. The designer seat made the SYMMETRICAL error the same hour,
   asserting an ABSENCE from a phrase-grep over one directory. **A ruling is recorded in whatever
   words the recording session chose, so a phrase-grep can only ever disconfirm the phrase.**
2. **Relayed a peer's attribution as fact** ("they shipped migration 687 today") and misrouted the
   owner's guides ruling twice. Check: `git log -- <file>` settles authorship in three seconds.
3. **A new detector's first run measures the detector.** Four false positives and one false clean,
   including a clean **0** on the owner's own email because the needle was empty and `grep -Fc ""`
   returns a tidy zero. Check: for every non-zero print the matched context; for every zero print
   the needle's length and a positive control.

New landmines filed today: **`set -o pipefail` + `grep -q` inverts an if-guard** (a matching
pipeline reports failure via the producer's SIGPIPE — general to every shell script in the estate);
**a runtime-assembled `UPDATE` is invisible to a literal-SQL census** (the 417/420 lane's find,
confirmed here at the lines: a literal census reports 3 writers of `sites.email` where there are 4,
and the invisible one is the unconditional admin PATCH); and an addendum that **`date -u -d '<bare
timestamp>'` parses in LOCAL time** and only renders in UTC.

---

## 6. FALSIFIERS — check before believing this file

A newer handoff in this dir or the boxingonline one · the served state of `/index.html` (the two
approved edits were pending a publish tick when this was written; re-probe, do not quote) ·
`customer_access_tokens` still 0 · a NEWER chassis roll (v1.0.1359 here; re-read the stamp, PER
SERVICE, and pick the negative control by the property it must lack — a commit made AFTER the
stamp, not merely your latest) · whether `bugs_open/457` candidate 4 landed and the 36-card
duplication cleared · whether the owner answered the logo question.

## 7. Read order, cold

This file → `SUMMARY_2026-09-03_site_delivery_and_editor.md` (the read-aloud version) →
`README_where_we_are.md` (his plain-prose log, newest at the bottom) → `RUNBOOK…` (the corrected
delivery recipes, the chrome-refresh-by-hand recipe, and the prepared index rebuild) →
`APPROVAL_READOUT_2026-09-02…` (three columns: verified / not fixed / built-but-inert) →
`SWEEP_2026-09-03_boxingonline_output.txt` → the boxingonline session's 09-03 handoff.
