# PLAN 2026-09-04 — the provenance rail (bugs_open/427, remaining work)

Written by the session named `427`, resuming the lane after it went dormant 2026-09-03
21:19 BST. Supersedes nothing: `PLAN_2026-09-02_bugfix_427_event_render.md` covered the
populate/correct/render phases, which are **built and live**. This plan covers what is
left, which turned out not to be what §23 of the bug file expected.

> **The plan §23 proposed has been overtaken by evidence and this plan replaces it.**
> §23.2 proposed three checker layers. Two peer lanes then demonstrated that all three,
> as scoped, score a live fabricating tool CLEAN. This plan does not widen them. It
> changes what is being detected.

## 1. What is already true (verified 2026-09-04, do not re-derive)

| half | state | evidence |
|---|---|---|
| **populate** | LIVE | `feed-triage` extracts dated event facts into `site_specs.aspect='evidence_base'`. `[MEASURED 2026-09-04]` 472 facts / 25 sites, **83 carrying `event_date`**, written by `evidence-refresher` (78) and `evidence-researcher` (5) |
| **correct** | LIVE | the existing citation-refresh arm of `refresh_evidence_base_action.go` re-verifies any fact carrying `source.citation`. No new code was needed |
| **render** | LIVE | `query.upcoming_events` (`queryresolve/upcoming_events.go`), registered PBP-048 |
| **the page** | **WORKING** | boxingonline.com/tools/fight-calendar serves two real, dated, cited fixtures via `event-list` (`3647c0c2`), rebuilt 2026-09-04 10:36Z. Confirmed by the `boxingonline.com` lane at the artefact |

So 427's originating symptom — *a calendar tool page with zero fixtures* — is **fixed**.
The mechanism it said did not exist now exists and is running.

## 2. The defect that remains, stated precisely

**Provenance is computed, stored, and then discarded at the last step, so nothing
downstream can tell a fact-backed fixture from a fabricated one.**

Verified at the source, both ends:

- `projectUpcomingEvents` (`queryresolve/upcoming_events.go` ~246) builds every item as
  `{"fact_id": html.EscapeString(e.FactID), "title": …, "date": …, "disclaimer": …}`.
  **The provenance is there**, and `content_components.input_schema` for `event-list`
  declares `fact_id` as a field of `items` with `"source": "query.upcoming_events"`.
- The `event-list` `html_template` renders `.date`, `.title`, `.venue`, `.broadcaster`
  and `.source_url`. **It never mentions `.fact_id`.** One field, never referenced.

`[MEASURED 2026-09-04]` consequences, two independent measurements agreeing:

- on the served calendar page: `data-fact-id` **0**, `CIT-` **0** (measured by the
  `boxingonline.com` lane, cache-busted, control `<body>`=1);
- fleet-wide across **335** active `component_level='tool'` components: `data-fact-id`
  **0**.

> **CORRECTED 2026-09-04, before submission, by the blast-radius query this plan's own §6
> asks for — and the correction makes the case stronger, not weaker.** The paragraph above
> was drafted saying provenance is dropped, full stop. It is not. It is dropped by two
> component families out of three, and **carried by the third under an attribute name no
> checker knows**. The narrow claim would have shipped a design keyed on the wrong property.
> What caught it: running §6's own disconfirming query instead of quoting the plan at
> myself.

**The measured state, fleet-wide** `[MEASURED 2026-09-04]`, over all **3,475**
`page_components`. Fourteen placements were rendered from fact-backed data
(`content_data` carrying a `fact_id`):

| component | placements | carries provenance to markup | how |
|---|---|---|---|
| `evidence-chart` | 10 (9 pages) | **0** | — |
| `evidence-timeseries` | 3 (3 pages) | **3** | `data-series="{{$s.fact_id}}"` |
| `event-list` | 1 (1 page) | **0** | — |

So **3 of 14** fact-backed placements declare where their figures came from, and they do it
under a name nothing reads: `[MEASURED]` exactly **one** active component emits
`data-series`, and **nothing** in `platform/`, `internal/` or `pkg/` — Go, CSS or JS —
consumes it. It is a write-only provenance marker, i.e. the "field with no reader" the
estate already warns about, arrived at independently by whoever wrote that template.

### 2.1 The defect, restated from the measurement

**Provenance emission is per-component folklore.** One family solved it privately, two did
not, and there is no shared vocabulary, so no checker can be written that reads all three.
That is a **shared-vocabulary** defect, which is the class CLAUDE.md's owner ruling of
2026-07-29 §1 is about — and it is a much better-founded problem than "one template forgot
a field", which is what this plan said before the query was run.

The three fact-bearing components also draw from **three different source kinds**, which is
the second thing the measurement corrected:

| component | declared source of the fact-bearing field |
|---|---|
| `event-list` | `"source": "query.upcoming_events"` — a `queryresolve` resolver |
| `evidence-chart` | `"source": "site_specs.evidence_base.facts"` / `.charts` — a direct spec read |
| `evidence-timeseries` | `"source": "llm"`, with `fact_id` a **required** property of each item |

### 2.2 Why that is the whole problem, and not a cosmetic one

The page that is doing everything right and the page that is fabricating **score
identically zero** on `data-fact-id`. The `boxingonline.com` lane's framing, which this plan
adopts:

> the "0 matches" figure is **anti-informative** for this class, not merely incomplete —
> it cannot see the defect either way.

The estate has already written this rule down for itself, in the doc comment of the field
that exists to prevent it (`discovery_checks/registry.go:110`, `CheckResult.Resolved`):

> *an empty result indistinguishable from a healthy site (016b §9: a gate's 0 findings has
> two causes with opposite fixes)*.

## 3. Why NOT a wider fabrication pattern

Enumerating fabrication shapes has now failed **five** times in two days, on one component
(`tool-fight-countdown-boxingonline-com`, live, six invented fights):

1. an ISO-date-literal pattern — the fixtures use `year: 2025, month: 5, day: 14`;
2. a `data-fact-id` resolution check — the component carries no `data-*` at all;
3. `check_event_fixture_completeness` — the other side of the seam (facts declared
   without evidence, not content with no fact behind it);
4. `check_tool_fabrication_action.go` Tier A — wants the model to *declare* it invented data;
5. Tier B — needs 15+ entity records (there are 6) **and** corroboration that an original
   was data-backed, which does not exist at birth.

The shape moved each time: ISO string → numeric triplet → keyed object instead of array →
6 records instead of 15. **A blacklist of shapes cannot bound the next generator's
output.** So this plan inverts the polarity: make provenance a **positive, declared,
mechanically checkable** signal, so that its ABSENCE is detectable whatever shape a
fabrication takes.

> **The `482` lane found the sharpest version of this while this plan was being written,
> and it is recorded here because it changes how small the fence's fix is.** Running
> `DetectToolFabrication` over this lane's five census hits: three of them **populate
> `Signals`** — "large literal record array (~20 / ~24 / ~30 entity objects)", all over
> `fabLiteralRecordThreshold = 15` — and still return `Fabricated=false`, because the
> corroboration arm gates on `dataBacked`, which is structurally false at birth. The
> signature is reachable and reached; the **conviction** is unreachable. The gate already
> knows and throws the finding away.
>
> That is the same shape as this bug's own half: the resolver computes `fact_id` and the
> template throws it away. **Two independent instances of "the estate already knows and
> discards it" on one root cause.** Neither needs new detection. Both need the existing
> evidence to survive to a reader.

## 4. The design

### 4.1 The seam, and why it is the right one

`RenderTemplate` (`component_library.go:1060`) is **the only spelling** —
`render_seam_one_spelling_test.go` fails the build if a second appears (owner ruling
2026-08-21, RFC_041 §5). All ~13 render call sites arrive there. The file states the
governing idiom three times, for `form_action`, for `AbsentRequiredFields` and for
`InstanceID`:

> *make the guarantee mechanical HERE, where they all arrive* … *Enumerating call sites is
> what goes stale, so the report lives HERE.*

It already carries `ctx.InputSchema` (the component's declared field contract) and the
template text. That is everything needed to answer *"does this component declare a
fact-bearing source whose provenance the template does not carry?"* — with **no database
handle, no network, and no knowledge of the site**. A pure function of two inputs the seam
already holds.

### 4.2 Phase 1 — declare which query sources are fact-bearing, in ONE place

> **CORRECTED 2026-09-04: this phase was drafted as a property of the `query.*` RESOLVER
> and that would have covered 1 of the 3 fact-bearing components.** §2's table is why:
> the three draw from a resolver, a direct spec read, and an LLM field with a required
> `fact_id`. Keying on `queryresolve` would have seen only `event-list`.

**Fact-bearing is a property of the declared SCHEMA, not of the data's source.** A field is
fact-bearing if its declared item shape carries a property named `fact_id` or `*_fact_id`.
That definition is source-agnostic, and `[MEASURED 2026-09-04]` it is already satisfied by
all three components today — `event-list`'s `items.fact_id`, `evidence-chart`'s
`facts`/`charts` from the register, and `evidence-timeseries`'s `series[].fact_id`
(`"required": ["fact_id"]`) plus its `max_fact_id`.

It lives beside `queryresolve`'s `queryHandlers`, which is already *"the ONE home of the
`query.*` vocabulary"* (its own comment, `:55`) — but as a schema predicate the render seam
can evaluate, so it covers the two components that never touch a resolver.

- **A ratchet test** asserts that every component whose schema declares a fact-bearing
  field emits the shared provenance attribute, so a future one that forgets fails the
  build. This is the mechanism the `482` lane is building on their side
  (`tool_content_writer_coverage_test.go` / `provenanceExemptWriters`); by agreement today
  this rail's writers join **that** map rather than a second one, satisfying its contract
  by declaring provenance rather than by calling the fabrication gate.

This is the estate's single-source rule, learned expensively twice (`council-scope.sh`
vs `098`'s hand-kept `SCOPE_PATHS`; the dedup index vs `workItemTerminalStatuses`). A
checker that enumerates fact-bearing sources itself would drift the same way.

### 4.3 Phase 2 — carry provenance into the markup

Adopt **one** attribute — `data-fact-id` — across all three fact-bearing components, each
`{{if}}`-guarded so a missing id never renders `data-fact-id=""`.

- `event-list` — add it to the `<article>` element. Currently emits nothing.
- `evidence-chart` — add it to each rendered figure. Currently emits nothing (10 placements).
- `evidence-timeseries` — **add `data-fact-id` alongside its existing `data-series`; do not
  rename.** `[MEASURED]` nothing in the estate reads `data-series`, but "no reader in the
  DB and no reader in the repo" does not prove no reader in a deployed page's own CSS, and
  this change gains nothing by removing it. Whether `data-series` should then be retired is
  §9's question, not this phase's.

**This is three shared rows, not a per-site migration.** `[MEASURED 2026-09-04]` there is
exactly **one** row per function — `event-list` is `3647c0c2`, `component_level='section'`,
shared, `is_active` — and the deployed blast radius is **14 placements on 13 pages**. That
matters because this lane already burned three migrations (719/727/728, §19/§21)
discovering that per-plan section rows get reverted by the tier-1 loader sync-down. A
shared library row is not that class of object.

**The attribute goes on a VISIBLE element carrying `data-*`, deliberately** — the shape
§23.1's review already established. It puts the fixture inside the claims perimeter
(`ExtractAssertionText` walks text nodes *outside* `script`/`style`/`noscript`/`template`;
an `<article>` is inside), and it gives a no-JS fallback for free. A JSON blob in a
`<script>` would put it back in the blind spot §22.5 is about.

### 4.4 Phase 3 — make the gap detectable at the seam. REPORT ONLY, OPT-IN, DEFAULT OFF

In `RenderTemplate`: if `ctx.InputSchema` declares a field whose `source` is a fact-bearing
`query.*` and the template does not reference that field's provenance key, publish it on a
new `RenderContext` OUTPUT field (working name `ProvenanceUnrendered []string`), mirroring
`AbsentRequiredFields` exactly.

**It must NOT be a refusal, and that is a ruling, not a preference.** The long note at the
`UnboundInstanceToken` write site records that council round 1 on `661bcf00` came back
REVISE on a guardian HIGH for exactly this move on exactly this function: converting the
seam's log-only path into a hard error is *"new authority on a shared seam"* shipped
*"unconditional on the strength of a census of TODAY's templates"* — and *"a census of
TODAY's live templates cannot bound tomorrow's caller"*. CLAUDE.md's owner ruling of
2026-08-02 §2 governs: new authority on a shared seam ships as an **opt-in field with the
unsafe default OFF**.

### 4.5 Phase 3b — THE READER, so the field cannot rot

**This is a hard requirement of the plan, not a nicety.** The same struct already carries a
cautionary example: `UnboundInstanceToken` has **no reader** — `[MEASURED]` its only
non-test occurrence outside its own declaration is the line that writes it — and its own
doc comment says so:

> *NOTHING READS IT YET … A field with no reader is a mechanism that can rot — so if
> RFC_050 is answered "do not arm", DELETE this field rather than leaving it as
> decoration.*

So the reader ships in the same commit, modelled on the fully-worked precedent in
`render_site_components_action.go:1187` — `recordAbsentRequiredFields(config)` gating an
`emitRequiredFieldsMissing(...)` work-item write, council-reviewed (`98852baa`,
`3626629a`), with its own comment stating the discipline:

> *OPT-IN, unsafe default OFF: this is a new DB write on a shared render path … Unset means
> today's behaviour, byte for byte.*

A `provenance_unrendered` work item, dedup-keyed per page+slot, filed only when the config
key is set. **No refusal arm is proposed at all** — not even unarmed.

### 4.6 Phase 4 — the artefact checker, with a DENOMINATOR

New discovery check `rendered_fact_provenance`. For each page component whose stored
`content_data` carries items with a `fact_id`:

| number | meaning |
|---|---|
| `facts_in_content_data` | **the denominator** — how many fact-backed records this component was rendered from |
| `facts_carried_to_markup` | how many appear in `rendered_html` as `data-fact-id` |
| `facts_resolving` | how many of those resolve to a **current** `evidence_base` fact with citation url AND quote |

**What makes it non-anti-informative, stated as the disconfirming result it must produce:**

- a **zero denominator on a component whose schema declares a fact-bearing source** is a
  FINDING, not a pass. That is precisely today's drop, and it is the case a naive "all
  `data-fact-id` resolve" checker scores clean;
- on the working calendar **today, before Phase 2**: denominator 2, carried **0** →
  **FINDING**. After Phase 2: 2 / 2 / 2 → clean, and it **retracts** its own item;
- on the fabricating countdown: denominator 0 **and no fact-bearing source declared** →
  reported **OUT OF SCOPE**, explicitly, not as clean. That is the honest boundary with the
  `482` lane's fence, which owns "asserts facts, declares no source".

It populates `CheckResult.Resolved` (retracting), which RFC_010's owner ruling requires
before anything is armed into the `capability_gap` pile — `[MEASURED 2026-09-03]` 334
filed, 1 closed.

### 4.7 Phase 5 — arm `check_event_fixture_completeness`

Built, council-reviewed, retracting, and armed on **zero** of the five discovery agents.
Its commit `d6a952249` is an ancestor of the live chassis, so arming cannot fail the run
step. It fires ~nothing today and establishes the baseline.

## 5. Ordering, and why it is load-bearing

**1 → 2 → 3+3b → 4 → 5.** The rule the estate learned the hard way: never ship a control
before the thing it controls can express the good shape. A birth-time refusal shipped
before a generator could emit provenance would make calendars unbuildable (§23.2's own
sequencing note); a checker shipped before Phase 2 would file a finding against every
consuming page at once with no remedy available.

Phase 4 is deliberately AFTER Phase 2 for a second reason: until provenance is in the
markup, the checker's denominator is the only number it can report, and a checker that can
only ever say "0 carried" teaches nobody anything.

## 6. Blast radius — the queries, and what would disconfirm the design

All five were RUN before submission, not left as intentions. Two of them changed the
design (§2's and Phase 1's correction blocks) — which is the argument for running a
disconfirming query rather than listing it.

| step | query | result `[MEASURED 2026-09-04]` | would have disconfirmed if |
|---|---|---|---|
| 2 | `count(*) FROM content_components WHERE function='event-list'` | **1** | **> 1** — not a single shared row, so Phase 2 becomes a per-site migration |
| 2 | fact-bearing placements, by component | **14 on 13 pages** (10 chart / 3 timeseries / 1 event-list) | a large number — a template change re-renders every one |
| 4 | `count(*) FROM page_components WHERE content_data::text LIKE '%fact_id%'` | **14** | **0** — the checker would have no subjects and Phase 4 should wait |
| 4 | provenance actually in markup | `data-fact-id` **0**, `data-series` **3**, of 3,475 placements | all 14 already carrying it — no defect to fix |
| 1 | fact-bearing sources other than `query.upcoming_events` | **yes — two more kinds** (a direct `site_specs.evidence_base.*` read, and an `llm` field with a required `fact_id`) | this is the one that DID disconfirm: it broke Phase 1's original resolver-keyed definition |
| 3 | consumers of `render_site_components` | **7 carriers, no `ActionInputSpec`** | — see §7's caveat; the budget check cannot see a key added there |

## 7. Council and architecture scope

**Ordinary council gate, not an RFC.** Reasoning, offered for the council to reject rather
than asserted:

- Phase 2 is additive and inert — nothing reads `data-fact-id` until Phase 4. Under the
  owner ruling of 2026-07-29 §1, additive-and-inert is not the same as
  additive-and-guarantee-changing, and only the second is architecture-scope.
- Phase 3 adds new authority to a shared seam, which *is* the architecture trigger — except
  that RFC_022's narrowing (owner, 2026-08-11) exempts an opt-in field whose unsafe default
  is OFF and which no live consumer names, and the direct precedent on **this same
  function** (`recordAbsentRequiredFields`, council `98852baa`) went through the ordinary
  gate in exactly this shape. **The consumers must be enumerated, not asserted** — RFC_022
  says asserting it without the query is itself the objection — and that enumeration is
  §6's third row.

**Honest caveat on the budget check.** `render_site_components` has **no
`ActionInputSpec`**, so `audit-optional-key-budget.sh` reports its optional surface as
UNKNOWABLE rather than zero, and a config key added there is invisible to the N=10 budget.
That is a pre-existing gap in the instrument, not a licence: it is stated here so the
council can weigh it, and it argues for declaring a spec rather than for quietly adding an
unseen key.

## 8. Explicitly NOT in this plan

- **The fabrication fence** — the `482` lane's, by agreement today: routing all three
  tool-writing paths through the existing gate and ratcheting membership
  (`tool_content_writer_coverage_test.go` / `provenanceExemptWriters`), plus census
  remediation.
- **Widening `ExtractAssertionText` to script bodies.** Architecture-scope under the
  2026-07-29 ruling. Both lanes agree today that if the rail lands and the fence consumes
  it, this becomes **duplicative rather than deferred** — a declared provenance removes the
  need to *infer* factuality from text, which is the unbounded half. Recorded in the bug
  file's §23.2 so a future reader does not file it as an outstanding gap.
- **Forward event research** (only 2 upcoming facts on boxingonline; the consumer is
  showing everything it has). That is a supply gap and it belongs to `news_feed_ingestion`.
- **Rebuilding the fight-calendar tool component** (`e5e8fa33`, inactive). A site decision
  for the owner and `site_delivery_and_editor`, and one overtaken by the page now working
  via `event-list`.

## 9. Open questions this plan does not settle

- Whether the provenance key should reuse `site_plan_sections.assigned_fact_ids`'
  vocabulary rather than introduce a second one. `assigned_fact_ids` binds a SECTION to
  facts at plan time; `data-fact-id` binds a RENDERED RECORD to the fact it came from.
  They are different scopes and probably both right, but the question deserves the
  council's answer rather than mine. `plan_sections_action.go:1566` already carries a
  remedy string for "matches no current evidence_base fact", which is the same predicate
  Phase 4 evaluates — that duplication is the part to scrutinise.
- Whether Phase 4 should read the SERVED artefact or the stored `rendered_html`. Stored is
  cheaper and is what every other discovery check does; served is what a visitor gets, and
  §22.4 proved on this very bug that assembly can drift from the declared plan. `[UNMEASURED]`
