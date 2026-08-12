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
