package actions

import "testing"

func TestCheckMetaCommentary(t *testing.T) {
	// The prose actually persisted as page content on robot-hands'
	// product-card-with-cta (found 2026-07-14).
	storedApology := `<div class="product-card"><p>The data schema for this section requires ` +
		`product array data sourced from query.affiliate_products. Per the schema definition, ` +
		`this field is marked required: true with on_missing: skip_section — meaning without ` +
		`actual affiliate product data (names, prices, ratings…)</p></div>`

	issues := checkMetaCommentary(storedApology)
	if len(issues) == 0 {
		t.Fatal("expected the stored robot-hands apology to be flagged, got no issues")
	}
	for _, issue := range issues {
		if issue.Severity != "blocker" {
			t.Errorf("meta-commentary must block deployment, got severity %q", issue.Severity)
		}
	}

	refusal := `<section><p>I don't have access to the product catalog, so as an AI I cannot generate this listing.</p></section>`
	if got := checkMetaCommentary(refusal); len(got) == 0 {
		t.Error("expected first-person refusal prose to be flagged")
	}

	legit := `<section><h2>Gripper Selection</h2><p>Choosing the right gripper requires ` +
		`matching payload, stroke and IP rating to your application. Our MatchMatrix tool ` +
		`scores 200+ models against your constraints.</p></section>`
	if got := checkMetaCommentary(legit); len(got) != 0 {
		t.Errorf("legitimate content flagged as meta-commentary: %+v", got)
	}
}
