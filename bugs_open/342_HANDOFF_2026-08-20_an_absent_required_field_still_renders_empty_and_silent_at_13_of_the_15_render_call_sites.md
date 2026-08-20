# 342 — An ABSENT required field still renders empty and silent; the gate that catches it runs at 2 of the 15 render call sites

**Filed 2026-08-20** by the `bugs_open/260` renderer-half lane, at the council gate's request
(trail `a44d9eb8`: the `bug_historian` seat's **gating** objection in round 1, and `architecture`'s
advisory in round 2 — *"worth a follow-up ticket with a target date rather than an open-ended
note"*). **UNOWNED.**

**Status: OPEN, mechanism verified first-hand, no owner.**

> **WHY THIS IS FILED RATHER THAN FOLDED INTO 260.** `bugs_closed/260` fixed the *mistyped* shape:
> a field of the wrong TYPE used to degrade a whole section silently, and now fails loudly. This is
> its **sibling**, deliberately scoped out of that change and registered in three places rather
> than fixed quietly: an *absent* field is a different mechanism (Go's `missingkey=zero`, not the
> deleted fallback), it is covered by a different gate, and closing it changes behaviour on content
> that renders successfully today. Folding it in would have made a reviewed change unreviewable.

## 1. The defect

Go's `text/template` is configured with `Option("missingkey=zero")` (`call_agent.go`,
`executeGoTemplate`). A field the template references but the content does not supply therefore
renders as **empty, with NO error**. The section is then structurally present and visually empty —
and page assembly drops visually-empty sections, so the content silently vanishes.

This is not a theory: it is the mechanism behind this estate's recorded fleet-wide blanking of
article bodies (`bugs_closed/004`/`005`), and 016b §9 records it as a live pattern.

## 2. What covers it today, measured rather than assumed `[MEASURED 2026-08-20]`

`missingRequiredLLMFields` (`json_envelope.go:451-474`) is the presence gate. It is called from
**exactly two** places:

| call site | path |
|---|---|
| `v3_site_actions.go:2386` (`RenderComponentAction`) | page BUILD |
| `rerender_page_sections_action.go:396` (pre-check) | page/section RERENDER |

There are **fifteen** render call sites (enumerated in `STY-057`). So **13 of 15 are uncovered**,
including both section-editor routes that write `rendered_html` straight to an already-live page,
all three chrome renderers, library assembly, and the offline audit.

And the two that ARE covered are covered narrowly: the gate fires only for fields the schema marks
**both** `source: "llm"` AND `required: true`. An optional field, or a field on one of the
**75 of 253 active components that declare no schema at all**, is invisible to it everywhere.

## 3. Why the 260 fix does not close it

260 deleted a fallback that fired on template *execution errors*. An absent field is not an
execution error — `missingkey=zero` makes it a successful render of an empty string. The new error
channel never fires, the new type checker deliberately never fires (`IsEmptyContentValue` — absent,
nil and empty are the presence gate's question at every declared type, and two gates disagreeing
about one field is its own defect), and the opt-in `refuse_mistyped_llm_fields` gate is keyed on
declared TYPE, not presence.

## 4. Fix candidates, costed — from `RFC_041` §5, not re-derived

1. **At the seam (the structural fix).** Have the render itself report which declared-required
   `source:"llm"` fields were absent. The machinery already exists: `missingBareFields`
   (`component_library.go`) walks the template and tests the data map, and its result is already
   returned by `RenderTemplateReportingMissing`. This makes the coverage question **disappear**
   rather than move — every call site inherits it — and it is the only candidate that makes the bad
   state unrepresentable. Cost: it changes behaviour at 15 call sites at once, so it is
   architecture-scope and needs the opt-in-default-OFF shape (owner ruling 2026-08-02 §2).
2. **Call the existing gate at the other thirteen sites.** Thirteen edits, no new mechanism, and it
   leaves optional fields and schema-less components exactly as they are. Cheap, partial, and the
   partiality would be easy to mistake for coverage — which is the failure mode this estate keeps
   re-closing.
3. **Do nothing.** Stated so the choice is a choice: the cost is a live, recorded, fleet-wide
   silent-blanking mechanism with 13 of 15 doors open.

## 5. What a fixing lane owes

- **Do not "fix" it by making the type checker report absence.** That is candidate (2) wearing a
  disguise and it re-introduces the two-gates-disagree defect the 260 lane closed by sharing
  `datahelpers.IsEmptyContentValue`.
- **Measure the population before arming anything**, top-level AND nested, and **read the rows, not
  the count** — the 260 lane's own checker had a false positive that only an estate-wide census
  found (a live page storing five empty-string array fields, gated by `{{if}}`, serving cleanly).
  The query shape is in `sql_for_agents/502_bugfix_260_arm_mistyped_llm_fields.sql`'s header.
- **Expect the 75-of-253 no-schema components to be the hard part**, not the seam.

## 6. Evidence

- `json_envelope.go:451-474` (`missingRequiredLLMFields`) and its two call sites, greppable:
  `grep -rn "missingRequiredLLMFields(" --include=*.go platform/ | grep -v _test`
- `call_agent.go` (`executeGoTemplate`): `Option("missingkey=zero")`
- `STY-057` (concept register) — the 15-call-site table and this gap in its landmine list
- `RFC_041` §5 — the three candidates above, and the record that this is unowned
- `bugs_closed/004`/`005` — the recorded fleet-wide blanking this mechanism caused
- `bugs_closed/260` — the sibling (mistyped shape), fixed and live on v1.0.1319
- 016b §9, the 260 entry — corrected 2026-08-20, and its banner names this gap

**No `090` diagnosis run.** Stated plainly per the owner ruling of 2026-07-31 rather than omitted:
the substitute is first-hand verification — the two call sites were enumerated by grep, the gate's
predicate read in full, the `missingkey=zero` option read at its source, and the 15-site inventory
is the one this lane built and the council reviewed twice. What is NOT established first-hand is
how often the absent case actually fires today, because its whole nature is that it logs nothing —
which is itself the argument for candidate (1).
