# SUMMARY 2026-08-23 — the one-shot route dies at hop two, on a website it is not allowed to read

*Third in this lane's series. Predecessors: `SUMMARY_2026-08-18_the_no_prompt_build_put_a_credit_broker_live.md`
and `SUMMARY_2026-08-18b_the_guard_holds_and_the_site_is_live.md`. Each is a new file; the series is
the record.*

## What we are trying to do

Find out what the framework builds when it is given a domain name and nothing else. No brief, no
contact details, no seed — just the name. The point is not to produce a nice site. It is to see the
machine's own judgement unassisted, because whatever it does without help is what it will do for a
customer who gives us nothing, and every place it stumbles is a defect we can hand to the team that
owns it.

## Where we have come from

The first attempt at this, on `loanzy.uk` earlier in the month, built a credit broker. Given only a
loan-shaped name it invented a regulated business, complete with an eligibility checker and a lender
panel, and put it on the public internet. We fixed that at the root — the classifier is now told a
regulated business is not an available answer unless someone asks for one — and proved the fix by
running the same domain twice and changing nothing else.

That left a question the loans vertical could not answer: how much of what went wrong was about
loans, and how much was the route itself? So the owner chose a deliberately dull, unregulated test
domain — `garden-tools.uk` — wired up its DNS, and left it waiting.

## What we have done

Today we ran it. The domain went in at 17:17 with nothing attached, after checking that the ground
was clean and that the submission had actually landed rather than trusting the script, which reports
success either way.

The framework's first act was good. It read the name, decided this should be an independent
editorial hub — buying guides and honest comparisons, paid for by affiliate links — and said in as
many words that a regulated direction did not apply here. It planned twelve pages. When we later ran
the same step a second time from scratch, it returned **the identical verdict, down to a confidence
score of 0.82 to two decimal places.** That is a better result than we expected and worth saying
plainly: the machine's opinion of a domain is stable, not a fresh guess each time.

Then it died, and it died in a way nobody had seen.

The second step of the build studies three examples of the best sites in the field. It chose
Gardeners' World, The Spruce and Which?. Our scraping provider **refuses to fetch The Spruce at
all** — a flat "we do not support this site". The step has no tolerance for that: one refusal threw
away the whole stage, including the pages it had already fetched successfully.

It tried three times over an hour and a half. Each time it picked **the same three sites**, shuffled
into a different order, and died on The Spruce at whatever position it landed in. Then it stopped
for good. The step that hands the build on to the next stage is the last one in that sequence, and
it turns out to be the only thing in the entire system that can start the stage after it. So the
build has four pages of notes about what the site should be, and no site, and nothing anywhere will
ever pick it up again.

We filed that as a new bug, and while we were in there we found and fixed a related nuisance: the
record that step keeps of its own work says **"success"** for the fetch that was refused. That flag
only means the request was sent. Anyone checking whether a fix had worked would have read it and
concluded everything was fine.

Two other things came out of the day, both corrections to ourselves. A safety check we had been told
to run **after** the build — confirming we had not damaged another site's calculators — we ran
before it instead, and found the fingerprints had already changed three days earlier for unrelated
and harmless reasons. Run in the order we were given, we would have loudly reported destroying
another site's content. And a bug this lane filed four days ago turned out to name the wrong cause
entirely; another team picked it up, found the real one, and we have corrected it in the four
documents that had repeated it.

## Where we are now

The route cannot currently build a site in this vertical, and we know exactly why and exactly where.
That is a good outcome for an experiment, even though it is a bad outcome for the website.

Nothing has been papered over. We did not fix the scraping problem to make the build finish, because
a build we helped is not a measurement any more.

## Where we are going

The new bug needs an owner. The cheap half of the fix is to let the step carry on when one example
cannot be read — three examples is research, not an all-or-nothing transaction. The half that
actually pays is to remember which sites cannot be fetched and stop nominating them, because at the
moment nothing anywhere learns, and every vertical whose obvious examples are big publishers will
walk into the same wall.

Until one of those lands, a greenfield build is a coin toss decided by whether the machine happens
to admire a website our scraper is allowed to read.
