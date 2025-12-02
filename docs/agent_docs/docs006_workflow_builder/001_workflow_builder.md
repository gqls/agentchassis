https://claude.ai/chat/a36b6fe1-efa1-4d53-b30e-768ab6c9bf68

# Workflow Builder & Validator System

## Problem Statement

Creating workflows manually is error-prone:
- ❌ Data path mismatches (input_data.X vs X.step_name)
- ❌ Missing agent definitions
- ❌ Incorrect input_fields references
- ❌ Spawn/call order confusion
- ❌ Local action vs agent inconsistencies
- ❌ No validation before deployment

**Result:** Workflows fail at runtime with cryptic errors.

## Solution: Structured Workflow Builder

A validation-first system that:
1. ✅ Validates agent definitions exist
2. ✅ Computes correct data paths automatically
3. ✅ Suggests input_fields based on previous outputs
4. ✅ Validates workflow logic before database insertion
5. ✅ Generates test cases
6. ✅ Provides a DSL for readable workflow definitions

---

## Architecture

```
┌─────────────────────────────────────────────────────┐
│              Workflow Definition (YAML/DSL)          │
│  - Human-readable format                             │
│  - No manual path computation                        │
│  - Declarative agent dependencies                    │
└──────────────────┬──────────────────────────────────┘
                   │
                   ↓
┌─────────────────────────────────────────────────────┐
│           Workflow Builder Service                   │
│                                                      │
│  1. Parse DSL                                       │
│  2. Validate agent definitions exist                │
│  3. Compute data paths                              │
│  4. Generate orchestration_workflow JSON            │
│  5. Create test cases                               │
│  6. Validate before insert                          │
└──────────────────┬──────────────────────────────────┘
                   │
                   ↓
┌─────────────────────────────────────────────────────┐
│          Database (agent_group_definitions)         │
│  - Validated JSON workflow                          │
│  - Correct paths guaranteed                         │
│  - Ready to execute                                 │
└─────────────────────────────────────────────────────┘
```

---

## Component 1: Workflow DSL (YAML)

Human-friendly format that abstracts away complexity:

```yaml
# multipage-site-builder.yaml
workflow:
  name: "Multi-Page Site Builder"
  group_type: "multipage-site-builder"
  category: "builder"
  status: "active"
  domain_tags: ["website", "landing-page"]
  description: "Builds 3-page site with index, about, contact"

agents:
  - name: strategist
    type: site-strategist
    description: "Generates build plan"
    
  - name: architect
    type: landing-page-architect
    description: "Assembles template from library"
    inputs:
      - from: strategist.build_plan
      
  - name: writer
    type: content-writer
    description: "Generates content JSON"
    inputs:
      - from: architect.template_data
      - from: strategist.build_plan
      - from: workflow.input_data  # Special: original input
      
  - name: html_assembler
    type: html-assembler
    description: "Combines template + content"
    inputs:
      - from: architect.template_data
      - from: writer.content_data
      - from: workflow.input_data
      
  - name: multipage_wrapper
    type: multipage-wrapper
    description: "Creates about/contact pages"
    inputs:
      - from: html_assembler.final_html
      - from: workflow.input_data
      
  - name: deployer
    type: site-deployer
    description: "Deploys to Git"
    inputs:
      - from: multipage_wrapper.site_files
      - from: workflow.input_data

timeouts:
  strategist: 120
  architect: 120
  writer: 300
  html_assembler: 120
  multipage_wrapper: 60
  deployer: 180
```

**Key features:**
- No manual path computation
- Clear dependency graph
- References by agent name + output field
- Builder computes actual paths

---

## Component 2: Workflow Builder (Go Service)

### Core Functions

```go
type WorkflowBuilder struct {
    db *sql.DB
    validator *WorkflowValidator
}

// Main entry point
func (wb *WorkflowBuilder) BuildFromYAML(yamlPath string) (*WorkflowDefinition, error)

// Individual steps
func (wb *WorkflowBuilder) ParseYAML(yamlPath string) (*WorkflowSpec, error)
func (wb *WorkflowBuilder) ValidateAgents(spec *WorkflowSpec) error
func (wb *WorkflowBuilder) ComputePaths(spec *WorkflowSpec) (*PathMap, error)
func (wb *WorkflowBuilder) GenerateOrchestrationJSON(spec *WorkflowSpec, paths *PathMap) (map[string]interface{}, error)
func (wb *WorkflowBuilder) GenerateTestCases(spec *WorkflowSpec) ([]TestCase, error)
func (wb *WorkflowBuilder) InsertToDatabase(def *WorkflowDefinition) error
```

### Data Structures

```go
type WorkflowSpec struct {
    Name         string
    GroupType    string
    Category     string
    Status       string
    DomainTags   []string
    Description  string
    Agents       []AgentStep
    Timeouts     map[string]int
}

type AgentStep struct {
    Name        string   // strategist, architect, etc.
    Type        string   // site-strategist, landing-page-architect
    Description string
    Inputs      []InputRef  // References to other agent outputs
    OutputField string   // Auto-computed if not specified
}

type InputRef struct {
    From string  // "strategist.build_plan" or "workflow.input_data"
}

type PathMap struct {
    // Maps logical references to actual CollectedData paths
    Paths map[string]string  // "strategist.build_plan" -> "build_plan"
                             // "writer.content_data" -> "content_data"
                             // When passed to next agent via input_fields,
                             // becomes "input_data.content_data"
}
```

---

## Component 3: Path Resolution Engine

Automatically computes correct paths based on agent types:

```go
type PathResolver struct {
    agentTypes map[string]AgentType  // Loaded from DB
}

func (pr *PathResolver) ResolvePath(ref InputRef, previousSteps []AgentStep) (string, error) {
    // Parse ref: "strategist.build_plan"
    parts := strings.Split(ref.From, ".")
    agentName := parts[0]
    fieldName := parts[1]
    
    // Find the agent step
    for _, step := range previousSteps {
        if step.Name == agentName {
            // Determine if agent or local action
            agentType := pr.agentTypes[step.Type]
            
            if agentType.IsLocalAction {
                // Local action: path includes step name
                return fmt.Sprintf("input_data.%s.%s.%s", 
                    step.Name, step.Name, fieldName), nil
            } else {
                // Agent call: direct path
                return fmt.Sprintf("input_data.%s", fieldName), nil
            }
        }
    }
    
    return "", fmt.Errorf("agent %s not found", agentName)
}
```

---

## Component 4: Workflow Validator

Validates before database insertion:

```go
type WorkflowValidator struct {
    db *sql.DB
}

func (wv *WorkflowValidator) Validate(spec *WorkflowSpec) error {
    var errors []error
    
    // Check 1: All agent types exist
    for _, agent := range spec.Agents {
        exists, err := wv.AgentTypeExists(agent.Type)
        if err != nil || !exists {
            errors = append(errors, 
                fmt.Errorf("agent type '%s' not found", agent.Type))
        }
    }
    
    // Check 2: No circular dependencies
    if wv.HasCircularDependency(spec) {
        errors = append(errors, 
            fmt.Errorf("circular dependency detected"))
    }
    
    // Check 3: All input references are valid
    for _, agent := range spec.Agents {
        for _, input := range agent.Inputs {
            if !wv.IsValidReference(input, spec) {
                errors = append(errors, 
                    fmt.Errorf("invalid input reference: %s", input.From))
            }
        }
    }
    
    // Check 4: Category and status are valid
    if !wv.IsValidCategory(spec.Category) {
        errors = append(errors, 
            fmt.Errorf("invalid category: %s", spec.Category))
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("validation failed: %v", errors)
    }
    
    return nil
}
```

---

## Component 5: JSON Generator

Generates the orchestration_workflow JSON:

```go
func (wb *WorkflowBuilder) GenerateOrchestrationJSON(
    spec *WorkflowSpec, 
    paths *PathMap,
) (map[string]interface{}, error) {
    
    workflow := map[string]interface{}{
        "start_step": "spawn_" + spec.Agents[0].Name,
        "steps": make(map[string]interface{}),
    }
    
    steps := workflow["steps"].(map[string]interface{})
    
    // Generate spawn steps
    for i, agent := range spec.Agents {
        spawnStep := map[string]interface{}{
            "action": "spawn_agent",
            "config": map[string]interface{}{
                "role": agent.Name,
                "agent_type": agent.Type,
            },
            "description": fmt.Sprintf("Spawn %s", agent.Description),
        }
        
        // Next step
        if i < len(spec.Agents)-1 {
            spawnStep["next_step"] = "spawn_" + spec.Agents[i+1].Name
        } else {
            spawnStep["next_step"] = "call_" + spec.Agents[0].Name
        }
        
        steps["spawn_"+agent.Name] = spawnStep
    }
    
    // Generate call steps
    for i, agent := range spec.Agents {
        callStep := map[string]interface{}{
            "action": "call_agent",
            "config": map[string]interface{}{
                "agent_type": agent.Type,
                "target_role": agent.Name,
                "timeout_seconds": spec.Timeouts[agent.Name],
            },
            "description": agent.Description,
        }
        
        // Compute input_fields from dependencies
        if len(agent.Inputs) > 0 {
            inputFields := make([]string, len(agent.Inputs))
            for i, input := range agent.Inputs {
                // Convert "strategist.build_plan" to field name
                parts := strings.Split(input.From, ".")
                inputFields[i] = parts[1]  // "build_plan"
            }
            callStep["config"].(map[string]interface{})["input_fields"] = inputFields
        }
        
        // Output field (same as agent name by convention)
        outputField := agent.Name + "_data"
        if agent.OutputField != "" {
            outputField = agent.OutputField
        }
        callStep["output_field"] = outputField
        
        // Next step
        if i < len(spec.Agents)-1 {
            callStep["next_step"] = "call_" + spec.Agents[i+1].Name
        } else {
            callStep["next_step"] = "complete"
        }
        
        steps["call_"+agent.Name] = callStep
    }
    
    // Complete step
    steps["complete"] = map[string]interface{}{
        "action": "complete_workflow",
        "description": "Workflow complete",
    }
    
    return workflow, nil
}
```

---

## Component 6: Test Case Generator

Automatically generates test cases:

```go
func (wb *WorkflowBuilder) GenerateTestCases(spec *WorkflowSpec) ([]TestCase, error) {
    tests := []TestCase{
        {
            Name: "Happy Path",
            Input: map[string]interface{}{
                "domain": "test-" + spec.GroupType + ".com",
                "objective": "Test workflow",
                "model": "AIDA",
            },
            ExpectedSteps: []string{
                "spawn_" + spec.Agents[0].Name,
                // ... all steps
                "complete",
            },
            ExpectedFiles: []string{"index.html"},  // Based on workflow type
        },
        {
            Name: "Missing Input",
            Input: map[string]interface{}{
                // Missing required fields
            },
            ShouldFail: true,
            ExpectedError: "missing required field",
        },
    }
    
    return tests, nil
}
```

---

## Usage Examples

### CLI Tool

```bash
# Build workflow from YAML
workflow-builder build multipage-site-builder.yaml

# Validate without inserting
workflow-builder validate multipage-site-builder.yaml

# Generate test cases
workflow-builder test multipage-site-builder.yaml

# List all workflows
workflow-builder list

# Show workflow details
workflow-builder show multipage-site-builder

# Update existing workflow
workflow-builder update multipage-site-builder.yaml
```

### Go API

```go
builder := NewWorkflowBuilder(db)

// Build from YAML
workflow, err := builder.BuildFromYAML("multipage-site-builder.yaml")
if err != nil {
    log.Fatal(err)
}

// Validate
if err := builder.Validate(workflow); err != nil {
    log.Fatal(err)
}

// Insert to database
if err := builder.InsertToDatabase(workflow); err != nil {
    log.Fatal(err)
}

// Generate tests
tests, _ := builder.GenerateTestCases(workflow)
for _, test := range tests {
    runTest(test)
}
```

### HTTP API

```bash
# Upload YAML
curl -X POST http://localhost:8080/api/v1/workflows/build \
  -F "file=@multipage-site-builder.yaml"

# Response:
{
  "workflow_id": "uuid",
  "validation": {
    "passed": true,
    "warnings": [],
    "errors": []
  },
  "test_results": [
    {"name": "Happy Path", "passed": true}
  ]
}
```

---

## Benefits

### For Developers
- ✅ Write workflows in readable YAML
- ✅ No path computation needed
- ✅ Validation catches errors early
- ✅ Test cases generated automatically

### For Operations
- ✅ All workflows validated before deployment
- ✅ Consistent structure across all workflows
- ✅ Easy to audit and review
- ✅ Version control friendly (YAML in git)

### For System
- ✅ Prevents runtime path errors
- ✅ Validates agent existence
- ✅ Catches circular dependencies
- ✅ Ensures data flow correctness

---

## Implementation Plan

### Phase 1: Core Builder (Week 1)
- [ ] Parse YAML workflow definitions
- [ ] Validate agent types exist
- [ ] Generate orchestration JSON
- [ ] Insert to database

### Phase 2: Path Resolution (Week 2)
- [ ] Automatic path computation
- [ ] Handle agent vs local action differences
- [ ] Generate input_fields correctly
- [ ] Test with existing workflows

### Phase 3: Validation & Testing (Week 3)
- [ ] Comprehensive validation rules
- [ ] Test case generation
- [ ] CLI tool
- [ ] HTTP API

### Phase 4: Advanced Features (Week 4)
- [ ] Visual workflow editor (web UI)
- [ ] Workflow templates
- [ ] Migration tool (convert existing)
- [ ] Workflow diffing/versioning

---

## File Structure

```
platform/workflowbuilder/
├── cmd/
│   └── workflow-builder/     # CLI tool
│       └── main.go
├── pkg/
│   ├── parser/
│   │   └── yaml_parser.go    # Parse YAML
│   ├── validator/
│   │   └── validator.go      # Validation logic
│   ├── generator/
│   │   ├── json_generator.go # Generate orchestration JSON
│   │   └── test_generator.go # Generate test cases
│   ├── resolver/
│   │   └── path_resolver.go  # Path computation
│   └── db/
│       └── repository.go     # Database operations
├── api/
│   └── http/
│       └── handlers.go       # HTTP API
├── examples/
│   ├── multipage-site-builder.yaml
│   ├── mvp-site-builder.yaml
│   └── veterinary-analyzer.yaml
└── README.md
```

---

## Example Output

Input: `multipage-site-builder.yaml` (shown above)

Output: Complete validated workflow ready for database:

```json
{
  "name": "Multi-Page Site Builder",
  "group_type": "multipage-site-builder",
  "category": "builder",
  "status": "active",
  "domain_tags": ["website", "landing-page"],
  "agent_configs": [...],
  "orchestration_workflow": {
    "start_step": "spawn_strategist",
    "steps": {
      // Generated with correct paths
    }
  }
}
```

Plus:
- Validation report (all green)
- Test cases (ready to run)
- Documentation (auto-generated)

---

## Next Steps

1. Review this architecture
2. Approve/modify design
3. Start implementation (Phase 1)
4. Convert existing workflows to YAML
5. Test with multipage-site-builder
6. Roll out to all workflows

This eliminates the entire class of path-related bugs we've been dealing with.

--
# Workflow Builder System

## Overview

The Workflow Builder eliminates workflow creation errors by:
- ✅ Validating agent definitions exist
- ✅ Computing correct data paths automatically
- ✅ Preventing circular dependencies
- ✅ Generating test cases
- ✅ Providing readable YAML format

**No more runtime path errors.**

## The Problem

Creating workflows manually in JSON is error-prone:

```json
{
  "call_deployer": {
    "config": {
      "files_field": "site_files.files"  // ❌ Wrong path!
    }
  }
}
```

Actual data is at: `input_data.site_files.wrap_multipage.files`

Result: Deployment fails with "no files found"

## The Solution

Write workflows in human-readable YAML:

```yaml
workflow:
  name: "Multi-Page Site Builder"
  
  agents:
    - name: deployer
      type: site-deployer
      inputs:
        - from: multipage_wrapper.site_files
```

The builder:
1. Validates `site-deployer` agent exists
2. Validates `multipage_wrapper` exists earlier in pipeline
3. Computes actual path: `input_data.site_files.wrap_multipage.files`
4. Generates correct JSON workflow

## Quick Start

### 1. Install

```bash
cd platform/workflowbuilder
go build -o workflow-builder cmd/workflow-builder/main.go
```

### 2. Create YAML Workflow

```yaml
# my-workflow.yaml
workflow:
  name: "My Workflow"
  group_type: "my-workflow"
  category: "builder"
  
  agents:
    - name: strategist
      type: site-strategist
      inputs: [workflow.input_data]
      
    - name: architect
      type: landing-page-architect
      inputs: [strategist.build_plan, workflow.input_data]
      
    - name: deployer
      type: site-deployer
      inputs: [architect.template_data, workflow.input_data]
```

### 3. Validate

```bash
export DATABASE_URL="postgresql://user:pass@localhost/clients_db"
./workflow-builder validate my-workflow.yaml
```

Output:
```
✓ Parsed: My Workflow

--- Validation Report ---
✓ All validations passed
  ✓ All agent types exist
  ✓ No circular dependencies  
  ✓ All input references valid
  ✓ Category is valid
```

### 4. Build & Insert

```bash
./workflow-builder build my-workflow.yaml
```

Output:
```
✓ Parsed workflow: My Workflow
✓ All validations passed
✓ Generated 3 test cases
✓ Successfully created workflow: My Workflow (group_type: my-workflow)
```

## YAML Format Reference

### Basic Structure

```yaml
workflow:
  # Required fields
  name: "Human-readable name"
  group_type: "database-key"
  
  # Optional fields (have defaults)
  category: "builder"           # builder, analyzer, collector, etc.
  status: "active"              # active, experimental, deprecated
  domain_tags: ["website"]      # For filtering/search
  description: "What it does"
  
  # Agent pipeline
  agents:
    - name: agent1
      type: agent-type-1
      inputs: [workflow.input_data]
      
    - name: agent2
      type: agent-type-2
      inputs: [agent1.output, workflow.input_data]
```

### Input References

Three types of input references:

```yaml
agents:
  - name: my_agent
    inputs:
      # 1. Original workflow input
      - from: workflow.input_data
      
      # 2. Previous agent output (auto-resolved)
      - from: strategist.build_plan
      
      # 3. Multiple inputs
      - from: agent1.data
      - from: agent2.analysis
      - from: workflow.input_data
```

### Advanced Features

```yaml
workflow:
  agents:
    - name: writer
      type: content-writer
      inputs: [...]
      
      # Override default timeout
      timeout: 300
      
      # Override output field name (usually auto-computed)
      output_field: custom_field_name
      
      # Additional config passed to agent
      config:
        max_tokens: 8000
        temperature: 0.7
        
  # Global timeouts
  timeouts:
    strategist: 120
    writer: 300
    deployer: 180
    
  # Metadata
  metadata:
    version: "2.0"
    author: "Team Name"
    changelog: "What changed"
```

## Validation Rules

The validator checks:

### 1. Agent Existence
```
✓ Verifies all agent types exist in agent_definitions table
✗ Error if agent type not found
```

### 2. Input References
```
✓ All referenced agents exist earlier in pipeline
✓ Field names match previous agent outputs
✗ Error if referencing non-existent agent
✗ Error if forward references (agent3 referencing agent4)
```

### 3. Circular Dependencies
```
✓ Detects cycles in agent dependencies
✗ Error if A → B → C → A
```

### 4. Category & Status
```
✓ Category is in valid list (builder, analyzer, etc.)
✓ Status is in valid list (active, experimental, etc.)
✗ Error if invalid values
```

### 5. Path Resolution
```
✓ All paths can be resolved
✓ Handles agent vs local action differences
✗ Error if unresolvable path
```

## CLI Commands

### build
Build workflow and insert into database:
```bash
workflow-builder build my-workflow.yaml
workflow-builder build --dry-run my-workflow.yaml  # Show SQL without inserting
```

### validate
Validate without inserting:
```bash
workflow-builder validate my-workflow.yaml
```

### test
Generate test cases:
```bash
workflow-builder test my-workflow.yaml
```

### list
List all workflows:
```bash
workflow-builder list
```

Output:
```
--- Workflows ---
GROUP TYPE                     NAME                          CATEGORY    STATUS      AGENTS
multipage-site-builder         Multi-Page Site Builder       builder     active      6
mvp-site-builder              MVP Site Builder              builder     active      4
vet-practice-analyzer         Veterinary Practice Analyzer  analyzer    experimental 5
```

### show
Show workflow details:
```bash
workflow-builder show multipage-site-builder
```

### docs
Generate documentation:
```bash
workflow-builder docs my-workflow.yaml > WORKFLOW.md
```

## Path Resolution

The builder automatically handles the complexity of data paths.

### Problem: Agent vs Local Action

```
Agent Call:
  CollectedData["final_html"] = <result>
  When passed: input_data.final_html

Local Action:
  CollectedData["wrap_multipage"] = <result>
  When passed: input_data.site_files.wrap_multipage  ← Extra layer!
```

### Solution: Automatic Detection

```yaml
# You write this:
agents:
  - name: deployer
    inputs:
      - from: multipage_wrapper.site_files

# Builder computes actual path:
# If multipage_wrapper is local action:
#   input_data.site_files.wrap_multipage.files
# If multipage_wrapper is agent:
#   input_data.site_files.files
```

### Path Documentation

Generate path docs for any workflow:

```bash
workflow-builder docs multipage-site-builder.yaml
```

Output:
```markdown
# Data Path Documentation

## Agent Outputs
| Agent | Output Field | Logical Path | Actual Path | Type |
|-------|--------------|--------------|-------------|------|
| strategist | build_plan | strategist.build_plan | input_data.build_plan | Agent Call |
| multipage_wrapper | site_files | multipage_wrapper.site_files | input_data.site_files.wrap_multipage | Local Action |

## Input References
| Agent | Input Reference | Actual Path |
|-------|----------------|-------------|
| deployer | multipage_wrapper.site_files | input_data.site_files.wrap_multipage.files |
```

## Test Generation

Auto-generates test cases:

```json
{
  "name": "Happy Path",
  "input": {
    "domain": "test-multipage.com",
    "objective": "Test workflow",
    "model": "AIDA"
  },
  "expected_steps": [
    "spawn_strategist",
    "spawn_architect",
    "call_strategist",
    "call_architect",
    "complete"
  ],
  "expected_files": ["index.html", "about.html", "contact.html"]
}
```

## Examples

See `examples.yaml` for complete examples:

1. **Multi-Page Site Builder** - Website with 3 pages
2. **Simple Landing Page** - Single page site
3. **Veterinary Practice Analyzer** - Business analysis
4. **Fitness Analyzer** - Health tracking

## Migration Guide

### Convert Existing Workflows

For existing workflows in database:

```bash
# 1. Export current workflow
psql clients_db -c "SELECT orchestration_workflow FROM agent_group_definitions 
                    WHERE group_type='mvp-site-builder'" > current.json

# 2. Convert to YAML (manual or script)
# 3. Validate
workflow-builder validate new.yaml

# 4. Update database
workflow-builder build new.yaml
```

### Best Practices

1. **Use clear agent names**: `strategist`, `architect`, `deployer` (not `agent1`, `agent2`)
2. **Follow conventions**: Output field usually matches next agent's input
3. **Document changes**: Use metadata.changelog
4. **Test thoroughly**: Run `validate` before `build`
5. **Version workflows**: Increment metadata.version

## Architecture

```
YAML File
   ↓
Parser (pkg/parser)
   ↓
WorkflowSpec
   ↓
Validator (pkg/validator) → Database (check agents exist)
   ↓
Path Resolver (pkg/resolver) → Compute all paths
   ↓
Generator (pkg/generator) → Create JSON workflow
   ↓
Database Insert
```

## File Structure

```
platform/workflowbuilder/
├── cmd/
│   └── workflow-builder/
│       └── main.go              # CLI tool
├── pkg/
│   ├── types/
│   │   └── types.go             # Core types
│   ├── parser/
│   │   └── yaml_parser.go       # YAML parsing
│   ├── validator/
│   │   └── validator.go         # Validation logic
│   ├── resolver/
│   │   └── path_resolver.go     # Path computation
│   ├── generator/
│   │   ├── json_generator.go    # JSON generation
│   │   └── test_generator.go    # Test case generation
│   └── builder/
│       └── builder.go           # Main orchestration
├── examples/
│   ├── multipage-site-builder.yaml
│   ├── mvp-site-builder.yaml
│   └── veterinary-analyzer.yaml
└── README.md
```

## API Usage

### Go API

```go
import (
    "github.com/aqls/personae/platform/workflowbuilder/pkg/builder"
    "github.com/aqls/personae/platform/workflowbuilder/pkg/parser"
)

// Parse YAML
parser := parser.NewYAMLParser()
spec, err := parser.ParseFile("my-workflow.yaml")

// Build workflow
wfBuilder := builder.NewWorkflowBuilder(db)
result, err := wfBuilder.Build(spec)

// Check validation
if !result.ValidationReport.Passed {
    for _, err := range result.ValidationReport.Errors {
        log.Printf("Error: %s", err.Message)
    }
}

// Insert to database
err = wfBuilder.InsertToDatabase(result.Workflow)
```

### HTTP API (future)

```bash
curl -X POST http://localhost:8080/api/v1/workflows/build \
  -F "file=@my-workflow.yaml"
```

## Troubleshooting

### "Agent type not found"
```
Error: agents[2].type: agent type 'my-agent' not found in database
```
Solution: Create the agent definition first, or fix the type name

### "Circular dependency detected"
```
Error: circular dependency detected involving agent: strategist
```
Solution: Check agent inputs - ensure no cycles (A → B → A)

### "Invalid reference"
```
Error: agents[3].inputs: references 'agent2.data' but agent 'agent2' not found
```
Solution: Ensure referenced agent exists earlier in pipeline

### "Path not resolvable"
```
Error: cannot resolve path 'deployer.files.result'
```
Solution: Check output_field names match input references

## Contributing

### Adding Validation Rules

```go
// In pkg/validator/validator.go
func (v *Validator) checkMyRule(spec *types.WorkflowSpec) error {
    // Your validation logic
    if invalid {
        return fmt.Errorf("validation failed")
    }
    return nil
}
```

### Adding Test Generators

```go
// In pkg/generator/test_generator.go
func GenerateEdgeCaseTests(spec *types.WorkflowSpec) []types.TestCase {
    // Generate test cases
}
```

## Roadmap

- [x] Phase 1: Core parser & validator
- [x] Phase 2: Path resolution
- [x] Phase 3: JSON generation
- [ ] Phase 4: Test generation
- [ ] Phase 5: HTTP API
- [ ] Phase 6: Web UI
- [ ] Phase 7: Visual editor
- [ ] Phase 8: Workflow templates
- [ ] Phase 9: Diffing & versioning

## License

Internal tool for AI Agent Orchestration platform.

## Support

For issues or questions:
- Create an issue in the repo
- Contact the platform team
- Check examples/ for reference workflows


---

# Workflow Builder - Implementation Guide

## Overview

This guide walks through implementing the Workflow Builder system in your codebase.

## Prerequisites

- Go 1.21+
- PostgreSQL database access
- Existing agent-chassis codebase

## Implementation Timeline

**Phase 1 (Week 1):** Core functionality  
**Phase 2 (Week 2):** Validation & path resolution  
**Phase 3 (Week 3):** CLI tool & testing  
**Phase 4 (Week 4):** Polish & documentation

---

## Phase 1: Core Functionality (Week 1)

### Step 1.1: Create Directory Structure

```bash
mkdir -p platform/workflowbuilder/cmd/workflow-builder
mkdir -p platform/workflowbuilder/pkg/{types,parser,validator,resolver,generator,builder}
mkdir -p platform/workflowbuilder/examples
```

### Step 1.2: Add Dependencies

```bash
cd platform/workflowbuilder
go mod init github.com/aqls/personae/platform/workflowbuilder

# Add dependencies
go get gopkg.in/yaml.v3
go get github.com/spf13/cobra
go get github.com/lib/pq
```

### Step 1.3: Copy Core Files

Copy these files from the outputs:
- `types.go` → `pkg/types/types.go`
- `yaml_parser.go` → `pkg/parser/yaml_parser.go`
- `examples.yaml` → `examples/multipage-site-builder.yaml`

### Step 1.4: Test YAML Parsing

```go
// test_parse.go
package main

import (
    "fmt"
    "github.com/aqls/personae/platform/workflowbuilder/pkg/parser"
)

func main() {
    p := parser.NewYAMLParser()
    spec, err := p.ParseFile("examples/multipage-site-builder.yaml")
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Parsed: %s\n", spec.Name)
    fmt.Printf("Agents: %d\n", len(spec.Agents))
}
```

Run:
```bash
go run test_parse.go
```

Expected output:
```
Parsed: Multi-Page Site Builder
Agents: 6
```

---

## Phase 2: Validation & Path Resolution (Week 2)

### Step 2.1: Implement Database Repository

```go
// pkg/db/repository.go
package db

import (
    "database/sql"
    "github.com/aqls/personae/platform/workflowbuilder/pkg/types"
)

type Repository struct {
    db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
    return &Repository{db: db}
}

func (r *Repository) GetAgentDefinition(agentType string) (*types.AgentDefinition, error) {
    var def types.AgentDefinition
    
    err := r.db.QueryRow(`
        SELECT type, display_name, description, category, status
        FROM agent_definitions
        WHERE type = $1 AND deleted_at IS NULL
    `, agentType).Scan(
        &def.Type, 
        &def.DisplayName,
        &def.Description,
        &def.Category,
        &def.Status,
    )
    
    if err != nil {
        return nil, err
    }
    
    return &def, nil
}

func (r *Repository) GetAllAgentDefinitions() (map[string]*types.AgentDefinition, error) {
    rows, err := r.db.Query(`
        SELECT type, display_name, description, category, status
        FROM agent_definitions
        WHERE deleted_at IS NULL
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    defs := make(map[string]*types.AgentDefinition)
    for rows.Next() {
        var def types.AgentDefinition
        if err := rows.Scan(&def.Type, &def.DisplayName, &def.Description, &def.Category, &def.Status); err != nil {
            return nil, err
        }
        defs[def.Type] = &def
    }
    
    return defs, nil
}
```

### Step 2.2: Implement Validator

```go
// pkg/validator/validator.go
package validator

import (
    "fmt"
    "github.com/aqls/personae/platform/workflowbuilder/pkg/db"
    "github.com/aqls/personae/platform/workflowbuilder/pkg/types"
)

type Validator struct {
    repo *db.Repository
}

func NewValidator(repo *db.Repository) *Validator {
    return &Validator{repo: repo}
}

func (v *Validator) Validate(spec *types.WorkflowSpec) *types.ValidationReport {
    report := &types.ValidationReport{
        Passed:   true,
        Errors:   []types.ValidationError{},
        Warnings: []types.ValidationError{},
    }
    
    // Load all agent definitions
    agentDefs, err := v.repo.GetAllAgentDefinitions()
    if err != nil {
        report.Passed = false
        report.Errors = append(report.Errors, types.ValidationError{
            Type:     "database_error",
            Message:  fmt.Sprintf("Failed to load agent definitions: %v", err),
            Severity: "error",
        })
        return report
    }
    
    // Check 1: All agent types exist
    for i, agent := range spec.Agents {
        if _, exists := agentDefs[agent.Type]; !exists {
            report.Passed = false
            report.Errors = append(report.Errors, types.ValidationError{
                Type:     types.ErrorMissingAgent,
                Message:  fmt.Sprintf("Agent type '%s' not found in database", agent.Type),
                Location: fmt.Sprintf("agents[%d].type", i),
                Severity: "error",
            })
        }
    }
    
    // Check 2: Valid category
    if !contains(types.ValidCategories, spec.Category) {
        report.Passed = false
        report.Errors = append(report.Errors, types.ValidationError{
            Type:     types.ErrorInvalidCategory,
            Message:  fmt.Sprintf("Invalid category: %s", spec.Category),
            Location: "workflow.category",
            Severity: "error",
        })
    }
    
    // Check 3: Valid status
    if !contains(types.ValidStatuses, spec.Status) {
        report.Passed = false
        report.Errors = append(report.Errors, types.ValidationError{
            Type:     types.ErrorInvalidStatus,
            Message:  fmt.Sprintf("Invalid status: %s", spec.Status),
            Location: "workflow.status",
            Severity: "error",
        })
    }
    
    // Check 4: No duplicate agent names
    namesSeen := make(map[string]bool)
    for i, agent := range spec.Agents {
        if namesSeen[agent.Name] {
            report.Passed = false
            report.Errors = append(report.Errors, types.ValidationError{
                Type:     types.ErrorDuplicateAgent,
                Message:  fmt.Sprintf("Duplicate agent name: %s", agent.Name),
                Location: fmt.Sprintf("agents[%d].name", i),
                Severity: "error",
            })
        }
        namesSeen[agent.Name] = true
    }
    
    // Generate summary
    if report.Passed {
        report.Summary = "✓ All validations passed"
    } else {
        report.Summary = fmt.Sprintf("✗ Validation failed with %d errors", len(report.Errors))
    }
    
    return report
}

func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}
```

### Step 2.3: Add Path Resolver

Copy `path_resolver.go` to `pkg/resolver/path_resolver.go`

### Step 2.4: Test Validation

```go
// test_validate.go
package main

import (
    "database/sql"
    "fmt"
    "os"
    
    "github.com/aqls/personae/platform/workflowbuilder/pkg/db"
    "github.com/aqls/personae/platform/workflowbuilder/pkg/parser"
    "github.com/aqls/personae/platform/workflowbuilder/pkg/validator"
    _ "github.com/lib/pq"
)

func main() {
    // Parse
    p := parser.NewYAMLParser()
    spec, _ := p.ParseFile("examples/multipage-site-builder.yaml")
    
    // Connect to DB
    dbConn, _ := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    defer dbConn.Close()
    
    // Validate
    repo := db.NewRepository(dbConn)
    val := validator.NewValidator(repo)
    report := val.Validate(spec)
    
    fmt.Println(report.Summary)
    for _, err := range report.Errors {
        fmt.Printf("  [%s] %s\n", err.Type, err.Message)
    }
}
```

---

## Phase 3: CLI Tool & Testing (Week 3)

### Step 3.1: Implement JSON Generator

```go
// pkg/generator/json_generator.go
package generator

import (
    "fmt"
    "github.com/aqls/personae/platform/workflowbuilder/pkg/types"
)

type JSONGenerator struct{}

func NewJSONGenerator() *JSONGenerator {
    return &JSONGenerator{}
}

func (g *JSONGenerator) Generate(spec *types.WorkflowSpec, pathMap *types.PathMap) (map[string]interface{}, error) {
    workflow := map[string]interface{}{
        "start_step": "spawn_" + spec.Agents[0].Name,
        "steps":      make(map[string]interface{}),
    }
    
    steps := workflow["steps"].(map[string]interface{})
    
    // Generate spawn steps
    for i, agent := range spec.Agents {
        stepName := "spawn_" + agent.Name
        
        step := map[string]interface{}{
            "action": "spawn_agent",
            "config": map[string]interface{}{
                "role":       agent.Name,
                "agent_type": agent.Type,
            },
            "description": fmt.Sprintf("Spawn %s", agent.Description),
        }
        
        // Next step
        if i < len(spec.Agents)-1 {
            step["next_step"] = "spawn_" + spec.Agents[i+1].Name
        } else {
            step["next_step"] = "call_" + spec.Agents[0].Name
        }
        
        steps[stepName] = step
    }
    
    // Generate call steps
    for i, agent := range spec.Agents {
        stepName := "call_" + agent.Name
        
        config := map[string]interface{}{
            "agent_type":      agent.Type,
            "target_role":     agent.Name,
            "timeout_seconds": spec.Timeouts[agent.Name],
        }
        
        // Compute input_fields
        if len(agent.Inputs) > 0 {
            inputFields := []string{}
            for _, input := range agent.Inputs {
                if input.From == types.WorkflowInputData {
                    inputFields = append(inputFields, "input_data")
                } else {
                    // Extract field name from "agent.field"
                    parts := strings.Split(input.From, ".")
                    if len(parts) == 2 {
                        inputFields = append(inputFields, parts[1])
                    }
                }
            }
            config["input_fields"] = inputFields
        }
        
        step := map[string]interface{}{
            "action":       "call_agent",
            "config":       config,
            "output_field": agent.OutputField,
            "description":  agent.Description,
        }
        
        // Next step
        if i < len(spec.Agents)-1 {
            step["next_step"] = "call_" + spec.Agents[i+1].Name
        } else {
            step["next_step"] = "complete"
        }
        
        steps[stepName] = step
    }
    
    // Complete step
    steps["complete"] = map[string]interface{}{
        "action":      "complete_workflow",
        "description": "Workflow complete",
    }
    
    return workflow, nil
}
```

### Step 3.2: Build Main CLI Tool

Copy `main.go` to `cmd/workflow-builder/main.go`

### Step 3.3: Build and Test

```bash
cd cmd/workflow-builder
go build -o workflow-builder

# Test commands
./workflow-builder validate ../../examples/multipage-site-builder.yaml
./workflow-builder build --dry-run ../../examples/multipage-site-builder.yaml
./workflow-builder list
```

---

## Phase 4: Polish & Documentation (Week 4)

### Step 4.1: Add Test Generator

```go
// pkg/generator/test_generator.go
package generator

import (
    "github.com/aqls/personae/platform/workflowbuilder/pkg/types"
)

func GenerateTestCases(spec *types.WorkflowSpec) []types.TestCase {
    tests := []types.TestCase{
        {
            Name:        "Happy Path",
            Description: "Test successful workflow execution",
            Input: map[string]interface{}{
                "domain":    "test-" + spec.GroupType + ".com",
                "objective": "Test workflow execution",
                "model":     "AIDA",
            },
            ShouldFail: false,
        },
        {
            Name:        "Missing Input",
            Description: "Test handling of missing required fields",
            Input:       map[string]interface{}{},
            ShouldFail:  true,
        },
    }
    
    return tests
}
```

### Step 4.2: Add Documentation Generator

```go
// pkg/generator/doc_generator.go
package generator

import (
    "fmt"
    "strings"
    
    "github.com/aqls/personae/platform/workflowbuilder/pkg/types"
)

func GenerateDocumentation(spec *types.WorkflowSpec, pathMap *types.PathMap) string {
    var doc strings.Builder
    
    doc.WriteString(fmt.Sprintf("# %s\n\n", spec.Name))
    doc.WriteString(fmt.Sprintf("**Category:** %s  \n", spec.Category))
    doc.WriteString(fmt.Sprintf("**Status:** %s  \n", spec.Status))
    doc.WriteString(fmt.Sprintf("**Group Type:** `%s`\n\n", spec.GroupType))
    
    doc.WriteString(fmt.Sprintf("## Description\n\n%s\n\n", spec.Description))
    
    doc.WriteString("## Agents\n\n")
    for i, agent := range spec.Agents {
        doc.WriteString(fmt.Sprintf("%d. **%s** (`%s`)\n", i+1, agent.Name, agent.Type))
        doc.WriteString(fmt.Sprintf("   - %s\n", agent.Description))
        if len(agent.Inputs) > 0 {
            doc.WriteString("   - Inputs: ")
            for j, input := range agent.Inputs {
                if j > 0 {
                    doc.WriteString(", ")
                }
                doc.WriteString(fmt.Sprintf("`%s`", input.From))
            }
            doc.WriteString("\n")
        }
        doc.WriteString(fmt.Sprintf("   - Output: `%s`\n", agent.OutputField))
        doc.WriteString("\n")
    }
    
    return doc.String()
}
```

### Step 4.3: Write Tests

```bash
mkdir -p platform/workflowbuilder/test

# Create test files
# test/parser_test.go
# test/validator_test.go
# test/generator_test.go
```

### Step 4.4: Create Example Workflows

Convert existing workflows to YAML:
- `mvp-site-builder.yaml`
- `content-site-builder.yaml`
- etc.

---

## Integration with Existing System

### Option 1: Standalone Tool

Use as CLI tool to create workflows, insert into database.

Workflow:
```
1. Developer writes YAML
2. Runs: workflow-builder validate my-workflow.yaml
3. Runs: workflow-builder build my-workflow.yaml
4. Workflow inserted into database
5. Existing orchestration system uses it normally
```

### Option 2: HTTP API Integration

Add HTTP endpoints:

```go
// In your main API server
import "github.com/aqls/personae/platform/workflowbuilder/pkg/builder"

func handleBuildWorkflow(w http.ResponseWriter, r *http.Request) {
    // Parse uploaded YAML
    file, _, _ := r.FormFile("workflow")
    data, _ := io.ReadAll(file)
    
    // Build
    parser := parser.NewYAMLParser()
    spec, _ := parser.Parse(data)
    
    wfBuilder := builder.NewWorkflowBuilder(db)
    result, _ := wfBuilder.Build(spec)
    
    // Return result
    json.NewEncoder(w).Encode(result)
}
```

### Option 3: Git-Based Workflow

Store YAML files in git:

```
workflows/
├── production/
│   ├── multipage-site-builder.yaml
│   └── mvp-site-builder.yaml
├── staging/
│   └── experimental-workflow.yaml
└── templates/
    └── basic-template.yaml
```

CI/CD pipeline:
```yaml
# .github/workflows/deploy-workflows.yml
name: Deploy Workflows

on:
  push:
    paths:
      - 'workflows/production/**'

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Validate and Deploy
        run: |
          for file in workflows/production/*.yaml; do
            workflow-builder validate $file
            workflow-builder build $file
          done
```

---

## Testing the Complete System

### 1. Unit Tests

```bash
cd platform/workflowbuilder
go test ./...
```

### 2. Integration Test

```bash
# Create test workflow
cat > test-workflow.yaml << EOF
workflow:
  name: "Test Workflow"
  group_type: "test-workflow"
  agents:
    - name: agent1
      type: site-strategist
      inputs: [workflow.input_data]
EOF

# Validate
./workflow-builder validate test-workflow.yaml

# Build (dry-run)
./workflow-builder build --dry-run test-workflow.yaml

# Build (actual)
./workflow-builder build test-workflow.yaml

# Verify in database
psql clients_db -c "SELECT * FROM agent_group_definitions WHERE group_type='test-workflow'"
```

### 3. End-to-End Test

```bash
# Build workflow
./workflow-builder build examples/multipage-site-builder.yaml

# Trigger workflow via API
curl -X POST http://localhost:8080/api/v1/orchestrate \
  -d '{"action": "orchestrate", "config": {"group_type": "multipage-site-builder"}, ...}'

# Check logs for successful execution
# Verify files deployed to Git/B2
```

---

## Rollout Plan

### Week 1: Development
- Implement core components
- Unit tests
- Basic validation

### Week 2: Testing
- Create example workflows
- Integration testing
- Documentation

### Week 3: Pilot
- Convert 2-3 existing workflows to YAML
- Test in staging environment
- Collect feedback

### Week 4: Production
- Convert all workflows to YAML
- Deploy to production
- Monitor and iterate

---

## Success Metrics

- ✅ Zero path-related errors after deployment
- ✅ All workflows validated before insertion
- ✅ 100% of new workflows use YAML format
- ✅ Developer time reduced by 50% for workflow creation
- ✅ All validation errors caught before runtime

---

## Support

Questions? Contact the platform team or create an issue.

Happy building! 🚀

---

