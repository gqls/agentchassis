# 219 — `checkMetaCommentary` scans HTML COMMENTS, so a developer note in a component template permanently blocks that page from ever being rebuilt

**Filed 2026-08-08 by the loancalculator voice-H rollout lane.** Not urgent — it
breaks no live page. It makes three pages **permanently un-rebuildable**, and it will
do the same on any site whose component templates carry the conscientious thing to
write: a comment explaining the template.

## The defect

`validate_page_content`'s meta-commentary check is a raw, case-insensitive substring
scan over the **entire assembled page HTML**:

```go
// platform/orchestration/actions/validate_page_content.go:332
issues = append(issues, checkMetaCommentary(htmlStr)...)

// :1230
func checkMetaCommentary(html string) []ValidationIssue {
	lower := strings.ToLower(html)
	for _, p := range metaCommentaryPatterns {
		idx := strings.Index(lower, p.Pattern)   // <- no comment stripping
```

`metaCommentaryPatterns` (`:1209`) includes `input_schema`, `on_missing`,
`skip_section`, `required: true`. Those are exactly the words a developer uses when
**documenting a component template** — and a template's documentation lives in an
HTML comment, which is part of `htmlStr`.

So the check cannot tell *"the model wrote about its task instead of doing it"* — its
own error message — from *"a human documented this template two weeks ago"*. It is a
**blocker** severity, so the page's build fails and never saves.

## Measured, with a control group that could have come out otherwise

On loancalculator.co.uk, 12 pages carry a locked tool component. **Exactly 3 active
pages have a tool template containing `input_schema`:**

```
index                            tool-loan-repayment       input_schema in template: t
tool-car-finance-calculator      tool-car-finance-pcp-hp   input_schema in template: t
tool-interest-rate-stress-test   tool-rate-stress-test     input_schema in template: t
tool-standard-calc  (archived)   tool-loan-repayment       input_schema in template: t
-- the other 8 tool pages: f
```

A voice rewrite was dispatched at **all 12**. **Those 3 failed. All 9 others passed.**
`tool-car-finance-calculator` was fired **three times** and failed identically each
time, quoting the same snippet — the failure is deterministic, not a model wobble:

```
LLM meta-commentary in content: 'input_schema' (schema vocabulary in copy)
location: ymbol come from the input_schema. FIXED 2026-08-03 — 0% APR, as its own
          change. The whole calc
```

That snippet is **`content_components.html_template` for `tool-car-finance-pcp-hp`** —
a developer changelog comment recording a real fix on 2026-08-03:

```
   3. Labels, defaults and the currency symbol come from the input_schema.

   FIXED 2026-08-03 — 0% APR, as its own change. The whole calculation used to
   be wrapped in `if (apr > 0)`, so an interest-free deal — a re…
```

**It predates the rewrite by five days and it already ships to readers** (1 occurrence
in the served `car-finance-calculator.html`, inside a comment). Nothing the writer
produced is involved. The prose baseline for that page does **not** contain the string
(checked before concluding).

> **CORRECTION, same session.** I first recorded this as *"the writer invented meta
> commentary"*, because that is what the validator's message asserts and I repeated it.
> Wrong. The `page-content-writer` prompt interpolates
> `{{.current_section.existing_content_html}}` and the content-direction fields — it is
> never handed the tool component's template. The predictor above (3 of 12, and exactly
> those 3) is what falsified my first reading. **The error message names a cause it has
> not established**, and it is convincing enough that I adopted it for an hour.

## Why this matters beyond three pages

- **It is permanent and silent-until-you-try.** The page builds fine until someone asks
  for a content change; then it fails every time, with an error blaming the model.
- **It punishes the good practice.** The 2026-07-31 landmine already tells authors that
  template syntax inside a comment breaks the parse; the natural response is prose
  comments explaining the template — which is what this check convicts.
- **It is a fleet-wide surface, not a site one.** Any `content_components.html_template`
  containing `input_schema`/`on_missing`/`skip_section`/`required: true` in a comment
  disables rebuilds for every page using that component. Census before fixing:
  ```sql
  SELECT function, count(*) FROM content_components
  WHERE html_template ~* '(input_schema|on_missing|skip_section|required: true)'
  GROUP BY 1;
  ```

## Fix candidates, ordered by what closes the door

1. **Strip HTML comments before scanning** in `checkMetaCommentary` (and audit the
   other whole-HTML scanners for the same assumption). Makes the bad state
   unrepresentable: a comment can never again be read as copy. ⚠ Deliberately keep
   scanning comments for the *first-person AI disclosure* patterns if a model emitting
   `<!-- as an AI -->` is judged worth blocking — that is a separate decision, and it
   should be made explicitly rather than inherited.
2. **Scan only the blocks this run WROTE**, not the whole assembled page. Stronger in
   principle — the check exists to grade generated copy — but a bigger change, and it
   would stop the check catching contamination that arrives via a template.
3. Strip the comments from the offending templates. **Rejected as the fix** — it treats
   the instance, leaves the class, and edits a locked calculator's template to satisfy
   a validator, which is the wrong direction of travel. Reasonable as a *workaround* if
   a page must be rebuilt before 1 ships.

## How to verify a fix

- Re-fire a `content_rewrite` at `tool-car-finance-calculator`; it must reach
  `save_page_sections` instead of failing `validate_content`.
- **Negative control, or the test proves nothing:** a page whose *generated prose*
  genuinely contains `input_schema` must still be blocked. Induce it — put the string
  in a prose row and re-run — rather than assuming the check still fires.
- The census query above returns the same rows before and after (the fix must not
  require editing any template).

## Filing basis

**CLAUDE.md requires a cross-cutting root-cause claim to go through the `090` loop, or
the filing session to state plainly why it substituted equivalent first-hand
verification. This is that statement.** Substituted, because the mechanism was
*induced five times* and then *predicted*: 12 pages were dispatched blind, the 3 with
`input_schema` in their tool template failed and the other 9 passed, which is a control
group that could have come out otherwise and did not. The check's source was then read
directly (`:332`, `:1230`, `:1209`) and confirms a raw substring scan with no comment
handling. What a `090` run would add is an independent reader; the symptom to file
would be *"validate_page_content's meta-commentary check scans the whole page HTML
including comments, so a developer note in a component template blocks the rebuild of
every page using that component"*.

**State of the three pages: unchanged and healthy.** The build failed before
`save_page_sections`, so their prose still carries its original bytes and row ids, and
all three serve HTTP 200. They are the only pages of loancalculator's 26 that the
voice-H rollout could not convert.
