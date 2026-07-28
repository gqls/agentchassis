# webdesign.co.uk — what the news feed taught us

**Written 2026-07-28 evening.** A new file, not an edit of
`SUMMARY_2026-07-27_corrections_and_handover.md`. Written because the owner asked
for the news problems and choices in one place, and because two days of work
turned a feature we thought was working into a fleet-wide platform fix.

---

## What we're trying to do

webdesign.co.uk is a 98-page site built from two older domains we owned. It has
two jobs that pull in different directions, and we have deliberately not averaged
them.

The **practitioner half** — 63 interactive tools and 31 guides — is the traffic
engine. It ranks, it gets shared, and it stays.

The **buying-design half** is a section for the person commissioning a £100k to
multi-million pound project. It does not need volume. Its whole argument is that
we run an AI build system across a thousand sites and are **not bidding for the
reader's work**, which makes us the only party in that market with nothing riding
on the answer about AI. We earn that by publishing where it fails.

A **news section** sits across both: a curated feed on design and web design, so
the site is a place people return to rather than a reference they consult once.

## Where we've come from

The site went live on 26 July with everything apparently working. Since then,
almost everything we checked properly turned out to be quietly broken.

Every content link on the home page was a 404 — ten of thirteen — and it had
survived inspection because the menu worked, so the site *felt* fine until you
clicked a card. The three automated link checks that should have caught it had
never run, on any site, ever.

The news feed had been "armed" for days and had ingested nothing, because creating
the sources is not what switches a feed on. Analytics had been "enabled" and was
collecting nothing. And when we did fix things, we kept finding that the fix had
not actually shipped.

The through-line, and it has now caught us four separate times: **a green status is
not evidence.** A successful deploy, a completed job, a file that exists, a
configuration value that is set — each of those was true at a moment when the thing
itself was broken.

## What we've done

**Analytics is live.** The tracking beacon is on all 99 pages, confirmed by
fetching every one of them rather than trusting the job queue. Getting there took
three attempts, because the place the previous session's notes told us to put the
token is a place the page builder never reads — so it looked correct at every step
and produced nothing.

**Every tool page now leads somewhere.** All 63 were dead ends: not one had a
single onward link, while the guides already pointed at tools. Each now carries the
guide that explains its subject and two or three related tools. No new copy was
written for it — every description is lifted word-for-word from the site's own
index pages, because writing 189 fresh descriptions is exactly where invented
claims creep in.

**Every site in the fleet has a proper 404 page.** Twenty of them. Before, a
missing page returned a raw error blob with internal storage paths in it — on a web
design site, which is a poor look for the page a lost visitor is most likely to
see.

**And the news feed turned out to be a platform bug, not a content problem.** It
had never once searched the news. Every stage of the system asked for a news
search — one part of the code forced it, the next passed it on, the search service
received it and wrote it into its own log — and then handed the actual search only
two things: the words, and how many results. There was nowhere for "news" to
travel, so it was thrown away at the final step. That is why the feed returned
encyclopaedia entries, vendor marketing and, in one case, a market forecast for
2034.

That is now fixed, fleet-wide, by another team within hours of us filing it. We ran
the confirmation on this site: articles carrying publication dates went from none
at all to every single one.

## Where we are now

The feed works. Switching on the recency window took it from **4 articles to 21**,
all published within the last month, the newest that same morning. The browser
coverage is genuinely good — Safari, Firefox and Chrome releases as they happen.
Typography gave us a major rebrand and a new type foundry.

But reading the articles rather than counting them showed the next problem, and
it is the interesting one.

**A query built from words that describe a transaction returns every industry's
transactions.** Asking for "design agency acquisition merger industry report"
brought back pharmaceutical layoffs, a supermarket merger, a law-firm private
equity piece, and an insurance broker that matched only because the company is
called Creative Planning. Meanwhile the queries that work are built from words
that exist nowhere else — CSS, browser, WCAG, typeface. **We got this wrong three
times in a row before the pattern was visible**, and each time we replaced one
generic phrasing with another and expected a different result.

We also learned that some subjects simply do not produce news. Web standards
returned nothing in a month, because specifications move yearly. The instinct was
to widen that source's window — but the platform discards anything older than
thirty days no matter what, so a wider window would have fetched a year of material
purely to throw it away.

**The owner's ruling on that is the right one and it is now the rule here: an empty
feed is acceptable if it is genuinely empty, and fewer articles on topic beat more
off it.** So the dead source has been repurposed rather than padded — pointed at
AI's influence on design, which is where the pertinent topics for modern design
now are.

The pool has been cleaned to match. Fifty-three articles collected before the fix
have been deleted, along with the off-subject ones. **Fifteen genuinely on-topic,
recently-published articles remain** — a smaller and more honest number than
yesterday's seventy-eight.

One decision closed today: **how exposed the buying-design section will be about
our own failures.** The answer is our own named failures and fixes, with the
errors listed as often as we like and the fixes claimed only once or twice. The
asymmetry is the whole design — a failure followed every time by a redemption reads
as boasting, and readers spot it within three examples.

## Where we're going

The immediate step is to read the next feed, not build on it. Two queries changed
this evening and neither has been tested; a count is not a verdict, and we have
been wrong about that three times already.

Then the news page itself. It waits for one clean tick, so the section does not
launch with a supermarket merger on it.

After that, the buying-design section can finally be written, now that the
exposure question is settled. The proof standard is the hard part: showing a
failure happened is easy and our records are unusually good, because they were
written as things went wrong rather than afterwards. Showing a fix *worked* is
harder, and it is the half that will be tested. One phrase in particular — that
our fixes are "almost comparable to a human and sometimes more" — has no
measurement behind it yet, and either needs one or needs softening.

Two things stay open and are not ours to close. Broken links still return a raw
error blob rather than the new 404 pages, which needs a small change at the edge.
And the automated link checks still have never run on any site — which is the
reason the original problem went unnoticed for so long, and the reason it would
again.
