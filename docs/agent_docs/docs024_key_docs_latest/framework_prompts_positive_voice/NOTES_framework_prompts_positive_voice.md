# NOTES — framework prompts, positive voice (append-only, newest at the bottom)

## 2026-09-03 (afternoon) — lane opened; the state had moved since the handoff; three owner decisions

**What the handoff said and what was true by the time it was read.** The 09:58 handoff (`HANDOFF_2026-09-03_continue_here.md`)
says migration 641 is the apis.uk lane's file and that they will write the final SQL from the finetuning
CONTRIB. By 13:00 `[MEASURED]`: 641 is committed at HEAD carrying candidate C (commits `7da6c6a46`,
`6e8d04b6b`), with the `sections_for_render` append, a both-halves verify and a pre/post em-dash
EQUALITY census; council run `6c92d154` reached `complete_approved` at 09:22Z; the apply was rehearsed
twice under ROLLBACK; and the apis.uk session handed the apply to this lane by cross-session message
("Owner direction: you can do the 641 apply"). A survey agent I ran at ~12:30 reported the file "still
carries draft 1" from `git log` at `c9c9b75ec`: stale by two commits within the hour. Lesson, already in
memory: a record goes stale faster than its reader can tell; I re-read the file and `git log` before
acting, which is what caught it.

**The owner's "baseline" was ambiguous and I asked.** Four readings (candidate C; the house voice row;
the model choice; the v3 style reference). Answer: "all of the above, I was referring to Candidate C originally".

**Mid-plan, relayed by the finetuning session, the owner's new shape for the block:** *"maybe something
like this 'If you'd like to prepare in advance of your hour, you might want to get these things ready'
(or something like that) (is this possible with the current prompt variable injection?)"* and *"please
talk to the prompts lane about this, we're working on it."* Mechanics (finetuning's note, verified against
`plan_sections_action.go:1223` and `datahelpers.RenderPromptTemplate`): the renderer substitutes strings
and generates no prose; the only authored per-section string today is `sectionPlanItem.Subject`
(`site_plan_sections.subject` on tier 1, `pages.section_subjects[]` on tier 3). `page_components.content_brief`
is a second per-slot free-text channel but exists only on REBUILDS and renders under an instruction-register
heading, so it is the wrong door.

**Three owner decisions, asked in this session:**
1. This lane edits 641 directly.
2. The model arm runs this session, after the new prompts exist, under a $25 ceiling; no model change ships.
3. Lead sentence: option A, one field authored in the voice. The block prints the subject verbatim as this
   section's line and lists the other sections' subjects; a planner nudge follows on 640's lineage.

**Measurements taken today** (all `[MEASURED 2026-09-03]`): 141 live prompt strings, 7 read `{{.voice_style}}`,
3 read `{{.build_standard}}`; models sonnet-5 87 steps/21 types, sonnet-4-6 42/33, haiku-4-5 26/23,
opus-4-6 6/3, opus-4-8 1/1; writer `generate_content` 5,058 calls / 38.39M in / 6.94M out in 7 days,
cache reads 0 on every call; the aiservice splits at `<!--CACHE_BREAKPOINT-->` (`anthropic.go:136`) and the
writer prompt has no marker, with its volatile line first.

**Two errors in the handoff, corrected in place:** `bugs_open/121` is `bugs_closed/121`;
`scripts/who-owns.py` cannot resolve a prompt name.

**Missteps so far:** none of mine that reached a document, but the plan file carried "641 still carries
draft 1" for about an hour on a stale agent report until the direct `git show HEAD:` check replaced it.
The check that caught it: read the file at HEAD, not a report about the file.
