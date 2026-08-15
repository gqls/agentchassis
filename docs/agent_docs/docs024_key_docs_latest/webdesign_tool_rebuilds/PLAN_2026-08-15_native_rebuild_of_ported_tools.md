# PLAN — native rebuild of webdesign.co.uk's ported tools (owner-directed, 2026-08-15)

**Owner directive (2026-08-15, in-session):** "let's do both and speed up the native rebuilds"
— (1) the audit fix is owned by the `bugfix_281_tool_audit_ported/` lane (Track 1, rides the
next roll; not this lane's work); (2) THIS lane accelerates replacing the 63 ported tools with
native framework components. Decomposition (byte-faithful conversion) remains a separate,
decision-pending proposal with preconditions:
`bugfix_281_tool_audit_ported/PROPOSAL_2026-08-15_decompose_webdesign_tools.md` — this lane
does NOT execute it and must not be confused with it.

## The mechanism (all verified 2026-08-15, live DB + code)

- Native tools arrive via `add_tool` work items → build-dispatch-loop → **tool-generator**
  (novel build; spec carries `name/function/priority/complexity/description`) → deploy.
  Produced organically by `check_missing_tools` → tool-suggester (gap analysis — it ADDS
  missing tools; **nothing was replacing ported tools before this lane**).
- The item key convention `add_tool_novel_<domain>` is **per-site, serialising**: the dedup
  index allows one open item at a time. Deliberate throttle; respect it until the pilot
  proves parallel filing safe.
- Deploy onto a page that already exists attaches the native slot ALONGSIDE the ported one
  (measured: `tool-ab-test-calculator`, native slot position 2, ported slot left `pending`,
  raw `{{.section_heading}}` served). **So every replacement needs an explicit retire step:**
  set the page's `ported-page` slot `build_status='removed'` (the documented tombstone —
  `rerender_single_page_action.go:843`, `v3_site_actions.go:4591` exclude it) and re-render.
- Generator output quality: the newest build (`tool-css-unit-converter`, 2026-08-15 00:07)
  stored clean HTML, no raw tags. The ab-test raw-tag defect did not recur there.

## Phasing and decisions

1. **Pilot (filed 2026-08-15): `tool-aspect-ratio`** — simple, no fleet-wide function
   collision (checked), same URL. Item `add_tool_novel_webdesign.co.uk`, `triaged`.
   After the generator completes: retire the ported slot, re-render, verify at the served
   artefact (RUNBOOK). Only then batch.
2. **Batch, simple tools first.** Converters/calculators/generators with self-contained
   logic. Serial by default (respect the item-key throttle); revisit parallelism only after
   ≥3 clean serial replacements.
3. **Rich apps are NOT rebuild candidates — decision recorded.** Mind Map Studio Pro,
   Meme Studio, Logic Architect Pro, Flat-File Micro CMS, Pasteboard Manager and similar are
   hand-built applications; a generator rebuild is a REIMPLEMENTATION from a one-paragraph
   description and would silently downgrade them. They wait for the decomposition route
   (byte-faithful) or per-tool owner sign-off. Classify each tool before filing it.
4. **The owner sees each replaced tool** the same way he gated the ink change: the served
   page, not the item status.

## Constraints this lane inherits

- 13 ported tools keep logic in external `<script src>` files (TL-032) — for those, the
  ported page is not self-describing; the generator spec must be written from the LIVE tool's
  behaviour, and the external asset retired with the slot.
- The shared `ported-page` wrapper was poisoned 08-14 by a tool-improver shared write and
  **restored 2026-08-15** (v3 snapshot; 114 placements un-flipped; see `bugs_open/281`
  addendum + contribution). Do not touch the shared component from this lane; Track 1's
  `allow_shared_component_write` guard lands with the next roll.
- `bugs_open/204` (positional slots never resolve on rebuild) does NOT bite this route —
  replacement pages end with a single named tool slot, not positional `prose-N` slots. It
  bites the decomposition route; it is listed here so nobody re-derives the distinction.

## Verification per tool (the artefact, never the status)

- Served page at the same URL shows exactly ONE tool, interactive, no `{{.` tags.
- The page's `page_components`: one `component_level='tool'` slot deployed; ported slot
  `removed`.
- The fixed audit (Track 1, once live) lists the tool under clause (a) — the census in
  `bugfix_281_tool_audit_ported/RUNBOOK` counts it.
