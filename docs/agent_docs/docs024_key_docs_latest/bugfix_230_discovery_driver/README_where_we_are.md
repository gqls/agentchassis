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
