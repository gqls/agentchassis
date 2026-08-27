# HANDOFF 2026-08-26 (night) — the delivery chain is LIVE CONFIG on a LIVE BINARY; what remains is a secret, one apply-cycle, and a rehearsal

**Supersedes, for this lane's pickup:** the joint cold-start chain ending at
`../webdesign_uk_build_service/HANDOFF_2026-08-21_continue_here.md` (the joint file's
webdesign half still holds that lane's detail — they have their own active threads).
**Milestone read-out:** `SUMMARY_2026-08-26_site_delivery_and_editor.md` (same directory).
**Depth:** `NOTES_site_delivery_and_editor.md` 2026-08-25 → 08-26 entries (the whole arc:
listener, Fable review, chain build, roll, seeds).

## 0. State in one paragraph

Phase 4 is DONE and LIVE end-to-end as of tonight: the delivery-only listener (SYS-095)
serves `/c/` from the internet (outside-verified 08-26 afternoon); the second-click
confirmation page is live and outside-verified; migration 650 is applied; the roll at
stamp **`b34c24f4c`** (chassis AND core-manager, same commit — both my chain commits
`10a963da2`/`aca0afe1d` proven ancestors, controls passed) carries `send_delivery_email`,
`refresh_zip_link` and the `/d/` handler; seeds **651 + 652 are APPLIED and verified**
(all three agents active: `delivery-review-filer`, `delivery-email-sender`,
`zip-link-refresher`; the `zip-link-refresh` schedule enabled at 6h). Every council trail
is APPROVED (25cd3044 r4, 0b84970d, 5a33a174 r1, c618d189 r1; the 0b84970d post-approval
hardening round was submitted 08-26 night — check it landed). The owner ruled BOTH open
product questions (his edit pass = internal, invisible; no customer edits at launch,
VOICE EDITOR is the next build), and email authentication is proven at Gmail itself
(dkim/spf/dmarc all pass). `customer_access_tokens` = 0, handovers 0/0 — nothing customer-
facing has happened yet.

## 1. NEXT, in order (the short list)

1. **OWNER: create the mail secret** (one command, he holds the password):
   `kubectl -n ai-persona-system create secret generic delivery-smtp-secrets --from-literal=DELIVERY_SMTP_PASS='<password>'`
2. **The `DELIVERY_SMTP_*` env reaches chassis pods at the next `apply -k` + restart** —
   it was committed tonight (`6d76dab1e`, in the chassis overlay) AFTER the roll, so
   tonight's pods do NOT carry it. No urgency: `send_delivery_email` fails loudly at its
   sender-first check, BEFORE the once-only stamp, until env+secret exist.
   ⚠ `optional: true` on the secret ref is LOAD-BEARING — without it a missing secret
   takes down the ~46-pod chassis fleet. Never "tidy" it away.
   > ✅ **DONE 2026-08-27 (overnight roll to v1.0.1346):** chassis pods carry
   > HOST/PORT/USER/FROM — verified by printenv on a live pod. PASS still absent
   > (secret not created). ⚠ secretKeyRef resolves at container START — after the
   > secret is created, PASS arrives only at the NEXT roll/restart.
3. **OWNER box step: re-apply `links.webdesign.uk.nginx`** — it gained the `/d/` location
   block (GET-only, same anchored token regex). Gated on the roll, which has now
   happened, so it is UNBLOCKED. After apply, `/d/<43-char junk>` from outside should
   return the uniform "no longer active" page (200), and `/d/x` a box-local 404.
   > ✅ **DONE 2026-08-27 (session — sessions CAN ssh the box, runbook 08-25
   > correction):** backup at `/root/links.webdesign.uk.bak-2026-08-27`, applied,
   > reloaded; outside-verified exactly as predicted (`/d/<43>` → 200 "no longer
   > active", `/d/x` → 404, `/c/<43>` → 200). Same morning: the box's nginx had been
   > DEAD since 06:22 (unattended-upgrade restart lost a cluster-DNS race — uniform
   > 502s, LANDMINES 2026-08-27 entry); started + hardened with a retry drop-in.
4. **THE REHEARSAL** — the whole flow on a site of our own before any customer:
   dispatch `delivery-review-filer` (recipes in `sql_for_agents/651_...HOLD.sql`'s
   header) → the item appears at `needs_human_review` → owner edits the site, presses
   **APPROVE** on admin.apis.uk (NOT resolve — resolve writes the key the gate ignores;
   this is also the first real exercise of the approve path, which has run once ever,
   on a test item in March) → `zip-deliverable-dispatch {domain}` (459's header) → note
   `presigned_url` + `expiry_minutes` from its output → dispatch `delivery-email-sender`
   `{site_id, customer_email, live_site_url, zip_presigned_url, zip_presign_minutes}` →
   receive the email, click every link FROM OUTSIDE, confirm dkim=pass on the received
   headers. ⚠ Respect the ~300s no-dispatch window after any chassis restart.
5. **Read the pending council verdict** on the 0b84970d hardening round; act on a REVISE.
   > ✅ **DONE 2026-08-26 ~21:15Z (fresh session): APPROVED** — verdict actually landed
   > 08-26 **10:29** (same morning as the resubmission, not night), 1 advisory, none
   > high-severity. The advisory (no producer shown calling `ReviewItemRequiredSpec`)
   > is closed by seed 651: live `delivery-review-filer` row verified carrying
   > `spec_literal = {"checkpoint": true}`. Items 1/3 re-checked the same hour: secret
   > still absent, `/d/` still 404 from outside — both stand. NOTES (late night) has
   > the evidence.

## 2. What is LIVE vs INERT, precisely

| thing | state | proof |
|---|---|---|
| delivery listener :8090 + `/c/` | LIVE, outside-verified 08-26 | pod triple 404/401/200; box table + suffix control from outside |
| second-click page | LIVE, outside-verified | GET 200 render-only / POST confirms; suffix 404 |
| `/d/` route in the binary | LIVE in core-manager (`b34c24f4c`) | ancestry + the roll |
| `/d/` at the BOX | **INERT until the owner re-applies the vhost** (item 3) | file committed, box unapplied |
| migration 650 (stored_url) | APPLIED + recorded | guard+verify passed; columns confirmed nullable |
| seeds 651/652 (3 agents + schedule) | **APPLIED + verified tonight** | agents active; schedule enabled 6h; recorded in schema_migrations (2nd attempt — see §4) |
| DELIVERY_SMTP env | committed, **NOT on pods yet** | reaches pods at next apply+restart; secret does not exist yet |
| refresher schedule | ARMED; fires only when a due token exists (0 today) | pre_query returns nothing → no dispatch |
| email copy | LIVE CONFIG, owner-editable | `delivery-email-sender.default_config...body_template`; figures = attested register |
| weekly chase + retraction job | **NOT BUILT** (plan items 4–5) | nothing takes sites down; 42d column records intent only |
| P5 domain wiring | NOT BUILT; **contract SETTLED** | BRIEF_2026-08-26 (same dir): hostname=domain ONLY when `zone_live_at`; unknown mode → slug; email branches on `commercial` |
| voice editor (Phases 5–6) | NEXT BUILD post-launch (owner ruling 08-26) | NOTES 08-26 |

## 3. The verification recipes that keep being needed

- **Stamp a service:** provenance line scrolls FAST (chassis: gone in 17 min). Fallback
  is NOT grepping your commit's sha in the binary — **the binary carries ONE stamp (the
  build commit), not ancestry**. Probe `/proc/1/exe` against CANDIDATE full shas from
  `git log --since`, with an absent-sha control. Then `git merge-base --is-ancestor
  <your-commit> <stamp>` + the reversed control.
- **Anything customer-facing: verify from OUTSIDE.** The box's anchored regex means
  cluster-internal curls prove nothing (LANDMINES, footprint `links.webdesign.uk.nginx`).
- **Refresher health = the OUTCOME, never run status** (spawn→call fails ~50% and
  self-heals next tick): `SELECT count(*) FROM customer_access_tokens WHERE
  purpose='zip_download' AND revoked_at IS NULL AND expires_at > now() AND
  stored_url_expires_at < now();` — want 0.

## 4. Traps laid or found tonight (read before touching)

- **`ON CONFLICT (type)` does not exist on `agent_definitions`** — seeds use
  `WHERE NOT EXISTS ... AND deleted_at IS NULL` (459's real pattern). My first apply
  failed on the invented form.
- **Apply → verify → record, THREE separate acts.** I batched the ledger INSERT with the
  applies and briefly recorded two failures as applied (WRONG_CALLS 08-26: "a record in
  the same batch is a prediction").
- **TWO files named 651** in sql_for_agents (mine + robot-hands gripper). Resolve by
  filename, always.
- **The empty-stamp ancestry trap**: `git merge-base --is-ancestor X ""` returns false
  and reads as "NOT deployed". If your stamp variable can be empty, your control must
  fail — mine did, which is what caught it.
- The email action's guards are ORDER-dependent: sender construction and template-vs-links
  validation sit BEFORE `delivery.Claim` on purpose (a failure after the stamp strands a
  site handed-over-but-unemailed; recovery = the operator recipe in 651's header).
- **The 30/42 gap is deliberate** (promise 30, serve 42; owner 08-25) — a test fails if
  equalised. The email template says {{days}} = 30 via `AdvertisedLiveWindowDays`.

## 5. Owner-facing summary of what he still owes (all small)

Mail secret (§1.1) · box vhost re-apply (§1.3) · the rehearsal's approve press (§1.4) ·
Stripe Payment Links when he chooses (keys are LIVE in the cluster since tonight,
`ad8a9b596` — the email's domain section gains real links then, via the
`domain_rent_url`/`domain_buy_url`/`stripe_portal_url` config slots) · the two Cloudflare
facts for DNS plan B (`PLAN_2026-08-25_dns_plan_b.md`) · the EPP first `--apply`
(spends money; his moment).

## 6. Falsifiers

A newer handoff in this dir or the webdesign dir; the hardening-round verdict; whether
the mail secret / box apply / rehearsal have happened (check `customer_access_tokens`
count and `sites.handed_over_at` FIRST — non-zero means a delivery happened and every
"nothing at risk" line above is expired); pod stamps re-read per service (tags roll
daily); the webdesign lane's launch state (their SUMMARY_2026-08-25 + runbook —
unpark may have happened).

## 7. Read order, cold

This file → `SUMMARY_2026-08-26` (the milestone) → NOTES 08-25/08-26 tail (the arc and
the traps) → `sql_for_agents/651_...HOLD.sql` header (the flow + dispatch + recovery
recipes) → register **DGH-017/DGH-018/SYS-095** → `BRIEF_2026-08-26_...` (the settled P5
contract) → `RUNBOOK_site_delivery_and_editor.md` (SMTP + DKIM state).
