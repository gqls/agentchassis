# Where we are — webdesign.uk, selling web design and build

Plain-prose running log, append-only, newest at the bottom.

---

## 2026-07-28 — first session, planning only

You asked for thinking and planning, so nothing has been built and nothing has
been committed to a design. What follows is what I found and what I think.

**The first thing worth saying is that you already planned this.** Last night's
buying-design plan for webdesign.co.uk has a section near the end recording your
next build — the website creation form, the tools-api, and standing up a copy of
the chassis in its own cluster with its own database. It was written down
deliberately as "recorded, not started". So this is not a new idea being invented;
it is that section being picked up.

**Most of what you want already exists, in pieces, and has taken real money.**
idea.uk is the whole shape of this product already working: a small Go program on
a VM, a page, a form, Stripe, an order, a human check, an email out. Its payment
code is behind a clean interface, it verifies the webhook signature properly, and
it has a fake payment provider for testing — that is the hard, boring, easy-to-get-
wrong half of a shop, and it is done and has survived a real sale. tools-api is
the other half: a public endpoint with rate limiting and input caps that talks to
nobody in your cluster. Between them, the shopfront is largely a copying job.

**The bit that genuinely does not exist is the one you named.** Nothing outside
the cluster can start a build. There *is* an internal admin API that can trigger a
pipeline, but it is behind an admin login and it is meant for your dashboard, not
for the public. So you are right about the gap.

**Where I want to push back, gently, is on what to do about it.** You said a
dedicated cluster so the existing one doesn't get hacked. I think there are two
different worries inside that, and only one of them needs a second cluster.

The security worry — a public website reaching into your production cluster — has
a much cheaper answer, and it is one you already chose three days ago. When the VM
estate plan asked whether the framework should push config out to the island, you
ruled that the island should pull, outbound only, so the cluster never holds a
credential to it. Turn that round and it solves this: the webdesign.uk box holds
the orders, and the cluster reaches *out* to collect the paid ones. The box never
dials in and holds no cluster credential. If someone owns that box completely,
they have the box and the orders on it, and no route to your database. That is a
single outbound HTTP call, not a second Kubernetes cluster.

But there is a second worry hiding inside yours, and I think it is the real one,
and it *is* worth money. When we build a site for a customer we scrape whatever
domain they typed, feed that content into the model, and write the results into
the same database that serves your fourteen live sites. That is untrusted content
influencing agents with write access to production. A firewall does nothing about
it. So if you want isolation — and I think eventually you should — that is the
reason to give, and it changes what needs isolating: the database and the agents,
not the front door.

The practical suggestion is to take that decision later, but not much later: after
the shop is up and money is coming in, and before the first paid build runs
against somebody's scraped website. Because of the pull design, deferring it costs
nothing — the collector points at whichever cluster exists.

**One change I'd like to argue for in the product itself.** You have the briefing
questionnaire as the optional better route. I think it has to be compulsory before
we take money and build. Not for lead capture — because it is the thing that stops
us inventing the customer's telephone number. This system has published a
fabricated contact address before, for hours, because the check that was supposed
to catch it quietly passes when a site has no email at all. And it has invented
statistics twice. A demo we build for ourselves getting something wrong is
embarrassing; a site we sell to a real business with an invented phone number or
an invented accreditation on it is their liability, published under their name,
paid for. Requiring the questionnaire turns that from a hope into a rule, and it
gives the page something true and unusual to say: we will not write a word about
your business that you haven't told us or we haven't read on your site.

The free "type your domain in" teaser still works — it just gets built only from
what we can actually read on the site that's there, and leaves the contact details
blank on purpose rather than by accident.

**On the chat box** — a genuine chat with a model is a much larger and riskier
thing than it looks, because it is unbounded input and unbounded cost from
strangers. For the job it's doing, which is collecting one domain name, a form
that asks one question at a time gives you the same conversational feel for a
fraction of the exposure. Worth a decision either way.

**One thing I did not go looking for and think you should see.** The buying-design
positioning says we run this system across about a thousand sites. The database
says thirty-two rows, of which fourteen are deployed sites and seventeen are empty
placeholders. Tracing it back, the thousand comes from the architecture threads,
where it is a *target* used to argue about how the code should scale — a perfectly
proper use — and it has drifted into outward-facing copy as a present-tense claim.
It may well be true about domains you own, which is a different sentence. It needs
pinning to its noun before it is published, and webdesign.uk makes that urgent
rather than academic: the whole pitch is that we are the ones who tell the truth
about AI web builds, and that audience checks. The honest version is at least as
good — fourteen production sites, run by the system, on the record including where
it went wrong.

**What I'd do next, if you agree with the shape:** put the shop up on a VM with
Stripe in test mode and nothing behind it, and see whether anyone types a domain
in and goes through with it. idea.uk did exactly that with a fake door before it
built the engine. Then the free teaser, then real money with me or another session
doing the build by hand — which is the "next day or so" on the page, honestly. All
three of those are a working business and none of them touch the platform code, so
none of them need the council, a new cluster, or a decision you'd be stuck with.

The full plan, with the phases and the seven open questions, is in
`PLAN_2026-07-28_webdesign_uk_build_service.md`.

---

## 2026-07-28 — your three rulings, and what they change

You ruled on all three the same evening. Recording what each one moves, because
two of them went against what I'd recommended and that has consequences worth
seeing rather than discovering later.

**The trust boundary: yes, ship it now, decide isolation at P3.** That's now a
constraint rather than a preference, and the thing to protect is the direction of
travel — the box never dials into the cluster and never holds a credential to it.
Any later shortcut that reverses that gives back the entire security argument for
free, so it's worth being unfashionably strict about. The one thing I've written
down so it doesn't get skipped: the isolation decision has a specific trigger —
the first paid build that scrapes a domain we don't own — and whoever runs that
build owes the decision. It would be very easy, with money coming in by then, for
that moment to pass unnoticed.

**The questionnaire stays optional.** Fine, and I've kept my argument in the plan
rather than deleting it, because the risk it describes doesn't disappear with the
gate — it just has to live somewhere else. The good news is that "somewhere else"
is stronger than the gate was. A gate is a rule someone can be talked out of; what
I've put in its place is structural. We simply don't generate a contact block at
all unless you gave us the details — not a placeholder that might slip through,
absent. The bug that let a fabricated address reach production was a *check*
failing open, and a field that is never written can't fail open. Alongside that,
the platform already has the right mechanism: seed the site's evidence base before
the first page is written, built from the two things we genuinely have — what the
customer told the chat, and what we read on their existing site, each attributed.

There is one condition attached, and I'd rather flag it now than have it arrive as
a surprise. The optional questionnaire is safe *while a human looks at every site
before it goes out*. P5 is the phase that automates release. When release stops
passing a human, this decision needs re-opening — either the gate comes back, or
those two controls have to be shown to be doing the work by themselves. I've
written that into the phase itself so the step that removes the safety net is the
step that notices.

**A real chat, then.** You were right that it buys something my version didn't,
and I'd missed it: a real conversation can *do* the briefing. It can ask for the
phone number and the services naturally, while someone is engaged, which is a far
better collection mechanism than an optional form afterwards. So your two rulings
work together better than either does alone — the chat is much less likely to
leave us with an empty brief than a stepped form plus an optional questionnaire
would have been.

The cost of it is that it moves work earlier. A fake door with a form costs
nothing to run and could go up bare. A fake door with a real chat spends money on
every visitor, including hostile ones, so the limits — per-visitor rate, a cap on
how long one conversation can run, and above all a per-day ceiling on total spend
— have to ship *with* it rather than after. The first phase is therefore a bigger
piece of work than it looked this afternoon, and it's the one not to rush.

The one part I'd think about hardest is not a setting. What someone types into
that chat ends up in the brief, and the brief is read by the agents that build the
site. Someone typing "ignore your instructions and…" is writing into a document
our own system will later read. The answer is that the transcript enters the build
as quoted customer statements in a named field, never as loose prose pasted into a
prompt — which is the same discipline as the evidence base, arrived at from a
different direction.

**Still needing you:** the price (unblocked as soon as I measure what a build
actually costs), whether the preview sites sit on `*.preview.webdesign.uk` or a
subdomain of another of your domains, and the thousand-sites figure.

---

## 2026-07-29 — Fable 5, and the offer takes shape

**On Fable 5, first, because the answer has a trap in it.** I checked what the
fleet is actually running before changing anything: over the last four days it
made about 1,900 model calls, and **not one of them was Fable**. It's almost all
Sonnet 5, with some Sonnet 4.6 and a little Gemini and Mistral. So "we're using
Fable 5" is true of this session and true of where you want the builds to go, but
it is not yet true of the system — which matters, because pointing a lane at a new
model is a live database change that takes effect immediately, with no rebuild to
slow it down.

Fable is Anthropic's most capable model and it costs $10 per million tokens in and
$50 out — twice Opus 5, and about five times what the fleet pays for Sonnet 5
today on its introductory rate. For a site we're selling at a proper price that's
noise, and what Fable is *for* reads like a description of this exact job: long
autonomous runs, getting a well-specified thing right first time, checking its own
work. So I've written the builds onto Fable as you asked.

Three things have to happen before we point anything at it, and I've put them at
the front of the first phase. Fable refuses to run at all for an organisation on
zero data retention — every request fails, and the error looks like a bad request
rather than a settings problem, which is exactly how you lose an afternoon. It
also rejects several of the parameters our chassis currently sets on model calls,
so swapping the model is not just a config edit if the calling code passes them —
and we know it passes some, because every council seat sets a token limit. And
we should measure what one real Fable build costs before we put a number on the
page. That measurement is now the only thing standing between us and a price.

**The offer.** Full sites at a proper price, money back if they don't want it,
and changes cost money afterwards — I think that's right, and I've written it up
with the joins made explicit, because the three parts only work together if
acceptance is the hinge.

Here's the shape. They pay, we build, the site goes up on the preview domain, and
they look at it. Until they accept it, the guarantee is the only thing in play:
if they want their money back they get all of it, the preview comes down, and
they keep nothing. The moment they accept, the guarantee ends and the fee model
starts. That gives us one clean line instead of two overlapping ones, and it
means "corrections carry a fee" never collides with "full money back" — they
simply can't both be live at the same time.

On the fee boundary I'd draw one line inside it, and this is the part I'd argue
for rather than just record. Charging for changes they want is obviously right —
that's your time, and your time is the genuinely scarce thing here. But a broken
link or an invented phone number is our defect, and I think those get fixed free,
indefinitely. Two reasons. The first is that the entire pitch — on webdesign.co.uk
and here — is that we're the ones who tell the truth about AI builds; "they
charged me to fix their own bug" is the cheapest possible thing for a competitor
to quote back at us. The second is more practical: it turns the anti-fabrication
work into something that protects money rather than just reputation. Every
invented detail that reaches a customer is now a free repair we owe, so the rule
about never generating a contact block we weren't given is defending margin.

One consequence worth flagging: the preview domain is no longer just where the
site goes. It's the mechanism the guarantee runs on — a refund is literally "the
preview comes down". So when you pick the short domain, it wants to be somewhere
a customer will believe a real deliverable lives, and it has to be a zone we
control the DNS for so we can issue the wildcard certificate. Nothing before the
third phase needs the name, so there's no rush.

**The thousand sites.** Taken as read, as you said — the item's closed. One note
for whoever writes the copy later: it's a forward-looking number now, so it has a
shelf life, and webdesign.uk's own claims should come from this project's
measurements rather than inheriting webdesign.co.uk's prose. That site describes
what we do; this one takes money for it.

**Also folded in, from another thread's bad afternoon.** The gauntlet lane just
found that its per-visitor rate limit had never worked: behind Cloudflare and a
reverse proxy, every request arrived looking like it came from the same address,
so all the visitors on earth shared one bucket. It only showed up because someone
counted distinct values — from one machine it looks perfectly healthy. We're
about to build the same shape of thing, and our per-visitor limit is what stops
one stranger spending the whole day's model budget, so I've written the trap into
the plan with the check that actually detects it.

**Still needing you:** the price (unblocked the moment we've measured a Fable
build), and the short domain when you've picked it.

---

## 2026-07-31 — the SSRF gap you asked about is now a fix, and it turned out to be live already

You asked me to write the SSRF code — server-side request forgery, the attack
where a service fetches a URL on someone's behalf and gets pointed at your own
internal network instead — and figure out how to make it reusable for other
domains. Both done, and along the way it stopped being a hypothetical for a
product we haven't built yet and became a fix to something already running.

Tracing the actual scrape pipeline, I found the real exposure: when the
platform scrapes a customer's page, it also pulls out the images that page
mentions and downloads them, unchecked, from the pod itself. A page could name
an image URL pointing at the cluster's own internal address book instead of a
real picture, and the pod would dutifully fetch it. That's not specific to
webdesign.uk — I measured it and ten different agent types across the fleet
already use this scraping path, so it's been sitting there as a live gap the
whole time, on ordinary site builds. I've filed it properly and fixed it in
the same piece of work.

The fix is a small, general-purpose package — a version of an HTTP client that
checks where a URL actually leads before connecting, and re-checks on every
redirect, so a URL that looks fine on the surface can't quietly send a fetch
somewhere private partway through. It sits alongside the platform's existing
"who's allowed to call us" guard, as the mirror of it — "what are we allowed
to call on someone else's behalf." I've wired it into the one place that was
genuinely exposed and left the rest of that file alone, and I've flagged one
thing it does *not* cover — a headless browser that navigates to a page
directly can't be protected the same way, and that's a separate piece of work
for later, not something I've quietly assumed is fine.

One small thing worth mentioning because it's exactly the kind of thing we
care about getting right: I wrote a comment in the new code making a specific
security claim, then wrote the test meant to prove it — and the test proved
me wrong instead. A caught mistake, fixed before it ever shipped, which is
the system working as intended rather than something to worry about.

It's gone through the platform's own review process and is committed. With
this done, the only thing left before this whole business idea is buildable
is measuring one complete site build end to end, which needs the actual
build-trigger machinery from a later phase — everything else this early
checklist was meant to answer (cost, whether Fable-5 will even run for us,
and now this) is in hand.

---

## 2026-07-31 (evening) — P4 is planned and proven; next thread builds the shop

Three things happened since the last note, and the middle one is the good news.

**The SSRF fix went live on your chassis roll.** I checked it properly rather
than assuming — grepped the running pods for a string the fix added, with a
control to prove the grep worked, on all three replicas rather than one. It's
there, so that bug is closed rather than sitting open with the code merged.

**P4 turned out to be about a tenth of the work we thought, and I proved it
rather than asserting it.** The machinery to take "a domain plus a brief" and
turn it into a real site build already exists in the platform and has been
quietly running every two minutes this whole time — it's just been idle, because
nothing has been putting anything into it since March. So P4 isn't building a
pipeline; it's putting one row into one table that already works.

I didn't want to take that on trust after four months of nobody using it, so I
ran a real test row through the live system. It worked — including the part that
mattered most, that the customer's brief travels through intact rather than
getting dropped somewhere. I did it in a way that couldn't accidentally start a
real site build and spend money: there's a specific condition that makes the
system skip a site, and I used it deliberately as a handbrake. Nothing ran,
nothing was spent, and I cleaned up afterwards back to exactly the numbers I
started with.

Worth mentioning because it nearly went wrong: **your fresh chassis rolled out
in the middle of the test**, and for about five minutes everything went silent
in a way that looks identical to "this is broken". It wasn't — there's a known
five-minute dead zone after a restart. I'd checked the timing before starting
precisely because you'd said a build was coming, so I recognised it rather than
writing the whole plan off as a failure. I've recorded that trap so nobody else
loses an afternoon to it.

**Next is P1 — the shop itself**, and I've written the handoff for a fresh
thread to pick it up. The short version of why P1 and not P4: P4 has nothing to
collect from until the shop exists, and the shop is the thing that tells us
whether anyone actually wants this. idea.uk is the cautionary tale sitting right
next to us — complete, working, live, and still no strangers buying.

The genuinely useful discovery for P1 is that the payment side is close to a
copying job. idea.uk's engine already does Stripe properly, stores orders
safely, and even handles the fiddly cases like a payment notification arriving
twice. It also already has a "domain" field on an order, which is the whole of
your intake. The new work is the chat, and the spend controls that have to ship
with it — a chat costs money on every visitor, including hostile ones, so those
aren't a later polish.

One thing I've written into both plans so it can't get lost: **the shop needs a
small "has the cluster collected this order yet?" marker built in from the
start.** It's about thirty lines, nothing uses it until P4, and adding it later
means surgery on a box that by then is taking real money.

**Still waiting on you:** the short preview domain, the price (unblocked now —
we know the cost), and DNS for webdesign.uk. None of them block starting P1.

---

**2026-07-31 (evening) — the DNS answer, and one of my own claims falling over**

You asked what you need to do about DNS for webdesign.uk and about the price, and
told me the preview domain is ugg2.com. Before answering I did the thing I should
have done three days ago and actually looked.

**webdesign.uk is already pointed.** I had been writing "DNS not pointed yet" in
the plan, the runbook and my own notes since the 28th, on the strength of your
remark that you hadn't done it. It's been sitting on Cloudflare's nameservers,
proxied, this whole time. That's my mistake, not yours — a second-hand claim about
infrastructure that one command would have checked, repeated for three days into
work that depended on it. I've logged it where we log those.

**But it's pointed at the wrong thing, and that turned out to be useful.** Type
webdesign.uk into a browser today and you get a raw error in developer-speak,
because it's wired into the same path that serves your twelve static sites and
there's simply no page there for it. The error is chatty, and what it chats about
is exactly the thing I'd written "I haven't checked this" against in the plan: how
a domain name turns into a file in storage. It's the domain name itself. So a
question I'd flagged as needing research answered itself by being visited.

**That matters for ugg2.com far more than it does for webdesign.uk.** The plan had
previews needing their own web server, a wildcard security certificate, an API
token, and renewal machinery — a few days' fiddly work. If the mapping really is
just "the address is the folder name", then previews need none of that: point
ugg2.com's subdomains at the path the static sites already use, and
`someclient.ugg2.com` serves itself. Certificates come free. Taking a preview down
for a refund becomes deleting some files rather than touching a server.

I want to be straight about how solid that is: I've seen it work for one address,
and the code that does it lives in Cloudflare, not in our repo, so I can't read it.
It's a strong hint, not a proven fact. **There's a ten-minute test** — point one
test subdomain at it, upload a file, see if it loads — and if it works, a chunk of
planned work simply disappears.

**What you actually need to do, in order:**

*ugg2.com* — it's delegated to Cloudflare, which is the important half, but the
only record on it still points at the registrar's parking page and nothing
answers. So: delete that record, and decide the mechanism (I'd run the ten-minute
test first). One trap worth naming — the preview records must be **proxied**, the
orange cloud, not grey. Grey means every customer gets a browser security warning
on the one page whose whole job is to make a stranger trust us enough to accept
and pay.

*webdesign.uk* — **nothing yet, and please don't pre-stage it.** It can't point at
P1's box until P1's box exists, and the way I'd recommend connecting it writes its
own DNS record and would overwrite anything set up in advance.

*One thing that isn't obvious and bites late* — order confirmation emails. If they
send from an @webdesign.uk address without the right records in place they go
straight to spam, and a confirmation for a four-figure purchase landing in spam is
the worst possible first impression. It's fifteen minutes of DNS, but it has to
happen before the first real sale, and customers will reply to that email, so
there needs to be a mailbox behind it.

**On the price — I got this wrong too, in a more interesting way.** I'd written
that the price was waiting on us measuring what a build actually costs in model
spend. That's backwards. You'd already ruled the price is quality-based, not
cost-plus, and I quietly reintroduced cost-plus by making the number wait for a
cost figure. Even on a deliberately pessimistic estimate the model spend is
somewhere near a hundred to two hundred dollars a site — at a four-figure price
that's a rounding error, and it stays a rounding error even if I'm wrong by
threefold. The measurement still matters, but for margin tracking and for deciding
how much free stuff we can give away, not for setting the price.

**So the price isn't blocked. My recommendation is £1,200, paid up front, fully
refundable until they accept it.** The reasoning: the customer will compare us to
a freelancer (£500–£3,000) rather than to Wix, because a human reviews it and the
money is guaranteed. At £1,200 you can afford the best part of a day on an awkward
one and still be well ahead — and your time is the genuinely scarce ingredient
here, not compute. And the guarantee only means anything at a price where a refund
hurts; "money back" on a £99 site is noise. If you want fewer and better customers,
£1,950 does that and the price itself filters out the ones who'd have been trouble.

**Two smaller things I need from you eventually, neither blocking:** a number for
the paid corrections after acceptance (I'd suggest £150 a change or £600 a day —
without one, "their changes are paid" isn't really enforceable), and whether the
price is quoted with or without VAT, which matters because business buyers reclaim
it and consumers don't.

Worth knowing: P1 runs Stripe in test mode, so nobody pays this. That makes the
number safe to commit to now — but it should be the price you actually mean to
charge, because the whole point of P1 is finding out whether anyone will.

**One thing I found that isn't about webdesign at all, and you should know.** While
checking what your VM setup normally looks like, I found that **idea.uk is not
behind Cloudflare** — it's on Hetzner's nameservers pointing straight at the
machine. Its own runbook says it is, and has a whole section of planned security
work built on that assumption. Two consequences: the rate limiting on that box is
actually fine, so that work isn't needed; but there's no firewall, no bot
protection and no DDoS layer in front of a live site taking real money, which the
runbook assumed was there. Also, instructions in it that say "purge the Cloudflare
cache" do nothing at all. Relojistas *is* behind Cloudflare, so the two boxes
aren't the same shape even though the docs treat them as one pattern. I've written
it down where the next person to touch that box will see it, and left a dated
correction in their runbook rather than rewriting someone else's file.

---

**2026-07-31 (late) — the A record you added, and what it's doing**

You pointed both domains at `116.203.204.115` and took the Worker routes off.
That address is idea.uk's live box, and it only knows how to serve one site — so
right now **webdesign.uk, ugg2.com and idea.uk all serve the exact same page**,
byte for byte. Anyone typing webdesign.uk gets idea.uk's shop.

That's worth fixing today rather than tomorrow, because the shop *works*. The
engine never checks which domain it was asked for, so someone could start buying
an idea.uk report from webdesign.uk and get bounced to idea.uk halfway through
paying. I haven't tested that end to end — doing so would create a real order and
a real Stripe session — but the page is fully live and there's nothing in the code
stopping it.

**The fix takes three minutes and doesn't touch the live box.** In Cloudflare, add
a redirect rule sending webdesign.uk to webdesign.co.uk — real content, no
confusion, and you delete it when P1 launches. Make it a **temporary** redirect,
not permanent; browsers cache permanent ones almost forever and you'd be fighting
it later. For ugg2.com just delete the A record — it's a delivery address, not
somewhere people should be typing yet.

**One genuinely good thing came out of this.** Because ugg2.com now sits behind
Cloudflare, the hardest piece of the preview plan has quietly disappeared.
Cloudflare issues certificates for subdomains automatically, so the wildcard
certificate, the challenge process, the API token and the renewal machinery I'd
planned are all unnecessary. That work was in the plan because I copied the
pattern from idea.uk — which *isn't* behind Cloudflare and therefore has to do all
of it by hand. When the preview domain turned out to be a Cloudflare domain, the
reason for that work went away and I hadn't noticed.

**On putting webdesign.uk on idea.uk's box — I'd rather we didn't**, and one of
the reasons is our own doing. The setup script we'd use to build the new box
resets the firewall as part of taking a machine over; your own runbook already
warns never to run it against the live idea.uk box. Beyond that, the shopfront is
an open chat box that spends money on every visitor including hostile ones, and
that shouldn't share a disk with your Stripe keys and your live order file. A
separate machine is about £7 a month.

**Separately, idea.uk itself needs a decision from you.** It isn't behind
Cloudflare at all — its own runbook says it is, and plans security work on that
belief. So either leave it direct, accept there's no firewall or bot protection in
front of a live earning site, and delete that planned work; or move it behind
Cloudflare and get all three. If you do move it, two things stop being optional:
telling nginx how to read the real visitor address, or the rate limiting silently
becomes useless while still looking like it works; and blocking direct access to
the machine's own address, or attackers simply walk around the protection. I'd
move it — but on its own, not mixed into this.

---

**2026-07-31 (late, second) — bucket or box, and a product question hiding in it**

You asked whether ugg2.com should point at the storage bucket or at a machine, and
guessed it depends on what kind of site got built. That's a fair reading of your
estate — you genuinely have both — but for what we're selling the answer collapses
to one, and the code settles it rather than my preference.

**Everything the system builds is static.** Where a page has a contact form, the
builder rewrites it to open the visitor's email program rather than submit to a
server, and there's a comment in that code saying plainly that no such backend has
ever existed. The two sites of yours that live on machines — idea.uk and
relojistas — are there because they carry their own custom software, which isn't
something a customer buys.

**So ugg2.com points at the bucket, and there is no preview machine at all.** That
kills the last piece of the preview plan: no server to build, no certificates, no
web server config, nothing to keep patched. It also keeps the shopfront machine
cleanly separate, because it never has to serve customer sites.

**But that turned up something you should decide on, and it's about the offer
rather than the plumbing.** The contact form on a delivered site opens the
visitor's email program. That's a deliberate and defensible choice — the
alternative the code refuses to do is invent an address that nobody reads, which
looks fixed while silently losing every message. Still: someone paying four
figures may well click it, see their mail app open, and call that broken. Under
the guarantee we've agreed, our defects are free forever — so if we haven't said
what they're getting, that's an unlimited free-repair claim waiting.

It needs one sentence in the offer: the contact form opens an email, and a form
that posts to a server is a paid extra. I'd rather say it plainly up front than
argue it after someone has paid. It's the only limitation that's invisible in a
screenshot and obvious on the first click.

**I've also written up the idea.uk situation as its own handoff** so whoever picks
that box up gets the whole thing in one place: what I measured, what it means for
the security section that's built on the wrong assumption, and the two sensible
options with the traps in each. I changed only facts in their documents and left
every decision alone — it isn't my lane, and pulling someone's security plan out
of a live earning service from the outside isn't a correction, it's a decision.

---

**2026-08-02, later — the two domains are now actually wired up.**

You gave me Cloudflare access, so I've done the things I'd previously written up as
instructions for you. Three changes, on webdesign.uk and ugg2.com only.

webdesign.uk now sends everyone to webdesign.co.uk with a temporary redirect, and
I've pointed it at a dead address on purpose. That second part matters more than it
sounds: until tonight, webdesign.uk was still serving the idea.uk shop, and because
that software doesn't check which domain it's being asked for, someone could have
placed a real order through it. The redirect alone would have hidden that, but it
would have been the only thing standing in the way. Now if the redirect were ever
switched off or mis-configured, the domain simply fails instead of quietly selling
something. I did the redirect first and the address change second, so there was no
moment where the domain was just broken.

ugg2.com previews now work, and this settles the open question from last week. I
added the wildcard record, and two made-up subdomains reached the storage bucket on
the first try, each looking for a file named after itself. That means we can give
every customer their own preview address with no setup per customer — **no server,
no certificates, nothing to renew.** A whole chunk of planned work just disappeared.

Two corrections I owe you.

First, I told you ugg2.com needed two things and had neither. It needed one — the
other half was already in place and had been all along. My evidence couldn't
actually tell the two apart, and I should have said so rather than naming both.

Second, and worse: partway through I told you idea.uk had gone down. It hadn't. It
was working perfectly the entire time. What happened is that you've moved idea.uk
onto Cloudflare since I last looked, and locked the old server so only Cloudflare
can reach it — which is exactly right, and is the safer option I'd recommended. But
my machine was still remembering the old address, so everything I sent went to a
door that's now bricked up. I checked DNS, saw the correct new answers, and took
that as confirmation the problem was at your end. It wasn't; that check was looking
somewhere else entirely. I've written the trap down properly because anyone else
touching that box this week will hit the same thing.

One genuine loose end on idea.uk, and it's the kind that doesn't announce itself.
When a site moves behind Cloudflare, the server needs a small configuration change
or it stops being able to tell visitors apart — everyone starts looking like the
same person, and the anti-abuse limit becomes one shared bucket that still *looks*
like it's working. I can't check that from outside; nobody can. It needs a look on
the machine itself. If whoever did the migration also did that step, there's nothing
to do.

Still yours to decide: the price (I've recommended £1,200), the correction fee, and
the VAT position.

---

**2026-08-03 — the shopfront exists, and it's two clicks from being live.**

I've built the webdesign.uk page: the offer, the £1,200 price you confirmed, the
money-back-until-you-accept-it guarantee, how it works, and the questions people
actually ask. It's uploaded and I've proved it renders properly by serving it
through the exact same machinery webdesign.uk will use — you can look at it right
now at **https://preview.ugg2.com/**. What you see there is byte-for-byte what
webdesign.uk will serve.

It doesn't take money and it doesn't have the chat. Both of those need the new box.
For now, someone who wants a site emails us with their domain, using a link that
fills most of the message in for them. That's weaker than a proper form, but it's
honest and it works from day one.

**What stopped me finishing.** Partway through, Cloudflare started refusing my
changes — the token is locked to particular IP addresses and this machine's address
had changed. So webdesign.uk still shows last night's redirect to webdesign.co.uk.
Either add `5.65.164.9` to the token's allowed addresses, or just make the two
changes yourself in the Cloudflare dashboard: delete the redirect rule on
webdesign.uk, and change its A record from `192.0.2.1` to `199.59.243.228`, leaving
it orange. That's it — the page is already sitting in the bucket waiting.

**On the Mythic Beasts box.** Short answer: **2 cores, 4 GB of RAM, 40–60 GB of
disk, Ubuntu 24.04, and you almost certainly don't need to pay for an IPv4
address** — because we'll reach it through a Cloudflare tunnel, so nothing needs to
connect to it from outside at all. We already run a box on exactly this pattern
(the tools-api one), so we're copying something that works rather than inventing.
The RAM is the one number I wouldn't trim; everything else has room.

The thing to be careful about on that box isn't the size, it's that it faces
strangers and spends money on every visitor. The limits on how much any one person
can use it have to go in when it's built, not afterwards — and there's a specific
trap where a rate limit behind Cloudflare silently becomes one shared bucket for
everybody while still looking like it's working. That's written down.

**Still yours:** the correction fee (nothing on the page quotes one yet), and
whether £1,200 includes VAT. The page carefully doesn't say either way, but that
needs settling before we take a real payment.

---

**2026-08-03, later — VAT sorted, and why Cloudflare keeps locking me out.**

You're not VAT registered, so £1,200 is simply £1,200. I've put that on the page in
three places: next to the price, as its own question ("Is there VAT on top?"), and
in the footer. It's worth being explicit rather than silent — a business buyer
assumes a quoted price is before VAT unless told, so saying nothing quietly reads as
"£1,440 really". It's live now; you can see it at https://preview.ugg2.com/.

I haven't put "not VAT registered" on the page. What a buyer needs to know is that
nothing gets added to the price, and that's what it says. Your registration status
is a fact about your turnover and doesn't need publishing.

**On Cloudflare blocking me.** The token is restricted to particular IP addresses,
and this machine's address had changed since the token was made — so Cloudflare
refused it. Three things worth knowing:

The obvious way to test a token says everything is fine. Cloudflare's "verify this
token" check answered "yes, active" at the exact moment real requests were being
refused, because that check is exempt from the address restriction. So it answers
"is this token still alive", not "can you use it from where you are". Anyone
debugging this will be misled by it.

Your address changes on its own. It worked last night and had rotated by this
afternoon. So adding today's address fixes it for a few days and then it breaks
again. Better to allow a range rather than one address — that survives the rotation
and still keeps out everyone who isn't on your broadband.

And the machine has two addresses, an old-style one and a new-style one, and picks
between them unpredictably. That's why it half-worked for a while. I've changed our
commands to always use the old-style one so there's only a single address to allow.

My honest recommendation: allow the range, keep the restriction. That token can
change DNS on all 36 of your domains, so the restriction is doing real work — it's
the main thing limiting the damage if the token ever leaked. Don't remove it just to
make my life easier. Longer term, once the new box exists, we run these changes from
there instead, and lock the token to that one fixed address — which is both tighter
and never in the way.

Either way you don't need to unblock me to finish the shopfront: the two changes
that put webdesign.uk live are dashboard clicks and need no token.

---

**2026-08-04, evening. The plan for putting webdesign.uk on its own machine.**

You asked for the whole thing through the framework, and the honest position is
that the framework already does three quarters of it. It builds the site, that's
running now. It can deploy to a machine instead of the bucket, and that switch is
a single database row, because relojistas already works this way and deployed
this very morning. And it has a health check for exactly this kind of site,
built, waiting to be switched on.

The plan reuses working parts rather than inventing. The box pulls its pages from
a repository every five minutes, the same way idea.uk's box does, which also means
the order of work can't go wrong: pages can pile up in the repository before the
machine even exists, and it simply catches up on its first pull. The machine needs
no public address at all, because it dials out through a tunnel, which is also
what makes the visitor identity trustworthy for rate limiting.

Two pieces are genuinely by hand, once each. Someone runs the provisioning script
on the new box, the same script family idea.uk used, kept in the repo so the
estate plan can absorb it later. And the chat service gets written once, small,
with the spending limits built in from the first line: a per-visitor cap, a hard
ceiling per day that fails to showing the contact details rather than an error,
and every conversation kept, because the conversations are the whole point of the
exercise, they tell us what people actually want.

The current holding redirect stays up until the very last step, so at no point is
anything half-built visible to the public.

What I need from you, in order: order the Mythic Beasts machine (the small spec
we discussed, and you don't need to pay for an IPv4 address), add a deploy key to
GitHub when the script pauses and asks, one look in the Cloudflare dashboard to
find how webdesign.uk is bound to the old worker so it can be unhooked at
cutover, and an Anthropic API key just for the chat, separate from the fleet's.
The site build itself needs nothing from you, it's the framework's job and it's
already in motion, just held up behind that platform bug from Sunday which
another team owns.

---

**2026-08-04, later still. One box or several?**

You asked whether one rented box could host all the dynamic sites: idea.uk,
relojistas, webdesign.uk and the ones coming later, including customer orders.

Putting many sites on one box is fine, and it's already how the machinery is
built: one folder per site, one nginx entry per site, one tunnel carrying all the
names. Density isn't the risk. The risk is mixing sites that shouldn't share a
disk.

There are really three kinds of site here. idea.uk takes real card payments, and
its keys and orders sit on its disk; that box should stay its own box, and it was
also only just properly secured, so moving it has risk and no benefit. Our own
product sites, webdesign.uk and the ones coming, share a trust class and belong
together on the one new box; each extra one costs nearly nothing to add. And the
customer sites are the surprise: they don't need a box at all. The framework
builds static sites, they go to the storage bucket like everything else, and
scale in customers is scale in files, not servers. A customer would only ever
need a machine if we deliberately sold them a backend feature, which today we
can't generate anyway, so that would be a priced decision on its own day.

So: rent the one box for our product sites, take 8 GB rather than 4 if several
chat-carrying sites are really coming, leave idea.uk and relojistas where they
are for now, and revisit folding relojistas in when the estate work makes that a
checked, mechanical move rather than a hopeful one. The version to avoid is the
one that saves a tenner a month by putting the live payment keys, an anonymous
chat that spends money, and every future site behind one kernel. One bad day
there takes all the revenue at once, and we've already had the near-miss with
the firewall-resetting script.

---

**2026-08-04, night. The order sheet for the new machine.**

Full spec, checked against what Mythic Beasts actually sell tonight rather than
memory: their virtual server line, 2 cores, 8 GB of memory (up from 4, since
you've said more sites are coming to this box), 50 GB on SSD, Ubuntu 24.04, in
one of their UK sites so customer conversations stay in the country. Their
pricing starts under a fiver a month and a year costs ten months.

One correction to what I told you before. I'd said you needn't pay for an IPv4
address. Then I checked, and GitHub and Stripe, of all things, still have no
IPv6 presence at all, and those are precisely the deploy path and the money
path. Mythic Beasts do provide a free translation service that bridges the gap,
so going without would work, but for what it costs, take the address: it keeps
the two paths that matter most off shared infrastructure. Nothing will ever
connect to it inbound either way; everything still arrives through the tunnel.

On backup: the trick is that almost nothing on this box needs backing up,
because almost everything on it is rebuilt from the repository on demand. The
pages, the configuration, the chat service itself: all versioned, all
re-creatable by running the provisioning script again. The one thing that is
genuinely unique is the conversations and orders, and those get dumped nightly,
encrypted, and pushed off the box to our storage, the same pattern our island
box has been using. Mythic's own backup add-on is worth having as a second copy
of those dumps, but it's the belt, not the braces. And before go-live we restore
one dump once, deliberately, because a backup that has never been restored is a
hope, not a backup.

---

**2026-08-05. The Gemini decision: wait.**

You've chosen to let the usage tier upgrade itself, which happens automatically
once the account's total spend crosses a hundred dollars, a few weeks away at the
current rate. Until then the whole fleet shares two hundred and fifty of the
expensive-model calls a day, refreshing around midnight. Practical consequence:
anything that fails late in the day gets re-run the next morning rather than
immediately, and a "quota" failure anywhere in the fleet between now and the
upgrade is this same known thing, not a new problem.

The landing page build re-runs on the fresh quota; everything after it is
already wired, so a successful build goes all the way to the box on its own.

---

**2026-08-05, midday. The site exists, built properly, and it's serving.**

The model swap did it. Within half an hour of switching the writer to Claude,
the page generated, passed every check, landed in the repository, and the new
machine pulled it and is serving it. I've read the served page character by
character against your rules: no em dash anywhere in the copy (one had crept
into the page title through a side door the checks don't cover; found it, fixed
it at source, and that hole is now written up for every future site), no "a
person checks it" phrasing, no speed promises. The writer did try "we usually
get back to you the same day" on one attempt; your rules blocked it, and the
rewrite came back clean. The price is on the page as £1,200, total, no VAT, and
the three-to-four-day timescale is there too.

What's left before the world sees it: your one click to authorise the tunnel,
then the cutover, which stays parked until you've had a look at the page
yourself, through the tunnel, and are happy. The chat box comes after that.

Worth saying plainly: this morning the shopfront was a hand-made stand-in and
the pipeline had never once run to the end. It's now the real thing, built by
the machinery we're selling, checked by rules you set, on a machine nothing can
reach except by permission. That's the product, working.

---

**2026-08-06, late evening. The missing-styling mystery is solved, and it's good news.**

You asked why the page arrived with no styling. We now know exactly what
happened, and it isn't a hole in the machinery. The pipeline did build
everything — the stylesheet, the logo, the pictures, the little bits of
JavaScript — and it filed them all correctly, by the rules it had at the time.
The catch: on the morning of the 4th, when those files were made, the site's
record still said "publish to the old shared shelf". We only pointed the record
at the new machine that evening, after the box was ordered. So the page, which
was re-made on the 5th, went to the new machine — and the styling, made in the
morning, is still sitting on the old shelf. Nothing re-sent it.

The fix costs nothing: the rebuild we already planned makes fresh styling and
pictures anyway, and the record now points at the right place, so everything
will land together this time. We watched the same machinery do exactly that for
idea.uk the day after the record was corrected, so this isn't hope — it's
observed.

One new thing to check before we press the button on the rebuild: the address
webdesign.uk currently forwards visitors to your webdesign.co.uk site. If the
rebuild's first step — the part that looks at what's already on the domain —
follows that forwarding, it would read the wrong site's hundred pages and get
the wrong idea, the same class of mistake that produced the one-page version.
We'll check how it behaves before submitting, and park the forwarding during
the rebuild if we have to.

---

**2026-08-08, afternoon. The rebuild is loaded and aimed; one switch needs your hand.**

We checked the thing we said we'd check before pressing the button. The crawler
that reads a domain at the start of a build does follow the forwarding, and it
came back with your webdesign.co.uk site — title, copy, links, the lot —
presented as if that were what lives at webdesign.uk. If we submitted now, the
machinery would study the wrong site's hundred pages and shape the new build
around them. Same family of mistake that produced the one-page version, so we
didn't submit.

Everything else is ready. The five-page plan is written — home, how it works,
what you get, the questions page with the price and VAT answers, and contact —
with the chat box explicitly parked for the phase after the chat service
exists. The rejected first version has been moved out of the way so the new
plan starts clean, and the two leftover to-do items pointing at it are closed.

The one thing between here and pressing go: the forwarding rules need switching
off for roughly half an hour while the build's first look at the domain
happens, then switching back on. The account this session runs under isn't
allowed to touch those rules itself, so we need you to either run the two
commands we've prepared, or allow the session to call the Cloudflare API and
we'll handle the off, the timing, and the on. Nothing visitors would notice
much either way: for that half hour, webdesign.uk would simply time out instead
of forwarding to webdesign.co.uk, and the preview link stays up throughout.

---

**2026-08-08, late evening. The rebuild is under way, and this time it read the domain right.**

We pressed go. Before submitting we found one more trap: the crawler keeps a
copy of what it last saw, so even with the forwarding switched off it would
have "seen" your other site again from memory — and the copy it held was one
our own checking had put there. We told the pipeline's first-look step to
always fetch fresh, switched the forwarding off, submitted, and watched. The
classifier's own notes this time read: no existing site, five named pages from
the roadmap, no testimonials because the service is new, no chat box in phase
one. Five pages, exactly as planned. The forwarding went back on within the
hour; visitors never saw more than a brief timeout.

One thing we had to work around, worth knowing about: the machinery that is
supposed to notice queued build work and start it isn't running on a timer —
it only moves when someone gives it a push, and the shared queue has a
fortnight of other sites' work piled ahead of ours. So we're walking this
build through step by step by hand: the domain study is done, the market
research is done, and the strategy step was running as this was written. Each
step finishes in a few minutes once started; there are a handful more before
the five pages themselves get written and appear on the preview. The queue
problem itself is written up for the platform people with everything they need.

---

**2026-08-08, midnight. The five-page site is built and serving on the preview.**

Home, how it works, what you get, the questions page, and contact — all built
by the machinery, all live on preview.webdesign.uk, with the styling, the
pictures, the logo and the menus all in place and all arriving in the right
repository this time. We read every page's visible words against your fifteen
banned-claim rules: zero hits. Your no-dashes rule: zero in anything a visitor
reads.

Worth telling you what the checks caught on the way, because it's the system
earning its keep. The writer tried "we usually reply the same day" on the
contact page — your rules blocked the save, and the rewrite came back clean.
The what-you-get page said "not a single template page with your logo dropped
on it" — which breaks your rule about not describing the product that way even
to deny it; the automatic gate missed that one (it only inspects a small slice
of page types — a real gap, now written up), our own sweep caught it, and the
page was rewritten by the machinery, not by hand. And the home page title had
picked up a long dash through the same side door as last time; fixed the same
way as last time.

Two things wait on you, same as before: the chat box phase (needs your
Anthropic key), and your review of the preview — the site's five "get started"
buttons currently have nowhere final to point, because the thing they should
point at is the chat box. That's the next build, not a defect in this one.

---

**2026-08-09, late morning. Your three complaints, answered — and one number needs your say-so.**

First, the "built by hand" worry: it wasn't. Every page and picture on the
preview arrived through the machinery, and the paper trail proves it — the
checking system even blocked some of the machinery's own drafts along the way,
which no hand-made page could experience. What made it look hand-made was your
second complaint: the home page's big picture was silently missing (its file
had been filed to the wrong shelf back on the 4th, and a missing background
picture leaves no broken-image icon, just a bare dark block). That picture is
now regenerated and showing, along with the smaller versions of the same
problem we found on the way.

The wording about changes and refunds took the most work, because the promise
wasn't one sentence — it had soaked into every layer of the machine's notes
about the site, and each rewrite faithfully brought it back. All of those
layers are now corrected: no page anywhere promises unlimited changes or an
open-ended walk-away right. What the pages don't yet do is state the new caps
in numbers, and that's deliberate: the numbers we've pencilled in — two rounds
of revisions included, fourteen days to review — are proposals waiting on you.
Say yes (or give different numbers) and the pages will state them within the
hour. The machinery is loaded to do it.

Small print: the little browser-tab icon is still missing (the tool that makes
it hit a platform fault — reported), and the email address on the contact page
is disguised by Cloudflare against spam harvesters, which looks odd to our
checking tools but works normally for a person in a browser.

---

**2026-08-09, midday. The numbers are on the pages. The site is ready for your eye again.**

You confirmed two rounds and fourteen days, and both are now stated plainly
wherever the pages talk about changes and refunds: the home page's summary row
reads "14 days refund window, 2 rounds revisions included"; How It Works
explains that acceptance or the end of the fourteen days closes the refund
right, whichever comes first; the FAQ answers it in full; What You Get states
the two rounds and points at the detail. The words are the machine's, written
under your rules, and the last "not a template" line the gate couldn't see was
caught by our own sweep and rewritten.

Finding the right lever took longer than it should have, and the reason is
written up for the platform people: the machinery keeps its facts ledger and
its writer's brief as two separate things, and a fact added to the ledger never
reaches the writer unless it's also copied into the brief. Nothing warns about
this. Once the terms went into the brief itself, every page stated them first
time.

The full check has just run against the live preview: every page serving, every
picture showing including all five heroes, zero rule violations anywhere, clean
titles. The one cosmetic gap is still the browser-tab icon, waiting on a
platform fault we've reported. The preview is at preview.webdesign.uk, ready
when you are.

---

**2026-08-09/10. A real bug in the chat box fixed and proven; a serious fault
in the building machinery found, reported, and worked around; one accidental
slip on the live contact page, caught and put right the same session; the
chat box on the page itself built but not yet switched on.**

First, a genuine problem in the chat service found before it caused trouble:
it was only ever hearing the visitor's LATEST message, with no memory of
anything said earlier in the same conversation. Fixed, tested, and proven
with a real two-message exchange through the live box — it now correctly
remembers what a visitor told it a moment before.

Then, trying to get the input box actually placed on the contact page, I hit
a real fault in the platform's own machinery for hand-driving work when its
normal queue is stalled (which it still is). Sending the platform a
perfectly correctly-worded instruction sometimes silently does nothing at
all while reporting success — and, worse, the same instruction can work once
and then fail the next time. I've written this up in detail for the platform
side to fix; it isn't specific to this site.

While tracking that fault down, two of what I believed were harmless test
instructions actually ran for real against your live contact page and
changed its picture back to a generic placeholder — the exact kind of slip
this project's own notes had already warned about. I caught it within
minutes by checking the actual page content rather than trusting a "success"
message, found the last good version in the site's own history, and put it
back byte-for-byte. I then ran the full site check again: all five pages
serving cleanly, right pictures, no rule violations, nothing else disturbed.
The contact page is back exactly as you last saw it.

The chat box itself is built and sitting ready — plain, honest copy ("Talk to
us", explains up front that it's an automated first step, points at the email
and phone below for anything else), all wired to the real chat service. It is
recorded correctly in the system's own records but I was blocked, by the
tool's own safety check, from pushing the very last step — putting the file
live — after the earlier slip. That's a sensible caution rather than a
problem to solve around, so I've left it for a fresh start or for you to say
go ahead.

Once it's live, the box sits between the picture at the top of the contact
page and the plain email/phone details underneath. The nine parked "where
should this button go" items from earlier can then all point at it.

---

**2026-08-10, later. The chat box is live. The deposit is priced and live
everywhere. All nine buttons now go somewhere real.**

The write block from earlier turned out to be about that one file in that
one moment, not a standing rule — everything went through cleanly once I
tried again. The chat box is now sitting on the contact page exactly as
planned, between your picture and the plain email and phone details. There
is one small, harmless delay: Cloudflare (the network that serves the site)
caches the box's script file for a few hours, so it may take up to a couple
more hours from writing this before it's clickable for every visitor. The
page itself is never delayed, only that one file, and it fixes itself with
no action needed.

On the deposit: we talked through the pricing and you settled on £75,
non-refundable, out of the £1,200. I checked it against what comparable
"AI builds you a website" tools charge, and against what this exact build
actually cost you in AI compute (a few dollars, measured from the real
usage logs) — both pointed well below the £80-150 you'd first suggested, so
£75 is the number I'd actually recommend rather than just the one you
picked. Every place on the site that used to say "full refund" now says it
straight: decline within 14 days and you get £1,125 back, not £1,200. I
asked the chat bot directly and it gave the right answer unprompted,
including correcting a visitor who asked about a "full refund" in those
words.

The nine buttons that had nowhere to go — the ones flagged weeks ago — now
all point somewhere real: the main "get in touch" ones lead to the contact
page (so people actually meet the chat box, not just fire off an email and
never see it), and the others go to whichever page or phone number the
button's own words already promised.

On your worry about people taking the £75 route deliberately: I'd genuinely
not build anything to stop that yet. Worst case someone pays £75 for a
proper built site and walks away, which isn't a bad outcome in its own
right, and you don't hand over the domain or hosting unless they actually
accept. This first stretch of real traffic is exactly how you'd find out if
it's a real pattern rather than a worry, at a cost you can afford to learn
from.

You just tried the chat box for real and got nothing back when you
submitted — that's not the box being broken, it's Cloudflare still serving
your browser the version of the page script from before the box existed.
It caches that file for up to 4 hours after any change, and I'd only just
made the change when we last spoke. Checked it directly: the correct
version is sitting there ready at the origin, Cloudflare just hasn't
picked it up yet. Should clear on its own within a couple of hours. If you
want it fixed right now rather than waiting, you (or whoever has
Cloudflare dashboard access) can manually purge the cache for
`webdesign.uk` — the API token I have doesn't carry that permission, so I
can't do it from here.

Follow-up on that, and it turns out there were two separate things wrong,
not one.

The cache thing was right, and I can now prove it rather than just assert
it. The web server on the box keeps its own log of every request that
reaches it, and for the whole of today there are exactly two chat
submissions in it — both mine, at 12:54 and 16:18. Yours, at around ten
past two, isn't there at all. Not as a failure, not as an error — it
simply never arrived. That's the signature of exactly what I described:
your browser had the old file, so the "send" button had nothing attached
to it and quietly did nothing. That cache has now expired on its own and
the correct file is being served, so that particular problem is gone.

But underneath it there was a second, completely unrelated problem, and
I'd have missed it if I'd taken the cache clearing as the end of the
story. I sent a real message through myself just now and got back the
"please contact us directly" message rather than a proper reply. The
reason is in the box's logs: **the Anthropic account the chat runs on has
hit a spending limit**, and the error says access comes back on the 1st of
September.

Two things worth saying about that. First, it isn't the chat box's doing —
this thing has cost well under a penny in its entire life, five messages
total, and our own safety ceiling is set at ten dollars a day, so it's
nowhere near anything I control. Whatever used up the account's allowance,
it wasn't this. Second, the fact that it answered with your phone number
and email instead of an error or a hang is the safety net working as
designed. That was the whole point of building it that way — if the AI
isn't available for any reason, the visitor still gets a real way to reach
you rather than a broken page.

**This one needs you**: it's a limit set on the Anthropic account itself,
which I can't change from here. If you raise or remove the usage limit in
the Anthropic console, the chat box will start working properly
immediately — nothing else needs redeploying or rebuilding. Until then
it's live and politely handing everyone your email and phone number.

One small thing I've deliberately not changed on my own: while the limit
is in force, that fallback message opens with "Thanks for your patience",
which hints at a short wait. If the limit really does run to September,
that's three weeks and the wording slightly oversells it. It's
customer-facing copy so I'd rather you decide, same as we did with the
deposit wording.

## 2026-08-12 — the chat bot now reads its facts from the database (the small solution, built)

You asked for the small version first: teach the existing chat program to read
each site's facts from the database instead of having them baked into its code,
so a fact you change in one place is the fact the bot states, with no rebuild.
That's built and, apart from one owner-run deploy step, proven end to end.

Why this matters concretely: two days ago the bot told visitors they'd get a
"full refund" while the database already said "£1,200 minus a £75 deposit" —
because the bot's facts were a copy, frozen at the moment its code was written,
with nothing connecting it back to the real figures. This removes the copy.

The wrinkle I had to design around: the database deliberately can't be reached
from the outside world — it only answers to a handful of services inside the
cluster, on purpose. So the chat computer doesn't talk to the database directly.
Instead there's now a tiny, read-only, password-protected service inside the
cluster whose only job is to hand out one site's facts, and the chat computer
asks *that* over a private encrypted tunnel. The database stays as locked-down
as it was; nothing new is exposed to the internet.

Setting that private tunnel up turned up a real latent fault, worth telling you
about because it had been sitting there unnoticed: the tunnel *looked* connected
(the encryption handshake succeeded) but was silently dropping every packet,
because a single system setting was switched off. It had never been caught
because nobody had ever actually pushed real traffic through it end to end — the
one time it was "tested" before, it was only checked from the wrong side. Found
it, fixed it properly, and it's now proven with real traffic: the chat computer
successfully reached a service inside the cluster.

What's left is one deploy: the small facts service ships with the next
whole-fleet release (it's harmless until then — it simply isn't there yet).
Once that's out, one config value on each side switches the bot from its old
baked-in facts to live database facts, and I've written the exact steps and an
end-to-end proof (change a fact in the database, watch the bot's answer change
with no redeploy) into the runbook. Until I flip that switch the bot runs
exactly as it did, so there's no risk in the meantime.

I put the cluster-side service through the council review as usual; that verdict
is pending. And to be careful: the new service is built so that if it ever
can't reach the database, the bot **refuses to start** rather than quietly fall
back to the old baked-in facts — because those baked-in facts are exactly the
stale copy this whole change exists to get rid of.

---

2026-08-13, evening — the bot's live facts feed broke this afternoon, and the
fix needs one command from you.

This morning's win (the chat bot reading live £149 facts from the database) got
knocked over by this afternoon's fleet release, in a way nobody had spotted:
the secret token the bot uses to authenticate was added to the cluster by hand,
but that particular store is rebuilt from a fixed list every time the fleet is
released — and the token wasn't on the list, so the release deleted it. The bot
has carried on with the facts it fetched this morning (visitors still hear £149
and the right terms), but if it restarts for any reason before the fix lands,
the chat on webdesign.uk goes down entirely. It has been in that fragile state
since about 2pm.

I've now put the token on the fixed list properly, so a release can never
delete it again — but actually writing secrets into the cluster is (rightly)
something my permissions don't allow me to do alone. You need to run one
terraform apply and one restart; the exact commands are in the runbook under
"Restoring or rotating the facts-relay token", and the recommended variant
reuses the token already on the box so nothing else needs touching. Until then
the bot works but is fragile; after it, we're back to "change a fact in the
database, the bot follows within five minutes".

---

2026-08-14, morning — halfway healed overnight; one command left, and it's
yours.

Another thread's work last night happened to run the same "apply the fixed
list" step my fix was waiting on, so the cluster side is now right — and I've
proven it from the inside: the facts service hands over the correct £149 facts
when shown the new key, and still turns away anyone without it. What's left is
the other half of the pair: the box next to the website still holds the old
key, so the bot still can't refresh. That's one command from you — it's
written out in the technical notes and in my chat summary — and it also
restarts the bot cleanly. After that the whole loop is back (change a fact in
the database, the bot follows within five minutes), and this time a fleet
release can't knock it over.

2026-08-14, a few minutes later — done. You ran the command, the bot restarted
cleanly and immediately fetched its fifteen facts from the database, and I
asked it the price to be sure: "£149 as a one-off payment… you approve the
finished site before you pay." Worth saying plainly: at no point did a visitor
see wrong information — the bot rode out the whole outage on the facts it had
already fetched. The live-facts loop is back, and it can no longer be broken
by a release, because the key it depends on is now part of what a release
installs rather than something a release forgets.

---

2026-08-15 — the claim above got its real test, and passed. Today's fresh
release ran the exact step that deleted the key on Wednesday, and this time
everything simply carried on: key present, bot refreshing, nobody had to do
anything. So "a release can't break it any more" is now a measured fact, not
a promise. Also done since you last read this: the chat box is registered as
a proper reusable tool in the framework's library, with a safety gate the
review council approved — it can only ever be offered to sites that actually
have a backend for it to talk to, so it can't end up as a dead widget on an
ordinary static site. Two decisions wait on you here (kick off the chat
experience plan; and confirm the chat box stays on the contact page — an
automated tidy-up tried to remove it and the lock stopped it), plus the
money-loop decisions on the other thread — the full list is in today's chat
summary and the new handoff.

Same evening — you decided: run the chat experience plan; find and fix
whatever made the automated tidy-up try to remove the chat box (it stays put
until that's fixed); Stripe keys a bit later; the payment confirmation will
come in through the webdesign.uk box for now; and take money first — which
sets the other thread's copy change in motion. You left the build-time
wording and the button colour open on purpose. All of it is written into the
new handoff, ready for a fresh chat.

---

2026-08-15 (evening). Your four evening decisions all got acted on today, and
three of the four work items are done.

The chat experience plan you said GO to has been written and approved. The
review council passed it with one advisory concern, and it's a real one: the
site's contact email on record is webdesign@contactforsales.com, which doesn't
match the domain. Everything the chat box falls back to — "here's how to reach
us" when someone hits the rate limit, or when the AI is down — points at those
contact details. So the plan makes checking that address its first gate, and
there's a question only you can answer: is webdesign@contactforsales.com a
deliberate choice (a real sales inbox you own), or a leftover that should be a
webdesign.uk address? There's already a review item open about it. Nothing
breaks either way; the plan just won't call those fallback journeys "proven
honest" until you've ruled.

The Stripe webhook path is built and tested end to end, without needing the
keys: a request from the internet reaches the payment service in the cluster
and comes back with its honest "no payment provider configured" answer in a
fraction of a second. That answer flips to real payment handling the moment
you add the keys — nothing else to build. One thing to know when you register
the webhook with Stripe: the main webdesign.uk address currently redirects
everything to webdesign.co.uk, and Stripe gives up on redirects — so the
webhook address to give Stripe is the preview.webdesign.uk one, unless you
carve an exception into the redirect rule in Cloudflare.

The box can now look up cluster services by name — the thing that made us pin
raw numeric addresses (which silently break if a service is ever recreated) is
fixed. Both places that had a pinned number now use proper names, and the chat
bot came back up cleanly on the named address, still serving the 15 live facts.

The improvement-loop bug you said to check and fix — the one where a sweep
tried to remove the locked chat box — is in the diagnosis loop now, framed the
way you'd want: the lock did its job, the defect is that the loop keeps
proposing page layouts that don't know locked sections exist. The diagnosis
run is minutes from its verdict as I write this. The chat box lock stays on,
as you ruled, until the fix lands.

Late addendum, same evening: the diagnosis came back and — usefully — proved
my first framing wrong. The function I named was found to be innocent: it
receives the page layout ready-made from an earlier stage and can't drop
anything. The real culprit is that earlier stage, the one that draws up the
layout list, and a second diagnosis run is already pointed at it. This is the
process working as designed: being proven wrong by the loop costs minutes;
writing the wrong cause into the record would have cost every thread that
believed it. The chat box lock stays on throughout, so nothing is at risk
while this gets pinned down.
