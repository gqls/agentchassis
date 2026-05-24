# FOCUS — Adopting the Best-Practice Doc Structure (Careful First Path)

**Status:** actionable plan, for following or amending. The careful, incremental way to get value from the best-practice doc structure against the **current** setup, without committing to the atomic rewrite, the mediator, or the routing build.

Companions: `FOCUS_self_development_coding_pipeline_reasoning.md` (coordination, the mediator), `FOCUS_best_practice_doc_tree.md` (the optimal atomic structure — the eventual target, not this plan), `FOCUS_mediator_routing_model.md` (how a mediator would route — deferred).

---

## 1. Grounding fact: the corpus does not fit in context

Measured: ~200 files, ~6.7MB, roughly **1.0–1.7M tokens** (prose-based estimate at the low end; markup and SQL push it up). At or above a 1M window, and growing with every handoff. So:

- "Send everything to project files / stuff it all into context" is not a reliable option for the whole corpus, and the gap widens over time.
- The question is therefore **what to load completely vs what to search.**

---

## 2. The core separation

Two independent decisions. Adopt the structure first; retrieval is separable and comes after.

- **Structure** — the constitution + concern/`applies_to` tagging. Pays off regardless of retrieval mechanism.
- **Retrieval** — how content reaches a prompt. Decided after the structure exists, and informed by §5.

### Two retrieval needs with opposite characteristics

| Need | Size | Requirement | Right tool |
|---|---|---|---|
| The **rules** for a task (constitution + matched concern standards) | Tiny — fits in context easily | **Completeness** — must not miss a governing rule | Tag-based deterministic selection |
| The **broad corpus** (handoffs, focus/status docs, how-it-works narration) | 1M+ | Recall; completeness impossible anyway | Embeddings (semantic search) |

Loading rules by similarity risks dropping the one rule that mattered (a missed contract rule ships a broken adapter). Searching the broad corpus by tag is impossible (most of it isn't tagged rules). So the two needs take different mechanisms — not either/or.

---

## 3. Benefits vs the current setup

- **Query by concern instead of reading across subsystem docs.** Today the rules for "writing a Go action" are spread across 001, 002, 003, the naming FOCUS, and others, because docs are organized by subsystem. Tagging makes that a query that returns just the governing standards.
- **A small always-on baseline.** Today there's no tight baseline — you either load the big core docs (token cost, attention dilution) or rely on remembering the standing rules. The constitution is a tiny composable block that replaces both.
- **Completeness for rules.** Tag selection returns every matching standard, exactly. Current retrieval (project search or RAG) returns whole docs or top-k similar chunks and can miss one.
- **Lower tokens, less noise per prompt** from pre-filtering to the relevant slice.
- **Surfaces duplication and drift.** Tagging exposes the same rule living in multiple docs; the eventual atomic form removes it.
- **Separates normative from descriptive** — pull "the rules" cleanly out of "the narration," which isn't possible now.

---

## 4. Phased adoption plan

None of this requires the mediator or the atomic rewrite up front. Each phase stands alone and delivers value.

### Phase 1 — Write the constitution (zero infrastructure)
Distill the standing rules into one short doc (1–2 pages) that points to fuller docs for detail. Candidate contents, already visible across 001/003 and stated working preferences: reuse/search before creating (STEP ZERO); check schema before SQL; workflows simple, complexity in Go; keep workflow variable names in sync with what actions expect; every agent is an orchestrator; agents reply to the caller's responses topic; no `logger.Debug`; don't rename variables silently; prefer fixing structural issues over quick fixes.

- **Payoff:** every dev session and agent prompt starts from a tight baseline instead of three large docs or nothing.
- **Effort / risk:** low / low. One new doc, no existing docs touched.

### Phase 2 — Tag existing docs by concern + `applies_to` (overlay, no rewrite)
Add frontmatter to existing docs (or their sections): `concern`, `applies_to`, and a `kind: rule | reference` flag. No content rewrite. A small script reads frontmatter into a manifest (JSON or a `standards`/`doc_index` table).

- **Payoff:** "what governs a `go_action` change" becomes answerable; load only those docs/sections.
- **Effort / risk:** low–medium / low. Mechanical tagging pass; the judgement is choosing the concern set and the `applies_to` vocabulary (reuse the lists in the doc-tree and routing FOCUS docs).

### Phase 3 — Make the retrieval split real
Keep loading the rules by tag (Phase 2 manifest). Point the **existing** nomic/pgvector/ollama RAG (the `rag_actions` with nomic prefixes already in the stack) at the full corpus for semantic exploration. Not building retrieval — aiming the stack already running at the dev docs.

- **Payoff:** comprehensive search over the 1M+ corpus, plus exact rule loading. Serves both a custom dev assistant and the pipeline agents.
- **Effort / risk:** medium / low–medium. Chunk + embed the corpus; the infra exists, the work is wiring and a refresh job.

### Phase 4 — Atomic extraction, evidence-driven (deferred)
Only where usage proves it: extract hot sections into atomic standards with generated views (the full structure in the doc-tree FOCUS). Driven by observed need — a section repeatedly loaded in isolation, a rule repeatedly violated — not done up front.

- **Payoff:** single source per rule, generated handbooks + manifest from one source.
- **Effort / risk:** higher / medium. Defer until Phases 1–3 are in use.

---

## 5. Retrieval decision, stated plainly

- **Rules:** tag-based, deterministic, complete. Small enough to always load.
- **Broad corpus:** embeddings via the owned nomic/pgvector/ollama stack — the "more comprehensive search."
- **Claude.ai project files:** fine for these dev conversations now, but they don't help the pipeline agents and the retrieval tuning isn't yours. If the assistant outgrows chat, lean on the stack you own.
- **Not either/or:** tags for rules, embeddings for exploration.

---

## 6. Explicitly deferred (so the first path stays small)

- The atomic-standard rewrite (Phase 4 above; full design in the doc-tree FOCUS).
- The mediator and the coordination decision (positions A/B/C).
- The routing build (classifier, trigger policy) — that's a later layer that consumes the Phase 2 tags.

---

## 7. One-line state and next action

Adopt structure before retrieval; structure is the constitution (Phase 1) + concern/`applies_to` tagging (Phase 2), retrieval is tags-for-rules + the existing embedding stack for exploration (Phase 3), atomic rewrite deferred (Phase 4). **Next concrete action:** draft the constitution from the standing rules across 001/003 and the stated preferences, as an artifact to react to.
