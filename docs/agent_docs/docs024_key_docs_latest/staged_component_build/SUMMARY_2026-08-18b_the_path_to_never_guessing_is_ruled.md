# SUMMARY — 2026-08-18b: the path to "never guess" is ruled

*(Series: follows `SUMMARY_2026-08-18_the_instrument_caught_its_first_real_bug.md`, same day —
the morning one records the instrument working; this one records the plan it produced.)*

**What we're trying to do.** When one part of the platform hands work to another, it sometimes
has to hunt for a value rather than being told where it is, and when the hunt finds two
different answers it used to pick one by luck. The goal, ruled by the owner weeks ago, is that
the platform never guesses: an explicit instruction wins, and a genuine ambiguity is refused
rather than resolved by chance.

**Where we've come from.** A recorder has been writing down every disagreement the hunt finds —
about seven and a half thousand notes in two days. This morning we believed one buildable fix
would clear most of them. Opening the one code citation that claim rested on showed it was
wrong: the fix could only reach about a quarter of the noise, because most of the searching is
done by a safety net that tops up six values on every lookup whether anyone asked or not.

**What we've done.** Resolved a day's worth of recorded disagreements by hand and split the
noise into four kinds — most of it wrong answers nobody reads, a small part right answers held
up by an unwritten rule of code ordering. Filed that as a proper bug, with the fixes ranked.
Filed a second bug from yesterday's outage: the retry machinery burned all three of its
attempts inside the same three-hour blip, and the owner has ruled a blip should re-queue work,
not kill it. Built, tested and submitted the quarter-of-the-noise fix. Measured exactly who
depends on the safety net: the three business values are used undeclared in fifty-five places
and must never be touched; the three page values have one soft dependent, fixable with a
one-line setting.

**Where we are now.** The owner has chosen the sequenced path: ship the fix, pin the unwritten
rule, gate the three page values (settings fix first), repair the few real same-page-two-shapes
clashes at their source, and only then arm "refuse rather than guess" — deliberately armed
last, when there should be almost nothing left to refuse, so it stands as a guarantee for the
future rather than a cleanup of the past. Each step advances on what the recorder shows, not on
a date. The first fix's review verdict is expected imminently.

**Where we're going.** Steps two and three are the next sessions' builds; step one ships with
the next fleet release. When the sequence completes, every disagreement the hunt can find will
either be impossible, explained, or refused — and the recorder goes quiet because there is
nothing left to record, which is the only kind of quiet worth trusting.
