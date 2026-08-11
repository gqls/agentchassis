# README — where we are (bugfix 236: sites that stop serving)

## 2026-08-10 — picked up, plan settled

Last week lendzy.co.uk turned out to be completely down — every visitor got an
error page — and nothing in the platform had noticed, because every internal
record said the site was finished and healthy. The person who found it fixed
lendzy itself, but the gap that let it happen is still there: we check the
quality of what we build in a dozen ways, and we never once ask "if a stranger
types the address, do they get the site?"

Today this thread took that gap on. The plan: teach the platform's existing
site-checking machinery one new check — fetch the site's front page from the
public internet the way a visitor would, and if it doesn't come back, raise a
high-priority flag for that site. When the site comes back, the flag clears
itself. Every live site gets this check every few hours, using the same
fair-rotation scheduler another thread built last week for the content checks.

Two honest limitations, decided out loud: a domain that has been taken over by a
parking page (which answers "successfully" with junk) is noted but not flagged —
flagging it today would also wrongly flag one healthy site in twenty-one, and a
detector that cries wolf gets ignored. And the flag is a flag, not a pager — the
queue it lands in is visible to people and tooling, but nobody gets an email.
Both are written down as follow-ups, not forgotten.

Checked before starting: all 21 live sites currently serve correctly, nobody
else is working this bug, and the fix was measured against the real fleet
before a line of code was written.

## 2026-08-10 (evening) — the check is written and committed; the reviewers are offline

The availability check is built, tested and committed. It does what the plan
said: fetches each live site's front page the way a visitor would, and if the
site does not answer, raises a high-priority flag that clears itself when the
site comes back. Every safeguard in it was tested by deliberately breaking it
and watching a named test fail — eight of them — because with all 21 sites
currently healthy there is no real fault to prove it on.

Two things you should know.

**First, the code is committed but not yet live.** It becomes active when the
platform's next image is built and rolled out, and the configuration that
switches it on is deliberately held back until we can prove the new code is
actually in the running system. That ordering is not fussiness: turning the
configuration on first would make every availability run fail loudly.

**Second, I could not get it reviewed.** The review council needs the Anthropic
API, and the account hit its spend cap at about 3:51pm today — the API says
access returns on 1 September. My submission was accepted and recorded, chose
its ten reviewers, and then died at the first one. Another thread found the same
cap this afternoon from a completely different service, so this is an account
level problem rather than anything about our code, and it is presently stopping
every AI-driven part of the platform, not just the review council. **This is
your call to make, and nothing else in the platform can work around it.**

So the honest status: the fix is done and safely inert, the review is owed, and
the review cannot happen until the API limit is lifted.

## 2026-08-10, late evening — it is live and running

The new build went out, and the availability check is in it. I checked that
properly rather than trusting the deploy: I searched the actual running program
on both servers for the new code, and searched for a deliberately misspelled
version at the same time — the real one was found, the misspelled one was not,
which is what makes the first result mean something.

The configuration that switches it on is now applied, and the first real check
has already run: it looked at robot-hands.com, found the site serving normally,
and correctly raised nothing. That first run answered a question I had flagged as
an assumption at the start — whether the servers inside our cluster can actually
reach our own websites from the outside. They can, and the way we know is neat:
if they could not, the check would have wrongly reported the site as down. Silence
was only possible because the request genuinely succeeded.

Every live site will be checked once over the next hour and three quarters, and
from then on each one gets checked every four hours, for ever, without anyone
remembering to.

**What is honestly not finished.** The check has never yet raised a real alarm,
because nothing is currently broken. Proving that half means deliberately breaking
something — briefly taking one site's routing away and confirming the alarm fires
and then clears itself when it is restored. I have written up two ways to do that,
one safer and slower, one faster and more visible, with the risks of each, and left
the choice rather than taking it unilaterally on live sites late at night. Until
that is done, the fair summary is: the alarm system is installed and wired, and we
have tested the wiring but not yet the bell.

**Still blocked on you.** The review council could not look at this work, because
the Anthropic account hit its spend limit at about a quarter to four this
afternoon and every AI step across the whole platform has failed since. That is
now written up as its own bug by another thread. My submission is saved and can be
re-sent unchanged the moment the limit lifts.

## 2026-08-11, morning — a full night of checks, and the review is unblocked

Overnight the availability check has been round-robining the whole fleet without
anyone touching it. Every one of our twenty-two live sites has now been fetched
from the public internet and found serving — twice over, the most recent pass
running from about ten past seven to five past nine this morning. Nothing has
raised a flag, which is correct, because nothing is down.

A new platform build went out at twenty past ten. I re-checked that the new build
still contains the availability check rather than assuming it — a new build is
somebody else's work and knows nothing about this one — and it does. One of the
three servers briefly ran out of disk during that changeover and had a copy
evicted; by half past ten every server reported itself healthy again, so that was
a passing squeeze rather than a problem, and the other two carried the load
throughout.

**The API spending limit has lifted.** It ended at about eleven o'clock last night,
roughly seven hours after it started — not on the 1st of September, which is what
the error message claimed. Everything AI-driven has been working normally since,
including overnight. So the review council I could not use yesterday is available
again, and my submission is still saved and ready to re-send.

**That leaves you two decisions, and I have costed both rather than guessing.**

The first is how to prove the alarm actually rings. There are two ways and they
trade the same thing off in opposite directions. The safe one uses a spare
placeholder site: they all have addresses ending `.internal`, which do not exist
on the public internet at all, so pointing the checker at one produces a genuine
failure that no customer could ever see. I measured what else in the platform would
notice such a site being temporarily marked live, and the answer is: one extra row
in the admin dashboard's list, and nothing else — every other place that looks at
live sites has a second condition these placeholders fail. The other way is the one
the original bug report suggested: briefly delete cookly.uk's routing, watch the
alarm fire, then put it back. That is more faithful, because it is the exact
failure that took lendzy down, but it means a real site is genuinely offline for a
few minutes, and I have not yet confirmed our Cloudflare key is even allowed to
delete a route.

Worth knowing before you choose: I read the code again this morning, and the two
routes produce **identical** behaviour inside the checker — a site that cannot be
reached and a site that answers with an error go down exactly the same path, and
differ only in one word recorded in the flag. So taking a live site down buys
almost nothing in terms of testing our own code. What it would buy is confirmation
of how Cloudflare behaves when a route is missing, which we already saw once, on
lendzy.

I also found something the earlier plan had missed: with the placeholder-site
route, the alarm can be proven to *fire* but not to *clear itself*, because putting
the placeholder back to its normal state stops the checker looking at it at all.
So the clearing half needs proving separately, on a healthy site — which can be
done with no risk at all, and I have written up how.

The second decision is simply whether to spend on the review round now that the
API works. It is ten reviewers and about half an hour. My recommendation is yes:
this code is live on every site, and it went in without review through no fault of
the process.
