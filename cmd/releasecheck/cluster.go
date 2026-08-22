// FILE: cmd/releasecheck/cluster.go
//
// The one part of the census that must talk to Kubernetes.
//
// It lives HERE and not in pkg/releaseset deliberately, and the split is the
// same one the rest of this tool makes: judgement in the package, I/O in the
// command. `pkg/releaseset` must not import client-go, because a predicate that
// can only be exercised against a live cluster is a predicate nobody exercises —
// the whole census is table-tested against a []Workload with no cluster in
// sight, and this file is the only thing that would need one.
//
// It reads Deployments, CronJobs and DaemonSets, because all three carry one of
// our images today: Deployments for the services, CronJobs for the daily checks,
// a DaemonSet for `node-config` (BLD-021). Reading only Deployments would have
// missed every frozen check service, which is most of bugs_open/318's evidence.
//
// LIST ONLY. Nothing here mutates, execs, or reads a Secret.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/gqls/agentchassis/pkg/releaseset"
)

// k8sClient returns a clientset for in-cluster use OR for a human at a terminal.
//
// Both paths matter and neither is a fallback for a failure of the other:
// in-cluster is how a future CronJob would run this, and kubeconfig is how it is
// run by hand today. Trying in-cluster first is right — `rest.InClusterConfig`
// fails fast and unambiguously when the service-account files are absent, so it
// can never half-succeed against the wrong cluster.
func k8sClient() (*kubernetes.Clientset, string, error) {
	if cfg, err := rest.InClusterConfig(); err == nil {
		cs, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			return nil, "", fmt.Errorf("in-cluster client: %w", err)
		}
		return cs, "in-cluster service account", nil
	}

	path := os.Getenv("KUBECONFIG")
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, "", fmt.Errorf("no KUBECONFIG and no home directory: %w", err)
		}
		path = filepath.Join(home, ".kube", "config")
	}
	cfg, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, "", fmt.Errorf("reading kubeconfig %s: %w", path, err)
	}
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, "", fmt.Errorf("kubeconfig client: %w", err)
	}
	return cs, "kubeconfig " + path, nil
}

// readWorkloads lists everything in the namespace that carries an image.
//
// ⚠ A LIST ERROR IS RETURNED, NEVER SWALLOWED. If the CronJob list fails and the
// Deployment list succeeds, silently returning the Deployments would report a
// clean fleet while blind to every scheduled check — which is the exact
// population bugs_open/318 is about. A partial read is not a read.
func readWorkloads(ctx context.Context, cs *kubernetes.Clientset, namespace string) ([]releaseset.Workload, error) {
	var out []releaseset.Workload

	deps, err := cs.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing deployments in %s: %w", namespace, err)
	}
	for _, d := range deps.Items {
		for _, c := range d.Spec.Template.Spec.Containers {
			repo, tag := releaseset.SplitImageTag(c.Image)
			out = append(out, releaseset.Workload{Name: d.Name, Kind: "Deployment", Image: repo, Tag: tag})
			break // the first container is the service; sidecars are not the subject
		}
	}

	crons, err := cs.BatchV1().CronJobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing cronjobs in %s: %w", namespace, err)
	}
	for _, cj := range crons.Items {
		for _, c := range cj.Spec.JobTemplate.Spec.Template.Spec.Containers {
			repo, tag := releaseset.SplitImageTag(c.Image)
			out = append(out, releaseset.Workload{Name: cj.Name, Kind: "CronJob", Image: repo, Tag: tag})
			break
		}
	}

	dss, err := cs.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing daemonsets in %s: %w", namespace, err)
	}
	for _, ds := range dss.Items {
		for _, c := range ds.Spec.Template.Spec.Containers {
			repo, tag := releaseset.SplitImageTag(c.Image)
			out = append(out, releaseset.Workload{Name: ds.Name, Kind: "DaemonSet", Image: repo, Tag: tag})
			break
		}
	}
	return out, nil
}

// runCensus is the --census mode.
func runCensus(root, registry, namespace string) (int, error) {
	mf, err := os.Open(root + "/makefile")
	if err != nil {
		return exitCannotRun, fmt.Errorf("opening %s/makefile: %w", root, err)
	}
	defer mf.Close()
	decl, err := releaseset.ParseMakefileDecls(mf)
	if err != nil {
		return exitCannotRun, err
	}

	cs, via, err := k8sClient()
	if err != nil {
		return exitCannotRun, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	workloads, err := readWorkloads(ctx, cs, namespace)
	if err != nil {
		return exitCannotRun, err
	}

	res, err := releaseset.Census(decl, workloads, registry)
	if err != nil {
		return exitCannotRun, err
	}

	// Print the measurement FIRST and always. A report that does not say how
	// much it looked at cannot be told apart from one that looked at nothing,
	// and for a detector that is the difference that matters.
	fmt.Printf("%sRelease census%s — namespace %s, via %s: %d of %d workloads run a %s/ image; "+
		"the fleet is on %s.\n",
		green, reset, namespace, via, res.Examined, res.Total, registry, res.FleetTag)

	if len(res.Findings) == 0 {
		fmt.Printf("%s  Nothing to report: every workload of ours is on the fleet tag, every declared "+
			"service is running, and every running service of ours is declared.%s\n", dim, reset)
		return exitOK, nil
	}

	fmt.Fprintf(os.Stderr, "%s%d finding(s) (bugs_open/318, register BLD-026)%s\n", yellow, len(res.Findings), reset)
	for _, v := range res.Findings {
		fmt.Fprintf(os.Stderr, "%s  %s%s\n", yellow, v.String(), reset)
		fmt.Fprintf(os.Stderr, "%s      %s%s\n", dim, v.Remedy, reset)
	}
	return exitViolation, nil
}
