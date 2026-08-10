package main

// main.go — entrypoint for the noted.co.uk engine.
//
// Follows the house pattern for this estate's engines (site-engine, idea.uk,
// webdesign-chat): one self-contained Go binary, cross-compiled and shipped to
// the box, config from the environment, and FAIL LOUDLY AT STARTUP rather than
// binding a port in a broken state. A notes service that starts without a
// database is worse than one that never started, because the first thing a
// visitor does is trust it with something they cannot retype.
//
// BINDS LOOPBACK BY DEFAULT, ON PURPOSE. nginx terminates and proxies; the
// cloudflared tunnel dials out. Nothing on this box should accept a connection
// from the internet directly. webdesign-chat on this same machine listens on
// *:8081 and is protected only by ufw — one `ufw allow` from being public. This
// process must not repeat that, so the default is 127.0.0.1 and a non-loopback
// bind has to be typed out deliberately.

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		log.Fatalf("%s invalid: %q", key, v)
	}
	return n
}

func main() {
	dsn := os.Getenv("NOTED_DATABASE_URL")
	if dsn == "" {
		log.Fatal("NOTED_DATABASE_URL not set — refusing to start")
	}

	addr := env("NOTED_LISTEN_ADDR", "127.0.0.1:8090")

	// Per-account media allowance. This is the control that stops noted's
	// unbounded growth filling the 50 GB disk it shares with the webdesign.uk
	// shopfront — see schema.sql. It is a real safety valve, not a product
	// decision, so it is tunable without a copy sign-off.
	quotaMB := envInt("NOTED_MEDIA_QUOTA_MB", 50)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("cannot open database: %v", err)
	}
	db.SetMaxOpenConns(envInt("NOTED_MAX_CONNS", 10))
	db.SetConnMaxLifetime(time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("cannot reach the database: %v", err)
	}

	store := &Store{DB: db, QuotaBytes: int64(quotaMB) * 1024 * 1024}
	if err := store.Migrate(ctx); err != nil {
		log.Fatalf("schema migration failed: %v", err)
	}

	// Secure cookies are the default because the only way in is HTTPS via
	// Cloudflare. NOTED_INSECURE_COOKIES exists solely for a plain-HTTP local
	// test and is named so it cannot be set by accident and look harmless.
	srv := &Server{
		Store:          store,
		SecureCookies:  os.Getenv("NOTED_INSECURE_COOKIES") == "",
		SessionTTL:     time.Duration(envInt("NOTED_SESSION_DAYS", 30)) * 24 * time.Hour,
		MaxUploadBytes: int64(envInt("NOTED_MAX_UPLOAD_MB", 25)) * 1024 * 1024,
	}

	// Sweep expired sessions rather than letting the table grow for ever.
	go func() {
		for {
			n, err := store.PurgeExpiredSessions(context.Background())
			if err != nil {
				log.Printf("session purge failed: %v", err)
			} else if n > 0 {
				log.Printf("purged %d expired session(s)", n)
			}
			time.Sleep(6 * time.Hour)
		}
	}()

	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           srv.Routes(),
		ReadHeaderTimeout: 10 * time.Second,
		// Generous: a media upload on a slow mobile connection is a legitimate
		// slow request, and cutting it off loses the recording.
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
		IdleTimeout:  2 * time.Minute,
	}

	go func() {
		log.Printf("noted engine listening on %s (media quota %d MB/account)", addr, quotaMB)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Print("shutting down")
	shutCtx, shutCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer shutCancel()
	_ = httpSrv.Shutdown(shutCtx)
	_ = db.Close()
}
