https://claude.ai/chat/0527c72a-f629-4503-ae6f-e24920633e1f

The worker's probe status_command encodes exactly this taxonomy, and it lines up with the plan:

ALIVE — pgrep finds 02_train_llama_3_3_70b or /workspace/run.sh (still training) → reset_streak.
DONE_OK — RUN_SH_DONE in train.log and /workspace/adapter_out/adapter_config.json exists → mark_complete → decommission.
DONE_FAIL — RUN_SH_FATAL in train.log → mark_failed → decommission.
GONE_UNKNOWN — process gone, no DONE/FATAL marker (crash/OOM/reap without a marker) → bump_streak, and mark_failed once consecutive_unreachable_probes hits the threshold (3).