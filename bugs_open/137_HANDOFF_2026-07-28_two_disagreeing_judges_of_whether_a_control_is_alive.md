# 137 — Two mechanisms judge "is this control alive", in the same function, and they disagree

**Filed** 2026-07-28 by the experience-register session, at the explicit request of the council
gate's `reuse_agent` seat (corr `99f2a5e6-e934-4ca1-addb-f16a29b38b0f`), which objected that
deferring this to "the register" would leave the two unreconciled:

> *"leaving two disagreeing judges of control-liveness unreconciled is the reuse risk, and it
> should at minimum be named to architecture review rather than deferred silently to 'the entry.'
> … the reconciliation of shell-dead-controls vs. the new attribute check belongs on someone's
> roadmap now, not only 'in the register' later."*

**Status:** OPEN, unowned. **Not a regression** — both mechanisms behave as designed. This is a
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
