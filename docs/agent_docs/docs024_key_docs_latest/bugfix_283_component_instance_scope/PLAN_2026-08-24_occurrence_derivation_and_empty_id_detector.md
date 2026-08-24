<!--
PROVENANCE: initial plan authored 2026-08-24 by a Fable 5 Plan agent at the owner's direction
(decisions 1+2 of 2026-08-24: build the occurrence-0 fix NOW, and fold the empty-id
detector/fail-loud fix into the SAME change). The agent explored the live repo and DB; its
three most load-bearing claims were independently re-verified by the supervising session
before this file was written (the log-only check at component_library.go:1103-1110; the
pattern-check seam regex matching both binder names; 30 multi-instance pairs / 16 with
repeated slot_names, live DB 2026-08-24). Re-run the dated measurements on build day — §
"Open questions" item 6 says which ones move.

THIS IS THE STARTING PLAN for the building thread, not a finished spec: the owner said
"prepare an initial plan and we can build that plan up in another thread." Corrections and
resizings land HERE, marked as corrections, per the standing-five practice.

Related: architecture_review/RFC_032…md §§8-9 (ruling + evidence) · council correlation
e8c7414c (retirement trail; round 3 cites this plan) · bugs_closed/283_CONTINUE_HERE_2026-08-24.md
-->

# PLAN — RFC_032 step 3: single-section renders derive the real occurrence, and the detector/seam stop being blind to empty ids

## Context

Two render paths that see one section at a time (`RenderComponentAction`, the section editor) stamp occurrence 0 on every instance via `BindSingleSectionInstanceToken`, so ANY per-section operation (page build, `content_rewrite`, section edit) re-collides every multi-instance page it touches — reproduced live 2026-08-23, re-confirmed at the artefact 2026-08-24 (`gaswholesalers.com/pricing-transparency.html` and `vetcomparison.uk/how-it-works.html` each serve `id="c-generic-text-block"` twice). The ruled fix (RFC_032 §9c) is to feed the canonical rule its real input — the occurrence counted from `page_components` — with an occurrence-0 fallback only where rows genuinely do not exist; the same change also closes two detector/seam blind spots around empty ids (Defect B), because a wrong occurrence is only detectable if the detector can actually see every broken id. One change, two halves, council-reviewed, staged in commits.

Measured this session (2026-08-24, live DB unless stated): 30 multi-instance (page, function) pairs on 30 pages; **16 of 30 have repeated slot_names** (slot alone cannot identify a placement — this kills any slot-only key and shapes Half A); `page_components.position` is `integer NOT NULL`, 1-based (`save_page_sections_action.go` writes `i+1`); 2 live (page_id, position) ties, **both cross-function** (tie-break unobservable today); 6 non-removed rows on 6 pages carry `id=""` in stored `rendered_html` (the RFC's "4 idea.uk pages" has grown — and the two extra are a THIRD cause: `category-listing`'s `id="{{.category_slug}}"`, a content field, rendering empty on dartsonline.com); `enforce_instance_scope` appears in live config only on `tool-generator` and `tool-deployer` (the rerender sweep is NOT armed anywhere).

---

## Half A — derive the occurrence from page_components

### A1. The canonical rule's exact semantics (verified, must be matched)

`InstanceCounter.Next` (component_instance_scope.go:140) keys on `strings.ToLower(strings.TrimSpace(function))`. The canonical walk is `loadStoredSections` (rerender_page_sections_action.go:1198): rows `WHERE page_id=$1 AND build_status IS DISTINCT FROM 'removed' ORDER BY position ASC`, and `Next` is called **only for sections that resolved a component** (carried sections — component missing, template invalid/empty, plan not ready — skip it; comment at :646-653). Equivalence claim for the SQL below: `lower(btrim(cc.function)) = lower(btrim($fn))` is exactly the counter's key equality; strict position ordering matches `ORDER BY position ASC`; the carried-section divergence is documented in A5 (it errs toward a HIGHER occurrence, i.e. a distinct token, never a collision).

### A2. New code — all in the seam's home

New file `platform/orchestration/actions/component_instance_occurrence.go` (keeps `component_instance_scope.go` pure; same package, `component_instance_*` family). Contents:

```go
// SectionPlacement is what a single-section render path can know about WHICH
// placement it is rendering.
type SectionPlacement struct {
    PageID       uuid.UUID // uuid.Nil = unknown -> occurrence 0 (today's behaviour)
    Position     int       // >0 = known (1-based, the editor path); <=0 = unknown
    RowID        uuid.UUID // tie-break identity when Position known; uuid.Nil ok
    SlotName     string    // used when Position unknown
    SameSlotRank int       // k-th same-slot item this pass (0 outside a loop)
}

// StoredOccurrence returns the occurrence this placement takes under the
// canonical position-order walk, derived=false when it cannot be derived.
// NEVER returns an error to be acted on by refusing a render.
func StoredOccurrence(ctx context.Context, db *sql.DB, function string, p SectionPlacement) (occ int, derived bool, err error)

// PlacementFromRenderStep extracts a SectionPlacement from a render_component
// step: page id via config "page_id_from" (ExtractNestedFieldString, with the
// unified extractor's input_data fallback), slot via "slot_name_from", and
// SameSlotRank from the loop-injected "loop_item_index"/"loop_name" config keys
// plus the "<loop_name>_item_<i>" entries the loop expander stores in
// CollectedData (loop_expansion_handler.go:153-156, :202-203).
func PlacementFromRenderStep(config, collected map[string]interface{}) SectionPlacement

// BindSingleSectionInstanceToken — SAME NAME, NEW SIGNATURE. Moves here from
// component_instance_scope.go. Keeping the name keeps pattern-check.py's
// INSTANCE_BIND_SEAM_RE (scripts/pattern-check.py:667) matching every call
// site, and the signature change makes the compiler find all three callers.
func BindSingleSectionInstanceToken(ctx context.Context, db *sql.DB,
    rc *RenderContext, function string, p SectionPlacement, logger *zap.Logger)
```

The binder internally: derive → `BindInstanceToken(rc, InstanceToken(function, occ))`; every failure branch binds occurrence 0 (never worse than today) with a named log. Rewrite the old doc comment: the "every interactive component appears once per page" licence was measured about getElementById components and was never true of the RFC_032 §8 templates (§9b's finding) — say so.

### A3. The two SQL shapes

Position known (section editor):

```sql
SELECT count(*)
FROM page_components pc
JOIN content_components cc ON cc.id = pc.component_id
WHERE pc.page_id = $1
  AND pc.build_status IS DISTINCT FROM 'removed'
  AND lower(btrim(cc.function)) = lower(btrim($2))
  AND (pc.position < $3 OR (pc.position = $3 AND pc.id < $4))
```

Position unknown — slot + same-slot rank (`RenderComponentAction`):

```sql
WITH live AS (
  SELECT pc.position, pc.id, pc.slot_name, lower(btrim(cc.function)) AS fn
  FROM page_components pc
  JOIN content_components cc ON cc.id = pc.component_id
  WHERE pc.page_id = $1 AND pc.build_status IS DISTINCT FROM 'removed'
), target AS (
  SELECT position, id FROM live WHERE slot_name = $2
  ORDER BY position, id OFFSET $3 LIMIT 1        -- $3 = SameSlotRank
)
SELECT (SELECT count(*) FROM live l
        WHERE l.fn = lower(btrim($4))
          AND (l.position < t.position OR (l.position = t.position AND l.id < t.id)))
FROM target t
```

Zero target rows → `derived=false`. The rank disambiguation is what handles the 16/30 repeated-slot pairs: the loop's k-th same-slot iteration maps to the k-th same-slot stored row in position order (composition is preserved on `content_rewrite`, which is the live defect's route).

Also tighten `loadStoredSections`'s `ORDER BY position ASC` to `ORDER BY position ASC, id ASC` (rerender_page_sections_action.go:1209) so the canonical walk is deterministic and matches the `(position, id)` comparison. Blast radius: only the 2 live tie pages (both cross-function) could see section-order change, and their current tie order is Postgres-arbitrary anyway. Flag it in the council submission.

### A4. Call-site edits (contested files, minimal lines)

- `v3_site_actions.go:2404` — replace the call and shrink the long occurrence-0 comment to ~4 lines pointing at the seam file:
  ```go
  BindSingleSectionInstanceToken(ctx, params.DB, renderCtx, comp.Function,
      PlacementFromRenderStep(config, params.CollectedData), params.Logger)
  ```
  (~10 changed lines total in the heavily contested file; everything else lives in the new seam file.)
  **Include in the SAME edit: delete the retired `renderCtx.ContentData["ComponentID"] = comp.ID`
  binding at :2385, twenty lines above this call site** — the third and last binding of RFC_032 §8's
  retirement, twice deferred because this contested file kept carrying other sessions' uncommitted
  hunks (2026-08-23: the 357 lane's, which became a council-veto passenger; 2026-08-24: the 345
  lane's, left uncommitted at their lane's close). It is inert (census 2026-08-24: 0 templates spell
  the placeholder) and belongs in this change because this change edits the same function anyway —
  one reviewed diff, no passenger risk that isn't already being managed for the call-site rewire.
- `section_editor_actions.go:1104` (applyContentEdit) and `:1275` (applyComponentSwap) — build `SectionPlacement{PageID: <pcData "page_id">, Position: <pcData "position">, RowID: <pcData "id">}` and call. `pcData` carries all three (`LoadEditContextAction`, section_editor_actions.go:206-214). Handle `position` arriving as `float64` after workflow-state persistence. For the swap, `function` is the NEW component's — the count-predecessors query is correct for it by construction.
- Delete the old two-line body from `component_instance_scope.go` (function moves); no pattern-check change needed since the name survives.
- Workflow config (DB, applied post-roll — a numbered file under `docs/agent_docs/sql_for_agents/`): add `"page_id_from": "input_data.current_page.id"` to the `render_section` AND `render_from_template` step configs of the live `page-content-writer` agent_definitions row. `slot_name_from: "current_section.name"` is already present on `render_section` (verified live); add it to `render_from_template` if absent. Include a rollback SQL. Until this config lands, the v3 path binds occurrence 0 exactly as today — the wiring is the opt-in.

### A5. Fallback semantics and the honest blind spots

| condition | occurrence | log |
|---|---|---|
| PageID unknown (no config / path unresolved / non-loop caller such as tool recreation) | 0 | Debug "no page context — occurrence 0 as before" |
| query error | 0 | Warn — the lookup is an input-improver, never a gate; a render must not fail on it |
| no row matches slot, or rank beyond stored rows | 0 | Info "placement not stored yet — build in progress or new section" |
| derived | count | Debug with occurrence and source |

Blind spots, stated rather than assumed away:
1. **First build of a new page**: no rows exist → every instance gets 0 → a NEW multi-instance page still collides on its first build. Corrected by its first full-page rerender; recorded meanwhile by the (now widened) sweep at rerender_page_sections_action.go:789. This is §9c's irreducible residue.
2. **Composition/ready-set drift on a rebuild**: the rank counts among THIS build's ready sections; stored rows reflect the previous build. Where they differ around a repeated slot (e.g. apis.uk's 7-of-8 shortfall page if the missing section is a repeated one), the derived occurrence can be off by one. Never worse than the constant-0 status quo; the next canonical rerender restores agreement.
3. **Repeated slot outside a loop**: SameSlotRank=0 → resolves to the first same-slot placement. Editor path is immune (position-exact).
4. **Concurrency**: nothing spans count→save in one transaction; a concurrent full-page render can shift rows between our read and our write. Last save wins per row; divergence is transient; `DetectInstanceCollisions` at the next assembly remains the backstop (unchanged role).
5. **Carried-section divergence**: the canonical counter does not advance for carried sections; the SQL counts every stored same-function row that joins to a component (it cannot see template validity). Where an earlier same-function section is unresolvable, the SQL yields a HIGHER occurrence — a distinct token (no collision), byte-drift versus the canonical walk until the component repairs. Errs-safe direction; documented in the function's comment.
6. Rows with NULL/empty `slot_name` are invisible to the slot variant → fallback 0 → same as today, never worse.

Explicitly rejected inputs (do not re-litigate): slot-name-derived tokens (council 2026-08-16, re-refuted by §8.4 and by this session's 16/30 measurement); reconstructing the occurrence purely from loop items without the DB (contradicts the ruled shape and couples an action to loop-expander internals — noted as an open question only for the first-build residue).

## Half B — detector sees empty ids; the seam fails loud on an unbound identity binding

### B1. Detector widening (`component_instance_scope.go`)

Do NOT widen `reElementID` (`([^"{}]+)`, :215) — folding `""` into `DuplicateElementIDs` would only fire at ≥2, would print an empty string in `Summary()`, and would mis-diagnose (empty = binding failure; duplicate = occurrence failure). Instead add a class:

```go
var reEmptyElementID = regexp.MustCompile(`\sid=(?:""|'')`)
// InstanceCollisions gains:
EmptyElementIDs int // id attributes that rendered literally empty — one is already a defect
```
`Clean()` gains `&& c.EmptyElementIDs == 0`; `Summary()` gains `"N empty element id(s) — an id binding rendered empty"`. An unrendered `id="{{.X}}"` matches neither regex (braces) — the "unrendered template is not evidence" property is preserved and pinned by test. The converter's own harvest regexes (`component_instance_conversion.go`, `component_instance_bindings.go`) are separate copies — untouched.

Caller-by-caller behaviour change (all three non-test callers enumerated):

| caller | today | after |
|---|---|---|
| rerender sweep, `rerender_page_sections_action.go:789` (`!Clean()`) | empty ids read CLEAN | recorded in `out["instance_collisions"]` + Warn. Refusal only where `enforce_instance_scope` is armed on that step — measured 2026-08-24: armed nowhere on rerender workflows (only tool-generator/tool-deployer configs carry the key), so this is record-only. The 6 stored empty-id pages start being recorded at their next rerender. |
| `GateConvertedTemplate`, `component_instance_conversion.go:432` | cannot see `id=""` (its own `id="-` check covers only token-empty) | add an explicit branch: `report.EmptyElementIDs > 0` → hard error ("an id rendered EMPTY across a real-token render — transform defect"), NOT judged-pool. Inherited by `fix_component_template_action.go` (:1272, :1470, :1613), `component_instance_judged.go:196`, `tool_birth_instance_scope.go:56` (where `enforce_instance_scope` IS armed → tool birth refuses such a template). Pre-roll measurement: run the widened detector over every active template's doubled real render (extend `cmd/instanceaudit`), expect 0 trips — cite the number. |
| `cmd/instanceaudit/main.go:108-109` | prints named fields | add an `EmptyElementIDs` column (report-only tool). |

### B2. Render-time fail-loud

The check ALREADY EXISTS as a log-only Error at `component_library.go:1103-1110` (`TemplateNeedsInstanceID(templateStr)` + empty `ctx.ContentData[InstanceContentKey]`). The change is to arm it — this estate has an owner ruling that a named log is not escalation (cited at component_library.go:1216).

**Failure direction: refuse** — `return "", nil, nil, fmt.Errorf("template namespaces ids with {{.InstanceID}} but no per-instance token is bound — bind via BindInstanceToken/BindSingleSectionInstanceToken (bugs_open/283 seam)")`. Every caller already handles this error channel correctly since bugs_open/260: rerender carries stored HTML, the editor refuses and leaves the live page, the build fails the step visibly. Refusing is the bugs_open/260 class ("output that is structurally wrong for every instance"), unlike the 342 absent-content report which deliberately does not refuse because that content legitimately renders today.

**Authority decision, conditional on a measurement the builder MUST run first** (owner ruling 2026-08-02 §2 — new authority needs opt-in OR measured zero):
1. Log census: the exact Error string at component_library.go:1106 has been shipping since v1.0.1304 — count fleet-wide occurrences over the retention window. Expected 0 since the 283 binder work completed.
2. Static census: pattern-check's `COMPONENT_RENDER_RE` file list — every non-`INSTANCE_TOKEN_ALLOWED` caller binds (the allowed list is chrome/head/lint templates that cannot carry `{{.InstanceID}}`).
3. The 6 stored empty-id pages are STORED bytes, never re-rendered by themselves — they cannot trip a render-time check.

If (1) is zero → ship the refusal unconditional, citing all three numbers in the council submission. If non-zero → keep log-only, fix the discovered unbound caller(s) in this same change, re-measure, and only then arm (do NOT invent a RenderContext opt-in bool — a per-caller flag re-creates the per-call-site wiring this seam exists to remove).

Division of labour, stated: the seam refusal covers the deterministic identity-binding case; the widened detector covers every OTHER cause of an empty id (e.g. `id="{{.SomeUnboundField}}"` as a whole value — the live dartsonline `category_slug` case) at gate/sweep time. No new ctx publication field and no new work-item type — the sweep's recorded output is the queue-adjacent surface (trade-off stated for the council).

## Test plan (all mutation-proven; patterns from component_instance_scope_test.go and v3_render_slot_name_test.go, sqlmock per v3_render_slot_name_test.go's harness)

New `component_instance_occurrence_test.go`:
- `TestStoredOccurrence_countsSameFunctionPredecessors` — rows [hero, gtb, gtb, faq, gtb], target 3rd gtb by position → occ 2. Revert to constant 0 → fails (0≠2): pins the defect fix itself.
- `TestStoredOccurrence_equivalentToInstanceCounter` — for a fixture page list incl. case/whitespace-differing functions ("FAQ " vs "faq"), tokens via the derived occurrence == `InstanceTokensForPage` output. Breaking the lower/trim predicate → fails: pins the key-equality claim.
- `TestStoredOccurrence_slotRankDisambiguatesRepeatedSlots` — two rows sharing slot_name+function, ranks 0/1 → occ 0/1. Dropping the OFFSET → both 0 → fails: pins the 16/30 repeated-slot case.
- `TestStoredOccurrence_fallsBackToZeroNeverErrors` — no rows / query error → occ 0, derived=false, render unaffected. Making the binder propagate the error → fails: pins "input-improver, not gate".
- `TestPlacementFromRenderStep_readsLoopPrefix` — CollectedData with `<loop>_item_0..3`, `loop_item_index=3`, two prior same-slot items → SameSlotRank 2. Breaking the item-key format → fails.
- `TestSingleSectionBinder_twoInstancesGetDistinctTokens_endToEnd` — drive `RenderComponentAction` twice via the sqlmock harness (renderSlotParams pattern) as a content_rewrite would, page rows mocked → assembled output passes `DetectInstanceCollisions`; control: no rows mocked → both occurrence 0 → detector reports. Reverting the v3 call site → control passes but the main assertion fails: proves the wiring, not just the helper.
- `TestSectionEditorBinder_usesStoredPosition` — position-known branch incl. float64 position; swap case counts the NEW function.

Half B, in `component_instance_scope_test.go` + `render_seam_no_fallback_test.go`:
- `TestDetect_emptyIdIsItsOwnClass` — one `id=""` → EmptyElementIDs 1, Clean false, DuplicateElementIDs empty (class separation); `id=''` counted; `id="{{.X}}"` NOT counted (unrendered-template control). Revert regex → fails.
- `TestGate_refusesEmptyIdResidue` — converted template with bound `{{.InstanceID}}` plus `<div id="{{.Missing}}">` → gate hard error, `needsJudged=false`. Remove the gate branch → fails.
- `TestRenderTemplate_refusesUnboundInstanceToken` — needs-token template + empty binding → error naming the seam; bound → renders. Revert to log-only → fails.
- UPDATE `TestRenderLayer_twoInstancesOnOnePageGetDifferentIDs` — its unbound-render mutation control (currently `mustRender` with `&RenderContext{}`, component_instance_scope_test.go:233-240) becomes "must refuse"; keep a collision control on hand-built HTML so the detector half stays mutation-proven.

Full suite run; any other test rendering an InstanceID template unbound will fail loudly — that is the property, not collateral.

## Rollout order

1. **Commits** (one reviewed change, staged): (1) Half A — occurrence derivation, call-site rewires, `loadStoredSections` tie-break, tests, the sql_for_agents config file (not yet applied); (2) Half B — detector class, gate branch, seam refusal (per the measured branch), instanceaudit column, tests; (3) docs — RFC_032 §10 note, council submission with this session's dated measurements (30 pairs / 16 repeated-slot / 2 cross-function ties / 6 empty-id pages / repro pages ×2 tokens / arming census).
2. **Council round** over the whole change. Name explicitly: the tie-break's 2-page blast radius, the gate's new hard-error class, the refusal's measured-zero (or the fallback taken), and the first-build residue.
3. **Image roll** (Go inert until then). Post-roll the editor path is live; the v3 path still binds 0.
4. **Apply the workflow-config SQL** (`page_id_from`, and `slot_name_from` on `render_from_template`) — DB config is live immediately; this is the activation switch for the build/rewrite path.
5. **Re-repair**: file `page_rerender` items for the 30 multi-instance pages (the queue that fixed 9/12 on 2026-08-23 is canonical). `apis.uk/index.html` rebuilds itself from its standing `needs_rebuild` — the build-path proof, no setup.
6. **Verify at the artefact, never at a status**:
   - `curl -s https://gaswholesalers.com/pricing-transparency.html | grep -o 'id="c-generic-text-block[^"]*"' | sort | uniq -c` → expect bare token ×1 and `-2` ×1 (today: bare ×2, re-measured 2026-08-24). Same for `vetcomparison.uk/how-it-works.html`.
   - `apis.uk/index.html` post-rebuild → 6 distinct tokens.
   - **Stickiness proof (the actual defect)**: after repair, trigger or await one `content_rewrite` on a repaired page and re-count distinct tokens at the served page — this is what flapped on 2026-08-23 17:41.
   - DB sweep: per multi-instance page, extract `id="c-…"` tokens from non-removed `rendered_html` and assert per-page distinctness.
7. Out of scope, named as unblocked follow-ups: arming `enforce_instance_scope` on the rerender workflow (RFC_032 §8.5's config migration), retiring ~~the `pricing` template's placeholder~~ **(done 2026-08-23 — `SQL_2026-08-23b`; the follow-up that remains is the THIRD binding at v3_site_actions.go:2385, blocked on that file carrying other sessions' WIP)**.

## Open questions for the building session

1. Does `"input_data.current_page.id"` resolve inside the page-content-writer workflow, or does the unified extractor's `input_data.` fallback make `"current_page.id"` the right spelling? Read `datahelpers/unified_extractor.go` and confirm against one live run.
2. Do `sections_for_render.sections_ready` items carry the slot under `"name"` after `resolve_internal_links_action.go:259` rebuilds the list? (Needed by the SameSlotRank count.)
3. Are there OTHER live workflows with `render_component` steps that should get the config keys? (Query `agent_definitions.default_config` for `"action": "render_component"` fleet-wide.)
4. The log-census result for "no per-instance token was bound" — decides refusal-unconditional vs fix-then-arm (B2).
5. Should the slot match apply `NormalizeComponentFunction` to both sides? Verify one live page's `slot_name` vs `current_section.name` byte-equality first.
6. Re-measure the empty-id census and the 30-page list on build day; both moved between 2026-08-22 and 2026-08-24.

## Critical files

- `platform/orchestration/actions/component_instance_scope.go` (seam home: token rule, detector, the binder being replaced; new sibling `component_instance_occurrence.go` lands beside it)
- `platform/orchestration/actions/v3_site_actions.go` (RenderComponentAction call site, :2384-2404 — **contested file**, minimal edit)
- `platform/orchestration/actions/section_editor_actions.go` (applyContentEdit :1104, applyComponentSwap :1275, pcData shape :206-214)
- `platform/orchestration/actions/component_library.go` (RenderTemplate seam, :1103-1110 log-to-refusal)
- `platform/orchestration/actions/component_instance_conversion.go` (GateConvertedTemplate :405-454, new empty-id branch)
