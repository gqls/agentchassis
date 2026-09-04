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

## 2026-09-03 (~16:00Z) — the owner's three decisions, 641 edited to A4, apply rehearsed green

**His answers, verbatim:** *"1: drop it and use fuller sentences/phrases 2: keep the prohibitions but
perhaps make the whole prompt read more positive as a whole. 3: enable cache and also test against Grok
now that it has some credit."*

**Decision 1 — the block.** 641's INSERTED TEXT is now candidate A4: the subject printed verbatim on its
own line, C's sibling list unchanged. The "You'll want to know ___" frame is gone. **This is a data
decision, not a taste one**, and the harness is why: the frame is a fixed sentence with a hole, so only a
short lower-case fragment fits, and both the real gamedesign planner subjects and his own example line
render broken (the latter with a doubled full stop no rule catches). Subjects are now authored as the
section's opening line, in fuller sentences or phrases. A phrasing spec is owed to finetuning and apis.uk
before either authors data; my earlier planner-nudge draft asked for the OPPOSITE (short lower-case
clauses) and was retracted to apis.uk before they cut it - confirmed nothing had landed.

**Decision 2 — the house voice row: keep every prohibition, change the register only.** So the trade I
offered (drop a prohibition where a detector already enforces it) is DECLINED, and that simplifies the
work: no rule leaves the prompt, meaning-preservation is a form question, and nothing depends on the
detectors holding. "The whole prompt read more positive as a whole" is taken as the row first, then the
writer prompt's own rule block by the same treatment.

**Decision 3 — caching is now authorised** (its own migration, after 641, on the same row) **and Grok
joins the model arm.** Credentials checked at the pod, value never printed: `XAI_API_KEY` and
`GROK_API_KEY` both SET, `grok-4-1-fast` is the model the platform already uses via the xAI Responses API
(`feed_actions.go:741-758`). The 08-31 arm was blocked on credits; he says there is credit now.
⚠ The chassis image has **no curl and no python3, only BusyBox wget**, which drops 4xx bodies - the
documented reason the 08-31 Grok 403 was unreadable. Find a pod with curl, or read the status line
explicitly, before calling any failure a credit failure.

**641 edited and rehearsed.** Only the two copies of the text and the header prose describing them changed;
the pre-flight already-applied probe (keys on `{{if .current_section.subject}}`, which A4 keeps), the
version-shadow guard, the both-halves verify and the em-dash equality census are byte-for-byte as apis.uk
wrote them. Rehearsed under BEGIN/ROLLBACK against the live row: `NOTICE: 641 applied … em-dash census 10
(unchanged)`, then ROLLBACK, then a post-rollback control showing the live row still clean (block position
0, `sections_for_render` absent).

> **[MEASURED 2026-09-03 ~16:00Z] The live writer prompt moved under me during this session: 14,111 chars
> at ~12:30Z, 14,621 at 16:00Z, and its em-dash count went 9 to 10.** Another lane edited the same row
> while this lane was drafting against it. Nothing here broke, because 641's census asserts pre/post
> EQUALITY rather than a literal - the apis.uk lane replaced a literal 5 with equality on 2026-09-02 for
> exactly this reason and it has now paid off twice. This is also the empirical form of their warning:
> **record no absolute positions or lengths of a shared prompt**; my own PLAN quotes 14,111 and is already
> stale, corrected here rather than there.

**Also confirmed for the finetuning lane:** their Stage B dispatch gate keys on the literal
`current_section.subject`, which A4 keeps, so the gate still fits.

**Misstep, mine, caught by running the check:** the RUNBOOK's two-copies byte-compare printed DIVERGED on
two byte-identical copies, because the comment copy is reconstructed line by line and gains a trailing
newline. Fixed to `rstrip` both sides, with the harness template as the authoritative third comparison.
A checker that false-alarms is worse than no checker, because it gets ignored.

## 2026-09-03 (~19:35Z) — 641 IS LIVE. Gate 2 cleared, gate 1 re-probed, three lanes landed in the right order by luck as much as planning

**The owner's read (gate 2).** He was shown the exact INSERTED TEXT bytes plus one filled-in render and
answered, in full: *"yes"*. Recorded verbatim, with the bytes and the round-1 diff, in the apis.uk lane's
`NOTES_apis_uk_bees_homepage.md` (their file, appended per their convention) and named in 641's APPLIED line.

**Gate 1, re-probed immediately before applying** because there was a roll today. Both replicas, three-way:
`section_subjects` 3 (the capability), `section_facts` 3 (positive control), `zzz_absent_zzz` 0 (absent
control). The absent control is the part that matters: without it a probe returning a number proves only
that grep ran.

> **MISSTEP, mine:** I first tried the documented `logs -l app=agent-chassis --tail=3000 | grep -m1 'build
> provenance'` route. On this service that streamed megabytes of council-gate orchestration JSON and **timed
> out at 2 minutes without ever printing a verdict**. CLAUDE.md already says the startup line scrolls; what
> it does not say is that on a busy chassis the *attempt* is itself expensive. **Go straight to the binary
> probe, which has no shelf life, and bound it with `timeout`.** Cheap check that would have saved it: the
> provenance line is a STARTUP line and these pods have been up for hours, so it could not have been in range.

**Applied**, output `NOTICE: 641 applied: block + input_fields in one transaction; em-dash census 10
(unchanged)`. **Live row verified after COMMIT**, six ways: block present, old frame absent, sibling range
present, block precedes the Verified Facts block, `input_fields` carries `sections_for_render`, em dashes 10,
template 14,914 chars.

**The ordering across three lanes, which came out right and should not be assumed next time.** Within about
an hour: the finetuning lane re-authored and applied all three `section_subjects` arrays to the new spec; the
apis.uk lane cut and applied the planner nudge (migration 762) carrying the spec's rule-17 text verbatim,
live but inert to the writer; then this block applied. So the first build after the apply reads new-register
subjects through a new-register block, with **no intermediate state in which the block prints old-register
text as a page's opening line**. Nobody sequenced that centrally. It worked because each lane held its half
until the spec existed, which is the argument for writing the spec before any of the three edits rather than
after the first.

**Council round 2 (corr 6c92d154) was still open at apply time.** Applied ahead of it on the owner's approval
of the exact words, which is the binding gate; commit trailer is `Council-Submitted:`. **Still owed: read the
verdict and act on a REVISE** - the change is LIVE now, not merely on the branch, so a revision is a new
migration rather than an edit.

**Next, in order:** (1) read the round-2 verdict; (2) the house voice row rewrite - his decision is KEEP every
prohibition, change the register only, so this is a form rewrite of 16 paragraphs with no rule removed;
(3) the cache-breakpoint migration on the same writer row, now unblocked since 641 has landed; (4) the model
arm including the Grok arm he asked for. The first real evidence for (2) will be the opening lines Stage B
produces, which the finetuning lane is sending.

## 2026-09-03 (~20:00Z) — 641 council round 2 APPROVED, all three advisories adjudicated at the artefact

Verdict `complete_approved` 19:30:57Z, *"approved with 3 advisory objection(s) — none high-severity"*.
Read in full and dispositioned; **two of the three were true of my SUBMISSION and false of the FILE**,
which is a lesson about how I wrote the submission, not a defect in the change.

- **editquality (medium)** — the sketch diffed only the `E'...'` string while the rationale claimed both
  copies changed, so a stale header copy would mean the owner reads the old frame at gate 2.
  **Already satisfied in the artefact**: both copies were changed together and byte-compared after
  decoding, plus a third comparison against the harness template. Their concern was the right one to
  raise from what I gave them; I showed one hunk and asserted two.
- **bug_historian (medium)** — `{{.current_page.title}}` is the field family in the OPEN `bugs_open/085`
  (render data advertises `current_page` and supplies empty on some paths), so the unchanged sibling-list
  line could render " also covers, each in its own section:" with a blank title, silently. They asked for
  a live spot-check rather than a fixture replay. **Done, post-apply, first-hand:** all six
  `generate_content` prompts of orchestration `89059f29` render the title as `The Technical Details |
  FineTuning`, non-empty, block 596 chars each. `[MEASURED 2026-09-03 ~20:00Z]` **Scope stated honestly:
  one orchestration, one site, six iterations — this shows 085 does not reach this path on this build,
  not that it cannot fleet-wide.**
- **debug_historian (medium + low)** — the `replace()` anchor is a bare needle with no occurrence count,
  on a row documented as concurrently edited this session. **Premise is false and the file answers it:**
  641 line 179-182 counts the anchor inside the same `DO` block at apply time and RAISEs unless the count
  is exactly 1, so a drifted or duplicated anchor aborts rather than multiply-inserting. Again my
  submission's omission: I described the guards as "unchanged" without showing them.

**Transferable lesson, and it cost two seats' attention:** a submission that says "the existing guards
still hold" without quoting them invites objections the file already answers. **Quote the guard, do not
cite it.** Cheap check that would have avoided both: before submitting, grep the file for each guard the
plan claims is inherited and paste the line into `grounded_in`.

Commit trailer stays `Council-Submitted:` (098 credits it now the correlation is approved); forward-only
forbids an amend and none is needed.

## 2026-09-03 (~20:05Z) — house voice v3 drafted

`DRAFT_2026-09-03_voice_style_block_v3.md`. Form rewrite per the owner's decision: **every rule and every
prohibition kept**, including all 15 cut-list items and the em-dash ban; what changes is which half of each
sentence leads. 16 paragraphs in and out, 5,862 -> 6,173 chars, 0 em dashes, 0 exclamation marks, all 16
rule carriers and all 15 cut-list words asserted present programmatically before the file was written.
Two things deliberately NOT changed and flagged for him: the two finance-specific worked examples (removing
them is a content decision and a documented exemplar-lift risk, not a form one), and the short first line
(it renders as a heading). Owner read, then the blind before/after, then the migration.

## 2026-09-03 (~20:30Z) — the diagnosis run came back UNVERIFIABLE, and the clean control does not exist in the data yet

**Verdict, verbatim from the work item (`2b9733ae`, status complete):** *"Diagnosis NOT confirmed
(stopped: iteration-cap). Best-effort trail attached for a human; no fix proposed."* Status
`UNVERIFIABLE`, `is_fix: false`.

**That is neither CONFIRMED nor REFUTED, and I will not report it as either.** It stopped at the cap.
Worth recording why it plausibly could not settle it: its own evidence bundle's "in-scope code" section
retrieved `truncate` from the banana image provider, `truncate` from the tools-api gripper store, and
`applyAddToPage` — none of which is on the path named in the symptom. The evidence that decides this
lives in `llm_call_log.prompt_rendered` and in one `agent_definitions` row, not in a code search.
**A symptom that points at a TABLE may be a poor fit for a loop whose retrieval is code-shaped.**

**What is established first-hand and is not in doubt** (all `[MEASURED 2026-09-03]`): the same
3,295-char brief, identical md5 `b4fd73f0…`, is rendered into all six section prompts under
"## Rewrite Guidance (IMPORTANT: incorporate this into the content)"; the code chain from
`aliasGuidanceIntoSuggestion` through page-build-handler's `rewrite_guidance` mapping to the writer's
loop; and that "## What To Write" identifies the section by `{{.current_section.name}}`, the COMPONENT
type, so three `generic-text-block` slots read the identical instruction there.

**What is NOT established: that removing or relabelling the brief would stop the convergence.**

**The sharpest circumstantial evidence, and its n.** Two pages built today converged, and both converged
on their brief's **section (2)** specifically: technical-details on "which model and its licence",
your-own-model on "how it works, in three steps". Different sites' briefs, different topics, same
POSITION in the brief. `[n=2, MEASURED 2026-09-03]` Small, and it points at the brief rather than at the
sections. your-own-model's brief was written 2026-08-24 and NOT touched today, so the effect is not an
artefact of the rewritten technical-details brief.

**The observational control does not exist yet.** Census of the last 4 hours: every orchestration
carrying the new subject block (`89059f29`, `fadecb26`, `ac71fc02`) also carries the page brief, 6 of 6
prompts each. Over 3 days there are 643 orchestrations WITHOUT a brief and 207 with, but the ones
without predate the subject block, so they lack BOTH inputs and cannot separate the two causes.
**The comparison that decides it is: subjects present, brief absent — and no such build has happened.**

**Two ways to get it, both real:**
1. **Prompt replay A/B, off-line.** Take `89059f29` iteration 2's exact `prompt_rendered` (which
   converged), run it as-is and with the Rewrite Guidance block deleted, n=2 each. Newly possible: the
   images have no curl and no python3, but **BusyBox wget supports `--header`/`--post-file`**, recipe now
   in the RUNBOOK. About four calls, pennies. Decides causation directly.
2. **Build one page through the framework with subjects and no page-level brief.** Slower, but it is the
   estate's own path and produces a real artefact rather than a replay.

**Not doing either without the owner**, because both spend money and (2) writes a live page. Recommending
(1) first: it is cheaper, it is reversible, and if it comes back showing the section opening on its own
subject, the fix is scoped before anyone edits a prompt.

## 2026-09-04 (~16:00Z) — fleet roll v1.0.1361 and this morning's credit outage: this lane is clear, and the roll sets a PRECONDITION on both pending experiments

**Checked, not assumed** (the comms session's numbers, re-run here):

- **Credit outage 11:21:11–11:56:49Z** killed 146 runs fleet-wide, 92 of them council-gate. **Nothing of
  this lane's**: council `6c92d154` reached `complete_approved` 2026-09-03 19:30:57Z and diagnosis
  `fae94be1` COMPLETED 2026-09-03 20:04:55Z, both a day before the window. Nothing was submitted today.
- **641 is intact and unedited since apply**: block present, `sections_for_render` present, template still
  **14,914 chars** — the same length as at apply, so no other lane has touched the writer row since.
- **A roll cannot ship or endanger a prompt change.** `make release` replaces binaries; `agent_definitions`
  is untouched. `service_binary_capabilities` is the wrong instrument for anything this lane changes — the
  live row is the fact. Worth stating because this lane's whole output is DB config.

> **PRECONDITION, and it binds both pending experiments.** v1.0.1361 carries **aiservice token-budget
> work** (`max_tokens.go`, `thinking_budget`, the `llm_budget` ladder) among 18 Go-touching commits since
> the running v1.0.1360. Both experiments this lane owes — the house voice before/after replay, and the
> model arm — measure **writer output**, which is exactly what a token-budget change can move.
> **Neither may straddle the restart.** Run both arms on one side of it, after the pods are up and past
> the ~300s dispatch-drop window, and record the running tag beside the numbers. A before/after taken
> across a binary change measures the binary as much as the prompt, and nothing in the output would say so.
