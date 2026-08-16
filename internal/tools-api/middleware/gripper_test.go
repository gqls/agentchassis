package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/platform/httpguard"
)

func TestInternalKeyRefusesEmptyKeyAndMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, configured := range []string{"", "0123456789abcdef0123456789abcdef"} {
		r := gin.New()
		r.GET("/x", InternalKey(configured), func(c *gin.Context) { c.String(200, "ok") })
		for _, sent := range []string{"", "wrong", "0123456789abcdef0123456789abcdef"} {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/x", nil)
			if sent != "" {
				req.Header.Set(InternalKeyHeader, sent)
			}
			r.ServeHTTP(rec, req)
			want := 401
			if configured != "" && sent == configured {
				want = 200
			}
			if rec.Code != want {
				t.Errorf("configured=%q sent=%q: code %d want %d", configured, sent, rec.Code, want)
			}
		}
	}
}

func TestBandedRateLimitRefusesWithRetryAfter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/x", BandedRateLimit(httpguard.NewLimiter(httpguard.Band{Window: time.Hour, Max: 2})), func(c *gin.Context) { c.String(200, "ok") })
	codes := []int{}
	var retry string
	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/x", nil)
		req.RemoteAddr = "203.0.113.9:1234"
		r.ServeHTTP(rec, req)
		codes = append(codes, rec.Code)
		retry = rec.Header().Get("Retry-After")
	}
	if codes[0] != 200 || codes[1] != 200 || codes[2] != 429 {
		t.Fatalf("codes = %v", codes)
	}
	if retry == "" {
		t.Fatalf("429 without Retry-After")
	}
}
