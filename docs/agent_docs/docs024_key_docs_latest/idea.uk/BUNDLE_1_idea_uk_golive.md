# Bundle 1 — idea.uk go-live (next)

**What this chat is for:** finish idea.uk and take it from "live on a fake payment provider"
to "earning". Two steps: (1) switch email on (the in-flight redeploy), (2) Stripe live mode.

idea.uk is one self-contained Go binary with file-based persistence (`store.go`).
**No SQL schema or table-content files are needed.**

---

## Attach — start here
- `HANDOFF.md`
- `running_notes.md`

## Attach — the live code (download from `outputs/idea-go/`; not in your chassis project)
- `service.go`, `engine.go`, `prompts.go`, `audience_check.go`, `store.go`, `billing.go`,
  `main.go`, `service_test.go`
- `page.html` (the embedded landing page — must sit beside the .go files for `go:embed`)
- `go.mod`
- `deploy/setup.sh`, `deploy/idea.env.example`, `deploy/README.md`

## Attach — docs (download from `outputs/`)
- `idea_uk_architecture_and_deployment.md` — architecture, hosting, deploy, email live-state
- `EMAIL_identity_in_site_spec.md` — email design
- `LIABILITY_AND_TERMS.md` — legal posture (solicitor review still pending)
- `016_debugging_guide_v2_32.md` — §11 covers idea.uk page-serving + deploy gotchas
- `PLAN_stripe_billing_integration.md` — the Stripe step (taking real money)

## Optional supporting
- `RUNBOOK_idea_uk.md`, `DEVELOPMENT_RUNBOOK.md` — the phased plan
- `terms_preview.html`, `refund_policy_preview.html`, `privacy_preview.html` (same content is
  embedded in `service.go`)

---

## The immediate sequence (already in HANDOFF.md)
1. Fix the laptop `golang_files/service.go` sync (it should be the 658-line canonical file),
   confirm a clean `go build -o idea .`.
2. cPanel → Default Address for **leopardess.uk** → Forward to `aaa@designconsultancy.co.uk`
   (it was left on the system@ POP route).
3. Set the email env in `/etc/idea/idea.env` (block below).
4. Redeploy (rebuild + `mv -f` binary swap; `&&`-chain it).
5. Test order → check it arrives AND From = `idea.uk <idea-uk@leopardess.uk>`.
6. Then Stripe: set `STRIPE_SECRET_KEY` + `STRIPE_WEBHOOK_SECRET`, webhook
   `https://idea.uk/stripe/webhook` on `checkout.session.completed`, own-card test, refund
   via `/refund`.

## Facts not to re-derive
- Box: Hetzner, Nuremberg, x86, **116.203.204.115** (confirm). nginx = TLS + proxy; the Go
  binary serves everything incl. the embedded page (so page changes need a rebuild).
- Build env: `GOPROXY=off GOTOOLCHAIN=local`, target `GOOS=linux GOARCH=amd64`.
- Email: Clook SMTP **mail.leopardess.uk:465 (implicit TLS)**, auth `system@leopardess.uk`,
  send as `idea-uk@leopardess.uk`. The mailer's `smtpSend` already handles 465. The one
  unproven bit: sending From a different local part than the login — if it errors "sender not
  allowed", set `SMTP_FROM=system@leopardess.uk` and keep `SMTP_REPLY_TO=idea-uk@`.
- **Currently on FakeProvider** — Stripe keys are what turn on real payments.

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
