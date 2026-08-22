# NOTES — live-object declaration drift

Append-only, newest at the bottom. Technical log: evidence, commands, what the system
actually said, and every misstep.

---

## 2026-08-22 — session start: the lane exists because bug 317 did NOT

Asked to fix `bugs_open/317`. **It is not there.** Closed and moved to `bugs_closed/`
on 2026-08-21 (`c12133e2e`), one day before this session.

I did not take the close banner's word for it. Three legs, re-checked today:

1. **Live column** — the artefact, not the file:
   ```sql
   SELECT name, enabled, (regexp_match(pre_query, 'NOT IN \(([^)]*)\)'))[1]
   FROM scheduled_tasks WHERE name='claimed-item-timeout';
   ```
   `enabled=t`, exclusion list = **14 item types, `dark_section_audit` present**. [MEASURED 2026-08-22]

2. **The union of both gate rosters is exactly those 14.** I nearly filed a false finding here —
   see the misstep below.

3. **The guard is green at committed HEAD.** `git archive HEAD` into a clean tree (not the
   shared dirty one), `go test ./platform/orchestration/actions/
   -run TestClaimTimeoutExclusionCoversBothCompletionGates -count=1` → **ok, exit 0**, at
   `45b728b01`. [MEASURED 2026-08-22]

317's fix holds. The lane below is a DIFFERENT defect, found while verifying it.

### MISSTEP 1 — I counted 12 where the live list said 14, and briefly believed the guard was broken

`grep -rn "RegisterVerifier(" .../discovery_checks/` returned **11** verifier types.
`noChangeGates` holds **1** (`dark_section_audit`). 11 + 1 = 12, against a live list of 14.
The two unaccounted types were `hardcoded_section_colors` and `needs_brand_head_assets`, and
for a few minutes I had a reverse-direction guard failure that the test manifestly was not
reporting.

**The error:** I grepped one spelling of the registration call. There are **two** —
`RegisterVerifier(` and `RegisterVerifierWithPolicy(`, the second added when RFC_017 made
verifier error-handling fail-closed with an opt-in field. Both funnel into the same registry
(`verifiers.go:143-149`), and `RegisteredVerifierItemTypes()` returns both. 13 verifiers + 1
gate-1b entry = **14**, exactly the live list.

**The cheap check that would have caught it:** grep the REGISTRY, not the call site —
`RegisteredVerifierItemTypes()` is the function the guard itself uses, and reading its
implementation names every writer of the map in one step. A grep for one call spelling
measures my guess about the API, not the API. Logged to `WRONG_CALLS.md`.

### The real finding: a guard that reads history and calls it a specification

While chasing the above I read the guard, and the thing it reads is not the live object.

`claim_timeout_exclusion_lockstep_test.go:48` globs
`*_claimed_item_timeout_generic_evidence.sql` and parses `item_type NOT IN (...)` out of the
newest match. That match is `482`, and in `482` the clause exists **only in a comment**
(line 27) — deliberately, and the file says so in its own header: the migration applies a
`replace()` of a tail fragment, so the applied SQL never contains the full clause, and the
comment is nominated as the declaration for the test to parse.

Meanwhile the live object has since been edited again by
`524_claimed_item_timeout_honours_the_cooldown.sql` — **a filename the glob does not match.**

So: the guard reads a comment in a historical file; the sweep reads
`scheduled_tasks.pre_query`; nothing ties the two.

### Confirmed absences (the disconfirming checks, both came back negative)

- `schema_migrations` records a **checksum of the FILE** (`run-migrations.sh:265,453`), never
  of the live object. It answers "was this file applied", never "does the live object still
  match what we believe".
- **Nothing in the repo reads `pg_get_functiondef`, `pg_trigger` or
  `information_schema.triggers`** — `grep -rn` over `*.go|*.py|*.sh` excluding
  `sql_for_agents`: **zero hits**. There is no live-object drift check of any kind.

Both of these could have come out the other way, and would have shrunk or killed the finding.

### Census of the class [MEASURED 2026-08-22]

`scheduled_tasks`: **97 rows, 47 enabled, 35 with a `pre_query`, 24 enabled with one.**
Of the 35, **8 enabled rows embed a literal quoted vocabulary** in their predicate
(`pre_query ~* 'IN \s*\(\s*'''`): `claimed-item-timeout`, `database-cleanup`,
`detected-item-promoter`, `held-pair-canary-escalation`, the three
`site-discovery-rotation-*`, and `site-render-audit-rotation`.

Go guards that read a repo migration file as a specification — **9 test files**, of which
**7 assert a property of a LIVE DB-resident object**:

| # | guard | live object |
|---|---|---|
| 1 | `TestClaimTimeoutExclusionCoversBothCompletionGates` | `scheduled_tasks.pre_query` |
| 2 | `TestGoAndSQLAgreeOnTheCooldownBoundary` (506) | dispatch read SQL |
| 3 | `TestDivergenceVocabularyMatchesMigration344` | trigger `trg_site_component_archive` |
| 4 | `page_component_divergence_test.go:219` (357) | trigger pair `trg_page_component_artefact_archive_upd/_del` |
| 5 | `TestValidDocSubjectTypes_LockstepWithMigrationCheck` | CHECK constraint `doc_plans_subject_type_check` |
| 6 | `TestRenderSlotNameConfigKeyMatchesTheSeededWorkflow` | `agent_definitions` workflow (reads a SEED) |
| 7 | `TestMigration302CarriesTheCanonicalPredicateVerbatim` | dispatch predicate SQL |

`write_experience_pattern_test.go` and `contact_info_no_fabrication_test.go` also read the
migrations dir but may be asserting repo-corpus conventions, where the repo genuinely IS the
subject. **[UNCLASSIFIED]** — pending the plan's judgement, not asserted either way here.

### A weaker claim I am NOT making

I checked whether `524` introduced a cooldown-boundary drift against the `506` guard. It does
**not**: `524` *stamps* `retry_after` on the reset arm, it does not *read* it with a `<`/`<=`
boundary. Recording this because the analogy was tempting and wrong, and 016b's own history
says the untreated sibling is where the next incident comes from — but only when the mechanism
actually transfers. It did not here.

### Prior art checked before filing

- `bugs_open/136` (live config says `*_domain` where code reads `*_pipeline`) — **neighbour, not
  duplicate.** Its surface is `agent_definitions` step-config KEYS that nothing reads; mine is a
  live SQL/DDL object whose only guard reads an untied repo copy. 136 brings reusable machinery:
  `cmd/config-key-audit`, `scripts/audit-config-keys.sh`.
- `needs_diagnosis` queue: **0 rows** at `awaiting_diagnosis` before I filed. No duplicate.
- `bugs_open/341` is the only open bug on the `claimed-item-timeout` sweep and is **OWNED and
  ACTIVE** (`scripts/who-owns.py 341` → `bugfix_307_terminal_write_contract`, 17 commits/14d).
  Not touched.
- LANDMINES: the `doc_subjects subject_type` entry (line ~657) independently documents instance
  5 and its two-enforcement-point shape. Relevant, and its "grep the VALUE you are adding, not
  the table you are changing" check is the same family.

### Diagnosis loop

Filed per the OWNER RULING of 2026-07-31 (a `bugs_open/` file asserting a cross-cutting or
structural root cause is not filed until it has been through the loop).
`CORRELATION_ID=c236fbb4-ca7f-4540-8170-8b806f40fc54`,
**`RUN_CORRELATION_ID=c8ec6478-5a54-4a16-aaf1-1e3373684ba0`** ← the key the artifacts carry.
