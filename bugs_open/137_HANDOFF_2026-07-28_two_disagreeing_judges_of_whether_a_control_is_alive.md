# 137 — Two mechanisms judge "is this control alive", in the same function, and they disagree

**Filed** 2026-07-28 by the experience-register session, at the explicit request of the council
gate's `reuse_agent` seat (corr `99f2a5e6-e934-4ca1-addb-f16a29b38b0f`), which objected that
deferring this to "the register" would leave the two unreconciled:

> *"leaving two disagreeing judges of control-liveness unreconciled is the reuse risk, and it
> should at minimum be named to architecture review rather than deferred silently to 'the entry.'
> … the reconciliation of shell-dead-controls vs. the new attribute check belongs on someone's
> roadmap now, not only 'in the register' later."*

**Status:** ~~OPEN, unowned~~ → **FIXED IN CODE 2026-07-31 (`9f0a8ec5b`), OPEN only until the next
chassis roll.** Option 2 taken. Council correlation `4465f655-c6c6-49b4-a9b8-4ca7a5f647df`.
Working docs: `docs024_key_docs_latest/bugfix_137_control_liveness/`. See the addendum at the
bottom — including the part of the cause this file did not have.

**Not a regression** — both mechanisms behave as designed. This is a
contradiction that only became *visible* today, and it was invisible for a structural reason worth
keeping: you cannot detect a disagreement between two rules when only one of them can speak.

## The two judges

Both live inside `evaluateStaticCriteria`
(`platform/orchestration/actions/discovery_checks/check_tool_acceptance.go`).

**1. `shell-dead-controls`** — a built-in sweep, always on, independent of any criteria document.
It fails a page carrying a no-op href (`#`, `#!`, `javascript:void(0)` — `datahelpers.IsNoopHref`),
**unless the page contains `data-runtime-fill`**, in which case the whole sweep is skipped:

```go
// Runtime-fill shells are exempt (placeholder hrefs hydrate client-side)
if !strings.Contains(html, "data-runtime-fill") {
    if dead := datahelpers.DeadControlAnchors(html); len(dead) > 0 { ... fail ... }
}
```

**2. `attribute_absent`** (new, 2026-07-28, TL-031) — opt-in via a criteria document. It fails a
page when a matched element carries a forbidden attribute. **It has no runtime-fill exemption**, by
deliberate omission: the check asserts exactly what it says, and building in a page-wide escape
hatch would silently weaken every attribute assertion the experience register depends on.

## The element they disagree about, live today

`https://vonc.com/provocations/index.html` serves, verified 2026-07-28:

```html
<section class="provocations-archive-list-section provocations-archive"
         data-component="provocations-archive-list" data-runtime-fill="true">
  ...
  <a class="provocations-archive__item" data-archive-template hidden href="#">
```

- `shell-dead-controls`: **exempt** — the section declares `data-runtime-fill`, so the sweep never
  runs, and this is the *only* `href="#"` on the page.
- `attribute_absent` on `{{binding.item_template}}` for `href`: **FAIL** — "1 of 1 element(s)
  matching `.provocations-archive__item` carry a forbidden attribute (href): element 1 carries
  href=\"#\"".

Same element, same evaluator, same run. One says fine, one says broken.

## Which is right is genuinely open — that is why this is filed rather than fixed

Evidence for "it is a defect": the component's **own loader** calls it one, in its own comment
(`/assets/js/snippets.js`, the provocations-archive-loader DOM contract):

> *"An entry WITHOUT one is NOT offered as openable: class --static, the template's placeholder
> href is REMOVED (it is href=\"#\", a dead control), no tabindex, no handler."*

And the register entry `feed-driven-teaser-list` states the clause and its reason: *"the un-cloned
template is itself markup on the page; left with href='#' it is a dead control that no visitor can
see but every sweep can find."*

Evidence for "it is fine": the element is `hidden`, no visitor can reach it, the loader removes the
href before any clone is appended, and the platform made a considered decision that runtime-fill
shells' placeholders are legitimate pre-hydration.

`[MY READING, UNRESOLVED]` The clause is **mis-tiered** rather than wrong: "the loader removed its
href" is a claim about the post-hydration DOM, which is Tier 4, and asserting it statically judges
the page a moment before the sentence becomes true. But that reading conveniently makes my own
check's only red result go away, which is exactly the kind of reasoning this bug exists to have
checked by someone else.

## The structural question, which outlives the one element

Two mechanisms answering "is this control alive" with different exemption rules will drift, and the
drift will be silent because each is individually correct. Options, roughly ordered by what makes
the bad state unrepresentable rather than merely unlikely:

1. **One shared predicate.** `IsNoopHref` + the runtime-fill rule become a single exported
   judgement that both the sweep and attribute assertion consult, so "alive" has one definition.
   Cost: attribute assertion inherits a page-wide exemption, which weakens it for every entry —
   the exemption is currently keyed on the whole page containing the string anywhere, not on the
   element being inside the shell.
2. **Scope the exemption to the element, not the page.** The current test is
   `strings.Contains(html, "data-runtime-fill")` — a page-wide string match. An element-scoped
   version (is this element inside a `[data-runtime-fill]` subtree?) is now cheap, because
   attribute assertion brought a real DOM into this file. Both mechanisms could then share it
   honestly. This looks like the best answer and is the one I would start from.
3. **Declare them different questions and say so in both.** The sweep asks "would a visitor meet a
   dead control"; the attribute check asks "does the markup conform to the contract". Cheapest,
   and it leaves two answers standing.

## How to verify whatever is decided

```bash
curl -s https://vonc.com/provocations/index.html | grep -o '<a[^>]*href="#"[^>]*>'
# expect exactly one hit today: the data-archive-template row
```
Then run the entry's criteria through `discovery_checks.EvaluateStaticCriteriaJSON` and check that
the two judgements agree. There is a worked harness in
`docs/agent_docs/docs024_key_docs_latest/experience_register/NOTES_experience_register.md`
(2026-07-28e).

## Related

- TL-031 in the concept register (`docs026_concept_register/register/tool-lifecycle.md`) — the new
  check type, with this disagreement recorded as its second landmine.
- TL-013 — the Tier 2 checker and the anchor rule, now carrying a dated note that the
  "confirm, never refute" rule has one deliberate exception.
- `bugs_closed/023` — CTA label/URL pairing unchecked; the same family (a control's promise not
  being tied to anything checkable) and the origin of the register's `no-inert-control` invariant.

---

## 2026-07-31 — FIXED via option 2. The disagreement was a SYMPTOM of the exemption's scope.

Fix in `9f0a8ec5b`, inert until the roll. Working docs:
`docs024_key_docs_latest/bugfix_137_control_liveness/` (the standing five).
Council submitted before the commit, correlation `4465f655-c6c6-49b4-a9b8-4ca7a5f647df`.

**Premise re-checked before touching anything**, and it had not gone stale —
`curl https://vonc.com/provocations/index.html | grep 'href="#"'` still returns
exactly one hit, the `data-archive-template` row.

### What this file did not have, and it changes where the fix belongs

This file locates the two judges inside `evaluateStaticCriteria`. Reading the
code for the fix found the exemption they share is inlined **eight times** as
`strings.Contains(html, "data-runtime-fill")` — and that its blast radius is set
by **how finely each caller chunks its input**, not by the markup:

| caller passes | the line actually asks | verdict |
|---|---|---|
| one section (`pc.rendered_html` in a row loop) | *is this section a shell?* | right, by accident of framing |
| an assembled page (`repairOutboundPageLinks`, `validate_page_content`) | *does ANY section here hydrate?* | **one shell exempts every neighbour** |
| a fetched served page + chrome (this file's own sweep) | wider still | **worse** |

So the two judges disagreeing is the second-order symptom. **`save_sections_link_repair.go:67-71`
had already worked this out** and fixed it at its own call site — *"a runtime-filled section still
exempts itself while its statically-linked neighbours no longer ride on its exemption"* — which is
the right answer applied by convention, leaving every other caller to rediscover it.

### Measured on live artefacts, before submitting — not left for a reviewer

| page (assembled from `page_components`) | bytes | old exemption | new exemption | newly visible |
|---|---|---|---|---|
| `vonc.com/index` | 48,956 | **100%** | 2 spans, 6,172 B (**12.6%**) | 2 dead controls ("Get Started", "Learn More") + 2 link repairs (48,956→48,866 B) |
| `vonc.com/provocations-index` | 7,684 | **100%** | 1 span, 1,400 B (18.2%) | **0 and 0, byte-identical** — the element this bug is about |

The two newly-repairable links are `gauntlet-cta`'s **"Enter the Gauntlet"** and **"Find Your
Archetype"** — the very controls `check_dead_controls.go` names in its own header as the case that
check was built for. They had moved from `href="#"` to `href=""`, which shifted them out of the
sweep's class and into the repair path's, where the page-wide skip was hiding them.

**Fleet-wide blast radius, stated plainly:** exactly **one** page across all deployed
`page_components` has a non-shell component holding empty hrefs alongside a page-mate shell. And
**zero** served *tool* pages currently carry either a shell or a no-op href, so this file's own
sweep masks nothing **today** — it is the mechanism, and it was unguarded. The case for the fix is
structural, not volumetric, and it is not dressed up as otherwise.

### Which option, and what was deliberately NOT done

**Option 2**, as this file recommended. Option 1 is refused by this file's own cost analysis
(attribute assertion would inherit a page-wide exemption); element scope is what removes that
objection and makes sharing honest. Option 3 is what the `reuse_agent` seat objected to.

`datahelpers/runtime_fill.go` holds one marker constant with two representations: `RuntimeFillSpans`
(byte ranges, so `RepairPageLinks` keeps its byte-identical-when-unchanged guarantee — a goquery
round-trip would break that on every page) and `InRuntimeFillShell` (the DOM form).

**The boundary is the design.** *"Is this CONTROL alive?"* becomes element-scoped. *"Is this SECTION
a shell?"* stays whole-input and untouched (`check_empty_sections`, `check_component_standards`,
`check_component_template_corrupted`, `sectionHasVisibleContent`) — `HasRuntimeFillMarker` **names**
that predicate so its use is a choice, and a test pins it against a later tidy-up.

### Answering this file's open question — and it is a SKIP, not a pass

`[MY READING, UNRESOLVED]` above suspected the clause was **mis-tiered**: "the loader removed its
href" is a claim about the post-hydration DOM. **That reading is adopted, and the file was right to
distrust it on its own** — so it is implemented in the form that cannot launder a defect.
`static_attribute_checks.go`'s **rule 2** already confines refutation to elements actually in the
served HTML and cites this sweep as its precedent without applying it; an element inside a hydrating
subtree is that same claim one step earlier.

A shell-enclosed element therefore returns **SKIP** — never PASS. A skip can never satisfy
`experienceVerdict`, so nothing vouches for markup that was not checked (rule 1's vacuity problem),
and the detail discloses how many elements were set aside. The exemption is **per element**, so an
anchor outside the shell in the same document is still refuted exactly as before. That is what keeps
this a reconciliation rather than a blanket amnesty on any page containing a shell.

### Verification owed at the next roll

1. Pod-grep with a positive control **in the same exec** (`bugs_open/153`):
   `strings /app/agent-chassis | grep -c RuntimeFillSpans` (0 before the roll) and
   `... | grep -c DeadControlAnchors` (non-zero either way).
2. **Induce the exempting branch**, not just the finding branch: a component whose root carries the
   marker must still produce zero dead-control findings, or the exemption has been deleted rather
   than narrowed. `TestShellDeadControlsExemptionIsPerAnchor`'s second case asserts this at unit
   level; confirm it on a real check run.
3. Re-run the `vonc.com/index` measurement above and record both numbers — the 2 dead controls and
   the 2 repairs should now appear as real findings/repairs on a live pass.

### One pinned expectation was INVERTED, and that is where the decision lives

`TestEvaluateStaticCriteria_AttributeChecksFlowThrough` asserted **FAILED** for the shell-enclosed
template row, its comment noting the sweep was suppressed on the same element — **this file's
contradiction, encoded as an expectation.** It now asserts **SKIPPED**, with the date and reason
written into the test. If the reconciliation is judged the wrong way round, change it there.

Tests proven load-bearing by mutation in both directions: restoring the whole-document span fails 8
tests; deleting the exemption fails 10, including the pre-existing
`TestRepairPageLinks_RuntimeFillShellIsExempt`. A suite asserting only the first direction would
pass against a change that simply removed the exemption.

### Registered
- `016b` §9 — *"An exemption tested with `strings.Contains` over whatever the caller passed has no
  fixed blast radius"*.
- Concept register **LNK-025** (`link-management.md`).
- `LANDMINES.md` — the marker is tested against whatever you pass; and note the **SQL-side** copies
  (`LIKE '%data-runtime-fill%'`) that a `--include=*.go` grep reports as absent.
