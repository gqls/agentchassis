package main

// main_test.go — the key fingerprint. Every assertion here is written to fail
// under a NAMED mutation of keyFingerprint, because a hash function is exactly
// the shape that passes a shallow test while being wrong: a stub returning a
// constant, or one that trims its input, both produce twelve plausible hex
// characters and neither would be noticed by eye in a journal.
//
// The oracle is deliberately NOT computed with Go's own sha256 — that would
// only prove the function calls the library it visibly calls. The two expected
// digests below were produced by the shell recipe the runbook gives the owner
// (`printf %s "$KEY" | sha256sum | cut -c1-12`), so these tests prove the thing
// that actually matters: HIS digest and OURS are the same number, which is what
// makes the log line checkable by someone holding the key.

import (
	"strings"
	"testing"
)

func TestKeyFingerprintMatchesTheShellRecipeOwnersUse(t *testing.T) {
	// Mutation killed: any change of algorithm, encoding, length or offset
	// (sha512, base64, first 8, last 12) — all still "look like" a digest.
	const key = "sk-ant-test-key-not-real"
	const want = "5b03b249b74a" // printf %s 'sk-ant-test-key-not-real' | sha256sum | cut -c1-12
	if got := keyFingerprint(key); got != want {
		t.Fatalf("keyFingerprint(%q) = %q, want %q — the owner's shell recipe and this "+
			"function must produce the SAME digest or the runbook check is worthless", key, got, want)
	}
}

func TestKeyFingerprintDoesNotTrimTheKey(t *testing.T) {
	// Mutation killed: strings.TrimSpace(key) before hashing. It looks like
	// defensive tidiness and it silently breaks comparability with
	// `printf %s`, which trims nothing — the two would disagree only for keys
	// with stray whitespace, i.e. exactly the copy-paste accident a fingerprint
	// is supposed to expose.
	const padded = " sk-ant-test-key-not-real "
	const want = "dc1fd431ebce" // same recipe, over the padded string
	if got := keyFingerprint(padded); got != want {
		t.Fatalf("keyFingerprint(%q) = %q, want %q — the exact bytes must be hashed, untrimmed", padded, got, want)
	}
	if keyFingerprint(padded) == keyFingerprint(strings.TrimSpace(padded)) {
		t.Fatal("padded and trimmed keys share a fingerprint — the function is trimming its input")
	}
}

func TestKeyFingerprintDistinguishesKeys(t *testing.T) {
	// Mutation killed: `return "000000000000"` (or any constant). A constant
	// passes a "is it 12 hex chars" test and makes every instance on the box
	// look identically configured — the precise lie this line exists to prevent.
	a := keyFingerprint("sk-ant-account-one")
	b := keyFingerprint("sk-ant-account-two")
	if a == b {
		t.Fatalf("two different keys share fingerprint %q — a swap would be invisible", a)
	}
}

func TestKeyFingerprintIsTwelveLowerHex(t *testing.T) {
	got := keyFingerprint("sk-ant-test-key-not-real")
	if len(got) != 12 {
		t.Fatalf("fingerprint %q is %d chars, want 12", got, len(got))
	}
	for _, c := range got {
		if !strings.ContainsRune("0123456789abcdef", c) {
			t.Fatalf("fingerprint %q contains %q — not lowercase hex, so it will not match `sha256sum`'s output", got, c)
		}
	}
}

func TestKeyFingerprintOfNoKeyIsNamed(t *testing.T) {
	// Mutation killed: hashing the empty string instead of returning "none".
	// sha256("") is a real, constant, plausible-looking digest — a journal line
	// reading `sha256=e3b0c44298fc` for an unset key would be read as a
	// configured instance. (main() log.Fatals before this in practice; the
	// guard is here so a future caller cannot reintroduce the confusion.)
	if got := keyFingerprint(""); got != "none" {
		t.Fatalf("keyFingerprint(\"\") = %q, want \"none\" — an absent key must not render as a digest", got)
	}
}

func TestKeyFingerprintNeverContainsTheKey(t *testing.T) {
	// Mutation killed: returning key[:12], or appending the key to the digest.
	// This is the whole safety claim of the log line, so it is asserted rather
	// than assumed.
	const key = "sk-ant-api03-REALLOOKINGSECRETVALUE-000"
	got := keyFingerprint(key)
	if strings.Contains(got, key) || strings.Contains(key, got) {
		t.Fatalf("fingerprint %q overlaps the key it was made from — it is not safe to log", got)
	}
}
