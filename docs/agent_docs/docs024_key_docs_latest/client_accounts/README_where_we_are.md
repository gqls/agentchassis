# Client accounts — where we are

Plain-prose log for the owner. Append only, newest at the bottom.

---

## 2026-09-04 — opened the lane, and the first thing I found is that the link does not exist

You asked me to find everything we have planned and discussed about setting up client
accounts, and to let this lane own it. Here is what is actually there.

**First, a fork in the road, because "client accounts" means two different things this week.**
One is an account *with us* — somewhere a customer logs in and sees their site, their domain,
their hosting and their bill. The other is the account you were talking about yesterday: the
one at Netlify, that a customer has to create for themselves before they can host the site we
gave them, and which took you forty minutes. Both are written up. They are the same subject
from opposite ends, but they are different pieces of work, so I have put the question of which
one you want at the top of the plan rather than guessing.

**Second, the thing I did not expect.** Back in August you ruled that a customer's identity
should live on the chain we already have — a customer owns a network, a network owns sites. The
columns for that were added and they are there. But nothing has ever used them. There is exactly
one network in the whole system, called "Default Network", and all forty-two of our sites that
have a network at all point at it. There is one real customer row, "Boxing Online", created a
week ago with your email address on it — and it is not connected to boxingonline.com by
anything at all. Nothing in the code ever creates a network for a customer; there is a single
function that makes the one default one and that is it.

So the shape you ruled is real as a filing cabinet and empty as a record. Whatever we decide a
client account *does*, this is the first thing that has to be built, and it is small.

**Third, where a customer's details actually live today: four places, none of them joined.**
The address they paid with sits in the build queue. The payment sits in the billing tables. The
address the site publishes sits on the site row — and since the fix in August that is
deliberately a *different* address from the paying one. And who we actually delivered to is
being added by another session right now, as its own table, because the obvious column was
populated and wrong. None of these is the customer record you ruled canonical.

**Fourth, what already exists and works.** Every customer-facing link we send is a token of
ours — one hashed row per link, with an expiry, and a list of permitted kinds that is closed on
purpose so nobody can quietly invent a new one. Two kinds exist today: download your zip, and
confirm you have moved. A third is already reserved and named for a customer login, and it has
not been built. That is genuinely the whole of a "log in" mechanism in prospect: no passwords,
no signup, a link in an email.

**Fifth, the constraint that shapes everything.** There is no public way into our cluster at all
— no front door, only the private tunnel. That is why the design puts any customer login on the
box outside, calling in over the tunnel, rather than on a login page of our own. It is the same
reason the Stripe webhook still has nowhere to arrive.

**What I have not done:** anything. This is a survey. Two other sessions are working in this
area right now — one on the delivery pipeline generally, one on the follow-up emails — and the
customer editor is formally theirs, so before building I need to know which of the two readings
you meant, and whether you want the paid-hosting option from yesterday's list to be part of it.

---

## 2026-09-04, later — you pointed me at the delivery thread's pre-plan, and I am leading it from here

Two things I had missed, and both matter.

**The delivery thread had already written a starting plan for this thread**, the same day. It is
good, I have adopted its shape, and I have argued with it in four places where I had evidence it did
not. **And you had already ruled on the identity question the day before** — four identities:
who ordered, who runs the site, what the site shows the public, and who the site is about, with more
than one contact allowed for each. I had been about to ask you a question you have already answered,
which is worse than not asking, so it is written down where the next person will see it.

**What I am adding to their plan, in plain terms.**

*It is not just tidying old records.* Their first step is to give every site an owner. That is right,
but the part that lasts is the other half: the code that creates a site **cannot record an owner at
all** — it attaches every new site to the one shared "default" group, with no way to say otherwise.
So tidying the existing sixty would start going stale the next morning. The code has to learn it
first, then the tidying sticks.

*The link we already email people is tied to one site, not to a person.* Their cleverest idea is that
the first "account" needs no login at all — just a page at a secret link, like the download and
confirm links we already send. I agree, and it is much cheaper than a login. But the machinery
behind those links is built to point at a single site, so a page showing *someone's sites* needs a
small change first. That is the same question as "is an account a person or a site?", which is on
your list below.

*A login is dearer than it looks.* We have no public front door into our own system at all — nothing
of ours is reachable from the internet except through a private tunnel. So a customer login is not a
page we add to something we already run; it is a whole new public thing, with password resets and
support emails attached. That is the strongest argument for their no-login first step.

*On hosting past the thirty days, which you asked me to cost rather than build:* the surprising half
is that **keeping a site up costs nothing to build, because nothing currently takes one down.** The
six-week cut-off is written in the database and nothing acts on it. All of the work — and all of the
risk — is in the other direction: stopping service for someone who has not paid, without
accidentally unpublishing a business over a billing hiccup. I will cost it that way round.

**One more thing worth knowing.** The finetuning service thread has written down that it needs a
customer account system and that none exists. If we build this one for websites only, they will
build a second. They should be told before that happens.

**Four questions are yours, and they are in the plan.** The one that decides the most is the first:
do customers log in at all, or is a secret link to a read-only page enough?

---

## 2026-09-04, later still — your two answers, and the separate-cluster idea

**You settled the two questions that were holding things up.** Customers get a page at a secret link
now, with a real login as a stated destination later — which is different from "no login ever", and
it changes what I build: the page has to be built so that what the link opens and what a password
would later open are *the same thing*, with the password just being a second key to the same door.
And an account is one per **person**, not one per site — so someone with three sites is one customer,
which is what most of the rest of the design hangs off.

**On a separate cluster for customers.** It is a real idea with real history here, and the useful
thing I can tell you is that it means three different things and they have three different answers.

*Serving their sites:* a separate cluster would change nothing, because their sites do not run on our
cluster at all — they are files in Backblaze served through Cloudflare. The genuine shared-fate risk
at that layer is that every customer sits in **one Cloudflare account**, so one bad page could get
the whole account flagged and take everybody down. That risk is already written up and costed,
with a staged answer that kicks in somewhere around five hundred to a thousand domains.

*Running their builds:* this is the one that was actually proposed before, and it is further along
than anyone would guess. There is a service running in our cluster **right now**, and has been for
187 days, whose only job is to accept work sent from another cluster. It is idle. The mechanism for
"run this on a different cluster" is built and deployed; what is missing is a second cluster and a
reason. The honest caveat is that our message bus currently has **no access control at all** —
everything connects anonymously with full rights — so a customer-facing satellite would need that
fixed first, and that is a bigger job than the satellite.

*Keeping their data separate:* here it works against us right now. Today a customer's details are
already scattered across four different places that nothing joins up — and the fourth of those was
created this week precisely because something trusted the third. Splitting the database before we
have joined it up would add a fifth.

**So my recommendation is that it does not change the plan — it makes the first step more valuable.**
You cannot split customers onto their own cluster until you can answer "which things are this
customer's?", and that is exactly what we cannot answer today. That first step is the same work
whichever way the isolation question goes.

I have added it as a fourth line in the hosting costing you asked for, so it gets priced properly
rather than argued about. And I have sent it back to the scale review, which is where you parked the
cluster question yourself until after the first working site — I did not want to quietly un-park it
on your behalf.

---

## 2026-09-04, evening — the payments thread and this one have agreed where the line is, and I have written the rule

The session doing Stripe work asked to own the money — vouchers, orders, the webhook, checkout, any
page where a customer pays — and to leave identity to us. That is the right line and I took it.

Before agreeing I checked one thing they had asserted, and it turned out to be wrong in a way that
mattered to both of us. **They believed a payment is what creates a customer record. It isn't.** The
one real payment we have ever taken went through perfectly on 27 August, and created nothing: for a
one-off card payment Stripe doesn't bother making a customer, so there was no customer to record.
They then found the sharper version of it — **the customer record already existed eleven minutes
before the payment cleared.** It was made when the order was taken, not when the money arrived.

**I had made the same mistake in our own plan an hour earlier**, so I have corrected it where it sits
and logged it. It is worth knowing why two of us got it wrong independently: we both read the same
line in the register, which describes how the mechanism is *designed* to work. Nothing in that line
can tell you whether it has ever actually happened. Both of us heard "this is built" as "this is
working". One query separated them.

They then asked for the one thing that was blocking them: a settled rule for **"which customer does
this order belong to?"** — and they named the trap themselves. There are two ways to write that
question and only one is safe. Asking *"who is the customer for this site?"* sounds natural and is
the trap: **today it would answer "Default Client" for every site we have, including the one somebody
paid for**, and it would look completely correct doing it.

So the rule is keyed on the person, never on the site, and I have written it up properly. The short
version: **a customer is somebody who exists; the roles you settled last week — who ordered it, who
runs it, what it shows the public, who it is about — are jobs that person holds on a particular
site.** An order belongs to whoever ordered it. A site gets its owner from the order, never the other
way round.

Four things make it safe and they are all small: match email addresses in a plain way and no cleverer
than that (because wrongly merging two customers cannot be undone, while wrongly splitting them can);
add a database rule preventing two customer records sharing an address, which is missing today; make
the create-if-missing step atomic so two orders arriving together cannot make two records; and **if
there is no email at all, stop and ask a human — never quietly file it under the default customer**.
That last one is the same mistake that put your own email address on thirteen public pages in August,
wearing different clothes.

The useful part is that this makes the two threads' worry go away entirely. They were asking how our
two record-creating paths could be kept in step. The answer is that there is only one: theirs. Ours
reads the decision rather than making its own.
