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
