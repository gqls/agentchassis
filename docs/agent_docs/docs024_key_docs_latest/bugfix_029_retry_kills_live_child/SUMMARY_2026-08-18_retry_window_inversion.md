# SUMMARY — 2026-08-18 — bug 029: one defect fixed, one diagnosis withdrawn, the real hang still open

## What we're trying to do

Bug 029 is the oldest open entry in the estate's "builds mysteriously stop" family, filed
2026-07-19. The symptom is that site builds quietly stop happening and then quietly resume,
with nothing failing and nothing alerting. The job of this lane is to find out what actually
causes it and fix it at the framework level rather than for one pipeline.

## Where we've come from

The bug has already been through two explanations, and **both were wrong**. Its own title —
that hung jobs saturate a concurrency pool — was refuted inside the file on 2026-07-21. The
replacement explanation, that this is a downstream effect of a separate spawn-loss bug, is
right about the *damage* but never named what makes a job hang today.

So the lane inherited a file where the most confident-sounding sentences were the ones that
had already failed, and where an earlier wrong cause had been copied into a second bug file
as fact. That history set the standard of proof for this round.

## What we've done

**We found and fixed a real, measured defect.** When one agent waits on another and the wait
times out, it tries again. Each step declares how long it should be allowed — the build
dispatcher says fifteen minutes. That declaration was honoured on the first attempt **only**.
Every retry was silently given five minutes, and anything declaring more than thirty minutes
was given **three**. The longer a step asked for, the less it got. Thirty-three steps across
twenty-five agents were affected, including a step that waits for a human being and asks for
twenty-four hours.

That is fixed, tested (every test proven able to fail by mutation), reviewed and approved by
the council at the third round, and registered as a reusable mechanism. It is Go, so it does
nothing until the fleet next rolls.

**We also withdrew our own headline finding, the same day we made it.** We had claimed the
retry was killing work that was still running — a striking claim with a striking measurement
behind it. It was wrong: the job was already dead when the retry arrived, and the code we had
already read said so. A design pass we had explicitly briefed to disagree with us caught it,
and a check settled it — one that turned on the single odd data point we had reported rather
than tidied away.

**And the bug diagnosed itself, expensively.** The diagnosis run we filed on 029 was killed
by 029: the child finished its work successfully at 12:56:58 and the parent had given up at
12:54:27, two and a half minutes early. Forty-two minutes of analysis completed and discarded.

## Where we are now

One contributing defect is fixed and approved. **The bug is not closed and we are not close
to closing it.** What actually kills a job after it starts its work — the original "hung
spawn" the file is named for — remains unexplained. We have a candidate and we have marked it
as unproven.

The lane's account is now honest in a way the file's history has not always been: three
claims are recorded as withdrawn or unverified, the runtime evidence we could not obtain is
recorded as unobtained rather than as a clean result, and the two review passes that caught
our errors are written up as having caught them.

## Where we're going

Next is the actual hang. The narrow question — what kills the continuation after the first
handshake — is stated precisely enough to be filed on its own, and deliberately separated
from the mechanism we withdrew so that nobody investigates the wrong thing.

Beyond that there are three known, separable pieces of work, all scoped and none started: a
guard so a retry is never sent at a job that already exists; a shorter path to noticing a
frozen job than the current four hours; and a fresh question the approval round surfaced —
whether the *first* wait, not just the retries, is also failing to read its declared timeout.

The one thing we would not do is treat the fixed defect as the answer. It was real, it was
worth fixing, and it is not why builds stop.
