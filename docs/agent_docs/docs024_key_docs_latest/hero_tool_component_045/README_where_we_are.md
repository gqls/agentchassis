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
