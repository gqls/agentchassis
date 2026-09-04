# NOTES — tool fabrication fence (bugs_open/482)

Running record. Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-09-04 — session start: ownership, validity, and the two `[UNMEASURED]`s closed

**Ownership.** `scripts/who-owns.py 482` returned OWNED-or-recently-active, pointing at
`site_delivery_and_editor`. That verdict is about the lanes that *cite* it. Read individually:
the bug file's own header says **"Status: OPEN, unowned"**; §7 says *"None of these is
`calendar_component`'s build"*; the `site_delivery_and_editor` handoff (UPDATE 9) cites it as a
report, not as work. `ListAgents` showed no session named for 482 and no session claiming the
fix. **Resumed as the fixing lane.** The `calendar` lane confirmed by message: *"No objection,
it's yours."*
⚠ *The check that mattered was reading the five hits, not the verdict line.* `who-owns.py` is
advisory and answers "who has touched this", which for a widely cross-referenced bug is
everybody.

**Validity, re-derived rather than trusted** `[MEASURED 2026-09-04]`:

```sql
SELECT id, name, function, component_level, is_active, length(html_template),
       input_schema IS NULL, created_at
FROM content_components
WHERE function IN ('tool-fight-countdown','tool-fighter-comparator');
```
→ both `component_level='tool'`, both `is_active=t`, `input_schema` NULL on both,
13,279 and 21,426 bytes, born 2026-08-31. All six invented fights still present verbatim.
**The bug is live and unchanged.**

**§3's `[UNMEASURED]` — which agent generated these, and did it have `evidence_base`?**
Answered: the birth path is `tool-generator` → `create_tool_component_action.go`. Its workflow
steps are `ensure_site_record, load_brand_context, load_site_page_names, compose_plan,
write_plan, index_plan, generate_tool_html, suggest_related_pages, save_tool, enqueue_rerender,
complete`. **No step loads or consults `evidence_base`.** The `generate_tool_html`
`prompt_template` is 5,118 chars, 22 numbered rules about ids, colours, IIFEs and readouts, with
**zero** occurrences of fabricate / invent / evidence_base / provenance / verify. So the answer
to "ignored it or never had it" is **never had it** — which decides §3's own fork in favour of
"make the generator unable to fabricate", not "make the generator consult evidence_base".

**§6's `[UNMEASURED]` — is this a class?** See the misstep below. Short answer: yes, and my first
attempt to measure it said no.

---

## 2026-09-04 — the finding the whole fix now rests on

Looking for what *should* have caught this, rather than designing a new checker first.

`platform/orchestration/actions/check_tool_fabrication_action.go` — 459 lines, built for
`bugs_open/020`, council-reviewed, negation-aware via the shared `datahelpers.NegationGuard`,
tiered A (declaration) / B (corpus signature + corroboration). **It already exists and it is
live.** `[MEASURED 2026-09-04]` it is named in the `default_config` of exactly **one** active
agent definition: `tool-recreation-handler` (steps `check_fabrication`, `route_fabrication`,
`request_fabrication_review`). The birth path does not consult it.

**Probe method, so it can be criticised.** A temporary `_test.go` in the `actions` package
calling the exported pure core `DetectToolFabrication(recreation, original, analysisDataBacked)`
against real `html_template` bytes pulled from the live DB, **with a positive control in the same
run** (the vetcomparison shape from `bugs_open/020`: a "realistic, deterministic dataset" comment
plus `makePostcode`). File removed after measuring; it is re-created as a permanent regression
test by the fix.

```
countdown.html                     fabricated=false tier=""            signals=[]
comparator.html                    fabricated=false tier=""            signals=[]
tool-budget-kit-builder…           fabricated=false tier=""            signals=[large literal record array (~20 entity objects)]
tool-sfi26-revenue-stacker…        fabricated=false tier=""            signals=[large literal record array (~24 entity objects)]
vetcomp.html                       fabricated=false tier=""            signals=[large literal record array (~30 entity objects)]
CONTROL(vetcomp-020 shape)         fabricated=TRUE  tier="declaration" signals=[declared synthetic/fake data…; synthetic identifier generator introduced: makePostcode]
```

The control convicts, so the zeros are real zeros and not a broken harness.

> **⚠ CORRECTION 2026-09-04, to my own claim made earlier the same day.** I told the `427` and
> `boxingonline.com` lanes that *"Tier B is UNREACHABLE at birth"*. **Unreachable is the wrong
> word and the right word matters.** Three of the five hits **populate `Signals`** — 20, 24 and
> 30 entity objects, all over `fabLiteralRecordThreshold = 15`. The *signature* is reachable and
> reached. What is unreachable is the **conviction**: Tier B gates on
> `dataBacked && !preserved`, `dataBacked` comes from an `original` that a born tool has never
> had, so the branch cannot be taken and the computed signals are returned "for observability"
> and gated on nothing.
> **Why the distinction is load-bearing:** it means the birth arm is not new detection work. The
> evidence is already computed, already correct, and thrown away. That is a far smaller and more
> defensible change than widening any pattern, and it is the argument the fix should be built on.
> Caught by feeding the detector more inputs than the two tools in the bug — i.e. by the `427`
> lane's census, not by my own reasoning.

**Second instance of the same shape, from `427`:** `projectUpcomingEvents`
(`queryresolve/upcoming_events.go` ~246) emits `fact_id` into every item map; the `event-list`
template renders `.date/.title/.venue/.broadcaster/.source_url` and **never `.fact_id`**. So the
provenance is computed and discarded too. Two independent instances of *"the estate already knows
and throws it away"* on one root cause.

---

## 2026-09-04 — MISSTEP: I censused the shape I already knew about

Full entry in `WRONG_CALLS.md`. Short form, kept here because the correction is the useful part:

I ran §6's missing census over all **335** active tool components keyed on **date shapes**
(`year:`, `month:`, `new Date(20NN,`, `Date.UTC(`, ISO-valued `date:`, `data-fact-id`) and got
**1**. I reported that to two lanes as *"a first occurrence rather than a class"*. The `427` lane
censused the same rows on an **entity+attribute** axis — a record that identifies a real-world
thing and attributes a checkable property to it — and found five candidates, three of them real.
The worst, `tool-vet-comparison-vetcomparison-uk`, **contains no date at all**, so no widening of
my predicate could ever have reached it.

*The check that costs one sentence:* **write down what the CLASS is, in words, without reference
to the instance in hand, then check your predicate is a predicate for the class.** "Tools with
stale dates" and "tools asserting real-world facts nothing verified" are different populations and
only the second is the bug.

*The free check I skipped:* **ask the lane that has been reading the corpus what axis they would
use, before picking one.** `427` had the better predicate and was one message away.

---

## 2026-09-04 — verifying `427`'s census hits at first hand rather than carrying them forward

All four named rows exist, `component_level='tool'`, `is_active=t`: `tool-vet-comparison-…`
(16,944 B, born 09-02), `tool-sfi26-revenue-stacker-…` (17,323 B, 08-24),
`tool-budget-kit-builder-…` (19,575 B, 08-25), `tool-loot-table-balancer-…` (12,254 B, 07-17).

Vetcomparison re-derived independently: **30** postcode-bearing records, **30**
`https://example-vet-*.co.uk` hostnames, four sample records byte-identical to the quoted ones.
Placement `[MEASURED 2026-09-04]`: `vetcomparison.uk` **`/index.html`**, `build_status='deployed'`,
`pages.deployed_at = 2026-09-03 21:19:33+00`.

**The re-derivation is what earned its keep** — it turned up three things the report had not
mentioned:
- `:298` `// Bundled, verified sample of practices. Self-contained — no fetch().` — the component
  asserts its own verification;
- `:290-291`, tool-doc header: *"Never seeds or fabricates practice records beyond the bundled
  list; if this list needs to grow, it must be replaced with a verified set"* — the prohibition
  and the violation in one file;
- `:40`, **served to the public**: *"Practice details shown here are a fixed sample bundled with
  this tool… please confirm anything important directly with the practice. The RCVS maintains the
  official register of accredited practices."*

⚠ **Note for the fence design:** `:290`'s wording (*"Never … fabricates …"*) is exactly the shape
`fabNegationGuard` exists to suppress — a denial of fabrication in a file that is fabricating.
`bugs_open/222` built that guard because a *conscientious* model echoing the prompt's prohibition
was the common case. Here the same words sit above the violation. **The guard is still right** (it
prevented a real false positive class) but this is the first observed case where the denial and the
act are in the same artefact, and it belongs in the calibration argument for any birth arm.

**Bitter footnote:** `bugs_open/020`, the bug that caused the fabrication gate to be built, was
filed about **vetcomparison.uk**. The gate was written for that site; that site is still
fabricating, in a shape the gate was never taught to convict.

---

## 2026-09-04 — calibration set received

`427` committed `docs024_key_docs_latest/bugfix_427_event_render/CENSUS_2026-09-04_tool_embedded_datasets_calibration_set.txt`
(commit `561b66fc2`) — all **134** dataset-bearing tools with record counts and key sets.
They corrected their own earlier figure of 133 → **134** (case-folding of key names) before I
built on it. **134 is the false-positive denominator any threshold I commit must be simulated
against**, per `component_write_guard.go`'s own doctrine: every threshold there was simulated
against the full live history first, and **two candidate checks were dropped because the
simulation caught them misfiring on legitimate rewrites**. Neither column in that file is a
verdict — its own header says the wide count is mostly legitimate UI vocabulary.

---

## 2026-09-04 — THE SIMULATION KILLS PLAN ITEM B AS DRAFTED (89% false positives)

Ran the threshold simulation I owe under `component_write_guard.go`'s doctrine, over `427`'s
134-tool calibration set, **before** committing any threshold. The draft plan's item B was
"give Tier B a birth arm so a computed signature is not discarded". Simulated as the obvious
version of that — **drop the corroboration requirement at birth, keep `fabLiteralRecordThreshold
= 15`**:

```
tools with dataset_records >= 15 : 28
  of which entity+attribute >= 2 : 3     ← true positives
  FALSE POSITIVES                : 25    ← 89%
```

**A gate with an 89% false-positive rate is not a gate.** This estate's own written doctrine is
that *"a guard that refuses good work gets switched off, and then it protects nothing"*
(`component_write_guard.go` header), and 25 legitimate tools routed to human review on their
birth day is precisely how that happens. **This is the "second incident" the review was asked to
find, and the simulation found it in one command.** Draft item B is withdrawn in its stated form.

**It is worse than a tuning problem — the threshold fails in BOTH directions.** The distribution
over the 134 (min 2, median 7, p90 23, max 73) puts plenty of legitimate tools above 15, and the
motivating case sits below it:

| component | dataset_records | entity+attr | verdict |
|---|---|---|---|
| `tool-vet-comparison-vetcomparison-uk` | 30 | 30 | **fabrication** |
| `tool-sfi26-revenue-stacker-agritec-uk` | 25 | 24 | **fabrication** |
| `tool-budget-kit-builder-garden-tools-uk` | 18 | 20 | **fabrication** |
| **`tool-fight-countdown-boxingonline-com`** | **6** | **6** | **fabrication — THE BUG, below the threshold** |
| `tool-loot-table-balancer-gamesdesign-co-uk` | 3 | 3 | probably legitimate game vocabulary |
| `tool-bee-foraging-calendar-apis-uk` | **73** | 0 | legitimate (flowers, label, shift, stage) |
| `tool-garden-jobs-finder-homegarden-uk` | 45 | 0 | legitimate |
| `tool-archetype-clash-calculator-vonc-com` | 37 | 0 | legitimate |

So no threshold on record count separates these populations. The largest dataset in the corpus
(73 records) is entirely legitimate and the bug that started this lane has 6.

> **⚠ CORRECTION to this lane's own PLAN §3, same day, before it was acted on.** Item B said
> "calibrate the threshold against the 134-tool set". **The finding is that there is no threshold
> to calibrate.** Record count is not the discriminator in either direction. **The discriminator is
> the KEY SET** — does a record identify a real-world entity *and* attribute a checkable property
> to it (`postcode`, `website`, `rate`, `price`, `venue`, `event_date`) — which is exactly the axis
> `427` chose for their census and I did not choose for mine. Their `entity_attr_records` column
> separates 5 from 129 where the record count separates nothing.
> On `ea >= 2` the same simulation gives **5 convictions, ~4 true** (loot-table-balancer is the
> likely false positive) — ~20% FP, routed to human review rather than to breakage, which is a
> defensible gate. **That is the predicate to build, and it is not a widening of Tier B; it is a
> different question asked of the same corpus.**

⚠ **One inconsistency in the calibration data, flagged to `427` rather than relied on.**
`tool-budget-kit-builder-garden-tools-uk` reads `dataset_records = 18`, `entity_attr_records = 20`.
`ea` is documented as *"the narrower subset"* of `ds`, and a subset cannot be larger than its
superset — so the two columns are counting over different extractions, or one has an off-by-one
in its record splitting. **It does not change the conclusion** (that row is a true positive on
either number, and the false-positive count is driven by the `ea = 0` majority), but the
`ea` column cannot be quoted as a strict subset count until its author says which. Recorded
because a number I am about to build a predicate on had a visible defect and saying nothing would
have made it load-bearing by silence.

---

## 2026-09-04 — the `ds`/`ea` discrepancy is resolved, and the resolution changes which predicate I port

`427` diagnosed the `ds=18 / ea=20` row I flagged (`8f13f2043` corrects their file header).
**Neither column is a subset of the other; the header claiming so was wrong.** They are
independent predicates over the same extracted object set:

- `ds` requires a **human-readable string value** (≥8 chars containing a space);
- `ea` requires an **identity key AND an attribute key**, and says nothing about string content.

**Verified at first hand rather than accepted** `[MEASURED 2026-09-04]`, against the real template:

```js
{ name: 'Secateurs',   tier: 1, always: true,                              low: 20, high: 45,  guide: '/buying-guides/secateurs' }
{ name: 'Loppers',     tier: 2, activities: ['pruning','landscaping'],     low: 30, high: 65,  guide: '/buying-guides/loppers' }
{ name: 'Wheelbarrow', tier: 3, sizes: ['medium','large','allotment'],     low: 65, high: 140, guide: '/buying-guides/wheelbarrows' }
```
**Three** records with a single-word `name`, exactly as described — they carry an identity key and
an attribute key (so `ea` counts them) and no string value of ≥8 chars containing a space (so `ds`
does not). The mechanism is confirmed.

⚠ *Minor, recorded so the columns are not over-trusted:* the arithmetic does not close —
18 + 3 = 21, not 20 — so the two predicates must also differ on at least one further record.
**Does not affect any conclusion** (the row is a true positive on either count and the
false-positive figure is driven by the `ea = 0` majority), but the columns are not related by a
clean offset and should not be quoted as though they were.

> **⚠ CORRECTION to this lane's own reasoning of an hour earlier.** I wrote that I would "port the
> `ea` extractor" on the assumption it was the *narrower* of the two, i.e. the conservative choice.
> **That reasoning was wrong even though the choice was right.** `ea` is not narrower — it is
> differently permissive, and the difference matters in my favour: `ds`'s human-readable-string
> filter **discards single-word entity names**, and an invented product, an invented practice, an
> invented fighter is very often a single word. Had I ported the "narrower subset" I described, I
> would have inherited a filter that drops a large share of the fabrication class the gate exists
> to catch. **I picked the right predicate for a reason that was false**, which is worth more to
> record than picking a wrong one for a good reason: the next person reasoning "narrower =
> safer" over these two columns will be wrong the same way.

Also confirmed by `427`: both columns share one object regex matching **innermost** brace pairs
(`[^{}]`), so nested structures are counted at their leaves — a shared property, not a defect in
either, and not a bug I inherit.

---

## 2026-09-04 — the rail's shape, and the honest limit that keeps the content test alive

`427` submitted the provenance rail to council (`68a8e2a3-aa4b-477f-83f1-69a317cd82c4`); plan at
`docs024_key_docs_latest/bugfix_427_event_render/PLAN_2026-09-04_provenance_rail.md` (`c0d245bd0`).
Shape, as it bears on this lane's birth arm:

- `platform/orchestration/datahelpers/fact_provenance.go` — the single home of the vocabulary, in
  a package both lanes can import (`discovery_checks` already imports `datahelpers`; it cannot
  import `queryresolve`, which imports `discovery_checks`).
- `FactBearingFields(schema)` is a **SCHEMA predicate, not a source predicate** — a field is
  fact-bearing if its declared item shape carries `fact_id` / `*_fact_id`. Source-agnostic on
  purpose: the three fact-bearing components today declare three different sources.
- `ExtractRenderedFactIDs(html)` is attribute-scoped, never a substring scan.
- The declaration is **structured and resolvable, never prose** — stated as the design's first
  property, explicitly because of this lane's vetcomparison `:290` finding (a tool-doc header
  prohibiting fabrication, in the file doing it).

**What it gives this lane:** the birth arm can ask `len(FactBearingFields(schema)) > 0` — a
structural question about the *declaration* — instead of re-deriving factuality from key names.
**Neither lane then maintains an identity/attribute vocabulary**, which was the drift risk in my
draft.

⚠ **The limit, stated by `427` unprompted and worth keeping in front of the design.**
`FactBearingFields` reads the *declared schema*, and `[MEASURED 2026-09-04]`
`tool-fight-countdown`'s `input_schema` is **NULL** — as most fabricating tools' will be. So the
predicate separates *declared-and-backed* from *declared-nothing* cleanly, **but cannot
distinguish "declared nothing because it has no data" from "declared nothing because it invented
some"**. That second step still needs the `ea` content test. **The rail removes the need for a
vocabulary; it does not remove the need for a content test.** Any plan that treats the rail as
sufficient is wrong, and this lane will not write one.

---

## 2026-09-04 — a NULL schema is the DEFAULT, not a signal: the declaration half is near-inert as evidence of guilt

`427` corrected their own message within ten minutes, on two counts. The second is a design
constraint and this lane has re-derived it independently rather than accepting it
`[MEASURED 2026-09-04, by this lane]`:

```sql
SELECT count(*) AS active_tools,
       count(*) FILTER (WHERE input_schema IS NULL)   AS schema_null,
       count(*) FILTER (WHERE input_schema::text='{}') AS schema_empty
FROM content_components WHERE is_active AND component_level='tool';
--  335 | 287 | 0        → 85.7% NULL
```

**287 of 335 active tools have no `input_schema` at all.** All three fabricating components
discussed on this bug are among them — and so are 284 others, most of them entirely fine.

> **⚠ CORRECTION, to the design direction this lane recorded one hour ago.** The NOTES entry above
> records the rail's `FactBearingFields(schema)` as the thing that lets the birth arm ask a
> *structural* question instead of a content one. **On today's corpus that predicate is satisfied
> by 86% of tools, including nearly every innocent one, so as an INCULPATORY signal it is close to
> inert.** A conjunction that leans on it is the 89%-false-positive shape in a new costume — the
> same failure this lane already found by simulation, re-entering through a different door.
>
> **The asymmetry is the honest statement:**
> - as an **exculpatory** test the declaration is genuinely useful — a tool that *does* declare a
>   fact-bearing field and whose ids resolve is clean, cheaply and certainly, and this grows in
>   value as the rail lands;
> - as an **inculpatory** test it discriminates almost nothing today, because declaring nothing is
>   the norm.
>
> **So the `ea` content predicate does essentially all the discrimination in the birth arm, and
> the plan must say so rather than presenting a two-part conjunction as if both parts carried
> weight.** Credit: `427` caught this on their own mechanism and published it against their own
> interest, before any seat asked.

**Second count, recorded because it is a practice failure worth the fleet seeing** (`427` is
logging it in `WRONG_CALLS.md` themselves): they stamped `[MEASURED 2026-09-04]` on the
`input_schema IS NULL` fact **taken from this lane's message**, not run by them. The figure was
correct; the marker was not. A `[MEASURED]` marker on a number you did not measure yourself reads
as first-hand and is not — and the round trip here is a good illustration of the remedy: they
re-ran it, the figure held, **and re-running it is what surfaced the 86% that inverted the design
advice.** Re-deriving a correct number was not wasted work; it was the work.

**Consequence they published for their own rail, unprompted:** `FactBearingFields` reads the
declared schema, so the rail's reach is bounded by schema adoption — **48 of 335 tools**. It works
where it is aimed (all 14 fact-bearing placements declare one) but it is *not* a general
tool-provenance mechanism today, and they will not present it as one.
