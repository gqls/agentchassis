package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/internal/tools-api/config"
	"github.com/gqls/agentchassis/internal/tools-api/handlers"
	"github.com/gqls/agentchassis/internal/tools-api/middleware"
	"github.com/gqls/agentchassis/internal/tools-api/store"
	"github.com/gqls/agentchassis/platform/httpguard"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GripperDeps is what the gripper route group needs beyond the pool: its
// store, its LLM factory and the per-route limiters. Built by NewGripperDeps
// for production; tests hand in fakes. Nil means "not mounted".
type GripperDeps struct {
	Store     handlers.GripperStore
	Generator handlers.ChatGenerator
	Limiters  GripperLimiters
}

// GripperLimiters are the per-endpoint bands from DESIGN §2. Exposed so a
// process-wide sweeper (the poller's hourly hook) can reach them.
type GripperLimiters struct {
	Session, Chat, Submit *httpguard.Limiter
}

// Sweep drops idle keys from all three limiters.
func (l GripperLimiters) Sweep() {
	for _, lim := range []*httpguard.Limiter{l.Session, l.Chat, l.Submit} {
		if lim != nil {
			lim.Sweep()
		}
	}
}

// NewGripperLimiters builds the DESIGN §2 bands:
// /session 6/h + 20/d, /chat 60/h + 200/d, /submit 3/h + 10/d, per visitor.
func NewGripperLimiters() GripperLimiters {
	h, d := time.Hour, 24*time.Hour
	return GripperLimiters{
		Session: httpguard.NewLimiter(httpguard.Band{Window: h, Max: 6}, httpguard.Band{Window: d, Max: 20}),
		Chat:    httpguard.NewLimiter(httpguard.Band{Window: h, Max: 60}, httpguard.Band{Window: d, Max: 200}),
		Submit:  httpguard.NewLimiter(httpguard.Band{Window: h, Max: 3}, httpguard.Band{Window: d, Max: 10}),
	}
}

// NewGripperDeps wires the production dependencies for cfg.Gripper.
func NewGripperDeps(pool *pgxpool.Pool, g *config.GripperConfig) *GripperDeps {
	return &GripperDeps{
		Store:     &store.Gripper{Pool: pool},
		Generator: handlers.AnthropicChatGenerator(g),
		Limiters:  NewGripperLimiters(),
	}
}

// NewRouter constructs the gin engine with all routes for the tools-api service.
// gripper is nil when cfg.Gripper is nil (the group is simply not mounted).
// PlaygroundDeps is what the playground route group needs beyond the pool: an
// HTTP client for the model server and its per-route limiter. The client has no
// overall timeout because the per-request context in the handler carries one
// sized for a CPU-served reply (streaming, ~14 tok/s measured 2026-09-03).
type PlaygroundDeps struct {
	Client   *http.Client
	Limiters PlaygroundLimiters
}

// PlaygroundLimiters are the playground's per-endpoint bands. One route today
// (/chat); the struct exists so a /session for booked hours can join it without
// changing the sweeper wiring in main.
type PlaygroundLimiters struct {
	Chat *httpguard.Limiter
}

// Sweep drops idle per-IP entries; main runs it hourly like the gripper's.
func (l PlaygroundLimiters) Sweep() {
	if l.Chat != nil {
		l.Chat.Sweep()
	}
}

// NewPlaygroundLimiters builds the demo's bands: 60 replies an hour and 300 a
// day per address. Generous for a person, bounded for a script; the CPU is the
// real ceiling and these keep one address from owning it.
func NewPlaygroundLimiters() PlaygroundLimiters {
	h, d := time.Hour, 24*time.Hour
	return PlaygroundLimiters{
		Chat: httpguard.NewLimiter(httpguard.Band{Window: h, Max: 60}, httpguard.Band{Window: d, Max: 300}),
	}
}

// NewPlaygroundDeps is the production wiring for the playground group.
func NewPlaygroundDeps() *PlaygroundDeps {
	return &PlaygroundDeps{Client: &http.Client{}, Limiters: NewPlaygroundLimiters()}
}

// FormsDeps is what the forms route group needs beyond the pool: its store and
// its one per-route limiter. Nil means "not mounted".
type FormsDeps struct {
	Store    handlers.FormInboxStore
	Limiters FormsLimiters
}

// FormsLimiters is the forms group's band. One route needs one, but the struct
// exists so a second (a status endpoint, say) joins without changing the
// sweeper wiring in main — the shape the playground group settled on.
type FormsLimiters struct {
	Submit *httpguard.Limiter
}

// Sweep drops idle per-visitor entries; main runs it hourly like the others'.
func (l FormsLimiters) Sweep() {
	if l.Submit != nil {
		l.Submit.Sweep()
	}
}

// NewFormsLimiters builds the forms band: 5 submissions an hour and 20 a day
// per visitor. A person filling in a contact form does it once and occasionally
// twice; the gripper's /submit sits at 3/h and 10/d for the same reason, and
// this is a shade more generous because a form is likelier than a chat to be
// retried after a typo.
func NewFormsLimiters() FormsLimiters {
	h, d := time.Hour, 24*time.Hour
	return FormsLimiters{
		Submit: httpguard.NewLimiter(httpguard.Band{Window: h, Max: 5}, httpguard.Band{Window: d, Max: 20}),
	}
}

// NewFormsDeps is the production wiring for the forms group.
func NewFormsDeps(pool *pgxpool.Pool) *FormsDeps {
	return &FormsDeps{Store: &store.FormInbox{Pool: pool}, Limiters: NewFormsLimiters()}
}

// RouterOption attaches an optional route group to NewRouter without changing
// its signature for the callers that predate the option.
type RouterOption func(*routerOptions)

type routerOptions struct {
	playground *PlaygroundDeps
	forms      *FormsDeps
}

// WithPlayground mounts the playground group when the config for it is present.
func WithPlayground(deps *PlaygroundDeps) RouterOption {
	return func(o *routerOptions) { o.playground = deps }
}

// WithForms mounts the forms group when the config for it is present.
func WithForms(deps *FormsDeps) RouterOption {
	return func(o *routerOptions) { o.forms = deps }
}

func NewRouter(pool *pgxpool.Pool, cfg *config.Config, gripper *GripperDeps, opts ...RouterOption) *gin.Engine {
	var o routerOptions
	for _, opt := range opts {
		opt(&o)
	}
	r := gin.New()

	// gin.Logger() before Recovery so every request is recorded, including one
	// that panics. Added for bugs_open/083: the island ran gin.New() with only
	// Recovery, so `docker compose logs tools-api` showed nothing but the startup
	// banner. With no request log there was no denominator — the 503s could be
	// described as "bursty" from client-side sampling but no honest overall RATE
	// could be quoted, which is why 083 §1 carries an [UNMEASURED] marker.
	// Per-request status + latency here is what turns that into a measurement.
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Local health check — not platform/health per spec hard constraint 6.
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// apiGroup is retained for s3/s4/s5/s6 to attach middleware and routes.
	apiGroup := r.Group("/api/v1/tools/gauntlet")
	apiGroup.Use(middleware.CORSMiddleware(pool))

	// Rate limiting and input cap are ordered after CORS so that preflight
	// OPTIONS requests (already aborted with 204 by CORSMiddleware) never
	// reach these checks.
	apiGroup.Use(
		middleware.RateLimitMiddleware(cfg.RateLimitRPS, cfg.RateLimitBurst),
		middleware.InputCapMiddleware(cfg.MaxBodyBytes),
	)

	// /round — POST creates a round row and returns {round_id, provocation}.
	// OPTIONS is registered explicitly so gin matches the path and runs the
	// CORS middleware; CORSMiddleware aborts with 204 before the no-op body
	// is ever reached.
	apiGroup.POST("/round", handlers.RoundHandler(pool))
	apiGroup.OPTIONS("/round", func(c *gin.Context) {})

	// /position — POST returns {counter_position, challenge} via one LLM call.
	// OPTIONS is registered explicitly for preflight support.
	apiGroup.POST("/position", handlers.PositionHandler(pool, cfg))
	apiGroup.OPTIONS("/position", func(c *gin.Context) {})

	// /defend — POST returns {verdict, reasons} via one LLM call.
	// OPTIONS is registered explicitly for preflight support.
	apiGroup.POST("/defend", handlers.DefendHandler(pool, cfg))
	apiGroup.OPTIONS("/defend", func(c *gin.Context) {})

	// /publish — POST marks a completed round public and returns its slug.
	// Idempotent; refuses a round with no verdict (409, not 404).
	apiGroup.POST("/publish", handlers.PublishHandler(pool))
	apiGroup.OPTIONS("/publish", func(c *gin.Context) {})

	// /round/:slug — GET serves a PUBLISHED round for the public record page.
	// Read-only and no LLM call, so it is the only cheap route here.
	//
	// Registered AFTER POST /round: gin's tree keeps the static "/round" and the
	// "/round/:slug" child distinct, so this does not shadow the POST above.
	apiGroup.GET("/round/:slug", handlers.PublicRoundHandler(pool))
	apiGroup.OPTIONS("/round/:slug", func(c *gin.Context) {})

	if cfg.Gripper != nil && gripper != nil {
		mountGripper(r, pool, cfg.Gripper, gripper)
	}
	if cfg.Playground != nil && o.playground != nil {
		mountPlayground(r, pool, cfg.Playground, o.playground)
	}
	if cfg.Forms != nil && o.forms != nil {
		mountForms(r, pool, cfg.Forms, o.forms)
	}

	return r
}

// mountForms adds the fourth group: the static-site form receiver. Same two-group
// split as the gripper, and for the same reason — POST /submit is a browser
// route behind CORS, GET /requests is the cluster's and arrives with no Origin,
// so it is gated by X-Internal-Key INSTEAD. Putting /requests behind CORS 403s
// every pull; putting /submit outside it lets any page on the internet drive the
// intake.
//
// The browser half goes through mountBrowserGroup rather than re-deriving
// CORS→cap→band, which is what that helper's doc asks a third tool to do
// (council 63be72d1 round 4, architecture): the convergence is only worth
// anything if the next group inherits it instead of copying two Use() lines.
//
// ⚠ CORS IS A GATE HERE, NOT IDENTITY, and the distinction is the whole security
// shape of this group. CORSMiddleware resolves a site from the Origin header,
// which a caller sets freely; it is kept because it stops a page on an unrelated
// domain driving the intake casually, and because the resolved domain is what
// safeRedirect builds the thank-you URL from (so the redirect target can never
// be chosen by the request). WHOSE submission it is, is decided later and
// elsewhere: the cluster resolves the token against site_form_routes in
// clients_db, which this process cannot see. A group added here that ACTS on
// site_id — sends, pays, publishes — would be trusting Origin, and must not.
func mountForms(r *gin.Engine, pool *pgxpool.Pool, f *config.FormsConfig, deps *FormsDeps) {
	const prefix = "/api/v1/tools/forms"

	pub := mountBrowserGroup(r, pool, prefix, f.MaxBodyBytes)

	pub.POST("/submit", middleware.BandedRateLimit(deps.Limiters.Submit), handlers.FormSubmitHandler(deps.Store))
	pub.OPTIONS("/submit", func(c *gin.Context) {})

	internal := r.Group(prefix)
	internal.Use(middleware.InternalKey(f.PullKey))
	internal.GET("/requests", handlers.FormRequestsHandler(deps.Store, f.MaxPullBatch))
}

// mountGripper adds the second tool. Two gin groups on ONE path prefix, on
// purpose:
//
//   - the browser routes (/session, /chat, /submit) sit behind CORSMiddleware
//     exactly like the gauntlet's, then their own per-route bands and their
//     own body cap;
//   - GET /requests is the cluster's, arrives with no Origin header, and is
//     gated by X-Internal-Key INSTEAD of CORS. Put it in the CORS group and
//     every pull 403s; put the browser routes outside CORS and any page on the
//     internet can drive the intake. The split is the whole security shape of
//     the group and is the thing a third tool copying this must not merge.
//
// Ordering inside the browser group mirrors the gauntlet group: CORS first so
// a preflight (204-aborted there) never spends a rate-limit token or reads a
// body; then bands; then the cap.
func mountGripper(r *gin.Engine, pool *pgxpool.Pool, g *config.GripperConfig, deps *GripperDeps) {
	const prefix = "/api/v1/tools/gripper"

	pub := mountBrowserGroup(r, pool, prefix, g.MaxBodyBytes)

	pub.POST("/session", middleware.BandedRateLimit(deps.Limiters.Session), handlers.GripperSessionHandler(deps.Store))
	pub.OPTIONS("/session", func(c *gin.Context) {})

	pub.POST("/chat", middleware.BandedRateLimit(deps.Limiters.Chat), handlers.GripperChatHandler(deps.Store, g, deps.Generator))
	pub.OPTIONS("/chat", func(c *gin.Context) {})

	pub.POST("/submit", middleware.BandedRateLimit(deps.Limiters.Submit), handlers.GripperSubmitHandler(deps.Store))
	pub.OPTIONS("/submit", func(c *gin.Context) {})

	internal := r.Group(prefix)
	internal.Use(middleware.InternalKey(g.PullKey))
	internal.GET("/requests", handlers.GripperRequestsHandler(deps.Store))
}

// mountBrowserGroup is the one way a BROWSER-facing route group is built in
// this service: the group, then CORS (deployed-site Origin allowlist) so a
// preflight is answered before anything else runs, then the group's own body
// cap. The per-route band is applied at each route, because the bands differ
// per route (DESIGN §2) while these two layers never do. Both the gripper and
// the playground mount through it (council 63be72d1 round 3, reuse_agent: a
// third hand-copied CORS→cap→band block is the platform's own "two paths
// diverge quietly" case starting over). The gauntlet group predates it and
// keeps its flat RPS bucket; it is not a browser-band group of this shape.
//
// A THIRD browser-facing tool calls this rather than re-deriving the order
// (council 63be72d1 round 4, architecture): the convergence is only worth
// anything if the next tool added in a hurry inherits it instead of copying
// two Use() lines. The gauntlet group above is the one hand-built chain left,
// and deliberately so — its flat RPS bucket is a different shape.
//
// Cluster-facing routes (the gripper's /requests) never come through here:
// they need the internal-key group instead, and putting one behind CORS is
// the landmine on this file. A caller that needs both builds both.
func mountBrowserGroup(r *gin.Engine, pool *pgxpool.Pool, prefix string, maxBodyBytes int) *gin.RouterGroup {
	g := r.Group(prefix)
	g.Use(middleware.CORSMiddleware(pool))
	g.Use(middleware.InputCapMiddleware(maxBodyBytes))
	return g
}

// mountPlayground adds the third tool: finetuning.uk's public demo chat. One
// browser group only — there is no cluster-side pull here, so no internal-key
// group — built by mountBrowserGroup like the gripper's, with the per-route
// band on the one route. The model server is reached only from the host
// tools-api runs on, by the configured URL; this group is the only public door.
func mountPlayground(r *gin.Engine, pool *pgxpool.Pool, p *config.PlaygroundConfig, deps *PlaygroundDeps) {
	const prefix = "/api/v1/tools/playground"
	pub := mountBrowserGroup(r, pool, prefix, p.MaxBodyBytes)
	pub.POST("/chat", middleware.BandedRateLimit(deps.Limiters.Chat), handlers.PlaygroundChatHandler(p, deps.Client))
	pub.OPTIONS("/chat", func(c *gin.Context) {})
}
