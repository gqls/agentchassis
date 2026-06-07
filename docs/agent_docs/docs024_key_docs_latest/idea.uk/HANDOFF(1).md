# idea.uk — handoff (2026-06-06)

The one-page picture of where idea.uk is and exactly what to do next. For the full
cross-session journal see `running_notes.md`; for architecture/deployment detail see
`idea_uk_architecture_and_deployment.md`.

---

## Where it stands

- **idea.uk is live** on a Hetzner box (Nuremberg, x86), behind nginx + Let's Encrypt, as a
  systemd service. The landing page is embedded in the Go binary (`//go:embed page.html`),
  so changing the page means rebuilding and redeploying.
- It runs on a **Fake payment provider** — no Stripe keys yet, so it cannot take real money.
  That is the main thing standing between "live" and "earning".
- **Email is built and tested both ways** (see below). It is not yet switched on *in the
  service* — that happens at the next redeploy.
- **leopardess.uk** (the operator brand; "idea.uk by leopardess.uk") has a one-page site,
  live on Clook.
- **One blocker right now:** the laptop's local `service.go` is out of sync, so the build
  fails. Nothing is wrong with the code — it builds cleanly from the canonical files. Fixing
  the sync is step 1 below.

---

## Do this next (in order)

**1. Fix the build (laptop sync).** In the build directory
`~/projects/agentchassis/docs/agent_docs/docs024_key_docs_latest/idea.uk/golang_files`:

```
ls -la service.go          # present? not 0 bytes?
head -1 service.go         # must say: package main
wc -l service.go           # should be ~658
grep -c "func NewApp" service.go   # should be 1
ls -1 *.go                 # expect 8: audience_check billing engine main prompts service service_test store
```

If service.go is missing/tiny/wrong, re-download the latest `service.go` and overwrite it.
Then confirm a clean build (no output = success):

```
GOPROXY=off GOTOOLCHAIN=local GOOS=linux GOARCH=amd64 go build -o idea .
```

**2. Flip the leopardess.uk catch-all back to forwarding.** It was left on the system@ POP
route. cPanel → Email → **Default Address** → domain **leopardess.uk** (not the .co.uk twin)
→ **Forward to Email Address** → `aaa@designconsultancy.co.uk` → Save. Without this, inbound
sits in the cPanel mailbox instead of reaching Gmail.

**3. Set the email env** in `/etc/idea/idea.env` on the box (full block below).

**4. Redeploy** (rebuild + binary swap — the page is embedded, so a rebuild is required):

```
GOPROXY=off GOTOOLCHAIN=local GOOS=linux GOARCH=amd64 go build -o idea . && \
scp idea root@116.203.204.115:/opt/idea/idea.new && \
ssh root@116.203.204.115 'chmod 755 /opt/idea/idea.new && mv -f /opt/idea/idea.new /opt/idea/idea && systemctl restart idea'
```

(`mv -f`, not `cp` — the running binary is busy. The `&&` chain stops a failed build from
deploying a stale binary. Confirm the box IP is still **116.203.204.115**.)

This redeploy also ships the **privacy-policy correction** (Clook, not SES) and the
**"by leopardess.uk"** footers.

**5. Test.** Place a test order / trigger a confirmation, then check:
- it **arrives**, and
- the **From** reads `idea.uk <idea-uk@leopardess.uk>`.

If the send errors **"sender not allowed"**: set `SMTP_FROM=system@leopardess.uk` and keep
`SMTP_REPLY_TO=idea-uk@leopardess.uk` (replies still route via the catch-all), redeploy.
If the first send throws a **TLS certificate error**: the cert may not match
mail.leopardess.uk — note the exact message; it's a one-line server-name fix in `smtpSend`.

---

## Email — the working setup

- **Outbound:** the Go service sends via Clook SMTP (**mail.leopardess.uk:465, implicit
  TLS**) authenticating as `system@leopardess.uk`, From `idea-uk@leopardess.uk`. Confirmed to
  pass SPF, DKIM (aligned to leopardess.uk), DMARC at Gmail. Clook relays via MailChannels.
- **Inbound:** leopardess.uk **catch-all** (Default Address → forward) sends every address to
  the Gmail `aaa@designconsultancy.co.uk`. Per-site address is a deterministic encoding of
  the domain (dots → dashes): idea.uk → `idea-uk@leopardess.uk`.
- **Port 465 needed a code change:** Go's stock mailer does STARTTLS, not implicit TLS, so
  `service.go`'s `smtpSend` dials TLS directly for 465 (and still does STARTTLS for 587/25).
- **Parked (Workspace admin-gated):** "Check mail from other accounts" (POP) and "Send mail
  as" external SMTP are disabled on the Workspace account, so *personal* replies go out as
  aaa@designconsultancy.co.uk. Cosmetic; idea.uk's automated mail is unaffected.
- **Watch out:** leopardess.**uk** and leopardess.**co.uk** are both on the cPanel account —
  it is easy to configure the wrong one. Always check the domain selector. Test inbound from
  a non-Gmail sender (Gmail dedupes its own messages, so self-tests look like they vanished).

### Env block (`/etc/idea/idea.env`)

```
SMTP_HOST=mail.leopardess.uk
SMTP_PORT=465
SMTP_USER=system@leopardess.uk
SMTP_PASS=<the system@ mailbox password>
SMTP_FROM=idea-uk@leopardess.uk
SMTP_FROM_NAME=idea.uk
SMTP_REPLY_TO=idea-uk@leopardess.uk
CONTACT_EMAIL=idea-uk@leopardess.uk
OPERATOR_EMAIL=idea-uk@leopardess.uk
```

---

## Facts worth not re-deriving

- **Box:** Hetzner, Nuremberg, x86, IPv4 **116.203.204.115** (confirm). nginx does TLS +
  proxy; the Go binary serves everything including the embedded landing page.
- **DNS:** idea.uk at **Hetzner DNS**; leopardess.uk at **Clook** (cPanel zone).
- **Hosting model:** one self-contained Go binary ("ship the binary"). Static sites for the
  wider chassis go git → GitHub Actions → Backblaze B2; idea.uk does not.
- **Build env always:** `GOPROXY=off GOTOOLCHAIN=local`, target `GOOS=linux GOARCH=amd64`.
- **Operator domain:** leopardess.uk — neutral brand for all sites' system/transactional/
  support mail, not a bulk sender for the lead-gen long tail.
- **Pricing/product (locked):** idea.uk = **£29 full report + free 30-second taster**. First
  vertical tool = **SFI26 Diff Alerts**. Engine method has a **Risk column** (6th factor;
  gate Definition≥3 AND Willingness≥3).

---

## Backlog (deferred)

- **Stripe live mode** — the step that lets idea.uk take real money. Set
  `STRIPE_SECRET_KEY` + `STRIPE_WEBHOOK_SECRET`; webhook `https://idea.uk/stripe/webhook`
  on `checkout.session.completed`; test with own card; refund via `/refund`.
- **Solicitor review** of the Terms / Refund / Privacy drafts before taking real payments;
  fill the remaining `[bracketed]` items in Privacy (transfer safeguards, retention period).
- **Email design fold-in:** add the `email` aspect + a shared encoding function + the
  email-provisioner agent to the chassis (per EMAIL_identity_in_site_spec.md). Prefer
  per-site forwarders over the catch-all as the number of live sites grows.
- **nginx hardening** (security headers, logrotate, geo-whitelist) is in the canonical
  `deploy/setup.sh` but the box still has the first copy — applies on a `MODE=full` re-run.
- Phase A remainder (streaming progress page, refund polish, PII quote); Phase C (SFI26 Diff
  Alerts); Phase D (chassis-native engine); the service-deployer workflow (Layer 5).

---

## File map (all in `/mnt/user-data/outputs/`)

- `idea-go/` — the Go service. Key files: `service.go` (App/Config/NewApp, all HTTP
  handlers, the `a.page()` styled wrapper, `writeHTML`, the mailer incl. `smtpSend`,
  policy-page constants), `engine.go` + `prompts.go` (the ideation method), `page.html` (the
  embedded landing page — must stay in the module dir for `go:embed`), `audience_check.go`,
  `store.go`, `billing.go`, `main.go`, `service_test.go`, `go.mod`. `idea-go/deploy/` =
  `setup.sh`, `idea.env.example` (Clook + 465), `README.md`.
- `running_notes.md` — the full cross-session journal (the memory; read the tail first).
- `idea_uk_architecture_and_deployment.md` — architecture + hosting + deploy; email
  live-state in the 2026-06-06 update.
- `EMAIL_identity_in_site_spec.md` — the framework email design (per-site forwarders).
- `leopardess_uk_index.html` — the leopardess.uk one-pager (live on Clook).
- `idea_uk_fakedoor.html` — standalone mirror of the landing page.
- Policy previews: `terms_preview.html`, `refund_policy_preview.html`, `privacy_preview.html`.
- `016_debugging_guide_v2_32.md` — chassis pipeline guide (§11 = idea.uk page-serving +
  deploy gotchas).

---

## Working preferences (please keep to these)

Plain, matter-of-fact language — not LLM-speak (avoid honest/gate/deck/asset/leverage/
robust/seamless, "surface" as a verb, "X not Y" framing; punchy headlines are fine, prose
stays plain). No flattery or congratulating. **Confirm live API/schema/product facts before
writing code or asserting them — every time.** Reuse before rebuild. Honest caveats and
pushback. Prefer structural fixes over quick patches. British English. Low risk appetite
(minimise liability). Memory is off — keep `running_notes.md` current at each checkpoint.
