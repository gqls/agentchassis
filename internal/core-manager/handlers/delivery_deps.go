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
