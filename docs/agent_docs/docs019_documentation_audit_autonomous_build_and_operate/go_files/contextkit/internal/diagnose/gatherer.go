// gatherer.go — a Gatherer that shells out to cmd/bundle (the read-only
// gather+assemble wrapper) to produce each iteration's bundle. It translates a
// Scope into bundle flags and returns the written bundle path.
//
// READ-ONLY by construction (DESIGN §4): bundle runs dbcontext (\d / capped
// SELECT / existing-log read) + the pure assembler. Nothing here triggers a
// build, spawn, or write to the system under diagnosis. This adapter adds no
// new capability beyond what cmd/bundle already does — it just drives it per
// iteration with the loop's evolving scope.
package diagnose

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

// BundleGatherer drives cmd/bundle. Field values map to the bundle/assembler
// inputs that DON'T change across iterations (analysis, root, constitution,
// psql); the per-iteration Scope supplies the rest.
type BundleGatherer struct {
	BundleBin    string   // how to invoke bundle, e.g. "./cmd/bundle" (go run target)
	UseGoRun     bool     // true => `go run <BundleBin> …`; false => exec BundleBin directly
	AnalysisPath string   // -analysis
	Root         string   // -root
	Constitution string   // -constitution
	Docs         []string // -doc (repeatable): authored context pasted verbatim into every
	//                        bundle's "Reference documents" section — e.g. the relevant 016
	//                        §9 entry or a dev-guide section. Constant across iterations
	//                        (the per-iteration evidence is the code bodies + runtime), so the
	//                        verdicter judges against the SAME standing context each time.
	DocRules []DocRule // per-hypothesis doc selection: each iteration ALSO includes the
	//                     docs whose keywords/paths match the CURRENT hypothesis/scope
	//                     (the task-specific 003 contracts), via SelectDocs. The 003 contract
	//                     for component linkage rides in on a component-linkage hypothesis, the
	//                     CSS contract on a CSS hypothesis, neither otherwise — so the bundle
	//                     stays focused as the loop re-scopes. Loaded from a JSON catalogue.
	Psql   string // -psql (empty => bundle skips the DB gather)
	// SchemaTables are domain tables whose \d is included in EVERY bundle
	// (independent of the per-iteration scope), so the verdict can write
	// data_requests against confirmed columns rather than guessing — the
	// "confirm schema" step. Constant across iterations, like Docs; merged with
	// the scope's per-iteration Tables. Only takes effect when Psql is set (\d
	// needs a connection). Reuses the existing -schema-tables -> dbcontext -schema
	// path; no new mechanism.
	SchemaTables []string
	Step   string // -step (default "debug")
	OutDir string // where per-iteration bundles are written
	iter   int    // internal counter for output filenames
	// DryRun: if set, build the command and write the command line to the output
	// path instead of running bundle — lets the wiring be tested without a cluster.
	DryRun bool
}

// Gather assembles the bundle for this scope and returns its path. The
// hypothesis is forwarded to cmd/bundle as its required -task (the one-sentence
// framing the bundle is built around) — bundle rejects args without it.
func (b *BundleGatherer) Gather(hypothesis string, s Scope) (string, error) {
	b.iter++
	if b.Step == "" {
		b.Step = "debug"
	}
	if b.OutDir == "" {
		b.OutDir = os.TempDir()
	}
	outPath := filepath.Join(b.OutDir, "diag_bundle_"+strconv.Itoa(b.iter)+".md")

	args := b.buildArgs(hypothesis, s, outPath)

	if b.DryRun {
		// Write the command we WOULD run, so tests/inspection can assert the
		// scope→flags translation without a cluster.
		line := "go " + join(args)
		if !b.UseGoRun {
			line = b.BundleBin + " " + join(args[2:]) // strip "run <bin>"
		}
		if err := os.WriteFile(outPath, []byte(line+"\n"), 0o644); err != nil {
			return "", err
		}
		return outPath, nil
	}

	var cmd *exec.Cmd
	if b.UseGoRun {
		cmd = exec.Command("go", args...)
	} else {
		cmd = exec.Command(b.BundleBin, args[2:]...) // args[0:2] are "run <bin>"
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("bundle gather (iter %d): %w", b.iter, err)
	}
	return outPath, nil
}

// buildArgs translates the fixed config + the per-iteration Scope into the
// bundle argv. Mirrors cmd/bundle's own flag names exactly — including -task,
// which bundle REQUIRES (main.go: "required: -analysis -root -constitution -task").
func (b *BundleGatherer) buildArgs(hypothesis string, s Scope, outPath string) []string {
	args := []string{"run", b.BundleBin,
		"-analysis", b.AnalysisPath,
		"-root", b.Root,
		"-constitution", b.Constitution,
		"-task", hypothesis,
		"-step", b.Step,
		"-out", outPath,
	}
	for _, sym := range s.Symbols {
		args = append(args, "-scope", sym)
	}
	// Authored context: the always-include Docs plus the per-hypothesis matches
	// from DocRules (the 003 contract / 016 §9 entry that fits THIS hypothesis or
	// scope). De-duplicated so an always-include doc that also matches a rule
	// isn't pasted twice. Each goes to the assembler's verbatim "Reference
	// documents" section.
	docSeen := map[string]bool{}
	emitDoc := func(d string) {
		if d != "" && !docSeen[d] {
			docSeen[d] = true
			args = append(args, "-doc", d)
		}
	}
	for _, d := range b.Docs {
		emitDoc(d)
	}
	for _, d := range SelectDocs(hypothesis, s, b.DocRules) {
		emitDoc(d)
	}
	if b.Psql != "" {
		args = append(args, "-psql", b.Psql)
		// Merge the constant domain-schema tables (so the model always sees the
		// columns it may query) with this iteration's scope tables, de-duplicated.
		tblSeen := map[string]bool{}
		var schemaTables []string
		for _, t := range append(append([]string{}, b.SchemaTables...), s.Tables...) {
			if t != "" && !tblSeen[t] {
				tblSeen[t] = true
				schemaTables = append(schemaTables, t)
			}
		}
		if len(schemaTables) > 0 {
			args = append(args, "-schema-tables", join2(schemaTables, ","))
		}
		if s.RuntimeSite != "" {
			args = append(args, "-runtime-site", s.RuntimeSite)
			if s.RuntimePage != "" {
				args = append(args, "-runtime-page", s.RuntimePage)
			}
		}
		if s.Capabilities {
			args = append(args, "-capabilities")
		}
	}
	return args
}

func join(a []string) string { return join2(a, " ") }
func join2(a []string, sep string) string {
	out := ""
	for i, s := range a {
		if i > 0 {
			out += sep
		}
		out += s
	}
	return out
}
