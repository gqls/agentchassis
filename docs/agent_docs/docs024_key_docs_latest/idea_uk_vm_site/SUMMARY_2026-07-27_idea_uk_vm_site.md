# SUMMARY — idea.uk (2026-07-27)

> ## ▶ UPDATE, 2026-07-27 11:13 UTC — the last gap closed a few hours after this was written
>
> **The owner paid the £29 and received the report. idea.uk has made its first sale.** Order
> `ord_1785090638951163875` is `delivered`; Stripe's signed webhook arrived at 11:13:13, the report
> was emailed the same second, and the queue slot released itself (`{"active":0,"max":5}`).
>
> This corrects the central claim below — *"one thing has never happened in production: a customer
> paying and receiving a report"* — which was true when written and is now retired. **Every link in
> the chain has run for real.**
>
> Corrected in place rather than replaced by a `27b`, on the same reasoning the previous session
> used for `26b`: this is the milestone the summary below *already anticipated and named as the
> next step*, not a new turn in the understanding. A summary series earns its keep by recording
> where the thinking changed direction — and here it did not; it simply arrived.
>
> One caveat the body cannot know: the delivered report was generated at 18:40 on 26 July, so it
> predates the copy fixes and the model migration and still carries two of the three writing
> defects. The **next** order is the first clean report, the first on the Claude 5 family, and the
> first that measures per-report cost. Detail in NOTES §X.23.

*Fourth in the series, after `SUMMARY_2026-07-25`, `_2026-07-26` and `_2026-07-26b`. The last one
said the site was complete and one proof was outstanding: nobody had yet received a report in the
new format. That proof happened, and then the day turned into something else — so this one exists
because "where we are now" would be unrecognisable to anyone who only read 26b. Written to be read
aloud.*

## What we're trying to do

idea.uk is where someone with an idea works it out properly: guides for every stage of an idea's
life, free tools that give an honest steer, and one paid product — the £29 Verified Idea Report —
that everything funnels towards. It is also meant to become the worked example the rest of the
portfolio copies.

## Where we've come from

Two weeks ago this was a migration project: get the site and the tool onto one address. Ten days
ago that was done. Three days ago there were nine sound pages and an empty Guides section. By
yesterday morning the whole content pipeline existed — nine guides, two free tools, and the paid
report extended so that it actually assesses the idea you send it, cites sources you can check,
and says so when an idea is too early to judge.

The last summary was written at the point where all of that was built but the extended report had
never been through a real order. That gap closed at lunchtime yesterday: a report was submitted,
researched, reviewed and read, and every promise the sales page makes was true of what came back.

## What we've done since

**We stopped treating "it works" as the same thing as "it's safe", and that turned out to matter.**
While preparing a routine deployment, we read the payment and rate-limiting code properly for the
first time and found two live faults. Neither had been exploited. Neither would ever have raised an
alarm.

- **The £29 report could be taken without paying.** A shortcut built for local testing was reachable
  on the live site: add a word to a web address and the order was marked as paid and the report
  sent. It wasn't obscure, either — a customer who started a payment and clicked "cancel" was handed
  everything they needed to do it.
- **Anyone could pretend to be a different visitor.** The free tool on the site does real AI work
  every time someone uses it, and we cap each visitor to three goes an hour to stop that running
  away with us. The cap was read from a piece of information the visitor writes themselves, so it
  could be sidestepped indefinitely.

Both are fixed, deployed, and — this is the part we insist on — *proved* fixed by carrying out the
attacks ourselves against the live site and watching them fail.

**We took a real order most of the way through the funnel.** Submitted, confirmed, researched in
nine and a half minutes, drafted, reviewed and approved, with a genuine Stripe payment link issued.
The report itself is good, and honest where honesty costs us a sale: it told the submitter that a
free government service is about to do much the same thing, that a dozen competitors already exist,
and that they should spend fifty pounds testing the idea in one town before writing any software.

**We read that report rather than just checking the job succeeded**, which found three small
writing faults no automated check would ever catch — a doubled full stop, a scoring line missing
its number, a sentence that read "using A form the receptionist fills in". All three are fixed.

**And we moved the engine onto the current generation of AI models.** That was meant to be a
one-line change and very nearly broke the product: the code chose how to talk to the models from a
list of the ones it recognised, and anything unfamiliar got the old, now-rejected treatment. Since
an unfamiliar model is by definition a newer one, the code was set up to fail on every future
upgrade. It's now inverted, so unfamiliar means modern, and the trap is gone for good.

Four separate deployments went out across the day, each one verified against the running service
rather than against our own notes.

## Where we are now

**The site is complete and the tool is sound.** Nine guides in journey order, four tools, every
button real, all authored content locked. The box is healthy, the queue is open, and the running
software carries every fix from yesterday — checked individually, because "we deployed something"
is not evidence that a particular fix is live. We learned that one the hard way: our second
security fix was found *after* the deployment that was supposed to contain it, and only re-testing
caught that it wasn't there yet.

**One thing has never happened in production: a customer paying and receiving a report.** An order
is sitting in the queue right now with a live payment link, waiting on a £29 card payment. Every
other link in the chain has now run for real at least once. This one hasn't, and until it does the
business end of the product is unproven.

> **SUPERSEDED the same day — see the UPDATE at the top.** It happened at 11:13. Left standing
> because a paragraph that was true for six hours is the most honest thing in this file.

**One open question, deliberately unanswered:** what a report now costs us to produce. We know the
top-tier model costs what the old one did and the cheaper one is currently below its old price, and
we know the new models use somewhat more tokens per job. We are not guessing the net — the next
real report measures it exactly, because the engine logs its own token usage.

## Where we're going

In order: ~~someone pays the £29 and we watch a report land in a customer's inbox for the first
time~~ — **done, 11:13 the same morning.** Then a small gap in the automatic order expiry — an order interrupted mid-report still holds its
slot forever, which is the same failure we spent yesterday fixing, reached by a different door.
Then the pipeline grows sideways: more tools where a stage earns one, and the News section that is
still empty.

The larger prize is unchanged. idea.uk is meant to be the rung the rest of the portfolio climbs to,
and it is now close enough to that to be worth copying — with the caveat that this week taught us
what "close enough" hides. Two money-losing faults sat in a product we had every reason to think
was finished, and both were found by reading code and reading output, not by any check the system
runs on itself. That is worth remembering before the next site is declared done.
