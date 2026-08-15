# Where we are — the contrast bug (122)

Plain-prose log for the owner. Append below; never rewrite or reorder.

---

## 2026-08-06 — picked up the contrast bug, and most of it turned out to be already fixed

I went looking for the next open bug nobody else is working on. The contrast one
(number 122) came out coldest by a wide margin — it was filed on 27 July after you
reported a link on oufe.com as "dark blue on the black background and not easily
readable", and it has essentially been sitting since. The one team that had touched it
recently, the darts site lane, had written in the file itself that the main fix "still
belongs to whoever takes it", so it was genuinely free.

**Before planning anything I re-measured the whole fleet in a real browser**, because
the file's numbers are nine days old and this codebase takes something like fifteen
hundred commits a week. That was worth doing, because the file is substantially out of
date in our favour.

Two of its three named problems are **fixed**. The big one was that every site's
header button had white text hardcoded into it, so whether it was readable was pure
luck depending on how dark that site's accent colour happened to be — five of six
sites failed. That is now done properly: the button takes its text colour from the
site's own palette, and I checked all nineteen stored headers on the fleet — not one
still has white hardcoded. Separately, robot-hands.com's main "Run a MatchMatrix
Query" button used to be white text on a white button, i.e. genuinely invisible. That
has gone too.

The one problem from the original file that is still live is the vonc.com Gauntlet
buttons — purple text on a pink button. That surface belongs to the Gauntlet team and
the file already says to coordinate rather than reach into it, so I have left it
alone.

**What I did find is a different problem, and it is on more sites than the original.**
Twelve of the fifteen homepages I measured still fail, 109 failures in total. They
fall into three groups:

The first is a colour being asked to do two jobs at once. Each site's palette has a
"primary" colour. On the dark sites that primary is a near-black — which is perfect
when it is used as a *button background* with light text on top, and invisible when it
is used as *text* on the page. Seventeen of our eighteen layouts use it as text
somewhere. So the same value is simultaneously right and wrong, and repointing the
palette just moves the failure from one place to the other — the darts lane worked
that arithmetic out in July and correctly decided to change nothing. The real gap is
that we have never given components a "primary, but guaranteed readable" colour to
reach for. We already compute exactly that answer in two places in the code; we just
never offer it in that direction.

The second is the worst single thing I found and **it is in no bug file at all**:
ai-agent-orchestration.com is serving six headings in exactly the same colour as the
background behind them. Not low contrast — identical. That page has been publishing
six invisible headings and nothing noticed. I have deliberately **not** guessed at why,
because on paper it should be impossible, and a confident guess about shared code is
the specific mistake our own guidelines were rewritten to prevent. It is going to the
automated diagnosis loop.

The third is components with white text painted onto a mid-tone coloured button —
finetuning.uk, gaswholesalers.com and gamesdesign.co.uk between them. The frustrating
part: we *built* the fix for this on 27 July. The platform now works out a readable
text colour for exactly this situation and publishes it. **Nothing uses it.** Not one
template, layout, or page anywhere on the fleet references it. So we did the hard part
and never connected it.

There is a decision worth flagging to you. The fix I want to make adds one new
"guaranteed readable" colour variable to every site's stylesheet, computed by the
renderer. Nothing changes until a component chooses to use it, so it cannot break
anything on its own — but it is a shared mechanism, so it goes through the review
council before it ships, and gets registered so the next person can find it.

**Also, good news from the build you just deployed.** The tool that checks live pages
in a browser and files the failures as work items — the thing this bug asked for back
in July — is now fully built and running in the new image. I confirmed it is in the
binary and that the agent is correctly configured end to end. The only thing missing
is that **nothing ever tells it to run**. It has been run by hand exactly once, on
4 August, on one site. So what was described in July as "build the checker" is now a
single scheduling row. That is the cheapest useful thing in this whole file, and it is
the same pattern we keep hitting: something built correctly, then parked behind a
switch nobody turned on.

---

## 2026-08-06, evening — the reviewers caught me out, and they were right

Short version: the fix is written, reviewed, approved and committed. It is **not live
yet** — it needs an image roll, and the half that actually changes the pages hasn't been
applied. Along the way the review process caught a real error in my thinking, and I
found two more bugs I've filed separately rather than quietly folding in.

**What the reviewers caught.** I'd said three of the failing sites had components with
a colour "hard-coded" into them. One of the reviewers pointed out that my plan didn't
actually include any change that would fix those three. That was the thread to pull. I
went and opened the three templates — which I should have done first — and none of them
hard-codes anything. What's really happening is subtler and, once you see it, obvious:

Think of a colour like the brand's amber. There are two completely different questions
you can ask about it, and we only ever had an answer to one:

- *"What colour of text should sit on top of an amber button?"* — we've had a working
  answer to this since July.
- *"The amber itself is being used as the text colour. Is it readable?"* — nothing on
  the platform had ever computed this.

The three failing sites were split across both questions, and my fix only answered the
first one. It would have fixed one site and looked like it fixed three. So the fix now
names both directions, and that's what got approved.

I'd rather that had come from me than from a reviewer, but this is exactly what the
review is for, and it cost one round rather than a wasted deployment.

**Two new bugs, filed rather than folded in.** While checking the above I found:

- `bugs_open/212` — 47 of our 173 shared page components quietly overrule a colour
  decision the system makes for them, usually by writing "white" into the template. On a
  dark section that's fine. On a light one it's invisible text. It accounts for about
  24 of the fleet's failures. It's a much bigger change than the one I'm making, so it
  gets its own file and its own decision — I've laid out four options in there and
  flagged that the cleanest one is also the only one that could break something that
  currently works. **That's a call worth someone else looking at.**
- `bugs_open/211` — ai-agent-orchestration.com, our worst site at 30 failures, has six
  headings painted in exactly their own background colour. I now know the mechanism: the
  last chunk of its stylesheet is simply missing, and its absence knocks out a variable
  that the headings depend on. **I do not know why it's missing**, and I've said so in
  the file rather than guessing.

**On that last one — I sent the automated diagnosis on a wild goose chase.** Twice. The
first run failed because the symptom I wrote described the wrong mechanism entirely: I'd
stated as fact something about that site's stylesheet that turned out not to be in it.
The tool didn't come back and say "you're wrong" — it came back saying "couldn't
determine", which reads like a hard problem rather than a bad question. Worth knowing
about that tool. I refiled with a corrected description and it also couldn't crack it,
so there's a genuine gap there too, not just my error.

**One more thing I got wrong and want on the record.** I'd banked a "before" measurement
so we could prove the fix worked. It covered 10 of our 15 sites — and the five it missed
included the two sites this fix is actually for. So the safety net had a hole exactly
where it mattered, and nothing about the file looked wrong. Fixed; it's now complete.

**Where this leaves us.** The engine half is committed and reviewed. What's left is the
part that changes what visitors see: pointing the affected components and layouts at the
new colours, and switching on the automatic contrast check that's been built and sitting
idle since v1.0.1257 — nothing has ever scheduled it, so it has run exactly once, by
hand. Both are database changes, no rebuild needed. They're specified in detail in the
approved submission and in the handoff.

---

**2026-08-07.** The new engine is live. Somebody else's build shipped it — another
thread rolled a fresh chassis yesterday evening for their own fix, and ours went along
for the ride, which is how this tree works. I checked it properly rather than trusting
the version number: I asked the two running servers whether our new code is actually
inside them, and included a deliberate nonsense word in the same check so that if the
check were broken it would light up too. Real code present, nonsense word absent. It's
in.

The part that changes what a visitor sees is still not done. It's two database updates,
both fully written out in the approved plan. One small thing to flag: the file number I'd
reserved for the first one has been taken by another session in the meantime — no harm,
but it's a reminder that nothing here is reserved until it's committed.

**Then I went back to the bug I filed yesterday, and found I'd got it wrong in two ways.**

The first is embarrassing and cheap to describe. I listed four possible fixes and ranked
them, and I didn't do the arithmetic on any of them. When I did the arithmetic this
morning, the two I'd ranked highest turn out to make the actual broken page **no better**
— one of them very slightly worse. The colour the system would have imposed is almost
exactly as unreadable as the colour it would have replaced, because both are pale and the
panel behind them is pale cyan. The right answer was sitting in the same stylesheet the
whole time: the site already publishes a near-black colour specifically for use on top of
that cyan, and it scores 8.65 where everything else scores under 2. Five minutes of sums
would have caught this yesterday.

The second is more interesting, and it's the reason I'm not just fixing the page.

**The system already found this bug on its own, four days ago, and closed the ticket.**
On 3 August an automated design review looked at that page and wrote a description I
couldn't improve on — it named the section, named the colour, worked out that the site's
"primary" colour is cyan rather than the dark shade the component assumed, and concluded
the white text would be nearly illegible. It even attached a precise pass/fail test. The
ticket was handed to the tool that fixes exactly this, and marked **complete three
minutes later**. Nothing was changed. I can prove nothing was changed, because the
component was last edited ten and a half hours *before* the ticket was even created.

The reason is a mismatch nobody would spot by reading any one file. Two different parts
of the system file tickets under the same label. One of them — a routine scanner — checks
for a specific narrow thing, and it wrote the final "is this fixed?" test. The other —
the design review that found our bug — files under the same label but means something
much broader. So when our ticket came up for closing, it was graded by the scanner's
question, not its own. The scanner's question was "are there any hard-coded colour codes
left?" There weren't, and there never had been: our problem is a *reference* to a colour,
not a hard-coded one. So the grader said yes, all clear, and closed it. The grader wasn't
broken. It answered its own question correctly. It was just never the right question, and
the ticket's own attached test — the one that would have caught this — is read by
nothing.

I checked how widespread this is, and the pattern is stark: **every single ticket the
design review has ever filed on this route has been closed successfully — seven out of
seven. Every ticket that ever got stuck, or is still open, came from the other source —
six out of six.** A source whose work never fails isn't a sign of quality. It's a sign
that nobody is checking it. I've written that up separately as its own bug, because it's
got nothing to do with colours — it will quietly close any ticket whose author doesn't
happen to own the grader.

**Where that leaves us.** The colour fix itself is now much cheaper than I thought
yesterday — the tool to do it already exists and already handles this exact case, so this
is about making it actually run, not about building anything. The two database updates
for the original bug are still the next job. And there's a new, larger question I've
deliberately not answered: the renderer only knows about two backgrounds on a page, and
when a component paints its own, the renderer has no correct answer to give it. That's a
design question rather than a bug, and I've said so rather than picking one.

One process note. I sent both of today's findings to the automated diagnosis tool before
writing them down, as the rules require. The first came back "couldn't determine" — the
third time in a row for this workstream. Worth saying that it wasn't useless: before it
ran out of road it had independently worked out the same thing I had. But three
inconclusive runs in a row is starting to look like something about the tool rather than
something about our questions.

---

**2026-08-10 (afternoon).** The good day this workstream has been building towards. The
twelve page republishes all completed this morning, and I checked every one against the
live site rather than trusting the "complete" statuses — after the false-tick problem we
found last week, that check is not optional. All twelve genuinely changed.

Then I measured everything again: all fifteen sites, same tool, same method as the
baseline we banked four days ago. **Every single one of the ten failures we said this fix
would close is gone.** The orange-on-white links on the gas site, the invisible eyebrow
and links on the robotics site, the two washed-out buttons on the fine-tuning site, and
the near-invisible label on the darts site. That last one went the odd way — the darts
team had already deleted the offending section — but deleted is gone all the same.

Two things worth telling straight. First, the darts site promptly reintroduced the same
class of problem in the section that replaced the deleted one: six new invisible-text
failures, same cause — the brand colour used as a text colour on a dark page. The
difference from last week is that the machinery to fix this now exists and is proven, so
I applied the same one-line-per-rule fix to the new section, checked it against every
other site that uses it (it changes nothing where the colours are already legible), and
queued the darts homepage to republish. Found in the morning, fixed by mid-afternoon.

Second, the weekly automatic check is now switched on — the missing half of the approved
plan. Every site now gets its pages measured once a week, the way I've been doing by hand,
and any real failures get filed automatically to the repair queue. It ran its first sweep
within a minute of being turned on. The darts recurrence is exactly why this matters:
without it, this class of bug only gets caught when someone happens to look.

What remains, honestly stated: the two "By the Numbers" eyebrow labels (games site and
the pitch site) are still failing and this fix structurally cannot reach them — a
component that paints its own background is invisible to the renderer's colour logic.
That, plus about two dozen similar failures, is the design question we've put to a human
rather than answered ourselves. And the repair queue those automatic findings feed into
still has the ticket-closing bug I wrote up last week — tickets from the audit can be
marked done without the work happening — so the loop is armed but its far end is not yet
trustworthy.

---

**Later the same day.** The automatic check I switched on this morning ran for five and a
half hours and I've now turned it off. It did what it was meant to, and it also did
something nobody had spotted it would do, which is the more useful half of the story.

What it was meant to do: find pages needing a rebuild and queue them. It found plenty — the
queue of found-but-not-yet-actioned items dropped from 193 to 25.

What it also did: it kept finding new work faster than the system can do the work. Roughly a
hundred new items an hour arriving, about fifty an hour being finished. So the pile didn't
shrink. It moved from one column to another and got bigger overall — from around 270 items
to 544.

That matters more than it sounds, because of the safety valve I described this morning.
Sites with more than fifty outstanding items get skipped, so the checker can't pile work
onto a site that's already struggling. This morning I moved 226 items aside specifically to
get five struggling sites back under that limit — it worked, and only one site was left over
the line. Five hours later, eight sites are over the line. The checker filled the space I'd
cleared with its own findings and then overshot. The concession I made this morning has been
spent, and on that particular measure we're further behind than when I started.

I also found that something I'd written in my own notes yesterday was wrong, and it was the
thing the decision rested on. I'd said the rebuilding work only happens because this checker
drives it — so switching the checker off would stop the repairs too. That isn't true. The
repairs run on their own at about fifty an hour, and they were running at exactly that rate
this morning before I switched the checker on, and all day yesterday while it was off. So
switching it off costs us nothing in repairs; it only stops new findings arriving. Had I
checked that yesterday instead of assuming it, this would have been a much easier call.

On cost, which I know you care about: the number I'd said I'd watch — calls per hour — stayed
normal all afternoon, comfortably inside the range I'd called healthy. But actual consumption
ran about three times the morning average, because each of this checker's calls is a full
pass over a site rather than a quick look. I was watching the wrong meter. Not alarming for
five hours, but a poor thing to have left running overnight unattended.

Nothing is broken and nothing needs urgent attention. The 544 outstanding items should work
through at about fifty an hour now nothing new is arriving — roughly overnight — and as they
clear, the sites currently being skipped come back under the limit by themselves. The 226
items I parked this morning are all still parked and untouched, which was the thing I most
wanted to be true.

---

**Next morning.** It cleared overnight, exactly as I said it would. The 446 items waiting to
be rebuilt are now zero, 542 pages got rebuilt while we slept, and the count of sites being
skipped for having too much outstanding work has gone from eight back to none — all 22 are
eligible again. The failures went down too, 66 to 15. Nothing was touched to make that
happen; switching the checker off was the whole intervention.

Worth saying why I'm pleased about that rather than just relieved: it was a real test. If I'd
been wrong yesterday — if the rebuilding genuinely had depended on the checker driving it —
those 446 items would have sat there all night untouched. They didn't. So the correction I
made yesterday to my own notes is now confirmed by the system rather than just by my
arithmetic.

Then a less welcome finding, which came out of a promise I'd made to another thread. I'd told
them yesterday that switching the checker on would finally exercise a safety gate of theirs
that had never fired, and asked them to re-check today rather than trust yesterday's reading.
So I re-checked. The traffic did arrive — the gate's population went from 26 items to 44, and
for the first time those include work finished *after* their fix went live. And the gate still
reports nothing.

The reason isn't their bug. There are two different holes here and I'd only understood one.
Their fix covers the case where a checker exists but is checking for the wrong thing — it now
says so out loud instead of quietly approving. But there's a second case: some kinds of work
have **no checker registered at all**, and those just complete silently, which the code
explicitly intends. Fourteen items of that second kind finished yesterday afternoon and
evening, and not one of them was looked at by anything.

That matters directly to us, and it changes a decision I'd written down. I had told them the
226 contrast items I parked would be released when they closed their bug. That was wrong, and
I've withdrawn it. Our contrast items are in the *second* category — no checker registered —
so releasing them would produce 226 tickets that get marked done without anything verifying
the work, regardless of what happens to their bug. The park stays. The thing that actually
releases it is writing a checker for contrast failures, which is a smaller job than I'd
assumed because we already have the contrast measuring tool from earlier in this lane. I've
also told them they're no longer holding us up, so they shouldn't delay closing on our
account.

So the honest position: yesterday's cleanup worked and is finished; and the reason we're
holding 226 items is now better understood and slightly worse than I thought — not "waiting
for someone else's fix" but "we haven't written the check yet".

---

**Later, on your two decisions.** You asked for the automatic checking to come back slowly
rather than by widening the limit, so I've switched on one of the three purpose-built rotas —
the quality one — and slowed its polling from hourly to three-hourly.

Two things about it are worth knowing, and the second is the sort of thing that quietly
embarrasses people a week later.

First, the rota is self-limiting by design, in a way the sweep wasn't. It looks at one site per
poll, and it stamps a site the moment it picks it, so that site can't be picked again for seven
days. That means the *rate* is set by the seven-day cycle, not by how often it polls: 22 sites
over seven days is about three site checks a day, no matter what the interval says. The interval
only controls how quickly it works through the fleet the first time.

Second — and this is the useful bit — **it is switched on and it will do absolutely nothing until
Saturday 16th.** All 22 sites still carry stamps from the day and a half it ran back on the 9th
and 10th before you switched it off, and those stamps have another four days to run. I only know
this because I had the migration count how many sites were actually due and print the answer; it
said zero. Without that, it would have sat there for four days looking busy — polling on
schedule, logging normally, doing nothing — and someone would eventually have "fixed" a rota
that was working correctly.

I considered nudging one site's stamp so we'd get a real cost figure today rather than on
Saturday. I decided not to. The figure arrives on Saturday for free, and spending your credits
to get it four days earlier — four days after a cost surprise — isn't my call to make. What we
lose by waiting isn't the number, it's someone remembering to look, so I've written the date and
the exact queries into the handoff and into the migration itself.

On the second decision, writing the missing check for contrast problems: I've done the
groundwork rather than the work, because there's one genuine choice in it that I don't think
should be made quietly. Measuring contrast properly means measuring what the browser actually
renders, not reading the stylesheet — a colour can be written in the file and never applied. Our
contrast tool is a separate Python program, and this check has to run inside the main Go service
at the moment a ticket closes. So it's either route it through the same browser machinery that
found the problem (most accurate, but there's an existing recorded objection to doing browser
work at that moment, which I haven't yet read), rebuild the measurement in Go (a second version
of a thing we already have, which could disagree with the first and nobody would see), or write
a narrower check that just confirms the exact bad colour pairing is gone. I've written all three
up with the trade-offs and recommended reading that objection first, because if it stands the
choice makes itself.

Everything is committed and the handoff is current, so this can be picked up cold.

---

**12th August, later.** I went and read that objection, as I said I would. It stands, and reading
it changed the answer more than I expected — so the choice I was about to put to you isn't the
choice any more. Nothing is broken and nothing needs doing today; this is a direction note.

Three things, in order.

**First, the objection is bigger than I described it.** I'd said there was "an existing recorded
objection" to doing browser work at the moment a ticket closes. There are three, side by side,
and they're firm: we have deliberately refused to build this kind of check twice before, both
times for the same reason — closing a ticket shouldn't require reaching out to the network. That
rules out my option one. It also rules out options two and three, which I had not expected. All
three of my options measure the page, and measuring a page means fetching it; the third was
narrower in *what it asks*, not in *how it asks*, so it draws exactly the same objection. So the
choice I was going to hand you had no surviving option in it. That's my error, and reading the
file rather than trusting my note about the file is what caught it.

**Second, and more important: we have already decided this, and the reason we wrote down is
wrong.** Contrast problems aren't an oversight waiting for a ruling. There's a line in our code
that says, in effect, "this type doesn't need a closing check, because if the problem is still
there the next weekly audit will find it again, and if it isn't, it won't." That sounds
reasonable. Three things say it isn't.

The first is that it's the same argument you already overturned. On the 8th we found that this
exact reasoning — "we don't need to check now, re-detection will catch it later" — had been
tested and failed: the two cases it was ever asked to catch both slipped through, and five days
later nothing had re-detected them even though the problem was still plainly there. That's what
led to the rule change you approved. The contrast line was written on the 2nd, six days before
that, and nobody went back to it.

The second is that "it won't be found again" isn't an observation, it's a silence. Our audit
files problems; it has no ability to *close* one. So "the next audit didn't mention it" and "the
next audit never looked at that page" produce the identical result, and we can't tell them apart.
We have a written rule about this, in the very file that line leans on: you're not allowed to
conclude something is fixed from the fact that nobody mentioned it. The comparable check we point
to as the good example doesn't rely on silence — it actively re-checks and closes the ticket
itself. Ours copied the exemption without copying the thing that earns it.

The third is simply that it has never happened. Not once. Of the 226 contrast tickets sitting
parked, not one has ever been sent to a fixer, completed, or re-found — no contrast ticket ever
has, in the platform's whole history. So the safety net we're relying on has never been tested;
it's a plan, not a net. For contrast, by comparison, the one problem type that *does* have a
proper closing check has nine completions and all nine carry the evidence.

**Third, there's a cheaper and better option none of us had listed.** Don't check at closing time
at all — instead let the weekly audit close its own tickets. It's already out there every week
with a browser, looking at every page; it's already allowed to do that. It just never learned to
say "this one's fixed now." We already built the shared piece that does the closing, for exactly
this purpose, back on the 2nd — it's careful, it refuses to guess, and it won't close anything
the current run just raised. It sits in the same part of the code as the audit, so it can be
called directly.

There's one genuine gap and it's small. To close a ticket safely you have to know the audit
actually looked at that page this time. Right now the audit reports *how many* pages it covered
but not *which ones*, and you can't work it out from the results, because a page that's been
fixed reports nothing — which is indistinguishable from a page that was skipped. That's the one
case we need. So it needs the audit to list the pages it visited, not just count them. That's the
same small fix another lane made last night for the count, extended from "how many" to "which" —
so the hard half is already done and approved.

**One correction to what I told you yesterday.** I listed the "silent 25-page cap" as still open
and unstarted. It isn't — another lane fixed it on the 11th, it's live, it's been proven with a
deliberately truncated run, and it went through the review board. I had it in our list as a
lesser job to get to later. It's actually the thing that makes the option above possible at all:
before it, we couldn't tell a partial audit from a complete one, so we could never have closed a
ticket safely on the strength of one.

**So the order is now clear, and nothing needs your decision this minute:** teach the audit to
report which pages it visited, teach it to close tickets it confirms are fixed, and only then
unpark the 226. That way, if a fixer closes a contrast ticket without really fixing it, the next
week's audit catches it — which is the safety net we currently claim to have and don't. I have
not written any of it; the next person can pick this up cold, and I'd want the review board to
see it before it ships, because it changes something shared.

---

**12th August, after the new build went out.** The new build changes nothing here — I checked
rather than assumed, because "a fresh deployment" invites you to think something moved. The 226
tickets are exactly where they were, untouched, and no contrast ticket has still ever been sent
to anyone.

I've spent this stretch turning yesterday's recommendation into something the next person can
just build, rather than starting it and leaving it half-done. Two things made that the right
call: the change is small but it touches a shared boundary, so it needs the review board, and
while writing it down I found four traps that would each have cost someone a day.

The two that matter to you:

**The first is that this fix would quietly start closing the parked tickets.** Not a bug — I
think it's the right behaviour, and it's rather elegant: as each site gets its weekly look, any
contrast problem that has genuinely been fixed since would close by itself, and only the ones
still really broken would ever need sending to a fixer. That's better than the plan we had,
which was to unpark all 226 and hope. But it does change what "parked" means, and that should be
your call rather than something that just happens, so I've written it up as a decision for the
review board rather than burying it.

**The second is a safety catch that isn't working on our tickets.** The shared piece that closes
tickets has a guard so a run can't close something it just created. It works by stamping each
ticket with a run identifier — and our contrast tickets don't have one. None of them: nought out
of 226, where three other kinds of ticket I checked as a comparison have one on every single row.
So the guard silently does nothing for us. The fix is to start stamping them, which also brings
them in line with everything else. I nearly wrote this down as a guess based on reading one file;
the count is what turned it into a fact, and the comparison types are what proved the zero was
real rather than my query being broken.

Everything is written up and committed, and the handoff is deliberately detailed enough to pick
up cold in a fresh conversation — the design is settled, the traps are named, and the tests that
matter are listed. Nothing is on fire and nothing needs you today. The one dated thing still in
the diary is Saturday the 16th, when the discovery rota wakes up and we find out what it costs.

---

**2026-08-12 (evening).** The job we settled on this afternoon is built, tested and committed.
Short version: the weekly check that finds unreadable text on our sites can now also **close its
own tickets**. When it re-visits a page and the bad colour combination is gone, it closes the
ticket then and there, recording what it saw and when. That is the piece we were missing, and it
is why we no longer need the thing three separate reviewers had already refused to let us build.

**One thing I want to flag, because I went against the written plan.** The plan I picked up this
morning told me how to work out which problems are "still there": take the list of tickets the
check files, and anything not on it has been fixed. That is wrong, and it took reading the actual
code to see why. We deliberately **don't** file every problem we find. Two kinds get left out: ones
whose cause sits in a part of a site somebody has locked (hands off — we can see the problem but
we're not allowed to touch it), and ones that fall past a cap of sixty per run when a site has a
lot wrong with it. Those problems are still there. They just aren't on the list. Had I followed the
plan, the very first run would have closed those tickets and told us those faults were fixed when
they were not — which is exactly the disaster the 226 parked tickets are parked to avoid. So I
built it the other way round: work out what's still broken from what the check actually **measured**,
before any of the filtering happens. I've written this up in three places and put it to the
reviewers explicitly, because it is the kind of mistake that would have looked like success.

**A decision I made that you should know about.** Once this is running, it will start closing the
226 parked tickets on its own, a few at a time, as each site's weekly check confirms the fault is
gone. I could have prevented that and didn't. The reason those tickets are parked is that closing
them today would mark them "done" with nothing behind it — no evidence, no check, just a status
change. But a closure that comes from the weekly check is the opposite of that: it happens because
we went and looked, and it records what we saw. That is precisely the evidence the park was waiting
for. So the pile drains itself as things genuinely get fixed, and only what is really still broken
is left needing attention. The count is reported every run so it can't drain quietly without
anyone noticing.

**Nothing happens yet.** Two services have to be updated before any of this does anything, and if
only one of them goes out it does nothing at all — deliberately, so a half-update can't produce
wrong answers. I hear a fresh chassis is being built; that is one of the two. Until the other one
(the part that drives the browser) goes with it, this is inert.

**Nothing needs you today.** The reviewers' verdict is still pending and I'll act on it. The one
dated thing in the diary is unchanged: Saturday the 16th, when the discovery rota wakes up and we
find out what it costs.

**2026-08-12, later.** Two things happened with the chassis that went out this evening, and they
are worth keeping apart because I nearly ran them together. First, it **killed the review** that
was running on this work — it was thirteen minutes in when the new version came up, and a review
in progress does not survive that. This is a known hazard with a written remedy, which I followed;
it is re-submitted and running again. Nothing was lost and it needed no code change. Second — and
this is the one that matters — **that release does not contain this work.** The build was cut a
few minutes before I committed, so it went out with three of my commits sitting just behind it. I
checked that rather than assuming it, because "I watched the deployment happen" tells you nothing
about whether your own change was in it.

So the position is unchanged from earlier: built, committed, reviewed-pending, and **not live**. It
goes live on the next release. The genuinely good news from looking is that the two services this
needs both went out together on the same version, seconds apart — so it is one release that carries
both halves, not two separate things to coordinate.

**2026-08-12, end of the evening.** The reviewers approved it. Thirteen of them looked, and I
checked that none of their reviews had been cut off or failed to render before I took the approval
at face value — an approval from reviewers who couldn't read the submission would look identical
from the outside.

Five advisory notes came back, and **one of them caught a real gap in my homework.** I had been
careful to list everything that READS the data this change produces. I never listed everything that
WRITES the kind of ticket it closes. Those are different questions, and the second is the dangerous
one: if some other part of the system also files contrast tickets, my change would have been
closing their findings as well as its own. There is a written warning about exactly this trap, and
it has bitten this seam before. I went and counted: all 226 tickets come from one place, one shape,
and only one piece of code files them. So the answer is fine — but I didn't know that when I
submitted, and I should have.

A second note led to one extra test, guarding something genuinely easy to get wrong: making sure a
page called "/pricing" can never be mistaken for a different page called "/pricing.html" when
matching up tickets. Two notes I've accepted without fixing, and said so plainly: this solves the
problem for one kind of ticket while the underlying gap stays open for others (that belongs to
another workstream), and if a third part of the system ever needs this same trick, it should be
pulled out into one shared piece rather than copied a third time.

Still not live — that needs the next release. Nothing needs you.

**2026-08-13.** This evening's release carries the work, so it is **live** — both halves, checked
properly rather than assumed. Two things went wrong in the checking and both were my own fault in an
interesting way. The standard command for asking a service which version it is running now returns
rubbish, because our own written warnings *about* that command have been fed into the system's
prompts and get logged, so searching the logs for the phrase finds the warning instead of the
answer. And when I looked inside the running program for my own change, it said it wasn't there —
which is correct and misleading at once: the program records only the single version it was built
from, not every change inside it. Both are now written down for whoever hits them next.

**It has not actually done anything yet, and it can't until Monday the 17th.** The weekly check
visits each site once every seven days, and every site was visited on the 10th — so nothing is due.
The rota still ticks every hour and looks busy while dispatching nothing, which is exactly the sort
of thing that gets mistaken for a fault. It isn't one. I have deliberately not forced a run.

**What I did instead was write down, in advance, what should happen** — because a prediction that
can be wrong is worth more than a report written afterwards. Monday, robot-hands.com goes first. It
has 34 open tickets across 21 pages, so it cannot close more than 34. On one page in particular
there are three tickets, and I know a previous fix addressed two of them and not the third. So two
must close and one must stay open. If all three close, the thing is closing tickets too eagerly and
we stop and look — and that is a distinction you simply cannot draw from counting how many closed.

Nothing needs you. The next real moment is Monday.

---

## 2026-08-13 (evening) — from a different thread, not this lane: the thing we've been pointing components at doesn't do what we said it does

You reported the invisible links on the darts guide page yesterday. Another thread diagnosed it,
found it was a component we'd never fixed before, and handed the repair on rather than doing it. I
picked that up, and the intention was straightforward: do what we did before, but properly this time
— stop fixing these one at a time by hand and fix all of them at once. There turned out to be **168
components** with the same flaw, and we had fixed **four** in six days. You found the fifth by eye.
That ratio was the whole argument for automating it.

**I didn't do it, because I checked what we'd be pointing them at and it isn't what we thought.**

The way this fix works: a site's palette has a "primary" colour. Some components use it for text.
On a few sites that colour is nearly the same as the page background, so the text is invisible —
that's your darts page. Our repair was to invent a second variable, `primary-ink`, described as
"primary, adjusted until you can read it", and point the text at that instead. The plan said, in as
many words, that it "prefers a palette colour so the site keeps its character".

**It doesn't keep any character. `primary-ink` is just the body-text colour.** I downloaded the
stylesheet of every live site and checked: on all 16 places where that variable differs from the
colour it's meant to be adjusting, it is **exactly the site's ordinary body-text colour**. Not once
is it a lighter version of the brand colour. Never anything else at all.

It's one line of code. It tries a list of colours in order and takes the first readable one, and
body text is first in the list — and body text is the colour we *chose* to be readable on the
background, so it always wins. The other four options can never be reached.

**Why that mattered more than it sounds.** Had I gone ahead, on 14 sites I'd have quietly replaced
brand-coloured links with plain body text. On webdesign.co.uk — the site whose entire pitch is that
we do design — that means the warm tan links all go near-black. And here's the part I want you to
notice: **our contrast checker would have called that a perfect pass**, because near-black text on a
pale background is extremely readable. The tool measures readability. It has no opinion about
whether the page still looks like your brand. So the bug would have been marked fixed, the numbers
would all have been green, and the only thing that would ever have caught it is you opening a page
and thinking it looked wrong — which is exactly how this bug got found in the first place.

The four fixes we already shipped are still good, and I want to be fair about that: those elements
were **invisible**, and now they're readable. That's a real improvement and it was measured. What's
wrong is only our belief that the brand colour survived. It didn't.

**What I think we should do**, and it's cheaper than what I was about to do: fix the one line
instead of the 347 places. Make it try *darkening or lightening the brand colour itself* until it's
readable, and only fall back to body text if that can't be done. Then the four fixes we've already
shipped improve on their own, with no further edits, and the big sweep becomes safe to do. Doing the
sweep first would spend the brand on 14 sites to buy readability that the one-line fix gives us for
free.

**One honest admission.** I used a second AI to attack my own plan, which is what caught this. Its
argument was right and I checked it myself. But it also handed me a pile of supporting numbers and I
wrote several of them down as if I'd measured them. Two I later checked: one was wrong by a factor
of twenty. I've corrected them in place, marked the ones I couldn't re-check, and written the lesson
down — a number someone else measured is a quotation, not a measurement, and our notation had no way
to tell those apart.

**Nothing here needs you tonight, and I've changed nothing on any site.** Two things for when you're
next at a keyboard: the cluster login expired mid-session, so a few figures are still unverified and
one bookkeeping sync didn't run. And the decision about whether to fix the one line is worth your
view, because it changes how every site's links will look.

---

**2026-08-13, evening.** Picked the lane back up and checked on the other threads first, as asked.
Nothing here needs you, and Monday is still the date. But three things came out of the check that
are worth knowing.

First, I re-proved that the thing we shipped is still running. That sounds like paranoia, but the
fleet was released again this afternoon (v1.0.1295) after our verification this morning, and a
release rolls everything — so "it was live at lunchtime" is not the same claim as "it is live now".
It is live now: I asked the running program which commit built it, with a control to prove the
question was capable of answering "no", and ours is in there. The 226 parked tickets are untouched,
nothing has retracted, and the audit rota still has Monday 14:54 as the first site due. Everything
the last note predicted is still exactly on track.

Second — the more interesting one — another session is inside the same bug right now, working the
half of it we did not take. It found something that contradicts a claim in our own plan: the
"legible ink" colours we introduced do not, as we wrote at the time, preserve the site's brand
colour. They come out as the site's ordinary body-text colour, on every site. I checked three of
their sixteen sites myself before believing it, and they are right. That does not undo anything we
shipped — the elements we repaired were genuinely invisible before and are genuinely readable now —
but the reason we gave for why it was safe was wrong, and I have recorded that as a correction
rather than quietly moving on.

It also creates a subtle collision that I would not have spotted without looking, and it is the real
reason checking other threads was worth the time. Two of the three tickets in our Monday test are on
elements that use exactly the colour they are about to change. If their fix lands and that site gets
rebuilt before Monday, our test stops measuring the thing we built it to measure — it would be
measuring their new colour instead, and a confusing result could easily be misread as "the ticket
closing is broken". The third ticket is untouched by their work, so the half of the test that
guards against closing too much survives either way. The fix is free: either nobody rebuilds that
one site before Monday, or whoever does tells us and we re-state the prediction first. I have
written it into the handoff and messaged the other session.

Third, I now know why that third ticket fails, and it is a different kind of fault from everything
else in this bug. The button is meant to be a white pill with the brand blue as its label. But the
"brand blue" it reaches for is stored as a gradient — two blues fading into each other — and a
gradient is not a colour. A browser cannot use it as text colour, so it throws the whole instruction
away and the text falls back to inheriting white from the panel behind it. White text on a white
button. The safety net written into that line, which would have produced a perfectly good dark
colour, never runs, because it only helps when the value is *missing*, not when it is *present and
of the wrong sort*. That is a nasty one, because reading the code shows you a sensible-looking
safety net that does nothing.

The login came back before the end of the session, so I could finish this properly rather than
leaving it as a theory, and the answer is not a guess: the audit records what it actually measured
on each ticket, and for this button it recorded white text on a white background at the worst
possible score. Sixteen of the seventeen tickets of this kind across the whole estate are that exact
fault. The seventeenth is a genuinely different problem — a real colour that is simply too pale —
which is a useful thing to have, because it shows the two faults can be told apart at a glance.

I should own a mistake in the middle of this, because it nearly sent me the wrong way. My first
count said the problem hit three sites out of eight. Then a site turned up with the broken button
but, apparently, none of the cause — which looked like my explanation was simply wrong. It wasn't:
I had been reading the site's main stylesheet, and that particular site overrides the colour further
down, inside the page itself. The convenient place to look was the wrong place to look. Once I read
the value that actually wins, it is five sites in ten, not three in eight — so my easy check had been
quietly under-reporting the damage rather than over-reporting it, which is the direction you least
want to be wrong in. I have written that down as the check to use in future, because anyone
investigating this would reach for the same convenient command I did.

One more thing worth your attention, though it belongs to a different bug and I have not chased it:
on four unrelated sites, that gradient is the *identical* blue, and it is not a colour from any of
their palettes. That smells like a shared template quietly supplying a default nobody chose, which
is the same shape as the bug about generated palettes inheriting a layout's stray light colours.
I have flagged it rather than pulled the thread.

---

**2026-08-14.** A short one, but it contains the closest I have come this week to filing something
that was wrong about someone else's work.

The other session has now committed its fix to the colour derivation. It is not switched on yet, but
the way builds work here, the next time anybody releases anything, it goes out — so it stopped being
something they control and became something the calendar controls. They told me the new colour for
one of our sites, and when I measured that colour against the two backgrounds it has to sit on, one
of them failed. On the face of it their fix would have broken an element we repaired last week, and
broken it on the exact page my Monday test is built around.

That is a serious thing to say about another thread's committed code, so I checked it against the
code rather than sending it. It does not survive the check. The function they wrote refuses to
return a colour unless it clears every background it was given, so it cannot produce the value they
quoted. When I reimplemented their algorithm and ran it on the real palette, it produces a slightly
different colour, and that one clears both. **Their fix is sound and my Monday test survives it.**
The two colours are a hair apart — two of the smallest steps their search can take — and they land
on opposite sides of the pass mark, which is why the discrepancy was worth chasing rather than
shrugging at. I have asked them to pin the real value with a test before their review, because right
now neither of us can say with certainty which one their program produces.

The more useful thing came out of that. Their method deliberately makes the smallest change that
works — it stops the moment the colour is legible enough. Sensible, and it protects the brand
colour, which is the whole point. But it means every colour it produces sits *just barely* over the
line. I measured it across ten sites: seven of the twelve colours it would generate clear the bar by
less than a tenth of a point, one by two hundredths. And the catch is that it does this arithmetic
against the background the palette *declares*, while our audit measures the background the browser
actually *paints* — and those two are not always the same, because some panels are semi-transparent
and pick up whatever is behind them. When the margin is two hundredths, that difference is enough to
push it back under.

So my prediction, written down now so it can be judged later: after the next release we should
expect a fresh crop of contrast tickets on elements this fix just repaired, and they will look like
our ticket-closing machinery misbehaving. They won't be. The remedy is on their side and is small —
aim a little above the line instead of exactly at it. I have passed it on with the numbers.

Nothing here needs you, and Monday is still Monday. I have written the test so that whoever runs it
reads one colour off the page first, which tells them which of the two mechanisms they are actually
looking at — including the third case where it went wrong, so that outcome is labelled in advance
rather than argued about afterwards.

---

**2026-08-14, afternoon.** The new build is out and it carries everything: our ticket-closing work,
and the other team's colour repair in both its versions. I checked that at the running program
rather than trusting the version number, with a control to prove the check could have said no.

The useful news is that the colour repair is live but **asleep**. Nothing tells a site to rebuild
itself, so no site has actually taken the new colours yet — that is deliberate on their side, and
they intend to do one site first and wait for your say-so before widening. I confirmed it by reading
robot-hands' actual stylesheet: still the old colour.

That is the best possible position for Monday. Our test gets a clean run, measuring our own work
rather than sitting on top of somebody else's change.

One caution I have written into the handoff, because it is the sort of thing that catches people.
Until today there were two separate reasons our test was safe: the new colour code was not in the
system, *and* no site had rebuilt. Now there is only one. If anything rebuilds that site between now
and Monday afternoon, the colours move. So whoever runs the test on the day must read the colour off
the live page **on the day** — not trust today's reading. We made a version of that mistake last
week on the very same site, so it is written down twice.

I have written a fresh handoff so this can be picked up cold in a new conversation. Nothing needs
you before Monday, except that on Saturday there is a small pricing task on the other rota.

---

**2026-08-14, evening.** Both your decisions are recorded and the first one is in motion.

On the colours: AA-as-default is now written down as this lane's standard — the stricter variant I
showed you is dead unless a brief asks for it. The darts site's stylesheet rebuild has been filed
through the framework's own queue, the same mechanism we proved twice last week, with the guard
checks run first: the site's palette is pinned so the design pass cannot invent new colours, and
nothing else is currently touching its stylesheet. I did not rebuild it by hand — the whole point of
the queue is that the framework does its own work. What you should expect to see when you look: the
links and labels that today render in near-white body text will render in a muted navy instead —
readable, but recognisably the brand's colour rather than plain text. If it looks right, say so and
we widen one site at a time. If it looks washed out, say that instead — there is a stronger variant
ready and it costs four characters.

One honesty item: when I explained this decision I told you fourteen sites would take new link
colours. The right figure is four components on thirty-five pages across seventeen sites — I had
counted where the colour is *defined* rather than where it is *used*. The correction is in the notes
with the query that settles it.

On the renderer: your ruling is recorded, near-verbatim, in the bug file where the question lives —
the renderer learns about self-painted backgrounds, one agent owns the whole repair end to end (a
new agent if needed), and hand-fixes are retired. I have deliberately not started that work: it
changes what a shared mechanism promises, so by our own rules it goes through the architecture
track, and the write-up in the bug file is shaped as the brief for it. The two dozen live failures
stay open until it lands — they are the reason the work exists, and closing them by hand would
remove the evidence while leaving the defect.

Timing note so nobody is surprised: the darts site's own weekly check runs early Monday morning,
several hours after the robotics site's. With the new colours live, some of its seventeen open
tickets may close on their own in that run — that would be the new closing mechanism doing its job
on genuinely fixed elements, and I have written it down in advance so it reads as expected rather
than alarming.

---

## 2026-08-14, later — a second thread's read-out, and four rulings that arrived after the darts rebuild was already filed

This is a different session from the one that wrote the entries above. The owner asked me to explain
the outstanding decisions in more detail, then ruled on them. Recording both, plus three measurements
that change what the rulings mean in practice.

**What the owner ruled, in his own order.** One: yes, links and labels should be brand-coloured.
Two: aim for a contrast ratio of **5.0**, not the bare 4.5 minimum. Three: yes, build an off-switch.
Four: the big sweep goes ahead, but only after he has seen the darts site with his own eyes.

**The first thing to say is that ruling two arrived after the darts rebuild had already been filed,
and the two do not agree.** The filed job will produce colours at 4.60 and 4.54 — correct for the
old target, below the new one. At 5.0 the same site would get a slightly lighter navy and a slightly
lighter coral. Neither is wrong; they are answers to two different questions, and the job in the
queue is answering the older one. Nobody has done anything careless here: the job was filed on a
ruling that was accurate when it was made, and the newer ruling came through a different session an
hour later. This is simply what a fast-moving decision looks like when two people are carrying it.

**The second thing is a correction to a correction, and it matters for the check the owner is about
to make.** The entry above corrects fourteen sites down to "four components on thirty-five pages
across seventeen sites". That is right about the components and it misses a second surface: five of
the page layouts also carry the rule `a { color: ... }` pointing at the ink colour, which is *every
link in every piece of prose* on the sites that use them. I confirmed this by reading the live
stylesheets rather than the database — webdesign.co.uk serves that rule today.

The consequence is uncomfortable and worth stating plainly. **On the darts site, exactly one thing
changes: a single small uppercase label on the homepage.** Its stylesheet contains no live use of
either ink colour at all. On webdesign.co.uk, by contrast, every link in every article changes
colour. So the site chosen as the rehearsal is the site where the change is nearly invisible, and
the sites where it is dramatic are not being looked at. The staged order is still right — one site
first is the correct instinct — but the owner should know that approving the darts site approves a
great deal he will not have seen.

**The third thing is smaller and good news.** The predicted colour for the darts site's accent was
flagged by the other session as unverified. I recomputed it from scratch — a separate implementation
of the colour maths, with the input values read off the live stylesheet rather than taken from
anyone's notes — and it agrees exactly, on both colours and both ratios. That is this lane's own
rule about computing a number twice from separately-sourced inputs, applied to somebody else's
number rather than my own, and this time the number held.

**Two other sessions are working this same bug right now** — one on the derivation and the darts
rebuild, one on the machinery that closes contrast tickets. Both are active as of this afternoon.
Nothing here is a criticism of either; the point of writing it down is that three threads on one bug
is exactly when a ruling gets applied twice at different strengths.

What I have deliberately not done: I have not touched the queued rebuild, changed the contrast
target, or started the off-switch. The first belongs to the session that filed it, and the other two
are code changes in a lane I do not own.

### Later the same afternoon — the four rulings are built

Same session as the entry above. All four decisions are now implemented and committed. Nothing has
changed on any live site yet, and it cannot until the fleet is rolled.

**The queued darts rebuild is held, not cancelled.** It would have produced the 4.5 colours, which
you overruled an hour after it was filed. It is parked in a state the queue will not pick up, with a
one-line note saying how to release it. The session that filed it has confirmed it agrees and has
frozen its own edits on these files while I worked.

**One thing I found on the way in that I want you to know about, because it was nearly a silent
mistake.** The contrast number you raised was being read by *two* different mechanisms: the one that
colours links and small labels — which is what you ruled on — and a second one that colours the text
*on* buttons and filled bands, which you did not. They shared a single value in the code. Changing
that value would have quietly restyled every button label on every site as a side effect of a
decision about links. I split them into two separate numbers and left the button one alone. That is
a judgement I made about how far your ruling reaches, not something I measured, so it is written up
as the thing I most want the reviewers to argue with — and it is a two-line change if you'd rather
buttons moved too.

**The off-switch is built and defaults to ON.** That sounds backwards for a safety switch, and the
reasoning is worth a sentence: a switch that defaults to off would leave the *broken* colour
behaviour as the normal case, which is what the other session was right about when it originally
declined to build one. So instead it is an opt-out — everything on, and an operator can turn a
single named site back to the old behaviour by editing configuration, with no rebuild and no
deployment. That is the whole point of it: today, undoing this needs a code revert and a full fleet
roll; after this, it needs a config edit.

It also has a floor. The contrast target can be retuned by config, but not below the legal
accessibility minimum — a switch that can be configured to ship unreadable text through a config
edit would be worse than the bug it exists to undo.

**Why you can believe the new colours.** Every one of them was worked out twice, by two separate
pieces of code written independently, from palette values read off the live sites rather than copied
from any of our notes. The two agree exactly. As a check on the checker, the same independent code
reproduces all seven of the *old* colours too — so it is not simply agreeing with itself. I also
deliberately broke the code in five different ways and confirmed each break makes a specific test
fail, because a test that passes when you sabotage the thing it guards is not guarding anything.

**Where the register was lying.** Our own internal reference still told readers the colour mechanism
was broken and not to use it — true when written on Wednesday, false since Thursday. The automated
reviewers treat that file as fact, and one of them is reviewing this change right now. Corrected in
place, with the old wording kept visible underneath, because the wording is what misled people and
deleting it would lose the lesson.

**What happens next, and the order matters.** The fleet needs rolling — that is yours to run. Then
the darts site is released from its hold and rebuilt, along with webdesign.co.uk as the second
canary you asked for. Then you look. Only then does the wider sweep start.

One caution for when you look: on the darts site this changes exactly one small label on the
homepage. On webdesign.co.uk it changes every link in every article. That is why both are in the
check.

---

**2026-08-14, late — a correction to something I told you twice today, and it is good news.**

I said that once the colour repair was in the system, any rebuild of any of those sites, by anyone,
for any reason, would change its link colours — and that nobody controls that. The other session
went and measured it, and I have re-checked their measurement with a wider net: **it is not true.**
Only one thing in the whole fleet regenerates a site's stylesheet — the design agent, the exact
process our held rebuild request points at. The everyday page rebuilds that happen all the time
cannot touch the colours; they only link to the stylesheet, they never rewrite it.

So the exposure I described — "one unrelated rebuild away from going live" — overstated the risk by
a wide margin. What actually stands between the new colours and a visitor is: somebody deliberately
pointing the design agent at one of those sites. The one such request that exists is ours, and it is
held. Your decision is better protected than I told you it was.

The same evening gave us the proof it matters: another team was rebuilding page sections on the
darts site while we believed the broad version. Under that belief, they were one command from
spending your before-you-look gate. Under the measured truth, their work could never have touched
it — and telling them the narrow truth let them carry on instead of stopping for a hazard that
does not exist. A warning people must ignore to do their jobs is a warning they stop reading, so
the record now says the narrow, true version.

---

**2026-08-15, morning.** The overnight build changed three things, and one of them is you seeing
the new colours earlier than planned — through nobody's decision, and safely.

First: the new build carries the first half of the 5.0 work but not the second — the version the
reviewers asked to be revised, not the corrected one they approved. The difference doesn't change
any colour; it's belt-and-braces for future code. But your before-you-look check deserves to run on
the approved version, and the corrected half will be in your next release anyway. So the darts
rebuild stays held. The gating check we built did exactly its job: it looked at the running system,
said "not yet", and that was the right answer.

Second — the interesting one. Yesterday afternoon, a routine design-audit process rebuilt the
stylesheets of the robotics site and the cooking site, entirely on its own, for its own reasons. It
happened to run while the earlier 4.5 version of the colour repair was live. So those two sites are
already showing brand-tinted link colours — the robotics site's labels are now a muted navy rather
than body-text white. This is exactly the third-party possibility we wrote down two nights ago as
"rare but real"; it took less than a day to happen. It is safe — the colours are legible and
brand-hued, just at the old threshold rather than your revised one — and in one way it's useful:
if you look at robot-hands.com today, you are looking at approximately what the fleet-wide change
will feel like, a shade darker than what will actually ship at 5.0.

Third: one of the 226 parked tickets was cancelled by another team overnight — one of the
white-on-white button tickets, which belongs to a family another lane now owns. Not our closing
mechanism (that always signs its work); noted, not alarming. Monday's count expectations adjusted
by one.

Monday's test survives all of this — the two tickets that must close still must close, the one that
must stay open still must stay open — but whoever grades it reads the live page first, as the
handoff now says for the third time, having been proven right about it twice.

This conversation has run long, so the handoff file is fully up to date for a fresh start:
`HANDOFF_2026-08-14_continue_here.md`, status block dated today at the top. Also: today is the
day for pricing the discovery-rotation ramp.

---

## 2026-08-15 — the build that wasn't, and the owner's call to wait

Same session as the 08-14 entries above.

**A fresh chassis build was reported as deployed. It had not been.** The fleet is still running
last night's build, and I checked it three separate ways before saying so: the running containers
have not restarted since 20:36 last night, there is no newer image anywhere on the machine, and the
deployment reports itself fully settled on the old one.

**The likely explanation is worth knowing, because it will happen again.** The release command does
not pick a new version number for you — it reuses whatever the file currently says. Rebuilding
under a version number that already exists makes the machines keep the copy they already have. The
build runs, reports success, and nothing changes. The fix is to name the new number explicitly when
releasing.

**The good news underneath it:** last night's build already contains the change that produces the
5.0 colours. What it lacks is a small safety refinement the reviewers asked for, which guards
against a kind of future mistake rather than changing anything visible. I said so plainly and
recommended going ahead, because the colours would have been identical either way.

**The owner decided to wait for the reviewed version, and that decision is recorded as his.** It
costs a release cycle and changes nothing on screen; what it buys is that the page he signs off is
produced by exactly the code that was approved. I've written it into the handoff along with the
reasoning I used to argue the other way — so that a future session can see the argument was heard
and settled, rather than reopening it.

**Meanwhile, two sites already show the new look, by accident.** A routine design check re-rendered
robot-hands and cookly yesterday afternoon, before the roll. Their links and labels are now the
brand colour rather than plain body text. That is a real preview of the change, live today, at a
very slightly darker shade than the final. Nothing else has moved — the two sites chosen as the
formal check are untouched and still show the old near-white links.

That accident also disproved something I had written down two days ago: that nothing would trigger
a re-render, so the change would sit dormant until we chose to release it. Other teams run their
own schedules, and one of them reached these sites within a day of me claiming nobody would. The
claim has been corrected where I made it.

**15 August, later — the weekend homework: what will the site-checker cost us?** Before the slow
restart of the automatic site-checker was allowed to widen, we owed an answer on what it costs to
run. The answer is: **nothing on the AI meter, to a first approximation.** I checked it two ways.
First, the checker's code simply contains no AI calls — it is a set of database lookups (broken
links, placeholder text, unverified claims and so on). Second, it already did a complete sweep of
all 22 sites last weekend, and the AI spend log shows not a single call from it, while the same log
was happily recording thousands of calls from everything else. So the earlier estimate — up to four
million tokens for a full sweep — was borrowed from a much bigger mechanism and does not apply; the
real number is zero.

What it does cost is **to-do items**: that one sweep filed 774 of them across the fleet. Most were
worked through or judged within days. The follow-up cost therefore depends entirely on what later
consumes those items, not on the checker itself.

One surprise worth knowing: **the restart hasn't actually restarted yet.** Because last weekend's
sweep touched every site, its seven-day politeness rule means nothing has been eligible to check
since. The first site becomes eligible tomorrow morning (Sunday), and it will then work through the
fleet at one site every three hours — done in about three days. We'll glance at the meter mid-week
to confirm the zero holds in practice, but there is no cost reason to hold anything back.

**15 August, midday — the reviewed version is live, and both preview sites are ready to look at.**
The release you ran this morning carried exactly the code the reviewers approved — I checked the
running machines themselves before touching anything, because yesterday a "fresh build" claim
turned out to be false. It was real this time. So I released the held rebuild for dartsonline and
filed the second one for webdesign.co.uk, as agreed. Both ran within half an hour, first attempt.

Then I checked the actual pages rather than trusting the job reports. dartsonline now has its
links-and-labels colour at the agreed 5.0 legibility level — the small capitals label above the
homepage heading is the thing to look at there. webdesign.co.uk is the site where the change really
shows: every link inside the article text is now a warm brown that belongs to the brand palette,
instead of the plain dark grey it had collapsed to. On both sites, comparing against the saved
before-copies, precisely the intended lines changed and nothing else.

**The next move is yours: look at dartsonline.com and webdesign.co.uk.** If you like what you see,
say "Go" and we widen it to the rest of the sites one at a time, checking each as it lands. Nothing
else changes until then.
