The council is live and working, and it stopped on something that genuinely needs you.

What happened

Your new build (v1.0.1135) carried the docResolveSubject fix, so I applied seed 167 and fired the planner. The loop runs end to end: compose → persist to doc_plans → four critics → deterministic council_decide →
router → run_checks → recompose, superseding the plan each round. Six plan versions were written; the final is 18.4KB, is_current, with the ```criteria fence intact.

Running it flushed out four real defects, each fixed (6c5dc9e13):
1. ExtractFields→UnwrapDeep strips the {result,type} wrapper, so in template context an LLM step's output is the value — {{.proposal.result}} was wrong. Config paths like plan_body_field read raw collected_data and
   correctly keep .result; that asymmetry is the trap.
2. Critic max_tokens 4000→8000 — feasibility's JSON truncated mid-object and council_decide failed closed rather than waving a partial review through. That fail-closed behaviour is exactly right.
3. councilReview.Edit is an int; a critic emitted a string.
4. Two genuine gaps in the plan contract that the council itself proved: the MVP cut must be an ordered, gated step list with prerequisite data steps as step 0, and any number a visitor reads as a score must have
   its exact computation defined.

Why it escalated — two separate causes

Structural (fixed): decideCouncil maps any objection → revise with no severity threshold, so a single low-severity nit from any of four critics blocks approval and burns a round. diagnose_council_decide is shared
Go used by fix-proposer and the council-gate thread, so I did not touch its semantics — I fixed it inside my own agent's prompts instead (object only for medium/high; low nits go to notes with approve).

Substantive — this one is yours. Feasibility's HIGH objection: page_type='provocation' has zero prior rows and no proven build/render pipeline, and the plan folds authoring a whole new page-type render path into an
MVP step. The planner can't resolve this because D2 is stated in its prompt as an accepted default it must not relitigate. So it correctly escalated instead of pretending.

The decision

Per-provocation detail pages:
- A — keep D2's default (static per-provocation pages), but sequence the new page-type render path as its own prerequisite round before the MVP. Honest, slower, puts an unproven render path on the critical path.
- B — recommended: switch to PLAN §7's documented alternative, client-side detail rendering on the existing archive page. That page is already a runtime-fill shell hydrating from the same feed, so there's no new
  page type and no new render path; static pages move to LATER once the daily emitter exists.

Worth noting: the Gauntlet is not the blocker. Honesty approved the final plan outright, and feasibility confirmed the timer/scoring/reset are genuinely client-side-doable — so D1's minimal-real call holds.

Tell me A or B and I'll set it in the compose prompt's D2 block and re-fire; with verdict discipline now in, I'd expect convergence and CP2 closed. The current is_current plan is the escalated version and must not
be built until then.

I also folded the tooldoctraveller2 thread's ruling into the RUNBOOK for Phase 5 — journeys as an additive persistent-context path (never rework Execute), reuse evaluateOnPage for free forced_by overflow
attribution, navigate by symbol not line, and unify needs_experience_replan escalation with their bugs_open/010 candidate (b) rather than building it twice.

---

The fix is correct — and verifying it uncovered a live bug in two other agents

I checked it three ways rather than by reasoning alone: the docs, the code, and a real test.

The docs back it. 001_development_guide §"Common Data Shapes" documents execute_llm_prompt's raw shape as {type: "text", result: "..."} — that's what config dot-paths read, so plan_body_field: 'proposal.result' and
review_fields: ['review_journeys.result', …] must keep .result, and I left them alone. §Step 6 documents the other half: "{{.missing.nested}} → execution error" — exactly the failure I hit. So the asymmetry I
described is the documented contract, not a workaround.

The test proves it. TestUnwrapDeep_TemplateVsConfigPaths (committed, passing) locks all four halves: config paths still resolve with .result; ExtractFields→UnwrapDeep strips the wrapper so template data is the
value itself; {{.proposal.result}} errors on a text step; and {{.review_journeys.result}} renders "<no value>" on a JSON step.

That last one is the part that matters beyond my work. The loud failure is the safe one — I got an error and fixed it. The quiet one silently produces <no value>, and a live-DB sweep shows two agents I don't own
are sitting on it:

┌──────────────────┬────────────────────┬────────────────────────┐
│      Agent       │       Steps        │     What's dropped     │
├──────────────────┼────────────────────┼────────────────────────┤
│ fix-proposer     │ repropose, reframe │ all 6 reviewer outputs │
├──────────────────┼────────────────────┼────────────────────────┤
│ feature-designer │ repropose, reframe │ same family            │
└──────────────────┴────────────────────┴────────────────────────┘

Both inject the council's reviews as {{.review_X.result}}, and I confirmed on the live row that those reviewers are in input_fields, so they go through UnwrapDeep. Their revise loops have been re-proposing without
seeing the objections they're told to address — only the previous plan plus "address every objection". That's the "prompt seams dropping spec intent" trap the travelling-docs workstream already warned about, in a
new place.

I did not change their agents — they belong to the fixloop / feature-builder / council-gate threads, and the council-gate notes explicitly warn to diff the seed against the live row before re-applying because the
roster moves fast. Instead I filed bugs_open/016 with the evidence, the exact scoped fix (templates only — never the config paths), and one flagged uncertainty: content-creator-hero references
{{.call_researcher.result}}, but call_agent returns {response, response_status} and UnwrapDeep has a separate Pattern 4 for it, so that one needs checking separately rather than assuming.

One thing worth naming: this makes my own council's escalations more trustworthy, not less. Mine escalated with its critics' objections fully visible to the reviser — the plan grew 10.6KB → 18.4KB absorbing them,
and journeys went object → object → approve. A council whose reviser can't see the objections would tend to look like it converges while actually just re-proposing.

Want me to fix fix-proposer and feature-designer directly (diffing seed vs live row first),  Jump to bottom (ctrl+End) ↓ s to pick up?

--
Two things resolved, and one still waiting on you.

016 is corrected and closed out on my side. My original caveat told the council-gate thread to leave content-creator-hero alone because call_agent has a different envelope — that was wrong, and they'd followed it
precisely. I proved the general rule instead: unwrapRecursive recurses, so Pattern 4 unwraps the call_agent envelope and then keeps going. A result key can never survive extraction, whatever produced it. Verified
across all four real shapes:

┌──────────────────────────────────────┬───────────────────────────┐
│               Producer               │ {{.field.result}} renders │
├──────────────────────────────────────┼───────────────────────────┤
│ LLM text step                        │ execution error (loud)    │
├──────────────────────────────────────┼───────────────────────────┤
│ LLM json step                        │ <no value> (silent)       │
├──────────────────────────────────────┼───────────────────────────┤
│ call_agent, result-shaped child body │ execution error (loud)    │
├──────────────────────────────────────┼───────────────────────────┤
│ call_agent, domain-shaped child body │ <no value> (silent)       │
└──────────────────────────────────────┴───────────────────────────┘

So content-creator-hero is affected — its hero prompt has been rendering <no value> where the researcher's findings belong, meaning heroes were written without the research they commissioned. It's the last
.result}} in the fleet, so fixing it closes the class. I put the correction in 016 in place rather than editing the council-gate thread's write-up, and left the fix to whoever owns that agent.

I also checked my own council against the pathology, since the feature-builder thread proved it's real on their run 3b084712 and my whole D2 escalation depends on my reviser having actually seen its objections.
Mine is clean: verdicts flip and objection counts move every round (honesty went approve→object→approve, feasibility 5→5→4→3→4), which is a reviser demonstrably reacting — the opposite of their tell. Two reasons it
escaped: my template fix landed before any run that reached a verdict, and my first run died loudly on a text step rather than degrading silently; and my check_results reference was always .results_text, a field
on the unwrapped value, which is correct.

That validation changes nothing about the diagnosis but makes it evidenced: the non-convergence is oscillation across a four-critic panel, which the verdict-discipline change targets (applied, not yet exercised).

Still yours: the D2 decision. Feasibility's objection is HIGH and structural — page_type='provocation' has zero prior rows and no proven render path — so it survives any severity threshold. Re-running now would
burn another ~25 minutes and escalate again for the same reason. The choice is:

- A — keep static per-provocation pages, but sequence the new render path as its own prerequisite round before the MVP.
- B (recommended) — client-side detail rendering on the existing archive page, which is already a runtime-fill shell hydrating from the same feed. No new page type, no new render path, MVP ships on proven
  machinery; static pages move to LATER once the daily emitter exists.

Say A or B and I'll set it in the compose prompt, re-fire, and drive to CP2 and the build round.

B
-----------
