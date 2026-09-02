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

---

## 2026-08-31 — picked this thread back up, and built the first of the four things the paid build asked for

This lane had been quiet since 23 July. Two things had happened to it while it was quiet, and
both were worth knowing before doing anything.

The first is that other people finished its next job. Back in July this lane wrote a handoff
saying "start by building a contrast check, then a layout check" — and both of those exist now,
built by other threads for their own reasons. So the obvious next step was already done, and
starting it would have been pure waste.

The second is that the first paid customer build landed on 31 August, you reviewed it the same
evening, and the session that reviewed it wrote four defects into this lane's inbox. All four
are the shape this lane exists for: nothing was broken, every page passed every check, and the
site was still wrong. The clearest one was the home page. It was headed "Latest from the ring —
fresh news, previews and results", and underneath it listed four explanation pages about the
site's own tools. Not one article. Every piece was individually fine. The promise was broken.

I have built a check for that one, and it is running.

The interesting part is that the check I was asked to build would not have worked, and proving
that took about ten minutes. The suggestion was to look at how each page is *labelled* in our
database — flag a "latest news" list that contains pages labelled guide or tool. I measured it
first. Across the fleet, pages that are obviously guides are labelled "blog post" 246 times
across 30 sites, and labelled "guide" only 72 times across 9. It is simply how the system files
them. So on boxingonline — the exact page you complained about — that check would have found
nothing at all, and reported the site clean. It would have been a check that could only ever say
"fine".

What works instead is reading what each item says about *itself*: the address it lives at, and
the words in its own title. "Understanding Boxing Quiz | Guide", sitting at /guides/, is telling
you what it is regardless of how it is filed. That version found the boxingonline case, and it
also found a second one nobody had noticed — leopardessconsulting.co.uk's blog page, headed
"Latest Articles", where seven of the thirteen items are tool guides rather than articles. One
site's fault is an accident; two is a pattern, and now there is something watching for the third.

Three things I got wrong on the way, because they are the useful part.

I first read the section's *subtitle* as part of the promise. That produced eight findings, and
when I opened them, four were nonsense: the subtitle had simply mentioned a tool in passing —
"if you want a shorter list, the Garden Jobs Finder" — and for most listings on the fleet the
text I was reading was actually the first item's own description. I had built something that
reads the items and calls the result the promise. The heading alone carries it now.

I then made the same mistake I had just caught, one step along, by treating any /blog/ address
as "an article" — on an estate that keeps its guides in /blog/. That flagged two perfectly
correct pages. So the check now states plainly what it cannot do: it can catch guides showing up
under a news heading, and it cannot catch the reverse, because nothing on the page tells it.
Better a stated blind spot than a confident wrong answer.

And my own test suite passed while a whole function was missing, because no test called it. The
cluster found that one for me, by crashing.

One small thing worth mentioning because it is funny and slightly instructive: I picked the
boxingonline home page as the check's reference case — the known-broken page it must always
find. While I was building, another session repaired it. So the check's own health test started
reporting failure, on a working check, because the world had been fixed underneath it. Reference
cases now come from frozen copies, not live pages.

The check runs every morning at 07:25 and writes a line into the database whether it finds
anything or not, so silence can never be mistaken for health. Today it finds one thing:
leopardess. I have not touched that site — either the guides come out of the list or the heading
changes, and that is a decision for whoever owns it.

The other three asks from the paid build are still open: an index page with nothing to index
should be treated as a failed journey rather than a healthy page; tools should have to declare
what data *we* bring versus what we make the reader type in; and the content quality auditor we
already own and pay for never ran on the paid site at all.

---

## 2026-09-02 — two more checks from your second look at boxingonline, and one I have refused to build

The session that reviewed boxingonline sent me three more problems from your second round. All
three are the same shape as before: nothing broken, everything validated, the site still wrong.
I checked all three against the database myself before building anything, because a colleague's
report is a report, not a measurement. Two of the three stood up exactly as described, and one
turned out to be worse than described.

**Two "News" links in the top menu, going to two different pages.** Confirmed. There are now two
sections both labelled News — the original one and a second created since Sunday. A visitor has
no way to know which is which. Worth saying clearly, because it is the interesting part: the
menu machinery is not broken. It was proved working end to end the day before, and it put the
right five items in the right order. It simply had no opinion about a sixth item wearing one of
their names. A mechanism can be perfectly correct while the experience it produces is not, and
that gap is the whole reason this lane exists.

**The fight calendar page has no calendar.** Confirmed, and worse than reported: the page has no
calendar component at all. What it has is a heading and two thousand words explaining how we
maintain the calendar — how entries get added, what each listing gives you, that we correct dates
rather than leave stale ones up — sitting above nothing. And it never says it is empty, so a
reader cannot tell whether there are no fights this month or the page is broken. This was the
core thing the customer paid for.

Both now have a check running every morning, and both checks found their case. Along the way the
first version of the tool check was wrong in an instructive way: I counted whether a page had a
component *named* like a tool, and got 74 broken pages across the fleet. Opening them, the tools
were simply named differently — a naming habit I had mistaken for a fact. Judging what the page
actually serves to a reader instead, the real number is one. I would rather report one true thing
than seventy-four confident ones.

**The third problem I have refused to automate, and I want to be plain about why.** An article
titled "Last night's result: underdog shocks the champion" contains no result — it is a general
essay about why underdogs win, citing fights from 1990 and 2019. The damning detail is that our
own news page on the same site carries the actual story, dated 31 August. We had it. The article
did not use it. That is real, and it needs a reader's judgement, not a pattern. Anything I could
write today would flag every well-written general article and stay quiet on a specific-sounding
essay — a check that looks like the others and quietly does not work. So it is written down as
refused, with the reason, rather than half-built. The copy lane has the writing half of it.

One thing that keeps happening and is worth knowing: **your team is fixing this site faster than
I can build checks for it.** The duplicate News link was repaired five minutes before my first
run, so my new check reported nothing on the very case it was built for. Sunday's check had the
same experience with a different page. Both times the check was right and the world had moved.
Both now carry frozen copies of the original broken page as their reference case, so they keep
proving themselves after the real page is fixed. That is not a complaint — it is the pre-delivery
cut-line you asked for, working.

And one of my open questions is now closed by your ruling. I had proposed, and could only
propose, that a tool ought to declare what data *we* bring versus what we make the reader type
in. Your words on the comparator and the calendar settle it. The check I built today is the
mechanical half; the other half — refusing to choose a tool at planning time when we would have
no data to put in it — sits with the planner and is still to do.

---

**2026-09-02, later — you said the content quality auditor should be in the new build path. I
went to do that, and stopped, because it would not have worked.**

The instruction is right. What I found is that moving this checker earlier would not, on its own,
have caught anything you complained about on the boxing site — and I would rather say so now than
show you a green tick later.

Here is the problem in one sentence. The auditor only ever looks at four pages, chosen by name:
the home page, "about", "services" and "contact". Everything else on the site is invisible to it.

Boxing Online has twenty-two pages. The auditor can see three of them (there is no "services"
page). The guide pages you called padding, the articles index that wrote a manifesto instead of
listing articles, and the fighter comparison tool that made the reader type everything in — all
ten of those pages sit outside the four names. It was never going to mention them. Across the
whole estate it is worse, not better: thirty-six sites, 1,196 pages, and the auditor looks at 92
of them. Under eight per cent.

Three smaller things compound it. It reads only the first thousand characters of each page it
does look at, which on a typical home page is about four per cent of the text. Much of that
thousand is not text at all but styling instructions — on the boxing home page the styling starts
at the very first character, so essentially the entire sample the reviewer received was
stylesheet rather than words. And the order it reads a page's pieces in is not fixed, so the same
unchanged page gave a different sample this afternoon than it did at lunchtime.

To be fair to it: what it does produce is decent. On the boxing site it flagged the home page for
burying four tools inside one paragraph with nothing to click, and the about page for claiming
accuracy while offering no evidence and no named person behind it. Those are fair criticisms and
you would probably agree with them. It is a real reviewer with its eyes almost shut.

There is a second problem waiting behind the first, and it is the one worth your attention. Even
the findings it produces mostly go nowhere. The system has been quietly filing notes to itself
saying "no handler for this kind of finding" — thirty of them for calls-to-action alone, going
back a fortnight. So the checker notices something, writes it down, and nothing on the other end
is listening.

So the job is bigger than plugging it in, and I think that is the honest read of your
instruction rather than a departure from it. First give it eyes: let it see every page, read more
of each one, and stop feeding it stylesheets. Then give it the questions you actually care about
— does a list contain the kind of thing its heading promised, does a tool bring any data of its
own, does a guide earn being more prominent than the thing it explains. Then put it in the build
path, which is the easy part and the part I can do quickly.

**The one thing I need from you before I go further.** When this runs during a build and finds
something, should it just record the finding for whoever approves the site, or should it be
allowed to send pages back to be rewritten automatically? Automatic rewriting is the thing you
switched off on 25 August because it was causing bad and unexpected renders. Before delivery
there is a decent argument for switching it back on, since nobody has seen the site yet. But it
would be running immediately after the writer that just produced the page, with nothing in
between, and that is close to the arrangement that misbehaved before. My recommendation is to
record only, and to make the findings part of what you see when you approve the site — but it is
your call and it is the difference between a checker that reports and a checker that acts.

---

**2026-09-02, evening — the auditor now has its eyes open, and it is live.**

Earlier today I told you the content quality auditor was looking at four pages out of twenty-two,
and that plugging it into the build would not have caught anything you complained about. That is
fixed and running.

The change went through the reviewer council. It came back needing revision first, and the
objection was a good one — it found a real bug in my fix. My code for stripping stylesheets out of
a page had a subtle fault: on a page with two separate style blocks it would delete everything
between them, including the actual writing. I had not noticed because the page I tested on only
had one block. I measured how often it would bite — seven pages across the whole estate, but on
every one of those seven it destroyed text, up to nine thousand characters on the worst. It would
have done that silently, which is the same class of fault the change exists to prevent. I fixed it
by copying an approach we had already proven elsewhere, and I added a check that fails if the bad
version ever comes back. Second time round it was approved.

It is applied and I have watched it working. Every audit the system has run since sees fourteen to
eighteen pages instead of three. Boxing Online itself now resolves to eighteen pages across eight
kinds. One caution I want to be straight about: the sweep works through about fifty sites at one
per quarter hour, so it has not come back round to Boxing Online yet. The improvement is proven
across the estate but not yet on the site that prompted it, and I would rather say that than let a
good number stand in for the one you actually asked about.

**Two other threads asked me for things today, and one of them found a real hole in my own work.**

The designblog lane reported four pages that list nothing — a glossary with no terms, a directory
with no studios — each carrying prose about what it intends to contain. My detectors reported
nothing. My first thought was scheduling: they run early morning and the sites were built at
lunchtime. That was true and it was not the real answer. I ran them by hand and they still found
nothing, because the rule asks "is this index listing the wrong kind of thing?" and that question
quietly assumes it is listing something. **An index that lists nothing is invisible to it.** That
is now the second time this exact gap has shown up, which is what makes it worth building rather
than noting, and I have taken it on.

While I was there I found a smaller fault in my own tool: when you point it at one site, its
self-check reports a failure that actually means "not applicable". That matters more than it
sounds, because that line is the one that tells you whether a clean result can be trusted — and if
it cries wolf on every run, people stop reading it. I told the other thread to treat that
particular clean result as untested rather than good.

The vetcomparison lane asked for an experience read on their homepage. Their promise ledger came
back clean, with one finding: a tool page that was planned and never built, which returns a 404
but which nothing currently links to. It is half-wired rather than broken — it has a navigation
label set while sitting outside the navigation, so the moment anyone switches that on it becomes a
dead link. They tell me it was a deliberate hold from July and they are putting the decision to
you today.

**Your design directive about sites not all sharing the same nav-and-big-hero shape** was passed
to me but it is not my thread's work, so I did not quietly adopt it. It has been routed to the
threads that own composition — components, theme kits, the site design planner, the vigilant
designer and the editorial design uplift lane — and they have all acknowledged it.

**What is left on my side.** Build the empty-index rule. Fix the self-check bug above. Then the
last step of the original job — putting the auditor into the build sequence — which I am
deliberately holding until I have watched one real audit run on a freshly built site, because
wiring in a checker before you have seen it work is how we got here.
