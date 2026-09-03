# Where the "someone else's account details" listings actually are

Owner, 2026-09-03, on 50 .co.uk domains: *"They have someone else's account
details in the listing and some of them have only just been added."*

## Located: they are listed on SPACESHIP, and not under the Spaceship account we can see

`[MEASURED 2026-09-03]`

- **44 of the 50 are delegated to NamePros nameservers**
  (`ns1.namepros-dns.com` + `ns2.<random>.ns.namepros-dns.is`), and 2 more
  (`whiskysales.co.uk`, `parksforsale.co.uk`) to Spaceship's own
  `launch1/launch2.spaceship.net`.
- **Those nameservers serve a Spaceship for-sale lander.** Fetched
  `whiskysales.co.uk`: *"WhiskySales.co.uk is for sale on Spaceship"*, "Listed
  with spaceship.com", **asking $4,999**, Buy Now plus Make Offer. No seller
  identity is shown publicly.
- **They do NOT appear in the Spaceship SellerHub API export** taken from the
  account whose key this estate holds — 831 rows, byte-identical across 09-02
  and 09-03, and none of the 50 in it. SellerHub *can* list domains registered
  elsewhere (795 of its 831 rows are), so their absence is not explained by the
  domains being at Nominet.
- **Dynadot is ruled out** (their lane, same day): the account is verifiably the
  owner's (`dqls` / Anthony Appleby / uk@websy.uk), the names have no Dynadot
  listing, and Dynadot's listing API carries **no seller/account/payee field at
  all**, so the display he is describing cannot originate there.
- **The registry side is clean**: all 14 new names verify as his via the
  DESIGNCONSULT tag, four in the post-registration add period.

**So the listings are real, they are on Spaceship, and they are invisible to the
Spaceship account we hold a key for.** That is exactly the shape of "someone
else's account details in the listing", and it is a payout/transfer risk: a sale
through a listing under another account does not necessarily pay or transfer to
the owner.

## The one thing that muddies it, stated rather than hidden

The lander's **$4,999 is precisely the owner's own standard band** — 250 of his
419 Afternic buy-now asks are the identical figure `[MEASURED 2026-09-03]`. So
whoever created these listings used his pricing convention. That is consistent
with him (or someone acting for him) having listed them, and *inconsistent* with
a stranger listing domains they do not control. It does not resolve whose
account receives the money, which is the part that matters.

## What only the owner can do

The seller/payout fields are **dashboard-only** on every marketplace involved —
confirmed against the full API surface by both the spaceship and dynadot lanes.
Nothing further can be established from here. He should open the Spaceship
account these are listed under and check the payout/seller details, and tell us
which account it is.

⚠ **Do not price or re-list any of these 50 until that is settled.** A price
from this lane landing on a listing that pays someone else is the worst
available outcome.
