# RUNBOOK — bugfix 163

Every command that was hard to get right, with its gotcha attached. Fix it HERE when it changes.

## Is this bug still real? (the disconfirmable test)

Retention makes the log-based version of this claim unreproducible. Run the predicate the Go
**builds**, not the one you would write — that was the check that cost the filing lane two
rounds (`WRONG_CALLS.md:13881`):

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT 'path-bearing (as executed)' AS form, count(*) FROM code_symbols
 WHERE symbol ILIKE '%internal%' AND symbol ILIKE '%analysis%' AND symbol ILIKE '%symbolbody%'
   AND symbol ILIKE '%go%' AND symbol ILIKE '%ReadSymbolBody%'
UNION ALL SELECT 'bare name', count(*) FROM code_symbols WHERE symbol ILIKE '%ReadSymbolBody%'
UNION ALL SELECT 'correct split', count(*) FROM code_symbols
 WHERE path ILIKE '%internal/analysis/symbolbody.go%' AND symbol ILIKE '%ReadSymbolBody%';"
```
Expect `0 / 1 / 1` before the fix.

**GOTCHA — always select the denominator in the same query as the finding** (181's lesson):
```sql
SELECT count(*) AS n, count(*) FILTER (WHERE symbol LIKE '%/%') AS symbol_has_slash FROM code_symbols;
```
`0` is only meaningful beside its `4992`.

## Who owns a bug, really

`scripts/who-owns.py <n>` reads **commits**, so a session mid-fix is invisible. Add the
transcript pass — and **grep for the FIX SITE'S SYMBOL, not the mechanism's name**:

```bash
cd ~/.claude/projects/-home-ant-projects-agentchassis/
for f in $(find . -name "*.jsonl" -mmin -300); do
  c=$(grep -c "symbolTokenClause" "$f"); [ "$c" -gt 0 ] && echo "$(basename $f|cut -c1-8) [$(date -r $f +%H:%M)]: $c"
done
```
**GOTCHA:** counting `bugs_open/NNN` is nearly useless — every session that runs
`ls bugs_open/` picks up all 60 numbers. And counting the *mechanism* name ("landmine-verifier")
over-reports: every adjacent lane says it. On 08-02 that exact confusion parked this bug for
two days. The function only a fixer would open is the discriminator.

## The code index — read this before you believe a 0-row result

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT commit_sha, count(*) AS symbols, max(updated_at) AS refreshed FROM code_symbols GROUP BY 1;"
```

**GOTCHA, and it will make a correct fix look broken:** the index is pinned at
`d98010e8bc9e…` (2026-07-28), single-commit, 4,992 rows — roughly 1,600 commits behind HEAD.
**Any symbol you added today is legitimately absent from it.** Verify this fix with a symbol
that exists AT THAT COMMIT. `symbolTokenClause`, `identifierTokens` and `answerCodeCheck` all
do — check before you use one:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT path, symbol, kind FROM code_symbols WHERE symbol IN ('symbolTokenClause','ReadSymbolBody');"
```

## Firing the landmine verifier

```bash
./scripts/landmines-verify-dispatch.sh          # for new/changed entries
./scripts/trigger-landmine-verifier.sh '<doc_notes.source slug>'   # for one entry, any time
```
**GOTCHA — do NOT run `landmines-sync.py --apply` first.** The `NEEDS_VERIFICATION` signal is
a diff between the markdown and the DB, and `--apply` overwrites the DB with the markdown, so
it **consumes the signal**. The wrapper then exits 0 saying "Nothing needs verification" and
fires nothing. CLAUDE.md names only the inner script, which is how this trap gets sprung
(`WRONG_CALLS.md:13574`). Use the dispatcher *instead of* the sync.

Read the verdicts back:
```sql
SELECT created_at, subject_key, body FROM doc_notes
 WHERE categories ? 'landmine-verification' ORDER BY created_at DESC LIMIT 5;
```

## Guards that fire on this file

```bash
python3 scripts/pattern-check.py
```
`DECLARED_PAIRS` couples `FROM code_symbols` ↔ `codeIndexFreshness` (`scripts/pattern-check.py:288`):
a renderer reading the index must carry the freshness fact. Adding a second `FROM code_symbols`
to this file is fine — the pair is satisfied in-file — but **run it, do not assume it**.

## Proving the deploy

```bash
kubectl get pods -n ai-persona-system -l app=agent-chassis \
  -o custom-columns=NAME:.metadata.name,IMAGE:.spec.containers[0].image --no-headers
kubectl exec -n ai-persona-system <pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "<string my change ADDED>"'
```
**GOTCHA:** three probes, every replica — a string the change **added**, a **positive control**
that pre-existed (proves the pipeline and your grep), and a **negative control** that must
return `0`. A roll is not evidence your fix shipped (`bugs_open/153`).
