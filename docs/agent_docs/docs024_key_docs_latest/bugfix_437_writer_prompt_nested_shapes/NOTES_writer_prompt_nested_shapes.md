# NOTES — bugs_open/437, writer prompt nested item shapes

Append-only, newest at the bottom. The missteps are the point.

## 2026-09-03 — session start, and the thing I nearly got wrong first

Session opened named `bugs_open/437`, with an unsolicited FYI from the
`portfolio_positioning` lane: advertise.co.uk has two `needs_page` items dead on this
defect, and would I say whether a fresh mint or a retry is needed once a fix lands.

The bug file offered two candidate mechanisms — "either the schema changed under the
writer, or the writer's prompt never learned the nested shape" — and marked the whole
section *"narrowed, not proven to the line"*.

**What I nearly did:** treat the 119-failure census as evidence about the writer. Every
artefact points that way. The error names the model's output (`got string`), the census
counts writer failures, the component's schema reads correctly when you open it, and the
adjacent closed bug (260) had already split "the writer-output half" off to another lane.
The obvious next step was to go and look at writer reliability or at coercion.

**What I did instead, and it was decisive in one query:** read the prompt the model was
actually sent. `llm_call_log` keeps replies verbatim (it is the training corpus), so the
instruction and the obedient answer sit in the SAME ROW —
`34f25815-42d3-4057-b42a-b8b42189ae7e`, 2026-09-02 19:07Z, advertise.co.uk. Prompt line
234:

```
"steps": [{ "body": "...", "branches": "...", "marker": "...", "note": "...", "title": "..." }]
```

The prompt declared `branches` a **string**. The reply obeyed. Neither candidate in the
bug file was "the prompt is generated from a lossy projection and states the wrong type",
which is what it is: `extractArrayItemFields` (`plan_sections_action.go:3277`) returns
`[]string`, so a nested collection flattens to a bare name and the exemplar renders it as
a scalar. The type gate then refused correctly, 119 times, deterministically — which is
why there were no lucky passes, a fact the census made visible and nobody had asked about.

**The lesson, and it generalises past this bug:** a correct guard downstream of a wrong
instruction makes a bad instruction look like bad output. The instruction is a *rendered*
artefact; it cannot be read in the schema, the component, or the code.

## 2026-09-03 — measurement: the census that would have under-reported

First blast-radius census only walked the JSON-Schema `items.properties` dialect and
returned 1 row. Correct, and incomplete: the library's majority dialect is flat
example-value `items` (values are type NAMES), plus `item_schema`. Re-ran across all
three; both others returned **0**, so the answer stands at **1** — but it stands for a
reason now rather than by luck. Recorded dated, per the counts rule: `[MEASURED
2026-09-03]`.

## 2026-09-03 — MISSTEP: my migration guard's arithmetic was wrong, and it would have refused my own splice

I wrote 724's balance guard asserting `{{if ` rises by **+2** — one for the new
`{{if .item_notes}}`, one for the new `{{if $f.value_shape}}`.

**It is +1.** The second edit does not add an `{{if `; it converts the existing
`{{if $f.item_fields}}` into `{{else if $f.item_fields}}`, consuming one. I reasoned about
what the replacement text CONTAINS and never about what the anchor it replaces stops
containing.

Caught by rehearsing the splice as string algebra before running any SQL — the rehearsal
printed `'{{if ' delta +1 expected +2 MISMATCH` beside five checks that passed. Cost:
nothing, because it was caught. Had it not been, the migration would have RAISEd on a
correct change while applying to a live row two other lanes also edit, and the natural
reaction to a guard firing is to doubt the splice rather than the guard.

Now practice, in the RUNBOOK and in `WRONG_CALLS.md`: **derive a diff-guard's expected
numbers from a rehearsal, never from reading the replacement.** An assertion about a diff
is a claim about what LEAVES as well as what arrives.

Second rehearsal (the real SQL in a rolled-back transaction) then proved the guards, the
verify block and the `_ROLLBACK` round trip — template md5
`7c7f1ffe9273e94f9952ab4e6f5205d9` before and after. A smaller misstep inside it: I first
tried to assert the md5 with psql's `:'orig_md5'` inside a dollar-quoted `DO` block, where
psql variables are not interpolated. Compared client-side instead.

## 2026-09-03 — a premise I inherited and had to correct

`bugs_closed/260` §5 dismisses candidate 4 ("ask the writer to obey the schema") as the
weakest option, and I read that at first as "prompt work on this class is not worth
doing". That would have been the wrong inference and it nearly cost the real fix.

The distinction: 260's candidate 4 proposed *asking a model to be careful, with no check*.
This change *fixes a false statement in a generated prompt*, with the mechanical check
still in place and unchanged. Written up for the `copy_quality_two_stage` lane, whose
2026-08-12 ruling ("not achievable by instruction … must be a mechanical check plus a
repair step") is CONFIRMED by this case and sharpened by it: the check was necessary,
armed, and working — it caught all 119 — and it could never have told us the writer was
handed a different contract.

## 2026-09-03 — tests: all six passed first time, so I mutated them

Passing on the first run is not evidence. Three mutations, each caught by a different
test: rendering a nested array as a scalar (reproduced the production error string
verbatim through `ContentTypeViolations`), removing the required-suppression, and removing
the structured-only gate (which would have churned every component's prompt). The frozen
pre-437 exemplar is kept in the actions test as a permanent mutation control.

One property I proved rather than assumed, and it is the one the deploy rests on:
`{{if $f.value_shape}}` on a spec map lacking the key is **falsy**, so an un-upgraded
chassis renders the new template byte-identically. ⚠ That path runs under text/template's
DEFAULT missingkey (`invalid`) — `RenderPromptTemplate` sets no `Option()` — not
`missingkey=zero` as I first assumed from the render paths elsewhere in the codebase. The
`{{if}}` behaviour is the same either way, but a BARE print of an absent key would emit a
literal `<no value>` into the prompt, so both new keys live only inside their guards and
the test asserts on that string.

## 2026-09-03 — shipped state at end of session

- `a0044e73b` — Go + tests + migration + livespec + PBP-052 + bug file. **Go INERT until
  the next chassis roll**; cluster on v1.0.1356, makefile staged at v1.0.1357 which does
  NOT carry this commit.
- `f88789e37` — gofmt + the register entry naming its sha (both raised by the pre-commit
  pattern check; the other two flags on that commit were checked and are not this task's).
- Migration **724 APPLIED and verified at the live row** 2026-09-03 09:44:42Z, recorded
  `--record-only`. All four declared counts hold, including the flat-arm survival check.
- Council corr `6de0f6f2-4f37-492a-9cbd-1ae886311a9b`, submitted before the commit
  (`Council-Submitted:` trailer). ⚠ **Submitted while a v1.0.1357 roll was pending** — a
  roll kills an in-flight council run. If it died, resubmit with `RESUBMIT_CORR=` that
  same correlation so the trailer still resolves at 098 report time.
- Open and NOT done by this work: 437 candidates 2 (repair path for items already branded
  terminal) and 3 (nothing escalates an active, linked, never-built page). The mixed
  deploy state (new template + old chassis) is live right now and is proven safe by test,
  but had NOT yet been observed at the artefact when this was written — no
  page-content-writer call had run in the ~20 minutes since the migration. That is the
  first thing to check next session.

## 2026-09-03 — council round 1: REVISE, and all three gating objections were MINE, in the submission rather than the code

Verdict on corr `6de0f6f2` at 09:56Z: **revise**, `decided_by: gating objection from
editquality`. `bug_historian` approved. Four seats abstained.

**editquality's three objections were all artefacts of my abbreviated SKETCHES**, and it
was right to gate on them, because a reviewer can only judge what I showed:

1. **HIGH.** My sketch wrote 724's `repl_A` as
   `$ra724$...anchor_A...{{if .item_notes}}…$ra724$` — a **placeholder** where the anchor's
   repeated text belongs. Read as the deployable artefact, that migration would delete the
   anchor and splice the literal string `...anchor_A...` into the live prompt. The seat
   also worked out that my guards check anchor COUNTS and a self-consistent length delta,
   so **they would not catch it** — which is a fair criticism of the guard design, not just
   of the sketch.
2. **MEDIUM.** I elided `structuredPropNote`'s object arm as `// ... object arm`, so it
   read as unimplemented.
3. **MEDIUM.** Could not confirm the reused `datahelpers` helpers exist.

**The committed code was correct on all three**, and for (1) the live row proves it: after
applying, `Each item is an object with exactly these fields:` is present and
`...anchor_A...` appears **zero** times. (2) is implemented and pinned by
`TestStructuredItemShape_NestedObjectProperty`. (3) they all exist; the package builds and
vets clean.

**The misstep is mine and the runbook had already warned me:** *"reviewers judge the
sketch"* and *"the sketch must be code, and it must show the part under objection."* I
abbreviated to stay inside the 32KB plan budget and it cost a round. The sharper version of
that lesson, now in the RUNBOOK: **a placeholder in a sketch does not read as an
abbreviation, it reads as a defect in the deployable artefact** — and where the artefact is
a migration against a live row, it reads as the worst kind. Resubmitted round 2 with every
sketch verbatim from the committed file (31,991 bytes, just inside the cap), on
`RESUBMIT_CORR=6de0f6f2` so the trail accumulates and the existing commit trailer still
resolves.

**The genuinely useful finding came from the seat that APPROVED.** `bug_historian` noted
that my omission advice ("or use `[]`") rested on `IsEmptyContentValue`'s live precedent,
which is a nested empty **STRING** — so `[]` at a nested position was reasoned **by
analogy, not measured**. That is precisely this estate's documented failure mode and I had
walked into it. Now measured (`53b2f46af`): all five spellings the note can produce — empty
array, absent, explicit nil, empty string, whitespace — are driven through the real
`ContentTypeViolations` at the nested position and all five pass, with
nested-string-branches as the control that must still fail. **Had any failed, the prompt
would have been recommending a shape the gate rejects — manufacturing this bug's own
failure on exactly the pages where the writer obeyed most carefully.** The answer was
favourable, but it was not knowable without the test.

Its second advisory: does closed bug 044 (`plan_sections` defers empty-schema components by
name heuristic) constrain the new branch? **Checked: disjoint.** 044 concerns components
whose `input_schema` is EMPTY; this code runs only inside the per-field loop on the
`source=='llm'` arm, which such a schema never reaches. No code change needed.

**A residual worth carrying, from objection (1)'s deeper point:** an anchored-`replace()`
guard built from occurrence counts and a length delta **cannot detect a replacement that
fails to preserve its anchor's own text** — both numbers stay self-consistent. 724 happens
to be covered, because its verify block asserts the flat arm and the field-list sentence
survive, and the applied live row confirms it. That was fortunate rather than designed. The
next anchored-replace migration on this estate should assert the anchor text is still
present after the splice.

## 2026-09-03 — the mixed state, observed at the artefact; and the control found somebody else's bug

The background poll finally caught writer calls under the new template (first at 10:35:54Z,
51 minutes after the migration — the writer is busy but bursty).

**Result, and it was alarming for about ninety seconds:** 2 of the first 3 prompts contained
`<no value>` — the exact string my own test asserts must never appear, and the thing my
whole deploy-order argument rests on not happening.

**The control settled it decisively, and it is the reason to always have one.**
`[MEASURED 2026-09-03]` `<no value>` appears in **420 of 643** page-content-writer prompts
from BEFORE migration 724 (65%), against 3 of 5 after (60%). It is long-standing, unrelated
to this change, and if anything marginally less frequent after. Had I checked only the
post-migration rows — which is the natural thing to do when verifying your own change — I
would have concluded I had broken the writer prompt and rolled back a correct migration.

**What it actually is:** `Location: {{.reviewed_brief.headquarters}}` rendering into the
block headed *"Official Contact Information (USE ONLY THESE - DO NOT INVENT)"*. So the
writer is told the business's location is the literal string `<no value>`, inside the one
block it is instructed to treat as authoritative — structurally `bugs_open/387`'s
stand-in-token class, except manufactured by the renderer rather than authored by a human.
Live damage today is **zero** (0 of 3,228 `page_components` carry it, against a writer that
ran 648 times today), so the writers are declining to copy it — a behaviour we rely on and
never asked for.

**Contributed into `bugs_open/453`** (filed this morning by the `apis_uk_bees_homepage`
lane; `who-owns.py` names them, so contribute rather than compete) rather than filed as a
new bug. The contribution's substance is a correction to their fix candidate 1: their lint
diffs template ROOT identifiers against `input_fields`, and **this instance's root
(`reviewed_brief`) IS in `input_fields`** — it is the SUB-FIELD that is absent, in the data,
per site. So there is a third failure shape their candidate cannot see, and it is the
highest-volume one. The cheap cover already exists: `RenderPromptTemplate` detects and counts
`<no value>` at render time and then logs a Warn nobody reads — 260 §9b's pattern exactly.

**Nothing about 437 changed.** My two new directives are inside `{{if}}` guards and neither
prints a bare key; none of the three observed calls was a mechanism-flow page, so the nested
exemplar has still not been exercised in production. That remains owed after the roll.

## 2026-09-03 — two lanes ask whether the legacy `items` dialect is my defect: it is not, and the check strengthened my own claim

The `components` lane (routed via `bugs_open/427`, independently re-derived and relayed by
`bugsweep4`) reported that `mechanism-flow` still carries the LEGACY JSON-Schema `items`
dialect — the `bugs_open/240` shape — and asked, without proposing a mechanism, whether that
is the same defect as 437 or two in one place.

**Settled: two, in the same place, fixed sequentially.**

- **240 was the DIALECT being misread.** A JSON-Schema `items` block read naively as a flat
  name→type map yields the JSON-Schema KEYWORDS, so `mechanism-flow` once shipped `steps`
  keyed `properties`/`required`/`type` and rendered empty. Fixed: `extractArrayItemFields`
  discriminates on `items["properties"]` being a map, and names this very component in its
  own doc comment.
- **437 is one layer past it.** The dialect is read CORRECTLY; the per-item NAMES come out
  right. What is lost is the nested TYPE.

**The proof they are distinct was already in my evidence and I had not noticed it as
proof:** the failing prompt listed `body, branches, marker, note, title` — the real per-item
names. Had the dialect been mishandled, it would have listed `properties, required, type`.
So dialect handling was working correctly throughout all 119 failures. My fix neither
depends on it nor changes it, and re-declares no component schema.

**The valuable part was the cross-check, and it could have gone against me.** Their census
named a population I did not choose: **5** components carry the legacy dialect (`checklist`,
`comparison-table`, `evidence-timeseries`, `mechanism-flow`, `period-calendar`). If any of
the other four had a nested structured item property, my "exactly 1 qualifying component"
blast-radius claim — which the council submission, the register entry and the commit message
all rest on — would have been wrong and the fix would have silently under-covered. Re-ran
the structured-property test over exactly those five: **1 of 21** item properties is
structured (`mechanism-flow.steps.branches`, `array`); the other 20 are `string`/`number`.
Same answer, someone else's population. That is a better form of the claim than my own sweep
alone and it is now recorded in PBP-052 with its provenance.

**Also inherited, a query trap worth more than the finding** (from `bugsweep4`, having cost
two lanes real time): `jsonb_typeof(def->'items'->'properties') = 'object'` is **NULL** on a
flat component, so `WHERE NOT (…)` discards every flat row under three-valued logic and
returns 43 flat components as ZERO — a wrong answer shaped exactly like a finding. My own
censuses happen to be safe because they use a POSITIVE test rather than a negation, but that
was luck rather than care. Written into the RUNBOOK beside the census it protects.

Their stated negative is worth preserving too: `bugsweep4` checked the dialect against
`bugs_open/361`'s render-check findings and reports **no intersection** — none of the 18
regressions sits on a legacy-dialect component, and the 3 of 460 unbaselined that do are
field-level (`.section_title empty_heading`), not per-item. Recorded so nobody later infers
a link that both lanes explicitly looked for and did not find.

## 2026-09-03 post-roll — THE FIX IS LIVE AND DOES NOT WORK, and I could not explain it

`v1.0.1358` rolled at 12:06Z carrying `a0044e73b`. Builds still fail with the identical
error; the writer prompt still shows the old flat exemplar. **Candidate 1 is not closed.**

**Every "is it really deployed?" answer came back YES, with controls** — ancestry against
the image revision label (negative control behaved), the helper's literal and the
`value_shape` struct tag both in `/proc/1/exe`, the call present twice in the built
revision's own tree at the single append site, no service skew across the three chassis
deployments, migration 724 intact in the live row, and the helper returning the correct
skeleton when fed the exact live schema bytes. And yet the emitted spec for mechanism-flow's
`steps` carries `item_fields` and no `value_shape`, which under `omitempty` means the helper
returned empty at runtime on a field where it demonstrably does not.

**Wrong turns, recorded because they cost real time:**

- **I read the wrong collected_data key for an hour.** The step's `output_field` is
  **`section_plan`**; I kept querying `plan_sections`. Both keys exist and they are not the
  same object. (The answer turned out identical either way, which is luck, not vindication.)
- **I concluded "plan ran on the new binary" from the WRITER CALL time.** `plan_sections`
  runs earlier in the same orchestration, so a writer call after the roll proves nothing
  about when the plan was built. I had to go back for the orchestration `created_at`. It
  happened to hold, but the inference was unsound when I made it.
- **`collected_data::text LIKE '%value_shape%'` returned TRUE and I briefly believed it.**
  It was matching the TEMPLATE TEXT (`{{if $f.value_shape}}`) echoed in the config, not
  data. A substring test over a blob that contains your own needle as source code is not a
  measurement — exactly the class this lane has been writing landmines about all day, and I
  walked into it on my own change.
- **I searched `$.**.llm_field_specs[*] ? (@.name == "steps")` and got 12 specs with zero
  shapes, and nearly reported that as the finding.** `steps` is a common field name; those
  were mostly other components, correctly emitting nothing. The question only became
  meaningful once filtered to `function = 'mechanism-flow'`.

**The one concrete anomaly, and the top hypothesis:** the component payload carried in the
plan serialises `component.input_schema` as a **JSON STRING**, not an object. If the loader
hands `plan_sections` a differently-shaped schema than the raw DB row,
`extractArrayItemFields` still succeeds (it needs only `items.properties`) while
`StructuredItemShape` returns early — because its first guard,
`declaresArray(fieldDef["type"])`, is **stricter than `extractArrayItemFields`' entry
condition**. That asymmetry is mine and is worth removing whether or not it is the cause:
two readers of the same field disagreeing about whether they can read it is the shape of
this bug one level up.

**Next session: instrument, do not query.** A temporary Warn in the `source == "llm"` branch
printing the runtime `%T` and value of `fieldDef["type"]` settles it in one run. Full
elimination list and ranked hypotheses: `bugs_open/437` §POST-ROLL and
`HANDOFF_2026-09-03_continue_here.md`.

**Docs corrected rather than left optimistic:** `bugs_open/437` candidate 1 and register
PBP-052 both now lead with the post-roll failure, because as written they would have told
the next reader this mechanism works.

## 2026-09-03 12:55Z — I probed the wrong pod, and the estate's own landmine says so

Asked to verify a fresh chassis build. **There isn't one:** makefile `IMAGE_TAG`, the
`agent-chassis` deployment spec and the running pods are all `v1.0.1358` from 12:06Z, rollout
complete, no newer local image. Reported rather than papered over.

But looking for it surfaced what I had missed all along. `kubectl get pods --sort-by=
.status.startTime` shows **per-agent pods** — `agent-page-build-handler-*`,
`agent-page-content-writer-*`, `agent-site-review-agent-*` — each running the chassis image,
spawned per job, a fresh crop at 12:49–12:52Z. **`plan_sections` runs in those, not in the
`agent-chassis` deployment pod I spent the morning probing.**

So elimination #2 of my own handoff ("not missing from the binary") was verified on a pod
that never ran the code. Re-probed on `agent-page-build-handler-2c993dc2-cvfl5`: the fix
literal IS present, control behaves. Today's agent pods carry the fix.

**Which reframes everything.** The five failing orchestrations ran 12:07:57–12:20:25Z, within
~14 minutes of the roll. Agent pods outlive a single job, so those executions plausibly ran
on pods created before 12:06Z on the previous image — reconciling every observation with a
fix that is simply correct. Unprovable retroactively (the pods are gone) and I am not
asserting it, but it displaces the `comp.InputSchema` hypothesis at the top of the list.
Checked and excluded: neither agent is image-pinned (`pin_image_tag` unset,
`agent_image.go:201`) and both rows read `v1.0.1358`.

**The lesson is one this estate already wrote down and I still walked into:** *"`-l
app=<subsystem>` may be the WRONG SERVICE (one image, every label); before believing a clean
grep, ask which pod could have produced the line."* I asked which pod ran the code only after
exhausting every other explanation — having first written a landmine, this same day, about
reading the artefact rather than the thing that shares its name. **Probe the pod that does
the work.**

Status: the fix is **un-exercised, not disproven**. No mechanism-flow write since the fresh
pods. A background watcher is armed on the next such prompt; if it shows the nested exemplar
the lane closes, and if it does not — on a pod proven to carry the fix — the
`comp.InputSchema` hypothesis returns and the next step is instrumentation.
