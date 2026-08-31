# SUMMARY 2026-08-31 — cascade attribution (bugs_open/390, now CLOSED)

## What we're trying to do

Our sites are checked automatically for text that is too faint to read against its
background. When the checker finds a spot, a repair agent writes a small stylesheet fix. The
problem this lane existed to fix: the agent's fixes were technically correct but kept
losing — the page's own styling out-ranked them, so the "repaired" spot stayed broken, the
checker filed the same complaint again three days later, and the loop never ended. Nobody
could see this, because the repair was marked complete the moment it was *written*, not when
it *worked*. We set out to make the checker measure exactly what a repair must beat, hand
that measurement to the repair agent, and prove the whole loop closes: repair, re-check,
complaint gone.

## Where we've come from

The bug was split out of an earlier case (352) on 25 August by owner ruling. Measurement
first: across seven sites, three-quarters of "completed" repairs had been filed again,
97 times with byte-identical colours. Three changes shipped 25–26 August, each through the
review council: the repair agent's prompt stopped instructing the losing move (mig 616); the
checker gained *cascade attribution* — for every faint spot it now identifies the exact
styling rule that wins the fight, proves it by removal in a real browser, and files the
requirement a repair must exceed (commit `ea64845e0`); and the agent now consumes that
requirement, with an honest "this one is genuinely unwinnable" park for the rare rest
(mig 635). A fourth change (mig 655) fixed a subtlety found live on day one: two co-present
prompt instructions were being adjudicated by the model rather than by our precedence
language, so we fenced them so only one ever renders. Five predictions (P1–P5) were
registered in advance, each with its disconfirming result stated before the audits ran.

## What we've done

Proven every prediction at the artefact, not the status. The repairs themselves: 16 of 16
theme-branch repairs after the prompt fence carry no `!important` and every selector strictly
exceeds its measured requirement, verified by our own arithmetic against the served file
(served, git and database copies hash-identical on every graded site). The attribution: ~60
worked pairings across seven sites, the designed under-claim (inherited footer colours) the
only systematic gap, zero cases of the prover fooling itself, and the "unwinnable" park never
once needed. And the closing observation, graded 31 August at the 29 August re-audits: **the
checker that filed the original complaints went back to all three repaired sites and filed
none of them again** — on the remortgage site it demonstrably completed (it flagged a
brand-new problem the same minute and said nothing about the five repaired spots). On the way
we also found, filed and evidenced two platform bugs that are not ours: 416 (every audit of a
site past ~two dozen pages has been silently timing out for weeks) and 396 (a site's next
full design run erases appended repairs).

## Where we are now

**Case 390 is closed** — fixed, live, council-reviewed, and confirmed by the platform's own
instrument. The file has moved to `bugs_closed/`, the mechanism is registered (VIZ-018), and
the transferable lessons are in the debugging guide: assert the *outcome* of a repair, never
the method; and fence co-present prompt instructions, because the model — not your precedence
prose — adjudicates between them.

## Where we're going

This lane ends here. Four loose ends live elsewhere, all recorded and unowned: bug 416 (the
big-site timeout — cheap interim fix spelled out in the bug file, burned thirteen more audits
over the weekend); bug 396 (design-run erasure — it will eat these repairs at each site's
next design run, the audit will honestly re-file and the agent re-repair, an expensive loop
396 must break); a sweep of other agent prompts for the same contradictory-instruction shape;
and a probe extension for inherited-colour attribution sketched in VIZ-018.
