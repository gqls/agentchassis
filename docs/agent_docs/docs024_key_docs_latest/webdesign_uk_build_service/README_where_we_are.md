# Where we are — webdesign.uk, selling web design and build

Plain-prose running log, append-only, newest at the bottom.

---

## 2026-07-28 — first session, planning only

You asked for thinking and planning, so nothing has been built and nothing has
been committed to a design. What follows is what I found and what I think.

**The first thing worth saying is that you already planned this.** Last night's
buying-design plan for webdesign.co.uk has a section near the end recording your
next build — the website creation form, the tools-api, and standing up a copy of
the chassis in its own cluster with its own database. It was written down
deliberately as "recorded, not started". So this is not a new idea being invented;
it is that section being picked up.

**Most of what you want already exists, in pieces, and has taken real money.**
idea.uk is the whole shape of this product already working: a small Go program on
a VM, a page, a form, Stripe, an order, a human check, an email out. Its payment
code is behind a clean interface, it verifies the webhook signature properly, and
it has a fake payment provider for testing — that is the hard, boring, easy-to-get-
wrong half of a shop, and it is done and has survived a real sale. tools-api is
the other half: a public endpoint with rate limiting and input caps that talks to
nobody in your cluster. Between them, the shopfront is largely a copying job.

**The bit that genuinely does not exist is the one you named.** Nothing outside
the cluster can start a build. There *is* an internal admin API that can trigger a
pipeline, but it is behind an admin login and it is meant for your dashboard, not
for the public. So you are right about the gap.

**Where I want to push back, gently, is on what to do about it.** You said a
dedicated cluster so the existing one doesn't get hacked. I think there are two
different worries inside that, and only one of them needs a second cluster.

The security worry — a public website reaching into your production cluster — has
a much cheaper answer, and it is one you already chose three days ago. When the VM
estate plan asked whether the framework should push config out to the island, you
ruled that the island should pull, outbound only, so the cluster never holds a
credential to it. Turn that round and it solves this: the webdesign.uk box holds
the orders, and the cluster reaches *out* to collect the paid ones. The box never
dials in and holds no cluster credential. If someone owns that box completely,
they have the box and the orders on it, and no route to your database. That is a
single outbound HTTP call, not a second Kubernetes cluster.

But there is a second worry hiding inside yours, and I think it is the real one,
and it *is* worth money. When we build a site for a customer we scrape whatever
domain they typed, feed that content into the model, and write the results into
the same database that serves your fourteen live sites. That is untrusted content
influencing agents with write access to production. A firewall does nothing about
it. So if you want isolation — and I think eventually you should — that is the
reason to give, and it changes what needs isolating: the database and the agents,
not the front door.

The practical suggestion is to take that decision later, but not much later: after
the shop is up and money is coming in, and before the first paid build runs
against somebody's scraped website. Because of the pull design, deferring it costs
nothing — the collector points at whichever cluster exists.

**One change I'd like to argue for in the product itself.** You have the briefing
questionnaire as the optional better route. I think it has to be compulsory before
we take money and build. Not for lead capture — because it is the thing that stops
us inventing the customer's telephone number. This system has published a
fabricated contact address before, for hours, because the check that was supposed
to catch it quietly passes when a site has no email at all. And it has invented
statistics twice. A demo we build for ourselves getting something wrong is
embarrassing; a site we sell to a real business with an invented phone number or
an invented accreditation on it is their liability, published under their name,
paid for. Requiring the questionnaire turns that from a hope into a rule, and it
gives the page something true and unusual to say: we will not write a word about
your business that you haven't told us or we haven't read on your site.

The free "type your domain in" teaser still works — it just gets built only from
what we can actually read on the site that's there, and leaves the contact details
blank on purpose rather than by accident.

**On the chat box** — a genuine chat with a model is a much larger and riskier
thing than it looks, because it is unbounded input and unbounded cost from
strangers. For the job it's doing, which is collecting one domain name, a form
that asks one question at a time gives you the same conversational feel for a
fraction of the exposure. Worth a decision either way.

**One thing I did not go looking for and think you should see.** The buying-design
positioning says we run this system across about a thousand sites. The database
says thirty-two rows, of which fourteen are deployed sites and seventeen are empty
placeholders. Tracing it back, the thousand comes from the architecture threads,
where it is a *target* used to argue about how the code should scale — a perfectly
proper use — and it has drifted into outward-facing copy as a present-tense claim.
It may well be true about domains you own, which is a different sentence. It needs
pinning to its noun before it is published, and webdesign.uk makes that urgent
rather than academic: the whole pitch is that we are the ones who tell the truth
about AI web builds, and that audience checks. The honest version is at least as
good — fourteen production sites, run by the system, on the record including where
it went wrong.

**What I'd do next, if you agree with the shape:** put the shop up on a VM with
Stripe in test mode and nothing behind it, and see whether anyone types a domain
in and goes through with it. idea.uk did exactly that with a fake door before it
built the engine. Then the free teaser, then real money with me or another session
doing the build by hand — which is the "next day or so" on the page, honestly. All
three of those are a working business and none of them touch the platform code, so
none of them need the council, a new cluster, or a decision you'd be stuck with.

The full plan, with the phases and the seven open questions, is in
`PLAN_2026-07-28_webdesign_uk_build_service.md`.

---

## 2026-07-28 — your three rulings, and what they change

You ruled on all three the same evening. Recording what each one moves, because
two of them went against what I'd recommended and that has consequences worth
seeing rather than discovering later.

**The trust boundary: yes, ship it now, decide isolation at P3.** That's now a
constraint rather than a preference, and the thing to protect is the direction of
travel — the box never dials into the cluster and never holds a credential to it.
Any later shortcut that reverses that gives back the entire security argument for
free, so it's worth being unfashionably strict about. The one thing I've written
down so it doesn't get skipped: the isolation decision has a specific trigger —
the first paid build that scrapes a domain we don't own — and whoever runs that
build owes the decision. It would be very easy, with money coming in by then, for
that moment to pass unnoticed.

**The questionnaire stays optional.** Fine, and I've kept my argument in the plan
rather than deleting it, because the risk it describes doesn't disappear with the
gate — it just has to live somewhere else. The good news is that "somewhere else"
is stronger than the gate was. A gate is a rule someone can be talked out of; what
I've put in its place is structural. We simply don't generate a contact block at
all unless you gave us the details — not a placeholder that might slip through,
absent. The bug that let a fabricated address reach production was a *check*
failing open, and a field that is never written can't fail open. Alongside that,
the platform already has the right mechanism: seed the site's evidence base before
the first page is written, built from the two things we genuinely have — what the
customer told the chat, and what we read on their existing site, each attributed.

There is one condition attached, and I'd rather flag it now than have it arrive as
a surprise. The optional questionnaire is safe *while a human looks at every site
before it goes out*. P5 is the phase that automates release. When release stops
passing a human, this decision needs re-opening — either the gate comes back, or
those two controls have to be shown to be doing the work by themselves. I've
written that into the phase itself so the step that removes the safety net is the
step that notices.

**A real chat, then.** You were right that it buys something my version didn't,
and I'd missed it: a real conversation can *do* the briefing. It can ask for the
phone number and the services naturally, while someone is engaged, which is a far
better collection mechanism than an optional form afterwards. So your two rulings
work together better than either does alone — the chat is much less likely to
leave us with an empty brief than a stepped form plus an optional questionnaire
would have been.

The cost of it is that it moves work earlier. A fake door with a form costs
nothing to run and could go up bare. A fake door with a real chat spends money on
every visitor, including hostile ones, so the limits — per-visitor rate, a cap on
how long one conversation can run, and above all a per-day ceiling on total spend
— have to ship *with* it rather than after. The first phase is therefore a bigger
piece of work than it looked this afternoon, and it's the one not to rush.

The one part I'd think about hardest is not a setting. What someone types into
that chat ends up in the brief, and the brief is read by the agents that build the
site. Someone typing "ignore your instructions and…" is writing into a document
our own system will later read. The answer is that the transcript enters the build
as quoted customer statements in a named field, never as loose prose pasted into a
prompt — which is the same discipline as the evidence base, arrived at from a
different direction.

**Still needing you:** the price (unblocked as soon as I measure what a build
actually costs), whether the preview sites sit on `*.preview.webdesign.uk` or a
subdomain of another of your domains, and the thousand-sites figure.
