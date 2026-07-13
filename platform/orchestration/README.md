https://claude.ai/chat/2d57dc36-1686-4021-8382-919342d4fa6e

What is a Workflow Plan?
A workflow plan is the complete blueprint of all steps in a workflow:
type WorkflowPlan struct {
StartStep string          `json:"start_step"`  // Where to begin
Steps     map[string]Step `json:"steps"`       // ALL steps in the workflow
}

// Example workflow plan:
plan := WorkflowPlan{
StartStep: "validate",
Steps: map[string]Step{
"validate": Step{
Action: "validate_input",
NextStep: "search",
},
"search": Step{
Action: "web_search",
Topic: "system.agent.web-search.process",
NextStep: "analyze",
},
"analyze": Step{
Action: "reasoning",
Topic: "system.agent.reasoning.process",
NextStep: "generate",
},
"generate": Step{
Action: "content_creation",
Topic: "system.agent.content-creator.process",
NextStep: "complete",
},
"complete": Step{
Action: "complete_workflow",
},
},
}

What is a Step?
A step is a single action within the workflow. It can be:

Local: Executed immediately by the orchestrator (validate_input, transform_data)
Remote: Sends a message to another agent and waits for response (web_search, content_creation)

The Current Problem
Right now, when the orchestrator executes a workflow:

It receives a workflow plan and starts executing
For local actions, it executes them immediately
For remote actions, it sends a message and updates state to "awaiting response"
When a response comes back, it needs to know "what's next?"

The problem: After executing a local action or receiving a response, the orchestrator needs the original workflow plan to know what step comes next. Currently, we lose this information.
Solution Options in Detail
Option 1: Store Workflow Plan in Database State (Recommended)
How it works:

-- Add workflow_plan column to orchestrator_state table
ALTER TABLE orchestrator_state ADD COLUMN workflow_plan JSONB;

// Update the state structure
type OrchestrationState struct {
CorrelationID      string
WorkflowPlan       models.WorkflowPlan  // Full plan stored here
CurrentStep        string               // "search"
Status             OrchestrationStatus  // "AWAITING_RESPONSES"
CollectedData      map[string]interface{} // Results so far
// ... other fields
}

Advantages:

✅ Complete workflow visibility at any time
✅ Can resume from any point after crashes
✅ Easy debugging - see entire workflow in database
✅ Supports long-running workflows (days/weeks)
✅ Can modify workflow while in-flight if needed

Disadvantages:

❌ More storage space (but JSONB is efficient)
❌ Need to migrate existing database

Implementation:

func (s *SagaCoordinator) ExecuteWorkflow(ctx context.Context, plan models.WorkflowPlan, headers map[string]string, initialData []byte) error {
// Get or create state WITH plan
state, err := s.getOrCreateState(ctx, correlationID, clientID, plan, initialData)

    // Now state always has the full plan
    // Can continue execution from any point
}

// When handling responses:
func (s *SagaCoordinator) HandleResponse(ctx context.Context, headers map[string]string, response []byte) error {
state, _ := repo.GetState(ctx, correlationID)
// state.WorkflowPlan has the full plan!

    // Continue to next step
    nextStep := state.WorkflowPlan.Steps[state.CurrentStep]
    return s.executeStep(ctx, nextStep, state)
}

Option 2: Pass Plan Through Execution Chain
How it works:
Keep passing the plan through all method calls.
Advantages:

✅ No database changes needed
✅ Simple to understand

Disadvantages:

❌ Can't resume after system restart
❌ Plan lost if process crashes
❌ Doesn't work well with async responses
❌ Not suitable for long-running workflows

Option 3: Store Plan Separately with Reference
How it works:

-- Create workflow_plans table
CREATE TABLE workflow_plans (
id UUID PRIMARY KEY,
plan JSONB NOT NULL,
created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Reference from state
ALTER TABLE orchestrator_state ADD COLUMN workflow_plan_id UUID;

Advantages:

✅ Reusable plans (many executions, same plan)
✅ Smaller state records
✅ Can version plans

Disadvantages:

❌ Extra join needed
❌ More complex

Option 4: Hybrid Approach (Best for Complex Systems)
How it works:
Combine stored plans with dynamic modification:

type OrchestrationState struct {
// Base plan reference
WorkflowPlanID     string

    // Runtime modifications
    PlanOverrides      map[string]Step  // Steps modified during execution
    DynamicSteps       map[string]Step  // Steps added during execution
    
    // Execution state
    CurrentStep        string
    ExecutionPath      []string         // History of steps taken
    DecisionPoints     map[string]interface{} // Why certain paths were chosen
}

My Recommendation: Option 1 with Enhancements
Here's why and how:
1. Store the Full Plan in State
   -- Migration
   ALTER TABLE orchestrator_state
   ADD COLUMN workflow_plan JSONB,
   ADD COLUMN execution_metadata JSONB DEFAULT '{}';

2. Enhanced State Structure
   type OrchestrationState struct {
   CorrelationID      string
   ClientID           string
   Status             OrchestrationStatus

   // Workflow definition
   WorkflowPlan       models.WorkflowPlan

   // Execution state
   CurrentStep        string
   ExecutionPath      []ExecutionRecord  // History

   // Data management  
   CollectedData      map[string]interface{}
   InitialRequestData json.RawMessage
   FinalResult        json.RawMessage

   // Async handling
   AwaitedSteps       []string  // Request IDs we're waiting for

   // Debugging/Monitoring
   ExecutionMetadata  ExecutionMetadata
   Error              string

   CreatedAt          time.Time
   UpdatedAt          time.Time
   }

type ExecutionRecord struct {
Step      string
Action    string
StartTime time.Time
EndTime   *time.Time
Result    string // success/failed/skipped
Error     string
}

type ExecutionMetadata struct {
TotalSteps     int
CompletedSteps int
SkippedSteps   int
RetryCount     map[string]int
Checkpoints    map[string]time.Time
}

3. Execution Flow

func (s *SagaCoordinator) ExecuteWorkflow(ctx context.Context, plan models.WorkflowPlan, headers map[string]string, initialData []byte) error {
state, err := s.getOrCreateState(ctx, correlationID, clientID, plan, initialData)
if err != nil {
return err
}

    // State now includes the full plan
    return s.continueExecution(ctx, state, headers)
}

func (s *SagaCoordinator) continueExecution(ctx context.Context, state *OrchestrationState, headers map[string]string) error {
// Get current step from the stored plan
currentStep, exists := state.WorkflowPlan.Steps[state.CurrentStep]
if !exists {
return s.failWorkflow(ctx, state, "current step not found in plan")
}

    // Record execution
    s.recordExecutionStart(state, currentStep)
    
    // Execute based on action type
    switch {
    case isLocalAction(currentStep.Action):
        return s.executeLocalAction(ctx, state, currentStep, headers)
    case currentStep.Topic != "":
        return s.executeRemoteAction(ctx, state, currentStep, headers)
    default:
        return s.executeBuiltinAction(ctx, state, currentStep, headers)
    }
}

// When handling async responses:
func (s *SagaCoordinator) HandleResponse(ctx context.Context, headers map[string]string, response []byte) error {
state, err := s.repo.GetState(ctx, headers["correlation_id"])
if err != nil {
return err
}

    // Process response...
    s.processResponse(state, headers["causation_id"], response)
    
    // Continue workflow with stored plan
    if len(state.AwaitedSteps) == 0 {
        return s.continueExecution(ctx, state, headers)
    }
    
    return nil
}

4. Benefits for Complex Workflows
   This approach enables:
   Dynamic Workflows:

// Workflow can modify itself based on results
if searchResults.Count > 10 {
// Add a summarization step dynamically
state.WorkflowPlan.Steps["summarize"] = Step{
Action: "summarize_results",
NextStep: state.WorkflowPlan.Steps["analyze"].NextStep,
}
state.WorkflowPlan.Steps["analyze"].NextStep = "summarize"
}

Parallel Execution:
// Fan-out to multiple agents
state.WorkflowPlan.Steps["gather_data"] = Step{
Action: "fan_out",
SubTasks: []SubTask{
{StepName: "search_web", Topic: "system.agent.web-search.process"},
{StepName: "search_db", Topic: "system.agent.db-search.process"},
{StepName: "search_docs", Topic: "system.agent.doc-search.process"},
},
NextStep: "merge_results",
}

Conditional Branching:
type Step struct {
Action       string
Conditions   []Condition  // New field
NextStep     string       // Default next
// ...
}

type Condition struct {
Expression string  // "result.score > 0.8"
NextStep   string  // Where to go if true
}

Monitoring:
// Query to see workflow progress
SELECT
correlation_id,
workflow_plan->>'start_step' as workflow_type,
current_step,
execution_metadata->>'completed_steps' as completed,
execution_metadata->>'total_steps' as total,
status,
created_at,
updated_at
FROM orchestrator_state
WHERE client_id = ? AND status != 'COMPLETED'
ORDER BY updated_at DESC;

Summary
For a system with many agents and complex workflows, storing the complete workflow plan in the database state is the best approach because:

Resilience: Survives restarts, crashes, deployments
Visibility: Can query "show me all workflows stuck at step X"
Debugging: Complete execution history with the plan
Flexibility: Can modify workflows in-flight
Scale: Supports workflows running for hours/days/weeks
Monitoring: Rich metadata for dashboards and alerts

The small overhead of storing the plan (usually just a few KB of JSON) is worth the massive operational benefits you get in return.