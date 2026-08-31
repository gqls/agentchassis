# HANDOFF 2026-08-31 (Monday evening) — THE COLLECTOR IS LIVE AND THE FIRST PAID BUILD IS RELEASING; what remains is the build itself, the owner's approve, and the delivery email

**Supersedes** `HANDOFF_2026-08-27_continue_here.md` (all its items closed — its ✅ blocks
say how) **and covers the JOINT picture** (site_delivery_and_editor + webdesign lanes —
one session has driven both since 08-18; findings live where the work happened).
**Owner context**: back from days away; API budget raised; he deployed a fresh chassis
build this evening at my request and asked for this handoff to continue in a new chat.

## 0. State in one paragraph

**webdesign.uk is LIVE** (unparked 08-27, verified from outside; ordering still CLOSED —
"Not active yet" ×2 served, re-placed after its THIRD rebuild-strip on 08-30).
**The first real order exists and is PAID**: billing_orders `36744bf0`, £30 (ruled
voucher WD-9FAB5-2NVNF redeemed), reference **BR-9AUZ59** (Boxing Online /
boxingonline.com / aaa@designconsultancy.co.uk), verified at the DB row AND the Stripe
dashboard. **P5 (build-release wiring) is BUILT, COUNCIL-APPROVED (7e3dd082, REVISE→r2
APPROVED, advisories acted on), and IN THE RUNNING BINARY** (symbol probe + control
passed on tonight's fresh pods). **Seed 661 is APPLIED and the `order-intake-collect`
task was ENABLED tonight** (900s cadence) — the collector polls the box, finds the paid
brief, and releases it into build_queue; the release may already have happened by the
time you read this (§1.1 verifies). Delivery machinery is fully armed end-to-end (SMTP
env + password on pods since 08-27; /c/ and /d/ outside-verified; email auth proven at
Gmail). The chat runs on the fleet key still — the owner holds a separate-account key
and the one-env-line swap recipe (§4).

## 1. NEXT, in order

1. **Verify the release happened** (the collector's first tick was due within 15 min of
   ~enable time tonight):
   ```sql
   SELECT domain, status, priority FROM build_queue WHERE direction->>'order_reference'='BR-9AUZ59';
   SELECT id, domain, email, company_name FROM sites WHERE domain ILIKE '%boxingonline%';
   SELECT aspect, source, created_by, jsonb_array_length(data->'facts') FROM site_specs
    WHERE site_id=(SELECT id FROM sites WHERE domain ILIKE '%boxingonline%') AND aspect='evidence_base' AND is_current;
   ```
   Expect: build_queue row `queued`→`seeded`; a sites row carrying
   email=aaa@designconsultancy.co.uk + company_name='Boxing Online'; ONE evidence_base
   row, source='order_intake', 2 facts. That triple IS P5's live proof — record it.
   > ✅ **DONE — VERIFIED LIVE the same evening, minutes after enabling.** The
   > collector's FIRST tick released the brief; all three legs measured in one frame:
   > build_queue `boxingonline.com | seeded` · sites `d2aa5206` with
   > email=aaa@designconsultancy.co.uk, company_name='Boxing Online' · register
   > source='order_intake', created_by='seed_build_queue', 2 facts,
   > verification_status='customer_attested'. First work item `needs_domain_research`
   > already **claimed** by domain-research-classifier — **the first customer-shaped
   > build is RUNNING as of 2026-08-31 evening.** Pickup starts at §1.2: watch the
   > build, then §1.3 the rehearsal.
   If NO build_queue row: read the collector's runs —
   `SELECT status, current_step, created_at FROM orchestration_states WHERE agent_type='order-intake-collector' ORDER BY created_at DESC LIMIT 3;`
   (spawn→call fails ~half the time and SELF-HEALS next tick — do not cancel a FAILED
   row; also check the box endpoint: the brief lives on the box, served at
   /internal/orders to the collector's bearer token, WEBDESIGN_BOX_ORDERS_TOKEN on pods.)
   ⚠ A repeat-domain or no-domain paid order goes to needs_human_review INSTEAD of
   building — that path is correct behaviour, not a failure.
2. **Watch the build** — the normal pipeline takes it from `seeded` (needs_domain_research
   first item, objective = the chat brief). This is the FIRST customer-shaped build:
   watch site_work_items for the site, and expect the honesty guards ARMED (the
   evidence_base register above is what validate_page_content reads — claims checks run
   on this site, unlike every prior estate build).
3. **When the build completes: the rehearsal-on-a-real-order** (recipes verbatim in
   `sql_for_agents/651_delivery_review_and_email_agents_HOLD.sql` header):
   dispatch `delivery-review-filer` {site_id, domain, site_url, brief} → owner EDITS the
   site then presses **APPROVE on admin.apis.uk** (NOT resolve — resolve writes the key
   the gate ignores; first real exercise of the approve path) →
   `zip-deliverable-dispatch {domain}` → note presigned_url + expiry →
   `delivery-email-sender` {site_id, customer_email, live_site_url, zip_presigned_url,
   zip_presign_minutes} → owner clicks every link FROM OUTSIDE, confirm dkim=pass.
   ⚠ handover stamp is ONCE-ONLY; recovery recipe for stamped-but-unemailed in 651's
   header. ⚠ ~300s no-dispatch after any chassis restart.
4. **The domain question for THIS site**: the customer (the owner) asked for
   boxingonline.com — an EXTERNAL domain he owns. Serving/pointing is the P5 domain
   contract's territory: hostname=domain ONLY when `zone_live_at`; unknown mode → slug
   (site serves at its ugg2 slug until domain wiring is done). Do NOT invent domain
   wiring during the rehearsal — deliver at the slug URL; the domain programme
   (BRIEF_2026-08-26, same dir) is its own workstream.

## 2. What is LIVE vs staged (the joint table)

| thing | state | proof/key |
|---|---|---|
| shopfront both hostnames | LIVE, ordering closed, label ×2 | 08-27 unpark + 08-30 label re-place (vm-sites `55835ad`) |
| first paid order | PAID £30, BR-9AUZ59 | billing_orders `36744bf0`; Stripe dashboard |
| P5 seeding | in the RUNNING binary (probe+control tonight) | commits on trail `Council-Submitted: 7e3dd082` (APPROVED r2) |
| collector (661) | agent active; task **ENABLED tonight 900s** | scheduled_tasks `order-intake-collect` |
| delivery chain | fully armed (env+PASS on pods; /c/ /d/ live outside; dkim pass) | 08-27 handoff ✅ blocks |
| chat | working on the FLEET key; owner holds separate-account key | swap recipe §4; the two limit outages 08-27/08-30 |
| /pay/success + /pay/cancel | **DO NOT EXIST — every buyer 404s after paying** | owed BEFORE ordering opens; stripe.go mints the URLs |
| weekly chase + 42d retraction | NOT BUILT | nothing takes sites down |
| voice editor | next build post-launch (owner ruling 08-26) | — |
| Payment Links (£10/mo, £59.99) + label removal + GTM/GA4 | owner checklist, opens ordering | webdesign HANDOFF_2026-08-26 §1 |

## 3. Traps for whoever picks this up (hard-won this week)

- **Chat failure shapes, three of them**: contactLine AS THE REPLY = server-side Claude
  failure (journalctl -u webdesign-chat; both outages were the ACCOUNT usage limit);
  generic "Something went wrong" under the input = HTTP failure (access.log status:
  429 = the service's own per-IP limiter — 5 NEW conversations/hour, counters in-memory,
  restart clears; 503 = nginx limit_req); box-wide fast 502 "error code: 502" =
  nginx dead on the box, NEVER the cluster (LANDMINES 08-27; retry drop-in self-heals).
- **This workstation shares the owner's public IP** — your curls burn his chat
  rate-limit allowance.
- **The label strips on EVERY index rerender** (three times now). Standing check:
  `curl -s https://preview.webdesign.uk/ | grep -c 'hand-placed 2026-08-25'` → 2.
  Re-place = vm-sites, two insertion points, push, `systemctl start sitesync` on box.
- **Enable-gates probe the RUNNING BINARY, never git ancestry** (661's header has the
  worked recipe; a same-tag cached rebuild defeats ancestry — council 7e3dd082).
- **write_site_spec's deep-merge OVERWRITES arrays** — never route evidence_base facts
  through it onto an existing register (its own header now warns; seedCustomerIdentity's
  guarded insert is the safe seed-time path).
- **The box is reachable**: `ssh -i ~/.ssh/webdesign_box_ed25519 root@webdesign.vs.mythic-beasts.com`.
  Kubeconfig token expires every ~3 days (owner refreshes).
- **661/651 are _HOLD seeds applied BY HAND** — both applied; do not re-apply
  (651+652 on 08-26, 661 tonight after its agent_category fix).

## 4. The API-key/budget question (owner has a separate-account key)

Discussion plan for another thread: `../webdesign_uk_build_service/PLAN_2026-08-31_api_budget_separation.md`.
The immediate act is owner-run (key never enters a session): edit ANTHROPIC_API_KEY in
`/etc/webdesign-chat.env` on the box, `systemctl restart webdesign-chat`, send one chat
message. Per-site chats each have their own env (`/etc/sitechat/<domain>.env`).
Own-cluster restructuring: costed in the plan, not recommended on budget grounds alone.

## 5. Falsifiers

Whether the release fired (§1.1 FIRST — its result changes everything downstream);
`SELECT count(*) FROM customer_access_tokens` + handed_over sites (0|0 as of tonight —
non-zero means delivery happened); the label (§3); pod stamps per service (tags roll
daily); a newer handoff in either lane dir; the chat key (fleet vs separate — read the
env file's key NAME never its value); whether ordering opened (changes the label rule
and makes /pay/success URGENT).

## 6. Read order, cold

This file → `../webdesign_uk_build_service/NOTES_webdesign_uk_build_service.md` 08-27→08-31
entries (the launch week: three chat arms, trial run 1, P5's two council rounds) →
`SUMMARY_2026-08-27b_...` (the first-order milestone read-out) →
`sql_for_agents/661_order_intake_collector_HOLD.sql` header (flow + enable-gate) →
`651_..._HOLD.sql` header (rehearsal recipes) → `PLAN_2026-08-31_api_budget_separation.md`
(the key question) → LANDMINES entries: the page-rule UPDATE 08-27, the box-nginx
502 entry, the three-stores identity entry.
