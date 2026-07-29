# SUMMARY — 2026-07-29: the news feature is finished, and Buying design has its first two pages

*(Series: follows `SUMMARY_2026-07-28_what_the_news_feed_taught_us.md`. Written to be read aloud.)*

## What we're trying to do

Make webdesign.co.uk two things at once without letting them blur: a
practitioner site whose hundred-odd tool and guide pages bring the traffic, and
a commissioning resource — "Buying design" — for the people who pay for large
web projects. The commissioning side sells nothing and ranks nobody. Its
credibility rests on us publishing what our own AI build system gets wrong,
which the owner ruled we may do freely, provided we claim our fixes sparingly
and only with evidence.

## Where we've come from

Yesterday the news feed worked but had no page: twenty-five on-topic items in
the database, one query still on probation, and a news page that had sat marked
"planned" for two days. The buying-design section existed only as a plan,
unblocked the day before by the owner's exposure ruling. And the handoff
promised that once the news page deployed, its menu entry would reappear on its
own.

## What we've done

The news page is live. It was never going to build itself — nothing on the
platform watches for planned pages, so we filed the build request by hand and
the machinery did the rest in three minutes. Reading the result found two
things every status had called fine: a headline describing a page that doesn't
exist (release notes for our own tools, on an industry news page) and a
call-to-action with no buttons, because the writer supplied button text but no
destinations and the template silently drops a button without an address. Both
fixed. The probationary news query came good on its third wording: five or six
of nine stories genuinely about AI in web design, and the platform's own triage
flagged the duds.

The menu promise turned out to be false, and the reason was a platform bug:
one word missing from the list that decides which pages count as a section's
front page. Our news page sits at the standard address every future news page
will get, and that exact address is the one the missing word made invisible.
We wrote the one-word fix, tested it, put it through the review council — it
approved first round — and rolled it to the fleet within the hour. Exactly one
page fleet-wide was wrongly hidden, and it was ours.

Buying design is live at /buying-design/: a front door that names five of our
own failures from the closed bug record and claims exactly one fix, with its
evidence; and an accessibility page giving the buyer the law (checked against
the government's own text that morning), the standard to write into a contract,
and three failures anyone can see using tools the site already had. Both pages
hand-written — the rules for this section don't allow a language model near it.

## Where we are now

Everything is deployed and verified on the wire: the news page serves real
stories, the two buying-design pages return 200 with every internal link
resolving, and all three are in the site search. The platform fix is live on
v1.0.1198. The site menu across all pages is mid-rebuild — the page-by-page
refresh moves at the platform's own pace and takes a few hours — after which
News and Buying design appear in every header.

## Where we're going

Next for Buying design: the remaining pillars, starting with why large projects
fail and what the buyer should own at handover, and then the first buyer tool —
the side-by-side site benchmark is the anchor, but it needs real build work.
For news: nothing; it runs itself now, and the one recorded flaw (five outlets
covering one story count as five items) is a platform matter, filed not fixed.
The practitioner copy rewrite remains the standing engine work, and ordering by
popularity waits until that rewrite lands, per the owner's sequencing.
