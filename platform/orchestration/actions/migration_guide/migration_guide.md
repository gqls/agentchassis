# Actions Package Refactor - Migration Guide

## Overview

This refactor reorganizes the actions from a single flat package into categorized subpackages with a central registry that includes metadata (category, status, description, domain tags).

## New Directory Structure

```
platform/orchestration/actions/
├── registry/           # Central registry with ActionDefinition type
│   └── registry.go
├── all/                # Import-only package to register all actions
│   └── all.go
├── control/            # Workflow control: complete, await, branch
│   └── control.go
├── agent/              # Agent management: spawn, call, discover
│   └── agent.go
├── llm/                # LLM operations: execute_llm_prompt
│   └── llm.go
├── data/               # Data manipulation: validate, transform, aggregate
│   └── data.go
├── io/                 # External I/O: git, http, notifications
│   └── io.go
├── scrape/             # Web scraping: firecrawl_*, scrape_web
│   └── scrape.go
├── storage/            # Storage: S3, hosting, assets
│   └── storage.go
├── html/               # HTML operations: generate, assemble, validate
│   └── html.go
├── image/              # Image generation
│   └── image.go
├── memory/             # Memory and caching
│   └── memory.go
├── hitl/               # Human-in-the-loop approvals
│   └── hitl.go
└── planning/           # Planning and evaluation
    └── planning.go
```

## Action Categories

| Category   | Description                                    | Example Actions                          |
|------------|------------------------------------------------|------------------------------------------|
| `control`  | Workflow control                               | complete_workflow, conditional_branch    |
| `agent`    | Agent lifecycle management                     | spawn_agent, call_agent, discover_agents |
| `llm`      | AI/LLM operations                              | execute_llm_prompt                       |
| `data`     | Data manipulation                              | validate_input, transform_data, calculate|
| `io`       | External I/O                                   | git_commit, http_request                 |
| `scrape`   | Web scraping                                   | scrape_web, firecrawl_scrape             |
| `storage`  | Storage operations                             | upload_to_s3, deploy_to_hosting          |
| `html`     | HTML generation/manipulation                   | generate_html, assemble_from_library     |
| `image`    | Image generation                               | generate_image                           |
| `memory`   | Memory and caching                             | store_memory, cache_lookup               |
| `hitl`     | Human-in-the-loop                              | await_approval, process_approval_decision|
| `planning` | Planning and evaluation                        | plan_agent_team, evaluate_task           |

## Migration Steps

### Step 1: Create the new package structure

Copy the files from `actions_refactor/` to `platform/orchestration/actions/`:

```bash
# From project root
cp -r actions_refactor/* platform/orchestration/actions/
```

### Step 2: Migrate implementations

Each package has stub implementations. Migrate the actual code from the old files:

| Old File                      | New Package | Notes                              |
|-------------------------------|-------------|------------------------------------|
| workflow_actions.go           | control/    | complete_workflow, await_response  |
| spawn_actions.go              | agent/      | spawn_agent, spawn_group           |
| call_agent.go                 | agent/      | call_agent                         |
| discovery_actions.go          | agent/      | discover_agents                    |
| ai_actions.go                 | llm/        | execute_llm_prompt                 |
| basic_actions.go              | data/       | validate_input, transform_data     |
| aggregate_data.go             | data/       | aggregate_data                     |
| aggregate_webpage.go          | data/       | aggregate_webpage                  |
| calculate_actions.go          | data/       | calculate                          |
| git_deployer_actions.go       | io/         | git_commit                         |
| generic_actions.go            | io/         | http_request, send_notification    |
| webscrape_actions.go          | scrape/     | All firecrawl and scrape actions   |
| storage_actions.go            | storage/    | upload_to_s3, store_result         |
| html_actions.go               | html/       | generate_html, process_html        |
| assemble_from_library.go      | html/       | assemble_from_library              |
| generate_image_actions.go     | image/      | generate_image                     |
| hitl_actions.go               | hitl/       | All approval actions               |

### Step 3: Update coordinator to use new registry

In `platform/orchestration/coordinator.go`, update:

```go
// OLD
import "github.com/aqls/personae/platform/orchestration/actions"

func getActionHandler(action string) (actions.ActionFunc, error) {
    fn, exists := actions.GlobalActionRegistry[action]
    if !exists {
        return nil, fmt.Errorf("unknown action: %s", action)
    }
    return fn, nil
}

// NEW
import (
    "github.com/aqls/personae/platform/orchestration/actions/registry"
    _ "github.com/aqls/personae/platform/orchestration/actions/all" // Register all actions
)

func getActionHandler(action string) (registry.ActionFunc, error) {
    def, exists := registry.Get(action)
    if !exists {
        return nil, fmt.Errorf("unknown action: %s", action)
    }
    
    // Log warning for deprecated actions
    if def.Status == registry.StatusDeprecated {
        logger.Warn("Using deprecated action", zap.String("action", action))
    }
    
    return def.Func, nil
}
```

### Step 4: Update ActionParams

The `registry.ActionParams` type may need to be adjusted. Either:
1. Use the one defined in `registry/registry.go`
2. Keep using a shared type from a common package

### Step 5: Update imports in main.go

```go
import (
    // This triggers all action registrations
    _ "github.com/aqls/personae/platform/orchestration/actions/all"
)
```

### Step 6: Delete old files

Once everything is migrated and tested, remove the old action files from the flat structure.

## New Registry Features

### Query actions by category

```go
// Get all LLM actions
llmActions := registry.ListByCategory(registry.CategoryLLM)

// Get all active actions
activeActions := registry.ListByStatus(registry.StatusActive)

// Get registry summary
summary := registry.Summary()
// Returns: map[string]int{"control": 4, "agent": 5, "llm": 1, ...}
```

### Check deprecation

```go
if err := registry.Validate("spawn_agent_k8s"); err != nil {
    // Action is deprecated or doesn't exist
}

// Or just warn
registry.WarnIfDeprecated("spawn_agent_k8s", logger)
```

### Get action metadata

```go
def, ok := registry.Get("execute_llm_prompt")
if ok {
    fmt.Printf("Category: %s\n", def.Category)
    fmt.Printf("Status: %s\n", def.Status)
    fmt.Printf("Description: %s\n", def.Description)
}
```

## API Endpoint for Action Discovery

You can add an API endpoint to list available actions:

```go
func handleListActions(w http.ResponseWriter, r *http.Request) {
    actions := registry.GetAll()
    
    type ActionInfo struct {
        Name        string   `json:"name"`
        Category    string   `json:"category"`
        Status      string   `json:"status"`
        Description string   `json:"description"`
        DomainTags  []string `json:"domain_tags,omitempty"`
    }
    
    result := make([]ActionInfo, 0, len(actions))
    for name, def := range actions {
        result = append(result, ActionInfo{
            Name:        name,
            Category:    def.Category,
            Status:      def.Status,
            Description: def.Description,
            DomainTags:  def.DomainTags,
        })
    }
    
    json.NewEncoder(w).Encode(result)
}
```

## Testing

After migration, verify:

1. All actions are registered:
   ```go
   actions := registry.List()
   // Should contain all expected action names
   ```

2. Actions execute correctly:
   ```bash
   # Test mvp-site-builder workflow
   curl -X POST http://localhost:8080/api/v1/orchestrate \
     -d '{"action": "orchestrate", "config": {"group_type": "mvp-site-builder"}, ...}'
   ```

3. Deprecated actions log warnings (check logs)

## Rollback

If issues occur, the old `registry.go` with `GlobalActionRegistry` can be temporarily restored by:

1. Keeping the old registry.go renamed to registry_old.go
2. Updating imports to use it instead of the new package structure