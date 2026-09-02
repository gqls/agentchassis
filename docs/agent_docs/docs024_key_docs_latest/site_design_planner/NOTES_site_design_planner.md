# NOTES — site-design-planner (running record, append-only, newest at bottom)

## 2026-09-02 — lane opened, survey done, no code touched yet

Session was renamed "site design planner" by the owner, with the instruction to
pick up the thread if one exists or take responsibility for it if not.

**Checked for an existing thread first.** `MEMORY_workstreams.md` has no entry
for it; no `docs024_key_docs_latest/site_design_planner*` directory existed;
`bugs_open/`/`bugs_closed/` grep for `site-design-planner`/`resolve_composition`/
`install_site_composition` returns only `bugs_open/113` (a real hit, but owned by
the `brochure_component_library` thread, not a dedicated site-design-planner
lane) and `bugs_closed/291` (a passing mention — 291 is about a different phantom
handler, `hitl-review`, hit incidentally on one of this mechanism's own work
items — see below). `who-owns.py` returns no match for `site-design-planner`,
`needs_composition`, or any of the three domains below. **Conclusion: no active
thread. Took responsibility per the owner's fallback instruction.**

**Read the concept register** (`design-composition.md`, DES-001 through DES-062,
freeze date 2026-07-13) to get the mechanism's shape and history rather than
guessing from code cold. Cross-checked live: `agent_definitions` row for
`type='site-design-planner'` exists, `status='active'`, single version.

**Read `bugs_open/113` in full** (huge file, 2026-07-27 → 2026-08-12, six
sessions' worth of corrections-on-corrections). It is the most complete recent
account of how this mechanism actually behaves under load, including two
self-corrected wrong attributions in the same session (both instructive — see
PLAN §1). Its tail names the fix that closed the "no re-resolve" platform gap
(`allow_reinstall`, per-request, chassis v1.0.1290) — checked live in RUNBOOK §3
is copied straight from that file's worked example.

**Queried live `site_work_items`** for anything in this mechanism's item types
still open. Found three (PLAN §2). Checked each one's actual current state
rather than trusting status alone:
- `adversecreditmortgage.co.uk`'s `needs_composition` looked like a live stuck
  item at first read (`unresolved`, empty spec/result). Read `load_work_item_actions.go`
  around the two-strike anti-churn logic and realised `unresolved` at
  `attempt_count=0` means "this key already failed twice recently", not "this
  attempt failed" — the row itself never ran. Then checked `site_specs` and
  found `resolved_composition` was written 2026-08-25, after this item's last
  update — the site got composed some other way and this row is stale. Then
  pulled the site's full work-item history and found ~230 other unresolved rows
  from an unrelated Anthropic billing outage on 2026-08-25/27. **This item is not
  a composition-mechanism defect** — recorded so nobody re-derives the same dead
  end.
- `loancalculator.co.uk`'s `needs_composition` is a deliberate 2026-08-12 park,
  and that domain already has its own active session (`loancalculator` in
  `ListAgents`). Not this thread's to un-park.
- `ai-agent-orchestration.com`'s `needs_new_layout_candidate` is the one
  genuinely open, in-scope question — see PLAN §3. **Not yet investigated
  further this session.**

**Cross-session coordination.** Mid-investigation, `bugs_open/427` messaged
asking whether this thread touches `build-site-planner`/`write_site_plan_action.go`/
`validate_site_plan` (they're starting on `bugs_open/428`, a build-site-planner
bug, and the name similarity worried them). Replied: no overlap, this thread is
the composition-resolution agent only. Worth recording because the two agent
names (`site-design-planner` vs `build-site-planner`) are close enough to
collide in a skim, and apparently already have, at least once, in a cross-session
check rather than a wasted edit.

**What's NOT done yet:** the one real open item (§3 in the PLAN) — whether
`ai-agent-orchestration.com` now has real classification tags and could resolve
to something other than `brochure-formal`. That's the natural next step if this
thread continues.
