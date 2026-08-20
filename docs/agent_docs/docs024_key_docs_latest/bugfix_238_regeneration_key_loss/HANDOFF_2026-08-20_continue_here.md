# HANDOFF 2026-08-20 — `bugs_open/238`, continue here

**Lane:** `docs/agent_docs/docs024_key_docs_latest/bugfix_238_regeneration_key_loss/`
**Bug:** `bugs_open/238_HANDOFF_2026-08-09_content_regeneration_drops_template_required_image_url_fields_leaving_empty_src_live.md`
— read **§11** first (this session's work); §1–§10 are the earlier history and §9.3/§10 contain
corrections to themselves.

**Read before acting:** `SUMMARY_2026-08-20_the_detector_was_off_and_the_repair_was_not_needed.md`
(the read-out), `RUNBOOK_regeneration_key_loss.md` §"2026-08-20 additions" (every query below, with
its gotcha attached), `NOTES_regeneration_key_loss.md` (the full log incl. both missteps).

---

## 1. State in one paragraph

The **mechanism** is fixed and live and now *proven* on fleet traffic. The **detection** half was
armed nowhere for nine days (a stale register line is why); its record-only half is now **armed,
council-APPROVED, and PROVEN AT THE ARTEFACT** — the first `dead_url_control` item in the
platform's history was filed 2026-08-20 17:21 by an owner-authorised demand control. The
**remediation** everyone plans for this bug has **zero population** — measured — because the class
is "the declared source never existed on this site", not "a value was lost". The bug stays **OPEN**,
and the remaining work is a decision and some content, not code.

## 2. What shipped this session (all committed, all on `087_towards_multiple_domains`)

| what | where | state |
|---|---|---|
| `record_dead_url_controls` declared on the action that reads it | `bb6600e48`, `rerender_page_sections_action.go` | **LIVE on v1.0.1319** (rev `447f3a8a8`, controls both ways) |
| test + mutation for that declaration | `dead_url_guard_test.go` `TestDeadURLGuardConfigKeysAreDeclared` | green; M1 (undeclare) fails it with the exact report text |
| **arming migration** | `sql_for_agents/504_bugfix_238_arm_dead_url_record_on_rerender.sql` (+ `_ROLLBACK`) | **APPLIED**, 1/1 steps fleet-wide, recorded in `schema_migrations` |
| council submission | `COUNCIL_SUBMISSION_2026-08-20_arm_dead_url_record.json` | **APPROVED r2**, corr `8a2aab7c-2ffa-469d-bb55-ce5a11126613` |
| **RFC_042** (the merge-vs-replace split) | `architecture_review/RFC_042_content_data_has_nine_writers_*.md` | DRAFT, **awaiting owner decision** |
| decision record | `doc_notes` `subject_key='bugs_open/238:dead_url_control_arming_sequence'` | written |
| register corrections | PBP-040 status (was stale 9 days), PBP-039 verify-later **discharged** | committed |
| LANDMINES + 016b §9 + WRONG_CALLS ×2 | strengthened existing entries, not duplicated | committed |

## 3. ⚠ The three things that will mislead you

1. **A migration/RFC number read more than a few minutes ago is stale.** 497→504 and RFC_041→042
   both collided mid-write. `ls | tail -1` is not enough; **allocate immediately before naming the
   file**, and note 497/498 each still carry two lanes' files.
2. **The `page_component_history` probe is wrong in BOTH directions** (LANDMINES, strengthened
   entry; 016b §9). Loose (`page_id` only) over-counts by slot; strict (`slot_name` +
   `source='artefact_archive_trigger'`) under-counts by writer and returns a confident **0** —
   app-written rows carry NULL `slot_name` and they are the ones holding the values. Discriminator
   that works: **content identity** (how many deployed components on that page declare the field).
3. **A zero from this emit means "nothing re-rendered", not "no damage"** — it fired the first time
   a re-render actually ran, so the nine days of zero were traffic, not blindness. Establish which
   with the archive count (`page_component_history` rows in the window), never by waiting. And
   note it can only ever report UNGATED fields, so its silence says nothing about the gated class.

## 4. Next actions, in order

### 4.1 ✅ DONE — the emit is proven at the artefact
Owner-authorised demand control ran 2026-08-20 17:21. Hand-patch check clean first (all 8 sections
machine-made). **First `dead_url_control` item in platform history:**
`dead_url_control:index:case-studies-grid:card1..5_image_url`, `needs_human_review`, no handler,
`refused=false`. Page unharmed: empty `src` 5→5, anchors 0→0 as predicted; the only delta was
another lane's SEO-005 Open Graph improvement arriving for free. Full account: bug file §11.10.

⚠ **The item names only the five UNGATED fields.** That is a live confirmation of the blind spot —
and it refuted a claim the council round rested on. `href=""` is already covered by
`empty_internal_href`; the gated/vanished class is covered by **nothing**, including this. So the
detector picture is now **three producers on the ungated symptom, none on the gated one**. A fourth
producer here should consolidate, not add (`WRONG_CALLS` 2026-08-20; `RFC_030`'s instinct).

### 4.2 ⏸ OWNER DECISION — *(was the demand control; done, folded into 4.1)*

### 4.3 The real next build: make the 10 damaged (page, slot) pairs visible
They exist only as `agent_error_log` findings today. Two halves that **must land together**:
- a **backfill** minting `required_fields_missing` items (producer's key shape
  `required_fields_missing:<page_id>:<slot_name>`) for the pairs in §11.4 of the bug file;
- a **park route for resolver-sourced fields** in the existing router
  (`sql_for_agents/410_required_fields_missing_router.sql`) — **widen it, do NOT clone it**:
  `RFC_030` is RULED (owner 2026-08-15) and a fourth bespoke router is forbidden.

⚠ **Without the park route the backfill is harmful**: those items fall to the `partial` route →
`content_rewrite` at the prose writer, which structurally cannot emit resolver-sourced keys. That
misroute is live in the router **today** (it has no population only because the producer never
files these items) — proven by the `410` `asset_sourced` canary refusal.
⚠ **Do NOT build the `source_resolves` / `history_restorable` routes** the earlier plan proposed.
Measured: **0 population each**. §11.4 has the numbers.

### 4.4 ⏸ Then, and only then: arm the refusal (`380_..._HOLD.sql`)
Gated on four measurable drain conditions — see the `doc_notes` decision row (authoritative) or
§11.2. Apply **by hand**; the runner must never take a `_HOLD`. Re-run its binary pre-checks at
apply time. `504`'s negative control asserts the refusal is armed nowhere; expect to remove nothing.

### 4.5 ⏸ OWNER DECISION — `RFC_042`
Four costed options for the merge-vs-replace split. Lane recommendation: **(e) now, (c) next, (b)
only if (c) shows real non-funnel losses** — build the detector before the guard, because the
guard's population is currently an inference. Flagged to be answered **jointly with `RFC_008`**
(same question, sibling column of the same table).

## 5. What is deliberately NOT being done, with the trigger that would change it

- **The two carry gaps** (a blank resolved value beats a good stored one; a `query.*` resolver ERROR
  drops the key) — real in the code, **0 observed instances**. Recorded in RFC_042 §4.6, not shipped.
  **Ship them when:** the generation-pairing query returns one loss event with a
  `site_specs.*`/`site_assets.*`/`query.*` source, or a `STRUCTURAL_KEY_CARRY_MISS` row whose page
  and slot DID hold the value in the prior generation.
- **The 8 unguarded `RenderTemplate` call sites** (`dead_url_guard.go`'s own header names them) —
  council-ruled RFC-shaped; overlaps `RFC_041`'s error contract. Not a rider on a bug fix.
- **Widening `check_required_fields_missing`** to `site_specs.*`/`pages.*` — census-gated, and the
  discovery rotations are paused on cost (`bugs_open/230`), so a widened check has no driver.

## 6. Closing the bug

Bar is fixed **AND** live. Prevention: done and proven. Detection (ungated): armed and **proven at the artefact** (§4.1).
Detection (gated): **still nothing** — the honest residual, precisely bounded in §11.6. Remediation:
10 pairs need §4.3 then human decisions. When it moves, **name both paths on the commit**
(`git commit bugs_open/OLD.md bugs_closed/NEW.md`) and verify at HEAD with `git ls-tree`, not `ls`.
