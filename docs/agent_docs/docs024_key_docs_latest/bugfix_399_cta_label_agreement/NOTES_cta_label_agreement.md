# NOTES — CTA label/destination agreement (bugs_open/399)

Append-only, newest at the bottom. The missteps are not an appendix — they are the point.

## 2026-08-26 — picked the bug up

`who-owns.py 399` said OWNED-or-recently-active, but that was only the filing commit `3e1b890f1`
(the `dartsonline_traffic` lane, 2026-08-25). No lane directory, no follow-up commits. Took it.

The nearest neighbour `bugfix_389_cta_relevance` (bug 391) IS live — its NOTES was written 45
seconds before I looked. Checked: it is a docs+SQL lane with **zero** `platform/` commits across all
16 of its commits, so there is no file collision, only a sequencing duty.

### The bug is still valid, and worse than filed

The evidence row is unchanged and was **re-minted at 2026-08-25 20:58:09Z — 17 minutes after the
bug was filed** — still carrying the contradiction. The defect re-mints faster than any sweep.

Two of the file's own figures had gone stale by addition, in the four hours since filing:
- 646 → **665** components carrying a `_target_title`;
- "1 of 3 discovery agents" → **1 of 5** (still only `completeness-discovery-agent`).

Both were dated in the file, which is the only reason the staleness was detectable. The
count-with-date rule earning its keep on someone else's text.

### The measurement that decided the whole design

The question the filing session could not answer: *does the writer actually receive the destination
title in its prompt?* If not, the fix is to feed it. If so, the writer is ignoring it and only a
check will do.

It receives it, **twice**:
- `content_components.input_schema` → `cta_text.llm_guidance` says *"the link destination is
  already fixed: write this CTA text FOR that destination"*;
- migration 476 armed `stamp_cta_destination_guidance` and 477 (2026-08-20 07:17Z) made it actually
  reach the prompt: **781 of 2,297** writer prompts over 3 days carry `Destination (fixed):`.

And of the pairs written **since** that pipe went live, **155 of 1,060 (14.6%)** still contradict
their destination `[MEASURED 2026-08-26]`.

**That number could have come out near zero and did not.** Prompt text is not a control.

⚠ The before/after split (23.0% → 14.6%) is **soft** and must not be quoted as the guidance's effect
size: the "before" bucket is only rows not rewritten since 2026-08-20 (n=135), i.e.
survivorship-biased. The post-pipe figure needs no comparison and is the one to cite.

## 2026-08-26 — MISSTEP 1: my first census over-reported, and I nearly quoted it

I compared label tokens to `_target_title` tokens and got **228 mismatches (19.1%)**. Then I sampled
22 rows before believing it, and several were **correct destinations**: "Calculate Overpayment
Savings" → *"What overpaying could save you"*; "Run the Scorecard Simulator" → *"Where you stand
before you apply"*; "Read the case studies behind these posts" → *"What we have built"*. The
`_target_title` is frequently a **marketing** title that shares no word with a perfectly good label.

Re-ran against the destination page's `name` + `title` + `nav_label` with 5-char stems: **186
(15.6%)**. Residual known false positive: my `length(w)>3` filter drops short tokens, so
`"Read what to do if you can't pay"` → `/cant-pay.html` still reads as a mismatch.

**The cheap check: sample the rows before quoting the count.** A census over a heuristic is a
hypothesis until you have read some of what it selected.

## 2026-08-26 — MISSTEP 2: I accepted the bug's framing, and it was a third predicate

I was going to build exactly what 399's candidate 1 says — compare the label's tokens to the
adjacent `_target_title`. Adversarial review caught it. That comparison is a **third** definition of
"misdirected" beside the detector's and the writers', which is precisely the re-drift **RFC_047 §9**
rejects by name, and it is the shape `bugfix_203/CALIBRATION_2026-08-11` measured brittle — nine
already-correct CTAs on gaswholesalers.com flipped to the wrong tool over a stray hyphen in
"Break-Even".

Worse, the bug's *remedy* half ("regenerate the label from the title") is actively harmful: it is
what `stampCTADestinationGuidance` already asks the writer to do, so it converts a mismatch into a
**lock** — moving rows out of the ~60 label-less bucket a ranking fix reaches into the ~20
label-locked bucket only an LLM copy pass clears (`bugs_open/391`).

**The cheap check: before adding a comparison, ask which existing consumer already asks this
question.** The answer was `BestLabelMatchForPage`, live, in the detector.

The filing session agreed and added the same error one level down: its "reuse the token reducer"
suggestion was written *to avoid* a second definition of "distinctive token" and would have created
a third predicate.

## 2026-08-26 — MISSTEP 3: I nearly gated the half of the estate that matters least

I had the seam at `RenderComponentAction` — the first point where the fresh label and the resolved
destination coexist. Adversarial review claimed `RerenderPageSectionsAction` bypasses it. I checked
rather than accepting: **true**, `rerender_page_sections_action.go:662` calls `RenderTemplate`
directly. A gate there would have been blind to the repair loop — which is the loop actually minting
the churn (**182** `misdirected_cta` item_keys have been filed more than once).

Both writers converge on `save_page_sections`, verified in the live agent definitions:
`page-build-handler → call_content_writer → save_sections` and
`page-rerender → rerender_sections → save_sections`.

**The cheap check: enumerate the writers of the table you are guarding.** A seam that looks
canonical may be bypassed by the repair path.

## 2026-08-26 — MISSTEP 4 (caught before writing code): the shared predicate nearly broke a live check

My plan put a `ClassifyLinkScope` guard **inside** `JudgeCTALabel`, on the reasoning that a `tel:`
destination is not a page link. `check_cta_nonpage.go:141` calls `ctaClassifyAnchor` with anchors it
has deliberately filtered **to** `LinkScopeMailto` — so the guard would have silently switched off
misdirect detection on every phone and email button on the estate. A live check broken without being
edited, by a change that reads as tidying.

Scope is the **caller's**, and the two callers are right to differ. Written into the predicate's
header as a ⚠, and my own caller does its own filtering.

## 2026-08-26 — the design, sized

Of 186 mismatched pairs, what could the existing matcher actually repair?

| the copy names | count |
|---|---|
| exactly one other page | **13** |
| two or more (RFC_047: refuse, never guess) | **78** |
| no page at all | **95** |

So an automatic repoint reaches **7%** and inherits `bugs_open/248`'s clobber. 173 of 186 are
undecidable by any matcher. **That inverts the usual reading: record-only is not a cautious first
step, it is very nearly the whole available action, and the record is the deliverable.**

## 2026-08-26 — MISSTEP 5: "exactly 2 steps" was wrong, and the estate's own guard caught the other one

My approved plan said the migration should assert **2** matched `save_page_sections` steps
(page-build-handler, page-rerender). A recursive census found **six** live agents with one, four of
them inside a loop `sub_workflow` where a top-level `{workflow,steps,save_sections}` path cannot
reach them: `pageflow-builder`, `page-rebuild`, `site-work-orchestrator` (+ `tool-recreation-handler`
at top level).

This matters more for an instrument than for a guard. A guard armed on half its writers is visibly
partial. **An instrument armed on half its writers reports a RATE that reads as fleet-wide and is
silently biased by whichever writers it missed** — and the rate is the entire deliverable here.

Separately, `TestFindingCodeScanEveryWriteIsRegistered` failed my build within a minute of my adding
`CTA_LABEL_MISMATCH`, telling me exactly what to do and why (`bugs_open/358`: `LINK_CONTEXT_UNAVAILABLE`
reached the live table past a source-side warning that could not see it). Registered as
`instrumented` — **not** `consumed`, which asserts a reader that does not exist.

## 2026-08-26 — mutation results

Six mutations run against the shipped code; all six killed a test:

| mutation | killed by |
|---|---|
| delete the `auditCTALabelAgreement` call from the save path | wiring scan |
| `JudgeCTALabel` returns `Agrees` unconditionally | the filed-defect test |
| delete the `ClassifyLinkScope` caller guard | the tel:-destination test |
| delete the `title == ""` scope test | the unresolved-destination test |
| count ambiguous findings as contradictions | the RFC_047 ambiguity test |
| re-inline the comparison instead of calling the shared predicate | the anti-drift scan |

The extraction proof: `check_misdirected_cta_test.go` and `cta_classify_anchor_test.go` pass
**unchanged**.

⚠ My first test fixture failed for a reason worth recording: I hand-wrote the component
`input_schema` as a flat field map, and production uses the v2 `{"fields": {...}}` wrapper, so
`DeriveCTAURLFields` returned nothing and the pass was **silently inert**. Every positive test
failed and told me. Fixture now copied from a live `content_components` row. *A fixture I compose
exercises the rule I imagined, not the one production uses.*

## 2026-08-26 — the council said REVISE, and two of its objections were real defects

Verdict `revise`, gated by `debug_historian`, corr `e9bda035`. Nine seats approved, four objected.
Submitting was cheaper than the defects it found — as it has been every time.

### MISSTEP 6 — I shipped an ordering constraint as PROSE, and a banner cannot hold a migration

`debug_historian`, **severity HIGH**: migration 643 armed a config key with no stated precondition
that the binary reading it had rolled. My PLAN said "image first, then the migration" and the
migration header said the same — **and the runner takes every pending file in the directory
regardless of what its comments say.** CLAUDE.md states this outright under migration practice: an
ordering-critical file cannot be held by a banner, it has to be named `_HOLD.sql`. I had read that
line and still shipped the banner version.

Fixed: `643_audit_cta_label_agreement_HOLD.sql`, with the discharge condition being a **pod probe**
(both the config key and the `CTA_LABEL_MISMATCH` literal present in `/proc/1/exe`, with a control
that must be absent) rather than a commit sha.

**The check: if a file's correctness depends on something else happening first, the constraint goes
in the FILENAME. Prose in the header is read by humans, and the thing that applies it is not one.**

### MISSTEP 7 — my "both writers converge" claim was TWO OF THREE, and the third is live

`bug_historian` asked whether another write path persists CTA `content_data` outside the six censused
steps. It does. `ApplySectionEditAction` (`section_editor_actions.go` — `updatePageComponent`,
`updatePageComponentSwap`) writes `page_components.content_data` directly, never through
`SavePageSectionsAction`.

I had checked the *save-step* census recursively and thought that was the writer census. It was not:
I enumerated everything reaching **one action**, then reported it as everything reaching **the
table**. Those are different questions and I answered the easier one.

`[MEASURED 2026-08-26]`: 144 `section_edit` items, 132 complete, newest **today** — live, not
dormant. CTA exposure **3 of 144**, which is why the limit is stated rather than closed by widening.

**The check: `grep -rn "UPDATE <table>\|INSERT INTO <table>"` before claiming coverage of a table.**
"Every path that reaches this action" is not "every writer of this table", and my own census
question hid the difference. Note the shape: I did the recursive census *correctly* and still got the
wrong answer, because the recursion was over the wrong root.

### MISSTEP 8 — I wrote a MUTATION CLAIM THAT WAS FALSE, in the test whose whole job is honesty

`guardian` and `bug_historian` both objected that "the pass can never fail a save" was **asserted and
not shown**. Correct. So I added a `recover()` and a test, and wrote the customary comment:
*"MUTATION: delete the recover. This test fails."*

**I then ran it, and the mutation SURVIVED.** With `params.DB` nil the function returns at its first
guard, so the nil logger I was relying on to panic was never touched — the test never reached the
code it claimed to pin, and deleting the recover left the suite green.

Fixed with a `sqlmock` DB so control genuinely reaches the covered region. And that rewrite exposed a
**second, real defect**: my recover handler logged through the same nil logger that had panicked, so
the handler panicked too — **a recover that re-raises contains nothing**, and the save it exists to
protect would have aborted anyway. Both mutations now kill; both were verified by running, not by
writing.

**The check: RUN the mutation before writing the sentence that says it fails.** The estate's lesson
is "a mutation that passes has usually hit a guard in series" — this one is worse and worth naming
separately: *the mutation never reached the code at all.* An unreached mutation and a redundant guard
both present as "still green".

### The rest of the round, and what it was worth

- `editquality`: the verify block counted the config key alone, so `jsonb_set`'s `create_missing`
  writing a **dead key at a path nothing reads** would still have verified green. Now the assertion
  requires the key to sit on a step whose `action` really is `save_page_sections`.
- `guardian`: arming six pipelines at once with no canary. Split into `643` (two primary writers) +
  `645` (the remaining four), with the rate explicitly unreadable between them.
- `debug_historian`: no `snapshot_agent()` pre-image before `jsonb_set` on live rows. Added.
- `architecture`: run the optional-key budget before adding a key to a shared action. **Answered
  from the counter's own source rather than by running it**: `optionalbudget.go`'s header says it
  counts `RegisterActionInputSpec` Optional lists and that **`ConfigKeys` is deliberately NOT
  counted** ("settings rather than input references"). `save_page_sections` registers no InputSpec at
  all, so its counted set is **zero** and a step-config bool cannot move it.
- `prior_art_librarian`: `LoadCTALabelUniverse` asserted to exist with no supporting context. It does
  — `platform/orchestration/datahelpers/cta_label_universe.go:99`, register LNK-036 — and the
  resubmission carries the citation. Fair objection about the submission, not the code.
- `guardian` also caught a **process** hole worth keeping: the comment-only `label_match.go` edit was
  disclosed in prose but excluded from the edit list because the gate refuses comment-only sketches,
  so the council reviewed a plan that was not the whole diff. It now rides as a listed edit with
  surrounding context.
