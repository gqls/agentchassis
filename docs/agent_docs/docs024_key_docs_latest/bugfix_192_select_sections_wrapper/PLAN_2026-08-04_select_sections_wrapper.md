# PLAN 2026-08-04 — `bugs_open/192`: the "pass-through" that re-wraps the section plan

**Lane claimed 2026-08-04 ~09:40 BST.** Filed by the `bugfix_154_work_item_routing_columns`
lane while live-verifying `bugs_open/178`; filed explicitly as *"not yet diagnosed"* and
handed off ("a `090` diagnosis run is owed"). This lane took the handoff, ran the `090`
loop, and diagnosed it first-hand in parallel.

---

## The mechanism, confirmed

One cause, four links:

1. `page-build-handler.plan_sections` writes `collected_data.section_plan` = a **flat**
   plan: `{sections_ready, sections_deferred, ready_count, …}`.
2. `page-build-handler.load_current_section_content` (new, `bugs_open/178`'s fix, commit
   `08d0515f3`, seed `sql_for_agents/299_*.sql`) declares **`output_field: section_plan`**
   — deliberately reusing the key so no input_mapping needed changing. But
   `load_current_section_content_action.go` returns a **wrapper**
   `{section_plan, applied, reason|matched}` on **every** return path, including all eight
   it calls "pass-through". `coordinator.go:1859-1861` (`storeActionResult`) stores an
   action's return value **wholesale** under `output_field`, so `section_plan` becomes the
   wrapper on **every page build**, in every mode — not only `edit_live`.
3. `call_content_writer` maps `"section_plan": "section_plan"`, forwarding the wrapper
   verbatim as the writer's `input_data.section_plan`.
4. In `page-content-writer`, **both** of `select_sections`' fallback paths die from that one
   cause:
   - path 2 (`input_data.section_plan.sections_ready`) directly — the plan moved one level
     down, to `input_data.section_plan.section_plan.sections_ready`;
   - path 1 (`resolved_links.response.link_resolution.sections_ready`) *indirectly* —
     `resolve_links`' input_mapping is `"sections?": "input_data.section_plan.sections_ready"`,
     so the link-resolver child is handed **no sections** and returns `sections_ready: null`.
     `resolve_links` also sets `error_step: select_sections`, so a resolver failure lands
     here too.

   `ExtractFieldsAction` (`v3_site_actions.go:4232-4356`) then **omits the target key and
   returns success**, so `sections_for_render = {}`; `process_sections_loop` hard-fails at
   `loop_actions.go:144/751` with `key 'sections_ready' not found at position 1` — an error
   naming the **symptom**, never the cause.

## The three properties that made this invisible

- **`output_field` reuse is a silent overwrite.** A "pass-through plus bookkeeping" return
  shape re-wraps the key it was supposed to preserve, and the wrong result looks exactly
  like the right one: all the data is still there, one level down.
- **`extract_fields` fails by doing nothing.** No path resolved ⇒ key omitted ⇒ success.
- **The loop's error names the missing key, not the failed extraction.** It printed
  `key 'sections_ready' not found`; had it printed the keys actually present
  (`[applied reason section_plan]`) the wrapper would have been visible in the first
  failure.

## Fix, ordered by what closes the door

| # | change | layer | why |
|---|---|---|---|
| 1 | `load_current_section_content` returns **the plan itself** on every path | Go | makes the wrapper unrepresentable at the seam |
| 2 | `extract_fields` gains an **opt-in `required`** list — a listed target that resolves on no path fails the step, naming the cause | Go | closes the **class**: future shape drift dies loudly, at the right step |
| 3 | loop path-miss error lists the keys present at the failing level | Go | the diagnostic that would have solved this in one read |
| 4 | seed: third fallback path on `select_sections` + opt into `required` | DB config | **immediate relief on the running binary**; the shim goes structurally dead once 1 rolls |

Rejected: coordinator auto-unwrap (magic on the widest seam — RFC territory); repointing
configs at the nested path (enshrines the wrapper); a fresh `output_field` (regresses 178 —
the enrichment would never reach the writer).

**`required` is opt-in with the unsafe default OFF**, per the OWNER RULING of 2026-08-02 §2
(new authority on a shared seam ships as a field, not a doc comment). Measured collision
check: exactly **2** `extract_fields` steps fleet-wide (`research-agent/extract_topic`,
`page-content-writer/select_sections`), **neither** carrying a `required` key — so nothing
is shadowed and `research-agent` is untouched.

## Order of application

1. **Seed first, deliberately** — a stated deviation from "image first, then seeds". It is
   safe because the seed **names no action**: the third path is plain data and `required` is
   a config key the running binary provably ignores (`ExtractFieldsAction` reads only
   `fields`, `field_map`, `defaults`). It unblocks page builds on the binary already running.
2. Go edits + register entry in **one pathspec commit** (register in the same commit that
   ships the mechanism, per the 2026-07-28/29 ruling). Council submitted alongside.
3. Image roll takes the Go half; then a cleanup seed removes the shim path.

## Verification that can come out FALSE

- **V1** — run the rewritten action test against the **unmodified** action: it must **FAIL**
  on the wrapper assertions. A test that passes both ways tests nothing.
- **V2** — after the seed alone, a real dispatch reaches `compile_page`, and
  `collected_data.sections_for_render ? 'sections_ready'` is true. Fails loudly if the
  wrapper diagnosis is wrong.
- **V3** — post-roll: `collected_data->'section_plan' ? 'applied'` is **false** and
  `? 'sections_ready'` is **true** on a fresh `page-build-handler` run.
- **V4** — post-roll: `bugs_open/178` not regressed — an `edit_live` rewrite still carries
  `existing_content_html`, now at the **flat** path.

## Blast radius, and who is told

`section_plan` consumers are all inside `page-build-handler` (`plan_sections`,
`check_has_ready_sections`, this step, `call_content_writer`) plus the writer's
`resolve_links`/`select_sections` — every one of them wants the flat shape, so the unwrap
restores each and breaks none. No Go code outside the action's own test reads the wrapper's
`applied`/`matched`/`reason`.

Told, not merely measured (OWNER RULING 2026-07-29 §3): the **178 lane** (owns the action;
its guarantee is unchanged in substance and the pass-through promise gets *stronger*), the
**087 lane** (shares this exact error signature and greps that string — it gains a suffix),
and the **research-agent** owner (the other `extract_fields` consumer; behaviour unchanged,
default OFF).
