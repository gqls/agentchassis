// FILE: internal/adapters/thunder/ssh/secrets.go
//
// k8s Secret CRUD for Thunder SSH keypairs.
//
// Naming convention: `thunder-ssh-<instance_uuid>` — deterministic so
// the reaper and decommission action can find the Secret without an
// extra lookup. The instance_uuid is the UUID returned by Thunder's
// CreateInstance API, NOT the numeric identifier.
//
// Secret data shape (corev1.SecretTypeOpaque):
//   private_key: OpenSSH-format PEM
//   public_key:  OpenSSH authorized_keys single-line
//
// Labels for filtering:
//   app.kubernetes.io/managed-by: thunder-adapter
//   thunder.adapter/instance-uuid: <uuid>
//
// Testability: SecretManager wraps a kubernetes.Interface, so unit tests
// inject `kubernetes/fake.NewSimpleClientset()`. The adapter calls
// NewInClusterSecretManager which builds a real clientset from the
// in-cluster service-account token.

package ssh

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// SecretNamePrefix is the static prefix used for all Thunder SSH Secrets.
// The thunder-adapter Role grants access scoped to this prefix
// (via name pattern in code, not k8s RBAC — see rbac.yaml).
const SecretNamePrefix = "thunder-ssh-"

// SecretNameFor returns the deterministic Secret name for a Thunder
// instance UUID. Use this everywhere to avoid mismatches between
// create and lookup sites.
func SecretNameFor(instanceUUID string) string {
	return SecretNamePrefix + instanceUUID
}

// SecretManager owns the kubernetes.Interface used for Secret CRUD.
// Safe for concurrent use (client-go is goroutine-safe).
type SecretManager struct {
	clientset kubernetes.Interface
	namespace string
	logger    *zap.Logger
}

// NewSecretManager constructs a SecretManager around an arbitrary
// kubernetes.Interface. Used directly by unit tests (passing a fake
// clientset); production code calls NewInClusterSecretManager.
func NewSecretManager(clientset kubernetes.Interface, namespace string, logger *zap.Logger) *SecretManager {
	return &SecretManager{
		clientset: clientset,
		namespace: namespace,
		logger:    logger.Named("ssh_secrets"),
	}
}

// NewInClusterSecretManager builds a SecretManager from the in-cluster
// ServiceAccount token. Returns an error if the pod isn't running with
// a SA mount, which would indicate a deployment misconfiguration.
func NewInClusterSecretManager(namespace string, logger *zap.Logger) (*SecretManager, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("in-cluster k8s config: %w", err)
	}
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("k8s clientset: %w", err)
	}
	return NewSecretManager(cs, namespace, logger), nil
}

// CreateKeypairSecret persists a generated Keypair as a k8s Secret.
// Returns the Secret name (deterministic via SecretNameFor).
//
// Not idempotent on AlreadyExists — that's an unexpected state
// (the caller shouldn't try to create the same instance's Secret twice)
// and gets surfaced as an error.
func (m *SecretManager) CreateKeypairSecret(
	ctx context.Context,
	instanceUUID string,
	kp *Keypair,
) (secretName string, err error) {
	name := SecretNameFor(instanceUUID)
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: m.namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by":  "thunder-adapter",
				"thunder.adapter/instance-uuid": instanceUUID,
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"private_key": []byte(kp.PrivatePEM),
			"public_key":  []byte(kp.PublicAuthorizedKey),
		},
	}

	_, err = m.clientset.CoreV1().Secrets(m.namespace).Create(ctx, sec, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("create secret %q: %w", name, err)
	}

	m.logger.Info("Created SSH keypair secret",
		zap.String("secret_name", name),
		zap.String("instance_uuid", instanceUUID),
	)
	return name, nil
}

// GetPrivateKey returns the private_key bytes from the named Secret.
// Used by ssh_exec actions in Phase 4. Returns an error wrapping
// kubernetes errors so callers can check apierrors.IsNotFound(err).
func (m *SecretManager) GetPrivateKey(ctx context.Context, secretName string) ([]byte, error) {
	sec, err := m.clientset.CoreV1().Secrets(m.namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get secret %q: %w", secretName, err)
	}
	priv, ok := sec.Data["private_key"]
	if !ok || len(priv) == 0 {
		return nil, fmt.Errorf("secret %q has no private_key data", secretName)
	}
	return priv, nil
}

// DeleteKeypairSecret removes the named Secret. Idempotent: a NotFound
// error from the API is treated as success (the desired end state is
// "secret doesn't exist", and that's what we get).
func (m *SecretManager) DeleteKeypairSecret(ctx context.Context, secretName string) error {
	err := m.clientset.CoreV1().Secrets(m.namespace).Delete(ctx, secretName, metav1.DeleteOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			m.logger.Info("SSH keypair secret already gone (idempotent delete)",
				zap.String("secret_name", secretName))
			return nil
		}
		return fmt.Errorf("delete secret %q: %w", secretName, err)
	}
	m.logger.Info("Deleted SSH keypair secret",
		zap.String("secret_name", secretName))
	return nil
}
