

// ============================================================================
// KEY DESIGN POINTS
// ============================================================================
//
// 1. AGENTS NEVER CLEAN UP TOPICS
//    - Not on shutdown, not on idle timeout, not on error
//    - The agent creates topics at spawn time and forgets about them
//    - This means every agent is free to spawn children without worrying
//      about topic lifecycle
//
// 2. EXTERNAL CLEANUP IS CONSERVATIVE
//    - Only deletes topics with no matching running pod
//    - Runs every 10 minutes (CronJob schedule)
//    - Topics with running consumers are always kept
//    - Kafka 7-day retention is the backstop
//
// 3. FUTURE: SHARED TOPICS ELIMINATE THE PROBLEM
//    - When we move to shared topics per agent type, there are no
//      per-spawn topics to clean up at all
//    - The CronJob section becomes a no-op and can be removed
//
// 4. IDLE TIMEOUT STILL WORKS
//    - Agents still exit via idle timeout — they just don't clean up topics
//    - The pod exits, K8s Job completes, CronJob cleans up the Job
//    - Topics persist until the CronJob's next pass confirms no pods need them
//
// 5. TOPIC ACCUMULATION IS BOUNDED
//    - With idle timeout, agents exit in minutes, not hours
//    - CronJob runs every 10 min → topics live at most ~20 min after pod exit
//    - At peak: maybe 50-100 orphan topics between cleanup cycles
//    - Each topic is a few KB — negligible resource usage
