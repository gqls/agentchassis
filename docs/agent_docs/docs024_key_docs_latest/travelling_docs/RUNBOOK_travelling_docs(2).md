# RUNBOOK — Travelling Docs (PLAN + NOTES): Tools, Complex Components, Pipelines

**Created:** 2026-07-04
**Last updated:** 2026-07-04 (rev 3 — subject-generic tables; diagnosis persist step; pipeline procedures; tool acceptance & iteration)
**Applies to:** tools (`content_components.component_level='tool'`),
interactive/stateful components, and pipelines (keyed by
`site_work_items.pipeline`).

---

## What we are aiming to achieve

Operational how-to for the travelling-docs feature: where a subject's `PLAN` and
`NOTES` live, how they are written/versioned, how they are indexed, and how a
fixer loads them at fix time. Assumes `PLAN_travelling_docs.md` (rev 3). Extends —
does not replace — the tool-doc header (019). Source of truth is Postgres; git is
an optional non-authoritative mirror.

> Prereqs not in place are flagged **[TODO]**.

---

## 1. Where the docs live

- **Truth (Postgres):** `doc_plans` (current + history) and `doc_notes`
  (append-only), keyed by `(subject_type, subject_key)`:
  `('tool', <function>)` — byte-for-byte `content_components.function`;
  `('pipeline', <pipeline>)` — a `site_work_items.pipeline` value.
  **[TODO: migration; confirm no name collision and the pipeline enum on the live
  DB first — don't take the schema dump as decisive.]**
- **Retrieval index (derived):** `knowledge_base` collections (`tool_docs`,
  `pipeline_docs`) via the existing `rag_index`; queried by `rag_lookup`
  (discovery only — no key filter). Never edited directly.
- **Mirror (optional):** rendered markdown committed to a docs repo. If a commit
  fails, truth is untouched.

---

## 2. Writing / updating a PLAN

**Tools — when:** at first creation of a `function` (inside `tool-generator` →
`create_tool_component`), and on intent change. **Not** on `deploy_tool_to_site`
forks. Draft from the generator's reasoning: spec slice, delivery mechanism +
why, dependencies, deliberate decisions, and **acceptance criteria** (see §5).
Multi-page tools must state the page set, per-page roles, and the inter-page
contract (URLs, shared state keys, data feeds).

**Pipelines — when:** authored initially (distil from 004–008), superseded on
direction change. Sections: Aim · Invariants · Branch rationale · Seams ·
Deliberate decisions. **Do not embed the step map** — topology is derived from
`agent_definitions` (callgraph pattern) on demand.

**How (`write_doc_plan`, one tx):**
```
UPDATE doc_plans SET is_current=false, superseded_at=now()
 WHERE subject_type=$1 AND subject_key=$2 AND is_current;
INSERT INTO doc_plans (subject_type, subject_key, body, source, source_agent,
                       source_item_id, notes, created_by) VALUES (...);
```
Then a `rag_index` step (matching collection). Rollback = restore a prior row.
`pinned=true` marks a human hold. Keep orchestration ids/agent names/dates out of
the inline `<script>` header (019 rule) — provenance lives in `source_*` columns.

---

## 3. Appending a NOTES entry

**When:** any change or fix (`tool-improver`, `update_component_html`,
`component-template-fixer`, manual) — the fixing agent's **last step**; any
workflow-altering **migration** (pipeline note: migration number, what, why); and
**every diagnosis run** whose subject is explicit (§4).

**How (`append_doc_note`, one INSERT — no read-modify-write):**
```
INSERT INTO doc_notes (subject_type, subject_key, site_id, body, categories,
                       source, source_agent, source_item_id, created_by)
VALUES ($1,$2,$3,$4,$5::jsonb,$6,$7,$8,$9);
```
`body` = the uniform entry (Observed / Root cause / Fix / Verified); `site_id`
for a per-site incident. Re-`rag_index` the digest (can batch).

**Categories:** rev-2 taxonomy + `diagnosis`, `unconfirmed-diagnosis`,
`migration`, `seam`, `acceptance-fail`.

---

## 4. Persisting diagnosis output (rev 3)

A config-gated step **`persist_diagnosis_note`** runs **after** `diagnose_emit`
in the diagnosis agent's workflow (emit stays read-only). Mapping:
- subject: an explicit `function` or `pipeline` in `input_data` — **skip rather
  than guess** if absent;
- `Observed:` the seed symptom/hypothesis; `Root cause:` the conclusion when
  `status=CONFIRMED`, else `unconfirmed (<stopped_by>)`; `Verified:` a one-line
  evidence-trail summary; categories `diagnosis` (+ `unconfirmed-diagnosis` when
  not confirmed — dead ends are wanted, they stop retries).
- `source='diagnosis-loop'`, `source_item_id` = the triggering work item if any.

---

## 5. Tool acceptance & iteration (rev 3)

1. **Criteria live in the tool's PLAN** (written at creation): checkable
   statements — root element boots, no `<no value>`, `/tools/assets/<function>.js`
   present, cross-page links resolve, shared-state keys present, feed fresh.
2. **Ladder:** Tier 1 structural (`check_tool_health`, exists) → Tier 2
   contract-presence (thin check asserting criteria selectors/assets against
   deployed HTML — catches `empty-shell`, `detool-on-rebuild`, `js-not-extracted`)
   **[TODO]** → Tier 3 acceptance audit (`tool-auditor` extended to judge deployed
   pages vs PLAN criteria; failures spawn improvement items carrying the failing
   criterion — findings `acceptance_test` pattern) **[TODO Phase B]** → Tier 4
   headless behavioural (separate future decision).
3. **Iteration:** deploy → audit → failing criterion → improvement item → fixer
   **loads PLAN+NOTES first** → fix → `append_doc_note` → re-audit. A tool is
   *working* when its criteria pass. Do not mark a complex tool done on deploy.
4. **Multi-page prerequisite:** the preserve-sections re-render +
   interactivity-aware save guard (pending; 016b/005/020/026) before scaling
   page counts — each extra page is another de-tool surface.

---

## 6. Loading context at fix time

1. **Direct-by-key (primary):** thin read action (maintenance mould):
```
SELECT body FROM doc_plans  WHERE subject_type=$1 AND subject_key=$2 AND is_current;
SELECT body, categories, created_at FROM doc_notes
 WHERE subject_type=$1 AND subject_key=$2 ORDER BY created_at DESC LIMIT $3;
```
   Read the PLAN's **Deliberate decisions** before changing anything.
2. **Code-diagnosis loop:** sibling read action resolves the in-scope subject and
   returns PLAN+NOTES text; one compose line in `diagnose_assemble_bundle`.
   Pipeline PLANs can additionally enter the `docselect` catalogue via
   `path_globs` once the git mirror supplies files (Phase B).
3. **Semantic:** `rag_lookup` on the collections (discovery only).

---

## 7. Verification

1. One current PLAN per subject: the partial unique index enforces it; after a
   supersede, exactly one `is_current=true` remains.
2. NOTES append visible with non-empty `categories`; diagnosis runs on an
   explicit subject leave a `diagnosis` entry (unconfirmed ones tagged).
3. Roll-up uses the GIN index:
   `SELECT subject_key, created_at FROM doc_notes WHERE categories ? 'detool-on-rebuild';`
4. `rag_lookup` surfaces the collections.
5. Tool criteria pass on the ladder tiers in place; failures created improvement
   items with the criterion in spec.
6. Inline header untouched: sentinel present in `rendered_html`; stripped from the
   shipped page and `/tools/assets/<function>.js`; `check_tool_health` raises no
   `no_doc_header`/`malformed_doc_header`.
7. Pipeline topology claims: regenerate from `agent_definitions` — never trust an
   embedded step map.

Don't treat a 0-row result as decisive before ruling out the query (right subject
key, right `is_current`, right collection).

---

## 8. Gotchas

- **Truth is Postgres, not git**: the commit path rejects empty `Domain`,
  force-prefixes `{domain}/`, no file-read, whole-file commits, no conflict
  retry, single serialised adapter — never the home of an append.
- **NOTES = one row per entry**, never a shared file (RMW/lost-update risk).
- **Library vs site:** one current PLAN per subject; per-site incidents are notes
  with `site_id`.
- **Skip, don't guess** the diagnosis subject; a mis-filed note poisons history.
- **Don't hand-draw pipeline topology** — derive it; author only invariants,
  branch rationale, seams, decisions.
- **Don't duplicate the inline header** (short anchor) in the PLAN (substantive
  intent); keep consistent.
- **`rag_index` labels `source_type='scrape'`** — parameterise if the label
  matters.
- Workflows stay flat; complexity in Go; `logger.Info`; check schemas before SQL;
  pods `-n ai-persona-system`, Kafka `-n kafka`; deploys via GitHub → Actions →
  Backblaze, not built in-container.
