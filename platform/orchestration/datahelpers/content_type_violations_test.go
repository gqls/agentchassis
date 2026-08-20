package datahelpers

import "testing"

// The live fixture: mechanism-flow's real input_schema shape, trimmed to the
// two fields that matter. `branches` is declared INSIDE `steps`' item shape,
// which is the whole reason the checker recurses — the violation that motivated
// bugs_open/260 is nested, and a top-level-only check reports nothing at all
// for it.
func mechanismFlowSchema() map[string]interface{} {
	return map[string]interface{}{
		"fields": map[string]interface{}{
			"intro": map[string]interface{}{"type": "text", "source": "llm"},
			"steps": map[string]interface{}{
				"type":     "array",
				"source":   "llm",
				"required": true,
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"title": map[string]interface{}{"type": "string"},
						"body":  map[string]interface{}{"type": "string"},
						"branches": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"label": map[string]interface{}{"type": "string"},
									"body":  map[string]interface{}{"type": "string"},
								},
							},
						},
					},
				},
			},
		},
	}
}

func TestNestedViolationIsFoundWithItsPath(t *testing.T) {
	content := map[string]interface{}{
		"steps": []interface{}{
			map[string]interface{}{"title": "Assess", "body": "We look."},
			map[string]interface{}{"title": "Advise", "body": "We say."},
			map[string]interface{}{
				"title": "Decide",
				// The live defect: the writer produced a sentence where the
				// schema declares a list of objects.
				"branches": "Either we file, or we appeal.",
			},
		},
	}

	got := ContentTypeViolations(mechanismFlowSchema(), content)
	if len(got) != 1 {
		t.Fatalf("want exactly one violation, got %d: %v", len(got), got)
	}
	if got[0].Path != "steps[2].branches" {
		t.Errorf("Path = %q, want steps[2].branches — the index is the diagnosis", got[0].Path)
	}
	if got[0].Declared != "array (items: object)" || got[0].Actual != "string" {
		t.Errorf("Declared/Actual = %q/%q, want %q/%q", got[0].Declared, got[0].Actual,
			"array (items: object)", "string")
	}
}

// CONTROL: the same schema with correctly shaped content must be silent. A
// checker that fires on everything is not a checker, and the A/B pair is what
// makes the test above evidence rather than assertion.
func TestCorrectlyShapedContentHasNoViolations(t *testing.T) {
	content := map[string]interface{}{
		"steps": []interface{}{
			map[string]interface{}{
				"title": "Decide",
				"branches": []interface{}{
					map[string]interface{}{"label": "File", "body": "We file."},
					map[string]interface{}{"label": "Appeal", "body": "We appeal."},
				},
			},
		},
	}
	if got := ContentTypeViolations(mechanismFlowSchema(), content); len(got) != 0 {
		t.Fatalf("correct content produced violations: %v", got)
	}
}

// Absent, nil and empty are the presence gate's business, never this one's.
// Reporting them here would make two gates disagree about the same content, and
// the render-time refusal would fire on a field the writer legitimately omitted.
func TestAbsentNilAndEmptyAreNeverViolations(t *testing.T) {
	for name, content := range map[string]map[string]interface{}{
		"absent": {"intro": "hello"},
		"nil":    {"steps": nil},
		"empty":  {"steps": []interface{}{}},
		// THE LIVE ROW. fundamentallyai.com's production-backend-engineering
		// page stores five steps whose `branches` are the empty string, and it
		// renders clean and serves today because the template gates them
		// (`{{if $s.branches}}` guards the `{{range}}`). The first version of
		// this checker called that a violation, which would have refused a
		// rebuild of a healthy live page. Nothing but the census found it.
		"empty string in an array-declared field": {"steps": []interface{}{
			map[string]interface{}{"title": "Assess", "branches": ""},
		}},
		"whitespace string": {"steps": []interface{}{
			map[string]interface{}{"title": "Assess", "branches": "   "},
		}},
	} {
		if got := ContentTypeViolations(mechanismFlowSchema(), content); len(got) != 0 {
			t.Errorf("%s: want no violations, got %v", name, got)
		}
	}
}

// The top-level shape violation — the simplest form of the same defect.
func TestTopLevelStringWhereArrayDeclared(t *testing.T) {
	got := ContentTypeViolations(mechanismFlowSchema(),
		map[string]interface{}{"steps": "We assess, then we advise."})
	if len(got) != 1 || got[0].Path != "steps" || got[0].Actual != "string" {
		t.Fatalf("want one violation on steps/string, got %v", got)
	}
}

// An array of STRINGS under an items:object declaration: each element is
// reported with its index, because "which one" is the part a human needs.
func TestArrayOfScalarsUnderObjectItems(t *testing.T) {
	got := ContentTypeViolations(mechanismFlowSchema(),
		map[string]interface{}{"steps": []interface{}{"one", "two"}})
	if len(got) != 2 {
		t.Fatalf("want one violation per element, got %v", got)
	}
	if got[0].Path != "steps[0]" || got[1].Path != "steps[1]" {
		t.Errorf("paths = %q, %q — want indexed element paths", got[0].Path, got[1].Path)
	}
	if got[0].Declared != "object" || got[0].Actual != "string" {
		t.Errorf("element violation = %q/%q, want object/string", got[0].Declared, got[0].Actual)
	}
}

// The OTHER live items dialect (12 of the 14 llm array fields): items is an
// example-value map, not a JSON-Schema block. Both must be understood, or the
// checker is silent on the majority of the estate.
func TestExampleValueItemsDialect(t *testing.T) {
	schema := map[string]interface{}{
		"fields": map[string]interface{}{
			"questions": map[string]interface{}{
				"type":   "array",
				"source": "llm",
				"items":  map[string]interface{}{"question": "string", "answer": "string"},
			},
		},
	}
	if got := ContentTypeViolations(schema, map[string]interface{}{"questions": "just one, really"}); len(got) != 1 {
		t.Fatalf("want a violation for a string where the faq array is declared, got %v", got)
	}
	ok := map[string]interface{}{"questions": []interface{}{
		map[string]interface{}{"question": "q", "answer": "a"},
	}}
	if got := ContentTypeViolations(schema, ok); len(got) != 0 {
		t.Fatalf("correct faq content produced violations: %v", got)
	}
}

// `list` is a real declared type on this estate (5 fields), not a typo to be
// normalised away.
func TestListIsTreatedAsArray(t *testing.T) {
	schema := map[string]interface{}{"fields": map[string]interface{}{
		"points": map[string]interface{}{"type": "list", "source": "llm"},
	}}
	if got := ContentTypeViolations(schema, map[string]interface{}{"points": "a, b, c"}); len(got) != 1 {
		t.Fatalf("a `list` declaration must be checked like `array`, got %v", got)
	}
}

// Everything the checker must stay OUT of: non-llm sources (they have their own
// guard), undeclared types, schema-less components, and declared scalars.
func TestConservativeSkips(t *testing.T) {
	cases := map[string]struct {
		schema  map[string]interface{}
		content map[string]interface{}
	}{
		"resolver-filled array": {
			map[string]interface{}{"fields": map[string]interface{}{
				"news": map[string]interface{}{"type": "array", "source": "query"}}},
			map[string]interface{}{"news": "not an array"},
		},
		"undeclared type": {
			map[string]interface{}{"fields": map[string]interface{}{
				"blob": map[string]interface{}{"source": "llm"}}},
			map[string]interface{}{"blob": "anything"},
		},
		"declared scalar holding a map": {
			map[string]interface{}{"fields": map[string]interface{}{
				"heading": map[string]interface{}{"type": "text", "source": "llm"}}},
			map[string]interface{}{"heading": map[string]interface{}{"a": "b"}},
		},
		"no schema at all": {nil, map[string]interface{}{"steps": "wrong"}},
		"bare example-value schema": {
			map[string]interface{}{"headline": "string"},
			map[string]interface{}{"steps": "wrong"},
		},
	}
	for name, c := range cases {
		if got := ContentTypeViolations(c.schema, c.content); len(got) != 0 {
			t.Errorf("%s: want silence, got %v", name, got)
		}
	}
}

func TestDescribeTypeViolationsIsEmptyForNone(t *testing.T) {
	if s := DescribeTypeViolations(nil); s != "" {
		t.Errorf("DescribeTypeViolations(nil) = %q, want empty so callers can append it unconditionally", s)
	}
	s := DescribeTypeViolations([]TypeViolation{{Path: "steps[2].branches", Declared: "array", Actual: "string"}})
	if s != "steps[2].branches: declared array, got string" {
		t.Errorf("DescribeTypeViolations = %q", s)
	}
}
