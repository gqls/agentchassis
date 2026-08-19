# 327 — a PARTIAL write to `content_direction` silently shrinks the brief the writer reads, and the document keeps growing so nothing looks wrong

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

**The same ordering is in the adoption path** — `apply_adoption_plan_action.go:280`
formats `directionData` (what this run produced) and never sees the merged result.

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

1. **Compute `formatted` from `merged`** (`site_spec_actions.go`, move the block from
   :212 to after :247; same in `apply_adoption_plan_action.go`). Removes the failure for
   every future write. **A shared platform seam** — RFC/council per CLAUDE.md, and it
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
