-- 366_tool_recreation_reads_the_evidence_register.sql
--
-- PLAN_2026-08-09_facts_into_tool_acceptance.md, Piece 1 — the half that needs
-- no Go and no image.
--
-- WHAT IS WRONG TODAY. `tool-recreation-handler` rebuilds a calculator from the
-- original's source plus a functional specification, and it is never told what
-- this site has verified. The register is not missing from its world: the
-- `load_site_specs` step calls `read_site_spec` with NO `aspect` in config, and
-- that mode returns EVERY current aspect keyed by aspect name
-- (site_spec_actions.go, "All aspects mode"). So `.site_specs.specs.evidence_base`
-- already resolves in this very prompt's data — the prompt simply never mentions
-- it, while it does mention `.site_specs.specs.identity.industry`,
-- `.site_specs.specs.design` and `.site_specs.specs.site_archetype.visual_character`.
-- Measured 2026-08-10: `evidence_base` appears 0 times in the 6,499-character
-- template. The facts arrive and are thrown away.
--
-- WHY IT MATTERS MORE FOR A TOOL THAN FOR COPY. The claims gate already stops a
-- PAGE asserting an unregistered number. Nothing looks inside a calculator's
-- JavaScript, so a threshold baked into a script is unexamined by every rung of
-- the acceptance ladder. mortgagecalculator.co.uk's original stamp-duty tool
-- granted first-time-buyer relief between £500,000 and £625,000 for sixteen
-- months after the rules withdrew it (bugs_open/225), and this agent's standing
-- instruction is to study the original and "achieve the same functionality" —
-- i.e. to reproduce exactly that, and make it look freshly built.
--
-- WHAT THIS DOES. Inserts one section immediately before `## Design Context`,
-- stating that a registered fact OVERRIDES both the original source and the
-- functional specification. It injects the register's own composed
-- `writer_block` — the prose roster the evidence-freshness sweep maintains
-- (`composeWriterBlock`) and the page writer is already given — rather than
-- raw `facts[]` JSON, so there is one wording of a fact on this estate and not
-- two that drift.
--
-- WHAT IT DOES NOT DO, said plainly so nobody reads a guarantee into it: an LLM
-- shown a fact may still ignore it. This lowers the odds of F1 (a tool BUILT
-- with a stale constant); it does not close the door. The closing mechanisms are
-- Pieces 2-4 of the plan and they need Go. In particular the code comment asked
-- for below is a TRACE for a human reader — it must never become the machine
-- declaration of which facts a tool encodes, because a comment enforces nothing.
--
-- SAFE ON THE CURRENT BINARY — verified, not assumed (2026-08-10):
--   * Prompt text only. No workflow step, no action, no config key. Nothing here
--     is read by Go; the template is rendered by the existing LLM step.
--   * The two-level access on a possibly-absent aspect is already PROVEN IN THIS
--     SAME TEMPLATE: `{{if .site_specs.specs.site_archetype.visual_character}}`
--     has the identical shape and ships today. A site with no evidence_base row
--     therefore takes the else branch exactly as a site with no site_archetype
--     already does.
--   * 10 of 11 current registers carry `writer_block`; the eleventh takes the
--     else branch, whose wording is true for "no register" and for "register
--     with no composed block" alike.
-- Fleet effect: additive on 10 sites, inert on the rest. No ordering constraint
-- is claimed (owner ruling 2026-07-29, condition 1 retired).
--
-- ROLLBACK: 366_tool_recreation_reads_the_evidence_register_ROLLBACK.sql removes
-- the inserted section by the same anchor.

BEGIN;

SELECT snapshot_agent('tool-recreation-handler',
  '366_tool_recreation_reads_the_evidence_register.sql: pre-update') AS snapshot_id;

UPDATE agent_definitions SET default_config = jsonb_set(
  default_config,
  '{workflow,steps,recreate_tool,config,prompt_template}',
  to_jsonb(
    replace(
      default_config #>> '{workflow,steps,recreate_tool,config,prompt_template}',
      '## Design Context',
      E'## Verified facts — these OVERRIDE the original tool AND the specification\n'
      || E'{{if .site_specs.specs.evidence_base.writer_block}}\n'
      || E'This site keeps a register of verified facts. Each is traced to an official published source and re-checked automatically every day. Where the original tool''s code, the functional specification above, or any figure you would otherwise reach for disagrees with a fact below, THE FACT WINS — the original is a defect to correct, not a behaviour to reproduce.\n\n'
      || E'That is not hypothetical. A calculator on this estate applied a tax-relief threshold that legislation had withdrawn sixteen months earlier. A faithful recreation would have carried that error forward and made it look freshly built.\n\n'
      || E'{{.site_specs.specs.evidence_base.writer_block}}\n\n'
      || E'How to use it:\n'
      || E'- Take the exact figure. Do not round it, restate it in other units, or infer a neighbouring threshold from it.\n'
      || E'- Where you hard-code a constant that comes from this register, put the fact''s wording beside it in a short code comment, so a later reader can see where the number came from.\n'
      || E'- Do NOT state a rule that is not in the register. If the tool needs a threshold that is not listed, implement what the specification says and claim nothing about its legal standing.\n'
      || E'- If a listed fact contradicts the original tool, implement the fact and say so plainly in the tool''s own help text, so a visitor who knows the old behaviour is not left guessing.\n'
      || E'{{else}}\n'
      || E'No verified-fact block is available for this site. Implement the functional specification as written, and do not introduce rates, thresholds or legal rules of your own.\n'
      || E'{{end}}\n\n'
      || '## Design Context'
    )
  )
)
WHERE type = 'tool-recreation-handler'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Guard the exact post-condition, inside the transaction so a miss rolls back.
-- Three assertions, each of which could fail independently:
--   1. exactly one row changed (not zero — a silent no-op is this file's main
--      failure mode, and `UPDATE ... WHERE` reports success on zero rows);
--   2. the anchor was inserted BEFORE, not instead of, `## Design Context`
--      (the section must still be there once);
--   3. the writer_block reference appears exactly TWICE — once in the `{{if}}`
--      guard and once in the interpolation. The first draft of this guard
--      asserted ONCE and the dry-run refused the file, which is the guard doing
--      its job on its author: the EXPECTATION was wrong, not the edit. Left
--      exact rather than loosened to `>= 1`, because the count is what makes a
--      double-application visible.
--   4. the section heading appears exactly once — the idempotency check proper.
DO $$
DECLARE
  t text;
  n_anchor  int;
  n_block   int;
  n_section int;
  n_rows    int;
BEGIN
  SELECT count(*) INTO n_rows FROM agent_definitions
   WHERE type = 'tool-recreation-handler'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n_rows <> 1 THEN
    RAISE EXCEPTION 'expected exactly 1 live tool-recreation-handler row, found %', n_rows;
  END IF;

  SELECT default_config #>> '{workflow,steps,recreate_tool,config,prompt_template}'
    INTO t FROM agent_definitions
   WHERE type = 'tool-recreation-handler'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  n_anchor := (length(t) - length(replace(t, '## Design Context', '')))
              / length('## Design Context');
  n_block  := (length(t) - length(replace(t, '.site_specs.specs.evidence_base.writer_block', '')))
              / length('.site_specs.specs.evidence_base.writer_block');

  IF n_anchor <> 1 THEN
    RAISE EXCEPTION 'expected 1 "## Design Context" after the insert, found %', n_anchor;
  END IF;
  IF n_block <> 2 THEN
    RAISE EXCEPTION 'expected 2 evidence_base.writer_block references ({{if}} + interpolation), found %', n_block;
  END IF;

  n_section := (length(t) - length(replace(t, '## Verified facts', '')))
               / length('## Verified facts');
  IF n_section <> 1 THEN
    RAISE EXCEPTION 'expected exactly 1 inserted section, found % (double application?)', n_section;
  END IF;
  IF position('## Verified facts' in t) > position('## Design Context' in t) THEN
    RAISE EXCEPTION 'the section landed AFTER ## Design Context, not before it';
  END IF;
END $$;

COMMIT;
