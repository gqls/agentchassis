// Command embed builds and queries a semantic vector index over the analyser's
// symbols — the recall layer for target resolution (and later matched-guidelines
// retrieval) that sits on top of the lexical baseline in resolve_targets.go.
//
// Model-agnostic by design: the embedder is an interface, so the foundation model
// is swappable. Two implementations:
//   -ollama URL -model M : call an Ollama embeddings endpoint — your ollama-adapter /
//                          a small CPU model now, Thunder GPU for bulk if it's slow.
//   -local               : a deterministic offline stand-in (token hashing). It proves
//                          the pipeline (index → cosine → rank) WITHOUT a model, but is
//                          NOT semantic — use -ollama for real recall.
//
//   embed build -analysis analysis.json -out embeddings.json -ollama http://host:11434 -model nomic-embed-text
//   embed query -embeddings embeddings.json -task "why does plan_sections see no ready sections" -n 12 -ollama ... -model ...
//
// The index and query MUST use the same embedder (same vector space).
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"hash/fnv"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	acode "contextkit/internal/analysis"
	"contextkit/internal/candidates"
)

// Embedder turns texts into vectors. The only model-specific seam.
type Embedder interface {
	Embed(texts []string) ([][]float32, error)
}

// --- Ollama embedder (POST {model,input[]} to /api/embed -> {embeddings[][]}) ---
type ollamaEmbedder struct {
	base, model string
	client      *http.Client
}

func (o ollamaEmbedder) Embed(texts []string) ([][]float32, error) {
	body, _ := json.Marshal(map[string]interface{}{"model": o.model, "input": texts})
	url := strings.TrimRight(o.base, "/") + "/api/embed"
	resp, err := o.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		var b bytes.Buffer
		b.ReadFrom(resp.Body)
		return nil, fmt.Errorf("ollama %d: %s", resp.StatusCode, strings.TrimSpace(b.String()))
	}
	var out struct {
		Embeddings [][]float32 `json:"embeddings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if len(out.Embeddings) != len(texts) {
		return nil, fmt.Errorf("ollama returned %d embeddings for %d inputs", len(out.Embeddings), len(texts))
	}
	return out.Embeddings, nil
}

// --- offline stand-in: token-hashing bag of words. Deterministic, NOT semantic. ---
type hashEmbedder struct{ dim int }

func (h hashEmbedder) Embed(texts []string) ([][]float32, error) {
	out := make([][]float32, len(texts))
	for i, t := range texts {
		v := make([]float32, h.dim)
		for _, tok := range tokenize(t) {
			hh := fnv.New32a()
			hh.Write([]byte(tok))
			v[hh.Sum32()%uint32(h.dim)]++
		}
		out[i] = v
	}
	return out, nil
}

func tokenize(s string) []string {
	var toks []string
	var cur strings.Builder
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			cur.WriteRune(r)
		} else if cur.Len() > 0 {
			toks = append(toks, cur.String())
			cur.Reset()
		}
	}
	if cur.Len() > 0 {
		toks = append(toks, cur.String())
	}
	return toks
}

func l2normalize(v []float32) {
	var s float64
	for _, x := range v {
		s += float64(x) * float64(x)
	}
	if s == 0 {
		return
	}
	n := float32(math.Sqrt(s))
	for i := range v {
		v[i] /= n
	}
}

func dot(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}
	var s float64
	for i := range a {
		s += float64(a[i]) * float64(b[i])
	}
	return s
}

type indexItem struct {
	Path string    `json:"path"`
	Name string    `json:"name"`
	Kind string    `json:"kind"`
	Vec  []float32 `json:"vec"`
}

type vindex struct {
	Model string      `json:"model"`
	Dim   int         `json:"dim"`
	Items []indexItem `json:"items"`
}

type analysis = acode.Output

func embedderFrom(local bool, ollamaURL, model string) (Embedder, string, error) {
	if local {
		return hashEmbedder{dim: 256}, "local-hash-256", nil
	}
	if ollamaURL == "" {
		return nil, "", fmt.Errorf("need -local or -ollama <url>")
	}
	if model == "" {
		model = "nomic-embed-text"
	}
	return ollamaEmbedder{base: ollamaURL, model: model, client: &http.Client{Timeout: 120 * time.Second}}, model, nil
}

// symbolText is what gets embedded: name, kind, location, signature, doc.
func symbolText(pkg, path, name, kind, sig, doc string) string {
	return fmt.Sprintf("%s %s\npackage %s, file %s\n%s\n%s", kind, name, pkg, path, sig, doc)
}

func ck(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: embed build|query [flags]")
		os.Exit(2)
	}
	mode := os.Args[1]
	fs := flag.NewFlagSet(mode, flag.ExitOnError)
	local := fs.Bool("local", false, "use the offline hashing stand-in (not semantic)")
	ollamaURL := fs.String("ollama", "", "Ollama base URL (e.g. http://ollama-adapter.ai-persona-system.svc.cluster.local:11434)")
	model := fs.String("model", "", "embedding model (default nomic-embed-text)")

	switch mode {
	case "build":
		analysisPath := fs.String("analysis", "", "analyser JSON")
		out := fs.String("out", "embeddings.json", "output index path")
		batch := fs.Int("batch", 64, "embedding batch size")
		fs.Parse(os.Args[2:])
		emb, modelName, err := embedderFrom(*local, *ollamaURL, *model)
		ck(err)
		raw, err := os.ReadFile(*analysisPath)
		ck(err)
		var an analysis
		ck(json.Unmarshal(raw, &an))

		var texts []string
		var items []indexItem
		for _, f := range an.Files {
			for _, fn := range f.Functions {
				if fn.Name == "init" {
					continue
				}
				texts = append(texts, symbolText(f.Package, f.Path, fn.Name, "func", fn.Signature, fn.Doc))
				items = append(items, indexItem{f.Path, fn.Name, "func", nil})
			}
			for _, td := range f.Types {
				texts = append(texts, symbolText(f.Package, f.Path, td.Name, td.Kind, "", td.Doc))
				items = append(items, indexItem{f.Path, td.Name, td.Kind, nil})
			}
		}
		for i := 0; i < len(texts); i += *batch {
			j := i + *batch
			if j > len(texts) {
				j = len(texts)
			}
			vecs, err := emb.Embed(texts[i:j])
			ck(err)
			for k, v := range vecs {
				l2normalize(v)
				items[i+k].Vec = v
			}
			fmt.Fprintf(os.Stderr, "embedded %d/%d\n", j, len(texts))
		}
		dim := 0
		if len(items) > 0 {
			dim = len(items[0].Vec)
		}
		b, _ := json.Marshal(vindex{Model: modelName, Dim: dim, Items: items})
		ck(os.WriteFile(*out, b, 0644))
		fmt.Fprintf(os.Stderr, "wrote %s: %d symbols, dim %d, model %s\n", *out, len(items), dim, modelName)

	case "query":
		embeddingsPath := fs.String("embeddings", "embeddings.json", "index path")
		task := fs.String("task", "", "task text")
		n := fs.Int("n", 12, "top N")
		jsonOut := fs.Bool("json", false, "emit ranked candidates as JSON (for fuse/eval)")
		fs.Parse(os.Args[2:])
		if *task == "" {
			fmt.Fprintln(os.Stderr, "need -task")
			os.Exit(2)
		}
		emb, _, err := embedderFrom(*local, *ollamaURL, *model)
		ck(err)
		raw, err := os.ReadFile(*embeddingsPath)
		ck(err)
		var idx vindex
		ck(json.Unmarshal(raw, &idx))
		qv, err := emb.Embed([]string{*task})
		ck(err)
		q := qv[0]
		l2normalize(q)

		type scored struct {
			it indexItem
			s  float64
		}
		ss := make([]scored, 0, len(idx.Items))
		for _, it := range idx.Items {
			ss = append(ss, scored{it, dot(q, it.Vec)})
		}
		sort.SliceStable(ss, func(i, j int) bool { return ss[i].s > ss[j].s })

		lim := *n
		if lim > len(ss) {
			lim = len(ss)
		}
		if *jsonOut {
			out := candidates.File{Task: *task, Method: "semantic"}
			for i := 0; i < lim; i++ {
				out.Candidates = append(out.Candidates, candidates.Candidate{
					Path: ss[i].it.Path, Name: ss[i].it.Name, Kind: ss[i].it.Kind,
					Score: ss[i].s, Rank: i + 1,
				})
			}
			b, _ := json.MarshalIndent(out, "", "  ")
			fmt.Println(string(b))
			return
		}

		fmt.Printf("Semantic top %d (model %s) for task:\n  %q\n\n", *n, idx.Model, *task)
		for i := 0; i < lim; i++ {
			fmt.Printf("  %.3f  %s:%s (%s)\n", ss[i].s, ss[i].it.Path, ss[i].it.Name, ss[i].it.Kind)
		}
		fmt.Printf("\nSuggested scope (top symbols):\n")
		for i := 0; i < lim && i < 4; i++ {
			fmt.Printf("  -scope %s:%s\n", ss[i].it.Path, ss[i].it.Name)
		}
		if *local {
			fmt.Printf("\n(Stand-in embedder: ranking proves the pipeline, not semantic quality. Run with -ollama <url> -model <m> for real recall, and merge with the lexical results from resolve_targets.)\n")
		}

	default:
		fmt.Fprintln(os.Stderr, "unknown mode "+mode+" (use build|query)")
		os.Exit(2)
	}
}
