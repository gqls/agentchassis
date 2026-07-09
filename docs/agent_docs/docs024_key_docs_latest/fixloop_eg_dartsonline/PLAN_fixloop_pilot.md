# PLAN — F0.1 plumbing, then the dartsonline pilot as a known-answer benchmark

Written 2026-07-09. Supersedes the "order of work" paragraph in the intake.
Companion documents: RUNBOOK_diagnosis_fix_loop(10).md (task statement, what
exists, phases, boundaries), NOTES_running_fixloop(10).md (evidence trail and
the reasoning that produced this plan).

## 0. What changed, and why the plan changed with it

The intake ordered: land F0.1 plumbing, then run the guides bug through it as
F0's acceptance test, with success = "a diagnosis carrying a static-tier
citation naming the code that drops or fails to route guide pages".

The mandatory pre-check ran first, as doctrine requires. **It produced that
citation by hand, in about a dozen queries and four file reads** (evidence in
NOTES(10), turn 1). The standing hypothesis was refuted; the real mechanism is
a success-labelled error terminal plus a nav query filtering on the wrong
column. The pilot can therefore no longer discover anything.

The plan below keeps the original order — plumbing first — and changes only
what the pilot *is*: from a discovery run whose answer nobody knows into a
**graded benchmark whose answer we hold**. This is a strictly stronger F0
acceptance test. Four of the five pre-registered pilot criteria were always
about the plumbing; only the fourth was about the answer, and that is the one
we can now mark objectively instead of taking on trust.

## 1. Slice F0.1 — the plumbing (do this first; unchanged in substance)

Three thin slices, each independently landable, each with criteria fixed
before the code is written.

### F0.1a — `diagnosis_artifacts` table
Columns: `id`, `correlation_id`, `iteration` (int), `kind` (text, CHECK IN
('bundle','iteration_note')), `body` (text), `created_at`, plus a retention
knob per kind. Index on `(correlation_id, iteration, kind)`.

*Criteria*: migration applies clean and is idempotent; a hand-inserted row of
each kind round-trips; the CHECK rejects a third kind.

### F0.1b — write-through inside the assemble action
`platform/orchestration/actions/diagnose_assemble_bundle_action.go` —
`DiagnoseAssembleBundleAction` (line 107) already composes the bundle in
memory and returns at line 219. Persist there, in Go, before returning. Zero
workflow-shape change; zero contact with the tools chat's live
`emit → persist_note → complete` surface. That isolation was the whole point
of choosing this placement (Q-A, decided 2026-07-07).

*Criteria*: a loop run leaves one `kind='bundle'` row per iteration; the
persisted body is byte-identical to what the action returned; a write failure
degrades to a logged warning and **never** fails the diagnosis (the loop's
read-only contract outranks our observability).

### F0.1c — the `needs_diagnosis` envelope
A work item with `pipeline='diagnose'` (the column exists and defaults to
`'build'`), `site_id` nullable, envelope extending
`drafts/084_TRIGGER_diagnose_v1.sh`, carrying `subject_type`/`subject_key` so
the tools chat's persist_note subject gate opens once their 3b lands.

*Criteria*: an item inserted with a null `site_id` dispatches; the anchorless
run survives (their `load_runtime` error-routing is already applied); the
manual trigger still works.

**Gotcha to honour while touching any workflow**: `error_step` belongs *inside*
a step's `config`. Step-level `error_step` is silently ignored (001 §16).
Note that `page-build-handler`'s live definition carries `error_step` at
*both* levels on `deploy_page`, `plan_sections`, `save_sections`,
`validate_content` and `load_spec_sections` — the inner one is load-bearing,
the outer is dead. Correct adjacent instances as a noted change if we touch it.

## 2. Regenerate the bundle before the benchmark run

`z_bundles/BUNDLE_fixloop_F0.md` is code+docs only — its Schema, Database
capabilities and Runtime evidence sections are all placeholders. Rebuild with
`-psql` as **one quoted argument, no `-it`/`-t`** (a TTY corrupts captured
output; `cmd/bundle`'s skip message and usage text were patched to say so).
Add `-schema-tables pages,page_components,site_work_items,site_specs,sites,
agent_definitions,site_plans,site_plan_pages` — note `site_plans` and
`site_plan_pages`, which the original invocation omitted and which the guides
evidence turned out to live in.

## 3. The benchmark — pre-registered scoring rubric

The loop receives **only the original symptom string**, with no hint of the
answer:

> "dartsonline.com published a Guides nav link and a /guides/index.html page,
> but the page is blank and no guide pages exist — while gamesdesign.co.uk, on
> the same platform, has working guides (and games and tools), and
> gaswholesalers.com has a working news feed."

Score its emitted diagnosis against these, fixed now, before the run:

| # | Claim the loop must reach | Tier | Weight |
|---|---|---|---|
| 1 | `pages.sections` empty ⟺ page not built (the exact 5/10 partition) | live data | must |
| 2 | `check_has_ready_sections` routes sectionless pages to `complete_error` | static | must |
| 3 | `complete_error` is a `complete_workflow` — a **success** terminal — so the item is stamped `complete` | static | must |
| 4 | nav selects on `pages.status`, not `build_status` (`populate_nav_tables_action.go:242-243`) | static | must |
| 5 | the `page-build-handler` work-item `result` lacks `deploy_result` for the 10 | runtime | should |
| 6 | gamesdesign's guide pages have sections and use the same handler ⇒ handler is not the discriminator | live data | should |
| 7 | `mark_no_sections` is referenced in a comment and does not exist | static | bonus |
| 8 | the two intake paths (`WriteBuildItemsAction` vs `reconcile_site_plan`) disagree on `unavailableBuilders` | static | bonus |

**Pass** = all four *must* claims, each with a citation that resolves to the
named file/step. **Refutation credit**: the loop should *not* assert the
standing hypothesis (that `reconcile_site_plan`'s routing table drops guide
pages). If it asserts it, that is a scored failure regardless of the rest —
the hypothesis is false and a citing loop that confirms a false hypothesis is
worse than one that abstains. **Abstention** on any *should*/*bonus* row is
neutral, not penalised: cite-or-abstain is the contract.

Also measured, independent of the answer: (a) intake landed via the documented
route; (b) every iteration's bundle is fetchable from `diagnosis_artifacts`;
(c) per-iteration notes were written; (d) iteration count and wall-clock, as
the first entry in the loop's own performance record.

**Blinding.** The benchmark is only worth running if the loop cannot read the
answer. NOTES(10) and this plan live in the repo. Confirm the loop's corpus and
`-doc` selection exclude `docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/`
before the run, or the result is worthless. This is the single most important
setup step and the easiest to forget.

## 4. F1 stretch — the constrained edit plan

If the benchmark passes, F1 emits an edit plan on a branch. Its target is the
**platform**, not the site (NOTES(10), DECISIONS 2026-07-09):

1. `page-build-handler` workflow: give `check_has_ready_sections` an
   `else_step` that flags rather than succeeds — build the `mark_no_sections`
   step the guard comment at `load_work_item_actions.go:756` already assumes
   exists, setting `needs_human_review`. The completion guard at `:759-766`
   then preserves it, and the whole thing works as designed.
2. `populate_nav_tables_action.go:243`: ground nav in the built set —
   `AND build_status = 'deployed'`, or an explicit join to the built pages.
   This is the "nav never links unbuilt pages" principle the roadmap work
   already identified; it wants to be a guideline amendment too (side-task,
   per Q-D).
3. Leave cause A (the planner under-populating `sections`) to the builder
   thread. It overlaps item 6/7 territory and a fix there is a bigger change
   than F1 should attempt on its first outing.

Validation before any PR: `gofmt` + `go build` in a spawned job (Q-C).
Verification after: rebuild dartsonline and assert nav contains no link whose
page is not `deployed` — a natural first job for the tools chat's Stage-6
browser-runner adapter.

## 5. Order of work

1. F0.1a migration → 2. F0.1b write-through → 3. F0.1c envelope →
4. regenerate bundle with `-psql` → 5. confirm blinding → 6. benchmark run →
7. score against §3 → 8. F1 edit plan if passed.

Steps 1–3 are unblocked and independent of everything above; they can start
now. Step 5 gates step 6 absolutely.

## 6. What this pilot has already taught the workstream

Three pilot candidates, three dissolutions by cheap pre-check. The pattern is
now strong enough to name: **on this platform, bug mechanisms tend to be
legible to schema access plus grep.** The loop's value is therefore unlikely to
be *discovery* on bugs of this shape. It is more plausibly: (a) doing this
unattended, at 3am, on a bug nobody has looked at; (b) doing it with citations
a human can audit; (c) doing it consistently across a class of bugs. The
benchmark measures exactly (a) and (b). Worth stating plainly rather than
letting the workstream's premise drift unexamined.
