# Constitution (thin slice)

The always-on rules for any task on this codebase. Included in full in every bundle. These are the things true regardless of task; task-specific standards (the detailed contracts in 003) are listed at the end and included only when a task touches them.

This is the flat-file version for the thin slice. Later it becomes the `standards` rows with `scope = constitution`; the content is the same.

---

## Reuse and structure

- **Reuse before recreate.** Before writing a new function, struct, or component, look for an existing one that does the same or similar and improve/alter it. Recreating something the system already has is a defect, not a shortcut.
- **Fix structural problems, not symptoms.** Prefer the fix that addresses the cause over the quick patch, even when the patch is faster. Note when you are knowingly deferring a structural fix.

## Agents and workflows

- **Every agent is an orchestrator.**
- **Distinct responsibilities, minimal overlap.** Each agent owns its area; agents overlap as little as possible.
- **Reply to the caller's responses topic.** An agent always responds on its parent's (caller's) responses topic, never on its own responses topic.
- **Workflows stay simple; complexity lives in Go action code.** Keep the workflow declarations thin and put the real logic in the actions.
- **No subworkflows in SQL — spawn sub-agents instead.** When work needs to branch, spawn a sub-agent with its own workflow rather than nesting subworkflows in SQL. This keeps logs clear, maintenance easier, and responsibilities separate.
- **Keep workflow variable names in sync with what the actions expect.** A workflow variable name must match the name the action reads. Do not let them drift.

## Code and data conventions

- **Check the database schema before writing SQL.** Always inspect the real schema first; never write SQL against an assumed shape.
- **Parameterised queries only.** Pass values as query parameters. Never build SQL by interpolating values into a template string.
- **Don't change variable names silently.** Keep names stable; if a rename is deliberate, say so explicitly and note it.
- **String-value naming — the two-and-a-half conventions (003):**
  - `snake_case` for **identifier-shaped** values used as keys in code: `switch case` constants, `map` keys, action-registry names, work-item `item_type`, dispatch routing, Kafka topic segments, k8s labels. (e.g. `needs_blog_posts`, `create_blog_posts`.)
  - `kebab-case` for **data-shaped** values that describe what a thing is and never act as code identifiers — they end up in CSS, URLs, HTML, prompts. (e.g. `social-proof`, `blog-post`.)
  - lowercase single word where no separator is needed (e.g. `planned`, `triaged`).
  - The test: does any Go `case`/`map` key/route/label use the value? Yes → snake. No → kebab.
- **Storage conventions (chassis):** enum-like columns are `text` + `CHECK`, not native enums; versioned entities use `version` + `previous_version_id`; soft-delete via `deleted_at`, not a `status = archived`.

## Logging

- **Don't use `logger.Debug`** — it does not show in the logs. Use a level that is actually emitted.
- **Put the run id in log lines.** Log the `orchestration_id` (and `correlation_id`) so a run can be traced across agents and tables. (Coverage of this is not yet verified everywhere — treat its absence in a line as unknown, not as "didn't happen.")
- **Log agent creation and the messages between agents** (headers and body), so the spawn tree and message flow are reconstructable.

## Deployment and infrastructure

- **Deployment path:** write to GitHub (or a future adapter) → GitHub Actions triggers a write to Backblaze S3.
- **Kubernetes namespaces:** main is `ai-persona-system` (e.g. `kubectl -n ai-persona-system get pods`); Kafka is in `kafka` (e.g. `kubectl -n kafka get pods`).
- **Kafka cluster:** `personae-kafka-cluster-*` (combined-pool and entity-operator pods). Use the real cluster names.

## Tone of generated text

- Plain, pragmatic, concrete. Avoid hype words and filler. This governs generated content and commit messages, not just chat.

---

## Task-specific standards (the 003 contracts) — included only when a task touches them

These are not always-on; the bundle pulls the relevant one in when the task's area matches. Listed here so the index is visible:

- Component Naming Contract (kebab `function`; one function, one active component; `data-component` flow).
- JS Content Separation Contract (HTML/JS split; asset path convention; `js_snippets`).
- Component Creation & Regeneration Contract (return statuses; version-history preservation).
- Site Component Linkage Contract (slot ↔ function mapping; `unlinked_site_components` check).
- CSS Colour Inheritance Model and Section Context (dark sections) Contract.
- CSS Theme Template Contract (responsibility split; theme storage/lineage columns; review gate; forking rules).
- Query Database Parameterisation Contract (the parameterised-query rule above, with examples).

For the thin slice these stay in 003; when a task touches one, paste that section alongside this constitution.
