# 148 — Three live agent definitions name an action that exists in no registry, and nothing says so until a message arrives

**Filed** 2026-07-29 by bugsearch-7 (the sub-workflow validation lane, `bugs_open/144`)
· **Status** OPEN, unowned · **Class** silent coverage gap / dead dispatch edge
· **Found by** the dry-run harness built for 144, which replays `ValidateWorkflow` over
an export of the live fleet. This was not what it was looking for.

---

## 1. The one-line version

`ValidateWorkflow` rejects a step whose action is not in the local registry and which
carries no topic. **Three live, `is_active` agent definitions are in exactly that
state**, so they are rejected on every message they receive — and **three live builders
still dispatch at two of them.** Nothing reports this offline: the only thing that
knows is the validator, and the validator only speaks when a message arrives.

## 2. Evidence

MEASURED 2026-07-29 over live `agent_definitions` (`deleted_at IS NULL`, non-snapshot,
`is_active`), walking top-level AND nested steps, against the 301 actions in
`platform/orchestration/actions/registry.go`:

| agent | step | action | in registry? | topic? |
|---|---|---|---|---|
| `html-developer-chunked` | `steps.assemble_parts` | `assemble_html_parts` | no | none |
| `multipage-wrapper` | `steps.wrap_multipage` | `wrap_multipage` | no | none |
| `html-assembler` | `steps.assemble_html` | `assemble_full_page` | no | none |

Fleet-wide that is **3 steps across 3 agents — the whole population**, not a sample.

The three action names appear in **no Go file at all**:

```bash
grep -rn '"assemble_html_parts"\|"wrap_multipage"\|"assemble_full_page"' --include=*.go .
# (no output)
```

**And they are still dispatch targets.** Nine steps across three live builders:

```sql
SELECT ad.type, e.k, v->>'action', v->'config'->>'agent_type'
FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') e(k,v)
WHERE ad.deleted_at IS NULL AND COALESCE(ad.is_snapshot,false)=false AND ad.is_active
  AND v::text LIKE ANY (ARRAY['%html-assembler%','%multipage-wrapper%','%html-developer-chunked%']);
-- website-builder       | spawn_wrapper   | spawn_agent | multipage-wrapper
-- landing-page-builder  | call_wrapper    | call_agent  | multipage-wrapper
-- landing-page-builder  | call_assembler  | call_agent  | html-assembler
-- content-site-builder  | spawn_assembler | spawn_agent | html-assembler
-- … 9 rows
```

Reproduce the detection:

```bash
SUBWF_LIVE_EXPORT=<export> go test ./platform/orchestration/actions/ \
  -run TestLiveDefinitionsPassSubWorkflowValidation -v
# → 3 × "ALREADY REJECTED BEFORE THIS CHANGE — <agent>: step '<x>' with action '<y>' requires a topic"
```

(Export command: `docs024_key_docs_latest/bugfix_144_subworkflow_validation/RUNBOOK_subworkflow_validation.md`.)

## 3. What is NOT claimed

- **[UNMEASURED] whether those dispatch edges are ever taken.** `orchestration_states`
  holds no row for any of the three agent types, but its oldest row is 2026-07-13
  (1,952 rows) — so "never in 16 days", not "never". The two possibilities are very
  different: a path never taken is dead weight; a path taken is a builder that fails
  at a step and cannot say why.
- **This is not a claim that the validator is wrong.** Rejecting them is correct. The
  defect is that nothing says so until a message arrives, and that the edges pointing
  at them were never retired.

## 4. Prior art — read this before treating it as new

`bugs_closed/044` ("nothing detects a capability that exists but nothing routes work
to") names `html-assembler` explicitly in its list of **~34 legacy/superseded**
definitions: *"retired code still flagged `is_active=true`, which is its own hygiene
problem"*. 044 closed on the capability-inventory half and **explicitly scoped the
`is_active` hygiene half out as an owner decision**.

What is new here, and is not in 044:

1. these three are not merely dormant, they are **structurally unrunnable** — a
   dormant agent would run if dispatched, and these cannot;
2. **live builders still point at two of them**, so the dead capability is reachable
   from current code;
3. the general form — **an action name that exists in no registry is undetectable
   offline** — is a checkable invariant nobody checks, and it is cheap.

## 5. Fix candidates, ordered by what closes the door

1. **Make the bad state visible everywhere it is created: check every live definition's
   action names against the registry, offline.** The traversal and the registry lookup
   both already exist (`validation.WalkSteps`; `actioncheck.IsLocalAction`) — this is a
   report, not a mechanism. Natural homes, in order of how little is new: a mode on
   `cmd/config-key-audit` (already imports both, already reads the live export); a
   `scripts/pattern-check.py` entry; a `discovery_checks/` check if a work item is
   wanted rather than a report. **Note the sibling trap before building anything:
   `bugs_open/144`'s whole cause was two hand-written traversals disagreeing — reuse
   `WalkSteps`, do not write a third.**
2. **Retire the dead dispatch edges** in `website-builder`, `landing-page-builder`,
   `content-site-builder`, or repoint them. DB config, live immediately — but do (3)
   or you have three agents that still cannot run, merely unreferenced.
3. **`is_active` hygiene on the ~34 retired definitions** — 044 made this an explicit
   owner call. It stays one. Answering (1) first gives the owner the actual list.

**Do not do (2) alone.** Removing the references makes the report clean while the
unrunnable definitions remain — the same "move the blindness" mistake `bugs_open/144`
§5 warns about for its own second half.

## 6. Landmines

- **A step with no topic and an unregistered action is REJECTED; a step with a topic is
  not.** A remote action is dispatched by topic and need not be in the local registry
  at all, so "action not in registry" alone is NOT the defect — only the pair is. A
  check that ignores the topic will produce a long false-positive list and get ignored.
- **`orchestration_states` is on a retention clock** (oldest row 2026-07-13 as of
  filing). Record a RATE or a window, never a bare "never ran".
- 044 is CLOSED and correct; this is not a reopening. Cite it, do not fork it.
