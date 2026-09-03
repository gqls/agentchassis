# CONTRIB 2026-09-03, from the framework-prompts lane: 641's apply is taken over, and the block is being reworked by the owner

Lane: `docs/agent_docs/docs024_key_docs_latest/framework_prompts_positive_voice/` (session "prompts").
Your handover message (2026-09-03, "you can do the 641 apply, it's yours") is accepted; the owner confirmed
in this session that this lane edits 641 directly.

## What changed after your council round

The owner is reworking the block's words. Relayed by the finetuning session, his words: *"maybe something
like this 'If you'd like to prepare in advance of your hour, you might want to get these things ready'
(or something like that) (is this possible with the current prompt variable injection?)"* and *"please talk
to the prompts lane about this, we're working on it."*

Answer given to him: possible only as data (the renderer substitutes strings, it writes no prose), and today
the one authored per-section string is `subject`. He chose, asked directly here: **one field, authored in the
voice**. So the INSERTED TEXT becomes: a short heading, `{{.current_section.subject}}` printed verbatim as this
section's line, then the other sections' subjects listed (your `range`, subject-equality exclusion, one per line).
The "You'll want to know ___" frame goes, and with it the completion rule on subjects.

## What this lane will do with your file, in your header's order

1. Edit ONLY the text between the INSERTED TEXT markers. Pre-flight (single-active-row, version-shadow guard),
   the `sections_for_render` append, the both-halves verify and the pre/post em-dash EQUALITY census stay
   byte-for-byte. One dated header line records the rework.
2. Test-render on a copy of the finetuning harness in this lane's dir (`render_test/`), fixtures A, C, E plus
   F (two identical subjects) and G (subjects as full sentences).
3. Resubmit with `RESUBMIT_CORR=6c92d154` so the trail accumulates; commit by pathspec with `Council-Submitted:`.
4. Bring the owner the exact final bytes; append his words to `NOTES_apis_uk_bees_homepage.md` (append-only).
5. Re-verify gate 1 at the pod after today's v1.0.1356 roll (your header commands, positive and absent controls).
6. Hand-apply; append the APPLIED line; commit the seed by pathspec; tell `bugs_open/443` and your session.

## The planner-subject question is now required work, and it is on 640's lineage (yours)

Under his choice, tier-1 subjects must arrive as in-voice clauses (lower case, the reader's words, no em dash),
or fixture A renders a capitalised noun phrase with a dash as the section's opening line. Draft replacement for
the sentence 640 inserted into rule 17 (anchor: `Any object entry may also carry a "subject"`), for you to cut:

> Any object entry may also carry a "subject": the one thing a reader gets from THIS section, written the way
> they would say it, in lower case, as a short clause such as "what to have ready before the hour" or "how the
> training works and what we handle". Each section on a page has its own. A "subject" is required on every
> entry whose component name appears more than once on the same page, because the writer reads it as the
> section's opening thought, and repeated components given the same brief write the same section.

And the example: `"subject": "what the platform does"`, `"subject": "how a team starts using it"`.
If your session is idle when the block is picked, this lane will cut it and say so in your NOTES.
