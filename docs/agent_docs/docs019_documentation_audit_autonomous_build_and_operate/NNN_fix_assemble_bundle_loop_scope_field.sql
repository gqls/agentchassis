-- NNN_fix_assemble_bundle_loop_scope_field.sql
--
-- Fixes the diagnose loop's RE-SCOPE (the thing the whole loop exists to do).
--
-- diagnose_assemble_bundle reads its loop scope from loop_scope_field "route.scope"
-- through datahelpers.ExtractStringListHelper — i.e. it expects a []string of
-- "path:Symbol" entries. But diagnose_route writes route.scope =
-- diagnose.EncodeScope(NextScope), which is json.Marshal of the Scope struct. The
-- Scope struct carries NO json tags, so the marshalled object uses the Go field
-- names as keys:
--     {"Symbols":[...],"Tables":[...],"RuntimeSite":"...","RuntimePage":"...","Capabilities":false}
-- A string-list reader over that OBJECT coerces to empty, so on every loop-back the
-- scope fell through the fallback chain (seed_scope -> code_results) and never
-- advanced. First-pass diagnosis still worked (iteration 1 uses code_results); only
-- the re-scope (iterations 2+) was silently inert — which would defeat §6E and the
-- §6G eval (REFUTE -> REFUTE -> CONFIRM requires the scope to move).
--
-- Fix: point loop_scope_field at the Symbols LIST inside the object,
-- "route.scope.Symbols" (capital S — matches the untagged marshalling).
-- ExtractNestedField traverses three levels here (precedent in the same configs:
-- input_data.section_plan.sections_ready, site_specs.identity.team), so
-- route.scope.Symbols resolves to the []string the reader expects.
--
-- Config-only; NO rebuild. The READ-ONLY loop is unchanged — only the field path
-- moves. snapshot first (standing rule). Alternative considered (NOT taken, would
-- need a rebuild + contextkit engine sync): add json tags to the Scope struct and
-- use route.scope.symbols, or have diagnose_route write NextScope.Symbols (a list)
-- directly at route.scope. The field-path move is the smallest structural fix that
-- makes producer and consumer agree.

BEGIN;

SELECT snapshot_agent(
  'diagnose-agent',
  'fix assemble_bundle.loop_scope_field route.scope -> route.scope.Symbols (re-scope read the EncodeScope object as a string list and fell back to the seed scope each iteration)'
);

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,assemble_bundle,config,loop_scope_field}',
      '"route.scope.Symbols"'::jsonb,
      true
    ),
    updated_at = now()
WHERE type = 'diagnose-agent';

-- verify — expect exactly: route.scope.Symbols
SELECT default_config #>> '{workflow,steps,assemble_bundle,config,loop_scope_field}'
         AS loop_scope_field
FROM agent_definitions
WHERE type = 'diagnose-agent';

COMMIT;
