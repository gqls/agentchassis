# Where we are — news feeds across the domain portfolio

Plain-prose running log, append-only, newest at the bottom.

---

## 2026-07-19 — first look

The question was: we're rolling out thousands of domains, most should have a news
feed, and we don't want to pay for each one separately. Could we get away with a
dozen feeds that cover most of the verticals?

The short answer is yes, a dozen is about right — but with one important
correction, which is that a dozen pools covers about two thirds of the domains,
not "the vast majority". The other third shouldn't have a news feed at all, and
that turns out to be the more useful finding.

Here's why. I went through the 1,625 domains on the list and sorted them by what
they're about. Roughly a third are either short brandable names with no meaning in
them — things like `2v.uk` and `aakn.com`, 126 of them are six characters or
shorter — or they're very narrow product sites: adjustable walking sticks, plastic
ducks, tile trimmers. There is no news about plastic ducks. There is no supplier we
could buy it from and no clever architecture that conjures it. If we point a news
feed at those domains, what we get is not a cheap news feed, it's invented filler
sitting under a heading that says "Latest News". Across hundreds of domains that is
actively harmful — these sites' main asset is that they look legitimate, and
nothing undermines that faster than a page of obvious generated padding.

So the useful way to divide the portfolio isn't by industry, it's by whether real
news actually exists for it. Where it does, we pool. Where it doesn't, we should
be doing something else entirely — seasonal and evergreen content refreshes rather
than a feed. The Christmas gift domains are the clearest example: there are sixteen
of them, and what they need is a calendar, not a news wire.

On the two thirds that do justify a feed, the good news is that they concentrate
much harder than I expected. The single biggest group is money — mortgages, loans,
savings, credit, insurance, pensions, investing — and that's about 231 domains all
of which can be fed by exactly one stream of UK financial news. Base rate moves,
FCA announcements, Budget, lender changes. One feed, 231 domains. That alone is a
seventh of the portfolio. After that it's marketing and web services (218),
construction trades, industrial and plant, travel, AI and tech, vehicles, health,
business services, property, energy, vets, and jobs. Thirteen pools, about a
thousand domains.

On the cost question itself, the reason it currently costs per-domain money is that
every part of the pipeline is built per-site, including the parts that have nothing
site-specific about them. Fetching an RSS feed and reading what's in it produces the
identical result no matter which of our domains asked for it. The only genuinely
per-site step is deciding which stories matter to that particular site — and that
step, it turns out, we can do for free, because the database already has the vector
search extension installed and we already run our own embedding model in-house for
other purposes. So the ranking that makes each site's feed feel bespoke costs us
nothing per site.

The effect is that our costs stop scaling with the number of domains and start
scaling with the number of news stories in the world, which doesn't go up when we
buy another domain. That's the whole design in one sentence.

Two things I got wrong today and corrected. I initially reported that 1,176 sites
were already set up wanting news feeds. That was wrong — I'd counted rows in a
versioned table rather than sites. The real number is four. Not four hundred, four:
gaswholesalers, relojistas, robot-hands and ai-agent-orchestration. The whole news
pipeline is a four-site prototype. That's genuinely good news, because it means
there's nothing to migrate and we can make the pooled design the default rather
than retrofitting it. The second thing was nearly reporting that a piece of the
scheduling machinery wasn't configured at all, when in fact its configuration lives
somewhere I wasn't looking. Both are written up properly in the notes.

One thing I need a decision on. There are a lot of near-duplicate domains — eleven
variations on "insurance", ten on "landlord insurance", nine on "health insurance",
`bestinsurancerate.co.uk` alongside `bestinsurancerate.uk` and so on. 358 domains
collapse into 146 concepts. Are those meant to be separate live sites, or are most
of them redirects pointing at one? It matters, because if they're separate sites
showing the same six headlines, that's a duplicate-content footprint across the
portfolio that a search engine will notice even though it's invisible when you look
at any one site.

Also worth flagging: 21.7% of these domains have any traffic at all, and one
domain — wayfaringlondoner.com — is 27% of the entire portfolio's views on its own.
So whatever we build, it's worth starting with the handful of domains that have
actual readers rather than spreading it thin across the dormant majority.

Nothing has been built. This is design only, and the plan, notes and runbook are
now in place alongside this file.

---

## 2026-07-19 — answers back, and a new directory

Owner came back on all three open questions.

The duplicate domains are meant to be **separate sites aimed at different target
markets**, so that the news on each can be angled differently. That's a bigger
decision than it sounds, because it moves the per-site ranking layer from being a
cost optimisation to being the thing that makes two similarly-named domains
genuinely different sites rather than the same site twice. It also means we now
need a real definition of who each domain is *for* — and that can't be derived
from the domain name, because the names are nearly identical in exactly the
families we most need to separate. I've flagged that as the critical unknown to
settle before anything gets built.

We're starting with the domains that have traffic. Agreed and recorded.

On the paid tier — yes, but the owner's point is that dedicated news sources
alone aren't a product, and I think that's right. A buyer can't tell a dedicated
feed from a pooled one. So that's filed as a direction with candidate axes
(bespoke research, calculators and tools, data pages that update themselves, lead
capture) rather than a spec, to be decided once the free tier actually works.

The most interesting thing to come out of this round is the owner's idea for
**packaged topic features**: pick a topic each week, break it into its parts —
for gas prices you'd pull in oil price history, the Hormuz situation, inflation,
collected opinion — and build a proper feature out of it that gets updated until
it stops being relevant. That's a much better instrument than a headline list,
and it's the owner's own answer to the duplicate content problem.

I'd add one thing to it, which I've written into the file. If we write one
package and publish it across all 231 money domains, that's *worse* for duplicate
content than headlines were, not better — long-form near-identical prose is the
shape that gets punished hardest. The fix is the same trick that makes the feed
pooling work in the first place, applied one level up: the research is shared and
done once (the expensive part), and each site generates its own angle on top of
it (the cheap part). Same facts and citations, genuinely different articles,
because they really are different articles. A haulage site gets what Hormuz means
for diesel; a landlord insurance site gets energy costs and tenant arrears. That
also happens to be exactly what the owner asked for when he said each domain
should be focused on a different target market.

New directory this session: **`features_open/`** at the repo root. We had nowhere
to put "things we've decided to build" — `bugs_open/` is explicitly for what's
broken in production right now, and putting design ideas there would wreck the
one question that directory exists to answer. So `features_open/` mirrors it, with
`FEATURE_` files for things to build and `RISK_` files for known hazards in things
we're about to build. Three entries so far: the packaged features idea, the paid
tier, and the duplicate content risk.

The duplicate content problem is worth reading properly (it's `002`). The short
version is that it's invisible from inside any single site — every site looks
fine on its own, and the defect only exists in the relationship between them. And
because so many of these domains are the same phrase on different endings, the
risk isn't just that a duplicated block ranks badly, it's that the whole estate
starts to look like a low-quality domain network. I've written down how to
actually measure it rather than just worry about it, and the rule that we test it
on one pool before rolling anything out fleet-wide — because pulling the feeds
back later doesn't undo the footprint.

027 is another thread's, so I've left it alone.

---

## 2026-07-19 — the piggyback question, and a design of mine that was wrong

Owner asked me to look hard at how the platform already handles decisions that
span multiple sites, so we could lean on that rather than invent something. Good
instinct — it paid off twice, once by finding what to reuse and once by showing me
something I'd got wrong.

**The thing I got wrong.** When I first sketched how pooling would work, I said
we'd add a "pool" column to the feed tables and allow the site column to be empty
for pooled items — and I pointed out that the items table already allows an empty
site, which I read as evidence someone had anticipated pooling. That was a
plausible story and it was wrong. This platform has an established way of handling
work that doesn't belong to any one customer site, and it's the opposite of what I
proposed: it creates a **fake site** to own the shared work. There's one already,
called `system.internal`, and the reason it exists is written down — they
deliberately chose a synthetic site record over allowing an empty owner, so that
all the ordinary per-site machinery keeps working unchanged on shared things.

Applied to us, that means **each news pool should just be a site**. A fake one,
flagged as such, that subscribes to sources and ingests articles exactly the way a
real site does today. The upshot is much less work: no schema changes, none of the
existing fetching or de-duplication code has to be touched, and we avoid changing
a database index that has already caused a fleet-wide outage in this repo once
before. Only the "which articles does this real site show" part is new.

I've written that up as a visible correction rather than quietly editing the old
plan, because the reason I got it wrong is worth keeping: nothing about my first
sketch looked wrong from the inside. It was coherent and I had a story for it. It
was wrong because I hadn't looked at how the platform already solves that shape of
problem — which is exactly the failure mode our own guidance warns about.

**On the target market question**, the news is better than I expected. Site
records already carry a free-form set of "aspects" — identity, strategy, design
intent and so on — with no restriction on what you can add, so a new one costs
nothing. And there's already an `audience` aspect with a couple of entries in it,
plus a piece of dead code in the admin API that was clearly written in
anticipation of exactly this. Someone reached for this once and stopped. On top of
that, every one of our eleven sites already has a written description of who it's
for, and several are genuinely good — relojistas' says "enthusiasts, collectors
and the curious about watchmaking in Spain, Mexico, Chile and the rest of Latin
America". So we're not starting from a blank page.

The rule for how a pool's default profile relates to a single site's own profile
is one we already have and have already been burned by: the component library. A
shared component sits in a library, a site takes a copy when it needs its own
version, and the standing rule is that you only edit the shared one for neutral
improvements — anything with the site's own voice in it has to be its own copy.
That maps onto our problem exactly. A pool gets a default audience profile;
near-identical domains like the insurance ones must take their own copy, because
sharing one guarantees they rank the pool identically, which is the whole thing
we're trying to avoid.

**One genuine gap, which nothing can cover.** Everywhere the platform talks about
what makes a site distinctive, it means distinctive *versus outside competitors* —
never versus another one of ours. There is no field anywhere that says "this site
is positioned differently from that site of ours". That's the one actually new
thing needed, and it's small, but it has to be deliberate.

**Two warnings I'd want us to heed.** A very similar audience design was written
before and then deliberately reverted, and I don't know why — worth finding out
before we rebuild the same shape. And the question "who is your target audience?"
used to be a required question when a site was set up, and has since been dropped
from the form. We've been quietly losing this information, which is part of why
we're now short of it.

Last useful find: the duplicate-content check I said we'd need is nearly written
already. There's a check that spots two sites accidentally sharing a colour
palette, and it works by comparing every site against every other site. That's the
same shape as comparing news blocks — same join, same output, different thing
compared. So the measurement isn't a research project, it's a sibling of something
that already runs.

The packaged-features idea is now written up as the design rather than a footnote,
including the point that the audience profiles have to come first — build packages
before profiles and you get 231 variations on one article.

---

## 2026-07-20 — four decisions executed, and the first live change of the workstream

Today the owner settled four things and we acted on all of them.

First, the reassuring discovery. Yesterday I'd flagged a worry: an earlier
audience design had been built and then rolled back, reason unknown, and I wanted
to know why before rebuilding the same shape. The answer turned out to be
harmless — the rollback was cleaning up two bad scripts wholesale, and the
audience fields just happened to be on board. Nobody ever rejected them on their
merits. Even better, they belonged to an agent that's been superseded anyway. So
what looked like a warning is actually a head start: someone already arrived at
the same design, and we can use it.

Second, the two existing audience descriptions in the database have been properly
migrated — the first live change this workstream has made. Reading them closely
turned out to matter: they weren't really descriptions of an audience at all,
they were instructions to copywriters that happened to mention the audience along
the way ("CTA language is 'Technical Discovery Call' sitewide"). A ranking system
can't do anything with that. So the new shape keeps three things separate: who
the reader is, how this domain differs from our other domains, and what all that
means for the copy. Only the first two feed the news selection. The migration
kept every sentence of the originals, invented nothing (where the originals had
no answer to "how does this differ from our other sites", the field is
deliberately empty rather than made up), and the old versions are preserved
underneath, so nothing is lost.

Third, the pools got sharper. With the owner's go-ahead to be more specific, the
big money group split into insurance, mortgages and lending, and savings and
investing — three genuinely different news streams — and the big
marketing group split into design, web, and digital marketing. Seventeen pools
now instead of thirteen, covering the same two thirds of the portfolio.

Fourth, the pilot is chosen: the domains that actually have readers. Thirty-one
domains, about four thousand views between them, touching fourteen of the
seventeen pools. Two things came out of picking it that need saying. A dozen of
the domains we'd ruled out of news feeds altogether turn out to carry real
traffic — including the portfolio's number two. Most are correctly excluded, but
a couple (the sports kit shop, the business supplies site) are arguable, and I'd
rather the owner spent ten minutes on that list than have a rule decide. And the
pilot is multilingual from day one — Romanian, Dutch, German and Spanish-market
domains are all in it — which turns the "news pages are hardcoded to English" bug
from a someday problem into a before-the-pilot problem.

The council seat idea is written up and ready but deliberately not created yet.
It has nothing to review until pooled selection exists, and it will give better
verdicts once the duplication measurement is running and it can point at numbers
instead of instinct.

---

## 2026-07-20 later — every live site now has its audience profile; digging into the mystery domains

Three more calls from the owner today, all acted on.

The mystery domains are being investigated. Several of the domains with the most
traffic have opaque names — zdec, komunikatif, makeitaquote and friends — and the
traffic has to be coming from somewhere: old links pointing at whatever used to
live there. A background research job is now working through the Internet
Archive's snapshots of nine of them to find out what they hosted, when they died,
and which old pages still attract visitors. That history will tell us what to
build on each one to inherit the traffic rather than squander it. Results will
land here when it finishes.

Languages are officially an exception, not a problem. Foreign-language domains
get feeds in their own language, the way relojistas already runs Spanish news
from Spanish sources. One consequence stays firm: the bug that hardcodes English
into news listings has to be fixed before the pilot, because the pilot includes
Romanian, Dutch, German and Spanish-market domains.

The domains with real readers get their own dedicated news sources on top of the
shared pool — the owner's phrase was "to give them the best chance in life".
Pleasingly, this costs nothing to build: it's the existing per-site machinery,
which our design deliberately left intact. The pool is the floor; dedicated
sources are the ceiling.

And the audience profiles are done — for everything that actually exists. All
eleven live sites now carry a structured profile: who the site is for, how it
differs from our sibling sites, each in the new three-part shape. The only three
sites that genuinely compete with each other — our AI services trio — got the
"how we differ" part written properly, each one defined against the other two:
one sells outcomes to small businesses, one sells architecture to enterprises,
one sells delivery track record to scale-ups. Everywhere else that field is
deliberately empty with a note saying which future site will make it necessary —
the twenty-plus vet domains being the clearest case.

The thirty-one pilot domains don't get profiles yet, on purpose. They aren't
sites yet, and writing a profile from nothing but a domain name is exactly the
mistake this whole design guards against. They'll get theirs at onboarding, when
the research and classification machinery has something real to say about them.

---

## 2026-07-20 evening — the mystery domains gave up their secrets, and two of them raise questions

The archive dig on the nine opaque traffic domains is done (the research job hit
a usage limit partway through its checking phase, so I finished the last two
domains by hand and marked clearly which findings are triple-checked and which
are single-source). The full write-up is in the research file alongside this one.
The headlines:

**The good inheritances.** The business-supplies domain shadows a real Surrey
office-supplies company whose website died in 2024 leaving customers with
nowhere to go — that's purchase-shaped demand, and it moves the domain into the
business pool. The komunikatif domain was a genuine Indonesian regional news
site for Central Java, publishing until 2023 — the strongest news-shaped legacy
of the lot, and a perfect fit for the language-exception rule. The makeitaquote
domain gets its traffic from people guessing the web address of a Discord bot
installed in over a million servers — they arrive wanting to turn text into a
quote image, and we build interactive tools for a living; that one should be a
tool, not a news site. The old fax service domain still gets people hitting its
dead login page — "how to send a fax online" content would serve exactly what
they came for.

**The warning.** The biggest number of the lot — zdec, 409 views — is the one to
distrust. It was a Chinese industrial company site that got hacked years ago and
stuffed with casino spam links, with backlinks literally offered for sale. Its
traffic is the residue of that spam network, not real demand. Big numbers aren't
automatically good news.

**The questions I need answered.** Four of the nine aren't what the list says
they are. Two of them (bigotime, buysportskit) currently redirect to
domain-for-sale pages — if we own them, why are they listed for sale? And two
(nanangmrk, ijih) are serving live content right now — nanangmrk is a thriving
Indonesian networking-tutorials site fed by a YouTube channel with half a
million subscribers. If those are ours, who set them up? If they're not, they
shouldn't be on our list. Nothing gets built on those four until that's
straightened out. One measurement lesson came free: nanangmrk looked empty to
our counting because it blocks robots — some of our "empty" domains may not be.

**One line we must hold.** Two of these domains shadow real businesses — one
dead, one still trading under a different address. Inheriting their *visitors*
is fair game; inheriting their *identity* is not. Anything built there serves
the arriving demand under our own name, never theirs.
