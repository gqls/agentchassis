# 342 — An ABSENT required field still renders empty and silent; the gate that catches it runs at 2 of the 15 render call sites

**Filed 2026-08-20** by the `bugs_open/260` renderer-half lane, at the council gate's request
(trail `a44d9eb8`: the `bug_historian` seat's **gating** objection in round 1, and `architecture`'s
advisory in round 2 — *"worth a follow-up ticket with a target date rather than an open-ended
note"*). ~~**UNOWNED.**~~ **OWNED since 2026-08-22 by `bugfix_342_absent_required`**
(`docs/agent_docs/docs024_key_docs_latest/bugfix_342_absent_required/`).

**Status: OPEN — the SILENCE is FIXED AND LIVE on `agent-chassis` v1.0.1322 as at 2026-08-21 17:00Z**
(probed on both replicas: the report literal and the config key PRESENT, two independent
removed-string controls ABSENT, nonsense control absent). **Stays OPEN deliberately** — see the
residual below: nine of fifteen call sites report, six do not, and ~~no refusal was added
anywhere~~ **the refusal half is BUILT as of 2026-08-22** (banner below), pending its roll and
arming.

> ⚠ **The live-page routes ESCALATE as well as detect (2026-08-21, council round 5).** Two seats
> gated on the fact that this file called `applyContentEdit`/`applyComponentSwap` "the two with the
> most exposure" and then gave them a log line only. Both now file a `required_fields_missing`
> item, unconditionally — the write is per-EDIT, not per-build (271 such edits in four months), so
> it is not the fleet-wide new authority that makes the chrome sibling opt-in. ~~**That commit is
> NOT in v1.0.1322** and is inert until the next roll.~~
> **CORRECTED 2026-08-22: the roll happened — the escalation is LIVE on v1.0.1323.** Evidence at
> the artefact, not the tag: `cd90e8b27`/`65f1b0b95`/`af4743464` are all ancestors of the
> v1.0.1323 build stamp `70e7b4f9c` (`git merge-base --is-ancestor`), and the stamp was probed in
> the binary on BOTH replicas (`grep -aq` on `/proc/1/exe`, nonsense control absent).

> **2026-08-22 — WHAT THE OWNING LANE DID (council submission alongside the commit).**
> 1. **The REFUSAL half, at the section-editor persist switch.** The two editor routes filed the
>    item and then persisted the blank anyway. Now `ApplySectionEditAction` refuses to persist
>    when the seam published absent required fields — ONE gate at the ONE persist switch (the
>    file's own idiom: link repair, envelope refusal), so a future edit branch inherits it. The
>    item is filed BEFORE the refusal, so a refused edit still leaves its queue entry. **Opt-in,
>    default OFF** (`refuse_absent_required_fields`, owner ruling 2026-08-02 §2 — such an edit
>    SUCCEEDS today, so refusing is new authority; RFC_022's three conditions hold: opt-in ✓,
>    unsafe default OFF ✓, zero live consumers name it ✓ by `agent_definitions` scan). Armed by
>    migration `551_…_HOLD.sql` ONLY after a binary carrying the code rolls. ⚠ Interaction with
>    `bugs_open/344` stated in both the code and the migration: `apply_edit` has no `error_step`,
>    so the DRIVING item of a refused edit may read `complete` until 344 lands — the live page is
>    protected either way, and the filed item survives.
> 2. **The chrome record ARMED** (migration `550`, appliable on sight — the Go half is live).
>    Measured first per §5: the chrome store (`site_components`) references only components with
>    ZERO required llm fields, so the arm fires on 0 rows today — free now, and the door closes
>    before a chrome component that declares required fields (five exist in the library) is ever
>    adopted.
> 3. **§5's "expect the 75-of-253 no-schema components to be the hard part" has DISSOLVED** —
>    re-measured 2026-08-22: 100 of 283 active components have no schema, but **95 are
>    `component_level='tool'`**, self-contained by design (`isSelfContainedSection` codifies it).
>    The real class is **5 non-tool components, ONE page_components usage each**, 2 with template
>    placeholders (`report-request-form`, `audience-check-form`). Small data work for a content
>    lane, not chassis work; the seam covers each the moment a schema exists.
> 4. Six unwired sites re-verified individually — each has a mechanism reason (raw candidate
>    templates with no component row; a contact-info block whose callers hold no schema; a
>    stitched TEMPLATE whose content arrives later; audit probes that remove fields by design).
>    Still no change owed there.
>
> **Remaining to close this file:** apply 551 after the next roll + the live canary (one
> refusing edit with the live section byte-identical, one clean edit persisting), and a decision
> on the 5 no-schema components (or an explicit scope-out recorded here).

> **WHAT WAS DONE.** RFC_041 §5's candidate (a), the structural one: `RenderContext` now carries the
> component's `InputSchema`, and the **seam** applies `missingRequiredLLMFields` — the same function
> the two pre-render gates call, deliberately not a second implementation — logging absent required
> fields at **Error** for every caller, and **PUBLISHING them on the RenderContext** so a caller
> with a database handle can escalate.
>
> **ESCALATION, not just a log** (council `bb7f5d0e` round 1 was gated on exactly this: *"a named
> log is not escalation"*, `bugs_open/054`). The chrome path files a **`required_fields_missing`**
> item — the type that already exists for this defect, with a router already seeded
> (`bugs_open/277`) — opt-in, unsafe default OFF (`record_absent_required_fields`). ⚠ **It reaches
> a population the existing producer structurally cannot:** `check_required_fields_missing` scans
> `pc.build_status = 'deployed'`, i.e. rows that made it, while this defect's sections render
> empty, get dropped by assembly, and never become one. Survivorship — the same shape as
> `bugs_closed/260`'s "no live damage" headline, and the reason this queue has looked quiet.
>
> **COVERAGE: NINE of fifteen call sites pass a schema**, and the arithmetic closes because it was
> wrong once and the council caught it. Wired: `v3_site_actions` (build), `render_site_components`
> (chrome store), `rerender_page_sections`, `assemble_from_library`, `RenderHeader`,
> `RenderFooter`, `RenderHead`, and **both section-editor routes** (`applyContentEdit`,
> `applyComponentSwap`) — the two that write `rendered_html` straight to an already-live page, and
> the two an earlier round had left out. **NOT wired (6):** `GateConvertedTemplate` and
> `tool_birth_instance_scope` (raw candidate templates, no component), the legacy head render
> (loads a template only), `RenderTemplateWithMap` (a different executor, not this seam), and the
> offline audit's two calls (it probes with fields REMOVED on purpose, so a report there would
> fire on every probe by design). **9 + 6 = 15.**
>
> ⚠ **The count read "nine" once before while the enumeration listed seven** — from a
> `grep -c 'InputSchema = '` that also matched `ci.InputSchema`, a DIFFERENT struct with an
> identically named field. Count with a grep that names the receiver.
>
> **The schema is not a new field.** It reuses the slot of the dead `SchemaSnapshot`, and
> `SchemaMode` + `RenderOptions` were deleted with it — all three had zero readers since
> `bugs_closed/260` deleted `RenderTemplateWithValidation`. The control-field property that guarded
> `schema_mode` moved to `InputSchema`, where it matters more: content that could set it would hand
> the renderer its own contract and switch off its own check.
>
> ⚠ **NO REFUSAL was added, deliberately.** Refusing at the seam would be new authority over
> content that renders successfully today at sites that never asked for it (owner ruling
> 2026-08-02 §2), and the two paths that want to refuse already do, before the render. What ships
> is a report PLUS a queue entry. **Refusal per path is the remaining work and is not scoped
> here** — this file stays OPEN for that and for the six unwired sites.
>
> ⚠ **Nothing is watching the log line.** Checked rather than assumed: no PrometheusRule exists in
> the namespace and no kustomize manifest consumes log level, so the Error lines feed no alerting
> surface. That is the honest limit, and it is why the work item matters more than the log.
>
> **A finding the tests forced out, worth knowing before you touch this:** the seam and the
> pre-render gate give DIFFERENT answers on the same content, and both are right. The seam judges
> the merged map the template actually sees, where `contextToInterfaceMap` supplies fleet defaults
> (`cta_text` → "Get Started"), so a required field the writer never produced can still render
> something. The gate judges the writer's output. "Did the WRITER supply it?" and "will it RENDER
> EMPTY?" are different questions, and this bug is the second one — so the seam's report is a strict
> SUBSET of the gate's. Pinned by `TestSeamReportsASubsetOfThePreRenderGateAndSaysWhy`; do not
> "fix" one to match the other.
>
> Commit + council trail `bb7f5d0e-d125-42c9-a155-9bac866a5017`. Tests:
> `render_seam_absent_required_test.go` — six, three of them controls, mutation-proven.

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
