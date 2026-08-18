# 303 — the tool-birth truncation guard counts tag SUBSTRINGS, so a tool whose own code mentions a structural tag cannot be born — and it is rejected with "the generation was cut", which is false

**Filed 2026-08-18** by the `webdesign_tool_rebuilds` lane, which hit it building an HTML minifier.
**Status: OPEN, UNOWNED. Live. It is blocking real work right now.**

## The one-paragraph version

`create_tool_component` refuses to persist a generated tool whose HTML fails
`componentTemplateValid(html, "tool")` → `toolTemplateValid`. That predicate is a raw substring
count — `strings.Count(folded, "<script") > strings.Count(folded, "</script>")` and the same for
`<style`, `<section`, `<div`, `<fieldset` — with **no parsing and no awareness of string literals,
regular expressions, code comments or visible prose.** A tool that *manipulates HTML* necessarily
mentions those tags in its own source. One unpaired mention — a comment saying
`// protect <style> and <script> blocks`, a regex `/<style[^>]*>/`, or body text explaining what the
tool does — makes opens exceed closes and the tool is refused **at birth, permanently**. The error
says *"refusing to persist truncated tool … the generation was cut; regenerate"*, which sends the
next reader after a truncation that did not happen.

## Evidence `[MEASURED 2026-08-18, live DB + source]`

- **The generation was NOT cut.** `llm_call_log` for the failing run: `output_tokens` **13,979**
  against `max_tokens` **32,000** (`claude-sonnet-5`) — nowhere near the ceiling.
- **The guard's own recorded context says the same:** `agent_error_log`,
  `error_code='tool_birth_truncation_blocked'`, context `{"html_length": 12248, "ends_cleanly": true}`.
  **`ends_cleanly: true` means the "ends mid-token" half PASSED** — the rejection came purely from
  the tag-balance half, on a document the guard itself agrees terminates properly.
- **The predicate, read from source** (`plan_sections_action.go:1853-1867`, pairs at
  `component_write_guard.go:146-152`):
  ```go
  for _, pair := range balancedPairs {
      if strings.Count(folded, pair.open) > strings.Count(folded, pair.close) { return false }
  }
  ```
- **Both live tools of this class pass with EXACTLY ZERO MARGIN** — measured over their stored
  `html_template`:
  | component | `<script`/`</script>` | `<style`/`</style>` | `<div`/`</div>` | `<fieldset`/`</fieldset>` |
  |---|---|---|---|---|
  | `c0cfb873` tool-html-minifier | 1 / 1 | 1 / 1 | 4 / 4 | 1 / 1 |
  | `88b70065` tool-svg-optimizer | 1 / 1 | 1 / 1 | 6 / 6 | 1 / 1 |
  They survived only because the generator happened to write its protect regex with an alternation
  group — `/<(pre|textarea|script|style)\b[^>]*>/` — in which the substring `<script` never occurs.
  **Phrase it as `/<script[^>]*>/` and the identical tool is unbornable.** Zero margin is the finding:
  this class is not comfortably passing, it is one word away from rejection at all times.
- **The calibration explains the gap, and is honest about its population.** The source comment:
  *"Calibrated against all 27 active tool components on 2026-07-20"* and (for the pair list)
  *"across every active component fleet-wide this 5-pair predicate flagged EXACTLY the 9 known
  truncation…"*. On 2026-07-20 **no tool on the estate manipulated HTML**. The first ones were created
  by this lane on 2026-08-17. The predicate was 0-false-positive against a population that could not
  contain its false-positive class.

## Why it matters beyond one tool

- **It makes a whole category unbuildable**: HTML/XML minifiers, formatters, escapers, sanitisers,
  template previewers, code-snippet tools, anything with markup in an example. Not site-specific.
- **It is worse for FIXES than for first builds.** A first build can be re-rolled until the model
  happens to phrase things safely (which is what produced the two live tools). Acting on a *reported
  defect* means asking for specific behaviour — and the more precisely you specify HTML handling, the
  more likely the implementation names a tag. This bug bites hardest exactly when a human has
  complained.
- **The message actively misdirects.** Two of this lane's builds failed at `save_tool` minutes apart
  with `SQLSTATE 23505` on two *different* constraints, and a third with this guard. All three read as
  "the same save_tool problem" at a glance. This one additionally asserts a cause (truncation) that
  its own recorded evidence (`ends_cleanly: true`) contradicts.

## Fix candidates, ordered by what closes the door

1. **Strip JS/CSS regions and HTML comments before counting.** The guard already has the concept —
   the tools it rejects are the ones that protect those very regions. Counting tags only outside
   `<script>`, `<style>` and `<!-- -->` removes the entire false-positive class and keeps every true
   positive (a mid-JavaScript cut still leaves an unclosed `<script`, which is the shape
   `bugs_open/046` was about).
2. **Count only tags, not substrings** — require the char after the name to be `>`, whitespace or `/`,
   and ignore matches inside quoted strings. Cheaper than (1), still parser-free, but does not help a
   regex like `/<div[^>]*>/`.
3. **Make the failure recoverable rather than terminal**: keep refusing the write, but route to
   `needs_human_review` with the *actual* signals (`ends_cleanly`, per-pair counts, output vs max
   tokens) instead of asserting truncation. Composes with 1 and 2 and is worth doing regardless — the
   current message cost this lane a wrong first hypothesis.
4. **At minimum, correct the wording.** "Structurally incomplete" is defensible; "the generation was
   cut" is a claim the guard has already disproved in its own context payload.

## How to verify a fix

Take `c0cfb873`'s stored template, insert `// protect <style> and <script> blocks` into its script,
and confirm `toolTemplateValid` returns **true** — today it returns false. Negative control: truncate
the same template mid-`<script>` and confirm it still returns **false**.

## Why this was NOT put through the `090` diagnosis loop (CLAUDE.md, owner ruling 2026-07-31)

Substituted equivalent first-hand verification, stated plainly as the ruling requires: I read the
deciding predicate and its calibration comment in source; measured `output_tokens` vs `max_tokens`
from `llm_call_log`; read the guard's own `ends_cleanly: true` from `agent_error_log`; and measured
the zero-margin tag counts across both live tools of the affected class. The cause is not inferred
from a symptom — the failing line is quoted above and its inputs are measured. A loop run would
re-read the same function.

## Related

- `bugs_open/046` / `bugs_open/021` INSTANCE 1 / `bugs_open/012` — the truncation class this guard
  was built for. **The guard is right to exist**; this is a precision defect, not an argument to remove it.
- `bugs_open/024` — the previous misclassification in the opposite direction (healthy tools ending
  `</script>` read as truncated), which is why `toolTemplateValid` was split out of `sectionTemplateValid`.
- `architecture_review/RFC_036` — the other structural blocker this lane surfaced (three uniqueness
  gates on one INSERT, one of which makes any tool un-rebuildable).
- `docs024_key_docs_latest/webdesign_tool_rebuilds/NOTES_…` 2026-08-18 — the full sequence.

## ADDENDUM 2026-08-18 — a second, STRICTLY WORSE class: tools whose OUTPUT is a script tag

Found filing `tool-seo-injector` (a JSON-LD generator). The class in the body above — tools that
mention a tag in a comment, regex or prose — can be dodged by rewording. **This one cannot be
reworded, because the imbalance is forced by JavaScript itself.**

A tool that emits a script tag must contain the opening text `<script type="…">` in its template, and
**must escape the closing tag** (`<\/script>`), because an unescaped closing tag inside an inline
script terminates that script — that is a hard language rule, not a style choice. So:

- `strings.Count(folded, "<script")` counts the opening literal, **and**
- `strings.Count(folded, "</script>")` does **not** match the escaped `<\/script>`.

⇒ opens exceed closes **by construction**, for every correct implementation. The only escape is to
assemble both tags from concatenated pieces so neither literal appears — which is what this lane now
has to write into the work item's description. **A guard that forces generated code into an obfuscated
form to pass is inverting its own purpose**: the concatenated version is harder to read and no less
likely to be truncated.

Affected beyond this tool: anything that outputs embedded markup — JSON-LD/schema generators,
analytics snippet builders, embed-code generators, CSP or meta-tag helpers, "copy this snippet" tools
of any kind. On a web-design site that is a natural and recurring product category.

**This strengthens fix candidate 1** (strip script/style/comment regions before counting): a tag
mentioned inside a template literal is exactly the case region-stripping handles and rewording cannot.
It also means candidate 4 (fix the wording only) is not sufficient — the message is misleading AND the
predicate blocks correct code.
