// Command fuse merges ranked candidate lists (the -json output of resolve_targets
// and embed query) into one ranking, using reciprocal-rank fusion. RRF combines by
// RANK, not score, so the lexical integer scores and the semantic cosine scores —
// which aren't on a comparable scale — can be merged without normalising either.
//
//	resolve_targets -analysis a.json -task "T" -n 25 -json > lex.json
//	embed query -embeddings e.json -task "T" -n 25 -json -ollama ... -model ... > sem.json
//	fuse -in lex.json -in sem.json -n 12
//
// score(item) = sum over lists of 1/(k + rank_in_list), k=60 (standard).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"

	"contextkit/internal/candidates"
)

type inputs []string

func (i *inputs) String() string { return fmt.Sprint(*i) }
func (i *inputs) Set(v string) error {
	*i = append(*i, v)
	return nil
}

func main() {
	var in inputs
	flag.Var(&in, "in", "a candidate JSON file (repeatable)")
	n := flag.Int("n", 12, "how many fused candidates to show")
	k := flag.Int("k", 60, "RRF damping constant")
	jsonOut := flag.Bool("json", false, "emit the fused ranking as JSON")
	flag.Parse()
	if len(in) < 1 {
		fmt.Fprintln(os.Stderr, "need at least one -in file")
		os.Exit(2)
	}

	type agg struct {
		path, name, kind string
		score            float64
		ranks            map[string]int // method -> rank, for transparency
	}
	byKey := map[string]*agg{}
	var methods []string
	task := ""
	for _, path := range in {
		raw, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read %s: %v\n", path, err)
			os.Exit(1)
		}
		var cf candidates.File
		if err := json.Unmarshal(raw, &cf); err != nil {
			fmt.Fprintf(os.Stderr, "parse %s: %v\n", path, err)
			os.Exit(1)
		}
		if task == "" {
			task = cf.Task
		}
		method := cf.Method
		if method == "" {
			method = path
		}
		methods = append(methods, method)
		for _, c := range cf.Candidates {
			key := c.Path + ":" + c.Name
			a := byKey[key]
			if a == nil {
				a = &agg{c.Path, c.Name, c.Kind, 0, map[string]int{}}
				byKey[key] = a
			}
			a.score += 1.0 / float64(*k+c.Rank)
			a.ranks[method] = c.Rank
		}
	}

	fused := make([]*agg, 0, len(byKey))
	for _, a := range byKey {
		fused = append(fused, a)
	}
	sort.SliceStable(fused, func(i, j int) bool {
		if fused[i].score != fused[j].score {
			return fused[i].score > fused[j].score
		}
		return fused[i].path < fused[j].path
	})

	lim := *n
	if lim > len(fused) {
		lim = len(fused)
	}

	if *jsonOut {
		out := candidates.File{Task: task, Method: "fused"}
		for i := 0; i < lim; i++ {
			out.Candidates = append(out.Candidates, candidates.Candidate{
				Path: fused[i].path, Name: fused[i].name, Kind: fused[i].kind,
				Score: fused[i].score, Rank: i + 1,
			})
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(b))
		return
	}

	fmt.Printf("Fused top %d (RRF over %d lists: %v) for task:\n  %q\n\n", lim, len(in), methods, task)
	for i := 0; i < lim; i++ {
		a := fused[i]
		fmt.Printf("  %.4f  %s:%s (%s)  ranks=%v\n", a.score, a.path, a.name, a.kind, a.ranks)
	}
	fmt.Printf("\nSuggested scope (top symbols):\n")
	for i := 0; i < lim && i < 4; i++ {
		fmt.Printf("  -scope %s:%s\n", fused[i].path, fused[i].name)
	}
}
