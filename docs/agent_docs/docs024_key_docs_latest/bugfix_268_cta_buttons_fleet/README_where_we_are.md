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
