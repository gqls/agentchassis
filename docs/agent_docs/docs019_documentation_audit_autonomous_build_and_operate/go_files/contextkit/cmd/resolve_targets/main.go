// Command resolve_targets is a first-cut target resolver: given a task string
// and the analyser JSON, it proposes which symbols/files to -scope, by lexical
// overlap between the task's distinctive words and each symbol's name, path, and
// docstring. This is the deterministic baseline — the layer that runs before any
// embeddings. It does not decide; it proposes a ranked candidate set for a human
// (or the assembler) to confirm.
//
//	resolve_targets -analysis analysis.json -task "why does plan_sections see no ready sections" -n 12
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"contextkit/internal/analysis"
	"contextkit/internal/candidates"
)

// analysis.Output is the analyser contract (internal/analysis).

type cand struct {
	path, name, kind string
	score            int
	matched          map[string]bool
}

var stop = map[string]bool{
	"the": true, "and": true, "for": true, "how": true, "why": true, "what": true,
	"is": true, "are": true, "was": true, "does": true, "do": true, "did": true,
	"find": true, "from": true, "that": true, "this": true, "with": true, "into": true,
	"its": true, "it": true, "a": true, "an": true, "of": true, "to": true, "in": true,
	"on": true, "no": true, "not": true, "any": true, "all": true, "actual": true,
	"see": true, "sees": true, "reads": true, "read": true, "computed": true, "compute": true,
	"source": true, "where": true, "which": true, "when": true, "then": true,
}

// splitWords breaks an identifier or phrase into lowercase word tokens,
// splitting on camelCase, digits, and non-alphanumerics.
func splitWords(s string) []string {
	var words []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			words = append(words, strings.ToLower(cur.String()))
			cur.Reset()
		}
	}
	runes := []rune(s)
	for i, r := range runes {
		isUpper := r >= 'A' && r <= 'Z'
		isLetter := isUpper || (r >= 'a' && r <= 'z')
		isDigit := r >= '0' && r <= '9'
		if !isLetter && !isDigit {
			flush()
			continue
		}
		// boundary: upper after lower/digit, or upper followed by lower (acronym end)
		if isUpper && cur.Len() > 0 {
			prev := runes[i-1]
			prevLowerOrDigit := (prev >= 'a' && prev <= 'z') || (prev >= '0' && prev <= '9')
			nextLower := i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z'
			if prevLowerOrDigit || nextLower {
				flush()
			}
		}
		cur.WriteRune(r)
	}
	flush()
	return words
}

func taskTokens(task string) []string {
	seen := map[string]bool{}
	var out []string
	for _, w := range splitWords(task) {
		if len(w) < 3 || stop[w] {
			continue
		}
		if !seen[w] {
			seen[w] = true
			out = append(out, w)
		}
	}
	return out
}

// fileWords are the meaningful words in a file basename (minus .go and an
// _action/_actions suffix), used so a file's name contributes to its symbols.
func fileWords(path string) []string {
	base := path
	if i := strings.LastIndexByte(base, '/'); i >= 0 {
		base = base[i+1:]
	}
	base = strings.TrimSuffix(base, ".go")
	base = strings.TrimSuffix(base, "_actions")
	base = strings.TrimSuffix(base, "_action")
	return splitWords(base)
}

func main() {
	analysisPath := flag.String("analysis", "", "path to the analyser JSON")
	task := flag.String("task", "", "the task description")
	n := flag.Int("n", 12, "how many candidates to show")
	jsonOut := flag.Bool("json", false, "emit ranked candidates as JSON (for fuse/eval)")
	flag.Parse()
	if *analysisPath == "" || *task == "" {
		fmt.Fprintln(os.Stderr, "need -analysis and -task")
		os.Exit(2)
	}
	raw, err := os.ReadFile(*analysisPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read analysis: %v\n", err)
		os.Exit(1)
	}
	var an analysis.Output
	if err := json.Unmarshal(raw, &an); err != nil {
		fmt.Fprintf(os.Stderr, "parse analysis: %v\n", err)
		os.Exit(1)
	}

	tokens := taskTokens(*task)
	tokenSet := map[string]bool{}
	for _, t := range tokens {
		tokenSet[t] = true
	}

	score := func(name, path, doc string) (int, map[string]bool) {
		matched := map[string]bool{}
		s := 0
		nameWords := map[string]bool{}
		for _, w := range splitWords(name) {
			nameWords[w] = true
		}
		fileWordSet := map[string]bool{}
		for _, w := range fileWords(path) {
			fileWordSet[w] = true
		}
		lowName := strings.ToLower(name)
		docWords := map[string]bool{}
		for _, w := range splitWords(doc) {
			docWords[w] = true
		}
		for t := range tokenSet {
			switch {
			case nameWords[t]:
				s += 3
				matched[t] = true
			case strings.Contains(lowName, t):
				s += 1
				matched[t] = true
			}
			if fileWordSet[t] {
				s += 2
				matched[t] = true
			}
			if docWords[t] {
				s += 1
				matched[t] = true
			}
		}
		return s, matched
	}

	var cands []cand
	includes := map[string]bool{}
	for _, f := range an.Files {
		// flag wiring/registration files for -include (call graph can't reach them)
		fw := fileWords(f.Path)
		for _, w := range fw {
			if w == "registry" || w == "register" || w == "registration" {
				includes[f.Path] = true
			}
		}
		for _, fn := range f.Functions {
			if fn.Name == "init" {
				continue
			}
			s, m := score(fn.Name, f.Path, fn.Doc)
			if s > 0 {
				cands = append(cands, cand{f.Path, fn.Name, "func", s, m})
			}
		}
		for _, td := range f.Types {
			s, m := score(td.Name, f.Path, td.Doc)
			if s > 0 {
				cands = append(cands, cand{f.Path, td.Name, td.Kind, s, m})
			}
		}
	}
	sort.SliceStable(cands, func(i, j int) bool {
		if cands[i].score != cands[j].score {
			return cands[i].score > cands[j].score
		}
		return cands[i].path < cands[j].path
	})

	if *jsonOut {
		lim := *n
		if lim > len(cands) {
			lim = len(cands)
		}
		out := candidates.File{Task: *task, Method: "lexical"}
		for i := 0; i < lim; i++ {
			out.Candidates = append(out.Candidates, candidates.Candidate{
				Path: cands[i].path, Name: cands[i].name, Kind: cands[i].kind,
				Score: float64(cands[i].score), Rank: i + 1,
			})
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(b))
		return
	}

	fmt.Printf("Task tokens: %s\n\n", strings.Join(tokens, ", "))
	if len(cands) == 0 {
		fmt.Println("No lexical matches. Broaden the task wording, or this area may use names that don't echo the task (the case for embeddings).")
		return
	}

	fmt.Printf("Top %d candidate symbols (score — path:symbol — matched):\n", *n)
	limit := *n
	if limit > len(cands) {
		limit = len(cands)
	}
	// distinct files among the top candidates, in first-seen order
	var topFiles []string
	seenFile := map[string]bool{}
	for i := 0; i < limit; i++ {
		c := cands[i]
		var mk []string
		for k := range c.matched {
			mk = append(mk, k)
		}
		sort.Strings(mk)
		fmt.Printf("  %2d  %s:%s (%s)  [%s]\n", c.score, c.path, c.name, c.kind, strings.Join(mk, " "))
		if !seenFile[c.path] {
			seenFile[c.path] = true
			topFiles = append(topFiles, c.path)
		}
	}

	fmt.Printf("\nSuggested scope (top symbols):\n")
	for i := 0; i < limit && i < 4; i++ {
		fmt.Printf("  -scope %s:%s\n", cands[i].path, cands[i].name)
	}
	if len(includes) > 0 {
		var incs []string
		for p := range includes {
			incs = append(incs, p)
		}
		sort.Strings(incs)
		fmt.Printf("\nSuggested wiring includes (call graph can't reach these):\n")
		for _, p := range incs {
			fmt.Printf("  -include %s\n", p)
		}
	}
	fmt.Printf("\n(Deterministic lexical match only — confirm before use; symbols that don't echo the task in their names won't appear. That recall gap is what embeddings would later fill.)\n")
}
