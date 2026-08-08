# 219 — `checkMetaCommentary` scans HTML COMMENTS, so a developer note in a component template permanently blocks that page from ever being rebuilt

> **FIXED 2026-08-08, commit `744bfdb3d` — but the title above is WRONG and so was
> this file's leading fix candidate. Read the correction in §"What the title got
> wrong" before you use anything here.** Short version: the notes are **JavaScript
> `/* … */` comments inside `<script>`, not HTML comments**, and two of the three
> templates contain no HTML comment at all — so "strip HTML comments before
> scanning" would have shipped **inert** and the three pages would still be blocked.
> The defect is the scan's **scope**, not comment syntax. The fix re-scopes the
> check to assertion text (`datahelpers.ExtractAssertionText` + `headProseBlocks`),
> the seam `bugs_open/218` settled for the sibling placeholder scan.
> **STAYS OPEN until the fleet rolls past `744bfdb3d`** — the defect is reproducible
> until the image ships (fleet was at `v1.0.1264`; the fix needs `v1.0.1265`).

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

## What the title got wrong (CORRECTION 2026-08-08, by the same lane, before fixing)

**The offending notes are not in HTML comments.** They are JavaScript block comments
inside the tool's own `<script>`. Measured on all three components, live:

```
        function         | hit_pos | last_script_open | next_script_close | last_html_comment_open
 tool-car-finance-pcp-hp |    8048 |             5379 |             11603 |   (none)
 tool-loan-repayment     |    8224 |             5019 |              9568 |     3560  <- before the script opens
 tool-rate-stress-test   |    5692 |             3884 |              7303 |   (none)
```

Every hit lies between a `<script>` open and its `</script>` close. Two templates
contain no `<!--` at all, and `tool-loan-repayment`'s sole HTML comment is at 3560,
**before** its script opens at 5019 — so the hit is not inside it either.

**Therefore fix candidate 1 below, "strip HTML comments before scanning", would have
unblocked NONE of the three pages.** It would have passed review, shipped, and left
the bug exactly where it was — with the fix marked done. What made this visible was
running the extraction as a query (`regexp_matches(html_template, '<!--(.*?)-->')`)
and getting **0 meta-commentary hits** where the whole-template scan got 3. The
disagreement between two counts is the finding; either alone reads as confirmation.

I wrote candidate 1 in this file, so this is my own error, and it is the same shape
as the one recorded in §"CORRECTION, same session" above: **the evidence said
"comment", and I did not ask which kind of comment.** "Inside a comment" was true and
useless. The fix candidate has to name the mechanism, not the appearance.

## Fix candidates, ordered by what closes the door

> **SUPERSEDED — see the correction above. 1 was inert; the shipped fix is 2's
> principle applied via the existing prose extractor.**

1. ~~**Strip HTML comments before scanning** in `checkMetaCommentary` (and audit the
   other whole-HTML scanners for the same assumption). Makes the bad state
   unrepresentable: a comment can never again be read as copy.~~ **INERT — would have
   fixed nothing. The comments are JS, not HTML.** ⚠ Deliberately keep
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

## What was actually shipped (`744bfdb3d`, 2026-08-08)

`checkMetaCommentary` now scans **assertion text** —
`datahelpers.ExtractAssertionText(html)` plus `headProseBlocks(html)` — instead of the
raw assembled page. That is candidate 2's principle ("grade the copy, not the
artefact") reached through machinery that already exists and is already trusted at
blocker severity in this same file: `bugs_open/218` re-scoped the sibling placeholder
scan to it this morning, after a council REVISE said *reuse it, do not add a second
stripper*. Its non-assertion set is `script, style, noscript, template, code, pre,
svg, iframe, textarea, select, option, head`, its walk skips comments and doctypes,
and attributes are never assertion text — so **every** non-prose context goes out of
reach at once, not just the one that happened to bite.

`headProseBlocks` is appended deliberately: `<head>` is a non-assertion context, but
`<title>` and the meta descriptions are prose a visitor reads. Without them this would
have been a narrowing rather than a re-scoping — and it is the exact escape the
placeholder scan shipped without and had to reopen hours earlier (`35889819c`).

**Proven on the real bytes, not on a fixture.** The three runs that failed this
morning still hold their exact input in `collected_data` (unpruned, ~24h window):

```
orchestration_states                    page                            page_html   old code      new code
fbd3da9d-f9d9-42d4-8b68-3a8e6d4ce788    index                           14,436 B    1 blocker  →  0 issues
0752258f-b03b-4a5a-9f46-61a67bb2d15d    tool-interest-rate-stress-test  10,475 B    1 blocker  →  0 issues
01072bcf-88e1-46d1-b738-db590e8989a2    tool-car-finance-calculator     14,147 B    1 blocker  →  0 issues
```

The old-code half of that pair is production's own record, not a local replica:
`error = "step validate_content failed: … content validation failed: 1 blockers, 0
errors"` on all three. Each artefact still contains `input_schema` — asserted in the
test, so a truncated or wrong file cannot masquerade as a pass.

**The negative control is induced, not assumed** (`validate_page_content_meta_scope_test.go`):
`input_schema` in a `<p>`, `on_missing` in an `<li>`, refusal prose, `as an ai` in a
meta description and `the data schema` in a `<title>` are each still blockers, each
with a location.

**Measured widening.** Assertion blocks collapse whitespace, so a multi-word pattern
split across a line break now matches where the raw substring scan did not. On live
content it convicts nothing new: 1,244 `page_components` rows on active pages give
**1** collapsed hit and the **same 1** raw hit (positive control: 268 match
`calculator`), and 585 active pages give **0** title/meta_description hits (positive
control: 122). `<title>`/meta were already in the old whole-HTML scan's reach, so they
are not new surface.

## Council verdict — APPROVED round 1, and the one objection worth answering

`Council-Reviewed: c9104844-b303-43dd-a426-73386ebbb25e` — *approved with 2 advisory
objections, none high-severity*; 6 seats abstained, `gated_by_truncation: false`.
`reuse_agent` approved it explicitly as reuse-over-reimplementation (the 218 REVISE
precedent), `editquality` confirmed the causal path and that `221` was "correctly
scoped out rather than papered over".

**The objection that had teeth**, raised by `bug_historian` at medium and echoed by
`guidelines`: `ExtractAssertionText` excludes `<script type="application/ld+json">`,
this file already carries a landmine that the banned-claims sweep misses JSON-LD, and
**I measured `<title>`/meta but never measured JSON-LD** — so I had created a possible
blind spot in the same shape the file has already been burned by, and asked a human to
sign it off. That is a fair hit: the measurement was missing, not merely unpersuasive.

**Answered by measurement, both directions, each with a control that could have come
out otherwise:**

1. **There was no JSON-LD coverage to lose.** Of the 37 assembled pages still in
   `collected_data` — the validator's *actual input*, `page_content.response.page_html`
   — **0 contain `application/ld+json` at all**. Positive control: 19 of the same 37
   contain `<script>`, so the query finds scripts when they are there.
2. **The mechanism says why**, so the zero is not a sampling accident: JSON-LD is
   appended to `<head>` at **render** time, after content validation —
   `rerender_single_page_action.go:931` and `data_helpers.go:1505`
   (`doc.Find("head").AppendHtml(…<script type="application/ld+json">…)`). It is not in
   the string this check has ever scanned. Independently corroborated by a dated
   measurement already in the code: *"none: measured 2026-07-28, ZERO of 14 live sites
   emitted any application/ld+json"* (`rerender_single_page_action.go:602`).
3. **And where JSON-LD does exist, it carries none of the vocabulary.** Across the 27
   `ld+json` blocks stored in `page_components.rendered_html` on active pages (3
   sites): **0** match any meta-commentary pattern. Positive control: 25 of the 27
   match `schema.org`.

So the re-scoping removed no JSON-LD coverage from this check, because it never had
any. **The landmine the seat cited is real and stays real for `banned_claims`** — that
check runs on a different surface — and nothing here changes it.

The second objection (low): meta-commentary inside `<code>`/`<pre>`/`<textarea>` or
attributes now passes **silently**. Accepted as the deliberate trade, stated in the
submission's risk 1 and unchanged by the above: a model apologising inside a `<pre>`
is rarer than a human documenting a template, and only one of the two can permanently
disable a page.

## What this fix does NOT fix — `bugs_open/221`

That single live hit is a **different defect**, found while measuring this one:
webdesign.co.uk's `tools-index` carries the copy *"LocalBusiness schema, as an
AI-builder prompt"*, and `as an ai` matches it in genuinely **visible prose**. The new
check was run over that page's real stored HTML and **still returns a blocker** — no
re-scoping can help, because the string really is copy. That is pattern precision, not
scan scope, and loosening the pattern is a false-negative decision that deserves its
own evidence. Filed as `221`; deliberately not folded in here.

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
