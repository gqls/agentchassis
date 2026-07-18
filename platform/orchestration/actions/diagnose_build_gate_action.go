// FILE: platform/orchestration/actions/diagnose_build_gate_action.go
//
// F1.1b(c) part 2b: the BUILD GATE (owner decision 2026-07-12: option B —
// "I don't want to approve PRs for broken code; good to have tested it in a
// container"). Before create_pull_request runs, the fix branch must survive
// gofmt + go build inside a short-lived Kubernetes Job spun from a stock
// golang image — the chassis image deliberately carries no toolchain.
//
// A failed build is a RESULT (passed=false + the build log), not an action
// error: the workflow routes on it (green → create_pull_request; red →
// escalate with the log attached). Only Job-machinery failures (cannot create,
// cannot read) error the step.
//
// Two hard-won scoping rules (this repo, 2026-07-12): `go build ./...` at the
// repo root FAILS on pre-existing docs-dir package clashes, and repo-wide
// `gofmt -l` flags files that were never formatted — so the gate builds
// TARGETED paths and formats ONLY the implementation's changed files, else
// every run fails on inherited mess the plan never touched.
//
// The Job reads GITHUB_READ_TOKEN via secretKeyRef (personae-platform-secrets)
// exactly like the diagnose-agent spawn gate: the chassis pod never holds the
// token. TTLSecondsAfterFinished keeps the Job (and its logs) inspectable for
// an hour, then k8s reaps it.
package actions

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var DiagnoseBuildGateInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"branch"},
	Optional: []string{
		"repo_owner", "repo_name", "changed_files_field",
		"build_targets", "image", "timeout_seconds", "namespace",
		"test_packages_field",
	},
	Defaults: map[string]interface{}{
		"repo_owner":          "gqls",
		"repo_name":           "agentchassis",
		"changed_files_field": "commit_prep.files",
		"build_targets":       []interface{}{"./platform/...", "./internal/...", "./pkg/...", "./cmd/..."},
		"image":               "golang:1.24",
		"timeout_seconds":     600,
		"namespace":           "ai-persona-system",
	},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("diagnose_build_gate", DiagnoseBuildGateInputSpec)
}

// DiagnoseBuildGateAction runs gofmt+build for a fix branch in a golang Job.
func DiagnoseBuildGateAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger.With(zap.String("action", "diagnose_build_gate"))
	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, config, DiagnoseBuildGateInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}
	branch := strings.TrimSpace(inputs.Get("branch"))
	if branch == "" {
		return nil, fmt.Errorf("branch is empty")
	}
	owner := datahelpers.GetStringField(config, "repo_owner", "gqls")
	repo := datahelpers.GetStringField(config, "repo_name", "agentchassis")
	image := datahelpers.GetStringField(config, "image", "golang:1.24")
	namespace := datahelpers.GetStringField(config, "namespace", "ai-persona-system")
	timeoutSecs := datahelpers.GetIntField(config, "timeout_seconds", 600)
	targets := configStringSlice(config, "build_targets", []string{"./platform/...", "./internal/...", "./pkg/...", "./cmd/..."})

	// The changed .go files (keys of prepare's files map) — the gofmt scope.
	var goFiles []string
	if cf := datahelpers.GetStringField(config, "changed_files_field", "commit_prep.files"); cf != "" {
		if m, ok := datahelpers.ExtractNestedField(params.CollectedData, cf).(map[string]interface{}); ok {
			for path := range m {
				if strings.HasSuffix(path, ".go") {
					goFiles = append(goFiles, path)
				}
			}
			sort.Strings(goFiles) // deterministic script for identical inputs
		}
	}

	// Stage-loop end gate (delta 2, E2/D6): go test over packages the ROUTER
	// derived from the plan's edited .go files — never a model-declared list.
	// Unset keeps the gate build-only (the per-stage and fix-loop behaviour).
	// CONFIGURED-BUT-EMPTY IS AN ERROR (council-gate review 5a65ec4c,
	// bug-historian): an unset field means "build-only gate", but a field that
	// IS configured and resolves to nothing means the derived package list never
	// arrived — and silently running a build-only gate would forfeit exactly the
	// D6 guarantee this mode exists to provide (the model cannot narrow its own
	// test surface). Fail loudly instead of gating on less than intended.
	var testPkgs []string
	if tf := datahelpers.GetStringField(config, "test_packages_field", ""); tf != "" {
		testPkgs = collectedStringSlice(params.CollectedData, tf)
		if len(testPkgs) == 0 {
			return nil, fmt.Errorf("test_packages_field %q is configured but resolved to no packages — refusing to run a build-only gate in end-gate mode", tf)
		}
	}

	script := buildGateScript(owner, repo, branch, goFiles, targets, testPkgs)
	jobName := gateJobName(branch)

	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("in-cluster config: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("k8s client: %w", err)
	}

	// Re-runs replace a finished gate Job of the same name.
	if existing, err := clientset.BatchV1().Jobs(namespace).Get(ctx, jobName, metav1.GetOptions{}); err == nil && existing != nil {
		policy := metav1.DeletePropagationBackground
		if err := clientset.BatchV1().Jobs(namespace).Delete(ctx, jobName, metav1.DeleteOptions{PropagationPolicy: &policy}); err != nil && !k8serrors.IsNotFound(err) {
			return nil, fmt.Errorf("delete prior gate job: %w", err)
		}
		// Give the API a moment to reap before recreating under the same name.
		time.Sleep(3 * time.Second)
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":        "diagnose-build-gate",
				"fix-branch": sanitizeLabel(branch),
			},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            int32Ptr(0),    // one attempt — a retry would hide flakiness
			TTLSecondsAfterFinished: int32Ptr(3600), // logs inspectable for an hour, then reaped
			ActiveDeadlineSeconds:   int64Ptr(int64(timeoutSecs)),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "diagnose-build-gate"},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{{
						Name:    "build-gate",
						Image:   image,
						Command: []string{"/bin/sh", "-ec", script},
						Env: []corev1.EnvVar{{
							// Same secretKeyRef pattern as the diagnose-agent
							// spawn gate: the token goes straight from the
							// Secret into this pod.
							Name: "GITHUB_READ_TOKEN",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{Name: "personae-platform-secrets"},
									Key:                  "GITHUB_READ_TOKEN",
								},
							},
						}},
					}},
				},
			},
		},
	}

	if _, err := clientset.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{}); err != nil {
		return nil, fmt.Errorf("create gate job: %w", err)
	}
	logger.Info("diagnose_build_gate: job created",
		zap.String("job", jobName),
		zap.String("branch", branch),
		zap.String("orchestration_id", orchIDForLog(params)))

	// Poll to completion. The Job's ActiveDeadlineSeconds is the true bound;
	// the loop deadline runs slightly past it so k8s marks the failure first.
	deadline := time.Now().Add(time.Duration(timeoutSecs+60) * time.Second)
	passed := false
	outcome := "timeout"
	for time.Now().Before(deadline) {
		got, err := clientset.BatchV1().Jobs(namespace).Get(ctx, jobName, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("poll gate job: %w", err)
		}
		if got.Status.Succeeded > 0 {
			passed, outcome = true, "succeeded"
			break
		}
		if got.Status.Failed > 0 {
			passed, outcome = false, "failed"
			break
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("gate polling cancelled: %w", ctx.Err())
		case <-time.After(5 * time.Second):
		}
	}

	logTail := fetchGateLog(ctx, clientset, namespace, jobName)

	logger.Info("diagnose_build_gate: finished",
		zap.String("job", jobName),
		zap.String("outcome", outcome),
		zap.Bool("passed", passed),
		zap.String("orchestration_id", orchIDForLog(params)))

	return map[string]interface{}{
		"passed":   passed,
		"outcome":  outcome, // succeeded | failed | timeout
		"job_name": jobName,
		"log":      logTail,
	}, nil
}

// buildGateScript renders the container script. Pure — tested directly.
// gofmt checks ONLY the changed .go files; go build runs ONLY the configured
// targets (see the file header for why neither may be repo-wide); go test runs
// ONLY the router-derived packages, and only in end-gate mode (testPkgs set).
func buildGateScript(owner, repo, branch string, goFiles, targets, testPkgs []string) string {
	var b strings.Builder
	b.WriteString("set -e\n")
	b.WriteString("echo '=== build gate: clone ==='\n")
	fmt.Fprintf(&b, "git clone --depth 1 --branch %q https://x-access-token:${GITHUB_READ_TOKEN}@github.com/%s/%s.git /workspace\n", branch, owner, repo)
	b.WriteString("cd /workspace\n")
	if len(goFiles) > 0 {
		b.WriteString("echo '=== build gate: gofmt (changed files only) ==='\n")
		fmt.Fprintf(&b, "UNFORMATTED=$(gofmt -l %s)\n", shellQuoteAll(goFiles))
		b.WriteString("if [ -n \"$UNFORMATTED\" ]; then echo \"gofmt FAILED for: $UNFORMATTED\"; exit 1; fi\n")
	}
	b.WriteString("echo '=== build gate: go build (targeted) ==='\n")
	for _, t := range targets {
		fmt.Fprintf(&b, "go build %s\n", shellQuote(t))
	}
	if len(testPkgs) > 0 {
		b.WriteString("echo '=== build gate: go test (derived packages) ==='\n")
		fmt.Fprintf(&b, "go test -count=1 %s\n", shellQuoteAll(testPkgs))
	}
	b.WriteString("echo '=== build gate: PASS ==='\n")
	return b.String()
}

// gateJobName derives a k8s-legal Job name from the branch (lowercase
// alphanumerics and '-', bounded length). Deterministic so re-runs replace.
func gateJobName(branch string) string {
	s := strings.ToLower(branch)
	var out strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			out.WriteRune(r)
		} else {
			out.WriteRune('-')
		}
	}
	name := "build-gate-" + strings.Trim(out.String(), "-")
	if len(name) > 60 {
		name = name[:60]
	}
	return strings.TrimRight(name, "-")
}

func sanitizeLabel(s string) string {
	return gateJobName(s)[len("build-gate-"):]
}

// shellQuote single-quotes one token for /bin/sh.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func shellQuoteAll(items []string) string {
	quoted := make([]string, len(items))
	for i, it := range items {
		quoted[i] = shellQuote(it)
	}
	return strings.Join(quoted, " ")
}

// fetchGateLog returns the tail of the gate pod's log — the human-facing
// failure report. Best-effort: a log fetch failure must not mask the gate
// verdict, so it degrades to a note.
func fetchGateLog(ctx context.Context, clientset *kubernetes.Clientset, namespace, jobName string) string {
	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "job-name=" + jobName,
	})
	if err != nil || len(pods.Items) == 0 {
		return "(gate pod log unavailable)"
	}
	// Most recent pod wins (BackoffLimit 0 → normally exactly one).
	pod := pods.Items[len(pods.Items)-1]
	tail := int64(200)
	raw, err := clientset.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{TailLines: &tail}).DoRaw(ctx)
	if err != nil {
		return fmt.Sprintf("(gate pod log unavailable: %v)", err)
	}
	return string(raw)
}
