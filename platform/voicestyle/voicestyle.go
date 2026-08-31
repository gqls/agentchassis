// FILE: platform/voicestyle/voicestyle.go
//
// One source for each platform-wide prompt block.
//
// The rules that govern how generated copy READS live in exactly one place: the
// `voice_style_block` row of `agent_default_configs` (migration 240). Everything
// that writes copy reads them from here. Generalised 2026-08-31 (owner ruling:
// best-in-class propagation, copy_quality_two_stage PLAN_2026-08-25): the same
// mechanism now carries any number of named blocks — the second is
// `build_standard_block`, injected as {{.build_standard}} — with one cache and
// one fetch shape, so a third block is a const and a migration, not a package.
//
// Why not a Go const, which is where the voice rules briefly lived on
// 2026-07-27: the owner's directive was "one place for the prompt, and probably
// not in go by choice", and the reason is operational rather than aesthetic. A
// prompt in Go changes only with a build and a roll; the other copy of the same
// rules lived in a DB row that changes instantly. Two substrates with different
// change latencies guarantee drift, and the drift is invisible until somebody
// reads site copy and blog copy side by side. See bugs_open/121. The same
// argument put the build standard here rather than in a Go file when the owner
// asked (PLAN_2026-08-25 §2): mechanism in Go, words in one DB row.
//
// The fetcher indirection exists because the consumers use different drivers:
// the chassis holds *sql.DB, content-creator holds a pgxpool.Pool. Rather than
// pick one and force an adapter on the other, each supplies a three-line
// closure and the caching lives here, once.
package voicestyle

import (
	"context"
	"sync"
	"time"
)

// ConfigName is the house-voice row. Changing it means changing migration 240
// too.
const ConfigName = "voice_style_block"

// BuildStandardConfigName is the best-in-class build standard row (migration
// 675). Templates opt in by writing {{.build_standard}}; one that does not
// mention it is unaffected.
const BuildStandardConfigName = "build_standard_block"

// SQL is the house-voice query both original consumers run. Exported so the
// call sites cannot drift into asking for different things. Kept verbatim for
// the existing callers; new named-block callers use SQLByName with the name as
// the one bind argument.
const SQL = `SELECT config->>'text' FROM agent_default_configs WHERE config_name = 'voice_style_block'`

// SQLByName is the parameterised form: one bind argument, the config name.
const SQLByName = `SELECT config->>'text' FROM agent_default_configs WHERE config_name = $1`

// cacheTTL keeps a hot generation loop off the database while preserving the
// platform's "DB config is live immediately" property at a human timescale: an
// editor sees their change take effect within a minute.
const cacheTTL = 60 * time.Second

// Fetch returns the raw block text from the store.
type Fetch func(context.Context) (string, error)

type entry struct {
	text string
	at   time.Time
	// loaded records whether the row has EVER been read successfully. It is
	// what makes a later DB blip non-fatal: we keep serving the last known
	// block instead of silently dropping the block from every prompt.
	loaded bool
}

var (
	mu     sync.RWMutex
	blocks = map[string]entry{}
)

// Get returns the house voice block, and whether one is available. Kept as the
// two original call sites' API; identical to GetBlock(ctx, ConfigName, fetch).
func Get(ctx context.Context, fetch Fetch) (block string, ok bool) {
	return GetBlock(ctx, ConfigName, fetch)
}

// GetBlock returns the named block, and whether one is available.
//
// An empty return is deliberately NOT an error. A missing block must degrade to
// "generate without the block", never to "fail the generation": losing it on
// one page is recoverable, failing every content build is not. The caller
// decides whether to care.
func GetBlock(ctx context.Context, name string, fetch Fetch) (block string, ok bool) {
	mu.RLock()
	e := blocks[name]
	fresh := e.loaded && time.Since(e.at) < cacheTTL
	mu.RUnlock()

	if fresh {
		return e.text, e.text != ""
	}
	if fetch == nil {
		return e.text, e.loaded && e.text != ""
	}

	text, err := fetch(ctx)
	if err != nil {
		// Serve the last known good block rather than nothing. A transient DB
		// error must not quietly strip the block from generated copy — that
		// failure leaves no trace in the artefact and surfaces days later as
		// "the writing got worse", with nothing to point at.
		return e.text, e.loaded && e.text != ""
	}

	mu.Lock()
	blocks[name] = entry{text: text, at: time.Now(), loaded: true}
	mu.Unlock()

	return text, text != ""
}

// Invalidate drops every cached block so the next Get re-reads. For tests, and
// for an operator who has just edited a row and does not want to wait out the
// TTL.
func Invalidate() {
	mu.Lock()
	blocks = map[string]entry{}
	mu.Unlock()
}
