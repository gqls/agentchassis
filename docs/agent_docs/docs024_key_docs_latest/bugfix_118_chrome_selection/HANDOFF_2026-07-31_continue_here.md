# HANDOFF — bugfix_118_chrome_selection — continue here

**Written 2026-07-31 ~21:50 UTC.** Read this first, then `NOTES_chrome_selection.md`
(the technical log, newest at the bottom) and `SUMMARY_2026-07-31b_chrome_selection.md`
(current state, written to be read aloud).

## State in one paragraph

`bugs_closed/118` is **DONE** — one chrome-eligibility predicate, live and
pod-verified on `v1.0.1219`, council APPROVED at round 1, and the fleet repointed on
the owner's ruling (21 assignments, 28/28 header+footer slots now render from an
ACTIVE component). `bugs_open/166` — the repair that could never repair — is **FIXED
IN CODE and committed** (`39afbf697`), **awaiting a council verdict and a roll**.
`bugs_open/167` is filed and untouched (owner call). Nothing is half-edited; the
working tree is clean of this lane's work.

## THE ONE THING OWED RIGHT NOW

**Read the council verdict for `e242e9d3-1e5a-4b14-a16f-fbff9ca86d35` and act on it.**
The code is already on the shared branch under a `Council-Submitted:` trailer, which
is correct for a pre-verdict commit (098 credits it automatically once approved, no
amend needed). **UPDATED 22:0x UTC: round 1 came back REVISE and has been answered; round 2 is
running on the SAME correlation** (`RESUBMIT_CORR`, so the trail accumulates). What
is owed is now round 2's verdict. Budget ~30-60 min per round under current load —
round 1 took ~55 minutes against 118's nine, which is fleet latency, not a dropped
dispatch.

```sql
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = 'e242e9d3-1e5a-4b14-a16f-fbff9ca86d35';
SELECT metadata->>'decision', body FROM diagnosis_artifacts
 WHERE correlation_id='e242e9d3-1e5a-4b14-a16f-fbff9ca86d35' AND kind='council_report'
 ORDER BY created_at DESC LIMIT 1;
```

- **APPROVED** → answer any medium objections in code where they earn it (that is what
  the 118 round did: 3 of 5 earned edits), then commit with
  `Council-Reviewed: e242e9d3-1e5a-4b14-a16f-fbff9ca86d35`.
- **REVISE** → the objections come back with the reviewers' own checks already
  answered. Resubmit with `RESUBMIT_CORR=e242e9d3-…` so the trail accumulates.
- **NEVER** write `Council-Reviewed:` on a verdict you have not read — that is the
  coverage report's dishonesty surface.

**Round 1's outcome, so you do not have to re-read the report: it stopped a
genuinely damaging change.** My repoint trigger was "not ELIGIBLE chrome", which
matched three live rows it must not have — `idea.uk`'s section-level header and
footer, and **`leopardessconsulting.co.uk`'s own ACTIVE FORK**, which my code would
have replaced with the house header on the next render. Narrowed to `NOT is_active`
(RETIREMENT), which is exactly what `deactivated_site_components` detects.
Eligibility decides what may be CHOSEN as a default; retirement decides what may no
longer be SERVED. **Do not widen it back** — two tests and a source-level assertion
guard that, and `WRONG_CALLS` 2026-07-31 (late) carries the full account.

**The objection that may still come**, because the submission flags it as risk #1 itself:
`render_site_components` now REASSIGNS an assigned-but-ineligible slot where before it
only assigned an UNASSIGNED one. That widens what a shared action promises its
callers. If the architecture or guardian seat calls that architecture-scope, the
answer is **not** to resubmit with better measurements — it is to route it to
`architecture_review/` on its own merits (CLAUDE.md, "a veto on SCOPE is not answered
by resubmitting"). The precedent in its favour: the owner ruled today to do exactly
this repoint by hand, 21 times.

## What shipped, so you do not re-derive it

| what | where | state |
|---|---|---|
| One chrome predicate + `ResolveChromeComponent` + `ChromeSlotFunction` | `component_library.go` | LIVE v1.0.1219, pod-verified both replicas |
| Two assignment call sites routed through it | `render_site_components_action.go`, `link_site_components_action.go` | LIVE |
| `GetComponentByFunction` given `ORDER BY name` (answer measured unchanged) | `component_library.go` | LIVE |
| `repointRetiredChromeSlot` + build_status-aware idempotence exit | `render_site_components_action.go` | committed `39afbf697`, **narrowed at round 1 in `60fd06e68`**, inert until a roll |
| Tests incl. source-scanning lockstep + ordering assertion | `chrome_selection_test.go` | green, non-vacuity proven by induced fault |
| Concept register | **CLC-013** in `register/component-lifecycle.md` | includes the 166 extension |
| §9 pattern + §10 rows for 118/166/167 | `016b_debugging_guide_8_consolidated.md` | done |
| Landmine + its dated correction | `LANDMINES.md`, synced to `doc_notes` | done |

## The three residuals, in the order they will bite

1. **195 of 206 `page_rerender` items are still `triaged`** two hours after the chrome
   rebuilds created them; the oldest stuck one fleet-wide is from 13:59 UTC. **Stored
   chrome is correct on all 14 sites and the DEPLOYED pages still serve the old
   footer** — `curl relojistas.com | grep -o '<h4>[^<]*</h4>'` shows `Our Services`
   (old) until it shows `Explore` (new). That queue is `bugs_open/149`'s lane, not
   this one. Do not "fix" it here; do not read "28/28 slots active" as "the fleet
   looks right".
2. ~~**`bugs_open/167`**~~ — **PICKED UP AND FIXED BY ANOTHER LANE the same evening**
   (`8b29404d6`, `11f8b9e08`, closed in `306130ba3` → `bugs_closed/167`), which is the
   filing rule working exactly as intended: I scoped it out as an owner call and named
   it, and a lane that wanted it took it within hours. **They found a FOURTH chrome
   path I had missed and filed it as `bugs_open/170`** — the style-collection pin
   (`style_collections.header_component_id`), which applies no eligibility predicate at
   all and has three deployed sites pinned to a deactivated header. So the census in
   this lane's PLAN ("the question is asked in four places") was itself one short.
   **Read 170 before touching chrome again**, and note their closure says NOT LIVE.
3. **No active `head` component exists fleet-wide.** 13 head slots still point at
   deactivated components and `repointRetiredChromeSlot` correctly declines rather
   than churn them. Activating one changes every page's `<head>` (the build path falls
   through to `RenderFallbackHead` today), so it wants the same one-site-first
   treatment the footers got. Data call, not code.

## Traps this lane paid for — do not re-learn them

- **A COMPLETED chrome run that changed nothing is the normal failure.** Two distinct
  gates cause it and they look identical: `rerender-pages` needs
  `refresh_site_components: true` in `input_data`, and the `!force` exit skips any slot
  holding bytes regardless of whether the component changed. Read `site_components.updated_at`,
  not the orchestration status. (Fixed by `39afbf697`, so this trap expires at the roll.)
- **Do NOT clear `rendered_html` to force a render.** I did; I had no copy; it
  recovered only because the artefact regenerates from the template. `build_status='pending'`
  is the supported signal.
- **An `ORDER BY` added to an existing `LIMIT 1` is a behaviour change until measured.**
  RUNBOOK R3 is the query.
- **Do not build under `/tmp`** on this box — it is a 16G tmpfs other sessions fill,
  and a truncated `git archive` gives a 0-byte `go.mod` and an error that reads like a
  broken repo. Use `$HOME/.cache/`. RUNBOOK R8.
- **The working tree carries other lanes' mid-edit compile errors.** Verify against a
  clean `git archive HEAD` extraction, which is what `make build-*` uses anyway.
- **`git status` two tool calls before a commit is a guess.** Run it in the same
  command. I named three same-file passengers that were not in my commit because of this.

## Verify the 166 fix once it rolls

Pod-grep first, with a positive control in the same exec, on **both** replicas:

```sh
kubectl -n ai-persona-system exec <pod> -- sh -c '
  strings /app/agent-chassis | grep -c "repointed a slot off a RETIRED component"        # NEW, want >0
  strings /app/agent-chassis | grep -c "RenderSiteComponentsAction"                     # control, want >0'
```

Then the behavioural proof, which needs an induced fault because the fleet is now
clean: deactivate a spare footer component, point one site's footer slot at it, run a
chrome render, and confirm the slot comes back on `footer-theme-chrome` with
`build_status='rendered'`. Restore afterwards. `site_components_repoint_backup_20260731`
holds the pre-repoint mapping if you need to reason about what was where.
