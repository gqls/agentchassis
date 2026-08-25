# HANDOFF — noted.co.uk, continue here

**Written 2026-08-25 late evening (second handoff of the day — the morning one
is `HANDOFF_2026-08-25_continue_here.md`; its traps all stand). Standalone:
start from this file.** Then: `NOTES_noted_rebuild.md` (the three 08-25
sessions are the live entries), `README_where_we_are.md` (owner's plain-prose
log), `PLAN_2026-08-24_media_pasteboard.md` (+ both 08-25 rulings addenda),
`PLAN_2026-08-25_account_deletion.md`, `SUMMARY_2026-08-25_noted_rebuild.md`,
`RUNBOOK_noted_rebuild.md` / `RUNBOOK_cutover.md`.

## 1. What this is

noted.co.uk — a notes product (text, photos, GIFs, video, audio) — LIVE at the
apex as a framework site since 08-16. Engine (accounts/notes/media API,
hand-written Go + Postgres, by design: the framework has no end-user accounts)
on the webdesign box, loopback behind nginx + cloudflared. Media bytes in
Backblaze B2. Every page framework-owned; every promise on the site has a
mechanism behind it as of tonight.

## 2. CURRENT STATE — everything below verified at the artefact, 2026-08-25

- **The full pasteboard programme is LIVE**: media in notes (paste/drop/pick;
  players inline) · the mobile-first BOARD (pointer events, 44px handles,
  fractional-of-width coordinates, arrangement rides Save) · captions (on the
  media ROW) · image rotate/crop editing that NEVER mutates in place (copy
  uploads first, original removed only on the 2xx; GIFs refused with the
  reason; board items retarget preserving position).
- **Immediate account deletion LIVE**: password re-typed, B2 objects deleted
  BEFORE the row cascade, all-or-nothing, goodbye only after everything is
  gone. Exercised in production five times on 08-25 (the smoke's self-cleanup
  + draining the four historical throwaways). **Exactly ONE account exists:
  the owner's.** The smoke now deletes its own throwaway every run.
- **Media storage = B2** bucket `personae-noted-media` (private; bucket-scoped
  key with listFiles/readFiles/writeFiles/deleteFiles only; the engine REFUSES
  a broader key). Startup line `media storage: B2 bucket personae-noted-media`
  is the greppable truth. **B2 API is v4 — the live service refuses v2/v3
  `b2_authorize_account`.** Quota 50 MB/account, 25 MB/file (env:
  `NOTED_MEDIA_QUOTA_MB`, `NOTED_MAX_UPLOAD_MB`). Zero media has ever sat in
  the pg dumps.
- **Privacy page carries the owner-approved backup sentence, 26/26 sentences
  verbatim** (25 measured on the live page; the email sentence measured at the
  BOX — the edge obfuscator rewrites addresses, the documented instrument
  case). The 30-day figure is **evidence_base fact #1** (7 bans intact).
- **Proof state**: engine suite 16/16 + live-gated B2 round-trip; editor
  harness 61 checks (cases A–O incl. a 390px touch viewport); live smoke
  **17/17 self-cleaning**; ~14 guards mutation-verified red across 08-24/25.
- Post-chassis-roll control (08-25 late): apex 200, engine health ok, editor
  page 200, shopfront 200. Nothing of this lane rides the cluster roll.
- Rollback binaries on the box: `/root/noted-engine.pre-20260824` (pre-media),
  `…pre-20260825-b2` (stage-1, restored from git after a re-run clobbered it),
  `…pre-20260825-deletion` (B2+board, pre-deletion).

## 3. OPEN — the OWNER'S decisions (explained in chat 08-25 late)

1. **Paid storage tier — when, and its shape.** Copy is APPROVED and recorded
   (PLAN_2026-08-24, final section) but HELD: shipping it before a purchase
   mechanism exists would promise a door that isn't there (161's class). To
   build it he must pick: price and size of the step (the quota env var is the
   lever), one-off top-up vs subscription (subscription touches the delivery
   lane's Stripe groundwork — memory: £10/mo = Payment Links, never the
   PAY-007 scaffold), and whether to wait for real users first.
2. **The "editor note" surface** — his vision (recorded verbatim in
   PLAN_2026-08-24's 08-25 addendum): a note type that becomes an editor —
   GIF-making, video editing, server-side when bandwidth allows. Decisions:
   now vs after the board beds in; client-side-only next steps (cheap) vs
   server-side processing (compute on the shared box, or a new box — cost).
3. **Mail routing for `noted@contactforsales.com`** (open since the 08-19
   handoff): the site and the privacy page NAME it; if mail there reaches
   nobody, the contact promise is hollow. One test email settles it.
4. **The backup-drill identity** (RUNBOOK, restore drill section): the age
   key still lives ONLY on the operator workstation. Owner: store it in the
   password manager; we re-run the drill FROM that copy, then delete the
   workstation file.
5. **Inviting real users** — every promise now has a mechanism; launch timing
   is product judgement, his.

## 4. OPEN — ours, not his

- CTA-override follow-up (08-25 disposition): a REFUSED `header_cta_url`
  override logs a Warn only — should raise an owner-visible work item.
  Platform change, own council round when picked up.
- Migration `608` (ledger backfill, idempotent no-op) — pending the runner
  sweep; nothing to do, just don't be surprised by it in a dry run.
- Experience patterns still `proposed`; three inert `detected` items (08-10).
- Stage-3 continuation beyond the seed (see owner decision 2).

## 5. TRAPS — the 08-25 harvest (earlier ones: 08-24 + 08-19 handoffs §5/§3)

- **B2 speaks v4 only** — probe the real service with the real credential
  before trusting any stub you wrote; v4 nests under `apiInfo.storageApi`.
- **A box check inside the sitesync tick reads yesterday's world** (5-min
  cadence; bitten twice — check AFTER the tick, or at the repo).
- **A `&&`-chain's steps BEFORE the guard run on every retry** — the install
  recipe's backup step clobbered its own rollback copy on a re-run; recipe now
  checks `.new` exists FIRST. General shape, worth carrying.
- **A directory-pathspec commit silently excludes untracked files** — two
  commits described b2.go while it sat untracked; the tell is a diffstat that
  cannot contain the described change (WRONG_CALLS 08-25).
- **One-shot migration scripts read as reusable** — apply/embed (08-12)
  cannot re-run (gone clause; presence≠currency). Future privacy-copy edits:
  `refresh_privacy_copy_from_draft.py` (re-runnable, self-verifying) → 074b.
- **Instrument-before-subject, three in one chain**: a figure regex that only
  knew marketing shapes read "30 days" as "none"; a verify probe straddling a
  hard-wrap read a present sentence as absent (normalise BOTH sides); urllib's
  default UA drew a Cloudflare 403 that read as a page failure (fetch with
  curl). And the standing one: **a live grep for an EMAIL tests the edge
  obfuscator, not the page** — measure at the box.
- **Box writes**: the classifier refused them cold on 08-24, allowed on 08-25
  after the owner's explicit go-ahead. Expect either; the RUNBOOK blocks are
  written to hand over (`!` prefix). Checksum before any overwrite.
- **The B2 secret lives ONLY in `/etc/noted/noted.env`** (box). Losing the box
  = cut a NEW key (CLI has account authority); the old one is unrecoverable.

## 6. COMMANDS

All in `RUNBOOK_noted_rebuild.md`: engine build/test/deploy (tests need the
throwaway postgres:16 container; deploy = two blocks, chmod-first);
editor update = `scripts/initial_messages/140_tool_suggester/077_update_noted_write_tool.sh`
(replace_existing; does NOT enqueue the build) → hand-filed `page_rerender`
(temp-table copy; RESET the `page_id` COLUMN; claim gates filter
`status IN ('triaged','approved')`) → live smoke
`/home/ant/.venvs/vonc_pw/bin/python docs/agent_docs/docs024_key_docs_latest/noted_rebuild/editor_tool/smoke_live_editor.py https://noted.co.uk`
(17 checks, self-cleaning). Privacy copy edits: edit the DRAFT →
`COPY_2026-08-12_privacy_check.py` → `refresh_privacy_copy_from_draft.py` →
`074b_section_editor_noted_privacy_copy.sh` → verify 26/26 (email sentence at
the box). Editor harness after ANY noted-write.html edit:
`/home/ant/.venvs/vonc_pw/bin/python docs/agent_docs/docs024_key_docs_latest/noted_rebuild/editor_tool/test_editor_degraded.py`.
