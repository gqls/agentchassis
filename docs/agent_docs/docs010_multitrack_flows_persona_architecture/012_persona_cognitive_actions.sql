-- ============================================================================
-- PERSONA COGNITIVE ACTIONS
-- Implementation guide for cognitive subsystem actions
-- ============================================================================

/*
This file documents the custom actions needed for the persona cognitive architecture.
Each action can be implemented with LLMs initially, then swapped for specialized systems.

Actions are implemented in Go in the agent-chassis codebase.
*/

-- ============================================================================
-- ACTION: load_cognitive_system
-- ============================================================================
/*
Loads all active cognitive components for a persona and initializes the instance.

Input:
  - persona_id: UUID
  - orchestration_id: UUID

Output:
  - persona_profile: JSONB (personality DNA)
  - cognitive_components: JSONB (map of component_type -> implementation)
  - instance_id: UUID
  - working_memory: JSONB
  - episodic_memory: JSONB
  - semantic_memory: JSONB

Go Implementation Sketch:
```go
func LoadCognitiveSystem(ctx ActionContext) error {
    personaID := ctx.Config["persona_id"]
    orchestrationID := ctx.OrchestrationID

    // 1. Load persona personality DNA
    persona := loadPersona(personaID)

    // 2. Load cognitive components
    components := getCognitiveSystems(personaID)
    /*
    Returns map like:
    {
        "perception": {"type": "llm_based", "model": "claude-sonnet-4", ...},
        "working_memory": {"type": "json_state", ...},
        "episodic_memory": {"type": "llm_based", ...},
        "semantic_memory": {"type": "llm_based", ...},
        "knowledge_retrieval": {"type": "llm_based", ...},
        "reasoning_engine": {"type": "llm_based", ...},
        "response_generator": {"type": "llm_based", ...},
        "style_applicator": {"type": "llm_based", ...}
    }
    *\/

    // 3. Get or create instance (loads existing memory or initializes new)
    instanceID := getOrCreatePersonaInstance(personaID, orchestrationID)
    instance := loadPersonaInstance(instanceID)

    // 4. Store in context for subsequent steps
    ctx.PersonaProfile = persona.PersonalityDNA
    ctx.CognitiveComponents = components
    ctx.InstanceID = instanceID
    ctx.WorkingMemory = instance.WorkingMemory
    ctx.EpisodicMemory = instance.EpisodicMemory
    ctx.SemanticMemory = instance.SemanticMemory

    return nil
}
```
*/

-- ============================================================================
-- ACTION: perceive_task
-- ============================================================================
/*
Understands the incoming task and classifies it.

Input:
  - task_input: JSONB (raw task requirements)
  - cognitive_components.perception: implementation config

Output:
  - task_classification: TEXT ('write_content', 'expert_review', 'debate', etc.)
  - task_complexity: DECIMAL (0-1)
  - required_capabilities: TEXT[]
  - subtasks: JSONB[] (if decomposition needed)

Phase 1 Implementation (LLM-based):
```go
func PerceiveTask(ctx ActionContext) error {
    perceiver := ctx.CognitiveComponents["perception"]

    if perceiver.Type == "llm_based" {
        prompt := fmt.Sprintf(`
You are the perception system for %s.

Analyze this task:
%s

Classify the task type (write_content, expert_review, debate, teach, etc.)
Assess complexity (0-1 scale)
List required capabilities
If complex, decompose into subtasks

Return JSON:
{
    "task_classification": "...",
    "task_complexity": 0.0-1.0,
    "required_capabilities": [...],
    "subtasks": [...]
}
        `, ctx.PersonaProfile.Name, ctx.InputData["task_input"])

        result := callLLM(perceiver.Provider, perceiver.Model, prompt)
        parsed := parseJSON(result)

        ctx.TaskClassification = parsed.TaskClassification
        ctx.TaskComplexity = parsed.TaskComplexity
        ctx.RequiredCapabilities = parsed.RequiredCapabilities
        ctx.Subtasks = parsed.Subtasks
    }

    return nil
}
```

Future Implementation (Specialized):
- Could use a fine-tuned classifier model
- Could use rule-based system for common tasks
- Could use multi-stage analysis pipeline
*/

-- ============================================================================
-- ACTION: retrieve_from_memory
-- ============================================================================
/*
Retrieves relevant information from persona's memory systems.

Input:
  - task_classification: TEXT
  - task_input: JSONB
  - episodic_memory: JSONB
  - semantic_memory: JSONB
  - knowledge_base_query: TEXT (optional)

Output:
  - relevant_experiences: JSONB[] (from episodic memory)
  - relevant_facts: JSONB[] (from semantic memory)
  - relevant_knowledge: JSONB[] (from knowledge base)

Phase 1 Implementation (LLM-based):
```go
func RetrieveFromMemory(ctx ActionContext) error {
    // 1. Search episodic memory (past experiences)
    episodicRetriever := ctx.CognitiveComponents["episodic_memory"]

    if episodicRetriever.Type == "llm_based" {
        prompt := fmt.Sprintf(`
Review past experiences and identify relevant ones:

Past experiences:
%s

Current task:
%s

Which experiences are relevant? Return JSON array of experience IDs.
        `, ctx.EpisodicMemory, ctx.InputData["task_input"])

        relevantExperiences := callLLM(episodicRetriever.Provider, episodicRetriever.Model, prompt)
        ctx.RelevantExperiences = filterExperiences(ctx.EpisodicMemory, relevantExperiences)
    }

    // 2. Search semantic memory (learned facts)
    semanticRetriever := ctx.CognitiveComponents["semantic_memory"]

    if semanticRetriever.Type == "llm_based" {
        prompt := fmt.Sprintf(`
What facts from semantic memory are relevant to this task?

Semantic memory:
%s

Task:
%s

Return relevant facts as JSON array.
        `, ctx.SemanticMemory, ctx.InputData["task_input"])

        relevantFacts := callLLM(semanticRetriever.Provider, semanticRetriever.Model, prompt)
        ctx.RelevantFacts = parseJSON(relevantFacts)
    }

    // 3. Search knowledge base (if query provided)
    if ctx.InputData["knowledge_base_query"] != "" {
        knowledgeResults := searchPersonaKnowledge(
            ctx.PersonaID,
            ctx.InputData["knowledge_base_query"],
            nil, // domain
            10   // limit
        )
        ctx.RelevantKnowledge = knowledgeResults
    }

    return nil
}
```

Future Implementation:
- Vector similarity search for episodic memory
- Knowledge graph traversal for semantic memory
- Hybrid retrieval with re-ranking
*/

-- ============================================================================
-- ACTION: reason_and_plan
-- ============================================================================
/*
Plans how to approach the task using reasoning.

Input:
  - task_classification: TEXT
  - task_complexity: DECIMAL
  - subtasks: JSONB[]
  - relevant_experiences: JSONB[]
  - relevant_facts: JSONB[]
  - personality_profile: JSONB

Output:
  - execution_plan: JSONB
  - reasoning_trace: JSONB[]

Phase 1 Implementation (LLM-based):
```go
func ReasonAndPlan(ctx ActionContext) error {
    reasoner := ctx.CognitiveComponents["reasoning_engine"]

    if reasoner.Type == "llm_based" {
        prompt := fmt.Sprintf(`
You are the reasoning system for %s.

Personality:
%s

Task: %s
Complexity: %.2f
Relevant past experiences: %s
Relevant facts: %s

Plan how to approach this task. Consider:
1. What's the best strategy given personality and expertise?
2. What order to do subtasks?
3. What potential issues to watch for?
4. How to apply personality traits appropriately?

Return execution plan as JSON:
{
    "strategy": "...",
    "steps": [
        {"step": 1, "action": "...", "rationale": "..."}
    ],
    "considerations": [...],
    "expected_challenges": [...]
}
        `,
            ctx.PersonaProfile.Name,
            ctx.PersonaProfile,
            ctx.InputData["task_input"],
            ctx.TaskComplexity,
            ctx.RelevantExperiences,
            ctx.RelevantFacts
        )

        plan := callLLM(reasoner.Provider, reasoner.Model, prompt)
        ctx.ExecutionPlan = parseJSON(plan)
    }

    return nil
}
```

Future Implementation:
- Tree-of-thought reasoning
- Multi-agent debate planning
- Formal logic system
- Causal reasoning engine
*/

-- ============================================================================
-- ACTION: generate_response
-- ============================================================================
/*
Generates the actual content/response for the task.

Input:
  - task_classification: TEXT
  - execution_plan: JSONB
  - relevant_knowledge: JSONB[]
  - personality_profile: JSONB

Output:
  - raw_response: TEXT or JSONB

Phase 1 Implementation (LLM-based):
```go
func GenerateResponse(ctx ActionContext) error {
    generator := ctx.CognitiveComponents["response_generator"]

    if generator.Type == "llm_based" {
        // Build comprehensive prompt with all context
        prompt := fmt.Sprintf(`
You are %s.

Background: %s
Expertise: %s
Communication style: %s

Task: %s

Execution plan: %s

Relevant knowledge:
%s

Now execute the task following the plan.
        `,
            ctx.PersonaProfile.Name,
            ctx.PersonaProfile.Biographical,
            ctx.PersonaProfile.Expertise,
            ctx.PersonaProfile.CommunicationStyle,
            ctx.InputData["task_input"],
            ctx.ExecutionPlan,
            ctx.RelevantKnowledge
        )

        response := callLLM(generator.Provider, generator.Model, prompt)
        ctx.RawResponse = response
    }

    return nil
}
```

Future Implementation:
- Fine-tuned model specific to this persona
- Multi-model generation (different models for different content types)
- Iterative refinement with self-critique
*/

-- ============================================================================
-- ACTION: apply_style
-- ============================================================================
/*
Applies persona's unique voice/style to the generated response.

Input:
  - raw_response: TEXT
  - personality_profile.communication_style: JSONB

Output:
  - styled_response: TEXT

Phase 1 Implementation (LLM-based):
```go
func ApplyStyle(ctx ActionContext) error {
    styler := ctx.CognitiveComponents["style_applicator"]

    if styler.Type == "llm_based" {
        commStyle := ctx.PersonaProfile.CommunicationStyle

        prompt := fmt.Sprintf(`
Apply %s's distinctive communication style to this content:

Original content:
%s

Style characteristics:
- Vocabulary level: %s
- Sentence structure: %s
- Formality: %.2f
- Rhetorical devices: %s
- Speech patterns: %s
- Preferred words: %s
- Avoided words: %s

Rewrite to match this style while preserving meaning.
        `,
            ctx.PersonaProfile.Name,
            ctx.RawResponse,
            commStyle.VocabularyLevel,
            commStyle.SentenceStructure,
            commStyle.Formality,
            commStyle.RhetoricalDevices,
            commStyle.SpeechPatterns,
            commStyle.PreferredWords,
            commStyle.AvoidedWords
        )

        styled := callLLM(styler.Provider, styler.Model, prompt)
        ctx.StyledResponse = styled
    }

    return nil
}
```

Future Implementation:
- Style-specific fine-tuned model
- Rule-based post-processing
- Multi-pass refinement
*/

-- ============================================================================
-- ACTION: update_memory
-- ============================================================================
/*
Updates persona's memory systems with learnings from this task.

Input:
  - task_input: JSONB
  - task_output: JSONB (styled_response)
  - execution_plan: JSONB
  - success: BOOLEAN
  - user_feedback: TEXT (optional)

Output:
  - memory_updates: JSONB

Phase 1 Implementation (LLM-based):
```go
func UpdateMemory(ctx ActionContext) error {
    learner := ctx.CognitiveComponents["learning_system"]

    if learner.Type == "llm_based" {
        prompt := fmt.Sprintf(`
Analyze this completed task and extract learnings:

Task: %s
Output: %s
Success: %v
User feedback: %s

What should be remembered?
1. New episodic memory (this experience)
2. Updates to semantic memory (new facts learned)
3. Procedural improvements (better ways to do this)

Return JSON:
{
    "episodic_entry": {
        "task_type": "...",
        "topic": "...",
        "outcome": "...",
        "learned": "..."
    },
    "semantic_updates": [
        {"fact": "...", "confidence": 0.8}
    ],
    "procedural_improvements": [...]
}
        `,
            ctx.InputData["task_input"],
            ctx.StyledResponse,
            ctx.Success,
            ctx.InputData["user_feedback"]
        )

        learnings := callLLM(learner.Provider, learner.Model, prompt)
        updates := parseJSON(learnings)

        // Update instance memory
        updatePersonaMemory(
            ctx.InstanceID,
            nil, // working memory cleared after task
            updates.EpisodicEntry,
            updates.SemanticUpdates,
            nil  // emotional state
        )

        // Log execution for analytics
        logPersonaTaskExecution(
            ctx.InstanceID,
            ctx.PersonaID,
            ctx.TaskClassification,
            ctx.InputData,
            ctx.StyledResponse,
            ctx.CognitiveTrace,
            ctx.Success,
            ctx.QualityScore,
            updates
        )
    }

    return nil
}
```

Future Implementation:
- Automatic knowledge graph updates
- Vector embedding updates
- Active learning (identify knowledge gaps)
*/

-- ============================================================================
-- COMPLETE COGNITIVE WORKFLOW
-- ============================================================================
/*
Example agent workflow using all cognitive actions:

{
    "workflow": {
        "start_step": "initialize",
        "steps": {
            "initialize": {
                "action": "load_cognitive_system",
                "config": {
                    "persona_id": "dr-michael-bimpton"
                },
                "next_step": "perceive"
            },

            "perceive": {
                "action": "perceive_task",
                "next_step": "retrieve"
            },

            "retrieve": {
                "action": "retrieve_from_memory",
                "next_step": "reason"
            },

            "reason": {
                "action": "reason_and_plan",
                "next_step": "generate"
            },

            "generate": {
                "action": "generate_response",
                "next_step": "style"
            },

            "style": {
                "action": "apply_style",
                "next_step": "learn"
            },

            "learn": {
                "action": "update_memory",
                "next_step": "complete"
            },

            "complete": {
                "action": "complete_workflow"
            }
        }
    }
}
*/

-- ============================================================================
-- EVOLUTION PATH: LLM → SPECIALIZED
-- ============================================================================
/*
Phase 1: All LLM-based (Start here)
- perception: LLM prompt
- memory: LLM searches JSON
- reasoning: LLM planning
- generation: LLM content creation
- style: LLM style application
- learning: LLM extraction

Phase 2: Hybrid (Selective upgrades)
- perception: LLM → Fine-tuned classifier
- episodic_memory: LLM → Vector database (Pinecone/Weaviate)
- semantic_memory: LLM → Knowledge graph (Neo4j)
- reasoning: LLM → Tree-of-thought or ReAct
- generation: LLM → Fine-tuned persona model
- style: LLM → Style-specific model
- learning: LLM → Automatic KG updates

Phase 3: Full specialized (Advanced)
- Each component optimized for its function
- Multi-model generation
- Graph-based reasoning
- Active learning systems
- Real-time adaptation

The key: Same interfaces, swappable implementations.
Change component implementation in persona_cognitive_components table,
workflow stays the same.
*/