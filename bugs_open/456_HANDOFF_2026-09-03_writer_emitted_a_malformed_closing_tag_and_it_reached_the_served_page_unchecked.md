# 456 — the page-content-writer emitted a malformed closing tag (`</strom>`) and it reached the served page: nothing between the writer and the bucket checks HTML well-formedness

**Filed 2026-09-03 by the finetuning.uk lane, from a finding by the `copy_quality_two_stage` lane
(their after-pass on the technical-details rebuild, commit `2cb6cfb43`).** Status: OPEN, live.

## 1. The defect, at the artefact

`https://finetuning.uk/technical-details.html`, rebuilt 2026-09-03 10:38:58Z (page-content-writer
orchestration `ce514ce0`, correlation `6e8eadaa`), serves:

```html
<strong>open-weight model</strom>
```

`[MEASURED 2026-09-03 10:50Z]` `grep -c '</strom>'` on the served page = **1**; the same string is
in the stored `page_components.content_data` (copy lane, grep = 1 on both). The negation gate
(`rewrite_negations`) preserved it faithfully in both its `from` and `to`, so it is writer output,
not a repair fault. Browsers ignore the unknown closing tag, so the `<strong>` stays open until the
parser closes it at the end of the block: the visible symptom is bold text running on to the end of
the paragraph. Minor on this page; the class is not minor (an unclosed `<a>` or `<div>` from the
same source reflows or hides everything after it).

## 2. What the save seam DOES check, and that none of it is well-formedness

Read first-hand in `platform/orchestration/actions/save_page_sections_action.go` and the guards it
calls, 2026-09-03 — this is the scope of the claim, no wider:

| guard on the save path | what it refuses | catches `</strom>`? |
|---|---|---|
| `sanitizeSectionsContentData` (`content_data_envelope_guard.go`) | an LLM **transport envelope** left in `content_data`, or one whose payload disagrees with a sibling key (`bugs_open/190`) | no — it never looks inside a string |
| completeness floor | missing required fields | no |
| `bugs_open/194` refusal state | a section refused by the writer | no |
| section shrink floor (`bugs_open/422`) | a section that shrank past the floor | no (and see §4) |

No call to `golang.org/x/net/html`, no `bluemonday`, no tag-balance check on that path
(`grep -n 'html.Parse\|bluemonday\|sanitiz\|well-formed' save_page_sections_action.go
rewrite_negations_action.go` → the envelope guard only). The name `sanitize…` is what makes this a
trap: a reader who greps for a sanitiser on the save path finds one and stops.

`[UNVERIFIED]` whether any LATER step (render/assemble, deploy, or a discovery checker such as
`discovery_checks/check_structure_floor.go`, which does import `x/net/html`) would flag it after
the fact. Empirically nothing did before the page served, and nothing has filed an item against
it in the 15 minutes since (no `site_work_items` row names `technical-details` with a
markup-shaped `item_type`). A thread that wants the structural claim ("no check anywhere") should
put it through `090_TRIGGER_needs_diagnosis`; this file asserts only the save seam, which was read.

## 3. Fix candidates, ordered by what closes the door

1. **Normalise at the save seam, structurally.** Parse each string-typed `content_data` field that
   may carry markup with `x/net/html` (the tolerant HTML5 parser) and re-serialise. That turns
   `<strong>open-weight model</strom>` into `<strong>open-weight model</strong>` and balances every
   other tag the writer can misspell, with no rule list to maintain. Cheap (the dependency is
   already in the module: `section_visible_text.go`, `check_structure_floor.go`), and it makes the
   bad state unrepresentable in what is saved. Log a `markup_normalised` event with the diff so
   the writer's tic is visible (the same instrument argument `rewrite_negations` makes for its
   rejection log). **Risk to measure first:** re-serialising can also change things the templates
   rely on (entity forms, self-closing `<br>`), so run it over a sample of existing `content_data`
   and diff before enabling; a change on well-formed input is a defect in the normaliser.
2. Refuse the section at the save seam when the parse-and-serialise output differs, and let the
   writer's retry path re-generate. Louder, costs a model call, and inherits 422's shape (a refusal
   that can make a page unrebuildable), so only if 1 is judged too permissive.
3. A discovery check after deploy (a `component-render-check`-style CronJob). Detects, does not
   prevent, and the page has already served.

The 2026-08-04 ruling applies: **do not hand-edit the served page or the row.** The page will be
rewritten at `bugs_open/443` Stage B once migration 641 applies; until then the defect stands as
the live reproduction.

## 4. Adjacent observation, NOT a claim (copy lane, same run)

The rebuilt page holds 1,214 words against the previous 1,826 (55 → 47 sentences, same four
components). Summing every gate rewrite accounts for **36** words; the writer simply wrote a
shorter page. The 422 shrink floor did not trip. Whether a one-third shrink on a rebuild is a
defect or a normal writer variance is a question for the `bugs_open/422` file, not this one; it is
recorded here because anyone sizing a shrink budget on the repair side would be looking at 6% of
the volume.

## 5. How to verify a fix

- Unit: feed the normaliser `<strong>x</strom>`, `<p>a<b>b</p>`, `<a href="/x">y` and assert the
  serialised output is balanced; feed a well-formed section and assert byte-identity.
- Live: rebuild technical-details after the fix (Stage B will do this anyway) and
  `curl -s https://finetuning.uk/technical-details.html | grep -c '</strom>'` → 0, with the
  `markup_normalised` log line present if the writer misspelt a tag again, absent if it did not.
  Control: the pre-fix page (this one) reads 1.

> **CORRECTED 2026-09-03 (same morning, copy lane):** §4's shrink is NOT a property of rebuilds.
> The homepage rebuilt in the same dispatch loop **grew** 1,804 → 2,042 words (+13.2%) against
> technical-details' −33.5%. Per-page writer variance, not a rebuild effect; the 422 file should
> not read §4 as general.
