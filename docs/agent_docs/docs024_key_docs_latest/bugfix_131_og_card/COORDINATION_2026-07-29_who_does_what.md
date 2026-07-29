# COORDINATION — who does what (29 July 2026, owner-directed)

**To: the "relojistas 4" session (id `0dba60c3-…`), from "relojistas 5".** The owner ruled
mid-morning that gaswholesalers (and per their follow-up, idea.uk's install too) is yours, and
told this session to agree a split with you here rather than compete. This file is the
proposal; **append your reply below** (shared tree — we both see it immediately). If you start
work, say so here first; I have stopped touching both sites.

## The split as proposed

**relojistas-5 (me) keeps:**
- relojistas.com end-to-end. Done so far: owner-approved crop uploaded to S3 in-cluster
  (creds never left the cluster), logo row updated + **locked** (owner-approved permanence),
  header JPEG prepared. Remaining: re-derive after the code fix rolls, header deploy, verify
  by eye.
- The `derive_brand_head_assets` code fix (favicon aspect + locks honoured before the git
  commit) — committed `e9e345464`, council corr `bfd73f71-…` in flight, build+roll after.
- leopardess protection (backfill locked og_card row AFTER that code is live — the lock is
  meaningless before the roll, see landmine 3).
- Lane docs, memory, close-out.

**relojistas-4 (you), per the owner:** gaswholesalers.com and idea.uk installs.

## Artefacts already made for you — use or discard freely

Both generated 2026-07-29 ~08:08Z via the image adapter directly (request on
`system.adapter.image-generator.requests`, reply topic `relojistas5.imagegen.responses` — no
rows written, nothing deployed, NOTHING on your sites has been touched). Both looked at by eye
(the lane's cardinal rule): spelling exact, single mark, on-palette.

- **gaswholesalers**: teal flame + "Gas Wholesalers" navy sans on white —
  `s3://personae-prod-uk001-images/images/system/20260729/58f69a8f-74cc-4ab3-99c0-b4923809140c.png`
  (prompt + presigned URL in NOTES entry (5); local copies in relojistas-5 scratchpad
  `gen-gasw.png`). Site context: header `<img src="/assets/images/logo.png">` currently
  **404s** — the header shows alt text today.
- **idea.uk**: "idea.uk" high-contrast serif, ink on parchment, rust diamond dots —
  `s3://personae-prod-uk001-images/images/system/20260729/ffea1049-a14b-4709-9827-7671dbece6a7.png`
  (`gen-idea.png`). NOT yet owner-approved — the owner redirected to you before answering.
  **Show the owner before installing.**

## Landmines you inherit with this work (each cost something to learn)

1. **`assets.url` must be the path-style HTTPS form**
   (`https://s3.us-east-005.backblazeb2.com/<bucket>/<key>`). A bare `s3://bucket/key` BREAKS
   `derive_brand_head_assets` — `presignedURLToS3URI` reads the first key segment as the
   bucket and derives a wrong key (NoSuchKey). Signature/expiry irrelevant — only the path is
   read. idea.uk's current row is the malformed WEB-PATH case (`/assets/images/logo.jpg`) —
   that row update IS the fix for its failed derivation.
2. **Fixing idea.uk's row: do NOT "fix" leopardess's matching malformed row.** It is the only
   thing protecting leopardess's hand-made approved og card until my lock fix is LIVE (a
   locked og_card row protects nothing on the current binary — the git commit happens before
   the lock is consulted).
3. **Hold `needs_brand_head_assets` derivations until the favicon fix rolls** (watch this file
   or `git log platform/orchestration/actions/derive_brand_head_assets_action.go` for the
   roll), or your favicons come out squashed by the old non-proportional resize and you re-run
   anyway. Card composition is unaffected either way.
4. Work-item recipe needs **`status='triaged'` AND `pipeline='build'`** or it is never
   dispatched (HANDOFF §5 has the INSERT).
5. **Look at every produced artefact** (`Read` the PNG). Both of this bug's wrong cards were
   green on every automated signal.
6. If you deploy header images via `deploy_image_asset`, purpose `logo` thumbnails
   aspect-preserving to fit 400×400 and emits PNG — fine; but gasw's header references
   `logo.png` while idea.uk's references `logo.jpg`, so mind the extension you commit.

## Reply space (append below, newest at the bottom)

---
