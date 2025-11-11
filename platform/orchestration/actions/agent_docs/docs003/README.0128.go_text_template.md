tmpl, err := template.New("condition").Parse(conditionTemplate)

Example Workflow Conditions Using text/template

{
"steps": {
"check_simple": {
"action": "evaluate_condition",
"config": {
"condition": "{{.input_data.extract_structured}}",
"default": false
},
"next_step": {
"true": "do_extraction",
"false": "do_basic_scrape"
}
},

    "check_multiple": {
      "action": "evaluate_condition",
      "config": {
        "condition": "{{and .input_data.crawl_pages (gt .input_data.crawl_limit 0)}}",
        "default": false
      },
      "next_step": {
        "true": "execute_crawl",
        "false": "skip_crawl"
      }
    },
    
    "check_previous_result": {
      "action": "evaluate_condition",
      "config": {
        "condition": "{{.execute_scrape.success}}",
        "default": false
      },
      "next_step": {
        "true": "process_results",
        "false": "handle_error"
      }
    },
    
    "check_or_condition": {
      "action": "evaluate_condition",
      "config": {
        "condition": "{{or .input_data.force_crawl (gt .execute_scrape.page_count 5)}}",
        "default": false
      },
      "next_step": {
        "true": "deep_crawl",
        "false": "complete"
      }
    }
}
}



Template Functions Available
Since you're using text/template, you have access to these built-in functions:

and - Returns true if all arguments are true
or - Returns true if any argument is true
not - Negates a boolean
eq - Equal comparison
ne - Not equal
lt - Less than
le - Less than or equal
gt - Greater than
ge - Greater than or equal

Complex Condition Examples for Your Webscrape Workflow

--
test this later;
-- Update your agent group with more sophisticated conditions
UPDATE agent_group_definitions
SET orchestration_workflow = '{
"start_step": "analyze_request",
"steps": {
"analyze_request": {
"action": "evaluate_condition",
"description": "Check if structured extraction is needed",
"config": {
"condition": "{{or .input_data.extract_structured (eq .input_data.mode \"full\")}}",
"default": false
},
"next_step": {
"true": "spawn_extractor",
"false": "check_simple_scrape"
}
},

    "check_simple_scrape": {
      "action": "evaluate_condition",
      "description": "Verify URL is provided",
      "config": {
        "condition": "{{and .input_data.target_url (ne .input_data.target_url \"\")}}",
        "default": false
      },
      "next_step": {
        "true": "spawn_basic_scraper",
        "false": "return_error"
      }
    },
    
    "check_crawl_needed": {
      "action": "evaluate_condition",
      "description": "Complex crawl decision",
      "config": {
        "condition": "{{and .input_data.crawl_pages (or (gt .input_data.crawl_limit 1) .execute_basic_scrape.found_pagination)}}",
        "default": false
      },
      "next_step": {
        "true": "execute_crawl",
        "false": "aggregate_results"
      }
    }
}
}'::jsonb
WHERE group_type = 'website-analyzer';
