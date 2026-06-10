# idea.uk — handoff (2026-06-10)

The one-page picture of where idea.uk is and exactly what to do next. Full cross-session
journal: `running_notes.md`. Architecture/deploy detail: `idea_uk_architecture_and_deployment.md`
(read the **2026-06-10 update** for the current email reality — earlier sections say 465, which
is superseded).

---

## Where it stands

- **idea.uk is live** on a Hetzner box (Nuremberg, x86, IPv4 **116.203.204.115** — confirm),
  behind nginx + Let's Encrypt, as a systemd service. The landing page is embedded in the
  binary (`//go:embed page.html`), so changing the page means a rebuild + redeploy.
- It runs on a **Fake payment provider** — no Stripe keys yet, so it cannot take real money.
  That is the main gap between "live" and "earning".
- **Email:** the mailer is built and **deployed**, and the **service→Clook submission works on
  port 587**. The current open item is **MailChannels content-filtering** (below) — one test
  is outstanding before email is proven end-to-end to a real recipient.
- **leopardess.uk** (the operator brand) has a one-page site, live on Clook.

---

## Email — current reality (supersedes the earlier "465 works" note)

- **Port is 587, not 465.** A port sweep from the box showed Hetzner blocks outbound 25/465/2525;
  **587 (submission) is open**. The cPanel "Connect Devices" page advertises 465, but 465 is
  unreachable from the box (the send timed out). So `SMTP_PORT=587`, and `smtpSend` takes the
  `smtp.SendMail` STARTTLS path. The 465 implicit-TLS branch stays in the code for hosts that
  need it, but is unused here.
- **Service→Clook works.** A test submission authenticated (`dovecot_plain`) as
  `system@leopardess.uk`, From `idea-uk@leopardess.uk`, and Clook accepted it — confirmed in the
  cPanel delivery report, with no error in the journal.
- **The open item — MailChannels content filter.** Clook relays outbound via MailChannels, which
  spam-filters. The operator "NEW REQUEST" notification (a *forward*: `idea-uk@` → the Gmail) was
  rejected `550 5.7.1 [CS]` — MailChannels Insights shows **"Blocked (Spam Content)"**. Its body
  embeds a `From: name <email>` line (reads as a forged header) and raw `POST /confirm {json}` —
  both trip content filters. The **buyer confirmation** body is clean prose + a link, so it
  should fare better; that's the next test.
- **Mailer code (deployed):** `makeDeliver` now sends in a **goroutine** (it was inline, so a
  failed connect froze the request path ~2 min); `smtpSend`'s 465 path got a **10s dial timeout +
  30s conn deadline**. Added the `net` import. Builds + `go vet` clean (Go 1.22).

---

## Do this next (in order)

**1. Test the customer path.** Place a request using a real **Gmail** you control as the buyer
email, run the operator confirm against that order, and watch the cPanel delivery report for the
message addressed to **that Gmail** (the pay-link). Also click **"Not Spam"** on the blocked
notification — it trains MailChannels on the false positive.

**2. Branch on the result.**
- **If the buyer confirmation lands** → only the operator notification is being filtered. Tidy
  that one body: relabel `From:` → `Requester:`, and replace the `POST /… {json}` lines with a
  plain sentence + the order id. One more redeploy and email is done.
- **If the buyer confirmation is also blocked** as spam content → MailChannels is filtering
  legitimate outbound. Ask Clook to relax outbound filtering for the account, or move
  transactional mail to a dedicated sender over 443 (the more dependable route for a service).
  Don't keep patching bodies — escalate to one of those.

**3. Then Stripe (idea.uk £29-report billing)** — the step that lets idea.uk take real money.
Confirmation/receipt emails must be delivering first, since the buyer gets the pay-link and
receipt by email. Set `STRIPE_SECRET_KEY` + `STRIPE_WEBHOOK_SECRET`; webhook
`https://idea.uk/stripe/webhook` on `checkout.session.completed`; own-card test; refund via
`/refund`. (These steps live in `billing.go` + here — note `PLAN_stripe_billing_integration.md`
is a *separate* plan for the chassis build/host product, not the £29 report.)

### Env block (`/etc/idea/idea.env`) — note 587

```
SMTP_HOST=mail.leopardess.uk
SMTP_PORT=587
SMTP_USER=system@leopardess.uk
SMTP_PASS=<the system@ mailbox password>
SMTP_FROM=idea-uk@leopardess.uk
SMTP_FROM_NAME=idea.uk
SMTP_REPLY_TO=idea-uk@leopardess.uk
CONTACT_EMAIL=idea-uk@leopardess.uk
OPERATOR_EMAIL=idea-uk@leopardess.uk
```

If a send errors **"sender not allowed"**: set `SMTP_FROM=system@leopardess.uk` and keep
`SMTP_REPLY_TO=idea-uk@leopardess.uk` (replies still route via the catch-all), then redeploy.
Inbound: leopardess.uk **catch-all** (Default Address → Forward) → `aaa@designconsultancy.co.uk`
(currently set correctly). Always check the **leopardess.uk vs .co.uk** twin in the domain
selector. Test inbound from a **non-Gmail** sender (Gmail dedupes its own self-sends).

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
- **Pricing/product (locked):** **£29** full report + free 30-second taster. First vertical tool =
  **SFI26 Diff Alerts**. Engine method has a **Risk column** (gate Definition≥3 AND Willingness≥3).

---

## Backlog (deferred)

- **Stripe live mode** — the step that lets idea.uk take real money (see "Do this next" §3).
- **Solicitor review** of the Terms / Refund / Privacy drafts before taking real payments; fill
  the remaining `[bracketed]` Privacy items (transfer safeguards, retention period).
- **Email fold-in to the chassis:** the `email` aspect + a shared encoding function + the
  email-provisioner agent (per `EMAIL_identity_in_site_spec.md`); prefer per-site forwarders over
  the catch-all as the number of live sites grows; and plan transactional sending that won't be
  content-filtered by a shared relay (this session showed why).
- **nginx hardening** (security headers, logrotate, geo-whitelist) — in the canonical
  `deploy/setup.sh`; the box has the first copy; applies on a `MODE=full` re-run.
- Phase A remainder (streaming progress page, refund polish, PII quote); Phase C (SFI26 Diff
  Alerts); Phase D (chassis-native engine); the service-deployer workflow (Layer 5).

---

## File map (all in `/mnt/user-data/outputs/`)

- `idea-go/` — the Go service. Key files: `service.go` (App/Config/NewApp, all HTTP handlers,
  the `a.page()` styled wrapper, `writeHTML`, the mailer incl. `smtpSend` + async `makeDeliver`,
  policy-page constants), `engine.go` + `prompts.go` (the ideation method), `page.html` (embedded
  landing page — must stay in the module dir for `go:embed`), `audience_check.go`, `store.go`,
  `billing.go`, `main.go`, `service_test.go`, `go.mod`. `idea-go/deploy/` = `setup.sh`,
  `idea.env.example`, `README.md`.
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
