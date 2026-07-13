Here's the complete flow from initial request to domain-analyst response:
1. Initial Message to Generic Agent
   Topic: system.agent.generic.requests
   Message: Website design request
   Headers: correlation_id, client_id, request_id, etc.
   Function Flow:

Generic Agent receives message → agentbase.(*AgentServer).processMessage()
→ messaging.(*MessageProcessor).ProcessMessage()
→ messaging.(*MessageProcessor).process()
→ messaging.(*MessageProcessor).executeWorkflow() (generic has a workflow)
→ orchestration.(*SagaCoordinator).ExecuteWorkflow()
→ orchestration.(*SagaCoordinator).getOrCreateState() (creates parent orchestration state)
→ orchestration.(*SagaCoordinator).continueExecution() with step: spawn_website_team

2. Spawn Website Team

→ orchestration.(*SagaCoordinator).executeLocalAction() for spawn_group
→ actions.SpawnGroupAction()

Queries DB for agent_groups where group_type='website-builder'
For each agent in group, calls SpawnAgentAction()


→ actions.SpawnAgentAction() spawns each agent (domain-analyst, site-architect, etc.)


Creates K8s jobs
Returns spawned agent IDs

3. Start Child Orchestration

→ Next step: start_website_workflow
→ actions.StartOrchestrationAction()


Creates new correlation_id for child
Calls SagaCoordinator.CreateNewOrchestration()


→ orchestration.(*SagaCoordinator).CreateNewOrchestration()


Creates child orchestration state in DB
Sets parent_correlation_id
Starts child execution

4. Child Orchestration Executes

→ Child continueExecution() with step: analyze_domain
→ executeLocalAction() for call_agent with agent_type: domain-analyst
→ actions.CallAgentAction()
→ actions.callSpecificAgent() with domain-analyst ID
→ actions.sendNewFormatMessage()

Message Sent to Kafka:
Topic: system.agent.domain-analyst.requests
Headers:

correlation_id: (child's correlation ID)
request_id: (new UUID)
responses_topic: system.agent.website-builder.responses

5. Domain Analyst Processes

Domain Analyst receives → agentbase.(*AgentServer).processMessage()
→ messaging.(*MessageProcessor).ProcessMessage()
→ messaging.(*MessageProcessor).executeWorkflow() (domain-analyst's workflow)
→ Executes workflow step: execute_llm_prompt
→ actions.ExecuteLLMPromptAction() generates analysis

6. Domain Analyst Responds

→ Domain analyst completes workflow
→ messaging.(*MessageProcessor).sendResponse()

Response Message to Kafka:
Topic: system.agent.website-builder.responses (from responses_topic)
Headers:

correlation_id: (child's correlation ID)
causation_id: (the request_id from step 18)

7. Website-Builder Child Orchestration Receives Response

Website-builder (child orchestration) listening on responses → agentbase.(*AgentClient).processResponse()
→ messaging.(*MessageProcessor).ProcessResponse()
→ orchestration.(*SagaCoordinator).HandleResponse()
→ Stores response in state.CollectedData
→ Removes from AwaitedSteps
→ Continues to next step: architect_site

This continues until all steps complete, then:
8. Child Notifies Parent

Child completes all steps → completeWorkflow()
→ Sends completion message to parent

Completion Message to Kafka:
Topic: system.orchestrator.responses or system.agent.generic.responses
Headers:

correlation_id: (parent's correlation ID)
causation_id: (child's correlation ID)

9. Parent Receives Completion

Parent orchestration → HandleResponse()
→ Continues parent workflow or completes

The key issue points to check:

Are the agent workflows calling themselves (self-recursion)?
Is the responses_topic being set correctly?
Are correlation IDs being maintained properly between parent/child?
Is the website-builder agent actually listening to its responses topic?
