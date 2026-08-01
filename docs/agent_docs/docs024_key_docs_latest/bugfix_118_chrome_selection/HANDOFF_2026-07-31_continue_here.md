# HANDOFF — bugfix_118_chrome_selection — continue here

**Written 2026-07-31 ~21:50 UTC; UPDATED 2026-08-01 08:15 UTC — the lane closed.**
Read this first, then `NOTES_chrome_selection.md`
(the technical log, newest at the bottom) and `SUMMARY_2026-07-31b_chrome_selection.md`
(current state, written to be read aloud).

## State in one paragraph — THIS LANE IS DONE

**Both bugs are CLOSED, LIVE and PROVEN.** `bugs_closed/118` (one chrome-eligibility
predicate) is live on `v1.0.1219`; `bugs_closed/166` (the repair that could never
repair) is live on `v1.0.1225`, council-APPROVED at round 2, and proven by an induced
fault on the failing branch. The fleet was repointed on the owner's ruling — 28 of 28
header/footer slots render from an active component. `bugs_closed/167` was taken and
fixed by another lane the same evening. **Nothing in this lane is owed to anyone.**

The one thing still standing between this work and a visitor is not ours: **195
`page_rerender` items have been stuck at `triaged` since 2026-07-31 19:25 UTC** (13+
hours), so stored chrome is correct everywhere and deployed pages still serve the old
footer. That is `bugs_open/149`'s lane. Do not read "28/28 slots active" as "the fleet
looks right".

## What is left, and none of it is this lane's

1. **`bugs_open/170`** — the style-collection PIN (`style_collections.header_component_id`)
   applies no eligibility predicate at all; three deployed sites are pinned to a
   deactivated header. Filed by the `bugfix_167_chrome_build_path` lane, who found the
   path this lane's census missed. **Owner call** (repointing is a visible markup
   change). ⚠ **`forked_from IS NULL` is RIGHT for pool selection and WRONG for a pin** —
   pinning a site to its own fork is the intended use, and copying the pool predicate
   makes the detector's first output a false positive against the only correctly
   configured site.
2. **No active `head` component exists fleet-wide.** 13 head slots point at deactivated
   ones; `repointRetiredChromeSlot` correctly declines rather than churn them, and logs
   at ERROR **with a full stacktrace** on every render of those sites. Activating one
   changes every page's `<head>`, so it wants the one-site-first treatment the footers
   got. **Data call, not code** — and it is what silences the stacktraces.
3. The stuck rerender queue above.

## What shipped, so you do not re-derive it

| what | where | state |
|---|---|---|
| One chrome predicate + `ResolveChromeComponent` + `ChromeSlotFunction` | `component_library.go` | LIVE v1.0.1219, pod-verified both replicas |
| Two assignment call sites routed through it | `render_site_components_action.go`, `link_site_components_action.go` | LIVE |
| `GetComponentByFunction` given `ORDER BY name` (answer measured unchanged) | `component_library.go` | LIVE |
| `repointRetiredChromeSlot` + build_status-aware idempotence exit | `render_site_components_action.go` | **LIVE v1.0.1225**, both replicas, induced-fault proven |
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
