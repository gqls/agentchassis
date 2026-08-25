# HANDOFF — webdesign tool rebuilds. START HERE. Written 2026-08-25 ~11:15Z. Supersedes `HANDOFF_2026-08-22_continue_here.md`.

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
   "asserts something untrue about itself" class — **15 sightings**; the brief's job is to name the
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

## csp-builder — analysis DONE 2026-08-25 (file next, with related_pages ['learn-security-cdn-risks','learn-security-xss-vulnerability'])

Slot `5190c47e-72d5-4328-a073-0f992151cf1d` (13,190, md5 `36710d18…`), page `a5803b85…`, sidecar
`policy-generator.js` (3,318 B, read in full). Ported defects for the brief:
- **STATE BUG, the load-bearing one: `resetPolicy()` resets only script/style/img/font — a
  `connect-src` value added by a checkbox or textarea is NEVER removed when unchecked/cleared**
  (the Set accumulates for the session), so the displayed policy does not reflect the current
  selections — the output lies about the inputs. The rebuild regenerates every directive from the
  current inputs each run.
- Copy: `.then` only, no `.catch` (rule 15); `window.copyCSP` global + inline onclick (rule 16).
- No validation/weakening warnings: `'unsafe-inline'`, `'unsafe-eval'`, `*` typed into a textarea
  land silently in the policy — a CSP builder should visibly flag policy-weakening values (the
  teaching point). Garbage tokens should get an inline message.
- Keep: the sensible fixed defaults (`object-src 'none'`, `base-uri 'self'`, `default-src 'self'`)
  stated on the page; the Google-Fonts → `font-src fonts.gstatic.com` smart-add (make it symmetric
  — it already is, since font-src is reset each run); newline-stripped header-ready copy.

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
