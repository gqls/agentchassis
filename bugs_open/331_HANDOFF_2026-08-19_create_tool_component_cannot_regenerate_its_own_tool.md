# 331 — `create_tool_component` can CREATE a tool for a site but cannot REPLACE one it built there: the per-site probe no-ops, `UNIQUE(name)` collides, and the old slot is retired by hand in a race

**Filed 2026-08-19 by the `bugs_open/286` session while closing 286 (the PAGE half of the same wall). Status: OPEN — FIX BUILT + TESTED 2026-08-19 (register **TL-047**, council round `7a82c943-a68a-4e12-bd4b-80db5b4cad72` SUBMITTED 21:45Z), inert until a chassis roll + seed `496_tool_generator_replace_existing_HOLD.sql`.**
Diagnosis: `090` intake `b6e461d0-0374-41cf-8f60-71e22af2ea89`, run `44ada235-64a7-4f25-b8bd-8fadb0c78a4e`, filed 20:58Z — **CONFIRMED, first iteration (21:02Z); see §6.**
Owner of the affected lane: `docs024_key_docs_latest/webdesign_tool_rebuilds/` (the lane that walked into every gate below; their RUNBOOK works round all of them by hand). This bug is filed INTO that lane's problem, not beside it.

## 1. Symptom, in the lane's own words

NOTES 2026-08-18 13:51Z: *"the generator can build a given tool for a given site EXACTLY ONCE, EVER … The recipe as written could ship a tool but never fix one."* A re-fix of a native tool today needs, by hand, before filing: `UPDATE content_components SET is_active=false` on the old row (RUNBOOK "Before REFILING…"), a RENAME of the old row (`…-v1-retired-<date>`) to free `content_components_name_key`, and after the build a guarded `UPDATE page_components SET build_status='removed'` on the old slot **before the generator's own rerender is claimed** — margins measured 2–96 min, "no floor", lost once (RUNBOOK "The retire race"; the page served BOTH tools).

## 2. The write history names the gates — each fired exactly once, on consecutive days `[MEASURED 2026-08-19 16:20Z]`

`agent_error_log`, all-time, `duplicate key` messages for `create_tool_component`:

| date | constraint | scope | state |
|---|---|---|---|
| 08-15 18:29Z | `pages_site_id_name_key` | page | **fixed + live** — `bugs_closed/286`, TL-044, `v1.0.1304` |
| 08-17 12:12Z | `idx_cc_tool_function_unique` | fleet-wide library claim | **fixed + LIVE** — RFC_036 §9.3, `e24bc9c0f`, council `ceae30f2` APPROVED r1, ancestor of the `v1.0.1316` stamp `07eeba4a1` (rolled 17:13Z 08-19) |
| 08-18 13:51Z | `content_components_name_key` — `UNIQUE(name)`, **no predicate** | fleet-wide, total | **no handler anywhere** (grep: no `ON CONFLICT (name)`, no retry, no suffix in any Go writer) |

Plus the gate that fires silently: the action's per-site probe (`create_tool_component_action.go` "Check if already exists": `cc.function + component_level='tool' + p.site_id + cc.is_active`, no `page_components.build_status` predicate) returns `{already_exists:true, component_id, function}` — **nothing reads `already_exists`** (no step, condition or test; only docs), and the downstream `enqueue_rerender` step reads `create_result.page_id`, which that map lacks, so the run completes having written nothing (LANDMINES "The tool generator's 'already exists' probe ignores build_status…" — filed by the lane 08-16).

Three consecutive days, three gates, each found by walking into it. That pattern — not any one constraint — is the bug: **the action has a CREATE path and no REPLACE path**, so every replacement is a hand-assembled sequence of DB edits around it.

## 3. Mechanism (read at HEAD `6bb3779aa`, 2026-08-19)

`platform/orchestration/actions/create_tool_component_action.go`:
- probe → `already_exists` early return (the per-site throttle; LANDMINES says correctly "why not just fix the query": `is_active` is a real signal and the probe stops two sessions building the same tool twice);
- `componentName := fmt.Sprintf("%s-%s", function, domainSlug)` → bare `INSERT INTO content_components (… name …)`; a second generation for the same (function, site) is the identical string ⇒ 23505 on `content_components_name_key` even after the incumbent is `is_active=false`;
- §9.3 library-claim lookup (`component_level='tool' AND forked_from IS NULL AND is_active`, no site filter) ⇒ on a second-ROW rebuild of a native tool it finds the incumbent itself and records v2 as `forked_from = v1.id`; retiring v1 then leaves the function with **no library entry** — for webdesign.co.uk, whose product is tools other sites fork, every re-fix would silently drop the tool from the library. (`[INFERRED from the code; not yet exercised]` — no second-row native rebuild has run since §9.3 rolled.)
- No Go code sets `content_components.is_active=false` or writes `page_components.build_status='removed'` — both exist only as hand-run SQL (15 `removed` rows live).

**The estate already solved this class for the other writer.** `store_generated_component` (sections) regenerates **in place** — CTS-009 "created/regenerated; `already_exists` removed": same `component_id`, FKs intact, no relink; `lookupBaseComponent` (`component_storage_identity.go:157-184`) deliberately omits `is_active` *because a regeneration of a deactivated row otherwise falls to the creation branch and hits the unique-on-name constraint* — the very collision above, learned once already. `update_component_html_action.go:272-338` is the in-place template updater (snapshot to `component_versions`, UPDATE, placements → `pending`). RFC_036 §11 (architecture seat, APPROVED round) asked that the tool writer's collision convention be recorded there; §12 now does.

## 4. Fix candidates, ranked by what closes the door

1. **Regenerate in place under a per-ITEM `replace_existing` input (TAKEN — built this session, register TL-047).** Optional input on `CreateToolComponentInputSpec`, mapped by a seed line `"replace_existing": "input_data.spec.replace_existing"` on `tool-generator.save_tool`; absent ⇒ byte-identical today (the probe's throttle stays). When set and the probe finds an incumbent: snapshot the incumbent's `html_template` to `component_versions` (the `update_component_html` statements, extracted to one helper), then in ONE transaction: lock the incumbent's live placements on this site's pages, `UPDATE content_components SET html_template, display_name, description, category, source_agent_type, source_orchestration_id, updated_at WHERE id=<incumbent> AND is_active AND component_level='tool'`, `UPDATE page_components SET rendered_html=<new html>, build_status='deployed', updated_at WHERE component_id=<incumbent> AND page_id IN (this site) AND build_status IS DISTINCT FROM 'removed' AND <agent-writable>` (tools carry the template verbatim as `rendered_html` — the assembler reads `rendered_html`, never the template; this UPDATE fires `trg_page_component_artefact_archive_upd`, so `page_component_history` gets the old bytes — the revert handle a status flip never produced); 0 rows on either ⇒ rollback + typed refusal. Return early with `regenerated:true`, same `component_id`, `page_id`/`page_url`/`function` (what `compose_plan`→`enqueue_rerender` read). `forked_from`, `name`, `function` untouched ⇒ zero uniqueness gates, the §9.3 wrinkle cannot occur, the page never holds two slots, no race. Unsafe default OFF and per item, so a duplicate `add_tool` from the suggester stays a no-op.
2. Rename the incumbent + insert a second row + atomic slot swap (`removed`). Closes the race, but must deactivate the incumbent BEFORE the §9.3 lookup or v2 forks from v1; invents a third collision convention; first programmatic writer of the `removed` tombstone. Rejected for 1.
3. Typed early refusal when the name is taken (RFC_036 option 3's shape). Converts a 23505-after-LLM-spend into an early refusal; unblocks nothing.
4. Leave it: the RUNBOOK's three hand steps per re-fix, and the race, for every site that ever re-fixes a tool.

## 5. How to close (fixed AND live)

1. Roll carries the fix commit: probe the stamp, `git merge-base --is-ancestor <fix> <stamp>`, with a junk control.
2. Apply the HOLD seed (`docs/agent_docs/sql_for_agents/495_tool_generator_replace_existing_HOLD.sql` → rename/apply): adds the `replace_existing` mapping to `tool-generator.save_tool` ONLY.
3. Re-fix one already-native tool (the lane's next v2 candidate) with `"replace_existing": true` in the item spec; expect `create_result.regenerated=true`, `component_id` == the incumbent, a new `page_component_history` row for the slot (old bytes), exactly one deployed tool slot on the page, NO new `content_components` row, and the served page cache-busted with an old element id as the negative control.
4. Then this file → `bugs_closed/`, and the lane's RUNBOOK "deactivate first / rename / retire before the rerender" steps marked retired for re-fixes.

## 6. Diagnosis-loop verdict

**CONFIRMED, first iteration, 2026-08-19 21:02Z** (run corr `44ada235-64a7-4f25-b8bd-8fadb0c78a4e`; `orchestration_states.collected_data->'verdict'->'result'`, code-tier run, no `doc_notes` row by design). Six citations: three static — the probe SQL (`… cc.is_active = true LIMIT 1`, "already-exists probe"), `componentName := fmt.Sprintf("%s-%s", function, domainSlug)`, the bare `INSERT INTO content_components` — and three runtime, the three `agent_error_log` rows above (`fresh` 08-15 / 08-17 / 08-18). Its `symptom_check`: *"the name is derived deterministically … the bare INSERT carries no ON CONFLICT/collision handling … for the same site the log shows successive regeneration attempts dying on three different uniqueness constraints in turn"* — `explained: true`. **One honest caveat:** the INSERT it quotes has nine bind parameters (no `forked_from`), i.e. the loop read a code snapshot that predates `e24bc9c0f` (§9.3, committed 16:01Z the same day) — so parts (a) and (b) of the symptom are CONFIRMED on evidence and part (c) (the §9.3 fork-of-the-replaced wrinkle) was not examined against the new lookup; it stays `[INFERRED from the code]` above. The fix taken makes (c) moot (no second row is ever written), so nothing rests on it.

## 7. Fix BUILT 2026-08-19 (TL-047) — what shipped in the commit

`create_tool_component_regenerate.go` (the arm + `replaceExistingRequested`), the Optional entry + a 6-line probe branch in `create_tool_component_action.go`, `create_tool_component_regenerate_test.go` (5 arms; two mutation-proven: deleting the slot UPDATE breaks the walk at COMMIT, deleting the zero-placement guard turns the typed refusal into an unexpected UPDATE), `single_slot_floors.go` scope-table correction + `page_component_writer_coverage_test.go` declared exemption with reason, `check.py` `OPTIONAL_KEY_COUNTS["create_tool_component"]` 3→4 (parity test green; **re-apply the overlay after the roll or the daily cron keeps the old literal**), seed `496_…_HOLD.sql`, RFC_036 §12, register TL-047. Full `platform/orchestration/actions` suite green on archive-HEAD + these files. Council `7a82c943` submitted with the rationale naming the write history, the race, the §9.3 interaction and the section writer's convention; consumers = 1 step; RFC_022 Optional 3→4.

## Relations
`bugs_closed/286` (page half; TL-044) · `architecture_review/RFC_036` §9.3 (library half, live) + §11/§12 (conventions) · `bugs_open/311` / CLC-020 (section writer's in-place regeneration + `resolveStorageIdentity`) · CTS-009 · LANDMINES "already exists probe ignores build_status" · lane: `docs024_key_docs_latest/webdesign_tool_rebuilds/` (RUNBOOK §"Before REFILING", §"The retire race"; NOTES 2026-08-18 13:51Z).

## 8. Pattern-check advisories on the fix commit `d375a0801`, answered honestly

- **`unrepaired-component-write`** (both `create_tool_component_action.go` and the new arm): the regenerate arm writes `page_components.rendered_html` without `repairComponentHTMLBeforePersist` — **exactly the status of this action's own creation link INSERT**, which the checker reports by design: `scripts/pattern-check.py`'s `COMPONENT_WRITE_ALLOWED` says *"the tool-markup writers are deliberately absent"* so the detector keeps naming them (`bugs_closed/136` names them as the next candidate; `bugs_closed/180` is why it is not trivial — the repair rewrote JavaScript that BUILDS anchors, which is what tool markup is). NOT allow-listed here; the arm inherits the open exposure and this line is the "say so in the bug file". A tool regeneration carries the same kind of markup the creation path already persists unrepaired; it adds no new class of link.
- **`handrolled-shipped-predicate`** at the arm's `SET build_status = 'deployed'`: a WRITE to `page_components`, not a `pages` liveness test — the same shape `fix_component_template_action.go` is allowed for; added to `SHIPPED_PREDICATE_ALLOWED` with that reason (follow-up commit).
- **`architecture signal` (migration + platform code in one commit)**: the seed is `_HOLD` and documents the order (image before config); the change is an opt-in field with the unsafe default OFF and one named consumer — the 2026-08-11 RFC_022 ruling's shape, not architecture-scope. Recorded in RFC_036 §12 regardless, as the tool writer's third collision convention.

## 9. Council round 1 (`7a82c943`, 21:19Z): REVISE — gating objection from `bug_historian`, and it is RIGHT

**Objection (high, edit 2):** the arm writes `req.htmlContent` verbatim into `html_template` AND the slot's `rendered_html` and flips `deployed` with **no check that the incoming markup is non-hollow** before overwriting a WORKING tool — the exact shape of `bugs_closed/012` (improver truncates and destroys what it repaired) and `bugs_closed/056` (regeneration silently drops content), and 016b §9's "the guard's discriminator is EMPTY-vs-content". Deferring that to "286's related finding" fixes the duplicate-key symptom while leaving the silent-loss mechanism live on the one path that overwrites in place. **(medium, edit 4):** the floor exemption opts out entirely instead of narrowing — no substitute content check. Advisory (low): edits 4/5 were one hunk listed twice.

**Accepted — and BUILT 2026-08-20 (next section). The plan as specified at the time:**
1. In `regenerateToolComponentInPlace`, BEFORE the snapshot/tx: a **non-hollow gate** on the incoming template — visible text (tags stripped, `<script>`/`<style>` bodies removed, whitespace collapsed) must be ≥ a floor, and the incoming must not be SHORTER in visible text than a fraction of the incumbent's (the existing text-shrink floor's own discriminator, reused via `evaluateSectionShrink`/`countComponentClasses`-style helpers from `single_slot_floors.go` — do not write a third text counter). Refuse loudly (`replace_existing refused: hollow regeneration …`), record via `recordComponentWriteRejection` with a typed error_code, leave the incumbent untouched. The ab-test hollow shell (13,284 chars, ZERO visible text) is the fixture that must FAIL the gate; `adoptTestToolHTML` must pass it.
2. Replace the floor EXEMPTION with this as the **substitute guard** (narrowed for the component_swap shape: classes may change, text may not vanish) — update `page_component_writer_coverage_test.go`'s entry and `single_slot_floors.go`'s scope row accordingly.
3. Tests: arm F — hollow incoming ⇒ refused before any write (mutation: delete the gate ⇒ the walk reaches the snapshot); arm B unchanged.
4. Submission: merge edits 4+5 into one; name the gate as the answer to 012/056.
The code on HEAD (`d375a0801`) is inert until a roll and behind the HOLD seed, so nothing is exposed in production by this gap; the HOLD stays until round 2 is approved and built.

## 10. Revise BUILT 2026-08-20; round 2 SUBMITTED on the same correlation; v1.0.1320 state

- **The non-hollow gate is in the arm** (`create_tool_component_regenerate.go`, before any write):
  ABSOLUTE — `visibleTextLength(incoming) == 0` refuses unconditionally (no config off-switch; a tool
  with no visible text has no UI); RELATIVE — the shared `evaluateSectionShrink` on the same calibrated
  axis, config key (`sectionShrinkFloorKey`, 0 disables) and minimum (`minShrinkGuardVisibleChars`) as
  both existing callers, incumbent template vs incoming. Refusal is loud and typed
  (`recordComponentWriteRejection`, error_code `tool_regeneration_hollow_blocked`), incumbent untouched.
  `shrink_axis_coverage_test` holds the new caller to the axis mechanically.
- **The fixture that must FAIL is the objection's own artefact:** `hollowToolHTML` (arm F) is
  byte-for-byte what the shared test fixture USED to be — zero visible text — which is itself evidence
  the objection was right; the shared fixture now carries visible text like a real tool. Arm G pins the
  relative half (incumbent ≥200 visible chars, incoming <50% ⇒ refused).
- **Mutation run 2026-08-20:** gate deleted ⇒ arms F and G fail on the walk reaching the placement
  census unexpectedly ⇒ restored. Round 1's two mutations still hold. Full actions suite green on
  archive-HEAD; exemption reworded as a NARROWING (class floor not applied — component_swap shape;
  text axis enforced in the arm) in both the coverage test and `single_slot_floors.go`'s scope table
  (now `SUBSTITUTE`, not `EXEMPT`).
- **Round 2 submitted** on the SAME correlation (`RESUBMIT_CORR`), run orch `ed2c500e-a127-478c-903c-789e5d48145b`;
  edits 4+5 merged per the advisory. Verdict pending at time of writing.
- **Roll state:** `v1.0.1320` (pods 16:09Z 2026-08-20, stamp `a255551e0` probed with a junk control)
  CARRIES round 1's code (`d375a0801` is an ancestor) — **inert**: the flag is unmapped (seed 496 still
  `_HOLD`) and the default-OFF path is byte-identical, so the round-1 gap was never reachable in
  production. **Order to close:** round 2 APPROVED → a roll whose stamp's ancestry includes the GATE
  commit (not just `d375a0801`) → lift + apply seed 496 → the live re-fix in §5.

## 11. Round 2 (21:56Z 08-20): REVISE again — and the gating objection was aimed at a STALE LANDMINE, not the code; round 3 submitted with measurements

Round 2's gating objection (editquality, high): the gate's premise contradicts the LANDMINE "the
per-slot TEXT floor counts CSS and JAVASCRIPT as text". **Reconciled, not conceded:** that landmine
describes the RETIRED tag-stripped axis (`shrinkGuardTagStripper`); its own "the check" line
prescribes `datahelpers.VisibleTextFromHTML` — the parsed walk `visibleTextLength` delegates to and
the gate uses; both `evaluateSectionShrink` callers moved to it 2026-08-17 (`bugs_open/293`), the
coverage test holds every caller to it, and arm F is the mechanical proof (fixture's only content
inside `<script>` measures 0). **The landmine headline was corrected in place** (dated, with the cost
— it misled the council's own seat) and re-synced via `landmines-verify-dispatch.sh`. The remaining
seats' asks, each answered with a run measurement in the round-3 rationale: fixture consumers = 2
sites, no shape assertions; shared helpers diff = 0 lines (read-only callers); the census re-run over
the WHOLE config document (sub_workflow-visible) = 1 row; round 1's report row quoted with its query
(the prior_art seat could not find it). Operations corrected to `modify` (the round-1 file is at HEAD).
**Round 3 = round 2's code byte-identical**, run orch `4dd5bea8-524d-4ef1-9f5e-21f176b78579`; verdict
pending. Note for whoever reads the trail: two of three rounds' objections found real defects (round
1: the missing hollow gate — real; round 2: a stale landmine — real, but in the DOCS, and the fix was
correcting the landmine, not the code).

## 12. Round 3 (23:5xZ 08-20): REVISE — prior_art_librarian, two HIGHs, both "true claim, no attached evidence"; round 4 submitted with the code IN the submission

Round 3 gated on exactly two things, both evidentiary: (a) the claim that `update_component_html`
does not write `rendered_html` and flips placements fleet-wide to `pending` (the reason it cannot be
the tool path) had no attached code; (b) the visible-text-axis claims (both callers moved 08-17, the
coverage test enforces it, the 117-pair calibration) were "specific, checkable, unverified". Both are
true at HEAD; the seat's index is declarations-only and could not see them. **Round 4 = the same code,
with the evidence quoted verbatim into `grounded_in`** (EVIDENCE-1…5: the whole update_component_html
write block including its own "Do NOT set rendered_html here" comment and the unfiltered
`WHERE component_id = $1`; the floor maps built with `visibleTextLength`; the coverage test's regex,
failure message and caller-count assertion; the calibration header with both ab-test hollow pairs at
visible 684→0; and both tool-slot writers binding the TEMPLATE into `rendered_html`). Run orch
`ec04b29b-43b8-4f3b-9a14-87dbd175920e`; verdict pending. The round tally so far: r1 caught a real code
gap (the hollow gate), r2 caught a real DOCS defect (a stale landmine headline), r3–r4 are evidence
formatting for a seat that cannot read HEAD — each round cheaper than the one before.
