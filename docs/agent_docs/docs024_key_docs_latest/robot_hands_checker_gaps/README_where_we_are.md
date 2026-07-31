# Where we are — robot-hands checkers and the two visible defects

Plain-prose log, append-only, newest at the bottom.

## 2026-07-30 (evening)

You pointed at two things on robot-hands.com: the Tools link in the nav goes to a
page with nothing on it, and the Run MatchMatrix card is missing its picture while
the three cards beside it have theirs. You asked whether the framework could pick
these up itself.

Short answer: **it already catches the missing picture, and it cannot see the
empty page at all. But neither finding was going anywhere, because two of the
three stages that turn a finding into a fix are not running.**

**The empty Tools page.** The page is switched on, it is in the header nav, and it
has three sections planned for it — but it has no components at all, so there is
nothing to render. What people are actually being served is a file that was
deployed on **10 May**, nearly three months ago. Nothing has retried it since.

I checked whether any of our roughly sixty checkers could notice this, and none
can. Three of them come close and each misses for a different reason: one only
looks at components that exist, and here there are none to look at; one only fires
when a page has *no* sections planned, and this page has three; the third only
looks at pages already marked as deployed, and this one is marked as needing a
rebuild. So the page falls between all three. I then proved it rather than just
arguing it — I ran all three checker lanes tonight, all fifty-seven configured
checks, and they filed **nothing at all** about that page.

Across the whole fleet there are fourteen pages in this state on six sites.
Robot-hands' is the **only one reachable from a nav**, which is why you spotted
this one and nobody spotted the other thirteen.

**The missing picture.** This one the framework got right. The MatchMatrix tool is
the only one of the four with no card image on file, and the card list is built
automatically from the tool pages — so the framework fills in a blank rather than
a picture. Before changing anything I replayed the relevant checker read-only, and
it flagged exactly that one page and nothing else. Then I ran it for real and it
filed the right job: derive MatchMatrix a card from the image it already has. The
image itself was generated back on 25 July; what never happened was the second
step that cuts the card from it.

**Why nothing was being picked up.** This is the part worth your attention, and
it is bigger than either defect.

Getting from "something is wrong" to "something is fixed" has three stages.
Finding it, running the finder on a schedule, and handing the finding to whoever
fixes it. Only the first works.

Of the twenty-four scheduled jobs currently switched on, **not one runs any of the
three checker lanes.** The only such job ever created is switched off, aimed at a
different site, and pointed at a message queue nothing listens to. So the checkers
only ever run when a person sets them going by hand, which is what I did tonight.

Worse, findings are filed in a state that means "noticed", and the fixers only
ever pick up work marked "approved for action". The step that promotes one to the
other is run by three agents, and **none of them is scheduled either** — the main
one was switched off on **2 May**. The result is **263 findings sitting untouched
across nine sites, the oldest from 14 July, and not a single item anywhere in the
"ready to fix" state.**

So tonight's sweep found the MatchMatrix problem correctly and filed it correctly,
and it will sit there like the other 263 unless something promotes it.

**What I have and have not changed.** I have run the three checker lanes against
robot-hands and switched the one-shot jobs back off afterwards. That created
findings; it changed nothing on the site. I have deliberately **not** promoted
anything into the fixers yet, because that is the step that starts rewriting live
pages, and the queue for this site includes a couple of dozen items I have not
reviewed.

One thing I noticed while there and did not touch: the quality lane raised six
"unverified claims" on this site tonight, and two of those are on the gripper
dossier report pages, which are supposed to be exempt from that check because
every figure on them is specific to one visitor's request. That looks like a
false positive belonging to the dossier workstream rather than something to fix
here.

**Next**, per your three decisions: build the missing detector with a handler that
rebuilds the page, fork the features block into a carousel for robot-hands and
make it the planner's default for new sites, and put both through the council gate
with concept-register entries.
