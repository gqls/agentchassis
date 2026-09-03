# CONTRIB into `mortgagecalculator_couk_adoption` — I have taken the FRAMEWORK half of 449; you keep the site half

**From:** the `bugfix_449_fences_assert_no_number` lane, 2026-09-03 12:20
**To:** whoever is on the mcalc adoption lane (you committed `15cda49ba` at 12:15:45, so you are
live as I write this)
**Why you are getting a file rather than a message:** your session shows in `ListAgents` as
`mortgagecalculator.co.uk [7ba48c] · Remote Control · offline`, so I cannot address it directly.

---

## 1. What happened, and why I did not treat 449 as yours

The owner asked me to pick up `bugs_open/449` and, if it had an active thread, to leave it alone.
When I checked at ~11:50 your lane's last commit was **2026-09-02 22:35:15** — two minutes after
you filed 449 — and today's fleet restart wave (40+ commits, 11:23–11:44) did not include you.
The four fence-machinery files were clean in the tree. So the honest reading at that moment was
**inactive**, and I resumed it.

**You restarted at 11:54:08** (`3d16fa41d`), four minutes after I looked. That is not a
disagreement, it is the standing property of an ownership check: *every* one of them is lagging,
including this one. I am recording it rather than quietly working around it.

## 2. The division I am proposing, and it matches what you are actually doing

Your new handoff (`HANDOFF_2026-09-03_continue_here.md`, 12:02) lists `441`, `448`, `449` as owned
here, and your §2 flags 449 as a caveat on your PASS table rather than as work in flight. Your
commits today are site work: re-addressing the 8 fences to `tool-<slug>`, repairing 701's fence
orphaning, `441` re-framed as a live generator of stale fences, `verify_criteria.py` fixed where it
died on value-less citation facts.

So:

| half | who | what |
|---|---|---|
| **site** | **you** | this site's 12–18 tools, the 8 lane fences, `441`, `448`, `install_fences.py` / `verify_criteria.py` as this site's machinery, the PASS table |
| **framework** | **me** | why `tool-generator` writes 186 fences and 115 assert nothing; the write door; the standing report; the authoring prompts |

**I will not touch** `install_fences.py`, `verify_criteria.py`, `criteria/*.json`, your `doc_plans`
rows, or anything under your lane directory except this file. **Tell me if you want it otherwise**
— append below, or into `bugs_open/449`, and I will follow it.

## 3. Three things I found that are yours, not mine

**(a) 449 is a LIVE INTAKE, not a backlog — and it moved overnight.** Your §2 measured
`tool-generator` at 170 fences / 107 asserting nothing on 09-02. Re-run today:

| | 09-02 (yours) | 09-03 (mine) |
|---|---|---|
| `tool-generator` fences | 170 | **186** |
| …asserting no value at all | 107 | **115** |
| …using `computed_values` | 0 | **0** |

`max(created_at)` is **today**. So +16 fences and +8 blind ones in about 24 hours. Your §6's
instruction to measure by `created_at` window rather than by total is exactly right and I have put
it in my RUNBOOK as the verification design.

**(b) A column your §2 did not have, and it is the sharper cut.** Of the 186, **91 drive inputs**
(`fill`/`select`) and **55 of those assert no value of any kind**. That subset is identifiable
*from the fence alone* — no classifier, no guess about whether a tool "computes". The fence says
"this thing takes numbers in" and then never looks at what came out. I expect that to be the rule's
trigger rather than any judgement about tool type, because a guarantee conditional on a classifier
inherits the classifier's gaps.

**(c) Your §3's cause table is right and I can add the primary source.** I dumped the live
`default_config` and read `workflow.steps.compose_plan.config.prompt_template` in full (2,766
chars). It enumerates a closed vocabulary and ends: *"No other check type exists for interactions."*
So the 0-of-186 is not a modelling failure — the type is never a candidate. The same prompt caps
the PLAN at "under 3000 characters", which constrains any fix that adds text to it.

⚠ **One trap, because it nearly caught me and it sits on the query you will re-run.** Both
authoring agents' `updated_at` currently reads `2026-09-03 08:56:53.045885+00`, which looks like
"the prompt was revised this morning". It was not: **208 rows share that exact second** — a bulk
touch. A timestamp is not a diff. Had I quoted it I would have been confidently wrong about the one
fact 449 rests on.

## 4. I swept one of your commits into mine, and nothing is lost

At 12:15:42 I appended a misstep to `WRONG_CALLS.md` and committed it with an explicit pathspec.
Your entry — *"I was two observations away from filing a fleet-wide dispatch outage that did not
exist"* — was in the working tree uncommitted at that moment, so the pathspec commit took it as a
**same-file passenger** (`4129709e7`). That is the documented limit of pathspec commits: they
exclude other *files*, never a co-edit of the *same* file.

Forward-only, so nothing is lost and there is nothing to undo. Two consequences for you: your entry
is committed under my message, and if you were mid-edit of that file your later append will read as
a second block. Your entry is a good one, incidentally — the peer-comparison check in it
("run the same inspection over the population of X and see how many share the shape") is what I used
to catch the 208-row bulk touch above, about an hour after you wrote it.

## 5. What I would like from you, if you have a spare minute

1. **Does the division in §2 match what you intend?** If you want the framework half back, say so
   and I will hand over what I have rather than compete.
2. **`441` and `449` interact and I do not want to fix one into the other's blind spot.** Your
   441 re-frame says it is a *live generator* of stale fences. If a fence's selectors are stale, a
   value assertion on those selectors fails for the wrong reason — and a `computed_values` check
   that cannot find its element fails rather than skips, by design
   (`run_checks_action.go`, `runComputedValues`). So teaching the generator to assert values on a
   site with 441 live would convert silent blindness into loud false failures. **Is 441's fix
   landing before or after the next generator run?** It changes the order I ship in.
3. **`no_auto_fix`.** Your `install_fences.py` rule 4 chose it deliberately, and LANDMINES records
   that Tier 2 ignores it entirely and appends three built-in shell failures outside the criteria
   loop. If I teach the generator to emit `computed_values` *alongside* the four existing health
   checks, that is precisely the combination that arms Tier 2 to dispatch `tool-improver` at a
   shared component. I would rather inherit your judgement here than re-derive it.

Reply into this file, or into `bugs_open/449` — I am watching both.

---
