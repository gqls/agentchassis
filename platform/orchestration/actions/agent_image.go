// FILE: platform/orchestration/actions/agent_image.go
package actions

// bugs_open/066 — spawned agent pods pinned stale image tags.
//
// A dedicated agent pod takes its container image from the agent_definitions
// row (image_repository + image_tag). A chassis roll updates the Deployment,
// never those rows, so every roll widened the gap: on 2026-07-24 the fleet sat
// at v1.0.1151 while the Deployment ran v1.0.1155, and a feature-implementer
// run failed on a bug that had already been fixed and pod-grepped green. The
// deployment pod-grep is a FALSE GREEN for that class of agent — it proves the
// image exists, not that the agent will run it.
//
// The fix removes the second authority rather than trying to keep two in step.
// The tag a spawned chassis pod should run is not a fact about a database row;
// it is a fact about the process doing the spawning. This file asks Kubernetes
// what image THIS pod is running and hands that to the child, so a rolled
// chassis reaches spawned pods with no deploy step to remember and no row to
// update. The DB columns become a fallback and a record, not the authority.
//
// Why not just sync the row at deploy time (the deploy does still sync it, as
// hygiene — see scripts/deploy/update-agent-images.sh): a deploy-time sync
// cannot survive a roll that does not go through the makefile, and this repo
// has several — `kubectl apply -k …` is written as a comment at makefile:1037,
// `kubectl set image` is scripts/deploy/deploy-agents.sh. Worse, it gets
// ROLLBACK backwards: `kubectl rollout undo` is exactly when spawned pods most
// need to follow the chassis down, and it is exactly when the makefile is not
// involved.
//
// Deliberately NOT done, and why:
//   - An env var carrying the tag. The chassis Deployment already has two —
//     AGENT_IMAGE_TAG=v1.0.82 and (via personae-prod-config) agent_image_tag=
//     v1.0.44, both measured live 2026-07-27 inside a pod running v1.0.1173.
//     Neither is read by any Go code and both are ~1,100 versions stale. A
//     duplicated tag rots here; that is evidence, not a prediction.
//   - A chassis-startup UPDATE writing the running tag back over the rows.
//     During a rolling deploy the old and new pods would fight over the column,
//     and every spawned pod runs this same binary, so the write would fan out
//     fleet-wide on every spawn. Drift is surfaced in the spawn log instead.

import (
	"context"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Where a resolved image came from — logged on every spawn so the choice is
// legible in the pod logs rather than inferred from the resulting pod spec.
const (
	imageSourceRunningChassis = "running_chassis" // this pod's own image
	imageSourceDefinition     = "agent_definition"
	imageSourcePinned         = "pinned"
)

// serviceAccountNamespaceFile is written into every in-cluster pod by the
// kubelet. It is the one namespace source that cannot be forgotten in a
// manifest, which is why it is preferred here over an env var.
const serviceAccountNamespaceFile = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

// defaultAgentNamespace matches the namespace hard-coded throughout
// spawn_actions.go; used only if the service-account file is unreadable.
const defaultAgentNamespace = "ai-persona-system"

// selfImageRetryInterval bounds how often a FAILED self-lookup is retried. A
// success is cached for the life of the process (a pod's image cannot change
// without the pod being replaced); a failure is not, or one transient API error
// at startup would silently demote the pod to the old behaviour for as long as
// it lived.
const selfImageRetryInterval = 60 * time.Second

// selfImageLookupTimeout caps the API call so a slow or unreachable API server
// cannot add latency to a spawn. On timeout the caller falls back to the row.
const selfImageLookupTimeout = 5 * time.Second

// ResolvedAgentImage is the outcome of deciding which image a spawned agent
// pod should run.
type ResolvedAgentImage struct {
	Repository string
	Tag        string
	// Source is one of the imageSource* constants.
	Source string
	// DriftedFrom carries the tag recorded on the agent_definitions row when it
	// differs from the tag actually used. Empty when the row agreed. This is
	// the 066 gap, made visible at the moment it would otherwise have bitten.
	DriftedFrom string
}

// Ref renders the reference for a container spec.
func (r ResolvedAgentImage) Ref() string {
	return r.Repository + ":" + r.Tag
}

var (
	selfImageMu      sync.Mutex
	selfImageCache   []string // container images of this pod; nil until resolved
	selfImageLastTry time.Time
)

// resolveAgentImage decides which container image a spawned agent pod should
// run, and logs the decision. Callers must use the returned values rather than
// agentDef.ImageRepository/ImageTag for anything that becomes a pod.
//
// It never fails: if this process cannot discover its own image (not in a
// cluster, RBAC denied, API slow), the agent definition's own values are used,
// which is exactly the pre-066 behaviour.
func resolveAgentImage(ctx context.Context, agentDef *AgentDefinition, logger *zap.Logger) ResolvedAgentImage {
	resolved := chooseAgentImage(selfPodImages(ctx, logger), agentDef)

	fields := []zap.Field{
		zap.String("agent_type", agentDef.Type),
		zap.String("image", resolved.Ref()),
		zap.String("image_source", resolved.Source),
	}
	if resolved.DriftedFrom != "" {
		// Not an error — this is the bug being corrected in flight. It is a
		// warning because the row is now known to be wrong, and the censuses
		// in the RUNBOOKs read that row.
		logger.Warn("bugs_open/066: agent_definitions.image_tag trails the running chassis; spawning on the running tag",
			append(fields, zap.String("row_image_tag", resolved.DriftedFrom))...)
	} else {
		logger.Info("Resolved spawn image", fields...)
	}
	return resolved
}

// chooseAgentImage is the whole decision, with no I/O, so it can be tested.
// selfImages are the container images of the pod running this process, in
// pod-spec order; it may be empty.
//
// Order:
//  1. An explicit pin on the agent row wins — that is what a pin is for.
//  2. Otherwise, if any container of this pod runs the SAME repository the row
//     asks for, that container's tag wins. Matching on repository is what keeps
//     this honest: an agent whose row names a different image (a non-chassis
//     agent, some other tool's image) is left completely alone.
//  3. Otherwise the row is used verbatim.
func chooseAgentImage(selfImages []string, agentDef *AgentDefinition) ResolvedAgentImage {
	repo := strings.TrimSpace(agentDef.ImageRepository)
	tag := strings.TrimSpace(agentDef.ImageTag)

	if agentImageIsPinned(agentDef) {
		return ResolvedAgentImage{Repository: repo, Tag: tag, Source: imageSourcePinned}
	}

	for _, img := range selfImages {
		selfRepo, selfTag, ok := parseImageRef(img)
		if !ok {
			continue
		}
		// An empty repository on the row means "whatever the chassis runs" —
		// the column is nullable and nothing else could sensibly be meant.
		if repo != "" && !sameRepository(selfRepo, repo) {
			continue
		}
		drifted := ""
		if tag != selfTag {
			drifted = tag
			if drifted == "" {
				drifted = "(null)"
			}
		}
		return ResolvedAgentImage{
			Repository:  selfRepo,
			Tag:         selfTag,
			Source:      imageSourceRunningChassis,
			DriftedFrom: drifted,
		}
	}

	return ResolvedAgentImage{Repository: repo, Tag: tag, Source: imageSourceDefinition}
}

// agentImageIsPinned reports whether the row asks to be left on its recorded
// tag. This is the deliberate override that keeps pinning possible without
// making it the default: set default_config.pin_image_tag = true on the agent
// definition and the row's image_tag is used verbatim, roll or no roll. It is
// the supported form of the interim rule 066 was filed with, and the deploy's
// row sync (scripts/deploy/update-agent-images.sh) honours the same flag, so a
// pin means one thing in both places.
func agentImageIsPinned(agentDef *AgentDefinition) bool {
	if agentDef.DefaultConfig == nil {
		return false
	}
	pinned, ok := agentDef.DefaultConfig["pin_image_tag"].(bool)
	return ok && pinned
}

// parseImageRef splits a container image reference into repository and tag.
//
// Digest references (repo@sha256:…) return ok=false deliberately: the caller
// then falls back to the row rather than inventing a tag. Nothing in this fleet
// deploys by digest, and guessing would be worse than not answering.
func parseImageRef(ref string) (repository, tag string, ok bool) {
	ref = strings.TrimSpace(ref)
	if ref == "" || strings.Contains(ref, "@") {
		return "", "", false
	}
	// A colon BEFORE the last slash is a registry port (registry:5000/x/y),
	// not a tag separator.
	lastColon := strings.LastIndex(ref, ":")
	lastSlash := strings.LastIndex(ref, "/")
	if lastColon < 0 || lastColon < lastSlash {
		// No tag: Kubernetes resolves this as :latest, so say so rather than
		// produce a reference with an empty tag.
		return ref, "latest", true
	}
	repository = ref[:lastColon]
	tag = ref[lastColon+1:]
	if repository == "" || tag == "" {
		return "", "", false
	}
	return repository, tag, true
}

// sameRepository compares two repositories the way a registry would, so that
// "aqls/agent-chassis" and "docker.io/aqls/agent-chassis" are one repository.
func sameRepository(a, b string) bool {
	return normaliseRepository(a) == normaliseRepository(b)
}

func normaliseRepository(repo string) string {
	repo = strings.TrimSpace(repo)
	repo = strings.TrimPrefix(repo, "index.docker.io/")
	repo = strings.TrimPrefix(repo, "docker.io/")
	repo = strings.TrimPrefix(repo, "library/")
	return strings.Trim(repo, "/")
}

// selfPodImages returns the container images of the pod running this process,
// cached. It returns an empty slice rather than an error: every caller's
// fallback is the same, and a spawn must not fail because the lookup did.
func selfPodImages(ctx context.Context, logger *zap.Logger) []string {
	selfImageMu.Lock()
	defer selfImageMu.Unlock()

	if selfImageCache != nil {
		return selfImageCache
	}
	if !selfImageLastTry.IsZero() && time.Since(selfImageLastTry) < selfImageRetryInterval {
		return nil
	}
	selfImageLastTry = time.Now()

	images, err := lookupSelfPodImages(ctx)
	if err != nil {
		// Info, not Warn: outside a cluster (tests, local runs) this is the
		// normal path, and what it falls back to is the old behaviour.
		logger.Info("Could not read this pod's own image; spawn images will come from agent_definitions (bugs_open/066 fallback)",
			zap.Error(err))
		return nil
	}
	selfImageCache = images
	logger.Info("Resolved this pod's own container images for spawning",
		zap.Strings("images", images))
	return images
}

func lookupSelfPodImages(ctx context.Context) ([]string, error) {
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	// A pod's hostname is its name unless a manifest overrides it, and nothing
	// here does. This avoids depending on a POD_NAME env var that a future
	// manifest could omit — the same failure mode that left AGENT_IMAGE_TAG
	// stale for a thousand versions.
	podName, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	lookupCtx, cancel := context.WithTimeout(ctx, selfImageLookupTimeout)
	defer cancel()

	pod, err := clientset.CoreV1().Pods(selfNamespace()).Get(lookupCtx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	images := make([]string, 0, len(pod.Spec.Containers))
	for _, c := range pod.Spec.Containers {
		if c.Image != "" {
			images = append(images, c.Image)
		}
	}
	return images, nil
}

func selfNamespace() string {
	if data, err := os.ReadFile(serviceAccountNamespaceFile); err == nil {
		if ns := strings.TrimSpace(string(data)); ns != "" {
			return ns
		}
	}
	if ns := strings.TrimSpace(os.Getenv("POD_NAMESPACE")); ns != "" {
		return ns
	}
	return defaultAgentNamespace
}
