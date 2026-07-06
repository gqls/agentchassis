# PLAN — Tool Acceptance Runner (headless browser, desktop + mobile)

**Created:** 2026-07-04
**Last updated:** 2026-07-06 (rev 2 — request/response contract pinned to the 035 Adapter Guide, which is normative)
**Status:** initial plan (P0 not started). This is the Tier-4 rung of the tool
verification ladder in `PLAN_travelling_docs.md`, kept as an active goal:
**iterate a tool until it meets its acceptance criteria**, verified by actually
driving it in a browser, on desktop and mobile profiles.

---

## Aim

Given a deployed tool (a `function` on one or more pages of a site), load its
acceptance criteria from its travelling PLAN, drive the deployed page(s) in a
headless browser under desktop and mobile profiles, judge each criterion
pass/fail, and feed failures back into the improvement loop — repeating until
the criteria pass. This turns "the tool is deployed" into "the tool works".

## Reuse (the mould)

- **Adapter pattern, not an in-chassis action.** A browser is heavy infra; the
  chassis stays lean. Follow the analyser-adapter shape exactly
  (`request_repo_analysis` → adapter pod does the work → response on the
  caller's topic, awaited): a **browser-runner adapter** deployment in
  `ai-persona-system`, its image carrying Playwright + Chromium, consuming a
  request topic and replying with structured results. Recommend
  `playwright-go` (Go, matches the repo; device descriptors give mobile
  emulation). `chromedp` is the fallback if the Playwright image is
  objectionable, at the cost of built-in device profiles.
- **Criteria source:** `load_doc_context` already extracts the PLAN's
  ```criteria block (`criteria_json`). The runner consumes that — no second
  criteria store.
- **Iteration machinery:** failures become `improve_tool` work items (existing
  handler `tool-improver`), each carrying the failing criterion as the
  finding-style `acceptance_test`, with the existing `max_fix_attempts`
  convention as the loop bound. Fix agents already load PLAN+NOTES first and
  append a note — the convergence memory is already designed.
- **Results:** work-item result + a `doc_notes` entry per run
  (`categories: ["acceptance-run"]`, plus `acceptance-fail` per failed
  criterion). No new results table until volume demands (graduation rule).

## Criteria contract (v0 — lives in the PLAN's ```criteria block)

```criteria
{
  "profiles": ["desktop", "mobile"],
  "checks": [
    {"id": "boots",        "type": "selector_exists",   "selector": "#tool-root"},
    {"id": "console",      "type": "no_console_errors"},
    {"id": "asset",        "type": "asset_loads",       "path": "/tools/assets/{function}.js"},
    {"id": "calc-flow",    "type": "interaction",
      "steps": [{"action": "fill", "selector": "#hours", "value": "10"},
                 {"action": "click", "selector": "#calculate"}],
      "expect": {"selector": "#result", "text_matches": "\\d"}},
    {"id": "mobile-fit",   "type": "no_horizontal_overflow", "profiles": ["mobile"]}
  ]
}
```

Check types v0: `selector_exists`, `selector_count`, `no_console_errors`,
`asset_loads`, `interaction` (fill/click/select steps + expect),
`no_horizontal_overflow`, `page_status_ok`. A check without `profiles` runs on
all. Deterministic only in v0 — no LLM drives the browser; an LLM-exploratory
mode is a later phase, if ever.

## Profiles

- **desktop:** Chromium, 1366×900, DPR 1, mouse.
- **mobile:** a Playwright device descriptor (e.g. Pixel 7 or iPhone 13 —
  pick one and keep it stable so results are comparable): real viewport, DPR,
  touch, mobile UA. Real devices are out of scope; emulation first.

## Request/response contract (adapter) — conforms to 035 §1 (normative)

- **Topics (035 §2.10, Convention A):** requests on
  `system.adapter.browser-runner.requests`; consumer group
  `browser-runner.adapter.group`; env overrides `REQUESTS_TOPIC` /
  `CONSUMER_GROUP`. The reply goes to the request's reply topic — read
  `responses_topic` from the headers, falling back to `reply_to_topic` in
  headers then body (accept all three; senders set both).
- **Request envelope (035 §1.2):** `action` from `body.action`
  (`"run_checks"`), payload at `body.data`:
  `{urls, profiles, criteria_json, function, site_id}`.
  `correlation_id`/`request_id` come from the message headers. (`run_id` is a
  payload detail only — it is NOT the matcher.)
- **Response (035 §1.3/1.4/1.8 — the favoured git/thunder pattern):** a typed
  local header struct with real bool `is_complete`/`is_error`;
  `in_response_to_request_id` = the **incoming** `request_id` (the load-bearing
  matcher); `request_id` reused from the request; `message_id` a fresh UUID;
  echo `client_id`, `correlation_id`, `orchestration_id`,
  `in_response_to_step_id`, `in_response_to_step_name`;
  `sender_agent_type: "browser-runner-adapter"`;
  `status`: `complete` | `error_recoverable` | `error_unrecoverable`. Body
  envelope `{headers, body: {success, data: {results: [{check_id, profile,
  url, pass, detail}]}}}`. Sent via **`ProduceWithValidation`** — never plain
  `Produce` (§1.9 anti-pattern: string bools, fresh request_id without
  in_response_to_request_id, no message_id, plain Produce).
- **§2.1 decision check:** heavy external dependency (a browser), multiple
  future callers (acceptance agent, audits) → adapter is the right pattern.
- **Health endpoints, config YAML, Deployment, creds:** per 035 §2.9–2.13
  (no external credentials needed for P0; public URLs only).

## Flow (one acceptance cycle)

```
acceptance work item (item_type: acceptance_run, handler: tool-acceptance-agent)
  → load_doc_context (criteria_json; skip+flag if none — an undocumented tool
    gets a needs_criteria note, not a fake pass)
  → resolve deployed URLs for the function
  → request browser runs (desktop + mobile) via adapter, await
  → judge: all pass → doc_note (acceptance-run, pass) + item complete
           any fail → doc_note per failure (acceptance-fail, criterion id)
                      + improve_tool item carrying the failing criterion
                        (acceptance_test) — bounded by max_fix_attempts
  → tool-improver fixes (loads PLAN+NOTES first) → redeploy → next acceptance run
```

Trigger points: after tool creation/recreation deploys; after any
`improve_tool` completes; periodic via `check_tool_health` Tier 2 queue.

## Phasing

- **P0 — adapter skeleton + boot checks.** Image (Playwright+Chromium),
  Convention-A topic + consumer group, typed response-header struct +
  `ProduceWithValidation` wiring, health endpoints, config YAML,
  `selector_exists` + `no_console_errors` + `page_status_ok`, desktop only.
  Exit = the 035 §2.15 smoke: hand-produce a request, expect a response with
  `in_response_to_request_id=$REQUEST_ID` and `status=complete`, results
  matching manual inspection of one live tool page.
- **P1 — criteria interpreter + mobile profile.** Full v0 check set, both
  profiles, `no_horizontal_overflow`.
- **P2 — interactions.** fill/click/expect flows (the "does it actually
  calculate" tier).
- **P3 — evidence.** Screenshots per failure uploaded via the existing deploy
  path (Backblaze), paths in the doc_note.
- **P4 (optional, later) — LLM-exploratory mode** for tools whose criteria
  can't be fully declared.

## Open questions

- **URL the runner drives:** the deployed public URL (Backblaze/CDN) is
  simplest and tests what users get; an in-cluster preview would test
  pre-deploy — start public, revisit if pre-deploy gating is wanted.
- **Adapter image weight + concurrency:** one run at a time per pod first;
  scheduler concurrency group if runs queue.
- **Multi-page tools:** the criteria block may name per-page checks (add a
  `url_role` field to checks) — design when the first multi-page tool lands.
- **Where `tool-acceptance-agent` sits:** own agent definition (orchestrator,
  per the constitution) — drafted after P0's adapter proves the contract.
