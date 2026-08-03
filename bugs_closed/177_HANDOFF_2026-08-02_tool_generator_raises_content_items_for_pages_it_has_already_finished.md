# 177 — `tool-generator` raises a "write content" item for a tool page it has just finished building; 8 of 8 have failed identically since 2026-07-14

**Filed 2026-08-02** (session "bugfix 19"), while triaging the item that stalled
the fleet in `bugs_closed/176`. **Status: OPEN, undiagnosed at the generator.**
The *consequence* is measured and certain; the *cause inside `tool-generator`* is
`[UNVERIFIED]` — I have not read the generator's workflow, and this file does not
claim to know why it emits the item.

Related: `bugs_closed/176` (the dispatch stall this fed), `bugs_open/033` (the
needs_human_review queue these accumulate in), `bugs_open/015` (same handler
no-op signature, different cause — stranded nav pages).

## The defect

`tool-generator` builds a tool page, its component and its `page_components` row,
and then **45 milliseconds later** raises a `needs_content_page` work item asking
for content to be written for that page:

```
content_components 769c0b80   updated 2026-07-31 12:27:28.312336
pages              1f939069   created 2026-07-31 12:27:28.326869
page_components    (1 slot)   updated 2026-07-31 12:27:28.342761
site_work_items    0733a7a4   created 2026-07-31 12:27:28.387981   <- 45ms later
```

`page-build-handler` claims it, finds no sections to build, and no-ops:

```
page-build-handler no-op: no sections ready to build (empty spec sections,
or all sections deferred for missing data) — the target section was NOT rebuilt
```

**The handler is not at fault and neither is the routing.** There genuinely are no
sections to build, and routing the no-op to `needs_human_review` instead of
silently completing is the WDS-004 fix behaving exactly as designed. The defect
is that the item was raised at all.

## One slot IS the finished shape of a tool page

Measured fleet-wide 2026-08-02 — this is the load-bearing measurement, because it
is what makes the item spurious rather than merely early:

```
slots | tool_pages
    0 |  14        (never built)
    1 | 116        <- the normal, complete shape
    2 |   7
    3 |   1
    4 |   2
    6 |   1
```

All five *deployed* robot-hands tool pages have exactly 1 slot. The page in
question is `deployed`, has `deployed_at` set, and carries a 10,336-char tool
component. So the item asks for work the platform does not do for tool pages.

## It is systematic: 8 of 8, every one since the class began

```sql
SELECT status, count(*) FROM site_work_items WHERE item_key LIKE 'tool_content:%' GROUP BY 1;
--  needs_human_review | 8      <- and nothing else, ever
```

Every one at `attempt_count = 1`, every one with the byte-identical error,
spanning 2026-07-14 → 2026-07-31 across 4 sites. **Not one `tool_content` item
has ever completed.** The class has a 0% success rate over its entire history.

## Why it matters more than 8 stuck rows

These items are not inert. On 2026-08-02 one of them (`0733a7a4`) was named in
another item's `depends_on`, and because **a dependency is only released by
`complete`/`verified`** — `wont_fix`, `rejected`, `cancelled` and `failed` all
leave the dependent blocked for ever — it became a permanent blocker. Combined
with the selector/loader disagreement in `bugs_closed/176`, that single row
stalled **the entire fleet** twice in one day (89 min and 68 min). See 176 for
that mechanism; it is fixed. What is not fixed is the generator that keeps
minting the blockers.

`93f2a3b7` was the FIRST `tool_crosslink` item ever to carry a `depends_on`, so
this coupling is new behaviour and the blast radius is growing, not shrinking.

## Fix candidates, ordered by what closes the door

1. **Do not raise the item when the page is already complete.** If
   `tool-generator` has just written the page's only planned slot and deployed
   it, there is nothing to ask for. Highest-value: it makes the bad state
   unrepresentable rather than cleaned up afterwards. Needs a read of the
   generator's workflow to know where the emit sits relative to the build.
2. **Raise it only when the page spec actually declares prose sections.** Weaker
   but more general — it also covers a future tool page that legitimately wants
   content. `[UNVERIFIED]` whether the spec is available at emit time.
3. **Give `page-build-handler`'s no-op a "nothing to do" outcome distinct from
   "needs a human".** Fleet-wide change, and it would mask 1 and 2 rather than
   fix them — a page that SHOULD have sections and has none is a real defect
   (`bugs_open/015`), so this must not become a blanket "no sections = fine".
   Do not do this without separating the two cases first.
4. **Sweep the 7 remaining rows.** Cleanup, not a fix; they block nothing (only
   `0733a7a4` ever had a dependent). Belongs with whichever of 1–3 ships.

## Verify a fix

Generate a tool page and assert no `tool_content:%` item is raised for it, or that
one is raised only when the page spec declares sections beyond the tool:

```sql
SELECT left(id::text,8), status, created_at, left(error,80)
FROM site_work_items WHERE item_key LIKE 'tool_content:%' ORDER BY created_at DESC;
-- expect: no new row for a single-slot tool page
```

Do NOT verify by watching the item succeed — the correct outcome is that it is
never created. A run that produces a completed `tool_content` item means
something built prose sections, which is a different (and possibly also wrong)
behaviour.

## What was done at triage time (2026-08-02)

`sql_for_agents/286` — 2 rows only, on the pair that stalled the fleet:

- `0733a7a4` → **`wont_fix`**, with the original handler error preserved inside
  the new reason string. Deliberately **not** `complete`: no work was ever
  performed, and `complete` would assert a success that did not happen — the
  silent-completion pathology committed by hand. (It is also the only status that
  would have released the dependency, which is precisely why it had to be
  refused.)
- `93f2a3b7` → `depends_on` cleared, left `triaged` to be attempted on its own
  merits. The crosslink is genuine outstanding work: verified that none of the
  three `page_components` on `/how-to-specify-a-gripper.html` references
  `gripper-safety-factor-calculator`.

**Watch item, not part of this bug:** 5 of 5 previous `tool_crosslink` items
failed at `validate_page_content` ("content validation failed: N blockers,
M errors"). If `93f2a3b7` fails the same way, that is a separate defect in the
crosslink path and should be filed as such — not re-diagnosed as a dependency
problem.

---

## UPDATE 2026-08-03 (154 lane session) — a 090 run is IN FLIGHT for this bug; do not fire another

- This session, continuing the 154 lane whose handoff listed 177 as "unstarted",
  dispatched a 090 diagnosis at ~10:47Z: intake
  `needs_diagnosis:177-tool-content-item-unsatisfiable-at-birth`, **RUN
  correlation `da59941f-8d16-4c3a-9812-e9f76064de28`** (artifacts key). Minutes
  later the untracked `bugfix_177_tool_content_items/` PLAN (11:41 BST)
  surfaced: the lane had been taken concurrently, and neither session could see
  the other — who-owns reads commits, the trigger's probes read
  `site_work_items`, and the new lane had touched neither. The misstep and its
  check are in `WRONG_CALLS.md` (2026-08-03, "177 is unstarted").
- **Do not dispatch a second 090** — the in-flight intake's seed_scope covers
  `create_tool_component_action.go` + `deploy_tool_action.go`, so the trigger's
  coverage probe will refuse anyway. Find the verdict under the run correlation
  above (`diagnosis_artifacts`, `doc_notes`).
- **Read the verdict against this caveat:** the dispatched symptom wrongly
  attributes the empty `sections` declaration to BOTH emit paths. The 177
  lane's sharper finding (deploy path declares 4 sections at
  `deploy_tool_action.go:343-346`; every dead item is create-path) supersedes
  that framing — a verdict refuting the both-paths claim refutes the SYMPTOM's
  framing, not the lane's diagnosis. What the run is still good for: the
  2026-07-31 owner ruling asks for a loop pass over a first-hand-verified
  structural root cause (the 155 precedent), and this is that pass.
- One observation for the close-out, found during the quick look:
  `spec.content_guidance` is written by all four tool emit sites and read by
  NO handler on the work-item path — its only reader
  (`apply_gap_plan_action.go:178`) takes it from a gap plan, not from an item
  spec. It is dead weight even on satisfiable items.

### Verdict landed (same day, ~3-minute run): REFUTED — and the refutation independently re-derives the lane's asymmetry

- Run `da59941f` completed in ONE iteration: outcome **REFUTED**, citing
  `deploy_tool_action.go`'s explicit
  `sectionsJSON = ["hero-tool","tool-guide-intro","<fn>","tool-cta"]` against
  the symptom's both-paths framing, and the create path's `pages` INSERT that
  never mentions the `sections` column. Revised hypothesis, core verbatim:
  *"The two emitters are not siblings behaving identically … the no-op, if it
  occurs, is more likely isolated to (or at least differs between) the
  create_tool_component_action.go path and whatever page-build-handler
  actually does with those specific section names, not a uniform
  platform-wide mismatch."*
- **That is an independent re-derivation of this lane's PLAN's central
  finding** (create path omits the declaration; deploy path declares four),
  from the seed code alone — the 2026-07-31 loop pass, delivered in the
  REFUTED-is-success form. The loop then stopped honestly
  (scope-not-narrowing; run status UNVERIFIABLE; "hand to a human, do NOT
  auto-conclude").
- Full verdict JSON (citations, data_requests, next_scope) preserved before
  the ~24h reaper:
  `bugfix_154_work_item_routing_columns/EVIDENCE_2026-08-03_177_verdict_da59941f.json`;
  bundle iter 1 durable in `diagnosis_artifacts` under `da59941f%`; terminal
  summary in `doc_notes` (categories diagnosis, unconfirmed-diagnosis).
- A rigor point worth keeping: the verdict declined to treat even
  `sections=[]` as established for the create path (the bundle never showed
  the column default). Both this session and the owning lane verified the
  LIVE row is `[]` — the claim stands on state evidence, not on inference
  from the emitter's code.

---

## CLOSED 2026-08-03 — fixed AND live (v1.0.1241, pod-proven both replicas)

**Root cause, verified** (was `[UNVERIFIED]` above): not a workflow defect in
tool-generator but a two-file asymmetry. `create_tool_component_action.go`
copied `deploy_tool_action.go`'s item emission WITHOUT its section declaration
— its `pages` INSERT names no `sections` column, so the page is born `[]`, and
page-build-handler's whole resolution chain (site_plan_sections →
`site_specs.site_plan` → `pages.sections` → sibling synthesis, which needs plan
membership) finds nothing. Control that pinned it: `tool_guide:%` items from
the SAME two files, whose page DOES declare sections, ran 4 complete / 1
review. All 9 `tool_content` items ever created were the create path's; zero
were tool-deployer's.

**Per the owner ruling of 2026-07-31**: no `090` run was filed for this root
cause; the substitute was equivalent first-hand verification, stated plainly —
both emit sites read at HEAD, the live `page-build-handler` workflow config
read from `agent_definitions`, the loader action read in full, the 9/9
attribution and the tool_guide control measured live, and the guard's edge case
(33 current-plan section rows for tool-named pages) measured before design. A
diagnosis note on this exact mechanism (doc_notes, pipeline/tool-generator,
2026-08-03) was independently filed by another session and is answered/
superseded by the fix note beneath it.

**What shipped** (council `982507b0` APPROVED round 1; commit `74655b709`,
`Council-Reviewed` trailer):

- Fix candidate 1+2 merged as an emit-side satisfiability guard:
  `raiseToolContentItem` (`tool_content_item.go`), one seam for both tool
  paths. Resolves the page's declared sections read-only in the handler's own
  priority order; raises only when a prose section beyond the widget exists;
  skip surfaced as `content_item: skipped_no_prose_sections` in the action
  output. Write routed through the shared `insertWorkItem`
  (`recurrenceExpected: true`, the `gapPlanWorkItem` precedent). 8 sqlmock
  tests including mutation-hardened guards.
- Candidate 4: `sql_for_agents/297` applied — 8 zombies → `wont_fix` (original
  errors preserved), NOT `complete` (no work happened; `complete` would release
  dependents on a lie).
- Candidates 3 (handler no-op split) deliberately NOT done — see
  `bugs_closed/015`: empty sections can be a real defect the no-op must keep
  surfacing.

**Deviation from the approved plan, narrowing (recorded in the lane PLAN):**
the two `tool_crosslink` dependents (`9e9ec430`, `18bc832c`) were NOT released.
Between verdict and apply, the diagnosis on their class completed: dispatching
one (`93f2a3b7`, released 08-02 on exactly the "stands on its own merits"
reasoning this plan copied) regenerated whole slots and dropped paragraphs —
`bugs_open/178`'s mechanism. They stay dep-blocked as a visible interlock;
`bugs_open/178` (owner: the 154 lane) releases them with its fix — contributed
into that file 2026-08-03.

**Verification record:**
- Unit: 8/8 pass; 4 mutations each caught (prose guard off, recurrence off,
  widget-as-prose, priority-order broken).
- Pod, both replicas, same exec: `raiseToolContentItem` 7 (added),
  `Failed to create tool content work item` **0** (removed — non-zero on any
  pre-fix image, so the zero is disconfirmable), `skipped_no_prose_sections` 1.
- Queue: `tool_content:%` = 9 wont_fix, 0 open; the two interlocked dependents
  still `triaged` + blocked, by design.
- **Watch (the live skip arm):** tool generation is work-item driven (~1 per
  2 days fleet-wide). On the next generation of a NOVEL tool for a planless
  page, expect NO new `tool_content:%` row and a logged
  `skipped_no_prose_sections`; a deploy-path fork should still mint one.
  `SELECT left(item_key,60), status, created_at FROM site_work_items WHERE
  item_key ~ '^tool_content:' ORDER BY created_at DESC LIMIT 3;`

**Left open, tracked elsewhere:** `bugs_open/187` (the 24+ `needs_page` rows
with the same no-op error, five other emitters — filed at the council's
direction); TL-009 (should generated tool pages DECLARE prose sections? owner
call); the companion-guide emits still hand-roll their INSERT (works today,
4/5 complete — a tidy-up, not a defect).

Lane docs: `docs024_key_docs_latest/bugfix_177_tool_content_items/`.
§9 pattern: "A work item can be UNSATISFIABLE AT BIRTH".

## ADDENDUM 2026-08-03 ~20:10 UTC — the RAISE arm witnessed live, hours after close

`2a9d693d` (source `tool-deployer`, `tool_content:tool-model-approach-selector`,
a NEW site — not the old zombie's site) was minted at 17:55 UTC **through the
guard** on a post-fix binary, and at 18:15 UTC became **the first completed
`tool_content` item in the class's history**. The page now holds the
four-section shape the code comments have described since April: `hero-tool`
(3,255 B) / `tool-guide-intro` (8,043 B) / the widget **intact** (16,467 B) /
`tool-cta` (8,487 B) — prose written around the tool, widget not clobbered
(the TL-001 worry this file raised about a completing item did not bite,
because the deploy path declares the widget as its own section).

So the two arms stand: **raise arm witnessed and completed** (deploy path,
declared sections); **skip arm still watch-listed** (create path, planless
page — no generation has occurred since the roll; the census shows 9 wont_fix
+ 1 complete + 0 parked, i.e. no new zombie). Fix re-proven at the artefact on
v1.0.1243 (two fleet rolls later): added symbol present, removed emit string 0,
both replicas.
