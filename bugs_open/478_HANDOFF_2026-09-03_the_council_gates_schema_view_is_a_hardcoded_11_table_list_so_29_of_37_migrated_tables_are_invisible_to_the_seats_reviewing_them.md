# 478 — the council gate's schema view is a hardcoded 11-table list framed as "the ONLY tables available", so 29 of the 37 tables that in-scope migrations write to are invisible to the seats reviewing them — including `site_specs`, the second most-written table in the estate

**Filed:** 2026-09-03, `copy_quality_two_stage` lane, at the owner's direction after two seats on
council corr `e7e0ae76` (migration `748`) objected that they *"cannot independently verify"* a
row in `agent_default_configs` because the table *"is not in the schema this council can query"*.
**Status: OPEN. Diagnosed first-hand at every link (the query, the prompt framing, the sibling
code path, the blast radius); no `090` run — stated substitution per the 2026-07-31 ruling, because
every claim below is a read of a file or a query the reader can re-run, not an inference.**

## 1. The symptom, as the seats reported it

Two seats on `748`'s round, verbatim:

> *guardian/medium:* "Target table `agent_default_configs` is not in the schema this council can
> query — I cannot independently verify the row is unique per config_name … or how many live rows
> share config_name='build_standard_block'."
>
> *prior_art_librarian/medium:* "Migration targets `agent_default_configs` … a table absent from
> the schema I can query — I cannot verify this table's existence or the carrier row's current
> shape via SQL, so the whole guarded-migration design (idempotency check, verify block) rests on
> the author's word."

The seats were right that they could not see it. They were wrong — and told to be wrong — about
what that absence meant.

## 2. The mechanism (read, not inferred)

**The list is hardcoded in one `query_database` step.** `agent_definitions` row `type='council-gate'`,
step `load_schema_hint`, `output_field: schema_hint`:

```sql
SELECT string_agg(t.line, chr(10)) AS text FROM (
  SELECT 0 AS ord, table_name || '(' || string_agg(column_name || ' ' || data_type, …) || ')'
    FROM information_schema.columns
   WHERE table_schema = 'public'
     AND table_name IN ('pages','sites','site_plans','site_plan_pages','site_work_items',
                        'content_components','page_components','agent_definitions',
                        'diagnosis_artifacts','agent_error_log','doc_notes')
   GROUP BY table_name
  UNION ALL … (the LANDMINES index)
) t
```

Eleven names. Nothing else exists, as far as the seats are told.

**And they are told it in the strongest possible terms.** Every seat's prompt (all 17 `review_*`
steps, read from `review_guardian`, shared prefix hoisted by migration `377`) opens its evidence
block with:

```
## Schema (the ONLY tables available to checks)
{{.schema_hint.text}}
```

**"ONLY."** So a table absent from the eleven is not presented as *unlisted* — it is presented as
*nonexistent for review purposes*. The seat's correct response to that framing is exactly what
happened: decline to verify and say so. The defect is upstream of the seat.

**The sibling code path already solved this and the council did not inherit it.** The diagnose
loop builds its schema in `platform/orchestration/actions/diagnose_load_runtime_action.go` and is
built differently on all three counts:

| | diagnose loop (`diagnose_load_runtime_action.go`) | council gate (`load_schema_hint`) |
|---|---|---|
| selection | `defaultSchemaInclude` prefix match (`site%`, `page%`, `content%`, `flow%`, `workflow%`) **plus** `schemaAlwaysTables` (7 names) **plus** the tables the bundle draws evidence from | 11 literal names, full stop |
| framing | `schemaFilterNotice`: *"This listing is FILTERED, not the whole database: N of M public tables are shown … A table's absence here is NOT evidence it does not exist."* | *"the ONLY tables available to checks"* |
| escape | *"To read one that is not listed, issue a `data_request` against `information_schema.columns` — you do not need a human to confirm it."* | none stated; `data_request` appears in the council config 3 times but never as a route to an unlisted table |

The diagnose loop's own comment on why the notice exists (`:854`) is the exact failure `748` hit:
*"a table that was filtered out and a table that does not exist render IDENTICALLY, and the
verdict prompt's cite-or-abstain acts on absence."* The council gate has the identical trap with
none of the mitigation, and the two lists have drifted independently — `agent_definitions` is in
both; `agent_default_configs` is in neither; `site_specs` matches the diagnose loop's `site%`
prefix and is absent from the council's eleven.

**Verified the council does NOT receive the diagnose bundle:** corr `e7e0ae76` has exactly two
`diagnosis_artifacts` rows, `fix_plan` and `council_report`; neither contains `Schema (live
tables)` or `This listing is FILTERED`. The council's schema view is `schema_hint` and nothing else.

## 3. Blast radius — measured, and it is the finding

**Census 2026-09-03** over `docs/agent_docs/sql_for_agents/[6-7]??_*.sql` (migrations 600–799,
i.e. roughly the last five weeks), counting `UPDATE` / `INSERT INTO` / `DELETE FROM` targets:

| writes | table | council can see it? |
|---:|---|---|
| 177 | `agent_definitions` | yes |
| **115** | **`site_specs`** | **NO** |
| 59 | `site_work_items` | yes |
| 49 | `content_components` | yes |
| 33 | `page_components` | yes |
| 29 | `scheduled_tasks` | NO |
| 21 | `sites` | yes |
| 20 | `governor_state` | NO |
| 17 | `schema_migrations` | NO |
| 16 | `site_plan_sections` | NO |
| 15 | `migration_backups` | NO |
| 12 | `pages` | yes |
| 12 | `governor_config` | NO |
| 9 | `doc_notes` | yes |
| 7 | `governor_work_class_map` · `content_sources` | NO |
| 6 | `form_submissions_inbox` | NO |
| 5 | `site_plan_imagery` · **`agent_default_configs`** | NO |

**37 distinct tables written. 8 visible. 29 invisible.** Migrations have been in council scope
since 2026-08-19 (`bugs_open/314` — *"a migration IS the running system, live the moment it
applies"*), which is the whole reason a schema view exists for the seats at all. For the second
most-written table in that window — `site_specs`, which holds `evidence_base`, `identity`,
`commercial`, `content_direction`, `strategy`: every site's premise — **the council has been
reviewing 115 writes to a table it was told does not exist.** `agent_default_configs`, the
instance that raised this, is five writes; it is the smallest thing on the list.

The census counts statements in files, not applied migrations, and covers one number range; it
is a lower bound on the class, not a census of the estate. **Counted 2026-09-03.**

## 4. What it costs, concretely

- A seat that cannot see a table cannot check uniqueness, versioning, current shape, or whether
  the guard the author describes is even well-formed against real columns. On `748` two seats
  said exactly that and approved anyway on the author's word — which is the outcome the council
  exists to avoid. On the 115 `site_specs` writes the same thing happened silently: no seat said
  "I cannot see this", because the framing gave them no reason to think there was anything to see.
- The approval is credited by `098` as a review. So the coverage report reads "reviewed" for
  changes whose target table no reviewer could inspect.
- The failure is invisible in the same way the diagnose loop's comment describes: absence reads
  as nonexistence, and nothing flags it.

## 5. Fix candidates, ordered by what closes the door

1. **Replace the literal list with the diagnose loop's selection, and reuse its notice.** Make
   `load_schema_hint` call the same table-selection logic (`defaultSchemaInclude` +
   `schemaAlwaysTables` + evidence tables), and put `schemaFilterNotice`'s text into the seat
   prompt in place of "the ONLY tables available". One selection, one framing, one escape, and the
   two paths stop drifting. This is the structural fix and the reuse-first one.
2. **At minimum, change the framing.** Even with the list unchanged, replace *"the ONLY tables
   available to checks"* with the FILTERED wording and the `data_request` escape. This costs one
   prompt edit across the shared prefix (migration `377`'s hoisted block — see the `099_SYNC`
   suspension in CLAUDE.md before touching it) and converts a dead end into one more read-only
   query, which is what the diagnose loop's comment says it did there.
3. **Derive the list from the submission.** Add the tables the plan's `edits[].file` migrations
   write to (a grep the trigger script could do client-side) to `schema_hint` per round. Closes
   the gap for the specific change under review without widening the default.
4. **Do nothing but document it** — rejected: the seats' framing actively tells them the table does
   not exist, so a note nobody reads does not change the behaviour that produced `748`'s two
   objections.

**Not a fix:** adding `agent_default_configs` (or `site_specs`) to the eleven. That closes two
instances and leaves the class, the drift, and the framing.

## 6. How to verify a fix

Resubmit a migration touching `site_specs` (any `_HOLD` in `sql_for_agents/` will do) and read
the seats' objections: **none may say "not in the schema I can query"** for a table the plan
names. Then the negative control: submit one touching a genuinely absent table name and confirm
a seat says so via the escape rather than by abstaining. A fix that silences the first without
preserving the second has hidden the problem, not fixed it.

## 7. Related

`bugs_open/314` (migrations enter council scope — the reason a schema view matters) ·
`diagnose_load_runtime_action.go:854` (the sibling's own statement of this trap, and its fix) ·
`016b` §9, *"a reviewer told a listing is exhaustive reads absence as nonexistence"* (added with
this file) · `WRONG_CALLS.md` 2026-09-03, the `Council-Submitted`/latest-verdict entry (why an
approval on unseen evidence is credited as a review) · CLAUDE.md's `099_SYNC --apply SUSPENDED`
note (the shared seat prefix must be edited surgically, not regenerated).
