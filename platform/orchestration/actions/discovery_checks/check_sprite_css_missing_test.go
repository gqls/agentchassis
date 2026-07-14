package discovery_checks

import (
	"strings"
	"testing"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/imageryplan"
)

// The live grid on robot-hands (3×3, verified at the B11 eyeball gate).
var testCells = []string{"check", "gauge", "gripper", "cog", "chart", "download", "arrow", "info", "warning"}

func stampedRow(sig, emittedAt string) spriteSheetPlanRow {
	return stampedRowFormat(sig, emittedAt, imageryplan.SpriteCSSFormat)
}

func stampedRowFormat(sig, emittedAt string, format int) spriteSheetPlanRow {
	r := spriteSheetPlanRow{Key: "sprite_sheet_main", Rows: 3, Cols: 3, CellNames: testCells, Verified: true}
	r.SpritesCSS = &struct {
		EmittedAt string `json:"emitted_at"`
		SheetPath string `json:"sheet_path"`
		Signature string `json:"signature"`
		Format    int    `json:"format"`
	}{EmittedAt: emittedAt, SheetPath: "/assets/images/sprite-sheet-main.jpg", Signature: sig, Format: format}
	return r
}

func TestSpriteCSSStaleness(t *testing.T) {
	sig := imageryplan.SpriteGridSignature(3, 3, testCells)
	emitted := "2026-07-14T13:44:00Z"
	emittedTime, _ := time.Parse(time.RFC3339, emitted)

	t.Run("never emitted → missing", func(t *testing.T) {
		row := spriteSheetPlanRow{Key: "sprite_sheet_main", Rows: 3, Cols: 3, CellNames: testCells, Verified: true}
		reason, stale := spriteCSSStaleness(row, time.Time{})
		if !stale || reason != "missing" {
			t.Errorf("got (%q, %v), want (\"missing\", true)", reason, stale)
		}
	})

	t.Run("emitted from the current grid, sheet unchanged → fulfilled (no re-emit)", func(t *testing.T) {
		// The idempotence case: without this, every discovery pass would queue
		// another needs_sprite_css and re-commit an identical stylesheet forever.
		row := stampedRow(sig, emitted)
		if reason, stale := spriteCSSStaleness(row, emittedTime.Add(-time.Hour)); stale {
			t.Errorf("expected fulfilled, got stale: %q", reason)
		}
	})

	t.Run("cell names re-verified → stale (CSS now slices the wrong glyphs)", func(t *testing.T) {
		reordered := []string{"gauge", "check", "gripper", "cog", "chart", "download", "arrow", "info", "warning"}
		row := stampedRow(imageryplan.SpriteGridSignature(3, 3, reordered), emitted)
		// Row still describes the ORIGINAL vocabulary, so the stamp disagrees.
		reason, stale := spriteCSSStaleness(row, time.Time{})
		if !stale {
			t.Fatalf("expected stale on grid change, got fulfilled")
		}
		if reason == "missing" {
			t.Errorf("wrong reason %q — should report a grid/cell change", reason)
		}
	})

	t.Run("grid geometry changed (3x3 → 4x4) → stale", func(t *testing.T) {
		row := stampedRow(imageryplan.SpriteGridSignature(4, 4, testCells), emitted)
		if _, stale := spriteCSSStaleness(row, time.Time{}); !stale {
			t.Errorf("expected stale when the stamped geometry differs from the plan")
		}
	})

	t.Run("sheet regenerated after the CSS was emitted → stale", func(t *testing.T) {
		row := stampedRow(sig, emitted)
		if _, stale := spriteCSSStaleness(row, emittedTime.Add(time.Minute)); !stale {
			t.Errorf("expected stale when the asset is newer than the emit stamp")
		}
	})

	t.Run("unparseable emit timestamp → re-emit rather than trust it", func(t *testing.T) {
		row := stampedRow(sig, "not-a-timestamp")
		if _, stale := spriteCSSStaleness(row, time.Time{}); !stale {
			t.Errorf("expected stale on an unparseable emitted_at")
		}
	})

	// The sheet never moves when the EMITTER changes (I2.5 added the
	// .sprite-bullets container opt-in to an unchanged 3×3 grid). Without a
	// format version, the signature would still match and every site would serve
	// the pre-I2.5 stylesheet forever.
	t.Run("emitter format bumped → stale even though the sheet is unchanged", func(t *testing.T) {
		row := stampedRowFormat(sig, emitted, imageryplan.SpriteCSSFormat-1)
		reason, stale := spriteCSSStaleness(row, emittedTime.Add(-time.Hour))
		if !stale {
			t.Fatalf("expected stale on an older CSS format version")
		}
		if !strings.Contains(reason, "format") {
			t.Errorf("reason %q should name the format version", reason)
		}
	})
}
