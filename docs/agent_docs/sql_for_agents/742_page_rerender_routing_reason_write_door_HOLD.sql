-- 742_page_rerender_routing_reason_write_door_HOLD.sql
--
-- RFC_062 PHASE 3, THE WRITE DOOR (OWNER RULING D3). The other half of 741.
--
-- ⚠⚠ _HOLD — same release condition as 741: the 404 lane's co-sign (OWNER RULING D2).
-- ⚠ That lane is NOT abandoned, whatever earlier revisions of these files said: last own
-- commit 2026-09-02 16:24Z, and its own r4 verdict is APPROVED-BUT-UNREAD since 16:33Z the
-- same day. Corrected 2026-09-03 (night) — the claim came from directory mtimes, not git log.
-- APPLY ORDER: 741 (the read door) THEN this file. Both intermediate states are safe —
-- read door alone REFUSES a bad key at the gate; write door alone REJECTS the INSERT —
-- but stopping after either one leaves the pair half-built, and the producer class this
-- file exists for (hand-written migrations minting page_rerender items) is the one no Go
-- guard and no gate condition can see at authoring time.
--
-- ── WHY THIS IS ITS OWN FILE, AND IT WAS NOT AT FIRST ─────────────────────────
--
-- It was step 5 of 741 until the council's `guardian` seat objected [medium], round
-- 56047b18: "ALTER TABLE ... ADD CONSTRAINT ... NOT VALID on site_work_items acquires an
-- ACCESS EXCLUSIVE lock on the WHOLE table regardless of the CHECK's item_type-scoped
-- predicate. This table is the dispatch queue for every pipeline (build, diagnose,
-- report, improvement, etc.), not just page-rerender ... bundling it with a
-- single-pipeline workflow-step insert removes the option to schedule the riskier DDL
-- separately."
--
-- CONCEDED — and it is a better argument than the one 741 had pre-rebutted. That
-- rebuttal was against splitting in order to "get the rest in", i.e. shipping a gate and
-- never returning for the write door; it did not address SCHEDULING. Bundled, a lock
-- timeout on this DDL rolls back the workflow edit too, and the two changes want
-- different windows: one touches a single agent_definitions row, this one takes a
-- table-wide lock on the estate's busiest queue.
--
-- ── THE LOCK, MEASURED BOTH WAYS ──────────────────────────────────────────────
--
-- `NOT VALID` skips the table SCAN. It does NOT skip the LOCK: ADD CONSTRAINT still needs
-- ACCESS EXCLUSIVE, which waits for every in-flight transaction touching the table.
-- MEASURED 2026-09-03, the SAME statement, both outcomes: one dry run acquired it in
-- **2 ms**; an earlier one was **still waiting after 2 minutes** and had to be killed. So
-- a fast reading is not evidence that it is fast — the wait is probabilistic, and one
-- measurement certifies whichever answer you happened to get.
--
-- Worse than the wait: a QUEUED ACCESS EXCLUSIVE request blocks every subsequent reader
-- and writer of `site_work_items` behind it in the lock queue. An unbounded wait here
-- does not merely delay this migration — it stalls the whole work-item pipeline for as
-- long as it waits. Hence `lock_timeout`, and hence the instruction NOT to raise it.
--
-- ── WHAT THE CONSTRAINT SAYS ──────────────────────────────────────────────────
--
-- Scoped to `item_type='page_rerender'` — every other item_type is untouched, which is
-- what makes this safe for the other pipelines sharing the table. A NULL `spec` or a NULL
-- `routing_reason` SATISFIES it: that is the annotation-only case (an item carrying prose
-- in `spec.reason` and no routing key), legal forever under owner ruling D4, which is also
-- why nothing here validates `spec.reason`.
--
-- NOT VALID, so history is not scanned and the ALTER cannot fail on an existing row.
-- VALIDATE is a separate, later, explicit step (D3) — the census that says whether it
-- would pass is section B of `741_..._HOLD_VERIFY.sql`, and it read **0** rows over the
-- WHOLE table (every status, not the pending window) on 2026-09-03. VALIDATE takes only
-- SHARE UPDATE EXCLUSIVE, so concurrent reads and writes continue; it fails on the first
-- violating row, which is why you count before you run it.
--
-- Companion: 742_..._HOLD_ROLLBACK.sql (one DROP CONSTRAINT). The post-apply checks live
-- in 741's VERIFY companion, which deliberately covers both files.

BEGIN;

-- ⚠ SEE THE HEADER. Failing fast is the safe direction; re-run in a quiet window rather
-- than raising this. One transaction, so a timeout changes nothing.
SET LOCAL lock_timeout = '5s';

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_constraint
              WHERE conname='chk_page_rerender_routing_reason_vocabulary'
                AND conrelid='site_work_items'::regclass) THEN
    RAISE EXCEPTION '742 REFUSED: constraint chk_page_rerender_routing_reason_vocabulary already exists — this file (or a hand edit) has already been applied; do not stack';
  END IF;

  -- The write door is the weaker half on its own, but applying it BEFORE 741 is still
  -- safe, so this is a NOTICE and not a refusal. Stopping here is the failure mode.
  IF (SELECT default_config #> '{workflow,steps,check_routing_key_known}' FROM agent_definitions
        WHERE type='page-rerender' AND is_active
          AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) IS NULL THEN
    RAISE NOTICE '742: 741 is NOT applied yet (no check_routing_key_known step). That is safe — this constraint stands alone — but the pair is incomplete until 741 lands. Do not stop here.';
  END IF;
END $$;

-- THE WRITE DOOR. Vocabulary PASTED from livespec.RerenderSectionReasonNames(); a sixth
-- value means regenerating this list and the gate clause in ONE migration, as 741's
-- header sets out.
ALTER TABLE site_work_items
  ADD CONSTRAINT chk_page_rerender_routing_reason_vocabulary
  CHECK (
    item_type <> 'page_rerender'
    OR spec->>'routing_reason' IS NULL
    OR spec->>'routing_reason' IN ('image_landed', 'section_data_resolved', 'cta_links_stale', 'template_changed', 'literal_markdown')
  ) NOT VALID;

-- VERIFY in a DO block: ON_ERROR_STOP ignores a non-empty SELECT result, so a
-- SELECT-based verify cannot stop the COMMIT (LANDMINES / RFC_006).
DO $$
DECLARE
  def text;
  validated boolean;
BEGIN
  SELECT pg_get_constraintdef(oid), convalidated INTO def, validated
    FROM pg_constraint
   WHERE conname='chk_page_rerender_routing_reason_vocabulary'
     AND conrelid='site_work_items'::regclass;

  IF def IS NULL THEN
    RAISE EXCEPTION '742 VERIFY FAILED: the constraint is absent after the ALTER';
  END IF;
  IF validated THEN
    RAISE EXCEPTION '742 VERIFY FAILED: the constraint is VALIDATED; it must be added NOT VALID so history is unscanned (D3 validates separately, after the census)';
  END IF;

  -- The scope is the whole safety argument for a table five other pipelines share:
  -- without the item_type guard this would refuse THEIR inserts too.
  IF def NOT LIKE '%item_type%' THEN
    RAISE EXCEPTION '742 VERIFY FAILED: the constraint is not scoped by item_type — it would apply to every pipeline sharing site_work_items. Definition: %', def;
  END IF;
  -- All five vocabulary values must be present, or the door refuses legitimate traffic.
  IF def NOT LIKE '%image_landed%' OR def NOT LIKE '%section_data_resolved%'
     OR def NOT LIKE '%cta_links_stale%' OR def NOT LIKE '%template_changed%'
     OR def NOT LIKE '%literal_markdown%' THEN
    RAISE EXCEPTION '742 VERIFY FAILED: the vocabulary in the constraint is incomplete. Definition: %', def;
  END IF;

  RAISE NOTICE '742 OK: write door up, NOT VALID, scoped to page_rerender. Next: section B of 741_..._HOLD_VERIFY.sql (it read 0 rows that would fail), then ALTER TABLE site_work_items VALIDATE CONSTRAINT chk_page_rerender_routing_reason_vocabulary;';
END $$;

COMMIT;
