# idea.uk — handoff (2026-06-11)

The one-page picture of where idea.uk is and exactly what to do next. Full cross-session
journal: `running_notes.md` (read the tail). Architecture/deploy detail:
`idea_uk_architecture_and_deployment.md`. **Email moved to AWS SES on 2026-06-11** — any earlier
note about Clook/MailChannels or port 465 is superseded by the SES section below.

---

## Where it stands

- **idea.uk is live** on a Hetzner box (Nuremberg, x86, IPv4 **116.203.204.115** — confirm),
  behind nginx + Let's Encrypt, as a systemd service. The landing page is embedded in the
  binary (`//go:embed page.html`), so changing the page means a rebuild + redeploy.
- It runs on a **Fake payment provider** — no Stripe keys yet, so it cannot take real money.
  That is the main gap between "live" and "earning".
- **Email works.** Outbound moved to **AWS SES** (London) and reaches the inbox, authenticated
  (DKIM `d=leopardess.uk`, DMARC + SPF pass). From is still `idea-uk@leopardess.uk`. See the SES
  section below.
- **Order flow has a switch.** `REVIEW_BEFORE_PAY` (default **on**) decides whether the engine
  runs and you approve the report **before** the buyer is billed, or the old charge-first order.
  See "Operator workflow" below.
- **The product pipeline is proven end-to-end** — a (fake) payment ran the engine and produced a
  full draft report. Only real payment (Stripe) is missing.
- **leopardess.uk** (the operator brand) has a one-page site, live on Clook.

---

## Email — current reality (AWS SES; supersedes all earlier Clook/MailChannels/465 notes)

- **Sender = AWS SES, London (`eu-west-2`).** SMTP `email-smtp.eu-west-2.amazonaws.com`, STARTTLS
  on **587** (Hetzner blocks 25/465/2525; 587 is open, which is why this works and `smtpSend`
  takes the `smtp.SendMail` STARTTLS path). From stays `idea-uk@leopardess.uk`, signed with
  leopardess.uk's own DKIM, so recipients see the same address — SES is just a more reliable pipe.
- **Why we left Clook.** Clook relays outbound via **MailChannels**, which blocked leopardess.uk's
  legitimate outbound wholesale — `550 5.7.1 [CS] Message blocked`, even a clean four-line message
  to one recipient. Not a wording or forwarding problem; the relay rejects the domain. So sending
  moved to SES. (Operator notifications to `idea-uk@` still deliver — `idea-uk@` is now a real
  Clook mailbox, so they land **locally on Clook**, no MailChannels; read them in Clook webmail.
  Do **not** set `idea-uk@` to forward onward, or it becomes outbound → MailChannels → blocked.)
- **SES setup (done):** domain `leopardess.uk` verified with Easy DKIM (3 CNAMEs in the Clook
  zone); account is in **production** (50k/day, not the 200/day sandbox), so it sends to any
  recipient. SMTP credentials created in eu-west-2.
- **The credential gotcha that cost time:** `535 Authentication Credentials Invalid` meant the env
  had `SMTP_USER` = the **IAM user name** (`ses-smtp-user.…`). SES authenticates with the **SMTP
  Username = the access key id (`AKIA…`)**, not the IAM user name. Fixing `SMTP_USER` to the
  `AKIA…` value cleared it. `SMTP_PASS` must be the **SMTP Password** SES shows on the create
  screen / CSV — not the IAM secret key.
- **Verified working:** a confirmation pay-link reached a real Workspace **inbox** (not spam),
  DKIM/DMARC/SPF pass.
- **Mailer code (deployed):** `makeDeliver` sends in a **goroutine** (a failed connect must not
  freeze the request path); it now declares **UTF-8** (`Content-Type: text/plain; charset=UTF-8`,
  `Content-Transfer-Encoding: 8bit`) and RFC 2047-encodes the Subject + From-name, which fixes the
  `â‰¥`/`â€` mojibake seen earlier. `smtpSend`'s 465 path keeps a 10s dial timeout + 30s deadline
  (unused on SES/587). Builds + `go vet` + tests clean (Go 1.22).

---

## Do this next (in order)

**1. Stripe (idea.uk £29-report billing)** — the step that turns "live" into "earning". Email is
now delivering, so the pay-link and receipt will reach buyers. Set `STRIPE_SECRET_KEY` +
`STRIPE_WEBHOOK_SECRET`; webhook `https://idea.uk/stripe/webhook` on `checkout.session.completed`;
own-card test; refund via `/refund`. (These steps live in `billing.go` + here — note
`PLAN_stripe_billing_integration.md` is a *separate* plan for the chassis build/host product, not
the £29 report.)

**2. Report + audience-check language and layout, and an optional PDF.** Both the paid report and
the free taster read too dense — a rewrite of the prompt output (`prompts.go`, `engine.go`,
`audience_check.go`) for clearer, less packed writing and nicer structure, plus an optional PDF of
the report (email attachment or link). Agreed as its own focused pass. (The UTF-8 mailer fix
already handles the garbled characters.)

**3. Form-abuse handling.** Bots hit the open `/request` form (injection probes, spam fills). Low
risk today — the flow is operator-gated, so nothing auto-charges or auto-emails a customer, and the
store is JSON + email is plaintext, so probes are inert. Add a honeypot field + light validation to
`handleRequest` to silently drop spam.

### Env block (`/etc/idea/idea.env`) — SES on 587

```
SMTP_HOST=email-smtp.eu-west-2.amazonaws.com
SMTP_PORT=587
SMTP_USER=<SES SMTP Username — the AKIA… access key id, NOT the ses-smtp-user.… IAM name>
SMTP_PASS=<SES SMTP Password from the create screen / CSV — NOT the IAM secret key>
SMTP_FROM=idea-uk@leopardess.uk
SMTP_FROM_NAME=idea.uk
SMTP_REPLY_TO=idea-uk@leopardess.uk
CONTACT_EMAIL=idea-uk@leopardess.uk
OPERATOR_EMAIL=idea-uk@leopardess.uk
REVIEW_BEFORE_PAY=true
```

`REVIEW_BEFORE_PAY=true` is review-before-pay (engine + your approval before the buyer is billed);
set it to `false` for the old charge-first order. The startup log prints `review_before_pay=…`.
Inbound: leopardess.uk **catch-all** (Default Address → Forward) → `aaa@designconsultancy.co.uk`.
Operator notifications go to `idea-uk@` (a real Clook mailbox now) — read them in **Clook webmail**;
don't forward `idea-uk@` onward or it becomes MailChannels-blocked outbound. Always check the
**leopardess.uk vs .co.uk** twin in the domain selector.

---

## Operator workflow (review-before-pay, the default)

Each request now needs two operator actions, both on the box (`X-Internal-Key` from the env):

```
KEY=$(grep '^INTERNAL_API_KEY=' /etc/idea/idea.env | cut -d= -f2)

# 1. confirm — runs the engine; you then get the draft by email to review
curl -s localhost:8080/confirm -H "X-Internal-Key: $KEY" -H 'content-type: application/json' -d '{"order_id":"ord_..."}'

# 2a. approve — bills the buyer (sends the pay-link); use once the draft looks good
curl -s localhost:8080/approve -H "X-Internal-Key: $KEY" -H 'content-type: application/json' -d '{"order_id":"ord_..."}'

# 2b. or decline — no charge
curl -s localhost:8080/decline -H "X-Internal-Key: $KEY" -H 'content-type: application/json' -d '{"order_id":"ord_...","reason":"weak differentiator"}'
```

`/confirm` → `running` → the draft arrives as a `[idea.uk] REVIEW …` email (in the `idea-uk@` Clook
mailbox); `/approve` → buyer gets the pay-link → on payment the **already-generated** report is sent
(no second engine run); `/decline` → buyer told, no charge. Newest pending order id:
`ssh root@116.203.204.115 "python3 -c \"import json;o=json.load(open('/var/lib/idea/orders.json'))['orders'];r=sorted([x for x in o.values() if x['status']=='requested'],key=lambda x:x['created_at']);print(r[-1]['id'], r[-1]['email'])\""`
With `REVIEW_BEFORE_PAY=false`, `/confirm` instead sends the pay-link straight away (charge first).

---

## Facts worth not re-deriving

- **Box:** Hetzner, Nuremberg, x86, IPv4 **116.203.204.115** (confirm). nginx does TLS + proxy;
  the Go binary serves everything incl. the embedded page. **Outbound SMTP from the box: 587
  only** (25/465/2525 blocked by Hetzner).
- **DNS:** idea.uk at **Hetzner DNS**; leopardess.uk at **Clook** (cPanel zone).
- **Hosting model:** one self-contained Go binary ("ship the binary"); env at `/etc/idea/idea.env`
  (systemd `EnvironmentFile` — comments on their own line, **never** inline after a value). Static
  chassis sites go git → GitHub Actions → Backblaze B2; idea.uk does not.
- **Build/redeploy:** `GOPROXY=off GOTOOLCHAIN=local`, target `GOOS=linux GOARCH=amd64`. Redeploy =
  the **`&&`-chained** build → scp → `mv -f` → `systemctl restart` (`mv -f` because the binary is
  busy; the chain stops a failed build from shipping a stale binary).
- **Operator domain:** leopardess.uk — neutral brand for system/transactional/support mail.
- **Email sender:** AWS SES, London (`eu-west-2`), STARTTLS 587, From `idea-uk@leopardess.uk` with
  leopardess.uk Easy DKIM; account in production. SES **SMTP Username = the `AKIA…` access key id**
  (not the `ses-smtp-user.…` IAM name); SMTP Password is the value from the create screen (not the
  IAM secret). MailChannels (Clook's relay) blocked leopardess.uk outbound — do not route customer
  mail back through Clook.
- **Pricing/product (locked):** **£29** full report + free 30-second taster. First vertical tool =
  **SFI26 Diff Alerts**. Engine method has a **Risk column** (gate Definition≥3 AND Willingness≥3).

---

## Backlog (deferred)

- **Stripe live mode** — the step that lets idea.uk take real money (see "Do this next" §3).
- **Solicitor review** of the Terms / Refund / Privacy drafts before taking real payments; fill
  the remaining `[bracketed]` Privacy items (transfer safeguards, retention period).
- **Email fold-in to the chassis:** the `email` aspect + a shared encoding function + the
  email-provisioner agent (per `EMAIL_identity_in_site_spec.md`). Use a **dedicated transactional
  sender (AWS SES or equivalent)**, not a shared cPanel relay — this session is the proof:
  MailChannels content-filtered legitimate outbound and SES fixed it. Prefer per-site DKIM
  identities as the number of live sites grows.
- **nginx hardening** (security headers, logrotate, geo-whitelist) — in the canonical
  `deploy/setup.sh`; the box has the first copy; applies on a `MODE=full` re-run.
- Phase A remainder (streaming progress page, refund polish, PII quote); Phase C (SFI26 Diff
  Alerts); Phase D (chassis-native engine); the service-deployer workflow (Layer 5).

---

## File map (all in `/mnt/user-data/outputs/`)

- `idea-go/` — the Go service. Key files: `service.go` (App/Config/NewApp, all HTTP handlers incl.
  `/confirm` `/approve` `/decline`, the review-before-pay flow via `fulfil` + `sendPayLink` +
  `deliverReport`, the `a.page()` styled wrapper, `writeHTML`, the mailer incl. `smtpSend` +
  async UTF-8 `makeDeliver`, policy-page constants), `engine.go` + `prompts.go` (the ideation
  method), `page.html` (embedded landing page — must stay in the module dir for `go:embed`),
  `audience_check.go`, `store.go`, `billing.go`, `main.go`, `service_test.go`, `go.mod`.
  `idea-go/deploy/` = `setup.sh`, `idea.env.example`, `README.md`.
- `running_notes.md` — the full cross-session journal (the memory; read the tail first).
- `idea_uk_architecture_and_deployment.md` — architecture + hosting + deploy; **email reality in
  the 2026-06-10 update** (earlier 465 sections superseded).
- `EMAIL_identity_in_site_spec.md` — the framework email design (per-site forwarders) + the
  2026-06-10 operational note (cloud-box ports, relay content-filtering).
- `RUNBOOK_idea_uk.md` — the service runbook (flow, local tests, go-live).
- `PLAN_stripe_billing_integration.md` — Stripe plan for the **chassis build/host product** (site-owner billing); **separate** from idea.uk's £29-report billing, which is inline in this handoff + `billing.go`.
- `leopardess_uk_index.html` — the leopardess.uk one-pager (live on Clook).
- `idea_uk_fakedoor.html` — standalone mirror of the landing page.
- Policy previews: `terms_preview.html`, `privacy_preview.html` (refund text is embedded in
  `service.go`).
- `016_debugging_guide_v2_32.md` — chassis pipeline guide (§11 = idea.uk page-serving + deploy).

---

## Working preferences (please keep to these)

Plain, matter-of-fact language — not LLM-speak (avoid honest/gate/deck/asset/leverage/robust/
seamless, "surface" as a verb, "X not Y" framing; punchy headlines are fine, prose stays plain).
No flattery or congratulating. **Confirm live API/schema/product facts before writing code or
asserting them — every time.** Reuse before rebuild. Honest caveats and pushback. Prefer
structural fixes over quick patches. British English. Low risk appetite (minimise liability).
Memory is off — keep `running_notes.md` current at each checkpoint.
