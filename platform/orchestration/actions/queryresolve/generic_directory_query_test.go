package queryresolve

import (
	"strings"
	"testing"
)

// The generic register arms (`directory:<kind>` / `directory_full:<kind>`) exist so a new
// directory kind is a component declaration, not a Go change. These tests pin the three
// properties that make that true, each with a control that fails if the property is lost.

func TestGenericDirectoryArmsResolveAnyKind(t *testing.T) {
	// Every kind the register carries today, plus one that does not exist yet: the point of
	// the arm is that it does NOT know the kind list, so a future kind must be accepted too.
	for _, name := range []string{
		"directory:model", "directory:company", "directory:protocol",
		"directory:mortgage-lender", "directory:savings-provider", "directory:health-insurer",
		"directory:copywriter", "directory:a-kind-invented-tomorrow",
		"directory_full:copywriter", "directory_full:model",
	} {
		if !IsKnownQueryName(name) {
			t.Errorf("IsKnownQueryName(%q) = false, want true — a component declaring this source would be refused at plan time", name)
		}
	}
	// CONTROL: the arm must not have made every name resolvable.
	for _, name := range []string{"directoryy:model", "not_a_directory:model", "directory_partial:model"} {
		if IsKnownQueryName(name) {
			t.Errorf("IsKnownQueryName(%q) = true, want false — the base lookup has stopped discriminating", name)
		}
	}
}

func TestLiteralPerKindArmsStillResolve(t *testing.T) {
	// Live components on shipped pages declare these literal names. The generic arm is an
	// ADDITION; if a literal stopped resolving, its listing would empty with no error.
	for _, name := range []string{
		"model_directory", "model_directory_full",
		"adoption_tracker", "adoption_tracker_full",
		"protocol_tracker", "protocol_tracker_full",
		"mortgage_lender_directory", "mortgage_lender_directory_full",
		"savings_provider_directory", "savings_provider_directory_full",
		"health_insurer_directory", "health_insurer_directory_full",
		"business_directory",
	} {
		if !IsKnownQueryName(name) {
			t.Errorf("IsKnownQueryName(%q) = false — a live component's source has stopped resolving", name)
		}
	}
}

func TestGenericDirectoryArmsDeclareTheirDependency(t *testing.T) {
	// bugs_open/384: a source whose data changed must be able to tell its consumer pages to
	// re-resolve. A generic arm with no dependency entry would go silently stale.
	for _, name := range []string{
		"directory:copywriter", "directory_full:copywriter",
		"directory:model", "directory_full:a-kind-invented-tomorrow",
	} {
		if !SourceReads(name, DepDirectoryEntities) {
			t.Errorf("SourceReads(%q, DepDirectoryEntities) = false — pages fed by this arm would never be re-resolved", name)
		}
	}
	// CONTROL: the dependency is specific, not universal.
	if SourceReads("directory:copywriter", DepPageCardImages) {
		t.Error("SourceReads(directory:copywriter, DepPageCardImages) = true — the deps entry is over-broad")
	}
	if SourceReads("pages_where_type:tool", DepDirectoryEntities) {
		t.Error("SourceReads(pages_where_type:tool, DepDirectoryEntities) = true — an unrelated source now claims the register dependency")
	}
}

func TestGenericArmIsDiscoverableInTheVocabulary(t *testing.T) {
	// KnownQueryBases feeds validation messages that must name the real vocabulary; a base a
	// planner cannot see is a base nobody will declare.
	bases := strings.Join(KnownQueryBases(), ",")
	for _, want := range []string{"directory", "directory_full"} {
		if !strings.Contains(","+bases+",", ","+want+",") {
			t.Errorf("KnownQueryBases() does not list %q — planners and validators cannot discover it (got %s)", want, bases)
		}
	}
}
