-- 645_design_critique_agent.sql — the 018 design critic, Phase 1: a critic that REPORTS.
-- DB-only INSERT of a NEW agent type (idempotent via WHERE NOT EXISTS; no snapshot —
-- nothing pre-existing to restore). features_open/018 (owner-raised 2026-07-24);
-- owner-directed to BUILD 2026-08-25 in the leopardess lane; design is the approved
-- plan's Part B (~/.claude/plans/let-s-do-1-2-and-3-ancient-crab.md), which follows the
-- vigilant_designer_offer_analysis Phase 2 shape and its four recorded owner decisions
-- (manual cadence; designer before offer analyser; trial both models — Gemini first per
-- the 2026-07-24 owner call; broad autonomy). That lane was told in writing
-- (CONTRIB 2026-08-25 in their dir, commit 7ce1bb6c5).
--
-- ORDERING: the Go half is ALREADY LIVE. `design-critique-agent` entered
-- isStorageEnabledAgent in commit 04c49f8f0 (Council-Submitted 30d5fdde), and the
-- running chassis was capability-probed 2026-08-26 with both controls:
--   grep -aq 'design-critique-agent' /proc/1/exe  -> PRESENT   (the grant)
--   grep -aq 'tool-acceptance-agent' /proc/1/exe  -> PRESENT   (positive control)
--   grep -aq 'zzz_invented_string_xq9' /proc/1/exe -> absent   (negative control)
-- So the 243 ordering trap (seed before build => every run fails 'no storage client')
-- does NOT apply at the time of writing. If this file is ever re-applied to a fresh
-- database, re-probe before the first run. Every action named below is registered and
-- live (ensure_site_record, query_database, request_render_audit, execute_vision_prompt,
-- append_doc_note, write_render_audit_findings, complete_workflow).
--
-- THE STRUCTURAL SPLIT (the design principle the whole Part B argues for):
-- auto-filed work items draw ONLY from the browser's deterministic measurements —
-- `file_measured_findings` is write_render_audit_findings over the SAME render_audit
-- response, the proven drain (seed 301) with the proven dedup keys, and it runs BEFORE
-- the vision call so a failed critique can never cost the measured findings. The vision
-- model's output lands ONLY in a doc_notes report (declared reader: the owner, plus
-- load_doc_context for any later agent). No prompt edit can leak a taste claim into the
-- auto-file path — moving it would be a workflow change, which is reviewable.
-- write_render_audit_findings makes this agent a SECOND producer on contrast_failure /
-- undeployed_asset keys beside render-audit-agent's rotation — CO-DEDUP on identical
-- keys is that drain's designed-for pattern (its own header: "the shared namespace is
-- the point"), so the duplication is absorbed, not multiplied.
--
-- WHAT THIS DELIBERATELY DOES NOT DO (Phase 1 scope, all stated in the approved plan):
--   * No compose-time role; no external reference-site capture; no new item_type
--     (needs_design_review's fifth producer needs the deferred semantics ruling first);
--   * NO cadence — manual trigger only, per the recorded owner decision. Nothing emits
--     work for this agent; nothing promotes into it; the only path is a hand dispatch:
--       ./docs/leopardessconsulting/scripts/orchestrate_safe.sh design-critique-agent \
--         '{"site_id":"…","domain":"…"}'
--   * No council seat; no due-gate (an unbuilt check_critique_due is not simulated —
--     a manual-only agent needs none).
--
-- LOAD_DESIGN_CONTEXT reads the RIGHT things — palettes.colours via the proven join
-- sites → style_collections → css_themes → palettes (verified live for leopardess
-- 2026-08-26: leopardess-dark-gold, 22 colour keys), plus site_specs.design_intent,
-- which nothing else reads. It deliberately does NOT read
-- style_collections.color_palette, which holds seed defaults for every forked
-- collection and is precisely why visual-design-auditor called a dark-gold site
-- "corporate blue" (the approved plan's §"A fixable bug in the existing auditor").
--
-- COST: 8 pages x 2 viewports = 16 images, exactly execute_vision_prompt's max_images
-- default ceiling — set EXPLICITLY below so excess drops are a config choice, not an
-- inherited accident. Vision calls log NULL tokens (known bug, plan §cost envelope), so
-- the report's header line carries the image count for at least countable spend.

BEGIN;

INSERT INTO agent_definitions
  (type, display_name, description, category, agent_category, status,
   input_contract, output_contract, default_config)
SELECT
  'design-critique-agent',
  'Design Critique Agent (018 Phase 1)',
  'The specialist design critic (features_open/018): captures rendered screenshots of up to 8 pages at two viewports via the render audit, files the audit''s deterministic measurements through the proven drain, then asks a vision model for a taste critique — hierarchy, rhythm, whitespace, card composition, imagery cohesion — judged against the site''s own design_intent and live palette. The critique is a doc_notes report for a person; it files nothing. Manual trigger only.',
  'design',
  'analyst',
  'active',
  '{"required": ["site_id", "domain"], "optional": ["spec", "page_names"]}'::jsonb,
  '{"design_report": {"note_id": "uuid", "pages_seen": "int", "measured_findings_filed": "int"}}'::jsonb,
  $config$
  {
    "workflow": {
      "start_step": "ensure_site_record",
      "processing_mode": "orchestrator",
      "timeout_seconds": 900,
      "steps": {
        "ensure_site_record": {
          "action": "ensure_site_record",
          "config": { "store_brief_in_content_data": false },
          "next_step": "audit",
          "description": "Load the site record",
          "output_field": "site_record"
        },
        "audit": {
          "action": "request_render_audit",
          "config": {
            "site_id_field": "site_record.site_id",
            "domain_field": "site_record.domain",
            "max_pages": 8,
            "capture_renders": true,
            "error_step": "complete_error"
          },
          "next_step": "file_measured_findings",
          "description": "One browser sweep serves both halves: deterministic contrast/overflow/broken-image measurements AND full-page renders at two viewports. AWAITS the adapter.",
          "output_field": "render_audit"
        },
        "file_measured_findings": {
          "action": "write_render_audit_findings",
          "config": {
            "site_id": "site_record.site_id",
            "error_step": "complete_error"
          },
          "next_step": "load_design_context",
          "description": "The auto-file half, and the ONLY auto-file half: the browser's own measurements through the proven seed-301 drain (contrast_failure -> css-patch-agent, attributed broken images -> undeployed_asset -> asset-deployer), born detected. Runs BEFORE the vision call so a failed critique never costs the measured findings.",
          "output_field": "findings_written"
        },
        "load_design_context": {
          "action": "query_database",
          "config": {
            "query": "SELECT jsonb_build_object('design_intent', (SELECT ss.data FROM site_specs ss WHERE ss.site_id = $1::uuid AND ss.aspect = 'design_intent' AND ss.is_current), 'palette_name', p2.name, 'palette', p2.colours, 'style_collection', sc.name) AS design_context FROM sites s JOIN style_collections sc ON sc.id = s.style_collection_id JOIN css_themes ct ON ct.id = sc.css_theme_id JOIN palettes p2 ON p2.id = ct.palette_id WHERE s.id = $1::uuid",
            "params": ["site_record.site_id"],
            "output_format": "array",
            "error_step": "complete_no_critique"
          },
          "next_step": "critique",
          "description": "The context the existing auditor gets wrong: the LIVE palette via sites -> style_collections -> css_themes -> palettes (never style_collections.color_palette, which holds seed defaults for forked collections), plus site_specs.design_intent, which nothing else reads.",
          "output_field": "design_context"
        },
        "critique": {
          "action": "execute_vision_prompt",
          "config": {
            "images_field": "render_audit",
            "max_images": 16,
            "output_type": "text",
            "error_step": "complete_no_critique",
            "ai_service": {
              "provider": "gemini",
              "model": "gemini-pro-latest",
              "api_key_env_var": "GEMINI_API_KEY",
              "max_tokens": 6000
            },
            "prompt_template": "You are a senior designer reviewing a junior designer's build of a website. You are looking at full-page screenshots of its pages, captured at two viewport widths. Automated checks have already measured contrast, overflow and broken images — those are handled elsewhere and are not your job. Your job is taste: the things only an experienced eye catches.\n\nImages, in order:\n{{ .vision_image_manifest }}\n\nThe site's own declared design intent and live palette (judge the pages against THIS, not against generic taste — the client chose this direction deliberately):\n{{ .design_context.results }}\n\nAssess, in plain prose, page by page where it matters and site-wide where the point is shared:\n- Visual hierarchy: does each page have one clear focal point, and does the eye travel in the intended order?\n- Rhythm and whitespace: density, breathing room, whether sections repeat one visual beat monotonously.\n- Card and grid composition: balance, orphaned last rows, cards restating their neighbour's point visually.\n- Imagery: cohesion of treatment across pages, sameness (many pages sharing one identical image is a defect of distinctiveness), and whether images explain something or merely decorate.\n- Distinctiveness: would a visitor who saw two of these pages recognise the third as the same site — and would they mistake this site for a template?\n\nEvery finding must be concrete enough to act on mechanically: name the page (by image index), the region, the property, and the direction of change — 'the services cards: tighten vertical padding by roughly a third and left-align the titles' is usable; 'feels dated' is noise and is worth nothing.\n\nDo NOT report any of the following. Each is a known artefact of how these pictures are taken, not a fault in the page:\n- The navigation bar appearing part-way down the image rather than at the top: these are full-page captures and a sticky header paints where it was scrolled to.\n- The image being extremely tall, or content looking small relative to the image: a full-page capture is not a viewport view.\n- Any page that looks blank, half-populated, or mid-interaction: say INCONCLUSIVE for that image and move on.\n\nStructure the report: one short site-wide paragraph first (at most three points), then per-page notes (at most six points across all pages), then — only if genuinely warranted — a short list of detail-level nits (at most three). If a page is genuinely well composed, say so in one line; finding little is a perfectly good answer and is much better than inventing something marginal.\n\nBegin the report with one line stating how many images you were shown and at which widths, so the reader knows what this critique could and could not see.\n\nWrite plain prose for a person to read. No JSON, no code fences, no headings beyond simple line breaks, no preamble."
          },
          "next_step": "write_report",
          "description": "The taste half. Judged against the site's own design_intent + live palette. Output is prose for a person; it can file nothing, by construction.",
          "output_field": "critique"
        },
        "write_report": {
          "action": "append_doc_note",
          "config": {
            "subject_type": "pipeline",
            "subject_key": "design-critique",
            "note_body_field": "critique.result",
            "note_categories": ["design-report"],
            "note_source": "design-critique-agent",
            "created_by": "design-critique-agent",
            "error_step": "complete_no_critique"
          },
          "next_step": "complete",
          "description": "Persist the critique under its own category (design-report, per DOC-068 per-site verdicts belong in notes; note_site_id_field defaults to input_data.site_id). Declared reader: the owner, plus load_doc_context for any later agent.",
          "output_field": "design_report_note"
        },
        "complete_no_critique": {
          "action": "complete_workflow",
          "config": {
            "success_message": "Render audit completed and measured findings filed; the vision critique did not produce a report",
            "multiple_output_fields": ["findings_written"]
          },
          "description": "SUCCESS terminal, not an error: the deterministic half stands on its own (the 317 lesson — a failed look is not a failed run). complete vs complete_no_critique on the row says which happened without reading __step_error."
        },
        "complete_error": {
          "action": "complete_workflow",
          "config": {
            "success_message": "Design critique run failed before the audit produced usable results",
            "error": true
          },
          "description": "The audit itself failed or its findings could not be filed — a failed audit and a clean audit must never be read the same way."
        },
        "complete": {
          "action": "complete_workflow",
          "config": {
            "success_message": "Design critique complete: measured findings filed, taste report written",
            "multiple_output_fields": ["findings_written", "design_report_note"]
          },
          "description": "Both halves done."
        }
      }
    }
  }
  $config$::jsonb
WHERE NOT EXISTS (
  SELECT 1 FROM agent_definitions
  WHERE type = 'design-critique-agent'
    AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
);

-- verification that can stop the COMMIT (RFC_006: bare SELECTs cannot)
DO $$
DECLARE n int; steps int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='design-critique-agent' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION 'design-critique-agent rows: % (want exactly 1)', n;
  END IF;
  SELECT count(*) INTO steps FROM agent_definitions,
    LATERAL jsonb_object_keys(default_config->'workflow'->'steps') k
   WHERE type='design-critique-agent' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF steps <> 8 THEN
    RAISE EXCEPTION 'design-critique-agent workflow has % steps (want 8)', steps;
  END IF;
END $$;

COMMIT;
