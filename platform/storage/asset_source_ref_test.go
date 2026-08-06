// FILE: platform/storage/asset_source_ref_test.go
// AssetSourceRef resolves an assets row's source object from the row itself.
// The cases below are the five row shapes measured live on 2026-08-06
// (bugs_open/152 + /155 lane) plus the traps that motivated the contract.

package storage

import "testing"

func TestAssetSourceRef(t *testing.T) {
	const (
		presigned = "https://s3.us-east-005.backblazeb2.com/personae-prod-uk001-images/images/system/20260722/abc.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Expires=604800"
		httpsObj  = "https://s3.us-east-005.backblazeb2.com/personae-prod-uk001-images/images/system/20260722/abc.png"
		s3URI     = "s3://personae-prod-uk001-images/images/system/20260722/abc.png"
		bareKey   = "images/uploads/5fe15466/20260729/7b21c824.png"
		localPath = "/assets/images/logo.png"
		literal   = "/assets/images/input-data.asset-key.jpg"
	)

	cases := []struct {
		name        string
		storagePath string
		url         string
		want        string
	}{
		// The five live row shapes.
		{"presigned url only (205 rows)", "", presigned, s3URI},
		{"flipped url with https storage_path (107 rows)", httpsObj, localPath, s3URI},
		{"flipped url, no storage_path (49 rows) — unresolvable", "", localPath, ""},
		{"bare-key storage_path (gaswholesalers logo)", bareKey, literal, bareKey},
		{"presigned url AND storage_path — storage_path wins", httpsObj, presigned, s3URI},

		// Contract edges.
		{"s3 storage_path passes through", s3URI, "", s3URI},
		{"s3 in url passes through", "", s3URI, s3URI},
		{"template-literal url with https storage_path", httpsObj, literal, s3URI},
		{"local storage_path is a deployed location, not a source", localPath, presigned, s3URI},
		{"both local — unresolvable", localPath, localPath, ""},
		{"both empty", "", "", ""},
		{"slashless value identifies nothing", "logo.png", "", ""},
		{"whitespace only", "   ", "  ", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := AssetSourceRef(tc.storagePath, tc.url); got != tc.want {
				t.Fatalf("AssetSourceRef(%q, %q) = %q, want %q", tc.storagePath, tc.url, got, tc.want)
			}
		})
	}
}

// The result contract: never an https URL, never a site-local path — the two
// forms ExtractKeyFromS3URI would mangle (it passes non-s3 input through
// as-is, so an https URL would become a "key" no bucket holds by accident
// rather than by the documented parity rule, and a local path would be
// downloaded as if it were a source).
func TestAssetSourceRefNeverReturnsWebForms(t *testing.T) {
	inputs := [][2]string{
		{"https://example.com/page.html", "/assets/images/hero.jpg"},
		{"/assets/images/og-card.png", ""},
		{"", "https://s3.us-east-005.backblazeb2.com/onlybucket"},
	}
	for _, in := range inputs {
		got := AssetSourceRef(in[0], in[1])
		if got != "" && (got[0] == '/' || (len(got) > 8 && got[:8] == "https://")) {
			t.Fatalf("AssetSourceRef(%q, %q) returned a web form: %q", in[0], in[1], got)
		}
	}
}
