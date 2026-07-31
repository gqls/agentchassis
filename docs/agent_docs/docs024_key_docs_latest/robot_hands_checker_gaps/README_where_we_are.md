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

## 2026-07-31 (afternoon) — the framework fixed your image, almost

The MatchMatrix card image is **made and live**. The framework did it, not me by
hand: the sweep spotted it, and once I let the queue move, the asset-deployer cut
the card from the picture that was generated back on 25 July. You can fetch it —
it is a real 46KB file at the right address.

**It is still not on the page though, and the reason is worth knowing.** The list
of tool cards on the home page is not written down anywhere; it is worked out
fresh each time the page is built, from the set of tool pages. There are two kinds
of "rebuild" here: a light one that reuses the words already stored, and a full one
that works everything out again. The page had a light one, an hour and a half after
the image landed, so it never went looking for the new picture. The job was marked
done, the file is genuinely there, and the thing you reported is still on screen.
I only caught it because I checked the page itself rather than the job status.

**The new detector for the empty Tools page passed review.** The council approved
it. It is committed, but it cannot actually run until the next time the chassis is
rebuilt and rolled out, and I have deliberately not done that yet — rolling it
would have killed the review while it was still running.

**On the carousel, I changed my recommendation and you agreed.** I was about to
build a new component. I then found we already have one — `hero-card-carousel` —
which does exactly what you asked, has the accessibility work already done
(keyboard, screen-reader labels, a pause button, and it does not auto-advance
unless told to), and has **never been used once**. So we are using that instead of
building a second one to maintain.

Three card pictures are being generated now, one per text block. They are queued
behind the backlog we released earlier, and they come through one at a time, so
this is minutes rather than seconds. I picked the style deliberately rather than
guessing: the site's own imagery settings only specify the blue line-drawing look
for one category of image, and the default is *photography*. Had I asked for
"icons" we would have got photographs that looked wrong beside the existing tool
cards.

**One thing I want to flag about the carousel, because it affects you.** The
component's card text and picture are both treated as things the writer owns. That
means if a full rebuild runs later, it can overwrite the picture paths with
invented ones. So the plan is to supply the card content myself and use the *light*
rebuild, which needs no writer at all — that also means no risk of the copy being
re-invented, which matters on this page specifically because it has previously
re-invented a statistic days after we corrected it. The longer-term fix is for the
platform to resolve per-card pictures from the real asset list instead of trusting
the writer, and that is worth doing properly rather than now.

**A judgement I need to put to you:** the three text blocks are full paragraphs,
and the carousel is designed for a heading plus one short line per card. I intend
to shorten each to one faithful line and link each card to the page that carries
the full detail. That keeps the design honest but does move text off the home page,
so say if you would rather keep the paragraphs and accept a taller card.

I also nearly caused a silent mess: generating three pictures that share one
category would have deployed **the same image three times**, each reporting
success. That is a known logged bug; the workaround is written down and I am
following it, and I will check the three files actually differ rather than trusting
the success message.
