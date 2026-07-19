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
The run just landed and is composing. Here's where things stand.

D2=B is applied, and running it found three more defects — all mine

I wrote B into the compose prompt with its rationale (so the planner can't re-derive it), plus a hard constraint so B couldn't reintroduce the original defect in a new coat: opening an archive entry must be a real observable
outcome (deep-linkable fragment + a populated detail region), not a class toggle or an empty panel.

Four council runs followed. All escalated — and every escalation was correct, each exposing a bug of mine rather than a fault in the plan:

Verdict discipline worked. Objections per critic fell from 3–5 to 1–2, low-severity noise largely vanished, honesty approved in every round. Two runs came within a single medium objection of approval.

load_context was lying by omission. It filtered component_level='tool', but gauntlet-interface, gauntlet-cta, provocation-card and lobby-grid are all level='section'. So the context showed one component when five are active and
attached — and tool-gauntlet genuinely does have gauntlet-interface attached. The planner asserted the component existed; the critic could see no evidence and objected. Correctly, five rounds running, across two independent
runs. I'd been reading that as a stubborn critic; it was a truthful one reading a misleading brief.

A compose truncation death spiral. llm_call_log settled it: recompose output ran 13303 → 12599 → 14138 → 15499 → 16000/16000. Each round the plan grew absorbing objections until it hit the ceiling and truncated inside the §5
criteria fence — at which point two critics objected that it was truncated, an objection revising can never clear because revising makes it longer. Same class as the article-body fix, new place. Fixed at both ends (32000
ceiling plus a length-discipline rule prioritising the criteria fence), because raising the ceiling alone only delays it. Worth noting: the council caught its own document being truncated — otherwise a plan with an unparseable
criteria fence would have reached the acceptance ladder.

My own operator error. The first D2=B trigger sat in the Kafka topic ~10 minutes before being consumed, my poll window missed it, I re-fired, and two councils ran concurrently on the same key. Harmless here — and the two runs
agreeing was actually informative — but it's exactly what the coverage rule exists to prevent. I now check for in-flight runs before every fire.

Right now

Run 054b358a is live and composing with all three fixes in. It had been queued behind a fleet-wide quiet period (chassis consumed nothing from 15:27); I deliberately did not restart the chassis — single healthy replica, 6h19m
uptime, zero AWAITING_RESPONSES backlog, and a prior run had landed fine after similar lag, so the evidence pointed to lag rather than failure, and a restart would have disrupted every concurrent thread.

Judge it by two things: whether the persisted plan ends with <!-- END EXPERIENCE_PLAN --> and a closed ```criteria fence (proves the truncation fix), and whether feasibility stops objecting about component existence (proves the
context fix). If it converges, that's CP2 closed and the MVP build round is next.

It runs ~25 minutes. Want me to keep watching it through to a verdict, or pick it up next session?

---

## 2026-07-19, later — the reviewer finally stopped finding our bugs and started finding the plan's

Picking this up from the resume handoff. The first job was the small resilience change I'd recommended last time: stop one flaky critic from killing a whole council run.

Reading the code before writing it changed the shape of the fix, in two ways I want to record because the original recommendation would have caused a quieter problem than the one it solved.

The idea was to let a critic that errors simply not vote. The machinery already supports that — a missing opinion counts as an abstention, and it refuses outright if *every* critic goes missing, so silence can never be mistaken for approval. But my earlier note said to send a failed critic straight to the counting step. That would have skipped every critic *after* the failed one too, turning one dead reviewer into three. So each critic now falls through to the next one instead.

The second correction matters more. I'd said to apply this to all four critics. Three, in the end. The honesty auditor — the one that checks nothing is invented — is the only seat with a blocking veto, so letting it quietly abstain would mean a plan could be approved with the anti-fabrication check never actually run. That is precisely the failure this whole subproject exists to prevent. So a dead honesty auditor still stops the run, and I've written the reason into the file next to the setting, because the inconsistency looks like an oversight and someone will otherwise tidy it away.

Then I ran it. **It survived all five rounds — the first time that has ever happened.** Every previous run died in round one or escalated early. Round four got to three approvals out of four: one objection away.

It still didn't get an approved plan, but the reason has changed, and that's the news. Every previous escalation was the council correctly reporting a bug in our own tooling. This one isn't. Both big fixes from last session held up under sustained pressure — the plan document stayed complete and stable in size through all five rounds instead of growing until it truncated, and the critic that had been objecting for five rounds that it couldn't verify a component now approves in round one and spends its objections on real build risks instead.

What it escalated on this time is the plan being too ambitious, and us not listening.

One critic — the one whose job is arguing for a smaller first version — said essentially the same thing in four of the five rounds: this Arena rebuild isn't needed for the core game to be playable, defer it. The plan never cut it. Meanwhile every revision answered a different critic by adding more detail, which made the plan bigger, which set the scope critic off again. Round four to five is the clearest case: two critics went to approve, then the revision that satisfied a third pushed one straight back to objecting.

I found why, and it was our fault rather than the model's. The revision prompt introduced that critic as "advisory". That's true of its voting rights — it can't veto. But it's false in practice, because the council revises if *any* critic objects at all. So we told the writer that objection was optional, and it treated it as optional, entirely reasonably. Meanwhile it was blocking approval exactly like a veto would.

You chose to fix that by making the scope cuts binding rather than by silencing the critic, which I think is right: that critic has been correct every round — the plan genuinely is over-scoped. The revision step now has to either cut what it's told to cut, or say in one line why the game can't work without it. And it's told outright that answering one critic must never grow the plan to satisfy another. Run eight is going now.

Two other things worth knowing.

The resilience fix I opened with is deployed but **unproven**. No critic flaked during the run, so the new path never actually executed. I'd rather say that plainly than count it as tested.

And chasing a warning from last session's notes — that some site components were "deactivated across sixteen pages" — turned up something worth having, though it wasn't what the note said. That claim mixed up two different things; every one of the site's page-level components is fine. What is real: the site's header, footer and page-head all point at library entries marked inactive, and the repair job for that has been sitting in a queue since the 11th, eleven days, never picked up. Nothing is visibly broken, because the already-built versions are still being served — but it's stale, and it's the same fault another thread diagnosed this morning on the buttons bug: the system detects a problem, files it, and then nothing ever consumes it. Theirs died in one queue, this one dies in another, one step earlier. I added the evidence to their bug rather than opening a second one, since it's the same disease.

Worth flagging for when we build: this site's homepage is queued for a generic rebuild, and the council itself spotted that our planned edits could be wiped out by it. That's a sequencing problem to solve before any building starts, not after.

---

## 2026-07-19, later still — it approved a plan

Run eight came back unanimous. All four critics, no objections, the workflow finished on "complete" instead of escalating. That's CP2, the thing this phase has been trying to reach since the 17th.

Two things I checked before believing it, both of which mattered.

The first is a trap I'd created myself earlier today. Because I'd just made three of the four critics able to drop out without killing the run, an "approved" verdict could in principle mean *nobody objected because nobody voted*. So I checked the count: four reviewers, zero abstentions. It's a real unanimous approval, not silence dressed up as consent. I've written that check into the runbook as mandatory for every future approval, because it's exactly the kind of thing that gets skipped once the result starts looking routine.

The second is whether it approved for the right reason or just got tired. The scope critic — the one we'd been ignoring for four rounds — asked repeatedly that the Arena rebuild be dropped from the first version. The approved plan now marks the Arena "coming soon" and defers it. And where it *does* keep something that critic had questioned, it now says in one line why: those buttons need only a data fix, no new build, so testing them costs nothing. That's precisely the discipline I put in this morning — either cut it, or say why you're keeping it — and the writer followed it in its own words. It also reached that position at round three and held it through rounds four and five, where in the previous run it had flipped straight back to objecting. The oscillation is gone.

The plans also stopped growing. Across the run they went 15.5k, 13.6k, 13.7k, 13.9k, 14.4k — it shrank after the first round and stayed flat, where the run before it drifted upward and the run before that ran off the end of its size limit mid-sentence.

What I'd flag, plainly: the flaky-critic fix from this morning has still never actually run. No critic has failed since I put it in, across ten rounds. It's correct as written and it's in place, but I'd be overstating things to call it tested, so I'm not going to.

The build phase is now unblocked — until today the live plan was an unapproved draft that nothing was allowed to build from, and now it's the approved one. Before any building starts, though, the homepage is queued for a generic rebuild that could wipe out the edits we're planning, which the council spotted on its own. That's a sequencing job to sort out first.

---

## 2026-07-19, end of session — I sent my own finding to the diagnosis loop and it told me I was wrong

You asked me to put the clobbering problem through the diagnosis loop rather than just fix it. That was the right call, and not for the reason either of us expected.

Before filing I grepped the bug directories, as the house rules require, and found that another thread had filed a closely related bug hours earlier — a tool fix that gets written to the database correctly and never actually reaches the live page. Reading theirs sharpened mine: they had found that one of the two page-republishing routes serves a stale cached copy instead of re-rendering. I had found that the *other* route doesn't publish component JavaScript. So I filed it as one bug: two routes, each carrying half of what a component needs, so neither can fully publish a change.

The loop came back in about nine minutes and refuted it. I checked its reasoning against the code myself before accepting it, and it is right.

My mistake was a specific and slightly embarrassing one. I had searched the file for a keyword, seen two matches, and concluded from the matches alone that the bulk route re-renders page sections from the component template. It doesn't. When I actually opened the function the loop pointed me at, it reads the same cached copy the other route does. The two matches I'd relied on turned out to be loading the page's `<head>` block and a contact-info block — neither of them anything to do with page sections. I asserted a structural claim about the whole fleet on the strength of two search hits without opening the function.

What survives is the narrower half, and it happens to be the half that was already written down in a comment in the source: only one of the two routes publishes component JavaScript. That part is real.

The corrected picture is simpler than mine and rather worse. Both routes serve a cache. The real question — which the loop put its finger on and which I had framed past — is whether anything ever refreshes that cache. The other thread's bug already answers it: no, because the flag that is supposed to mark a page as needing regeneration is read by nothing at all. So theirs is the substantive bug, and mine was a mis-framing of it rather than an extension of it.

I've not re-run the loop. The one real asymmetry left is already documented in the code itself, so there's nothing to *diagnose* — what remains is a decision about what to do, which isn't the loop's job.

Two things worth saying plainly. First, the advice I gave you earlier about how to build — prefer per-page data over shared component JavaScript — is unaffected. I checked rather than assumed: it rested on three risks, and none of them depended on the part that was refuted. Second, the standing guidance says that for a bug you can see, you'll out-diagnose the loop yourself for free. I had full context here and still got it wrong, and the loop caught it in nine minutes for a few credits. I wouldn't rewrite the rule on one data point, but it's worth knowing that the loop is useful as a check on a confident claim, not only as a way to crack hard bugs.
