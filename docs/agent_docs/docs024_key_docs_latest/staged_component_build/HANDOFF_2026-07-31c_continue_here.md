# HANDOFF — staged_component_build, 2026-07-31c (fresh chat starts here)

**This supersedes `HANDOFF_2026-07-31b_continue_here.md`.** That file's own open item (the Go
gate proof, its §3) is DONE — read this file instead; only come back to `b` if you want the
proof's own detail, which also lives in `SUMMARY_2026-07-31b_the_go_gate_is_proven.md` and
`NOTES_staged_component_build.md`.

---

## 1. Read these, in this order

| doc | why |
|---|---|
| `README_where_we_are.md`, last four entries | plain prose, fastest way in |
| `SUMMARY_2026-07-31b_the_go_gate_is_proven.md` | current-state milestone read-out for the just-finished piece |
| `NOTES_staged_component_build.md`, last two entries (search `## 2026-07-31, later still`) | the reconnaissance for the NEXT piece — read this before writing any code |
| `PLAN_2026-07-30_staged_component_build.md`, sections **P1** and **P2** | P1's tail is not yet finished — see §2 below, this is easy to misread from `b`'s own open-items list |
| `RUNBOOK_staged_component_build.md` §8–§9 | how to author and prove a criteria fence before publishing it |

Register: **DOC-068** (`subject_type='component'`, both halves now proven live),
**DOC-070** (the probe instrument), **TL-036** (`try_fence.go` / `prove_fence_can_fail.go`).

---

## 2. State — what is live and proven, what is not

**DONE and proven at the artefact:**

- **The component `subject_type`'s two enforcement points both accept it, proven at runtime**
  (not inferred from a build date) — correlation `8f564028-6fc6-488c-96d2-c2e362b243b2`,
  detail in `NOTES` and `SUMMARY_2026-07-31b_…`. `doc_plans` is back to **0** `component` rows
  — the capability is proven, nothing has used it yet.
- **P1a, the three-way naming contract for tools, is CLOSED** — 30 canonical tool components,
  0 broken.
- **`tool-review-council-simulator`'s 18-check fence is live and cluster-proven** (unrelated to
  components, but the pattern to copy).
- **`bugs_open/157` is CLOSED** (another lane, `browser-runner-adapter v1.0.1216`):
  `has_visible_area` now works. It was deliberately left out of every fence this lane built
  while 157 was open — that restriction is lifted. Use it in any new fence.

**Chassis image has moved since the Go-gate proof ran, for reasons unconnected to this
lane** — currently `v1.0.1219` (both replicas), following unrelated fixes
(`bugs_open/165`, `138`, `137`). Re-check the running image before citing any figure from
before 2026-07-31 evening as current; the `doc_subjects_common.go` vocabulary itself was not
touched by any of those.

**NOT yet done — read carefully, `b`'s "open items" list skips a step:**

`PLAN_2026-07-30…`'s own P1 is not finished. Its stated tail, in order, is: *verdict → image
→ pod-grep → apply migration 273 → **then one real component (`teaser-reveal-panel`) gets a
PLAN with a criteria fence, and its NOTES backfilled from
`brochure_component_library/NOTES_brochure_component_library.md`**. Gate: the fence exists,
passes the ten-rule validator, and every criterion has been watched to pass by hand.* Only
verdict/image/pod-grep/migration are done. **No component has a real fence yet** — the only
`doc_plans` row ever written for `subject_type='component'` was the throwaway probe body,
already deleted.

`b`'s "Open items, ordered" §4 calls the next step "S6 for components" and describes it as
wiring a dispatch. That is **P2**, and P2 depends on P1's unfinished tail — you cannot
dispatch a fence to `browser-runner-adapter` that does not exist yet. **So THE NEXT ACTION is
in two parts, in this order:**

---

## 3. THE NEXT ACTION, part A — author `teaser-reveal-panel`'s real fence

`teaser-reveal-panel` is a real, live component — `content_components.id =
'22c12251-73aa-4232-bd67-ef9edcfe8061'`, name confirmed live 2026-07-31 evening. It was
chosen in the PLAN "because its history is fully written down" — that history is
`brochure_component_library/NOTES_brochure_component_library.md`, which is where its NOTES
backfill comes from.

**Before writing anything**, find a live placement to test against:
```sql
SELECT pc.page_id, p.site_id, p.url FROM page_components pc
JOIN pages p ON p.id = pc.page_id
WHERE pc.component_id = '22c12251-73aa-4232-bd67-ef9edcfe8061'
ORDER BY p.updated_at DESC;
```
(5 placements across 2 sites, measured 2026-07-31 evening — re-run, this moves.) Pick one
concrete `(site_id, page_id, url)` to develop and prove the fence against.

1. **Author the fence** — a `criteria` JSON block, same shape as the tool fences (RUNBOOK §8),
   scoped to what a *component* can actually assert: presence, visible area (now safe to use,
   157 is closed), any interactive behaviour the component has, no console errors. It is not a
   tool — there is no arithmetic to assert, no "spec.function" — so do not copy a tool fence's
   check list mechanically; read what the component actually does first
   (`content_components.html_template`, `input_schema` for `teaser-reveal-panel`).
2. **Prove it before publishing**, exactly per RUNBOOK §8:
   ```bash
   go run scripts/try_fence.go <fence.json> https://<domain>/<path>
   go run scripts/prove_fence_can_fail.go <fence.json> https://<domain>/<path>
   ```
   Run **both** — RUNBOOK §8 is explicit that (1) alone is not evidence; the same lane's own
   `tool-review-council-simulator` fence went 36/36 green on the first pass while one check
   asserted nothing.
3. **Write the PLAN** into `doc_plans` by hand, per RUNBOOK §9's supersede pattern
   (`idx_doc_plans_current` is a partial unique index — you cannot just insert), then backfill
   its NOTES from `NOTES_brochure_component_library.md`.

Gate for this part: the fence exists, passes the ten-rule validator, and every criterion has
been watched to fail by hand (not just to pass).

---

## 4. THE NEXT ACTION, part B — wire S6 (dispatch the fence to `browser-runner-adapter`)

**This is genuinely smaller construction than `b`'s "wiring, not construction" framing
suggests — read this before starting, it will save you re-discovering it.**

`request_browser_run` (`platform/orchestration/actions/tool_acceptance_actions.go:87-152`,
the action `tool-acceptance-agent`'s `request_run` step calls) resolves the page to test by
reading a `function` string and looking it up **directly against `pages.name`** (lines
95-145): `SELECT url FROM pages WHERE site_id=$1 AND name IN ($2, 'tool-'||$2)`. That encodes
**one function ⇒ exactly one page** — the same identity P1a's naming check polices for tools.

**That identity does not hold for a component, and it is not close.** Checked live 2026-07-31:
`teaser-reveal-panel` is placed via `page_components` on **5 distinct pages across 2 distinct
sites**. `content_components` has no `site_id` at all — components are fleet-shared by design
(RUNBOOK §5 already says this for a different reason). So "the page for this component" is not
a well-formed question the way "the page for this tool" is, and nothing in
`request_browser_run` can express "which of the 5" today.

**The `smart-contrast` pilot, cited by the PLAN as proof S6 "works end to end," is a tool
(11/11 checks asserting real arithmetic) — not a component.** Nothing in this lane's history
shows a component ever going through `request_browser_run`. Worth being precise: the mechanism
is proven for tools, not yet exercised for components.

**Two ways to close the gap — a real design choice, not decided here:**

- **(a) Extend `request_browser_run`** with an opt-in `page_id_field` (+ implicit `site_id`)
  that, when present, bypasses the `pages.name` lookup and resolves instead via
  `page_components.component_id = <uuid> AND page_id = <given>`. Less code, one action's config
  surface; adds a branch to a function that has exactly one path today.
- **(b) A sibling action** (e.g. `request_component_browser_run`) that never touches the
  tool path. Zero blast radius on a working mechanism; near-duplicate plumbing (envelope,
  headers, profile handling — roughly lines 153-230 of the same file, which are already
  identity-agnostic and could be shared via a helper either way).

**Council-gate note:** either option edits `platform/orchestration/actions/`, which is
in-scope for the advisory council gate (CLAUDE.md, "platform seams"). Per the 2026-07-29 owner
ruling, an addition needs architecture review only if it changes what the shared mechanism
*guarantees* — an opt-in field or a new sibling action that nothing calls until a component
agent names it does not change `request_browser_run`'s existing guarantee for tools, so the
normal council gate looks right. Confirm against the actual diff once one exists; do not take
this paragraph as a ruling on a change that doesn't exist yet.

**Dispatch shape to copy** — `tool-acceptance-agent`'s live workflow (checked 2026-07-31
evening):
```
ensure_site_record → load_docs (load_doc_context, subject_type: tool)
                   → request_run (request_browser_run)
                   → judge (judge_acceptance_results)
                   → complete
```
For a first proof, you likely do **not** need a real `agent_definitions` row at all — the
Go-gate probe in this same lane (`scripts/PROBE_doc_subject_go_gate.sh`) proved that an inline
`config.workflow` in the dispatched message (`selectWorkflow` Priority 1,
`processor.go:922-928`) is enough, with nothing to seed and nothing to clean up afterwards.
Same trick applies here: set `load_docs`'s `subject_type: component`, `subject_key_field`
pointing at `teaser-reveal-panel`, and whichever `request_run` shape you built in the field
above pointing at the one `(site_id, page_id)` you picked in §3.

**Gate:** a deliberately broken `teaser-reveal-panel` (or a deliberately wrong page_id) makes
S6 go red — run the negative control **in the same dispatch**, the same habit the Go-gate
probe used and for the same reason: a green run and a run that skipped silently look identical
unless something is watched to fail.

---

## 5. Open items behind this one, unchanged from `b`

3. **The authoring backlog: 10 tools resolve to a page but have NO PLAN at all.** Honest, not
   a check failure. **Do NOT wire the naming check into the tool-birth path** — tools born via
   `create_tool_component_action` are already born compliant.
4. **`features_open/028`** (rename orphaning) — filed, unowned.
5. **`has_visible_area` checks are owed to every existing fence**, now that `157` is closed.
   Not this lane's to duplicate.

---

## 6. Do NOT do these (unchanged from `b`, still correct)

- **Do not rebuild the eight-stage ladder.** Owner cut it (D8).
- **Do not take `bugs_open/157`.** It is closed; nothing left to take.
- **Do not fire an acceptance run at the arena tool.** `gauntlet_dead_cta` lane's decision.
- **Do not run `./scripts/migration/run-migrations.sh --apply`** blind — it takes every
  pending file, and some are known-bad. Apply single files by hand + `--record-only`.
- **Do not roll the chassis to ship anything.** Builds come from committed HEAD; your commit
  ships on anyone's next roll regardless. A deliberate roll kills in-flight council rounds and
  imposes a ~300s dispatch blackout.
- **Do not adopt `features_open/015`.** Accepted decomposition stands: 015 = rung vocabulary,
  027 = gate mechanism, 026 = missing instrument.

---

## 7. The one thing to carry into this piece specifically

Every past miss in this lane was a check that reported health it never measured. The new one
found this session is a sibling shape: **a plan that named the next step "wiring" without
having checked whether the thing being wired already assumes an identity the new subject
doesn't have.** `request_browser_run`'s tool-shaped page lookup is not wrong — it has simply
never been asked the question a component asks. Read the code the recommended route actually
calls before trusting a phrase like "wiring, not construction," the same lesson this lane
already logged once this week about a predicted failure string.
