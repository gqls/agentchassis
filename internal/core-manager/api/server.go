package api

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/gqls/agentchassis/internal/core-manager/handlers"

	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/internal/core-manager/admin"
	"github.com/gqls/agentchassis/internal/core-manager/database"
	"github.com/gqls/agentchassis/internal/core-manager/middleware"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/kafka"
	"go.uber.org/zap"
)

// Server represents the Core Manager API server
type Server struct {
	ctx           context.Context
	cfg           *config.ServiceConfig
	logger        *zap.Logger
	router        *gin.Engine
	httpServer    *http.Server
	personaRepo   models.PersonaRepository
	kafkaProducer kafka.Producer

	// deliveryServer is the second listener: customer delivery routes ONLY, on
	// their own port, so a misconfiguration at the box can expose nothing but
	// those routes. nil when server.delivery_port is unset, which is the
	// default and means the delivery routes are mounted nowhere at all.
	deliveryServer *http.Server
}

// deliveryRoutePrefixes are the customer-facing paths that must never appear on
// the main (admin) router. They are short because they are read aloud and
// retyped out of an email, and they are listed here because that shortness is
// exactly what makes an accidental re-mount easy to miss in review.
var deliveryRoutePrefixes = []string{"/c/", "/d/"}

// assertNoDeliveryRoutes is the mechanism behind the comment in setupRoutes.
//
// The containment this service now buys is precisely "the admin port serves no
// customer route". That is a property, and a property held only by review is one
// this tree has repeatedly failed to hold — so it is checked against the real
// route table at construction, and a violation refuses to build the server.
//
// Fail-closed is deliberate. The only way to trip this is to mount a customer
// route on the admin port, which is the exact mistake it exists to catch; a roll
// that crash-loops says so immediately, where a logged warning would be read by
// nobody and the hole would serve traffic in the meantime.
// It refuses two shapes, not one. The direct registration is obvious; the second
// is a WILDCARD whose static prefix sits at or above a delivery path — gin routes
// `/*any` to a single handler, so a catch-all would serve /c/ by prefix dispatch
// while the route table contains no /c/ entry at all, and a check that only
// compared paths would report the admin port clean while it answered customer
// links. The council's guardian seat raised exactly this (council 25cd3044,
// round 1, medium): the guard "is blind to a wildcard route that later dispatches
// by path prefix internally". Measured when that objection landed: core-manager
// registers no wildcard and no r.Any() route today, so this arm defends a
// property that currently holds rather than fixing a live hole.
//
// ⚠ STATED LIMIT: gin does not list a NoRoute handler in Routes(), so a NoRoute
// that proxied by path would be invisible here. Nothing registers one today
// (grep NoRoute in internal/core-manager: no hits), and if something ever does,
// this function cannot see it — that is a residual, not a claim.
func assertNoDeliveryRoutes(routes gin.RoutesInfo) error {
	for _, r := range routes {
		for _, prefix := range deliveryRoutePrefixes {
			if strings.HasPrefix(r.Path, prefix) {
				return fmt.Errorf(
					"delivery route %s %s is mounted on the main admin router: it belongs on the delivery listener only (RFC_054 Q2, owner ruling 2026-08-25)",
					r.Method, r.Path)
			}
		}
		// A wildcard's static prefix is everything before "/*". If a delivery
		// path starts with that prefix, this route can capture it.
		if i := strings.Index(r.Path, "/*"); i >= 0 {
			static := r.Path[:i+1]
			for _, prefix := range deliveryRoutePrefixes {
				if strings.HasPrefix(prefix, static) {
					return fmt.Errorf(
						"wildcard route %s %s on the main admin router can capture delivery path %s by prefix dispatch: the delivery routes belong on the delivery listener only (RFC_054 Q2, owner ruling 2026-08-25)",
						r.Method, r.Path, prefix)
				}
			}
		}
	}
	return nil
}

// newDeliveryEngine builds the router for the delivery listener. It registers
// the delivery routes and NOTHING else — no health endpoint, no metrics, no
// catch-all. "Delivery-only" is the whole value of this listener, so anything
// added here gives back some of what the change bought.
//
// Routes come from the handler's own RegisterRoutes, which is the single
// definition of the route table (the guardian seat's objection on council
// ea99befa): there is no second copy here to drift.
func newDeliveryEngine(h *handlers.DeliveryHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	h.RegisterRoutes(r)
	return r
}

// newDeliveryServer applies the opt-in. An empty port returns nil: no listener,
// and — because these routes are mounted nowhere else — no customer route served
// anywhere in this process.
//
// This is a function rather than an inline `if` so that the default-OFF property
// can be asserted by a test. The safety of this whole mechanism rests on which
// way the empty case falls, and that is not something to leave to a comment.
func newDeliveryServer(port string, h *handlers.DeliveryHandler) *http.Server {
	if port == "" {
		return nil
	}
	return &http.Server{
		Addr:    ":" + port,
		Handler: newDeliveryEngine(h),
	}
}

// NewServer creates a new API server instance
//func NewServer(ctx context.Context, cfg *config.ServiceConfig, logger *zap.Logger, templatesDB, clientsDB *sql.DB) (*Server, error) {

func NewServer(ctx context.Context, cfg *config.ServiceConfig, logger *zap.Logger, templatesDB, clientsDB *sql.DB) (*Server, error) {
	// Initialize repositories
	personaRepo := database.NewPersonaRepository(templatesDB, clientsDB, logger)

	// Create Gin router
	router := gin.New()
	router.Use(gin.Recovery())

	// Initialize Kafka Producer for admin handlers
	kafkaProducer, err := kafka.NewProducer(cfg.Infrastructure.KafkaBrokers, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer for admin handlers: %w", err)
	}

	// Initialize auth middleware config
	authConfig, err := middleware.NewAuthMiddlewareConfig(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize auth middleware: %w", err)
	}

	server := &Server{
		ctx:           ctx,
		cfg:           cfg,
		logger:        logger,
		router:        router,
		personaRepo:   personaRepo,
		kafkaProducer: kafkaProducer,
	}

	// Setup routes with configured auth middleware
	server.setupRoutes(authConfig)

	// The admin port must carry no customer-facing route. Checked against the
	// real route table rather than trusted (see assertNoDeliveryRoutes).
	if err := assertNoDeliveryRoutes(router.Routes()); err != nil {
		return nil, err
	}

	// Create HTTP server
	server.httpServer = &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	// The delivery listener. OPT-IN: an unset delivery_port means the customer
	// routes are served nowhere, which is the safe direction — the box then gets
	// a 404 rather than the admin API. See config.ServerConfig.DeliveryPort.
	deliveryHandler := handlers.NewDeliveryHandler(
		handlers.NewDBDeliveryDeps(personaRepo.ClientsDB(), logger))
	server.deliveryServer = newDeliveryServer(cfg.Server.DeliveryPort, deliveryHandler)
	if server.deliveryServer != nil {
		logger.Info("Delivery listener configured (customer routes only)",
			zap.String("address", server.deliveryServer.Addr))
	} else {
		logger.Info("Delivery listener NOT configured; customer delivery routes are served nowhere",
			zap.String("config_key", "server.delivery_port"))
	}

	return server, nil
}

// setupRoutes configures all API routes with auth config
func (s *Server) setupRoutes(authConfig *middleware.AuthMiddlewareConfig) {
	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(s.cfg, s.logger)
	templateHandler := handlers.NewTemplateHandler(s.personaRepo, s.logger)
	instanceHandler := handlers.NewInstanceHandler(s.personaRepo, s.logger)

	// Initialize admin handlers
	personaRepoImpl := s.personaRepo.(*database.PersonaRepository)
	clientHandlers := admin.NewClientHandlers(personaRepoImpl.ClientsDB(), s.logger)
	customerHandlers := admin.NewCustomerHandlers(personaRepoImpl.ClientsDB(), s.logger)
	systemHandlers := admin.NewSystemHandlers(personaRepoImpl.ClientsDB(), personaRepoImpl.TemplatesDB(), s.kafkaProducer, s.logger)
	agentAdminHandlers := admin.NewAgentHandlers(personaRepoImpl.ClientsDB(), personaRepoImpl.TemplatesDB(), s.kafkaProducer, s.logger, s.personaRepo)
	siteAdminHandlers := admin.NewSiteAdminHandlers(personaRepoImpl.ClientsDB(), s.logger)
	pageAdminHandlers := admin.NewPageAdminHandlers(personaRepoImpl.ClientsDB(), s.logger)
	specAdminHandlers := admin.NewSpecAdminHandlers(personaRepoImpl.ClientsDB(), s.logger)
	assetAdminHandlers := admin.NewAssetAdminHandlers(personaRepoImpl.ClientsDB(), s.logger)
	pipelineAdminHandlers := admin.NewPipelineAdminHandlers(personaRepoImpl.ClientsDB(), s.logger)

	// Initialize the bootstrap handler
	bootstrapHandler := handlers.NewBootstrapHandler(s.logger, personaRepoImpl.ClientsDB())

	// Health check (no auth)
	s.router.GET("/health", healthHandler.HandleHealth)

	// NOTE: Admin dashboard SPA and auth proxy are served by api-gateway (nginx).
	// Core-manager only handles authenticated API requests.

	// Agent Bootstrap Endpoint (Special Authentication, bypasses AuthMiddleware)
	// This endpoint is for agents to register with a bootstrap key, not a JWT.
	s.router.POST("/api/v1/agents/bootstrap", bootstrapHandler.HandleAgentBootstrap)

	// Site-facts relay (Special Authentication, bypasses AuthMiddleware —
	// same rationale as bootstrap: the caller is a headless box service with
	// a static token, not a user with a JWT). Read-only; serves evidence_base
	// FACTS by domain so box-hosted tools hold live DB truth instead of
	// compiled-in copies. Reached over WireGuard; ClusterIP only, no ingress.
	siteFactsHandler := handlers.NewSiteFactsHandler(personaRepoImpl.ClientsDB(), s.logger)
	s.router.GET("/api/v1/site-facts/:domain", siteFactsHandler.HandleGetSiteFacts)

	// THE CUSTOMER DELIVERY ROUTES (/c/, later /d/) ARE DELIBERATELY NOT HERE.
	//
	// They live on a SECOND listener with its own router (newDeliveryEngine
	// below), so that this port — which serves every site's data, the work-item
	// queue and the pipeline controls — carries no publicly-reachable customer
	// route at all. Before 2026-08-25 they were mounted here, and the only thing
	// keeping the admin API off the internet was an anchored `location` regex in
	// nginx on the webdesign.uk box. Widening that location by one character
	// would have exposed this entire API, and nothing in the binary would have
	// refused. RFC_054 Q2, owner ruling 2026-08-25.
	//
	// This is enforced, not merely intended: assertNoDeliveryRoutes refuses to
	// construct the server if a delivery path is ever mounted on this router.
	// Adding one back here does not produce a quietly-reopened hole; it produces
	// a core-manager that will not start.

	// API v1 group with authentication
	apiV1 := s.router.Group("/api/v1")
	apiV1.Use(middleware.AuthMiddleware(authConfig))
	{
		// Template Management (Admin Only)
		templates := apiV1.Group("/templates")
		templates.Use(middleware.AdminOnly())
		{
			templates.POST("", templateHandler.HandleCreateTemplate)
			templates.GET("", templateHandler.HandleListTemplates)
			templates.GET("/:id", templateHandler.HandleGetTemplate)
			templates.PUT("/:id", templateHandler.HandleUpdateTemplate)
			templates.DELETE("/:id", templateHandler.HandleDeleteTemplate)
		}

		// Persona Instance Management (Tenant-scoped)
		instances := apiV1.Group("/personas/instances")
		instances.Use(middleware.TenantMiddleware(s.logger))
		{
			instances.POST("", instanceHandler.HandleCreateInstance)
			instances.GET("", instanceHandler.HandleListInstances)
			instances.GET("/:id", instanceHandler.HandleGetInstance)
			instances.PATCH("/:id", instanceHandler.HandleUpdateInstance)
			instances.DELETE("/:id", instanceHandler.HandleDeleteInstance)
		}

		// Admin Management (Admin Only)
		adminGroup := apiV1.Group("/admin")
		adminGroup.Use(middleware.AdminOnly())
		{
			// Client Management
			adminGroup.POST("/clients", clientHandlers.HandleCreateClient)
			adminGroup.GET("/clients", clientHandlers.HandleListClients)
			adminGroup.GET("/clients/:client_id/usage", clientHandlers.HandleGetClientUsage)

			// Customer Management (website customers on the clients->networks->sites
			// chain, migration 375 — a different population from /clients above,
			// which serves the per-client-schema tenant machinery)
			adminGroup.POST("/customers", customerHandlers.HandleCreateCustomer)
			adminGroup.GET("/customers", customerHandlers.HandleListCustomers)
			adminGroup.GET("/customers/:customer_id", customerHandlers.HandleGetCustomer)
			adminGroup.PATCH("/customers/:customer_id", customerHandlers.HandleUpdateCustomer)

			// System & Workflow Management
			adminGroup.GET("/system/status", systemHandlers.HandleGetSystemStatus)
			adminGroup.GET("/system/kafka/topics", systemHandlers.HandleListKafkaTopics)
			adminGroup.GET("/workflows", systemHandlers.HandleListWorkflows)
			adminGroup.GET("/workflows/:correlation_id", systemHandlers.HandleGetWorkflow)
			adminGroup.POST("/workflows/:correlation_id/resume", systemHandlers.HandleResumeWorkflow)

			// Agent Definition Management
			adminGroup.GET("/agent-definitions", systemHandlers.HandleListAgentDefinitions)
			adminGroup.POST("/agent-definitions", agentAdminHandlers.HandleCreateAgentDefinition) // NEW
			adminGroup.PUT("/agent-definitions/:type_name", systemHandlers.HandleUpdateAgentDefinition)

			// Kafka Topic Management for Agents (NEW)
			adminGroup.GET("/agent-definitions/:type/topics/verify", agentAdminHandlers.HandleVerifyAgentTopics)
			adminGroup.POST("/agent-definitions/:type/topics/recreate", agentAdminHandlers.HandleRecreateAgentTopics)

			// Agent Instance Management
			adminGroup.GET("/agent-instances", agentAdminHandlers.HandleListAgentInstances)                 // NEW
			adminGroup.GET("/agent-instances/:agent_id", agentAdminHandlers.HandleGetAgentInstance)         // NEW
			adminGroup.PUT("/agent-instances/:agent_id/status", agentAdminHandlers.HandleToggleAgentStatus) // NEW
			adminGroup.POST("/agent-instances/:agent_id/restart", agentAdminHandlers.HandleRestartAgent)    // NEW
			adminGroup.PUT("/clients/:client_id/instances/:instance_id/config", agentAdminHandlers.HandleUpdateInstanceConfig)

			// Topic Management
			adminGroup.POST("/system/cleanup-topics", systemHandlers.HandleCleanupStaleTopics)

			// Site Administration
			siteGroup := adminGroup.Group("/sites")
			{
				siteGroup.GET("", siteAdminHandlers.HandleListSites)
				siteGroup.GET("/:site_id", siteAdminHandlers.HandleGetSite)
				siteGroup.PATCH("/:site_id", siteAdminHandlers.HandleUpdateSite)
				siteGroup.POST("/:site_id/lock", siteAdminHandlers.HandleLockSite)
				siteGroup.POST("/:site_id/unlock", siteAdminHandlers.HandleUnlockSite)
				siteGroup.PATCH("/:site_id/specs/:aspect", siteAdminHandlers.HandleUpdateSiteSpec)

				// Spec Direction Control (Phase 4)
				siteGroup.GET("/:site_id/specs", specAdminHandlers.HandleListSpecs)
				siteGroup.POST("/:site_id/specs/:aspect/pin", specAdminHandlers.HandlePinSpec)
				siteGroup.POST("/:site_id/specs/:aspect/unpin", specAdminHandlers.HandleUnpinSpec)
				siteGroup.POST("/:site_id/specs/:aspect/propagate", specAdminHandlers.HandlePropagateSpec)

				// Page Structure (Phase 2)
				siteGroup.GET("/:site_id/pages", pageAdminHandlers.HandleListPages)
				siteGroup.GET("/:site_id/pages/:page_name/components", pageAdminHandlers.HandleListComponents)
				siteGroup.PATCH("/:site_id/pages/:page_name/components/:component_id", pageAdminHandlers.HandleUpdateComponent)
				siteGroup.POST("/:site_id/pages/:page_name/components/:component_id/lock", pageAdminHandlers.HandleLockComponent)
				siteGroup.POST("/:site_id/pages/:page_name/components/:component_id/unlock", pageAdminHandlers.HandleUnlockComponent)
				siteGroup.DELETE("/:site_id/pages/:page_name/components/:component_id", pageAdminHandlers.HandleRemoveComponent)
				siteGroup.POST("/:site_id/pages/:page_name/restore-section", pageAdminHandlers.HandleRestoreSection)
				siteGroup.POST("/:site_id/pages/:page_name/components/:component_id/regenerate", pageAdminHandlers.HandleRegenerateComponent)
				siteGroup.POST("/:site_id/pages/:page_name/regenerate", pageAdminHandlers.HandleRegeneratePage)
				siteGroup.PATCH("/:site_id/pages/:page_name/spec", pageAdminHandlers.HandleUpdatePageSpec)

				// Site-Wide Components (Phase 7)
				siteGroup.GET("/:site_id/site-components", pageAdminHandlers.HandleListSiteComponents)
				siteGroup.PATCH("/:site_id/site-components/:slot_name", pageAdminHandlers.HandleUpdateSiteComponent)
				siteGroup.POST("/:site_id/site-components/:slot_name/lock", pageAdminHandlers.HandleLockSiteComponent)
				siteGroup.POST("/:site_id/site-components/:slot_name/unlock", pageAdminHandlers.HandleUnlockSiteComponent)

				// Assets (Media)
				siteGroup.GET("/:site_id/assets", assetAdminHandlers.HandleListAssets)
				siteGroup.GET("/:site_id/assets/:asset_id/references", assetAdminHandlers.HandleAssetReferences)
				siteGroup.PATCH("/:site_id/assets/:asset_id", assetAdminHandlers.HandleUpdateAsset)
				siteGroup.DELETE("/:site_id/assets/:asset_id", assetAdminHandlers.HandleDeleteAsset)
			}

			// Work Item Administration + HITL Review
			workItemGroup := adminGroup.Group("/work-items")
			{
				workItemGroup.POST("", siteAdminHandlers.HandleCreateWorkItem)
				workItemGroup.GET("", siteAdminHandlers.HandleListWorkItems)
				workItemGroup.GET("/:item_id", siteAdminHandlers.HandleGetWorkItem)
				workItemGroup.PATCH("/:item_id", siteAdminHandlers.HandleUpdateWorkItem)
				workItemGroup.POST("/:item_id/retry", siteAdminHandlers.HandleRetryWorkItem)
				workItemGroup.POST("/:item_id/resolve", siteAdminHandlers.HandleResolveWorkItem)
				workItemGroup.POST("/:item_id/approve", siteAdminHandlers.HandleApproveWorkItem)
			}

			// Pipeline Administration (scheduled tasks control)
			pipelineGroup := adminGroup.Group("/pipelines")
			{
				pipelineGroup.GET("", pipelineAdminHandlers.HandleListPipelines)
				pipelineGroup.GET("/stats", pipelineAdminHandlers.HandlePipelineStats)
				pipelineGroup.PATCH("/:name", pipelineAdminHandlers.HandleUpdatePipeline)
				pipelineGroup.POST("/:name/trigger", pipelineAdminHandlers.HandleTriggerPipeline)
			}
		}
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Info("Starting Core Manager API server", zap.String("address", s.httpServer.Addr))
	return s.httpServer.ListenAndServe()
}

// StartDelivery starts the delivery listener. It returns http.ErrServerClosed on
// a graceful shutdown, exactly as Start does.
//
// It returns nil immediately when no delivery port is configured, so the caller
// can always launch it and does not have to mirror the opt-in decision. A caller
// that treats a nil return as "the listener stopped" would be wrong either way —
// the same is true of Start.
func (s *Server) StartDelivery() error {
	if s.deliveryServer == nil {
		return nil
	}
	s.logger.Info("Starting Core Manager delivery listener (customer routes only)",
		zap.String("address", s.deliveryServer.Addr))
	return s.deliveryServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.kafkaProducer.Close()
	if s.deliveryServer != nil {
		// Shut the public door first, and do not let its error mask the main
		// one: a customer mid-click is a smaller loss than an admin request
		// killed without the error being reported.
		if err := s.deliveryServer.Shutdown(ctx); err != nil {
			s.logger.Error("Delivery listener shutdown failed", zap.Error(err))
		}
	}
	return s.httpServer.Shutdown(ctx)
}

// DeliveryAddress returns the delivery listener's address, or "" when it is not
// configured. Exists so a caller can log or probe what is actually listening
// rather than assuming the configured value took effect.
func (s *Server) DeliveryAddress() string {
	if s.deliveryServer == nil {
		return ""
	}
	return s.deliveryServer.Addr
}

// Address returns the server's address
func (s *Server) Address() string {
	return s.httpServer.Addr
}
