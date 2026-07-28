package main

// main.go — entrypoint. `idea` starts the HTTP service; `idea internal D A S`
// runs the engine once against one of our own domains (no billing).

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func loadConfig() Config {
	maxActive, _ := strconv.Atoi(env("MAX_ACTIVE_ORDERS", "8"))
	staleReview, _ := strconv.Atoi(env("STALE_REVIEW_DAYS", "14"))
	stalePayment, _ := strconv.Atoi(env("STALE_PAYMENT_DAYS", "7"))
	price, _ := strconv.Atoi(env("REPORT_PRICE_GBP", "199"))
	// Example places ship OFF: 0 disables the tier. Turning it on is a deliberate
	// act once the copy and consent wording are live, not a default that could
	// start selling cheap reports the moment this binary rolls.
	examplePrice, _ := strconv.Atoi(env("EXAMPLE_PRICE_GBP", "0"))
	examplePlaces, _ := strconv.Atoi(env("EXAMPLE_MAX_PLACES", "0"))
	pub := env("PUBLIC_BASE_URL", "http://localhost:8080")
	var origins []string
	for _, o := range strings.Split(env("ALLOWED_ORIGINS", pub), ",") {
		if o = strings.TrimSpace(o); o != "" {
			origins = append(origins, o)
		}
	}
	return Config{
		PriceGBP:         price,
		AutoDeliver:      strings.ToLower(os.Getenv("AUTO_DELIVER")) == "true",
		ReviewBeforePay:  strings.ToLower(env("REVIEW_BEFORE_PAY", "true")) == "true",
		PublicBaseURL:    pub,
		InternalAPIKey:   os.Getenv("INTERNAL_API_KEY"),
		OperatorEmail:    env("OPERATOR_EMAIL", "ops@idea.uk"),
		ContactEmail:     env("CONTACT_EMAIL", ""),
		Slots:            env("MONTH_SLOTS", ""),
		MaxActive:        maxActive,
		ExamplePriceGBP:  examplePrice,
		ExampleMaxPlaces: examplePlaces,
		StaleReviewDays:  staleReview,
		StalePaymentDays: stalePayment,
		AllowedOrigins:   origins,
	}
}

func makeProvider(cfg Config) Provider {
	sk, wh := os.Getenv("STRIPE_SECRET_KEY"), os.Getenv("STRIPE_WEBHOOK_SECRET")
	if sk != "" && wh != "" {
		return &StripeProvider{secretKey: sk, webhookSecret: wh,
			publicBaseURL: cfg.PublicBaseURL}
	}
	log.Println("No Stripe keys — using FakeProvider (local/testing only).")
	return &FakeProvider{publicBaseURL: cfg.PublicBaseURL}
}

func main() {
	if len(os.Args) >= 5 && os.Args[1] == "internal" {
		rep, err := RunMethod(os.Args[2], os.Args[3], os.Args[4])
		if err != nil {
			log.Fatal(err)
		}
		os.Stdout.WriteString(rep.Text)
		return
	}

	cfg := loadConfig()
	store, err := NewStore(env("IDEA_DB_PATH", "idea_orders.json"))
	if err != nil {
		log.Fatal(err)
	}
	app := NewApp(cfg, store, makeProvider(cfg))
	// Recover runs abandoned by the previous process, then release slots held by
	// cold orders — at startup and hourly thereafter. Without this, ActiveCount
	// only ever grows and the service quietly closes itself to new work (exactly
	// what happened on 2026-07-26). Both halves are inside StartSweeper because
	// recovery has to happen before the first sweep.
	app.StartSweeper()
	addr := ":" + env("PORT", "8080")
	log.Printf("idea.uk service on %s (review_before_pay=%v, auto_deliver=%v, price=£%d)", addr, cfg.ReviewBeforePay, cfg.AutoDeliver, cfg.PriceGBP)
	log.Fatal(http.ListenAndServe(addr, app.routes()))
}
