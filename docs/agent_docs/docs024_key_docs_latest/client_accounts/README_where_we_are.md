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
