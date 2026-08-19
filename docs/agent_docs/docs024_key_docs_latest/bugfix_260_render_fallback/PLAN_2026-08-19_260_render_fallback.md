# PLAN 2026-08-19 — make a component render either execute or fail, with no third state

Design, phasing, decisions **and their reasons**. Corrections to the originating brief live here,
marked as corrections.

Scope: `bugs_open/260`'s **renderer half**. The writer-output half is `copy_quality_two_stage`'s
(owner direction 08-12); the dead-link consequence is `bugs_open/328`'s class.

---

## The design in one line

**Delete the regex fallback and give the render seam an error channel.** Every call site then
makes an explicit, reviewed failure decision, and a schema type-checker supplies the actionable
diagnosis — as an opt-in pre-render refusal on the two gated paths, and unconditionally as an
*enricher* on any render that has already failed.

Ranked by what makes the bad state unrepresentable:

1. **Hard error at the seam.** The renderer can no longer emit output it did not execute.
2. **Per-caller failure behaviour.** The error lands safely on each of the 13 paths.
3. **Pre-render type gate.** Earlier failure and a named cause; opt-in, default OFF.
4. **Blocker-cap honesty** (`FindAllString(html, 10)`) — separate small commit, not in the
   council's edit budget.

### Why the hard error satisfies the brace-literal constraint by construction

The design contains **no output brace-scan at all.** It fires exactly when `text/template`
returns an error, so content that merely *contains* `{{ }}` — a prompt library, a syntax gallery,
the tool pages the owner has ruled every site should be able to have — never touches it. This is
the constraint satisfied structurally rather than by heuristic, which is the whole reason not to
reach for a detector. It also argues **against** tightening `checkUnrenderedTemplates`: with the
seam failing loud, no leaked HTML reaches that detector at all.

## Edit 1 — the seam (`component_library.go`)

```go
func RenderTemplateReportingMissing(templateStr string, ctx *RenderContext, logger *zap.Logger) (string, []string, []string, error)
func RenderTemplate(templateStr string, ctx *RenderContext, logger *zap.Logger) (string, error)
```

On `executeGoTemplate` error, return the error wrapped; delete the fallback block entirely.
`executeGoTemplate` itself is unchanged — its wrapped errors already carry line, field and the
offending value (`range can't iterate over "…"`). Keep the `form_action` seeding, the InstanceID
report, `missingBareFields` and the `<no value>` strip; skip the missing-fields report on error,
there being no output to report against.

**Decision, and the rejected alternative — name it in the submission.** Keeping
`RenderTemplate(…) string` and returning `""` on error was rejected: it is a *second silent
shape*. `assembleComponents` would stitch a page missing a section; `GateConvertedTemplate` would
gate an empty render. **The compile-breaking signature is the feature** — it forces every
caller's decision into the diff, which is the review surface the council actually needs.

**Dies in the same commit** (each verified to have no other caller):
`RenderTemplateWithValidation` (dead), `validateContentAgainstSchema`, `contextToMap`,
`renderEachBlocks`, `renderIfBlocks`, `renderGoStyleSubstitutions`,
`renderHandlebarsSubstitutions`, `applyBidirectionalAliases` (already caller-less), and the
fallback-only `{{nav_items_html}}` / `{{quick_links_html}}` replacements.

Two consequences worth writing down:

- **`contextToMap` dying takes `bugs_open/203`'s two remaining fabricate-a-URL members with it**
  (`primary_cta_url → "/contact.html"`, `secondary_cta_url → "/about.html"`). ⚠ **The bug file's
  §5 claims this happens "as a side effect" of deleting the fallback — it only happens if
  `RenderTemplateWithValidation` is deleted in the SAME commit**, because that is `contextToMap`'s
  other caller. 203 stays OPEN until the image rolls (fixed-and-live bar).
- **A template naming `{{nav_items_html}}` cannot parse as a Go template** ("function not
  defined"), so after this change it hard-fails rather than being regex-patched. Measured 0 of
  253 use it — but the retirement must be in the register entry so nobody re-authors one.

Tests to **rework, not silently delete**: `component_library_cta_url_fallback_test.go`,
`component_library_form_action_test.go`, `render_context_current_page_test.go`,
`render_context_derivation_test.go`.

## Edit set 2 — the call sites, each with its decided behaviour

| # | call site | path | on render error |
|---|---|---|---|
| 1 | `v3_site_actions.go:2273` (`RenderComponentAction`) | **build** — where 25 of 26 events flow | **Fail the step**, same shape as the `missingRequiredLLMFields` refusal 20 lines above. Append the type diagnosis. The parked item then carries the cause, not 20 capped symptoms |
| 2 | `rerender_page_sections_action.go:600` | rerender — the repair vehicle | **Carry the stored HTML** using the path's existing carried-section machinery. Never replace good stored bytes with a failed render; never fail the whole page — that would deadlock the path's own remedy |
| 3 | `render_site_components_action.go:889` | **chrome store — no gate downstream** | **Do not store**; the existing row keeps serving. First build with no existing row → fail the step: a site must not go live with a silently missing chrome slot |
| 4–6 | `component_library.go:1969/2038/2311` (`RenderHeader/Footer/Head`) | chrome | **Return the error.** `InjectHeader` (:2143) *already* answers an error with `RenderFallbackHeader`; the page gets well-formed fallback chrome carrying its `<!-- SOURCE: fallback -->` marker. Cheapest correct behaviour, no new machinery |
| 7 | `rerender_pages_actions.go:532` (head) | legacy whole-page rerender | Take the **existing** fallback-head else-branch |
| 8 | `assemble_from_library.go:297` | library assembly | **Fail the step.** No writer data is in scope, so an execute error means a corrupt template. A silently missing section is the 018 class |
| 9–10 | `section_editor_actions.go:886/996` | **live-page edit/swap** | **Return the error**; the edit is refused and the live row untouched. Closes the ungated route |
| 11 | `component_instance_conversion.go:271` (`GateConvertedTemplate`) | 283-lane acceptance gate | **Propagate → the gate refuses.** Today it can green-light a template the real renderer cannot execute |
| 12 | `cmd/component-render-check/rendercheck.go:653,680` | offline audit CronJob | Route to the existing `unanalysed` bucket. **Verify baseline noise = 0 with a live run before commit** — predicted zero, but predict-then-measure |
| **13** | **`rerender_pages_actions.go:782` (`RenderTemplateWithMap`)** | **contact-info rerender** | **Return the input HTML unchanged.** See below — this one is not in the bug file at all |

### ⚠ The 13th seam, verified here and recorded as §13g of the bug file

`RenderTemplateWithMap` is a **second, independent** Go template executor with **no FuncMap and
no `missingkey=zero`**, returning `""` on parse or execute error at `Warn`. Its caller does
`html = contactInfoRe.ReplaceAllString(html, renderedContactInfo)` — so an error **deletes the
live contact block**, which is worse than mangling it and leaves nothing for any detector to
find.

`[MEASURED 2026-08-19]` **Latent, not firing**: the one active `contact-info` component renders
clean there (2,453 bytes). Probed with that seam's *own* configuration and controlled both ways —
`{{safe .x}}` must fail there (it did, proving the probe faithful to the absent FuncMap), a plain
`{{.email}}` must pass (it did). **But the trigger is one ordinary edit away**: adding `{{safe}}`,
`{{default}}` or `{{isset}}` — normal everywhere else in this library — makes it fail to *parse*
here. The two seams do not accept the same language and nothing says so.

## Edit 3 — the type gate

**Checker (new, pure):** `datahelpers.ContentTypeViolations(fields, content) []TypeViolation`,
beside `SchemaContentFields`, returning `{Path, Declared, Actual}` — e.g.
`steps[2].branches / array (items: object) / string`. **Recursion into `items` is mandatory**:
the live case is a *nested* violation, so a top-level-only check would miss the very instance
that motivated this.

Rules, deliberately conservative:
- declared `type IN ('array','list')` — `list` is real (5 fields) — present, non-nil, non-`[]`
  → must be `[]interface{}`;
- `items` declaring object shape → each element a `map[string]interface{}`; recurse for nested
  arrays;
- **absent / nil / empty are NEVER violations** (§2's table); unknown or undeclared types are
  skipped, never guessed. Only `source: "llm"` fields — resolver fills already have
  `resolvedValueSatisfiesDeclaredType`.

**Two modes with different authority, and the split is the point:**

1. **Enricher — unconditional.** At sites 1 and 2, when the render has *already failed*, append
   the violation list to the error. It adds no authority (it fires only on failure), so it needs
   no opt-in key. It turns `range can't iterate over "…"` into
   `component=mechanism-flow field=steps[2].branches declared=array(items:object) actual=string`.
2. **Pre-render refusal — opt-in, default OFF**, key `refuse_mistyped_llm_fields`, read exactly
   as `refuse_dead_url_controls` is. **Why opt-in when the presence gate beside it is
   unconditional:** the checker keys on *schema*, not template, so a mistyped field the template
   never references builds fine today — an unconditional refusal is therefore **new authority
   over content that currently renders**. That is precisely the owner ruling of 2026-08-02 §2,
   and the dead-URL guard's own history (three seats independently demanded default-OFF for this
   shape) is the local precedent. **The hard error remains the complete detector; the gate is the
   early, named one.**

Arming: `sql_for_agents/NNN_bugfix_260_arm_type_gate_HOLD.sql` patterned on `380`, asserting the
step count so a future third render step fails loudly. **Image first, then seed.**

RFC_022 housekeeping: add the key to `RenderComponentInputSpec` and
`RerenderPageSectionsInputSpec`, run `scripts/audit-optional-key-budget.sh` and the cron parity
test, and **enumerate the consumers in the submission — asserting it without the query is itself
the objection.** By the RFC_022 ruling an opt-in field whose unsafe default is OFF and which no
live consumer names is *not* architecture-scope.

**Coverage honesty for the register entry:** 107 of the 110 exposed components carry a `fields`
schema, but **75 of 253 actives declare no schema at all** — a green gate is not fleet coverage
and must not be reported as one.

## Explicitly NOT in this change

**Coercion / bounded repair** (string → `[{body: s}]`). It belongs to the writer-output half by
owner direction, and two further reasons: repairing at render time **permanently silences the
only measure of the writer's violation rate** (the your-own-action-silences-your-detector class),
and it rewrites writer output at a layer that cannot see the page. If that lane builds it, it
composes cleanly *in front of* this gate.

## Regression set — fixtures are live row bytes

- **A (negative):** real `mechanism-flow` template + real string-`branches` data → **error**
  naming the field; no output.
- **B (control, mandatory):** same with `branches` coerced to the declared shape → renders, no
  `{{` in output. **A fix that makes both fail is not a fix.**
- **Positive control for the brace constraint (required):** `body = "Use {{ variable }} in your
  prompt"` through `{{.body}}` → **passes**, braces present in output; plus a fixture shaped like
  the live benign row. Pins that the design never confuses renderer failure with brace-bearing
  copy.
- **Parse-error case:** `{{if $.x}}` inside a CSS comment (the LANDMINE's case) → error, not
  degraded output.
- **Checker units:** the nested `steps[2].branches` violation with correct path/declared/actual;
  absent / nil / `[]` → no violation; array-of-strings under `items:object` → violation; `list`
  covered; schema-less component → nil.
- **Call-site behaviour:** rerender carries stored bytes; chrome store refused with the row
  untouched; `InjectHeader` output contains `<!-- HEADER SOURCE: fallback -->`;
  `GateConvertedTemplate` refuses; the contact-info caller leaves HTML unchanged.
- **Mutation proof (run, do not commit):** re-add a one-line fallback, or swallow the error —
  tests A, the parse-error case and the gate test **must** fail. A quiet pass otherwise proves
  the tests do not guard the seam.
- **Live, after the roll:** per-service provenance check; §7's estate sweep with its positive
  control in the same query; the loanzy lane's greenfield build.

**Success criterion, stated so it cannot be misread: the build fails EARLY naming `branches` —
not "the page builds."** The 24 parked items still hold mistyped content; building them is the
writer half's job.

## Corrections this plan makes to the bug file and to my own brief

1. **"The page-BUILD path has no schema gate at all" is FALSE at HEAD** — two callers, including
   `v3_site_actions.go:2252` inside the build render step. The build path has a *presence* gate
   and lacks the *type* half. Corrected in §5 of the bug file; it was the sentence a reviewer
   would check first.
2. **Candidate 1 as the file states it under-specifies.** "Delete the regex fallback" without an
   error channel yields `""`/partial output — a new silent shape.
3. **Chrome needs error *plumbing*, not new mechanism** — `Inject*` already owns a complete safe
   fallback ladder, and `resolvedValueSatisfiesDeclaredType` is a live precedent for refusing a
   declared-type violation on this estate.
4. **Line drift throughout** (`:2179`→`:2273`, `:333`→`:396`). Re-cite at commit time;
   `render_content_envelope_guard.go:13` carries a stale citation too, so code comments are not
   a safe source here either.

## Process

One council round for the whole coherent task, `Council-Submitted:` on the commit, register entry
**in the same commit** (condition 2), no ordering-constraint claim, consumers **told** not merely
measured, LANDMINES corrections appended then `landmines-verify-dispatch.sh`. Go change is inert
until the roll; the arming migration is `_HOLD` and runs after.
