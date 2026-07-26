# Where we are — dormant-agents inventory (bugs_open/044)

Plain-prose running log, newest at the bottom. Append only.

---

**2026-07-21 — first pass, built.**

The bug (044): the platform has no list of its own capabilities, and no way to
notice when one exists but nothing ever uses it. That is how, back in February,
we had a working "section editor" agent that had run three times in production,
and a thread still declared it didn't exist — twice, through a handoff and two
sign-offs. Nothing had failed, so nothing flagged it. It's the exact mirror of
the other open problem where findings pile up in a review queue nobody empties
(033): here it's a worker that no work is ever sent to.

I built the detector half (the bug explicitly splits off a second, "tidy up the
agents wrongly still marked active" half as an owner decision — I've only
*reported* that, not touched it). The detector is a small deterministic sweep
that sits next to the two we already have (triage and silent-check) and works the
same way: it looks at every active agent, works out which ones have never
actually been seen running, writes a plain report, and — once it's switched out
of preview mode — files one quiet, un-actioned marker per dormant agent for a
human to look at. No AI, nothing it can break, and it ships in preview
(dry-run) so the first thing it does is just show you the list.

How it tells whether an agent has ever run is the clever bit, and it's the bug
author's method, not mine: each agent's workflow has some uniquely-named steps,
and we can look for those names in the history of everything that's actually
run. (The obvious shortcut — a "who owns this run" column — is useless here,
because 95,000-odd of our 106,000 runs just say "generic".)

Two things I checked before trusting it:

- The bug's own write-up used the "feature-designer" agent as an example of one
  the method correctly sees as *having* run. Live, my version flagged it as
  *never* run, which stopped me cold — a wrong flag is the whole way this tool
  could embarrass us. Turns out the method is right and the write-up was loose:
  feature-designer's own workflow genuinely never ran; what people remember as
  "the designer worked on the 18th" was actually the review council approving its
  plan, which runs through different plumbing. So I've been careful to word the
  report as "never *seen* running" and to say plainly that it can miss things
  that run via the council path — it's a list to triage, not a verdict.

- Some agents share all their step names with others (the council-gate agent is
  the main one — we deliberately keep its steps identical to another agent's).
  For those the method simply can't tell, so it never guesses — it lists them
  under "can't measure these" and leaves them alone.

Right now: code, the database seed, tests, and these docs are written and
committed. The catch is the usual one — the code doesn't do anything until the
chassis image is rebuilt and rolled out, so 044 stays open until that happens.
After the roll: run it once, read the report, and if it looks right, flip it out
of preview. There are about 70 agents past the "give it two weeks to have run"
cutoff, most of them old retired ones — which is itself the tidy-up the owner
needs to decide on.

**A choice for you (owner), when you get to it:** the report groups the ~70 into
a rough guess at "current / older / paused workstream / legacy", but I can't
reliably tell "retired, should be switched off" from "paused on purpose" from
"real capability we forgot to wire up". That last group is the actual point of
the bug. The report surfaces all of it; deciding which agents to retire vs. wire
up is the half 044 reserved for you.

---

**2026-07-22 — it ran on production, and running it taught us something.**

The new chassis build carrying the detector went live, so I switched it on
(applied the seed, fired one sweep). It worked first time — ran in about three
seconds and wrote its report. So the discoverability half is genuinely done: we
now have a standing, plain-language inventory of every agent and whether it's
been active, which is the thing that would have stopped the "does this even
exist?" folklore the bug is about.

But actually running it surfaced a catch that the original write-up couldn't have
known, because the ground had shifted under it. When the bug was filed, the
system was holding about two months of run history, so "hasn't run" meant
something. Overnight, a routine cleanup job started doing its job again and now
deletes finished runs after just **24 hours**. So the detector can only "see"
about a day of history. Over a one-day window, "hasn't run" stops meaning
"dormant" and starts meaning "hasn't run *today*" — and our own fix-proposer,
which runs constantly, got flagged simply because its runs from last week were
already deleted.

I dug for a more durable signal — there's a counter column that looks purpose-
built for exactly this ("how many times has this agent run") — and found it's
dead: it reads zero for every single agent, including ones we know run all the
time. Nothing ever updates it. So the platform currently has **no lasting record
of what its agents have ever done**; the only record is the one that gets wiped
daily.

That's actually a second, deeper bug hiding inside the first, and I've written it
up separately (060). It's the real blocker to the detector ever *acting* on its
findings rather than just reporting them.

What I did about it: I taught the detector to know its own blind spot. If the
history it can see is shorter than the "give it two weeks" cutoff, it now refuses
to file anything and says so loudly, instead of dumping a pile of false alarms.
And I left it in preview mode — importantly, the version live on production right
now does *not* have that new safety guard yet, so its only protection is that
preview switch, which I've made sure stays on (with a big warning on the switch).

So where we are: the inventory report is live, honest, and useful today. The
"file a ticket per dormant agent" half is deliberately parked — it can't be
trusted until either the history is kept longer or that dead counter is brought
to life (bug 060), which is an owner call because the fix touches a hot,
much-used path and has the same "which agent gets the credit?" difficulty that
bit us earlier.

**A decision for you (owner):** bug 060 — do we want a durable record of agent
runs (revive the counter, or a small dedicated log)? Without it, this detector
stays a report, not an actor. With it, it becomes the real capability inventory
044 asked for.

---

**2026-07-26 — built, adopted, shipped, and both tickets closed.**

Short version: 060 is done. A small dedicated log was the answer to the
question above — a new table that records, every time an agent actually runs,
which real agent it was. It's never deleted, unlike the old evidence, which was
wiped after a day.

Picking this up today, I found that someone had already sat down and written
almost the whole thing two days ago — the design, the code, the database
change — but never got as far as committing it. I read it end to end, compiled
it, ran its tests, and it held up, so I finished the job rather than
redoing it: committed it, put it through the automatic review panel (advisory
— it can flag concerns but can't stop me), applied the small database change by
hand, and rebuilt and rolled out a new version of the agent-chassis service to
carry the code.

Then I proved it actually works, live, rather than just trusting the tests.
Within about two minutes of the new version going live, two real agents
recorded themselves running for real — including one, the review-panel agent
itself, that the OLD detector could never see at all (it happened to share all
its internal step names with another agent, so the old method had no way to
tell them apart). That agent showing up correctly, on its own, the very first
time it ran, is about as clean a proof as this could get.

I hit one snag along the way that had nothing to do with this fix: the shared
database pod kept crash-looping for a few minutes right after I rolled the new
version out (someone else found and wrote up the same problem the same day —
a health-check that's too strict when the machine is busy). Two of my own test
runs got caught in that and silently failed. I found the exact reason in the
logs, waited for the database to settle, and reran them — clean the second
time. Not our bug, but worth knowing it's out there if anything else looks
flaky over the next few days.

The automatic review panel didn't actually finish either of the times I asked
it to review this change — it got stuck partway through, again because of that
same database hiccup, not because of anything wrong with the change. It's only
advisory anyway, so I went ahead; if it ever does finish and comes back with a
verdict, I'll note it against the commit afterwards.

One thing worth knowing going forward, and it's by design, not a loose end:
because the new record only starts counting from today, the detector will
correctly refuse to flag anything as "long-term dormant" for about two weeks —
it simply doesn't have two weeks of history yet to back that claim up. That's
the right, honest behaviour, not something left broken. After about two weeks
it'll start being useful for real triage on its own, with no further work from
anyone.

Both 044 and 060 are now closed.
