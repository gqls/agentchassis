# Summary — 2026-09-03 — feed markdown reaching visitors (bug 332)

Written to be read aloud. Current state only; the history is in `README_where_we_are.md` and
the evidence in `NOTES`.

---

## What we're trying to do

Stop raw formatting symbols from third-party news sources appearing as literal text on
customer sites — on every surface that shows that text, not just the one we fixed in August.

## Where we've come from

In August a lane found that news summaries were reaching pages with markdown still in them and
fixed the news page's own renderer. In closing, it noticed the same stored text was also
published as an RSS feed, with no cleaning, and filed that as bug 332 — explicitly **latent**:
only one site published RSS, that site's sources emitted no markdown, and zero visitors could
see anything wrong. It filed a re-review trigger: *watch for a second site turning RSS on.*

That trigger never fired. A different one did. On 2 September, reviewing the first paid
customer build, the owner saw broken markdown on the boxingonline news page — the surface the
bug file recorded as **fixed**. The boxingonline session added an addendum saying so and
handed it on. Nobody picked it up; the delivery lane's handoff listed it as unstaffed.

## What we've done

Re-measured everything at the served artefact, then designed the fix and put it through two
independent reviews.

The measurement changed the shape of the bug three times.

**The August fix works — it is blind, not broken.** Those pages carry zero stray `#` headings
while 1,177 stored feed rows contain them. So the cleaner runs and succeeds. Every defect
visible today is a *half* pattern — a link chopped mid-address so the closing bracket never
arrives — and the cleaner's rules require the closing bracket.

**We chop them in half ourselves.** The addendum flagged this as unverified; it is our own
search adapter, cutting each summary to 200 characters with no regard for what it cuts
through. A related, unnoticed cost: in the summaries containing links, a third of that
200-character budget is spent on the web address — an average of 69 characters no visitor ever
sees. Cleaning before cutting would hand the customer 69 more characters of real writing.

**There is a bigger surface nobody had named, and it wins.** Beside every news page we publish
a public data file that a *second* piece of code produces with no cleaning at all — currently
serving seven headings and nine broken links on a paying customer's site. Every news page
loads a script that fetches that file and replaces the cleaned page with it. So for a
JavaScript-enabled visitor, the August fix is cosmetic. Five sites are affected, not one.

The reviews were worth their cost. The sharpest catch was against our own first design: the
obvious way to clean a chopped link would also have deleted the "…" that marks the cut, turning
a visibly-broken fragment into a grammatical sentence the source never wrote — prettier, and
dishonest, on a paid page, with no test or check able to see it. Two other proposed patterns
were dropped as buying nothing or risking more than they fixed.

## Where we are now

The plan is approved and the first code is being written. Coordination messages have gone to
the four lanes whose surfaces this touches; two have already replied, and the components lane
returned a full library census that improved the design.

We know the size of the job precisely, and it is smaller than it looked: the nine already-
damaged page sections rewrite themselves every few hours from the live feed, so fixing the
producing code repairs them within about a day. No repair campaign, no work-item backlog.

One deliberate departure from what the owner picked, flagged rather than buried: he chose to
fix the truncation at source as well, and both reviews came back against touching that
particular file — it would write a permanent, un-undoable change into records that also feed
our own model training, on a file another session was editing the same afternoon. That half
has been handed to its owning lane with the measurement attached, plus an offer to make the
one uncontested part of it (the cut is done by raw byte position and can slice an accented
character in half; two records already carry the corruption).

## Where we're going

Seven narrow commits: the shared display cleaner that all three readers call, the new rules
for chopped-off markdown, the one contested test assertion changed with its argument written
out, a protective cap on an unrelated uncapped field, and the sweep script taught to look at
the data file and the feed as well as the page — because today it checks only the page, which
is why this went unseen.

Then the council gate on the code, a build, and verification at the served artefact with
controls that could come out otherwise: the five pages, both data files, and the feed — where
the signal we watch for is not "clean" but "did we accidentally empty a live feed".

Three things go out rather than in, each to a named owner: the truncation at source, the
unescaped script insertion, and ESPN's own site menu being scraped in as article text. That
last one markdown cleaning cannot fix, and it is worth the owner knowing that the news page
will still read a little oddly until it is.
