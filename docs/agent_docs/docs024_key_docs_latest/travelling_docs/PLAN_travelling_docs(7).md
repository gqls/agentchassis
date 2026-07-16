# PLAN — Travelling Docs (PLAN + NOTES) for Tools, Complex Components, and Pipelines

**Created:** 2026-07-04
**Last updated:** 2026-07-16 (rev 7 — THE WHOLE LOOP IS PROVEN GREEN ON A REAL BUG. Tier-4 P1 (mobile) + P2 (interactions) live; behavioural attribution (tool vs site-chrome) + routing; a genuine site-chrome footer overflow found, routed, fixed at the DURABLE content_component layer, deployed, and re-verified `mobile-fit@mobile` GREEN. See the new "Completed loop" section and the RUNBOOK §0 position line. Prior rev 5 line below.)
**Prior:** 2026-07-10 (rev 5 — Phase A write-hooks COMPLETE AND PROVEN: PLANs at birth (Task 3, run `1923badd`) and NOTES at every fix (Task 4, two machine `fix` notes from the economy-simulator recreation). KB indexing gated on the chunkContent fix deploy — see Rollout outcomes.)
**Status:** **COMPLETE end-to-end and proven on a real defect** — birth-PLAN → continuous discovery → Tier-4 (desktop+mobile, interactions) → attribution → routing → durable fix → deploy → re-verify. Remaining work is polish (P3 screenshots; per-site override option for shared-template fixes). This document is the spec; live position lives in the RUNBOOK §0 tracker; the blow-by-blow lives in the HANDOFF Turn log (T14–T17).

---

## Completed loop (2026-07-16) — what "done" turned out to mean

The self-verifying tool loop is whole. Proven on a genuine, non-manufactured bug
(not a manufactured test):

1. **Birth PLAN** — the composer writes acceptance criteria at tool creation, and
   (migration 149) now emits a `container` selector copied from the generated
   HTML so a later behavioural run can tell where the tool ends.
2. **Continuous discovery** — a scheduled sweep (`tool_acceptance_due`) queues an
   `acceptance_run` for any deployed tool with criteria that is due.
3. **Tier-4 behavioural run** — `browser-runner-adapter` drives real headless
   Chromium on the live page, **desktop + mobile**, running P0 (status/selector/
   console), P1 (`no_horizontal_overflow`) and P2 (`interaction`: fill/click/
   select then assert). Every result is labelled `id@profile`.
4. **Attribution** — a document-level failure (overflow) is attributed to the
   TOOL or to SITE CHROME by asking the browser which element overflows and
   whether it lies inside the tool's `container`. `scope` = tool | chrome |
   unknown; `unknown` never blames chrome.
5. **Routing** — a tool defect → `improve_tool` (tool-improver); a site-chrome
   defect → `responsive_fix` / `chrome_overflow_fix` (component-template-fixer),
   deduped per (component, profile) so one footer bug is ONE site ticket.
6. **Durable fix** — `chrome_overflow_fix` patches the DURABLE layer
   (`content_components.html_template`, resolved via the slot's backing
   component), so it survives `refresh_site_components`; it falls back to the
   rendered artifact only when a slot has no component, reporting that honestly as
   TRANSIENT, and reports the shared-site blast radius.
7. **Deploy + re-verify** — a rerender deploys the fix; re-running Tier-4 confirms
   the once-failing check now PASSES.

**The two lessons the arc banked (durable):**
- *"Completeness + validation passed" ≠ working, at every layer.* Tier 4 caught a
  FIXER that reported `fixed:true` and changed nothing (it defaulted to the wrong
  slot). Behavioural re-verification is the only thing that proved the fix real.
- *Fix the source, not the artifact.* `site_components.rendered_html` is
  regenerated from its content_component template; a patch there is wiped by the
  next refresh. The durable fix lives in the template.

---

## What we are aiming to achieve

Give every tool/complex component — and every pipeline — its own **travelling
documentation**: a `PLAN` (intent) and a `NOTES` log (history), keyed by a stable
subject key, so that whenever an agent or a human fixes something or the next
improvement-loop cycle touches it, it **loads the subject's intent and history
first**. For tools this includes a per-tool definition of *working* (acceptance
criteria) that verification holds still across iterations — up to and including
driving the deployed tool in a headless browser on desktop and mobile until the
criteria pass (see `PLAN_tool_acceptance_runner.md`).

This **extends** the tool-doc header system (shipped 2026-06-11; 019 §Tool Doc
Header). The inline header stays the short audit/parse anchor; PLAN and NOTES are
the substantive intent and the missing change-history.

---

## Framing (agreed 2026-07-04) — where each artifact sits

- **Plan** = enforced desired state (site_plans + specs; the reconciler drives
  realised state toward it — "the plan table is ground truth").
- **Pipeline** = the compiled happy-path runbook (workflow SQL + Go actions).
- **Runbook docs** = the un-compiled residue: exception knowledge, failure lore,
  judgment. Entries retire when compiled into guards/fixes.
- **NOTES** = the reasoning log nothing machine-side captures (why, dead ends,
  iteration memory). `diagnose_emit` deliberately persists nothing — that gap is
  what `persist_diagnosis_note` now fills.
- **Contracts/constitution (003, guidelines)** = admission rules above all of it.

**Graduation rule:** prose → structured → enforced, only when recurrence proves
the need. (Locks stay deferred; criteria live as a fenced block, not a column,
until a checker consumes them at volume.)

---

## Reuse baseline — what already exists (do not rebuild)

| Piece | Where | What it gives us |
|---|---|---|
| Inline tool-doc header | `platform/content/tool_doc_header.go`; gate in `create_tool_component`; 019 | `function`/`purpose`/`behaviour`/`inputs`/`outputs` anchor; stripped at deploy. |
| Generation-time completeness | `check_tool_completeness` (marker `<!-- tool-recreation-complete -->`, balanced script/style, length; flags-but-passes) | Tier 0 of the ladder — output integrity before deploy. |
| Creation provenance | `content_components.source_agent_type`, `source_orchestration_id` | Who/what created the component — provenance is data. |
| Versioning pattern | `site_specs` + supersede log (`is_current`, `superseded_at`, `source*`, `pinned`, `notes`) | Proven Postgres document versioning; reused re-keyed in `doc_plans`. |
| Per-iteration acceptance | 004 findings: each requires `current_value`, `acceptance_test`, `suggestion`, `max_fix_attempts` | The iteration-scale done-ness mechanism the standing criteria SEED. |
| Site-level durable musts | `direction.must_have` (auditor-loaded) | The site-scale analogue; home for per-site tool parametrisation. |
| RAG index + retrieval | `rag_index`/`rag_lookup` (`rag_actions.go`), `knowledge_base` | Chunk→embed→store + vector/trigram lookup; the derived index. |
| Diagnosis loop output | `diagnose_emit` (status, conclusion, evidence_trail) | The producer `persist_diagnosis_note` persists. |
| Adapter mould | analyser adapter (`request_repo_analysis` → adapter → response topic) | The shape of the browser-runner adapter (Tier 4). |
| Callgraph/derivation | `callgraph.go`, `agent_definitions` | Pipeline topology is derivable — generate, don't author. |
| Pipeline identity | `site_work_items.pipeline` (text NOT NULL DEFAULT 'build', **no CHECK** — convention) | The runtime key pipeline docs attach to; live values to confirm. |
| Migration ledger | e.g. 005 "SQL Migrations Applied" table | Pipeline NOTES in embryo — the write-hook precedent. |
| Lifecycle hooks | `create_tool_component`; `deploy_tool_to_site`; `tool-improver`/`update_component_html`/`component-template-fixer`; `tool-auditor`; `check_tool_health` | The steps docs are written/updated/consumed at. |

---

## Storage decision — DB is the source of truth; git is an optional mirror

(Unchanged from rev 2; evidence: the git commit path rejects empty `Domain`,
force-prefixes `{domain}/`, has no file-read, whole-file commits, no conflict
retry, single serialised adapter; `knowledge_base` is content-addressed with no
version chain — a derived index's shape.)

- **Source of truth = Postgres** (`doc_plans`, `doc_notes` — migration drafted,
  `drafts/0NN_doc_plans_and_notes.sql`).
- **Retrieval index = `knowledge_base`** via `rag_index` (`collection='tool_docs'`
  / `'pipeline_docs'`), `rag_lookup` for discovery (no key filter).
- **Git = optional non-authoritative mirror** for human browsing/diffs.

## Table design (drafted; see migration)

`doc_plans` — supersede-versioned intent, one `is_current` row per
`(subject_type, subject_key)`; partial unique index enforces it.
`doc_notes` — append-only history, one row per entry, `categories jsonb` with
GIN (`jsonb_ops` so `?` is indexable), `site_id` for per-site incidents.
Subjects: `('tool', content_components.function)`, `('pipeline',
site_work_items.pipeline value)`. Both carry `source`/`source_agent`/
`source_item_id`/`created_by` provenance.

---

## Where acceptance criteria live (decided 2026-07-04)

Candidates judged on **key, lifecycle, owner**:

- **`site_specs`** — right machinery (supersede/pin, auditor-loaded), wrong key:
  everything is `(site_id, aspect)`; tool criteria describe a `function` shipped
  to many sites. Per-site duplication drifts. `direction.must_have` stays the
  **site-scale** durable-musts home — and is where rare **per-site tool
  parametrisation** goes ("this site's calculator uses GBP").
- **`site_plans`/directives** — wrong lifecycle and owner: the churniest artifact
  (superseded per re-plan, re-derived by sync — see the plan↔pages divergence
  saga), written by the planner LLM. Criteria exist to hold the bar still;
  never store the bar in the artifact that regenerates most.
- **Findings' `acceptance_test`** — right pattern, wrong duration: per-iteration
  done-ness that dies with the work item. The standing criteria **seed** it.
- **The tool's `doc_plans` PLAN — chosen.** Matches on all three axes: keyed by
  `function` (travels), supersede-versioned + pinnable (the same machinery,
  re-keyed), owned by creation + humans (not the planner), loaded at exactly the
  consumption moments (fix time, audit time, acceptance run).

Tool criteria differ in kind from site criteria (behavioural assertions checked
against deployed pages by a browser/DOM check vs qualities checked over content
by an LLM auditor) — the second reason not to co-locate.

**Format:** a machine-extractable fenced ` ```criteria ` JSON block inside the
PLAN body under "## Acceptance criteria" (the tool-doc-header precedent: a
structured block parsed from a larger artifact). `load_doc_context` extracts it
as `criteria_json`. Schema v0 in `PLAN_tool_acceptance_runner.md`. Lifts to a
column only on volume (graduation rule).

---

## Rollout outcomes that amended this spec (2026-07-09/10)

- **Recreation notes are pipeline-scoped.** `tool-recreation-handler` writes
  page sections and never creates a `content_components` row, so its NOTES
  subject is `('pipeline','build')` + `note_site_id`, NOT `('tool',
  spec.function)` — a tool-keyed note there would dangle. General rule: the
  subject must be something the agent actually owns/creates (migration 137).
- **The seam rule for spec-carrying prompts.** A requirement can survive the
  analysis step and still be ignored at generation if the generation prompt
  never renders it: `analyze_tool` rendered `spec.interactive_features`,
  `recreate_tool` didn't, and the model trusted the visible source HTML over
  the buried analysis JSON. Every prompt that consumes a spec field must
  render it explicitly, marked as overriding the source (migration 138 added
  "Mandatory Behaviour Requirements" to `recreate_tool`).
- **"Passed checks" ≠ "working" — twice demonstrated.** The June recreation
  introduced the economy-simulator's two bugs and passed; run 2 of the repair
  faithfully recreated them and passed. Tier 0/validation measure output
  integrity, not behaviour. This is the standing argument for Tier 2 (cheap
  static confirm, anchor rule) and Tier 4 (behavioural) — build them next.
- **`rag_index` reuse caveat closed.** `chunkContent()` looped forever on
  content > chunk_size (cause of both chassis OOMKills, both on PLAN-sized
  bodies through `index_plan`). Fixed with regression tests; `index_plan` is
  bypassed (migration 140) until the fixed image deploys, then 141 re-enables.
  The `tool_docs` KB write — this spec's derived index — first happens with
  that proof run.
- **Migrations are now a system, and they write pipeline NOTES.** Numbered
  files in `sql_for_agents/` (baseline 124), `schema_migrations` ledger,
  `scripts/migration/run-migrations.sh`. The travelling-docs arc is 125–144.
  The "NOTES — from workflow-altering migrations" write-hook below is live
  practice from 140 onward, not aspiration.
- **Criteria describe DELIVERED reality, not aspiration (decided 2026-07-10,
  Option B).** The composer asserted a designed-but-never-built JS extraction
  (`asset_loads /tools/assets/<fn>.js`) in every PLAN; Tier-2's first sweep
  failed every tool on it by construction. Remedy: PLANs superseded to inline
  delivery (143), composer emits four standard checks and an inline delivery
  line (144). An aspiration belongs in a roadmap; if it ships, PLANs
  supersede forward. Corollary: Tier-2's standard checks are now
  boots/console/status/mobile-fit (+ optional interaction from real
  selectors).

## Write hooks

- **PLAN — at creation / intent change.** `write_doc_plan` (drafted). Tools:
  from `create_tool_component`, drafted from the generator's reasoning (spec
  slice, delivery mechanism, deliberate decisions, **acceptance criteria**).
  Pipelines: authored initially (distil from 004–008), superseded on direction
  change. Not on `deploy_tool_to_site` forks.
- **NOTES — at modification.** `append_doc_note` (drafted) as the **last step**
  of the fix agents (symptom → root cause → fix → verification + categories).
- **NOTES — from the diagnosis loop.** `persist_diagnosis_note` (drafted),
  config-gated step **after** `diagnose_emit` (emit stays read-only). Explicit
  subject only — skip, don't guess. UNVERIFIABLE runs persist tagged
  `unconfirmed-diagnosis` (dead ends prevent retries).
- **NOTES — from workflow-altering migrations (pipelines).** Migration number,
  what changed, why.
- **NOTES — from acceptance runs (Tier 4).** Pass/fail per run;
  `acceptance-fail` entries carry the failing criterion.
- **(Optional wiring)** `check_tool_completeness` `complete=false` flags — today
  log-only — can append a note (`truncated-output`) once wired.
- **Index step (reuse).** `rag_index` after writes. **Mirror step (optional,
  Phase B).**

Keep workflows flat: each write is one Go action; no sub-workflows.

---

## Pipeline documentation

**Derive the topology; author the intent.** Step maps generate from
`agent_definitions` (callgraph pattern) — never hand-drawn. The authored
pipeline PLAN holds only: **end-to-end invariants** ("an interactive section
survives every rebuild route"), **branch rationale** (page-build-handler
re-resolves sources vs page-rebuild deliberately doesn't), **seams** (pipelines
sharing one handler), **deliberate decisions** (priorities, cooldowns,
NULL-as-stale). Pipeline NOTES = incidents + migration entries + persisted
diagnoses; 016/016b stays the global roll-up. Retrieval symmetry: pipeline scope
IS code, so `docselect` + `path_globs` fits pipelines (needs the git mirror for
files — Phase B); 004–008 remain the prose base for first PLAN bodies.

---

## Tool assurance: contracts, acceptance, iteration

- **Behaviour contract = acceptance criteria** in the PLAN (fenced block).
  Multi-page tools also state the **page set + per-page roles** and the
  **inter-page contract** (URLs, shared state keys, data feeds).
- **Verification ladder:**
  - **Tier 0 — generation-time output integrity (exists):** `HasToolDocHeader`
    gate + `check_tool_completeness` (completion marker, balanced script/style,
    length; deliberately flags-but-passes for review).
  - **Tier 1 — structural post-deploy (exists):** `check_tool_health`.
  - **Tier 2 — contract-presence (Phase A):** thin check asserting criteria
    selectors/assets against deployed HTML (catches `empty-shell`,
    `detool-on-rebuild`, `js-not-extracted`).
  - **Tier 3 — acceptance audit (Phase B):** `tool-auditor` extended to judge
    deployed pages vs PLAN criteria; failures spawn improvement items carrying
    the failing criterion (findings `acceptance_test` pattern).
  - **Tier 4 — behavioural, headless browser (ACTIVE PLAN):** drive the deployed
    tool on **desktop and mobile** profiles until criteria pass —
    `PLAN_tool_acceptance_runner.md` (adapter mould, Playwright, P0–P4).
- **Iteration loop.** deploy → acceptance run → failing criterion →
  `improve_tool` item (criterion as `acceptance_test`, bounded by
  `max_fix_attempts`) → fixer **loads PLAN+NOTES first** → fix →
  `append_doc_note` → re-run. A tool is *working* when its criteria pass.
- **Prerequisite before multi-page tools scale:** the pending preserve-sections
  re-render + interactivity-aware save guard (016b/005/020/026).

---

## Retrieval

- **Direct-by-key (primary, fix-time):** `load_doc_context` (drafted) — current
  PLAN + latest-N NOTES + extracted `criteria_json`. No-plan is a normal state
  (`has_plan=false`), not an error.
- **Code-diagnosis loop:** hand `doc_context` to `diagnose_assemble_bundle` the
  way `runtime_evidence` is handed in (one compose line).
- **Semantic discovery:** `rag_lookup` on the collections.
- **`docselect` catalogue:** pipelines, Phase B (needs the mirror).

## Deliberate-decisions — prose now, no locks
Unchanged: a PLAN section ("Deliberate decisions — do not re-fix") + NOTES
narrative; protective because loaded at fix time. Enforcement is the compilation
target if a regression class shows prose isn't honoured.

## Document formats

**Tool PLAN sections:** Aim · Source spec · Behaviour contract → **## Acceptance
criteria** (with the fenced ```criteria block; multi-page: Page set &
inter-page contract) · Delivery mechanism + why · Dependencies · Deliberate
decisions.

**Pipeline PLAN sections:** Aim · Invariants · Branch rationale · Seams ·
Deliberate decisions · (topology = derived, referenced not embedded).

**NOTES entry (uniform, dated, one row each):**
```
## 2026-07-04 — <short title>
Observed: <symptom, where>
Root cause: <cause — or "unconfirmed (<stopped_by>)" for a diagnosis trail>
Fix: <what changed> (<site_id> if per-site)
Verified: <how confirmed>
Categories: detool-on-rebuild, diagnosis, ...
```

**Category taxonomy:** rev-2 set + `diagnosis`, `unconfirmed-diagnosis`,
`migration`, `seam`, `acceptance-run`, `acceptance-fail`, `truncated-output`,
`needs_criteria`.

---

## Phasing

**Phase A (in rollout — tables live 2026-07-04; actions deploying; wiring next)**
1. Formats locked (above; criteria block v0 in the runner PLAN).
2. Migration APPLIED 2026-07-04 — gates clean (no collision; pipeline values
   `build`/`content`/`design`/`maintenance`; no CHECK); statement tally
   verified (2 tables, 5 indexes, 3 comments).
3. Actions ON PRODUCTION 2026-07-04: `write_doc_plan`, `append_doc_note`,
   `load_doc_context` (with criteria extraction), `persist_diagnosis_note` —
   workflow wiring migrations are the next step (Stages 3–4).
4. ✅ Wire — DONE AND PROVEN: tool-generator → PLAN at birth (Task 3, 2026-07-09);
   fix agents → note-append last step (Task 4, two machine `fix` notes
   2026-07-09); diagnosis agent → persist step (Stage 3, 2026-07-07).
   `rag_index` after writes is gated on the chunkContent fix deploy (140/141).
5. Tier-2 contract-presence check fed by `criteria_json` — **next up**; the
   anchor rule is settled (validate the leftmost id/class token only; static
   checks confirm, never refute; `-EDIT` ids skipped).
6. KB `tool_docs` write — becomes real with the 141 proof run (first rows).

**Phase B**
7. Tier-3 acceptance audit (`tool-auditor` extension).
8. **Tier-4 acceptance runner P0** (adapter skeleton, desktop boot checks) —
   see `PLAN_tool_acceptance_runner.md`; then P1 mobile profile.
9. Pipeline PLAN bodies distilled from 004–008; migration write-hook live.
10. Optional git mirror; `docselect`/assembler injection for pipeline docs.
11. Category roll-up query surfaced into 016/016b.

---

## Open questions / dependencies
- ~~Live-DB verifications~~ DONE 2026-07-04: no collision; live pipeline
  values = `build`, `content`, `design`, `maintenance`; no CHECK constraint.
- KB `tool_docs` write — confirm whether it exists (uploaded action lacks it).
- `deploy_tool_to_site` — confirm forks stamp `source_*`; NOTES-only on fork.
- `rag_index` `source_type` — parameterise or accept `'scrape'`.
- Acceptance runner P0 open questions — in its own PLAN.
- `site_plan_directives` — not present in the schema dump; unverified (not
  needed for the criteria decision, which stands on lifecycle/owner).

## Reuse ledger
site_specs supersede pattern · rag_index/rag_lookup/knowledge_base ·
source_* provenance convention · maintenance-mould read actions ·
diagnose_emit output (producer) · findings acceptance_test + max_fix_attempts ·
check_tool_completeness (Tier 0) · check_tool_health (Tier 1) · tool-auditor
(Tier 3) · analyser-adapter mould (Tier 4 runner) · callgraph/agent_definitions
derivation · docselect (pipelines, later) · content_components.function +
site_work_items.pipeline (keys) · inline tool-doc header (untouched) ·
lifecycle agents as write hooks · git adapter (optional mirror).
