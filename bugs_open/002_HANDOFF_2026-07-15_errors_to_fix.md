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

**Sibling handoffs in this dir:** `001_…replan_clobbers_built_pages_FIX.md`
(the planner discarding built pages' composition) and error **C** below are the
SAME class — the `site_plan_sections` authoritative store diverging from what's
deployed. The `section_source_drift` check built this session (activates next
image) is a **detector** for that class: whoever fixes 001 can use it to verify
sources stay aligned after a re-plan. `003_…spawn_lost_child_response.md`
supersedes error **A** below.

---

## A. Spawned child response lost, parent hangs — SUPERSEDED BY `003_HANDOFF_spawn_lost_child_response.md`

**⚠️ 2026-07-15 CORRECTION — my root cause here was WRONG. Read 003 instead;
it is the authoritative diagnosis with far stronger evidence.** I am leaving
this entry as a record of a retracted diagnosis, not a live task.

### What I got right
The symptom: firing an orchestrator-mode agent (one that internally does
`spawn_agent`→`call_agent`, e.g. `page-build-handler`) hangs. It runs
`plan_sections`, spawns `page-content-writer` (child logs init OK and sends its
init response), then the child idles 180s with `awaiting_count: 0` and shuts
down without receiving `call_content_writer`; the parent sits at
`spawn_content_writer`/`AWAITING_RESPONSES` until the reaper fails it at 90 min.

### What I got WRONG (retracted)
I diagnosed this as an `action=orchestrate` generic-orchestrate wrapper
**correlation** mismatch, and claimed it "manifests only on manual invocation,
never production." **003 disproves both.** The real cause (003, with node-level
evidence) is a **Kafka broker-2 network path failure from certain worker
nodes** — child pods that land on a bad node hit
`dial tcp 10.20.99.93:9092: i/o timeout`, so they process their `initialize`
message (hence the init response I saw) but their request/response consumers
can never dial the broker to pick up the work or publish back. It is
**FLEET-WIDE** across production paths (`spawn_dispatch`, `call_content_writer`,
`process_item_iter_*_spawn_handler`, image gen, …), not manual-only.

Where my reasoning failed: my "121 completed vs 2 hung" was survivorship —
those children happened to land on good nodes. My two direct invocations landed
children on bad nodes, and I over-fit a correlation theory onto a 2-sample
observation instead of reading pod logs (which would have shown the dial
timeout, as 003 did). The `action=orchestrate` wrapper is almost certainly a
red herring.

### Consequence for the workaround (also corrected — see §RUNBOOK note below)
"Use the dispatch path" is NOT a reliable fix — 003 shows dispatch paths hang
too (`spawn_dispatch` had 38, `process_item_iter_*_spawn_handler` ~19). My
dispatch rebuilds succeeded by node-landing luck. The practical mitigation is
**retry** (a re-driven item eventually lands its child on a healthy node) until
the infra/platform fix in 003 lands. → **All fix work for this belongs in 003.**

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

### tool-guide-intro on gripper-cycle-time-estimator — attempted, hit a guard
Live section renders `<h1 class="tgi-headline"></h1>` (empty heading);
`required_fields_missing` flags 12 missing `source: llm` fields. **I re-drove
it 2026-07-15 via dispatch and it FAILED — the failure is the finding.**
`save_page_sections` has a **content-regression guard**
(`save_page_sections_action.go:335-371`): it blocks the save when the newly
generated page text is `< existingTextLen/4`. This rebuild produced 6911 chars
vs 31001 existing (threshold 7750) → blocked → handler `complete_error` → item
`failed` (attempt 1/3). The stale duplicate item was retired (`wont_fix`).

**So this is NOT a quick re-drive — it's a genuine tension.** A page-scoped
rebuild regenerates the WHOLE page; on a content-rich page (this one has a
calculator tool + a large FAQ) the regenerated text comes out thinner and
trips the regression guard. The guard is correct (it stops LLM failures wiping
good content), but it also means **an empty section on an otherwise-rich page
can't be repaired by the page-scoped handler.** Fix needs one of: a TARGETED
single-section repair (fill only the empty fields, no whole-page re-save);
understanding why whole-page regen is thinner (is the tool/FAQ content not
preserved on rebuild?); or richer content-writer output. A real architectural
gap (guard vs. repair), its own small investigation — not the 10-minute job an
earlier draft implied.

### Dedup debt
`news-index` still carries a `failed` + a `detected` item for the same slot
(tool-guide-intro's duplicate was already retired). Harmless but messy.

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
1. **A** — SUPERSEDED; all work is in `003_HANDOFF_spawn_lost_child_response.md`
   (Kafka broker-2 node network path + two platform gaps). Highest leverage —
   it hurts the whole fleet. Do NOT chase my retracted action=orchestrate
   theory.
2. **B** — red test; quick to triage, unblocks a clean `go test`.
3. **C** — one migration, low risk, template exists (154).
4. **D** — news-listing is another subsystem (news feed). tool-guide-intro is
   NOT a quick fix: it hits the content-regression guard (see its entry) — a
   real guard-vs-repair gap needing a targeted-repair approach.
5. **E** — dartsonline decision; not urgent but do before its next rebuild.

---

## F — On-demand discovery dispatch: envelope accepted, nothing runs (2026-07-18)

**Found** trying to point the auditors at idea.uk (which had never had a
discovery run — see `017`). Two distinct problems, both in *how you ask for a
discovery run*, not in the checks themselves.

**F.1 — `ensure_site_record` resolves BY DOMAIN.** An envelope carrying only
`site_id` dies at the first step:
`step ensure_site_record failed: … domain not found in input_data`
(orchestration `974a56f9-a109-414b-85ef-7b83aa9a4642`). The canonical trigger
already knows this — `scripts/initial_messages/060improvement_loop/076_improvement_loop_trigger.sh`
passes **both** `site_id` and `domain`. Anything new must too.

**F.2 — a well-formed envelope produced NO orchestration at all.** After adding
`domain`, correlations `cd2459ce-06f4-4f64-aa09-66ebd9ccdd3f` and
`199ba851-f0fd-4fa7-8909-de1a9cc790a5` created **zero** `orchestration_states`
rows — accepted by Kafka, never executed, no error anywhere. Not the documented
300s-post-restart drop: the chassis pod was ~6h old (v1.0.1135). Unexplained.

*Difference worth testing first:* the working trigger uses `action=process` with
an inline `spawn_agent`/`call_agent` workflow and `kcat -P **-c 1**`; the failing
one used `action=orchestrate` with `config.agent_type` and plain `kcat -P`. So the
suspects are (a) `action=orchestrate` + bare `agent_type` no longer being a
supported entry for these agent types, or (b) the missing `-c 1` letting kcat
publish a malformed/extra message (the known kcat line-splitting trap).

**Why it matters beyond one run:** discovery is how the fleet finds dead
controls, phantom links and misdirected CTAs. If the only reliable way to run it
is the full improvement-loop, that should be *documented as the way* — and if
`action=orchestrate` is genuinely dead for these agents, other hand-rolled
triggers across the repo are silently no-ops too.

**Do first:** re-run via `076_improvement_loop_trigger.sh <site_id> <domain>`
(note it has hardcoded SITE_ID/DOMAIN overrides near the top that shadow its own
arguments — read before running). Reuse it; do not write another trigger.
