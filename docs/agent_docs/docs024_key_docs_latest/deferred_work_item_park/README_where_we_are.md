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
