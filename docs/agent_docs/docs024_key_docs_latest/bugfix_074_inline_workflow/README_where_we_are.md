# Where we are — bug 074, the scheduled task that fired every day and did nothing

Plain-prose log, append-only, newest at the bottom.

---

## 2026-07-26, evening

Someone filed a bug yesterday about scheduled tasks. The shape of it: you can write a task's
"programme of work" into the task row itself, the system accepts it without a murmur, and then
ignores it completely. The task still fires on schedule, still reports success, still updates
every timestamp anyone would look at. It just never does the work. Three tasks were written that
way. One of them had been doing nothing since the day it was created.

The thread that filed it fixed its own two tasks and left the general problem, plus one task
belonging to another team, for whoever picked it up. That's this.

**The first thing I did was read the code rather than the bug report**, and it changed the fix.
The report said the system doesn't support programmes-of-work written into the message. That isn't
quite right — it does, and other parts of the platform use it constantly (58 live jobs came
through that route). What's actually broken is narrower and more interesting: the scheduler builds
the message itself, out of the task's own columns, and buries whatever the author wrote one level
deeper than the place that reads it. Both halves look correct if you only look at one of them.

That distinction mattered because it meant there were two honest fixes: teach the scheduler to
find the buried programme, or refuse to accept the shape at all. A closed case from five days ago
had already ruled on almost exactly this question — don't teach the scheduler to go rummaging in
the payload, because sooner or later it eats a legitimate field with an unlucky name. So: refuse.
The owner agreed, with the alternative on the table.

**What went in.** A database rule that rejects the bad shape the moment anyone tries to write it —
live immediately, no deployment needed, and it fails loudly at the point where the mistake is
actually made rather than quietly at three in the morning every night. Then the repair: the one
remaining broken task got its programme moved to where the system genuinely reads it, copying the
pattern the other thread had already proven yesterday. And a change to the scheduler itself so
that if this ever happens again it says so and refuses to fire, instead of manufacturing a
successful-looking job for work it never sent. That last piece is committed but won't be live
until someone builds the scheduler image; I've said so plainly rather than implying it shipped.

**The repaired task is the good news.** It belongs to the claims-verification work — its job is to
re-check every published number on our sites against the database that produced it, and flag
anything that has drifted. It had never run. Not once. It ran for the first time this evening: it
checked 24 figures across four sites, re-synced 13 of them, and raised three items for a human to
rule on where a published figure has moved outside what the wording allows. Council round counts
have each gone up by one; the platform now runs 14 live sites where the copy says 12.

I staged that carefully — a dry run first, which reports what it *would* do and writes nothing —
because it rewrites the approved-numbers list on two sites other people are working on right now.
That turned out to be worth it twice over: the dry run's report caught me quoting a stale figure
of my own (see below), and gave the two affected threads something concrete to read.

**Then I broke something on purpose.** A sweep that reports "all fine" proves it was deployed, not
that it works — you only learn whether a detector detects by giving it something to find. So I
corrupted one figure on one site to ten times its real value and forced a run. It caught it, said
exactly why, and put the real number back. That is the proof; the green run before it was not.

**And that induced fault turned up a second, smaller problem**, which I've filed rather than
fixed. When one of these "a number has drifted" notices is already open for a site and a
*different* number drifts, the new one is dropped — and the run still reports that it raised it.
Nothing is lost forever (it comes back next time round), but while the notice sits open it
describes the wrong problem, and the report claims a record that doesn't exist. The fix touches
machinery shared by every detector in the fleet, so it belongs to the team that owns that
machinery, not to me on my way past.

**One thing I got wrong, recorded properly.** I wrote a count of "how many sites have figures to
check" into my notes from a query I'd run minutes earlier. Another session wrote to three of those
same rows in between, and my figure was already wrong when I typed it. The sweep's own report is
what contradicted me. The lesson isn't "check your figures" — I did check, and the check aged in
about ten minutes. It's that in this tree a measurement starts going stale the moment the query
returns, including your own.
