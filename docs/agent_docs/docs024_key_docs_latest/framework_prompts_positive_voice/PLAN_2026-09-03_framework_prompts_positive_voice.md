# PLAN 2026-09-03 — framework prompts in the positive voice, and the four baselines to revisit

Lane copy of the approved plan (source: this session's plan file, approved by the owner 2026-09-03). Decisions and corrections land here as dated additions; the plan text below is the approved version, unedited.


Lane directory: `docs/agent_docs/docs024_key_docs_latest/framework_prompts_positive_voice/`
Cold-start read: `HANDOFF_2026-09-03_continue_here.md` in that directory (commit 79c952692).

## Context

On 2026-09-02 the owner ruled that framework prompts should be written positively, in the
language the reader of the output expects, rather than as lists of prohibitions ("if you say
don't think of an elephant the llm will start thinking of elephants"). He picked candidate C for
the section-subject block in migration 641, then asked for a handoff so a new thread could look
at every prompt in the framework. Today (2026-09-03) he added that he wants to revisit "the
existing baseline choice", clarified as all four of:

1. **Candidate C**, the 641 block ("You'll want to know {{subject}}. That's what this section is for.").
2. **The house voice row** (`agent_default_configs.voice_style_block`), the base voice every copy prompt inherits.
3. **The model choice** (page-content-writer runs `claude-sonnet-5`).
4. **The style reference** (`travelling_docs/pitch_pdf_source/REVERSE_ENGINEERED_STYLE_PROMPT_v3.md`).

And then, relayed by the finetuning session mid-plan, his new shape for the block, in his words:
*"maybe something like this 'If you'd like to prepare in advance of your hour, you might want to
get these things ready' (or something like that) (is this possible with the current prompt
variable injection?)"* and *"please talk to the prompts lane about this, we're working on it."*

Anthropic's own prompt-audit guidance (bundled `claude-api` skill, `shared/prompt-audit.md`)
says the same thing: "a prohibition against a failure the model wasn't going to make can anchor
it toward that failure"; "the prompt's register becomes the output's register"; and for Sonnet 5,
"positive examples showing the desired concision tend to be more effective than telling the
model what not to do". Its one exception is the one this estate needs: prohibitions encoding a
real business constraint (no invented numbers, testimonials, contact details, accuracy promises)
stay, with their reason beside them. Style prohibitions get restated as what good looks like.

## What is true today (all MEASURED 2026-09-03 unless marked)

| fact | value | where |
|---|---|---|
| live writer prompt | 14,111 chars, unchanged; model `claude-sonnet-5`; `input_fields` lacks `sections_for_render` | live `agent_definitions` row |
| **migration 641** | **committed at HEAD with candidate C** + `sections_for_render` append + both-halves verify + pre/post em-dash EQUALITY census (commits 7da6c6a46, 6e8d04b6b). Council run **6c92d154 APPROVED** (`complete_approved`, 09:22 UTC). Rehearsed twice under ROLLBACK. Gate 1 (binary + 639) pod-verified. **Gate 2, the owner's read of the exact block, is now superseded by his rework.** The apis.uk session has handed the apply to this lane. | file header lines 43-52 (INSERTED TEXT markers); apis.uk peer message |
| migration 640 (planner subject rule) | APPLIED 2026-09-02; example subjects are capitalised noun phrases; real ones carry em dashes (fixture A) | `sql_for_agents/640_*_HOLD.sql:45`; `finetuning_uk_service/render_test_641/OUTPUT.txt` |
| per-section strings reaching the writer | exactly ONE authored string per section: `sectionPlanItem.Subject` (tier 1 from `site_plan_sections.subject`, tier 3 from `pages.section_subjects[]`). `content_brief` (purpose/tone_direction/section_guidance) exists per `page_components` slot but only on REBUILDS, rendered under an instruction-register heading | `plan_sections_action.go:1223`; `v3_site_actions.go:5324-5370`; `save_page_sections_action.go:1091` |
| renderer | Go `text/template`; FuncMap `toJSON`/`placeholder`/`rangeStart`/`rangeEnd`; substitutes strings only, generates no prose | `datahelpers.RenderPromptTemplate` |
| house voice row | 5,862 chars; text is migration 628's (2026-08-25, form-only; `updated_at` still reads 08-13 because 628 never bumped it); scanner population F at 0 demonstrations; read by **7 of 141** prompt strings; rendered under a `## ` prefix in the writer, so its first line becomes a heading | live row; `copy_quality_two_stage/AUDIT_prompts/PHASE2_2026-08-25_*.md:77-104`; `sql_for_agents/628_house_voice_v3_form_rewrite.sql` |
| build standard row | 894 chars (2026-08-31); read by 3 prompts | live row |
| fleet models | sonnet-5: 87 steps / 21 types · sonnet-4-6: 42 / 33 · haiku-4-5: 26 / 23 · opus-4-6: 6 / 3 (planner) · opus-4-8: 1 / 1 | live `ai_service` blocks |
| writer volume, 7 days to 09-03 | `generate_content` steps: **5,058 calls, 38.39M in, 6.94M out**; all steps 7,005 calls; **cache reads 0 on every call** | `llm_call_log` |
| model screen (08-31) | sonnet NEG=5; `claude-fable-5` NEG=0 ×2, grounding kept, owner: too dense; gemini failed grounding; Grok blocked on xAI funding; decision "7. Model choice" left with the owner, then dropped from the 09-02 handoff. **No committed harness**: calls were made from a chassis pod with platform credentials, key never read into a session | `copy_quality_two_stage/AUDIT_prompts/EXPERIMENT_2026-08-31_model_trials.md`; `HANDOFF_2026-08-31_continue_here.md:108` |
| pricing ($/MTok in/out, Anthropic first-party, cached 2026-06-24) | sonnet-5 2/10 · sonnet-4-6 3/15 · opus-5 5/25 · opus-4-6 5/25 · fable-5 and fable-5-1 10/50 · haiku-4-5 1/5 | bundled `claude-api` skill |
| style reference v3 | 2026-07-17, 18 rules for REWRITING pitch copy; rules 8-10 and 14-16 are markdown/table/fenced-log mechanics with no place in a voice block; not read by any agent, but named as an attribution line inside the meta-description backfiller's prompt text (`488_*.sql:215`) | `travelling_docs/pitch_pdf_source/` |
| detectors at the output | `rewrite_negations` gate (every non-exempt hit repaired since 08-31; budget machinery inert), BANNED_REGISTER v2 / `registerwords.go`, `evidence_base` gating, `brief-negation-check` (reads BRIEFS, not prompts), `audit_prompt_demonstrations.py` (prompts; population F = the row), `count_negation_tells.py` (served pages) | `rewrite_negations_action.go`; `negationtells.go:125-139`; copy lane scripts |
| two handoff errors | `bugs_open/121` is `bugs_closed/121` (fixed, live); `scripts/who-owns.py` resolves bug numbers only, so every prompt name returns "no match" | agent survey |

## Owner decisions taken during planning (2026-09-03)

1. **This lane edits 641 directly** (the apis.uk session had already handed the apply over).
2. **The model arm runs this session, after the new prompts exist**, under the $25 ceiling; no model change ships.
3. **Lead sentence: option A, one field authored in the voice, now.** The subject is the sentence or clause; 641's block prints it verbatim as this section's line and lists the other sections' subjects; a small planner nudge on the 640 lineage follows; the two-field shape (B) stays available later without changing the block.

## Sequencing, with reasons

```
0. lane setup, peer replies, three CONTRIBs                     (first; nothing else can honestly start)
1. the 641 block: answer the owner's question, redraft, test-render, he picks, edit 641, resubmit, apply
2. house voice row v3 (+ mapping table) -> off-line 2x2 experiment -> owner reads -> migration -> canary
3. style reference v4, from the mapping table                    (human record of the same standard)
4. observations for the owner (no work)
5. follow-on: the 141-prompt sweep, method only
```

Why 641 first: it is the owner's original meaning, it is live in his hands right now, and it is
council-approved and rehearsed, so the only new work is the block's words. Why the row before v4:
the row is the production lever, read on 5,058 writer calls a week; v3 is a rewriter prompt for
pitch decks that no agent reads, and half its rules do not belong in a voice block. Building v4
first would spend the owner's reading on a document nothing consumes, then re-derive the row.
One mapping table, two renderings: row first, v4 second. Why the model experiment sits INSIDE
step 2 rather than after it: the 08-31 screen ran off-line against a `prompt_rendered` row; the
same method with the candidate row substituted in IS the row's pre-apply test, and it arrives
with the owner's read ("with this text, sonnet went from NEG 5 to x"). The framework canary is the
post-apply confirmation.

## Step 0: lane setup (first, ~30 min)

- Standing five in the lane dir: `PLAN_2026-09-03_framework_prompts_positive_voice.md` (this
  plan, as the lane's copy), `RUNBOOK_framework_prompts_positive_voice.md`,
  `NOTES_framework_prompts_positive_voice.md`, `README_where_we_are.md`; a `SUMMARY_*` only at a
  milestone. One line in `~/.claude/projects/-home-ant-projects-agentchassis/memory/MEMORY_workstreams.md`.
- Reply to the three peer sessions (SendMessage): finetuning (cold-start received; coordinating on
  the owner's new sentence; do not carry the old C block to him for a read), apis.uk (641 apply
  accepted by this lane; the block is being reworked so the approved bytes will change and
  6c92d154 will be resubmitted with `RESUBMIT_CORR`; their handover checklist will be followed on
  apply), 443 session (told after apply, per apis.uk's handover).
- Three CONTRIBs: `apis_uk_bees_homepage/CONTRIB_2026-09-03_from_framework_prompts_641_taken_over_block_under_rework.md`;
  `finetuning_uk_service/CONTRIB_2026-09-03_from_framework_prompts_the_lead_sentence_options.md`
  (their backfill arrays are the data this decides); `copy_quality_two_stage/CONTRIB_2026-09-03_from_framework_prompts_house_voice_and_model_choice_are_being_worked_here.md`.
- Fix the two handoff errors in place (visible corrections).

## Step 1: the 641 block, reworked around the owner's sentence

**The answer to his question.** The template cannot compose prose; it substitutes strings. A
sentence like his has to arrive as DATA, one authored string per section. Today there is exactly
one such string, `subject`. So: yes, possible today, if the subject is authored as the sentence
(or a clause that reads as one); properly, with a second per-section field. Four mechanisms:

| option | what changes | cost | what it reads like |
|---|---|---|---|
| **A. one field, authored in the voice** (finetuning's options 2/4 merged) | 641's block only: frame drops to a heading, the subject prints verbatim as this section's line, siblings listed by subject. Planner nudge on the 640 lineage so tier-1 subjects arrive as in-voice clauses/sentences. Finetuning re-author their three arrays | one migration on the writer row (641 revised) + one small planner migration; no Go, no roll | his sentence as this section's line; the sibling list is a list of clauses or sentences, which reads well for clauses and oddly for full sentences |
| **B. two fields, `lead` + `subject`** | `pages.section_leads[]` beside `section_subjects`, the plan table, `sectionPlanItem.Lead`, the 443 fallback loader, the planner prompt asked for both; 641 prints `lead` here and `subject` for siblings | a chassis change and a roll across the 443 and apis lanes, then 641 | exactly his sentence for this section; short subjects in the list |
| C. `content_brief.section_guidance` | none | none | rebuild-only, and rendered under "Admin Content Brief (follow these instructions closely)": instruction register, the wrong door |
| D. the page outline, no lead | 641's block only | as A without the planner nudge | every subject in page order, this one marked; accepts any subject format; gives him no sentence |

**Recommendation: A now, B as the follow-on if the sentence-as-list-item reads badly on real
pages.** A answers "is this possible" with yes-today, and B's data shape can be added later
without changing A's block (the block would print `lead` when present, else `subject`). Whichever
he picks, the block under it is written in the voice: no "That's what this section is for", no
"each in its own section".

**One thing to say to him plainly before he picks:** a lead written in page register will tend to
be reproduced on the page close to verbatim (memory `a-quoted-exemplar-in-a-prompt-is-copied-verbatim`).
With A that is the intent: the subject is the section's opening thought and the writer continues
it. If he would rather the writer find its own first sentence, D is the shape.

**Candidates to test-render** (copy `finetuning_uk_service/render_test_641/main.go` + `fixtures.json`
into the lane dir; do not edit theirs). Fixtures: A (real gamedesign planner subjects), C
(playground backfill), E (a sibling with no subject), new F (two identical subjects, marker must
not double), new G (subjects authored as full sentences, to show the list shape under A).
Renders shown to the owner blind, letters randomised, C-as-committed included as control.

**Edit 641** once he picks: the INSERTED TEXT block only (his old approval attaches to those bytes
and is already void by his rework); keep the pre-flight, the `sections_for_render` append, the
both-halves verify and the equality census untouched (do not "fix" a census number; the file
computes it). Header gains a dated line recording the rework. Resubmit to the council with
`RESUBMIT_CORR=6c92d154`; commit by pathspec with `Council-Submitted:`; he reads the exact final
bytes; record his words in `apis_uk_bees_homepage/NOTES_apis_uk_bees_homepage.md` (append-only);
re-verify gate 1 at the pod (header commands); hand-apply; append the APPLIED line; commit; tell
443 and apis.uk. `_HOLD` in the filename never changes on apply.

**Planner nudge** (option A): a separate small migration on `build-site-planner` anchored on 640's
rule 17 sentence, asking for the subject as the thing the reader wants, in the reader's words,
lower case, no em dash. Its own council round.

## Step 2: the house voice row, rewritten in the voice

**Deliverables.** (a) v3 text for `voice_style_block`, 4,500-7,000 chars, first line short (it
renders as a heading), no example page sentences (the two quoted ones today, the debts and the
loan, are the exemplar-lift risk and go); (b) a 17-row mapping table in the lane NOTES: live
paragraph, new paragraph, reason kept, and the DETECTOR that guarantees the rule at the output
where one exists; (c) migration + `_ROLLBACK`; (d) the 2x2 experiment (below); (e) one framework
canary after apply.

**Method.** For each of the live row's paragraphs: name the constraint and its reason; rewrite as
prose in the voice that describes the copy that results, carrying the reason. Where the rule is
enforced at the output by a detector (em dash: regex in the gate; the cut-list words:
`registerwords.go` / BANNED_REGISTER v2; contrastive pairs: `negationtells.go`; exclamation
marks: regex), the prompt need not carry the list: naming fourteen banned words IS the elephant.
The mapping table records the detector as the guarantee, and the experiment measures incidence
with and without. Rules with no detector (word-weight, contractions, headings, read-it-aloud)
stay, as positive descriptions. Fold in the two density rulings (08-31 ruling 13, 09-02 "guides
can be shorter") as two sentences. Keep the precedence sentence's meaning (site spec outranks the
house voice outranks marketing instinct). Business-constraint rules keep their reason beside them.

**The migration.** Copy 628's skeleton (`migration_backups` row, drift anchor in a pre-flight
`DO` block, `jsonb_set(config,'{text}',…)`, paired `_ROLLBACK.sql` restoring the previous text)
and add the verify half 628 lacks: from 240, exactly-one-row, a length band, `IF t LIKE '%—%'
THEN RAISE` (absolute form: the row bans em dashes outright), zero `!`, one landmark phrase per
surviving rule (`position()` checks, the closest SQL gets to "meaning kept"), and the six
negation regexes from `audit_prompt_demonstrations.py:49-56` ported to `regexp_matches` so the
row's own demonstrations are asserted at 0 in the transaction. Anchor the drift check on a 628
phrase ("No em dashes, anywhere, ever" or "A contraction keeps the company of its neighbours"),
not on the deleted v2 wording. Bump `updated_at`. `INSERT INTO schema_migrations` in the file.
Dry-run under `BEGIN … ROLLBACK` against the live row first; induce one verify failure to prove
the block can fail. Council gate (migration in scope). Not `snapshot_agent`: it does not cover
this table.

**The 2x2 experiment (the row's pre-apply test, and the model arm in one).** P0 = a real writer
`prompt_rendered` from `llm_call_log`; P1 = P0 with today's row text replaced by the candidate
(assert exactly one replacement, since the writer prefixes it with `## `). Models: `claude-sonnet-5`,
`claude-fable-5-1` (re-baselines the 08-31 Fable 5 result at the same price), plus on P1 only
`claude-fable-5-1` + the vendor's density line ("Please remove all mannered prose.") and
`claude-opus-5`. Six sections from three sites, one of them finetuning's about-content for
continuity with 08-31; n=2. 6 sections × (2×2 + 2) cells × 2 = 72 calls; at the measured averages
(7,057 in / 1,624 out) about $9; ceiling $25. Production temperature and `wire_max_output_tokens`.
Calls run FROM a chassis pod using the platform's own credentials, as the 08-31 screen did: the
key never enters the session (memory `never-extract-keys-probe-from-the-pod`; BusyBox `wget`
drops 4xx bodies, so check for `curl` in the image first and record the recipe in the RUNBOOK).
Measured per output: the 305 negation battery, cut-list hits, em dashes, `!`, words per sentence
and paragraphs per 300 words (density), every Verified Fact and internal link present, invented
digit-runs. Owner reads three sections × cells blind, letters randomised, key sealed in NOTES
until his ranking is recorded. Pre-registered question: does P1 change the sonnet-versus-Fable
gap? Cost report: per call and per week at the measured volume (**$146 sonnet-5 / $365 opus-5 /
$731 fable-5-1 per week** on the 09-03 numbers), no recommendation; the cost call is his.

**Owner read, apply, canary.** The row in full read as prose (that is the test), the mapping
table with every dropped prohibition beside its detector, the experiment numbers, the blind
outputs. On his pick: migration, gate, apply (live within the 60 s cache), then ONE page rebuilt
through the framework: served by a `{{.voice_style}}` prompt, rebuildable (finetuning about and
services are unrebuildable per `bugs_open/422`), carries Verified Facts so invention detectors
have teeth, no subjects so 641 cannot confound. Before and after, dated: `count_negation_tells.py`
on the served URL, `audit_prompt_demonstrations.py` on the rendered prompt, BANNED_REGISTER v2,
`evidence_base` check, em dash / `!` / cut-list counts, density proxies. Every other served page
keeps the old voice until rebuilt; size any wider rerender by the pages, never by the change.
Ownership: the row is the copy lane's; the CONTRIB carries the ready migration and asks for
written consent for this lane to apply, else they apply.

## Step 3: style reference v4

New file `travelling_docs/pitch_pdf_source/REVERSE_ENGINEERED_STYLE_PROMPT_v4.md`; v1-v3 untouched.
Written from the Step 2 mapping table as the human record of the same standard: the voice rules
in the positive form, each with its reason and a worked before/after (examples belong here, never
in the row), plus the rewriter-only mechanics v3 carries (tables, fenced logs, staccato lists)
kept as a separate section addressed to rewriting, not to a writer's voice block. Cross-reference
the row's migration number so the two cannot drift silently. Nothing depends on v4; it is the
first thing to cut if short.

## Step 4: observations for the owner (report only)

- 42 live steps across 33 agent types run `claude-sonnet-4-6` at $3/$15; `claude-sonnet-5` is $2/$10.
- The planner runs `claude-opus-4-6`; `claude-opus-5` is the same price.
- Nothing in the fleet runs Fable; the Fable question is the writer's volume, which Step 2 prices.
- **Writer cache reads are zero on all 7,005 calls last week.** The mechanism exists:
  `platform/aiservice/anthropic.go:136` splits any prompt at a literal `<!--CACHE_BREAKPOINT-->`
  and sets `cache_control` on the prefix (how migration 377 got the council seats a measured
  68% saving). The writer prompt has no marker, and its first line is the volatile one ("Write
  content for the X section of Y") ahead of ~20k stable chars, so nothing could hit anyway.
  Reordering plus the marker would remove most of the input side (about $77 of the $146/week)
  at any model. A structural change to the writer row: after 641, its own council round, not
  this session.
- The meta-description backfiller's prompt names the v3 file in an attribution line that ships
  into the rendered prompt (`488_*.sql:215`); a sweep item.

## Step 5: the 141-prompt sweep (follow-on, method only)

Re-run the handoff's census with today's date; order the copy-producing prompts by
`audit_prompt_demonstrations.py`'s league table (writer, planner, component-creator,
grounded-explainer, brief-writer, meta-description-backfiller, content-creator blog prompts).
One at a time, the same loop: draft in the voice, test-render on real loop rows, detectors,
exact text to the owner, council gate, migration, `Council-Submitted:` trailer. The bundled
skill's `shared/prompt-audit.md` is a usable checklist for the reviewer-seat prompts too: it
separates prohibitions with provenance from style prohibitions, which is exactly the cut needed.

## Verification, end to end

1. Every candidate block renders through the harness on fixtures A, C, E, F, G with no `<no value>`
   and no silent empties; outputs recorded in NOTES with the date.
2. Scanners at 0 demonstrations and 0 em dashes on every new prompt text; the row's migration
   asserts it in-transaction and was induced to fail once.
3. Experiment outputs scored by the detectors; the owner's blind ranking recorded in
   `README_where_we_are.md` in his words; every figure `[MEASURED <date>]`.
4. 641: resubmitted round completes; owner read recorded; gate 1 re-verified at the pod; applied
   by hand; live row re-read (`SELECT … prompt_template`) shows the block and `sections_for_render`
   in `input_fields`; 443 and apis.uk told.
5. Row: applied; live row re-read; one framework rebuild; served page through the detectors,
   before/after pair dated.
6. Council: one submission per coherent change; `Council-Submitted:` if committing before a
   verdict; never `Council-Reviewed:` on a verdict not read.

## Risks, and what to cut if short

- **The old approval is void and the read must not happen on the old block.** The finetuning
  session was carrying C to the owner; Step 0's reply stops that. Any edit inside the INSERTED
  TEXT markers restarts the read; make one edit, once.
- **Same-row collision.** Only 641 touches the writer row in this plan; the cache-marker change
  waits.
- **Losing a guarantee in the row rewrite.** The mapping table plus detectors plus rollback are
  the control; a positive rewrite that drops "no invented numbers" is a regression.
- **Lexical detectors are not his ear.** The blind read is mandatory, not optional.
- **Three other owners** (apis.uk, finetuning, copy lane) and idle peer sessions: CONTRIBs may sit
  unread; the owner's direct word to this lane covers 641; for the row, get consent in writing
  or let the copy lane apply.
- **Chassis roll today (owner, ~09:00).** No orchestration dispatch within ~300 s of a chassis
  restart; the canary waits for the roll to settle, and gate 1 for 641 is re-verified at the pod
  after it.
- **Cut order:** v4 first (nothing depends on it); then shrink the experiment to four sections,
  n=2 (about $4), keeping the sonnet/fable P1 cells; then defer the framework canary, keeping the
  off-line detectors. Never cut the peer replies, the standing five, the test-render, or
  exact-text discipline.
