# Where we are — site delivery + customer editor

Owner's plain-prose log. Append-only, newest at the bottom.

---

**14 Aug 2026.** The idea: as well as the preview, publish each finished
customer site as a real hosted copy, and give the customer an editor so they
can change their own site after handover. A reference document proposed doing
this with a new central engine and Netlify; the review concluded our platform
already IS that engine, and that the automated route is Cloudflare Pages
(Netlify's ownership hand-over needs a manual dashboard step, which fails the
"completely automated" bar; and since we keep the ability to update sites,
handing the hosting away doesn't fit anyway — what the customer owns is the
ZIP, as the £149 terms already say). The customer's editor will be one small
service on our existing box; customers get in through an emailed link after
handover; their edits flow through the same machinery our own edits use, so
the existing locking system referees between their changes and ours. Six
phases, each reviewable on its own. The first (the "usually next day"
promise on webdesign.uk) went live the same evening.

**15–16 Aug 2026.** Phase 2 is built and working: finished sites can now be
published to a real hosted copy automatically, and we have watched it happen.

How it works, in plain terms. Each site has a switch that is off by default —
nothing changes for any site until someone turns it on for that site. A
timer wakes up, picks the site that has been checked least recently, and asks
one question: has anything about the built site changed since we last
published it? If nothing has, it stops and does nothing. If something has, it
copies the site to its hosted address, then — and this is the part that
matters — it fetches the page back from the public address and checks the
bytes match the original. Only if they match does it record the publish as
done. So we never mark something as published on the strength of the copy
"looking like it worked".

It did not work first time, and that is the useful part of this story. The
first live run failed on the very first file: the storage service refuses an
upload unless it is told the size in advance, and our code was streaming files
without knowing the size. Nothing was half-copied and nothing was falsely
recorded as published — the failure was clean, which is what the design was
for. We fixed it, and the fix was proven by deliberately putting the old
mistake back to confirm the test caught it. Yesterday that same site failed on
file one; today it copied all eight files and the published page is
byte-for-byte identical to the original. Ten minutes later the timer came round
again and correctly did nothing, because nothing had changed.

Two other things worth knowing. The reviewers turned the work down on the first
pass, and they were right: the way I was taking a safety backup of a shared
setting would have quietly dropped part of it, so a restore would not have
worked. Fixed and re-submitted, approved second time. And once it was running I
noticed each check was starting a whole new worker just to answer "nothing has
changed" — about a hundred and fifty a day for one site — so I slowed the timer
from every ten minutes to every hour. A finished site changes a few times a
day at most, so nobody will notice the difference, and it is one line to put
back if you want it faster.

Where that leaves us: the publishing half of the delivery promise is done. Next
is the ZIP file the customer actually owns, then the handover step, then the
emailed login, then the editor itself. Still waiting on you for two keys: the
Stripe ones for taking payment, and a Cloudflare one if we later want the
sites hosted there rather than on our own storage — the current route needs
neither, so nothing is blocked.

**17 Aug 2026.** We spent today deciding, not building — five rounds of
discussion, and the shape of the whole delivery business came out the other
side. Written down properly in the plan file dated today; here is what it
means in ordinary words.

What a customer will experience: they ask for a site and, while they wait for
it to be built, we invite them to spend two minutes creating their own free
account at Netlify (a big site-hosting company) and clicking one Connect
button. If they do — and we expect most will — their finished site is
delivered straight into an account that belongs to them from the first
moment. Their hosting, their bandwidth, their bill if they ever outgrow the
free tier; nothing for us to run. If they skip it, no harm: the build and the
sale go ahead regardless, the site is always viewable on our preview address,
and their own domain shows a friendly "choose where your site should live"
page until they pick — their free Netlify, our own hosting, or just taking
the files away. We deliberately priced our own hosting high, because we don't
want to be a hosting company; the page always shows the free option next to
it.

Money: two separate, simple charges. Keeping the domain name we chose for
them is £10 a month — we register it ourselves at trade price, since we're a
Nominet registrar, so this is good margin and completely automatic. Hosting
with us is the expensive option almost nobody should take. Everything is a
link in an email: the payment links, the file download, and a page run by
Stripe where customers manage their own subscriptions — we build no billing
screens at all.

One technical decision worth saying aloud: we will run our own nameservers —
the machines that answer "where does this domain point?". Because we register
every customer domain, that answering job is ours whichever company hosts the
site, and doing it ourselves means moving a customer between "our hosting",
"their Netlify" and "still choosing" is a one-line change on our own
machines, forever, with no outside company able to hold it hostage.

Parked for later, on purpose: giving paying hosting customers extra content
like news feeds (your idea, kept, not now); the big review of how everything
scales, including whether we need more clusters; and what to do about sites
that get seriously busy. None of these block anything.

Next: back to building. The ZIP file the customer downloads is the next piece
— it is also exactly what gets uploaded to their Netlify account, so one
piece of work serves both doors.
