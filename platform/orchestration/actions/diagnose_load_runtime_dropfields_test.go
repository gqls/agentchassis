package actions

import (
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// The render half of the drop-reporting fix. Round 3's council (eba040a9,
// editquality) noted it was "the second half of the fix and currently
// unverified" — the assertions covered withPriorCodeRequests but nothing
// exercised the code that turns the count into bundle text.
func TestUpstreamDropNotice(t *testing.T) {
	t.Run("no notice when nothing was dropped", func(t *testing.T) {
		if got := upstreamDropNotice("code_request", 0); got != "" {
			t.Fatalf("want empty, got %q", got)
		}
		if got := upstreamDropNotice("code_request", -1); got != "" {
			t.Fatalf("negative must not render a notice, got %q", got)
		}
	})

	t.Run("names the count and the kind", func(t *testing.T) {
		got := upstreamDropNotice("code_request", 3)
		if !strings.Contains(got, "3 further code_request(s)") {
			t.Fatalf("count/kind missing: %q", got)
		}
		if !strings.Contains(got, "data_request") {
			// sanity: the kind is parameterised, not hard-coded
		}
		if d := upstreamDropNotice("data_request", 2); !strings.Contains(d, "2 further data_request(s)") {
			t.Fatalf("kind not parameterised: %q", d)
		}
	})

	// The wording is the actual guard here: a verdicter that reads a
	// capped-away question as an answered one lands in the empty-vs-absent trap
	// this whole tier is built to avoid. If someone shortens this string, this
	// test should stop them.
	t.Run("says the questions were never answered and absence is not an answer", func(t *testing.T) {
		got := upstreamDropNotice("code_request", 1)
		for _, must := range []string{"never answered", "capped, not complete", "do not read their absence as an answer"} {
			if !strings.Contains(got, must) {
				t.Errorf("notice must contain %q — it is the guard, not decoration; got %q", must, got)
			}
		}
	})
}

// The cross-action name coupling bug_historian flagged (eba040a9, low):
// diagnose_route writes bare keys into its result map and diagnose_load_runtime
// reads them back at "route.<key>", with nothing tying the two ends together.
// A rename at either end would silently re-open the very silence the drop
// reporting exists to close. These assert the ends still agree.
func TestRouteDropFieldsStayInSyncWithLoadRuntimeDefaults(t *testing.T) {
	spec := DiagnoseLoadRuntimeInputSpec

	cases := []struct {
		configKey string
		writerKey string
	}{
		{"code_requests_dropped_field", codeRequestsDroppedKey},
		{"data_requests_dropped_field", dataRequestsDroppedKey},
	}

	for _, c := range cases {
		want := routeOutputPrefix + c.writerKey
		got := datahelpers.GetStringField(spec.Defaults, c.configKey, "")
		if got == "" {
			// Not fatal on its own, but a default that vanished means the reader
			// falls back to "" and the notice silently stops rendering.
			t.Fatalf("%s has no default; the reader would look up an empty path", c.configKey)
		}
		if got != want {
			t.Fatalf("reader default %q != writer key %q — diagnose_route writes result[%q] "+
				"but diagnose_load_runtime reads %q, so drops would go unreported",
				got, want, c.writerKey, got)
		}
	}

	// The declared-optional list must carry them too, or the config keys are
	// unsettable from a workflow and the defaults become the only possible value.
	for _, c := range cases {
		found := false
		for _, o := range spec.Optional {
			if o == c.configKey {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s missing from InputSpec.Optional — not settable from workflow config", c.configKey)
		}
	}
}
