# HANDOFF 2026-08-04 — bugfix 140 / RFC_009 · **all eight plan items discharged; the element-level check is STANDING**

**Read this first.** This supersedes `HANDOFF_2026-08-03_continue_here.md`. That doc set
out eight items; **all eight are done**, and item 1's check is deployed, running daily and
proven in the cluster. Nothing is half-finished. Every commit is narrow, everything that
ships is proven at the artefact, and the two non-blocking follow-ons are named below.

History: `NOTES_…` (technical, with this session's three missteps),
`README_where_we_are.md` (plain prose, the owner's), `RUNBOOK_…` §R11 (the new commands),
the `SUMMARY_…` series.

---

## One-line state

> **SUPERSEDING UPDATE, 2026-08-04 19:55 UTC — follow-on 1 is DONE and only follow-on 2
> remains.** The owner ruled "flip it": the ungated `skip_field` class now **exits 1**,
> live and proven in-cluster (commit `2b1684314`). RFC_009 B+C re-proven again on
> **`v1.0.1251`** — the `v1.0.1250` numbers below were superseded within hours, as this
> doc warned they would be. The single remaining item is watching the 06:55 UTC firing on
> **2026-08-05**; everything else on this page is discharged.

`bugs_closed/140` CLOSED. RFC_009 B and C live. The 68 `skip_field` fields gated
(migration 295). **The element-level blindness the 08-03 recheck exposed is now covered by
a STANDING check**: `component-render-check` runs daily at 06:55 UTC as a CronJob
(CGV-030), image `v1.0.1250`, baseline `go:embed`-ed into the binary, proven in-cluster —
manual Job Succeeded, `doc_notes` row verified by query, not inferred. **All eight plan
items from the 08-03 handoff are discharged.**

RFC_009 B re-proven on the chassis **three times** during this session, because it kept
rolling under me: `v1.0.1246`, then `v1.0.1247`, then **`v1.0.1250`** (both replicas each
time) — compiled markers `a component must not assert…` 2, `template invents` 1,
`library_fabricated_hours` 1 (C), negative control 0, on every replica. Markers chosen with
`scripts/pick-pod-marker.py` rather than by hand, which is the point of having built it.
*Each re-run exists because the previous proof named a tag that was superseded within the
hour: a proof carries the tag it was taken on, and a rolled fleet retires it.* **The live
tag when this was written is `v1.0.1250` — check `kubectl get deploy agent-chassis -o
jsonpath='{...image}'` before quoting any of these numbers.*

---

## What was done, with its commit

| item | what | commit |
|---|---|---|
| 2 | `.githooks/commit-msg` **blocks** a non-UUID `Council-Submitted:`/`Council-Reviewed:` value. An absent trailer still passes — review stays advisory | `9cfab6f4d` |
| 6 | lint `load()` retries the truncating fetch 3× — and a **second flake shape** was found: rc=0 + cut body used to exit **1**, the findings code | `6f0808aea` |
| 5 | `scripts/pick-pod-marker.py` — markers proven against a binary built from `git archive`; comment-only literals shown as the trap | `eadf3a77e`, DOC-073 |
| 7 | `bugs_open/190` — the two undecoded-envelope rows, with per-row repairability MEASURED through the live parser | `ba91d07a4` |
| 4 | queue read + contributed into the finetuning lane; **no work fired** | `03e485d55` |
| 8 | the config-migration review gap put to the owner in prose | `cf7879e85` |
| 1 | `cmd/component-render-check` — the check itself | `c6ae2e300`, `bb8be9cd1`, `c77795cf9` |
| 1 | its baseline + mutation-proven growth detection (and the blind-pass defect a mutation found) | `d0b44e6b1`, `fd872bc91` |
| 1 | the CARRIER: direct-Postgres path, image, daily CronJob, make targets, `go:embed`-ed baseline | `0971af4f6`, `7a150dc7e`, `83755f449`, `1194e42f0`, CGV-030 |
| — | paper trail: LANDMINE (synced) + WRONG_CALLS row for the `strings`/non-ASCII trap | `b2e2b2953` |

Three findings from doing it that change what the next thread should believe:

1. **The blog rebuild did not stall — it was REFUSED, protectively.** `needs_page`
   `d96aee06` re-confirmed 1 of 3 stored sections (33% < `prune_floor_ratio` 0.50), so the
   whole save was refused and nothing was written. The blank hero survives *because* the
   guard worked. Do not "retry" it without deciding whether the shrinkage is genuine.
2. **finetuning's `insights`/`ai-guides` empty_sections are downstream of a missing
   COMPONENT**, not of missing content: `article-grid`/`category-section` generation has
   failed 3× at `store_component` on "template variables and schema fields do not match"
   (the 287 desync class). Nothing about those items will resolve until that does.
3. **Half of `bugs_open/190` is mechanically repairable and half is not**, and only
   running the parser tells you which — measured, in the bug file.

---

## What is OWED

**Nothing from the 08-03 plan.** Two follow-ons, neither blocking:

### 1. ~~The lint's exit code (old plan item 3)~~ — **DONE 2026-08-04 evening, owner ruled "flip it"**

> **UPDATE 2026-08-04, 19:55 UTC.** No longer owed. The owner was asked and chose to
> enforce; `check_placeholder_fallbacks.py` now exits **1** on the ungated `skip_field`
> class as well as on a fabricated fact. Live: commit `2b1684314`, ConfigMap
> `component-fallback-check-script-7862bfkf9t` applied, manual `--from=cronjob` run
> **Completed with exit code 0 read off the pod**, `doc_notes` row 19:54:35Z. Clean at the
> flip — 178 components — so the ratchet closed at zero and nothing is red.
>
> **Proven by mutation, and it had to be:** with the population at 0 an honest run exits 0
> whichever version ships. Three fixtures through the pre-change script (pinned
> `6f0808aea`) and the new one — clean `0→0`, **ungated `0→1`**, fabricated `1→1`.
> Reproduce it from RUNBOOK §R9, which now also carries the apply command and the
> read-the-exit-code-off-the-POD form.
>
> **Correction that mattered, made before the decision rather than after:** this lint
> never gated a build and still doesn't. It is not in the pre-commit hook and no build
> runs it. Flipping changed the daily Job's status and a session's exit code, nothing
> else. Any doc telling you "ungated-only exits 0" predates 08-04 and is wrong.
>
> ⚠ **The known consequence, written down where it will be met:** a lane whose job goes
> red can clear the finding with a no-op gate (`{{if .v}}{{.v}}{{end}}` in a fixed cell)
> and leave the identical blank. `component-render-check` is the counter — run
> `/tmp/rck --compare` before and after any such repair. LANDMINES, footprint
> `check_placeholder_fallbacks.py`.

### 2. Watch the first UNATTENDED firing (06:55 UTC) — **STILL OWED; it is the ONLY thing owed**

> **19:55 UTC 08-04: not checkable yet, and `LAST SCHEDULE <none>` is CORRECT, not a
> fault.** The CronJob was created ~11:30 UTC, after that morning's 06:55 slot, so its
> first unattended firing is **06:55 UTC on 2026-08-05**. Do not read the empty
> LASTSCHEDULE as a broken schedule before then — and do not "fix" it.
> (Verified this evening: the *sibling* `component-fallback-check` shows LASTSCHEDULE 12h
> and its 06:40 run wrote a clean row, so the CronJob substrate itself is working.)

The manual Job proves the image, the query, the credentials and the write. It does **not**
prove the schedule. Check tomorrow:

```bash
kubectl -n ai-persona-system get cronjob component-render-check     # LASTSCHEDULE set?
kubectl -n ai-persona-system get pods -l app=component-render-check # pod Completed, not BackOff
```
```sql
SELECT created_at, left(body,120) FROM doc_notes
WHERE source='component_render_check' ORDER BY created_at DESC LIMIT 3;
```

⚠ **Read the POD, not the Job.** `ImagePullBackOff` leaves the Job reporting `Running`
with `0/1` for its whole deadline — it is never `Failed`, so a dead check looks like a slow
one. That is exactly how this nearly shipped broken (LANDMINES, footprint
`imagePullSecrets`).

### Banking a fix, when someone fixes a real hole

The baseline is embedded, so it is deliberately not editable in place:

```bash
go build -o /tmp/rck ./cmd/component-render-check/
/tmp/rck --write-baseline cmd/component-render-check/baseline.json   # regenerate
git commit cmd/component-render-check/baseline.json -m "..."         # the diff IS the record
make build-component-render-check push-component-render-check deploy-component-render-check
```

There is no ConfigMap to patch and no file in the image to edit — silencing a finding costs
a reviewable commit, by construction.

## Explicitly NOT owed

- **No pod-grep outstanding for RFC_009.** Re-proven on `v1.0.1250`, both replicas, B and C
  markers + negative control. To re-prove after the next roll, use
  `scripts/pick-pod-marker.py 87ea0a5e7` rather than picking a marker by hand.
- **No rerenders owed** (item 4 stands: the residual rows route through existing queue
  items, and one of them is protectively refused).
- **No council verdict outstanding.** Nothing this session touched a shared seam: two new
  files under `cmd/`/`scripts/`, one hook, one allow-list entry, one Python retry. The
  config-migration review gap is item 8 and is the owner's.
- **Nothing about the carrier.** Image built from committed HEAD, pushed, CronJob
  deployed and run; makefile has `build-`/`push-`/`deploy-component-render-check` and
  `component-render-check-now`.
- **`bugs_open/190` is filed, not owned.** Its fix candidate 1 (guard the `content_data`
  write seam) needs a writer census first, which is stated in the file as `[UNMEASURED]`.

---

## Health checks

```bash
python3 scripts/check_placeholder_fallbacks.py            # CLEAN across 178 (08-04 19:30). exit 1 = EITHER class since 08-04; exit 2 = flake (retries 3× first)
python3 scripts/check_placeholder_fallbacks.py --selftest # 10 must-refuse / 14 must-allow
go test ./platform/orchestration/actions/ -run TestFabricatedFallback
go build -o /tmp/rck ./cmd/component-render-check/ && /tmp/rck --compare  # expect 0 NEW, 0 UNCOVERED, exit 0
kubectl get cronjob component-fallback-check component-render-check -n ai-persona-system   # LASTSCHEDULE today
kubectl -n ai-persona-system get pods -l app=component-render-check        # Completed — read the POD, not the Job
```

```sql
-- bugs_open/190's closing condition:
SELECT count(*) FROM page_components
WHERE content_data ? 'type' AND content_data ? 'result' AND content_data->>'type'='text';  -- 2 today
```

## The five things worth knowing before you touch any of it

1. **`data-runtime-fill` must be honoured by anything that flags emptiness** — vonc's
   shells are `check_empty_sections.go`'s documented first false positive.
   `component-render-check` demotes them to info at COMPONENT granularity (its input is one
   template, never a page — `bugs_open/137`'s failure shape needs a page-shaped input).
2. **The positive control is per FIELD.** A field whose marker never reaches the baseline
   (condition-only, attribute-only, `{{else}}`-branch) is UNCHECKED, not clean, and the
   tool says so in its own section. 30 fields sit there.
3. **Baseline-subtract or blame the wrong field**: ~44 components are blank at full data.
4. **`strings` splits at non-ASCII bytes** — a pod-grep marker containing `…` returns 0
   against a binary that contains it, and a negative control cannot catch it (LANDMINES,
   footprint `strings /app/`).
5. **The lint's real file is under `deployments/kustomize/.../base/`**; `scripts/…` is a
   symlink (kustomize refuses a source outside its root). Edit through the real path — the
   Write tool refuses the symlink, which is how you find out.
