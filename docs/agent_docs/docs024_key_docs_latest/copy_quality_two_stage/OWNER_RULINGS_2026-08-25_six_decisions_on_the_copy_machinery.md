# OWNER RULINGS 2026-08-25 — the six decisions, ruled in chat, verbatim in substance

Context: the replay experiment (`AUDIT_prompts/EXPERIMENT_2026-08-25_about_section_replay.md`) and
the decision list put to the owner the same evening. His words quoted; my reading after each.

1. **The planner premise** — *"yes please research this."* → Research where `build-site-planner`
   mints self-limiting page premises ("About Home Garden — Editorial Approach and What We Will Not
   Do") and fix at the plan layer. AUTHORISED as the priority.

2. **The writer's July substitutes** — *"the other option is to not say it at all. I approve of
   deleting the substitutes."* → Delete the three "say this instead" clauses from
   `page-content-writer`'s template (the STRICT-RULE remedy sentence, the Operating-history
   "method = sourcing" sentence, rule 19's "we cannot tell you X" substitute). **Every ban stays.
   Nothing is added in their place** — his "not say it at all", and the experiment's arm D earned
   nothing over deletion.

3. **Tool-fallibility mandate** — *"keep the tool fallibility mandate."* → The clause "Where the
   copy describes an interactive tool, it must say the tool can give a wrong answer" STAYS,
   fleet-wide. (It survives the deletion in 2 as its own sentence.)

4. **Uncertainty-acknowledgement** — *"for finance we acknowledge but for planting grass seed we
   don't need to acknowledge it, the onus is on telling what we do know and what the research
   tells us."* → SPLIT BY STAKES: finance keeps acknowledge-uncertainty as a trust device;
   low-stakes editorial does not — the emphasis is what we DO know and what the research says.
   Affects the per-site briefs' trust lines (homegarden's *"Explicit honesty about uncertainty is
   itself a trust signal"* is the wrong device for that site under this ruling).
   **Plus a standing mission requirement:** *"Somewhere the mission for every site we build must
   be to make it the best in class - please check that is still there. For that to happen we can
   research the latest findings, the latest products, the most trusted product reviews and
   anything else that will make us best in class."* → CHECK ordered: does a best-in-class mission
   exist in the seeds/specs/prompts, and is research wired to serve it?

5. **House voice** — *"It can be rewritten - what did I say?"* → Rewrite APPROVED (form-only, no
   rule's meaning changes, written in its own recommended shape); and he asked to be shown what
   the approved v2 text says — answered in chat and the text lives at
   `agent_default_configs.config->>'text'` WHERE `config_name='voice_style_block'` (CQ-022).

6. **Empty social-proof slots** — *"Stop planning them on sites that can't fill them."* → A
   PLANNER-side change: testimonial/case-study slots are not planned where no real content
   exists. Folds into ruling 1's planner work. (The writer's rules 16–17 filler instructions
   become dead letters once the slots stop appearing; their removal can follow the planner fix.)

**Standing constraints unchanged:** bans stay (no-overclaiming, no-history-claims, the "honest"
word ban); D2 (stage 2 proposes, human approves); D3 (`rather than` mild in copy); prompt changes
ship as migrations through the council gate.

---

## Execution record, same day

**Rulings 2, 3, 5, 6 SHIPPED 2026-08-25** — migrations `627` (writer substitutes deleted; every ban
and the tool-fallibility mandate asserted present post-edit), `628` (house voice form-only rewrite —
scanner: 17 negation demonstrations → **0**, live row verified at 5,863 chars), `629` (planner: the
worked example no longer demonstrates a testimonials slot — it literally contained
`{"name": "testimonials"}` — plus the explicit rule). All applied with `migration_backups` rows and
per-file ROLLBACKs, recorded in `schema_migrations`, verified at the live rows independently of the
migrations' own NOTICEs. Council: `Council-Submitted: 6a0f8b99-f4c5-4e47-981f-cd59043b462d`.
Canary: the next greenfield build's about page.

**Ruling 1 research (first cut, `[MEASURED 2026-08-25]`):** the planner template itself never says
"will not do" (0 hits) — the premise chain is: the MISSION SEED is prohibition-rich (homegarden's:
*"Say what a thing is for and, just as clearly, what it is not for … We have not tested anything
and should never imply that we have"*) → the planner amplifies mission prohibitions into page
PREMISES. It does this despite already carrying a guard (*"a methodology or 'how we assess' page
may be planned only as what the site IS and does for a reader, never as a description of
practice"*) — a prohibition with no demonstrated good example, the displacement shape again.
Fleet blast radius: ~6 of 23 about pages carry self-limiting titles (lendzy *"Independent, Not a
Lender, Never Will Be"*; loancalculator *"No Advice, No Data Stored"*; homegarden the sharpest).
**Fix candidate, NOT shipped (awaiting the owner):** give the planner one demonstrated GOOD
about-page premise (his register: *"We're hoping you can get a lot of useful tips from this
site…"*) plus the rule that mission constraints become internal rules, never page titles or
premises.

**Ruling 4 check (best-in-class):** the mission IS still there — the `domain-research-classifier`
prompt carries a "Build standard" block: *"Aim for best-in-class quality in this site's field. The
bar is not 'competent template' but 'stands comparison with the strongest sites in this
vertical'…"*. But it lives ONLY there: **0 of 51 sites' current specs contain the phrase**, the
planner and writer never see it, and the submit script records *"Best-in-class design research (a
design-research sub-agent + best-in-class mission default) is a separate TODO and is not part of
this trigger"* — the research-serves-best-in-class wiring was never built. The stakes-split of the
uncertainty device (finance acknowledges; grass seed does not) needs a category model over the
briefs — scoped as the next piece, briefs untouched today.

---

## Second wave, 2026-08-26

**Ruling 1 SHIPPED** — migration `630`, the TESTED second draft (draft 1 held 1 of 2 planner
replays — its failure, "Practical UK Guidance, No Products to Sell", was the rule's own named
error; draft 2's hard format clause held 3 of 3: "About | Home Garden"). Applied, backed up,
rollback on file, recorded, verified at the live row alongside 629's rule.
`Council-Submitted: 5f084feb-7c2a-4412-9363-29d5108eed5b`.

**Ruling 4 PLANNED** — `PLAN_2026-08-25_best_in_class_propagation.md`: carrier-row + injection
(the Go question answered there: text in Go would not break anything but optimises for
immutability when the mission needs reach and cheap tuning — mechanism in Go, words in one DB row,
the proven CQ-022 shape), per-site `strategy.benchmark` at birth, the research inventory → refresh
→ routing phases, and the finance-vs-editorial stakes split over the briefs.

**Wave-1 council verdict: APPROVED** (`6a0f8b99`, round 1, 2 advisories, none high), advisories
checked not banked: the seats read `agent_default_configs` as "no such table" — it exists (queried
throughout; `voicestyle.go:35` is the code's own query against it) and the 628 verification ran at
the concrete path, though the submission's edit-2 symbol WAS vaguer than edits 1/3 (noted for
future submissions); the guardian's two-active-rows trap was CHECKED — all three wave-1 targets
single-row — and found REAL elsewhere (four types carry 2 active rows) → filed to `LANDMINES.md`
with the in-file `RAISE count<>1` check prescribed, verifier armed; the fleet-wide-immediacy
concern is answered by the offline replay-before-apply pattern both waves used plus one-file
rollbacks.

---

## 2026-08-26 evening — two further instructions, on his read of finetuning.uk

**7. THE TRUNCATION TRIAL** — *"as a trial, whenever we want to write the second half of one of
these sentences, we should just stop before the negative (or the 'not' or the 'instead of') and
leave that part of the comparison out all together. We don't need to sound competitive like this.
There is no hidden competition. We offer what we offer straight up."*
**Context measured before implementing:** his quoted specimens ("we don't sell one", "not tied",
"nothing runs", "The right model for the job, not the only one…", "a mechanism for catching
mistakes early, not a pr…") ALL sit in components **not yet rebuilt** — the canary queue had
pushed only 2 of 9 pages; index/approach/faq carry the same copy he rejected on 08-25. **BUT the
two rebuilt pages independently FAIL P2b** (3 and 5 constructions vs the pre-registered ≤2/page),
and the character is diagnostic: the **tool-fallibility mandate (his ruling 3) satisfied in the
negation shape** — "a starting point for a conversation, not a verdict" ×4, "a guide rather than
a guarantee" — an affirmative instruction executed in the model's preferred competitive form on
pages with a near-zero demonstration stack. **EXECUTED same evening:** the gate's repair prompt
(`negationRepairPrompt`) is now truncation-first — end the sentence before the comparison, keep
the first half, leave the alternative out, shorter is the point; safety instructions kept; the
no-demonstration guard test passes on the new text. Go — **inert until the next roll**; tonight's
remaining rebuilds run the old repair, covered by his forward-only standing instruction. Council
`Council-Submitted: 82b800e1-4af7-44af-a41d-504386584498`. **Open sub-question flagged, not
bundled:** whether the trial's "whenever" also repeals D3's mild-forgiveness (≤2/page) — his call.

**8. HERO LENGTH** — *"on mortgagecalculator and others, the hero text has become way too long
and boilerplate. It should be shorter but if it is to be long it should be composed better."*
Census `[MEASURED 2026-08-26]`: homepage hero visible text — cv1.co.uk **16,088** chars (an
outlier pathology of its own), cookly 2,723, **mortgagecalculator 1,380**, robot-hands 1,311,
fundamentallyai 1,138; median fleet ~900. Routed to the audit's next verdicts (hero components'
`llm_guidance` + the brief sweep) — the fix shape is a hero length/composition rule at the
guidance layer, not yet designed. mortgagecalculator also carries the v2 voice FOSSIL
(CONTRIB 08-26 in their lane), a suspect for "boilerplate".

---

## 2026-08-26 night — the roll landed; the trial is LIVE and APPROVED; two questions now carry numbers

**Ruling 7's trial is LIVE at the artefact**: pod on v1.0.1345 probed — both added literals
present (×1), the REPLACED instruction absent (×0, the removed-string control), and the tone-route
bound's literal present too (`items_copy_edit_bound_unevaluated` ×1 — CQ-030's bound is now live
as well). Council **APPROVED, all reviewers, round 1** (`82b800e1`, 19:10Z). Fleet still mixed
(19× 1344 / 106× 1345), converging.

**The canary's round-2 scoring sharpens the two open questions into decisions with numbers**
(canary doc, round 2): rebuilt pages still ship ~10 constructions per multi-section page because
(a) **D3's mild-forgiveness is per SECTION** (the 305 landmine), so `rather than` ≤2 per section
× 6–7 sections ≈ 6/page never even reaches the truncation repair; and (b) **"instead of" and
"not just" are outside the gate's five shapes** — the gate cannot see them, and the owner NAMED
"instead of" in the trial. So:
- **Q-A (owner):** does the trial's "whenever" repeal D3's mild-forgiveness — repair every
  `rather than` (truncation makes repairs cheap and safe now), or keep an allowance, and if so
  per PAGE rather than per section?
- **Q-B (ordered, execution pending):** add `instead_of` (named in the ruling) — and `not_just`?
  — to `ScanDefineByNegation`'s shapes; classification (mild vs always-repair) follows Q-A.

---

## 2026-08-31 — four further instructions (rulings 9–12)

**9. THE WORD "PLAINLY" JOINS THE BAN** — his specimen from finetuning.uk: *"Real projects,
described plainly"* — *"we just don't need to say that and no one says 'plainly'."* Same class as
the July "honest" ban: never LABEL the register; delete the label word. Fix shape: extend the
writer's rule-19-style word ban (fleet template, migration).

**10. NO INTERNAL-DIRECTIVE EXPLANATIONS IN COPY** — his specimen: *"… that is what we have
written, because a number we cannot stand behind is not one we will print."* His words: *"written
too dense and too competitively and it's very much about our internal directives rather than what
it should be, which is focusing on describing what we do in terms that the user will understand
might be a benefit to them."* This is P4's unmoved candour/self-narration class, now with the fix
DIRECTION ruled: benefit-led description, not self-justification. Remedied via ruling 12's
approach, not another prohibition.

**11. MODEL TRIALS** — *"Let's try different models until we find the best. I think Claude Fable
will probably be too expensive but let's try it as it might give us a benchmark for the other
models, then try Grok and Gemini next."* Directly motivated by the canary's P2a result (the
carrier is the MODEL PRIOR — so change the model). Method: offline replays of a constant
post-fix rendered prompt across models, battery + read, pre-registered; Fable = the benchmark.

**12. BENEFIT-LED HERO COPY, WORKED OUT WITH THE OFFER/BENEFIT ANALYSIS THREAD** — *"correspond
with the offer analysis and benefit analysis thread to try and iron out what sort of approach we
can use with these hero titles and copy to explain what we do in terms that clients can see how
it might work for them and why they might think these things that we do are useful to them. We
mustn't presume to know what they want or what things actually are useful to them because that
presumptive approach doesn't work well. It's all subtle and if each piece of copy requires
discussion between agents then so be it."* Note the last sentence: PER-COPY multi-agent
discussion is explicitly licensed. Non-presumption is a constraint of its own (the "Most
visitors arrive with one specific question" tell was his earlier example of presumption).

**Also expected:** correspondence about farmerinsurance.uk copy ("the old AI type of content").
**Still open from before: Q-A and Q-B** — not yet answered.

---

## 2026-08-31 (later) — the co-design's first round, and DECISION C for the owner

The offer-analysis thread answered with measurements
(`CONTRIB_2026-08-31_from_the_offer_analysis_lane_the_benefit_set_exists_at_scale_and_no_writer_reads_it.md`):
the benefit set exists at scale — **32 sites, 187 ranked `lead_with` points with provenance —
and NO writer-side agent reads any of it**; finetuning's own `avoid_leading_with` already forbids
the exact class the owner rejected (*"an unenforced judgement, not a missing one"*). They
correctly REJECTED our "heroes draw only from the set" as written: **18 of 187 points carry the
banned register** (feeding them as data would inject it through the mandated-phrase channel — the
strongest transfer mechanism we have measured), and only 22 of 187 open conditionally. So: **the
set supplies WHAT and in what order; this lane governs HOW; and the set needs ONE register pass
BEFORE it becomes a source.** Their three asks, answered: (1) the banned register now exists AS
DATA — `AUDIT_prompts/BANNED_REGISTER_v1.json`, versioned, provenance per entry, structured-input-only
usage rule; (2) rank PINNED during the register pass — a form fix must not change substance, and
any fix that would is not a form fix (it goes back to their analyser for re-judgement); (3) the
non-presumption constraint binds the SERVED `point` only in v1 — the unserved `why` fields join
the audit's census as a known demonstration reservoir (reachability measured before any wider
scope).

**DECISION C (owner):** authorise wiring `offer_ordering` into `page-content-writer` (config on
the shared writer, council scope) — after the register pass over the 187 points — with
finetuning.uk as the first case. This is the concrete mechanism behind ruling 12; both lanes are
agreed on the shape and neither will wire it without his word.

> **DECISION C AMENDED 2026-08-31 (later), on the offer lane's own correction, INDEPENDENTLY
> RE-RUN AND CONFIRMED by this lane** (51 of 187 points dirty — 27%, not their first-reported
> ~10%; 23 of 32 sites; shapes dominate: `x_not_y` 28, `rather_than` 13; **7 of 31 rank-1 hero
> candidates dirty**): the set is a DEMONSTRATION RESERVOIR of the exact comparison shape the
> canary proved the writer produces unprompted — feeding it unwashed through the mandated-phrase
> channel would re-teach the shape through the most effective transfer mechanism we have
> measured. TWO changes to C's packaging: (1) the register pass over the 51 is a PRECONDITION for
> any dirty site's wiring, but **9 of 32 sites are clean TODAY**, so a first wiring can proceed
> on a clean site without waiting; (2) **"finetuning first" is WITHDRAWN — it is the WORST site
> in the corpus (6 dirty points, its rank-1 included)**: the first outing should be a clean site
> chosen deliberately, with finetuning as the best demonstration OF THE PASS, not of the set.
> Their figure is dated 2026-08-31 and describes the EXISTING corpus — re-measure after any
> writer-model change.

---

## 2026-08-31 (evening) — rulings 13–14, from his critique of the Fable benchmark output

**13. DENSITY IS A FAULT CLASS OF ITS OWN — models compress; copy must EXPAND.** His specimen:
*"a small first project lets you judge the work on results"* — *"more of a riddle. I have to work
out what results we're talking about and what work we're judging… it collapses so many thoughts
into just a few words. It is the way models work so we need to put in effort to expand the words
more to get a more readable and understandable, less dense text every time."* Consequence for the
model trials: **the Fable verdict is DOWNGRADED to "register pass, density fail"** — the trials
continue per ruling 11, with density a first-class criterion of the read from now on. (Instrument
note, recorded against ourselves: this lane's read scored the output "passes decisively" — the
read was against the criteria as they stood, and his ear extended them. The same lesson as the
checklist: our read is not his ear either.)

**14. RELEVANCE IS SUBTLE AND ORDERING FOLLOWS THE USER'S OWN QUESTIONS.** His specimen: *"No
vendor pays us, so the choice is made on fit alone"* — *"unnecessarily direct and strong. It
might be the fact but the relevance of the facts are more subtle"*: the real benefit (results not
influenced by whoever might pay us) needs UNFOLDING, and it is probably far down the user's list —
their first doubt is *"what this tool as a whole is going to get them and how much work it's
going to be"*, then practicality, then data independence. The offer/benefit thread is asked to
help *"see what users are after and what hidden questions they may be asking in their minds."*
**Clarification implicit in his own message:** modelling the user's question hierarchy to ORDER
content is licensed (he does it himself here); ASSERTING the hypothesis in copy is the presumption
ruling 12 bans. Ordering: hypothesis-driven. Copy: offered, never presumed.
