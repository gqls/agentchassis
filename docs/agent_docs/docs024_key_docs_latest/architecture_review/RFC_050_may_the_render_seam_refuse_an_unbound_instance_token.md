# RFC_050 — may the shared render seam REFUSE an unbound instance token, and on which callers?

**Status:** FILED 2026-08-24 by the `bugfix_283_component_instance_scope` §10 building thread.
**Needs an owner decision.** The disputed change is **already contained** — see §1 — so this is not
a request to leave anything armed while it is discussed.

**Provenance:** council correlation `661bcf00-131d-4e4c-9815-218647812907`, round 1 → REVISE,
decided by a **gating HIGH from the `guardian` seat**, with the `architecture` seat naming this
document as the remedy in the same round. Commits: `120131549` (the change, refusal included),
`<containment commit>` (refusal → published report, everything else kept).

---

## 1. What was shipped, what was objected to, and what is contained NOW

`platform/orchestration/actions/component_library.go`'s `RenderTemplate` has, since v1.0.1304,
detected a specific defect and only written a `logger.Error` about it: a template that namespaces
its element ids with `{{.InstanceID}}`, rendered by a path that bound no token, takes
`missingkey=zero` — an empty string — so **every instance on the page lands on identical element
ids**, and `document.getElementById` hands every lookup to the first copy. On a calculator that is
a page which shows a visitor a figure computed from numbers they never entered.

Commit `120131549` converted that log into a **hard refusal**, unconditional. The guardian seat
objected, at HIGH:

> *"RenderTemplate is a shared render seam with 6 binding callers across multiple pipelines …
> Converting its log-only defect path into a hard error is new authority on a shared seam, shipped
> unconditional rather than opt-in-default-OFF. The author names this exact conflict with the
> 2026-08-02 §2 ruling in risks#1 but ships anyway on the strength of today's census. **A census of
> TODAY's live templates cannot bound tomorrow's caller**; this needs an explicit council sign-off
> on the authority question, not a footnote."*

**The seat is right, and the estate has litigated this shape once already.** `RFC_044` shipped a
default-ON annotation on two shared render actions, drew the same HIGH, and was **VETOED** on the
next round with *"routing a scope objection to architecture review does not license deploying the
disputed change"*. Re-arguing with better measurements is also what the owner ruling of 2026-07-28
forbids: *"a veto on SCOPE is not answered by resubmitting with better measurements."*

So the refusal is **withdrawn** and replaced with a **published report** —
`RenderContext.UnboundInstanceToken`, written by the seam, readable by a caller. That is not a
retreat to the log: it is the shape the owner ruling the original submission *cited* actually used
on this very function. The `editquality` seat caught that too, at medium:

> *"Diagnosis explicitly cites an owner ruling (component_library.go:1216) that this same
> logger.Error is 'not escalation.' The plan arms it into a hard refusal without acknowledging it is
> overriding that specific ruling."*

Reading it properly: `bugs_open/054` (2026-07-22) ruled *"make it MEAN something"* about a named
log — and its remedy on this same function was `ctx.AbsentRequiredFields`, **published, not
refused**, because *"the seam still cannot escalate — it has no database handle and no site
identity — so it publishes here and a caller that has both acts."* Everything true of that field is
true of this one.

**Everything else from `120131549` stands and is not in question here**: the `EmptyElementIDs`
detector class, the gate's hard-error branch (with a typed sentinel), and `instanceaudit --gate`.
The detector change means the class is no longer *silent* whichever way this RFC is decided.

## 2. The question for the owner

> **Should `RenderTemplate` refuse to render an `{{.InstanceID}}` template with no token bound —
> and if so, on all callers or only some?**

Four answers, costed:

| | answer | cost |
|---|---|---|
| **a** | **Publish only** (today's contained state) | The field has **no reader**. A published fact nobody consumes is a mechanism that rots — this estate's own repeated finding. The section editor still writes `id=""` to a live page. |
| **b** | **Refuse fleet-wide, unconditional** (what was withdrawn) | Answers the defect completely. Is new authority on a seam with **15** non-test call sites as of 2026-08-24 (`grep -rn '\bRenderTemplate(' --include=*.go platform/ cmd/`, defs and `RenderTemplateWithMap` excluded; an earlier draft of this row said 12 — miscounted, and undated, which is its own rule violation; §3's per-caller table is the load-bearing census), licensed only by a survey that cannot bound a future caller — the guardian's objection, unaddressed. |
| **c** | **Refuse where the caller opts in**, via the existing `enforce_instance_scope` key | Uses machinery that already exists (`enforceInstanceScope`, `instanceScopeEnforceConfigKey`, armed today on `tool-generator`/`tool-deployer`). But `RenderTemplate` takes no step config, so the flag must reach it through `RenderContext` — which the §10 plan explicitly forbade, on the ground that a per-caller flag re-creates the per-call-site wiring the seam exists to remove. |
| **d** | **Refuse only on the ungated live-page routes**, i.e. the two section-editor call sites | Smallest blast radius, targets the actual damage (those two write `rendered_html` to an already-live page with **no `validate_content` between**). Leaves the build and rerender paths reporting. Is still per-call-site wiring, but only two sites, both of which already refuse on a render error. |

**The building thread's recommendation is (d), then (c) if a general switch is wanted** — but this
is recorded as a recommendation, not taken.

## 3. The blast-radius data, already gathered (all 2026-08-24, all disconfirmable)

**Nothing renders unbound today.** Static, tree-wide: 11 of 972 non-test `.go` files call a
`RenderTemplate*` helper — 6 bind a token, 5 are allow-listed in `pattern-check.py`'s
`INSTANCE_TOKEN_ALLOWED` as slots occurring once per document, **0 unbound**. *Demand control:*
deleting the bind call from each of the 6 flips it to an `unscoped-component-render` finding, 6 of
6 — so the zero could have come out otherwise.

**At the artefact.** Of 2,020 `page_components` rows, **0** carry the unbound shape `id="-…"`;
**374** carry a bound `id="c-…"` (the control proving the query can find rendered ids). Sharpest
arm: 134 of the 140 active token-bearing templates spell `id="{{.InstanceID}}-suffix"`, but **6
spell `id="{{.InstanceID}}"` exactly** — `generic-text-block`, `faq`, `pricing`, `mechanism-flow`,
`evidence-timeseries`, `illustrated-text-block` — whose unbound render is `id=""`.
`generic-text-block` began spelling it at `2026-08-23 12:32:24+00`; **155 of its rows written
since, 155 of 155 bound, 0 empty.**

**Blast radius of a refusal.** 0 chrome-level templates spell the token (header 4, footer 1, site
6, element 1, all active), so no header, footer or `<head>` render could fail. All 140 that spell it
are section (30) or tool (110) level.

**The per-caller contract is NOT uniform — and the original submission asserted that it was.** This
is the guardian's second objection and it was correct. Read on 2026-08-24, of the **7** call sites
that can reach an `{{.InstanceID}}` template:

| call site | on a render error |
|---|---|
| `assemble_from_library.go:303` | returns wrapped error — step fails |
| `section_editor_actions.go:1113` (content edit) | returns error, **live section left unchanged** |
| `section_editor_actions.go:1277` (component swap) | returns error, **live section left unchanged** |
| `v3_site_actions.go:2465` | returns error — step fails |
| `component_instance_conversion.go:419` | returns error (this is the gate itself) |
| `tool_birth_instance_scope.go:67` | returns error — birth refuses |
| **`rerender_page_sections_action.go:661`** | **CARRIES the stored HTML and continues** — deliberate: this action *is* the repair vehicle, and *"a re-render that refuses on the state it was dispatched to fix would deadlock its own remedy"* |

Two further callers soften a render error and **cannot** reach such a template (chrome census 0):
`render_site_components_action.go:1074` (leaves working chrome in place unless the site has none)
and `adopt_fragment_section.go:123` (`Warn`, returns false). So under answer (b) the refusal would
have been a hard failure on 6 paths and a silent carry on the 7th.

## 4. The second door, now reported rather than merely documented

`RenderTemplateWithMap` (`rerender_pages_actions.go:818`) is an **independent render path** that
does not share this one's `FuncMap` (`bugs_open/260` §13g). It had no instance-token check at all.
The `bug_historian` seat objected, at medium, that a chrome-only census is *"a snapshot, not a
guard — nothing detects if a section/tool template ever reaches that second path in future"*.

Correct, and fixed: that path now carries the same report. It is log-only there (no `RenderContext`
to publish onto), so **whatever this RFC decides must cover both paths** or the invariant is
enforced on one and not the other — which is precisely the `016b` §9 shape (*"one call site of a
shared judgement gets the rigorous fix, the sibling stays heuristic"*) that the seat named.

## 5. Why this is not simply "add a default-OFF switch and move on"

Because the owner has already ruled against requiring one, for a stated reason. CLAUDE.md,
2026-07-29 §2:

> *"The only mechanism that actually holds a seam back is a default-OFF switch, and the owner has
> ruled we will NOT require one (its cost is a mechanism rotting unexercised, which this platform
> has been bitten by before). So: review here is after the fact, by design."*

And the 2026-08-02 §2 ruling points the other way for **new authority on a shared seam**. Both
rulings are live, both are being followed by different seats, and this change sits exactly where
they meet. **That is why it needs a human**, and it is the same collision `RFC_044` ran into.

The contained state (a) is not free either: it leaves `UnboundInstanceToken` with no reader, and a
field with no reader is the rot the 2026-07-29 ruling warns about. **If the answer is "do not
arm", the honest follow-through is to DELETE the field**, not leave it as decoration — that
instruction is written at the field's declaration so a later reader does not have to find this
document.

## 6. What happens if this is not decided

The defect stays detectable and undefended: `DetectInstanceCollisions` now reports empty ids
(so the corpus sweep catches them after the fact), the gate refuses them at conversion and at tool
birth (armed today on `tool-generator`/`tool-deployer`), and both render paths log. What no longer
happens is the render being stopped — so a caller added tomorrow that forgets to bind will ship a
page whose instances all answer to `id=""`, and we will find it in the next census rather than at
the moment it was written.

**Related:** `RFC_044` (the same shape, on the same kind of seam, vetoed) · `RFC_022` (the
narrowing that does NOT cover this: it exempts opt-in-default-OFF fields with no live consumer, and
this is neither) · `bugs_open/054` (a named log is not escalation) · `bugs_open/260` (execute or
fail, no third state — the ruling the withdrawn refusal leaned on) · `bugs_open/283` /
`architecture_review/RFC_032` §10 · plan
`docs024_key_docs_latest/bugfix_283_component_instance_scope/PLAN_2026-08-24_occurrence_derivation_and_empty_id_detector.md`
