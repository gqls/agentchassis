. Pinpointing the Failure

Here's the journey of the responses_topic and where it gets dropped:

    Source (kubectl): You send the header string ...responses_topic:system.agent.generic.responses\t...

    Ingestion (agent.go): The agent.processMessage function receives the raw Kafka message.

    Parsing (messaging/context.go - Implied): processMessage calls messaging.NewMessageContext(...) to parse the raw message. This is the point of failure. The custom key:value,key:value header format is not being correctly parsed into the structured msgCtx.ExecutionContext.ResponsesTopic field. The field is left as an empty string.

    Storage (coordinator.go): The coordinator.StartNewWorkflow function is called with the MessageContext containing the now-empty ResponsesTopic. It then saves this empty value into the permanent orchestration state under the key __initial_responses_topic__.

    Retrieval (workflow_actions.go): Much later, CompleteWorkflowAction retrieves the empty value from the state, leading to the warning you saw.