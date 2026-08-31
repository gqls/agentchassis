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

## 2026-08-26 — round 2 was REFUSED before review, and the reason is a real tension worth recording

The round-2 resubmission reached `complete_invalid` without a single seat running:

> `step persist_submission failed: … plan failed validation: edit 8: sketch is comment-only — a fix
> plan proposes changes, not observations; drop the edit or make it real`

**The client dry-run passed and the server refused.** The two checks are not the same: the client
tests whether *every non-blank line* is a comment; the server appears to test whether every **added**
line is. My sketch had diff context lines (a real `func` signature) but every `+` line was a comment,
because the change genuinely is comment-only.

**This is a real tension, not a formatting slip.** The council's `guardian` seat objected in round 1
that a comment-only edit rode along unlisted — *"the council is reviewing a sketch that is not the
full diff that runs — precisely the failure mode this gate exists to prevent"*. The gate then refuses
that same edit, by design. Both are right, and there is no submission that satisfies both.

Resolved by dropping the edit and **attesting it in the rationale in full**, naming the commit
(`00cf81437`) so any seat can read the diff. What I did **not** do is fabricate a code line to make
the sketch pass — that would have satisfied the gate by defeating its purpose, which is the shape of
error this whole lane exists to catch.

⚠ **Two costs of the refusal, worth knowing before anyone repeats it.** It cost a full dispatch round
(the run publishes, queues, and dies at `persist_submission`), and — more importantly — **a refused
submission leaves the trail with round 1's `revise` as its latest verdict**, so the correlation reads
as un-revised until a round actually completes. `DRY_RUN=1` does **not** catch this class, so treat a
comment-only edit as inadmissible from the start rather than trying to shape one past the gate.

## 2026-08-26 — APPROVED at round 3 (corr `e9bda035`), 12 of 15 seats

Rounds: **revise** (gated by `debug_historian`) → **refused before review** (`complete_invalid`,
comment-only edit) → **revise** (gated by `editquality`, on my truncated sketches) → **approved**.

Approving: `bug_historian`, `reuse_agent`, `guidelines`, `diagnosis_guardian`, `improvement_guardian`,
`compliance`, `render_guardian`, `debug_historian`, `constitution`, `mission`, `prior_art_librarian`,
`architecture`. `architecture` approving matters: it means this was ruled a **point fix**, not
architecture-scope, rather than my having assumed so.

**Three rounds, and only the FIRST was about the change.** Round 1 found three real defects in the
code. Rounds 2 and 3 were about my *submission* — truncated sketches, a missing test-file edit, an
inadmissible comment-only edit. That is a cheap way to spend a review budget and worth saying
plainly: the gate charges the same for a bad submission as for a bad change.

### Residual objections, checked rather than waved through

`editquality` kept three MEDIUMs, all of the form "this symbol is undefined, it will not compile".
**All three are pre-existing same-package helpers, and the package builds and tests green:**

| symbol | actually lives at |
|---|---|
| `ctaTargetTitleField` | `resolve_internal_links_action.go:634` (same package) |
| `pqStringArray` | `fixloop_digest_action.go:487` (same package) |
| `readSourceFile` | `render_sitemap_test.go:17` (same test package) |

They are artefacts of a seat reading one file in isolation — a real limit of sketch-based review, not
a defect. Recorded rather than dismissed, because "the reviewer was wrong" is the most dangerous
sentence to write without evidence, and the evidence here is `go build` plus `go test` plus
`verify-head-builds.sh` against HEAD.

### `guardian`'s medium — CHECKED, and it does not bite

> *"the estate has a live landmine that four agent types carry two ACTIVE rows under the same type.
> If any of the six affected types are among those four, the six-step census and the per-type UPDATE
> could both be silently wrong."*

Exactly the right question to ask of a `type IN (...)` UPDATE. Verified `[MEASURED 2026-08-26]`:

```
page-build-handler 1 · page-rerender 1 · page-rebuild 1 · pageflow-builder 1
site-work-orchestrator 1 · tool-recreation-handler 1
```

**All six carry exactly one active row**, so neither the census nor the UPDATEs can double-touch or
undercount. The migrations' `RAISE EXCEPTION` on `count <> 6` also fails loudly if that changes
before they are applied — which is the property that makes this safe to leave rather than re-engineer.

### `guardian`'s standing caution, accepted and NOT closed

> *"Unchanged tests prove the CASES THEY COVER are preserved, not that no live finding population
> shifts — RFC_047's ambiguity split is new surface the old tests never exercised."*

This is right and I have no counter-evidence. The extraction is behaviour-preserving *by
construction* (`ctaClassifyAnchor` maps the three verdicts back onto its original two returns), but
"the detector's live finding population is unchanged" is a claim about production that only
production can settle. **Added to the post-roll verification owed**: compare `misdirected_cta`
finding volume across the roll, and treat a shift as this change's until shown otherwise.

---

## 2026-08-31 — the canary passed, `645` discharged, and the instrument is now unbiased

### The demand the 08-26 session could not wait for

Anchored on the migration's own `applied_at` (`2026-08-26 22:17:08Z`), never on this session's start
— that anchoring error is the one the previous session nearly filed as a defect.

```
cta_saves_since_arming = 214        [MEASURED 2026-08-31 15:03Z]
```

`CTA_LABEL_MISMATCH` records, by producer [MEASURED 2026-08-31 15:04Z]:

| producer | records | contradicts | ambiguous | first | last |
|---|---|---|---|---|---|
| `page-build-handler` | 61 | 34 | 59 | 08-26 22:46:53Z | 08-31 13:08:48Z |
| `page-rerender` | 83 | 1 | 124 | 08-26 22:51:32Z | 08-31 14:59:50Z |

**Both armed producers fired.** That is the canary the guardian seat asked for, and it is the whole
of the decision rule in the 08-26 handoff §1. Applied `645`.

A sampled record is well-formed and is a real case of the *destination-KIND* blind spot the 391 lane
named: `"Browse all tools"` → `/tools/css-specificity-calculator/index.html`, verdict `no_opinion`,
`silence: ambiguous`. The reason code is doing its job — this is exactly the bucket that used to be
indistinguishable from "the judge had nothing to say".

### The canary's real job was pipeline health, not record presence — checked separately

Records firing proves the pass RUNS. It does not prove the two armed pipelines are HEALTHY, and the
guardian's risk was the second thing: *a new per-page DB read landing on every content-writing
pipeline at once*. So:

```
page-rerender      FAILED at current_step='save_sections' : 7
  of which OWNED_PAGE_GUARD : 7        of which anything else : 0
```

All seven are a **different guard firing correctly** — `OWNED_PAGE_GUARD` refusing to let a generic
section save clobber a tool/widget-owned page (`llm-cost-calculator`, `electric-vs-pneumatic-economics`).
Unrelated to this pass, and it appears only on `page-rerender` because that is the writer that
re-renders existing pages.

> ### ⚠ MISSTEP — I built a column that matched everything by construction
>
> My first pass at this asked for `count(*) FILTER (WHERE ... collected_data::text ILIKE
> '%audit_cta_label_agreement%')` and got **5 of 5, 7 of 7 — a perfect hit rate**, which I briefly
> read as "every failure touches the CTA seam". It cannot come out any other way: an armed run
> embeds its own step config in `collected_data`, so the filter matches **every armed run, failed or
> not**. It was measuring arming, not causation.
> The check that actually discriminated was reading the `error` text, where all seven said
> `OWNED_PAGE_GUARD`. Logged in `WRONG_CALLS.md`.

### ⚠ A before/after failure-rate comparison is NOT available from `orchestration_states`

`min(created_at)` in that table is **2026-08-30 14:53Z** — roughly a one-day retention window. Asking
it for the four days before arming returns **zero rows**, which reads as "no failures before" and is
really "no data before". Same rolling-window trap as `site_work_items`. Ask the failures what they
SAY; do not ask the table for a rate it cannot hold.

### `645` applied — the four remaining writers

Discharged exactly as `643` was: renamed out of `_HOLD` (content byte-identical to the approved file,
`git diff --cached` reports **R100**), applied out-of-band with `psql -f` because `--apply` takes
every pending file and other lanes have work in that directory, then recorded `--record-only`.

```
DO / UPDATE 2 / UPDATE 1 / UPDATE 1 / DO
NOTICE: audit_cta_label_agreement now armed on all 6 save_page_sections steps
COMMIT
645_audit_cta_label_agreement_remaining_writers.sql | 2026-08-31 15:09:38.146832+00 | record-only
```

Pre-condition guard passed (2 primary armed) and the verify passed at 6. **Verified independently of
the migration's own verify** — a migration that checks itself is the weaker half of the pair:

```
page-build-handler t · page-rerender t · page-rebuild t · pageflow-builder t
site-work-orchestrator t · tool-recreation-handler t
content-writer f · council-gate f      ← controls, and they are why the six t's mean anything
```

⚠ The 08-26 arming census had a **mixed** expected answer (2 t, 4 f) and that mixture was its own
control. After `645` the expected answer is all-true, so the mixture is gone — **carry two known-false
types in the query** or an all-true result is indistinguishable from a predicate that matches
anything.

### The number `645` was already taken, and that is tolerated

`645_design_critique_agent.sql` (a different lane) was applied 2026-08-26 14:21Z. The ledger keys on
`filename`, not on the number, and already carries **two `648`s** — so the collision is precedent, not
a defect. Renumbering would have broken every citation of "645" in `bugs_open/399`, the RUNBOOK,
LNK-040 and `643`'s own header, so the number stands.

### Binary probe, with both controls in the same exec

Fleet is on `v1.0.1349` (the pass shipped in `v1.0.1345`). Every agent type runs the **one**
`agent-chassis` image — there are no per-agent images — so the four newly-armed writers necessarily
run the same binary as the two that have been firing all week.

```
PRESENT  audit_cta_label_agreement      PRESENT  CTA_LABEL_MISMATCH
ABSENT   cta_label_agreement_NOT_A_REAL_SYMBOL   ABSENT  audit_cta_kind_agreement
```

### §6's owed check — `misdirected_cta` volume across the roll — DONE, and it found a shift

The guardian's standing caution asked for this, with the burden set against us. There **is** a shift:

```
08-24: 35 · 08-25: 19 · 08-26: 16 · 08-27: 10 · 08-28..08-31: 0
```

> #### ⚠ THE ITEM TYPE IS NOT THE CHECK NAME — the obvious query returns a false zero
>
> The 08-26 handoff told me to "compare `misdirected_cta` finding volume". Doing that literally —
> `WHERE item_type='misdirected_cta'` — returns **zero rows in all of history, live and archive**,
> which reads as *"this check has never found anything"*. The check is **named** `misdirected_cta`
> (`check_misdirected_cta.go:64`) and **files** `item_type='cta_names_unknown_destination'`
> (`:352`). Same trap shape as `audit_cta_label_agreement` vs `cta_label_audit.go` — appended to
> `LANDMINES.md`.

Discriminating the shift, in order:

1. **Is the host agent dormant?** No. Sibling checks on the *same* `completeness-discovery-agent`
   produced findings over 08-28..08-31: `head_essentials_missing` **38**, `canonical_mismatch` **36**.
   So the agent runs and this check specifically went quiet. The control does **not** exonerate.
2. **Has the population shrunk?** Yes, substantially. Same census, same predicate as 08-26:

   | | 2026-08-26 | 2026-08-31 |
   |---|---|---|
   | mismatched | 186 | **126** |
   | pairs | ~1,192 | **1,779** |
   | rate | 15.6% | **7.1%** |
   | sites | — | 23 |

   Pairs **grew 40%** while mismatches **fell 60 in absolute terms** — so this is not dilution by new
   good pages, it is repair. The 389/391 lanes have been draining exactly this population all week.
3. **Would the survivors even refile?** Mostly no. The convicting subset was only **13 of 186** on
   08-26 (the rest name no single page); scaled to 126 that is ~9 pages. And **99**
   `cta_names_unknown_destination` items sit in `needs_human_review` — non-terminal, so the dedup
   index suppresses a refile of those same `item_key`s.

**Conclusion: the shift is explained without invoking the extraction, and the burden is discharged.**
`[INFERRED]`, not proven — the decisive test would be running the check's own ranked predicate
against a page known to convict and watching it convict. I did not do that. What I can say is that
every cheap alternative explanation (dormant agent, shrinking population, dedup suppression) is
independently evidenced above, and the detector-broke hypothesis has no evidence for it at all.
