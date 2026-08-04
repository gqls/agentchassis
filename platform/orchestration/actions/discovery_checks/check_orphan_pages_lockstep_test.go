// FILE: platform/orchestration/actions/discovery_checks/check_orphan_pages_lockstep_test.go
//
// The orphan census and the retraction graph audit read the SAME three link
// surfaces to answer opposite questions (what IS unreachable / what BECOMES
// unreachable). The surface list is declared once — datahelpers.
// InboundLinkSurfaces — and this test is the discovery_checks half of the
// lockstep: a surface added to the shared list fails here until this census
// answers for it; a surface dropped from this query fails here immediately.
// (Council round 5a965452, 098 debt 3: the lockstep used to be a comment
// saying "CHANGE ONE, READ THE OTHER", and a comment is not a mechanism.)
//
// A green run here is a DRIFT ALARM only, not a semantic proof: it asserts the
// surface's NAME appears in the SQL, not that the query reads it correctly
// (council `architecture` seat, round 37593214 — do not over-trust it).
package discovery_checks

import (
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func TestOrphanCensusReadsEveryDeclaredLinkSurface(t *testing.T) {
	if len(datahelpers.InboundLinkSurfaces) == 0 {
		t.Fatal("datahelpers.InboundLinkSurfaces is empty — the lockstep contract has been deleted, not satisfied")
	}
	for _, surface := range datahelpers.InboundLinkSurfaces {
		if !strings.Contains(findOrphanPagesSQL, surface) {
			t.Errorf("the orphan census does not read %q — a page linked ONLY from that surface would be reported as an orphan. "+
				"Either the census gained a blind spot, or a new link surface was added to datahelpers.InboundLinkSurfaces "+
				"and this query has not answered for it yet", surface)
		}
	}
}
