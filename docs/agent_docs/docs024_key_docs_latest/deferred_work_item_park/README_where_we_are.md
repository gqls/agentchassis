# Where we are — work parked in the queue that nothing will ever pick up

The owner's running log. Append only; newest at the bottom.

## 2026-08-25 — what this is, and why it is smaller than I first said

While finishing the dead-links bug this morning I tried to ask the system to rebuild a
mortgagecalculator guide page, and it refused: there was already a request for that page in the
queue. It had been sitting there since the 3rd of August — twenty-two days — and had never been
attempted once. It was parked in a state that nothing in the platform ever looks at, so it was
never going to run; and because it was still occupying the queue slot for that page, it also
prevented anyone from asking again. The page was, quietly, unrequestable. I woke the existing
request up and it rebuilt in two minutes.

I reported that there were **297** such jobs. **That number was too big, and the honest figure is
118.** Three different things were sitting under the same label and I had counted them as one.

About 75 of them are **entirely correct**. The platform deliberately uses this parked state to mean
"here is a piece of work we can see and have no way of doing yet" — a roadmap note rather than a
job. Those are written that way on purpose, with comments in the code explaining it, and there are
two places in the system that read them. Nothing is wrong there.

Another 87 were parked on 11 August by a deliberate, recorded database change, which stamped every
row with who parked it, why, and what would un-park it. Fourteen days later you can still read all
of that off the row. Those belong to an existing bug that another thread is actively working, so I
have left them alone.

That leaves **118**. These name a real handler — so they look like ordinary queued work to anyone
reading the queue — and they carry **no record whatsoever** of who parked them or why. And here is
the part I cannot yet explain: **there is no code in the platform that can produce them.** Every
place in the system that writes this parked state deliberately leaves the handler blank, and no
code anywhere moves an existing job into this state. So something outside the normal machinery did
it, roughly ten times, mostly one site at a time, and left no trace.

I have deliberately **not** guessed. I have handed the question to the platform's own diagnosis
loop, which reads the real code and the live database and comes back with a cited answer — that is
the house rule for exactly this kind of claim, and it is the right one here, because the obvious
guess (that earlier sessions parked them by hand) is comfortable and I have no evidence for it.

I should also say I got one thing wrong along the way. I ran a test to work out whether these jobs
were *created* parked or *moved* into it later, and got a clean-looking answer — and the test
cannot actually tell those two apart, because every write to a job updates the same timestamp I was
measuring. I caught it, but only because a second thing disagreed, not by re-reading my own work.
That is twice today.

Nothing here is urgent and nothing is broken for a customer right now. The cost is that work we
decided to do can vanish silently, and the page it belonged to cannot be asked again.

## 2026-08-25 (later) — the machine could not find the culprit either, and that is worth knowing

I handed the question to the platform's own diagnosis loop rather than guessing. It ran four
rounds, read the real code and the live database, and came back **"not confirmed"** — its own words
were *"zero remaining named candidates in the read code… hand to a human; do not auto-conclude."*

That is a genuine result, not a wasted run, and it earned its cost twice.

**It caught something I had missed.** I had been separating the honest parked jobs from the
suspicious ones by looking for the two "who parked this and why" stamps I knew the platform used.
There was a **third** one I had never heard of, and four jobs carry it — an owner-sanctioned park
from 12 August, fully explained. Four out of a hundred-odd, so the headline barely moves (118 down
to 114). But the mistake underneath is the useful bit: **I tested for the labels I already knew
instead of asking what labels exist.** That is one query, and it is now written down. It caught
something in the other direction within the minute, too: another field looked like a "why it was
parked" note on 22 jobs and is actually a "why it was found" note, so trusting it would have made
the problem look smaller than it is.

**And it got one thing wrong, which I checked rather than took on trust.** It ended by asking for
the code of a particular function, calling it the only place that could do what we are looking for.
That function **does not exist** anywhere in our codebase. Its own report flagged that it had only
seen the name and not the body — which was the clue. Had I taken the verdict at face value I would
have spent another round chasing something that isn't there. Worth saying plainly: the loop is very
good and it is not an oracle, in either direction.

**Where it leaves us.** I have filed it as **bug 396**, with the root cause openly recorded as *not
established* rather than filling the gap with the comfortable answer. What the file does carry is
eight specific explanations we can now rule out, each with the evidence, so nobody has to walk that
ground again. Two possibilities remain and both are labelled as unproven: that a routine site scan
parks them as a side effect (we know one was finishing at exactly the right minute on the right
site — but something happening at the same time is not the same as something doing it), and that
earlier sessions parked them by hand (for which I have no evidence at all beyond having run out of
alternatives, which is not evidence).

**One thing did get fixed on the way**, and it may matter more than the 114. There is a setting a
future session could put on any agent that would create this exact problem deliberately and at
speed — it writes whatever word you give it straight into the job's status, with nothing checking
that the word is one the system understands. "Deferred" is the natural word to choose and it is
precisely the wrong one. All four places currently using it happen to have chosen a safe value, by
habit rather than by any rule. That trap is now written into the landmines file, where a session
about to touch it will see it first.

**Nothing here is urgent and nothing is broken for a customer.** The cost is that work we decided
to do can vanish without a trace, and the page it belonged to cannot be asked for again.

## 2026-08-25 (later still) — a second opinion took the bug apart, and it was right on both counts

You asked for a different model to reconsider this. It refuted my filing within the hour and
answered the question that had beaten both me and the platform's own diagnosis loop. I checked
every point myself before accepting any of it; all of it held.

**First, the number was wrong.** I said 114 jobs were parked with no record of who parked them or
why. **Sixty-two of them are fully recorded** — the note is just stored in a different field of the
row than the one I looked in. One lane stamped its sixty with its own name, the reason (an
owner-ordered rebuild), and even the condition for releasing them again. The genuinely unrecorded
number is **52**.

What stings is that I had written the rule for avoiding exactly this a couple of hours earlier — *do
not check for the labels you already know about, ask what labels exist* — put it in the runbook, in
the debugging guide and in my own notes, **and then ran that check on one field and not the other**.
I applied the lesson to the example that taught it instead of to the kind of mistake it was.

**Second, and more useful: we now know what parked them.** Sessions did, by hand, deliberately —
holding a site's job queue still while they rebuilt that site. Three of the four were acting on your
instructions. The evidence was sitting in the repository the whole time: one lane's own handoff
document contains the exact database command it ran on a fifteen-second loop, and that command sets
the status and nothing else, **which is precisely why its thirty-eight jobs have no note attached**.
Another lane's commit message says plainly "14 items held".

I had searched the *code* for something that could have done it, a dozen times over, and never once
searched the *documents* — on a system whose defining characteristic is that many sessions work it
by hand and write down what they did. "No code does this" and "nobody did this" are not the same
statement, and I treated them as one.

**The best part is the last part.** The lane that parked thirty-eight of them had itself written, in
a later session, *"unverified what deferred them — a hand-park is the obvious guess and I did not
establish it"* — in the same folder as the instructions it had followed. It had forgotten its own
action. I read that note, took it at face value, and re-derived the entire mystery from scratch.

**This changes the fix, and improves it.** The problem is not a broken piece of machinery. It is a
**missing one**: four separate lanes each needed to pause a site's work queue, none had a supported
way to do it, so each invented the same hand-written database command — and only the ones who
thought of it left a note. The system has now grown **six** different homemade ways of recording the
same act. That is a far better argument for building the feature properly than fifty-two mystery
rows ever were.

**One warning I have put everywhere it belongs:** sixty of those parked jobs are being held on
purpose right now, with a stated condition for letting them go. Nobody should release them in bulk.
Ask the lane that is holding them.

Bug 396 is corrected in place — old claims struck through rather than edited away, so the record
shows what we believed and when.

## 2026-08-25 (evening) — the review said my fix was the wrong fix, and it was right

You asked for the park verb first, then the config guard. Both got built. Then the review panel
looked at the park verb and said, in effect: *you have built a tidier version of the wrong thing.*

**A way to hold a site's work queue already existed.** Three of our fifty-one sites are held by it
right now — including one you halted yourself on 18 August, with your reason still readable on the
row. It works by telling the dispatcher to skip the site. **It never touches the individual jobs**,
which is exactly why it causes none of the damage this bug is about: nothing gets stranded, nothing
blocks its own replacement, and the page that sat unbuildable for twenty-two days could not have
happened under it.

I wrote "there is no way to do this" into a formal submission without checking. **That is the third
time in one day I have claimed something did not exist without looking**, and this one got as far
as an approved plan and a live database change before a reviewer stopped it.

**What was genuinely missing turned out to be much smaller**: the existing hold is all-or-nothing.
You cannot say "hold this site *except* these three jobs" — which is precisely what the
mortgagecalculator thread needed, and why it wrote the improvised command that started all this.

**So that is what I built instead**, and it is better than the verb in the way that matters: it
still never touches a job. The site keeps its hold; a named exception list lets specific work
through. The column is live and does nothing yet; the code is committed and takes effect at the
next release; and the switch that connects them is **deliberately held back**, because turning it
on before the code ships would convert a full hold into no hold at all on exactly the sites someone
has deliberately locked. That is written at the top of the file in capitals.

**The second review round found something too, and it was subtler.** A reviewer thought I had
copied one piece of SQL into a place it could not work. I had not — but they were right that
*somebody eventually will*, because the same rule genuinely is written twice in two different
shapes, for a real reason. So I have written down why they cannot be merged, and added a test that
fails if anyone tries. Their concern is now a red build rather than an outage.

**Worth saying plainly, because it is the pattern of the day:** four separate times today I asserted
something did not exist and was wrong, and every single one was already written down somewhere in
our own notes. Not bad reasoning — not looking. I have logged all four, and the cheap check is now
the first line of this lane's runbook: before writing "nothing does X", grep the notes, then the
documents, then the code.

**Nothing is broken and nothing needs a decision from you.** The remaining step is the next release,
after which the held switch can go on and be tested on one real site.

## 2026-08-26, afternoon — one thing proved itself for free, and one safety check turned out to be looking the other way

Picked this back up expecting to have nothing to do. The handoff said, correctly, that nothing was
blocked and nothing was owed. Two things came out of simply re-checking what we had already written
down, and the second one matters more than the first.

**The good news first, and it cost nothing.** We built a way to hold a site — to stop the system
automatically doing work on it — and this morning we could only prove it worked in a laboratory
sense: we ran the rules by hand against real data, got the right answers, and undid it all. What we
could not show was the real scheduler actually obeying the hold out in production. The scheduler
always picks the site with the oldest waiting job, there were about 1,400 jobs waiting, so we could
not push a held site to the front of that queue to watch it get skipped. I wrote that down as an
honest gap.

It closed on its own. This afternoon the eight oldest waiting jobs in the whole fleet happen to sit
on `adversecreditmortgage.co.uk` — a site you personally halted on 18 August pending a decision. So
the held site is now at the very front of the queue, and the hold is being tested every single time
the scheduler runs. I asked the scheduler's real query what it would pick: it answered
`agritec.uk`. Then I asked the identical query with only the hold rule removed: it answered
`adversecreditmortgage.co.uk`. The two questions differ by one clause, so that clause is what is
keeping your halted site alone. **67 waiting jobs across 3 held sites are being held by it right
now.** That is no longer a feature we believe in; it is one we have watched work.

**The less good news.** When the reviewers approved this work, their one substantive comment was
that our automated test protects the rule only inside our program code — and this particular rule
also exists as a piece of database configuration, which a future colleague could edit directly with
nothing checking their work. I answered that by pointing at our written warnings file, which has an
entry about exactly this rule, and said in effect: anyone editing this must read that first.

I went back and actually ran the check that entry tells people to run. **It cannot tell a correct
rule from a broken one.** It looks for a word appearing anywhere in the configuration. I gave it
four versions of the rule — the correct one, and three broken ones — and it declared all four fine.
One of those broken versions switches the hold off on every site in the fleet, which would release
your halted site on the next run. Another, caused by nothing more than a missing pair of brackets,
would widen the queue from about 1,100 jobs to over 15,000 and start re-running work that had
already finished or already failed.

The uncomfortable part is that **the check was not wrong when it was written.** Back in August the
problem was that the rule was missing entirely, and looking for a word is a perfectly good way to
spot something missing. Our own change in the last two days is what made the rule conditional — and
from that moment on, "the word is present" stopped meaning "the rule is correct". We inherited a
safety check straight across the change that broke it, and nobody noticed because it kept saying
everything was fine.

I have replaced it with a check that watches what the rule actually *does* rather than how it is
spelled, and I have been explicit in writing that the new check still cannot catch the
missing-brackets version — that one is only visible by reading it. Better to say so than to leave
someone trusting a check that has a hole in it, which is the whole lesson here.

**Nothing is blocked.** The one long-standing gap is unchanged: nothing stops someone editing the
database by hand to park a job, and short of a database-level trigger nothing can.
