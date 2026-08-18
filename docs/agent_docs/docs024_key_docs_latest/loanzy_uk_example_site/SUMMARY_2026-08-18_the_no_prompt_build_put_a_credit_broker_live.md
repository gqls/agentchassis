# SUMMARY 2026-08-18 — the no-prompt build invented a credit broker, and one page of it went live

Written to be read aloud. It is mostly an account of a mistake, so the mistake is not
buried: **a page presenting Loanzy as an FCA-panel credit broker was published to a live UK
domain, and it was published after I had already decided that must not happen.**

## What we are trying to do

webdesign.uk sells a one-shot website build: you give it a prompt, you pay, you get a
finished starter site. The page lead the owner approved on 18 August is *show the work,
promise nothing* — the promise being that a visitor can see real sites and the exact prompt
that produced each one. That copy is currently forbidden from mentioning examples at all,
because the four sites we could point at were built by hand-steered runs, not by the
one-shot route a customer actually gets. So the missing artefact is one honest pair: a
prompt, and the site that prompt produced. loanzy.uk — a spare domain, already wired to the
edge, with nothing on it — was chosen to carry it.

## Where we came from

Three earlier threads had touched the domain and none had built anything. One delegated it
to Cloudflare in August; one found its delegation dangling and repaired it; one asked the
owner who owned it and was told it belonged with the webdesign lane. The owner then ruled
that it would be the first example site, that it would get no positioning entry of its own,
and — the instruction that shaped today — that the framework should decide what the site is
without any prompt at all, because the pipeline already has a research stage for exactly
that.

## What we did

We dispatched the build with the domain string and nothing else: no mission, no contact
details, no seeded facts. The framework did precisely what it was asked to do, and did it
well. Within forty-five minutes it had researched the name, found two unrelated companies
trading as Loanzy elsewhere in the world, and concluded that loanzy.uk should be a UK loan
comparison and matching platform. The strategist then designed the business: lenders pay a
fee per qualified referral; an eligibility checker qualifies the visitor; a panel of
FCA-regulated lenders sits behind it; the representative APR is a featured element on every
page that mentions rates. The planner turned that into twenty pages, four of them tools —
an eligibility checker, a loan comparison tool, a repayment calculator, and a "is a loan
right for me" diagnostic — plus a lender directory and lender profile pages.

That is a credit broker. Credit broking is a regulated activity in the UK, and a site
carrying a lender panel and an eligibility checker is making a claim about itself whether or
not any business exists behind it.

I had written that exact risk down before dispatching, in this lane's notes, so that it
could not be rationalised afterwards. Having written it down, I fired anyway.

## The mistake, in full

**The first error was firing at all without a way to stop.** I identified in advance that
the only signal available was a domain name that reads as a lender, and that a lender-shaped
result would have to be retracted rather than published. I then started a build on a live,
publicly-resolving domain with no hold in place. The correct order was the reverse: arrange
the containment first — no route to the edge, or a domain that does not resolve — and only
then let the framework answer the question. Everything that followed was a consequence of
having no brake.

**The second error was believing the brake was somewhere it is not.** When the strategy came
back as a broker, I moved to stop the build before pages could deploy, and my first thought
was to cancel the rerender step, because rerendering is what publishes a site. That is wrong
here. The page builder has its own deploy step: every page ships itself as it finishes.
Holding the rerender would have held nothing at all. I caught this by reading the agent's
step list rather than assuming, which is the only reason the damage was one page instead of
twenty.

**The third error is the one that actually put a page on the internet.** At 13:57 I
cancelled the thirty-three pending items — fifteen page builds, sixteen image jobs, the
design step and the rerender. One page build had already been claimed by an agent a few
minutes earlier. Cancelling a queued item stops it being picked up; it does nothing to an
agent already running. At 14:01, four minutes after the cancellation, the About page
deployed. Its title was *"About Loanzy — A Credit Broker, Not a Lender"*, and it told
visitors that we search a panel of FCA-authorised lenders and that our eligibility checker
will not affect their credit score. Neither of those things is true of anything the owner
operates.

**A fourth, smaller error made all of this slower to see.** I had armed a watcher on the
build an hour earlier which reported nothing at all for a full hour — while the build was
producing forty-odd work items. Its query was wrapped in `2>/dev/null` and `|| true`, so a
failing query and a quiet queue looked identical. That trap is written down in our own
landmines file, with the remedy — test the watcher in the foreground before arming it —
which I had not done. The replacement watcher emits explicitly when its own query fails.

**Retracting it was not one step either.** The platform has a page retraction path, and it
refused the first attempt: it will not retract a live page, because retracting a live page
is not what archiving means. The documented order is to archive the page row by hand first,
then retract. Done in that order it worked — the file was deleted from the sites repository
with a commit naming the retraction — and the edge then has to catch up through a bucket
sync and a cache purge, which is minutes, not seconds. Until that lands the page is still
being served, so "retracted" is a claim to check at the URL, not to take from a success flag.

## What the framework got right, which matters more than it sounds

The briefing agent, two steps before any page was written, recorded what it did not know.
Its gap list includes *"FCA authorisation number — not yet known; must be obtained before
launch and added to footer"*, *"Lender panel — specific lenders not confirmed; lender
directory content cannot be populated until commercial agreements are in place"*, and
*"Legal entity name — not confirmed"*.

So the system knew. It wrote down that the business needed an authorisation it did not have,
and then built the page anyway, because a gap is a note and nothing gates on it. That is the
most useful thing this run produced: not that the model hallucinated a business, but that
the machinery noticed and had no authority to stop.

## Where we are now

Nothing of the broker site remains queued: thirty-three items cancelled, each carrying the
reason. Twenty pages sit at "planned" and were never built. The one page that reached the
public has been archived and retracted, and the edge is being verified rather than assumed.
The research, strategy and briefing specs are intact, because they are the finding.

The lane has cost one page build, one image-less design run, and about seventy minutes of
pipeline time. The example site does not exist yet.

## Where we are going

The owner has chosen to re-run with one short prompt — still a single customer input, still
no positioning entry, still the framework doing everything else — naming a business that is
not in a regulated trade. Today's run stays on the record as the evidence for why the prompt
is not optional on a domain whose name implies a regulated one.

Alongside that, the owner has asked for the framework itself to be told that a regulated
business model is not an available answer unless the brief explicitly asks for it. His
instinct that this might belong in the briefing agent is half right and worth stating
precisely: the briefing agent is where the system already *notices*, but the classifier is
where the direction is *chosen*, and every spec downstream inherits that choice. So the rule
goes to the classifier, and the briefing agent's existing gap detection is the natural
candidate for a second, harder gate — a build whose own briefing says an authorisation must
be obtained before launch should not quietly proceed to write pages that claim it.

Before either, one thing changes about how this lane works: the next build gets its
containment arranged before it is dispatched, not after it surprises us.
