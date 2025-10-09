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


