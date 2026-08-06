# PLAN 2026-08-06 — bugs_open/189: the stored slot name travels with the section it names

Second arc of this session (first: 204, fixed+approved, open until live). 189
gates 204's closure canary and the loancalculator lane. Ownership verified
2026-08-06: the filing session (8134dee6) is not fixing it (3 tail mentions,
historic); `who-owns 189` collides with the OTHER 189 (split_symbol lane) —
resolved by slug; `save_page_sections_action.go` git log shows no 189 fix; the
queue's `lock_blocked_change` items are the guard WORKING (name-matched case),
not fix work. Defect re-verified: armed census still 12+2; the tool-loan-vs-
savings remediation held (4 sections served).

## The defect, one sentence

Save persists a section under a name derived from the COMPONENT
(`component_function` first), so the first time a positionally-named slot
resolves, its stored identity is overwritten (rename) and the locked-row guard
— which matches by that same name — silently fails to fire (duplication).

## Decisions and their reasons

1. **A new metadata field, `stored_slot_name`, carries the page_components row's
   own slot identity from every producer that KNOWS it; save prefers it
   VERBATIM.** This is the bug file's candidate 1 + candidate 2 as one move:
   the slot name (which section this is) and the component function (what
   renders it) stop sharing a field, which the bug names as the underlying
   defect. Verbatim, never normalised: it is a row identity being matched back
   to the row that issued it — normalisation could un-match a legacy spelling
   (`NormalizeComponentFunction` stays on the DERIVED-name path only, where 041
   put it).
2. **Absence of the field = today's behaviour, byte for byte.** Three save
   consumers measured live (page-rerender, page-build-handler,
   tool-recreation-handler); the tool-recreation path regenerates single-tool
   HTML with no structured slot identity and must keep working unchanged.
   Backward compatible with in-flight orchestrations expanded pre-roll.
3. **Producers:** (a) rerender's success entry and `carryStoredSection` add
   `stored_slot_name: s.slotName` — they hold the stored row in hand; (b) the
   build path's `RenderComponentAction` gains an optional `slot_name_from`
   config path (resolved via the same per-iteration mechanism as
   `component_from` — `setLoopVariable` puts `current_section` into
   CollectedData), outputs `stored_slot_name` when it resolves;
   `extractSectionFromMap` forwards it alongside the other meta keys.
   The plan item's `Name` is the `pages.sections` entry — for a decomposed
   page that IS the stored slot name; for a function-named page it equals
   today's derived name, so preferring it is a no-op there.
4. **`matchLockedRow` is deliberately untouched.** It already matches the
   incoming ComponentName exact-then-normalised with a consumed flag; once
   ComponentName is the stored slot name, the guard works. Matching by
   component_id instead was considered and rejected: duplicate slots
   legitimately share a component_id (11 pages), which would make the match
   ambiguous where names are unambiguous.
5. **Config change ships as seed + live UPDATE, applied after the image.**
   `page-content-writer`'s `render_section` / `render_from_template` steps gain
   `slot_name_from: "current_section.name"`. Old binary ignores the unknown
   key; new binary treats absence as today — both orders are safe, but the
   estate's convention (image first, then config) is followed. Seed file
   `023_page_content_writer_agent.sql` updated in the same commit.
6. **Save does not rewrite `pages.sections`** (verified: no UPDATE pages in the
   action), so stopping the slot_name drift at the insert is the whole fix —
   the 189 incident's pages.sections/page_components divergence cannot recur.

## Verification

- Tests (mutation-tested): locked positional row matched via stored_slot_name
  (no duplicate; guard consumes it); positional names survive a build save;
  field absent → identical to today (function-first + normalise); verbatim
  round-trip (no normalisation of the stored name); render action emits the
  field only when `slot_name_from` resolves; extractSectionFromMap forwards it.
- Post-roll (189's own §how-to-verify): fire `section_data_resolved` on
  tool-loan-vs-savings; assert exactly 4 rows, `tool-2` still locked at
  position 3, id/locked_at/locked_by unchanged; slot names `prose-*` intact.
  Then the 204 canary runs un-gated.
