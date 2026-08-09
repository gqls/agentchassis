# SUMMARY — 2026-08-09 — phantom CTA cleanup: the resolver capability is built

## What we're trying to do

Fix a class of button on our sites that lies to visitors: a call-to-action whose visible
label promises one thing — "Run the Risk Checker" — while the link underneath sends you
somewhere else entirely, sometimes to something completely unrelated like a password
strength tool. The original instance was one page on one site. The question this whole
lane exists to answer is whether that was a one-off content mistake or something the
platform itself keeps doing, and if it's the platform, to fix the actual mechanism rather
than repair pages one at a time forever.

## Where we've come from

It was the platform. The button's destination is chosen by code that picks the site's
"top" pages by menu order and hands them out to buttons in position — first button gets
the first pick, second button gets the second — without ever reading what either button's
label actually says. That's fine when the labels are generic ("Get Started"), and
invisible when they're not, because a specifically-worded button is exactly the case where
a visitor notices the mismatch.

A live test on one page proved this cleanly: fixing the button by hand (asking our content
system to edit just that one link) worked mechanically, but exposed that the underlying
picker doesn't have a way to be told which target is right — so the fix removed the wrong
link but couldn't put the right one in its place. We also checked the rest of the fleet at
that point and found the identical wrong-destination pattern on other pages, on other
sites, unrelated to the one we started with — confirming it wasn't a one-off.

## What we've done

Built the missing capability: the picker now checks whether a button already has a
published label, and if it does, tries to match that label's own wording to a real page
before falling back to its old generic pick. We didn't invent a new matching system for
this — we reused the exact word-matching logic our own site-auditing tool already uses to
*detect* this problem, so the checker and the fixer now agree with each other by
construction instead of by coincidence.

While tracing through the code to build this, we also found that the platform's existing
"repair a bad link" mechanism — the thing that's supposed to run when our own audit flags
a mismatch — had a bug of its own: it would see a link that pointed at *some* real page and
leave it alone, without checking whether it was the *right* page. That's the same bug in a
different spot, so we fixed both at once.

Before asking anyone to sign off on this, we tested it against a copy of the real, live
site data — not a guess, the actual code, run for real against everything on the fleet
that has this kind of button. Out of about 1,250 buttons checked, roughly 160 would get a
different, better destination than they have today. We read through those by hand rather
than trusting the count: most are clear improvements — a button that says "open the drop
rate tuner" now genuinely opens the drop rate tuner — and we caught one real mistake before
it shipped, where a handful of common question-words ("what", "how", "why") weren't being
ignored properly and caused one button to match the wrong page purely by accident. That's
fixed now too.

## Where we are now

The code is written, tested, and was sent for independent review last night. That review
didn't actually complete — not because anything was wrong with the work, but because the
whole system's AI credits ran out partway through (a fleet-wide billing issue another team
also hit the same evening, already flagged to you). We confirmed this morning that the
credits are back and the system is working normally again, and resent the same review
request. We're waiting on its answer now — nothing else is blocking it.

Two things are deliberately left undone, on purpose, not by oversight: fixing this stops
new pages from getting the wrong link, but it doesn't reach back and correct pages that
already have one — that's a second, smaller piece of work, written down and ready for
whoever picks it up. And the content-writing side of the system still doesn't get told
which page the picker chose, so it can't yet write the button's words to match — also
written down.

## Where we're going

Read the review's answer, and act on it: if it's approved, the code needs to be built
into a new version of the platform and rolled out before it does anything (writing the
code isn't the same as it running live — nothing changes until that happens). If the
reviewers ask for changes, make them and resend. Once it's live, the natural next step is
the smaller follow-up work already named above — reaching back to fix the handful of pages
that are already wrong today, and closing the last gap so the button's words and its
destination are always decided together instead of independently.
