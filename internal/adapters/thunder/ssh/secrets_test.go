// FILE: internal/adapters/thunder/ssh/secrets_test.go

package ssh

import (
	"context"
	"errors"
	"strings"
	"testing"

	"go.uber.org/zap/zaptest"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

const testNamespace = "ai-persona-system"
const testUUID = "abc123-def456-ghi789"

func testKeypair(t *testing.T) *Keypair {
	t.Helper()
	kp, err := GenerateKeypair("test-comment")
	if err != nil {
		t.Fatalf("GenerateKeypair: %v", err)
	}
	return kp
}

func newTestManager(t *testing.T) *SecretManager {
	t.Helper()
	cs := fake.NewSimpleClientset()
	return NewSecretManager(cs, testNamespace, zaptest.NewLogger(t))
}

// ─────────────────────────────────────────────────────────────────────────
// SecretNameFor
// ─────────────────────────────────────────────────────────────────────────

func TestSecretNameFor(t *testing.T) {
	got := SecretNameFor(testUUID)
	want := "thunder-ssh-" + testUUID
	if got != want {
		t.Errorf("SecretNameFor: got %q want %q", got, want)
	}
}

// ─────────────────────────────────────────────────────────────────────────
// CreateKeypairSecret
// ─────────────────────────────────────────────────────────────────────────

func TestCreateKeypairSecret_StoresDataAndLabels(t *testing.T) {
	m := newTestManager(t)
	kp := testKeypair(t)
	ctx := context.Background()

	name, err := m.CreateKeypairSecret(ctx, testUUID, kp)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if name != "thunder-ssh-"+testUUID {
		t.Errorf("returned name: got %q", name)
	}

	// Inspect what got stored via the fake clientset directly.
	sec, err := m.clientset.CoreV1().Secrets(testNamespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Get back from fake: %v", err)
	}

	// Data fields
	if got := string(sec.Data["private_key"]); got != kp.PrivatePEM {
		t.Errorf("private_key stored != generated")
	}
	if got := string(sec.Data["public_key"]); got != kp.PublicAuthorizedKey {
		t.Errorf("public_key stored != generated")
	}

	// Labels
	if got := sec.Labels["app.kubernetes.io/managed-by"]; got != "thunder-adapter" {
		t.Errorf("managed-by label: got %q", got)
	}
	if got := sec.Labels["thunder.adapter/instance-uuid"]; got != testUUID {
		t.Errorf("instance-uuid label: got %q", got)
	}
}

func TestCreateKeypairSecret_AlreadyExistsSurfaced(t *testing.T) {
	m := newTestManager(t)
	kp := testKeypair(t)
	ctx := context.Background()

	if _, err := m.CreateKeypairSecret(ctx, testUUID, kp); err != nil {
		t.Fatalf("first create: %v", err)
	}
	_, err := m.CreateKeypairSecret(ctx, testUUID, kp)
	if err == nil {
		t.Fatal("expected AlreadyExists error on second create, got nil")
	}
	// fake clientset surfaces this as a k8s AlreadyExists status error,
	// which our error wrapping doesn't strip — apierrors.IsAlreadyExists
	// should still detect it via errors.As / Unwrap chain.
	if !apierrors.IsAlreadyExists(unwrapAll(err)) {
		t.Errorf("expected apierrors.IsAlreadyExists, got: %v", err)
	}
}

// ─────────────────────────────────────────────────────────────────────────
// GetPrivateKey
// ─────────────────────────────────────────────────────────────────────────

func TestGetPrivateKey_HappyPath(t *testing.T) {
	m := newTestManager(t)
	kp := testKeypair(t)
	ctx := context.Background()

	name, err := m.CreateKeypairSecret(ctx, testUUID, kp)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	priv, err := m.GetPrivateKey(ctx, name)
	if err != nil {
		t.Fatalf("GetPrivateKey: %v", err)
	}
	if string(priv) != kp.PrivatePEM {
		t.Error("returned private key != stored")
	}
	if !strings.Contains(string(priv), "OPENSSH PRIVATE KEY") {
		t.Errorf("private key missing PEM header, got prefix: %q", string(priv)[:60])
	}
}

func TestGetPrivateKey_NotFound(t *testing.T) {
	m := newTestManager(t)
	_, err := m.GetPrivateKey(context.Background(), "thunder-ssh-nonexistent")
	if err == nil {
		t.Fatal("expected NotFound error, got nil")
	}
	if !apierrors.IsNotFound(unwrapAll(err)) {
		t.Errorf("expected apierrors.IsNotFound, got: %v", err)
	}
}

// ─────────────────────────────────────────────────────────────────────────
// DeleteKeypairSecret
// ─────────────────────────────────────────────────────────────────────────

func TestDeleteKeypairSecret_HappyPath(t *testing.T) {
	m := newTestManager(t)
	kp := testKeypair(t)
	ctx := context.Background()

	name, err := m.CreateKeypairSecret(ctx, testUUID, kp)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := m.DeleteKeypairSecret(ctx, name); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	// Confirm gone
	_, err = m.clientset.CoreV1().Secrets(testNamespace).Get(ctx, name, metav1.GetOptions{})
	if !apierrors.IsNotFound(err) {
		t.Errorf("expected NotFound after delete, got: %v", err)
	}
}

func TestDeleteKeypairSecret_IdempotentOn404(t *testing.T) {
	m := newTestManager(t)
	// Never created — delete should still succeed.
	if err := m.DeleteKeypairSecret(context.Background(), "thunder-ssh-never-existed"); err != nil {
		t.Errorf("Delete of nonexistent secret should be idempotent, got: %v", err)
	}
}

// ─────────────────────────────────────────────────────────────────────────
// helper
// ─────────────────────────────────────────────────────────────────────────

// unwrapAll walks the errors.Unwrap chain to the innermost error.
// Needed because we wrap k8s errors in fmt.Errorf("...: %w", err), and
// the helpers like apierrors.IsNotFound only inspect the immediate type.
func unwrapAll(err error) error {
	for {
		inner := errors.Unwrap(err)
		if inner == nil {
			return err
		}
		err = inner
	}
}
