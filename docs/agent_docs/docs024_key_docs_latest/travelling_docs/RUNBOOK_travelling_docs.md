# RUNBOOK — Tool & Complex-Component Travelling Docs (PLAN + NOTES)

**Created:** 2026-07-04
**Last updated:** 2026-07-04
**Applies to:** tools (`content_components.component_level = 'tool'`) and, later,
interactive/stateful non-tool components.

---

## What we are aiming to achieve

Operational how-to for the travelling-docs feature: where a tool's `PLAN` and
`NOTES` live, when and how they are written, how they are made retrievable, and
how a fixer loads them at the moment of a fix. This runbook assumes the design in
`PLAN_travelling_docs.md`. It extends — does not replace — the tool-doc header
system (019 §Tool Doc Header). If a step here appears to conflict with the header
rules, the header rules win for the header block; this runbook governs the
separate PLAN/NOTES files.

> Prereqs not yet in place are flagged **[DEPENDENCY]**. Do not skip them — the
> feature fails the same way the header gate fails without its prompt migration.

---

## 1. Where the docs live

- **Source of truth (git):** `<docs-repo>/tools/<function>/PLAN.md` and
  `<docs-repo>/tools/<function>/NOTES.md`. A **new** repo the framework can write
  to — **not** `agentchassis` (read-only to the running framework) and **not**
  under any single domain directory in the sites repo (these docs are
  library-level, shared across sites). **[DEPENDENCY: docs repo + adapter write
  target provisioned.]**
- **Retrieval index (DB, derived):** a `knowledge_base` collection (e.g.
  `tool_docs`) populated by `rag_index`. This is a copy for lookup, never edited
  directly.
- **NOTES-in-DB (only if adopted in Phase B):** one row per entry, keyed by
  `function` (+ `site_id` for per-site incidents). PLAN stays in git regardless.

Key rule: **`function` is the identity.** It must match
`content_components.function` byte-for-byte (kebab-case, `tool-` prefix for
tools). Docs are keyed by `function` so they survive forks and regenerations.

---

## 2. Writing a PLAN (at creation)

**When:** once, when a new `function` is first created — inside the
`tool-generator` → `create_tool_component` path. **Not** on `deploy_tool_to_site`
forks (a fork reuses the library `function`; no new PLAN).

**Who:** the creation action drafts it from material already in hand — the spec
slice it used, the delivery mechanism it chose, the deliberate decisions. This is
the reasoning currently discarded after `description` is written.

**How:**
1. Draft `PLAN_<function>.md` with the sections in `PLAN_travelling_docs.md`
   (Aim, Source spec, Behaviour contract, Delivery mechanism + why, Dependencies,
   Deliberate decisions).
2. Commit to `<docs-repo>/tools/<function>/PLAN.md` via the docs-repo adapter.
3. `rag_index` the file into the `knowledge_base` doc collection.
4. Register a catalogue entry (§4).

**Do not** put orchestration ids, agent names, or dates in the inline `<script>`
header — those live in `source_*` columns and in the PLAN/NOTES, never in code
that could ship if stripping regressed (019 rule).

---

## 3. Appending a NOTES entry (at every modification)

**When:** any time a tool/component is changed or a bug is fixed —
`tool-improver`, `update_component_html`, `component-template-fixer`, or a manual
fix. The entry is the fixing agent's **last step**.

**Who:** the agent (or human) doing the fix.

**How:** append one uniform, dated entry (keep it import-shaped so a later move to
a DB table is an import, not a reformat):

```
## <YYYY-MM-DD> — <short title>
Observed: <symptom, and where — page/site if relevant>
Root cause: <the actual cause, not the surface symptom>
Fix: <what changed> (<site_id> for a per-site incident)
Verified: <how you confirmed it — query, rendered check, sweep clean>
Categories: <one or more taxonomy tags>
Dead end: <optional — what was tried and abandoned, so it isn't retried>
```

Then re-`rag_index` the NOTES digest so retrieval stays current.

**Category taxonomy** (reuse; extend as needed): `css-variable-mismatch`,
`empty-shell`/`mode-b-template`, `broken-template-slots`,
`content-vs-runtime-mismatch`, `detool-on-rebuild`, `js-not-extracted`,
`js-bundle-stale`, `schema-template-drift`.

---

## 4. Making a doc retrievable (catalogue entry)

Add one `DocRule` to `diagnose_doc_catalogue*.json` per tool. `SelectDocs` matches
on `always`, OR a keyword being a substring of the hypothesis, OR a path-glob
being a substring of an in-scope symbol path (all case-insensitive).

```json
{
  "doc": "tools/<function>/PLAN.md",
  "keywords": ["<function>", "<one or two symptom terms>"],
  "path_globs": ["<function>", "create_tool_component"]
}
```

Notes:
- The `doc` value is a path forwarded verbatim to the bundle via `-doc`.
  **[DEPENDENCY: confirm how `-doc` resolves a path in the docs repo — the
  diagnosis loop may need its doc-root repointed. This is acceptable per the
  agreed design.]**
- A false match only adds a doc (not misinformation); a miss degrades to "no
  extra doc" (today's behaviour). Benign both ways.

---

## 5. Loading context at fix time (the payoff)

Order of preference for a fixer that needs to get up to speed:
1. **Direct-by-function** — the agent already knows the `function` it is fixing;
   load `PLAN_<function>.md` + `NOTES_<function>.md` (git file, or DB rows in
   Phase B) directly. This is the primary path.
2. **Deterministic catalogue** — for the diagnosis loop starting from a
   hypothesis/scope that has not yet localised to a `function`, `SelectDocs`
   pulls the PLAN in alongside the 003/016 docs.
3. **Semantic** — `rag_lookup` over the `knowledge_base` copy for free-text
   recall.

Read the PLAN's **Deliberate decisions** before changing anything — that section
is the guard against re-fixing an intentional choice (no locks in place yet).

---

## 6. Verification (after any change to the feature or a new tool)

Mirror the header/thin-slice verification discipline:

1. **PLAN present in git:** `<docs-repo>/tools/<function>/PLAN.md` exists and its
   `function:`/Aim matches `content_components.function`.
2. **Catalogue matches:** a hypothesis naming the `function` (or a scope path
   containing it) returns the PLAN from `SelectDocs`.
3. **RAG returns it:** `rag_lookup` for the tool's purpose surfaces the PLAN
   collection row.
4. **NOTES append works:** after a fix, the new dated entry appears with a
   `Categories:` line and re-indexes.
5. **Inline header untouched:** the `<script>` still opens with the sentinel
   block; `StripToolDocHeader` still removes it from shipped page/asset (grep the
   committed page and `/tools/assets/<function>.js` — `tool-doc` must find
   nothing); `tool_health` raises no `no_doc_header`/`malformed_doc_header`.

Remember: **do not treat a 0-row result as decisive** until the query itself is
ruled out (right `function`, right collection, right catalogue path).

---

## 7. Gotchas

- **`agentchassis` is read-only to the framework.** The framework cannot
  auto-commit docs there. That is why docs go in a *new* writable repo (or DB),
  and why the header/provenance rollout was human-applied migrations + a binary
  release.
- **Library vs site.** PLAN describes the tool *type* (one per `function`). A
  per-site incident is a NOTES entry that names the `site_id` — not a new PLAN,
  and never filed under a domain directory in the sites repo.
- **Don't duplicate the inline header's job.** The header is the short
  audit/parse anchor (6–12 lines). The PLAN is the substantive intent. Keep them
  consistent but don't restate the whole PLAN in the header.
- **NOTES git-append friction.** If concurrent appends to a shared `NOTES.md`
  cause merge pain once agents write them, that is the Phase B trigger to move
  NOTES to a DB table (PLAN stays git).
- **Kubernetes / build reminders:** actions run in the chassis binary; changes
  are drafted for review and shipped via GitHub Actions → Backblaze, not built
  in-container. Pods: `kubectl -n ai-persona-system get pods`; Kafka:
  `kubectl -n kafka get pods`. Use `logger.Info` (not `logger.Debug`).
