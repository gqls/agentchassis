# What it would take to set up hosting accounts for our customers

Written 2026-09-04 for the owner, who asked after performing the Netlify handoff himself and finding
it took about forty minutes. This is a costed set of options, not a recommendation dressed as one —
three of the four are business decisions rather than engineering ones.

## What we know, measured today

**We already host the site, for six weeks.** `LiveLinkWindow = 6 * 7 * 24h` = **42 days**, and the
email tells the customer **30**. That gap is deliberate and documented in `prepare.go`: *"we promise
30 and serve 42 on purpose… a composer must never derive days from `LiveLinkExpiresAt`, or the email
will promise the slack away."* So the handoff deadline is a policy we chose, not a technical limit.

**There is no Netlify integration.** One mention of the word in the whole codebase, and it is a
comment in `zip_deliverable_action.go`.

**The customer's real experience, timed by the owner** (`DRAFT_2026-09-04…v3`): signup demanded on
the drop, a "security check" with no security check shown, a good 11-character password rejected, a
slow confirmation email, and a site that is **private by default and looks public to the person who
uploaded it**. Forty minutes.

**What Netlify gives them once they are in**, which changes the story: an **AI editing agent** on the
project screen, and **Domain management** that accepts an existing custom domain behind one Verify
button. The owner's framing, and it is the right one: *"that makes it easier for us to hand off to
the client without having to explain things too much or be too negative about it."*

---

## Option A — we create the Netlify account for them

**Blocked, and not by effort.** Netlify requires email verification to a mailbox the account holder
controls. We would either need the customer's inbox, or to use an address we control — at which point
the account is not theirs, and handing it over means handing over credentials we generated and
holding a copy until they change them.

For a product whose pitch is *"the files are yours"*, us holding the keys to their hosting is the
wrong shape regardless of whether it is possible.

## Option B — deploy to OUR Netlify account and invite them

Technically available: Netlify supports team members and site transfer, and there is an API.

**But it inverts the point of the handoff.** Their site would live under our team and our billing, so
we are still the host — just via a third party, with an extra dependency and a free-tier ceiling we
have not measured. The customer gains an editor and loses independence, and we gain a support
surface without gaining revenue.

Worth it only as a **stepping stone inside** Option C, never as a destination.

## Option C — keep hosting it ourselves, and charge for it

We host it today. The 30-day cut is a decision, not a constraint. Turning it into a paid tier removes
the entire Netlify problem for any customer willing to pay, and it is the only option that makes the
£59.99 domain buy-out obviously worth having — a domain they own, pointed at hosting that does not
expire.

**This is a pricing and commitment question, not an engineering one.** What it costs us is a
recurring obligation: uptime, and someone to answer when it breaks. What it is worth is that the
awkward part of the handoff simply stops existing.

**The engineering to support it is small** — the machinery already runs, and `live_link_expires_at`
is already a per-site column rather than a global. Extending or removing an expiry for a paying site
is a value change, not a build.

## Option D — leave the handoff, and make it good

**Cheapest, already largely done, and the owner's own reframe is what makes it viable.**

The instructions are rewritten from his real run (`DRAFT_2026-09-04…v3`), including the private-by-
default trap and the signed-out check that is the only way to catch it. His screenshots go on the
page. And the two things he spotted turn the handoff from an apology into an offer:

- **Netlify's AI agent edits their site**, so *"no changes are included"* stops being the last word.
  The honest positive line is that their host has an editor built in and they do not need us to
  change a headline.
- **Domain management takes a custom domain with one Verify button**, so a customer who buys their
  domain from us at £59.99 has a coherent, permanent, free end state: their domain, their hosting,
  their account, no monthly anything.

---

## What I would do, stated plainly

**D is done and should ship regardless of the others.** It costs nothing and it is strictly better
than what a customer would have received yesterday.

**C is the real question and it is yours.** It is the only option that removes the problem rather than
documenting it, and the engineering is not the hard part — the recurring obligation is.

**A is closed** by email verification and by the credential-custody problem.
**B is not a destination**, only a possible mechanism inside C.

**One thing worth deciding either way:** whether we recommend Netlify at all now that we know the real
cost is an account, a password fight, an email wait, and a privacy step that silently fails to
nobody-can-see-it. The alternative is not a different host with the same shape — every free static
host has an account and most now default to private. The alternative is C.
