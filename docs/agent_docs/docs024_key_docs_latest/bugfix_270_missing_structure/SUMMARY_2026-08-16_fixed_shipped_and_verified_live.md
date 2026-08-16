# SUMMARY — bugfix 270, 2026-08-16: fixed, shipped, verified live on a real site. CLOSED.

## What we're trying to do

Stop a maintenance check called `missing_structure` from ordering a full
rebuild of a website on every single check, forever, for no reason — it had
been doing this since April, and by the time this was found it had already
caused about 31 pointless full-site rebuilds.

## Where we've come from

The check was asking a reasonable question — "does this site's header and
footer actually exist?" — but reading the answer from three database columns
that had been empty on every page of every site for months, because the
platform had moved that information somewhere else without updating this
one check to follow. So the check could never hear "yes, it's fine" — every
single time it ran, on every site, it concluded something was broken and
ordered a rebuild.

That was found and written up two days before this fix started, by a
different team who correctly judged it wasn't theirs to fix and left it
documented for whoever had time. This work picked it up, confirmed it was
still happening (worse than when it was found), had a fix plan drawn up
independently rather than taking the obvious-looking shortcut, and wrote and
tested the actual code change: the check now asks the right table the right
question, and — because it keeps the exact same identity — the roughly
seventeen already-wrongly-filed items sitting in the queue are able to close
themselves automatically once the fix is live, using a self-correction
feature the platform already had built in for exactly this situation.

The fix went through the standard review before being deployed and was
approved without any serious concerns.

## What we've done (this session)

You deployed a fresh build of the system containing the fix. This session's
job was to confirm it actually took effect, not just assume it from the
deploy having happened.

That confirmation had two parts, and neither was a formality:

First, checking that the new code was genuinely the thing running. Rather
than trusting a version label (which turned out to be unreadable for reasons
unrelated to whether the fix was there), the check asked the running program
directly whether it still contained the exact function this fix deleted. It
didn't — on both copies of the service that handle this. That's about as
direct a proof as this kind of question gets.

Second, and more important: confirming the fix actually behaves correctly
once live, not just that the code is present. The natural way to observe
this — waiting for the system's normal daily housekeeping to re-check each
site — turned out to be unavailable, because that particular piece of
housekeeping has been switched off for an unrelated reason that other people
had already found and documented days earlier. Rather than wait indefinitely
on something that wasn't going to happen, or force the housekeeping back on
(which this codebase's own recent experience says can look like it worked
while actually doing nothing), the fix was proven directly: one specific
site with a wrongly-filed item was asked to check itself again, right now,
for real, in production. It did, and the wrong item closed itself
immediately, carrying the exact explanation the fix was written to give:
"this site's chrome is fine — the earlier flag came from a broken question."

That is the strongest kind of proof available — not a test, not a
simulation, an actual live site correcting its own record in real time.

## Where we are now

Fixed, shipped, and proven correct on a real site. The other sixteen
sites still carrying an old wrongly-filed item weren't individually
re-checked — the mechanism is proven, their underlying data was already
confirmed healthy by a direct database check, and there was no need to
spend sixteen more real production actions proving the same thing sixteen
more times. They'll clear themselves whenever their own site is next
checked by anything at all — gated by that same already-known, separately
owned scheduling issue, not by anything wrong with this fix.

Bug 270 is closed and moved to the closed-bugs record.

## Where we're going

Nothing further on this specific bug. Two loose threads it surfaced along
the way remain, each already handed off properly: a second, smaller version
of the same underlying mistake was found in a different, unrelated check
and filed as its own separate report rather than folded in here, since
nobody has hit a wrong result from it yet. And the switched-off housekeeping
task that blocked the natural verification path is somebody else's
already-documented issue to pick up, not created or owned by this work.
