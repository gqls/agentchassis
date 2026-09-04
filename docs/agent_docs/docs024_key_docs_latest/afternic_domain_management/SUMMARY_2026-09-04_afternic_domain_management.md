# SUMMARY — Afternic domain management (2026-09-04)

## What we're trying to do

The owner asked for his domains at Afternic (afternic.com, GoDaddy-owned,
absorbed Dan.com) to be manageable by Claude sessions: which domains are
listed, at what prices, how they're performing (views/leads/sales), and
which ones are actually verified and delegated to the marketplace versus
just sitting there. The scope he set, in three parts: listings + pricing,
sales + leads, verification + NS state.

## Where we've come from

The starting assumption — that this would be an "API setup" task — was
wrong. Afternic has no self-serve seller API. Their real, documented APIs
are for registrar partners in the Fast Transfer programme; what they built
for ordinary sellers is a chat assistant inside their own dashboard, not
anything with a key. People who automate Afternic today do it by borrowing
their browser login session, which works but is unofficial and breaks
whenever the session expires. The owner was offered that route, official
API access (slow, possibly refused), or a no-credential loop — export the
portfolio, we parse it; when he wants changes made, we fill Afternic's own
bulk-upload spreadsheet and he uploads it. He chose the no-credential loop.

## What we've done

Built a parser (`scripts/domains/afternic-csv.py`, register entry OPP-013)
that reads the owner's portfolio export defensively: it matches columns by
their header text, never by position, and refuses outright — rather than
guessing — any row whose cell count doesn't match its header. That
specific discipline exists because of a real incident in July: a
positional read of a pasted dashboard invented a $0 minimum offer that
reached five separate documents before the owner caught it. The parser
also takes a value the owner quotes off his own screen and checks the
parse against it, so a wrong column mapping fails loudly instead of
quietly.

The owner supplied his first real export on 2026-09-03 — 1,634 domains.
All three prices he quoted from the dashboard matched exactly, which
proved the parse was reading the right columns. One thing the real data
taught us that no documentation said: Afternic writes `0` in a price
column to mean "not set," not an actual zero price — the parser was
updated to treat 0 as no-value on Buy Now and Floor, and that's now
confirmed against real data rather than assumed.

A second Claude session working on a whole-portfolio valuation asked for a
normalised copy of the same data, since the owner considers his Afternic
prices generally too high and plans to reprice a large tranche of the
portfolio for sale. We built that hand-off (`valuation-csv`) as a standing
step after every ingest — it picks one usable price per domain (preferring
an actual asking price over a floor or minimum offer, and naming which one
it used, because a valuation must not mistake a floor for an asking
price), and it now runs automatically. Two follow-up refinements the
valuation session asked for — treating a floor honestly as a floor rather
than a price, and marking the currency as an assumption "USD-assumed"
inside the data itself rather than leaving it to prose — were both built
in, and their own cross-check of our feed against their registrar census
turned up a genuinely useful finding: 689 of our 1,634 Afternic domains
aren't in the registrar estate anyone has enumerated yet, meaning the
Afternic list is effectively a preview of an unlisted portion of the
owner's Nominet-registered names.

The valuation lane also relayed two owner decisions that reached back into
this lane's own plan. First, a set of six Afternic listings turned out to
advertise domains the owner no longer owns — three unregistered, two now
owned by someone else running live sites on them, one lost to NameSilo,
which the owner has ruled entirely out of scope. We confirmed all six were
genuinely still listed (headline prices $10,000–$50,000) and queued their
removal; before that queued item was ever acted on through our tooling,
the owner removed all six himself directly in the dashboard and registered
a new domain, `enables.uk`, in the process. Second — and this changes how
future repricing will work — the owner ruled that marketplace floor prices
and the pricing tier shown on a site's own about-page should be unlinked
entirely; it turned out that link had only ever been a stated intention in
an earlier plan and nothing in the actual code enforced it. That
simplifies the plan considerably: repricing will now come from one prices
file the owner edits, not from a per-site lookup that barely existed.

## Where we are now

The read side of the loop works and has been proven against real data.
The owner's export is parsed, cross-checked, snapshotted, and its
findings have already changed real state — six bad listings are gone and
one new domain is registered. The valuation lane has a working feed of
our data and is actively using it. What doesn't exist yet is the write
side: nothing currently turns a desired price list into an upload Afternic
will accept, because we don't yet have Afternic's own bulk-upload template
file, and building a writer against a guessed format was deliberately
avoided.

## Where we're going

Two things are needed from the owner to unlock the write half: the
`bulk_upload_sample_v3.xlsx` template from Afternic's own bulk-upload
page, and — separately, whenever he's ready — the actual prices he wants
set. Once both exist, the generate half gets built, and it has two things
already queued to go out in its first real run: the valuation lane's
repricing (once they finish their appraisal), and nothing else outstanding
— the six dead listings are already gone by the owner's own hand. Beyond
that, the lane will keep ingesting whatever export the owner drops in
next, each one diffing automatically against the last so sales and price
changes surface without being asked for.
