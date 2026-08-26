# HANDOFF 2026-08-26 (night) — launch-ready, Stripe reverted-by-roll then durably fixed, trial loop next

**SUPERSEDES** `HANDOFF_2026-08-21_continue_here.md` (and the 08-18 pair). Joint with
`../site_delivery_and_editor/` stands: one session may drive both lanes; findings live
where the work happened.

**Read order, cold:** this file → `SUMMARY_2026-08-26_webdesign_uk_build_service.md`
(the launch-ready milestone read-out; 08-25's pause snapshot is history, keep it so) →
NOTES tail (the whole of 2026-08-26 is one long day) → `RUNBOOK_go_live_webdesign_uk.md`
(the unpark procedure) → `../site_delivery_and_editor/BRIEF_2026-08-26_domain_find_register_point_service.md`
(the domain programme, rulings + built chain).

## 0. State in one paragraph

The SITE is finished, verified and parked: all seven pages carry the owner's repositioning
(starter sites for experienced web designers; "No changes are included" at subtitle
prominence on index/what-you-get and as its own heading on how-it-works; categories;
not-a-hosting-company; 30 days; £59.99 domain buy-out — ruled and served 2026-08-26
night), the "Not active yet" label is up (×2, survives checks), GTM is durable, and the
customer door (second-click page over the :8090 delivery listener) is proven from the
internet. STRIPE is configured by the owner (restricted key, webhook endpoint
`we_1U8mp202nQ76FNifIrpKLN3s` with checkout.session.completed + charge.refunded) and WAS
verified keyed (webhook 400-rejects garbage) — then the night's fleet roll SILENTLY
REVERTED it to 503: terraform's `047-base-configs` owns `personae-platform-secrets`'
data map wholesale (LANDMINES entry 2026-08-26). The durable fix is COMMITTED
(`0cdc9e2d9`: the Stripe pair are now REQUIRED terraform variables) and needs the
OWNER's two steps in §2. The £30 trial-voucher variant is IN the running auth-service
(verified by ancestry against the post-roll stamp; council APPROVED r1, `e5c25b0b`).

## 1. THE OWNER'S CHECKLIST (as agreed in chat, in order)

1. **Restore Stripe, durably** — §2. (Blocks the till, not the site.)
2. **Remove the Cloudflare page rule** → site live. Box vhost verified `/c/`-free;
   RUNBOOK gates all green. After it: §3 verification.
3. **Second GTM container** (~2 min in his Tag Manager dashboard, or a TM-API service
   account = the same grant Search Console needs). Ruled 2026-08-26; creation blocked on
   his Google access; the analytics lane ("google" session) holds the click-by-click and
   receives the GTM-XXXX id.
4. **Publish GA4 into GTM-PQ3WCTBD** (his own Publish click; estate sites only).
5. **Payment Links** (dashboard): domain rental £10/month recurring + Customer Portal
   ON; domain buy-out one-off **£59.99**.
6. **First paid domain registration** when he picks a name (~£4, `--apply` is his).
7. Background: chase Nominet's second TAG (not blocking; DESIGNCONSULT is the ruled
   interim).

## 2. STRIPE — the revert and the two owner steps

> **✅ DONE 2026-08-26 (late night), verified in-session:** owner completed both steps.
> Secret carries both STRIPE_* keys again (names listed, values never read), auth-service
> restarted (fresh pods, rollout complete, 0 keyless warnings), box probe → **400**
> (keyed). Durable half confirmed: local `terraform.tfvars.secret` holds both variable
> names (mtime 22:02 tonight), commit `0cdc9e2d9` in history. Evidence: NOTES tail.
> Checklist item 1 is CLOSED; the webhook round-trip (test event → `billing_events`)
> still waits for the unpark, per §3.

The roll wiped `STRIPE_SECRET_KEY`/`STRIPE_WEBHOOK_SECRET` from the secret (webhook
503 again; keys-listing shows 9 entries, no STRIPE_*). Fix committed: they are now
REQUIRED variables in `deployments/terraform/environments/production/uk001/047-base-configs/`
— a release WITHOUT values now fails loudly naming the variable. Owner owes:

1. Add two lines to his LOCAL `deployments/terraform/environments/production/uk001/047-base-configs/terraform.tfvars.secret`
   (never committed): `stripe_secret_key = "rk_live_..."` and
   `stripe_webhook_secret = "whsec_..."` (his values; never through a session).
2. Re-run the immediate restore (until the next release picks up the tfvars):
   `kubectl -n ai-persona-system patch secret personae-platform-secrets --type merge -p '{"stringData":{"STRIPE_SECRET_KEY":"...","STRIPE_WEBHOOK_SECRET":"..."}}'`
   then `kubectl -n ai-persona-system rollout restart deployment/auth-service`.

**Verify (session-side, proven recipe):** rollout status; keyless-warning grep = 0; and
the artefact check from the box:
`ssh -i ~/.ssh/webdesign_box_ed25519 root@webdesign.vs.mythic-beasts.com 'curl -s -o /dev/null -w "%{http_code}" -X POST -d x=y http://auth-service.ai-persona-system.svc.cluster.local:8081/api/v1/billing/webhooks/stripe'`
→ **400** = keyed; 503 = keyless. After EVERY future roll, re-run this probe (the
landmine's own instruction).

## 3. AT THE UNPARK (whoever is in session)

`RUNBOOK_go_live_webdesign_uk.md` end to end. Post-verification additions since it was
written: the h1 needle is `A starter website, built once` (~66KB body); ALSO run the
Stripe round-trip (dashboard "send test event" → `billing_events` row) and close the
LANDMINES parking-rule entry (§4 of the runbook). Label: if ordering is still closed,
confirm ×2 at the served apex.

## 4. THE TRIAL LOOP (owner as customer zero) — ready once §2 + unpark land

£30 voucher variant LIVE in auth-service. Mint codes: `POST /api/v1/admin/billing/vouchers`
(admin JWT via the owner's console; no FE screen yet) with drops_price_to_pence 3000.
Loop per run: brief-starter intake → voucher at checkout → £30 → build → owner's
internal edit pass (ruled 2026-08-26: internal-only, invisible to customers — the
approval-language OFFER-SHAPE ban is ARMED and claimscan-proven so it cannot leak) →
collect at the hosted link (delivery email joins when the delivery lane ships it).
Trial sites default to `<slug>.ugg2.com`; pointing a couple at the owner's OWN portfolio
domains additionally exercises P4 pointing at zero registration cost. Cost per run =
Stripe fees (~65p).

## 5. THE DOMAIN SERVICE (severable layer — all proven except the paid step)

Chain: **find** (`../site_delivery_and_editor/find_customer_domain.sh`, VMB-018 —
generic-only stems via inline LLM workflow, brand tokens FORBIDDEN, shortest-first,
hyphens = fallback tier; fixture proven: leedsgas.uk topped a 28-name live check) →
**register** (`idea_uk_vm_site/box/nominet-epp-domain-register.py`, VMB-017, dry-run
proven from a cluster pod under DESIGNCONSULT; the first `--apply` spends ~£4 and is the
OWNER's) → **zone+point** (`../site_delivery_and_editor/cf_customer_domain_zone.sh`,
wraps the proven 08-25 recipe; emits the `zone_live_at` stamp instruction ONLY on its
verified path) → serve. Layer shape agreed with the P5 reader (delivery lane):
`site_config.domain_programme {mode, domain, tag, registered_at, zone_live_at,
commercial: rent|bought}`; default absent = `<slug>.ugg2.com`; unrecognised mode FAILS
SAFE to slug; revert of the whole programme = mode flips. EPP allow-list: ONLY the
cluster's five egress IPs work (this machine and the webdesign box are both refused —
measured). Inline LLM steps need a FULL ai_service block (provider/model/api_key_env_var).

## 6. OTHER LANES' STATE (as relayed/verified 2026-08-26)

- **Delivery email** (site_delivery_and_editor): claim layer live+hardened (DGH-017);
  copy UNBLOCKED by the two product rulings (internal-only owner edits; NO customer
  edits at launch, voice-edit their next build AFTER launch); still owed: the
  needs_delivery_review producer + copy + send. The uniform 200 on `/c/` is DESIGNED
  (no-oracle) — never "fix" it.
- **Analytics** ("google"): GTM durable fleet-wide; GA4 unblocked pending owner Publish;
  second-container creation blocked on owner Google access; the empty-container-records-
  nothing trap is banked (needs a consent-denied GA4 tag at publish to actually count).
- **Tool rebuilds**: co.uk only, nothing touches this site. The FALSE-ACCEPTANCE chain
  on tool-website-brief-starter is DEFERRED ×2 (items 0559eb67, 41d82357) until the
  acceptance checker's instance-prefix fix (staged_component_build lane) is LIVE —
  un-defer then; do NOT let an improve_tool rewrite the flagship tool on a false verdict.
- **Refunds**: DECISION_2026-08-25 Option A (webhook-as-truth, UNADVERTISED — nothing
  may enter the register); build owed in the payments/delivery lane before real sales.

## 7. STANDING CHECKS AND TRAPS (the ones a fresh session must not relearn)

- **Label**: any framework rerender of `index` strips it. Re-place anchors: above
  `class="btn btn-primary"` and above `<div class="cta-buttons">` in vm-sites
  `webdesign.uk/index.html`; push; box syncs ≤5 min. The GTM-redeploy wipe already
  happened and was the last EXPECTED one; the design rotation can still cause one.
- **Direct-fire recipe** (queue starvation recurs): pre-claim the work item, publish the
  077-style envelope with `client_id=system` via `kafka_publish_checked`, poll the
  correlation with `</dev/null` on kubectl inside read-loops (kubectl -i EATS loop stdin).
- **Served-page md5 is not a baseline** (CF email obfuscation re-keys per response);
  verify at the vm-sites REPO. webdesign.uk serves from gqls/VM-SITES via box sitesync,
  NOT gqls/sites.
- **A new FIGURE needs a fact VALUE** (unregistered_number) and each rebuild rolls fresh
  gate-collision dice — one retry on a rolled banned phrase is normal; a REPEATED
  identical refusal is deterministic, stop and read the guard.
- **Fact SOURCES keep historical prices/quotes** — clean-sweep guards scope to claims/
  writer_lines/writer_block only.
- **`kubectl patch` on personae-platform-secrets dies at the next release** (LANDMINES
  2026-08-26) — any new key goes in terraform 047 + the owner's tfvars.
- The three-aspect supersede cross-join trap: aggregate `retire` to one row before the
  INSERT joins it.

## 8. Falsifiers

- Every dated claim above; the roll cadence is daily — re-ask each SERVICE its stamp.
- `customer_access_tokens` / handovers / `billing_orders` all 0 as of 2026-08-26 night.
- The edge: apex still 302-parked; preview/links/admin controls green at last check.
- A newer handoff here or in `../site_delivery_and_editor/`.
