# apis.uk — where we are

Plain-prose running log for the owner. Append only; newest at the bottom.

---

**2026-08-22 — the bees page, and the good news about the API.**

You asked for a page about bees on the apis.uk home page, without disturbing the
tools API that runs on the same domain. I took the second half first, because
that is the half that can break something that is currently working.

The short version: **nothing needs to change in DNS at all, and the API was never
really at risk from the DNS anyway.** But there *is* a real way to kill it by
accident, it is not the one everyone has been writing down, and I have now written
the real one down.

Three of our own documents say that when the bees page exists we will need to
"repoint the apex at its hosting — one record swap". That was true when it was
written and it is not true now. The apex of apis.uk is *already* being served by
the same Cloudflare worker that serves all the portfolio sites. What happened is
that a worker "route" was added for the bare domain at some point, and a route
overrides where DNS points — Cloudflare intercepts the request at its own edge
before it ever goes looking for the origin server. So the DNS entry for the bare
domain still says "send this to the island machine", and that instruction is now
simply never consulted. The upshot is that putting the bees page in the usual
place publishes it, with no zone changes whatsoever. The safest possible change
to the API's DNS is the one we are making: none.

I did not take that from the configuration files, because the configuration files
are exactly what misled the earlier documents. I asked the four hostnames
directly, and they give themselves away by *how* they fail. All four return "not
found", but the bare domain and the www one return it in the worker's words,
while the API hostname and a made-up subdomain return it with an empty body,
which is the island machine's way of saying it. Same status code, four names,
three different servers, all distinguishable.

**The thing that could actually kill the API.** It is not the DNS records — those
are per-name and independent, so the bare domain and the API subdomain were never
going to interfere with each other. It is the *worker route pattern*. Our routes
today name the bare domain and www specifically. If anyone ever adds a wildcard
route — the "everything under this domain" form — it would swallow the API
hostname too, hand it to the portfolio worker, which would look for a web page
that does not exist and return "not found". The API would be dead, no DNS record
would have changed, nothing would look wrong, and our own tidy-up script would
report success. Twenty-four other domains already have that wildcard route, quite
correctly, because they do not have an API living on a subdomain. This one does.
I have filed that as a landmine so the next person to run a "give every domain the
standard treatment" sweep is warned before they run it, not after.

Before touching anything I also checked the API was alive, properly rather than
lazily. Worth knowing: the API's front door returns "404 not found" and always
has — that is just how it is configured, and someone has previously read that as
proof it was dead and been wrong. The honest test is to call a real endpoint, so I
did, and it answered correctly. That is now the before-and-after control for
anything done to this domain.

**On the page itself.** You chose a home page only, as a personal enthusiast page
rather than a beekeeping guide or a conservation campaign, and I have set it up
that way and handed it to the build pipeline. It writes the content, not me —
that is the rule and it is a good one. I gave it a brief explaining that the
domain name is the joke (Apis is the genus the honey bee belongs to, so a domain
called apis.uk ought to be about bees), that the visitor is someone curious who
knows nothing much about bees, and that the page sells nothing, collects nothing
and asks nothing of anybody.

I also constrained it to build *only* the home page. Left alone the pipeline
would plan a whole site — an about page, a contact page, guides and so on — so
there is a specific instruction in place that says build one page and do not
invent others.

**One thing I spent real care on, which I think is the right call but you should
know about.** Bees are a subject made almost entirely of famous numbers: the
share of our food that depends on pollinators, the two million flowers in a jar
of honey, the miles flown, the tens of thousands of bees in a hive, the percentage
declines. Every one of them is repeated everywhere and none of them has been
checked by us. There is also a very well-known quote about having four years left
to live that is confidently attributed to Einstein and that he did not say.

So I have set the page up to assert **no quantities at all** — no counts,
percentages, distances, weights, temperatures or lifespans, whether in digits or
spelled out. That sounds like a severe restriction and I do not think it is,
because none of what makes bees genuinely remarkable needs a number: a returning
forager telling the others where she has been by dancing on the comb, direction
read against gravity; wax worked into a shape that wastes nothing; a colony
dividing its work by age so a bee's job changes as she ages; the fact that most
bees are not honey bees at all but solitary insects nobody notices. If we later
want a specific figure on the page, the way to get it is to look it up properly,
record where it came from, and then it is allowed. The friction is deliberate.

Worth saying plainly: I tested that restriction rather than assuming it worked,
and **it failed three times before it passed.** One rule was written in a way that
looked correct, was technically valid, and would never have matched anything — the
kind of error that reports "all clean" for ever. Another missed "two million
flowers" entirely, which is the most repeated bee statistic there is, because it
was only looking for digits sitting directly against the word. Those are now
fixed and there is a test holding them in place.

**Where it stands.** The site is seeded and the build was submitted at lunchtime.
These take a while to come through — anything from a quarter of an hour to half
an hour just to start under normal load — so the next step is simply watching it
build, then checking the page that comes out and confirming, separately, that the
API still answers. Those are two independent facts and I will check them as two.

**One thing you may want to decide.** The bare domain used to be part of a traffic
probe — a passive listener recording who was still asking for apis.uk and what
they wanted, which was due a read on the 8th of August and has not had one. The
apex arm of that probe has in fact been silently inactive since the worker route
was added, so the bees page is not taking anything away that was still working.
But the wildcard arm is still listening on every other subdomain, and the log has
never been read. It might be worth reading before we lose interest in it — that is
a separate small job and I have not done it.
