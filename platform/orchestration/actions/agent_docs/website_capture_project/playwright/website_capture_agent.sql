-- Website Capture Agent Definition
-- Handles screenshot capture and DOM extraction using Playwright

INSERT INTO agent_definitions (
    id,
    type,
    display_name,
    description,
    category,
    default_config,
    is_active,
    capabilities,
    image_repository,
    image_tag,
    orchestration_workflow
) VALUES (
             gen_random_uuid(),
             'website-capture',
             'Website Capture Agent',
             'Captures screenshots, HTML, CSS, and interaction states from websites',
             'data-extraction',
             '{
                 "workflow": {
                     "start_step": "prepare_capture",
                     "steps": {
                         "prepare_capture": {
                             "action": "validate_url",
                             "description": "Validate and prepare URL for capture",
                             "config": {
                                 "url_field": "target_url",
                                 "add_protocol_if_missing": true
                             },
                             "next_step": "capture_desktop"
                         },
                         "capture_desktop": {
                             "action": "capture_site",
                             "description": "Capture desktop version of website",
                             "config": {
                                 "adapter_type": "playwright",
                                 "capture_config": {
                                     "viewport": {"width": 1920, "height": 1080},
                                     "full_page": true,
                                     "wait_until": "networkidle",
                                     "capture_dom": true,
                                     "capture_styles": true,
                                     "extract_computed_styles": true
                                 }
                             },
                             "next_step": "capture_mobile"
                         },
                         "capture_mobile": {
                             "action": "capture_site",
                             "description": "Capture mobile version of website",
                             "config": {
                                 "adapter_type": "playwright",
                                 "capture_config": {
                                     "viewport": {"width": 390, "height": 844},
                                     "full_page": true,
                                     "user_agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X)",
                                     "wait_until": "networkidle"
                                 }
                             },
                             "next_step": "capture_interactions"
                         },
                         "capture_interactions": {
                             "action": "capture_hover_states",
                             "description": "Capture hover and interaction states",
                             "config": {
                                 "adapter_type": "playwright",
                                 "capture_config": {
                                     "selectors": ["a", "button", "[role=\"button\"]", ".interactive"],
                                     "capture_hover": true,
                                     "capture_focus": true,
                                     "max_elements": 50
                                 }
                             },
                             "next_step": "capture_scroll_behavior"
                         },
                         "capture_scroll_behavior": {
                             "action": "capture_scroll_animation",
                             "description": "Capture scroll animations and parallax effects",
                             "config": {
                                 "adapter_type": "playwright",
                                 "capture_config": {
                                     "scroll_intervals": [0, 25, 50, 75, 100],
                                     "capture_at_each": true,
                                     "detect_parallax": true,
                                     "detect_sticky_elements": true
                                 }
                             },
                             "next_step": "extract_assets"
                         },
                         "extract_assets": {
                             "action": "extract_website_assets",
                             "description": "Extract images, fonts, and other assets",
                             "config": {
                                 "extract_images": true,
                                 "extract_fonts": true,
                                 "extract_icons": true,
                                 "compress_assets": true
                             },
                             "next_step": "upload_to_storage"
                         },
                         "upload_to_storage": {
                             "action": "upload_to_s3",
                             "description": "Upload captured data to S3",
                             "config": {
                                 "bucket": "website-captures",
                                 "organize_by": "domain_and_timestamp",
                                 "generate_manifest": true
                             },
                             "next_step": "complete"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "description": "Return capture results with S3 paths"
                         }
                     }
                 },
                 "processing_mode": "task",
                 "adapter_topics": {
                     "playwright": "system.adapter.playwright.requests"
                 },
                 "timeout_seconds": 180,
                 "retry_config": {
                     "max_retries": 3,
                     "backoff_multiplier": 2
                 }
             }'::jsonb,
             true,
             ARRAY['capture', 'playwright', 'scraping', 'screenshot'],
             'docker.io/aqls/agent-chassis',
             'v1.0.407',
             NULL
         );