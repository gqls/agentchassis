package actions

import (
	"context"
	"testing"

	"github.com/gqls/agentchassis/platform/voicestyle"
)

// The call-site semantics the CQ-033 submission asserts (council round
// b5a642b7, editquality: "verified by a test, not only asserted in prose").
func TestInjectPlatformBlocks(t *testing.T) {
	ctx := context.Background()
	blocks := map[string]string{
		voicestyle.ConfigName:              "HOUSE VOICE TEXT",
		voicestyle.BuildStandardConfigName: "BUILD STANDARD TEXT",
	}
	get := func(_ context.Context, name string) (string, bool) {
		b, ok := blocks[name]
		return b, ok && b != ""
	}

	t.Run("both injected when absent", func(t *testing.T) {
		td := map[string]interface{}{"other": 1}
		injectPlatformBlocks(ctx, td, get)
		if td["voice_style"] != "HOUSE VOICE TEXT" || td["build_standard"] != "BUILD STANDARD TEXT" {
			t.Fatalf("injection: %v", td)
		}
	})

	t.Run("step-supplied value is never clobbered", func(t *testing.T) {
		td := map[string]interface{}{"build_standard": "STEP OVERRIDE", "voice_style": ""}
		injectPlatformBlocks(ctx, td, get)
		if td["build_standard"] != "STEP OVERRIDE" {
			t.Fatalf("step-supplied build_standard clobbered: %v", td["build_standard"])
		}
		if td["voice_style"] != "" {
			t.Fatalf("step-supplied empty voice_style clobbered: %v — presence, not truthiness, is the override key", td["voice_style"])
		}
	})

	t.Run("missing row degrades to no key, not a failure", func(t *testing.T) {
		td := map[string]interface{}{}
		injectPlatformBlocks(ctx, td, func(context.Context, string) (string, bool) { return "", false })
		if _, has := td["build_standard"]; has {
			t.Fatalf("missing row still injected: %v", td)
		}
		if _, has := td["voice_style"]; has {
			t.Fatalf("missing row still injected voice: %v", td)
		}
	})

	// Control for the vacuousness trap: the map driving injection must contain
	// exactly the two reviewed blocks — a third entry added without review
	// makes this fail and points at the council convention.
	t.Run("reviewed block set is exactly two", func(t *testing.T) {
		if len(platformPromptBlocks) != 2 {
			t.Fatalf("platformPromptBlocks has %d entries — a new platform block needs its own reviewed carrier migration and this test updated in the same commit", len(platformPromptBlocks))
		}
	})
}
