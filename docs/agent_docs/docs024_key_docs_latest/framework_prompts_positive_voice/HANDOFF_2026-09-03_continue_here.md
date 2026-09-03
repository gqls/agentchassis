# HANDOFF 2026-09-03 — the 641 prompt change, written up for the thread that will look at every prompt in the framework

**COLD-START for a new lane.** The owner's ask, verbatim (2026-09-03, in the finetuning.uk thread):
*"please write up the prompt change as a handoff, I will start another thread where we can look at
all the prompts in the framework."* This file is that write-up. It is written so a thread with no
context can start; the finetuning lane's own record is in
`docs/agent_docs/docs024_key_docs_latest/finetuning_uk_service/DRAFT_2026-09-02_641_positive_prompt_candidates.md`
(every draft, every rejection, the test-render) and `…/finetuning_uk_service/NOTES_finetuning_uk_service.md`
(entries dated 2026-09-02).

**First thing this lane owes itself:** create the standing five in this directory (CLAUDE.md,
"Working docs"), add a line to `~/.claude/projects/*/memory/MEMORY_workstreams.md`, and run
`scripts/who-owns.py` on every prompt before touching it (section 6 lists the ones in flight).

## 1. The owner's directive, in his words

On how to prompt (2026-09-02):

> "if you say don't think of an elephant the llm will start thinking of elephants, so we want to
> turn that around and prompt with positive prompting and also in the language that we'd expect it
> to use (not telling it to use the language) and using the sorts of terms that a person reading
> the response might be expecting but writing the prompt in that language."

On the first redraft (same day): *"It has started to hardcode what should be in it and it doesn't
have any example language or text."* Describing the arrangement in our production vocabulary
(section, subject, sibling, reader) is still the wrong register.

On the second redraft, three scene-setting candidates (2026-09-02, late): *"can you try again, they
all sound a bit AI."*

On the third (2026-09-02, late): *"go with C."*

One question was put to him and answered: **the prompt's own prose is the demonstration; there is
NO specimen answer.** A quoted exemplar in a prompt ships verbatim into live pages (memory
`a-quoted-exemplar-in-a-prompt-is-copied-verbatim`), so the voice is shown by writing the prompt
in it, never by pasting an example of the output.

Three rules of the estate that bear on all of this and were ruled before it: **the framework writes
the content, not you** (owner 2026-08-06, and 2026-08-04's "every site goes through the
framework"); the owner's tested style prompt for plain human copy is
`docs/agent_docs/docs024_key_docs_latest/travelling_docs/pitch_pdf_source/REVERSE_ENGINEERED_STYLE_PROMPT_v3.md`
(check for a higher `_vN` first) and it is the bar the second redraft was written against; and the
owner's verdict of 2026-08-25 that the served copy is "very AI sounding" belongs to the
`copy_quality_two_stage` lane, whose `PLAN_2026-08-25` put the house voice into one DB row.

## 2. The worked example: one block, four drafts, and what changed between them

The block is the piece of the page-content-writer prompt that tells a section what it is about,
so three text blocks on one page stop writing the same section (`bugs_open/443`). It renders only
when the plan assigned a subject, immediately before the Verified Facts block.

**Draft 1, migration 641 as first written (REJECTED at the owner's read, 2026-09-02):**

    ## This section's subject
    {{.current_section.subject}}

    Write THIS section about that subject specifically. Sibling sections on
    this page carry their own subjects - do not restate theirs, and do not
    widen this section into a general treatment of the page's topic.

Two negatives, capitals for emphasis, and every noun is ours (section, subject, sibling, restate,
widen). Nothing in it sounds like a page.

**Draft 2 (REJECTED: "hardcode[s] what should be in it … no example language").** Positive, but
it described the arrangement ("this section covers X; the sections around it cover Y; keep to X")
in the same production vocabulary. Not kept verbatim; the lesson is that removing the negatives
is not the change.

**Draft 3, three candidates that each set a scene (REJECTED: "a bit AI").** The lane's
recommended one, for the record:

    ## What this part covers

    What to bring to the session.

    A reader gets to this part of The playground: an hour with your own model
    wanting to know what to bring. The parts around it cover what the playground
    is, how the hour works, what you learn, questions people ask, and booking,
    each written on its own. This one has what to bring to itself.

What reads as AI in it, named so the next thread can spot the same tells in other prompts: a
staged scene ("A reader gets to this part … wanting to know"); padding phrases that perform
consideration ("each written on its own", "has what to bring to itself", and in its siblings
"give it the room it needs", "with room to do it properly"); a tidy tricolon list read out with
rhythm; and a heading in instruction register ("What you are writing").

**Draft 4, candidate C, the owner's pick (2026-09-02, late).** Filled in for the playground page,
section "what to bring":

    ## This section

    You'll want to know what to bring to the session. That's what this section is for.

    The playground: an hour with your own model also covers, each in its own section:
    - what the playground is
    - how the hour works
    - what you learn
    - questions people ask
    - booking

As the template (exclusion of the current section is by subject, because section names repeat;
the list is one per line because the renderer has no join function, see section 3):

```
{{if .current_section.subject}}## This section

You'll want to know {{.current_section.subject}}. That's what this section is for.

{{.current_page.title}} also covers, each in its own section:
{{range $s := .sections_for_render.sections_ready}}{{if and $s.subject (ne $s.subject $.current_section.subject)}}- {{$s.subject}}
{{end}}{{end}}
{{end}}
```

**What changed between draft 1 and draft 4, as transferable moves:**

1. The subject is stated the way the reader meets it ("You'll want to know what to bring"), not
   the way the pipeline holds it ("This section's subject: what to bring").
2. The rest of the page is a plain list of what it covers, which is the positive form of "do not
   restate theirs". Nothing is forbidden; the writer can see the territory is taken.
3. No instruction about language or voice. The two sentences are in the voice; that is the whole
   instruction.
4. No emphasis devices, no capitals, no "specifically", nothing the writer is told to feel.
5. The words are the site visitor's words (session, hour, book), so the response continues in them.

**And one consequence for data, not prose:** the sentence only reads if every subject completes
"You'll want to know ___". "how to book" does; "booking" does not. That rule now binds the
subject arrays the finetuning lane writes, and it exposed that the planner writes subjects in a
different register ("Publication identity and editorial scope — what gamedesign.uk covers and who
it is written for", real, gamedesign.uk 2026-09-02). Open with the owner; see section 6.

## 3. How a prompt reaches the model here: read this before editing any of them

**Where they live.** `agent_definitions.default_config`, as `prompt_template` (and a few
`system_prompt`) strings at any depth of the workflow JSON. Seeds are in
`docs/agent_docs/sql_for_agents/NNN_*.sql`, but **the live row is the fact and the seed is
history** (memory `seed-sql-is-history-live-row-is-fact`). Enumerate them live:

```sql
SELECT type, length(q.v #>> '{}') AS chars
FROM agent_definitions, LATERAL (SELECT jsonb_path_query(default_config, 'strict $.**.prompt_template') AS v) q
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL AND jsonb_typeof(q.v)='string'
ORDER BY 2 DESC;
```

**The house voice is ONE row, and most prompts do not read it.** `platform/voicestyle/voicestyle.go`
injects `agent_default_configs.config_name='voice_style_block'` (5,862 chars, last updated
2026-08-13) into any template that writes `{{.voice_style}}`, and `build_standard_block` into any
that writes `{{.build_standard}}` (`bugs_open/121`; generalised 2026-08-31 for the copy-quality
lane). `[MEASURED 2026-09-03]` **7 of 141** live prompt strings reference `{{.voice_style}}`. So
for those seven a voice change is one row, live at once; for the rest it is a migration each.
That is the first structural fact to act on: the lever exists and is mostly unconnected.

**How the template is rendered.** `ExecuteLLMPromptAction` (`platform/orchestration/actions/ai_actions.go`)
renders against `ExtractFields(CollectedData, input_fields)` (`platform/orchestration/datahelpers/unified_extractor.go:315`),
which copies ONLY the keys the step's `input_fields` names (with `input_data`'s keys promoted to
the root). A key at the CollectedData root is invisible to the template unless named, and a
`{{range}}` over it renders EMPTY with no error. This bit the 641 work on 2026-09-02
(`WRONG_CALLS.md` 2026-09-02(d)). Go `text/template`, default options (a missing map key prints
`<no value>`, which the renderer then warns on), FuncMap of exactly `toJSON`, `placeholder`,
`rangeStart`, `rangeEnd`; no `join`, no arithmetic, so a comma list with "and" cannot be built.
Go 1.24, so `and`/`or` short-circuit. **Test-render before proposing:** the harness at
`docs/agent_docs/docs024_key_docs_latest/finetuning_uk_service/render_test_641/` rebuilds the
renderer's construction against fixtures taken from live `orchestration_states` rows.

**Change control for a prompt edit.** It is a migration, so the council gate is in scope
(CLAUDE.md, `sql_for_agents` widened 2026-08-19). The owner's approval attaches to the EXACT
committed text and is voided by any later edit (`RFC_016 §5.2`; that is why 641 went back to him).
Ordering-sensitive migrations are `_HOLD.sql`, applied by hand, and **`schema_migrations` never
records them**: read the live row, not the ledger. Verify blocks anchor on verbatim lines and
pin an em-dash census, so a prompt with 5 em dashes in it must keep 5. And a prompt change is
inert for every page already served until that page is rebuilt (memory
`a-stale-page-holds-every-improvement-since-it-rendered`); size the rerender by the pages, never
by the change.

## 4. The census, 2026-09-03 (the starting map, not a judgement)

`[MEASURED 2026-09-03]` from live `agent_definitions`, all `prompt_template` and `system_prompt`
strings on active, non-snapshot, non-deleted rows. "Negations" is a crude regex
(`do not | don't | never | NOT | no `), so it ranks, it does not measure.

| | |
|---|---|
| prompt strings | **141** across **64** agent types |
| total size | **674,201** chars |
| negation matches | **1,762** |
| em dashes inside prompts | **1,560** (the house voice bans them in output) |
| strings reading `{{.voice_style}}` | **7** |

By agent type, top of the list:

| type | prompts | chars | negations | reads voice_style |
|---|---|---|---|---|
| fix-proposer | 21 | 150,131 | 431 | 0 |
| council-gate | 17 | 143,930 | 407 | 0 |
| feature-designer | 10 | 66,878 | 165 | 0 |
| experience-planner | 8 | 28,375 | 113 | 0 |
| build-site-planner | 1 | 30,896 | 64 | 0 |
| domain-research-classifier | 2 | 21,389 | 61 | 0 |
| experience-approval-council | 5 | 14,353 | 50 | 0 |
| **page-content-writer** | 1 | 14,111 | 48 | **1** |
| diagnose-agent | 1 | 12,700 | 43 | 0 |
| component-creator | 1 | 18,816 | 35 | 0 |
| grounded-explainer | 3 | 6,850 | 28 | 1 |
| tool-generator | 3 | 9,026 | 26 | 0 |
| brief-writer | 1 | 6,340 | 20 | 0 |
| meta-description-backfiller | 1 | 2,862 | 16 | 0 |

Two readings of that table. The top of it is reviewer seats (fix-proposer and council-gate are
17-seat rosters that mirror each other, `099_SYNC_gate_roster.py`), whose prompts instruct
reviewers, a different case from copy. The copy-producing prompts a visitor's eyes end up on are
`page-content-writer`, `build-site-planner` (writes the subjects and briefs the writer sees),
`component-creator`, `grounded-explainer`, `brief-writer`, `meta-description-backfiller`, and the
content-creator service's blog prompts (a separate binary, same voice row). The owner's directive
was raised about the first; the second is where the subject-phrasing question lives.

## 5. The hypothesis to test first, stated as a hypothesis

`[INFERRED, NOT MEASURED]` The page-content-writer prompt is a house-voice block that itself bans
things, followed by 18 "STRICT RULES" of which most are prohibitions ("NEVER invent", "Do NOT
write", "do not state, in any tense"), and its output was judged "very AI sounding" on 2026-08-25.
By the owner's elephant rule the prohibition register may be part of the cause: a model continues
in the register it is given. The finetuning lane noted this on 2026-09-02 and did not test it.

A test that would settle it and stay inside the rules: pick one section on one framework-built
page; keep every fact and every honesty guarantee (no invented numbers, testimonials, contact
details, accuracy promises); rewrite ONLY the form of the rules into the positive ("Every number
on this page is in the Verified Facts list below" for "NEVER invent statistics"); render both
through the framework, never by hand; put the two in front of the owner blind. Then measure
whether the guarantees still hold with the estate's own detectors (the banned-claim sweeps,
`evidence_base` gating, `brief-negation-check`). A positive rewrite that loses a safety rule is a
regression, not a style win, so the detectors are the test, not the prose.

## 6. In flight: what this lane inherits, and what it must not collide with

- **Migration 641 (page-content-writer, the block above) is the `apis.uk` lane's file**
  (`docs/agent_docs/docs024_key_docs_latest/apis_uk_bees_homepage/`, PBP-049). They write the
  final SQL from `…/apis_uk_bees_homepage/CONTRIB_2026-09-02_from_finetuning_owner_picked_C_and_the_test_render_found_two_things.md`,
  which also requires `sections_for_render` added to the writer's `input_fields` in the same
  migration and asserted in its verify block. The owner then reads the exact final words.
  **Do not edit the page-content-writer row while 641 is in flight**: a second migration on the
  same row breaks 641's anchors and voids its approval.
- **Planner subject phrasing** (`build-site-planner`, the 598/362 lineage): the owner has not yet
  chosen between nudging the planner to write each subject as the thing a reader wants to know,
  lower case, no em dash, or loosening C's first sentence. The finetuning lane recommended the
  planner nudge as a separate small migration. This is the natural first edit for this lane if
  the owner picks it.
- **`bugs_open/443`** (the mechanism that makes subjects reach the writer at all) is owned by the
  session of that name; its fallback half shipped in chassis `v1.0.1355` (probed 2026-09-03).
  The finetuning lane runs the backfill and rebuild for its own pages.
- **The house voice row** and the "very AI sounding" verdict belong to `copy_quality_two_stage`;
  contribute there (a CONTRIB file in their directory) rather than editing the row from here.

## 7. Next session, in order

0. Standing five here; workstream line in memory; `who-owns` each prompt you intend to touch.
1. Read, in this order: `finetuning_uk_service/DRAFT_2026-09-02_641_positive_prompt_candidates.md`,
   the style prompt (`travelling_docs/pitch_pdf_source/REVERSE_ENGINEERED_STYLE_PROMPT_v3.md`),
   `platform/voicestyle/voicestyle.go`, `bugs_open/121`, `copy_quality_two_stage/PLAN_2026-08-25*`,
   `RFC_016` §5.2.
2. Re-run the census above with today's date and pick the first target. The writer prompt is in
   flight; the two that are not are the voice row's reach (7 of 141) and the planner's subject
   instruction (if the owner picks the nudge).
3. Design the section-5 test before rewriting anything, and run it through the framework.
4. Every prompt edit: test-render against real loop rows, council gate, exact text to the owner,
   `Council-Submitted:` trailer if committing before the verdict.
