Dispatch loop spawns handler → waits for response
↓ 120s of no messages (handler is doing LLM work)
↓ Idle timeout kills dispatch loop pod → Job completes → CronJob cleans it up
↓ Handler finishes → sends response to dead dispatch loop topic
↓ Item stays claimed forever
↓ claimed-item-timeout isn't firing (pre_query returns no rows)
↓ Pipeline stalls