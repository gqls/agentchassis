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
