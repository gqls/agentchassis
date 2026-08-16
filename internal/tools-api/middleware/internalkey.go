package middleware

import (
	"crypto/subtle"

	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/internal/tools-api/httperr"
)

// InternalKeyHeader is the shared-secret header the cluster's pull actions send
// (platform/orchestration/actions/ndjson_feed.go sets it; intent_collector and
// pull_report_requests both ride that transport).
const InternalKeyHeader = "X-Internal-Key"

// InternalKey gates a route on X-Internal-Key matching key, in constant time.
// It is the auth for the ONE route in this service that a browser never calls
// and the cluster does — GET /api/v1/tools/gripper/requests. That route must
// NOT sit behind CORSMiddleware (no Origin header on a server-to-server GET, so
// CORS would 403 it), which is why the gripper handlers are split across two
// gin groups in api/server.go: browser routes under CORS, this one under here.
//
// An empty configured key refuses everything: a missing secret must read as
// "locked", never as "open".
func InternalKey(key string) gin.HandlerFunc {
	want := []byte(key)
	return func(c *gin.Context) {
		got := []byte(c.GetHeader(InternalKeyHeader))
		if len(want) == 0 || len(got) != len(want) || subtle.ConstantTimeCompare(got, want) != 1 {
			httperr.JSONError(c, 401, "unauthorised")
			c.Abort()
			return
		}
		c.Next()
	}
}
