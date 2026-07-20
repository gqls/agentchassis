# HANDOFF — errors surfaced but NOT fixed (route each to its own chat)

> **STATUS 2026-07-19 (bugfix thread).**
> **B ✅ FIXED** (already resolved by the truncation work in `/bugs_closed/005`;
> verified green — the entry was stale, not the code).
> **C ✅ FIXED** — migrations **175** + **176**, both applied and
> ledger-recorded. It turned out to be **two sites, not one**; fleet-wide
> section drift is now **0 pages**.
> **A** superseded by `003`. **D**, **E**, **F** still open — D and E
> re-grounded against the live system 2026-07-19 (see Priorities).
> This file stays in `/bugs_open/` because D/E/F remain live.

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

## B. FAILING TEST — `TestParseLLMJSON_RepairsLiveEnvelopes` (14 fixtures) — ✅ FIXED

> **RESOLVED 2026-07-19 (bugfix thread).** The test no longer exists under that
> name and the package's envelope tests are green. It was rewritten as
> `TestParseLLMJSON_LiveEnvelopeDistribution`
> (`json_envelope_test.go:26`) when the truncation root cause was found and
> closed — see `/bugs_closed/005_…article_body_root_cause_is_truncation_FIXED.md`.
>
> The rewrite answers the exact question this entry said nobody had determined
> ("should the repair logic handle these, or are they genuinely unrepairable?").
> The answer was **both, and the test now asserts the split**: 12 of the 14
> fixtures are genuinely truncated by the old `max_tokens=2000` ceiling and must
> stay permanently unparseable ("no repair can complete a sentence the model
> never finished"); 2 are complete documents whose only fault was raw newlines /
> unescaped attribute quotes, and `repairJSONStringLiterals` now handles those.
> The test fails if that 2/12 distribution ever changes — so it also guards
> against the repair function silently accepting a partial article.
>
> **Verified 2026-07-19:**
> `go test ./platform/orchestration/actions/ -run 'TestParseLLMJSON|TestRepairJSON|TestMissingRequired'`
> → `ok` — all 9 envelope tests PASS, including all 14 fixtures.
>
> **Note for whoever runs the full package:** `go test ./platform/orchestration/actions/`
> currently still FAILS, but on `TestReconcile_BuiltPageCompositionSurvivesReplan`
> and `TestReconcile_BuiltPageOmittedByLLMIsUnioned` in
> `v3_site_reconcile_test.go` — an **untracked** file belonging to another
> session working error 001. That is not this entry, and not yours to fix.

**Original entry (retained for the record):**

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

## C. DATA DRIFT — `contact` page section sources disagree (robot-hands) — ✅ FIXED

> **RESOLVED 2026-07-19 (bugfix thread), and it was TWO sites, not one.**
> Migrations `175_robot_hands_contact_plan_sections_fix.sql` and
> `176_leopardess_aspect_generic_text_block_fix.sql`, both applied and
> ledger-recorded. **Fleet-wide drift is now 0 pages** (query below).
>
> ### C.1 robot-hands `contact` — intended component was `contact-block`
> The mechanism claim in this entry is **CONFIRMED**, not assumed: resolution
> Pass 1 is an *exact* match on `content_components.name`
> (`v3_site_actions.go:3383-3398`), so a plan naming `contact-info` binds the
> component named `contact-info` and `contact-block` is never a candidate. The
> swap was real.
>
> `contact-block` is intended. Sources 2 (aspect), 3 (`pages.sections`),
> `page_components` and the live page **all already said `contact-block`** —
> only the authoritative table, rewritten by the 2026-07-08 replan, disagreed.
> It is also bespoke to this one site, 28 schema fields against `contact-info`'s
> 6, template 12066 chars against 2573, and maintained 4 months more recently.
> Swapping would additionally have rendered an *incomplete* section:
> `contact-info` needs a business contact email this site does not supply.
> Migration 175 corrects source 1 only — unlike 154, the resurrection had not
> happened yet, so `page_components`/`pages.sections` were already right and
> were left alone.
>
> ### C.2 leopardess `index` + `case-studies` — same class, OTHER source
> Found while verifying C.1. leopardess has a current `site_plans` row with
> **zero** `site_plan_sections` rows, so source 1 misses and the **aspect** is
> authoritative for it. 16 leopardess pages carry a deployed
> `generic-text-block` (added 2026-07-18 at the `page_components` level, never
> written back up); on 14 it is harmless (page absent from the aspect, or its
> aspect entry has `"sections": null`, so source 2 misses). On **two** the
> aspect holds a real array that omits the block, so a rebuild would have
> **deleted live editorial copy** — verified live 2026-07-19: "The whole thing
> on one page / Every figure here comes from our own database…" (index) and
> "The three systems, as a route map…" (case-studies). Migration 176 aligns the
> aspect *up* to what is deployed — it makes no editorial decision. If the
> leopardess workstream wants those blocks gone, remove them from the aspect
> **and** `pages.sections` together.
>
> ### Work items closed
> `94de6b92` (robot-hands contact drift) — fixed by 175.
> `f50a8161` (leopardess `who-we-help` drift) — **stale**: that drift had
> self-resolved when the page rebuilt on 2026-07-18. The predicted revert *did*
> fire and restored `case-studies-grid` — but with the **audited honest
> framing** ("not client pitches dressed up as outcomes"), so no fabrication was
> resurrected. Checked because that site's audit had removed invented client
> case studies.
>
> ### Missteps worth keeping (they cost me time, or nearly cost a finding)
> - **I hypothesised the planner flip-flops between component names fleet-wide**
>   (the plan history alternates `contact-info`/`contact-block` and
>   `hero-contact`/`contact-hero` across five replans). **REFUTED** — a
>   fleet-wide comparison found **1 drifted page out of 91**. The alternation is
>   historical, not an active fleet defect. Confidence was not a signal.
> - **`hero-contact` is not a component at all.** Only `contact-hero`,
>   `contact-info`, `contact-block` exist. The slot binds `contact-hero` because
>   resolution falls back to `content_components.section_type`, which the
>   component-creator writes from the *requested* name while the LLM names the
>   row whatever it likes (`store_generated_component_action.go:636-645`). **Any
>   `section_type` value is a permanent, invisible alias**, and nothing enforces
>   `name == section_type`. Do not assume a section name is a component name.
> - **My first fleet drift query only compared the TABLE path** and so missed
>   every aspect-authoritative site — which is how leopardess nearly went
>   unnoticed. The effective authoritative source is *table if present, else
>   aspect*; a query that checks only one is not a fleet check.
> - **A scalar/`null` `sections` value silently drops pages from the
>   comparison** (`ERROR: cannot extract elements from a scalar`, or a NULL that
>   a `WHERE auth IS NOT NULL` then filters out). Guard with
>   `jsonb_typeof(...)='array'` on **both** sides or the query under-reports and
>   looks clean.
>
> ### Where the two migrations actually live (git provenance is misleading here)
> `176` is in the fix commit for this entry. **`175` is NOT** — between my
> `git add` and my `git commit`, another session's broad `add` swept my staged
> file into **`754577564`** ("v1.0.1138 — multiple docs, robot hands contact
> plan sections fix sql, reasoning debate", 42 files). Content verified
> byte-identical, nothing lost, and forward-only means it stays there. Recorded
> because `git log` on this bug will not show 175, and because it is a live
> instance of the hazard CLAUDE.md § Git already warns about: committing per
> task protects others from you, not you from others. The practical lesson is
> narrower than "add early" — it is **`add` and `commit` in the same breath for
> new files**, since the exposure window is exactly the gap between them.
>
> ### Verify (expect 0 rows)
> The corrected fleet-wide drift query, with its three gotchas, is in
> `docs024_key_docs_latest/empty_sections_loop_integrity/RUNBOOK_empty_sections_loop_integrity.md`
> **§5c-bis**. Ran clean (0 rows) after both migrations.

**Original entry (retained for the record):**

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

> **STATUS 2026-07-19 (bugfix thread): B and C are DONE. A, D, E, F remain.**

1. **A** — SUPERSEDED; all work is in `003_HANDOFF_spawn_lost_child_response.md`
   (Kafka broker-2 node network path + two platform gaps). Highest leverage —
   it hurts the whole fleet. Do NOT chase my retracted action=orchestrate
   theory.
2. ~~**B** — red test~~ — ✅ **FIXED** (was already fixed by the truncation work
   in `/bugs_closed/005`; verified green 2026-07-19). Note the package still
   fails on another session's untracked `v3_site_reconcile_test.go` (error 001).
3. ~~**C** — one migration~~ — ✅ **FIXED** 2026-07-19, migrations **175** and
   **176**. It was **two sites**, not one; fleet-wide drift now 0 pages.
4. **D** — news-listing is another subsystem (news feed). tool-guide-intro is
   NOT a quick fix: it hits the content-regression guard (see its entry) — a
   real guard-vs-repair gap needing a targeted-repair approach.
   *Re-grounded 2026-07-19:* those items are now `unresolved` (2 attempts
   spent), not `detected`/`failed`, and the news-listing half is additionally
   covered by `/bugs_open/026` (hardcoded English + dropped `h1`) and
   `/bugs_open/027` (news pages render no news without JavaScript). Read those
   before treating it as a data-supply problem.
5. **E** — dartsonline decision; not urgent but do before its next rebuild.
   *Re-grounded 2026-07-19: unchanged and still latent* — `products` for
   dartsonline is still **0**, and both `product-grid` sections are still
   frozen `rendered_html` last touched 2026-07-06 (index 3055 chars,
   new-arrivals 3048). Still the owner's call, so deliberately not actioned.
6. **F** — on-demand discovery dispatch. Untouched; needs a live trigger run.

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

> **⚠️ 2026-07-20 — suspect (a) is REFUTED, and the mechanism is now filed as
> `/bugs_open/034`.**
>
> **`action=orchestrate` is NOT dead.** It is the *first* clause of
> `isOrchestrationAction` (`platform/messaging/processor.go:983-988`); the path
> reads `config.agent_type` (`extractGroupInfo`, `processor.go:991-1057`) and
> looks the workflow up from `agent_definitions` (`FindByType`,
> `agent_discovery.go:109-125`) — no inline workflow required. Git history shows
> no removal or rename. All three discovery agents verified live 2026-07-20:
> `is_active=true`, not deleted, `default_config->'workflow'` present and an
> object. **So the closing worry — "other hand-rolled triggers across the repo
> are silently no-ops too" — does not follow from `action=orchestrate`.** It may
> still be true for the *other* reason below.
>
> **Two corrections to the premise while I was there.** The canonical trigger's
> `action=process` works **not** because `process` is a valid action — it is not
> in the set — but because an inline `config.workflow` is checked at Priority 1,
> *before* any action test (`processor.go:893-899`). And the action is read from
> **Kafka headers only** (`types/context.go:545`): `action` in the JSON body but
> not passed as `-H` yields `action == ""`, which falls through to the
> *consuming agent's own default workflow* with no log saying so. There is no
> `default:` branch and no dead-letter on that path.
>
> **The likely real cause of F.2**, and why it left no trace:
> `coordinator.go:142-144` returns `client_id is required to execute a workflow`
> **before** `getOrCreateState`, so no `orchestration_states` row is ever
> created; and `agent.go:828-845` swallows any error whose text contains
> `"is required"` / `"validation"` / `"invalid"` — skipping the error response
> to the parent, the retry, and any durable record. The only trace is one
> stdout line in a pod that rotates ~3.6k lines/10min. **That is exactly the
> "accepted, never executed, no error anywhere" symptom.**
>
> **Not proven for these two correlations** — the trigger was deleted
> (`15f612346`) and the 07-18 pod logs and metrics are gone. The mechanism is
> proven from code; its application to these runs is a strong hypothesis. The
> fact that it *cannot* be checked is itself the finding, which is why 034 is
> filed as a diagnostic-surface bug rather than closed here.
>
> Suspect **(b)** (`kcat -P` without `-c 1`) is untouched and still open.

**Why it matters beyond one run:** discovery is how the fleet finds dead
controls, phantom links and misdirected CTAs. If the only reliable way to run it
is the full improvement-loop, that should be *documented as the way* — and if
`action=orchestrate` is genuinely dead for these agents, other hand-rolled
triggers across the repo are silently no-ops too.

**Do first:** re-run via `076_improvement_loop_trigger.sh <site_id> <domain>`
(note it has hardcoded SITE_ID/DOMAIN overrides near the top that shadow its own
arguments — read before running). Reuse it; do not write another trigger.
