// Command eval_targets scores a resolver's candidate list (the -json output of
// resolve_targets, embed query, or fuse) against a ground-truth set of tasks mapped
// to the symbols they actually needed. It turns "the semantic/fused list looks
// better" into numbers: recall@N over the decisive symbols, and the rank of the
// first decisive hit (MRR contribution). Run it on each method's output for the
// same task to compare them.
//
//	eval_targets -truth groundtruth_targets.json -candidates lex.json -n 12
//	eval_targets -truth groundtruth_targets.json -candidates fused.json -n 12
//
// Match is on "path:name". The ground-truth task id is matched to the candidate
// file by -task-id (or, if omitted, the single task in the truth file is used).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"contextkit/internal/candidates"
)

type truth struct {
	Tasks []struct {
		ID     string   `json:"id"`
		Task   string   `json:"task"`
		Expect []string `json:"expect"`      // decisive — must be found
		Useful []string `json:"also_useful"` // helpful but not scored as decisive
	} `json:"tasks"`
}

func main() {
	truthPath := flag.String("truth", "groundtruth_targets.json", "ground-truth file")
	candPath := flag.String("candidates", "", "candidate JSON (from resolve_targets/embed/fuse -json)")
	taskID := flag.String("task-id", "", "which ground-truth task id (default: the only one)")
	n := flag.Int("n", 12, "evaluate recall within the top N candidates")
	flag.Parse()
	if *candPath == "" {
		fmt.Fprintln(os.Stderr, "need -candidates")
		os.Exit(2)
	}

	traw, err := os.ReadFile(*truthPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read truth: %v\n", err)
		os.Exit(1)
	}
	var t truth
	if err := json.Unmarshal(traw, &t); err != nil {
		fmt.Fprintf(os.Stderr, "parse truth: %v\n", err)
		os.Exit(1)
	}
	craw, err := os.ReadFile(*candPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read candidates: %v\n", err)
		os.Exit(1)
	}
	var cf candidates.File
	if err := json.Unmarshal(craw, &cf); err != nil {
		fmt.Fprintf(os.Stderr, "parse candidates: %v\n", err)
		os.Exit(1)
	}

	// pick the ground-truth task
	idx := -1
	for i, tk := range t.Tasks {
		if (*taskID != "" && tk.ID == *taskID) || (*taskID == "" && len(t.Tasks) == 1) {
			idx = i
			break
		}
	}
	if idx < 0 {
		fmt.Fprintf(os.Stderr, "task id not found (have %d tasks; pass -task-id)\n", len(t.Tasks))
		os.Exit(1)
	}
	tk := t.Tasks[idx]

	// candidates within top N
	type cc struct {
		path, name string
		rank       int
	}
	var top []cc
	for _, c := range cf.Candidates {
		if c.Rank <= *n {
			top = append(top, cc{c.Path, c.Name, c.Rank})
		}
	}
	// matchRank: expected is "path:name"; match by name and a path-suffix rule so a
	// basename in the ground truth matches a full relative path in the candidates.
	matchRank := func(expected string) int {
		ep, en := expected, ""
		if i := strings.LastIndexByte(expected, ':'); i >= 0 {
			ep, en = expected[:i], expected[i+1:]
		} else {
			ep, en = "", expected
		}
		best := 0
		for _, c := range top {
			if c.name != en {
				continue
			}
			if ep == "" || c.path == ep || strings.HasSuffix(c.path, "/"+ep) || strings.HasSuffix(c.path, ep) {
				if best == 0 || c.rank < best {
					best = c.rank
				}
			}
		}
		return best
	}

	found := 0
	firstHitRank := 0
	fmt.Printf("Task %q  (method: %s)\n", tk.ID, cf.Method)
	fmt.Printf("Decisive expected symbols (recall@%d):\n", *n)
	for _, e := range tk.Expect {
		if r := matchRank(e); r > 0 {
			found++
			if firstHitRank == 0 || r < firstHitRank {
				firstHitRank = r
			}
			fmt.Printf("  HIT  rank %-3d %s\n", r, e)
		} else {
			fmt.Printf("  MISS  (not in top %d) %s\n", *n, e)
		}
	}
	recall := 0.0
	if len(tk.Expect) > 0 {
		recall = float64(found) / float64(len(tk.Expect))
	}
	mrr := 0.0
	if firstHitRank > 0 {
		mrr = 1.0 / float64(firstHitRank)
	}
	fmt.Printf("\nrecall@%d = %d/%d = %.2f    first-decisive-hit rank = %s (1/rank = %.3f)\n",
		*n, found, len(tk.Expect), recall, rankStr(firstHitRank), mrr)

	if len(tk.Useful) > 0 {
		uf := 0
		for _, u := range tk.Useful {
			if matchRank(u) > 0 {
				uf++
			}
		}
		fmt.Printf("also-useful found: %d/%d\n", uf, len(tk.Useful))
	}
}

func rankStr(r int) string {
	if r == 0 {
		return "none"
	}
	return fmt.Sprintf("%d", r)
}
