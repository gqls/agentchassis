Role-Based Task Queue Architecture
Instead of spawning agents tied to specific IDs, agents claim work based on their capabilities and assigned roles:

// Agent announces its role and capabilities when starting
type AgentRegistration struct {
AgentID      string
AgentType    string
Role         string
Capabilities []string
Status       string // "ready", "busy", "failed"
LastHeartbeat time.Time
}

// Work items are posted to role-specific queues
type WorkItem struct {
WorkID        string
RequiredRole  string
RequiredType  string
OrchestrationID string
Priority      int
Data          interface{}
ClaimedBy     *string // nil if unclaimed
ClaimedAt     *time.Time
ExpiresAt     time.Time
}

Implementation Approach
1. Role-Based Topics

system.roles.adder.pending        # Work waiting for adder role
system.roles.multiplier.pending   # Work waiting for multiplier role
system.roles.{role}.claimed      # Work being processed

2. Agent Lifecycle

// Agent starts up
func (a *Agent) Initialize() {
// Register with role
a.registerRole(a.Role)

    // Start consuming from role queue
    topic := fmt.Sprintf("system.roles.%s.pending", a.Role)
    a.consumeFrom(topic)
    
    // Heartbeat to show we're alive
    go a.heartbeat()
}

// Claim work
func (a *Agent) claimWork(workItem WorkItem) bool {
// Atomic claim via DB
query := `
        UPDATE work_items 
        SET claimed_by = $1, claimed_at = NOW() 
        WHERE work_id = $2 AND claimed_by IS NULL
        RETURNING work_id
    `
// If successful, process work
}

3. Orchestrator Changes

// Instead of spawning specific agents
func CallAgentAction(ctx context.Context, params ActionParams) (interface{}, error) {
targetRole := config["target_role"].(string)

    // Create work item
    workItem := WorkItem{
        WorkID:       uuid.New().String(),
        RequiredRole: targetRole,
        RequiredType: "calculator",
        Data:         params.CollectedData[inputField],
        ExpiresAt:    time.Now().Add(30 * time.Second),
    }
    
    // Post to role queue
    topic := fmt.Sprintf("system.roles.%s.pending", targetRole)
    sendWorkItem(topic, workItem)
    
    // Track for response
    return map[string]interface{}{
        "work_id": workItem.WorkID,
        "role": targetRole,
        "await_response": true,
    }, nil
}

4. Spawn Strategy

// Spawn pool of workers with roles
func SpawnAgentAction() {
// Check if we have enough workers for this role
activeWorkers := countActiveWorkers(role)
if activeWorkers < minWorkers {
// Spawn new worker with this role
spawnWorkerWithRole(role)
}
}

Benefits

No message stealing: Work items are claimed atomically
Automatic failover: If an agent dies, unclaimed work can be picked up by another
Elastic scaling: Spawn more workers when queue builds up
Role flexibility: Agents can have multiple roles/capabilities
Clean separation: No more broadcast confusion

Migration Path

Phase 1: Keep existing spawn but add role registration
Phase 2: Implement work queue for new agents
Phase 3: Migrate calculators to role-based model
Phase 4: Extend to all agent types

This aligns perfectly with your Kubernetes job model - if a pod crashes, another can claim its incomplete work items. The role becomes the contract, not the agent ID.