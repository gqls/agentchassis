# 173 — a loop's error routing is all-or-nothing: one substep cannot be made recoverable without making every substep recoverable

**Filed 2026-08-01** by the `bugs_open/165` sites B+C lane, **at the direction of
the council gate's `architecture` and `constitution` seats** (corr
`c69e935a-7134-45c1-81c3-2f1da7831827`, round 2). Status: **OPEN, UNOWNED.**
Latent — nothing is failing today because of it; it is a missing degree of freedom
that pushed a bug fix into a workaround, and the seats caught the workaround.

> **VERIFICATION STATEMENT, per the owner ruling of 2026-07-31.** That ruling says
> a `bugs_open/` file asserting a cross-cutting or structural cause is not filed
> until it has been through `090_TRIGGER_needs_diagnosis_v1.sh`, or the filing
> session states plainly why it substituted equivalent first-hand verification.
> **I substituted first-hand verification, and here is exactly what it consisted
> of**, so the substitution can be judged rather than taken on trust:
>
> - I read the whole routing path in the running source, not from memory or grep
>   hits: `loop_actions.go:66` (reads `continue_on_error` from the loop step's
>   config), `loop_expansion_handler.go:39` and `:157` (propagates one loop-level
>   value onto EVERY injected iteration step), `loop_error_handler.go:71-87`
>   (`shouldContinueLoopOnError` reads it back off the iteration step),
>   `coordinator.go:907` and `:3350-3363` (`routeToErrorStepOrFail` → `failWorkflow`).
> - I confirmed the live config shape by query, not by assumption: the loop that
>   motivated this (`multipage-website-builder.workflow.steps.generate_pages_loop`)
>   has `continue_on_error` unset, and its `substeps.extract_links` has no
>   `error_step`.
> - **What makes this a safe substitution:** the claim is a *closed* one about a
>   single code path I have read end to end, not a hunt for a cause that might live
>   somewhere else. There is no "the cause is not where the symptom is" risk here
>   because there is no symptom — the finding is that a knob does not exist.
> - **What would still be worth the loop, and I have NOT done:** whether any live
>   workflow is *currently* mis-shaped because of this (i.e. loops that set
>   `continue_on_error: true` and thereby silently swallow substep failures they
>   should not). That is a fleet-wide question and is listed under "What is not
>   established" below rather than asserted.

## The defect

`continue_on_error` is a **loop-level** flag. `LoopAction` reads it once from the
loop step's config (`loop_actions.go:66`), passes it through the expansion
(`loop_expansion_handler.go:39`), and stamps the **same value onto every injected
iteration step** (`:157`). `shouldContinueLoopOnError` then reads it back per
iteration step (`loop_error_handler.go:83`).

So the only two states a loop can be in are:

| `continue_on_error` | effect |
|---|---|
| unset / false | **any** substep error fails the entire orchestration (`routeToErrorStepOrFail` finds no `error_step` and falls through to `failWorkflow`) |
| true | **every** substep error is swallowed and the iteration is skipped |

There is no way to say "a failure in *this* substep should skip the item, but a
failure in *that* one should still fail the build". `error_step` exists at step
level and loop expansion does carry it (`loop_expansion_handler.go:480` prefixes
`error_step` among the step-reference fields), so a per-substep `error_step` is
routable — but it requires naming a real recovery step, which is a different thing
from "tolerate this one and carry on".

## Why it matters — the concrete case that produced this file

`bugs_open/165` site C put a completeness floor on `extract_and_sync_links`. That
action is reachable from exactly one live agent and it is **nested**:
`multipage-website-builder` → `workflow.steps.generate_pages_loop` →
`config.substeps.extract_links`. The loop carries no `continue_on_error`.

Consequence: **a refusal on one page's link extraction fails the entire site
build.** For a table that is regeneratable, feeds no rendered output, and is
currently empty fleet-wide, that is wildly disproportionate.

The fix lane's round-2 response was to make the action never error — return a
`(detail, bool)` and skip the sync. **Four council seats ruled that backwards and
they were right**, which is the useful part of this file:

- `constitution` (HIGH): "the plan's own rationale identifies the real cause as the
  workflow's missing `continue_on_error`/`error_step` … then works around that gap
  by making the action silently skip (never error) instead of fixing the workflow's
  error-routing config. This is a fix whose rationale names the mechanism it steps
  around rather than repairs it."
- `architecture` (medium): "the action launders its own failure into a soft
  `maySync=false` skip. That is a workaround built because the core routing
  couldn't carry the load, not a property of `link_registry` itself." And: it
  "establishes a reusable pattern … with no shared vocabulary — the next nested
  substep that hits the same disproportionate-failure problem will invent its own
  ad hoc bool/sentinel rather than the engine offering per-substep
  `continue_on_error`."
- `bug_historian` (HIGH): the skip turned a refusal into a success-shaped return
  whose only signal is a queue.
- `guardian` (medium): a shared-mechanism contract change on the strength of a
  single-caller claim.

The action was reverted to the uniform error contract. **The gap is real and is
this file.**

## Why the obvious config change is NOT the answer either

Setting `continue_on_error: true` on `generate_pages_loop` would fix the
disproportion — and would change failure semantics for **every** substep in the
page-generation loop, turning a page that genuinely failed to build into a
silently skipped one. That is the same silent-drop class this platform keeps
finding, moved rather than removed. It is a workflow-owner's decision and it
should not be a rider on someone's bug fix, which is precisely why it is filed
here instead of done.

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Per-substep `continue_on_error`.** Let a substep declare its own tolerance,
   defaulting to the loop's value. Smallest change that removes the false choice:
   `loop_expansion_handler.go:157` already writes the flag onto each injected step,
   so the plumbing is one lookup — read `substep.Config["continue_on_error"]` first
   and fall back to the loop's. Closes the door because the shape "one tolerant
   substep in a strict loop" becomes expressible rather than requiring an action to
   lie about its own outcome.
2. **A named severity on the substep** (`on_error: fail | skip_item | continue`)
   instead of a bool. More expressive, more vocabulary to agree, and it is a shared
   mechanism change — so it wants an RFC rather than a bug patch.
3. **Leave it, and require a recovery step.** Per-substep `error_step` already
   works; document it as the answer and accept that every tolerant substep needs a
   real step to route to. Cheapest, but it is the option that leaves the next
   thread inventing a bool.

## What is NOT established, marked so nobody inherits it as a finding

- **Whether any live workflow is currently mis-shaped by this.** Loops that DO set
  `continue_on_error: true` are, by this same all-or-nothing property, swallowing
  every substep failure — including ones that should fail loudly. **[UNMEASURED]**
  I did not survey them. The query to start from:
  ```sql
  SELECT a.type, s.key, s.value->'config'->>'continue_on_error' AS coe,
         jsonb_object_keys(s.value->'config'->'substeps') AS substep
  FROM agent_definitions a, LATERAL jsonb_each(a.default_config->'workflow'->'steps') s
  WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
    AND s.value->'config' ? 'substeps';
  ```
- **Whether per-substep granularity has been raised in a prior council round.** The
  `architecture` seat flagged that it could not confirm this from its tier and
  asked for it to be checked. I have not checked it either. **[UNVERIFIED]**
- Whether candidate 1 interacts with the retry/`fallback_step` fields that
  `loop_expansion_handler.go:480` prefixes alongside `error_step`. Not looked at.

## How to verify a fix

The mechanism is inert until induced, like every guard in this family, so a green
build proves nothing. Induce it: give a loop two substeps, make the *tolerant* one
fail and confirm the iteration is skipped while the orchestration continues; then
make the *strict* one fail in the same loop and confirm the orchestration FAILS.
Both branches, or the flag is untested in the direction that matters.

## Provenance

- Council corr `c69e935a-7134-45c1-81c3-2f1da7831827`, round 2 — `constitution`
  (HIGH, edit 4), `architecture` (medium ×2, edit 4), `bug_historian` (HIGH,
  edit 4), `guardian` (medium, edit 4).
- The case that produced it: `bugs_open/165` site C,
  `platform/orchestration/actions/link_registry_prune_floor.go`.
- Code: `loop_actions.go:66`, `loop_expansion_handler.go:39,157,480`,
  `loop_error_handler.go:71-87`, `coordinator.go:907,3350-3363`.
