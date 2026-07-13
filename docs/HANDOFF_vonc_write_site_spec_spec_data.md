# HANDOFF — vonc.com `write_site_spec` / `spec_data` failure

A self-contained brief to START a diagnosis in a fresh chat. It carries the
EVIDENCE and the CONTEXT; it does NOT contain a diagnosis — the cause is still to
be read from the real code (the log line below is the symptom). Memory is off in
the new chat, so this doc is the context; the runbooks/NOTES referenced are the
durable record if deeper history is needed.

## The bug (the symptom, as logged)

Surfaced from the LIVE `agent_error_log` for site `vonc.com` (read read-only via
the diagnosis harness's runtime gather against `postgres-clients-0`). Two failures,
seconds apart, identical shape:

```
2026-06-22 17:02:59  generic  persist_mission_brief  write_site_spec
   step persist_mission_brief failed: failed to execute action write_site_spec:
   spec_data must be a JSON object, got string
2026-06-22 17:03:00  generic  persist_roadmap_brief  write_site_spec
   step persist_roadmap_brief failed: failed to execute action write_site_spec:
   spec_data must be a JSON object, got string
```

So: the `write_site_spec` action received `spec_data` as a **string** when it
expects a **JSON object** (`map[string]interface{}`). Both failing steps —
`persist_mission_brief` and `persist_roadmap_brief` — are in the same agent
(`agent_type = generic`, i.e. a workflow running on the generic chassis worker).

## Downstream effect (same site, `site_work_items`)

Not necessarily caused by the above — gathered alongside it, hold as context, do
NOT assume a link until the evidence shows one:

```
needs_page:index                 failed              3/3   "Claim timed out (attempts exhausted)"   2026-06-22 17:13
needs_page (index)               complete            2/3   "Claim timed out — handler pod likely died"  2026-06-23 00:15
page_rerender_index_...          complete            0/3                                            2026-06-23 00:39
unresolved_cta_index_hero_...    needs_human_review  0/3                                            2026-06-22 21:08
```

Reading: index page-build first FAILED on claim timeouts (17:13), then later
SUCCEEDED on a rerender (next day). A hero CTA is parked for human review. The
"Claim timed out — handler pod likely died" rows are a SEPARATE concern (worker
liveness / claim mechanics), almost certainly unrelated to the spec_data shape
error — do not conflate them.

## What is CONFIRMED vs STILL TO READ

CONFIRMED (from the evidence):
- `write_site_spec` rejected `spec_data` because it was a string, not a JSON object.
- It happened on two brief-persisting steps in a `generic` workflow for vonc.com.
- The error is a DATA-SHAPE mismatch at an action input boundary.

STILL TO READ (the new chat must do this BEFORE diagnosing — the log is the
symptom, the cause is in the code/workflow):
1. The `write_site_spec` action itself — where it validates `spec_data` and emits
   "spec_data must be a JSON object, got string". Find the exact check. Does it
   accept a JSON-STRING and unwrap it (many actions do, via the datahelpers
   unwrap/StripCodeFences/tryParseJSON helpers), or does it strictly require an
   already-parsed object? `grep -rn "spec_data must be a JSON object" platform/`.
2. The two workflow steps `persist_mission_brief` / `persist_roadmap_brief` in the
   relevant agent_definition's `default_config` workflow — how do they supply
   `spec_data`? Is it `{{...}}`-templated (which can stringify an object), a
   `*_field` path reference, or a prior step's output read under the wrong key?
   The mismatch is almost certainly HERE: a producing step emits the brief as a
   JSON string (or a template renders an object to a string) and `write_site_spec`
   wants the object.
3. Which agent runs these steps — `agent_type = generic` means the generic worker;
   find the agent_definition whose workflow contains `persist_mission_brief` /
   `persist_roadmap_brief` (`grep -rn "persist_mission_brief" ` migrations + any
   workflow SQL). That row's workflow is the thing to inspect/repair.

## Likely shape of the fix (a HYPOTHESIS to falsify, NOT a conclusion)

Two candidate structural fixes; the code read decides which (or neither):
- **Producer side:** the step feeding `spec_data` is handing over a serialized
  JSON string; pass the OBJECT through instead (or stop a template stringifying it).
- **Consumer side:** `write_site_spec` should tolerate a JSON-STRING input by
  unwrapping it with the existing datahelpers (the codebase already has
  `SafeUnmarshalString`, `tryParseJSON`, `StripCodeFences`, `UnwrapDeep`) — matching
  how sibling actions accept either shape. This is the more robust structural fix IF
  other callers also pass strings, but confirm first that it doesn't paper over a
  producer that should emit an object.
Prefer the structural fix over a patch; decide from the actual producer/consumer
code, not from this doc.

## How to get the evidence into the new chat

EASIEST: attach these bundle artefacts from the run (kept on disk):
- `/tmp/diag_bundle_1.md` (and `_2`, `_3`) — the assembled bundles per iteration.
- `/tmp/bundle-<id>/runtime.md` — the runtime evidence shown above (the
  `agent_error_log` + `site_work_items` rows). The bundle dir for iter 1 of the
  last run was `/tmp/bundle-3132848707/` (re-run to regenerate if cleared — /tmp
  may not survive a reboot).
The bundles alone WILL let the new chat start, but they are code-context for the
ORIGINAL (scripted) hypothesis, not for spec_data — so the new chat should re-scope
to `write_site_spec` and read items 1–3 above. Supplying this handoff + the
`write_site_spec` action file + the agent_definition workflow is faster than bundles
for THIS bug.

To regather fresh runtime evidence for vonc.com directly (read-only):
```
go run ./cmd/diagnose -analysis /tmp/vonc-analysis.json -root ~/projects/agentchassis \
  -constitution ./thin_slice_constitution.md \
  -psql 'kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db' \
  -seed-hypothesis "write_site_spec rejects spec_data as a string on persist_mission_brief/persist_roadmap_brief" \
  -seed-scope "<write_site_spec action path>" -runtime-site vonc.com -runtime-page index \
  -callgraph /tmp/vonc-analysis.json -dry-bundle
```
(Drop `-dry-bundle` and add a `-verdict-script` only when actually running the loop;
without a model the stub abstains. The runtime gather is read-only.)

## Standing rules the new chat MUST follow (carried from this project)

- Schema-before-SQL: run `\d <table>` before writing any SQL. A 0-row result is NOT
  decisive until the query itself is checked.
- Reuse before create: search for an existing action/struct/helper that does the job
  (or can be altered) before writing new code. STEP ZERO of the dev guide.
- Structural fix over quick patch.
- Read the REAL code/signatures/schema, don't assume — verify the datahelpers/action
  APIs by grepping the package (this project has been bitten repeatedly by assumed
  helpers and stale copies).
- Workflows stay thin; complexity lives in Go action code. Keep workflow variable
  names in sync with what actions read. No sub-workflows in SQL — spawn sub-agents.
- Every agent is an orchestrator; agents reply to the CALLER's responses topic.
- No `logger.Debug` (won't show); no summary docs unless asked; British English;
  flag any variable/signature changes explicitly.
- k8s namespace: `-n ai-persona-system` (DB pod `postgres-clients-0`,
  db `clients_db`, user `clients_user`); Kafka in `-n kafka`.
- Follow the repo's agent-creation + debugging guides (016 debugging guide; the dev
  guide). The 016 §9 catalogue + the cross-module/assumed-helper entries are relevant
  if the cause turns out to be a data-shape/contract drift like several recent ones.

## One-line orientation for the new chat

"Diagnose why `write_site_spec` fails with 'spec_data must be a JSON object, got
string' on the `persist_mission_brief` / `persist_roadmap_brief` steps for vonc.com.
Read the action + the two workflow steps that feed `spec_data` first; the log line is
the symptom. Prefer a structural fix (producer emits an object, or consumer unwraps a
JSON-string via existing datahelpers) over a patch."
