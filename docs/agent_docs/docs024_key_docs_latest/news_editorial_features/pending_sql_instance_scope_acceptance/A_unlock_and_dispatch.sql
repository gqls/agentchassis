-- news_editorial_features lane — 2026-08-25
-- ACCEPT the 283/RFC_034 instance-scope conversion on the two locked
-- evidence-timeseries instances (owner ruling, 2026-08-25).
--
-- > **CORRECTED 2026-08-25 (v2): THIS SCRIPT DID NOT UNLOCK ANYTHING.**
-- > v1 cleared `lock_type` and `locked_by` and left `locked_at` SET. But the
-- > canonical "may automation write this row?" predicate is
-- > `AgentWritableSQLFor` (platform/orchestration/datahelpers/chrome_render_inputs.go:91):
-- >     (locked_at IS NULL OR (lock_type='timed' AND lock_expires_at IS NOT NULL
-- >                            AND lock_expires_at < NOW()))
-- > so **`locked_at` is the load-bearing column** — `lock_type`/`locked_by` are
-- > classification and messaging only. Worse, `classifyComponentLock`
-- > (lock_helpers.go:100) treats a row with `locked_at` set and NO `lock_type`
-- > as HARD/permanent ("conservative — never silently overwrite what we can't
-- > classify"), so v1 left the rows locked and UNCLASSIFIED.
-- > Caught 2026-08-25 by running it: both re-dispatched items were claimed, ran,
-- > and returned `{"skipped":true,"locked":true,"success":true}` with the reason
-- > `is locked by ""` (the empty string is NULL `locked_by` read back). Nothing
-- > was written — stored bytes unchanged at 7,742 / 7,182.
-- > The framework's own admin unlock clears ALL FOUR columns
-- > (internal/core-manager/admin/page_admin_handlers.go:450); this now matches it.
--
-- Dry-run gate PASSED before this was written: the local harness reproduces
-- both stored rows BYTE-FOR-BYTE from the v1 template + live content_data, and
-- rendering the LIVE template with InstanceID bound changes ONLY the id:
--   robot-demand-step-change : evidence-timeseries-ifr          -> c-evidence-timeseries  (-2 B)
--   darts-calendar-density   : evidence-timeseries-pdc-calendar -> c-evidence-timeseries  (-11 B)
--
-- Idempotent. Safe to re-run: every statement is guarded and RETURNS what it did,
-- so "0 rows" is diagnosable rather than ambiguous.
--
-- AFTER this runs the two rows are GENUINELY UNLOCKED (v1 never was). Re-lock
-- with B_relock.sql as soon as verification passes — do not leave them unlocked
-- overnight. B_relock.sql restores `locked_at` to its captured original, NOT
-- NOW(), so the lock's provenance is not rewritten.

BEGIN;

-- 0. Record the pre-unlock lock state, so the restore in B_relock.sql can be
--    checked against what was actually here rather than against a handoff.
--    Captured 2026-08-25 before this ran:
--      d344585f… locked_at 2026-08-19 15:17:43.126181+00
--      ea6b4ca7… locked_at 2026-08-20 16:59:30.895515+00
--      both: locked_by 'news_editorial_features-lane', lock_type 'permanent',
--            lock_expires_at NULL
SELECT id, slot_name, locked_at, locked_by, lock_type, lock_expires_at
  FROM page_components
 WHERE id IN ('d344585f-6f79-4a18-a1fe-3116b68a4c52',
              'ea6b4ca7-7717-4e29-ae1c-88844040b0d2');

-- 1. Unlock the two instances so apply_section_edit can write.
--    ALL FOUR columns — clearing fewer does not unlock (see the correction above).
--    Guarded on the exact lock we own: this cannot touch another lane's lock.
UPDATE page_components
   SET locked_at = NULL, locked_by = NULL, lock_type = NULL, lock_expires_at = NULL
 WHERE id IN ('d344585f-6f79-4a18-a1fe-3116b68a4c52',
              'ea6b4ca7-7717-4e29-ae1c-88844040b0d2')
   AND locked_by = 'news_editorial_features-lane'
RETURNING id, slot_name,
          (locked_at IS NULL OR (lock_type = 'timed' AND lock_expires_at IS NOT NULL
                                 AND lock_expires_at < NOW())) AS agent_writable_now;

-- 2. Re-dispatch the delivery the lock refused on 2026-08-23.
--    Shape copied verbatim from the two refused items (53a7fdd0…, baa9b873…):
--    same source/handler_agent/priority/approval_mode/pipeline/spec/item_key.
--    Created at 'triaged' because the dispatcher claims
--    status IN ('triaged','approved')  [load_work_item_actions.go:701].
--    created_by names THIS lane, not component-template-fixer — we are the
--    author of this re-drive and the record should say so.
INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, spec,
   priority, handler_agent, status, created_by, approval_mode, pipeline, item_key)
SELECT v.site_id::uuid, 'side_effect', 'section_edit', 'medium', v.summary, v.spec::jsonb,
       60, 'section-editor', 'triaged', 'news_editorial_features-lane', 'auto', 'build', v.item_key
  FROM (VALUES
    ('00ff3af5-dad8-4770-9f70-3edc267a3c92',
     'Deliver template fix to owned page via section-editor: robot-demand-step-change',
     '{"reason":"template_changed","page_id":"0c297014-1d0d-444d-b673-a75f1ee706fc","edit_type":"content_edit","page_name":"robot-demand-step-change","slot_name":"evidence-timeseries-ifr","component_id":"fb870e82-2f01-46e4-9552-e764515e18d8","field_updates":{}}',
     'section_edit_tplfix_0c297014-1d0d-444d-b673-a75f1ee706fc'),
    ('5fe8785b-223d-41a3-88ee-c07187622381',
     'Deliver template fix to owned page via section-editor: darts-calendar-density',
     '{"reason":"template_changed","page_id":"cccc84fa-4b17-4dfb-aecd-fb776d220c33","edit_type":"content_edit","page_name":"darts-calendar-density","slot_name":"evidence-timeseries-pdc-calendar","component_id":"fb870e82-2f01-46e4-9552-e764515e18d8","field_updates":{}}',
     'section_edit_tplfix_cccc84fa-4b17-4dfb-aecd-fb776d220c33')
  ) AS v(site_id, summary, spec, item_key)
 WHERE NOT EXISTS (
   SELECT 1 FROM site_work_items w
    WHERE w.item_key = v.item_key
      AND w.status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled')
 )
RETURNING id, site_id, status, item_key, created_at;

COMMIT;

-- 3. Read the state back raw, so an empty RETURNING above can be told apart
--    from "already applied". agent_writable MUST be t on both rows — that is
--    the check v1 lacked, and its absence is why v1's failure was invisible
--    until the dispatcher had already spent both work items.
SELECT pc.id, pc.slot_name, pc.locked_at, pc.locked_by, pc.lock_type,
       (pc.locked_at IS NULL OR (pc.lock_type = 'timed' AND pc.lock_expires_at IS NOT NULL
                                 AND pc.lock_expires_at < NOW())) AS agent_writable,
       pc.updated_at
  FROM page_components pc
 WHERE pc.id IN ('d344585f-6f79-4a18-a1fe-3116b68a4c52',
                 'ea6b4ca7-7717-4e29-ae1c-88844040b0d2');

SELECT id, status, created_by, created_at, left(summary,60) AS summary
  FROM site_work_items
 WHERE item_key IN ('section_edit_tplfix_0c297014-1d0d-444d-b673-a75f1ee706fc',
                    'section_edit_tplfix_cccc84fa-4b17-4dfb-aecd-fb776d220c33')
 ORDER BY created_at DESC;
