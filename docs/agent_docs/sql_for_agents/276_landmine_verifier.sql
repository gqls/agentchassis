-- 276_landmine_verifier.sql
--
-- landmine-verifier -- RFC_005 section 3.2 (owner-ratified 2026-07-31): a dedicated,
-- single-pass verifier for LANDMINES.md entries, NOT the multi-seat council
-- (wrong tool for a fact-check task; see the RFC for why) and NOT a
-- mechanical sync-script grep (the owner's explicit choice, over the cheaper
-- alternative, because a footprint-exists check alone would miss a "the tell
-- no longer matches the check" drift).
--
-- Reuses existing actions only -- no new Go code, no platform/-scope change:
--   query_database       -- load one entry's body from doc_notes (categories ? 'landmine')
--   diagnose_code_lookup -- the SAME action the diagnose loop uses for its own
--                           "[static]" evidence citations (confirmed live in
--                           bugs_open/155's verification run); reused here
--                           rather than building a second code-search path
--   execute_llm_prompt   -- twice: once to derive code_checks from the entry's
--                           footprint text, once for the final verdict
--   append_doc_note      -- persists the verdict as a NEW doc_notes row
--                           (categories: landmine-verification). Deliberately
--                           does NOT edit LANDMINES.md itself -- that file is
--                           human-curated and append-only by design (the same
--                           concurrency reasoning that guards it against
--                           careless deletion applies to an automated writer
--                           too), so the verdict is queryable, not in-file.
--
-- One entry per invocation (input_data.source = 'LANDMINES.md#<slug>'),
-- straight-line pipeline, no hypothesis iteration -- that is what makes this
-- "single-pass" rather than a reuse of the full diagnose-orchestrator loop.
--
-- Known limitation, stated rather than hidden (see PLAN_2026-07-31 Part A):
-- "the check" field is free text and sometimes names a shell/kubectl command
-- (strings /app/agent-chassis, grep -ac ...). No existing action runs an
-- arbitrary shell command from inside a workflow. Those checks fall back to
-- LLM judgement on internal consistency only, and the verdict says so via
-- NEEDS_HUMAN_REVIEW rather than guessing past what could actually be run.
--
-- Trigger: not wired to a schedule yet. Manually, per entry, via a kafka
-- message to system.agent.generic.requests (see
-- scripts/initial_messages/130_section_editor/073_section_editor.sh for the
-- message shape) with agent_type "landmine-verifier" and
-- input_data {"source": "LANDMINES.md#<slug>", "ref": "<git ref under test>"}.
-- Dispatching automatically on new/changed entries (rather than by hand) is
-- the open wiring question the PLAN leaves for the next increment.

INSERT INTO agent_definitions (
    type,
    display_name,
    description,
    category,
    default_config,
    is_active,
    capabilities,
    image_repository,
    image_tag,
    resources,
    topics,
    version,
    agent_category,
    status,
    domain_tags,
    input_contract,
    output_contract
) VALUES (
    'landmine-verifier',
    'Landmine Verifier',
    'Single-pass fact-check of one LANDMINES.md entry: confirms the footprint still resolves and the tell still matches the check, via a fresh code lookup. Persists a verdict to doc_notes; never edits LANDMINES.md.',
    'specialist',
    '{"workflow": {"start_step": "load_entry", "processing_mode": "orchestrator", "timeout_seconds": 300, "steps": {"load_entry": {"action": "query_database", "config": {"query": "SELECT source, body FROM doc_notes WHERE source = $1 AND categories ? ''landmine'' ORDER BY created_at DESC LIMIT 1", "params": ["input_data.source"], "output_format": "object"}, "next_step": "derive_checks", "description": "Load one LANDMINES.md entry (any of its footprint rows carries the same body)", "output_field": "entry"}, "derive_checks": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-opus-4-6", "provider": "anthropic", "max_tokens": 1200, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["entry"], "output_format": "json", "prompt_template": "You are extracting mechanically-checkable code references from a LANDMINES.md entry.\n\nThe entry (raw markdown, one bullet per field):\n\n{{.entry.body}}\n\nLook at the `footprint` field specifically (the line starting \"- **footprint:**\"). It lists one or more file paths, `path:Symbol` references, table names, or bare commands, separated by commas.\n\nFor each DISTINCT item in the footprint list that names something a code search could confirm still exists, emit one entry:\n- kind \"symbol\" -- for a `path:Symbol` reference (a file path followed by a colon and a Go identifier)\n- kind \"ls\" -- for a bare file or directory path with no symbol\n- kind \"content\" -- for anything else worth a text search: a table name, a distinctive string, a command name\n\nSkip items that are pure prose description (a phrase like \"any docs/agent_docs/.../HANDOFF_*.md belonging to another thread\" describes a category, not a specific target -- do not emit a check for it).\n\nReturn ONLY this JSON shape, nothing else:\n{\"code_checks\": [{\"kind\": \"symbol|ls|content\", \"query\": \"the exact path or path:Symbol or search text\", \"why\": \"one clause: what this confirms\"}]}\n\nIf nothing in the footprint is mechanically checkable, return {\"code_checks\": []}."}, "next_step": "run_checks", "description": "Extract mechanically-checkable footprint items as code_checks", "output_field": "derived"}, "run_checks": {"action": "diagnose_code_lookup", "config": {"code_check_fields": ["derived.result.code_checks"], "max_checks": 10}, "next_step": "verify", "description": "Run the derived checks against the live code index", "output_field": "lookup"}, "verify": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-opus-4-6", "provider": "anthropic", "max_tokens": 1200, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["entry", "lookup", "input_data"], "output_format": "json", "prompt_template": "You are verifying whether a LANDMINES.md entry is still an accurate description of the system -- NOT diagnosing a new bug, NOT iterating on a hypothesis. One pass, one verdict.\n\nThe entry:\n\n{{.entry.body}}\n\nFresh code-lookup results, run just now against the current checked-out ref:\n\n{{.lookup.results_text}}\n\nJudge:\n1. Does the footprint still resolve (the lookup results show the file/symbol/table exists)?\n2. Where the lookup found the cited code, does it still say what \"the tell\" and \"the check\" describe? (Read the actual excerpt, don''t assume.)\n3. Is the entry internally consistent -- does the footprint still relate to the \"fires when\" clause, or has the code moved on since this was written?\n\nIf the lookup ran no checks at all (footprint had nothing mechanically checkable, or every check errored), say so plainly and judge only on internal consistency of the text itself.\n\nReturn ONLY this JSON shape:\n{\n  \"status\": \"STILL_VALID | STALE | NEEDS_HUMAN_REVIEW\",\n  \"rationale\": \"one paragraph: what you found, citing the lookup results by file/line where possible\",\n  \"body\": \"a short markdown note recording the verdict, formatted as: **last verified (landmine-verifier): <STATUS>.** <one-sentence rationale>. Checked against {{.input_data.ref}}.\"\n}\n\nSTALE means the lookup evidence contradicts the entry (footprint gone, or the tell no longer matches). NEEDS_HUMAN_REVIEW means genuinely ambiguous -- a symbol that moved but might still guard the same thing, a check that couldn''t run. Do not guess past what the lookup evidence actually shows."}, "next_step": "persist_verdict", "description": "Single-pass verdict: STILL_VALID / STALE / NEEDS_HUMAN_REVIEW", "output_field": "verdict"}, "persist_verdict": {"action": "append_doc_note", "config": {"subject_type": "landmine", "subject_key_field": "input_data.source", "note_body_field": "verdict.result.body", "note_categories": ["landmine-verification"], "note_source": "landmine-verifier"}, "next_step": "complete", "description": "Persist the verdict as a new doc_notes row (never edits LANDMINES.md itself)", "output_field": "persisted"}, "complete": {"action": "complete_workflow", "config": {"output_fields": ["verdict", "persisted"]}, "description": "Landmine verification complete"}}}, "processing_mode": "orchestrator", "timeout_seconds": 300}'::jsonb,
    true,
    '["doc-review", "landmine-verification"]'::jsonb,
    'docker.io/aqls/agent-chassis',
    'latest',
    '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "128Mi"}}'::jsonb,
    '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
    1,
    'specialist',
    'experimental',
    '["docs", "review", "landmine", "fact-check"]'::jsonb,
    '{"required": ["source"], "optional": ["ref"], "description": "source = the LANDMINES.md#<slug> value (doc_notes.source). ref = git ref under test, for the verdict body only (defaults unset)."}'::jsonb,
    '{"produces": {"verdict": "STILL_VALID | STALE | NEEDS_HUMAN_REVIEW, with rationale", "persisted": "doc_notes row id"}}'::jsonb
)
ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
    description = EXCLUDED.description,
    image_tag = EXCLUDED.image_tag,
    input_contract = EXCLUDED.input_contract,
    output_contract = EXCLUDED.output_contract,
    status = EXCLUDED.status,
    agent_category = EXCLUDED.agent_category,
    domain_tags = EXCLUDED.domain_tags,
    updated_at = NOW();
