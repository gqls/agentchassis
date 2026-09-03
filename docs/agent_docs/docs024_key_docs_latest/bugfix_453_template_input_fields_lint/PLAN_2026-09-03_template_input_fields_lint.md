# PLAN — 2026-09-03 — the template-variable ↔ `input_fields` lint (`bugs_open/453` candidate 1)

## What this lane is building

`bugs_open/453` fix candidate 1: a fleet-wide check that asks, of every live step that renders
a prompt template, whether each template variable's ROOT can actually reach the template.

Home: `cmd/config-key-audit --template-input-fields`, the binary that already sweeps
`agent_definitions` config and already owns the shared `validation.WalkSteps` descent.

## Ownership (checked before starting, 2026-09-03)

`scripts/who-owns.py 453` → **no owning workstream**. Two lanes touch the file and BOTH
explicitly declined this build: the filer (`apis_uk_bees_homepage`) wrote *"what is OPEN is a
build decision (candidate 1), which is a human's call"*, and the contributor
(`bugs_open/437` lane) wrote *"I am not proposing to build candidate 1"*. So this is an
unowned candidate, not a competing fix. Findings go back INTO `bugs_open/453`.

## The three failure shapes, and WHICH ONES THIS CLOSES

The 437 lane's contribution (2026-09-03) is right that the seam has three shapes, and that
candidate 1 as originally worded did not say which it closes. It closes **shape 2 only**, and
the deliverable says so in its own `--help` text and in the finding names:

| shape | what it is | closed by this lint? |
|---|---|---|
| 1 | no `input_fields` at all → randomised recursive resolution (the LANDMINE sibling) | **no** — reported as context, not convicted |
| 2 | root missing from `input_fields` → the variable can NEVER resolve | **YES — this is the deliverable** |
| 3 | root present, SUB-FIELD absent in the row's data → `<no value>` | **no, and no static lint over config ever can** — the config is correct and the shape depends on a row |

Shape 3's remedy is the 437 lane's suggestion (promote `RenderPromptTemplate`'s existing
`<no value>` scan from a `Warn` to a durable row). That is a shared prompt seam, wants its own
round, and is NOT in this lane's scope. Stated here so the next reader does not think this
lint was meant to cover it and failed.

## Design decisions, and the reasons

### D1 — parse the template with Go's own parser, not a regex

`text/template/parse` is what the runtime uses, so the check inherits the real scoping rules
for free: inside `{{range}}` / `{{with}}` the dot is REBOUND, so `{{range .items}}{{.name}}{{end}}`
must not report `name` as a missing root. A regex over `{{\.(\w+)}}` reports `name`, which on
the fleet's biggest template (`page-content-writer`, ~40 range bodies) is a wall of false
positives that would get the check switched off in its first week.

`{{$var := ...}}` variables and function-call arguments fall out correctly for the same reason.

### D2 — the "available roots" model is READ FROM THE OWNING CODE, never copied

`bugs_open/453` names this explicitly: *"the extractor's speciallyHandled set must be read from
ONE place or the lint inherits the classifier-gap problem."* So:

- `datahelpers` exports the specially-handled set and the input-field→root rule, and
  `ExtractFields` itself uses the exported forms. A copy in the check could drift silently;
  a shared symbol cannot.
- The per-action injected roots are exported from `actions` — because they are NOT uniform,
  which the first cut of this lane's own sizing script got wrong. `execute_vision_prompt`
  injects `vision_image_manifest` and does NOT inject the platform blocks; `execute_llm_prompt`
  injects `voice_style`/`build_standard` and not the manifest. A single global set reported
  both vision steps as broken. That was two false positives out of twelve in the first run.

### D3 — three finding kinds, and only ONE of them fails the run

- `unreachable_root` (**exit 1**): `input_data` is NOT among `input_fields`, so the root set is
  fully determined by config and the variable can never resolve, on any row, ever.
- `conditional_root` (**exit 0**): `input_data` IS among `input_fields`. `ExtractFields` promotes
  every key of the runtime `input_data` map to the root, so whether the variable resolves depends
  on a row this check cannot see. Reported, never convicted.
- `declared_unread` (**exit 0**): the cheap reverse direction 453 also asked for — an
  `input_fields` entry no template variable reads.

The precedent for "report the undecidable class but never fail on it" is `defaultshadow.go`'s
`dotted_conditional`, same binary.

### D4 — the agent-level prompt tier is in scope, and its absence must be LOUD

`getPromptWithPriority` has three tiers; tier 2 is `agent_definitions.default_config.prompt_template`,
used by **6** live steps as of 2026-09-03. Covering it needs one extra key in the export, which
the other audit scripts' exports do not carry. A missing key would make the tier silently
invisible — the "reads as clean while blind" shape. So the mode REFUSES (exit 2) when no decoded
row carries the key at all, which is exactly the wrong-export case: `jsonb_build_object` emits
the key as JSON `null` when the projection is present, so "present on zero rows" can only mean
the projection is missing.

## Acceptance

`bugs_open/453` §How to verify: *"expecting AT LEAST one true positive it can regression-pin"*,
and a zero result on first run is to be treated as a suspected wrong path. Both hold — see NOTES
for the measured first run.
