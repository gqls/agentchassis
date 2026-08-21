// FILE: internal/core-manager/handlers/delivery.go
//
// The customer-facing half of handover: the link a customer clicks in the
// delivery email to tell us their site (or their bought domain) has moved off
// our infrastructure. Recording that click IS the state — owner ruling
// 2026-08-19: no form, no reply parsing — and it is what stops the weekly chase
// email.
//
// # Why this is in core-manager and not somewhere public-looking
//
// There is NO Ingress in this cluster (checked 2026-08-21: `kubectl get ingress
// -A` is empty). Customer traffic arrives at the webdesign.uk box, whose nginx
// listens on loopback behind a cloudflared tunnel and proxies named paths to
// cluster services over WireGuard by cluster-DNS name. `/stripe/webhook` ->
// auth-service is the proven instance of that shape; this is the second.
//
// So the exposure added here is EXACTLY the paths nginx names, never
// core-manager as a whole. That distinction is load-bearing and it is why the
// nginx block uses `location =` / a bounded prefix rather than a catch-all: the
// site-facts relay two files over reasons explicitly about core-manager not being
// publicly fronted, and this must not quietly make that false.
//
// # Why there is no /d/<token> here, and it is not an omission
//
// The Phase 4 plan pairs this with `/d/<token>`, which was specified to "mint a
// clamped presign and redirect". IT CANNOT BE BUILT IN THIS SERVICE, and the
// reason is a standing owner directive rather than a technical gap.
//
// Object-store credentials in this estate live ONLY in spawned pods. They were
// REMOVED from the standing agent-chassis on 2026-08-11 under `bugs_open/245`,
// on the owner's directive ("the agent chassis shouldn't carry b2 credentials"),
// and the spawn path now injects secretKeyRef references into spawned pod specs
// instead. Measured 2026-08-21: agent-chassis, auth-service, core-manager and
// admin-dashboard all have no B2 credentials in their pod env, and the only
// manifests carrying them are adapters, the database-backup CronJob, and spawn
// templates.
//
// Presigning inside core-manager would therefore mean giving a standing service
// the credentials another standing service had deliberately taken away ten days
// earlier. That is an architecture decision with a real blast radius, not an
// implementation detail, so it is written up and put to the owner rather than
// decided here. See the webdesign_uk_build_service lane's
// DECISION_2026-08-21_zip_download_link_needs_a_credential_home.md.
//
// # The prefetch hazard, stated because it is a real one
//
// This is a GET that mutates state, which is what an emailed confirmation link
// has to be if there is to be no form. Mail scanners and link-preview bots fetch
// links in email. Such a fetch would stamp the site as transfer-confirmed
// without the customer having clicked anything, and — IF the token were minted
// single-use — would also spend the token so the customer's own click then
// fails. The handler cannot prevent the first; it is written so the second
// cannot happen if the minting step follows the guidance in ConfirmTransfer's
// doc comment. The mitigation and its cost are in the decision doc; the ruling
// stands until the owner changes it.
package handlers

import (
	"errors"
	"html/template"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/delivery"
)

// DeliveryHandler serves the customer-facing handover links.
type DeliveryHandler struct {
	deps DeliveryDeps
}

// DeliveryDeps is the handler's dependencies, as an interface so a test can
// drive the HTTP layer without a database. The package's own tests cover the
// SQL; these cover routing, method, status and what a human is shown.
type DeliveryDeps interface {
	ConfirmTransfer(c *gin.Context, token string) error
	Logger() *zap.Logger
}

// NewDeliveryHandler creates the handler.
func NewDeliveryHandler(deps DeliveryDeps) *DeliveryHandler {
	return &DeliveryHandler{deps: deps}
}

// maxTokenLen bounds what we will even hash. Tokens are 32 random bytes in
// base64url = 43 characters; anything materially longer is not a token of ours,
// and refusing early keeps a hostile URL from reaching the database at all.
const maxTokenLen = 128

// HandleConfirmTransfer is GET /c/:token.
//
// 200 on success AND on every failure, with different HTML. Deliberate: the
// reader is a customer, not an API client, and a 404 page from a link we sent
// them reads as "we have lost your site". The distinction that matters for
// security is that the failure page NEVER says which failure it was — unknown,
// expired, revoked, already spent and wrong-purpose are one message, because
// telling an attacker which guess was closer is the only thing distinguishing
// them would achieve. That is the same rule delivery.ErrTokenNotFound encodes.
func (h *DeliveryHandler) HandleConfirmTransfer(c *gin.Context) {
	log := h.deps.Logger()

	token := strings.TrimSpace(c.Param("token"))
	if token == "" || len(token) > maxTokenLen {
		// NOT logged with the token in it. An emailed link ends up in access
		// logs, Referer headers and browser history already; this process is
		// not going to be a fourth copy.
		log.Info("confirm-transfer refused: malformed token", zap.Int("len", len(token)))
		h.renderConfirm(c, confirmFailed)
		return
	}

	err := h.deps.ConfirmTransfer(c, token)
	switch {
	case err == nil:
		log.Info("transfer confirmed by customer click")
		h.renderConfirm(c, confirmOK)
	case errors.Is(err, delivery.ErrTokenNotFound):
		log.Info("confirm-transfer refused: token not valid")
		h.renderConfirm(c, confirmFailed)
	default:
		// A database error is OURS, and the customer must not be told their
		// link is invalid when it is not: that would send them to an inbox we
		// do not staff (no pre-sales service) about a problem they cannot fix.
		log.Error("confirm-transfer failed", zap.Error(err))
		h.renderConfirm(c, confirmError)
	}
}

type confirmOutcome int

const (
	confirmOK confirmOutcome = iota
	confirmFailed
	confirmError
)

// confirmPage is deliberately self-contained: no external stylesheet, no font,
// no script. It is served from a service with no CDN in front of it to a
// customer who has just been emailed, and every external reference is a way for
// this page to look broken on someone else's bad day.
//
// The copy follows the register's voice rules (plain, direct, British English,
// NO EM DASHES — the ban is armed and the writer_block's first rule) and it
// makes no commercial claim, so it needs no evidence_base fact. It deliberately
// does not invite a reply: there is no pre-sales or support service.
var confirmPage = template.Must(template.New("confirm").Parse(
	`<!doctype html>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<meta name="robots" content="noindex,nofollow">
<title>{{.Title}}</title>
<style>
 body{font:16px/1.6 system-ui,-apple-system,"Segoe UI",Roboto,sans-serif;
      max-width:34rem;margin:12vh auto;padding:0 1.25rem;color:#111;background:#fff}
 h1{font-size:1.4rem;margin:0 0 .75rem}
 p{margin:0 0 .75rem}
 @media (prefers-color-scheme:dark){body{color:#eee;background:#111}}
</style>
<h1>{{.Title}}</h1>
{{range .Paragraphs}}<p>{{.}}</p>
{{end}}`))

type confirmView struct {
	Title      string
	Paragraphs []string
}

func (h *DeliveryHandler) renderConfirm(c *gin.Context, outcome confirmOutcome) {
	var v confirmView
	switch outcome {
	case confirmOK:
		v = confirmView{
			Title: "Thank you, that is recorded.",
			Paragraphs: []string{
				"We have noted that you have moved everything across. You will not get any more reminders about it.",
				"Your site files are yours to keep, and so is anything you have bought outright.",
			},
		}
	case confirmFailed:
		v = confirmView{
			Title: "That link is no longer active.",
			Paragraphs: []string{
				"It may already have been used, or it may have run out. Either way there is nothing wrong at your end.",
				"If you have already told us you have moved, you do not need to do anything else.",
			},
		}
	default:
		v = confirmView{
			Title: "Something went wrong at our end.",
			Paragraphs: []string{
				"Your link is fine. Please try it again in a few minutes.",
			},
		}
	}
	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/html; charset=utf-8")
	// Never cached: the page's content depends on state this request changed,
	// and a tunnel or a corporate proxy caching the success page would show it
	// to the next person to click any link on this host.
	c.Header("Cache-Control", "no-store")
	// The token is in the path. Without this, it travels to any resource the
	// page references and into any analytics the customer's browser runs.
	c.Header("Referrer-Policy", "no-referrer")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-Frame-Options", "DENY")
	if err := confirmPage.Execute(c.Writer, v); err != nil {
		h.deps.Logger().Error("confirm page render failed", zap.Error(err))
	}
}
