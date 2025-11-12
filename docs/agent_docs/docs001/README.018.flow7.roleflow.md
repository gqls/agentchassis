The flow for the role would be:

When agent is spawned: Role comes from the spawn message body
Agent stores it: The spawned agent saves the role in its Agent.Role field
Agent includes in identity: When sending messages, GetIdentity() includes the role
Appears in ExecutionContext: The role is in ExecutionContext.Sender.Role
Available in actions: Actions access it via params.ExecutionContext.Sender.Role

However, for this to work, you need to ensure the spawned agent actually extracts and stores its role when it receives the initialization message. This happens in the agent's message processor when handling the "initialize" action.