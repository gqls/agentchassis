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
