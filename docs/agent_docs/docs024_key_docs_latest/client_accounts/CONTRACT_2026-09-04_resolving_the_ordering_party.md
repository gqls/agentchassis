# CONTRACT — 2026-09-04 — resolving "which client is this order for"

**Written by `client_accounts` for the `stripe` lane, who asked for rule (1) settled before drafting
the order-side contract.** This is the identity half only. The money half — when it is called, what
the order carries, what the webhook does — is theirs.

**Status: proposed by this lane, not yet built and not yet council-reviewed.** It is written to be
argued with; the measurements in it are the durable part.

---

## 1. What is settled, and by whom

- **Owner, 2026-09-04:** *"one account per PERSON, not per site."* An account is the customer;
  someone with three sites is one customer with three sites.
- **Owner, 2026-09-03 (`RFC_058`):** four identities — **ordering party · operating party ·
  published contact · subject** — with **more than one contact each** (not deferred), and a fifth
  selling-party identity named and deferred.
- **Owner, 2026-08-10:** customer identity rides the existing `clients → networks → sites` chain;
  `sites` gains no ownership columns.

**The synthesis these three force**, and the sentence the whole contract rests on:

> **A party is WHO EXISTS. An identity is a ROLE that party holds ON A SITE. An order belongs to the
> party holding the ORDERING role.**

`clients` is the parties table. `billing_orders.client_id` is *the ordering party*. It is not "the
owner", not "the client of the site", and not "the account that will log in" — those are all
different questions and conflating them is how the estate got four unjoined identity stores.

## 2. The rule: one function, and it is keyed on the PARTY

> **`ResolveOrderingParty(ctx, tx, email, displayName) (clientID, error)`**
>
> Find the party whose contacts include this email. If there is none, create one. Return its id.
> **Never keyed on a site. Never creates a site. Never merges two parties.**

The direction of the dependency is the load-bearing part, and it runs **one way**:

```
order → party → network → site
```

**not** site → client. A site is attributed *from* an order; an order is never attributed from a site.

## 3. Why the site-keyed version is the trap — measured, not argued

The `stripe` lane named this correctly: *"find-or-create the party for this email"* and
*"find-or-create the client for this site"* are different functions, and the second is the trap. The
measurement that shows why:

`[MEASURED 2026-09-04]` **60 sites · 42 carry a `network_id` · 1 distinct network between them · 18
carry none.** `EnsureSiteRecordAction` (`platform/orchestration/actions/site_db_actions.go:178`)
attaches every site to that one default network and takes no network parameter.

So **`the client for this site` resolves to `Default Client` for every site in the estate** —
including boxingonline.com, the one site a customer actually paid for. A site-keyed resolver would
therefore attribute a paying customer's order to the default row and look entirely correct doing it.
It would also have to invent a customer for each of the 17 `pool` and 2 `test` sites, which have no
customer and never will.

**A site does not determine a party.** Most sites are ours; some will have several parties in
different roles; a party may hold the ordering role on three sites at once.

## 4. Implementation now, and the shape it must not cement

RFC_058's contacts relation **does not exist yet**, so today the lookup can only read a column.

- **Today:** match on a canonical form of `clients.email`.
- **When the contacts relation lands:** the same function matches a contact row instead.
- **The signature does not change**, so no caller is rewritten. That is the point of stating it now.

⚠ **Do not let the temporary implementation become the model.** The pre-plan's §6 and RFC_058's
addition 2 both say it: modelling contacts as columns is debt, and the owner's deferral of the fifth
identity is only cheap under a relation. **This contract is party-keyed precisely so the column is an
implementation detail rather than the contract.**

## 5. What makes it safe — four requirements, each with its reason

**5a. Canonicalisation is `lower(trim(email))`, and deliberately nothing cleverer.**
Reuse the estate's proven shape rather than inventing a third:
`canonicalEmail(e) = strings.ToLower(strings.TrimSpace(e))`
(`docs/agent_docs/docs024_key_docs_latest/noted_rebuild/box/noted-engine/store.go:95`, which also
stores the address **as typed** alongside the canonical form — do the same).
`internal/auth-service/user/repository.go:110` already lowercases too.

**No plus-tag stripping, no dot removal.** Those merge parties a human would call distinct, and
**merging is unrecoverable while splitting is not.**

**5b. A UNIQUE index on the canonical email, or find-or-create races.**
`[MEASURED 2026-09-04]` `clients` has exactly two indexes — `clients_pkey` and
`clients_external_id_key` — so **nothing stops two rows sharing an email today.** Two simultaneous
orders from a new customer would create two parties, and both inserts would look correct.

The index must be **partial** (`WHERE email IS NOT NULL AND btrim(email) <> ''`) because two of the
three live rows have no email and must stay legal.
`[MEASURED 2026-09-04]` there are **no duplicate canonical emails** in `clients` today and only
**one** row carries an email at all, so the index applies cleanly with no reconciliation.

**5c. Find-or-create must be atomic and inside the caller's transaction** — hence `tx` in the
signature. The proven idiom is noted-engine's: `INSERT … RETURNING`, catching `23505` and
re-selecting (`store.go:140-149`). Not read-then-insert.

**5d. No email ⇒ REFUSE. Never fall back to `Default Client`.**
This is the estate's own most-repeated lesson wearing a new coat: `bugs_open/420` happened because an
empty field invited a fill, and RFC_058 §5.2 records the same trap proved at two more layers. A
resolver that silently attributes an unidentifiable order to the default row produces a wrong answer
that is **indistinguishable from a right one** at every call site. Fail closed and let a human look.

## 6. Who may create a party

- **`ResolveOrderingParty` — the only automated creator.** Any path that needs a party for an order
  calls it; no path does its own find-or-create.
- **`POST /admin/customers`** (`internal/core-manager/admin/customer_handlers.go:190`) — a human
  deliberately creating one. Stays. It is how `Boxing Online` exists.
- **Nothing else.** In particular **Phase 0 does not decide**: `EnsureSiteRecordAction` will *consume*
  a network belonging to an already-resolved party, not resolve one itself.

**So the answer to the `stripe` lane's worry is stronger than "our two producers agree": there is one
producer.** The build path reads the decision; it does not make it.

## 7. What Phase 0 changes, and the notice owed

`EnsureSiteRecordAction` gains an optional network parameter, supplied from the order. **Opt-in,
unsafe side default OFF** (owner ruling 2026-08-02 §2): absent the parameter it behaves exactly as
today and attaches to the default network, so nothing existing changes behaviour.

**This lane will tell the `stripe` lane before that ships**, as they asked. It is also council-scope
(`platform/`) and a register entry is owed in the same commit.

## 8. Open, and honestly so

- **The payer's email may differ from the intake email.** Then this rule produces **two parties**,
  correctly — we cannot know they are one person. **Merging is a human decision through the admin
  surface, never automatic** (5a's reasoning). Whether the delivery/account page should offer "this
  is also me" is a product question nobody has asked the owner.
- **Our own sites need a party too.** The 17 `pool`, 2 `test` and 1 `system` sites should be attributed
  to an explicit *us* party rather than left on `Default Client` by inheritance — otherwise "unowned"
  and "ours" stay indistinguishable, which is the ambiguity Phase 0 exists to end.
- **`clients.external_id` is NOT part of this contract.** `[MEASURED 2026-09-04]` it is empty on the
  only real customer row and the one live paid event carried `"customer": null`. It becomes
  load-bearing under `mode=subscription` / `customer_creation:"always"` — i.e. the £10/month rental —
  and that is the `stripe` lane's decision to take deliberately.
