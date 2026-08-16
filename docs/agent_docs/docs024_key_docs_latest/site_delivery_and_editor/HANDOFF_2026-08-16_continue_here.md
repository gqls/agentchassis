# HANDOFF 2026-08-16 — continue here: Phase 2 (publish seam) is COMPLETE AND PROVEN IN PRODUCTION; next is Phase 3 (ZIP deliverable)

**Start here cold.** Read order: this file → `PLAN_2026-08-14_site_delivery_and_editor.md`
(the owner-approved design, Phases 2–6) → NOTES tail (the 2026-08-15/16 entries,
newest at the bottom — they carry every misstep and its fix).

## 0. State in one paragraph

**Phase 2 is DONE.** The publish seam is live and proven end to end on
**v1.0.1304**: `platform/publish` + `publish_site` + the reconciler
(`site-publish-reconciler` → `publish-reconciler` → spawned `site-publisher`),
council **APPROVED** (corr `21aba3f5`), migrations 412 + 423 applied, seed 422
hand-applied. The canary **noted.co.uk → noted.ugg2.com passed on 2026-08-16
16:01Z**: 8/8 objects copied, served `index.html` sha256 byte-identical to the
origin hash captured before any publish existed, `published_hash` written only
after that acceptance; the next tick returned `no drift` and published nothing.
The canary is **left armed** as continuous proof (one no-op tick/hour). Exactly
1 site is opted in; every other site is `publish_target` NULL = OFF.
**Phases 3–6 are unstarted.** cfpages remains deliberately unarmed.

## 1. NEXT: Phase 3 — the ZIP deliverable (PLAN Part 2e)

New `zip_deliverable_action.go`: ListObjects prefix → stream through
`archive/zip` (first use in the repo) → Upload under `deliverables/<domain>/`
→ presigned URL. Council run + register entry, per the PLAN roll-up.

Carry these forward — they are paid-for, not theoretical:
- ⚠ **Do NOT copy b2worker's whole-buffer upload pattern for the ZIP's own
  output.** b2worker buffers each small site file because B2's S3 gateway
  411s a non-seekable body; a whole site ZIP is a different size class.
  Stream with a known length, or multipart. A truncated ZIP is a silent
  contractual failure (the PLAN's own ranked risk 3).
- The action must run in a **spawned storage-enabled pod** for the same
  reason `publish_site` does (chassis has no B2 credentials, owner ruling
  2026-08-08). `site-publisher` is the allow-listed type already; adding a
  new type means a spawner allow-list change.
- Acceptance is the artefact: `unzip -l` count == object count, extracted
  index.html sha == the B2 object, presigned 200 in-expiry / 403 after.
- Reuse `publish.S3Source`/`ObjectStore` for listing+reading — it exists,
  is tested, and already strips the `<domain>/` prefix.

Then Phases 4 (handover state), 5 (magic-link customer auth), 6 (the editor —
**site_id always from the session; the cross-tenant probe is the acceptance**),
and cfpages arming (owner token + **write up the Direct Upload protocol
decisions FIRST** — the architecture seat's standing obligation).

## 1a. If you need to re-verify or re-run the canary

```sql
-- state
SELECT domain, publish_target, publish_project, published_hash, published_at
  FROM sites WHERE domain='noted.co.uk';
-- force a republish (proves the whole path again)
UPDATE sites SET published_hash=NULL WHERE domain='noted.co.uk';
-- turn it off entirely
UPDATE sites SET publish_target=NULL WHERE domain='noted.co.uk';
```
Reconciler interval is **3600s** (raised from 600s on 2026-08-16: every tick
spawns a pod even for a no-op, because the drift check needs bucket
credentials and cannot run on the chassis). The original acceptance recipe,
kept because it is still how you check a publish:

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
