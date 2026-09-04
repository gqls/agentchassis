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

**Your Afternic prices are confirmed high, and now there's a number on it.**
Where a name has both an Afternic asking price and an independent appraisal,
the ask is a median of **5.4 times** the appraisal. Also worth knowing: only
419 of your 1,634 Afternic entries have an actual asking price at all — the
rest carry only a minimum-offer floor, and those floors run about 5.8 times
appraisal too.

On valuations themselves: the appraisal tool turned out to value *any* domain
name, not just ones you own — which matters because it refuses .co.uk entirely.
So the .co.uk half of the estate can be valued through its .com equivalent
instead, clearly labelled as the proxy it is. The daily limit is 300 names, so
full coverage of 2,945 domains takes roughly a week of windows; today's ran to
the limit.

## 2026-09-03 (later) — correcting the paragraph above about your Afternic prices

Two corrections, one to the finding and one to how I recorded it.

**The finding.** I said your asks run about 5.4 times the appraisal. That is
arithmetically right and it describes the wrong problem. Your prices were
applied in **bulk bands**, not per domain: of 419 asking prices, 250 are the
identical figure $4,999 and 136 are $25,000; of 1,215 minimum-offer floors,
845 are exactly $10,000. Because one side of that "5.4 times" is nearly a
constant, the ratio was really measuring how much the appraisals vary, not how
you priced.

What the bands do show is worse than a factor being too high: they don't track
quality at all. Names in the $4,999 band and names in the $25,000 band have
essentially the same appraisals ($1,549 against $1,646 median) — equal names
priced five times apart. The $10,000 floor band covers names appraised anywhere
from $25 to $24,511.

So the job isn't to scale your prices down. It's to give the portfolio
per-domain pricing for the first time, which will move many prices down a long
way and some genuinely good names **up**. Had you acted on my first version,
you'd have cut everything by a factor and kept the actual defect.

**The record.** I edited that paragraph in place rather than correcting it
here, which is against this file's own append-only rule — the point of the rule
being that what we believed at the time is part of the record. The original
wording is restored above and this is the correction. The underlying lesson
(check whether a variable actually varies before quoting a ratio over it) is
logged in the fleet-wide WRONG_CALLS.md.

## 2026-09-03 (evening) — a fence that was hiding live sites, and a £5,000 lesson

Three things happened today that changed the shape of the job.

**Thirty-three of your live websites were sitting in the "sell" pile.** The
do-not-sell list is built from the framework's own records, which only know
about sites hosted on Cloudflare. Your older sites sit on Clook — a different
host entirely — so the list simply couldn't see them. All 28 Cloudflare ones
were correctly protected; every single miss was a Clook site. Among them:
wpx.uk, which is the domain your own email address is on; designconsultancy.co.uk,
the company behind your Nominet tag; and leopardess.co.uk and leopardess.uk,
adjacent to a client site. I found them by checking every Nominet domain's
nameservers myself and then actually fetching each page, rather than trusting
the list. You've since gone through all 39 by name — 17 stay out, 21 released.

**You then reversed the premise: list the live sites too, priced high.** That's
now a separate track, and it needs a different instrument, because I can show
the appraisal tool is not up to it. Your own floor on relojistas.com is $12,000;
the tool prices it at $1,490 — eight times under. Your figure for webdesign.uk
is over a million; the tool would say about a thousand. The reason is
mechanical: it prices a *name* in its market and cannot see either a
category-defining exact match or a business attached to it. It stays the right
tool for the ordinary tail and the wrong one for your best names. I'm now
gathering real, realised sale prices — actual transactions, not asking prices —
to price that tier on evidence rather than on an algorithm that demonstrably
doesn't reach it.

**And the £5,000 lesson.** You mentioned in passing that cartoon.co.uk cost you
over £5,000. It was sitting in ordinary stock at the time, where it would have
been priced from an appraisal and sold keenly for a fraction of that. It escaped
only because you happened to say so. Nothing anywhere in this estate records
what you paid for anything — the registrar exports carry expiry dates and
nameservers and nothing else. So the same risk applies to the other 2,859, and
it breaks an assumption underneath the whole bottom-500 idea: pricing "keenly"
assumes the alternative is holding a cheap renewal, which isn't true for
anything you bought at auction. **This is the one thing I need from you before
any tail price goes anywhere** — even a partial answer would bound it: which
ones you paid real money for, or just "anything from an auction", or a rough
year. Nothing will be priced until that comes back.

---

**2026-09-04 — the safety net had a hole in it, and running today's job was
what would have torn it open.**

Today's job was the appraisal window: 300 valuations a day is all the account
allows, and we're 19% of the way through the estate. The handoff said do the
premium names first — the single dictionary words, the very short ones, the
four-letter .coms. Those are your best names and the ones the tool has been
worst at.

Before spending anything I checked the list of what to appraise, and it was out
of date in a way that mattered. It was built at ten in the morning on the 3rd.
You ruled at around seven that evening that all the financial domains are kept
together as an advertising network. The list still had 95 financial names at the
top — so a third of today's allowance would have gone on valuing things that are
never going to be sold. It also still contained all 23 of the domains you've
withdrawn, including the family names. I rebuilt it, and I wrote a script to
rebuild it, because "remember to rebuild the list" had been written down since
yesterday and hadn't happened — there was nothing to run.

**Then the real problem.** We have four guards that hold your best names out of
the automatic pricing — a name is held back if it's a single dictionary word, if
it's three characters or shorter, if it's a four-letter .com, or if you've ever
given me a real figure for it. Two of those guards were written as "hold this
name back *if we don't have an appraisal for it*."

Today's job was to go and get appraisals for exactly those names. So the act of
doing the work would have quietly switched the guards off. Not all at once, and
nothing would have looked wrong: the names would simply have stopped being
flagged and started carrying automatic prices. By the time I caught it, 61 of
them had already flipped — `healthcare.uk` among them, the £40,000 name that is
the entire reason the guards exist.

It was already happening, before today, on four names. `effectiveness.uk` is the
clearest: an ordinary English word, appraised at $3,576, sitting in the sell
list at **$350**. Nobody would have spotted that by reading the list.

**The fix is that the guards no longer care whether we have an appraisal.** And
today's numbers say that's right, because having one doesn't help. Asked
directly, the tool values `healthcare.uk` at $18,193 — against the £40,000 you
paid, so about $2,000 once it's discounted for being a .uk and priced to move.
`free.uk` it puts at $67,926, which becomes about $8,600 — against the roughly
£160,000 its sibling `free.co.uk` sold for. Both land at about **4%** of a real
number we actually know. An appraisal doesn't make these names safe to price; it
makes them *confidently* wrong, which is worse.

**One thing in the tool's favour, though.** I've been saying the appraiser can't
see your best names. Today refines that: it's the *fallback* that can't. When a
name has no appraisal of its own, the model borrows the middle of its category,
and that's where `healthcare.uk` got its $149 — the same figure it gave
`healthcarecareers.uk`. Asked about `healthcare.uk` directly, it says $18,193.
Still too low to use, but that's 2.8 times under rather than 340. So getting
through the remaining 81% of the estate is worth more than I thought.

**Nothing here changes the rule that appraisals don't set prices.** They're an
input to a conversation with you. What changed today is that the names that most
need that conversation will still be flagged for it tomorrow.

**2026-09-04, later — the tool was reading UK domains wrong, and finding that
out was an accident.**

Before spending the day's 300 valuations I wanted to settle something nobody had
checked: when the tool values `healthcare.uk`, is it valuing *that domain*, or
is it really valuing the word "healthcare" and ignoring the ending? It matters,
because our model takes the tool's number and then knocks it down to about a
fifth on the grounds that a .uk sells for less than a .com. If the tool had
already done that, we were doing it twice.

Fifteen valuations settled it. `ant.uk` comes back at $23,144 and `ant.com` at
**$8.2 million**. `design.uk` $23,558, `design.com` $3.1 million. The tool knows
exactly which ending it is looking at.

**So we were discounting twice, and had been all along.** Every UK domain with
its own valuation was being marked down about five times too far. That is what
put `effectiveness.uk` — appraised at $3,576 — into the sell list at $350. Fixed.
The tool's number now stands as it is for a domain valued directly, and the
discount is applied only where we're borrowing a .com value as a stand-in.

**An unexpected bonus.** Yesterday we set that UK discount to about a fifth,
worked out from real UK sale prices we'd gathered. Today's fifteen valuations
let me check it a completely different way — the tool's own view of .uk versus
.com on the same word. Across the ordinary names it comes out between a ninth
and a fifth. Two methods with nothing in common landing in the same place. That
number is now on firmer ground than anything else in the model.

**And fixing it broke something else, which is the part worth telling you.** With
the double discount gone, `vetzy.co.uk` — an invented name, not a real word —
jumped from $325 to $6,250. That was wrong, and the reason was the same fault
you already know about, running backwards. When a domain has no valuation of its
own we use the middle of its category. That category's middle had been set by
`felines.co.uk`, `veterinary.co.uk` and `bunnies.co.uk` — three real dictionary
words, all valued in the tens or hundreds of thousands, and all three names we
have *refused to price* precisely because the tool can't handle them. They were
being kept out of the sell list and still allowed to set their neighbours'
prices. Now the middle of a category is worked out only from the ordinary names
we would actually sell, and `vetzy.co.uk` is back to $325.

That has an honest cost I want to be straight about: taking those names out
leaves some categories with too few ordinary valuations to be reliable, so about
400 domains now sit on a weaker footing than they appeared to yesterday. They
were always on that footing — we just couldn't see it. About 158 more valuations
fixes every one of them, which is half a day's allowance.

**Where the day leaves us.** We went from 19% of the estate valued to 29%, and
**every one of your premium names — the real words, the short ones, the
four-letter .coms — now has a valuation and every one is still held back from
automatic pricing.** That was the whole point: get the numbers, and don't let
having them talk us into trusting them.

Two you asked about, `mieleonline.com` and `webuyanycarandvan.com`, are out —
the other thread tells me you withdrew both. For the record they since found a
live US trademark on "We Buy Any Car" covering exactly online car buying, so
that one was a real risk rather than a hunch.

**2026-09-04, evening — you said one number and it turned the day's conclusion
around.**

`scales.co.uk` cost you £3,500. The model had it at **$393,917**.

That is the first time one of your figures has caught the model being too
*high*. Every previous one caught it being too low, and I'd written the whole
thing up this afternoon on that basis — that the tool undervalues your best
names. That framing was wrong, and your one line is what showed it.

Here is what was happening. For a `.co.uk` the tool won't give a valuation at
all, so we look up the `.com` instead and knock 15% off. For an ordinary
made-up name that's a reasonable stand-in. For a real dictionary word it is
nonsense, because the `.com` of a real word is a landmark asset and the `.co.uk`
of the same word is an ordinary domain. `scales.com` is valued at $463,432, so
`scales.co.uk` was being carried at $393,917 — against the £3,500 you actually
paid.

**And it had already caught `cartoon.co.uk`, which is the part that bothers me
most.** That one has been sitting in my notes since yesterday as the example of
the tool being too *cheap* — it said $2,934 against the £5,000+ you paid. After
today's valuations came in, the same domain now reads **$739,424**. The same
example in the same table, quietly turned upside down overnight, with nothing to
show it had moved.

**The upshot: 72 of your domains were being carried this way, and they made up
74% of the estate's total value.** I have stopped the model producing a figure
for them at all rather than inventing a better discount — we have no real
evidence for what a one-word `.co.uk` is worth, and guessing is what got us
here. The estate's stated value goes from about $23.7m to **$6.1m**, with those
72 names listed separately as "cannot value".

Two things worth saying plainly. **First, no price was ever at risk** — all 72
were already held back from automatic pricing by the guards, so none of them
ever carried an asking figure. The protection worked; it was the *value* written
next to it that was wrong. **Second, that $23.7m never reached you** — I checked
before changing it. It only ever appeared in the tool's own output.

**And the standing request now has a number on it.** Today I spent the full
daily allowance — 299 machine valuations — and moved coverage ten points. Your
one sentence about `scales.co.uk` reversed the direction of the central finding
and took three quarters off the estate's stated value. That is the ratio I keep
asking about. Anything you can remember paying — even roughly, even "that one
was an auction" — is worth more than another day of the machine.
