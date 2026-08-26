# PLAN 2026-08-26 — a coverage cursor with a priority regrade set (`bugs_open/394`)

Design, phasing, decisions **and their reasons**. Corrections live here marked as corrections,
never silently edited away. Evidence is in `NOTES_render_audit_rotation_cursor.md`; commands in
`RUNBOOK_render_audit_rotation_cursor.md`.

---

## 1. The defect, in one paragraph

`request_render_audit` selects a site's live pages `ORDER BY COALESCE(nav_order,999), name` and
takes the first `max_pages`. It takes the **same prefix every run**, so pages past the cap are
not "audited less often" — they have never been audited and never will be. `bugs_closed/242`
made this LOUD (a durable `RENDER_AUDIT_TRUNCATED` row) and raised the cap 25→60 as a stated
mitigation. The signal has no reader and the mitigation has been outgrown.

## 2. The three decisions, and why

### D1 — Candidate 1 (cursor), not candidate 3 (raise the cap). **Decided on a measurement.**

`[MEASURED 2026-08-26]` webdesign.co.uk's tail is not a count, it is a **class**: sorted the way
the audit sorts, `nav_order` is `0..90` (6 nav pages), `100` (94 tools, alphabetical), `200`
(48 `tool-*-guide` pages), `201` (1). A cap of 60 cuts between `tool-head-architect` and
`tool-html-minifier`, so **all 45 remaining guide pages are unreachable at any cap below 98** —
on a site that grew 15 pages in two days. A constant cannot chase that.

And the premise of "raise the cap" is itself false: `max_pages` is **per-dispatch**, not
per-agent (the `5 of 26` row, §3 below). There is no single "the cap" to raise.

With a cursor, `max_pages` stops being a coverage cliff and becomes what it should always have
been: a per-run spend and latency dial.

⚠ Consistent with `bugs_closed/242` §5's explicit prohibition — *"do NOT sweep in random page
order to spread the misses"*. A keyset cursor is **deterministic rotation with a recorded
window**, not randomisation: the gap is stable and reportable within any run, and the union
converges.

### D2 — The window is a UNION, not a slice. **This is the decision the plan turned on.**

A plain cursor takes webdesign's per-page re-measurement latency to `3 days × ceil(146/60)` =
**~9 days**. `sql_for_agents/469_render_audit_rotation_three_day_window.sql` is an **owner
instruction of 2026-08-18** whose stated why is:

> The render audit is the only thing that **GRADES a contrast repair** … Its eligibility window
> is therefore **the confirmation latency of the whole repair loop**: a fix that shipped today
> could wait up to **SEVEN DAYS** to be graded.

So a plain cursor does not merely underperform that ruling — it **exceeds the condition the
owner ordered removed**. Shipping it as an accepted cost would trade away something already
decided, inside a commit whose purpose is coverage.

It dissolves instead of trading, because the protected population is not the site — it is the
pages awaiting a grade. `[MEASURED 2026-08-26]`, with the grader's own predicate:
**webdesign.co.uk has 3 open `contrast_failure` items across 3 paths.** Fleet maximum is
robot-hands.com at 17 paths, on a 37-page site that never truncates at cap 60.

```
window  =  (pages carrying an open contrast_failure)  ∪  (next N−k pages from the cursor)
```

Cost today: **3 of 60 slots (5%)**. Grading latency stays at 469's 3 days; the cursor covers
the rest.

> **⚠ CORRECTED 2026-08-26, before this plan was first committed.** My first census of that
> population used `status NOT IN ('complete','cancelled','rejected')`, which is **not** the
> platform's closed set — `workItemClosedStatuses` is
> `{complete, verified, rejected, wont_fix, cancelled}`, and it deliberately excludes
> `unresolved`/`failed` (RFC_010, owner ruling 2026-08-02 "Decision 2: `unresolved` is OPEN").
> The number was unchanged on webdesign (3 and 3) **by site-specific coincidence, not by
> rigour** — it carries no `verified`/`wont_fix` contrast rows. The predicate is now derived
> from `work_item_retraction.go:118-128`, the grader's own candidate query. `WRONG_CALLS.md`,
> 2026-08-26.

### D3 — The commissioned reader is STILL OWED, and its alarm condition changes

The owner commissioned *a reader for `RENDER_AUDIT_TRUNCATED`* (decision 4, 2026-08-25); the
registry entry says "upgraded when it ships". A cursor does not discharge it:

1. The cursor ships **opt-in on the rotation caller only**, so `design-critique-agent` keeps
   writing prefix-mode rows by design — rows somebody must read.
2. The cursor **changes what the row means**, from "these pages are never audited" to "this run
   covered this window". A row meaning "healthy pagination" needs a reader precisely so the
   *unhealthy* variants become findings rather than folklore.
3. `DBG-075` requires the registry flip to `consumed` with `reader`/`reader_sink` **in the
   shipping commit**, and the checker OPENS the named file and requires it to contain both the
   code and the sink — a claim that cannot be satisfied by typing.

New alarm conditions: **(a)** a `coverage_mode` absent-or-`"prefix"` row from
`render-audit-agent` → RED (the rotation caller must rotate; a prefix row means the config flip
regressed or the binary is old); **(b)** consecutive cursor rows for one `(domain, agent_type)`
with a repeating `window_first` → RED (stalled cursor); **(c)** prefix rows from
`design-critique-agent` → acked at birth with a written reason, so the baseline is quiet and a
NEW unacked caller pages. Vacuity guard (DBG-077): refuse to report over an empty
`agent_error_log` — "the check could not run" must never read as "the check passed".

## 3. Facts the design rests on, all `[MEASURED 2026-08-26]`

| fact | value | why it constrains the design |
|---|---|---|
| webdesign live pages / audited / tail | 146 / 60 / **86** | the bug said 131/71 two days earlier |
| tail composition | all 45 `tool-*-guide` at `nav_order` 200 | kills candidate 3 |
| callers | `render-audit-agent` (cap 60), `design-critique-agent` (cap 8) | the cursor must be per-caller |
| the `5 of 26` row | `{"max_pages":5}`, `render-audit-agent` | **the cap is per-dispatch** |
| sites truncating at cap 8 | **25** | design-critique is the fleet-wide half |
| open `contrast_failure`, webdesign | 3 items / 3 paths | the priority set is affordable |
| retraction scope | `Summary.PagesAudited`; *"AllOfType is never set, and must not be"* | a cursor cannot silently close skipped pages |
| durable per-page audited list | **does not exist** — only counts | the bug's own acceptance is unrunnable today |
| optional-key budget | 6 of N=10, 0 actions over budget | one new key is affordable |
| `design-critique-agent` cadence | **manual only**, no scheduler | its truncation is a sample, not an accruing debt |

## 4. The change

**One new optional config key, `rotate_coverage` (bool, default `false`).** Default-OFF is the
owner's 2026-08-02 §2 rule for new authority on a shared seam — and per RFC_022's 2026-08-11
narrowing this is NOT architecture-scope (opt-in; the unsafe side is the default; the live
consumers are enumerated, not asserted: two, both listed above). Budget goes 6 → 7 of 10.

**Persistence: a NEW table `render_audit_page_cursor`,** not a column on
`site_discovery_rotation`. Three reasons, any one sufficient: that table's own `COMMENT`
declares it *"Written by the site-discovery-rotation-* pre_queries"* and it has never had a Go
writer; its `last_selected_at` is `NOT NULL` with ruled semantics that a cursor UPSERT would
have to invent a value for; and its `agent_type` means "a scheduled task's target", which
coincides with "the agent that dispatched with rotation on" today and is not guaranteed to.

Keyset cursor on `(COALESCE(nav_order,999), name)` — **the existing ORDER BY**, so the cursor
and the ordering cannot disagree. Key `(site_id, agent_type)`. Boundary found by exact tuple
match first, falling back to first-strictly-greater, so a **deleted or renamed cursor page
neither stalls the rotation nor restarts it from the top**. Final window clears the row; a
cursor past the end restarts this run and must never return an empty window (an empty `urls`
upstream reads as the `no_deployed_pages` skip — a false skip).

**The cursor never stores the cap.** A one-off override of 5 takes a 5-page window and advances
by 5. Any design storing the cap would corrupt on override; this one cannot.

**Advance at DISPATCH, written AFTER a successful produce.** Dispatch-advance is the estate's
existing ruling for this rotation family (migration 346: *"a site whose run fails must not pin
the rotation head and starve the fleet"*), and response-advance would have to live in a
different action entirely, since an awaiting step's result never survives the park. The failure
modes are not symmetric: dispatch-advance skips one window for one cycle, visibly;
response-advance retries a wedging page **for ever** and starves everything behind it. Written
after the produce — deliberately the opposite ordering from the truncation row, which must land
*before* the send so a failed dispatch cannot unrecord it. A failed cursor write is non-fatal
but LOUD, naming the window that will repeat.

**Priority set**: open `contrast_failure` rows only — the one type this reply grades
(`write_render_audit_findings_action.go:791`). Bounded at `max_pages/2` so the rotation always
keeps at least half the window and coverage can never stall; selection is cyclic-from-cursor so
the dropped excess is not the same pages every run (the deterministic-prefix disease reappearing
in miniature). Priority pages are intersected with the live set, never advance the cursor, and
are de-duplicated at fill time.

**Context keys** (the reader keys on these, never on prose): `coverage_mode`, `window_first`,
`window_last`, `cursor_cleared`, `audited_paths`, `priority_paths`, `priority_open_items`,
`priority_dropped`, `priority_not_live` — all present on every cursor-mode row, **zeros and
empty arrays included**, because a key that appears only on a bad run makes its absence
ambiguous between "none" and "binary too old to count".

**Message**: mode-split. Prefix mode keeps today's text — *"the unaudited tail is the SAME pages
every run"* remains literally true for that mode. Cursor mode gets new text. Removing the old
sentence fleet-wide would be the stale-assertion error in the other direction.

## 5. Phasing

1. Council submission (this plan).
2. Code + tests; `go test` green; existing truncation tests pass **unchanged** — that is itself
   the compatibility guard.
3. Commit → `make build-agent-chassis` → bump `IMAGE_TAG` → roll → verify at the binary with a
   positive control.
4. Apply `646_render_audit_coverage_cursor_HOLD.sql` (table + config flip in one transaction,
   `DO`/`RAISE` verify — a bare `SELECT` verify cannot stop the `COMMIT`).
5. Registry flip to `consumed` in the shipping commit; reader lands with or before it.

**What commutes:** every skew order is inert, never wrong (old binary + new config = unknown key
ignored, since the spec sets neither `CheckConfig` nor `StrictConfig`; new binary + old config =
default false = today). So migration-last is **convention plus verifiability**, not a hard
safety constraint — applying it early would make a post-roll prefix row ambiguous between
"binary old" and "code broken". Say that in the migration header rather than claiming a false
hard dependency.

## 6. What is NOT in scope

Adapter echo of the window into `findings_written` (needs a browser-runner roll; the
`agent_error_log` route already satisfies the acceptance) · an owner ruling on whether
design-critique should rotate or curate · `page_names`, declared in the action's spec and
**never read** — a separate defect worth its own filing · webdesign's own page growth · the
`agent_error_log` retention ambiguity (30 d under migration 466 vs 365 d under 567), which
whoever grades the acceptance must resolve before promising a union window.
