// Package candidates defines the ranked-candidate contract that resolve_targets,
// embed, and fuse emit (with -json) and that fuse and eval_targets read. Defined
// once here so the shape isn't re-declared as `candFile`/`jc` in each tool.
//
// Score is float64 to hold both the lexical resolver's integer scores and the
// semantic/fused cosine-derived scores; readers that only rank can ignore it.
package candidates

type Candidate struct {
	Path  string  `json:"path"`
	Name  string  `json:"name"`
	Kind  string  `json:"kind"`
	Score float64 `json:"score"`
	Rank  int     `json:"rank"`
}

type File struct {
	Task       string      `json:"task"`
	Method     string      `json:"method"` // lexical | semantic | fused
	Candidates []Candidate `json:"candidates"`
}
