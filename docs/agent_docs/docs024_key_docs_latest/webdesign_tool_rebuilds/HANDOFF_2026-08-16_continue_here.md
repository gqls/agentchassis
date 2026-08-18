> **SUPERSEDED 2026-08-17 by `HANDOFF_2026-08-17_continue_here.md`.** Kept for the record: its
> "pilot is complete" section is still accurate, but the state table, the batch scope and the
> next-actions list are all out of date, and it does not know about the fleet-wide unique index
> (`idx_cc_tool_function_unique`) that blocks two of the tools it tells you to file.

# HANDOFF — webdesign tool rebuilds. START HERE. Written 2026-08-16 ~10:15Z, UPDATED 16:50Z. Supersedes `HANDOFF_2026-08-15_continue_here.md`.

Read `PLAN_2026-08-15_…` (design + the owner ruling + its 16:20Z correction), `RUNBOOK_…` (commands),
`NOTES_…` (evidence + missteps, newest at the bottom), `SUMMARY_2026-08-16_…` (the milestone read-out),
`bugs_closed/286`-or-`bugs_open/286` (the pilot's cause + fix — check which dir it is in now).

## THE PILOT IS COMPLETE AND LIVE. The replacement recipe is proven end to end. (16:47Z)

`https://webdesign.co.uk/tools/aspect-ratio/index.html` serves the framework-built tool. Graded at the
served bytes 16:48Z: `class="ported-page"` **0** (was 1), `{{\.` **0**, the tool's own controls present
(`ratio-width`/`ratio-height`/`target-ratio` ×3 each, `preset-btn` ×8, 6 inputs, 5 buttons),
**861 visible chars**, md5 changed, negative + positive controls both run. DB: ported slot `removed`
with its 5,850 chars retained, `tool-aspect-ratio` `deployed` at position 2 — one live slot.

**The recipe, in the order that matters (each step earned its place — see NOTES 16:47Z):**
1. File the `add_tool` item, description written from the LIVE tool's behaviour, self-contained copy.
2. **Grade the generated component in the DB before retiring anything** — `{{\.` count 0,
   visible chars > 300, `<script` ≥1, no `<script src=`, **and read the JS** to confirm it does the job.
3. **Retire the ported slot BEFORE the generator's own rerender is claimed.** The generator files that
   rerender itself, already assemble-only — you do not file one. Win this race and the page renders
   once, correctly; lose it and the page serves two tools (the ab-test shape).
4. Let that rerender run, then grade at the served URL **with controls**.

## Verified state (2026-08-16 16:50Z — supersedes the 15:40Z table)

| thing | state |
|---|---|
| fleet binary | `v1.0.1304`, stamp `5de6cddbe`; `88897190e` IS an ancestor ⇒ **286 fix LIVE**. |
| seed 435 | APPLIED 15:15Z (`ee0228813`): `tool-generator.save_tool.adopt_existing_page = true`. |
| **pilot `tool-aspect-ratio`** | **DONE — built, graded, retired, re-rendered, PASS at the served page.** |
| ab-test | ported tool restored and served; **rebuild candidate #1 of the batch.** |
| batch scope | **62 ported tools remain.** 97 `ported-page` slots exist but only 63 were tools. |
| owner gate | **OUTSTANDING — PLAN §4: the owner sees the served page before the batch starts.** |

## Next actions, in order

1. **Show the owner the served page** (`/tools/aspect-ratio/index.html`). PLAN §4 gates the batch on
   this, and it is the one step no machine can do. **Do not start the batch before it.**
2. **ab-test first, and it needs one statement before it is filable.** The generator's already-exists
   probe has **no `build_status` predicate**, so the withdrawn fork blocks its own rebuild and the run
   completes having written nothing (LANDMINES, footprint `create_tool_component_action.go`; confirmed
   from source, not inherited):
   `UPDATE content_components SET is_active=false, updated_at=now() WHERE id='cd60486c-f5e1-4d80-9676-0d65024f0372' AND function='tool-ab-test-calculator' AND is_active;` (1 row) — then file.
3. **Then the simple batch, serial** (PLAN §2), recipe above, one tool at a time. Scope it with the
   RUNBOOK's "Scope the batch correctly" query — **filter `p.name LIKE 'tool-%'`, not
   `p.url LIKE '/tools/%'`** (`tools-index` is a ported page and is not a tool).
4. **Rich apps last and one at a time** (owner ruling reversed PLAN §3 — all 63 are in scope). Spec
   written from the live tool's behaviour; the grade is a feature list in a browser, not a tag count.
5. Owed to others, unchanged (122 ink lane's audit checks ~08-18; mindmap junk text = owner's localStorage).

## Traps this lane has paid for (all in LANDMINES / WRONG_CALLS / NOTES)

- **`/tools/<x>/` is a 404; the page is `/tools/<x>/index.html`.** Every "is it clean now?" count
  passes perfectly against a 404 — it contains none of what you are grepping for. Assert `http=200` first.
- **There is no `page_component_history` archive row for a retire** — the trigger is
  `AFTER UPDATE OF rendered_html`. PLAN's 16:20Z correction supersedes the ruling's condition 3: the
  undo handle is the ported `page_components` row + its md5, recorded before the retire.
- **Two different sets of exactly 13**: `<script src=` tools (the PLAN/TL-032 class) and
  `content_data ? 'repair'` tools. **Intersection 4.** Do not use one as a proxy for the other.
- An empty `final_result` is the workflow's `output_fields` shape, not a failure; read `agent_error_log`.
- `spec_data` as a MAP is silently unread — census `jsonb_typeof` on any copied `create_work_item` step.
- A tool slot can be 13 KB and have zero visible text; the class-attr floors will not tell you.
- Guard "open items on this page" on `('triaged','approved','claimed','pending')` only.
- **A `090` on a symbol in a file over ~60 KB returns bundles and no verdict** — looks like a run in progress.

## Not this lane's, but it sits under the batch

`bugs_open/289` (OPEN, unowned): a `loop_complete` substep re-aggregates the whole loop from inside each
iteration, so `collected_data` doubles per lap. **`build-dispatch-loop` — the dispatcher every item on
this lane goes through — is exposed and was at 13 MB**; `tool-auditor` is already dead at 22 MB. A batch
of 62 serial items runs through it. Watch for the dispatcher stalling rather than assuming a slow queue.

## Owner ruling 2026-08-16 (unchanged) — PLAN §3 REVERSED

The rich hand-built apps ARE rebuild candidates (Mind Map Studio Pro, Meme Studio, Logic Architect Pro,
Flat-File Micro CMS, Pasteboard Manager, and the 13 external-script tools): option (a), generator rebuild
anyway, **accepting that it is a reimplementation, not a preservation**. No excluded class remains, and
`bugs_open/204` / the byte-faithful route is no longer a prerequisite for any webdesign tool.
Full ruling + reasoning: PLAN §"OWNER RULING 2026-08-16", README_where_we_are 2026-08-16 afternoon.

---

> **[Dated pointer from the bugfix_283 lane, 2026-08-18 — appended, your words untouched.]**
> Your native tools' component TEMPLATES now carry `{{.InstanceID}}` element-id namespacing
> (68 components converted 2026-08-18 via `scope_component_instance`, RFC_034; snapshot per row in
> `component_versions`, `change_source='scope_component_instance'`). Three things that touch you:
> **(1)** your OWNED pages were deliberately NOT rerendered — `save_page_sections` refuses generic
> saves on them (correctly), so their served bytes keep pre-conversion ids until YOUR pipeline next
> renders; the fixer now skips owned pages entirely (migration 462). **(2)** your render path is
> outside `RenderTemplate`, so when it next renders a converted template, check once that
> `{{.InstanceID}}` is either bound or absent from your path's inputs — an unbound token renders as
> an empty string. **(3)** if you REGENERATE a component, regeneration REPLACES the template and the
> conversion goes with it — that is expected, not a defect; the conversion items are idempotent and
> can re-run after your rebuild. Contact/context: `bugs_open/283` §13.6.
