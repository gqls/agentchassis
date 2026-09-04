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
// # Two clicks, and why the GET does nothing
//
// An emailed confirmation link is fetched by things that are not the customer:
// corporate mail scanners, inbox link checkers, preview bots. Until 2026-08-24
// the link click WAS the confirmation, so any of those would have stamped the
// site as transfer-confirmed with nobody having decided anything — after which
// we stop chasing the customer and retract their live site on schedule,
// believing they had moved.
//
// Owner ruling 2026-08-24 (*"we can't have email scanners clicking the accept
// button so we'll need a separate page"*): the endpoint splits BY METHOD.
//
//	GET  /c/<token>  renders a page with one button. It reads nothing and
//	                 writes nothing — no database access of any kind — so a
//	                 scanner following the link cannot change our state, and
//	                 cannot learn whether the token is even real.
//	POST /c/<token>  is the confirmation. Scanners follow links; they do not
//	                 submit forms.
//
// The 2026-08-19 ruling that recording the click IS the state is AMENDED, not
// repealed: the BUTTON press is the state, the LINK click is navigation.
//
// SAME PATH, method as the distinction, pinned 2026-08-24. The box vhost admits
// exactly `^/c/[A-Za-z0-9_-]{20,128}$` and permits GET and POST on it, so a
// suffix route such as /c/<token>/confirm would be 404'd at the box and never
// reach this file. Change the vhost and the route together or not at all.
//
// The speculative-fetch guard below stays on BOTH handlers. On the POST it is
// what it always was. On the GET it is now belt to that braces — the ruling
// keeps it as the outer layer rather than resting the whole hazard on one
// mechanism, and it costs a header check.
//
// See DECISION_2026-08-24_confirmation_needs_a_second_click.md in the
// webdesign_uk_build_service lane for the ruling and what it does NOT change.
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
	// ZipDownloadURL redeems a zip_download token for its stored presigned URL.
	// delivery.ErrTokenNotFound -> the uniform failure page; delivery.ErrZipURLStale
	// (reachable only with a VALID token) -> the honest refresh page.
	ZipDownloadURL(c *gin.Context, token string) (string, error)
	// RecordStaleZipLink persists a stale hit where the fleet's sweeps read.
	// Best-effort: recording must never break rendering.
	RecordStaleZipLink(c *gin.Context)
	Logger() *zap.Logger
}

// NewDeliveryHandler creates the handler.
func NewDeliveryHandler(deps DeliveryDeps) *DeliveryHandler {
	return &DeliveryHandler{deps: deps}
}

// RegisterRoutes mounts the endpoint's two verbs.
//
// It exists because the router shape must have ONE definition. The guardian
// seat's objection on council ea99befa was exact: delivery_test.go held its own
// copy of the route table, "mirrors api/server.go" was discipline rather than a
// mechanism, and a later edit to one copy would leave the suite green while
// production diverged. Both callers now register through here, so the mirror
// cannot drift — there is nothing to keep in step.
//
// HEAD is deliberately NOT registered: gin does not route HEAD to a GET handler,
// and adding one would widen the public surface to make a test easier. The
// handlers keep their own HEAD refusal for the day someone reaches for r.Any(),
// and the tests drive that arm directly rather than through this function.
func (h *DeliveryHandler) RegisterRoutes(r gin.IRouter) {
	r.GET("/c/:token", h.HandleConfirmPage)
	r.POST("/c/:token", h.HandleConfirmTransfer)
	// GET only, deliberately: a download link is idempotent and scanners
	// following it can learn nothing and change nothing (LookupZipURL mutates
	// only use_count, which exists to measure exactly such visits).
	r.GET("/d/:token", h.HandleZipDownload)
}

// maxTokenLen bounds what we will even hash. Tokens are 32 random bytes in
// base64url = 43 characters; anything materially longer is not a token of ours,
// and refusing early keeps a hostile URL from reaching the database at all.
const maxTokenLen = 128

// HandleConfirmPage is GET /c/:token: the page with the button, and nothing else.
//
// IT PERFORMS NO DATABASE ACCESS. That is the whole mitigation, so it is worth
// stating as a property rather than describing it: there is no call here to
// spend a token, none to look one up, and none to ask whether one exists. A mail
// scanner that opens this link gets byte-identical HTML whatever it holds, which
// is also why this page cannot be used to test a guessed token without pressing
// the button.
//
// It follows that the page cannot say "that link is no longer active" for a
// token that is spent or expired: the customer learns that when they press. That
// cost was put to the owner and accepted (2026-08-25), against the alternative
// of a read-only lookup, which buys slightly better copy and hands anyone
// holding a guess a free validity oracle.
func (h *DeliveryHandler) HandleConfirmPage(c *gin.Context) {
	log := h.deps.Logger()

	// Both refusals are retained from the mutating days on the owner's ruling.
	// Nothing here mutates, so neither is load-bearing now; they are kept
	// because a guard removed on the reasoning "the thing it protects moved" is
	// how the next refactor gets a surprise.
	if c.Request.Method == http.MethodHead {
		log.Info("confirm page refused: HEAD")
		h.renderSpeculativeRefusal(c)
		return
	}
	if reason := speculativeFetchReason(c.Request); reason != "" {
		log.Info("confirm page refused: speculative fetch", zap.String("signal", reason))
		h.renderSpeculativeRefusal(c)
		return
	}

	// Same guard as the POST, on a handler that does not use the token: a
	// hostile URL must get the same answer from both verbs, or the pair of them
	// becomes the oracle that neither is on its own.
	token := strings.TrimSpace(c.Param("token"))
	if token == "" || len(token) > maxTokenLen {
		log.Info("confirm page refused: malformed token", zap.Int("len", len(token)))
		h.renderConfirm(c, confirmFailed)
		return
	}

	h.renderConfirm(c, confirmButton)
}

// HandleConfirmTransfer is POST /c/:token — the button press, and the only thing
// on this endpoint that changes state.
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

	// Never mutate for a request that announces itself as speculative, and never
	// for HEAD. Browsers do not speculatively POST, so on this verb the guard is
	// mostly defence against a future routing change — gin does not route HEAD
	// to a POST handler today, but that is one r.Any() away from being untrue,
	// and this is the file that must not depend on it.
	//
	// ⚠ WHAT CLOSES THE MAIL-SCANNER VECTOR IS THE METHOD SPLIT, NOT THIS
	// GUARD. A corporate security gateway sends a plain GET with none of these
	// headers and is indistinguishable here from a person; it now lands on
	// HandleConfirmPage, which changes nothing. Do not merge the two handlers
	// back together on the reasoning that this guard covers it. It does not, and
	// it never did (DECISION_2026-08-24, closing DECISION_2026-08-21b s4).
	if c.Request.Method == http.MethodHead {
		log.Info("confirm-transfer refused: HEAD, state not changed")
		h.renderSpeculativeRefusal(c)
		return
	}
	if reason := speculativeFetchReason(c.Request); reason != "" {
		// Logged so this is MEASURABLE. If it never fires we have learned that
		// browser prefetch is not a real vector for our mail; if it fires we
		// have learned it is, before a customer is affected. A guard nobody can
		// count is a guard nobody can review.
		log.Info("confirm-transfer refused: speculative fetch, state not changed",
			zap.String("signal", reason))
		h.renderSpeculativeRefusal(c)
		return
	}

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

// speculativeFetchReason returns the header that marks this request as a
// browser prefetch, prerender or link preview, or "" if it looks like a human
// following the link. The returned string is the SIGNAL NAME only, never a
// header value: values are attacker-supplied and this goes to the log.
//
// All four headers are checked because they are four different vendors' takes
// on the same idea and a browser sends whichever its generation used. Checking
// only Sec-Purpose (the current standard) would silently miss older Chrome,
// Firefox and Safari.
func speculativeFetchReason(r *http.Request) string {
	// Sec-Purpose: the Speculation Rules standard. Values seen in the wild
	// include "prefetch" and "prefetch;prerender", so substring, not equality.
	if v := strings.ToLower(r.Header.Get("Sec-Purpose")); strings.Contains(v, "prefetch") || strings.Contains(v, "prerender") {
		return "Sec-Purpose"
	}
	if strings.EqualFold(strings.TrimSpace(r.Header.Get("Purpose")), "prefetch") {
		return "Purpose"
	}
	switch strings.ToLower(strings.TrimSpace(r.Header.Get("X-Purpose"))) {
	case "preview", "prefetch":
		return "X-Purpose"
	}
	if strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Moz")), "prefetch") {
		return "X-Moz"
	}
	return ""
}

// renderSpeculativeRefusal answers a speculative fetch without changing
// anything.
//
// WHY A 4xx AND NOT A FRIENDLY 200. A 200 would be a valid prefetch result, and
// the browser may serve it back when the customer actually clicks — so the
// human would see a page while the click never reached us, which is the failure
// this guard exists to prevent, wearing a different hat. A non-2xx makes the
// browser discard the speculation and re-fetch on the real navigation, which is
// exactly what we want: the customer's own click then mutates normally, because
// it carries none of these headers.
//
// no-store belongs here for the same reason and is not redundant with the
// status: it is the half that does not depend on the browser agreeing with our
// choice of code.
func (h *DeliveryHandler) renderSpeculativeRefusal(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Header("Referrer-Policy", "no-referrer")
	c.Header("X-Content-Type-Options", "nosniff")
	c.String(http.StatusPreconditionFailed, "This link is confirmed by opening it yourself.\n")
}

type confirmOutcome int

const (
	confirmOK confirmOutcome = iota
	confirmFailed
	confirmError
	// confirmButton is the GET page: the only outcome reached without the
	// database having been asked anything.
	confirmButton
	// zipStale: the download token is VALID but its stored presign has aged
	// out. Honest, and deliberately different from failure: the holder of a
	// real link must never be told their files are gone.
	zipStale
	// zipFailed: the uniform download failure page. Same no-oracle rule as
	// confirmFailed: unknown, expired and revoked are one message.
	zipFailed
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
 button{font:inherit;margin-top:.5rem;padding:.7rem 1.25rem;border:0;
        border-radius:.4rem;background:#111;color:#fff;cursor:pointer}
 @media (prefers-color-scheme:dark){body{color:#eee;background:#111}
                                    button{background:#eee;color:#111}}
</style>
<h1>{{.Title}}</h1>
{{range .Paragraphs}}<p>{{.}}</p>
{{end}}{{if .Button}}<form method="post"><button type="submit">{{.Button}}</button></form>
{{end}}`))

type confirmView struct {
	Title      string
	Paragraphs []string
	// Button, when set, renders the form that POSTs back to this same URL.
	//
	// The form deliberately carries NO action attribute: a form without one
	// submits to the current document URL, so the token stays in the request
	// line where it already is and never enters the HTML. action="/c/<token>"
	// would write the secret into the page body, and from there into anything
	// that keeps a copy of the page.
	Button string
}

// HandleZipDownload is GET /d/:token: redeem the token, 302 to the stored
// presigned URL while it is fresh, and never redirect to a stale one (an
// expired presign answers 403 SignatureDoesNotMatch, which reads as broken
// credentials, not an old link).
func (h *DeliveryHandler) HandleZipDownload(c *gin.Context) {
	log := h.deps.Logger()

	token := strings.TrimSpace(c.Param("token"))
	if token == "" || len(token) > maxTokenLen {
		log.Info("zip-download refused: malformed token", zap.Int("len", len(token)))
		h.renderConfirm(c, zipFailed)
		return
	}

	url, err := h.deps.ZipDownloadURL(c, token)
	switch {
	case err == nil:
		log.Info("zip download redirected")
		// 302, not 301: the target URL changes every refresh cycle and nothing
		// downstream may cache the mapping.
		c.Redirect(http.StatusFound, url)
	case errors.Is(err, delivery.ErrZipURLStale):
		log.Warn("zip download link stale: rendering refresh page and recording")
		h.deps.RecordStaleZipLink(c)
		h.renderConfirm(c, zipStale)
	case errors.Is(err, delivery.ErrTokenNotFound):
		log.Info("zip download refused: token not valid")
		h.renderConfirm(c, zipFailed)
	default:
		log.Error("zip download lookup failed", zap.Error(err))
		h.renderConfirm(c, confirmError)
	}
}

func (h *DeliveryHandler) renderConfirm(c *gin.Context, outcome confirmOutcome) {
	var v confirmView
	switch outcome {
	case confirmOK:
		v = confirmView{
			Title: "Thank you, that is recorded.",
			Paragraphs: []string{
				"We have noted that you have moved everything across.",
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
	case confirmButton:
		// No em dashes, no external reference, nothing named. The page says what
		// pressing does and that nothing has happened yet, because a customer
		// who thinks the link already did it will not press.
		//
		// ⚠ IT SAYS ONLY WHAT PRESSING RECORDS, AND THAT IS A CORRECTION
		// (bugs_open/477, 2026-09-04). This page and the success page above both
		// used to end on "You will not get any more reminders about it." Nothing
		// in the estate sends a reminder: exactly one agent can send mail
		// (delivery-email-sender), no scheduled task targets it, and
		// transfer_confirmed_at, the record this button writes, had no reader
		// outside platform/delivery/handover.go. The sentence was DELETED rather
		// than reworded, because the button's stated motivation was the false
		// part and inventing a replacement repeats the defect one wording along.
		//
		// RESTORE IT when the follow-up sender exists and this stamp suppresses
		// it, and delete the test that guards this
		// (TestNoConfirmPagePromisesRemindersStopWhileNothingSendsThem) in the
		// same commit. The test is what makes that a deliberate act.
		v = confirmView{
			Title: "Confirm you have moved your site",
			Paragraphs: []string{
				"Pressing the button below tells us you have moved everything across.",
				"Nothing is recorded until you press it.",
			},
			Button: "Yes, I have moved everything",
		}
	case zipStale:
		// The holder of a REAL link. Their files are safe; the link's inner
		// plumbing has aged and the refresher has not caught up. Honest about
		// the wait, silent about the mechanism, and it invites the retry that
		// will succeed once re-stamped rather than a support email.
		v = confirmView{
			Title: "Your download link is being refreshed.",
			Paragraphs: []string{
				"Your files are safe. This link renews itself from time to time, and it is mid renewal right now.",
				"Please try the same link again a little later. It is the same link; you do not need a new one.",
			},
		}
	case zipFailed:
		// The uniform download failure. Same no-oracle rule as confirmFailed:
		// unknown, expired and revoked are one message.
		v = confirmView{
			Title: "That download link is no longer active.",
			Paragraphs: []string{
				"It may have run out. There is nothing wrong at your end.",
				"If you still need your files, reply to the email your link arrived in.",
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
