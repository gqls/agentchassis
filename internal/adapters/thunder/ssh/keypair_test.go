// FILE: internal/adapters/thunder/ssh/keypair_test.go

package ssh

import (
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestGenerateKeypair_RoundTrip(t *testing.T) {
	kp, err := GenerateKeypair("thunder-instance-test-uuid")
	if err != nil {
		t.Fatalf("GenerateKeypair: %v", err)
	}

	// Sanity: non-empty outputs.
	if kp.PrivatePEM == "" {
		t.Fatal("PrivatePEM is empty")
	}
	if kp.PublicAuthorizedKey == "" {
		t.Fatal("PublicAuthorizedKey is empty")
	}

	// Private key should be parseable back into an *ed25519.PrivateKey.
	signer, err := ssh.ParsePrivateKey([]byte(kp.PrivatePEM))
	if err != nil {
		t.Fatalf("parse private key: %v", err)
	}
	if signer.PublicKey().Type() != "ssh-ed25519" {
		t.Errorf("private key type: got %q want ssh-ed25519", signer.PublicKey().Type())
	}

	// Public key should be ssh-ed25519 with the comment we provided.
	if !strings.HasPrefix(kp.PublicAuthorizedKey, "ssh-ed25519 ") {
		t.Errorf("public key should start with 'ssh-ed25519 ', got: %s", kp.PublicAuthorizedKey)
	}
	if !strings.HasSuffix(kp.PublicAuthorizedKey, " thunder-instance-test-uuid") {
		t.Errorf("public key should end with comment, got: %s", kp.PublicAuthorizedKey)
	}

	// Public key should parse via ssh.ParseAuthorizedKey.
	pubKey, comment, _, _, err := ssh.ParseAuthorizedKey([]byte(kp.PublicAuthorizedKey))
	if err != nil {
		t.Fatalf("parse authorized_key: %v", err)
	}
	if pubKey.Type() != "ssh-ed25519" {
		t.Errorf("authorized_key type: got %q want ssh-ed25519", pubKey.Type())
	}
	if comment != "thunder-instance-test-uuid" {
		t.Errorf("comment: got %q want thunder-instance-test-uuid", comment)
	}

	// The signer's public key should match the parsed authorized_key.
	// (Comparing wire bytes via MarshalAuthorizedKey is the simplest way.)
	signerPubLine := strings.TrimRight(string(ssh.MarshalAuthorizedKey(signer.PublicKey())), "\n")
	parsedPubLine := strings.TrimRight(string(ssh.MarshalAuthorizedKey(pubKey)), "\n")
	if signerPubLine != parsedPubLine {
		t.Errorf("private key's public != separate public key:\n  priv: %s\n  pub:  %s",
			signerPubLine, parsedPubLine)
	}
}

func TestGenerateKeypair_NoComment(t *testing.T) {
	kp, err := GenerateKeypair("")
	if err != nil {
		t.Fatalf("GenerateKeypair: %v", err)
	}
	// No trailing space when comment is empty.
	if strings.HasSuffix(kp.PublicAuthorizedKey, " ") {
		t.Errorf("PublicAuthorizedKey has trailing space when comment is empty: %q", kp.PublicAuthorizedKey)
	}
}

func TestGenerateKeypair_UniqueKeys(t *testing.T) {
	// Two consecutive generations must produce different keys
	// (sanity check that we're actually reading from crypto/rand, not stub).
	kp1, _ := GenerateKeypair("a")
	kp2, _ := GenerateKeypair("b")
	if kp1.PrivatePEM == kp2.PrivatePEM {
		t.Error("two GenerateKeypair calls produced identical private keys")
	}
}
