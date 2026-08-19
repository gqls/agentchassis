-- 487 — seed the bindings-repair batch (bugs_open/324; parent 283 §14).
--
-- ⚠ _HOLD: ORDERING-CRITICAL, TWICE — and BOTH preconditions are ENFORCED, not prose
-- (council round 6, guardian + debug_historian):
-- (1) fix_type 'repair_instance_scope_bindings' exists only in chassis builds carrying the
--     2026-08-19 code. RUNBOOK §1 digest-verify first, THEN ask the pod binary itself, with
--     a present AND an absent control in the same breath (the documented discipline):
--       POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1)
--       kubectl -n ai-persona-system exec ${POD#pod/} -- grep -ac "repair_instance_scope_bindings" /proc/1/exe   # MUST be >0
--       kubectl -n ai-persona-system exec ${POD#pod/} -- grep -ac "no_such_symbol_283_xyzzy"      /proc/1/exe   # MUST be 0
--     (both >0 or both 0 = your probe is broken, not the binary.)
-- (2) 486 (the judged branch) must be applied FIRST: 5 of these rows will refuse to the
--     judged pool, and without 486 a refusal completes as a quiet no-op instead of failing
--     to a human. The DO block below RAISES if the fixer's workflow lacks 486's steps, so a
--     manual apply-order mistake fails loudly inside the transaction.
-- Rename away from _HOLD when applying.
--
-- One item per CONVERTED row, derived at apply time (never a pasted id list — the corpus
-- drifts daily). The repair arm no-ops loudly-but-cleanly on sound rows, so seeding all
-- converted rows costs ~37 no-ops and buys a fully derived, re-runnable batch. Priority and
-- severity are raised for rows whose placements are SERVING the converted bytes (those are
-- live-broken today — 14 rows / 15 placements measured 2026-08-19).
--
-- Progress:   SELECT status, count(*) FROM site_work_items
--             WHERE item_type='instance_scope_conversion' AND created_by='324-bindings-repair-seed'
--             GROUP BY 1;
-- Done-check: cmd/instanceaudit <converted-export> --bindings exits 0 (the 5 judged rows
--             leave via the judged pipeline, not this batch).
BEGIN;

-- Precondition (enforced): 486 must already be live in the fixer's workflow.
DO $$
DECLARE cfg jsonb;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type='component-template-fixer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF cfg IS NULL OR NOT (cfg #> '{workflow,steps}') ?& array['check_scope_route','judged_refusal'] THEN
    RAISE EXCEPTION '487: migration 486 (judged branch) is NOT applied — without it a judged refusal completes as a quiet no-op. Apply 486 first.';
  END IF;
END $$;

INSERT INTO site_work_items
  (site_id, source, pipeline, item_type, severity, summary, priority, handler_agent, status, created_by, spec, item_key)
SELECT
  (SELECT p2.site_id FROM page_components pc2 JOIN pages p2 ON p2.id=pc2.page_id
    WHERE pc2.component_id = c.id ORDER BY (pc2.rendered_html LIKE '%id="c-' || c.function || '-%') DESC, p2.site_id LIMIT 1),
  'manual', 'build', 'instance_scope_conversion',
  CASE WHEN serving.n > 0 THEN 'high' ELSE 'medium' END,
  'Repair instance-scope bindings (324): ' || c.function ||
    CASE WHEN serving.n > 0 THEN ' [SERVING BROKEN on ' || serving.n || ' page(s)]' ELSE '' END,
  CASE WHEN serving.n > 0 THEN 30 ELSE 60 END,
  'component-template-fixer', 'triaged', '324-bindings-repair-seed',
  jsonb_build_object(
    'fix_type', 'repair_instance_scope_bindings',
    'component_id', c.id::text,
    'category', 'seam',
    'note', 'bugs_open/324: pass-5 binding repair; arm no-ops on sound rows, refuses judged rows to needs_human_review (post-484)'),
  'bindings-repair-' || c.id::text
FROM content_components c
CROSS JOIN LATERAL (
  SELECT count(*) AS n FROM page_components pc JOIN pages p ON p.id=pc.page_id
  WHERE pc.component_id = c.id AND pc.rendered_html LIKE '%id="c-' || c.function || '-%'
) serving
WHERE c.is_active
  AND c.html_template LIKE '%InstanceID%'
  AND EXISTS (SELECT 1 FROM page_components pc3 WHERE pc3.component_id = c.id)
  AND NOT EXISTS (
    SELECT 1 FROM site_work_items w
    WHERE w.item_key = 'bindings-repair-' || c.id::text
      AND w.status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled'));

DO $$
DECLARE n int; hi int;
BEGIN
  SELECT count(*), count(*) FILTER (WHERE severity='high')
    INTO n, hi FROM site_work_items WHERE created_by='324-bindings-repair-seed';
  IF n < 30 THEN
    RAISE EXCEPTION '487: only % items seeded — expected one per converted+placed row (69 at design time; derive, but <30 means the predicate broke)', n;
  END IF;
  IF hi < 5 THEN
    RAISE EXCEPTION '487: only % high-severity items — the serving-broken detection predicate broke (14 rows measured 2026-08-19)', hi;
  END IF;
  RAISE NOTICE '487: seeded % repair items (% serving-broken, priority 30)', n, hi;
END $$;

COMMIT;
