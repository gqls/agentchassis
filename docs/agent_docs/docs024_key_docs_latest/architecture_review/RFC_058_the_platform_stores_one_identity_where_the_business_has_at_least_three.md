# RFC 058 — the platform stores ONE identity where the business has at least three, and the published contact is the one nobody agreed to

**Status: DRAFT, raised 2026-09-02 by the `bugfix_417_420` lane, on an explicit OWNER RULING.**
Raised out of `bugs_open/420` §C, whose class fix is APPROVED and LIVE — and which closed the
*filed* defect while leaving the general one open. This RFC is the general one.

**The owner's words, 2026-09-02 (relayed via the boxingonline session from his review thread):**

> *"The identity of the person creating the site can be independent of the operation of the site
> (they might be the design agency) so the contact details need to be independent of the site
> build. We will need to replumb the identities and think hard about what identities we need to
> store."*

This RFC **proposes no schema**. The owner asked for thinking about *what identities we need*, so
its job is to name them, measure what exists, and put the decision — not to pre-empt it with a
migration.

---

## 1. What the thing IS, in plain terms

When a website is built, several different people or organisations are involved, and the platform
currently records them as if they were one.

- **The ordering party** pays. We bill them, we deliver the finished site to them, and we email
  them when it is ready. On the first paid build this was the owner himself, acting as customer
  zero.
- **The operating party** runs the site afterwards. They receive the handover and they are who we
  would contact about the site itself. **They may be a different organisation entirely** — the
  owner's example is a design agency ordering on behalf of its client.
- **The published contact** is what the finished website shows the public. It may be neither of the
  above — a support desk, a sales alias — and it may legitimately be **nothing at all**.
- **The subject** is the business the site is *about*, whose claims the evidence register holds.
  On boxingonline this is a fan brand with no legal entity; the ordering party is a design
  consultancy. They are not the same thing and the register currently cannot tell.

**The rule this RFC is built on:** an identity collected for one purpose is licensed for that
purpose only, and each of the four above has a different consent and a different consumer.

**How the current case measures against it:** the platform has one column, `sites.email`, and
until 2026-08-31 it meant all four. The first paid build therefore published the payer's personal
address on 13 pages. `bugs_open/420` split *billing* from *published*. **It did not split
ordering from operating, and it did not give the subject an identity of its own** — which is
exactly the gap the owner has now named.

---

## 2. What exists today `[MEASURED 2026-09-02]`

Five stores hold identity-shaped data, and no two agree on what they mean:

| store | what it holds | rows |
|---|---|---|
| `clients` (id, name, email, phone, tier, customer_status) | a billing-ish party | **3 rows**, **1** with an email |
| `sites.email / phone / contact_address / company_name / logo_text` | the PUBLISHED identity (post-420) | 54 sites; **34** with an empty email |
| `build_queue.direction.customer_email / customer_name / order_reference` | the ORDERING party, from intake | **3** queue rows; **1** carries them |
| `site_specs` `identity` / `briefing` | spec-layer identity, written by the classifier | **28** current specs carry a contact email |
| `site_specs` `evidence_base` facts | the SUBJECT's attested claims | seeded once, fleet-wide |

**The linkage that would connect a site to its client barely exists**: `sites.network_id` is set on
**33 of 54** sites, but `networks.client_id` is set on **1** network. So there is a path from a
site to a client on paper, and it is populated for approximately one site. Any design that assumes
`sites → networks → clients` works today is assuming something that is 1/33 true.

**Nothing anywhere records an operating party.** It is not under-populated; it does not exist.

---

## 3. Why this is architecture-scope and not another opt-in field

`bugs_open/420` shipped `published_contact` as an opt-in direction key with the unsafe side OFF,
under the 2026-08-02 ruling. That was right for what it was: **one** new authority on **one** seam,
inert until a chat asks for it.

This is not that shape, for three reasons:

1. **It changes what the platform GUARANTEES about a site**, which is the 2026-07-29 trigger test.
   "Whose site is this?" currently has one answer and would come to have several, each with its own
   consumer. Every reader of `sites.email`, `sites.company_name` and the evidence register inherits
   a new question: *which* identity did you mean?
2. **RFC_022's opt-in exception explicitly does not apply** — it requires zero live consumers, and
   an identity model has consumers on day one (delivery, chrome, the register, the classifier, the
   admin UI).
3. **A tenth opt-in field is the accumulation RFC_022 exists to catch.** Adding
   `operating_contact` beside `published_contact` beside `customer_email` is how a shared action
   acquires a surface nobody understands. The owner's own phrasing — *"replumb"* — is a rejection
   of the incremental route.

---

## 4. The decision to put, and the options

**The question for the owner is not "how many columns" — it is which identities the business
actually needs to distinguish, because each one we store is one we must then keep true.**

### Option A — two identities: *counterparty* and *published contact*

Collapse ordering and operating into one "counterparty" (whoever we deal with), keep the published
contact separate as 420 already does.

- **For:** smallest change; 420's split is already live and this only names it properly. Matches
  every order to date, where orderer and operator are the same.
- **Against:** it fails the owner's own example. A design agency that orders ten sites for ten
  clients has one counterparty and ten operators, and handover, support and renewal all belong to
  the operator, not the agency.

### Option B — three identities: *ordering party*, *operating party*, *published contact*

The owner's sentence, taken literally.

- **For:** models the agency case correctly; makes "who do we email at handover?" a different
  question from "who paid?", which is the question the delivery chain actually needs to ask.
- **Against:** the operating party is frequently unknown at build time and may never be supplied.
  A field that is empty on 90% of rows invites readers to fall back to the ordering party — which
  silently rebuilds the conflation this RFC exists to remove. **Any Option B design must state
  what happens when the operating party is absent, and "fall back to the orderer" is the answer
  that makes it pointless.**

### Option C — three identities plus the *subject* (four)

Adds the business the site is *about*, distinct from anyone we transact with.

- **For:** the evidence register genuinely needs this. `business_name` is currently seeded from the
  ordering party's name, which on boxingonline meant a fan brand's register carrying a design
  consultancy's identity. The register's whole job is to hold claims about the *subject*.
- **Against:** the most work, and the subject is often only discoverable after research, so it
  cannot be an intake field. It may be better modelled as *derived and revisable* rather than
  *supplied*.

**This lane's recommendation, offered as input and not as a decision: B as the contract, with the
subject (C) treated as a derived identity the register owns rather than a fourth intake field.**
The reasoning is that ordering and operating are both *transactional* parties we are told about,
while the subject is a *research finding* we establish — and giving them the same shape would
invite the same fallback confusion the register already suffered.

---

## 5. What any option must satisfy, whichever is chosen

1. **No identity may reach a published surface without an explicit, recorded consent** — the
   2026-08-31 class ruling, which this RFC must not weaken. Absent an answer, the site publishes
   none.
2. **Absence must not fall back to another identity, and "deliberately absent" must be
   REPRESENTABLE — distinct from "not yet known".** This is the trap that makes an identity model
   decorative, and it has now been proved twice at two different layers:
   - at the COLUMN layer, `bugs_open/420` §C is live proof — `sync_site_identity` reads a
     classifier-derived contact into the published column today, with nobody asked, precisely
     because an empty field invited a fill;
   - at the SPEC layer (contributed by the `site_delivery_and_editor` lane from the boxingonline
     incident, 2026-09-02): **the fill-only-if-empty guards INVERTED into a refill vector the
     moment emptiness became deliberate.** Those guards exist to protect an operator's correction,
     and they are correct while "empty" means "not yet known". The owner's ruling made empty mean
     "we asked and the answer is none" — and every one of those guards then read it as a hole to
     fill.
   So a two-state field (set / empty) cannot express this model. Whatever wins must distinguish
   **not yet known** from **deliberately none**, or every fill-only-if-empty writer in the estate
   silently overturns a customer's decision. This is the sharpest single constraint in this RFC.
3. **The evidence register must be able to say whose claim it holds.** A `business_name` fact
   seeded from the ordering party is a claim about the wrong entity.
4. **Every reader moves with the write.** `bugs_open/420`'s census — **4 writers and 14 readers of
   `sites.email` as of 2026-08-31**, including an admin PATCH endpoint that writes it
   unconditionally — is the shape of the work, and that census must be re-run and re-dated before
   any migration, because a census goes stale by addition.

   > **PARTIAL REFRESH 2026-09-03** (prompted by the `site_delivery_and_editor` lane, on the
   > grounds that the owner is engaging with this RFC now and §5.4's own instruction makes the
   > census the first thing to go stale). **Writers are still FOUR, no growth in three days:**
   > `sync_site_identity_action.go`, `seed_build_queue_action.go`, `v3_site_actions.go`, and
   > `internal/core-manager/admin/site_admin_handlers.go` (the unconditional admin PATCH).
   > **The 14 readers were NOT re-counted** — that half remains dated 2026-08-31 and still owes a
   > re-run before any migration. Do not quote a refreshed reader figure; there isn't one.

   ⚠ **TRAP FOR WHOEVER RE-RUNS THIS: the admin writer is INVISIBLE to a literal-SQL census.**
   `site_admin_handlers.go:389` builds `UPDATE sites SET %s` with the column list assembled at
   **runtime** from `setClauses`, so the file contains no literal `email` adjacent to any `UPDATE`.
   Two separate greps on this refresh returned it as **absent**. A naive re-run therefore reports
   **three** writers and silently drops the unconditional one — **precisely the writer this
   constraint exists to worry about**, and the one an identity model most needs to account for.
   Search the file set for the **column name independently of the SQL verb**, then read each
   candidate. A count that cannot see a dynamically-built writer is not a census of writers.
5. **The delivery chain's source must stay explicit.** It reads
   `build_queue.direction.customer_email` today (**convention, not code** — no code reads
   `sites.email` for delivery, measured 2026-08-31). Whichever identity becomes the delivery
   address, it must be named in the 651 recipe, not inherited.

---

## 6. Consumers to tell, per the 2026-07-29 ruling

Naming them, because measuring that nothing breaks is not the same as their owners agreeing:

- **`site_delivery_and_editor`** — owns the delivery-contact reader and P5's seed; asked to be
  named on the consumer list. The delivery address is presumably one of the three identities and
  the choice changes their recipe.
- **the boxingonline.com lane** — the worked example, and the site where the conflation was
  visible to a customer.
- **the webdesign / intake-chat lane (box-side)** — every identity this RFC names has to be *asked
  for* somewhere, and that somewhere is the chat, in the owner's own environment. **No platform
  design survives if the chat cannot collect what it assumes.**
- **the classifier / identity-spec owners** — `sync_site_identity` and `update_site_content`'s
  sync arm are the writers that would need to learn which identity they are filling.
- **the admin-dashboard lane** — the unconditional PATCH writer, and the surface where an operator
  would set an operating party.

---

## 7. What is NOT proposed here

No schema, no migration, no code. `bugs_open/420`'s fix stays exactly as it is — APPROVED, live,
and correct for the defect it closed. This RFC does not reopen it; it names the larger question
the owner asked after reading it.

**Status of the residual meanwhile:** `bugs_open/420` §C stays OPEN and stays accurate — a
classifier-derived contact can still reach the published column with nobody asked. It is a live
gap, not a theoretical one, and it is the strongest argument for doing this properly rather than
adding a sixth store.
