# SUMMARY — the claims checker learns what kind of page it is reading

**2026-07-29** (the work ran through the evening of 07-28). First summary in this
workstream's series. Bug `bugs_closed/102`. Session "bugsearch 7".

---

## What we're trying to do

We publish sites that make claims about our business — how many records we hold, how
many clients we serve. We once published invented figures, so we built a checker that
reads every page before it goes live and asks, of every number: is this asserted as a
fact about the business, and does anything in our evidence register support it? If
not, the page does not ship until a person has looked.

That checker has to be trusted to be useful. A checker that cries wolf gets ignored,
and then it is worse than not having one — because everybody assumes it is watching.

## Where we've come from

The checker decides whether a number is a business claim by looking at the words
around it: clients, customers, records, users, uptime. That is a reasonable question
on a sales page, where nearly every sentence is about the business.

It is the wrong question on a page that is *teaching* something. When a games-design
article explains drop rates and writes "in a game with 10,000 active players farming
that item, roughly 180 of them will hit that wall", the checker sees a large number
surrounded by business-sounding words. The sentence is a worked example. There is
nothing to verify and nothing to correct.

Another team had filed this in July as a nuisance blocking one site: webdesign.co.uk
could not be given an evidence register at all, because switching it on would have
raised fifteen complaints about fifteen pieces of perfectly good teaching copy.

## What we've done

**We measured it first, against the live system, before writing any code** — and the
bug turned out to be considerably larger than the report said.

Nine of our sites have the checker armed. Between them they were carrying 124 of these
findings. Fifty-nine sat on teaching pages, and every single one of the fifty-nine was
a false alarm. The findings on ordinary business pages, meanwhile, were the real ones
the checker exists to catch.

It was also not merely noisy. A finding of this type is filed at a severity that stops
a page being rebuilt, so four blog posts on gamesdesign.co.uk simply could not be
regenerated, because probability examples inside them looked like lies about the
company.

The fix gives the checker the one signal it was missing and which we already record:
what kind of page this is. On teaching pages — guides, blog posts, news listings,
interactive tools — it stops guessing at numbers in prose.

**The care in the fix is in what it does *not* stop.** We also keep a per-site list of
specific claims a human has audited out as false. Those are still matched on every page
of every kind, because the very first time this checker ran for real, what it caught
was one of those banned claims sitting in a guide. Figures printed in stat cards are
still checked everywhere too. Only the *guessing* stops, and only where guessing was
measured to be wrong every time.

We put it through the reviewer council before committing. It was approved — and two of
its three comments were right, and narrowed the fix. I had included two page types on
reasoning rather than evidence. One of them, "section-index", turned out to cover
gamesdesign's **about** and **contact** pages: ordinary marketing pages filed under an
index-sounding name. Both came out. The lesson is worth keeping: **an "-index" suffix
is a naming convention, not a guarantee about what is on the page.**

## Where we are now

Live and verified. You deployed chassis v1.0.1196; I checked the running pods rather
than the tag or the build log, and the new code is in both, with a control check to
prove the test itself works.

The measured result, over an identical export before and after: **124 findings became
65 — fifty-nine false alarms gone, and nothing new appearing.** The two that remain
from the narrowing are one quoted market-share sentence on two index pages: a known
false positive, which we chose deliberately over an unknown blind spot on an about
page.

One claim is written down as an inference rather than a fact, because that is what it
is: I can prove the fix is in the running binary, but not *which of my two versions* —
the difference is two words that appear in several other files anyway, and our images
carry no revision stamp. The build ran fourteen minutes after the final commit and
builds always take committed code, so it is almost certainly the narrowed one. The case
file says exactly that, in those words.

The ticket is closed and moved to the closed-bugs folder. webdesign.co.uk can now be
given its evidence register — the obstacle this bug describes is gone.

## Where we're going

Three things are recorded rather than quietly dropped:

1. **Report pages still throw fourteen false alarms**, for an entirely different reason:
   product model numbers ("Schunk EGP 40-N-S-B") read as figures. Folding them into this
   fix would have been a coincidence rather than a mechanism.
2. **Teaching the checker to recognise tutorial phrasing** ("let's say", "for example")
   is still worth doing, and is the only thing that helps a worked example sitting on an
   ordinary business page.
3. **Nothing yet watches for a page being filed under the wrong type in future.** The
   reviewers were right to press on this. I checked it by hand today and found the two
   pages named above. That is a check, not a control, and the difference matters.

One thing worth carrying beyond this bug. Mid-afternoon, another session was editing the
same four files for a related fix. Its commit took my half-finished edits along with it —
normal here, and nothing was lost — but it took the callers of my new code without the
code itself, so for four minutes the shared codebase did not compile, and anyone starting
an image build in that window would have failed. I found it by building the committed
code in a clean directory rather than by reading the diff. **A green test suite tells you
your own filesystem compiles; it says nothing about what ships.** That is now written
into the debugging guide, with the three-second check that catches it.
