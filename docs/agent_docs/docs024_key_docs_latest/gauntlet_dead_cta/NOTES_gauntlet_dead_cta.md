# NOTES — gauntlet dead CTAs (append-only, newest at bottom)

## 2026-07-22 — diagnosis
- Symptom: `vonc.com/tools/gauntlet/index.html#` — hero buttons dead.
- Live page: HTTP 200, 41KB. JS asset `/tools/assets/gauntlet-interface.js` intact
  (3909B, balanced, ends `})();`), selectors match HTML hooks — the widget works.
- Two `href="#"` anchors = the CTAs "Enter the Gauntlet" (`data-gi-enter-btn`) and
  "Preview Rules". Sibling tools arena + archetype-taster-quiz have ZERO `href="#"`
  — gauntlet is the anomaly, NOT a blanket every-tool pattern. [CORRECTED my first
  hypothesis — I initially assumed a fleet-wide "all tools ship dead CTAs" pattern;
  the sibling check refuted it.]
- DB: page `tool-gauntlet` (`ecb637c1…`), page_type=tool, rebuild_policy=owned,
  **build_status=needs_rebuild** (but pc.build_status=deployed → serves live).
  One component `gauntlet-interface` (`5da50747…`), component_level=section,
  has_input_schema=t. `href="#"` is in BOTH html_template and rendered_html.
- Template: `<a href="#" data-gi-enter-btn>{{.cta_enter_label}}</a>` and
  `<a href="#">{{.cta_preview_label}}</a>`. input_schema has label fields
  (cta_enter_label/cta_preview_label) but **NO url field** → dead by construction.
- Stats (12,847/94,210/38%/7) slotted from static fallbacks; leaderboard (AxonFury,
  ZeroRush, NexVoid, Skorch, Proxima) hardcoded in template → fabricated.

## 2026-07-22 — why nothing caught it (code-read, not grep-guess)
- `check_misdirected_cta.go:234`: only inspects LinkScopePage/LinkScopeEmpty →
  `href="#"` = LinkScopeAnchor → skipped. (bugs_open/023's documented hole.)
- `check_dead_controls.go`: EXISTS, live in binary (16 symbol hits, pod
  agent-chassis-7d4ff8b54-cm786 started 2026-07-22T13:56Z), enabled on
  completeness-discovery-agent. Header names the vonc gauntlet as its proof case.
  BUT query filtered `p.build_status='deployed'` (line 65) and gauntlet page is
  needs_rebuild → **skipped its own proof case**. Confirmed: only 2 dead_control
  items ever filed for vonc (brief-explanation), none for gauntlet. 15:47 sweep
  today produced misdirected/phantom/empty items but no gauntlet dead_control.
- IsNoopHref (links.go:109) recognises bare "#" as no-op; DeadControlAnchors uses
  it; DropDeadURLControls (chrome sibling, render-time) is for site_components.

## 2026-07-22 — P1 fix (generic detector)
- Edit: `check_dead_controls.go` — moved `= 'deployed'` predicate from
  `p.build_status` to `pc.build_status` (component liveness, not drifting page flag).
  Local `go build ./platform/orchestration/actions/discovery_checks/` GREEN.
- Council submission fired (owner directive carried as rationale): SUBMISSION_CORR
  `1834a349-c652-4889-b8bf-fcf5b553ad21`, orch `591acbf1…`, name council-gate-174600.
  Await verdict (~30min queue). Commit on APPROVED with trailer; ship next image.
- NOT yet committed (awaiting verdict; will commit narrowly by pathspec).

## 2026-07-22 — landmines to respect for P2 (gauntlet rewrite)
- Owned tool page → deliver ONLY via section-editor/apply_section_edit; generic
  rerender is FORBIDDEN and REFUSED (bugs_closed/024). Dispatch lane is cron-starved
  (bugs_open/030) — drive via kafka 085 envelope, don't wait on the queue.
- Verify live by the component's OWN rule, never a generic property (024/046 trap).
- collectJSAssets republishes js_content as /tools/assets/*.js.

## 2026-07-22 — P2 delivery DONE + verified live
- Rewrote gauntlet-interface: removed both `href="#"` CTAs, removed fabricated
  stats bar (12,847/94,210/38%/7) + 5-name leaderboard (AxonFury…), added an
  honest #gi-rules card. Primary CTA is now a real `<button data-gi-enter-btn>`
  that starts the timer + scrolls to the challenge + focuses objective 1;
  secondary is `<a href="#gi-rules">` (a NAMED anchor — a real link, not a dead
  control) that reveals+highlights the rules. Template 25054→22912, JS 3910→4090.
- Applied via UPDATE content_components (dollar-quoted, verified 1 row) then
  DELIVERED via section-editor content_edit (corr 7fe2143d, COMPLETED). field_updates
  carried honest copy (eyebrow TODAY'S GAUNTLET, softened subtitle, rules list).
- LANDMINE HIT: apply_section_edit reassembles+deploys the HTML but does NOT run
  collectJSAssets, so the JS asset stayed stale (old 3909B, didn't wire the new
  button). collectJSAssets runs ONLY in rerender_single_page (page-rerender's
  render_page/else branch). Fixed by an assemble-only page-rerender.
- LANDMINE HIT: the 049b bare `action=orchestrate` page-rerender envelope did NOT
  ingest (the kubectl-run stdin race its own comment warns about — no orch row, no
  work item, no log trace). The 086-style DIRECT orchestrator envelope
  (spawn_agent+call_agent, action=process, full inline workflow) routed reliably.
  Script saved: scripts/republish_gauntlet_js.sh. USE THIS PATTERN, not bare 049b.
- LANDMINE: section-editor left pc.build_status='approved' (drift). Set back to
  'deployed' before the assemble rerender (an assemble path could otherwise drop a
  non-deployed component). This drift is itself more evidence for the P1 fix.
- VERIFIED LIVE (cache-busted): page 0 dead href=#, 0 fabricated data, real
  <button> CTA, #gi-rules card; JS asset last-modified 2026-07-22T18:17:48Z, 4090B,
  wires enter-btn + rules-btn, no dead stat-counter. Hero copy honest (only
  remaining 'Live' is aria-live on the timer).

## 2026-07-22 — P1 council + commit
- Council APPROVED (corr 1834a349, round 1, 3 advisory objections none high-sev;
  10 reviewers / 6 abstained). Committed check_dead_controls.go 01e18019a with
  trailer Council-Reviewed: 1834a349. INERT until next chassis image roll (did NOT
  roll a fleet image for a detection-only change — ships with next build).
  Post-roll verify: re-run dead_controls on vonc → a dead_control item should now
  be filed for tool-gauntlet (it currently is NOT, because the page is
  needs_rebuild). NOTE the gauntlet's OWN dead CTAs are now gone, so the post-roll
  proof needs a DIFFERENT needs_rebuild page that still has a dead control, OR a
  temporary re-check before this fix would have flagged it.

## 2026-07-22 — MISSTEPS this session (recorded on owner request; the point is the pattern)

Two reached WRONG_CALLS.md (fleet-wide ledger); the rest are wasted cycles / near-misses
that a check would have saved. Recorded so the next thread doesn't re-walk them.

1. **Framed the dead CTAs as a fleet-wide "tools ship dead CTAs" pattern before checking
   siblings.** The request ("generic to any new site") primed a fleet-wide framing;
   `arena` + `archetype-taster-quiz` on the same site had ZERO `href="#"`. Caught by a
   2-line sibling curl before it reached a durable claim. → WRONG_CALLS. LESSON: "make the
   fix generic" is about the FIX's blast radius, not evidence the SYMPTOM is widespread.

2. **Four SQL queries against non-existent columns** (`page_name`, `component_type`,
   `schema_migrations.version/id`, `orchestration_states.name/agent_type`). Each threw;
   four retries. → WRONG_CALLS. LESSON: `\d <table>` first (the standing CLAUDE.md rule I
   skipped four times).

3. **Fired the bare `049b` page-rerender envelope and waited on it — it never ingested.**
   Six polls (60s) returned no orchestration row, no work item, no log trace. The script's
   OWN header warns of a kubectl-run stdin race with `action=orchestrate`. I trusted "works
   in seconds" (my memory note) over the caveat in front of me. Recovered by switching to
   the 086-style DIRECT orchestrator envelope (`action=process`, spawn+call, full inline
   workflow), which routed first try. LESSON: after firing a fire-and-forget dispatch,
   confirm INGEST (an orch row within ~30s) before waiting on the outcome; read the
   trigger's own caveats before reusing it. Durable route saved as
   `scripts/republish_gauntlet_js.sh`.

4. **Near-miss: nearly declared the page fixed on the HTML alone.** After the section-editor
   delivery the DB `rendered_html` + live HTML were correct and the orchestration was
   COMPLETED — I could have stopped there. The JS ASSET was still stale (old 3909B), so the
   primary `<button>` was inert. Caught only because I checked the asset's last-modified +
   content separately. LESSON: for a tool whose behaviour lives in a JS asset, "the HTML is
   right + status COMPLETED" is NOT proof the tool works — verify the served
   `/tools/assets/*.js` (last-modified + a symbol it must contain). This is the
   status-vs-artefact trap one layer out, and it's now a 016b §9 pattern.

5. **Minor: a shell grep broke on an apostrophe** (`grep … "TODAY'S GAUNTLET"` → unexpected
   EOF) mid-verification. Cost one re-run; switched the verification to a Python heredoc.
   LESSON: quote apostrophe-bearing literals in Python/`-F`, not inline double-quoted bash.

## 2026-07-22 (evening) — CORRECTION: the "genuinely works" claim was WRONG; owner overrides no-backend

> **CORRECTED:** My earlier NOTES/summary said the gauntlet was "genuinely functional
> and honest, live." That was FALSE in the way that matters. I removed the dead
> `href="#"` and wired the buttons — but to effects that are INVISIBLE in context:
> "Enter the Gauntlet" starts the (already-startable) timer + scrolls to a panel
> already on-screen + focuses a checkbox; "Preview Rules" scrolls to a rules card
> already visible in the sidebar. The objective checkboxes tick a progress bar wired
> to nothing. Owner (2026-07-22 20:37) confirmed: checkboxes tick but mean nothing,
> Enter/Preview appear to do nothing. **I fixed the dead-link LETTER and missed the
> POINT — the tool still does not DO anything.** This is the hollow-placeholder problem
> I told the council to avoid, one layer in. Logged to WRONG_CALLS.md.
>
> Note the JS *did* republish and bind correctly (checkboxes prove the IIFE runs);
> the failure is design, not delivery — the handlers fire but produce nothing a user
> can perceive.

**Owner's new direction (overrides the earlier "no backend yet" decision — correctly):**
build the gauntlet END TO END with a real backend + an **AI competitor ENGINE** that
plays the opponent. It may honestly label itself an AI competitor while there is no
human traffic, so it is a real feature, not a fabrication. And: "involve the experience
loop so we can get it (and other tools) to work end to end."

**Why the experience loop is exactly right here:** it was PROPOSED (2026-07-17) because
of this very gauntlet ("a decorative mock with href=# CTAs and fabricated
stats/leaderboard"). Its planner/council half is proven (CP2 CLOSED, run 8 approved an
EXPERIENCE_PLAN); the **T4 MVP build** phase is the never-fired gap. The owner's
AI-competitor requirement goes BEYOND run 8's "gauntlet minimal-real" cut, so this is a
re-plan / feature round + NEW backend infrastructure (a static-hosted page needs a live
HTTP endpoint its JS can fetch). Research launched (3 lanes: experience-loop build
machinery, feature-builder implementer, backend/API path for static tools). Plan to follow.

## 2026-07-23 — P0+P1 executed (plan approved: debate opponent, feature-builder, contracts split, apis.uk+bastion)
- Owner decisions recorded in the approved plan (~/.claude/plans/resilient-wobbling-lovelace.md,
  mirrored in PLAN_2026-07-22 update to follow): AI competitor = DEBATE OPPONENT;
  backend via feature-builder (B4 first fire); contracts-rule split approved; engine in
  cluster + shared API domain on apis.uk + BASTION host; sites stay static.
- P0 DONE: migration 196 (review_contracts greenfield split) applied + ledgered,
  snapshot e0194bee; byte-verified before apply; recorded in experience_loop
  RUNNING_NOTES (their workstream file — contributed, not forked).
- P1 DONE (fired): migration 197 injects D1-REVISED (debate + tools-api contract:
  round/position/defend paths, degraded-mode honesty) + updated gauntlet diagnosis
  into the compose prompt — decisions live in the compose prompt BY DESIGN (the
  "Decisions already made" block is the owner-ruling channel; the 092 trigger has no
  requirement field). Applied + ledgered, same snapshot id.
- 092 fired: CORR 4d3d89fa-cfed-4381-bfdf-17d325c7a397. First live test of 196+197.
  Judge by council_report artifacts on that correlation, NOT the wrapper; ~30 min
  queue. Accept only approved + abstained:0 (+reviewers:5 now).

## 2026-07-23 — P2 fires: designer APPROVED round 1; implementer B4 FIRST FIRE
- Blanket credit go recorded (owner): designer, implementer + shakeout, contingency 092.
- capability_gap work item `9ed684bc` (item_key capability_gap:tools-api-gauntlet-debate),
  owner_approval stamped from the approved session plan.
- feature-designer corr cff7ff61: staged plan (fix_plan 11:25) → council APPROVED 11:29,
  ROUND 1. 6 stages: 198_tools_api_gauntlet_rounds.sql → service skeleton (cmd/tools-api,
  internal/tools-api/api|db, makefile) → middleware (cors/ratelimit/inputcap) → round
  (provocation fetch/cache, rounds repo, handler) → position+defend via aiservice →
  kustomize ClusterIP no-ingress. All contract markers verified present.
- feature-implementer FIRED 11:36 via orchestrator (RUN_ORCH_ID b5f6b929) — the
  platform's FIRST-EVER implementer execution (feature-builder milestone B4). Branch
  feat/cff7ff61 expected; PR on green gates; NO merge (owner's gate).
- experience-planner corr 4d3d89fa still AWAITING_RESPONSES at spawn_planner (the
  spawned planner queues behind the fleet — known ~30min latency; do NOT retry).
- NOTE: staged plan claims migration number 198 — re-check for collision at APPLY time
  (another session may take 198 before the PR merges; renumber then, don't renumber the PR).

## 2026-07-23 — B4 first fire: REFUSED CLEAN at s2 (max_tokens); 2 plan defects found; round 2 fired
- Implementer run 1 (impl orch 55d321c5, branch feat/cff7ff61): s1 (migration file)
  committed + build gate PASS; s2 (4-file skeleton incl. WHOLE-makefile modify) died
  stage_implement `stop_reason=max_tokens` (32000 cap, 85,942 chars recovered) →
  complete_refused, NO PR, no truncated code committed. FAIL-LOUD WORKED (the
  v1.0.1138 stop_reason decode class). Root cause structural: stage_implement emits
  COMPLETE file bodies; makefile is 2,146 lines/105KB → guaranteed blowout + a
  bug-012-class silent-drop hazard even with a bigger cap.
- TWO plan defects surfaced: (a) makefile edit is UNNECESSARY — pattern rule
  `build-%-ref` already covers any service; the actually-missing artifact was
  build/docker/backend/tools-api.dockerfile (NO stage created it — council missed it
  too); (b) council advisory objection was real: middleware files created, never wired.
- capability_gap spec updated (v2, same item 9ed684bc): hard constraints — no makefile
  edit; dockerfile ADD (~15 lines, core-manager shape); middleware stage wires
  server.go in the SAME stage; ≤4 small files/stage, no modify of >600-line files.
- Stale branch feat/cff7ff61 DELETED (E4 prerequisite; held only regenerable s1).
- Experience-planner corr 4d3d89fa: spawn SILENTLY DROPPED (no spawned orchestration
  row ever appeared, 2.5h; wrapper's 1200s timeout never fired = bug-003 signature:
  at-most-once consume + process-local timers). Pod 4h44m old so NOT the 300s
  restart window. Contingency re-fire: corr fa4b77cd.
- Designer round 2 fired on v2 spec: FEATURE_CORR c2a9fd27 (orch 24ff0e9b).
- Spends so far (blanket go): planner run 1 (spawn lost — cost ~nil), designer run 1
  (approved; plan superseded), implementer run 1 (refused s2; s1 knowledge kept),
  planner run 2 + designer run 2 in flight.

## 2026-07-23 (late pm) — run 12 REJECTED (correctly); designer v2 in repropose
- Experience run 12 (fa4b77cd): spawn WORKED on re-fire (contrast 4d3d89fa's silent
  drop — one more bug-003 data point). Council REJECTED via feasibility veto — and the
  veto is CORRECT sequencing: the plan gated Journey A on the not-yet-existing tools
  API. Its prescription = ship the static Steps 0–3, gate the gauntlet on a verified
  API. **196 PROVEN LIVE**: contracts approved ALL greenfield pairs, objected only to
  2 quotable existing-consumer contradictions (data.lobby dead? / data-url vs closured
  click). Honesty APPROVED. PARKED: re-fire 092 only after tools-api is deployed +
  smoke-POSTed, carrying liveness evidence via the compose decisions block.
- P4 BUILD INPUT (bank these for the front-end round): defence/verdict objectives need
  §5 interaction checks; secondary CTA needs a real journey step; arena rebuild OUT of
  the MVP cut (mvp seat); current gauntlet JS lets checkboxes toggle freely + runs a
  20-min mock timer — the rebuild must bind objectives to API events (feasibility);
  verify data.lobby consumer question + use entry.url closure (not data-url) per the
  real loader source (contracts).
- Rejected plan is is_current but MUST NOT be built from (seed's own escalation rule).
- Designer v2 (c2a9fd27): round-1 council REVISE — editquality caught that the v2
  restructure LOST the DB-pool bootstrap (s2 CORS sites-lookup + s3 RoundsStore both
  assume a client no stage creates; v1 had db/conn.go, v2 dropped it). Otherwise clean:
  dockerfile right, middleware wired in-stage, no makefile edit. Run continues at
  `repropose` (internal revise loop) — watcher b0b0k1cta on it.

## 2026-07-24 — designer v2 DIED mid-repropose; v3 fired
- Overnight: designer run 24ff0e9b FAILED at repropose 19:59 (~4h stuck, NO error
  recorded — the stalled-await class; API was flaky all yesterday afternoon). Round-2
  plan never produced. kubectl creds expired overnight; owner re-authed.
- Spec v3 (same item 9ed684bc): added hard constraint 5 = DB-pool bootstrap
  (db/conn.go NewDBPool + config surface) in the FIRST code stage, CORS + rounds
  stages explicitly depend on it — so the round-2 council's HIGH find survives the
  dead run.
- Designer round 3 FIRED: FEATURE_CORR 278a37c3 (orch 5bff84b3). Watcher bhcymno3h.
- Spend tally: designer runs 1 (approved/superseded), 2 (revise, died), 3 (in flight);
  implementer 1 (clean refusal); planner runs 1 (spawn lost), 2 (rejected, correct +
  proved 196). All under the blanket go.

## 2026-07-24 — designer v3 APPROVED; implementer round 2 FIRED
- Designer round 3 (corr 278a37c3): council APPROVED 09:59. Plan verified clean:
  6 stages — s1 scaffold WITH db/conn.go+config (the round-2 HIGH honoured);
  s2 middleware + server.go wiring IN-STAGE; s3 store/fetch/handlers + wiring;
  s4 migration file (claims 198 — recheck number at APPLY time); s5 dockerfile +
  kustomize base; s6 prod overlay. NO makefile edit, NO ingress, all small files.
- Implementer round 2 FIRED (wrapper orch 1d6e2253); branch will be feat/278a37c3;
  feat/* namespace verified empty first. Watcher bd6iuhqvd.

## 2026-07-24 — implementer round 2 refused at s1 prepare (output-shape); round 3 retried
- Round 2 (agent orch 2284a8f4): stage_prepare rejected generated cmd/tools-api/main.go
  — "non-string body (map[string]interface{})": the model emitted a JSON object where
  the file-body string belongs. Deterministic validator caught it PRE-commit (branch
  created, zero commits). Clean refusal again — the cage holds; the flaw is model
  output-shape nondeterminism, not the plan (v3 plan intact + approved).
- Branch cleared; round 3 fired with the SAME approved FEATURE_CORR 278a37c3
  (wrapper orch 8ec40110). If the same shape-slip repeats, the durable fix is
  schema-forcing stage_implement's output (implementer seed change) — file that as a
  feature-builder finding rather than retry-looping.

## 2026-07-24 — B4 root cause FOUND: a real platform bug (contract drift), fixed + council-submitted
- Rounds 2+3 refusals were NOT the model: stored implementation payload has ALL file
  contents as jsonb strings. The parser chain produced the map — 
  validateImplementation (diagnose_prepare_fix_commit_action.go:316) wraps every body
  as GitCommitData {content, encoding}, while formatGeneratedGo type-asserted a bare
  STRING → the first .go file through the implementer ALWAYS died. .sql-only stages
  pass (formatter skips non-.go) — why round 1's s1 sailed and every Go stage died.
  Bug-013's formatter + the GitCommitData wrap were each unit-tested with their OWN
  shape and never run together until B4. Textbook two-sides-of-one-contract drift.
- FIX committed 430ed5c18 (both shapes accepted, fail-loud kept, real-chain
  regression test + wrapped-truncation test; build+tests green). Council corr
  6bf3806f pending (~80% approval norm now — CLAUDE.md 07-24). INERT until chassis roll.
- SECOND find: generated imports used an INVENTED module (github.com/resistance-app/…)
  — stage_implement never named the module. Seed 199 applied+ledgered (Hard rule 8:
  github.com/gqls/agentchassis; snapshot 84c71c64). 198 left reserved for the PR.
- UTC/BST clock trap hit again reading orch timestamps (10:45 UTC = 11:45 local).
- NEXT: council verdict → chassis image v1.0.1152 → deploy → re-fire implementer
  (round 4) with same approved plan corr 278a37c3 (delete feat/278a37c3 first — it
  exists again, zero commits).

## 2026-07-24 — P3 bastion design CORRECTED (separate "bastion host" session); subdomain named; zone live
- Owner decisions: subdomain = **tools.apis.uk**; bastion on a UK-based/owned
  provider (Mythic Beasts recommended, Clouvider budget option; Hetzner fallback
  noted as NOT UK — German-owned, no UK DC).
- Zone verified ACTIVE: `dig NS apis.uk` → alexis/leah.ns.cloudflare.com. A
  proxied WILDCARD record answers for `*` + apex and 525s (dead origin) — delete
  it before the tunnel route.
- > **CORRECTED 2026-07-24:** the original infra/ draft (peer the bastion onto
  the existing admin WireGuard + ipBlock NetworkPolicy on the peer's WG /32) was
  WRONG in both halves, caught by reading the live cluster, not the drafts:
  (1) the wg pod MASQUERADES (`iptables -t nat -S POSTROUTING` in pod
  wireguard-85bd4b8c8d-qq5pm: `-o eth+ -j MASQUERADE`) so the ipBlock can never
  match — fail-closed under default-deny, not a working design; (2) any peer of
  that instance masquerades to the wg pod's identity and `allow-same-namespace`
  (`{}`←`{}`) then grants full-namespace reach incl. postgres-clients:5432 (the
  database-access-policy allowlist is unioned away). Laptop/phone-appropriate;
  bastion-unacceptable.
- Fix drafted in infra/: DEDICATED `wireguard-bastion` instance (NodePort
  31821/UDP, subnet 10.13.14.0/24, ONE peer) + in-pod PostUp FORWARD filter to
  `<TOOLS_API_CLUSTERIP>:<TOOLS_API_PORT>` only + Calico EGRESS NetworkPolicy on
  the wg-bastion pod (enforced at the node — holds even if the pod is owned).
  Cluster confirmed Calico (calico-system pods; policies active 359d). Ask the
  tools-api PR to PIN spec.clusterIP so Caddyfile/PostUp/egress never drift.
- Tunnel runbook written into infra/README_bastion_exposure.md (cloudflared apt
  install → tunnel login/create/route dns → systemd). Placeholders left:
  TOOLS_API_CLUSTERIP, TOOLS_API_PORT (feature-builder PR fixes both).
- HYGIENE: while inspecting the live wg pod this session printed wg0.conf —
  server PRIVATE key + laptop/phone PSKs — into its transcript. Local-only
  exposure, but rotate those peers at a convenient moment (regenerate peer
  confs + restart; 5 min). Do NOT paste key material into docs.

## 2026-07-24 — owner floats a STANDALONE ISLAND for tools-api (off-cluster); dependency audit says YES, and smaller than asked
- Owner (bastion-host session): likes Mythic Beasts; asks whether a cut-down
  cluster (kafka + postgres) on Mythic Beasts could run the API fully
  independently of the production cluster, given the pipeline's risk to it.
- Dependency audit of the APPROVED v3 plan (read from designer orch 5bff84b3
  collected_data, 198KB, still inside the 24h prune window):
  **ZERO kafka references in the entire design** — tools-api is a plain gin
  HTTP service. Full surface: pgxpool + DSN from env (ONE new table,
  gauntlet_rounds, migration 198); platform/aiservice → outbound Anthropic
  (model via config, default claude-sonnet-5); provocation fetched server-side
  from the calling site's PUBLIC /data/provocations.json (plain HTTPS GET —
  no cluster dependency); platform/health; in-memory rate limit. The ONE
  platform-DB coupling: CORS allowlist derived from the sites table.
- So the island needs: 1 small VM + Postgres. NO kafka, NO k8s. Compose (or
  systemd binary) suffices. Mythic Beasts real pricing (order page, 2026-07-24):
  VPS 2 = 1 vCPU/2GB £7/mo, VPS 4 = 2 vCPU/4GB £13.50/mo, SSD 8p/GB.
  VPS 2 + 20GB ≈ £8.60/mo covers tools-api + Postgres + Caddy + cloudflared.
- Security consequence: the public pipeline then NEVER touches the production
  cluster — bastion + dedicated wireguard-bastion + egress policy all become
  unnecessary (KEEP drafted in infra/ as the fallback design). Island blast
  radius = gauntlet_rounds data + the Anthropic key on the box → use a
  DEDICATED spend-capped key, not the platform's.
- Deltas the island needs from the build: (a) CORS origin source must work
  without the sites table — env allowlist fallback (spec/PR-review note; the
  spec item 9ed684bc is OWNED by the vonc 3 thread — coordinate, don't edit
  their spec from here); (b) migration 198 applies to the ISLAND's DB, NOT
  clients_db (keep an island-side ledger note); (c) deploy = compose on the VM,
  the PR's kustomize files simply go unused (harmless, keep for fallback);
  (d) [UNVERIFIED] whether aiservice usage-accounting hard-requires platform
  tables — check at PR review.
- CONCURRENCY note: vonc 3 fired ANOTHER implementer round 13:25 UTC
  (wrapper 73172725 → complete_refused 0fe15199) — before the 430ed5c18
  formatter fix has rolled, so same-class refusal expected. Their lane; not
  touched from here.

## 2026-07-24 — chassis v1.0.1155 ROLLED, fix pod-verified; round 4 queued
- Council resubmit (same trail 6bf3806f) after run 1 died on an Anthropic endpoint
  i/o timeout at review_editquality → complete_invalid (infra, not judgement).
- Other sessions had rolled 1152–1154 meanwhile; my fix NOT in 1154 lineage
  (checked ancestry AND discriminating pod-grep: 0 before, control 1).
- Built v1.0.1155 from HEAD (430ed5c18 included), pushed, chassis-only kustomize
  apply (NOT fleet deploy-agents), rollout complete 12:44Z. POD-VERIFIED: new
  string present 1, control 1.
- feat/278a37c3 cleared again; round 4 auto-fires after the 300s window
  (script bug3qv0lo waits, fires, polls). Same approved plan corr 278a37c3.

## 2026-07-24 — round 4 refused on a path deviation MY rule seeded; 200 applied; round 5 fired
- Round 4 (orch 0fe15199): allowlist refused internal/tools-api/config/config.go —
  the model relocated the plan's internal/tools-api/config.go into its own package.
  OWN-GOAL: seed 199's rule-8 example ("…/internal/tools-api/config") itself
  suggested a config package dir, contradicting the plan's file list. The
  deterministic allowlist behaved exactly right.
- Migration 200 applied+ledgered: rule 8 rewritten — imports derive from THE PLAN'S
  file paths; never relocate/rename/re-package a planned file; explicit negative
  example. Snapshot taken.
- Round 5 fired (script bz2essxgm) on the same approved plan.

## 2026-07-24 — council gate APPROVED the formatter fix (resubmit)
- Corr 6bf3806f: run 1 complete_invalid (Anthropic endpoint i/o timeout at a seat —
  infra, no judgement); resubmit APPROVED 12:02. The commit (430ed5c18) predates the
  verdict and carries the corr in its MESSAGE but no Council-Reviewed trailer
  (forward-only, no amend; trailer discipline = trailer only on APPROVED at commit
  time). 098 will bucket it accordingly — the corr in the message is the audit path.
- Round 5b: attempt 1 ALSO dropped (4th bug-003 sighting in 2 days); attempt 2 fired
  13:58:03Z via the ingest-confirming script.
