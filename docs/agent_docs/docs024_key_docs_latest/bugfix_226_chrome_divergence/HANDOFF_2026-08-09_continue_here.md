# HANDOFF 2026-08-09 — bugfix 226 chrome divergence guard: continue here

**For the next session picking up this lane. Read PLAN (design + blast radius)
and NOTES (evidence trail) beside this; this file is only what to DO next.**

## State in one paragraph

The mechanism is SHIPPED and LIVE end to end: mig 344 (archive trigger,
fail-closed, sole trigger on `site_components`, probe-verified) has been live
since 08-08 and has already archived four real production overwrites
(webdesign.uk + leopardessconsulting.co.uk chrome, all `unstamped`, zero
errors); the Go half (digest stamp + classify + `chrome_divergence_overwritten`
item + ledger read-back fallback) is on **v1.0.1270**, binary-verified on both
main `agent-chassis` replicas with positive AND negative greps. Commits:
`1eae32644` (round 1), `6bee2708e` (round 2), plus the round-3 commit that
carries this file. Council trail `cffbfec4-3bec-4577-8844-d17c546ded3e`:
round 1 REVISE (guardian) answered, round 2 REVISE (bug_historian) answered,
**round 3 pending** (run orch `50924d69-73a9-4af9-9649-4c3bcd74aa8f`,
submitted ~09:XX UTC 08-09).

## Do next, in order

1. **Read the round-3 verdict.** Query WITHOUT a time filter (a BST-vs-UTC
   filter blinded the round-2 watcher — WRONG_CALLS 2026-08-09):
   `SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts WHERE
   correlation_id='cffbfec4-3bec-4577-8844-d17c546ded3e' AND
   kind='council_report' ORDER BY created_at;` — three rows exist (r1+r2
   revise); the fourth is round 3. Body: same query with `body`, **the column
   is `body` not `content`**. APPROVED → nothing to do (commits carry
   `Council-Submitted:`, 098 credits automatically). REVISE on the page-side
   scope again → do NOT expand the code; the seat split is recorded as an
   OWNER CALL block in `bugs_open/229` — escalate to the owner, that is the
   designed path. REVISE on anything else → judge it on its evidence.
2. **Run close criterion 2 (end-to-end protocol)** — the bug file's CLOSE
   CRITERIA block has the corrected steps. The trap: all 57 rows were
   unstamped at roll time, so it is a TWO-step: (a) rebuild a test site's
   chrome once → the new code stamps `rendered_html_digest`; (b) hand-patch a
   throwaway string into that slot's `rendered_html` via psql; (c) rebuild
   again → require the WARN (`were overwritten and archived`), the
   `chrome_divergence_overwritten` item (key carries the patched digest's
   first 12 chars), and the `hand_patched` ledger row; (d) negative control:
   rebuild once more untouched → byte-identical → no new archive row, no item.
   Mind the ~300s no-dispatch window after any chassis pod restart.
3. **Watch the 117 wave's first pass (close criterion 3).** `SELECT
   count(*) FILTER (WHERE rendered_html_digest IS NOT NULL) FROM
   site_components;` — 0/57 as of 08-09 morning means the wave has not fired.
   After it fires: archive rows for every changed slot, zero trigger errors,
   stamped count climbing. Also the guardian's watch item: item count per
   site+slot vs distinct digests (runaway-loop accumulation check).
4. **Then the bug is done in substance** — per the owner 08-06 ruling it
   STAYS in `bugs_open/` (do not move to bugs_closed); update its header and
   the MEMORY topic file `bugfix-226-chrome-divergence.md` + the
   MEMORY_workstreams line instead.

## Standing landmines for whoever touches this

- The archive is a **DB trigger** — no Go grep shows it (LANDMINES entry,
  synced). A chrome write failing with `site_component_history` in the error
  is the trigger refusing to destroy unarchived bytes — fix the ledger, never
  drop the trigger casually (ROLLBACK sidecar exists for a deliberate one).
- `rendered_html_digest` belongs to the render path ONLY. Never backfill it,
  never re-stamp it from another writer — either move silences the detector.
- Pod verification: enumerate by IMAGE (65 pods run the chassis image; the
  label selector shows 2). The deployment that matters is `agent-chassis`.
- `bugs_open/229` (page-side sibling) is UNOWNED and carries the design
  constraints — do not copy-paste the chrome trigger there (DELETE+INSERT
  family; shared-abstraction question; possible RFC per the architecture
  seat).

## Unrelated but load-bearing, left for its owners

- HEAD's test suite is red on `TestValidDocSubjectTypes_Lockstep` since
  `e1628f7df` (RFC_015: migration 340 added `'decision'`, the Go
  `validDocSubjectTypes` list was not updated — the fourth instance of the
  both-halves landmine at LANDMINES.md:646). Owning sessions were active on
  08-08 evening; if it is still red when you read this, it may be abandoned —
  check `who-owns` and live transcripts before touching.
