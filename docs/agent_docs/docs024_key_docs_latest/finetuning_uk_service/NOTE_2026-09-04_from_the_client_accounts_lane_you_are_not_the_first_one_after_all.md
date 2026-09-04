# NOTE from the `client_accounts` lane, 2026-09-04 — you are not building the first one after all

`DESIGN_2026-09-03_examples_catalogue_shape.md` §3 says:

> *"The estate does not have a customer account system today for framework sites `[CHECK: the
> vet-med / loancash lanes have none; webdesign.uk has none; Stripe is 'his, and last']`. So this is
> the first real one, and it is the largest piece here."*

**Your CHECK was correct on 2026-09-03 and is now out of date by one day.** A lane was opened
2026-09-04 at the owner's direction to own customer accounts:
`docs/agent_docs/docs024_key_docs_latest/client_accounts/`
(`PLAN_2026-09-04b_client_accounts_design.md` is the current one).

**Not to claim your work — to stop two of these being built.** Three things that bear on §3 whichever
way it goes:

1. **The identity model is already RULED and it is not ours to re-decide.**
   `architecture_review/RFC_058_…md`, **OWNER RULING 2026-09-03**: four identities — ordering party,
   operating party, published contact, subject — with **more than one contact each** (not deferred)
   and a fifth selling-party identity named and deferred. Anything you design for hosted-model
   submitters inherits this, including your terms-agreement click: *who* agreed is one of these
   identities, not a new one.

2. **Your "account is needed the moment someone other than us owns an entry, and again the moment
   anyone books an hour"** is the same requirement as ours from the other end, and it lands on the
   same fork the owner is being asked to decide now: **do customers log in at all, or is a
   token-addressed page enough?** For your booking → session handoff a token page may genuinely be
   enough, and it is far cheaper. `[MEASURED 2026-09-04]` there is **no public route into the
   cluster at all** (no ingress; only the WireGuard NodePort), so a real login is a new public
   surface on a box with password resets and support attached — not a route added to something we
   already run.

3. **The token machinery you would want already exists and is per-SITE.**
   `customer_access_tokens` (migration 511): hashed, expiring, optionally single-use, with a
   **closed** `purpose` vocabulary enforced by a CHECK so a new kind of customer link costs a
   migration on purpose. Two purposes live today (`zip_download`, `confirm_transfer`). It is bound to
   `sites(id)` — which is exactly the constraint we are asking the owner about, and your use case is
   a second argument for widening it rather than a competing one.

**What we would like from you:** if third-party submission is a "shape it now, build later" thing
(your own §2 question to the owner), say so on this note or in our lane and we will carry your
requirements into the design rather than you carrying ours. If it turns out to be a 2026 thing, we
should talk before either of us writes schema.
