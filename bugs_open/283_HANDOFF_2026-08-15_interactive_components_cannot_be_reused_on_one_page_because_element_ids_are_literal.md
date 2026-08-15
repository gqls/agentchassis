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
