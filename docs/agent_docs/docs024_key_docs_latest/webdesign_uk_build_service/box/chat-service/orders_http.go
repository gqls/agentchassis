package main

// orders_http.go — the collection contract the cluster's collector calls
// (PLAN_2026-07-31_p4_order_intake §8: a list of uncollected orders and an
// acknowledge path; the collection marker itself lives in orders.go). The box
// only ever ANSWERS — it never dials in, per the standing trust boundary.
//
// TRANSPORT: the PUBLIC edge, over the Cloudflare tunnel, exactly as P4 §2
// states ("GETs the box's orders endpoint over HTTPS") and as idea.uk already
// runs its INTERNAL_API_KEY routes. Measured 2026-08-26: cluster pods have NO
// route to this box's wg0 address (10.13.13.4 — the tunnel was built for
// box→cluster flows only, and a pod-side probe times out), so a
// WireGuard-only collector cannot exist without new routing infra. The lock
// is therefore the bearer token alone: ORDERS_API_TOKEN, 64 random hex chars,
// constant-time compared. Unset token = endpoints answer 503 and say so at
// startup: the feature is OFF, loudly, never open. A first cut of this file
// also refused any request carrying CF-Connecting-IP ("wg-only, public
// traffic can never belong here") — retired the same evening it was written,
// because it refused the only transport that actually exists.

import (
	"crypto/subtle"
	"encoding/json"
	"log"
	"net/http"
)

type ordersAPI struct {
	store *OrderStore
	token string
}

// authorize is the bearer check. Returns false after writing the refusal, so
// handlers read as: if !authorize { return }.
func (a *ordersAPI) authorize(w http.ResponseWriter, r *http.Request) bool {
	if a.token == "" {
		http.Error(w, "orders collection not configured", http.StatusServiceUnavailable)
		return false
	}
	const prefix = "Bearer "
	auth := r.Header.Get("Authorization")
	if len(auth) <= len(prefix) || auth[:len(prefix)] != prefix ||
		subtle.ConstantTimeCompare([]byte(auth[len(prefix):]), []byte(a.token)) != 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return false
	}
	return true
}

// handleList — GET /internal/orders. Serves every uncollected brief, oldest
// first. Re-serving a brief whose ack was lost is the DESIGN (the collector
// is idempotent end to end), so there is no pagination and no cursor: the
// uncollected set is bounded by how fast the collector drains it.
func (a *ordersAPI) handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !a.authorize(w, r) {
		return
	}
	orders := a.store.ListUncollected()
	if orders == nil {
		orders = []BriefOrder{} // an empty list, never null — the collector's decoder should not need a special case
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"orders": orders})
}

type ackRequest struct {
	References []string `json:"references"`
}

// handleAck — POST /internal/orders/ack {"references": ["BR-..."]}. Marks the
// named orders collected. Idempotent (orders.go): a retry after a lost
// response changes nothing and reports collected=0.
func (a *ordersAPI) handleAck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !a.authorize(w, r) {
		return
	}
	var req ackRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64*1024)).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if len(req.References) == 0 {
		http.Error(w, "references is empty", http.StatusBadRequest)
		return
	}
	changed, err := a.store.Ack(req.References)
	if err != nil {
		log.Printf("orders ack: persist failed after marking %d: %v", changed, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"collected": changed})
}
