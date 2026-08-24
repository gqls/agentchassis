# HANDOFF — noted.co.uk, continue here

**Written 2026-08-24 ~18:45 UTC. Supersedes `HANDOFF_2026-08-19_continue_here.md`
(read it for the CSS-incident mechanism, §3 there — still the sharpest account,
and bug 198's guards were built by that lane since: 198 is CLOSED, but its
residuals are that lane's, not ours).** Standalone otherwise. Then:
`NOTES_noted_rebuild.md` (08-24 entries are the live ones),
`README_where_we_are.md`, `PLAN_2026-08-24_media_pasteboard.md`,
`RUNBOOK_noted_rebuild.md` (engine deploy + tool update recipes added 08-24).

## 1. What this is

noted.co.uk — notes with text, photos, GIFs, video and audio — LIVE as a
framework site at the apex since 08-16. Engine (accounts/notes/media API,
hand-written Go + Postgres) on the webdesign box, loopback behind nginx +
cloudflared. **Nothing is hand-built as a page; the engine is hand-written by
design (the framework has no end-user accounts).**

## 2. CURRENT STATE (all verified at the artefact, 2026-08-24 evening)

- **Pasteboard stage 1 LIVE + smoke-proven 15/15**: any media in a note —
  paste/drop/pick in `/tools/write/`, images inline, video/audio players,
  remove-with-confirm (leaves only on a 2xx), storage meter, pre-flight
  size/type refusals. Engine: `video` kind, `DELETE /api/media/{id}`, Range
  serving (seeking works), unified per-note `media` array, `/api/me.max_upload`.
  Old engine binary at `/root/noted-engine.pre-20260824` on the box.
- The editor's honesty contract now has FOUR clauses (tool-doc in
  `editor_tool/noted-write.html`), each mutation-verified in
  `editor_tool/test_editor_degraded.py` (38 checks). **Re-run after ANY edit.**
- Council on the CTA override: **APPROVED 08-18** (`89f3331e…`) — commit
  `229e14e74` auto-credits via its `Council-Submitted:` trailer. 5 advisories
  sit unread-in-detail in the full report (diagnosis_artifacts, that corr).
- Contact email, privacy copy (22/22 verbatim), CSS v23 — all as the 08-19
  handoff left them.

## 3. NEXT: stage 2 — the pasteboard proper (PLAN_2026-08-24 §Stage 2)

`notes.layout` JSONB (idempotent ADD COLUMN; client-owned, versioned shape
`{v:1, items:[{type, media_id?, text?, x,y,w,h,z}]}`), saved through the
existing note save so the Saved contract covers it free. Editor grows a board
view (drag/resize/z-order); linear editor stays the fallback; absence of
`layout` = today's behaviour. Stage 3 (edit in place) after.

## 4. OPEN THREADS, in order

| # | thread | state |
|---|---|---|
| 1 | **Stage 2 build** | Not started; design pinned in the PLAN |
| 2 | **Quota sizing** (50 MB/account, 25 MB/file) | Owner's call — env-tunable (`NOTED_MEDIA_QUOTA_MB`, `NOTED_MAX_UPLOAD_MB`); the account quota protects the shared 50 GB disk |
| 3 | **Account deletion** | Still absent; privacy copy promises it; smoke tests leave throwaway accounts; more urgent now accounts hold real files |
| 4 | **CTA-override advisories** (5, non-gating) | Undispositioned — read the full report once, record dispositions |
| 5 | **Experience patterns at `proposed`** | Unchanged from 08-19 §4 |
| 6 | **Three `detected` items (08-10)** | Still inert, still known |

## 5. TRAPS — the ones this lane keeps proving

- **Verify at the artefact, never the status** — this session again: rerenders
  for this site ran at 13:57 that PREDATED the 15:45 regeneration; only the
  box-file grep says what serves.
- **This session's classifier REFUSES box writes** (scp/install/restart) — the
  OWNER runs the two RUNBOOK blocks (`!` prefix); reads are fine. Checksum
  before the overwrite.
- **A regeneration does not enqueue the page build** — hand-file the
  page_rerender; **RESET the `page_id` COLUMN** (the handler ignores your
  spec); claim gates filter `status IN ('triaged','approved')`.
- **An EMPTY poll is "instrument down", not "unchanged"** — the 08-24 watcher
  lost stderr and read 13 empty statuses during a transient exec failure. Keep
  stderr on watchers.
- **kcat publishes need a receipt** — use `kafka_publish_checked`
  (077 does; 076 predates it).
- Everything in the 08-19 handoff §5 still applies (sitesync `--delete`,
  fd-3 loops, byte-vs-char, nginx `.bak`, cloudflared restart-not-HUP,
  `vm-sites` load-bearing).

## 6. COMMANDS

All in `RUNBOOK_noted_rebuild.md` (engine build/test/deploy; tool update via
`scripts/initial_messages/140_tool_suggester/077_update_noted_write_tool.sh`;
live smoke `editor_tool/smoke_live_editor.py https://noted.co.uk` — 15 checks
incl. the media round-trip; theme-health one-liner in the 08-19 handoff §6).
