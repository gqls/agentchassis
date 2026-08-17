# The new commercial terms — which attested facts break, and in what order to change them

**Owner, 2026-08-17:** *"The payment sentence should not have a preview step. The preview
step is for me and not the user. It is more of a take it or leave it, you've paid anyway
sort of deal but we can't phrase it like that."* Plus: exclusions should carry **lists and
links to third parties** who do take that work on; the hero must not mention a preview
stage and can be **bolder about the no-frills positioning**; and **do not claim anything
untenable**, because a second, fiercer brand (cheaper, even less service) is coming from
the same owner.

**Nothing here is applied.** The wording is the owner's to rule. This file exists so the
change is made once, completely, in the right order.

## 1. THE URGENT PART: the chat bot is already telling customers the old terms

Measured 2026-08-17, one live turn through `preview.webdesign.uk/api/chat`:

> Q: *"Do I get to see the site before I pay?"*
> A: **"Yes. You'll get a private preview link to review the finished site. You only pay
> the £149 once you've approved it."**

The bot builds its system prompt from `evidence_base` facts fetched through the relay, so
it states whatever the register attests. **Pages can wait for the rewrite; the bot cannot** —
it is answering the pre-sale question, live, with terms the owner says are wrong.

**The useful consequence:** fixing the REGISTER fixes the bot **without any rebuild or
deploy** — the relay refreshes every 5 minutes and the next conversation uses the new
facts. That is the fastest correct lever available.

## 2. Three facts break, not one — and fixing only the obvious one ships a contradiction

| fact | status under the new terms | why |
|---|---|---|
| `payment_after_approval` | **FALSE — must be replaced** | claim: *"The customer sees the finished site on a private preview link and pays after they have approved it. Nothing is taken before that."* This is the sentence the owner named. |
| `no_refund` | **CONCLUSION SURVIVES, JUSTIFICATION BREAKS** | claim: *"No refund is offered. **The customer approves the site before paying, so there is nothing to return**…"* and its `writer_line` is literally *"We do not offer refunds. **You pay once you have approved the site.**"* |
| `delivery_preview_and_zip` | **PARTIALLY FALSE — needs an owner decision** | writer_line: *"You get a **private preview link**, then a ZIP…"* If the preview is internal, does the customer get a preview link at all, or only the finished site plus the ZIP? |

**This is the trap.** `no_refund` is where the copy would silently contradict itself: change
only `payment_after_approval`, and the writers still hold a `writer_line` telling them to
say *"you pay once you have approved the site"* — so a page can state "pay upfront" and
"pay after approval" in the same breath, each traceable to an attested fact, and the claims
gate would pass both because both are attested. **All three change together or none does.**

Facts that are UNAFFECTED and must not be disturbed: `price_total`, `price_is_total_no_vat`,
`ai_built`, `build_duration`, `no_changes_included`, `no_lock_in`,
`hosting_and_domain_not_included`, `taking_it_further`, `yours_to_change`, `queue_limited`,
`contact`, `third_party_options`.

## 3. Two questions only the owner can answer

1. **Does the customer see the site at all before paying?** "Take it or leave it, you've paid
   anyway" reads as *pay first, then receive*. But `delivery_preview_and_zip` may still be
   true AFTER payment (preview link to view, ZIP to download). Which is it?
2. **What replaces the reason in `no_refund`?** The current justification ("you approved it
   first") is gone. The honest replacement is a plain statement of the bargain — the price
   buys one build, taken as it comes — but the wording is his.

## 4. "Do not claim anything untenable" — the concrete constraint

A cheaper, less-service sibling brand is coming **from the same owner**, so any claim that
positions webdesign.uk as the floor becomes false the day that launches, and it would be
undercut by its own author. **Ban the superlative, keep the absolute:**

- **Do not claim:** cheapest / lowest price anywhere / nothing simpler exists / no one does
  less / the leanest offer / best value — any comparative or market-wide superlative.
- **Safe:** absolute statements about THIS product — £149 total, no VAT, no changes, no
  refunds, files are yours, no lock-in. These stay true whatever else the owner launches.

Recommended as `banned_claims` additions when the register is next superseded, so the
constraint is mechanical rather than remembered. (The sibling lane owns the register trail —
coordinate; do not fork it.)

## 5. The positioning the owner is asking for

*"More proud of the positioning… it is an unusual brand and we can be more bold about it."*
The material is already attested and currently reads as apology. Stated as deliberate
choices it reads as confidence: one price, one build, no upsell, no lock-in, no monthly fee,
files handed over outright, and a named third party for everything not included
(`third_party_options` already holds six, by category — this is the "lists and links" the
owner asked for, and it already exists as a fact). House voice still binds: no em dashes, no
agency-marketing weight, plain British English.

## 6. One thing worth checking before the terms ship — flagged, not decided

Payment in full **before** the customer sees anything, with no refund and no changes, is a
stronger position than the current one, and it is the owner's to take. Worth a deliberate
check on **who the customer is**: for business customers this is ordinary freedom of
contract, but the Consumer Contracts Regulations 2013 give a **consumer** buying at a
distance a 14-day cancellation right, which sits awkwardly with pay-upfront-no-refund unless
handled explicitly. Most buyers here are businesses, so this may be a non-issue — but the
register is the estate's claims layer, and a term that cannot be enforced is exactly the
kind of claim it exists to keep off the page. **Owner's call; noted once, not repeated.**

## 7. Order of operations, when the owner rules

1. Owner rules the terms and the two questions in §3.
2. **SUPERSEDE** the three facts in one transaction (never edit in place; inherit `pinned`;
   claimscan against the live corpus AND the current register as control first).
   → the chat bot self-corrects within ~5 minutes, no deploy.
3. Add the §4 superlative bans in the same supersede.
4. Only then rewrite the pages (hero, the exclusions/positioning copy, `bugs_open/299`'s
   CTA), through the spec and planner, per the owner's instruction.
5. Re-verify the chat's answer to *"Do I get to see the site before I pay?"* — that question
   is now the acceptance test for this whole change.
