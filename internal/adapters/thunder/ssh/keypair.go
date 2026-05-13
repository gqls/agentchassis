package ssh

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"strings"

	"golang.org/x/crypto/ssh"
)

// FILE: internal/adapters/thunder/ssh/keypair.go
//
// ed25519 keypair generation for SSH access to Thunder Compute instances.
//
// Output formats:
//   - Private key: OpenSSH PEM ("-----BEGIN OPENSSH PRIVATE KEY-----...")
//     Compatible with `ssh -i <file>` and all modern OpenSSH clients.
//   - Public key: single-line OpenSSH authorized_keys format
//     ("ssh-ed25519 AAAA... <comment>"), suitable for the `public_key`
//     field of Thunder's POST /instances/create.
//
// Why ed25519: small keys, fast, modern. The Thunder API accepts any
// SSH key type but ed25519 is the only one we use to keep things simple.
//
// No I/O: this package only produces strings. The caller (secrets.go)
// is responsible for persisting them.

// Keypair holds the OpenSSH-formatted private and public key materials.
type Keypair struct {
	// PrivatePEM is the OpenSSH-format private key, complete with PEM
	// headers and trailing newline. Store verbatim; ssh -i will accept it.
	PrivatePEM string

	// PublicAuthorizedKey is the single-line public key (no trailing
	// newline) suitable for ssh authorized_keys files and Thunder's
	// public_key API field. Format: "ssh-ed25519 AAAA... <comment>"
	PublicAuthorizedKey string
}

// GenerateKeypair produces a new ed25519 keypair. The comment is appended
// to the public key for traceability (typically the Thunder instance UUID
// or a similar correlation identifier — e.g. "thunder-<uuid>").
//
// Returns an error only if the underlying crypto/rand source fails, which
// indicates a system-level problem (entropy exhaustion or similar).
func GenerateKeypair(comment string) (*Keypair, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate ed25519 key: %w", err)
	}

	// Serialize private key in OpenSSH format.
	// ssh.MarshalPrivateKey requires golang.org/x/crypto v0.13.0+ (Sep 2023);
	// our base image is golang:1.24-alpine which has a recent x/crypto.
	privBlock, err := ssh.MarshalPrivateKey(priv, comment)
	if err != nil {
		return nil, fmt.Errorf("marshal ssh private key: %w", err)
	}
	privPEM := pem.EncodeToMemory(privBlock)

	// Public key in OpenSSH authorized_keys format.
	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		return nil, fmt.Errorf("ssh public key wrap: %w", err)
	}
	// MarshalAuthorizedKey returns "<type> <base64-blob>\n" without
	// the trailing comment. We append the comment ourselves so callers
	// can trace which instance a key belongs to from authorized_keys.
	base := strings.TrimRight(string(ssh.MarshalAuthorizedKey(sshPub)), "\n")
	pubLine := base
	if comment != "" {
		pubLine = base + " " + comment
	}

	return &Keypair{
		PrivatePEM:          string(privPEM),
		PublicAuthorizedKey: pubLine,
	}, nil
}
