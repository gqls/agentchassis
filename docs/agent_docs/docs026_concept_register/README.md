# docs026 — Concept Register

Created 2026-07-13. **Stage 1 complete as of 2026-07-13:** 1,627 concepts across
107 category register files, consolidated from 2,185 raw extraction blocks
swept from ~4,111 files under `docs/`. **Stage 2 complete 2026-07-14** (see
`006_VERIFICATION_stage2.md`). 4 more concepts added 2026-07-16 for a subsystem
that shipped after extraction froze (1,631 total) — see `register/000_concept_index.md`'s
"2026-07-16 addition" note. Master index at `register/000_concept_index.md`;
final taxonomy note at `005_TAXONOMY_final.md`.

A complete register of every concept — scope, responsibility,
or behaviour — found anywhere under `docs/`, classified and status-tagged, built
as **stage 1** of a three-stage programme:

1. **Extract (this directory):** sweep every file under `docs/`, extract concepts,
   classify into categories, tag status signals, record provenance.
2. **Verify (later):** analyse the agent-chassis code, workflows, and DB to determine
   the true state of each concept (the extraction's status is only a *documentary
   signal* — what the docs claim, not verified ground truth).
3. **Council agents (later still):** create an expert agent per concept area — fully
   versed in its responsibilities and provenance — to join the council of decision
   makers in the diagnosis/fix loop (see
   `docs024_key_docs_latest/fixloop_eg_dartsonline/SUMMARY_write_step_position_2026-07-12.md`).

Nothing outside this directory is modified by this work.

## Categories are open

The docs024 documentation index provides a starting spine of categories, but the
taxonomy is NOT restricted to it. New categories are expected and welcome wherever
the content demands them — extraction agents propose them freely (`NEW:<slug>`),
and consolidation settles the final set. The goal is categories that could each
back an expert council agent, not conformance to the existing index.

## Directory layout

```
README.md                  — this file: method + format specs
PLAN_concept_register.md   — the living 3-stage plan, current status, backlog
RUNBOOK_concept_register.md — human-operator tasks (rotations, judgment calls, resumes)
RUNNING_NOTES_concept_register.md — turn-by-turn discussion log, updated every turn
SUMMARY_where_we_are_2026-07-16.md — calm, read-aloud narrative summary (companion to PLAN)
PILOT_bug_historian_reviewer.md — stage-3 pilot spec: charter, curated context, prompt, exact patch (not yet applied)
001_PROMPT_charter.md      — the rewritten task charter
002_PLAN_extraction.md     — stage-1 work-unit ledger (26 units), extraction-specific
003_TAXONOMY_seed.md       — starting categories + tagging rules (open-ended)
005_TAXONOMY_final.md      — final taxonomy note (stage 1)
006_VERIFICATION_stage2.md — stage-2 method + full findings (105 corrections)
extractions/UNN_<slug>.md  — raw per-unit harvest (one file per work unit)
register/<category>.md     — consolidated concept entries, one file per category
register/000_concept_index.md — master index: every concept, one line, by category
stage2_workflow.js         — the stage-2 verification workflow script (reusable for superseded/abandoned sweep)
```

## What counts as a concept

A **nameable scope of responsibility or behaviour** — something an expert agent
could own, a doc section could define, or a fix-loop council member could hold an
opinion about. Examples: "work item lifecycle", "hard allowlist safety check in the
fix-implementer", "asset locking mirrors page_components", "section-contrast model",
"imagery kind enum + chk_kind constraint". Granularity guide: bigger than a single
fact, smaller than a whole subsystem. Missions, pipelines, mechanisms, conventions,
contracts, agents, tools, and *ideas that were never built* all qualify.

Do NOT create one concept per paragraph — merge repeats; a concept mentioned in
forty documents is one concept with forty provenance entries (cite the best 3–6).

## Extraction file format

Each work unit produces `extractions/UNN_<slug>.md`:

```markdown
# EXTRACTION UNN — <source area>
Extracted 2026-07-13. Files in scope: <n>. Concepts found: <m>.

## Coverage
| file | treatment |
|---|---|
| <relative path> | full \| family-latest \| family-delta \| header-scan \| skipped-binary \| skipped-generated |

## Concepts
### <concept name>
- **category:** <slug from 003_TAXONOMY_seed.md, or `NEW:<proposed-slug>`>
- **status-signal:** deployed | partial | aspirational | superseded | abandoned | unknown
- **status-evidence:** <the dated phrase/table/section that justifies the signal>
- **what:** 2–4 sentences — scope, responsibility, behaviour.
- **sources:** <path#section, best 3–6>
- **relations:** <related concept names, free text>
- **verify-later:** <code paths / DB tables / workflow names to check in stage 2>
```

### Treatment rules

- **Version families** (`name.md`, `name(1).md` … `name(N).md`): read the
  highest-numbered fully (`family-latest`); scan earlier ones only for concepts
  absent from the latest — dropped ideas are often the superseded/abandoned
  concepts we most want (`family-delta`).
- **SQL files**: agent-definition seeds are concept-rich (each agent/workflow IS a
  concept — read header comments and prompt bodies). Table DDL: read comments and
  constraints. Mechanical patches: `header-scan`.
- **Shell/Go/JS/py files**: read header comments for intent (`header-scan`); do not
  analyse code bodies (that is stage 2).
- **Binaries (png/jpg), generated data (large json/html/css), site captures**: list
  as `skipped-binary` / `skipped-generated`. Every file in the unit MUST appear in
  the coverage table — that is the audit trail for "every single file was accounted for".

## Status-signal vocabulary

- **deployed** — doc states it is live/verified (dates, ✅ tables, "committed, tests green").
- **partial** — some phases live, some not; or live but with known quality gaps.
- **aspirational** — planned/designed, no claim of implementation.
- **superseded** — an explicit replacement exists (name it in `relations`).
- **abandoned** — idea appears in older docs/versions and silently vanishes.
- **unknown** — cannot tell from documents alone.
- **convention** *(added in stage 2, 2026-07-14)* — a design doctrine, methodology,
  or one-off analysis, not a code/DB/infra artifact. Stage 1 extraction had no way
  to express this and defaulted such concepts to `deployed` (the docs describe them
  as current/established practice); stage 2 verification found ~48 cases where this
  produced a misleading "deployed" tag with nothing to point to in the codebase. See
  `006_VERIFICATION_stage2.md`.

## Consolidation format (register/)

One file per category: `register/<category-slug>.md`, entries sorted by concept id
`<PREFIX>-NNN` (prefix per category, assigned at consolidation). Same fields as the
extraction entry plus `id` and a merged `sources` list. The index file lists every
concept as `| id | name | status | one-line summary | register file |`.
