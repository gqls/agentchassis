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
