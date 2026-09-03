package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gqls/agentchassis/internal/tools-api/api"
	"github.com/gqls/agentchassis/internal/tools-api/config"
	"github.com/gqls/agentchassis/internal/tools-api/db"
	"github.com/gqls/agentchassis/internal/tools-api/gripper"
	"github.com/gqls/agentchassis/internal/tools-api/store"
	"github.com/gqls/agentchassis/platform/mailer"
)

// shutdownGrace is how long in-flight requests and the poller get on SIGTERM.
// Compose's default stop grace is 10s and the ENTRYPOINT is exec-form, so the
// signal reaches this process directly; 8s leaves headroom under that.
const shutdownGrace = 8 * time.Second

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewDBPool(ctx)
	if err != nil {
		log.Fatalf("failed to create db pool: %v", err)
	}
	defer pool.Close()

	// The gripper route group and its poller are one feature: both or neither.
	// cfg.Gripper is nil unless GRIPPER_ANTHROPIC_API_KEY is set (config.go).
	var gdeps *api.GripperDeps
	var wg sync.WaitGroup
	if cfg.Gripper != nil {
		gdeps = api.NewGripperDeps(pool, cfg.Gripper)
		sender, err := mailer.New(cfg.Gripper.SMTP)
		if err != nil {
			log.Fatalf("gripper mailer: %v", err)
		}
		p := &gripper.Poller{
			Store:  &store.Gripper{Pool: pool},
			Sender: sender,
			Log:    log.Default(),
			Hourly: gdeps.Limiters.Sweep,
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.Run(ctx)
		}()
		log.Printf("gripper route group mounted (model=%s, smtp=%s:%s, from=%s)",
			cfg.Gripper.Model, cfg.Gripper.SMTP.Host, cfg.Gripper.SMTP.Port, cfg.Gripper.SMTP.From)
	} else {
		log.Printf("gripper route group NOT mounted (%s unset)", config.GripperAPIKeyEnv)
	}

	var ropts []api.RouterOption
	if cfg.Playground != nil {
		pdeps := api.NewPlaygroundDeps()
		ropts = append(ropts, api.WithPlayground(pdeps))
		wg.Add(1)
		go func() {
			defer wg.Done()
			t := time.NewTicker(time.Hour)
			defer t.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-t.C:
					pdeps.Limiters.Sweep()
				}
			}
		}()
		log.Printf("playground route group mounted (ollama=%s, model=%s, max_tokens=%d)",
			cfg.Playground.OllamaURL, cfg.Playground.Model, cfg.Playground.MaxTokens)
	} else {
		log.Printf("playground route group NOT mounted (%s unset)", config.PlaygroundOllamaURLEnv)
	}

	// The forms group is the only one with no LLM key and no mailer: it stores
	// what a static site's form posted and nothing else. The cluster's collector
	// pulls via GET /requests, resolves the token against site_form_routes in
	// clients_db — which this process cannot reach — and notifies. Opt-in on
	// FORMS_PULL_KEY, because a receiver nobody collects from is worse than no
	// receiver: the visitor is told their message was sent.
	if cfg.Forms != nil {
		fdeps := api.NewFormsDeps(pool)
		ropts = append(ropts, api.WithForms(fdeps))
		wg.Add(1)
		go func() {
			defer wg.Done()
			t := time.NewTicker(time.Hour)
			defer t.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-t.C:
					fdeps.Limiters.Sweep()
				}
			}
		}()
		log.Printf("forms route group mounted (max_body=%d, max_pull_batch=%d)",
			cfg.Forms.MaxBodyBytes, cfg.Forms.MaxPullBatch)
	} else {
		log.Printf("forms route group NOT mounted (%s unset)", config.FormsPullKeyEnv)
	}

	r := api.NewRouter(pool, cfg, gdeps, ropts...)
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// gin's r.Run is ListenAndServe on r.Handler(); this is the same server
	// with a shutdown path so the poller can finish a send and in-flight
	// requests can complete on SIGTERM.
	errCh := make(chan error, 1)
	go func() {
		log.Printf("tools-api listening on %s", srv.Addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	case <-ctx.Done():
		log.Printf("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownGrace)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown: %v", err)
	}
	stop() // cancels the poller's ctx if the exit came from a server error
	wg.Wait()
	log.Printf("tools-api stopped")
}
