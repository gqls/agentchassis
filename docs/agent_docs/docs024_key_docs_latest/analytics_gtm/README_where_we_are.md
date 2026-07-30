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
