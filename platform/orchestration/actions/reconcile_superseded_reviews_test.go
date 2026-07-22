package actions

import "testing"

// The context shape is the real CONTENT_VALIDATION_BLOCKER_DETAIL payload
// observed on fundamentallyai.com (bugs_open/056 regeneration-loss, diagnosis
// corr b361298a): issues[] carrying type/value/severity plus page-level fields.
const realBlockerContext = `{
	"issues": [{
		"type": "cross_site_domain",
		"value": "leopardessconsulting.co.uk",
		"category": "contamination",
		"location": "atform generated on leopardessconsulting.co.uk. We corrected it.",
		"severity": "blocker",
		"description": "Found domain 'leopardessconsulting.co.uk' in content for 'fundamentallyai.com'"
	}],
	"page_name": "model-fine-tuning",
	"error_count": 0,
	"blocker_count": 1
}`

func TestFlaggedValueFindings(t *testing.T) {
	t.Run("flagged value absent from new content = dropped, not resolved", func(t *testing.T) {
		findings := flaggedValueFindings([]byte(realBlockerContext),
			"<section><p>Fine-tuning is a means to an end.</p></section>")
		if len(findings) != 1 {
			t.Fatalf("want 1 finding, got %d", len(findings))
		}
		f := findings[0]
		if f["value"] != "leopardessconsulting.co.uk" || f["type"] != "cross_site_domain" {
			t.Fatalf("finding mangled: %+v", f)
		}
		if f["present_in_new_content"].(bool) {
			t.Fatal("value is absent from the content but reported present")
		}
	})

	t.Run("flagged value still present is reported present, case-insensitively", func(t *testing.T) {
		findings := flaggedValueFindings([]byte(realBlockerContext),
			"<p>Our own site, LeopardessConsulting.co.uk, once shipped a wrong claim.</p>")
		if len(findings) != 1 {
			t.Fatalf("want 1 finding, got %d", len(findings))
		}
		if !findings[0]["present_in_new_content"].(bool) {
			t.Fatal("value is present in the content but reported absent")
		}
	})

	t.Run("no blocker context yields no findings, not a failure", func(t *testing.T) {
		if got := flaggedValueFindings(nil, "<p>x</p>"); len(got) != 0 {
			t.Fatalf("nil context: want 0 findings, got %d", len(got))
		}
		if got := flaggedValueFindings([]byte("not json"), "<p>x</p>"); len(got) != 0 {
			t.Fatalf("malformed context: want 0 findings, got %d", len(got))
		}
	})

	t.Run("issues with empty values are skipped", func(t *testing.T) {
		ctx := `{"issues":[{"type":"x","value":"  ","severity":"blocker"},{"type":"y","value":"real","severity":"error"}]}`
		findings := flaggedValueFindings([]byte(ctx), "real content")
		if len(findings) != 1 {
			t.Fatalf("want 1 finding (empty value skipped), got %d", len(findings))
		}
		if findings[0]["type"] != "y" {
			t.Fatalf("wrong finding survived: %+v", findings[0])
		}
	})
}
