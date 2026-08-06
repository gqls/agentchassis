# Where we are — fragment blind spot (plain prose, append-only, newest at the bottom)

## 2026-08-06

We went looking for the next open bug nobody was working on. The picker that
worked was measuring how "warm" each bug is in the other sessions' live
transcripts, then confirming the coldest candidates by searching for the actual
function names a fixer would have to touch. Two bugs that looked free by every
other test turned out to have sessions actively working right next to them; the
function-name search is what saved us from starting a duplicate.

We landed on the last unowned piece of bug 071. The story of 071: the platform
detects broken links on every build, and for a while it detected them and then
shipped them anyway. Most of that is fixed now. The piece nobody took is the
"jump to a section" kind of link — `#pricing` at the end of a URL. Nothing in
the platform ever checks that the section a link jumps to actually exists, and
for a while nearly every such link on the estate jumped nowhere.

First finding: the damage has quietly healed since the bug was filed — today's
fragment links (66 of them) all work, mostly because pages got repaired by hand
and the writing agent got given a list of real pages to link to. But nothing
GUARDS this: the next page build can ship dead section-links again and nothing
would notice, which is exactly how it happened three times in a row on one site
in July.

The plan: teach the existing post-deploy link checker (which already runs, and
whose findings already get worked) to also check the section half of every
link, reusing the id-detection logic another check already paid to get right;
and add one sentence to the writer's instructions so it stops inventing section
links in the first place. No new check to schedule, no new queue to forget.
Deliberately NOT doing yet: blocking builds on this, auto-deleting dead
fragments, or making the platform stamp ids on every section — each is
recorded with reasons, and each needs its own decision.

Next: write the code and tests, prove on today's data that the new check would
raise zero false alarms (and one real one when we plant it), put it through the
council, commit.

## 2026-08-06 (later)

Done and committed. The check now has a fourth thing it looks for: a link whose
"#jump-to-section" half names a section that does not exist on the page it lands
on. It sits inside the link checker that already runs and whose findings already
get worked, rather than being a new check somebody would have to remember to
switch on — we have been bitten by that exact thing before. The writer has also
been told to stop inventing section links, since nobody ever gives it a list of
real ones.

Before shipping it I ran the new code over every page on every site to see what
it would complain about: nothing, on all 67 of today's section links. Then I
planted two broken ones into a copy of the same data and it caught both — a
"zero problems found" that has never been shown capable of finding a problem is
worth nothing.

The council approved it first time. One reviewer asked a question I had not
thought to ask myself: I had rearranged a shared piece of code, and it turns out
a third place uses it — the gate that refuses to publish a broken tool. Changing
that by accident would have been bad. I checked properly by running the old and
new versions side by side over 4,036 real documents; they agree everywhere. My
first attempt at that check was itself worthless (both versions agreed that
nothing was wrong with anything, which proves nothing), and only a guard I had
written into it beforehand stopped me reporting it as a pass. That is the third
time today that asking "could this test have failed?" changed the answer.

What is left: the new code only starts working when the next chassis image is
built and rolled — this is Go, so it is inert until then. After that roll,
someone needs to confirm it is in the binary and plant one broken link on a real
site to watch it get reported. The steps and exact commands are in the lane's
RUNBOOK and handoff.

Three related things are deliberately NOT done, each written down with why: the
pre-publish gate still cannot check section links (it cannot see the site's
header and footer, so it would raise false alarms); nothing repairs a broken one
yet; and no page section publishes a stable name to jump to, so these links can
only be avoided rather than made to work. That last one is the real capability
and it needs its own decision, because it changes every page on every site.

## 2026-08-06 (after the roll)

It is live and it works. The build that went out this afternoon carries it — I
checked the actual running binaries rather than trusting the version number, on
both of them, including a control string that should be there and one that should
not.

Then I proved it properly instead of just looking at an empty queue. An empty queue
would have been the *expected* result, because none of our real section links are
broken today, so it would have told me nothing. So I built a page with four section
links — two deliberately broken, two working — on an internal placeholder site that
nobody's browser can reach, and ran the real checker at it. It complained about
exactly the two broken ones and said nothing about the two working ones. Then I
fixed one of the two, left the other, and ran it again: one complaint, the right
one. That is the difference between "the code is in there" and "the code does the
thing". The test page is deleted and I confirmed the site is back as I found it.

A nice accident: the same test page had a "#" link that goes nowhere, and a
*different* checker picked that one up instead of mine. That is exactly the division
of labour I designed, so one test proved both that mine fires when it should and
that it keeps its hands off the neighbour's territory.

One loose end, written down: the bit that double-checks a fix before closing a
problem hasn't been run yet — its database queries are verified but the code around
them only runs when a real one of these gets fixed, and forcing that would have
meant kicking off a page build I didn't want on a placeholder site.

There is now a five-part summary in the lane you can read out to someone, and the
handoff is marked up so a new session can see at a glance that the roll-time checks
are done rather than repeating them.
