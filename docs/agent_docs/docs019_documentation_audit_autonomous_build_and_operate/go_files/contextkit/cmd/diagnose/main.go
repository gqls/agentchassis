// Command diagnose wires the diagnosis-loop scaffold (internal/diagnose) to its
// real adapters: the BundleGatherer (shells out to cmd/bundle, read-only) and the
// AnalysisCallGraph (follows the analyser's `calls` for re-scope, DESIGN §1a).
//
// THE VERDICT STEP IS NOT THE REAL MODEL HERE. The cite-or-abstain LLM verdicter
// (DESIGN §2) is a chassis-side follow-on (needs a model). This entrypoint ships
// two stand-ins so the loop is runnable + inspectable now:
//
//	-verdict-script FILE : a JSON array of verdicts, one per iteration (scripted
//	                       testing — drive the loop through a known reasoning path)
//	(default)            : a trivial stub that always returns UNVERIFIABLE, so a
//	                       real run just exercises the gather+guard plumbing and
//	                       stops at the iteration cap (no fabricated conclusions).
//
// READ-ONLY + HUMAN-GATED (DESIGN §4): emits a diagnosis + evidence trail; never
// a fix, never a triggered run.
//
// Usage (scripted, no cluster — uses bundle dry-run):
//
//	diagnose -analysis a.json -root /repo -constitution c.md \
//	         -seed-hypothesis "…symptom…" -seed-scope save_page_sections_action.go:SavePageSectionsAction \
//	         -callgraph a.json -verdict-script verdicts.json -dry-bundle
//
// Usage (live gather, real model not yet wired — will stop at cap):
//
//	diagnose -analysis a.json -root /repo -constitution c.md -psql '…' \
//	         -seed-hypothesis "…" -seed-scope file.go:Sym -callgraph a.json
package main

import (
	"fmt"
	"os"
	"strings"

	"contextkit/internal/diagnose"
)

func main() {
	var (
		analysisPath, root, constitution, psql  string
		seedHypothesis, seedScopeCSV            string
		seedTablesCSV, runtimeSite, runtimePage string
		schemaTablesCSV                         string
		capabilities                            bool
		callgraphPath                           string
		verdictScript                           string
		dryBundle                               bool
		maxIter                                 int
		bundleBin                               string
		noFollow                                bool
		docs                                    []string
		docCatalogue                            string
	)
	bundleBin = "./cmd/bundle"
	maxIter = 5

	args := os.Args[1:]
	need := func(i *int) string {
		if *i+1 >= len(args) {
			fmt.Fprintf(os.Stderr, "flag %s needs a value\n", args[*i])
			os.Exit(2)
		}
		*i++
		return args[*i]
	}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-analysis":
			analysisPath = need(&i)
		case "-root":
			root = need(&i)
		case "-constitution":
			constitution = need(&i)
		case "-psql":
			psql = need(&i)
		case "-seed-hypothesis":
			seedHypothesis = need(&i)
		case "-seed-scope":
			seedScopeCSV = need(&i)
		case "-seed-tables":
			seedTablesCSV = need(&i)
		case "-schema-tables":
			schemaTablesCSV = need(&i) // constant domain schema \d'd into every bundle (for data_requests)
		case "-runtime-site":
			runtimeSite = need(&i)
		case "-runtime-page":
			runtimePage = need(&i)
		case "-capabilities":
			capabilities = true
		case "-callgraph":
			callgraphPath = need(&i)
		case "-doc":
			docs = append(docs, need(&i)) // authored context pasted verbatim into every bundle
		case "-doc-catalogue":
			docCatalogue = need(&i) // JSON catalogue; per-hypothesis doc selection
		case "-verdict-script":
			verdictScript = need(&i)
		case "-dry-bundle":
			dryBundle = true
		case "-max-iter":
			fmt.Sscanf(need(&i), "%d", &maxIter)
		case "-bundle-bin":
			bundleBin = need(&i)
		case "-no-follow":
			noFollow = true
		case "-h", "-help", "--help":
			usage()
			return
		default:
			fmt.Fprintf(os.Stderr, "unknown flag: %s\n", args[i])
			usage()
			os.Exit(2)
		}
	}
	if analysisPath == "" || root == "" || constitution == "" || seedHypothesis == "" || seedScopeCSV == "" {
		fmt.Fprintln(os.Stderr, "required: -analysis -root -constitution -seed-hypothesis -seed-scope")
		usage()
		os.Exit(2)
	}

	// Gatherer (read-only bundle wrapper).
	docRules, err := diagnose.LoadDocCatalogue(docCatalogue)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load doc catalogue: %v\n", err)
		os.Exit(2)
	}
	g := &diagnose.BundleGatherer{
		BundleBin: bundleBin, UseGoRun: true,
		AnalysisPath: analysisPath, Root: root, Constitution: constitution,
		Docs: docs, DocRules: docRules,
		SchemaTables: splitCSV(schemaTablesCSV),
		Psql:         psql, Step: "debug", DryRun: dryBundle,
	}

	// CallGraph (re-scope by following calls).
	var cg diagnose.CallGraph
	if callgraphPath != "" && !noFollow {
		acg, err := diagnose.NewCallGraphFromFile(callgraphPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "load callgraph: %v\n", err)
			os.Exit(1)
		}
		cg = acg
	}

	// Verdicter: scripted, or the trivial UNVERIFIABLE stub.
	var v diagnose.Verdicter
	if verdictScript != "" {
		sv, err := loadScript(verdictScript)
		if err != nil {
			fmt.Fprintf(os.Stderr, "load verdict script: %v\n", err)
			os.Exit(1)
		}
		v = sv
	} else {
		v = stubVerdicter{}
		fmt.Fprintln(os.Stderr, "NOTE: no -verdict-script; using the UNVERIFIABLE stub. The real model verdicter is a chassis-side follow-on. The loop will exercise gather+guards and stop at the cap without concluding.")
	}

	seed := diagnose.Scope{
		Symbols:      splitCSV(seedScopeCSV),
		Tables:       splitCSV(seedTablesCSV),
		RuntimeSite:  runtimeSite,
		RuntimePage:  runtimePage,
		Capabilities: capabilities,
	}
	cfg := diagnose.Config{MaxIterations: maxIter, FollowCallGraph: !noFollow}

	res, err := diagnose.Run(seedHypothesis, seed, g, v, cg, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "loop error: %v\n", err)
		os.Exit(1)
	}
	printResult(res)
}

// stubVerdicter always abstains — so a model-less run never fabricates a verdict.
type stubVerdicter struct{}

func (stubVerdicter) Assess(_, _ string) (diagnose.Verdict, error) {
	return diagnose.Verdict{Outcome: diagnose.Unverifiable,
		NeededEvidence: "no model wired (stub verdicter); supply -verdict-script or the chassis model"}, nil
}

// scriptVerdicter replays verdicts from a JSON file, one per iteration.
type scriptVerdicter struct {
	verdicts []diagnose.Verdict
	i        int
}

func (s *scriptVerdicter) Assess(_, _ string) (diagnose.Verdict, error) {
	if s.i >= len(s.verdicts) {
		return diagnose.Verdict{Outcome: diagnose.Unverifiable, NeededEvidence: "script exhausted"}, nil
	}
	v := s.verdicts[s.i]
	s.i++
	return v, nil
}

func loadScript(path string) (*scriptVerdicter, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// The script is the MODEL WIRE FORMAT (string outcomes/tiers) — the same
	// bytes the real model would emit — so a script is a faithful model stand-in.
	vs, err := diagnose.ParseVerdicts(raw)
	if err != nil {
		return nil, err
	}
	return &scriptVerdicter{verdicts: vs}, nil
}

func printResult(r diagnose.Result) {
	fmt.Println("================ DIAGNOSIS ================")
	fmt.Println(r.Conclusion)
	fmt.Printf("\nstopped by: %s   |   iterations: %d\n", r.StoppedBy, len(r.Trail))
	fmt.Println("\n---------------- EVIDENCE TRAIL ----------------")
	for _, s := range r.Trail {
		fmt.Printf("\n[iter %d] hypothesis: %s\n", s.Iteration, s.Hypothesis)
		fmt.Printf("  scope: %s\n", strings.Join(s.Scope.Symbols, ", "))
		fmt.Printf("  bundle: %s\n", s.BundlePath)
		fmt.Printf("  verdict: %s\n", s.Verdict.Outcome)
		for _, c := range s.Verdict.Citations {
			fmt.Printf("    [%s] %s — %q\n", c.Tier, c.Where, c.Quote)
		}
		if s.Verdict.RevisedHypothesis != "" {
			fmt.Printf("    → revised: %s\n", s.Verdict.RevisedHypothesis)
		}
		if len(s.Verdict.NextScope) > 0 {
			fmt.Printf("    → next scope: %s\n", strings.Join(s.Verdict.NextScope, ", "))
		}
		if s.Verdict.NeededEvidence != "" {
			fmt.Printf("    → needed: %s\n", s.Verdict.NeededEvidence)
		}
	}
	fmt.Println("\n(diagnosis + trail for a HUMAN; the loop never applies a fix)")
}

func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func usage() {
	fmt.Fprint(os.Stderr, `diagnose — run the read-only diagnosis loop (scaffold + adapters)

REQUIRED:
  -analysis FILE        analyser JSON (for bundle)
  -root DIR             repo root
  -constitution FILE    flat constitution
  -seed-hypothesis STR  the starting hypothesis (the symptom)
  -seed-scope CSV       initial scope symbols (path or path:Symbol), comma-separated

OPTIONAL:
  -psql STR             psql invocation (omit => bundle skips DB gather)
  -seed-tables CSV      initial -schema-tables
  -runtime-site DOMAIN  runtime evidence site   -runtime-page NAME
  -capabilities         include \dx/\df
  -callgraph FILE       analyser JSON for call-graph re-scope (DESIGN §1a)
  -verdict-script FILE  JSON array of verdicts, one per iteration (scripted test)
  -dry-bundle           bundle writes the command it WOULD run (no cluster needed)
  -max-iter N           iteration cap (default 5)
  -no-follow            disable call-graph following (use verdict NextScope verbatim)

Without -verdict-script, a stub verdicter abstains every iteration (no model
wired; the real cite-or-abstain model is a chassis-side follow-on).
`)
}
