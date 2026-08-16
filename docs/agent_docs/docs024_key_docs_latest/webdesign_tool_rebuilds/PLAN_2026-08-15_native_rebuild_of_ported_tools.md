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

## OWNER RULING 2026-08-16 — §3 is REVERSED: the rich apps ARE rebuild candidates

> **CORRECTION to §3 above (do not read §3 as current).** §3 recorded "Rich apps are NOT rebuild
> candidates" — Mind Map Studio Pro, Meme Studio, Logic Architect Pro, Flat-File Micro CMS,
> Pasteboard Manager and similar — deferring them to the byte-faithful route or per-tool sign-off.
> **The owner has ruled option (a): generator rebuild anyway, accepting that it will be a
> REIMPLEMENTATION rather than a preservation.** The trade §3 warned about (an AI rebuild from a
> short description silently downgrades a hand-built app) was put to him in those terms and
> accepted. So there is no excluded class left: all 63 are in scope for this lane, and the
> decomposition/byte-faithful route is no longer a prerequisite for any of them.
> Recorded by the `bugfix_285_shared_template_write` lane, which asked the question at closure;
> this lane owns the execution.

What the ruling does NOT change — the existing conditions in this PLAN still bind, and they carry
more weight for a rich app than for a converter:

1. **The spec is written from the LIVE tool's behaviour, not from its page.** Already required for
   the 13 external-script tools (Constraints, above); a rich app is the same shape for a different
   reason — its value is in interactions no static read of the page describes. Reimplementation is
   the accepted outcome; an ACCIDENTAL reimplementation of half the features is not, and the only
   thing separating them is how the description is written.
2. **Grade the generated component BEFORE retiring the ported slot** (the ab-test lesson, measured:
   a 13 KB slot with zero visible text and 47 raw `{{.` tags reached the served page). For a rich
   app the grade is a feature list checked in a browser, not a raw-tag count.
3. **Retire, never delete.** `build_status='removed'` is the documented tombstone; the ported bytes
   stay recoverable from `page_component_history` (357 archive) and `cmd/webdesignport` is
   idempotent. Note the archive row id per tool in NOTES as you go — that is what makes a bad
   reimplementation reversible in one statement rather than a re-import.
4. **RECOMMENDED, lane's call: rich apps go LAST and one at a time**, after the simple batch has
   proved the recipe end to end, with the owner seeing each served page (§4). The ruling settles
   WHETHER, not the order; front-loading the hardest reimplementations before the recipe is proven
   spends the owner's review attention on the least certain output first.

### CORRECTION 2026-08-16 16:20Z — condition 3 of the ruling above asks for a row that cannot exist

> Condition 3 says: "Note the archive row id per tool in NOTES as you go — that is what makes a bad
> reimplementation reversible in one statement rather than a re-import." **The retire we prescribe
> creates no archive row.** The trigger is `trg_page_component_artefact_archive_upd AFTER UPDATE **OF
> rendered_html** ON page_components … WHEN (old.rendered_html IS NOT NULL AND new.rendered_html IS
> DISTINCT FROM old.rendered_html)`, with an AFTER DELETE twin. Setting `build_status='removed'`
> changes neither `rendered_html` nor deletes the row, so neither fires. Measured on the pilot page
> `00979b9e`: `page_component_history` = **0 rows**. (This is why the 08-16 NOTES entry recorded
> "No archive row for the pair" on ab-test — true, and its mechanism was not known at the time.)
>
> **The substance of condition 3 is unaffected and is in fact stronger than it was written.** The
> reversibility comes from the same choice — retire, never delete — but the handle is the surviving
> `page_components` row, not an archive row: it keeps its `rendered_html` verbatim, so the revert is
> one status flip plus a re-render (proven on ab-test, 2026-08-16 10:0xZ). **So record, per tool:
> the ported slot's row id, its `rendered_html` length and md5, and its pre-state `build_status`** —
> the md5 is what lets a later session prove the bytes it is restoring are the bytes that were served.
> `page_component_history` remains the right place to look for content-CHANGING writes (a section
> edit, a rerender that rewrites a slot); it is simply not part of the retire path.
