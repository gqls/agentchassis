# SUMMARY 2026-08-19 — the internal linker can link again, and the whole fleet is checked for the defect that silenced it

**What we're trying to do.** Make the internal linker actually work — the agent whose job is to
add links between a site's pages so new pages stop being orphans — and make sure the defect that
silenced it can't quietly happen again anywhere in the fleet.

**Where we've come from.** The linker had a broken "are there any candidate pages?" test from the
day it was created in April: the step that fetches candidates returns a plain list, and the test
asks that list for a named property (`count`) that plain lists don't have. The test could never
pass, so for four months every run ended "no candidates" — including one holding fifteen
candidates at the time — and the agent completed 57 jobs without ever planning a single link.
Nothing errored; the runs read `complete`. A sister ticket (298) noted the same fetch also
capped candidates at the first 15 alphabetically, which never mattered only because of the dead
test. The lane that found both closed yesterday, handing the tickets over unowned.

**What we've done.** One database migration (490) fixed both bugs together, in the order the bug
files demanded: the fetch now returns its result in the shape the test expects, the alphabetical
cap is gone (the per-page text was already bounded, so the payload stays modest, and the cut it
does make is now marked), the prompt was updated to match the new shape, and the repaired test
step opted into a new tripwire so this exact silence can never recur there. Then two
framework-wide pieces so the *class* is closed, not just the instance: an offline checker that
scans every live agent for "a test asking a list for a named property" (it found exactly one —
ours — on its first run), and an opt-in switch on the test step itself that makes an
un-evaluable numeric comparison fail loudly instead of silently picking "no". The whole set went
through the reviewer council: round one sent it back — correctly, catching two real things (a
snapshot-ordering flaw in the migration, accepted with its cost bounded, and an imprecise claim
of mine about a budget counter, retracted and logged in WRONG_CALLS) — and round two **approved
it unanimously**. The migration is applied and verified live.

**Where we are now.** The checker paid for itself the same afternoon: a *brand-new* agent, seeded
today by another workstream, carried the identical defect — caught in hours rather than months,
handed to its owners, and fixed by them (they adopted the tripwire too) before the day was out.
The fleet-wide check now reads **clean: 193 agents, 147 conditionals, zero findings.** A separate
workstream also confirmed that once our linker produces links, a known work-item collision would
have silently dropped all but one link per plan — their fix is going in independently, confirmed
non-interfering. Everything is committed and the council trail is complete.

**Where we're going.** One thing keeps the two bug files open: the real-world proof. The linker
runs a few times a day off its queue (twenty jobs waiting); the first run since the fix will
write the first LLM call in the agent's entire history, and the check is one query — a
`plan_links` row in `llm_call_log` with page names in the rendered prompt. When that lands, 313
and 298 move to closed and this lane ends. The two code-level safety nets ride the next platform
build, with their post-build checks already written down.
