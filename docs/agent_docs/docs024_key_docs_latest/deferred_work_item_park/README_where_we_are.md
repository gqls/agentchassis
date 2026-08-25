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
