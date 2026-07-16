# PILOT — "Bug Historian" council reviewer (stage 3, candidate B)

**Status: LIVE as of 2026-07-16.** Applied to `clients_db` (`postgres-clients-0`,
`ai-persona-system`) with explicit owner sign-off, via
`fixloop_eg_dartsonline/0NN_fix_proposer_v6_bug_historian.sql`. `fix-proposer`'s
`updated_at` bumped to `2026-07-16 18:24:20 UTC`; the pre-update row was
snapshotted first (`snapshot_agent`, id `f9d90a2d-cdfc-403d-a949-6e327db7c9a3`,
confirmed present in `agent_definitions_backup` — rollback available). Verified
live in the DB: `review_bug_historian` step present, wired
`review_editquality → review_bug_historian → review_guardian`,
`council_decide` and `escalate` both carry all three `review_fields`, prompt
content intact (3,986 chars). Committed:
`docs024_key_docs_latest/fixloop_eg_dartsonline/0NN_fix_proposer_v6_bug_historian.sql`
(`187a1208e`). This document remains the design record — charter, curated
context, prompt, and the patch that was applied.

---

## 1. Why this seat, and why now

Stage 3's design (`PLAN_concept_register.md`) resolved *how* to add a council
seat — mechanically, it's a new named step in `fixloop_eg_dartsonline/0NN_fix_proposer.sql`
plus a role prompt — but left *which* seat to build first as an open question,
to be answered from evidence rather than a guess.

Two candidates came out of counting which register concepts are independently
rediscovered most often across the project's documentation history (the
signal fixloop's own `FIX-036` names as the strongest one available). This
pilot is candidate B, **"bug-historian,"** chosen over candidate A
("reuse-agent," `tool-lifecycle.md`) because it is concretely anchored to real,
recent evidence rather than a general discipline: while writing up the
`missingkey=zero` structural defect behind fixloop's own real-case queue
(`STY-049`), the same failure shape — **content silently vanishing during a
rebuild/rerender, because a schema-required field went missing and nothing
flagged it** — turned up independently in **five other places** across the
platform's history, in completely different subsystems, built by different
people, years apart, none of them aware of the others:

| Concept | Where | What happened |
|---|---|---|
| `TL-001` | tool-lifecycle | An interactive tool/game, stored only as raw HTML/JS, got silently deleted by a routine content rebuild — the prose-based regression guard couldn't see script-heavy content was missing. |
| `PBP-012` | page-build-pipeline | The exact same DELETE+INSERT rebuild pattern destroyed a working A* pathfinding game; recurred independently on a second site (vonc) months later. |
| `PBP-019` | page-build-pipeline | The page assembler's own visible-content filter is a second, independent silent-drop path for the same class of content. |
| `STY-004` | styling-render-pipeline | An early defence layer (empty-section filter, schema validation) built specifically because LLM output was silently producing broken/empty sections. |
| `STY-019` | styling-render-pipeline | The visible-content filter needed a second fix because it silently dropped *intentionally* empty runtime-filled sections too. |
| `STY-049` | styling-render-pipeline | This week's incident: `missingkey=zero` silently blanked article bodies platform-wide; only one of the affected call sites is guarded so far. |
| `CLC-003` | component-lifecycle (tool-library.md) | A component regeneration silently renamed schema fields, breaking every dependent instance in one 16ms batch — the direct ancestor of today's field-contract guard. |

A council reviewer who had this history in front of it would have recognised
the shape of `STY-049` on sight, instead of it needing a full incident
write-up to notice the family it belongs to. That's the seat's whole job.

## 2. Charter

**The bug-historian judges one question only: "has this shape of failure
happened before, and does this plan account for it?"** It does not re-judge
edit quality (that's `review_editquality`) or blast radius (that's
`review_guardian`) — it has a narrower, deeper job: pattern-match the proposed
fix against a **curated history of recurring platform bug shapes**, and flag
it if either (a) the plan is *itself* proposing something with a known bad
shape, or (b) the plan fixes the symptom but not the pattern, leaving the next
occurrence uncaught.

**It cannot block.** Because the council's decision code treats *any*
reviewer's `veto` as an automatic rejection regardless of `hard_veto_from`
(verified directly in `platform/orchestration/actions/diagnose_council_decide_action.go:236-238`
— `hard_veto_from` only changes the audit label, not the outcome), giving this
seat a `veto` option would make it a second full gatekeeper, not an advisory
historian. Its prompt deliberately offers only `approve | object` — at most it
triggers a `revise` round, never an outright rejection. This is a considered
design choice, not an oversight.

## 3. Curated context (v1)

Seeded from exactly the 7 concepts in the table above — not the whole
register, not even the whole `tool-lifecycle`/`styling-render-pipeline`
categories. This keeps the pilot small, concrete, and reviewable; broadening
the curated corpus (via the register's rediscovery-frequency signal
generally, or fixloop's own `FIX-036`-named categories) is future work once
this seat proves useful.

The context block (`known_patterns` — the exact text the prompt below embeds):

```
KNOWN RECURRING PATTERN: "Silent content loss during rebuild/rerender"

This exact shape has independently recurred at least 6 times across this
platform's history, in different subsystems, built by different people:

1. An interactive tool/game (raw HTML/JS in rendered_html) was silently
   deleted by a routine content rebuild — the prose-based regression guard
   couldn't see script-heavy content was missing (tool-lifecycle TL-001).
2. The identical DELETE+INSERT rebuild pattern destroyed a working A*
   pathfinding game; recurred independently on a second site months later
   (page-build-pipeline PBP-012).
3. The page assembler's own visible-content filter is a SECOND, independent
   silent-drop path for the same class of content (page-build-pipeline
   PBP-019).
4. LLM-generated sections were silently rendering broken/empty; an early
   defence layer (schema validation, empty-section filter) was built
   specifically for this (styling-render-pipeline STY-004).
5. That same visible-content filter needed a second fix because it silently
   dropped INTENTIONALLY empty runtime-filled sections too (styling-render-
   pipeline STY-019).
6. A component regeneration silently renamed schema fields, breaking every
   dependent instance in one 16ms batch (component-lifecycle CLC-003).
7. MOST RECENT (2026-07-16): Go's template engine renders any missing
   required field as empty with NO ERROR (Option("missingkey=zero")); this
   silently blanked article bodies platform-wide. Only ONE call site is
   guarded so far — the root behaviour itself remains generic and unpatched,
   so ANY other unaudited call site has the identical exposure
   (styling-render-pipeline STY-049).

THE PATTERN: something is silently dropped, overwritten, or rendered empty —
no error, no warning, no failed work item — because a rebuild/regeneration/
render path didn't check whether what it was about to discard or skip was
actually required or actually present. Every instance was caught only after
real content was already lost.

WHAT TO LOOK FOR in a fix plan: (a) does this plan touch a rebuild, rerender,
regeneration, or template-render code path? (b) if so, does it introduce a
NEW way for something required-but-missing to fail silently rather than
loudly? (c) does the plan's fix patch ONE call site of a shared underlying
mechanism (like missingkey=zero) while leaving the mechanism itself generic
and exploitable elsewhere? (d) is the plan fixing a SYMPTOM of this pattern
without a corresponding "fail loud, not silent" guard?
```

## 4. Prompt template (draft, matches the existing reviewers' style/contract exactly)

```
# Council reviewer: BUG HISTORIAN

You judge one thing: does this platform have a documented history of this
exact failure shape, and does the plan account for it? You change nothing;
you judge.

{{.known_patterns}}

Judge the plan against the pattern(s) above: (a) does the plan's own proposed
change risk reproducing a known pattern (e.g. adding a new silent-discard path
in a rebuild/rerender/render code path); (b) does the plan fix only a
SYMPTOM of a known pattern, leaving the underlying mechanism exploitable
elsewhere — if so, name specifically what else is still exposed; (c) is there
a simpler, already-proven fix shape from a past occurrence that this plan
should reuse instead of inventing a new one.

Verdicts: approve (no known-pattern concern), object (this plan risks or
incompletely addresses a documented recurring pattern — say which one and
why, in objections). You do NOT have a veto — if you see something severe
enough that this fix should not proceed at all, put it in objections at
"high" severity and trust the router (severe cases naturally raise revise
rounds; if it's a true architecture-level concern, note that explicitly so a
human sees it either way).

CHECKS: if a verdict hinges on a fact a read-only SQL query could settle,
put that query in checks as {"sql": "SELECT ...", "why": "what this settles"}
— SELECT/WITH only, never writes. Write checks ONLY against the tables/columns
in the Schema section below.

## Schema (the ONLY tables available to checks)
{{.schema_hint.text}}

## The diagnosis
{{.diagnosis_row.conclusion}}

## The plan
{{.plan_persisted.plan_json}}

## Output — ONLY this JSON
{"reviewer": "bug_historian", "verdict": "approve|object", "objections": [{"edit": 1, "problem": "...", "severity": "low|medium|high"}], "missing": [], "checks": [{"sql": "SELECT ...", "why": "what this settles"}], "notes": "..."}
```

## 5. Exact wiring changes (all inside `0NN_fix_proposer.sql`'s workflow JSON)

The two existing reviewers run sequentially (`review_editquality` →
`review_guardian` → `council_decide`, chained by `next_step` — not a parallel
fan-out, confirmed by reading the live file). The new reviewer inserts into
that same chain, between the two:

```
persist_plan → review_editquality → review_bug_historian (NEW) → review_guardian → council_decide
```

Four edits, all mechanical:

1. **`review_editquality.next_step`**: `'review_guardian'` → `'review_bug_historian'`
2. **New step `review_bug_historian`** (full block below)
3. **`council_decide.config.review_fields`**: add `'review_bug_historian.result'`
   to the array (alongside the existing two)
4. **`repropose.config.input_fields`**: add `'review_bug_historian'` (so a
   repropose round can see and address its objections, same as it already
   does for the other two reviewers)

New step block (drop-in, matches the file's existing style exactly):

```sql
        'review_bug_historian', jsonb_build_object(
          'action', 'execute_llm_prompt',
          'description', 'Council reviewer — bug historian: does this platform have a documented history of this failure shape? Advisory only (no veto).',
          'output_field', 'review_bug_historian',
          'next_step', 'review_guardian',
          'config', jsonb_build_object(
            'error_step', 'complete_refused',
            'ai_service', jsonb_build_object(
              'model', 'claude-sonnet-4-6',
              'provider', 'anthropic',
              'api_key_env_var', 'ANTHROPIC_API_KEY',
              'max_tokens', 3000
            ),
            'temperature', 0.0,
            'input_fields', jsonb_build_array('diagnosis_row', 'plan_persisted', 'schema_hint', 'known_patterns'),
            'output_format', 'json',
            'prompt_template',
              -- (the full prompt in §4 above, as one Go template string,
              --  matching the review_editquality/review_guardian literal-
              --  concatenation style used elsewhere in this file)
          )
        ),
```

Note: `known_patterns` needs to reach the step as a static input field. The
simplest mechanism consistent with this file's existing patterns is a new
`query_database`-free constant — either (a) a `load_known_patterns` step
(no-op `query_database` returning a literal, mirroring `load_schema_hint`'s
shape) inserted right after `load_schema_hint`, or (b) inlining the
`known_patterns` text directly into the prompt template as a literal string
rather than a templated field (simpler, avoids a step for a value that never
changes at run time). Recommend (b) for v1 — one fewer moving part — and only
promote it to a queried/versioned value if a future iteration wants the
curated corpus to grow without editing this SQL file each time.

## 6. What happened (applied 2026-07-16)

1. Wrote the versioned file
   `0NN_fix_proposer_v6_bug_historian.sql` with the five edits (the four
   originally planned, plus one caught during a final review pass: the
   `escalate` step's `review_fields` also needed the new reviewer, so the
   human hand-off package shows all three reviews, not just two).
2. No chassis image dependency — the new reviewer uses only the existing
   `execute_llm_prompt` action, so this was a pure DB change, live immediately.
3. Ran it against `clients_db` after explicit, specifically-named owner
   confirmation (an auto-mode safety classifier had blocked even a read-only
   query until the target was named — a deliberate gate, not worked around).
   `snapshot_agent` backed up the prior row first (confirmed present in
   `agent_definitions_backup`).
4. Verified post-apply: `review_bug_historian` present and correctly wired,
   both `review_fields` arrays (`council_decide` and `escalate`) carry all
   three reviewers, prompt content intact.

**Still open — watch the next real fix-loop run** to confirm the new
reviewer's output parses cleanly in production traffic (only verified via
direct DB read so far, not a live workflow execution) — a malformed reviewer
output fails the step closed by design, so a problem would surface
immediately, not silently, but it hasn't been exercised end to end yet.

**Candidate A (reuse-agent, `tool-lifecycle.md`) remains unbuilt** — the same
process (charter → curated context → prompt → patch) would apply if a second
seat is wanted later.
