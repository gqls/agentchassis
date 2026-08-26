# HANDOFF — 333 owned-page door, 2026-08-25

> **CLOSED 2026-08-26 — the bug is in `bugs_closed/333_…` with the full close-out; the milestone read-out is
> `SUMMARY_2026-08-26_owned_page_door.md`. Nothing below is a work instruction any more; it stands as the
> record of how the lane ran.**

**COLD-START: read this file, then `bugs_open/333_HANDOFF_2026-08-19_producers_route_content_findings_at_page_build_handler_without_reading_rebuild_policy_so_owned_pages_queue_findings_that_can_only_be_refused.md` (its POST-ROLL section), then `RUNBOOK_owned_page_door.md` for the queries.**

Lane dir: `docs/agent_docs/docs024_key_docs_latest/bugfix_333_owned_page_door/`

---

## 1. Where this stands in one paragraph

The bug: about two dozen producers file content findings at `page-build-handler` without checking whether the
page belongs to a tool. On such a page that handler is forbidden to act, so it refuses, and the finding is
recorded as `wont_fix` — "we decided not to fix this" — on a real defect. The fix (candidate 1 of the bug's own
list) puts one check at the shared write seam: if the page is owned **and** the target handler has *declared*
it refuses owned pages, the finding is parked visibly instead of being routed into a guaranteed refusal.
**That fix is built, council-APPROVED, live, and proven on live demand.** The bug is still OPEN, deliberately,
because three findings were refused *after* the fix went live via a producer that bypasses the seam entirely.

> **CORRECTED 2026-08-25:** three → **sixteen**, in three populations, only 3 of which bypass the seam —
> see the corrections at §4 and the owner rulings at §5.

## 2. State — what is done

| item | state |
|---|---|
| Fix candidate 1 (the door) | **LIVE + PROVEN**, chassis provenance `4c996e1b5`, both replicas, verified 2026-08-25 09:39Z |
| Council | **APPROVED round 2**, corr `9813dec8-5ce1-48ab-bb77-e3f601f9f64c` (round 1 REVISE; two objections changed the code) |
| Register | **WII-028** (`work-item-integrity.md` + index row), status LIVE AND PROVEN |
| Landmines | 3 entries (re-typing a parked row; a fail-open guard voiding seam tests; **a marker has a birth date**) |
| 016b §9 | 1 pattern entry (N producers route by type; ask the HANDLER, not a name list) |
| WRONG_CALLS | 3 entries, one with a 3-step correction chain |
| Standing five | PLAN, RUNBOOK, NOTES, README all present and current; no SUMMARY yet (see §7) |

**Code commits:** `6ab0b3434` (door) → `1789489bf` (round-1 revision). Both are ancestors of the deployed stamp.

**The door, in one sentence of mechanism:** in `writeWorkItem`, before the anti-churn brake, if the item carries
a `page_id` at a dispatchable status → read `pages.rebuild_policy` (a PK read) → if owned, probe whether the
handler declares `refuse_owned_page` (mig 488's key) → if so, rewrite the item via the pure
`ownedPageParkedItem`: `status='deferred'`, `handler_agent=''`, priority 200, `recurrenceExpected`, **item_type
and item_key UNCHANGED**, spec gains `gap_kind`/`builder_needed`/`what_to_do`/`owned_page_guard`, error leads
with `OWNED_PAGE_GUARD`. Kill switch `DISABLE_OWNED_PAGE_DOOR_DEMOTION`, ships armed.

## 3. Proof it works [MEASURED 2026-08-25 09:39Z] — do not re-derive, re-run

Door live since **2026-08-24 19:19:13Z**. On `rebuild_policy='owned'` pages since then:

- **32 findings PARKED** by the door (`required-fields-missing-handler` 28, `generic` discovery 4)
- **3 still REFUSED** (`offer-analysis` — bypasses the seam)
- **243 untouched and completing** (`rerender-pages`, `generic`)
- **35 parked rows** total, 5 item types, 4 sites
- **Per finding, not per page — PROVEN**: page `c67ed17b` holds 2 parked findings under 2 distinct `item_key`s
- **Consumer reads them**: roadmap sweep groups all 35 under one `builder_needed` line
- **Negative controls held**: `page-rerender` on owned pages 244 complete / 10 failed; `page-build-handler` on
  generic pages still runs

## 4. WHY IT IS STILL OPEN — the one thing that matters

Three `offer-analysis` rows created **2026-08-24 22:08:39Z, three hours after the door went live**, all died
`wont_fix` on the ownership guard. They never met the door: `write_audit_findings_action.go:987` is a raw
`INSERT INTO site_work_items`. **The bug's title is still literally true of that producer.**

> **CORRECTED 2026-08-25 (adversarial review, Fable; each point re-verified in-lane):** "three findings" was
> a `page_id`-JOIN census and undercounts. Since the door went live, **16** owned-page refusals are 333's
> defect: the 3 `offer-analysis` rows above PLUS **13** `needs_page` rows from `page-rerender`'s
> `content_data_backfill` escalation (`escalateRerenderToWriter`,
> `platform/orchestration/actions/rerender_page_sections_action.go:1397-1435`) — which are **NOT a
> bypass**: they pass through `insertWorkItem`/`writeWorkItem` and the door stands down because the
> producer omits `pageID` (resolved at `:196-215`, dropped from the literal at `:1415`), and they are born
> `status='triaged'` (`:1425`), so **no promoter-side remedy can ever see them**. A further 8 rows
> (`misdirected_cta` findings at `page-rerender`) are the CTA lane's (`bugs_open/389`) — not 384's as this
> lane first said, and not 333's. Rulings in §5.

## 5. What to do next, in priority order

1. **Close the bypass gap — the actual remaining work.** Two shapes:
   - **(recommended) Promoter-side predicate.** Add the ownership test to `workItemRoutableSQL` /
     `TriageDetectedItemsAction` so a `detected` row at a refusing handler on an owned page is held back. This
     covers `write_audit_findings` (which births `detected`) **without editing 9 call sites**, and it mirrors
     `bugs_closed/284`/WDS-017, which did exactly this for the registration predicate. Smaller and more general.
   - **(alternative) Route the raw writers through `writeWorkItem`.** 9 files, all enumerated in the WII-028
     register entry by path. More churn, touches other lanes' files.
   ⚠ **Whichever you pick, the promoter option changes a SHARED gate — it needs its own council round.**
2. **Decide the 111 legacy rows** (59 failed / 36 unresolved / 16 needs_human_review, **0 dispatchable**).
   Nothing is burning; they are a false record. Owner call: a `_HOLD` migration re-typing them into the parked
   shape, or a deliberate mass-cancel. Do not do this silently.
3. **Optional, probably NOT this bug**: 1,438 rows carry `spec.page_name` and no `page_id` — name-only ACTION
   REQUESTS, a different kind of item from this bug's content findings.

### RULED 2026-08-25 (owner) — the list above is superseded; three decisions landed

1. **Bypass gap: route `write_audit_findings_action.go:987` through `writeWorkItem`.** The promoter-side
   option is **dropped**: the promoter that actually promoted the 3 `offer-analysis` rows is the scheduled
   task `detected-item-promoter`'s raw-SQL `pre_query` (all three carry the identical
   `triaged_at = 2026-08-24 22:21:38.324` of one 900s tick) — NOT the Go `TriageDetectedItemsAction`, so a
   `workItemRoutableSQL` predicate would not have covered them, and a SQL-side park would re-implement the
   parked shape in a second medium (the "one contract in two media" drift `work_items_common.go` already
   warns about). The Go promoter is not dead — `improvement-loop` runs periodically (owner, 2026-08-25) —
   it just was not the promoter of record here. Council round owed with the change (shared seam file).
2. **The 13 `page-rerender` escalations: owner chose (i) — check `rebuild_policy` before escalating** in
   `escalateRerenderToWriter`, stopping the false alarm at source. NOT (ii) pass-the-pageID-and-park: the
   key `needs_page:<name>` is deliberately shared with `reconcile_site_plan`, and a parked (non-terminal)
   row would hold that dedup slot where today's terminal rows free it. Name-resolution at the door is NOT
   taken — the class's live producer holds the id and drops it; all-history for name-only refusals is
   **272 rows since 08-03 (live+archive)**, not the 53 this handoff first said (live-table-only read).
3. **Legacy rows: no relabelling.** Only **58 of the 111** carry an ownership-refusal error (58 of 59
   `failed`; **0** of 36 `unresolved`, **0** of 16 `needs_human_review` [MEASURED 2026-08-25]) — re-typing
   the 111 would overwrite 53 rows' real failure history; **31 of the 59 `failed`** would collide with a
   live open row of the same `item_key` under `idx_swi_dedup` if made non-terminal; and the
   `failed`/`wont_fix` halves (incl. the 46 `tool_crosslink` `wont_fix`) are archiver-eligible within days
   (`archive_completed_work_items(7,1000)`). If re-typing is ever revisited: dedupe by key first, exclude
   rows whose recorded error is not an ownership refusal.

**IMPLEMENTED 2026-08-25 (same day): rulings 1 and 2 are BUILT, COMMITTED and COUNCIL-APPROVED —
`0ad313f02` + round-2 revision `fb2bd056d`, corr `70a1e557` APPROVED round 2 (round 1 REVISE: the
guardian's HIGH drew the 30-day census — 22/579 would-be-branded, the brake's design class — and the
owner's explicit full-routing sign-off; bug_historian's MEDIUM added the durable per-page
`owned_page_review` record on every escalation skip, then flipped to approve). **LIVE AT THE ARTEFACT since the 2026-08-25 19:07Z roll** — both replicas probed ~19:4xZ: `skipped_owned_page` ×2 and `routed via the work-item seam` ×1 present in `/proc/1/exe`, door control present, must-be-absent control 0. **Behavioural verification is still OWED, because demand has been ZERO on both paths since the roll** (0 audit filings on owned pages; 0 escalation attempts — both log literals at 0 in both pods' logs), and the RUNBOOK's own rule stands: a zero is only a pass with a demand control. Re-run the RUNBOOK §"Verifying the residual fixes" queries once either denominator is non-zero.** Four guards mutation-proven (NOTES,
2026-08-25 later entry); post-roll verification queries with their demand controls are in the RUNBOOK
("Verifying the residual fixes"). Ruling 3 needs no code. WII-028's raw-writer census updated 9→8.

## 6. Traps that will bite you specifically on this lane

- **A marker has a BIRTH DATE.** `OWNED_PAGE_GUARD` was only added to `SavePageSectionsAction`'s refusal on
  **2026-08-19**. Any historical census keyed on it silently drops everything older and the dropped half reads
  as a different defect — this cost me a wrong correction sent to a peer. For history match
  `error LIKE '%rebuild_policy=owned%' OR '%OWNED_PAGE_GUARD%'`. For post-roll door rows the marker alone is
  correct (they all carry it by construction). Full entry in `LANDMINES.md`.
- **A census of `OWNED_PAGE_GUARD` now counts PARKED rows as refusals.** Add `AND status <> 'deferred'` when you
  mean refusals.
- **Adding any guard to `writeWorkItem` silently voids existing sqlmock tests.** It fails open, and sqlmock's
  "unexpected query" error is exactly what the fall-through swallows — 21 test cases across 8 files were in that
  state. Call `expectWorkItemDoorStandsDown(mock)` / `expectWorkItemDoorGenericPage(mock)` (in
  `work_item_owned_page_door_test.go`), placed after the `ExpectBegin()` of the transaction containing the
  `INSERT INTO site_work_items` — **not after every `ExpectBegin()`**. And if your fixture has no `pageID`, the
  door never fires and the helpers must be ABSENT (the 326 lane's dual of the same trap).
- **The door's unit is the AGENT, not the branch.** `page-rerender` is both the estate's principal owned-page
  route (5,216 completes) and capable of refusing one, depending on `spec.reason`. It must never declare
  `refuse_owned_page`. A producer targeting such a handler needs a consumer-side exclusion instead — this was
  ruled for the `bugs_open/384` lane and is in the register entry.
- **Verify at the artefact, per service.** `logs … | grep -m1 'build provenance'` scrolls; fall back to
  `kubectl exec <pod> -- grep -aq DISABLE_OWNED_PAGE_DOOR_DEMOTION /proc/1/exe` **with a must-be-absent control
  in the same breath**, on BOTH replicas.

## 7. Housekeeping owed

- **No SUMMARY has been written for this lane.** By the cadence rule that is correct so far (one arc, no
  inflection yet). **The right moment is when the bypass gap closes** — that is when "where we are now" genuinely
  changes.
- Two orphan register entries exist that are NOT mine (`CQ-031`, `PLAN-053` — entries with no index row);
  flagged in the index narrative for their owning lanes.
- The landmine verifier returns `NEEDS_HUMAN_REVIEW` on this lane's entries: **stale-index false negative**, the
  code index was last built 2026-08-24 08:40Z. Both entries carry a note saying so. **⚠ A "known false negative,
  ignore it" line is the most perishable kind (375 lane, 2026-08-25) — it makes a reader SKIP work, and skipping is
  invisible. Re-establish before obeying: `SELECT max(created_at) FROM code_index_symbols;` (or the index's own
  freshness row) against `git log -1 --format=%ci -- platform/orchestration/actions/work_item_owned_page_door_test.go`.
  If the index is NEWER than the file and still says NEEDS_HUMAN_REVIEW, that is a real finding, not this note.**

## 8. Consumers already told (do not re-notify)

`bugs_open/326` (colliding in the same function — landed on top, lane closed), `bugs_open/367` (their
`from_rfm` rows are the biggest parked population — 28 of the 32), `webdesign-tool-rebuilds`,
`staged_component_build`/`bugs_open/353`, `bugs_open/384` (the per-agent/per-branch ruling).
`vigilant_designer_offer_analysis` — **told 2026-08-25** via
`docs/agent_docs/docs024_key_docs_latest/vigilant_designer_offer_analysis/CONTRIB_2026-08-25_write_audit_findings_bypasses_the_new_owned_page_door.md`
(their session was not running, so it is a file in their own CONTRIB_ convention). It states the measured cost,
recommends the promoter-side remedy that would cover them **without touching their file**, and records that
routing their action through `writeWorkItem` instead is theirs to choose. **Nobody is owed a notification now.**

> **UPDATE 2026-08-25:** that sentence is superseded — three more notifications were owed and sent (commit
> `42a89c47f`): `bugs_open/389` + `bugfix_308_cta_destination_provenance` (their stated belief that
> owned-page `cta_links_stale` rows park under this door is FALSE — 0 ever have, the door keys on the
> target handler's declaration and `page-rerender` declares none; corrected in both docs plus a CONTRIB in
> their dir), and an ADDENDUM to the `vigilant_designer_offer_analysis` CONTRIB withdrawing its
> promoter-side recommendation (owner ruled seam-routing; see §5 rulings).
