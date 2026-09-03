# Where we are — lendzy.co.uk

The owner's running log. Plain prose, append-only, newest at the bottom. Anyone may add; nobody
rewrites what is already here. Corrections go **below** as dated entries, never as edits.

---

**2026-09-02 — the lane opens, and the first day's findings.**

Lendzy now has its own lane. Until today it didn't have one: it was built back on 2 August as the
very first site the framework made end to end, and since then four different threads have each
worked on a piece of it and moved on. Nothing was being neglected exactly, but nobody was holding
the whole thing. That's what changed today.

The first thing worth saying is that lendzy is healthy. The site serves, the pages are real, and
the five calculators I checked all work. There was one live problem and it turned out to be quite a
neat one.

**Three of the nine calculators were in a strange state: perfectly fine to a visitor, and recorded
in our database as never having been built.** Both of those things were true at once, and both of
our checks were reading correctly — they were just asking different questions. If you asked the
website, the pages were there. If you asked our records, they had never been published.

The cause goes back to the site's birth. When lendzy was built on 2 August, three of its
calculators were stored in a way that predates how we do tools now — the page had a chunk of
finished HTML attached to it, but nothing saying *which* component that HTML was. The other six
calculators were rebuilt properly later, in mid and late August, and they're fine. The three that
weren't rebuilt have been quietly failing every time the system tried to re-publish them ever
since, because the re-publisher looks up the component, finds nothing, and gives up. It gives up
loudly, which is correct behaviour and a fix somebody made deliberately — but nothing was listening
for that particular noise.

The reason it never healed itself is a bit circular: the page is marked "needs rebuilding", so it
gets picked up and rebuilt, and the rebuild fails the same way, so it stays marked "needs
rebuilding". Six attempts since 25 August, all of which report as completed.

That single fault explains all three of the things you asked about. The **47 internal links** are
not broken — every one of them points at a page that loads fine. They were flagged because the
flagger asks our records rather than the website. Fix the record and all 47 clear themselves
without anyone editing a link. The same fault is also why those three calculators are **missing
from the sitemap**, so search engines can't see them.

So the repair is to give those three calculators the component record they never got — **using the
HTML that is already live, not by generating new ones**. You said keep the tools if they're
working, and they are working, so nothing about them changes. I checked that this is safe: the
stored HTML for all three is self-contained, with no template placeholders in it at all, so
adopting it as-is loses nothing. And I've written down the number of input boxes each one has today
(3, 1 and 2) so that after the repair we can prove we haven't quietly swapped a working calculator
for a new one.

I've put the root cause through our diagnosis loop rather than just asserting it, because it's the
kind of claim that other threads would then build on. I'll record the verdict when it lands, including
if it disagrees with me.

**On the FCA handbook — the big one.**

I've started on this and I want to flag one thing early, because it changes the design.

The good news is that most of what you're asking for already exists. We have a mechanism that
stores a fact along with the exact web page and the exact sentence it came from, and re-fetches
that page **every day** to check the sentence is still there. That is your "check with their online
version each time", already built and already running. It just has nothing to check on lendzy,
because lendzy has no facts registered at all — it is one of five finance sites in that state. So
the single highest-value thing is unglamorous: go through lendzy's financial claims and register
each one against the specific handbook rule it comes from. That needs no new code and it starts the
daily checking immediately.

The FCA handbook is fetchable — I pulled the consumer credit cost-cap section today, 477KB, no
login needed, with the rule numbers right there in the page.

**But here's the thing that worried me, and it's the reason I'd want us to be careful.** The FCA's
handbook site returns "success" for *every* address you ask it for, including rules that don't
exist. I asked it for an invented rule and it cheerfully returned a page. If we build something
that trusts that, we would end up with a system that says "checked against the FCA handbook" while
happily verifying facts against rules that were never written — which is a more sophisticated
version of exactly the problem we had before. The way to tell a real rule from a fake one is the
page title, and every check we build will have to ask it for a rule we know doesn't exist, in the
same run, to prove it can still tell the difference.

There's a second thing to design around: our daily checker currently fetches one page per fact with
no pacing at all. That's fine when the whole estate has 39 facts. If lendzy alone cites dozens of
handbook rules, we'd be hammering the FCA's website every morning. We should sort the pacing out
before we scale up, not after they block us.

For the local copy you asked about: we have a good precedent — the Companies House pipeline pulls a
whole external register into our own database, deliberately throttled to a small fraction of what
we're allowed, refreshed on a schedule and then queried locally. I'd build the handbook mirror the
same way, as its own thing, rather than trying to stuff a handbook into the facts table.

I've been in touch with the claims-verification thread, since that's their area, and they've come
back with all of this confirmed and the boundaries drawn. Their view and mine agree: register the
facts first (no code, immediate benefit), then build the mirror and the "what changed in the
handbook this week" detection, which is the genuinely new part.

**One thing I want to be firm about, and I'd like you to push back if you disagree.** I don't think
we should put a sentence on the site saying we check against the handbook until the checking is
actually running and proven. The original problem wasn't that the sentence was wrong in spirit —
it's that it existed before the thing it described. I'd rather build the mechanism, prove it, and
then say something true and dated about it. So: mechanism first, sentence last, and the sentence is
your call when we get there.

**2026-09-02, later — we found two wrong rule references on the live site, and the checking idea has already paid for itself.**

I went through every lendzy page, pulled out every claim that cites an FCA rule, and checked each
one against the actual handbook text. Five are right. Two are wrong.

The site tells people their lender can't roll a loan over more than twice, and says that rule is
CONC 6.7.17. It isn't — 6.7.17 is the definitions section that explains what "refinance" means. The
actual rule is CONC 6.7.23. That one is arguable, because 6.7.17 does introduce the block of rules
that contains the limit, so you could call it imprecise rather than wrong.

The second one isn't arguable. The site tells people their lender only gets two attempts to take a
card payment, and says that's CONC 6.7.23 — which is the rollover rule I just mentioned. The card
payment rule is CONC 7.6.12, in a different chapter entirely.

**Both of the things the site actually tells people are true.** Two rollovers, two card attempts —
correct. It's only the rule numbers that are wrong, and they're wrong in a shifted pattern, as if
they slipped by one when they were written.

I want to be clear about why this matters more than a typo would. This is a site telling people in
financial trouble which law protects them, and inviting them to check. Someone who follows the
reference we gave them lands on the wrong rule and finds nothing about their situation — and the
natural conclusion is that we made the protection up. The claim is true and the citation makes it
look false.

The other thing worth saying: **nothing we already had could have caught this.** All our claim
checking asks "is this claim supported?" — none of it asks "is the rule number you cited actually
the rule that says this?" That's a different question, and it only exists on sites that cite law.

**I haven't changed the copy.** Those sentences are on live pages, and rewriting published wording
off the back of an automated finding is exactly the authority you held back in August. So it's your
call. My recommendation is that we fix both — the substance doesn't change at all, only the two
rule numbers — and that we do it through the framework rather than by hand.

**On the calculator repair.** The fix is written and committed but I haven't run it yet. I put the
root cause through our diagnosis loop for an independent check, and it came back inconclusive —
not disagreeing with me, but it spent its budget looking at pages belonging to other sites and one
of its own database queries failed, so it never actually examined the three pages in question. I've
recorded that as "not confirmed" rather than quietly treating no-disagreement as agreement. My own
evidence is good and I've written down exactly what it consists of, but it is my evidence rather
than an independent check, and you should know that.

The repair is also with our review council and that verdict hasn't come back yet either. I'll run
it once that lands. The repair refuses to run at all if anything I measured has changed in the
meantime, so a stale assumption can't do damage quietly.

**2026-09-02, evening — you ruled on all five; here's what happened with each.**

The citation fix is written and with the review council. Going through every layer that stores
those sentences turned up two things worth telling you. First, the wrong number was also in
lendzy's *writing instructions* — the notes our content system consults when it writes — so if
we'd only fixed the pages, the next rewrite would have put the error straight back. Both layers
are in the fix. Second, the wrong rollover number had already travelled: loancash runs a copy of
lendzy's rollover checker, copied while the error was in it, and it's serving the wrong rule
number today. Same sentence, our copying mechanism, so the fix covers it and I'm telling you
rather than asking — it felt inside the spirit of "fix both".

Your go-ahead on the better checker went to the claims-verification thread the same hour, and
your sentence ruling is now written down as the standing rule: nothing on any site claims
handbook-checking until the checking is real and that new checker is watching it regularly.

The mirror will start with just the chapters we actually cite, built so widening it later is a
configuration change rather than a redesign.

On the other four sites: I've sent the method and the tools to the two with live sessions —
loanzy.uk (whose thread also holds farmerinsurance) and loancalculator. **The one with nobody
home is loancash.co.uk** — no session at all. Farmerinsurance technically has no session of its
own either, but the loanzy thread owns its lane, so it's covered. Loancash is the gap, and it's
now also the site where we're fixing a wrong rule number, so it's the one I'd give a session to
first.

And noted on the new chassis within the hour — I'm holding both apply steps until it's rolled and
settled, because the re-render jobs the fixes queue shouldn't race a restart.

**2026-09-02, end of day — everything you asked for this morning is done or armed.**

The two wrong rule references are fixed and verified on the live pages of both sites, calculators
untouched. The three stuck calculators are properly registered, published, and stamped for the
first time in their lives. And lendzy now has its facts register: eight FCA rules, each with the
exact sentence it rests on, checked against the FCA's live handbook every morning from tomorrow —
plus five banned-phrase tripwires adapted from a sibling site, each tested against our own pages
first so they can't convict legitimate copy.

Two things will finish themselves overnight and I'll confirm both: the sitemap picks up the three
repaired calculators, and the 47 stale link reports clear as the system re-checks them against the
now-published pages. The claims thread is building the rule-number checker you approved; the
sentence stays off until that's running.

Worth saying at the end of a long day: the review council earned its keep repeatedly — it caught a
repair that would have been inert, an insert that would have crashed, and a pattern-seeding bug
that would have silently disarmed the tripwires. Every catch is now a guard or a test, not a memory.

**2026-09-03, morning — the checking is running. This morning it ran for the first time, on its
own schedule, and passed.**

Just after nine this morning the daily checker made its first pass over lendzy's register: all
eight facts re-fetched from the FCA's live handbook, every quoted sentence found exactly where we
said it was, nothing flagged. So the thing you asked for on Tuesday — "check all financial facts
against the FCA handbook" — is now something the system does every morning without anyone asking,
and this morning it said: all true.

Worth being precise about why this morning's quiet means something, because quiet usually doesn't:
the pass demonstrably ran (all eight facts carry today's date), so the absence of alarms is a
verified clean rather than a switched-off check. And the daily writer handled our custom fields —
the rule numbers, the correction records, the banned-phrase tripwires — without losing any of
them, which was the one thing three separate reviewers worried about.

The sitemap also caught up overnight: all nine calculators now listed, thirty pages in total.

One item left from Tuesday: the 47 stale link reports. They turned out to have missed being
cleared yesterday by forty-four seconds — the daily sweep ran just before the repaired pages got
their stamps. Today's sweep, mid-afternoon, should clear them; I'm watching it.
