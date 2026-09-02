# Where we are — Google Tag Manager across the domains

Plain-prose log, append-only, newest at the bottom.

---

## 2026-07-30

You asked for the Google Tag Manager tags on every page of idea.uk, and then for the
best way to track all the domains. Here is what happened.

**The tags are live on the idea.uk website.** All 21 pages carry the script high in
the head — immediately after the character-set line — and the noscript iframe as the
very first thing after `<body>`, exactly as Google specifies. I checked every page
that actually serves, not just the homepage, and confirmed it on the live site rather
than in the database.

One thing I changed about *how* it was applied, and it is worth knowing. The `<head>`
on idea.uk is not idea.uk's own — it is a shared component used by **nine** of our
sites. If I had simply pasted the container into it, eight other domains would now be
reporting into idea.uk's Google account. So instead the container ID is stored as a
per-site setting, and the shared template only emits the tag if that site has an ID.
Other sites render exactly as before. The useful side effect is that turning GTM on
for any other domain is now a one-line change rather than a rebuild.

**Then I found something more important, and it is the reason to read this.**

idea.uk is really two things behind one address. Most of it is the site we build and
publish. But the paid tool — the order form, the payment, the legal pages — is a
separate program running on the server, and nginx hands it about sixteen addresses.
Those pages never pass through our site build, so they did not get the tag.

That includes the two pages that matter most:

- the **"Request received"** page, after someone submits the £29 order, and
- the **"Payment received"** page, after they actually pay.

Without the tag on those, Google would show you visitors and **no sales at all**. The
danger is that this looks like bad news about the site rather than a missing tag — you
would be reading a chart that says "nobody buys" when the truth is "nobody told Google
when they did". It also affects the privacy, terms and refund pages, because the
versions in our site build quietly redirect to the tool's own copies.

I caught it because one URL out of twenty behaved differently, and I looked at it
instead of at the summary.

**I have written that fix but I have not deployed it.** The change is small and
tested, but it means restarting the program that takes the money, and there is
currently one live order in progress. That is your call, not mine. Rolling back is a
straight swap to the previous version, which is already the routine there.

**On the wider question — tracking all the domains.** My recommendation is one
Google Analytics property and one Tag Manager container for the whole estate, not one
per domain. The sites are built the same way, so the tags would be identical, and
maintaining fourteen copies of the same thing is how they drift apart. You still get
per-domain reporting, because every event already records which domain it came from —
it becomes a filter rather than fourteen separate reports. The one reason to split is
access: if a domain is ever handed to a client who must not see the others, that
domain wants its own container. That is a business decision, and the way I have set it
up keeps it open.

Two things I would not do. I would **not** switch on cross-domain tracking — that
feature is for one customer journey that crosses two of your sites, and these are
separate businesses; turning it on would merge unrelated visits and make the numbers
worse, not better. And I would not go to server-side tagging yet; it costs real money
to run and only idea.uk has a revenue path to justify it.

**One thing I think you should decide on separately.** These are UK sites, so
analytics cookies need consent, and there is no consent banner on any of them. Adding
GTM does not create that problem — it was already there the moment we plan to measure
anything — but it does make it concrete. Consent Mode is the standard way to handle
it and it needs a banner to feed it. Worth a conversation before we roll the tag to
the other thirteen domains.

So: idea.uk's website is done, the tool's pages are written and waiting on your word,
and the other domains are a repeat of the same recipe — three shared head templates
and six header templates cover all fourteen sites, so it is a short job once you are
happy with the shape.

---

## 2026-07-31

The chassis was rebuilt and deployed overnight. I re-checked idea.uk afterwards, because
a rebuild is exactly when something quietly reverts — it hadn't. The tags are still on
every page and the co-tenant sites are still clean.

Then I found something while preparing the rollout to the other domains, and it is worth
telling you plainly because it changes the plan.

**We already had an analytics mechanism, and I didn't know.** Four of the sites —
ai-agent-orchestration, finetuning, gaswholesalers and leopardessconsulting — share a
different page header from idea.uk's, and that one has had a Google Analytics hook built
into it since **May**. It is the same idea I built this week, just wired for Google
Analytics directly rather than Tag Manager. Nobody ever switched it on: no site has the
setting filled in, so it has sat there doing nothing for two months.

I found it by accident, a day after shipping, while checking something unrelated. That is
my mistake and I have written it up properly. I searched the documentation for prior work,
which is what the house rules ask for — but this thing was never written down anywhere. It
exists only as a row in the database. **On this platform a capability can be pure
configuration and be invisible to every document search.** I should have queried the live
components, which would have taken about a second.

**Why it matters to you rather than just to me:** if we roll Tag Manager onto those four
sites without dealing with the old hook, and someone later switches the old one on, those
sites would report every visit **twice** — once directly and once through Tag Manager.
You'd be looking at doubled traffic figures with no obvious reason. So the rollout now has
a decision in front of it: keep Tag Manager and retire the old hook. Retiring it is
completely safe right now precisely because nothing uses it — that stops being true the
moment anyone turns it on.

**So the other thirteen domains are waiting on two things from you**, and neither is
something I can decide:

First, **which container**. `GTM-PQ3WCTBD` is the one you gave me for idea.uk. Putting
thirteen other businesses into it is what I recommended — one container, one property,
with each domain identified automatically — but it is your call, and if you'd rather they
were separate I need a container ID for each.

Second, **the old hook** — my recommendation is to retire it, as above.

**And idea.uk's paid tool is still waiting.** That is the one with the "Payment received"
and "Request received" pages. Until it goes out, Google will show you visitors to idea.uk
and no sales at all. The change is built and tested; it needs a restart of the program
that takes the money, and there was an order in progress each time I looked. I have
written the exact steps down so it can be done in a couple of minutes whenever you want,
including waiting for the queue to be empty first if you'd prefer.

I've also written a full handoff document, so if we pick this up in a fresh conversation
nothing is lost.

---

## 2026-07-31 (afternoon) — it is all applied

Both things you approved are done.

**The paid tool is tagged.** I checked the order queue properly before restarting rather
than trusting the count — the one "active" order was sitting waiting for payment, not
mid-run, so a restart couldn't interrupt anything. Backed up the binary and the orders
file, deployed, restarted. The order queue came back untouched and the payment provider
is still the real Stripe one, which is the check that matters: if the restart had mangled
the settings file, the service would quietly have fallen back to a fake payment provider
and looked fine. **"Payment received" and "Request received" now carry the tag**, so
Google can finally see a sale.

**All fourteen domains are set up.** One container for the estate, as recommended, and I
retired that old Google Analytics hook I found this morning so we're not running two
systems that would double-count.

**One honest thing about the rollout.** Re-tagging every page meant re-publishing 377
pages, and each page publishes as its own commit, and each commit triggers its own deploy
job. So I queued up about 230 deploy jobs on a machine that runs two at a time. They're
all doing the same thing — each one republishes a whole site — so they're largely
redundant. Nothing is broken and nothing is lost; they just have to work through the
queue, roughly an hour and a half. **Nine of the fourteen domains were confirmed live
when I checked; the other five were still sitting behind that queue serving Monday's
files.** They'll come good on their own.

I could have avoided that by publishing the pages first and then triggering one deploy at
the end, and I've written that down so it isn't repeated. It also means other work of
yours that deploys through the same machine will be slow for the next hour or so, which I
should have thought about before firing it.

If you want to check later, the quickest test is to load any of the sites and search the
page source for "googletagmanager" — you should find it twice.

**Still open, and it is the one I'd not leave too long:** there's no cookie consent banner
on any of these sites, and now they're all running analytics. You decided to press on,
which is fair, but it's worth picking up as its own piece of work.

Everything is written up in a handoff so a fresh conversation can carry straight on.

## 2026-08-25 — picking up your "google tags" request, and two things I found on the way

You asked to be walked through the Google tag setup, pointed at Agent Chassis rather than idea.uk.
The walkthrough is in the handoff (`HANDOFF_2026-08-25_continue_here.md` here, and the apis.uk one),
and I have put it in the chat as well. Before writing a word of it I checked what is actually live.

**Nothing is being recorded yet.** The container that every site loads is still empty — version 2,
no tags, no measurement id. That matches your screenshots from last night. Until you press
Submit → Publish with a *Google Tag* pointed at the Agent Chassis measurement id, the snippet on the
pages does nothing at all. There is one command now that reads the container and says so plainly:
`scripts/check_gtm_state.sh`. Run it after you publish; it should flip to "PUBLISHED" within a minute.

**I owe you a correction from July.** On 31 July I told you "they're all running analytics" and
warned about cookies being set on fourteen sites. I had checked that the snippet was on the pages;
I had not checked the container it loads, which had no tag then and has none now. So no cookie was
ever set by it. The consent question is still real — it becomes real the moment you publish — but it
was not yet real when I raised it. I have logged that against myself.

**The second finding is the one that needs a decision from you.** When the tag was rolled out to the
newer sites on 24 August, it was written into the *stored copy* of each site's page header, not into
the site setting that the header is generated from. That works until the platform regenerates the
header — which it does whenever a site's navigation changes — and then the tag silently disappears.
It has already happened once: agritec.uk lost it six hours after it was added, when a tool page was
added to that site. Twelve more sites are in the same state (the loan calculators, noted, cookly,
garden-tools, webdesign.uk, apis.uk), and four sites built since (agritec, cv1, homegarden,
lampenkap) never had it at all. The fourteen from July are fine.

The proper fix is a small setting per site, and I have it written and tested in a dry run. **But
writing that setting makes the platform rebuild each site** — twelve sites, 241 pages, each page a
deploy job on the runner that does two at a time, so a couple of hours of queue, on top of what is
still draining from Sunday. Nothing breaks and nothing needs babysitting; it is a question of *when*
you want that queue occupied, and whether to include the four untagged sites in the same pass
(sixteen sites, 280 pages) so it is one wave rather than two. Say the word and the time.

**Still needing only you:** the consent position before publishing (the 039 reference sets it out);
a Google Cloud service account if you want Search Console automated — there is no Google credential
anywhere on our side, so I cannot start that half.

## 2026-08-26 — it stopped being theoretical overnight, so on your "carry on" I fixed it

The rebuild I asked you to pick a time for — the platform picked its own time. A nightly sweep that
was switched back on yesterday regenerated page headers across the estate overnight, and every site
where the tag lived only in the stored copy lost it: ten sites, including apis.uk, between nine last
night and six this morning. Every site with the proper setting kept it — seven out of seven, ten out
of ten the other way, as clean a confirmation as we will ever get.

So when you said "please carry on" this morning, I did the durable fix: the per-site setting is now
written on all seventeen sites that needed it, checked, and nothing else in their settings was
touched. The platform will now rebuild each of those sites in its own time — that is the tag coming
back, not more churn — and the two big loan-calculator sites never lost theirs.

One practical thing for your Google walkthrough: until apis.uk rebuilds, it serves no tag — so when
you get to the "check it works" step, load **idea.uk or vonc.com** instead and watch Realtime there.
The container itself still has no published tag; that click is still yours.

## 2026-08-26 — your lampenkap ruling is done, and your question deserves a straight answer

Your "leave lampenkap google tag" is recorded and nothing was changed — it keeps the tag and will
report like every other site.

You asked *"is that what GA4 is?"* — fair question, and the honest answer is: not quite, and the
difference matters for what happens next. The thing on every site (GTM-PQ3WCTBD) is **Tag Manager —
a socket**. GA4 is **the appliance** you plug into it, and the plug is the Measurement ID from your
Agent Chassis property. I re-checked the socket just now: it still has **nothing plugged in** — zero
tags — so no site is reporting anything to Google yet, lampenkap included. The moment you finish the
walkthrough (Google Tag with the Agent Chassis id, then Submit → Publish), every site starts
reporting at once. That one step is still the whole ballgame.

## 2026-08-26 (night) — your second-container ruling: two minutes of your clicks, then it's all mechanism

You ruled a second, cookie-light container for customer sites. Right call — it frees you to publish
GA4 for your own sites without ever touching a customer visitor. I can't create it for you: it lives
inside your Google account and there's still no machine credential on our side. It is genuinely two
minutes:

1. `tagmanager.google.com` → **Admin**
2. In the left-hand **Account** column (gqls), click **＋ Create Container**
3. Name: `Customer sites — cookie-light` · Target platform: **Web** → **Create**
4. Copy the new `GTM-XXXX` id from the top bar and paste it to me or any session — that id becomes
   the standard default wired into customer builds.

Leave the new container **empty** — empty is what makes it cookie-free. One thing to know for
later, so it doesn't surprise you: empty also means it *counts nothing*. When you want to see
whether delivered customer sites get visitors (the reason behind this ruling), the cookie-free way
is a GA4 tag in that container with consent defaults set to "denied" — it counts visits without
setting any cookie. I'll have that speced and ready for when you say go; it is not needed for the
container to exist or for anything to launch.

And to keep the two straight: **your** sites' container (`GTM-PQ3WCTBD`) is still waiting on its
own one click — the Google Tag publish from the walkthrough. That's unchanged and now fully
unblocked.

## 2026-09-02 — the customer container exists

You created it: `GTM-TH5XGNQ4`. I checked it from the outside straight away — it serves, it's
empty, it can't set a cookie. That's exactly the state it should stay in for now, and I've handed
the id to the webdesign.uk lane as the standard default for customer builds. (My earlier
instructions said to look for an Admin button that isn't on the page you were on — the three-dots
menu on the account row was the right way in, and I've fixed the written steps.)

The one Google thing left from this whole piece of work is still the same single click: publishing
the GA4 tag in your own container (the idea.uk-named one) so your sites start counting. Nothing new
has appeared in it since August — it's still recording nothing.

## 2026-09-02 (night) — it's counting

You pressed Publish and it worked: your container now carries one Google Tag reporting to
G-Y26N29T4KH, and every one of the thirty durable sites is sending page views from this evening
onwards. Load any of your sites and watch Reports → Realtime — you should see yourself.

Full honesty: my checking tool told me, for about two minutes, that your publish hadn't worked —
its parser broke on a format change at Google's end, and the tool's own output contradicted itself,
which is how I caught and fixed it. The world was fine; the instrument was wrong. Logged against
the tool, fixed, and re-proven in both directions.

Two things this changes. History starts tonight — GA4 knows nothing about anything before this
evening, so give it a few days before the numbers mean much, and always split reports by hostname.
And your sites now set cookies for real, with no consent banner anywhere — that was your call and
it stands, but it has moved from "theoretical" to "live", and it's the one compliance item I'd not
leave too long.
