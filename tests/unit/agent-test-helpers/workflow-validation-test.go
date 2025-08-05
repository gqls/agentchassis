// Workflow validation
func TestWorkflowValidation(t *testing.T) {
tests := []struct {
name     string
workflow models.WorkflowPlan
wantErr  bool
}{
{
name: "valid simple workflow",
workflow: validSimpleWorkflow(),
wantErr: false,
},
{
name: "missing start step",
workflow: workflowWithoutStartStep(),
wantErr: true,
},
// ... more cases
}
}

// State management
func TestOrchestrationState(t *testing.T) {
// Test state creation, updates, retrieval
}