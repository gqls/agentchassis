# farmerinsurance.uk — where we are

Plain-prose log for the owner. Append only, newest at the bottom.

---

## 2026-09-04 — the site now has its own lane, and here is what it looks like today

You asked me to pick up the farmer insurance thread. Farmer did not have a thread of its own —
it was built by the loanzy lane as the second test of the greenfield route, and since then bits
of it have been handled by five other lanes (copy, news, directories, components, evidence
registers). Nobody owned the site itself. So this lane now does, the same way lendzy got its
own lane on Tuesday. I have not taken anything off anyone: the route stays with loanzy, the
copy edits stay with the copy lane, and so on. What is new is that somebody is now looking at
the SITE, regularly, and telling you what a visitor actually sees.

**The good news first.** The site is in better shape than its own to-do list suggests. All 18
pages load. Every link on every page works — all 27 of them, checked one by one this morning.
Every page has a title, a description and a proper canonical tag. The 21 tool pages you had
deleted are still gone and still return "not found". The queue attached to this site lists
about 274 outstanding items, and I have now measured that **158 of them — 58% — are either
about pages you deleted, or were true when written and have since been fixed**. The biggest
single group, 104 items, all say the same thing: "this link points at a page that has never
been built". They all mean `/claims.html`, and `/claims.html` was rebuilt and published at
2:18 this morning. Nothing goes back and retires a complaint once its cause is gone, so they
sit there looking like 104 problems.

**Two things a visitor can see, and both are on the front page.**

The first is the one I did not expect. Farmer carries a directory of **UK private health
insurers** — Bupa, AXA Health, Vitality, WPA, Saga — on its own page, linked from the homepage
under "A directory of UK health insurers". On a site about insuring farm buildings, livestock
and machinery. I found why, and it is a single line of code, not a wandering AI. The platform
has a small table that says "a site about X should carry a directory of Y". It has entries for
mortgages, savings and health insurance, and then a catch-all entry for "insurance" that hands
out the health-insurer directory, with the written reason: *"the one insurer kind built so far;
more kinds follow"*. Farmer matched the catch-all — twice over, once on its industry and once
on its own domain name containing the word "insurance". The same table, two entries further
down, refuses to do this for the word "finance", and gives the reason: *a wrong directory on a
site is worse than none.* That rule is right and it simply was not applied to the word next to
it. No other site in the estate is affected today; the mechanism would hit any pet, car,
travel or farm insurance site we build next. I have filed it and put it through the diagnosis
loop rather than just asserting it.

The second is one you already found on 31 August: **the news on the front page is American**.
Four days on it still is, and it is worse than "American" — today's headlines are corporate
takeover news. "Aon to acquire USI to establish the premier U.S. middle-market platform",
"Aon buys mid-market insurance specialist in $17B deal", and one about the Governor of Texas
directing the Texas Department of Insurance. A farmer in Devon has no use for any of it,
British or not. The search terms behind it are "insurance market, insurance regulation, claims,
premiums" — no country, and no farmer in them. The news lane measured that the platform has no
region setting at all (none of its 48 news configurations has one), so this needs building
rather than switching on. One piece of good news inside the bad: the other complaint about that
section — that its links were broken Google redirects rather than real articles — is fixed.
Every headline on the page today links to a real publisher.

**One thing is waiting on you.** Fourteen rewritten pages have been sitting in the admin queue
since Tuesday evening, ready for the batch review you asked for ("I'll review them as a batch,
present them on the admin page and let me know when to look"). Thirteen passed their automatic
checks; the fourteenth, the farm buildings page, is flagged for you because the rewrite removes
a button that pointed at one of the deleted tools — sensible, but a deletion, so it is your
call. Nobody told you they were ready. They are ready.

**What I would like a decision on: the health insurer directory.** Three options. Take it off
the site, which is quick and leaves a gap where the homepage promises a directory. Replace it
with the directory the site's own plan actually asked for — FCA-regulated agricultural
insurance brokers, NFU Mutual and the like — which is the right answer and means building a new
directory type the platform does not have yet. Or leave it, which I would not recommend on a
site whose whole claim is that it knows farm insurance.

---

## 2026-09-04, later — a correction to my own figure, and what it turned into

I told you this morning that 58% of farmer's outstanding items were dead. **That number was
wrong and the right one is 43%** — 117 of 269. I had counted a group of 42 items into the total
that were already closed off as failed, while the total I was comparing them against deliberately
left closed items out. Same day, same session, so nothing has been built on it; but it is the
kind of mistake that reads perfectly plausible either way, so it is written down where I made it.

The recount turned up something more useful than the headline. Of the 117 dead items, **63 will
clear themselves and 54 never will.** The platform has an automatic sweep that re-checks stale
items and closes the ones whose cause has gone away — but it only looks at items in two states,
and 54 of farmer's are in a third state it does not read. Forty of those 54 are complaints about
pages you deleted a week ago. Nothing will ever retire them; they are permanent furniture in this
site's list, and they are part of why the list looks alarming when the site is fine.

That is worth knowing beyond farmer, because every site's queue is read the same way.

**Correction to the outage I reported.** I told you the API credit outage ran from about 11:21
to 14:20. It ran from 11:21 to 11:56 — thirty-six minutes, not three hours. When I checked for
recovery I asked for "the last ninety minutes", got nothing but successes, and wrote down the
earliest one I could see as the moment it ended. It was only the moment I started looking. A
colleague session challenged the figure, I checked properly, and they were right: the two and a
half hours I had written off as outage contain 416 successful calls.
