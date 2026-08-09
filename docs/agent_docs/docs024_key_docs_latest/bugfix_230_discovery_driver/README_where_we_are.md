# Where we are — discovery driver (bug 230)

## 2026-08-09 — picked up, verified, plan written

We found that the machinery which examines live sites for problems — sixty-one separate
checks across three inspector agents — has no clock. It runs only when a person points it
at a site. So a broken page on a site nobody happens to be working on stays broken and
unrecorded, indefinitely. This was filed yesterday as bug 230 by a thread that then
finished up, so nobody owned it. I checked nobody else was working it, re-ran every
measurement in the bug file against the live database today, and it all still holds. The
clearest example: two pages on finetuning.uk — a customer site — are serving empty
sections right now, the check that would catch them works, and nothing has run it.

Two useful things came out of the research beyond the bug itself. First, the pause was
deliberate: the old sweep was switched off in May during core build, and that decision is
on record — this isn't a mystery, it's a paused mechanism nobody un-paused. Second, the
old sweep couldn't simply be switched back on anyway: its site-picking is unfair in a way
that was already on record, and it skips any site with a big backlog of open findings —
and the two sites with the biggest backlogs today are exactly the two being worked on
hardest. Switching it on would examine everything except the sites that most need it.

The plan: give the three inspector agents a fair rota. A small table remembers when each
site was last examined by each inspector; every hour, each inspector takes the site that
has waited longest, if any site has waited more than a week. No site can fall off the
rota by being uninteresting or by having too many findings. It only *detects* — it files
findings and fixes nothing, which is the mode the system was designed to support while
the fixing side stays a separate decision (that decision, bug 083's, is still open and
this work deliberately doesn't touch it). A daily check watches the rota itself and
reports if any site is going unexamined or if runs are being silently dropped — so the
silence this bug is about can't come back quietly.

Cost, measured rather than guessed: one full examination of one site used two AI calls.
The rota at full speed is about nine examinations a day. This is small.

Decision I'd like a view on (not blocking — defaults are conservative and every knob is
one UPDATE, listed in the runbook): the rota ships ON at one site per inspector per hour
maximum, each site roughly weekly. If you'd rather it start OFF, or slower, say so and
it's one statement. The separate, bigger question — when to turn the *fixing* side back
on — stays with you, as bug 083 records; the findings this rota accumulates should make
that call better-informed.

Off to the council for review before anything is applied.

## 2026-08-09 (later) — the question answered itself, and the plan holds

While this was at the council, the thread that filed the bug recorded your ruling in the
bug file: the missing driver is a defect, and there was never a cost decision standing
behind the switched-off rows. Better still, the real reason for the original pause turned
out to be on record all along — it was about sequencing (don't detect what nothing can
fix yet), not budget. That's exactly the line this plan already walks: the rota only
detects; the fixing side stays off pending your separate call on bug 083. So nothing in
the plan changes, and the "start ON or OFF?" question above is settled — ON.

## 2026-08-09 (afternoon) — approved, switched on, and working

The council approved it first time round, with four pieces of advice and no serious
objection. Two of them were genuinely useful and got built in before anything went live:
the installer now refuses to run if the three inspector names don't match real, active
inspectors (a typo would have made a rota that silently never ran — exactly the kind of
failure this whole job is about), and the new machinery now has its own note in the
system's records so the next person to touch it finds one authoritative description
rather than a trail of bug files.

Then it went live, and the first cycle worked exactly as drawn: within two minutes the
three inspectors had each taken the first site on the rota (robot-hands.com), examined
it, and filed real findings — seventeen assets that were never deployed, some stray
markdown, and more. The daily watchdog ran too, reported everything healthy, and wrote
its report where a missing report is itself detectable.

One proof remains, and it's the one the bug itself asked for: those two empty pages on
finetuning.uk should be found by the rota on its own, with nobody pointing at them —
that's due within a few hours as the rota works through the site list, and I'm watching
for it. When that lands, the bug is closed in substance: a broken page on a site nobody
is looking at can no longer stay invisible.
