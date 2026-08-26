# Where we are — pages we retired that are still on the internet

Plain prose, append-only, newest at the bottom. The owner maintains this too — add below,
never rewrite.

---

## 2026-08-26 (afternoon) — what the bug is, and what I found when I went and looked

When we decide a page should no longer be published, we mark it "archived" in the database.
That mark does two things and, crucially, it does not do a third. It takes the page out of
every list the site builds from, and it stops anything rebuilding it. What it does **not** do
is remove the page from the internet. The file we last published is still sitting on the
server, and anyone with the link — or Google — still gets it.

Removing the file is a separate action. We built that action back in early August. Nothing has
ever checked that it ran.

So there is a gap between "we retired this page" and "this page has actually gone", and nothing
anywhere looks into that gap. That is the whole bug. It was filed four days ago by the thread
that noticed it while closing a different piece of work, and nobody has picked it up since.

I went and measured it this afternoon rather than taking the filing on trust, because the filed
numbers were four days old and this is exactly the kind of number that moves.

**There are 39 pages we have retired that were published at some point. Seven of them are still
being served to the public right now.** They are on six different sites, including two on
robot-hands.com and two on fundamentallyai.com. One of them — robot-hands.com's gripper catalogue
— was already recorded as being in this state on the 14th of August. That is twelve days, and
nothing raised a flag, opened a ticket, or wrote a note in all that time. Not because anything
failed. Because nothing was looking.

I want to be careful about how I measured that, because a naive version of this test gives the
wrong answer in two opposite directions. Some domains are configured to answer *every* address
with a page, so "the page loads" would look true for a page that genuinely does not exist. And
if a site is simply down, every page looks absent — which for this particular question reads as
good news when it is the worst possible news. So for each site I also asked for a made-up
address that must not load, and a page we know is live that must load. Every site passed both.
That is what makes seven a real seven, and thirty-two a real thirty-two.

The other thing worth saying plainly: **this is a flow, not a backlog.** Two of the pages the
original bug report caught have since been taken down, and five of the seven I found today were
not in its list at all. So going in and tidying up these seven by hand would look like a fix,
change nothing about the mechanism, and leave us exactly where we started in a fortnight. What
is missing is not a cleanup. It is a meter.

What I am proposing to build is that meter: a check that runs as part of the routine sweep every
site already gets, asks the question for each retired page, and raises a visible flag when the
answer is wrong. Deliberately a flag and not an automatic deletion — a page that was archived by
mistake and is serving perfectly well is a real possibility, and quietly deleting a good live
page is the kind of failure that is worse than the bug it was fixing. A person should look at the
first ones.

I have asked for an independent design pass on the detail, and it will go through the reviewer
council before it ships.
