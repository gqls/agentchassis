# README — where we are (plain prose, append-only, newest at the bottom)

## 2026-08-22 — picking the bug up, and what turned out to already be done

Bug 342 is about a quiet way pages lose content: if the writer never supplies a piece of
text the template needs (say a headline that the component's contract marks as required),
the page doesn't fail — it just renders that spot as blank, and then the assembly step
throws the visually-empty section away. The reader never sees an error and neither do we.

The previous lane fixed the biggest part of this: every render now *notices* and *says*
when a required field came up empty, and the two riskiest editing routes (the ones that
write straight onto a live page) also file a work item about it. We checked today and all
of that is genuinely running in production — including a piece the bug file still said was
waiting to ship. The bug file was out of date on that point and we are correcting it.

What is still missing, and what this lane is doing:

1. **The editors still ship the blank.** They notice, they file the note — and then they
   write the damaged section to the live page anyway. We are adding the refusal: if the
   edit would blank out a required field, the edit is declined and the live page keeps what
   it had. This is switched off by default and turned on deliberately per agent, which is
   the house rule for giving the system new powers.
2. **A safety switch that was built but never turned on.** The site-chrome route (headers
   and footers) can file the same work item, but the switch has been off everywhere. We
   measured: turning it on today fires on nothing at all (no header or footer in use even
   has required fields), so it is free to arm now and it protects us the day someone adopts
   a chrome component that does have them.
3. **The "hard part" shrank.** The bug file worried about ~75 components with no contract
   at all. It turns out almost all of those are self-contained tools that don't need one by
   design. Only five real ones remain, each used on exactly one page. That's a small
   tidy-up job for a content lane, not a platform change, and we've written it down rather
   than done it here.

Next: put the code change through the review council, commit it for the next build, and
apply the two small config changes (one now, one held until the new code has rolled).
