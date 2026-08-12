# 265 — the legacy `input_schema` dialect is declared EXTINCT in a doc comment, is being reintroduced steadily, and the tripwire built to catch that only writes a `Warn`

**Filed 2026-08-12.** Found by the `copy_quality_two_stage` lane while sizing
`bugs_open/260`'s exposure; the reintroduction dates were surfaced by the
`brochure_component_library` front and are reproduced here with their query. **Two
defects that compound: a stale invariant readers trust, and the detector that should have
retired it firing where nobody reads.**

---

## Defect 1 — a load-bearing doc comment asserts an extinction that is false

`platform/orchestration/datahelpers/component_schema_fields.go:53-56`, in
`SchemaContentFields`' own header:

> *"fromLegacy is the fail-loud signal: the legacy dialect is **extinct fleet-wide (0 of
> 173 as at 2026-07-21)**, so a true here means a regression reintroduced it"*

The census was true when written. It is now false, and **every instance postdates it**:

```sql
SELECT function, created_at::date, is_active FROM content_components WHERE input_schema ? 'properties' ORDER BY created_at;
--  report-dossier      | 2026-07-27 | t
--  mechanism-flow      | 2026-07-28 | t
--  evidence-timeseries | 2026-07-28 | t
--  loans-consolidation | 2026-08-10 | t     ← two days before filing
```

`[MEASURED 2026-08-12]` Four active components, none forked, spanning 15 days and still
arriving. **This is not a residue the census missed — it is a steady reintroduction that
began six days after the dialect was declared dead.** The likely producer is the
component-creator path; that is `[UNVERIFIED]` and is the first thing the fixing thread
should establish.

**Why a comment earns a bug file.** It is read as an invariant by anyone writing code
against `input_schema`, and it has already caused one error: this lane specified a
`bugs_open/260` type gate *"against the house dialect"* on the strength of the dialect
question looking settled. That gate would have been blind to all four — **including
`mechanism-flow`, the only component with a proven live render failure.** A comment that
turns a reader's gate into an inert one is doing the work of a defect.

## Defect 2 — the tripwire is well-wired and effectively silent

The platform anticipated exactly this. `WarnLegacyDialect` / `WarnIfLegacyDialect` is
called from **six** sites — both render gates (`v3_site_actions.go:2019`,
`rerender_page_sections_action.go:332`), `plan_sections_action.go:2015`, and two discovery
checks (`check_required_fields_missing.go:105`,
`check_image_source_unsatisfiable.go:129`). The wiring is not the problem.

**Its only output is a `Warn` log line.** The sibling front measured the entire
`RenderTemplate` log family **absent from a 4,661-line 24-hour window on `agent-chassis`**.
So four components have been passing through six tripwires for up to 15 days and nothing
has surfaced. **A detector whose sole output is a warn line on a busy service is not a
detector** — it is the `a-hook-that-writes-to-stderr-reaches-nobody` shape: measure a check
at its READER, not at its call site.

⚠ **The two discovery checks are the sharp end.** They produce work items. A legacy-dialect
component reaching `check_required_fields_missing` degrades a *work-item-producing* check
silently, so the failure is not merely unlogged — it is a check quietly doing less than it
reports. `[UNMEASURED — I have not established what those two checks return for the four
components; that is the highest-value next measurement and it decides this bug's severity.]`

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Correct the comment and date it** — the smallest honest step, and it stops the next
   reader building an inert gate. Not sufficient alone: it will go stale again the same way.
2. **Give the tripwire a destination that is read.** `WarnLegacyDialect` should raise a
   work item (or write a `doc_note`) rather than log, on at least the two discovery-check
   call sites. This converts a silent regression into a queued one.
3. **Refuse the dialect at creation.** If the component-creator path is confirmed as the
   producer, validate there — a component cannot be stored in a dialect the estate has
   retired. **This is the candidate that makes the bad state unrepresentable**, and it is
   the only one that stops the count growing.
4. **Do NOT delete legacy support in `SchemaContentFields`.** It is currently the only
   thing keeping those four renderable, and `mechanism-flow` is already failing for an
   unrelated reason (`bugs_open/260`).

## How to verify a fix

```sql
-- must stay at 4 and stop growing; a 5th means candidate 3 is not in place
SELECT count(*), max(created_at)::date FROM content_components WHERE input_schema ? 'properties';
```
⚠ **Induce before trusting**: create a scratch component in the legacy dialect and confirm
the chosen mechanism actually fires. A zero from a detector nobody reads and a zero from a
fixed producer look identical — which is this bug's whole subject.

## Related

- `bugs_open/260` — the render failure on `mechanism-flow`, one of the four; its fix
  candidate 2 (type-check `content_data` against `input_schema`) is the gate this defect
  would have made inert.
- `docs024_key_docs_latest/copy_quality_two_stage/PLAN_2026-08-12_two_stage_copy.md` §10 —
  where the inert-gate error was made and corrected.
- ⚠ **`component_schema_fields.go`'s cited prior art is stale**: it points at
  `bugs_open/026`, which is CLOSED and is about news-listing hardcoded English. Do not
  follow that pointer expecting dialect history.

---

## ADDENDUM 2026-08-12 — the UNMEASURED severity question, MEASURED: Defect 2's sharp end is REFUTED, and the failure direction is the opposite of the one feared

Contributed by the `brochure_component_library` fact-assignment front (`bugs_open/260`'s owner),
answering §"Defect 2"'s `[UNMEASURED]` note. **Nothing below touches Defect 1** (the stale
invariant and the growing count), which stands as filed.

### 1. REFUTED — the two work-item-producing checks are NOT degraded

The claim under test: *"a legacy-dialect component reaching `check_required_fields_missing`
degrades a work-item-producing check silently … a check quietly doing less than it reports."*

Two independent grounds, both read in code rather than inferred:

- **Both checks read through the dialect-aware helper**, as this file already cites
  (`check_required_fields_missing.go:100`, `check_image_source_unsatisfiable.go:124` call
  `datahelpers.SchemaContentFields`), each with a comment stating that is why. So a legacy
  component is *projected*, not skipped.
- **The projection is LOSSLESS for these four components' actual key usage.** The helper copies
  only `source`, `on_missing`, `fallback`, `missing_reason`, `items`, `min_items` (+ `type`,
  and `llm_guidance`/`description`) and silently drops any other property-level key. Enumerating
  **every** property-level key across all four legacy components — not grepping for expected
  ones — returns exactly four keys:

  | property key | occurrences | components | survives projection |
  |---|---|---|---|
  | `type` | 11 | evidence-timeseries, mechanism-flow, report-dossier | yes |
  | `items` | 2 | evidence-timeseries, mechanism-flow | yes |
  | `minItems` | 2 | evidence-timeseries, mechanism-flow | yes (remapped to `min_items`, helper `:92-96`) |
  | `description` | 1 | report-dossier | yes (falls back to `llm_guidance`, helper `:105-107`) |

  **All four survive. Nothing these components declare is dropped**, so the checks see the same
  field set they would see from a v2 component.

- **And the checks are live, not dormant** — so this is a real negative, not an untested path:
  **111** `required_fields_missing` and **49** `image_source_unsatisfiable` work items exist,
  most recent 2026-08-11.

**Severity consequence: Defect 2's sharp end drops away.** The tripwire is still effectively
silent (Defect 2's main claim, unaffected), but the two discovery checks are not doing less than
they report. Fix candidate 2's urgency for *those two call sites specifically* is therefore lower
than filed — nothing is being lost there today.

### 2. The residual risk is real but points the OTHER way: over-report, not under-report

**No property in any of the four declares `source`** (see the enumeration above — `source` does
not appear). The helper therefore applies its documented default of `source: "llm"` to **every
field of all four components** (`component_schema_fields.go:112-114`, with the reasoning in its
own comment: *"A property with no explicit source is content the writer must supply"*).

So any field of these four that is genuinely renderer- or query-supplied is presented to the
checks as writer-authored. That biases toward **false positives** — a check demanding an
LLM-authored value for a field nothing asks the writer for. That is the opposite failure
direction from the one this bug feared, and it is worth stating explicitly because "the check is
blind" and "the check over-reports" have different remedies and different queues.

### 3. One genuine gap remains, and the dialect is not its cause

`loans-consolidation` is the one of the four carrying **no top-level `required[]`** array. The
projection folds `required` from that array, so **no field of that component is ever marked
required**, and `check_required_fields_missing` can never fire for it. That is a real blind spot
— but it is the component's own schema declaring nothing required, not a dialect-projection
defect, and a v2 component with no `required` flags would be equally invisible.

### 4. What this does NOT clear

Defect 1 is untouched: four components in a dialect the estate's own comment calls extinct, all
four created after the census that declared it so, the newest two days ago. Candidate 3 (refuse
at creation) is unaffected by everything above and remains the only one that stops the count
growing — arguably *strengthened*, since with the checks shown sound, the growing count is now
the principal ongoing harm rather than one symptom among several.
