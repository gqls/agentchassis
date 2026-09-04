# Where we are — bug 233, our own keys written into the logs

Plain-prose log for the owner. Append only, newest at the bottom.

---

**2026-09-04 — the leak is fixed and proven fixed. One thing is left, and it is yours, not mine.**

Some background in plain terms, because the fix and the problem are not the same thing.

For about ten months, two of our own programs printed live passwords into their ordinary logs. One
printed the pair of keys that give access to our file storage, every time any service connected to
it — which is most of them. The other printed the main database password every time we started up a
new worker. Not encrypted, not masked: the actual values, sitting in the log where anyone who can
read our logs could see them.

**The printing was stopped on 9 August and I have now confirmed it is genuinely gone**, on the build
that went out today. I checked it the careful way rather than trusting the deploy: I asked the running
program for a phrase the fix *added*, a phrase it *removed*, and a third phrase that should be there
either way. That third one matters — without it, a broken check answers "not found" to everything and
I would have reported good news that wasn't true. Then I checked every service in the fleet is on the
same, current build. All 36 are.

**I also closed the one gap the original fix was honest about leaving open.** It had checked for
passwords printed directly, but not for passwords that might leak by printing out a whole settings
object. I have now swept that, and it is clean — with one thing worth telling you because it nearly
went the other way. A first pass flagged **81** of our agents as apparently holding a secret in their
configuration. That would have been alarming. Looking at the *names* rather than the count, every
single one was a setting that holds the *name* of a password (like "the key lives in
ANTHROPIC_API_KEY"), not a password. Eighty-one false alarms, nought real. Counting would have told
me the opposite of the truth.

**What is left is a decision only you can make: those credentials should be changed.**

Stopping the printing does not un-print ten months of it. The storage keys and the database password
have been sitting in readable logs since late October, so the safe assumption is that anyone who
could read our logs in that time could have taken them. The original write-up deliberately left this
to you and I am not going to quietly close it as if the code fix were the whole job.

**Two things that make this easier than it was.** First, there used to be a reason to wait: one
service was running an old build that would have written the *new* key into its log the moment it
restarted, so changing the keys would have recreated the problem. **That service has been updated —
changing the keys today is safe, and nothing in the fleet would re-leak them.** Second, and less
comfortably: that "wait" instruction had been out of date for **fifteen days** before I found it. The
thing it was waiting for was finished on 19 August and nobody went back to update the note, so the
right action kept looking premature. I have written that up, because the lesson is not about this bug
— it is that a stale "wait for X" note does not merely mislead, it quietly stops the work.

**I could not check whether you have already rotated them.** The only clue Kubernetes offers is the
date the secret was created, and that date does not change when the contents are updated — so it
cannot tell the two cases apart. And I am not permitted to read the values. So: have they been
changed? If yes, I will close this today. If not, that is the last thing standing.
