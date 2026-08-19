package git

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"testing"

	"go.uber.org/zap"
)

// bugs_open/315 / RFC_038. files_sha256 is the fingerprint the whole fix turns
// on: it is what lets a later check ask "are the bytes the origin serves the
// bytes we sent?" as a one-step comparison, instead of pulling rendered_html out
// of the database and grepping the served page for a needle.
//
// The value is only worth anything if it equals a sha256 of what is SERVED. That
// is why the base64 branch is tested first and hardest: a base64 file's `content`
// string is a transport wrapper, not the file, and hashing the wrapper produces a
// value that can never match — silently, for ever, on every asset.

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func TestSHA256OfFiles_Base64IsHashedDECODED(t *testing.T) {
	// The live producers of base64 files are derive_card_asset_action.go and
	// derive_brand_head_assets_action.go — PNG bytes, so this branch is
	// exercised in production, not hypothetically.
	raw := []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0xff}
	encoded := base64.StdEncoding.EncodeToString(raw)

	got := sha256OfFiles(map[string]interface{}{
		"assets/favicon.png": map[string]interface{}{"content": encoded, "encoding": "base64"},
	}, zap.NewNop())

	want := sha256Hex(raw)
	if got["assets/favicon.png"] != want {
		t.Errorf("base64 file hashed as %q, want the hash of the DECODED bytes %q",
			got["assets/favicon.png"], want)
	}
	// The mutation this test exists to catch: hashing the wrapper.
	if got["assets/favicon.png"] == sha256Hex([]byte(encoded)) {
		t.Error("the base64 STRING was hashed, not the bytes — this value can never equal a sha256 of the served file, and nothing would ever report the mismatch")
	}
}

func TestSHA256OfFiles_TextShapesAgree(t *testing.T) {
	const html = "<html><body>hi</body></html>"
	want := sha256Hex([]byte(html))

	// The two shapes CommitToRepo itself accepts: a bare string (legacy) and
	// {content, encoding}. They must agree, or the same page hashes differently
	// depending on which producer wrote it.
	got := sha256OfFiles(map[string]interface{}{
		"legacy.html":    html,
		"explicit.html":  map[string]interface{}{"content": html, "encoding": "utf-8"},
		"defaulted.html": map[string]interface{}{"content": html}, // encoding omitted => utf-8
	}, zap.NewNop())

	for _, k := range []string{"legacy.html", "explicit.html", "defaulted.html"} {
		if got[k] != want {
			t.Errorf("%s hashed as %q, want %q", k, got[k], want)
		}
	}
}

func TestSHA256OfFiles_KeysAreTheCallersPaths(t *testing.T) {
	// CommitToRepo prefixes paths with {domain}/ on its own copy. The chassis
	// looks a page up by the path IT sent, so the fingerprint map must be keyed
	// that way — a domain-prefixed key would simply never be found.
	got := sha256OfFiles(map[string]interface{}{"tools/x/index.html": "<html/>"}, zap.NewNop())
	if _, ok := got["tools/x/index.html"]; !ok {
		t.Fatalf("expected the caller's own path as the key, got keys: %v", keysOf(got))
	}
}

func TestSHA256OfFiles_UnusableInputIsOMITTEDNotWrong(t *testing.T) {
	// A missing key means "no fingerprint available" and a reader can act on it.
	// A wrong key means "this page is broken" and it cannot. Omission is the only
	// safe direction.
	got := sha256OfFiles(map[string]interface{}{
		"good.html":       "<html/>",
		"weird.html":      12345, // not a file shape at all
		"undecodable.png": map[string]interface{}{"content": "!!!not base64!!!", "encoding": "base64"},
	}, zap.NewNop())

	if _, ok := got["good.html"]; !ok {
		t.Error("the usable file lost its fingerprint")
	}
	if _, ok := got["weird.html"]; ok {
		t.Error("an unhashable file data type produced a fingerprint — it must be omitted")
	}
	if _, ok := got["undecodable.png"]; ok {
		t.Error("a file whose base64 did not decode produced a fingerprint — it must be omitted, never hashed as its wrapper")
	}
	if len(got) != 1 {
		t.Errorf("expected exactly 1 fingerprint, got %d: %v", len(got), keysOf(got))
	}
}

func keysOf(m map[string]string) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}
