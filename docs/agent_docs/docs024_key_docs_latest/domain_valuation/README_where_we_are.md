# Where we are — domain valuation (append-only, newest at the bottom)

## 2026-09-02 — lane opened: gathering the lists

You asked for every domain across Dynadot, Porkbun, Nominet and Spaceship to be
listed and valued — .co.uk, .uk and now the .coms too — with a view to selling
roughly the bottom 500 at keen prices, keeping whole categories together rather
than splitting up, say, the financial names.

Done today: asked each of the four registrar sessions for their full domain
list plus any valuations their registrar can produce; asked the Afternic
session to bring over its current asking prices as a comparison (you've said
they're generally too high, so they're an input, not an answer); and set a
search running over the earlier conversations where we discussed .co.uk/.uk
values, so we start from what was already agreed rather than from scratch.

The one list nobody holds yet is Nominet's — the ~1,500 .uk domains. The
walk failed on a connection blip earlier; when you have a minute, run in the
Nominet session (or here):

    ! python3 scripts/domains/nominet.py login
    ! python3 scripts/domains/nominet.py walk --months 120 > all_domains.txt

## 2026-09-02, late evening — three lists in, categorisation running, method drafted

The registrar sessions moved fast. Dynadot (451 domains), Porkbun (683) and
Spaceship (203) have all delivered their full lists — 1,337 domains in hand,
almost all .com, and about 85% of them parked at Afternic. None of those three
registrars offers valuations through their API except Dynadot, whose appraisal
tool ("Dynappraisal") is being fetched by that session now. Porkbun's
marketplace gave us something useful instead: asking-price comparables — 774
UK listings pulled already, with a .com pull to follow.

A first categorisation pass has sorted a third of the names into families
(financial, home & garden, AI, web design, and so on); a second pass on the
long tail is running now. The valuation method is drafted: every domain gets a
tier with its reasoning recorded, the bottom-500 sale is assembled from whole
weak categories so the financial names (and any kept family) stay together,
and keen prices never go below the £150 you already charge as a transfer-away
fee. The 19 domains carrying live sites are marked keep regardless.

One honest miss: I searched all 646 stored conversations on this machine and
the earlier .co.uk/.uk valuation discussion you remember isn't in any of them —
it likely happened on claude.ai in the browser, or on another machine. What I
did recover: your $12,000 Afternic floor on relojistas.com, the £150
transfer-away ruling, and the domain-value ladder doc. **If you can find that
old conversation (or just paste its conclusions), it would sharpen the
starting point — otherwise we price from the data now arriving.**

Things only you can do, gathered in one place:
1. Nominet walk (the two `!` commands above) — the ~1,500 .uk list.
2. Afternic: export your portfolio CSV into the afternic lane's inbound/
   folder, plus their bulk-upload template — that's the current-prices column
   AND the eventual repricing vehicle.
3. Atom.com: 58 of the Spaceship domains are listed there — an export from
   that dashboard if you want those asks in the comparison.
4. Porkbun: flip the global API-access toggle (Account Settings → API) —
   needed later for repricing writes, not for the valuation itself.

## 2026-09-03 — the whole estate is finally visible: 2,945 domains

Three things landed today and together they change the picture.

**The Nominet list arrived** — 1,606 .uk domains, the list that had never
successfully been produced before. Added to the 1,339 at Dynadot, Porkbun and
Spaceship, and with no overlap at all between them, **your estate is 2,945
domains**.

**You asked which domains aren't in the registrar lists, expecting them to be
expired.** Answer: almost none. There were 692 candidates, and 683 of them were
simply the Nominet names, invisible until this morning. Nine were genuinely
unaccounted for, and of those, **five are actually gone**: three
(chicklets.co.uk, demisexual.uk, protecty.co.uk) have lapsed and nobody holds
them; two (cheapbuild.co.uk, enables.co.uk) lapsed and were picked up by
somebody else, who is running live sites on them now — those are only
recoverable by buying them back. Two others sit at registrars we don't ask —
one at Dynadot but missing from Dynadot's own listing, one at **NameSilo**,
which may mean an account holding domains this valuation cannot see. Two .co
names could not be checked (that registry publishes no lookup service we can
reach). Worth acting on: **the three dead names are still listed for sale on
Afternic**, advertising domains that can't be delivered.

**Your Afternic prices need more than lowering — corrected later the same day.**
I first reported that your asks run about 5.4 times the independent appraisal.
That figure is arithmetically right but it describes the wrong problem, and the
real one matters more. Your prices were applied in **bulk bands**, not per
domain: of 419 asking prices, 250 are the identical figure $4,999 and 136 are
$25,000; of 1,215 minimum-offer floors, 845 are exactly $10,000. And the bands
do not track quality — names in the $4,999 band and names in the $25,000 band
have essentially the same appraisals ($1,549 against $1,646 median), so two
names of equal worth are priced five times apart. The $10,000 floor band covers
names appraised anywhere from $25 to $24,511.

So the job isn't to scale your prices down by a factor. It's to give the
portfolio per-domain pricing for the first time — which will move some prices
down a long way and some genuinely good names **up**. Also worth knowing: only
419 of your 1,634 Afternic entries carry an asking price at all; the rest have
only a minimum-offer floor.

On valuations themselves: the appraisal tool turned out to value *any* domain
name, not just ones you own — which matters because it refuses .co.uk entirely.
So the .co.uk half of the estate can be valued through its .com equivalent
instead, clearly labelled as the proxy it is. The daily limit is 300 names, so
full coverage of 2,945 domains takes roughly a week of windows; today's ran to
the limit.
