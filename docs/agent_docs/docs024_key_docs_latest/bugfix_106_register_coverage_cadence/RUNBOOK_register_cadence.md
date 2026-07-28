# RUNBOOK — bug 106, register coverage cadence

## Run the sensor by hand (still supported, now not the only path)

```bash
./docs/agent_docs/docs026_concept_register/102_CHECK_register_coverage.py          # drift vs ratchet
./docs/agent_docs/docs026_concept_register/102_CHECK_register_coverage.py --all    # every uncovered subsystem
./docs/agent_docs/docs026_concept_register/102_CHECK_register_coverage.py --strict # exit 1 on NEW drift (CI)
```

## Measure a new pattern-check's fire rate (the file's own bar for inclusion)

`scripts/pattern-check.py` requires a new check to state its rate against real
history before it ships. This is the loop:

```bash
fires=0; n=0
for c in $(git log --format=%h -n 1500); do
  n=$((n+1))
  out=$(timeout 25 ./scripts/pattern-check.py --commit "$c" 2>/dev/null | grep -c "<your-kind>" || true)
  [ "${out:-0}" -gt 0 ] && fires=$((fires+1))
done
python3 -c "print(f'{100*$fires/$n:.2f}%')"
```

Gotchas: `--commit <sha>` audits that commit **in isolation** (`sha~1..sha`), which
is what you want; `|| true` on the `grep -c` is load-bearing because grep exits 1
on no match and the loop runs under an implicit failure-sensitive shell; and 1,500
commits is ~2 minutes, so run it once and record the number rather than re-deriving.

**Rate alone is not the bar.** A very low rate can mean high precision or a dead
check, and the two look identical. Inspect every fire.

## Induce the gap (the verification 106 demands)

A green report on a register somebody just hand-patched proves only that the patch
happened. Prove the detector *can* fire, and that both silencing routes work:

```bash
D=docs/agent_docs/docs024_key_docs_latest
mkdir -p "$D/zz_probe_a" "$D/zz_probe_b"
echo scratch > "$D/zz_probe_a/PLAN_probe.md"; echo scratch > "$D/zz_probe_b/PLAN_probe.md"
git add "$D/zz_probe_a/PLAN_probe.md" "$D/zz_probe_b/PLAN_probe.md"

./scripts/pattern-check.py            # ARM 1 — both must fire

echo zz_probe_a >> docs/agent_docs/docs026_concept_register/102_coverage_ratchet.txt
./scripts/pattern-check.py            # ARM 2 — only B fires

# restore the ratchet, then give B a register entry instead
./scripts/pattern-check.py            # ARM 3 — B quiet, A fires

git restore --staged "$D/zz_probe_a/PLAN_probe.md" "$D/zz_probe_b/PLAN_probe.md"
rm -rf "$D/zz_probe_a" "$D/zz_probe_b"
```

**Back up the ratchet first** (`cp` it to your scratchpad) — arm 2 edits a real,
shared file, and another session may commit while your probe is mid-flight. Restore
it before you finish, and `git status` to prove no residue.

## Why the check imports the sensor instead of copying `is_covered()`

One matching rule, one implementation. Two hand-maintained copies is the
`idx_swi_dedup` ↔ `workItemTerminalStatuses` drift class. If you change how
coverage is decided, change it in `102_CHECK_register_coverage.py` and the commit
hook follows automatically.

The import is `importlib.util.spec_from_file_location` because the sensor lives
under `docs/` and is not an importable package. It is guarded: if the sensor is
moved or deleted the check returns silently rather than breaking commits.
