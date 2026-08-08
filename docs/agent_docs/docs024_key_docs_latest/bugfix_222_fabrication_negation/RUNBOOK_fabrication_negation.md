# RUNBOOK — bugfix 222 (fabrication gate negation-blindness)

Commands that were hard to get right, with the gotcha attached. Update this file
the moment a command changes, not later.

## Build/test loop for this fix

```bash
go build ./platform/orchestration/actions/... ./platform/orchestration/datahelpers/...
go test ./platform/orchestration/actions/... ./platform/orchestration/datahelpers/...
```

**Gotcha:** the shared tree carries other sessions' uncommitted files in the SAME
packages. Two failures show up unconditionally
(`TestValidDocSubjectTypes_LockstepWithMigrationCheck`,
`TestEveryCheckProducedItemTypeIsClassified`) — confirmed pre-existing at clean
HEAD (see below), unrelated to this fix. Don't chase them; don't let them block
this task's own green bar.

## Testing against a clean HEAD when the shared tree is dirty

```bash
mkdir -p "$SCRATCH/clean_head_222"
git archive HEAD | tar -x -C "$SCRATCH/clean_head_222"
cp platform/orchestration/datahelpers/claims.go "$SCRATCH/clean_head_222/platform/orchestration/datahelpers/claims.go"
cp platform/orchestration/datahelpers/claims_test.go "$SCRATCH/clean_head_222/platform/orchestration/datahelpers/claims_test.go"
cp platform/orchestration/actions/check_tool_fabrication_action.go "$SCRATCH/clean_head_222/platform/orchestration/actions/check_tool_fabrication_action.go"
cp platform/orchestration/actions/check_tool_fabrication_action_test.go "$SCRATCH/clean_head_222/platform/orchestration/actions/check_tool_fabrication_action_test.go"
cd "$SCRATCH/clean_head_222" && go build ./... && go test ./platform/orchestration/actions/... ./platform/orchestration/datahelpers/...
```

This isolates "does MY change work" from "is the shared tree currently compilable
end to end" — the two pre-existing failures reproduce on bare `git archive HEAD`
with none of my files copied in yet, which is the proof they are not mine.

## Reproduction-first (write the failing test before the fix)

```bash
go test ./platform/orchestration/actions/... -run 'TestDetect_DeniedFabricationComment_NotGated|TestDetect_RealDeclarationWithUnrelatedNegatorElsewhere_StillGated|TestDetect_NegatorInsideMatchSpan_NotGated|TestDetect_DenialVocabularySweep_NotGated|TestPostPositionedDenialIsAKnownResidual' -v
```

**Gotcha, twice, both caught by running this BEFORE trusting the fixture:** a test
row that doesn't fail red against unfixed code is not proof of anything — it may
mean the fixture never matched any regex at all (`fabDataNearQualifier`'s
data-noun group has no bare `\bdata\b`, only `dataset`/`records`/…), or it may
convict via an entirely different, unrelated match than the one the row was meant
to exercise (`seeded` unanchored inside "pre-seeded"). Check every row's *reason*
for failing, not just that it failed.

## Mutation-and-control triple (T-6)

Run against the clean-archive copy above, one file at a time, reverting fully
before the next:

```bash
FILE="$SCRATCH/clean_head_222/platform/orchestration/datahelpers/claims.go"
# Mutation A — guard never fires:
python3 - "$FILE" <<'PY'
import sys; p=sys.argv[1]; s=open(p).read()
old="func (g NegationGuard) NegatedAt(text string, pos int) bool {\n\twindow := text[maxInt(0, pos-g.Window):pos]"
new="func (g NegationGuard) NegatedAt(text string, pos int) bool {\n\treturn false // MUTATION A\n\twindow := text[maxInt(0, pos-g.Window):pos]"
open(p,"w").write(s.replace(old,new,1))
PY
# ... go test, then restore FILE from the real working tree, repeat with `return true` for Mutation B.
```

**Gotcha:** a bare `sed -i 's/func (g NegationGuard) NegatedAt/XXX_UNUSED/'`
mangles the signature line and produces a syntax error, not a clean mutation —
match the FULL line-pair (signature + first body line) so the replacement stays
syntactically valid Go. Restore by copying the real file back from the working
tree (`cp platform/orchestration/datahelpers/claims.go "$CLEAN/..."`), not by
trying to un-sed.

Control mutation (proves the new tests aren't vacuously sensitive):
```bash
sed -i 's/const fabLiteralRecordThreshold = 15/const fabLiteralRecordThreshold = 14/' \
  "$SCRATCH/clean_head_222/platform/orchestration/actions/check_tool_fabrication_action.go"
# go test — everything must still PASS. Then revert the sed (15 for 14).
```

## Council submission

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh \
  docs/agent_docs/docs024_key_docs_latest/bugfix_222_fabrication_negation/submission_222_r1.json
```
Save the printed `SUBMISSION_CORR`. Budget ~30 min; find the run by payload, not
by the printed id:
```sql
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
```
