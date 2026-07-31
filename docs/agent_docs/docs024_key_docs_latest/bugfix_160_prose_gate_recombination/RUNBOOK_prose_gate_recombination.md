# RUNBOOK — 160, prose gate recombination

Every command here cost something to get right. Gotchas attached.

## Read the council's objections — NOT from the doc_note

The `doc_notes` row is the human-readable note, and its embedded prior-art block is
**truncated mid-sentence per line** (`…`), so the objection you most need is the one you
cannot read. The full review array is the artifact's **`body`** column — `metadata` is only a
tally (`decision`, `abstained`, `reviewers`).

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -At -c "SELECT body FROM diagnosis_artifacts
          WHERE correlation_id='<SUBMISSION_CORR>' AND kind='council_report'
          ORDER BY created_at DESC LIMIT 1;" > report.json
python3 -c "
import json; d=json.load(open('report.json'))
for r in d['reviews']:
    print('=====', r['reviewer'].upper(), '->', r['verdict'])
    for o in r.get('objections',[]):
        print('  [%s | edit %s] %s' % (o.get('severity'), o.get('edit'), o.get('problem')))
    print('  notes:', (r.get('notes') or '')[:400])
"
```

Poll the run by **payload**, never by the printed run id:

```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>'
ORDER BY created_at DESC LIMIT 2;   -- round 2 first, round 1's complete_revise below it
```

`orchestration_states` has **no `id` column** — `WHERE id='…'` errors out; that is not a
permissions or latency problem, it is the wrong column.

Round 1 took **~12 minutes** here, not the ~30 the runbook budgets.

## The mutation harness — restore-verified, or you will ship a mutant

Never mutate with the editor and undo by hand. Copy first, restore from the copy, and
**assert the restore with md5** — a mutation left in the tree is a fix that never shipped.

```bash
F=platform/orchestration/actions/verify_report_prose_action.go
cp $F $S/orig.go
sed -i 's/<the one clause>/<broken>/' $F
go test ./platform/orchestration/actions/ -run TestX 2>&1 | grep -E 'passed the SKU|^ok|build failed'
cp $S/orig.go $F
md5sum -c <(md5sum $S/orig.go | sed "s|$S/orig.go|$F|")   # must print OK
```

Grep for `build failed` in the same expression as the failure line: a mutant that breaks the
build prints `FAIL <pkg> [build failed]` with **no** `--- FAIL:` lines, which looks exactly
like a caught mutant if you only grep for `FAIL` (`LANDMINES.md`, and this tree has ~30
sessions editing the same package).

**Read the mutation result as evidence about the TEST, not only the code.** A mutation that
does not fail means the case never reached the rule. Three times in this lane; see NOTES.

## Reproduce and verify

```sql
-- the failing run: the cause is in __step_error, NOT in the work item's error column
SELECT collected_data->'__step_error' FROM orchestration_states
WHERE collected_data::text LIKE '%model-like token%';

-- blast radius: who actually consumes this action
SELECT type, is_active FROM agent_definitions
WHERE default_config::text LIKE '%verify_report_prose%' AND deleted_at IS NULL;
```

Verify the **image** before pushing it, and the **pod** after deploying — `skuTokenTraces` and
`qualifierWords` never existed in any prior build, so their presence proves the binary is new
(a roll is not evidence, `bugs_open/153`). Positive control in the same exec:

```bash
docker run --rm --entrypoint sh aqls/agent-chassis:<tag> -c \
  'strings /app/agent-chassis | grep -c "qualifierWords\|skuTokenTraces"; \
   strings /app/agent-chassis | grep -c "names model-like token"'
# 3 1 is the pass. 0 1 means the image predates the fix. 0 0 means the grep is wrong.
```

## Fixture gotchas in this package

- `realScoring(t, mass, ipMin)` — the fact block carries `required protection IP54` **only
  when `ipMin > 0`** (`buildFactBlock`, `score_grippers_action.go:737-739`). Any test case
  built around `IP54-*` is refused for an untraceable head at `ipMin 0` and never reaches the
  tail rule.
- `ipMin 54` with mass 2.5 is a **zero-match** scenario, so the no-match contract activates:
  the summary must contain `noMatchSentence` or you get an unrelated violation in the list.
  Harmless for reject tests that assert on the specific message; fatal for a `len(v)==0` test.
- `go build ./...` fails at HEAD on `cmd/reasoningset` (committed breakage, not yours). Build
  `./platform/...` and `./cmd/agent-chassis/`.
- `go vet ./platform/orchestration/actions/` reports pre-existing unreachable code in
  `load_component_library_actions.go` — not this lane's.
