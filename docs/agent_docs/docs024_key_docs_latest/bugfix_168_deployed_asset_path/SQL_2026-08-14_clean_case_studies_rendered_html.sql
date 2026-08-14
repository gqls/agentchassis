-- Second surface: the same minimal deletion applied to page_components.rendered_html.
--
-- WHY A SECOND EDIT IS NEEDED, and this corrected my own plan mid-task:
-- RerenderSinglePageAction is "Simple concatenation - no template re-rendering"
-- (its own file header) — it assembles the page from stored per-component
-- rendered_html and does NOT regenerate that HTML from content_data. So the
-- content_data edit alone would have been republished with the claim intact.
-- ScanDeployedClaims reads BOTH surfaces (rendered_html for the number scan,
-- content_data for the stat scans) and the claim-granular gate searches
-- html||content_data, so both must be clean anyway.
--
-- NOT the section editor: this item type is HITL-terminal and its own fix text
-- says "Truth decisions are human — do not auto-rewrite". An LLM re-render of
-- the section would be exactly the auto-rewrite the check forbids. A deletion
-- of the identical 36-char plain-ASCII sentence from a text node is the same
-- minimal act as the content_data edit, and changes no markup.
--
-- The trg_page_component_artefact_archive_upd trigger archives the old
-- rendered_html on update, so this is recoverable independently of the guard.
\set ON_ERROR_STOP on

BEGIN;

SET LOCAL app.expect_delta = :expect_delta;

DO $$
DECLARE
  v_pc      uuid := '9c9aaed8-d13f-4825-afdb-e449bfe8f92b';
  v_needle  text := '75,061 orchestration state records. ';
  v_delta   int  := current_setting('app.expect_delta')::int;
  v_old     text;
  v_new     text;
  v_locked  timestamptz;
  v_hits    int;
  v_rows    int;
BEGIN
  SELECT rendered_html, locked_at INTO v_old, v_locked
    FROM page_components WHERE id = v_pc;

  IF v_old IS NULL OR v_old = '' THEN
    RAISE EXCEPTION 'ABORT: no rendered_html on %', v_pc;
  END IF;

  -- Respect the platform's own write guard for this table.
  IF v_locked IS NOT NULL THEN
    RAISE EXCEPTION 'ABORT: component is human-locked (locked_at %)', v_locked;
  END IF;

  v_hits := (length(v_old) - length(replace(v_old, v_needle, ''))) / length(v_needle);
  IF v_hits <> 1 THEN
    RAISE EXCEPTION 'ABORT: needle occurs % time(s) in rendered_html, expected exactly 1', v_hits;
  END IF;

  v_new := replace(v_old, v_needle, '');

  UPDATE page_components
     SET rendered_html = v_new,
         updated_at    = now()
   WHERE id = v_pc AND locked_at IS NULL;
  GET DIAGNOSTICS v_rows = ROW_COUNT;
  IF v_rows <> 1 THEN
    RAISE EXCEPTION 'ABORT: updated % row(s), expected exactly 1', v_rows;
  END IF;

  -- Re-read rather than trust the local variable: this asserts what the table HOLDS.
  SELECT rendered_html INTO v_new FROM page_components WHERE id = v_pc;

  IF position('75,061' in v_new) > 0 THEN
    RAISE EXCEPTION 'ABORT: 75,061 survives in rendered_html';
  END IF;

  IF length(v_old) - length(v_new) <> v_delta THEN
    RAISE EXCEPTION 'ABORT: rendered_html shrank by % chars, expected %',
      length(v_old) - length(v_new), v_delta;
  END IF;

  IF replace(v_old, v_needle, '') <> v_new THEN
    RAISE EXCEPTION 'ABORT: rendered_html differs by more than the removed sentence';
  END IF;

  RAISE NOTICE 'OK: rendered_html -% chars, 1 row, markup untouched.', v_delta;
END $$;

COMMIT;
