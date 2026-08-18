# PLAN — bugfix 303: the truncation guards count tag SUBSTRINGS; make them count MARKUP

**Workstream:** `bugfix_303_tool_birth_guard` · started 2026-08-18 · owner thread: session "303, bugs_open/298"
**Bug:** `bugs_open/303_HANDOFF_2026-08-18_tool_birth_guard_counts_tag_substrings_so_a_tool_that_manipulates_html_cannot_be_born.md`
**Landmine:** LANDMINES.md § "The tool-birth guard counts tag SUBSTRINGS…" (added 2026-08-18 by the filer)

## The defect, in one paragraph

Every truncation guard on component HTML decides "was this generation cut?" by
`strings.Count(folded, "<script") > strings.Count(folded, "</script>")` (and the same for
`<style`, `<section`, `<div`, `<fieldset`) — a raw substring count over the whole file. A tool
whose own JavaScript *mentions* a tag (a comment `// protect <style> blocks`, a regex
`/<script[^>]*>/`, prose) tips opens over closes and is refused as "truncated" — permanently,
with an error message its own context payload (`ends_cleanly: true`) disproves. Both live
HTML-manipulating tools pass with exactly zero margin today.

## Affected surfaces (all four verified at HEAD 2026-08-18 — one more than the bug file names)

| surface | file | exposure |
|---|---|---|
| `toolTemplateValid` | `plan_sections_action.go:1853` | tool BIRTH (create_tool_component) **and LOAD** — a failing component is silently dropped from the schemas map, so an affected tool can never re-render (bugs_open/024's shape) |
| `hasUnbalancedStructuralTags` | `component_write_guard.go:169` | section BIRTH (store_generated_component:343) |
| `componentRegressionIssues` check 2 | `component_write_guard.go:219` | REWRITE gate — comparative, so it fires exactly when a fix ADDS a legitimate mention (the "worse for fixes than for first builds" case) |
| `unterminatedTagPairs` | `discovery_checks/check_truncated_component.go:95` | fleet sweep → needs_human_review items; also its VERIFIER (an item can never resolve) and `newestIntactVersion` (calls intact versions damaged) |

The fourth surface (the discovery check) *mirrors* the pair list because `actions` imports
`discovery_checks`, so the reverse import is a cycle — a hand-maintained mirror guarded by a test.

## Decision: fix candidate 1 (+3's honesty +4's wording), at the framework level

Bug file candidates, and what we take:

1. **Strip JS/CSS regions and HTML comments before counting** — TAKEN, as the core.
2. Count only tag-shaped tokens (delimiter after the name) — TAKEN, folded into the same
   scanner (also fixes `<div` matching `<divider`).
3. Recoverable failure with the actual signals — TAKEN as: per-pair counts + `ends_cleanly`
   in the error context payload; routing (workflow error_step → needs_human_review) unchanged.
4. Correct the wording — TAKEN: "structurally incomplete: <script> opened N, closed M
   (counting markup only)" instead of asserting "the generation was cut".

**Where the shared code lives: `platform/content` (new file `markup_balance.go`).** Reasons:
- It is a LEAF package (stdlib imports only), already imported by the birth-path action
  (`create_tool_component_action.go` imports it for `HasToolDocHeader`), and importable by
  `discovery_checks` with no cycle — which **deletes the mirror** and its drift risk, the exact
  drift class the council reviews for.
- Its own header says it is the home for pure content/markup functions shared across the render
  path and validation ("place wherever … so the render path imports it without a cycle").

## The scanner (semantics, so the council can judge it without reading Go)

Case-insensitive single pass. Counts an open/close for the five structural pairs ONLY when the
token is **tag-shaped** (the name is followed by whitespace, `>`, or `/`) and **in markup
context**:

- `<!--` … `-->` comment bodies are skipped entirely (unterminated comment ⇒ rest of input
  skipped; tag balance before the comment still counts).
- `<script`/`<style` open tags are counted, then the RAW-TEXT BODY is skipped up to the first
  case-insensitive `</script`/`</style` — which is exactly where a browser ends the element, so
  a `</script>` inside a JS string is treated as the close **because the browser treats it so**.
  No close found ⇒ the open stays uncounted ⇒ imbalance, which is precisely the mid-JavaScript
  truncation signature the guard exists for (bugs_open/012/046).
- Inside an open tag, `>` inside a quoted attribute value does not end the tag.

Why this keeps every true positive: a generation cut mid-JavaScript leaves `<script` with no
`</script` anywhere after it — the scanner reaches end-of-input still inside the body and the
open is left unmatched. A cut mid-markup leaves a counted `<div`/`<section` open. A cut inside
a comment is invisible to tag-balance under BOTH old and new counting (no regression), and the
tool path's `endsCleanly` still catches it.

## Calibration duty (the guard header's standing instruction — RE-RUN THE SIMULATION)

Before shipping, against live `clients_db`:
1. **Absolute predicates** over every `content_components` row and every `component_versions`
   row: old vs new flag sets. Every row that FLIPS is inspected by hand. Expected flips:
   false positives only (mentions in JS). The 9 known 046 casualties and the 3 open
   `truncated_component` queue items must STAY flagged.
2. **Comparative simulation** over consecutive `component_versions` pairs (the header's own
   query): the 012 wrecks must still block; no new blocks on legitimate transitions.
3. **The bug's verification recipe:** `c0cfb873`'s stored template + injected
   `// protect <style> and <script> blocks` ⇒ new predicate passes (old fails — that is the
   bug); the same template truncated mid-`<script>` ⇒ both fail.

## Edits (≤8, for the council submission)

1. NEW `platform/content/markup_balance.go` — `StructuralTagPairs` (canonical),
   `StructuralTagCounts`, `UnbalancedStructuralTags`.
2. NEW `platform/content/markup_balance_test.go` — scanner unit tests incl. the bug's recipe
   shapes and the negative controls.
3. `platform/orchestration/actions/component_write_guard.go` — `balancedPairs` /
   `hasUnbalancedStructuralTags` / check 2 delegate to `content`; messages updated to name
   markup-context counts.
4. `platform/orchestration/actions/plan_sections_action.go` — `toolTemplateValid` delegates.
5. `platform/orchestration/actions/create_tool_component_action.go` — refusal message states
   the measured signals; context payload gains per-pair counts.
6. `platform/orchestration/actions/discovery_checks/check_truncated_component.go` — import
   `content`, delete the mirror list; mirror test becomes "uses the canonical list".
7. Test updates: `component_write_guard_test.go` (+ the adds-a-mention rewrite case),
   `check_truncated_component_test.go`.
8. `store_generated_component_action.go:345` — message wording only.

## Decisions and their reasons

- **Error code `tool_birth_truncation_blocked` KEPT.** It is queried by name (LANDMINES, bug
  file, agent_error_log history); renaming it breaks every existing query for zero gain. The
  honest part is the message and payload, not the identifier.
- **`endsCleanly` untouched.** Its calibration (tool-only, 36 FPs fleet-wide as a sweep) is
  documented and not implicated by this bug.
- **The discovery check's min-length stub tolerance (100) stays where it is** — behaviour
  preserved, only the counting changes.
- **No RFC.** This narrows what an existing guard REFUSES (precision fix); it changes no shared
  guarantee, adds no namespace/contract, no new authority (OWNER RULING 2026-07-29 §1). Council
  gate: yes — platform code.
