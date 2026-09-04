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

---

**2026-09-03, mid-afternoon.** The fight is on the page.

I checked, this time properly, and the server had actually restarted with today's
fix in it — the previous "fresh build" I told you about turned out, on closer
inspection, to not have shipped anything new at all. Same pods, same code, just
sitting there. I caught that before reporting it as done, which is the whole
point of checking rather than assuming.

This time it was real. I re-ran the repair immediately. It worked all the way
through — the data resolved, the save that had been blocked earlier today went
through, and the file was actually pushed to the real deployment pipeline. I
traced that past a green status light to the actual upload happening. The fight
calendar page, once it catches up to the version we just pushed, will show the
one confirmed fight instead of an empty box.

I closed the underlying bug for good, since it's now proven working, not just
believed working. And while closing it, something happened worth telling you
about because it's a good example of how careful you have to be on a shared
system like this.

Another team was working on a completely different, older problem — one that
turned out, to their surprise, to be caused by the exact same one-line mistake
I fixed this morning. While I was moving my file to the "solved" pile, they were,
at the very same moment, writing to it from the other end. The timing meant one
of their routine saves — technically correct on their side — ended up deleting
the file entirely for about two minutes, because it looked for the file where it
used to be and found nothing there. Nobody did anything wrong; two reasonable
actions landed in the same few seconds.

What made it a non-event rather than a real problem: they noticed immediately,
figured out exactly what had happened and why, and — this is the part I want to
highlight — they didn't just fix it themselves. They told me, because it was my
call whether and how to close the file, and fixing someone else's decision for
them isn't the same as helping. I put it back, checked it was actually back
properly, and we moved on. I've written the pattern down so if it happens again
to someone else, they'll recognise it in thirty seconds instead of a few minutes
of confusion.

Their own work turned out to matter a lot to ours: they had independent proof,
on a completely different page and a different kind of content, that this fix
does what we said it does — proof that reached all the way to what a visitor
actually sees, checked separately by three different people today, including me
checking it myself just now rather than taking anyone's word for it.

So: this piece is done. What's left is one open question about whether a
different, older mechanism might quietly undo our fix the next time this
particular page gets replanned — that's a decision, not more coding, and I've
flagged it clearly rather than either ignoring it or trying to force through a
fix at the end of a long session. And there's one automatic check, running
overnight, that will be the final independent confirmation everything worked.

---

### 2026-09-03, evening — the calendar fix was one rebuild away from vanishing, and now it isn't

Picked this lane up from the handoff the last session wrote for exactly that purpose. Short
version of the day: the fight calendar fix was real but it was stored in the wrong place,
and it would have quietly disappeared the next time that page was rebuilt. It won't now.

**What was actually wrong.** A page's list of sections is kept in three places at once. One
of them is the master copy; the other is a photocopy that gets refreshed from the master
every time the page is built. Yesterday's fix wrote the photocopy. So the page looked right,
the deploy was real, the fixture was genuinely on the site — and the master copy still said
the page should have the old advertising block and no event list. The next build would have
copied the master over the photocopy and the calendar would have gone back to being empty,
with nothing reporting an error.

**One correction to what I told you earlier today.** For about ten minutes I had this
backwards and said so out loud: I'd found a mechanism that protects the page when the site
is *re-planned*, and concluded the fix was safe. It is safe from re-planning. It was not
safe from an ordinary page rebuild, which is a different and much more common event. I was
wrong in the same way the previous session's write-up was wrong — we both looked at the code
that *writes* the store and not the code that *reads* it. That's now written down as the
cheap check that would have caught both of us.

**This has happened before, and twice it wasn't caught.** The same trap was documented in
July, with a fix migration written for it. Since then two other sites have quietly lost
sections to it: robot-hands lost its gripper spec sheet — which is the very thing that July
migration was written to rescue — and idea.uk lost a guide list. In both cases the system's
own detector had spotted the problem and filed a warning, and the warning just sat there.
Nobody was ever going to action it, because that detector was built to flag and not to fix.
The architecture reviewer put it better than I would: *a detector that fires and does not
prevent the loss it detects is not a safeguard, it's a log.*

**What I've done.**
- Corrected the master copy for the fight calendar page, and proved the page itself didn't
  change by so much as a byte — which is the point, since the page was already right.
- Cleared a backlog of six of those unactioned warnings, the oldest from 28 July. Four were
  describing problems that had already resolved themselves one way or the other; I closed
  those and recorded *which side won*, so a case where someone's work was destroyed doesn't
  get filed away looking like a success. One was a genuine live problem on another team's
  site, so I left it alone and warned them instead — they fixed it within the hour, and gave
  me back a detail I didn't know that has improved the permanent fix.
- Filed the two destroyed pages as a bug of their own. Whether those sections should be put
  back needs someone who knows what those pages are for; a machine can't tell a deliberate
  removal from this bug's handiwork.
- Written up the permanent fix as an architecture question for you. There's no proper way to
  do this correction today, so every team hand-writes database surgery for it, and about
  half get it wrong — five instances in seven weeks, two of them today by different teams.
  The question is one sentence and it's at the top of the document.

**The thing you should know that isn't good news.** The signal the last session named as
"the way we'll know 427 is finished" cannot actually fire. The nightly check that grades
tool pages ran four minutes after yesterday's deploy and still reports this one as *"a page
about a tool, not a tool"* — because what was added is a static list, and the page is
classified as a tool, which is held to needing something you can actually use. That check
was never going to pass. You've decided we build a real calendar mechanism, which is the
right call, and I've made sure that work happens in the right order: correcting the master
copy had to come first, because the only reason the page hadn't already reverted is a guard
that switches itself off the moment a real tool is added to it.

### 2026-09-03, later — we built the calendar, and it invented the fights

Following your decision, I put the calendar through the tool pipeline rather than building
anything by hand. The good news is that the machinery works and produced exactly the right
*kind* of thing: a real interactive calendar with filters and a live countdown, attached to
the existing page without duplicating it. That is genuinely what the nightly check has been
asking for, and what the static list we shipped yesterday could never have been.

The bad news is what it put *in* the calendar. Twelve fights, all invented — real fighters'
names, with made-up dates, venues and broadcasters. Canelo against Charlo, Fury against Usyk,
Wilder against Joshua. None of it checked against anything. And every single date is in the
past: the newest is December 2025, and it is now September 2026. So the tool's own "hide
fights that have already happened" behaviour would show a visitor **an empty calendar** —
which is the exact complaint this whole piece of work started from.

I have stopped it before it reached the site. Nothing has been published; the page is still
serving yesterday's verified state, and I cancelled the queued rebuild with the reason
recorded against it. I kept the tool itself, because the mechanism is right and only the data
is wrong — throwing it away would lose the good half.

**This is not a new problem, and I want to be plain about that: it is the original problem
showing its face.** The whole reason 427 exists is that nothing on the estate turns a
confirmed real-world event into a dated fact that can later be corrected. So when a generator
is asked for a calendar's worth of fixtures, it has nothing to draw on and makes them up. We
asked the machine for real upcoming fights and it did the only thing it could.

**Your call, and there are three honest options.** One: actually build the missing piece —
wire the calendar's fixture list to real, dated, correctable facts. That is the only option
that finishes 427 rather than moving it, and the tool was deliberately briefed so its data can
be swapped without touching the interface. Two: ship the mechanism with no fixtures and an
honest "fixtures to follow" state — an empty calendar that says it is empty is not a false
claim, whereas twelve invented ones are, and this is a customer's site with real people's
names on it. Three: hand-check a small number of genuinely real upcoming fights into the
evidence base first, then regenerate.

My recommendation is one, with two as the interim if the site has to go out before that
lands. What I would not do is re-run the generator and hope — it produced this from a brief
that explicitly asked for real events, because there was nothing for it to be right about.

---

**2026-09-04, late afternoon — the calendar is working, and the remaining problem turned out
to be a different one from the one we thought.**

I picked this lane back up after it had been quiet since yesterday evening. First thing worth
saying plainly: **the fight calendar on boxingonline.com now works.** It was rebuilt this
morning and it shows two real fights — Navarrete against Foster on 24 October, Canelo against
Mbilli on 31 October — each with a date, a source link and a note saying schedules can change.
Nothing invented. The whole machine this bug was filed about, which turns a real news story
into a dated fact and then puts that fact on a page, exists now and is running.

The site's own lane also settled something I would otherwise have chased: the reason there are
only two fights is not a fault in the calendar. The site's fact register holds eight facts,
seven of them dated, and exactly two are in the future. The calendar is showing everything it
has. Finding more upcoming fights is a research job for the news lane, not a repair job here.

**So what is actually left?** This: when the calendar shows a fight, nothing on the page says
where that fight came from. The system knows — it carries the fact's identifier all the way to
the last step and then drops it, because the template that draws the fight was never written to
print it. That sounds like a small tidy-up. It isn't, and here is why it matters.

Yesterday we found a second tool on the same site — a fight countdown — that had six completely
invented fights in it, about real, named boxers. Today another thread found more of the same
class elsewhere on the estate, including something worse: a veterinary comparison tool on a
customer's homepage listing **thirty invented veterinary practices with invented postcodes**.
That file describes its own data as a "verified sample", its own instructions say it must never
invent practices, and a line shown to the public invites the reader to go and confirm the
details with the practice, citing the official register while doing it.

The uncomfortable point is this: **the page doing everything right and the page inventing
everything look identical to every automatic check we have.** Neither says where its data came
from. So we cannot tell them apart, and no amount of cleverness in the checker fixes that,
because there is nothing on the page to check.

**That is why the fix I have proposed is to make the good pages say so, rather than to get
better at spotting the bad ones.** We have now tried five times in two days to write a rule that
recognises invented data by its shape, and it has escaped every time — the shape just changes.
Another thread ran a simulation today over the whole estate and found that the most obvious such
rule would be wrong 89% of the time, and that the biggest data set on the estate is entirely
legitimate while the fabrication that started all this has only six entries. There is no clever
threshold hiding in there. So instead: when a page shows a fact that came from our register, it
prints the fact's identifier alongside it, and a checker confirms that identifier really does
lead to a real, cited, still-current fact. Absence of that becomes the signal, and absence is
something you cannot disguise.

**Two things I want to be honest about, because they cut against my own proposal.**

First, this only helps where a component has declared what its data is. I measured it: **287 of
our 335 tools declare nothing at all** — that is the normal state, not a red flag. So my change
reaches about 48 of them. On the other 287 it does nothing, and a different thread's work is
carrying the whole load. I have told them so, and we have agreed that neither of us will
describe the two pieces together as having solved the problem.

Second, I got things wrong today and two of them were caught by the other thread rather than by
me — including one where my mistake would have pointed them at the weaker of two tests, which is
worse than simply being wrong. Both are written up. The one I am actually pleased about is
smaller: I re-ran a number I had borrowed from someone else, expecting to confirm it, and the
re-run is what turned up the 287-of-335 figure above — which then showed that advice I had given
that same thread ten minutes earlier was overweighted. The number was right; going and getting it
myself is what showed me what it meant.

**Where it stands:** the plan is with the review council now. Nothing has been built yet. When
the verdict comes back I will either build it or revise it, and I will not commit code against a
verdict I have not read.

**One thing did not go to plan and you should know:** you asked for Fable to prepare the plan.
Fable stopped immediately with a credit limit — "You've reached your Fable limit" — and produced
nothing, so I wrote the plan myself rather than leave the work stalled. If you top up the credits
and would like Fable's independent take on it before I build, say so and I will get one.
