// FILE: internal/adapters/thunder/ssh_exec_actions.go
//
// Phase 4 SSH actions: ssh_exec and ssh_get_status.
//
// VERIFIED MECHANICS (2026-05-24, against a live adapter-provisioned instance):
//   - SSH is a DIRECT x/crypto/ssh dial to instance_ip:ssh_port. No proxy, no
//     tunnel, no shelling out to the `tnr` CLI.
//   - The login user is `ubuntu` (NOT root — Thunder's printed ssh_command says
//     root@, but the real user is ubuntu; every root attempt was denied).
//   - The adapter-generated keypair (public half sent as public_key on create,
//     private half in k8s Secret thunder-ssh-<db_row_id>) authenticates as
//     ubuntu. Thunder honours our public_key.
//   - ssh_port is the /instances/list `port` (captured at provision into
//     thunder_instances.ssh_port). Verified == the tnr-connect-resolved port
//     and directly dialable.
//   - The instance reaches Thunder status RUNNING BEFORE sshd is accepting, so
//     a fresh dial can get "connection refused" for ~30–60s. Execute therefore
//     waits-for-sshd (retry the dial) before giving up.
//
// These actions look up the instance row themselves (ip/port/ssh_user/secret)
// rather than depending on the store's Instance struct shape, keeping the file
// self-contained and resilient to store changes.

package thunder

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	gossh "golang.org/x/crypto/ssh"

	"github.com/gqls/agentchassis/internal/adapters/thunder/store"
)

const (
	// sshDialTimeout bounds a single dial+handshake attempt.
	sshDialTimeout = 15 * time.Second
	// sshWaitForSSHD bounds the total wait-for-sshd window (RUNNING ≠ sshd up).
	sshWaitForSSHD = 90 * time.Second
	// sshWaitInterval is the gap between wait-for-sshd retries.
	sshWaitInterval = 5 * time.Second
	// sshCommandTimeout bounds command execution once connected.
	sshCommandTimeout = 5 * time.Minute
)

// privateKeyGetter is the subset of *ssh.SecretManager this action needs.
// The existing secretManager interface (provision_action.go) only exposes
// Create/Delete; ssh_exec needs to READ the key, so it declares its own
// narrow interface — same pattern as the other actions' narrow interfaces.
type privateKeyGetter interface {
	GetPrivateKey(ctx context.Context, secretName string) ([]byte, error)
}

// ─────────────────────────────────────────────────────────────────────────
// Request / Result shapes
// ─────────────────────────────────────────────────────────────────────────

// SSHExecRequest runs a command on a provisioned instance. Identify the
// instance by provisioning_id (DB row UUID) — the action resolves ip/port/
// user/key from the row. `command` is run via a non-interactive SSH session.
type SSHExecRequest struct {
	ProvisioningID string `json:"provisioning_id"`
	Command        string `json:"command"`
	// TimeoutSeconds optionally overrides the command execution timeout.
	TimeoutSeconds int `json:"timeout_seconds,omitempty"`
}

// SSHGetStatusRequest checks reachability + (optionally) a status command.
// Identify by provisioning_id. If StatusCommand is empty, a default
// readiness probe (`echo ready`) is used — useful as a wait-for-sshd gate.
type SSHGetStatusRequest struct {
	ProvisioningID string `json:"provisioning_id"`
	StatusCommand  string `json:"status_command,omitempty"`
}

// SSHExecResult is the shared result shape.
type SSHExecResult struct {
	ProvisioningID string `json:"provisioning_id"`
	ExitCode       int    `json:"exit_code"`
	Stdout         string `json:"stdout"`
	Stderr         string `json:"stderr"`
	Reachable      bool   `json:"reachable"` // true if sshd accepted + authed
}

// connectionInfo — what dial() needs, derived from store.Instance.
type connectionInfo struct {
	IP            string
	Port          int
	SSHUser       string
	SSHSecretName string
}

// SSHExecAction runs commands over SSH on provisioned instances. It holds the
// DB (to resolve the instance via the store) and a privateKeyGetter (to fetch
// the private key). Constructed in NewAdapter alongside the other actions.
type SSHExecAction struct {
	db        *sql.DB
	secretMgr privateKeyGetter
	logger    *zap.Logger
}

// NewSSHExecAction builds an SSHExecAction.
func NewSSHExecAction(db *sql.DB, secretMgr privateKeyGetter, logger *zap.Logger) *SSHExecAction {
	return &SSHExecAction{
		db:        db,
		secretMgr: secretMgr,
		logger:    logger.Named("ssh_exec"),
	}
}

// loadConnectionInfo resolves connection details via the store (reusing
// store.LookupByID — the same lookup decommission uses), then validates the
// fields ssh_exec needs.
func (s *SSHExecAction) loadConnectionInfo(ctx context.Context, provisioningID string) (*connectionInfo, error) {
	id, err := uuid.Parse(provisioningID)
	if err != nil {
		return nil, fmt.Errorf("provisioning_id %q is not a valid UUID: %w", provisioningID, err)
	}
	inst, err := store.LookupByID(ctx, s.db, id)
	if err != nil {
		if err == store.ErrInstanceNotFound {
			return nil, fmt.Errorf("no thunder_instances row with id %s", provisioningID)
		}
		return nil, fmt.Errorf("lookup instance %s: %w", provisioningID, err)
	}
	if inst.InstanceIP == "" {
		return nil, fmt.Errorf("instance %s has no instance_ip (not yet running?)", provisioningID)
	}
	if !inst.SSHPort.Valid || inst.SSHPort.Int64 == 0 {
		return nil, fmt.Errorf("instance %s has no ssh_port stored (provisioned before port capture? re-provision needed)", provisioningID)
	}
	if inst.SSHKeySecretName == "" {
		return nil, fmt.Errorf("instance %s has no ssh_key_secret_name", provisioningID)
	}
	user := inst.SSHUser
	if user == "" {
		user = "ubuntu"
	}
	return &connectionInfo{
		IP:            inst.InstanceIP,
		Port:          int(inst.SSHPort.Int64),
		SSHUser:       user,
		SSHSecretName: inst.SSHKeySecretName,
	}, nil
}

// dial establishes an SSH client, waiting for sshd to come up (RUNNING ≠ sshd
// ready). Returns a connected client or an error after the wait window.
func (s *SSHExecAction) dial(ctx context.Context, ci *connectionInfo) (*gossh.Client, error) {
	keyPEM, err := s.secretMgr.GetPrivateKey(ctx, ci.SSHSecretName)
	if err != nil {
		return nil, fmt.Errorf("fetch private key from secret %s: %w", ci.SSHSecretName, err)
	}
	signer, err := gossh.ParsePrivateKey(keyPEM)
	if err != nil {
		return nil, fmt.Errorf("parse private key from secret %s: %w", ci.SSHSecretName, err)
	}

	cfg := &gossh.ClientConfig{
		User:            ci.SSHUser,
		Auth:            []gossh.AuthMethod{gossh.PublicKeys(signer)},
		HostKeyCallback: gossh.InsecureIgnoreHostKey(), // Thunder hosts are ephemeral; we trust the network path. TODO: pin if Thunder exposes host keys.
		Timeout:         sshDialTimeout,
	}
	addr := net.JoinHostPort(ci.IP, strconv.Itoa(ci.Port))

	deadline := time.Now().Add(sshWaitForSSHD)
	var lastErr error
	for attempt := 1; ; attempt++ {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("ctx done while waiting for sshd at %s: %w", addr, ctx.Err())
		}
		client, derr := gossh.Dial("tcp", addr, cfg)
		if derr == nil {
			s.logger.Info("SSH connected",
				zap.String("addr", addr), zap.String("user", ci.SSHUser), zap.Int("attempt", attempt))
			return client, nil
		}
		lastErr = derr
		// "connection refused" == sshd not up yet; keep waiting. An auth error
		// would also surface here — we still retry until the window closes, but
		// log it so a persistent auth failure is visible.
		s.logger.Info("SSH dial not ready yet, waiting for sshd",
			zap.String("addr", addr), zap.Int("attempt", attempt), zap.Error(derr))
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("sshd not reachable at %s after %s: %w", addr, sshWaitForSSHD, lastErr)
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("ctx done while waiting for sshd at %s: %w", addr, ctx.Err())
		case <-time.After(sshWaitInterval):
		}
	}
}

// runCommand runs a single command on an established client and captures
// stdout/stderr/exit. A non-zero exit is NOT a Go error — it's returned in
// ExitCode so callers can branch on it.
func (s *SSHExecAction) runCommand(client *gossh.Client, command string) (stdout, stderr string, exitCode int, err error) {
	session, err := client.NewSession()
	if err != nil {
		return "", "", -1, fmt.Errorf("new ssh session: %w", err)
	}
	defer session.Close()

	var outBuf, errBuf bytes.Buffer
	session.Stdout = &outBuf
	session.Stderr = &errBuf

	runErr := session.Run(command)
	stdout = outBuf.String()
	stderr = errBuf.String()
	if runErr == nil {
		return stdout, stderr, 0, nil
	}
	// Distinguish "command exited non-zero" (expected, captured) from a real
	// SSH-level failure.
	if exitErr, ok := runErr.(*gossh.ExitError); ok {
		return stdout, stderr, exitErr.ExitStatus(), nil
	}
	return stdout, stderr, -1, fmt.Errorf("ssh run: %w", runErr)
}

// Exec resolves the instance, dials (wait-for-sshd), runs the command.
func (s *SSHExecAction) Exec(ctx context.Context, req SSHExecRequest) (*SSHExecResult, error) {
	if req.ProvisioningID == "" {
		return nil, fmt.Errorf("provisioning_id is required")
	}
	if req.Command == "" {
		return nil, fmt.Errorf("command is required")
	}
	ci, err := s.loadConnectionInfo(ctx, req.ProvisioningID)
	if err != nil {
		return nil, err
	}
	client, err := s.dial(ctx, ci)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	stdout, stderr, exitCode, err := s.runCommand(client, req.Command)
	if err != nil {
		return nil, err
	}
	return &SSHExecResult{
		ProvisioningID: req.ProvisioningID,
		ExitCode:       exitCode,
		Stdout:         stdout,
		Stderr:         stderr,
		Reachable:      true,
	}, nil
}

// GetStatus probes reachability + runs a status command (default `echo ready`).
// Unlike Exec, a dial failure here is returned as Reachable=false (not an
// error), because "is it up yet?" is the question being asked.
func (s *SSHExecAction) GetStatus(ctx context.Context, req SSHGetStatusRequest) (*SSHExecResult, error) {
	if req.ProvisioningID == "" {
		return nil, fmt.Errorf("provisioning_id is required")
	}
	statusCmd := req.StatusCommand
	if statusCmd == "" {
		statusCmd = "echo ready"
	}
	ci, err := s.loadConnectionInfo(ctx, req.ProvisioningID)
	if err != nil {
		return nil, err
	}
	client, err := s.dial(ctx, ci)
	if err != nil {
		// Not reachable yet — a valid status answer, not an error.
		return &SSHExecResult{
			ProvisioningID: req.ProvisioningID,
			Reachable:      false,
			Stderr:         err.Error(),
			ExitCode:       -1,
		}, nil
	}
	defer client.Close()

	stdout, stderr, exitCode, err := s.runCommand(client, statusCmd)
	if err != nil {
		return &SSHExecResult{
			ProvisioningID: req.ProvisioningID,
			Reachable:      true, // connected, but command errored at SSH level
			Stderr:         err.Error(),
			ExitCode:       -1,
		}, nil
	}
	return &SSHExecResult{
		ProvisioningID: req.ProvisioningID,
		ExitCode:       exitCode,
		Stdout:         stdout,
		Stderr:         stderr,
		Reachable:      true,
	}, nil
}

// ─────────────────────────────────────────────────────────────────────────
// Adapter handler methods (dispatch targets)
// ─────────────────────────────────────────────────────────────────────────

func (a *Adapter) handleSSHExec(
	body map[string]interface{},
	reqHeaders map[string]string,
	replyToTopic string,
	l *zap.Logger,
) {
	if replyToTopic == "" {
		l.Warn("No reply_to_topic on ssh_exec request — cannot send response")
		return
	}
	if a.sshExecAction == nil {
		a.sendErrorResponse(reqHeaders, replyToTopic, "ssh_exec",
			"ssh_unavailable", "ssh exec action not configured", "error_unrecoverable", l)
		return
	}

	bodyBytes, merr := json.Marshal(body)
	if merr != nil {
		a.sendErrorResponse(reqHeaders, replyToTopic, "ssh_exec",
			"invalid_request_body", fmt.Sprintf("could not re-marshal body: %v", merr),
			"error_unrecoverable", l)
		return
	}
	var req SSHExecRequest
	if uerr := json.Unmarshal(bodyBytes, &req); uerr != nil {
		a.sendErrorResponse(reqHeaders, replyToTopic, "ssh_exec",
			"invalid_request_body", fmt.Sprintf("could not unmarshal into SSHExecRequest: %v", uerr),
			"error_unrecoverable", l)
		return
	}

	timeout := sshCommandTimeout
	if req.TimeoutSeconds > 0 {
		timeout = time.Duration(req.TimeoutSeconds) * time.Second
	}
	// Budget = command timeout + the wait-for-sshd window.
	ctx, cancel := context.WithTimeout(context.Background(), timeout+sshWaitForSSHD)
	defer cancel()

	result, err := a.sshExecAction.Exec(ctx, req)
	if err != nil {
		// Infrastructure/transient failures (instance still booting, sshd not
		// up) → recoverable so the caller can retry; everything else
		// (bad request, parse, auth) → unrecoverable.
		status := "error_unrecoverable"
		if isInfrastructureError(err) {
			status = "error_recoverable"
		}
		a.sendErrorResponse(reqHeaders, replyToTopic, "ssh_exec",
			"ssh_exec_failed", err.Error(), status, l)
		return
	}

	a.sendSuccessResponse(reqHeaders, replyToTopic, "ssh_exec",
		map[string]interface{}{
			"provisioning_id": result.ProvisioningID,
			"exit_code":       result.ExitCode,
			"stdout":          result.Stdout,
			"stderr":          result.Stderr,
			"reachable":       result.Reachable,
		}, l)
}

func (a *Adapter) handleSSHGetStatus(
	body map[string]interface{},
	reqHeaders map[string]string,
	replyToTopic string,
	l *zap.Logger,
) {
	if replyToTopic == "" {
		l.Warn("No reply_to_topic on ssh_get_status request — cannot send response")
		return
	}
	if a.sshExecAction == nil {
		a.sendErrorResponse(reqHeaders, replyToTopic, "ssh_get_status",
			"ssh_unavailable", "ssh exec action not configured", "error_unrecoverable", l)
		return
	}

	bodyBytes, merr := json.Marshal(body)
	if merr != nil {
		a.sendErrorResponse(reqHeaders, replyToTopic, "ssh_get_status",
			"invalid_request_body", fmt.Sprintf("could not re-marshal body: %v", merr),
			"error_unrecoverable", l)
		return
	}
	var req SSHGetStatusRequest
	if uerr := json.Unmarshal(bodyBytes, &req); uerr != nil {
		a.sendErrorResponse(reqHeaders, replyToTopic, "ssh_get_status",
			"invalid_request_body", fmt.Sprintf("could not unmarshal into SSHGetStatusRequest: %v", uerr),
			"error_unrecoverable", l)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), sshWaitForSSHD+30*time.Second)
	defer cancel()

	result, err := a.sshExecAction.GetStatus(ctx, req)
	if err != nil {
		// Lookup/parse errors are unrecoverable; reachability is reported via
		// the result, not an error.
		a.sendErrorResponse(reqHeaders, replyToTopic, "ssh_get_status",
			"ssh_get_status_failed", err.Error(), "error_unrecoverable", l)
		return
	}

	a.sendSuccessResponse(reqHeaders, replyToTopic, "ssh_get_status",
		map[string]interface{}{
			"provisioning_id": result.ProvisioningID,
			"reachable":       result.Reachable,
			"exit_code":       result.ExitCode,
			"stdout":          result.Stdout,
			"stderr":          result.Stderr,
		}, l)
}
