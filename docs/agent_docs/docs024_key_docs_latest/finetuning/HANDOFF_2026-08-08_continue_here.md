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
