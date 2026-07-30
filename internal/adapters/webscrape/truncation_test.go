// FILE: internal/adapters/webscrape/truncation_test.go
package webscrape

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
	"testing"

	"go.uber.org/zap"
)

// The exact sentence bugs_open/133 is about. It must not be constructible.
const theOldLie = "full version in S3"

func bigString(n int) string { return strings.Repeat("x", n) }

// ---------------------------------------------------------------------------
// The marker itself — the one place either claim exists.
// ---------------------------------------------------------------------------

func TestMarkerCannotClaimAStoredCopyWithoutNamingIt(t *testing.T) {
	withoutURI := transportTruncationMarker("")
	if strings.Contains(withoutURI, "stored at") {
		t.Errorf("marker with no URI claims a stored copy: %q", withoutURI)
	}
	if !strings.Contains(withoutURI, "DISCARDED") {
		t.Errorf("marker with no URI must say the remainder is gone, got %q", withoutURI)
	}

	const uri = "s3://bucket/webscrape/c1/20260730/abc/raw.html"
	withURI := transportTruncationMarker(uri)
	if !strings.Contains(withURI, uri) {
		t.Errorf("marker must NAME the URI it claims, got %q", withURI)
	}
	if strings.Contains(withURI, "DISCARDED") {
		t.Errorf("marker naming a URI must not also say discarded: %q", withURI)
	}
}

// A regression guard on the literal string, because the defect WAS a string:
// the claim used to be spellable with no URI in scope at all.
func TestTheOldUnconditionalClaimIsUnreachable(t *testing.T) {
	for _, m := range []string{transportTruncationMarker(""), transportTruncationMarker("s3://b/k")} {
		if strings.Contains(m, theOldLie) {
			t.Errorf("the unconditional claim is back: %q", m)
		}
	}

	// And no STRING LITERAL in the package can produce it — a second copy of the
	// sentence somewhere else would be the same bug in a new place.
	//
	// This parses and inspects literals rather than grepping the file, because a
	// substring scan cannot tell an emittable string from a comment describing
	// the defect (this very file's truncation.go quotes the old marker to explain
	// it). That distinction is not pedantry: bugs_open/153 built an evidence
	// table on five pod-grep markers, four of which were doc prose and one a Go
	// comment, and concluded the wrong thing about which binary was running.
	// A claim only ships if it is in a literal.
	for _, f := range []string{"adapter.go", "truncation.go", "batch_handler.go"} {
		for _, lit := range stringLiteralsIn(t, f) {
			if strings.Contains(lit, theOldLie) {
				t.Errorf("%s has a string literal containing the unconditional S3 claim: %q", f, lit)
			}
		}
	}
}

// stringLiteralsIn returns every string literal in a Go file — the only text
// that can reach a message. Comments are deliberately excluded.
func stringLiteralsIn(t *testing.T, filename string) []string {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", filename, err)
	}
	var out []string
	ast.Inspect(file, func(n ast.Node) bool {
		if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			if unquoted, err := strconv.Unquote(lit.Value); err == nil {
				out = append(out, unquoted)
			}
		}
		return true
	})
	if len(out) == 0 {
		t.Fatalf("%s yielded no string literals — the parse or the path is wrong, "+
			"and this check would pass vacuously", filename)
	}
	return out
}

// ---------------------------------------------------------------------------
// The measured incident, as a test.
// ---------------------------------------------------------------------------

// bugs_open/133's evidence: raw_html of 53,805 chars, upload_results false, so
// no storage map at all. 3,805 chars were destroyed and the message claimed a
// copy existed.
func TestTruncationWithNoUploadSaysDiscarded(t *testing.T) {
	result := map[string]interface{}{"raw_html": bigString(53805)}

	cut := truncateResultForTransport(result, storageInfoOf(result), zap.NewNop())

	if len(cut) != 1 || cut[0] != "raw_html" {
		t.Fatalf("cut = %v, want [raw_html]", cut)
	}
	got := result["raw_html"].(string)
	if !strings.Contains(got, "DISCARDED") {
		t.Errorf("no upload happened but the marker does not say discarded: %q", got[len(got)-160:])
	}
	if strings.Contains(got, "s3://") || strings.Contains(got, theOldLie) {
		t.Errorf("marker names or implies a store that was never written: %q", got[len(got)-160:])
	}
	if len(got) != maxTransportContentLen+len(transportTruncationMarker("")) {
		t.Errorf("unexpected truncated length %d", len(got))
	}
}

func TestTruncationWithUploadNamesThatFieldsURI(t *testing.T) {
	result := map[string]interface{}{
		"raw_html": bigString(60000),
		"storage": map[string]interface{}{
			"raw_html_uri": "s3://bucket/webscrape/c1/20260730/abc/raw.html",
		},
	}

	truncateResultForTransport(result, storageInfoOf(result), zap.NewNop())

	got := result["raw_html"].(string)
	if !strings.Contains(got, "s3://bucket/webscrape/c1/20260730/abc/raw.html") {
		t.Errorf("marker does not name the URI that was written: %q", got[len(got)-200:])
	}
	if strings.Contains(got, "DISCARDED") {
		t.Errorf("a stored field must not be marked discarded: %q", got[len(got)-200:])
	}
}

// The case per-field resolution exists for: the uploads are independent and
// best-effort, so one can succeed while another fails. A single "was there an
// upload?" flag would mark both fields as stored.
func TestTruncationResolvesURIsPerFieldNotPerRequest(t *testing.T) {
	result := map[string]interface{}{
		"html_content":     bigString(60000),
		"markdown_content": bigString(60000),
		"raw_html":         bigString(60000),
		"storage": map[string]interface{}{
			// html uploaded; markdown's upload failed (Warn, no key); raw_html too.
			"html_uri": "s3://bucket/content.html",
		},
	}

	truncateResultForTransport(result, storageInfoOf(result), zap.NewNop())

	if got := result["html_content"].(string); !strings.Contains(got, "s3://bucket/content.html") {
		t.Error("html_content was stored but its marker does not name the URI")
	}
	for _, field := range []string{"markdown_content", "raw_html"} {
		got := result[field].(string)
		if !strings.Contains(got, "DISCARDED") {
			t.Errorf("%s upload failed, so its marker must say discarded: %q", field, got[len(got)-160:])
		}
		if strings.Contains(got, "s3://") {
			t.Errorf("%s marker names a URI that was never recorded for it", field)
		}
	}
}

// ---------------------------------------------------------------------------
// The compaction trap — resolving a page URI by position would be wrong.
// ---------------------------------------------------------------------------

// uploadScrapingResults appends to storage.pages only for pages that uploaded
// something, so storage.pages[i] is NOT result.pages[i]. Here page 0 stored
// nothing, so the single storage entry belongs to page 1. Resolving by position
// would attach it to page 0 — a false claim of exactly the kind this bug is.
func TestPageURIIsResolvedByEmbeddedIndexNotByPosition(t *testing.T) {
	result := map[string]interface{}{
		"pages": []interface{}{
			map[string]interface{}{"markdown": bigString(60000)}, // page 0: nothing stored
			map[string]interface{}{"markdown": bigString(60000)}, // page 1: stored
		},
		"storage": map[string]interface{}{
			"pages": []map[string]string{
				{"markdown_uri": "s3://bucket/base/pages/page_1.md"},
			},
		},
	}

	truncateResultForTransport(result, storageInfoOf(result), zap.NewNop())

	pages := result["pages"].([]interface{})
	page0 := pages[0].(map[string]interface{})["markdown"].(string)
	page1 := pages[1].(map[string]interface{})["markdown"].(string)

	if strings.Contains(page0, "s3://") {
		t.Errorf("page 0 stored nothing but its marker names a URI: %q", page0[len(page0)-160:])
	}
	if !strings.Contains(page0, "DISCARDED") {
		t.Error("page 0 stored nothing, so its marker must say discarded")
	}
	if !strings.Contains(page1, "page_1.md") {
		t.Errorf("page 1 was stored but its marker does not name its URI: %q", page1[len(page1)-160:])
	}
}

// page_1 must not match page_10's URI. The trailing dot in the needle is what
// makes that true, so it is worth a test rather than a comment.
func TestPageURIIndexMatchIsNotAPrefixMatch(t *testing.T) {
	storage := map[string]interface{}{
		"pages": []map[string]string{
			{"markdown_uri": "s3://bucket/base/pages/page_10.md"},
		},
	}
	if uri := pageStorageURIFor(storage, "markdown", 1); uri != "" {
		t.Errorf("page 1 matched page 10's URI: %q", uri)
	}
	if uri := pageStorageURIFor(storage, "markdown", 10); uri == "" {
		t.Error("page 10 did not match its own URI")
	}
}

// Fields with no stored copy at any setting must never resolve a URI, even when
// the storage map is full of other pages' URIs.
func TestPageFieldsThatAreNeverUploadedResolveNoURI(t *testing.T) {
	storage := map[string]interface{}{
		"pages": []map[string]string{
			{"markdown_uri": "s3://bucket/base/pages/page_0.md", "html_uri": "s3://bucket/base/pages/page_0.html"},
		},
	}
	for field := range pageFieldsNeverUploaded {
		if uri := pageStorageURIFor(storage, field, 0); uri != "" {
			t.Errorf("page field %q is never uploaded but resolved %q", field, uri)
		}
	}
}

// storage.pages survives a JSON round trip as []interface{}; both shapes must
// resolve, or the marker silently reverts to claiming nothing was stored.
func TestPageURIResolvesThroughJSONShape(t *testing.T) {
	storage := map[string]interface{}{
		"pages": []interface{}{
			map[string]interface{}{"markdown_uri": "s3://bucket/base/pages/page_2.md"},
		},
	}
	if uri := pageStorageURIFor(storage, "markdown", 2); uri != "s3://bucket/base/pages/page_2.md" {
		t.Errorf("JSON-shaped storage.pages did not resolve: %q", uri)
	}
}

// ---------------------------------------------------------------------------
// The drift guard — two lists that must agree, and a real cross-check.
// ---------------------------------------------------------------------------

// Every field the truncator cuts must have a ruling: either a URI key, or an
// explicit entry saying no copy is ever stored. Adding a field to one list and
// not the other fails here rather than shipping a marker that guesses.
func TestEveryTruncatedFieldHasAURIRuling(t *testing.T) {
	for _, field := range truncatableTopLevelFields {
		if _, ok := topLevelFieldURIKeys[field]; !ok {
			t.Errorf("top-level field %q is truncated but has no URI mapping — add it to "+
				"topLevelFieldURIKeys, or its marker will claim nothing was stored when it was", field)
		}
	}
	for _, field := range truncatablePageFields {
		_, mapped := pageFieldURIKeys[field]
		if !mapped && !pageFieldsNeverUploaded[field] {
			t.Errorf("page field %q is truncated but has no ruling — map it in pageFieldURIKeys "+
				"or record it in pageFieldsNeverUploaded", field)
		}
		if mapped && pageFieldsNeverUploaded[field] {
			t.Errorf("page field %q is both mapped and declared never-uploaded", field)
		}
	}
}

// A wrong URI key is worse than a missing one: it puts another field's URI in
// this field's marker. So check the keys against the uploader's own source
// rather than against a second hand-written list, which would just be more
// drift surface.
func TestURIKeysAreOnesTheUploaderActuallyWrites(t *testing.T) {
	src, err := os.ReadFile("adapter.go")
	if err != nil {
		t.Fatalf("read adapter.go: %v", err)
	}
	source := string(src)

	// Positive control: if this needle stops matching, the assertions below
	// would all pass vacuously.
	if !strings.Contains(source, `uploadInfo["base_path"]`) {
		t.Fatal("cannot find uploadInfo assignments in adapter.go — this test has gone blind, " +
			"not the mapping (uploadScrapingResults may have been renamed or moved)")
	}

	for field, key := range topLevelFieldURIKeys {
		needle := fmt.Sprintf("uploadInfo[%q]", key)
		if !strings.Contains(source, needle) {
			t.Errorf("field %q maps to %q, but adapter.go never writes %s", field, key, needle)
		}
	}
	for field, key := range pageFieldURIKeys {
		needle := fmt.Sprintf("pageInfo[%q]", key)
		if !strings.Contains(source, needle) {
			t.Errorf("page field %q maps to %q, but adapter.go never writes %s", field, key, needle)
		}
	}
}

// ---------------------------------------------------------------------------
// Signals and no-ops.
// ---------------------------------------------------------------------------

// The only pre-existing signal that a reply was short was English prose inside
// the content. A consumer should not have to read a marker to know.
func TestTruncationSetsMachineReadableFlags(t *testing.T) {
	result := map[string]interface{}{"raw_html": bigString(60000)}
	truncateResultForTransport(result, nil, zap.NewNop())

	if result["truncated"] != true {
		t.Error("truncated flag not set")
	}
	fields, ok := result["truncated_fields"].([]string)
	if !ok || len(fields) != 1 || fields[0] != "raw_html" {
		t.Errorf("truncated_fields = %v, want [raw_html]", result["truncated_fields"])
	}
}

func TestNothingOversizedIsLeftAlone(t *testing.T) {
	result := map[string]interface{}{
		"raw_html":         "short",
		"markdown_content": bigString(maxTransportContentLen), // exactly at the cap
	}
	cut := truncateResultForTransport(result, nil, zap.NewNop())

	if len(cut) != 0 {
		t.Errorf("cut %v, want nothing", cut)
	}
	if _, present := result["truncated"]; present {
		t.Error("truncated flag set when nothing was cut")
	}
	if result["raw_html"] != "short" {
		t.Error("an under-cap field was modified")
	}
	if len(result["markdown_content"].(string)) != maxTransportContentLen {
		t.Error("a field exactly at the cap was cut — the cap is inclusive")
	}
}

func TestScreenshotBase64IsAlwaysDropped(t *testing.T) {
	result := map[string]interface{}{"screenshot_base64": "AAAA", "screenshot_uri": "s3://b/s.png"}
	truncateResultForTransport(result, nil, zap.NewNop())

	if _, present := result["screenshot_base64"]; present {
		t.Error("screenshot_base64 survived")
	}
	if result["screenshot_uri"] != "s3://b/s.png" {
		t.Error("the screenshot URI must survive — it is the useful form")
	}
}

func TestTruncateHandlesNilResult(t *testing.T) {
	if cut := truncateResultForTransport(nil, nil, nil); cut != nil {
		t.Errorf("nil result returned %v", cut)
	}
}

// ---------------------------------------------------------------------------
// The degraded form, used once after an oversize refusal.
// ---------------------------------------------------------------------------

func TestStripResultForRetryDropsRawAndShrinksTheRest(t *testing.T) {
	result := map[string]interface{}{
		"raw_html":          bigString(60000),
		"screenshot_base64": "AAAA",
		"markdown_content":  bigString(60000),
		"pages": []interface{}{
			map[string]interface{}{"raw_html": bigString(60000), "content": bigString(60000)},
		},
	}

	stripResultForRetry(result, storageInfoOf(result))

	if _, present := result["raw_html"]; present {
		t.Error("raw_html survived the strip — if the reply did not fit with it, it is what did not fit")
	}
	if _, present := result["screenshot_base64"]; present {
		t.Error("screenshot_base64 survived the strip")
	}
	if got := len(result["markdown_content"].(string)); got > oversizeStripContentCap+len(transportTruncationMarker("")) {
		t.Errorf("markdown_content is %d chars after the strip, cap is %d", got, oversizeStripContentCap)
	}

	page := result["pages"].([]interface{})[0].(map[string]interface{})
	if _, present := page["raw_html"]; present {
		t.Error("page raw_html survived the strip")
	}
	if got := len(page["content"].(string)); got > oversizeStripContentCap+len(transportTruncationMarker("")) {
		t.Errorf("page content is %d chars after the strip", got)
	}

	// A stub must be visibly a stub, or the next investigation starts from a
	// page that appears to have scraped fine.
	if result["truncated"] != true || result["degraded_for_transport"] != true {
		t.Error("a degraded reply must say so in the reply, not only in this pod's logs")
	}
}

// The strip must not resurrect the lie either.
func TestStripResultForRetryDoesNotClaimAnUnwrittenStore(t *testing.T) {
	result := map[string]interface{}{"markdown_content": bigString(60000)}
	stripResultForRetry(result, storageInfoOf(result))

	got := result["markdown_content"].(string)
	if strings.Contains(got, "s3://") || strings.Contains(got, theOldLie) {
		t.Errorf("degraded marker claims a store that was never written: %q", got[len(got)-160:])
	}
	if !strings.Contains(got, "DISCARDED") {
		t.Error("degraded marker must say the remainder is gone when nothing was stored")
	}
}

func TestStripResultForRetryHandlesNil(t *testing.T) {
	stripResultForRetry(nil, nil) // must not panic
}
