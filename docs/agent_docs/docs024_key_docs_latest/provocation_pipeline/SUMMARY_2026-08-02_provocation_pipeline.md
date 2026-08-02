# SUMMARY — provocation pipeline, 2 August 2026

*Third in the series. The previous two (`SUMMARY_2026-07-31`, `_2026-07-31b`) were
written when the mechanism was designed and then built. This one is written the
day it went live, which is why it exists: the read-out has genuinely changed.*

---

## What we're trying to do

vonc.com promises a new provocation every day. It has never delivered one. The
same entry sat under the words "Today's Provocation" for weeks, because the file
that feeds the site was hand-written and hand-committed — there was no machine
anywhere that could change it. The goal of this workstream is to make the promise
true: one provocation a day, selected and published automatically, with the site's
own claim about itself becoming a statement of fact rather than an aspiration.

## Where we've come from

The rotation problem was never the hard part. A short Python script solved it in
an afternoon — pick the newest entry whose publication date has arrived, put
everything older in the archive, newest first — and we proved it across thirty-nine
dates. The hard part was that the script lived in a documentation folder, and the
cluster cannot run a documentation folder. It could not be scheduled, could not be
pointed at, could not be deployed. We had a correct answer with no way to ask the
question.

So the work became: move the proven logic into the platform proper. A pool table
to hold provocations, a Go action to select and publish, an agent to run it and a
schedule to fire it every six hours. That went through the council review twice —
rejected the first time, correctly, for a reason worth remembering: the action
existed but nothing in the system could invoke it. We had built the engine and no
ignition. The second submission was approved.

## What we've done

The owner rolled a build, and everything downstream of that ran today.

We confirmed the new code was genuinely in the running system rather than trusting
the version number, which this platform has learned the hard way not to do. Then,
before switching anything on, we predicted what the first run would do: we ran the
old script for today's date and compared it against what the website is actually
serving. Identical. That meant a real run could only conclude "nothing has
changed" and stop — which made it the safest possible moment to go live, because
the whole machine would be exercised and the site could not move. That is exactly
what happened, in under two seconds.

Then we made it do the half that had never run. Skipping is the easy path; the
part that writes to the website's repository had never once executed, and an
unwatched writer is a classic way this platform gets hurt. Because today's content
was provably identical, we forced a commit where the only thing that could
possibly change was a timestamp. It committed cleanly.

**And that is where the day earned its keep.** That commit rewrote all 119 lines
of the file to publish one timestamp. Comparing what the new code writes against
what the old script wrote turned up two differences, neither of which breaks
anything: the new code was converting the italic markup in headlines into escape
codes, and it writes the file's sections in alphabetical rather than human order.

Both survive being read by a browser perfectly. That is precisely why nothing
caught them. Our parity test — the one we relied on, the one we told the council
about — compares the two versions *after* reading them in, so to that test the two
files were identical. It had been passing happily the entire time the escaped
version was going to production. Only looking at the published artefact showed it.

We fixed the escaping. We deliberately did not fix the ordering, and wrote down
why: that cost is paid once when the writer changes, not once a day, and pinning it
would mean maintaining a second copy of the file's structure that could drift from
the real one — the exact class of problem this feed exists to remove.

## Where we are now

**Every part of the machine works, and both of its paths have been watched
working.** Selection, feed assembly, verification, the comparison against what is
live, the decision to skip, and the decision to commit. There is no code left to
write for rotation.

**And the site is still wrong.** It still says "Today's Provocation" above an entry
dated 26 July, because the pool contains nothing newer than 26 July and the
selector is correctly serving the most recent thing that has arrived. The machine
is not failing; it is faithfully publishing the fact that we have run out of
material.

So the remaining gap is content, and nothing else. Adding a provocation is now a
database insert — no code, no build, no deployment. That is a deliberate stopping
point rather than an unfinished one: provocations go out as the owner's opinions
under his name, and choosing them is an editorial act, not a build step.

One fix is committed and waiting for the next build to take effect. Nothing is
broken in the meantime.

## Where we're going

The immediate next step belongs to the owner: provocations dated forward from
today. The moment they exist, the site starts rotating daily on its own with no
further intervention.

After that, the generative half — the part that was always the point. A model
proposes candidate provocations from current topics, and a gate checks each one is
safe, interesting, genuinely current, and actually a good provocation rather than
merely a controversial sentence. The pool already has the columns waiting for it:
where a provocation came from, what the gate decided, and when. The owner has
settled that there will be no human approval queue in front of publishing, so the
gate has to be good enough to be trusted alone.

Then categories, which are a bigger change than they sound — the game engine reads
exactly one provocation per site, so more than one a day means changing the engine,
not the feed. And after that, paired mode: private, team-based, organiser-set
provocations, which needs an identity system we do not yet have.

The honest summary of today: we finished the machine, proved it, found a defect in
it that our own tests were structurally incapable of seeing, and stopped at the
point where the next decision is a human one.
