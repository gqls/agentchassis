package voicestyle

import (
	"context"
	"errors"
	"testing"
)

// The generalisation's contract: blocks cache PER NAME, and the second name
// cannot bleed into the first. A single shared cache here would hand the build
// standard to every {{.voice_style}} template on the fleet, which is exactly
// the cross-contamination the named map exists to prevent.
func TestGetBlockCachesPerName(t *testing.T) {
	Invalidate()
	ctx := context.Background()

	voice, std := 0, 0
	vFetch := func(context.Context) (string, error) { voice++; return "VOICE", nil }
	sFetch := func(context.Context) (string, error) { std++; return "STANDARD", nil }

	if got, ok := GetBlock(ctx, ConfigName, vFetch); !ok || got != "VOICE" {
		t.Fatalf("voice block: got %q ok=%v", got, ok)
	}
	if got, ok := GetBlock(ctx, BuildStandardConfigName, sFetch); !ok || got != "STANDARD" {
		t.Fatalf("standard block: got %q ok=%v", got, ok)
	}
	// Cached: neither fetch runs again inside the TTL.
	GetBlock(ctx, ConfigName, vFetch)
	GetBlock(ctx, BuildStandardConfigName, sFetch)
	if voice != 1 || std != 1 {
		t.Fatalf("fetch counts: voice=%d standard=%d, want 1 and 1", voice, std)
	}
}

// Get is the two original call sites' API and must remain exactly the
// ConfigName block — this pins the delegation so a refactor cannot quietly
// point it elsewhere.
func TestGetDelegatesToConfigName(t *testing.T) {
	Invalidate()
	ctx := context.Background()
	GetBlock(ctx, ConfigName, func(context.Context) (string, error) { return "HOUSE", nil })
	if got, ok := Get(ctx, nil); !ok || got != "HOUSE" {
		t.Fatalf("Get after GetBlock(ConfigName): got %q ok=%v, want HOUSE true", got, ok)
	}
}

// A transient DB error serves the last known good block rather than silently
// stripping it from every prompt — per name.
func TestGetBlockServesStaleOnError(t *testing.T) {
	Invalidate()
	ctx := context.Background()
	name := BuildStandardConfigName

	GetBlock(ctx, name, func(context.Context) (string, error) { return "GOOD", nil })
	Invalidate()
	// Re-seed loaded state, then age it out so the failing fetch is consulted.
	GetBlock(ctx, name, func(context.Context) (string, error) { return "GOOD", nil })
	mu.Lock()
	e := blocks[name]
	e.at = e.at.Add(-2 * cacheTTL)
	blocks[name] = e
	mu.Unlock()

	got, ok := GetBlock(ctx, name, func(context.Context) (string, error) { return "", errors.New("db blip") })
	if !ok || got != "GOOD" {
		t.Fatalf("stale-on-error: got %q ok=%v, want GOOD true", got, ok)
	}

	// Control: with no prior load, the same failing fetch yields not-ok — the
	// stale path above passed because of the seed, not vacuously.
	Invalidate()
	if got, ok := GetBlock(ctx, name, func(context.Context) (string, error) { return "", errors.New("db blip") }); ok || got != "" {
		t.Fatalf("no-seed control: got %q ok=%v, want empty false", got, ok)
	}
}
