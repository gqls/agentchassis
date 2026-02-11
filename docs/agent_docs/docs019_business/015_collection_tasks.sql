-- collection_tasks_unique_pending_index.sql
--
-- Prevents duplicate pending tasks for the same business+task_type.
-- If a promote run is re-executed, the ON CONFLICT DO NOTHING in
-- promote_candidates will silently skip already-queued businesses.
--
-- Partial index: only covers status='pending' rows, so completed/in_progress
-- tasks don't block new ones (e.g. re-verification after a previous run).

CREATE UNIQUE INDEX IF NOT EXISTS idx_collection_tasks_unique_pending
    ON business_intel.collection_tasks (business_id, task_type)
    WHERE status = 'pending';