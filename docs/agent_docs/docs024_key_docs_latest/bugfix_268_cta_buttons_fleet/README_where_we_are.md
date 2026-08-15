# Where we are — the CTA buttons fix (bug 268)

Append-only, newest at the bottom. Plain prose.

## 2026-08-13 — picked up, checked nobody else has it, confirmed the suspect code

What this is about: across 19 of our live sites, 214 call-to-action buttons
have quietly vanished. The button text is still stored, but the link it should
point at got dropped whenever a page went through a full content regeneration,
and our templates (sensibly) refuse to draw a button that goes nowhere. Nothing
errored; five of our six checks stayed green. Only webdesign.uk has been
repaired and locked so far — the other 18 sites are still damaged.

Today: a fresh thread (this one) took the handoff. First we checked nobody else
is already working it — the queue, the ownership tool, and the other live
sessions all say no, it's ours. The platform got a fresh chassis build deployed
mid-morning; we checked what's in it — one nearby change (a guard against
rewrites flattening page layout) but nothing that touches this bug.

Then we read the suspect code properly. The picture is a bit different from
what the handoff guessed: there's a rescue mechanism (built for the earlier
bug 238) that would have saved these links — but the fields in question are
routed around it by an early shortcut in the code, so the rescue never even
gets asked. That's still a theory until the diagnosis loop confirms it; firing
that run is the next step, and it's how we always test a claim like this
before building a fix on it.

The plan after that, in order: fix the leak, then repair the 214 damaged
buttons from history, then re-render, and only at the very end unlock
webdesign.uk. Repairing before fixing would just see the repairs wiped out
again — that already happened once.

## 2026-08-14 — theory traced through, a decision made, the diagnosis running

We dug through the history of the code in question. The shortcut that routes
these link fields around the rescue mechanism turns out to be older than the
rescue mechanism itself — it was never a deliberate protection, just an early
"nothing to do here" shortcut that quietly became a trap when the rescue was
added later and nobody revisited it. We also traced exactly where a rescued
link value would travel, and it comes out in both places we need it (the page
as rendered, and the stored copy the next build reads). And we found the one
thing that looked like a reason NOT to do this — an old rule that keeps these
fields out of a certain data channel — and satisfied ourselves it doesn't
apply: the rescue only ever puts back the page's own last good value; it can't
invent a link or undo someone's edit.

One decision was needed and the owner made it: the rescue applies to all
fields of this kind everywhere, on by default, rather than being switched on
field by field. A recent fix on this same pipeline set that precedent, and the
review council will be asked to judge the call explicitly.

The independent diagnosis run is now in flight — it checks our theory against
the change history in the database rather than taking our word for it. Nothing
gets committed until its verdict is in. The count of damaged buttons crept up
overnight (217 now, was 216), which is expected — the leak keeps dripping
until the fix ships.

## 2026-08-14 (later) — the checker broke, we checked by hand, and the count tells a different story

> **Correction to the entry above:** "the leak keeps dripping" over-stated it.
> See below — most of the missing buttons were never wired up in the first
> place, not deleted.

The independent diagnosis run failed on its own plumbing (its final write-up
step ran out of room and the platform correctly refused the truncated
answer), so it never ruled on our theory. Rather than fire it again, we did
what it would have done, by hand, and wrote it down: the database's own
change history shows the links present on webdesign.uk's pages right up to
the moment each rewrite ran, and gone from what the rewrite saved. Together
with the test that reproduces the loss mechanically, we're satisfied the
mechanism is proven, and the fix is committed.

But checking the history fleet-wide surprised us. Of the 217 buttons missing
across the fleet, only about ten ever HAD a link that got deleted — those we
can restore from history. The other two hundred or so never had one: the
part of the system that picks destinations for buttons never found one for
them, said so at the time in its own quiet queue, and the button was born
label-only. So this bug's fix stops real, ongoing deletions (webdesign.uk
lost seven in one afternoon, and ten more fell fleet-wide this week) — but
most of the missing buttons are a different, older problem: nobody ever
decided where they should point. That is a separate piece of work, and the
plan has been corrected so nobody expects this fix alone to bring the count
to zero.

## 2026-08-14 (afternoon) — approved, shipped, and running in production

The review council approved the fix first time, with a few advisory notes
(the useful one: check the neighbouring code paths the same way — queued for
the next session). The fix went out in this morning's platform release and
we've confirmed, at the running binaries themselves, that both servers are
running code that contains it. From this point on, a page rewrite cannot
delete a button's destination link.

What remains, in order: prove it on one real page (a controlled rewrite,
watching the links survive); restore the ten genuinely-deleted links from
history; then unlock webdesign.uk. The two-hundred-odd buttons that never had
destinations are a separate decision for you — resolve destinations for them
site by site, or accept them as label-only; the handoff file lays out the
options. A fresh session can pick all of this up from
HANDOFF_2026-08-14_canary_and_repair.md.

## 2026-08-14 (evening) — the fix is proven on a real page, and the ten links are restored

We ran the controlled rewrite on a real page: the beginners' guide on the
darts site. The writer genuinely rewrote the prose, and every button link
survived — and not by luck: the build's own record shows the new safety net
is what supplied them. Before the fix, this exact operation is what deleted
links. The page redeployed to the live site a minute later. One wrinkle
worth knowing about: the work item's STATUS said "failed", because the very
last step — reporting its own result back — hit a messaging snag. The work
itself succeeded, and we verified that on the page rather than the status,
which is exactly the habit this platform keeps teaching.

Then the repair: the ten genuinely-deleted links were restored from the
history archive (every one checked live first — all their target pages
still exist), and the seven affected pages were queued to re-render so the
restored links reach the published HTML. Re-renders are running as I write.

Two corrections to earlier notes. First, the webdesign.uk button we thought
was locked isn't — the lock went onto its neighbour, so no unlocking was
needed for the repair. Second, small print: the dispatch queue serves the
oldest work first, fleet-wide, and it is currently digesting a backlog of
several hundred items — our canary would have waited hours, so I moved our
own items to the front by adjusting their filed-time. Noted in the records
so nobody misreads those timestamps later.

Still open: the permanence proof (one more rewrite on a restored page, to
show the fix and the repair hold together), the fresh fleet count, and your
call on the two-hundred-odd never-had-a-destination buttons — for scale,
the queue that should list them only holds 71 entries across 6 sites, so
most of them have never even been queued for a decision.

## 2026-08-14 (night) — done: proven, repaired, and closed

The second rewrite test passed: we rewrote one of the freshly repaired
pages again, and the restored links came through untouched — the fix and
the repair hold together. All ten restored links are verified on the live
sites. The bug file has moved to the closed pile with the full record in
its closing section.

The two questions left are yours, and neither is urgent. First: the ~194
buttons that never had destinations — do we re-run destination-picking
site by site, accept them as label-only, or open a dedicated piece of
work? Worth knowing: the system's own to-do list for these holds only 71
entries across 6 sites, so most of them have never even been put in front
of anyone for a decision. Second: webdesign.uk's eight emergency locks
from before the fix — they can stay on as belt-and-braces, or come off now
that the fix protects those rows. Say the word either way.

## 2026-08-15 — your two decisions, actioned

You said: re-run resolution per site, and take the webdesign.uk locks off.

The locks are off — all eight, verified, with the chat-box lock (which
belongs to the other workstream) left alone. The fix that shipped this week
is what makes this safe: rewrites no longer delete button links, so the
emergency freeze has done its job and is retired.

The resolution re-run is underway. The machinery already exists: the same
re-render that repairs a mispointed button can also give a destination to a
button that never had one — it matches the button's own wording against the
site's real pages first, and falls back to the site's main tool or content
pages. It also, deliberately, keeps any link that is already valid. We're
running it site by site: the darts site is going through now as the test
case; the other twenty follow once it verifies. One honest caveat: a site
with no plausible destination pages gains nothing from this — those buttons
will stay label-only and we'll list them for you at the end rather than
invent links for them.
