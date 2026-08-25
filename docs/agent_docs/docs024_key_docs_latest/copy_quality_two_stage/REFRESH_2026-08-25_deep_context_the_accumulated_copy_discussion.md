# REFRESH 2026-08-25 — the accumulated copy discussion, assembled in one place

**Why this file exists.** Owner instruction of 2026-08-25, routed via the `loanzy_uk_example_site`
lane after his review of `homegarden.uk`: *"We have discussed copy at length. Please ask the copy
quality two stage to do a deep search and refresh their context on this before suggesting fixes."*
This is that refresh: four parallel sweeps over the lane's PLAN/NOTES/summaries (2,634 NOTES lines
read in full), the cross-lane CONTRIB series, the concept register + landmines + bugs, and a dated
census of the fleet's entire prompt surface. Every claim below carries its source. **No fixes are
proposed here — that is the point of the instruction's ordering.** What this licenses is at the end.

Canonical owner records this rests on:
`loanzy_uk_example_site/OWNER_REVIEW_2026-08-25_homegarden_and_what_it_says_about_every_site.md` ·
`copy_quality_two_stage/CONTRIB_2026-08-25_OWNER_ESCALATION_finetuning_pages_fail_the_would_a_person_say_this_test_after_a_maximal_seed.md`

---

## 1. The owner has issued one instruction five times, each time sharper

1. **08-12** — *"pin down why everything is negatively framed as a default and we should change that
   first."* Found: the negativity is **written down before any writer sees it** —
   `identity.key_differentiators[0]` drives the lead, and a differentiator written as a subtraction
   makes the writer lead with a loss (CONTRIB_2026-08-12_why_the_default_is_negative §1).
2. **08-15** — *"Sounds like AI with 'this not that'"* (webdesign.co.uk's "A workbench, not a sales
   pitch") — the tell got its name and its taxonomy (NOTES 08-15).
3. **08-18** — three directory pages: *"looks like it didn't go through the framework"* … *"ensure
   that that sort of copy never leaves this framework again"* — the gate era (CONTRIB_2026-08-18).
4. **08-24, finetuning.uk after a MAXIMAL seed** — *"It fails the 'would a person actually say this'
   test really badly, and it sounds so methodical like AI"* … *"the whole site could be rewritten in
   better language"* (CONTRIB_2026-08-25_OWNER_ESCALATION §1).
5. **08-25, homegarden.uk** — *"It's a whole category of not thinking about the user and what they
   are after"* … *"Stop talking about us in such a technical way"* … and the instruction this file
   answers, including: *"audit every prompt in the database and code and ask of it whether it is
   contributing to good readable copy or whether it is encouraging AI styles of writing (bad)"*
   (OWNER_REVIEW_2026-08-25 §4, §7).

The trajectory matters: each round, a remedy aimed at the previous complaint's SURFACE was met by
the same instinct in a new form. That is also the exact meta-lesson of his own style-prompt work
(§6): *"a phrase-level patch tends to resurface in new grammar next round."*

## 2. The single strongest finding: demonstrations govern, instructions do not

- *"The example is the instruction; the rule is commentary"* — proven by deleting a rule, leaving
  its worked examples, and watching behaviour not move (NOTES 08-12, inherited from
  fleet_copy_quality). Extended 08-24: **"an instruction is also an example"**
  (CONTRIB_2026-08-24_from_the_finetuning_lane §3).
- The writer's own rendered prompt **demonstrates the banned form 16× per call** (`rather than` ×8,
  `, not ` ×8 — structural prompt text, identical across three calls)
  (bugfix_305 CONTRIB_2026-08-20 §2). The per-site brief layer adds 7–8 more
  (`content_direction.formatted` on finetuning.uk).
- The controlled-ish test: de-demonstrating ONLY what the lane owned took `X, not Y` **3 → 0**
  (its demonstrations removed) while `rather than` went **6 → 8** (the fleet's 7 demonstrations left
  standing) (CONTRIB_2026-08-24 ADDENDUM). The classes track their demonstration counts.
- The negation rule **survived its own ban with worked examples**: the prohibition carried two
  negative exemplars, and the writer produced the move "in a third grammatical costume neither
  example covers" (NOTES 08-12 evening).
- Corollaries, all measured: **prohibitions displace** (banning loss-framing-first moved the loss to
  the sentence end — the owner's next objection, CONTRIB_2026-08-12 §6); **prescriptions become
  tics** ("Start with the fact" made a hundred sections open identically, NOTES 08-12); **a
  prompt that talks about X makes the output mention X** (016b §9 :12508 — the gate's own
  instruction manufactured its false positives, bugs_open/222).

**This is why the owner's audit question is exactly the right one, with one refinement the evidence
forces: the audit must judge what each prompt DEMONSTRATES — in its own prose, its examples, its
tone — not merely what it commands.** A prompt whose rules are impeccable and whose own sentences
are written in the register it forbids is a net teacher of that register.

## 3. Pattern lists have a measured ceiling, and the owner's ear is above it

- The maximal seed on finetuning.uk (positive exemplars + guards + gains-framed identity + fact-first
  rules + count-matched subjects + the live 305 gate) took tells **9 → 9 → 6 and floored**. The
  owner's rejected specimens ARE that floor (CONTRIB_2026-08-25_OWNER_ESCALATION §2).
- The lane's own checklist scored the owner-rejected "Who is actually running this" section
  **CLEAN**, and the session summary praised it hours before he rejected it: *"a pattern list is not
  the owner's ear"* (same file §3).
- What his ear catches that no current detector models: the **methodical scaffold** ("All three
  arrangements matter for the same reason:", enumerate-then-summarise), the **performed-candour /
  self-narrating honesty beat** ("That won't scale forever, and we'll say plainly if that changes"),
  the **presumptive reader move** ("Most visitors arrive with one specific question:"), word-weight
  in the falsely-humble direction ("say plainly" — *"people just don't say that"*), and the even
  essayistic cadence (§3 + OWNER_REVIEW §4).
- The 305 gate's guarantee is deliberately narrow and honest: *"what a gate can hold is a form"* —
  five named shapes, two forgiven per page, never in headlines; brief-supplied sentences exempt
  (bugfix_305 CONTRIB_2026-08-20). It is a floor-holder, not an acceptance test.

## 4. The homegarden review adds a LAYER, not another tell: content selection

His sharpest sentence is not about register: *"the premise of the page is wrong."* 14 of 17
about.html headings are about the site's own methodology — "Sourced and dated", "No fabricated
tests", "Timing stated plainly", "What this site will not do" (twice) (OWNER_REVIEW §4). His model
of right: *"We're hoping you can get a lot of useful tips from this site…"* — brief, about what we
do FOR THEM.

**Hypothesis, marked as such — the audit is its test:** the methodology self-description is the
platform's own integrity instructions leaking into copy as content. Suggestive evidence, none
sufficient alone: `evidence_base.writer_block` held **31 of 45 tells** on webdesign.uk — more than
the whole brief (NOTES 08-19); the finetuning "self-narrating honesty beat" and homegarden's
"Sourced and dated / No brands, no fabricated tests" headings read like prompt-integrity language
rendered as headings; a stand-in token from a writer_block instruction has already shipped verbatim
as public copy (bugs_open/387); and the writer, briefed with nothing reader-facing to say for an
about page, plausibly falls back on narrating its own constraints — the one subject every prompt
gives it in abundance. `[INFERRED — the audit's phase-2 question 2 tests this: find the prompts
whose instruction text matches the shipped self-descriptions.]`

Related, already-diagnosed content-selection mechanics (distinct from register): the section plan
carries no per-section subject (`pages.sections` parses as `[]string`), so every slot gets the same
brief and the writer restates the most concrete material — four builds proved it structural
(apis_uk NOTES 08-23/24, owner: *"it's the mechanism that would be good to fix here"*); and a
`content-listing` with nothing to list skips silently, leaving copy that describes the concept of
the missing content (bugs_open/384, OWNER_REVIEW §5).

## 5. What actually reaches the writer — the audit's reachability map

The writer reads **exactly five spec fields**, derived at runtime from the live config:
`content_direction.formatted` · `identity.key_differentiators` · `identity.target_audience` ·
`evidence_base.writer_block` · `design_intent.imagery_direction` (NOTES 08-19; LANDMINES :13318).

Dead surfaces that look live (findings against them are findings about the DETECTOR, not the
writer): `site_specs.voice` (0 of 1,338 prompts; it feeds the voice detector), `tone_of_voice`,
`voice_and_tone`, `audience` (most-populated aspect, zero readers), `editorial`,
`content_standards` (LANDMINES :12682–:12698, :2351–:2359). The `example_phrases.characteristic`
ARRAY is unread — only its serialisation into `formatted` matters. `evidence_base.facts[]` is
bookkeeping; `writer_block` is the wire (:7863). A field whose `source` is not `"llm"` never calls
a model at all (:5932).

Census trap for the audit: `page-content-writer`'s 12,813–14,897-char prompt lives under
`sub_workflow`, one level below where a steps walk looks — a house-voice census confidently
reported 6 of a true 7 (LANDMINES :7807). The 08-25 census (PLAN_2026-08-25_prompt_audit) used a
full recursive JSON walk for this reason.

## 6. The owner already wrote the definition of good copy, and tested it

`travelling_docs/pitch_pdf_source/REVERSE_ENGINEERED_STYLE_PROMPT_v3.md` (2026-07-17) — 18 rules
reverse-engineered from his own hand-edits and refined through two rounds of his corrections. It is
the closest thing to his ear in writing, and its three meta-lessons are this month's copy findings
stated a month earlier:

- **Rule 3's evolution**: the negative-frame rule was rewritten *"against the underlying move, not
  the specific grammar"* after the same instinct resurfaced in new sentence shapes twice — exactly
  the lane's displacement finding, and exactly homegarden's "premise, not sentences".
- **Rule 4**: word-weight overclaim in BOTH directions — grandiose AND falsely-humble ("nothing
  exotic here" still asks to be impressed). "Say plainly" / "Timing stated plainly" is this.
- **Rules 5, 11–13 + the coda**: self-flagging commentary, landing beats, manufactured contrast,
  cadence templates — *"performing insistence instead of explaining clearly"*. The methodical
  scaffold and the honesty beat are these.

Register caveat: v3 was built for outward-facing pitch prose; site copy is not identical. But his
homegarden model sentence ("We're hoping you can get a lot of useful tips from this site…") obeys
v3's rules exactly, so v3 is the right STARTING criteria set for the audit, adapted, not adopted
blind.

## 7. Mechanisms that exist, with their honest limits

| mechanism | state | honest limit |
|---|---|---|
| house voice row (CQ-022, `{{.voice_style}}`) | LIVE fleet-wide | improves copy; did NOT suppress the negation tell (2.72→2.85/1k words); emptying the row silently removes ALL voice |
| 305 negation gate (CQ-026) | LIVE; D3 mildness ruling inert until roll | holds five FORMS, ≤2/page, headlines always; brief-supplied text exempt BY DESIGN; counting default-OFF outside the writer |
| brief-negation-check cron (CQ-027) | LIVE daily | detects briefs that hand the writer the construction; files for a human, no handler |
| stage 2 / copy-editor (CQ-024) | LIVE, dispatchable (CQ-030), bounded 08-25 | ranks restatement above register — its ranking is not the owner's; 3-edit budget; proposal-only (D2) |
| per-field `llm_guidance` (CQ-031) | LIVE | strongest single writer lever (76% vs 7% behaviour split); also the carrier of CQ-005's baked-in site-specific copy |
| `audit_writer_brief.py` (CQ-025) | tool, unscheduled | resolves the true writer-visible surface first — the founding rule any audit must inherit |
| writer-side type/content gates | partial | `validate_page_content` guards the BUILD path only; section-editor path has no content validation |

## 8. Withdrawn or corrected figures — do not re-quote

- The 2.72→2.85 before/after as evidence of anything but "no suppression" — adjacent windows give
  4.35→2.85; the weekly series has no trend; *"the method cannot answer it"* (CONTRIB_2026-08-18
  answers, correction block).
- "60 exemplars → 3 verbatim = 5.8% transfer" — withdrawn: transfer is a function of exemplar
  CHARACTER; a concrete on-topic exemplar came back verbatim as a hero subheadline
  (WRONG_CALLS 08-24; CONTRIB_2026-08-24 exemplar correction).
- "Not reachable by prompt on this model" — retracted within the hour; apis.uk went 12 → 0
  whole-page tells when guard + rule + subjects landed together (CONTRIB_2026-08-23 SECOND ADDENDUM).
- Loss/gain marker ratios (16:11 vs 9:14) — *"trust neither ratio; the ranking is the finding"*.
- "unbounded auto-dispatch" — corrected 08-25: the anti-churn brake already rate-limits same-key
  re-files; the true gap was cross-type/cross-source (WRONG_CALLS 08-25).

## 9. What this refresh licenses (and what it does not)

The owner's sequencing stands: fixes come AFTER this refresh and through the audit. What the
assembled context establishes for the audit (PLAN_2026-08-25_prompt_audit.md):

1. **Judge demonstrations first** — count what each prompt's own prose models, in the five gate
   shapes AND the wider register (scaffold, candour beat, presumption, word-weight) (§2, §3).
2. **Weight by reachability** — a finding against a dead surface is noise; CQ-025's runtime
   derivation is the method (§5).
3. **Check the leak hypothesis** — prompts that instruct ABOUT methodology/integrity, matched
   against shipped self-description copy (§4).
4. **Prefer replacement-by-good-example over added prohibition** in whatever fixes follow — the
   displacement evidence forbids the other shape (§2).
5. **Acceptance is his sentence, not our regexes** — any "fixed" verdict needs a sample the owner
   (or v3-as-proxy, applied with judgment) would pass, because the pattern-list ceiling is measured
   (§3, §6).
