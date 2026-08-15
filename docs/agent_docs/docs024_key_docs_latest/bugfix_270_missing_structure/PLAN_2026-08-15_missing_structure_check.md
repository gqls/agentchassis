# PLAN — bugfix 270: retype `missing_structure`'s predicate to the real chrome store

**Date:** 2026-08-15 · **Bug:** `bugs_open/270_HANDOFF_2026-08-13_missing_structure_check_fires_on_vestigial_columns_so_every_run_orders_a_full_site_rerender.md` · **Status:** plan written, no code changed yet.

All DB figures in this file were verified live on 2026-08-15 by the prior
verification pass and are cited `[MEASURED 2026-08-15]`. File/line citations were
re-read from the working tree today by this session.

---

## 1. What the bug is, in plain terms

`missing_structure` is one of ~30 discovery checks run per site by
`completeness-discovery-agent`. Its job is to notice a site whose pages went out
without their shared chrome (header, footer, `<head>`) and order a full-site
reassembly to repair it.

Its predicate (`platform/orchestration/actions/discovery_checks/check_missing_structure.go:92-101`)
reads `pages.rendered_header/rendered_footer/rendered_head` — three columns that
are empty on all 694 pages fleet-wide, because chrome actually lives in
`site_components` [MEASURED 2026-08-15; standing landmine, LANDMINES.md
"pages.rendered_header / rendered_footer / rendered_head are VESTIGIAL",
2026-08-03]. Its status filter `p.status IN ('active','deployed')` is also
non-discriminating: `pages.status` takes exactly two values fleet-wide, `active`
(658) and `archived` (36); `'deployed'` never occurs — that value lives in a
different column, `pages.build_status` [MEASURED 2026-08-15; second standing
landmine, same file].

So the predicate is true for every non-archived page, always. Every discovery
pass on every site with an active page files one `needs_rerender` work item
(`item_key = "missing_structure:rerender"`, severity `high`, priority 30,
`refresh_site_components: true` — a full reassembly) on a diagnosis that is false
on its face. Census: 50 items since 2026-04-24, ~31 completed full-site
reassemblies dispatched for nothing, still firing the day before this plan
(max `created_at` 2026-08-14 16:35) [MEASURED 2026-08-15]. The reason string
("Pages deployed without header/footer…") has already misled two other
investigations (`bugs_open/232`, `bugs_open/235`).

## 2. The decision: Candidate 1 — retype the predicate. Candidate 2 rejected.

The bug file offers two fixes. This plan weighs both and picks one, plainly.

### Candidate 1 — retype the predicate to `site_components` (CHOSEN)

Ask the store chrome actually lives in. A slot (`header`/`footer`/`head`) is
healthy when a `site_components` row exists for it with non-empty
`rendered_html`. Flag the site when any slot is unhealthy and the site has at
least one active page. Keep the check name, item type, item key, and dispatch
route exactly as they are. Add a `Resolved` entry on the healthy path so the
17 open false items close themselves through the framework's own RFC_010
machinery.

Why this wins:

- **It keeps a repair route that nothing else provides.** The check's dispatch
  arm (`refresh_site_components: true` → `rerender-pages`) is the one mechanism
  that orders `render_site_components` plus reassembly for a site whose chrome
  store is empty. `head_essentials_missing` (Candidate 2's replacement) detects
  the *symptom* at the served page, per page, but its work items are per-page
  `head_essentials_missing` items — it does not route the site-level repair.
  The two are complementary, not substitutes.
- **The stale items close themselves.** `CheckResult.Resolved`
  (`registry.go:112-142`) exists precisely for a check that no longer finds a
  defect it previously filed. The runner applies retractions in the same
  transaction as inserts, after them, and never on an errored run
  (`discovery_checks.go:266-283`) — so a blinded check cannot retract real
  defects. No hand-written cleanup SQL. This is "reuse existing machinery
  before building new", exactly.
- **Zero config change.** The check name stays `"missing_structure"`, so the
  live `checks` array in `agent_definitions` is untouched. That avoids the
  documented ordering hazard entirely: the runner hard-fails on a config-listed
  name the binary does not register (`check_site_unreachable.go:36-37`'s own
  header records holding migration 368 until the image rolled, for exactly this
  reason). On a tree where any session's roll ships your commit, a fix that
  needs *no* config/binary lockstep is worth a lot.
- **It can be false.** The retyped predicate returns zero findings on every
  currently-serving site (22 sites × 3 slots, all `rendered`, all non-empty
  [MEASURED 2026-08-15]) and fires only on a real absence. The current check
  cannot pass that first half — which is the whole defect.

Residual risk, stated rather than hidden: "zero `site_components` rows" is not an
unconditionally safe proxy for "this site has no chrome". On 2026-08-03,
`loanandmortgagecalculator.co.uk` served full chrome on 41 pages with zero
`site_components` rows, because an older pipeline had baked chrome into the
deployed artefacts. That state no longer exists anywhere (that site was
backfilled by `nav-updater`; every non-pool/non-system site now has 3/3 rendered
slots) [MEASURED 2026-08-15], but the class could return. If it does, the
retyped check files **one** item, the reassembly renders the components and
converges the site onto the current pipeline, and the check goes quiet — a
bounded, self-terminating false positive whose "repair" is the same
convergence `nav-updater` performed deliberately. That is an acceptable residual;
the unbounded always-fire behaviour it replaces is not. The check that is
genuinely immune to this class is the served-page probe
(`head_essentials_missing`), which is approved and will cover it when its owning
workstream switches it on — see §5.

### Candidate 2 — retire the check (REJECTED, with reasons)

Deregister `missing_structure` from the live config, delete the file, clean up
the open items by hand. The argument for it is real: `head_essentials_missing`
(`check_site_structural_validity.go:991-1043`) probes the actual served page over
HTTP for `<title>`, skip-link and `<footer>` — strictly more robust than any
DB-column proxy, immune to baked-vs-stored chrome and to stale bookkeeping.

Why it loses anyway:

1. **It cannot reuse RFC_010.** A retired check never runs again, so it can
   never emit a `Resolved` entry. The 17 open rows (14 `unresolved`, 1
   `detected`, 2 `deferred` — all outside `workItemClosedStatuses`,
   `work_items_common.go:83-89`, hence all still dispatchable) would need a
   one-off cleanup: a numbered migration under `docs/agent_docs/sql_for_agents/`
   (next free number after 411, with a `_ROLLBACK` counterpart — the live
   convention, verified by listing that directory) doing
   `UPDATE site_work_items SET status='cancelled', result = coalesce(result,'{}'::jsonb) || jsonb_build_object('resolved_by','bugfix_270','reason','check retired: predicate read vestigial columns','resolved_at',now()) WHERE item_key='missing_structure:rerender' AND status NOT IN ('complete','verified','rejected','wont_fix','cancelled')`.
   Buildable, but it is new one-off machinery where Candidate 1 needs none.
2. **It requires the config/binary lockstep Candidate 1 avoids.** Removing the
   name from the config array must not land before a binary that no longer
   registers it — and conversely deleting the Go file first makes the runner
   hard-fail on the still-configured name. Two artefacts held in order, on a
   shared-HEAD tree where "hold the deploy" is not available. This is a known
   hazard class here, not a hypothetical.
3. **It opens a coverage gap of indeterminate length.** `head_essentials_missing`
   is approved but deliberately NOT yet in the live checks config; switching it
   on is explicitly reserved to the `portfolio_positioning` workstream on its own
   timeline (its HANDOFF says so). Retiring `missing_structure` under 270 either
   forces that other workstream's decision or leaves the fleet with no detector
   at all for "chrome store never rendered" until someone else acts. A bug fix
   should not annex another lane's decision.
4. **Even switched on, the replacement lacks the repair route** (first bullet
   under Candidate 1).

Retirement remains the right *eventual* question once `head_essentials_missing`
is live and proven — but that is that workstream's call to make with a working,
honest `missing_structure` in hand, not a hole where it used to be.

### CORRECTION to the originating bug file (marked per the PLAN convention)

`bugs_open/270`'s Candidate 1 sketch proposed flagging "when every row is
`build_status='pending'`". **This plan deliberately drops the `build_status`
half.** `build_status='pending'` is used elsewhere in this codebase as "a
re-render has been scheduled", not "this slot has no content" —
`chrome_link_policy.go:127-137` sets `pending` as the supported force-re-render
signal precisely *because* `rendered_html` still holds valid, currently-serving
content while it is set (and its LANDMINES entry records that blanking the HTML
instead once nearly lost an artefact). Gating on `build_status` would therefore
manufacture a new false-positive class: every slot queued for a routine
re-render would read as "missing structure" and trigger the same spurious
full-site rerender this fix exists to stop. The predicate keys on
`rendered_html` content only.

## 3. Is this architecture-scope? No — and here is the measurement, not the assertion

What the rule is: a change needs architecture review when it changes what a
**shared mechanism guarantees**, not merely because it touches something shared
(owner ruling 2026-07-29 §1, narrowed further by RFC_022 for opt-in fields).

What this change is: a retype of one check's internal predicate, plus adoption
of an existing opt-in field. Check name, `ItemType` (`needs_rerender`),
`ItemKey` (`missing_structure:rerender`), severity, priority, handler and
dispatch route are all unchanged. The discovery framework's guarantees — checks
run per site, file work items, may positively retract their own — are untouched;
`Resolved` was shipped through RFC_010 with per-check opt-in as the intended
adoption path (`registry.go:107-111`: "the other 49 checks are unaffected until
each is edited deliberately").

The consumer enumeration (grep `missing_structure` across `*.go, *.sql, *.yaml,
*.py, *.sh, *.ts, *.js`, run 2026-08-15, this session):

- `check_missing_structure.go` — the file itself;
- `discovery_checks_registration_test.go:45` — asserts the binary registers the
  *name*; unchanged, stays green;
- `sql_for_agents/074, 033, 070` — historical seed migrations listing the *name*
  in checks arrays; the name is unchanged and seeds are history, not live state;
- `check_site_structural_validity.go` — prose cross-reference to bug 270;
- one historical doc SQL (`idea_uk_vm_site/.../p3_02_promote_rerender_item.sql`) —
  prose mention.

No consumer parses the check's findings shape or its reason string (the two
things this change rewrites). The only live coupling is the name and the item
key, both preserved. Conclusion: normal council-gate submission (it is
`platform/` code, so the gate applies), not architecture review.

## 4. The design, precisely

All edits in `check_missing_structure.go` only. The header comment is rewritten
to describe the real predicate and to point at `bugs_open/270` and the vestigial
landmine. `findPagesWithMissingStructure` and `missingStructureFinding` are
deleted.

### 4.1 The query

One row, four booleans, one round trip:

```sql
SELECT
  EXISTS (SELECT 1 FROM pages p
          WHERE p.site_id = $1 AND p.status = 'active')          AS has_active_pages,
  EXISTS (SELECT 1 FROM site_components sc
          WHERE sc.site_id = $1 AND sc.slot_name = 'header'
            AND coalesce(length(sc.rendered_html), 0) > 0)       AS header_ok,
  EXISTS (SELECT 1 FROM site_components sc
          WHERE sc.site_id = $1 AND sc.slot_name = 'footer'
            AND coalesce(length(sc.rendered_html), 0) > 0)       AS footer_ok,
  EXISTS (SELECT 1 FROM site_components sc
          WHERE sc.site_id = $1 AND sc.slot_name = 'head'
            AND coalesce(length(sc.rendered_html), 0) > 0)       AS head_ok
```

Decisions and their reasons:

- **`p.status = 'active'`, not `<> 'archived'`.** The live vocabulary is exactly
  `active`/`archived` [MEASURED 2026-08-15]; naming the value that exists is the
  spelling the `'deployed'` landmine teaches. If a third status ever appears,
  `= 'active'` errs in the quiet direction (fewer firings), which is right for a
  check whose action is a high-priority full-site rerender.
- **`coalesce(length(rendered_html),0) > 0`, never `build_status`.** See the
  CORRECTION in §2 — `pending` coexists with valid serving content.
- **`head` is now first-class.** The old query selected `missing_head` as
  information but filtered only on header/footer. Every serving site holds 3/3
  rendered slots [MEASURED 2026-08-15: 22 sites × 3, all `(rendered, non-empty)`],
  and a missing `head` is as structural as a missing footer, so all three gate
  equally. This adds no firing anywhere in the current fleet.
- **Site-level, not per-page.** Chrome is a site property and the check already
  dispatches one site-level item; the per-page listing was theatre on top of a
  broken predicate. The unhealthy *slots* are the finding.

### 4.2 The branches

- **All three slots healthy** → empty `Findings`, plus:

  ```go
  result.Resolved = append(result.Resolved, ResolvedFinding{
      ItemType: "needs_rerender",
      ItemKey:  "missing_structure:rerender",
      Reason:   "site_components healthy: header, footer and head all hold non-empty rendered_html (bugfix 270 — earlier items were filed by a predicate reading vestigial pages columns)",
  })
  ```

  Narrow `ItemKey` branch, not `AllOfType`: the check files under exactly one
  fixed key per site, so the narrow claim is the honest one, and `AllOfType`
  would close *other* producers' `needs_rerender` items — the destructive breadth
  the runner's validation exists to keep deliberate
  (`work_items_common.go:258-276`). The retraction fires regardless of the
  active-page count — chrome demonstrably exists, so the item's premise is
  positively refuted, which is RFC_010's required shape (retraction only on
  positive observation, never inferred from an empty result).

- **Any slot unhealthy AND the site has ≥1 active page** → one finding
  (`"check": "missing_structure"`, `"missing_slots": [...]`) and one work item —
  identical `WorkItemSpec` to today's (`SiteID/Source/Pipeline/ItemType/
  Severity: "high"/Priority: 30/HandlerAgent: "rerender-pages"/Status:
  "detected"/ItemKey: "missing_structure:rerender"/BatchID`) except the text:

  ```go
  Summary: fmt.Sprintf("Site chrome incomplete: %s missing or empty in site_components — full reassembly needed",
      strings.Join(missingSlots, ", ")),
  ```

  and `SpecJSON` keeps `"check"` and `"refresh_site_components": true` but
  replaces the misleading reason with one that names the real predicate:
  `"site_components rows for <slots> are absent or hold empty rendered_html — chrome cannot assemble until render_site_components runs"`.
  The old text ("Pages deployed without header/footer — likely built before
  site_components were rendered") has already misled two bug threads; the new
  text states the observation, not a guess at the history.

  **Keeping `ItemKey` unchanged is load-bearing**: the RFC_010 retraction
  matches on it, so a new key would orphan the 17 historical open rows forever.

- **Any slot unhealthy AND zero active pages** → log at Info, return empty
  result: no finding, no work item, no retraction. Nothing to reassemble, and
  we make no health claim we did not observe. (Belt-and-braces: `pool`-status
  sites — the only fleet population in this state, 2 of 18 with pages
  [MEASURED 2026-08-15] — never reach this check anyway; zero
  `missing_structure:rerender` items have ever been filed for a `pool` site, so
  the exclusion lives upstream of `discovery_checks.go` [MEASURED 2026-08-15,
  dispatch mechanism not further traced — out of scope].)

- **Query error** → `return nil, err`. The runner then skips `Resolved`
  entirely (`discovery_checks.go:266-269`), preserving RFC_010's safety
  property: a blinded check cannot retract.

### 4.3 What closes the stale items

Nothing manual. On each affected site's next discovery pass, the healthy branch
emits the retraction; `resolveWorkItems` flips every row with that
`item_type`/`item_key` whose status is outside `workItemClosedStatuses`
(`complete/verified/rejected/wont_fix/cancelled`) to `complete`, annotating
`result` with `resolved_by`/`reason`/`resolved_at`, in the same transaction,
never touching rows this batch raised. That sweeps all 17 open rows — the 14
`unresolved` and 1 `detected` the bug counts, and the 2 `deferred` as well,
since `deferred` is also outside the closed set (`work_items_common.go:83-89`).

## 5. Relations — including one finding that must NOT be folded into this fix

- **File `bugs_open/278` (next free number after 277, verified by listing) —
  `check_decision_guards.go` reads the same vestigial columns.** Out of scope
  for 270's fix; recorded here so it is not lost. Proposed problem statement for
  that filing:

  > `check_decision_guards.go`'s `storedPageAssemblySQL`
  > (`check_decision_guards.go:72-78`) builds "the page, as stored" by
  > concatenating `pages.rendered_header` and `pages.rendered_footer` — vestigial
  > columns, empty on all 694 pages fleet-wide [MEASURED 2026-08-15] — with
  > `page_components.rendered_html`. Every decision-guard verdict is therefore
  > reached against an assembly that silently omits chrome/nav content: a
  > `contains` guard asserting header/footer content would fire a false
  > violation, and a `not_contains` guard on chrome content would falsely pass.
  > No wrong verdict has yet been observed — only 5 decision-record rows exist
  > fleet-wide and none of their patterns reference chrome content (the closest,
  > `D-001-free-beside-paid`, asserts a page-body CTA link) [MEASURED
  > 2026-08-15] — but the defect is structural and silent, and
  > `storedPageAssemblySQL` is deliberately shared between the check and its
  > verifier, so any fix must retype both in lockstep (assemble chrome from
  > `site_components`, or explicitly redefine and document the assembly as
  > body-only). `bugs_open/232` (2026-08-09) already identified this exact
  > second caller of the vestigial columns and nobody followed up. Per the
  > 2026-07-31 owner ruling, the filing session should state its verification
  > substitution as 270's did — the mechanism here is seven lines of quoted SQL
  > plus two fleet-wide counts, all read or measured directly.

- After this fix lands, the vestigial-columns landmine's reader count drops to
  one (`check_decision_guards.go`, pointed at 278) — update the LANDMINES entry's
  pointer accordingly when closing 270.
- `bugs_open/232`, `bugs_open/235` — both had to reason around items this check
  filed; note the fix in their files when 270 closes.
- `check_site_structural_validity.go:991-1025` (`head_essentials_missing`) —
  already cross-references 270; no edit needed there. Its switch-on remains the
  `portfolio_positioning` workstream's own decision (§2, Candidate 2, reason 3).
- `bugfix 117` / `stale_chrome` churn investigations — this check was a standing
  unattributed rerender source; after the fix it stops confounding them.

## 6. Phasing

1. **File `bugs_open/278`** with the paragraph in §5 (separate task, separate
   commit, per the one-commit-per-task rule).
2. **Edit `check_missing_structure.go`** per §4: rewritten header comment, new
   query, three branches, deleted page-level plumbing.
3. **Add `check_missing_structure_test.go`** following the sibling sqlmock
   pattern (`check_componentless_pages_test.go`): the two predicate tests that
   matter (healthy → zero findings + one narrow `ResolvedFinding`; one empty
   slot → one finding naming the slot + one work item under the unchanged key),
   plus the no-active-pages branch (no item, no retraction) and the error path
   (`Run` returns the error, so the runner's no-retract-on-error guard engages).
   Assert the emitted SQL does not reference `pages.rendered_*` — the sibling
   file's rationale (a one-line predicate regression leaves happy-path tests
   green) applies here exactly.
4. **Council gate**: submit rationale + plan via `097_TRIGGER_council_review_v1.sh`
   before or alongside the commit (platform code; budget ~30 min end-to-end).
   Commit narrowly by pathspec; if the verdict has not landed, use
   `Council-Submitted: <corr>` and let the 098 report credit it on approval.
5. **Ship**: the Go change is inert until an image builds from committed HEAD and
   rolls (fleet releases are whole-fleet, owner-run). After the roll, verify at
   the artefact: read the service's `build provenance` stamp and check
   `git merge-base --is-ancestor <fix-commit> <stamp>` — per service, not per
   fleet.
6. **Watch one discovery rotation**, then run §7's fleet checks.
7. **Close out**: move `bugs_open/270` → `bugs_closed/` only when fixed AND
   live AND the stale items are closed; update the LANDMINES reader-count
   pointer (§5); append the outcome to this workstream's NOTES and
   `README_where_we_are`.

## 7. How to verify

**Unit (pre-ship).** The retyped predicate returns zero findings and one
retraction for a healthy fixture (3 slots, non-empty `rendered_html`) — the
current code cannot pass this half. It returns one finding and one work item,
under `ItemKey "missing_structure:rerender"`, for a fixture missing any one
slot — this is also the demand control proving the check still *can* fire. No
item and no retraction when the site has no active pages. Error from the query
propagates as an error.

**Fleet (post-roll, after one full discovery rotation).**

- The retractions are the positive proof the new code ran (better evidence than
  any absence): the 17 open rows flip to `complete` with
  `result->>'resolved_by'` set and the §4.2 reason —
  `SELECT status, count(*) FROM site_work_items WHERE item_key='missing_structure:rerender' GROUP BY 1`
  should show `detected`/`unresolved`/`deferred` at 0 and `complete` up by ~17.
- `max(created_at)` for that key stops advancing on serving sites (pre-fix
  baseline to beat: 2026-08-14 16:35 [MEASURED 2026-08-15]). One clean rotation
  is decisive here, not merely suggestive: the old predicate fired on *every*
  pass for *every* site with an active page, so ~22 sites passing quietly is ~22
  trials of a previously always-firing event.
- The zero has its demand control at unit level (the synthetic unhealthy
  fixture); at fleet level the 17 positive retraction annotations distinguish
  "check ran and found health" from "check silently stopped running".
