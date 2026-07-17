# SCHEMA (draft, for owner sign-off) — the staged plan (`stages[]`), feature-builder delta 1

*2026-07-17, "fixloop feature builder" thread. This is the sign-off artifact the
handoff (`HANDOFF_2026-07-17_feature_builder_thread.md` step 2) requires BEFORE
any code. Parent design: `DESIGN_feature_builder_and_council_gate.md` §1. Nothing
here is built; §7 lists the decisions this draft needs from the owner.*

## 0. What this schema is

The fix loop's `fix_plan` artifact is ONE constrained edit plan
(`{summary, edits[], grounded_in[], risks}`, validated in
`diagnose_persist_fix_plan_action.go`). The feature builder's plan is a SEQUENCE
of such plans — stages — with declared inter-stage dependencies, per-stage gate
criteria, and an encoded image-then-seed discipline. The council router judges
whatever artifact it is given (confirmed against
`diagnose_council_decide_action.go`); the cage does not change.

Grounding read before drafting (per the handoff): `0NN_fix_proposer.sql` (v6
live shape), `0NN_fix_implementer.sql`, `0NN_fix_implementer_orchestrator.sql`,
`diagnose_persist_fix_plan_action.go`, `diagnose_prepare_fix_commit_action.go`,
`diagnose_read_repo_files_action.go`.

## 1. The schema (annotated)

```json
{
  "plan_format": "staged-v1",
  "summary": "one paragraph: the capability being built and the shape of the build",
  "stages": [
    {
      "id": "s1",
      "title": "short name of what this stage delivers",
      "goal": "what exists after this stage that did not exist before it",
      "depends_on": [],
      "edits": [
        {
          "file": "repo-relative/path.go",
          "symbol": "FunctionOrStep",
          "operation": "modify|add|remove|config_change",
          "artifact_role": "code|seed|doc",
          "rationale": "why THIS edit, tracing to the approved spec/design",
          "sketch": "the intended change, described precisely"
        }
      ],
      "expected_symbols": ["NewFunctionName", "action_registry_key"],
      "gate": { "build": true }
    }
  ],
  "post_merge_checklist": [
    { "order": 1, "act": "image_deploy",
      "detail": "make build-agent-chassis-ref REF=<merge commit>; bump IMAGE_TAG; verify the running pod binary" },
    { "order": 2, "act": "seed_apply", "file": "docs/.../0NN_new_agent.sql",
      "detail": "psql clients_db -f <the seed file, from the merged tree>" },
    { "order": 3, "act": "verify",
      "detail": "how the owner confirms the feature is live end to end" }
  ],
  "grounded_in": ["verbatim quotes from the approved spec / design the plan rests on"],
  "risks": "what could this break; what a reviewer should check"
}
```

## 2. Field rules

**Top level.**
- `plan_format` — REQUIRED, literal `"staged-v1"`. The discriminator: its
  presence (with `stages`) selects staged validation; its absence selects the
  legacy single-plan path unchanged. `stages` and a top-level `edits` are
  mutually exclusive.
- `summary`, `grounded_in`, `risks` — same rules as today (summary and
  grounded_in must be non-empty). For a feature, `grounded_in` quotes the
  APPROVED SPEC rather than a diagnosis — same field, different source, so the
  persist/PR machinery needs no change.
- `post_merge_checklist` — REQUIRED whenever any edit has
  `artifact_role: "seed"` or operation `config_change`; otherwise optional.
  Rules in §4.

**Stage.**
- `id` — REQUIRED, unique, `^[a-z0-9_-]{1,16}$`. Referenced by `depends_on`.
- `title`, `goal` — REQUIRED, non-empty. `goal` is the reviewable contract:
  the council judges stage boundaries by it.
- `depends_on` — ids of stages this stage needs. MUST reference strictly
  EARLIER stages (array order is execution order; v1 schedules nothing in
  parallel — the field is validated documentation for the council and a hook
  for later parallelism, not a scheduler input).
- `edits[]` — same shape and rules as today's `fixPlanEdit` (repo-relative
  path, no traversal/whitespace; operation in the allowlist; rationale and
  sketch non-empty; the no-op phrase rejections stand). One addition:
  `artifact_role` (default `"code"`) — see §4.
- `expected_symbols` — optional list. At implement time, each symbol must
  appear verbatim in at least one of that stage's produced file bodies —
  a deterministic post-LLM check in the prepare step, same spirit as the
  allowlist. (At plan time only non-emptiness of each string is checkable.)
- `gate.build` — default `true`: gofmt + targeted go build on the stage's
  changed files (existing `diagnose_build_gate`, scoped per stage). `false` is
  legal ONLY for a stage whose edits are all `artifact_role` seed/doc (nothing
  to build); validation enforces that implication.

**End gate (not in the schema).** After the final stage, `go test` runs over
the union of packages containing edited `.go` files — DERIVED from the plan,
not declared in it, so the model cannot narrow its own test surface. (Optional
widening by the owner at fire time is a trigger flag, not a plan field.)

## 3. Cross-stage file rules (the new-file discipline)

1. A path may be `add`ed at most ONCE across the whole plan.
2. `modify` of a path that is `add`ed in a LATER stage is invalid
   (modify-before-create).
3. `modify` of a path `add`ed in an EARLIER stage is legal — this is the
   normal "create it, then wire it" shape.
4. `remove` of a path `add`ed earlier in the same plan is invalid (churn — the
   plan should not create what it then deletes).
5. The implementer's hard allowlist is now PER STAGE: stage N's commit may
   touch only stage N's modify/add files (`validateImplementation` runs per
   stage with that stage's edit list). To-be-created files enter the allowlist
   through the stage's own `add` edits — the allowlist mechanism itself is
   unchanged.
6. Reads for stage N happen at the FIX BRANCH as of stage N-1's commit (stage
   1 reads the base ref). This is why F1.2 — ref/base as a per-run input,
   today a config literal at `diagnose_read_repo_files_action.go:101` — is a
   structural PRECONDITION of the stage loop, not just a cleanup. The pilot
   (§6) fixes it first.

## 4. Seed/apply discipline, encoded (design delta 3)

- `artifact_role: "seed"` marks an edit whose file is a DB-side seed shipped
  AS A FILE in the PR (e.g. `docs/.../0NN_new_agent.sql`). The builder NEVER
  executes seeds.
- Validation rules:
  1. every `seed` edit must be referenced (by `file`) from exactly one
     `post_merge_checklist` entry with `act: "seed_apply"`;
  2. if any checklist entry is `seed_apply`, at least one `image_deploy` entry
     must exist with a strictly LOWER `order` — image-first-then-seed is thus
     structurally unexpressible in the wrong order, not merely documented;
  3. `order` values are unique positive integers; the PR body renders the
     checklist sorted by `order` as the owner's apply checklist.
- `config_change` keeps its existing meaning (a described change to a LIVE
  agent_definitions row, carried in the PR body, applied by a human). Guidance
  for the designer prompt: NEW agent definitions ship as seed files; changes
  to EXISTING live rows may be either a seed file (preferred — reviewable,
  reapplicable) or a `config_change` description.

## 5. What each existing component needs (compatibility map)

| Component | Change for staged-v1 |
|---|---|
| `diagnose_persist_fix_plan_action.go` | `validateFixPlan` branches on `plan_format`/`stages` presence: legacy path byte-for-byte unchanged; staged path adds §2–§4 rules. `max_plan_bytes` default rises for staged plans (proposed 131072); metadata gains `stage_count`. |
| `diagnosis_artifacts` kinds | NONE — staged plans persist as `kind='fix_plan'` (discriminated by `plan_format` in the body). No DDL. (Alternative: a new `feature_plan` kind — §7 D1.) |
| council (3 seats + `diagnose_council_decide`) | Router/aggregator unchanged (judges the artifact given). Reviewer prompts in the NEW feature-designer workflow gain staged-plan judging criteria (stage boundaries, dependency order, reuse-before-recreate); the fix-proposer's live prompts are untouched. |
| `diagnose_read_repo_files_action.go` | Already distinguishes modify (must exist) from add (expected absent). Needs: ref resolvable per run/stage (F1.2), and per-stage edit-list scoping. |
| `diagnose_prepare_fix_commit_action.go` | `validateImplementation` unchanged in mechanism, invoked per stage with the stage's edits; adds the `expected_symbols` check; branch naming decision §7 D2; base_branch per-run (F1.2). |
| `diagnose_build_gate_action.go` | Unchanged; invoked once per stage with that stage's changed files. |
| implementer workflow | New stage-loop router (deterministic Go, mirroring the diagnosis loop's iteration pattern): `create_branch` once → per stage {read at branch → sketch_to_files → prepare → commit → build gate} → end test gate → ONE PR. Build delta 2 — separate sign-off when its design is drafted. |
| intake | Triage already routes `capability_gap` to the roadmap (`FIX-051` in the concept register) — that pool is the feed; spec approval remains a human act. |

Concept-register note (per this thread's boot instruction): the register's
fix-loop category (`FIX-*`) documents the chain this extends; the staged-plan
schema, once built, is itself register material (a new `FIX-*` concept), and
the bug-historian seat — live but unexercised — will see staged plans through
the same council seam it already occupies. Nothing in this schema changes its
prompt or contract.

## 6. Worked example — the pilot (F1.2 self-hosted)

The owner-suggested pilot as a staged-v1 instance (abridged sketches):

```json
{
  "plan_format": "staged-v1",
  "summary": "Make the fix-implementer's read ref and base branch per-run inputs (today config literals, live-set to a stale branch), so re-fires cannot silently read/branch from the wrong base — and so the coming stage loop can read its own branch between stages.",
  "stages": [
    { "id": "s1", "title": "ref/base as resolvable inputs",
      "goal": "both Go actions resolve ref/base_branch from input_data at run time, config literal then 'main' as fallbacks",
      "depends_on": [],
      "edits": [
        { "file": "platform/orchestration/actions/diagnose_read_repo_files_action.go",
          "symbol": "DiagnoseReadRepoFilesAction", "operation": "modify", "artifact_role": "code",
          "rationale": "ref is a config literal (line ~101); a stale live value silently reads the wrong tree",
          "sketch": "resolve ref via the action-input spec (input_data.ref → config → 'main')" },
        { "file": "platform/orchestration/actions/diagnose_prepare_fix_commit_action.go",
          "symbol": "DiagnosePrepareFixCommitAction", "operation": "modify", "artifact_role": "code",
          "rationale": "base_branch has the same literal-only resolution",
          "sketch": "same input-spec resolution for base_branch" }
      ],
      "expected_symbols": ["input_data.ref"],
      "gate": { "build": true } },
    { "id": "s2", "title": "re-seed the implementer workflow",
      "goal": "the live workflow passes input_data.ref/base through, and create_branch's from_branch comes from the same input",
      "depends_on": ["s1"],
      "edits": [
        { "file": "docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/0NN_fix_implementer_v2_ref_input.sql",
          "operation": "add", "artifact_role": "seed",
          "rationale": "the workflow config must name the new input fields; seeds ship as PR files, never executed",
          "sketch": "v2 seed: read_current_files/prepare gain ref/base field refs; create_branch from_branch via data_fields" }
      ],
      "gate": { "build": false } }
  ],
  "post_merge_checklist": [
    { "order": 1, "act": "image_deploy", "detail": "make build-agent-chassis-ref REF=<merge>; bump IMAGE_TAG; verify pod binary" },
    { "order": 2, "act": "seed_apply", "file": "docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/0NN_fix_implementer_v2_ref_input.sql",
      "detail": "apply to clients_db; snapshot_agent runs inside the seed" },
    { "order": 3, "act": "verify", "detail": "delete stale fix/* branches; fire the implementer on a known approved correlation with an explicit ref; confirm it reads/branches from it" }
  ],
  "grounded_in": ["HANDOFF_2026-07-17_feature_builder_thread.md: 'make implementer ref/base a per-run INPUT (they are live-set to a stale branch)'"],
  "risks": "behavioural default must remain 'main' so existing fires are unchanged; the stale live literals are the bug being removed"
}
```

This instance exercises every new mechanism once: multi-stage deps, add of a
to-be-created file, the seed role, checklist ordering, a build-gated and a
non-build stage — while being small and gradable.

## 7. Decisions needed from the owner (sign-off gates)

- **D1 — artifact kind.** Recommended: reuse `kind='fix_plan'` with the
  `plan_format` discriminator (no DDL, implementer load queries unchanged).
  Alternative: new `kind='feature_plan'` (cleaner queries; needs a constraint
  change + parallel load steps).
- **D2 — branch prefix.** `feat/<short-corr>` for feature builds (recommended:
  distinguishes them from `fix/*` in every listing and in the stale-branch
  cleanup habit) vs reusing `fix/*`.
- **D3 — caps.** Proposed defaults: max 6 stages, 8 edits/stage (today's
  per-plan cap becomes per-stage), 24 edits total, 131072 plan bytes. All
  config-overridable per run, as today.
- **D4 — seed discipline as validation.** Confirm §4's rules (seed⇒checklist
  coverage; image strictly before seed) as HARD validation failures, not
  advisories.
- **D5 — pilot.** Confirm F1.2 (§6) as the first feature build, and that its
  s1 lands via the feature builder itself once deltas 1–2 exist (self-hosting)
  rather than being hand-built first. Note the circularity is benign: the
  PLAN above is hand-written; only its implementation runs through the loop.
- **D6 — end gate.** Confirm derived-from-plan `go test` packages (model
  cannot narrow its own test surface), owner-widenable at fire time.

Sign-off on D1–D6 unblocks build delta 1 (schema + validation in
`diagnose_persist_fix_plan_action.go`) and the feature-designer agent draft.
Nothing is coded until then.
