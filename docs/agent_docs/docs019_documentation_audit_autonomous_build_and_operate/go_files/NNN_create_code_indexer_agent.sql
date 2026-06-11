-- FILE: migrations/NNN_create_code_indexer_agent.sql  (renumber to your sequence)
--
-- DRAFT — modelled on the real agent_definitions rows you sent
-- (site-adoption-agent: a code-driven orchestrator; build-site-planner:
-- input_contract/output_contract shape). Confirm the live schema before
-- applying:
--
--   \d agent_definitions
--
-- Specifically verify: NOT NULL columns, defaults (id / created_at / etc.),
-- and the CHECK on agent_category (allowed: strategist, executor, analyst,
-- integrator, coordinator, specialist — per doc 003). This INSERT sets the
-- columns the real rows populate and leaves the rest to defaults.
--
-- The code-indexer is an ORCHESTRATOR (every agent is): its whole workflow is
-- request the analyser adapter → index the returned symbols → complete. The
-- substantive work (parsing) runs in the analyser adapter pod; the indexing
-- upsert runs in the index_code_symbols action. So this orchestrator does only
-- coordination — appropriate to run via the generic entry point or a wrapper.
--
-- Snapshot before re-applying to an existing row:  SELECT snapshot_agent('code-indexer');

INSERT INTO agent_definitions (
    type, display_name, description,
    category, agent_category, status,
    image_repository, image_tag,
    resources, topics, health_config,
    capabilities, domain_tags,
    default_config, input_contract, output_contract,
    is_active
) VALUES (
    'code-indexer',
    'Code Indexer',
    'Indexes a source repository into the code_symbols table. Asks the analyser adapter to parse the repo at a ref into symbols, then upserts them (embedding changed symbols, pruning symbols absent from the commit). The retrieval side is the lookup_code_symbols action used by other agents.',
    'code-driven',
    'integrator',
    'experimental',
    'docker.io/aqls/agent-chassis',
    'v1.0.1059',
    '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
    '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
    '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb,
    '["code-indexing", "symbols", "embeddings"]'::jsonb,
    ARRAY['code-context', 'indexing'],
    '{
      "workflow": {
        "start_step": "request_analysis",
        "steps": {
          "request_analysis": {
            "action": "request_repo_analysis",
            "config": {
              "owner_field": "input_data.owner",
              "repo_field": "input_data.repo",
              "ref_field": "input_data.ref",
              "language": "go"
            },
            "next_step": "index_symbols",
            "output_field": "repo_analysis",
            "description": "Ask the analyser adapter to parse the repo at ref into symbols; await the reply"
          },
          "index_symbols": {
            "action": "index_code_symbols",
            "config": {
              "repo_field": "repo_analysis.repo",
              "commit_field": "repo_analysis.commit_sha",
              "analysis_field": "repo_analysis.output",
              "embedding_service": {
                "provider": "ollama",
                "model": "nomic-embed-text",
                "api_url": "http://ollama-adapter.ai-persona-system.svc.cluster.local:11434"
              }
            },
            "next_step": "complete",
            "output_field": "index_result",
            "description": "Upsert symbols into code_symbols, embed changed ones, prune by commit"
          },
          "complete": {
            "action": "complete_workflow",
            "config": {"output_fields": ["repo_analysis", "index_result"]},
            "description": "Indexing complete"
          }
        }
      },
      "processing_mode": "orchestrator",
      "timeout_seconds": 300
    }'::jsonb,
    '{"required": ["owner", "repo"], "optional": ["ref"], "expects": {"input_data.owner": "string - GitHub owner/org", "input_data.repo": "string - repository name", "input_data.ref": "string - branch, tag, or SHA (optional; defaults to HEAD)"}}'::jsonb,
    '{"produces": {"repo_analysis": "analyser AnalyseResult (owner, repo, ref, commit_sha, output)", "index_result": "counts: symbols, upserted, embedded, pruned"}}'::jsonb,
    true
);
