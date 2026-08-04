# NOTES — `bugs_open/087`, the self-planning writer

Append-only, newest at the bottom. Missteps are not an appendix — they are the point.

---

## 2026-08-04 ~10:15 — picking the bug, and why not 192

The session was opened as "bugfix 192 then 087". **192 was already owned** and I did not
take it. `who-owns.py` alone would not have told me — it reads commits, so a session
mid-fix is invisible — so I swept the live session transcripts as well:

```bash
for f in ~/.claude/projects/-home-ant-projects-agentchassis/*.jsonl; do …
  grep -c 'bugs_open/192\|bugfix_192' "$f"; done
```

Session `424d6591` had 53 hits and a modification time of *this minute*; it had committed
`2b9d84072 fix(192)` ten minutes earlier. Its first user message is the same
pick-a-bug prompt as mine. Competing would have been the `bugs_open/023` mistake again.

I also mapped every live session to the bugs it was actually touching (tool-use inputs and
assistant text, not just any mention) so I could pick something genuinely free. ~22 lanes
were live; most recent numbers were claimed. `087` showed only cross-references — session
`e37fd473` had opened it purely to check whether the failure it was seeing was already
filed, decided it was not, and filed `192` instead.

## 2026-08-04 ~10:20 — 087 is still valid, but not for the reason its file gives

The file's own account: candidate A applied 2026-07-27 (mig 246), verified live 07-28,
left OPEN only because its acceptance bar could not be met — the target was
`rebuild_policy=owned`, and then `bugs_open/125` blocked it.

Both of those blockers are gone: **125 is CLOSED, fixed and live on v1.0.1217**, and there
are 303 `generic`/`deployed` pages to choose from. And I re-read the live `page-rebuild`
row: mig 246's three edits are all still there (`start_step: plan_sections`, the step,
`section_plan` in a 9-key `input_mapping`). So on the file's own terms 087 was one live
run away from closing.

**What the file does not say is that candidate A was applied to one caller out of three
broken ones.** Four live agent definitions reference `page-content-writer`;
`pageflow-builder` and `site-work-orchestrator` have `start_step: write_page_content`, no
planning step anywhere, and no `section_plan` in the writer's mapping. 087's title —
"the writer has no section plan and nothing builds one" — is *literally true* for both.

## 2026-08-04 ~10:25 — a census trap that made the writer look fixed

My first census used `default_config::text LIKE '%plan_sections%'`. **In SQL `LIKE`, `_` is
a single-character wildcard**, so that pattern also matches the substring `plan.sections`
inside `section_plan.sections_ready` — and `page-content-writer`, which had no planning
step at all, came back as a positive.

It did not change my conclusion (the two broken callers were `false` under a *more*
permissive pattern, which is a stronger negative, and I read the step keys directly
afterwards) — but the wrong answer here looks exactly like the right one, which is the
definition of a landmine. Filed as one; recorded in the RUNBOOK with the query that is
actually safe.

## 2026-08-04 ~10:30 — the plan, and the two things I corrected in it

Fable planned candidate **D**: make the writer self-sufficient, so no caller can get it
wrong, rather than a third and fourth copy of mig 246. I took the shape and re-derived the
mechanics, and **two of its load-bearing claims were wrong**:

1. It claimed `resolve_links`' existing mapping would resolve a writer-local plan via
   `FindByPath`'s input_data fallback (`content_search.go:84-94`). **It will not.** Both
   fallback branches there guard `i == 0`. For the path
   `input_data.section_plan.sections_ready`, `input_data` *resolves* at position 0, so the
   miss happens at position 1 where neither branch applies. Consequence: on the self-plan
   branch the link resolver is handed nothing. That turned "no change needed" into a
   deliberate, documented decision (below).
2. It proposed repointing that mapping anyway. I did **not**. Seed 308 pins it on purpose
   and the 192 lane is mid-flight on the Go half; the degradation it would fix is exactly
   the one the estate has already accepted fleet-wide for the pre-roll window; and the
   repoint swaps an *exact* path hit for a three-deep fallback on the one branch that
   currently works — inside the very change I was using to prove 087. Recorded as owed,
   with the one-line edit and its post-roll test.

Everything else I verified in code before writing SQL: the bare-path truthiness grammar
(`conditional_branch_action.go:305-315`, `valueIsTruthy` `:527-551`), Strategy 5's
recursive search making a 192-wrapped plan read truthy (`:396-411`), `ExtractNestedField`'s
strict traversal (`data_helpers.go:1199-1234`), `ExtractFieldsAction` prefixing but never
stripping `input_data.` (`v3_site_actions.go:4284-4305`), and `plan_sections` making zero
LLM calls.

## 2026-08-04 ~10:40 — MISSTEP: a verify check that could never have failed

Seed 309's `DO` block asserted

```sql
IF cfg #>> '{steps,process_sections_loop,config,items_field}' <> 'sections_for_render.sections_ready'
```

Two faults compounding. The loop's key is **`iterate_over`**, not `items_field` — and
`#>>` on a missing path yields NULL, so `NULL <> 'x'` is **NULL, not TRUE**, and the
`RAISE` could never fire. A check that sits green regardless of the truth is worth less
than no check, because it reads as coverage.

Caught by reading the live step rather than trusting the name I had in my head from
`page-rebuild`'s loop (which *does* use `items_field` — two loop actions, two key names).
Fixed to `iterate_over` and every string comparison in the block moved to
`IS DISTINCT FROM`, every `@>` wrapped in `COALESCE(…, false)`.

Then I **induced** it: ran the `DO` block alone against the unmodified row and required an
exception. It raised `087/309: one of the four new steps is missing`. A verify block you
have not seen fail is a claim, not a check.

## 2026-08-04 ~10:45 — seed 309 applied

Four new steps in `page-content-writer` (`check_section_plan` → `plan_sections` →
`check_planned_sections` → `fail_no_ready_sections`), `build_render_context.next_step`
rewired, and a fourth `select_sections` fallback path appended under a containment guard so
it is commutative with the 192 lane's owed cleanup. `resolve_links` untouched, and the
verify block asserts that it *stayed* untouched.

`14 steps, branch wired, select_sections has 4 paths, resolve_links untouched`. Snapshot
`5946a27b` taken first. Live on apply — no image roll, which matters because the fleet is
mid-flight on 192's Go half and this change does not queue behind it.

The one declared behaviour change beyond 087's symptom: a writer whose *own* plan yields
zero ready sections now fails loudly instead of compiling an empty page over a real one.
`page-build-handler` already refuses that case before it will even spawn the writer; the
writer had no equivalent.

## 2026-08-04 ~10:47 — acceptance run, and picking a target without hurting anyone

`page-rebuild` rebuilds **every** armed page on the site (`build_statuses: ["needs_rebuild"]`,
`max_iterations: 20`), so "arm one page" is only bounded if the site has none armed already.
Seven sites qualified. I dropped `loancalculator.co.uk` — a live session had 846 transcript
hits on it — and picked `vetcomparison.uk` / `tool-cma-obligation-checker-guide`:
`generic` policy, three sections, untouched since 08-02, no live session mentioning it, and
crucially `name ≠ url` stem, so `/tool-cma-obligation-checker-guide.html` is a real negative
control for the `bugs_closed/125` path assertion (404 before).

Armed **inside a transaction that aborts unless the count is exactly 1** — checking
afterwards is not the same thing on a cluster this contended.

## 2026-08-04 ~10:52 — 087's acceptance test PASSES on every assertion it named

Correlation `3fdf4acf-5f96-49f9-8801-28047aae92ef`, 09:47–09:50Z. `page-rebuild`,
`page-content-writer`, `internal-link-resolver`, `content-reviewer` and `deployer-agent`
all `COMPLETED`.

- **branch asserted, not the happy path** — the writer child's `input_data` keys are
  `current_page, db_sync, hero_url, logo_url, reviewed_brief, section_plan, site_plan,
  site_record, style_collection`: the rebuild shape, including mig 246's `section_plan`.
- `sections_for_render.sections_ready` = 3, loop iterated, `compile_page` reached, run
  `COMPLETED`.
- `build_status` → `deployed`; all three `page_components` rewritten.
- canonical URL 200 with all three `data-component` slots; the name-derived URL **still
  404**. 125's fix holds.

And the **regression check on my own change**, which is the part that could have come out
otherwise: `check_section_plan.condition_met = true` and `collected_data ? 'plan_sections'`
is **false** — the caller's plan was recognised and kept verbatim, and the writer did *not*
run its own planner. The truthy branch behaves exactly as before.

## 2026-08-04 ~10:55 — the acceptance run found a second, unrelated defect: filed as 194

The rebuild worked and **NULLed `content_data` on all three components** (644 / 3,810 / 420
chars before). `rendered_html` is fine and the page serves correctly, so nothing visible
broke — but `content_data` is what the rerender path regenerates from.

**Not caused by my change**, and the evidence is the same telemetry that proved the fix:
the writer took the pre-existing truthy branch (`condition_met = true`, no local plan), so
its behaviour was identical to before; and `save_page_sections` is downstream of everything
309 touches. The mechanism is a config key four of six callers never set —
`sections_metadata_field`. Filed as `bugs_open/194` with the six-caller table.

**MISSTEP, small but worth the line:** my first look at this used
`md5(pc.content_data::text)`, which returned blank, and I briefly read blank as "the query
is odd". Blank *was* the finding — `md5(NULL)` is NULL. A NULL rendering identically to a
formatting quirk is how a real result gets dismissed; `IS NULL` as its own column is what
settled it.

I fixed `page-rebuild`'s instance (seed 310) rather than only filing it, because this
lane's own acceptance run is what NULLed those three components — recording damage and
leaving it is not an option. The other three callers are deliberately left to 194: two are
the dormant pair 087 also found broken, and `tool-recreation-handler` runs a different
writer flow whose response shape I have not read, so copying the key there would be an
unmeasured claim.

**Restoring from `page_component_history` would have been the wrong repair** even though
the archive is faithful (the three rows hold exactly 644 / 3,810 / 420). It would pair the
*old* structured content with the *new* HTML, and the next rerender would regenerate the old
page over the new one. Re-running the build writes both halves together.

## 2026-08-04 ~11:10 — MISSTEP found in the final state check: `updated_at` has no trigger

Re-reading the live row to confirm nobody had overwritten my config, I saw
`page-content-writer.updated_at = 09:01:35Z` — the **192 lane's** write — on a row that
plainly carried my 14 steps and 4 fallback paths.

`agent_definitions` has **no trigger** on `updated_at` (`pg_trigger` has no non-internal
row for the table). The column is current only when a seed sets it. Seeds `246` and `308`
do; **`309` and `310` as I first wrote them did not.**

Three things wrong, in increasing order of how much they mattered:

1. the two live rows carried a stale timestamp — fixed by hand, `10:11:38Z`;
2. both seeds would replay the same way — fixed (309 gains a stamping UPDATE inside the
   transaction, 310 sets it in the same `SET`);
3. **the RUNBOOK told the next session to rely on that column.** That is the real error.
   Corrected in place with the evidence, and the reliable check (`md5(default_config::text)`
   diffed across the read-to-write window) put in its place.

I had followed schema-first for every column I *wrote* and not for the one I *reasoned
from* — I took `updated_at`'s semantics from its name. Filed to `WRONG_CALLS.md` and, since
it fires on touch with no symptom, to `LANDMINES.md`.

## 2026-08-04 ~11:15 — final state

- `bugs_closed/087` — fixed, live, acceptance passed twice. Commit `8bafcf9d4`; the move
  verified at HEAD (`git ls-tree` returns exactly one path), no same-file passengers.
- `bugs_open/194` — filed, 1 of 4 instances fixed and proven live.
- `page-content-writer`: 14 steps, `build_render_context → check_section_plan`, 4
  `select_sections` paths, `resolve_links` mapping untouched as designed.
- `page-rebuild`: `sections_metadata_field` present.
- Target page `deployed`, 3 components, **3 with `content_data`**; nothing left armed on
  `vetcomparison.uk`.
- Concept register PBP-**030** (not 021 — `PBP-021` was already
  `load_page_record lookup semantics`; the highest `### ` heading in a category file is
  **not** the highest id in the series). Drift pair clean: 1,764 rows = 1,764 entries,
  0 each way.

**Owed, deliberately:** repoint `resolve_links`' `sections?` to the unprefixed
`section_plan.sections_ready` **after** 192's Go roll, then exercise the falsy branch via a
`pageflow-builder` / `site-work-orchestrator` dispatch — that branch is the half of this
fix no run has yet taken.
