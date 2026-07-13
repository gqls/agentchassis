I've refactored CallAgentAction into distinct, focused functions with clear responsibilities:
Key Improvements:
1. Main Function Flow

The refactored CallAgentAction now clearly orchestrates 9 distinct steps:

    Extract configuration
    Find target agent
    Extract data
    Determine action
    Build request body
    Create child orchestration
    Build request message
    Send message
    Build result

2. Configuration Extraction

   extractCallConfiguration() - Simple extraction of target agent type and role
   Returns individual values, not a new struct

3. Agent Discovery

   findTargetAgent() - Main function to locate the agent
   findAgentByRole() - Search by role
   findAgentByType() - Search by type
   isStandardAgent() - Check for legacy agents
   Uses minimal TargetAgentInfo struct to hold found agent data

4. Data Extraction (Much Clearer!)

   extractDataForAgent() - Main data extraction logic
   Clear search path hierarchy (most specific to least)
   Explicit logging of what path was used
   Clean fallback logic

5. Request Building

   determineTargetAction() - Clear action selection
   buildRequestBody() - Assembles complete request body
   extractContext() - Handles context extraction separately
   createChildOrchestration() - Simple ID generation

6. Message Construction

   buildCallRequestMessage() - Single responsibility for message creation
   Clear parameter list shows all required inputs
   Proper parent/child orchestration tracking

7. Message Sending

   sendAgentRequest() - Handles marshaling, sending, and tracing
   Single responsibility for communication

8. Result Building

   buildCallResult() - Constructs the return value
   Clear about what information is returned

Key Benefits:

    Debugging Data Issues: The extractDataForAgent() function now clearly shows:
        What paths are searched
        Which path succeeded
        What data was found
        Clear fallback behavior
    Agent Discovery: The agent finding logic is now separated and clear:
        Role-based search
        Type-based search
        Legacy agent handling
    Single Responsibility: Each function does one thing:
        Extract config
        Find agent
        Get data
        Build message
        Send message
    Testability: Each function can be tested independently
    Minimal New Types: Only added TargetAgentInfo to hold found agent data, reusing existing types everywhere else

This refactoring makes it much easier to:

    Track where data comes from
    Debug why certain data isn't found
    Understand the message flow
    Add new features without breaking existing logic

The code is now more maintainable and the data flow is transparent at each step.

--

Priority 1 (Message Prompt)**
```
Parent sends: call_agent with prompt="Test custom prompt"
Expected: Child uses "Test custom prompt"
Log: "Using prompt from incoming message (Priority 1)"
```

**Test 2: Priority 2 (Agent Config)**
```
Parent sends: call_agent WITHOUT prompt field
Agent type: content-creator-hero (has prompt_template in database)
Expected: Child uses hero-specific prompt from database
Log: "Using prompt from agent config (Priority 2)"
```

**Test 3: Priority 3 (Fallback)**
```
Agent has NO prompt_template in database
Workflow step has generic prompt
Expected: Uses workflow step prompt
Log: "Using prompt from workflow step config (Priority 3 - fallback)"

===

--

===


## Part 3: How the Flow Works Now

**Scenario 1: Parent sends custom prompt (Priority 1)**
```
Parent workflow step:
action: call_agent
config:
agent_type: content-creator-hero
prompt: "Write a space-themed hero section for Mars Colony Inc..."

→ CallAgentAction puts prompt in requestBody["prompt"]
→ Child receives message with prompt in body
→ ExecuteLLMPromptAction checks StepConfig.Config["prompt"] ✓ FOUND
→ Uses parent's custom prompt (Priority 1)
```

**Scenario 2: No parent prompt, use agent's default (Priority 2)**
```
Parent workflow step:
action: call_agent
config:
agent_type: content-creator-hero
# NO prompt specified

→ CallAgentAction doesn't add prompt to requestBody
→ Child receives message without prompt
→ ExecuteLLMPromptAction checks StepConfig.Config["prompt"] ✗ NOT FOUND
→ Checks agentConfig["prompt_template"] ✓ FOUND
→ Uses agent's database-defined prompt (Priority 2)
```

**Scenario 3: Fallback (Priority 3)**
```
Agent has NO prompt_template in database
→ ExecuteLLMPromptAction checks all priorities
→ Falls back to workflow step config or generic
→ Uses workflow fallback (Priority 3)
