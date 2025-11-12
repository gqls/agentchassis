# Timeout Configuration - Quick Reference and Examples

## Quick Reference

### Default Timeouts
- **Approval Actions**: 24 hours (86400 seconds)
- **Other Actions**: 3 minutes (180 seconds)

### Timeout Limits
- **Minimum**: 60 seconds (1 minute)
- **Maximum**: 604800 seconds (7 days)

### Timeout Specification
```json
{
  "action": "await_approval",
  "config": {
    "timeout_seconds": 86400
  }
}
```

## Example Workflows

### 1. Simple Content Approval (24 hours)

```json
{
  "start_step": "generate_content",
  "steps": {
    "generate_content": {
      "action": "execute_llm_prompt",
      "config": {
        "prompt_template": "Write a blog post about {{.topic}}"
      },
      "next_step": "await_approval"
    },
    "await_approval": {
      "action": "await_approval",
      "config": {
        "timeout_seconds": 86400,
        "approval_type": "content_review",
        "approval_fields": ["generate_content"],
        "ui_config": {
          "title": "Content Approval Required",
          "description": "Please review the generated content within 24 hours"
        }
      },
      "next_step": "process_decision"
    },
    "process_decision": {
      "action": "process_approval_decision",
      "config": {
        "stop_on_reject": true
      },
      "next_step": "publish"
    },
    "publish": {
      "action": "publish_content",
      "next_step": "done"
    },
    "done": {
      "action": "complete_workflow"
    }
  }
}
```

### 2. Long-Form Article Review (48 hours)

```json
{
  "start_step": "research",
  "steps": {
    "research": {
      "action": "call_agent",
      "config": {
        "agent_type": "researcher",
        "target_role": "content_researcher"
      },
      "next_step": "write_article"
    },
    "write_article": {
      "action": "execute_llm_prompt",
      "config": {
        "prompt_template": "Write a detailed article using research: {{.research}}"
      },
      "next_step": "editorial_review"
    },
    "editorial_review": {
      "action": "await_approval",
      "config": {
        "timeout_seconds": 172800,
        "approval_type": "editorial_review",
        "approval_fields": ["write_article"],
        "ui_config": {
          "title": "Editorial Review Required",
          "description": "Please review this long-form article within 48 hours",
          "priority": "medium"
        }
      },
      "next_step": "process_editorial_decision"
    },
    "process_editorial_decision": {
      "action": "process_approval_decision",
      "config": {
        "stop_on_reject": false
      },
      "next_step": "revise_or_publish"
    },
    "revise_or_publish": {
      "action": "conditional_branch",
      "config": {
        "condition": "{{.approved}}",
        "if_true": "publish",
        "if_false": "revision_feedback"
      }
    },
    "revision_feedback": {
      "action": "send_notification",
      "config": {
        "message": "Article requires revision: {{.comments}}"
      },
      "next_step": "done"
    },
    "publish": {
      "action": "publish_content",
      "next_step": "done"
    },
    "done": {
      "action": "complete_workflow"
    }
  }
}
```

### 3. Legal Document Approval (7 days maximum)

```json
{
  "start_step": "draft_document",
  "steps": {
    "draft_document": {
      "action": "execute_llm_prompt",
      "config": {
        "prompt_template": "Draft legal document for: {{.case_details}}"
      },
      "next_step": "legal_review"
    },
    "legal_review": {
      "action": "await_approval",
      "config": {
        "timeout_seconds": 604800,
        "approval_type": "legal_review",
        "approval_fields": ["draft_document"],
        "ui_config": {
          "title": "Legal Review Required",
          "description": "Comprehensive legal review - up to 7 days allowed",
          "priority": "high",
          "reviewers": ["legal-team@company.com"]
        }
      },
      "next_step": "process_legal_decision"
    },
    "process_legal_decision": {
      "action": "process_approval_decision",
      "config": {
        "stop_on_reject": true
      },
      "next_step": "finalize"
    },
    "finalize": {
      "action": "finalize_document",
      "next_step": "done"
    },
    "done": {
      "action": "complete_workflow"
    }
  }
}
```

### 4. Multi-Stage Approval Pipeline

```json
{
  "start_step": "create_design",
  "steps": {
    "create_design": {
      "action": "execute_llm_prompt",
      "config": {
        "prompt_template": "Create design for: {{.project_spec}}"
      },
      "next_step": "initial_review"
    },
    "initial_review": {
      "action": "await_approval",
      "config": {
        "timeout_seconds": 14400,
        "approval_type": "design_review",
        "ui_config": {
          "title": "Initial Design Review",
          "description": "Quick design review - 4 hours"
        }
      },
      "next_step": "process_initial"
    },
    "process_initial": {
      "action": "process_approval_decision",
      "next_step": "refine_design"
    },
    "refine_design": {
      "action": "execute_llm_prompt",
      "config": {
        "prompt_template": "Refine design based on: {{.comments}}"
      },
      "next_step": "final_approval"
    },
    "final_approval": {
      "action": "await_approval",
      "config": {
        "timeout_seconds": 86400,
        "approval_type": "final_approval",
        "ui_config": {
          "title": "Final Design Approval",
          "description": "Final sign-off required - 24 hours"
        }
      },
      "next_step": "process_final"
    },
    "process_final": {
      "action": "process_approval_decision",
      "config": {
        "stop_on_reject": true
      },
      "next_step": "implement"
    },
    "implement": {
      "action": "implement_design",
      "next_step": "done"
    },
    "done": {
      "action": "complete_workflow"
    }
  }
}
```

## Common Timeout Scenarios

### Quick Approval (1 hour)
```json
"timeout_seconds": 3600
```
Use for: Urgent decisions, time-sensitive content

### Standard Business Day (8 hours)
```json
"timeout_seconds": 28800
```
Use for: Same-day review expectations

### Next Business Day (24 hours)
```json
"timeout_seconds": 86400
```
Use for: Standard review processes

### Extended Review (2-3 days)
```json
"timeout_seconds": 172800  // 2 days
"timeout_seconds": 259200  // 3 days
```
Use for: Complex content, multiple reviewers

### Weekly Review (7 days)
```json
"timeout_seconds": 604800
```
Use for: Legal, compliance, or executive reviews

## Timeout Behavior

### Before Timeout
- Workflow waits in `StatusAwaitingResponses`
- Approval request stored in database
- UI shows approval pending with countdown
- Orchestration state persisted in database

### At Timeout
- `handleRequestTimeout` is triggered
- Attempts retry (up to 3 times)
- After max retries, workflow fails with timeout error
- Error includes: request_id, step_name, timeout duration

### Approval Received Before Timeout
- Normal workflow continuation
- Approval response processed
- Next step executed immediately

### Pod Restart During Wait
- State recovered from database
- Timeout goroutine restarted
- Wait continues from TimeoutAt timestamp

## Logging Examples

### Successful Conversion
```
INFO  Converted timeout_seconds to step.Timeout
  action: "await_approval"
  timeout_seconds: 86400
  step_timeout: "24h0m0s"
```

### Timeout Below Minimum
```
WARN  Approval timeout below minimum, using minimum
  requested: 30
  minimum: 60
```

### Timeout Above Maximum
```
WARN  Approval timeout exceeds maximum, using maximum
  requested: 1000000
  maximum: 604800
```

### Approval Wait Started
```
INFO  Action requires waiting for response
  request_id: "9178e309-db08-42e6-bdc7-bc7e1b700fa3"
  target_agent_type: "human"
  responses_topic: "system.agent.generic.responses"
  timeout_formatted: "1.0 days"
```

## Troubleshooting

### Timeout Not Respected
**Issue**: Approval times out after 3 minutes despite config
**Solution**: Ensure `ConvertStepTimeout` is called in coordinator

### Timeout Too Long
**Issue**: Timeouts set beyond 7 days
**Solution**: System enforces 7-day maximum automatically

### Timeout Too Short
**Issue**: Timeouts set below 1 minute
**Solution**: System enforces 60-second minimum automatically

### State Loss on Restart
**Issue**: Timeout lost when pod restarts
**Solution**: Implement timeout recovery in agent startup

## Best Practices

1. **Choose Appropriate Timeouts**
    - Consider reviewer availability
    - Account for time zones
    - Add buffer for complexity

2. **Use UI Config Effectively**
    - Set clear expectations in description
    - Indicate urgency with priority
    - Specify exact reviewers when known

3. **Handle Rejections Gracefully**
    - Set `stop_on_reject` appropriately
    - Provide revision paths
    - Log rejection reasons

4. **Monitor Timeout Rates**
    - Track how often approvals timeout
    - Adjust timeout durations based on data
    - Alert on unusual timeout patterns

5. **Test Multi-Day Waits**
    - Verify pod restart recovery
    - Test database state persistence
    - Confirm timeout handler triggers correctly