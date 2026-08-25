# HANDOFF — loancalculator.co.uk · post-roll acceptance is GREEN; `bugs_open/385`'s cause is the only open item (2026-08-25)

> Supersedes `docs/agent_docs/docs024_key_docs_latest/loancalculator_couk/HANDOFF_2026-08-24_continue_here.md`.
> That file's "next actions" 1 and 2 are **done and verified**; its item 3 (385's cause) is
> carried here with three new constraints. Read 08-24 for how the harness broke and how 385
> was found; read this for state. Evidence: NOTES `## 2026-08-25`. Owner prose:
> `README_where_we_are.md`. The bug: `bugs_open/385_HANDOFF_2026-08-24_a_rebuild_appends_an_unlinked_copy_of_the_locked_section_it_just_repositioned.md`.

```
site        loancalculator.co.uk   0162cde4-633e-45e9-8ca6-87a6b2fe1d26
pages       28 active · 28 serving 200 · ZERO 404s · 0 pages with duplicate ids
locks       11 held · 0 orphan (component_id IS NULL) rows on this site
chassis     v1.0.1339, build provenance git_commit a7459a44b — ASKED THE SERVICE, both pods,
            started 2026-08-25 19:07. My last commit (3973260c5) IS an ancestor of it.
harness     ✅ toolgolden WORKING; --selftest green 2026-08-25
golden      ✅ acceptance/GOLDEN_2026-08-24_post_385_repair_tool_values.json
            [MEASURED 2026-08-25] all 11 reproduce it EXACTLY, exit 0, ZERO divergences —
            after the roll AND a 24-page rerender wave the same day
open        bugs_open/385 — damage repaired & verified; CAUSE NOT ESTABLISHED; writer LIVE
```

## RUN THIS FIRST, ALWAYS

```bash
cd docs/agent_docs/docs024_key_docs_latest/loancalculator_couk
python3 toolgolden.py --selftest        # green => a divergence is about the PAGE
```

The harness reports its own faults in the vocabulary of the page under test — that is what
cost this lane a day (`LANDMINES.md`, *"a browser harness that 'is down' is probably reading
`$TMPDIR`"*). The selftest drives a fixture whose answers are computed by hand from the
driver's own rules and is proven disconfirmable by mutation. **Green, or nothing you measure
afterwards is quotable.** Full acceptance run:

```bash
python3 toolgolden.py --compare acceptance/GOLDEN_2026-08-24_post_385_repair_tool_values.json \
  $(the 11 URLs — take them from pages.url or from the golden's own "pages" key, never
    name-derived)
```

## WHAT MOVED UNDER THIS THREAD, and how it was checked

750 commits landed between `3973260c5` and HEAD. Checked, not assumed:

- **Nothing of this lane's was touched** — `git log 3973260c5..HEAD --` over `toolprobe.py`,
  the whole `loancalculator_couk/` directory and the 385 bug file returns **empty**.
- **385's writer is unchanged in the running build.** `save_page_sections_action.go` has one
  commit in the window (`c735bfd9c`, `bugs_open/375`'s verifier gate) whose diff **for that
  file is empty**. So the roll did not fix 385 and nobody claims to have.
- **A peer lane is on the neighbouring seam**: `bugs_open/357` (a whole tool page stored in a
  slot claiming to be a hero component) is ACTIVE and shipped opt-in, default-OFF identity
  work on the compile hop. Cross-check with it before touching slot identity — but note it is
  **not** 385, and `who-owns.py 385` returns this lane.

## `bugs_open/385` — the one open item, with three constraints added today

**Not recurred, anywhere.** `[MEASURED 2026-08-25]` 385's own discriminator —
`byte_twins_on_page`, **not** a bare `component_id IS NULL` count, which over-reports by 10
of 11 — is **0** for all 9 remaining orphan rows fleet-wide.

**All ten locked tool pages rebuilt clean today**, 13:04–13:34, 4 of 5 rows rewritten, locked
row untouched, none duplicated — **including `/tools/loan-vs-savings.html` at 13:11, the
victim itself.**

> ⚠ **THE TRAP IN THAT SENTENCE, and it is the most important line in this handoff.** That
> wave was **24 `page_rerender` items, `source='side_effect'`**, from a `tool-cta` template
> change — the **rerender** arm. The 08-23 damage came from a **`needs_page`** item on the
> **build** arm (`page-build-handler` → compile → `save_page_sections`). **Two different
> upstreams into the same INSERT.** Today establishes only that *the rerender arm is clean on
> a locked positional tool page*. **The build arm has not run on one since it failed.** A fix
> verified through `page_rerender` alone has not been verified against this bug.

**A lead chased and CLOSED negatively — do not re-chase it.** The
`source='save_page_sections_overwrite'` rows in `page_component_history` are a **pre-overwrite
snapshot of rows that already existed** (`save_page_sections_action.go:830-843`,
`SELECT pc.id …`) — they describe the page BEFORE a rebuild, not the composition it received.
The calculator's marker carries a `page_components` **row id** where a `content_components` id
would go, which reads like a type confusion explaining everything. It is not: **153 of 23,627
markers fleet-wide carry a row id, 101 of them on this site across 12 pages between 08-03 and
08-25, against ONE duplication.** It tracks locked/retained rows. **A 101-to-1 ratio is not a
cause.** ("The code changed underneath the data" was checked and refused too —
`git show ec653247f:` carries the identical snapshot SQL.)

### The next move, sharper than it was

§5b of the bug file still holds: the plan demonstrably does **not** list the tool twice
(`site_plan_sections` for `tool-loan-vs-savings` is `hero · tool-loan-vs-savings ·
ported-prose · faq · tool-cta`, the same shape as two pages that rebuilt correctly), so
something between the plan and the save added a sixth entry. Unread: **LOCK-008's merge**
(`platform/orchestration/datahelpers/locked_page_sections.go`, `MergeLockedPageSlots` — it
pairs list entries to locked rows in three arms and inserts every *unpaired* locked row at its
live position) against **`matchLockedRow`**'s arms in `save_page_sections_action.go`. If the
two disagree about what "already in the list" means, the merge inserts an entry the guard then
fails to pair — which is this bug's exact shape. The file's own header says the arms mirror
each other; **that claim is the thing to test, not to trust.**

⚠ **The `090` loop is not available for this.** It returned `UNVERIFIABLE — iteration-cap`
with zero non-bundle artifacts; `v3_site_actions.go` is 344 KB and
`save_page_sections_action.go` is 89 KB. If you re-file, name **ONE** symbol
(`MergeLockedPageSlots` is the obvious candidate and is in a small file), and read the
LANDMINE first — a single-symbol scope has still failed on other lanes.

## Standing cautions (carried; all still true)

- **Prove a deploy at the artefact.** Ask the service for its `build provenance`, per SERVICE
  not per fleet; never grep the binary for a fix's own sha.
- Verify tool placement at `site_plan_sections`, never `pages.sections` (LOCK-008 merges).
- ⚠ **`UPDATE page_components SET position` does NOT touch `updated_at`.** "The locked row's
  `updated_at` never moved" proves the BYTES were not rewritten, **not** that the row was not
  repositioned.
- ⚠ **Before any repair of 385's shape, check `pages.sections`** — a stale entry in that
  materialised cache lets an assemble re-materialise the duplicate and makes the repair look
  done.
- **A single sample during a deploy or rerender wave proves nothing — of a 404, and equally of
  BYTES.** A page differing from its peers mid-sweep looks *skipped* and usually means *not
  yet reached*. This cost a false "the publish failed" report on 08-24 (`WRONG_CALLS.md`).
- A hand-filed or un-parked work item must be `triaged`; the dispatcher cannot see `detected`.
- `retract_page_deployment` REFUSES an active page (archive first) and its DEFAULT selection
  also takes `tool-standard-calc` — use explicit `page_ids`.
- Query runs BY CORRELATION, never `now()`-interval; a planner run's `collected_data` can
  purge within ~2 hours.
- **Before any planner run**, the four cautions in `HANDOFF_2026-08-23_continue_here.md` still
  apply verbatim (`checkpoint_postplan.sh` immediately; check item KEYS against the plan the
  run just wrote; re-verify identity flags; Pass C2 will not save you).
