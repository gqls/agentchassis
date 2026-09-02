# HANDOFF — bugfix 417/420 lane — 2026-09-02, continue here

**One lane, two bugs, both class fixes SHIPPED, APPROVED and LIVE.** Three further bugs and one
RFC were spawned by the work and are with other people. This document is the cold start: what is
done, what is proven, what is still owed, and the decisions that are the owner's.

**Bug files (resolve by SLUG — both numbers are ambiguous):**
- `bugs_open/417_HANDOFF_2026-08-31_planner_logo_exemplar_licenses_a_wordmark_it_never_names_so_the_image_model_invents_a_brand.md`
- `bugs_open/420_HANDOFF_2026-08-31_order_intake_publishes_the_billing_email_as_the_sites_public_contact_and_registers_it_as_a_renderable_claim.md`
  ⚠ **420 is ambiguous** — the other 420 is the negation gate's prose walker. **417 is not ambiguous.**

**Working docs:** `docs/agent_docs/docs024_key_docs_latest/bugfix_417_logo_text_policy/` and
`…/bugfix_420_contact_consent/` (PLAN, RUNBOOK, NOTES, README_where_we_are, SUMMARY in each).

---

## 1. STATE — verified at the artefact on the NEW build, not inferred

Chassis rolled again **2026-09-02 15:39 / 15:53Z** (both replicas). Probed the RUNNING binary with
a control pair — never the `build provenance` log line, which the estate's landmine calls
unreliable:

| needle | result | meaning |
|---|---|---|
| `Render a text-free mark` | PRESENT | 417's policy clause is live |
| `image_kind_conflict` | PRESENT | 417's detectors are live |
| `published_contact` | PRESENT | 420's opt-in is live |
| `email_was_published_contact` | PRESENT | 420's log field is live |
| `resolveLogoIntent` | **absent** | **removed-string control** — round 3 deleted it, so absence proves round 3+, not round 2 |
| `email_was_intake_value` | **absent** | **removed-string control** — 420's commit deleted it |
| `zzz_impossible_control` | absent | the grep can return zero, so the PRESENTs mean something |

**Both fixes survived the new build.** `go build ./platform/...` clean; my files all committed.

---

## 2. WHAT WAS DONE

**417 — the logo text policy.** The estate's "no lettering" rule was attached to the prompt's
SOURCE (a fallback builder only reached when a plan carries no prompt) instead of the asset's
PURPOSE. Every planner-built site supplies a prompt, so the rule governed only the population that
never needed it. Moved to the generation choke point (`GenerateImageAction`), where every prompt
from every producer passes — including work items already queued. Opt-in
`constraints.wordmark_text` permits lettering only when it names the EXACT brand string, validated
against the site's own naming. Council **APPROVED round 4**, `bb099a3d-0555-4fcf-b12a-31652b59f8b9`.
Migration **680** applied (washed the race-tail prompt).

**420 — a billing address is not consent to publish.** `sites.email` carried two contracts; the
first paid build published the payer's address on 13 pages. The payer's address is now
delivery-only; the published contact comes only from an opt-in `published_contact` intake answer;
absent it, the site publishes none. The evidence-register fact that LICENSED re-publication is
minted only from the consented address. No schema change, no backfill. Council **APPROVED**,
`2026df60-b91e-4f1c-8425-a3f6f14e6309`.

---

## 3. WHAT IS PROVEN, AND WHAT IS ONLY EVIDENCE

**Disconfirmation A (did the guard REACH the generation?) — 6 for 6 `[MEASURED 2026-09-02 19:45]`.**
Every logo generation that has COMPLETED since the guard went live carries the clause in
`origin_prompt`: boxingonline.com (10:40Z), advertise.co.uk (14:48Z), designblog.co.uk (17:03Z),
seotools.co.uk (17:10Z), **gamedesign.uk (17:58Z)** and **websitepromotion.co.uk (18:00Z)**. Neither timestamp column can be trusted (§6) — and **nor can the work-item trail on its
own**: boxingonline's regeneration filed no work item at all. Settle it on the **storage key's date
directory**, which caught all three and is sound by construction (§6).

**Disconfirmation C (did the model OBEY it?) — 5 for 5 `[MEASURED 2026-09-02, eye-checked]`.**
Five of the six have been opened and looked at (gamedesign.uk not yet — it is the ONLY one left):
- **boxingonline.com** — fist-in-a-square mark. Zero lettering, single composition. (Delivery
  lane's original check, re-done independently this session from `boxingonline.ugg2.com`.)
- **advertise.co.uk** — broadcast/signal mark, concentric arcs off a mast. Zero lettering, single
  composition, no invented brand name.
- **designblog.co.uk** — geometric star mark, generated 17:03:23Z off the queue. Zero lettering,
  single composition. Carries the clause.
- **seotools.co.uk** — compass/target mark, generated 17:10:10Z off the queue. Zero lettering,
  single composition. Carries the clause.
- **websitepromotion.co.uk** — paper-plane-and-signal mark, generated 18:00:38Z on the second
  attempt (the first was refused by 424's guard). Zero lettering, single composition. Verified at
  BOTH the source object and the deployed file.

**421's two-panel design-comp shape did not recur on any of them.**

**Disconfirmation D (not over-applied) — holding.** A post-roll `content_hero` carries no policy
sentinel; heroes are untouched.

**The detectors have never fired** (`image_generation_without_kind`, `image_kind_conflict`,
`logo_wordmark_rejected` all zero) against a demand control of **702 `agent_error_log` rows in 24h**
and **0 `unattributed`** rows. Read as: every generation so far supplied a resolvable kind.
**Evidence the legacy-parent path is dormant — NOT proof those callers are dead.**

---

## 4. WHAT IS LEFT ON THIS LANE

Ordered. Nothing here blocks moving to other bugs; items 1–2 are cheap and time-sensitive.

1. ~~**Eye-check advertise.co.uk's logo.**~~ **DONE 2026-09-02** — clean, see §3. The RUNBOOK now
   carries the fetch recipe (§"Fetch a generated asset's BYTES and LOOK at it"); it was not a
   2-minute job, because the customer's own domain is not ours and the bytes had to come out of the
   bucket through a pod.
2. ~~**The three queued logos.**~~ **ALL THREE DONE, all clean** — designblog 17:03Z, seotools
   17:10Z, websitepromotion 18:00Z (second attempt; the first was correctly refused by 424's guard
   at `border_keyed=0`). `gamedesign.uk` (17:58Z) also generated and carries the clause but has
   **not been eye-checked** — that is the one cheap action left on this lane.
   **A 424 refusal is not a 417 signal** and a veiled background is 424's defect, not 417's.
2b. **NEW, and not this lane's: 5 of 34 sites with a logo asset still render a TEXT header.**
   `site_components.slot_name='header'` holds `logo-text` and no `logo-img` on websitepromotion.co.uk,
   webdesign.co.uk, ai-agent-orchestration.com, loanandmortgagecalculator.co.uk, cookly.uk — while
   29 of 34 render the image correctly, so the mechanism works in general. **At least two distinct
   causes:** websitepromotion's header rendered 17:30Z with `render_inputs.plan_logo = ""`, half an
   hour BEFORE its logo existed, and the 18:01Z page re-render re-used the stored header (a page
   rebuild does not rebuild the header) — whereas **webdesign.co.uk's header rendered AFTER its logo
   with a real `plan_logo` digest and still emitted text**, which is a different defect. Not
   diagnosed, not filed, needs an owner — the header/chrome pipeline, not 417.
3. **The fence decision itself** — recorded in 417 with its grounds and trigger, deliberately NOT
   taken on n=1. See §5.
4. **417 and 420 stay OPEN** until their classes are bounded, per the fixed-AND-live bar. 420 also
   stays open on its §C residual (owner decision, §5).
5. **Not this lane's, but spawned by it and unowned:** `bugs_open/421` (the multi-panel design comp
   — still no owner).

---

## 5. THE DECISIONS THAT ARE THE OWNER'S

**These are the only things genuinely waiting on a person. Everything else has a next action.**

### Decision 1 — the identity model (RFC_058). The big one.
`docs/agent_docs/docs024_key_docs_latest/architecture_review/RFC_058_the_platform_stores_one_identity_where_the_business_has_at_least_three.md`

Your ruling: *"the identity of the person creating the site can be independent of the operation of
the site (they might be the design agency) … we will need to replumb the identities and think hard
about what identities we need to store."* The RFC proposes **no schema**, on purpose — it names the
candidates, measures what exists, and puts the choice:

- **A — two identities** (counterparty, published contact). Smallest change; **fails your own
  agency example.**
- **B — three** (ordering party, operating party, published contact). Models the agency case.
  ⚠ Must state what happens when the operating party is unknown, and *"fall back to the orderer"*
  is the answer that makes it pointless.
- **C — four** (B plus the *subject*, the business the site is about). The register genuinely needs
  this: `business_name` is currently seeded from the ordering party, so a fan brand's register
  carried a design consultancy's identity.

Lane recommendation, offered as input: **B as the contract, with the subject derived and owned by
the register** rather than a fourth intake field.

**The binding constraint, whichever you choose:** a two-state field cannot express this. *"We asked
and the answer is none"* must be distinguishable from *"not yet known"*, or every fill-only-if-empty
writer in the estate silently overturns a customer's decision. That is not theoretical — it is what
inverted into a refill vector during the incident.

### Decision 2 — the 420 §C residual: does the narrow ruling extend to derived contacts?
A contact the **classifier derives** (from brief prose, or research) can still reach the published
column with nobody asked — `sync_site_identity` reads `identity.contact.email` and writes
`sites.email`. **28 current specs carry a contact email.**

- **Narrow reading** (both lanes would argue for it): a derived contact is not explicit consent, so
  that writer should require the consent key too.
- **Wide reading:** it re-plumbs how the **24 of 57** sites with legitimate contacts get them.

Neither lane will patch this unilaterally — it is an estate-level call, and it is arguably
subsumed by Decision 1.

### Decision 3 — ordering cannot reopen until the intake chat asks the contact question.
Consequence of your own ruling, and it is **box-side, in your environment** — no session here can
close it. Until the chat asks *"what contact details should the site show?"*, **every new customer
site ships with no contact at all**, including customers who wanted one. This is the gating item
before order 2.

### Decision 4 — `bugs_open/421` has no owner.
The boxingonline logo was also a **two-panel design comp** — the mark twice, on two grounds. Filed,
diagnosed, fix candidates ranked (cheapest is a dimensional envelope at store time, no classifier
needed). It needs routing to a lane.

---

## 6. TRAPS THIS LANE HIT — read before touching assets or censusing generations

- **`assets` UPSERTS on regeneration**: `created_at` keeps the ORIGINAL generation's date. A census
  filtered on `created_at > <roll>` returned ZERO logo generations on a day one had completed, and
  I reported that stale zero to two lanes. **LANDMINES entry.**
- **…and `updated_at` lies the other way** — ANY write bumps it. **Measured 2026-09-02: of 10 logo
  rows with `updated_at` today, only 2 were real regenerations**; the other 8 carried storage keys
  dated 06-21 to 08-26. **Neither timestamp is a regeneration signal.**
- **…and the WORK-ITEM TRAIL is incomplete too** — this handoff used to send you there. boxingonline's
  10:40Z regeneration filed **no `needs_imagery` row at all** (every item type on that site since
  09-01 checked). A trail-keyed census reports one regeneration on a day there were two.
- **The instrument that caught all three, sound BY CONSTRUCTION:** the storage key's date directory.
  `dynamic_adapter.go:717` builds every key as `images/<client>/<YYYYMMDD>/<fresh uuid>.png`, so a
  regeneration can never re-use an old key:
  ```sql
  substring(a.storage_path from 'images/[^/]+/([0-9]{8})/') AS key_date
  ```
  Blind spot: operator-supplied rows (`amend-asset.sh`) have no dated generator key and return NULL.
- **⚠ The `.png` in that key is a LIE — 12 of 12 sampled logo objects are JPEG.** Read the magic
  bytes before you trust the format; a JPEG cannot hold alpha at all. LANDMINES entry; `bugs_open/433`.
- **Transparency cannot be verified by looking** — viewers draw a checkerboard FOR real alpha, so
  paint and real alpha are visually identical. Only a PNG chunk scan settles it (`bugs_open/424`).
- **Two producers' clauses differed by one capital letter** — my Go clause and migration 680's wash
  both contain "text-free mark". Pick a needle unique to the producer you mean.
- **`doc_notes.subject_type` is CHECK-constrained to eight values** and `'site'` is not one. A
  best-effort writer fails SILENTLY. **LANDMINES entry.** Reuse `LogActionEntry`/`agenterrors`.

---

## 7. SPAWNED, AND WITH OTHERS

| item | state |
|---|---|
| `bugs_open/421` multi-panel design comp | **filed, UNOWNED** |
| `bugs_open/424` transparency capability gap | filed by me, implemented by the 424 lane (`6440ec968`), council APPROVED — **but its matte ran for the first time 2026-09-02 17:03Z and FAILED: zero transparent pixels, and its fail-closed guard scored the run 1.000.** Structural (the guard counts flood *reachability*, not transparency), so it survives any threshold change. Contributed to their lane + bug file; NOT taken over. `CONTRIB_2026-09-02_from_417_lane_…` |
| `bugs_open/433` empty `mime_type` (910/1,277) | filed by another lane, diagnosis owed |
| `RFC_058` identity model | **DRAFT, awaiting the owner** |
| boxingonline pre-delivery | delivery lane; interim solid-ground logo shipped |

---

## 8. IF YOU READ ONE THING

**The census and the eye are different instruments and neither substitutes for the other.** The
census proves the instruction ARRIVED; only a human looking at the PNG proves it was OBEYED — and
`bugs_closed/390` measured that co-present prompt instructions are adjudicated by the model, not by
precedence language, which is exactly what this fix relies on. Both known instances of the original
bug were found by a person opening an image. **Nothing automated in this estate can currently see a
logo that says the wrong words.**
