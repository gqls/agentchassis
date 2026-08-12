# dispatch/pool lane — HANDOFF 2026-08-12b: D2 is APPROVED and live, and its post-roll check found the metric was scraped from neither chassis pod

**Cold-start for a new chat. Read this, then `bugs_open/259_HANDOFF_2026-08-12_three_processor_paths…`.**
The previous handoff in this directory (`HANDOFF_2026-08-12_continue_here.md`) is history —
its two owed items are both discharged below. **One of its stated facts was wrong and is
corrected in §2; do not carry it forward.**

## 1. Both owed items are discharged

**(a) D2's council verdict: APPROVED.** Correlation `e3aa14c5-adcd-4472-b0ee-213ae043e378`,
decided 2026-08-12 17:44:34Z. Nothing was owed on the code, so the `Council-Submitted:`
trailer on `11abe7a41` will be credited automatically by `098` at report time. No REVISE,
no REJECTED, no action.

**(b) The post-roll check ran, and it came out disconfirming.** D2 shipped in **v1.0.1293**
and the Go half is provably correct — `[MEASURED 2026-08-12]` both chassis pods serve all
nine `go_sql_*` series on `:9090`, reading
`go_sql_max_open_connections{db_name="clients_db"} 12`, matching `CHASSIS_DB_MAX_OPEN_CONNS=12`.
**But Prometheus was scraping neither of them**, and had never scraped either.

## 2. ⚠ CORRECTION — the previous handoff's reader-check was wrong, in the direction that reads as success

It said, and SYS-091 said with it:

> PodMonitor … selects `app in (agent-chassis, dynamic-agent)`, path `/metrics`,
> **`targetPort: 9090` numeric while the pod declares only containerPort 8080** — that
> numeric target is what makes the scrape work, so do not "tidy" it. Prometheus reports both
> chassis pods **`health:"up"`**…

The selector and endpoint are stated correctly. **Both conclusions drawn from them are
false.**

- The operator compiles a numeric `targetPort` into
  `keep __meta_kubernetes_pod_container_port_number == "9090"` (read verbatim from
  `/etc/prometheus/config_out` on the live Prometheus). That arm matches the port a pod
  **DECLARES**, never the port it serves. So the numeric target is not what made the scrape
  work — it is precisely what dropped the chassis.
- `[MEASURED]` **0 of 141** active targets, across every scrape pool, was an
  `agent-chassis-*` pod. The pods read as `up` because the `job` label is the **PodMonitor's
  name**: all ~108 spawned-agent targets carry `job="ai-persona-system/agent-chassis"` and
  look like the chassis at a glance.

**Fixed** in `889a7c055`: `base/deployment.yaml` now declares `9090 name: metrics`, matching
what `spawn_actions.go` already does for spawned pods. **Not** switched to a port name —
that was the original author's rejected alternative and fails from the other side. SYS-091,
`podmonitor.yaml`'s comment, `LANDMINES.md` and `WRONG_CALLS.md` are all updated;
`bugs_open/040` has the contribution (that lane owns the file — do not compete on it).

**The lesson, because it survived a council round, a mutation test and a deliberate
reader-check:** an aggregate over a superset is not evidence about a member. The query
`go_sql_max_open_connections` returned 108 healthy series and zero rows from the two pods
the instrument existed for. **Filter by the thing you mean** (`{pod=~"agent-chassis-.*"}`),
and ask the **target list for the pod name**, never the metric for a health status.

## 3. THE ONE THING OWED NOW

**After the next roll, assert D2 at the artefact.** The manifest change is inert until a
roll applies it (owner runs `make release`; whole-fleet).

```bash
kubectl -n monitoring exec prometheus-kube-prometheus-stack-prometheus-0 -c prometheus -- \
  wget -qO- 'http://localhost:9090/api/v1/query?query=go_sql_max_open_connections%7Bpod%3D~%22agent-chassis-.*%22%7D'
```

Expect **two series reading `12`**, one per chassis pod. **An empty vector is the
disconfirming result and is what this returns today** — which is the only reason the check
is worth running. Do not substitute the unfiltered query: it is healthy right now and says
nothing about the chassis.

If it is still empty after a roll that carries `889a7c055`, the next suspect is the
`allow-monitoring` NetworkPolicy (040's layer 3) — the chassis pod IP is reachable from the
`monitoring` namespace today, so it is not currently implicated, but that was checked
against a pod Prometheus was not scraping.

## 4. What remains on the lane

1. **`bugs_open/259` (slug `three_processor_paths…`) — site A is next and untouched.**
   B is proven dead code and C is redundant; both are `delete`, and candidate 1 (delete
   `p.sqlDB` entirely) takes them together. **A must be assessed first** — removing the
   handle turns A's guard into an ordinary `p.db` read, a live-behaviour change. The open
   question on A (`process()`, genuinely reachable) is whether the dead early-return causes
   a duplicate parent response or is belt-and-braces over a guard elsewhere; it needs a live
   behavioural measurement. ⚠ Do **not** mechanically apply 239's fallback to B — that
   resurrects a dead path.
2. **Memory-index compaction** (was TASK 3): the hook fires at 92% of the 25,000B cap; the
   sanctioned exit is moving closed-and-live bug entries to `MEMORY_closed.md`. Count is the
   binding axis, not bytes.
3. **Optional, and it is 040's call not ours:** `podmonitor.yaml` is live but **not in the
   kustomize build** (`base/kustomization.yaml` lists only `deployment.yaml`) — hand-applied,
   reconciled by nothing, drift silent in both directions. Recorded in the 040 contribution.

## 5. Standing facts and traps for this lane

- **Prove a deploy** at the artefact. The `build provenance` log line **scrolls within the
  hour** on the chassis (absent from a 100k-line tail on pods 1.5h old — that is "out of
  range", not "unstamped"). A sha probe of `/proc/1/exe` needs a **narrow** candidate list:
  the fleet is committing fast enough that the 40 most recent commits spanned ~28 minutes on
  2026-08-12, so a 40-sha window can miss a build from two hours ago entirely, and a
  400-pattern `grep -f` against the binary times out.
- **kcat -P sends one message per LINE** — single-line `<<<` envelopes only.
- **A sha is generated output, never retyped** (`git rev-parse`, or pasted from `git log`).
- Peer coordination: the lane on the "bugfix 238" socket is the shared-pool-ownership lane
  (`bugfix_246_shared_pool_ownership/`), NOT the 238 lane.
- Council: one run per coherent task; `Council-Submitted:` when committing before the
  verdict; a scope veto is answered by recording, not resubmitting. Deployment manifests and
  docs are **out of council scope** (it takes `platform/`, `internal/`, `pkg/`), which is why
  `889a7c055` carries no trailer.

## 6. Where everything is written down

- `bugs_closed/239_…` — the whole 239 story. `bugs_closed/246` — pool ownership.
- `architecture_review/RFC_023_…` — the behaviour-not-package-count scope ruling.
- Register **SYS-090** (dispatch seam) and **SYS-091** (pool instrumentation, status
  corrected today).
- `LANDMINES.md` — "a PodMonitor's numeric `targetPort` keys on the port a pod DECLARES".
- `WRONG_CALLS.md` — 2026-08-12, "the reader was checked first".
- `bugs_open/040` — the metrics-plumbing bug this rides on, with today's contribution at the
  foot.
