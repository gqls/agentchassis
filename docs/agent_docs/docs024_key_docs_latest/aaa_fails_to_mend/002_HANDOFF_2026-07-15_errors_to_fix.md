# HANDOFF — errors surfaced but NOT fixed (route each to its own chat)

**Created 2026-07-15 from the `empty_sections_loop_integrity` workstream.**
Each section below is an INDEPENDENT error — self-contained, can be handed to a
separate chat in any order. They were found while doing other work and
deliberately left unfixed (blast radius, out of scope, or pre-existing). None
block that workstream, which is complete and proven live.

Context you may want first: the parent workstream docs are in
`docs024_key_docs_latest/empty_sections_loop_integrity/` (PLAN / RUNNING_NOTES
/ RUNBOOK). Standing mechanisms (DB access, re-drive, deploy) are in that
RUNBOOK §0-§5. Testbed site robot-hands.com =
`00ff3af5-dad8-4770-9f70-3edc267a3c92`; dartsonline.com =
`5fe8785b-223d-41a3-88ee-c07187622381`.

**NOT errors — do not "fix" these:** two things built this session register on
the NEXT chassis image and are waiting, not broken — the
`section_source_drift` discovery check and the `refresh_product_specs` action /
`product-spec-refresher` agent. SQL 155 + 156 are already applied. They are
inert until an image rebuild, by design.

---

## A. PLATFORM BUG — `action=orchestrate` of an orchestrator-mode agent hangs its child handshake  ★ highest value

**Severity: real bug, core messaging layer. Deliberately not fixed here
(blast radius). This is the one worth a focused effort.**

### Symptom
Invoking an orchestrator-mode agent (one whose workflow does
`spawn_agent` → `call_agent` internally) DIRECTLY via a kcat
`{action: orchestrate, config.agent_type: <agent>}` envelope hangs forever at
the spawn step. Concretely, firing `page-build-handler` this way:
- runs `plan_sections` fine,
- spawns the `page-content-writer` child (child logs init OK, sends its init
  response to `system.agent.generic.responses`, logs "starting agent's own
  workflow"),
- then the child idles 180s with `awaiting_count: 0` and shuts down — it NEVER
  receives the `call_content_writer` work request,
- the parent sits at `spawn_content_writer` / `AWAITING_RESPONSES` until the
  stale-orchestration reaper fails it at 90 min
  (`error: "reaper: stale AWAITING_RESPONSES for >90 min"`).

### Root cause (investigated, not guessed)
`action=orchestrate` wraps the target agent's workflow in a
**generic-orchestrate** context (see `platform/messaging/processor.go` ~L1355+,
`isOrchestrateAction`/ProcessMessage). The wrapper is a top-level orchestration
owned by `generic`. The handler's internal `spawn_agent`→`call_agent`
handshake registers its awaited-request correlation relative to that wrapper,
but the spawned child replies with `parent_orchestration_id` = the wrapper's
id, and the init response never matches the waiting spawn step's awaited
request. So the parent never advances from `spawn_content_writer` to
`call_content_writer`. Via the normal dispatch path
(`build-dispatch-loop` → `page-build-handler`), the handler is itself a
correctly-nested child and the correlation lines up — so it works there.

### Evidence it is invocation-shaped, not a product bug
In the same window, 121 orchestrations reached `COMPLETED` while only the 2
direct `action=orchestrate` invocations hung. Production never invokes
`page-build-handler` directly — it always dispatches. So this bites operators
doing manual rebuilds, not the fleet.

### Where to look
- `platform/orchestration/coordinator.go` — response correlation / awaited-request
  matching (search `awaitedReq`, `spawn_agent`, `TargetAgentType`, ~L2400-2500
  handles the spawn/call response wrapping).
- `platform/messaging/processor.go` ~L1355-1410 — the orchestrate wrapper and
  `__work_request__`/reply-to metadata handling.
- Reproduce: RUNBOOK §5b of the parent workstream has the exact kcat command
  and the working (dispatch) alternative.

### Suggested direction (for the fixer to confirm)
Make the generic-orchestrate wrapper propagate orchestration identity into
nested spawns so a child's init/response correlates to the step that is
actually awaiting it — OR detect orchestrator-mode agents at the
`action=orchestrate` entry and run them with the same nesting the dispatch
path gives them. This is core-coordinator work: it needs its own test surface
(a direct-invoke integration test of a spawn+call agent) and the platform
owner in the loop. Do NOT patch it as a side edit — every agent interaction
flows through this code.

### Workaround already in place
Rebuild pages via the dispatch path (re-drive an `empty_section` work item),
never direct kcat. Documented in the parent RUNBOOK §5b. So this is a
correctness/ergonomics fix, not an outage.

---

## B. FAILING TEST — `TestParseLLMJSON_RepairsLiveEnvelopes` (14 fixtures)

**Severity: pre-existing red test in `go test ./platform/orchestration/actions/`.
Not introduced by this workstream (touches none of its code).**

### Symptom
`go test ./platform/orchestration/actions/` FAILS. 14 subtests under
`TestParseLLMJSON_RepairsLiveEnvelopes/<uuid>.json` fail with e.g.:
```
json_envelope_test.go:41: repair failed: unrecoverable after control-char repair: unexpected end of JSON input
json_envelope_test.go:41: repair failed: unrecoverable after control-char repair: invalid character 'w' after object key:value pair
```

### Where
- Test: `platform/orchestration/actions/json_envelope_test.go`
- Under test: the `parseLLMJSON` / JSON-repair helper it calls (find with
  `grep -rn "func parseLLMJSON\|ParseLLMJSON" platform/orchestration/actions/`).
- Fixtures: saved LLM-envelope `.json` files the test loads (uuids in the
  failure output) — likely under a `testdata/` dir near the test.

### What to determine (I did not)
Whether these fixtures represent envelopes the repair logic SHOULD handle (→
improve the repair function) or genuinely-unrepairable truncated output (→ the
fixtures should be quarantined / marked expected-unrepairable, or the test
relaxed). The two error shapes ("unexpected end of JSON input" = truncated;
"invalid character X after object key:value pair" = malformed mid-object)
suggest a mix. Start by eyeballing 2-3 of the named fixtures.

### Why deferred
Orthogonal to the product-data/loop-integrity work; a fix risks changing LLM
JSON-repair behaviour used broadly, so it deserves its own focused change.

---

## C. DATA DRIFT — `contact` page section sources disagree (robot-hands)

**Severity: latent — a rebuild of the contact page will silently swap a
component. Caught by the new `section_source_drift` check.**

### The drift
For robot-hands `contact`, the three section stores disagree:
- `site_plan_sections` table (AUTHORITATIVE): `hero-contact, contact-form, contact-info`
- `pages.sections` (deployed cache):        `hero-contact, contact-form, contact-block`

On the next rebuild of `contact`, `load_page_sections_from_spec` serves the
table and syncs it down — so `contact-block` (currently deployed) would be
replaced by `contact-info`. Whether that's correct depends on which component
is the intended one (is the live page right, or the plan?).

### Fix pattern
Decide the intended layout, then align ALL sources + `page_components` in one
migration — exactly what `sql_for_agents/154_product_detail_plan_sections_fix.sql`
did for product-detail (it's a copy-paste template: fix `site_plan_sections`,
re-do the `page_components` swap, realign `pages.sections`). The parent
RUNBOOK §5c explains the three-source model and has the diagnostic queries.

### Note
This is the ONE page the drift check flags today; it found nothing false. Once
the next image ships and `section_source_drift` runs, it will emit a
`needs_human_review` work item for this page automatically.

---

## D. GENUINE empty sections still live (robot-hands) — 6 work items, 4 slots

**Severity: real current defects (verified still-empty 2026-07-15), correctly
flagged, left for their owning subsystems.** These are NOT zombies — the ~30
stale zombies were triaged and closed; these 6 are what remained.

| page | slot | items | nature |
|---|---|---|---|
| gripper-catalog-index | news-listing | 1 detected | news-feed data gap |
| news | news-listing | 1 detected | news-feed data gap |
| news-index | news-listing | 1 failed + 1 detected | news-feed data gap |
| gripper-cycle-time-estimator | tool-guide-intro | 1 failed + 1 detected | LLM-content gap |

### news-listing (3 pages)
Empty because the site's news feed has no populated source. Owned by the news
subsystem (`check_news_feed` / `check_empty_blog` / news feed population), NOT
page-build-handler. Fixing = give the site a news source, or remove the
news-listing sections. Out of the parent workstream's scope.

### tool-guide-intro on gripper-cycle-time-estimator
An LLM-content gap — `required_fields_missing` also flags it (12 missing
`source: llm` fields: headline, lead_paragraph, step_1..3_title/desc, etc.).
Because the fields are LLM-sourced, a normal rebuild SHOULD fill them: re-drive
via the parent RUNBOOK §3 (reset one of its `empty_section` items to
`triaged`, `attempt_count=0`, and let build-dispatch-loop run it). This is the
cleanest test of the now-honest loop genuinely FIXING (not just honestly
escalating). If it fills, close the duplicate; if it can't, it will honestly
land at `needs_human_review`.

### Dedup debt
`news-index` and `tool-guide-intro` each carry a `failed` + a `detected` item
for the same slot. Harmless but messy; a fixer may want to collapse each pair.

---

## E. LATENT — dartsonline product-grids will VANISH on next rebuild

**Severity: latent; surfaced while investigating the resolver. dartsonline's
own concern, flagged for awareness.**

### The situation
dartsonline has two `product-grid` sections (`index`, `new-arrivals`), each
~3KB of rendered product cards — the "14 real cards" the original
empty-product handoff cited as proof the pipeline was sound. But:
- `products` rows for dartsonline: **0**
- product-grid field source: `query.products`, `on_missing: skip_section`,
  `required: true`, `min_items: 1`

The `query.products` resolver did not exist in code until this workstream added
`resolveProducts` (queryresolve.go). Those cards are **frozen `rendered_html`**
from some earlier mechanism that no longer runs. On the next rebuild,
`query.products` returns `[]` → required+min_items unmet → `skip_section` fires
→ **both product grids disappear from the deployed pages.**

### Fix options (dartsonline owner's call)
- Populate `products` for dartsonline (real rows, same shape as robot-hands'
  gripper rows — see `sql_for_agents/152`), OR
- accept the grids are decorative/stale and remove them, OR
- point them at whatever mechanism originally filled them (if it's meant to
  exist).

### Why it matters now
This workstream's `resolveProducts` makes `query.products` LIVE for the first
time — so the next dartsonline rebuild behaves per the schema (skip) rather
than serving frozen HTML. That's correct behaviour, but it changes what
dartsonline ships. Worth deciding before a rebuild surprises someone.

---

## Priorities (suggested)
1. **A** — the real platform bug; highest leverage (fixes manual invocation of
   every spawn+call agent). Its own focused chat + platform owner.
2. **B** — red test; quick to triage, unblocks a clean `go test`.
3. **C** — one migration, low risk, template exists (154).
4. **D** — mostly other-subsystem; the tool-guide-intro re-drive is a 10-min
   win.
5. **E** — dartsonline decision; not urgent but do before its next rebuild.
