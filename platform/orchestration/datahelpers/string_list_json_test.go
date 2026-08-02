// FILE: platform/orchestration/datahelpers/string_list_json_test.go
//
// bugs_open/174 — ExtractStringListHelper is the last gate on a chain that lost a
// caller's `seed_scope` three lanes' diagnoses in a row, and it is the gate that
// no amount of fixing the config would have opened.
//
// The chain: a work item's jsonb `spec` -> QueryDatabaseAction, which scans every
// column into interface{} and stringifies any []byte it gets back -> input_mapping,
// which passes values through unchanged (ResolveInputMapping does no coercion) ->
// here. So a list that was a JSON array in the database arrives as the STRING
// `["a","b"]`, and returning nil for it is indistinguishable from "the caller
// supplied nothing".
//
// These tests exist to stop that arm being tidied away again. The helper's
// contract is WIDENING only, and the second half of this file is what pins that:
// everything that returned nil before must still return nil.

package datahelpers

import (
	"reflect"
	"testing"
)

func TestExtractStringListHelper_AcceptsTheShapesTheDataPathActuallyDelivers(t *testing.T) {
	want := []string{"platform/x.go:Foo", "platform/y.go"}

	cases := map[string]interface{}{
		// Already decoded — a seed passed straight down a Kafka envelope.
		"[]interface{} of strings": []interface{}{"platform/x.go:Foo", "platform/y.go"},
		"[]string":                 []string{"platform/x.go:Foo", "platform/y.go"},
		// THE 174 SHAPE. A jsonb column after query_database has stringified it.
		"JSON-array string": `["platform/x.go:Foo","platform/y.go"]`,
		// The same value if a driver ever hands back raw bytes instead. The fix is
		// deliberately correct either way, so nobody has to know which pgx does.
		"JSON-array []byte": []byte(`["platform/x.go:Foo","platform/y.go"]`),
		// Whitespace is what a jsonb_pretty or a hand-written spec produces.
		"JSON-array string, padded": "  [\n \"platform/x.go:Foo\",\n \"platform/y.go\"\n]  ",
	}

	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			if got := ExtractStringListHelper(in); !reflect.DeepEqual(got, want) {
				t.Errorf("got %#v, want %#v", got, want)
			}
		})
	}
}

// The widening-only guarantee, which is the whole argument that this change is
// safe for the other callers (v3_site_actions' config lists,
// assemble_upload_manifest, thunder_prepare_object_urls). Every input below
// returned nil before and must still return nil — a helper that started guessing
// at scalars would silently hand those callers a one-element list.
func TestExtractStringListHelper_StillRejectsEverythingItRejectedBefore(t *testing.T) {
	cases := map[string]interface{}{
		"nil":                    nil,
		"plain string":           "platform/x.go:Foo",
		"comma-separated string": "platform/x.go:Foo,platform/y.go",
		"JSON object":            `{"path":"platform/x.go"}`,
		"JSON scalar string":     `"platform/x.go"`,
		"JSON number":            `42`,
		"malformed JSON array":   `["platform/x.go"`,
		"empty string":           "",
		"empty bytes":            []byte{},
		"int":                    7,
		"map":                    map[string]interface{}{"a": "b"},
		// An array of non-strings yields nil, matching the []interface{} arm's
		// long-standing behaviour of skipping items it cannot read as strings.
		"JSON array of numbers": `[1,2,3]`,
		"JSON empty array":      `[]`,
	}

	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			if got := ExtractStringListHelper(in); got != nil {
				t.Errorf("got %#v, want nil — this helper's tolerance must not grow past a JSON array of strings", got)
			}
		})
	}
}

// Mixed arrays keep the existing skip-what-you-cannot-read behaviour rather than
// failing the whole list, so the JSON arm and the decoded arm answer identically.
func TestExtractStringListHelper_JSONArmMatchesDecodedArmOnMixedInput(t *testing.T) {
	decoded := ExtractStringListHelper([]interface{}{"a", 1, "b", nil, map[string]interface{}{}})
	fromJSON := ExtractStringListHelper(`["a",1,"b",null,{}]`)
	if !reflect.DeepEqual(decoded, fromJSON) {
		t.Errorf("decoded arm gave %#v but JSON arm gave %#v — the two must not disagree, or the shape a value happens to arrive in changes the answer", decoded, fromJSON)
	}
	if want := []string{"a", "b"}; !reflect.DeepEqual(fromJSON, want) {
		t.Errorf("got %#v, want %#v", fromJSON, want)
	}
}
