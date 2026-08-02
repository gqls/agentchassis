# Where we are — the three unguarded deletes (bugs_open/165)

Plain-prose log, append-only, newest at the bottom.

---

**2026-07-31, evening.**

Background, in one paragraph. A few of our build steps work by deleting
everything they previously wrote and then writing back what they just produced.
That is a sensible way to keep things tidy — it stops stale rows piling up
forever. It goes badly wrong in exactly one situation: when the thing producing
the new content only managed to produce *some* of it. Nothing errors, because a
short answer is still an answer. So the delete removes everything the short run
did not replace, and what you are left with is not a broken page that shouts at
you — it is a page that quietly has less on it than it used to, which looks
exactly like a page that never had more.

We fixed one of these last week (the code index). The fix deliberately built the
*rule* as a shared, reusable thing, and deliberately did not go around converting
the other three places that have the same problem. When that fix went through our
review council, one of the reviewers objected — reasonably — that leaving the
other three is itself the mistake we make most often: one place gets the careful
fix, its siblings stay exposed. So the previous session wrote the objection up as
its own piece of work and left it for someone to pick up. That is this.

There are three left. Today I did the one that matters most: the table that holds
the actual sections of every page on every site. It is the one that has genuinely
lost customer content before — twice that we know of, including an interactive
game that a routine content rebuild simply deleted, on two different sites months
apart.

The interesting part of today was not writing the guard. It was choosing what
"enough" means, because a guard that fires when it shouldn't is worse than no
guard at all — the first person it blocks unnecessarily will delete it, and then
we have neither. So I measured before deciding, and two of my first instincts
turned out to be wrong.

The first wrong instinct was written into the task itself: check each *slot* on
the page separately. It turns out almost every slot on every page holds exactly
one section, so checking per slot means "if you removed one section, refuse" —
which would have blocked 89 perfectly ordinary edits over the last four months.

The second one was mine, and it is the one worth remembering. I had the guard
compare what a rebuild produces against what the page *plans* to contain — a good
second opinion, because the plan is written by different code entirely. Running it
across every page, three tripped. One of them was the idea.uk homepage: plan says
six sections, and my query said it had two. I wrote that down as the guard finding
real damage.

It was not damage. The page has all six sections. Four of them are **locked** —
someone deliberately marked them as not-to-be-touched by automation. My count was
only counting the ones a rebuild is allowed to change, which is correct, but I
read "2" as "this page only has 2 sections". The consequence was in the guard, not
just in my notes: it meant a *perfect* rebuild of that page would have been
refused, because it compared "2 sections written" against "6 sections planned"
when four of those six were never on the table. In other words, my guard would
have blocked every future rebuild of exactly the pages someone had cared enough
about to protect. Which is precisely the way these things get ripped out.

The fix was small once seen — don't count the locked ones in the target. After it,
the check trips on none of the 238 pages it can actually reach. What caught it was
opening the six rows instead of trusting the one number, and I have written that
up because the number was not wrong, it just answered a slightly different
question from the one I was asking it.

Where that leaves things. The guard is written, tested (seventeen tests, and I
broke it four different ways to confirm the tests actually notice), committed, and
submitted to the review council. It is **not yet proven working in production** —
our code only takes effect after the next deployment, and more importantly a
successful run proves nothing here, because the guard is designed to do absolutely
nothing when everything is healthy. The only real proof is to deliberately break a
rebuild, watch it refuse, confirm nothing was deleted, then undo the breakage and
watch a normal rebuild go through. That is the next job, and it needs the roll
first.

The other two sites are still open. One of them I deliberately did not touch,
because another session had that file open at the time and landing a change
underneath them on a shared tree causes exactly the sort of mess this project has
already had enough of.

One thing I needed a decision on and got it: the machine's temporary disk was
completely full — 13GB of leftover scratch from 125 past sessions — and commands
were failing mid-run. I cleared the ~4GB belonging to sessions that had been
finished for over six hours, leaving anything recent alone.

---

**2026-08-01, morning.**

The guard is finished, it is running in production, and we have watched it do
both of the things it is supposed to do. That is the headline; the rest of this
is what it cost and what it turned up.

Proving it needed a bit of theatre. The whole point of this guard is that it does
nothing at all when everything is healthy — so a successful run tells you
absolutely nothing about whether it works. The only real test is to deliberately
break something and watch it refuse. So I took a page on one of our sites, told
the system the page was supposed to have twenty sections when it really has
seven, and asked it to rebuild. It refused, said "35% — 7 of 20", wrote a note for
a human to look at, and — the important bit — left all seven of the real sections
completely untouched. Then I put the page back to normal, asked again, and it
rebuilt happily. Both halves confirmed, and the page is exactly as it was.

The interesting result was *which* half of the guard fired. I built it with two
independent checks. The obvious one — "are you writing back roughly as much as is
already there?" — looked at that page and saw seven out of seven and was perfectly
happy. It was the second check, the one comparing against what the page is
*supposed* to contain, that caught it. If I had built only the obvious check, that
run would have sailed straight through. It is rare to get that clean a
demonstration that the belt-and-braces bit was the bit doing the work.

The review council approved it first time, with seven advisory notes. Four of
them, independently, spotted the same genuine bug in my code: the note it writes
for a human would have gone quiet after two occurrences, on precisely the page
that keeps failing and most needs looking at. They were right, I fixed it, and the
live test then confirmed the fix.

Meanwhile another session picked up the two remaining places with the same
problem, and did something better than copying my work — they generalised the
useful half of it into a shared piece that all four places now use. I threw away my
private copy in favour of theirs. One nice side effect: the test I wrote to pin
that four-seat bug now protects all three sites instead of just mine.

Two things I found and did not fix, both deliberate.

The first is a small lie in the refusal message. It ends by telling the operator
that the leftovers will be cleaned up by a later run. That is true for the original
case this machinery was built for, and false for three of the four places now using
it, where the whole operation is refused and nothing gets cleaned up later — an
operator could easily read it as "this sorts itself out". It does not. The sentence
lives in the shared piece, so changing it touches all four, and another session had
that file open. Written up rather than grabbed.

The second is more interesting. Chasing down whether a refusal actually stops the
work or gets quietly swallowed, I read how the system handles a failing step and
found there is a setting that would make it skip the failed page and carry on
reporting success — which would be the exact silent failure this whole project
exists to stop. It is switched off everywhere it matters, so we are fine. But the
same fact means a refusal on one page currently aborts an entire multi-page build.
That is a real disproportion, another session had already filed it as its own
problem an hour earlier from their end, and I have added the measurements showing
it affects four places rather than the one they found.

Last thing, and it is a small point about how I work rather than about the code. I
had planned to prove those last three cases by running three more live tests.
Reading the code instead told me the general rule in one pass, covered all of them
plus any future ones, and cost a fraction as much. Three passing tests would have
proved three things; the rule proves the class. Worth remembering the next time I
write "induce it" in a plan.

---

**2026-08-02, late morning — site B is now proven too, and site C turns out to be
unprovable for a reason worth understanding.**

Quick recap of the shape of this job, because it matters for what follows. Four
places in the system rebuild something by deleting what was there and writing
what they just produced. That is fine when the thing producing the new version
saw everything; it is silent data loss when it only saw half. We are adding a
check to each one: prove you saw enough before you are allowed to delete.

Site A (page content) was done and proven last week. Today site B (site
navigation) got the same treatment.

**What I did.** The fleet had rolled, so the code for B and C was actually
running in production for the first time. I checked that directly against the
running containers rather than trusting the version number — a roll is not
evidence that your change is in it.

Then I deliberately broke a site to watch the guard catch it. I picked oufe.com
because nothing else was touching it, added sixteen obviously-fake navigation
entries so that a rebuild would look like it was throwing most of the nav away,
and asked the system to rebuild the navigation. It refused, said exactly why
("nav items 33% of 24"), left every real navigation entry untouched, and filed a
note for a human. Then I removed the fakes and checked, across the whole fleet
rather than just that one site, that none were left behind.

**The bit I want to flag, because it is a small win.** My plan then said: run a
normal rebuild and watch it succeed, to prove the guard doesn't block ordinary
work. That would have been expensive — a full nav refresh triggers a re-render of
every page on the site, which costs real money across the fleet. Before spending
it, I asked whether the system had already done that on its own. It had: a
routine build on a different site the previous morning had passed the same check
and recorded its numbers. A real build passing is better evidence than a staged
one, and it cost one query instead of a site rebuild.

That is the second time in two days that looking for evidence the system already
produced beat manufacturing it. I have written the pattern down.

**Site C is a different story, and the answer is "you cannot do this yet".** Site
C guards the internal link registry. That table is completely empty — no rows,
ever, on any site — and the only thing that writes to it has never run. So the
guard cannot be made to refuse (with nothing stored, there is nothing to protect)
and it cannot be made to pass (nothing runs it). It is not that I ran out of
time; there is genuinely nothing to point at. The code is live and was tested
offline by deliberately breaking it and confirming the tests caught it. Why that
table is empty is a separate open case someone else owns.

**One thing that got better as a side effect.** A while back I flagged that the
refusal message tells the operator something untrue for three of the four sites —
it says the leftover rows will be tidied up by a later run, which is only true
for the original site. Until today that was an argument. Now it is a real message
that a real operator could be shown, produced by a real refusal. Same fix, much
better evidence for it.

**Where that leaves things.** A and B are done, live, and proven. C is live but
blocked. I am not closing the case: the reviewers specifically asked that it not
be closed the moment the important site was finished, and closing it now needs a
judgement call from you — whether "live, tested offline, and provably harmless
because it cannot currently do anything" is good enough for C, or whether it
waits for the other case to unblock it. That is a decision rather than a
measurement, which is why I am leaving it with you.

---

**2026-08-02, evening — finished, and the last thing shipped with a proper proof.**

The refusal message fix went live on the new build. This is the one where the
system used to tell an operator "don't worry, a later run will tidy up the rows we
left behind" — true for one of the four places it was printed, and false for the
other three, where the whole operation had been refused and nothing would be tidied
up later. Each of the four now says what it actually did.

Worth a line on how it was checked, because it is the first time in this job I
could do it properly. Normally I confirm a change shipped by looking for new text
in the running program. That only proves something arrived; it cannot prove the old
thing left. This time the old sentence and the new one were the same piece of text
in the program, so I could check that the old version reads **zero** — it is gone,
not merely accompanied. Both copies of the service, positive and negative.

**Where the whole job stands: done.** All four of the risky delete-and-rebuild
steps now refuse to delete unless they can show they saw enough of what they were
rebuilding. Three of the four have been caught doing it for real in production —
deliberately broken, watched to refuse, confirmed nothing was lost, then put back
and watched to work normally. The fourth is guarded but cannot be tested, because
the thing it protects has never had any data in it and the agent that would write
it is the one we retired today.

**Two things I want to leave with you rather than bury.**

The first is that I got something wrong and it is worth knowing why. I wrote, in
four different documents, that a certain agent had not run "in the last twenty
days". It was actually the last **twenty-four hours** — the table I read that from
throws away completed records after about a day, and the way I measured its age was
fooled by a handful of odd leftover rows. The conclusion turned out to be right
anyway, for a completely different reason I found later, but the reason I published
was wrong. The uncomfortable part: I had added that "twenty days" caveat *to be
careful*, and it was the caveat itself that was false — which made the claim look
more checked than it was.

The second is the bigger question I have not touched. Retiring that agent showed
that four of the five remaining site builders have also never been used, and that
the whole intake path they belong to shows no recent activity — while sites are
still being created daily. So something else is building sites now. I have
deliberately not retired the others: the honest reading is that an entire
subsystem may have been superseded, and that is worth understanding properly rather
than removing four agents one at a time because a query says nobody picked them.

Handoff for whoever picks this up: `HANDOFF_2026-08-02_continue_here.md`.

---

**2026-08-02, evening — the B+C side is finished.**

The chassis rolled twice today. After the first roll I checked the two guards I
wrote were actually in the running binary rather than just in git, and they were.
After the second roll the last outstanding piece — the fix to the refusal message,
which had been telling operators to wait for a tidy-up that never comes — went live
too. All four of the destructive deletes this whole exercise was about are now
guarded, live, and three of them have been proven by deliberately breaking them and
watching the guard refuse.

The interesting part of today was not the guards, though. It was that three separate
times I wrote a sentence pointing at something whose status had already changed
underneath me — a bug I said was open had closed two days earlier, a decision I
called "an owner call" had been made seven hours before I wrote it, and the bug I
was working on closed while I was typing about it. None of these were wrong
measurements; every number I checked was right. They were all the same mistake:
citing a status instead of checking one. I fixed nine of them in the morning, wrote
up why it happens, and then did it twice more in the afternoon. That is worth
knowing about how this goes wrong — it is not a knowledge problem, and being told
about it does not fix it. What fixes it is one query, run at the moment you write
the claim rather than before or after.

Two things are left and neither belongs to this lane: bug 173 (a missing degree of
freedom in how loops handle errors) has no owner, and there are roughly a hundred
citations across the codebase pointing at bug files that have since moved. The
second one should be a check that runs when you edit a file, not a hundred-file
sweep — the numbering makes a mechanical rewrite genuinely unsafe.
