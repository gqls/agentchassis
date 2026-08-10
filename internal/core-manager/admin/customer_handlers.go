// FILE: internal/core-manager/admin/customer_handlers.go
//
// Customer identity endpoints, backed by the clients -> networks -> sites FK
// chain (owner ruling 2026-08-10, ai_site_selling_automation PLAN §1; columns
// added by migration 375). Deliberately separate from ClientHandlers: those
// endpoints serve the per-client-schema tenant machinery via the clients_info
// side table, which is a different population from website customers.
package admin

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CustomerHandlers handles admin operations for website customers
type CustomerHandlers struct {
	clientsDB *sql.DB
	logger    *zap.Logger
}

// NewCustomerHandlers creates new customer admin handlers
func NewCustomerHandlers(clientsDB *sql.DB, logger *zap.Logger) *CustomerHandlers {
	return &CustomerHandlers{clientsDB: clientsDB, logger: logger}
}

// CustomerSummary is one row of the customers list
type CustomerSummary struct {
	ID             string  `json:"id"`
	ExternalID     *string `json:"external_id"`
	Name           string  `json:"name"`
	Email          *string `json:"email"`
	Phone          *string `json:"phone"`
	Tier           *string `json:"tier"`
	CustomerStatus *string `json:"customer_status"`
	SiteCount      int     `json:"site_count"`
	CreatedAt      string  `json:"created_at"`
}

// CustomerSite is one site owned by a customer (via networks)
type CustomerSite struct {
	ID          string  `json:"id"`
	Domain      string  `json:"domain"`
	Status      *string `json:"status"`
	Email       *string `json:"email"`
	Phone       *string `json:"phone"`
	NetworkName string  `json:"network_name"`
}

// CreateCustomerRequest creates a clients row. Name is the only required field.
type CreateCustomerRequest struct {
	Name           string  `json:"name" binding:"required"`
	ExternalID     *string `json:"external_id"`
	Email          *string `json:"email"`
	Phone          *string `json:"phone"`
	Tier           *string `json:"tier"`
	CustomerStatus *string `json:"customer_status"`
	Notes          *string `json:"notes"`
}

// UpdateCustomerRequest is a partial update; absent fields are left unchanged.
type UpdateCustomerRequest struct {
	Name           *string `json:"name"`
	ExternalID     *string `json:"external_id"`
	Email          *string `json:"email"`
	Phone          *string `json:"phone"`
	Tier           *string `json:"tier"`
	CustomerStatus *string `json:"customer_status"`
	Notes          *string `json:"notes"`
}

const customerSiteCountSubquery = `(SELECT count(*) FROM sites s
	JOIN networks n ON s.network_id = n.id WHERE n.client_id = c.id)`

// HandleListCustomers returns every clients row with its owned-site count
func (h *CustomerHandlers) HandleListCustomers(c *gin.Context) {
	rows, err := h.clientsDB.QueryContext(c.Request.Context(), `
		SELECT c.id, c.external_id, c.name, c.email, c.phone, c.tier,
		       c.customer_status, to_char(c.created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
		       `+customerSiteCountSubquery+` AS site_count
		FROM clients c
		ORDER BY c.created_at DESC`)
	if err != nil {
		h.logger.Error("Failed to list customers", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list customers"})
		return
	}
	defer rows.Close()

	customers := []CustomerSummary{}
	for rows.Next() {
		var cust CustomerSummary
		var createdAt sql.NullString
		if err := rows.Scan(&cust.ID, &cust.ExternalID, &cust.Name, &cust.Email,
			&cust.Phone, &cust.Tier, &cust.CustomerStatus, &createdAt, &cust.SiteCount); err != nil {
			h.logger.Error("Failed to scan customer row", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read customer row"})
			return
		}
		cust.CreatedAt = createdAt.String
		customers = append(customers, cust)
	}
	if err := rows.Err(); err != nil {
		h.logger.Error("Customer list iteration failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list customers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"customers": customers, "count": len(customers)})
}

// HandleGetCustomer returns one customer with notes and its sites
func (h *CustomerHandlers) HandleGetCustomer(c *gin.Context) {
	customerID, ok := parseCustomerID(c)
	if !ok {
		return
	}

	var cust CustomerSummary
	var notes sql.NullString
	var createdAt sql.NullString
	err := h.clientsDB.QueryRowContext(c.Request.Context(), `
		SELECT c.id, c.external_id, c.name, c.email, c.phone, c.tier,
		       c.customer_status, c.notes,
		       to_char(c.created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
		       `+customerSiteCountSubquery+`
		FROM clients c WHERE c.id = $1`, customerID).
		Scan(&cust.ID, &cust.ExternalID, &cust.Name, &cust.Email, &cust.Phone,
			&cust.Tier, &cust.CustomerStatus, &notes, &createdAt, &cust.SiteCount)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}
	if err != nil {
		h.logger.Error("Failed to get customer", zap.Error(err), zap.String("customer_id", customerID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get customer"})
		return
	}
	cust.CreatedAt = createdAt.String

	rows, err := h.clientsDB.QueryContext(c.Request.Context(), `
		SELECT s.id, s.domain, s.status, s.email, s.phone, n.name
		FROM sites s JOIN networks n ON s.network_id = n.id
		WHERE n.client_id = $1 ORDER BY s.domain`, customerID)
	if err != nil {
		h.logger.Error("Failed to list customer sites", zap.Error(err), zap.String("customer_id", customerID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list customer sites"})
		return
	}
	defer rows.Close()

	sites := []CustomerSite{}
	for rows.Next() {
		var site CustomerSite
		if err := rows.Scan(&site.ID, &site.Domain, &site.Status, &site.Email,
			&site.Phone, &site.NetworkName); err != nil {
			h.logger.Error("Failed to scan customer site row", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read customer site row"})
			return
		}
		sites = append(sites, site)
	}
	if err := rows.Err(); err != nil {
		h.logger.Error("Customer sites iteration failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list customer sites"})
		return
	}

	resp := gin.H{"customer": cust, "sites": sites}
	if notes.Valid {
		resp["notes"] = notes.String
	}
	c.JSON(http.StatusOK, resp)
}

// HandleCreateCustomer inserts a clients row and echoes it back
func (h *CustomerHandlers) HandleCreateCustomer(c *gin.Context) {
	var req CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var id string
	err := h.clientsDB.QueryRowContext(c.Request.Context(), `
		INSERT INTO clients (name, external_id, email, phone, tier, customer_status, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`,
		req.Name, req.ExternalID, req.Email, req.Phone, req.Tier,
		req.CustomerStatus, req.Notes).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), "clients_external_id_key") {
			c.JSON(http.StatusConflict, gin.H{"error": "A customer with that external_id already exists"})
			return
		}
		h.logger.Error("Failed to create customer", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create customer"})
		return
	}

	h.logger.Info("Customer created", zap.String("customer_id", id), zap.String("name", req.Name))
	c.JSON(http.StatusCreated, gin.H{"id": id, "name": req.Name})
}

// HandleUpdateCustomer applies a partial update to a clients row
func (h *CustomerHandlers) HandleUpdateCustomer(c *gin.Context) {
	customerID, ok := parseCustomerID(c)
	if !ok {
		return
	}

	var req UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Name != nil && *req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name cannot be empty"})
		return
	}

	res, err := h.clientsDB.ExecContext(c.Request.Context(), `
		UPDATE clients SET
		  name            = COALESCE($2, name),
		  external_id     = COALESCE($3, external_id),
		  email           = COALESCE($4, email),
		  phone           = COALESCE($5, phone),
		  tier            = COALESCE($6, tier),
		  customer_status = COALESCE($7, customer_status),
		  notes           = COALESCE($8, notes),
		  updated_at      = now()
		WHERE id = $1`,
		customerID, req.Name, req.ExternalID, req.Email, req.Phone,
		req.Tier, req.CustomerStatus, req.Notes)
	if err != nil {
		if strings.Contains(err.Error(), "clients_external_id_key") {
			c.JSON(http.StatusConflict, gin.H{"error": "A customer with that external_id already exists"})
			return
		}
		h.logger.Error("Failed to update customer", zap.Error(err), zap.String("customer_id", customerID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update customer"})
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"updated": customerID})
}

// parseCustomerID validates the :customer_id path param as a uuid; on failure
// it writes the 400 itself and returns ok=false.
func parseCustomerID(c *gin.Context) (string, bool) {
	raw := c.Param("customer_id")
	id, err := uuid.Parse(raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "customer_id must be a uuid"})
		return "", false
	}
	return id.String(), true
}
