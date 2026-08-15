# HANDOFF 2026-08-15 — continue here: Phase 2 (publish seam) is BUILT, APPROVED, LIVE and mid-canary; next is reading the canary result, then Phase 3 (ZIP)

**Start here cold.** Read order: this file → `PLAN_2026-08-14_site_delivery_and_editor.md`
(the owner-approved design, Phases 2–6) → NOTES tail (the 2026-08-15 entries,
newest at the bottom — they carry every misstep and its fix).

## 0. State in one paragraph

The publish seam is **live end to end**: `platform/publish` + the `publish_site`
action shipped in the v1.0.1303 roll (verified at the binary: stamp
`5e075a6f9…` on both replicas, `71e4d9736` is its ancestor), council **APPROVED**
round 2 on corr `21aba3f5-ca44-4220-a680-d99f5ef0a90b` (round 1 REVISE was
correct — see §3), migrations 412 + 423 applied (columns + `publish_project`
uniqueness), seed 422 **hand-applied** ~22:00Z (site-publisher repurposed with a
full-row snapshot in `agent_definitions_backup`, publish-reconciler +
600s schedule live). The **canary is in flight**: `noted.co.uk` opted in to
`b2worker` → `noted.ugg2.com` at ~22:04Z; the first real reconciler pass was due
~22:11:23Z. Everything else in the fleet is publish_target NULL = OFF.

## 1. FIRST: re-arm the canary once the NEXT release rolls — the first pass ran and FOUND A REAL DEFECT (now fixed, committed, inert until the roll)

> **SUPERSEDES the "in flight" language below (same night):** the first
> reconciler pass executed the full chain — tick 22:11:23Z → stamp →
> publish-reconciler → spawned `agent-site-publisher-c08f7091-rl8hc` →
> `publish_site` ran — and failed at the FIRST upload with **HTTP 411
> MissingContentLength**: B2's S3 gateway refuses a bare stream, and
> `copyOne` piped the download straight into PutObject. **Fixed in
> `b4981634d`** (buffer to a seekable reader; both test fakes now enforce
> the gateway's contract, mutation-proven). Zero objects landed under
> `noted.ugg2.com/` (it failed on the first file), `published_hash` was
> never written, and the **canary is DE-ARMED** (`publish_target=NULL`,
> `publish_project` kept) so the 600s retry loop doesn't feed the failure
> sweeps while the fix waits on the next owner release.
>
> **To re-arm after the next roll** (verify `b4981634d` is an ancestor of
> the new stamp first, per §"verify the running pod" in
> `422_site_publish_reconciler_HOLD.sql`):
> ```sql
> UPDATE sites SET publish_target='b2worker', updated_at=now()
>  WHERE domain='noted.co.uk';  -- publish_project already set
> ```
> Then follow the acceptance section below unchanged.

## 1b. The canary acceptance (once re-armed on the fixed binary)

The expected chain, and where each link leaves evidence:

1. `scheduled_tasks` row `site-publish-reconciler` fires (600s interval):
   `last_triggered_at` advances past 22:11Z.
2. Its pre_query stamps `site_publish_checks` (one row, noted.co.uk) and
   dispatches `publish-reconciler` → orchestration named
   `sched-site-publish-reconciler-<ts>` in `orchestration_states`.
3. publish-reconciler spawn→calls `site-publisher` (a SPAWNED pod —
   `agent-site-publisher-*`; the standing chassis has no B2 creds, by owner
   ruling 2026-08-08). ⚠ The spawn→call handshake **fails ~half the time
   fleet-wide** (memory: spawn-call-handshake-races) — a FAILED orchestration
   here is NOT a defect of this lane; the next 600s tick retries. Do not
   cancel the failing row.
4. `publish_site` in the spawned pod: lists `portfolio-sites/noted.co.uk/`,
   TreeHash, publishes on drift (published_hash was NULL = max drift), copies
   under `noted.ugg2.com/`, ETag-verifies at the destination listing, fetches
   `https://noted.ugg2.com/index.html?pub=<hash>` and compares sha256 vs
   origin, and ONLY then writes `sites.published_hash/published_at`.

Acceptance queries (all in one):
```sql
SELECT domain, publish_target, publish_project, published_hash, published_at
  FROM sites WHERE domain='noted.co.uk';
SELECT * FROM site_publish_checks;
SELECT orchestration_name, current_step, status FROM orchestration_states
 WHERE created_at > '2026-08-15T22:11:00Z' AND orchestration_name LIKE '%publish%'
 ORDER BY created_at DESC;
```
Independent artefact checks (never trust the status):
- `b2 ls b2://portfolio-sites/noted.ugg2.com/` — the copied tree.
- `curl -s https://noted.ugg2.com/index.html | sha256sum` must equal the origin
  hash **`b4416c3208f9df047c044a526246f06c4fca03c4b02ec470e9e6af4e01f82ceb`**
  (captured pre-publish; origin copy also in the prior session's scratchpad).
- Second pass no-drift proof: wait one more 600s tick, then the newest
  publish orchestration's result should carry `skipped: no drift` and
  `published_at` must NOT advance. (Result payloads live in the orchestration
  row's collected_data / the spawned pod's logs.)
- Isolation: `SELECT count(*) FROM sites WHERE publish_target IS NOT NULL;` = 1.

If the canary FAILS repeatedly (3+ ticks, no published_hash): read the spawned
pod's logs (`kubectl logs agent-site-publisher-…`), not the chassis logs —
`orchestration_states.processing_node` names the pod that ran each step.
Plausible causes in order: handshake race (retry, benign), storage env missing
in the spawned pod (would say "portfolio store unavailable — is this a
storage-enabled spawned pod?" — means the spawner gate changed), CF cache
serving stale on the acceptance fetch (would be `accepted:false` with the two
sha256s in the reason — check the served copy by hand with a different
cache-buster).

## 2. What is DONE (all committed; do not redo)

- `platform/publish/` (seam, TreeHash `th1:`, S3Source, b2worker backend,
  cfpages as a LOUD refusal) + `publish_site` action + registry entry —
  commit `71e4d9736`, tests green including the negative proofs.
- Migration **412** applied (4 nullable columns on sites, all-OFF default,
  guarded); **423** (`423_publish_project_unique.sql` — ⚠ the number is
  SHARED with an unrelated concurrent migration; the FILENAME is the
  identity) applied: partial unique index on publish_project.
- Seed **422** (`_HOLD` sidecar) hand-applied on v1.0.1303 — its header
  carries the STATUS line, the apply command, the re-verification
  obligations, and ⚠ there is NO `--record-only` for sidecars (the runner
  refuses by design; the record is the header + NOTES).
- Council: corr `21aba3f5` APPROVED r2. Trailers: `Council-Submitted:` on
  `71e4d9736`/`cd5490866`, `Council-Reviewed:` on `d00647ef4` (verdict was
  read first). Advisories triaged in NOTES (2026-08-15 late evening entry).
- Register: **DGH-008** (+ index row) — status updated as things landed.
- LANDMINES: new entry — the all-zeros sha is git's null-sha constant and a
  FALSE negative-control for binary probes; use a random 40-hex value.

## 3. Lessons this lane paid for today (read before touching anything here)

- **Round 1 REVISE was CORRECT**: a hand-rolled 9-column "snapshot" of an
  agent_definitions row would have dropped topics/capabilities/image fields
  from any restore. The sanctioned mechanism is `snapshot_agent(type, reason)`
  → `agent_definitions_backup`, verified to hold the PRE-change config, found
  again by `snapshot_reason` + `snapshot_taken_at DESC` (never id/created_at —
  they're copied from the source row).
- `agent_definitions.agent_category` is CHECK-constrained to
  strategist/executor/analyst/integrator/coordinator/specialist — the
  unconstrained `category` column is where 'orchestrator' lives.
- Migration numbers moved 411→421 under this session and 423 was taken
  concurrently — numbers are NEVER reserved; re-list at write time, resolve
  by filename.
- The provenance log line rotates out within ~2h on the chassis — the binary
  probe (with the random-hex control) is the durable check.

## 4. Work list (after the canary is proven)

1. **Phase 3 — ZIP deliverable** (PLAN Part 2e): `zip_deliverable_action.go`,
   `archive/zip` first use in the repo, stream never truncate, presigned URL.
   Council run + register entry. The b2worker canary tree is a ready test
   subject.
2. **Phase 4 — handover state** (`sites.handed_over_at`, gates ONLY editor
   access; `platform/mailer` exists and is sanctioned).
3. **Phase 5 — magic-link customer auth** (clone `sitefacts.go`; add
   `"customer"` to `humanLockSources`).
4. **Phase 6 — the editor** (extract `HandleUpdateComponent`'s tx; box
   service port 8083; **site_id always from the session — the cross-tenant
   probe is the acceptance**).
5. **cfpages arming** (owner-gated): token into `personae-platform-secrets`
   + spawner env injection for site-publisher + build the Direct Upload
   client verify-as-you-go, **writing up the protocol decisions FIRST**
   (architecture seat's standing obligation, recorded in DGH-008).
6. Advisory follow-ups parked in NOTES: shared fetch/verify helper,
   contentTypeFor dedup — Phase 3+ candidates, not now.

## 5. Falsifiers (re-check before trusting this file)

A newer handoff here; `SELECT count(*) FROM sites WHERE publish_target IS NOT
NULL` (1 = canary only, as written); the chassis stamp (image label per
service — a NEWER roll than v1.0.1303 changes nothing here as long as
`71e4d9736` remains an ancestor); whether the canary's published_hash landed
(§1 supersedes this file's "in flight" the moment it does); the council
coverage report (`098`) crediting the two `Council-Submitted:` commits.
