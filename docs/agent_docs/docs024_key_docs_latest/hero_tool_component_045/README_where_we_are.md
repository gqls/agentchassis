# Where we are — the "Bayesian hero on every tool page" bug (045)

_Append-only, newest at the bottom. Plain prose._

---

**2026-07-21 — what this was and what I did.**

Every tool page on the platform that asked for a "tool hero" (the big banner at
the top of the page) was getting the banner for one specific product — a Bayesian
ranking calculator — no matter what the page was actually about. So an LLM cost
calculator and an ROI estimator ended up with buttons saying "Start Ranking Free"
and "Try the Bayesian Ranker". The owner spotted a couple of these on
2026-07-19.

The reason turned out to be simple and a bit surprising: the platform picks a
component for a section by looking up its "section type", and there was only ever
**one** component in the whole library that could answer the request for a "tool
hero" — the Bayesian one. So the system was doing exactly what it was told; it
just had nothing neutral to reach for. It's a missing-part problem, not a
faulty-decision problem.

The fix was to **build the missing neutral part**: a plain tool-hero banner with
no product-specific wording — the headline, sub-heading and any statistics are
written fresh for each page by the content writer, and the buttons only appear if
the system can find a real page to point them at (so no more buttons that go
nowhere). I then **moved the Bayesian one out of the shared pool** so it can't be
picked for a generic request any more — but I did **not** delete it, because it's
the only copy of that component and one live page genuinely uses it
(gamesdesign.co.uk's Bayesian ranking page). That page is untouched; if it's ever
rebuilt it'll get the neutral banner, and its actual ranking calculator is a
separate part of the page, so nothing is lost.

This was a database/configuration change, so it's **live now** — no software
release needed. I've confirmed the system now finds only the neutral banner for a
tool-hero request, and that the new banner contains none of the old Bayesian
wording.

One thing I corrected from the original bug note: it said two pages were affected
and the Bayesian component wasn't used anywhere. When I checked the live system it
was actually **three** pages, and the Bayesian one **is** live on that one
gamesdesign page. That changed how careful I had to be with the retire step
(supersede, never delete), so it was worth catching.

**What's left:** the real proof is to rebuild one of the affected pages and watch
it come out clean — a rebuild is the exact thing that used to "arm" the bug, so a
clean rebuild is the definitive test. The two affected pages are already queued
for rebuild.

---

**2026-07-26 — the proof turned up by itself, and the bug is now closed.**

Back on the 21st I decided not to force a rebuild to prove the fix. Rebuilding is
done a whole site at a time, it costs real money, and two other people were working
on the two sites in question. The bet was that the platform would get round to
those pages on its own and hand us the proof for free.

It did, though not from the page I was watching. On the 25th the platform built a
**different** tool page — the LLM cost calculator on fundamentallyai.com — and that
build went through the full path that chooses components. It picked the neutral
banner. The live page has none of the old Bayesian wording on it and the headline
reads "Compare LLM provider costs before you commit", which is the page talking
about itself rather than about somebody else's product. That is exactly what this
bug was about, so the bug is closed.

Two smaller things came out well. The new banner shows **no buttons at all** on that
page, which is the behaviour we wanted: buttons only appear when there's a real page
to send someone to, so the worst case is a missing button rather than a dead one.
And it showed **no statistics**, because we made those optional rather than let a
writer invent numbers to fill a slot.

I also checked the whole fleet for the old Bayesian wording. It survives in exactly
one place — the gamesdesign ranking page, where it is correct — and nowhere else.

**One thing I got wrong, and it nearly slipped through.** The instructions I wrote
on the 21st for checking the live page had the wrong web address in them: a slash on
the end instead of ".html". That address doesn't exist, so the site returns a small
error page. The check was "count the Bayesian words on this page, expect zero" — and
an error page has zero Bayesian words on it, so **the check passed on a page that
wasn't there.** It would have passed before the fix too. I only noticed because I
happened to ask for the status code and saw a 404 sitting next to a clean result.

The lesson is a general one and I've written it up for everyone: if you're checking
that something bad is *absent*, that check also passes when the page is missing,
when you typo the address, or when your search pattern is wrong. So always pair it
with something that must be *present* — I now also check the page contains the
banner's own tag, which correctly returns "yes" on the real page and "no" on the
broken address. The instructions are fixed, with the old version left visible and
marked as wrong rather than quietly deleted.

**What's left over.** One idea from the original bug never got built: an automatic
warning when the only component available for a generic request is a product-specific
one — the thing that would have caught this bug on day one. It was meant to be built
alongside a related bug which has since been closed, so I've handed it to the
existing "which components never get used" feature note, where it fits naturally.
Separately, two items are still sitting in the human review queue complaining about
Bayesian buttons on pages that no longer have them; that queue's clean-up is somebody
else's active job, so I've pointed it out to them rather than reaching into it.
