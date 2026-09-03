# RUNBOOK — layout fit / bugs_open/445

Every command here had to be got right once. Gotchas attached.

## r1. Where a site's layout actually lives (three hops, no single table)

A peer lane looked for this and could not find it: `css_themes` has no `site_id`,
`style_collections` has no `layout_id`, `site_plan_sections.layout_id` is per-section.

```sql
SELECT s.domain, l.name AS layout
FROM sites s
JOIN style_collections sc ON sc.id = s.style_collection_id
JOIN css_themes t        ON t.id  = sc.css_theme_id
JOIN layouts l           ON l.id  = t.layout_id;
```
⚠ **This is the only census that sees every site.** `resolved_composition` covers 33 of 38 —
`SelectStyleCollectionAction` (`v3_site_actions.go:67`) points a site at an existing collection
and writes no lineage row at all.

## r2. The matcher's own recorded score — and the regex you need until the roll

```sql
SELECT s.domain, ss.data->>'layout_name', ss.data->'lineage'->>'layout_source',
       ss.data->>'reasoning'
FROM site_specs ss JOIN sites s ON s.id = ss.site_id
WHERE ss.aspect = 'resolved_composition' AND ss.is_current;
```
The score is inside `reasoning` as prose: `weighted match: score 3.05 (tags 2.30), …`.
Parse with `score ([\d.]+) \(tags ([\d.]+)\)`. **After commit `76db94fc7` rolls**, read
`lineage.layout_match_score` and `lineage.layout_fit` instead — structured, no regex.

## r3. ⚠ ANY all-time count of work items MUST union the archive

```sql
SELECT 'live' src, count(*) FROM site_work_items      WHERE item_type = $1
UNION ALL
SELECT 'archive', count(*) FROM site_work_items_archive WHERE item_type = $1;
```
Closing a row **moves** it to `site_work_items_archive` (33,350 rows, back to 2026-02-22). I
published "one item, fleet-wide, ever" to four lanes from the live table alone; it is two.
Before quoting any all-time figure:
```sql
SELECT table_name FROM information_schema.tables
WHERE table_schema='public' AND table_name LIKE '%work_item%';   -- four here
```
⚠ **And the estate does this TWO different ways.** `site_specs` does **not** archive — it
versions in place under `is_current = false`. A table listing finds the work-item archive and
tells you nothing about specs; there, `is_current` is the trap.

## r4. What the model was actually SENT (not what the template says)

The single most valuable query of the day. A prompt template is not a prompt.

```sql
SELECT substring(prompt_rendered from position('Current library tags (match these' in prompt_rendered) for 400)
FROM llm_call_log
WHERE step_name = 'classify_and_extract' AND prompt_rendered LIKE '%Current library tags (match these%'
ORDER BY created_at DESC LIMIT 1;
```
⚠ **Anchor on a long, unique phrase.** `position('Current library tags' …)` matches the *JSON
schema example* earlier in the same prompt and returns the wrong region — it looked like a
successful read and showed the wrong text. The column is `prompt_rendered` (there is no
`prompt` column); `prompt_template` is the unrendered source.

Companion — the allow-list that decides what reaches the template at all:
```sql
SELECT default_config #>> '{workflow,steps,classify_and_extract,config,input_fields}'
FROM agent_definitions WHERE type='domain-research-classifier'
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```
A step can populate `collected_data` and still be invisible to the model if its key is not in
that list. That was the cause of the `null` tag list.

## r5. Applying one migration without sweeping every other lane's

`--apply` takes EVERY pending file. Scope by directory:
```bash
S=<scratch>/mig735; mkdir -p "$S"; cp docs/agent_docs/sql_for_agents/735_*.sql "$S"/
MIGRATIONS_DIR="$S" bash scripts/migration/run-migrations.sh            # dry run
MIGRATIONS_DIR="$S" bash scripts/migration/run-migrations.sh --apply
```
⚠ The dry run prints *"ran to its own COMMIT without error (everything rolled back)"* — that
means the SQL executed cleanly, **not** that it is applied.
⚠ A verify block made of `SELECT`s **cannot stop the COMMIT** (`ON_ERROR_STOP` ignores a
non-empty result set). Use `DO $$ … RAISE EXCEPTION … $$;`. 735 does, and its guards abort if
an anchor phrase is missing rather than silently replacing nothing.

## r6. Building while another session's WIP breaks the tree

`go build ./platform/...` failed on `tool_acceptance_actions.go` — a file I never touched, dirty
in another session's tree. Do **not** stash (forbidden, and hook-blocked).
```bash
scripts/verify-head-builds.sh --with <my file> --with <my other file> --test -- ./platform/orchestration/actions/
```
Overlays only your files onto clean HEAD. `--with` is repeatable; `--test` runs tests.
⚠ It surfaced a **pre-existing** failure too —
`TestStylesheetGutted_TokenSetMatchesCanonicalCSSTokens` in `discovery_checks` fails on clean
HEAD with zero changes. Re-run with no `--with` before assuming a failure is yours.

## r7. Firing 090 on a code-only symptom

```bash
SEED_SCOPE="platform/.../file.go:SymbolName,platform/.../other.go:OtherSymbol" \
  ./docs/.../090_TRIGGER_needs_diagnosis_v1.sh "<symptom>"
```
⚠ **Without `SEED_SCOPE` a code-only run FAILS** at `assemble_bundle`
(`no scope (tried "route.scope.Symbols", "input_data.seed_scope", then code_results)`) after
~6 minutes, and burns the item's only attempt (`max_attempts=1`). The script warns
`nothing to key coverage on … dispatching blind` and continues. Read its output to the end.

## r8. The council submission schema (it is NOT what the header implies)

`.plan` is an **object**, not an array:
```json
{ "rationale": "…", "plan": { "summary": "…", "edits": [ … ≤8 … ], "grounded_in": [ "…" ] } }
```
`grounded_in` lives **inside** `.plan`. `operation` ∈ `modify|add|remove|config_change` —
`create` is rejected. The commit-msg hook refuses a `Council-Submitted:` value that is not a
UUID, so **submit first (seconds), then commit naming the correlation**.

## r9. Simulating a layout before recommending it

`scratchpad/score.py` re-implements the matcher and **reproduces the system's own recorded score
on 29 of 30 sites** — that agreement is the control; without it the simulation proves nothing.
`simulate.py` adds a candidate layout and re-scores the fleet, printing who moves.
⚠ Re-derive `df`/`N` after appending a layout, or every weight is computed against the old
library size.

## r10. Is the Go fix in the running fleet? (three instruments, in order, with controls)

The `build provenance` recipe failed here for **two** reasons at once. Do them in this order.

```bash
# (a) which build, and how old are the pods?
kubectl -n ai-persona-system get pods -l app=agent-chassis \
  -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.containers[0].image}{"\t"}{.status.startTime}{"\n"}{end}'

# (b) the startup stamp — MATCH THE STRUCTURE, never the phrase
kubectl -n ai-persona-system logs <pod> --tail=400 2>/dev/null \
  | grep -oE '"caller":"[a-z-]+/main\.go:[0-9]+","msg":"build provenance","git_commit":"[a-f0-9]{40}"'
#   then: git merge-base --is-ancestor <your-commit> <that sha>   (+ a control that must NOT be)
```
⚠ **`grep 'build provenance'` on the chassis returns a HIT that is not a stamp.** LANDMINES prose
about build provenance is synced into `doc_notes`, injected into agent prompts, and logged. I got
back the landmine warning me about this exact trap. The structural pattern above cannot match it.
⚠ The stamp is a **startup** line: at 3h05m old both replicas had rotated it out. Empty there means
*out of range*, **not** *unstamped*.

```bash
# (c) the binary probe — no shelf life. BOTH controls, EVERY replica.
for lit in <your-new-literal> <another> <must-be-PRESENT-control> <must-be-ABSENT-control>; do
  kubectl -n ai-persona-system exec <pod> -- grep -ac "$lit" /proc/1/exe
done
```
Measured 2026-09-03 on `v1.0.1359` (both pods): `weak_tag_fit` 1, `layout_match_score` 1,
`enforceListingItemSources` 2, `zzq_literal_that_cannot_exist_4417` 0. On `v1.0.1358` the first two
were **0**, which is what makes this a measurement rather than a reading.
⚠ Never grep the binary for **your commit sha** — only the build's own commit is stamped, so absent
proves nothing. Ancestry is a git question.

## r11. Did my council submission actually get a verdict — or did the run die?

**`098` cannot tell you this and will say "queued" for ever.** Ask the run, not the report:

```sql
SELECT orchestration_name, current_step, status, last_activity
FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';

SELECT kind, created_at FROM diagnosis_artifacts       -- a verdict exists iff council_report does
WHERE correlation_id = '<SUBMISSION_CORR>' ORDER BY created_at DESC;
```
`FAILED` + no `council_report` ⇒ **resubmit**, do not wait. A `fix_plan` row alone is not a verdict.
⚠ The bare correlation query is a **jsonb path scan and times out** on this table. Bound it:
`WHERE updated_at > now() - interval '10 hours' AND owner_agent_type='council-gate'`, then filter.

Reading the verdict itself (the note is a summary; the report has the per-seat objections):
```sql
SELECT body FROM doc_notes WHERE categories ? 'council-gate' AND body LIKE '%<corr-prefix>%';
SELECT body FROM diagnosis_artifacts
WHERE correlation_id='<corr>' AND kind='council_report' ORDER BY created_at DESC LIMIT 1;
```

## r12. Answering an "asserted, not verified" objection about tag matching

Seats are told `layouts`/`site_specs` are outside their schema, so they flag tag-vocabulary claims
they cannot check. The answer is usually one grep, and **both sides must canonicalise**:

```bash
grep -n "func canonicalTag" -A 25 platform/orchestration/actions/fork_theme_composition.go
grep -n "siteTerms := canonicalSet\|layoutTags := canonicalSet" platform/orchestration/actions/fork_theme_composition.go
```
`editorial|publication|magazine|editorial-content → editorial-publication` (`:146`); site terms
canonicalised at `:227`, layout tags at `:278`. So a site emitting `editorial` **does** reach a
layout tagged `editorial-publication`. Checking only one side would have "confirmed" the objection.
