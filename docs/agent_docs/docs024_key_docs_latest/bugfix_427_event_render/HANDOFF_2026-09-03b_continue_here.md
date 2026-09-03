# HANDOFF — bugs_open/427 + 428 + 454, continue here (supersedes HANDOFF_2026-09-03)

Written 2026-09-03, late morning. Continues from `HANDOFF_2026-09-03_continue_here.md` in this
directory, which is now **superseded in its central claim**: the defect that handoff called
"undiagnosed, needs a fresh diagnosis pass" is diagnosed, fixed and council-approved. Read this
one; read that one only for the arc from filing to yesterday.

**The live record is the bug files themselves:**
- `bugs_open/454_HANDOFF_2026-09-03_the_light_rerender_computes_a_section_plan_and_drops_it_so_every_page_is_rendered_from_its_own_stored_data.md` — the new case. §11 is the verdict.
- `bugs_open/427_HANDOFF_2026-09-02_...md` — §14 and §15 are today.
- `bugs_open/428_HANDOFF_2026-09-02_...md` — §11 plus the CONTRIB and its addendum at the foot.
- Milestone read-out: `SUMMARY_2026-09-03_bugfix_427_event_render.md` (this directory).
- Narrative + evidence: `NOTES_bugfix_427_event_render.md` (three entries dated today) and
  `README_where_we_are.md` (plain prose, for the owner).

## 0. One paragraph on each, current as of this update

**454 (NEW)** — `classifyStoredSection` in `rerender_page_sections_action.go` computes a
section plan, branches on it, and returns without setting `c.plan`. The field is **read at
exactly one line in the repository and written at none**. So since `94f81cc60` rolled
(2026-09-02 12:27 BST) every light re-render has composed `base ⊕ stored content_data` with
`plan.ResolvedData == nil` and persisted the stored map unchanged — the estate's repair vehicle
delivering nothing while reporting success. Fix `9831e9ab4` (one line + a two-level regression
test), council `075cfedd` **APPROVED** round 1. **OPEN because it is INERT until a chassis
image rolls.**

**427** — Everything this lane built was correct; it was feeding a pipeline that had stopped
delivering. The page is live and shows the honest empty state. It fills in on its own once 454
ships. Council `ff91e666` **round 2 submitted** (dispatch `c46cf6c2`) — verdict still pending
at the time of writing, the only `council_report` on that correlation remains the 09-02 REVISE.
A defect in this lane's own migration 719 was found and fixed as migration `727`.

**428** — Unchanged in its own substance. Its §4 open item (audit what migration 687 actually
produces) was answered by the `gamedesign.uk` lane and is now ticked; the residual it exposed —
a compliant reason can be a false one — is adjudicated as a NEW residual of 687, not a
reopening. Its one original open item (a human uses the release surface) is still open and is a
decision, not a code task.

## 1. Decisions/actions that need a person, not a session

1. **A chassis roll.** Everything 427 and 454 still owe is gated on it. Releases are
   whole-fleet and the owner runs `make release`; nothing here should attempt a one-service
   apply.
2. **A human uses bug 428's release surface** on a real flagged verdict — the tool has been
   live end to end since yesterday and nobody has clicked it. Worked case ready to hand:
   boxingonline's own `e3c2b440-c006-40ec-be7a-88d0b689ed1e`.
3. **Why did `blog-content-planner` stop running on 2026-04-24?** **Now filed as
   `bugs_open/460`** (2026-09-03, `787283cc9`) — still unowned, no root cause asserted. It
   decides whether the 687 residual is an outage or a truth-telling defect. Since this handoff
   was first written the producer has been shown **DRIVEN, not merely wired**: 14
   `needs_blog_posts` items (13 complete), 2026-03-14 → 2026-04-24, agreeing to the day with
   the `llm_call_log` window — so the planner's citation is STALE, not invented.
   `[NOT ESTABLISHED]` why it stopped; do not assert.
4. **Same open decisions as the previous handoff, still open**: whether `news_feed_ingestion`
   extraction should run beyond boxingonline.com; arming `event_fixture_completeness` in a live
   `run_checks.config.checks` array.

## 2. The first thing to do when a chassis carrying `9831e9ab4` rolls

```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor 9831e9ab4 <the stamp>     # exit 0 = the fix is in that binary
```
The provenance line is a STARTUP line and scrolls; an empty result means "not in range", not
"unstamped" — fall back to the binary probe with a known-absent sha as a control. Then
re-dispatch the page-rerender (recipe in the RUNBOOK) and read the **artefact**:

```sql
SELECT pc.content_data->'items' AS items, length(pc.rendered_html)
FROM page_components pc
WHERE pc.page_id='4b74ff1f-455a-4bb2-b81d-e1d0ec824f33' AND pc.slot_name='event-list';
```
`items` non-empty, `rendered_html` no longer 1,813 bytes, then curl the served page.

⚠ **Do NOT read `escalated`/`rerendered` as evidence that data moved.** 454 is exactly the
proof they cannot tell you: every count was healthy for a fortnight while nothing was
delivered.

## 3. What was verified this session (don't re-derive)

- **454's mechanism**, by grep and then by a failing test, not by inference: `grep -n '\.plan\b'`
  on the action file returns ONE hit, a read, with no assignment anywhere in the package.
- **Population `[MEASURED 2026-09-03]`**, `build_status='deployed'` joined to the component:
  **206** rows / **196** pages / **21** functions declare a `query.*` source; **1,855** /
  **838** / **82** declare any non-`llm` source. Query in `bugs_open/454` §4. Re-run before
  quoting — a census goes stale by addition.
- **HEAD is green** at `13aac933f` (`verify-head-builds.sh --test ./platform/orchestration/actions/`,
  "ok … 7.215s"), with the fix at line 1482 and both new tests part of that green.
- **The `event-list` row's own facts**, answering the round-1 council objections at the row:
  `input_schema` top-level keys exactly `{notes, fields}` (native dialect, so
  `chk_input_schema_no_legacy_dialect` is satisfied); `component_level='section'`, deliberately
  not `'tool'`; `query.*`-sourced components **33 total / 30 active**.
- **`pages.sections` is order-bearing by INDEX** and 719 had lost that order; `727` restored it
  and the live array now indexes exactly onto `hero-tool@1, event-list@2`.
- **`create_blog_posts_action.go:238` creates `blog-post` pages without the planner**, routed to
  by `discovery_checks/check_empty_blog.go` and handled by a live `blog-content-planner` row.
  **Driven 14 times, then stopped dead**: items 2026-03-14 → 2026-04-24, `llm_call_log` 10 calls
  over the same window and none since. Both re-verified first-hand. Now `bugs_open/460`.
- **Migration 730's rule 20 is live in the reworded form** (`updated_at 11:03:58Z`), naming the
  dormant mechanism rather than asserting no pass exists.

## 4. Three traps this session hit, so the next one doesn't

- **Two rolling windows that both read as "this never happened".** `orchestration_states`
  spans ~24 HOURS (`[MEASURED 2026-09-03]` 9,155 rows, 2026-09-02 10:41Z → 2026-09-03 10:58Z),
  so a zero there cannot support an all-history claim — use `llm_call_log`, which keeps replies
  verbatim as the training corpus. And **`site_work_items` is not the population**: a
  `needs_blog_posts` census over it returns **ZERO fleet-wide across every status** while all
  14 rows sit in `site_work_items_archive`. Closing a row archives it out of the table you
  queried, so a *successful* mechanism erases its own evidence from the live table. Query both.
- **A pathspec commit takes a same-file passenger.** This session's one-line fix carried the
  `bugs_open/450` lane's uncommitted rework of the same file into HEAD and broke the build.
  Their closure landed as `587666be8`. Cheap warning sign: `git status --porcelain <the file>`
  showing it already dirty when your own edit is one line.
- **Reviewers judge the SKETCH.** The one advisory on an otherwise clean APPROVED verdict was
  drawn by sketching one test where two had been written. Four lines would have avoided it.

## 5. Named and deliberately NOT done

- **No fleet census** of pages whose `pages.sections` order no longer indexes onto their own
  `page_components` positions. `727` fixed ONE page.
- **`"advertising"`** remains declared in that page's array with no `page_components` row.
  Pre-dates 719 and this lane; left exactly as found.
- **Migration 719's header is unedited**, including its now-refuted paragraph about the items
  defect. It is an applied migration and the runner's drift guard hashes it; the correction
  lives in `bugs_open/427` §14 and `bugs_open/454` §5 instead.
- **No `090` diagnosis run on 454.** Substituted for, with the reason stated in `bugs_open/454`
  §7 per the 2026-07-31 owner ruling — not silently omitted.
