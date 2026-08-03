# NOTES — bugfix 134 (optional-marker keys) — append-only, newest at the bottom

## 2026-08-03 — lane opened, bug claimed

**How the bug was picked** (recorded because the picking discipline is half the
work on this tree): ranked all 56 `bugs_open/` numbers by reference-heat over the
39 transcripts touched in the last 4h, ascending. Coldest genuine candidates:
146 (27), 116 (36), 121 (42), 163 (49), 134 (51). Disqualified on reading:

- **146 #1 (unreachable tool pages)** — its nav mechanisms are the active
  `bugfix_149_nav_membership` lane's plan, heat 432. 146 #2 (ported pages outside
  acceptance tiers) is genuinely unowned but candidate 1 is blocked on an
  owner/architecture call about whether ported pages are deliberately out of scope.
- **163 (landmine verifier symbol lookup)** — unowned on paper, but the fix file
  (`diagnose_code_lookup_action.go`) was being read line-by-line by the session
  working bug 181 (rowCap greps, 11:30) — a same-file passenger waiting to happen.
- **116 (link checks never run)** — valid, but sits in the checker-layer
  schedule/dispatch territory where the 149 queue and today's 185/186 filings are
  active; high collision surface.
- **098** — memory said unowned; `who-owns` showed a `seam(098)` commit TODAY.
  A memory entry's ownership claim ages in hours.
- **186** appeared in `bugs_open/` between two `ls` runs four minutes apart —
  being filed live by another session while I was picking.

**134 validity re-check (all four figures, live, ~11:50 BST):** live row still
carries `"category?"`/`"limit?"`; agent runs = 0 all-history; fleet punctuation
sweep = the same 2 rows and nothing else; action reads `category`/`limit`
unsuffixed at `refresh_product_specs_action.go:211,215`, spec has no
`CheckConfig`. The five open `site_work_items` matching "product-spec" are about
a robot-hands page *component* — unrelated. `needs_diagnosis` queue: empty.

**No fresh 090 run, stated per the 2026-07-31 ruling's escape hatch:** the filed
root cause is fully cited, re-verified first-hand today, and the mechanism is
self-evidencing (exact-key extraction; no Go reads a `?`-suffixed key — grep).

**Plan:** `PLAN_2026-08-03_optional_marker_keys.md`. Three parts: data (seed 156 +
migration 298), contract (`CheckConfig: true`), class (`--suspicious-keys` audit
mode fleet-wide, including non-opted-in actions).

**Council:** submitted 12:0x BST, `SUBMISSION_CORR =
9521d62f-f239-4590-875b-4c4f2e9c4343` (submission JSON in this dir). Budget ~30
min for the queue; find the run by payload, not the printed id.

**Implementation:** dispatched to an opus subagent (files: the action, seed 156,
new migration 298, `cmd/config-key-audit` + test, `scripts/audit-config-keys.sh`),
with git and DB explicitly withheld — the parent session commits and applies.

**Claims verified rather than carried:** "no automated caller of
audit-config-keys.sh" — grepped tree-wide before writing it into the submission's
risks block; all hits are comments. The no-op-detector phrase blocklist checked
against every sketch before submitting (RUNBOOK landmine, corr 4a227ed9).

## 2026-08-03 (later) — implementation, migration applied, council round 1: REVISE

> **CORRECTED 2026-08-03:** the entry above says "agent runs = 0 all-history".
> **Wrong** — `orchestration_states` is retention-clocked and cannot support
> "all-history"; five `products.verified_date` stamps dated 2026-07-22 are
> consistent with a reaped manual run. Caught by the council's
> `prior_art_librarian` seat. Full row in `WRONG_CALLS.md` 2026-08-03. The
> safety argument now rests on: the bad keys never resolved (so any historical
> run behaved as defaults), `scheduled_tasks` = 0, and the only dispatch path is
> manual kcat.

Implementation landed by the opus subagent (all six files), verified
independently: `go test ./cmd/config-key-audit/... ./platform/orchestration/datahelpers/...`
both ok; live report BEFORE the migration showed the defect through both new
detectors (`refresh_product_specs: category?, limit?` under UNKNOWN KEYS — the
CheckConfig arm working from HEAD source — and both rows under SUSPICIOUS KEYS),
exit 1.

Migration 298 applied by hand 12:06 BST: snapshot f4f12221 (NOTICE confirms),
UPDATE 1, both DO/RAISE verify blocks passed, COMMIT. Artefact verified: live
config now exactly `{"limit": "input_data.limit", "site_id":
"site_record.site_id", "category": "input_data.category"}`. Post-migration
report: SUSPICIOUS KEYS `none`, this agent gone from UNKNOWN KEYS; residual
exit 1 is three pre-existing unknown keys from the bug-136 family
(`plan_sections: domain`, `run_discovery_checks: check_domain`,
`triage_detected_items: target_domain`) — other lanes' territory, not touched.
Recorded in the ledger via `--record-only` with the verification note.

**Council round 1 (corr 9521d62f): REVISE**, gating HIGH from `editquality` —
the plan's rollback story used `snapshot_agent(text,text)` without proving which
table the two-arg overload writes (there are two overloads writing to two
different tables — a recorded landmine). Measured both ways:
`pg_get_functiondef` shows `snapshot_agent(text)` → INSERT INTO
`agent_definitions` (is_snapshot row) and `snapshot_agent(text,text)` → INSERT
INTO `agent_definitions_backup`; and at the artefact, my applied migration's
snapshot row IS in `agent_definitions_backup` (snapshot_taken_at 11:06:17Z,
reason names 298) **with the old `category?` key preserved** — rollback is real.
Zero `is_snapshot=true` rows for this type in `agent_definitions`. The
`doc_notes` migration row also landed (11:06:17Z, pipeline/["migration"]) —
answers `tooling_provenance`'s silent-failure concern at the artefact.
All other objections measured or answered in the round-2 rationale; resubmitting
on the same correlation.

## 2026-08-03 (close) — ticket moved to bugs_closed; landmine appended

Bug file moved to `/bugs_closed/` with the closing account appended; 016b §10 row
updated to CLOSED (with the run-count overreach corrected in place). LANDMINES
gained the `go run`-collapses-exit-status entry (found by this lane's own first
draft); `landmines-sync.py --apply` ran — and then `landmines-verify-dispatch.sh`
said "Nothing needs verification", which is EXACTLY the trap recorded in
`bugs_open/163`'s WRONG_CALLS cross-reference: running `--apply` first consumes
the NEEDS_VERIFICATION signal. Not fought here — 163 also documents the verifier
cannot mechanically confirm path-bearing symbols anyway; the entry stands on its
measured evidence.
