# RUNBOOK — bugs_open/437, the writer prompt's nested item shapes

Every command here had a gotcha attached. The gotcha is the reason the line exists.

## 1. Read the instruction the model was ACTUALLY sent

This is the whole diagnosis, and it is two minutes. Do not read the schema and infer.

```sql
SELECT id, created_at, output_tokens
  FROM llm_call_log
 WHERE agent_type = 'page-content-writer'
   AND created_at BETWEEN '2026-09-02 17:00Z' AND '2026-09-02 19:30Z'
   AND prompt_rendered LIKE '%mechanism-flow%'
 ORDER BY created_at DESC LIMIT 5;
```

Then pull the pair — prompt AND reply — from ONE row:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -At -c "SELECT prompt_rendered FROM llm_call_log WHERE id='<id>';" > prompt.txt
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -At -c "SELECT response_text  FROM llm_call_log WHERE id='<id>';" > reply.txt
grep -n "exactly these fields\|branches" prompt.txt | head
grep -o '"branches"[^,}]*' reply.txt | head -5
```

⚠ **Bound the time range.** `llm_call_log` is the training corpus, not a log — it is
large, and a bare `LIKE '%...%'` over it will run past a 120s tool timeout. My first
attempt did, *and* its `created_at >` bound was accidentally in the FUTURE, so it scanned
for nothing at length. Check `date -u` against your literal before you run it.

## 2. Census: which components can even hit this?

The blast-radius question is "does any element property declare a COLLECTION", and it must
be asked of all three live element dialects or it under-reports.

```sql
WITH f AS (
  SELECT function, key AS field_name, value AS fielddef
    FROM content_components, jsonb_each(input_schema->'fields')
   WHERE COALESCE(is_active,true) AND input_schema ? 'fields'
)
SELECT function, field_name, ikey AS item_prop, ival->>'type' AS item_prop_type
  FROM f, jsonb_each(fielddef->'items'->'properties') AS e(ikey, ival)
 WHERE fielddef->>'type' IN ('array','list')
   AND fielddef->>'source' = 'llm'
   AND jsonb_typeof(fielddef->'items'->'properties') = 'object'
   AND ival->>'type' IN ('array','object');
```

`[MEASURED 2026-09-03]` → **1 row** (mechanism-flow.steps.branches). Run the sibling
queries over the flat-`items` dialect (values are type NAMES, so test `ival #>> '{}' IN
('array','list','object')`) and over `item_schema` — both returned **0**. ⚠ A census of
only the JSON-Schema dialect would have been correct and incomplete; the flat dialect is
the majority of the library.

## 3. Rehearse a prompt-template migration BEFORE writing its guard numbers

Two rehearsals, both cheap, and the first one caught a real defect in my own guard.

**(a) String algebra, in Python, against the captured live template:**

```python
new = t.replace(A, RA).replace(B, RB)
for tok in ("{{if ", "{{end}}", "{{range ", "{{else}}", "{{.", "{{else if "):
    print(tok, new.count(tok) - t.count(tok))          # DERIVE the deltas; do not predict them
back = new.replace(TAIL, "").replace(RB, B)
print("rollback byte-exact:", back == t)
```

⚠ **Derive the expected deltas from this run, never from reading your replacement text.**
I asserted `{{if ` would be **+2** and it is **+1**: converting `{{if $f.item_fields}}`
into `{{else if $f.item_fields}}` CONSUMES one. A replacement that rewrites a token is a
deletion as well as an insertion. As written my guard would have refused my own correct
splice (`WRONG_CALLS.md`, 2026-09-03).

**(b) The real SQL, in a transaction you throw away.** Exercises the actual guards,
`snapshot_agent`, and the verify block, against the real live row:

```bash
sed 's/^COMMIT;$/ROLLBACK;/' <migration>.sql | kubectl -n ai-persona-system exec -i \
  postgres-clients-0 -- psql -U clients_user -d clients_db -v ON_ERROR_STOP=1
```

For the apply-then-rollback round trip, concatenate the migration and its `_ROLLBACK`
(strip each file's own `BEGIN;`/`COMMIT;`), wrap the pair in one `BEGIN; … ROLLBACK;`, and
compare `md5()` of the template before and after. ⚠ **psql `:'var'` interpolation does NOT
work inside a dollar-quoted `DO` block** — my first round-trip probe died on exactly that.
Select the md5 as a plain statement and compare client-side instead.

## 4. Apply it — by hand, not through the runner

```bash
./scripts/migration/run-migrations.sh --no-probe          # READ the Pending list first
```

⚠ **Never `--apply` here.** On 2026-09-03 that list held ~100 pending files belonging to
other threads. Apply your own file and record it:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 < docs/agent_docs/sql_for_agents/724_....sql
./scripts/migration/run-migrations.sh --record-only 724_....sql --note '<what you verified>'
```

`--record-only` **requires** `--note` and refuses without one; say what you verified, not
what you did.

## 5. Verify at the live row, against the declaration

Do not trust the runner's "recorded". Read the object back and count:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -c "
SELECT default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}'
FROM agent_definitions WHERE type='page-content-writer' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;" > live_after.txt
```

Then assert the four counts the `livespec` declaration
`workflow.page-content-writer.prompt_item_shape` states: nested exemplar 1, item_notes
tail 1, pre-437 spelling **0**, flat `item_fields` arm still 1. The fourth is the one
people forget — it is the machine-checkable form of "every other component's prompt is
byte-identical".

⚠ Use `#>>` on the path, not `default_config::text`: `#>>` returns the template unescaped,
so your needles are the template's own spelling rather than a JSON-escaped third copy.

## 6. Post-roll checks (the Go half)

```bash
kubectl -n ai-persona-system get pods -l app=agent-chassis \
  -o custom-columns='IMAGE:.spec.containers[0].image,AGE:.metadata.creationTimestamp'
git merge-base --is-ancestor a0044e73b <the service's build-provenance stamp>
```

Then, at the artefact — the prompt first, the page second:

```sql
SELECT prompt_rendered LIKE '%"branches": [{%'
  FROM llm_call_log WHERE agent_type='page-content-writer'
   AND prompt_rendered LIKE '%mechanism-flow%' ORDER BY created_at DESC LIMIT 1;
```

⚠ **A post-fix zero in the failure census needs a DEMAND control** — count writer runs on
mechanism-flow pages in the same window, or "no failures" is indistinguishable from "no
builds".

## 7. Council

```bash
DRY_RUN=1 ./docs/.../097_TRIGGER_council_review_v1.sh <submission.json>   # free; checks admission
```

⚠ `.plan.summary` is REQUIRED and is not in the header's field list I first worked from —
the run refuses with `ERROR: .plan.summary is empty` after you have written everything
else. Write it first. Corr for this change: `6de0f6f2-4f37-492a-9cbd-1ae886311a9b`.

⚠ **NEVER put a placeholder in a sketch, however obviously it reads as shorthand to you.**
Round 1 of this change was REVISED on exactly that: to fit the 32KB plan budget I wrote the
migration's `repl_A` as `$ra724$...anchor_A...{{if .item_notes}}…$ra724$`, meaning "the
anchor text, repeated, plus the tail". The `editquality` seat read it as the deployable
artefact — correctly, that is the rule — and reported a HIGH objection that the migration
would splice the literal string `...anchor_A...` into the live prompt. **A placeholder does
not read as an abbreviation; it reads as a defect**, and where the artefact is a migration
against a live row it reads as the worst kind. The committed file was always correct and
the applied live row proves it, but the round was still spent. If you cannot fit a sketch,
cut a different edit's prose — not the part under objection.

Resubmit on the same trail so the accumulated round history and your existing commit
trailer both keep working:

```bash
RESUBMIT_CORR=<the original corr> ./docs/.../097_TRIGGER_council_review_v1.sh <submission.json>
```

⚠ **Read the APPROVING seats' objections too.** Round 1's gating objections were all mine
in the submission rather than in the code; the only finding that changed the artefact came
from `bug_historian`, which **approved** — it spotted that the prompt's "or use `[]`"
advice rested on the empty-STRING precedent by analogy rather than measurement. The test it
prompted (`TestStructuredItemShape_EveryOmissionSpellingTheNoteRecommendsPassesTheGate`)
is now the guard against a prompt that recommends a value the gate would refuse.
