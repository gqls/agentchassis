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
