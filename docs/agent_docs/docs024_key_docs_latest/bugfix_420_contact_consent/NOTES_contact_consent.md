# NOTES — 420 contact consent (append-only, newest at the bottom)

## 2026-08-31 — ownership, checked by asking rather than by tooling

`who-owns.py` reported the delivery lane ACTIVE on this bug and its session showed BUSY. That is
exactly the state where the tool cannot distinguish "working the incident" from "working the
class fix", because it reads commits. Messaged them. They replied that they hold the incident and
the boxingonline pre-delivery list only, and that the class fix and my proposed split matched
their file's intent — plus four do-not-touch items and one condition (census the readers before
moving the contract). **That exchange is what made the work safe to start, and no amount of
grepping would have produced it.**

## Half 2 was not a prediction — it had already happened

The bug file called the register licence "the subtler one" and framed it as what *could* happen
on a rebuild. It had already happened on order 1, and I could establish it by elimination:

- the build-briefing agent reads specs with `"aspect": "all"` (`050_build_briefing_agent.sql:57-66`);
- the briefing spec written **12:33:38**, twelve minutes after seeding, carried
  `contact.contact_email = <payer>`;
- the identity spec's contact block was **all null** (checked, not assumed);
- the customer's brief prose did **not** contain the address
  (`direction->>'objective' ILIKE '%designconsultancy%'` → false, over 855 chars).

Nothing else could have supplied it. **A registered claim propagated into a second regeneration
source inside one build** — which is precisely why removing the address from the column and the
pages could never have been sufficient.

## Two things in the bug file's own framing that turned out to be wrong

Both make the fix *cheaper*, which is the useful direction to be wrong in, but both would have
led to worse code if believed.

1. **"651's delivery-email-sender reads `sites.email`."** No code in `send_delivery_email_action.go`
   or `platform/delivery/` reads the column; `customer_email` is REQUIRED `input_data` supplied
   at dispatch. The coupling is a **recipe**, not a dependency. Had I believed it, I would have
   designed a migration and a reader-move for a chain that was never coupled.
2. **"chrome suppresses the footer contact block unless a flag is set" (candidate 3).** The block
   was **already** gated on a non-empty email (`component_library.go:1988`, the bugs 111 gate).
   Candidate 3 would have added a guard to a door that was already shut, while leaving the
   two-contract column — the actual defect — in place. The defect was the VALUE.

Also found a fourth writer the file did not list: the **admin PATCH endpoint**
(`site_admin_handlers.go:363-367`), unconditional — the deliberate operator override.

## A guard we got for nothing

`validate_page_content` loads `sites.email` as the one licensed "official" address and PASSES any
page publishing it. That is what made the leak invisible to our own honesty checks. Post-fix the
column no longer holds the payer address, so **any residual occurrence is now FLAGGED instead of
licensed.** That is the bug file's own fix candidate 3, obtained without building anything —
worth noticing, because it means the weakest candidate was already half-satisfied by the strong one.

## The trap I went looking for rather than hitting

Before finishing I asked what would happen if someone simply retried the boxingonline build.
`build_queue.direction.customer_email` is durable (UNIQUE(domain), no cleanup path), re-seeding is
the canonical retry (bugs_open/326), and `sites.email` is now legitimately **empty** — so the
fill-only-if-empty guard, which exists to protect an operator's correction, **inverts into a
refill mechanism**. A retry would put the address straight back.

Every sweep you would run to check reads clean, because they all describe the state after the
*last* seed and say nothing about the *next* one. LANDMINES entry added; live until the roll.

## Council submission — schema errors, and one I inflicted on myself

`DRY_RUN=1` refused `operation: "create"` (must be `add`) and `risks` as an array (must be a
string). I then "consistently" converted `grounded_in` to a string as well — and it must be an
array. **I applied one field's error message to a different field, and broke one that was already
correct.** The validator was right there and would have said so for free if I had re-run after
each single change rather than batching a fix with a guess. Logged in `WRONG_CALLS.md`.

## Tests: the old ones PINNED THE DEFECT

`seed_customer_identity_test.go` asserted that the payer's email lands in `sites.email` and is
minted as an "Enquiries reach" fact — i.e. the suite would have gone green on the exact behaviour
that published the owner's address. Rewritten to the new contract, and the **negative** assertion
made load-bearing: `argContaining` gained an `absent` list, because a test that only checks what
IS in the payload cannot see a leak, and "the payer's address is not here" is the entire claim.

Six mutations, each applied ALONE, each broke its named test — including rebinding `$2` back to
the payer email and re-minting the fact from it.
