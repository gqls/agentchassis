# Context pack — idea.uk go-live: Stripe + email (fresh thread)

Starting context for finishing idea.uk. **One self-contained Go binary** with file-based persistence (no external DB), live on a Hetzner box behind nginx + Let's Encrypt as a systemd service. Separate from the chassis platform — its files live in the idea.uk outputs, **not** in the chassis project, so they must be attached.

---

## State + next action

idea.uk is **live** but on a **Fake payment provider** (no Stripe keys) — that's the one thing between "live" and "earning". Email is built and tested both ways but **not yet switched on in the service** (happens at the next redeploy). The landing page is embedded (`//go:embed page.html`), so page changes need a rebuild+redeploy. **One blocker right now:** the laptop's `service.go` is out of sync so the build fails — the code is fine, it builds cleanly from the canonical files; fixing the sync is step 1.

**Next action (in order):**
1. **Fix the build (laptop sync).** In the build dir, confirm `service.go` is present, ~658 lines, `package main`, has `func NewApp`; if missing/tiny/wrong, re-download the latest and overwrite. Confirm a clean build: `GOPROXY=off GOTOOLCHAIN=local GOOS=linux GOARCH=amd64 go build -o idea .`
2. **Flip the leopardess.uk catch-all to forwarding.** cPanel → Email → Default Address → domain **leopardess.uk** (not the .co.uk twin) → Forward to `aaa@designconsultancy.co.uk`.
3. **Set the email env** in `/etc/idea/idea.env` on the box (block below).
4. **Redeploy** (rebuild + binary swap with `mv -f`, not `cp` — the running binary is busy; the `&&` chain stops a failed build deploying a stale binary). Confirm box IP **116.203.204.115**. This also ships the privacy-policy correction and the "by leopardess.uk" footers.
5. **Test.** Place a test order / trigger a confirmation; check it **arrives** and the **From** reads `idea.uk <idea-uk@leopardess.uk>`. If "sender not allowed": set `SMTP_FROM=system@leopardess.uk`, keep `SMTP_REPLY_TO=idea-uk@leopardess.uk`, redeploy. If a TLS cert error: the cert may not match mail.leopardess.uk — one-line server-name fix in `smtpSend`.

## Standing rules (working preferences)

Plain, matter-of-fact language (no LLM-speak, no flattery). **Confirm live API/schema/product facts before writing code or asserting them — every time.** Reuse before rebuild. Structural fixes over quick patches. British English. Low risk appetite (minimise liability). Memory off — keep `running_notes.md` current at each checkpoint. Build env always `GOPROXY=off GOTOOLCHAIN=local`, target `GOOS=linux GOARCH=amd64`.

## Attach — code (download from the idea.uk outputs; NOT in the chassis project)

`idea-go/`: `service.go` (App/Config/NewApp, HTTP handlers, the `a.page()` wrapper, `writeHTML`, the mailer incl. `smtpSend`'s 465 implicit-TLS path, policy-page constants), `engine.go` + `prompts.go` (the ideation method), `audience_check.go` (the free taster), `store.go` (persistence), `billing.go` (Stripe / Fake provider), `main.go`, `service_test.go`, `page.html` (must sit beside the .go files for `go:embed`), `go.mod`, `deploy/` (`setup.sh`, `idea.env.example`, `README.md`).

## Attach — docs

`HANDOFF.md` (the idea.uk one), `running_notes.md` (read the tail first), `idea_uk_architecture_and_deployment.md`, `EMAIL_identity_in_site_spec.md`, `LIABILITY_AND_TERMS.md`, `016_debugging_guide_v2_32.md` (§11 = idea.uk page-serving + deploy gotchas), and for the next likely task **`PLAN_stripe_billing_integration.md`**.

## Schema / live data

**None** — file-based persistence via `store.go`. No SQL schema or table-content files needed.

## Email — the working setup (don't re-derive)

Outbound: Go service sends via Clook SMTP (mail.leopardess.uk:465, implicit TLS) auth `system@leopardess.uk`, From `idea-uk@leopardess.uk`; passes SPF/DKIM/DMARC at Gmail (Clook relays via MailChannels). Inbound: leopardess.uk catch-all forwards every address to Gmail `aaa@designconsultancy.co.uk`; per-site address encodes the domain (dots→dashes): idea.uk → `idea-uk@leopardess.uk`. Port 465 needed a code change (Go's stock mailer does STARTTLS, so `smtpSend` dials TLS directly for 465, STARTTLS for 587/25). **Watch out:** leopardess.**uk** and leopardess.**co.uk** are both on the cPanel account — easy to configure the wrong one; always check the domain selector; test inbound from a non-Gmail sender (Gmail dedupes its own messages).

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

## Backlog (deferred)

**Stripe live mode** (the earning step): set `STRIPE_SECRET_KEY` + `STRIPE_WEBHOOK_SECRET`; webhook `https://idea.uk/stripe/webhook` on `checkout.session.completed`; test with own card; refund via `/refund`. **Solicitor review** of Terms/Refund/Privacy before taking real payments; fill the remaining `[bracketed]` Privacy items. Email design fold-in to the chassis (email aspect + shared encoding fn + email-provisioner agent; prefer per-site forwarders as live sites grow). nginx hardening (security headers, logrotate, geo-whitelist) on a `MODE=full` re-run. Phase A remainder / Phase C (SFI26 Diff Alerts) / Phase D (chassis-native engine) / the service-deployer workflow.

## Facts worth not re-deriving

Box: Hetzner, Nuremberg, x86, IPv4 **116.203.204.115** (confirm); nginx does TLS + proxy, the Go binary serves everything incl. the embedded page. DNS: idea.uk at Hetzner DNS; leopardess.uk at Clook (cPanel). Hosting model: idea.uk ships **one binary**; static chassis sites go git → GitHub Actions → Backblaze B2 (idea.uk does **not**). Pricing (locked): £29 full report + free 30-second taster; first vertical tool = SFI26 Diff Alerts; engine has a Risk column (gate Definition≥3 AND Willingness≥3).

## Minimum set to start fast

For the go-live: `idea-go/` + `HANDOFF.md` + `running_notes.md` + the architecture and email docs + `PLAN_stripe_billing_integration.md`. That's enough.
