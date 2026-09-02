# CONTRIB 2026-09-02, from the finetuning.uk lane: a phrasing rule for `section_subjects` arrays, forced by the owner's chosen 641 text

The owner picked the redrafted 641 block tonight. Its first sentence is
`You'll want to know {{.current_section.subject}}. That's what this section is for.` and the
sibling list is `- {{$s.subject}}` one per line (full text and the render evidence:
`finetuning_uk_service/DRAFT_2026-09-02_641_positive_prompt_candidates.md`, harness
`finetuning_uk_service/render_test_641/`).

**So every entry in a `pages.section_subjects` array must complete "You'll want to know ___",
lower case, no em dash.** Worked on the playground page:

| slot | reads well | reads badly |
|---|---|---|
| hero | what the playground is | The playground |
| text block | how the hour works | Hour structure |
| text block | what to bring to the session | What to bring |
| faq | what people usually ask | questions people ask (fine in the list, odd after "to know") |
| cta | how to book | booking |

Suggest adding that line to the RUNBOOK's backfill template comment, beside the same-length /
same-order rule; the alignment guard cannot see phrasing. Two things the render also settled that
touch the backfill: the current section is excluded from the sibling list by SUBJECT, so duplicate
subjects in one array would silently drop both from each other's lists (keep them distinct); and a
`null` slot drops out of the list cleanly (fixture E), so null for "no subject" stays correct.

Not yours but adjacent: tier-1 planner-written subjects do not obey this phrasing (real
gamedesign.uk example in the apis.uk CONTRIB of the same date); raised to the owner as a planner
nudge, not a backfill matter.
