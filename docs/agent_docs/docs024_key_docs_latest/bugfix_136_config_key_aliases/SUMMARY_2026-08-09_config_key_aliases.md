# SUMMARY — 2026-08-09 — config key aliases (`bugs_open/136`)

*Written to be read aloud. First summary in this lane; the milestone is that the work is
finished and the acceptance report is clean.*

## What we're trying to do

Make it impossible for a setting in an agent's configuration to look live while being read
by nobody. The estate configures its agents in the database: each step of a workflow carries
a little bag of settings, and the Go code picks the ones it knows about out of that bag. If
someone renames a setting in the code and does not rename it in the database — or the other
way round — nothing breaks loudly. The old setting is simply ignored, the code falls back to
a built-in default, and because the default is usually the same value the config was asking
for, everything keeps working and every test passes. The configuration becomes decoration.
This bug was one instance of that, and the aim was both to fix the instance and to leave
behind a mechanism that makes the next one visible.

## Where we've come from

The instance was a rename from the word "domain" to the word "pipeline". It landed in the Go
code and never in the database, on three separate actions. One of those actions had a
hand-written patch to cope; the other two did not, and had been quietly ignoring their
configuration for months. It only stopped being harmless on 4 August, when two checks moved
onto an agent whose configuration said "content" — and they filed their findings under
"design" instead, because the ignored setting meant the built-in default won.

The fix, shipped on 8 August, was to give the platform a proper way to say "this old setting
name still works, and here is what it is now called" — declared once, honoured everywhere,
and listed by an audit report so nobody has to remember which definitions still use the old
name. That went live, and was proven working in production with a deliberately designed
experiment that could have failed and didn't.

What was left over was residue: records already filed under the wrong label, dead settings
still sitting in live configuration, and definitions still spelling things the old way. Those
were listed and deferred.

## What we've done

Cleared all of it, on your instruction. The mislabelled records are corrected. The dead
setting on the page builder is deleted. Every place in live configuration that still used an
old spelling now uses the one the code reads. The last action that had never opted in to the
"tell me about settings nobody reads" check is now opted in, with its contract written down
and pinned by tests. Two database migrations, both applied and recorded; one code change,
reviewed by the council and approved first time.

The acceptance report — the thing that tells us whether any of this is real — now returns
completely clean on both of its counts, where it previously named one bad setting and three
families of out-of-date ones.

Two things went differently from the plan, and both are worth knowing.

The first is that our own checking query was wrong. The plan said thirteen places needed
renaming; there were nineteen. The six it missed were nested one level deeper, inside loops,
and the query only ever looked at the top level. It did not fail — it returned a confident,
plausible, incomplete answer. The platform had already fixed this exact mistake in its own
code and left a note explaining why; our workstream just kept using the old query. It is
corrected now, and written up as a trap.

The second is that finishing the last item exposed a live problem nobody knew about. Before
switching on the new check, I listed everything it would complain about. That turned up a
setting called `spec`, used in three places, that the code has never read — and in one of
them it matters: the improvement loop has been trying to say "also rebuild the shared header
and footer" and that instruction has been discarded every single time, sixteen records out
of sixteen.

## Where we are now

The original bug is finished in substance. It stays open in the file, as you ruled, with the
completion written into it rather than the file being moved.

One decision is waiting for you, and it is genuinely a judgement call rather than a task.
Restoring that lost instruction would start triggering full header-and-footer rebuilds about
twice a day — and there is an open bug saying that kind of rebuild can silently throw away
hand-patched content. So the three options are: turn it back on, delete it and accept that
this path never refreshes the shared furniture, or turn it back on only once the other bug
is guarded. Two of the three places are harmless to fix whenever, because they have never
actually run. It is written up as its own bug so it cannot get lost.

One consequence to expect, so it doesn't look like a regression: once the next chassis image
rolls, the audit will start naming that `spec` setting. That is the new check doing its job
on a real defect. The alternative — declaring the setting so the report stays at zero — is
precisely the mistake that a previous fix in this family made, and we are not repeating it.

## Where we're going

Nothing further is planned on this lane. The alias mechanism stays declared even though
nothing uses the old names any more: it is the safety net for old snapshots and stragglers,
and an empty list is the net working, not the net being unnecessary.

The two pieces of general learning have been pushed somewhere they will reach people who
have never read this lane: the "your query only looked at the top level" trap and the
"renaming a setting in a seed can break a later migration that deletes it" trap are both
recorded fleet-wide, and the diagnosis loop's own inability to read agent configuration —
which is why its verdict on the new bug was "cannot tell" rather than "confirmed" — is now
in the debugging guide, because it will mislead the next person who uses that loop on a
configuration question.
