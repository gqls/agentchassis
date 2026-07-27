// FILE: platform/voicestyle/voicestyle.go
//
// One source for the house voice.
//
// The rules that govern how generated copy READS live in exactly one place: the
// `voice_style_block` row of `agent_default_configs` (migration 240). Everything
// that writes copy reads them from here.
//
// Why not a Go const, which is where they briefly lived on 2026-07-27: the
// owner's directive was "one place for the prompt, and probably not in go by
// choice", and the reason is operational rather than aesthetic. A prompt in Go
// changes only with a build and a roll; the other copy of the same rules lived
// in a DB row that changes instantly. Two substrates with different change
// latencies guarantee drift, and the drift is invisible until somebody reads
// site copy and blog copy side by side. See bugs_open/121.
//
// The fetcher indirection exists because the two consumers use different
// drivers: the chassis holds *sql.DB, content-creator holds a pgxpool.Pool.
// Rather than pick one and force an adapter on the other, each supplies a
// three-line closure and the caching lives here, once.
package voicestyle

import (
	"context"
	"sync"
	"time"
)

// ConfigName is the single row this package reads. Changing it means changing
// migration 240 too.
const ConfigName = "voice_style_block"

// SQL is the query both consumers should run. Exported so the two call sites
// cannot drift into asking for different things.
const SQL = `SELECT config->>'text' FROM agent_default_configs WHERE config_name = 'voice_style_block'`

// cacheTTL keeps a hot generation loop off the database while preserving the
// platform's "DB config is live immediately" property at a human timescale: an
// editor sees their change take effect within a minute.
const cacheTTL = 60 * time.Second

// Fetch returns the raw block text from the store.
type Fetch func(context.Context) (string, error)

var (
	mu       sync.RWMutex
	cached   string
	cachedAt time.Time
	// loadedOnce records whether the row has EVER been read successfully. It is
	// what makes a later DB blip non-fatal: we keep serving the last known
	// block instead of silently dropping the house voice from every prompt.
	loadedOnce bool
)

// Get returns the house voice block, and whether one is available.
//
// An empty return is deliberately NOT an error. A missing block must degrade to
// "generate without the house voice", never to "fail the generation": losing the
// voice on one page is recoverable, failing every content build is not. The
// caller decides whether to care.
func Get(ctx context.Context, fetch Fetch) (block string, ok bool) {
	mu.RLock()
	fresh := loadedOnce && time.Since(cachedAt) < cacheTTL
	stale, hadValue := cached, loadedOnce
	mu.RUnlock()

	if fresh {
		return stale, stale != ""
	}
	if fetch == nil {
		return stale, hadValue && stale != ""
	}

	text, err := fetch(ctx)
	if err != nil {
		// Serve the last known good block rather than nothing. A transient DB
		// error must not quietly strip the house style from generated copy —
		// that failure leaves no trace in the artefact and surfaces days later
		// as "the writing got worse", with nothing to point at.
		return stale, hadValue && stale != ""
	}

	mu.Lock()
	cached, cachedAt, loadedOnce = text, time.Now(), true
	mu.Unlock()

	return text, text != ""
}

// Invalidate drops the cache so the next Get re-reads. For tests, and for an
// operator who has just edited the row and does not want to wait out the TTL.
func Invalidate() {
	mu.Lock()
	cachedAt = time.Time{}
	mu.Unlock()
}
