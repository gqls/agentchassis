# Where we are — a site putting its own page in its own header

Plain prose, append-only, newest at the bottom. The owner maintains this too — add below,
never rewrite.

---

## 2026-08-26 (evening) — the header is the site's to decide now, and there is one thing I need you to rule on

You asked for finetuning.uk's £99 offer page to go in the header, and named four pages you were
happy to lose to make room. Contact went. **`Pricing` took the slot instead**, and a second page
had to be displaced before the page you actually asked for turned up. Nothing warned anybody;
the database said that page was in the header the whole time.

Here is why, in plain terms. The platform kept **one list of page names, shared by all 51 sites**,
ranking what it thought mattered — index, services, about, contact at the top; blog, pricing,
case studies next; everything else last. A site's own ordering number only sorted pages *within*
one of those bands. So finetuning.uk's offer page, which the site had marked as its most
important, sat behind every "pricing" and "blog" on the shelf above it, and the header only fits
eight.

The three ways to get a page promoted were: rename it to a name on the shared list, edit that
list for every site at once, or knock enough higher-ranked pages out that yours arrives last man
standing. All three are the same problem — **the site had no way to say what it wanted.**

**What I've built is that.** A site can now name its own header — an ordered list of its own
pages — and those go in, in that order, ahead of anything the shared list thinks. The shared
list doesn't go away; it becomes what it's actually good at, a sensible default for a brand-new
site that hasn't said anything yet. Nothing changes for any site until it speaks: right now
**none of the 51 has**, so the next server update changes nothing anywhere, and I can prove that
rather than assert it.

**Why an ordered list and not just "bump this page up a rank."** That was the obvious cheaper
fix and I checked it before choosing. It would have fixed finetuning.uk and done nothing at all
for gaswholesalers.com, where the page that's missing and the three pages beating it are *all in
the same band, all with the same ordering number*, and the winner is decided by nothing more
principled than the order the database handed them over. Only a real ordering fixes that.

**Something I need you to rule on.** To make this work properly, a site's declaration overrides
three separate rules the platform normally applies — including "pages of type *tool* are never
in a header". idea.uk is the case: it has marked `/report.html` as header material, given it
position 3, and gets nothing, because the platform has decided pages like that don't belong in
headers. I think a site's explicit word should beat a fleet-wide default, and I've built it that
way — but the reviewers flagged, fairly, that this is wider than "fix the ordering" and ought to
be your decision rather than mine. **If you'd rather it stayed narrow, say so and I'll pull that
part**: the fix still solves finetuning.uk and gaswholesalers either way, and idea.uk's report
page stays where it is. Two rules are *not* overridable in either version — a site can't declare
its privacy page or its 404 page into the header, and if it tries it gets told rather than
obeyed.

**Two things I got wrong, both worth recording.**

The first is almost silly and cost the most. Deciding where a site's declaration should be
stored, I checked whether the platform already had somewhere for this and concluded it didn't. It
does — I'd listed the tables and cut the listing at twenty rows, and the one I needed was the
twenty-first. Worse, that table already holds *other header settings* for three sites. So my
first version would have put "which pages are in the header" in a different place from "what the
header's button says". The reviewers caught it — they couldn't see my command, they just knew the
estate — and I moved it before it shipped. It's better off for the move: where it lives now keeps
a history, records who asked for it and why, and can be marked so an automated process won't
overwrite it.

The second is the same mistake I made this morning, in a different disguise. Test files here
carry a table saying "break this line and that test fails" — the proof each safety catch is real.
I ran mine, and two didn't hold. This morning's pair failed because the tests were tripping over
a *different* problem; this afternoon's failed because the examples I'd written couldn't produce
the failure at all — one of them listed two pages in an order that happened to already be
alphabetical, so a bug that re-sorted them changed nothing. Both fixed, and both now check their
own examples are still capable of failing.

The last thing: I re-measured what this bug actually costs before and after, and **the filed
figure was too high**. It said eight pages across five sites; the file itself warned that a second
unrelated problem was mixed in and that its author couldn't separate them. I found a way to —
not with better timestamps, which is what had failed, but by re-running the platform's own ranking
logic as a query and comparing it to what the sites are actually serving. The real figure is
**five pages across three sites**. Worth having, because otherwise we'd have "fixed" this and
watched the number not fall to zero, with no idea why.
