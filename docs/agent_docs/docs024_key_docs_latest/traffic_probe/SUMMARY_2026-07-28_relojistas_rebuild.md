# relojistas.com — where we are, 28 July 2026

*Series: 07-19 · 07-24 · 07-25 · 07-26 · **07-28**. Written to be read aloud.*

---

## What we're trying to do

Take a dead Spanish watch forum — a domain with years of history and nothing left on it — and
turn it into a living Spanish-language watch news portal that runs itself. Keep faith with the
people who still pull the old feed. Find out whether anyone is actually out there. And leave the
domain worth more than we found it.

## Where we've come from

The forum died. The domain sat parked. In mid-July we found that the old RSS feed was still
being fetched about 136 times a day and returning 404 to every one of them — people and machines
still knocking on a door that had been bricked up.

We rebuilt the site as a news portal, reactivated the feed at its original address, and by the
24th it was publishing on its own with nobody watching. What remained was a list of small
things: three variants of the old feed URL still 404ing, a search feature built but not
switched on, and a way to tell real visitors apart from Cloudflare's servers. All of it needed
one session on the box, and the handover said only the owner could do that.

## What we've done — and the hoops, since you asked

**The first thing we found was that we had been lying to ourselves.** The handover certified the
site healthy, and one of its proofs was a check that searched the homepage for two specific
broken links we had fixed the week before. Of course it found none. It would have found none on
a page made entirely of broken links. Behind that reassuring zero, the main button on the front
page — *"Leer las últimas noticias"* — had been pointing at a page that does not exist. In
English. On a Spanish site. Since at least the 18th.

That set the tone. **Almost everything that followed was a check that couldn't fail, or a fix
that fixed the wrong thing.**

Then you asked for the contact route to come off, and it turned into five separate problems
wearing one coat. There was a contact page whose form submitted to nowhere — it discarded every
message while promising *"we read all of them"*. There was the dead button. There was a link in
the footer of every page. There was an empty "Contact" heading. And there was the address
itself, which turned out to be feeding two different parts of the site from a single database
column. Fixing any one of them left the other four serving.

**Three of those took me a wrong turn each, and the wrong turns are the useful part.**

I tried to remove a button by deleting its text, on the reasonable theory that a button with no
words wouldn't appear. The platform filled the gap with its own defaults and published *"Get
Started"* — in English, pointing at a 404. **You cannot remove a button here by taking its
words away; the absence is what triggers the default.** Two pages were briefly worse than when
I started.

I archived the contact page and expected it to vanish from the menus. It didn't, because the
navigation is kept in its own table that archiving never touches — so for about ten minutes
every page on the site advertised a link to a page I had just deleted.

And then the footer simply would not change, through three separate theories and eighteen page
rebuilds. The answer was that the header and footer aren't rebuilt when a page is rebuilt at
all — they're stored once and reused for months. relojistas' had been frozen since the 16th.
Worse, when I finally rebuilt it, it *still* didn't change, because the site had been rendering
its footer from a component that was switched off — the code that picks one ignores whether it's
active and takes whichever name sorts first alphabetically. That is true of every site we run.

**Then I broke the site's publishing pipeline.** The box session — which, it turned out, never
needed you at all, because I could reach the machine perfectly well and nobody had tried in four
days — went cleanly, until I ran the setup script without one parameter. The runbook mentioned it
in its header and omitted it from the section I was following. Dropping it handed the web
directory to the wrong user, and every automated deploy failed until you pasted me the error.
**Nothing I had checked would have caught it**: every test I ran was me *reading* the public
site, and what I had broken was somebody else *writing* to it. The site looked perfect because
the last good publish was still sitting there.

**What we got for all that.** The three legacy feed addresses now answer. Search is live and
returns real results — five for *tourbillon*, three for *cronógrafo*, a proper empty page for
something that isn't there. The site has a favicon for the first time since launch, drawn from
the gear in its own logo. Every internal link on all nineteen pages now resolves — twenty-five
distinct addresses, nothing broken. And we can finally see real visitor addresses instead of
Cloudflare's, which is what everything below depends on.

One item on that list turned out to be impossible rather than pending: the visitor-intent
collector was waiting on an endpoint that has never existed in the engine. The source calls it
*"the future /events collector"*. No box session could ever have finished it, and the runbook had
been telling threads to book one.

## Where we are now

**The site is healthy and running itself.** Feed rebuilding daily without help, no broken links,
no contact route, search working, the whole legacy address space answering.

**And the first day of real visitor data has changed the picture completely.** Now that we can
see who is knocking, the answer is: **nine out of every ten requests to this domain are a 404 for
the forum that died.** Over thirty-three hours — 243,000 requests, 225,000 of them errors, 946
successes.

Most of it is a machine sweep: 184,000 requests a day, from 1,409 different addresses, walking
25,000 old image URLs harvested out of somebody's archived copy of the forum. It is not an
audience and never will be.

**The part that matters is what it's doing to the crawlers we actually want.** Google, Bing,
Apple and the rest are here — and 78% of everything they fetch is the dead forum. Seven requests
in thirty-three hours reached the news, the guides or the glossary. A site publishing curated
Spanish watch news every single day is, as far as search is concerned, very nearly invisible —
not because the writing is poor, but because the crawlers are spending their entire visit on a
corpse.

Two things underneath that nobody had looked at: we have no sitemap, and our `robots.txt` isn't
ours at all — it's Cloudflare's default, and it blocks every AI crawler outright. That may be
what you want. It isn't what you chose.

**On the for-sale block**, the copy is written in Spanish and the machinery to carry it is built.
That took its own detour: the block existed only in English, and the wording you approved doesn't
survive translation — the natural Spanish for "for sale" is the classified-ad register, and this
site's own style rules ban shop language by name. So it was written against the site's voice
rather than translated into it.

## Where we're going

**The next move is the one the crawler data just handed us, and it's the same trick that worked
on the feed.** Tell the machines the dead forum is gone for good — a "410 Gone" rather than a
"404 not found", which is the only answer that makes a crawler stop asking — and publish a
sitemap pointing at the living site. That should hand the search engines' attention back to
content that actually exists. I've said in advance what I expect it to do, so we can check
whether it did.

Then two decisions that are yours, not mine.

**The AI crawlers.** Right now they're all refused, by a setting we inherited rather than picked.
Being read and cited by an AI answer is exposure; being training material is not. Those are
separable, and at the moment we're refusing both.

**And the price.** You've said the twelve thousand is aspirational — that we build the domain up
to meet it rather than trimming the number to fit. I think that's right, and today's evidence
supports it: a domain that 1,409 machines still sweep daily, and that a live Spanish watch forum
still links to, has a residual standing you don't get from a fresh registration. What it doesn't
yet have is a search presence, and that is fixable and cheap. **Worth saying plainly, though:
those are machines, not readers.** The human number is small. If the twelve thousand is a target
to build toward, the thing to build is the audience — and the first step is being findable.
