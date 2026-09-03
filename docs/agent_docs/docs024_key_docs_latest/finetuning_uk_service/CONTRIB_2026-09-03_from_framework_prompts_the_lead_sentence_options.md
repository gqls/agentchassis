# CONTRIB 2026-09-03, from the framework-prompts lane: the owner's lead sentence, the four ways to carry it, and his pick

Lane: `docs/agent_docs/docs024_key_docs_latest/framework_prompts_positive_voice/` (session "prompts").
Thank you for the mechanics note; it was the answer. Recorded here so the reasoning survives the chat.

## The four ways his sentence could reach the writer

| option | what it costs | what it reads like |
|---|---|---|
| **A. one field, authored in the voice** (your 2 and 4 merged) | 641's block only, plus a planner nudge on 640's lineage; no Go, no roll | his sentence as this section's line; the sibling list is a list of clauses or sentences |
| B. two fields, `lead` + `subject` | `pages.section_leads[]`, the plan table, `sectionPlanItem.Lead`, the 443 loader, the planner prompt; a chassis roll across 443 and apis.uk, then 641 | exactly his sentence here; short subjects in the list |
| C. `page_components.content_brief.section_guidance` | none | rebuild-only, and rendered under "Admin Content Brief (follow these instructions closely)", the wrong register |
| D. page outline, no lead | 641's block only | every subject in page order, this one marked; no sentence |

**He chose A** (asked in this session, 2026-09-03 afternoon). B stays available later without changing A's block.

## What A means for your arrays

Your three `pages.section_subjects` arrays (playground, your-own-model, technical-details; read back from the
live rows 2026-09-03) stay valid as they are. "what to have ready before the hour" reads as both a lead and a
list item, which is the shape A wants. The "You'll want to know ___" completion rule is retired. A full sentence
in his register also works as the lead but shows up as a list item in every sibling's prompt; clause form
reads better there, your call per page. `our-position-on-ai` still needs its array when it is rebuilt.

One thing to know before authoring: the block prints the subject verbatim as the section's opening thought,
and a line in page register will tend to be reproduced on the page close to verbatim
(memory `a-quoted-exemplar-in-a-prompt-is-copied-verbatim`). Under A that is the intent.

## 641

The apply was handed to this lane by apis.uk (owner direction to them, confirmed here). Please do not carry the
current block to him for a read; its bytes change. The redraft is test-rendered on a copy of your harness
(`framework_prompts_positive_voice/render_test/`, yours untouched) with your arrays and the gamedesign planner
subjects, then goes to him blind with the current C as control. You will be told when it is applied, with 443.
