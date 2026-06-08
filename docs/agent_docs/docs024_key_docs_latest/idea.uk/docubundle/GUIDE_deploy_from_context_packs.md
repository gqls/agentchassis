# Guide — working and deploying the projects from their context packs

How to take a context pack into a fresh thread, gather what it lists, do the work, and ship it. One general loop, then the deploy mechanics (which differ per project), then a per-project quick reference.

> Division of labour: **you** run all SQL, kubectl, builds, and deploys; the assistant reads and reasons. Where exact commands depend on your env, this points at the authoritative file rather than inventing them.

---

## The general loop (same for every pack)

1. **Open the fresh thread** and attach the pack's listed docs and code. Start the thread by stating the task and the precise next action from the pack.
2. **Gather the live context** the pack lists — schema (`dbcontext -schema …`), live rows (`dbcontext -rows …`), runtime (`kubectl …`), and any **fresh-from-system code** the pack flags as not-on-disk. Paste these in.
3. **Verify the decisive fact before acting.** Every pack names a fork or a fact to confirm first (e.g. "are the section rows actually read?", "does `CurrentStep` hold the expanded loop name?", "is `service.go` ~658 lines and building?"). Confirm it before prescribing — the packs are a restatement of earlier context and inherit its staleness, so treat the fresh pull as the source of truth.
4. **Do the work** within the standing rules (snapshot before DB change, reuse before recreate, schema before SQL, no `logger.Debug`, don't rename vars, structural over patches).
5. **Deploy** via the mechanism(s) for that project (below).
6. **Verify the deploy** (rollout healthy / orchestration progressed / work-item completed with positive evidence / email arrived).

---

## Deploy mechanisms

Most projects use more than one. Know which target you're touching — the **chassis platform** (the Go agent system, a container image in k8s) is a different thing from the **sites it generates** (static output shipped to Backblaze).

### A. Chassis platform image (Go action/agent code changes)
Any change to chassis Go code (actions, agents, dispatches) compiles into the chassis image and only takes effect once the image is rebuilt and rolled out.
- Shape: **build & push the image → bump the image tag → apply/roll out the k8s deployment.** Use the targets in your `makefile.txt` and the tag in `kustomization.yaml`/`deployment.yaml` — confirm the exact `make` targets there rather than assuming.
- Bump the tag (e.g. from `v1.0.1057`) so the rollout actually pulls the new image; a reused tag may not.
- Roll out and watch: `kubectl -n ai-persona-system rollout status deploy/<chassis-deploy>` and `kubectl -n ai-persona-system get pods` (new pods Running, no crash-loop).
- Gotcha (from thunder): editing a script without re-packaging produces a byte-identical artifact (same md5) — the change won't ship. Re-tar/rebuild so the artifact actually changes.

### B. Database (SQL and numbered migrations)
- Apply via psql, through your access pattern: `kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -f <file.sql>` (or `-c "…"`).
- **Snapshot first** (`CREATE TABLE … _bak_<tag> AS SELECT …`; short names; remember `CREATE TABLE IF NOT EXISTS … AS SELECT` is a no-op on re-run, so don't trust a re-run's printed counts).
- Resolve ids fresh in the SQL (e.g. `site_id` changes every teardown → `(SELECT id FROM sites WHERE domain='…')`).
- Verify by re-querying, not by assuming the write took.

### C. Work-items (trigger a pipeline)
- Insert a `site_work_items` row; `build-dispatch-loop` claims it (`triaged → claimed → complete`). Shapes: re-render existing components → `item_type='needs_page'`; generate content → `item_type='needs_content_page'`; both `handler_agent='page-build-handler'`, `status='triaged'`, `ON CONFLICT DO NOTHING`. (Exact spec JSON is in the adoption pack / the gamesdesign handoff.)
- Verify: the item's status and **positive evidence** (the page actually has components / is deployed), not just "completed" — "complete" can mean "we stopped", so check the artifact.

### D. Orchestration trigger (start/re-run an orchestrator)
- Produce an `orchestrate` message to `system.agent.generic.requests` via kcat. The pattern is in your trigger scripts (`initial_messages…`, `initial_vet_practice_check_message`): `kubectl -n kafka run … kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 -t system.agent.generic.requests -H … <<JSON {"action":"orchestrate","config":{"agent_type":"<orchestrator>"},"input_data":{…}} JSON`.
- For thunder, the trigger targets `model-trainer` with a known-good export id — the exact JSON is in `RUNBOOK_phase_b_c_d_deploy.md` §4b; re-run with export `146a9a12` or `fef7be6b` (not `a8484922`).
- Verify: `orchestration_states` (status/current_step/error) and the relevant pod logs filtered by your correlation id.

### E. Generated static sites (downstream, mostly automatic)
- A built/deployed page ships **git → GitHub Actions → Backblaze B2**. This is the site output, not the platform — it follows once a page reaches `build_status='deployed'`. Verify by checking the deployed git contents / the live page renders the expected components.

### F. idea.uk binary (separate from all of the above)
- Self-contained Go binary, file-based persistence, **not** k8s, **not** Backblaze. Build `GOPROXY=off GOTOOLCHAIN=local GOOS=linux GOARCH=amd64 go build -o idea .`, then `scp` → `mv -f` (the running binary is busy; the `&&` chain stops a failed build deploying a stale binary) → `systemctl restart idea`. The page is embedded (`go:embed`), so any page change needs this rebuild+redeploy. Email/forwarding are cPanel steps. **Exact commands and the env block are in the idea.uk pack** — follow them verbatim.

---

## Per-project quick reference

### gamesdesign adoption — skinner-box (`CONTEXT_PACK_adoption_skinner_box.md`)
- **Gather:** the fresh-from-chassis action code (`load_page_sections_from_spec`, `plan_sections`); the schema list; the skinner-box trio + queue-health rows; pod state.
- **Verify first:** are the section rows present **and read** (the decisive fork)?
- **Deploy:** mostly **B** (add section rows, set `pages.sections`) + **C** (re-issue the `needs_content_page` item, monitor). **Only if** the fork shows the relational rows aren't the source plan_sections reads → also **A** (fix the action, rebuild+redeploy the chassis). Page output then ships via **E**.
- **Verify deploy:** item runs >90s, `page_components` non-empty, `build_status='deployed'`, the skinner-box card gains its description.

### Flywheel-C thunder — checkpoint race (`CONTEXT_PACK_thunder_checkpoint_race.md`)
- **Gather:** `CHASSIS_await_loop_extract.txt`, `thunder_prepare_object_url_dispatch.go`; `\d awaited_requests`; launcher-def-config and export-rows queries.
- **Verify first:** `ActionParams.CurrentStep` holds the expanded loop-substep name at dispatch time.
- **Cleanup first:** kill the stuck launcher jobs; confirm no live `thunder_instances`.
- **Deploy:** **A** — add the `preRegisterAwaitedRequest` call, rebuild + redeploy the chassis (bump from `v1.0.1057`); then **D** — re-run via the `model-trainer` orchestrate trigger. (Fallback route also touches **B** for a migration.)
- **Verify deploy:** each `presign_checkpoints_iter_N_presign_one` logs `ClaimAwaitedRequest … claimed:true`; loop → `flatten` → `MANIFEST_WRITTEN` → `ssh_exec_launch`; `orchestration_states` healthy.

### idea.uk — go-live (`CONTEXT_PACK_idea_uk_golive.md`)
- **Gather:** nothing from a DB (file-based). Just the `idea-go/` code + the docs.
- **Verify first:** `service.go` present, ~658 lines, clean build.
- **Deploy:** **F** only — fix the build, flip the cPanel catch-all to forwarding, set `/etc/idea/idea.env`, rebuild + binary-swap + restart.
- **Verify deploy:** a test order's confirmation **arrives** and the **From** reads `idea.uk <idea-uk@leopardess.uk>`.

### imagery (the live example) (`CONTEXT_PACK_imagery_sprite_sheet.md`)
- Not being deployed yet — it's the tool's example. When built it follows the chassis pattern: **B** (migration adding `kind='sprite_sheet'`) + **A** (planner/adapter/action code → rebuild+redeploy) + **E** (the generated site shows the sprites). Gate each phase on an eyeball before the next.

---

## Cross-cutting cautions

- **Confirm before acting, every time** — the packs restate earlier context and can be stale; the fresh pull (schema/rows/code/pod state) is the truth.
- **Snapshot before any DB change**, resolve ids fresh, verify writes by re-querying.
- **Bump image tags** so rollouts pull the new build; re-tar artifacts so changes actually ship.
- **"Complete" ≠ "succeeded"** — verify positive evidence (components present, page deployed, manifest written, email received), not just a terminal status.
- **Know your target:** chassis image (platform) vs Backblaze (generated sites) vs the idea.uk box — they deploy by different paths and a change to one doesn't ship the other.
