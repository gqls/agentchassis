# PLAN — bugs_open/281: tool audit by INSTANCE (2026-08-15)

Owner brief: take the next unowned bug (281), research docs + DB, prefer a robust
framework-level fix over the individual case, council-review it, keep the docs updated,
commit what should ride the next chassis build; "please consider decomposing all the tools
so we can manage them in the framework properly".

## What is wrong (verified live 2026-08-15)

1. `check_tool_health.go:68` scoped `cc.component_level='tool'` → 4 of webdesign.co.uk's 67
   tools. The other 63 are page instances of ONE shared section-level `ported-page` component
   (`a7daa5c5…`, 115 instances / 2 sites); their tool code is per-page `rendered_html`.
2. tool-auditor's `load_tool` (live row = seed 088) resolves by component only, `LIMIT 1` →
   an arbitrary instance for shared components; its prompt reads `html_template` (the shared
   wrapper for a ported instance).
3. **Found during research, not in the bug file:** `check_tool_acceptance.go` (Tier 2)
   already admits ported instances (TL-033) and files `improve_tool` → tool-improver with the
   SHARED component_id. tool-improver's writeback (`update_component_html`) rewrites the
   shared template and flips all placements to pending — it did so on 2026-08-05 and
   2026-08-14 (`component_versions` v1/v3). Live latent hazard: the wrapper's template is
   currently tool-improver output; all 115 instances `pending`; ~~not yet propagated to any~~ [CORRECTION 2026-08-16: propagated to ONE page via the improver's delivery step, served ~23.5 h; found + restored by the 285 lane, which also restored the template + flags on 2026-08-15] not (otherwise) propagated to any
   rendered_html (verified by timestamps + content).

## Decisions and their reasons

- **D1 — routing.** Ported findings from BOTH producers file a new handler-less item type
  `ported_tool_fix` (`needs_human_review`), key `ported_tool_fix:<check>:<subjectKey>:<site>`.
  Reason: the only fixer writes the shared template (two incidents); ~~0 tool PLANs exist (89~~ [CORRECTION 2026-08-16: wrong table — `doc_plans` holds 143 tool PLANs, 14 for the ported 63; the decision stands on "no per-instance writeback exists", not a PLAN count] (89
  `needs_criteria`); orphan_element_refs is the written precedent for "no fixer until a PLAN
  and a per-instance path exist"; `section_edit` is not a proven per-instance fixer (no
  instruction payload, 1 completed ported-slot edit ever). Forks unchanged byte-for-byte.
- **D2 — Tier 2 LLM audit ON for ported instances.** The wait-for-PLANs argument governed the
  fixer, not the reviewer; only LLM review catches junk content. Auditor OUTPUT is gated in
  seed 425: ported findings → review item, never improve_tool.
- **D3 — load_tool pins the instance** (`AND pc.page_id=$2`, `input_data.spec.page_id`).
  Strict, not `COALESCE`-guarded: a nil param hard-fails (loud), and both seeds pre-flight
  that no open item lacks spec.page_id (0 live). Prompt reads `source_html`.
- **D4 — write fence** in `update_component_html`: refuse `component_level<>'tool'` AND
  >1 page unless `allow_shared_component_write` (opt-in default OFF, owner ruling 2026-08-02
  §2). Threshold from the census: only 2 tool-level forks are multi-placed (2 pages/2 sites,
  established pattern → WARN only); both incidents were section-level.
- **D5/D6 — keys and cooldowns** on the subject key / per instance for ported; forks
  byte-identical (subject key == function for level='tool').
- **D7 — volume.** Tier-1 uncapped (that visibility is the request); Tier-2 audit_tool
  capped 12/run with `ORDER BY p.name`.
- **TL-033 reversal stated explicitly** in code header, register (TL-042 + note in TL-033),
  and the council submission.
- **Track 2 (decompose all 63 ported tools): proposal only, owner decision.** Reasons in
  `PROPOSAL_2026-08-15_decompose_webdesign_tools.md`: bug 204 (open, unowned) makes
  decomposed pages unrebuildable; B2 omits `component_level='tool'` and verifiably REDUCED
  audit eligibility (1/18 LMC tool pages eligible); PLAN re-keying; external-script tools.

## Phasing

1. Code + tests + seeds + register + docs (this session) → council submit → pathspec commit.
2. Apply seeds 425/426 (dry-run + round-trip proven first) — config live immediately.
3. Go rides the next chassis roll (not built/rolled by this session unless asked).
4. Post-roll verification per RUNBOOK; append census to NOTES; close 281 only when
   fixed AND live.
