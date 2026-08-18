# 303 — the tool-birth truncation guard counts tag SUBSTRINGS, so a tool whose own code mentions a structural tag cannot be born — and it is rejected with "the generation was cut", which is false

**Filed 2026-08-18** by the `webdesign_tool_rebuilds` lane, which hit it building an HTML minifier.
~~**Status: OPEN, UNOWNED. Live. It is blocking real work right now.**~~
**Status: FIXED AT SOURCE 2026-08-18 (`6d962bcf8` + `e21b172f0`), OPEN until a chassis image rolls** — Go changes are inert until then, so the birth refusal is still live in the fleet today and the LANDMINES workaround (tag names without angle brackets in `add_tool` descriptions) still applies. Owner: session "303, bugs_open/298", lane `bugfix_303_tool_birth_guard`. Council: `Council-Submitted: 70cf0da5-e91a-42f0-8dd6-0cb5710b51dc`. Fix record at the bottom of this file.

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

## FIX RECORD 2026-08-18 — fixed at source (session "303, bugs_open/298"), OPEN until a chassis roll

**The fix is candidate 1 (+3's honesty, +4's wording), taken at the framework level.** One shared
markup-context scanner — `platform/content/markup_balance.go`: `StructuralTagPairs` /
`StructuralTagCounts` / `UnbalancedStructuralTags` — counts a token only when it is tag-shaped
(name then whitespace/`>`/`/`) and in markup context: `<!-- -->` bodies skipped, `<script>`/`<style>`
raw-text bodies skipped to the first close token, exactly where a browser ends the element. A cut
mid-JavaScript still leaves `<script` unmatched, so the true-positive class (012/021/046) is kept.

**Where it landed** (commits `6d962bcf8` and, for the shared file, `e21b172f0` — that hunk rode the
309 lane's commit as a declared same-file passenger by agreement between the two sessions):

- `toolTemplateValid` (birth AND load — the load half was worse than this file knew: a failing tool
  was silently dropped from the schemas map, the `bugs_open/024` shape again),
- `hasUnbalancedStructuralTags` (section birth) and `componentRegressionIssues` check 2 (rewrite),
- `check_truncated_component` (sweep + verifier + intact probe; the hand-mirrored pair list is
  RETIRED — the leaf package ends the import cycle that forced it),
- `check_tool_completeness` (advisory, tool-recreation flow; was also case-SENSITIVE),
- `store_generated_component` Check 2 + the five-pair birth message.

**The refusal is honest now** (candidates 3+4): it states the measured signals — per-pair
markup-context counts and `ends_cleanly`, both in the message and in the `agent_error_log` context
(`tag_imbalance`) — and points at `llm_call_log` output_tokens vs max_tokens instead of asserting
"the generation was cut". `error_code='tool_birth_truncation_blocked'` is KEPT (queried by name).

**Calibration** (the guard header's standing instruction, re-run 2026-08-18 over the full live DB):

- `component_versions`, 264 rows: **26 flagged by BOTH old and new, 0 verdict flips** — every
  recorded casualty kept. Comparative check-2 over all 121 consecutive version pairs: 1 block under
  both (the 012 write), 0 disagreements.
- `content_components`, 300 rows: old flagged 11, **new flags 8 — a strict subset, zero new
  positives**. The 3 cleared rows were each hand-read and all are mentions inside CSS comments:
  `79c34359` Agent Protocol Tracker ("component's `<style>` block is unstyled…", line 3, inside the
  style block), `fc56f085` info-card-grid ("do not move this block above the base `<style>`. */",
  line 233), `71a54cc2` gauntlet-round-record-vonc-com ("Inline `<script>` means no /tools/assets
  publication step", line 13, inside the style block). All 8 still-flagged rows are inactive
  historical casualties; **no active component is a true casualty today**.
- **Two of the three cleared rows carry OPEN `truncated_component` work items that are FALSE ALARMS
  of the substring count**: `91007600` (info-card-grid, "intact v1 available to restore" — restoring
  v1 would have REPLACED A GOOD TEMPLATE) and `6e2c9ebf` (gauntlet-round-record, "needs
  regeneration" — regeneration risks fabrication per `bugs_open/020`). Left open deliberately: after
  the fix rolls, their verifier resolves them as balanced; nobody should act on their remedy text
  before then. The third open item's component (`c4f94a99` tool-llm-cost-calculator) passes under
  BOTH predicates today — repaired after filing; its item is stale, not false.
- **The verification recipe from this file, run on the real stored templates** (`c0cfb873`,
  `88b70065`): as stored, both predicates pass; with `// protect <style> and <script> blocks`
  injected, old refuses (this bug), new passes; truncated at 60%, BOTH refuse, new naming
  `<script`. Pinned as unit tests, including `TestSubstringCountWouldRefuseMentionTool`, which
  reproduces the OLD predicate so the fixture can never silently stop exercising the defect.

**The ADDENDUM class is covered too**, and is the strongest argument for region-stripping: in
`<script>const s=`+"`"+`<script type="…">…<\/script>`+"`"+`;</script>` both mentions sit in the
outer script's raw-text body, and the element balances exactly as a browser reads it — pinned as
the "tool that OUTPUTS a script tag" test. No concatenation obfuscation needed once this rolls.

**Council:** `Council-Submitted: 70cf0da5-e91a-42f0-8dd6-0cb5710b51dc` (verdict recorded here when read).
**Register:** CLC-019. **Verify after the roll:** re-run an `add_tool` whose description names tags
with angle brackets; confirm birth; confirm the two false-alarm items resolve via their verifier.
