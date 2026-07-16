# HANDOFF — empty_sections_loop_integrity: start a fresh chat HERE

*2026-07-16. Self-sufficient cold-start. Read this top to bottom; it is enough
to continue without re-reading the whole workstream. Deeper detail:
`PLAN_…md` (phases), `RUNNING_NOTES_…md` (turn-by-turn + the MISSTEP entry),
`RUNBOOK_…md` (operator procedures). Origin:
`../HANDOFF_2026-07-14_empty_product_sections.md`.*

**Testbed:** robot-hands.com `00ff3af5-dad8-4770-9f70-3edc267a3c92`;
dartsonline.com `5fe8785b-223d-41a3-88ee-c07187622381`.
**DB:** `kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db`
**Chassis at handoff:** `v1.0.1123`.

---

## 1. The immediate next action (this is why you're here)

**`refresh_product_specs` runs but refreshes NOTHING — diagnose and fix it.**

Live run 2026-07-16 (`product-spec-refresher`, robot-hands): orchestration
COMPLETED, but `products: 5, refreshed: 0, failed: 5`, every product
`no_fields_extracted`; `verified_date` unchanged at 2026-07-14.

What's already ruled out (don't redo):
- NOT a scrape failure — status would be `scrape_failed`; **Firecrawl worked**.
- NOT a missing model — `mistral-small3.1:latest` IS on ollama-adapter
  (`/api/tags`), and no "LLM call failed" was logged, so the HTTP call worked.
- So: scrape OK → model call OK → model returned nothing usable.

**Leading hypothesis (UNTESTED — treat as hypothesis, not fact):** the action
truncates page markdown to **6000 chars** before prompting
(`refresh_product_specs_action.go`, `if len(md) > 6000`). Manufacturer pages
front-load nav/cookie/marketing, so the spec table may sit beyond the cut,
leaving the model correctly returning `{}` per my own prompt ("if the text is
not a spec page for this product, return {}"). Weak corroboration only:
sibling `../aaa_fails_to_mend/005_…article_body_root_cause_is_truncation_FIXED.md`
— truncation has bitten this codebase before.

**Do this first — the run is now self-diagnosing.** I added logging (in the
working tree, **needs an image build**) that prints the model's own output on
unparseable content and warns on an empty object with `markdown_chars_sent`.
So:
1. Rebuild the chassis image (`make quick-agent-update` or the owner's flow).
2. Re-run the refresher (RUNBOOK §5d has the exact kcat command).
3. Read the log: `kubectl logs -n ai-persona-system <chassis-pod> | grep refresh_product_specs`
   — it will now say whether the model returned `{}` (→ it never saw the specs
   → truncation/wrong-page) or emitted non-JSON (→ prompt/format problem).
4. If truncation: don't just raise 6000 — better to select the spec-bearing
   region of the markdown (or use Firecrawl's extract/JSON mode) than to shove
   a whole page at a 24B model.

**Verify by evidence, not by plausibility** — see the MISSTEP entry in
RUNNING_NOTES for why that sentence is in this handoff.

---

## 2. What this workstream is (one paragraph)

agentchassis autonomously builds/operates a fleet of content sites. Discovery
checks find defects → `site_work_items` → dispatch loops → handler agents.
robot-hands.com was serving live product pages with **every value empty**,
while the platform had already marked the matching work items **complete** — a
fix loop that closes without fixing. This workstream made the loop honest, then
fixed the pages.

## 3. What is DONE and PROVEN LIVE (do not redo)

- **Loop integrity (Phase 1).** A per-item-type **verifier registry** consulted
  by `CompleteWorkItemAction` before it stamps `complete`; defect-persists →
  routes into the normal attempt machinery. Plus SQL 149: page-build-handler's
  no-op exits now flag `needs_human_review` instead of exiting "successfully".
  **Proven** by re-driving the *original* falsely-completed item — it now stops
  honestly and can never false-complete.
- **`required_fields_missing` check (Phase 2).** Flags components whose
  schema-required LLM fields are empty. **Proven live:** 8 flags on
  robot-hands, **0 on dartsonline** (negative control held).
- **Meta-commentary guard (Phase 3).** Blocks LLM apologies shipping as page
  copy. In the binary; not yet exercised by a real case.
- **robot-hands product pages (Phase 4).** Category-wrong cart furniture
  replaced with a `gripper-spec-sheet` fed by a new **live `query.products`
  resolver** (`resolveProducts`) + 5 real, sourced gripper products
  (Schunk/OnRobot/Robotiq/Zimmer/Festo, each with source_url + verified_date).
  **Live and correct:** both `/entities/gripper-detail.html` and
  `/product-detail.html` serve real specs, zero cart furniture, zero empty
  shells.
- **`section_source_drift` check.** **Proven live 2026-07-16:** flags exactly
  one real drift (`contact`), zero false positives.
- **Backlog triaged.** 36 items → 6 genuine; ~30 verified-stale zombies closed.

## 4. What is BUILT but NOT working / NOT proven

| thing | state |
|---|---|
| `refresh_product_specs` | **runs, refreshes 0** — §1 above. Needs image build + diagnosis. |
| my new extraction logging | in working tree, **needs image build** |
| meta-commentary guard | deployed, never exercised by a real case |

## 5. Open decisions / known-unfixed (each already has an owner doc)

Errors live in `../aaa_fails_to_mend/` — **`002_HANDOFF_2026-07-15_errors_to_fix.md`
is mine**; read it for C/D/E:
- **A (spawn hang)** → **SUPERSEDED by `003_…spawn_lost_child_response.md`**
  (Kafka broker-2 node network path, fleet-wide). ⚠️ My original
  `action=orchestrate` diagnosis was **WRONG and is retracted** — do not chase
  it. Mitigation meanwhile: **retry**; dispatch is NOT a reliable workaround.
- **B** `TestParseLLMJSON_RepairsLiveEnvelopes` — 14 fixtures red, pre-existing,
  unrelated to this workstream. `go test ./platform/orchestration/actions/`.
- **C** `contact` page section drift — now auto-flagged by the drift check.
  Decide which component is intended, then align ALL THREE sources in one
  migration (template: `sql_for_agents/154`).
- **D** 6 genuine empty sections: news-listing ×3 pages (news-feed data gap —
  other subsystem) + `tool-guide-intro` (**hits the content-regression guard**:
  whole-page regen came out 6911 chars vs 31001 existing, guard blocks
  `< existing/4` → an empty section on a content-rich page can't be repaired by
  the page-scoped handler. Needs targeted single-section repair. Real gap.)
- **E** dartsonline product-grids will **vanish** on next rebuild (0 `products`
  rows behind `query.products`, `on_missing: skip_section`). Decide before it
  rebuilds.
- Sibling `001_…replan_clobbers_built_pages_FIX.md` is the same class as C —
  the drift check is a ready-made verifier for that fix.

## 6. Traps that cost real time (do not relearn)

1. **Verify deployed code against the POD, never git/tag.**
   `kubectl exec <pod> -- grep -ac "<distinctive string>" /app/agent-chassis`.
   Files saved *during* a docker build silently miss the `COPY . .` layer (cost
   a whole cycle on v1.0.1116).
2. **Section lists have THREE sources**, read in priority order and the winner
   is **synced DOWN over `pages.sections`**: `site_plan_sections` table
   (authoritative) → `site_specs.site_plan` aspect → `pages.sections`. Editing
   only the lower ones silently reverts on rebuild (cost the product-detail
   regression; SQL 154 fixed it). RUNBOOK §5c.
3. **Re-driving a work item:** ALWAYS `attempt_count=0` with
   `status='triaged'` + clear claim fields, or it sits undispatchable forever.
4. **Dispatch pickup lags ~5-7 min.** Don't call it a stall too early.
5. `products` needs `slug` NOT NULL.
6. **The reaper** fails stuck orchestrations at 30 min (dispatch loops) / 90
   min (everything else) — that's the backstop, not a bug in your workflow.

## 7. Housekeeping

- `README_where_we_are.md` in this dir is a **stale artifact** (a pasted chat
  reply from the first session). Ignore or delete it; this handoff supersedes.
- Migration numbering collision: my `151_gripper_spec_sheet_component.sql`
  shares its number with someone else's `151_footer4col_flexwrap.sql`. Both are
  applied; cosmetic only.
- Applied by this workstream: SQL **149, 150, 151(gripper), 152, 153, 154, 155,
  156**. Go: verifier registry + gate, `required_fields_missing`,
  `section_source_drift`, `resolveProducts`, meta-commentary guard,
  `refresh_product_specs`.
