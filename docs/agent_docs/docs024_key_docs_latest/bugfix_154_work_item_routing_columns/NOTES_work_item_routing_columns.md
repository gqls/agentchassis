# NOTES — `bugs_open/154`, work-item routing columns

Append-only, newest at the bottom. Technical log: evidence, commands, what the
system actually said, and every misstep.

---

## 2026-07-31 — session "bugfix 19", picking the bug

Asked to take the next `bugs_open/` case nobody else is on. The ownership check
is the part worth recording, because the naive version of it is wrong.

`git log` over `bugs_open/` is **lagging** — it cannot see a session mid-fix. So
I read the live `.jsonl` transcripts instead, tallying `bugs_open/NNN` mentions
in the tail of every session touched today. That produced a clear map of what 31
concurrent sessions are on (149, 157, 138, 161, 151, 156, 155, 099, 072, 139,
135, 153, 143, 125, 083, 018, 120, 017 all dominantly owned).

Then, per the memory note, I re-checked the *shortlist* by **code symbol rather
than by number** — `tool-improver|load_tool|improve_tool|tool-auditor`. That
turned up two sessions with high counts:

- `c4daed6f` (158 hits) — last written 10:05, i.e. ended. This is the `131`/`083`
  lane that *filed* 154.
- `631baa00` (15 hits) — **active**, 16:10. Worth a second look, and the second
  look mattered: it says *"I deliberately did NOT fire a cluster acceptance run …
  On a failing verdict the judge inserts an `improve_tool` work item routed to
  `handler_agent='tool-improver'` — an automated fixer, pointed at a page whose
  … [owner] … isn't mine to choose."* That is a session **declining** to touch
  tool-improver, which is the opposite of owning the fix.

**Lesson, and it is the one the memory note already makes:** a high symbol count
is not ownership. It can be a session explaining why it is staying away. Read
the surrounding text, not the tally.

## Confirming the cause rather than inheriting it

154 marked its own explanation `[INFERRED — not yet read in the code]` and named
the two things to read first. Both reads, live (`agent_definitions`, not seeds):

`build-dispatch-loop`, `process_item.sub_workflow.call_handler.input_mapping`:

```json
"component_id?": "current_item.spec.component_id"
```

`tool-improver`, `load_tool`:

```json
{"action": "query_database",
 "config": {"params": ["input_data.component_id"],
            "query": "SELECT ... FROM content_components cc JOIN page_components pc ON pc.component_id = cc.id JOIN pages p ON pc.page_id = p.id WHERE cc.id = $1::uuid AND cc.is_active = true LIMIT 1"}}
```

Then `LoadWorkItemsAction`'s SELECT — and there it is:

```sql
wi.severity, wi.summary, wi.spec, wi.page_id,
```

`page_id` and **not** `component_id`, though `\d site_work_items` shows
`component_id | uuid` as a real column alongside `entity_id` and `affected_url`.
So `current_item` never carries them, and `current_item.spec.component_id` is the
only reachable path. `ResolveInputMapping` skips an unresolved `?` path silently
(`input_mapping.go:122-129`), so the value dies with no log line and surfaces two
agents later as a nil query param.

**The framework, not either agent, is what drops the value.** The creator that
used the schema properly is the one whose items cannot be dispatched.

## Census, 2026-07-31 (all `site_work_items`)

| column | column set | column set AND no spec key | spec only |
|---|---|---|---|
| `component_id` | 21 | **16** | 235 |
| `entity_id` | 0 | 0 | 12 |
| `affected_url` | **0** | 0 | **0** |
| `page_id` | 1681 | 100 | **218** |

`improve_tool` alone: **4 of 4** `tool-auditor` rows are column-only, all
`failed`/`wont_fix`; **16 of 16** rows from `tool-acceptance-agent`,
`design-discovery-agent` and `generic` are spec-only and fine.

**Still live:** the newest row is `a5d11c86`, robot-hands.com, dated
**2026-07-31** — a fresh failure one day after filing. The bug is not stale.

## The check that could have made the whole fix inert, and nearly went unrun

`load_tool` queries `content_components`, but `page_components.id` is a
*different* id — and 016b already records "do not key re-validation on a stored
`component_id`" because `page_components.id` is not stable across re-renders. If
`tool-auditor` had been writing a `page_components.id` into the column, then
delivering it would satisfy the mapping and **still** find no row, and I would
have shipped a fix that changed the error message and nothing else.

```
item     | in_content_components | active | in_page_components | joins_to_a_page
a5d11c86 | t                     | t      | f                  | t
ee745694 | t                     | t      | f                  | t
7c2d898a | t                     | t      | f                  | t
5b4fd5cc | t                     | t      | f                  | t
```

4/4 satisfy **every clause** of `load_tool`'s query. The fix is not inert.
`[VERIFIED]` — query above, run 2026-07-31.

## The design candidate I rejected, and what killed it

The obvious fix is to **backfill the column value into the item's `spec` map**:
it needs no config change at all, because `current_item.spec.component_id` would
simply start resolving. I had written most of the argument for it before
checking who else reads `spec.component_id`.

`rerender-pages` does. And `create_rerender_items_action.go:219`:

```go
scoped := (reason == "section_data_resolved" || reason == "image_landed") && componentIDStr != ""
```

So a write into `spec` can flip a **site-wide rerender into a component-scoped
one** — in a path with nothing to do with this bug. Rejected on that.

**Misstep worth recording:** my first instinct was to reassure myself with the
data — the two live `needs_rerender` rows in that shape both carry an empty
`reason`, so `scoped` would be false today anyway. That is a *data-dependent*
safety argument about a *code* property, and it goes stale the moment a rerender
is raised with one of those two reasons. The design should not rest on it, and
the note in the code says so. I kept the structural argument and threw the
comforting measurement away.

## What I deliberately did NOT change

- **`page_id` keeps its column-only behaviour.** Extending the spec fallback to
  it "for symmetry" would newly expose `current_item.page_id` on **218** rows.
  Widening what reaches a handler changes it without editing it. There is no bug
  here to close — `page_id` is already exposed — so there is no reason to.
- **The key stays ABSENT, not `""`,** when neither source has a value. An
  optional mapping path that *resolves* is forwarded; one that is *missing* is
  skipped. Materialising `""` would turn "not supplied" into "supplied as empty"
  for handlers that gate on presence — `create_rerender_items` again.
- **`tool-auditor`'s site-scoped `item_key`** (`audit_fix_<domain>`, one key per
  site on a per-tool fix) — 154's second finding. A separate defect in item
  *creation* with fleet-wide dedup consequences, marked `[UNVERIFIED]` in the bug
  file as possibly intended. Left open and recorded, not silently dropped.

## Reuse, caught by the compiler

Wrote a `uuidPtrString` helper; `go build` refused it —
`fork_theme_from_site_action.go:450` already had a byte-identical one. Deleted
mine. Small, but it is the CLAUDE.md rule ("reuse existing machinery before
building new") being enforced by the toolchain rather than by me remembering.

## Proving the test can fail

A passing test proves nothing until you have watched it fail. Deleted the
`component_id` exposure line and re-ran:

```
--- FAIL: TestLoadWorkItems_ExposesRoutingColumns
    column-only item: component_id = <nil>, want e56f1d80-… — this is the bugs_open/154 failure
    spec-only item:   component_id = <nil>, want c5bb011e-… — the fallback regressed the 235-row majority
```

Both arms fire — the bug itself **and** a regression of the majority path.
Restored, full `./platform/orchestration/...` suite green.

## Filed and submitted

- Diagnosis loop (`090`), per CLAUDE.md's default for a durable structural
  claim: `RUN_CORRELATION_ID=21758756-d7b3-444a-844e-b37e09b5c9ce`.
  **Retention trap, exactly as 154 warned:** the `orchestration_states` rows for
  it were already gone ~25 minutes later. Track it via
  `site_work_items` (`7d5d34eb`) and `diagnosis_artifacts` instead.
- Council gate: `SUBMISSION_CORR=10be5ed9-3bd0-45ed-b6bb-4385a887967d`.
  Committing before the verdict, so the commit carries `Council-Submitted:`,
  never `Council-Reviewed:` on a verdict I have not read.
