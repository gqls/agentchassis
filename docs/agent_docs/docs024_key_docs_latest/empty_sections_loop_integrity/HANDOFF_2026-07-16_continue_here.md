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

### STATUS 2026-07-17: `refresh_product_specs` WORKS. Two follow-ups remain.

The zero-refresh bug is **fixed and proven end-to-end** (v1.0.1128). First
working run: **4/5 refreshed, 0 LLM timeouts** (was 0/5, 5/5 timeouts). Details
in RUNNING_NOTES "Session 10 (cont.)". What's left for the next chat:

1. **Deploy the degradation guard.** The working run refreshed correctly but
   *degraded* 5 values by dropping hand-verified qualifiers ("6 mm per jaw" →
   "6 mm" — halves a parallel gripper's stated stroke). No fabrication; the
   model copied the page's value cell and lost the label's meaning. Fixed in the
   working tree (`specValueIsRestatement`, 12 green tests) but **NOT yet in an
   image**. The refresher is manual, so nothing re-degrades until it's re-run —
   but deploy this before the next run. **DB already repaired (SQL 157).**
2. **Festo returns {} — acceptable, not a bug.** Its source_url is an RS-Online
   distributor listing, not a manufacturer spec page; the model correctly
   refuses it. Give it a real spec URL if you want it refreshed (human judgement
   — that's the by-design discovery boundary).

Everything below this line is the diagnosis journey; keep for the reasoning.

---

**`refresh_product_specs` — root cause FOUND and MEASURED; fix written and
proven against the live model; DEPLOYED v1.0.1128 and proven live.**

### The cause was NOT truncation. That hypothesis is RETRACTED.

v1.0.1125 shipped the self-diagnosing logging; the re-run answered it flatly:

- **Zero** `markdown_chars_sent` warnings — the model never returned `{}`. It
  never returned anything.
- **5/5** products died identically on `Post ".../api/chat": context deadline
  exceeded (Client.Timeout exceeded while awaiting headers)`, spaced exactly
  **92s** apart = the action's 90s HTTP timeout + 1.5s pacing delay.

**Root cause: the 90s HTTP timeout was unreachable on this hardware.** Measured
live: **no GPU on any node**; mistral-small3.1 is 24B/Q4 on 8 CPUs; prompt eval
runs at **~3 tok/s**; Mistral-Small's chat template costs **~360 tokens (~120s)
before our text**; markdown tables tokenize at ~2.6 chars/token. The shipped
6000-char slice ≈ 3400 tokens ≈ **~19 min of inference behind a 90s timeout —
off by ~12×**. Content never mattered; every product was doomed identically.

The earlier "no LLM call failed was logged, so the HTTP call worked" conclusion
came from **already-rotated logs** (~3.6k lines/10min on this pod). Absence of
evidence read as evidence of absence. Capture with `kubectl logs -f > file`
*before* triggering.

### The fix (in the working tree; needs v1.0.1126 to be live)
1. **Timeout 90s → 600s** — matches the same-shape sibling
   `vet_med_price_scrape_action.go`, which has used 600s against this same model
   all along. The original build copied that sibling's shape but neither its
   timeout nor its text budget.
2. **`selectSpecRegion()` replaces `md[:6000]`** — 1500-char budget, picking the
   *densest spec-signal region* rather than the head of the page. At 1500 chars,
   WHICH 1500 decides success. Also stays under Ollama's default **2048 num_ctx**
   (above it, Ollama silently drops the START of the prompt — a second silent
   -truncation trap).
3. **num_predict 400 → 200** — output decodes at ~3 tok/s; 8 fields need ~70.

**Proven before deploy** (port-forward to the live model, exact action prompt):
**8/8 fields extracted correctly** — Schunk / 3 mm / 140 N / 0.7 kg / 0.19 kg /
IP40 / Digital I/O / 24 V DC. The model and prompt were never the problem.
5 unit tests green on `selectSpecRegion`.

### So do this
1. Confirm v1.0.1126 is on the pod **and contains the fix** (trap 1):
   `kubectl exec -n ai-persona-system <pod> -- grep -ac "selected page region" /app/agent-chassis`
2. Re-run the refresher (RUNBOOK §5d).
3. Expect **~5-6 min per product, ~30 min for 5** — that is normal on CPU, not a
   stall. Workflow `timeout_seconds: 900` → the reaper kills at **3× = 45 min**
   (`coordinator.go:780-788`), so 5 products fit with ~13 min headroom. **Beyond
   ~7 products this needs `limit` batching or a raised `timeout_seconds`** — the
   constraint is CPU inference and it scales linearly.
4. Verify: `SELECT name, content_data->>'verified_date', specifications FROM
   products WHERE site_id='00ff3af5-…' AND category='gripper';` — verified_date
   should advance to today.

**Verify by evidence, not by plausibility** — see the MISSTEP entry in
RUNNING_NOTES. The truncation hypothesis was plausible, had a sibling
precedent, and was wrong; one captured log line settled it. That is now twice
this workstream has paid for the same lesson.

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
| `refresh_product_specs` | **diagnosed + fixed, not yet proven live.** Root cause was the 90s timeout vs ~3 tok/s CPU inference (§1), NOT truncation. Fix proven against the live model offline (8/8 fields) + 5 unit tests; needs the v1.0.1126 run to be called done. |
| extraction logging | **LIVE in v1.0.1125** — it is what produced the diagnosis. |
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
- **RESOLVED 2026-07-16 (by the travelling-docs chat):** 151–156 were applied
  but never recorded in `schema_migrations`, so the runner kept replaying 151,
  hitting its duplicate-component error, and blocking there ("not cosmetic"
  after all — it gated every later workstream's migrations). Each file's
  artifacts were verified live in the DB (component, 5 products, page slots,
  plan section, drift check in completeness-discovery config, refresher agent)
  and ledger rows backfilled (`applied_by='ledger-backfill'`). Runner now
  reports "Up to date". Convention going forward: whoever applies a migration
  records it (the runner does this automatically; out-of-band applies must
  insert the ledger row themselves).
