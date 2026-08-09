# HANDOFF — finetuning.uk service — continue here (2026-08-08)

Cold-start for a fresh thread. Two workstreams meet at this domain; this file
covers both and says which is which:

- **The SERVICE** (this directory, `finetuning/`) — turning finetuning.uk into a
  paying fine-tuning-pilot product. Early; mostly decisions and verified
  plumbing.
- **The SITE REPAIR** (`../finetuning_uk_repair/`) — the 08-03/04 lane that fixed
  the broken site and drove the improvement loop. Nearly done; two bounded tasks
  remain (below).

Read this file, then the two PLANs it points at. Everything dated was verified on
that date against the live system, not carried forward.

---

## 1. Rulings in force (owner, dated, recorded)

1. **2026-08-05 — fine-tuning FIRST, RAG second; both get built.** The offer is a
   **bounded diagnostic pilot**: fine-tune small subsets of the customer's data on
   a **choice of small models** for a modest outlay, producing evidence-based
   answers — *enough data? bigger model needed? what's missing? what should the
   full corporate build be?* "You actually need RAG" is a valid pilot outcome and
   is how the second product connects. Recorded at the top of
   `BUSINESS_PLAN_finetuning_uk.md` with the April RAG-first passages CORRECTED in
   place, not rewritten. **Pricing deliberately not decided** ("we'll have to talk
   about that but we can set up the service first"). §5/§6/§7 of the business plan
   are RAG-tier numbers — flagged do-not-quote for the pilot.
2. **2026-08-04 — per-site loop firing is the standing method.** Not re-enabling
   `improvement-sweep`, not bulk-promoting. Entry point:
   `../finetuning_uk_repair/294_TRIGGER_improvement_loop_v1.sh <site_id> [domain]`
   (registered IMP-050). Its two pre-flight refusals are fleet protection — do not
   weaken, do not runbook `FORCE=1` as normal.
3. **2026-08-04 (CLAUDE.md) — every site goes through the framework.** No
   hand-built pages, and for this lane specifically: the product front end is a
   framework build like any other site surface.

## 2. Verified facts a fresh thread can rely on (all 2026-08-08 unless dated)

**Thunder API — BOTH keys work.**
- The local token is at **`~/.config/thundercompute/token`** (NOT
  `thunderadapter`, the path in the original request — that does not exist).
  65 bytes, written 2026-08-03 09:52.
- Local token: `GET /v1/instances/list` → **HTTP 200** (`{}` — zero instances,
  consistent with the fleet below). **Negative control run**: invalid token →
  401, no token → 401. So the 200 is discriminating, not a permissive endpoint.
- Cluster key (`THUNDER_COMPUTE_API_KEY` via `personae-default-secrets`,
  `envFrom` on thunder-adapter): same call from **inside the adapter pod** using
  its own env — 200-class (wget, which hard-fails on 401). Whether local and
  cluster hold the *same* string is unknown (secret read is classifier-blocked)
  and **no longer matters: both authenticate independently.**
- Base URL `https://api.thundercompute.com:8443/v1`; auth `Authorization: Bearer`;
  `/instances/list` returns a JSON **object keyed by instance id**, not a list
  (`internal/adapters/thunder/api/client.go`, verified comment 2026-05-20).

**GPU / training lane — cold but healthy plumbing.**
- `thunder_instances`: 23 rows, **all decommissioned**, newest 2026-06-18
  (measured 08-04). Zero running (the `{}` above corroborates from the API side).
- `thunder_config`: $30/day cap, max 2 concurrent, not paused.
- `thunder-training-monitor` scheduled task: **disabled, never triggered**.
  `thunder-reaper`: enabled, 900s.
- Training bucket `personae-model-training` wired in the adapter (startup log).
- **[UNVERIFIED] whether a training run has EVER completed end to end and
  produced a usable adapter.** The phase5 docs stop on a checkpoint-upload race
  (`working/phase5/HANDOFF_2026-06-06_checkpoint_upload_loop_await_race(2).md`,
  `CONTEXT_PACK_thunder_checkpoint_race.md`). Establishing this is task #1 below.

**The site (repair lane) — done except two bounded tasks.**
- Both pages that had the 19 broken icon-images serve clean (verified 08-03 at
  the served HTML; fleet census of that defect went 31 → 0 with a
  relaxed-predicate control on the zero).
- The improvement-loop queue **fully drained**: 259 complete, 0 triaged/claimed
  left; 85 needs_human_review (positioning/social-proof — genuinely human), 13
  failed, 25 unresolved, 11 blocked.
- **The chassis roll HAPPENED** (between 08-05 and 08-08): the bare-token checker
  is **LIVE** — pod-grep on both replicas: `image_url_404:bare-token-src` = 1,
  control `empty-src` = 1. The framework can now SEE the icon-in-image-slot class
  on its own. Council trail `cfc94d91-3d17-4f29-a370-2b91d1a59a6f`: REVISE round 1
  (right — the overlap landmine), APPROVED round 2.
- **Five case-study images still 404** (`/assets/images/case-study-*.jpg`,
  re-checked 08-08). Their 11 `image_url_404` items sit `blocked` ("No
  handler_agent set" — flag-only by design). This is the repair lane's task A.

## 3. Next actions, in order — each cheap enough to stop after

**A. (site) Generate the five case-study images** — plan §Phase 1 in
`../finetuning_uk_repair/PLAN_2026-08-04_imagery_then_visual_designer.md`, where
the two open questions are already resolved: the five paths are template-hardcoded
on `/case-studies.html` AND content-driven on `/index.html`, same five filenames,
so **five assets keyed `case-study-<slug>` fix both pages**; `image-build-handler`
is proven on this exact site (10/10 completions 08-03). **The live trap is the
extension**: both surfaces reference `.jpg` — if the generator emits `.png`,
`DeployedWebPath` yields a path nothing references and the repair reports success
while fixing nothing. Copy a completed `needs_imagery` item's spec from this site;
verify at the served URL.

**B. (site) The visual-designer pass** — owner asked for it explicitly, AFTER the
images are real. It is **not new machinery**: `visual-design-auditor` is spawned by
`design-audit-agent` inside the improvement loop, so it is one more per-site
firing of `294_TRIGGER` once A is verified at the artefact. Firing early wastes
the LLM call on holes it already reported.

**C. (service) Establish whether a training run has ever finished** — read
`working/phase5/` NOTES tail + the checkpoint-race handoff; if the race is still
the blocker, that bug is the first engineering task of the product. The
run-to-adapter path IS the pilot service.

**D. (service) The model menu.** "A choice of small models" is a product
requirement phase5 never had (single-model runs). Even two entries (a 1–3B and a
7–8B class) shape the eval report, and **the eval report is the deliverable**.
Owner conversation, then config.

**E. (service) Hosting: follow the webdesign.uk lane, do not fork one.**
`../webdesign_uk_build_service/PLAN_2026-08-04_webdesign_uk_vm_hosting.md` +
`SUMMARY_2026-08-04b_dynamic_site_capability.md` are the authority: framework
builds/deploys/monitors VM sites in production (relojistas.com, 20 pages); it does
NOT generate backend code — `site-engine` is one hand-written Go binary, and the
pilot's backend (upload → provision Thunder box → track job → return adapter +
eval report → charge) is the same shape: **hand-written once, in this repo,
deployed by the machinery that lane is proving this week.** Check that lane's
HANDOFF before starting; it takes the first-time costs.

**F. (service) Front end LAST.** It is the piece the framework already does well,
and it needs a backend to talk to. Owner wants it fully framework-hosted
dynamically — which per E is exactly the relojistas pattern.

## 4. Traps for the fresh thread (each cost this lane real time)

- **Token path**: `~/.config/thundercompute/token`. The `thunderadapter` path in
  the original request does not exist.
- **thunder-adapter logs B2 credentials in PLAINTEXT at INFO on every startup**
  (`storage/s3.go:32`, `DEBUGaa:` lines). Found incidentally 08-04, still
  unfixed, deliberately not bundled into this lane — worth its own small fix.
  > **CORRECTED 2026-08-09: fixed in tree (`bugs_open/233`), inert until the
  > next fleet roll.** Two things the trap line under-stated, per that file:
  > the leaking constructor is shared by 8 call sites, so EVERY storage-touching
  > service leaked, not just thunder-adapter; and the fix's own grep found a
  > same-class second leak — `CLIENTS_DB_PASSWORD` at INFO on every dynamic-agent
  > spawn (`spawn_actions.go`), fixed in the same commit. Rotation of the exposed
  > credentials (live in logs since 2025-10-28) is an OWNER decision, not taken.
- **A `page_rerender` without `spec.reason` re-staples STORED html** — it will
  complete, report success, and preserve the defect. Reason
  `section_data_resolved` regenerates. (Landmine + worked examples
  `sql_for_agents/294/295/296`.)
- **Postgres regex is not Go regex**: `\b` is a BACKSPACE there (`\y` is the
  boundary); a mis-ported pattern returns 0 rows at exit 0, which reads as "the
  fleet is clean". Census rule: prove the query on a known-positive first; when
  the population is legitimately zero, relax the predicate as the control.
- **`/instances/list` returns `{}`** when empty — an object, not `[]`; the
  instance id is the map KEY.
- **Council trail for the checker**: `cfc94d91-…` is fully resolved (APPROVED,
  trailer written on `a1aaec7b9`). Commit `1985c0433` carries a literal
  `Council-Submitted: pending` — a recorded wrong call (WRONG_CALLS 08-03), do
  not "fix" it; forward-only.

## 5. The documents, by question

| question | doc |
|---|---|
| what is the offer? | `BUSINESS_PLAN_finetuning_uk.md` — 08-05 decision block at top |
| what exists to deliver it? | `SUMMARY_2026-08-04_where_we_are_on_offering_the_service.md` (with 08-05 update block) |
| flywheel vs product split | `working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md` (08-05 update at §1) |
| where training stopped | `working/phase5/HANDOFF_2026-06-06_checkpoint_upload_loop_await_race(2).md` |
| site repair, full story | `../finetuning_uk_repair/` — PLAN ×2, RUNBOOK, NOTES (missteps), README (owner log), SUMMARY ×2 |
| imagery + designer sequencing | `../finetuning_uk_repair/PLAN_2026-08-04_imagery_then_visual_designer.md` |
| VM hosting authority | `../webdesign_uk_build_service/` — their HANDOFF is the live state |
| the per-site trigger | `../finetuning_uk_repair/294_TRIGGER_improvement_loop_v1.sh` (IMP-050) |

---

## 6. FROM THE `bugfix_201` LANE, 2026-08-09 — two of your pages are serving empty sections, and the record says they were fixed

Left here rather than acted on: **it is your site, so nothing has been dispatched at it.**
Everything below is read-only measurement against the live DB.

**The pages.** `ai-guides` (`69a50d5d-3732-4efe-9a79-f887b072fa86`) and `insights`
(`8867b4d5-12d1-4ecc-8956-109a80395a18`), slot **`featured-content`**, component
`b3e0c2c0-a9c4-4ccb-ac04-fe79b92909a3`, **334 bytes each** — empty shells, live now.

**Why your queue looks clean.** The `empty_section` items you have for these pages name the
slot **`featured-article`**, not `featured-content`. A repair on 08-03 satisfied them by
**replacing** the component rather than filling it, and the replacement came back under a new
slot name. So:

- grepping `item_key` for the slot you can see on the page returns **0** — it reads as
  "never detected", and that is what my own handoff wrongly said before I checked;
- the completion verifier went looking for the component the item named
  (`a553f25f…` / `a390860e…`), found it deleted, and could not tell "fixed" from "removed".
  Pre-`RFC_017` that errored **fail-OPEN**, so both items were stamped `complete`:

```
result->'_verification' = {"status":"error","item_type":"empty_section",
  "error":"cannot verify: component … no longer exists
           (genuinely fixed or silently deleted — indistinguishable here)"}
```

Both of your checks — the queue and the item status — therefore lie in the **same direction**.
Your `NOTES` line 344 (service lane) traces these to the missing `article-grid`/`category-section`
components, which was right for *those* slots; `featured-content` is a separate, still-open one.

**Nothing has re-detected it because nothing has run.** Site discovery has no recurring
driver at all — all five `scheduled_tasks` rows targeting the discovery agents are disabled
one-offs, including `oneshot-completeness-discovery-fai-20260803`, your own 08-03 run, switched
off after it fired. Filed fleet-wide as **`bugs_open/230`**. `findEmptySections`' predicate
matches both pages *today*, so the detector is not blinded — it simply has not been asked.

**What I'd suggest, and it is entirely your call:**

- Fire detect-only discovery when it suits you — `bugfix_201…/TRIGGER_fire_quality_discovery.sh
  finetuning.uk`. It **cannot** change a work item's status (that tail was deliberately removed
  from the 075 original, which carried a hardcoded `UPDATE … domain='finetuning.uk'`), so it
  raises items and triggers no rebuilds.
- **Before you let a repair run at these two**, know the cost that is now armed: with
  `RFC_017` fail-closed live on `v1.0.1268`, an `empty_section` whose target vanishes again
  burns up to `max_attempts` (3) **full page rebuilds** before a human sees it. If the handler
  replaces-rather-than-fills a second time, that is exactly what happens.
- These two rows are, as far as the registry's history shows, the **only two occurrences** of
  the absent-target case in its entire life — so if you do re-run this, it is also the most
  likely route to the first live execution of the fail-closed branch, which has never fired.

Evidence, queries and the slot-rename trap: `bugs_open/230`, `LANDMINES.md` (2026-08-08,
"an `empty_section` item's `item_key` names the slot as it was WHEN FILED"), and
`bugfix_201_page_content_writer_dispatch/NOTES` (2026-08-08 late).

— `bugfix_201_page_content_writer_dispatch` lane

---

## 7. SESSION 2026-08-09 — credential fix shipped; imagery Phase 1 (task A) dispatched

Written mid-flight so a fresh thread can take over cleanly. Everything below was
measured today unless dated otherwise.

### 7.1 The B2 credential trap in §4 is FIXED and LIVE — with one straggler

`bugs_open/233` (council **APPROVED** round 1, corr
`7490388d-c945-42c0-b3c4-c452741a10cd`; commits `43c1801d6`, `08a94b474`,
`aeeffb800`, `3ece76804`).

- The §4 trap **under-stated the scope twice.** `NewS3Client` is shared by **8
  call sites**, so every storage-touching service leaked, not just
  thunder-adapter; and the same-class grep found a second leak —
  `CLIENTS_DB_PASSWORD` logged at INFO on **every dynamic-agent spawn**
  (`spawn_actions.go`). Both are now presence booleans.
- **Verified live on v1.0.1274** at the pod, added-string 1 / removed-string 0,
  both chassis replicas and thunder-adapter.
- **STILL LEAKING: `render-audit-adapter`** (runs `browser-runner-adapter:v1.0.1194`,
  80 tags behind). The credential is in its retained log buffer **now**. Cause is
  a release-coverage gap — the service appears **nowhere** in the makefile, so no
  release ever bumps its tag. Filed as **`bugs_open/237`**, landmine added.
- **⚠ THIS GATES THE OWNER'S KEY ROTATION.** That pod reads the credential from
  env at startup, so **rotating while it is still on v1.0.1194 means the next
  restart logs the NEW key in plaintext.** Roll it to ≥v1.0.1274 first. Do **not**
  `kubectl delete pod` it as a shortcut — its overlay pins the old tag, so it
  returns on the same leaking image and writes a fresh credential line.

### 7.2 Task A (five case-study images) — QUEUED AND DISPATCHED, verification pending

`ORCH_ID=18b299ff-59d0-4676-b969-650f38ded505`, corr
`80076b13-b8fe-4d15-9d2e-03f9b87e4b44`, fired 14:50Z via the standing 294
trigger (both pre-flights passed).

**The extension trap §3A warned about is RESOLVED — and the answer inverts the
worry.** The generator **does** emit `.png`; the deploy leg publishes `.jpg`.
Proof is a completed run's own record (08-03 `llm-cost-calculator`): S3 original
`…796d3589….png`, `deploy_result.data.file_path`
`/assets/images/content-hero-llm-cost-calculator.jpg`. Confirmed independently by
`ImagePurposes["content_hero"] = {1600,900,85,"jpg"}` and by three live assets on
this site (200), with `.png` and underscore variants as negative controls (404).
**Reading only the generator gives a confident wrong answer here.**

**What was queued:** five `needs_imagery` rows, `handler_agent =
image-build-handler`, `purpose = content_hero`, `asset_key = case-study-<slug>`,
`status = triaged`, prompts derived from the site's own card titles/excerpts and
existing `card*_image_alt` art direction. SQL kept at
`scratchpad/queue_case_study_imagery.sql` (reproduced in the lane NOTES).

**Three mechanism facts a successor needs:**
- `detected` is **not claimable** — `load_work_items` takes
  `status IN ('triaged','approved') AND attempt_count < max_attempts AND
  approval_mode='auto'`. Hand-raised items must be written `triaged`.
- `build-dispatch-loop.load_items` sets **`max_items: 5`**, no pipeline/handler
  filter, `ORDER BY priority ASC, created_at ASC`. Five items and no other
  `triaged` row = one clean batch; a sixth would have been silently left behind.
- `check_image_url_404` is flag-only by design, so **discovery will never raise
  these** — hand-queueing is the plan's Phase 1, not a bypass.

**VERIFY AT THE SERVED URL, not the item status:**
```
for s in facilities legal-rag private-ai financial-data logistics-strategy; do
  curl -s -o /dev/null -w "$s %{http_code}\n" \
    "https://finetuning.uk/assets/images/case-study-$s.jpg"; done
```
All five were **404** at 14:50Z. If they are 200, task A is done and **task B
(the visual-designer pass) is simply the 294 trigger fired again** — but only
once these are verified at the artefact, or the designer spends its LLM call
re-reporting the holes.

**If an item goes `failed`:** read `result->'_verification'` and `error` before
re-firing. `max_attempts` is 3 and each attempt is a real generation.

### 7.3 Other threads are active on this site — check before dispatching

Runs at 13:51–14:44Z today (discovery + a design review) came from another lane;
all COMPLETED, none in flight at 14:50Z. Two `needs_imagery` items appeared at
13:51 for `model-approach-selector` and `tool-automation-savings-estimator` —
**other pages' content heroes, unrelated to these five.** Also note five
`phantom_internal_link` items (`needs_human_review`): the case-study cards link to
`/case-studies/<slug>.html` pages that **do not exist**. Related to the cards but
a separate, human-decision task — the images will render on cards whose links
still 404.

### 7.4 CORRECTION to 7.2 — the first dispatch COMPLETED and served none of them

> **CORRECTED 2026-08-09 15:45Z.** §7.2 says task A is "QUEUED AND DISPATCHED,
> verification pending". The dispatch completed and **did not touch these five
> items.** Do not read §7.2 as "in progress, just wait".

`ORCH 18b299ff` ran 48 minutes → `complete / COMPLETED`, no error, and left all
five `triaged` / `attempt_count=0`. It dispatched a real batch of five — none of
them these.

**The mechanism, which is the transferable part:** `improvement-loop` runs
`triage_findings` **before** its dispatch step, and that step promotes
`detected → triaged` in bulk across the whole site. This run promoted ~95
findings (20 at priority 35, 46 at priority 80). `load_work_items` is
`ORDER BY priority ASC, created_at ASC LIMIT max_items(5)`. So five items that
were *the only claimable rows on the site* when I queued them became
*positions ~80–84 of a 95-item queue draining five per run* — **because of the
run I fired to serve them.** At five per firing that is ~16 more loop runs.

**So: firing the 294 trigger does not mean your item gets worked.** Before
relying on queue position, read the priority histogram as the dispatcher will
see it, *after* triage:

```sql
SELECT priority, item_type, status, count(*) FROM site_work_items
WHERE site_id='1368e337-dd1d-4799-bbb3-8221a1b79bcc'
  AND status IN ('triaged','approved') GROUP BY 1,2,3 ORDER BY priority ASC;
```

**Current state (15:45Z):** the five are re-prioritised to **`priority = 1`** —
exactly one `max_items` batch, with an in-transaction assertion that zero
claimable items sit ahead. `1` and not something softer because triage mints low
numbers itself (the batch that displaced them carried a priority-5 item). A
`build-dispatch-loop` is **live on the site now** (`0512f186`, claimed a
`phantom_internal_link` at 15:39Z), so the next `load_items` should take these
five. A watcher is running; authority is the served URL.

**Two traps confirmed while diagnosing this, both costly if hit:**
- **Do NOT hand-dispatch `build-dispatch-loop` with a bare `action=orchestrate`
  to skip the expensive loop.** It reports COMPLETED and processes nothing
  (`LANDMINES.md` 2026-08-08, webdesign_uk_build_service). The trigger's
  spawn+call is the only supported invocation.
- **Do NOT re-fire the 294 trigger while an item is `claimed`** — the pre-flight
  refuses, correctly, and `FORCE=1` here would race a live dispatch loop.

Full account and the cheap check that would have caught it:
`WRONG_CALLS.md` 2026-08-09, and the lane NOTES correction of the same date.

### 7.5 FINAL STATE 2026-08-09 ~16:10Z — images done, pages are the remaining half

**The five images exist and serve.** All five `/assets/images/case-study-*.jpg`
return **200 `image/jpeg`**, 52–94KB, distinct sizes; a made-up sixth filename
404s, so the check discriminates. Work items all `complete`. The
re-prioritisation to `priority = 1` was what unblocked it — claimed one at a
time from 15:48Z, all five serving by 16:01Z.

**But neither page showed them, and the reason differs per page.** Verifying at
the asset URL alone would have reported task A complete while both surfaces were
still wrong — and while the homepage was newly broken.

**`/index.html` — REGRESSED TODAY, filed as `bugs_open/238`.** A `tone_shift`
item (`page-build-handler`) regenerated the `case-studies-grid` section at
15:17–15:18Z. The new content keeps every `card*_image_alt` and **drops every
`card*_image_url`**, which the template requires — so the live homepage serves
five `<img class="csg-card-image" src="">`. It also **rewrote which case studies
the cards describe** (card 4 went from private-AI deployment to "coordinating
agent processes"), so the five generated images no longer map 1:1 and **this is
not repairable by pasting the old URLs back**. Read 238 before touching it.
That run was dispatched by the improvement-loop firing made in §7.2 — the defect
is the handler's, the timing is mine.

**`/case-studies.html` — fixable with no content decisions, IN FLIGHT.** Its
`case-studies-list` template **hardcodes** the five paths (unchanged since
2026-03-09); only its `rendered_html` was stale (2026-04-23, predating the
assets). Queued a `page_rerender` with **`spec.reason = "image_landed"`** at
priority 1 — claimed 16:09:28Z by the live `build-dispatch-loop`.

> **Why `image_landed` and not a bare rerender.** Three
> `page_rerender_case-studies_…_assemble` items (no `spec.reason`) have already
> **completed** on this page and left the images absent — that is the
> assemble-only path, which re-staples STORED html and cannot pick up a template
> change. `reason ∈ {image_landed, section_data_resolved}` routes through
> `rerender_page_sections`: re-render from template + stored content_data,
> **no LLM**. No-LLM is deliberate — a regenerating rerender is exactly what
> caused `bugs_open/238`, and this page's images live in the template, so
> nothing needs regenerating.

**Verify (authority is the served page, not the item):**
```
curl -s https://finetuning.uk/case-studies.html | grep -o "case-study-[a-z-]*\.jpg" | sort -u | wc -l   # want 5
curl -s https://finetuning.uk/index.html | grep -c 'csg-card-image" src=""'                              # want 0, is 5
```

**Next actions, revised:**
1. Finish `/case-studies.html` (in flight) — verify at the served page.
2. `bugs_open/238` — decide the homepage repair. It needs content carrying image
   URLs for the *rewritten* case studies; a fix at the generator (candidate 1 in
   238) is the durable answer.
3. **Task B (visual-designer pass) is NOT ready.** Its whole premise is "fire it
   once the images are real"; with five empty `src` on the homepage it would
   spend its LLM call reporting the holes. Do 2 first.
4. **⚠ Before firing the 294 trigger again, read the priority histogram** (§7.4)
   **and know that the loop dispatches content-REGENERATING items**
   (`tone_shift`, `content_rewrite`, `cta_improvement`) which, while 238 is open,
   can drop image URLs on whatever page they touch. That is a live risk, not a
   theoretical one — it already happened once today.

### 7.6 The `/case-studies.html` rerender ESCALATED — and the blocker is the contact-block, not the imagery

The `image_landed` rerender completed at 16:09:57Z **without rendering**:
`rerender_sections` returned `escalated: true, escalation: "raised"` and
`check_escalated` overrode the next step straight to `complete`, skipping
`save_sections` → `render_page` → `deploy`. It raised
**`needs_page:case-studies`** ("Full rebuild of case-studies — a section had no
stored content_data", `triaged`, priority 90, `page-build-handler`).

**This is the guard working, not a failure.** `rerender_page_sections` refuses to
re-render a section whose stored content is incomplete, because doing so would
render an empty section and **overwrite good HTML with a blank shell** — the
defect that once blanked live article bodies. It leaves the existing HTML intact
and escalates to the writer.

**The section responsible is `contact-block`, and it has nothing to do with the
case studies or the images.** It is missing **7 of its required `source:"llm"`
fields**:

```
form_heading · privacy_note · section_intro · form_subheading
section_heading · response_time_note · message_placeholder
```

```sql
SELECT pc.slot_name, f.key, (pc.content_data ? f.key) AS present
FROM pages p JOIN page_components pc ON pc.page_id=p.id
JOIN content_components cc ON cc.id=pc.component_id
CROSS JOIN LATERAL jsonb_each(cc.input_schema->'fields') AS f(key,val)
WHERE p.id='386683a5-eb6b-4256-bacb-420c44bf4c4a'
  AND f.val->>'source'='llm' AND COALESCE((f.val->>'required')::bool,false)
  AND NOT (pc.content_data ? f.key);
-- control (must be non-zero): drop the last line -> 9 required llm fields on the page
```

> **⚠ THE `->'fields'` IS LOAD-BEARING.** `input_schema` nests everything under a
> `fields` key. My first version iterated `jsonb_each(cc.input_schema)` and
> returned **0 rows** — it was iterating the single key `"fields"`, which has no
> `source`, so it would have printed "nothing missing" whichever way the truth
> lay. Always run the no-filter control alongside it.

**So the honest state of `/case-studies.html`:** the five images serve, its
template hardcodes their paths, and the only thing standing between them and the
page is a **full rebuild** — which is now queued as `needs_page:case-studies`.

**Decide before letting that run** (deliberately NOT dispatched by me):
- `needs_page` has a poor record on this site: **5 failed, 4 wont_fix, 2
  rejected** against 20 complete.
- A full rebuild regenerates content with an LLM, which is the exact operation
  that caused `bugs_open/238` on the homepage two hours earlier. This page's
  images are template-hardcoded so they should survive, but its *copy* would be
  rewritten — including the case-study text.
- The cheaper alternative is to fix `contact-block`'s 7 missing fields first,
  then re-run the `image_landed` rerender, which needs no LLM and cannot rewrite
  copy. **That is the recommended order** — it also serves the 25
  `required_fields_missing` items already sitting in `needs_human_review`, and
  `contact-block` is separately implicated in `bugs_open/228`.
