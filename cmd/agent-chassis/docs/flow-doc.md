1. Client sends message → agent-chassis
2. agent-chassis creates workflow state in DB
3. Orchestrator executes fan_out action
4. Orchestrator sends messages to:
    - system.agent.reasoning.process (with request_id: a14a52ce-...)
    - system.adapter.image.generate (with request_id: 4278de45-...)
5. Orchestrator updates state to AWAITING_RESPONSES
6. Other agents consume these messages
7. e.g. Reasoning agent processes and sends response to → system.responses.generic
      Headers: correlation_id=ed240794..., causation_id=a14a52ce-...
8. e.g. Image generator processes and sends response to → system.responses.generic
   Headers: correlation_id=ed240794..., causation_id=4278de45-...

The orchestrator is stateless - it doesn't keep workflows in memory. Instead:

Workflow state is in the database
Multiple orchestrator instances can handle any workflow
Responses can arrive at any instance

So the flow is:

Response arrives at ANY agent-chassis instance
That instance loads the workflow state from DB
Updates the state with the response
Continues execution if all responses are received

--

The orchestrator has all the logic to:

Find the workflow
Match responses to awaited steps
Continue execution
Complete the workflow