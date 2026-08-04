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

### 1. The lint's exit code (old plan item 3) — now genuinely unblocked

The "68 predate it" rationale expired (count is 0). The reason to wait was that flipping it
alone enforces a check satisfiable by a no-op gate (`<td>{{if .v}}{{.v}}{{end}}</td>` —
same blank, finding cleared). **That objection is now answered**: the output-level check is
live and catches exactly what the no-op leaves behind. Flipping
`check_placeholder_fallbacks.py` to exit 1 on the ungated class is defensible today. It is
still a decision, not a chore — it changes what an existing check does to other people's
work, which is the point the owner was asked about in `README_where_we_are.md`.

### 2. Watch the first UNATTENDED firing (06:55 UTC)

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
python3 scripts/check_placeholder_fallbacks.py            # CLEAN across ~176. exit 2 = flake (now retries 3× first)
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
