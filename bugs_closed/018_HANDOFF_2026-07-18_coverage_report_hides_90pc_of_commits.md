> **✅ CLOSED 2026-07-21 — both defects fixed & live; a third robustness gap of the
> same class was found during verification and fixed.** Closed by the
> bug-backlog-clearing thread. This is a **shell-script** fix, so "live" = committed
> (the script runs from the working tree, like config/SQL — no image roll). Moved to
> `/bugs_closed/`. The *other* `018` (idea.uk chrome, `018_HANDOFF_2026-07-19_…`) is
> unrelated and stays open — resolve by slug, not number.
>
> **Defect 1 — stdin theft — FIXED & VERIFIED.** `1493b74f3` dropped `-i` from the
> in-loop `db_decision` call. The full run *with the DB* (the path defect 1 broke)
> now returns the same in-scope count as the raw query at every window — 2026-07-21:
> 3d 62==62, 7d 111==111, 14d 150==150. Pre-fix it stopped at the first trailered
> commit (4 of 41).
>
> **Defect 2 — a fix-proposer-approved commit read as MISMATCH — RESOLVED & VERIFIED.**
> `9c56128c5` resolves the trailer against **either** `correlation_id` **or**
> `orchestration_id`; proven live — `db_decision('8c9adc27')` (a *run* id) →
> `approved|run` → REVIEWED, and `db_decision('17be3962')` (a correlation) →
> `approved|correlation` → REVIEWED. `ee5a9bed9` added the **EVIDENCE GONE** bucket:
> the handoff's own example `f32b208e5` (trailer `53da3a30` = orchestration of
> correlation `e505f70f`) has only `fix_plan` rows and no `council_report` — the
> report was *cleared* (the 091 "clear them first" practice), not mis-`kind`ed.
> `e505f70f` was genuinely council-approved (BUG A / MDL-038), so "evidence gone" is
> the honest verdict — the report no longer accuses it (the original defect) nor
> credits it without evidence. The handoff's open question ("does a fix-proposer run
> write its verdict under a different `kind`?") is answered **no**.
>
> **Defect 3 (found here) — a mangled trailer hid a false claim of review — FIXED.**
> `9c94cc842` authored a non-conforming trailer
> (`Council-Reviewed: 5a65ec4c-…-7c7fce03a779 (verdict: revise; these`); the parser
> stripped whitespace and glued the prose onto the id, so `db_decision`'s
> `tr -cd '0-9a-fA-F-'` kept hex-ish garbage, the `LIKE` prefix ran *longer* than the
> real 36-char id and matched nothing → the commit fell through to EVIDENCE GONE.
> Its real verdict is a `council_report`/`revise` — a **false claim of review**,
> silently excused, which is exactly the dishonesty this report exists to surface.
> Fix (this thread): take only the trailer's first whitespace-delimited token
> (`awk '{print $1}'`) before resolving. Post-fix `9c94cc842` correctly buckets as
> MISMATCH and the count invariant is unchanged.

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
