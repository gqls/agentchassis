# HANDOFF — 2026-08-15 (b), fresh chat starts here: RFC_029 Phase 1 is IMPLEMENTED and COMMITTED — inert until the next roll; the council verdict and the observation window are what remain

**Supersedes `HANDOFF_2026-08-15_continue_here.md`.** That file's §2 named the RFC_029 §9
implementation as the lane's remaining platform work; an afternoon session (~13:00–16:00Z the
same day) did it — one coherent council-gated task, three commits, all verified against a
clean `git archive HEAD`. **The full implementation record is RFC_029 §10 — read that section
before touching anything resolver-shaped; it is newer than every other account, including
parts of this directory.**

## 1. What is now true

- **Phase 1 of the §9 ruling is on the shared branch** (commits `927e12bd9` test repairs,
  `1806371ef` implementation, `6b0736eed` NOTES pointer). All Go — **NOTHING is live until a
  chassis image built from a commit ≥ `1806371ef` rolls.** What it does once rolled:
  - `findFieldRecursive` (the whole-tree search) is deterministic: collect-all, sorted-key
    DFS, shallowest-first winner. Conflicting candidates STILL RESOLVE in this build, but log
    **WARN `aggressive search: conflicting candidates`** (field, every candidate path,
    winner). Unique/agreeing candidates behave exactly as before.
  - The mapped-field class from bugs_closed/213 §D logs **WARN `aggressive search: explicit
    single-segment mapping bypassed`** — observation only, no behaviour change.
  - The opt-in **`!` strict marker** (mirror of `?`) parses on both surfaces
    (`ExtractActionInputs` config keys; `ResolveInputMapping` dest fields): explicit
    resolution or loud failure, never the search. **Zero live carriers** until migration 417
    is applied (below).
  - The inner extraction chain has an enforced arm budget (floor 5 / ceiling 8,
    `resolver_arm_budget_test.go`) and descriptive arm names — the two-chains-both-called-
    "Strategy N" citation trap is closed (migration 402 carries the dated correction).
- **One correction to the ruling, made on its own evidence (§10.3):** D3's second named
  adopter — build-dispatch-loop's `asset_id?` — is **NOT flipped and must not be**: that
  mapping serves 636+ item types (402's own measurement); `!` there hard-fails every
  non-asset dispatch fleet-wide. Only image-build-handler adopts (measured first: 13/13
  spawns resolve `asset_id`, zero refusal-branch spawns in the retained window).
- **Registered:** concept register **CTS-060**; two new LANDMINES entries (the
  `!`-before-roll trap; migration-number races) plus a dated update to the 2026-08-08
  randomised-search entry — all synced to `doc_notes` via `landmines-verify-dispatch.sh`
  (verifications dispatched). Session record incl. missteps: NOTES `## 2026-08-15 (later)`.

## 2. OWED, in order — this is the fresh session's worklist

1. **READ the council verdict.** SUBMISSION_CORR `75091072-9d65-433e-8a30-84719dc3f30f`,
   submitted ~15:15Z 2026-08-15; both platform commits carry `Council-Submitted:`, so
   approval credits automatically — but a **REVISE/REJECTED must be acted on** (the code is
   already on the shared branch). Find the run by payload, not by printed id:
   ```sql
   SELECT current_step, status FROM orchestration_states
   WHERE collected_data->'input_data'->>'fix_correlation_id' = '75091072-9d65-433e-8a30-84719dc3f30f';
   ```
   ⚠ At session end (~15:50Z) `postgres-clients-0` was timing out on exec — known flakiness
   (bugfix_270's note, `90e3da509`); earlier queries the same session were fine. A missing
   row is dispatch latency (measured 29 min under load), NOT a dropped dispatch — do not
   resubmit on that evidence.
2. **After the next fleet roll:** verify the binary actually carries the change — ask the
   service (`logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'`, then
   `git merge-base --is-ancestor 1806371ef <stamp>`; startup line scrolls, so fall back to
   the known-value `/proc/1/exe` probe with a two-way control). Per SERVICE, not fleet.
3. **Then the observation window (48h minimum, a week preferred):** grep BOTH WARN messages
   fleet-wide — **stream `logs -f`, never `--tail` an old pod** (measured 92-second
   retention, CTS-059). What the two message populations decide is written in §10.2.
4. **Then hand-apply `417_image_build_handler_asset_id_goes_strict_HOLD.sql`** — it is
   SIDECAR_RE-excluded from the runner on purpose; its header carries the mandatory binary
   check. Applying it before the roll silently re-arms the search it forbids (LANDMINES).
   Afterwards confirm the child still resolves `asset_id` (13/13 baseline, §10.3).
5. **Phase 2 (conflicts resolve NOTHING) only after 1–4**, on §9 D2's precondition: zero
   conflict WARNs over the window, or every observed pair explicitly mapped first. The flip
   site and the tests that change with it are marked in code and in
   `unified_extractor_search_test.go`'s header. Phase 2 is its own council-gated task.

## 3. Traps measured THIS session (fresh session: read before acting)

- **The working tree does not compile** — another session's untracked
  `publish_site_action.go` breaks `platform/orchestration/actions`. Not ours; do not fix it
  into your task. Build/test against `git archive HEAD` + your files (worked recipe in NOTES).
- **Migration numbers race.** 413→416 were claimed by concurrent sessions within hours during
  this one session (our file was renumbered twice, ending at 417). Re-check your number
  against BOTH the tree (untracked files included) and `git log --all` immediately before
  committing — full check in the new LANDMINES entry.
- **LANDMINES.md is high-traffic today:** our uncommitted appends rode to HEAD inside another
  lane's commit (`90e3da509`) as same-file passengers — harmless, forward-only held, but
  expect the same; and use `./scripts/landmines-verify-dispatch.sh`, **never**
  `landmines-sync.py --apply` directly (the sync consumes the verification signal — its own
  LANDMINES entry).
- **Do not revert a probe with `git checkout --`** on a file carrying uncommitted work — it
  wiped this session's `unified_extractor.go` edits mid-task (recovered from context; NOTES
  records it). Edit the probe back out instead.

## 4. Session-start checklist

1. `git log --oneline -10`; re-read THIS file from disk (it goes stale in hours here).
2. Read **RFC_029 §10** (implementation record + corrections), then §9 (the ruling). Only
   then the code — the doc block on `findFieldRecursive` and the `!` sections in
   `action_inputs.go` / `input_mapping.go` are the contract.
3. Work §2's list top-down; items 2–5 are gated on the roll, so if no roll has happened,
   item 1 (the verdict) is likely the only actionable piece — everything else in this lane
   is done and closed.
4. If anything asset-shaped resurfaces: the morning handoff's §3–§4 items (dormant
   `unresolved` rows, the closed 248 rituals) still stand — census with `AND
   status='active'`, wire-check with must-be-absent controls, `sites.locked_at IS NULL`.
