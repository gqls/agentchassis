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
