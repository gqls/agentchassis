# PLAN — noted.co.uk media + pasteboard (owner ask, 2026-08-22)

The owner's words: *"I want to be able to save images, videos, gifs, audio etc
too, all in one note. Can we have a sort of pasteboard so you can paste any of
them onto it and move them around and even edit them. That would be amazing but
we can get there in stages."*

The last clause is the licence for this plan's shape: plumbing first, then the
board, then editing. Each stage is independently shippable and independently
verifiable at the artefact.

## Where we start from (measured 2026-08-24, at the source in `box/noted-engine/`)

- The engine already stores media: `media` table (`kind IN ('audio','image')`,
  bytes in Postgres, per-account quota enforced in-transaction, default 50 MB,
  `NOTED_MAX_UPLOAD_MB` 25). `POST /api/notes/{id}/media?kind=`,
  `GET /api/media/{id}` (account-scoped), legacy-backup import fills it.
- **No `video` kind, no media delete endpoint, no Range support** (whole-body
  write — seeking in a `<video>` would not work).
- The editor (`/tools/write/`, framework tool `tool-write`) is title + textarea
  only; its tool-doc says so honestly. The engine's media surface has **zero UI**
  — imported legacy photos/recordings are invisible to their owners today.
- The editor's load-bearing contract (mutation-verified): "Saved ✓" only on a
  2xx; failure loud, text untouched; beforeunload while dirty.

## Stage 1 — any media in a note, end to end (THIS stage now)

**Engine** (hand-written, on the box — not council scope; source stays versioned
in `box/noted-engine/`):

1. `kind` widens to `('audio','image','video')`. The CHECK lives in an
   already-created table, so `CREATE TABLE IF NOT EXISTS` cannot change it: the
   idempotent migration is `DROP CONSTRAINT IF EXISTS media_kind_check` + re-add,
   in schema.sql, run at every startup like the rest.
2. `DELETE /api/media/{id}` — account-scoped in the SQL like every other query;
   the existing trigger frees the quota on DELETE already.
3. `GET /api/media/{id}` serves via `http.ServeContent` (Range → video/audio
   seeking works). `nosniff` + `inline` stay.
4. `GET /api/me` also reports `max_upload`, so the client can pre-check size
   without hardcoding a number that lives in the engine's environment.
5. List payload: each note gains a unified `media` array (all kinds, upload
   order) — the shape a pasteboard needs. The grouped `audio`/`images` fields
   stay (cheap; same query).
6. Tests for each in `engine_test.go` (real Postgres via a throwaway
   `postgres:16` container; the suite skips loudly without one).

**Editor** (framework tool page; update via `create_tool_component` with
`replace_existing` (TL-047) + hand-filed page build item with the `page_id`
COLUMN set — regeneration does not enqueue the build):

- Paste (clipboard), drag-drop, and a picker button. Kind from MIME:
  `image/*` (GIF included) → image, `video/*` → video, `audio/*` → audio;
  anything else refused with a plain sentence.
- A media strip on the open note: images/GIFs as thumbnails, video/audio as
  players (`src="/api/media/{id}"` — same origin, cookie auth, Range works).
  Remove per item (confirm; gone from the strip only on a 2xx).
- Upload honesty is the SAME contract family: a new note is saved first (its id
  is needed), each item shows "Uploading…" → stored only on 2xx; a failed
  upload keeps the bytes in memory with a Try again, and never claims storage.
  A media upload does NOT mark the text dirty (the rows are server-side);
  beforeunload extends to in-flight/failed uploads.
- Storage meter from `/api/me` (`media_bytes`/`media_quota`), refreshed after
  upload/delete; the engine's quota refusal shown verbatim.
- `test_editor_degraded.py` gains media cases, mutation-verified like the
  original three; `smoke_live_editor.py` gains a real upload/serve/delete pass
  (file picker via `set_input_files` — a synthetic paste is not provable).
- Existing `#nw-*` ids are load-bearing (experience-pattern selectors): add,
  never rename.

**Immediate product win**: legacy-imported photos and recordings become visible
to their owners for the first time.

## Stage 2 — the pasteboard (arrangement)

- `notes.layout` JSONB (idempotent `ADD COLUMN IF NOT EXISTS`), client-owned,
  versioned shape: `{v:1, items:[{type:text|media, media_id?, text?, x,y,w,h,z}]}`.
  Saved through the existing note save (4 MB body cap is ample), so the Saved
  contract covers layout for free.
- Editor grows a board view: absolute-positioned draggable/resizable items —
  media items reference stage-1 rows; text blocks live in the layout itself.
  The linear editor remains the fallback and `content` remains the plain-text
  truth for notes that never touch the board (absence of `layout` = today's
  behaviour, per the estate's opt-in rule of thumb).

## Stage 3 — editing in place

- Client-side image crop/rotate (canvas → re-upload as a NEW media row, then
  delete the old on 2xx — never destructive in one step), captions, z-order
  controls; audio/video trim only if the owner still wants it after using 1+2.

## Decisions and their reasons

- **Media stays in Postgres under the quota.** schema.sql's own header states
  the revisit point (a few GB total) and the migration door (`storage_key`
  beside `bytes`). Video brings that day closer; the quota is the safety valve
  meanwhile. **Raising the 50 MB default is the owner's call** — it is the
  control protecting the shared 50 GB disk under webdesign.uk's shopfront.
- **GIFs need no kind of their own** — `image/gif` animates in `<img>`.
- **Immediate upload, not upload-on-save**: held blobs die with the tab, and
  the engine's quota refusal should arrive while the person is still looking at
  the file, not at save time. The cost — an empty note row when someone pastes
  before typing — is honest and small.
- **25 MB per file / 50 MB per account will pinch for video.** Stage 1 ships
  under today's limits and surfaces both numbers in the UI; sizing is flagged
  to the owner alongside the missing account-deletion mechanism (open thread 3,
  HANDOFF 2026-08-19 §4).
