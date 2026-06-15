# PLAN — Bundle Shape Contract

**Status:** contract specification. The bundle is what the bundle builder assembles for one task and hands to the generation step. It is the thing that replaces today's "paste some code and a doc into a chat." The whole-plan review flagged its shape as undefined while every consumer needs it fixed — the same gap the active-config schema had, one step further along. This settles it.

---

## 1. What the bundle is

A bundle is a task-scoped package of context. The builder assembles it fresh for a single task — a feature, a bug, a maintenance change — from the current active config and the current code and state. It is built new each time, not stored and reused, so it is never stale.

It has two jobs: give the generation step everything it needs to do the task correctly, and record what it was given so the decision can be audited later.

---

## 2. The parts of a bundle

A bundle has these sections. Each section says plainly what goes in it.

- **Metadata.** Bundle id, tenant id, the task it was built for, the step type (see §4), the time it was assembled, and a pointer to the config state it was built from (which atom versions — this is the provenance, §6).
- **Task and target.** What is being done, and where: the area (which objective-tree node) and the specific files or functions involved.
- **Authored layer — the "why" and "how well".** Pulled from the active config:
  - The constitution (the always-on rules).
  - The why-chain: the path from the mission down to this task's area, with each step's one-line purpose.
  - The priority profile for this area (which dimensions win when they conflict).
  - The direction-of-travel for this area (current heading, anything settled, anything deliberately temporary).
  - The standards that apply to this task (matched by the task's change types).
- **Code context — the "what is".** Pulled from the code analysis:
  - The in-scope code: the files or functions being changed, in full.
  - The neighbourhood: the signatures (not the bodies) of the things that call, or are called by, the target — so the generator sees the shape around it without drowning in detail.
  - Reuse-search results: existing functions or structs — **and existing definition rows (§2.1)** — that already do something similar, put in front of the generator so reuse-before-recreate is the default, not a hope.
  - Relevant schema, for tasks that touch the database.
  - **Definition data (§2.1):** in a system where behaviour is stored as data, the thing being worked on is often a database row, not a file. The in-scope agent, workflow, tool, or prompt definitions belong here, treated like source.
- **Database data — three kinds (§2.1, §3).** Not one vague "relevant rows" line. The bundle distinguishes definition data (the system's design, fetched routinely), operational data (what's happening now, for debug/ops), and content data (the output, gated in the service case). How they are fetched safely is the multipass flow in §3.
- **Pointers.** Links to everything that was *not* put in the bundle in full — other standards, other code, fuller docs, larger query results. The generator (or a person) can follow these for more if needed, but they don't bloat the bundle.
- **Provenance.** The exact list of what went in: atom ids and versions, code references (commit or file identity), **the queries run and the form each result took (§3),** and the assembly time. This is what gets logged (§6).

### 2.1 The three kinds of database data

In this framework a lot of what you'd call "the code" lives in the database, so reading it is usually no more sensitive than reading a `.go` file. Three kinds, handled differently:

- **Definition data — the system's design, as data.** Verified against the live schema: workflows are `jsonb` columns on `agent_definitions` (`task_workflow`, `orchestrator_workflow`); tools are rows in `content_components` where `component_level = 'tool'`; prompts are text columns; scheduled jobs are in `scheduled_tasks`. (Note these are *not* where you'd assume — there is no `workflows` table and no `tools` table; check the schema, don't guess.) This is core context for most tasks here — fetched **routinely**, scoped to the task's target (the agents/workflows/tools in scope in full, neighbours summarised), and **covered by the reuse-search** so the common duplication in a data-defined system — a near-copy of an existing workflow or agent — is caught the way a duplicate function would be. Not sensitive for our own use.
- **Operational data — what's happening now.** `site_work_items`, `maintenance_queue`, `build_queue`, `orchestration_states`, `awaited_requests`, `agent_error_log`, `llm_call_log`. This is the system's own telemetry — needed mainly for debug and ops tasks. Not sensitive for our own use, but these tables get large, so the multipass caps apply (§3).
- **Content data — the output.** `sites`, `pages`, `content_items`, `research_results`, assets, and in the service case a tenant's own data. This is the smaller, **gated** set: for our own use it's just our content; in the service it's where privacy matters and reads must be bounded and, for free-form queries, confirmed.

### 2.2 Runtime evidence — what actually happened on a run

For debugging, the most useful context is often not a row's current state but the *narrative of a run*. Your tables make this reconstructable from one key — every run carries an `orchestration_id` (and a `correlation_id`) across `orchestration_states`, `llm_call_log`, and `agent_error_log` — so the bundle can pull a coherent run trace with three cheap reads:

- **The run trace / spawn tree** — `orchestration_states` for the run: the agent spawn tree (`parent_orchestration_id`, `subtree_agents`), the owner agent, the Kafka topics it used (`requests_topic`, `responses_topic`), the status, and any error. This is the "which agents were created and the messages between them" record, in one row.
- **The step sequence** — `llm_call_log` for the run, time-ordered: each `agent_type` + `step_name` + `model` + `success`. The run's actual narrative of what ran and what failed.
- **The error trail** — `agent_error_log` for the run.
- **The pod logs** — `kubectl -n ai-persona-system logs <pod> --since=<t> | grep <orchestration_id>` for the agents involved; `kubectl -n kafka …` for topic/consumer state. The layer the DB tables don't capture. Fetched through the multipass gate (§3) because logs are the worst offender for size.
- **The messages** — the Kafka messages on the run's request/response topics (headers + body), where the failure is in the message flow.

All of it is keyed by the `orchestration_id`, so the evidence is coherent — one run, not a scatter of unrelated lines.

**Run signatures (expected vs actual).** A healthy run of a workflow has an expected step sequence (the `step_name` order from `llm_call_log` on a known-good run) and an expected spawn-tree shape ending in `status = complete`. Captured once from good runs and confirmed, this is stored as authored reference. On a debug task the bundle pulls the actual run's sequence and tree, **diffs them against the signature, and surfaces the divergence point** — the step where `success` flipped, the expected step that never ran, the child agent that should have spawned and didn't. This turns "read the logs" into "matched the healthy path to step 7, then diverged here." It is verification applied to runtime: expected behaviour versus actual, with the break localized.

**Diagnostic playbooks (authored).** The existing debugging guides, the `FOCUS_*_diagnostic` docs, and the failure writeups (dispatch failures, consumer-group race, silent completion, stuck state) are authored diagnostic knowledge — known failure fingerprints, the commands that confirm them, the fixes. A fingerprint = a known failure's signature + the commands to confirm it + the fix pattern. The bundle surfaces the relevant playbook for a debug task the way it surfaces the relevant standards, seeded from those guides rather than authored fresh, and grown over time as run-signature diffs reveal new ones. (Where these live — reference atoms in `standards`, or a sibling table — is a small open decision, not yet pinned.)

### 2.3 These data and runtime capabilities are codebase-conditional

The definition-data, runtime-evidence, and run-signature capabilities above rest on facts about how the codebase is structured — that behaviour is stored as data, that a run-correlation key spans the telemetry tables, that steps are named and logged, that logs are fetchable a particular way. **All of these hold on our own codebase, so we build the optimal version against it first.** They are not assumptions baked into the engine: the stack-discovery agent discovers each per codebase (`PLAN_onboarding_agent_specs` §2.9), records which hold, and the bundle builder offers only the capabilities the codebase actually supports — degrading to a weaker form or stating "unavailable, because this codebase has no X" rather than breaking. A capability's availability scales with what the codebase provides. The limitation is named per capability there so it is not forgotten when the tool meets a codebase unlike ours.

---

## 3. What goes in full, what is only pointed at, and the multipass fetch

The rule is: **put in full only what is needed for this task at this moment; point at the rest.** A bundle full of every standard and every nearby file would bury the few that matter. Inlining the relevant pieces and linking the rest keeps the bundle focused, and — because the links are followed fresh when needed — avoids carrying stale copies of things that may have moved. This is the same "references, not copies" idea the whole design rests on, applied to one task's context.

**Query results need a different mechanism, because you don't know a result's size until you run the query.** A SELECT that looks small can return far more than expected. So the builder can include query results, but it does so in a **multipass flow** rather than running a query and dumping whatever comes back:

1. **Probe.** Run a bounded probe first — a `count`, and a `LIMIT 1` for shape — to learn how big the result is and what it looks like, without pulling it all. (For very large tables, an estimate or a counted query with a timeout, so the probe itself stays cheap.)
2. **Check against the gates.** Two gates: a **size** gate (is the row count / payload within the bundle's cap?) and a **sensitivity** gate (is this definition/operational data, which is routine, or content data, which is gated — §2.1?).
3. **Fetch the right form.** Based on the check:
   - Within size, routine → **include the rows in full.**
   - Over size → **don't dump.** Reduce instead: an **aggregate** (counts, distributions), a **representative sample** (`LIMIT N`), or a **pointer** ("N rows matched; here is the query to see them") — whichever serves the task.
   - Sensitive → **gate it:** confirm before including, or include redacted/sampled, or pointer-only.

So "inline what's needed, point at the rest" becomes, for data: probe, check, then include-in-full / reduce / point-at. An unexpectedly large result turns into an aggregate or a pointer, never an unbounded dump that swamps the bundle.

**On the builder running its own queries.** For definition and operational data the queries are known and parameterised (fetch the workflow for these agents, the work items in this state) — safe by construction, still size-probed. The builder may also propose a query for content/exploration; that goes through the sensitivity gate (confirm-not-initiate for the query) as well as the size probe. The common case is safe known reads; the freer case is the gated minority.

Caps and the sensitivity classification are configurable (per task type and, for the service, per tenant), so the gates can be tuned without changing the flow.

---

## 4. Altitude — the step type decides what to emphasise

The same task at different stages needs different context. The bundle's `step_type` says which, and the builder composes accordingly:

- **Framing or decision step** — the generator is deciding *what* to do. Give it the full intent: the why-chain, the priority profile, the direction-of-travel. Keep code detail light.
- **Implementation step** — the generator is writing code. Give it the full code context and a thin tether to intent: the constitution and the immediate purpose, not the whole why-chain.
- **Debug step** — lead with the errors and the code involved, plus the **runtime evidence** for the failing run (§2.2): the run trace, the step sequence, and — where a run signature exists — the expected-vs-actual diff that localizes the break. Add the intent needed to judge a fix.

This is the "right altitude at the right moment" idea made concrete: the bundle is not one fixed shape, it leans toward intent or toward code depending on what the step is for.

---

## 5. Two forms of the same bundle

Every bundle exists in two forms, both from one assembly:

- **Structured form** — the object with all the sections and the provenance. This is the canonical form, used by programs: the cascade router, the verification step, the logger.
- **Rendered form** — the structured form turned into readable text or markdown, for a model to read or a person to paste. It is produced fresh from the structured form each time; it is not stored as the source of truth.

The structured form is the real bundle; the rendered form is a view of it.

---

## 6. Provenance is what gets logged

The provenance section of the bundle is exactly what the decision log records as `inputs_used` for the bundle-assembly decision (`decision_log` contract §3). So assembling a bundle is a logged decision: the log keeps the atom ids and versions, the code references, **the queries run with the form each result took and its row count (full / sample / aggregate / pointer — §3),** and the time, and — combined with the versioned atoms and the versioned definition rows — that is enough to reconstruct later what context the task was actually built with. This is the "compute fresh, log what you used" rule applied to the bundle. (Definition rows carry `version`/`previous_version_id`, so a workflow or agent definition in the bundle is logged by id and version exactly as a standard atom is.)

---

## 7. Storage and lifecycle

A bundle is **derived**, not authored: it is true-at-the-time, built from sources that are themselves the real thing. It is not a source of truth and is not maintained — the active config and the code are.

- The provenance is always logged (§6).
- The full structured bundle may be stored as a task artifact so a person can inspect "what was this task built with," but it is kept as a record, not as something to edit or trust over the live sources.
- Bundles are not reused across tasks; a new task gets a new bundle.

---

## 8. What this reuses

The bundle builder is mostly composition over things already specified:

- The **authored layer** comes from the active-config read patterns already written (`PLAN_active_config_schema` §4).
- The **code context** comes from the Go analyser (tool plan Phase 1).
- The **definition and operational data** come from known, parameterised reads against the existing tables (`agent_definitions`, `content_components`, `site_work_items`, etc.), with the multipass probe in front.
- The **reuse-search over definitions** uses the existing pgvector/nomic store — index the definition rows so "is there already a workflow/agent like this" is the same similarity search already run over the corpus.
- The **provenance** maps straight onto the decision log's `inputs_used`.
- The **rendering** into text, and the **multipass fetch**, are the main genuinely new pieces.

So this contract is not much new machinery — it is the agreed shape of the thing those pieces produce together.

---

## 9. Open

- **The rendering template.** Exactly how the sections become prose for a model — ordering, headings, how pointers are shown. Best tuned by trying it on real tasks.
- **Neighbourhood depth.** How far out the call-graph slice reaches (direct callers/callees, or two hops). Tune against real tasks; too little starves the generator, too much buries it.
- **Size caps and the probe.** The exact row/payload caps per task type, and how the probe estimates size cheaply on large tables (a counted query with a timeout, or `pg_class` estimates). Tune against real tasks.
- **Sensitivity classification.** Which tables/columns count as content/sensitive versus routine — straightforward for our own use (almost all routine), the real work in the service case.
- **Run-signature capture.** How healthy-run signatures are first captured and kept current (observe known-good runs, confirm, refresh as workflows change). The diff logic and where signatures are stored.
- **Log correlation.** Whether the `orchestration_id` reliably appears in pod log lines — if it does, grep-by-run works; if not, that is worth fixing precisely so runtime evidence can be correlated. A logging convention to confirm/repair, not just consume.
- **Diagnostic-playbook home.** Whether failure fingerprints live as reference atoms in `standards` or a sibling table; seeded from the existing debugging guides.
- **Following pointers in an agentic loop.** When the generator can ask for a pointed-at item — or a fuller version of a reduced query result or a longer log slice — and get it added, rather than a person following the link. A later capability, not needed for the first version.

---

## 10. One-line state

A bundle is the task-scoped context the builder assembles fresh for one task: metadata, task/target, the authored layer (constitution, why-chain, priority profile, direction-of-travel, matched standards), the code context (in-scope code, neighbourhood signatures, reuse-search, schema, and definition data — workflows/agents/tools, which live in the database here), the operational and content data fetched through a probe-gate-fetch multipass so an oversized result becomes an aggregate or a pointer rather than a dump, pointers to the rest, and provenance. What goes in full versus pointed-at, and whether it leans to intent or code, is decided by the step type. It exists as a canonical structured form and a rendered text view; its provenance (including each query and the form its result took) is logged to the decision log; it is derived, not authoritative; and it is mostly composition over the active-config reads, the Go analyser, known parameterised data reads, and the decision log.
