# Where we are — tool-suggester phantom tool links (bugs_open/029)

Plain-English running log, newest at the bottom. Append; don't rewrite.

---

## 2026-07-21 — first look, and the bug is a bit worse than it was written up as

The complaint that started this: a visitor clicked a link on leopardessconsulting.co.uk and
got a blank 404. The link pointed at a "tool" page — one of those little interactive
calculators — that had never been built.

Here's what actually happens. When the system decides a site could use a new tool, it does two
things in one go: it queues the tool to be built, and — at the same moment — it writes a set of
instructions telling the copywriter to mention that tool on a few existing pages and to link to
it. The problem is the second step guesses the tool's web address from the tool's name, using a
fixed pattern, before the tool page has actually been created. So the copywriter is handed an
address that doesn't exist yet and dutifully puts it on the page.

Two things make it worse than the original write-up said:

1. It's not that "the copywriter invented a plausible link". The *system itself* made up the
   address and even wrote an acceptance test demanding that exact (wrong) link appear. The
   copywriter was just following orders.
2. It's not only tools that never got built. I checked every such link across the fleet — 24 of
   them, across three sites — and **not one** points at a real page. Even the one tool that
   *was* successfully built has a broken link, because the real address the tool ended up at
   doesn't match the pattern the guesser used. The tools get filed under a handful of different
   address shapes and there's no way to predict which — so guessing was never going to work.

The fix is straightforward in principle: don't guess the address, and don't write the "mention
this tool" instructions until the tool page actually exists. Conveniently, the part of the
system that builds the tool page already knows the real address and already has the list of
pages to cross-link — so the cleanest place to write those instructions is right there, after
the page is made, using the real address. That removes the guessing and the timing race in one
move.

I've written the diagnosis into the bug file and the working docs, and I'm about to build the
fix. Because this changes behaviour across every site, I'll run it past the automated reviewer
panel (the "council gate") before committing, then rebuild and check it against a real tool
suggestion on a test site.

Two things this fix deliberately does NOT do, so nobody expects them:
- It doesn't retro-fix the 24 already-broken links on live pages — that's a separate cleanup
  I'll line up with the related "broken links" bug (049).
- If a tool page gets created but its build later fails, a link to it could still 404 — but
  that's a different, broader bug (049 again), not this one.

---

## 2026-07-26 — the fix is written and the emitter has been switched off in production

Picked this back up. Nobody else had touched it in the four days since (checked the commit log
and the ownership tool), and the bug had quietly got bigger: 27 broken links now instead of 24,
on four sites instead of three. Still not one of them points at a real page.

What I've built is what the plan said: the instructions to "mention this tool on that page" are
now written by the part of the system that *builds* the tool, at the moment it builds it, using
the address it has just created. Nothing guesses an address any more. Where the old code
constructed one from the tool's name, the new code simply refuses to write an instruction at all
unless it has been handed a real address.

I also added something the plan had left out, because on reflection it's the difference between
"the address is correct" and "the link works". A tool page is created before its content is
written, so for a while it exists but isn't published. If that content step never finishes — and
on the current fleet a lot of them are sitting waiting for a human to look at them — the link
would be permanently dead, which is exactly the leopardess damage arriving by a new route. So
each "mention this tool" instruction is now held back until the tool page has actually gone
live, using the system's own existing dependency mechanism. If the tool page never goes live,
the mention is never written. No link is a non-event; a dead link is the bug.

Two things are done as of today:

- **The old emitter is switched off in production.** A config change (migration 211) deletes the
  step that was writing these instructions at suggestion time, so no new broken links can be
  created from now on, regardless of when the new code ships. That took effect immediately.
- **The code is written and compiles**, and I've put it in front of the automated reviewer panel.
  It only takes effect once a new chassis image is built and rolled out, which is the next step.

Still deliberately not done, same as before: the 27 already-broken links on live pages are not
cleaned up by this — that's a separate sweep to line up with bug 049.
