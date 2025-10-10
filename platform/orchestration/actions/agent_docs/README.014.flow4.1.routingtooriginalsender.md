The key information for routing back to the original sender is stored in the __execution_context__ in CollectedData when the request first arrives. This contains the original responses_topic and request_id.
Here's how the flow should work with the reverted ProcessResponse:

Parent (Generic) calls Child (Calculator):

Parent stores child's request_id in AwaitedRequests
Child receives request with responses_topic: system.agent.generic.responses
Child stores this in its __execution_context__


Child processes and completes:

Child's CompleteWorkflowAction extracts parent info from __execution_context__
Sends response to system.agent.generic.responses with parent's request_id


Parent receives response:

ProcessResponse finds parent's state via FindByAwaitedRequestID
Stores response in parent's CollectedData
When parent completes, it has the original request info in its own __execution_context__



The critical fix needed is ensuring the parent's CompleteWorkflowAction can find the ORIGINAL client request info. This should already be in the parent's __execution_context__. Let me verify: