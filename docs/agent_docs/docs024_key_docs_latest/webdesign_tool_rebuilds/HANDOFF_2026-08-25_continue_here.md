# HANDOFF — webdesign tool rebuilds. START HERE. Written 2026-08-25 ~11:15Z. Supersedes `HANDOFF_2026-08-22_continue_here.md`.

**STATE (updated 20:15Z): 43 of 63. NOTHING IN FLIGHT. Two things changed under the lane while it
was idle, both verified first-hand (NOTES 20:15Z):
⚠ **A CHASSIS ROLL LANDED 19:07Z, after every build in the 08-25 grind.** The "fresh roll 09:27Z —
verified for this lane" paragraph below is TWO ROLLS STALE — do not read it as current. Re-verified
on the new binary: `adopt_existing_page` and `page_adopted` PRESENT with a junk-literal control
absent, live config flag still `true`. **NOT re-verified: the 360 tombstone guard** — the next
filing's post-retire re-read is that check; treat it as unverified until then.
⚠ **NO TOOL ON THIS SITE HAS EVER HAD A `tool_acceptance` ROW** — 0 rows all-history, because
`site-discovery-rotation-design` is `enabled=f` (alone of the four) and `design-discovery-agent`
has 0 runs since 08-11. **The serve-grades in NOTES are the ONLY grading these 43 rebuilds have
ever had.** Nothing in this lane may defer behavioural correctness to that pipeline while the
switch is off. Re-enable is the OWNER's call (put to him in README); `bugs_open/401` covers the
watchdog that missed it. Prior line follows.**

**STATE (updated 16:20Z): 43 of 63 SERVE-CONFIRMED (#43 golden-ratio, NOTES 16:20Z). NEXT:
monolith-splitter (9,037), then head-architect (9,212), asset-formatter (9,222), layout-generator
(9,223), insight-injector (9,369). NOTHING IN FLIGHT.
⚠ **THE BRIEF RULE FROM 15:45Z IS CORRECTED** — a 2,701-char brief died inside the "safe" window.
It is not character count, it is SURFACE AREA, and the lever is: **describe the tool to BUILD, not
the defects to fix.** Golden-ratio built at 1,551 chars once the defect archaeology came out; keep
that story in NOTES. RUNBOOK corrected in place.
⚠ **TWO ITEMS OWED, both real capability gaps, both wanting their own `replace_existing` filing:**
(1) cubic-bezier — keyboard access to the two drag handles; (2) golden-ratio — a REAL crop export
(cropped to the ratio, guides NOT burned in). The rebuilt golden-ratio has NO download at all: I cut
it to fit the token budget, which is a removal, not a fix. Do not let it be found later as a mystery.
Prior line follows.**

**STATE (updated 15:45Z): 42 of 63 SERVE-CONFIRMED (#42 cubic-bezier, NOTES 15:45Z). NEXT:
golden-ratio (8,754), then monolith-splitter (9,037), head-architect (9,212), asset-formatter
(9,222). NOTHING IN FLIGHT.
⚠ TWO NEW RULES, both learned the expensive way today, both now in the RUNBOOK:
(a) **keep the add_tool `description` near 2,000–2,800 chars** — a 4,431-char brief killed
cubic-bezier's first build at `max_tokens` and the WORK ITEM STILL SAID `complete` with `error`
NULL; only the RUN grade showed it. (b) **`related_pages` mentions never land on this site** —
`deferred` is a terminus, not a gate (0 of 80 `tool_crosslink` rows have ever completed). Keep
filing the key; never claim a mention. Prior line follows.**

**STATE (updated 15:00Z): 41 of 63 SERVE-CONFIRMED (#41 text-sanitizer done, NOTES 15:00Z — the
"remove invisible characters" toggle that replaced zero-width chars with a SPACE is dead; sighting
#19). Phase B continues smallest-first: cubic-bezier (8,754) next, then golden-ratio (8,754),
monolith-splitter (9,037), head-architect (9,212), asset-formatter (9,222). NOTHING IN FLIGHT.
⚠ READ THE CROSSLINK CORRECTION in NOTES 15:00Z before you report a cross-mention: `deferred` is a
TERMINUS, not a gate — 0 of 71 `tool_crosslink` rows have ever completed on this site. Keep filing
`related_pages` (the finding is targeted and parked, and that is the raw material); just never claim
a mention landed. Prior line follows.**

**STATE (updated 13:50Z): 40 of 63 SERVE-CONFIRMED (#40 entropy-meter done, NOTES 13:50Z — the Infinity-bits overflow dead, all arithmetic in log space). Phase B continues: text-sanitizer (8,607) next. Prior line follows.**

**STATE (updated 13:10Z): 39 of 63 SERVE-CONFIRMED — #39 focus-ring's owed grade PAID (NOTES 13:10Z). Phase B continues smallest-first: entropy-meter (8,325) next. Superseded owed-line follows.**

**STATE (updated 12:45Z): 38 confirmed + #39 focus-ring RETIRED with serve-grade OWED (rerender `4f0a3002`; controls pinned in NOTES 12:45Z — grade it FIRST on pickup, then dispatch nothing for it: self-contained, no orphan). Phase B continues smallest-first: entropy-meter (8,325) next. Prior state line follows.**

**STATE (updated 12:10Z): 38 of 63 — PHASE C COMPLETE (fluid-typography and vibe-equalizer both DONE, NOTES 11:35Z/12:10Z; micro-cms reclassified to the rich-app finale — it IS Flat-File Micro CMS). NEXT: Phase B, 23 self-contained ≥8 KB tools smallest-first (focus-ring 8,148, entropy-meter 8,325, text-sanitizer 8,607, cubic-bezier 8,754, golden-ratio 8,754, …), then the FIVE rich apps one at a time, owner-reviewed. New milestone summary: SUMMARY_2026-08-25_phase_c_complete.md. Then Phase B: 24
self-contained ≥8 KB tools (incl. head-architect, reclassified 08-24) with the FIVE rich apps LAST,
one at a time, owner-reviewed (standing ruling). NOTHING IN FLIGHT — no open add_tool, no pending
retires, no unwatched rerenders.**

**Fresh chassis roll 2026-08-25 09:27Z — verified for this lane** (probes 11:10Z): binary carries
`adopt_existing_page` AND the 360 tombstone guard (junk control clean); live config carries the
adopt flag and the `suggest_related_pages` picker wiring. #35 built ON this roll, so it is also
behaviourally proven. Nothing this lane depends on regressed.

Read: this file → `PLAN_2026-08-15_…` (rulings) → `RUNBOOK_…` (all recipes: retire race, tombstone
re-read, Phase C asset half, related_pages) → `NOTES_…` (evidence, newest at bottom — every entry
since 2026-08-22 11:06Z is this arc) → `SUMMARY_2026-08-21_phase_a_complete.md` (a new summary is
DUE at Phase C completion, not before).

## The recipe (proven 35 times; Phase C variant)

1. Fetch the live page AND its sidecar(s) cache-busted; read the sidecar IN FULL. Expect the
   "asserts something untrue about itself" class — **19 sightings as of 2026-08-25**; the brief's job is to name the
   honesty invariant that makes the claim TRUE by construction.
2. Gates before filing: library-claim (0 rows or pin fork identity), local active component (0),
   open add_tool (0), **adopt flag present** (it was removed un-snapshotted once — migration 558),
   margin on the page's queued rerender. **`related_pages`: 1–3 EXISTING non-tool `pages.name`
   values picked by topic** (RUNBOOK section; names verified resolving pre-file). The ask-when-absent
   picker (353 lane, mig 602) is live and PROVEN (NOTES 2026-08-25 10:00Z) as the safety net — but
   deliberate picks beat it; carry the key.
3. File with the full spec; attend with foreground poll loops (≤600 s per call, chained).
4. Grade the RUN (`page_adopted='true'`, no `already_exists`, no `__step_error` — an item can read
   complete/error NULL on a dead run), quick structural grade, then **RETIRE IMMEDIATELY** (guarded
   txn, DO/RAISE asserts, md5 pinned, post-commit re-read — a typo'd heredoc once silently skipped
   the whole retire and only the re-read caught it).
5. Mechanism-grade the component at the DECIDING code arms (never the tool-doc header — assembly
   STRIPS it; pin serve-grade literals from CODE, twice burned). Validate every control BOTH ways.
6. Attend the assemble; serve-grade cache-busted: http=200 first, last-modified > completed_at
   (**the S3 write lands 1–2 min AFTER the item completes** — a completed_at-adjacent fetch reads
   as a false FAIL), negatives 0 incl. `src="<sidecar>"`, positives present. Tombstone re-read.
7. Dispatch the sidecar's dry-run retraction, record the refusal + orphan (bugs_open/365 list,
   **8 so far**; the shared `/tools/assets/webdesign-couk-header.js` goes with the LAST ported page).

## Next up — monolith-splitter (9,037), analysis NOT yet done

Nothing is pre-analysed — start at step 1 of the recipe. Phase B order from the census (re-run it,
do not trust this list): monolith-splitter 9,037 · head-architect 9,212 · asset-formatter 9,222 ·
layout-generator 9,223 · insight-injector 9,369 · … then the FIVE rich apps LAST, one at a time,
owner-reviewed.

**Two items OWED, from #42 and #43 — each wants its own `replace_existing` filing, not a fold-in:**
1. **cubic-bezier — keyboard access** (arrow-key nudge) for the two drag handles, cut to fit the
   token budget. A real gap on a site that publishes `/learn/accessibility/focus-states.html`.
2. **golden-ratio — a REAL crop export.** The rebuild has NO download at all; the ported one had a
   "Download Crop" button that exported the photograph with the guides burned into it and cropped
   nothing. Wanted: the image cropped to the chosen ratio, centred on the overlay, guides NOT drawn
   on it — which is exactly what the ported code's own comment was reaching for and never did.

## Standing rules (unchanged, load-bearing)

- ONE at a time (serial item key); file ONLY what you attend in-turn; margin-gate the page's queued
  rerender; `ahead=0` is ambiguous (the dispatcher FIFO is fleet-wide — other sites' rows matter).
- Retire = status flip ONLY, never delete; revert handle = row id + length + md5 recorded pre-file.
- Grade the RUN, the COMPONENT (by mechanism), the SERVED page — never a status.
- Counts carry their census date (owner 08-24). A `[MEASURED]` claim needs a disconfirmable shape.
- Probe traps: exit 137 (timeout-killed exec) looks like grep's exit-1 "absent" if you test only
  $?≠0; a bare leading `sleep` is blocked (poll the actual condition); `.git/index.lock` can be a
  concurrent session's transient — check age + live processes before touching.

## Open threads

- `bugs_open/365` (sidecar files have no retirement path) — routed at DGH-010 owner; orphan list in
  the file, 8 entries as of 08-25.
- `bugs_closed/360` residuals: check_literal_markdown still SCANS tombstones (277 lane, 356 §6-B);
  486's INSERT predicate (283 lane). Both now land on the live guard and skip — nuisance, not damage.
- 098 credits: corr `4007ce96` (tombstone guard) and `a367b63e` (558 restore) both APPROVED —
  trailers resolve automatically; nothing owed.
- The 353 lane's picker demand proof: delivered PASS 2026-08-25 (NOTES 10:00Z); their close is theirs.
- Phase B rich-app cadence (owner sight per page) unchanged; a decision on WHEN is the owner's.

## Cold-start dependencies

DB: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`
Site `6b49db8e-d447-4467-8277-4f3018af9897` · tally check: the RUNBOOK's GROUP BY build_status
query (expect removed=35+N) · adopt-flag + picker check queries: NOTES 12:35Z 08-22 and this file's
gates line. All ids/md5s for every rebuilt tool: NOTES, per-tool entries.
