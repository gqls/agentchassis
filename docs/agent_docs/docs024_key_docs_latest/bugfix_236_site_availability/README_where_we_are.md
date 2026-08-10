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
