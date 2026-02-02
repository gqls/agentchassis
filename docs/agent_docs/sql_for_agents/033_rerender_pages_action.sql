/*not sql just the trigger - using an inline workflow, but we now have an agent workflow type='rerender-agent':
#!/bin/bash
# Re-render all deployed pages with current components
# This applies component updates (head, header, footer, CSS links) without regenerating content

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"
SITE_ID="4851f6fc-71cf-4160-a270-e03d6d3e0732"
DOMAIN="leopardessconsulting.co.uk"

echo "========================================="
echo "Re-rendering site pages with updated components"
echo "========================================="
echo "  Correlation ID:      $CORRELATION_ID"
echo "  Orchestration ID:    $ORCHESTRATION_ID"
echo "  Site ID:             $SITE_ID"
echo "  Domain:              $DOMAIN"
echo "  Time:                $TIMESTAMP"
echo "========================================="

kubectl -n kafka run -i --rm kcat-rerender-pages \
--image=edenhill/kcat:1.7.1 \
--restart=Never -- \
kcat -P \
-b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
-t system.agent.generic.requests \
-H correlation_id=$CORRELATION_ID \
-H orchestration_id=$ORCHESTRATION_ID \
-H request_id=$REQUEST_ID \
-H message_id=$MESSAGE_ID \
-H message_type=request \
-H client_id=$CLIENT_ID \
-H action=process \
-H sender_agent_type=cli \
-H sender_agent_id=cli-user \
-H timestamp=$TIMESTAMP <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"rerender_pages","steps":{"rerender_pages":{"action":"rerender_site_pages","config":{"site_id_field":"input_data.site_id","domain_field":"input_data.domain","include_statuses":["deployed","active"]},"description":"Re-render all pages with current components","next_step":"deploy_pages","output_field":"rerender_result"},"deploy_pages":{"action":"loop","config":{"items_field":"rerender_result.pages","item_variable":"current_page","max_iterations":50,"sub_workflow":{"start_step":"commit_page","steps":{"commit_page":{"action":"git_commit","config":{"domain_field":"input_data.domain","content_field":"current_page.html","page_field":"current_page","commit_message":"Re-render page: {{.filename}}"},"description":"Commit re-rendered page via git adapter","next_step":"done","output_field":"commit_result"},"done":{"action":"loop_complete"}}}},"description":"Deploy each re-rendered page","next_step":"complete","output_field":"deploy_result"},"complete":{"action":"complete_workflow","config":{"output_fields":["rerender_result","deploy_result"]}}}}},"input_data":{"site_id":"${SITE_ID}","domain":"${DOMAIN}"}}
JSON

echo ""
echo "Message sent. Monitor with:"
echo "  kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=100"*/

---

-- sql for workflow:


-- Rerender Pages Agent
-- Re-assembles all deployed pages with current components
-- Use for applying component updates without regenerating content

INSERT INTO agent_definitions (
    type,
    version,
    display_name,
    description,
    category,
    default_config,
    is_active,
    capabilities,
    input_contract,
    output_contract
) VALUES (
    'rerender-pages',
    1,
    'Rerender Pages Agent',
    'Re-assembles all deployed pages with current components (head, header, footer). Use when component templates are updated, CSS links change, or navigation structure changes.',
    'builder',
    '{
        "processing_mode": "orchestrator",
        "timeout_seconds": 300,
        "workflow": {
            "start_step": "rerender_pages",
            "steps": {
                "rerender_pages": {
                    "action": "rerender_site_pages",
                    "config": {
                        "site_id_field": "input_data.site_id",
                        "domain_field": "input_data.domain",
                        "include_statuses": ["deployed", "active"]
                    },
                    "description": "Re-render all pages with current components",
                    "next_step": "deploy_pages",
                    "output_field": "rerender_result"
                },
                "deploy_pages": {
                    "action": "loop",
                    "config": {
                        "items_field": "rerender_result.pages",
                        "item_variable": "current_page",
                        "max_iterations": 50,
                        "sub_workflow": {
                            "start_step": "commit_page",
                            "steps": {
                                "commit_page": {
                                    "action": "git_commit",
                                    "config": {
                                        "domain_field": "input_data.domain",
                                        "content_field": "current_page.html",
                                        "page_field": "current_page",
                                        "commit_message": "Re-render page: {{.slug}}"
                                    },
                                    "description": "Commit re-rendered page via git adapter",
                                    "next_step": "done",
                                    "output_field": "commit_result"
                                },
                                "done": {
                                    "action": "loop_complete"
                                }
                            }
                        }
                    },
                    "description": "Deploy each re-rendered page",
                    "next_step": "complete",
                    "output_field": "deploy_result"
                },
                "complete": {
                    "action": "complete_workflow",
                    "config": {
                        "output_fields": ["rerender_result", "deploy_result"]
                    }
                }
            }
        }
    }',
    true,
    '["rerender", "maintenance", "components", "css"]',
    '{"required": [], "optional": ["site_id", "domain"], "description": "Provide either site_id or domain to identify the site"}',
    '{"produces": {"rerender_result": "Pages rendered with updated components", "deploy_result": "Git commit results for each page"}}'
)
ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
                                                                                                                                                                                                                                                                                                                                                                                                      description = EXCLUDED.description,
                                                                                                                                                                                                                                                                                                                                                                                                      default_config = EXCLUDED.default_config,
                                                                                                                                                                                                                                                                                                                                                                                                      capabilities = EXCLUDED.capabilities,
-- one page at a time
-- Update rerender-pages agent to use loop-based workflow
-- This processes one page at a time, avoiding large message payloads
--
-- Flow: get_pages → loop → (render_page → deploy_page → update_status) → complete

-- ============================================================
-- 1. Update rerender-pages agent workflow
-- ============================================================
UPDATE agent_definitions
SET default_config = '{
    "workflow": {
        "start_step": "get_pages",
        "steps": {
            "get_pages": {
                "action": "get_pages_for_rerender",
                "config": {
                    "site_id_field": "input_data.site_id",
                    "domain_field": "input_data.domain",
                    "include_statuses": ["deployed", "active"]
                },
                "description": "Get page metadata for rerender loop",
                "output_field": "rerender_pages",
                "next_step": "deploy_loop"
            },
            "deploy_loop": {
                "action": "loop",
                "config": {
                    "items_field": "rerender_pages.pages",
                    "item_variable": "current_page",
                    "mode": "sequential",
                    "max_iterations": 50,
                    "sub_workflow": {
                        "start_step": "render_page",
                        "steps": {
                            "render_page": {
                                "action": "rerender_single_page",
                                "config": {
                                    "page_id_field": "current_page.page_id",
                                    "site_id_field": "rerender_pages.site_id",
                                    "domain_field": "rerender_pages.domain",
                                    "max_nav_items": 6
                                },
                                "description": "Render single page from stored sections",
                                "output_field": "rendered_page",
                                "next_step": "deploy_page"
                            },
                            "deploy_page": {
                                "action": "git_commit",
                                "config": {
                                    "repo_name": "sites",
                                    "domain_field": "rendered_page.domain",
                                    "content_field": "rendered_page.html",
                                    "filename_field": "rendered_page.filename",
                                    "commit_message": "Rerender: {{.filename}}"
                                },
                                "description": "Deploy rendered page to git",
                                "output_field": "deploy_result",
                                "next_step": "update_status"
                            },
                            "update_status": {
                                "action": "update_page_status",
                                "config": {
                                    "status": "deployed",
                                    "page_id_field": "current_page.page_id",
                                    "commit_from": "deploy_result.commit_sha"
                                },
                                "description": "Update page status in database",
                                "output_field": "status_updated",
                                "next_step": "loop_done"
                            },
                            "loop_done": {
                                "action": "loop_complete",
                                "description": "Page rerender complete"
                            }
                        }
                    }
                },
                "description": "Loop through pages, render and deploy each one",
                "output_field": "pages_processed",
                "next_step": "complete"
            },
            "complete": {
                "action": "complete_workflow",
                "config": {
                    "output_fields": ["rerender_pages", "pages_processed"]
                },
                "description": "Rerender complete"
            }
        }
    },
    "processing_mode": "orchestrator",
    "timeout_seconds": 1800
}'::jsonb,
version = version + 1,
updated_at = NOW()
WHERE type = 'rerender-pages';

-- ============================================================
-- 2. If agent doesn't exist, create it
-- ============================================================
INSERT INTO agent_definitions (type, name, description, default_config, version)
SELECT
    'rerender-pages',
    'Rerender Pages Agent',
    'Re-renders all deployed pages with current components (header, footer, nav, contact info)',
    '{
        "workflow": {
            "start_step": "get_pages",
            "steps": {
                "get_pages": {
                    "action": "get_pages_for_rerender",
                    "config": {
                        "site_id_field": "input_data.site_id",
                        "domain_field": "input_data.domain",
                        "include_statuses": ["deployed", "active"]
                    },
                    "description": "Get page metadata for rerender loop",
                    "output_field": "rerender_pages",
                    "next_step": "deploy_loop"
                },
                "deploy_loop": {
                    "action": "loop",
                    "config": {
                        "items_field": "rerender_pages.pages",
                        "item_variable": "current_page",
                        "mode": "sequential",
                        "max_iterations": 50,
                        "sub_workflow": {
                            "start_step": "render_page",
                            "steps": {
                                "render_page": {
                                    "action": "rerender_single_page",
                                    "config": {
                                        "page_id_field": "current_page.page_id",
                                        "site_id_field": "rerender_pages.site_id",
                                        "domain_field": "rerender_pages.domain",
                                        "max_nav_items": 6
                                    },
                                    "description": "Render single page from stored sections",
                                    "output_field": "rendered_page",
                                    "next_step": "deploy_page"
                                },
                                "deploy_page": {
                                    "action": "git_commit",
                                    "config": {
                                        "repo_name": "sites",
                                        "domain_field": "rendered_page.domain",
                                        "content_field": "rendered_page.html",
                                        "filename_field": "rendered_page.filename",
                                        "commit_message": "Rerender: {{.filename}}"
                                    },
                                    "description": "Deploy rendered page to git",
                                    "output_field": "deploy_result",
                                    "next_step": "update_status"
                                },
                                "update_status": {
                                    "action": "update_page_status",
                                    "config": {
                                        "status": "deployed",
                                        "page_id_field": "current_page.page_id",
                                        "commit_from": "deploy_result.commit_sha"
                                    },
                                    "description": "Update page status in database",
                                    "output_field": "status_updated",
                                    "next_step": "loop_done"
                                },
                                "loop_done": {
                                    "action": "loop_complete",
                                    "description": "Page rerender complete"
                                }
                            }
                        }
                    },
                    "description": "Loop through pages, render and deploy each one",
                    "output_field": "pages_processed",
                    "next_step": "complete"
                },
                "complete": {
                    "action": "complete_workflow",
                    "config": {
                        "output_fields": ["rerender_pages", "pages_processed"]
                    },
                    "description": "Rerender complete"
                }
            }
        },
        "processing_mode": "orchestrator",
        "timeout_seconds": 1800
    }'::jsonb,
    1
    WHERE NOT EXISTS (SELECT 1 FROM agent_definitions WHERE type = 'rerender-pages');

-- ============================================================
-- 3. Verify the update
-- ============================================================
SELECT type, display_name,
       default_config->'workflow'->>'start_step' as start_step,
    version
FROM agent_definitions WHERE type = 'rerender-pages';

-- much the same:

-- Update rerender-pages agent to use loop-based workflow
-- This processes one page at a time, avoiding large message payloads
--
-- Flow: get_pages → loop → (render_page → deploy_page → update_status) → complete

-- ============================================================
-- 1. Update rerender-pages agent workflow
-- ============================================================
UPDATE agent_definitions
SET default_config = '{
    "workflow": {
        "start_step": "get_pages",
        "steps": {
            "get_pages": {
                "action": "get_pages_for_rerender",
                "config": {
                    "site_id_field": "input_data.site_id",
                    "domain_field": "input_data.domain",
                    "include_statuses": ["deployed", "active"]
                },
                "description": "Get page metadata for rerender loop",
                "output_field": "rerender_pages",
                "next_step": "deploy_loop"
            },
            "deploy_loop": {
                "action": "loop",
                "config": {
                    "items_field": "rerender_pages.pages",
                    "item_variable": "current_page",
                    "mode": "sequential",
                    "max_iterations": 50,
                    "sub_workflow": {
                        "start_step": "render_page",
                        "steps": {
                            "render_page": {
                                "action": "rerender_single_page",
                                "config": {
                                    "page_id_field": "current_page.page_id",
                                    "site_id_field": "rerender_pages.site_id",
                                    "domain_field": "rerender_pages.domain",
                                    "max_nav_items": 6
                                },
                                "description": "Render single page from stored sections",
                                "output_field": "rendered_page",
                                "next_step": "deploy_page"
                            },
                            "deploy_page": {
                                "action": "git_commit",
                                "config": {
                                    "repo_name": "sites",
                                    "domain_field": "rendered_page.domain",
                                    "content_field": "rendered_page.html",
                                    "filename_field": "rendered_page.filename",
                                    "commit_message": "Rerender: {{.filename}}"
                                },
                                "description": "Deploy rendered page to git",
                                "output_field": "deploy_result",
                                "next_step": "update_status"
                            },
                            "update_status": {
                                "action": "update_page_status",
                                "config": {
                                    "status": "deployed",
                                    "page_id_field": "current_page.page_id",
                                    "commit_from": "deploy_result.commit_sha"
                                },
                                "description": "Update page status in database",
                                "output_field": "status_updated",
                                "next_step": "loop_done"
                            },
                            "loop_done": {
                                "action": "loop_complete",
                                "description": "Page rerender complete"
                            }
                        }
                    }
                },
                "description": "Loop through pages, render and deploy each one",
                "output_field": "pages_processed",
                "next_step": "complete"
            },
            "complete": {
                "action": "complete_workflow",
                "config": {
                    "output_fields": ["rerender_pages", "pages_processed"]
                },
                "description": "Rerender complete"
            }
        }
    },
    "processing_mode": "orchestrator",
    "timeout_seconds": 1800
}'::jsonb,
version = version + 1,
updated_at = NOW()
WHERE type = 'rerender-pages';

-- ============================================================
-- 2. If agent doesn't exist, create it
-- ============================================================
INSERT INTO agent_definitions (type, display_name, description, category, default_config, version)
SELECT
    'rerender-pages',
    'Rerender Pages Agent',
    'Re-renders all deployed pages with current components (header, footer, nav, contact info)',
    'specialist',
    '{
        "workflow": {
            "start_step": "get_pages",
            "steps": {
                "get_pages": {
                    "action": "get_pages_for_rerender",
                    "config": {
                        "site_id_field": "input_data.site_id",
                        "domain_field": "input_data.domain",
                        "include_statuses": ["deployed", "active"]
                    },
                    "description": "Get page metadata for rerender loop",
                    "output_field": "rerender_pages",
                    "next_step": "deploy_loop"
                },
                "deploy_loop": {
                    "action": "loop",
                    "config": {
                        "items_field": "rerender_pages.pages",
                        "item_variable": "current_page",
                        "mode": "sequential",
                        "max_iterations": 50,
                        "sub_workflow": {
                            "start_step": "render_page",
                            "steps": {
                                "render_page": {
                                    "action": "rerender_single_page",
                                    "config": {
                                        "page_id_field": "current_page.page_id",
                                        "site_id_field": "rerender_pages.site_id",
                                        "domain_field": "rerender_pages.domain",
                                        "max_nav_items": 6
                                    },
                                    "description": "Render single page from stored sections",
                                    "output_field": "rendered_page",
                                    "next_step": "deploy_page"
                                },
                                "deploy_page": {
                                    "action": "git_commit",
                                    "config": {
                                        "repo_name": "sites",
                                        "domain_field": "rendered_page.domain",
                                        "content_field": "rendered_page.html",
                                        "filename_field": "rendered_page.filename",
                                        "commit_message": "Rerender: {{.filename}}"
                                    },
                                    "description": "Deploy rendered page to git",
                                    "output_field": "deploy_result",
                                    "next_step": "update_status"
                                },
                                "update_status": {
                                    "action": "update_page_status",
                                    "config": {
                                        "status": "deployed",
                                        "page_id_field": "current_page.page_id",
                                        "commit_from": "deploy_result.commit_sha"
                                    },
                                    "description": "Update page status in database",
                                    "output_field": "status_updated",
                                    "next_step": "loop_done"
                                },
                                "loop_done": {
                                    "action": "loop_complete",
                                    "description": "Page rerender complete"
                                }
                            }
                        }
                    },
                    "description": "Loop through pages, render and deploy each one",
                    "output_field": "pages_processed",
                    "next_step": "complete"
                },
                "complete": {
                    "action": "complete_workflow",
                    "config": {
                        "output_fields": ["rerender_pages", "pages_processed"]
                    },
                    "description": "Rerender complete"
                }
            }
        },
        "processing_mode": "orchestrator",
        "timeout_seconds": 1800
    }'::jsonb,
    1
    WHERE NOT EXISTS (SELECT 1 FROM agent_definitions WHERE type = 'rerender-pages');

-- ============================================================
-- 3. Verify the update
-- ============================================================
SELECT
    type,
    display_name,
    default_config->'workflow'->>'start_step' as start_step,
    default_config->'workflow'->'steps'->'deploy_loop'->'config'->'sub_workflow'->'steps'->'render_page'->>'action' as render_action,
    version
FROM agent_definitions
WHERE type = 'rerender-pages';


-- use input_fields
--
-- Update rerender-pages agent to use loop-based workflow
-- This processes one page at a time, avoiding large message payloads
--
-- Flow: get_pages → loop → (render_page → deploy_page → update_status) → complete

-- ============================================================
-- 1. Update rerender-pages agent workflow
-- ============================================================
UPDATE agent_definitions
SET default_config = '{
    "workflow": {
        "start_step": "get_pages",
        "steps": {
            "get_pages": {
                "action": "get_pages_for_rerender",
                "config": {
                    "input_fields": ["site_id", "domain"],
                    "include_statuses": ["deployed", "active"]
                },
                "description": "Get page metadata for rerender loop",
                "output_field": "rerender_pages",
                "next_step": "deploy_loop"
            },
            "deploy_loop": {
                "action": "loop",
                "config": {
                    "items_field": "rerender_pages.pages",
                    "item_variable": "current_page",
                    "mode": "sequential",
                    "max_iterations": 50,
                    "sub_workflow": {
                        "start_step": "render_page",
                        "steps": {
                            "render_page": {
                                "action": "rerender_single_page",
                                "config": {
                                    "page_id_field": "current_page.page_id",
                                    "site_id_field": "rerender_pages.site_id",
                                    "domain_field": "rerender_pages.domain",
                                    "max_nav_items": 6
                                },
                                "description": "Render single page from stored sections",
                                "output_field": "rendered_page",
                                "next_step": "deploy_page"
                            },
                            "deploy_page": {
                                "action": "git_commit",
                                "config": {
                                    "repo_name": "sites",
                                    "domain_field": "rendered_page.domain",
                                    "content_field": "rendered_page.html",
                                    "filename_field": "rendered_page.filename",
                                    "commit_message": "Rerender: {{.filename}}"
                                },
                                "description": "Deploy rendered page to git",
                                "output_field": "deploy_result",
                                "next_step": "update_status"
                            },
                            "update_status": {
                                "action": "update_page_status",
                                "config": {
                                    "status": "deployed",
                                    "page_id_field": "current_page.page_id",
                                    "commit_from": "deploy_result.commit_sha"
                                },
                                "description": "Update page status in database",
                                "output_field": "status_updated",
                                "next_step": "loop_done"
                            },
                            "loop_done": {
                                "action": "loop_complete",
                                "description": "Page rerender complete"
                            }
                        }
                    }
                },
                "description": "Loop through pages, render and deploy each one",
                "output_field": "pages_processed",
                "next_step": "complete"
            },
            "complete": {
                "action": "complete_workflow",
                "config": {
                    "output_fields": ["rerender_pages", "pages_processed"]
                },
                "description": "Rerender complete"
            }
        }
    },
    "processing_mode": "orchestrator",
    "timeout_seconds": 1800
}'::jsonb,
version = version + 1,
updated_at = NOW()
WHERE type = 'rerender-pages';

-- ============================================================
-- 2. If agent doesn't exist, create it
-- ============================================================
INSERT INTO agent_definitions (type, display_name, description, category, default_config, version)
SELECT
    'rerender-pages',
    'Rerender Pages Agent',
    'Re-renders all deployed pages with current components (header, footer, nav, contact info)',
    'specialist',
    '{
        "workflow": {
            "start_step": "get_pages",
            "steps": {
                "get_pages": {
                    "action": "get_pages_for_rerender",
                    "config": {
                        "input_fields": ["site_id", "domain"],
                        "include_statuses": ["deployed", "active"]
                    },
                    "description": "Get page metadata for rerender loop",
                    "output_field": "rerender_pages",
                    "next_step": "deploy_loop"
                },
                "deploy_loop": {
                    "action": "loop",
                    "config": {
                        "items_field": "rerender_pages.pages",
                        "item_variable": "current_page",
                        "mode": "sequential",
                        "max_iterations": 50,
                        "sub_workflow": {
                            "start_step": "render_page",
                            "steps": {
                                "render_page": {
                                    "action": "rerender_single_page",
                                    "config": {
                                        "page_id_field": "current_page.page_id",
                                        "site_id_field": "rerender_pages.site_id",
                                        "domain_field": "rerender_pages.domain",
                                        "max_nav_items": 6
                                    },
                                    "description": "Render single page from stored sections",
                                    "output_field": "rendered_page",
                                    "next_step": "deploy_page"
                                },
                                "deploy_page": {
                                    "action": "git_commit",
                                    "config": {
                                        "repo_name": "sites",
                                        "domain_field": "rendered_page.domain",
                                        "content_field": "rendered_page.html",
                                        "filename_field": "rendered_page.filename",
                                        "commit_message": "Rerender: {{.filename}}"
                                    },
                                    "description": "Deploy rendered page to git",
                                    "output_field": "deploy_result",
                                    "next_step": "update_status"
                                },
                                "update_status": {
                                    "action": "update_page_status",
                                    "config": {
                                        "status": "deployed",
                                        "page_id_field": "current_page.page_id",
                                        "commit_from": "deploy_result.commit_sha"
                                    },
                                    "description": "Update page status in database",
                                    "output_field": "status_updated",
                                    "next_step": "loop_done"
                                },
                                "loop_done": {
                                    "action": "loop_complete",
                                    "description": "Page rerender complete"
                                }
                            }
                        }
                    },
                    "description": "Loop through pages, render and deploy each one",
                    "output_field": "pages_processed",
                    "next_step": "complete"
                },
                "complete": {
                    "action": "complete_workflow",
                    "config": {
                        "output_fields": ["rerender_pages", "pages_processed"]
                    },
                    "description": "Rerender complete"
                }
            }
        },
        "processing_mode": "orchestrator",
        "timeout_seconds": 1800
    }'::jsonb,
    1
    WHERE NOT EXISTS (SELECT 1 FROM agent_definitions WHERE type = 'rerender-pages');

-- ============================================================
-- 3. Verify the update
-- ============================================================
SELECT
    type,
    display_name,
    default_config->'workflow'->>'start_step' as start_step,
    default_config->'workflow'->'steps'->'deploy_loop'->'config'->'sub_workflow'->'steps'->'render_page'->>'action' as render_action,
    version
FROM agent_definitions
WHERE type = 'rerender-pages';


---

-- Update rerender-pages agent to use loop-based workflow
-- This processes one page at a time, avoiding large message payloads
--
-- Flow: get_pages → loop → (render_page → deploy_page → update_status) → complete

-- ============================================================
-- 1. Update rerender-pages agent workflow
-- ============================================================
UPDATE agent_definitions
SET default_config = '{
    "workflow": {
        "start_step": "get_pages",
        "steps": {
            "get_pages": {
                "action": "get_pages_for_rerender",
                "config": {
                    "input_fields": ["site_id", "domain"],
                    "include_statuses": ["deployed", "active"]
                },
                "description": "Get page metadata for rerender loop",
                "output_field": "rerender_pages",
                "next_step": "deploy_loop"
            },
            "deploy_loop": {
                "action": "loop",
                "config": {
                    "items_field": "rerender_pages.pages",
                    "item_variable": "current_page",
                    "mode": "sequential",
                    "max_iterations": 50,
                    "sub_workflow": {
                        "start_step": "render_page",
                        "steps": {
                            "render_page": {
                                "action": "rerender_single_page",
                                "config": {
                                    "input_fields": ["current_page", "rerender_pages"],
                                    "max_nav_items": 6
                                },
                                "description": "Render single page from stored sections",
                                "output_field": "rendered_page",
                                "next_step": "check_skipped"
                            },
                            "check_skipped": {
                                "action": "conditional",
                                "config": {
                                    "condition": "rendered_page.skipped == true",
                                    "then_step": "loop_done",
                                    "else_step": "deploy_page"
                                },
                                "description": "Skip deploy if page has no stored sections"
                            },
                            "deploy_page": {
                                "action": "git_commit",
                                "config": {
                                    "repo_name": "sites",
                                    "domain_field": "rendered_page.domain",
                                    "content_field": "rendered_page.html",
                                    "filename_field": "rendered_page.filename",
                                    "commit_message": "Rerender: {{.filename}}"
                                },
                                "description": "Deploy rendered page to git",
                                "output_field": "deploy_result",
                                "next_step": "update_status"
                            },
                            "update_status": {
                                "action": "update_page_status",
                                "config": {
                                    "status": "deployed",
                                    "page_id_field": "current_page.page_id",
                                    "commit_from": "deploy_result.commit_sha"
                                },
                                "description": "Update page status in database",
                                "output_field": "status_updated",
                                "next_step": "loop_done"
                            },
                            "loop_done": {
                                "action": "loop_complete",
                                "description": "Page rerender complete"
                            }
                        }
                    }
                },
                "description": "Loop through pages, render and deploy each one",
                "output_field": "pages_processed",
                "next_step": "complete"
            },
            "complete": {
                "action": "complete_workflow",
                "config": {
                    "output_fields": ["rerender_pages", "pages_processed"]
                },
                "description": "Rerender complete"
            }
        }
    },
    "processing_mode": "orchestrator",
    "timeout_seconds": 1800
}'::jsonb,
version = version + 1,
updated_at = NOW()
WHERE type = 'rerender-pages';

-- ============================================================
-- 2. If agent doesn't exist, create it
-- ============================================================
INSERT INTO agent_definitions (type, display_name, description, category, default_config, version)
SELECT
    'rerender-pages',
    'Rerender Pages Agent',
    'Re-renders all deployed pages with current components (header, footer, nav, contact info)',
    'specialist',
    '{
        "workflow": {
            "start_step": "get_pages",
            "steps": {
                "get_pages": {
                    "action": "get_pages_for_rerender",
                    "config": {
                        "input_fields": ["site_id", "domain"],
                        "include_statuses": ["deployed", "active"]
                    },
                    "description": "Get page metadata for rerender loop",
                    "output_field": "rerender_pages",
                    "next_step": "deploy_loop"
                },
                "deploy_loop": {
                    "action": "loop",
                    "config": {
                        "items_field": "rerender_pages.pages",
                        "item_variable": "current_page",
                        "mode": "sequential",
                        "max_iterations": 50,
                        "sub_workflow": {
                            "start_step": "render_page",
                            "steps": {
                                "render_page": {
                                    "action": "rerender_single_page",
                                    "config": {
                                        "input_fields": ["current_page", "rerender_pages"],
                                        "max_nav_items": 6
                                    },
                                    "description": "Render single page from stored sections",
                                    "output_field": "rendered_page",
                                    "next_step": "check_skipped"
                                },
                                "check_skipped": {
                                    "action": "conditional",
                                    "config": {
                                        "condition": "rendered_page.skipped == true",
                                        "then_step": "loop_done",
                                        "else_step": "deploy_page"
                                    },
                                    "description": "Skip deploy if page has no stored sections"
                                },
                                "deploy_page": {
                                    "action": "git_commit",
                                    "config": {
                                        "repo_name": "sites",
                                        "domain_field": "rendered_page.domain",
                                        "content_field": "rendered_page.html",
                                        "filename_field": "rendered_page.filename",
                                        "commit_message": "Rerender: {{.filename}}"
                                    },
                                    "description": "Deploy rendered page to git",
                                    "output_field": "deploy_result",
                                    "next_step": "update_status"
                                },
                                "update_status": {
                                    "action": "update_page_status",
                                    "config": {
                                        "status": "deployed",
                                        "page_id_field": "current_page.page_id",
                                        "commit_from": "deploy_result.commit_sha"
                                    },
                                    "description": "Update page status in database",
                                    "output_field": "status_updated",
                                    "next_step": "loop_done"
                                },
                                "loop_done": {
                                    "action": "loop_complete",
                                    "description": "Page rerender complete"
                                }
                            }
                        }
                    },
                    "description": "Loop through pages, render and deploy each one",
                    "output_field": "pages_processed",
                    "next_step": "complete"
                },
                "complete": {
                    "action": "complete_workflow",
                    "config": {
                        "output_fields": ["rerender_pages", "pages_processed"]
                    },
                    "description": "Rerender complete"
                }
            }
        },
        "processing_mode": "orchestrator",
        "timeout_seconds": 1800
    }'::jsonb,
    1
    WHERE NOT EXISTS (SELECT 1 FROM agent_definitions WHERE type = 'rerender-pages');

-- ============================================================
-- 3. Verify the update
-- ============================================================
SELECT
    type,
    display_name,
    default_config->'workflow'->>'start_step' as start_step,
    default_config->'workflow'->'steps'->'deploy_loop'->'config'->'sub_workflow'->'steps'->'render_page'->>'action' as render_action,
    version
FROM agent_definitions
WHERE type = 'rerender-pages';

-- add page rerender agent to remove sub-workflow


-- ============================================================
-- 2. Update rerender-pages to spawn and call page-rerender agent
-- ============================================================
UPDATE agent_definitions
SET default_config = '{
    "workflow": {
        "start_step": "get_pages",
        "steps": {
            "get_pages": {
                "action": "get_pages_for_rerender",
                "config": {
                    "input_fields": ["site_id", "domain"],
                    "include_statuses": ["deployed", "active"]
                },
                "description": "Get page metadata for rerender",
                "output_field": "rerender_pages",
                "next_step": "check_pages_exist"
            },
            "check_pages_exist": {
                "action": "conditional",
                "config": {
                    "condition": "rerender_pages.page_count > 0",
                    "then_step": "spawn_page_agent",
                    "else_step": "complete"
                },
                "description": "Skip if no pages to process"
            },
            "spawn_page_agent": {
                "action": "spawn_agent",
                "config": {
                    "role": "page_renderer",
                    "agent_type": "page-rerender"
                },
                "description": "Spawn page rerender agent",
                "output_field": "page_agent",
                "next_step": "deploy_loop"
            },
            "deploy_loop": {
                "action": "loop",
                "config": {
                    "mode": "sequential",
                    "items_field": "rerender_pages.pages",
                    "item_variable": "current_page",
                    "max_iterations": 50,
                    "sub_workflow": {
                        "start_step": "call_page_agent",
                        "steps": {
                            "call_page_agent": {
                                "action": "call_agent",
                                "config": {
                                    "agent_type": "page-rerender",
                                    "target_role": "page_renderer",
                                    "input_mapping": {
                                        "page_id": "current_page.page_id",
                                        "site_id": "rerender_pages.site_id",
                                        "domain": "rerender_pages.domain"
                                    },
                                    "timeout_seconds": 120
                                },
                                "description": "Call page-rerender agent for this page",
                                "output_field": "page_result",
                                "next_step": "loop_done"
                            },
                            "loop_done": {
                                "action": "loop_complete",
                                "description": "Page iteration complete"
                            }
                        }
                    }
                },
                "description": "Loop through pages, call agent for each",
                "output_field": "pages_processed",
                "next_step": "complete"
            },
            "complete": {
                "action": "complete_workflow",
                "config": {
                    "output_fields": ["rerender_pages", "pages_processed"]
                },
                "description": "Rerender complete"
            }
        }
    },
    "processing_mode": "orchestrator",
    "timeout_seconds": 1800
}'::jsonb,
version = version + 1,
updated_at = NOW()
WHERE type = 'rerender-pages';

-- ============================================================
-- 3. Verify both agents
-- ============================================================
SELECT
    type,
    display_name,
    default_config->'workflow'->>'start_step' as start_step,
    jsonb_object_keys(default_config->'workflow'->'steps') as steps,
    version,
    status
FROM agent_definitions
WHERE type IN ('rerender-pages', 'page-rerender')

--

-- Fix rerender-pages condition to use boolean instead of numeric comparison
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,check_pages_exist,config,condition}',
        '"rerender_pages.has_pages == true"'
                     ),
    updated_at = NOW()
WHERE type = 'rerender-pages';

-- Verify
SELECT type,
       default_config->'workflow'->'steps'->'check_pages_exist'->'config'->'condition' as condition
FROM agent_definitions
WHERE type = 'rerender-pages';


