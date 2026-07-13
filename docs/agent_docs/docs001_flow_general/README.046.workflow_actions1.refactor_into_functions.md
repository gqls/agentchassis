Looking at CompleteWorkflowAction, you're right that we can simplify it. Every orchestration should complete the same way - by sending a response to whoever is waiting for it. The distinction between "root" and "child" shouldn't matter if our message structure is consistent.

Let me refactor CompleteWorkflowAction with clear, focused functions:
Refactored CompleteWorkflowAction with Clear Responsibilities
Code

I've refactored CompleteWorkflowAction with clear, focused functions that handle distinct responsibilities:
Key Improvements:
1. Simplified Main Flow

The refactored CompleteWorkflowAction now has 5 clear steps:

    Extract final result
    Determine response target
    Build response message
    Send response
    Return local result

2. Removed Root vs Child Distinction

As you correctly pointed out, we don't need to know if it's "root completing" - every orchestration completes the same way by responding to whoever is waiting. The logic is now unified.
3. Clear Functions with Single Responsibilities

   extractFinalResult() - Determines what data to return as the final result
   Checks multiple locations in priority order
   Filters out internal data (keys starting with __)
   determineResponseTarget() - Figures out who's waiting for the response
   Returns a ResponseTarget struct with all needed info
   Returns nil if no response is needed (some workflows might not respond)
   determineResponseTopic() - Finds where to send the response
   Checks multiple sources in priority order
   Environment variable is most reliable for spawned agents
   findOriginalRequestID() - Locates the request we're responding to
   Checks parent request ID, stored requests, context
   Clear priority order
   buildCompletionResponse() - Constructs the response message
   Properly sets all headers
   Distinguishes between "my" orchestration and target orchestration
   sendCompletionResponse() - Handles the actual sending
   Marshaling, headers, and production
   Doesn't fail the workflow if send fails

4. ResponseTarget Structure

Introduced a simple struct to hold response targeting information, making the code clearer.
5. Bonus: FailWorkflowAction

Also added a refactored FailWorkflowAction using the same patterns for consistency.
Key Design Decisions:

    Graceful Degradation: If we can't send the response, we still complete locally and return success. This prevents cascading failures.
    Explicit Priority Orders: Each "find" function has clear priority for where to look for data.
    Consistent Message Structure: The response message is built the same way regardless of whether it's a child or root orchestration.
    Environment Variable Priority: PARENT_RESPONSES_TOPIC is given highest priority as it's the most reliable for spawned agents.

This refactoring makes the workflow completion logic much clearer and easier to debug. Each function has a single, clear purpose and the overall flow is straightforward.

Would you like me to also refactor executeLocalAction next? It has quite a bit of complexity that could benefit from similar decomposition.