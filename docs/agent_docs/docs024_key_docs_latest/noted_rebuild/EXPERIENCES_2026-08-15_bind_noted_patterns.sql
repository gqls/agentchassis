-- =============================================================================
-- noted.co.uk — bind the two experience patterns to the BUILT components
-- 2026-08-15. PLAN §4 step 1, which could not happen until the editor existed.
-- =============================================================================
--
-- The bind action (bind_site_experience_action.go) refuses four things, each a
-- door: an unclosed binding (every schema key bound, every supplied key used), an
-- empty value, a selector with no parseable anchor, and a page that resolves to
-- nothing. Every value below is a real element id on a deployed page — read from
-- editor_tool/noted-write.html and legacy_tool/noted-legacy-rescue.html, not
-- guessed. Both patterns are still `draft`, so the action records the fork as
-- `proposed`, not `bound` — that IS PLAN §4.1's `proposed → bound → verified`,
-- so nothing here jumps a step.
--
-- ONE SCHEMA CHANGE, STATED: legacy-local-data-adoption declares `adopt_control`
-- as REQUIRED and uses it in ZERO checks or contract clauses. It was declared on
-- 08-11 for the "sign in and copy everything up" route, which was deliberately
-- NOT built (the download route is the one §6 makes mandatory — no account
-- needed). Binding a required key to a selector that does not exist would be
-- exactly the "check that cannot fail" the action's own header warns about. So
-- adopt_control is made OPTIONAL here, with the reason in the schema itself, and
-- becomes required again if and when the adopt route ships.
--
-- WHAT THIS DOES NOT DO: it does not verify. Only a green run may write
-- `verified`, and three of the checks are known-unrunnable by the platform
-- (HANDOFF 08-11 §5) — those are covered by the Playwright probes.
-- =============================================================================

BEGIN;

UPDATE experience_patterns
SET binding_schema = jsonb_set(
      jsonb_set(binding_schema, '{required}', '["tool_section","summary_region","download_control"]'::jsonb),
      '{properties,adopt_control,note}',
      to_jsonb('OPTIONAL since 2026-08-15: the adopt (sign-in-and-copy-up) route was deliberately not built for the first release; the download route is the mandatory one. Used by no check. Make required again when the adopt route ships.'::text)),
    updated_at = now()
WHERE name = 'legacy-local-data-adoption';

DO $$
DECLARE req jsonb;
BEGIN
  SELECT binding_schema->'required' INTO req FROM experience_patterns WHERE name='legacy-local-data-adoption';
  IF req ? 'adopt_control' THEN RAISE EXCEPTION 'adopt_control still required'; END IF;
END $$;

COMMIT;

-- The bindings themselves go through the ACTION (081b_bind_noted_experiences.sh),
-- so its four doors run. Values, for the record — all real ids on deployed pages:
--
-- authenticated-note-sync  → /tools/write/index.html  (editor_tool/noted-write.html)
--   tool_section    #noted-write          sign_in_form    #nw-auth-form
--   email_input     #nw-email             password_input  #nw-password
--   sign_in_submit  #nw-signin            note_list       #nw-list
--   note_editor     #nw-content           save_indicator  #nw-status
--   api_base        /api                  sample_email / sample_password: a
--                                         throwaway the runner may use
-- legacy-local-data-adoption → /tools/legacy-rescue/index.html
--   tool_section    #legacy-rescue        summary_region  #lr-counts
--   download_control #lr-download         empty_state     #lr-empty
