Step 1: Initial Request Arrives at Generic Agent
Agent: Generic Agent (agent-chassis-f8bc84dcf-z42vl)
Function: ProcessMessage → process (processor.go)
Time: 10:01:50

Message arrives with action: "calculate" and calculation data
Creates orchestration ID: 82e595f2-de57-44f8-9291-0a6e6e0f74c4
Loads workflow from DB with start step: spawn_calculator

Step 2: Workflow Execution Begins
Agent: Generic Agent
Function: executeWorkflow → ExecuteWorkflow (coordinator.go)
Time: 10:01:50

Creates orchestration state in database
Status: INITIALIZED → EXECUTING_STEP
Current step: spawn_calculator

Step 3: Execute Spawn Action
Agent: Generic Agent
Function: executeLocalAction → SpawnAgentAction (spawn_actions.go)
Time: 10:01:50-51

Generates agent ID: 30e48b55-9bc5-480a-b494-312e5556eacf
Creates agent in database
Creates Kubernetes Job: agent-calculator-30e48b55
Sends initialization message to system.agent.calculator.requests
Request ID: fd1b94ed-3f20-485c-8213-6f7534717fa4
Response topic: system.agent.generic.responses

Step 4: Action Marked as Awaiting Response
Agent: Generic Agent
Function: executeLocalAction (coordinator.go line 702)
Time: 10:01:52

Adds request to AwaitedRequests map
Sets state status to AWAITING_RESPONSES
Saves state to database
Returns from continueExecution with "waiting for responses"

Step 5: Calculator Agent Starts
Agent: Calculator Agent (agent-calculator-30e48b55-9sj79)
Function: Agent initialization
Time: 10:02:08

Kubernetes job starts
Receives initialization message
Function: SendInitializationResponse (agent.go line 943)
Sends response to system.agent.generic.responses

Step 6: WHERE IT BREAKS - Response Not Processed
Expected: Generic agent should receive initialization response on system.agent.generic.responses
Expected Function: Response processor → ProcessResponse → handleCompleteResponse
What Should Happen:

Match response to request ID fd1b94ed-3f20-485c-8213-6f7534717fa4
Remove from AwaitedRequests
Advance to next step: first_calculation

What Actually Happens: Nothing - silence for 2 minutes
Step 7: Timeout and Retry
Agent: Generic Agent
Function: handleRequestTimeout (coordinator.go)
Time: 10:03:52

Request times out after 120 seconds
Attempts retry

The Problem
The generic agent IS consuming from system.agent.generic.responses (we can see it created the consumer at startup), but when the calculator's initialization response arrives, it's not being processed correctly.
Looking at the calculator logs, the initialization response is sent with:

in_response_to_request_id: "fd1b94ed-3f20-485c-8213-6f7534717fa4"
Topic: system.agent.generic.responses

The issue is likely in ProcessMessage in processor.go. When a response arrives, it should:

Detect it's a response (line 1236: checks MessageType == "response")
Route to orchestrator (line 1241: p.orchestrator.ProcessResponse)

But the orchestrator's ProcessResponse needs to properly extract the request ID from the response headers.