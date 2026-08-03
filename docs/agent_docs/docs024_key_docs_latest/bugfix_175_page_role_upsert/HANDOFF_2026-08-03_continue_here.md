# HANDOFF — `bugfix_175_page_role_upsert`, 2026-08-03

**Read this first, then `NOTES` for the missteps and `RUNBOOK` for the commands.**
Written because the originating session's context ran long, not because the work stalled.

## State in one paragraph

`bugs_open/175` is **CLOSED, moved to `bugs_closed/`, and fully live** on chassis
`v1.0.1237` (both replicas, pod-verified 2026-08-03 08:47Z). The council trail is
**approved → REVISE → approved** on one correlation, `e78c62e3-7f01-48f1-b083-924eaccd195a`.
`architecture_review/RFC_010` is **RATIFIED by the owner and implemented**. One follow-up
was filed and is **unowned**: `bugs_open/185`. Nothing is in flight; nothing is owed to a
council; the working tree carries none of this lane's changes uncommitted.

## What was built

`UpsertPageForRole` — `platform/orchestration/actions/page_role_upsert.go`, register entry
**PBP-027**. One collision seam for any arm whose `page_type` is a compile-time constant.
Four outcomes on `INSERT … ON CONFLICT (site_id, name) DO NOTHING`:

| collision | outcome |
|---|---|
| none | **created** |
| row holds the SAME role | **refreshed** — the caller's declared `Refresh` subset only |
| DIFFERENT role, never shipped, `AdoptUnshippedRows: true` | **adopted** — whole takeover incl. `page_type` |
| DIFFERENT role, never shipped, field NOT set | **refused** — nothing mutated, nothing filed |
| DIFFERENT role, HAS shipped | **refused** — nothing mutated, `mistyped_deployed_page` filed `needs_human_review` |

Callers (all four declare `AdoptUnshippedRows: true`): `deploy_tool_action.go:369` (tool
page) and `:524` (companion guide), `create_tool_component_action.go:412` (companion
guide), `create_report_page_action.go:159` (report). `apply_gap_plan_action.go`'s
`applyNewPage` is **deliberately not a caller** — its role comes from an LLM plan, and
`bugs_closed/081` declined exactly that authority. What the two DO share is
`fileMistypedLivePageItem`, so both refusal sites produce one item shape and one
`item_key`.

Recurrence is closed mechanically: `check_partial_page_upsert` in
`scripts/pattern-check.py` (measured 4 findings at pristine HEAD → 0 after).

## The owner's four rulings, all implemented and live

1. **ADOPT is opt-in, default OFF** (`AdoptUnshippedRows`). Generalised into CLAUDE.md:
   *new authority on a shared seam ships as an opt-in field with the unsafe default OFF,
   not as a documented contract.*
2. **Converging producers onto one `item_type` needs no RFC**, provided the producer set
   and shared `item_key` are named in the register entry. Also in CLAUDE.md.
3. **Refusal-becomes-error stands** — two arms hard-error, two log and continue.
4. **`bugs_closed/081`'s guard converged now**: it read `build_status = 'deployed'`, it now
   asks `NOT (datahelpers.NeverDeployedPagePredicate)`.

## What is NOT done — pick up here

1. **`bugs_open/185` is filed, unowned, census complete, no fix attempted.** ~10 detectors
   select `p.build_status = 'deployed'` on `pages` and are blind to **28 active** (35
   counting archived-but-possibly-served) pages that HAVE shipped under another status.
   False negatives, not corruption. Its fix candidate 1 warns: converging them will make
   ten checks start reporting on pages they have never seen, so **measure what each newly
   reports before shipping** — the first consequence of fixing a false-negative is a burst
   of findings.
2. **`bugs_open/185` fix candidate 2** is the one two council seats pushed hardest:
   `datahelpers.NeverDeployedPagePredicate` and `queryresolve.FetchablePageEligibilitySQL`
   are one judgement in two constants. They now cross-reference each other by comment; the
   seats wanted actual convergence. Read `queryresolve.go:210-236` first — it documents a
   deliberate family of three, so this is not a free merge.
3. **DEPLOYED IS NOT PROVEN.** No arm has hit a collision since the roll. Nothing has
   exercised the branches in production, `mistyped_deployed_page` still has **zero rows**
   from any of its three producers, and there is no live `PageRoleAdopted` /
   `PageRoleRefused` log line. If you want that proof, the reachable case is a tool named
   so its page collides on `robot-hands.com` (`gripper-selection-guide` /
   `selection-guide`, both `content`, both deployed).
4. **Not converted, deliberately:** `create_tool_component_action.go:288` creates its own
   tool page with a plain `INSERT` and no `ON CONFLICT` at all — a collision raises a
   unique violation, deletes the component and errors. Loud and fail-closed, so outside
   175's silent class. Converting it would make re-runs idempotent, a behaviour change
   nobody asked for.

## The five things this lane learned that will cost you if you skip them

1. **A detector must be seen to FIRE before you trust its clean run.** I measured my new
   `pattern-check` rule against a tree that already had my fix, got 0 findings over 1,120
   Go files, and nearly recorded that as "no false positives". Pristine HEAD: 4, exactly
   the census.
2. **Ask whether the shared PREDICATE exists, not just whether the shared HELPER does.**
   The council asked about a helper; I checked, answered, stopped — and shipped a
   hand-rolled liveness predicate that was **wrong on 11 live rows** (three of them lendzy
   tool pages created that day). `datahelpers.NeverDeployedPagePredicate` already existed
   with three consumers and a test forbidding the exact clause I added. The reusable unit
   was one line of SQL, which is what nobody searches for.
3. **When a mutation PASSES, the test is failing to see the guard.** Reverting 081's
   widened guard left every test in its file green — both predicates agree on the inputs
   those tests supply. The discriminating input is a `needs_rebuild` page that HAS shipped.
   Related: a mutation that fails to **compile** is not a mutation test (mine removed the
   last use of a package; re-run with the import kept referenced).
4. **`GROUP BY` a status column BEFORE filtering a census on it.** I published "28" from a
   `status='active'` filter I never looked behind; 7 archived-but-possibly-served rows were
   hiding there. A trap recorded against `sites.status` is a trap against the **column
   shape**, not that table.
5. **A deployment claim has a shelf life of minutes here.** I told the owner twice that
   nothing was live; another session rolled in between, twice. Grep the pod at the moment
   you assert it.

All five are in `NOTES`; (1), (2), (4) and (5) are in `WRONG_CALLS.md`.

## Landmines this lane added (both synced to `doc_notes`)

- **`pages.build_status = 'deployed'` is NOT "is this page live"** — 35 of 46
  `needs_rebuild` rows carry a `deployed_at`. Use the shared predicate; do **not** name the
  status (its own test forbids that — naming it produced a 34-page false-positive class for
  the nav lane).
- **THREE `pages` upsert helpers with OPPOSITE collision policies.**
  `site_db_actions.upsertPage` **re-types** whatever it hits (correct — plan-sync is the
  authority on what a page is); `UpsertPageForRole` **refuses** a live row of another role.
  Both compile, both return a page id. **Choose by where your `page_type` comes from.**

## Verification recipe, if you touch any of this

`RUNBOOK_page_role_upsert.md` has the commands with their gotchas. The two worth repeating:
the generic `ON CONFLICT (site_id, name) DO UPDATE SET` grep is **not** a negative control
(it returns 4 and that is correct — five arms keep the statement deliberately); and the
shared-predicate count is the strongest check available, because that constant is
interpolated per consumer, so the count IS the consumer count (3 pre-existing + this seam
= 4; 5 once 081's guard converged).

## Commits, newest first

`12e9335db` liveness record · `d7c119c11` RFC/register live · `004206913` + `b460937de` +
`1c71d7cff` the two corrections and their wrong-calls · `47b4d9f8b` the sibling audit +
`bugs_open/185` · `0f2d57a9b` RFC REVISE record · `85d7e0ca5` + `4ee695cc1` the ruling
round · `023f6624a` the predicate correction · `6192cc9d2` close + RFC_010 · `588adb6f1`
016b §9 follow-up · `9897cb82f` docs · `cbbecb021` **the fix itself**.
