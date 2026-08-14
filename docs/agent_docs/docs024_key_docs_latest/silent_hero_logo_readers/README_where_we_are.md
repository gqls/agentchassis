# Where we are — making the silent hero/logo readers speak up

Plain prose, append-only, newest at the bottom. The owner maintains this too — add below, never
rewrite above.

---

## 2026-08-11 — the small job that turned out to be the important one

This is commission item 2, the one you approved with "2. yes." It is the smallest of the five and
it has quietly become the one that unblocks the big one.

**The problem in one sentence.** Three places in the code take the web address of a freshly
deployed hero image or logo, and when that address isn't there they do nothing whatsoever — no
complaint, no note, nothing. So a page that shipped with no hero and no logo looked exactly like a
page that never wanted one. That is how this went unnoticed for five weeks.

**What I changed.** Those three readers now say something when a deploy actually happened and its
result came back without the address. They say which key they wanted and — this is the useful part
— which keys the result *did* carry, because that list is a fingerprint: seeing
`response / response_status / response_received_at` and no address tells you immediately that this
is the known fault and not something new.

**They stay silent when no image was wanted**, which matters more than it sounds. Most pages
deploy no hero and no logo, and the naive version of this change would have filed a complaint on
every page of every site — thousands of them, burying the handful that mean something. So the test
is "did something actually try to deploy an image", and the presence of the result is what answers
it. There is a test that fails if that gate is ever removed, and I proved it fails by deliberately
breaking the gate and watching it go red.

**One judgement call I want to flag, because it goes slightly beyond what you approved.** The
commission asked for a log line. I've added a log line *and* a durable record in the database. The
reason is that a log line here would not survive long enough to be read. Two measurements, neither
of them mine:

- the service these readers run in is the busiest we have, and its own start-up line was already
  measured as having scrolled out of reach within hours;
- the run record that holds this evidence is deleted after **four hours** — because it sits in the
  "waiting for a reply" state, which is the shortest-lived category in that table. That is also
  the real explanation for something that had puzzled us: this fault kept appearing and vanishing
  (nothing on the 6th, two cases on the 9th, nothing on the 11th). It was never coming and going.
  It was one four-hour window opening and closing.

So a log-only version would have reproduced the exact problem it exists to solve: evidence that
evaporates before anyone reads it. The database table I'm using is the one the codebase itself
documents as the only one that outlives this kind of step. I've put this at the very top of the
council submission rather than hiding it in the small print, so if the reviewers think I overstepped
they will say so plainly.

**Why this is now the unblocker.** Item 5 (last week's job) fixed the diagnosis tool, and the
re-run proved the tool works and the evidence is gone. Chasing the evidence by looking for it is a
race against a four-hour deletion, and we have lost that race three times. This change stops
looking and starts catching: the next time it happens, it writes itself down at the moment it goes
wrong, into a table nothing prunes on that schedule.

**Something I found on the way, which is not mine to fix.** There is already a workaround in the
deploy code — added back in February — that writes the address to a second place precisely because
someone had noticed the first place gets overwritten. And the codebase separately records that this
*second place is also wiped* for exactly this kind of awaited step. If both of those are true, that
is a real candidate for the root cause we have been unable to pin down, and it has been sitting in
two comments that nobody had read side by side. I have written it into the bug file as a lead
rather than acting on it, because the fix is the design decision you have reserved for yourself.

**An incident worth knowing about, since it is about how we work rather than what we built.** While
I was mid-verification, another session committed my four files into the repository under its own
commit message — it swept up everything in the shared workspace, including work in progress. Nothing
was lost. But the timing was uncomfortably close: minutes earlier I had been *deliberately breaking*
one of those files to prove a test could catch it, and if the sweep had landed then, a knowingly
disabled safety check would now be in the codebase under someone else's commit, with my own notes
showing a clean test run. I checked instead of assuming — the version that landed is the correct
one, and a clean copy of the repository builds and passes. I have written the lesson into the
runbook: when you break something on purpose, restore it and verify the restore in the same breath,
not at the end of the session.

**Where it stands.** The code is in the repository and will go out with the next fleet release. It
is submitted to the council. The one thing that will actually prove it works is a real site build
that deploys a hero or a logo, after the release — and then a record appearing at the moment it
fails. I cannot force that from here.

---

## 2026-08-11 (evening) — approved, and the reviewers earned their keep

The council approved it, first round, in about eleven minutes. No serious objections. Several
reviewers went out of their way to say the judgement call I flagged — adding the durable record on
top of the log line you approved — was the right way to handle a scope decision: declare it and let
someone rule, rather than slip it in.

**Two reviewers pushed back in ways that were genuinely useful, and neither could be answered by
just agreeing.**

The first caught me repeating a number instead of checking it. The four-hour deletion window is the
fact my whole argument rests on, and I had taken it from another session's notes. The reviewer's
point was blunt and correct: it should not be treated as settled just because it was stated with a
specific number. I went and read the live cleanup job. **It is four hours** — the claim holds, and
it is now something I have measured rather than something I inherited. Worth recording that the
reviewer had no way to check it either; it flagged the gap rather than waving it through.

The second asked a harder question: I fixed the three places the bug report named — but is the same
silent-failure pattern lurking elsewhere? Rather than promise to look, I counted. There are 64
places in that package using the same shape. Almost all are reading configuration, not results, so
they are irrelevant. Four are genuinely reading the result of a dispatched job, and I opened all
four. **None of them has this bug.** Two treat a missing value as "then do the other thing" and go
and do it. One fails outright with a clear error. So the real distinguishing feature isn't the code
shape at all — it is whether a missing value has any consequence beyond the page quietly coming out
worse. That was true in our three, and false in the other four.

Four more I could not classify, because they look up a key that comes from configuration rather
than being written in the code, so you cannot tell what they mean by reading them. I have written
those down as unfinished rather than counting them as clean.

**One reviewer suggestion I did not take**, and I want to be upfront about it: it asked me to leave
a note in a particular shared table. There is a proper tool for doing that for one category of note
and none that I could find for this one, and hand-writing rows into a shared table is exactly the
habit that tool exists to prevent. So the reasoning lives in six places a reader will actually look
instead. I have logged that as an open loose end, not as done.

**Still outstanding:** the release, and then a real site build that deploys an image, which is the
only thing that will prove any of this works in anger. And separately, the lead I mentioned — I have
put that to the automated diagnosis loop rather than write it up as fact, and its answer is still
pending.

---

## 2026-08-12 — it's live, and the automated diagnosis told us something better than an answer

**Item 2 is live.** The new build went out overnight and I checked the running service itself rather
than trusting the version number — the new record type is compiled into both copies of the service,
and I ran the check alongside a deliberately impossible search to prove the check isn't just saying
yes to everything.

**It has not fired yet, and I can tell you why that means nothing.** There are no records — but
also **nothing has deployed a hero or a logo since the release**. So the path hasn't had a chance to
run. That is the difference between "nothing broke" and "nothing was tried", and it is only visible
because I asked the second question alongside the first. A quiet result here is not yet good news;
it is not news at all. It still needs a real site build.

**Now the part worth your attention.** That lead I put to the automated diagnosis loop came back
"unverifiable" — but read what it could not do, because it is more useful than an answer would have
been. It said, in effect: *I cannot see the code you are asking me about.* It was given one function's
body and a single line of another, and nothing at all for the two that mattered.

So I checked whether the code was missing from our searchable index. **It isn't.** All four functions
are there in full, with correct line numbers. The index had everything; the evidence pack the loop
actually reads passed on almost none of it.

**That is the same fault I fixed last week, one layer over.** Last week's job was that the pack
listed the columns of one database table while showing rows from six — and, crucially, never said it
was showing you a filtered view, so "not there" and "not included" looked identical. This is that
exact shape again, in the code half rather than the data half: it holds four functions and shows
one, and the loop is instructed to abstain rather than guess when it cannot cite something. So it
abstained. Correctly.

**And I have to own a mistake here.** Last week I wrote in the bug file that this very blocker was
"clear", and my evidence was that the index was fresh and carried the functions. That was true, and
it was an answer to the wrong question — the loop had complained about the pack, not the index. A
fresh index says "present" whether or not the pack passes it on, so my check could never have come
out any other way. It has cost one diagnosis run to be told the same thing again. I have corrected
the bug file and logged it in our wrong-calls log, including the check I should have run: read the
pack itself, which was sitting in the database the whole time.

**Where that leaves the lead:** neither proved nor disproved. It is still the best explanation we
have for both halves of the original bug, and it is still marked unverified. Chasing it again through
the same loop will fail the same way until the code half of the evidence pack is fixed — so that is
now the thing standing in the way, and it is a fault in our own diagnosis tooling rather than in the
image code.

**What I'd suggest next**, though the choice is yours: fix the code tier of the evidence pack. It is
the same shape as the fix that already went through review and worked, it unblocks this lead, and it
unblocks every future question whose answer lives in a function body — which is most of them.

---

## 2026-08-12, afternoon — I fixed the evidence pack, and it was a spelling mistake between two halves of our own tooling

I said yesterday that the thing standing in the way was the code half of the evidence pack. I have
now found out why it was broken, fixed it, and put the fix through review. Here is what it turned
out to be, because it is simpler and sillier than I expected.

**Two parts of our system write down the name of a function in two different ways, and nobody had
ever checked they agreed.**

When we index the codebase, a function that belongs to a type gets written down in the style Go
programmers use in documentation — with the type in brackets in front of it, like
`(*SagaCoordinator).applyResponseToState`. That is the name that goes in the index, and it is the
name the evidence pack shows you when it lists search results.

But the piece of code that actually goes and fetches the text of a function expected the name
written a different way, without the brackets. So when the diagnosis loop asked for a function using
the only name it had ever been shown, the fetcher looked for it, did not recognise the form,
and reported "symbol not found" — which reads exactly like "that function does not exist".

The loop is built to abstain rather than guess when it cannot cite evidence. So it abstained. The
function was in the index the whole time, all 4,746 characters of it.

**The scale of it surprised me.** I counted every diagnosis run we have ever done: 335 function
bodies lost this way, across 47 separate runs. Every indexed method in the codebase — 1,170 of
them — was unfetchable by its own official name. There was a second, smaller version of the same
fault too: package-level values (things like limits and error messages) have been in the index since
a change in July, but the fetcher was never taught to look for them, so all 1,238 of those were
invisible as well.

**The detail I found most uncomfortable is that our own written guidance caused it.** We keep a file
of traps for the team, and yesterday somebody added an entry saying, in effect, "when you ask about a
method, be sure to use the bracketed name, or you will be told it does not exist." That advice is
correct for searching the index. It is exactly wrong for fetching a function body, because those are
two different pieces of code that never shared a rulebook. Following our own instructions was what
walked into the fault. I have corrected that entry and written the plain rule underneath it.

**Why it survived so long:** the test we had did check this, but it checked the name written in the
style *nothing in the system actually produces*. So the test and the code were wrong in the same
direction, and passed each other. That is the failure I want us to notice more than the bug itself.

**Two things I got wrong today, in the open.**

The first is that I wrote a claim into the review submission before I had checked it. I said all
twenty of the odd cases were of one type. I had recognised them by looking at their names. When I
ran the actual query ten minutes later, nineteen were — one was a different problem entirely (a
function too new for the snapshot we had indexed), and my fix does not cover it. So the honest
figure is 334 of 335, not all of them. The annoying part is that the whole point of that sentence
was to be the one that could have proved me wrong, and I published it before letting it. I have
logged it.

The second is smaller but says something about how we work: I checked that bug number 260 was free,
spent half an hour writing the file, and by the time I saved it another session had taken 260 for
something else. So the fix's commit message points at a bug number that now belongs to a different
problem. Ours is 261. On a tree this busy, "I checked it was free" has a shelf life of about half an
hour.

**Where this leaves us.** The fix is written, tested, committed and submitted for review. It is Go
code, which means it does nothing until the next time we build and roll the images — so the fault is
still live in production right now, and still costing us: fourteen more function bodies were lost in
the forty minutes it took me to write this up.

**What I need from you is the same choice as yesterday, now better informed:** which piece of the
commission to take next. Item 1 is the big investigative one whose design decision you have
reserved; item 3 is the medium one that needs a routing call. Neither is blocked. My own suggestion
is that once this fix has rolled, the cheapest valuable next move is to re-run the diagnosis on the
original hero-and-logo bug — it has failed twice for want of exactly the function bodies this fix
restores, and it is the last unanswered question in that bug.

---

**2026-08-12, evening — the bug that was hiding behind yesterday's bug is now fixed too.**

Quick recap of where this sits, because the chain matters. Bug 261 was that our diagnosis tool
could not read function bodies when they were written in a particular style — it asked for them,
got nothing back, and gave up. We fixed that. And the moment it could ask properly, it ran straight
into a second problem that had been sitting behind the first one all along, invisible because
nothing ever got far enough to reach it. That is bug 267, and it is the one I have just fixed.

**What 267 actually is.** When the tool asks to read a file that is too big to fit in its working
space, it says so — honestly, with the numbers. And then, in the same sentence, it tells the model
to *ask for that file again on its own*. Which cannot work. The file was 169,000 characters and the
space is 60,000. It does not matter how you ask; it will not fit. So the model does as it is told,
spends one of its three attempts finding out, and by the time it has worked out the right thing to
ask for, it has no attempts left. In the case that started this, it named exactly the four functions
it needed — they would have fitted four times over — and had nowhere to put the request.

Measured across every one of these bundles we have ever produced: six wasted attempts across five
investigations that came back with **no code at all**.

**What I changed.** The tool now checks the arithmetic it already had in front of it before it
offers advice. If the thing genuinely would fit on its own, it still says so, in exactly the same
words — that mattered to me, because the easy mistake here is to delete the sentence and quietly
make the common case worse. If it would not fit, the tool says so plainly and then does something
more useful: it names the largest pieces of that file that *would* fit, with their exact sizes, so
the model can ask for those instead.

**Two things I found that were not in the bug report, and one of them is the important half.**

The report named two places that gave the impossible advice. There were four. The fourth is not a
sentence at all — it is a piece of logic that says "we have already shown you this whole file, so
there is nothing else to show you from it". Which is false in precisely the case where we could not
show the file. So the one file the model most needed a map of was the one file we withheld the map
for. Fixing the advice without fixing that would have moved the dead end rather than closed it.

I found it by asking what the model does *next* after we turn it down, rather than by searching for
where else the sentence appeared. Worth remembering — searching for the string finds the instances
you already know about.

**One mistake, caught by reading the output rather than by a test.** My new sentence told the model
that the rest of the file's contents were listed further down the page. They are not, quite — that
section lists functions only, and my count included a couple of other kinds of thing. Small, but it
is a false statement in most files, and no test I had written would ever have caught it. I only saw
it because I printed the finished text and read it as the model would. Narrowed the wording.

**One thing I deliberately did not fix.** There is a related inconsistency — the same function can
appear under two different names in one report — and my change makes it more visible than it was.
It is one line, in a file I already had open, and I left it alone: it belongs to bug 261, which
explicitly recorded it as a separate job. Folding it in would have been me overruling someone
else's decision quietly, in a commit about something else. I have written down that my change
strengthens the case for doing it.

**Where this leaves us.** Same shape as yesterday: written, tested, committed, submitted for
review. It is Go code, so it changes nothing in production until the next time we build and roll
the images — the waste is still happening right now. And the same choice as before is still open on
which piece of the commission to take next; nothing here has changed that, except that the
diagnosis loop will be meaningfully less wasteful once both of these have rolled together.

---

**2026-08-13 — the review came back approved, but it took three goes and it caught me out twice.**

The 267 fix is through the council gate: approved, twelve of thirteen reviewers, one advisory note
left over. That took three rounds and I think the two rejections were worth more than the approval.

**What the reviewers found that I had missed.** The sharpest one first. My fix has two halves, and the
second half only works if one piece of information reaches it. Every test I had written checked that
half by handing it that information *directly* — which proves the half works and proves nothing about
whether the real system ever hands it over. A reviewer asked the obvious question I hadn't: what if it
doesn't arrive? The honest answer was that the fix would quietly do nothing, and **not one of my tests
would have noticed** — which is the exact failure this bug is about, reproduced inside my own fix. I
wrote the missing test, then deliberately broke the wiring to confirm the test bites. It does.

That is the second time in two days something on this lane has been wrong in the shape of "the test
and the code agreed with each other and both were wrong". I am starting to think the useful question
isn't "does my test pass" but "what would this test do if the thing it checks were never wired up".

**The one I got wrong myself, and it was worse because of how I got it wrong.** A reviewer asked me
to say how we'd confirm this fix was actually live once we ship it. I told it its information was out
of date and quoted the newer method: ask the running service which version of itself it's running.
That method does not work on this particular service — it announces its version once when it starts
up, and this service is so chatty that the announcement scrolls out of the log within minutes. We have
a written warning about exactly that, dated the day before, naming this service. And the document I
was quoting says so itself, four lines below the bit I quoted.

So I read half a paragraph, skipped the half that named my own service, and used it to correct
somebody. Two reviewers caught it independently. Everything else in that submission was carefully
evidenced — the one unchecked sentence was the one where I was telling someone else they were behind.
**Confidence in that particular register is what stopped me opening the file.** I've logged it, and
I've rewritten how we verify this fix: instead of asking the service what it is, we look for something
in its output that only the new code could possibly produce — plus a second check that tells us
whether the situation even arose, because "we saw none of the new behaviour" means nothing if nothing
triggered it.

**One thing four reviewers agreed on and I had got wrong in judgement rather than fact.** I had
noticed that my change added a third copy of a small piece of naming logic, and I filed it as a future
job. Four separate reviewers said the same thing: it's two lines, extract it now, don't file it. They
were right, and my reason for deferring didn't survive being said out loud — it was an argument for not
touching *someone else's* code, which I'd stretched into an argument for not tidying my own. Done now.

**What's left, and it isn't nothing.** The remaining reviewer note says this is the fourth bug in a row
against one small underlying mechanism, and asked for that to be visible to whoever plans the bigger
pieces of work rather than sitting as a cross-reference in a bug file. I've written it up as RFC_027.
It needs a decision from you eventually, and the decision might well be "four bugs on something this
small is acceptable, stop filing it" — that's a legitimate answer and I've said so in the file.

**And the caveat that has not changed since yesterday.** All of this is Go code. It does nothing in
production until we next build and roll the images. The waste is still happening right now.

---

**2026-08-13, afternoon — the first fix is live and closed, the second one is written, and I caught
myself over-correcting yesterday's mistake.**

**267 is done.** The new build went out, I checked whether our fix was actually in it, and it is. That
sounds routine and it wasn't, because of what happened yesterday.

Yesterday I told a reviewer that the way to check "is my code actually running" — ask the service which
version of itself it's running — was the current method, and I was wrong: on this particular service the
announcement scrolls out of the log within minutes. I logged that as a mistake and wrote down that the
method **doesn't work here**.

Today it worked, first try. The pods were seven minutes old and the announcement was still there. **So
my correction was also wrong**, in the opposite direction. The truth is narrower than either version:
the method works if you ask soon after a deployment, and there is a one-line check that tells you
whether you're still inside that window. I had read that check yesterday and not carried it into my own
note.

I've written this up as a second mistake rather than quietly fixing the first, because the shape is the
interesting part: **both times I took a single measurement of something time-dependent and turned it into
a permanent property of the service.** And the over-correction was the more expensive one — "it never
works here" would have sent the next person to a different check that the same warning describes as
returning "not found" even when your code *is* in the binary. My tidy-up pointed at a worse tool than the
one it dismissed.

**One thing that worked exactly as intended.** Yesterday I paired the "is it live" check with a second
check asking "did the situation even arise?". Today the first check came back empty — and the second
explained why: no diagnosis has run at all since the deployment, so there was nothing for it to see.
Without the pair, that empty result would have read as "the fix didn't ship". That distinction is the
whole difference between an honest inconclusive answer and a confident wrong one, and it cost about two
lines to build in.

**269 is written, and the measurement turned out to matter more than I expected.** This is the one where
our tool could hand the model the *wrong function* — not an error, just quietly the wrong piece of code,
labelled as the right one. I said yesterday it was unmeasured, so I measured it.

Of the 1,175 methods in our codebase, **48 sit in a position where this could fire** — that's 4.1%. But
4.1% understates it, because in the worst spots *six* different types share one method name, and there a
wrong answer is five times more likely than a right one. And the file I'd most want to be reliable is on
the list: **the diagnosis tool's own source code.** Both of the investigations we ran this week were
investigations *of that file*. Either could have been handed the wrong function and we'd never have known.

I checked one thing before believing any of it: that the data actually stores names in the format my
query assumes. It does — all 1,175 of them, none otherwise. If it hadn't, the query would have produced a
confident number that meant nothing.

**Where things stand.** 267: live, verified, closed. 269: fixed, tested, committed, in review — and like
everything else it does nothing until the next build and roll. The remaining piece is RFC_027, which asks
whether the underlying naming machinery deserves one proper owner rather than four bug fixes; that's a
decision for you and "no, four is acceptable" is a fine answer.

**2026-08-14.** The overnight build carries the fix, and I checked that properly rather than assuming
it: the running binary on both replicas says which commit it was built from, our fix is inside that
commit's history, and a commit made after the build is absent — so the check could have failed and
didn't. Then the behaviour: the first diagnosis bundle since the roll listed twelve method references,
and every one of them is in the new unambiguous form, none in the old form. The old code would have
written all twelve the old way, so that's real evidence, not a vacuous zero.

One honest caveat, written into the closed file rather than smoothed over: nothing since the roll has
touched a file where two functions actually share a name, so the part that picks the right one of two
has test proof but no live sighting yet. The part that stops the ambiguous references being offered at
all — which is the defect we filed — is proven live. You approved closing on that basis this morning,
so 269 is closed. That's the whole chain done: 261, 267, 269.

Then I went to the hero/logo bug this was all in aid of, and found something embarrassing in a useful
way: the "cheapest next move" we wrote down on the 12th — ask one narrow question about one function —
was never actually run. The retry that evening asked the old broad question again, and then died
without producing an answer, and nothing recorded why; the evidence of what went wrong has since been
cleaned away by the database's own housekeeping. So I fired the narrow version today. The answer should
land within the hour and will go in the bug file.

Still waiting on you: the RFC_027 ruling — whether the naming machinery that produced four bugs gets
one proper owner, or whether four bugs on something this small is acceptable and we move on.

**2026-08-14, an hour later.** The answer came back in half an hour, and it's the one we suspected
since the 11th but couldn't prove: when a workflow pauses to wait for another service, the pause
throws away everything the current step had just worked out and not yet saved. The function that
writes the pause reloads the record from the database and copies across only three bookkeeping
fields — everything else the step computed in memory is simply not part of what gets written. The
line that briefly made us doubt this turned out to be a look-don't-touch check, not a rescue.

And this time it isn't just code-reading: two live workflows, caught mid-pause, both show exactly
the predicted hole — the waiting step's data missing, every finished step's data intact. It also
explains why a workaround someone added back in February has never once worked: its output was
being thrown away by the same pause it was trying to survive.

So the hero/logo bug finally has its cause. What it needs now is a decision from you rather than
more digging: the repair sits on a seam every waiting workflow in the platform shares (the RFC_012
question that's been parked). Fixing it there fixes a whole family of silent losses, of which the
missing hero and logo images are just the two we happened to catch.

**2026-08-14, afternoon.** A fresh build went out this morning (nothing of ours in it — checked it
anyway, same method as before, both replicas). With ten diagnosis bundles now through since the fix
rather than one, the picture is firmer: not a single ambiguous method reference in any of them, and
not a single wasted "read the whole file" iteration either. The one thing still unseen live is the
rare case where two same-named functions collide in one file — no investigation has happened to touch
such a file yet; the test suite covers it. Next piece of real work on this lane, when a session picks
it up: the list-of-neighbours a bundle shows is capped at about ten per file, and we have a recorded
case of that cap hiding exactly the three functions an investigation needed. Your two decisions
(RFC_012 — the pause that loses work; RFC_027 — one owner for the naming machinery) are still open.
