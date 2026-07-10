# RUNBOOK — Diagnosis→Fix Loop (v2 of the diagnosis loop)

Rev 10, 2026-07-09. Supersedes RUNBOOK_diagnosis_fix_loop(9).md. The change in
this revision is confined to the ★ F0 PILOT section, which now records a
**closed, evidence-backed diagnosis** rather than a standing hypothesis, and to
the pre-check SQL, which had a wrong column name. Everything else is carried
forward. Detailed evidence in NOTES_running_fixloop(10).md; forward plan in
PLAN_fixloop_pilot.md.

## THE TASK (read this first if you are new)

The platform already has a working, read-only **diagnosis loop**: given a bug
symptom, an agent forms a hypothesis, gathers scoped evidence (real code bodies
from an indexed corpus, read-only database rows, runtime records), issues a
verdict that must CITE evidence or ABSTAIN, and re-scopes by FOLLOWING what the
evidence names — until it confirms a cause with citations across all three tiers
(static code / live data reads / runtime records). It is deliberately
human-gated: it emits a diagnosis and changes nothing.

**This workstream develops it into a diagnosis→fix system** with, in order:
1. **An easy, documented route in and out**: one clear way to input a task or
   bug; live monitoring of what the loop is doing and why (per-iteration — and
   per-step — reasoning written to a task-specific running-notes file); and a
   usable result out, including the ability to **fetch the bundles** the loop
   builds each iteration (today they are ephemeral, in-memory).
2. **Fixes on a branch**: the confirmed diagnosis drives a proposed fix
   committed to a separate git branch, so the human can amend, ditch, or apply
   it. The loop's core stays read-only; the write surface is isolated.
3. **A council of reviewers** before any fix is finalised: independent
   specialist agents each judging the proposal from their own perspective and
   sending opinions to a **decision-maker** that weighs them all. Initial
   roster: a **guidelines agent** (does the fix adhere to guidelines 000-0xx —
   or did the guidelines fall short?); a **reuse agent** (are we building a new
   route where a tried-and-tested solution exists — checking BOTH code and
   docs); a **bug-historian** (catch early; record bug categories so the same
   class never repeats); a **compliance/legal eye**; **pipeline guardians** —
   one per master workflow — checking the fix doesn't infringe on another
   workflow; and **specialist knowledge agents** (e.g. a trigger expert, a
   site-work-items triage expert) that answer "we already have one of these".
4. **Architecture-change visibility**: make it loud when a proposed change is
   accidentally fundamental — touching platform contracts, message shapes, many
   packages, exported signatures — before it ships.
5. **Learning**: recorded bugs, proposed guideline amendments, corpus and doc
   enrichment feeding back in.

**Mission for the tool**: use everything available to reach the right result —
the code corpus, schemas, runtime records, the guidelines themselves — with
checks, balances and second opinions built in.

## What already exists (do not rebuild)

- The **live loop** (chassis `pkg/diagnose` + `diagnose_*` actions): three-tier
  CONFIRMED diagnosis achieved; engine guards (named-scope narrowing, capped
  call-graph expansion, cite-or-abstain, SQL guard read-only allowlist); the
  §7D resolver. `RUNBOOK_code_retrieval_route.md` is the closed record.
- **contextkit CLI** (`cmd/bundle` + `cmd/analyse`, RUNBOOK_31_.md).
  `example_bundle.txt` is a real invocation.
- The **code_symbols corpus** + vector/trigram lookup.
- The **work-item relay + immune system** (builder thread §B2/§B3).
- The **tools chat's travelling-docs infrastructure** (rev-22): `doc_plans` /
  `doc_notes` live; the diagnose-agent workflow is rewired by them to
  `emit → persist_note (config.error_step="complete") → complete`; the subject
  gate is the action's first check; their 3b (threading
  `subject_type`/`subject_key`) is in flight. `load_runtime` error-routing is
  applied — **anchorless runs survive**. Canonical trigger:
  `drafts/084_TRIGGER_diagnose_v1.sh`. Their Stages 5–6 give a static Tier-2
  contract check and a Playwright **browser-runner adapter** (035-conformant) —
  acceptance/verification, not a rival loop.
- The **builder thread's pipeline map** (RUNBOOK_builder_route.md §B0–§B3) —
  guardian seed material.

## Phased plan (thin slices; pre-registered criteria per slice)

**F0 — Intake, observability, egress.** F0.1 bundle egress (persist each
iteration's bundle + one documented fetch route); F0.2 task input (one
documented way in); F0.3 per-task running notes (reasoning per iteration AND
per step). *Concrete slices, criteria and order: PLAN_fixloop_pilot.md §1.*

**F1 — Fix on a branch.** F1.1 fix-proposer turns a CONFIRMED diagnosis into a
patch on a new branch via the git adapter; PR opened; human amends/ditches/
applies. Write-token isolated to the proposer (spawn token-gate pattern).
F1.2 the per-task notes gain proposal rationale + diff summary.

**F2 — The council.** Independent reviewers, each with curated context (Q-G),
producing a structured opinion (verdict + citations + objections + suggested
alternative); a decision-maker aggregates; the human sees diagnosis + proposal
+ council report (Q-H). Architecture-change detector runs as one reviewer (Q-E).

**F3 — Learning.** `bug_records` (category taxonomy, recurrence checks feeding
the historian); guideline-amendment proposals routed to the human; corpus/doc
enrichment.

## Boundaries
- **Tools chat**: owns `doc_plans`/`doc_notes` + tool docs + their
  `diagnose_load_runtime` draft. F0.3 reuses rather than reinvents. Their
  diagnose-agent workflow is an ACTIVE surface — any change is fetch-first and
  coordinated; our egress lands Go-side in assemble precisely to stay off it.
- **Builder thread**: owns the relay/spine. The pipeline map is INPUT here.
  Guardian findings implying relay changes route back through it. **Causes A
  and the two-intake-path disagreement below belong to them.**
- **Quality thread**: a future consumer of fixes.
- **Imagery**: another chat.

## QUESTIONS — decided vs open

**DECIDED 2026-07-07 (owner).** **Q-A** `diagnosis_artifacts` table, written
through inside assemble (`kind ∈ {bundle, iteration_note}`). **Q-B** intake =
`needs_diagnosis` item in a new `pipeline='diagnose'` namespace (~~null-site
allowed~~ — **CORRECTED 2026-07-09, see below**; envelope extends 084; manual
trigger retained). **Q-C** separate fixer
agent (isolated write token; constrained edit plan; gofmt+build in a spawned
job pre-PR). **Q-D** flag-based `hard_veto`; guideline-gap = SIDE-TASK
(amendment PR against the guideline docs; human terminal; fix unblocked; F3
recurrence record). **Q-F** shape (c): working notes in our own storage; only
the TERMINAL note lands in `doc_notes` via the tools chat's `persist_note`.

**Q-B CORRECTION (2026-07-09, established from schema + code, not assumed).**
"Null-site allowed" is **impossible**, twice over: `site_work_items.site_id` is
`NOT NULL`, and `LoadWorkItemsAction` parses `site_id` as a required uuid and
filters `WHERE wi.site_id = $1` — the relay's loader is site-anchored by
construction, so a NULL-site item could never be loaded even with the constraint
dropped. **Instead, reuse the existing `system.internal` pseudo-site**
(`eac60db8-b032-432b-b36d-76f37632045d`, `sites.status='system'`), which already
carries platform-wide `maintenance` work. Every `needs_diagnosis` item anchors
there — *including* site-specific bugs — because `build-dispatch-loop`'s
`load_items` step is configured with only `{site_id, max_items}` and has **no
`item_pipeline` filter**, so any item parked on a real site is claimed by that
site's next build dispatch. The site under diagnosis travels in
`spec.site_id`/`spec.runtime_site`. Items are written at `status='detected'`,
outside the loader's `('triaged','approved')` filter, as a second guard.
Intake route: `090_TRIGGER_needs_diagnosis_v1.sh`. Automatic dispatch of the
`diagnose` pipeline remains unbuilt — it needs a pipeline-filtered loop, or an
`item_pipeline` filter on `build-dispatch-loop`, which is the builder thread's
surface and their call.

**STILL OPEN (F2-phase).**
- **Q-E architecture-change signals**: packages touched; `platform/` vs
  `actions/`; exported-signature diffs vs the corpus; message/topic/schema
  changes; migration presence. Which are load-bearing?
- **Q-G reviewer context**: per-reviewer contextkit bundles vs one shared bundle
  + role prompts vs curated RAG corpora per specialist.
- **Q-H the human-facing result**: PR link + diagnosis + council report + task
  notes link — what exactly lands, and where.
- **Q-D sub-question**: where the `hard_veto` flag physically lives (reviewer
  definition column vs per-pipeline council config vs both; most-specific wins?).

## LOOP-WORTHINESS TEST (doctrine — apply before every intake)
A task is loop material when ALL hold: (1) it is a SYMPTOM about system
behaviour, not a feature request; (2) a causal mechanism plausibly exists in
code + data + runtime; (3) it is not answerable by one or two direct queries —
**run the cheap pre-check first**; (4) it is bounded to one symptom; (5) the
symptom is verified CURRENT at intake.

**AMENDMENT, 2026-07-09.** Three candidates have now been dissolved by their
pre-checks (chrome: fixed before start; nav-to-unbuilt-pages: root cause found
in two files; guides: fully diagnosed in a dozen queries). Criterion 3 is doing
almost all the work, and it keeps saying no. The honest reading is that on this
platform bug mechanisms tend to be **legible to schema access plus grep**, so
"the loop discovers what a human could not" is the wrong value proposition.
The right one is *unattended, cited, consistent* diagnosis. Accordingly a
dissolved candidate is no longer discarded — a bug with a **known answer** is
promoted to a **benchmark** for grading the loop. See PLAN_fixloop_pilot.md §3.

## ★ F0 PILOT — DIAGNOSIS CLOSED BY PRE-CHECK 2026-07-09

**SYMPTOM (the intake string, unchanged — use verbatim for the benchmark run):**
"dartsonline.com published a Guides nav link and a /guides/index.html page, but
the page is blank and no guide pages exist — while gamesdesign.co.uk, on the
same platform, has working guides (and games and tools), and gaswholesalers.com
has a working news feed."

Site `5fe8785b-223d-41a3-88ee-c07187622381`. Built by the relay 2026-07-06.

### It was never a guides bug
Ten of dartsonline's fifteen pages were never built. `content` (3) and
`landing` (2) are `deployed`; `blog-post` (4), `entity-directory` (2),
`entity-page` (2), `section-index` (1) and `tool` (1) are all `planned`.
Guides is merely the one that also got a nav link.

### The mechanism (three causes, chained)

**A — the planner under-populates `sections`.** `build-site-planner` wrote 15
rows to `site_plan_pages` but authored `sections` for only the 5
`content`/`landing` pages. `jsonb_array_length(pages.sections) > 0` ⟺
`build_status='deployed'` is an **exact partition, 5 v 10, no exceptions**.

**B — a success-labelled error terminal swallows the drop.** The live
`page-build-handler` workflow (agent_definitions, v1, active) runs
`plan_sections` → `check_has_ready_sections`, a `conditional` with
`condition: "section_plan.ready_count > 0"`, `else_step: "complete_error"`.
`complete_error` is **`action: complete_workflow`** — a *success* terminal —
`success_message: "Content writer skipped — page has no sections defined"`,
`output_fields: ["page_content","site_record"]`. The dispatch loop then stamps
the work item `status='complete'`. The page row is never touched
(`updated_at == created_at` on all ten). The observed `result` payloads confirm
the path: the 5 built pages carry `deploy_result` with written files; the 10
carry only `site_record` (or a bare design-tokens blob) — precisely
`complete_error`'s `output_fields`.

*The platform already knows.* `load_work_item_actions.go:750-756` states it:
"the dispatch loop calls complete_work_item on every successful handler saga,
and page-build-handler's complete_error is a SUCCESS-labelled complete_workflow"
— and names the remedy, `mark_no_sections`, which would flag the item
`needs_human_review`. **`mark_no_sections` does not exist**: not in the live
workflow's 18 steps, and nowhere in the repo but that comment. The completion
guard at `:759-766` preserves a flag nothing ever sets.

**C — nav is grounded in the wrong column.**
`populate_nav_tables_action.go:242-243`:
`FROM pages WHERE site_id = $1 AND status IN ('active','deployed','pending')`.
`pages.status` is a lifecycle column defaulting to `'active'`; `build_status`
is never consulted. `guides-index` (`build_status='planned'`, `status='active'`,
`in_header=true`, `page_type='section-index'` — absent from the
`neverPrimaryTypes` set `{blog-post, tool, entity-page}`) is published into the
primary nav. **This is "the system linked to something it never built."**

### The standing hypothesis is REFUTED
It named `reconcile_site_plan`'s routing table. The table is real but lives in
**`WriteBuildItemsAction`** (`load_work_item_actions.go:218-228`), and absence
from it **does not drop a page**: `:239` defaults `handlerAgent =
"page-build-handler"` and `:283` warns `"Unknown page_type, using
page-build-handler"` before falling through. What drops pages is the separate
`unavailableBuilders` map (`:233-237`: `tool`, `entity-directory`,
`entity-page`) whose branch hits `continue // Skip — don't create a dispatch
work item`. Meanwhile `reconcile_site_plan_action.go:213-217` hardcodes
`handler_agent='page-build-handler'` for **every** plan page with no type switch
at all. The guides nav link is caused by **B + C**. A loop that *confirms* the
standing hypothesis has failed the benchmark.

### The differential, explained
gamesdesign's `guide` and `section-index` pages each carry `sections` and are
`deployed` — **through the same `page-build-handler`**. The handler is not the
discriminator; `sections` is. The build-route variable the runbook told us to
establish by evidence is now established: gamesdesign's items are
`needs_content_page` (the `availableBuilders` path in `WriteBuildItemsAction`);
dartsonline's are `needs_page` (emitted by `reconcile_site_plan`).

### Fourth finding (unlooked-for): the two intake paths disagree
`WriteBuildItemsAction` skips `tool`/`entity-directory`/`entity-page`;
`reconcile_site_plan` emits `needs_page` for them regardless. dartsonline's
`shop-index`/`brands-index` have completed work items that could never have
built anything. The paths must agree; which way is a builder-thread decision.

### Corrected pre-check SQL
The (9) runbook's second query referenced `attempts`; the column is
`attempt_count`. Working set:
```sql
-- the partition that cracks it open (run this one first)
SELECT name, page_type, build_status,
       jsonb_array_length(COALESCE(sections,'[]'::jsonb)) AS n_sections
FROM pages WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381'
ORDER BY n_sections DESC, name;

-- what the handler returned: deploy_result present ⟺ built
SELECT item_key, status, LEFT(result::text,120) AS result
FROM site_work_items
WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND item_type='needs_page'
ORDER BY item_key;

-- THE DIFFERENTIAL: same handler, but sections present
SELECT s.domain, p.name, p.page_type, p.build_status,
       jsonb_array_length(COALESCE(p.sections,'[]'::jsonb)) AS n_sections
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE s.domain='gamesdesign.co.uk' AND p.page_type IN ('guide','section-index')
ORDER BY p.page_type, p.name;
```

### Pilot criteria (revised)
(1) intake via the documented route; (2) per-iteration bundles fetchable from
`diagnosis_artifacts`; (3) per-iteration notes written; (4) **the emitted
diagnosis scores a pass against the pre-registered rubric in
PLAN_fixloop_pilot.md §3** — replacing the old "reaches a cited mechanism",
which we can now grade rather than eyeball; (5) stretch: F1 emits a constrained
edit plan on a branch, targeting the **platform** (the missing
`mark_no_sections` step; the nav column) and not dartsonline's data.

**BLINDING IS MANDATORY.** This runbook, NOTES(10) and PLAN contain the answer.
Exclude `docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/` from
the loop's corpus and `-doc` selection before the benchmark run, or the result
means nothing.

## REGENERATING THE CONTEXT BUNDLE (contextkit) — verified end-to-end 2026-07-09

The original `BUNDLE_fixloop_F0.md` was built without `-psql` and carries code +
docs only. This is the exact, tested procedure to rebuild one *with* the database
half. Total time ≈ 23 s.

### Where the tool actually lives
**Not** `cmd/bundle` in the chassis. contextkit is a separate Go module vendored at:

```
docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/go_files/contextkit
```

`bundle` is a thin wrapper that shells out to **`go run ./cmd/dbcontext`** and
**`go run ./cmd/assembler`** using *relative* paths. **You must `cd` into that
directory first** or it fails to find them. `dbcontext` is the only component
that touches SQL (bounded, read-only); the assembler never opens a connection.

### Step 1 — regenerate the analysis index
The chassis stores archived copies of its own source under `docs/`, so without
`-exclude` the index double-counts symbols.

```bash
cd docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/go_files/contextkit

go run ./cmd/analyser /home/ant/projects/agentchassis \
  -exclude 'go_files_old/,go_files/,docubundle/,thin_slice_run/,scripts/documentation_project/,docs024_key_docs_latest/,docs/_archive/' \
  > /tmp/analysis_repo.json
```

*Verify* (do not skip — a leak silently inflates the corpus):
```bash
python3 -c "import json;d=json.load(open('/tmp/analysis_repo.json'));p=[f['path'] for f in d['files']];\
print(len(p),'files;','docs024 leaks:',sum('docs024' in x for x in p),'go_files leaks:',sum('go_files' in x for x in p))"
# expect: 468 files; docs024 leaks: 0 go_files leaks: 0
```

### Step 2 — dry-run, then build
`-dry-run` prints the `dbcontext` and `assembler` commands it *would* run and
exits. Use it whenever you change flags.

```bash
go run ./cmd/bundle \
  -analysis /tmp/analysis_repo.json -root /home/ant/projects/agentchassis \
  -constitution thin_slice_constitution.md -step debug \
  -psql 'kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db' \
  -schema-tables pages,site_plans,site_plan_pages,site_work_items,page_components,site_specs,sites,agent_definitions,diagnosis_artifacts \
  -runtime-site dartsonline.com \
  -capabilities \
  -task "one sentence stating the bug or slice" \
  -scope platform/orchestration/actions/populate_nav_tables_action.go \
  -scope platform/orchestration/actions/reconcile_site_plan_action.go \
  -scope platform/orchestration/actions/load_work_item_actions.go \
  -include platform/orchestration/actions/registry.go \
  -out /path/to/BUNDLE.md
```

**`-psql` is ONE quoted argument and must NOT contain `-it`/`-t`.** It is passed
through to `dbcontext` as a single argv element and run via `exec.Command` with no
shell; a TTY either errors ("input device is not a TTY") or corrupts the captured
output. Without `-psql`, the tool prints a warning and silently produces a
code-and-docs-only bundle — which is exactly how F0's bundle came to be deficient.

`-schema-tables` **must include `site_plans` and `site_plan_pages`** — the original
invocation omitted both, and the guides evidence lives there. Add
`diagnosis_artifacts` now that F0.1a has landed.
`-include` forwards a file to the assembler as a whole-file `-scope`.

### Step 3 — verify the DB half actually landed
```bash
grep -c "_none provided\|not available in the thin slice" BUNDLE.md   # expect exactly 1
grep -n '^## ' BUNDLE.md
```
Expect these sections populated: `Recent errors (agent_error_log)`,
`Work-item lifecycle (site_work_items)`, `Schema`, `Database capabilities`,
`Installed extensions`, `Functions`, and the code sections.

**Known wart, do not be misled:** the heading `## Runtime evidence` *always* reads
"not available in the thin slice". `bundle` feeds `dbcontext`'s runtime output to
the assembler as a `-doc`, so the real runtime rows appear under **"Recent errors"**
and **"Work-item lifecycle"** instead; the assembler's own `-runtime` slot is never
wired by the wrapper. One placeholder is correct; two or more means `-psql` was
skipped or a gather failed. Watch stderr for `gathered schema|capabilities|runtime`.

Verified output 2026-07-09: `z_bundles/BUNDLE_fixloop_F1.md` (306,897 B, 468 files
analysed, all three gathers succeeded). The old, deficient
`z_bundles/BUNDLE_fixloop_F0.md` (199,579 B) is left in place for comparison.

### BLINDING — what it does and does NOT cover
The context bundle is for a **human or chat**, not for the loop. The live
`diagnose-agent` never reads it: it runs `analyse_repo_local` + `lookup_code_symbols`
over a checkout at an explicit ref, and `diagnose_assemble_bundle` reads only the
**bodies of in-scope Go symbols**. The analyser walks Go source; the live
diagnose-agent workflow has **no doc step** (verified). So the markdown in this
directory is structurally unreachable by the loop — blinding is largely automatic.

The two things that *can* leak the answer into a benchmark run, and must be checked:
1. **the symptom string** — pass the original one verbatim (★ F0 PILOT above), which
   describes only what a user could observe;
2. **`seed_scope`** — run the benchmark with **no seed scope at all**. Seeding it
   with `populate_nav_tables_action.go` or `load_work_item_actions.go` hands the loop
   the answer and makes the result meaningless. Absent a seed, the fallback chain
   uses `lookup_code_symbols`' `code_results`, which is the honest starting point.

## Inherited gotchas (diagnose-relevant subset)
- Loop core is READ-ONLY by contract; `sqlguard` allowlists reads — keep it so.
  The F1 write surface is a separate agent with isolated credentials.
- Result contracts: `result_from`/`output_fields` both live post-Option-A; the
  response size guard (`max_response_bytes`) exists — bundle egress via
  completion payloads is BOUNDED; persist and reference, don't ship megabytes.
- `agent-type` (hyphen) pod label; `is_active` gates spawn; seeds must copy
  image columns from a live donor; check body.status not just header status.
- Schema before SQL; `snapshot_agent` before `agent_definitions` updates;
  0 rows isn't decisive until the query is checked; explicit git refs.
- **`error_step` goes INSIDE a step's `config`** — step-level `error_step` is
  silently ignored (001 §16). Live dormant instances confirmed in
  `page-build-handler` (`deploy_page`, `plan_sections`, `save_sections`,
  `validate_content`, `load_spec_sections` carry it at both levels). Correct
  adjacent instances when touching a workflow, as a noted change.
- Idle pods reap ≈3600s; a ProcessingHistory state dump is the accepted
  post-reap evidence substitute.
- `site_work_items.attempt_count`, not `attempts`. `pages.status` (lifecycle,
  default `'active'`) is **not** `pages.build_status` (build state). This
  distinction is itself the nav bug — expect it to bite elsewhere.
- `site_work_items.site_id` is **NOT NULL**, and `LoadWorkItemsAction` is
  site-anchored (`WHERE wi.site_id = $1`). Platform-wide work anchors to the
  `system.internal` pseudo-site. There is no site-less work item.
- **Nothing in the relay filters work items by pipeline where it matters.**
  `build-dispatch-loop`'s `load_items` config is `{site_id, max_items}` only.
  `build-pipeline-trigger`'s `find_dispatchable_site` has no pipeline filter
  either, despite its description saying "pending *build* items". Any item of
  any pipeline, on a selected site, at `status IN ('triaged','approved')`, is
  claimed and handed to its `handler_agent`. **The `maintenance` pipeline is
  dispatched only because of this accident** — so adding `item_pipeline='build'`
  to `build-dispatch-loop` would orphan it. Builder thread's call.
- **`triage_detect_items` is not pipeline-safe**: it promotes
  `WHERE site_id=$1 AND status='detected'` and **rewrites `pipeline` to
  `'build'`**. Its own comment asserts the dispatch loop filters on
  `item_pipeline='build'`; it does not. So `'detected'` is not safe parking for
  a non-build item.
- **`claimed-item-timeout` resets any `claimed` item older than 40 minutes to
  `'triaged'`** (or `'failed'` once `attempt_count+1 >= max_attempts`), and
  `find_dispatchable_site` excludes any site holding a `'claimed'` item. So
  `'claimed'` is a shared, swept status — a long-running non-build handler
  should not sit in it.
- **Pattern for a private pipeline**: give it statuses no sweep names
  (`diagnose` uses `awaiting_diagnosis` → `diagnosing`), claim atomically with
  `UPDATE … FOR UPDATE SKIP LOCKED … RETURNING` rather than `claim_work_item`
  (which only claims `triaged|approved`), and **reap your own dead runs** —
  opting out of the shared sweeps means opting out of their cleanup too.
  The dispatcher claims, not the handler: `page-build-handler` neither claims
  nor completes its own item.
- `snapshot_agent(type, reason)` writes to **`agent_definitions_backup`**, not
  to `agent_definitions` with `is_snapshot=true`. Verify there.
- **`orchestration_states.last_activity` is `timestamp WITHOUT time zone`** while
  `created_at` is `timestamptz`. Any `NOW() - last_activity` arithmetic is wrong
  by the UTC offset — the `idle_s` query printed by 084/090 included. And this
  dev host runs **BST (+0100)** while the DB is UTC, so `stat`/`date` output and
  DB timestamps differ by an hour. A run can look like it finished before it
  started.
- **`code_symbols` indexes `.go` files only.** Workflow definitions live in
  `agent_definitions.default_config` as JSON and are therefore INVISIBLE to the
  loop's static tier — even though they contain load-bearing control flow (the
  pilot's cause B is a workflow step). Reach them with a `data_request`.

## BENCHMARK RESULT — 2026-07-09 (first run; correlation 4d43d002-671f-496f-a64a-c3bb8ffe35e2)
Plumbing **passed**: intake via the documented route; five per-iteration bundles
persisted and fetched back out of `diagnosis_artifacts`; the terminal note landed
in `doc_notes` (`pipeline/build`, category `diagnosis`) proving Q-F's integration.
Per-iteration notes were not written — F0.3 is not built.

Rubric: **0 of 4 must-claims**, verdict `CONFIRMED` regardless. It passed the
refutation check (never blamed `reconcile_site_plan`) and contributed one *new,
independently verified* finding: the four "guide" pages are `blog-post` rows at
`/blog/*.html` with `site_area_id IS NULL`, so `resolveSectionIndexForType` could
never bind them to `guides-index` — they would not have appeared under `/guides/`
even if built. **Add that as rubric claim 9.**

Three engine defects it exposed, in value order:
1. **No symptom-explanation gate.** Cite-or-abstain permits a CONFIRMED verdict
   whose cited cause does not explain the reported symptom. Every citation was
   real; the conclusion never accounted for why a nav link was published.
2. **Static tier is Go-only** (see the corpus gotcha above), so the causal
   workflow step was unreachable.
3. **Symbol-granular retrieval overshoots.** Iteration 1 scoped
   `populate_nav_tables_action.go:isLegalPage` (a trivial helper) while the bug
   sits in `loadPagesForNav` in the same file; that function entered no bundle.
   When retrieval implicates a file, offer the verdict its sibling symbols.

Also: `page-build-handler/complete_error (complete_workflow) fatal: …` appeared
verbatim in the `agent_error_log` section of **all five** bundles. The engine
does not enforce the runbook's own "FOLLOW what the evidence names" principle.

## BENCHMARK RUN 2 — 2026-07-09 (corr `dd1186b9`; after F0.4 a/b/c/e on v1.0.1101)
Same symptom string, no seed scope, site data untouched. 5 iterations, ~18.5 min,
CONFIRMED, terminal note persisted. **Score moved from 0/4 musts to 1 pass +
2 partial + 1 fail**: claim 3 (the success-labelled `complete_error`) is now
**cited**, the static citation sourced from the new agent_definitions enrichment
section — run 1's structurally unreachable cause became run 2's quoted evidence.
Claims 1–2 partial (mechanism described; site-wide partition and the routing
conditional not established). Claim 4 **failed and was actively dismissed**
("not a nav issue") while confirming — the compound symptom's nav clause is the
unexplained residue, which is the empirical case for F0.4d (symptom-closure
gate, run 3's variable). One false side-claim ("those pages built") shows
periphery statements still escape the citation discipline. Refutation credit
passed in both runs. Two verification gotchas for future runs: fire NOTHING into
a settling rollout (run-2 attempt 1 died silently at spawn in the chassis pod's
rebalance window); and the completion signal (`diagnosis` in collected_data)
precedes persist_note by ~2 min — don't read doc_notes in that gap.
- `cmd/bundle` needs `-psql` as ONE quoted argument with NO `-it`/`-t`;
  without it the bundle silently carries code+docs only.
- **`max_tokens` at a step-config's root is DEAD CONFIG** for
  `execute_llm_prompt`: the action reads it from the agent's top-level config
  or from inside the step's `ai_service` block only (ai_actions.go:252-256),
  and the Anthropic client then defaults to **2048 output tokens**. The
  diagnose-agent verdict step ran capped at 2048 through all five benchmark
  runs; the fix-proposer's plan was truncated mid-JSON twice before this was
  found (byte counts ~7.5-8.1KB ≈ 2048 tokens — the tell). Put `max_tokens`
  INSIDE `ai_service`. Same family as 001 §16's error_step placement bug:
  config that looks right, parses fine, and is silently ignored.

## BENCHMARK RUN 3 — 2026-07-10 (corr `5120c0dc`; F0.4d live on v1.0.1102)
Honest UNVERIFIABLE in 3 iterations (~16 min): iteration 2's state-only
CONFIRMED was **coerced by the tier guard in production** (its exact message in
the trail); the final output names `complete_error` + `sections=[]` as prime
suspect and instructs "hand to a human; do NOT auto-conclude". Primary
criterion (no confident half-answer) PASSED; no symptom_check gaming observed.
Third engine defect found: **data_request answers are one-shot** — they appear
only in the bundle immediately after the requesting verdict, so a guard-refused
confirm loses the fetched evidence, the loop re-requests, and
scope-not-narrowing fires. Fix = F0.5 (persist answered requests across
iterations). Run arc: run 1 wrong-confidently → run 2 half-right-confidently →
run 3 honest abstention with the mechanism in its sights.

## BENCHMARK RUN 4 — 2026-07-10 (corr `5179a2ea`; F0.5 live on v1.0.1103)
**The first full-coverage CONFIRMED**: 2 iterations, ~8 min, citation + tier +
closure guards all passed legitimately; five-entry `symptom_check` rendered in
the conclusion. Blank-page chain correct and cited end-to-end (sections=[] →
complete_error success terminal → unbuilt → blank; the step definition quoted
via the F0.4b enrichment). Nav clause now visible in coverage but SHALLOW
("the nav row exists") — the gate enforces declared coverage, not depth;
must-claim 4 (`loadPagesForNav` on `status` not `build_status`) remains the
one unreached mechanism across all four runs. Control clauses
(gamesdesign/gaswholesalers) were marked explained while self-described as
unverifiable → F0.6: a `context` disposition + citation-backed `explained`.
Four-run arc: wrong-confidently → half-right-confidently → honest abstention →
right-with-declared-coverage. F0 functionally complete bar F0.3.

## CURRENT POSITION — 2026-07-09
F0/F1 design questions all decided (2026-07-07). The pilot bug is **fully
diagnosed ahead of the loop** and has been reframed as a known-answer
benchmark. F0.1's three slices (artifacts migration, assemble write-through,
`needs_diagnosis` envelope) are unblocked and independent — start there.
Q-E/G/H open the F2 discussion when reached. The four findings above should be
relayed to the builder thread (causes A + the intake-path disagreement) and
carried into F1 (causes B + C).
