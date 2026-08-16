package gripper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gqls/agentchassis/platform/mailer"
)

// Request statuses. The lifecycle is
//
//	pending → pulled → fulfilled → emailed
//	                 ↘ failed  ─┐
//	pending/pulled → expired  ─┴→ (apology sent: failure_notified_at set)
//	fulfilled → email_failed (link email refused MaxEmailAttempts times)
//
// Every transition is a guarded UPDATE (WHERE status = expected) in the store,
// so a poller tick that overlaps a /requests write, or two ticks that overlap
// each other, cannot move a row twice.
const (
	StatusPending     = "pending"
	StatusPulled      = "pulled"
	StatusFulfilled   = "fulfilled"
	StatusEmailed     = "emailed"
	StatusEmailFailed = "email_failed"
	StatusFailed      = "failed"
	StatusExpired     = "expired"
)

// Timing constants from DESIGN §5.2 and §2.
const (
	FirstCheckAfter  = 2 * time.Minute
	EarlyCheckEvery  = 5 * time.Minute
	LateCheckEvery   = 15 * time.Minute
	EarlyWindow      = time.Hour
	RequestTTL       = 24 * time.Hour
	SidecarTimeout   = 15 * time.Second
	DefaultTick      = 60 * time.Second
	MaxEmailAttempts = 3
	EmailRetryAfter  = 10 * time.Minute
	SessionIdleTTL   = 24 * time.Hour      // transcript GC (retention Q7)
	PIIRetention     = 90 * 24 * time.Hour // email + ip null after terminal (Q7)
	batchLimit       = 50
)

// NextCheck returns when a not-yet-ready request should be polled again:
// every 5 minutes for the first hour after it was created, then every 15.
func NextCheck(createdAt, now time.Time) time.Time {
	if now.Sub(createdAt) < EarlyWindow {
		return now.Add(EarlyCheckEvery)
	}
	return now.Add(LateCheckEvery)
}

// Request is the poller's view of a gripper_report_requests row.
type Request struct {
	ID            string
	SiteDomain    string
	Email         string
	Status        string
	CreatedAt     time.Time
	ExpiresAt     time.Time
	EmailAttempts int
	ReportURL     string
}

// RequestStore is what the poller needs from persistence. *store.Gripper
// implements it against Postgres; tests use an in-memory fake.
type RequestStore interface {
	// DueChecks: status IN (pending, pulled) AND next_check_at <= now.
	DueChecks(ctx context.Context, now time.Time, limit int) ([]Request, error)
	MarkFulfilled(ctx context.Context, id, reportURL string, now time.Time) (bool, error)
	MarkFailed(ctx context.Context, id string, now time.Time) (bool, error)
	MarkExpired(ctx context.Context, id string, now time.Time) (bool, error)
	RescheduleCheck(ctx context.Context, id string, next time.Time) error

	// DueLinkEmails: status = fulfilled AND email_attempts < max AND next_check_at <= now.
	DueLinkEmails(ctx context.Context, now time.Time, maxAttempts, limit int) ([]Request, error)
	// DueApologies: status IN (failed, expired) AND failure_notified_at IS NULL
	// AND email_attempts < max AND next_check_at <= now.
	DueApologies(ctx context.Context, now time.Time, maxAttempts, limit int) ([]Request, error)
	// ClaimEmailAttempt increments email_attempts and pushes next_check_at to
	// retryAt, guarded on the row still being in one of expectStatus. It is
	// committed BEFORE the send, so a crash mid-send costs at most one retry,
	// never an unbounded stream of duplicates.
	ClaimEmailAttempt(ctx context.Context, id string, expectStatus []string, retryAt time.Time) (bool, error)
	MarkEmailed(ctx context.Context, id string, now time.Time) (bool, error)
	MarkEmailFailed(ctx context.Context, id string) (bool, error)
	MarkFailureNotified(ctx context.Context, id string, now time.Time) (bool, error)

	// Retention.
	ExpireIdleSessions(ctx context.Context, idleBefore time.Time) (int64, error)
	ScrubTerminalPII(ctx context.Context, terminalBefore time.Time) (int64, error)
}

// Poller drives report requests from pulled to emailed. One instance per
// process; it runs on a ticker and skips a tick that would overlap the last.
type Poller struct {
	Store  RequestStore
	Sender mailer.Sender
	HTTP   *http.Client
	Log    *log.Logger
	Now    func() time.Time
	Tick   time.Duration
	// Hourly is an optional maintenance hook (e.g. limiter sweeps) run once
	// per hour from the same goroutine.
	Hourly func()

	running   atomic.Bool
	lastMaint time.Time
}

func (p *Poller) now() time.Time {
	if p.Now != nil {
		return p.Now()
	}
	return time.Now()
}

func (p *Poller) logf(format string, args ...interface{}) {
	if p.Log != nil {
		p.Log.Printf("gripper/poller: "+format, args...)
	}
}

// Run ticks until ctx is cancelled. Safe to call once.
func (p *Poller) Run(ctx context.Context) {
	tick := p.Tick
	if tick <= 0 {
		tick = DefaultTick
	}
	if p.HTTP == nil {
		p.HTTP = &http.Client{Timeout: SidecarTimeout}
	}
	t := time.NewTicker(tick)
	defer t.Stop()
	p.logf("started tick=%s", tick)
	for {
		select {
		case <-ctx.Done():
			p.logf("stopped: %v", ctx.Err())
			return
		case <-t.C:
			p.RunOnce(ctx)
		}
	}
}

// RunOnce performs one tick: the three lanes, then hourly maintenance. It is
// exported so tests (and an operator) can drive it without a ticker. A tick
// that finds the previous one still running returns immediately.
func (p *Poller) RunOnce(ctx context.Context) {
	if !p.running.CompareAndSwap(false, true) {
		p.logf("tick skipped: previous tick still running")
		return
	}
	defer p.running.Store(false)

	now := p.now()
	p.checkLane(ctx, now)
	p.linkLane(ctx, now)
	p.apologyLane(ctx, now)

	if now.Sub(p.lastMaint) >= time.Hour {
		p.lastMaint = now
		p.maintenance(ctx, now)
	}
}

// checkLane polls the sidecar for every request that is due.
func (p *Poller) checkLane(ctx context.Context, now time.Time) {
	due, err := p.Store.DueChecks(ctx, now, batchLimit)
	if err != nil {
		p.logf("check lane: DueChecks FAILED err=%v", err)
		return
	}
	for _, r := range due {
		if ctx.Err() != nil {
			return
		}
		st, sidecarURL, err := p.fetchSidecar(ctx, r, now)
		switch {
		case err == nil && st == "ready":
			link := p.reportLink(r, sidecarURL)
			ok, uerr := p.Store.MarkFulfilled(ctx, r.ID, link, now)
			p.logf("request=%s sidecar=ready fulfilled=%v err=%v", r.ID, ok, uerr)
		case err == nil && st == "failed":
			ok, uerr := p.Store.MarkFailed(ctx, r.ID, now)
			p.logf("request=%s sidecar=failed marked=%v err=%v", r.ID, ok, uerr)
		default:
			// Not there yet, or an unreadable answer: both mean "ask again later",
			// unless the request has run out of time.
			if !now.Before(r.ExpiresAt) {
				ok, uerr := p.Store.MarkExpired(ctx, r.ID, now)
				p.logf("request=%s EXPIRED after %s marked=%v err=%v (last sidecar: status=%q err=%v)",
					r.ID, RequestTTL, ok, uerr, st, err)
				continue
			}
			next := NextCheck(r.CreatedAt, now)
			if rerr := p.Store.RescheduleCheck(ctx, r.ID, next); rerr != nil {
				p.logf("request=%s reschedule FAILED err=%v", r.ID, rerr)
			}
			if err != nil {
				p.logf("request=%s sidecar not ready: %v (next %s)", r.ID, err, next.Format(time.RFC3339))
			}
		}
	}
}

// fetchSidecar GETs https://<domain>/reports/<id>.json with a cache-buster and
// returns its status ("ready" | "failed" | other) and its url field. A 404 is
// the normal "not yet" and is returned as an error so the caller reschedules.
func (p *Poller) fetchSidecar(ctx context.Context, r Request, now time.Time) (status, sidecarURL string, err error) {
	u := SidecarURL(r.SiteDomain, r.ID) + "?cb=" + strconv.FormatInt(now.UnixNano(), 10)
	cctx, cancel := context.WithTimeout(ctx, SidecarTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(cctx, http.MethodGet, u, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	resp, err := p.HTTP.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("sidecar %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return "", "", err
	}
	var sc struct {
		Status string `json:"status"`
		URL    string `json:"url"`
	}
	if err := json.Unmarshal(body, &sc); err != nil {
		return "", "", fmt.Errorf("sidecar unparseable: %w", err)
	}
	return sc.Status, sc.URL, nil
}

// SidecarURL is the one URL the island polls for a request (DESIGN §5.2).
func SidecarURL(domain, id string) string {
	return "https://" + domain + "/reports/" + id + ".json"
}

// DefaultReportLink is where the cluster's create_report_page puts the page.
func DefaultReportLink(domain, id string) string {
	return "https://" + domain + "/reports/" + id + ".html"
}

// reportLink resolves the link to email. The sidecar's url is site-relative
// ("/reports/<id>.html") and is honoured when it resolves to the same host —
// it is the artefact's own statement of where it lives — otherwise the
// conventional path is used and the disagreement is logged.
func (p *Poller) reportLink(r Request, sidecarURL string) string {
	fallback := DefaultReportLink(r.SiteDomain, r.ID)
	if strings.TrimSpace(sidecarURL) == "" {
		return fallback
	}
	base, _ := url.Parse("https://" + r.SiteDomain + "/")
	ref, err := url.Parse(sidecarURL)
	if err != nil {
		p.logf("request=%s sidecar url unparseable, using default", r.ID)
		return fallback
	}
	abs := base.ResolveReference(ref)
	if abs.Scheme != "https" || !strings.EqualFold(abs.Host, r.SiteDomain) {
		p.logf("request=%s sidecar url is off-host (%s), using default", r.ID, abs.Host)
		return fallback
	}
	if abs.String() != fallback {
		p.logf("request=%s sidecar url differs from convention: using %s", r.ID, abs.String())
	}
	return abs.String()
}

// linkLane sends the "ready" email for fulfilled requests.
func (p *Poller) linkLane(ctx context.Context, now time.Time) {
	due, err := p.Store.DueLinkEmails(ctx, now, MaxEmailAttempts, batchLimit)
	if err != nil {
		p.logf("link lane: DueLinkEmails FAILED err=%v", err)
		return
	}
	for _, r := range due {
		if ctx.Err() != nil {
			return
		}
		claimed, err := p.Store.ClaimEmailAttempt(ctx, r.ID, []string{StatusFulfilled}, now.Add(EmailRetryAfter))
		if err != nil || !claimed {
			p.logf("request=%s link claim=%v err=%v", r.ID, claimed, err)
			continue
		}
		serr := p.Sender.Send(ctx, LinkMessage(r.Email, r.ReportURL))
		if serr == nil {
			ok, uerr := p.Store.MarkEmailed(ctx, r.ID, now)
			p.logf("request=%s link SENT emailed=%v err=%v", r.ID, ok, uerr)
			continue
		}
		attempt := r.EmailAttempts + 1
		p.logf("request=%s link send FAILED attempt=%d/%d err=%v", r.ID, attempt, MaxEmailAttempts, serr)
		if attempt >= MaxEmailAttempts {
			ok, uerr := p.Store.MarkEmailFailed(ctx, r.ID)
			p.logf("request=%s link GIVEN UP email_failed=%v err=%v", r.ID, ok, uerr)
		}
	}
}

// apologyLane sends the apology for failed and expired requests, once.
func (p *Poller) apologyLane(ctx context.Context, now time.Time) {
	due, err := p.Store.DueApologies(ctx, now, MaxEmailAttempts, batchLimit)
	if err != nil {
		p.logf("apology lane: DueApologies FAILED err=%v", err)
		return
	}
	for _, r := range due {
		if ctx.Err() != nil {
			return
		}
		claimed, err := p.Store.ClaimEmailAttempt(ctx, r.ID, []string{StatusFailed, StatusExpired}, now.Add(EmailRetryAfter))
		if err != nil || !claimed {
			p.logf("request=%s apology claim=%v err=%v", r.ID, claimed, err)
			continue
		}
		serr := p.Sender.Send(ctx, ApologyMessage(r.Email))
		if serr == nil {
			ok, uerr := p.Store.MarkFailureNotified(ctx, r.ID, now)
			p.logf("request=%s apology SENT notified=%v err=%v", r.ID, ok, uerr)
			continue
		}
		p.logf("request=%s apology send FAILED attempt=%d/%d err=%v", r.ID, r.EmailAttempts+1, MaxEmailAttempts, serr)
	}
}

func (p *Poller) maintenance(ctx context.Context, now time.Time) {
	if n, err := p.Store.ExpireIdleSessions(ctx, now.Add(-SessionIdleTTL)); err != nil {
		p.logf("maintenance: ExpireIdleSessions FAILED err=%v", err)
	} else if n > 0 {
		p.logf("maintenance: %d idle session(s) expired, transcripts dropped", n)
	}
	if n, err := p.Store.ScrubTerminalPII(ctx, now.Add(-PIIRetention)); err != nil {
		p.logf("maintenance: ScrubTerminalPII FAILED err=%v", err)
	} else if n > 0 {
		p.logf("maintenance: %d terminal row(s) scrubbed of email/ip", n)
	}
	if p.Hourly != nil {
		p.Hourly()
	}
}
