# HANDOFF — the council coverage report hid 90% of in-scope commits (stdin theft in a read loop)

**Created 2026-07-18 from the travelling-docs thread**, while checking whether my
own platform commits had been council-reviewed. **Defect 1 is proven and FIXED
in this commit; defect 2 is flagged, not diagnosed** — it belongs to the
council-gate owner.

**Severity: high for trust, zero for runtime.** Nothing breaks; the report simply
tells every thread that coverage is far better than it is. It reported **2**
unreviewed commits when there were **40**.

Script: `docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/098_REPORT_unreviewed_commits_v1.sh`

---

## Defect 1 — `kubectl exec -i` inside a `while read` loop eats the loop's stdin (FIXED)

**Symptom.** The report under-counts drastically and silently:

| Window | Report said | `git log --since=<w> -- platform internal pkg` |
|---|---|---|
| 2 days | 3 in-scope | 6+ |
| 3 days | **4 in-scope** | **41** |

It always stopped at the first commit carrying a `Council-Reviewed:` trailer —
which is why the cut-off moved as new commits landed on top.

**Diagnose.** The classifier loop is fed by process substitution:

```
while IFS='|' read -r sha short date subject; do ... done < <(git log ... -- platform internal pkg)
```

and for any commit WITH a trailer it calls `db_decision`, which ran
`kubectl -n "$NS" exec **-i** "$PG_POD" -- psql ... -c "<sql>"`. `-i` makes
kubectl read **stdin — which inside that loop is the git-log stream**. The first
trailered commit therefore drains the remaining commits into kubectl and the loop
ends. Everything older is invisible.

**Proof (the script's own escape hatch).** `NO_DB=1` skips `db_decision`
entirely, so the loop keeps its stdin:

```
NO_DB=1 ./098_REPORT_unreviewed_commits_v1.sh 3   →  In-scope commits found: 41
        ./098_REPORT_unreviewed_commits_v1.sh 3   →  In-scope commits found: 4
git log --since="3 days ago" --pretty='%h' -- platform internal pkg | wc -l  →  41
```

**Fix (applied).** Drop `-i` from that call — the SQL is passed with `-c`, so
stdin was never needed. Post-fix the report returns **41 in-scope**, matching the
raw query. The *other* `kubectl exec -i` in the file (the `PERSIST=1` block)
legitimately needs `-i`: it feeds SQL via a heredoc, and it runs OUTSIDE the loop.

**Transferable rule.** Any command that may read stdin — `kubectl exec -i`,
`ssh`, `psql` without `-c`, `ffmpeg` — **truncates the loop it is called from**.
Inside a `while read` loop, either close its stdin (`< /dev/null`), drop the
interactive flag, or read the loop from a dedicated FD. The failure mode is
silent early exit, not an error, so it looks like "there were only 4 commits".

## Defect 2 — the one trailered commit does not resolve to a verdict (FLAGGED, for the gate owner)

Post-fix, `f32b208e5` ("detect truncated LLM responses in GenerateText") carries
`Council-Reviewed: 53da3a30` and classifies as **MISMATCH**, not REVIEWED.

Established: that id has **3 rows in `diagnosis_artifacts`, all of
`kind='fix_plan'`; none of `kind='council_report'`** — and `db_decision` filters
`WHERE kind = 'council_report'`.

CLAUDE.md documents the trailer as accepting **either** the gate's
`SUBMISSION_CORR` **or** the orchestration id of a **fix-proposer** council run.
If a fix-proposer run records its verdict under a different `kind`, the lookup
matches both id *columns* but only one artifact *kind*, so a legitimately
reviewed commit reads as a mismatch. **Not asserted** — the owner should confirm
which kind a fix-proposer run writes before changing the filter. (Pre-fix the same
commit printed REVIEWED; I would not trust that verdict either, as it came from a
call whose stdin was polluted with git-log lines.)

## Verify after fixing

1. `./098_REPORT_unreviewed_commits_v1.sh <n>` in-scope count == `git log --since="<n> days ago" --pretty='%h' -- platform internal pkg | wc -l`.
2. A commit with a known-good trailer classifies as REVIEWED, not MISMATCH.

## References

- `CLAUDE.md` § "Council review of platform changes" (the trailer + dual-id rule).
- 016b §9 "A command that reads stdin truncates the loop that calls it".
