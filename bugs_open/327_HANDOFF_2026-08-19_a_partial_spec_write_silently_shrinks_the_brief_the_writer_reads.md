# 327 — a PARTIAL write to `content_direction` silently shrinks the brief the writer reads, and the document keeps growing so nothing looks wrong

> ## ⚠ THE NUMBER 327 IS AMBIGUOUS — refer to this case BY SLUG
> Another lane filed `bugs_open/327_HANDOFF_2026-08-19_the_build_trigger_can_publish_nothing_and_exit_zero.md`
> the same day. Both are real, neither is renumberable (numbering is never reassigned), and
> `scripts/who-owns.py 327` now prints the ambiguity warning. This case is
> **`a_partial_spec_write_silently_shrinks_the_brief_the_writer_reads`**. `git log` the FILE PATH.

> ## ✅ FIXED IN CODE 2026-08-20 — `c9a71388f` — AND STILL OPEN, because it is inert until a chassis roll
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
