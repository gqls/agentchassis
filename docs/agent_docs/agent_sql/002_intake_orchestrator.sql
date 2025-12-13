{
  "workflow": {
    "start_step": "fetch_available_builders",
    "steps": {
      "fetch_available_builders": {
        "action": "query_agent_definitions",
        "description": "Discover what builder agents are available",
        "config": {
          "fields": ["type", "display_name", "description"],
          "filter": {
            "type_pattern": "%-builder"
          }
        },
        "next_step": "spawn_classifier",
        "output_field": "available_builders"
      },

      "spawn_classifier": {
        "action": "spawn_agent",
        "description": "Spawn site classifier agent",
        "config": {
          "role": "classifier",
          "agent_type": "site-classifier"
        },
        "next_step": "spawn_briefer"
      },

      "spawn_briefer": {
        "action": "spawn_agent",
        "description": "Spawn briefing agent",
        "config": {
          "role": "briefer",
          "agent_type": "briefing-agent"
        },
        "next_step": "call_classifier"
      },

      "call_classifier": {
        "action": "call_agent",
        "description": "Classify the site type from domain and objective",
        "config": {
          "agent_type": "site-classifier",
          "target_role": "classifier",
          "input_fields": [
            "input_data",
            "available_builders"
          ],
          "timeout_seconds": 30
        },
        "next_step": "hitl_confirm_type",
        "output_field": "classification"
      },

      "hitl_confirm_type": {
        "action": "request_human_input",
        "description": "Human confirms or adjusts the site type classification",
        "config": {
          "title": "Confirm Site Type",
          "request_type": "confirmation",
          "message": "Please confirm or adjust the site classification",
          "skip_if": "input_data.hitl_mode == auto",
          "timeout_seconds": 86400,
          "fields": [
            {
              "name": "site_type",
              "type": "select",
              "label": "Site Type",
              "options": ["landing", "content", "portfolio", "brochure"],
              "default_from": "classification.classify_site.result.site_type"
            },
            {
              "name": "recommended_builder",
              "type": "dynamic_select",
              "label": "Builder",
              "default_from": "classification.classify_site.result.recommended_builder",
              "options_from": "available_builders.agents",
              "option_label_field": "display_name",
              "option_value_field": "type"
            }
          ]
        },
        "next_step": "fetch_questionnaire",
        "output_field": "confirmed_type"
      },

      "fetch_questionnaire": {
        "action": "fetch_agent_questionnaire",
        "description": "Fetch the briefing questionnaire for the target builder",
        "config": {
          "agent_type_field": "confirmed_type.recommended_builder"
        },
        "next_step": "call_briefer",
        "output_field": "questionnaire"
      },

      "call_briefer": {
        "action": "call_agent",
        "description": "Run the briefing questionnaire",
        "config": {
          "agent_type": "briefing-agent",
          "target_role": "briefer",
          "input_fields": [
            "input_data",
            "classification",
            "confirmed_type",
            "questionnaire"
          ],
          "timeout_seconds": 120
        },
        "next_step": "hitl_review_brief",
        "output_field": "brief_data"
      },

      "hitl_review_brief": {
        "action": "request_human_input",
        "description": "Human reviews the completed brief",
        "config": {
          "title": "Review Brief",
          "request_type": "review",
          "message": "Please review and adjust the briefing answers if needed",
          "data_field": "brief_data",
          "editable": true,
          "skip_if": "input_data.hitl_mode == auto",
          "timeout_seconds": 86400
        },
        "next_step": "spawn_builder",
        "output_field": "reviewed_brief"
      },

      "spawn_builder": {
        "action": "spawn_agent",
        "description": "Spawn the appropriate builder agent with all collected data",
        "config": {
          "role": "builder",
          "agent_type_field": "confirmed_type.recommended_builder",
          "input_fields": [
            "input_data",
            "classification",
            "brief_data",
            "reviewed_brief"
          ]
        },
        "next_step": "call_builder",
        "output_field": "spawned_builder"
      },

      "call_builder": {
        "action": "call_agent",
        "description": "Call the spawned builder to execute the site build",
        "config": {
          "target_role": "builder",
          "input_fields": [
            "input_data",
            "classification",
            "confirmed_type",
            "brief_data",
            "reviewed_brief",
            "questionnaire"
          ],
          "timeout_seconds": 600
        },
        "next_step": "complete",
        "output_field": "build_result"
      },

      "complete": {
        "action": "complete_workflow",
        "description": "Intake complete - builder has been spawned"
      }
    }
  },

  "processing_mode": "orchestration",
  "timeout_seconds": 600
}
