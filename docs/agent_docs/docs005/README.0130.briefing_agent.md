User Input → Briefing Agent (gathers context, asks clarifying questions)
→ Human reviews/approves brief
→ Strategist (uses the enriched brief)
→ Architect → Content Creator → Deployer
The briefing agent could:

Take initial input
Ask targeted questions ("What's your unique selling point?", "Who are your competitors?", "What tone - corporate/friendly/technical?")
Compile a structured brief
Pause for human approval before proceeding

meanwhile here's a brief starter prompt:
{
"action": "orchestrate",
"config": {
"group_type": "mvp-site-builder"
},
"input_data": {
"domain": "ai-agent-orchestration.com",
"objective": "Sell an AI multi-agent orchestration framework as a service. The framework enables businesses to automate complex workflows by coordinating specialized AI agents that work together. Key capabilities include: automated website creation from just a domain name, content generation using proven persuasion models (AIDA, PAS, Cialdini), intelligent component assembly from pre-built libraries, and automated deployment to live hosting. The service offering: we set up custom agent workflows tailored to client objectives - whether that's instant website creation, content pipelines, data processing, or any multi-step AI workflow. Target audience: businesses wanting to leverage AI automation without building infrastructure, agencies needing scalable content/site production, and developers wanting a ready-made orchestration backbone. Emphasize: simple instruction, no-code setup for clients, human-in-the-loop at key decision points, scales from single sites to thousands, pay-per-use pricing model.",
"model": "AIDA with Cialdini social proof and authority",
"repo_name": "sites"
}
}
```

---

Briefing Agent Summary (for later):
A new agent type briefing-agent that:

Sits before chief-strategist in the workflow
Takes raw user input (domain, rough objective)
Generates clarifying questions based on the industry/objective detected
Collects answers (via Kafka message back to client, or stored for human review)
Outputs a structured brief: target audience, tone, USPs, competitor context, key messages, desired sections
Passes enriched brief to strategist

Workflow change: spawn_briefing → call_briefing → (human approval pause) → spawn_strategist → ...