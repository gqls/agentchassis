# 327 — a PARTIAL write to `content_direction` silently shrinks the brief the writer reads, and the document keeps growing so nothing looks wrong

> ## ⚠ THE NUMBER 327 IS AMBIGUOUS — refer to this case BY SLUG
> Another lane filed `bugs_open/327_HANDOFF_2026-08-19_the_build_trigger_can_publish_nothing_and_exit_zero.md`
> the same day. Both are real, neither is renumberable (numbering is never reassigned), and
> `scripts/who-owns.py 327` now prints the ambiguity warning. This case is
> **`a_partial_spec_write_silently_shrinks_the_brief_the_writer_reads`**. `git log` the FILE PATH.

> ## ✅ FIX IS **LIVE** ON `v1.0.1319` (2026-08-20 ~14:38Z) — and this file stays OPEN, for a reason worth reading
> **Binary-probed with full controls, not inferred from a tag.** The literal `merged_keys`, which
> only this fix introduces, reads **1** on a `v1.0.1319` pod and read **0** on `v1.0.1317` eleven
> hours earlier; the pre-existing literal `formatted_len` reads **1** on both (so the probe is not
> simply always agreeing) and a deliberately impossible string reads **0** (so it can return zero).
> Two distinct pods, one of them `agent-build-dispatch-loop-…`.
>
> **Why it is not CLOSED.** The bar is fixed AND live, and the *defect* is both — a partial write
> can no longer shrink a brief. But **the damage is not repaired and this fix does not repair it**:
> the three fragment briefs stay fragments until something writes those specs again.
> `[MEASURED 2026-08-20 14:40Z]` **no `content_direction` write has occurred on any site since the
> roll**, so nothing has been restored yet and nothing has been observed on a live write. Closing on
> a binary probe alone would be trusting a status over an artefact, which is the one thing this
> lane's own rules forbid.
>
> **What closes it:** one real `content_direction` write, then `audit_writer_brief.py <domain>`
> showing zero non-empty dropped keys for that site. The three affected sites belong to other lanes,
> so the write is theirs to make — see the restoration warning in `LANDMINES.md`, which is now
> unconditional.

> ## (superseded) FIXED IN CODE 2026-08-20 — `c9a71388f` — inert until a chassis roll
> Both defects, one commit: `formatted` is now built from `merged` (`site_spec_actions.go`), and
> `FormatContentDirection` sorts its keys at both levels so an identical spec renders identically.
> **Council-Submitted: `db3c158b-4dab-4a1b-bb2b-875dbac98358`** (advisory; the verdict is still owed
> a read, and a REVISE must be acted on — the code is already on the shared branch).
>
> **The test that was missing now exists** — `platform/orchestration/actions/site_spec_formatted_from_merged_test.go`.
> It drives the real `WriteSiteSpecAction` through `sqlmock` and asserts on the `data` JSONB actually
> inserted, because a formatter-only test passes in the broken world too: the formatter was never
> wrong, the action was handing it the wrong map. **Mutation-proven against genuine `HEAD`** (both
> files restored from git, not hand-mutated): the label test names the six keys that really vanish —
> `heading_style, terminology, example_phrases, persuasion_approach, things_to_avoid, writing_rules` —
> and the determinism test fails at iteration 0. ⚠ My first hand-mutation changed only the call's
> ARGUMENT and not the block's POSITION, and failed for a different reason (one missing key); that is
> why the faithful comparison was run, and it is the general lesson — **a mutation that fails can
> still be failing for the wrong reason.**
>
> **What remains before this closes:**
> 1. **A roll.** Verify at the artefact, per service, not at git.
> 2. **The three fragment briefs stay fragments until each is written again** — the fix stops the next
>    write destroying more; it restores nothing by itself. §5.2's warning stands: a backfill is a
>    content change, and on `ai-agent-orchestration.com` part of what returns teaches the very
>    construction the owner objected to.
> 3. `audit_writer_brief.py --fleet` expecting zero non-empty dropped keys — **after** those sites
>    are written again, not after the roll. A green run before then would be measuring nothing.
>
> ⚠ **The actions package is RED on HEAD for unrelated reasons** — three failures belonging to
> another lane's `bugs_open/336` (a live WORKFLOW_INVALID incident) plus one order-dependent identity
> test. Verified identical with and without this change, so the honest claim is "my tests pass and I
> add no new failures", not "the suite is green".


**Filed 2026-08-19** by `copy_quality_two_stage`.
**Diagnosis loop:** `090` filed the same session, `RUN_CORRELATION_ID=8be5f6e9-d0b3-43f7-9ee4-dee2432dd8b1`
(per the owner ruling of 2026-07-31 — verdict appended below when it lands).

> ## What this is, in one paragraph, before any jargon
> A site's page brief is a JSON document with about twenty parts — voice, things to
> avoid, example phrases, heading style. The writer does not read that document. It
> reads **one field of it**, called `formatted`, which is supposed to be the whole
> document turned into readable prose. That field is rebuilt on every write — but it is
> rebuilt from **only the part being written**, before that part is merged into what was
> already there. So a small update to two sections replaces the brief with a rendering
> of those two sections, and the other eighteen stop reaching the writer. They are still
> in the document, so every query, every dashboard and every reviewer still sees them.

## 1. The mechanism, in code

`platform/orchestration/actions/site_spec_actions.go`:

```
:212    formatted := datahelpers.FormatContentDirection(specMap)   // specMap == the INCOMING partial
:213    specMap["formatted"] = formatted
...
:247    merged := siteSpecDeepMerge(currentData, specMap)           // merge happens AFTER
```

`FormatContentDirection` walks whatever map it is handed and renders every key
(`datahelpers/format_content_direction.go`). Handed the incoming partial, it renders the
partial. The deep merge then puts that short `formatted` over the previous full one,
because `formatted` is just another key and the newer value wins.

> **⚠ CORRECTED 2026-08-19 — the adoption path does NOT have this bug, and the loop caught me.**
> This file first read: *"the same ordering is in the adoption path —
> `apply_adoption_plan_action.go:280` formats `directionData` and never sees the merged result."*
> The ordering is the same and **the consequence is not**, because that path never merges at all:
> `apply_adoption_plan_action.go:386-421` does a straight `UPDATE … is_current=false` then a
> fresh `INSERT` of `directionData` wholesale, with a comment saying so
> (*"adoption re-derives them from the crawl wholesale"* — only the `structure` aspect
> carries keys forward). With no merge there is no divergence: `formatted` describes the whole
> document because the document **is** what was just written.
> **Adoption has a different, visible failure instead** — it can shrink the stored document, which
> a reader can see. This bug is specifically about the document staying complete while the brief
> silently does not.
> Found by the `090` run's iteration 2, from a path I had not taken: *"apply_adoption_plan_action.go's
> content_direction branch, by contrast, does a full supersede+insert with no merge at all, so
> that path can't exhibit the hypothesized bug."* **It is right.**

**So the invariant that would fix it is one line long:** `formatted` must be computed
from `merged`, never from the incoming partial.

## 2. It fired, and three sites are still living with it `[MEASURED 2026-08-19]`

Every write where the brief lost a quarter or more of its size, all history, all sites:

| date | site | transition | `formatted` before → after |
|---|---|---|---|
| 2026-04-18 | `leopardessconsulting.co.uk` | `domain-research-classifier` → `build-site-planner` | 10,263 → 3,766 |
| 2026-04-18 | `ai-agent-orchestration.com` | same | 9,279 → **3,538** |
| 2026-04-18 | `finetuning.uk` | same | 9,288 → 3,081 |
| 2026-04-18 | `robot-hands.com` | same | 10,135 → 3,324 |

**The key sets prove it is the partial and not a rewrite.** For
`ai-agent-orchestration.com`, the keys whose labels appear in each row's `formatted`:

- classifier, 18:31Z — `content_depth, cta_style, example_phrases, heading_style,
  paragraph_style, persuasion_approach, sentence_style, social_proof_style, terminology,
  things_to_avoid, things_to_emulate, voice, writing_rules` (13)
- planner, 18:40Z — `avoid_phrases, blog_strategy, emphasis, social_proof_style, voice` (5)

Those five are the planner's own write (`blog_strategy` is new in that row and appears in
no earlier `formatted`). **`formatted`'s key set is the new write's key set, not the
merged document's** — which is the signature the code predicts. The merged document went
to 19 keys at the same moment.

`finetuning.uk` recovered on 2026-08-12 when an operator wrote a full document. The other
three have been serving a fragment **since 2026-04-18** — four months, every page.

## 3. What each affected site is not showing its writer `[MEASURED 2026-08-19]`

Tool: `copy_quality_two_stage/audit_writer_brief.py <domain>` (`--fleet` for all).

| site | keys dropped | writer sees | document is |
|---|---|---|---|
| `robot-hands.com` | 14 | 5,077 chars | 19,988 |
| `leopardessconsulting.co.uk` | 12 | 7,669 | 17,074 |
| `ai-agent-orchestration.com` | 12 | 8,517 | 15,760 |
| `loanandmortgagecalculator.co.uk` | 3 | 19,909 | 44,033 |
| `loancalculator.co.uk` / `loancash.co.uk` | 2 | — | — |
| `gamesdesign.co.uk` | 1 | — | — |
| the other 17 sites | 0 | — | — |

For `ai-agent-orchestration.com` the dropped keys are `writing_rules` (1,428 chars),
`things_to_emulate`, `things_to_avoid`, `content_depth`, `persuasion_approach`,
`example_phrases`, `sentence_style`, `heading_style`, `terminology`, `paragraph_style`,
`cta_style`, `trust_signals`. Its `things_to_avoid` is eight specific bans — *"the word
'seamless'"*, *"urgency or scarcity language"*, *"generic AI hype vocabulary"*,
*"passive voice in technical descriptions"*. **None of them has reached the writer since
April.**

⚠ An empty key is not a loss and is reported separately: `compliance_rules: []` is absent
from 13 sites' briefs and takes nothing with it. The tool's control for that arm exists
precisely so the headline count cannot be inflated by empties.

## 4. ⚠ What this does NOT explain — stated plainly, because the temptation is obvious

**This is NOT the cause of the owner's define-by-negation complaint** (`bugs_open/305`),
even though it was found while investigating it and even though it hits the same site.
Two reasons, and both are checkable:

1. The dropped `things_to_avoid` **never mentions the construction.** It bans hype
   vocabulary and urgency language. Restoring it would not have stopped a hero being
   written as *"X, not Y"*.
2. The dropped `example_phrases.characteristic` is **itself written in the
   construction** — *"Agents fail in isolation — not in cascades"*, *"Speed comes from
   engineering discipline, not from skipping the hard parts"*. On this estate's own
   measured principle that the example is the instruction, restoring that key naively
   would push the writer **towards** the shape the owner objected to, not away from it.

305's root cause stands where it is: a **supplied** phrase — the canonical tagline —
transfers verbatim (1,369 rendered prompts → 409 responses, re-measured today). This bug
is a separate, larger defect that happens to live in the same field.

## 5. Fix candidates, ordered by what makes the bad state unrepresentable

1. **Compute `formatted` from `merged`** (`site_spec_actions.go`, move the block from :212 to
   after :247). **`apply_adoption_plan_action.go` needs no change** — see the correction in §1;
   it does not merge, so its `formatted` already describes its whole document. Removes the
   failure for every future write. **A shared platform seam** — RFC/council per CLAUDE.md, and it
   changes what every `content_direction` consumer sees, so the other consumers must be
   **told**, not merely measured (owner ruling 2026-07-29 §3).
2. **Then backfill the three sites** by recomputing `formatted` from the current merged
   document. ⚠ Backfill is **not** a no-op on output: it restores ~10,000 chars of brief
   to sites whose pages were written without it, and — see §4.2 — some of what returns
   teaches the construction. Recompute and **read the diff** before writing it; do not
   run it as a sweep.
3. **A drift check**, cheap and mechanical: every current `content_direction` whose
   document keys are not all represented in `formatted` is a defect.
   `audit_writer_brief.py --fleet` is that check today, as an observation; a CronJob
   writing one `doc_notes` row is the automatable form.
4. **Weakest, and named only to be dismissed:** telling authors to always write full
   documents. It is an operator-must-remember rule on a tree many sessions share, and
   the next partial write is one migration away.

## 6. How to verify a fix

- **Unit**: write a two-key partial over a ten-key spec; assert every one of the ten
  labels is present in the resulting `formatted`. The bug is that this test does not
  exist — `FormatContentDirection` is tested on the map it is given, which is exactly the
  wrong scope, because the defect is in **which map it is given**.
- **Live**: `audit_writer_brief.py --fleet`, expect zero non-empty dropped keys.
- ⚠ **The failure is silent in every place an operator would look.** The document is
  complete, the write succeeded, `formatted_len` is logged as a healthy number for the
  partial, and no error is raised. Only the comparison of document-keys to brief-labels
  shows it.

## 7. What has NOT been done

- **No code change** — a shared seam, and this lane is config-only by design.
- **No backfill** — see §5.2; and `robot-hands.com` / `leopardessconsulting.co.uk` are
  other lanes' sites. Told them (CONTRIBs of this date).
- **No edit to any spec.** ⚠ **Anyone about to "fix the briefs" must read §1 first**: a
  targeted correction written as a partial will itself collapse the brief to whatever it
  touched. That trap is filed in `LANDMINES.md`.

---

# `090` ITERATION 1 returned `UNVERIFIABLE` and named four gaps. All four are closed below, first-hand.

> ## ⚠ CORRECTED WITHIN THE HOUR — I READ AN INTERMEDIATE STEP AS THE FINAL VERDICT
>
> This section was first written under the heading **"`090` VERDICT IS IN"**. It was not.
> `UNVERIFIABLE` with a populated `code_requests` / `data_requests` block is **the loop asking
> for its next bundle**, not a conclusion — and seven minutes after I read it, orchestration
> `6073488a` moved from `verdict` back to **`assemble_bundle`** and began iteration 2 with the
> three symbols and the query it had asked for. `[MEASURED 2026-08-19 16:02Z]`
>
> **What gave it away was the step, not the outcome field.** The tell is one query:
> `SELECT current_step, status FROM orchestration_states WHERE correlation_id='<corr>'` — a run
> still cycling `assemble_bundle → call_diagnoser → verdict` has not finished, however
> conclusive its latest `outcome` string reads. I had already read this lane's own handoff
> warning that a missing row is latency rather than a dropped dispatch; the same impatience in
> the other direction produced this.
>
> **What survives unchanged:** the four gap-closures below are first-hand code and data, and
> they do not depend on the loop's outcome at all. **What changes:** iteration 2 may reach a
> different outcome, and if it does, it is reading evidence I did not have when I wrote this.
> Whoever picks this up should read the FINAL verdict rather than quoting this section.

**Run `8be5f6e9-d0b3-43f7-9ee4-dee2432dd8b1`, orchestration `6073488a-3082-447d-8bd0-d8ee53000136`.**
Iteration 1's outcome was **`UNVERIFIABLE`** — which here means *not closed on that bundle*, not
*refuted*. Read its own words before mine: it **confirmed the ordering** from the code it could
see —

> *"FormatContentDirection(specMap) runs on the incoming partial spec_data **before currentData
> is even read from the DB**, and siteSpecDeepMerge is called afterward."*

— with two `static`-tier citations at `site_spec_actions.go:WriteSiteSpecAction`, and then
declined to conclude, because its bundle contained neither of the two function bodies the
hypothesis turns on, nor any `site_specs` rows. **That is a bundle-assembly limit, not a
counter-argument** — the same class of gap the previous `090` run in this lane hit when symbol
bodies came back "unavailable". Per the owner ruling of 2026-07-31 I am **stating plainly that
first-hand verification is substituted here**, and naming what for, rather than leaving the
outcome to read as agreement.

**Its gap 1 — `siteSpecDeepMerge`'s body**, *"unverified whether it overwrites the prior
'formatted' string with the new partial's value"*. It does, unconditionally
(`site_spec_actions.go:512-535`):

```go
srcMap, srcIsMap := srcVal.(map[string]interface{})
dstMap, dstIsMap := dstVal.(map[string]interface{})
if srcIsMap && dstIsMap {
    result[k] = siteSpecDeepMerge(dstMap, srcMap)
} else {
    result[k] = srcVal          // <- `formatted` is a STRING, so this arm always runs
}
```

`formatted` is a scalar, so the map-merge arm is unreachable for it and the newest — shortest —
value wins outright. **This is the load-bearing step and it behaves exactly as the hypothesis
required.**

**Its gap 2 — what `FormatContentDirection` renders, and whether it could reach `currentData`.**
It cannot: its only parameter is the map it is handed
(`format_content_direction.go`, `func FormatContentDirection(spec map[string]interface{}) string`),
it has no DB handle, no context, and its whole body is `for key, val := range spec`. Handed the
partial, it can only render the partial.

**Its gap 3 — the second write path it could not see.** ⚠ **My answer here was WRONG and
iteration 2 corrected it.** `apply_adoption_plan_action.go:280` does format `directionData` before
writing — but that path **supersedes and re-inserts wholesale with no merge** (`:386-421`), so
`formatted` and the document always agree and the bug cannot occur there. I asserted "same
ordering, same consequence"; only the first half holds. See the correction block in §1. **This is
the run paying for itself**: one round refuted a claim I had already committed, and the fix
candidate in §5.1 is narrower as a result.

**Its gap 4 — *"no observed instance where `jsonb_object_keys(data)` contains labels absent from
`data->>'formatted'`"*.** That is §2 and §3 of this file, run before the verdict arrived, and it
is the loop's own requested query shape: 8 of 25 sites, three of them 12–14 keys, with the
`2026-04-18` classifier→planner transition dated and its surviving key set identified as the
planner's own partial. The tooling for it is `audit_writer_brief.py --fleet` (register CQ-025).

**What the verdict is worth keeping for.** It was right that the bundle could not close this,
and its `next_scope` list — `siteSpecDeepMerge`, `format_content_direction.go`,
`apply_adoption_plan_action.go` — is exactly the three places the answer was. A run that names
its own gaps precisely is doing its job even when it stops short.

---

# ✅ FINAL `090` VERDICT: **CONFIRMED** (iteration 3, orchestration COMPLETED) — with one of its own citations corrected, and a SECOND defect it surfaced by accident

**Run `8be5f6e9-d0b3-43f7-9ee4-dee2432dd8b1`. Item `complete`, orchestration
`6073488a-3082-447d-8bd0-d8ee53000136` `COMPLETED`, outcome `CONFIRMED`.** It took three
iterations; the two `UNVERIFIABLE` rounds above were it ordering the bundles it needed.

**Its four static citations are the mechanism, and they match the reading in §1 exactly** —
`formatted := datahelpers.FormatContentDirection(specMap)` and
`merged := siteSpecDeepMerge(currentData, specMap)` both in `WriteSiteSpecAction`, the
`else { result[k] = srcVal }` arm in `siteSpecDeepMerge`, and the action's own comment that the
writer reads `formatted` *"as one field regardless of how the structured data is organised"*.
Its own summary of the third: the writer *"has no other path to the deeper keys."*

## ⚠ But its STATE evidence does not say what it says — and checking it found something else

The verdict's fourth `symptom_check` cites three `loanzy.uk` rows from 2026-08-18 (13:36, 15:58,
20:19) with an **identical 14-key document**, whose `formatted` opens on *"Compliance rules:"*,
*"Voice:"* and *"Content depth:"* respectively — read as *"'formatted' switches wholesale between
… each describing only the [subset]"*, i.e. the bug firing live last week.

**It is not.** All three rows are **complete** `[MEASURED 2026-08-19]`:

| row | `formatted` chars | labels present / keys | opens on |
|---|---|---|---|
| 13:36 | 9,802 | **13 / 13** | `Content depth:` |
| 15:58 | 10,115 | **13 / 13** | `Voice:` |
| 20:19 | 10,108 | **13 / 13** | `Compliance rules:` |

Same content, **different order**. The loop compared opening lines and inferred subsets.

## The second defect: `formatted` is regenerated in a RANDOM key order on every write

`FormatContentDirection` builds its output with `for key, val := range spec` — and **Go
randomises map iteration order by design**. Nothing sorts the keys before or after. So two writes
of an identical document produce two completely different briefs, character for character, with
the same content shuffled.

**This is a separate defect from the one this file is about** — different failure, same function,
and the remedy is one line (iterate a sorted key list). Consequences, in the order they bite:

1. **You cannot tell whether a brief changed by comparing `formatted`.** A text diff between two
   writes reports ~100% changed whether or not anything did. ⚠ **That lands directly on the work
   this lane and `portfolio_positioning` are about to do**: correcting a brief and then verifying
   the correction landed. Verify by **label presence and key content**, never by diffing the
   rendered brief — which is what `audit_writer_brief.py` already does, by luck rather than
   design when it was written.
2. **The writer meets its instructions in an arbitrary order** that changes on every spec write —
   compliance rules first on one write, content depth first on the next. Whether that moves
   output is **[UNMEASURED]**, and it is a genuinely testable question.
3. **Prompt-cache cost: checked and DISMISSED.** The brief sits ~21% into a ~45,000-char rendered
   prompt (position ~9,450, n=3), so a reordering would invalidate the cached prefix from there
   on — **but the order is fixed once stored**, so every call between two spec writes sees the
   same text, and a spec write would break the prefix anyway by changing content. No additional
   cost. Recorded because the arithmetic looked alarming until it was done.

## What the run was worth, stated plainly

It **confirmed** the mechanism from citations I had also read, **refuted** a claim I had already
committed (the adoption path — see §1's correction block), and **misread one state citation**,
which is how the ordering defect was found. Three useful outcomes, two of them corrections to
somebody's committed work. That is a good return on one run, and it is an argument for reading a
verdict's evidence rather than its outcome field.

---

# COUNCIL ROUND 1: **REVISE** (gating objection from editquality), 12 reviewers, 5 abstained. What each objection produced.

⚠ **Read the verdict BY CORRELATION, not from the newest `doc_notes` row.** My first attempt
returned another lane's `bugs_open/336` verdict, which is the documented trap:
`SELECT body FROM diagnosis_artifacts WHERE correlation_id='db3c158b-4dab-4a1b-bb2b-875dbac98358' AND kind='council_report';`

**Round 2 submitted on the same correlation** (`RESUBMIT_CORR`, so the trail accumulates), run
orch `8f5b5dfe-0459-47a8-98ef-2112374504fe`. Two objections changed the submission; the rest were
answered with a check.

### editquality, HIGH — "the formatter might embed the OLD brief inside the NEW one"

The best objection of the round, and it was right that the assumption was **asserted, not shown**.
Rendering from `merged` means the previous `formatted` string is present at call time (the partial
has no such key, so the merge keeps the old one). If the formatter walked it, every write would
nest the whole previous brief inside the new one — **compounding, and worse than the data loss it
replaces.**

It does not, and never did: `format_content_direction.go:41-44` has carried
`// Skip the formatted field itself / if key == "formatted" { continue }` all along. **But nothing
tested it**, which is the whole of the criticism. Now something does —
`TestBriefDoesNotNestItsOwnPreviousRendering` asserts every label appears **exactly once**, because
a nested render repeats every label the old one held. Mutation: delete the skip and it fails with
`label "Voice:" appears 2 times, want exactly 1`.

### editquality, LOW — "sorting covers only two levels" · and the shape measurement it asked for

Factually the sort is applied wherever a map is rendered and `FormatSpecValue` recurses into
itself, so **every** depth goes through it — but that was asserted too. `[MEASURED 2026-08-20]`
across all 25 current specs: **224 object / 102 array / 24 string** top-level values, **zero**
three-level maps, **zero** arrays-of-objects. So production is at most two levels deep today, and
the new `TestBriefIsDeterministicAtEveryDepth` uses a deliberately **four**-level fixture, deeper
than anything live, plus an assertion that the deepest values actually render so stability cannot
be satisfied trivially. Mutation: unsorting **only** the nested level fails it.

⚠ **Separate pre-existing drop, found by that measurement and NOT fixed here:** an array whose
elements are objects is discarded entirely by `InterfaceSliceToStrings`. Zero live instances, so it
is a landmine to record rather than a change to smuggle into a bugfix round.

### bug_historian, MEDIUM — "you are patching one aspect branch of a shared mechanism that has two other recorded silent-drop shapes"

Audited, and the answer is clean. `grep -n 'if aspect == ' site_spec_actions.go` returns exactly
three: `:180` an emptiness guard, `:240` `identity` → `normaliseServicesField(merged)`, `:265`
`content_direction`. **`normaliseServicesField` already operates on `merged`, so after this change
there is NO remaining pre-merge derivation in the function** — the shape being worried about does
not exist elsewhere in it.

The two siblings are **different root causes**, checked not assumed:
- **`pinned` ignored** — `grep -c pinned site_spec_actions.go` = **0**. That defect is "the column
  is never read", not "a derived value is computed pre-merge".
- **`structure` opt-in flags dropped on re-adoption** — that is in `apply_adoption_plan_action.go`,
  which does not merge at all (supersede + wholesale insert, `:386-421`) and already carries its
  own `carryForwardStructureSpecKeys` remedy.

### prior_art_librarian, MEDIUM — "the reused matcher and the no-tests claim are asserted"

Both confirmed. `captureArg` is at `tool_acceptance_convergence_test.go:61` (added `b13238be6`);
my file declares no second copy — the **first build failed with `captureArg redeclared`, which is
how I found it**, and reusing it was the fix. And
`grep -rln 'FormatContentDirection|siteSpecDeepMerge' --include=*_test.go platform/` returned
nothing before this change, so there was no colliding file.

### debug_historian, MEDIUM — "no pod-verification recipe" · ACCEPTED, and this is the close-out

Primary — ask the service, per SERVICE not per fleet:
```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --limit-bytes=900000 | grep -m1 'build provenance'
git merge-base --is-ancestor 90930a4a8 <the stamped sha>
```
If the startup line has scrolled (it had on 2026-08-19), probe the binary for a literal this change
**adds**, with a control in the same breath: `grep -aq 'merged_keys' /proc/1/exe` must hit,
`incoming_keys` must hit, and a sha created after the build must be absent. Never `strings`; never a
discovery grep for "some 40-hex string".

### architecture, APPROVE with a follow-up — recorded here so it is not left in a verdict nobody re-reads

*"This pattern (format-before-merge racing a partial write) is generic to any `WriteSiteSpecAction`
aspect that also auto-derives a formatted/summary field from `specMap` before the merge — if other
aspects grow that shape, the fix belongs in `siteSpecDeepMerge` or a shared post-merge hook, not
per-aspect."* Agreed. **The general shape stays reachable for any future aspect that copies it**,
and closing it properly would be architecture-scope. Named latent risk, not a to-do for this round.

### guardian MEDIUM + compliance HIGH — "the fix un-suppresses known-bad brief content automatically; add a scrub or a key-skip guard"

**Decision: no code gate, and this is the reasoning rather than a dismissal.** The seats are right
that the fix converts a latent exposure into an active one.

1. **A measurement that lowers the severity, and it is checkable.** The phrase with the *proven*
   transfer chain — the canonical tagline, 1,369 prompts → 409 responses — is **already in the live
   brief today**, in `emphasis`, one of the five keys that survived the collapse, in a block that
   also **orders** it into every hero, the footer and every meta description. What the fix restores
   is `example_phrases`, whose transfer into output is **unmeasured**. The worst payload is live
   now and this change does not touch it.
2. **A site-specific key-skip in a shared action is the wrong shape.** "Skip a key flagged in
   `bugs_open/305`" hardcodes one bug's payload into the merge path for every site and every
   aspect, with no expiry — the accumulation-on-a-shared-seam the architecture seat exists to
   catch, and it makes correctness conditional on a content list somebody must remember to prune.
3. **The suppression being removed is an ACCIDENT, not a control.** The bad content is in the
   document and reaches nothing only because of a data-loss bug. Keeping that is keeping a defect
   as a safety feature — and an undetectable one, which is how it lasted four months.
4. **So the remedy is the payload, and it is another lane's site config.** Contained instead by:
   three CONTRIBs, a fleet LANDMINE (2026-08-20) so a session that never reads them still meets the
   warning, and a per-site command that lists exactly which keys will reappear, largest first, so
   the call is made deliberately **in the same write**. The roll gives that decision a deadline,
   which beats a status quo where nobody knew the keys existed. I have offered to make the
   `example_phrases` edit myself if that lane wants it.
5. **If the seats still want a gate**, the containable version is a per-site **opt-in field with
   the unsafe default OFF** (the 2026-08-02 ruling's shape), not a hardcoded key list. I do not
   think one instance of unmeasured-transfer exemplars justifies a new opt-in on this seam, and
   putting that judgement to the round is the point of submitting.

---

# COUNCIL ROUND 2: **REVISE** again — gated by `compliance`, on the same HIGH. Two objections acted on, one escalated to the owner.

Round 2 (orch `8f5b5dfe-0459-47a8-98ef-2112374504fe`, 12 reviewers, 5 abstained) accepted the two
round-1 fixes: nobody re-raised the nesting hazard or the depth question. Three things came back.

### 1. `bug_historian`, MEDIUM — "the audit closes today's three branches; the SHAPE stays reachable" · **BUILT**

Their sharpest line, and it is right: *"a bug-file note IS exactly that unread verdict"* — they cite
closed cases 086, 093 and 042 as "the underlying rule stayed generic and got hit again on the
untouched branch". So the note became a test: **`TestNoAspectBranchDerivesBeforeTheMerge`**
enumerates the aspect branches (`identity`, `content_direction`) and asserts **every derivation sits
after the merge**, by source position. A fourth branch fails it and has to be accounted for
deliberately.

⚠ It scans source, so it is built against the known trap that *a source-scanning test makes your
comments load-bearing*: comment lines are **stripped before anything is counted**, and a decoy
comment (`// if aspect == "strategy" { derive from specMap }`) leaves it passing. Three mutations:
genuinely pre-fix source **fails** on the ordering arm, a fourth branch **fails** on the enumeration
arm, the decoy is correctly ignored.

### 2. `bug_historian`, LOW — "open `bugs_closed/025` before merging" · **READ, and it corroborates**

025 is `pages.content_direction` — the **per-page column**, a different thing that shares the name,
filed 2026-07-19 and closed 2026-07-21 by wiring it up. Its own evidence says *"the writer prompt
only ever dereferences `.site_specs.specs.content_direction.formatted`"*, which is this round's
premise arriving from a different direction. **No contradiction, and no re-derivation** — but the
name collision it warns about is real and worth carrying: `pages.content_direction` and the
`site_specs` aspect are unrelated.

### 3. `debug_historian`, LOW — "your verification recipe samples the wrong pods" · **CORRECT, and my recipe was wrong. Replaced.**

They said `-l app=agent-chassis` returns 2 pods while ~41 run the binary. `[MEASURED 2026-08-20]`
**it is worse: 75 pods run the `agent-chassis` image and the selector matches 2** — 70 are labelled
`dynamic-agent`, plus one `vet-intel` and one `business-intel`. A spec write almost certainly
executes on a `dynamic-agent` pod, so **the recipe I published would have proved nothing about the
code path that matters.** Superseding it:

```bash
# 1. Is the fleet uniform? Filter on the IMAGE, never on a label.
kubectl -n ai-persona-system get pods \
  -o jsonpath='{range .items[*]}{.spec.containers[0].image}{"\n"}{end}' \
  | grep 'agent-chassis:' | sort | uniq -c        # more than one line == MIXED, stop here
# 2026-08-20: `75 docker.io/aqls/agent-chassis:v1.0.1317` — uniform, and pre-fix.
# 2. Probe a pod that actually runs the path (a dynamic-agent one), WITH a control:
kubectl -n ai-persona-system exec <a dynamic-agent pod> -- sh -c \
  'grep -ac merged_keys /proc/1/exe; grep -ac incoming_keys /proc/1/exe; grep -ac formatted_len /proc/1/exe'
# want 1,1,1. The third is the PRE-EXISTING literal: if it is 0 the probe itself is broken.
```

**Current state, measured this way: `0 0 1` on the running binary — the fix is NOT live**, and the
pods started `2026-08-19T22:26:25Z`, eleven hours before the commit, so two independent routes agree.
⚠ **`IMAGE_TAG` in the makefile is already `v1.0.1318` while the fleet runs `1317`**, so a build is
imminent and this fix will ride it.

### 4. `compliance`, HIGH — held for a second round. **NOT resubmitting a third time; this is an OWNER DECISION.**

Their round-2 wording is stronger than round 1 and quotes my own concession back at me:

> *"The chosen mitigation is notification-only (CONTRIBs + a fleet landmine + a listing command) —
> none of it stops the automatic write or requires review before the restored content reaches a
> writer prompt. A per-site opt-in default-OFF gate (which the author concedes as the containable
> alternative) is the correct shape here; relying on other lanes reading a CONTRIB before their next
> scheduled write is not a control."*

**They are quoting the owner's own 2026-08-02 ruling** — *"a comment is not a control on a tree this
many sessions share"* — and by that ruling they have the better of the argument on form.

**Why I am not simply complying, and why this goes to the owner rather than to a third round.**
CLAUDE.md's rule for exactly this situation: a judgement about *consequence* rather than measurement
is not answered by resubmitting with better numbers — *"record it where the change lives … and let a
human break it — especially when seats disagree with each other"*, which they do (`architecture`
approved this as a `point_fix`; `compliance` calls it HIGH).

**The decision, with both options costed.**

- **(A) Ship as committed.** The 8 lossy sites' briefs repair themselves at their next spec write.
  Cost: on `ai-agent-orchestration.com` the restored `example_phrases` include define-by-negation
  exemplars, reaching a writer prompt with no human in the loop.
- **(B) Add a per-site opt-in, default OFF.** Cost: **the bug stays live by default on all 25
  sites** until someone opts each one in — a data-loss defect preserved as a safety feature, and
  this estate has been bitten by mechanisms rotting unexercised behind a default-OFF switch. The
  unsafe side here is genuinely ambiguous, which is what makes it an owner call rather than a
  mechanical application of the 08-02 ruling: that ruling assumes the new authority is the risk,
  and here the new authority is *the repair*.

**What I did to narrow the gap instead** — because "unmeasured" was the load-bearing word in my own
answer and I owed the round better than an absence of evidence. `[MEASURED 2026-08-20]` 60 distinct
`example_phrases` strings from the 21 sites whose exemplars DO reach the writer, checked verbatim
against every successful `page-content-writer` call: **52 reached a prompt, 3 appeared in output, 18
total appearances — and none of the three carries the construction** (they are hooky openers). For
contrast, the single **mandated** tagline produced **409**. So exemplar transfer is real but two
orders of magnitude below a supplied-and-mandated phrase.
⚠ **The 60 were the alphabetically first of 194, not a random sample**, so 3-of-52 is indicative and
not a rate to quote; and a low *literal* rate says nothing about whether FORM transfers, which is
still open in `bugs_open/305`. **This narrows the risk; it does not retire it.**

**Recommendation to the owner: (A), plus fixing that one spec key before the roll** — which is
`ai-agent-orchestration.com`'s lane's call on their own site config, and I have offered to make the
edit. That keeps the repair and removes the payload, rather than preserving a defect to contain its
side effect.

---

# ✅ VERIFIED AT THE ARTEFACT 2026-08-23 — on two real writes, by two different producers. Closing.

The fix went live on `v1.0.1319` (binary-probed) but **nothing had exercised it** — zero
`content_direction` writes had occurred. Two have now happened, and both behave:

| write | producer | doc keys | keys reaching the brief |
|---|---|---|---|
| `loanzy.uk`, 2026-08-22 09:04 | `owner_release_20260822` (an operator write) | 14 | **all 14 with content** |
| `apis.uk`, 2026-08-22 12:26 | `domain-research-classifier` | 13 | **12, the 13th being an empty `compliance_rules`** |

`audit_writer_brief.py` reports `none — every key with content reaches the brief` for both. **The
defect is fixed, live, and now proven on real writes by two independent producers**, which is the
bar. Moving to `bugs_closed/`.

⚠ **The first reading of this was WRONG, in the direction that would have wasted someone's day.**
`loanzy.uk` initially reported `layout_preservation` (370 chars) as DROPPED on a **post-fix** write —
i.e. "the fix does not work". It was a false positive in my own tool: the brief labels that block
`Layout preservation (rerender rule):`, and the check tested for `Layout preservation:` exactly. The
tell was arithmetic that did not fit the story — **`formatted` had grown by ~400 chars on the very
write that supposedly dropped a 370-char key.** Fixed (label must start a line and reach a colon
without crossing a newline), with the 12 genuine drops on `ai-agent-orchestration.com` as the control
that the loosening did not erase real findings.

## The owner's rulings, 2026-08-21 — the council escalation is ANSWERED

The `compliance` seat's HIGH objection was escalated rather than resubmitted (council round 2). The
owner ruled:

1. **Ship as-is. No gate.** The repair stays ungated; the fix is live and unconditional.
2. **The three fragment briefs are their own lanes' to repair** — not to be pushed by this lane.
   They have each been told twice and given the command that lists exactly what will return.
3. (Separately) run 4's proposal was approved and is applied — see `CQ-024`.

**So the residual is DATA, not defect, and it is assigned.** `ai-agent-orchestration.com`,
`robot-hands.com` and `leopardessconsulting.co.uk` still serve fragments
(3,558 / 3,324 / 3,774 chars as of 2026-08-23) and repair themselves the moment anything writes those
specs. The restoration trap — including that one site's returning `example_phrases` teach the
construction the owner objected to — is in `LANDMINES.md` so a session meets it before acting, which
was the whole point of declining a code gate.
