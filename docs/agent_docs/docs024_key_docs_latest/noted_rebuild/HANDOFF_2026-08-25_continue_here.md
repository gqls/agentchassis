# HANDOFF — noted.co.uk, continue here

**Written 2026-08-25 evening. Supersedes `HANDOFF_2026-08-24_continue_here.md`**
(its §5 traps all still hold — read them; the CSS-incident mechanism stays in
the 08-19 handoff §3). Then: `NOTES_noted_rebuild.md` (08-25 entries),
`PLAN_2026-08-24_media_pasteboard.md` (+ its 08-25 owner-rulings addendum),
`PLAN_2026-08-25_account_deletion.md`, `SUMMARY_2026-08-25_noted_rebuild.md`,
`RUNBOOK_noted_rebuild.md`.

## CURRENT STATE (verified at the artefact, 2026-08-25 evening)

- **Pasteboard stages 1 + 1b + 2 LIVE, smoke 15/15 on the apex**: media in
  notes; media bytes in B2 (`personae-noted-media`, private, bucket-scoped key;
  engine refuses a broader key; startup line `media storage: B2 bucket …` is
  the greppable truth); the board (List|Board toggle, pointer events, 44px
  handles, fractional-of-width coords, arrangement rides Save).
- **B2 API is v4** — the live service REFUSES v2/v3 `b2_authorize_account`;
  the client + stub + live-gated `TestB2LiveRoundTrip` all speak v4. Engine
  suite 13/13 + live test; editor harness 51 checks; 8 guards mutation-proven
  across the two days.
- Existing pre-B2 media rows still serve from Postgres (`storage_key IS NULL`
  path) — no drain needed; optional later.
- Old binaries on the box: `/root/noted-engine.pre-20260824`,
  `/root/noted-engine.pre-20260825-b2`.
- Migration `608` (ledger backfill of the CTA-override config write) is
  committed and PENDING — idempotent no-op; the next runner sweep applies it.

## OPEN, in order

1. **Account deletion — build on the owner's word**
   (`PLAN_2026-08-25_account_deletion.md`): his two decisions are immediate-vs-
   grace (immediate recommended) and the privacy backup sentence.
2. **Stage 3** (edit in place: crop/rotate/captions) — after the owner has
   lived with the board.
3. **Paid storage tier** — the B2 move was for this; quota env
   `NOTED_MEDIA_QUOTA_MB` is the lever.
4. CTA-override follow-up recorded 08-25: a REFUSED `header_cta_url` override
   should raise an owner-visible work item, not a Warn log (platform change,
   own council round).
5. Experience patterns at `proposed`; three inert `detected` items (08-10) —
   unchanged.

## NEW TRAPS from these two days (older ones: 08-24 handoff §5)

- **B2 docs of record show retired API versions** — probe the real service
  with the real credential before trusting any stub you wrote (v2 AND v3
  authorize are refused; v4 nests everything under `apiInfo.storageApi`).
- **A box check inside the sitesync tick reads yesterday's world** — deploy
  commits to git; the box follows up to 5 min later.
- **A directory-pathspec commit silently excludes untracked files** — two
  commits described b2.go while it sat untracked (WRONG_CALLS 08-25; the tell
  is a diffstat that cannot contain the described change).
- **The B2 secret lives ONLY in `/etc/noted/noted.env` (box) now** — it never
  entered the repo or the conversation; the scratchpad copy dies with the
  session. Losing the box means cutting a NEW key (the CLI has account
  authority), not recovering the old one.

## COMMANDS

RUNBOOK_noted_rebuild.md has engine build/test/deploy (box writes may need the
owner's hand — the classifier allowed the install once on 08-25 after his
explicit go-ahead, refused it cold on 08-24) and the tool-update recipe (077 →
hand-filed page_rerender with the `page_id` COLUMN reset → live smoke:
`editor_tool/smoke_live_editor.py https://noted.co.uk`).
