-- 646 — the render-audit COVERAGE CURSOR: table + the opt-in flip (bugs_open/394).
--
-- ⚠ _HOLD: THIS MUST NOT BE APPLIED BEFORE THE CHASSIS IMAGE CARRYING THE CODE
-- HAS ROLLED. `rotate_coverage: true` against a binary that does not know the key
-- is harmless (the action reads it with a comma-ok assert and the spec sets
-- neither CheckConfig nor StrictConfig, so an unknown key is ignored, not
-- warned) — but `rotate_coverage: true` against a binary that DOES know it, with
-- no `render_audit_page_cursor` table, makes the first cursor read fail. The
-- action degrades to the prefix window and logs, so even that is not damage; it
-- is a silently inert feature, which is worse to diagnose than a loud one.
-- The `_HOLD` suffix matches run-migrations.sh's SIDECAR_RE, so the runner LISTS
-- this file and SKIPS it. Drop the suffix or apply it by hand, after the roll.
--
-- ── THE DEFECT ─────────────────────────────────────────────────────────────
--
-- `request_render_audit` selects a site's live pages
-- `ORDER BY COALESCE(nav_order,999), name` and takes the first `max_pages`. The
-- SAME prefix, every run. Pages past the cap were never audited and never would
-- be. `bugs_closed/242` made that loud (RENDER_AUDIT_TRUNCATED) and raised the
-- cap 25 -> 60 as a stated MITIGATION (migration 392); the mitigation has been
-- outgrown. MEASURED 2026-08-26: webdesign.co.uk is 146 live pages, 60 audited,
-- 86 never — and the tail is a CLASS rather than a count. Its nav_order bands
-- are 0..90 (6 nav pages), 100 (94 tools, alphabetical), 200 (48 tool-*-guide),
-- 201 (1). The cap cuts between `tool-head-architect` and `tool-html-minifier`,
-- so ALL 45 remaining guide pages are unreachable at ANY cap below 98, on a site
-- that grew 15 pages in two days. Raising the cap again cannot work.
--
-- ── WHY A NEW TABLE AND NOT A COLUMN ON site_discovery_rotation ─────────────
--
-- Three reasons, any one sufficient. (1) That table's own COMMENT (migration
-- 346) declares it "Written by the site-discovery-rotation-* pre_queries; safe
-- to TRUNCATE" — it has never had a Go writer, and this cursor must be written
-- from the action. (2) Its `last_selected_at` is NOT NULL with ruled semantics
-- ("when the scheduler last PICKED the site"); an action-side UPSERT would have
-- to invent a value for a column another system reads for fairness ordering.
-- (3) Its `agent_type` means "a scheduled task's target", which coincides with
-- "the agent that dispatched with rotation on" today and is not guaranteed to.
--
-- ── BLAST RADIUS ───────────────────────────────────────────────────────────
--
-- One new table nothing else reads, and ONE key on ONE step of ONE agent.
-- `design-critique-agent` also carries a `request_render_audit` step (cap 8) and
-- is deliberately NOT flipped: it is a manual sampler with no cadence (its own
-- seed, 645, says so in terms), and its 8 pages are plausibly meant as the most
-- important 8 rather than any 8. Rotating what a vision critique looks at is a
-- product decision nobody has taken. It keeps writing prefix-mode truncation
-- rows, which the commissioned reader acks by name.
--
-- FAIL-OPEN BY CONSTRUCTION: with the key absent the code takes today's prefix,
-- and TRUNCATEing this table resets every cursor to the top of the ordering —
-- which is exactly pre-394 behaviour, not a broken state.

BEGIN;

-- ── 1. the cursor table ────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS render_audit_page_cursor (
    site_id         uuid        NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    agent_type      text        NOT NULL,
    after_nav_order int         NOT NULL,
    after_name      text        NOT NULL,
    updated_at      timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (site_id, agent_type)
);

COMMENT ON TABLE render_audit_page_cursor IS
    'bugs_open/394: keyset coverage cursor for request_render_audit, used only by '
    'callers whose step config sets rotate_coverage=true. A row PRESENT means the '
    'site is mid-cycle; a row ABSENT means the next capped run starts from the top '
    'of ORDER BY COALESCE(nav_order,999), name. Written ONLY from Go '
    '(request_render_audit_action.go), AFTER a successful dispatch — a cursor is a '
    'commitment about the next run, so it must not outlive a produce that failed. '
    'Safe to TRUNCATE: every cursor resets to the top, which is exactly the '
    'behaviour this table was added to change, so truncation costs coverage '
    'progress and nothing else.';

COMMENT ON COLUMN render_audit_page_cursor.after_nav_order IS
    'The COALESCE(nav_order,999) value of the last page taken — the SORTED value, '
    'never the raw column, so the cursor and the ORDER BY cannot disagree.';
COMMENT ON COLUMN render_audit_page_cursor.after_name IS
    'pages.name of the last page taken. The pair (after_nav_order, after_name) is '
    'a keyset position, not an index: an index would break the moment a page is '
    'added or removed, and the page set moved by 15 rows in two days on the site '
    'this exists for.';
COMMENT ON COLUMN render_audit_page_cursor.agent_type IS
    'The DISPATCHING agent. Two callers with different caps share a site '
    '(render-audit-agent at 60, design-critique-agent at 8), so one cursor per '
    'site would have them fighting over one position.';

-- ── 2. the opt-in flip, guarded ────────────────────────────────────────────
--
-- Guards copied from migration 392 (council round 700da63e, five seats), because
-- the failure they prevent is unchanged: some agent types carry TWO active rows
-- and only the higher version loads, so an UPDATE keyed on is_active alone could
-- write a row the loader never reads while a verify agrees with itself
-- vacuously. And jsonb_set would happily INSERT the key at a path the loaded
-- workflow does not read if the step layout has moved.

DO $$
DECLARE
  v_rows int;
  v_has_step int;
BEGIN
  SELECT count(*) INTO v_rows
  FROM agent_definitions
  WHERE type = 'render-audit-agent'
    AND is_active
    AND COALESCE(is_snapshot, false) = false
    AND deleted_at IS NULL;
  IF v_rows IS DISTINCT FROM 1 THEN
    RAISE EXCEPTION '646: expected exactly 1 active render-audit-agent row, found % — resolve which row the loader reads before applying', v_rows;
  END IF;

  -- Anchor on a key that must ALREADY exist at the path, so a moved step layout
  -- aborts instead of growing a key nothing reads.
  SELECT count(*) INTO v_has_step
  FROM agent_definitions
  WHERE type = 'render-audit-agent'
    AND is_active
    AND COALESCE(is_snapshot, false) = false
    AND deleted_at IS NULL
    AND default_config->'workflow'->'steps'->'audit'->'config' ? 'max_pages';
  IF v_has_step IS DISTINCT FROM 1 THEN
    RAISE EXCEPTION '646: workflow.steps.audit.config.max_pages not found on the live row — the step layout has changed; do not invent the path';
  END IF;
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
      '{workflow,steps,audit,config,rotate_coverage}', 'true'::jsonb, true),
    updated_at = NOW()
WHERE type = 'render-audit-agent'
  AND is_active
  AND COALESCE(is_snapshot, false) = false
  AND deleted_at IS NULL;

-- ── 3. verify with DO/RAISE — a SELECT verify cannot stop the COMMIT ───────
--
-- ON_ERROR_STOP ignores a non-empty result set, so a verify block made of
-- SELECTs reports the problem and commits anyway. This is the trap LANDMINES
-- records against migration verify blocks; induce it once if you doubt it.

DO $$
DECLARE
  v_rotate text;
  v_tbl    int;
BEGIN
  SELECT default_config->'workflow'->'steps'->'audit'->'config'->>'rotate_coverage'
    INTO v_rotate
  FROM agent_definitions
  WHERE type = 'render-audit-agent'
    AND is_active
    AND COALESCE(is_snapshot, false) = false
    AND deleted_at IS NULL;
  IF v_rotate IS DISTINCT FROM 'true' THEN
    RAISE EXCEPTION '646: render-audit-agent audit.config.rotate_coverage is % (expected true) — aborting', v_rotate;
  END IF;

  SELECT count(*) INTO v_tbl FROM pg_tables
   WHERE schemaname = 'public' AND tablename = 'render_audit_page_cursor';
  IF v_tbl IS DISTINCT FROM 1 THEN
    RAISE EXCEPTION '646: render_audit_page_cursor was not created — the flag would be on with nowhere to store a position';
  END IF;

  -- The other carrier must NOT have been flipped. Checked rather than assumed:
  -- a WHERE clause that widened by accident is exactly the kind of thing that
  -- reads as success.
  IF EXISTS (
    SELECT 1 FROM agent_definitions
     WHERE type = 'design-critique-agent'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
       AND default_config->'workflow'->'steps'->'audit'->'config' ? 'rotate_coverage'
  ) THEN
    RAISE EXCEPTION '646: design-critique-agent was flipped too — it is a manual sampler and must keep its prefix; aborting';
  END IF;
END $$;

COMMIT;
