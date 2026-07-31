# Where we are — the three unguarded deletes (bugs_open/165)

Plain-prose log, append-only, newest at the bottom.

---

**2026-07-31, evening.**

Background, in one paragraph. A few of our build steps work by deleting
everything they previously wrote and then writing back what they just produced.
That is a sensible way to keep things tidy — it stops stale rows piling up
forever. It goes badly wrong in exactly one situation: when the thing producing
the new content only managed to produce *some* of it. Nothing errors, because a
short answer is still an answer. So the delete removes everything the short run
did not replace, and what you are left with is not a broken page that shouts at
you — it is a page that quietly has less on it than it used to, which looks
exactly like a page that never had more.

We fixed one of these last week (the code index). The fix deliberately built the
*rule* as a shared, reusable thing, and deliberately did not go around converting
the other three places that have the same problem. When that fix went through our
review council, one of the reviewers objected — reasonably — that leaving the
other three is itself the mistake we make most often: one place gets the careful
fix, its siblings stay exposed. So the previous session wrote the objection up as
its own piece of work and left it for someone to pick up. That is this.

There are three left. Today I did the one that matters most: the table that holds
the actual sections of every page on every site. It is the one that has genuinely
lost customer content before — twice that we know of, including an interactive
game that a routine content rebuild simply deleted, on two different sites months
apart.

The interesting part of today was not writing the guard. It was choosing what
"enough" means, because a guard that fires when it shouldn't is worse than no
guard at all — the first person it blocks unnecessarily will delete it, and then
we have neither. So I measured before deciding, and two of my first instincts
turned out to be wrong.

The first wrong instinct was written into the task itself: check each *slot* on
the page separately. It turns out almost every slot on every page holds exactly
one section, so checking per slot means "if you removed one section, refuse" —
which would have blocked 89 perfectly ordinary edits over the last four months.

The second one was mine, and it is the one worth remembering. I had the guard
compare what a rebuild produces against what the page *plans* to contain — a good
second opinion, because the plan is written by different code entirely. Running it
across every page, three tripped. One of them was the idea.uk homepage: plan says
six sections, and my query said it had two. I wrote that down as the guard finding
real damage.

It was not damage. The page has all six sections. Four of them are **locked** —
someone deliberately marked them as not-to-be-touched by automation. My count was
only counting the ones a rebuild is allowed to change, which is correct, but I
read "2" as "this page only has 2 sections". The consequence was in the guard, not
just in my notes: it meant a *perfect* rebuild of that page would have been
refused, because it compared "2 sections written" against "6 sections planned"
when four of those six were never on the table. In other words, my guard would
have blocked every future rebuild of exactly the pages someone had cared enough
about to protect. Which is precisely the way these things get ripped out.

The fix was small once seen — don't count the locked ones in the target. After it,
the check trips on none of the 238 pages it can actually reach. What caught it was
opening the six rows instead of trusting the one number, and I have written that
up because the number was not wrong, it just answered a slightly different
question from the one I was asking it.

Where that leaves things. The guard is written, tested (seventeen tests, and I
broke it four different ways to confirm the tests actually notice), committed, and
submitted to the review council. It is **not yet proven working in production** —
our code only takes effect after the next deployment, and more importantly a
successful run proves nothing here, because the guard is designed to do absolutely
nothing when everything is healthy. The only real proof is to deliberately break a
rebuild, watch it refuse, confirm nothing was deleted, then undo the breakage and
watch a normal rebuild go through. That is the next job, and it needs the roll
first.

The other two sites are still open. One of them I deliberately did not touch,
because another session had that file open at the time and landing a change
underneath them on a shared tree causes exactly the sort of mess this project has
already had enough of.

One thing I needed a decision on and got it: the machine's temporary disk was
completely full — 13GB of leftover scratch from 125 past sessions — and commands
were failing mid-run. I cleared the ~4GB belonging to sessions that had been
finished for over six hours, leaving anything recent alone.
