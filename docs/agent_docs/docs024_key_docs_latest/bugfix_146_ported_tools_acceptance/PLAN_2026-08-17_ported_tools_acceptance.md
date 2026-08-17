# PLAN — bugfix 146: ported tool pages and the acceptance tiers (2026-08-17)

Bug: `bugs_open/146_HANDOFF_2026-07-29_ported_tool_pages_are_outside_every_acceptance_tier.md`
(the `ported_tool_pages` slug; 146 is an ambiguous number). Lane NOTES hold the evidence trail.

## What the bug turned out to be (re-validated 2026-08-17)

The filed mechanism (`tool_acceptance_due` only sees `component_level='tool'`) was fixed
**three hours after filing** (TL-033, `ac9f75a0c`, 2026-07-29 17:19): the shared
`toolEligibilityWhere` admits ported sole-component `page_type='tool'` pages. Nobody
updated the bug. What actually keeps the symptom alive — all 7 filed pages still overflow
at 390×844, re-measured live today (0 clean / 7 flagged, lane dir `scan_146_rerun…txt`) —
is two doors:

1. **No fence → no run.** Tier 4 is criteria-gated; 6 of the 7 pages have no current
   ```criteria``` PLAN, so the only instrument that can measure their defect
   (`no_horizontal_overflow`, Tier-4-only) never runs. 48 of 67 ported pages fleet-wide
   are unfenced (webdesign 45, loancash 3).
2. **Fenced + FAILED → silent sink.** `JudgeAcceptanceResultsAction` resolves its
   component by `cc.function = <subject key>` (tool_acceptance_actions.go:866-873); a
   ported instance's component is `ported-page`, so `componentID=""` and the else-arm
   (:1029-1040) writes the note ("route this manually") and files NOTHING. Live: 4 failing
   ported verdicts (pasteboard ×2, vibe-equalizer ×2 — the exact overflow this bug
   measured), 0 work items. This is `bugs_closed/281`'s Finding B, left OPEN and unowned.

## Decision: fix door 2 in code now; put door 1 to the owner as a costed choice

**Door 2 (this lane, now).** Make the Tier-4 judge the THIRD producer of
`ported_tool_fix` — the vocabulary TL-042 established for exactly this population,
already produced by `check_tool_health` (Tier 1) and `check_tool_acceptance` (Tier 2),
both handler-less at `needs_human_review` by design (tool-improver would rewrite the
SHARED wrapper; clobbered fleet-wide 2026-08-05 and 08-14). Owner ruling 2026-08-02 §1:
converging producers onto one item_type needs no RFC provided the register names the
producer set and the item_key shape — TL-042 is updated in the same commit.

Design (opt-in-by-evidence; default path byte-unchanged):
- In the `componentID==""` else-arm, read the run item's own spec
  (`input_data.spec.component_id`, present on every `check_tool_acceptance_due`-filed
  item; verified live). If absent → today's behaviour exactly.
- Resolve that id: `SELECT component_level FROM content_components WHERE id AND is_active`.
  **Only `component_level <> 'tool'` routes** — a fork whose function moved, a deleted
  component, or any lookup error keeps today's behaviour (fail-safe, nothing new fires).
- File `ported_tool_fix`, status `needs_human_review`, no handler, severity medium,
  priority 60, source 'acceptance', created_by 'tool-acceptance-agent',
  `item_key = ported_tool_fix:tool_acceptance_tier4:<subjectKey>:<siteID>` (sibling keys
  use their own check segment: `tool_health`, `tool_acceptance`). Spec carries
  check=tool_acceptance_tier4, subject_key, component_id (the wrapper), page_id,
  page_name, issue, failing_checks, failing_instances, acceptance_test (the criteria),
  screenshots, overflow_forced_by / overflow_fix_hint — everything a human needs to route.
- Insert uses the file's reviewed idiom (acceptance_stuck's): `ON CONFLICT (site_id,
  item_key) WHERE item_key IS NOT NULL AND status NOT IN (<terminal>) DO UPDATE` with a
  spec MERGE — a re-verdict refreshes the standing decision, never duplicates it, never
  clobbers a human's keys.
- The acceptance-fail note's `Fix:` line records the filing (the note is written from the
  outcome, per the file's own rule); result map gains `ported_tool_fix_filed` only when it
  fired (key-presence convention).

**Blast radius, measured not assumed:** exactly ONE live caller of
`judge_acceptance_results` fleet-wide (`tool-acceptance-agent`; active agent_definitions
census + all retained orchestration_states; no sub_workflow references). The component
browser-run path uses a different judge and is untouched.

**Door 1 (owner decision — do NOT take unilaterally).** Should ported tools get a criteria
fence automatically? Options:
  a. `adopt_verbatim` writes a baseline Tier-4-only fence (mobile-fit
     `no_horizontal_overflow` + `no_console_errors`) for future ported tool pages, plus a
     one-off backfill for the 48 unfenced. Cost: ~48 more subjects on the 7-day Tier-4
     cadence (spawned pod + browser + vision per run). ⚠ L8171: a ported fence must
     contain NOTHING Tier 2 can evaluate (no `page_status_ok`) — Tier 2 files
     `improve_tool` unconditionally at the shared wrapper; the 281 write-fence blocks the
     write but the item still mis-aims. ⚠ L8744: writing a PLAN also switches on Tier 2's
     three out-of-fence checks.
  b. Backfill only non-webdesign (loancash 3); webdesign's retire by attrition as the
     owner-directed rebuild lane replaces them (4 of 63 done as of today, one at a time,
     each rebuild writing a proper composer PLAN).
  c. Status quo: fence-writing stays a per-lane authored act; static tool_health is the
     only automatic coverage (it caught 2 of the 7).
Recommendation: (b) now, (a) for the PORTING PATH going forward (new ported tools arrive
fenced), revisit backfill scale after the rebuild lane's cadence is known.

## The seven pages themselves

Not fixed page-side here (the bug's own candidate 3 is marked NOT RECOMMENDED alone; the
rebuild lane owns their replacement). A dated pointer goes into the rebuild lane's NOTES
naming the seven as re-measured-broken today, for their queue ordering.

## Consumers told (shared-seam ruling 2026-07-29 §3)

- `bugfix_281_tool_audit_ported` (TL-042 owner) — register updated in-commit; their
  Finding B is what this closes.
- 291 lane (human-review drain) — 19→more `ported_tool_fix` rows possible; their phase 2
  demotion door does not touch born-`needs_human_review` items.
- `webdesign_tool_rebuilds` — pointer re the seven pages; population shrinks as they build.
- `mortgagecalculator_couk_adoption` — fence guidance unchanged; their A4 route benefits
  from door 2 for their 1 remaining ported instance.
- 285 lane — the per-instance fixer gap (TL-042 gap b) is unchanged by this; when a fixer
  exists, this producer's items are its natural feed.

## Verification

- Unit: new `tool_acceptance_ported_sink_test.go` — firing arm (ported spec → item with
  key/status/spec asserted from the SQL), negative controls (no spec.component_id →
  byte-identical behaviour; component resolves to a FORK → no item), and the existing
  suite must stay green untouched (it never supplies spec.component_id, so the default
  path is proven unchanged by construction).
- Mutation check: invert the `component_level <> 'tool'` guard locally and watch the fork
  control fail (a guard is proven by mutating the code, not by the mock's bookkeeping).
- Live (after the next roll — Go is inert until then): the standing failing case IS the
  induction — vibe-equalizer's next 7-day acceptance run on its real fence should file
  `ported_tool_fix:tool_acceptance_tier4:vibe-equalizer:6b49db8e…` with the overflow
  attribution, or the page gets rebuilt first and the case moves to pasteboard (fence
  present, clip overflow, not in the rebuild lane's first batches).

## Council

Submit via 097 before/alongside the commit; trailer per verdict state. Edits ≤8, this
file is the rationale's source.
