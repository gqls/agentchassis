-- Migration NNN — teach the tool-doc header to tool-generator and tool-improver.
--
-- Targets (verified from the live rows, 2026-06-11 dump — BOTH are TOP-LEVEL
-- steps, so the 072 nested-path trap does not apply here):
--   tool-generator → default_config #> '{workflow,steps,generate_tool_html,config,prompt_template}'
--   tool-improver  → default_config #> '{workflow,steps,improve_tool,config,prompt_template}'
--
-- Method: anchored replace() on the prompt TEXT, not whole-prompt overwrite —
-- survives unrelated prompt edits since the dump; ABORTS LOUDLY if an anchor
-- has drifted (re-pull the row and re-anchor rather than guessing).
-- Idempotent: re-running skips with a NOTICE.
--
-- PRE-CHECKS:
--   \df snapshot_agent          -- confirm the snapshot function exists
--   SELECT type FROM agent_definitions WHERE type IN ('tool-generator','tool-improver');
--
-- DEPLOYMENT ORDER (load-bearing): apply THIS before or with the binary
-- carrying create_tool_component's HasToolDocHeader gate — gate without
-- prompt means every generation fails at the gate.

-- ── Snapshots first (house discipline) ──
SELECT snapshot_agent('tool-generator');
SELECT snapshot_agent('tool-improver');

-- ── tool-generator: new rule 13 + the Structure example's <script> block ──
DO $gen$
DECLARE
    cur  text;
    new_ text;
    rule_anchor   text := $a$12. Include a clear heading and brief instruction text$a$;
    struct_anchor text := $b$<script>
  // Self-contained logic
</script>$b$;
    rule_add text := $c$
13. Begin the <script> block with exactly one tool-doc header comment between the sentinels /* === tool-doc === and === /tool-doc === */ stating: function: (the exact Function value above), purpose: (one sentence), behaviour: (the invariants your code keeps — units, ranges, no external calls), inputs: (what the user provides), outputs: (what the tool renders). Never put names, ids, or dates in it. Never use */ inside it.$c$;
    struct_new text := $d$<script>
/* === tool-doc ===
function: (the exact Function value above)
purpose: One sentence — what the tool computes for the visitor.
behaviour:
  - invariants the code keeps (units, ranges, rounding)
  - no external calls; all computation client-side
inputs: what the user provides
outputs: what the tool renders
=== /tool-doc === */
  // Self-contained logic
</script>$d$;
BEGIN
    SELECT default_config #>> '{workflow,steps,generate_tool_html,config,prompt_template}'
      INTO cur FROM agent_definitions WHERE type = 'tool-generator';
    IF cur IS NULL THEN
        RAISE EXCEPTION 'tool-generator prompt path not found — workflow shape changed; locate the prompt step first (072 lesson)';
    END IF;
    IF position('=== tool-doc ===' IN cur) > 0 THEN
        RAISE NOTICE 'tool-generator: header instructions already present — skipping';
        RETURN;
    END IF;
    IF position(rule_anchor IN cur) = 0 THEN
        RAISE EXCEPTION 'tool-generator: rule anchor not found — prompt drifted since the 2026-06-11 dump; re-pull and re-anchor';
    END IF;
    IF position(struct_anchor IN cur) = 0 THEN
        RAISE EXCEPTION 'tool-generator: structure anchor not found — prompt drifted since the 2026-06-11 dump; re-pull and re-anchor';
    END IF;

    new_ := replace(cur, rule_anchor, rule_anchor || rule_add);
    new_ := replace(new_, struct_anchor, struct_new);

    UPDATE agent_definitions
       SET default_config = jsonb_set(default_config,
              '{workflow,steps,generate_tool_html,config,prompt_template}',
              to_jsonb(new_)),
           updated_at = NOW()
     WHERE type = 'tool-generator';
    RAISE NOTICE 'tool-generator: header rule + structure example applied';
END
$gen$;

-- ── tool-improver: new rule 10 (preserve/update the header) ──
DO $imp$
DECLARE
    cur  text;
    new_ text;
    rule_anchor text := $e$9. Do not add external dependencies$e$;
    rule_add text := $f$
10. Preserve the tool-doc header comment block (/* === tool-doc === ... === /tool-doc === */) at the top of the <script>. If your fix changes the tool's behaviour, update its behaviour: lines to match. If the block is missing, add it using the Function value shown above. Never use */ inside it.$f$;
BEGIN
    SELECT default_config #>> '{workflow,steps,improve_tool,config,prompt_template}'
      INTO cur FROM agent_definitions WHERE type = 'tool-improver';
    IF cur IS NULL THEN
        RAISE EXCEPTION 'tool-improver prompt path not found — workflow shape changed; locate the prompt step first (072 lesson)';
    END IF;
    IF position('=== tool-doc ===' IN cur) > 0 THEN
        RAISE NOTICE 'tool-improver: header instruction already present — skipping';
        RETURN;
    END IF;
    IF position(rule_anchor IN cur) = 0 THEN
        RAISE EXCEPTION 'tool-improver: rule anchor not found — prompt drifted since the 2026-06-11 dump; re-pull and re-anchor';
    END IF;

    new_ := replace(cur, rule_anchor, rule_anchor || rule_add);

    UPDATE agent_definitions
       SET default_config = jsonb_set(default_config,
              '{workflow,steps,improve_tool,config,prompt_template}',
              to_jsonb(new_)),
           updated_at = NOW()
     WHERE type = 'tool-improver';
    RAISE NOTICE 'tool-improver: preserve-header rule applied';
END
$imp$;

-- ── VERIFY ──
SELECT type,
       position('=== tool-doc ===' IN default_config #>> '{workflow,steps,generate_tool_html,config,prompt_template}') > 0 AS generator_has_header,
       position('=== tool-doc ===' IN default_config #>> '{workflow,steps,improve_tool,config,prompt_template}') > 0 AS improver_has_header
FROM agent_definitions
WHERE type IN ('tool-generator','tool-improver');
-- Expect: tool-generator row generator_has_header = t; tool-improver row improver_has_header = t.
