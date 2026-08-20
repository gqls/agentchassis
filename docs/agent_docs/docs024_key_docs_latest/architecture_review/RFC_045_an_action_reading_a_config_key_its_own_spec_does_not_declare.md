# RFC 045 — an action reading a config key its own spec does not declare: the class behind `bugs_open/336`, and the one instrument nobody has

**Status: OPEN — raised 2026-08-20 by the `bugfix_336` lane. Nothing built, nothing changed.
This is a routing document: three council seats asked for it in a single round, and the
submission's own answer was a stated intention with no work attached — which is the thing
one of those seats objected to.**

Raised from council correlation `bc2f4b0e-45db-49c8-9f45-6af74a344cce` (round 1, REVISE).
Case file: `bugs_closed/336_HANDOFF_2026-08-20_deploy_result_field_is_declared_on_the_wrong_actions_spec_so_arming_it_hard_fails_every_workflow_that_stamps_a_page.md`.
Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_336_config_key_on_the_wrong_spec/`.

## 1. What the thing IS, before any rule about it

An **action input spec** is a Go declaration listing the step-config keys an action reads
(`datahelpers.ActionInputSpec`, registered per action name in an `init()`). It is the
action's own written statement of its config contract. Two mechanisms consume it:
`UnknownConfigKeys` reports a live step carrying a key the spec does not list, and
`StrictConfig` promotes that report from a warning into a hard validation error that
rejects the whole workflow.

**The class this RFC is about: an action reads a key that its own spec does not declare.**
The key may be declared nowhere, or — `bugs_open/336`'s shape — declared on a *different
action's* spec that happens to sit in the same file.

The rule that makes it dangerous is `StrictConfig`. Where the reading action is strict, the
key is unrecognised *for that action*, so arming it on a live step does not merely warn: it
fails validation on every message. That is what took the publish path down fleet-wide on
2026-08-20 — 8 items across 4 item types, 123 `page_rerender` queued and none draining.

## 2. Why it exists rather than another line in a risks block

Three seats said the same thing in one round:

> `bug_historian [medium]`: "This plan fixes the loud instance and leaves the
> silent-failure-prone general mechanism fully exploitable elsewhere. … the plan should
> either widen Test 3's scan or explicitly file the fleet-wide audit as follow-up work
> rather than a stated intention with no work item attached."

> `reuse_agent [medium]`: "The plan's own rationale names `cmd/config-key-audit` as the
> right home for the fleet-wide version but then writes a second, narrower scanning
> mechanism instead of adding a single-action mode to (or calling into) the existing one."

> `architecture`: "Building that fleet-wide check here would itself be scope creep into a
> shared cross-cutting tool from inside a single-bug patch — the plan is right not to do
> that."

The last two are not in conflict; together they say exactly where the work belongs and that
a bug patch is not it. This file is that destination.

## 3. The population, measured [MEASURED 2026-08-20, at HEAD `ade78a426`]

`go run ./cmd/config-key-audit --specs`, counted from its own JSON:

| | count |
|---|---|
| registered action input specs | **173** |
| opted into unknown-key detection (`opted_in`) | **82** |
| with a non-empty `ConfigKeys` list | 16 |
| **not opted in at all** | **91** |
| **`StrictConfig: true`** | **2** |

The two strict actions are the entire loud blast radius, and they are named by one grep —
`grep -rn "StrictConfig:" --include=*.go` returns exactly two non-test hits:
`update_page_status` (`platform/orchestration/actions/v3_site_actions.go:665`) and
`create_work_item` (`platform/orchestration/actions/create_work_item_action.go:145`).

**Both are clean at HEAD, checked exhaustively rather than sampled:**

- `update_page_status` — its handler body (`v3_site_actions.go:727-1168`) indexes config for
  exactly six keys (`status`, `page_id_field`, `site_id_field`, `page_name_field`,
  `page_component_id_field`, `deploy_result_field`), and its spec declares exactly those
  six. No helper in the body takes `config` as an argument, so there is no indirect read to
  miss.
- `create_work_item` — reads `loop_iteration` and `loop_var_name` without declaring them,
  and that is **correct, not a finding**: both are framework-reserved keys injected by loop
  expansion (`platform/orchestration/datahelpers/action_inputs.go:218-224`), and
  `UnknownConfigKeys` recognises the framework set. Every other key it reads is declared.

**So the loud class has zero live instances today.** That is a real result and it is also
the reason this is an RFC and not a `bugs_open/` entry: there is no reproducible damage to
point at. What there is, is 91 actions where the same mistake produces **no signal at all**
— not a failure, not a warning — which is the `bugs_open/234` shape: a key an author wrote,
that reads as wired, that no code path resolves.

### What would have disconfirmed the above

Stated because a figure that could not come out otherwise is not evidence. The strict-action
census would have failed if `grep "StrictConfig:"` had returned a third hit, or if either
handler had indexed a key absent from its spec, or if a helper in either body had taken
`config` — all three were checked and all three could have gone the other way. The
`render_component` arm of the 336 verification (below) had a live control: the same
all-depths `jsonb_path_query('$.**')` search that returned **0** carriers for
`render_component` returned **3 of 7** for `update_page_status`, so the zero is a
measurement, not a broken query.

## 4. Why no existing instrument sees this class

This is the part that makes it worth a tool rather than a habit. From 336's own landmine:

- The key literal **is in the binary**, put there by the reading action's `zap.String` call,
  so a `/proc/1/exe` probe says PRESENT no matter which spec lists it.
- `git log -S'"deploy_result_field",'` matches that same `zap` call, so it **names the
  commit that shipped the reader** and reads as though it named the declaration.
- An arming precondition written about the reader ("wait for a build carrying the reader's
  commit") is **satisfied** while the declaration sits on the wrong spec.
- `cmd/config-key-audit --suspicious-keys` inspects key *names* for documentation
  punctuation. It has no opinion on where a key is declared.
- `--live-pairs` and `--specs` between them hold the *declared* half and the *in-use* half.
  Neither holds the **read** half, and the read half is the only one that can settle this.

Every instrument agreed, and none of them was looking at the list inside the named spec.

## 5. What a `--read-vs-declared` mode has to do, and the hard part

The declared half is already there (`--specs` emits it, keyed by registered action name).
The missing half is a source scan of each action's handler for the keys it actually reads.

**The hard part is that a literal `config["..."]` scan is not sufficient, and we already
have the counter-example in the tree.** `RenderComponentInputSpec` declares
`refuse_dead_url_controls` with a comment recording that the action has read it since the
guard shipped — but through `shouldRefuseDeadURLControls(config, ...)`, not through a
literal index — so the 2026-08-18 census, which was a grep over the function, could not see
it. A grep-only mode would therefore report false negatives on exactly the indirection it
most needs to follow, and would do so silently.

This is not hypothetical for this lane either: 336's own sibling-key check was recorded
**INCONCLUSIVE rather than clean** for precisely that reason, and it was right to be.

Sketch of what would actually work, in rough order of cost:

1. **Literal scan + an explicit unresolved list.** Walk each handler with `go/ast`, collect
   `config[<literal>]`, and *separately* list every call in the body that receives `config`
   (or `params.StepConfig.Config`) as an argument. Report the second list as
   "unresolved — read sites this mode cannot follow", so the blind spot is printed rather
   than absent. Cheap, honest, and immediately better than a grep.
2. **One level of interprocedural follow.** Resolve those callees within the package and
   scan their bodies too. Covers the `shouldRefuseDeadURLControls` shape, which is the one
   we know exists.
3. **Whole-program.** `golang.org/x/tools/go/packages` plus a call-graph walk. Correct, and
   almost certainly more machinery than this class warrants today.

Option 1 is the one that closes the honesty gap; option 2 closes the known instance. The
mode belongs in `cmd/config-key-audit`, where `--specs`, `--removed-keys-in-use` and the
`relaygaps`/`defaultshadow`/`sharedoutputs` scanners already live, and where the daily
CronJob wiring already exists — per `reuse_agent`, and per RFC_024's finding that a tenth
standalone meta-check is not what this estate needs.

## 6. What is NOT proposed

- **Not making more actions strict.** The 91 non-opted-in actions are non-opted-in for a
  stated reason (`action_inputs.go`'s own header: an over-strict validator is a worse bug
  than the inert key it chases, and a fleet-wide flag day would reject working definitions).
  This RFC asks for a *report*, not a gate.
- **Not a runtime check.** An action cannot tell at execution time whether its own spec is
  complete; like `--single-owner-actions` and `findUnregisteredActions`, this has to be
  offline and fleet-wide.
- **Not a per-action regression test per action.** 336 shipped one for `update_page_status`
  (`platform/orchestration/actions/update_page_status_config_contract_test.go`). Doing that
  173 times by hand is the duplication this RFC exists to avoid.

## 7. The decision asked for

Whether to build option 1 (or 1+2) as a `--read-vs-declared` mode in `cmd/config-key-audit`,
reported daily alongside the existing checks; or to record that the loud class is bounded to
two verified-clean actions and accept the silent class as a known, unguarded surface until
an instance appears.

Either answer is defensible on today's evidence. What is not defensible is a third round of
"recorded, not actioned" — which is how this file came to exist.
