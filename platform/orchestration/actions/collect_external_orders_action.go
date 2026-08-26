// FILE: platform/orchestration/actions/collect_external_orders_action.go
//
// P4 of the webdesign.uk order intake (PLAN_2026-07-31_p4_order_intake, owner
// GO 2026-08-26): poll the site box's committed-brief list, release each brief
// whose payment has landed into build_queue, and acknowledge what was taken so
// the box stops offering it. Everything downstream of the INSERT is existing,
// running machinery (seed_build_queue → work item → dispatch), verified live
// end to end on 2026-07-31.
//
// THE JOIN IS AN ORDER REFERENCE, NOT THE BRIEF (owner ruling 2026-08-26,
// "the brief will change"): the chat mints BR-XXXXXX when it stores the brief,
// the customer quotes it at payment, billing_orders carries it as
// external_reference, and this action releases a brief to the pipeline ONLY
// when a PAID billing order names its reference. An unpaid brief stays on the
// box, re-listed every poll, costing nothing.
//
// The repeat-domain rule is the owner's 2026-07-31 ruling verbatim: a domain
// already 'queued' is a lost-ack retry (acknowledge, change nothing); a domain
// whose row is PAST 'queued' means a real build already happened, and the
// order is RECORDED FOR A HUMAN (needs_human_review, the checkpoint pattern)
// and never silently dropped — a paid customer swallowed by a unique
// constraint is this product's worst failure. A paid brief naming NO domain
// takes the same human-review path: inventing a domain here would put a name
// nobody chose on a paid order.
//
// Transport: the box's PUBLIC edge over HTTPS with a bearer token (P4 §2;
// measured 2026-08-26 that cluster pods have no route to the box's WireGuard
// address — the tunnel only carries box→cluster flows). The token arrives via
// env (terraform-owned secret; see 047-base-configs), NEVER via action config:
// config lives in the DB, and a credential in agent config is readable by
// every session and survives in history.
//
// Workflow config example (the order-intake-collector agent's single step):
//
//	"collect": {
//	    "action": "collect_external_orders",
//	    "config": {
//	        "orders_url": "https://preview.webdesign.uk/internal/orders",
//	        "max_orders": 10
//	    },
//	    "output_field": "collect_result",
//	    "next_step": "complete"
//	}

package actions

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/fetchguard"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var CollectExternalOrdersInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"orders_url"},
	Optional: []string{"token_env_var", "max_orders", "attention_site_domain"},
	Defaults: map[string]interface{}{
		"token_env_var":         "WEBDESIGN_BOX_ORDERS_TOKEN",
		"max_orders":            10,
		"attention_site_domain": "webdesign.uk",
	},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("collect_external_orders", CollectExternalOrdersInputSpec)
}

// boxBriefOrder mirrors the box's BriefOrder wire shape (orders_http.go).
type boxBriefOrder struct {
	Reference    string `json:"reference"`
	ContactEmail string `json:"contact_email"`
	ContactName  string `json:"contact_name"`
	Domain       string `json:"domain"`
	Brief        string `json:"brief"`
	CreatedAt    string `json:"created_at"`
}

// collectOrdersHTTP indirects the box calls so tests can substitute a stub
// server; production uses a fetchguard-guarded client (never a bare
// &http.Client{} — the box is our own host, but a compromised one redirecting
// inward is exactly fetchguard's case, and using it costs nothing).
var collectOrdersHTTP = func() *http.Client {
	cfg := fetchguard.DefaultConfig()
	cfg.MaxResponseBytes = 4 * 1024 * 1024 // a brief list is KBs; 4MB is generous
	return fetchguard.NewClient(cfg)
}

func CollectExternalOrdersAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "collect_external_orders"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	config := params.StepConfig.Config
	ordersURL := datahelpers.GetStringField(config, "orders_url", "")
	if ordersURL == "" {
		return nil, fmt.Errorf("orders_url is required")
	}
	tokenEnvVar := datahelpers.GetStringField(config, "token_env_var", "WEBDESIGN_BOX_ORDERS_TOKEN")
	maxOrders := datahelpers.GetIntField(config, "max_orders", 10)
	attentionDomain := datahelpers.GetStringField(config, "attention_site_domain", "webdesign.uk")

	token := os.Getenv(tokenEnvVar)
	if token == "" {
		// Fail loudly, never poll unauthenticated: a 401 loop reads as "the
		// box is broken" when the truth is "this pod was never given the key".
		return nil, fmt.Errorf("%s is not set in this pod's environment — the collector cannot authenticate to the box", tokenEnvVar)
	}

	client := collectOrdersHTTP()
	orders, err := fetchBoxOrders(ctx, client, ordersURL, token)
	if err != nil {
		return nil, fmt.Errorf("list box orders: %w", err)
	}
	if len(orders) == 0 {
		return map[string]interface{}{"listed": 0, "queued": 0, "awaiting_payment": 0, "human_review": 0, "acked": 0}, nil
	}
	if len(orders) > maxOrders {
		logger.Info("CollectExternalOrders: truncating batch",
			zap.Int("listed", len(orders)), zap.Int("max_orders", maxOrders))
		orders = orders[:maxOrders]
	}

	queued, awaitingPayment, humanReview := 0, 0, 0
	var ackRefs []string

	for _, o := range orders {
		if o.Reference == "" || strings.TrimSpace(o.Brief) == "" {
			// A malformed row would re-list for ever; that is the box's bug to
			// fix, and this log line is how it gets found. Not acked: acking
			// would silently discard a customer's brief.
			logger.Warn("CollectExternalOrders: malformed order skipped",
				zap.String("reference", o.Reference))
			continue
		}

		// THE GATE: a paid billing order must name this reference.
		var paidOrderID string
		err := params.DB.QueryRowContext(ctx, `
			SELECT id FROM billing_orders
			WHERE external_reference = $1 AND status = 'paid'
			ORDER BY paid_at DESC LIMIT 1`, o.Reference).Scan(&paidOrderID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				awaitingPayment++
				continue // stays on the box; next poll re-checks
			}
			return nil, fmt.Errorf("paid check for %s: %w", o.Reference, err)
		}

		domain := strings.ToLower(strings.TrimSpace(o.Domain))
		disposition := "" // for logging
		switch {
		case domain == "":
			// Paid, but nobody chose a domain: a human decides (the domain
			// programme exists for exactly this), never an invented name.
			if err := fileOrderAttention(ctx, params, o, attentionDomain,
				"paid order names no domain — choose one (domain programme) and queue the build by hand"); err != nil {
				return nil, err
			}
			humanReview++
			disposition = "human_review_no_domain"
		default:
			var existingStatus string
			err := params.DB.QueryRowContext(ctx,
				`SELECT status FROM build_queue WHERE domain = $1`, domain).Scan(&existingStatus)
			switch {
			case errors.Is(err, sql.ErrNoRows):
				direction := map[string]interface{}{
					"objective":       o.Brief,
					"source":          "site-chat-intake",
					"order_reference": o.Reference,
					"customer_email":  o.ContactEmail,
					"customer_name":   o.ContactName,
				}
				dirJSON, mErr := json.Marshal(direction)
				if mErr != nil {
					return nil, fmt.Errorf("marshal direction for %s: %w", o.Reference, mErr)
				}
				if _, iErr := params.DB.ExecContext(ctx, `
					INSERT INTO build_queue (domain, direction, priority, status)
					VALUES ($1, $2::jsonb, 10, 'queued')
					ON CONFLICT (domain) DO NOTHING`, domain, string(dirJSON)); iErr != nil {
					return nil, fmt.Errorf("queue %s for %s: %w", domain, o.Reference, iErr)
				}
				queued++
				disposition = "queued"
			case err != nil:
				return nil, fmt.Errorf("build_queue check for %s: %w", domain, err)
			case existingStatus == "queued":
				// Lost-ack retry: the work is already waiting. Nothing to do
				// but acknowledge.
				disposition = "already_queued"
			default:
				// The ruled worst-failure guard: this domain was built before.
				// Rebuild, refund or duplicate is a human's call.
				if err := fileOrderAttention(ctx, params, o, attentionDomain,
					fmt.Sprintf("paid order for domain %q whose build_queue row is already %q — rebuild, refund or duplicate is a human decision", domain, existingStatus)); err != nil {
					return nil, err
				}
				humanReview++
				disposition = "human_review_repeat_domain"
			}
		}

		ackRefs = append(ackRefs, o.Reference)
		logger.Info("CollectExternalOrders: order processed",
			zap.String("reference", o.Reference),
			zap.String("domain", domain),
			zap.String("disposition", disposition),
			zap.String("billing_order_id", paidOrderID))
	}

	acked := 0
	if len(ackRefs) > 0 {
		ackURL := strings.TrimRight(ordersURL, "/") + "/ack"
		acked, err = ackBoxOrders(ctx, client, ackURL, token, ackRefs)
		if err != nil {
			// The DB work is committed; the box will re-list these next poll
			// and every branch above is idempotent (paid check, ON CONFLICT,
			// already-queued, work-item dedup key). Report the failure so it
			// is visible, but do not pretend the collection did not happen.
			return nil, fmt.Errorf("processed %d orders but ack failed (safe to retry, all branches idempotent): %w", len(ackRefs), err)
		}
	}

	result := map[string]interface{}{
		"listed":           len(orders),
		"queued":           queued,
		"awaiting_payment": awaitingPayment,
		"human_review":     humanReview,
		"acked":            acked,
	}
	logger.Info("CollectExternalOrders: complete", zap.Any("result", result))
	return result, nil
}

// fileOrderAttention records a paid order a machine must not decide about,
// as a needs_human_review work item (the checkpoint_for_review pattern —
// same item_type, same 'human-review' handler, so it lands on the same
// screen). Dedup key order_attention_<reference>: re-collection after a lost
// ack cannot file twice.
func fileOrderAttention(ctx context.Context, params ActionParams, o boxBriefOrder, attentionDomain, reason string) error {
	logger := params.Logger
	var siteID uuid.UUID
	err := params.DB.QueryRowContext(ctx,
		`SELECT id FROM sites WHERE domain = $1`, attentionDomain).Scan(&siteID)
	if err != nil {
		return fmt.Errorf("attention site %q not found (the work item needs a site to hang on): %w", attentionDomain, err)
	}

	spec := map[string]interface{}{
		"order_reference": o.Reference,
		"customer_email":  o.ContactEmail,
		"customer_name":   o.ContactName,
		"domain":          o.Domain,
		"brief":           o.Brief,
		"reason":          reason,
	}
	specJSON, err := json.Marshal(spec)
	if err != nil {
		return err
	}

	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := insertWorkItem(ctx, tx, workItem{
		siteID:   siteID,
		source:   "order-intake",
		pipeline: "build",
		itemType: "needs_human_review",
		severity: "high",
		summary:  fmt.Sprintf("Paid order %s needs a human: %s", o.Reference, reason),
		spec:     string(specJSON),
		priority: 10,
		// EMPTY on purpose (the voice_tells posture, LANDMINES "A
		// HITL-terminal item type with a NON-EMPTY handler_agent"): a
		// descriptive value like 'human-review' is selected by the
		// detected-item-promoter, held as "handler not a live agent" —
		// which reads as a routing defect — and after 3 days
		// held-pair-canary-escalation asks a human to hand-canary the very
		// thing that must never be auto-dispatched. An empty handler is
		// excluded by the promoter's scored CTE by construction.
		// (checkpoint_for_review's 'human-review' literal is the measured
		// safe-by-accident case that landmine documents — copied here
		// first, corrected on the council round's prior_art advisory,
		// corr aa5a40a2.)
		handlerAgent: "",
		status:       "needs_human_review",
		createdBy:    "collect_external_orders",
		itemKey:      "order_attention_" + o.Reference,
	}, logger); err != nil {
		return fmt.Errorf("file order attention for %s: %w", o.Reference, err)
	}
	return tx.Commit()
}

func fetchBoxOrders(ctx context.Context, client *http.Client, url, token string) ([]boxBriefOrder, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4*1024*1024))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		// 503 means the box has no token configured; 401 means ours is wrong.
		// Different fixes, identical symptoms otherwise — keep the status.
		return nil, fmt.Errorf("box orders endpoint returned %d: %s", resp.StatusCode, truncateForLog(strings.TrimSpace(string(body)), 200))
	}
	var out struct {
		Orders []boxBriefOrder `json:"orders"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("box orders response not decodable: %w", err)
	}
	return out.Orders, nil
}

func ackBoxOrders(ctx context.Context, client *http.Client, url, token string, refs []string) (int, error) {
	payload, err := json.Marshal(map[string]interface{}{"references": refs})
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("box ack endpoint returned %d: %s", resp.StatusCode, truncateForLog(strings.TrimSpace(string(body)), 200))
	}
	var out struct {
		Collected int `json:"collected"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return 0, fmt.Errorf("box ack response not decodable: %w", err)
	}
	return out.Collected, nil
}
