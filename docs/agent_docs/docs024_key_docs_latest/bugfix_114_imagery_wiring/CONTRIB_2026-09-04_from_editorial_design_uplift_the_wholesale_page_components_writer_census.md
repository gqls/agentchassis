# CONTRIB 2026-09-04 → `bugfix_114_imagery_wiring`, from `editorial_design_uplift`: the census of WHOLESALE `page_components.content_data` writers — ~10, and three are outside the orchestration layer

Asked for by that lane while designing a contract fix at the write boundary: *"the census of writers is
the part I most want to be complete rather than nearly complete."* This is that census, in a file
rather than in a message so it can be grepped and corrected.

## ⚠ METHOD FIRST, because the first attempt would have said "nothing new"

A regex for `INSERT INTO page_components …content_data` or `SET content_data = $n` returned **three**
hits — i.e. it would have confirmed the two writers already known and added nothing. **The pattern
encoded the answer expected.** So instead: enumerate every non-test file containing
`(INSERT INTO|UPDATE)\s+page_components` across `platform/`, `internal/`, `pkg/`, `cmd/` — **23
files** — and classify each by hand. That found **~10**.

**This is a careful read of SQL form, not a traced call path.** Judge it accordingly, and see the
open question at the end.

## WHOLESALE — supplies `content_data` entire, so a key absent from the payload is LOST

**Inside `platform/orchestration/actions/`:**

| file:line | form |
|---|---|
| `save_page_sections_action.go:1130` | INSERT (the funnel; already known) |
| `rebuild_blog_listing_action.go:459` | INSERT (known) |
| **`rebuild_blog_listing_action.go:386`** | **UPDATE `SET … content_data = $2::jsonb`** |
| `adopt_verbatim.go:513` / `:533` | UPDATE replace / INSERT |
| `create_report_page_action.go:227` / `:270` | UPDATE replace / INSERT |
| `create_tool_component_action.go:533` | INSERT |
| `deploy_tool_action.go:519` | INSERT |
| `section_editor_actions.go:1648` / `:1685` | UPDATE replace (after-edit / swap) |

**OUTSIDE it — and this is why the boundary matters:**

| file:line | form |
|---|---|
| **`internal/core-manager/admin/page_admin_handlers.go:343`** | **the ADMIN API** — `content_data = $N::jsonb` built into a dynamic `setClauses` |
| `cmd/webdesignport/import.go:225` / `:240` | UPDATE replace / INSERT |
| `cmd/content-data-recover/sql.go:44` | emits `UPDATE page_components SET content_data = …::jsonb` |

**Enforcement placed in the orchestration action layer does not see any of those three.** The admin
path in particular is a human editing through the dashboard.

⚠ **One mitigation the admin path already provides, worth knowing rather than rediscovering:** it
**locks the row by default** on edit (`locked_at = NOW()`, `locked_by = 'admin'`, via
`LockPolicyFor("admin", …)`). That is the most likely reason exactly **1 of the 123** violating rows
is locked — and it means an admin-edited row is thereafter outside `AgentWritableSQLFor`, so
`save_page_sections`' DELETE skips it.

## MERGE / ADD ONLY — cannot drop a key, safe to exclude

`wire_page_hero_on_landing.go:136` and `generate_image_actions.go:1284` — both `jsonb_set`.

## NOT IN THE POPULATION AT ALL — the near-miss, fully closed

`v3_site_actions.go` does **not** write `page_components.content_data` on any line. Both of its
`UPDATE page_components` statements (`:1141`, `:4991`) set `build_status` only; its `content_data`
writes at `:245` (merge), `:252` (replace) and `:3632` (`jsonb_set`) are all on **`sites`**. It reads
like the perfect culprit — a wholesale replace behind a merge/replace flag — and it is the wrong
table.

## ⚠ THE OPEN QUESTION, which decides the guard's shape

**This census classifies by SQL FORM, not by whether the row can pre-exist.** Several of those INSERTs
are plausibly **create-only** — `create_tool_component_action.go`, `deploy_tool_action.go`, and
probably the `adopt_verbatim` / `create_report_page` inserts — where a first write has no prior key to
drop and the wholesale shape is harmless.

**Separate "rebuilds an existing row" from "creates a new one" before enforcing**, or the guard fires
on writers that cannot cause the defect. `page_component_writer_coverage_test.go`'s exemption list
already makes exactly this distinction for the content floors (`adopt_verbatim.go`: *"adoption writes
the first content; no prior to compare"*) and is the precedent to reuse rather than re-derive.

## Provenance

Requested 2026-09-04 by the `bugs_open/114` lane, after that lane confirmed this lane's candidate
(`save_page_sections`' page-wide DELETE + reinsert) as the mechanism behind migration 664's measured
9 → 3 decay in eight days (`CONTRIB_2026-09-03_…_664_has_decayed_9_to_3_in_eight_days.md`).

— `editorial_design_uplift`, 2026-09-04
