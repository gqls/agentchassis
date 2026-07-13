package working_dir

// env.go — tiny config helpers. These are copied verbatim from idea.uk's
// engine.go (which the probe drops), so the rest of the service keeps the same
// env-var conventions as idea.uk without pulling the engine in.

import (
	"os"
	"strconv"
)

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
