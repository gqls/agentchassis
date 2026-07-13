INSERT INTO agent_definitions (
    type, display_name, description, category, default_config, is_active
) VALUES (
    'rag-test-agent',
    'RAG Test Agent',
    'Minimal deterministic agent exercising rag_index and rag_lookup to verify chassis registration. Takes input_data.content and input_data.query.',
    'experimental',
    '{
      "workflow": {
        "start_step": "index_content",
        "processing_mode": "orchestrator",
        "timeout_seconds": 120,
        "steps": {
          "index_content": {
            "action": "rag_index",
            "config": {
              "content_field": "input_data.content",
              "collection": "flywheel_b_chassis_test",
              "chunk_size": 500,
              "chunk_overlap": 100,
              "embedding_service": {
                "provider": "ollama",
                "model": "nomic-embed-text",
                "api_url": "http://ollama-adapter.ai-persona-system.svc.cluster.local:11434"
              }
            },
            "next_step": "lookup_query",
            "output_field": "index_result",
            "description": "Index the provided content into the test collection"
          },
          "lookup_query": {
            "action": "rag_lookup",
            "config": {
              "query_field": "input_data.query",
              "collection": "flywheel_b_chassis_test",
              "top_k": 3,
              "embedding_service": {
                "provider": "ollama",
                "model": "nomic-embed-text",
                "api_url": "http://ollama-adapter.ai-persona-system.svc.cluster.local:11434"
              }
            },
            "next_step": "complete",
            "output_field": "lookup_result",
            "description": "Query the test collection"
          },
          "complete": {
            "action": "complete_workflow",
            "config": {
              "output_fields": ["index_result", "lookup_result"],
              "success_message": "RAG chassis test complete"
            }
          }
        }
      }
    }'::jsonb,
    true
)
ON CONFLICT (type, version) WHERE deleted_at IS NULL DO UPDATE
                                                           SET default_config = EXCLUDED.default_config,
                                                           updated_at = NOW();
