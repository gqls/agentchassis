1. Client → Generic Agent (Initial Request)
   Agent: Generic
   Function: ProcessMessage (processor.go)

// Line 1088: Create message context with perspective transformation
msgCtx, err := NewMessageContext(msg, headers, p.agentType) // p.agentType = "generic"

BEFORE: No context exists

AFTER NewMessageContext:
Since this is a request from client, the generic agent creates its OWN orchestration
execCtx.OrchestrationID = NEW_GENERIC_ID (generated) - NEW_GENERIC_ORCH_123
execCtx.ParentOrchestrationID: "" (root)
execCtx.Sender: {AgentType: "generic"}
ResponsesTopic: "system.agent.generic.responses"
RequestID: CLIENT_REQ_456

In ExecuteWorkflow (coordinator.go line 92):
Uses execCtx.OrchestrationID to create/find state ✅
State gets created with OwnerAgentType: "generic" ✅

2. Generic → Calculator (CallAgentAction)
   Agent: Generic
   Function: CallAgentAction

Function: executeLocalAction (coordinator.go line 745)
// execCtx represents GENERIC's perspective
execCtx.OrchestrationID = NEW_GENERIC_ID // Generic's own
execCtx.Sender = {AgentType: "generic"}
In CallAgentAction:
Creates child request with ResponsesTopic: "system.agent.generic.responses" ✅
Stores __execution_context__ with generic's context for later retrieval ✅

SENDING:
OrchestrationID: NEW_GENERIC_ORCH_123 (my orchestration)
RequestID: CALC_REQ_789
ResponsesTopic: "system.agent.generic.responses"

3. Calculator Receives Request
   Agent: Calculator
   Function: ProcessMessage (processor.go)

msgCtx, err := NewMessageContext(msg, headers, p.agentType) // p.agentType = "calculator"

AFTER NewMessageContext:
execCtx.OrchestrationID: NEW_CALC_ORCH_ABC (my new orchestration) - Calculator creates ITS OWN orchestration
execCtx.ParentOrchestrationID: NEW_GENERIC_ORCH_123 - from sender (who called me) (not actually NEW)
execCtx.Sender: {AgentType: "calculator"}
execCtx.ResponsesTopic: "system.agent.generic.responses" (preserved)
RequestID: CALC_REQ_789

4. Calculator → Generic (CompleteWorkflowAction)
   Agent: Calculator
   Function: CompleteWorkflowAction (workflow_actions.go line 58)
// Check if child
if params.ExecutionContext.ParentOrchestrationID != "" {
// YES - calculator has ParentOrchestrationID = NEW_GENERIC_ORCH_123
 
Line 62-75: Extracts parent info from __execution_context__:
Gets parentResponseTopic: "system.agent.generic.responses" ✅
Gets parentRequestID from stored context ✅

Line 134: Sends response with:
InResponseToParentOrchID: NEW_GENERIC_ORCH_123 ✅
Topic: "system.agent.generic.responses" ✅

SENDING RESPONSE:
MyOrchestrationID: NEW_CALC_ORCH_ABC
InResponseToParentOrchID: NEW_GENERIC_ORCH_123
InResponseToRequestID: CALC_REQ_789

5. Generic Receives Response
   Agent: Generic
   Function: ProcessMessage (processor.go)

msgCtx, err := NewMessageContext(msg, headers, p.agentType) // p.agentType = "generic"

AFTER NewMessageContext:
The perspective flip happens here
execCtx.OrchestrationID: NEW_GENERIC_ORCH_123 (back to my context!)
execCtx.InResponseTo: {
ParentOrchestrationID: NEW_CALC_ORCH_ABC (who responded)
RequestID: CALC_REQ_789
}
Sender: {AgentType: "generic"}

In ProcessResponse (coordinator.go line 133):

Uses FindByAwaitedRequestID to find state ✅
Finds GENERIC's state (not calculator's) ✅
Stores response in CollectedData ✅

6. Generic → Client (Final Response)
   Agent: Generic
   Function: CompleteWorkflowAction (workflow_actions.go line 173)**
// YES - generic is root, no parent

Check ParentOrchestrationID == "" (yes, I'm root)
Get original request from CollectedData["__original_request__"]
Send to: "system.agent.generic.responses"
RequestID: CLIENT_REQ_456

Line 176-194: Tries to find original client info:

Gets originalResponseTopic from ExecutionContext.ResponsesTopic ✅
Gets originalRequestID from ExecutionContext.RequestID ✅
Sends final response to client ✅

Key Points

Each agent's ExecutionContext.OrchestrationID is THEIR OWN after transformation
ParentOrchestrationID correctly identifies who called them
ResponsesTopic is preserved through the chain
Original client request info is stored in CollectedData for final response

The critical fix is in ProcessMessage line 1088:

// Change from:
msgCtx, err := NewMessageContext(msg, headers)

// To:
msgCtx, err := NewMessageContext(msg, headers, p.agentType)

This ensures every agent sees the conversation from their own perspective.

===

Issues Found

In ProcessResponse (coordinator.go):

After perspective transformation, execCtx.OrchestrationID should already be the receiver's
No need to look up by InResponseTo.ParentOrchestrationID
The state lookup by FindByAwaitedRequestID should work correctly


In CompleteWorkflowAction:

For root orchestration, the original request info needs to be preserved
Currently relies on ExecutionContext having the original values
But after transformation, these might be lost

