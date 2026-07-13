// callgraph.go — a CallGraph backed by the analyser's recorded `calls`, so the
// diagnosis loop can re-scope by FOLLOWING the call graph from an evidence-named
// site (DESIGN §1a) rather than re-searching the symptom (the B4a ceiling).
//
// HONEST LIMIT (from the analyser): `calls` is NAME-BASED, not type-resolved —
// `x.Bar()` and `pkg.Bar()` both record "Bar" (analyse.go callsIn). So a callee
// NAME can map to MANY defining symbols. The adapter resolves names to symbols
// and DELIBERATELY DROPS ubiquitous names (Run, String, Error, …) that would
// explode the neighbourhood into noise — following them is worse than useless,
// it's the symptom-vocabulary trap in call-graph form. The loop's narrowing
// guard is the backstop, but dropping known-ubiquitous names keeps re-scope
// sharp at the source.
package diagnose

import (
	"encoding/json"
	"os"
	"sort"
	"strings"

	"contextkit/internal/analysis"
)

// AnalysisCallGraph implements CallGraph over an analyser Output.
type AnalysisCallGraph struct {
	// symbol "path:Name" -> the callee NAMES in its body
	callsBySym map[string][]string
	// callee NAME -> the symbols ("path:Name") that DEFINE a function of that name
	defsByName map[string][]string
	// names too common to follow (resolved to > maxDefsPerName symbols, or in the
	// stop-list) — following them adds noise, not signal
	ubiquitous map[string]bool
	maxFanout  int
}

// ubiquitousStopList: names so common across a Go codebase that following them
// is noise. Mirrors the kind of vocabulary that flooded the B4a semantic top-12.
var ubiquitousStopList = map[string]bool{
	"Run": true, "String": true, "Error": true, "Errorf": true, "Sprintf": true,
	"Info": true, "Warn": true, "Debug": true, "Error_": true, "Close": true,
	"len": true, "make": true, "append": true, "Background": true, "New": true,
	"Marshal": true, "Unmarshal": true, "Get": true, "Set": true, "Now": true,
	"WithCancel": true, "Load": true, "Parse": true, "Scan": true, "Next": true,
}

// NewCallGraphFromFile loads an analyser JSON and builds the call graph.
func NewCallGraphFromFile(path string) (*AnalysisCallGraph, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out analysis.Output
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return NewCallGraph(out), nil
}

// NewCallGraph builds the graph from an in-memory analysis.
func NewCallGraph(out analysis.Output) *AnalysisCallGraph {
	g := &AnalysisCallGraph{
		callsBySym: map[string][]string{},
		defsByName: map[string][]string{},
		ubiquitous: map[string]bool{},
		maxFanout:  8, // a callee name resolving to more than this = too ambiguous to follow
	}
	for _, f := range out.Files {
		for _, fn := range f.Functions {
			sym := f.Path + ":" + fn.Name
			g.callsBySym[sym] = fn.Calls
			g.defsByName[fn.Name] = append(g.defsByName[fn.Name], sym)
		}
	}
	for name := range ubiquitousStopList {
		g.ubiquitous[name] = true
	}
	// also mark names that resolve to too many definitions as ubiquitous
	for name, defs := range g.defsByName {
		if len(defs) > g.maxFanout {
			g.ubiquitous[name] = true
		}
	}
	for _, defs := range g.defsByName {
		sort.Strings(defs)
	}
	return g
}

// Neighbourhood returns the symbols reachable in ONE hop from sym: the symbols
// that DEFINE the callees in sym's body. Ubiquitous callee names are dropped.
// Accepts "path:Name" (preferred) or a bare "Name" (resolves all defs of that
// name as the starting set, then their callees).
func (g *AnalysisCallGraph) Neighbourhood(sym string) []string {
	var startSyms []string
	if strings.Contains(sym, ":") {
		startSyms = []string{sym}
	} else {
		// bare name: start from every symbol defining it
		startSyms = g.defsByName[sym]
	}

	out := map[string]bool{}
	for _, s := range startSyms {
		for _, callee := range g.callsBySym[s] {
			if g.ubiquitous[callee] {
				continue // drop noise — do not follow ubiquitous names
			}
			for _, def := range g.defsByName[callee] {
				if def != s { // don't include self
					out[def] = true
				}
			}
		}
	}
	res := make([]string, 0, len(out))
	for s := range out {
		res = append(res, s)
	}
	sort.Strings(res)
	return res
}

// CalleesOf is a lower-level helper (returns the raw callee NAMES of a symbol,
// ubiquitous ones included) — useful for diagnostics/tests, not used by the loop.
func (g *AnalysisCallGraph) CalleesOf(sym string) []string {
	return g.callsBySym[sym]
}
