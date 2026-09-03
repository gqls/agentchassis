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

## 2026-09-03 (~15:00Z) — Step 1: four block candidates rendered; blind key sealed here before the owner sees them

Harness: `render_test/` (this lane's copy), `go run . > OUTPUT.txt`. Fixtures added: C2 (the LIVE playground array,
read back 2026-09-03), F (two sections share a subject: a data defect 640 forbids), G (subjects as full sentences,
fixture text I wrote, not copy). All four candidates: 0 em dashes in the template, no `<no value>` on any fixture,
B (null subjects) renders nothing, D (siblings absent from input_fields) renders an empty list SILENTLY as before,
which is why 641's input_fields half stays. F: A2 marks BOTH duplicates (`!! MARKER DOUBLED`, visible); A1 and A3
drop both from the list (silent). Either way the fix is the data.

**Blind key (seed 20260903), sealed before presentation:**

```
R = C_control_as_committed_in_641
P = A1_rest_of_the_page
Q = A2_page_in_order_this_one_marked
S = A3_elsewhere_on_the_page
```

The four templates are in `render_test/main.go`; the renders in `render_test/OUTPUT.txt`. Under option A every
candidate prints the subject verbatim, so on planner-written subjects (fixture A) they all show the capitalised,
em-dashed planner phrase as the section's line: the planner nudge is required data work, not a block choice.

## 2026-09-03 (~15:30Z) — the owner read the four blind, leaned R (the committed control), and the frame's cost is now MEASURED

He picked **R**, which is the currently committed candidate C, his own 2026-09-02 pick chosen again without
knowing it. His words: *"I think R because it is a bit more verbose and friendly, but not sure."*

The "not sure" is well placed, and the harness answers it. R's first sentence is a frame with a hole, and only
one shape of subject fits the hole. Both renders below are real harness output, not composed:

- On the REAL gamedesign.uk planner subjects (fixture A): *"You'll want to know Brief description of the
  sister-site relationship with gamesdesign.co.uk and what each site covers. That's what this section is for."*
- Carrying HIS OWN example sentence (fixture G): *"You'll want to know If you'd like to prepare in advance of
  your hour, you might want to get these things ready.. That's what this section is for."* (note the doubled
  full stop, which no rule catches).

**So R's frame and his own sentence idea are mutually exclusive**, and that is the decision put back to him:
keep the frame and subjects are permanently short lower-case "what/how" clauses on every site including
planner-written ones, or drop the frame (A4 = R's second half unchanged, subject printed verbatim) and the
opening line can be written per section in his words. A4 was added to the harness for this and renders clean
on every fixture. Awaiting his answer.

**Whatever he picks, this lane owes a PHRASING SPEC for subjects**, because two other lanes author that data:
finetuning's backfill arrays and apis.uk's index page (generic-text-block ×6 with zero subjects on the plan
rows, per the 443 lane's exposure CONTRIB). The 443 CONTRIB's "must complete 'You'll want to know ___'" rule
is stale under the rework and must not be followed until the spec lands.

**From apis.uk, recorded in the RUNBOOK because they are traps, not preferences:** the block's text lives
twice in 641 (header comment and the `E'...'` string) and must be byte-compared after decoding; the pre-flight
"already applied" probe keys on the opening `{{if .current_section.subject}}` literal; `::text LIKE` on the
planner row returns a clean false for any literal containing a double quote (JSON stores `\"`); that prompt
moved 169 chars in an hour, so no absolute offsets; and the 450 lane's migration 729 pins the same rule-17
anchor and is blocked on an owner permission decision, so they re-anchor on top of the nudge, not the reverse.
