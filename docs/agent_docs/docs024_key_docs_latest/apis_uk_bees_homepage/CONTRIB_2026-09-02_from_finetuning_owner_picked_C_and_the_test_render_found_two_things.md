# CONTRIB 2026-09-02, from the finetuning.uk lane: the owner picked the 641 block's text, and the test-render found two things your SQL must carry

**For:** whoever writes the final `641_page_content_writer_prompt_v5_section_subject_HOLD.sql`
(your file, PBP-049's seed). **Owner approval attaches to the EXACT final text** (RFC_016 §5.2);
he will read your final words before the apply, so change nothing in the block without saying so.
Evidence and harness: `finetuning_uk_service/render_test_641/` (`OUTPUT.txt` is the render of every
fixture below). Full account: `finetuning_uk_service/DRAFT_2026-09-02_641_positive_prompt_candidates.md`.

## 1. The block the owner chose (verbatim, his pick "C")

Rendered for playground / "what to bring", the text he read and approved the shape of:

    ## This section

    You'll want to know what to bring to the session. That's what this section is for.

    The playground: an hour with your own model also covers, each in its own section:
    - what the playground is
    - how the hour works
    - what you learn
    - questions people ask
    - booking

As the template. Plain hyphens only (641's em-dash census stays at 5), the `subject` guard kept
so an unassigned section's prompt is byte-identical to v4, placed immediately before the Verified
Facts block (i.e. before `{{if .current_section.facts_scoped}}`). The range line has no trailing
whitespace; the block ends with one blank line before the facts block:

```
{{if .current_section.subject}}## This section

You'll want to know {{.current_section.subject}}. That's what this section is for.

{{.current_page.title}} also covers, each in its own section:
{{range $s := .sections_for_render.sections_ready}}{{if and $s.subject (ne $s.subject $.current_section.subject)}}- {{$s.subject}}
{{end}}{{end}}
{{end}}
```

`{{.current_page.title}}` resolves: `current_page` lives under `input_data` and the extractor's
`input_data` special case promotes it to the root (`unified_extractor.go:40-55`); the v4 first
line already relies on it. Exclusion is by **subject**, not name, because names repeat
(`generic-text-block` ×3 on the real playground row `5c804a5b`); a name test would drop all three.
Go 1.24 short-circuits `and`, so a sibling with no `subject` key drops out without error
(fixture E).

## 2. REQUIRED in the same migration: `sections_for_render` into the writer's `input_fields`

The prompt is NOT rendered against CollectedData. `ExecuteLLMPromptAction` renders against
`ExtractFields(CollectedData, input_fields)` (`ai_actions.go`, `unified_extractor.go:315`), which
copies only the keys the step names. Live today:

```
generate_content.config.input_fields = [current_section, render_context, reviewed_brief,
  current_page, link_context, site_plan, site_specs, existing_content, build_mode, rewrite_guidance]
```

No `sections_for_render`. **Fixture D (that config, this block) renders "also covers, each in its
own section:" followed by NOTHING, with no error.** That is the 443 failure one level up. So 641
must also:

```sql
-- same UPDATE, second jsonb_set, path:
-- {workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,input_fields}
-- append "sections_for_render" (it is set before the loop; it is the loop's own iterate_over source)
```

and the verify `DO` block must RAISE unless BOTH the template contains
`{{range $s := .sections_for_render.sections_ready}}` AND `input_fields` contains
`"sections_for_render"`. A verify that checks only the template passes on a config that renders
an empty list. Note `input_fields` names a CollectedData KEY; the extractor stores it under the
same simple name, so the template path is `.sections_for_render.sections_ready`.

## 3. A subject must complete "You'll want to know ___" (owner question raised, not decided)

Our backfill arrays will obey this ("what to bring to the session", "how the hour works",
"how to book", never "booking"); carried to the 443 lane's RUNBOOK by CONTRIB. On tier-1 sites the
PLANNER writes the subject, and fixture A (gamedesign.uk run `3ed7cdfd`, real) rendered:

    You'll want to know Brief description of the sister-site relationship with gamesdesign.co.uk and what each site covers. That's what this section is for.

    gamedesign.uk — The Practice of Game Design also covers, each in its own section:
    - Publication identity and editorial scope — what gamedesign.uk covers and who it is written for
    - The most recent or most significant published article, surfaced as the lead piece

Capitalised noun phrases with em dashes. Two options are with the owner: nudge the planner's
subject instruction (`build_site_planner`, the 598 lineage) to write each subject as the thing a
reader wants to know, lower case, no em dash; or loosen C's first sentence. The finetuning lane
recommended the planner nudge, as a separate small migration. **Do not change C's words to
absorb this; the owner's pick was the words.**

## 4. What happens next

You write the SQL → tell the finetuning lane (a CONTRIB in `finetuning_uk_service/`, or a line in
your NOTES we will read) → we carry the EXACT final text to the owner for the fresh read → apply.
Stage B of the 443 backfill plan waits on this; Stage A does not.
