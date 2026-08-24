# RFC_051 — `call_agent`'s `input_mapping` allow-list silently manufactures dead config keys, and the recurrence rate is now the argument

**Status:** OPEN — filed 2026-08-24 by the `bugs_open/382` lane, at the direction of the council
gate's `architecture` seat, which raised it as an on-record objection while approving the fix it
was objecting past. This is not a proposal to change anything today; it is the design pass the
seat asked for, with the recurrence measured rather than asserted.

**Raised by:** council gate `e53f57ae-3bb1-442c-8e7b-742a1c2bb0ad`, seat `architecture`, verdict
`object` inside an overall **approved** round (1 advisory objection, none high-severity):

> *"This is the third bug (011, 390, 382) against the same `call_agent`/`input_mapping` seam
> producing a dead step-config key. The fix is correctly scoped to `routing.go`, but the seam that
> keeps generating this class (`extractDataForAgent` silently dropping non-mapped config) is
> untouched and will produce a fourth instance for the next caller that adds a config key expecting
> it to reach `input_data`. … I'm recording that the recurrence rate is a cost-of-not-changing
> signal on the input-mapping seam itself, worth a design pass before a fourth instance, not a
> blocking condition on this submission."*

## 1. The mechanism, in plain terms

A workflow step that calls another agent is written as a `call_agent` step with a `config` block.
Inside that block, one key — `input_mapping` — is a list of *"give the callee this field, taken
from this path in my collected data"*. `extractDataForAgent`
(`platform/orchestration/actions/call_agent.go:974-1018`) resolves that mapping and returns it.
**That is the entire payload.** Every other key in the step's `config` is either consumed by the
coordinator for its own purposes (`agent_type`, `target_role`, `timeout_seconds`, `output_mapping`)
or consumed by nobody at all.

So if an author writes a config key intending the callee to receive it, the callee never sees it.
Nothing errors. No validation rejects an unknown config key. The step succeeds. The callee takes
its zero-value branch and carries on.

## 2. Why this shape in particular is hard to see

The dead key is usually spelled to *match the callee's own fallback*, because that is what the
author was reading when they wrote it. `image-build-handler` carried `default_kind: "hero"` beside
a `resolveKind` that reads `inputData["default_kind"]`. The word `hero` is right there in the
config, one line above a working `input_mapping`, on a step that has never failed.

**A reviewer asking "does this branch supply a kind?" sees the word and stops.** That is not a
hypothetical reviewer — it is what happened to migration `390`, which diagnosed this exact class,
wrote *"`default_kind` here has never done anything"* into its own header, fixed two of the three
branches, and then asserted in its blast-radius paragraph that the third *"already forward[s] kind"*.
It did not. That false sentence is the direct cause of `bugs_open/382`.

## 3. The recurrence, measured

| # | date | instance | how it was found |
|---|---|---|---|
| 1 | 2026-07-18 | `bugs_open/011` — the routing `switch`'s silent `default:` (the ancestor of the class, not an `input_mapping` case itself) | a human looked at a gibberish diagram on a client homepage |
| 2 | 2026-08-11 | migration `390` — `call_hero_gen` / `call_logo_gen` carry a config-level `default_kind` that nothing reads | a `needs_logo` slug investigation; an owner-rejected hero |
| 3 | 2026-08-24 | `bugs_open/382` — `call_variant_gen`, same shape, missed by 390's own blast-radius claim | the owner noticed a face |

**Three in about six weeks, all on the same file family, and not one found by a check.** Every one
was found by a person looking at an image. That is the seat's point: the class does not produce
evidence of itself.

## 4. The live surface, so the next reader does not re-measure it

Across every live `call_agent` step, the config-key population **as of 2026-08-24**:

| key | live steps | read by |
|---|---|---|
| `target_role` | 59 | coordinator |
| `timeout_seconds` | 58 | await machinery |
| `input_mapping` | 57 | `extractDataForAgent` — the only key that reaches the callee |
| `agent_type` | 42 | coordinator |
| `output_mapping` | 8 | coordinator |
| `error_step` | 7 | **[UNVERIFIED]** — `error_step` is a *step*-level field; these are inside `config` |
| `default_kind` | ~~3~~ **0** | nothing — all three deleted by migration `586` |
| `prompt` | 1 | **[UNVERIFIED]** |
| `input_data` | 1 | `extractDataForAgent`, deprecated branch |

So the known dead-or-doubtful surface is **at most 8 steps**, and after `586` the *confirmed* dead
count is **0**. `call_agent` declares no `ConfigKeys`, so `cmd/config-key-audit`'s
declared/conditional/deprecated/removed buckets **cannot see any of this** — an empty audit here is
not evidence.

## 5. The options, costed

**(a) Do nothing; rely on the callee.** What `382` actually shipped: the callee's absent-field
branch now takes the safe default and emits a durable `MISSING_IMAGE_KIND` row naming the caller.
*Cost:* it fixes one callee. The seam still manufactures dead keys for every other `call_agent`
caller, and the next callee will not have a routing table to be loud with.

**(b) Declare `ConfigKeys` for `call_agent`, then audit.** Makes the existing
`cmd/config-key-audit` machinery able to see this seam at all, and `--removed-keys-in-use` becomes
meaningful for it. *Cost:* `call_agent` is on 59 live steps; a declaration that is wrong in the
strict direction hard-fails validation on a binary roll (`bugs_open/234`'s shape). Needs the
`error_step`/`prompt`/`input_data` residuals resolved first, which is ~an hour of reading.
**This is the cheapest option that turns the class from invisible to enumerable, and it changes no
runtime behaviour.**

**(c) Warn at resolve time on an unrecognised `config` key.** `extractDataForAgent` knows the key
set it consumes; anything else could be logged once per step. *Cost:* it is a log line, and this
estate's own record says a warning living only in a pod log is how defects survive for months
(011's council residual said exactly that). Would want the `reported_conditions`/`agent_error_log`
treatment to be worth building, which is a bigger change than it looks.

**(d) Make config keys reach the callee — i.e. make `default_kind` work.** **Rejected by the
`382` lane and recorded here so it is not re-proposed.** It widens a seam 57 live steps depend on,
turning every stray config key into payload, in order to rescue a spelling that 3 steps used and 0
now use. It converts a silent no-op into a silent *behaviour change*, which is worse.

## 6. What this RFC actually asks

Nothing urgent, and no code today. It asks a human to decide between **(a)** and **(b)**, and it
exists so that the fourth instance meets a written record instead of a fresh surprise. The seat was
explicit that this is a cost-of-not-changing signal, not a blocker.

**If it is left at (a), the honest statement is: the class is unguarded, the surface is ~8 steps as
of 2026-08-24, and it grows by addition every time someone writes a `call_agent` step.** Per the
2026-08-22 counting ruling, re-run §4's census before quoting its numbers:
`git log --since=2026-08-24 -- docs/agent_docs/sql_for_agents/` will show whether new steps have
landed since.

## 7. Cross-refs

`bugs_open/382` §7a (the case) · `docs/agent_docs/sql_for_agents/390` (the migration whose own
header both diagnosed the class and mis-stated its blast radius) · `586` (deleted the last three
dead keys) · `LANDMINES.md`, *"A key in a `call_agent` step's `config` is read by NOTHING"* ·
`bugs_open/231` (the parent class: a static config value for a field the spec was meant to supply
is dead) · `RFC_045` (an action reading a config key its own spec does not declare — the mirror
image of this, and probably the same design pass).
