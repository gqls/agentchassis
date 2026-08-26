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

> **CORRECTED 2026-08-26 (same evening).** The entry below was committed in `2e52c174b` in a
> way that **replaced this file instead of appending to it** — the afternoon entry above was
> gone from the working tree for about a minute, and the pre-commit hook caught it
> (`readme-not-appended: 54 line(s) removed from the owner's log`). Restored from git in the
> next commit; nothing is lost, and this note stays because a silent restore would hide that
> the append-only rule was broken on the very file whose header states it. What made it
> survivable: the file was in git, and the hook prints the warning FIRST, which is exactly the
> half a `| tail -N` cuts off. **Read the hook output, not just git's summary line.**

## 2026-08-26 (evening) — the meter is built, and what it refuses to do is the interesting part

The detector is written, tested and committed. It is deliberately switched off until the code
is actually running on the servers — turning it on early would break two other working checks,
for reasons I'll come back to.

Here is what it does, in plain terms. Every site already gets a routine health sweep, roughly
every four hours. This adds one question to that sweep: for each page we have retired, go and
ask the internet whether it still answers. If it does, raise a flag a person can see.

**It raises a flag and nothing else, and that was a deliberate choice.** I could have made it
delete the page automatically — we already have the machinery to do that. I didn't, because
there is no way to tell, from the outside, the difference between a page that was retired on
purpose and is wrongly still live, and a page that was retired *by accident* and is quite
correctly still live. Both look identical: a working page at a URL we've marked "retired". The
first needs deleting. The second needs un-retiring. Getting that backwards means quietly
deleting a good page off a customer's website, which is worse than the problem I'm fixing. So
the flag says both things and a person decides.

**The part I want to explain properly is what the check does when it can't tell.**

Every other check of this kind we have reports a *problem* — a broken link, a missing file, a
site that's down. If one of those checks goes blind, it under-reports: it finds fewer problems
than there are. Annoying, but the shape of the failure is honest.

This check is the other way round. Its "problem" is a page that *works*. So if the whole site
happens to be down when the check runs, every retired page looks correctly gone, and the check
reports **nothing wrong**. And "nothing wrong" is exactly what it reports when everything
genuinely is fine. A broken instrument and a clean bill of health would look the same in the
records — which would make the check worse than useless, because people would trust it.

So before it judges anything, it proves its own instrument twice, on each site. It asks for a
page that definitely does not exist and requires a "not found" — because some domains answer
*everything*, and on one of those every retired page would look alive. And it asks for a page
we know is live and requires a "here it is" — because if that fails, the site is down and
nothing else it sees means anything. **If either check fails, it refuses to judge, says so in
a permanent record, and — this is the important bit — is structurally prevented from marking
any existing flag as resolved.** A blind run cannot quietly tidy away real problems.

**One thing I got wrong, because it is the useful part.** Test files here carry a table saying
"break this line, and that test will fail" — the evidence that each safety catch really is
load-bearing. I wrote mine from the design, since I knew which test was aimed at which catch.
Then I actually ran it: eleven of thirteen behaved. **The two that didn't were the two
instrument checks I've just spent three paragraphs explaining.** I could delete either of them
and the test still passed.

Neither test was wrong about what it was checking. Both were passing for the wrong reason: with
the safety catch removed, the run simply tripped over the *next* thing and produced an error
anyway, and the test only asked "was there an error?". So it saw one, and was happy. Both are
fixed — they now set the scene so the damage genuinely could happen, and they insist on knowing
*which* safety catch stopped it. And I've written the mistake down, because a table like that,
written from intent rather than from a result, is a design document dressed up as proof, and
the next person to read it would have believed it.

I've also put the whole thing through the reviewer council, and there is one more thing to do
after the next server update: switch it on, and check it flags the pages we already know are
live and does *not* flag the ones we know are gone. Until it has done both, a quiet result from
it means nothing.
