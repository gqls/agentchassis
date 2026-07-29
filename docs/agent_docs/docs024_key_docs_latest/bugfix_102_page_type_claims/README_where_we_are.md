# Where we are — the claims checker and teaching pages (bug 102)

Plain prose, append-only, newest at the bottom.

---

## 2026-07-28, evening — what this bug was, and what it turned out to be

We have a checker that reads every page before it goes live and asks: is there a
number here that we are stating as a fact about the business, which nothing in our
evidence register supports? It exists because we once published invented figures.
It works by looking at the words around a number — "clients", "records", "users",
"uptime" — and deciding from that whether the number is a business claim.

That question is the right one on a sales page. It is the wrong one on a page that
is *teaching* something. When a games-design article explains drop rates and says
"in a game with 10,000 active players farming that item, roughly 180 of them will
hit that wall", the checker sees a big number surrounded by business-sounding words
and flags it. The sentence is a worked example. There is nothing to verify.

The thing that separates the two is not in the words at all — it is that we already
record what kind of page each one is. Guide, blog post, landing page, tool. The
checker simply never looked.

**The bug had been filed as a nuisance blocking one site. It is bigger than that,
and I measured it before writing a line of code.** Nine of our sites have the
checker armed. Between them they carry 124 of these findings right now. Sixty-one
of those sixty-one are on teaching pages, and I read every one: all sixty-one are
false alarms. Meanwhile the sixty-three on ordinary business pages include the real
ones the checker exists to catch.

And it is not only noise. A finding of this kind is filed at a severity that
**stops the page being rebuilt**. So today, four blog posts on gamesdesign.co.uk
cannot be regenerated, because a probability example inside them looks like a lie
about the company.

## What I changed

The checker now knows what kind of page it is reading. On teaching pages — guides,
blog posts, index listings, interactive tools — it stops guessing at numbers in the
prose.

**But it does not stop checking those pages.** That distinction is the whole of the
care in this fix, and the bug report as filed would have got it slightly wrong. We
keep a per-site list of *specific* claims a human has audited out as false — "70+
agents across eight departments" and the like. Those are still matched on every
page, of every kind, because the very first time this checker ran for real, the
thing it caught was one of those banned claims sitting in a guide. Figures printed
in stat cards are still checked everywhere too, for the same reason: a stat card is
a claim by construction, wherever it sits.

So the change is narrow: only the *guessing* part stops, and only where guessing has
been measured to be wrong every single time.

I then re-ran the identical fleet-wide scan with the fixed code: 124 findings became
63. The 61 that vanished are exactly the 61 false alarms, and nothing new appeared.
The tests check it in both directions — including that the same sentences on a
business page still get flagged, because a checker that has been switched off and a
checker that has been fixed look identical if you only count what stopped.

## Two things worth knowing

**One.** This unblocks something the owner asked for. webdesign.co.uk — our largest
site — could not have the checker armed at all, because doing so would have raised
fifteen complaints about perfectly good teaching copy. All fifteen were on guide
pages. That obstacle is gone.

**Two, and this is the honest bit.** Another session was rewriting the same four
files at the same minute, for a related bug. Their commit landed four minutes before
mine and took my half-finished edits with it — which is normal here and nothing was
lost — but it took the *callers* of my new code without the code itself, so for
about five minutes the shared codebase did not compile at all. Anyone building an
image in that window would have failed. I found it by building the committed code in
a clean directory rather than by reading the diff, and committed the missing half
immediately. Worth saying out loud because it is a cost of many sessions sharing one
working tree, and it is invisible unless someone checks.

## What is left

Three things, each recorded in the bug file rather than quietly done:

- Report pages still throw fourteen false alarms, but for a completely different
  reason: product model numbers ("Schunk EGP 40-N-S-B") read as figures. I did not
  fold that into this fix, because it would have been a coincidence rather than a
  mechanism.
- The narrower idea in the original report — teaching a checker to recognise phrases
  like "let's say" and "for example" — is still worth doing, and is the only thing
  that helps a worked example sitting on an ordinary page. It is a separate change.
- The code is committed but does not take effect until the platform's next image
  roll. There is a one-line check in the runbook to confirm it, and I have written
  down the trap: check for a string the change *introduced*, with a control, rather
  than one that was always there.

## 2026-07-28, late — live, and closed

You deployed a fresh chassis (v1.0.1196) and I checked it against the running pods rather than
against the tag: the new code is in both. The checker now knows what kind of page it is reading,
on the live system.

One honest caveat, written into the case file as well. Between my first commit and the build I
narrowed the fix — the council review had come back approved but with three comments, and two of
them were right. I had included two page types on reasoning rather than evidence, and one of
them turned out to cover gamesdesign's *about* and *contact* pages, which are filed in the
database as "section-index" despite being ordinary marketing pages. So I took both out. What I
cannot prove by inspecting the binary is *which of my two versions* is in it — the difference is
two words that appear in several other files anyway. The build happened fourteen minutes after
the narrowing commit and builds always take the committed code, so it is almost certainly the
narrowed one; I have written that down as an inference rather than a fact, because it is one.

The ticket is closed and moved to the closed-bugs folder. Three follow-ons are recorded there
rather than quietly dropped — including one the reviewers were right to press me on: nothing yet
*watches* for a page being filed under the wrong type in future. Today I checked it by hand and
it was fine except for the two pages above. That is a check, not a control, and the file says so.

**Correction to the numbers above, 2026-07-28 (not an edit — the earlier text stands as what I
believed at the time).** The two sections written before the review say "61 of 61 false alarms"
and "124 became 63". After dropping the two page types the reviewers questioned, the shipped
figures are: **59 false alarms removed, 124 became 65.** The two that come back are one quoted
market-share sentence appearing on two robot-hands index pages — a known false positive I chose
over an unknown blind spot. If you quote a number from this file, quote these.
