# HANDOFF — `bugs_open/215` quiet mode, continue here (2026-08-11 evening)

Cold-start for a fresh chat. The fix is **built, approved, live and enabled
nowhere**. Nothing is half-finished in the tree; the next moves are measurement
and two decisions.

## 1. What this lane is

`bugs_open/215`'s **crash mode** (two plan entries canonicalising to one name,
killing the whole replan write) was fixed and live back on `v1.0.1276`. The bug
stayed open for the **quiet mode**: a replan can carry a page under a spelling
that DENOTES an already-realised page, nothing reconciles the two, and the
name-keyed upsert inserts a **second `pages` row** — a permanent 404 that is a
valid internal-link target from creation, or worse, a second page that also gets
built. Owner ruling 2026-08-11 §1 left this "in `bugs_open/215`'s remaining scope
on its own merits".

## 2. State, as of this handoff

| | |
|---|---|
| code | **LIVE on chassis `v1.0.1288`**, artefact-verified on BOTH replicas |
| council | **APPROVED** round 3, corr `56e13695-17cb-48ec-bc6b-0371fde8b717` (revise → revise → approved) |
| enabled on | **no site.** All three gates default OFF |
| damage remaining | **7 both-deployed twin pairs, 4 domains** — untouched, needs an owner call per pair |
| bug file | stays **OPEN**; full account in `bugs_open/215_HANDOFF_...md`, newest sections at the bottom |
| register | **PLAN-048** in `docs026_concept_register/register/site-plan-and-reconciler.md` |

Commits: `65c1984d0` (shared identity keys + site policy) · `7a066dba1` (writer
guard, both surfaces) · `b36163fb3` (reconciler layers) · `038211dd8` (the gating
REVISE response) · docs/register/runbook/landmine in `25fdc53ab`, `d5dd65a06`,
`86ed9d723`, `88897f0c6`, `a8d678800`, `10f794e7d`, `b8457a9fc`, `3e9d2035d`.

## 3. What the fix actually does (one paragraph)

Two coupled halves. **(a)** `reconcilePlanWithRealised` gained three twin-identity
match layers — normalised URL path, predicted canonical identity, and a stem
heuristic (both directions) — all sharing ONE extracted snap arm with the
pre-existing exact-URL rename, so the `bugs_open/050` empty-page rules and the
`bugs_open/151` fact carry cannot drift between them. **(b)** Both
canonicalisation surfaces now HONOUR a realised page's stored identity instead of
re-deriving it. Half (b) is not in the filed bug and is the part to keep
distrusting: it exists because `CanonicalisePage` **cannot express a legacy
identity**, and **71 live shipped rows fleet-wide are not fixed points of it** —
for those, a reconciler-only fix is silently undone downstream.

## 4. DO THIS FIRST next session — read the dark-launch population

The gates are off, but the layers still COUNT what they would have done. **The
baseline at the roll was 0/0/0, and those zeros are not evidence** — no replan had
run through the new path yet.

```sql
SELECT error_code, count(*), min(occurred_at), max(occurred_at)
FROM agent_error_log
WHERE error_code IN ('PLAN_PAGE_IDENTITY_TWIN_OBSERVED','PLAN_PAGE_STEM_TWIN_OBSERVED',
                     'PLAN_PAGE_STEM_TWIN_REFUSED','PLAN_PAGE_IDENTITY_SNAPPED')
GROUP BY 1 ORDER BY 1;
```

Every `*_OBSERVED` row is a second page identity that was about to be written.
That population is the entire evidential basis for enabling anything. If it is
still all-zero, **no replan has run** — check before concluding the layers are
inert (`SELECT max(created_at) FROM site_plans;`).

## 5. The two open decisions (both owner's)

- **O1 — enable the gates, per site.** Three keys in the `site_specs`
  **`structure`** aspect, beside `url_shape`: `honour_realised_identity`,
  `twin_identity_snap`, `stem_twin_snap`. Recommended order: read §4, then enable
  `honour_realised_identity` + `twin_identity_snap` on ONE site (fundamentallyai
  is the natural pilot — it holds three of the known twin shapes), leave
  `stem_twin_snap` until its own OBSERVED count has been read.
  **Rollback is safe:** turning it off does not move live URLs; the next replan
  simply reverts to minting twins, i.e. the pre-fix bug, no worse.
  **Do NOT enable on the five decomposed sites** until `bugs_open/204` is fixed.
- **O2 — the 7 both-deployed pairs.** The fix REFUSES these on purpose (snapping
  would hand the writer two entries with one name, and richer-wins would evict a
  live page). Procedure, population and ordering:
  `RUNBOOK_2026-08-11_duplicate_page_identity_remediation.md` in this directory.
  Needs a survivor decision per pair before anything is executed; the sweep front
  owns fundamentallyai's execution.

## 6. Traps this lane hit — do not re-walk them

- **Verifying the deploy.** The documented
  `logs -l app=agent-chassis | grep 'build provenance'` returned **1.4MB of
  council payloads quoting the phrase**, and the startup line had rotated out of
  both pods. Probing `/proc/1/exe` for my **commit SHAs** returned absent for
  every sha *including the fabricated control* — no positive control, so it
  proved nothing. What works: probe **added string literals** with a **one-letter
  near-miss** negative control, on **both** replicas.
- **The identity marker's route.** `identity_authority` travels from the
  reconciler to the write surfaces through **`collected_data.site_plan`**, NOT
  through `site_plan_pages` (which has no column for it). `validate_plan`'s
  `output_field` is `site_plan`; both surfaces read it via `extractPagesFromPlan`,
  which appends each page map WHOLE. Make that extraction field-selective and the
  writer guard silently dies — `TestReconcile_MarkerSurvivesTheStepBoundary`
  exists to fail loudly instead.
- **Re-adopting a site DROPS all three flags** (`apply_adoption_plan_action.go`
  replaces the structure spec; `WriteSiteSpecAction` deep-merges and is safe).
  Fails safe, but silently. Now in `LANDMINES.md`; re-check the spec after any
  adoption run, with `data ? 'key'`, never `->>'key' = 'true'`.
- **Three `pages` upsert helpers with opposite policies.** Only ONE of the two
  guarded surfaces writes `pages` at all (`SyncPagesToDBAction` → `upsertPage`);
  `WriteSitePlanAction` writes the PLAN table. `upsertPage` is the correct helper
  by that landmine's own rule (a role arriving from a plan belongs on the
  plan-sync path).

## 7. Loose ends, honestly listed

- **090 diagnosis `38099787-c7f9-46d4-b75e-3a1867fcaf41`** (archived pages being
  rebuilt and re-deployed) completed with **evidence bundles but no verdict
  artifact**. Nobody has read a root cause. The narrower question it should be
  read as: *should the build/deploy path refuse a page whose `status` is
  `archived`, rather than relying on the plan never naming it?*
- **Two `PLAN_PAGE_MERGE_LOSSY` rows** from the 08-11 census replan tripped the
  standing richer-wins revisit trigger — composed-vs-composed, which is not the
  case richer-wins was ratified on. That is the same underlying duplicate-page
  condition as O2 and is on the owner's radar via the lane README.
- **Council submission JSON** lives at
  `COUNCIL_SUBMISSION_215_quiet_mode_2026-08-11.json` — sketches match the shipped
  tree as of round 3. If you resubmit anything here, update the sketches: round 2
  was a REVISE purely because the document had gone stale against the code.
- The council correlation `3cd9fd92-...` on commit `25fdc53ab` is **dead** (that
  submission was rejected as invalid before any seat ran). The live one is
  `56e13695-...`.

## 8. Three wrong calls recorded against this lane

In `WRONG_CALLS.md`, 2026-08-11, worth reading before doing similar work:
a "direction flips" invariant implemented in **one** direction (caught only
because the fixture came from live data rather than invention); a root cause
inferred from a log row's **title** when the row carried the answer in its
**fields**; and a finding filed as a "distinct defect" that the owning lane had
**predicted in writing five hours earlier** — the cheap check being `ls -t` on the
owning lane's directory before filing anything about its site.
