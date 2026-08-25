# 377 — the placeholder pattern `"your company"` convicts ordinary B2B prose; every recorded firing is a false positive, and it serially blocked finetuning.uk's builds for three weeks

**Filed 2026-08-24 by the finetuning_uk_service lane.** Fix committed same day
(one pattern-list line + two regression tests); ~~INERT until a post-fix chassis
roll — check the pod's build-provenance ancestry before believing it live.~~
**CLOSED 2026-08-25 — fixed AND live; moved to `bugs_closed/`.**

> ## ✅ CLOSED 2026-08-25 — proven at the ARTEFACT, not at the tag
> Both halves re-verified first-hand this session by the filing lane, and the proof is the
> served page rather than a pod probe, because this defect's whole signature is a build that
> *fails*:
> - **The literal is gone at HEAD.** `validate_page_content.go:141` now carries only the
>   comment recording the removal; `grep -n '"your company"'` returns the comment line and no
>   pattern entry `[MEASURED 2026-08-25]`.
> - **The motivating case re-drove and SHIPPED.** `/your-own-model.html` is
>   `build_status=deployed`, `deployed_at = 2026-08-24 19:58:47Z` — **after** the 18:32Z roll
>   that carried the fix — and the live page serves the exact sentence the pattern convicted:
>   `Your company's voice, in a model you own` (HTTP 200, 38,194 bytes, `[MEASURED 2026-08-25]`).
>   A build carrying that sentence could not have completed with the pattern still in the list;
>   the 08-24 blocker row names that very string.
>
> **Residual left deliberately open** (unchanged from "what is owed", step 3): the other
> prose-plausible entries — `"coming soon"`, `"not provided"`, unbracketed `"tbd"` — were never
> measured and are not this fix's scope. Census them the same way before touching them.
> ⚠ Note also that this file has since gained a **new** neighbour in the same function: the
> `bugs_open/387` numeric-stand-in detector (`29ca1b35e`…`5aed9ca74`) adds blocker patterns to
> `checkPlaceholderPatterns`. A future false-positive here is more likely to be 387's shapes
> than this one's.

## The mechanism

`platform/orchestration/actions/validate_page_content.go` carries a static
`placeholderPatterns` list; every hit is `severity=blocker` and fails the build
(`content validation failed: N blockers`). The entry
`{"your company", "placeholder prompt"}` is a bare substring match, so it fires
on any second-person B2B sentence — "your company data stays private", "what AI
can do for your company", "fill in the form with your company details". It was
presumably meant for template residue ("[Your Company]", "Your Company Name
Here"), but the bracketed shapes have their own entries (`[your `, `[company`)
and the bare form matches the copy this estate is PAID to write.

The list's own comment already records this class once: bare `"placeholder"`
was removed for convicting `<input placeholder="...">`. `bugs_open/218` is the
adjacent-but-different defect (the scan read CODE as prose; fixed via
`ExtractAssertionText` scope). 377 is a pattern that convicts genuine PROSE, so
scope-fixes cannot help it.

## The evidence (all `[MEASURED 2026-08-24]`, `agent_error_log`, error_code `CONTENT_VALIDATION_BLOCKER_DETAIL`)

- **46 recorded firings of the pattern, 2026-08-03 → 2026-08-24** (the detail
  row exists since ~08-03, so this is the pattern's whole observable life),
  across 6 domains. The disconfirming result was available: a genuine template
  residue in the sample. **All 46 locations were read; every one is ordinary
  prose.** The nearest to a true positive is webdesign.co.uk's article that
  DESCRIBES placeholder headings (`…older heading like "Your Company Name", and
  the layout…`) — prose about placeholders, not residue.
- **41 of the 46 are finetuning.uk** — its privacy-first copy says "your
  company('s) data" constantly, and its owner-ratified proposition literally
  opens "Your company's voice, in a model you own". The 08-24 firing blocked
  the £99 offer page's first build (work item
  `gap_plan_new_your-own-model_1368e337…`, parked `needs_human_review`); the
  40 before it are three weeks of serial re-blocks of content builds on the one
  site whose message the pattern matches.
- The blocker detail row (10:26:20) names it exactly:
  `type=placeholder_text, value="your company", location="Your company's voice,
  in a model you own", severity=blocker`.

## Why no 090 run (first-hand verification substituted, per the 2026-07-31 ruling)

The root cause is one static literal, verified four ways: the line itself
(`validate_page_content.go:141` pre-fix); the live blocker row naming the
pattern and label; the 46/46 false-positive census read location by location;
and the test suite reproducing both sides of the fix. There is no hidden
mechanism for a diagnosis loop to find — the entire cause fits in one line and
its evidence is already cited above.

## The fix (committed with this file)

Remove the bare pattern; keep every guard that catches real residue:

- `{"your company", …}` → **removed**, with a comment recording the census and
  mirroring the bare-`"placeholder"` precedent.
- `{"your company name here", …}` → **added** (the unambiguous prose-form
  residue).
- Regression pair in `validate_page_content_placeholder_test.go`:
  second-person possessive prose must NOT convict (killed by re-adding the bare
  pattern); "Your Company Name Here" must still convict.

Verified against `git archive HEAD` + only these two files: full
`platform/orchestration/actions` package passes. (The working tree's
concurrent `TestUpdateWorkItemStatus` failures belong to another session's
in-flight `load_work_item_actions.go` change — measured by the same
archive-HEAD baseline passing them.)

## How to verify live, and what is owed after the roll

1. Chassis provenance ancestry: `git merge-base --is-ancestor <this commit>
   <pod's stamped sha>`.
2. Re-drive the motivating case (your-action-can-silence-your-own-detector:
   this fix is proven only when the case that paid for it passes): reset
   finetuning.uk's offer-page item `gap_plan_new_your-own-model_…` from
   `needs_human_review` to `triaged` (attempt_count is untouched at 0) and let
   the build-pipeline sweep it. The page carried FIVE further "your company"
   sentences in its written copy, so a re-block would be immediate and loud if
   the fix did not ship.
3. Residual to leave alone deliberately: other prose-plausible entries remain
   in the list (`"coming soon"`, `"not provided"`, `"tbd"` unbracketed). They
   were NOT measured here and are not this fix's scope; census them the same
   way before touching them.

---

> **CORRECTED 2026-08-24, same session:** §"what is owed after the roll" step 2 said
> the page carried *"FIVE further 'your company' sentences in its written copy"*.
> **Wrong — the written copy carries ZERO.** The five sentences quoted in the census
> came from OLDER firings (08-03..08-16 rows on other finetuning.uk pages); I read
> them as belonging to today's page. Measured against the writer's actual output
> (preserved in the build orchestration's `page_content.response.page_html`,
> `llm_call_log 774ca9c5` 10:24:23): 0 occurrences in the written prose; the one
> conviction was the ASSEMBLED hero headline ("Your company's voice, in a model you
> own" — sourced from the seeded proposition, not from the writer's sections). The
> re-drive after the roll is still the right verification — the hero line alone
> re-blocks if the fix did not ship — but it is one loud sentence, not six. What
> caught it: running the register checks on the preserved copy instead of trusting
> my reading of the census locations. The cheap check: filter the census by
> `occurred_at`/orchestration before attributing rows to a specific build.
