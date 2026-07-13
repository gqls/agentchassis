complete flow:
1. Client sends request → main topic → consumer → starts workflow
2. Workflow needs help from other agents → sends to their topics
3. Other agents work and send responses → response topic → responseConsumer
4. responseConsumer → HandleResponse → workflow continues
5. Workflow completes → sends final response

visual example

Main Consumer (system.agent.generic.process):
← "Create marketing content about AI"

Your agent starts workflow, realizes it needs:
→ Send to reasoning agent: "Analyze AI benefits"
→ Send to image agent: "Create AI visualization"

Response Consumer (system.responses.generic):
← Response from reasoning: "AI benefits include..."
← Response from image: "Image URL: ..."

Your agent continues workflow with both responses

Your agent is both a SERVICE (processing requests via main consumer) 
and a CLIENT (receiving responses via response consumer) in the multi-agent system!

--
separating out client and server
AgentServer - Handles incoming requests
├── Listens to main topic (system.agent.generic.process)
├── Processes new work requests
└── Creates/manages workflows

AgentClient - Handles responses from other agents
├── Listens to response topic (system.responses.generic)
├── Routes responses to orchestrator
└── Manages response handling

Agent (Coordinator)
├── Owns AgentServer
├── Owns AgentClient
└── Coordinates between them