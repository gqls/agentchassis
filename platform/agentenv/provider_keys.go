// Package agentenv holds the environment contract for SPAWNED agent pods.
//
// WHY THIS PACKAGE EXISTS (bugs_open/112, 2026-07-27)
// ---------------------------------------------------
// A spawned agent pod does NOT inherit the spawner's environment, and it does
// not take `envFrom: secretRef: personae-default-secrets` the way the long-lived
// Deployments do — its EnvFrom carries only the ConfigMap. Every secret it gets
// is named explicitly in Go.
//
// That list is deliberately an allow-list and not a blanket secretRef: spawned
// pods run arbitrary agent definitions, and handing all twelve keys in
// personae-default-secrets to every one of them would undo the same
// least-privilege boundary that keeps the repo-read GitHub token scoped to
// repo-cloning agents only. The allow-list stays.
//
// What was NOT deliberate was maintaining it TWICE. Two spawners build a pod
// env — the chassis (platform/orchestration/actions/spawn_actions.go) and the
// remote job spawner (cmd/remote-job-spawner/main.go) — and they had drifted:
// the chassis carried ANTHROPIC and GROK, the remote spawner carried ANTHROPIC
// alone, and NEITHER carried GEMINI when page-content-writer was switched to
// Gemini in the live database. The writer then failed at client construction in
// a spawned pod, and with no error_step on that step it took the whole page
// build with it.
//
// So the cost of the allow-list is paid once here instead of once per spawner:
//
//	ADDING A PROVIDER IS A TWO-PLACE CHANGE — the agent definition's
//	ai_service.api_key_env_var, and this list. Both spawners read this list, so
//	the second place is now singular. A key absent here is absent at runtime
//	however many keys the secret holds, because the client does
//	os.Getenv(apiKeyEnvVar) and fails at construction on an empty string
//	(platform/aiservice/gemini.go and its siblings).
//
// This half is Go, so it is inert until the image is rebuilt and rolled.
package agentenv

import corev1 "k8s.io/api/core/v1"

// defaultSecretName is the Secret every provider API key is read from. It is the
// same Secret the long-lived Deployments take wholesale via envFrom; spawned
// pods take only the named subset below.
const defaultSecretName = "personae-default-secrets"

// providerKeyNames is THE list of provider API keys a spawned agent pod may use.
//
// Each entry is both the environment variable name and the key within the
// Secret — they are identical by convention, and keeping them identical is what
// lets an agent definition name a provider by its env var alone.
//
// Order is not significant. Add a provider here and BOTH spawners get it.
var providerKeyNames = []string{
	"ANTHROPIC_API_KEY",
	"GROK_API_KEY",
	"GEMINI_API_KEY",
}

// ProviderKeyEnv returns the provider API key environment variables that every
// spawned agent pod receives, as SecretKeyRef references rather than values —
// nothing is read into the spawner's own memory, and the value never appears in
// a pod spec, a log line, or a dispatch message.
//
// Callers append the result to their own env list; it deliberately contains ONLY
// provider API keys. Database passwords, the bootstrap key and any per-agent-type
// scoped credential (such as the repo-read GitHub token) stay with their caller,
// because those are not shared between the two spawners and at least one of them
// is intentionally granted to a single agent type.
func ProviderKeyEnv() []corev1.EnvVar {
	env := make([]corev1.EnvVar, 0, len(providerKeyNames))
	for _, name := range providerKeyNames {
		env = append(env, corev1.EnvVar{
			Name: name,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: defaultSecretName,
					},
					Key: name,
				},
			},
		})
	}
	return env
}

// ProviderKeyNames returns the env var names in the allow-list. It exists so a
// test or a preflight check can assert reachability of a configured
// ai_service.api_key_env_var WITHOUT constructing a pod spec.
func ProviderKeyNames() []string {
	out := make([]string, len(providerKeyNames))
	copy(out, providerKeyNames)
	return out
}
