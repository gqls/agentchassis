package actions

import (
	"context"
	"strings"
	"testing"

	"go.uber.org/zap"
)

// The 056 shape made unrepresentable (council 6acf8e4e round 1, gating HIGH):
// a regeneration must refuse empty bytes BEFORE any read, lock, or write —
// whatever upstream produced them — so a refused/failed stage can never blank
// a live placement's rendered_html. Callable with a zero ActionParams because
// the check precedes every use of params.
func TestRegenerateToolRefusesEmptyBytes(t *testing.T) {
	for _, req := range []toolRegenerateRequest{
		{function: "my-tool", htmlContent: "", renderedHTML: "<div>x</div>"},
		{function: "my-tool", htmlContent: "<div>x</div>", renderedHTML: ""},
		{function: "my-tool", htmlContent: "   \n", renderedHTML: "  "},
	} {
		_, err := regenerateToolComponentInPlace(context.Background(), ActionParams{}, zap.NewNop(), req)
		if err == nil || !strings.Contains(err.Error(), "bugs_closed/056") {
			t.Fatalf("empty bytes must refuse naming the 056 class, got: %v", err)
		}
	}
}
