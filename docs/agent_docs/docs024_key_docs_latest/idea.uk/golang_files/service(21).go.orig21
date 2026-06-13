package main

// service.go — the idea.uk service. Two front doors, one engine:
//   external: request → operator confirm/decline → pay → run → deliver
//   internal: /internal/run (auth, no billing)
// Webhook is the source of truth; webhooks are idempotent; AUTO_DELIVER off by
// default holds paid runs for operator review (honours the 72h/refund promise).

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"mime"
	"net"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"
)

// pageHTML is the idea.uk landing page (the taster + request/subscribe forms),
// embedded into the binary so the service is a single self-contained artefact:
// nginx just terminates TLS and proxies, and when the deployer ships the binary
// the page travels with it. Edit page.html and rebuild to change the page.
//
//go:embed page.html
var pageHTML []byte

type Config struct {
	PriceGBP        int
	AutoDeliver     bool
	ReviewBeforePay bool // run engine + operator approval BEFORE taking payment (else charge first)
	PublicBaseURL   string
	InternalAPIKey  string
	OperatorEmail   string
	ContactEmail    string // public-facing support address shown on pages
	Slots           string // header capacity phrase, e.g. "a limited number of" or "5"
	MaxActive       int
	AllowedOrigins  []string
}

type App struct {
	cfg         Config
	store       *Store
	provider    Provider
	engine      EngineFunc
	audience    func(business, audience, assets string) (audienceResult, error) // step 1 only; reused by /audience-check
	deliver     func(to, subject, body string)
	deliverHTML func(to, subject, text, htmlBody string)
	dispatch    func(func()) // how background work runs (goroutine in prod, inline in tests)
	taster      *rateLimiter // per-IP rate limiter for the public /audience-check endpoint
	landingHTML []byte       // the landing page with CONTACT_EMAIL / MONTH_SLOTS placeholders filled
}

func NewApp(cfg Config, store *Store, p Provider) *App {
	contact := cfg.ContactEmail
	if contact == "" {
		contact = cfg.OperatorEmail
	}
	slots := cfg.Slots
	if slots == "" {
		slots = "a limited number of"
	}
	rendered := strings.NewReplacer(
		"CONTACT_EMAIL", contact,
		"MONTH_SLOTS", slots,
	).Replace(string(pageHTML))
	return &App{
		cfg:         cfg,
		store:       store,
		provider:    p,
		engine:      RunMethod,
		audience:    runAudience,
		deliver:     makeDeliver(cfg.OperatorEmail),
		deliverHTML: makeDeliverHTML(cfg.OperatorEmail),
		dispatch:    func(f func()) { go f() },
		taster:      newRateLimiter(),
		landingHTML: []byte(rendered),
	}
}

func newID() string {
	return "ord_" + fmt.Sprintf("%d", time.Now().UnixNano())
}

func (a *App) operatorOK(r *http.Request) bool {
	k := r.Header.Get("X-Internal-Key")
	return a.cfg.InternalAPIKey != "" && k == a.cfg.InternalAPIKey
}

// orderToken is a per-order capability token for the click-through operator
// links. It's an HMAC of the order id under the operator secret, so it needs no
// storage and can't be guessed; it authorises actions on that ONE order only.
// (Actions stay gated by order status, so a token can't, say, confirm twice.)
func (a *App) orderToken(orderID string) string {
	mac := hmac.New(sha256.New, []byte(a.cfg.InternalAPIKey))
	mac.Write([]byte("op:" + orderID))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))[:24]
}

// opLink is the full URL the operator clicks in the request/review emails.
func (a *App) opLink(orderID string) string {
	return fmt.Sprintf("%s/op?o=%s&t=%s", a.cfg.PublicBaseURL, orderID, a.orderToken(orderID))
}

// opReq is an operator action request, parsed from either a JSON body (curl/API)
// or form/query values (the click-through page's buttons).
type opReq struct {
	OrderID string `json:"order_id"`
	Token   string `json:"token"`
	Reason  string `json:"reason"`
}

func parseOp(r *http.Request) opReq {
	var req opReq
	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		json.NewDecoder(r.Body).Decode(&req)
	} else {
		r.ParseForm()
		req = opReq{OrderID: r.FormValue("order_id"), Token: r.FormValue("t"), Reason: r.FormValue("reason")}
	}
	if req.OrderID == "" {
		req.OrderID = r.URL.Query().Get("o")
	}
	if req.Token == "" {
		req.Token = r.URL.Query().Get("t")
	}
	return req
}

// opAuthorised allows an operator action via either the internal-key header
// (curl) or a valid per-order token (the email link's button).
func (a *App) opAuthorised(r *http.Request, orderID, token string) bool {
	if a.operatorOK(r) {
		return true
	}
	return orderID != "" && token != "" && hmac.Equal([]byte(token), []byte(a.orderToken(orderID)))
}

// wantsHTML is true for a browser (so we return a friendly page); curl/API get JSON.
func wantsHTML(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "text/html")
}

// opRespond returns an HTML page to a browser, or JSON to curl/API.
func (a *App) opRespond(w http.ResponseWriter, r *http.Request, j map[string]any, htmlTitle, htmlBody string) {
	if wantsHTML(r) {
		writeHTML(w, a.page(htmlTitle, htmlBody))
		return
	}
	writeJSON(w, 200, j)
}

// ── fulfilment ───────────────────────────────────────────────────────────────
func (a *App) fulfil(id string) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("fulfil: panic for %s: %v", id, rec)
			a.store.Update(id, func(o *Order) { o.Status = "failed" })
			a.deliver(a.cfg.OperatorEmail, "[idea.uk] RUN PANIC "+id,
				fmt.Sprintf("Order %s panicked during fulfilment: %v\nReview & refund per the 72h promise.", id, rec))
		}
	}()
	o, ok := a.store.Get(id)
	if !ok || (o.Status != "paid" && o.Status != "running") {
		log.Printf("fulfil: %s not fulfillable", id)
		return
	}
	a.store.Update(id, func(o *Order) { o.Status = "running" })
	log.Printf("fulfil: %s running engine (business=%q)", id, o.Domain)
	rep, err := a.engine(o.Domain, o.Audience, o.Assets)
	if err != nil {
		log.Printf("fulfil: %s engine error: %v", id, err)
		a.store.Update(id, func(o *Order) { o.Status = "failed" })
		a.deliver(a.cfg.OperatorEmail, "[idea.uk] RUN FAILED "+id,
			fmt.Sprintf("Order %s engine error: %v\nReview & refund per the 72h promise.", id, err))
		return
	}
	log.Printf("fulfil: %s engine done (report %d chars)", id, len(rep.Text))
	// Auto-deliver straight to the buyer only in the charge-first flow, where
	// they have already paid. In review-before-pay we always hold for operator
	// approval — the buyer has not paid yet, and delivery happens after payment.
	if a.cfg.AutoDeliver && !a.cfg.ReviewBeforePay {
		a.store.Update(id, func(o *Order) { o.Status = "delivered"; o.Report = rep.Text; o.ReportHTML = rep.HTML })
		log.Printf("fulfil: %s delivering report to buyer %s", id, o.Email)
		a.deliverHTML(o.Email, "Your idea.uk report", rep.Text, rep.HTML)
	} else {
		a.store.Update(id, func(o *Order) { o.Status = "awaiting_review"; o.Report = rep.Text; o.ReportHTML = rep.HTML })
		lead := "Paid order ready for review before sending"
		if a.cfg.ReviewBeforePay {
			lead = "Draft ready for review — /approve to bill the buyer, or /decline (no charge), before anything is sent"
		}
		subject := fmt.Sprintf("[idea.uk] REVIEW %s (%s)", id, o.Domain)
		link := a.opLink(o.ID)
		text := fmt.Sprintf("%s to %s.\n\nApprove (send the buyer the pay link) or decline here:\n%s\n\n--- DRAFT REPORT ---\n\n%s", lead, o.Email, link, rep.Text)
		htmlBody := fmt.Sprintf(`<p style="font-family:-apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;color:#4A4540">%s to %s.</p><p style="font-family:-apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif"><a href="%s">Approve or decline this order</a></p>%s`,
			html.EscapeString(lead), html.EscapeString(o.Email), html.EscapeString(link), rep.HTML)
		log.Printf("fulfil: %s draft ready, emailing review to operator %s", id, a.cfg.OperatorEmail)
		a.deliverHTML(a.cfg.OperatorEmail, subject, text, htmlBody)
	}
}

// sendPayLink creates a checkout and emails the buyer the pay link, moving the
// order to awaiting_payment. Shared by confirm (charge-first) and approve
// (review-first) so the buyer-facing wording lives in exactly one place.
func (a *App) sendPayLink(o *Order) (string, error) {
	sessID, checkoutURL, err := a.provider.CreateCheckout(o.ID, o.Email)
	if err != nil {
		return "", err
	}
	a.store.Update(o.ID, func(o *Order) { o.Status = "awaiting_payment"; o.ProviderSessionID = sessID })
	a.deliver(o.Email, "Your idea.uk report — confirmed, ready to pay",
		fmt.Sprintf("Hi %s,\n\nWe can do a useful job on this. To start your report (£%d), pay here:\n\n%s\n\n"+
			"Delivered within 72 hours of payment. Full refund if we can't find anything worth acting on.",
			o.Name, a.cfg.PriceGBP, checkoutURL))
	return checkoutURL, nil
}

// deliverReport releases the already-generated report to the buyer once payment
// lands in the review-first flow. The report was produced and approved before
// payment, so this only sends it — it does not run the engine again.
func (a *App) deliverReport(id string) {
	o, ok := a.store.Get(id)
	if !ok {
		return
	}
	if o.Report == "" {
		log.Printf("deliverReport: %s paid but has no stored report", id)
		a.deliver(a.cfg.OperatorEmail, "[idea.uk] DELIVER FAILED "+id,
			fmt.Sprintf("Order %s is paid but has no stored report — investigate before the 72h promise lapses.", id))
		return
	}
	a.store.Update(id, func(o *Order) { o.Status = "delivered" })
	a.deliverHTML(o.Email, "Your idea.uk report", o.Report, o.ReportHTML)
}

// ── handlers ─────────────────────────────────────────────────────────────────
func (a *App) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"ok": true, "auto_deliver": a.cfg.AutoDeliver,
		"price_gbp": a.cfg.PriceGBP, "provider": fmt.Sprintf("%T", a.provider)})
}

func (a *App) capacity(w http.ResponseWriter, r *http.Request) {
	n := a.store.ActiveCount()
	writeJSON(w, 200, map[string]any{"open": n < a.cfg.MaxActive, "active": n, "max": a.cfg.MaxActive})
}

func (a *App) subscribe(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	email := r.FormValue("email")
	if email == "" {
		http.Error(w, "missing email", 400)
		return
	}
	a.store.AddSubscriber(email)
	writeHTML(w, a.page("You're on the list",
		`<h1>You're on the list.</h1>
<p>Thanks. The next time we run the method on a real business, we'll email you the write-up — so you can see exactly what the £29 report looks like before you ever pay for one.</p>
<p>That's all we'll send, and you can unsubscribe from any email at any time.</p>
<a class="back" href="/">← Back to idea.uk</a>`))
}

func (a *App) handleRequest(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name, email := r.FormValue("name"), r.FormValue("email")
	business, audience := r.FormValue("business"), r.FormValue("audience")
	notes := r.FormValue("notes")
	if name == "" || email == "" || business == "" || audience == "" {
		http.Error(w, "missing required field", 400)
		return
	}
	id := newID()
	now := time.Now().UTC()
	a.store.Save(&Order{ID: id, Name: name, Email: email, Domain: business,
		Audience: audience, Assets: notes, Status: "requested",
		CreatedAt: now, UpdatedAt: now})
	a.deliver(a.cfg.OperatorEmail, fmt.Sprintf("[idea.uk] New report request %s", id),
		fmt.Sprintf("New idea.uk report request.\n\nRequester: %s (%s)\nBusiness: %s\nAudience: %s\nNotes: %s\n\nOrder id: %s\n\nReview and decide (confirm to run the report, or decline) here:\n%s\n",
			name, email, business, audience, notes, id, a.opLink(id)))
	writeHTML(w, a.page("Request received",
		`<h1>Thanks, `+html.EscapeString(firstName(name))+` — we've got your request.</h1>
<div class="accent">
<p>We'll read it properly and reply by email within 24 hours — either to confirm a place and a delivery date with a link to pay, or to let you know if we're full this month and offer you the next one.</p>
<p><strong>Nothing is charged until you've seen that and said yes.</strong> Once you go ahead, the report arrives by email within a few days. If it doesn't turn up anything worth acting on, you get your money back.</p>
</div>
<a class="back" href="/">← Back to idea.uk</a>`))
}

func (a *App) confirm(w http.ResponseWriter, r *http.Request) {
	req := parseOp(r)
	if !a.opAuthorised(r, req.OrderID, req.Token) {
		http.Error(w, "unauthorised", 401)
		return
	}
	o, ok := a.store.Get(req.OrderID)
	if !ok {
		http.Error(w, "no such order", 404)
		return
	}
	if o.Status != "requested" {
		http.Error(w, "order is "+o.Status+", not 'requested'", 409)
		return
	}
	if a.store.ActiveCount() >= a.cfg.MaxActive {
		writeJSON(w, 409, map[string]any{"error": "at_capacity",
			"active": a.store.ActiveCount(), "max": a.cfg.MaxActive,
			"hint": "wait for a slot to free, or raise MaxActive"})
		return
	}
	if a.cfg.ReviewBeforePay {
		// review-first: run the engine now and hold the draft for you. Nothing
		// is charged until you /approve (or you /decline, at no charge).
		a.store.Update(o.ID, func(o *Order) { o.Status = "running" })
		a.dispatch(func() { a.fulfil(o.ID) })
		a.opRespond(w, r,
			map[string]any{"status": "running", "note": "engine started — you'll get the draft to review, then /approve to bill or /decline"},
			"Confirmed — report running",
			`<h1>Confirmed — the report is being generated.</h1>
<div class="accent"><p>You'll get the draft by email shortly, with a link to approve it (which sends the buyer the pay link) or to decline. Nothing is charged until you approve.</p></div>
<a class="back" href="/">← idea.uk</a>`)
		return
	}
	// charge-first: bill before the engine runs
	checkoutURL, err := a.sendPayLink(o)
	if err != nil {
		http.Error(w, "checkout error: "+err.Error(), 502)
		return
	}
	a.opRespond(w, r,
		map[string]any{"status": "awaiting_payment", "checkout_url": checkoutURL},
		"Pay link sent",
		`<h1>Pay link sent to the buyer.</h1><p>The buyer has been emailed a link to pay. The report runs once they've paid.</p><a class="back" href="/">← idea.uk</a>`)
}

// approve is the operator's "this draft is good, bill the buyer" step in the
// review-first flow: it sends the pay link for an order already reviewed.
func (a *App) approve(w http.ResponseWriter, r *http.Request) {
	req := parseOp(r)
	if !a.opAuthorised(r, req.OrderID, req.Token) {
		http.Error(w, "unauthorised", 401)
		return
	}
	o, ok := a.store.Get(req.OrderID)
	if !ok {
		http.Error(w, "no such order", 404)
		return
	}
	if o.Status != "awaiting_review" {
		http.Error(w, "order is "+o.Status+", not 'awaiting_review'", 409)
		return
	}
	checkoutURL, err := a.sendPayLink(o)
	if err != nil {
		http.Error(w, "checkout error: "+err.Error(), 502)
		return
	}
	a.opRespond(w, r,
		map[string]any{"status": "awaiting_payment", "checkout_url": checkoutURL},
		"Approved — pay link sent",
		`<h1>Approved.</h1><div class="accent"><p>The buyer has been emailed the pay link. Once they pay, the report is sent to them automatically.</p></div>
<a class="back" href="/">← idea.uk</a>`)
}

func (a *App) decline(w http.ResponseWriter, r *http.Request) {
	req := parseOp(r)
	if !a.opAuthorised(r, req.OrderID, req.Token) {
		http.Error(w, "unauthorised", 401)
		return
	}
	o, ok := a.store.Get(req.OrderID)
	if !ok {
		http.Error(w, "no such order", 404)
		return
	}
	reason := req.Reason
	if reason == "" {
		reason = "the differentiator is not strong enough yet"
	}
	a.store.Update(o.ID, func(o *Order) { o.Status = "declined" })
	a.deliver(o.Email, "About your idea.uk request",
		fmt.Sprintf("Hi %s,\n\nThanks for the request. Honestly, we don't think we'd produce something "+
			"worth £%d for this right now — %s. No charge, and we'd rather say so than sell you a weak report.",
			o.Name, a.cfg.PriceGBP, reason))
	a.opRespond(w, r,
		map[string]any{"status": "declined"},
		"Declined",
		`<h1>Declined.</h1><p>The requester has been emailed a polite note, at no charge.</p><a class="back" href="/">← idea.uk</a>`)
}

func (a *App) webhook(w http.ResponseWriter, r *http.Request) {
	payload := readBody(r)
	evt, err := a.provider.ParseWebhook(payload, r.Header.Get("Stripe-Signature"))
	if err != nil {
		http.Error(w, "bad webhook: "+err.Error(), 400)
		return
	}
	if a.store.MarkEventSeen(evt.EventID) {
		writeJSON(w, 200, map[string]any{"status": "duplicate_ignored"})
		return
	}
	if !evt.Paid || evt.OrderID == "" {
		writeJSON(w, 200, map[string]any{"status": "ignored", "type": evt.Type})
		return
	}
	o, ok := a.store.Get(evt.OrderID)
	if !ok {
		writeJSON(w, 200, map[string]any{"status": "unknown_order"})
		return
	}
	switch o.Status {
	case "paid", "running", "awaiting_review", "delivered":
		writeJSON(w, 200, map[string]any{"status": "already_processed"})
		return
	case "awaiting_payment":
		a.store.Update(evt.OrderID, func(o *Order) { o.Status = "paid" })
		if a.cfg.ReviewBeforePay {
			a.dispatch(func() { a.deliverReport(evt.OrderID) }) // already generated & approved
		} else {
			a.dispatch(func() { a.fulfil(evt.OrderID) }) // charge-first: generate now
		}
		writeJSON(w, 200, map[string]any{"status": "accepted"})
	default:
		writeJSON(w, 200, map[string]any{"status": "unexpected_state:" + o.Status})
	}
}

func (a *App) internalRun(w http.ResponseWriter, r *http.Request) {
	if !a.operatorOK(r) {
		http.Error(w, "unauthorised", 401)
		return
	}
	var d struct{ Domain, Audience, Assets string }
	json.NewDecoder(r.Body).Decode(&d)
	if d.Domain == "" || d.Audience == "" || d.Assets == "" {
		http.Error(w, "missing field", 400)
		return
	}
	rep, err := a.engine(d.Domain, d.Audience, d.Assets)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, 200, map[string]any{"report": rep.Text, "report_html": rep.HTML})
}

func (a *App) orderSuccess(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("fake") != "" { // FakeProvider local-test shortcut
		id := r.URL.Query().Get("o")
		if o, ok := a.store.Get(id); ok && o.Status == "awaiting_payment" {
			a.store.Update(id, func(o *Order) { o.Status = "paid" })
			if a.cfg.ReviewBeforePay {
				a.dispatch(func() { a.deliverReport(id) })
			} else {
				a.dispatch(func() { a.fulfil(id) })
			}
		}
	}
	writeHTML(w, a.page("Payment received",
		`<h1>Payment received — thank you.</h1>
<div class="accent">
<p>We're putting your report together now. It'll arrive by email within 72 hours, and usually sooner. A person reads it before it goes out.</p>
<p>If, after reading it, you don't think it turns up anything worth acting on, just email us within 14 days and we'll refund you in full.</p>
</div>
<a class="back" href="/">← Back to idea.uk</a>`))
}

func (a *App) orderCancel(w http.ResponseWriter, r *http.Request) {
	writeHTML(w, a.page("Nothing was charged",
		`<h1>No problem — nothing was charged.</h1>
<p>Your card hasn't been billed. If you cancelled by mistake, you can start again whenever you're ready.</p>
<a class="back" href="/">← Back to idea.uk</a>`))
}

func (a *App) termsPage(w http.ResponseWriter, r *http.Request) {
	writeHTML(w, a.page("Terms", strings.ReplaceAll(termsBody, "{{EMAIL}}", a.contactEmail())))
}

func (a *App) refundPage(w http.ResponseWriter, r *http.Request) {
	writeHTML(w, a.page("Refund policy", strings.ReplaceAll(refundBody, "{{EMAIL}}", a.contactEmail())))
}

func (a *App) privacyPage(w http.ResponseWriter, r *http.Request) {
	writeHTML(w, a.page("Privacy", strings.ReplaceAll(privacyBody, "{{EMAIL}}", a.contactEmail())))
}

// ── routing + CORS ───────────────────────────────────────────────────────────
func (a *App) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", a.health)
	mux.HandleFunc("/capacity", a.capacity)
	mux.HandleFunc("/audience-check", a.audienceCheck) // free public taster — Step 1 only
	mux.HandleFunc("/subscribe", a.subscribe)
	mux.HandleFunc("/request", a.handleRequest)
	mux.HandleFunc("/confirm", a.confirm)
	mux.HandleFunc("/approve", a.approve)
	mux.HandleFunc("/decline", a.decline)
	mux.HandleFunc("/op", a.opPage) // click-through operator page (token in link)
	mux.HandleFunc("/stripe/webhook", a.webhook)
	mux.HandleFunc("/internal/run", a.internalRun)
	mux.HandleFunc("/order/success", a.orderSuccess)
	mux.HandleFunc("/order/cancel", a.orderCancel)
	mux.HandleFunc("/terms", a.termsPage)
	mux.HandleFunc("/refund-policy", a.refundPage)
	mux.HandleFunc("/privacy", a.privacyPage)
	mux.HandleFunc("/", a.home) // landing page at "/"; 404 for anything else unmatched
	return a.cors(mux)
}

// home serves the embedded landing page at "/" only. ServeMux routes every
// otherwise-unmatched path here, so anything that isn't exactly "/" gets a 404
// rather than the page.
func (a *App) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(a.landingHTML)
}

func (a *App) cors(next http.Handler) http.Handler {
	allowed := map[string]bool{}
	for _, o := range a.cfg.AllowedOrigins {
		allowed[o] = true
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && (allowed["*"] || allowed[origin]) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST")
			w.Header().Set("Access-Control-Allow-Headers", "*")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ── small helpers ────────────────────────────────────────────────────────────
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

// opPage is the click-through landing page from the operator emails. It is a
// side-effect-free GET (so a mail scanner pre-fetching the link can't trigger an
// action); the action only happens when the operator clicks a button, which POSTs
// the per-order token. Which buttons appear depends on the order's status.
func (a *App) opPage(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("o")
	token := r.URL.Query().Get("t")
	o, ok := a.store.Get(id)
	if !ok || token == "" || !hmac.Equal([]byte(token), []byte(a.orderToken(id))) {
		w.WriteHeader(404)
		writeHTML(w, a.page("Link not valid",
			`<h1>This link isn't valid.</h1><p>The order may not exist, or the link is incomplete. Check the most recent request email.</p><a class="back" href="/">← idea.uk</a>`))
		return
	}
	esc := html.EscapeString
	var b strings.Builder
	fmt.Fprintf(&b, `<h1>Report request — %s</h1>`, esc(o.Domain))
	fmt.Fprintf(&b, `<p class="meta">Order %s &middot; status: <strong>%s</strong></p>`, esc(o.ID), esc(o.Status))
	fmt.Fprintf(&b, `<p><strong>Requester:</strong> %s (%s)</p>`, esc(o.Name), esc(o.Email))
	fmt.Fprintf(&b, `<p><strong>Audience:</strong> %s</p>`, esc(o.Audience))
	if o.Assets != "" {
		fmt.Fprintf(&b, `<p><strong>Notes:</strong> %s</p>`, esc(o.Assets))
	}
	btn := func(action, label, color string) string {
		return `<form method="POST" action="` + action + `" style="display:inline-block;margin:0 10px 10px 0">` +
			`<input type="hidden" name="order_id" value="` + esc(o.ID) + `">` +
			`<input type="hidden" name="t" value="` + esc(token) + `">` +
			`<button type="submit" style="background:` + color + `;color:#fff;border:0;border-radius:6px;padding:12px 18px;font-size:15px;cursor:pointer">` + esc(label) + `</button></form>`
	}
	switch o.Status {
	case "requested":
		b.WriteString(`<div class="accent"><p>Confirming runs the report now (it spends on the engine). You'll then get the draft to review before anything is billed.</p></div>`)
		b.WriteString(`<div style="margin-top:18px">` + btn("/confirm", "Confirm and run the report", "#15243d") + btn("/decline", "Decline (no charge)", "#7d2a12") + `</div>`)
	case "awaiting_review":
		b.WriteString(`<div class="accent"><p>The draft is ready (in your email). Approving sends the buyer the pay link; the report is delivered once they pay.</p></div>`)
		b.WriteString(`<div style="margin-top:18px">` + btn("/approve", "Approve and send the pay link", "#15243d") + btn("/decline", "Decline (no charge)", "#7d2a12") + `</div>`)
	case "running":
		b.WriteString(`<p>The report is being generated now. You'll get the draft by email shortly — reopen this link then to approve or decline.</p>`)
	case "awaiting_payment":
		b.WriteString(`<p>The pay link has been sent to the buyer. Waiting for payment.</p>`)
	case "delivered":
		b.WriteString(`<p>This report has been delivered to the buyer.</p>`)
	case "declined":
		b.WriteString(`<p>This order was declined.</p>`)
	default:
		fmt.Fprintf(&b, `<p>No action available for status "%s".</p>`, esc(o.Status))
	}
	b.WriteString(`<a class="back" href="/">← idea.uk</a>`)
	writeHTML(w, a.page("Order "+o.ID, b.String()))
}

func writeHTML(w http.ResponseWriter, s string) {
	w.Header().Set("content-type", "text/html; charset=utf-8")
	w.Write([]byte(s))
}

// firstName returns the first word of a name for a friendly greeting.
func firstName(full string) string {
	if f := strings.Fields(full); len(f) > 0 {
		return f[0]
	}
	return "there"
}

// contactEmail returns the public support address (ContactEmail, or OperatorEmail
// if unset).
func (a *App) contactEmail() string {
	if a.cfg.ContactEmail != "" {
		return a.cfg.ContactEmail
	}
	return a.cfg.OperatorEmail
}

// page wraps body HTML in a full, brand-styled document so the post-submit, order,
// and policy pages look like idea.uk instead of bare unstyled text. Body is trusted
// (we build it ourselves); any user input interpolated into it must be escaped
// by the caller.
func (a *App) page(title, body string) string {
	contact := a.contactEmail()
	return `<!doctype html><html lang="en"><head>
<meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>` + html.EscapeString(title) + ` · idea.uk</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=Fraunces:opsz,wght@9..144,400;9..144,600&family=IBM+Plex+Sans:wght@400;500&display=swap" rel="stylesheet">
<style>
:root{--ink:#1A1816;--ink-soft:#4A4540;--ink-mute:#837C72;--paper:#EFE7D6;--paper-2:#E8DFCC;--rust:#A8391A;--rust-deep:#7D2A12;--serif:"Fraunces","Times New Roman",Georgia,serif;--sans:"IBM Plex Sans",-apple-system,BlinkMacSystemFont,sans-serif}
*{box-sizing:border-box}
html,body{margin:0;background:var(--paper);color:var(--ink);font-family:var(--sans)}
.bar{border-bottom:1px solid var(--ink);padding:20px 24px}
.wordmark{font-family:var(--serif);font-weight:600;font-size:22px}
.wordmark .dot{color:var(--rust)}
main{max-width:680px;margin:0 auto;padding:64px 24px;padding-left:max(24px,env(safe-area-inset-left));padding-right:max(24px,env(safe-area-inset-right))}
h1{font-family:var(--serif);font-weight:600;font-size:32px;line-height:1.15;margin:0 0 18px}
h2{font-family:var(--serif);font-weight:600;font-size:21px;line-height:1.25;margin:34px 0 12px;color:var(--ink)}
p{font-size:17px;line-height:1.62;color:var(--ink-soft);margin:0 0 16px}
ul,ol{margin:0 0 16px;padding-left:22px}
li{font-size:17px;line-height:1.6;color:var(--ink-soft);margin:0 0 8px}
strong{color:var(--ink)}
.meta{font-size:14px;color:var(--ink-mute);margin:0 0 26px}
.note{background:var(--paper-2);border-left:3px solid var(--rust);padding:16px 20px;margin:22px 0;font-size:15px;color:var(--ink-soft)}
a{color:var(--ink);text-decoration-color:var(--rust)}
.accent{border-left:3px solid var(--rust);padding-left:20px;margin:24px 0}
.back{display:inline-block;margin-top:20px;font-size:15px}
.foot{margin-top:44px;padding-top:20px;border-top:1px solid var(--ink-mute);font-size:14px;color:var(--ink-mute)}
</style></head><body>
<div class="bar"><span class="wordmark">idea<span class="dot">.</span>uk</span></div>
<main>` + body + `
<p class="foot">Questions? Email <a href="mailto:` + contact + `">` + contact + `</a> &nbsp;·&nbsp; <a href="/">idea.uk</a> &nbsp;·&nbsp; <a href="/terms">Terms</a> &nbsp;·&nbsp; <a href="/refund-policy">Refunds</a> &nbsp;·&nbsp; <a href="/privacy">Privacy</a> &nbsp;·&nbsp; by <a href="https://leopardess.uk">leopardess.uk</a></p>
</main></body></html>`
}

func readBody(r *http.Request) []byte {
	b, _ := io.ReadAll(r.Body)
	return b
}

// sendOne builds and sends a single email. If htmlBody is non-empty it sends a
// multipart/alternative message (plain-text fallback + HTML); otherwise plain
// text only. Synchronous — callers wrap it in a goroutine so SMTP I/O never
// blocks the HTTP request path; failures are logged, not returned (the
// webhook/store is the source of truth for fulfilment).
func sendOne(operatorEmail, to, subject, text, htmlBody string) {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		fn := fmt.Sprintf("delivered_%s_%d.md",
			strings.NewReplacer("@", "_at_", "/", "_").Replace(to), time.Now().Unix())
		_ = os.WriteFile(fn, []byte(text), 0o600)
		log.Printf("SMTP not configured — wrote %s", fn)
		return
	}
	from := env("SMTP_FROM", operatorEmail) // bare envelope sender
	fromHeader := from
	if name := os.Getenv("SMTP_FROM_NAME"); name != "" {
		fromHeader = fmt.Sprintf("%s <%s>", mime.QEncoding.Encode("utf-8", name), from)
	}
	var msg strings.Builder
	fmt.Fprintf(&msg, "From: %s\r\nTo: %s\r\nSubject: %s\r\n", fromHeader, to, mime.QEncoding.Encode("utf-8", subject))
	if rt := os.Getenv("SMTP_REPLY_TO"); rt != "" {
		fmt.Fprintf(&msg, "Reply-To: %s\r\n", rt)
	}
	msg.WriteString("MIME-Version: 1.0\r\n")
	if htmlBody == "" {
		// Plain text. UTF-8 declared so £, —, ≥, smart quotes render correctly
		// instead of as mojibake (e.g. â‰¥) when a client assumes Latin-1.
		msg.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
		msg.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
		msg.WriteString(text)
	} else {
		// multipart/alternative: plain-text fallback first, HTML second — clients
		// pick the richest part they can render.
		boundary := fmt.Sprintf("idea-bnd-%d", time.Now().UnixNano())
		fmt.Fprintf(&msg, "Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", boundary)
		fmt.Fprintf(&msg, "--%s\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\nContent-Transfer-Encoding: 8bit\r\n\r\n%s\r\n", boundary, text)
		fmt.Fprintf(&msg, "--%s\r\nContent-Type: text/html; charset=\"UTF-8\"\r\nContent-Transfer-Encoding: 8bit\r\n\r\n%s\r\n", boundary, htmlBody)
		fmt.Fprintf(&msg, "--%s--\r\n", boundary)
	}
	port := env("SMTP_PORT", "587")
	if err := smtpSend(host, port, os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASS"), from, []string{to}, []byte(msg.String())); err != nil {
		log.Printf("email to %s failed: %v", to, err)
		return
	}
	log.Printf("email to %s sent: %q", to, subject)
}

func makeDeliver(operatorEmail string) func(to, subject, body string) {
	return func(to, subject, body string) { go sendOne(operatorEmail, to, subject, body, "") }
}

func makeDeliverHTML(operatorEmail string) func(to, subject, text, htmlBody string) {
	return func(to, subject, text, htmlBody string) { go sendOne(operatorEmail, to, subject, text, htmlBody) }
}

// smtpSend delivers one message. Port 465 uses implicit TLS (the connection is
// encrypted from the start — what cPanel/Clook offers); any other port (587, 25)
// uses smtp.SendMail, which negotiates STARTTLS. Auth is sent only over TLS.
func smtpSend(host, port, user, pass, from string, to []string, msg []byte) error {
	addr := host + ":" + port
	var auth smtp.Auth
	if user != "" {
		auth = smtp.PlainAuth("", user, pass, host)
	}
	if port != "465" {
		return smtp.SendMail(addr, auth, from, to, msg) // STARTTLS path
	}
	// implicit TLS (465); bound the connect and the whole conversation so a
	// network problem fails in ~10s instead of hanging on the OS TCP timeout.
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 10 * time.Second}, "tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return err
	}
	_ = conn.SetDeadline(time.Now().Add(30 * time.Second))
	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer c.Close()
	if auth != nil {
		if err := c.Auth(auth); err != nil {
			return err
		}
	}
	if err := c.Mail(from); err != nil {
		return err
	}
	for _, rcpt := range to {
		if err := c.Rcpt(rcpt); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return c.Quit()
}

// ── policy page content ──────────────────────────────────────────────────────
// Plain-language DRAFTS. Have a UK solicitor review before taking real payments.
// {{EMAIL}} is filled in at serve time; [bracketed] items are for you to complete.

const termsBody = `<h1>Terms</h1>
<p class="meta">Last updated: 5 June 2026</p>
<p>These terms cover your use of idea.uk and the reports we provide. By requesting a report, you agree to them. Please read them alongside our <a href="/refund-policy">refund policy</a>.</p>

<h2>Who we are</h2>
<p>idea.uk is operated under the name idea.uk, in the United Kingdom. If you have any questions, email us at <a href="mailto:{{EMAIL}}">{{EMAIL}}</a>.</p>

<h2>What you get</h2>
<p>A written report about possible AI products for your business. We use AI tools to come up with ideas, test them, and check the facts behind them, and a person reviews the report before it is sent to you. You receive it by email, normally within five working days of payment and usually sooner.</p>

<h2>What this is not</h2>
<p>This is an information and research service. It is not professional advice. We are not your accountant, solicitor, financial adviser, or business consultant, and the report should not be treated as advice of that kind. The decisions you make about your business are yours. Every factual claim in the report comes with a source you can check, and you should check those sources before you spend money or commit to anything based on the report.</p>

<h2>Ordering and payment</h2>
<p>You send us a request first, with no payment. We reply within 24 hours, either to confirm a place and a delivery date with a link to pay, or to let you know we are full for the time being. We only take payment after you have accepted a confirmed place. The price is shown when you order (currently £29 per report).</p>

<h2>Your part</h2>
<p>The report is only as good as what you tell us, so please describe your business and the people you want to reach accurately. You agree not to use the service for anything unlawful, and not to present the report as your own professional advice to others.</p>

<h2>The report is produced using AI, and AI can be wrong</h2>
<p>We use artificial intelligence to research, generate, and check the ideas in your report. AI is useful but imperfect, and it can be confidently wrong. It can invent facts, figures, companies, prices, quotes, or sources that look convincing but are not real — this is often called "hallucination" — and it can misread information or rely on something out of date. A person reviews each report before it goes out, but that review cannot catch everything, and we do not claim that it does.</p>
<p>So please treat everything in the report as something to check, not as established fact. Markets, prices, and rules also change over time. Check the sources we cite, confirm any figure, competitor, price, or rule for yourself, and take your own professional advice before you spend money or commit to anything. We do not promise that the report is accurate, complete, current, or suitable for any particular purpose.</p>
<p>What you do with the report, and any decision you make based on it, is entirely your responsibility and not ours.</p>

<h2>Our responsibility to you</h2>
<p>We provide the report with reasonable care and skill. As far as the law allows, we are not responsible for business losses, lost profits, or other losses that follow from decisions you make based on the report, and our total responsibility to you for anything connected with a report will not be more than the amount you paid for it. Nothing in these terms limits anything that cannot be limited by law — for example, liability for fraud, or for death or personal injury caused by our negligence. None of this affects your rights as a consumer.</p>

<h2>Using the report</h2>
<p>Once you have paid, you may use the report freely within your own business. We keep the rights in the methods, templates, and formats we use to produce it.</p>

<h2>The information you give us</h2>
<p>We use the details you submit only to produce and deliver your report and to contact you about it, and we do not sell your information. For the full picture — what we collect, who processes it, and your rights — see our <a href="/privacy">privacy policy</a>.</p>

<h2>Refunds</h2>
<p>If the report does not turn up anything worth acting on, you can ask for a full refund within 14 days. The full details are in our <a href="/refund-policy">refund policy</a>.</p>

<h2>Changes to these terms</h2>
<p>We may update these terms from time to time. The version in force when you place your order is the one that applies to that order.</p>

<h2>Which law applies</h2>
<p>These terms are governed by the law of England and Wales, and the courts of England and Wales will deal with any dispute. [Confirm the right jurisdiction with your adviser.]</p>

<h2>Contact</h2>
<p>Questions about these terms? Email <a href="mailto:{{EMAIL}}">{{EMAIL}}</a>.</p>
<a class="back" href="/">← Back to idea.uk</a>`

const refundBody = `<h1>Refund policy</h1>
<p class="meta">Last updated: 5 June 2026</p>
<p>The short version: you do not pay until we confirm we can do a useful job, and if the report does not turn up anything worth acting on, you get your money back.</p>

<h2>You pay only after we confirm</h2>
<p>You send a request first, with no payment. We only send a payment link once we have confirmed a place and a delivery date. If we are full, or we do not think we can do a useful job for you, we say so and you pay nothing.</p>

<h2>If the report is not useful</h2>
<p>When you have read your report, if you do not think it turns up anything worth acting on, email us within 14 days of receiving it and we will refund you in full. You do not need to give a detailed reason. A line about what was missing is always welcome, so we can do better, but it is not a condition of the refund.</p>

<h2>If something goes wrong at our end</h2>
<p>If we fail to deliver your report, or it arrives plainly incomplete or unusable, you are entitled to a full refund regardless of the 14-day period.</p>

<h2>How to ask for a refund</h2>
<p>Email <a href="mailto:{{EMAIL}}">{{EMAIL}}</a> from the address you used to order, or quote your order reference. We aim to reply within two working days and to return the money to your original payment method within five to ten working days.</p>

<h2>Your legal rights</h2>
<p>This policy is offered on top of your rights under UK consumer law; it does not replace or reduce them, and nothing here affects your statutory rights. [Ask your adviser to confirm how the Consumer Contracts Regulations 2013 apply to a bespoke report that you have asked us to begin.]</p>

<h2>Contact</h2>
<p>Questions about a refund? Email <a href="mailto:{{EMAIL}}">{{EMAIL}}</a>.</p>
<a class="back" href="/">← Back to idea.uk</a>`

const privacyBody = `<h1>Privacy policy</h1>
<p class="meta">Last updated: 5 June 2026</p>
<p>This explains what personal information idea.uk collects, why, who else handles it, and the rights you have. It is written to be read, not to be clever. If anything is unclear, email us at <a href="mailto:{{EMAIL}}">{{EMAIL}}</a>.</p>

<h2>Who is responsible for your information</h2>
<p>idea.uk, operating under the name idea.uk in the United Kingdom, is responsible for the personal information described here (the "data controller"). You can reach us at <a href="mailto:{{EMAIL}}">{{EMAIL}}</a>.</p>

<h2>What we collect</h2>
<ul>
<li><strong>When you request a report:</strong> your name, your email address, and the business and audience details you choose to tell us, including anything you put in the notes field.</li>
<li><strong>When you ask for updates:</strong> your email address.</li>
<li><strong>When you pay:</strong> your payment is handled by Stripe. We receive confirmation that you have paid and limited details such as your name and the last few digits of your card — we never see or store your full card number.</li>
<li><strong>When you use the free audience check:</strong> we process the business and audience you type in to produce the result. We do not store it unless you ask us to.</li>
<li><strong>Automatically:</strong> our server keeps standard technical logs, such as your IP address and the time of a request, to keep the service running and secure.</li>
</ul>

<h2>Why we use it, and our legal basis</h2>
<ul>
<li>To prepare and deliver your report and to reply to you — because it is needed to provide the service you have asked for (performance of a contract, and steps taken before one).</li>
<li>To keep the service running, secure, and free from abuse — because we have a legitimate interest in doing so.</li>
<li>To send you the updates you signed up for — on the basis of your consent, which you can withdraw at any time.</li>
<li>To meet our legal duties, such as keeping basic records of payments.</li>
</ul>

<h2>Who else handles your information</h2>
<p>We do not sell your information, and we do not share it for advertising. We use a small number of trusted suppliers to run the service, and they only process your information on our instructions:</p>
<ul>
<li><strong>Stripe</strong> — to take and process payments.</li>
<li><strong>Anthropic</strong> — our AI provider. The business and audience details you give us are sent to Anthropic so the report can be researched, generated, and checked. We do not send your name or email to the AI for this.</li>
<li><strong>Hetzner</strong> — our hosting provider; the servers that run the service are in Germany.</li>
<li><strong>Clook</strong> — our email provider (UK); it sends our email and receives the email you send us.</li>
<li><strong>Google (Gmail)</strong> — the email you send us is forwarded on to a Gmail inbox that we read.</li>
</ul>

<h2>Information that leaves the UK</h2>
<p>We keep your information in the UK or the EU where we can: our servers are in Germany, and our email is handled in the UK by Clook. Information goes outside the UK in two places: the business and audience details are sent to Anthropic, in the United States, so your report can be produced; and any email you send us is forwarded on to a Gmail inbox, so it is handled by Google in the United States. Where information goes outside the UK, we rely on the protections allowed under UK data protection law, such as an adequacy decision or the UK's standard contractual terms for international transfers. [Ask your adviser to confirm the exact safeguards for each supplier.]</p>

<h2>How long we keep it</h2>
<ul>
<li>Your request, your report, and related correspondence: for as long as we need it to provide the service and to meet our legal and accounting duties, after which we delete it. [Set your retention period here — financial records in the UK are commonly kept for six years.]</li>
<li>Update subscribers: until you unsubscribe.</li>
<li>The free audience check: not stored.</li>
</ul>

<h2>Your rights</h2>
<p>Under UK data protection law you can ask us to: give you a copy of the information we hold about you; correct anything that is wrong; delete it; restrict or object to how we use it; or provide it in a portable form. Where we rely on your consent, you can withdraw it at any time. To exercise any of these, email <a href="mailto:{{EMAIL}}">{{EMAIL}}</a> and we will respond within the time the law allows.</p>
<p>If you are unhappy with how we have handled your information, you can complain to the Information Commissioner's Office (ICO) at <a href="https://ico.org.uk">ico.org.uk</a>. We would appreciate the chance to put things right first.</p>

<h2>Cookies and tracking</h2>
<p>The idea.uk website does not use cookies, analytics, or third-party advertising or tracking.</p>

<h2>Keeping your information safe</h2>
<p>We take reasonable steps to protect the information we hold, but no service can promise that information sent over the internet is completely secure. Please do not send us anything more sensitive than the service needs.</p>

<h2>Children</h2>
<p>idea.uk is a service for businesses and is not intended for anyone under 18.</p>

<h2>Changes to this policy</h2>
<p>We may update this policy from time to time. The version published here is the one that applies.</p>

<h2>Contact</h2>
<p>Questions about your information, or want to exercise a right? Email <a href="mailto:{{EMAIL}}">{{EMAIL}}</a>.</p>
<a class="back" href="/">← Back to idea.uk</a>`
