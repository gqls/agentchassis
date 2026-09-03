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

---

**2026-09-03, later that morning.** The mystery is solved, and it turned out
not to be our mystery at all.

The calendar page wasn't showing the one real fight because of a bug in a
completely different piece of the system — the bit that re-renders a page when
its data changes. Yesterday lunchtime another session tidied up that code:
split one long function into two smaller ones, which is normal, sensible work
and their notes about it are unusually careful. In the split, one value stopped
being handed across from the first function to the second. In Go that isn't an
error — the second function just quietly gets an empty value instead — so
nothing complained, nothing crashed, and nothing looked wrong.

The value that stopped being handed across was the freshly-looked-up data. So
since lunchtime yesterday, every "re-render this page" across the whole estate
has been taking the page's own previously-stored content and rendering that
back at itself. It never breaks anything and it never blanks a page — it just
silently achieves nothing. There is no error message for that. The only way to
notice is to know what a page *should* now say and find that it doesn't, which
is exactly the position we were in yesterday with the fight calendar.

I found it by doing the first of the two things yesterday's notes said I hadn't
tried yet — actually reading that function line by line instead of searching it.
It took about four minutes. One search command told the whole story: the value is
read in one place and set in none.

I wrote a test first, before fixing anything. The test builds the real calendar
component and the real fight fact, runs the re-render, and — with no cluster
involved at all — produces exactly the empty "no confirmed fixtures yet" box
that's on the live page. That's what turned "I've found something odd" into
"this is definitely the thing we've been chasing". Then the fix, which is one
line. With the line, the test passes; without it, it fails. I checked that
against the properly committed code rather than my own working copy, because
someone else's half-finished work was in the shared tree at the time.

Two honest corrections to what I wrote yesterday, both now written into the
notes rather than quietly tidied away. First, I said I couldn't find any sign of
the data-lookup code running, and concluded it wasn't running. It was — I just
wasn't capturing its output properly. I read a gap in my own instrument as a
fact about the code. Second, I framed the question as a choice between two
explanations and the real answer was a third one I hadn't thought of.

How big is this? I measured it today: 1,855 live page sections across 838 pages
use data that gets looked up rather than written by an author. Every one of them
has been getting nothing from a re-render since yesterday lunchtime.

One thing I have to own. When I committed my one-line fix, the same file already
had someone else's unfinished work in it, and the way we're all told to commit
takes whatever is in the file. So their half-finished change went in with mine
and the shared code temporarily stopped compiling. I worked out exactly which
six files would fix that, checked it, and messaged the session that owns them —
rather than committing six files of somebody else's in-progress work on my own
guess about whether it was ready. That's their call, not mine.

What this means for the calendar page: nothing we built for it was wrong. The
component, the schema, the evidence check, the fact itself, the way it was
attached — all correct, all in place, and all being fed into a pipeline that had
stopped delivering three hours before we attached it. The fix is in the code but
not yet running: this kind of change only takes effect when a new server image
is built and rolled out. Once that happens the page should fill in on its own,
with no further work from us.

I've also sent the fix to the reviewer council, as we do for anything touching
shared platform code, and I've written the whole thing up as its own bug (454)
so nobody has to rediscover it.

---

**2026-09-03, late morning.** Three things since the last note, and one of
them is me finding a mistake of our own.

The reviewer council approved the fix. Twelve of thirteen reviewers said yes,
and the one objection was fair and was about my paperwork rather than the
code: I had written two tests and only showed them one of them, so they quite
reasonably asked where the other was. It was there all along. The lesson is
one our own runbook already spells out — the reviewers only see the summary
you give them, so a summary that undersells the work draws an objection the
work didn't deserve. Four lines would have avoided it.

Second, I resubmitted the older review that had been sitting unanswered since
yesterday — the one about the calendar component itself. Its blocking
objection was that we'd claimed there was no safe way to attach the component
except a full page rebuild, without checking. The reviewer was right, we did
check afterwards, there was a safer way, and we used it. So the resubmission
answers them with what actually happened rather than with a better argument.
I also widened it: the original only asked them to review the safe part
(adding a component to the library), leaving the two changes that actually
touched the live site unreviewed. On this system a database change IS the
running system, so reviewing only the safe third isn't really reviewing it.

Third, and this is the one worth telling you about: while writing that
honest account I found a bug in our own change from yesterday. When we
updated the page's list of sections, the database command we used rebuilt
that list without preserving its order. The list went from
[hero, old-section, advertising] to [advertising, hero, calendar] when it
should have been [hero, calendar, advertising]. That order matters — three
different parts of the system use "the Nth entry in this list" to identify
"the Nth section on the page", so after our change the first section on the
page was matching against the wrong name.

I want to be straight about the size of it: this was not causing damage
today. The one place that would have tripped over it only does so for
sections that have no name of their own, and both of ours have names. It was
a trap waiting rather than a fire burning. But it was silent in both
directions — a wrong-but-present name doesn't error, it just quietly matches
nothing — so it would have surfaced one day as an inexplicable failure with
no obvious cause.

Fixed, with the order restored. Before applying it I rehearsed the whole
thing inside a transaction that I then rolled back, to check it did what I
expected and left nothing behind, and then I deliberately made it write the
WRONG order to confirm the safety check would actually catch that and refuse.
It did. That second step is the one that's easy to skip and it's the only one
that proves the safety check works at all.

Two things I deliberately did NOT do, so they don't look handled: there's
still an entry in that page's list ("advertising") pointing at something the
page doesn't have — that pre-dates all of this work, so I left it exactly as
found and wrote down that it's there. And I haven't checked whether other
pages across the estate have the same ordering problem. I fixed one page and
I'm claiming nothing wider than that.

The calendar page itself still can't show the fight until a new server image
is rolled out. Nothing more to do on it here.

---

**2026-09-03, early afternoon.** The new server image went out, and the
answer is: our part works. The fight now resolves.

I ran the re-render as soon as I'd confirmed the new image actually contained
this morning's fix — and I want to flag that confirming it was not a
formality. The obvious way to ask "what is the server running?" gave me the
wrong answer: it returned a temporary worker pod still running yesterday's
code, and if I'd trusted it I'd have concluded the fix hadn't shipped and
stopped. Asking the two long-running pods specifically, and checking they
agreed with each other, gave the right answer.

The re-render did what we've been waiting two days for. The calendar section
picked up the Canelo/Mbilli fixture — one item where there had been none —
and its HTML grew from 1,813 to 2,498 bytes. I checked that against a
snapshot taken minutes before, not against memory.

There was a bonus I wasn't looking for, and it's the more telling result.
The hero banner section on the same page quietly regained its background
image settings in the same pass. Nothing about our work touched hero images.
That section had been losing its data to the same one-line bug, silently,
along with everything else on the estate — so this is the "1,855 affected
sections" claim from this morning showing up in front of me rather than in a
spreadsheet. One line, two sections, two completely unrelated kinds of data.

**But the page still hasn't changed, and I need to correct something I told
you this morning.** I said it would fill in on its own once the fix shipped.
It didn't. The last step — writing the result back — was refused by a
different safety check that went live in the very same image, put there by
another team working on a different problem. It blocks changes to pages that
are marked as "tool pages" but have no tool component on them, on the
grounds that rebuilding such a page would publish waffle about a tool that
doesn't exist. Our calendar page technically matches that description, so
the check is doing what it says.

I have not tried to work around it, and I don't think we should. I measured
how far it reaches — 53 live, currently-serving pages across 9 sites,
including 16 on the loan-and-mortgage calculator site — and sent that to the
team who own it, because they'd asked to be told if it got in our way. It's
their decision, not mine.

The unlucky part is pure timing and nobody's fault. Until this morning, a
re-render on those pages ran, said "success", and quietly achieved nothing.
So refusing to save its result cost nothing you could see. Both changes
landed in the same image, and now the same refusal blocks a repair that
would actually do something.

So: everything we built is proven correct, the data flows, and one decision
by another team stands between that and the page a visitor sees.

The lesson I'm taking, having now hit it twice in one day: "it'll work once
X lands" is a prediction about a whole chain, and I had only checked our own
link in it.
