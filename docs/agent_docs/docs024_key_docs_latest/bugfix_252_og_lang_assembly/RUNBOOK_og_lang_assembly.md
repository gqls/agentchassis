# RUNBOOK — bug 252 (og/lang slug)

Every command this lane needed, with the gotcha attached. When one changes, change it HERE.

---

## 1. Measure the defect (all of these are re-runnable as verification)

**The stored heads — how many assert a page identity they cannot know:**
```sql
SELECT count(*) AS heads,
  count(*) FILTER (WHERE rendered_html LIKE '%og:url%')                            AS og_url,
  count(*) FILTER (WHERE rendered_html LIKE '%property="og:title" content=""%')    AS blank_ph
FROM site_components WHERE slot_name='head';
-- 2026-08-19: 24 | 22 | 4
```
⚠ **The columns are `rendered_html` and `slot_name`** — not `html` / `component_type`. Both of my
first two attempts failed on guessed names. `\d site_components` first, per CLAUDE.md.

**Per-site breakdown with the assembled-page count, which is what makes it a blast radius:**
```sql
SELECT s.domain,
  (SELECT count(*) FROM pages p WHERE p.site_id=s.id
     AND EXISTS (SELECT 1 FROM page_components pc WHERE pc.page_id=p.id)) AS assembled_pages,
  (length(sc.rendered_html)-length(replace(sc.rendered_html,'og:title','')))/8 AS n_og_title
FROM site_components sc JOIN sites s ON s.id=sc.site_id
WHERE sc.slot_name='head' ORDER BY assembled_pages DESC;
```
The `n_og_title` trick (length-difference / needle length) counts occurrences without a regex; `>1`
is the duplicated-tag signature.

**Which component each site's head actually points at** — there are THREE, not two, and the third
has its own `function`, so a `WHERE function='head'` query misses it:
```sql
SELECT cc.id, cc.name, cc.function, md5(cc.html_template),
       (cc.input_schema ? 'fields') AS wrapped, count(sc.id) AS sites_using
FROM site_components sc JOIN content_components cc ON cc.id=sc.component_id
WHERE sc.slot_name='head' GROUP BY 1,2,3,4,5 ORDER BY sites_using DESC;
```
⚠ `wrapped` is the load-bearing column: **Document Head is FLAT, head-seo-standard is WRAPPED**, so
the two `jsonb_set` paths differ (`'{lang}'` vs `'{fields,lang}'`). And the entry must be
**map-valued** or the resolver silently skips it as "not a field descriptor".

## 2. Verify at the ARTEFACT — the only instrument that settles this class

```bash
curl -s "https://<domain>/<inner-page>" \
  | grep -oE '<meta property="og:[^>]*>|<link rel="canonical"[^>]*>|<html[^>]*>'
```
⚠ **Run it on an inner page AND the homepage.** A per-page defect is invisible on the homepage,
where the site-level value happens to be correct — checking only `/` would have shown nothing wrong.
⚠ `curl | head -N` on a large page prints `curl: (23) Failure writing output` when head closes the
pipe. Harmless, but it looks like a fetch failure; use `grep -oE` (consumes the stream) or `-s -o
/dev/null` patterns instead.

**Post-fix expectations, per page:** `og:url` == the page URL == the canonical href; exactly ONE
`og:title`, carrying the page title; zero `content=""` og tags; homepage `og:url` is the bare `/`,
not `/index.html`.

## 3. Grepping for HTML that Go EMITS

```bash
grep -rn '<html lang' platform/ internal/ pkg/ --include=*.go     # right
grep -rn 'lang="en"'  platform/ internal/ pkg/ --include=*.go     # WRONG — silently misses the emitters
```
⚠ The emitters build the string in Go, so the source reads `lang=\"en\"`. A pattern containing a
quote cannot match a literal that escapes it, and the failure is **silent and directional**: it
returns the decorative occurrences (prompt templates, doc strings) and hides the load-bearing ones.
This cost a wrong "already fixed" call — `WRONG_CALLS.md`, 2026-08-20.

## 4. Build and test when the working tree does not compile

Another session's in-flight work routinely breaks `package actions`. Do not fight it:
```bash
SC=<scratchpad>; rm -rf $SC/head252 && mkdir -p $SC/head252
git archive HEAD | tar -x -C $SC/head252
cp platform/orchestration/actions/{head_assembly.go,head_assembly_test.go,\
rerender_single_page_action.go,render_site_components_action.go} $SC/head252/platform/orchestration/actions/
cd $SC/head252 && go test ./platform/orchestration/actions/ -count=1
```
This is also the honest build: `make build-*` builds from committed HEAD, so this is what will ship.

## 5. Mutation-proving a test — and proving the MUTATION

```bash
cp $F $F.bak
python3 -c "import io;p='$F';s=io.open(p).read();b=s;s=s.replace(OLD,NEW,1);assert s!=b,'did not apply';io.open(p,'w').write(s)"
grep -n '<the mutated line>' $F        # <- DO THIS
go test ... ; cp $F.bak $F
```
⚠ **`sed` with regex metacharacters or backslashes fails silently and exits 0.** One mutation here
reported PASS because the file was never modified — a **false pass**, indistinguishable from a test
that does not discriminate. Use Python with an `assert s != before`, and grep the line before reading
the verdict.

⚠ **A mutation that passes usually means your FIXTURE cannot express the property.** The ordering
mutation passed because the fixture (copied from a live head) carries the exact blank description tag
and so drives `spliceMetaDescription` down its targeted path, where the hazard cannot fire at all.
Build the discriminating input — here: blank og placeholders and NO blank description tag — and print
both orders side by side in a throwaway test. That took 15 lines and settled it in one run.

## 6. Council gate

```bash
DRY_RUN=1 ./docs/.../097_TRIGGER_council_review_v1.sh <submission.json>   # free; tests admission
./docs/.../097_TRIGGER_council_review_v1.sh <submission.json>             # real; prints SUBMISSION_CORR
```
⚠ `.plan.summary` is required and is **not** listed in the script's header comment block — the
validator refuses without it. ⚠ Migrations are now IN scope (2026-08-19, `bugs_open/314`), so the Go
and the SQL go in ONE submission. ⚠ The `Council-Submitted:` trailer is checked by a **commit-msg
gate**: a non-UUID value (I tried `pending`) BLOCKS the commit. Submit first, or omit the trailer.

Watch it:
```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1;
```

## 7. The 090 diagnosis loop

```bash
./docs/.../090_TRIGGER_needs_diagnosis_v1.sh "<symptom>"
```
It prints TWO correlations. **The second one (`RUN_CORRELATION_ID`) is the key the artefacts are
written under** — the intake correlation resolves nothing. Read the outcome from the work item, not
`diagnosis_artifacts` (whose rows here are `kind='bundle'` with no verdict in `metadata`):
```sql
SELECT status, left(result::text, 4000) FROM site_work_items
WHERE item_type='needs_diagnosis' AND spec->>'dispatch_correlation_id'='<RUN_CORR>';
```
⚠ **The loop cannot see served bytes.** For a defect in deployed markup it returns UNVERIFIABLE and
asks for the artefact — it queries `pages.rendered_head` (VESTIGIAL, 0 rows fleet-wide) and truncates
`site_components.rendered_html` before the `</head>` tail. That is not a refutation. Full trap:
`LANDMINES.md`, "The 090 diagnosis loop CANNOT SEE SERVED BYTES".

## 8. Migration numbering

```bash
ls docs/agent_docs/sql_for_agents/ | grep -oE "^[0-9]+" | sort -n | uniq | tail -4
```
⚠ **Run this at the moment you COMMIT, not when you start authoring.** I read 501 as the max, wrote
two files, and by commit time 502–506 had all been taken by three other lanes. 497 and 498 are each
already doubled from earlier collisions. Renumbering afterwards means a `git mv` **naming both paths
on the commit** (a one-sided pathspec ships a copy), plus fixing every internal self-reference —
including any `created_by` marker the ROLLBACK matches on.

## 9. Rollout, in order (the ordering is load-bearing)

1. Roll the chassis; **prove the binary per service, with controls**, before anything else.
2. **Canary the og half — it needs NO migration.** Direct dispatch (`049b_deploy_single_page.sh`),
   not the `spawn_agent`→`call_agent` wrapper, which has hung. Two pages on a duplicated-tag site
   (ai-agent-orchestration.com `/about.html` + `/index.html`), verified with §2.
3. Apply `507` then `508`; record with `run-migrations.sh --record-only` and **scope the directory** —
   `--apply` takes every pending file.
4. The staleness pipe fans out on its own (`stale_chrome` → `rerender-pages` with
   `refresh_site_components:true`). Direct lever if it stalls: the `rerender-chrome` agent (seed 351).
5. Sweep all 26 sites with §2, across **all four head families** — Document Head, head-seo-standard,
   the webdesign.co.uk fragment, and the site with no stored head.

⚠ **THE TRAP: never apply the migrations before the binary is live.** DB config is live on apply, Go
is inert until the roll. The old code would consume the staleness edge, restamp `render_inputs`, and
the pipe would go **quiet with the fleet still wrong**. Both files are `_HOLD` for this reason.
