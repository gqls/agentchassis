// FILE: platform/orchestration/actions/validate_page_content_contamination_test.go
//
// Contract for the cross-site contamination check's per-site allowlist
// (bugs_open/055). A portfolio / meta site such as fundamentallyai.com
// legitimately names another of our sites (the owner-approved self-correction
// case study naming leopardessconsulting.co.uk). Without an allowlist the
// mention is a blocker; with the domain on the site's allowlist it must not be
// flagged — and the allowlist is specific, so an *un*-listed known site is
// still caught. Absent allowlist (nil) reproduces the original behaviour.

package actions

import "testing"

// A page for fundamentallyai.com that tells the self-correction story (names
// leopardessconsulting.co.uk on purpose) and, to prove the allowlist is
// specific, also happens to contain an un-listed known domain.
const contaminationTestHTML = `
<section>
  <h2>How we caught our own mistake</h2>
  <p>Our sibling site leopardessconsulting.co.uk once published invented
     details. Our verification system flagged it; we corrected it. That is
     Leopardess Consulting, named openly on purpose.</p>
  <p>Unrelated stray mention of dartsonline.com that nobody approved.</p>
</section>`

func TestContamination_NoAllowlist_FlagsKnownDomain(t *testing.T) {
	issues := checkDomainContamination(contaminationTestHTML,
		"fundamentallyai.com", "FundamentallyAI", nil)

	var sawLeopardess bool
	for _, is := range issues {
		if is.Value == "leopardessconsulting.co.uk" {
			sawLeopardess = true
			if is.Severity != "blocker" || is.Category != "contamination" {
				t.Errorf("expected leopardess flagged as contamination blocker, got %+v", is)
			}
		}
	}
	if !sawLeopardess {
		t.Fatalf("without an allowlist, leopardessconsulting.co.uk must be flagged; issues=%+v", issues)
	}
}

func TestContamination_Allowlist_SuppressesListedDomainAndCompany(t *testing.T) {
	allow := map[string]bool{"leopardessconsulting.co.uk": true}

	issues := checkDomainContamination(contaminationTestHTML,
		"fundamentallyai.com", "FundamentallyAI", allow)

	for _, is := range issues {
		if is.Value == "leopardessconsulting.co.uk" {
			t.Errorf("allowlisted domain must not be flagged, got %+v", is)
		}
		if is.Value == "Leopardess Consulting" {
			t.Errorf("company of an allowlisted domain must not be flagged, got %+v", is)
		}
	}

	// The allowlist is specific: an un-listed known domain is still caught.
	var sawDarts bool
	for _, is := range issues {
		if is.Value == "dartsonline.com" && is.Severity == "blocker" {
			sawDarts = true
		}
	}
	if !sawDarts {
		t.Errorf("an un-allowlisted known domain must still be flagged; issues=%+v", issues)
	}
}
