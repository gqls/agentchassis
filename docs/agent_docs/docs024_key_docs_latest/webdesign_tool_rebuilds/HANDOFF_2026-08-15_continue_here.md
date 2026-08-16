# HANDOFF — webdesign tool rebuilds + owner-gate follow-ups. START HERE. Written 2026-08-15 ~22:30Z.

> **SUPERSEDED 2026-08-16 ~10:15Z by `HANDOFF_2026-08-16_continue_here.md`** — read that one. Note its correction: §Next actions 2 below ("matches the spawn→call handshake failure") was WRONG; the pilot died on `pages_site_id_name_key` in `create_tool_component` (bugs_open/286, `WRONG_CALLS.md`).

The owner's session (visual gate → tool defects → 281/285 → rebuild lane) ran long; this is
the continuation point. Read `PLAN_2026-08-15_native_rebuild_of_ported_tools.md` for the
design and `NOTES_…` for the evidence trail. Everything below was verified at the stated time.

## Verified state (2026-08-15 ~22:20Z)

| thing | state |
|---|---|
| fleet binary | `v1.0.1303`, pods up 18:45Z; probe: build sha `5e075a6f9` in `/proc/1/exe`, junk control absent; `d7b2d9994` IS an ancestor ⇒ **281 Track 1 + 285 fence LIVE** |
| mindmap repairs (contrast ×2, usage hint) | **LIVE + verified with controls** at the served page |
| occlusion placeholder removal | **LIVE + verified** ('Uncanny Valley' control present) |
| 285 learn-page restore (other lane's) | serving clean (`portedPageAssetList` absent, control present) |
| shared `ported-page` wrapper | restored (4,664 chars, `{{.body}}`), 115 placements `deployed` |
| ab-test page | fork slot: raw tag REMOVED from stored html; ported slot `removed`; **`page_rerender:owner-gate:tool-ab-test-calculator` queued** |
| aspect-ratio pilot | item complete but **EMPTY RUN** — see next actions |
| audits (4 filed) | css-unit-converter COMPLETE; css-specificity claimed; llm-cost triaged; ab-test FAILED (pod-timeout retries) |

## Next actions, in order

1. **Verify the ab-test rerender at the artefact** once its item completes: served page shows
   exactly ONE calculator (fork), zero `{{.` tags, `<script>` ≥1. `grep -c 'ported-page-section'`
   must be 0. Positive control: 'A/B Test Significance'.
2. **Diagnose the pilot's empty run, then refile.** Orchestration `5ef53886-e97a-4302-b194-401dc78dd290`
   (tool-generator, 18:28Z, 47 s, `final_result` empty, no component created). This matches the
   spawn→call handshake failure (MEMORY: ~half fail; **never cancel/refile pre-diagnosis**).
   Read its `execution_path`/`current_step` and the chassis logs around 18:28–18:29Z. Note the
   two OTHER empty-result generator runs same evening (18:36 `tool-storage-risk-explainer`,
   19:27 `tool-overpayment-priority`, different sites) — if all three are empty the defect is
   in the generator/handshake, not this item. The item key `add_tool_novel_webdesign.co.uk`
   is free again once diagnosed (item is `complete`, dedup ignores it).
3. **The two gap-shape `improve_tool` items** (`audit_fix_webdesign.co.uk_c5baf147…`, `_1c1b186e…`)
   fail `load_tool` on `spec.page_id` null — created between seed 426 (pins loader to
   component+page) and the roll that fixed the producer. A backfill UPDATE keyed on
   `spec->>'component_id'` matched **0 rows** — the spec's key names differ from the
   tool_health shape; read one item's full spec first. New audits (post-roll) emit the correct
   shape, so this is a 2-item cleanup, not a class.
4. **285 close-out is now runnable** (fence live): follow `bugs_open/285` §How to close —
   induce a refusal (synthetic `improve_tool` at the shared component `a7daa5c5…` must draw
   the refusal + a `ported_tool_fix` row, not a write), confirm wrapper still 4,664/`{{.body}}`
   + placements `deployed`, then move to `bugs_closed/` (or hold open on the delivery-shape
   residual their contribution re-scoped). Coordinate with `bugfix_285_shared_template_write`
   — they own the lane; contribute into the bug file.
5. **Then the batch** (PLAN §2): only after the pilot end-to-end (generate → deploy → retire
   ported slot → rerender → artefact grade) has succeeded ONCE. Simple tools first, serial,
   RUNBOOK has the guarded SQL. Rich apps stay excluded (PLAN §3).

## Owed elsewhere (not this lane, do not lose)

- **122 ink lane**: owner PASSED dartsonline visually; webdesign explained (accent-ink has no
  visible surface — cascade correction in that handoff §5). **Widening (its §4 step 5) still
  awaits the owner's explicit per-site Go.** Owed audit check: dartsonline (baseline 17) due
  ~08-18 00:58Z, webdesign (baseline 7) ~08-18 02:59Z, via `site-render-audit-rotation`;
  gate on the rotation stamp moving, then count `contrast_failure` arrivals by `created_at`.
  robot-hands' Monday canary ~08-17 14:54Z. NOTE: that lane now has a `HANDOFF_2026-08-15b`
  (another session) — read it first.
- **Mindmap junk node text**: NOT a site defect — owner's browser localStorage (tool seeds
  "Central Idea"; contentEditable + autosave). Told the owner; nothing to do.

## Traps this session paid for (all in LANDMINES/WRONG_CALLS/MEMORY too)

- A "clean" grep needs its fingerprint taken FROM the bad artefact (my false "checked clean",
  `WRONG_CALLS.md` 2026-08-15) — and a first CDN fetch can serve a stale 9-byte "Not found"
  under HTTP 200; cache-bust before diagnosing.
- The section-editor REGENERATES; ported blobs have NO safe framework editor — the slot
  floors refuse (28→4 / 11→4 class attrs). Surgical guarded `replace()` + `page_rerender`
  (assembly path) is the working repair recipe; archive trigger banks pre-states.
- A "failed" section_edit may have half-applied: the ab-test edit's content landed, its
  reassembly failed — read the artefact, not the status, in BOTH directions.
- The build-dispatch-loop is SERIAL per site; `add_tool`'s per-site item key is a deliberate
  throttle.
