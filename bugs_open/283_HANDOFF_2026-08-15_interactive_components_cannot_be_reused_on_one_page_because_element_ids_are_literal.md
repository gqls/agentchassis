# 283 — Interactive components cannot be instantiated twice on one page: their element ids are literal, and the platform's own fix does not work on both render paths

**Filed 2026-08-15.** Status: **OPEN — needs an OWNER DECISION before any code is written.**
Raised out of the `loanandmortgagecalculator.co.uk` lane (`bugs_open/263`, Track B2), where
the reuse demonstration in that lane's handoff §6 item 6 could not be done the obvious way
and had to be run on its own throwaway page instead.

---

## 1. What the thing is, in plain terms

A **component** is a reusable block of a page — a calculator, an FAQ, a text section. It is
stored once, as an HTML template, and a page uses it by pointing at it. The whole promise is
that one component can appear on many pages, and more than once on the same page, each time
with its own copy.

Inside an interactive component's HTML, individual elements carry an **id** — a name like
`loanAmount` or `btn-calculate`. The component's JavaScript uses those names to find the
elements it needs to work with: read what the visitor typed into `loanAmount`, write the
answer into `displayMonthly`.

**An id is required to be unique across the whole page.** That is not a house style, it is
how HTML works: `document.getElementById('loanAmount')` returns **one** element — the first
one it finds — and silently ignores any others.

## 2. The defect

The calculator components hardcode their ids as fixed text. Put two of them on one page and
both copies claim the same names, so the browser hands every lookup to the first copy.

The second calculator on the page is then **not broken in a way anyone would notice**. It
renders perfectly. Its fields accept typing. Its button responds. But it reads the *first*
calculator's inputs and writes into the *first* calculator's results area — so the visitor
types into one box and watches a different box answer, or sees nothing happen at all.

On a consumer-credit site, that means a page that can quietly show someone a repayment figure
computed from **numbers they did not enter**. That is the reason this is filed rather than
left as a note: the failure produces a plausible wrong number, not an error.

## 3. Evidence — all measured 2026-08-15, live

**The specific component the lane hit** (`mortgages-repayment`, freshly converted to B2):

- **9 literal ids in the template**, none templated:
  `loanAmount`, `interestRate`, `termYears`, `btn-calculate`, `resultsArea`,
  `displayMonthly`, `displayTotalInterest`, `displayTotalRepayable`, `amortizationTable`
- **7 `getElementById` calls**, every one a bare literal string.

**Fleet-wide, across 240 active components:**

| measure | count |
|---|---|
| components hardcoding at least one element id | **173** |
| of those, ids are **literal only** (no template expression) | **166** |
| components that template their id (per-instance capable) | **7** |
| components that bind by `getElementById` | **100** |

**Reuse already happens — so this is not hypothetical demand.** Pages carrying the same
component more than once, live right now:

| domain | page | component | instances |
|---|---|---|---|
| loancalculator.co.uk | index | `ported-prose` | **4** |
| loancalculator.co.uk | tool-damage-checker | `ported-prose` | **4** |
| ai-agent-orchestration.com | containment-first-architecture | `generic-text-block` | 3 |
| leopardessconsulting.co.uk | how-it-works | `generic-text-block` | 3 |
| gaswholesalers.com | pricing-transparency | `generic-text-block` | 2 |

**Why nothing is on fire today.** Every component in that list is prose. `ported-prose` has
**0 ids**; `generic-text-block` has 1, and it is *templated*, not literal. **No component
that binds by `getElementById` is currently instantiated twice on any page.** So the defect
today is a **wall, not a fire**: it blocks reuse of interactive components rather than
corrupting a live page. It becomes a fire the moment someone places two calculators on one
page, which is an ordinary thing to want and which nothing currently prevents.

## 4. ⚠ The obvious fix is already in the estate — and it does not straightforwardly work

The platform has a convention for exactly this: a template writes `id="{{.ComponentID}}"`
and the renderer substitutes a value. Seven components use it — `faq`, `generic-text-block`,
`mechanism-flow`, `evidence-timeseries`, `pricing`, plus `product-grid` and
`category-listing` on their own domain fields.

**The trap: the two render paths substitute different things, and only one of them is unique
per instance.**

- `platform/orchestration/actions/assemble_from_library.go:277`
  ```go
  componentID := fmt.Sprintf("component_%s_%d", comp.Function, idx)
  ```
  Includes the loop index → **unique per instance on the page.** Reuse-safe.

- `platform/orchestration/actions/v3_site_actions.go:2055`
  ```go
  renderCtx.ContentData["ComponentID"] = comp.ID
  ```
  `comp.ID` is the **component's** id from `content_components` — **the same value for every
  instance of that component.** Two instances collide exactly as before.

They are not even the same shape: one is `component_<function>_<n>`, the other a UUID
(`uuid.Parse(comp.ID)` at `v3_site_actions.go:2106`). **So "just use `{{.ComponentID}}`" is
a fix on one path and a no-op on the other**, and which path renders a given page decides
whether it works.

> **[UNVERIFIED]** Which of these two paths renders the LMC calculator pages on a
> `page_rerender`. It does not change the diagnosis — those templates use literal ids, so
> they collide on either path — but it is load-bearing for fix candidate A and must be
> established before that fix is trusted. Do not assume from this file.

## 5. The decision the owner needs to make

Not "should we fix it" — **"how much do we want reuse to be a real property of the
platform, given 166 components would need touching to get there."**

The candidates, ordered by **what closes the door** — i.e. by how much they make the broken
state impossible to represent, rather than merely currently-absent:

**A. Namespace ids per instance, and scope the script to its own block.**
Ids become `{{.ComponentID}}`-prefixed; the script stops calling `document.getElementById`
and instead searches within its own root element. Closes the door properly: with lookups
scoped to the instance, a duplicate id cannot reach across to another copy even if one
survives. **Cost:** touches the 100 components that bind by id, and depends on resolving §4
first. This is the only candidate where "two on a page" becomes ordinary rather than
dangerous.

**B. Same as A, but only for components someone actually wants twice.**
Convert on demand. **Cost:** low now. **What it gives up:** the estate stays in a state where
the wall is invisible until someone hits it, and the person who hits it is whoever placed two
calculators on a page — after it has rendered.

**C. Refuse the bad state instead of fixing it.**
Leave the templates alone; make it an error to place two instances of an id-bearing component
on one page — enforced where sections are saved, not in a doc comment. **Cost:** low, and it
converts a silent wrong answer into a loud refusal. **What it gives up:** reuse of exactly
the components where reuse is most valuable. Honest, and a genuine narrowing of the estate's
promise — worth pairing with A or B rather than shipping alone.

**D. Do nothing, record it.**
Defensible today, because no live page is broken. **The cost is that this file is the only
thing standing between the estate and a page that shows a stranger a repayment figure
computed from someone else's inputs** — and files do not stop anyone.

A recommendation, since one is owed: **C now, A for the calculators, B for the rest.** C is
cheap and removes the silent-wrong-answer class immediately; A pays for itself where the
numbers are consumer-credit figures; B stops us rewriting 166 templates for a property most
of them will never exercise.

## 6. How to verify any fix

The gate is that the **second** instance computes independently — a check that must be able
to fail:

1. Put two instances of one interactive component on a test page with different copy.
2. Enter **different** numbers in each. The second must answer from its own inputs.
   Assert both answers; a test that only reads the first instance passes while blind.
3. Assert **zero duplicate id attributes** in the assembled HTML.
4. Run the arithmetic oracle against the **reused** instance, with its expectation control —
   a mutation control here is not optional, because "both instances show the same number" is
   indistinguishable from a correct answer when the inputs happen to match.
5. Do it on an unlinked `noindex` page and retract it afterwards, per the owner's
   2026-08-15 ruling on the reuse demonstration.

**Precedent, already run:** the LMC lane did steps 1–5 on a single reused instance
(second row on the `mortgages-repayment` component, own page, arithmetic 12/12 plus the
expectation control, then retracted; delete commit `10cbd6116`, live 404 confirmed against a
200 control). It demonstrated reuse **on its own page**, which is precisely the workaround
this bug is about — it did not put two instances on one page, so it did not test this defect.

## 7. Provenance of this filing

**This is first-hand verification, not a diagnosis-loop run**, and CLAUDE.md's 2026-07-31
ruling requires me to say so plainly rather than omit it. What I did: read the live component
templates and their scripts out of the database; counted the literal/templated/`getElementById`
split across all 240 active components; enumerated every live page carrying a component more
than once and checked whether those components carry ids; and read both `{{.ComponentID}}`
injection sites in the Go source.

What that does **not** cover is §4's `[UNVERIFIED]` item — which render path serves these
pages. **If the owner wants the structural claim graded before anyone writes code, that is
the `090` needs_diagnosis trigger's job**, and the sharpest symptom to file is the
two-injection-sites disagreement, not the id collision itself: the collision is settled, the
consequence for the fix is not.

**Related:** `bugs_open/263` (Track B2, the lane this came from) ·
`docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk/` NOTES (d)–(g).

---

## 8. UPDATE 2026-08-15 — owner ruled; halves one and two BUILT; §4's UNVERIFIED is CLOSED; and §5 was wrong about the number of collision classes

**Owner ruling, same day:** *"I'd like reuse to be a genuine property of the platform
because if we chose to list all the calculators on one page we'd hope it would work."*
So candidate **A + C**, not the staged B. Work commits: `03c1b0b90` (per-instance token +
detector + register CLC-014), `9372a82c3` (gofmt + index row), `1e19aa6ab` (the guard).
All three carry `Council-Submitted: 07635a2f-3605-4e67-9a6d-7636b07f16ca`.

### 8.1 §4's `[UNVERIFIED]` is closed, and the answer is the unsafe one

The LMC pages re-render through `page-rerender` → `rerender_page_sections`, whose render
site is `rerender_page_sections_action.go:494`: `rc.ContentData["ComponentID"] = comp.ID`.
That is the **content component row id — shared by every instance.** So on the path that
actually serves these pages, `{{.ComponentID}}` namespaces nothing, exactly as §4 warned it
might. Also established: `storedSection` (`:103`) does **not** load `page_components.id` at
all — it carries `componentID, slotName, contentData, renderedHTML, position`.

### 8.2 §2 UNDERSTATED THE DEFECT: there are THREE collision classes, and ids are only one

This is the correction that matters most, because a fix addressing only §2 would leave two
of them live:

| class | mechanism | affected |
|---|---|---|
| element ids | `getElementById` returns the first match | 22 of 22 calculators |
| global function names | the second `<script>`'s `function calc()` **replaces** the first | 16 declare at top level |
| `window.onload` | a **single slot** — last assignment wins, so every earlier component **never initialises** | 8 (and those 8 are the only components fleet-wide that assign it) |

Classes 2 and 3 have nothing to do with ids. "Namespace the ids" would have produced a
page where the second calculator has unique fields and still does not run.

### 8.3 And the owner's case is worse than a repeated component

`btn-calculate` is a **single id shared by NINE different calculator components**; `price`,
`rate` and `years` are each shared by two, and in those cases both components bind them with
`getElementById`. So listing *different* calculators on one page collides — a check scoped to
"the same component twice" would report clean on exactly the page the owner described.

### 8.4 ⚠ A DEFECT IN THE FIX AS COMMITTED — read this before converting any template

The token is `c<position>` from `page_components.position`, chosen because position is
unique per page and stable across re-renders. **Measured after committing: the LMC tool slot
sits at position 0 on 7 pages and position 1 on the other 16.** So the same component renders
**different element ids on different pages**, and any selector addressing it is coupled to the
page's section order.

That has a concrete cost already visible: `oracle.py` — the lane's arithmetic proof, **170
checks** — addresses every calculator by literal CSS id (`#loanAmount`, `#displayMonthly`).
Under a position-derived token it would need **per-page** knowledge rather than one prefix
per tool, and it would silently break whenever a page's sections were reordered.

**Recommendation, and it should land WITH the template conversion rather than before it:**
derive the token from the component **function plus an occurrence index within the page** —
`c-mortgages-repayment` for the first occurrence, `c-mortgages-repayment-2` for a second.
Stable under reordering; identical on every page for a single-instance component, so the
oracle takes **one mechanical prefix per tool** instead of a per-page map.
`rerender_page_sections` already loops over the full section list, so the occurrence index is
computable there. Nothing is broken by the committed version in the meantime — no live
template references `{{.InstanceID}}`, so it is inert.

### 8.5 What is NOT done

**The 22 calculator templates are unconverted**, so the defect is still live and this bug
stays OPEN. That work needs, in order: the §8.4 token change; a converter that namespaces
ids, scopes lookups to the instance root, wraps each script in an IIFE, and replaces
`window.onload` with a scoped listener; the oracle's selectors updated **in lockstep**; and
a two-instance proof page. It changes the rendered bytes of 22 live pages, so it also ends
the byte-identical property the lane currently verifies against — `b2_verify`'s "verbatim
render" check will need rebaselining, and that is a deliberate decision, not an accident to
be discovered mid-batch.

---

## 9. UPDATE 2026-08-16 — the council's REVISE is acted on, §8.4's defect is fixed, and the fix now has a control that does not go stale

Commit `32d6e980a`, `Council-Submitted: 07635a2f-3605-4e67-9a6d-7636b07f16ca` (round 2 of the
same correlation, so the trail accumulates). **283 stays OPEN: the 22 calculator templates are
still unconverted, so the defect is still live.**

### 9.1 The token rule is settled: component FUNCTION + OCCURRENCE

§8.4's recommendation is implemented. `c-mortgages-repayment`, then `c-mortgages-repayment-2`
for a second copy on the same page. `InstanceToken(function, occurrence)` is the rule;
`InstanceCounter` is the one derivation, walked in position order by every path that can see
the whole page.

Three candidates were on the table and the deciding question was not uniqueness — it was what
a **selector** has to know:

| candidate | unique within a page | same on every page | verdict |
|---|---|---|---|
| `position` (shipped 08-15) | yes — measured, zero duplicates fleet-wide | **no** — the LMC tool slot is position 0 on 7 pages, 1 on the other 16 | rejected |
| `page_components.data_uuid` | **provably** — 1,580 rows, 1,580 distinct | **no**, by construction — it is per row | rejected |
| function + occurrence | derived, not provable | **yes** for a component appearing once | **chosen** |

`oracle.py` addresses all 170 of its checks by literal CSS id. Under either rejected candidate
it needs per-page knowledge of every tool; under this rule it needs one prefix per tool. The
cost is that uniqueness is **derived from the page's ordered section list** rather than read
off a unique column — and `DetectInstanceCollisions` is what pays it.

**This was free to change because nothing consumes it yet: measured 2026-08-16, 0 of 243
active component templates reference `{{.InstanceID}}`.** That query is the reason the shape
could be revised at all, and it is the one to re-run before revising it again.

### 9.2 The council's objections, and what each one turned into

- **reuse_agent, HIGH — "the same key under a weaker guarantee".** Accepted in full.
  `InstanceTokenFromSlot` is **deleted**. The two paths that cannot see the whole page
  (`RenderComponentAction`; the section editor) now supply **occurrence 0 to the same rule**
  rather than deriving a token of their own. That is a possibly-wrong *input* to one rule, not
  a second rule with a second guarantee — and a wrong occurrence produces a **collision**,
  which the detector already reports. `BindSingleSectionInstanceToken` is the seam, and its
  agreement with the canonical token is asserted by a test, not by a comment.
- **bug_historian, MEDIUM — "three call sites patched, the mechanism left generic".** Right
  about the class, **wrong about the members, and that is the whole lesson.** It named five
  files; **four of them call no `RenderTemplate*` helper at all.** Measured across the whole
  repo: **8 non-test files, 14 calls.** The real gap was `section_editor_actions.go` — two
  sites, page-embedded, binding nothing — which was on **nobody's** list, including mine. So
  the answer is not a better list (see 9.4).
- **prior_art_librarian, MEDIUM — `data_uuid`.** Checked and rejected on the table above. The
  seat was right that it was unchecked; the check changed the argument, not the answer.
- **guardian, LOW — "no end-to-end test".** Added:
  `TestRenderLayer_twoInstancesOnOnePageGetDifferentIDs` drives template → binding →
  `RenderTemplate` → detector, with a mutation that renders the same template through a path
  binding nothing (which must collide).
- **editquality + guardian, HIGH — "does not compile".** A submission artefact, fixed by
  including every named symbol in round 2's sketches.

### 9.3 What is bound where, measured — the census the council asked for

| file | sites | before | now |
|---|---|---|---|
| `assemble_from_library.go` | 1 | hand-written map write | `BindInstanceToken` + counter |
| `rerender_page_sections_action.go` | 1 | hand-written map write | `BindInstanceToken` + counter |
| `v3_site_actions.go` (`RenderComponentAction`) | 1 | slot-derived token | `BindSingleSectionInstanceToken` |
| `section_editor_actions.go` | 2 | **nothing — the real gap** | `BindSingleSectionInstanceToken` |
| `component_library.go` (header/footer/head) | 3 | nothing | allow-listed: chrome, one per document |
| `render_site_components_action.go` | 1 | nothing | allow-listed: chrome slots |
| `rerender_pages_actions.go` | 1 | nothing | allow-listed: `<head>` only |
| `cmd/component-render-check/rendercheck.go` | 4 | nothing | allow-listed: offline lint, writes `doc_notes` (:507), never a served page |

### 9.4 The durable half — why this is not another list

Two censuses of these call sites went stale **inside one council round**: the council's, and
mine. Mine grepped `platform/` and `internal/` and missed `cmd/component-render-check`
entirely, which is a scope error in the *question*, not a slip in the reading.

So the control is mechanical:

1. **The shared render layer reports it.** `RenderTemplateReportingMissing` now logs at
   **Error** when a template references `{{.InstanceID}}` and no token was bound. It reports
   and does **not** substitute: this layer cannot see the page, so any token it invented would
   either collide (no better than empty) or **disagree with the token the page's other paths
   use for the same instance** — which is worse than empty, because the ids would then depend
   on which action last touched the section.
2. **`scripts/pattern-check.py` refuses a new unbound render call site.**
   `check_unscoped_component_render` fires on any changed non-test `.go` that calls a
   `RenderTemplate*` helper and calls neither binding seam, unless allow-listed with a measured
   reason. **Proven on the motivating case before being trusted: 4 findings at HEAD, 0 now.**
   It matches the **call**, not the argument's name — an earlier draft required the argument to
   be spelled `htmlTemplate`, which is the same staleness one rung down and would have missed a
   new site passing `tpl`.

### 9.5 Deploy status — the §2 warning in the CONTINUE_HERE is CLOSED

The halves committed on 08-15 **are live**. Chain, each link checkable:

```
pods  -l app=agent-chassis   imageID   docker.io/aqls/agent-chassis@sha256:f208f01d…
local aqls/agent-chassis:v1.0.1303     RepoDigests  ["…@sha256:f208f01d…"]   ← same bytes
   docker image inspect … --format '{{index .Config.Labels "org.opencontainers.image.revision"}}'
                                        → 5e075a6f949bbb08d164bc1293b8c990068d917f
git merge-base --is-ancestor 03c1b0b90 5e075a6f9   → yes (and 9372a82c3, 1e19aa6ab, 06a74ba7d)
```

**The digest match is the load-bearing step, not the label.** A local tag can be rebuilt at any
time by any session; it is only evidence about the running pod once its `RepoDigests` equals
the pod's `imageID`. This works where the two recipes in the CONTINUE_HERE failed — the
`build provenance` startup line had scrolled out of `--tail=20000` on both pods, and grepping
the binary for a candidate sha only ever confirms a **guess** (the binary carries its own build
stamp, not its ancestors).

**Round 2's code is committed but NOT yet built or rolled**, so the section-editor binding and
the pattern-check are inert in production until the next chassis release.

### 9.6 Still not done — unchanged from §8.5 except that the token question is now answered

The 22 calculator templates are unconverted. That work needs: a converter that namespaces ids,
scopes lookups to the instance root, wraps each script in an IIFE (16 declare at top level) and
replaces `window.onload` (8 assign it); `oracle.py`'s selectors updated **in lockstep**; a
two-instance proof page; and `b2_verify`'s verbatim-render baseline reset, because converting
ends the byte-identical property that lane verifies against. DB writes need the owner.

---

## 10. UPDATE 2026-08-16 (later) — round 2 is **APPROVED and LIVE**, and the six advisory objections are worked through

`decision: approved` at 10:25 UTC, *"approved with 6 advisory objection(s) — none high-severity"*.
Reviewers: **approve** — constitution, mission, debug_historian, tooling_provenance, architecture;
**object (advisory)** — editquality, bug_historian, reuse_agent, guardian, render_guardian,
prior_art_librarian; 6 abstained. **283 still stays OPEN: the 22 templates are unconverted.**

**It is live.** Chassis `v1.0.1304`, pods started 2026-08-16T10:41 UTC. Same digest chain as §9.5:
pod `imageID` `sha256:75ae5902…` == local `v1.0.1304` `RepoDigests`, whose
`org.opencontainers.image.revision` is `5de6cddbe`, of which `32d6e980a` is an ancestor. So the
section-editor binding, the shared-layer report and the new token rule are all running in
production — **and still inert, because nothing consumes the token.**

### 10.1 The two "this needs a code_check, not an assumption" objections — both checked, both fine

The `editquality` seat was right that the *sketch* did not evidence either claim. Both are true in
the code, and here is the evidence rather than the assertion:

1. **"The report is on `RenderTemplateReportingMissing`, but every call site calls plain
   `RenderTemplate` — if these are two distinct functions the report never fires."** They are not:
   `component_library.go:952` is a three-line wrapper —
   `func RenderTemplate(...) string { out, _, _ := RenderTemplateReportingMissing(...); return out }`.
   Every call site reaches the report. ✅
2. **"The rationale says the counter does not advance for a CARRIED section, but the sketch shows an
   unconditional `instances.Next(...)` in the loop."** The carried branch `continue`s at
   `rerender_page_sections_action.go:399`, before the render block at `:511` where `Next` is called.
   The stated behaviour holds. ✅

### 10.2 `prior_art_librarian` — three new factual claims, unverified in the submission. Verified here.

| claim | check |
|---|---|
| `pattern-check.py` already has the `check_unrepaired_component_write` idiom with a raw-body/stripped-body split | `scripts/pattern-check.py:608` (the check), `:567` (`_calls_repair_seam`, the raw-body scan), `:576-605` (`COMPONENT_WRITE_ALLOWED`, reasons-not-exemptions). The new check mirrors all three. ✅ |
| `cmd/component-render-check/rendercheck.go` exists and is one of the call sites | it exists; 4 `RenderTemplate*` calls; renders every active component through the production entry point; writes its report to `doc_notes` (`:507`) and never to a served page; synthesises a marker for every referenced field (`:315-350`). ✅ |
| `section_editor_actions.go`'s two sites render page-embedded components and bound no token | `applyContentEdit` (`RenderTemplate(htmlTemplate, …)`) and `applyComponentSwap` (`RenderTemplate(comp.HTMLTemplate, …)`); neither bound anything before `32d6e980a`. ✅ |

### 10.3 `guardian` — the census was Go-file-level, not pipeline-level. Here is the pipeline level.

Live, non-snapshot agent types touching the edited actions — and the measurement itself needed a
correction, which is the more useful half:

| action | agent types that EXECUTE it | that MENTION it |
|---|---|---|
| `render_site_components` | 7 — nav-link-fixer, nav-updater, pageflow-builder, rerender-chrome, rerender-pages, rerender-site, site-work-orchestrator | 7 |
| `assemble_from_library` | 4 — content-site-architect, landing-page-architect, portfolio-architect, site-component-architect | 4 |
| `rerender_page_sections` | 1 — page-rerender | 3 (council-gate + fix-proposer are footprint mentions, not steps) |
| `apply_section_edit` | 1 — section-editor | 2 (tool-improver) |
| `render_component` | **1 — page-content-writer** | 1 |

**≈15 distinct live agent types across 4+ pipelines** — wider than the Go-file census implied, and
the guardian was right to ask. All of it inert today (0 of 243 templates), which is what makes it
safe rather than lucky.

> ⚠ **`render_component` first measured as ZERO, and that was wrong.** A census over
> `workflow.steps.*.action` cannot see an action nested in a `loop` step's
> `config.sub_workflow.steps.*.action` — fleet-wide that hides **80 invocations across 19 action
> names**. `render_component` is executed inside `page-content-writer`'s `process_sections_loop`.
> Caught by a text control disagreeing with the structured query; without it I would have published
> "that render path is dormant". Full trap in `LANDMINES.md`.

### 10.4 `bug_historian` — the objection to take seriously, and it is NOT closed

> *"This ships a second instance of the `missingkey=zero` exposure rather than closing the
> mechanism … the realistic outcome is that `InstanceID` joins the population of 'detected, never
> enforced' fields already documented elsewhere on this platform."*

**Recorded as a known-unresolved root cause, which is exactly what the seat asked for.** Stating it
plainly:

- Go's `missingkey=zero` blanks **any** absent field in **every** template, silently. 283 adds a
  report for **one** field name. Every other required field on every other template is exactly as
  exposed as it was before, and **nothing in this bug attempts the general fix.**
- The general fix would be at the template-execution layer — refuse (or loudly mark) any bare
  output placeholder that renders empty. That is a fleet-wide behaviour change to every render, and
  it is not this bug's to make.
- `enforce_instance_scope` stays **off**, so today a collision is recorded and nothing acts on it.
  It cannot default on: 13 active pages already collide via `generic-text-block`, so arming it
  would fail their next re-render.
- **The sequence that actually closes it** — and the only honest one — is: convert the templates,
  which removes the 13 collisions' cause; re-measure; then arm. Arming a narrowed guard *now*
  would be a mechanism that fires on nothing, which is the failure mode the owner's 2026-07-29
  ruling warns about.

### 10.5 `reuse_agent` — "'filed as the follow-up' has no locator". Correct; now it does.

I wrote that the `ComponentID` unification was "filed as the follow-up the architecture seat asked
for". **Nothing had been filed** — there was a `verify-later` line in CLC-014 and a sentence in a
commit message. The seat's objection is upheld and the artifact now exists:

**`architecture_review/RFC_032_three_render_context_builders_disagree_about_what_an_instance_is.md`**

It also carries the `architecture` seat's own follow-up ask (§5 there): the RFC_022 exception 283
was approved under holds **only while zero live templates reference `{{.InstanceID}}`**.

> **UPDATE, same day: that trigger is now BUILT, DEPLOYED AND PROVEN.**
> `instance-token-adoption-check`, a daily CronJob at 07:40 UTC
> (`deployments/kustomize/services/instance-token-adoption-check/`). It counts active components
> referencing `{{.InstanceID}}`: **0 = the exception holds; non-zero = it has expired and `RFC_032`
> is owed a round.** One `doc_notes` row per run (`subject_key='instance-token-adoption'`) even
> when quiet, so a missing row means the job did not run.
>
> **Not the pattern-check the seat suggested, and the reason matters:** `pattern-check.py` is a
> commit-time lint over repo files, and an `html_template` is written by the component-creator
> agent, by hand-authored SQL, by migrations and by the admin UI — four routes, none through a
> commit. Only a clock against live data can see the first adopter.
>
> **Its healthy answer is ZERO, so it carries a DEMAND CONTROL** — the same `LIKE` in the same
> statement also counts `{{.ComponentID}}`, and the job **refuses (exit 2)** if that returns 0,
> rather than reporting a zero it has not earned. First live run 15:29 UTC: adopters 0, control 5,
> active 243; both report branches were exercised via `--stdin` before deploy, with exit codes
> checked, so the tripped branch is known reachable rather than hoped-for.
>
> ⚠ **A trip is not a defect** — conversion is the intended next phase. **Retire the job once it
> trips**, or a fired tripwire becomes a muted one.

### 10.6 The remaining advisories, and what was done

- **`debug_historian`, medium — compile atomicity** ("six files must land in one commit; round 1's
  own HIGH objection was a merge-half-done"). They did: `32d6e980a` is a single commit containing
  the seam and all five callers. ✅
- **`debug_historian`, low — no deploy-verification for the new logging seam.** Honest answer: the
  log line **cannot fire in production today**, because it fires only when a template references
  `{{.InstanceID}}` and none is bound, and no template references it. What *is* verified is that
  the code is in the running binary (§10's digest chain). The pod-level check becomes meaningful
  the moment the first template converts:
  `kubectl -n ai-persona-system logs -l app=agent-chassis --tail=2000 | grep 'no per-instance token was bound'`.
- **`guardian`, low — "confirm the 0-template count holds at merge time".** Re-run post-merge and
  post-roll, 2026-08-16: **0 of 243.** ✅
- **`guardian`/`bug_historian`, low — the pattern-check allow-list is keyed on file basename, so a
  render call added to an allow-listed file inherits the exemption silently.** Real, unfixed, and
  it matters most for `component_library.go`, which hosts both the chrome renderers and the shared
  seam. Keying on file+symbol is the improvement; not done.
- **`render_guardian`, low — the report logs and does not refuse.** Same substance as 10.4, same
  answer. Its warning is worth carrying verbatim: *"if `enforce_instance_scope` is ever flipped on
  before the 13 known-colliding pages are fixed, that would be a high-severity fail-loud
  violation."*

---

## 11. UPDATE 2026-08-17 — the conversion is SCOPED, and this file's own premise was stale

The owner gave the go-ahead for the conversion. Before writing any of it, the scope was measured —
and **"the 22 calculator templates" understates the work by a factor of four.** Full proposal and
the decision the owner now owes: **`architecture_review/RFC_034`**. Summary here so the bug file
does not send the next reader to a stale number.

### 11.1 The measured scope (live, 2026-08-17)

Every figure shares one denominator — **component ROWS that bind by `getElementById` and are placed
on a live page**, which is the unit that actually gets converted:

| measure | value |
|---|---|
| **component rows to convert** | **91** |
| distinct `function`s among them | 83 |
| functions with MORE THAN ONE active row (forks) | 4 — `tool-llm-cost-calculator` ×4, `tool-automation-savings-estimator` ×3, `tool-affordability-complaint-checker` ×3, `tool-model-approach-selector` ×2 |
| live pages affected | 94 |
| domains affected | 22 |
| literal `id=` attributes | 1,346 |
| `getElementById` calls | 886 |

**Blast radius per unit is small:** measured per ROW, **1** row is placed on more than one domain
(max 2) and **3** on more than one page (max 2). Converting one row changes one page.

> ⚠ **I first measured that sharing by `function` and got "4 components across up to 5 domains".**
> Wrong, and wrong in the flattering direction — the four widest-spread functions are exactly the
> four carrying forks, so grouping merged several single-domain rows into one apparently-shared
> function. **Convert by `content_components.id`, never by `function`:** a function-keyed
> conversion would also silently skip 9 forked rows. Caught by re-measuring at the row level
> before publishing.

### 11.2 ⚠ CONVERTING THE IDS ALONE MAKES THE PAGE READ CLEAN AND LEAVES IT BROKEN

This is the finding that decides how the work is sequenced, and it is **proven, not argued** —
`TestIDOnlyConversion_readsCleanOnIDsAndIsStillBroken` in `component_instance_scope_test.go`.

Namespace the ids on today's real shape, render two instances, and the detector reports:

- **0 duplicate element ids** — the mechanical half genuinely worked, and
- **2 surviving `window.onload` assignments** and **2 surviving global-scope scripts**.

Both instances still declare `function runCalc()` at top level; the second replaces the first; every
instance's `onclick="runCalc()"` resolves to that one surviving function. So every button runs the
**last** instance's logic — against correctly namespaced, correctly unique fields. The page now
passes an id check while producing exactly the wrong answer this bug was filed about.

Mutation-checked: making the fixture script-scoped fails the test, so it is sensitive to what it
claims. **Consequence: ids and scripts convert in ONE step per component. A phased "ids now,
scripts later" plan is worse than doing nothing**, because it removes the only signal that anything
is wrong.

### 11.3 ⚠ The IIFE route is FORCED — `{{.InstanceID}}` is not a valid JS identifier

It renders as `c-mortgages-repayment`. The obvious de-collision — `function runCalc_{{.InstanceID}}()`
— is a **syntax error**, because the token contains hyphens. Asserted by
`TestInstanceToken_isNotAValidJSIdentifier` so a converter author meets it in a test rather than on
a shipped page. Each script must therefore be IIFE-wrapped, **which forces the 22 inline `on*=`
handlers to be rewired** to `addEventListener` — the step that is not safely mechanical.

### 11.4 The two quiet surfaces that break with no error at all

Of the 91 rows: **58 carry `<label for="…">`** and **33 reference an id from CSS inside a `<style>`
block**. Neither throws when the id underneath it is renamed. A conversion that handles `id=` and
`getElementById` but forgets these ships pages whose labels no longer focus their input and whose
component styling silently stops applying — across 94 live pages. Also present: 34 rows use
`querySelector`, where the selector may be built by concatenation rather than written literally.

### 11.5 What is NOT decided, and is the owner's

`RFC_034` §4 puts three shapes with their costs: a deterministic converter as a new
`fix_component_template` fix_type (reuses live machinery, cannot do the script half); an LLM rewrite
(can, and is the `bugs_open/012` truncation class); or a hybrid gated on the detector. It also names
what must be settled **with** the shape rather than after it: rebaselining `b2_verify`'s
byte-identical check, moving `oracle.py`'s 170 selectors in lockstep, the ordering against
`RFC_032`, and when `enforce_instance_scope` is armed.

**Nothing has been converted. 283 stays OPEN and the defect is still live.**

---

## 12. UPDATE 2026-08-17 (afternoon) — owner ruled, detector corrected, converter BUILT; round 3 approved clean

### 12.1 The owner ruling (RFC_034, now DECIDED)

**Shape C (hybrid), LMC first, and the conversions RUN THROUGH THE FRAMEWORK** — work items and
dispatched actions, the detector as the gate, never hand-applied SQL or hand-edited templates.

### 12.2 The corpus numbers were corrected TWICE in one day — read §3a of RFC_034, not any earlier figure

Regex triage said 24 need script work; the detector said 88; the truth is **25**, because the
detector itself had a defect: its accepted-wrapper regex anchored at the body's first byte, and the
estate's tool templates conventionally open with a `/* tool-doc */` comment — **62 correctly
IIFE-wrapped scripts read as global** (a 70% false-flag rate, found by sampling one flag). Detector
fixed (`5b30a831b`), council round 3 **approved, all reviewers approve** (2026-08-17 12:43 UTC).
Final split: **66 mechanical-only / 25 judged**, and the 25 are the 23 `loans-*`/`mortgages-*`
calculators plus two tools — the original "22 templates" scope, rediscovered from the other side.

### 12.3 The deterministic half is BUILT (`b7b396cb3`) — fix_type `scope_component_instance`, register CLC-017

`ConvertTemplateToInstanceScope` + `GateConvertedTemplate` + the framework seam in
`FixComponentTemplateAction`. The gate enforces §2.1 mechanically: a component whose script still
declares into global scope gets `fixed:false, action:"needs_script_scoping"` and **nothing is
written** — ids-only conversion is unrepresentable through this path. Every ambiguity refuses
(hex-ambiguous ids, already-converted, unrecognised binding constructions). Fixtures are pinned
LIVE ROW BYTES, and one paid for itself immediately:

> ⚠ **`data-*` attributes can carry id references no call-site pass can see.**
> `tool-css-unit-converter`'s five copy buttons write `data-target="result-px"` and the script
> binds `getElementById(targetId)` off the attribute at runtime. A conversion that missed it would
> ship five silently-dangling buttons. The composed test never contained this; the real bytes did.
> There is now a `data-*` exact-match pass, and the fixture asserts it.

### 12.4 ⚠ THE "FRESH BUILD" OF 2026-08-17 14:43 DID NOT SHIP THIS — same-tag cache trap, live

The pods restarted at 14:43 UTC and still run **yesterday's digest** (`f90a7e88…`, revision
`6a782274b`): round 2's seam is live; **the detector fix and the converter are NOT.** A local
rebuild at 14:30 (`89a0cbeb7`) contains both and was pushed — but under the **same tag**
`v1.0.1305`, so the restart served the node's cached image. This is CLAUDE.md's "bump `IMAGE_TAG`
for every build" trap, observed live (and the same day another lane measured 203 commits unshipped
the same way). **The fix rides the next roll under a bumped tag; do not one-service-apply it**
(releases are whole-fleet, owner runs `make release`).

### 12.5 What execution still needs, in order

1. **A roll with a bumped tag** (12.4). Verify at the artefact: pod digest == local digest, then
   the revision label, then ancestry of `b7b396cb3`.
2. **The work-item seed** — `instance_scope_conversion` items, one per component ROW
   (`content_components.id`, never function), LMC's rows first. Not written.
3. **Pre-conversion owed steps**: rebaseline `b2_verify`'s byte-identical check; move `oracle.py`'s
   170 selectors in lockstep with each converted calculator.
4. **The judged pipeline for the 25** — design not started; every LMC calculator is in this pool,
   so LMC-first means the judged pipeline is needed early, with the truncation check
   (`output_tokens == max_tokens` means CUT) on every rewrite.
5. **Per component: convert → rerender → redeploy.** A `fixed:true` changes no served page until
   the rerender.

---

## 13. UPDATE 2026-08-18 — the canary is DONE END-TO-END, and its real product was a vocabulary fix the whole estate needed

### 13.1 The canary's first run exposed the gap

Item `38efde3b` converted the template **exactly to prediction** (§ 12.3's counts, to the digit),
the fixer raised its rerender, a **111-page site-wide fan-out** completed, the page deployed —
**and the served page still carried the OLD ids.** The mechanism, read at the config:
`page-rerender.check_rerender_mode` routes to the sections re-render (the path that renders from
`html_template`) only for `image_landed | section_data_resolved | cta_links_stale`; everything
else ASSEMBLES STORED HTML. *"The component's template changed" had no reason in the vocabulary at
all* — so every template fix on the estate shipped stale bytes under a green status. This is
CLC-002's own recorded history repeating ("carried no reason, making the triggered re-render
assemble-only").

### 13.2 Migrations 460+461 (APPLIED, live, council round 4 submitted)

**460**: `template_changed` joins the reason vocabulary (verbatim pre-image guard; additive), and
the fixer's `create_rerender` is replaced — one **page-scoped** `page_rerender` per page carrying
the fixed component, across every site it is placed on, instead of a site-wide reason-less
fan-out. **461**: corrects 460's embedded query (`pages` has no `filename` column — the error was
DATA to 460's probe run and became SQL only at step-execution time; nothing reads `spec.filename`,
so the key is dropped) and adds the missing check class: **PREPARE-compile the embedded query in
the verify block**. Landmine written; snapshots in `agent_definitions_backup`.

### 13.3 The canary, completed through the new path

Item `82265e18` (`reason=template_changed`) took the sections path: stored render flipped at
11:54 UTC, and the served page reads **34 new instance-scoped ids, 0 old, 0 unrendered tokens,
and all five copy-button `data-target`s PAIRED** with their renamed elements. The full chain —
convert → gate → snapshot → write → reason-carrying rerender → sections re-render → deploy — has
now run once, end to end, through the framework.

The 07:40 tripwire **tripped on schedule** (adopters 1) — the RFC_022 expiry doing its job;
acknowledged by RFC_034 DECIDED; retire the CronJob.

### 13.4 The batch is RELEASED

Seed applied 2026-08-18 ~12:1x UTC: **70 `instance_scope_conversion` items** queued (the eligible
count moved 66→69→70 across two days — the corpus drifts, which is why eligibility is derived at
apply time). Each conversion now auto-files its page-scoped `template_changed` rerenders through
the fixed `create_rerender`. Monitor armed on the drain. The judged pool (25: the 23 LMC
calculators + 2 tools) remains untouched — its pipeline is the next design task, LMC-first with
the oracle, per the ruling.

### 13.5 The batch is DRAINED — 68 converted, both refusals are the refusal arms working, tripwire retired

Final reconciliation, counts read fresh at terminal state and tied item-by-item:

| | count |
|---|---|
| batch items | 70 |
| `fixed:true` (converted, gated, written, snapshot taken) | **68** |
| `fixed:false` — hex-ambiguity refusal | 1 — `tool-process-automation-scorer` declares an id literally named **`ec2`**; the `#id` pass cannot tell it from the colour `#ec2`. Route: judged pool, or rename the id first |
| `failed` — gate hard-error | 1 — `tool-spawn-rate-balancer`, the **pre-existing internal duplicate** (`chartTitle` twice within one copy, §3a's off-by-one); prefixing cannot fix a within-instance duplicate, and the gate refused to write. Route: repair the duplicate, then reconvert |
| adopters (templates carrying `{{.InstanceID}}`) | **69** = canary + 68 ✓ exact |
| page-scoped `template_changed` rerenders auto-filed by the fixed `create_rerender` | **72** (68 components; some sit on 2 pages), draining under a monitor |

**Both non-conversions are the failure-direction design proving itself**: each refused loudly,
wrote nothing, and named its route in the message. Neither is a converter defect.

**The tripwire is RETIRED** (CronJob deleted 2026-08-18; retirement note in `doc_notes`;
CLC-016 updated). It lived three days, ran daily, caught real corpus drift on its second run,
tripped on the first conversion exactly as designed, and was removed before its firing could
become noise — the full life cycle the design asked for.

**Remaining in 283:** the 72 rerenders drain and a sample of served pages gets spot-checked;
the two refused rows get their small repairs; then the JUDGED pool — 25 components, 23 of them
the LMC calculators, LMC-first with the oracle, pipeline not yet designed. The mechanical
three-quarters of RFC_034 is done.

### 13.6 The rerender drain found the OWNED-page seam — 17 items, one more guard proving itself, migration 462

Of the 72 auto-filed rerenders, **17 target `rebuild_policy='owned'` pages** (the tool pipeline's
native pages, mostly webdesign.co.uk tools — including the three components that joined the corpus
mid-programme). `save_page_sections` **correctly refused** the generic save on each ("a generic
section save would clobber it"): 2 failed before the pattern was diagnosed, and the remaining 15 —
a deterministic, fully-diagnosed refusal — were **cancelled** rather than left to feed the fleet's
failure sweeps (doc_notes records it; the never-cancel-pre-diagnosis rule is about undiagnosed
rows, and these are the opposite).

**What this means for those 17 components:** their TEMPLATES are converted (snapshot-reversible);
their served pages keep pre-conversion ids until the **owning pipeline** next renders them, or a
targeted `apply_section_edit` (which binds `InstanceID`) is used. Their owning lane
(webdesign tool rebuilds) renders outside `RenderTemplate`, so no empty-token hazard arises from
this seam — but that pipeline's next render of a converted template is worth one deliberate look.

**Migration 462 (applied, PREPARE-checked):** the fixer's `create_rerender` now excludes owned
pages, so it stops filing items that exist only to be refused. The remaining ~55 generic-page
rerenders continue draining.
