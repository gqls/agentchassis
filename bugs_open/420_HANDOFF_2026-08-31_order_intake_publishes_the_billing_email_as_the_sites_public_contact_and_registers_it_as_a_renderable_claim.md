# 420 — order intake publishes the BILLING email as the site's public contact, and registers it as a renderable claim — two contracts on one column, plus a claims-register licence to re-publish it

**Filed 2026-08-31** by the delivery-lane session, from the incident on the FIRST PAID
build (boxingonline.com, order BR-9AUZ59, site `d2aa5206-…`). Incident measurements by
the boxingonline critique session (their OWNER_REVIEW_2026-08-31 §0, in
`docs/agent_docs/docs024_key_docs_latest/site_delivery_and_editor/`); the defect
framing and the second half below are this lane's. **The incident is being remediated;
this file is about the class**, which fires again on order 2.

## What happened (all measured 2026-08-31)

The customer paid with `aaa@designconsultancy.co.uk` (the owner, acting as customer
zero — so the leaked address was HIS, which is the only reason this was cheap). P5's
`seedCustomerIdentity` (seed_build_queue_action.go, council-approved trail 7e3dd082)
wrote it to `sites.email` at build release, by design: sites.email is the canonical
identity store used to DELIVER (651's delivery-email-sender reads it). The site chrome
then assembled the PUBLIC footer Contact block from the same column — rendered as
`<p><a href="mailto:…">…</a></p>` inside `div.footer-contact` on **every deployed page**
(13/13), plus 4 further occurrences on /contact.html via contact components. The owner
saw his personal address published on a public site and ordered it off immediately.

## The defect, half 1 — two contracts on one column

`sites.email` means BOTH "where the platform reaches the customer" (billing/delivery
contact; what the delivery email gate reads) AND "the contact address the site
publishes" (what footer chrome and contact components assemble from). Nothing in the
schema, the seed, or the chrome distinguishes them. For a real order 2, the platform
will publish whatever address the customer happened to PAY with — a billing address is
not consent to publish.

## The defect, half 2 — the register licenses re-publication (the subtler one)

`seedCustomerIdentity` also minted an evidence_base FACT whose claim text was
**"Enquiries reach aaa@designconsultancy.co.uk."** (id 'contact', kind 'entity',
customer_attested). That is a RENDERABLE registered claim: section planning assigns
fact ids to sections, writers write from facts, and validate_page_content would
rightly PASS a page publishing it. Deleting the address from `sites.email` and the
rendered pages therefore does NOT make it stay gone — any rebuild consuming the
register could legitimately re-emit it, validated clean. The intake wrote the same
address into the `briefing` spec (`contact.contact_email`) too — a second regeneration
source. **An address you must not publish cannot be represented as a publishable fact
anywhere in the spec stack.**

## Incident remediation (done 2026-08-31, this site only — not the class fix)

- boxingonline session: sites.email→NULL, contact components scrubbed/rewritten,
  whole-site rerender fired (corr 3f604312), 0 component/page rows carry the string.
- this session: evidence_base superseded (contact fact REMOVED, business_name kept);
  briefing superseded (`contact.contact_email` nulled, rest verbatim). Verified:
  **0 current site_specs rows carry the address.** identity spec contact block was
  already all-null, so the fill-only-if-empty column syncs (sync_site_identity,
  update_site_content) have nothing to refill from — checked, not assumed.
- served-page verification pending the rerender wave + publish re-mirror (this
  session, in flight at filing time).

## ADDENDUM 2026-08-31 (evening) — the address lived in a FOURTH place, and the removal READ as complete while it was not

Found by the boxingonline session after three clean sweeps still left the served
pages carrying the address. **`site_components` (slot_name='footer') held the mailto
in its `rendered_html`** — `pages.rendered_footer` is NULL site-wide, so the deploy
assembles the footer FROM THAT ROW. The earlier "clean" verifications queried
`sites`, `page_components`, `pages` and (this session) `site_specs` — four sweeps,
each internally correct, none of which looked at `site_components`.

**The sharper defect: `rerender-pages` with `refresh_site_components:true`
refreshed the `head` and `header` slots and SKIPPED `footer`** (measured: head +
header updated 15:39:06 by the rerender wave; footer still at its 13:31:54 bake).
So a full-site rerender fired specifically to flush removed data reported success
while re-serving the removed data — a data-removal that reads as done. Whether
footer's exclusion is a bug or an undocumented exclusion is UNSETTLED; either way
the flag's name promises all slots. This half may deserve its own file if the fix
splits from 420's contract work.

Incident tail: the boxingonline session surgically removed the whole
`footer-contact` block from the footer row (guarded transaction, verified pre-COMMIT)
and re-fired the whole-site rerender (corr `eef4de19`). **Pre-delivery obligation
(this lane's list): the footer row is now a hand-patched artefact — rebuild it from
content_data through the normal component-render path before the customer handover.**

For the class fix, add to the verify recipe: the served-page probe (not any set of
DB sweeps) is the only complete check, because the set of places a value is baked
into is not enumerable from the schema — four independent sweeps each missed one.

Two more instances of the same family from the same evening (both caught by the
boxingonline session): **a chrome-only change is invisible to the page content
hash**, so a whole-site rerender no-ops on every page and reports success — a
change complete at its source, invisible to the thing whose job is to notice
changes (targeted per-page rerenders are the working path); and **a removal sweep
must enumerate its pages from `pages WHERE deployed`, never from the pages someone
remembered to probe** — a one-page watcher reported CLEARED at 16:19 while six of
nineteen deployed pages still served the address. Every false "clean" tonight was
an incomplete enumeration reading as a result.

Late additions, same evening: **the lag-vs-upstream discriminator** (boxingonline
session) — a still-dirty served page has two opposite diagnoses that grep
identically: compare the served object's last-modified against the page's
deployed_at; **older means wait (publish lag), newer means look upstream (a dirty
source was freshly published)**. And the footer's tail: the hand-patched footer
row turns out to be UNREGENERABLE — two forced render_site_components runs
declined the footer slot silently with every surfaced reason field empty
(rendered.footer=false; no lock, no ineligible, no error) — so the hand-patch
serves for a reason no regeneration has ever tested. Diagnosis run
`387c0a2d-7fd7-460c-b7cf-fb46ff50b13f` owns the mechanism; do not quote one
until it reports.

## Owner rulings recorded (2026-08-31, relayed via the boxingonline session)

1. "There should be no contact email or address on this site because I didn't ask for
   one" — for THIS site, sites.email/phone/contact_address stay NULL, and nothing may
   re-populate them.
2. The brief never mentioned contact at all — the contact page itself was
   planner-invented (its form posts to '#contact', i.e. nowhere; delete-vs-wire is
   with the owner as a separate decision).

## Fix candidates (class fix — needs the owner's contract ruling + council for the code)

Ranked by what makes the bad state unrepresentable:
1. **Separate the contracts at the schema/seed level**: the intake email lands in a
   delivery-contact field (or stays only on billing_orders), and `sites.email` (the
   published contact) is written ONLY from an explicit customer answer "what contact
   details should the site show?" — absent that answer, the site publishes none. The
   chrome's source column then never holds an unconsented address.
2. **Stop minting the 'contact' evidence fact from the ordering email** — a billing
   address is attested for billing, not as a publishable business claim. If a contact
   fact is wanted, it comes from the same explicit intake answer as (1).
3. Weakest: chrome suppresses the footer contact block unless a published-contact
   flag is set — leaves the two-contract column in place, guards one consumer.

## How to verify a fix

Run an order whose payer email differs from any contact the brief asks to publish;
assert the built site serves ZERO occurrences of the payer email (all pages), the
evidence_base contains no fact rendering it, and the delivery email still reaches the
payer address. The probe-with-controls recipe is in this incident's NOTES entries.

---

## CLASS FIX SHIPPED — bugfix 417/420 lane, 2026-08-31 (taken with the filing lane's agreement)

**Status: committed (`162877051`), INERT until the next chassis roll — so this file stays OPEN.**
Council: `Council-Submitted: 2026df60-b91e-4f1c-8425-a3f6f14e6309`; verdict NOT yet read.
The filing lane keeps the incident, the boxingonline pre-delivery list, and the ADDENDUM's
`refresh_site_components` footer-skip half, which splits into its own file as this file predicted.

### 1. The owner's contract ruling — the thing this file said was missing

Put to the owner directly, 2026-08-31, as the CLASS question (the existing ruling was per-site):

> **When a customer pays but is never explicitly asked what contact details the site should
> show, the site publishes NONE. The payer's address is used only to deliver.**

That generalises the boxingonline ruling and is the safe side, so the code is built to it and the
architecture does not change if the owner later widens it — only a value does.

### 2. Half 2 was not hypothetical — it FIRED on order 1, established by elimination

This file called the register licence "the subtler one". It is also the one that already
happened. `[MEASURED 2026-08-31]`: the build-briefing agent reads specs with `aspect: "all"`
(`050_build_briefing_agent.sql:57-66`), and the briefing spec it wrote at 12:33:38 carried
`contact.contact_email = <payer>` — while the identity spec's contact block was **all null** and
the customer's brief prose did **not** contain the address
(`direction->>'objective' ILIKE '%designconsultancy%'` → false, 855 chars). **The seeded register
fact was the only possible source.** A registered claim propagated into a second regeneration
source inside twelve minutes of the same build. That is the mechanism this file predicted,
observed.

### 3. The fix, and the two corrections the census forced

**Fix candidates 1 and 2 of this file, both taken.** `sites.email` is now **only** the published
contact. The payer's address is delivery-only and stays in `build_queue.direction.customer_email`
— where the collector already put it and where the delivery dispatch already takes it from. The
published contact comes only from a new OPTIONAL `direction.published_contact` key (the
2026-08-02 shape: opt-in, unsafe side OFF, visible in the producer's own direction literal rather
than licensed by a comment). The `contact` evidence fact is minted only from that consented
address. `business_name` is kept — it is **constitutive** (no page can exist without naming the
business, so giving it to a site-building order is inseparable from consenting to publish it),
where the email is **severable** and was never asked about. That is the same line the incident's
own owner-reviewed remedy drew.

**No schema change and no backfill.** The delivery-contact store already exists; adding a
`sites.delivery_email` column would put a second contact-ish column on the table every render
path loads, guarded by a comment — which the 2026-08-02 ruling says does not survive this tree.
Verified rather than assumed: the seed ran exactly once fleet-wide and 0 current `site_specs`
rows carry the address.

Two things in this file's framing turned out to be wrong, and both make the fix cheaper:

- **"651's delivery-email-sender reads `sites.email`" is CONVENTION, NOT CODE.** `[MEASURED
  2026-08-31]` No code in `send_delivery_email_action.go` or `platform/delivery/` reads the
  column; `customer_email` is REQUIRED `input_data` supplied at dispatch. The split is therefore
  a recipe update, not a chain change.
- **The footer block was ALREADY gated** on `ctx.Email != ""` (`component_library.go:1988`, the
  bugs_open/111 gate). The defect was never a missing gate — it was the **value** the gate let
  through.

**Census, dated as the rule requires: 4 writers / 14 readers of `sites.email` as of 2026-08-31**
— including an **admin PATCH writer this file had missed** (`site_admin_handlers.go:363-367`,
unconditional, the deliberate operator override). The two spec-copying writers
(`sync_site_identity`, `update_site_content`) need no edit: on a customer build they can only
copy what the seed put there, so closing the seed closes them at their source.

**Also removed: the `info@<domain>` synthesis** on two render paths
(`section_editor_actions.go`, `multipage_actions.go`). Post-fix an empty column is the CORRECT
and common state, and fabricating an address nobody owns and no mailbox answers would make
"the site publishes no contact" quietly false. Flagged severable in the submission; blast radius
is 4 estate sites losing a fake display string at their next rerender.

### 4. A guard obtained for free, and a live trap until the roll

**Free guard:** `validate_page_content` loads `sites.email` as the one licensed "official"
address and passes any page publishing it. Post-fix the column no longer holds the payer
address, so **any residual occurrence on a page is now FLAGGED rather than licensed** — this
file's fix candidate 3, without the flag.

**⚠ LIVE TRAP until the chassis rolls (LANDMINES entry added).** boxingonline's `build_queue`
row still holds `direction.customer_email`, and re-seeding is the canonical build retry
(bugs_open/326). Because `sites.email` is now legitimately EMPTY, the fill-only-if-empty guard
**inverts into a refill mechanism**: a retry would put the address back. The register survives
(`WHERE NOT EXISTS`); the column does not. **Do not re-seed that site before the roll.** Every
sweep you would run to check reads clean, because they all describe the state after the last
seed and say nothing about the next one.

### 5. This file's verification lesson, built into the plan

The demanded recipe is an order-2 rehearsal, and it is written to be immune to all three
enumeration traps recorded here: pages enumerated from `pages WHERE deployed` (never a
remembered list), served-bytes probed with a demand control (never a DB sweep as the verdict —
four sweeps read clean while `site_components.rendered_html` still served the address), and no
step treating a rerender status or content hash as proof (a chrome-only change no-ops the hash).

### 6. Open, for the owner, and out of scope here

- `[UNVERIFIED — lives in the box repo]` whether the intake chat's `customer_name` guarantees a
  BUSINESS name rather than a personal one. If personal, the same publish-without-consent shape
  exists for `company_name` at lower severity, since `business_name` is minted as a fact and
  rendered. Flagged rather than fixed, because the answer is not in this repo.
- The `published_contact` wire shape (`{email, phone, …}`) needs agreeing with the intake-chat
  lane. The platform half reads only `.email` today and ignores unknown keys, so the shape can
  widen without another platform change.

---

## COUNCIL APPROVED (2026-08-31) — and its objections found a RESIDUAL the fix does not close

Verdict on `2026df60-b91e-4f1c-8425-a3f6f14e6309`: **approved with 3 advisory objections, none
high-severity.** The commit carries `Council-Submitted:`, so 098 credits it automatically. Read,
and acted on — two objections were fair hits and one exposed a real gap. **This section exists so
the approval is not read as "the class is closed".**

### A — my "4 estate sites" blast-radius figure was wrong, and unverifiable as stated

`debug_historian` objected that no query was shown for it. Running it: my first attempt used
`WHERE status='active'`, which returned **0 with an email and 0 without** — `sites.status` has no
`'active'` value (it is `deployed | pool | system | test`). Corrected, with the denominator so it
is self-checking:

```sql
SELECT count(*) FROM sites;                                    -- 54
SELECT count(*) FROM sites WHERE COALESCE(email,'')='';        -- 34
SELECT count(*) FROM sites WHERE COALESCE(email,'')<>'';       -- 20
```

`[MEASURED 2026-08-31]` **54 sites; 34 with an empty email column; 20 with one.** So removing the
`info@<domain>` synthesis touches a population of up to **34, not 4** — an order of magnitude out.
Logged in `WRONG_CALLS.md`. (Not all 34 would have shown `info@`: some carry a contact email in
their specs which the sync writer fills from — see C.)

### B — the "0 current site_specs carry the address" claim, now with its query AND a demand control

`editquality` and `debug_historian` both flagged this as asserted without visible mechanism, on a
jsonb column where string-absence claims are a known trap. Re-run:

```sql
SELECT count(*) FROM site_specs WHERE is_current AND data::text ILIKE '%<the address>%';   -- 0
SELECT count(*) FROM site_specs WHERE is_current AND data::text ILIKE '%boxingonline%';    -- 2  (DEMAND CONTROL)
SELECT count(*) FROM site_specs WHERE created_by='seed_build_queue';                       -- 1  (the seed ran ONCE, fleet-wide)
```

**The claim holds and the zero is now a measurement**, because the same query shape against the
same column finds something that IS there. Without the control it was a zero that could not have
come out otherwise.

### C — ⚠ THE RESIDUAL: `sync_site_identity` can still publish an address nobody consented to

`editquality` objected that my reason for NOT editing the two spec-copying writers — *"on a
customer build those stores receive the address only via the seed's own writes"* — was an
assumption about internal data flow, not a verified fact.

**It was, and it is FALSE.** `[MEASURED 2026-08-31]` **28 current `identity`/`briefing` specs
carry a contact email**, and the sync writer reads exactly those keys
(`sync_site_identity_action.go:104-137`: `identity.contact.email`, then `briefing.contact_email`)
and writes `sites.email` — the published-contact column. Example from the live data:
`cv1.co.uk` has `identity.contact.email = hello@cv1.co.uk` while its `sites.email` is EMPTY, i.e.
the fill is pending, not absent.

**So the fix closes the path the bug FILED — the payer's billing address — and does NOT close the
general one.** If a customer's brief prose mentions an email, or a research step derives one, the
classifier can land it in `identity.contact.email`, and it becomes the site's published contact
with no `published_contact` answer anywhere. That is the same defect one layer along: **an address
reaching the published-contact column without anyone having been asked.**

**Deliberately NOT fixed here, and this is a scope judgement the owner should confirm rather than
a session's call.** Gating that writer would change how the 20 estate sites with legitimate
contacts get them, and those flows are how estate sites are supposed to work. The narrow reading
of the owner's 2026-08-31 ruling ("absent an explicit answer, publish none") appears to cover this
case too; the wide one would re-plumb the estate's own identity pipeline. **Routed to the owner as
an open decision, stated rather than left implicit.**

### D — the guardian's product point, which is the ruling's known cost made concrete

*"the default publish-nothing behaviour ships fleet-wide the moment this rolls, ahead of the box
intake-chat sending `published_contact`… a customer-facing regression across the whole
order-intake pipeline until an out-of-repo change ships."* Correct, and it is the direct
consequence of the ruling rather than a surprise: **until the intake chat asks the question, every
new customer site publishes no contact at all**, including customers who would have wanted one.
The alternative is publishing an address nobody consented to, which is this bug. Flagged to the
owner; the fix is the chat change, and it is the critical-path item for order 2.

Also accepted, unfixed and recorded: the admin PATCH writer remains an unconditional,
unannotated route to the published-contact column (no way to distinguish operator-set from
customer-consented), and the two `info@` deletions are duplicated rather than shared, which is a
drift surface the next editor of one path will meet.

**C-bis — boxingonline is already defended against this residual, and that is not luck.**
Raised by the delivery lane, verified here first-hand rather than taken on report `[MEASURED
2026-08-31]`: on boxingonline's CURRENT specs, `identity.data->'contact'->>'email'` and
`briefing.data->>'contact_email'` are **both null** — and those are precisely the two expressions
`sync_site_identity_action.go:104-137` reads. Demand control: the same two expressions return a
non-null value on **28** current specs fleet-wide, so the null is a measurement, not a broken
path. `sites.email/phone/contact_address` all remain empty.

So the pending-fill state `cv1.co.uk` sits in — spec populated, column empty, waiting for the sync
to publish it — **cannot arise on boxingonline.** The incident's briefing scrub was therefore
load-bearing rather than belt-and-braces: had only the register and the column been cleaned, the
briefing key would have survived as a live re-publication route through a writer nobody had
identified as one at the time. **The scrub defended against a mechanism that had not yet been
found.**

Ownership confirmed with the delivery lane: the residual stays here and on the owner's decision
list. It is the identity-sync seam (estate pipeline), not the delivery chain, and gating that
writer reshapes how 20 estate sites legitimately get their contacts — an owner call about the
estate, not a patch either lane should make. Both lanes would ARGUE for the narrow reading
(a classifier-derived contact is not explicit consent, so the sync writer should require the
consent key too); arguing it is as far as either will go unilaterally.

**Delivery-side items now closed by that lane** (verified at HEAD, commit `6eea185e6`): 651's
header and the delivery RUNBOOK carry the corrected dispatch source
(`SELECT direction->>'customer_email' FROM build_queue WHERE domain=$1`, never `sites.email`, with
the convention-not-code note), plus the do-not-re-seed block. LANDMINES carries it from this side,
the RUNBOOK from theirs.

**Order-2 critical path, per the guardian objection:** the box intake chat must ask "what contact
details should the site show?" before ordering reopens, or every customer site ships contactless.
That change is box-side (owner-run env, webdesign lane), so it needs the owner's hands or a
briefed session — it is not a platform task and no session here can close it.
