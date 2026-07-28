package actions

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The fixture is a REAL harvested entry's binding schema, read off disk — the
// 2026-07-27 lesson: a fixture invented to satisfy the code under test proves
// only that the code is self-consistent.
func harvestedSchemaAndUse(t *testing.T, file string) (map[string]interface{}, map[string]bool) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "agent_docs",
		"docs024_key_docs_latest", "experience_register", "harvest", "entries", file))
	if err != nil {
		t.Fatalf("cannot read harvested entry %s: %v", file, err)
	}
	entry := mustJSON(t, string(raw))
	schema, _ := entry["binding_schema"].(map[string]interface{})
	if len(schema) == 0 {
		t.Fatalf("%s has no binding_schema", file)
	}
	used := map[string]bool{}
	for k, v := range entry {
		if k == "binding_schema" {
			continue
		}
		collectExperiencePlaceholders(v, used)
	}
	if len(used) == 0 {
		t.Fatalf("%s uses no placeholders — fixture is not exercising closure", file)
	}
	return schema, used
}

// A binding set that closes: every placeholder the entry uses, and nothing else.
func closingBindings(used map[string]bool, schema map[string]interface{}) map[string]interface{} {
	b := map[string]interface{}{}
	for name := range used {
		switch experienceBindingKind(schema, name) {
		case "page":
			b[name] = "/archive"
		default:
			b[name] = ".some-anchor"
		}
	}
	return b
}

func TestBindSiteExperience_AcceptsAClosingBindingSet(t *testing.T) {
	schema, used := harvestedSchemaAndUse(t, "CC-001_feed-driven-teaser-list.json")
	if got := validateExperienceBindings(used, schema, closingBindings(used, schema)); len(got) != 0 {
		t.Fatalf("a closing binding set was refused: %v", got)
	}
}

func TestBindSiteExperience_RefusesAnUnboundPlaceholder(t *testing.T) {
	// The check would run against the literal placeholder text: it runs, it
	// looks green, it asserts nothing. Same family as the -EDIT ids.
	schema, used := harvestedSchemaAndUse(t, "CC-001_feed-driven-teaser-list.json")
	b := closingBindings(used, schema)
	var dropped string
	for k := range b {
		dropped = k
		delete(b, k)
		break
	}
	got := strings.Join(validateExperienceBindings(used, schema, b), " | ")
	if !strings.Contains(got, "{{binding."+dropped+"}}") || !strings.Contains(got, "not bound") {
		t.Fatalf("dropping %q was accepted: %s", dropped, got)
	}
}

func TestBindSiteExperience_RefusesAnEmptyValue(t *testing.T) {
	// The one that matters most: an empty selector SATISFIES closure, so it
	// passes every check that only asks "is it bound?", and then produces a
	// check that cannot fail.
	schema, used := harvestedSchemaAndUse(t, "CC-001_feed-driven-teaser-list.json")
	b := closingBindings(used, schema)
	for k := range b {
		b[k] = "   "
		break
	}
	if got := strings.Join(validateExperienceBindings(used, schema, b), " | "); !strings.Contains(got, "cannot fail") {
		t.Fatalf("an empty binding was accepted: %s", got)
	}
}

func TestBindSiteExperience_RefusesAnUnanchoredSelector(t *testing.T) {
	// Tier 2 SKIPS a selector it cannot anchor, and a skipped check reads green.
	schema, used := harvestedSchemaAndUse(t, "CC-001_feed-driven-teaser-list.json")
	b := closingBindings(used, schema)
	var target string
	for k := range b {
		if experienceBindingKind(schema, k) == "selector" {
			target = k
			break
		}
	}
	if target == "" {
		t.Skip("this entry declares no selector binding")
	}
	b[target] = "[data-role=list]" // an attribute selector has no leftmost anchor token
	got := strings.Join(validateExperienceBindings(used, schema, b), " | ")
	if !strings.Contains(got, "no leftmost") {
		t.Fatalf("an unanchorable selector was accepted: %s", got)
	}
}

func TestBindSiteExperience_RefusesAnUnresolvedPlaceholderInAValue(t *testing.T) {
	// A fork is where placeholders END. One surviving here means the generating
	// step handed its own template through unrendered.
	schema, used := harvestedSchemaAndUse(t, "CC-001_feed-driven-teaser-list.json")
	b := closingBindings(used, schema)
	for k := range b {
		b[k] = "{{binding.something_else}}"
		break
	}
	if got := strings.Join(validateExperienceBindings(used, schema, b), " | "); !strings.Contains(got, "placeholders END") {
		t.Fatalf("a placeholder value was accepted: %s", got)
	}
}

func TestBindSiteExperience_RefusesABindingNothingReads(t *testing.T) {
	schema, used := harvestedSchemaAndUse(t, "CC-001_feed-driven-teaser-list.json")
	b := closingBindings(used, schema)
	b["a_binding_the_entry_never_reads"] = ".x"
	if got := strings.Join(validateExperienceBindings(used, schema, b), " | "); !strings.Contains(got, "never reads it") {
		t.Fatalf("an unread binding was accepted: %s", got)
	}
}

func TestBindSiteExperience_ReportsEveryProblemAtOnce(t *testing.T) {
	schema, used := harvestedSchemaAndUse(t, "CC-001_feed-driven-teaser-list.json")
	b := closingBindings(used, schema)
	n := 0
	for k := range b {
		b[k] = "" // every value empty
		n++
	}
	got := validateExperienceBindings(used, schema, b)
	if len(got) < n {
		t.Fatalf("expected at least %d problems reported together, got %d: %v", n, len(got), got)
	}
}

// TestExperienceAnchorRE_MatchesTheCheckersOwn holds the bind-time anchor rule
// to the checker that actually enforces it at run time. If Tier 2's idea of an
// anchor changes and this copy does not, we would accept selectors the checker
// then SKIPS — and a skipped check reads as green, which is the failure this
// whole register is built to stop.
func TestExperienceAnchorRE_MatchesTheCheckersOwn(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("discovery_checks", "check_tool_acceptance.go"))
	if err != nil {
		t.Fatalf("cannot read the checker: %v", err)
	}
	m := regexp.MustCompile("anchorRe = regexp\\.MustCompile\\(`([^`]+)`\\)").FindSubmatch(src)
	if m == nil {
		t.Fatal("could not find anchorRe in check_tool_acceptance.go — if it moved or was renamed, fix this test rather than deleting it: it is the only thing tying bind-time acceptance to run-time skipping")
	}
	if theirs, ours := string(m[1]), experienceAnchorRE.String(); theirs != ours {
		t.Errorf("anchor rules have drifted — bind-time would accept selectors the checker SKIPS:\n  checker: %s\n  ours:    %s", theirs, ours)
	}
}

// TestExperienceBindingKinds_CoverTheHarvestedCorpus holds the binding-type
// vocabulary to the entries that actually use it.
//
// The list was wrong the first time — it had "path" and "role", which no entry
// declares, and lacked "asset_path" and "url_param", which they do. That is the
// same mistake as the invented `contract` shape and the enumerated document
// list: I wrote a plausible vocabulary instead of counting the real one. This
// test means the corpus decides, not me.
func TestExperienceBindingKinds_CoverTheHarvestedCorpus(t *testing.T) {
	dir := filepath.Join("..", "..", "..", "docs", "agent_docs", "docs024_key_docs_latest",
		"experience_register", "harvest", "entries")
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("cannot read harvested entries: %v", err)
	}
	seen := map[string]string{}
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", f.Name(), err)
		}
		entry := mustJSON(t, string(raw))
		schema, _ := entry["binding_schema"].(map[string]interface{})
		for name, decl := range schema {
			d, ok := decl.(map[string]interface{})
			if !ok {
				continue
			}
			if typ := experienceString(d["type"]); typ != "" {
				seen[typ] = f.Name() + ":" + name
			}
		}
	}
	if len(seen) == 0 {
		t.Fatal("no binding types found in the corpus — fixture path is wrong")
	}
	for typ, where := range seen {
		if !experienceBindingKinds[typ] {
			t.Errorf("the harvested corpus declares binding type %q (%s) which experienceBindingKinds does not know — every fork using it would be refused", typ, where)
		}
	}
}
