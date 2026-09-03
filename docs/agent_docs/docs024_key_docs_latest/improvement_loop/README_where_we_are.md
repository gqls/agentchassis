# Where we are — the improvement loop

The owner's running log. Plain prose, append-only, newest at the bottom.

---

**2026-09-02.**

You asked me to take responsibility for the improvement loop, so the first thing I did
was find out whether it was running. That turned out to matter, because the written
record says it isn't.

**The loop is on.** There is a standing ruling of yours from 29 July that it was stopped
deliberately during a heavy development phase, and several documents still repeat it. It
was switched back on since — by a migration whose name says so — and it has been running
ever since. It fires every fifteen minutes, it picked up 32 different sites over the last
two days, and it last ran twelve minutes before I looked. Anything anyone has written
that reasons from "the sweep is off" is now wrong, and I have said so in the plan.

**The machinery is in good order.** There used to be a rule that a site got three quality
audits in its lifetime and then went quiet, which was crude — a site that has been left
alone is not the same as a site that is fine. That was replaced at the start of August by
something better: the loop takes a fingerprint of the site, and only pays for a full
audit if the site has actually changed or a fortnight has passed. I checked whether that
gate is doing real work rather than just always saying no, and it is: it ran the full
audit on about a quarter of visits and skipped the rest. That is the behaviour we want.

**Now the problem, and it is a good one to have found.**

The loop finds things. Most of what it finds is a job — something is wrong, and there is
an agent that can fix it, so the finding names that agent and gets sent along. But some
findings are deliberately not jobs. Nobody can automatically repaint a brand or repoint
somebody else's broken image, so the checker leaves the "who fixes this" box empty on
purpose and the finding is supposed to sit there for a person to read.

In August we fixed a bug where those person-shaped findings were being shoved into the
machine anyway and coming back stamped "could not be routed" — a correct observation
filed as a breakdown. The fix was to stop the machine picking them up. That was right.

What nobody did was give them anywhere to go. They sit in the state that means "just
found, waiting to be sorted", and the sorting step is precisely the one that has been
told to leave them alone. Nothing else looks at them — I checked the one job that might
have, and its own code says in as many words that it excludes them. There is a perfectly
good "waiting for a human" shelf elsewhere in the system, with 912 findings on it, and
these are not on it.

**There are 1,385 of them, across 31 sites, the oldest from 26 July.** The lane that did
the August fix counted 722 of the same kind on 19 August, so the pile has roughly doubled
in a fortnight. And every fifteen minutes the loop walks past it and reports the site
clean.

**But before recommending anything, I went and looked at what is actually in the pile,
and it is not what the number suggests.**

Two thirds of it — 867 findings — is one thing: pages are missing a "skip link", the
hidden link at the top of a page that lets someone using a keyboard or a screen reader
jump past the navigation. That is a single omission in the shared page furniture, filed
once per page across 26 sites. It is one fix, not 867.

Fifty-six findings looked much more serious: pages with no title at all. I fetched them
rather than believing them, with a deliberately invented address on the same site as a
control so I could tell a real page from a site that answers everything.

On farmerinsurance.uk, 36 of those findings are simply **wrong**. The pages have titles
and footers; I have them in front of me. The reason is worth knowing, because it affects
everything else in the pile: when a checker finds the same problem twice, the second
report is thrown away rather than replacing the first, and this particular check only
withdraws a finding when *everything* it was complaining about is fixed. The skip link is
still missing, so the finding can never be withdrawn — and it goes on repeating a
complaint about the title and the footer that stopped being true some time ago. **Any
finding in this pile is a claim of unknown age.**

The other 20 are boxingonline.com, and they are true but they are not about our pages.
Every address on that domain — including the front page — returns a 114-byte stub that
bounces the visitor to a "lander" page. The domain is parked. It is not serving our site
at all.

**So I am not going to propose building a screen to show you 1,385 findings, because a
third of what it showed you would be false or beside the point, and you would rightly
stop trusting it.** The order I intend to work in is: correct the design document that
still describes the old three-audit rule; clear out the stale and mis-framed findings;
fix the skip link at the template, which retires 867 findings in one go; and only then
tackle the real structural question, which is that a "for a human" finding currently has
no way of ever reaching a human. That last one changes something ~26 different pieces of
code write to, so it goes through the review council rather than being slipped in.

**Two things I need from you when you have a moment.** Neither blocks me.

First, do we want skip links? If the answer is that we don't care about them, the honest
fix is to retire the check, and 867 findings go with it. If we do want them, it is a
change to the page furniture on every site we run.

Second, is boxingonline.com parked on purpose — a domain we hold but do not serve — or
has it come unpointed without anyone noticing? The answer decides whether those 20
findings are damage or noise.

---

**2026-09-02, later — you asked which domains to point. The answer is two, and it is the
nameservers, not the addresses.**

I checked all 34 of our live domains rather than just the two I'd tripped over, because
the way I found the first two was luck and I didn't want to hand you a partial list.
Thirty-one are serving properly. Two are not, and both are the same story.

**boxingonline.com** and **adversecreditmortgage.co.uk** were bought and never moved off
the domain marketplace that sold them. Their nameservers still say Afternic and Dan.com —
the same company behind both, which is why they show identical parking addresses. Every
address on them returns a 114-byte stub that bounces the visitor to a "lander" page.

**The important detail: pointing the A record at our servers will not work.** The domains
aren't delegated to us, so a record set at the registrar has nowhere to take effect. What
needs changing is the nameservers, to the pair every one of our working sites uses:

```
alexis.ns.cloudflare.com
leah.ns.cloudflare.com
```

Nothing on our side is broken and nothing needs rebuilding. **Forty pages are finished
and sitting behind a delegation nobody changed** — 21 on boxingonline, 19 on
adversecreditmortgage. Once you've pointed them they should simply appear.

The full list, the evidence and the re-check script are in
`docs/agent_docs/docs024_key_docs_latest/improvement_loop/POINTING_2026-09-02_domains_to_repoint.md`.
Tell me when you've pointed them and I'll confirm at the pages rather than assume.

One thing I want to flag rather than bury: **adversecreditmortgage.co.uk filed no quality
findings at all** while serving that stub — it is one of only two sites on the estate with
a completely clean sheet. That is not a clean bill of health, it is the checker not
looking. I don't yet know why, and I've put it on my list. It is the same lesson as the
false findings on farmerinsurance, from the other direction: I can't yet take either a
finding or the absence of one at face value, which is the first thing this lane needs to
fix.

Separately, noted.co.uk returns the home page for addresses that don't exist rather than
a proper "not found". Real pages are fine. It's not a pointing problem and not mine —
I've passed it to the lane that owns that site.

**2026-09-03, evening — the "held too long" report is built, and two things I owe you a
decision on.**

Yesterday's migration made every new site start life "on hold": it doesn't grow tools until a
person says it's ready. That leaves a gap I named at the time and have now closed: nothing told
anyone when a site had been sitting on hold, forgotten. A held site doesn't error. Its checks
report clean. It just quietly stops improving — which is the same shape as the 1,385 invisible
findings this lane started on.

**The report now exists.** Once a day it lists every held site and how long it has been held,
and raises a flag for any live site held longer than seven days. It also checks that the
"born on hold" rule itself is still switched on, because if someone removed it, a clean report
would be a lie. It is built and tested against the real database, but it is *not running yet*
— it is with the review council, and I'll switch it on when they approve. **Seven days is my
guess, not your ruling** — tell me the number you want and it's a one-line change.

To make "how long" answerable I also had to add a clock: until today, a hold recorded nothing
about when it was set or by whom. Two sites are held right now (gamedesign.uk and apis.uk),
each written by a different lane in a different shape, neither with a date. New holds will
carry one automatically from the next migration; for those two, the report will say "unknown"
until their lanes add the date, and will count from the first day it saw them.

**Housekeeping I found and fixed:** yesterday's migration was applied by hand and never
written into the migrations ledger. That matters because the shared tool that applies pending
migrations would have tripped over it and stopped, taking every other lane's pending work with
it. Recorded now.

**The two decisions:**

1. **copyonline.co.uk slipped through.** It was created about two minutes *before* the
   "born on hold" rule went in, so it is the one new site the rule didn't catch. It is now
   live and open, and the portfolio lane is actively building tools on it today. I have not
   touched it — putting it on hold now would freeze their work mid-stream. Do you want it held
   like the others, or left open because that lane is in the middle of it?

2. **The seven-day threshold** — see above.

**Where the skip-link drain is:** 626 findings still open, down from 978 yesterday; 343 pages
have re-checked clean since the release. What's left underneath is 458 findings of other
kinds, and that number is creeping *up*, not down — about 10% a day. That is the next thing to
look at properly, and I have not yet.
