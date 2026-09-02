-- gamedesign.uk — growth_posture = 'hold' (WDS-020, owner decision 5 of 2026-08-31; live in stamp
-- ebf27c60, c2349955d an ancestor, checked 2026-09-02 ~22:10Z).
-- WHY: bugs_open/447 — the improvement loop's tool-suggester planted the sibling's tools on this
-- site; the brief (v1, v2) and GD2 (hosts_tools=FALSE) say it hosts no tool pages. The hand-cancel
-- in SEED_2026-09-02d has to be repeated every loop run; this holds evaluate_tools + add_tool at
-- writeWorkItem in the RECORD shape (deferred, no handler — filed, never skipped, releasable by the
-- one-UPDATE verb stamped on each held row's spec). Set on 0 of 39 sites before this one.
-- Named by the improvement-loop owner as the instrument; applied by this lane on its own site.
\set ON_ERROR_STOP on
BEGIN;
UPDATE sites
   SET settings = jsonb_set(COALESCE(settings,'{}'::jsonb), '{maintenance_profile}',
                  COALESCE(settings->'maintenance_profile','{}'::jsonb)
                  || '{"growth_posture":"hold","growth_posture_set_by":"gamedesign_uk_rebuild lane 2026-09-02","growth_posture_reason":"bugs_open/447 — brief and GD2 say no tool pages; tools live on gamesdesign.co.uk"}'::jsonb,
                  true),
       updated_at = now()
 WHERE id='8f17eb73-fc74-4718-8371-b3125bc4e414';
COMMIT;
SELECT domain, settings->'maintenance_profile'->>'growth_posture' AS posture FROM sites WHERE id='8f17eb73-fc74-4718-8371-b3125bc4e414';
