# PLAN 2026-08-20 — a mechanical gate for define-by-negation, at the writer seam

**Lane:** `bugfix_305_negation_gate`. **Bug:** `bugs_open/305` (owned and diagnosed by
`copy_quality_two_stage`; this lane builds the platform half and contributes back — see
`docs/agent_docs/docs024_key_docs_latest/copy_quality_two_stage/CONTRIB_2026-08-20_from_the_305_lane_the_writer_side_gate_is_being_built_and_here_is_what_it_will_and_will_not_guarantee.md`).

## What we are trying to do

The owner read three directory pages and said the copy *"looks like it didn't go through the
framework"*, quoting two sentences built on the same mannerism: saying what a thing **is not** in
order to say what it is. He gave two instructions — fix those pages, and *"ensure that that sort of
copy never leaves this framework again"*. This lane builds the second half, mechanically, and hands
the first half to the site's owning lane because it is their site and their brief.

## The decision that shapes everything else

**We are not adding another rule to a prompt.** That has been tried twice and measured to fail:
migration `docs/agent_docs/sql_for_agents/228_writers_default_to_humanised_voice.sql` carries the rule
in its own header and says it *"did not work"*; the v2 house voice
(`agent_default_configs.voice_style_block`) says *"Say what a thing IS rather than what it is not"*
today, while the writer produced the construction for the complained-of site again on 2026-08-19. The
`fleet_copy_quality` lane's conclusion is the reason:

> *"A rule can only name a form. What goes wrong is an instinct."* … *"Prescriptions become tics."* …
> *"The example is the instruction; the rule is commentary."*
> (`docs/agent_docs/docs024_key_docs_latest/fleet_copy_quality/SUMMARY_2026-08-08_why_this_is_hard_and_what_we_have_learned_about_rules.md`)

So the fix is a **mechanical check on the output at the seam**, with a bounded LLM repair of the
specific sentences that trip it, and a measurement of what the repair does instead.

## Where the seam is, and why there

`page-content-writer`'s `process_sections_loop` runs `generate_content` (execute_llm_prompt) →
`render_section` (render_component). **Nothing sits between them.** Neither does anything sit at
save, at compile, or on the writer's output anywhere else: the only wired detector
(`platform/orchestration/datahelpers/voicetells.go`) runs **post-deploy**, only for the 9 of 43 sites
with `voice_gate.enabled`, and its `strawmanCommaRe` matches *"not X, but Y"* only — 1.5% of sections,
and neither of the owner's sentences.

Two pieces, deliberately split:

- **The count goes where it cannot be forgotten.** `render_component` annotates every LLM-authored
  section it renders (`copy_gate_findings`), default ON, no behaviour change. Every writer, every
  site, whether or not the fixer is wired. Precedent and reasoning:
  `save_page_meta_description_action.go:44-48` — *"a gate a workflow author can forget to wire is a
  comment"*.
- **The repair goes in its own action.** `rewrite_negations`, with its own `ActionInputSpec`, wired by
  migration into the writer's loop. **Not** a new key on `execute_llm_prompt`: that action has 66
  carriers, no input spec (so the RFC_022 optional-key budget cannot see it), and a prose-quality
  rewrite is writer behaviour, not transport behaviour.

## The rules the gate enforces

| rule | value | why this value |
|---|---|---|
| page budget | first **2** non-exempt hits pass | the house voice's own standard: *"earned once or twice per page at most"* |
| headline-class fields | **any** hit is rewritten | the owner's complaint was a headline and a subheadline; 14% of sections carry one |
| exemption | sentence-level match against **brief-supplied** fields, plus a regulatory allowlist | a phrase the brief hands the writer is the brief's decision, and *"a site's own voice specification outranks these rules"* |
| repair shape | one call, **sentence replacements**, spliced in Go | preserves the key set and types by construction |
| acceptance | per replacement, and only if it adds no hit in the five shapes **or** the neighbour set, and preserves digits/URLs/proper nouns | otherwise the repair adopts displacement |
| failure | never fails the step | style is softer than truth (the `voice_tells` severity precedent) |

**The per-page budget cannot live in a section.** Six sections at one hit each is six on the page and
every one passes a per-section threshold. Loop iterations share one workflow state
(`platform/orchestration/actions/loop_actions.go:209-239`), so the running count rides in
`CollectedData` under `__copy_gate.page_hits`. ⚠ **That is the one load-bearing thing in this design I
have not yet proven at runtime** — it must be shown by a live canary, not assumed.

## Three sub-designs that were refuted before any code was written

Recorded here rather than in a retrospective, because two of them are traps for anyone who builds in
this area and the third is a cost mistake.

1. **REFUTED — "re-ask for the whole section and keep whichever attempt scores lower."** Measured in
   the same 1,503-call corpus: `instead of` 5.9%, `isn't just/a` 6.4%, `more than (just)` 10.8%, em
   dash 0.5%. A rewrite to *"X instead of Y"* scores **zero** on the five shapes and wins the
   comparison while being the same instinct. This is `copy_quality_two_stage`'s *"a prohibition
   displaces a problem rather than solving it"* arriving from the measurement side. → per-replacement
   acceptance with a neighbour-set check, and **rejections logged with a reason**.
2. **REFUTED — "exempt a hit whose phrase appears verbatim in the rendered prompt."** The literal
   string `rather than` is in every rendered writer prompt (the house voice uses it six times; STRICT
   RULE 19 uses it), so the whole 43% arm would have been silently exempt. → sentence-level
   exemption against brief-supplied fields only.
3. **REFUTED — "quote the house-voice rule in the repair prompt."** That rule's own text carries the
   construction and a worked example of it. "The example is the instruction" applies to the fixer's
   prompt as much as the writer's. → the repair prompt carries no rule text and no example of the
   banned shape, only the positive instruction.

Also costed and rejected: the full-regeneration repair is ~$0.072/call against ~$0.0135 for the patch
shape (~$200/month vs ~$40 at today's 215 sections/day), and it can truncate, drop a key or change a
type, which loses the section at `render_component`'s required-field refusal
(`v3_site_actions.go:2388-2400`).

## What this does NOT do

- **It does not clean the three pages the owner named.** `in days, not months` is supplied by that
  site's brief and is therefore exempt **by design**; the gate counts it and leaves it. Only a brief
  edit plus a rerender moves it, and that belongs to `site_ai_agent_orchestration`.
- **It does not promise the instinct never leaves.** It holds five named **forms**.
- **It does not replace `copy-editor` (stage 2).** A section-scoped gate structurally cannot see the
  same argument made five times on one page.
- **It does not touch `bugs_open/327`** (`site_spec_actions.go` computing `formatted` from the
  incoming partial). Different bug, different owner, and it sits on the brief edits two lanes are
  about to make.

## Phasing

1. **Phase 1 — the gate.** Scanner + `render_component` annotation + `rewrite_negations` action +
   `compile_page_sections` page count + migration (held until the image is live). Council-submitted.
2. **Phase 2 — the brief-side check, scheduled.** `cmd/brief-negation-check` and a daily CronJob on
   the `verifier-remit-check`/WFA-013 shape. `copy_quality_two_stage`'s `audit_writer_brief.py` is the
   specification; the Python stays as the human-run tool. **Owner directed this on 2026-08-20**, after
   that lane had parked the scheduling question as an owner/architecture call.
3. **Phase 3 — the pages.** Not ours. CONTRIBs sent to the three affected lanes with the artefact-level
   verification queries.

## Corrections to the originating brief

- The bug's §7 says to verify by re-running the `llm_call_log` rate. **That is wrong for this fix and
  is corrected here:** the gate's own first attempts are logged in that table, so the per-call rate
  can rise while the artefact improves. Verification is at the artefact
  (`page_components.content_data`), at the `__copy_gate` marker, and on a first-attempt/retry split.
- The bug's title still overstates its evidence (its own §3 says so). Nothing in this plan depends on
  the v2 before/after comparison, which was withdrawn as unanswerable at that sample size.
