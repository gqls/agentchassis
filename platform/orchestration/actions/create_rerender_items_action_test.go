package actions

import (
	"testing"

	"go.uber.org/zap"
)

// singlePageFromScalars is the tool-birth path: a producer that made exactly
// one page names it by scalar config paths instead of a pages array.
func TestSinglePageFromScalars(t *testing.T) {
	logger := zap.NewNop()

	// The real shape from tool-generator's create_tool_component result.
	collected := map[string]interface{}{
		"create_result": map[string]interface{}{
			"page_id":  "f25dd4d8-6e25-44eb-a021-689d3057d7a3",
			"function": "tool-loot-table-balancer",
			"page_url": "/tools/tool-loot-table-balancer.html", // leading slash
		},
	}
	config := map[string]interface{}{
		"page_id_field":   "create_result.page_id",
		"page_name_field": "create_result.function",
		"filename_field":  "create_result.page_url",
	}

	page := singlePageFromScalars(collected, config, logger)
	if page == nil {
		t.Fatal("expected a single page, got nil")
	}
	if page["page_id"] != "f25dd4d8-6e25-44eb-a021-689d3057d7a3" {
		t.Errorf("page_id wrong: %v", page["page_id"])
	}
	if page["name"] != "tool-loot-table-balancer" {
		t.Errorf("name wrong: %v", page["name"])
	}
	// The leading slash must be stripped — the item spec wants "tools/x.html",
	// create_result carries "/tools/x.html".
	if page["filename"] != "tools/tool-loot-table-balancer.html" {
		t.Errorf("filename should have its leading slash trimmed, got %v", page["filename"])
	}
}

// Not configured for single-page mode → nil, so list mode is untouched.
func TestSinglePageFromScalars_NotConfigured(t *testing.T) {
	if singlePageFromScalars(map[string]interface{}{}, map[string]interface{}{}, zap.NewNop()) != nil {
		t.Error("no scalar fields configured → must return nil (list mode)")
	}
}

// Configured but the producer wrote nothing → nil (and a warning), never a
// silently-empty page.
func TestSinglePageFromScalars_ConfiguredButEmpty(t *testing.T) {
	config := map[string]interface{}{
		"page_id_field":   "create_result.page_id",
		"page_name_field": "create_result.function",
	}
	if singlePageFromScalars(map[string]interface{}{}, config, zap.NewNop()) != nil {
		t.Error("configured but nothing resolved → must return nil, not a partial page")
	}
}

// filename is optional: a page with no filename field still resolves.
func TestSinglePageFromScalars_NoFilename(t *testing.T) {
	collected := map[string]interface{}{
		"r": map[string]interface{}{"pid": "abc", "pn": "tool-x"},
	}
	config := map[string]interface{}{
		"page_id_field":   "r.pid",
		"page_name_field": "r.pn",
	}
	page := singlePageFromScalars(collected, config, zap.NewNop())
	if page == nil || page["filename"] != "" {
		t.Errorf("filename optional; expected empty, got %v", page)
	}
}
