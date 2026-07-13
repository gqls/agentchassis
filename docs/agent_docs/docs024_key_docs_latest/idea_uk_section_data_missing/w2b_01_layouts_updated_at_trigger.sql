-- W2b step 1: make layouts.updated_at update automatically.
-- GATE: only run if w2b_00 found NO existing equivalent function. CREATE FUNCTION
-- (deliberately not OR REPLACE) errors on a name collision rather than silently
-- overwriting different semantics — if it errors, reuse the existing function name
-- in the trigger instead and skip the CREATE.
-- Note: the codebase convention is explicit `updated_at = now()` in each UPDATE
-- (fix_nav_link_templates, IncrementUsageCount); the trigger coexists harmlessly
-- (it overwrites with the same now()). Scoped to layouts only; extendable later.
CREATE FUNCTION set_updated_at() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
  NEW.updated_at := now();
  RETURN NEW;
END $$;

DROP TRIGGER IF EXISTS trg_layouts_updated_at ON layouts;
CREATE TRIGGER trg_layouts_updated_at
BEFORE UPDATE ON layouts
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Bump the W1-stale row THROUGH the trigger (no-op column write) and verify:
UPDATE layouts SET name = name WHERE name = 'tool-portal-light';
SELECT name, updated_at FROM layouts WHERE name = 'tool-portal-light';
