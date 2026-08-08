# 221 — `as an ai` convicts legitimate copy ABOUT AI, so a page that sells AI tooling cannot be rebuilt

**Filed 2026-08-08 by the loancalculator voice-H lane, found while measuring the
fix for `bugs_open/219`.** Same check, same severity, same permanent consequence
— **different mechanism, which is why it is a separate bug and was deliberately
not fixed inside 219's change.**

219 was about the check's **scope** (it scanned the whole artefact, so a code
comment convicted a page). That is fixed. This one is about the check's
**pattern precision**, in text that is genuinely visible prose, where no
re-scoping can help.

## The defect

`metaCommentaryPatterns` (`platform/orchestration/actions/validate_page_content.go:1216`)
contains `{"as an ai", "first-person AI disclosure"}`, matched as a
case-insensitive **substring**. It is intended to catch a model disclosing
itself — *"As an AI, I cannot…"*.

It also matches every ordinary English sentence in which "AI" follows "as an".
Live, on webdesign.co.uk:

```
page   tools-index   slot ported-page   (site: webdesign.co.uk)

  <p class="index-subtitle">LocalBusiness schema, as an AI-builder prompt</p>
```

That is a **product description on a tools index** — the tool emits a prompt for
an AI builder. It is correct, deliberate, customer-facing copy. The check reads
`as an AI-builder` → matches `as an ai` → **blocker**.

A word boundary does **not** fix it: `\bas an ai\b` still matches `as an
AI-builder`, because `-` is a boundary. The pattern needs first-person context
(what it was always for), not tighter tokenisation.

## Measured

```sql
-- 1,244 page_components rows on active pages, scanned for the multi-word
-- meta-commentary patterns. Positive control: 268 rows match 'calculator',
-- so a zero here would have been a real zero.
   rows_scanned | multiword_pattern_hits | positive_control
           1244 |                      1 |              268
```

The single hit is the row above. **Verified against the code, not inferred**:
the new (post-219) `checkMetaCommentary` was run over that row's real stored
`rendered_html` (12,879 bytes) and still returns a blocker —

```
STILL A BLOCKER: value="as an ai" location="calBusiness schema, as an AI-builder prompt"
```

— which is the whole point of filing it separately. 219's fix moves the scan to
assertion text, and this copy **is** assertion text.

## Why it matters

- **Same permanence as 219.** A blocker makes `validate_page_content` return an
  error, so the step fails and the page is never saved. `tools-index` cannot be
  rebuilt while that sentence is in its copy.
- **It has not fired yet, and that is not reassurance.** `agent_error_log` has
  **0** meta-commentary rows for webdesign.co.uk (all 6 fleet-wide are
  loancalculator's, 2026-08-08). Nobody has asked for a content change on that
  page since the copy landed. This is 219's "silent-until-you-try" shape exactly
  — the failure is latent, not absent.
- **It targets the sites most likely to write it.** Any page whose subject is AI
  tooling — the estate builds several — will reach for "as an AI assistant",
  "as an AI-builder prompt", "as an AI product". The check is most likely to
  fire where the copy is most likely to be right.

## Fix candidates, ordered by what closes the door

1. **Require the first-person construction the pattern was written for.** Match
   `as an ai, i `, `as an ai i `, `as an ai language model`, `as an ai model`,
   `as an ai assistant, i ` — i.e. the disclosure, not the noun phrase. Makes the
   legitimate use unrepresentable as a hit rather than relying on copy avoiding a
   word. ⚠ Decide explicitly what to do about `as a language model`, which has
   the same shape but far less innocent usage.
2. **Drop the blocker severity for the first-person family, keep it for the
   schema/pipeline family.** The schema vocabulary (`input_schema`, `on_missing`)
   genuinely never belongs in copy; the AI family is a judgement call about
   English. A warning still surfaces it without disabling the page.
   Weaker than 1 — it leaves a real apology shipping — but it is the smaller
   change and it removes the permanence, which is the actual damage.
3. Reword the page. **Rejected as the fix** — same reason 219 rejected editing
   the templates: it treats the instance, leaves the class, and asks correct
   customer copy to bend around a scanner. Reasonable as a *workaround* if that
   page must be rebuilt before 1 or 2 ships.

## How to verify a fix

- The negative control is the whole test here: **`As an AI, I cannot generate
  this listing.` must still be a blocker.** Induce it in a prose row; do not
  assume the family still fires. `TestMetaCommentaryStillBlocksVisibleCopy`
  in `validate_page_content_meta_scope_test.go` already carries the refusal case
  and is where the new one belongs.
- `checkMetaCommentary` over webdesign.co.uk `tools-index`'s stored
  `rendered_html` must return **0** issues (it returns 1 today — the recipe is in
  this file's Measured section, and it is a real artefact, not a fixture).
- Re-run the census above: still 1 row matched by the *old* patterns, 0 by the
  new ones, positive control still 268.

## Consumers to tell

The webdesign.co.uk lane owns the affected page. **The useful message is not
"we changed a regex" but "your `tools-index` cannot currently be rebuilt, and
here is the sentence".** Not yet sent — see the loancalculator lane's NOTES for
2026-08-08.

## Filing basis

First-hand, and the mechanism was induced rather than argued: the census could
have returned zero and did not, the matching row was read, and the **current
code was executed against that row's real bytes** to confirm the conviction
survives 219's fix. No `090` run: this is a two-line pattern-precision defect in
one function that I read and executed, not a cross-cutting structural claim —
the cross-cutting half (the scan's scope) was 219 and is fixed. If a fixing
thread wants to widen this to "how should the fleet's blocker-severity string
scans be governed", **that** is a `090` or an RFC, not this file.
