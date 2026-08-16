# 265 — the legacy `input_schema` dialect is declared EXTINCT in a doc comment, is being reintroduced steadily, and the tripwire built to catch that only writes a `Warn`

> ## STATUS 2026-08-16 — TAKEN UP by the `bugfix_265_legacy_dialect_unrepresentable` lane. FIX BUILT, council submitted (`aba82416-de79-4452-8730-3e35ca0a15bb`), migration 437 written and probe-run; see §"2026-08-16" at the foot for the re-verification, the producer CORRECTION, and what ships where.
> Docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_265_legacy_dialect_unrepresentable/`.
> **Headline correction:** the producer is NOT the component-creator. All four legacy rows were hand-authored SQL seeds/scripts (`created_from='manual'`, `source_agent_type` NULL) — so the fix that stops the count growing is a **CHECK constraint on the table**, the one seam every producer passes through. Population today: **3** (loans-consolidation was converted to v2 by its own lane on 2026-08-15).

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
began six days after the dialect was declared dead.** ~~The likely producer is the
component-creator path; that is `[UNVERIFIED]` and is the first thing the fixing thread
should establish.~~
> **CORRECTED 2026-08-16 (fixing lane):** the producer is **hand-authored SQL**, not the
> component-creator. All four rows are `created_from='manual'`, `source_agent_type` NULL;
> the three seeds are on disk (`sql_for_agents/207` — committed 2026-07-25, four days after
> the census — `247`, `250`), and the fourth was the LMC lane's hand seed of 2026-08-10. The
> component-creator (`created_from='generated'`, 69 rows 03-31→07-06) has emitted **0**
> legacy rows in its life. What caught it: reading the provenance columns instead of
> guessing from the dialect's shape — one `GROUP BY created_from, source_agent_type`.
> Consequence: a gate on the component-creator would have stopped none of the four; the
> table itself is the only seam they all crossed. Full section at the foot of this file.

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

## ADDENDUM 2 (filing session, 2026-08-12) — refutation INDEPENDENTLY REPRODUCED, and the one thing it leaves contingent

**The refutation above is correct and I am not taking it on trust — I re-ran the load-bearing
measurement myself, blind to their method, and got the same answer.** Every property-level key
across all four components:

```sql
SELECT k.key, count(*) FROM content_components c
CROSS JOIN LATERAL jsonb_each(c.input_schema->'properties') AS p(key,value)
CROSS JOIN LATERAL jsonb_each(p.value) AS k(key,val)
WHERE c.input_schema ? 'properties' GROUP BY 1 ORDER BY 2 DESC;
--  type 11 · items 2 · minItems 2 · description 1
```

Four keys, nothing else. And the three survival paths are in the code as described:
`minItems` → `min_items` (`:92-96`), `description` → `llm_guidance` (`:105-107`), and the
`source: "llm"` default (`:112-114`). Both checks do call the helper
(`check_required_fields_missing.go:100`). **So my "a check quietly doing less than it
reports" claim is REFUTED. Struck, not softened.**

> **This is what the `[UNMEASURED]` marker is for.** I flagged that claim as unmeasured and
> named it as the measurement that would decide severity, rather than asserting it in the same
> voice as the rest of the file. It was wrong, it cost one query to find out, and it was
> refuted before anyone routed work at it. Recording that here because the marker earning its
> keep is worth more than the claim was.

**The sharpening the refutation opens, which changes nothing today and dates the result.**
The projection copies **exactly six** property keys — `source`, `on_missing`, `fallback`,
`missing_reason`, `items`, `min_items` (`:87`) — plus the three special cases above.
**Everything else in a legacy property is silently dropped.** So "lossless" is not a property
of the projection; it is a property of *what these four components happen to declare today*.
A fifth legacy component using any ordinary JSON Schema key outside that set — `maxItems`,
`enum`, `pattern`, `format`, `default` — loses it silently. `maxItems` is the sharpest
example, because `minItems` **is** remapped and its twin is not, so the two adjacent keys
behave differently with nothing to signal it.

**And the count is growing** — most recent arrival 2026-08-10, six days after the census that
called the dialect extinct. **So this is a clean result with an expiry date**, and re-running
the enumeration is a precondition for trusting it again, not a one-off. That is an argument
for **fix candidate 3 (refuse the dialect at creation)** and against treating the checks'
soundness as a reason to relax: today's soundness is luck about declaration style, and the
population producing it is still being added to.

⚠ **Do not read "the checks are sound" as "the dialect is harmless."** The two claims have
different lifetimes: the first is measured against four specific rows on 2026-08-12; the
second would need to hold for every legacy component anyone creates from here.

**Adopted from the refutation, unchanged:** the residual risk points at **false positives**,
not blindness — no property of the four declares `source`, so the helper presents every field
of all four as writer-authored. Any that is really renderer- or query-supplied is over-reported.
Different remedy, different queue, and worth carrying in the file rather than folding into the
severity note. And `loans-consolidation`'s invisibility to `check_required_fields_missing` is
its own missing `required[]`, not the projection's doing — a v2 component with no required
flags is equally invisible.

**Net effect on this file:** Defect 1 stands unchanged and is now the whole of the bug's
urgency. Defect 2's *silence* claim stands (the tripwire's only output is still a `Warn` that
nobody reads); Defect 2's *consequence* claim is refuted. Fix candidate 2's urgency drops for
those two call sites specifically; candidate 3 is strengthened.

---

## 2026-08-16 — TAKEN UP: re-verified, producer corrected, fix built (`bugfix_265_legacy_dialect_unrepresentable` lane)

### Re-verification [MEASURED 2026-08-16, clients_db]

| claim | today |
|---|---|
| 4 legacy rows, newest 08-10 | **3**: `report-dossier` (07-27), `mechanism-flow` (07-28), `evidence-timeseries` (07-28). `loans-consolidation` now carries `fields`; `updated_at` 2026-08-15 14:06:40Z; converted by `loanandmortgagecalculator_couk/b2_convert_oldshape.py` (its own lane), no `component_versions` row |
| producer `[UNVERIFIED]` = component-creator | **wrong direction** — see the CORRECTED block in §Defect 1. `created_from | source_agent_type | n | legacy`: `manual|∅|170|3`, `generated|∅|69|0`, `generated|tool-generator|33|NULL schemas`, `manual|tool-deployer|13|0`, `generated|generic|6|NULL schemas` |
| tripwire only Warns | unchanged (`component_schema_fields.go:130-137` at HEAD) |
| comment claims extinction | unchanged (`:53-56` at HEAD) |
| top-level `properties` anywhere else | `component_versions`: 0. Only `content_components` (live) and `component_versions` (history) carry the column outside `bak_*` tables; every Go reader joins `content_components` |

The bug is **still valid**; the count stopped growing only because the last seeder converted
their own row — nothing structural changed.

### What the correction means for the fix candidates

The file's candidate 3 said *"if the component-creator path is confirmed as the producer,
validate there."* It is not the producer, and RFC_009's own CronJob header already states the
general fact: *"content_components live only in the database. A component is routinely changed
by a migration or by hand with no commit at all."* So **candidate 3 lands as a CHECK constraint
on the table** — the seam every one of the four rows actually crossed — with the Go birth-path
check kept as legibility for the one LLM producer, not as the guarantee.

### The fix, as built (council `aba82416-de79-4452-8730-3e35ca0a15bb`, submitted 2026-08-16)

1. **Migration `437_content_components_refuse_legacy_input_schema_dialect.sql`** — guards
   (not applied; population is exactly the 3 ids; every property def is an object) → backup
   table `content_components_bak_20260816_265_legacy_dialect` → UPDATE converting the 3 rows
   by `SchemaContentFields`' projection written in SQL (behaviour-preserving) → `DO`/`RAISE`
   verify (0 legacy left; field-name sets equal old `properties` sets; every field has a
   `source`; `required` flags equal old `required[]`) → `ADD CONSTRAINT
   chk_input_schema_no_legacy_dialect CHECK (input_schema IS NULL OR NOT (input_schema ?
   'properties'))` + `COMMENT`. Probe-run 2026-08-16 (COMMIT swapped for ROLLBACK): all
   guards passed, `UPDATE 3`, verify NOTICE, ALTER, COMMENT; live table confirmed untouched
   after. `_ROLLBACK` sidecar drops the constraint FIRST, then restores by id.
   **Refuses the top level only** — nested `properties` under `fields.<x>.items` is the
   shape of an item (mechanism-flow and evidence-timeseries both carry one) and is fine.
2. **Go, inert until the next chassis roll:** `datahelpers.IsLegacyInputSchemaDialect` (the
   constraint's own predicate — deliberately wider than `fromLegacy`, so the birth path refuses
   exactly what the table refuses); a fourth birth check in `store_generated_component` that
   fails the step with a message naming the house dialect instead of SQLSTATE 23514 after
   `deriveRenderMode` has already mis-read the schema as `template`; the doc comment rewritten
   to cite the constraint (checkable at any moment) instead of a census (stale in four days);
   `WarnLegacyDialect` Warn→Error with a message that says what a firing now MEANS (the
   constraint is gone, or the schema came from a `bak_*`/versions copy or memory); the stale
   `bugs_open/026` pointer → `bugs_closed/026`.
3. **Proof:** `TestLegacyDialectConversionMatchesProjection` runs `SchemaContentFields` on
   the live before-JSON and 437's after-JSON (fixture captured from the rolled-back probe, not
   typed) and asserts identical field maps + `fromLegacy=false` after. Mutation-tested: a
   dropped `required` flag and an always-false predicate both fail. `go build ./platform/...`
   green.

**Deliberately not done:** no new CronJob (RFC_006's daily-check pattern is for invariants the
DB cannot express — this one it can; a constraint does not drift); the reader's legacy
projection is KEPT (deleting it restores the fail-OPEN blind spot 026 was opened about); the
three applied seeds are not rewritten (history — a hand re-run now fails loudly, which is the
intended behaviour); `component_versions`/`bak_*` untouched.

### Two things this surfaces for OTHER lanes (not fixed here — content decisions)

- **`report-dossier` `body` is marked `source: llm` by the projection and therefore by the
  conversion**, yet its seed (`207`) says the body is *"Never authored by an LLM and never
  assembled from a template."* That is exactly the over-report addendum 1 predicted. Behaviour
  is unchanged by 437 (the reader already defaulted it), and 0 `page_components` use the
  component today, so nothing fires — but the honest v2 value is the gripper-dossier lane's
  call (`source: renderer` is the vocabulary the fleet uses for pre-rendered fields: 134 rows).
- **`site-header` (2026-07-17) carries a THIRD shape**: v2 field definitions with no `fields`
  wrapper (`{"header_cta_url": {"type":"url","source":"config.chrome.header_cta_url",…}}`).
  `SchemaContentFields` returns `ok=false` for it ("no declared fields"). Harmless today — no
  `llm` field, chrome path — but it is a v2 schema the v2 reader cannot see. Not this bug's
  dialect and not refused by 437; noted so the next reader of that row does not assume the
  wrapper is present. `[OBSERVED, not measured for consequence]`

### Verification route (owner ruling 2026-07-31)

No `090` run for the producer claim. Substituted: a full-population enumeration on the
provenance columns (it could have shown `generated` rows carrying `properties`, and did not),
plus the three seed files read on disk with the dialect at the cited lines. The fix goes
through the council gate.

### How to verify (once 437 is applied; the Go half once rolled)

```sql
SELECT count(*) FROM content_components WHERE input_schema ? 'properties';         -- 0
SELECT conname FROM pg_constraint WHERE conname='chk_input_schema_no_legacy_dialect'; -- 1 row
-- INDUCE, in a transaction you roll back:
BEGIN; INSERT INTO content_components (name, html_template, function, input_schema)
  VALUES ('zz','<section></section>','zz-scratch-265','{"type":"object","properties":{"x":{"type":"string"}}}');
-- must fail: violates check constraint "chk_input_schema_no_legacy_dialect"
ROLLBACK;
```
A zero from the census is now evidence, because the refusal can be induced.

