# PLAN 2026-08-26 — large uploads (the 25 MB blocker), gating the paid tier

**Owner ruling (2026-08-26 evening):** the paid tier does not launch until the 25 MB
file-size blocker is fixed; this build is next. Tier copy stays HELD meanwhile.

## Why 25 MB is not a number we can just raise

Three limits stack, and only two are ours:

1. **Engine** buffers the whole upload in memory (`b2.go Upload(..., data []byte)`,
   single `b2_upload_file`; `server.go uploadMedia` behind `MaxBytesReader`). Ours.
2. **nginx** `client_max_body_size 30m` (`box/noted.co.uk.nginx:20`,
   `proxy_request_buffering off` — already streams). Ours.
3. **Cloudflare tunnel**: ~100 MB request-body cap on the current plan. NOT ours to
   raise. This is the binding constraint: any single-request upload tops out under
   100 MB no matter what the engine allows.

So gigabyte files mean a **chunked upload protocol**: the editor slices the file, each
part is one request under the Cloudflare cap, and the engine streams parts into B2's
large-file API — never holding the whole file, never landing it on the box disk.

## Design decisions (each with its reason)

- **Part size 64 MiB** — comfortable margin under Cloudflare's cap; B2's minimum part
  is 5 MB, max 10,000 parts, so 64 MiB parts reach ~640 GB/file; B2 large files go to
  10 TB if we ever raise the part size.
- **The small path (≤25 MB single POST) stays byte-identical** — proven, and free
  accounts keep exactly today's behaviour. The chunked path only opens where the
  account's effective max-upload allows it.
- **Per-account overrides, not a global bump**: `accounts.media_quota_override_bytes`
  and `accounts.max_upload_override_bytes`, both NULL = today's env defaults. This is
  also the paid-tier lever (1 TB quota = set the override), so the tier needs no
  further engine schema when its time comes. `/api/me` already reports
  `media_quota`/`max_upload`, so the editor UI adapts with no extra plumbing.
- **Quota is reserved at `begin`, charged at `finish`, released on abort/reap** — a
  declared 900 GB upload on a 1 TB quota must block a concurrent 200 GB `begin`.
  Reservations live in a `pending_uploads` table and count toward usage in the quota
  check.
- **Honesty contract unchanged and extended**: the editor shows an item as STORED only
  after `finish` returns 2xx (same as today's rule for the single POST); a failed part
  retries; `beforeunload` already guards in-flight uploads.
- **Reaper for abandoned uploads** (startup + daily): cancel unfinished B2 large files
  under our prefix older than 24 h AND clear stale `pending_uploads`. **B2 is the
  source of truth** (`b2_list_unfinished_large_files`), not our table — the
  thunder-reaper lesson: a tick that only reads its own bookkeeping has never reaped.
- **Memory bound**: one 64 MiB buffer per in-flight part, global semaphore of 2
  concurrent part-uploads → worst ~128 MiB on the shared box.
- **Sequential parts in the editor v1** (simple, clear progress = parts done/total,
  per-part retry ×3). Resume-across-page-reload is NOT in v1 — stated, not implied.

## The pieces

1. **b2.go**: `StartLargeFile`, `GetUploadPartURL`, `UploadPart`, `FinishLargeFile`,
   `CancelLargeFile`, `ListUnfinishedLargeFiles` — same stdlib-only v4 style, same
   per-call URL-fetch discipline the existing `Upload` documents.
2. **schema**: `pending_uploads` (id, account, note, b2_file_id, declared bytes,
   content_type, parts recorded as (n, size, sha1), created_at) + the two override
   columns on `accounts`. Live DB gets a hand-run ALTER (the box's own postgres —
   engine schema is not the fleet's migration runner's).
3. **server.go**: `POST /api/notes/{id}/media/uploads` (begin; validates type/size
   against effective limits + quota-with-reservations) · `PUT
   /api/uploads/{id}/parts/{n}` (bounded read, sha1, B2 part, idempotent per n) ·
   `POST /api/uploads/{id}/finish` (all parts present, sizes sum to declared,
   `b2_finish_large_file`, media row inserted, reservation → usage) · `DELETE
   /api/uploads/{id}` (cancel) · reaper goroutine.
4. **Editor** (`editor_tool/noted-write.html`): files over the small threshold take
   the chunked path; strip/meter behaviour identical to today from the user's side.
   Harness gains cases: chunked happy path, part failure → retry → success, finish
   refused on missing part, abort releases quota, oversize refused at begin — each
   mutation-verified per the lane's standard.
5. **nginx**: `client_max_body_size 100m` (covers 64 MiB parts + headers). Owner-run
   box block, RUNBOOK gets the command pair (edit + `nginx -t && reload`).
6. **Env defaults unchanged** (`NOTED_MAX_UPLOAD_MB=25`, `NOTED_MEDIA_QUOTA_MB=50`);
   nothing changes for existing accounts until an override is set.

## Order of work

b2.go + tests (incl. live-gated 2-part round-trip, ~10 MB — B2's 5 MB part minimum
sets the floor) → schema + store → server endpoints + engine suite → editor + harness
→ RUNBOOK deploy blocks (binary install, ALTER, nginx) → owner runs the box blocks →
live smoke + a real >30 MB upload through the wire as the acceptance proof (30 MB
beats nginx's old cap, proving the whole stack, while staying polite to the tunnel).

## Explicitly out of scope here

Stripe/payment linkage, prices, VAT, the unsubscribe policy (owner discussion open in
chat, 2026-08-26); resume-across-reload; parallel part upload; raising Cloudflare's
cap (plan-level, owner's account).
