# SUMMARY 2026-08-24 — the route builds two-thirds of a site, and the queue tells you it built all of it

*Fourth in this lane's series, and the first written from a completed greenfield build. Predecessors:
`SUMMARY_2026-08-18_the_no_prompt_build_put_a_credit_broker_live.md`,
`SUMMARY_2026-08-18b_the_guard_holds_and_the_site_is_live.md`,
`SUMMARY_2026-08-23_the_route_dies_at_hop_two_on_a_website_it_is_not_allowed_to_read.md`
(that one's central claim was wrong within the hour — kept, and corrected in place).*

## What we are trying to do

Find out what the framework builds when it is given a domain name and nothing else. No brief, no
contact details, no seed. Not to get a nice site out of it, but to see the machine's unassisted
judgement, because whatever it does with nothing is what it will do for a customer who gives us
nothing — and every place it stumbles is a defect we can hand to whoever owns it.

## Where we have come from

The first attempt, on `loanzy.uk`, invented a credit broker: a regulated business with a lender panel
and an eligibility checker, from a loan-shaped name and nothing else. We fixed that at the root and
proved it. But loans could not tell us how much of the trouble was loans and how much was the route,
so the owner picked a deliberately dull unregulated domain — `garden-tools.uk` — and wired it up.

## What we have done

Ran it. Submitted at 17:17 with nothing attached, and watched it end to end.

**The judgement was good.** It read the name and decided on an independent gardening review hub paid
for by affiliate links, said in its own reasoning that a regulated direction did not apply, and
planned twelve pages. Asked the same question again two hours later it gave the **identical** answer,
down to a confidence score of 0.82 to two decimal places. It planned an affiliate-disclosure page and
a "how we assess" page nobody asked for. When it wanted a contact email it did not have, **it refused
to invent one** and asked a human — the exact opposite of the loanzy failure.

**Then the execution fell short in two distinct ways.**

First, it nearly died at the second step. The framework studies three example sites; our scraping
service flatly refuses to fetch one of them. One refusal throws away the whole step, including the
sites it read successfully — and that step is the only thing in the entire system that can start the
one after it. The first submission died there permanently. A second submission escaped on a
one-in-five lucky draw. That is a new bug, filed.

Second, and more consequential: **five of the twelve planned pages never built.** Not through error —
each refused honestly, saying it had no content to work with. They are all pages whose content is
*other content*: a guides index with no guides, a brand directory with no brands, a blog post with no
subject. The planner plans a site; the builder builds pages; nothing builds the material those
container pages exist to present. That was already a known bug, filed three weeks ago by another
team, and it is with them now.

**The site is live and seven pages serve.** They are substantial and read well. But the home page has
three dead links, because the pages that did build cheerfully link to the ones that did not.

## Where we are now

We have a real site at `garden-tools.uk` that is about two-thirds finished and does not know it.

That last part is the finding I would most want remembered. **The work queue says the site is done.**
Twelve rerenders report success — five of them against pages that do not exist and return a 404. And
eight items say a human needs to look at broken buttons that were quietly fixed hours ago. Both
numbers are wrong, in opposite directions, from the same table. Anyone who checks this site by
reading its status will get a confident, wrong answer whichever way they lean. Only fetching the
pages tells the truth.

I have not repaired anything. Repairing it would have ended the measurement, which was the point.

## Where we are going

Four things need fixing before the next domain goes through, and they are listed with the evidence
in the handoff. The two that matter most: the scraping failure that can kill a build outright, and
the container pages that leave a third of the plan undelivered and the front page broken.

Neither is ours to fix — both belong to lanes already working them, and both now have first-hand
greenfield evidence they did not have yesterday. That is what this lane is for.
