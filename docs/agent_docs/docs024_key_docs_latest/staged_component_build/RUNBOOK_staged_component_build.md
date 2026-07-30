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

## 3. STAGED, NOT APPLIED — the migration that unblocks P1

**Deliberately not a numbered file in `sql_for_agents/`** (PLAN D5): the runner takes
*every* pending file in a directory, so an unreviewed `272_*.sql` could be swept in by
another session's unrelated `--apply`. It gets a number when it goes to the council
gate. Follow migration **270** exactly — it is the same shape, done for `landmine`.

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

```sql
-- a page_components row alone is a RENDER ARTEFACT; site_plan_sections is the PLAN FACT
SELECT ordering, function FROM site_plan_sections
 WHERE page_id = (SELECT id FROM pages WHERE site_id=$1 AND name=$2) ORDER BY ordering;
```

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
