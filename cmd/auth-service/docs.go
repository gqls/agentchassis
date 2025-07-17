// Package main AI Persona Platform API
//
// The AI Persona Platform provides APIs for managing AI personas, authentication, and content generation.
//
// Authentication:
// All API endpoints (except auth endpoints) require a valid JWT token in the Authorization header:
// Authorization: Bearer <token>
//
// Rate Limiting:
// API calls are rate-limited based on your subscription tier:
// - Free: 100 requests/hour
// - Basic: 1000 requests/hour
// - Premium: 10000 requests/hour
// - Enterprise: Unlimited
//
//	Schemes: https, http
//	Host: api.personaplatform.com
//	BasePath: /
//	Version: 1.0.0
//	Contact: AI Persona Support<support@personaplatform.com>
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Security:
//	- BearerAuth: []
//
//	SecurityDefinitions:
//	BearerAuth:
//	  type: apiKey
//	  name: Authorization
//	  in: header
//	  description: "Type 'Bearer' followed by a space and the JWT token"
//
// swagger:meta
package main

import (
	_ "github.com/gqls/agentchassis/internal/auth-service/auth"
	_ "github.com/gqls/agentchassis/internal/auth-service/project"
	_ "github.com/gqls/agentchassis/internal/auth-service/subscription"
	_ "github.com/gqls/agentchassis/internal/auth-service/user"
)
