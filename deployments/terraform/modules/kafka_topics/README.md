Standardize on system.agent.<agent-type>.process for primary agent requests.

Standardize on system.adapter.<adapter-name>.requests (or similar) for adapter requests.

Standardize on system.responses.<agent-or-adapter-type> for responses.

Standardize on dlq.<agent-or-adapter-type> for dead-letter queues.

Use system.tasks.<agent-type> for specific work delegation if distinct from direct requests. (If content-creator is mostly directly requested, system.tasks.content-creator might be redundant unless you have a specific workflow that uses it).

Ensure all topic definitions in modules/kafka_topics/main.tf match the RequestTopic, ResponsesTopic, etc., constants in your Go code and the TopicDefinition logic in platform/kafka/topic_manager.go.