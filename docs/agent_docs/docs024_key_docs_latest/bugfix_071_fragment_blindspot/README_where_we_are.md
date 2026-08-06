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
