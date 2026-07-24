-- SEED_wiring_probe_agents.sql — 2026-07-24, chat "diagnosis fixloop 5"
--
-- Two SCRATCH agents for the live induced-fault verification of the route-wiring
-- guard + the code-index freshness banner on v1.0.1155 (council 6cdbc374 r2 +
-- 8ed67200, both APPROVED; debug_historian: compile/pod-grep proves DEPLOYMENT,
-- only an induced fault proves the guard). Pattern = the 036scratch precedent:
-- probe on scratch definitions, live agents untouched, deactivate after (keep as
-- evidence).
--
--   diagnose-wiring-probe-ok  — CORRECT wiring (route step's output_field =
--     "route", matching the reader's defaults) + code_requests injected via
--     input_data, so the gather runs clean and RENDERS the freshness banner in
--     code_evidence. Zero LLM: load_runtime -> complete. Proves the healthy
--     branch of the guard AND the banner, live.
--   diagnose-wiring-probe-bad — MISMATCHED wiring (a diagnose_route step present
--     with output_field = "elsewhere" while the reader keeps route.* defaults).
--     The guard must FAIL the gather loudly with "route wiring mismatch" naming
--     "elsewhere". The route step is never executed — the guard reads the PLAN.
--     Zero LLM: fails before any model call.
--
-- Idempotent (ON CONFLICT (type) DO UPDATE — probe defs carry no concurrent state).
INSERT INTO agent_definitions (type, display_name, category, default_config, is_active)
VALUES
-- NOTE (2026-07-24, first dispatch): the ORIGINAL probe-ok carried BOTH a route
-- step AND code_requests_field=input_data.code_requests — and the live guard
-- correctly REJECTED it: with a route step present, pointing any coupled field
-- off the route namespace IS the divergence the guard exists to catch (risk #1
-- in the approved submission, behaving exactly as reviewed). The probe was
-- wrong, not the guard — an accidental extra falsification. Split since into:
--   probe-ok   (banner): NO route step -> guard skips -> injection legal.
--   probe-pass (healthy): route step + pure defaults -> guard passes -> COMPLETES.
('diagnose-wiring-probe-ok', 'SCRATCH: freshness-banner probe (no route step)', 'diagnostic',
 '{"workflow": {"start_step": "load_runtime", "steps": {
    "load_runtime": {"action": "diagnose_load_runtime",
      "config": {"code_requests_field": "input_data.code_requests"},
      "output_field": "load_runtime", "next_step": "complete",
      "description": "probe gather: no route step (guard skips), injected code_requests render the banner"},
    "complete": {"action": "complete_workflow",
      "config": {"result_from": "load_runtime"},
      "description": "stop before any LLM step"}
 }}}'::jsonb, true),
('diagnose-wiring-probe-pass', 'SCRATCH: route-wiring probe (healthy defaults)', 'diagnostic',
 '{"workflow": {"start_step": "load_runtime", "steps": {
    "load_runtime": {"action": "diagnose_load_runtime",
      "config": {},
      "output_field": "load_runtime", "next_step": "complete",
      "description": "probe gather: reader on route.* defaults, route step matches -> guard passes"},
    "route": {"action": "diagnose_route", "output_field": "route",
      "description": "correct wiring; never executed, the guard reads the PLAN"},
    "complete": {"action": "complete_workflow",
      "config": {"result_from": "load_runtime"},
      "description": "reached only if the guard passes healthy wiring"}
 }}}'::jsonb, true),
('diagnose-wiring-probe-bad', 'SCRATCH: route-wiring probe (mismatched)', 'diagnostic',
 '{"workflow": {"start_step": "load_runtime", "steps": {
    "load_runtime": {"action": "diagnose_load_runtime",
      "config": {},
      "output_field": "load_runtime", "next_step": "complete",
      "description": "probe gather: reader on route.* defaults"},
    "route": {"action": "diagnose_route", "output_field": "elsewhere",
      "description": "MISMATCH: writes under elsewhere while the reader expects route.*"},
    "complete": {"action": "complete_workflow",
      "config": {"result_from": "load_runtime"},
      "description": "unreachable if the guard works"}
 }}}'::jsonb, true)
ON CONFLICT (type, version) DO UPDATE SET
  default_config = EXCLUDED.default_config,
  display_name   = EXCLUDED.display_name,
  is_active      = EXCLUDED.is_active,
  updated_at     = now();
-- (unique key is (type, version); version defaults to 1 — checked 2026-07-24)
-- Deactivate after the probe:
--   UPDATE agent_definitions SET is_active=false, updated_at=now()
--   WHERE type LIKE 'diagnose-wiring-probe-%';
