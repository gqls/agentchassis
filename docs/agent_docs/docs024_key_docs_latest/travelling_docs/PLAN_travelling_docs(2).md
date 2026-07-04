# PLAN — Travelling Docs (PLAN + NOTES) for Tools, Complex Components, and Pipelines

**Created:** 2026-07-04
**Last updated:** 2026-07-04 (rev 3 — diagnosis→notes wiring folded into Phase A; tables generalised to subjects; pipeline docs; tool acceptance/testing)
**Status:** design agreed; not yet built. This document is the spec.

---

## What we are aiming to achieve

Give every tool/complex component — and now every pipeline — its own **travelling
documentation**: a `PLAN` (intent) and a `NOTES` log (history), keyed by a stable
subject key, so that whenever an agent or a human fixes something or the next
improvement-loop cycle touches it, it **loads the subject's intent and history
first**. The payoff is fixes that build on prior decisions instead of re-deriving
lost context and fighting earlier, deliberate choices — and, for tools, a per-tool
definition of *working* that acceptance checks can hold still across iterations.

This **extends** the tool-doc header system (shipped 2026-06-11; 019 §Tool Doc
Header). The inline header stays the short audit/parse anchor; PLAN and NOTES are
the substantive intent and the missing change-history.

---

## Framing (agreed 2026-07-04) — where each artifact sits

- **Plan** = enforced desired state (site_plans + specs at site level; the reconciler
  drives realised state toward it — "the plan table is ground truth").
- **Pipeline** = the compiled happy-path runbook (workflow SQL + Go actions).
- **Runbook docs** = the un-compiled residue: exception knowledge, failure lore,
  judgment. Entries retire when compiled into guards/fixes.
- **NOTES** = the reasoning log nothing machine-side captures (why, dead ends,
  iteration memory). `diagnose_emit` deliberately persists nothing today — that is
  the gap NOTES fills.
- **Contracts/constitution (003, guidelines)** = admission rules above all of it.

**Graduation rule:** prose → structured → enforced, only when recurrence proves the
need. (Locks stay deferred under this rule; acceptance criteria may later lift from
PLAN prose to a structured column when a checker consumes them at volume.)

---

## Reuse baseline — what already exists (do not rebuild)

| Piece | Where | What it gives us |
|---|---|---|
| Inline tool-doc header | `platform/content/tool_doc_header.go`; gate in `create_tool_component`; 019 | `function`/`purpose`/`behaviour`/`inputs`/`outputs` anchor; stripped at deploy. |
| Creation provenance | `content_components.source_agent_type`, `source_orchestration_id` | Who/what created the component — provenance is data. |
| Versioning pattern | `site_specs` + supersede log (`is_current`, `superseded_at`, `source*`, `pinned`, `notes`) | Proven Postgres document versioning; reuse re-keyed. |
| RAG index + retrieval | `rag_index`/`rag_lookup` (`rag_actions.go`), `knowledge_base` | Chunk→embed→store + vector/trigram lookup; the derived index. |
| Deterministic doc retrieval | `docselect.SelectDocs` + doc catalogue | Keyword/path-glob/always doc injection for the code-diagnosis loop. |
| Diagnosis loop output | `diagnose_emit` (status, conclusion, evidence_trail; deliberately not persisted) | An import-shaped NOTES entry, currently discarded — now a producer. |
| Findings pattern | Improvement-loop findings carry `acceptance_test` | Machine-checkable acceptance at work-item scale — the pattern tool criteria reuse. |
| Callgraph/derivation | `callgraph.go`, `agent_definitions`, code_symbols | Pipeline topology is derivable — generate, don't author. |
| Pipeline identity | `site_work_items.pipeline` (e.g. `'build'`) | The runtime key pipeline docs attach to. |
| Migration ledger | e.g. 005 "SQL Migrations Applied" table | Pipeline NOTES in embryo — the write-hook precedent. |
| Lifecycle hooks | `create_tool_component`; `deploy_tool_to_site`; `tool-improver`/`update_component_html`/`component-template-fixer`; `tool-auditor`; `check_tool_health` | The steps docs are written/updated/consumed at. |

---

## Storage decision — DB is the source of truth; git is an optional mirror

(Unchanged from rev 2; evidence: the git commit path rejects empty `Domain`,
force-prefixes `{domain}/`, has no file-read, whole-file commits, no conflict
retry, single serialised adapter; `knowledge_base` is content-addressed with no
version chain — a derived index's shape.)

- **Source of truth = Postgres tables** (below), written transactionally by the
  framework.
- **Retrieval index = `knowledge_base`** via `rag_index` (`collection='tool_docs'`,
  or `'pipeline_docs'`), queried by `rag_lookup` (discovery only — no key filter).
- **Git = optional non-authoritative mirror** for human browsing/diffs.

---

## Table design (rev 3 — generalised to subjects; PROPOSED rename, adopt unless vetoed)

Rather than `tool_doc_*` plus a later parallel `pipeline_doc_*`, one pair keyed by
subject. Confirm no name collision on the live DB before migrating.

```
doc_plans
  id             uuid pk default gen_random_uuid()
  subject_type   text not null check (subject_type in ('tool','pipeline'))
  subject_key    text not null    -- tool: content_components.function
                                  -- pipeline: site_work_items.pipeline value (verify enum)
  body           text not null    -- the PLAN markdown
  source         text             -- 'tool-generator' | 'human' | 'migration' | ...
  source_agent   text
  source_item_id uuid
  notes          text             -- why this version changed
  is_current     boolean not null default true
  pinned         boolean not null default false
  created_by     text
  created_at     timestamptz not null default now()
  superseded_at  timestamptz
  updated_at     timestamptz not null default now()
  UNIQUE (subject_type, subject_key) WHERE is_current
```
Write = supersede pattern in one tx (flip current → insert new). Rollback =
restore a prior row.

```
doc_notes
  id             uuid pk default gen_random_uuid()
  subject_type   text not null check (subject_type in ('tool','pipeline'))
  subject_key    text not null
  site_id        uuid             -- set for a per-site incident; null = library/global
  body           text not null    -- Observed / Root cause / Fix / Verified (or diagnosis report)
  categories     jsonb not null default '[]'::jsonb
  source         text
  source_agent   text
  source_item_id uuid
  created_by     text
  created_at     timestamptz not null default now()
  INDEX btree (subject_type, subject_key, created_at DESC)
  INDEX gin (categories jsonb_ops)   -- jsonb_ops so `categories ? 'tag'` is indexable
```
Append = one INSERT (no read-modify-write; concurrent-safe).

---

## Write hooks

- **PLAN — at creation / intent change.** `write_doc_plan` (supersede tx).
  Tools: from `create_tool_component`, drafted from the generator's reasoning
  (spec slice, delivery mechanism, deliberate decisions, acceptance criteria).
  Pipelines: authored initially (largely distilled from 004–008), superseded on
  direction change. Not written on `deploy_tool_to_site` forks.
- **NOTES — at modification.** `append_doc_note` (one INSERT) as the **last step**
  of `tool-improver` / `update_component_html` / `component-template-fixer`
  (symptom → root cause → fix → verification + categories; `site_id` if per-site).
- **NOTES — from the diagnosis loop (rev 3, Phase A).** A config-gated step
  `persist_diagnosis_note` placed **after** `diagnose_emit` in the diagnosis
  agent's workflow (emit itself stays read-only per its design comment). Maps the
  report (summary, conclusion, evidence trail) into a `doc_notes` row when the
  run's subject is explicit in `input_data` (a `function` or a `pipeline`);
  **skips rather than guesses** when it isn't. `status=CONFIRMED` → a root-cause
  entry; `UNVERIFIABLE` → still written, tagged `unconfirmed-diagnosis` — dead
  ends are part of NOTES' purpose. Category `diagnosis` + taxonomy tags.
- **NOTES — from workflow-altering migrations (pipelines).** The migration (or the
  agent applying it) appends a note: migration number, what changed, why —
  formalising the 005 migration-table practice.
- **Index step (reuse).** After any write, `rag_index` with the matching
  collection. (Optional: parameterise its hardcoded `source_type='scrape'`.)
- **Mirror step (optional, Phase B).** Render current docs → docs repo commit.

Keep workflows flat: each write is one Go action; no sub-workflows.

---

## Pipeline documentation (rev 3)

**Derive the topology; author the intent.** The step map / spawn graph / routing
is generated on demand from `agent_definitions` (callgraph pattern) — never
hand-drawn, so it can't drift. The authored pipeline PLAN holds only:

- **End-to-end invariants** — testable guarantees the pipeline must preserve, e.g.
  "an interactive section survives every rebuild route" (the unstated invariant
  whose violation was the de-tool bug); "nav stays in sync with pages".
- **Branch rationale** — why each route exists and when to use which, e.g.
  `page-build-handler` re-resolves section sources / `page-rebuild` deliberately
  does not; dispatch-loop `page_name` mapping rule.
- **Seams** — which pipelines share agents/tables (seam bugs, like de-tool, happen
  where two pipelines reuse one handler). 
- **Deliberate decisions** — priorities (cross-links at 110), cooldowns,
  NULL-as-stale choices.

Pipeline NOTES = incidents at pipeline scope (largely what 016/016b holds today) +
migration entries + persisted diagnoses. 016/016b remains the global rolled-up
guide; categories still graduate recurring lore into pipeline guards, after which
the entry is history.

**Retrieval symmetry:** a pipeline's scope IS code, so the `docselect` catalogue
with `path_globs` (agent/action file paths) fits pipelines the way it didn't fit
tool functions; `rag_lookup` for discovery; direct-by-key for agents that know
their pipeline. Relationship to 004–008: they remain the prose base and the first
`doc_plans` bodies can be distilled from them; the net-new authored content is the
invariants + branch rationale.

---

## Tool assurance: contracts, acceptance, iteration (rev 3)

"Fully tested" needs a per-tool definition of *working*. That is documentation's
job; the checking is machinery fed by it.

- **Behaviour contract = acceptance criteria.** The PLAN's contract section is
  written as checkable statements (root element boots; no `<no value>`; JS asset
  present; result renders for input X; data feed fresher than N days). For
  **multi-page tools** it must also state the **page set + per-page roles** and the
  **inter-page contract** (URLs, shared state keys, data feeds, snippet version
  expectations) — checks then include cross-page links and shared-state presence.
- **Verification ladder (reuse upward):**
  1. Structural — exists (`check_tool_health` Tier 1).
  2. Contract-presence — cheap DOM/selector/asset assertions from the criteria
     against deployed HTML (catches `empty-shell`, `detool-on-rebuild`,
     `js-not-extracted`). New thin check, no browser.
  3. Acceptance audit — extend `tool-auditor` from "code vs header invariants" to
     "**deployed pages vs PLAN criteria**"; failures spawn improvement items
     carrying the failing criterion (the findings `acceptance_test` pattern).
  4. Behavioural (headless browser actually driving the tool) — genuinely new
     infrastructure; a separate decision, not designed here. (023 checked: it is
     model/prompt evaluation only — no behavioural layer exists today.)
- **Iteration loop.** create → deploy → acceptance audit → failing criteria →
  improvement item → fixer **loads PLAN+NOTES first** → fix → `append_doc_note` →
  re-audit. A tool is *working* when its criteria pass; the criteria hold the bar
  still across iterations, the notes stop iterations fighting each other, and
  unconfirmed attempts are recorded as dead ends.
- **Prerequisite before multi-page tools scale:** the pending preserve-sections
  re-render + interactivity-aware save guard (016b/005/020/026) — every extra tool
  page is another page a content rebuild can silently de-tool.

---

## Retrieval (unchanged shape, subjects generalised)

- **Direct-by-key (primary, fix-time):** thin read action (maintenance mould):
  current `doc_plans` body + latest-N `doc_notes` for the subject.
- **Code-diagnosis loop:** a sibling read action resolves the in-scope subject
  (tool `function` from the runtime target, or the pipeline) and returns PLAN+NOTES
  text; one compose line in `diagnose_assemble_bundle`. Pipeline PLANs additionally
  fit the `docselect` catalogue via `path_globs` once the mirror supplies files.
- **Semantic discovery:** `rag_lookup` on the collections.

---

## Deliberate-decisions — prose now, no locks
Unchanged: a PLAN section ("Deliberate decisions — do not re-fix") + NOTES
narrative; protective because loaded at fix time. Enforcement (locks/`must_have`)
is the compilation target if a regression class shows prose isn't honoured.

---

## Document formats

**Tool PLAN sections:** Aim · Source spec · Behaviour contract → **Acceptance
criteria** (+ for multi-page: Page set & inter-page contract) · Delivery mechanism
+ why · Dependencies · Deliberate decisions.

**Pipeline PLAN sections:** Aim · Invariants (end-to-end) · Branch rationale ·
Seams · Deliberate decisions · (topology = derived, referenced not embedded).

**NOTES entry (uniform, dated, one row each):**
```
## 2026-07-04 — <short title>
Observed: <symptom, where>
Root cause: <cause — or "unconfirmed" + stopped_by for a diagnosis trail>
Fix: <what changed> (<site_id> if per-site)
Verified: <how confirmed>
Categories: detool-on-rebuild, diagnosis, ...
```

**Category taxonomy:** rev-2 set (`css-variable-mismatch`, `empty-shell`/
`mode-b-template`, `broken-template-slots`, `content-vs-runtime-mismatch`,
`detool-on-rebuild`, `js-not-extracted`, `js-bundle-stale`,
`schema-template-drift`) + `diagnosis`, `unconfirmed-diagnosis`, `migration`,
`seam`, `acceptance-fail`.

---

## Phasing

**Phase A (now)**
1. Lock formats (above).
2. Migration: `doc_plans` + `doc_notes` (confirm live DB has no collision; verify
   `site_work_items.pipeline` enum values for subject keys).
3. Actions: `write_doc_plan` (supersede), `append_doc_note` (INSERT), the
   direct-by-key read loader, and **`persist_diagnosis_note`** (config-gated step
   after `diagnose_emit`).
4. Wire: `create_tool_component` → PLAN (incl. acceptance criteria); fix agents →
   note-append as last step; `rag_index` step after writes.
5. Contract-presence check (ladder step 2) fed by tool criteria.
6. Verify/settle the KB `tool_docs` write claimed by 019.

**Phase B (when useful)**
7. Acceptance audit (`tool-auditor` extension: deployed pages vs criteria).
8. Pipeline PLAN bodies distilled from 004–008; migration write-hook live.
9. Optional git mirror; `docselect`/assembler injection for pipeline docs.
10. Category roll-up query surfaced into 016/016b.
11. Headless behavioural testing — separate infrastructure decision.

---

## Open questions / dependencies
- `site_work_items.pipeline` — confirm the value set on the live DB.
- KB `tool_docs` write — confirm whether it exists (uploaded action lacks it).
- `deploy_tool_to_site` — confirm forks stamp `source_*`; NOTES-only on fork.
- `rag_index` `source_type` — parameterise or accept `'scrape'`.
- Acceptance-audit consumer — how `tool-auditor` ingests criteria (prose via LLM
  first; lift to structured only on volume, per the graduation rule).
- Headless testing — future decision; not in scope.

## Reuse ledger
site_specs supersede pattern · rag_index/rag_lookup/knowledge_base ·
source_*/source_item_id convention · maintenance-mould read actions ·
diagnose_emit output (producer) · findings acceptance_test pattern ·
callgraph/agent_definitions derivation · docselect (pipelines, later) ·
content_components.function + site_work_items.pipeline (keys) · inline tool-doc
header (untouched) · lifecycle agents as write hooks · check_tool_health (ladder
step 1; can flag missing docs later) · git adapter (optional mirror).
