# DRAFT 2026-09-02 — the 641 section-subject block, redrafted to the owner's positive-prompting directive

**Status: AWAITING the owner's framing pick (question 1 below). Question 2 is ANSWERED.**
This supersedes the inserted block in `sql_for_agents/641_..._HOLD.sql` as the text the owner
will eventually read — **641 must NOT be applied as written**; its gate-2 owner read came back
as a REDRAFT direction, not an approval.
> **CORRECTED 2026-09-02 (late): 641 is the `apis.uk` lane's file, NOT `bugs_open/443`'s** — it
> is PBP-049's seed, their council round, their voids-on-edit convention. The 443 session caught
> my misattribution and has relayed this draft, the mechanics list and the untested sibling-range
> flag to apis.uk directly, so the redrafted text lands in the owner's file. My part is unchanged:
> owner picks the framing → I test-render → apis.uk writes the final SQL → owner reads the exact
> final words.

## The owner's directive, 2026-09-02 (his words, from chat)

> "if you say don't think of an elephant the llm will start thinking of elephants, so we want to
> turn that around and prompt with positive prompting and also in the language that we'd expect
> it to use (not telling it to use the language) and using the sorts of terms that a person
> reading the response might be expecting but writing the prompt in that language."

And on my first redraft attempt: *"It has started to hardcode what should be in it and it
doesn't have any example language or text."* — i.e. describing the arrangement in our production
vocabulary (section, subject, reader) is still the wrong register; the prompt itself must be
written in the voice we want back.

**Question 2, ANSWERED: written-in-the-voice only — NO specimen answer.** The prompt's own prose
is the demonstration. This deliberately avoids the documented estate trap where a quoted
exemplar ships verbatim into a live page (memory: a-quoted-exemplar-in-a-prompt-is-copied-verbatim).

## Question 1, OPEN — which scene? Three candidates, shown FILLED IN for the playground page,
## section "what to bring" (the shipped version substitutes live values — see Mechanics)

### Candidate 1 — "someone asked you" (reply to a person)

    ## What you are writing

    Someone has asked you about what to bring to the session. This is your answer.

    They are reading The playground: an hour with your own model, and further on
    they will find how the hour works, what you learn, questions people ask, and
    booking. So you can stay with what to bring and give it the room it needs.

Strongest pull toward conversational register (a reply to a person). Risk: every section may
come out answer-shaped — right for FAQ, odd for a hero.

### Candidate 2 — "a reader arrives" (write for a person) — LANE RECOMMENDATION

    ## What this part covers

    What to bring to the session.

    A reader gets to this part of The playground: an hour with your own model
    wanting to know what to bring. The parts around it cover what the playground
    is, how the hour works, what you learn, questions people ask, and booking,
    each written on its own. This one has what to bring to itself.

A person to write for (where natural register comes from) with no implied question, so no
pressure toward any shape.

### Candidate 3 — "this part has it to itself" (just the subject)

    ## What you are writing

    What to bring to the session.

    The rest of The playground: an hour with your own model is covered elsewhere:
    what the playground is, how the hour works, what you learn, questions people
    ask, booking. This part has what to bring to itself, with room to do it
    properly.

Smallest change, least steering either way.

**The dial, in one line: reply to a person → write for a person → just the subject.**

## Mechanics the redraft must honour (for whoever writes the final SQL)

- Substitutions: `{{.current_page.title}}`, `{{.current_section.subject}}`, and the SIBLING
  subject list. Sibling data is in scope (the loop persists `current_section` into the same
  CollectedData the template reads — `loop_expansion_handler.go:395-425`; the full plan is on
  `section_plan.sections_ready[]`) — **but the sibling-list range render is UNTESTED. Test-render
  against real loop data before proposing; a template that silently renders nothing is exactly
  the failure 443 was about.**
- **No em dashes in the inserted block** — 641's verify census pins the prompt's em-dash count
  at 5; plain hyphens only (its own header says so).
- The `{{if .current_section.subject}}` guard stays: a section with no subject must get a
  byte-identical v4 prompt.
- Placement unchanged: immediately before the Verified Facts block.
- The owner's approval attaches to the EXACT final text — bring him the final words, not a
  sketch, and any later edit voids it again (RFC_016 §5.2 pattern).
- British English throughout.

## Also noted for the wider register work (hypothesis, NOT measured)

The existing writer prompt is nearly all prohibition ("DO NOT INVENT", "Never approximate",
"Do not state, in any tense…"). By the owner's own elephant rule, that register may itself be
part of why output reads stiffly — a model continues in the register it is given. `[INFERRED]` —
worth an experiment on the copy lane's side, not an assertion.

---

## OWNER PICK, 2026-09-02 (late): the three above were REJECTED as "a bit AI"; a plainer set was drafted and he chose **C**

His words on the first three: *"can you try again, they all sound a bit AI"*. What read as AI: each
set up a small scene and then padded it ("give it the room it needs", "has what to bring to itself",
"each written on its own"). The second set was drafted against the owner's own style prompt
(`travelling_docs/pitch_pdf_source/REVERSE_ENGINEERED_STYLE_PROMPT_v3.md`) and showed the sibling
list ONE PER LINE, because the prompt renderer (`datahelpers.RenderPromptTemplate`) registers only
`toJSON`/`placeholder`/`rangeStart`/`rangeEnd`, so an inline list with "and" before the last item
cannot be rendered. He picked C: *"go with C"*.

**The chosen text, verbatim (filled in for playground / "what to bring"):**

    ## This section

    You'll want to know what to bring to the session. That's what this section is for.

    The playground: an hour with your own model also covers, each in its own section:
    - what the playground is
    - how the hour works
    - what you learn
    - questions people ask
    - booking

(The other two, for the record. A, "the plain brief": "What to bring to the session. / The page is
X. Its other sections cover: [list] / So this one sticks to what to bring." B, "the reader named":
"Someone reading X wants to know what to bring to the session. This section tells them. / The rest
of the page has its own sections for: [list]".)

## TEST-RENDER, 2026-09-02, done BEFORE handing to apis.uk (`render_test_641/`, this directory)

Harness: `render_test_641/main.go` builds the template the way `RenderPromptTemplate` does
(`template.New().Funcs().Parse().Execute()`, default options) and renders the C block against
five fixtures built from REAL `orchestration_states` rows (`render_test_641/fixtures.json`,
output in `render_test_641/OUTPUT.txt`). Go 1.24 (the chassis's `go.mod`), so `and` short-circuits.

| fixture | source | result |
|---|---|---|
| A | gamedesign.uk run `3ed7cdfd`, tier 1, real planner subjects | renders; **subjects read badly in C's sentence** (see below) |
| B | finetuning.uk playground run `5c804a5b`, all subjects null | block absent, v4 byte-identical, as designed |
| C | playground with the intended backfill subjects | exactly the text above |
| **D** | **siblings NOT in `input_fields` (the live config today)** | **renders "also covers, each in its own section:" followed by NOTHING. No error.** |
| E | one sibling with no `subject` key | that sibling drops out cleanly; no `<no value>` |

> **CORRECTED 2026-09-02 (late), caught by fixture D:** the Mechanics section above says *"Sibling
> data is in scope (the loop persists `current_section` into the same CollectedData the template
> reads)"*. **That is wrong for the template.** `ExecuteLLMPromptAction` renders against
> `ExtractFields(CollectedData, input_fields)` (`ai_actions.go` → `unified_extractor.go:315`), a
> SUBSET named by the step's `input_fields`, and the live `generate_content.config.input_fields` is
> `[current_section, render_context, reviewed_brief, current_page, link_context, site_plan,
> site_specs, existing_content, build_mode, rewrite_guidance]` — **no `sections_for_render`, no
> `section_plan`**. So `{{range .sections_for_render.sections_ready}}` sees nil and the list renders
> EMPTY, silently. **641 must add `sections_for_render` to that `input_fields` array in the same
> migration, and its verify block must assert it.** Only the range render was marked untested; the
> scoping claim was stated as fact. Logged in `WRONG_CALLS.md` 2026-09-02(d).

**Two rules the render forced, both carried to apis.uk in
`apis_uk_bees_homepage/CONTRIB_2026-09-02_from_finetuning_owner_picked_C_and_the_test_render_found_two_things.md`:**

1. **Exclude the current section from the sibling list by SUBJECT, not by name.** Real
   `sections_ready` on the playground row is `hero, generic-text-block, generic-text-block,
   generic-text-block, faq, call-to-action` — name-based exclusion would drop all three text
   blocks. Subject-based exclusion is what fixture C/E prove.
2. **A subject must complete "You'll want to know ___".** "what to bring to the session" does;
   "booking" does not ("You'll want to know booking."); "how to book" does. This binds the
   backfill arrays this lane writes (carried to the 443 lane's RUNBOOK by CONTRIB). It also
   exposes a question for the owner on tier-1 sites, where the PLANNER writes subjects: fixture A
   rendered *"You'll want to know Brief description of the sister-site relationship with
   gamesdesign.co.uk and what each site covers."* Capitalised noun phrases with em dashes, which
   is what the planner prompt currently produces. Either the planner's subject instruction is
   nudged to write "what/how/who…" phrases, or C's opening sentence is loosened. Owner's call;
   lane recommends the planner nudge, as a separate small migration on apis.uk's side.

**Status now: apis.uk writes the SQL from the CONTRIB; then the owner reads the EXACT final words.**
