============================================================================
FUTURE: Shared topic strategy
============================================================================

When we move to shared topics:

1. Pre-create long-lived topics per agent type:
     system.work.page-build-handler
     system.work.content-gap-planner
     system.work.webdesign-agent
     etc.

2. Each spawned agent uses a unique consumer group ID (already set):
     KAFKA_CONSUMER_GROUP = "{type}-group-{agentID[:8]}"
   This ensures only one consumer gets each message.

3. Message routing uses headers:
     - orchestration_id identifies which orchestration a response belongs to
     - reply_to_request_id identifies which awaited_request to match
   These are already in the message headers and used for matching.

4. Changes needed:
   a. setupAgentTopics() → return the shared topic names instead of creating new ones
   b. Stop setting EPHEMERAL_TOPICS → agents won't try to clean up shared topics
   c. Remove topic creation from createTopics() for job.* prefix
   d. Consumer group offset management — auto-reset to latest so new
      consumers don't replay old messages

5. Benefits:
   - No topic creation latency (5-10s saved per spawn)
   - No orphaned topics
   - Fewer Kafka partitions (currently 2 per topic × 2 topics × N spawns)
   - Simpler cleanup — no cleanup needed

6. Risks:
   - Message ordering: currently guaranteed by topic isolation.
      With shared topics, ordering is per-partition. Need to ensure
      messages for the same orchestration go to the same partition
      (use orchestration_id as the Kafka key).
   - Consumer group rebalancing: when a new consumer joins a shared
      topic, Kafka rebalances partitions. Short-lived consumers
      would cause frequent rebalancing. Mitigation: static group
      membership with group.instance.id = agent_id.

The current EPHEMERAL_TOPICS flag makes this transition clean:
  - Today: EPHEMERAL_TOPICS=true, unique topics, cleanup on shutdown
  - Later: remove EPHEMERAL_TOPICS, shared topics, no cleanup needed
  - The agent code doesn't change — it reads topics from env vars either way
