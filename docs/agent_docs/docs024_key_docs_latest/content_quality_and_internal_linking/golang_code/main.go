package main

// main.go — entrypoint for the capture backend. Loads config from env, opens the
// JSON store, starts the HTTP service that nginx proxies the capture paths to.

import (
	"log"
	"net/http"
	"strings"
)

func loadConfig() Config {
	accept := map[string]bool{}
	for _, h := range strings.Split(env("ACCEPT_HOSTS", ""), ",") {
		if h = strings.TrimSpace(strings.ToLower(h)); h != "" {
			accept[strings.TrimPrefix(h, "www.")] = true
		}
	}
	var origins []string
	for _, o := range strings.Split(env("ALLOWED_ORIGINS", ""), ",") {
		if o = strings.TrimSpace(o); o != "" {
			origins = append(origins, o)
		}
	}
	return Config{
		InternalKey:  env("INTERNAL_API_KEY", ""),
		GeoHeader:    env("GEO_COUNTRY_HEADER", "CF-IPCountry"),
		MaxValueLen:  envInt("MAX_VALUE_LEN", 500),
		ThanksPath:   env("THANKS_PATH", "/"),
		AcceptHosts:  accept,
		AllowOrigins: origins,
	}
}

func main() {
	cfg := loadConfig()
	store, err := NewStore(env("PROBE_DB_PATH", "probe_events.json"))
	if err != nil {
		log.Fatal(err)
	}
	app := NewApp(cfg, store)
	addr := ":" + env("PORT", "8080")
	log.Printf("traffic-probe capture backend on %s (accept_hosts=%d)", addr, len(cfg.AcceptHosts))
	log.Fatal(http.ListenAndServe(addr, app.routes()))
}
