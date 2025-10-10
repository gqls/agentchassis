system already has an elegant way to handle versioned workflows. It's built around the concept of an AgentGroup. Here's the logic, based on your files:

    Stored and Versioned: You have an agent_groups table in your database. Each row in this table represents a specific, versioned team of agents designed for a task (e.g., website-builder, content-team). Each of these groups has its own orchestration_workflow stored directly in its database row.

    Evolutionary Versions: The code in evolution.go is designed to create new versions of these groups. When a mutation is applied (e.g., adding a new "validator" agent to the team), it doesn't overwrite the old group. Instead, it creates a new group record with an incremented version number and a parent_id pointing back to the original. This gives you a complete, auditable history of how a workflow has evolved.

    Discovery by group_type: The entry point to this system is the FindBestGroup function in agent_discovery.go. When an action like spawn_group is triggered with a group_type (like "website-builder"), this function queries the database to find the best available version of that group, ordered by performance, usage, and version number.

In short: you don't send a workflow ID. You request a capability (the group_type), and the system intelligently selects the best version of the workflow to execute.



# Define your variables
CORRELATION_ID=$(uuidgen)
# ... etc.

# Send a message to spawn the group
kubectl -n kafka run -i --rm kcat-producer \
--image=edenhill/kcat:1.7.1 \
--restart=Never -- \
kcat -P \
-b personae-kafka-cluster-kafka-bootstrap:9092 \
-t system.agent.generic.requests \
-H correlation_id=$CORRELATION_ID \
# ... other headers ...
-H action=orchestrate <<EOF
{
"action": "spawn_group",
"config": {
"group_type": "welcome-message-generator"
},
"input_data": {
"business_type": "pizza restaurant",
"business_name": "Test Pizza"
}
}
EOF


--

Agent Group Behavior

The system is designed around a powerful concept of versioned, discoverable agent teams.

    Triggered by Type: A workflow is initiated not by sending the workflow itself, but by sending a message with an action of spawn_group and specifying a group_type like "website-builder".

    Discovery of the Best Version: The GroupDiscovery service receives this group_type. It then queries the agent_groups table to find the best-performing and most recent version of that group to execute the task.

    Execution of Stored Workflow: Once the best group is selected, the system retrieves its associated orchestration_workflow directly from the database and begins executing it.

    Evolution and Versioning: The EvolutionService is designed to create new versions of these groups. When a group is "mutated" (e.g., a new agent is added to the team), it creates a new row in the database with an incremented version and a parent_id linking to the old version. This provides a complete and auditable history of how your workflows change over time.


Combining Overrides with Agent Groups

It can have both systems working together, an inline workflow override in the message and the group system. The key is to establish a clear order of priority in your selectWorkflow function within processor.go.

Here is the recommended priority hierarchy:

    Highest Priority: Inline Workflow Override

        Check: Does the incoming message's config block contain a full workflow definition?

        Action: If yes, use this workflow immediately. This is for ephemeral, single-use tasks that don't need to be versioned.

    Second Priority: Group-Based Workflow

        Check: If there's no inline workflow, does the message's action field specify spawn_group?

        Action: If yes, extract the group_type from the message, use FindBestGroup to get the appropriate versioned workflow from the database, and use that.

    Fallback: Default Agent Workflow

        Check: If neither of the above conditions is met.

        Action: Fall back to the agent's hardcoded default workflow from the agent_definitions table. This is the original behavior.

This approach can run stable, versioned business processes by default but still override them with dynamic instructions when needed.