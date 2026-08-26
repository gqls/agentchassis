// FILE: internal/core-manager/handlers/delivery_deps.go
//
// The production wiring behind DeliveryDeps. Split from the handler so the
// handler's tests exercise HTTP behaviour (method, status, what a human is
// shown, what is NOT leaked) without a database, while the SQL semantics stay
// covered where they belong: platform/delivery's own tests, which run against
// real Postgres in a rolled-back transaction.
package handlers

import (
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/delivery"
)

// dbDeliveryDeps is the real implementation.
type dbDeliveryDeps struct {
	db     *sql.DB
	logger *zap.Logger
	// now is swapped by tests. Time is an input to token expiry, so a handler
	// that reads the wall clock directly cannot be tested at a boundary — and
	// the boundary is the whole behaviour.
	now func() time.Time
}

// NewDBDeliveryDeps builds the production dependencies.
func NewDBDeliveryDeps(db *sql.DB, logger *zap.Logger) DeliveryDeps {
	return &dbDeliveryDeps{db: db, logger: logger, now: time.Now}
}

// ConfirmTransfer spends the token and stamps the site. The site id it returns
// is deliberately dropped: nothing on the customer-facing page varies by site,
// and a page that named the site would confirm to whoever holds the link that
// the token is real and which site it belongs to.
func (d *dbDeliveryDeps) ConfirmTransfer(c *gin.Context, token string) error {
	_, err := delivery.ConfirmTransfer(c.Request.Context(), d.db, token, d.now())
	return err
}

// Logger satisfies DeliveryDeps.
func (d *dbDeliveryDeps) Logger() *zap.Logger { return d.logger }

// ZipDownloadURL redeems a zip_download token for its stored presigned URL.
// The two failure modes reach the handler as distinct errors on purpose:
// delivery.ErrTokenNotFound gets the uniform failure page (no oracle), and
// delivery.ErrZipURLStale — reachable only by a VALID token's holder — gets the
// honest "being refreshed" page, never a redirect to a URL that would 403 as
// SignatureDoesNotMatch and read like broken credentials.
func (d *dbDeliveryDeps) ZipDownloadURL(c *gin.Context, token string) (string, error) {
	return delivery.LookupZipURL(c.Request.Context(), d.db, token, d.now())
}

// RecordStaleZipLink persists the stale hit where the fleet's sweeps read
// (agent_error_log — the immune system's triage sweeps recorded failures), so a
// dead link becomes a row somebody sees rather than a customer's private
// dead-end (DECISION_2026-08-21b §4's requirement; the full work-item filing
// rides the refresher build). Errors here are logged and swallowed: failing to
// RECORD the staleness must not break RENDERING the honest page.
func (d *dbDeliveryDeps) RecordStaleZipLink(c *gin.Context) {
	_, err := d.db.ExecContext(c.Request.Context(), `
		INSERT INTO agent_error_log (agent_type, action, error_message, error_code, severity, context)
		VALUES ('core-manager', 'zip_download',
		        'a customer zip_download link was visited but its stored presign has aged out; the refresher has not re-stamped it',
		        'ZIP_LINK_STALE', 'error',
		        jsonb_build_object('path', '/d/', 'remedy', 'rerun zip-deliverable-dispatch for the site and re-stamp via delivery.MintZipToken'))
	`)
	if err != nil {
		d.logger.Error("failed to record stale zip link", zap.Error(err))
	}
}
