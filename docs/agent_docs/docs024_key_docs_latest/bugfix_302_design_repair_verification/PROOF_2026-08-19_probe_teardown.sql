-- Teardown for the bugs_open/302 proof harness. Removes ONLY what the setup created,
-- matched on the 'proof-302' markers it stamped. Deliberately narrow: no wildcard on
-- item_type or site_id, so it cannot reach a real row even if run twice.
--
-- Both deletes are safe to repeat (0 rows the second time). site_work_items carries only
-- an updated_at trigger — no archive-on-delete trigger — so a delete leaves nothing behind.

DELETE FROM site_work_items
WHERE item_key LIKE 'proof-302-%' AND source = 'proof-302' AND created_by = 'proof-302';

DELETE FROM agent_definitions WHERE type = 'proof-302-completion-gate-probe';

-- Both must read 0.
SELECT (SELECT count(*) FROM site_work_items WHERE item_key LIKE 'proof-302-%') AS items_left,
       (SELECT count(*) FROM agent_definitions WHERE type LIKE 'proof-302%')    AS agents_left;
