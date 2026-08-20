package storage

import "testing"

// The boundary is asserted AT the ceiling and one minute either side, because the
// live measurement that justifies this constant found the refusal one SECOND past
// 604800 — a test that only checks "six weeks clamps" would pass against an
// off-by-a-lot ceiling.
func TestClampPresignExpiryMinutes(t *testing.T) {
	cases := []struct {
		name     string
		in, want int
	}{
		{"the ceiling itself survives", MaxPresignExpiryMinutes, MaxPresignExpiryMinutes},
		{"one minute inside survives", MaxPresignExpiryMinutes - 1, MaxPresignExpiryMinutes - 1},
		{"one minute past clamps", MaxPresignExpiryMinutes + 1, MaxPresignExpiryMinutes},
		{"six weeks clamps", 6 * 7 * 24 * 60, MaxPresignExpiryMinutes},
		{"a normal hour survives", 60, 60},
		{"one minute survives", 1, 1},
		{"zero is a dead URL, so it clamps up", 0, MaxPresignExpiryMinutes},
		{"negative clamps up", -5, MaxPresignExpiryMinutes},
	}
	for _, c := range cases {
		if got := ClampPresignExpiryMinutes(c.in); got != c.want {
			t.Errorf("%s: ClampPresignExpiryMinutes(%d) = %d, want %d", c.name, c.in, got, c.want)
		}
	}
}

// The constant IS the protocol number. Asserted in seconds because that is the
// unit the signature carries and the unit the measurement was taken in — a minutes
// figure that looks plausible can still be the wrong number of seconds.
func TestTheCeilingIsTheSigV4Number(t *testing.T) {
	if got := MaxPresignExpiryMinutes * 60; got != 604800 {
		t.Fatalf("MaxPresignExpiryMinutes is %d minutes = %d seconds, want 604800 "+
			"(measured 2026-08-20: 604800 -> HTTP 404 NoSuchKey, 604801 -> HTTP 403 SignatureDoesNotMatch)",
			MaxPresignExpiryMinutes, got)
	}
}
