# Where we are — bugs_closed/040, partial page builds

Plain-prose running log, **append-only, newest at the bottom.** What you'd say out loud.

---

## 2026-07-26, evening — the bug that wasn't, and the one we found instead

**Short version:** the thing the last session left as an unsolved mystery was not a bug at
all — the experiment had already worked, one minute after they gave up on it. And in
chasing it we found something real and unrelated, which is that a single stuck job can
freeze the queue that every session's work goes through, for hours, until someone
restarts the service.

**The background.** Bug 040 was about page builds that failed halfway and then lied about
it — the page got stamped "deployed" and became invisible to the thing that would have
rebuilt it. That was fixed and closed days ago. What was left was a small tidying-up
piece: when one of those builds fails, the failure reason should be written onto the work
item, so whoever looks at the queue can see *why* it failed instead of a blank box. That
was written, reviewed, approved and shipped. The only thing missing was watching it
actually happen on the live system.

**The mystery.** The previous session built a little test harness — a fake job designed to
fail on purpose — and fired it five times. Nothing happened. No trace, no error, nothing
in the logs. They ruled out six explanations one by one, wrote it up under a heading in
capital letters saying UNSOLVED, and left three more theories for whoever came next.

**What actually happened.** I checked the test's results before working through the
theories, and they were sitting right there: the test ran at 21:16, passed both of its
checks, one minute after the handoff was written saying it never would. The messages had
been queued behind a jam. And the really galling part is that the same handoff document
*describes that jam* two sections further down, under "found on the way" — including a
printout showing the test messages sitting in the queue behind it. The evidence was in the
file, filed under a different story.

The lesson is a cheap one and I've written it into the fleet-wide log of wrong calls:
**when something looks like it vanished, check whether it arrived late before working out
why it never came.** A jammed queue and a lost message look exactly the same from the
sending end.

**So where does the fix stand?** Two of the three checks are now proven on the live system:
the failure reason gets written when it should, and — the one that actually mattered — it
does *not* get written onto jobs that succeeded. That second one sounds trivial and isn't:
the error is remembered for the rest of the job, so without a deliberate guard a job that
stumbled and then recovered would be recorded as a failure. It isn't. Good.

The third check needs one more run, which is queued now. It asks whether a job that sets
its *own* reason still keeps it, rather than being overwritten by the new automatic one.
I couldn't answer it from history, because every example on the system was written before
the change went live — they prove a message was written when there was nothing to compete
with it, which isn't the question.

**I also got something wrong myself, and caught it.** Looking at the test result I said the
"name the step that failed" part of the code had run. It hadn't — the message already had
the step name in it, so that part correctly did nothing. The output looks identical either
way, which is exactly why I should have checked the input instead of reading the output.
Corrected in all three places it was written.

## 2026-07-26, later — the queue jam is real, and it is worth someone's attention

While waiting for the test to run I worked out what the jam actually is, because it was
happening again in front of me.

**The dispatch queue is a single line, one job at a time, in order** — and the service does
not mark a message as done until the job it started has finished. That is deliberate and
it is the right design: it was changed to work that way so that a crash could not silently
swallow work in flight. But it has a cost nobody had written down. **The queue moves at the
speed of whatever is at the front of it** — and one of the things that regularly sits at the
front is a sixteen-reviewer council run, which takes about ten minutes. Everyone else's
work waits.

That is merely slow. The bad case is when the job at the front doesn't just take a long
time but **never finishes at all**. Then the queue stops dead, and nothing gets through
until somebody restarts the service.

I watched both happen:

- The healthy case: I sampled the queue position and the running job together every thirty
  seconds. The queue sat still for the entire ten minutes of a council run, then moved
  within half a minute of that council finishing. That's the link, demonstrated rather
  than guessed.
- The bad case: the jam the previous session hit was a council run that started at 18:44,
  ran for six and a half minutes, and then **stopped mid-way and never moved again** — it
  is still sitting there in that state now. The queue stayed frozen behind it from 18:51
  until 21:02. The previous session recorded that it "cleared on its own" at around 21:04.
  It didn't clear on its own. **A service restart at 21:02 cleared it**, and only because
  someone happened to be deploying something else.

Two and a half hours, six other sessions' review submissions stuck, two real site updates
stuck, and none of those people could see why.

**How often?** Not often. Five stuck jobs against about 1,566 that finished normally in
twenty-four hours. But the damage isn't proportional to how often it happens, because a
single one takes out everybody until a restart. Two of the five stopped within seconds of a
service restart, so at least some are jobs killed mid-flight by a deploy.

**What I have not done, deliberately.** I haven't tried to fix it. That queue belongs to
another active workstream — they closed a related case on it this morning and have been
working it all day — so I've written the evidence into their bug file rather than starting
a competing effort. **Why those jobs get stuck in the first place is still unknown**, and
that probably belongs with the separate lost-work investigation.

**The thing worth deciding, when someone has a moment:** there is currently nothing that
notices a job has stopped making progress. A watchdog that gave up on a job idle for, say,
fifteen minutes would turn "the whole system is mysteriously frozen for two hours" into "one
submission failed and can be retried". That's a design call for the owning workstream, not
something to bolt on from here.

## 2026-07-26, 22:05 — done, and the last check was the interesting one

The final run went through and all three checks passed. Bug 040 is now finished in full,
with nothing left owed.

The third check was the one I'd expected to be a formality and it turned out to be the one
worth doing. The question was: when a job sets its *own* failure reason, does the new
automatic reason overwrite it? To test that properly you need a job that has *already*
failed once — so the automatic reason is sitting there, loaded and ready — and then a step
that supplies its own wording. If the guard were wrong, the specific human-written reason
would be silently replaced by a generic one, and you would never notice, because both look
like perfectly good text in the box. It held: the job kept its own words.

I couldn't have answered that from the existing data, and it's worth saying why, because
it's a trap that will come up again. There *are* examples on the system of jobs carrying
their own reason — but every one of them was written before the change went live. They
prove a reason was written when there was nothing competing with it. That's a different
question, and answering the easy version would have looked exactly like answering the hard
one.

Same shape as a second thing I hit: the written-down way to check this fix was "watch the
count of blank failure reasons stop going up". That can't work. The count went *down*
between yesterday and today — from 21 to 14 — not because anything was fixed but because
old records age out, so the pool it's measured against keeps shrinking. And separately,
not a single real job has failed since the fix went live, so the count is quiet because
nothing has happened, not because something is working. Both of those would have read as
success. Deliberately breaking something and watching it, which is what we did, is the only
version of this that answers the question.

I've cleaned up the test fixtures I created on the live system — nothing left behind.
