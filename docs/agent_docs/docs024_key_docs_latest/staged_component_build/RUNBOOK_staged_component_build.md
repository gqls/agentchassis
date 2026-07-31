# RUNBOOK — staged component build

**The commands, each with the gotcha that made it hard to get right.** Fix a command
HERE when it changes, not in scrollback. Lane adopted 2026-07-30.

---

## 1. Read the travelling-docs tables before writing to them

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c '\d doc_plans'
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c '\d doc_notes'
```

**The gotcha, and it blocks P1:** both tables carry a CHECK on `subject_type`.
`doc_plans` allows `tool|pipeline|experience|action|experience-pattern`; `doc_notes`
the same plus `landmine`. **Neither allows `component`.** An insert fails; it does not
coerce. Measured 2026-07-30.

**The half that is good news:** `doc_notes` has `site_id`, `doc_plans` does not — which
is exactly the split a component needs (fleet-wide contract, per-site verdicts), so no
column has to be added. Only the constraint changes.

Population, to see what is actually in use:

```sql
SELECT 'doc_plans' AS t, subject_type, count(*) FROM doc_plans GROUP BY 1,2
UNION ALL SELECT 'doc_notes', subject_type, count(*) FROM doc_notes GROUP BY 1,2
ORDER BY 1,3 DESC;
```
2026-07-30: doc_notes pipeline 362 / tool 73 / landmine 57 / experience 56 / action 16;
doc_plans experience 59 / tool 35 / experience-pattern 24 / action 4.
**Re-run it; do not quote these.**

---

## 2. Count the criteria fences fleet-wide

```sql
SELECT count(*) FROM doc_plans
 WHERE body LIKE '%```criteria%' AND COALESCE(is_current,true);
```

**Gotcha:** this number moves daily — it read 23 on 07-29 and **25** on 07-30, inside
one report's editing window. Never carry it forward from a document.

---

## 3. The change that unblocks P1 — TWO halves, one commit, migration NOT applied

> **CORRECTED 2026-07-30 (see PLAN D5′). This section previously said the DDL was the
> whole change and should be withheld from `sql_for_agents/`. Both were wrong.**
> The DDL alone reproduces `bugs_open/064`, and withholding the number reddens HEAD
> because a test parses the numbered file. The DDL below is retained for reference; the
> shipped change is `273_doc_subjects_component.sql` + `doc_subjects_common.go`
> (`c659e312b`), council `e5673868-7c5b-489c-931a-7ba59b959b91`.

**Half 1 — Go, and it is the half that is easy to miss.** `validDocSubjectTypes` in
`platform/orchestration/actions/doc_subjects_common.go` gates `write_doc_plan`,
`append_doc_note`, `load_doc_context` and `persist_diagnosis_note`. Add `"component"`.
The file's own comment carries the rule: *a value the DB accepts but a Go gate rejects
is a split contract; move both together.* Migration 163 missed one gate; 184 moved the
DB CHECKs only and left its own seeded docs unreachable — that is `bugs_open/064`.

**Half 2 — the migration, numbered normally.** Follow **270** exactly (same shape, done
for `landmine`). It must be numbered because
`TestValidDocSubjectTypes_LockstepWithMigrationCheck` parses the newest numbered `.sql`
recreating `doc_plans_subject_type_check` and fails on drift — which is also why both
halves go in ONE commit.

**Prove the guard is real before trusting it** (this lane's own S2 rule, applied to
itself):
```bash
mv docs/agent_docs/sql_for_agents/273_doc_subjects_component.sql /tmp/hidden.sql
go test ./platform/orchestration/actions/ -run TestValidDocSubjectTypes_LockstepWithMigrationCheck   # expect FAIL
mv /tmp/hidden.sql docs/agent_docs/sql_for_agents/273_doc_subjects_component.sql
go test ./platform/orchestration/actions/ -run TestValidDocSubjectTypes_LockstepWithMigrationCheck   # expect ok
```
Measured 2026-07-30: fails with *"split contract: validDocSubjectTypes = [action
component …] but 218_experience_register_substrate.sql sets the CHECK to [action …]"*.

**ORDER: image, then migration.** Until an image carries half 1, applying half 2 widens
the CHECK past the Go gate — 184's split in a new spot. Mitigating fact so nobody
over-reacts to an early apply: **nothing writes component docs yet**, so the widened
CHECK is inert rather than broken.

```sql
-- allow subject_type='component' on doc_plans and doc_notes
-- Additive and inert: nothing reads 'component' until something writes it, so under
-- the owner ruling of 2026-07-29 §1 this is normal council-gate scope, not an RFC.
BEGIN;

DO $$
DECLARE found_def text;
BEGIN
  SELECT pg_get_constraintdef(oid) INTO found_def
    FROM pg_constraint
   WHERE conrelid = 'public.doc_plans'::regclass
     AND conname  = 'doc_plans_subject_type_check';
  IF found_def IS NULL THEN
    RAISE EXCEPTION 'doc_plans_subject_type_check is absent — read the table before applying';
  END IF;
  IF found_def LIKE '%component%' THEN
    RAISE EXCEPTION 'already allows component — record it, do not re-run';
  END IF;
END $$;

ALTER TABLE public.doc_plans DROP CONSTRAINT doc_plans_subject_type_check;
ALTER TABLE public.doc_plans ADD CONSTRAINT doc_plans_subject_type_check
  CHECK (subject_type = ANY (ARRAY['tool'::text,'pipeline'::text,'experience'::text,
    'action'::text,'experience-pattern'::text,'component'::text]));

-- doc_notes keeps 'landmine' — dropping it would silently invalidate 57 live rows.
ALTER TABLE public.doc_notes DROP CONSTRAINT doc_notes_subject_type_check;
ALTER TABLE public.doc_notes ADD CONSTRAINT doc_notes_subject_type_check
  CHECK (subject_type = ANY (ARRAY['tool'::text,'pipeline'::text,'experience'::text,
    'action'::text,'experience-pattern'::text,'landmine'::text,'component'::text]));

COMMIT;
```

**The gotcha that would bite hardest:** the `doc_notes` re-add must keep `'landmine'`.
Copying `doc_plans`' array across would drop it and orphan **57 live rows** written by
two other threads. Read the constraint you are replacing; do not assume the two tables
agree.

**Before applying, PREPARE-prove any new SQL against the live schema.** `go build`
cannot parse SQL, and the tools lane lost a pilot run to a Postgres type-deduction
error that compiled fine.

---

## 4. Prove a check type is in the RUNNING binary before authoring a gate against it

This is the one that produced a wrong answer first. **Two rules:**

```bash
# 1. There is no `strings` in the browser-runner image — grep the binary directly.
# 2. Use a LONG marker. Go compiles short string literals to immediate comparisons
#    that never reach rodata, so a short marker returns 0 on a binary that supports it.
POD=$(kubectl get pods -n ai-persona-system -o name | grep browser-runner | head -1 | cut -d/ -f2)
kubectl exec -n ai-persona-system "$POD" -- sh -c '
  echo "NEW marker:";  grep -ac "too small to see or click"        /app/browser-runner-adapter
  echo "CONTROL:";     grep -ac "page overflows horizontally"      /app/browser-runner-adapter
  echo "CONTROL:";     grep -ac "in the live DOM after settle"     /app/browser-runner-adapter'
```

> **RE-MEASURED 2026-07-31 and the answer FLIPPED: `has_visible_area` IS now in the
> running pod** — both long markers **1**, on `browser-runner-adapter` built 08:53:36 UTC,
> with three positive controls also 1. The 07-30 measurement below was correct on 07-30 and
> is stale now; keeping it because the *method* is what this section teaches.
> **And a GO from this check is no longer sufficient.** `bugs_open/157` is open and unfixed
> at HEAD (`run_checks_action.go:773-774` still comma-ok asserts `float64` while
> playwright-go returns `int` for whole numbers), so the type is present **and wrong**: any
> integer-sized axis measures 0 and it reports "too small to see or click". Presence in the
> binary answers "can it run", not "is it right" — **grep `/bugs_open/` for the check type
> by name as well as grepping the pod.**

**Measured 2026-07-30:** `has_visible_area`'s two long markers **0**; three long
pre-existing controls **1** each — so TL-034 is committed (`1850acb07`, 15:19) and
**not in the running pod**. My first attempt grepped the type names themselves and got
`has_visible_area` 0 *and* `selector_count` **0**, on a binary that demonstrably
supports `selector_count`. **A negative from a short marker is worthless.**

Why this matters more than a version check: an unknown type is **skipped, not failed**
(`run_checks_action.go`, `default: skip(...)`), and an all-skipped set reads as PASS
plus a 7-day cooldown. A gate authored against an unrolled type passes vacuously and
suppresses its own re-check for a week.

---

## 5. The component render harness (S2), and the two traps in it

Pattern: `scripts/render_teaser_reveal_panel.go` — 24 assertions run against
`html/template` exactly as the render path does, **before any DB write**.

```bash
go run docs/agent_docs/docs024_key_docs_latest/brochure_component_library/scripts/render_teaser_reveal_panel.go
```

- **Slice the `<style>` block away before counting elements.** Counting `.trp__media`
  across the whole document over-counts by exactly one — the CSS rule's own selector.
  This trap was hit **twice** in this lane; it is why 027's S5 gate names it explicitly.
- **A green check nobody has seen go red is not evidence.** Every assertion class needs
  a mutant. The mutants here caught two harness defects, not just template ones —
  including an unguarded `strings.Index` returning −1 that panicked the harness.

---

## 6. Verify a placement is durable (S4), not just present

> **CORRECTED 2026-07-31 — the query below never ran; `site_plan_sections` has neither
> `function` nor `page_id`.** Its columns are `plan_id, page_name, ordering,
> component_name, component_version_id, palette_id, layout_id, typography_set_id`.
> Read the schema before pasting a query out of a RUNBOOK, including this one — the
> original is kept here struck through because a command that was never executable is
> itself worth knowing about.

~~`SELECT ordering, function FROM site_plan_sections WHERE page_id = (SELECT id FROM pages WHERE site_id=$1 AND name=$2) ORDER BY ordering;`~~

```sql
-- a page_components row alone is a RENDER ARTEFACT; site_plan_sections is the PLAN FACT.
-- It keys on NAMES (plan_id + page_name + component_name), not ids.
SELECT sps.ordering, sps.component_name
  FROM site_plan_sections sps
  JOIN site_plans sp ON sp.id = sps.plan_id
 WHERE sp.site_id = $1 AND sps.page_name = $2
 ORDER BY sps.ordering;

-- And to ask the DIFFERENT question "is this component attached to any page at all?",
-- which is the one that refuted an 'orphan' diagnosis on 2026-07-31 — join through
-- page_components. content_components has NO site_id: components are fleet-shared and
-- keyed by `function`.
SELECT p.name, s.domain, p.url, pc.slot_name, pc.build_status
  FROM page_components pc
  JOIN content_components cc ON cc.id = pc.component_id
  JOIN pages p ON p.id = pc.page_id
  JOIN sites s ON s.id = p.site_id
 WHERE cc.function = $1;
```

**The gotcha that cost a false conclusion:** "no page found under the name I expected" is
**not** "no page". `tool-arena-interface` was recorded as an orphaned component with no page;
it is in fact live and serving on vonc.com under a page named `tool-arena`. Ask the
placement question above before concluding absence — and note that grepping the SERVED html
for the component's function name proves nothing either, because many components emit no
`data-component` attribute (that grep returns 0 on the page this one renders on).

**Gotcha:** `page_components.id` is **not stable across re-renders** — key placement
edits on `(page_id, function)`. And a page-level placement does not survive a
re-render: the rebuild resolves sections against `site_plan_sections`, so a panel
placed only in `page_components` is silently dropped **while the work item reports
`complete`**.

---

## 7. Verify the open/driven state, not the static markup (S5/S6)

```bash
python3 docs/agent_docs/docs024_key_docs_latest/brochure_component_library/scripts/probe_reveal_open_state.py
```

- **It prints the count it measured**, so "opened nothing" is distinguishable from a
  clean result. Copy that property into every new probe.
- **`render_audit.py` renders a LOCAL copy**, so `?open=<key>` never reaches
  `window.location` — a contrast run against it proves nothing about the open state.
- **Forcing `.open = true` on a DOM node is not a click.** Neither is calling the
  component's own function. Both were mistaken for verification, in two lanes, on the
  same day. Dispatch the visitor's real gesture.
- **`chrome --headless=new --screenshot` on a `file://` copy renders near-blank**
  (a few KB vs ~1MB for a live URL). For an open-state visual, use `?open=<key>`
  against the live URL — but smooth-scroll does not resolve inside
  `--virtual-time-budget`, so the shot lands scrolled to the top. Untried lever:
  force `prefers-reduced-motion` first.

---

## 8. Author a criteria fence, and prove it before publishing it

**The rule this discharges:** a fence is a tool's published contract. Until 2026-07-31 the
only way to exercise one was to write it into `doc_plans` and dispatch a cluster run — so the
first time anyone saw it run was *after* it was published. Both harnesses below import
`internal/adapters/browserrunner` and call `RunChecksAction.Execute`, so they are the **same
evaluator the fleet uses**, not a lookalike.

```bash
# 1. does the fence pass on the live page, on every profile, with nothing skipped?
go run docs/agent_docs/docs024_key_docs_latest/staged_component_build/scripts/try_fence.go \
  <fence.json> https://<domain>/<path>

# 2. can every check in it go RED? (baseline control + one mutation at a time)
go run docs/agent_docs/docs024_key_docs_latest/staged_component_build/scripts/prove_fence_can_fail.go \
  <fence.json> https://<domain>/<path>
```

**Run (2). (1) alone is not evidence** — the fence for `tool-review-council-simulator` went
36/36 green first time and one of its checks was still asserting nothing.

Gotchas, each of which cost something here:

- **`selector_count` does NOT assert a count.** `run_checks_action.go:497` passes on `n > 0`
  and `criteriaCheck` has no expected-count field — it is `selector_exists` with a friendlier
  detail line, and that detail line *reads* like an assertion. Assert counts through text the
  tool itself renders (`^26 seats on the panel$`), which also proves the tool can say what it
  did.
- **The fence must be the FIRST triple-backtick `criteria` block in the body.** Both
  extractors (`check_tool_acceptance.go:552`, `load_doc_context_action.go:143`) take the first
  one and read to the next triple-backtick. So prose that names the fence *in backticks*
  above the real one silently hijacks extraction and yields unparseable JSON. Refer to it in
  plain words in prose. Verify: the marker must appear **exactly once**.
- **Checks share ONE page per profile and run in fence order.** `evaluateOnPage` iterates the
  applicable checks against a single page instance, so interactions accumulate. Put structural
  and default-state checks first; sequence presets so a state-wiping one comes last.
  `no_console_errors` is forced last by the runner. **Reordering a fence changes what it tests.**
- **A check id ending `-EDIT` is silently skipped** as a placeholder selector
  (`splitByProfile`). Never name one that way.
- **A local copy of the page must proxy its assets** or the control is not fair: the copy
  404s `/assets/css/styles.css` and `/assets/js/snippets.js`, Chromium logs those as console
  errors, and `no_console_errors` goes red for a reason unrelated to the mutation.
  `prove_fence_can_fail.go` 302s every non-page request to the live origin.
- **`page_status_ok` cannot be falsified by any edit inside the page.** Its mutant has to be
  server-level (answer the path 404), which is why the prover has a `serveStatus` mutant kind.
- **A mutant that turns everything red proves nothing.** The prover requires `page-serves-200`
  to still pass under every body mutant, so a demolition cannot masquerade as validating the
  whole fence at once.

## 9. Write a PLAN body (or a fence) into `doc_plans` by hand

`idx_doc_plans_current` is a **partial unique index** on `(subject_type, subject_key) WHERE
is_current`, so you cannot just insert. Mirror `write_doc_plan_action.go:94-110`: supersede
then insert, **one transaction**.

```sql
BEGIN;
UPDATE doc_plans SET is_current=false, superseded_at=now(), updated_at=now()
 WHERE subject_type='tool' AND subject_key='<key>' AND is_current;
INSERT INTO doc_plans (subject_type, subject_key, body, source, created_by)
VALUES ('tool','<key>', $planbody$...$planbody$, '<lane>', 'operator:<lane>');
-- verify INSIDE the transaction, then COMMIT (or ROLLBACK for a dry run)
COMMIT;
```

- **Dollar-quote the body and generate the file from a script, never a shell string.** The
  body contains triple backticks; in a double-quoted bash string those are **command
  substitution**. Generating the `.sql` from Python and piping it avoids the shell entirely.
- **Dry-run it first with `ROLLBACK`**, per migration 270/273's precedent, with the
  verification queries *inside* the transaction.
- **Assert the stored `length(body)` equals the length you built.** That single check is also
  how you prove psql did not interpolate a `:name` inside the literal — a silent mangling that
  would otherwise reach production as the tool's contract.
- **Then read it back out and re-run the fence from the DB copy.** Writing the field is not
  reading the field: the only proof that the platform will run what you meant is to extract
  the fence *from the database* and pass that through the evaluator.

## 10. Fire an acceptance run at ONE tool, and read its result honestly

```bash
./docs/leopardessconsulting/scripts/tool_acceptance_run.sh <site_id> <domain> <function>
```

Generic despite living in that lane's directory. Its own header names the three things that
must line up; `CHECK_naming_contract.sh` checks the first two and `try_fence.go` the third.

```sql
-- state. NOTE: a FAILED run still reports status=COMPLETED, with current_step='complete_error'.
SELECT current_step, status, round(extract(epoch from (updated_at-created_at))) AS secs
  FROM orchestration_states WHERE correlation_id='<CID>';

-- the real error lives in __step_error, NOT in `error`
SELECT jsonb_pretty(collected_data->'__step_error')
  FROM orchestration_states WHERE correlation_id='<CID>';

-- the verdict
SELECT jsonb_pretty(collected_data->'request_run'->'response'->'summary')
  FROM orchestration_states WHERE correlation_id='<CID>';

-- SKIPS ARE NOT PASSES, and there are two kinds. Read the reason, never the count:
--   'not run on profile X'      = the fence's own gating. Intentional.
--   '<type> not implemented'    = the binary has no evaluator. A DEFECT — it reads as PASS
--                                 upstream and suppresses its own re-check for 7 days.
SELECT s->>'detail', count(*) FROM orchestration_states o,
       jsonb_array_elements(o.collected_data->'request_run'->'response'->'skipped') s
 WHERE o.correlation_id='<CID>' GROUP BY 1 ORDER BY 2 DESC;
```

### ⚠ THE 120-SECOND RUN DEADLINE — size the fence for the cluster, not for your laptop

`runDeadline = 120 * time.Second` covers **the entire request**: every url x every profile.
When it expires, `openChromium` returns `ctx.Err()`, so the error you get is
**`browser open failed for <url> [<profile>]: context deadline exceeded`** — which names the
browser and reads as infrastructure. **It usually means the fence is too big.**

Measured 2026-07-31, both numbers from real runs:

| | evaluations | wall clock |
|---|---|---|
| local (`try_fence.go`) | 36 | **10.6s** (x3 runs, stable) |
| cluster, v1 of the fence | 36 | **FAILED at 133s** |
| cluster, v2 (profile-gated) | 22 | **18s, 22 passed / 0 failed** |
| cluster, the only other run ever (`dc952633`) | ~21 | 48s |

So budget **~3-5s per evaluation in-cluster against ~0.3s locally**, and keep well under 120s.
The lever that costs nothing: **gate to desktop every check whose answer is
profile-independent**, and keep on mobile only what mobile can answer differently (status,
"did the JS run", horizontal overflow, console errors). That halved the run and lost no
assertion.

**And the limitation to state out loud: `try_fence.go` proves a fence is CORRECT; it cannot
prove it FITS.** It is an order of magnitude faster than the pod and does not model the
deadline. **A fence is not proven until it has completed once in the cluster.**

## 11. Rename a page so its tool becomes acceptance-testable — TWO rows, not one

The Tier-4 lookup is `name IN (function, 'tool-' || function)`, scoped by `site_id` **and
`status='active'`** (`tool_acceptance_actions.go:140-146`). A page named anything else makes
`request_browser_run` hard-error with *"no deployed page URL"*.

```sql
BEGIN;
-- Scope by ID, never by name: a concurrent rename must not be able to redirect this.
UPDATE pages           SET name = '<function>', updated_at = now()
 WHERE id = '<page id>' AND name = '<old name>';
UPDATE site_plan_pages SET name = '<function>'
 WHERE id = '<plan page id>' AND name = '<old name>';
-- verify INSIDE the transaction (below), then COMMIT — or ROLLBACK for a dry run
COMMIT;
```

**⚠ THE SECOND UPDATE IS THE ONE THAT IS EASY TO MISS AND EXPENSIVE TO OMIT.**
`check_sectionless_pages` (`discovery_checks/check_sectionless_pages.go:118`) joins
`site_plan_pages spp ON spp.name = p.name`. Renaming `pages.name` alone desynchronises that
join and the page **silently leaves that detector's population** — no error, no report, and
the page looks fine precisely because nothing is looking at it any more. Trading a naming
defect for a lost detection is the worse deal.

**Measure all of these first** (all were zero or accounted for on the arena rename, 07-31):

```sql
SELECT 'collision'      AS q, count(*) FROM pages p JOIN sites s ON s.id=p.site_id
        WHERE s.domain='<domain>' AND p.name='<function>'
UNION ALL SELECT 'sps rows on old page_name', count(*) FROM site_plan_sections WHERE page_name='<old name>'
UNION ALL SELECT 'imagery on old name',       count(*) FROM site_plan_imagery  WHERE scope_ref='<old name>';
-- and read these off the page: status must be 'active'; nav renders nav_label/title, NOT name;
-- site_plan_pages carries its own slug+url, so `name` is not the served filename.
```

**Then prove it red and green at the code's own query, not a paraphrase:**

```sql
-- BEFORE this must return 0 rows; AFTER it must return the url
SELECT COALESCE(url,'') FROM pages
 WHERE site_id=(SELECT id FROM sites WHERE domain='<domain>') AND status='active'
   AND name IN ('<function>'::text, 'tool-' || '<function>'::text)
 ORDER BY (name='<function>'::text) DESC LIMIT 1;

-- and the sectionless-check join must STILL match the page afterwards
SELECT p.name FROM pages p
  JOIN site_plans sp ON sp.site_id=p.site_id AND sp.is_current
  JOIN site_plan_pages spp ON spp.plan_id=sp.id AND spp.name=p.name
 WHERE p.site_id=(SELECT id FROM sites WHERE domain='<domain>')
   AND (p.sections IS NULL OR p.sections='[]'::jsonb) AND COALESCE(p.status,'')<>'deleted';
```

**Finally, diff the served page — and take the baseline IMMEDIATELY BEFORE the change.**
The arena page moved 31,431 → 32,553 bytes in two hours from its own lane's redeploy; a
baseline taken earlier would have attributed 1,122 bytes to a rename that changed nothing.
Worked example: `scripts/RENAME_arena_page_to_function.sql`.

**Before firing an acceptance run at another lane's tool, read the judge.** On a failing
verdict it inserts an `improve_tool` work item with `handler_agent='tool-improver'`
(`tool_acceptance_actions.go:711`) — an automated fixer. If the page is `rebuild_policy='owned'`
or the right remedy is a design decision, hand the dispatch to its owner instead of firing it.
