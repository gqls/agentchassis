# HANDOFF 2026-08-18 — JOINT cold-start: site delivery (Phase 4 next) + webdesign.uk build service — SUPERSEDES HANDOFF_2026-08-17_phase3_zip (complete) and, for operational pickup, webdesign_uk_build_service/HANDOFF_2026-08-18_continue_here.md (still holds that lane's detail)

**Why joint (owner, 2026-08-18):** one session now drives both lanes and the
owner has asked whether to merge the threads. The merge is OPERATIONAL: one
session, one cold-start (this file), because the work joins for real — the
delivery machinery (ZIP, domain rent/buy, delivery email) IS what
webdesign.uk's copy now promises customers. The two lanes KEEP their own
docs (NOTES/PLAN/RUNBOOK/README/SUMMARY) and their own registers; update
each lane's NOTES where the work happens.

## 0. State in one paragraph

Phase 3 (the ZIP deliverable) is COMPLETE and live-proven (register DGH-011;
v1.0.1308; canary 8/8 byte-verified; presign both directions — note B2
returns 401 UnauthorizedAccess + an expiry message for expired presigns,
never AWS's 403). webdesign.uk's index gained the chat box and the
payment-first terms on the served page (2026-08-18 morning). The owner then
issued copy directives (two rounds, both applied at the REGISTER — the wire):
one-shot/no approval stage stated bluntly; starter-site-with-initial-copy
framing; domain rent £10/mo (link arrives in the delivery email) or buy
£200 one-off then transfer freely; NO example-site links yet (none were
built by this one-shot route — and the framework's cross_site_domain guard
REFUSED them anyway); the post-payment link is never called a "preview";
keep-it-online hosting clarity (host it yourself after the month; free
options recommended; the ZIP's instructions walk through set-up); no
pre-sales service. Four page rewrites are queued/running against that
register: index, what-you-get, faq, how-it-works.

## 1. FIRST: verify the four webdesign.uk rewrites landed

Items (needs_content_page, owner-brief-2026-08-18): what-you-get `cf83a513`,
faq `f853f532`, how-it-works `8d969047`, index `5c6f73ac`.

```sql
SELECT spec->>'page_name', status FROM site_work_items
 WHERE id::text IN ('cf83a513-bc98-4b4d-b50b-d27001c76fdf','f853f532-ef9f-4951-9d45-27ed0757ae85',
                    '8d969047-88a3-4384-8376-6699135e67c7','5c6f73ac-5ad6-413a-84e3-c5ff112c57f5');
```
On failure, the blocker IS persisted — never grep pods:
```sql
SELECT occurred_at, jsonb_pretty(context) FROM agent_error_log
 WHERE error_code='CONTENT_VALIDATION_BLOCKER_DETAIL' ORDER BY occurred_at DESC LIMIT 3;
```
Verify at the SERVED pages (cache-busted), never item status. Expected on
every page: no approval/sign-off step; no "preview" naming for the
post-payment link (it is "your site, live at an address we provide, for
about a month"); no pay-after-preview sentence. FAQ specifically: text
answer gives both halves (initial copy included; starter site the customer
is expected to edit); domain answer states £10/mo rent (link in the
delivery email) and £200 one-off buy then free transfer; get-started answer
does NOT invite email/phone and says any sort of site (NO example links);
questions answer offers no pre-sales service. Known coin-flip failures the
gate correctly catches (just re-triage the item): a passing "refund"
mention (writer_block now steers: point at "the full terms"), any stray
cross-site domain. Locked CTA components may refuse the rewrite and keep
"Send us an email" copy: that is an owner unlock decision, do not force it.

## 2. THEN: Phase 4 — handover + the delivery email (the next build)

PLAN_2026-08-14 Phase 4 mechanics + PLAN_2026-08-17 decision 3, PLUS today's
addition: the email's domain links are now TWO — rent £10/mo (subscription)
AND buy £200 one-off with free transfer (owner, 2026-08-18, recorded in
site_delivery NOTES and as webdesign.uk register facts domain_rent_monthly /
domain_buy_once). The email carries: ZIP link (dispatch
`zip-deliverable-dispatch` with `{domain}` — recipes in
`sql_for_agents/459_zip_deliverer_agent_HOLD.sql`'s header, APPLIED), the
live-site link, Netlify-connect invite (request-phase repeat), both domain
links, Stripe hosted portal. `sites.handed_over_at` migration + single
reader. `platform/mailer` is the sanctioned mailer. One council run;
register entry same-commit.
**Blocked on the owner, gating first revenue:** Stripe keys; the webhook
edge exception (apex 302s); second Nominet TAG (domain programme only).

## 3. Standing hazards for this joint work

- The REGISTER is the wire: never steer via item-spec prose, never hand-edit
  HTML (owner ruling 2026-08-04). writer_block edits are BY ANCHOR with
  exactly-once guards (worked examples: SQL_2026-08-17b, 18b, 18c in the
  webdesign lane dir).
- Ban patterns must match PROMISE shapes only — a bare token bans the
  DENIAL too (the \brefunds?\b precedent), and the negation guard does NOT
  treat bare "no" as a negator.
- The framework's cross_site_domain guard blocks any other-site domain in
  webdesign.uk copy — example links need an allow-list mechanism first, if
  the owner ever wants them (he deferred: examples only once this route has
  produced sites).
- Stat/figure fields publish only attested numbers; hedged facts stay prose
  (the "1 day" lesson, 2026-08-18).
- Deploy proof: image label `org.opencontainers.image.revision` + ancestry +
  digest against the pod; a fresh build can ship no new code.

## 4. Read order (cold)

This file → site_delivery_and_editor/PLAN_2026-08-17_delivery_architecture_decisions.md
(OWNER DECISIONS + build order) → webdesign_uk_build_service/HANDOFF_2026-08-18_continue_here.md
(that lane's detail: bugs_open/299, owner decisions outstanding, prompt-maker
TODO) → both NOTES tails (2026-08-18 entries) → register DGH-011 for the ZIP
mechanism.

## 5. Falsifiers

A newer handoff in either lane dir; the four items' statuses (§1 — if all
complete and served pages verified, §1 is done, strike it); whether Stripe
keys / webhook exposure / second TAG have landed; `sites.handed_over_at`
existing already (someone started Phase 4); the webdesign.uk register's
updated_at (another session may have edited facts since 2026-08-18 ~12:30Z).
