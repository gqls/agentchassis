# HANDOFF 2026-08-17 — start here for Phase 3: the ZIP deliverable — SUPERSEDES HANDOFF_2026-08-16

> **⚠ PHASE 3 IS COMPLETE — code committed (`e1a7f1935`), council APPROVED
> round 1 (`4cc887b9`), seed 459 APPLIED, and the FULL §3 acceptance PASSED in
> production 2026-08-18 on v1.0.1308 (NOTES 2026-08-18 entry; register DGH-011
> holds the evidence). Do NOT re-implement and do NOT re-run acceptance.**
> One acceptance nuance for future phases: B2 refuses expired presigns with
> 401 `UnauthorizedAccess` (+ expiry message), not AWS's 403. NEXT PHASE = 4
> (handover + delivery email, `PLAN_2026-08-14` Phase 4 + the email-links list
> in `PLAN_2026-08-17` decision 3); dispatch the ZIP via
> `zip-deliverable-dispatch` with `{domain}` — recipes in 459's header.

**Start here cold.** Read order: this file →
`PLAN_2026-08-17_delivery_architecture_decisions.md` (the owner's decided
architecture — read at least OWNER DECISIONS + Build order) →
`PLAN_2026-08-14_site_delivery_and_editor.md` Part 2e Phase 3 (the mechanics
this phase executes) → NOTES tail (2026-08-16/17 entries).

## 0. State in one paragraph

Phase 2 (the publish seam) is COMPLETE and proven in production on
v1.0.1304+: `platform/publish` + `publish_site` + the hourly reconciler, with
the canary noted.co.uk → noted.ugg2.com live-verified both ways (publish on
drift, no-op without). The delivery ARCHITECTURE is now decided (owner,
2026-08-17): fees separate (domain £10/mo; our hosting deliberately
expensive), Netlify-connect via customer OAuth invited during the
request-phase build wait, no free custom-domain serving (choose-a-home page
until they pick), delivery email as the v1 account surface, own
authoritative DNS as part of the domain programme. **Phase 3 — the ZIP —
is next, and every decision above made it MORE central: it is what the
customer downloads, what a Netlify deploy uploads, AND what
"take it elsewhere" hands over. One build, three doors.**

## 1. Phase 3 scope (PLAN 2026-08-14 Part 2e, unchanged and re-ratified)

New `zip_deliverable_action.go`: list the built tree under
`portfolio-sites/<domain>/` → stream through `archive/zip` (FIRST use in the
repo) → upload under `deliverables/<domain>/` → presigned URL. Register the
action; council run for the phase; register entry per the PLAN roll-up.

**Reuse, do not rebuild** (all proven in Phase 2):
- `publish.S3Source` / `publish.ObjectStore` (platform/publish/publisher.go)
  for listing + reading — already strips the `<domain>/` prefix, already
  tested, already the pattern `publish_site_action.go` uses.
- The second-client construction idiom + spawned-pod requirement (below).
- `GetPresignedURL` exists on the S3 client (storage/interface.go).

## 2. Hazards carried from Phase 2 (each cost real time — do not re-learn)

- ⚠ **Do NOT copy b2worker's whole-buffer upload for the ZIP's own output.**
  b2worker buffers each SMALL site file because B2's S3 gateway 411s a
  non-seekable body (MissingContentLength — found by the first live canary,
  fixed in `b4981634d`). A whole-site ZIP is a different size class: stream
  with a known length (compose to a temp file then upload with size, or
  multipart) — a truncated ZIP is a silent contractual failure (the PLAN's
  own ranked risk 3). Note the test fakes now REFUSE non-seekable bodies,
  which will push you the right way.
- ⚠ **The action MUST run in a spawned storage-enabled pod** — the standing
  chassis carries no B2 credentials (owner ruling 2026-08-08). The
  site-publisher type is already on the spawner's storage allow-list; decide
  whether the ZIP action rides site-publisher's workflow or needs its own
  dispatch — either way, construct the portfolio-sites client in-action
  (the `newPortfolioStore` idiom in publish_site_action.go).
- ⚠ Migration numbers are NEVER reserved (two live collisions in two days —
  411→421 jump, and 423 taken concurrently). Re-list at write time; the
  FILENAME is the identity.
- ⚠ If you seed any workflow naming the new action: image-before-seed —
  `_HOLD.sql` sidecar with apply commands + pod-verification in the header
  (422's header is the worked example, INCLUDING its corrected negative
  control: random 40-hex, NEVER the all-zeros sha — it is git's null-sha
  constant and matches legitimately).
- ⚠ Deploy verification: the provenance stamp names the BUILD HEAD, not your
  commit — the test is `git merge-base --is-ancestor <your-sha> <stamp>`,
  with a discrimination control (a commit AFTER the stamp must be absent).
  And a "fresh build" can ship no new code (same-tag rebuild serves the
  cached image — fleet memory 08-17).

## 3. Acceptance (from the PLAN, plus the register discipline)

- `unzip -l` count == B2 object count for the domain.
- Extracted `index.html` sha256 == the B2 object's bytes.
- Presigned URL: 200 in-expiry, 403 after expiry (prove BOTH directions).
- Nothing written on any box; artefact-level checks, never item status.
- Size guard: alert (never truncate) past a threshold; the alert path needs
  a demand control (prove it fires on an induced oversize, not just that
  quiet == fine).

## 4. Council + register obligations

One council run for the phase (submit before/alongside the shipping commit;
`Council-Submitted:` trailer pattern; corr from the 097 trigger). Register
entry for the ZIP deliverable (DGH family or storage-architecture — writer's
call), in the SAME commit as the mechanism (ordering-exemption condition 2),
with the open review question stated. RFC_022 note: `publish_site` entered
the optional-key-budget counter as ZERO and was invisible until 2026-08-17 —
when adding the new action, run the cron parity test
(`cmd/config-key-audit/optional_budget_cron_parity_test.go`) and keep the
cron literal in step (CLAUDE.md's corrected RFC_022 section).

## 5. What Phase 3 does NOT include (decided, do not drift into it)

- No Netlify code (that is Phase 4b, the netlify-oauth publisher backend).
- No emails (Phase 4 owns the delivery email carrying the ZIP link).
- No domain work (parallel programme, gated on the second Nominet TAG).
- No forms/editor changes.

## 6. Falsifiers (re-check before trusting this file)

A newer handoff in this directory; the canary state
(`SELECT domain, published_hash, published_at FROM sites WHERE
publish_target IS NOT NULL` — expect noted.co.uk only, hash from 2026-08-16
16:01Z or a later legitimate republish); the chassis stamp per service
(v1.0.1305 at writing; ancestry `b4981634d` must hold); whether Stripe keys
/ webhook exposure / the second Nominet TAG have landed (they gate OTHER
tracks, not this one); `archive/zip` still absent from the repo (if present,
someone started — check git log before writing a line).
