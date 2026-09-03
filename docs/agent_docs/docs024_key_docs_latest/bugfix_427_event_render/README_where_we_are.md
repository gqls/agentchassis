README — where we are, bugfix_427_event_render (plain prose, append-only, newest at the bottom)

2026-09-02. Picked up bug 427 — boxingonline.com's fight calendar page had a
banner and some text describing a calendar, but no actual dates or fights
listed. The underlying problem: nothing on the platform ever turns "a
promoter announced a fight" into a piece of structured, correctable data the
site can show, and even where that data existed, nothing would put it on the
page.

Good news early on: another session (calling itself "feed lane") had already
started building the first half — the bit that reads a news article and turns
it into a registered fact. I checked their work rather than duplicating it,
found one real mistake in it before it shipped (they were about to label
these facts with a tag the system doesn't recognise, which would have quietly
broken some validation), told them, and they fixed it. A third session found
why the site's page planner sometimes skips building pages it clearly should
build — turns out that's a deliberate (if debatable) choice the AI planner
makes for itself, not a bug in the code, and it's now its own separate ticket
(428).

My part ended up being two things. First, a small but real gap: once these
"fight is confirmed" facts exist, the part of the system that writes summary
sentences about them (for marketing copy) would have printed the placeholder
text literally — things like "{event_date} at {venue}" showing up word for
word instead of an actual date and place. Fixed and shipped.

Second, and this is the main piece: even once the facts exist, nothing was
built to actually show them on the page. I built that — a way for the page to
pull in "what's the site's list of upcoming fights right now" and display it,
using the same mechanism the site already uses for showing news headlines and
business directories, rather than inventing something new. It's tested and
sent off for review.

What's left, and why I'm not doing it yet: the actual fight data doesn't exist
on boxingonline.com yet — the other session's populator still needs one more
step (a config change to the AI prompt that does the extraction) before real
facts start appearing. Once they do, there's a small follow-up piece — making
sure the calendar page actually asks for this new data — that I'll do then,
because building it now would just show an empty list, which isn't actually
fixing anything.

Both my changes are sitting in the platform's review queue (an AI council that
checks platform code changes before they're trusted) — I haven't heard back
yet, will note it here when I do.

---

2026-09-03. A new session picked this up where the last one left off, plus a
second, related job (deploying a frontend for the sister ticket, 428). Here's
where things actually stand now.

The review council came back on the "show fights on the page" piece three
times, not once — the reviewers found two real, good-faith problems (would a
schedule claim with no source citation be trusted blindly; would an incomplete
fact just vanish with no record) and both got fixed properly rather than
argued away. It's approved now.

Then came the part that was left open on purpose last time: actually putting
the fixture list on boxingonline's live page. The council flagged something
important on the FIRST attempt at this — it said "you're claiming the only
way to do this is a full page rebuild, and that's risky on a page a real
customer is already using; have you checked there isn't a smaller, safer way?"
That turned out to be exactly right. There WAS a smaller way — a targeted
"swap one section for another" tool the platform already has, built and
safety-tested for exactly this. I used it instead of the risky option. It
worked: the page's old, never-actually-shown "how we build this calendar"
text block is now the actual fixtures list, and nothing else on the page was
touched. I confirmed this properly — not just "the database says it worked"
but "I fetched the actual live page and the actual GitHub deploy log and saw
the real file change."

One genuine gap remains, and I want to be honest that I could not close it
this session: the fixtures list is showing on the page, but it's showing
EMPTY — "no confirmed fixtures yet" — even though there is one real, dated,
sourced fight in the system right now (Canelo Alvarez vs Christian Mbilli,
October 31st) that should be showing. I tried three times, including after a
brand-new version of the software was deployed, and each time the mechanism
that's supposed to pull that fight onto the page just... didn't. I couldn't
find why by reading the code — the code looks right, and I couldn't catch it
misbehaving in the logs either. This needs someone to sit down with a
debugger, or the platform's own "please investigate this properly" process,
rather than another guess from me. I've written down exactly how to reproduce
it so whoever picks this up doesn't have to rediscover the steps.

I also fixed a smaller thing this uncovered: after swapping the section, the
page's own internal "list of what sections this page has" still mentioned the
OLD section name. Left alone, that would have confused the platform's own
housekeeping into thinking something was broken and trying to rebuild the
whole page from scratch — the exact risky thing we were avoiding. Fixed with a
small, careful database change, checked before and after.

Separately, the OTHER ticket (428, the one about the site-planner skipping
page types it's supposed to build) needed a small admin webpage update
deployed so a human can actually click a button that already existed in the
backend. I built it, and I checked — properly, by looking inside the actual
built files, not just trusting the build succeeded — that it had the right
content. Then I hit a wall I don't control: actually publishing it is a
"push this to production" action, and this session's own safety settings
require a person to say yes to that, so I stopped and asked. By the time this
note is being written, it's live — either you or another session finished it,
and I've double-checked the live version really does have the right content.

So: the fix that visitors will actually notice (something instead of nothing
on the calendar page) is live but incomplete (empty rather than showing the
one real fight). The admin tool for 428 is live and usable. What's left is
one real, well-documented mystery about why the fixture data isn't flowing
through, and the ordinary next-step housekeeping (get that council review
result on the actual "swap the section" step, since I only just told them
about the safer approach they should check).
