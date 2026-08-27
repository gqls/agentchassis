# HANDOFF 2026-08-27 (midday) — SHOPFRONT LIVE; delivery fully armed pending ONE deploy that was MID-FLIGHT when the session ended; the rehearsal is scripted and the site is RULED

**Supersedes** `HANDOFF_2026-08-26_continue_here.md` (its items are ALL closed — the ✅
blocks in it say how). **Owner is away until ~Monday 2026-09-01; he left mid-deploy.**
Milestone read-out (fuller than usual, owner-requested, joint with the webdesign lane):
`../webdesign_uk_build_service/SUMMARY_2026-08-27_webdesign_uk_build_service.md`.

## 0. State in one paragraph

**webdesign.uk WENT LIVE 2026-08-27 ~10:00Z** (both parking page rules toggled OFF —
disabled not deleted, re-park = toggle back on; post-unpark table passed from outside;
the 522 trap and its fix are in the LANDMINES page-rule entry's dated UPDATE). Ordering
stays CLOSED — "Not active yet" served ×2. The delivery chain is fully live config on
live binaries; **the mail secret EXISTS** (owner created `delivery-smtp-secrets`
2026-08-27 morning); chassis pods carry `DELIVERY_SMTP_HOST/PORT/USER/FROM` since the
overnight v1.0.1346 roll. **The owner started a whole-fleet deploy ~11:00Z (his words:
"approximately 1 hour") whose job for this lane is carrying `DELIVERY_SMTP_PASS` onto
the pods** — its completion was NOT verified before the session ended. Rehearsal site
RULED by the owner: **remortgagecalculator.uk** (12 pages, deployed). Nothing
customer-facing has happened: tokens 0, handovers 0 (last read ~10:00Z).

## 1. NEXT, in order

1. **Verify the owner's deploy landed** (it was mid-flight; do not assume):
   ```bash
   POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis --no-headers -o custom-columns=:metadata.name | head -1)
   kubectl -n ai-persona-system exec "$POD" -- printenv DELIVERY_SMTP_PASS >/dev/null && echo PASS-ON-PODS || echo NOT-YET
   ```
   `NOT-YET` on fresh pods = secret missing or env gone (check both); on OLD pods
   (age > ~1 day) = the deploy never ran — ask the owner, do NOT one-service apply
   (whole-fleet rule stands).
   > ✅ **DONE 2026-08-27 afternoon (same session, after the usage-limit pause): the
   > deploy LANDED.** Fresh chassis generation (pod `7df947c88b-*`) carries the FULL
   > `DELIVERY_SMTP_*` set including PASS (verified by presence + length only — the
   > value was never read). **Email is fully armed.**
2. **Post-roll checks** (first release since the Stripe terraform fix `0cdc9e2d9`):
   `curl -sS -o /dev/null -w '%{http_code}' -X POST https://webdesign.uk/stripe/webhook -d x=y`
   → **400** (keyed). A **503** means the roll wiped the keys again and the terraform
   fix did not hold — that outranks everything else that day. Also the listener triple
   (recipes §3 of the 08-26 handoff) and stamps per service if anything looks stale.
   > ✅ **DONE 2026-08-27 afternoon: webhook → 400 — THE TERRAFORM FIX HELD through its
   > first release.** Listener routes re-verified outside post-roll (/c/ 200 · /d/ 200
   > "no longer active" · /other 404 · apex 200). Label still ×2. Counters 0|0.
   > **Monday's pickup is therefore ONLY §1.3's label re-check (days will have passed)
   > and §1.4 THE REHEARSAL.**
3. **Re-check the "Not active yet" label** — several days will have passed and the
   design rotation rerenders index unprompted (it stripped the label 08-26 evening;
   re-placed 08-27 morning, vm-sites `b72c608`):
   `curl -s https://preview.webdesign.uk/ | grep -c 'hand-placed 2026-08-25'` → want 2.
   If 0: re-place per the webdesign runbook gate 2 (two insertion points, markup in
   vm-sites `63bd5a6`/`b72c608`), push, `systemctl start sitesync.service` on the box.
4. **THE REHEARSAL — everything is decided, just run it** (recipes:
   `sql_for_agents/651_delivery_review_and_email_agents_HOLD.sql` header; ZIP recipe in
   459's header):
   - Site: **remortgagecalculator.uk** (owner ruling 2026-08-27). Email:
     **info@designconsultancy.co.uk** (proposed to the owner with no objection —
     confirm in passing when he's back).
   - Respect ~300s after any chassis restart, then dispatch `delivery-review-filer`
     `{site_id, domain, site_url, brief}` via kafka-publish-lib (OPP-009 — CHECK the
     receipt).
   - Item appears at `needs_human_review` → **owner edits then presses APPROVE on
     admin.apis.uk** (NOT resolve — resolve writes the key the gate ignores; the
     approve path has run once ever, on a March test item — this is also its first
     real exercise).
   - `zip-deliverable-dispatch {domain}` → note `presigned_url` + `expiry_minutes`.
   - Dispatch `delivery-email-sender` `{site_id, customer_email, live_site_url,
     zip_presigned_url, zip_presign_minutes}`.
   - Owner clicks every link FROM OUTSIDE; confirm `dkim=pass` on the received headers.
   - ⚠ The handover stamp is ONCE-ONLY and permanent — a second dispatch is REFUSED by
     design (recovery recipe for stamped-but-unemailed is in 651's header).
5. **After the rehearsal**: the delivery lane's next builds are the weekly chase + the
   42-day retraction job (still NOT BUILT — nothing takes sites down yet), then the
   VOICE EDITOR (owner ruling 08-26). Ordering-opening items (Payment Links £10/mo +
   £59.99 buy-out, label removal, GTM/GA4) are the webdesign lane's checklist.

## 1b. LATE-AFTERNOON ADDITIONS (owner stayed a few hours; all committed)

The site LAUNCHED and was immediately trialled — four more things happened, each with
its full account in the webdesign lane's NOTES 08-27 entries:

1. **Chat launch-day bug FIXED + released (`160546543`)**: gate 1 counted every message
   though its design says conversation STARTS — a real 5-turn intake died 429 mid-flow.
   Now start-only (continuations bounded by turn cap 20 + $10/day ceiling), self-minted
   ids still count as starts, blocked starts allocate nothing; two tests, mutation-proven;
   box-released with provenance verified; live 7-messages-one-conversation proof passed.
   ⚠ operator fact: this workstation's curls SHARE the owner's public IP — probes burn
   his 5-new-chats/hour band; heavy testing that OPENS many chats still trips it
   (in-memory: `systemctl restart webdesign-chat` clears).
2. **Brief-starter tool re-pointed at the chat** (improve_tool `be0bdf28`, complete +
   SERVED): no contact form exists — the chat IS the intake; "before we speak" gone.
3. **TRIAL RUN 1 IN FLIGHT**: brief committed via the chat, **reference BR-9AUZ59**
   (Boxing Online / boxingonline.com / aaa@designconsultancy.co.uk). Client row
   `a7395f69-e735-4390-98d7-9f17085338f4`. **NEXT trial step (owner, own terminal):**
   `webdesign_uk_build_service/trial_checkout.sh` — worked invocation in its header —
   mints the ruled £30 voucher + creates the order + prints the Stripe checkout URL.
   After he pays: webhook → paid → `collect_external_orders` releases the brief →
   **watch build_queue for the first customer-shaped build**, then the rehearsal's
   review/approve/deliver flow applies to THIS site. No email at intake is BY DESIGN
   (first email = delivery email); an intake-ack email is a future owner decision.
   > ✅ **PAID, same evening (owner stayed):** order `36744bf0` paid 14:40:22Z, £30,
   > BR-9AUZ59, voucher redeemed, verified at the DB row AND the Stripe dashboard —
   > **the first real payment through the platform**, and the script's run doubled as
   > the outstanding billing mint-path acceptance. TWO findings, full detail in the
   > webdesign NOTES 08-27 evening entry: (1) **`/pay/success`+`/pay/cancel` exist
   > nowhere** — every buyer lands on a bare 404 after paying; owed before ordering
   > opens, framework-built pages. (2) **The brief does NOT auto-release — correctly**:
   > seed 661 ships disabled pending P5 seeding (contact details + evidence_base;
   > honesty guards otherwise unarmed); token on pods, action in the running image
   > (ancestry + reversed control). **So Monday's build order is: P5 wiring → apply
   > 661 (asserts disabled) → enable with the owner → run 1's brief releases → build →
   > his edit pass + APPROVE → delivery email = the rehearsal, on a real paid order.**
   > Do NOT force-release before P5 — it builds a degraded site.
4. Blueprint Compiler: owner REVERSED the morning's Lovable/v0 removal before it ran —
   references STAY (deliberate; more third-party positioning later, not less).

## 2. What changed today (all committed; keys to find them)

| thing | state | where |
|---|---|---|
| shopfront | **LIVE both hostnames**, verified outside | LANDMINES page-rule entry UPDATE 2026-08-27; webdesign NOTES 08-27 |
| the 522 trap | apex/www DNS had NEVER pointed at the tunnel; fixed via `cloudflared tunnel route dns --overwrite-dns` (cert: `/root/.cloudflared/cert.pem.webdesign`) | same entries |
| box nginx | dead 06:22→08:32Z (unattended-upgrade restart lost the cluster-DNS race); started + **retry drop-in installed** (`/etc/systemd/system/nginx.service.d/retry-on-failure.conf`) — self-heals ≤15s now | LANDMINES 2026-08-27 entry (`error code: 502` signature = box-side, never cluster) |
| /d/ vhost | APPLIED + outside-verified (uniform "no longer active" 200 on token-shaped junk) | 08-26 handoff item 3 ✅ block |
| SMTP env | HOST/PORT/USER/FROM on pods (v1.0.1346); PASS = the mid-flight deploy | §1.1 above |
| mail secret | EXISTS; ruled OUT of terraform (047 owns only `personae-platform-secrets` — a release cannot touch a secret nothing owns) | NOTES 08-27 mid-morning |
| Blueprint Compiler | **v0/Lovable references STAY — owner ruling 08-27 (reversal)**; both improve_tool items cancelled unclaimed; do not re-file | NOTES 08-27 late morning |
| 'Not active yet' label | re-placed after the ba44c5c strip; standing check in §1.3 | webdesign NOTES 08-27 |

## 3. Falsifiers

The deploy's outcome (§1.1 decides everything downstream); whether anyone rehearsed
while the owner was away (`SELECT count(*) FROM customer_access_tokens;` and
`SELECT count(*) FROM sites WHERE handed_over_at IS NOT NULL;` — non-zero means the
rehearsal or worse happened, re-read before acting); the label (§1.3); a newer handoff
here or in the webdesign dir; pod stamps re-read per service (tags roll daily); whether
ordering opened (Payment Links existing changes the label instruction from keep to
remove).

## 4. Read order, cold

This file → `../webdesign_uk_build_service/SUMMARY_2026-08-27_...` (the milestone) →
NOTES 08-27 entries here (the day's evidence) → `651_..._HOLD.sql` header (rehearsal
recipes verbatim) → LANDMINES entries "A CLOUDFLARE PAGE RULE..." (UPDATE 08-27) and
"`systemctl restart nginx` ON THE WEBDESIGN BOX..." (new today).
