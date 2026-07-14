# Handoff — idea.uk chassis, continue from 2026-07-06

You are continuing work on a multi-agent Go/Kafka/Postgres system that plans and builds
multipage websites from a domain name. This handoff carries the state, the working rules,
and the concrete next actions. Read the two long-lived docs alongside it:
`RUNBOOK_scheme_to_components.md` (the operational record; its LAST "▶ WHERE WE ARE"
section supersedes earlier ones) and `running_notes_scheme_to_components.md` (chronological
checkpoints Sa–Uo).

## System in one paragraph
Every agent is an orchestrator owning a workflow of one-or-more steps that call Go
**actions**; children reply to the parent's `responses_topic`, never their own. The
platform intelligently plans and builds targeted multipage sites from a domain. Sites
deploy git → GitHub Actions → Backblaze B2. Kubernetes: main namespace `ai-persona-system`
(e.g. `postgres-clients-0`), Kafka in namespace `kafka` (cluster
`personae-kafka-cluster-...`). The reference site is **idea.uk**, site_id
`1244516d-014d-421c-88c6-090bb1e9552a`.

## Working rules (hold these)
Go, not Python. British English. Keep **workflows simple**; put complexity in Go action
code. Do **not** build sub-workflows in SQL — spawn sub-agents with their own workflows
(keeps logs clean and responsibilities separate). Keep workflow variable names in sync
with what actions expect; do not rename identifiers in current code except deliberately
and noted. Never use `logger.Debug` (won't show); use `logger.Info`. **Reuse and adapt
existing functions/structs before writing new ones** — check first, every time. **Schema
first**: always read `\d <table>` before writing SQL. Prefer **structural fixes over quick
patches**. Do not treat a 0-row result as decisive until you've checked the query itself
isn't at fault. Work in reasonable step sizes. Run SQL as files:
`PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"` then `$PSQL < file.sql`.

## What is DONE and deployed (do not redo)
Three threads are closed and verified on the deployed site:
1. **Scheme → components (the original P0), closed 2026-07-03.** idea.uk resolved a LIGHT
   scheme but was rendering dark chrome/sections. Root cause was not a bug but an
   incomplete convention: components carried literal colours instead of consuming the
   library's paired "on-colour" variables. Fix completed the paired-variable standard —
   one layout patched (CTA pair + contrast fixes + an `updated_at` trigger reusing the
   shared `set_updated_at`), ten templates de-hardcoded to consume pairs / `--hero-ink` /
   ambient core vars, the chrome components repointed and force-re-rendered, then a full
   page-build. All nine deployed-grep checks passed (the decisive one: `var(--accent-color`
   count 0 — the stale-section fossil gone).
2. **Brief-explanation section, closed 2026-07-04.** The section was dropped from index
   and tools during rebuild because `plan_sections` couldn't resolve its `illustration_url`
   and escalated `needs_section_data`. The skip_field fix already existed on disk but had
   never shipped (deployed-binary drift). Deploying it plus a template `{{if
   .illustration_url}}` gate returned the section; two manual `site_plan_imagery` rows +
   `needs_imagery` items generated the two illustrations. A long "index miss" sub-thread was
   ultimately a BLIND PROBE (the check tested the asset-key string, but B2 objects are
   UUID-named) — the resolver worked throughout.
3. **Presigned-URL expiry, closed 2026-07-06.** `assets.url` stored presigned B2 URLs with
   `X-Amz-Expires=604800`; baked into pages they would die in 7 days. Heroes were safe
   (they render local `/assets/images/` paths via a legacy site-level `content_data.hero_url`
   that shadows the presigned `background_image`), but section illustrations rendered the
   presigned URL. A backfill (`w9_04`) flipped all 18 idea.uk asset URLs to repo-local
   paths, preserving the S3 object path into `storage_path`; renders verified local
   (`t/f/f`).

**Go batch — deployed:** slice 1 (`plan_sections_action.go` Edits A+B: required-branch
`skip_field` case; `ensureAssets` section-scope query mapping by key and kind-alias),
slice 2 (`component_library.go` Edits C/D: scheme-aware `RenderFallbackHeader/Footer`
consuming chrome pairs; Edit E: eight `logger.Debug`→`Info`), and Edit F
(`deploy_image_asset_action.go`: post-commit UPDATE recording the repo-local URL on the
asset row so recurrence is prevented for all kinds).

**Go batch — DELIVERED, apply/confirm landed:**
- Slice 3 — the re-aimed `fix_forced_text_colours_action.go` (in outputs, 739 lines,
  structurally verified). It replaces an `is_dark_section`-keyed decision core with a
  painting **classifier** (pair band / ink model / palette-or-hex band / ambient) + a
  `--section-*` **declaration rewriter** (references only; ambient sections declare
  nothing), keeping the function signature (the `isDarkSection` param is deliberately
  ignored), the literal-strip machinery, and the contrast gate. The injector trio was
  deleted (removing a stray `logger.Debug`). The three palette-band strings use
  `var(--color-primary-text, var(--color-background))` — confirmed against the deployed
  page (`--color-primary-text: #ffffff` is defined). **Filename caution:** the repo file
  may be spelled `...colors...` (US); ensure the patched content lands in the ONE real
  filename, no parallel file. Compile gate: `go build ./...`.
- Slice 4a — creator prompt: **LANDED in the DB** (`slice4a_creator_prompt.sql` applied;
  gate t/t/t/t/f → UPDATE 1 t/t/t/f). The prompt now teaches SECTION PAINTING, the
  image-fields rule (optional + `skip_field` + template gate), the extended token
  vocabulary, and demotes `is_dark_section` to catalogue metadata.
- Slice 4b — `003_contracts_and_standards_7_.md` patched (in outputs): item 6 →
  painting contract, new 6b image-fields rule, both literal dark examples → reference
  form, narrative de-dark-keyed. Copy over the repo doc.
- Slice 4c — `flag_page_image_rebuild_action.go` edit (`gobatch_05_flag_section_scope.md`):
  section scope_refs (`<page>:<ordinal>`) prefix-split to the page and fall through to the
  existing emit path (so future section-imagery landings trigger their own rebuilds) + the
  header comment update; add `strings` to imports if absent. The step-description SQL
  (`slice4c_step_description.sql`) is **already applied** (UPDATE 1).

## OPEN ITEM 1 — build the three catalogued-but-uncomposed pages (the real next task)
idea.uk's nav references `news-index`, `guides-index`, `tool-audience-check`; their links
404 because the pages were catalogued but never composed. Confirmed via reads: all three
have `pages` rows (`section-index`/`section-index`/`tool`, `build_status=planned`) but
`sections=[]` and **no `site_plan_sections` rows**. This is NOT a planner bug — composition
happens in `build-site-planner` → `write_site_plan_action.go`, which writes
`site_plan_sections` for the pages in a plan run; idea.uk's current plan simply didn't
include these three (forward-looking catalogue).

**The fix is the tested route, and it is safe to re-run** — do NOT hand-write
`site_plan_sections`. `v3_site_actions.go`'s `normaliseRealisedToPlanPage` (~:4383) exists
so a re-plan loads the already-realised pages via `load_existing_pages`, carries their
existing `sections` forward, and UNIONS them with the LLM proposal (its own comment warns
that without carrying sections the upsert would clobber built pages). So re-running
`build-site-planner` for idea.uk composes the three missing pages while preserving the six
built ones; the cascade (`needs_composition` → site-design-planner, `needs_design` →
webdesign-agent) then `needs_page` builds render and deploy them.

**Before emitting, run `stepF_replan_read.sql`** (in outputs) and read: F.1/F.2
build-site-planner's workflow steps + the `needs_site_plan` item it consumes + the
`load_existing_pages`/`write_site_plan` step configs and input_mapping; F.3 (schema-first)
the historical `needs_site_plan` item shape to copy; F.4 that the three pages carry nav
intent (`in_header`/`nav_order`) so `load_existing_pages` surfaces them. Then emit a
`needs_site_plan` work item for idea.uk matching that shape, watch the plan supersede and
`site_plan_sections` populate for all pages (including the three), let the cascade run, and
verify the three pages build and deploy (nav links resolve). If the LLM proposal does not
re-include the three catalogued pages, the narrower option is to drive the same
`write_site_plan` path with a page list that unions the realised catalogue — F.1/F.2 show
whether the workflow already does this.

## OPEN ITEM 2 — Step D: supervised first run of the re-aimed fixer
Once slice 3 is built and deployed, run the re-aimed `fix_forced_text_colours` ONCE against
a disposable site under supervision before it is ever pointed at the improvement loop or a
second site. Use **dartsonline.com** (site_id `5fe8785b-223d-41a3-88ee-c07187622381`),
freshly built and safe to wipe/resubmit. Note it is a THIN specimen: its 16 components
already avoid literal `--section-*` (built after this work began); only two `product-grid`
rules carry hardcoded text hex. So the correct outcome is "few rewrites, two strips, rest
untouched" — that proves the run is safe, not weak. Spawn via the harness in the debugging
guide (016b), read the returned `details` JSON (per-component: which `--section-*` rewritten
to which class, colours stripped, contrast-gate skips) rather than trusting the render, then
re-read the components. The decision that follows is the library tail: if `details` are
clean, the fixer is the vehicle for the ~10 remaining surface-painting declarers + ~17
band-class components elsewhere; if misclassified, tune the classifier regexes first.
Do not point the improvement loop at the fixer until this run is judged. If a run makes a
mess, dartsonline is disposable — write a guarded delete cascade across the seven tables
(sites, pages, page_components, site_specs, site_work_items, assets, site_plan_imagery)
against the schema, then resubmit via `082_submit_domain_unified.sh`.

## OPEN ITEM 3 — re-seed check for slice 4a (5 minutes)
component-creator is dynamically registered and today's deploy left `default_config`
intact, so the DB row is authoritative. Confirm nothing in code can revert 4a:
`grep -rn "DARK SECTIONS (if the section has a dark background)" --include='*.go' .` in the
chassis repo. A hit → an in-code prompt template exists; mirror 4a's four changes there. No
hit → done.

## OPEN ITEM 4 — contact-form spam (new, see spam_read.sql)
idea.uk's report-request form is receiving spam (fake orders, test data). The owner wants
existing spam removed and an IP block list. This is schema-first and separate from the
above — `spam_read.sql` (in outputs) finds the submissions table and whether IP is
captured before any delete or block design.

## Hygiene backlog (no deadlines)
`site_work_items.updated_at` frozen through transitions; legacy `active`/`system` site
statuses; "Rerender:" commit message shared by full builds; claim-timeout retries on long
builds; the legacy pinned head component (source of a bare `var(--color-cta-bg)`);
scope_ref-by-ordinal drift; the site-wide hero currently shows hero-about's image on every
page (last-write-wins `purpose+"_url"`; per-page heroes need `background_image`
un-shadowing); an orphan `illustration-specimen-report.jpg` in the repo; the
component-creator definition `updated_at` bumping on deploy (benign so far).

## Key files in outputs
Docs: `RUNBOOK_scheme_to_components.md`, `running_notes_scheme_to_components.md`,
`003_contracts_and_standards_7_.md` (patched), `gobatch_04_fixer_reaim.md`,
`gobatch_05_flag_section_scope.md`. Applied/staged SQL: `slice4a_creator_prompt.sql`
(+rollback, applied), `slice4c_step_description.sql` (applied), `stepF_replan_read.sql`
(next), `stepD_and_pages_reads.sql`, `spam_read.sql`. Patched Go: `fix_forced_text_colours_action.go`.
